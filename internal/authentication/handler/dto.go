package handler

import (
	"time"

	"github.com/google/uuid"

	"github.com/VincentArjuna/RexiErp/internal/authentication/model"
)

// RegisterRequest represents the request payload for user registration
type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email" example:"user@example.com"`
	Password    string `json:"password" binding:"required,min=8" example:"SecurePass123!"`
	FullName    string `json:"full_name" binding:"required,min=2,max=255" example:"John Doe"`
	PhoneNumber string `json:"phone_number" binding:"omitempty,e164" example:"+6281234567890"`
	Role        string `json:"role,omitempty" binding:"omitempty,oneof=super_admin tenant_admin staff viewer" example:"staff"`
}

// LoginRequest represents the request payload for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"SecurePass123!"`
}

// RefreshTokenRequest represents the request payload for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// UpdateProfileRequest represents the request payload for updating user profile
type UpdateProfileRequest struct {
	FullName    *string `json:"full_name,omitempty" binding:"omitempty,min=2,max=255" example:"John Smith"`
	PhoneNumber *string `json:"phone_number,omitempty" binding:"omitempty,e164" example:"+6281234567890"`
}

// ChangePasswordRequest represents the request payload for changing password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required" example:"OldPass123!"`
	NewPassword     string `json:"new_password" binding:"required,min=8" example:"NewPass123!"`
}

// PasswordResetRequest represents the request payload for password reset
type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}

// ResetPasswordRequest represents the request payload for resetting password with token
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	NewPassword string `json:"new_password" binding:"required,min=8" example:"NewSecurePass123!"`
}

// AuthResponse represents the response payload for authentication
type AuthResponse struct {
	AccessToken  string    `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string    `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType    string    `json:"token_type" example:"Bearer"`
	ExpiresIn    int64     `json:"expires_in" example:"86400"`
	User         UserDTO   `json:"user"`
	SessionID    string    `json:"session_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// PasswordResetResponse represents the response payload for password reset request
type PasswordResetResponse struct {
	Message      string    `json:"message" example:"If an account with this email exists, a password reset link has been sent"`
	ResetTokenID string    `json:"reset_token_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	ExpiresAt    time.Time `json:"expires_at" example:"2024-01-16T10:30:00Z"`
	SentToEmail  string    `json:"sent_to_email" example:"us****@example.com"`
	RateLimited  bool      `json:"rate_limited" example:"false"`
}

// UserDTO represents the user data transferred in responses
type UserDTO struct {
	ID          uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	TenantID    uuid.UUID  `json:"tenant_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	Email       string     `json:"email" example:"user@example.com"`
	FullName    string     `json:"full_name" example:"John Doe"`
	PhoneNumber string     `json:"phone_number,omitempty" example:"+6281234567890"`
	Role        string     `json:"role" example:"staff"`
	IsActive    bool       `json:"is_active" example:"true"`
	LastLogin   *time.Time `json:"last_login,omitempty" example:"2024-01-15T10:30:00Z"`
	CreatedAt   time.Time  `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time  `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

// SessionDTO represents session data transferred in responses
type SessionDTO struct {
	ID           uuid.UUID              `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	SessionID    string                 `json:"session_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	DeviceInfo   *model.DeviceInfo      `json:"device_info,omitempty"`
	IPAddress    string                 `json:"ip_address" example:"192.168.1.100"`
	UserAgent    string                 `json:"user_agent,omitempty" example:"Mozilla/5.0..."`
	ExpiresAt    time.Time              `json:"expires_at" example:"2024-01-16T10:30:00Z"`
	LastActivity time.Time              `json:"last_activity" example:"2024-01-15T11:30:00Z"`
	IsActive     bool                   `json:"is_active" example:"true"`
	CreatedAt    time.Time              `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt    time.Time              `json:"updated_at" example:"2024-01-15T11:30:00Z"`
}

