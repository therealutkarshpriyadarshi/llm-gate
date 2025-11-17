package cost

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/yourusername/llm-gate/internal/core/models"
)

func setupTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, mr
}

func TestTracker_Record(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	tracker := NewTracker(TrackerConfig{
		RedisClient:   client,
		BufferSize:    2,
		FlushInterval: 1 * time.Second,
	})

	ctx := context.Background()

	// Test recording
	record := &UsageRecord{
		UserID:       "user1",
		TenantID:     "tenant1",
		Provider:     models.ProviderOpenAI,
		Model:        "gpt-4",
		InputTokens:  100,
		OutputTokens: 50,
		InputCost:    0.003,
		OutputCost:   0.006,
		TotalCost:    0.009,
		CacheHit:     false,
		RequestID:    "req1",
	}

	err := tracker.Record(ctx, record)
	assert.NoError(t, err)
	assert.NotEmpty(t, record.ID)
	assert.False(t, record.Timestamp.IsZero())

	// Buffer should have 1 record
	assert.Len(t, tracker.buffer, 1)

	// Add another record to trigger flush
	record2 := &UsageRecord{
		UserID:       "user1",
		TenantID:     "tenant1",
		Provider:     models.ProviderOpenAI,
		Model:        "gpt-4",
		InputTokens:  200,
		OutputTokens: 100,
		InputCost:    0.006,
		OutputCost:   0.012,
		TotalCost:    0.018,
		CacheHit:     false,
		RequestID:    "req2",
	}

	err = tracker.Record(ctx, record2)
	assert.NoError(t, err)

	// Buffer should be flushed (empty)
	time.Sleep(100 * time.Millisecond)
	assert.Len(t, tracker.buffer, 0)

	// Verify data in Redis
	key1 := "usage:record:" + record.ID
	exists, _ := client.Exists(ctx, key1).Result()
	assert.Equal(t, int64(1), exists)
}

func TestTracker_RecordFromCostInfo(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	tracker := NewTracker(TrackerConfig{
		RedisClient:   client,
		BufferSize:    100,
		FlushInterval: 1 * time.Second,
	})

	ctx := context.Background()

	costInfo := models.CostInfo{
		InputTokens:  150,
		OutputTokens: 75,
		InputCost:    0.0045,
		OutputCost:   0.009,
		TotalCost:    0.0135,
		Model:        "gpt-3.5-turbo",
		Provider:     models.ProviderOpenAI,
	}

	err := tracker.RecordFromCostInfo(ctx, costInfo, "user2", "tenant2", "req3", true)
	assert.NoError(t, err)

	// Flush to Redis
	err = tracker.Flush(ctx)
	assert.NoError(t, err)

	// Verify aggregations were updated
	dayKey := "usage:user:user2:day:" + time.Now().Format("2006-01-02")
	exists, _ := client.Exists(ctx, dayKey).Result()
	assert.Equal(t, int64(1), exists)
}

func TestTracker_Flush(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	tracker := NewTracker(TrackerConfig{
		RedisClient:   client,
		BufferSize:    100,
		FlushInterval: 1 * time.Hour, // Long interval, we'll flush manually
	})

	ctx := context.Background()

	// Add multiple records
	for i := 0; i < 5; i++ {
		record := &UsageRecord{
			UserID:       "user1",
			TenantID:     "tenant1",
			Provider:     models.ProviderOpenAI,
			Model:        "gpt-4",
			InputTokens:  100,
			OutputTokens: 50,
			TotalCost:    0.009,
			RequestID:    "req",
		}
		tracker.Record(ctx, record)
	}

	assert.Len(t, tracker.buffer, 5)

	// Flush
	err := tracker.Flush(ctx)
	assert.NoError(t, err)
	assert.Len(t, tracker.buffer, 0)

	// Verify timeseries was updated
	tsKey := "usage:timeseries"
	count, _ := client.ZCard(ctx, tsKey).Result()
	assert.Equal(t, int64(5), count)
}

func TestTracker_Close(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	tracker := NewTracker(TrackerConfig{
		RedisClient:   client,
		BufferSize:    100,
		FlushInterval: 1 * time.Hour,
	})

	ctx := context.Background()

	// Add records
	record := &UsageRecord{
		UserID:    "user1",
		TenantID:  "tenant1",
		Provider:  models.ProviderOpenAI,
		Model:     "gpt-4",
		TotalCost: 0.009,
	}
	tracker.Record(ctx, record)

	assert.Len(t, tracker.buffer, 1)

	// Close should flush
	err := tracker.Close(ctx)
	assert.NoError(t, err)
	assert.Len(t, tracker.buffer, 0)
}
