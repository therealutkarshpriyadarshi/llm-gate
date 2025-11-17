package bedrock

import (
	"fmt"
	"time"
)

// Config holds AWS Bedrock provider configuration
type Config struct {
	// AccessKeyID is the AWS access key ID
	AccessKeyID string

	// SecretAccessKey is the AWS secret access key
	SecretAccessKey string

	// SessionToken is the AWS session token (optional, for temporary credentials)
	SessionToken string

	// Region is the AWS region
	Region string

	// Timeout is the HTTP request timeout
	Timeout time.Duration

	// MaxRetries is the maximum number of retries
	MaxRetries int

	// RetryDelay is the delay between retries
	RetryDelay time.Duration
}

// DefaultConfig returns default AWS Bedrock configuration
func DefaultConfig() *Config {
	return &Config{
		Region:     "us-east-1",
		Timeout:    60 * time.Second, // Bedrock can be slower
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.AccessKeyID == "" {
		return fmt.Errorf("access key ID is required")
	}

	if c.SecretAccessKey == "" {
		return fmt.Errorf("secret access key is required")
	}

	if c.Region == "" {
		return fmt.Errorf("region is required")
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	return nil
}
