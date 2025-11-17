package providers

import (
	"fmt"
	"sync"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/interfaces"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

// Registry manages multiple LLM providers
type Registry struct {
	mu        sync.RWMutex
	providers map[models.ProviderType]interfaces.LLMProvider
	factory   *Factory
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[models.ProviderType]interfaces.LLMProvider),
		factory:   NewFactory(),
	}
}

// Register registers a provider
func (r *Registry) Register(provider interfaces.LLMProvider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	providerType := provider.Name()
	if _, exists := r.providers[providerType]; exists {
		return fmt.Errorf("provider %s already registered", providerType)
	}

	r.providers[providerType] = provider
	return nil
}

// RegisterWithConfig registers a provider using configuration
func (r *Registry) RegisterWithConfig(providerType models.ProviderType, config interface{}) error {
	provider, err := r.factory.CreateProvider(providerType, config)
	if err != nil {
		return fmt.Errorf("failed to create provider %s: %w", providerType, err)
	}

	return r.Register(provider)
}

// Unregister removes a provider from the registry
func (r *Registry) Unregister(providerType models.ProviderType) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	provider, exists := r.providers[providerType]
	if !exists {
		return fmt.Errorf("provider %s not found", providerType)
	}

	// Close the provider
	if err := provider.Close(); err != nil {
		return fmt.Errorf("failed to close provider %s: %w", providerType, err)
	}

	delete(r.providers, providerType)
	return nil
}

// Get retrieves a provider by type
func (r *Registry) Get(providerType models.ProviderType) (interfaces.LLMProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[providerType]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", providerType)
	}

	return provider, nil
}

// GetAll returns all registered providers
func (r *Registry) GetAll() []interfaces.LLMProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]interfaces.LLMProvider, 0, len(r.providers))
	for _, provider := range r.providers {
		providers = append(providers, provider)
	}

	return providers
}

// List returns a list of all registered provider types
func (r *Registry) List() []models.ProviderType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]models.ProviderType, 0, len(r.providers))
	for providerType := range r.providers {
		types = append(types, providerType)
	}

	return types
}

// Count returns the number of registered providers
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.providers)
}

// Close closes all registered providers
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errors []error
	for providerType, provider := range r.providers {
		if err := provider.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close provider %s: %w", providerType, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to close %d providers: %v", len(errors), errors)
	}

	r.providers = make(map[models.ProviderType]interfaces.LLMProvider)
	return nil
}
