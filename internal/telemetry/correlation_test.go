package telemetry

import (
	"context"
	"testing"
)

func TestCorrelationID(t *testing.T) {
	ctx := context.Background()

	t.Run("WithCorrelationID and GetCorrelationID", func(t *testing.T) {
		correlationID := "test-correlation-id"
		ctx = WithCorrelationID(ctx, correlationID)

		got := GetCorrelationID(ctx)
		if got != correlationID {
			t.Errorf("GetCorrelationID() = %v, want %v", got, correlationID)
		}
	})

	t.Run("GetCorrelationID from empty context", func(t *testing.T) {
		got := GetCorrelationID(context.Background())
		if got != "" {
			t.Errorf("GetCorrelationID() = %v, want empty string", got)
		}
	})

	t.Run("WithCorrelationID with empty string generates ID", func(t *testing.T) {
		ctx = WithCorrelationID(ctx, "")
		got := GetCorrelationID(ctx)
		if got == "" {
			t.Error("WithCorrelationID() with empty string should generate an ID")
		}
	})
}

func TestRequestID(t *testing.T) {
	ctx := context.Background()

	t.Run("WithRequestID and GetRequestID", func(t *testing.T) {
		requestID := "test-request-id"
		ctx = WithRequestID(ctx, requestID)

		got := GetRequestID(ctx)
		if got != requestID {
			t.Errorf("GetRequestID() = %v, want %v", got, requestID)
		}
	})

	t.Run("WithRequestID with empty string generates ID", func(t *testing.T) {
		ctx = WithRequestID(ctx, "")
		got := GetRequestID(ctx)
		if got == "" {
			t.Error("WithRequestID() with empty string should generate an ID")
		}
	})
}

func TestUserID(t *testing.T) {
	ctx := context.Background()

	t.Run("WithUserID and GetUserID", func(t *testing.T) {
		userID := "test-user-id"
		ctx = WithUserID(ctx, userID)

		got := GetUserID(ctx)
		if got != userID {
			t.Errorf("GetUserID() = %v, want %v", got, userID)
		}
	})

	t.Run("GetUserID from empty context", func(t *testing.T) {
		got := GetUserID(context.Background())
		if got != "" {
			t.Errorf("GetUserID() = %v, want empty string", got)
		}
	})
}

func TestTenantID(t *testing.T) {
	ctx := context.Background()

	t.Run("WithTenantID and GetTenantID", func(t *testing.T) {
		tenantID := "test-tenant-id"
		ctx = WithTenantID(ctx, tenantID)

		got := GetTenantID(ctx)
		if got != tenantID {
			t.Errorf("GetTenantID() = %v, want %v", got, tenantID)
		}
	})

	t.Run("GetTenantID from empty context", func(t *testing.T) {
		got := GetTenantID(context.Background())
		if got != "" {
			t.Errorf("GetTenantID() = %v, want empty string", got)
		}
	})
}

func TestSessionID(t *testing.T) {
	ctx := context.Background()

	t.Run("WithSessionID and GetSessionID", func(t *testing.T) {
		sessionID := "test-session-id"
		ctx = WithSessionID(ctx, sessionID)

		got := GetSessionID(ctx)
		if got != sessionID {
			t.Errorf("GetSessionID() = %v, want %v", got, sessionID)
		}
	})
}

func TestWithAllIDs(t *testing.T) {
	correlationID := "correlation-123"
	requestID := "request-456"
	userID := "user-789"
	tenantID := "tenant-abc"
	sessionID := "session-xyz"

	ctx := WithAllIDs(context.Background(), correlationID, requestID, userID, tenantID, sessionID)

	if got := GetCorrelationID(ctx); got != correlationID {
		t.Errorf("GetCorrelationID() = %v, want %v", got, correlationID)
	}
	if got := GetRequestID(ctx); got != requestID {
		t.Errorf("GetRequestID() = %v, want %v", got, requestID)
	}
	if got := GetUserID(ctx); got != userID {
		t.Errorf("GetUserID() = %v, want %v", got, userID)
	}
	if got := GetTenantID(ctx); got != tenantID {
		t.Errorf("GetTenantID() = %v, want %v", got, tenantID)
	}
	if got := GetSessionID(ctx); got != sessionID {
		t.Errorf("GetSessionID() = %v, want %v", got, sessionID)
	}
}

func TestGetAllIDs(t *testing.T) {
	correlationID := "correlation-123"
	requestID := "request-456"
	userID := "user-789"
	tenantID := "tenant-abc"
	sessionID := "session-xyz"

	ctx := WithAllIDs(context.Background(), correlationID, requestID, userID, tenantID, sessionID)

	gotCorrelation, gotRequest, gotUser, gotTenant, gotSession := GetAllIDs(ctx)

	if gotCorrelation != correlationID {
		t.Errorf("GetAllIDs() correlationID = %v, want %v", gotCorrelation, correlationID)
	}
	if gotRequest != requestID {
		t.Errorf("GetAllIDs() requestID = %v, want %v", gotRequest, requestID)
	}
	if gotUser != userID {
		t.Errorf("GetAllIDs() userID = %v, want %v", gotUser, userID)
	}
	if gotTenant != tenantID {
		t.Errorf("GetAllIDs() tenantID = %v, want %v", gotTenant, tenantID)
	}
	if gotSession != sessionID {
		t.Errorf("GetAllIDs() sessionID = %v, want %v", gotSession, sessionID)
	}
}

func TestGenerateID(t *testing.T) {
	id1 := GenerateID()
	id2 := GenerateID()

	if id1 == "" {
		t.Error("GenerateID() should not return empty string")
	}
	if id1 == id2 {
		t.Error("GenerateID() should generate unique IDs")
	}
}
