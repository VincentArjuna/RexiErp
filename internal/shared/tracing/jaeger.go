package tracing

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
)

// Config holds Jaeger configuration
type Config struct {
	ServiceName string
	AgentHost   string
	AgentPort   string
	SamplerType string
	SamplerParam float64
}

// DefaultConfig returns default Jaeger configuration
func DefaultConfig(serviceName string) *Config {
	return &Config{
		ServiceName: serviceName,
		AgentHost:   "jaeger-agent",
		AgentPort:   "6831",
		SamplerType: jaeger.SamplerTypeConst,
		SamplerParam: 1.0, // Sample all traces in development
	}
}

// InitTracer initializes Jaeger tracer
func InitTracer(cfg *Config) (opentracing.Tracer, io.Closer, error) {
	if cfg == nil {
		return nil, nil, fmt.Errorf("tracing config cannot be nil")
	}

	// Jaeger configuration
	jaegerConfig := jaegercfg.Configuration{
		ServiceName: cfg.ServiceName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  cfg.SamplerType,
			Param: cfg.SamplerParam,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: fmt.Sprintf("%s:%s", cfg.AgentHost, cfg.AgentPort),
		},
	}

	// Initialize tracer with logger and metrics factory
	tracer, closer, err := jaegerConfig.NewTracer(
		jaegercfg.Logger(jaegerlog.StdLogger),
		jaegercfg.Metrics(metrics.NullFactory),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize Jaeger tracer: %w", err)
	}

	// Set as global tracer
	opentracing.SetGlobalTracer(tracer)

	return tracer, closer, nil
}

// StartSpan starts a new span with the given operation name
func StartSpan(ctx context.Context, operationName string) (opentracing.Span, context.Context) {
	return opentracing.StartSpanFromContext(ctx, operationName)
}

// StartSpanWithParent starts a new span with the given operation name and parent context
func StartSpanWithParent(ctx context.Context, operationName string, parentCtx opentracing.SpanContext) (opentracing.Span, context.Context) {
	span := opentracing.StartSpan(operationName, opentracing.ChildOf(parentCtx))
	return span, opentracing.ContextWithSpan(ctx, span)
}

// SetSpanTags sets tags on the span
func SetSpanTags(span opentracing.Span, tags map[string]interface{}) {
	for key, value := range tags {
		span.SetTag(key, value)
	}
}

// SetBaggageItem sets a baggage item on the span
func SetBaggageItem(span opentracing.Span, key, value string) {
	span.SetBaggageItem(key, value)
}

// GetBaggageItem gets a baggage item from the span context
func GetBaggageItem(ctx context.Context, key string) string {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return ""
	}
	return span.BaggageItem(key)
}

// FinishSpan finishes the span with optional error
func FinishSpan(span opentracing.Span, err error) {
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.message", err.Error())
	}
	span.Finish()
}

// WithSpan executes a function within a span
func WithSpan(ctx context.Context, operationName string, fn func(context.Context) error) error {
	span, ctx := StartSpan(ctx, operationName)
	defer FinishSpan(span, nil)

	return fn(ctx)
}

// WithSpanOptions executes a function within a span with options
func WithSpanOptions(ctx context.Context, operationName string, opts SpanOptions, fn func(context.Context) error) error {
	span, ctx := StartSpan(ctx, operationName)
	defer FinishSpan(span, nil)

	// Apply tags
	if opts.Tags != nil {
		SetSpanTags(span, opts.Tags)
	}

	// Apply baggage items
	if opts.Baggage != nil {
		for key, value := range opts.Baggage {
			SetBaggageItem(span, key, value)
		}
	}

	return fn(ctx)
}

// SpanOptions holds options for span creation
type SpanOptions struct {
	Tags    map[string]interface{}
	Baggage map[string]string
}

// TraceIDFromContext extracts trace ID from context
func TraceIDFromContext(ctx context.Context) string {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return ""
	}

	if spanContext, ok := span.Context().(jaeger.SpanContext); ok {
		return spanContext.TraceID().String()
	}

	return ""
}

// SpanIDFromContext extracts span ID from context
func SpanIDFromContext(ctx context.Context) string {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return ""
	}

	if spanContext, ok := span.Context().(jaeger.SpanContext); ok {
		return spanContext.SpanID().String()
	}

	return ""
}

// InjectSpanContext injects span context into headers map
func InjectSpanContext(ctx context.Context, headers map[string]string) error {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return nil
	}

	carrier := opentracing.TextMapCarrier(headers)
	return span.Tracer().Inject(span.Context(), opentracing.TextMap, carrier)
}

// ExtractSpanContext extracts span context from headers map
func ExtractSpanContext(headers map[string]string) (opentracing.SpanContext, error) {
	tracer := opentracing.GlobalTracer()
	carrier := opentracing.TextMapCarrier(headers)
	return tracer.Extract(opentracing.TextMap, carrier)
}

// TimeTracker tracks duration for operations
type TimeTracker struct {
	span    opentracing.Span
	start   time.Time
	metrics map[string]time.Duration
}

// NewTimeTracker creates a new time tracker
func NewTimeTracker(ctx context.Context, operationName string) *TimeTracker {
	span, _ := StartSpan(ctx, operationName)
	return &TimeTracker{
		span:    span,
		start:   time.Now(),
		metrics: make(map[string]time.Duration),
	}
}

// TrackStep tracks a step within the operation
func (tt *TimeTracker) TrackStep(stepName string) {
	if tt.metrics != nil {
		tt.metrics[stepName] = time.Since(tt.start)
	}
	if tt.span != nil {
		tt.span.SetTag(fmt.Sprintf("step.%s.duration_ms", stepName), time.Since(tt.start).Milliseconds())
	}
}

// Finish finishes the time tracker and span
func (tt *TimeTracker) Finish(err error) {
	duration := time.Since(tt.start)
	if tt.span != nil {
		tt.span.SetTag("duration_ms", duration.Milliseconds())
		FinishSpan(tt.span, err)
	}
}

// GetMetrics returns all tracked metrics
func (tt *TimeTracker) GetMetrics() map[string]time.Duration {
	return tt.metrics
}