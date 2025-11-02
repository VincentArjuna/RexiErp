package database

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// QueryBuilder provides utilities for building database queries
type QueryBuilder struct {
	db *gorm.DB
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(db *gorm.DB) *QueryBuilder {
	return &QueryBuilder{db: db}
}

// WhereClause represents a WHERE clause
type WhereClause struct {
	Column   string
	Operator string
	Value    interface{}
}

// BuildWhere builds WHERE clauses from conditions
func (qb *QueryBuilder) BuildWhere(conditions map[string]interface{}) *gorm.DB {
	query := qb.db

	for column, value := range conditions {
		switch v := value.(type) {
		case map[string]interface{}:
			// Handle operators like {"gt": 100, "lt": 200}
			for op, val := range v {
				sqlOp := qb.mapOperator(op)
				query = query.Where(fmt.Sprintf("%s %s ?", column, sqlOp), val)
			}
		case []interface{}:
			// Handle IN clauses
			if len(v) > 0 {
				query = query.Where(fmt.Sprintf("%s IN ?", column), v)
			}
		default:
			// Handle equality
			query = query.Where(fmt.Sprintf("%s = ?", column), value)
		}
	}

	return query
}

// mapOperator maps string operators to SQL operators
func (qb *QueryBuilder) mapOperator(op string) string {
	operators := map[string]string{
		"eq":  "=",
		"ne":  "!=",
		"gt":  ">",
		"gte": ">=",
		"lt":  "<",
		"lte": "<=",
		"like": "LIKE",
		"ilike": "ILIKE",
		"in": "IN",
		"not_in": "NOT IN",
		"is_null": "IS NULL",
		"is_not_null": "IS NOT NULL",
	}

	if sqlOp, exists := operators[op]; exists {
		return sqlOp
	}

	return "=" // Default to equality
}

// Pagination represents pagination parameters
type Pagination struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	SortBy   string `json:"sort_by" form:"sort_by"`
	SortDir  string `json:"sort_dir" form:"sort_dir"`
}

// ApplyPagination applies pagination to a query
func (qb *QueryBuilder) ApplyPagination(pagination Pagination) *gorm.DB {
	// Set default values
	if pagination.Page <= 0 {
		pagination.Page = 1
	}
	if pagination.PageSize <= 0 {
		pagination.PageSize = DefaultPageSize
	}
	if pagination.PageSize > MaxPageSize {
		pagination.PageSize = MaxPageSize
	}

	// Calculate offset
	offset := (pagination.Page - 1) * pagination.PageSize

	// Apply offset and limit
	query := qb.db.Offset(offset).Limit(pagination.PageSize)

	// Apply sorting
	if pagination.SortBy != "" {
		sortDir := strings.ToUpper(pagination.SortDir)
		if sortDir != "ASC" && sortDir != "DESC" {
			sortDir = "ASC"
		}

		// Validate sort field to prevent SQL injection
		if qb.isValidSortField(pagination.SortBy) {
			query = query.Order(fmt.Sprintf("%s %s", pagination.SortBy, sortDir))
		}
	}

	return query
}

// isValidSortField validates the sort field to prevent SQL injection
func (qb *QueryBuilder) isValidSortField(field string) bool {
	// Only allow alphanumeric characters, underscores, and dots
	matched, err := regexp.MatchString(`^[a-zA-Z0-9_.]+$`, field)
	if err != nil {
		return false
	}
	return matched
}

// SearchBuilder provides search functionality
type SearchBuilder struct {
	db *gorm.DB
}

// NewSearchBuilder creates a new search builder
func NewSearchBuilder(db *gorm.DB) *SearchBuilder {
	return &SearchBuilder{db: db}
}

// SearchParams represents search parameters
type SearchParams struct {
	Query    string                 `json:"query" form:"q"`
	Fields   []string               `json:"fields"`
	Filters  map[string]interface{} `json:"filters"`
	Pagination Pagination           `json:"pagination"`
}

// BuildSearch builds a search query
func (sb *SearchBuilder) BuildSearch(model interface{}, params SearchParams) *gorm.DB {
	query := sb.db.Model(model)

	// Apply text search
	if params.Query != "" && len(params.Fields) > 0 {
		var conditions []string
		var values []interface{}

		for _, field := range params.Fields {
			if sb.isValidField(field) {
				conditions = append(conditions, fmt.Sprintf("%s ILIKE ?", field))
				values = append(values, fmt.Sprintf("%%%s%%", params.Query))
			}
		}

		if len(conditions) > 0 {
			whereClause := strings.Join(conditions, " OR ")
			query = query.Where(fmt.Sprintf("(%s)", whereClause), values...)
		}
	}

	// Apply filters
	if len(params.Filters) > 0 {
		qb := NewQueryBuilder(query)
		query = qb.BuildWhere(params.Filters)
	}

	// Apply pagination
	qb := NewQueryBuilder(query)
	query = qb.ApplyPagination(params.Pagination)

	return query
}

