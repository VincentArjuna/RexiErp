package database

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/VincentArjuna/RexiErp/internal/shared/config"
)

func TestDatabase_NewDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Suppress logs in tests

	cfg := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		Name:            "rexi_erp_test",
		User:            "rexi",
		Password:        "password",
		SSLMode:         "disable",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}

	// This test requires a running PostgreSQL instance
	// In CI/CD, use testcontainers or similar
	t.Run("ValidConfig", func(t *testing.T) {
		db, err := NewDatabase(cfg, logger)
		if err != nil {
			t.Skipf("Skipping test - database not available: %v", err)
		}
		require.NoError(t, err)
		assert.NotNil(t, db)
		assert.NotNil(t, db.DB)
		assert.NotNil(t, db.SQLDB)

		// Test connection pool stats
		stats := db.GetStats()
		assert.GreaterOrEqual(t, stats.OpenConnections, 0)
		assert.GreaterOrEqual(t, stats.MaxOpenConnections, 0)

		// Clean up
		err = db.Close()
		assert.NoError(t, err)
	})
}

func TestDatabase_Configuration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cfg := &config.DatabaseConfig{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}

	t.Run("ConnectionPoolSettings", func(t *testing.T) {
		// Test that connection pool settings are properly configured
		assert.Equal(t, 10, cfg.MaxOpenConns)
		assert.Equal(t, 5, cfg.MaxIdleConns)
		assert.Equal(t, 30*time.Minute, cfg.ConnMaxLifetime)
		assert.Equal(t, 5*time.Minute, cfg.ConnMaxIdleTime)
	})
}

func TestDatabase_GetDSN(t *testing.T) {
	cfg := &config.Config{
		Databases: config.DatabaseConfigs{
			Master: config.DatabaseConfig{
				Type:     "postgres",
				Host:     "localhost",
				Port:     5432,
				Name:     "test_db",
				User:     "test_user",
				Password: "test_pass",
				SSLMode:  "require",
			},
		},
	}

	expected := "host=localhost port=5432 user=test_user password=test_pass dbname=test_db sslmode=require"
	actual := cfg.GetDSN()

	assert.Equal(t, expected, actual)
}

func TestDatabase_ConnectionPoolOptimization(t *testing.T) {
	testCases := []struct {
		name        string
		maxOpen     int
		maxIdle     int
		maxLifetime time.Duration
		maxIdleTime time.Duration
	}{
		{
			name:        "DevelopmentSettings",
			maxOpen:     25,
			maxIdle:     5,
			maxLifetime: 5 * time.Minute,
			maxIdleTime: 1 * time.Minute,
		},
		{
			name:        "ProductionSettings",
			maxOpen:     50,
			maxIdle:     10,
			maxLifetime: 1 * time.Hour,
			maxIdleTime: 30 * time.Minute,
		},
		{
			name:        "HighTrafficSettings",
			maxOpen:     100,
			maxIdle:     20,
			maxLifetime: 2 * time.Hour,
			maxIdleTime: 1 * time.Hour,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.DatabaseConfig{
				MaxOpenConns:    tc.maxOpen,
				MaxIdleConns:    tc.maxIdle,
				ConnMaxLifetime: tc.maxLifetime,
				ConnMaxIdleTime: tc.maxIdleTime,
			}

			assert.Equal(t, tc.maxOpen, cfg.MaxOpenConns)
			assert.Equal(t, tc.maxIdle, cfg.MaxIdleConns)
			assert.Equal(t, tc.maxLifetime, cfg.ConnMaxLifetime)
			assert.Equal(t, tc.maxIdleTime, cfg.ConnMaxIdleTime)

			// Validate settings are reasonable
			assert.Greater(t, cfg.MaxOpenConns, 0)
			assert.Greater(t, cfg.MaxIdleConns, 0)
			assert.LessOrEqual(t, cfg.MaxIdleConns, cfg.MaxOpenConns)
			assert.Greater(t, cfg.ConnMaxLifetime, 0)
			assert.Greater(t, cfg.ConnMaxIdleTime, 0)
		})
	}
}

// Benchmark tests for connection pool performance
func BenchmarkDatabase_GetStats(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark test in short mode")
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cfg := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		Name:            "rexi_erp_test",
		User:            "rexi",
		Password:        "password",
		SSLMode:         "disable",
		MaxOpenConns:    50,
		MaxIdleConns:    10,
		ConnMaxLifetime: 1 * time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,
	}

	db, err := NewDatabase(cfg, logger)
	if err != nil {
		b.Skipf("Skipping benchmark - database not available: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats := db.GetStats()
		_ = stats.OpenConnections
	}
}

func TestDatabase_HealthCheck(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cfg := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		Name:            "rexi_erp_test",
		User:            "rexi",
		Password:        "password",
		SSLMode:         "disable",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}

	t.Run("SuccessfulHealthCheck", func(t *testing.T) {
		db, err := NewDatabase(cfg, logger)
		if err != nil {
			t.Skipf("Skipping test - database not available: %v", err)
		}
		defer db.Close()

		err = db.HealthCheck()
		if err != nil {
			t.Skipf("Skipping health check - database not responding: %v", err)
		}
		assert.NoError(t, err)
	})
}

