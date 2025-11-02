package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestMigrationManager_Initialize(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDatabase(t)
	defer cleanupTestDatabase(t, db)

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tempDir := t.TempDir()
	mm := NewMigrationManager(db, logger, tempDir)

	// Create test migration files
	err := createTestMigrationFiles(tempDir)
	require.NoError(t, err)

	ctx := context.Background()
	err = mm.Initialize(ctx)
	assert.NoError(t, err)

	// Check that migrations table was created
	var count int64
	err = db.Raw("SELECT COUNT(*) FROM migrations").Scan(&count).Error
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(0))
}

func TestMigrationManager_GetPendingMigrations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDatabase(t)
	defer cleanupTestDatabase(t, db)

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tempDir := t.TempDir()
	mm := NewMigrationManager(db, logger, tempDir)

	// Create test migration files
	err := createTestMigrationFiles(tempDir)
	require.NoError(t, err)

	ctx := context.Background()
	err = mm.Initialize(ctx)
	require.NoError(t, err)

	pending, err := mm.GetPendingMigrations()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(pending)) // We created 2 test migrations

	// Check ordering
	assert.True(t, pending[0].Version < pending[1].Version)
}

func TestMigrationManager_MigrateUp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDatabase(t)
	defer cleanupTestDatabase(t, db)

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tempDir := t.TempDir()
	mm := NewMigrationManager(db, logger, tempDir)

	// Create test migration files
	err := createTestMigrationFiles(tempDir)
	require.NoError(t, err)

	ctx := context.Background()
	err = mm.Initialize(ctx)
	require.NoError(t, err)

	// Apply migrations
	err = mm.MigrateUp(ctx)
	assert.NoError(t, err)

	// Verify migrations were applied
	pending, err := mm.GetPendingMigrations()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(pending))

	applied, err := mm.GetAppliedMigrations()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(applied))

	// Verify tables were created
	var tables []string
	err = db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name LIKE 'test_%'").Scan(&tables).Error
	assert.NoError(t, err)
	assert.Contains(t, tables, "test_users")
	assert.Contains(t, tables, "test_products")
}

func TestMigrationManager_MigrateDown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDatabase(t)
	defer cleanupTestDatabase(t, db)

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tempDir := t.TempDir()
	mm := NewMigrationManager(db, logger, tempDir)

	// Create test migration files
	err := createTestMigrationFiles(tempDir)
	require.NoError(t, err)

	ctx := context.Background()
	err = mm.Initialize(ctx)
	require.NoError(t, err)

	// Apply migrations first
	err = mm.MigrateUp(ctx)
	require.NoError(t, err)

	// Rollback the last migration
	err = mm.MigrateDown(ctx, "002")
	assert.NoError(t, err)

	// Verify one migration is still applied
	pending, err := mm.GetPendingMigrations()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(pending))
	assert.Equal(t, "002", pending[0].Version)

	// Verify one table still exists
	var count int64
	err = db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'test_users'").Scan(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Verify the other table was dropped
	err = db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'test_products'").Scan(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestMigrationManager_Dependencies(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDatabase(t)
	defer cleanupTestDatabase(t, db)

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tempDir := t.TempDir()
	mm := NewMigrationManager(db, logger, tempDir)

	// Create test migration files with dependencies
	err := createTestMigrationFilesWithDependencies(tempDir)
	require.NoError(t, err)

	ctx := context.Background()
	err = mm.Initialize(ctx)
	require.NoError(t, err)

	// Try to apply migration with missing dependency
	pending, err := mm.GetPendingMigrations()
	assert.NoError(t, err)

	// Find the dependent migration
	var dependentMigration *MigrationFile
	for _, migration := range pending {
		if migration.Version == "003" {
			dependentMigration = migration
			break
		}
	}
	require.NotNil(t, dependentMigration)

	// Should fail due to missing dependency
	err = mm.applyMigration(ctx, dependentMigration)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dependency")
}

func TestMigrationManager_Hooks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDatabase(t)
	defer cleanupTestDatabase(t, db)

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tempDir := t.TempDir()
	mm := NewMigrationManager(db, logger, tempDir)

	// Create test migration files
	err := createTestMigrationFiles(tempDir)
	require.NoError(t, err)

	ctx := context.Background()
	err = mm.Initialize(ctx)
	require.NoError(t, err)

	// Add hooks
	preHookCalled := false
	postHookCalled := false

	mm.AddPreHook("001", func(ctx context.Context, db *gorm.DB, migration *Migration) error {
		preHookCalled = true
		return nil
	})

	mm.AddPostHook("001", func(ctx context.Context, db *gorm.DB, migration *Migration) error {
		postHookCalled = true
		return nil
	})

	// Apply migration
	err = mm.MigrateUp(ctx)
	assert.NoError(t, err)

	// Verify hooks were called
	assert.True(t, preHookCalled)
	assert.True(t, postHookCalled)
}

