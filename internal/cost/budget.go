package cost

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	// ErrBudgetExceeded is returned when a budget limit is exceeded
	ErrBudgetExceeded = errors.New("budget limit exceeded")
	// ErrBudgetNotFound is returned when a budget is not found
	ErrBudgetNotFound = errors.New("budget not found")
)

// BudgetPeriod represents the period for budget limits
type BudgetPeriod string

const (
	BudgetPeriodHourly  BudgetPeriod = "hourly"
	BudgetPeriodDaily   BudgetPeriod = "daily"
	BudgetPeriodWeekly  BudgetPeriod = "weekly"
	BudgetPeriodMonthly BudgetPeriod = "monthly"
)

// Budget represents a budget configuration
type Budget struct {
	ID              string        `json:"id"`
	UserID          string        `json:"user_id,omitempty"`
	TenantID        string        `json:"tenant_id,omitempty"`
	Period          BudgetPeriod  `json:"period"`
	Limit           float64       `json:"limit"` // in USD
	AlertThresholds []float64     `json:"alert_thresholds"` // percentages (e.g., 0.8 for 80%)
	Enabled         bool          `json:"enabled"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

// BudgetStatus represents the current status of a budget
type BudgetStatus struct {
	Budget       *Budget   `json:"budget"`
	CurrentSpend float64   `json:"current_spend"`
	Remaining    float64   `json:"remaining"`
	Percentage   float64   `json:"percentage"`
	PeriodStart  time.Time `json:"period_start"`
	PeriodEnd    time.Time `json:"period_end"`
	IsExceeded   bool      `json:"is_exceeded"`
}

// BudgetAlert represents a budget alert
type BudgetAlert struct {
	ID          string    `json:"id"`
	BudgetID    string    `json:"budget_id"`
	Threshold   float64   `json:"threshold"`
	CurrentSpend float64  `json:"current_spend"`
	Timestamp   time.Time `json:"timestamp"`
	Message     string    `json:"message"`
}

// BudgetManager handles budget management
type BudgetManager struct {
	redis      *redis.Client
	aggregator *Aggregator
	alerts     chan *BudgetAlert
}

// NewBudgetManager creates a new budget manager
func NewBudgetManager(redisClient *redis.Client, aggregator *Aggregator) *BudgetManager {
	return &BudgetManager{
		redis:      redisClient,
		aggregator: aggregator,
		alerts:     make(chan *BudgetAlert, 100),
	}
}

// CreateBudget creates a new budget
func (bm *BudgetManager) CreateBudget(ctx context.Context, budget *Budget) error {
	if budget.ID == "" {
		budget.ID = fmt.Sprintf("budget:%d", time.Now().UnixNano())
	}
	budget.CreatedAt = time.Now()
	budget.UpdatedAt = time.Now()

	data, err := json.Marshal(budget)
	if err != nil {
		return err
	}

	key := bm.getBudgetKey(budget)
	return bm.redis.Set(ctx, key, data, 0).Err()
}

// GetBudget retrieves a budget
func (bm *BudgetManager) GetBudget(ctx context.Context, budgetID string) (*Budget, error) {
	// Try to find the budget by ID
	// In a real implementation, you would need to scan or maintain an index
	// For now, we'll return a not found error
	return nil, ErrBudgetNotFound
}

// UpdateBudget updates a budget
func (bm *BudgetManager) UpdateBudget(ctx context.Context, budget *Budget) error {
	budget.UpdatedAt = time.Now()

	data, err := json.Marshal(budget)
	if err != nil {
		return err
	}

	key := bm.getBudgetKey(budget)
	return bm.redis.Set(ctx, key, data, 0).Err()
}

// DeleteBudget deletes a budget
func (bm *BudgetManager) DeleteBudget(ctx context.Context, budget *Budget) error {
	key := bm.getBudgetKey(budget)
	return bm.redis.Del(ctx, key).Err()
}

// CheckBudget checks if a budget limit would be exceeded by a cost
func (bm *BudgetManager) CheckBudget(ctx context.Context, userID, tenantID string, cost float64) error {
	// Check user budget
	if userID != "" {
		if err := bm.checkUserBudget(ctx, userID, cost); err != nil {
			return err
		}
	}

	// Check tenant budget
	if tenantID != "" {
		if err := bm.checkTenantBudget(ctx, tenantID, cost); err != nil {
			return err
		}
	}

	return nil
}

// GetBudgetStatus retrieves the current status of a budget
func (bm *BudgetManager) GetBudgetStatus(ctx context.Context, budget *Budget) (*BudgetStatus, error) {
	periodStart, periodEnd := bm.getPeriodBounds(budget.Period)

	var currentSpend float64
	var err error

	if budget.UserID != "" {
		usage, err := bm.aggregator.GetUserUsage(ctx, budget.UserID, periodStart, periodEnd)
		if err != nil {
			return nil, err
		}
		currentSpend = usage.TotalCost
	} else if budget.TenantID != "" {
		usage, err := bm.aggregator.GetTenantUsage(ctx, budget.TenantID, periodStart, periodEnd)
		if err != nil {
			return nil, err
		}
		currentSpend = usage.TotalCost
	}

	remaining := budget.Limit - currentSpend
	percentage := 0.0
	if budget.Limit > 0 {
		percentage = (currentSpend / budget.Limit) * 100
	}

	status := &BudgetStatus{
		Budget:       budget,
		CurrentSpend: currentSpend,
		Remaining:    remaining,
		Percentage:   percentage,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
		IsExceeded:   currentSpend >= budget.Limit,
	}

	// Check for alert thresholds
	for _, threshold := range budget.AlertThresholds {
		if percentage >= threshold*100 {
			alert := &BudgetAlert{
				ID:           fmt.Sprintf("alert:%d", time.Now().UnixNano()),
				BudgetID:     budget.ID,
				Threshold:    threshold,
				CurrentSpend: currentSpend,
				Timestamp:    time.Now(),
				Message:      fmt.Sprintf("Budget %.0f%% threshold reached: $%.2f / $%.2f", threshold*100, currentSpend, budget.Limit),
			}
			select {
			case bm.alerts <- alert:
			default:
				// Alert channel is full, skip
			}
		}
	}

	return status, err
}

// GetAlerts returns the alerts channel
func (bm *BudgetManager) GetAlerts() <-chan *BudgetAlert {
	return bm.alerts
}

// ForecastSpending forecasts future spending based on historical data
func (bm *BudgetManager) ForecastSpending(ctx context.Context, userID, tenantID string, daysAhead int) (float64, error) {
	// Get historical data for the last 30 days
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -30)

	var historicalCost float64

	if userID != "" {
		usage, err := bm.aggregator.GetUserUsage(ctx, userID, startTime, endTime)
		if err != nil {
			return 0, err
		}
		historicalCost = usage.TotalCost
	} else if tenantID != "" {
		usage, err := bm.aggregator.GetTenantUsage(ctx, tenantID, startTime, endTime)
		if err != nil {
			return 0, err
		}
		historicalCost = usage.TotalCost
	}

	// Simple linear forecast: (historical cost / days) * forecast days
	dailyAverage := historicalCost / 30.0
	forecast := dailyAverage * float64(daysAhead)

	return forecast, nil
}

// Helper methods

func (bm *BudgetManager) getBudgetKey(budget *Budget) string {
	if budget.UserID != "" {
		return fmt.Sprintf("budget:user:%s:%s", budget.UserID, budget.Period)
	} else if budget.TenantID != "" {
		return fmt.Sprintf("budget:tenant:%s:%s", budget.TenantID, budget.Period)
	}
	return fmt.Sprintf("budget:global:%s", budget.Period)
}

func (bm *BudgetManager) checkUserBudget(ctx context.Context, userID string, cost float64) error {
	// Check each period type
	periods := []BudgetPeriod{BudgetPeriodHourly, BudgetPeriodDaily, BudgetPeriodWeekly, BudgetPeriodMonthly}

	for _, period := range periods {
		key := fmt.Sprintf("budget:user:%s:%s", userID, period)
		data, err := bm.redis.Get(ctx, key).Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			return err
		}

		var budget Budget
		if err := json.Unmarshal([]byte(data), &budget); err != nil {
			continue
		}

		if !budget.Enabled {
			continue
		}

		status, err := bm.GetBudgetStatus(ctx, &budget)
		if err != nil {
			return err
		}

		if status.CurrentSpend+cost > budget.Limit {
			return fmt.Errorf("%w: user %s %s budget ($%.2f / $%.2f)",
				ErrBudgetExceeded, userID, period, status.CurrentSpend+cost, budget.Limit)
		}
	}

	return nil
}

func (bm *BudgetManager) checkTenantBudget(ctx context.Context, tenantID string, cost float64) error {
	// Check each period type
	periods := []BudgetPeriod{BudgetPeriodHourly, BudgetPeriodDaily, BudgetPeriodWeekly, BudgetPeriodMonthly}

	for _, period := range periods {
		key := fmt.Sprintf("budget:tenant:%s:%s", tenantID, period)
		data, err := bm.redis.Get(ctx, key).Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			return err
		}

		var budget Budget
		if err := json.Unmarshal([]byte(data), &budget); err != nil {
			continue
		}

		if !budget.Enabled {
			continue
		}

		status, err := bm.GetBudgetStatus(ctx, &budget)
		if err != nil {
			return err
		}

		if status.CurrentSpend+cost > budget.Limit {
			return fmt.Errorf("%w: tenant %s %s budget ($%.2f / $%.2f)",
				ErrBudgetExceeded, tenantID, period, status.CurrentSpend+cost, budget.Limit)
		}
	}

	return nil
}

func (bm *BudgetManager) getPeriodBounds(period BudgetPeriod) (time.Time, time.Time) {
	now := time.Now()
	var start, end time.Time

	switch period {
	case BudgetPeriodHourly:
		start = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
		end = start.Add(time.Hour)
	case BudgetPeriodDaily:
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end = start.AddDate(0, 0, 1)
	case BudgetPeriodWeekly:
		// Start of week (Sunday)
		weekday := int(now.Weekday())
		start = time.Date(now.Year(), now.Month(), now.Day()-weekday, 0, 0, 0, 0, now.Location())
		end = start.AddDate(0, 0, 7)
	case BudgetPeriodMonthly:
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end = start.AddDate(0, 1, 0)
	default:
		start = now
		end = now
	}

	return start, end
}
