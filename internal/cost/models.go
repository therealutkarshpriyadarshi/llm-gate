package cost

import (
	"time"

	"github.com/yourusername/llm-gate/internal/core/models"
)

// UsageRecord represents a single usage record
type UsageRecord struct {
	ID           string             `json:"id"`
	Timestamp    time.Time          `json:"timestamp"`
	UserID       string             `json:"user_id,omitempty"`
	TenantID     string             `json:"tenant_id,omitempty"`
	Provider     models.ProviderType `json:"provider"`
	Model        string             `json:"model"`
	InputTokens  int                `json:"input_tokens"`
	OutputTokens int                `json:"output_tokens"`
	InputCost    float64            `json:"input_cost"`
	OutputCost   float64            `json:"output_cost"`
	TotalCost    float64            `json:"total_cost"`
	CacheHit     bool               `json:"cache_hit"`
	RequestID    string             `json:"request_id"`
}

// UsageAggregation represents aggregated usage statistics
type UsageAggregation struct {
	Period          string             `json:"period"` // hour, day, week, month
	StartTime       time.Time          `json:"start_time"`
	EndTime         time.Time          `json:"end_time"`
	TotalRequests   int                `json:"total_requests"`
	TotalInputTokens int               `json:"total_input_tokens"`
	TotalOutputTokens int              `json:"total_output_tokens"`
	TotalCost       float64            `json:"total_cost"`
	CacheHits       int                `json:"cache_hits"`
	CacheHitRate    float64            `json:"cache_hit_rate"`
	ByProvider      map[models.ProviderType]*ProviderStats `json:"by_provider"`
	ByModel         map[string]*ModelStats                 `json:"by_model"`
	UserID          string             `json:"user_id,omitempty"`
	TenantID        string             `json:"tenant_id,omitempty"`
}

// ProviderStats represents statistics for a specific provider
type ProviderStats struct {
	Provider      models.ProviderType `json:"provider"`
	Requests      int                 `json:"requests"`
	InputTokens   int                 `json:"input_tokens"`
	OutputTokens  int                 `json:"output_tokens"`
	TotalCost     float64             `json:"total_cost"`
	AverageCost   float64             `json:"average_cost"`
	CacheHits     int                 `json:"cache_hits"`
	CacheHitRate  float64             `json:"cache_hit_rate"`
}

// ModelStats represents statistics for a specific model
type ModelStats struct {
	Model         string  `json:"model"`
	Requests      int     `json:"requests"`
	InputTokens   int     `json:"input_tokens"`
	OutputTokens  int     `json:"output_tokens"`
	TotalCost     float64 `json:"total_cost"`
	AverageCost   float64 `json:"average_cost"`
}

// CostReport represents a detailed cost report
type CostReport struct {
	StartTime     time.Time          `json:"start_time"`
	EndTime       time.Time          `json:"end_time"`
	TotalCost     float64            `json:"total_cost"`
	TotalRequests int                `json:"total_requests"`
	ByUser        map[string]float64 `json:"by_user"`
	ByTenant      map[string]float64 `json:"by_tenant"`
	ByProvider    map[models.ProviderType]float64 `json:"by_provider"`
	ByModel       map[string]float64 `json:"by_model"`
	TopUsers      []UserCostSummary  `json:"top_users"`
	TopTenants    []TenantCostSummary `json:"top_tenants"`
}

// UserCostSummary represents cost summary for a user
type UserCostSummary struct {
	UserID        string  `json:"user_id"`
	TotalCost     float64 `json:"total_cost"`
	RequestCount  int     `json:"request_count"`
	TokensUsed    int     `json:"tokens_used"`
}

// TenantCostSummary represents cost summary for a tenant
type TenantCostSummary struct {
	TenantID      string  `json:"tenant_id"`
	TotalCost     float64 `json:"total_cost"`
	RequestCount  int     `json:"request_count"`
	TokensUsed    int     `json:"tokens_used"`
}
