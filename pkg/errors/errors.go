package errors

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidConfig indicates configuration validation failed
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrServerStartFailed indicates the server failed to start
	ErrServerStartFailed = errors.New("server failed to start")

	// ErrProviderUnavailable indicates a provider is unavailable
	ErrProviderUnavailable = errors.New("provider unavailable")

	// ErrRateLimitExceeded indicates rate limit has been exceeded
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrCacheMiss indicates a cache miss occurred
	ErrCacheMiss = errors.New("cache miss")

	// ErrInvalidRequest indicates the request is invalid
	ErrInvalidRequest = errors.New("invalid request")

	// ErrTimeout indicates a timeout occurred
	ErrTimeout = errors.New("timeout")
)

// APIError represents an API error with additional context
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

// NewAPIError creates a new APIError
func NewAPIError(code int, message string, details string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Details: details,
	}
}
