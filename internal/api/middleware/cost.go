package middleware

import (
	"context"
	"net/http"

	"github.com/yourusername/llm-gate/internal/cost"
)

// CostTrackingMiddleware creates middleware for cost tracking
func CostTrackingMiddleware(tracker *cost.Tracker, budgetManager *cost.BudgetManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract user and tenant IDs from request context
			// In production, these would come from authentication
			userID := r.Header.Get("X-User-ID")
			tenantID := r.Header.Get("X-Tenant-ID")

			// Store in context for later use
			ctx := r.Context()
			ctx = context.WithValue(ctx, "userID", userID)
			ctx = context.WithValue(ctx, "tenantID", tenantID)
			ctx = context.WithValue(ctx, "costTracker", tracker)
			ctx = context.WithValue(ctx, "budgetManager", budgetManager)

			// Continue with request
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