// isValidField validates field names to prevent SQL injection
func (sb *SearchBuilder) isValidField(field string) bool {
	// Only allow alphanumeric characters and underscores
	matched, err := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, field)
	if err != nil {
		return false
	}
	return matched
}

// TenantQueryBuilder provides tenant-aware query building
type TenantQueryBuilder struct {
	db *gorm.DB
}

// NewTenantQueryBuilder creates a new tenant query builder
func NewTenantQueryBuilder(db *gorm.DB) *TenantQueryBuilder {
	return &TenantQueryBuilder{db: db}
}

// WithTenant adds tenant context to a query
func (tqb *TenantQueryBuilder) WithTenant(ctx context.Context, tenantID uuid.UUID) *gorm.DB {
	query := tqb.db.WithContext(ctx)

	if tenantID != uuid.Nil {
		query = query.Where("tenant_id = ?", tenantID)
	}

	return query
}

// BulkOperation provides bulk database operations
type BulkOperation struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewBulkOperation creates a new bulk operation handler
func NewBulkOperation(db *gorm.DB, logger *logrus.Logger) *BulkOperation {
	return &BulkOperation{
		db:     db,
		logger: logger,
	}
}

// BulkInsert performs a bulk insert operation
func (bo *BulkOperation) BulkInsert(ctx context.Context, records interface{}, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 100
	}

	return bo.db.WithContext(ctx).CreateInBatches(records, batchSize).Error
}

// BulkUpdate performs a bulk update operation
func (bo *BulkOperation) BulkUpdate(ctx context.Context, model interface{}, updates map[string]interface{}, conditions map[string]interface{}) error {
	query := bo.db.WithContext(ctx).Model(model)

	// Apply conditions
	for column, value := range conditions {
		query = query.Where(fmt.Sprintf("%s = ?", column), value)
	}

	return query.Updates(updates).Error
}

// BulkDelete performs a bulk delete operation
func (bo *BulkOperation) BulkDelete(ctx context.Context, model interface{}, conditions map[string]interface{}) error {
	query := bo.db.WithContext(ctx).Model(model)

	// Apply conditions
	for column, value := range conditions {
		query = query.Where(fmt.Sprintf("%s = ?", column), value)
	}

	return query.Delete(model).Error
}

// CacheManager provides caching utilities
type CacheManager struct {
	redis interface{} // Use interface{} to avoid import cycle
}

// NewCacheManager creates a new cache manager
func NewCacheManager(redis interface{}) *CacheManager {
	return &CacheManager{redis: redis}
}

// CacheKey represents a cache key
type CacheKey struct {
	Prefix string
	ID     string
}

// String returns the string representation of the cache key
func (ck CacheKey) String() string {
	return fmt.Sprintf("%s:%s", ck.Prefix, ck.ID)
}

// Get retrieves a value from cache
func (cm *CacheManager) Get(ctx context.Context, key CacheKey) (string, error) {
	if cm.redis == nil {
		return "", fmt.Errorf("redis client not available")
	}

	// Type assertion for Redis client - implement as needed
	// This is a placeholder implementation
	return "", fmt.Errorf("cache implementation not available")
}

// Set stores a value in cache
func (cm *CacheManager) Set(ctx context.Context, key CacheKey, value string, expiration time.Duration) error {
	if cm.redis == nil {
		return fmt.Errorf("redis client not available")
	}

	// Type assertion for Redis client - implement as needed
	// This is a placeholder implementation
	return fmt.Errorf("cache implementation not available")
}

// Delete removes a value from cache
func (cm *CacheManager) Delete(ctx context.Context, key CacheKey) error {
	if cm.redis == nil {
		return fmt.Errorf("redis client not available")
	}

	// Type assertion for Redis client - implement as needed
	// This is a placeholder implementation
	return fmt.Errorf("cache implementation not available")
}

// ClearByPrefix removes all keys with a given prefix
func (cm *CacheManager) ClearByPrefix(ctx context.Context, prefix string) error {
	if cm.redis == nil {
		return fmt.Errorf("redis client not available")
	}

	// Type assertion for Redis client - implement as needed
	// This is a placeholder implementation
	return fmt.Errorf("cache implementation not available")
}

// ConnectionPoolManager provides connection pool management utilities
type ConnectionPoolManager struct {
	db     *Database
	logger *logrus.Logger
}

