package integration

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/llm-gate/internal/core/models"
	"github.com/yourusername/llm-gate/internal/cost"
	"github.com/yourusername/llm-gate/internal/optimization"
	"github.com/yourusername/llm-gate/internal/ratelimit"
)

// TestPhase7Integration tests the integration of Phase 7 features
func TestPhase7Integration(t *testing.T) {
	// Setup test Redis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	ctx := context.Background()

	t.Run("Cost Tracking End-to-End", func(t *testing.T) {
		tracker := cost.NewTracker(cost.TrackerConfig{
			RedisClient:   client,
			BufferSize:    10,
			FlushInterval: 1 * time.Second,
		})
		defer tracker.Close(ctx)

		// Record usage
		costInfo := models.CostInfo{
			InputTokens:  1000,
			OutputTokens: 500,
			InputCost:    0.03,
			OutputCost:   0.06,
			TotalCost:    0.09,
			Model:        "gpt-4",
			Provider:     models.ProviderOpenAI,
		}

		err := tracker.RecordFromCostInfo(ctx, costInfo, "user1", "tenant1", "req1", false)
		assert.NoError(t, err)

		// Flush to Redis
		err = tracker.Flush(ctx)
		assert.NoError(t, err)

		// Verify aggregation
		aggregator := cost.NewAggregator(client)
		usage, err := aggregator.GetUserUsage(ctx, "user1", time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour))
		assert.NoError(t, err)
		assert.Equal(t, 1, usage.TotalRequests)
		assert.Equal(t, 1000, usage.TotalInputTokens)
		assert.Equal(t, 500, usage.TotalOutputTokens)
		assert.InDelta(t, 0.09, usage.TotalCost, 0.001)
	})

	t.Run("Budget Management End-to-End", func(t *testing.T) {
		aggregator := cost.NewAggregator(client)
		budgetManager := cost.NewBudgetManager(client, aggregator)

		// Create budget
		budget := &cost.Budget{
			UserID:          "user2",
			Period:          cost.BudgetPeriodDaily,
			Limit:           100.0,
			AlertThresholds: []float64{0.8, 0.9},
			Enabled:         true,
		}

		err := budgetManager.CreateBudget(ctx, budget)
		assert.NoError(t, err)

		// Get status
		status, err := budgetManager.GetBudgetStatus(ctx, budget)
		assert.NoError(t, err)
		assert.Equal(t, 100.0, status.Remaining)
		assert.False(t, status.IsExceeded)

		// Check budget (should pass)
		err = budgetManager.CheckBudget(ctx, "user2", "", 50.0)
		assert.NoError(t, err)
	})

	t.Run("Rate Limiting End-to-End", func(t *testing.T) {
		config := ratelimit.GetDefaultConfig()
		limiter := ratelimit.NewLimiter(client, config)

		// Make requests within limit
		for i := 0; i < 5; i++ {
			status, err := limiter.Allow(ctx, "user3", "", ratelimit.TierFree, 0)
			assert.NoError(t, err)
			assert.True(t, status.Allowed)
		}

		// Check status
		status, err := limiter.GetStatus(ctx, "user3", "", ratelimit.TierFree, "requests:minute")
		assert.NoError(t, err)
		assert.Less(t, status.Remaining, config.Tiers[ratelimit.TierFree].RequestsPerMinute)

		// Reset limits
		err = limiter.Reset(ctx, "user3", "")
		assert.NoError(t, err)

		// Status should be reset
		status, err = limiter.GetStatus(ctx, "user3", "", ratelimit.TierFree, "requests:minute")
		assert.NoError(t, err)
		assert.Equal(t, config.Tiers[ratelimit.TierFree].RequestsPerMinute, status.Remaining)
	})

	t.Run("Token Optimization End-to-End", func(t *testing.T) {
		config := &optimization.OptimizerConfig{
			EnableCompression:     true,
			EnableTruncation:      true,
			MaxPromptTokens:       100,
			MaxResponseTokens:     50,
			EnableSmartTruncation: true,
		}
		optimizer := optimization.NewTokenOptimizer(config)

		// Create a request with excessive whitespace
		req := &models.UnifiedRequest{
			Messages: []models.Message{
				{Role: "system", Content: "You    are    a    helpful    assistant"},
				{Role: "user", Content: "What   is   the   meaning   of   life?"},
			},
		}

		// Optimize
		optimized, err := optimizer.OptimizeRequest(req)
		assert.NoError(t, err)
		assert.NotNil(t, optimized)

		// Verify compression
		assert.NotContains(t, optimized.Messages[0].Content, "    ")
		assert.NotContains(t, optimized.Messages[1].Content, "   ")

		// Calculate savings
		savings := optimizer.CalculateTokenSavings(req, optimized)
		assert.GreaterOrEqual(t, savings, 0)
	})

	t.Run("Full Workflow: Cost + Rate Limit + Optimization", func(t *testing.T) {
		// Setup all components
		tracker := cost.NewTracker(cost.TrackerConfig{
			RedisClient:   client,
			BufferSize:    10,
			FlushInterval: 1 * time.Second,
		})
		defer tracker.Close(ctx)

		aggregator := cost.NewAggregator(client)
		budgetManager := cost.NewBudgetManager(client, aggregator)

		rateLimiter := ratelimit.NewLimiter(client, ratelimit.GetDefaultConfig())

		optimizer := optimization.NewTokenOptimizer(&optimization.OptimizerConfig{
			EnableCompression:     true,
			EnableTruncation:      true,
			MaxPromptTokens:       1000,
			EnableSmartTruncation: true,
		})

		// Create budget for user
		budget := &cost.Budget{
			UserID:  "user4",
			Period:  cost.BudgetPeriodDaily,
			Limit:   1000.0,
			Enabled: true,
		}
		budgetManager.CreateBudget(ctx, budget)

		// Simulate request workflow
		userID := "user4"
		tenantID := "tenant4"

		// Step 1: Check rate limit
		status, err := rateLimiter.Allow(ctx, userID, tenantID, ratelimit.TierPro, 0)
		assert.NoError(t, err)
		assert.True(t, status.Allowed)

		// Step 2: Optimize request
		req := &models.UnifiedRequest{
			Messages: []models.Message{
				{Role: "user", Content: "Hello    world    with    spaces"},
			},
		}
		optimizedReq, err := optimizer.OptimizeRequest(req)
		assert.NoError(t, err)

		// Step 3: Process request (simulated)
		costInfo := models.CostInfo{
			InputTokens:  500,
			OutputTokens: 250,
			InputCost:    0.015,
			OutputCost:   0.03,
			TotalCost:    0.045,
			Model:        "gpt-3.5-turbo",
			Provider:     models.ProviderOpenAI,
		}

		// Step 4: Check budget
		err = budgetManager.CheckBudget(ctx, userID, tenantID, costInfo.TotalCost)
		assert.NoError(t, err)

		// Step 5: Record cost
		err = tracker.RecordFromCostInfo(ctx, costInfo, userID, tenantID, "req4", false)
		assert.NoError(t, err)

		// Verify everything worked
		assert.NotNil(t, optimizedReq)
		assert.Less(t, len(optimizedReq.Messages[0].Content), len(req.Messages[0].Content))
	})
}

