package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Config represents the application configuration
type Config struct {
	App      AppConfig      `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq"`
	JWT      JWTConfig      `yaml:"jwt"`
	APIKey   APIKeyConfig   `yaml:"api_key"`
	Log      LogConfig      `yaml:"log"`
	Monitoring MonitoringConfig `yaml:"monitoring"`
}

// AppConfig represents application-specific configuration
type AppConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
	Debug       bool   `yaml:"debug"`
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Timezone    string `yaml:"timezone"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	Name            string        `yaml:"name"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	SSLMode         string        `yaml:"ssl_mode"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

// RabbitMQConfig represents RabbitMQ configuration
type RabbitMQConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	VHost        string `yaml:"vhost"`
	MaxRetries   int    `yaml:"max_retries"`
	RetryDelay   string `yaml:"retry_delay"`
	Heartbeat    time.Duration `yaml:"heartbeat"`
	ConnectionTimeout time.Duration `yaml:"connection_timeout"`
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	Secret           string        `yaml:"secret"`
	ExpirationHours  int           `yaml:"expiration_hours"`
	Issuer           string        `yaml:"issuer"`
	RefreshTokenDays int           `yaml:"refresh_token_days"`
	AccessTokenTTL   time.Duration `yaml:"access_token_ttl"`
}

// APIKeyConfig represents API key configuration
type APIKeyConfig struct {
	Enabled    bool     `yaml:"enabled"`
	Keys       []string `yaml:"keys"`
	HeaderName string   `yaml:"header_name"`
}

// Validate validates the database configuration
func (c *DatabaseConfig) Validate() error {
	// Validate connection pool settings
	if c.MaxOpenConns <= 0 {
		return fmt.Errorf("max open connections must be greater than 0")
	}
	if c.MaxOpenConns > 1000 {
		return fmt.Errorf("max open connections (%d) is too high, maximum recommended is 1000", c.MaxOpenConns)
	}

	if c.MaxIdleConns < 0 {
		return fmt.Errorf("max idle connections cannot be negative")
	}
	if c.MaxIdleConns > c.MaxOpenConns {
		return fmt.Errorf("max idle connections (%d) cannot be greater than max open connections (%d)", c.MaxIdleConns, c.MaxOpenConns)
	}

	// Validate connection lifetime settings
	if c.ConnMaxLifetime <= 0 {
		return fmt.Errorf("connection max lifetime must be greater than 0")
	}
	if c.ConnMaxLifetime > 24*time.Hour {
		return fmt.Errorf("connection max lifetime (%v) is too long, maximum recommended is 24 hours", c.ConnMaxLifetime)
	}

	if c.ConnMaxIdleTime <= 0 {
		return fmt.Errorf("connection max idle time must be greater than 0")
	}
	if c.ConnMaxIdleTime > c.ConnMaxLifetime {
		return fmt.Errorf("connection max idle time (%v) cannot be greater than connection max lifetime (%v)", c.ConnMaxIdleTime, c.ConnMaxLifetime)
	}

	return nil
}

// LogConfig represents logging configuration
type LogConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	Output     string `yaml:"output"`
	MaxSize    int    `yaml:"max_size"`
	MaxAge     int    `yaml:"max_age"`
	MaxBackups int    `yaml:"max_backups"`
	Compress   bool   `yaml:"compress"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Port       int    `yaml:"port"`
	Path       string `yaml:"path"`
	MetricsURL string `yaml:"metrics_url"`
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() (*Config, error) {
	config := &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "RexiERP"),
			Version:     getEnv("APP_VERSION", "1.0.0"),
			Environment: getEnv("APP_ENV", "development"),
			Debug:       getEnvBool("APP_DEBUG", true),
			Host:        getEnv("APP_HOST", "0.0.0.0"),
			Port:        getEnvInt("APP_PORT", 8000),
			Timezone:    getEnv("TZ", "Asia/Jakarta"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvInt("DB_PORT", 5432),
			Name:            getEnv("DB_NAME", "rexi_erp"),
			User:            getEnv("DB_USER", "rexi"),
			Password:        getEnv("DB_PASSWORD", "password"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNECTIONS", 50),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNECTIONS", 10),
			ConnMaxLifetime: getEnvDuration("DB_CONNECTION_MAX_LIFETIME", 1*time.Hour),
			ConnMaxIdleTime: getEnvDuration("DB_CONNECTION_MAX_IDLE_TIME", 30*time.Minute),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
			PoolSize: getEnvInt("REDIS_POOL_SIZE", 10),
		},
		RabbitMQ: RabbitMQConfig{
			Host:               getEnv("RABBITMQ_HOST", "localhost"),
			Port:               getEnvInt("RABBITMQ_PORT", 5672),
			User:               getEnv("RABBITMQ_USER", "guest"),
			Password:           getEnv("RABBITMQ_PASSWORD", "guest"),
			VHost:              getEnv("RABBITMQ_VHOST", "/"),
			MaxRetries:         getEnvInt("MESSAGE_QUEUE_RETRY_ATTEMPTS", 3),
			RetryDelay:         getEnv("MESSAGE_QUEUE_RETRY_DELAY", "5s"),
			Heartbeat:          getEnvDuration("RABBITMQ_HEARTBEAT", 10*time.Second),
			ConnectionTimeout:  getEnvDuration("RABBITMQ_CONNECTION_TIMEOUT", 30*time.Second),
		},
		JWT: JWTConfig{
			Secret:           getEnv("JWT_SECRET", "your-super-secret-jwt-key-for-development-only"),
			ExpirationHours:  getEnvInt("JWT_EXPIRATION_HOURS", 24),
			Issuer:           getEnv("JWT_ISSUER", "RexiERP"),
			RefreshTokenDays: getEnvInt("JWT_REFRESH_TOKEN_DAYS", 7),
			AccessTokenTTL:   getEnvDuration("JWT_ACCESS_TOKEN_TTL", 24*time.Hour),
		},
		APIKey: APIKeyConfig{
			Enabled:    getEnvBool("API_KEY_AUTH_ENABLED", true),
			Keys:       getEnvSlice("API_KEYS", []string{"rexierp-api-key-2024-dev"}),
			HeaderName: getEnv("API_KEY_HEADER", "X-API-Key"),
		},
		Log: LogConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			Output:     getEnv("LOG_OUTPUT", "stdout"),
			MaxSize:    getEnvInt("LOG_MAX_SIZE", 100),
			MaxAge:     getEnvInt("LOG_MAX_AGE", 7),
			MaxBackups: getEnvInt("LOG_MAX_BACKUPS", 3),
			Compress:   getEnvBool("LOG_COMPRESS", true),
		},
		Monitoring: MonitoringConfig{
			Enabled:    getEnvBool("ENABLE_METRICS", true),
			Port:       getEnvInt("METRICS_PORT", 9000),
			Path:       getEnv("METRICS_PATH", "/metrics"),
			MetricsURL: getEnv("METRICS_URL", ""),
		},
	}

	// Validate configuration
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// validate validates the configuration
func (c *Config) validate() error {
	// Validate required fields
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}
	if c.JWT.Secret == "your-super-secret-jwt-key-for-development-only" && c.App.Environment == "production" {
		return fmt.Errorf("JWT secret must be changed in production")
	}

	// Validate port ranges
	if c.App.Port < 1 || c.App.Port > 65535 {
		return fmt.Errorf("invalid application port: %d", c.App.Port)
	}
	if c.Database.Port < 1 || c.Database.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", c.Database.Port)
	}
	if c.Redis.Port < 1 || c.Redis.Port > 65535 {
		return fmt.Errorf("invalid redis port: %d", c.Redis.Port)
	}
	if c.RabbitMQ.Port < 1 || c.RabbitMQ.Port > 65535 {
		return fmt.Errorf("invalid rabbitmq port: %d", c.RabbitMQ.Port)
	}

	// Validate database connection settings
	if err := c.Database.Validate(); err != nil {
		return fmt.Errorf("database configuration validation failed: %w", err)
	}

	return nil
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// GetRedisAddr returns the Redis connection address
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// GetRabbitMQURL returns the RabbitMQ connection URL
func (c *Config) GetRabbitMQURL() string {
	if c.RabbitMQ.VHost == "/" {
		return fmt.Sprintf("amqp://%s:%s@%s:%d/",
			c.RabbitMQ.User,
			c.RabbitMQ.Password,
			c.RabbitMQ.Host,
			c.RabbitMQ.Port,
		)
	}
	return fmt.Sprintf("amqp://%s:%s@%s:%d%s",
		c.RabbitMQ.User,
		c.RabbitMQ.Password,
		c.RabbitMQ.Host,
		c.RabbitMQ.Port,
		c.RabbitMQ.VHost,
	)
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// GetLogLevel returns the logrus log level
func (c *Config) GetLogLevel() logrus.Level {
	level, err := logrus.ParseLevel(c.Log.Level)
	if err != nil {
		return logrus.InfoLevel
	}
	return level
}

// Helper functions for environment variable parsing

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

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}