package cost

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/yourusername/llm-gate/internal/core/models"
)

// Tracker handles cost tracking and usage recording
type Tracker struct {
	redis      *redis.Client
	mu         sync.RWMutex
	buffer     []*UsageRecord
	bufferSize int
	flushTicker *time.Ticker
	stopChan   chan struct{}
}

// TrackerConfig holds configuration for the cost tracker
type TrackerConfig struct {
	RedisClient *redis.Client
	BufferSize  int
	FlushInterval time.Duration
}

// NewTracker creates a new cost tracker
func NewTracker(config TrackerConfig) *Tracker {
	if config.BufferSize == 0 {
		config.BufferSize = 100
	}
	if config.FlushInterval == 0 {
		config.FlushInterval = 10 * time.Second
	}

	tracker := &Tracker{
		redis:       config.RedisClient,
		buffer:      make([]*UsageRecord, 0, config.BufferSize),
		bufferSize:  config.BufferSize,
		flushTicker: time.NewTicker(config.FlushInterval),
		stopChan:    make(chan struct{}),
	}

	// Start background flusher
	go tracker.backgroundFlush()

	return tracker
}

// Record records a usage event
func (t *Tracker) Record(ctx context.Context, record *UsageRecord) error {
	if record.ID == "" {
		record.ID = uuid.New().String()
	}
	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now()
	}

	t.mu.Lock()
	t.buffer = append(t.buffer, record)
	shouldFlush := len(t.buffer) >= t.bufferSize
	t.mu.Unlock()

	if shouldFlush {
		return t.Flush(ctx)
	}

	return nil
}

// RecordFromCostInfo creates and records a usage record from cost info
func (t *Tracker) RecordFromCostInfo(ctx context.Context, costInfo models.CostInfo, userID, tenantID, requestID string, cacheHit bool) error {
	record := &UsageRecord{
		ID:           uuid.New().String(),
		Timestamp:    time.Now(),
		UserID:       userID,
		TenantID:     tenantID,
		Provider:     costInfo.Provider,
		Model:        costInfo.Model,
		InputTokens:  costInfo.InputTokens,
		OutputTokens: costInfo.OutputTokens,
		InputCost:    costInfo.InputCost,
		OutputCost:   costInfo.OutputCost,
		TotalCost:    costInfo.TotalCost,
		CacheHit:     cacheHit,
		RequestID:    requestID,
	}

	return t.Record(ctx, record)
}

