package routing

import (
	"fmt"
	"hash/fnv"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/interfaces"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

// CostBasedStrategy routes to the cheapest provider based on estimated cost
type CostBasedStrategy struct {
	matrix   *ModelCapabilityMatrix
	analyzer *QueryAnalyzer
}

// NewCostBasedStrategy creates a new cost-based routing strategy
func NewCostBasedStrategy(matrix *ModelCapabilityMatrix, analyzer *QueryAnalyzer) *CostBasedStrategy {
	return &CostBasedStrategy{
		matrix:   matrix,
		analyzer: analyzer,
	}
}

// SelectProvider selects the cheapest provider for the request
func (s *CostBasedStrategy) SelectProvider(providers []interfaces.LLMProvider, req *models.ChatRequest) (interfaces.LLMProvider, error) {
	if len(providers) == 0 {
		return nil, ErrNoProvidersAvailable
	}

	// Analyze the request to estimate tokens
	analysis := s.analyzer.Analyze(req)

	// Estimate output tokens (rough estimate: 50% of input)
	estimatedOutput := analysis.EstimatedTokens / 2

	var cheapestProvider interfaces.LLMProvider
	var lowestCost float64

	for _, provider := range providers {
		// Get model capability
		cap := s.matrix.GetCapability(req.Model)
		if cap == nil {
			continue
		}

		// Check if this provider supports the model
		if cap.Provider != provider.Name() {
			continue
		}

		// Calculate estimated cost
		cost := s.matrix.EstimateCost(req.Model, analysis.EstimatedTokens, estimatedOutput)

		if cheapestProvider == nil || cost < lowestCost {
			cheapestProvider = provider
			lowestCost = cost
		}
	}

	if cheapestProvider == nil {
		// Fallback to first provider if no cost info available
		return providers[0], nil
	}

	return cheapestProvider, nil
}

// Name returns the strategy name
func (s *CostBasedStrategy) Name() string {
	return "cost-based"
}

// LeastLatencyStrategy routes to the provider with the lowest latency
type LeastLatencyStrategy struct {
	mu          sync.RWMutex
	latencyMap  map[models.ProviderType]time.Duration
	healthMap   map[models.ProviderType]*HealthStatus
	matrix      *ModelCapabilityMatrix
}

// NewLeastLatencyStrategy creates a new least-latency routing strategy
func NewLeastLatencyStrategy(matrix *ModelCapabilityMatrix) *LeastLatencyStrategy {
	return &LeastLatencyStrategy{
		latencyMap: make(map[models.ProviderType]time.Duration),
		healthMap:  make(map[models.ProviderType]*HealthStatus),
		matrix:     matrix,
	}
}

// SelectProvider selects the provider with the lowest latency
func (s *LeastLatencyStrategy) SelectProvider(providers []interfaces.LLMProvider, req *models.ChatRequest) (interfaces.LLMProvider, error) {
	if len(providers) == 0 {
		return nil, ErrNoProvidersAvailable
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var fastestProvider interfaces.LLMProvider
	var lowestLatency time.Duration

	for _, provider := range providers {
		// Get latency from health map
		latency := s.getLatency(provider.Name())

		if fastestProvider == nil || latency < lowestLatency {
			fastestProvider = provider
			lowestLatency = latency
		}
	}

	if fastestProvider == nil {
		return providers[0], nil
	}

	return fastestProvider, nil
}

// UpdateLatency updates the latency for a provider
func (s *LeastLatencyStrategy) UpdateLatency(providerType models.ProviderType, latency time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.latencyMap[providerType] = latency
}

// getLatency gets the latency for a provider
func (s *LeastLatencyStrategy) getLatency(providerType models.ProviderType) time.Duration {
	if latency, ok := s.latencyMap[providerType]; ok {
		return latency
	}

	// Fallback to capability matrix
	// This is a hack - in production you'd have better provider->model mapping
	if s.matrix != nil {
		for _, cap := range s.matrix.capabilities {
			if cap.Provider == providerType {
				return cap.AverageLatency
			}
		}
	}

	return 1 * time.Second // Default
}

// Name returns the strategy name
func (s *LeastLatencyStrategy) Name() string {
	return "least-latency"
}

// WeightedStrategy routes based on weighted distribution
type WeightedStrategy struct {
	mu      sync.RWMutex
	weights map[models.ProviderType]int
	counter uint64
}

// NewWeightedStrategy creates a new weighted routing strategy
func NewWeightedStrategy(weights map[models.ProviderType]int) *WeightedStrategy {
	if weights == nil {
		weights = make(map[models.ProviderType]int)
	}
	return &WeightedStrategy{
		weights: weights,
		counter: 0,
	}
}

// SelectProvider selects a provider based on weights
func (s *WeightedStrategy) SelectProvider(providers []interfaces.LLMProvider, req *models.ChatRequest) (interfaces.LLMProvider, error) {
	if len(providers) == 0 {
		return nil, ErrNoProvidersAvailable
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Calculate total weight
	totalWeight := 0
	providerWeights := make(map[interfaces.LLMProvider]int)

	for _, provider := range providers {
		weight := s.getWeight(provider.Name())
		providerWeights[provider] = weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		// No weights configured, use round-robin
		index := int(s.counter) % len(providers)
		s.counter++
		return providers[index], nil
	}

	// Select based on weights
	s.counter++
	target := int(s.counter) % totalWeight
	current := 0

	for _, provider := range providers {
		current += providerWeights[provider]
		if current > target {
			return provider, nil
		}
	}

	return providers[0], nil
}

// SetWeight sets the weight for a provider
func (s *WeightedStrategy) SetWeight(providerType models.ProviderType, weight int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.weights[providerType] = weight
}

// getWeight gets the weight for a provider
func (s *WeightedStrategy) getWeight(providerType models.ProviderType) int {
	if weight, ok := s.weights[providerType]; ok {
		return weight
	}
	return 1 // Default weight
}

// Name returns the strategy name
func (s *WeightedStrategy) Name() string {
	return "weighted"
}

// StickySessionStrategy routes the same user to the same provider
type StickySessionStrategy struct {
	mu            sync.RWMutex
	sessionMap    map[string]models.ProviderType
	fallbackStrategy RoutingStrategy
}

// NewStickySessionStrategy creates a new sticky session strategy
func NewStickySessionStrategy(fallback RoutingStrategy) *StickySessionStrategy {
	if fallback == nil {
		fallback = NewRoundRobinStrategy()
	}
	return &StickySessionStrategy{
		sessionMap:       make(map[string]models.ProviderType),
		fallbackStrategy: fallback,
	}
}

// SelectProvider selects a provider with session stickiness
func (s *StickySessionStrategy) SelectProvider(providers []interfaces.LLMProvider, req *models.ChatRequest) (interfaces.LLMProvider, error) {
	if len(providers) == 0 {
		return nil, ErrNoProvidersAvailable
	}

	// Get session key (user ID or request ID)
	sessionKey := s.getSessionKey(req)
	if sessionKey == "" {
		// No session key, use fallback strategy
		return s.fallbackStrategy.SelectProvider(providers, req)
	}

	s.mu.RLock()
	providerType, exists := s.sessionMap[sessionKey]
	s.mu.RUnlock()

	if exists {
		// Find the provider
		for _, provider := range providers {
			if provider.Name() == providerType {
				return provider, nil
			}
		}
	}

	// Session provider not found or not in healthy providers
	// Select a new provider and save it
	provider, err := s.fallbackStrategy.SelectProvider(providers, req)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.sessionMap[sessionKey] = provider.Name()
	s.mu.Unlock()

	return provider, nil
}

// getSessionKey extracts the session key from the request
func (s *StickySessionStrategy) getSessionKey(req *models.ChatRequest) string {
	// Try user ID first
	if req.Metadata.UserID != "" {
		return req.Metadata.UserID
	}

	// Try user field
	if req.User != "" {
		return req.User
	}

	// Use request ID as last resort
	return req.Metadata.RequestID
}

// ClearSession removes a session mapping
func (s *StickySessionStrategy) ClearSession(sessionKey string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessionMap, sessionKey)
}

// Name returns the strategy name
func (s *StickySessionStrategy) Name() string {
	return "sticky-session"
}

// IntelligentStrategy combines multiple factors for optimal routing
type IntelligentStrategy struct {
	analyzer    *QueryAnalyzer
	matrix      *ModelCapabilityMatrix
	healthMap   map[models.ProviderType]*HealthStatus
	mu          sync.RWMutex

	// Weights for different factors
	costWeight      float64
	latencyWeight   float64
	capabilityWeight float64
}

// NewIntelligentStrategy creates a new intelligent routing strategy
func NewIntelligentStrategy(analyzer *QueryAnalyzer, matrix *ModelCapabilityMatrix) *IntelligentStrategy {
	return &IntelligentStrategy{
		analyzer:         analyzer,
		matrix:           matrix,
		healthMap:        make(map[models.ProviderType]*HealthStatus),
		costWeight:       0.4,
		latencyWeight:    0.3,
		capabilityWeight: 0.3,
	}
}

// SelectProvider selects the best provider using multiple criteria
func (s *IntelligentStrategy) SelectProvider(providers []interfaces.LLMProvider, req *models.ChatRequest) (interfaces.LLMProvider, error) {
	if len(providers) == 0 {
		return nil, ErrNoProvidersAvailable
	}

	// Analyze the query
	analysis := s.analyzer.Analyze(req)

	// Score each provider
	type providerScore struct {
		provider interfaces.LLMProvider
		score    float64
	}

	var scores []providerScore

	for _, provider := range providers {
		score := s.scoreProvider(provider, req, analysis)
		scores = append(scores, providerScore{
			provider: provider,
			score:    score,
		})
	}

	// Sort by score (higher is better)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	if len(scores) == 0 {
		return providers[0], nil
	}

	return scores[0].provider, nil
}

// scoreProvider calculates a score for a provider
func (s *IntelligentStrategy) scoreProvider(provider interfaces.LLMProvider, req *models.ChatRequest, analysis *QueryAnalysis) float64 {
	// Get model capability
	cap := s.matrix.GetCapability(req.Model)
	if cap == nil {
		return 0
	}

	// Check if provider matches
	if cap.Provider != provider.Name() {
		return 0
	}

	var score float64

	// 1. Cost score (lower cost = higher score)
	estimatedOutput := analysis.EstimatedTokens / 2
	cost := s.matrix.EstimateCost(req.Model, analysis.EstimatedTokens, estimatedOutput)
	costScore := 1.0 / (1.0 + cost*100) // Normalize cost
	score += costScore * s.costWeight

	// 2. Latency score (lower latency = higher score)
	latency := cap.AverageLatency.Seconds()
	latencyScore := 1.0 / (1.0 + latency)
	score += latencyScore * s.latencyWeight

	// 3. Capability score (how well the model fits the query)
	capabilityScore := s.calculateCapabilityScore(cap, analysis)
	score += capabilityScore * s.capabilityWeight

	return score
}

// calculateCapabilityScore calculates how well a model's capabilities match the query
func (s *IntelligentStrategy) calculateCapabilityScore(cap *ModelCapability, analysis *QueryAnalysis) float64 {
	var score float64

	// Match complexity with performance tier
	switch analysis.Complexity {
	case ComplexitySimple:
		// Prefer lower-tier models for simple queries
		score += 1.0 - float64(cap.PerformanceTier-1)/4.0
	case ComplexityMedium:
		// Prefer mid-tier models
		score += 1.0 - math.Abs(float64(cap.PerformanceTier-3))/2.0
	case ComplexityComplex:
		// Prefer high-tier models
		score += float64(cap.PerformanceTier) / 5.0
	}

	// Match categories
	for _, category := range analysis.Categories {
		for _, strength := range cap.BestFor {
			if category == strength {
				score += 0.3
			}
		}
	}

	// Context length support
	if cap.MaxContextLength >= analysis.EstimatedTokens*2 {
		score += 0.2
	}

	return math.Min(score, 1.0)
}

// SetWeights sets the weights for different factors
func (s *IntelligentStrategy) SetWeights(cost, latency, capability float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	total := cost + latency + capability
	s.costWeight = cost / total
	s.latencyWeight = latency / total
	s.capabilityWeight = capability / total
}

// Name returns the strategy name
func (s *IntelligentStrategy) Name() string {
	return "intelligent"
}

// HashBasedStrategy routes based on hash of request for consistent routing
type HashBasedStrategy struct{}

// NewHashBasedStrategy creates a new hash-based routing strategy
func NewHashBasedStrategy() *HashBasedStrategy {
	return &HashBasedStrategy{}
}

// SelectProvider selects a provider based on hash of the request
func (s *HashBasedStrategy) SelectProvider(providers []interfaces.LLMProvider, req *models.ChatRequest) (interfaces.LLMProvider, error) {
	if len(providers) == 0 {
		return nil, ErrNoProvidersAvailable
	}

	// Create hash from user or request content
	hashKey := req.Metadata.UserID
	if hashKey == "" {
		hashKey = req.User
	}
	if hashKey == "" && len(req.Messages) > 0 {
		hashKey = req.Messages[0].Content
	}

	hash := fnv.New32a()
	hash.Write([]byte(hashKey))
	index := int(hash.Sum32()) % len(providers)

	return providers[index], nil
}

// Name returns the strategy name
func (s *HashBasedStrategy) Name() string {
	return "hash-based"
}

// StrategyFactory creates routing strategies
type StrategyFactory struct {
	analyzer *QueryAnalyzer
	matrix   *ModelCapabilityMatrix
}

// NewStrategyFactory creates a new strategy factory
func NewStrategyFactory(analyzer *QueryAnalyzer, matrix *ModelCapabilityMatrix) *StrategyFactory {
	return &StrategyFactory{
		analyzer: analyzer,
		matrix:   matrix,
	}
}

// CreateStrategy creates a strategy by name
func (f *StrategyFactory) CreateStrategy(name string, config map[string]interface{}) (RoutingStrategy, error) {
	switch name {
	case "round-robin":
		return NewRoundRobinStrategy(), nil
	case "random":
		return NewRandomStrategy(), nil
	case "cost-based":
		return NewCostBasedStrategy(f.matrix, f.analyzer), nil
	case "least-latency":
		return NewLeastLatencyStrategy(f.matrix), nil
	case "weighted":
		weights := make(map[models.ProviderType]int)
		if config != nil {
			if w, ok := config["weights"].(map[models.ProviderType]int); ok {
				weights = w
			}
		}
		return NewWeightedStrategy(weights), nil
	case "sticky-session":
		return NewStickySessionStrategy(NewRoundRobinStrategy()), nil
	case "intelligent":
		return NewIntelligentStrategy(f.analyzer, f.matrix), nil
	case "hash-based":
		return NewHashBasedStrategy(), nil
	default:
		return nil, fmt.Errorf("unknown strategy: %s", name)
	}
}
