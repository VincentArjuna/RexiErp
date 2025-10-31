package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		wantFunc func(t *testing.T, cfg *Config)
		wantErr  bool
	}{
		{
			name: "default configuration",
			envVars: map[string]string{
				"APP_ENV": "development",
			},
			wantFunc: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "RexiERP", cfg.App.Name)
				assert.Equal(t, "1.0.0", cfg.App.Version)
				assert.Equal(t, "development", cfg.App.Environment)
				assert.Equal(t, true, cfg.App.Debug)
				assert.Equal(t, "0.0.0.0", cfg.App.Host)
				assert.Equal(t, 8000, cfg.App.Port)
				assert.Equal(t, "Asia/Jakarta", cfg.App.Timezone)
			},
			wantErr: false,
		},
		{
			name: "custom environment variables",
			envVars: map[string]string{
				"APP_NAME":        "TestApp",
				"APP_VERSION":     "2.0.0",
				"APP_ENV":         "production",
				"APP_DEBUG":       "false",
				"APP_HOST":        "127.0.0.1",
				"APP_PORT":        "9000",
				"TZ":              "UTC",
				"DB_HOST":         "test-host",
				"DB_PORT":         "5433",
				"DB_NAME":         "test_db",
				"DB_USER":         "test_user",
				"DB_PASSWORD":     "test_password",
				"JWT_SECRET":      "test-jwt-secret",
				"REDIS_HOST":      "test-redis",
				"REDIS_PORT":      "6380",
				"RABBITMQ_HOST":   "test-rabbitmq",
				"RABBITMQ_PORT":   "5673",
			},
			wantFunc: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "TestApp", cfg.App.Name)
				assert.Equal(t, "2.0.0", cfg.App.Version)
				assert.Equal(t, "production", cfg.App.Environment)
				assert.Equal(t, false, cfg.App.Debug)
				assert.Equal(t, "127.0.0.1", cfg.App.Host)
				assert.Equal(t, 9000, cfg.App.Port)
				assert.Equal(t, "UTC", cfg.App.Timezone)
				assert.Equal(t, "test-host", cfg.Database.Host)
				assert.Equal(t, 5433, cfg.Database.Port)
				assert.Equal(t, "test_db", cfg.Database.Name)
				assert.Equal(t, "test_user", cfg.Database.User)
				assert.Equal(t, "test_password", cfg.Database.Password)
				assert.Equal(t, "test-jwt-secret", cfg.JWT.Secret)
				assert.Equal(t, "test-redis", cfg.Redis.Host)
				assert.Equal(t, 6380, cfg.Redis.Port)
				assert.Equal(t, "test-rabbitmq", cfg.RabbitMQ.Host)
				assert.Equal(t, 5673, cfg.RabbitMQ.Port)
			},
			wantErr: false,
		},
		{
			name: "valid configuration - empty database host uses default",
			envVars: map[string]string{
				"DB_HOST": "",
			},
			wantFunc: func(t *testing.T, cfg *Config) {
				// Should use default value "localhost"
				assert.Equal(t, "localhost", cfg.Database.Host)
			},
			wantErr: false,
		},
		{
			name: "invalid configuration - production with default JWT secret",
			envVars: map[string]string{
				"APP_ENV":   "production",
				"JWT_SECRET": "your-super-secret-jwt-key-for-development-only",
			},
			wantErr: true,
		},
		{
			name: "invalid configuration - invalid port range",
			envVars: map[string]string{
				"APP_PORT": "70000",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				// Clean up environment variables
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			cfg, err := LoadConfig()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, cfg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, cfg)
				if tt.wantFunc != nil {
					tt.wantFunc(t, cfg)
				}
			}
		})
	}
}

func TestConfigGetDSN(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test_user",
			Password: "test_password",
			Name:     "test_db",
			SSLMode:  "disable",
		},
	}

	expected := "host=localhost port=5432 user=test_user password=test_password dbname=test_db sslmode=disable"
	assert.Equal(t, expected, cfg.GetDSN())
}

