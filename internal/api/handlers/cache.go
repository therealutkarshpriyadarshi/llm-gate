package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/cache"
)

// CacheHandler handles cache management requests
type CacheHandler struct {
	cachingRouter *cache.CachingRouter
}

// NewCacheHandler creates a new cache handler
func NewCacheHandler(cachingRouter *cache.CachingRouter) *CacheHandler {
	return &CacheHandler{
		cachingRouter: cachingRouter,
	}
}

// HandleGetStats handles GET /v1/cache/stats
func (h *CacheHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := h.cachingRouter.GetStats(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get cache stats")
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve cache stats")
		return
	}

	log.Info().
		Int64("total_entries", stats.TotalEntries).
		Int64("total_hits", stats.TotalHits).
		Int64("total_misses", stats.TotalMisses).
		Float64("hit_rate", stats.HitRate).
		Msg("cache stats retrieved")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}

// HandleClearCache handles DELETE /v1/cache
func (h *CacheHandler) HandleClearCache(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.cachingRouter.ClearCache(ctx); err != nil {
		log.Error().Err(err).Msg("failed to clear cache")
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to clear cache")
		return
	}

	log.Info().Msg("cache cleared")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Cache cleared successfully",
	})
}

// HandleDeleteCacheEntry handles DELETE /v1/cache/{key}
func (h *CacheHandler) HandleDeleteCacheEntry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := chi.URLParam(r, "key")

	if key == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Cache key is required")
		return
	}

	if err := h.cachingRouter.DeleteCacheEntry(ctx, key); err != nil {
		log.Error().
			Err(err).
			Str("key", key).
			Msg("failed to delete cache entry")
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to delete cache entry")
		return
	}

	log.Info().Str("key", key).Msg("cache entry deleted")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Cache entry deleted successfully",
		"key":     key,
	})
}

// writeError writes an error response
func (h *CacheHandler) writeError(w http.ResponseWriter, status int, errType, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := map[string]interface{}{
		"error": map[string]interface{}{
			"type":    errType,
			"message": message,
		},
	}

	json.NewEncoder(w).Encode(resp)
}
