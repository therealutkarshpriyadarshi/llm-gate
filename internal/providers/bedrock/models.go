package bedrock

// BedrockMessage represents a message in Bedrock format (for Claude models)
type BedrockMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// BedrockRequest represents a Bedrock API request
type BedrockRequest struct {
	// For Claude models on Bedrock
	AnthropicVersion string            `json:"anthropic_version,omitempty"`
	Messages         []BedrockMessage  `json:"messages,omitempty"`
	MaxTokens        int               `json:"max_tokens,omitempty"`
	Temperature      *float64          `json:"temperature,omitempty"`
	TopP             *float64          `json:"top_p,omitempty"`
	TopK             *int              `json:"top_k,omitempty"`
	StopSequences    []string          `json:"stop_sequences,omitempty"`
	System           string            `json:"system,omitempty"`
}

// BedrockResponse represents a Bedrock API response
type BedrockResponse struct {
	// For Claude models on Bedrock
	ID           string                  `json:"id"`
	Type         string                  `json:"type"`
	Role         string                  `json:"role"`
	Content      []BedrockContentBlock   `json:"content"`
	Model        string                  `json:"model,omitempty"`
	StopReason   string                  `json:"stop_reason"`
	StopSequence string                  `json:"stop_sequence,omitempty"`
	Usage        BedrockUsage            `json:"usage"`
}

// BedrockContentBlock represents a content block
type BedrockContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// BedrockUsage represents token usage
type BedrockUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// BedrockStreamChunk represents a streaming response chunk
type BedrockStreamChunk struct {
	Type         string                 `json:"type"`
	Message      *BedrockResponse       `json:"message,omitempty"`
	Index        int                    `json:"index,omitempty"`
	ContentBlock *BedrockContentBlock   `json:"content_block,omitempty"`
	Delta        *BedrockDelta          `json:"delta,omitempty"`
	Usage        *BedrockUsage          `json:"usage,omitempty"`
}

// BedrockDelta represents incremental content
type BedrockDelta struct {
	Type         string `json:"type"`
	Text         string `json:"text,omitempty"`
	StopReason   string `json:"stop_reason,omitempty"`
	StopSequence string `json:"stop_sequence,omitempty"`
}

// BedrockError represents an error from the Bedrock API
type BedrockError struct {
	Message string `json:"message"`
	Type    string `json:"__type"`
}

// ModelPricing maps Bedrock model IDs to their pricing
var ModelPricing = map[string]struct {
	InputCostPer1K  float64
	OutputCostPer1K float64
	MaxTokens       int
}{
	"anthropic.claude-3-opus-20240229-v1:0": {
		InputCostPer1K:  0.015,
		OutputCostPer1K: 0.075,
		MaxTokens:       200000,
	},
	"anthropic.claude-3-sonnet-20240229-v1:0": {
		InputCostPer1K:  0.003,
		OutputCostPer1K: 0.015,
		MaxTokens:       200000,
	},
	"anthropic.claude-3-haiku-20240307-v1:0": {
		InputCostPer1K:  0.00025,
		OutputCostPer1K: 0.00125,
		MaxTokens:       200000,
	},
	"anthropic.claude-v2:1": {
		InputCostPer1K:  0.008,
		OutputCostPer1K: 0.024,
		MaxTokens:       100000,
	},
	"anthropic.claude-v2": {
		InputCostPer1K:  0.008,
		OutputCostPer1K: 0.024,
		MaxTokens:       100000,
	},
	"anthropic.claude-instant-v1": {
		InputCostPer1K:  0.00163,
		OutputCostPer1K: 0.00551,
		MaxTokens:       100000,
	},
	// Meta Llama models
	"meta.llama2-70b-chat-v1": {
		InputCostPer1K:  0.00195,
		OutputCostPer1K: 0.00256,
		MaxTokens:       4096,
	},
	"meta.llama2-13b-chat-v1": {
		InputCostPer1K:  0.00075,
		OutputCostPer1K: 0.001,
		MaxTokens:       4096,
	},
	// Amazon Titan models
	"amazon.titan-text-express-v1": {
		InputCostPer1K:  0.0002,
		OutputCostPer1K: 0.0006,
		MaxTokens:       8000,
	},
	"amazon.titan-text-lite-v1": {
		InputCostPer1K:  0.00015,
		OutputCostPer1K: 0.0002,
		MaxTokens:       4000,
	},
}

// GetModelPricing returns pricing information for a model
func GetModelPricing(modelID string) (inputCost, outputCost float64, maxTokens int) {
	if pricing, ok := ModelPricing[modelID]; ok {
		return pricing.InputCostPer1K, pricing.OutputCostPer1K, pricing.MaxTokens
	}
	// Default to Claude 3 Haiku pricing if model not found
	return 0.00025, 0.00125, 200000
}
