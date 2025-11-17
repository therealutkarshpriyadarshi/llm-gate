package models

import "time"

// ChatResponse represents a unified chat completion response
type ChatResponse struct {
	// ID is the unique identifier for this completion
	ID string `json:"id"`

	// Object is the object type (e.g., "chat.completion")
	Object string `json:"object"`

	// Created is the Unix timestamp of when the completion was created
	Created int64 `json:"created"`

	// Model is the model used for completion
	Model string `json:"model"`

	// Choices is the list of completion choices
	Choices []Choice `json:"choices"`

	// Usage contains token usage information
	Usage Usage `json:"usage,omitempty"`

	// Provider is the name of the LLM provider that generated this response
	Provider string `json:"provider,omitempty"`

	// Metadata contains additional response metadata
	Metadata ResponseMetadata `json:"metadata,omitempty"`
}

// Choice represents a single completion choice
type Choice struct {
	// Index is the choice index
	Index int `json:"index"`

	// Message is the generated message
	Message Message `json:"message"`

	// FinishReason is the reason the model stopped generating tokens
	FinishReason string `json:"finish_reason"`
}

// Usage contains token usage statistics
type Usage struct {
	// PromptTokens is the number of tokens in the prompt
	PromptTokens int `json:"prompt_tokens"`

	// CompletionTokens is the number of tokens in the completion
	CompletionTokens int `json:"completion_tokens"`

	// TotalTokens is the total number of tokens
	TotalTokens int `json:"total_tokens"`
}

// ResponseMetadata contains additional information about the response
type ResponseMetadata struct {
	// RequestID links this response to the original request
	RequestID string `json:"request_id,omitempty"`

	// Latency is the time taken to generate the response
	Latency time.Duration `json:"latency,omitempty"`

	// CacheHit indicates if the response came from cache
	CacheHit bool `json:"cache_hit,omitempty"`

	// CacheSimilarity is the similarity score for semantic cache hits
	CacheSimilarity float64 `json:"cache_similarity,omitempty"`

	// Cost is the estimated cost of this request
	Cost float64 `json:"cost,omitempty"`

	// ProviderLatency is the time taken by the provider
	ProviderLatency time.Duration `json:"provider_latency,omitempty"`

	// Timestamp when the response was generated
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// StreamChunk represents a chunk in a streaming response
type StreamChunk struct {
	// ID is the unique identifier for this completion
	ID string `json:"id"`

	// Object is the object type (e.g., "chat.completion.chunk")
	Object string `json:"object"`

	// Created is the Unix timestamp of when the chunk was created
	Created int64 `json:"created"`

	// Model is the model used for completion
	Model string `json:"model"`

	// Choices is the list of completion choice deltas
	Choices []StreamChoice `json:"choices"`

	// Provider is the name of the LLM provider
	Provider string `json:"provider,omitempty"`
}

// StreamChoice represents a single completion choice in a stream
type StreamChoice struct {
	// Index is the choice index
	Index int `json:"index"`

	// Delta contains the incremental message update
	Delta MessageDelta `json:"delta"`

	// FinishReason is the reason the model stopped generating tokens
	FinishReason string `json:"finish_reason,omitempty"`
}

// MessageDelta represents an incremental update to a message
type MessageDelta struct {
	// Role is the role of the message sender (only present in first chunk)
	Role string `json:"role,omitempty"`

	// Content is the incremental content
	Content string `json:"content,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	// Error contains the error details
	Error ErrorDetail `json:"error"`

	// Provider is the name of the provider that returned the error
	Provider string `json:"provider,omitempty"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	// Message is the error message
	Message string `json:"message"`

	// Type is the error type
	Type string `json:"type"`

	// Code is the error code
	Code string `json:"code,omitempty"`

	// Param is the parameter that caused the error
	Param string `json:"param,omitempty"`
}
