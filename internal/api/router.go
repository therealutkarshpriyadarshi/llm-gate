package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/api/handlers"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/api/middleware"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/routing"
)

// Router creates and configures the HTTP router
func Router(version string, llmRouter *routing.Router) http.Handler {
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

	// API routes
	r.Route("/v1", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"LLM Gateway API v1","version":"` + version + `"}`))
		})

		// Chat completions endpoint
		chatHandler := handlers.NewChatHandler(llmRouter)
		r.Post("/chat/completions", chatHandler.HandleChatCompletion)
	})

	return r
}
