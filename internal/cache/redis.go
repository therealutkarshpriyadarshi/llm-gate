package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

const (
	// Redis key prefixes
	cachePrefix      = "llmgate:cache:"
	embeddingPrefix  = "llmgate:embedding:"
	statsPrefix      = "llmgate:stats:"
	vectorIndexName  = "llmgate:vector:idx"

	// Stats keys
	statsHitsKey   = statsPrefix + "hits"
	statsMissesKey = statsPrefix + "misses"
	statsTotalKey  = statsPrefix + "total"
)

// RedisCache implements semantic caching using Redis
type RedisCache struct {
	client              *redis.Client
	config              *CacheConfig
	embedder            Embedder
	normalizer          *RequestNormalizer
	similarityCalc      *SimilarityCalculator
	vectorIndexCreated  bool
}

// NewRedisCache creates a new Redis-based semantic cache
func NewRedisCache(config *CacheConfig, embedder Embedder) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Info().
		Str("addr", config.RedisAddr).
		Int("db", config.RedisDB).
		Msg("connected to Redis")

	cache := &RedisCache{
		client:         client,
		config:         config,
		embedder:       embedder,
		normalizer:     NewRequestNormalizer(),
		similarityCalc: NewSimilarityCalculator(),
	}

	return cache, nil
}

// Get retrieves a cached response for the given request
func (r *RedisCache) Get(ctx context.Context, req *models.ChatRequest) (*CacheEntry, error) {
	if !r.config.Enabled {
		return nil, nil
	}

	// Extract searchable text and generate embedding
	searchText := r.normalizer.ExtractSearchableText(req)
	if searchText == "" {
		return nil, nil
	}

	queryEmbedding, err := r.embedder.Embed(ctx, searchText)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate embedding")
		return nil, err
	}

	// Search for similar cached entries
	entry, err := r.findSimilarEntry(ctx, queryEmbedding)
	if err != nil {
		if err == redis.Nil {
			// Cache miss
			r.incrementMisses(ctx)
			return nil, nil
		}
		return nil, err
	}

	if entry == nil || entry.SimilarityScore < r.config.SimilarityThreshold {
		r.incrementMisses(ctx)
		return nil, nil
	}

	// Cache hit
	r.incrementHits(ctx)
	r.incrementHitCount(ctx, entry.Key)

	log.Info().
		Str("key", entry.Key).
		Float32("similarity", entry.SimilarityScore).
		Msg("cache hit")

	return entry, nil
}

// Set stores a request-response pair in the cache
func (r *RedisCache) Set(ctx context.Context, req *models.ChatRequest, resp *models.ChatResponse, ttl time.Duration) error {
	if !r.config.Enabled {
		return nil
	}

	// Generate cache key
	key, err := r.normalizer.GenerateCacheKey(req)
	if err != nil {
		return fmt.Errorf("failed to generate cache key: %w", err)
	}

	// Extract searchable text and generate embedding
	searchText := r.normalizer.ExtractSearchableText(req)
	if searchText == "" {
		return nil // Nothing to cache
	}

	embedding, err := r.embedder.Embed(ctx, searchText)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Create cache entry
	entry := &CacheEntry{
		Key:       key,
		Request:   req,
		Response:  resp,
		Embedding: embedding,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
		HitCount:  0,
	}

	// Serialize entry
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	// Store in Redis
	cacheKey := cachePrefix + key
	if err := r.client.Set(ctx, cacheKey, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to store cache entry: %w", err)
	}

	// Store embedding separately for vector search
	embeddingKey := embeddingPrefix + key
	embeddingData, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}

	if err := r.client.Set(ctx, embeddingKey, embeddingData, ttl).Err(); err != nil {
		return fmt.Errorf("failed to store embedding: %w", err)
	}

	// Update total entries counter
	r.client.Incr(ctx, statsTotalKey)

	log.Info().
		Str("key", key).
		Dur("ttl", ttl).
		Int("embedding_dim", len(embedding)).
		Msg("cached entry")

	return nil
}

