package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ==================== HTTP Metrics ====================

	// HTTPRequestsTotal tracks total HTTP requests
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llmgate_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestDuration tracks HTTP request duration
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "llmgate_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// HTTPRequestsInFlight tracks current in-flight requests
	HTTPRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "llmgate_http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
	)

	// ==================== LLM Provider Metrics ====================

	// LLMRequestsTotal tracks total requests to LLM providers
	LLMRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llmgate_llm_requests_total",
			Help: "Total number of LLM requests",
		},
		[]string{"provider", "model", "status"},
	)

	// LLMRequestDuration tracks LLM request duration
	LLMRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "llmgate_llm_request_duration_seconds",
			Help:    "LLM request duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
		},
		[]string{"provider", "model"},
	)

	// LLMTokensProcessed tracks tokens processed
	LLMTokensProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llmgate_llm_tokens_processed_total",
			Help: "Total number of tokens processed",
		},
		[]string{"provider", "model", "type"}, // type: input, output
	)

	// LLMCostTotal tracks total cost of LLM requests
	LLMCostTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llmgate_llm_cost_total_usd",
			Help: "Total cost of LLM requests in USD",
		},
		[]string{"provider", "model", "user_id", "tenant_id"},
	)

	// LLMErrorsTotal tracks LLM errors
	LLMErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llmgate_llm_errors_total",
			Help: "Total number of LLM errors",
		},
		[]string{"provider", "model", "error_type"},
	)

	// ==================== Cache Metrics ====================

	// CacheRequests tracks cache lookup requests
	CacheRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llmgate_cache_requests_total",
			Help: "Total number of cache lookup requests",
		},
		[]string{"result"}, // result: hit, miss
	)

	// CacheHitRate tracks cache hit rate (gauge for current rate)
	CacheHitRate = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "llmgate_cache_hit_rate",
			Help: "Current cache hit rate (0-1)",
		},
	)

	// CacheLookupDuration tracks cache lookup duration
	CacheLookupDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "llmgate_cache_lookup_duration_seconds",
			Help:    "Cache lookup duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5},
		},
	)

	// CacheSimilarityScore tracks semantic similarity scores
	CacheSimilarityScore = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "llmgate_cache_similarity_score",
			Help:    "Semantic similarity scores for cache lookups",
			Buckets: []float64{0.5, 0.6, 0.7, 0.75, 0.8, 0.85, 0.9, 0.95, 0.98, 1.0},
		},
	)

	// EmbeddingGenerationDuration tracks embedding generation time
	EmbeddingGenerationDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "llmgate_embedding_generation_duration_seconds",
			Help:    "Time to generate embeddings in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2},
		},
	)

	// CacheSize tracks current cache size
	CacheSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "llmgate_cache_size_bytes",
			Help: "Current size of cache in bytes",
		},
	)

	// CacheEntries tracks number of cache entries
	CacheEntries = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "llmgate_cache_entries",
			Help: "Current number of cache entries",
		},
	)

	// ==================== Routing Metrics ====================

	// RoutingDecisions tracks routing strategy decisions
	RoutingDecisions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llmgate_routing_decisions_total",
			Help: "Total number of routing decisions",
		},
		[]string{"strategy", "selected_provider"},
	)

	// FallbackActivated tracks fallback activations
	FallbackActivated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llmgate_fallback_activated_total",
			Help: "Total number of fallback activations",
		},
		[]string{"from_provider", "to_provider", "reason"},
	)

	// LoadBalancerWeight tracks current load balancer weights
	LoadBalancerWeight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "llmgate_load_balancer_weight",
			Help: "Current load balancer weight for each provider",
		},
		[]string{"provider"},
	)

	// ==================== Rate Limiting Metrics ====================

	// RateLimitRequests tracks rate limit checks
	RateLimitRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llmgate_ratelimit_requests_total",
			Help: "Total number of rate limit checks",
		},
		[]string{"tier", "result"}, // result: allowed, exceeded
	)

	// RateLimitTokensRemaining tracks remaining tokens in bucket
	RateLimitTokensRemaining = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "llmgate_ratelimit_tokens_remaining",
			Help: "Remaining tokens in rate limit bucket",
		},
		[]string{"user_id", "tier", "limit_type"}, // limit_type: requests, tokens
	)

	// RateLimitExceededDuration tracks time in rate limited state
	RateLimitExceededDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "llmgate_ratelimit_exceeded_duration_seconds",
			Help:    "Duration until rate limit resets",
			Buckets: []float64{1, 5, 10, 30, 60, 300, 600},
		},
		[]string{"tier"},
	)

	// ==================== Cost & Budget Metrics ====================

	// BudgetSpending tracks current spending against budget
	BudgetSpending = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "llmgate_budget_spending_usd",
			Help: "Current spending in USD",
		},
		[]string{"user_id", "tenant_id", "period"},
	)

	// BudgetLimit tracks configured budget limits
	BudgetLimit = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "llmgate_budget_limit_usd",
			Help: "Configured budget limit in USD",
		},
		[]string{"user_id", "tenant_id", "period"},
	)

	// BudgetUtilization tracks budget utilization percentage
	BudgetUtilization = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "llmgate_budget_utilization_percent",
			Help: "Budget utilization as a percentage (0-100)",
		},
		[]string{"user_id", "tenant_id", "period"},
	)

	// BudgetAlertsTriggered tracks budget alert activations
	BudgetAlertsTriggered = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llmgate_budget_alerts_triggered_total",
			Help: "Total number of budget alerts triggered",
		},
		[]string{"user_id", "tenant_id", "threshold"},
	)

	// ==================== Prompt Management Metrics ====================

	// PromptRequests tracks prompt template usage
	PromptRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llmgate_prompt_requests_total",
			Help: "Total number of prompt template requests",
		},
		[]string{"prompt_name", "version"},
	)

	// ABTestAssignments tracks A/B test variant assignments
	ABTestAssignments = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llmgate_abtest_assignments_total",
			Help: "Total number of A/B test variant assignments",
		},
		[]string{"experiment_id", "variant"},
	)

	// ABTestConversions tracks A/B test conversions
	ABTestConversions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llmgate_abtest_conversions_total",
			Help: "Total number of A/B test conversions",
		},
		[]string{"experiment_id", "variant"},
	)

	// ==================== Token Optimization Metrics ====================

	// TokenOptimizationSavings tracks tokens saved
	TokenOptimizationSavings = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llmgate_token_optimization_savings_total",
			Help: "Total number of tokens saved through optimization",
		},
		[]string{"optimization_type"}, // type: compression, truncation
	)

	// TokenOptimizationRatio tracks optimization ratio
	TokenOptimizationRatio = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "llmgate_token_optimization_ratio",
			Help:    "Token reduction ratio (0-1)",
			Buckets: []float64{0, 0.05, 0.1, 0.15, 0.2, 0.25, 0.3, 0.4, 0.5},
		},
	)

	// ==================== System Metrics ====================

	// RedisConnectionPoolSize tracks Redis connection pool usage
	RedisConnectionPoolSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "llmgate_redis_connection_pool_size",
			Help: "Current Redis connection pool size",
		},
		[]string{"state"}, // state: active, idle
	)

	// DatabaseConnectionPoolSize tracks database connection pool usage
	DatabaseConnectionPoolSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "llmgate_database_connection_pool_size",
			Help: "Current database connection pool size",
		},
		[]string{"state"}, // state: active, idle
	)

	// GoroutinesActive tracks active goroutines
	GoroutinesActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "llmgate_goroutines_active",
			Help: "Number of active goroutines",
		},
	)

	// MemoryAllocated tracks memory allocation
	MemoryAllocated = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "llmgate_memory_allocated_bytes",
			Help: "Currently allocated memory in bytes",
		},
	)

	// ==================== Custom Business Metrics ====================

	// ActiveUsers tracks active users in the last period
	ActiveUsers = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "llmgate_active_users",
			Help: "Number of active users",
		},
		[]string{"period"}, // period: 5m, 15m, 1h, 24h
	)

	// RequestLatencyPercentiles tracks request latency percentiles
	RequestLatencyPercentiles = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "llmgate_request_latency_percentiles_seconds",
			Help:       "Request latency percentiles",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.95: 0.005, 0.99: 0.001},
		},
		[]string{"endpoint"},
	)
)

