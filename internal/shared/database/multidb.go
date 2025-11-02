package database

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/VincentArjuna/RexiErp/internal/shared/config"
)

// RedisClient interface for both single and cluster Redis clients
type RedisClient interface {
	Ping(ctx context.Context) *redis.StatusCmd
	Close() error
}

// MultiDBManager manages multiple database connections
type MultiDBManager struct {
	databases map[string]*Database
	redis     RedisClient
	minio     *minio.Client
	logger    *logrus.Logger
	config    *config.Config
	mu        sync.RWMutex
}

// NewMultiDBManager creates a new multi-database manager
func NewMultiDBManager(cfg *config.Config, logger *logrus.Logger) *MultiDBManager {
	return &MultiDBManager{
		databases: make(map[string]*Database),
		logger:    logger,
		config:    cfg,
	}
}

// Initialize initializes all database connections
func (mdb *MultiDBManager) Initialize(ctx context.Context) error {
	// Initialize master database
	if err := mdb.initializeDatabase(ctx, "master", &mdb.config.Databases.Master); err != nil {
		return fmt.Errorf("failed to initialize master database: %w", err)
	}

	// Initialize replica databases
	for i, replica := range mdb.config.Databases.Replicas {
		if !replica.Enabled {
			continue
		}
		name := fmt.Sprintf("replica_%d", i+1)
		if err := mdb.initializeDatabase(ctx, name, &replica); err != nil {
			mdb.logger.WithError(err).WithField("replica", name).Warn("Failed to initialize replica database")
		}
	}

	// Initialize Redis
	if err := mdb.initializeRedis(ctx); err != nil {
		return fmt.Errorf("failed to initialize Redis: %w", err)
	}

	// Initialize MinIO
	if err := mdb.initializeMinIO(ctx); err != nil {
		return fmt.Errorf("failed to initialize MinIO: %w", err)
	}

	mdb.logger.Info("Multi-database manager initialized successfully")

	return nil
}

// initializeDatabase initializes a single database connection
func (mdb *MultiDBManager) initializeDatabase(ctx context.Context, name string, cfg *config.DatabaseConfig) error {
	if cfg.Type == "" {
		cfg.Type = config.DatabaseTypePostgreSQL // Default to PostgreSQL
	}

	var dialector gorm.Dialector
	var dsn string

	switch cfg.Type {
	case config.DatabaseTypePostgreSQL:
		// Use secure connection string construction with URL encoding
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=10",
			cfg.Host, cfg.Port, cfg.User, escapePassword(cfg.Password), cfg.Name, cfg.SSLMode)
		dialector = postgres.Open(dsn)

	case config.DatabaseTypeMySQL:
		// Use secure connection string with URL encoding for credentials
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s",
			cfg.User, escapePassword(cfg.Password), cfg.Host, cfg.Port, cfg.Name)
		dialector = mysql.Open(dsn)

	case config.DatabaseTypeSQLite:
		dsn = cfg.Name
		dialector = sqlite.Open(dsn)

	default:
		return fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	// Configure GORM
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:         mdb.getGORMLogger(),
		PrepareStmt:    true,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database %s: %w", name, err)
	}

	// Get underlying SQL DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB for %s: %w", name, err)
	}

	// Configure connection pool
	if err := configureConnectionPool(sqlDB, cfg, mdb.logger); err != nil {
		return fmt.Errorf("failed to configure connection pool for %s: %w", name, err)
	}

	// Test connection
	if err := testConnection(sqlDB, mdb.logger); err != nil {
		return fmt.Errorf("database connection test failed for %s: %w", name, err)
	}

	// Create database instance
	database := &Database{
		DB:     db,
		SQLDB:  sqlDB,
		Logger: mdb.logger,
		Config: cfg,
	}

	// Start health monitoring
	if err := database.startHealthMonitoring(); err != nil {
		mdb.logger.WithError(err).Warn("Failed to start health monitoring")
	}

	mdb.mu.Lock()
	mdb.databases[name] = database
	mdb.mu.Unlock()

	mdb.logger.WithFields(logrus.Fields{
		"name":   name,
		"type":   cfg.Type,
		"host":   cfg.Host,
		"port":   cfg.Port,
		"master": cfg.IsMaster,
	}).Info("Database connection established")

	return nil
}

