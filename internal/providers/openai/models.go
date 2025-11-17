package openai

// OpenAI-specific request/response models
// These models match the OpenAI API specification

// OpenAIChatRequest represents an OpenAI chat completion request
type OpenAIChatRequest struct {
	Model            string         `json:"model"`
	Messages         []OpenAIMessage `json:"messages"`
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

// OpenAIMessage represents a message in the OpenAI format
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// OpenAIChatResponse represents an OpenAI chat completion response
type OpenAIChatResponse struct {
	ID      string              `json:"id"`
	Object  string              `json:"object"`
	Created int64               `json:"created"`
	Model   string              `json:"model"`
	Choices []OpenAIChoice      `json:"choices"`
	Usage   OpenAIUsage         `json:"usage"`
}

// OpenAIChoice represents a completion choice
type OpenAIChoice struct {
	Index        int           `json:"index"`
	Message      OpenAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// OpenAIStreamChunk represents a streaming response chunk
type OpenAIStreamChunk struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []OpenAIStreamChoice   `json:"choices"`
}

// OpenAIStreamChoice represents a streaming choice
type OpenAIStreamChoice struct {
	Index        int                `json:"index"`
	Delta        OpenAIMessageDelta `json:"delta"`
	FinishReason string             `json:"finish_reason,omitempty"`
}

// OpenAIMessageDelta represents an incremental message update
type OpenAIMessageDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// OpenAIUsage represents token usage information
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAIError represents an OpenAI API error
type OpenAIError struct {
	Error OpenAIErrorDetail `json:"error"`
}

// OpenAIErrorDetail contains error details
type OpenAIErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
	Param   string `json:"param,omitempty"`
}

// ModelPricing contains pricing information for OpenAI models
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
	"gpt-4-turbo-preview": {
		InputCostPer1K:  0.01,
		OutputCostPer1K: 0.03,
		MaxTokens:       128000,
	},
	"gpt-3.5-turbo": {
		InputCostPer1K:  0.0005,
		OutputCostPer1K: 0.0015,
		MaxTokens:       16385,
	},
	"gpt-3.5-turbo-16k": {
		InputCostPer1K:  0.003,
		OutputCostPer1K: 0.004,
		MaxTokens:       16385,
	},
}

// GetModelPricing returns pricing information for a model
func GetModelPricing(model string) (inputCost, outputCost float64, maxTokens int) {
	if pricing, ok := ModelPricing[model]; ok {
		return pricing.InputCostPer1K, pricing.OutputCostPer1K, pricing.MaxTokens
	}
	// Default to GPT-3.5-turbo pricing if model not found
	return 0.0005, 0.0015, 4096
}
