package utils

import (
	"context"
	"time"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxAttempts int
	InitialWait time.Duration
	MaxWait     time.Duration
	Multiplier  float64
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		InitialWait: 100 * time.Millisecond,
		MaxWait:     5 * time.Second,
		Multiplier:  2.0,
	}
}

// Retry executes a function with exponential backoff retry logic
func Retry(ctx context.Context, config RetryConfig, fn func() error) error {
	var lastErr error
	wait := config.InitialWait

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// Execute function
		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// Check if we should retry
		if attempt >= config.MaxAttempts {
			break
		}

		// Wait before retry
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}

		// Increase wait time
		wait = time.Duration(float64(wait) * config.Multiplier)
		if wait > config.MaxWait {
			wait = config.MaxWait
		}
	}

	return lastErr
}
