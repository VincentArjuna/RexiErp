package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// APIKeyMiddleware provides API key authentication middleware
type APIKeyMiddleware struct {
	validKeys  map[string]bool
	headerName string
	logger     *logrus.Logger
}

// NewAPIKeyMiddleware creates a new API key authentication middleware
func NewAPIKeyMiddleware(apiKeys []string, headerName string, logger *logrus.Logger) *APIKeyMiddleware {
	validKeys := make(map[string]bool)
	for _, key := range apiKeys {
		validKeys[key] = true
	}

	return &APIKeyMiddleware{
		validKeys:  validKeys,
		headerName: headerName,
		logger:     logger,
	}
}

// RequireAPIKey returns a Gin middleware that requires valid API key
func (m *APIKeyMiddleware) RequireAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get API key from header
		apiKey := c.GetHeader(m.headerName)
		if apiKey == "" {
			// Also try Authorization header with Bearer token
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// Validate API key
		if apiKey == "" {
			m.logger.WithFields(logrus.Fields{
				"ip":     c.ClientIP(),
				"path":   c.Request.URL.Path,
				"method": c.Request.Method,
			}).Warn("API key authentication failed: missing key")

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "API key required",
				},
			})
			c.Abort()
			return
		}

		if !m.validKeys[apiKey] {
			m.logger.WithFields(logrus.Fields{
				"ip":       c.ClientIP(),
				"path":     c.Request.URL.Path,
				"method":   c.Request.Method,
				"key_hash": hashAPIKey(apiKey),
			}).Warn("API key authentication failed: invalid key")

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Invalid API key",
				},
			})
			c.Abort()
			return
		}

		// Log successful authentication
		m.logger.WithFields(logrus.Fields{
			"ip":       c.ClientIP(),
			"path":     c.Request.URL.Path,
			"method":   c.Request.Method,
			"key_hash": hashAPIKey(apiKey),
		}).Debug("API key authentication successful")

		// Add API key hash to context for potential audit logging
		c.Set("api_key_hash", hashAPIKey(apiKey))
		c.Next()
	}
}

// OptionalAPIKey returns a Gin middleware that optionally validates API key
// If provided, validates the key, but doesn't require it
func (m *APIKeyMiddleware) OptionalAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader(m.headerName)
		if apiKey == "" {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if apiKey != "" {
			if m.validKeys[apiKey] {
				c.Set("api_key_hash", hashAPIKey(apiKey))
				c.Set("authenticated", true)
			} else {
				m.logger.WithFields(logrus.Fields{
					"ip":       c.ClientIP(),
					"path":     c.Request.URL.Path,
					"method":   c.Request.Method,
					"key_hash": hashAPIKey(apiKey),
				}).Warn("API key authentication failed: invalid key")

				c.JSON(http.StatusUnauthorized, gin.H{
					"error": gin.H{
						"code":    "UNAUTHORIZED",
						"message": "Invalid API key",
					},
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// hashAPIKey creates a hash of the API key for logging purposes
func hashAPIKey(key string) string {
	if len(key) <= 8 {
		return strings.Repeat("*", len(key))
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}