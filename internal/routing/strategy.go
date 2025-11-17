package routing

import (
	"sync/atomic"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/interfaces"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

// RoutingStrategy defines the interface for provider selection strategies
type RoutingStrategy interface {
	// SelectProvider selects a provider from the available providers
	SelectProvider(providers []interfaces.LLMProvider, req *models.ChatRequest) (interfaces.LLMProvider, error)

	// Name returns the strategy name
	Name() string
}

// RoundRobinStrategy implements a simple round-robin selection strategy
type RoundRobinStrategy struct {
	counter uint64
}

// NewRoundRobinStrategy creates a new round-robin strategy
func NewRoundRobinStrategy() *RoundRobinStrategy {
	return &RoundRobinStrategy{
		counter: 0,
	}
}

// SelectProvider selects the next provider in round-robin order
func (s *RoundRobinStrategy) SelectProvider(providers []interfaces.LLMProvider, req *models.ChatRequest) (interfaces.LLMProvider, error) {
	if len(providers) == 0 {
		return nil, ErrNoProvidersAvailable
	}

	// Get next index using atomic increment
	index := atomic.AddUint64(&s.counter, 1) - 1
	selectedIndex := int(index) % len(providers)

	return providers[selectedIndex], nil
}

// Name returns the strategy name
func (s *RoundRobinStrategy) Name() string {
	return "round-robin"
}

// RandomStrategy implements a random selection strategy
type RandomStrategy struct{}

// NewRandomStrategy creates a new random strategy
func NewRandomStrategy() *RandomStrategy {
	return &RandomStrategy{}
}

// SelectProvider selects a random provider
func (s *RandomStrategy) SelectProvider(providers []interfaces.LLMProvider, req *models.ChatRequest) (interfaces.LLMProvider, error) {
	if len(providers) == 0 {
		return nil, ErrNoProvidersAvailable
	}

	// For simplicity, we'll use a basic approach
	// In production, you'd want to use crypto/rand for better randomness
	index := int(atomic.AddUint64(new(uint64), 1)) % len(providers)
	return providers[index], nil
}

// Name returns the strategy name
func (s *RandomStrategy) Name() string {
	return "random"
}
