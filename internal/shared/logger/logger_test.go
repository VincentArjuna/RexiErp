package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.Equal(t, logrus.InfoLevel.String(), config.Level)
	assert.Equal(t, "json", config.Format)
	assert.Equal(t, "rexi-erp", config.ServiceName)
	assert.Equal(t, "1.0.0", config.Version)
	assert.Equal(t, "development", config.Environment)
}

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectedLvl logrus.Level
		expectedFmt string
	}{
		{
			name:        "nil config uses defaults",
			config:      nil,
			expectedLvl: logrus.InfoLevel,
			expectedFmt: "json",
		},
		{
			name: "custom config",
			config: &Config{
				Level:       "debug",
				Format:      "text",
				ServiceName: "test-service",
				Version:     "2.0.0",
				Environment: "staging",
			},
			expectedLvl: logrus.DebugLevel,
			expectedFmt: "text",
		},
		{
			name: "production disables debug",
			config: &Config{
				Level:       "debug",
				Format:      "json",
				ServiceName: "prod-service",
				Version:     "1.0.0",
				Environment: "production",
			},
			expectedLvl: logrus.InfoLevel, // Debug should be disabled in production
			expectedFmt: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.config)
			assert.NotNil(t, logger)
			assert.Equal(t, tt.expectedLvl, logger.GetLevel())
			assert.Equal(t, tt.expectedFmt, logger.config.Format)
		})
	}
}

func TestLoggerWithContext(t *testing.T) {
	config := &Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "test-service",
		Version:     "1.0.0",
		Environment: "test",
	}

	logger := NewLogger(config)

	// Test with service context
	entry := logger.WithServiceContext()
	assert.NotNil(t, entry)

	// Test with request context
	entry = logger.WithRequestContext("corr-123", "tenant-456", "user-789")
	assert.NotNil(t, entry)

	// Test with error
	err := errors.New("test error")
	entry = logger.WithError(err)
	assert.NotNil(t, entry)
}

func TestLoggerJSONOutput(t *testing.T) {
	var buf bytes.Buffer

	config := &Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "test-service",
		Version:     "1.0.0",
		Environment: "test",
	}

	logger := NewLogger(config)
	logger.SetOutput(&buf)

	logger.WithRequestContext("corr-123", "tenant-456", "user-789").Info("test message")

	// Parse JSON output
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "test message", logEntry["message"])
	assert.Equal(t, "info", logEntry["level"])
	assert.Equal(t, "corr-123", logEntry["correlation_id"])
	assert.Equal(t, "tenant-456", logEntry["tenant_id"])
	assert.Equal(t, "user-789", logEntry["user_id"])
	assert.Equal(t, "test-service", logEntry["service"])
	assert.Equal(t, "1.0.0", logEntry["version"])
	assert.Equal(t, "test", logEntry["environment"])
	assert.NotEmpty(t, logEntry["@timestamp"])
}

func TestSetLevel(t *testing.T) {
	logger := NewLogger(&Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "test",
		Environment: "development",
	})

	// Test setting valid level
	err := logger.SetLevel("debug")
	assert.NoError(t, err)
	assert.Equal(t, logrus.DebugLevel, logger.GetLevel())

	// Test setting invalid level
	err = logger.SetLevel("invalid")
	assert.Error(t, err)

	// Test that production prevents debug level
	prodLogger := NewLogger(&Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "test",
		Environment: "production",
	})

	err = prodLogger.SetLevel("debug")
	assert.NoError(t, err) // Should not error but level should remain info
	assert.Equal(t, logrus.InfoLevel, prodLogger.GetLevel())
}

func TestCorrelationID(t *testing.T) {
	ctx := context.Background()

	// Test adding and getting correlation ID
	corrID := GenerateCorrelationID()
	ctx = WithCorrelationID(ctx, corrID)

	retrievedID := GetCorrelationID(ctx)
	assert.Equal(t, corrID, retrievedID)

	// Test empty correlation ID
	assert.Empty(t, GetCorrelationID(context.Background()))
}

func TestTenantAndUserID(t *testing.T) {
	ctx := context.Background()

	// Test adding and getting tenant ID
	tenantID := "tenant-123"
	ctx = WithTenantID(ctx, tenantID)

	retrievedTenantID := GetTenantID(ctx)
	assert.Equal(t, tenantID, retrievedTenantID)

	// Test adding and getting user ID
	userID := "user-456"
	ctx = WithUserID(ctx, userID)

	retrievedUserID := GetUserID(ctx)
	assert.Equal(t, userID, retrievedUserID)

	// Test empty values
	assert.Empty(t, GetTenantID(context.Background()))
	assert.Empty(t, GetUserID(context.Background()))
}

func TestExtractFromHeaders(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-123")
	req.Header.Set("X-Tenant-ID", "tenant-456")
	req.Header.Set("X-User-ID", "user-789")

	corrID, tenantID, userID := ExtractFromHeaders(req)

	assert.Equal(t, "corr-123", corrID)
	assert.Equal(t, "tenant-456", tenantID)
	assert.Equal(t, "user-789", userID)
}

func TestExtractFromHeadersGeneratesCorrelationID(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)

	corrID, tenantID, userID := ExtractFromHeaders(req)

	assert.NotEmpty(t, corrID) // Should generate a new correlation ID
	assert.Empty(t, tenantID)
	assert.Empty(t, userID)
}

