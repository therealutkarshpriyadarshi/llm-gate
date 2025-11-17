package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	// TracerName is the name of the tracer
	TracerName = "llm-gateway"

	// Service name for traces
	ServiceName = "llm-gateway"
)

// TracingConfig holds the configuration for tracing
type TracingConfig struct {
	Enabled           bool
	ServiceName       string
	ServiceVersion    string
	Environment       string
	OTLPEndpoint      string
	SamplingRate      float64
	ExportTimeout     time.Duration
	ExportBatchSize   int
	ExportMaxQueue    int
}

// DefaultTracingConfig returns default tracing configuration
func DefaultTracingConfig() TracingConfig {
	return TracingConfig{
		Enabled:         true,
		ServiceName:     ServiceName,
		ServiceVersion:  "1.0.0",
		Environment:     "development",
		OTLPEndpoint:    "localhost:4317",
		SamplingRate:    1.0,
		ExportTimeout:   30 * time.Second,
		ExportBatchSize: 512,
		ExportMaxQueue:  2048,
	}
}

// TracerProvider is a wrapper around the OpenTelemetry TracerProvider
type TracerProvider struct {
	provider *sdktrace.TracerProvider
	config   TracingConfig
}

// InitTracing initializes OpenTelemetry tracing
func InitTracing(ctx context.Context, config TracingConfig) (*TracerProvider, error) {
	if !config.Enabled {
		// Set no-op tracer provider
		otel.SetTracerProvider(trace.NewNoopTracerProvider())
		return &TracerProvider{
			provider: nil,
			config:   config,
		}, nil
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			attribute.String("environment", config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP exporter
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(config.OTLPEndpoint),
		otlptracegrpc.WithInsecure(), // Use insecure for local development
		otlptracegrpc.WithTimeout(config.ExportTimeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create sampler based on sampling rate
	var sampler sdktrace.Sampler
	if config.SamplingRate >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if config.SamplingRate <= 0.0 {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(config.SamplingRate)
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithMaxExportBatchSize(config.ExportBatchSize),
			sdktrace.WithMaxQueueSize(config.ExportMaxQueue),
			sdktrace.WithBatchTimeout(5*time.Second),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator to tracecontext (W3C Trace Context)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &TracerProvider{
		provider: tp,
		config:   config,
	}, nil
}

// Shutdown gracefully shuts down the tracer provider
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	if tp.provider == nil {
		return nil
	}
	return tp.provider.Shutdown(ctx)
}

// GetTracer returns a tracer instance
func GetTracer() trace.Tracer {
	return otel.Tracer(TracerName)
}

// StartSpan starts a new span with the given name and options
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return GetTracer().Start(ctx, name, opts...)
}

// AddSpanAttributes adds attributes to the current span
func AddSpanAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	if span != nil && span.IsRecording() {
		span.SetAttributes(attrs...)
	}
}

// AddSpanEvent adds an event to the current span
func AddSpanEvent(span trace.Span, name string, attrs ...attribute.KeyValue) {
	if span != nil && span.IsRecording() {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// RecordError records an error on the current span
func RecordError(span trace.Span, err error, attrs ...attribute.KeyValue) {
	if span != nil && span.IsRecording() && err != nil {
		span.RecordError(err, trace.WithAttributes(attrs...))
	}
}

// Common span attributes for LLM Gateway operations
var (
	// Provider attributes
	AttrProvider      = attribute.Key("llm.provider")
	AttrModel         = attribute.Key("llm.model")
	AttrRequestTokens = attribute.Key("llm.request_tokens")
	AttrResponseTokens = attribute.Key("llm.response_tokens")
	AttrTotalTokens   = attribute.Key("llm.total_tokens")
	AttrCost          = attribute.Key("llm.cost")

	// Cache attributes
	AttrCacheHit      = attribute.Key("cache.hit")
	AttrCacheKey      = attribute.Key("cache.key")
	AttrSimilarityScore = attribute.Key("cache.similarity_score")

	// Request attributes
	AttrUserID        = attribute.Key("user.id")
	AttrTenantID      = attribute.Key("tenant.id")
	AttrRequestID     = attribute.Key("request.id")
	AttrCorrelationID = attribute.Key("correlation.id")

	// Routing attributes
	AttrRoutingStrategy = attribute.Key("routing.strategy")
	AttrFallbackUsed    = attribute.Key("routing.fallback_used")

	// Rate limiting attributes
	AttrRateLimitTier = attribute.Key("ratelimit.tier")
	AttrRateLimitExceeded = attribute.Key("ratelimit.exceeded")
)

// TraceHTTPRequest adds HTTP-specific attributes to a span
func TraceHTTPRequest(span trace.Span, method, path, userAgent string, statusCode int) {
	if span == nil || !span.IsRecording() {
		return
	}

	span.SetAttributes(
		attribute.String("http.method", method),
		attribute.String("http.path", path),
		attribute.String("http.user_agent", userAgent),
		attribute.Int("http.status_code", statusCode),
	)
}

// TraceLLMRequest adds LLM-specific attributes to a span
func TraceLLMRequest(span trace.Span, provider, model string, inputTokens, outputTokens int, cost float64, cacheHit bool) {
	if span == nil || !span.IsRecording() {
		return
	}

	span.SetAttributes(
		AttrProvider.String(provider),
		AttrModel.String(model),
		AttrRequestTokens.Int(inputTokens),
		AttrResponseTokens.Int(outputTokens),
		AttrTotalTokens.Int(inputTokens+outputTokens),
		AttrCost.Float64(cost),
		AttrCacheHit.Bool(cacheHit),
	)
}

// TraceCacheLookup adds cache-specific attributes to a span
func TraceCacheLookup(span trace.Span, cacheKey string, hit bool, similarityScore float64) {
	if span == nil || !span.IsRecording() {
		return
	}

	span.SetAttributes(
		AttrCacheKey.String(cacheKey),
		AttrCacheHit.Bool(hit),
		AttrSimilarityScore.Float64(similarityScore),
	)
}

// TraceRouting adds routing-specific attributes to a span
func TraceRouting(span trace.Span, strategy string, fallbackUsed bool) {
	if span == nil || !span.IsRecording() {
		return
	}

	span.SetAttributes(
		AttrRoutingStrategy.String(strategy),
		AttrFallbackUsed.Bool(fallbackUsed),
	)
}

// TraceUser adds user-specific attributes to a span
func TraceUser(span trace.Span, userID, tenantID, correlationID string) {
	if span == nil || !span.IsRecording() {
		return
	}

	attrs := make([]attribute.KeyValue, 0, 3)
	if userID != "" {
		attrs = append(attrs, AttrUserID.String(userID))
	}
	if tenantID != "" {
		attrs = append(attrs, AttrTenantID.String(tenantID))
	}
	if correlationID != "" {
		attrs = append(attrs, AttrCorrelationID.String(correlationID))
	}

	span.SetAttributes(attrs...)
}
