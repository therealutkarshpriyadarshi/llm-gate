package middleware

import (
	"net/http"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/telemetry"
)

// CorrelationID middleware extracts or generates correlation IDs and adds them to the context
func CorrelationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract or generate correlation ID
		correlationID := r.Header.Get(telemetry.HeaderCorrelationID)
		if correlationID == "" {
			correlationID = telemetry.GenerateID()
		}
		ctx = telemetry.WithCorrelationID(ctx, correlationID)

		// Extract or generate request ID
		requestID := r.Header.Get(telemetry.HeaderRequestID)
		if requestID == "" {
			requestID = telemetry.GenerateID()
		}
		ctx = telemetry.WithRequestID(ctx, requestID)

		// Extract user ID if present
		if userID := r.Header.Get(telemetry.HeaderUserID); userID != "" {
			ctx = telemetry.WithUserID(ctx, userID)
		}

		// Extract tenant ID if present
		if tenantID := r.Header.Get(telemetry.HeaderTenantID); tenantID != "" {
			ctx = telemetry.WithTenantID(ctx, tenantID)
		}

		// Extract session ID if present
		if sessionID := r.Header.Get(telemetry.HeaderSessionID); sessionID != "" {
			ctx = telemetry.WithSessionID(ctx, sessionID)
		}

		// Add correlation and request IDs to response headers
		w.Header().Set(telemetry.HeaderCorrelationID, correlationID)
		w.Header().Set(telemetry.HeaderRequestID, requestID)

		// Update request with new context
		r = r.WithContext(ctx)

		// Call next handler
		next.ServeHTTP(w, r)
	})
}

// RequestContext middleware adds various request context information
// This should be used after CorrelationID middleware
func RequestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract additional context from headers if needed
		// For example, extract authorization information and add user/tenant context
		// This is where you would integrate with your auth system

		// For now, we just pass through
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
