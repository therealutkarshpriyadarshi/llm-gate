package azure

import (
	"fmt"
	"time"
)

// Config holds Azure OpenAI provider configuration
type Config struct {
	// APIKey is the Azure OpenAI API key
	APIKey string

	// Endpoint is the Azure OpenAI endpoint (e.g., https://your-resource.openai.azure.com)
	Endpoint string

	// APIVersion is the Azure OpenAI API version
	APIVersion string

	// DeploymentName is the deployment name for the model
	DeploymentName string

	// Timeout is the HTTP request timeout
	Timeout time.Duration

	// MaxRetries is the maximum number of retries
	MaxRetries int

	// RetryDelay is the delay between retries
	RetryDelay time.Duration
}

// DefaultConfig returns default Azure OpenAI configuration
func DefaultConfig() *Config {
	return &Config{
		APIVersion: "2024-02-15-preview",
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

	if c.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}

	if c.APIVersion == "" {
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
