// Package service provides authentication business logic for RexiERP.
//
// This package handles user authentication, authorization, JWT token management,
// and session management. All types and interfaces are centralized in this file
// to prevent circular dependencies and IDE resolution issues.
//
// IMPORTANT: To prevent future type resolution issues:
// 1. ALL shared types must be defined in this file
// 2. Never define the same type in multiple files
// 3. Update this file when adding new types
package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v5"

	"github.com/VincentArjuna/RexiErp/internal/authentication/model"
)

// ============================================================================
// Service Interfaces
// ============================================================================

// AuthService defines the contract for authentication business logic operations.
// This interface provides all core authentication functionality including user
// registration, login, token management, and session handling.
type AuthService interface {
	// User Management
	Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*model.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req *UpdateProfileRequest) (*model.User, error)

	// Authentication
	Login(ctx context.Context, req *LoginRequest, ipAddress, userAgent string) (*AuthResponse, error)
	Logout(ctx context.Context, sessionID string) error
	ChangePassword(ctx context.Context, userID uuid.UUID, req *ChangePasswordRequest) error

	// Password Reset Flow
	RequestPasswordReset(ctx context.Context, req *PasswordResetRequest) (*PasswordResetResponse, error)
	ValidateResetToken(ctx context.Context, token string) (*ResetTokenValidationResult, error)
	ResetPassword(ctx context.Context, req *ResetPasswordRequest) error

	// Token Management
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
	ValidateToken(ctx context.Context, tokenString string) (*TokenValidationResult, error)

	// Session Management
	DeactivateAllSessions(ctx context.Context, userID uuid.UUID) error
	GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*model.UserSession, error)
}

// JWTService defines the contract for JWT token operations.
// This interface handles the creation, validation, and parsing of JWT tokens
// for authentication and authorization purposes.
type JWTService interface {
	GenerateTokenPair(user *User, sessionID string) (accessToken, refreshToken string, err error)
	GenerateAccessToken(user *User, sessionID string) (string, error)
	GenerateRefreshToken(user *User, sessionID string) (string, error)
	ValidateToken(tokenString string) (*TokenClaims, error)
	ExtractTokenFromHeader(authHeader string) (string, error)
}

// ============================================================================
// Core Domain Types
// ============================================================================

// User represents a simplified user model for JWT token generation and validation.
// This is different from the full model.User and is optimized specifically for
// authentication operations and token claims.
type User struct {
	// ID is the unique identifier for the user
	ID uuid.UUID `json:"id"`

	// TenantID identifies the tenant context for multi-tenant operations
	TenantID uuid.UUID `json:"tenant_id"`

	// Email is the user's email address used for authentication
	Email string `json:"email"`

	// Role represents the user's permission level within the tenant
	Role string `json:"role"`
}

// TokenValidationResult represents the result of JWT token validation.
// It contains all necessary information to determine if a token is valid
// and to extract user context from valid tokens.
type TokenValidationResult struct {
	// IsValid indicates whether the token passed all validation checks
	IsValid bool `json:"is_valid"`

	// UserID is extracted from valid tokens and identifies the authenticated user
	UserID uuid.UUID `json:"user_id,omitempty"`

	// TenantID is extracted from valid tokens and identifies the user's tenant
	TenantID uuid.UUID `json:"tenant_id,omitempty"`

	// Role represents the user's permission level from the token claims
	Role string `json:"role,omitempty"`

	// SessionID identifies the specific session this token is associated with
	SessionID string `json:"session_id,omitempty"`

	// ExpiresAt indicates when the token becomes invalid
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// TokenClaims represents the JWT token claims structure.
// This implements the jwt.Claims interface for token validation.
type TokenClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	SessionID string    `json:"session_id"`
	TokenType string    `json:"token_type"` // "access" or "refresh"
	TokenHash string    `json:"token_hash"`
	jwt.RegisteredClaims
}

// ============================================================================
// Configuration Types
// ============================================================================

// AuthConfig represents authentication service configuration settings.
// All timing values use Go's time.Duration for clarity and consistency.
type AuthConfig struct {
	// Password Policy
	MinPasswordLength      int           `json:"min_password_length"`
	RequireSpecialChars    bool          `json:"require_special_chars"`
	RequireNumbers         bool          `json:"require_numbers"`
	RequireUppercase       bool          `json:"require_uppercase"`

	// Security Settings
	MaxLoginAttempts       int           `json:"max_login_attempts"`
	AccountLockoutDuration time.Duration `json:"account_lockout_duration"`
	SessionTimeout         time.Duration `json:"session_timeout"`

	// Token Lifetimes
	PasswordResetTokenTTL  time.Duration `json:"password_reset_token_ttl"`
	EmailVerificationTTL   time.Duration `json:"email_verification_ttl"`
	AccessTokenTTL         time.Duration `json:"access_token_ttl"`
	RefreshTokenTTL        time.Duration `json:"refresh_token_ttl"`

	// JWT Configuration
	JWTSecret              string        `json:"-"` // Hidden from JSON output for security
	JWTIssuer              string        `json:"jwt_issuer"`
}

// ============================================================================
// Request/Response Types
// ============================================================================

// RegisterRequest represents user registration input with validation requirements.
type RegisterRequest struct {
	Email       string    `json:"email"`
	Password    string    `json:"password"`
	FullName    string    `json:"full_name"`
	PhoneNumber string    `json:"phone_number"`
	Role        string    `json:"role"`
	TenantID    uuid.UUID `json:"tenant_id"`
}

// LoginRequest represents user login credentials.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UpdateProfileRequest represents user profile update data.
// Pointer types are used to distinguish between empty and unchanged fields.
type UpdateProfileRequest struct {
	FullName    *string `json:"full_name,omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty"`
}

// ChangePasswordRequest represents password change data with security validation.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// PasswordResetRequest represents password reset request with email.
type PasswordResetRequest struct {
	Email string `json:"email"`
}

// PasswordResetResponse represents response after password reset request.
type PasswordResetResponse struct {
	Message       string    `json:"message"`
	ResetTokenID  string    `json:"reset_token_id,omitempty"`
	ExpiresAt     time.Time `json:"expires_at"`
	SentToEmail   string    `json:"sent_to_email"`
	RateLimited   bool      `json:"rate_limited"`
}

// ResetPasswordRequest represents password reset with token verification.
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// ResetTokenValidationResult represents password reset token validation result.
type ResetTokenValidationResult struct {
	IsValid     bool      `json:"is_valid"`
	UserID      uuid.UUID `json:"user_id,omitempty"`
	TenantID    uuid.UUID `json:"tenant_id,omitempty"`
	Email       string    `json:"email,omitempty"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	ErrorMessage string   `json:"error_message,omitempty"`
}

// AuthResponse represents successful authentication response with tokens and user data.
type AuthResponse struct {
	User         *model.User `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	TokenType    string      `json:"token_type"`
	ExpiresIn    int64       `json:"expires_in"` // seconds
	SessionID    string      `json:"session_id"`
}

// NOTE: model.User and other model types are imported from the model package
// to avoid circular dependencies and maintain clean separation of concerns.