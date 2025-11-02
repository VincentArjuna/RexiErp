package tracing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitTracer(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:   "valid config",
			config: DefaultConfig("test-service"),
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracer, closer, err := InitTracer(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, tracer)
				assert.Nil(t, closer)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, tracer)
				assert.NotNil(t, closer)

				// Cleanup
				assert.NoError(t, closer.Close())
			}
		})
	}
}

func TestStartSpan(t *testing.T) {
	_, closer, err := InitTracer(DefaultConfig("test-service"))
	require.NoError(t, err)
	defer closer.Close()

	ctx := context.Background()
	span, ctx := StartSpan(ctx, "test-operation")

	assert.NotNil(t, span)
	assert.NotNil(t, ctx)

	FinishSpan(span, nil)
}

func TestSetSpanTags(t *testing.T) {
	_, closer, err := InitTracer(DefaultConfig("test-service"))
	require.NoError(t, err)
	defer closer.Close()

	ctx := context.Background()
	span, _ := StartSpan(ctx, "test-operation")

	tags := map[string]interface{}{
		"user.id":    "123",
		"tenant.id":  "456",
		"operation":  "test",
	}

	SetSpanTags(span, tags)
	FinishSpan(span, nil)
}

func TestBaggageItems(t *testing.T) {
	_, closer, err := InitTracer(DefaultConfig("test-service"))
	require.NoError(t, err)
	defer closer.Close()

	ctx := context.Background()
	span, ctx := StartSpan(ctx, "test-operation")

	// Set baggage items
	SetBaggageItem(span, "correlation_id", "test-correlation-123")
	SetBaggageItem(span, "user_id", "test-user-456")

	// Get baggage items
	assert.Equal(t, "test-correlation-123", GetBaggageItem(ctx, "correlation_id"))
	assert.Equal(t, "test-user-456", GetBaggageItem(ctx, "user_id"))
	assert.Equal(t, "", GetBaggageItem(ctx, "non_existent"))

	FinishSpan(span, nil)
}

func TestWithSpan(t *testing.T) {
	_, closer, err := InitTracer(DefaultConfig("test-service"))
	require.NoError(t, err)
	defer closer.Close()

	ctx := context.Background()
	err = WithSpan(ctx, "test-operation", func(ctx context.Context) error {
		// Verify we have a span in context
		span := opentracing.SpanFromContext(ctx)
		assert.NotNil(t, span)
		return nil
	})

	assert.NoError(t, err)
}

func TestWithSpanError(t *testing.T) {
	_, closer, err := InitTracer(DefaultConfig("test-service"))
	require.NoError(t, err)
	defer closer.Close()

	ctx := context.Background()
	testErr := assert.AnError

	err = WithSpan(ctx, "test-operation", func(ctx context.Context) error {
		return testErr
	})

	assert.Equal(t, testErr, err)
}

func TestTraceIDExtraction(t *testing.T) {
	_, closer, err := InitTracer(DefaultConfig("test-service"))
	require.NoError(t, err)
	defer closer.Close()

	ctx := context.Background()
	span, ctx := StartSpan(ctx, "test-operation")

	traceID := TraceIDFromContext(ctx)
	assert.NotEmpty(t, traceID)

	spanID := SpanIDFromContext(ctx)
	assert.NotEmpty(t, spanID)

	FinishSpan(span, nil)
}

func TestInjectExtractContext(t *testing.T) {
	_, closer, err := InitTracer(DefaultConfig("test-service"))
	require.NoError(t, err)
	defer closer.Close()

	ctx := context.Background()
	span, ctx := StartSpan(ctx, "test-operation")
	SetBaggageItem(span, "test_key", "test_value")

	// Inject context into headers
	headers := make(map[string]string)
	err = InjectSpanContext(ctx, headers)
	require.NoError(t, err)

	// Extract context from headers
	extractedCtx, err := ExtractSpanContext(headers)
	require.NoError(t, err)
	assert.NotNil(t, extractedCtx)

	FinishSpan(span, nil)
}

