package cost

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestBudgetManager_CreateBudget(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	aggregator := NewAggregator(client)
	manager := NewBudgetManager(client, aggregator)

	ctx := context.Background()

	budget := &Budget{
		UserID:          "user1",
		Period:          BudgetPeriodDaily,
		Limit:           100.0,
		AlertThresholds: []float64{0.8, 0.9},
		Enabled:         true,
	}

	err := manager.CreateBudget(ctx, budget)
	assert.NoError(t, err)
	assert.NotEmpty(t, budget.ID)
	assert.False(t, budget.CreatedAt.IsZero())
	assert.False(t, budget.UpdatedAt.IsZero())
}

func TestBudgetManager_UpdateBudget(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	aggregator := NewAggregator(client)
	manager := NewBudgetManager(client, aggregator)

	ctx := context.Background()

	budget := &Budget{
		UserID:  "user1",
		Period:  BudgetPeriodDaily,
		Limit:   100.0,
		Enabled: true,
	}

	// Create
	err := manager.CreateBudget(ctx, budget)
	assert.NoError(t, err)

	// Update
	budget.Limit = 200.0
	err = manager.UpdateBudget(ctx, budget)
	assert.NoError(t, err)
}

func TestBudgetManager_CheckBudget(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	aggregator := NewAggregator(client)
	manager := NewBudgetManager(client, aggregator)

	ctx := context.Background()

	// Create budget
	budget := &Budget{
		UserID:  "user1",
		Period:  BudgetPeriodDaily,
		Limit:   10.0,
		Enabled: true,
	}
	manager.CreateBudget(ctx, budget)

	// Check with cost below limit
	err := manager.CheckBudget(ctx, "user1", "", 5.0)
	assert.NoError(t, err)

	// Check with cost above limit would require actual usage data in Redis
	// For this test, we just verify the function runs without errors
}

func TestBudgetManager_GetBudgetStatus(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	aggregator := NewAggregator(client)
	manager := NewBudgetManager(client, aggregator)

	ctx := context.Background()

	budget := &Budget{
		UserID:          "user1",
		Period:          BudgetPeriodDaily,
		Limit:           100.0,
		AlertThresholds: []float64{0.8},
		Enabled:         true,
	}

	status, err := manager.GetBudgetStatus(ctx, budget)
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, budget, status.Budget)
	assert.Equal(t, 0.0, status.CurrentSpend)
	assert.Equal(t, 100.0, status.Remaining)
	assert.False(t, status.IsExceeded)
}

func TestBudgetManager_ForecastSpending(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	aggregator := NewAggregator(client)
	manager := NewBudgetManager(client, aggregator)

	ctx := context.Background()

	forecast, err := manager.ForecastSpending(ctx, "user1", "", 30)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, forecast, 0.0)
}

func TestBudgetManager_GetPeriodBounds(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	aggregator := NewAggregator(client)
	manager := NewBudgetManager(client, aggregator)

	tests := []struct {
		name   string
		period BudgetPeriod
	}{
		{"hourly", BudgetPeriodHourly},
		{"daily", BudgetPeriodDaily},
		{"weekly", BudgetPeriodWeekly},
		{"monthly", BudgetPeriodMonthly},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := manager.getPeriodBounds(tt.period)
			assert.True(t, start.Before(end))
			assert.True(t, end.After(time.Now()) || end.Equal(time.Now()))
		})
	}
}