func TestSetInHeaders(t *testing.T) {
	header := make(http.Header)
	SetInHeaders(header, "corr-123", "tenant-456", "user-789")

	assert.Equal(t, "corr-123", header.Get("X-Correlation-ID"))
	assert.Equal(t, "tenant-456", header.Get("X-Tenant-ID"))
	assert.Equal(t, "user-789", header.Get("X-User-ID"))
}

func TestContextWithRequestHeaders(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-123")
	req.Header.Set("X-Tenant-ID", "tenant-456")

	ctx := ContextWithRequestHeaders(context.Background(), req)

	assert.Equal(t, "corr-123", GetCorrelationID(ctx))
	assert.Equal(t, "tenant-456", GetTenantID(ctx))
	assert.Empty(t, GetUserID(ctx))
}

func TestRequestLoggerContext(t *testing.T) {
	ctx := context.Background()
	ctx = RequestLoggerContext(ctx, "corr-123", "tenant-456", "user-789")

	assert.Equal(t, "corr-123", GetCorrelationID(ctx))
	assert.Equal(t, "tenant-456", GetTenantID(ctx))
	assert.Equal(t, "user-789", GetUserID(ctx))

	// Test that it generates correlation ID if not provided
	ctx2 := RequestLoggerContext(context.Background(), "", "tenant-789", "")
	assert.NotEmpty(t, GetCorrelationID(ctx2))
}

func TestHTTPMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "test-service",
		Version:     "1.0.0",
		Environment: "test",
	}

	logger := NewLogger(config)
	middleware := NewHTTPMiddleware(logger)

	// Test correlation ID middleware
	router := gin.New()
	router.Use(middleware.CorrelationID())
	router.GET("/test", func(c *gin.Context) {
		corrID := GetCorrelationID(c.Request.Context())
		assert.NotEmpty(t, corrID)
		assert.Equal(t, corrID, c.Writer.Header().Get("X-Correlation-ID"))
		c.Status(200)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := &testResponseWriter{}
	router.ServeHTTP(w, req)
}

// testResponseWriter implements gin.ResponseWriter for testing
type testResponseWriter struct {
	statusCode int
	body       bytes.Buffer
	header     http.Header
}

func (w *testResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *testResponseWriter) Write(data []byte) (int, error) {
	return w.body.Write(data)
}

func (w *testResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func TestHTTPMiddlewareWithExistingHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "test-service",
		Version:     "1.0.0",
		Environment: "test",
	}

	logger := NewLogger(config)
	middleware := NewHTTPMiddleware(logger)

	router := gin.New()
	router.Use(middleware.CorrelationID())
	router.GET("/test", func(c *gin.Context) {
		corrID := GetCorrelationID(c.Request.Context())
		tenantID := GetTenantID(c.Request.Context())
		userID := GetUserID(c.Request.Context())

		assert.Equal(t, "existing-corr-123", corrID)
		assert.Equal(t, "existing-tenant-456", tenantID)
		assert.Equal(t, "existing-user-789", userID)
		assert.Equal(t, "existing-corr-123", c.Writer.Header().Get("X-Correlation-ID"))
		assert.Equal(t, "existing-tenant-456", c.Writer.Header().Get("X-Tenant-ID"))
		c.Status(200)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Correlation-ID", "existing-corr-123")
	req.Header.Set("X-Tenant-ID", "existing-tenant-456")
	req.Header.Set("X-User-ID", "existing-user-789")

	w := &testResponseWriter{}
	router.ServeHTTP(w, req)
}

func TestStandardFields(t *testing.T) {
	// Test ServiceFields
	fields := ServiceFields("test-service", "1.0.0", "development")
	assert.Equal(t, "test-service", fields[FieldService])
	assert.Equal(t, "1.0.0", fields[FieldVersion])
	assert.Equal(t, "development", fields[FieldEnvironment])

	// Test RequestContextFields
	fields = RequestContextFields("corr-123", "tenant-456", "user-789")
	assert.Equal(t, "corr-123", fields[FieldCorrelationID])
	assert.Equal(t, "tenant-456", fields[FieldTenantID])
	assert.Equal(t, "user-789", fields[FieldUserID])

	// Test HTTPRequestFields
	fields = HTTPRequestFields("GET", "/api/test", "param=value", "127.0.0.1", "test-agent")
	assert.Equal(t, "GET", fields[FieldMethod])
	assert.Equal(t, "/api/test", fields[FieldPath])
	assert.Equal(t, "param=value", fields[FieldQuery])
	assert.Equal(t, "127.0.0.1", fields[FieldClientIP])
	assert.Equal(t, "test-agent", fields[FieldUserAgent])

	// Test ErrorFields
	err := errors.New("test error")
	fields = ErrorFields(err, "stack trace")
	assert.Equal(t, "test error", fields[FieldError])
	assert.Equal(t, "stack trace", fields[FieldStackTrace])
}

func BenchmarkLoggerCreation(b *testing.B) {
	config := &Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "test-service",
		Version:     "1.0.0",
		Environment: "test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger := NewLogger(config)
		_ = logger
	}
}

func BenchmarkLoggingWithContext(b *testing.B) {
	config := &Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "test-service",
		Version:     "1.0.0",
		Environment: "test",
	}

	logger := NewLogger(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithRequestContext("corr-123", "tenant-456", "user-789").Info("test message")
	}
}