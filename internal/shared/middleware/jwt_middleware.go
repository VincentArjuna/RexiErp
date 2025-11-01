package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/VincentArjuna/RexiErp/internal/authentication/service"
)

// JWTMiddleware provides JWT token validation middleware
type JWTMiddleware struct {
	authService service.AuthService
	logger      *logrus.Logger
}

// NewJWTMiddleware creates a new JWT middleware instance
func NewJWTMiddleware(authService service.AuthService, logger *logrus.Logger) *JWTMiddleware {
	return &JWTMiddleware{
		authService: authService,
		logger:      logger,
	}
}

// RequireAuth creates a gin middleware that requires valid JWT authentication
func (m *JWTMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			m.logger.Debug("Missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
				"code":  "MISSING_AUTH_HEADER",
			})
			c.Abort()
			return
		}

		// Parse Bearer token
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
			m.logger.Debug("Invalid Authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format. Expected 'Bearer <token>'",
				"code":  "INVALID_AUTH_FORMAT",
			})
			c.Abort()
			return
		}

		token := tokenParts[1]
		if token == "" {
			m.logger.Debug("Empty token provided")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token cannot be empty",
				"code":  "EMPTY_TOKEN",
			})
			c.Abort()
			return
		}

		// Validate token
		result, err := m.authService.ValidateToken(context.Background(), token)
		if err != nil {
			m.logger.WithFields(logrus.Fields{
				"error": err,
			}).Debug("Token validation failed")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
				"code":  "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		if !result.IsValid {
			m.logger.Debug("Token is invalid or expired")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token is invalid or expired",
				"code":  "TOKEN_EXPIRED",
			})
			c.Abort()
			return
		}

		// Add user information to context
		c.Set("user_id", result.UserID)
		c.Set("tenant_id", result.TenantID)
		c.Set("user_role", result.Role)
		c.Set("session_id", result.SessionID)

		// Add logging context
		c.Set("logger", m.logger.WithFields(logrus.Fields{
			"user_id":    result.UserID,
			"tenant_id":  result.TenantID,
			"user_role":  result.Role,
			"session_id": result.SessionID,
		}))

		m.logger.WithFields(logrus.Fields{
			"user_id":   result.UserID,
			"tenant_id": result.TenantID,
			"role":      result.Role,
		}).Debug("JWT validation successful")

		c.Next()
	}
}

// RequireRole creates a gin middleware that requires specific user roles
func (m *JWTMiddleware) RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure user is authenticated
		m.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		// Get user role from context
		userRole, exists := c.Get("user_role")
		if !exists {
			m.logger.Debug("User role not found in context")
			c.JSON(http.StatusForbidden, gin.H{
				"error": "User role not found",
				"code":  "ROLE_NOT_FOUND",
			})
			c.Abort()
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			m.logger.Debug("Invalid user role type in context")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"code":  "INVALID_ROLE_TYPE",
			})
			c.Abort()
			return
		}

		// Check if user role is allowed
		allowed := false
		for _, allowedRole := range allowedRoles {
			if roleStr == allowedRole {
				allowed = true
				break
			}
		}

		if !allowed {
			m.logger.WithFields(logrus.Fields{
				"user_role":     roleStr,
				"allowed_roles": allowedRoles,
			}).Debug("User role not allowed")
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
				"code":  "INSUFFICIENT_PERMISSIONS",
			})
			c.Abort()
			return
		}

		m.logger.WithFields(logrus.Fields{
			"user_role": roleStr,
		}).Debug("Role validation successful")

		c.Next()
	}
}

// RequireTenant creates a gin middleware that validates tenant context
func (m *JWTMiddleware) RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure user is authenticated
		m.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		// Get tenant ID from context
		tenantID, exists := c.Get("tenant_id")
		if !exists {
			m.logger.Debug("Tenant ID not found in context")
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Tenant context not found",
				"code":  "TENANT_NOT_FOUND",
			})
			c.Abort()
			return
		}

		_, ok := tenantID.(uuid.UUID)
		if !ok {
			m.logger.Debug("Invalid tenant ID type in context")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"code":  "INVALID_TENANT_TYPE",
			})
			c.Abort()
			return
		}

		// Add tenant ID to header for downstream services
		c.Header("X-Tenant-ID", tenantID.(uuid.UUID).String())

		m.logger.WithField("tenant_id", tenantID).Debug("Tenant validation successful")

		c.Next()
	}
}

// OptionalAuth creates a gin middleware that optionally validates JWT token
// If token is present and valid, it adds user info to context
// If token is missing or invalid, it continues without authentication
func (m *JWTMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		// Parse Bearer token
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
			// Invalid format, continue without authentication
			c.Next()
			return
		}

		token := tokenParts[1]
		if token == "" {
			// Empty token, continue without authentication
			c.Next()
			return
		}

		// Validate token
		result, err := m.authService.ValidateToken(context.Background(), token)
		if err != nil || !result.IsValid {
			// Invalid token, continue without authentication
			c.Next()
			return
		}

		// Add user information to context for valid tokens
		c.Set("user_id", result.UserID)
		c.Set("tenant_id", result.TenantID)
		c.Set("user_role", result.Role)
		c.Set("session_id", result.SessionID)

		// Add logging context
		c.Set("logger", m.logger.WithFields(logrus.Fields{
			"user_id":    result.UserID,
			"tenant_id":  result.TenantID,
			"user_role":  result.Role,
			"session_id": result.SessionID,
		}))

		m.logger.WithFields(logrus.Fields{
			"user_id":   result.UserID,
			"tenant_id": result.TenantID,
			"role":      result.Role,
		}).Debug("Optional JWT validation successful")

		c.Next()
	}
}