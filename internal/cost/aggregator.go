package cost

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yourusername/llm-gate/internal/core/models"
)

// Aggregator handles cost aggregation and reporting
type Aggregator struct {
	redis *redis.Client
}

// NewAggregator creates a new cost aggregator
func NewAggregator(redisClient *redis.Client) *Aggregator {
	return &Aggregator{
		redis: redisClient,
	}
}

// GetUserUsage retrieves usage statistics for a user
func (a *Aggregator) GetUserUsage(ctx context.Context, userID string, startTime, endTime time.Time) (*UsageAggregation, error) {
	agg := &UsageAggregation{
		StartTime:  startTime,
		EndTime:    endTime,
		ByProvider: make(map[models.ProviderType]*ProviderStats),
		ByModel:    make(map[string]*ModelStats),
		UserID:     userID,
	}

	// Aggregate data from all days in the range
	current := startTime
	for current.Before(endTime) || current.Equal(endTime) {
		dayKey := fmt.Sprintf("usage:user:%s:day:%s", userID, current.Format("2006-01-02"))

		data, err := a.redis.HGetAll(ctx, dayKey).Result()
		if err != nil && err != redis.Nil {
			return nil, err
		}

		if len(data) > 0 {
			if requests, ok := data["requests"]; ok {
				if val, err := strconv.Atoi(requests); err == nil {
					agg.TotalRequests += val
				}
			}
			if inputTokens, ok := data["input_tokens"]; ok {
				if val, err := strconv.Atoi(inputTokens); err == nil {
					agg.TotalInputTokens += val
				}
			}
			if outputTokens, ok := data["output_tokens"]; ok {
				if val, err := strconv.Atoi(outputTokens); err == nil {
					agg.TotalOutputTokens += val
				}
			}
			if totalCost, ok := data["total_cost"]; ok {
				if val, err := strconv.ParseFloat(totalCost, 64); err == nil {
					agg.TotalCost += val
				}
			}
			if cacheHits, ok := data["cache_hits"]; ok {
				if val, err := strconv.Atoi(cacheHits); err == nil {
					agg.CacheHits += val
				}
			}
		}

		current = current.AddDate(0, 0, 1)
	}

	// Calculate cache hit rate
	if agg.TotalRequests > 0 {
		agg.CacheHitRate = float64(agg.CacheHits) / float64(agg.TotalRequests)
	}

	return agg, nil
}

// GetTenantUsage retrieves usage statistics for a tenant
func (a *Aggregator) GetTenantUsage(ctx context.Context, tenantID string, startTime, endTime time.Time) (*UsageAggregation, error) {
	agg := &UsageAggregation{
		StartTime:  startTime,
		EndTime:    endTime,
		ByProvider: make(map[models.ProviderType]*ProviderStats),
		ByModel:    make(map[string]*ModelStats),
		TenantID:   tenantID,
	}

	// Aggregate data from all days in the range
	current := startTime
	for current.Before(endTime) || current.Equal(endTime) {
		dayKey := fmt.Sprintf("usage:tenant:%s:day:%s", tenantID, current.Format("2006-01-02"))

		data, err := a.redis.HGetAll(ctx, dayKey).Result()
		if err != nil && err != redis.Nil {
			return nil, err
		}

		if len(data) > 0 {
			if requests, ok := data["requests"]; ok {
				if val, err := strconv.Atoi(requests); err == nil {
					agg.TotalRequests += val
				}
			}
			if inputTokens, ok := data["input_tokens"]; ok {
				if val, err := strconv.Atoi(inputTokens); err == nil {
					agg.TotalInputTokens += val
				}
			}
			if outputTokens, ok := data["output_tokens"]; ok {
				if val, err := strconv.Atoi(outputTokens); err == nil {
					agg.TotalOutputTokens += val
				}
			}
			if totalCost, ok := data["total_cost"]; ok {
				if val, err := strconv.ParseFloat(totalCost, 64); err == nil {
					agg.TotalCost += val
				}
			}
			if cacheHits, ok := data["cache_hits"]; ok {
				if val, err := strconv.Atoi(cacheHits); err == nil {
					agg.CacheHits += val
				}
			}
		}

		current = current.AddDate(0, 0, 1)
	}

	// Calculate cache hit rate
	if agg.TotalRequests > 0 {
		agg.CacheHitRate = float64(agg.CacheHits) / float64(agg.TotalRequests)
	}

	return agg, nil
}

