package logger

import (
	"context"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// HTTPMiddleware provides Gin middleware for logging and correlation ID propagation
type HTTPMiddleware struct {
	logger *Logger
}

// NewHTTPMiddleware creates a new HTTP middleware instance
func NewHTTPMiddleware(logger *Logger) *HTTPMiddleware {
	return &HTTPMiddleware{
		logger: logger,
	}
}

// RequestLogging middleware logs HTTP requests with correlation IDs
func (m *HTTPMiddleware) RequestLogging() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Extract context information
		correlationID := GetCorrelationID(param.Request.Context())
		tenantID := GetTenantID(param.Request.Context())
		userID := GetUserID(param.Request.Context())

		// Create log entry
		entry := m.logger.WithRequestContext(correlationID, tenantID, userID)

		// Add HTTP request fields
		fields := logrus.Fields{
			"method":           param.Method,
			"path":             param.Path,
			"status_code":      param.StatusCode,
			"latency":          param.Latency.String(),
			"client_ip":        param.ClientIP,
			"user_agent":       param.Request.UserAgent(),
			"request_size":     param.Request.ContentLength,
			"response_size":    param.BodySize,
		}

		// Add error if present
		if param.ErrorMessage != "" {
			fields["error"] = param.ErrorMessage
		}

		// Determine log level based on status code
		if param.StatusCode >= 500 {
			entry.WithFields(fields).Error("HTTP request completed with server error")
		} else if param.StatusCode >= 400 {
			entry.WithFields(fields).Warn("HTTP request completed with client error")
		} else {
			entry.WithFields(fields).Info("HTTP request completed")
		}

		// Return empty string since we're using structured logging
		return ""
	})
}

// CorrelationID middleware ensures correlation ID propagation
func (m *HTTPMiddleware) CorrelationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract correlation ID from headers or generate new one
		correlationID, tenantID, userID := ExtractFromHeaders(c.Request)

		// Add to context
		ctx := WithCorrelationID(c.Request.Context(), correlationID)
		if tenantID != "" {
			ctx = WithTenantID(ctx, tenantID)
		}
		if userID != "" {
			ctx = WithUserID(ctx, userID)
		}

		// Update request context
		c.Request = c.Request.WithContext(ctx)

		// Set response headers
		c.Header("X-Correlation-ID", correlationID)
		if tenantID != "" {
			c.Header("X-Tenant-ID", tenantID)
		}

		// Process request
		c.Next()
	}
}

// Combined middleware that provides both correlation ID propagation and request logging
func (m *HTTPMiddleware) Combined() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set start time for latency calculation
		startTime := time.Now()

		// Extract correlation ID from headers or generate new one
		correlationID, tenantID, userID := ExtractFromHeaders(c.Request)

		// Add to context
		ctx := WithCorrelationID(c.Request.Context(), correlationID)
		if tenantID != "" {
			ctx = WithTenantID(ctx, tenantID)
		}
		if userID != "" {
			ctx = WithUserID(ctx, userID)
		}

		// Update request context
		c.Request = c.Request.WithContext(ctx)

		// Set response headers
		c.Header("X-Correlation-ID", correlationID)
		if tenantID != "" {
			c.Header("X-Tenant-ID", tenantID)
		}

		// Create log entry
		entry := m.logger.WithRequestContext(correlationID, tenantID, userID)

		// Log request start
		entry.WithFields(logrus.Fields{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"query":      c.Request.URL.RawQuery,
			"client_ip":  c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		}).Info("HTTP request started")

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(startTime)

		// Create response log fields
		fields := logrus.Fields{
			"method":           c.Request.Method,
			"path":             c.Request.URL.Path,
			"status_code":      c.Writer.Status(),
			"latency":          latency.String(),
			"latency_ms":       latency.Milliseconds(),
			"client_ip":        c.ClientIP(),
			"user_agent":       c.Request.UserAgent(),
			"request_size":     c.Request.ContentLength,
			"response_size":    c.Writer.Size(),
		}

		// Add query parameters if present
		if c.Request.URL.RawQuery != "" {
			fields["query"] = c.Request.URL.RawQuery
		}

		// Log request completion
		if c.Writer.Status() >= 500 {
			if len(c.Errors) > 0 {
				fields["error"] = c.Errors.String()
			}
			entry.WithFields(fields).Error("HTTP request completed with server error")
		} else if c.Writer.Status() >= 400 {
			entry.WithFields(fields).Warn("HTTP request completed with client error")
		} else {
			entry.WithFields(fields).Info("HTTP request completed")
		}
	}
}

// PanicRecovery middleware recovers from panics and logs them
func (m *HTTPMiddleware) PanicRecovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		correlationID := GetCorrelationID(c.Request.Context())
		tenantID := GetTenantID(c.Request.Context())
		userID := GetUserID(c.Request.Context())

		entry := m.logger.WithRequestContext(correlationID, tenantID, userID)

		entry.WithFields(logrus.Fields{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"client_ip":  c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"panic":      recovered,
			"stack":      string(debug.Stack()),
		}).Error("HTTP request panic recovered")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":          "Internal Server Error",
			"correlation_id": correlationID,
		})
	})
}

// RequestContext provides a middleware that ensures request context is available
func RequestContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract correlation ID from headers or generate new one
		correlationID, tenantID, userID := ExtractFromHeaders(c.Request)

		// Add to context
		ctx := WithCorrelationID(c.Request.Context(), correlationID)
		if tenantID != "" {
			ctx = WithTenantID(ctx, tenantID)
		}
		if userID != "" {
			ctx = WithUserID(ctx, userID)
		}

		// Update request context
		c.Request = c.Request.WithContext(ctx)

		// Set response headers
		c.Header("X-Correlation-ID", correlationID)
		if tenantID != "" {
			c.Header("X-Tenant-ID", tenantID)
		}

		c.Next()
	}
}

// Helper function to get context information from Gin context
func ContextFromGin(c *gin.Context) context.Context {
	return c.Request.Context()
}

// Helper function to log from Gin handlers with proper context
func LogFromGin(c *gin.Context, logger *Logger) *logrus.Entry {
	correlationID := GetCorrelationID(c.Request.Context())
	tenantID := GetTenantID(c.Request.Context())
	userID := GetUserID(c.Request.Context())

	return logger.WithRequestContext(correlationID, tenantID, userID)
}