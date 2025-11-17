package vertex

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Location != "us-central1" {
		t.Errorf("DefaultConfig() Location = %v, want us-central1", config.Location)
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
			name: "valid config with API key",
			config: &Config{
				ProjectID:  "test-project-123",
				Location:   "us-central1",
				APIKey:     "test-api-key",
				Timeout:    30 * time.Second,
				MaxRetries: 3,
				RetryDelay: 1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid config with service account",
			config: &Config{
				ProjectID:          "test-project-123",
				Location:           "us-central1",
				ServiceAccountJSON: `{"type": "service_account"}`,
				Timeout:            30 * time.Second,
				MaxRetries:         3,
				RetryDelay:         1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid config with both API key and service account",
			config: &Config{
				ProjectID:          "test-project-123",
				Location:           "us-central1",
				APIKey:             "test-api-key",
				ServiceAccountJSON: `{"type": "service_account"}`,
				Timeout:            30 * time.Second,
				MaxRetries:         3,
				RetryDelay:         1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "missing project ID",
			config: &Config{
				Location:   "us-central1",
				APIKey:     "test-api-key",
				Timeout:    30 * time.Second,
				MaxRetries: 3,
				RetryDelay: 1 * time.Second,
			},
			wantErr: true,
			errMsg:  "project ID is required",
		},
		{
			name: "missing location",
			config: &Config{
				ProjectID:  "test-project-123",
				APIKey:     "test-api-key",
				Timeout:    30 * time.Second,
				MaxRetries: 3,
				RetryDelay: 1 * time.Second,
			},
			wantErr: true,
			errMsg:  "location is required",
		},
		{
			name: "missing authentication",
			config: &Config{
				ProjectID:  "test-project-123",
				Location:   "us-central1",
				Timeout:    30 * time.Second,
				MaxRetries: 3,
				RetryDelay: 1 * time.Second,
			},
			wantErr: true,
			errMsg:  "either API key or service account JSON is required",
		},
		{
			name: "invalid timeout",
			config: &Config{
				ProjectID:  "test-project-123",
				Location:   "us-central1",
				APIKey:     "test-api-key",
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
				ProjectID:  "test-project-123",
				Location:   "us-central1",
				APIKey:     "test-api-key",
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