func TestMigrationManager_GetStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDatabase(t)
	defer cleanupTestDatabase(t, db)

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tempDir := t.TempDir()
	mm := NewMigrationManager(db, logger, tempDir)

	// Create test migration files
	err := createTestMigrationFiles(tempDir)
	require.NoError(t, err)

	ctx := context.Background()
	err = mm.Initialize(ctx)
	require.NoError(t, err)

	// Check initial status
	status, err := mm.GetStatus(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, status.TotalMigrations)
	assert.Equal(t, 0, status.AppliedMigrations)
	assert.Equal(t, 2, status.PendingMigrations)

	// Apply one migration
	pending, err := mm.GetPendingMigrations()
	require.NoError(t, err)
	require.Greater(t, len(pending), 0)

	err = mm.applyMigration(ctx, pending[0])
	require.NoError(t, err)

	// Check updated status
	status, err = mm.GetStatus(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, status.TotalMigrations)
	assert.Equal(t, 1, status.AppliedMigrations)
	assert.Equal(t, 1, status.PendingMigrations)
}

func TestMigrationManager_RollbackScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("RollbackWithData", func(t *testing.T) {
		db := setupTestDatabase(t)
		defer cleanupTestDatabase(t, db)

		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)

		tempDir := t.TempDir()
		mm := NewMigrationManager(db, logger, tempDir)

		// Create migration with data
		err := createTestMigrationWithData(tempDir)
		require.NoError(t, err)

		ctx := context.Background()
		err = mm.Initialize(ctx)
		require.NoError(t, err)

		// Apply migration
		err = mm.MigrateUp(ctx)
		require.NoError(t, err)

		// Insert test data
		err = db.Exec("INSERT INTO test_data (id, name) VALUES (?, ?)", uuid.New(), "test").Error
		require.NoError(t, err)

		// Verify data exists
		var count int64
		err = db.Raw("SELECT COUNT(*) FROM test_data").Scan(&count).Error
		require.NoError(t, err)
		assert.Greater(t, count, int64(0))

		// Rollback migration
		err = mm.MigrateDown(ctx, "001")
		assert.NoError(t, err)

		// Verify table was dropped (should not error)
		err = db.Raw("SELECT COUNT(*) FROM test_data").Scan(&count).Error
		assert.Error(t, err) // Table should not exist
	})

	t.Run("RollbackFailedMigration", func(t *testing.T) {
		db := setupTestDatabase(t)
		defer cleanupTestDatabase(t, db)

		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)

		tempDir := t.TempDir()
		mm := NewMigrationManager(db, logger, tempDir)

		// Create migration that will fail
		err := createTestMigrationFailScenario(tempDir)
		require.NoError(t, err)

		ctx := context.Background()
		err = mm.Initialize(ctx)
		require.NoError(t, err)

		// Apply migration (should fail)
		err = mm.MigrateUp(ctx)
		assert.Error(t, err)

		// Verify system is in consistent state
		status, err := mm.GetStatus(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, status.AppliedMigrations) // No migrations should be applied
	})
}

func TestSeedManager(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDatabase(t)
	defer cleanupTestDatabase(t, db)

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tempDir := t.TempDir()
	sm := NewSeedManager(db, logger, tempDir)

	// Create test seed files
	err := createTestSeedFiles(tempDir)
	require.NoError(t, err)

	ctx := context.Background()
	err = sm.LoadSeeders(ctx)
	assert.NoError(t, err)

	// Test seeding
	err = sm.Seed(ctx, "development", "test_users")
	assert.NoError(t, err)

	// Verify data was seeded
	var count int64
	err = db.Raw("SELECT COUNT(*) FROM users").Scan(&count).Error
	if err == nil {
		assert.Greater(t, count, int64(0))
	}

	// Test seeder status
	status := sm.GetSeederStatus()
	assert.Equal(t, 2, status.TotalSeeders)
	assert.Greater(t, status.AppliedSeeders, 0)
}

func TestMigrationPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	db := setupTestDatabase(t)
	defer cleanupTestDatabase(t, db)

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tempDir := t.TempDir()
	mm := NewMigrationManager(db, logger, tempDir)

	// Create multiple test migration files
	err := createMultipleTestMigrations(tempDir, 20)
	require.NoError(t, err)

	ctx := context.Background()
	err = mm.Initialize(ctx)
	require.NoError(t, err)

	// Measure migration time
	start := time.Now()
	err = mm.MigrateUp(ctx)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Less(t, duration, 5*time.Second) // Should complete within 5 seconds

	// Verify all migrations were applied
	status, err := mm.GetStatus(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 20, status.AppliedMigrations)
}

// Helper functions for testing

func setupTestDatabase(t *testing.T) *gorm.DB {
	// Use environment variables or defaults for test database
	host := getEnv("TEST_DB_HOST", "localhost")
	port := getEnvInt("TEST_DB_PORT", 5432)
	user := getEnv("TEST_DB_USER", "rexi")
	password := getEnv("TEST_DB_PASSWORD", "password")
	dbName := getEnv("TEST_DB_NAME", "rexi_erp_test")

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping test - database not available: %v", err)
	}

	// Clean up any existing test tables
	cleanupTestTables(t, db)

	return db
}

