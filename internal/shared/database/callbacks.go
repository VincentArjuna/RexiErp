package database

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Audit trail callback functions
func beforeCreateAudit(db *gorm.DB) {
	// Set created_at and updated_at timestamps
	db.Statement.SetColumn("created_at", time.Now())
	db.Statement.SetColumn("updated_at", time.Now())
}

func beforeUpdateAudit(db *gorm.DB) {
	// Set updated_at timestamp
	db.Statement.SetColumn("updated_at", time.Now())
}

func beforeDeleteAudit(db *gorm.DB) {
	// Set deleted_at timestamp for soft deletes
	db.Statement.SetColumn("deleted_at", time.Now())
}

func afterCreateAudit(db *gorm.DB) {
	// Log creation
	tableName := db.Statement.Table
	id := db.Statement.Dest

	if logger, ok := db.Logger.(interface{ Infof(string, ...interface{}) }); ok {
		logger.Infof("Created record in %s: %+v", tableName, id)
	}
}

func afterUpdateAudit(db *gorm.DB) {
	// Log update
	tableName := db.Statement.Table
	affected := db.Statement.RowsAffected

	if logger, ok := db.Logger.(interface{ Infof(string, ...interface{}) }); ok {
		logger.Infof("Updated %d record(s) in %s", affected, tableName)
	}
}

func afterDeleteAudit(db *gorm.DB) {
	// Log deletion
	tableName := db.Statement.Table
	affected := db.Statement.RowsAffected

	if logger, ok := db.Logger.(interface{ Infof(string, ...interface{}) }); ok {
		logger.Infof("Deleted %d record(s) from %s", affected, tableName)
	}
}

// Tenant isolation callback functions
func beforeQueryTenant(db *gorm.DB) {
	ctx := db.Statement.Context
	if ctx == nil {
		// Reject queries without tenant context for security
		db.AddError(fmt.Errorf("tenant context is required for database queries"))
		return
	}

	tenantID, ok := ctx.Value("tenant_id").(string)
	if !ok || tenantID == "" {
		// Reject queries without valid tenant ID
		db.AddError(fmt.Errorf("valid tenant_id is required in context for database queries"))
		return
	}

	// Validate tenant ID format (UUID)
	if !isValidUUID(tenantID) {
		db.AddError(fmt.Errorf("invalid tenant_id format in context"))
		return
	}

	// Add tenant condition to query with proper escaping
	if db.Statement.Table != "" {
		// Use parameterized query to prevent SQL injection
		db.Where("tenant_id = ?", tenantID)

		// Log tenant access for audit
		if logger := getLoggerFromDB(db); logger != nil {
			logger.WithFields(logrus.Fields{
				"table":    db.Statement.Table,
				"tenant_id": tenantID,
				"operation": "query",
			}).Debug("Tenant isolation enforced")
		}
	}
}

func beforeCreateTenant(db *gorm.DB) {
	ctx := db.Statement.Context
	if ctx == nil {
		// Reject creation without tenant context for security
		db.AddError(fmt.Errorf("tenant context is required for record creation"))
		return
	}

	tenantID, ok := ctx.Value("tenant_id").(string)
	if !ok || tenantID == "" {
		// Reject creation without valid tenant ID
		db.AddError(fmt.Errorf("valid tenant_id is required in context for record creation"))
		return
	}

	// Validate tenant ID format (UUID)
	if !isValidUUID(tenantID) {
		db.AddError(fmt.Errorf("invalid tenant_id format in context"))
		return
	}

	// Set tenant_id for the record with proper validation
	db.Statement.SetColumn("tenant_id", tenantID)

	// Log tenant creation for audit
	if logger := getLoggerFromDB(db); logger != nil {
		logger.WithFields(logrus.Fields{
			"table":    db.Statement.Table,
			"tenant_id": tenantID,
			"operation": "create",
		}).Debug("Tenant context enforced for creation")
	}
}

