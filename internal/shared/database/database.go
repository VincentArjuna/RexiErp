package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/VincentArjuna/RexiErp/internal/shared/config"
)

// Database wrapper for GORM with connection pooling
type Database struct {
	DB     *gorm.DB
	SQLDB  *sql.DB
	Logger *logrus.Logger
	Config *config.DatabaseConfig
}

// NewDatabase creates a new database connection with optimized pooling
func NewDatabase(cfg *config.DatabaseConfig, logger *logrus.Logger) (*Database, error) {
	dsn := cfg.GetDSN()

	// Configure GORM logger
	gormLogger := logger.New()
	logLevel := logger.Silent
	if logger.Level == logrus.DebugLevel {
		logLevel = logger.Info
	}
	gormLogger.SetLevel(logLevel)

	// Open GORM connection with PostgreSQL driver
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logLevel,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	if err := configureConnectionPool(sqlDB, cfg, logger); err != nil {
		return nil, fmt.Errorf("failed to configure connection pool: %w", err)
	}

	// Test the connection
	if err := testConnection(sqlDB, logger); err != nil {
		return nil, fmt.Errorf("database connection test failed: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"host":            cfg.Host,
		"port":            cfg.Port,
		"database":        cfg.Name,
		"max_open_conns":  cfg.MaxOpenConns,
		"max_idle_conns":  cfg.MaxIdleConns,
		"conn_max_lifetime": cfg.ConnMaxLifetime,
		"conn_max_idle_time": cfg.ConnMaxIdleTime,
	}).Info("Database connection established with optimized connection pool")

	return &Database{
		DB:     db,
		SQLDB:  sqlDB,
		Logger: logger,
		Config: cfg,
	}, nil
}

// configureConnectionPool sets up the database connection pool with optimal settings
func configureConnectionPool(sqlDB *sql.DB, cfg *config.DatabaseConfig, logger *logrus.Logger) error {
	// Set maximum number of open connections
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	// Set maximum number of idle connections
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	// Set maximum lifetime of a connection
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Set maximum idle time for a connection
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	logger.WithFields(logrus.Fields{
		"max_open_conns":     cfg.MaxOpenConns,
		"max_idle_conns":     cfg.MaxIdleConns,
		"conn_max_lifetime":  cfg.ConnMaxLifetime,
		"conn_max_idle_time": cfg.ConnMaxIdleTime,
	}).Info("Database connection pool configured")

	return nil
}

// testConnection verifies the database connection is working
func testConnection(sqlDB *sql.DB, logger *logrus.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection test successful")
	return nil
}

// GetStats returns current connection pool statistics
func (d *Database) GetStats() sql.DBStats {
	return d.SQLDB.Stats()
}

// LogStats logs current connection pool statistics
func (d *Database) LogStats() {
	stats := d.GetStats()
	d.Logger.WithFields(logrus.Fields{
		"open_connections":     stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration,
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}).Debug("Database connection pool statistics")
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.SQLDB != nil {
		d.Logger.Info("Closing database connection")
		return d.SQLDB.Close()
	}
	return nil
}

// HealthCheck performs a health check on the database
func (d *Database) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := d.SQLDB.PingContext(ctx); err != nil {
		d.Logger.WithError(err).Error("Database health check failed")
		return fmt.Errorf("database health check failed: %w", err)
	}

	d.LogStats()
	return nil
}

// BeginTx starts a new transaction with the given options
func (d *Database) BeginTx(opts *sql.TxOptions) (*gorm.DB, error) {
	return d.DB.Begin(opts)
}

// GetTenantDB returns a database connection scoped to the specified tenant
func (d *Database) GetTenantDB(tenantID string) *gorm.DB {
	// Implement tenant isolation by setting search_path or schema
	// This is a placeholder for tenant-specific database configuration
	return d.DB.WithContext(context.WithValue(context.Background(), "tenant_id", tenantID))
}