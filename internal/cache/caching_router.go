package cache

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/routing"
)

// CachingRouter wraps a router and adds semantic caching capabilities
type CachingRouter struct {
	router *routing.Router
	cache  Cache
	config *CacheConfig
}

// NewCachingRouter creates a new caching router
func NewCachingRouter(router *routing.Router, cache Cache, config *CacheConfig) *CachingRouter {
	return &CachingRouter{
		router: router,
		cache:  cache,
		config: config,
	}
}

// RouteWithFallback routes a request with caching and fallback
func (cr *CachingRouter) RouteWithFallback(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	startTime := time.Now()

	// Check if caching is enabled for this request
	if cr.shouldUseCache(req) {
		// Try to get from cache
		entry, err := cr.cache.Get(ctx, req)
		if err != nil {
			log.Error().Err(err).Msg("cache get failed")
			// Continue with normal routing on cache error
		} else if entry != nil {
			// Cache hit!
			response := entry.Response

			// Update metadata to indicate cache hit
			response.Metadata.CacheHit = true
			response.Metadata.CacheSimilarity = float64(entry.SimilarityScore)
			response.Metadata.Latency = time.Since(startTime)
			response.Metadata.Timestamp = time.Now()

			log.Info().
				Str("request_id", req.Metadata.RequestID).
				Float32("similarity", entry.SimilarityScore).
				Dur("latency", response.Metadata.Latency).
				Msg("cache hit - returning cached response")

			return response, nil
		}
	}

	// Cache miss or caching disabled - route to provider
	response, err := cr.router.RouteWithFallback(ctx, req)
	if err != nil {
		return nil, err
	}

	// Store response in cache if applicable
	if cr.shouldUseCache(req) && response != nil {
		ttl := cr.getCacheTTL(req)
		if err := cr.cache.Set(ctx, req, response, ttl); err != nil {
			log.Error().Err(err).Msg("failed to cache response")
			// Don't fail the request if caching fails
		} else {
			log.Debug().
				Str("request_id", req.Metadata.RequestID).
				Dur("ttl", ttl).
				Msg("response cached")
		}
	}

	// Update metadata
	response.Metadata.CacheHit = false
	response.Metadata.Latency = time.Since(startTime)

	return response, nil
}

// Route routes a request to a provider (for streaming)
func (cr *CachingRouter) Route(ctx context.Context, req *models.ChatRequest) (interface{}, error) {
	// For streaming requests, we typically don't cache
	// because we'd need to buffer the entire response
	// So we just pass through to the underlying router
	return cr.router.Route(ctx, req)
}

// GetStats returns cache statistics
func (cr *CachingRouter) GetStats(ctx context.Context) (*CacheStats, error) {
	return cr.cache.GetStats(ctx)
}

// ClearCache clears all cache entries
func (cr *CachingRouter) ClearCache(ctx context.Context) error {
	return cr.cache.Clear(ctx)
}

// DeleteCacheEntry deletes a specific cache entry
func (cr *CachingRouter) DeleteCacheEntry(ctx context.Context, key string) error {
	return cr.cache.Delete(ctx, key)
}

// shouldUseCache determines if caching should be used for this request
func (cr *CachingRouter) shouldUseCache(req *models.ChatRequest) bool {
	// Don't cache if caching is disabled globally
	if !cr.config.Enabled {
		return false
	}

	// Don't cache streaming requests
	if req.Stream {
		return false
	}

	// Check if the request explicitly disables caching
	if !req.Metadata.CacheEnabled && req.Metadata.CacheEnabled == false {
		return false
	}

	// Don't cache requests with high temperature (very random responses)
	if req.Temperature != nil && *req.Temperature > 1.5 {
		return false
	}

	return true
}

// getCacheTTL determines the TTL for caching this request
func (cr *CachingRouter) getCacheTTL(req *models.ChatRequest) time.Duration {
	// If request specifies a TTL, use it
	if req.Metadata.CacheTTL > 0 {
		return req.Metadata.CacheTTL
	}

	// Otherwise use the default TTL
	return cr.config.DefaultTTL
}

// Close closes the cache connection
func (cr *CachingRouter) Close() error {
	return cr.cache.Close()
}

// GetHealthStatus returns provider health status from the underlying router
func (cr *CachingRouter) GetHealthStatus() map[models.ProviderType]*routing.HealthStatus {
	return cr.router.GetHealthStatus()
}
