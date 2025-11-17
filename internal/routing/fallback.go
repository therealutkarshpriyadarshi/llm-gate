package routing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/interfaces"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
	"github.com/therealutkarshpriyadarshi/llm-gate/pkg/utils"
)

// FallbackChain represents a chain of providers to try in order
type FallbackChain struct {
	providers        []interfaces.LLMProvider
	strategy         RoutingStrategy
	circuitBreaker   *CircuitBreakerManager
	retryConfig      utils.RetryConfig
	maxAttempts      int
	enableRetry      bool
	recordMetrics    bool
}

// FallbackChainConfig configures the fallback chain
type FallbackChainConfig struct {
	// Strategy for selecting providers
	Strategy RoutingStrategy

	// CircuitBreaker manager
	CircuitBreaker *CircuitBreakerManager

	// RetryConfig for individual provider attempts
	RetryConfig utils.RetryConfig

	// MaxAttempts across all providers
	MaxAttempts int

	// EnableRetry enables retry logic
	EnableRetry bool

	// RecordMetrics enables metrics recording
	RecordMetrics bool
}

// DefaultFallbackChainConfig returns a default configuration
func DefaultFallbackChainConfig() FallbackChainConfig {
	return FallbackChainConfig{
		Strategy:       NewRoundRobinStrategy(),
		CircuitBreaker: NewCircuitBreakerManager(DefaultCircuitBreakerConfig()),
		RetryConfig:    utils.DefaultRetryConfig(),
		MaxAttempts:    3,
		EnableRetry:    true,
		RecordMetrics:  true,
	}
}

// NewFallbackChain creates a new fallback chain
func NewFallbackChain(providers []interfaces.LLMProvider, config FallbackChainConfig) *FallbackChain {
	if config.Strategy == nil {
		config.Strategy = NewRoundRobinStrategy()
	}
	if config.CircuitBreaker == nil {
		config.CircuitBreaker = NewCircuitBreakerManager(DefaultCircuitBreakerConfig())
	}

	return &FallbackChain{
		providers:      providers,
		strategy:       config.Strategy,
		circuitBreaker: config.CircuitBreaker,
		retryConfig:    config.RetryConfig,
		maxAttempts:    config.MaxAttempts,
		enableRetry:    config.EnableRetry,
		recordMetrics:  config.RecordMetrics,
	}
}

// Execute executes a request with fallback
func (fc *FallbackChain) Execute(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	if len(fc.providers) == 0 {
		return nil, ErrNoProvidersAvailable
	}

	var lastErr error
	attemptCount := 0
	triedProviders := make(map[models.ProviderType]bool)

	for attemptCount < fc.maxAttempts {
		// Get available providers (not tried yet or circuit not open)
		available := fc.getAvailableProviders(triedProviders)
		if len(available) == 0 {
			break
		}

		// Select a provider
		provider, err := fc.strategy.SelectProvider(available, req)
		if err != nil {
			lastErr = err
			break
		}

		// Mark as tried
		triedProviders[provider.Name()] = true
		attemptCount++

		// Try to execute with this provider
		resp, err := fc.executeWithProvider(ctx, provider, req)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Check if we should retry
		if !fc.shouldRetry(err) {
			break
		}
	}

	if lastErr == nil {
		lastErr = ErrAllProvidersFailed
	}

	return nil, fmt.Errorf("fallback chain exhausted after %d attempts: %w", attemptCount, lastErr)
}

// executeWithProvider executes a request with a specific provider
func (fc *FallbackChain) executeWithProvider(ctx context.Context, provider interfaces.LLMProvider, req *models.ChatRequest) (*models.ChatResponse, error) {
	var resp *models.ChatResponse
	var err error

	// Execute with circuit breaker protection
	executeErr := fc.circuitBreaker.Execute(ctx, provider.Name(), func(ctx context.Context) error {
		if fc.enableRetry {
			// Execute with retry
			err = utils.Retry(ctx, fc.retryConfig, func() error {
				var retryErr error
				resp, retryErr = provider.ChatCompletion(ctx, req)
				return retryErr
			})
		} else {
			// Execute without retry
			resp, err = provider.ChatCompletion(ctx, req)
		}
		return err
	})

	if executeErr != nil {
		return nil, executeErr
	}

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// getAvailableProviders returns providers that haven't been tried or have closed circuits
func (fc *FallbackChain) getAvailableProviders(tried map[models.ProviderType]bool) []interfaces.LLMProvider {
	var available []interfaces.LLMProvider

	for _, provider := range fc.providers {
		// Skip if already tried
		if tried[provider.Name()] {
			continue
		}

		// Check circuit breaker state
		breaker := fc.circuitBreaker.GetBreaker(provider.Name())
		if breaker.State() == StateOpen {
			continue
		}

		available = append(available, provider)
	}

	return available
}

// shouldRetry determines if we should retry after an error
func (fc *FallbackChain) shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	// Don't retry on certain errors
	if errors.Is(err, ErrCircuitOpen) {
		return true // Try another provider
	}

	// Add more error type checks here
	// For example, don't retry on validation errors, but do retry on network errors

	return true
}

