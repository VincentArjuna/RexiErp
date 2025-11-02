package database

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/VincentArjuna/RexiErp/internal/shared/config"
)

func TestMultiDatabaseManager_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test configuration
	cfg := &config.Config{
		Databases: config.DatabaseConfigs{
			Master: config.DatabaseConfig{
				Type:            config.DatabaseTypePostgreSQL,
				Host:            getEnv("TEST_DB_HOST", "localhost"),
				Port:            getEnvInt("TEST_DB_PORT", 5432),
				Name:            getEnv("TEST_DB_NAME", "rexi_erp_test"),
				User:            getEnv("TEST_DB_USER", "rexi"),
				Password:        getEnv("TEST_DB_PASSWORD", "password"),
				SSLMode:         "disable",
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: 1 * time.Hour,
				ConnMaxIdleTime: 30 * time.Minute,
				IsMaster:        true,
				Enabled:         true,
			},
			Replicas: []config.DatabaseConfig{
				{
					Type:            config.DatabaseTypePostgreSQL,
					Host:            getEnv("TEST_DB_HOST", "localhost"),
					Port:            getEnvInt("TEST_DB_PORT", 5432),
					Name:            getEnv("TEST_DB_NAME", "rexi_erp_test"),
					User:            getEnv("TEST_DB_USER", "rexi"),
					Password:        getEnv("TEST_DB_PASSWORD", "password"),
					SSLMode:         "disable",
					MaxOpenConns:    5,
					MaxIdleConns:    2,
					ConnMaxLifetime: 1 * time.Hour,
					ConnMaxIdleTime: 30 * time.Minute,
					IsMaster:        false,
					Enabled:         true,
				},
			},
		},
		Redis: config.RedisConfig{
			Host:     getEnv("TEST_REDIS_HOST", "localhost"),
			Port:     getEnvInt("TEST_REDIS_PORT", 6379),
			Password: getEnv("TEST_REDIS_PASSWORD", ""),
			DB:       1, // Use different DB for testing
			PoolSize: 10,
		},
		MinIO: config.MinIOConfig{
			Endpoint:        getEnv("TEST_MINIO_ENDPOINT", "localhost:9000"),
			AccessKeyID:     getEnv("TEST_MINIO_ACCESS_KEY", "minioadmin"),
			SecretAccessKey: getEnv("TEST_MINIO_SECRET_KEY", "minioadmin"),
			UseSSL:          false,
			Region:          "us-east-1",
			Bucket:          "test-bucket-" + uuid.New().String()[:8],
			Timeout:         30 * time.Second,
			RetryAttempts:   3,
		},
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	ctx := context.Background()

	t.Run("InitializeMultiDB", func(t *testing.T) {
		multiDB := NewMultiDBManager(cfg, logger)
		defer multiDB.Close()

		err := multiDB.Initialize(ctx)
		if err != nil {
			t.Skipf("Skipping multi-database test - services not available: %v", err)
		}

		// Test master database
		master, err := multiDB.GetMaster()
		assert.NoError(t, err)
		assert.NotNil(t, master)

		// Test read replica
		replica, err := multiDB.GetReadReplica()
		assert.NoError(t, err)
		assert.NotNil(t, replica)

		// Test health check
		err = multiDB.HealthCheck(ctx)
		assert.NoError(t, err)

		// Test connection metrics
		metrics := multiDB.GetConnectionMetrics()
		assert.NotNil(t, metrics)
		assert.Contains(t, metrics, "databases")
	})

	t.Run("ReplicaFailover", func(t *testing.T) {
		multiDB := NewMultiDBManager(cfg, logger)
		defer multiDB.Close()

		err := multiDB.Initialize(ctx)
		if err != nil {
			t.Skipf("Skipping replica failover test - services not available: %v", err)
		}

		// Test multiple read replica calls
		for i := 0; i < 10; i++ {
			replica, err := multiDB.GetReadReplica()
			assert.NoError(t, err)
			assert.NotNil(t, replica)
		}

		// Should work even if replicas are not available (fallback to master)
	})
}