// RecordHTTPRequest records HTTP request metrics
func RecordHTTPRequest(method, path string, statusCode int, duration float64) {
	HTTPRequestsTotal.WithLabelValues(method, path, string(rune(statusCode))).Inc()
	HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
}

// RecordLLMRequest records LLM request metrics
func RecordLLMRequest(provider, model, status string, duration float64, inputTokens, outputTokens int, cost float64, userID, tenantID string) {
	LLMRequestsTotal.WithLabelValues(provider, model, status).Inc()
	LLMRequestDuration.WithLabelValues(provider, model).Observe(duration)
	LLMTokensProcessed.WithLabelValues(provider, model, "input").Add(float64(inputTokens))
	LLMTokensProcessed.WithLabelValues(provider, model, "output").Add(float64(outputTokens))
	LLMCostTotal.WithLabelValues(provider, model, userID, tenantID).Add(cost)
}

// RecordCacheLookup records cache lookup metrics
func RecordCacheLookup(hit bool, duration float64, similarityScore float64) {
	result := "miss"
	if hit {
		result = "hit"
	}
	CacheRequests.WithLabelValues(result).Inc()
	CacheLookupDuration.Observe(duration)
	if similarityScore > 0 {
		CacheSimilarityScore.Observe(similarityScore)
	}
}

// RecordRateLimitCheck records rate limit check metrics
func RecordRateLimitCheck(tier string, allowed bool) {
	result := "exceeded"
	if allowed {
		result = "allowed"
	}
	RateLimitRequests.WithLabelValues(tier, result).Inc()
}

// RecordRoutingDecision records routing decision metrics
func RecordRoutingDecision(strategy, selectedProvider string) {
	RoutingDecisions.WithLabelValues(strategy, selectedProvider).Inc()
}

// RecordFallback records fallback activation
func RecordFallback(fromProvider, toProvider, reason string) {
	FallbackActivated.WithLabelValues(fromProvider, toProvider, reason).Inc()
}

// UpdateBudgetMetrics updates budget-related metrics
func UpdateBudgetMetrics(userID, tenantID, period string, spending, limit, utilization float64) {
	BudgetSpending.WithLabelValues(userID, tenantID, period).Set(spending)
	BudgetLimit.WithLabelValues(userID, tenantID, period).Set(limit)
	BudgetUtilization.WithLabelValues(userID, tenantID, period).Set(utilization)
}

// RecordTokenOptimization records token optimization metrics
func RecordTokenOptimization(optimizationType string, tokensSaved int, ratio float64) {
	TokenOptimizationSavings.WithLabelValues(optimizationType).Add(float64(tokensSaved))
	TokenOptimizationRatio.Observe(ratio)
}
