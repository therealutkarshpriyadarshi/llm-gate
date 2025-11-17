package providers

import (
	"context"
	"testing"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/interfaces"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

// mockProvider is a mock implementation of LLMProvider for testing
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
	return models.ProviderCapabilities{
		Name: m.name,
	}
}

func (m *mockProvider) HealthCheck(ctx context.Context) models.ProviderHealth {
	return models.ProviderHealth{
		Name:    m.name,
		Healthy: true,
	}
}

func (m *mockProvider) GetModelInfo(modelID string) (models.ModelInfo, error) {
	return models.ModelInfo{}, nil
}

func (m *mockProvider) Close() error {
	return nil
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	provider := &mockProvider{name: "test"}
	err := registry.Register(provider)
	if err != nil {
		t.Errorf("Register() error = %v, want nil", err)
	}

	if registry.Count() != 1 {
		t.Errorf("Count() = %v, want 1", registry.Count())
	}
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	registry := NewRegistry()

	provider := &mockProvider{name: "test"}
	registry.Register(provider)

	// Try to register again
	err := registry.Register(provider)
	if err == nil {
		t.Error("Register() should return error for duplicate provider")
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()

	provider := &mockProvider{name: "test"}
	registry.Register(provider)

	retrieved, err := registry.Get("test")
	if err != nil {
		t.Errorf("Get() error = %v, want nil", err)
	}

	if retrieved.Name() != "test" {
		t.Errorf("Get() returned wrong provider: %v", retrieved.Name())
	}
}

func TestRegistry_GetNotFound(t *testing.T) {
	registry := NewRegistry()

	_, err := registry.Get("nonexistent")
	if err == nil {
		t.Error("Get() should return error for non-existent provider")
	}
}

func TestRegistry_Unregister(t *testing.T) {
	registry := NewRegistry()

	provider := &mockProvider{name: "test"}
	registry.Register(provider)

	err := registry.Unregister("test")
	if err != nil {
		t.Errorf("Unregister() error = %v, want nil", err)
	}

	if registry.Count() != 0 {
		t.Errorf("Count() = %v, want 0", registry.Count())
	}
}

func TestRegistry_GetAll(t *testing.T) {
	registry := NewRegistry()

	provider1 := &mockProvider{name: "test1"}
	provider2 := &mockProvider{name: "test2"}

	registry.Register(provider1)
	registry.Register(provider2)

	all := registry.GetAll()
	if len(all) != 2 {
		t.Errorf("GetAll() returned %v providers, want 2", len(all))
	}
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	provider1 := &mockProvider{name: "test1"}
	provider2 := &mockProvider{name: "test2"}

	registry.Register(provider1)
	registry.Register(provider2)

	types := registry.List()
	if len(types) != 2 {
		t.Errorf("List() returned %v types, want 2", len(types))
	}
}