// GetProviderUsage retrieves usage statistics for a provider
func (a *Aggregator) GetProviderUsage(ctx context.Context, provider models.ProviderType, startTime, endTime time.Time) (*ProviderStats, error) {
	stats := &ProviderStats{
		Provider: provider,
	}

	// Aggregate data from all days in the range
	current := startTime
	for current.Before(endTime) || current.Equal(endTime) {
		dayKey := fmt.Sprintf("usage:provider:%s:day:%s", provider, current.Format("2006-01-02"))

		data, err := a.redis.HGetAll(ctx, dayKey).Result()
		if err != nil && err != redis.Nil {
			return nil, err
		}

		if len(data) > 0 {
			if requests, ok := data["requests"]; ok {
				if val, err := strconv.Atoi(requests); err == nil {
					stats.Requests += val
				}
			}
			if inputTokens, ok := data["input_tokens"]; ok {
				if val, err := strconv.Atoi(inputTokens); err == nil {
					stats.InputTokens += val
				}
			}
			if outputTokens, ok := data["output_tokens"]; ok {
				if val, err := strconv.Atoi(outputTokens); err == nil {
					stats.OutputTokens += val
				}
			}
			if totalCost, ok := data["total_cost"]; ok {
				if val, err := strconv.ParseFloat(totalCost, 64); err == nil {
					stats.TotalCost += val
				}
			}
			if cacheHits, ok := data["cache_hits"]; ok {
				if val, err := strconv.Atoi(cacheHits); err == nil {
					stats.CacheHits += val
				}
			}
		}

		current = current.AddDate(0, 0, 1)
	}

	// Calculate averages and rates
	if stats.Requests > 0 {
		stats.AverageCost = stats.TotalCost / float64(stats.Requests)
		stats.CacheHitRate = float64(stats.CacheHits) / float64(stats.Requests)
	}

	return stats, nil
}

// GetModelUsage retrieves usage statistics for a model
func (a *Aggregator) GetModelUsage(ctx context.Context, model string, startTime, endTime time.Time) (*ModelStats, error) {
	stats := &ModelStats{
		Model: model,
	}

	// Aggregate data from all days in the range
	current := startTime
	for current.Before(endTime) || current.Equal(endTime) {
		dayKey := fmt.Sprintf("usage:model:%s:day:%s", model, current.Format("2006-01-02"))

		data, err := a.redis.HGetAll(ctx, dayKey).Result()
		if err != nil && err != redis.Nil {
			return nil, err
		}

		if len(data) > 0 {
			if requests, ok := data["requests"]; ok {
				if val, err := strconv.Atoi(requests); err == nil {
					stats.Requests += val
				}
			}
			if inputTokens, ok := data["input_tokens"]; ok {
				if val, err := strconv.Atoi(inputTokens); err == nil {
					stats.InputTokens += val
				}
			}
			if outputTokens, ok := data["output_tokens"]; ok {
				if val, err := strconv.Atoi(outputTokens); err == nil {
					stats.OutputTokens += val
				}
			}
			if totalCost, ok := data["total_cost"]; ok {
				if val, err := strconv.ParseFloat(totalCost, 64); err == nil {
					stats.TotalCost += val
				}
			}
		}

		current = current.AddDate(0, 0, 1)
	}

	// Calculate average cost
	if stats.Requests > 0 {
		stats.AverageCost = stats.TotalCost / float64(stats.Requests)
	}

	return stats, nil
}

// GenerateReport generates a comprehensive cost report
func (a *Aggregator) GenerateReport(ctx context.Context, startTime, endTime time.Time) (*CostReport, error) {
	report := &CostReport{
		StartTime:  startTime,
		EndTime:    endTime,
		ByUser:     make(map[string]float64),
		ByTenant:   make(map[string]float64),
		ByProvider: make(map[models.ProviderType]float64),
		ByModel:    make(map[string]float64),
		TopUsers:   make([]UserCostSummary, 0),
		TopTenants: make([]TenantCostSummary, 0),
	}

	// This is a simplified implementation
	// In production, you would scan through all user/tenant/provider/model keys
	// and aggregate the data

	// For now, return the basic report structure
	return report, nil
}
