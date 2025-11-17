package vertex

import (
	"fmt"
	"time"
)

// Config holds Google Vertex AI provider configuration
type Config struct {
	// ProjectID is the Google Cloud project ID
	ProjectID string

	// Location is the Google Cloud location/region
	Location string

	// APIKey is the API key for authentication (alternative to service account)
	APIKey string

	// ServiceAccountJSON is the service account JSON credentials
	ServiceAccountJSON string

	// Timeout is the HTTP request timeout
	Timeout time.Duration

	// MaxRetries is the maximum number of retries
	MaxRetries int

	// RetryDelay is the delay between retries
	RetryDelay time.Duration
}

// DefaultConfig returns default Google Vertex AI configuration
func DefaultConfig() *Config {
	return &Config{
		Location:   "us-central1",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.ProjectID == "" {
		return fmt.Errorf("project ID is required")
	}

	if c.Location == "" {
		return fmt.Errorf("location is required")
	}

	if c.APIKey == "" && c.ServiceAccountJSON == "" {
		return fmt.Errorf("either API key or service account JSON is required")
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	return nil
}
