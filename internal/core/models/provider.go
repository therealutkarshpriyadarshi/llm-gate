package models

import "time"

// ProviderType represents different LLM providers
type ProviderType string

const (
	ProviderOpenAI    ProviderType = "openai"
	ProviderAnthropic ProviderType = "anthropic"
	ProviderAzure     ProviderType = "azure"
	ProviderBedrock   ProviderType = "bedrock"
	ProviderVertex    ProviderType = "vertex"
)

// ProviderCapabilities describes what a provider supports
type ProviderCapabilities struct {
	// Name is the provider name
	Name ProviderType `json:"name"`

	// SupportsStreaming indicates if the provider supports streaming
	SupportsStreaming bool `json:"supports_streaming"`

	// SupportsFunctions indicates if the provider supports function calling
	SupportsFunctions bool `json:"supports_functions"`

	// SupportsVision indicates if the provider supports vision/image inputs
	SupportsVision bool `json:"supports_vision"`

	// MaxTokens is the maximum context window size
	MaxTokens int `json:"max_tokens"`

	// Models is the list of supported models
	Models []ModelInfo `json:"models"`

	// DefaultTimeout is the default request timeout
	DefaultTimeout time.Duration `json:"default_timeout"`
}

// ModelInfo contains information about a specific model
type ModelInfo struct {
	// ID is the model identifier
	ID string `json:"id"`

	// Name is the human-readable model name
	Name string `json:"name"`

	// MaxTokens is the maximum context window
	MaxTokens int `json:"max_tokens"`

	// InputCostPer1K is the cost per 1,000 input tokens (in USD)
	InputCostPer1K float64 `json:"input_cost_per_1k"`

	// OutputCostPer1K is the cost per 1,000 output tokens (in USD)
	OutputCostPer1K float64 `json:"output_cost_per_1k"`

	// SupportsStreaming indicates if this model supports streaming
	SupportsStreaming bool `json:"supports_streaming"`

	// SupportsFunctions indicates if this model supports function calling
	SupportsFunctions bool `json:"supports_functions"`
}

// ProviderHealth represents the health status of a provider
type ProviderHealth struct {
	// Name is the provider name
	Name ProviderType `json:"name"`

	// Healthy indicates if the provider is healthy
	Healthy bool `json:"healthy"`

	// Latency is the average latency to the provider
	Latency time.Duration `json:"latency,omitempty"`

	// ErrorRate is the recent error rate (0.0 to 1.0)
	ErrorRate float64 `json:"error_rate,omitempty"`

	// LastCheck is when the health check was last performed
	LastCheck time.Time `json:"last_check"`

	// Message contains additional status information
	Message string `json:"message,omitempty"`
}

// CostInfo contains cost information for a request
type CostInfo struct {
	// InputTokens is the number of input tokens
	InputTokens int `json:"input_tokens"`

	// OutputTokens is the number of output tokens
	OutputTokens int `json:"output_tokens"`

	// InputCost is the cost of input tokens
	InputCost float64 `json:"input_cost"`

	// OutputCost is the cost of output tokens
	OutputCost float64 `json:"output_cost"`

	// TotalCost is the total cost
	TotalCost float64 `json:"total_cost"`

	// Model is the model used
	Model string `json:"model"`

	// Provider is the provider used
	Provider ProviderType `json:"provider"`
}

// CalculateCost calculates the cost based on token usage and model pricing
func CalculateCost(usage Usage, modelInfo ModelInfo, provider ProviderType) CostInfo {
	inputCost := float64(usage.PromptTokens) / 1000.0 * modelInfo.InputCostPer1K
	outputCost := float64(usage.CompletionTokens) / 1000.0 * modelInfo.OutputCostPer1K

	return CostInfo{
		InputTokens:  usage.PromptTokens,
		OutputTokens: usage.CompletionTokens,
		InputCost:    inputCost,
		OutputCost:   outputCost,
		TotalCost:    inputCost + outputCost,
		Model:        modelInfo.ID,
		Provider:     provider,
	}
}