// findSimilarEntry searches for the most similar cached entry
func (r *RedisCache) findSimilarEntry(ctx context.Context, queryEmbedding []float32) (*CacheEntry, error) {
	// Get all embedding keys
	embeddingKeys, err := r.client.Keys(ctx, embeddingPrefix+"*").Result()
	if err != nil {
		return nil, err
	}

	if len(embeddingKeys) == 0 {
		return nil, redis.Nil
	}

	var bestEntry *CacheEntry
	var bestSimilarity float32 = -2.0

	// Iterate through all embeddings and find the most similar one
	for _, embKey := range embeddingKeys {
		// Get embedding
		embData, err := r.client.Get(ctx, embKey).Result()
		if err != nil {
			continue
		}

		var embedding []float32
		if err := json.Unmarshal([]byte(embData), &embedding); err != nil {
			continue
		}

		// Calculate similarity
		similarity, err := r.similarityCalc.CosineSimilarity(queryEmbedding, embedding)
		if err != nil {
			continue
		}

		if similarity > bestSimilarity {
			// Extract cache key from embedding key
			cacheKey := cachePrefix + embKey[len(embeddingPrefix):]

			// Get the full cache entry
			entryData, err := r.client.Get(ctx, cacheKey).Result()
			if err != nil {
				continue
			}

			var entry CacheEntry
			if err := json.Unmarshal([]byte(entryData), &entry); err != nil {
				continue
			}

			entry.SimilarityScore = similarity
			bestSimilarity = similarity
			bestEntry = &entry
		}
	}

	if bestEntry == nil {
		return nil, redis.Nil
	}

	return bestEntry, nil
}

// Delete removes a cache entry
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	cacheKey := cachePrefix + key
	embeddingKey := embeddingPrefix + key

	pipe := r.client.Pipeline()
	pipe.Del(ctx, cacheKey)
	pipe.Del(ctx, embeddingKey)

	_, err := pipe.Exec(ctx)
	return err
}

// Clear removes all cache entries
func (r *RedisCache) Clear(ctx context.Context) error {
	// Delete all keys with our prefixes
	cacheKeys, err := r.client.Keys(ctx, cachePrefix+"*").Result()
	if err != nil {
		return err
	}

	embeddingKeys, err := r.client.Keys(ctx, embeddingPrefix+"*").Result()
	if err != nil {
		return err
	}

	allKeys := append(cacheKeys, embeddingKeys...)
	if len(allKeys) > 0 {
		if err := r.client.Del(ctx, allKeys...).Err(); err != nil {
			return err
		}
	}

	// Reset stats
	r.client.Set(ctx, statsHitsKey, 0, 0)
	r.client.Set(ctx, statsMissesKey, 0, 0)
	r.client.Set(ctx, statsTotalKey, 0, 0)

	log.Info().Msg("cache cleared")
	return nil
}

// GetStats returns cache statistics
func (r *RedisCache) GetStats(ctx context.Context) (*CacheStats, error) {
	hits, _ := r.client.Get(ctx, statsHitsKey).Int64()
	misses, _ := r.client.Get(ctx, statsMissesKey).Int64()
	total, _ := r.client.Get(ctx, statsTotalKey).Int64()

	totalRequests := hits + misses
	hitRate := 0.0
	if totalRequests > 0 {
		hitRate = float64(hits) / float64(totalRequests)
	}

	stats := &CacheStats{
		TotalEntries: total,
		TotalHits:    hits,
		TotalMisses:  misses,
		HitRate:      hitRate,
	}

	return stats, nil
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// incrementHits increments the cache hit counter
func (r *RedisCache) incrementHits(ctx context.Context) {
	if r.config.EnableStats {
		r.client.Incr(ctx, statsHitsKey)
	}
}

// incrementMisses increments the cache miss counter
func (r *RedisCache) incrementMisses(ctx context.Context) {
	if r.config.EnableStats {
		r.client.Incr(ctx, statsMissesKey)
	}
}

// incrementHitCount increments the hit count for a specific cache entry
func (r *RedisCache) incrementHitCount(ctx context.Context, key string) {
	cacheKey := cachePrefix + key

	// Get current entry
	data, err := r.client.Get(ctx, cacheKey).Result()
	if err != nil {
		return
	}

	var entry CacheEntry
	if err := json.Unmarshal([]byte(data), &entry); err != nil {
		return
	}

	// Increment hit count
	entry.HitCount++

	// Update entry
	updatedData, err := json.Marshal(entry)
	if err != nil {
		return
	}

	// Preserve original TTL
	ttl := r.client.TTL(ctx, cacheKey).Val()
	r.client.Set(ctx, cacheKey, updatedData, ttl)
}
