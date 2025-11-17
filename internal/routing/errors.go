package routing

import "errors"

var (
	// ErrNoProvidersAvailable is returned when no providers are available
	ErrNoProvidersAvailable = errors.New("no providers available")

	// ErrNoHealthyProviders is returned when no healthy providers are available
	ErrNoHealthyProviders = errors.New("no healthy providers available")

	// ErrAllProvidersFailed is returned when all providers failed to handle the request
	ErrAllProvidersFailed = errors.New("all providers failed")

	// ErrProviderNotFound is returned when a specific provider is not found
	ErrProviderNotFound = errors.New("provider not found")
)
