package config

import (
	"fmt"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Log       LogConfig       `mapstructure:"log"`
	Cache     CacheConfig     `mapstructure:"cache"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Metrics   MetricsConfig   `mapstructure:"metrics"`
	Providers ProvidersConfig `mapstructure:"providers"`
	Routing   RoutingConfig   `mapstructure:"routing"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port            int           `mapstructure:"port"`
	Host            string        `mapstructure:"host"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	MaxHeaderBytes  int           `mapstructure:"max_header_bytes"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"` // json or console
}

// CacheConfig holds Redis cache configuration
type CacheConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// DatabaseConfig holds PostgreSQL database configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

// MetricsConfig holds Prometheus metrics configuration
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Path    string `mapstructure:"path"`
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Server.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be positive")
	}

	if c.Server.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout must be positive")
	}

	if c.Log.Level == "" {
		return fmt.Errorf("log level cannot be empty")
	}

	return nil
}

// GetAddress returns the server address in host:port format
func (c *ServerConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetDSN returns the PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

// GetRedisAddress returns the Redis address in host:port format
func (c *CacheConfig) GetRedisAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// ProvidersConfig holds configuration for LLM providers
type ProvidersConfig struct {
	OpenAI  OpenAIConfig `mapstructure:"openai"`
	Enabled []string     `mapstructure:"enabled"` // List of enabled providers
}

// OpenAIConfig holds OpenAI provider configuration
type OpenAIConfig struct {
	APIKey       string `mapstructure:"api_key"`
	BaseURL      string `mapstructure:"base_url"`
	Organization string `mapstructure:"organization"`
	Enabled      bool   `mapstructure:"enabled"`
}

// RoutingConfig holds routing configuration
type RoutingConfig struct {
	Strategy            string `mapstructure:"strategy"` // round-robin, random, least-latency, cost-optimized
	EnableHealthChecks  bool   `mapstructure:"enable_health_checks"`
	HealthCheckInterval int    `mapstructure:"health_check_interval"` // in seconds
	EnableFallback      bool   `mapstructure:"enable_fallback"`
}
