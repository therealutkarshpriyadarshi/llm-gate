package vertex

// VertexMessage represents a message in Vertex AI format
type VertexMessage struct {
	Role    string        `json:"role"`
	Parts   []VertexPart  `json:"parts"`
}

// VertexPart represents a part of a message
type VertexPart struct {
	Text string `json:"text"`
}

// VertexRequest represents a Vertex AI API request
type VertexRequest struct {
	Contents         []VertexMessage      `json:"contents"`
	GenerationConfig *VertexGenConfig     `json:"generationConfig,omitempty"`
	SafetySettings   []VertexSafety       `json:"safetySettings,omitempty"`
}

// VertexGenConfig represents generation configuration
type VertexGenConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
	TopK            *int     `json:"topK,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

// VertexSafety represents safety settings
type VertexSafety struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

// VertexResponse represents a Vertex AI API response
type VertexResponse struct {
	Candidates     []VertexCandidate   `json:"candidates"`
	UsageMetadata  VertexUsageMetadata `json:"usageMetadata,omitempty"`
}

// VertexCandidate represents a response candidate
type VertexCandidate struct {
	Content      VertexContent `json:"content"`
	FinishReason string        `json:"finishReason"`
	Index        int           `json:"index"`
}

// VertexContent represents response content
type VertexContent struct {
	Role  string       `json:"role"`
	Parts []VertexPart `json:"parts"`
}

// VertexUsageMetadata represents token usage
type VertexUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// VertexStreamChunk represents a streaming response chunk
type VertexStreamChunk struct {
	Candidates    []VertexCandidate   `json:"candidates"`
	UsageMetadata VertexUsageMetadata `json:"usageMetadata,omitempty"`
}

// VertexError represents an error from the Vertex AI API
type VertexError struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// ModelPricing maps Vertex AI model IDs to their pricing
var ModelPricing = map[string]struct {
	InputCostPer1K  float64
	OutputCostPer1K float64
	MaxTokens       int
}{
	// Gemini Pro models
	"gemini-pro": {
		InputCostPer1K:  0.000125,
		OutputCostPer1K: 0.000375,
		MaxTokens:       32760,
	},
	"gemini-pro-vision": {
		InputCostPer1K:  0.000125,
		OutputCostPer1K: 0.000375,
		MaxTokens:       16384,
	},
	"gemini-1.5-pro": {
		InputCostPer1K:  0.00125,
		OutputCostPer1K: 0.00375,
		MaxTokens:       1048576, // 1M tokens
	},
	"gemini-1.5-flash": {
		InputCostPer1K:  0.000075,
		OutputCostPer1K: 0.0003,
		MaxTokens:       1048576,
	},
	// PaLM models
	"text-bison": {
		InputCostPer1K:  0.000125,
		OutputCostPer1K: 0.000125,
		MaxTokens:       8196,
	},
	"text-bison-32k": {
		InputCostPer1K:  0.000125,
		OutputCostPer1K: 0.000125,
		MaxTokens:       32000,
	},
	"chat-bison": {
		InputCostPer1K:  0.000125,
		OutputCostPer1K: 0.000125,
		MaxTokens:       8196,
	},
	"chat-bison-32k": {
		InputCostPer1K:  0.000125,
		OutputCostPer1K: 0.000125,
		MaxTokens:       32000,
	},
}

// GetModelPricing returns pricing information for a model
func GetModelPricing(modelID string) (inputCost, outputCost float64, maxTokens int) {
	if pricing, ok := ModelPricing[modelID]; ok {
		return pricing.InputCostPer1K, pricing.OutputCostPer1K, pricing.MaxTokens
	}
	// Default to Gemini Pro pricing if model not found
	return 0.000125, 0.000375, 32760
}