// FallbackRouter combines routing with fallback capability
type FallbackRouter struct {
	router         *Router
	fallbackChain  *FallbackChain
	analyzer       *QueryAnalyzer
	matrix         *ModelCapabilityMatrix
}

// NewFallbackRouter creates a new fallback router
func NewFallbackRouter(router *Router, chain *FallbackChain, analyzer *QueryAnalyzer, matrix *ModelCapabilityMatrix) *FallbackRouter {
	return &FallbackRouter{
		router:        router,
		fallbackChain: chain,
		analyzer:      analyzer,
		matrix:        matrix,
	}
}

// Route routes a request with intelligent fallback
func (fr *FallbackRouter) Route(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	// Analyze the query first
	analysis := fr.analyzer.Analyze(req)

	// Get recommended models if not specified
	if req.Model == "" {
		recommendedModels := fr.analyzer.GetRecommendedModels(analysis)
		if len(recommendedModels) > 0 {
			req.Model = recommendedModels[0]
		} else {
			req.Model = "gpt-3.5-turbo" // Default fallback
		}
	}

	// Execute with fallback chain
	return fr.fallbackChain.Execute(ctx, req)
}

// GetAnalysis returns the query analysis for a request
func (fr *FallbackRouter) GetAnalysis(req *models.ChatRequest) *QueryAnalysis {
	return fr.analyzer.Analyze(req)
}

// RequestHedging implements request hedging - send to multiple providers and use first response
type RequestHedging struct {
	providers      []interfaces.LLMProvider
	hedgingDelay   time.Duration
	maxConcurrent  int
}

// NewRequestHedging creates a new request hedging instance
func NewRequestHedging(providers []interfaces.LLMProvider, hedgingDelay time.Duration, maxConcurrent int) *RequestHedging {
	return &RequestHedging{
		providers:     providers,
		hedgingDelay:  hedgingDelay,
		maxConcurrent: maxConcurrent,
	}
}

// Execute executes a request with hedging
func (rh *RequestHedging) Execute(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	if len(rh.providers) == 0 {
		return nil, ErrNoProvidersAvailable
	}

	type result struct {
		resp *models.ChatResponse
		err  error
	}

	resultChan := make(chan result, rh.maxConcurrent)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sentCount := 0

	// Send to first provider immediately
	go func() {
		resp, err := rh.providers[0].ChatCompletion(ctx, req)
		resultChan <- result{resp, err}
	}()
	sentCount++

	// Set up hedging timer
	hedgingTimer := time.NewTimer(rh.hedgingDelay)
	defer hedgingTimer.Stop()

	providerIndex := 1

	for {
		select {
		case res := <-resultChan:
			if res.err == nil {
				// Got a successful response, cancel others and return
				cancel()
				return res.resp, nil
			}

			// If all requests have been sent and this was the last response
			if sentCount >= len(rh.providers) && sentCount == len(rh.providers) {
				return nil, res.err
			}

		case <-hedgingTimer.C:
			// Time to send to another provider
			if providerIndex < len(rh.providers) && sentCount < rh.maxConcurrent {
				provider := rh.providers[providerIndex]
				providerIndex++
				sentCount++

				go func() {
					resp, err := provider.ChatCompletion(ctx, req)
					resultChan <- result{resp, err}
				}()

				// Reset timer for next hedge
				if providerIndex < len(rh.providers) {
					hedgingTimer.Reset(rh.hedgingDelay)
				}
			}

		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}
