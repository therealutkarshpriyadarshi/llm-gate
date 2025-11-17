package middleware

import (
	"net/http"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Tracing middleware adds OpenTelemetry tracing to HTTP handlers
func Tracing(next http.Handler) http.Handler {
	// Use the otelhttp instrumentation
	handler := otelhttp.NewHandler(next, "http.request",
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			return r.Method + " " + r.URL.Path
		}),
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the span from context
		span := trace.SpanFromContext(r.Context())

		// Add custom attributes
		if span.IsRecording() {
			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.path", r.URL.Path),
				attribute.String("http.user_agent", r.UserAgent()),
				attribute.String("http.remote_addr", r.RemoteAddr),
			)

			// Add correlation ID if present
			if correlationID := telemetry.GetCorrelationID(r.Context()); correlationID != "" {
				span.SetAttributes(attribute.String("correlation_id", correlationID))
			}

			// Add user ID if present
			if userID := telemetry.GetUserID(r.Context()); userID != "" {
				span.SetAttributes(attribute.String("user_id", userID))
			}

			// Add tenant ID if present
			if tenantID := telemetry.GetTenantID(r.Context()); tenantID != "" {
				span.SetAttributes(attribute.String("tenant_id", tenantID))
			}
		}

		// Serve the request
		handler.ServeHTTP(w, r)
	})
}

// TracingWithSpan is similar to Tracing but also stores the span in the context for later use
func TracingWithSpan(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Start a new span
			ctx, span := telemetry.StartSpan(ctx, r.Method+" "+r.URL.Path)
			defer span.End()

			// Add HTTP attributes
			telemetry.TraceHTTPRequest(span, r.Method, r.URL.Path, r.UserAgent(), 0)

			// Add correlation ID, user ID, etc.
			correlationID := telemetry.GetCorrelationID(ctx)
			userID := telemetry.GetUserID(ctx)
			tenantID := telemetry.GetTenantID(ctx)
			telemetry.TraceUser(span, userID, tenantID, correlationID)

			// Wrap response writer to capture status code
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Update context in request
			r = r.WithContext(ctx)

			// Call next handler
			next.ServeHTTP(wrapped, r)

			// Update span with status code
			span.SetAttributes(attribute.Int("http.status_code", wrapped.statusCode))
		})
	}
}
