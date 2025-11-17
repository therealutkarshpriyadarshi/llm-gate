package routing

import (
	"context"
	"testing"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/interfaces"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

// mockProvider is a mock implementation for testing
type mockProvider struct {
	name models.ProviderType
}

func (m *mockProvider) Name() models.ProviderType {
	return m.name
}

func (m *mockProvider) ChatCompletion(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	return &models.ChatResponse{}, nil
}

func (m *mockProvider) ChatCompletionStream(ctx context.Context, req *models.ChatRequest) (interfaces.StreamReader, error) {
	return nil, nil
}

func (m *mockProvider) GetCapabilities() models.ProviderCapabilities {
	return models.ProviderCapabilities{Name: m.name}
}

func (m *mockProvider) HealthCheck(ctx context.Context) models.ProviderHealth {
	return models.ProviderHealth{Name: m.name, Healthy: true}
}

func (m *mockProvider) GetModelInfo(modelID string) (models.ModelInfo, error) {
	return models.ModelInfo{}, nil
}

func (m *mockProvider) Close() error {
	return nil
}

func TestRoundRobinStrategy_SelectProvider(t *testing.T) {
	strategy := NewRoundRobinStrategy()

	providers := []interfaces.LLMProvider{
		&mockProvider{name: "provider1"},
		&mockProvider{name: "provider2"},
		&mockProvider{name: "provider3"},
	}

	req := &models.ChatRequest{
		Model: "test-model",
		Messages: []models.Message{
			{Role: "user", Content: "test"},
		},
	}

	// Test round-robin distribution
	selected := make(map[models.ProviderType]int)
	for i := 0; i < 9; i++ {
		provider, err := strategy.SelectProvider(providers, req)
		if err != nil {
			t.Fatalf("SelectProvider() error = %v", err)
		}
		selected[provider.Name()]++
	}

	// Each provider should be selected 3 times
	for name, count := range selected {
		if count != 3 {
			t.Errorf("Provider %s selected %d times, want 3", name, count)
		}
	}
}

func TestRoundRobinStrategy_NoProviders(t *testing.T) {
	strategy := NewRoundRobinStrategy()

	req := &models.ChatRequest{
		Model: "test-model",
		Messages: []models.Message{
			{Role: "user", Content: "test"},
		},
	}

	_, err := strategy.SelectProvider([]interfaces.LLMProvider{}, req)
	if err == nil {
		t.Error("SelectProvider() should return error for empty providers list")
	}
}

func TestRoundRobinStrategy_Name(t *testing.T) {
	strategy := NewRoundRobinStrategy()
	if strategy.Name() != "round-robin" {
		t.Errorf("Name() = %v, want 'round-robin'", strategy.Name())
	}
}