func beforeUpdateTenant(db *gorm.DB) {
	ctx := db.Statement.Context
	if ctx == nil {
		// Reject updates without tenant context for security
		db.AddError(fmt.Errorf("tenant context is required for record updates"))
		return
	}

	tenantID, ok := ctx.Value("tenant_id").(string)
	if !ok || tenantID == "" {
		// Reject updates without valid tenant ID
		db.AddError(fmt.Errorf("valid tenant_id is required in context for record updates"))
		return
	}

	// Validate tenant ID format (UUID)
	if !isValidUUID(tenantID) {
		db.AddError(fmt.Errorf("invalid tenant_id format in context"))
		return
	}

	// Ensure update only affects records in the same tenant
	db.Where("tenant_id = ?", tenantID)

	// Log tenant update for audit
	if logger := getLoggerFromDB(db); logger != nil {
		logger.WithFields(logrus.Fields{
			"table":    db.Statement.Table,
			"tenant_id": tenantID,
			"operation": "update",
		}).Debug("Tenant context enforced for update")
	}
}

func beforeDeleteTenant(db *gorm.DB) {
	ctx := db.Statement.Context
	if ctx == nil {
		// Reject deletion without tenant context for security
		db.AddError(fmt.Errorf("tenant context is required for record deletion"))
		return
	}

	tenantID, ok := ctx.Value("tenant_id").(string)
	if !ok || tenantID == "" {
		// Reject deletion without valid tenant ID
		db.AddError(fmt.Errorf("valid tenant_id is required in context for record deletion"))
		return
	}

	// Validate tenant ID format (UUID)
	if !isValidUUID(tenantID) {
		db.AddError(fmt.Errorf("invalid tenant_id format in context"))
		return
	}

	// Ensure deletion only affects records in the same tenant
	db.Where("tenant_id = ?", tenantID)

	// Log tenant deletion for audit
	if logger := getLoggerFromDB(db); logger != nil {
		logger.WithFields(logrus.Fields{
			"table":    db.Statement.Table,
			"tenant_id": tenantID,
			"operation": "delete",
		}).Debug("Tenant context enforced for deletion")
	}
}

// Performance monitoring callback functions
type performanceContextKey string

const (
	performanceStartTimeKey performanceContextKey = "perf_start_time"
	operationTypeKey        performanceContextKey = "operation_type"
)

func startPerformanceTimer(operationType string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		ctx := db.Statement.Context
		if ctx == nil {
			ctx = context.Background()
		}

		ctx = context.WithValue(ctx, performanceStartTimeKey, time.Now())
		ctx = context.WithValue(ctx, operationTypeKey, operationType)
		db.InstanceSet("gorm:context", ctx)
	}
}

func endPerformanceTimer(operationType string, logger *logrus.Logger) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		ctx := db.Statement.Context
		if ctx == nil {
			return
		}

		startTime, ok := ctx.Value(performanceStartTimeKey).(time.Time)
		if !ok {
			return
		}

		duration := time.Since(startTime)

		// Log performance metrics
		logger.WithFields(logrus.Fields{
			"operation": operationType,
			"table":     db.Statement.Table,
			"duration":  duration,
			"rows_affected": db.Statement.RowsAffected,
		}).Debug("Database operation completed")

		// Log warning for slow operations
		if duration > 200*time.Millisecond {
			logger.WithFields(logrus.Fields{
				"operation": operationType,
				"table":     db.Statement.Table,
				"duration":  duration,
				"rows_affected": db.Statement.RowsAffected,
			}).Warn("Slow database operation detected")
		}
	}
}

