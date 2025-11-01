package service

import (
	"time"

	"github.com/VincentArjuna/RexiErp/internal/authentication/config"
)

// NewAuthConfig creates a new AuthConfig from AuthServiceConfig
func NewAuthConfig(cfg *config.AuthServiceConfig) *AuthConfig {
	return &AuthConfig{
		MinPasswordLength:      cfg.Auth.MinPasswordLength,
		RequireSpecialChars:    cfg.Auth.RequireSpecialChars,
		RequireNumbers:         cfg.Auth.RequireNumbers,
		RequireUppercase:       cfg.Auth.RequireUppercase,
		MaxLoginAttempts:       cfg.Auth.MaxLoginAttempts,
		AccountLockoutDuration: parseDuration(cfg.Auth.AccountLockoutDuration),
		SessionTimeout:         parseDuration(cfg.Auth.SessionTimeout),
		PasswordResetTokenTTL:  parseDuration(cfg.Auth.PasswordResetTokenTTL),
		EmailVerificationTTL:   parseDuration(cfg.Auth.EmailVerificationTTL),
		AccessTokenTTL:         cfg.JWT.AccessTokenTTL,
		RefreshTokenTTL:        time.Duration(cfg.JWT.RefreshTokenDays) * 24 * time.Hour,
		JWTSecret:              cfg.JWT.Secret,
		JWTIssuer:              cfg.JWT.Issuer,
	}
}

// parseDuration parses a duration string and returns time.Duration
func parseDuration(durationStr string) time.Duration {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		// Return default values if parsing fails
		switch durationStr {
		case "account_lockout_duration":
			return 15 * time.Minute
		case "session_timeout":
			return 24 * time.Hour
		case "password_reset_token_ttl":
			return 1 * time.Hour
		case "email_verification_ttl":
			return 24 * time.Hour
		default:
			return 1 * time.Hour
		}
	}
	return duration
}