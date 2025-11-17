package interfaces

import (
	"context"
	"io"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

// LLMProvider defines the interface that all LLM providers must implement
type LLMProvider interface {
	// Name returns the provider name
	Name() models.ProviderType

	// ChatCompletion sends a chat completion request and returns a response
	ChatCompletion(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error)

	// ChatCompletionStream sends a streaming chat completion request
	ChatCompletionStream(ctx context.Context, req *models.ChatRequest) (StreamReader, error)

	// GetCapabilities returns the provider's capabilities
	GetCapabilities() models.ProviderCapabilities

	// HealthCheck checks if the provider is healthy
	HealthCheck(ctx context.Context) models.ProviderHealth

	// GetModelInfo returns information about a specific model
	GetModelInfo(modelID string) (models.ModelInfo, error)

	// Close closes the provider and cleans up resources
	Close() error
}

// StreamReader defines the interface for reading streaming responses
type StreamReader interface {
	// Recv receives the next chunk from the stream
	Recv() (*models.StreamChunk, error)

	// Close closes the stream
	Close() error
}

// ProviderFactory creates provider instances
type ProviderFactory interface {
	// CreateProvider creates a new provider instance
	CreateProvider(providerType models.ProviderType, config interface{}) (LLMProvider, error)
}

// StreamWriter defines the interface for writing streaming responses
type StreamWriter interface {
	io.Writer
	Flush() error
}
