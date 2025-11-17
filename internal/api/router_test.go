package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/routing"
)

func createTestRouter() *routing.Router {
	// Create empty router for testing endpoints that don't need providers
	return routing.NewRouter(nil, routing.DefaultConfig())
}

func TestRouter_HealthEndpoint(t *testing.T) {
	llmRouter := createTestRouter()
	defer llmRouter.Close()
	router := Router("test-version", llmRouter)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRouter_ReadinessEndpoint(t *testing.T) {
	llmRouter := createTestRouter()
	defer llmRouter.Close()
	router := Router("test-version", llmRouter)

	req := httptest.NewRequest(http.MethodGet, "/readiness", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRouter_LivenessEndpoint(t *testing.T) {
	llmRouter := createTestRouter()
	defer llmRouter.Close()
	router := Router("test-version", llmRouter)

	req := httptest.NewRequest(http.MethodGet, "/liveness", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRouter_V1RootEndpoint(t *testing.T) {
	llmRouter := createTestRouter()
	defer llmRouter.Close()
	router := Router("1.0.0", llmRouter)

	req := httptest.NewRequest(http.MethodGet, "/v1/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Expected non-empty response body")
	}
}

func TestRouter_ChatCompletionsEndpoint(t *testing.T) {
	llmRouter := createTestRouter()
	defer llmRouter.Close()
	router := Router("1.0.0", llmRouter)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 400 (bad request) or 503 (no providers) instead of 501 (not implemented)
	if w.Code == http.StatusNotImplemented {
		t.Errorf("Chat completions endpoint should be implemented, got status %d", w.Code)
	}
}

func TestRouter_MetricsEndpoint(t *testing.T) {
	llmRouter := createTestRouter()
	defer llmRouter.Close()
	router := Router("test-version", llmRouter)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}
