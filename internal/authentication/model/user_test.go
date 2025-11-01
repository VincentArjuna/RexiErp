package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestUser_IsSuperAdmin(t *testing.T) {
	tests := []struct {
		name     string
		role     UserRole
		expected bool
	}{
		{
			name:     "Super admin role",
			role:     RoleSuperAdmin,
			expected: true,
		},
		{
			name:     "Tenant admin role",
			role:     RoleTenantAdmin,
			expected: false,
		},
		{
			name:     "Staff role",
			role:     RoleStaff,
			expected: false,
		},
		{
			name:     "Viewer role",
			role:     RoleViewer,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.role}
			assert.Equal(t, tt.expected, user.IsSuperAdmin())
		})
	}
}

func TestUser_HasPermission(t *testing.T) {
	tests := []struct {
		name          string
		userRole      UserRole
		requiredRole  UserRole
		expected      bool
	}{
		{
			name:         "Super admin can access everything",
			userRole:     RoleSuperAdmin,
			requiredRole: RoleSuperAdmin,
			expected:     true,
		},
		{
			name:         "Super admin can access viewer resources",
			userRole:     RoleSuperAdmin,
			requiredRole: RoleViewer,
			expected:     true,
		},
		{
			name:         "Tenant admin can access staff resources",
			userRole:     RoleTenantAdmin,
			requiredRole: RoleStaff,
			expected:     true,
		},
		{
			name:         "Tenant admin cannot access super admin resources",
			userRole:     RoleTenantAdmin,
			requiredRole: RoleSuperAdmin,
			expected:     false,
		},
		{
			name:         "Staff can access viewer resources",
			userRole:     RoleStaff,
			requiredRole: RoleViewer,
			expected:     true,
		},
		{
			name:         "Staff cannot access tenant admin resources",
			userRole:     RoleStaff,
			requiredRole: RoleTenantAdmin,
			expected:     false,
		},
		{
			name:         "Viewer can only access viewer resources",
			userRole:     RoleViewer,
			requiredRole: RoleViewer,
			expected:     true,
		},
		{
			name:         "Viewer cannot access staff resources",
			userRole:     RoleViewer,
			requiredRole: RoleStaff,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.userRole}
			assert.Equal(t, tt.expected, user.HasPermission(tt.requiredRole))
		})
	}
}

func TestUser_UpdateLastLogin(t *testing.T) {
	user := &User{}

	// Initially nil
	assert.Nil(t, user.LastLogin)

	// Update last login
	user.UpdateLastLogin()

	// Should now be set and recent
	require.NotNil(t, user.LastLogin)
	assert.WithinDuration(t, time.Now(), *user.LastLogin, time.Second)
}

func TestUser_IsActiveUser(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		user      *User
		expected  bool
	}{
		{
			name: "Active user",
			user: &User{
				IsActive:  true,
				DeletedAt: gorm.DeletedAt{},
			},
			expected: true,
		},
		{
			name: "Inactive user",
			user: &User{
				IsActive:  false,
				DeletedAt: gorm.DeletedAt{},
			},
			expected: false,
		},
		{
			name: "Soft deleted user",
			user: &User{
				IsActive:  true,
				DeletedAt: gorm.DeletedAt{Time: now, Valid: true},
			},
			expected: false,
		},
		{
			name: "Inactive and soft deleted user",
			user: &User{
				IsActive:  false,
				DeletedAt: gorm.DeletedAt{Time: now, Valid: true},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.user.IsActiveUser())
		})
	}
}

func TestUser_SanitizeForResponse(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()

	user := &User{
		ID:           userID,
		TenantID:     tenantID,
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Test User",
		PhoneNumber:  "+6281234567890",
		Role:         RoleStaff,
		IsActive:     true,
	}

	sanitized := user.SanitizeForResponse()

	// Should contain non-sensitive fields
	assert.Equal(t, userID, sanitized.ID)
	assert.Equal(t, tenantID, sanitized.TenantID)
	assert.Equal(t, "test@example.com", sanitized.Email)
	assert.Equal(t, "Test User", sanitized.FullName)
	assert.Equal(t, "+6281234567890", sanitized.PhoneNumber)
	assert.Equal(t, RoleStaff, sanitized.Role)
	assert.True(t, sanitized.IsActive)

	// Should not contain sensitive fields
	assert.Empty(t, sanitized.PasswordHash)
}

func TestUserSession_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		expected  bool
	}{
		{
			name:      "Future expiration",
			expiresAt: time.Now().Add(time.Hour),
			expected:  false,
		},
		{
			name:      "Past expiration",
			expiresAt: time.Now().Add(-time.Hour),
			expected:  true,
		},
		{
			name:      "Current time",
			expiresAt: time.Now().Add(-time.Second),
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &UserSession{
				ExpiresAt: tt.expiresAt,
			}
			assert.Equal(t, tt.expected, session.IsExpired())
		})
	}
}

