package providers

import (
	"fmt"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/interfaces"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/providers/openai"
)

// Factory creates provider instances
type Factory struct{}

// NewFactory creates a new provider factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateProvider creates a new provider instance
func (f *Factory) CreateProvider(providerType models.ProviderType, config interface{}) (interfaces.LLMProvider, error) {
	switch providerType {
	case models.ProviderOpenAI:
		return f.createOpenAIProvider(config)
	case models.ProviderAnthropic:
		return nil, fmt.Errorf("Anthropic provider not yet implemented")
	case models.ProviderAzure:
		return nil, fmt.Errorf("Azure provider not yet implemented")
	case models.ProviderBedrock:
		return nil, fmt.Errorf("AWS Bedrock provider not yet implemented")
	case models.ProviderVertex:
		return nil, fmt.Errorf("Google Vertex provider not yet implemented")
	default:
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}
}

// createOpenAIProvider creates an OpenAI provider instance
func (f *Factory) createOpenAIProvider(config interface{}) (interfaces.LLMProvider, error) {
	var openAIConfig *openai.Config

	switch c := config.(type) {
	case *openai.Config:
		openAIConfig = c
	case openai.Config:
		openAIConfig = &c
	case map[string]interface{}:
		openAIConfig = f.mapToOpenAIConfig(c)
	default:
		return nil, fmt.Errorf("invalid config type for OpenAI provider: %T", config)
	}

	if err := openAIConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid OpenAI config: %w", err)
	}

	return openai.NewClient(openAIConfig)
}

// mapToOpenAIConfig converts a map to OpenAI config
func (f *Factory) mapToOpenAIConfig(m map[string]interface{}) *openai.Config {
	config := openai.DefaultConfig()

	if apiKey, ok := m["api_key"].(string); ok {
		config.APIKey = apiKey
	}
	if baseURL, ok := m["base_url"].(string); ok {
		config.BaseURL = baseURL
	}
	if org, ok := m["organization"].(string); ok {
		config.Organization = org
	}

	return config
}