func TestConfigGetRedisAddr(t *testing.T) {
	cfg := &Config{
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
	}

	expected := "localhost:6379"
	assert.Equal(t, expected, cfg.GetRedisAddr())
}

func TestConfigGetRabbitMQURL(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *Config
		expected string
	}{
		{
			name: "default vhost",
			cfg: &Config{
				RabbitMQ: RabbitMQConfig{
					Host:     "localhost",
					Port:     5672,
					User:     "guest",
					Password: "guest",
					VHost:    "/",
				},
			},
			expected: "amqp://guest:guest@localhost:5672/",
		},
		{
			name: "custom vhost",
			cfg: &Config{
				RabbitMQ: RabbitMQConfig{
					Host:     "localhost",
					Port:     5672,
					User:     "test_user",
					Password: "test_password",
					VHost:    "/test",
				},
			},
			expected: "amqp://test_user:test_password@localhost:5672/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.cfg.GetRabbitMQURL())
		})
	}
}

func TestConfigEnvironmentChecks(t *testing.T) {
	tests := []struct {
		name       string
		environment string
		isDevelopment bool
		isProduction  bool
	}{
		{
			name:         "development environment",
			environment:  "development",
			isDevelopment: true,
			isProduction:  false,
		},
		{
			name:         "production environment",
			environment:  "production",
			isDevelopment: false,
			isProduction:  true,
		},
		{
			name:         "staging environment",
			environment:  "staging",
			isDevelopment: false,
			isProduction:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				App: AppConfig{
					Environment: tt.environment,
				},
			}

			assert.Equal(t, tt.isDevelopment, cfg.IsDevelopment())
			assert.Equal(t, tt.isProduction, cfg.IsProduction())
		})
	}
}

func TestConfigHelperFunctions(t *testing.T) {
	t.Run("getEnv with existing value", func(t *testing.T) {
		os.Setenv("TEST_VAR", "test_value")
		defer os.Unsetenv("TEST_VAR")

		result := getEnv("TEST_VAR", "default_value")
		assert.Equal(t, "test_value", result)
	})

	t.Run("getEnv with default", func(t *testing.T) {
		result := getEnv("NON_EXISTENT_VAR", "default_value")
		assert.Equal(t, "default_value", result)
	})

	t.Run("getEnvInt with valid value", func(t *testing.T) {
		os.Setenv("TEST_INT", "123")
		defer os.Unsetenv("TEST_INT")

		result := getEnvInt("TEST_INT", 456)
		assert.Equal(t, 123, result)
	})

	t.Run("getEnvInt with invalid value", func(t *testing.T) {
		os.Setenv("TEST_INT", "invalid")
		defer os.Unsetenv("TEST_INT")

		result := getEnvInt("TEST_INT", 456)
		assert.Equal(t, 456, result)
	})

	t.Run("getEnvBool with true value", func(t *testing.T) {
		os.Setenv("TEST_BOOL", "true")
		defer os.Unsetenv("TEST_BOOL")

		result := getEnvBool("TEST_BOOL", false)
		assert.True(t, result)
	})

	t.Run("getEnvBool with false value", func(t *testing.T) {
		os.Setenv("TEST_BOOL", "false")
		defer os.Unsetenv("TEST_BOOL")

		result := getEnvBool("TEST_BOOL", true)
		assert.False(t, result)
	})

	t.Run("getEnvDuration with valid value", func(t *testing.T) {
		os.Setenv("TEST_DURATION", "5m")
		defer os.Unsetenv("TEST_DURATION")

		result := getEnvDuration("TEST_DURATION", 10*time.Minute)
		assert.Equal(t, 5*time.Minute, result)
	})

	t.Run("getEnvDuration with invalid value", func(t *testing.T) {
		os.Setenv("TEST_DURATION", "invalid")
		defer os.Unsetenv("TEST_DURATION")

		result := getEnvDuration("TEST_DURATION", 10*time.Minute)
		assert.Equal(t, 10*time.Minute, result)
	})
}