package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"github.com/VincentArjuna/RexiErp/internal/shared/config"
)

// RedisCache wraps Redis client with logging and configuration
type RedisCache struct {
	Client *redis.Client
	Logger *logrus.Logger
	Config *config.RedisConfig
}

// NewRedisCache creates a new Redis cache connection
func NewRedisCache(cfg *config.RedisConfig, logger *logrus.Logger) (*RedisCache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"host":     cfg.Host,
		"port":     cfg.Port,
		"db":       cfg.DB,
		"poolSize": cfg.PoolSize,
	}).Info("Redis connection established")

	return &RedisCache{
		Client: rdb,
		Logger: logger,
		Config: cfg,
	}, nil
}

// Set stores a key-value pair with expiration
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := r.Client.Set(ctx, key, jsonValue, expiration).Err(); err != nil {
		r.Logger.WithFields(logrus.Fields{
			"key":   key,
			"error": err,
		}).Error("Failed to set cache value")
		return fmt.Errorf("failed to set cache value: %w", err)
	}

	r.Logger.WithFields(logrus.Fields{
		"key":        key,
		"expiration": expiration,
	}).Debug("Cache value set")

	return nil
}

// Get retrieves a value by key
func (r *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	value, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			r.Logger.WithField("key", key).Debug("Cache miss")
			return fmt.Errorf("key not found: %s", key)
		}
		r.Logger.WithFields(logrus.Fields{
			"key":   key,
			"error": err,
		}).Error("Failed to get cache value")
		return fmt.Errorf("failed to get cache value: %w", err)
	}

	if err := json.Unmarshal([]byte(value), dest); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	r.Logger.WithField("key", key).Debug("Cache hit")
	return nil
}

// Delete removes a key
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	if err := r.Client.Del(ctx, key).Err(); err != nil {
		r.Logger.WithFields(logrus.Fields{
			"key":   key,
			"error": err,
		}).Error("Failed to delete cache value")
		return fmt.Errorf("failed to delete cache value: %w", err)
	}

	r.Logger.WithField("key", key).Debug("Cache value deleted")
	return nil
}

// Exists checks if a key exists
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.Client.Exists(ctx, key).Result()
	if err != nil {
		r.Logger.WithFields(logrus.Fields{
			"key":   key,
			"error": err,
		}).Error("Failed to check cache key existence")
		return false, fmt.Errorf("failed to check cache key existence: %w", err)
	}

	exists := result > 0
	r.Logger.WithFields(logrus.Fields{
		"key":    key,
		"exists": exists,
	}).Debug("Cache key existence checked")

	return exists, nil
}

// SetAdd adds a member to a set
func (r *RedisCache) SetAdd(ctx context.Context, key string, member interface{}) error {
	if err := r.Client.SAdd(ctx, key, member).Err(); err != nil {
		return fmt.Errorf("failed to add to set: %w", err)
	}
	return nil
}

// SetRemove removes a member from a set
func (r *RedisCache) SetRemove(ctx context.Context, key string, member interface{}) error {
	if err := r.Client.SRem(ctx, key, member).Err(); err != nil {
		return fmt.Errorf("failed to remove from set: %w", err)
	}
	return nil
}

// SetMembers returns all members of a set
func (r *RedisCache) SetMembers(ctx context.Context, key string) ([]string, error) {
	members, err := r.Client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get set members: %w", err)
	}
	return members, nil
}

// SetIsMember checks if a member is in a set
func (r *RedisCache) SetIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	isMember, err := r.Client.SIsMember(ctx, key, member).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check set membership: %w", err)
	}
	return isMember, nil
}

// Increment increments a numeric value
func (r *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	result, err := r.Client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment value: %w", err)
	}
	return result, nil
}

// IncrementWithExpiration increments a numeric value and sets expiration
func (r *RedisCache) IncrementWithExpiration(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	pipe := r.Client.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, expiration)

	if _, err := pipe.Exec(ctx); err != nil {
		return 0, fmt.Errorf("failed to increment value with expiration: %w", err)
	}

	result, err := incrCmd.Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get increment result: %w", err)
	}

	return result, nil
}

// Clear removes all keys from the current database
func (r *RedisCache) Clear(ctx context.Context) error {
	if err := r.Client.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	r.Logger.Info("Cache cleared")
	return nil
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	if r.Client != nil {
		r.Logger.Info("Closing Redis connection")
		return r.Client.Close()
	}
	return nil
}

// HealthCheck performs a health check on Redis
func (r *RedisCache) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := r.Client.Ping(ctx).Err(); err != nil {
		r.Logger.WithError(err).Error("Redis health check failed")
		return fmt.Errorf("redis health check failed: %w", err)
	}

	r.Logger.Debug("Redis health check successful")
	return nil
}

// GetStats returns Redis connection pool statistics
func (r *RedisCache) GetStats() *redis.PoolStats {
	stats := r.Client.PoolStats()
	r.Logger.WithFields(logrus.Fields{
		"hits":         stats.Hits,
		"misses":       stats.Misses,
		"timeouts":     stats.Timeouts,
		"total_conns":  stats.TotalConns,
		"idle_conns":   stats.IdleConns,
		"stale_conns":  stats.StaleConns,
	}).Debug("Redis connection pool statistics")

	return stats
}