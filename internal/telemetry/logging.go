package telemetry

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LoggingConfig holds the configuration for logging
type LoggingConfig struct {
	Level              string
	Format             string // json or console
	EnableCaller       bool
	EnableStackTrace   bool
	SamplingEnabled    bool
	SamplingRate       int // Log every Nth message when sampling is enabled
}

// DefaultLoggingConfig returns default logging configuration
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level:            "info",
		Format:           "json",
		EnableCaller:     true,
		EnableStackTrace: false,
		SamplingEnabled:  false,
		SamplingRate:     100,
	}
}

// InitLogger initializes the global logger with configuration
func InitLogger(config LoggingConfig) {
	// Set log level
	zerolog.TimeFieldFormat = time.RFC3339Nano
	logLevel := parseLogLevel(config.Level)
	zerolog.SetGlobalLevel(logLevel)

	// Set output format
	var output io.Writer = os.Stdout
	if config.Format == "console" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	// Create logger context
	loggerCtx := zerolog.New(output).With().Timestamp()

	// Add caller information if enabled
	if config.EnableCaller {
		loggerCtx = loggerCtx.Caller()
	}

	// Create base logger
	logger := loggerCtx.Logger()

	// Enable sampling if configured
	if config.SamplingEnabled {
		logger = logger.Sample(&zerolog.BasicSampler{N: uint32(config.SamplingRate)})
	}

	log.Logger = logger
}

// InitLoggerSimple initializes the global logger with simple parameters (backwards compatible)
func InitLoggerSimple(level, format string) {
	InitLogger(LoggingConfig{
		Level:            level,
		Format:           format,
		EnableCaller:     true,
		EnableStackTrace: false,
		SamplingEnabled:  false,
	})
}

// parseLogLevel parses string log level to zerolog.Level
func parseLogLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

// GetLogger returns a new logger instance with additional fields
func GetLogger(component string) zerolog.Logger {
	return log.With().Str("component", component).Logger()
}

// GetContextLogger returns a logger with context-specific fields (correlation ID, user ID, etc.)
func GetContextLogger(ctx context.Context, component string) zerolog.Logger {
	logger := GetLogger(component)

	// Add correlation ID if present
	if correlationID := GetCorrelationID(ctx); correlationID != "" {
		logger = logger.With().Str("correlation_id", correlationID).Logger()
	}

	// Add request ID if present
	if requestID := GetRequestID(ctx); requestID != "" {
		logger = logger.With().Str("request_id", requestID).Logger()
	}

	// Add user ID if present
	if userID := GetUserID(ctx); userID != "" {
		logger = logger.With().Str("user_id", userID).Logger()
	}

	// Add tenant ID if present
	if tenantID := GetTenantID(ctx); tenantID != "" {
		logger = logger.With().Str("tenant_id", tenantID).Logger()
	}

	// Add session ID if present
	if sessionID := GetSessionID(ctx); sessionID != "" {
		logger = logger.With().Str("session_id", sessionID).Logger()
	}

	return logger
}

// RequestLog represents a structured log entry for HTTP requests
type RequestLog struct {
	Method        string            `json:"method"`
	Path          string            `json:"path"`
	Query         string            `json:"query,omitempty"`
	StatusCode    int               `json:"status_code"`
	Duration      float64           `json:"duration_ms"`
	UserAgent     string            `json:"user_agent,omitempty"`
	RemoteAddr    string            `json:"remote_addr,omitempty"`
	RequestSize   int64             `json:"request_size_bytes,omitempty"`
	ResponseSize  int64             `json:"response_size_bytes,omitempty"`
	CorrelationID string            `json:"correlation_id,omitempty"`
	RequestID     string            `json:"request_id,omitempty"`
	UserID        string            `json:"user_id,omitempty"`
	TenantID      string            `json:"tenant_id,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
}

// LogRequest logs an HTTP request with structured data
func LogRequest(ctx context.Context, req RequestLog) {
	logger := GetContextLogger(ctx, "http")

	event := logger.Info()

	event.Str("method", req.Method).
		Str("path", req.Path).
		Int("status_code", req.StatusCode).
		Float64("duration_ms", req.Duration)

	if req.Query != "" {
		event.Str("query", req.Query)
	}
	if req.UserAgent != "" {
		event.Str("user_agent", req.UserAgent)
	}
	if req.RemoteAddr != "" {
		event.Str("remote_addr", req.RemoteAddr)
	}
	if req.RequestSize > 0 {
		event.Int64("request_size_bytes", req.RequestSize)
	}
	if req.ResponseSize > 0 {
		event.Int64("response_size_bytes", req.ResponseSize)
	}
	if req.Headers != nil && len(req.Headers) > 0 {
		event.Interface("headers", req.Headers)
	}

	event.Msg("HTTP request processed")
}

// LLMRequestLog represents a structured log entry for LLM requests
type LLMRequestLog struct {
	Provider      string  `json:"provider"`
	Model         string  `json:"model"`
	InputTokens   int     `json:"input_tokens"`
	OutputTokens  int     `json:"output_tokens"`
	TotalTokens   int     `json:"total_tokens"`
	Cost          float64 `json:"cost_usd"`
	Duration      float64 `json:"duration_ms"`
	CacheHit      bool    `json:"cache_hit"`
	Success       bool    `json:"success"`
	ErrorMessage  string  `json:"error_message,omitempty"`
	CorrelationID string  `json:"correlation_id,omitempty"`
	UserID        string  `json:"user_id,omitempty"`
	TenantID      string  `json:"tenant_id,omitempty"`
}

// LogLLMRequest logs an LLM request with structured data
func LogLLMRequest(ctx context.Context, req LLMRequestLog) {
	logger := GetContextLogger(ctx, "llm")

	event := logger.Info()
	if !req.Success {
		event = logger.Error()
	}

	event.Str("provider", req.Provider).
		Str("model", req.Model).
		Int("input_tokens", req.InputTokens).
		Int("output_tokens", req.OutputTokens).
		Int("total_tokens", req.TotalTokens).
		Float64("cost_usd", req.Cost).
		Float64("duration_ms", req.Duration).
		Bool("cache_hit", req.CacheHit).
		Bool("success", req.Success)

	if req.ErrorMessage != "" {
		event.Str("error_message", req.ErrorMessage)
	}

	event.Msg("LLM request processed")
}

// SlowQueryLog logs slow queries
func SlowQueryLog(ctx context.Context, operation string, duration time.Duration, threshold time.Duration) {
	if duration < threshold {
		return
	}

	logger := GetContextLogger(ctx, "performance")
	logger.Warn().
		Str("operation", operation).
		Float64("duration_ms", float64(duration.Milliseconds())).
		Float64("threshold_ms", float64(threshold.Milliseconds())).
		Msg("Slow operation detected")
}

// DebugLog logs debug information (only in debug mode)
func DebugLog(ctx context.Context, component, message string, fields map[string]interface{}) {
	logger := GetContextLogger(ctx, component)
	event := logger.Debug()

	for key, value := range fields {
		event.Interface(key, value)
	}

	event.Msg(message)
}
