package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PasswordResetToken represents a password reset token for users
type PasswordResetToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	TenantID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Token     string     `gorm:"type:varchar(255);not null;uniqueIndex" json:"-"`
	TokenHash string     `gorm:"type:varchar(255);not null;uniqueIndex" json:"token_hash"`
	Email     string     `gorm:"type:varchar(255);not null" json:"email"`
	ExpiresAt time.Time  `gorm:"not null;index" json:"expires_at"`
	UsedAt    *time.Time `gorm:"index" json:"used_at,omitempty"`
	IPAddress string     `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent string     `gorm:"type:text" json:"user_agent"`
	IsActive  bool       `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

// TableName specifies the table name for PasswordResetToken model
func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}

// IsExpired checks if the token has expired
func (t *PasswordResetToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsUsed checks if the token has been used
func (t *PasswordResetToken) IsUsed() bool {
	return t.UsedAt != nil
}

// IsValid checks if the token is valid (not expired, not used, and active)
func (t *PasswordResetToken) IsValid() bool {
	return t.IsActive && !t.IsExpired() && !t.IsUsed()
}

// MarkAsUsed marks the token as used
func (t *PasswordResetToken) MarkAsUsed() {
	now := time.Now()
	t.UsedAt = &now
	t.IsActive = false
}

// Deactivate deactivates the token
func (t *PasswordResetToken) Deactivate() {
	t.IsActive = false
}

// SanitizeForResponse returns a sanitized version of the token for API responses
func (t *PasswordResetToken) SanitizeForResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":         t.ID,
		"user_id":    t.UserID,
		"tenant_id":  t.TenantID,
		"email":      t.Email,
		"expires_at": t.ExpiresAt,
		"used_at":    t.UsedAt,
		"is_active":  t.IsActive,
		"created_at": t.CreatedAt,
	}
}

// TokenMetadata represents metadata about the token for audit purposes
type TokenMetadata struct {
	TokenID   uuid.UUID `json:"token_id"`
	UserID    uuid.UUID `json:"user_id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Email     string    `json:"email"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

// GetMetadata returns the token metadata for logging
func (t *PasswordResetToken) GetMetadata() TokenMetadata {
	return TokenMetadata{
		TokenID:   t.ID,
		UserID:    t.UserID,
		TenantID:  t.TenantID,
		Email:     t.Email,
		IPAddress: t.IPAddress,
		UserAgent: t.UserAgent,
		CreatedAt: t.CreatedAt,
	}
}

// BeforeCreate hook to set default values before creating a new token
func (t *PasswordResetToken) BeforeCreate(tx *gorm.DB) error {
	if t.IsActive && t.ExpiresAt.IsZero() {
		// Set default expiration to 1 hour if not specified
		t.ExpiresAt = time.Now().Add(1 * time.Hour)
	}
	return nil
}

// SetOldValues sets the old values for activity logging
func (t *PasswordResetToken) SetOldValues(oldValues map[string]interface{}) error {
	data, err := json.Marshal(oldValues)
	if err != nil {
		return err
	}
	// In a real implementation, you might have a field to store this
	// For now, this is a placeholder for the interface
	_ = data
	return nil
}

// SetNewValues sets the new values for activity logging
func (t *PasswordResetToken) SetNewValues(newValues map[string]interface{}) error {
	data, err := json.Marshal(newValues)
	if err != nil {
		return err
	}
	// In a real implementation, you might have a field to store this
	// For now, this is a placeholder for the interface
	_ = data
	return nil
}