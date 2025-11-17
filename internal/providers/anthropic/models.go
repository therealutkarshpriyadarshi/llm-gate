package anthropic

// AnthropicMessage represents a message in Anthropic format
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicRequest represents an Anthropic API request
type AnthropicRequest struct {
	Model       string              `json:"model"`
	Messages    []AnthropicMessage  `json:"messages"`
	MaxTokens   int                 `json:"max_tokens"`
	Temperature *float64            `json:"temperature,omitempty"`
	TopP        *float64            `json:"top_p,omitempty"`
	TopK        *int                `json:"top_k,omitempty"`
	Stream      bool                `json:"stream,omitempty"`
	StopSeqs    []string            `json:"stop_sequences,omitempty"`
	System      string              `json:"system,omitempty"`
	Metadata    map[string]string   `json:"metadata,omitempty"`
}

// AnthropicResponse represents an Anthropic API response
type AnthropicResponse struct {
	ID           string                   `json:"id"`
	Type         string                   `json:"type"`
	Role         string                   `json:"role"`
	Content      []AnthropicContentBlock  `json:"content"`
	Model        string                   `json:"model"`
	StopReason   string                   `json:"stop_reason"`
	StopSequence string                   `json:"stop_sequence,omitempty"`
	Usage        AnthropicUsage           `json:"usage"`
}

// AnthropicContentBlock represents a content block in Anthropic response
type AnthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// AnthropicUsage represents token usage
type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// AnthropicStreamEvent represents a server-sent event
type AnthropicStreamEvent struct {
	Type         string                  `json:"type"`
	Message      *AnthropicResponse      `json:"message,omitempty"`
	Index        int                     `json:"index,omitempty"`
	ContentBlock *AnthropicContentBlock  `json:"content_block,omitempty"`
	Delta        *AnthropicDelta         `json:"delta,omitempty"`
	Usage        *AnthropicUsage         `json:"usage,omitempty"`
}

// AnthropicDelta represents incremental content updates
type AnthropicDelta struct {
	Type         string `json:"type"`
	Text         string `json:"text,omitempty"`
	StopReason   string `json:"stop_reason,omitempty"`
	StopSequence string `json:"stop_sequence,omitempty"`
}

// AnthropicError represents an error from the Anthropic API
type AnthropicError struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// ModelPricing maps Anthropic model IDs to their pricing and capabilities
var ModelPricing = map[string]struct {
	InputCostPer1K  float64
	OutputCostPer1K float64
	MaxTokens       int
}{
	"claude-3-opus-20240229": {
		InputCostPer1K:  0.015,  // $15 per 1M tokens
		OutputCostPer1K: 0.075,  // $75 per 1M tokens
		MaxTokens:       200000,
	},
	"claude-3-sonnet-20240229": {
		InputCostPer1K:  0.003,  // $3 per 1M tokens
		OutputCostPer1K: 0.015,  // $15 per 1M tokens
		MaxTokens:       200000,
	},
	"claude-3-haiku-20240307": {
		InputCostPer1K:  0.00025, // $0.25 per 1M tokens
		OutputCostPer1K: 0.00125, // $1.25 per 1M tokens
		MaxTokens:       200000,
	},
	"claude-2.1": {
		InputCostPer1K:  0.008,  // $8 per 1M tokens
		OutputCostPer1K: 0.024,  // $24 per 1M tokens
		MaxTokens:       100000,
	},
	"claude-2.0": {
		InputCostPer1K:  0.008,  // $8 per 1M tokens
		OutputCostPer1K: 0.024,  // $24 per 1M tokens
		MaxTokens:       100000,
	},
	"claude-instant-1.2": {
		InputCostPer1K:  0.00163, // $1.63 per 1M tokens
		OutputCostPer1K: 0.00551, // $5.51 per 1M tokens
		MaxTokens:       100000,
	},
}

// GetModelPricing returns pricing information for a model
func GetModelPricing(modelID string) (inputCost, outputCost float64, maxTokens int) {
	if pricing, ok := ModelPricing[modelID]; ok {
		return pricing.InputCostPer1K, pricing.OutputCostPer1K, pricing.MaxTokens
	}
	// Default to Claude 3 Sonnet pricing if model not found
	return 0.003, 0.015, 200000
}
