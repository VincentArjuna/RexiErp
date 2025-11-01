package cache

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/VincentArjuna/RexiErp/internal/shared/config"
)

func TestRedisCache_Configuration(t *testing.T) {
	config := &config.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
		PoolSize: 10,
	}

	assert.NotNil(t, config)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 6379, config.Port)
	assert.Equal(t, 0, config.DB)
	assert.Equal(t, 10, config.PoolSize)
}

// TestNewRedisCache_Integration is an integration test that requires a real Redis instance
// This test should be run in a CI/CD environment with testcontainers or similar
func TestNewRedisCache_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := &config.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       1, // Use test database
		PoolSize: 5,
	}

	// This test will fail if Redis is not running
	cache, err := NewRedisCache(config, logger)
	if err != nil {
		t.Skipf("Redis not available for integration test: %v", err)
		return
	}
	defer cache.Close()

	ctx := context.Background()

	// Test basic operations
	key := "test-integration-key"
	value := map[string]interface{}{
		"field1": "value1",
		"field2": 123,
	}

	// Test Set
	err = cache.Set(ctx, key, value, time.Minute)
	assert.NoError(t, err)

	// Test Get
	var retrieved map[string]interface{}
	err = cache.Get(ctx, key, &retrieved)
	assert.NoError(t, err)
	assert.Equal(t, "value1", retrieved["field1"])
	assert.Equal(t, float64(123), retrieved["field2"])

	// Test Exists
	exists, err := cache.Exists(ctx, key)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test Delete
	err = cache.Delete(ctx, key)
	assert.NoError(t, err)

	// Verify deletion
	exists, err = cache.Exists(ctx, key)
	assert.NoError(t, err)
	assert.False(t, exists)

	// Test HealthCheck
	err = cache.HealthCheck()
	assert.NoError(t, err)
}