func cleanupTestDatabase(t *testing.T, db *gorm.DB) {
	cleanupTestTables(t, db)

	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

func cleanupTestTables(t *testing.T, db *gorm.DB) {
	tables := []string{"test_users", "test_products", "test_data", "migrations"}

	for _, table := range tables {
		db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table))
	}
}

func createTestMigrationFiles(dir string) error {
	// Create migration 001
	migration1Up := `CREATE TABLE test_users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		email VARCHAR(255) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);`

	migration1Down := `DROP TABLE IF EXISTS test_users;`

	err := os.WriteFile(filepath.Join(dir, "001_create_test_users.up.sql"), []byte(migration1Up), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(dir, "001_create_test_users.down.sql"), []byte(migration1Down), 0644)
	if err != nil {
		return err
	}

	// Create migration 002
	migration2Up := `CREATE TABLE test_products (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(255) NOT NULL,
		price DECIMAL(10,2),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);`

	migration2Down := `DROP TABLE IF EXISTS test_products;`

	err = os.WriteFile(filepath.Join(dir, "002_create_test_products.up.sql"), []byte(migration2Up), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(dir, "002_create_test_products.down.sql"), []byte(migration2Down), 0644)
	if err != nil {
		return err
	}

	return nil
}

func createTestMigrationFilesWithDependencies(dir string) error {
	// Create migration 001
	migration1Up := `-- depends:
	CREATE TABLE test_users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		email VARCHAR(255) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);`

	migration1Down := `DROP TABLE IF EXISTS test_users;`

	err := os.WriteFile(filepath.Join(dir, "001_create_test_users.up.sql"), []byte(migration1Up), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(dir, "001_create_test_users.down.sql"), []byte(migration1Down), 0644)
	if err != nil {
		return err
	}

	// Create migration 003 with dependency on 002 (which doesn't exist)
	migration3Up := `-- depends: 002
	CREATE TABLE test_orders (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);`

	migration3Down := `DROP TABLE IF EXISTS test_orders;`

	err = os.WriteFile(filepath.Join(dir, "003_create_test_orders.up.sql"), []byte(migration3Up), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(dir, "003_create_test_orders.down.sql"), []byte(migration3Down), 0644)
	if err != nil {
		return err
	}

	return nil
}

func createTestMigrationWithData(dir string) error {
	migrationUp := `CREATE TABLE test_data (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(255) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);`

	migrationDown := `DROP TABLE IF EXISTS test_data;`

	err := os.WriteFile(filepath.Join(dir, "001_create_test_data.up.sql"), []byte(migrationUp), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(dir, "001_create_test_data.down.sql"), []byte(migrationDown), 0644)
	if err != nil {
		return err
	}

	return nil
}

func createTestMigrationFailScenario(dir string) error {
	// Create a migration that will fail on rollback
	migrationUp := `CREATE TABLE test_fail (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(255) NOT NULL
	);`

	migrationDown := `DROP TABLE IF EXISTS test_fail;
	-- This will cause an error
	SELECT * FROM non_existent_table;`

	err := os.WriteFile(filepath.Join(dir, "001_create_test_fail.up.sql"), []byte(migrationUp), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(dir, "001_create_test_fail.down.sql"), []byte(migrationDown), 0644)
	if err != nil {
		return err
	}

	return nil
}

func createMultipleTestMigrations(dir string, count int) error {
	for i := 1; i <= count; i++ {
		version := fmt.Sprintf("%03d", i)
		tableName := fmt.Sprintf("test_table_%d", i)

		up := fmt.Sprintf(`CREATE TABLE %s (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);`, tableName)

		down := fmt.Sprintf(`DROP TABLE IF EXISTS %s;`, tableName)

		err := os.WriteFile(filepath.Join(dir, version+"_create_"+tableName+".up.sql"), []byte(up), 0644)
		if err != nil {
			return err
		}

		err = os.WriteFile(filepath.Join(dir, version+"_create_"+tableName+".down.sql"), []byte(down), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func createTestSeedFiles(dir string) error {
	// Create development seeder
	seeder1 := `{
		"name": "test_users",
		"description": "Test users for development",
		"environment": "development",
		"dependencies": [],
		"order": 1,
		"data": {
			"users": [
				{
					"id": "550e8400-e29b-41d4-a716-446655440001",
					"email": "test@example.com",
					"first_name": "Test",
					"last_name": "User"
				}
			]
		}
	}`

	err := os.WriteFile(filepath.Join(dir, "test_users.json"), []byte(seeder1), 0644)
	if err != nil {
		return err
	}

	// Create test seeder
	seeder2 := `{
		"name": "test_data",
		"description": "Test data for testing",
		"environment": "test",
		"dependencies": [],
		"order": 1,
		"data": {
			"test_records": [
				{
					"id": "550e8400-e29b-41d4-a716-446655440001",
					"name": "Test Record"
				}
			]
		}
	}`

	err = os.WriteFile(filepath.Join(dir, "test_data.json"), []byte(seeder2), 0644)
	if err != nil {
		return err
	}

	return nil
}

// Helper functions for environment variables
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}