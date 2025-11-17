package middleware

import (
	"net/http"
	"time"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/telemetry"
)

// responseWriter wraps http.ResponseWriter to capture status code and size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
	size       int64
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.size += int64(n)
	return n, err
}

// Logging middleware logs HTTP requests with enhanced telemetry
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := r.Context()

		// Wrap response writer
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call next handler
		next.ServeHTTP(wrapped, r)

		// Calculate duration
		duration := time.Since(start)

		// Create structured log entry
		reqLog := telemetry.RequestLog{
			Method:        r.Method,
			Path:          r.URL.Path,
			Query:         r.URL.RawQuery,
			StatusCode:    wrapped.statusCode,
			Duration:      float64(duration.Milliseconds()),
			UserAgent:     r.UserAgent(),
			RemoteAddr:    r.RemoteAddr,
			RequestSize:   r.ContentLength,
			ResponseSize:  wrapped.size,
			CorrelationID: telemetry.GetCorrelationID(ctx),
			RequestID:     telemetry.GetRequestID(ctx),
			UserID:        telemetry.GetUserID(ctx),
			TenantID:      telemetry.GetTenantID(ctx),
		}

		// Log the request
		telemetry.LogRequest(ctx, reqLog)

		// Record metrics
		telemetry.RecordHTTPRequest(r.Method, r.URL.Path, wrapped.statusCode, duration.Seconds())
	})
}
