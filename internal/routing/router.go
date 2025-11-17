package routing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/interfaces"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

// Router handles routing of requests to LLM providers
type Router struct {
	mu         sync.RWMutex
	providers  []interfaces.LLMProvider
	strategy   RoutingStrategy
	healthMap  map[models.ProviderType]*HealthStatus
	healthCheck time.Duration
	stopCh     chan struct{}
}

// HealthStatus tracks the health status of a provider
type HealthStatus struct {
	Provider    models.ProviderType
	Healthy     bool
	LastCheck   time.Time
	Latency     time.Duration
	ErrorRate   float64
	ErrorCount  int
	TotalCount  int
	LastError   error
}

// Config holds router configuration
type Config struct {
	// Strategy is the routing strategy to use
	Strategy RoutingStrategy

	// HealthCheckInterval is how often to check provider health
	HealthCheckInterval time.Duration

	// EnableHealthChecks enables automatic health checks
	EnableHealthChecks bool
}

// DefaultConfig returns a default router configuration
func DefaultConfig() *Config {
	return &Config{
		Strategy:            NewRoundRobinStrategy(),
		HealthCheckInterval: 30 * time.Second,
		EnableHealthChecks:  true,
	}
}

// NewRouter creates a new router
func NewRouter(providers []interfaces.LLMProvider, config *Config) *Router {
	if config == nil {
		config = DefaultConfig()
	}

	router := &Router{
		providers:   providers,
		strategy:    config.Strategy,
		healthMap:   make(map[models.ProviderType]*HealthStatus),
		healthCheck: config.HealthCheckInterval,
		stopCh:      make(chan struct{}),
	}

	// Initialize health status for all providers
	for _, provider := range providers {
		router.healthMap[provider.Name()] = &HealthStatus{
			Provider:  provider.Name(),
			Healthy:   true, // Assume healthy initially
			LastCheck: time.Now(),
		}
	}

	// Start health check goroutine if enabled
	if config.EnableHealthChecks {
		go router.healthCheckLoop()
	}

	return router
}

// Route routes a request to an appropriate provider
func (r *Router) Route(ctx context.Context, req *models.ChatRequest) (interfaces.LLMProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	// Get healthy providers
	healthyProviders := r.getHealthyProviders()
	if len(healthyProviders) == 0 {
		return nil, fmt.Errorf("no healthy providers available")
	}

	// Use strategy to select a provider
	provider, err := r.strategy.SelectProvider(healthyProviders, req)
	if err != nil {
		return nil, fmt.Errorf("failed to select provider: %w", err)
	}

	return provider, nil
}

// RouteWithFallback routes a request with automatic fallback on failure
func (r *Router) RouteWithFallback(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	r.mu.RLock()
	healthyProviders := r.getHealthyProviders()
	r.mu.RUnlock()

	if len(healthyProviders) == 0 {
		return nil, fmt.Errorf("no healthy providers available")
	}

	var lastErr error
	for i := 0; i < len(healthyProviders); i++ {
		// Select provider
		provider, err := r.strategy.SelectProvider(healthyProviders, req)
		if err != nil {
			return nil, fmt.Errorf("failed to select provider: %w", err)
		}

		// Try to send request
		resp, err := provider.ChatCompletion(ctx, req)
		if err == nil {
			// Success! Record success
			r.recordSuccess(provider.Name())
			return resp, nil
		}

		// Record failure
		r.recordError(provider.Name(), err)
		lastErr = err

		// Remove failed provider from list for next iteration
		healthyProviders = r.removeProvider(healthyProviders, provider)
	}

	return nil, fmt.Errorf("all providers failed, last error: %w", lastErr)
}

// AddProvider adds a new provider to the router
func (r *Router) AddProvider(provider interfaces.LLMProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.providers = append(r.providers, provider)
	r.healthMap[provider.Name()] = &HealthStatus{
		Provider:  provider.Name(),
		Healthy:   true,
		LastCheck: time.Now(),
	}
}

