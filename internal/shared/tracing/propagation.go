package tracing

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// HTTPHeaders represents HTTP headers for tracing
type HTTPHeaders map[string]string

// TraceMiddleware returns HTTP middleware for tracing
func TraceMiddleware(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract span context from incoming headers
			spanCtx, err := ExtractSpanContext(HTTPHeadersToMap(r.Header))
			if err != nil {
				// Start a new span if no context found
				span, ctx := StartSpan(r.Context(), fmt.Sprintf("%s.%s", serviceName, r.URL.Path))
				defer FinishSpan(span, nil)

				// Set standard tags
				SetSpanTags(span, map[string]interface{}{
					"http.method":     r.Method,
					"http.url":        r.URL.String(),
					"http.user_agent": r.UserAgent(),
					"http.remote_addr": r.RemoteAddr,
					"service.name":    serviceName,
				})

				// Continue with new context
				r = r.WithContext(ctx)
			} else {
				// Continue with existing span context
				span := opentracing.StartSpan(fmt.Sprintf("%s.%s", serviceName, r.URL.Path), ext.RPCServerOption(spanCtx))
				defer FinishSpan(span, nil)

				// Set standard tags
				SetSpanTags(span, map[string]interface{}{
					"http.method":      r.Method,
					"http.url":         r.URL.String(),
					"http.user_agent":  r.UserAgent(),
					"http.remote_addr": r.RemoteAddr,
					"service.name":     serviceName,
				})

				ctx := opentracing.ContextWithSpan(r.Context(), span)
				r = r.WithContext(ctx)
			}

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

			// Continue with the request
			next.ServeHTTP(wrapped, r)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// HTTPHeadersToMap converts http.Header to map[string]string
func HTTPHeadersToMap(headers http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range headers {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}

// InjectHTTPHeaders injects span context into HTTP headers
func InjectHTTPHeaders(ctx context.Context, headers http.Header) error {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return nil
	}

	carrier := opentracing.HTTPHeadersCarrier(headers)
	return span.Tracer().Inject(span.Context(), opentracing.HTTPHeaders, carrier)
}

// TraceHTTPRequest creates a span for an outbound HTTP request
func TraceHTTPRequest(ctx context.Context, method, url string) (context.Context, opentracing.Span) {
	span, ctx := StartSpan(ctx, fmt.Sprintf("http.%s", method))

	SetSpanTags(span, map[string]interface{}{
		"http.method": method,
		"http.url":    url,
		"span.kind":   "client",
	})

	ext.SpanKindRPCClient.Set(span)

	return ctx, span
}

// TraceHTTPResponse finishes the span for an HTTP response
func TraceHTTPResponse(span opentracing.Span, statusCode int) {
	if span != nil {
		span.SetTag("http.status_code", statusCode)
		if statusCode >= 400 {
			span.SetTag("error", true)
			ext.Error.Set(span, true)
		}
		FinishSpan(span, nil)
	}
}

// TraceHTTPRequestRoundTrip traces a complete HTTP request round trip
func TraceHTTPRequestRoundTrip(ctx context.Context, req *http.Request, transport http.RoundTripper) (*http.Response, error) {
	// Create span for the request
	spanCtx, span := TraceHTTPRequest(ctx, req.Method, req.URL.String())
	defer func() {
		if span != nil {
			FinishSpan(span, nil)
		}
	}()

	// Inject span context into request headers
	if err := InjectHTTPHeaders(spanCtx, req.Header); err != nil {
		// Log error but continue with request
		span.SetTag("inject.error", err.Error())
	}

	// Execute the request
	resp, err := transport.RoundTrip(req)
	if err != nil {
		if span != nil {
			span.SetTag("error", true)
			span.SetTag("error.message", err.Error())
		}
		return nil, err
	}

	// Set response status code
	if span != nil {
		span.SetTag("http.status_code", resp.StatusCode)
		if resp.StatusCode >= 400 {
			span.SetTag("error", true)
		}
	}

	return resp, nil
}

// TracedHTTPClient returns an HTTP client with tracing
func TracedHTTPClient(baseTransport http.RoundTripper) *http.Client {
	if baseTransport == nil {
		baseTransport = http.DefaultTransport
	}

	tracedTransport := &tracingTransport{
		base: baseTransport,
	}

	return &http.Client{
		Transport: tracedTransport,
	}
}

type tracingTransport struct {
	base http.RoundTripper
}

func (t *tracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	return TraceHTTPRequestRoundTrip(ctx, req, t.base)
}

// CorrelationIDHeader is the header name for correlation ID
const CorrelationIDHeader = "X-Correlation-ID"

// GetCorrelationID extracts correlation ID from context or headers
func GetCorrelationID(ctx context.Context, headers http.Header) string {
	// First try to get from context
	if correlationID := GetBaggageItem(ctx, "correlation_id"); correlationID != "" {
		return correlationID
	}

	// Then try to get from headers
	if headers != nil {
		if correlationID := headers.Get(CorrelationIDHeader); correlationID != "" {
			return correlationID
		}
	}

	// Finally try to extract from span
	if traceID := TraceIDFromContext(ctx); traceID != "" {
		return traceID
	}

	return ""
}

// SetCorrelationID sets correlation ID in context and headers
func SetCorrelationID(ctx context.Context, headers http.Header, correlationID string) context.Context {
	// Set in span baggage
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		SetBaggageItem(span, "correlation_id", correlationID)
	}

	// Set in headers
	if headers != nil {
		headers.Set(CorrelationIDHeader, correlationID)
	}

	return ctx
}

// WithCorrelationID ensures a correlation ID exists in the context
func WithCorrelationID(ctx context.Context) context.Context {
	if correlationID := GetCorrelationID(ctx, nil); correlationID != "" {
		return ctx
	}

	// Generate new correlation ID if none exists
	span, ctx := StartSpan(ctx, "generate_correlation_id")
	defer FinishSpan(span, nil)

	correlationID := TraceIDFromContext(ctx)
	if correlationID == "" {
		correlationID = generateUUID()
	}

	return SetCorrelationID(ctx, nil, correlationID)
}

func generateUUID() string {
	// Simple UUID generation - in production, use a proper UUID library
	return fmt.Sprintf("%d", time.Now().UnixNano())
}