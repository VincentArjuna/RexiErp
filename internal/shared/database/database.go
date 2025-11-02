package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/VincentArjuna/RexiErp/internal/shared/config"
)

// ConnectionMetrics holds connection pool metrics
type ConnectionMetrics struct {
	TotalConnections     int32
	ActiveConnections    int32
	IdleConnections      int32
	WaitCount            int64
	WaitDuration         time.Duration
	MaxIdleClosed        int64
	MaxLifetimeClosed    int64
	LastHealthCheck      time.Time
	HealthCheckFailures  int32
	ReconnectionAttempts int32
}

// Database wrapper for GORM with enhanced connection pooling and health monitoring
type Database struct {
	DB                *gorm.DB
	SQLDB             *sql.DB
	ReadReplicas      []*sql.DB
	Logger            *logrus.Logger
	Config            *config.DatabaseConfig
	metrics           ConnectionMetrics
	mu                sync.RWMutex
	healthCheckTicker *time.Ticker
	reconnecting      int32
	closed            int32
}

// NewDatabase creates a new database connection with optimized pooling
func NewDatabase(cfg *config.DatabaseConfig, logger *logrus.Logger) (*Database, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.SSLMode,
	)

	// Configure GORM logger
	var logLevel gormlogger.LogLevel
	if logger.Level == logrus.DebugLevel {
		logLevel = gormlogger.Info
	} else {
		logLevel = gormlogger.Silent
	}

	// Open GORM connection with PostgreSQL driver
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			gormlogger.Config{
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

	// Initialize database instance with enhanced features
	database := &Database{
		DB:           db,
		SQLDB:        sqlDB,
		ReadReplicas: []*sql.DB{},
		Logger:       logger,
		Config:       cfg,
		metrics: ConnectionMetrics{
			LastHealthCheck: time.Now(),
		},
	}

	// Start health monitoring
	if err := database.startHealthMonitoring(); err != nil {
		logger.WithError(err).Warn("Failed to start health monitoring")
	}

	return database, nil
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
	return d.DB.Begin(opts), nil
}

// GetTenantDB returns a database connection scoped to the specified tenant
func (d *Database) GetTenantDB(tenantID string) (*gorm.DB, error) {
	// Implement tenant isolation by setting search_path or schema
	// This is a placeholder for tenant-specific database configuration
	return d.DB.WithContext(context.WithValue(context.Background(), "tenant_id", tenantID)), nil
}

// startHealthMonitoring begins periodic health checks
func (d *Database) startHealthMonitoring() error {
	if atomic.LoadInt32(&d.closed) == 1 {
		return fmt.Errorf("database is closed")
	}

	d.healthCheckTicker = time.NewTicker(30 * time.Second)

	go func() {
		for range d.healthCheckTicker.C {
			if atomic.LoadInt32(&d.closed) == 1 {
				return
			}

			if err := d.performHealthCheck(); err != nil {
				d.Logger.WithError(err).Error("Health check failed")

				// Attempt reconnection if health check fails
				if atomic.CompareAndSwapInt32(&d.reconnecting, 0, 1) {
					go d.attemptReconnection()
				}
			}
		}
	}()

	return nil
}

// performHealthCheck conducts a comprehensive health check
func (d *Database) performHealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Update metrics
	d.updateMetrics()

	// Perform ping test
	if err := d.SQLDB.PingContext(ctx); err != nil {
		atomic.AddInt32(&d.metrics.HealthCheckFailures, 1)
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Check connection pool performance
	stats := d.SQLDB.Stats()
	if stats.WaitCount > 0 && stats.WaitDuration > 10*time.Millisecond {
		d.Logger.WithFields(logrus.Fields{
			"wait_count":    stats.WaitCount,
			"wait_duration": stats.WaitDuration,
		}).Warn("Connection pool experiencing delays")
	}

	// Check read replicas if configured
	for i, replica := range d.ReadReplicas {
		if err := replica.PingContext(ctx); err != nil {
			d.Logger.WithFields(logrus.Fields{
				"replica_id": i,
				"error":      err,
			}).Error("Read replica health check failed")
		}
	}

	atomic.StoreInt32(&d.metrics.HealthCheckFailures, 0)
	atomic.StoreInt64(&d.metrics.WaitCount, int64(stats.WaitCount))
	d.metrics.WaitDuration = stats.WaitDuration
	atomic.StoreInt64(&d.metrics.MaxIdleClosed, int64(stats.MaxIdleClosed))
	atomic.StoreInt64(&d.metrics.MaxLifetimeClosed, int64(stats.MaxLifetimeClosed))

	return nil
}