// initializeRedis initializes Redis connection
func (mdb *MultiDBManager) initializeRedis(ctx context.Context) error {
	cfg := mdb.config.Redis

	var redisClient RedisClient

	if cfg.Cluster.Enabled {
		// Redis Cluster
		redisClient = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        cfg.Cluster.Nodes,
			Password:     cfg.Cluster.Password,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.MinIdleConns,
			MaxRetries:   cfg.MaxRetries,
			DialTimeout:  cfg.DialTimeout,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			PoolTimeout:  cfg.PoolTimeout,
			IdleTimeout:  cfg.IdleTimeout,
			IdleCheckFrequency: cfg.IdleCheckFrequency,
		})
	} else if cfg.Sentinel.Enabled {
		// Redis Sentinel
		redisClient = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    cfg.Sentinel.Master,
			SentinelAddrs: cfg.Sentinel.Nodes,
			Password:      cfg.Sentinel.Password,
			PoolSize:      cfg.PoolSize,
			MinIdleConns:  cfg.MinIdleConns,
			MaxRetries:    cfg.MaxRetries,
			DialTimeout:   cfg.DialTimeout,
			ReadTimeout:   cfg.ReadTimeout,
			WriteTimeout:  cfg.WriteTimeout,
			PoolTimeout:   cfg.PoolTimeout,
			IdleTimeout:   cfg.IdleTimeout,
			IdleCheckFrequency: cfg.IdleCheckFrequency,
		})
	} else {
		// Single Redis instance
		redisClient = redis.NewClient(&redis.Options{
			Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Password:     cfg.Password,
			DB:           cfg.DB,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.MinIdleConns,
			MaxRetries:   cfg.MaxRetries,
			DialTimeout:  cfg.DialTimeout,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			PoolTimeout:  cfg.PoolTimeout,
			IdleTimeout:  cfg.IdleTimeout,
			IdleCheckFrequency: cfg.IdleCheckFrequency,
		})
	}

	// Test connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	mdb.redis = redisClient

	mdb.logger.WithFields(logrus.Fields{
		"host": cfg.Host,
		"port": cfg.Port,
		"db":   cfg.DB,
	}).Info("Redis connection established")

	return nil
}

// initializeMinIO initializes MinIO connection
func (mdb *MultiDBManager) initializeMinIO(ctx context.Context) error {
	cfg := mdb.config.MinIO

	// Create MinIO client
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Test connection and create bucket if it doesn't exist
	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	exists, err := minioClient.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return fmt.Errorf("failed to check MinIO bucket: %w", err)
	}

	if !exists {
		err = minioClient.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{
			Region: cfg.Region,
		})
		if err != nil {
			return fmt.Errorf("failed to create MinIO bucket: %w", err)
		}
		mdb.logger.WithField("bucket", cfg.Bucket).Info("MinIO bucket created")
	}

	mdb.minio = minioClient

	mdb.logger.WithFields(logrus.Fields{
		"endpoint": cfg.Endpoint,
		"bucket":   cfg.Bucket,
		"use_ssl":  cfg.UseSSL,
		"region":   cfg.Region,
	}).Info("MinIO connection established")

	return nil
}

// getGORMLogger returns the GORM logger configuration
func (mdb *MultiDBManager) getGORMLogger() gormlogger.Interface {
	return gormlogger.Default.LogMode(gormlogger.Silent)
}

// GetDatabase returns a database connection by name
func (mdb *MultiDBManager) GetDatabase(name string) (*Database, error) {
	mdb.mu.RLock()
	defer mdb.mu.RUnlock()

	db, exists := mdb.databases[name]
	if !exists {
		return nil, fmt.Errorf("database %s not found", name)
	}

	return db, nil
}

