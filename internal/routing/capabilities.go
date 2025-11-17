package routing

import (
	"time"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

// ModelCapability represents the capabilities of a specific model
type ModelCapability struct {
	// Model name
	Model string

	// Provider type
	Provider models.ProviderType

	// Maximum context length in tokens
	MaxContextLength int

	// Cost per 1K input tokens (in USD)
	CostPerInputToken float64

	// Cost per 1K output tokens (in USD)
	CostPerOutputToken float64

	// Performance tier (1-5, higher is better)
	PerformanceTier int

	// Capabilities
	SupportsStreaming      bool
	SupportsFunctionCalling bool
	SupportsVision         bool
	SupportsJSON           bool

	// Specialized strengths
	BestFor []string // e.g., "code", "reasoning", "creative", "speed"

	// Average latency (p50)
	AverageLatency time.Duration
}

// ModelCapabilityMatrix maintains capabilities for all models
type ModelCapabilityMatrix struct {
	capabilities map[string]*ModelCapability
}

// NewModelCapabilityMatrix creates a new model capability matrix
func NewModelCapabilityMatrix() *ModelCapabilityMatrix {
	matrix := &ModelCapabilityMatrix{
		capabilities: make(map[string]*ModelCapability),
	}

	// Initialize with known model capabilities
	matrix.initializeCapabilities()

	return matrix
}

// initializeCapabilities initializes the matrix with known model capabilities
func (m *ModelCapabilityMatrix) initializeCapabilities() {
	// OpenAI Models
	m.capabilities["gpt-4"] = &ModelCapability{
		Model:                  "gpt-4",
		Provider:               models.ProviderOpenAI,
		MaxContextLength:       8192,
		CostPerInputToken:      0.03,  // $30 per 1M tokens
		CostPerOutputToken:     0.06,  // $60 per 1M tokens
		PerformanceTier:        5,
		SupportsStreaming:      true,
		SupportsFunctionCalling: true,
		SupportsVision:         false,
		SupportsJSON:           true,
		BestFor:                []string{"reasoning", "code", "complex"},
		AverageLatency:         2 * time.Second,
	}

	m.capabilities["gpt-4-turbo"] = &ModelCapability{
		Model:                  "gpt-4-turbo",
		Provider:               models.ProviderOpenAI,
		MaxContextLength:       128000,
		CostPerInputToken:      0.01,  // $10 per 1M tokens
		CostPerOutputToken:     0.03,  // $30 per 1M tokens
		PerformanceTier:        5,
		SupportsStreaming:      true,
		SupportsFunctionCalling: true,
		SupportsVision:         true,
		SupportsJSON:           true,
		BestFor:                []string{"reasoning", "code", "complex", "vision"},
		AverageLatency:         1500 * time.Millisecond,
	}

	m.capabilities["gpt-4o-mini"] = &ModelCapability{
		Model:                  "gpt-4o-mini",
		Provider:               models.ProviderOpenAI,
		MaxContextLength:       128000,
		CostPerInputToken:      0.00015, // $0.15 per 1M tokens
		CostPerOutputToken:     0.0006,  // $0.60 per 1M tokens
		PerformanceTier:        3,
		SupportsStreaming:      true,
		SupportsFunctionCalling: true,
		SupportsVision:         true,
		SupportsJSON:           true,
		BestFor:                []string{"speed", "cost", "simple"},
		AverageLatency:         500 * time.Millisecond,
	}

	m.capabilities["gpt-3.5-turbo"] = &ModelCapability{
		Model:                  "gpt-3.5-turbo",
		Provider:               models.ProviderOpenAI,
		MaxContextLength:       16384,
		CostPerInputToken:      0.0005, // $0.50 per 1M tokens
		CostPerOutputToken:     0.0015, // $1.50 per 1M tokens
		PerformanceTier:        3,
		SupportsStreaming:      true,
		SupportsFunctionCalling: true,
		SupportsVision:         false,
		SupportsJSON:           true,
		BestFor:                []string{"speed", "cost", "simple"},
		AverageLatency:         800 * time.Millisecond,
	}

	// Anthropic Models
	m.capabilities["claude-3-opus"] = &ModelCapability{
		Model:                  "claude-3-opus-20240229",
		Provider:               models.ProviderAnthropic,
		MaxContextLength:       200000,
		CostPerInputToken:      0.015, // $15 per 1M tokens
		CostPerOutputToken:     0.075, // $75 per 1M tokens
		PerformanceTier:        5,
		SupportsStreaming:      true,
		SupportsFunctionCalling: true,
		SupportsVision:         true,
		SupportsJSON:           true,
		BestFor:                []string{"reasoning", "complex", "creative", "long-context"},
		AverageLatency:         2500 * time.Millisecond,
	}

	m.capabilities["claude-3-sonnet"] = &ModelCapability{
		Model:                  "claude-3-sonnet-20240229",
		Provider:               models.ProviderAnthropic,
		MaxContextLength:       200000,
		CostPerInputToken:      0.003, // $3 per 1M tokens
		CostPerOutputToken:     0.015, // $15 per 1M tokens
		PerformanceTier:        4,
		SupportsStreaming:      true,
		SupportsFunctionCalling: true,
		SupportsVision:         true,
		SupportsJSON:           true,
		BestFor:                []string{"balanced", "reasoning", "code"},
		AverageLatency:         1500 * time.Millisecond,
	}

	m.capabilities["claude-3-haiku"] = &ModelCapability{
		Model:                  "claude-3-haiku-20240307",
		Provider:               models.ProviderAnthropic,
		MaxContextLength:       200000,
		CostPerInputToken:      0.00025, // $0.25 per 1M tokens
		CostPerOutputToken:     0.00125, // $1.25 per 1M tokens
		PerformanceTier:        3,
		SupportsStreaming:      true,
		SupportsFunctionCalling: true,
		SupportsVision:         true,
		SupportsJSON:           true,
		BestFor:                []string{"speed", "cost", "simple"},
		AverageLatency:         400 * time.Millisecond,
	}

	// Azure OpenAI (similar to OpenAI but different provider)
	m.capabilities["azure-gpt-4"] = &ModelCapability{
		Model:                  "gpt-4",
		Provider:               models.ProviderAzure,
		MaxContextLength:       8192,
		CostPerInputToken:      0.03,
		CostPerOutputToken:     0.06,
		PerformanceTier:        5,
		SupportsStreaming:      true,
		SupportsFunctionCalling: true,
		SupportsVision:         false,
		SupportsJSON:           true,
		BestFor:                []string{"reasoning", "code", "complex"},
		AverageLatency:         2200 * time.Millisecond,
	}

	// AWS Bedrock Models
	m.capabilities["bedrock-claude-3-sonnet"] = &ModelCapability{
		Model:                  "anthropic.claude-3-sonnet-20240229-v1:0",
		Provider:               models.ProviderBedrock,
		MaxContextLength:       200000,
		CostPerInputToken:      0.003,
		CostPerOutputToken:     0.015,
		PerformanceTier:        4,
		SupportsStreaming:      true,
		SupportsFunctionCalling: true,
		SupportsVision:         true,
		SupportsJSON:           true,
		BestFor:                []string{"balanced", "reasoning"},
		AverageLatency:         1800 * time.Millisecond,
	}

	// Google Vertex AI Models
	m.capabilities["vertex-gemini-pro"] = &ModelCapability{
		Model:                  "gemini-pro",
		Provider:               models.ProviderVertex,
		MaxContextLength:       32768,
		CostPerInputToken:      0.00025,
		CostPerOutputToken:     0.0005,
		PerformanceTier:        4,
		SupportsStreaming:      true,
		SupportsFunctionCalling: true,
		SupportsVision:         false,
		SupportsJSON:           true,
		BestFor:                []string{"balanced", "cost"},
		AverageLatency:         1200 * time.Millisecond,
	}
}

// GetCapability returns the capability for a specific model
func (m *ModelCapabilityMatrix) GetCapability(model string) *ModelCapability {
	if cap, ok := m.capabilities[model]; ok {
		return cap
	}
	return nil
}

// AddCapability adds or updates a model capability
func (m *ModelCapabilityMatrix) AddCapability(cap *ModelCapability) {
	m.capabilities[cap.Model] = cap
}

// GetModelsByComplexity returns models suitable for the given complexity
func (m *ModelCapabilityMatrix) GetModelsByComplexity(complexity QueryComplexity) []*ModelCapability {
	var models []*ModelCapability

	for _, cap := range m.capabilities {
		switch complexity {
		case ComplexitySimple:
			if cap.PerformanceTier <= 3 {
				models = append(models, cap)
			}
		case ComplexityMedium:
			if cap.PerformanceTier >= 3 && cap.PerformanceTier <= 4 {
				models = append(models, cap)
			}
		case ComplexityComplex:
			if cap.PerformanceTier >= 4 {
				models = append(models, cap)
			}
		}
	}

	return models
}

// GetModelsByCategory returns models best suited for a specific category
func (m *ModelCapabilityMatrix) GetModelsByCategory(category string) []*ModelCapability {
	var models []*ModelCapability

	for _, cap := range m.capabilities {
		for _, strength := range cap.BestFor {
			if strength == category {
				models = append(models, cap)
				break
			}
		}
	}

	return models
}

// GetCheapestModel returns the cheapest model for a given input/output token count
func (m *ModelCapabilityMatrix) GetCheapestModel(inputTokens, outputTokens int) *ModelCapability {
	var cheapest *ModelCapability
	var lowestCost float64

	for _, cap := range m.capabilities {
		cost := (float64(inputTokens)/1000)*cap.CostPerInputToken +
			(float64(outputTokens)/1000)*cap.CostPerOutputToken

		if cheapest == nil || cost < lowestCost {
			cheapest = cap
			lowestCost = cost
		}
	}

	return cheapest
}

// GetFastestModel returns the model with the lowest average latency
func (m *ModelCapabilityMatrix) GetFastestModel() *ModelCapability {
	var fastest *ModelCapability

	for _, cap := range m.capabilities {
		if fastest == nil || cap.AverageLatency < fastest.AverageLatency {
			fastest = cap
		}
	}

	return fastest
}

// EstimateCost estimates the cost for a request with the given model
func (m *ModelCapabilityMatrix) EstimateCost(model string, inputTokens, outputTokens int) float64 {
	cap := m.GetCapability(model)
	if cap == nil {
		return 0
	}

	return (float64(inputTokens)/1000)*cap.CostPerInputToken +
		(float64(outputTokens)/1000)*cap.CostPerOutputToken
}

// SupportsContextLength checks if a model supports the required context length
func (m *ModelCapabilityMatrix) SupportsContextLength(model string, tokens int) bool {
	cap := m.GetCapability(model)
	if cap == nil {
		return false
	}
	return cap.MaxContextLength >= tokens
}