func TestUserSession_IsValid(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		session    *UserSession
		expected   bool
	}{
		{
			name: "Valid session",
			session: &UserSession{
				IsActive:     true,
				ExpiresAt:    now.Add(time.Hour),
				LastActivity: now,
			},
			expected: true,
		},
		{
			name: "Inactive session",
			session: &UserSession{
				IsActive:     false,
				ExpiresAt:    now.Add(time.Hour),
				LastActivity: now,
			},
			expected: false,
		},
		{
			name: "Expired session",
			session: &UserSession{
				IsActive:     true,
				ExpiresAt:    now.Add(-time.Hour),
				LastActivity: now,
			},
			expected: false,
		},
		{
			name: "Inactive for too long",
			session: &UserSession{
				IsActive:     true,
				ExpiresAt:    now.Add(time.Hour),
				LastActivity: now.Add(-25 * time.Hour),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.session.IsValid())
		})
	}
}

func TestUserSession_DeviceInfo(t *testing.T) {
	session := &UserSession{}

	deviceInfo := DeviceInfo{
		Platform:       "Web",
		Browser:        "Chrome",
		BrowserVersion: "120.0",
		OS:             "Windows",
		OSVersion:      "10",
		Device:         "Desktop",
		DeviceType:     "desktop",
		ScreenWidth:    1920,
		ScreenHeight:   1080,
		Language:       "en-US",
		Timezone:       "America/New_York",
	}

	// Test setting device info
	err := session.SetDeviceInfo(deviceInfo)
	require.NoError(t, err)
	assert.NotEmpty(t, session.DeviceInfo)

	// Test getting device info
	retrieved, err := session.GetDeviceInfo()
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, deviceInfo.Platform, retrieved.Platform)
	assert.Equal(t, deviceInfo.Browser, retrieved.Browser)
	assert.Equal(t, deviceInfo.BrowserVersion, retrieved.BrowserVersion)
	assert.Equal(t, deviceInfo.OS, retrieved.OS)
	assert.Equal(t, deviceInfo.OSVersion, retrieved.OSVersion)
	assert.Equal(t, deviceInfo.Device, retrieved.Device)
	assert.Equal(t, deviceInfo.DeviceType, retrieved.DeviceType)
	assert.Equal(t, deviceInfo.ScreenWidth, retrieved.ScreenWidth)
	assert.Equal(t, deviceInfo.ScreenHeight, retrieved.ScreenHeight)
	assert.Equal(t, deviceInfo.Language, retrieved.Language)
	assert.Equal(t, deviceInfo.Timezone, retrieved.Timezone)
}

func TestActivityLog_SetGetValues(t *testing.T) {
	log := &ActivityLog{}

	oldValues := map[string]interface{}{
		"email": "old@example.com",
		"role":  "viewer",
	}

	newValues := map[string]interface{}{
		"email": "new@example.com",
		"role":  "staff",
	}

	// Test setting old values
	err := log.SetOldValues(oldValues)
	require.NoError(t, err)

	// Test setting new values
	err = log.SetNewValues(newValues)
	require.NoError(t, err)

	// Test getting old values
	retrievedOld, err := log.GetOldValues()
	require.NoError(t, err)
	require.NotNil(t, retrievedOld)
	assert.Equal(t, "old@example.com", retrievedOld["email"])
	assert.Equal(t, "viewer", retrievedOld["role"])

	// Test getting new values
	retrievedNew, err := log.GetNewValues()
	require.NoError(t, err)
	require.NotNil(t, retrievedNew)
	assert.Equal(t, "new@example.com", retrievedNew["email"])
	assert.Equal(t, "staff", retrievedNew["role"])
}

func TestActivityLog_MarkAsFailed(t *testing.T) {
	log := &ActivityLog{
		Success: true,
	}

	assert.True(t, log.Success)
	assert.Empty(t, log.ErrorMessage)

	log.MarkAsFailed("Test error message")

	assert.False(t, log.Success)
	assert.Equal(t, "Test error message", log.ErrorMessage)
}

func TestActivityLog_Context(t *testing.T) {
	log := &ActivityLog{}

	context := ActivityContext{
		RequestID: "req-123",
		IPAddress: "192.168.1.1",
		UserAgent: "Test Agent",
		Metadata: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
		Component: "auth-service",
		TraceID:   "trace-123",
		SpanID:    "span-123",
	}

	// Test setting context
	err := log.SetContext(context)
	require.NoError(t, err)

	// Test getting context
	retrieved, err := log.GetContext()
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, context.RequestID, retrieved.RequestID)
	assert.Equal(t, context.IPAddress, retrieved.IPAddress)
	assert.Equal(t, context.UserAgent, retrieved.UserAgent)
	assert.Equal(t, context.Component, retrieved.Component)
	assert.Equal(t, context.TraceID, retrieved.TraceID)
	assert.Equal(t, context.SpanID, retrieved.SpanID)
	assert.Equal(t, "value1", retrieved.Metadata["key1"])
	assert.Equal(t, float64(42), retrieved.Metadata["key2"])
}