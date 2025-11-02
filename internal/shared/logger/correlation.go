package logger

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// correlationIDKey is the context key for correlation ID
type correlationIDKey struct{}

// tenantIDKey is the context key for tenant ID
type tenantIDKey struct{}

// userIDKey is the context key for user ID
type userIDKey struct{}

// WithCorrelationID adds a correlation ID to the context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey{}, correlationID)
}

// WithTenantID adds a tenant ID to the context
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantIDKey{}, tenantID)
}

// WithUserID adds a user ID to the context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

// GetCorrelationID retrieves the correlation ID from the context
func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(correlationIDKey{}).(string); ok {
		return id
	}
	return ""
}

// GetTenantID retrieves the tenant ID from the context
func GetTenantID(ctx context.Context) string {
	if id, ok := ctx.Value(tenantIDKey{}).(string); ok {
		return id
	}
	return ""
}

// GetUserID retrieves the user ID from the context
func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(userIDKey{}).(string); ok {
		return id
	}
	return ""
}

// GenerateCorrelationID creates a new correlation ID
func GenerateCorrelationID() string {
	return uuid.New().String()
}

// ExtractFromHeaders extracts correlation, tenant, and user IDs from HTTP headers
func ExtractFromHeaders(r *http.Request) (correlationID, tenantID, userID string) {
	// Get correlation ID from header or generate a new one
	correlationID = r.Header.Get("X-Correlation-ID")
	if correlationID == "" {
		correlationID = GenerateCorrelationID()
	}

	// Get tenant ID from header
	tenantID = r.Header.Get("X-Tenant-ID")

	// Get user ID from header
	userID = r.Header.Get("X-User-ID")

	return correlationID, tenantID, userID
}

// SetInHeaders sets correlation, tenant, and user IDs in HTTP headers
func SetInHeaders(header http.Header, correlationID, tenantID, userID string) {
	if correlationID != "" {
		header.Set("X-Correlation-ID", correlationID)
	}
	if tenantID != "" {
		header.Set("X-Tenant-ID", tenantID)
	}
	if userID != "" {
		header.Set("X-User-ID", userID)
	}
}

// ContextWithRequestHeaders extracts context information from HTTP headers and adds it to the context
func ContextWithRequestHeaders(ctx context.Context, r *http.Request) context.Context {
	correlationID, tenantID, userID := ExtractFromHeaders(r)

	ctx = WithCorrelationID(ctx, correlationID)
	if tenantID != "" {
		ctx = WithTenantID(ctx, tenantID)
	}
	if userID != "" {
		ctx = WithUserID(ctx, userID)
	}

	return ctx
}

// RequestLoggerContext creates a context with request-specific logging information
func RequestLoggerContext(ctx context.Context, correlationID, tenantID, userID string) context.Context {
	if correlationID == "" {
		correlationID = GenerateCorrelationID()
	}

	ctx = WithCorrelationID(ctx, correlationID)
	if tenantID != "" {
		ctx = WithTenantID(ctx, tenantID)
	}
	if userID != "" {
		ctx = WithUserID(ctx, userID)
	}

	return ctx
}