func TestGORMIntegration_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDatabase(t)
	defer cleanupTestDatabase(t, db)

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	gormConfig := &GORMConfig{
		Logger:     logger,
		LogLevel:   gormlogger.Silent,
		PrepareStmt: true,
	}

	_ = NewGORMManager(db, gormConfig) // Created but not used in this test

	t.Run("RepositoryPattern", func(t *testing.T) {
		// Create a test model
		type TestModel struct {
			BaseModel
			Name  string `gorm:"not null"`
			Value int    `gorm:"default:0"`
		}

		// Auto migrate
		err := db.AutoMigrate(&TestModel{})
		require.NoError(t, err)

		repo := NewRepository[TestModel](db)

		tenantID := uuid.New()

		// Test Create
		testModel := &TestModel{
			BaseModel: BaseModel{
				TenantID: tenantID,
			},
			Name:  "Test",
			Value: 42,
		}

		ctx := context.WithValue(context.Background(), "tenant_id", tenantID.String())
		err = repo.Create(ctx, testModel)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, testModel.ID)

		// Test FindByID
		found, err := repo.FindByID(ctx, testModel.ID)
		assert.NoError(t, err)
		assert.Equal(t, testModel.ID, found.ID)
		assert.Equal(t, tenantID, found.TenantID)

		// Test Update
		updates := map[string]interface{}{
			"value": 100,
		}
		err = repo.Update(ctx, testModel.ID, updates)
		assert.NoError(t, err)

		// Verify update
		updated, err := repo.FindByID(ctx, testModel.ID)
		assert.NoError(t, err)
		assert.Equal(t, 100, updated.Value)

		// Test Delete (soft delete)
		err = repo.Delete(ctx, testModel.ID)
		assert.NoError(t, err)

		// Should still find with unscoped query
		err = repo.HardDelete(ctx, testModel.ID)
		assert.NoError(t, err)

		// Should not find anymore
		_, err = repo.FindByID(ctx, testModel.ID)
		assert.Error(t, err) // Should return record not found
	})

	t.Run("TransactionManager", func(t *testing.T) {
		type TestModel struct {
			BaseModel
			Name string `gorm:"not null"`
		}

		err := db.AutoMigrate(&TestModel{})
		require.NoError(t, err)

		txManager := NewTransactionManager(db)
		tenantID := uuid.New()

		t.Run("SuccessfulTransaction", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "tenant_id", tenantID.String())

			err := txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
				model1 := &TestModel{
					BaseModel: BaseModel{TenantID: tenantID},
					Name:      "Model1",
				}
				if err := tx.Create(model1).Error; err != nil {
					return err
				}

				model2 := &TestModel{
					BaseModel: BaseModel{TenantID: tenantID},
					Name:      "Model2",
				}
				return tx.Create(model2).Error
			})

			assert.NoError(t, err)

			// Verify both records exist
			repo := NewRepository[TestModel](db)
			count, err := repo.Count(ctx, map[string]interface{}{})
			assert.NoError(t, err)
			assert.Equal(t, int64(2), count)
		})

		t.Run("FailedTransaction", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "tenant_id", tenantID.String())

			err := txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
				model1 := &TestModel{
					BaseModel: BaseModel{TenantID: tenantID},
					Name:      "Model3",
				}
				if err := tx.Create(model1).Error; err != nil {
					return err
				}

				// Return an error to trigger rollback
				return assert.AnError
			})

			assert.Error(t, err)

			// Verify no new records were added
			repo := NewRepository[TestModel](db)
			count, err := repo.Count(ctx, map[string]interface{}{})
			assert.NoError(t, err)
			assert.Equal(t, int64(2), count) // Should still be 2 from previous test
		})
	})

	t.Run("TenantIsolation", func(t *testing.T) {
		type TestModel struct {
			BaseModel
			Name string `gorm:"not null"`
		}

		err := db.AutoMigrate(&TestModel{})
		require.NoError(t, err)

		tenant1 := uuid.New()
		tenant2 := uuid.New()

		repo := NewRepository[TestModel](db)

		// Create record for tenant1
		ctx1 := context.WithValue(context.Background(), "tenant_id", tenant1.String())
		model1 := &TestModel{
			BaseModel: BaseModel{TenantID: tenant1},
			Name:      "Tenant1 Model",
		}
		err = repo.Create(ctx1, model1)
		require.NoError(t, err)

		// Create record for tenant2
		ctx2 := context.WithValue(context.Background(), "tenant_id", tenant2.String())
		model2 := &TestModel{
			BaseModel: BaseModel{TenantID: tenant2},
			Name:      "Tenant2 Model",
		}
		err = repo.Create(ctx2, model2)
		require.NoError(t, err)

		// Test tenant isolation
		count1, err := repo.Count(ctx1, map[string]interface{}{})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count1)

		count2, err := repo.Count(ctx2, map[string]interface{}{})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count2)

		// Verify tenant1 cannot see tenant2's data
		all1, err := repo.FindAll(ctx1, map[string]interface{}{})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(all1))
		assert.Equal(t, tenant1, all1[0].TenantID)

		// Verify tenant2 cannot see tenant1's data
		all2, err := repo.FindAll(ctx2, map[string]interface{}{})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(all2))
		assert.Equal(t, tenant2, all2[0].TenantID)
	})
}

