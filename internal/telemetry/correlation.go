package telemetry

import (
	"context"

	"github.com/google/uuid"
)

// Context keys for storing correlation IDs and other metadata
type contextKey string

const (
	// CorrelationIDKey is the context key for correlation ID
	CorrelationIDKey contextKey = "correlation_id"

	// RequestIDKey is the context key for request ID
	RequestIDKey contextKey = "request_id"

	// UserIDKey is the context key for user ID
	UserIDKey contextKey = "user_id"

	// TenantIDKey is the context key for tenant ID
	TenantIDKey contextKey = "tenant_id"

	// SessionIDKey is the context key for session ID
	SessionIDKey contextKey = "session_id"
)

// HTTP header names for correlation IDs
const (
	// HeaderCorrelationID is the HTTP header for correlation ID
	HeaderCorrelationID = "X-Correlation-ID"

	// HeaderRequestID is the HTTP header for request ID
	HeaderRequestID = "X-Request-ID"

	// HeaderUserID is the HTTP header for user ID
	HeaderUserID = "X-User-ID"

	// HeaderTenantID is the HTTP header for tenant ID
	HeaderTenantID = "X-Tenant-ID"

	// HeaderSessionID is the HTTP header for session ID
	HeaderSessionID = "X-Session-ID"
)

// GenerateID generates a new UUID
func GenerateID() string {
	return uuid.New().String()
}

// WithCorrelationID adds a correlation ID to the context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	if correlationID == "" {
		correlationID = GenerateID()
	}
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

// GetCorrelationID retrieves the correlation ID from the context
func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return id
	}
	return ""
}

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		requestID = GenerateID()
	}
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestID retrieves the request ID from the context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// WithUserID adds a user ID to the context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetUserID retrieves the user ID from the context
func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	return ""
}

// WithTenantID adds a tenant ID to the context
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

// GetTenantID retrieves the tenant ID from the context
func GetTenantID(ctx context.Context) string {
	if id, ok := ctx.Value(TenantIDKey).(string); ok {
		return id
	}
	return ""
}

// WithSessionID adds a session ID to the context
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, SessionIDKey, sessionID)
}

// GetSessionID retrieves the session ID from the context
func GetSessionID(ctx context.Context) string {
	if id, ok := ctx.Value(SessionIDKey).(string); ok {
		return id
	}
	return ""
}

// WithAllIDs adds all common IDs to the context
func WithAllIDs(ctx context.Context, correlationID, requestID, userID, tenantID, sessionID string) context.Context {
	ctx = WithCorrelationID(ctx, correlationID)
	ctx = WithRequestID(ctx, requestID)
	if userID != "" {
		ctx = WithUserID(ctx, userID)
	}
	if tenantID != "" {
		ctx = WithTenantID(ctx, tenantID)
	}
	if sessionID != "" {
		ctx = WithSessionID(ctx, sessionID)
	}
	return ctx
}

// GetAllIDs retrieves all common IDs from the context
func GetAllIDs(ctx context.Context) (correlationID, requestID, userID, tenantID, sessionID string) {
	return GetCorrelationID(ctx),
		GetRequestID(ctx),
		GetUserID(ctx),
		GetTenantID(ctx),
		GetSessionID(ctx)
}
