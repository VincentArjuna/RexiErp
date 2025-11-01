package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserSession represents a user session for token management
type UserSession struct {
	ID                uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID            uuid.UUID  `gorm:"type:uuid;not null;index:idx_session_user" json:"user_id"`
	TenantID          uuid.UUID  `gorm:"type:uuid;not null;index:idx_session_tenant" json:"tenant_id"`
	SessionID         string     `gorm:"type:varchar(255);not null;uniqueIndex" json:"session_id"`
	TokenHash         string     `gorm:"type:varchar(255);not null" json:"-"`
	RefreshTokenHash  string     `gorm:"type:varchar(255);not null" json:"-"`
	DeviceInfo        string     `gorm:"type:json" json:"device_info"`
	IPAddress         string     `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent         string     `gorm:"type:text" json:"user_agent"`
	ExpiresAt         time.Time  `gorm:"not null;index" json:"expires_at"`
	LastActivity      time.Time  `gorm:"not null" json:"last_activity"`
	IsActive          bool       `gorm:"not null;default:true" json:"is_active"`
	CreatedAt         time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"not null" json:"updated_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// DeviceInfo represents device fingerprinting data
type DeviceInfo struct {
	Platform       string `json:"platform"`
	Browser        string `json:"browser"`
	BrowserVersion string `json:"browser_version"`
	OS             string `json:"os"`
	OSVersion      string `json:"os_version"`
	Device         string `json:"device"`
	DeviceType     string `json:"device_type"`
	ScreenWidth    int    `json:"screen_width"`
	ScreenHeight   int    `json:"screen_height"`
	Language       string `json:"language"`
	Timezone       string `json:"timezone"`
}

// TableName returns the table name for the UserSession model
func (UserSession) TableName() string {
	return "user_sessions"
}

// BeforeCreate is a GORM hook that runs before creating a session
func (us *UserSession) BeforeCreate(tx *gorm.DB) error {
	if us.ID == uuid.Nil {
		us.ID = uuid.New()
	}
	if us.SessionID == "" {
		us.SessionID = uuid.New().String()
	}
	return nil
}

// SetDeviceInfo sets device information from JSON string
func (us *UserSession) SetDeviceInfo(deviceInfo DeviceInfo) error {
	jsonData, err := json.Marshal(deviceInfo)
	if err != nil {
		return err
	}
	us.DeviceInfo = string(jsonData)
	return nil
}

// GetDeviceInfo returns device information as struct
func (us *UserSession) GetDeviceInfo() (*DeviceInfo, error) {
	if us.DeviceInfo == "" {
		return nil, nil
	}

	var deviceInfo DeviceInfo
	err := json.Unmarshal([]byte(us.DeviceInfo), &deviceInfo)
	if err != nil {
		return nil, err
	}
	return &deviceInfo, nil
}

// IsExpired checks if the session is expired
func (us *UserSession) IsExpired() bool {
	return time.Now().After(us.ExpiresAt)
}

// IsInactive checks if the session is inactive
func (us *UserSession) IsInactive() bool {
	return !us.IsActive || time.Since(us.LastActivity) > 24*time.Hour
}

// UpdateActivity updates the last activity timestamp
func (us *UserSession) UpdateActivity() {
	us.LastActivity = time.Now()
}

// IsValid checks if the session is valid (active and not expired)
func (us *UserSession) IsValid() bool {
	return us.IsActive && !us.IsExpired() && !us.IsInactive()
}

// SanitizeForResponse returns a session object with sensitive data removed
func (us *UserSession) SanitizeForResponse() *UserSession {
	return &UserSession{
		ID:           us.ID,
		UserID:       us.UserID,
		TenantID:     us.TenantID,
		SessionID:    us.SessionID,
		DeviceInfo:   us.DeviceInfo,
		IPAddress:    us.IPAddress,
		UserAgent:    us.UserAgent,
		ExpiresAt:    us.ExpiresAt,
		LastActivity: us.LastActivity,
		IsActive:     us.IsActive,
		CreatedAt:    us.CreatedAt,
		UpdatedAt:    us.UpdatedAt,
	}
}

// Deactivate deactivates the session
func (us *UserSession) Deactivate() {
	us.IsActive = false
}

// Activate activates the session
func (us *UserSession) Activate() {
	us.IsActive = true
}

// ExtendExpiration extends the session expiration time
func (us *UserSession) ExtendExpiration(duration time.Duration) {
	us.ExpiresAt = time.Now().Add(duration)
}