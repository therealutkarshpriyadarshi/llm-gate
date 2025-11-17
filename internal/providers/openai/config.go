package openai

import (
	"errors"
	"time"
)

// Config holds OpenAI provider configuration
type Config struct {
	// APIKey is the OpenAI API key
	APIKey string

	// BaseURL is the API base URL (defaults to https://api.openai.com/v1)
	BaseURL string

	// Organization is the organization ID (optional)
	Organization string

	// Timeout is the request timeout
	Timeout time.Duration

	// MaxRetries is the maximum number of retries for failed requests
	MaxRetries int

	// RetryDelay is the initial delay between retries
	RetryDelay time.Duration
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return errors.New("OpenAI API key is required")
	}

	if c.Timeout <= 0 {
		c.Timeout = 60 * time.Second
	}

	if c.MaxRetries < 0 {
		c.MaxRetries = 3
	}

	if c.RetryDelay <= 0 {
		c.RetryDelay = 1 * time.Second
	}

	if c.BaseURL == "" {
		c.BaseURL = "https://api.openai.com/v1"
	}

	return nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL:    "https://api.openai.com/v1",
		Timeout:    60 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}
}
