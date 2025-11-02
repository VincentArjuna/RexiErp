package metrics

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/VincentArjuna/RexiErp/internal/shared/logger"
)

// MetricsMiddleware provides Gin middleware for Prometheus metrics collection
type MetricsMiddleware struct {
	metrics *PrometheusMetrics
	logger  *logger.Logger
}

// NewMetricsMiddleware creates a new metrics middleware instance
func NewMetricsMiddleware(metrics *PrometheusMetrics, log *logger.Logger) *MetricsMiddleware {
	return &MetricsMiddleware{
		metrics: metrics,
		logger:  log,
	}
}

// HTTPMiddleware returns a Gin middleware for HTTP metrics collection
func (mm *MetricsMiddleware) HTTPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Extract tenant ID from context or headers
		tenantID := logger.GetTenantID(c.Request.Context())
		if tenantID == "" {
			tenantID = c.GetHeader("X-Tenant-ID")
		}

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)
		durationSeconds := duration.Seconds()

		// Get response details
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := normalizePath(c.Request.URL.Path)

		// Update HTTP metrics
		labels := map[string]string{
			"method":      method,
			"path":        path,
			"status_code": strconv.Itoa(statusCode),
		}

		// Add tenant_id to labels if available
		if tenantID != "" {
			labels["tenant_id"] = tenantID
			mm.metrics.httpRequestsTotal.WithLabelValues(method, path, strconv.Itoa(statusCode), tenantID).Inc()
			mm.metrics.httpRequestDuration.WithLabelValues(method, path, strconv.Itoa(statusCode), tenantID).Observe(durationSeconds)
		} else {
			mm.metrics.httpRequestsTotal.WithLabelValues(method, path, strconv.Itoa(statusCode), "").Inc()
			mm.metrics.httpRequestDuration.WithLabelValues(method, path, strconv.Itoa(statusCode), "").Observe(durationSeconds)
		}

		// Update response size metric
		responseSize := float64(c.Writer.Size())
		mm.metrics.httpResponseSize.WithLabelValues(method, path, strconv.Itoa(statusCode)).Observe(responseSize)

		// Update API calls metric
		if tenantID != "" {
			mm.metrics.apiCalls.WithLabelValues(path, method, tenantID, getStatusCategory(statusCode)).Inc()
		} else {
			mm.metrics.apiCalls.WithLabelValues(path, method, "", getStatusCategory(statusCode)).Inc()
		}

		// Log slow requests
		if duration > time.Second {
			correlationID := logger.GetCorrelationID(c.Request.Context())
			logEntry := mm.logger.WithRequestContext(correlationID, tenantID, "")
			logEntry.WithFields(map[string]interface{}{
				"method":          method,
				"path":            path,
				"status_code":     statusCode,
				"duration_ms":     duration.Milliseconds(),
				"response_size":   responseSize,
				"performance_type": "slow_request",
			}).Warn("Slow HTTP request detected")
		}
	}
}

// DatabaseMiddleware creates middleware for database metrics collection
type DatabaseMiddleware struct {
	metrics *PrometheusMetrics
	logger  *logger.Logger
}

// NewDatabaseMiddleware creates a new database middleware instance
func NewDatabaseMiddleware(metrics *PrometheusMetrics, log *logger.Logger) *DatabaseMiddleware {
	return &DatabaseMiddleware{
		metrics: metrics,
		logger:  log,
	}
}

// RecordQuery records a database query
func (dm *DatabaseMiddleware) RecordQuery(ctx context.Context, table, operation string, duration time.Duration, tenantID string, success bool) {
	durationSeconds := duration.Seconds()
	status := "success"
	if !success {
		status = "error"
	}

	// Update database metrics
	if tenantID != "" {
		dm.metrics.dbQueryDuration.WithLabelValues(table, operation, tenantID).Observe(durationSeconds)
		dm.metrics.dbQueriesTotal.WithLabelValues(table, operation, status, tenantID).Inc()
	} else {
		dm.metrics.dbQueryDuration.WithLabelValues(table, operation, "").Observe(durationSeconds)
		dm.metrics.dbQueriesTotal.WithLabelValues(table, operation, status, "").Inc()
	}

	// Log slow queries
	if duration > time.Millisecond*100 {
		correlationID := logger.GetCorrelationID(ctx)
		logEntry := dm.logger.WithRequestContext(correlationID, tenantID, "")
		logEntry.WithFields(map[string]interface{}{
			"table":           table,
			"operation":       operation,
			"duration_ms":     duration.Milliseconds(),
			"performance_type": "slow_query",
		}).Warn("Slow database query detected")
	}
}

// RecordConnectionError records a database connection error
func (dm *DatabaseMiddleware) RecordConnectionError() {
	dm.metrics.dbConnectionErrors.Inc()
}

// UpdateConnectionCount updates the active database connections count
func (dm *DatabaseMiddleware) UpdateConnectionCount(count float64) {
	dm.metrics.dbConnections.Set(count)
}

// CacheMiddleware creates middleware for cache metrics collection
type CacheMiddleware struct {
	metrics *PrometheusMetrics
	logger  *logger.Logger
}

// NewCacheMiddleware creates a new cache middleware instance
func NewCacheMiddleware(metrics *PrometheusMetrics, log *logger.Logger) *CacheMiddleware {
	return &CacheMiddleware{
		metrics: metrics,
		logger:  log,
	}
}

