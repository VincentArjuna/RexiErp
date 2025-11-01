package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole represents the user role enum
type UserRole string

const (
	RoleSuperAdmin UserRole = "super_admin"
	RoleTenantAdmin UserRole = "tenant_admin"
	RoleStaff      UserRole = "staff"
	RoleViewer     UserRole = "viewer"
)

// User represents a user in the system
type User struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID    uuid.UUID  `gorm:"type:uuid;not null;index:idx_user_tenant" json:"tenant_id"`
	Email       string     `gorm:"type:varchar(255);not null;uniqueIndex:idx_user_email_tenant" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	FullName    string     `gorm:"type:varchar(255);not null" json:"full_name"`
	PhoneNumber string     `gorm:"type:varchar(20)" json:"phone_number"`
	Role        UserRole   `gorm:"type:varchar(20);not null;default:'viewer'" json:"role"`
	IsActive    bool       `gorm:"not null;default:true" json:"is_active"`
	LastLogin   *time.Time `gorm:"type:timestamp" json:"last_login"`
	CreatedAt   time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"not null" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	UserSessions []UserSession `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	ActivityLogs []ActivityLog `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName returns the table name for the User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate is a GORM hook that runs before creating a user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// IsSuperAdmin checks if the user is a super admin
func (u *User) IsSuperAdmin() bool {
	return u.Role == RoleSuperAdmin
}

// IsTenantAdmin checks if the user is a tenant admin
func (u *User) IsTenantAdmin() bool {
	return u.Role == RoleTenantAdmin
}

// IsStaff checks if the user is staff
func (u *User) IsStaff() bool {
	return u.Role == RoleStaff
}

// IsViewer checks if the user is a viewer
func (u *User) IsViewer() bool {
	return u.Role == RoleViewer
}

// HasPermission checks if the user has the required role or higher
func (u *User) HasPermission(requiredRole UserRole) bool {
	roleHierarchy := map[UserRole]int{
		RoleViewer:     1,
		RoleStaff:      2,
		RoleTenantAdmin: 3,
		RoleSuperAdmin: 4,
	}

	userLevel := roleHierarchy[u.Role]
	requiredLevel := roleHierarchy[requiredRole]

	return userLevel >= requiredLevel
}

// SanitizeForResponse returns a user object with sensitive data removed
func (u *User) SanitizeForResponse() *User {
	return &User{
		ID:          u.ID,
		TenantID:    u.TenantID,
		Email:       u.Email,
		FullName:    u.FullName,
		PhoneNumber: u.PhoneNumber,
		Role:        u.Role,
		IsActive:    u.IsActive,
		LastLogin:   u.LastLogin,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// UpdateLastLogin updates the last login timestamp
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLogin = &now
}

// IsActiveUser checks if the user account is active
func (u *User) IsActiveUser() bool {
	return u.IsActive && u.DeletedAt.Time.IsZero()
}