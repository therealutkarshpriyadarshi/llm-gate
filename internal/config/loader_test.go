package config

import (
	"os"
	"testing"
)

func TestLoad_WithDefaults(t *testing.T) {
	// Load config without file (should use defaults)
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check defaults
	if cfg.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Server.Port)
	}

	if cfg.Log.Level != "info" {
		t.Errorf("Expected default log level 'info', got %s", cfg.Log.Level)
	}

	if cfg.Cache.Port != 6379 {
		t.Errorf("Expected default Redis port 6379, got %d", cfg.Cache.Port)
	}
}

func TestLoad_WithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("LLMGATE_SERVER_PORT", "9090")
	os.Setenv("LLMGATE_LOG_LEVEL", "debug")
	defer func() {
		os.Unsetenv("LLMGATE_SERVER_PORT")
		os.Unsetenv("LLMGATE_LOG_LEVEL")
	}()

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("Expected port from env 9090, got %d", cfg.Server.Port)
	}

	if cfg.Log.Level != "debug" {
		t.Errorf("Expected log level from env 'debug', got %s", cfg.Log.Level)
	}
}
