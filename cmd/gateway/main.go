package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/api"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/config"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/providers"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/providers/openai"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/routing"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/telemetry"
)

const version = "0.2.0-phase2"

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	telemetry.InitLogger(cfg.Log.Level, cfg.Log.Format)
	log.Info().Str("version", version).Msg("Starting LLM Gateway")

	// Initialize provider registry
	registry := initializeProviders(cfg)
	defer registry.Close()

	// Initialize router
	llmRouter := initializeRouter(cfg, registry)
	defer llmRouter.Close()

	// Create API router
	router := api.Router(version, llmRouter)

	// Create HTTP server
	server := &http.Server{
		Addr:           cfg.Server.GetAddress(),
		Handler:        router,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	// Start server in a goroutine
	go func() {
		log.Info().
			Str("address", server.Addr).
			Msg("HTTP server starting")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server failed to start")
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
		os.Exit(1)
	}

	log.Info().Msg("Server exited gracefully")
}

// initializeProviders initializes and registers all configured providers
func initializeProviders(cfg *config.Config) *providers.Registry {
	registry := providers.NewRegistry()

	// Register OpenAI provider if enabled
	if cfg.Providers.OpenAI.Enabled {
		log.Info().Msg("Initializing OpenAI provider")

		openAICfg := &openai.Config{
			APIKey:       cfg.Providers.OpenAI.APIKey,
			BaseURL:      cfg.Providers.OpenAI.BaseURL,
			Organization: cfg.Providers.OpenAI.Organization,
			Timeout:      60 * time.Second,
			MaxRetries:   3,
			RetryDelay:   1 * time.Second,
		}

		if err := registry.RegisterWithConfig("openai", openAICfg); err != nil {
			log.Error().Err(err).Msg("Failed to register OpenAI provider")
		} else {
			log.Info().Msg("OpenAI provider registered successfully")
		}
	}

	log.Info().Int("count", registry.Count()).Msg("Provider initialization complete")
	return registry
}

// initializeRouter initializes the routing layer
func initializeRouter(cfg *config.Config, registry *providers.Registry) *routing.Router {
	// Get all registered providers
	providersList := registry.GetAll()

	if len(providersList) == 0 {
		log.Warn().Msg("No providers registered - gateway will not be able to serve requests")
	}

	// Create routing configuration
	routingCfg := &routing.Config{
		Strategy:            routing.NewRoundRobinStrategy(),
		HealthCheckInterval: time.Duration(cfg.Routing.HealthCheckInterval) * time.Second,
		EnableHealthChecks:  cfg.Routing.EnableHealthChecks,
	}

	// Create router
	router := routing.NewRouter(providersList, routingCfg)

	log.Info().
		Str("strategy", routingCfg.Strategy.Name()).
		Bool("health_checks", routingCfg.EnableHealthChecks).
		Dur("check_interval", routingCfg.HealthCheckInterval).
		Msg("Router initialized")

	return router
}
