package database

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// GORMConfig holds advanced GORM configuration
type GORMConfig struct {
	Logger             *logrus.Logger
	LogLevel           gormlogger.LogLevel
	SlowThreshold      time.Duration
	IgnoreRecordNotFound bool
	Colorful          bool
	PrepareStmt       bool
	AllowGlobalUpdate bool
}

// GORMManager provides advanced GORM integration with hooks and utilities
type GORMManager struct {
	db     *gorm.DB
	logger *logrus.Logger
	config *GORMConfig
}

// NewGORMManager creates a new GORM manager with enhanced configuration
func NewGORMManager(db *gorm.DB, cfg *GORMConfig) *GORMManager {
	if cfg == nil {
		cfg = &GORMConfig{
			LogLevel:              gormlogger.Silent,
			SlowThreshold:         200 * time.Millisecond,
			IgnoreRecordNotFound:  true,
			Colorful:              false,
			PrepareStmt:           true,
			AllowGlobalUpdate:     false,
		}
	}

	// Configure GORM callbacks and hooks
	setupGORMCallbacks(db, cfg)

	return &GORMManager{
		db:     db,
		logger: cfg.Logger,
		config: cfg,
	}
}

// setupGORMCallbacks configures GORM callbacks for enhanced functionality
func setupGORMCallbacks(db *gorm.DB, cfg *GORMConfig) {
	// Register callbacks for audit trail
	registerAuditCallbacks(db)

	// Register callbacks for tenant isolation
	registerTenantCallbacks(db)

	// Register callbacks for performance monitoring
	registerPerformanceCallbacks(db, cfg.Logger)
}

// registerAuditCallbacks registers callbacks for audit trail functionality
func registerAuditCallbacks(db *gorm.DB) {
	// Before create callback
	db.Callback().Create().Before("gorm:create").Register("audit:before_create", beforeCreateAudit)

	// Before update callback
	db.Callback().Update().Before("gorm:update").Register("audit:before_update", beforeUpdateAudit)

	// Before delete callback
	db.Callback().Delete().Before("gorm:delete").Register("audit:before_delete", beforeDeleteAudit)

	// After create/update/delete callbacks for logging
	db.Callback().Create().After("gorm:create").Register("audit:after_create", afterCreateAudit)
	db.Callback().Update().After("gorm:update").Register("audit:after_update", afterUpdateAudit)
	db.Callback().Delete().After("gorm:delete").Register("audit:after_delete", afterDeleteAudit)
}

// registerTenantCallbacks registers callbacks for multi-tenant isolation
func registerTenantCallbacks(db *gorm.DB) {
	// Before query callback to enforce tenant isolation
	db.Callback().Query().Before("gorm:query").Register("tenant:before_query", beforeQueryTenant)
	db.Callback().Create().Before("gorm:create").Register("tenant:before_create", beforeCreateTenant)
	db.Callback().Update().Before("gorm:update").Register("tenant:before_update", beforeUpdateTenant)
	db.Callback().Delete().Before("gorm:delete").Register("tenant:before_delete", beforeDeleteTenant)
}

// registerPerformanceCallbacks registers callbacks for performance monitoring
func registerPerformanceCallbacks(db *gorm.DB, logger *logrus.Logger) {
	// Before callbacks for timing
	db.Callback().Create().Before("gorm:create").Register("perf:before_create", startPerformanceTimer("create"))
	db.Callback().Update().Before("gorm:update").Register("perf:before_update", startPerformanceTimer("update"))
	db.Callback().Delete().Before("gorm:delete").Register("perf:before_delete", startPerformanceTimer("delete"))
	db.Callback().Query().Before("gorm:query").Register("perf:before_query", startPerformanceTimer("query"))

	// After callbacks for timing
	db.Callback().Create().After("gorm:create").Register("perf:after_create", endPerformanceTimer("create", logger))
	db.Callback().Update().After("gorm:update").Register("perf:after_update", endPerformanceTimer("update", logger))
	db.Callback().Delete().After("gorm:delete").Register("perf:after_delete", endPerformanceTimer("delete", logger))
	db.Callback().Query().After("gorm:query").Register("perf:after_query", endPerformanceTimer("query", logger))
}

// TransactionManager provides enhanced transaction management
type TransactionManager struct {
	db *gorm.DB
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(db *gorm.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

// WithTransaction executes a function within a database transaction
func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	tx := tm.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // Re-throw panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback().Error; rbErr != nil {
			return fmt.Errorf("transaction failed: %w, rollback failed: %v", err, rbErr)
		}
		return fmt.Errorf("transaction failed and rolled back: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// WithRetryTransaction executes a function with transaction and retry logic
func (tm *TransactionManager) WithRetryTransaction(ctx context.Context, maxRetries int, fn func(*gorm.DB) error) error {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := tm.WithTransaction(ctx, func(tx *gorm.DB) error {
			return fn(tx)
		})

		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable (deadlock, connection timeout, etc.)
		if !isRetryableError(err) {
			return err
		}

		if attempt < maxRetries {
			// Exponential backoff
			delay := time.Duration(attempt) * 100 * time.Millisecond
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("transaction failed after %d attempts: %w", maxRetries, lastErr)
}

// Repository provides base repository pattern with GORM
type Repository[T any] struct {
	db *gorm.DB
}

// NewRepository creates a new generic repository
func NewRepository[T any](db *gorm.DB) *Repository[T] {
	return &Repository[T]{db: db}
}

// Create creates a new record
func (r *Repository[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// CreateBatch creates multiple records
func (r *Repository[T]) CreateBatch(ctx context.Context, entities []*T) error {
	return r.db.WithContext(ctx).CreateInBatches(entities, 100).Error
}

// FindByID finds a record by ID
func (r *Repository[T]) FindByID(ctx context.Context, id interface{}) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// FindAll finds all records with optional filtering
func (r *Repository[T]) FindAll(ctx context.Context, conditions map[string]interface{}) ([]*T, error) {
	var entities []*T
	query := r.db.WithContext(ctx)

	for key, value := range conditions {
		query = query.Where(key, value)
	}

	err := query.Find(&entities).Error
	return entities, err
}

// Update updates a record
func (r *Repository[T]) Update(ctx context.Context, id interface{}, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Updates(updates).Error
}

// Delete soft deletes a record
func (r *Repository[T]) Delete(ctx context.Context, id interface{}) error {
	return r.db.WithContext(ctx).Delete(new(T), id).Error
}

// HardDelete permanently deletes a record
func (r *Repository[T]) HardDelete(ctx context.Context, id interface{}) error {
	return r.db.WithContext(ctx).Unscoped().Delete(new(T), id).Error
}

// Count returns the count of records
func (r *Repository[T]) Count(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(new(T))

	for key, value := range conditions {
		query = query.Where(key, value)
	}

	err := query.Count(&count).Error
	return count, err
}

// Exists checks if a record exists
func (r *Repository[T]) Exists(ctx context.Context, conditions map[string]interface{}) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(new(T))

	for key, value := range conditions {
		query = query.Where(key, value)
	}

	err := query.Count(&count).Error
	return count > 0, err
}

// GetDB returns the underlying GORM database instance
func (r *Repository[T]) GetDB() *gorm.DB {
	return r.db
}

// GetTenantDB returns a database instance scoped to a tenant
func (r *Repository[T]) GetTenantDB(ctx context.Context, tenantID string) *gorm.DB {
	return r.db.WithContext(context.WithValue(ctx, "tenant_id", tenantID))
}