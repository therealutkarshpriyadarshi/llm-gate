package cache

import (
	"context"
	"time"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

// Cache defines the interface for semantic caching operations
type Cache interface {
	// Get retrieves a cached response for the given request
	// Returns nil if no match is found or similarity is below threshold
	Get(ctx context.Context, req *models.ChatRequest) (*CacheEntry, error)

	// Set stores a request-response pair in the cache
	Set(ctx context.Context, req *models.ChatRequest, resp *models.ChatResponse, ttl time.Duration) error

	// Delete removes a cache entry
	Delete(ctx context.Context, key string) error

	// Clear removes all cache entries
	Clear(ctx context.Context) error

	// GetStats returns cache statistics
	GetStats(ctx context.Context) (*CacheStats, error)

	// Close closes the cache connection
	Close() error
}

// Embedder defines the interface for generating embeddings
type Embedder interface {
	// Embed generates an embedding vector for the given text
	Embed(ctx context.Context, text string) ([]float32, error)

	// EmbedBatch generates embeddings for multiple texts
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
}

// CacheEntry represents a cached request-response pair
type CacheEntry struct {
	Key             string                 `json:"key"`
	Request         *models.ChatRequest    `json:"request"`
	Response        *models.ChatResponse   `json:"response"`
	Embedding       []float32              `json:"embedding"`
	SimilarityScore float32                `json:"similarity_score"`
	CreatedAt       time.Time              `json:"created_at"`
	ExpiresAt       time.Time              `json:"expires_at"`
	HitCount        int64                  `json:"hit_count"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// CacheStats represents cache performance metrics
type CacheStats struct {
	TotalEntries    int64   `json:"total_entries"`
	TotalHits       int64   `json:"total_hits"`
	TotalMisses     int64   `json:"total_misses"`
	HitRate         float64 `json:"hit_rate"`
	AvgSimilarity   float64 `json:"avg_similarity"`
	CacheSize       int64   `json:"cache_size_bytes"`
	OldestEntry     time.Time `json:"oldest_entry"`
	NewestEntry     time.Time `json:"newest_entry"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	// Enabled indicates if caching is enabled
	Enabled bool

	// RedisAddr is the Redis server address
	RedisAddr string

	// RedisPassword is the Redis password
	RedisPassword string

	// RedisDB is the Redis database number
	RedisDB int

	// SimilarityThreshold is the minimum cosine similarity score (0-1) for a cache hit
	SimilarityThreshold float32

	// DefaultTTL is the default time-to-live for cache entries
	DefaultTTL time.Duration

	// MaxCacheSize is the maximum cache size in bytes (0 = unlimited)
	MaxCacheSize int64

	// EmbeddingModel is the model to use for generating embeddings
	EmbeddingModel string

	// EmbeddingDimension is the dimension of embedding vectors
	EmbeddingDimension int

	// UseCompression enables response compression
	UseCompression bool

	// EnableStats enables cache statistics tracking
	EnableStats bool
}

// DefaultCacheConfig returns the default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		Enabled:             true,
		RedisAddr:           "localhost:6379",
		RedisPassword:       "",
		RedisDB:             0,
		SimilarityThreshold: 0.95,
		DefaultTTL:          24 * time.Hour,
		MaxCacheSize:        0, // unlimited
		EmbeddingModel:      "text-embedding-3-small",
		EmbeddingDimension:  1536,
		UseCompression:      true,
		EnableStats:         true,
	}
}
