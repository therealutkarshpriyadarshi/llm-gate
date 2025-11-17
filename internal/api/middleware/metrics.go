package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/telemetry"
)

// Metrics middleware collects HTTP metrics
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Increment in-flight requests
		telemetry.HTTPRequestsInFlight.Inc()
		defer telemetry.HTTPRequestsInFlight.Dec()

		// Wrap response writer to capture status
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call next handler
		next.ServeHTTP(wrapped, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(wrapped.statusCode)

		telemetry.HTTPRequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
		telemetry.HTTPRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}
