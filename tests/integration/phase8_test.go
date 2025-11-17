package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/api/middleware"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/telemetry"
)

// TestPhase8_CorrelationIDMiddleware tests correlation ID middleware
func TestPhase8_CorrelationIDMiddleware(t *testing.T) {
	// Initialize logger
	telemetry.InitLoggerSimple("info", "json")

	handler := middleware.CorrelationID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify correlation ID is in context
		correlationID := telemetry.GetCorrelationID(r.Context())
		if correlationID == "" {
			t.Error("Expected correlation ID in context")
		}

		// Verify request ID is in context
		requestID := telemetry.GetRequestID(r.Context())
		if requestID == "" {
			t.Error("Expected request ID in context")
		}

		w.WriteHeader(http.StatusOK)
	}))

	t.Run("generates correlation ID when not provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		// Check response headers
		correlationID := rec.Header().Get(telemetry.HeaderCorrelationID)
		if correlationID == "" {
			t.Error("Expected correlation ID in response headers")
		}

		requestID := rec.Header().Get(telemetry.HeaderRequestID)
		if requestID == "" {
			t.Error("Expected request ID in response headers")
		}
	})

	t.Run("uses provided correlation ID", func(t *testing.T) {
		expectedCorrelationID := "test-correlation-id"
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(telemetry.HeaderCorrelationID, expectedCorrelationID)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		// Check response headers
		correlationID := rec.Header().Get(telemetry.HeaderCorrelationID)
		if correlationID != expectedCorrelationID {
			t.Errorf("Expected correlation ID %s, got %s", expectedCorrelationID, correlationID)
		}
	})

	t.Run("extracts user and tenant IDs from headers", func(t *testing.T) {
		expectedUserID := "user-123"
		expectedTenantID := "tenant-456"

		handler := middleware.CorrelationID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := telemetry.GetUserID(r.Context())
			if userID != expectedUserID {
				t.Errorf("Expected user ID %s, got %s", expectedUserID, userID)
			}

			tenantID := telemetry.GetTenantID(r.Context())
			if tenantID != expectedTenantID {
				t.Errorf("Expected tenant ID %s, got %s", expectedTenantID, tenantID)
			}

			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(telemetry.HeaderUserID, expectedUserID)
		req.Header.Set(telemetry.HeaderTenantID, expectedTenantID)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
	})
}

// TestPhase8_LoggingMiddleware tests enhanced logging middleware
func TestPhase8_LoggingMiddleware(t *testing.T) {
	// Initialize logger
	telemetry.InitLoggerSimple("info", "json")

	t.Run("logs HTTP requests with correlation IDs", func(t *testing.T) {
		handler := middleware.CorrelationID(
			middleware.Logging(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("test response"))
				}),
			),
		)

		req := httptest.NewRequest(http.MethodGet, "/api/test?param=value", nil)
		req.Header.Set(telemetry.HeaderUserID, "user-123")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		// Verify response
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
	})
}

// TestPhase8_TracingIntegration tests OpenTelemetry tracing integration
func TestPhase8_TracingIntegration(t *testing.T) {
	// Initialize tracing
	ctx := context.Background()
	config := telemetry.DefaultTracingConfig()
	config.Enabled = false // Disable for testing to avoid needing Jaeger
	_, err := telemetry.InitTracing(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize tracing: %v", err)
	}

	t.Run("creates spans for requests", func(t *testing.T) {
		ctx, span := telemetry.StartSpan(context.Background(), "test-operation")
		defer span.End()

		// Add attributes
		telemetry.AddSpanAttributes(span,
			telemetry.AttrProvider.String("openai"),
			telemetry.AttrModel.String("gpt-4"),
		)

		// Verify span is recording
		if !span.IsRecording() {
			t.Error("Expected span to be recording")
		}
	})

	t.Run("traces LLM requests", func(t *testing.T) {
		ctx, span := telemetry.StartSpan(context.Background(), "llm-request")
		defer span.End()

		telemetry.TraceLLMRequest(span, "openai", "gpt-4", 100, 200, 0.05, false)

		// Verify span has attributes
		if !span.IsRecording() {
			t.Error("Expected span to be recording")
		}
	})

	t.Run("traces cache lookups", func(t *testing.T) {
		ctx, span := telemetry.StartSpan(context.Background(), "cache-lookup")
		defer span.End()

		telemetry.TraceCacheLookup(span, "cache-key-123", true, 0.95)

		if !span.IsRecording() {
			t.Error("Expected span to be recording")
		}
	})
}