// ActivityLogDTO represents activity log data transferred in responses
type ActivityLogDTO struct {
	ID           uuid.UUID              `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID       *uuid.UUID             `json:"user_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	Action       string                 `json:"action" example:"login"`
	ResourceType string                 `json:"resource_type" example:"user"`
	ResourceID   *uuid.UUID             `json:"resource_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	IPAddress    string                 `json:"ip_address,omitempty" example:"192.168.1.100"`
	UserAgent    string                 `json:"user_agent,omitempty" example:"Mozilla/5.0..."`
	SessionID    string                 `json:"session_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	Success      bool                   `json:"success" example:"true"`
	ErrorMessage string                 `json:"error_message,omitempty" example:"Invalid credentials"`
	Context      *model.ActivityContext `json:"context,omitempty"`
	CreatedAt    time.Time              `json:"created_at" example:"2024-01-15T10:30:00Z"`
}

// ErrorResponse represents the standard error response
type ErrorResponse struct {
	Error   string            `json:"error" example:"Validation failed"`
	Message string            `json:"message" example:"Email is required"`
	Code    string            `json:"code,omitempty" example:"VALIDATION_ERROR"`
	Details map[string]string `json:"details,omitempty"`
}

// SuccessResponse represents the standard success response
type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Operation completed successfully"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Success     bool        `json:"success" example:"true"`
	Message     string      `json:"message" example:"Data retrieved successfully"`
	Data        interface{} `json:"data"`
	Total       int64       `json:"total" example:"100"`
	Page        int         `json:"page" example:"1"`
	PerPage     int         `json:"per_page" example:"20"`
	TotalPages  int         `json:"total_pages" example:"5"`
	HasNext     bool        `json:"has_next" example:"true"`
	HasPrevious bool        `json:"has_previous" example:"false"`
}

// Validation helper functions

// IsEmpty checks if a string pointer is nil or empty
func IsEmpty(s *string) bool {
	return s == nil || *s == ""
}

// ToPtr returns a pointer to the given value
func ToPtr(s string) *string {
	return &s
}

// UserToDTO converts a User model to UserDTO
func UserToDTO(user *model.User) *UserDTO {
	if user == nil {
		return nil
	}

	return &UserDTO{
		ID:          user.ID,
		TenantID:    user.TenantID,
		Email:       user.Email,
		FullName:    user.FullName,
		PhoneNumber: user.PhoneNumber,
		Role:        string(user.Role),
		IsActive:    user.IsActive,
		LastLogin:   user.LastLogin,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

// SessionToDTO converts a UserSession model to SessionDTO
func SessionToDTO(session *model.UserSession) *SessionDTO {
	if session == nil {
		return nil
	}

	var deviceInfo *model.DeviceInfo
	if session.DeviceInfo != "" {
		var err error
		deviceInfo, err = session.GetDeviceInfo()
		if err != nil {
			deviceInfo = nil
		}
	}

	return &SessionDTO{
		ID:           session.ID,
		SessionID:    session.SessionID,
		DeviceInfo:   deviceInfo,
		IPAddress:    session.IPAddress,
		UserAgent:    session.UserAgent,
		ExpiresAt:    session.ExpiresAt,
		LastActivity: session.LastActivity,
		IsActive:     session.IsActive,
		CreatedAt:    session.CreatedAt,
		UpdatedAt:    session.UpdatedAt,
	}
}

// ActivityLogToDTO converts an ActivityLog model to ActivityLogDTO
func ActivityLogToDTO(log *model.ActivityLog) *ActivityLogDTO {
	if log == nil {
		return nil
	}

	var context *model.ActivityContext
	if log.Context != "" {
		var err error
		context, err = log.GetContext()
		if err != nil {
			context = nil
		}
	}

	return &ActivityLogDTO{
		ID:           log.ID,
		UserID:       log.UserID,
		Action:       log.Action,
		ResourceType: log.ResourceType,
		ResourceID:   log.ResourceID,
		IPAddress:    log.IPAddress,
		UserAgent:    log.UserAgent,
		SessionID:    log.SessionID,
		Success:      log.Success,
		ErrorMessage: log.ErrorMessage,
		Context:      context,
		CreatedAt:    log.CreatedAt,
	}
}