// TestPhase7Performance tests performance aspects of Phase 7 features
func TestPhase7Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	mr, _ := miniredis.Run()
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	ctx := context.Background()

	t.Run("Cost Tracker Throughput", func(t *testing.T) {
		tracker := cost.NewTracker(cost.TrackerConfig{
			RedisClient:   client,
			BufferSize:    100,
			FlushInterval: 5 * time.Second,
		})
		defer tracker.Close(ctx)

		start := time.Now()
		numRecords := 1000

		for i := 0; i < numRecords; i++ {
			costInfo := models.CostInfo{
				InputTokens:  100,
				OutputTokens: 50,
				TotalCost:    0.01,
				Model:        "gpt-3.5-turbo",
				Provider:     models.ProviderOpenAI,
			}
			tracker.RecordFromCostInfo(ctx, costInfo, "user1", "tenant1", "req", false)
		}

		elapsed := time.Since(start)
		throughput := float64(numRecords) / elapsed.Seconds()

		t.Logf("Cost tracker throughput: %.2f records/sec", throughput)
		assert.Greater(t, throughput, 1000.0) // Should handle >1000 records/sec
	})

	t.Run("Rate Limiter Latency", func(t *testing.T) {
		limiter := ratelimit.NewLimiter(client, ratelimit.GetDefaultConfig())

		start := time.Now()
		numChecks := 1000

		for i := 0; i < numChecks; i++ {
			limiter.Allow(ctx, "user1", "", ratelimit.TierFree, 0)
		}

		elapsed := time.Since(start)
		avgLatency := elapsed / time.Duration(numChecks)

		t.Logf("Rate limiter average latency: %v", avgLatency)
		assert.Less(t, avgLatency, 5*time.Millisecond) // Should be <5ms per check
	})
}