// updateMetrics updates connection pool metrics
func (d *Database) updateMetrics() {
	stats := d.SQLDB.Stats()

	atomic.StoreInt32(&d.metrics.TotalConnections, int32(stats.OpenConnections))
	atomic.StoreInt32(&d.metrics.ActiveConnections, int32(stats.InUse))
	atomic.StoreInt32(&d.metrics.IdleConnections, int32(stats.Idle))
	d.metrics.LastHealthCheck = time.Now()
}

// attemptReconnection implements exponential backoff reconnection logic
func (d *Database) attemptReconnection() {
	defer atomic.StoreInt32(&d.reconnecting, 0)

	maxRetries := 10
	baseDelay := 1 * time.Second
	maxDelay := 30 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Calculate exponential backoff with jitter
		delay := time.Duration(math.Min(
			float64(baseDelay)*math.Pow(2, float64(attempt-1)),
			float64(maxDelay),
		))

		// Add jitter to prevent thundering herd
		jitter := time.Duration(rand.Float64() * float64(delay) * 0.1)
		delay += jitter

		d.Logger.WithFields(logrus.Fields{
			"attempt": attempt,
			"delay":   delay,
			"max_retries": maxRetries,
		}).Info("Attempting database reconnection")

		time.Sleep(delay)

		// Test connection
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := d.SQLDB.PingContext(ctx); err == nil {
			d.Logger.WithField("attempt", attempt).Info("Database reconnection successful")
			atomic.AddInt32(&d.metrics.ReconnectionAttempts, 1)
			cancel()
			return
		}
		cancel()
	}

	d.Logger.Error("Failed to reconnect to database after maximum retries")
}

// GetMetrics returns current connection pool metrics
func (d *Database) GetMetrics() ConnectionMetrics {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Update metrics before returning
	d.updateMetrics()

	return d.metrics
}

// AddReadReplica adds a read replica to the database configuration
func (d *Database) AddReadReplica(cfg *config.DatabaseConfig) error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to read replica: %w", err)
	}

	// Configure connection pool for replica
	if err := configureConnectionPool(db, cfg, d.Logger); err != nil {
		db.Close()
		return fmt.Errorf("failed to configure replica connection pool: %w", err)
	}

	d.mu.Lock()
	d.ReadReplicas = append(d.ReadReplicas, db)
	d.mu.Unlock()

	d.Logger.WithFields(logrus.Fields{
		"host": cfg.Host,
		"port": cfg.Port,
	}).Info("Read replica added successfully")

	return nil
}

// GetReadReplica returns a read replica connection using round-robin
func (d *Database) GetReadReplica() *sql.DB {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.ReadReplicas) == 0 {
		return d.SQLDB
	}

	// Simple round-robin selection
	index := int(atomic.AddInt64(&d.metrics.WaitCount, 1)) % len(d.ReadReplicas)
	return d.ReadReplicas[index]
}

// Close gracefully closes all database connections
func (d *Database) Close() error {
	if atomic.CompareAndSwapInt32(&d.closed, 0, 1) {
		// Stop health monitoring
		if d.healthCheckTicker != nil {
			d.healthCheckTicker.Stop()
		}

		// Close read replicas
		d.mu.Lock()
		for _, replica := range d.ReadReplicas {
			if replica != nil {
				replica.Close()
			}
		}
		d.ReadReplicas = nil
		d.mu.Unlock()

		// Close main connection
		if d.SQLDB != nil {
			d.Logger.Info("Closing database connection")
			return d.SQLDB.Close()
		}
	}
	return nil
}