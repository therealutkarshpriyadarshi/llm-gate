package anthropic

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.BaseURL != "https://api.anthropic.com" {
		t.Errorf("DefaultConfig() BaseURL = %v, want https://api.anthropic.com", config.BaseURL)
	}

	if config.Version != "2023-06-01" {
		t.Errorf("DefaultConfig() Version = %v, want 2023-06-01", config.Version)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("DefaultConfig() Timeout = %v, want 30s", config.Timeout)
	}

	if config.MaxRetries != 3 {
		t.Errorf("DefaultConfig() MaxRetries = %v, want 3", config.MaxRetries)
	}

	if config.RetryDelay != 1*time.Second {
		t.Errorf("DefaultConfig() RetryDelay = %v, want 1s", config.RetryDelay)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				APIKey:     "sk-ant-test-key",
				BaseURL:    "https://api.anthropic.com",
				Version:    "2023-06-01",
				Timeout:    30 * time.Second,
				MaxRetries: 3,
				RetryDelay: 1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "missing API key",
			config: &Config{
				BaseURL:    "https://api.anthropic.com",
				Version:    "2023-06-01",
				Timeout:    30 * time.Second,
				MaxRetries: 3,
				RetryDelay: 1 * time.Second,
			},
			wantErr: true,
			errMsg:  "API key is required",
		},
		{
			name: "missing base URL",
			config: &Config{
				APIKey:     "sk-ant-test-key",
				Version:    "2023-06-01",
				Timeout:    30 * time.Second,
				MaxRetries: 3,
				RetryDelay: 1 * time.Second,
			},
			wantErr: true,
			errMsg:  "base URL is required",
		},
		{
			name: "missing version",
			config: &Config{
				APIKey:     "sk-ant-test-key",
				BaseURL:    "https://api.anthropic.com",
				Timeout:    30 * time.Second,
				MaxRetries: 3,
				RetryDelay: 1 * time.Second,
			},
			wantErr: true,
			errMsg:  "API version is required",
		},
		{
			name: "invalid timeout",
			config: &Config{
				APIKey:     "sk-ant-test-key",
				BaseURL:    "https://api.anthropic.com",
				Version:    "2023-06-01",
				Timeout:    0,
				MaxRetries: 3,
				RetryDelay: 1 * time.Second,
			},
			wantErr: true,
			errMsg:  "timeout must be positive",
		},
		{
			name: "negative max retries",
			config: &Config{
				APIKey:     "sk-ant-test-key",
				BaseURL:    "https://api.anthropic.com",
				Version:    "2023-06-01",
				Timeout:    30 * time.Second,
				MaxRetries: -1,
				RetryDelay: 1 * time.Second,
			},
			wantErr: true,
			errMsg:  "max retries cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Config.Validate() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}