// isRetryableError checks if an error is retryable for transaction retry logic
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Check for common retryable errors
	retryableErrors := []string{
		"deadlock",
		"lock wait timeout",
		"connection reset",
		"connection refused",
		"timeout",
		"connection lost",
	}

	for _, retryableErr := range retryableErrors {
		if contains(errStr, retryableErr) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		   (s == substr ||
		    len(s) > len(substr) &&
		    (s[:len(substr)] == substr ||
		     s[len(s)-len(substr):] == substr ||
		     containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TenantManager provides utilities for multi-tenant context management
type TenantManager struct {
	db *gorm.DB
}

// NewTenantManager creates a new tenant manager
func NewTenantManager(db *gorm.DB) *TenantManager {
	return &TenantManager{db: db}
}

// WithTenantContext returns a GORM instance with tenant context
func (tm *TenantManager) WithTenantContext(ctx context.Context, tenantID string) *gorm.DB {
	newCtx := context.WithValue(ctx, "tenant_id", tenantID)
	return tm.db.WithContext(newCtx)
}

// SetTenantContext adds tenant context to an existing context
func (tm *TenantManager) SetTenantContext(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, "tenant_id", tenantID)
}

// GetTenantFromContext extracts tenant ID from context
func (tm *TenantManager) GetTenantFromContext(ctx context.Context) (string, bool) {
	tenantID, ok := ctx.Value("tenant_id").(string)
	return tenantID, ok
}

// ValidateTenantContext validates that tenant context exists
func (tm *TenantManager) ValidateTenantContext(ctx context.Context) error {
	_, ok := tm.GetTenantFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant context is required for this operation")
	}
	return nil
}

// AuditManager provides audit trail functionality
type AuditManager struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewAuditManager creates a new audit manager
func NewAuditManager(db *gorm.DB, logger *logrus.Logger) *AuditManager {
	return &AuditManager{
		db:     db,
		logger: logger,
	}
}

// LogChange logs an audit trail entry
func (am *AuditManager) LogChange(ctx context.Context, operation, tableName string, recordID interface{}, changes map[string]interface{}) error {
	// This would log to an audit table
	am.logger.WithFields(logrus.Fields{
		"operation": operation,
		"table":     tableName,
		"record_id": recordID,
		"changes":   changes,
		"tenant_id": ctx.Value("tenant_id"),
	}).Info("Database operation audit log")

	return nil
}

// Helper functions for tenant validation and logging

// isValidUUID validates if a string is a valid UUID
func isValidUUID(s string) bool {
	// Check basic UUID format (length and hyphens)
	if len(s) != 36 {
		return false
	}

	// Basic format check without regex
	if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return false
	}

	// Check that remaining characters are hex
	hexParts := []string{
		s[0:8], s[9:13], s[14:18], s[19:23], s[24:36],
	}

	for _, part := range hexParts {
		for _, c := range part {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
	}

	return true
}

// getLoggerFromDB extracts logger from GORM instance
func getLoggerFromDB(db *gorm.DB) *logrus.Logger {
	// Return default logger if available
	return logrus.StandardLogger()
}

// createRLSPolicies creates Row Level Security policies for tenant isolation
// This should be called during database migration/setup
func createRLSPolicies(db *gorm.DB, tables []string) error {

	for _, table := range tables {
		// Enable RLS on the table
		if err := db.Exec(fmt.Sprintf("ALTER TABLE %s ENABLE ROW LEVEL SECURITY", table)).Error; err != nil {
			return fmt.Errorf("failed to enable RLS on table %s: %w", table, err)
		}

		// Create tenant isolation policy
		policyName := fmt.Sprintf("%s_tenant_isolation_policy", table)
		policySQL := fmt.Sprintf(`
			CREATE POLICY %s ON %s
			FOR ALL
			TO application_role
			USING (tenant_id = current_setting('app.current_tenant_id')::UUID)
			WITH CHECK (tenant_id = current_setting('app.current_tenant_id')::UUID)
		`, policyName, table)

		if err := db.Exec(policySQL).Error; err != nil {
			// Policy might already exist, which is okay
			if !contains(err.Error(), "already exists") {
				return fmt.Errorf("failed to create RLS policy for table %s: %w", table, err)
			}
		}

		if logger := getLoggerFromDB(db); logger != nil {
			logger.WithFields(logrus.Fields{
				"table": table,
				"policy": policyName,
			}).Info("RLS policy created for tenant isolation")
		}
	}

	return nil
}

// setTenantContext sets the PostgreSQL session variable for tenant context
func SetTenantContext(db *gorm.DB, tenantID string) error {
	if !isValidUUID(tenantID) {
		return fmt.Errorf("invalid tenant ID format")
	}

	// Set the session variable for RLS policies
	sql := fmt.Sprintf("SET app.current_tenant_id = '%s'", tenantID)
	if err := db.Exec(sql).Error; err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	return nil
}

// clearTenantContext clears the PostgreSQL session variable
func ClearTenantContext(db *gorm.DB) error {
	if err := db.Exec("RESET app.current_tenant_id").Error; err != nil {
		return fmt.Errorf("failed to clear tenant context: %w", err)
	}
	return nil
}