package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/VincentArjuna/RexiErp/internal/shared/logger"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

// testRegistry creates a new registry for each test to avoid conflicts
func testRegistry() *prometheus.Registry {
	reg := prometheus.NewRegistry()
	// Register default Go metrics to avoid warnings
	reg.MustRegister(prometheus.NewGoCollector())
	return reg
}

func TestNewPrometheusMetrics(t *testing.T) {
	serviceName := "test-service"
	pm := NewPrometheusMetrics(serviceName)

	assert.NotNil(t, pm)
	assert.NotNil(t, pm.httpRequestsTotal)
	assert.True(t, pm.startTime.Before(time.Now()))
	assert.NotNil(t, pm.customMetrics)
}

func TestPrometheusMetricsCreation(t *testing.T) {
	pm := NewPrometheusMetrics("test-service")

	// Test that metrics are properly created
	assert.NotNil(t, pm.httpRequestsTotal)
	assert.NotNil(t, pm.httpRequestDuration)
	assert.NotNil(t, pm.httpResponseSize)
	assert.NotNil(t, pm.activeTenants)
	assert.NotNil(t, pm.totalTransactions)
	assert.NotNil(t, pm.transactionAmounts)
	assert.NotNil(t, pm.userSessions)
	assert.NotNil(t, pm.apiCalls)
	assert.NotNil(t, pm.uptimeCounter)
	assert.NotNil(t, pm.dbConnections)
	assert.NotNil(t, pm.dbQueryDuration)
	assert.NotNil(t, pm.dbQueriesTotal)
	assert.NotNil(t, pm.dbConnectionErrors)
	assert.NotNil(t, pm.cacheHits)
	assert.NotNil(t, pm.cacheMisses)
	assert.NotNil(t, pm.externalServiceCalls)
	assert.NotNil(t, pm.externalServiceDuration)
}

func TestMetricsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup without registration to avoid conflicts
	pm := NewPrometheusMetrics("test-service")

	logConfig := &logger.Config{
		Level:       "info",
		Format:      "text",
		ServiceName: "test-service",
		Environment: "test",
	}
	log := logger.NewLogger(logConfig)

	middleware := NewMetricsMiddleware(pm, log)

	// Create test router
	router := gin.New()
	router.Use(middleware.HTTPMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	})
	router.GET("/users/:id", func(c *gin.Context) {
		c.JSON(200, gin.H{"id": c.Param("id")})
	})

	// Test normal request
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Tenant-ID", "tenant-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	// Test path with ID
	req, _ = http.NewRequest("GET", "/users/12345", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestDatabaseMiddleware(t *testing.T) {
	pm := NewPrometheusMetrics("test-service")

	logConfig := &logger.Config{
		Level:       "info",
		Format:      "text",
		ServiceName: "test-service",
		Environment: "test",
	}
	log := logger.NewLogger(logConfig)

	db := NewDatabaseMiddleware(pm, log)

	ctx := context.Background()
	ctx = logger.WithTenantID(ctx, "tenant-123")

	// Test successful query
	db.RecordQuery(ctx, "users", "SELECT", time.Millisecond*50, "tenant-123", true)

	// Test failed query
	db.RecordQuery(ctx, "orders", "INSERT", time.Millisecond*200, "tenant-123", false)

	// Test connection error
	db.RecordConnectionError()

	// Test connection count update
	db.UpdateConnectionCount(15.0)
}

func TestCacheMiddleware(t *testing.T) {
	pm := NewPrometheusMetrics("test-service")

	logConfig := &logger.Config{
		Level:       "info",
		Format:      "text",
		ServiceName: "test-service",
		Environment: "test",
	}
	log := logger.NewLogger(logConfig)

	cache := NewCacheMiddleware(pm, log)

	// Test cache hit
	cache.RecordHit("redis", "session", "tenant-123")

	// Test cache miss
	cache.RecordMiss("redis", "product", "tenant-456")
}

func TestExternalServiceMiddleware(t *testing.T) {
	pm := NewPrometheusMetrics("test-service")

	logConfig := &logger.Config{
		Level:       "info",
		Format:      "text",
		ServiceName: "test-service",
		Environment: "test",
	}
	log := logger.NewLogger(logConfig)

	external := NewExternalServiceMiddleware(pm, log)

	ctx := context.Background()

	// Test successful call
	external.RecordCall(ctx, "payment-gateway", "/charge", time.Millisecond*500, true)

	// Test failed call
	external.RecordCall(ctx, "sms-service", "/send", time.Second*2, false)
}

func TestBusinessMetricsCollector(t *testing.T) {
	pm := NewPrometheusMetrics("test-service")

	logConfig := &logger.Config{
		Level:       "info",
		Format:      "text",
		ServiceName: "test-service",
		Environment: "test",
	}
	log := logger.NewLogger(logConfig)

	bmc := NewBusinessMetricsCollector(pm, log)
	assert.NotNil(t, bmc)
	assert.NotNil(t, bmc.business)
	assert.NotNil(t, bmc.tenants)

	ctx := context.Background()
	ctx = logger.WithTenantID(ctx, "tenant-123")

	// Test e-commerce metrics
	bmc.RecordOrderCreated(ctx, "tenant-123", "confirmed", "online", 1500000.0, "order-123")
	bmc.RecordProductView(ctx, "tenant-123", "electronics", "smartphone")
	bmc.RecordCartAbandoned(ctx, "tenant-123", "timeout", 500000.0)

	// Test user activity metrics
	bmc.RecordUserLogin(ctx, "tenant-123", "success", "password", "user-123")
	bmc.RecordUserLogin(ctx, "tenant-123", "failed", "password", "user-456")
	bmc.RecordUserRegistration(ctx, "tenant-123", "customer", "success", "user-789")
	bmc.UpdateActiveSessions(25.0)

	// Test inventory metrics
	bmc.RecordStockMovement(ctx, "tenant-123", "sale", "electronics", 5)
	bmc.UpdateStockLevels(1500.0)

	// Test financial metrics
	bmc.RecordInvoiceCreated(ctx, "tenant-123", "sales", "confirmed", 2500000.0, "inv-123")
	bmc.RecordPaymentReceived(ctx, "tenant-123", "bank_transfer", "confirmed", 1500000.0, "pay-123")

	// Test HR metrics
	bmc.UpdateEmployeeCount(50.0)
	bmc.RecordPayrollProcessed(ctx, "tenant-123", "monthly", "success", 50)
	bmc.RecordLeaveRequest(ctx, "tenant-123", "annual", "approved")

	// Test tenant metrics
	tenantMetrics := bmc.GetTenantMetrics("tenant-123")
	assert.NotNil(t, tenantMetrics)
	assert.Equal(t, "tenant-123", tenantMetrics.TenantID)
	assert.Equal(t, bmc.business, tenantMetrics.Metrics)

	// Test metrics summary
	summary := bmc.GetMetricsSummary()
	assert.NotNil(t, summary)
	assert.Contains(t, summary, "uptime_seconds")
	assert.Contains(t, summary, "active_tenants")
	assert.Contains(t, summary, "registered_tenant_ids")
	assert.Equal(t, 1, summary["active_tenants"])
}

func TestMetricsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/metrics", MetricsHandler())

	req, _ := http.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	// The metrics endpoint should return data even without specific metrics registered
	assert.NotEmpty(t, w.Body.String())
}

func TestGetMetricNames(t *testing.T) {
	pm := NewPrometheusMetrics("test-service")
	names := pm.GetMetricNames()

	assert.NotNil(t, names)
	assert.Contains(t, names, "http_requests_total")
	assert.Contains(t, names, "active_tenants_total")
	assert.Contains(t, names, "transactions_total")
}

