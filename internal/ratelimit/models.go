package ratelimit

import "time"

// Tier represents a rate limit tier
type Tier string

const (
	TierFree       Tier = "free"
	TierPro        Tier = "pro"
	TierEnterprise Tier = "enterprise"
)

// Limit represents a rate limit configuration
type Limit struct {
	// RequestsPerMinute is the number of requests allowed per minute
	RequestsPerMinute int `json:"requests_per_minute"`

	// RequestsPerHour is the number of requests allowed per hour
	RequestsPerHour int `json:"requests_per_hour"`

	// RequestsPerDay is the number of requests allowed per day
	RequestsPerDay int `json:"requests_per_day"`

	// TokensPerMinute is the number of tokens allowed per minute
	TokensPerMinute int `json:"tokens_per_minute"`

	// TokensPerDay is the number of tokens allowed per day
	TokensPerDay int `json:"tokens_per_day"`

	// BurstSize is the maximum burst size
	BurstSize int `json:"burst_size"`
}

// Status represents the current rate limit status
type Status struct {
	// Allowed indicates if the request is allowed
	Allowed bool `json:"allowed"`

	// Remaining is the number of remaining requests
	Remaining int `json:"remaining"`

	// Limit is the rate limit
	Limit int `json:"limit"`

	// ResetAt is when the limit resets
	ResetAt time.Time `json:"reset_at"`

	// RetryAfter is how long to wait before retrying (in seconds)
	RetryAfter int `json:"retry_after,omitempty"`
}

// Config holds rate limiter configuration
type Config struct {
	// DefaultTier is the default tier for users without a tier
	DefaultTier Tier `json:"default_tier"`

	// Tiers maps tier names to their limits
	Tiers map[Tier]*Limit `json:"tiers"`

	// Enabled indicates if rate limiting is enabled
	Enabled bool `json:"enabled"`
}

// GetDefaultConfig returns the default rate limit configuration
func GetDefaultConfig() *Config {
	return &Config{
		DefaultTier: TierFree,
		Enabled:     true,
		Tiers: map[Tier]*Limit{
			TierFree: {
				RequestsPerMinute: 10,
				RequestsPerHour:   100,
				RequestsPerDay:    1000,
				TokensPerMinute:   50000,
				TokensPerDay:      1000000,
				BurstSize:         20,
			},
			TierPro: {
				RequestsPerMinute: 60,
				RequestsPerHour:   1000,
				RequestsPerDay:    10000,
				TokensPerMinute:   500000,
				TokensPerDay:      50000000,
				BurstSize:         100,
			},
			TierEnterprise: {
				RequestsPerMinute: 1000,
				RequestsPerHour:   10000,
				RequestsPerDay:    100000,
				TokensPerMinute:   5000000,
				TokensPerDay:      500000000,
				BurstSize:         2000,
			},
		},
	}
}