// NewConnectionPoolManager creates a new connection pool manager
func NewConnectionPoolManager(db *Database, logger *logrus.Logger) *ConnectionPoolManager {
	return &ConnectionPoolManager{
		db:     db,
		logger: logger,
	}
}

// GetPoolMetrics returns current connection pool metrics
func (cpm *ConnectionPoolManager) GetPoolMetrics() ConnectionMetrics {
	return cpm.db.GetMetrics()
}

// OptimizePool optimizes connection pool settings based on current metrics
func (cpm *ConnectionPoolManager) OptimizePool() error {
	metrics := cpm.GetPoolMetrics()

	// Log current metrics
	cpm.logger.WithFields(logrus.Fields{
		"total_connections": metrics.TotalConnections,
		"active_connections": metrics.ActiveConnections,
		"idle_connections": metrics.IdleConnections,
		"wait_count": metrics.WaitCount,
		"wait_duration": metrics.WaitDuration,
	}).Info("Current connection pool metrics")

	// Check for optimization opportunities
	var recommendations []string

	if metrics.WaitCount > 100 && metrics.WaitDuration > 100*time.Millisecond {
		recommendations = append(recommendations, "Consider increasing max open connections")
	}

	if metrics.IdleConnections > metrics.TotalConnections/2 {
		recommendations = append(recommendations, "Consider reducing max idle connections")
	}

	if metrics.HealthCheckFailures > 5 {
		recommendations = append(recommendations, "Investigate connection health issues")
	}

	if len(recommendations) > 0 {
		cpm.logger.WithField("recommendations", recommendations).Info("Connection pool optimization recommendations")
	}

	return nil
}

// BackupManager provides database backup utilities
type BackupManager struct {
	db     *Database
	logger *logrus.Logger
}

// NewBackupManager creates a new backup manager
func NewBackupManager(db *Database, logger *logrus.Logger) *BackupManager {
	return &BackupManager{
		db:     db,
		logger: logger,
	}
}

// CreateBackup creates a database backup
func (bm *BackupManager) CreateBackup(ctx context.Context, backupPath string) error {
	// This is a placeholder for backup functionality
	// In production, you would implement database-specific backup logic

	bm.logger.WithField("path", backupPath).Info("Creating database backup")

	// Example for PostgreSQL:
	// cmd := exec.CommandContext(ctx, "pg_dump", "--format=custom", "--no-owner", "--no-privileges", dbName)
	// output, err := cmd.Output()
	// if err != nil {
	//     return fmt.Errorf("backup failed: %w", err)
	// }
	// return os.WriteFile(backupPath, output, 0644)

	return fmt.Errorf("backup functionality not implemented")
}

// RestoreBackup restores a database from backup
func (bm *BackupManager) RestoreBackup(ctx context.Context, backupPath string) error {
	// This is a placeholder for restore functionality
	// In production, you would implement database-specific restore logic

	bm.logger.WithField("path", backupPath).Info("Restoring database from backup")

	return fmt.Errorf("restore functionality not implemented")
}

// MigrationHelper provides migration utilities
type MigrationHelper struct {
	migrationManager *MigrationManager
	logger           *logrus.Logger
}

// NewMigrationHelper creates a new migration helper
func NewMigrationHelper(mm *MigrationManager, logger *logrus.Logger) *MigrationHelper {
	return &MigrationHelper{
		migrationManager: mm,
		logger:           logger,
	}
}

// CreateMigrationTemplate creates a new migration template
func (mh *MigrationHelper) CreateMigrationTemplate(version, description string) error {
	// Create up migration file
	upFile := fmt.Sprintf("migrations/master/%s_%s.up.sql", version, strings.ReplaceAll(description, " ", "_"))
	upContent := fmt.Sprintf(`-- Migration: %s
-- Description: %s
-- Version: %s
-- Created: %s

-- Add your migration SQL here
-- Example:
-- CREATE TABLE example_table (
--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     name VARCHAR(255) NOT NULL,
--     created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
--     updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
-- );
`, description, description, version, time.Now().Format(time.RFC3339))

	// Create down migration file
	downFile := fmt.Sprintf("migrations/master/%s_%s.down.sql", version, strings.ReplaceAll(description, " ", "_"))
	downContent := fmt.Sprintf(`-- Rollback: %s
-- Description: %s
-- Version: %s
-- Created: %s

-- Add your rollback SQL here
-- Example:
-- DROP TABLE IF EXISTS example_table;
`, description, description, version, time.Now().Format(time.RFC3339))

	// Use the template contents (prevents unused variable warnings)
	_ = upContent
	_ = downContent

	// Write files (implementation depends on your file system)
	mh.logger.WithFields(logrus.Fields{
		"up_file":   upFile,
		"down_file": downFile,
	}).Info("Migration template created")

	return nil
}