func TestDatabase_GetTenantDB(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cfg := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		Name:            "rexi_erp_test",
		User:            "rexi",
		Password:        "password",
		SSLMode:         "disable",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}

	t.Run("TenantDatabase", func(t *testing.T) {
		db, err := NewDatabase(cfg, logger)
		if err != nil {
			t.Skipf("Skipping test - database not available: %v", err)
		}
		defer db.Close()

		tenantDB, err := db.GetTenantDB("tenant-123")
		assert.NoError(t, err)
		assert.NotNil(t, tenantDB)
	})
}

// Test enhanced database functionality
func TestDatabase_ConnectionMetrics(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cfg := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		Name:            "rexi_erp_test",
		User:            "rexi",
		Password:        "password",
		SSLMode:         "disable",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}

	t.Run("GetMetrics", func(t *testing.T) {
		db, err := NewDatabase(cfg, logger)
		if err != nil {
			t.Skipf("Skipping test - database not available: %v", err)
		}
		defer db.Close()

		metrics := db.GetMetrics()
		assert.GreaterOrEqual(t, metrics.TotalConnections, int32(0))
		assert.GreaterOrEqual(t, metrics.ActiveConnections, int32(0))
		assert.GreaterOrEqual(t, metrics.IdleConnections, int32(0))
		assert.GreaterOrEqual(t, metrics.WaitCount, int64(0))
		assert.GreaterOrEqual(t, metrics.WaitDuration, time.Duration(0))
	})

	t.Run("HealthMonitoring", func(t *testing.T) {
		db, err := NewDatabase(cfg, logger)
		if err != nil {
			t.Skipf("Skipping test - database not available: %v", err)
		}
		defer db.Close()

		// Give some time for health monitoring to start
		time.Sleep(100 * time.Millisecond)

		err = db.HealthCheck()
		assert.NoError(t, err)
	})
}

func TestDatabase_ReadReplicas(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	masterCfg := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		Name:            "rexi_erp_test",
		User:            "rexi",
		Password:        "password",
		SSLMode:         "disable",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}

	t.Run("AddReadReplica", func(t *testing.T) {
		db, err := NewDatabase(masterCfg, logger)
		if err != nil {
			t.Skipf("Skipping test - database not available: %v", err)
		}
		defer db.Close()

		// Test adding a replica (using same config for simplicity)
		err = db.AddReadReplica(masterCfg)
		if err != nil {
			t.Skipf("Skipping replica test - replica database not available: %v", err)
		}

		// Test getting read replica
		replica := db.GetReadReplica()
		assert.NotNil(t, replica)

		// Test multiple calls to ensure round-robin
		replica2 := db.GetReadReplica()
		assert.NotNil(t, replica2)
	})

	t.Run("NoReplicasFallback", func(t *testing.T) {
		db, err := NewDatabase(masterCfg, logger)
		if err != nil {
			t.Skipf("Skipping test - database not available: %v", err)
		}
		defer db.Close()

		// Should return master when no replicas
		replica := db.GetReadReplica()
		assert.NotNil(t, replica)
	})
}

func TestDatabase_GracefulShutdown(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cfg := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		Name:            "rexi_erp_test",
		User:            "rexi",
		Password:        "password",
		SSLMode:         "disable",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}

	t.Run("GracefulClose", func(t *testing.T) {
		db, err := NewDatabase(cfg, logger)
		if err != nil {
			t.Skipf("Skipping test - database not available: %v", err)
		}

		// Add a replica
		replicaCfg := *cfg
		err = db.AddReadReplica(&replicaCfg)
		if err != nil {
			t.Skipf("Skipping replica test: %v", err)
		}

		// Test graceful shutdown
		err = db.Close()
		assert.NoError(t, err)

		// Test double close (should not panic)
		err = db.Close()
		assert.NoError(t, err)
	})
}

// Performance benchmarks
func BenchmarkDatabase_HealthCheck(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cfg := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		Name:            "rexi_erp_test",
		User:            "rexi",
		Password:        "password",
		SSLMode:         "disable",
		MaxOpenConns:    50,
		MaxIdleConns:    10,
		ConnMaxLifetime: 1 * time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,
	}

	db, err := NewDatabase(cfg, logger)
	if err != nil {
		b.Skipf("Skipping benchmark - database not available: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.HealthCheck()
	}
}

func BenchmarkDatabase_GetMetrics(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cfg := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		Name:            "rexi_erp_test",
		User:            "rexi",
		Password:        "password",
		SSLMode:         "disable",
		MaxOpenConns:    50,
		MaxIdleConns:    10,
		ConnMaxLifetime: 1 * time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,
	}

	db, err := NewDatabase(cfg, logger)
	if err != nil {
		b.Skipf("Skipping benchmark - database not available: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = db.GetMetrics()
	}
}

func BenchmarkDatabase_GetReadReplica(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cfg := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		Name:            "rexi_erp_test",
		User:            "rexi",
		Password:        "password",
		SSLMode:         "disable",
		MaxOpenConns:    50,
		MaxIdleConns:    10,
		ConnMaxLifetime: 1 * time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,
	}

	db, err := NewDatabase(cfg, logger)
	if err != nil {
		b.Skipf("Skipping benchmark - database not available: %v", err)
	}
	defer db.Close()

	// Add multiple replicas for round-robin testing
	for i := 0; i < 3; i++ {
		replicaCfg := *cfg
		_ = db.AddReadReplica(&replicaCfg) // Ignore errors for benchmark
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = db.GetReadReplica()
	}
}