package telemetry

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger initializes the global logger
func InitLogger(level, format string) {
	// Set log level
	zerolog.TimeFieldFormat = time.RFC3339Nano
	logLevel := parseLogLevel(level)
	zerolog.SetGlobalLevel(logLevel)

	// Set output format
	var output io.Writer = os.Stdout
	if format == "console" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	// Create logger
	logger := zerolog.New(output).With().Timestamp().Caller().Logger()
	log.Logger = logger
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
