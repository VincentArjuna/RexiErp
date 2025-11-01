package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestAPIKeyMiddleware_RequireAPIKey_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Suppress logs in tests

	validKeys := []string{"test-key-1", "test-key-2"}
	middleware := NewAPIKeyMiddleware(validKeys, "X-API-Key", logger)

	router := gin.New()
	router.Use(middleware.RequireAPIKey())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test with valid API key in header
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "test-key-1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIKeyMiddleware_RequireAPIKey_MissingKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	validKeys := []string{"test-key-1"}
	middleware := NewAPIKeyMiddleware(validKeys, "X-API-Key", logger)

	router := gin.New()
	router.Use(middleware.RequireAPIKey())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPIKeyMiddleware_RequireAPIKey_InvalidKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	validKeys := []string{"test-key-1"}
	middleware := NewAPIKeyMiddleware(validKeys, "X-API-Key", logger)

	router := gin.New()
	router.Use(middleware.RequireAPIKey())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "invalid-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPIKeyMiddleware_RequireAPIKey_BearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	validKeys := []string{"test-key-1"}
	middleware := NewAPIKeyMiddleware(validKeys, "X-API-Key", logger)

	router := gin.New()
	router.Use(middleware.RequireAPIKey())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer test-key-1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIKeyMiddleware_OptionalAPIKey_WithKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	validKeys := []string{"test-key-1"}
	middleware := NewAPIKeyMiddleware(validKeys, "X-API-Key", logger)

	router := gin.New()
	router.Use(middleware.OptionalAPIKey())
	router.GET("/test", func(c *gin.Context) {
		authenticated, _ := c.Get("authenticated")
		c.JSON(http.StatusOK, gin.H{"authenticated": authenticated})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "test-key-1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIKeyMiddleware_OptionalAPIKey_WithoutKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	validKeys := []string{"test-key-1"}
	middleware := NewAPIKeyMiddleware(validKeys, "X-API-Key", logger)

	router := gin.New()
	router.Use(middleware.OptionalAPIKey())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHashAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"short key", "1234", "****"},
		{"normal key", "test-key-1234", "test*****1234"},
		{"long key", "very-long-api-key-123456", "very****************3456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hashAPIKey(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}