func TestHealthCheck_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDatabase(t)
	defer cleanupTestDatabase(t, db)

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	database := &Database{
		DB:     db,
		Logger: logger,
		Config: &config.DatabaseConfig{
			Host: "localhost",
			Port: 5432,
			Name: "test",
		},
	}

	// Initialize database components
	sqlDB, err := db.DB()
	if err != nil {
		t.Skipf("Skipping health check test - database not available: %v", err)
	}
	database.SQLDB = sqlDB

	healthChecker := NewHealthChecker(nil, logger, nil)

	// Since we don't have a full MultiDBManager, we'll test individual components
	t.Run("DatabaseHealth", func(t *testing.T) {
		health := healthChecker.checkDatabase(context.Background(), "master", nil)

		// Should be healthy if database is available
		if database.SQLDB != nil {
			err := database.SQLDB.Ping()
			if err == nil {
				assert.Equal(t, HealthStatusHealthy, health.Status)
			} else {
				assert.Equal(t, HealthStatusUnhealthy, health.Status)
			}
		}
	})
}

func TestPerformance_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	db := setupTestDatabase(t)
	defer cleanupTestDatabase(t, db)

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("ConcurrentOperations", func(t *testing.T) {
		type TestModel struct {
			BaseModel
			Name  string `gorm:"not null"`
			Value int    `gorm:"default:0"`
		}

		err := db.AutoMigrate(&TestModel{})
		require.NoError(t, err)

		repo := NewRepository[TestModel](db)
		tenantID := uuid.New()

		const numGoroutines = 50
		const operationsPerGoroutine = 100

		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines)

		start := time.Now()

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				ctx := context.WithValue(context.Background(), "tenant_id", tenantID.String())

				for j := 0; j < operationsPerGoroutine; j++ {
					model := &TestModel{
						BaseModel: BaseModel{TenantID: tenantID},
						Name:      fmt.Sprintf("Model-%d-%d", goroutineID, j),
						Value:     goroutineID*1000 + j,
					}

					err := repo.Create(ctx, model)
					if err != nil {
						errors <- err
						return
					}

					// Read it back
					_, err = repo.FindByID(ctx, model.ID)
					if err != nil {
						errors <- err
						return
					}
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		duration := time.Since(start)

		// Check for errors
		for err := range errors {
			t.Errorf("Concurrent operation error: %v", err)
		}

		// Verify all records were created
		ctx := context.WithValue(context.Background(), "tenant_id", tenantID.String())
		count, err := repo.Count(ctx, map[string]interface{}{})
		assert.NoError(t, err)
		assert.Equal(t, int64(numGoroutines*operationsPerGoroutine), count)

		t.Logf("Completed %d operations in %v (%.2f ops/sec)",
			numGoroutines*operationsPerGoroutine*2, // 2 ops per iteration (create + read)
			duration,
			float64(numGoroutines*operationsPerGoroutine*2)/duration.Seconds())

		// Performance assertion - should complete within reasonable time
		assert.Less(t, duration, 30*time.Second)
	})

	t.Run("BulkOperations", func(t *testing.T) {
		type TestModel struct {
			BaseModel
			Name  string `gorm:"not null"`
			Value int    `gorm:"default:0"`
		}

		err := db.AutoMigrate(&TestModel{})
		require.NoError(t, err)

		bulkOp := NewBulkOperation(db, logger)
		tenantID := uuid.New()

		const batchSize = 1000
		const numBatches = 10

		ctx := context.WithValue(context.Background(), "tenant_id", tenantID.String())

		start := time.Now()

		// Create multiple batches
		for batch := 0; batch < numBatches; batch++ {
			var models []TestModel
			for i := 0; i < batchSize; i++ {
				models = append(models, TestModel{
					BaseModel: BaseModel{TenantID: tenantID},
					Name:      fmt.Sprintf("BulkModel-%d-%d", batch, i),
					Value:     batch*1000 + i,
				})
			}

			err := bulkOp.BulkInsert(ctx, &models, batchSize)
			assert.NoError(t, err)
		}

		duration := time.Since(start)

		// Verify all records were created
		repo := NewRepository[TestModel](db)
		count, err := repo.Count(ctx, map[string]interface{}{})
		assert.NoError(t, err)
		assert.Equal(t, int64(batchSize*numBatches), count)

		t.Logf("Bulk inserted %d records in %v (%.2f records/sec)",
			batchSize*numBatches,
			duration,
			float64(batchSize*numBatches)/duration.Seconds())

		// Performance assertion - should be fast
		assert.Less(t, duration, 10*time.Second)
	})

	t.Run("QueryPerformance", func(t *testing.T) {
		type TestModel struct {
			BaseModel
			Name  string `gorm:"not null;index"`
			Value int    `gorm:"index"`
		}

		err := db.AutoMigrate(&TestModel{})
		require.NoError(t, err)

		// Create test data
		const numRecords = 10000
		tenantID := uuid.New()

		start := time.Now()
		ctx := context.WithValue(context.Background(), "tenant_id", tenantID.String())

		// Bulk insert test data
		var models []TestModel
		for i := 0; i < numRecords; i++ {
			models = append(models, TestModel{
				BaseModel: BaseModel{TenantID: tenantID},
				Name:      fmt.Sprintf("Model-%d", i%100), // Only 100 unique names
				Value:     i,
			})
		}

		bulkOp := NewBulkOperation(db, logger)
		err = bulkOp.BulkInsert(ctx, &models, 1000)
		require.NoError(t, err)

		insertTime := time.Since(start)

		// Test query performance
		queryBuilder := NewQueryBuilder(db)

		start = time.Now()

		// Test filtered queries
		for i := 0; i < 100; i++ {
			conditions := map[string]interface{}{
				"name": fmt.Sprintf("Model-%d", i),
			}
			query := queryBuilder.BuildWhere(conditions)
			query = query.WithContext(ctx)

			var count int64
			err := query.Model(&TestModel{}).Count(&count).Error
			assert.NoError(t, err)
			assert.Greater(t, count, int64(0))
		}

		queryTime := time.Since(start)

		t.Logf("Inserted %d records in %v (%.2f records/sec)",
			numRecords, insertTime, float64(numRecords)/insertTime.Seconds())
		t.Logf("Executed 100 filtered queries in %v (%.2f queries/sec)",
			queryTime, 100.0/queryTime.Seconds())

		// Performance assertions
		assert.Less(t, insertTime, 30*time.Second)
		assert.Less(t, queryTime, 5*time.Second)
	})
}