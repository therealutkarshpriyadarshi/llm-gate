package azure

// AzureMessage represents a message in Azure OpenAI format
type AzureMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// AzureRequest represents an Azure OpenAI API request
type AzureRequest struct {
	Messages         []AzureMessage `json:"messages"`
	Temperature      *float64       `json:"temperature,omitempty"`
	MaxTokens        *int           `json:"max_tokens,omitempty"`
	TopP             *float64       `json:"top_p,omitempty"`
	N                *int           `json:"n,omitempty"`
	Stream           bool           `json:"stream,omitempty"`
	Stop             []string       `json:"stop,omitempty"`
	PresencePenalty  *float64       `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64       `json:"frequency_penalty,omitempty"`
	User             string         `json:"user,omitempty"`
}

// AzureResponse represents an Azure OpenAI API response
type AzureResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []AzureChoice  `json:"choices"`
	Usage   AzureUsage     `json:"usage"`
}

// AzureChoice represents a response choice
type AzureChoice struct {
	Index        int          `json:"index"`
	Message      AzureMessage `json:"message"`
	FinishReason string       `json:"finish_reason"`
}

// AzureUsage represents token usage
type AzureUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// AzureStreamChunk represents a streaming response chunk
type AzureStreamChunk struct {
	ID      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Model   string               `json:"model"`
	Choices []AzureStreamChoice  `json:"choices"`
}

// AzureStreamChoice represents a streaming choice
type AzureStreamChoice struct {
	Index        int                 `json:"index"`
	Delta        AzureStreamDelta    `json:"delta"`
	FinishReason string              `json:"finish_reason,omitempty"`
}

// AzureStreamDelta represents incremental content
type AzureStreamDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// AzureError represents an error from the Azure OpenAI API
type AzureError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// ModelPricing maps Azure deployment models to their base model pricing
// Note: Azure pricing depends on your specific Azure subscription
// These are approximate costs based on standard Azure OpenAI pricing
var ModelPricing = map[string]struct {
	InputCostPer1K  float64
	OutputCostPer1K float64
	MaxTokens       int
}{
	"gpt-4": {
		InputCostPer1K:  0.03,
		OutputCostPer1K: 0.06,
		MaxTokens:       8192,
	},
	"gpt-4-32k": {
		InputCostPer1K:  0.06,
		OutputCostPer1K: 0.12,
		MaxTokens:       32768,
	},
	"gpt-4-turbo": {
		InputCostPer1K:  0.01,
		OutputCostPer1K: 0.03,
		MaxTokens:       128000,
	},
	"gpt-35-turbo": {
		InputCostPer1K:  0.0015,
		OutputCostPer1K: 0.002,
		MaxTokens:       4096,
	},
	"gpt-35-turbo-16k": {
		InputCostPer1K:  0.003,
		OutputCostPer1K: 0.004,
		MaxTokens:       16384,
	},
}

// GetModelPricing returns pricing information for a model
func GetModelPricing(modelID string) (inputCost, outputCost float64, maxTokens int) {
	if pricing, ok := ModelPricing[modelID]; ok {
		return pricing.InputCostPer1K, pricing.OutputCostPer1K, pricing.MaxTokens
	}
	// Default to GPT-3.5-turbo pricing if model not found
	return 0.0015, 0.002, 4096
}
