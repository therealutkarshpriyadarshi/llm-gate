package providers

import (
	"fmt"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/interfaces"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/providers/anthropic"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/providers/azure"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/providers/bedrock"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/providers/openai"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/providers/vertex"
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
		return f.createAnthropicProvider(config)
	case models.ProviderAzure:
		return f.createAzureProvider(config)
	case models.ProviderBedrock:
		return f.createBedrockProvider(config)
	case models.ProviderVertex:
		return f.createVertexProvider(config)
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

// createAnthropicProvider creates an Anthropic provider instance
func (f *Factory) createAnthropicProvider(config interface{}) (interfaces.LLMProvider, error) {
	var anthropicConfig *anthropic.Config

	switch c := config.(type) {
	case *anthropic.Config:
		anthropicConfig = c
	case anthropic.Config:
		anthropicConfig = &c
	case map[string]interface{}:
		anthropicConfig = f.mapToAnthropicConfig(c)
	default:
		return nil, fmt.Errorf("invalid config type for Anthropic provider: %T", config)
	}

	if err := anthropicConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid Anthropic config: %w", err)
	}

	return anthropic.NewClient(anthropicConfig)
}

// createAzureProvider creates an Azure OpenAI provider instance
func (f *Factory) createAzureProvider(config interface{}) (interfaces.LLMProvider, error) {
	var azureConfig *azure.Config

	switch c := config.(type) {
	case *azure.Config:
		azureConfig = c
	case azure.Config:
		azureConfig = &c
	case map[string]interface{}:
		azureConfig = f.mapToAzureConfig(c)
	default:
		return nil, fmt.Errorf("invalid config type for Azure provider: %T", config)
	}

	if err := azureConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid Azure config: %w", err)
	}

	return azure.NewClient(azureConfig)
}

// createBedrockProvider creates an AWS Bedrock provider instance
func (f *Factory) createBedrockProvider(config interface{}) (interfaces.LLMProvider, error) {
	var bedrockConfig *bedrock.Config

	switch c := config.(type) {
	case *bedrock.Config:
		bedrockConfig = c
	case bedrock.Config:
		bedrockConfig = &c
	case map[string]interface{}:
		bedrockConfig = f.mapToBedrockConfig(c)
	default:
		return nil, fmt.Errorf("invalid config type for Bedrock provider: %T", config)
	}

	if err := bedrockConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid Bedrock config: %w", err)
	}

	return bedrock.NewClient(bedrockConfig)
}

// createVertexProvider creates a Google Vertex AI provider instance
func (f *Factory) createVertexProvider(config interface{}) (interfaces.LLMProvider, error) {
	var vertexConfig *vertex.Config

	switch c := config.(type) {
	case *vertex.Config:
		vertexConfig = c
	case vertex.Config:
		vertexConfig = &c
	case map[string]interface{}:
		vertexConfig = f.mapToVertexConfig(c)
	default:
		return nil, fmt.Errorf("invalid config type for Vertex provider: %T", config)
	}

	if err := vertexConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid Vertex config: %w", err)
	}

	return vertex.NewClient(vertexConfig)
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

// mapToAnthropicConfig converts a map to Anthropic config
func (f *Factory) mapToAnthropicConfig(m map[string]interface{}) *anthropic.Config {
	config := anthropic.DefaultConfig()

	if apiKey, ok := m["api_key"].(string); ok {
		config.APIKey = apiKey
	}
	if baseURL, ok := m["base_url"].(string); ok {
		config.BaseURL = baseURL
	}
	if version, ok := m["version"].(string); ok {
		config.Version = version
	}

	return config
}

// mapToAzureConfig converts a map to Azure config
func (f *Factory) mapToAzureConfig(m map[string]interface{}) *azure.Config {
	config := azure.DefaultConfig()

	if apiKey, ok := m["api_key"].(string); ok {
		config.APIKey = apiKey
	}
	if endpoint, ok := m["endpoint"].(string); ok {
		config.Endpoint = endpoint
	}
	if apiVersion, ok := m["api_version"].(string); ok {
		config.APIVersion = apiVersion
	}
	if deploymentName, ok := m["deployment_name"].(string); ok {
		config.DeploymentName = deploymentName
	}

	return config
}

// mapToBedrockConfig converts a map to Bedrock config
func (f *Factory) mapToBedrockConfig(m map[string]interface{}) *bedrock.Config {
	config := bedrock.DefaultConfig()

	if accessKeyID, ok := m["access_key_id"].(string); ok {
		config.AccessKeyID = accessKeyID
	}
	if secretAccessKey, ok := m["secret_access_key"].(string); ok {
		config.SecretAccessKey = secretAccessKey
	}
	if sessionToken, ok := m["session_token"].(string); ok {
		config.SessionToken = sessionToken
	}
	if region, ok := m["region"].(string); ok {
		config.Region = region
	}

	return config
}

// mapToVertexConfig converts a map to Vertex config
func (f *Factory) mapToVertexConfig(m map[string]interface{}) *vertex.Config {
	config := vertex.DefaultConfig()

	if projectID, ok := m["project_id"].(string); ok {
		config.ProjectID = projectID
	}
	if location, ok := m["location"].(string); ok {
		config.Location = location
	}
	if apiKey, ok := m["api_key"].(string); ok {
		config.APIKey = apiKey
	}
	if serviceAccountJSON, ok := m["service_account_json"].(string); ok {
		config.ServiceAccountJSON = serviceAccountJSON
	}

	return config
}
