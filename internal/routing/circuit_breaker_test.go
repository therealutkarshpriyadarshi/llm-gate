package routing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

func TestCircuitBreaker_StateClosed(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxFailures:      3,
		Timeout:          1 * time.Second,
		MaxConcurrent:    10,
		SuccessThreshold: 2,
		FailureRatio:     0.5,
		MinSamples:       5,
	}

	cb := NewCircuitBreaker(config)

	// Initial state should be closed
	if cb.State() != StateClosed {
		t.Errorf("Expected initial state to be Closed, got %s", cb.State())
	}

	// Successful requests should keep circuit closed
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		err := cb.Execute(ctx, func(ctx context.Context) error {
			return nil
		})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}

	if cb.State() != StateClosed {
		t.Errorf("Expected state to remain Closed, got %s", cb.State())
	}
}

func TestCircuitBreaker_StateOpen(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxFailures:      3,
		Timeout:          100 * time.Millisecond,
		MaxConcurrent:    10,
		SuccessThreshold: 2,
		FailureRatio:     0.5,
		MinSamples:       5,
	}

	cb := NewCircuitBreaker(config)
	ctx := context.Background()
	testErr := errors.New("test error")

	// Trigger failures to open the circuit
	for i := 0; i < 3; i++ {
		_ = cb.Execute(ctx, func(ctx context.Context) error {
			return testErr
		})
	}

	if cb.State() != StateOpen {
		t.Errorf("Expected state to be Open after %d failures, got %s", 3, cb.State())
	}

	// Requests should be rejected while circuit is open
	err := cb.Execute(ctx, func(ctx context.Context) error {
		return nil
	})

	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("Expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreaker_StateHalfOpen(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxFailures:      2,
		Timeout:          50 * time.Millisecond,
		MaxConcurrent:    10,
		SuccessThreshold: 2,
		FailureRatio:     0.5,
		MinSamples:       5,
	}

	cb := NewCircuitBreaker(config)
	ctx := context.Background()
	testErr := errors.New("test error")

	// Open the circuit
	for i := 0; i < 2; i++ {
		_ = cb.Execute(ctx, func(ctx context.Context) error {
			return testErr
		})
	}

	if cb.State() != StateOpen {
		t.Errorf("Expected state to be Open, got %s", cb.State())
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Next request should move to half-open
	err := cb.Execute(ctx, func(ctx context.Context) error {
		return nil // Success
	})

	if err != nil {
		t.Errorf("Unexpected error in half-open state: %v", err)
	}
}

func TestCircuitBreaker_Recovery(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxFailures:      2,
		Timeout:          50 * time.Millisecond,
		MaxConcurrent:    10,
		SuccessThreshold: 2,
		FailureRatio:     0.5,
		MinSamples:       5,
	}

	cb := NewCircuitBreaker(config)
	ctx := context.Background()
	testErr := errors.New("test error")

	// Open the circuit
	for i := 0; i < 2; i++ {
		_ = cb.Execute(ctx, func(ctx context.Context) error {
			return testErr
		})
	}

	// Wait for timeout to move to half-open
	time.Sleep(60 * time.Millisecond)

	// Succeed enough times to close the circuit
	for i := 0; i < config.SuccessThreshold; i++ {
		err := cb.Execute(ctx, func(ctx context.Context) error {
			return nil
		})
		if err != nil {
			t.Errorf("Unexpected error during recovery: %v", err)
		}
	}

	if cb.State() != StateClosed {
		t.Errorf("Expected state to be Closed after recovery, got %s", cb.State())
	}
}

func TestCircuitBreaker_FailureRatio(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxFailures:      10, // High so we test ratio instead
		Timeout:          1 * time.Second,
		MaxConcurrent:    10,
		SuccessThreshold: 2,
		FailureRatio:     0.5,
		MinSamples:       10,
	}

	cb := NewCircuitBreaker(config)
	ctx := context.Background()
	testErr := errors.New("test error")

	// Execute 10 requests with 6 failures (60% failure rate)
	for i := 0; i < 10; i++ {
		var err error
		if i < 6 {
			err = testErr
		}
		_ = cb.Execute(ctx, func(ctx context.Context) error {
			return err
		})
	}

	if cb.State() != StateOpen {
		t.Errorf("Expected state to be Open due to high failure ratio, got %s", cb.State())
	}
}

func TestCircuitBreaker_MaxConcurrent(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxFailures:      5,
		Timeout:          1 * time.Second,
		MaxConcurrent:    2,
		SuccessThreshold: 2,
		FailureRatio:     0.5,
		MinSamples:       5,
	}

	cb := NewCircuitBreaker(config)
	ctx := context.Background()

	// Start 2 concurrent requests
	done := make(chan bool, 2)
	for i := 0; i < 2; i++ {
		go func() {
			_ = cb.Execute(ctx, func(ctx context.Context) error {
				time.Sleep(50 * time.Millisecond)
				return nil
			})
			done <- true
		}()
	}

	// Wait a bit for the goroutines to start
	time.Sleep(10 * time.Millisecond)

	// Third request should fail due to max concurrent limit
	err := cb.Execute(ctx, func(ctx context.Context) error {
		return nil
	})

	if !errors.Is(err, ErrTooManyRequests) {
		t.Errorf("Expected ErrTooManyRequests, got %v", err)
	}

	// Wait for concurrent requests to finish
	<-done
	<-done
}

