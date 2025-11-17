package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/api/handlers"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/api/middleware"
)

// Router creates and configures the HTTP router
func Router(version string) http.Handler {
	r := chi.NewRouter()

	// Apply global middleware
	r.Use(middleware.Recovery)
	r.Use(middleware.Logging)
	r.Use(middleware.CORS)
	r.Use(middleware.Metrics)

	// Health check endpoints
	healthHandler := handlers.NewHealthHandler(version)
	r.Handle("/health", healthHandler)
	r.HandleFunc("/readiness", handlers.ReadinessHandler)
	r.HandleFunc("/liveness", handlers.LivenessHandler)

	// Metrics endpoint (on different path to avoid middleware)
	r.Handle("/metrics", promhttp.Handler())

	// API routes (to be implemented in later phases)
	r.Route("/v1", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"LLM Gateway API v1","version":"` + version + `"}`))
		})

		// Chat completions endpoint (placeholder for Phase 2)
		r.Post("/chat/completions", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte(`{"error":"not implemented yet - coming in Phase 2"}`))
		})
	})

	return r
}
