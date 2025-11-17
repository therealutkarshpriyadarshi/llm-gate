package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	// ErrRateLimitExceeded is returned when rate limit is exceeded
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)

// Limiter implements distributed rate limiting using Redis
type Limiter struct {
	redis  *redis.Client
	config *Config
}

// NewLimiter creates a new rate limiter
func NewLimiter(redisClient *redis.Client, config *Config) *Limiter {
	if config == nil {
		config = GetDefaultConfig()
	}
	return &Limiter{
		redis:  redisClient,
		config: config,
	}
}

// Allow checks if a request is allowed based on rate limits
func (l *Limiter) Allow(ctx context.Context, userID, tenantID string, tier Tier, tokens int) (*Status, error) {
	if !l.config.Enabled {
		return &Status{
			Allowed:   true,
			Remaining: -1,
			Limit:     -1,
		}, nil
	}

	limit, ok := l.config.Tiers[tier]
	if !ok {
		limit = l.config.Tiers[l.config.DefaultTier]
	}

	// Check per-minute request limit
	if limit.RequestsPerMinute > 0 {
		status, err := l.checkLimit(ctx, userID, tenantID, "requests:minute", limit.RequestsPerMinute, time.Minute, 1)
		if err != nil {
			return nil, err
		}
		if !status.Allowed {
			return status, ErrRateLimitExceeded
		}
	}

	// Check per-hour request limit
	if limit.RequestsPerHour > 0 {
		status, err := l.checkLimit(ctx, userID, tenantID, "requests:hour", limit.RequestsPerHour, time.Hour, 1)
		if err != nil {
			return nil, err
		}
		if !status.Allowed {
			return status, ErrRateLimitExceeded
		}
	}

	// Check per-day request limit
	if limit.RequestsPerDay > 0 {
		status, err := l.checkLimit(ctx, userID, tenantID, "requests:day", limit.RequestsPerDay, 24*time.Hour, 1)
		if err != nil {
			return nil, err
		}
		if !status.Allowed {
			return status, ErrRateLimitExceeded
		}
	}

	// Check per-minute token limit
	if limit.TokensPerMinute > 0 && tokens > 0 {
		status, err := l.checkLimit(ctx, userID, tenantID, "tokens:minute", limit.TokensPerMinute, time.Minute, tokens)
		if err != nil {
			return nil, err
		}
		if !status.Allowed {
			return status, ErrRateLimitExceeded
		}
	}

	// Check per-day token limit
	if limit.TokensPerDay > 0 && tokens > 0 {
		status, err := l.checkLimit(ctx, userID, tenantID, "tokens:day", limit.TokensPerDay, 24*time.Hour, tokens)
		if err != nil {
			return nil, err
		}
		if !status.Allowed {
			return status, ErrRateLimitExceeded
		}
	}

	return &Status{
		Allowed:   true,
		Remaining: -1,
		Limit:     -1,
	}, nil
}

// checkLimit checks a specific rate limit using token bucket algorithm
func (l *Limiter) checkLimit(ctx context.Context, userID, tenantID, limitType string, limit int, window time.Duration, cost int) (*Status, error) {
	key := l.getKey(userID, tenantID, limitType)

	// Lua script for atomic token bucket check and decrement
	script := `
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local window = tonumber(ARGV[2])
		local cost = tonumber(ARGV[3])
		local now = tonumber(ARGV[4])

		-- Get current value and ttl
		local current = tonumber(redis.call('get', key) or '0')
		local ttl = redis.call('ttl', key)

		-- If key doesn't exist or expired, reset
		if ttl == -1 or ttl == -2 then
			current = limit
			redis.call('set', key, current, 'EX', window)
			ttl = window
		end

		-- Check if we have enough tokens
		if current >= cost then
			-- Deduct tokens
			current = current - cost
			redis.call('set', key, current, 'KEEPTTL')
			return {1, current, limit, ttl}
		else
			-- Not enough tokens
			return {0, current, limit, ttl}
		end
	`

	result, err := l.redis.Eval(ctx, script, []string{key}, limit, int(window.Seconds()), cost, time.Now().Unix()).Result()
	if err != nil {
		return nil, err
	}

	results := result.([]interface{})
	allowed := results[0].(int64) == 1
	remaining := int(results[1].(int64))
	limitVal := int(results[2].(int64))
	ttl := int(results[3].(int64))

	status := &Status{
		Allowed:   allowed,
		Remaining: remaining,
		Limit:     limitVal,
		ResetAt:   time.Now().Add(time.Duration(ttl) * time.Second),
	}

	if !allowed {
		status.RetryAfter = ttl
	}

	return status, nil
}

// GetStatus returns the current rate limit status without consuming tokens
func (l *Limiter) GetStatus(ctx context.Context, userID, tenantID string, tier Tier, limitType string) (*Status, error) {
	if !l.config.Enabled {
		return &Status{
			Allowed:   true,
			Remaining: -1,
			Limit:     -1,
		}, nil
	}

	limit, ok := l.config.Tiers[tier]
	if !ok {
		limit = l.config.Tiers[l.config.DefaultTier]
	}

	var limitVal int
	var window time.Duration

	switch limitType {
	case "requests:minute":
		limitVal = limit.RequestsPerMinute
		window = time.Minute
	case "requests:hour":
		limitVal = limit.RequestsPerHour
		window = time.Hour
	case "requests:day":
		limitVal = limit.RequestsPerDay
		window = 24 * time.Hour
	case "tokens:minute":
		limitVal = limit.TokensPerMinute
		window = time.Minute
	case "tokens:day":
		limitVal = limit.TokensPerDay
		window = 24 * time.Hour
	default:
		return nil, fmt.Errorf("unknown limit type: %s", limitType)
	}

	key := l.getKey(userID, tenantID, limitType)

	current, err := l.redis.Get(ctx, key).Int()
	if err == redis.Nil {
		current = limitVal
	} else if err != nil {
		return nil, err
	}

	ttl, err := l.redis.TTL(ctx, key).Result()
	if err != nil {
		ttl = window
	}

	return &Status{
		Allowed:   current > 0,
		Remaining: current,
		Limit:     limitVal,
		ResetAt:   time.Now().Add(ttl),
	}, nil
}

// Reset resets the rate limit for a user or tenant
func (l *Limiter) Reset(ctx context.Context, userID, tenantID string) error {
	pattern := l.getKey(userID, tenantID, "*")

	// Note: In production, use SCAN instead of KEYS for better performance
	keys, err := l.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return l.redis.Del(ctx, keys...).Err()
	}

	return nil
}

// getKey generates a Redis key for rate limiting
func (l *Limiter) getKey(userID, tenantID, limitType string) string {
	if userID != "" {
		return fmt.Sprintf("ratelimit:user:%s:%s", userID, limitType)
	} else if tenantID != "" {
		return fmt.Sprintf("ratelimit:tenant:%s:%s", tenantID, limitType)
	}
	return fmt.Sprintf("ratelimit:global:%s", limitType)
}

// UpdateConfig updates the rate limiter configuration
func (l *Limiter) UpdateConfig(config *Config) {
	l.config = config
}

// GetConfig returns the current configuration
func (l *Limiter) GetConfig() *Config {
	return l.config
}