func TestCircuitBreaker_Stats(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	cb := NewCircuitBreaker(config)
	ctx := context.Background()

	// Execute some requests
	_ = cb.Execute(ctx, func(ctx context.Context) error {
		return nil
	})
	_ = cb.Execute(ctx, func(ctx context.Context) error {
		return errors.New("test error")
	})

	stats := cb.Stats()

	if stats.State != StateClosed {
		t.Errorf("Expected state Closed, got %s", stats.State)
	}

	if stats.Requests != 2 {
		t.Errorf("Expected 2 requests, got %d", stats.Requests)
	}

	if stats.Successes != 1 {
		t.Errorf("Expected 1 success, got %d", stats.Successes)
	}

	if stats.Failures != 1 {
		t.Errorf("Expected 1 failure, got %d", stats.Failures)
	}
}

func TestCircuitBreakerManager(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	manager := NewCircuitBreakerManager(config)

	// Get breaker for a provider
	breaker1 := manager.GetBreaker(models.ProviderOpenAI)
	if breaker1 == nil {
		t.Error("Expected breaker to be created")
	}

	// Getting again should return same instance
	breaker2 := manager.GetBreaker(models.ProviderOpenAI)
	if breaker1 != breaker2 {
		t.Error("Expected same breaker instance")
	}

	// Different provider should get different breaker
	breaker3 := manager.GetBreaker(models.ProviderAnthropic)
	if breaker1 == breaker3 {
		t.Error("Expected different breaker for different provider")
	}
}

func TestCircuitBreakerManager_Execute(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxFailures:      2,
		Timeout:          1 * time.Second,
		MaxConcurrent:    10,
		SuccessThreshold: 2,
		FailureRatio:     0.5,
		MinSamples:       5,
	}

	manager := NewCircuitBreakerManager(config)
	ctx := context.Background()
	testErr := errors.New("test error")

	// Execute with circuit breaker
	err := manager.Execute(ctx, models.ProviderOpenAI, func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Trigger failures
	for i := 0; i < 2; i++ {
		_ = manager.Execute(ctx, models.ProviderOpenAI, func(ctx context.Context) error {
			return testErr
		})
	}

	// Next request should fail with circuit open
	err = manager.Execute(ctx, models.ProviderOpenAI, func(ctx context.Context) error {
		return nil
	})

	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("Expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreakerManager_GetStats(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	manager := NewCircuitBreakerManager(config)
	ctx := context.Background()

	// Execute on different providers
	_ = manager.Execute(ctx, models.ProviderOpenAI, func(ctx context.Context) error {
		return nil
	})
	_ = manager.Execute(ctx, models.ProviderAnthropic, func(ctx context.Context) error {
		return nil
	})

	stats := manager.GetStats()

	if len(stats) != 2 {
		t.Errorf("Expected stats for 2 providers, got %d", len(stats))
	}

	if _, ok := stats[models.ProviderOpenAI]; !ok {
		t.Error("Expected stats for OpenAI provider")
	}

	if _, ok := stats[models.ProviderAnthropic]; !ok {
		t.Error("Expected stats for Anthropic provider")
	}
}

func TestCircuitBreakerManager_Reset(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxFailures:      2,
		Timeout:          1 * time.Second,
		MaxConcurrent:    10,
		SuccessThreshold: 2,
		FailureRatio:     0.5,
		MinSamples:       5,
	}

	manager := NewCircuitBreakerManager(config)
	ctx := context.Background()
	testErr := errors.New("test error")

	// Open the circuit
	for i := 0; i < 2; i++ {
		_ = manager.Execute(ctx, models.ProviderOpenAI, func(ctx context.Context) error {
			return testErr
		})
	}

	breaker := manager.GetBreaker(models.ProviderOpenAI)
	if breaker.State() != StateOpen {
		t.Error("Expected circuit to be open")
	}

	// Reset
	manager.Reset()

	if breaker.State() != StateClosed {
		t.Errorf("Expected circuit to be closed after reset, got %s", breaker.State())
	}

	stats := breaker.Stats()
	if stats.Failures != 0 || stats.Requests != 0 {
		t.Error("Expected stats to be reset")
	}
}

func BenchmarkCircuitBreaker_Execute(b *testing.B) {
	config := DefaultCircuitBreakerConfig()
	cb := NewCircuitBreaker(config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.Execute(ctx, func(ctx context.Context) error {
			return nil
		})
	}
}