// GetMaster returns the master database connection
func (mdb *MultiDBManager) GetMaster() (*Database, error) {
	return mdb.GetDatabase("master")
}

// GetReadReplica returns a read replica connection (round-robin)
func (mdb *MultiDBManager) GetReadReplica() (*Database, error) {
	mdb.mu.RLock()
	defer mdb.mu.RUnlock()

	// Filter enabled replicas
	var replicas []string
	for name := range mdb.databases {
		if strings.HasPrefix(name, "replica_") {
			replicas = append(replicas, name)
		}
	}

	if len(replicas) == 0 {
		// Fallback to master if no replicas
		return mdb.GetDatabase("master")
	}

	// Simple round-robin selection
	index := int(time.Now().UnixNano()) % len(replicas)
	selected := replicas[index]

	return mdb.databases[selected], nil
}

// GetRedis returns the Redis client
func (mdb *MultiDBManager) GetRedis() RedisClient {
	return mdb.redis
}

// GetMinIO returns the MinIO client
func (mdb *MultiDBManager) GetMinIO() *minio.Client {
	return mdb.minio
}

// HealthCheck performs health checks on all connections
func (mdb *MultiDBManager) HealthCheck(ctx context.Context) error {
	var errors []string

	// Check master database
	if master, err := mdb.GetMaster(); err == nil {
		if err := master.HealthCheck(); err != nil {
			errors = append(errors, fmt.Sprintf("master database: %v", err))
		}
	} else {
		errors = append(errors, fmt.Sprintf("master database not available: %v", err))
	}

	// Check Redis
	if mdb.redis != nil {
		if err := mdb.redis.Ping(ctx).Err(); err != nil {
			errors = append(errors, fmt.Sprintf("Redis: %v", err))
		}
	} else {
		errors = append(errors, "Redis not available")
	}

	// Check MinIO
	if mdb.minio != nil {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if _, err := mdb.minio.ListBuckets(ctx); err != nil {
			errors = append(errors, fmt.Sprintf("MinIO: %v", err))
		}
	} else {
		errors = append(errors, "MinIO not available")
	}

	if len(errors) > 0 {
		return fmt.Errorf("health check failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// Close closes all database connections
func (mdb *MultiDBManager) Close() error {
	var errors []string

	// Close databases
	mdb.mu.Lock()
	for name, db := range mdb.databases {
		if err := db.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("failed to close database %s: %v", name, err))
		}
	}
	mdb.databases = make(map[string]*Database)
	mdb.mu.Unlock()

	// Close Redis
	if mdb.redis != nil {
		if err := mdb.redis.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("failed to close Redis: %v", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors during close: %s", strings.Join(errors, "; "))
	}

	return nil
}

// GetConnectionMetrics returns metrics for all connections
func (mdb *MultiDBManager) GetConnectionMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// Database metrics
	mdb.mu.RLock()
	dbMetrics := make(map[string]ConnectionMetrics)
	for name, db := range mdb.databases {
		dbMetrics[name] = db.GetMetrics()
	}
	mdb.mu.RUnlock()
	metrics["databases"] = dbMetrics

	// Redis metrics (only available for single client)
	if mdb.redis != nil {
		if client, ok := mdb.redis.(*redis.Client); ok {
			poolStats := client.PoolStats()
			metrics["redis"] = map[string]interface{}{
				"hits":        poolStats.Hits,
				"misses":      poolStats.Misses,
				"timeouts":    poolStats.Timeouts,
				"total_conns": poolStats.TotalConns,
				"idle_conns":  poolStats.IdleConns,
				"stale_conns": poolStats.StaleConns,
			}
		} else {
			metrics["redis"] = map[string]interface{}{
				"type": "cluster/sentinel (metrics not available)",
			}
		}
	}

	return metrics
}

// escapePassword properly escapes special characters in database passwords
func escapePassword(password string) string {
	// Use URL percent-encoding to escape special characters
	return url.QueryEscape(password)
}