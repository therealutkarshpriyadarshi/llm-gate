package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
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

func TestLimiter_Allow(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	config := GetDefaultConfig()
	limiter := NewLimiter(client, config)

	ctx := context.Background()

	// First request should be allowed
	status, err := limiter.Allow(ctx, "user1", "", TierFree, 0)
	assert.NoError(t, err)
	assert.True(t, status.Allowed)

	// Make requests up to the limit
	limit := config.Tiers[TierFree].RequestsPerMinute
	for i := 1; i < limit; i++ {
		status, err = limiter.Allow(ctx, "user1", "", TierFree, 0)
		assert.NoError(t, err)
		assert.True(t, status.Allowed)
	}

	// Next request should exceed limit
	status, err = limiter.Allow(ctx, "user1", "", TierFree, 0)
	assert.Error(t, err)
	assert.Equal(t, ErrRateLimitExceeded, err)
	assert.False(t, status.Allowed)
}

func TestLimiter_AllowWithTokens(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	config := GetDefaultConfig()
	limiter := NewLimiter(client, config)

	ctx := context.Background()

	// Allow request with tokens
	status, err := limiter.Allow(ctx, "user1", "", TierFree, 1000)
	assert.NoError(t, err)
	assert.True(t, status.Allowed)

	// Exceed token limit
	tokenLimit := config.Tiers[TierFree].TokensPerMinute
	status, err = limiter.Allow(ctx, "user1", "", TierFree, tokenLimit)
	assert.Error(t, err)
	assert.Equal(t, ErrRateLimitExceeded, err)
}

func TestLimiter_GetStatus(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	config := GetDefaultConfig()
	limiter := NewLimiter(client, config)

	ctx := context.Background()

	// Get initial status
	status, err := limiter.GetStatus(ctx, "user1", "", TierFree, "requests:minute")
	assert.NoError(t, err)
	assert.True(t, status.Allowed)
	assert.Equal(t, config.Tiers[TierFree].RequestsPerMinute, status.Limit)

	// Make some requests
	limiter.Allow(ctx, "user1", "", TierFree, 0)
	limiter.Allow(ctx, "user1", "", TierFree, 0)

	// Get status after requests
	status, err = limiter.GetStatus(ctx, "user1", "", TierFree, "requests:minute")
	assert.NoError(t, err)
	assert.Less(t, status.Remaining, config.Tiers[TierFree].RequestsPerMinute)
}

func TestLimiter_Reset(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	config := GetDefaultConfig()
	limiter := NewLimiter(client, config)

	ctx := context.Background()

	// Make some requests
	limiter.Allow(ctx, "user1", "", TierFree, 0)
	limiter.Allow(ctx, "user1", "", TierFree, 0)

	// Reset
	err := limiter.Reset(ctx, "user1", "")
	assert.NoError(t, err)

	// Status should be reset
	status, err := limiter.GetStatus(ctx, "user1", "", TierFree, "requests:minute")
	assert.NoError(t, err)
	assert.Equal(t, config.Tiers[TierFree].RequestsPerMinute, status.Remaining)
}

func TestLimiter_DifferentTiers(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	config := GetDefaultConfig()
	limiter := NewLimiter(client, config)

	ctx := context.Background()

	tests := []struct {
		name string
		tier Tier
	}{
		{"free tier", TierFree},
		{"pro tier", TierPro},
		{"enterprise tier", TierEnterprise},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := limiter.Allow(ctx, "user_"+string(tt.tier), "", tt.tier, 0)
			assert.NoError(t, err)
			assert.True(t, status.Allowed)
		})
	}
}

func TestLimiter_Disabled(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	config := GetDefaultConfig()
	config.Enabled = false
	limiter := NewLimiter(client, config)

	ctx := context.Background()

	// When disabled, all requests should be allowed
	for i := 0; i < 1000; i++ {
		status, err := limiter.Allow(ctx, "user1", "", TierFree, 0)
		assert.NoError(t, err)
		assert.True(t, status.Allowed)
	}
}

func TestLimiter_TenantIsolation(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	config := GetDefaultConfig()
	limiter := NewLimiter(client, config)

	ctx := context.Background()

	// User in tenant1 makes requests
	limiter.Allow(ctx, "", "tenant1", TierFree, 0)
	limiter.Allow(ctx, "", "tenant1", TierFree, 0)

	// User in tenant2 should have separate limits
	status, err := limiter.Allow(ctx, "", "tenant2", TierFree, 0)
	assert.NoError(t, err)
	assert.True(t, status.Allowed)

	// Verify tenant1 status
	status1, _ := limiter.GetStatus(ctx, "", "tenant1", TierFree, "requests:minute")
	status2, _ := limiter.GetStatus(ctx, "", "tenant2", TierFree, "requests:minute")

	assert.Less(t, status1.Remaining, status2.Remaining)
}