// TestPhase8_MetricsCollection tests metrics collection
func TestPhase8_MetricsCollection(t *testing.T) {
	t.Run("records HTTP request metrics", func(t *testing.T) {
		telemetry.RecordHTTPRequest("GET", "/api/test", 200, 0.150)

		// Metrics are recorded asynchronously, so we just verify no panic
	})

	t.Run("records LLM request metrics", func(t *testing.T) {
		telemetry.RecordLLMRequest("openai", "gpt-4", "success", 2.5, 100, 200, 0.05, "user-123", "tenant-456")

		// Metrics are recorded asynchronously, so we just verify no panic
	})

	t.Run("records cache lookup metrics", func(t *testing.T) {
		telemetry.RecordCacheLookup(true, 0.010, 0.95)
		telemetry.RecordCacheLookup(false, 0.012, 0.0)

		// Metrics are recorded asynchronously, so we just verify no panic
	})

	t.Run("records rate limit checks", func(t *testing.T) {
		telemetry.RecordRateLimitCheck("pro", true)
		telemetry.RecordRateLimitCheck("free", false)

		// Metrics are recorded asynchronously, so we just verify no panic
	})

	t.Run("records routing decisions", func(t *testing.T) {
		telemetry.RecordRoutingDecision("cost-optimized", "openai")
		telemetry.RecordRoutingDecision("least-latency", "anthropic")

		// Metrics are recorded asynchronously, so we just verify no panic
	})

	t.Run("updates budget metrics", func(t *testing.T) {
		telemetry.UpdateBudgetMetrics("user-123", "tenant-456", "daily", 50.0, 100.0, 50.0)

		// Metrics are recorded asynchronously, so we just verify no panic
	})

	t.Run("records token optimization", func(t *testing.T) {
		telemetry.RecordTokenOptimization("compression", 150, 0.15)
		telemetry.RecordTokenOptimization("truncation", 300, 0.30)

		// Metrics are recorded asynchronously, so we just verify no panic
	})
}

// TestPhase8_StructuredLogging tests structured logging
func TestPhase8_StructuredLogging(t *testing.T) {
	// Initialize logger
	telemetry.InitLoggerSimple("info", "json")

	t.Run("logs HTTP requests", func(t *testing.T) {
		ctx := telemetry.WithAllIDs(
			context.Background(),
			"corr-123",
			"req-456",
			"user-789",
			"tenant-abc",
			"session-xyz",
		)

		reqLog := telemetry.RequestLog{
			Method:        "GET",
			Path:          "/api/test",
			Query:         "param=value",
			StatusCode:    200,
			Duration:      150.5,
			UserAgent:     "test-agent",
			RemoteAddr:    "127.0.0.1",
			RequestSize:   1024,
			ResponseSize:  2048,
			CorrelationID: "corr-123",
			RequestID:     "req-456",
			UserID:        "user-789",
			TenantID:      "tenant-abc",
		}

		telemetry.LogRequest(ctx, reqLog)

		// Verify no panic
	})

	t.Run("logs LLM requests", func(t *testing.T) {
		ctx := telemetry.WithCorrelationID(context.Background(), "corr-123")

		llmLog := telemetry.LLMRequestLog{
			Provider:      "openai",
			Model:         "gpt-4",
			InputTokens:   100,
			OutputTokens:  200,
			TotalTokens:   300,
			Cost:          0.05,
			Duration:      2500.0,
			CacheHit:      false,
			Success:       true,
			CorrelationID: "corr-123",
			UserID:        "user-789",
			TenantID:      "tenant-abc",
		}

		telemetry.LogLLMRequest(ctx, llmLog)

		// Verify no panic
	})

	t.Run("logs slow queries", func(t *testing.T) {
		ctx := telemetry.WithCorrelationID(context.Background(), "corr-123")

		telemetry.SlowQueryLog(ctx, "cache-lookup", 500*time.Millisecond, 100*time.Millisecond)

		// Verify no panic
	})
}

// TestPhase8_EndToEnd tests end-to-end observability
func TestPhase8_EndToEnd(t *testing.T) {
	// Initialize telemetry
	telemetry.InitLoggerSimple("info", "json")

	ctx := context.Background()
	config := telemetry.DefaultTracingConfig()
	config.Enabled = false // Disable for testing
	_, err := telemetry.InitTracing(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize tracing: %v", err)
	}

	// Create a handler with all middleware
	handler := middleware.CorrelationID(
		middleware.Logging(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()

				// Create a span
				ctx, span := telemetry.StartSpan(ctx, "handle-request")
				defer span.End()

				// Simulate LLM request
				telemetry.TraceLLMRequest(span, "openai", "gpt-4", 100, 200, 0.05, false)
				telemetry.RecordLLMRequest("openai", "gpt-4", "success", 2.5, 100, 200, 0.05,
					telemetry.GetUserID(ctx),
					telemetry.GetTenantID(ctx))

				// Simulate cache lookup
				telemetry.RecordCacheLookup(true, 0.010, 0.95)

				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			}),
		),
	)

	// Send request
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req.Header.Set(telemetry.HeaderUserID, "user-123")
	req.Header.Set(telemetry.HeaderTenantID, "tenant-456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Verify response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Verify correlation ID in response
	correlationID := rec.Header().Get(telemetry.HeaderCorrelationID)
	if correlationID == "" {
		t.Error("Expected correlation ID in response headers")
	}
}
