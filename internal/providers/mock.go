package providers

import (
	"context"
	"time"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/interfaces"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

// MockProvider is a mock LLM provider for testing
type MockProvider struct {
	name         models.ProviderType
	model        string
	responseFunc func(context.Context, *models.ChatRequest) (*models.ChatResponse, error)
	healthy      bool
	latency      time.Duration
}

// NewMockProvider creates a new mock provider
func NewMockProvider(name models.ProviderType, model string, responseFunc func(context.Context, *models.ChatRequest) (*models.ChatResponse, error)) *MockProvider {
	return &MockProvider{
		name:         name,
		model:        model,
		responseFunc: responseFunc,
		healthy:      true,
		latency:      100 * time.Millisecond,
	}
}

// Name returns the provider name
func (m *MockProvider) Name() models.ProviderType {
	return m.name
}

// ChatCompletion performs a chat completion
func (m *MockProvider) ChatCompletion(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	// Simulate latency
	time.Sleep(m.latency)

	if m.responseFunc != nil {
		return m.responseFunc(ctx, req)
	}

	// Default response
	return &models.ChatResponse{
		ID:      "mock-response",
		Model:   m.model,
		Created: time.Now().Unix(),
		Choices: []models.Choice{
			{
				Index: 0,
				Message: models.Message{
					Role:    "assistant",
					Content: "This is a mock response",
				},
				FinishReason: "stop",
			},
		},
		Usage: models.Usage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}, nil
}

// ChatCompletionStream performs a streaming chat completion
func (m *MockProvider) ChatCompletionStream(ctx context.Context, req *models.ChatRequest) (interfaces.StreamReader, error) {
	return &mockStreamReader{
		chunks: []*models.StreamChunk{
			{
				ID:      "mock-chunk",
				Model:   m.model,
				Created: time.Now().Unix(),
				Choices: []models.StreamChoice{
					{
						Index: 0,
						Delta: models.MessageDelta{
							Role:    "assistant",
							Content: "Mock response",
						},
					},
				},
			},
		},
		index: 0,
	}, nil
}

// GetCapabilities returns the provider's capabilities
func (m *MockProvider) GetCapabilities() models.ProviderCapabilities {
	return models.ProviderCapabilities{
		Name:              m.name,
		SupportsStreaming: true,
		SupportsFunctions: true,
		SupportsVision:    false,
		MaxTokens:         8192,
	}
}

// HealthCheck performs a health check
func (m *MockProvider) HealthCheck(ctx context.Context) models.ProviderHealth {
	return models.ProviderHealth{
		Name:      m.name,
		Healthy:   m.healthy,
		Latency:   m.latency,
		LastCheck: time.Now(),
	}
}

// GetModelInfo returns information about a specific model
func (m *MockProvider) GetModelInfo(modelID string) (models.ModelInfo, error) {
	return models.ModelInfo{
		ID:                modelID,
		Name:              modelID,
		MaxTokens:         8192,
		InputCostPer1K:    0.01,
		OutputCostPer1K:   0.03,
		SupportsStreaming: true,
		SupportsFunctions: true,
	}, nil
}

// Close closes the provider
func (m *MockProvider) Close() error {
	return nil
}

// SetHealthy sets the health status
func (m *MockProvider) SetHealthy(healthy bool) {
	m.healthy = healthy
}

// SetLatency sets the latency
func (m *MockProvider) SetLatency(latency time.Duration) {
	m.latency = latency
}

// mockStreamReader implements interfaces.StreamReader for testing
type mockStreamReader struct {
	chunks []*models.StreamChunk
	index  int
}

// Recv receives the next chunk from the stream
func (r *mockStreamReader) Recv() (*models.StreamChunk, error) {
	if r.index >= len(r.chunks) {
		return nil, nil // EOF
	}
	chunk := r.chunks[r.index]
	r.index++
	return chunk, nil
}

// Close closes the stream
func (r *mockStreamReader) Close() error {
	return nil
}

// Ensure MockProvider implements LLMProvider
var _ interfaces.LLMProvider = (*MockProvider)(nil)