// RemoveProvider removes a provider from the router
func (r *Router) RemoveProvider(providerType models.ProviderType) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, provider := range r.providers {
		if provider.Name() == providerType {
			r.providers = append(r.providers[:i], r.providers[i+1:]...)
			delete(r.healthMap, providerType)
			return nil
		}
	}

	return fmt.Errorf("provider %s not found", providerType)
}

// GetHealthStatus returns the health status of all providers
func (r *Router) GetHealthStatus() map[models.ProviderType]*HealthStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Create a copy to avoid race conditions
	statusCopy := make(map[models.ProviderType]*HealthStatus)
	for k, v := range r.healthMap {
		statusCopy[k] = &HealthStatus{
			Provider:   v.Provider,
			Healthy:    v.Healthy,
			LastCheck:  v.LastCheck,
			Latency:    v.Latency,
			ErrorRate:  v.ErrorRate,
			ErrorCount: v.ErrorCount,
			TotalCount: v.TotalCount,
			LastError:  v.LastError,
		}
	}

	return statusCopy
}

// Close stops the health check loop and cleans up
func (r *Router) Close() {
	close(r.stopCh)
}

// getHealthyProviders returns a list of healthy providers
func (r *Router) getHealthyProviders() []interfaces.LLMProvider {
	var healthy []interfaces.LLMProvider
	for _, provider := range r.providers {
		if status, ok := r.healthMap[provider.Name()]; ok && status.Healthy {
			healthy = append(healthy, provider)
		}
	}
	return healthy
}

// removeProvider removes a provider from the list
func (r *Router) removeProvider(providers []interfaces.LLMProvider, toRemove interfaces.LLMProvider) []interfaces.LLMProvider {
	var result []interfaces.LLMProvider
	for _, p := range providers {
		if p.Name() != toRemove.Name() {
			result = append(result, p)
		}
	}
	return result
}

// recordSuccess records a successful request
func (r *Router) recordSuccess(providerType models.ProviderType) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if status, ok := r.healthMap[providerType]; ok {
		status.TotalCount++
		status.ErrorRate = float64(status.ErrorCount) / float64(status.TotalCount)
	}
}

// recordError records a failed request
func (r *Router) recordError(providerType models.ProviderType, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if status, ok := r.healthMap[providerType]; ok {
		status.ErrorCount++
		status.TotalCount++
		status.LastError = err
		status.ErrorRate = float64(status.ErrorCount) / float64(status.TotalCount)

		// Mark as unhealthy if error rate is too high (>50%)
		if status.ErrorRate > 0.5 && status.TotalCount >= 10 {
			status.Healthy = false
		}
	}
}

// healthCheckLoop periodically checks provider health
func (r *Router) healthCheckLoop() {
	ticker := time.NewTicker(r.healthCheck)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.checkAllProviders()
		case <-r.stopCh:
			return
		}
	}
}

// checkAllProviders checks the health of all providers
func (r *Router) checkAllProviders() {
	r.mu.Lock()
	providers := make([]interfaces.LLMProvider, len(r.providers))
	copy(providers, r.providers)
	r.mu.Unlock()

	for _, provider := range providers {
		r.checkProvider(provider)
	}
}

// checkProvider checks the health of a single provider
func (r *Router) checkProvider(provider interfaces.LLMProvider) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	health := provider.HealthCheck(ctx)

	r.mu.Lock()
	defer r.mu.Unlock()

	if status, ok := r.healthMap[provider.Name()]; ok {
		status.Healthy = health.Healthy
		status.LastCheck = health.LastCheck
		status.Latency = health.Latency

		// Reset error count if healthy
		if health.Healthy {
			status.ErrorCount = 0
			status.TotalCount = 0
			status.ErrorRate = 0
			status.LastError = nil
		}
	}
}
