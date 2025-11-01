package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Permission represents a specific permission that can be granted to a role
type Permission struct {
	Resource string // e.g., "users", "orders", "products"
	Action   string // e.g., "read", "write", "delete", "admin"
}

// RolePermissions defines the permissions for each role
var RolePermissions = map[string][]Permission{
	"super_admin": {
		{Resource: "*", Action: "*"}, // Full access to everything
	},
	"tenant_admin": {
		{Resource: "users", Action: "read"},
		{Resource: "users", Action: "write"},
		{Resource: "users", Action: "delete"},
		{Resource: "orders", Action: "read"},
		{Resource: "orders", Action: "write"},
		{Resource: "orders", Action: "delete"},
		{Resource: "products", Action: "read"},
		{Resource: "products", Action: "write"},
		{Resource: "inventory", Action: "read"},
		{Resource: "inventory", Action: "write"},
		{Resource: "reports", Action: "read"},
		{Resource: "reports", Action: "write"},
		{Resource: "settings", Action: "read"},
		{Resource: "settings", Action: "write"},
	},
	"staff": {
		{Resource: "users", Action: "read"},
		{Resource: "orders", Action: "read"},
		{Resource: "orders", Action: "write"},
		{Resource: "products", Action: "read"},
		{Resource: "inventory", Action: "read"},
		{Resource: "inventory", Action: "write"},
		{Resource: "reports", Action: "read"},
	},
	"viewer": {
		{Resource: "users", Action: "read"},
		{Resource: "orders", Action: "read"},
		{Resource: "products", Action: "read"},
		{Resource: "inventory", Action: "read"},
		{Resource: "reports", Action: "read"},
	},
}

// RBACMiddleware provides role-based access control middleware
type RBACMiddleware struct {
	jwtMiddleware *JWTMiddleware
	logger        *logrus.Logger
}

// NewRBACMiddleware creates a new RBAC middleware instance
func NewRBACMiddleware(jwtMiddleware *JWTMiddleware, logger *logrus.Logger) *RBACMiddleware {
	return &RBACMiddleware{
		jwtMiddleware: jwtMiddleware,
		logger:        logger,
	}
}

// RequirePermission creates a gin middleware that requires specific permissions
func (m *RBACMiddleware) RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure user is authenticated
		m.jwtMiddleware.RequireAuth()(c)
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

		// Check if user has the required permission
		if !m.hasPermission(roleStr, resource, action) {
			m.logger.WithFields(logrus.Fields{
				"user_role": roleStr,
				"resource":  resource,
				"action":    action,
			}).Debug("User does not have required permission")
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
				"code":  "INSUFFICIENT_PERMISSIONS",
				"details": gin.H{
					"required_permission": resource + ":" + action,
					"user_role":          roleStr,
				},
			})
			c.Abort()
			return
		}

		m.logger.WithFields(logrus.Fields{
			"user_role": roleStr,
			"resource":  resource,
			"action":    action,
		}).Debug("RBAC validation successful")

		// Add permission info to context
		c.Set("resource", resource)
		c.Set("action", action)

		c.Next()
	}
}

// RequireRole creates a gin middleware that requires specific roles
func (m *RBACMiddleware) RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure user is authenticated
		m.jwtMiddleware.RequireAuth()(c)
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
				"details": gin.H{
					"allowed_roles": allowedRoles,
					"user_role":     roleStr,
				},
			})
			c.Abort()
			return
		}

		m.logger.WithField("user_role", roleStr).Debug("Role validation successful")

		c.Next()
	}
}

// TenantIsolation creates a middleware that ensures users can only access their own tenant's data
func (m *RBACMiddleware) TenantIsolation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure user is authenticated
		m.jwtMiddleware.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		// Get tenant ID from context (set by JWT middleware)
		userTenantID, exists := c.Get("tenant_id")
		if !exists {
			m.logger.Debug("Tenant ID not found in context")
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Tenant context not found",
				"code":  "TENANT_NOT_FOUND",
			})
			c.Abort()
			return
		}

		// Check for tenant ID in request (for APIs that allow cross-tenant access)
		// This is typically used by super_admins
		requestTenantID := c.Param("tenant_id")
		if requestTenantID == "" {
			requestTenantID = c.Query("tenant_id")
		}
		if requestTenantID == "" {
			requestTenantID = c.GetHeader("X-Tenant-ID")
		}

		// Get user role to check if they can access other tenants
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

		roleStr, _ := userRole.(string)

		// If request specifies a different tenant, check if user is allowed
		if requestTenantID != "" && requestTenantID != userTenantID.(string) {
			if roleStr != "super_admin" {
				m.logger.WithFields(logrus.Fields{
					"user_role":        roleStr,
					"user_tenant_id":   userTenantID,
					"request_tenant_id": requestTenantID,
				}).Debug("User attempting to access different tenant")
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Cross-tenant access not allowed",
					"code":  "CROSS_TENANT_ACCESS_DENIED",
				})
				c.Abort()
				return
			}
		}

		// Set the effective tenant ID for this request
		effectiveTenantID := userTenantID
		if requestTenantID != "" && roleStr == "super_admin" {
			effectiveTenantID = requestTenantID
		}

		// Add tenant context for downstream services
		c.Set("effective_tenant_id", effectiveTenantID)
		c.Header("X-Tenant-ID", effectiveTenantID.(string))

		m.logger.WithFields(logrus.Fields{
			"user_role":           roleStr,
			"user_tenant_id":      userTenantID,
			"request_tenant_id":   requestTenantID,
			"effective_tenant_id": effectiveTenantID,
		}).Debug("Tenant isolation validation successful")

		c.Next()
	}
}

// hasPermission checks if a role has the required permission
func (m *RBACMiddleware) hasPermission(role, resource, action string) bool {
	permissions, exists := RolePermissions[role]
	if !exists {
		return false
	}

	for _, permission := range permissions {
		// Wildcard permissions
		if permission.Resource == "*" && permission.Action == "*" {
			return true
		}
		if permission.Resource == "*" && permission.Action == action {
			return true
		}
		if permission.Resource == resource && permission.Action == "*" {
			return true
		}

		// Exact match
		if permission.Resource == resource && permission.Action == action {
			return true
		}
	}

	return false
}

// GetRolePermissions returns all permissions for a given role
func GetRolePermissions(role string) []Permission {
	return RolePermissions[role]
}

// AddRolePermission adds a permission to a role (for dynamic configuration)
func AddRolePermission(role string, permission Permission) {
	if RolePermissions[role] == nil {
		RolePermissions[role] = []Permission{}
	}
	RolePermissions[role] = append(RolePermissions[role], permission)
}

// ResourceFromPath extracts resource type from HTTP path
func ResourceFromPath(path string) string {
	// Remove common prefixes
	path = strings.TrimPrefix(path, "/api/v1/")
	path = strings.TrimPrefix(path, "/api/")

	// Extract first segment as resource
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) > 0 {
		return parts[0]
	}

	return "unknown"
}

// ActionFromMethod extracts action type from HTTP method
func ActionFromMethod(method string) string {
	switch strings.ToUpper(method) {
	case "GET":
		return "read"
	case "POST", "PUT", "PATCH":
		return "write"
	case "DELETE":
		return "delete"
	default:
		return "unknown"
	}
}