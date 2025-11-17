package bedrock

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Region != "us-east-1" {
		t.Errorf("DefaultConfig() Region = %v, want us-east-1", config.Region)
	}

	if config.Timeout != 60*time.Second {
		t.Errorf("DefaultConfig() Timeout = %v, want 60s", config.Timeout)
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
				AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
				SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				Region:          "us-east-1",
				Timeout:         60 * time.Second,
				MaxRetries:      3,
				RetryDelay:      1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid config with session token",
			config: &Config{
				AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
				SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				SessionToken:    "test-session-token",
				Region:          "us-east-1",
				Timeout:         60 * time.Second,
				MaxRetries:      3,
				RetryDelay:      1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "missing access key ID",
			config: &Config{
				SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				Region:          "us-east-1",
				Timeout:         60 * time.Second,
				MaxRetries:      3,
				RetryDelay:      1 * time.Second,
			},
			wantErr: true,
			errMsg:  "access key ID is required",
		},
		{
			name: "missing secret access key",
			config: &Config{
				AccessKeyID: "AKIAIOSFODNN7EXAMPLE",
				Region:      "us-east-1",
				Timeout:     60 * time.Second,
				MaxRetries:  3,
				RetryDelay:  1 * time.Second,
			},
			wantErr: true,
			errMsg:  "secret access key is required",
		},
		{
			name: "missing region",
			config: &Config{
				AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
				SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				Timeout:         60 * time.Second,
				MaxRetries:      3,
				RetryDelay:      1 * time.Second,
			},
			wantErr: true,
			errMsg:  "region is required",
		},
		{
			name: "invalid timeout",
			config: &Config{
				AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
				SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				Region:          "us-east-1",
				Timeout:         0,
				MaxRetries:      3,
				RetryDelay:      1 * time.Second,
			},
			wantErr: true,
			errMsg:  "timeout must be positive",
		},
		{
			name: "negative max retries",
			config: &Config{
				AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
				SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				Region:          "us-east-1",
				Timeout:         60 * time.Second,
				MaxRetries:      -1,
				RetryDelay:      1 * time.Second,
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
