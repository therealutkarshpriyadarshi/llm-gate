package providers

import (
	"testing"
	"time"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/providers/anthropic"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/providers/azure"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/providers/bedrock"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/providers/openai"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/providers/vertex"
)

func TestNewFactory(t *testing.T) {
	factory := NewFactory()
	if factory == nil {
		t.Error("NewFactory() returned nil")
	}
}

func TestFactory_CreateProvider_OpenAI(t *testing.T) {
	factory := NewFactory()

	tests := []struct {
		name    string
		config  interface{}
		wantErr bool
	}{
		{
			name: "valid OpenAI config struct",
			config: &openai.Config{
				APIKey:     "sk-test-key",
				BaseURL:    "https://api.openai.com/v1",
				Timeout:    30 * time.Second,
				MaxRetries: 3,
				RetryDelay: 1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid OpenAI config map",
			config: map[string]interface{}{
				"api_key":  "sk-test-key",
				"base_url": "https://api.openai.com/v1",
			},
			wantErr: false,
		},
		{
			name: "invalid OpenAI config - missing API key",
			config: &openai.Config{
				BaseURL:    "https://api.openai.com/v1",
				Timeout:    30 * time.Second,
				MaxRetries: 3,
				RetryDelay: 1 * time.Second,
			},
			wantErr: true,
		},
		{
			name:    "invalid config type",
			config:  "invalid-config",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := factory.CreateProvider(models.ProviderOpenAI, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && provider == nil {
				t.Error("CreateProvider() returned nil provider")
			}
			if !tt.wantErr && provider.Name() != models.ProviderOpenAI {
				t.Errorf("CreateProvider() returned wrong provider type: %v", provider.Name())
			}
		})
	}
}

func TestFactory_CreateProvider_Anthropic(t *testing.T) {
	factory := NewFactory()

	tests := []struct {
		name    string
		config  interface{}
		wantErr bool
	}{
		{
			name: "valid Anthropic config struct",
			config: &anthropic.Config{
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
			name: "valid Anthropic config map",
			config: map[string]interface{}{
				"api_key":  "sk-ant-test-key",
				"base_url": "https://api.anthropic.com",
				"version":  "2023-06-01",
			},
			wantErr: false,
		},
		{
			name: "invalid Anthropic config - missing API key",
			config: &anthropic.Config{
				BaseURL:    "https://api.anthropic.com",
				Version:    "2023-06-01",
				Timeout:    30 * time.Second,
				MaxRetries: 3,
				RetryDelay: 1 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := factory.CreateProvider(models.ProviderAnthropic, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && provider == nil {
				t.Error("CreateProvider() returned nil provider")
			}
			if !tt.wantErr && provider.Name() != models.ProviderAnthropic {
				t.Errorf("CreateProvider() returned wrong provider type: %v", provider.Name())
			}
		})
	}
}

func TestFactory_CreateProvider_Azure(t *testing.T) {
	factory := NewFactory()

	tests := []struct {
		name    string
		config  interface{}
		wantErr bool
	}{
		{
			name: "valid Azure config struct",
			config: &azure.Config{
				APIKey:         "test-api-key",
				Endpoint:       "https://test.openai.azure.com",
				APIVersion:     "2024-02-15-preview",
				DeploymentName: "gpt-4",
				Timeout:        30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid Azure config map",
			config: map[string]interface{}{
				"api_key":         "test-api-key",
				"endpoint":        "https://test.openai.azure.com",
				"api_version":     "2024-02-15-preview",
				"deployment_name": "gpt-4",
			},
			wantErr: false,
		},
		{
			name: "invalid Azure config - missing endpoint",
			config: &azure.Config{
				APIKey:         "test-api-key",
				APIVersion:     "2024-02-15-preview",
				DeploymentName: "gpt-4",
				Timeout:        30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := factory.CreateProvider(models.ProviderAzure, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && provider == nil {
				t.Error("CreateProvider() returned nil provider")
			}
			if !tt.wantErr && provider.Name() != models.ProviderAzure {
				t.Errorf("CreateProvider() returned wrong provider type: %v", provider.Name())
			}
		})
	}
}

func TestFactory_CreateProvider_Bedrock(t *testing.T) {
	factory := NewFactory()

	tests := []struct {
		name    string
		config  interface{}
		wantErr bool
	}{
		{
			name: "valid Bedrock config struct",
			config: &bedrock.Config{
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
			name: "valid Bedrock config map",
			config: map[string]interface{}{
				"access_key_id":     "AKIAIOSFODNN7EXAMPLE",
				"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				"region":            "us-east-1",
			},
			wantErr: false,
		},
		{
			name: "invalid Bedrock config - missing secret key",
			config: &bedrock.Config{
				AccessKeyID: "AKIAIOSFODNN7EXAMPLE",
				Region:      "us-east-1",
				Timeout:     60 * time.Second,
				MaxRetries:  3,
				RetryDelay:  1 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := factory.CreateProvider(models.ProviderBedrock, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && provider == nil {
				t.Error("CreateProvider() returned nil provider")
			}
			if !tt.wantErr && provider.Name() != models.ProviderBedrock {
				t.Errorf("CreateProvider() returned wrong provider type: %v", provider.Name())
			}
		})
	}
}

func TestFactory_CreateProvider_Vertex(t *testing.T) {
	factory := NewFactory()

	tests := []struct {
		name    string
		config  interface{}
		wantErr bool
	}{
		{
			name: "valid Vertex config struct with API key",
			config: &vertex.Config{
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
			name: "valid Vertex config map",
			config: map[string]interface{}{
				"project_id": "test-project-123",
				"location":   "us-central1",
				"api_key":    "test-api-key",
			},
			wantErr: false,
		},
		{
			name: "invalid Vertex config - missing project ID",
			config: &vertex.Config{
				Location:   "us-central1",
				APIKey:     "test-api-key",
				Timeout:    30 * time.Second,
				MaxRetries: 3,
				RetryDelay: 1 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := factory.CreateProvider(models.ProviderVertex, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && provider == nil {
				t.Error("CreateProvider() returned nil provider")
			}
			if !tt.wantErr && provider.Name() != models.ProviderVertex {
				t.Errorf("CreateProvider() returned wrong provider type: %v", provider.Name())
			}
		})
	}
}

func TestFactory_CreateProvider_UnknownProvider(t *testing.T) {
	factory := NewFactory()

	provider, err := factory.CreateProvider("unknown-provider", map[string]interface{}{})
	if err == nil {
		t.Error("CreateProvider() should return error for unknown provider")
	}
	if provider != nil {
		t.Error("CreateProvider() should return nil provider for unknown provider")
	}
}

func TestFactory_MapToConfig_AllProviders(t *testing.T) {
	factory := NewFactory()

	// Test all map conversion methods
	tests := []struct {
		name         string
		providerType models.ProviderType
		configMap    map[string]interface{}
		wantErr      bool
	}{
		{
			name:         "OpenAI map conversion",
			providerType: models.ProviderOpenAI,
			configMap: map[string]interface{}{
				"api_key":  "sk-test",
				"base_url": "https://api.openai.com/v1",
			},
			wantErr: false,
		},
		{
			name:         "Anthropic map conversion",
			providerType: models.ProviderAnthropic,
			configMap: map[string]interface{}{
				"api_key":  "sk-ant-test",
				"base_url": "https://api.anthropic.com",
				"version":  "2023-06-01",
			},
			wantErr: false,
		},
		{
			name:         "Azure map conversion",
			providerType: models.ProviderAzure,
			configMap: map[string]interface{}{
				"api_key":     "test-key",
				"endpoint":    "https://test.openai.azure.com",
				"api_version": "2024-02-15-preview",
			},
			wantErr: false,
		},
		{
			name:         "Bedrock map conversion",
			providerType: models.ProviderBedrock,
			configMap: map[string]interface{}{
				"access_key_id":     "AKIAIOSFODNN7EXAMPLE",
				"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				"region":            "us-east-1",
			},
			wantErr: false,
		},
		{
			name:         "Vertex map conversion",
			providerType: models.ProviderVertex,
			configMap: map[string]interface{}{
				"project_id": "test-project",
				"location":   "us-central1",
				"api_key":    "test-key",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := factory.CreateProvider(tt.providerType, tt.configMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && provider == nil {
				t.Error("CreateProvider() returned nil provider")
			}
		})
	}
}
