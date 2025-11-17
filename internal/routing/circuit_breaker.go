package routing

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

var (
	// ErrCircuitOpen is returned when the circuit breaker is open
	ErrCircuitOpen = errors.New("circuit breaker is open")

	// ErrTooManyRequests is returned when too many requests are in progress
	ErrTooManyRequests = errors.New("too many requests")
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	// StateClosed means requests are allowed
	StateClosed CircuitState = iota

	// StateOpen means requests are blocked
	StateOpen

	// StateHalfOpen means testing if the service has recovered
	StateHalfOpen
)

// String returns the string representation of the circuit state
func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	// MaxFailures is the number of failures before opening the circuit
	MaxFailures int

	// Timeout is how long to wait before trying again after opening
	Timeout time.Duration

	// MaxConcurrent is the maximum number of concurrent requests
	MaxConcurrent int

	// SuccessThreshold is the number of successes needed to close the circuit from half-open
	SuccessThreshold int

	// FailureRatio is the ratio of failures to total requests that triggers opening
	FailureRatio float64

	// MinSamples is the minimum number of samples before considering the failure ratio
	MinSamples int
}

// DefaultCircuitBreakerConfig returns a default configuration
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxFailures:      5,
		Timeout:          30 * time.Second,
		MaxConcurrent:    100,
		SuccessThreshold: 2,
		FailureRatio:     0.5,
		MinSamples:       10,
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	config CircuitBreakerConfig

	mu               sync.RWMutex
	state            CircuitState
	failures         int
	successes        int
	consecutiveSuccesses int
	requests         int
	lastStateChange  time.Time
	concurrentReqs   int
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config:          config,
		state:           StateClosed,
		lastStateChange: time.Now(),
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(context.Context) error) error {
	// Check if we can execute
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	// Execute the function
	err := fn(ctx)

	// Record the result
	cb.afterRequest(err)

	return err
}

// beforeRequest checks if the request should be allowed
func (cb *CircuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check concurrent request limit
	if cb.concurrentReqs >= cb.config.MaxConcurrent {
		return ErrTooManyRequests
	}

	// Check circuit state
	switch cb.state {
	case StateClosed:
		// Allow request
		cb.concurrentReqs++
		return nil

	case StateOpen:
		// Check if timeout has elapsed
		if time.Since(cb.lastStateChange) > cb.config.Timeout {
			// Move to half-open state
			cb.state = StateHalfOpen
			cb.lastStateChange = time.Now()
			cb.consecutiveSuccesses = 0
			cb.concurrentReqs++
			return nil
		}
		return ErrCircuitOpen

	case StateHalfOpen:
		// Allow limited requests to test
		if cb.concurrentReqs < 1 {
			cb.concurrentReqs++
			return nil
		}
		return ErrCircuitOpen

	default:
		return ErrCircuitOpen
	}
}

// afterRequest records the result of a request
func (cb *CircuitBreaker) afterRequest(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.concurrentReqs--
	cb.requests++

	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}
}

// onSuccess handles a successful request
func (cb *CircuitBreaker) onSuccess() {
	cb.successes++
	cb.failures = 0 // Reset consecutive failures

	switch cb.state {
	case StateHalfOpen:
		cb.consecutiveSuccesses++
		if cb.consecutiveSuccesses >= cb.config.SuccessThreshold {
			// Close the circuit
			cb.state = StateClosed
			cb.lastStateChange = time.Now()
			cb.reset()
		}

	case StateClosed:
		// Check if we should reset counters
		if cb.requests >= cb.config.MinSamples*2 {
			cb.reset()
		}
	}
}

// onFailure handles a failed request
func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.consecutiveSuccesses = 0

	switch cb.state {
	case StateHalfOpen:
		// Failed during testing, go back to open
		cb.state = StateOpen
		cb.lastStateChange = time.Now()

	case StateClosed:
		// Check if we should open the circuit
		if cb.shouldOpen() {
			cb.state = StateOpen
			cb.lastStateChange = time.Now()
		}
	}
}

// shouldOpen determines if the circuit should be opened
func (cb *CircuitBreaker) shouldOpen() bool {
	// Check consecutive failures
	if cb.failures >= cb.config.MaxFailures {
		return true
	}

	// Check failure ratio
	if cb.requests >= cb.config.MinSamples {
		failureRatio := float64(cb.failures) / float64(cb.requests)
		if failureRatio >= cb.config.FailureRatio {
			return true
		}
	}

	return false
}

// reset resets the counters
func (cb *CircuitBreaker) reset() {
	cb.failures = 0
	cb.successes = 0
	cb.consecutiveSuccesses = 0
	cb.requests = 0
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Stats returns the current statistics
func (cb *CircuitBreaker) Stats() CircuitBreakerStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return CircuitBreakerStats{
		State:            cb.state,
		Failures:         cb.failures,
		Successes:        cb.successes,
		Requests:         cb.requests,
		ConcurrentReqs:   cb.concurrentReqs,
		LastStateChange:  cb.lastStateChange,
	}
}

// CircuitBreakerStats contains circuit breaker statistics
type CircuitBreakerStats struct {
	State           CircuitState
	Failures        int
	Successes       int
	Requests        int
	ConcurrentReqs  int
	LastStateChange time.Time
}

// CircuitBreakerManager manages circuit breakers for multiple providers
type CircuitBreakerManager struct {
	mu       sync.RWMutex
	breakers map[models.ProviderType]*CircuitBreaker
	config   CircuitBreakerConfig
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager(config CircuitBreakerConfig) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[models.ProviderType]*CircuitBreaker),
		config:   config,
	}
}

// GetBreaker gets or creates a circuit breaker for a provider
func (m *CircuitBreakerManager) GetBreaker(provider models.ProviderType) *CircuitBreaker {
	m.mu.RLock()
	breaker, exists := m.breakers[provider]
	m.mu.RUnlock()

	if exists {
		return breaker
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if breaker, exists := m.breakers[provider]; exists {
		return breaker
	}

	// Create new breaker
	breaker = NewCircuitBreaker(m.config)
	m.breakers[provider] = breaker
	return breaker
}

// Execute executes a function with circuit breaker protection for a provider
func (m *CircuitBreakerManager) Execute(ctx context.Context, provider models.ProviderType, fn func(context.Context) error) error {
	breaker := m.GetBreaker(provider)
	return breaker.Execute(ctx, fn)
}

// GetStats returns statistics for all circuit breakers
func (m *CircuitBreakerManager) GetStats() map[models.ProviderType]CircuitBreakerStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[models.ProviderType]CircuitBreakerStats)
	for provider, breaker := range m.breakers {
		stats[provider] = breaker.Stats()
	}

	return stats
}

// Reset resets all circuit breakers
func (m *CircuitBreakerManager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, breaker := range m.breakers {
		breaker.mu.Lock()
		breaker.reset()
		breaker.state = StateClosed
		breaker.lastStateChange = time.Now()
		breaker.mu.Unlock()
	}
}
