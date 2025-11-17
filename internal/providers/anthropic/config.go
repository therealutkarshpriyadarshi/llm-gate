package anthropic

import (
	"fmt"
	"time"
)

// Config holds Anthropic provider configuration
type Config struct {
	// APIKey is the Anthropic API key
	APIKey string

	// BaseURL is the Anthropic API base URL
	BaseURL string

	// Version is the API version
	Version string

	// Timeout is the HTTP request timeout
	Timeout time.Duration

	// MaxRetries is the maximum number of retries
	MaxRetries int

	// RetryDelay is the delay between retries
	RetryDelay time.Duration
}

// DefaultConfig returns default Anthropic configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL:    "https://api.anthropic.com",
		Version:    "2023-06-01",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	if c.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}

	if c.Version == "" {
		return fmt.Errorf("API version is required")
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	return nil
}
