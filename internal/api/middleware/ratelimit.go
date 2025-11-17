package middleware

import (
	"fmt"
	"net/http"

	"github.com/yourusername/llm-gate/internal/ratelimit"
)

// RateLimitMiddleware creates middleware for rate limiting
func RateLimitMiddleware(limiter *ratelimit.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract user and tenant IDs from request context
			userID := r.Header.Get("X-User-ID")
			tenantID := r.Header.Get("X-Tenant-ID")
			tierStr := r.Header.Get("X-Tier")

			// Default to free tier if not specified
			tier := ratelimit.TierFree
			if tierStr != "" {
				tier = ratelimit.Tier(tierStr)
			}

			// Check rate limit (0 tokens for request-based limit)
			status, err := limiter.Allow(r.Context(), userID, tenantID, tier, 0)
			if err != nil {
				if err == ratelimit.ErrRateLimitExceeded && status != nil {
					w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", status.Limit))
					w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", status.Remaining))
					w.Header().Set("X-RateLimit-Reset", status.ResetAt.Format("2006-01-02T15:04:05Z07:00"))
					if status.RetryAfter > 0 {
						w.Header().Set("Retry-After", fmt.Sprintf("%d", status.RetryAfter))
					}

					w.WriteHeader(http.StatusTooManyRequests)
					w.Write([]byte(`{"error": "rate limit exceeded"}`))
					return
				}

				// Other error
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "internal server error"}`))
				return
			}

			// Add rate limit headers to response
			if status != nil {
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", status.Limit))
				w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", status.Remaining))
				w.Header().Set("X-RateLimit-Reset", status.ResetAt.Format("2006-01-02T15:04:05Z07:00"))
			}

			// Continue with request
			next.ServeHTTP(w, r)
		})
	}
}
