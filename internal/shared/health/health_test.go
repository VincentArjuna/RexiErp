package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/VincentArjuna/RexiErp/internal/shared/config"
)

func TestHealthChecker_BasicHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cfg := &config.Config{
		App: config.AppConfig{
			Version:     "1.0.0",
			Environment: "test",
		},
	}

	healthChecker := NewHealthChecker(nil, cfg, logger)
	router := gin.New()
	healthChecker.RegisterRoutes(router)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
}

func TestHealthChecker_Liveness(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cfg := &config.Config{
		App: config.AppConfig{
			Version:     "1.0.0",
			Environment: "test",
		},
	}

	healthChecker := NewHealthChecker(nil, cfg, logger)
	router := gin.New()
	healthChecker.RegisterRoutes(router)

	req := httptest.NewRequest("GET", "/health/live", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.Contains(t, response, "uptime")
}

func TestHealthChecker_Readiness(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cfg := &config.Config{
		App: config.AppConfig{
			Version:     "1.0.0",
			Environment: "test",
		},
		Database: config.DatabaseConfig{
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 1 * time.Hour,
			ConnMaxIdleTime: 30 * time.Minute,
		},
	}

	healthChecker := NewHealthChecker(nil, cfg, logger)
	router := gin.New()
	healthChecker.RegisterRoutes(router)

	req := httptest.NewRequest("GET", "/health/ready", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, StatusUnhealthy, response.Status)
	assert.Contains(t, response.Checks, "database")
	assert.Equal(t, StatusUnhealthy, response.Checks["database"].Status)
}

func TestHealthChecker_DetailedHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cfg := &config.Config{
		App: config.AppConfig{
			Version:     "1.0.0",
			Environment: "test",
		},
		JWT: config.JWTConfig{
			Secret: "test-secret",
		},
		APIKey: config.APIKeyConfig{
			Keys: []string{"test-key"},
		},
		Database: config.DatabaseConfig{
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 1 * time.Hour,
			ConnMaxIdleTime: 30 * time.Minute,
		},
	}

	healthChecker := NewHealthChecker(nil, cfg, logger)
	router := gin.New()
	healthChecker.RegisterRoutes(router)

	req := httptest.NewRequest("GET", "/health/detailed", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, StatusDegraded, response.Status) // Database is unhealthy, but others are healthy
	assert.Equal(t, "1.0.0", response.Version)
	assert.Equal(t, "test", response.Environment)
	assert.Contains(t, response.Checks, "database")
	assert.Contains(t, response.Checks, "configuration")
	assert.Contains(t, response.Checks, "system")
	assert.Contains(t, response.System.GoVersion, "go")
}

func TestHealthChecker_ConfigurationCheck(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("ValidConfiguration", func(t *testing.T) {
		cfg := &config.Config{
			JWT: config.JWTConfig{
				Secret: "valid-secret",
			},
			APIKey: config.APIKeyConfig{
				Keys: []string{"valid-key"},
			},
		}

		healthChecker := NewHealthChecker(nil, cfg, logger)
		result := healthChecker.checkConfiguration()

		assert.Equal(t, StatusHealthy, result.Status)
		assert.Equal(t, "Configuration valid", result.Message)
	})

	t.Run("MissingJWTSecret", func(t *testing.T) {
		cfg := &config.Config{
			JWT: config.JWTConfig{
				Secret: "",
			},
			APIKey: config.APIKeyConfig{
				Keys: []string{"valid-key"},
			},
		}

		healthChecker := NewHealthChecker(nil, cfg, logger)
		result := healthChecker.checkConfiguration()

		assert.Equal(t, StatusUnhealthy, result.Status)
		assert.Contains(t, result.Message, "JWT secret not configured")
	})

	t.Run("DefaultProductionSecret", func(t *testing.T) {
		cfg := &config.Config{
			App: config.AppConfig{
				Environment: "production",
			},
			JWT: config.JWTConfig{
				Secret: "your-super-secret-jwt-key-for-development-only",
			},
			APIKey: config.APIKeyConfig{
				Keys: []string{"valid-key"},
			},
		}

		healthChecker := NewHealthChecker(nil, cfg, logger)
		result := healthChecker.checkConfiguration()

		assert.Equal(t, StatusDegraded, result.Status)
		assert.Contains(t, result.Message, "Using default JWT secret in production")
	})

	t.Run("NoAPIKeys", func(t *testing.T) {
		cfg := &config.Config{
			JWT: config.JWTConfig{
				Secret: "valid-secret",
			},
			APIKey: config.APIKeyConfig{
				Keys: []string{},
			},
		}

		healthChecker := NewHealthChecker(nil, cfg, logger)
		result := healthChecker.checkConfiguration()

		assert.Equal(t, StatusDegraded, result.Status)
		assert.Contains(t, result.Message, "No API keys configured")
	})
}

func TestHealthChecker_SystemResourcesCheck(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			MaxOpenConns: 10,
		},
	}

	healthChecker := NewHealthChecker(nil, cfg, logger)
	result := healthChecker.checkSystemResources()

	// System resources should typically be healthy in test environment
	assert.Equal(t, StatusHealthy, result.Status)
	assert.Contains(t, result.Message, "System resources healthy")
	assert.Contains(t, result.Details, "alloc_mb")
	assert.Contains(t, result.Details, "num_goroutine")
}

func TestHealthChecker_GetSystemInfo(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cfg := &config.Config{}
	healthChecker := NewHealthChecker(nil, cfg, logger)

	systemInfo := healthChecker.getSystemInfo()

	assert.NotEmpty(t, systemInfo.GoVersion)
	assert.Greater(t, systemInfo.NumGoroutine, 0)
	assert.GreaterOrEqual(t, systemInfo.MemoryUsage.Alloc, uint64(0))
	assert.GreaterOrEqual(t, systemInfo.MemoryUsage.TotalAlloc, uint64(0))
	assert.GreaterOrEqual(t, systemInfo.MemoryUsage.Sys, uint64(0))
}

func TestSetupDefaultHealthRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cfg := &config.Config{
		App: config.AppConfig{
			Version:     "1.0.0",
			Environment: "test",
		},
	}

	router := gin.New()
	SetupDefaultHealthRoutes(router, nil, cfg, logger)

	// Test that routes are registered
	routes := router.Routes()

	routePaths := make(map[string]bool)
	for _, route := range routes {
		routePaths[route.Path] = true
	}

	assert.True(t, routePaths["/health"])
	assert.True(t, routePaths["/health/live"])
	assert.True(t, routePaths["/health/ready"])
	assert.True(t, routePaths["/health/detailed"])
}

func TestBToMb(t *testing.T) {
	testCases := []struct {
		name     string
		bytes    uint64
		expected uint64
	}{
		{"Zero", 0, 0},
		{"OneMB", 1024 * 1024, 1},
		{"FiveMB", 5 * 1024 * 1024, 5},
		{"PartialMB", 1024 * 1024 + 512 * 1024, 1}, // Should floor
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := bToMb(tc.bytes)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestHealthStatus_String(t *testing.T) {
	assert.Equal(t, "healthy", string(StatusHealthy))
	assert.Equal(t, "unhealthy", string(StatusUnhealthy))
	assert.Equal(t, "degraded", string(StatusDegraded))
}