func TestTimeTracker(t *testing.T) {
	_, closer, err := InitTracer(DefaultConfig("test-service"))
	require.NoError(t, err)
	defer closer.Close()

	ctx := context.Background()
	tracker := NewTimeTracker(ctx, "test-operation")

	// Simulate some work
	time.Sleep(10 * time.Millisecond)
	tracker.TrackStep("step1")

	time.Sleep(10 * time.Millisecond)
	tracker.TrackStep("step2")

	metrics := tracker.GetMetrics()
	assert.Len(t, metrics, 2)
	assert.True(t, metrics["step1"] > 0)
	assert.True(t, metrics["step2"] > metrics["step1"])

	tracker.Finish(nil)
}

func TestTraceMiddleware(t *testing.T) {
	_, closer, err := InitTracer(DefaultConfig("test-service"))
	require.NoError(t, err)
	defer closer.Close()

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := opentracing.SpanFromContext(r.Context())
		assert.NotNil(t, span)

		correlationID := GetCorrelationID(r.Context(), r.Header)
		assert.NotEmpty(t, correlationID)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with tracing middleware
	wrappedHandler := TraceMiddleware("test-service")(handler)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	wrappedHandler.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "OK", rr.Body.String())
}

func TestTraceMiddlewareWithHeaders(t *testing.T) {
	_, closer, err := InitTracer(DefaultConfig("test-service"))
	require.NoError(t, err)
	defer closer.Close()

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := GetCorrelationID(r.Context(), r.Header)
		assert.Equal(t, "test-correlation-123", correlationID)

		w.WriteHeader(http.StatusOK)
	})

	// Wrap with tracing middleware
	wrappedHandler := TraceMiddleware("test-service")(handler)

	// Create test request with existing correlation ID
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(CorrelationIDHeader, "test-correlation-123")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	wrappedHandler.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestTracedHTTPClient(t *testing.T) {
	_, closer, err := InitTracer(DefaultConfig("test-service"))
	require.NoError(t, err)
	defer closer.Close()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	// Create traced HTTP client
	client := TracedHTTPClient(nil)

	// Make request
	req, err := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestWithCorrelationID(t *testing.T) {
	_, closer, err := InitTracer(DefaultConfig("test-service"))
	require.NoError(t, err)
	defer closer.Close()

	ctx := context.Background()

	// Should add correlation ID if none exists
	ctx = WithCorrelationID(ctx)
	correlationID := GetCorrelationID(ctx, nil)
	assert.NotEmpty(t, correlationID)

	// Should return existing correlation ID
	existingID := GetCorrelationID(ctx, nil)
	assert.Equal(t, correlationID, existingID)
}

func TestSetGetCorrelationID(t *testing.T) {
	_, closer, err := InitTracer(DefaultConfig("test-service"))
	require.NoError(t, err)
	defer closer.Close()

	ctx := context.Background()
	headers := make(http.Header)

	// Start a span so we can set baggage items
	span, ctx := StartSpan(ctx, "test-span")
	defer FinishSpan(span, nil)

	// Set correlation ID
	testID := "test-correlation-123"
	ctx = SetCorrelationID(ctx, headers, testID)

	// Get from context
	assert.Equal(t, testID, GetCorrelationID(ctx, nil))

	// Get from headers
	assert.Equal(t, testID, headers.Get(CorrelationIDHeader))
}

func TestSpanOptions(t *testing.T) {
	_, closer, err := InitTracer(DefaultConfig("test-service"))
	require.NoError(t, err)
	defer closer.Close()

	ctx := context.Background()

	opts := SpanOptions{
		Tags: map[string]interface{}{
			"user.id":   "123",
			"tenant.id": "456",
		},
		Baggage: map[string]string{
			"correlation_id": "test-123",
			"session_id":     "session-456",
		},
	}

	err = WithSpanOptions(ctx, "test-operation", opts, func(ctx context.Context) error {
		// Verify tags are set (can't directly access, but span should have them)
		span := opentracing.SpanFromContext(ctx)
		assert.NotNil(t, span)

		// Verify baggage items
		assert.Equal(t, "test-123", GetBaggageItem(ctx, "correlation_id"))
		assert.Equal(t, "session-456", GetBaggageItem(ctx, "session_id"))

		return nil
	})

	assert.NoError(t, err)
}