// RecordHit records a cache hit
func (cm *CacheMiddleware) RecordHit(cacheType, keyPrefix string, tenantID string) {
	cm.metrics.cacheHits.WithLabelValues(cacheType, keyPrefix).Inc()
}

// RecordMiss records a cache miss
func (cm *CacheMiddleware) RecordMiss(cacheType, keyPrefix string, tenantID string) {
	cm.metrics.cacheMisses.WithLabelValues(cacheType, keyPrefix).Inc()
}

// ExternalServiceMiddleware creates middleware for external service metrics collection
type ExternalServiceMiddleware struct {
	metrics *PrometheusMetrics
	logger  *logger.Logger
}

// NewExternalServiceMiddleware creates a new external service middleware instance
func NewExternalServiceMiddleware(metrics *PrometheusMetrics, log *logger.Logger) *ExternalServiceMiddleware {
	return &ExternalServiceMiddleware{
		metrics: metrics,
		logger:  log,
	}
}

// RecordCall records an external service call
func (esm *ExternalServiceMiddleware) RecordCall(ctx context.Context, serviceName, endpoint string, duration time.Duration, success bool) {
	durationSeconds := duration.Seconds()
	status := "success"
	if !success {
		status = "error"
	}

	// Update external service metrics
	esm.metrics.externalServiceCalls.WithLabelValues(serviceName, endpoint, status).Inc()
	esm.metrics.externalServiceDuration.WithLabelValues(serviceName, endpoint).Observe(durationSeconds)

	// Log slow external calls
	if duration > time.Second*5 {
		correlationID := logger.GetCorrelationID(ctx)
		logEntry := esm.logger.WithRequestContext(correlationID, "", "")
		logEntry.WithFields(map[string]interface{}{
			"service":         serviceName,
			"endpoint":        endpoint,
			"duration_ms":     duration.Milliseconds(),
			"performance_type": "slow_external_call",
		}).Warn("Slow external service call detected")
	}
}

// BusinessMetrics provides methods for recording business metrics
type BusinessMetrics struct {
	metrics *PrometheusMetrics
	logger  *logger.Logger
}

// NewBusinessMetrics creates a new business metrics instance
func NewBusinessMetrics(metrics *PrometheusMetrics, log *logger.Logger) *BusinessMetrics {
	return &BusinessMetrics{
		metrics: metrics,
		logger:  log,
	}
}

// RecordTransaction records a business transaction
func (bm *BusinessMetrics) RecordTransaction(ctx context.Context, transactionType, transactionID string, amount float64, tenantID string, success bool) {
	status := "success"
	if !success {
		status = "failed"
	}

	// Update transaction metrics
	if tenantID != "" {
		bm.metrics.totalTransactions.WithLabelValues(transactionType, tenantID, status).Inc()
		bm.metrics.transactionAmounts.WithLabelValues(transactionType, tenantID).Observe(amount)
	} else {
		bm.metrics.totalTransactions.WithLabelValues(transactionType, "", status).Inc()
		bm.metrics.transactionAmounts.WithLabelValues(transactionType, "").Observe(amount)
	}

	// Log significant transactions
	if amount > 1000000 { // > 1 million IDR
		correlationID := logger.GetCorrelationID(ctx)
		logEntry := bm.logger.WithRequestContext(correlationID, tenantID, "")
		logEntry.WithFields(map[string]interface{}{
			"transaction_type": transactionType,
			"transaction_id":   transactionID,
			"amount":           amount,
			"business_type":    "high_value_transaction",
		}).Info("High value transaction recorded")
	}
}

// RecordUserSession records a user session
func (bm *BusinessMetrics) RecordUserSession(tenantID string, success bool) {
	status := "success"
	if !success {
		status = "failed"
	}

	if tenantID != "" {
		bm.metrics.userSessions.WithLabelValues(tenantID, status).Inc()
	} else {
		bm.metrics.userSessions.WithLabelValues("", status).Inc()
	}
}

// UpdateActiveTenants updates the active tenants count
func (bm *BusinessMetrics) UpdateActiveTenants(count float64) {
	bm.metrics.activeTenants.Set(count)
}

// normalizePath normalizes URL paths for metrics by removing IDs and other variables
func normalizePath(path string) string {
	// Remove UUIDs from paths
	path = strings.ReplaceAll(path, "/[a-f0-9-]{36}", "/:id")

	// Remove numeric IDs
	path = strings.ReplaceAll(path, "/\\d+", "/:id")

	// Remove common parameter patterns
	path = strings.ReplaceAll(path, "/[^/]+/[^/]+$", "/:id")

	// Normalize specific API patterns
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		if len(segment) > 0 && segment[0] == ':' {
			segments[i] = ":param"
		}
		// Convert numeric segments to :id
		if _, err := strconv.Atoi(segment); err == nil {
			segments[i] = ":id"
		}
		// Convert UUID-like segments to :id
		if len(segment) == 36 && strings.Count(segment, "-") == 4 {
			segments[i] = ":id"
		}
	}

	return strings.Join(segments, "/")
}

// getStatusCategory converts HTTP status codes to categories
func getStatusCategory(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return "success"
	case statusCode >= 300 && statusCode < 400:
		return "redirect"
	case statusCode >= 400 && statusCode < 500:
		return "client_error"
	case statusCode >= 500:
		return "server_error"
	default:
		return "unknown"
	}
}