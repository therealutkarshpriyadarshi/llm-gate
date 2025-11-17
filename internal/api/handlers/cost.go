package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yourusername/llm-gate/internal/cost"
)

// CostHandler handles cost-related API requests
type CostHandler struct {
	tracker        *cost.Tracker
	aggregator     *cost.Aggregator
	budgetManager  *cost.BudgetManager
}

// NewCostHandler creates a new cost handler
func NewCostHandler(tracker *cost.Tracker, aggregator *cost.Aggregator, budgetManager *cost.BudgetManager) *CostHandler {
	return &CostHandler{
		tracker:        tracker,
		aggregator:     aggregator,
		budgetManager:  budgetManager,
	}
}

// GetUserUsage handles GET /api/v1/cost/usage/user/:userID
func (h *CostHandler) GetUserUsage(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		http.Error(w, "user ID is required", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	startTime, endTime := h.parseTimeRange(r)

	usage, err := h.aggregator.GetUserUsage(r.Context(), userID, startTime, endTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usage)
}

// GetTenantUsage handles GET /api/v1/cost/usage/tenant/:tenantID
func (h *CostHandler) GetTenantUsage(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	if tenantID == "" {
		http.Error(w, "tenant ID is required", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	startTime, endTime := h.parseTimeRange(r)

	usage, err := h.aggregator.GetTenantUsage(r.Context(), tenantID, startTime, endTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usage)
}

// CreateBudget handles POST /api/v1/cost/budget
func (h *CostHandler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	var budget cost.Budget
	if err := json.NewDecoder(r.Body).Decode(&budget); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.budgetManager.CreateBudget(r.Context(), &budget); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(budget)
}

// GetBudgetStatus handles GET /api/v1/cost/budget/:budgetID/status
func (h *CostHandler) GetBudgetStatus(w http.ResponseWriter, r *http.Request) {
	budgetID := chi.URLParam(r, "budgetID")
	if budgetID == "" {
		http.Error(w, "budget ID is required", http.StatusBadRequest)
		return
	}

	budget, err := h.budgetManager.GetBudget(r.Context(), budgetID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	status, err := h.budgetManager.GetBudgetStatus(r.Context(), budget)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// ForecastSpending handles GET /api/v1/cost/forecast
func (h *CostHandler) ForecastSpending(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	tenantID := r.URL.Query().Get("tenant_id")
	daysAhead := 30 // default

	if userID == "" && tenantID == "" {
		http.Error(w, "user_id or tenant_id is required", http.StatusBadRequest)
		return
	}

	forecast, err := h.budgetManager.ForecastSpending(r.Context(), userID, tenantID, daysAhead)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"user_id":         userID,
		"tenant_id":       tenantID,
		"days_ahead":      daysAhead,
		"forecast":        forecast,
		"forecast_period": time.Now().AddDate(0, 0, daysAhead).Format("2006-01-02"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GenerateReport handles GET /api/v1/cost/report
func (h *CostHandler) GenerateReport(w http.ResponseWriter, r *http.Request) {
	startTime, endTime := h.parseTimeRange(r)

	report, err := h.aggregator.GenerateReport(r.Context(), startTime, endTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// parseTimeRange parses time range from query parameters
func (h *CostHandler) parseTimeRange(r *http.Request) (time.Time, time.Time) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var startTime, endTime time.Time

	if startStr != "" {
		startTime, _ = time.Parse(time.RFC3339, startStr)
	} else {
		// Default to 30 days ago
		startTime = time.Now().AddDate(0, 0, -30)
	}

	if endStr != "" {
		endTime, _ = time.Parse(time.RFC3339, endStr)
	} else {
		endTime = time.Now()
	}

	return startTime, endTime
}