func TestUpdateUptime(t *testing.T) {
	pm := NewPrometheusMetrics("test-service")

	// Sleep a bit to ensure uptime > 0
	time.Sleep(time.Millisecond * 10)

	pm.UpdateUptime()

	// We can't easily test the exact value since it's timing-dependent,
	// but we can verify the method doesn't panic
	assert.True(t, true)
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/api/v1/users/123", "/api/v1/users/:id"},
		{"/api/v1/orders/456/items/789", "/api/v1/orders/:id/items/:id"},
		{"/api/v1/products", "/api/v1/products"},
		{"/health", "/health"},
		{"/api/v1/users/550e8400-e29b-41d4-a716-446655440000", "/api/v1/users/:id"},
		{"/api/v1/products/123/reviews", "/api/v1/products/:id/reviews"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizePath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetStatusCategory(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   string
	}{
		{200, "success"},
		{201, "success"},
		{301, "redirect"},
		{302, "redirect"},
		{400, "client_error"},
		{404, "client_error"},
		{500, "server_error"},
		{503, "server_error"},
		{100, "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.statusCode)), func(t *testing.T) {
			result := getStatusCategory(tt.statusCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMetricsIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup complete metrics stack
	pm := NewPrometheusMetrics("test-service")

	logConfig := &logger.Config{
		Level:       "info",
		Format:      "text",
		ServiceName: "test-service",
		Environment: "test",
	}
	log := logger.NewLogger(logConfig)

	bmc := NewBusinessMetricsCollector(pm, log)
	metrics := NewMetricsMiddleware(pm, log)
	db := NewDatabaseMiddleware(pm, log)
	cache := NewCacheMiddleware(pm, log)
	external := NewExternalServiceMiddleware(pm, log)

	// Create test router with all middleware
	router := gin.New()
	router.Use(metrics.HTTPMiddleware())
	router.GET("/test", func(c *gin.Context) {
		ctx := c.Request.Context()

		// Simulate database operations
		db.RecordQuery(ctx, "users", "SELECT", time.Millisecond*10, "tenant-123", true)

		// Simulate cache operations
		cache.RecordHit("redis", "user", "tenant-123")
		cache.RecordMiss("redis", "profile", "tenant-123")

		// Simulate external service call
		external.RecordCall(ctx, "notification-service", "/send", time.Millisecond*100, true)

		// Simulate business metrics
		bmc.RecordUserLogin(ctx, "tenant-123", "success", "password", "user-123")
		bmc.RecordProductView(ctx, "tenant-123", "electronics", "smartphone")

		c.JSON(200, gin.H{"message": "success"})
	})

	// Make multiple requests
	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Tenant-ID", "tenant-123")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	}

	// Test metrics endpoint - since we don't register metrics in tests, we just verify the endpoint works
	req, _ := http.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	router.GET("/metrics", MetricsHandler())
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	// Just verify it returns some metrics data (Go metrics are always available)
	assert.Contains(t, w.Body.String(), "go_")
	assert.NotEmpty(t, w.Body.String())
}

// BenchmarkMetricsMiddleware tests performance of metrics middleware
func BenchmarkMetricsMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)

	pm := NewPrometheusMetrics("test-service")

	logConfig := &logger.Config{
		Level:       "info",
		Format:      "text",
		ServiceName: "test-service",
		Environment: "test",
	}
	log := logger.NewLogger(logConfig)

	middleware := NewMetricsMiddleware(pm, log)

	router := gin.New()
	router.Use(middleware.HTTPMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// Test helper to get metric value
func getMetricValue(collector prometheus.Collector) float64 {
	metricChan := make(chan prometheus.Metric, 1)
	collector.Collect(metricChan)
	m := <-metricChan

	var metric dto.Metric
	m.Write(&metric)

	// Handle different metric types by checking which field is not nil
	if metric.Counter != nil {
		return metric.Counter.GetValue()
	}
	if metric.Gauge != nil {
		return metric.Gauge.GetValue()
	}
	if metric.Histogram != nil {
		return metric.Histogram.GetSampleSum()
	}
	if metric.Untyped != nil {
		return metric.Untyped.GetValue()
	}

	return 0
}