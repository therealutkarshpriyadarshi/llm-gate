package models

import "time"

// ChatRequest represents a unified chat completion request
type ChatRequest struct {
	// Model is the name of the model to use
	Model string `json:"model"`

	// Messages is the array of conversation messages
	Messages []Message `json:"messages"`

	// Temperature controls randomness (0.0 to 2.0)
	Temperature *float64 `json:"temperature,omitempty"`

	// MaxTokens is the maximum number of tokens to generate
	MaxTokens *int `json:"max_tokens,omitempty"`

	// TopP controls nucleus sampling (0.0 to 1.0)
	TopP *float64 `json:"top_p,omitempty"`

	// N is the number of completions to generate
	N *int `json:"n,omitempty"`

	// Stream indicates whether to stream the response
	Stream bool `json:"stream,omitempty"`

	// Stop sequences where the API will stop generating
	Stop []string `json:"stop,omitempty"`

	// PresencePenalty penalizes new tokens based on whether they appear in the text (-2.0 to 2.0)
	PresencePenalty *float64 `json:"presence_penalty,omitempty"`

	// FrequencyPenalty penalizes new tokens based on their frequency in the text (-2.0 to 2.0)
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`

	// User is a unique identifier for the end-user
	User string `json:"user,omitempty"`

	// Metadata for tracking and analytics
	Metadata RequestMetadata `json:"metadata,omitempty"`
}

// Message represents a single message in the conversation
type Message struct {
	// Role is the role of the message sender (system, user, assistant)
	Role string `json:"role"`

	// Content is the message content
	Content string `json:"content"`

	// Name is an optional name for the participant
	Name string `json:"name,omitempty"`
}

// RequestMetadata contains additional information about the request
type RequestMetadata struct {
	// RequestID is a unique identifier for this request
	RequestID string `json:"request_id,omitempty"`

	// UserID is the identifier of the user making the request
	UserID string `json:"user_id,omitempty"`

	// TenantID is the identifier of the tenant
	TenantID string `json:"tenant_id,omitempty"`

	// Tags for categorization and filtering
	Tags []string `json:"tags,omitempty"`

	// Timestamp when the request was created
	Timestamp time.Time `json:"timestamp,omitempty"`

	// CacheEnabled indicates if caching should be used
	CacheEnabled bool `json:"cache_enabled,omitempty"`

	// CacheTTL is the cache time-to-live
	CacheTTL time.Duration `json:"cache_ttl,omitempty"`
}

// Validate validates the chat request
func (r *ChatRequest) Validate() error {
	if r.Model == "" {
		return ErrInvalidModel
	}

	if len(r.Messages) == 0 {
		return ErrNoMessages
	}

	for i, msg := range r.Messages {
		if msg.Role == "" {
			return NewValidationError("message", i, "role cannot be empty")
		}
		if msg.Content == "" {
			return NewValidationError("message", i, "content cannot be empty")
		}
		if msg.Role != "system" && msg.Role != "user" && msg.Role != "assistant" {
			return NewValidationError("message", i, "invalid role: "+msg.Role)
		}
	}

	if r.Temperature != nil && (*r.Temperature < 0 || *r.Temperature > 2.0) {
		return ErrInvalidTemperature
	}

	if r.TopP != nil && (*r.TopP < 0 || *r.TopP > 1.0) {
		return ErrInvalidTopP
	}

	if r.PresencePenalty != nil && (*r.PresencePenalty < -2.0 || *r.PresencePenalty > 2.0) {
		return ErrInvalidPresencePenalty
	}

	if r.FrequencyPenalty != nil && (*r.FrequencyPenalty < -2.0 || *r.FrequencyPenalty > 2.0) {
		return ErrInvalidFrequencyPenalty
	}

	return nil
}

// GetTemperature returns the temperature value or default (1.0)
func (r *ChatRequest) GetTemperature() float64 {
	if r.Temperature != nil {
		return *r.Temperature
	}
	return 1.0
}

// GetMaxTokens returns the max tokens value or 0 (unlimited)
func (r *ChatRequest) GetMaxTokens() int {
	if r.MaxTokens != nil {
		return *r.MaxTokens
	}
	return 0
}

// GetTopP returns the top_p value or default (1.0)
func (r *ChatRequest) GetTopP() float64 {
	if r.TopP != nil {
		return *r.TopP
	}
	return 1.0
}

// GetN returns the number of completions or default (1)
func (r *ChatRequest) GetN() int {
	if r.N != nil {
		return *r.N
	}
	return 1
}