// Flush flushes buffered records to Redis
func (t *Tracker) Flush(ctx context.Context) error {
	t.mu.Lock()
	if len(t.buffer) == 0 {
		t.mu.Unlock()
		return nil
	}

	// Copy buffer and clear it
	records := make([]*UsageRecord, len(t.buffer))
	copy(records, t.buffer)
	t.buffer = t.buffer[:0]
	t.mu.Unlock()

	// Store in Redis
	pipe := t.redis.Pipeline()
	now := time.Now()

	for _, record := range records {
		// Store individual record
		key := fmt.Sprintf("usage:record:%s", record.ID)
		data, err := json.Marshal(record)
		if err != nil {
			continue
		}
		pipe.Set(ctx, key, data, 7*24*time.Hour) // Keep for 7 days

		// Add to time-series sorted set
		tsKey := "usage:timeseries"
		pipe.ZAdd(ctx, tsKey, redis.Z{
			Score:  float64(record.Timestamp.Unix()),
			Member: record.ID,
		})

		// Update aggregations by user
		if record.UserID != "" {
			t.updateUserAggregation(ctx, pipe, record, now)
		}

		// Update aggregations by tenant
		if record.TenantID != "" {
			t.updateTenantAggregation(ctx, pipe, record, now)
		}

		// Update aggregations by provider
		t.updateProviderAggregation(ctx, pipe, record, now)

		// Update aggregations by model
		t.updateModelAggregation(ctx, pipe, record, now)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// backgroundFlush periodically flushes buffered records
func (t *Tracker) backgroundFlush() {
	for {
		select {
		case <-t.flushTicker.C:
			ctx := context.Background()
			if err := t.Flush(ctx); err != nil {
				// Log error (in production, use proper logging)
				fmt.Printf("Error flushing cost tracker: %v\n", err)
			}
		case <-t.stopChan:
			t.flushTicker.Stop()
			return
		}
	}
}

// Close stops the tracker and flushes remaining records
func (t *Tracker) Close(ctx context.Context) error {
	close(t.stopChan)
	return t.Flush(ctx)
}

// Helper functions for updating aggregations

func (t *Tracker) updateUserAggregation(ctx context.Context, pipe redis.Pipeliner, record *UsageRecord, now time.Time) {
	// Hourly aggregation
	hourKey := fmt.Sprintf("usage:user:%s:hour:%s", record.UserID, now.Format("2006-01-02-15"))
	pipe.HIncrBy(ctx, hourKey, "requests", 1)
	pipe.HIncrBy(ctx, hourKey, "input_tokens", int64(record.InputTokens))
	pipe.HIncrBy(ctx, hourKey, "output_tokens", int64(record.OutputTokens))
	pipe.HIncrByFloat(ctx, hourKey, "total_cost", record.TotalCost)
	if record.CacheHit {
		pipe.HIncrBy(ctx, hourKey, "cache_hits", 1)
	}
	pipe.Expire(ctx, hourKey, 30*24*time.Hour) // Keep for 30 days

	// Daily aggregation
	dayKey := fmt.Sprintf("usage:user:%s:day:%s", record.UserID, now.Format("2006-01-02"))
	pipe.HIncrBy(ctx, dayKey, "requests", 1)
	pipe.HIncrBy(ctx, dayKey, "input_tokens", int64(record.InputTokens))
	pipe.HIncrBy(ctx, dayKey, "output_tokens", int64(record.OutputTokens))
	pipe.HIncrByFloat(ctx, dayKey, "total_cost", record.TotalCost)
	if record.CacheHit {
		pipe.HIncrBy(ctx, dayKey, "cache_hits", 1)
	}
	pipe.Expire(ctx, dayKey, 90*24*time.Hour) // Keep for 90 days
}

func (t *Tracker) updateTenantAggregation(ctx context.Context, pipe redis.Pipeliner, record *UsageRecord, now time.Time) {
	// Hourly aggregation
	hourKey := fmt.Sprintf("usage:tenant:%s:hour:%s", record.TenantID, now.Format("2006-01-02-15"))
	pipe.HIncrBy(ctx, hourKey, "requests", 1)
	pipe.HIncrBy(ctx, hourKey, "input_tokens", int64(record.InputTokens))
	pipe.HIncrBy(ctx, hourKey, "output_tokens", int64(record.OutputTokens))
	pipe.HIncrByFloat(ctx, hourKey, "total_cost", record.TotalCost)
	if record.CacheHit {
		pipe.HIncrBy(ctx, hourKey, "cache_hits", 1)
	}
	pipe.Expire(ctx, hourKey, 30*24*time.Hour)

	// Daily aggregation
	dayKey := fmt.Sprintf("usage:tenant:%s:day:%s", record.TenantID, now.Format("2006-01-02"))
	pipe.HIncrBy(ctx, dayKey, "requests", 1)
	pipe.HIncrBy(ctx, dayKey, "input_tokens", int64(record.InputTokens))
	pipe.HIncrBy(ctx, dayKey, "output_tokens", int64(record.OutputTokens))
	pipe.HIncrByFloat(ctx, dayKey, "total_cost", record.TotalCost)
	if record.CacheHit {
		pipe.HIncrBy(ctx, dayKey, "cache_hits", 1)
	}
	pipe.Expire(ctx, dayKey, 90*24*time.Hour)
}

func (t *Tracker) updateProviderAggregation(ctx context.Context, pipe redis.Pipeliner, record *UsageRecord, now time.Time) {
	// Daily aggregation by provider
	dayKey := fmt.Sprintf("usage:provider:%s:day:%s", record.Provider, now.Format("2006-01-02"))
	pipe.HIncrBy(ctx, dayKey, "requests", 1)
	pipe.HIncrBy(ctx, dayKey, "input_tokens", int64(record.InputTokens))
	pipe.HIncrBy(ctx, dayKey, "output_tokens", int64(record.OutputTokens))
	pipe.HIncrByFloat(ctx, dayKey, "total_cost", record.TotalCost)
	if record.CacheHit {
		pipe.HIncrBy(ctx, dayKey, "cache_hits", 1)
	}
	pipe.Expire(ctx, dayKey, 90*24*time.Hour)
}

func (t *Tracker) updateModelAggregation(ctx context.Context, pipe redis.Pipeliner, record *UsageRecord, now time.Time) {
	// Daily aggregation by model
	dayKey := fmt.Sprintf("usage:model:%s:day:%s", record.Model, now.Format("2006-01-02"))
	pipe.HIncrBy(ctx, dayKey, "requests", 1)
	pipe.HIncrBy(ctx, dayKey, "input_tokens", int64(record.InputTokens))
	pipe.HIncrBy(ctx, dayKey, "output_tokens", int64(record.OutputTokens))
	pipe.HIncrByFloat(ctx, dayKey, "total_cost", record.TotalCost)
	pipe.Expire(ctx, dayKey, 90*24*time.Hour)
}
