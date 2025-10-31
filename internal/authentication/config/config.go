package config

import (
	"github.com/VincentArjuna/RexiErp/internal/shared/config"
	"os"
	"strconv"
)

// AuthServiceConfig represents authentication service specific configuration
type AuthServiceConfig struct {
	config.Config `yaml:",inline"`
	Auth         AuthConfig `yaml:"auth"`
}

// AuthConfig represents authentication-specific configuration
type AuthConfig struct {
	MinPasswordLength      int    `yaml:"min_password_length"`
	RequireSpecialChars    bool   `yaml:"require_special_chars"`
	RequireNumbers         bool   `yaml:"require_numbers"`
	RequireUppercase       bool   `yaml:"require_uppercase"`
	MaxLoginAttempts       int    `yaml:"max_login_attempts"`
	AccountLockoutDuration string `yaml:"account_lockout_duration"`
	SessionTimeout         string `yaml:"session_timeout"`
	PasswordResetTokenTTL  string `yaml:"password_reset_token_ttl"`
	EmailVerificationTTL   string `yaml:"email_verification_ttl"`
}

// LoadAuthServiceConfig loads authentication service configuration
func LoadAuthServiceConfig() (*AuthServiceConfig, error) {
	baseConfig, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	authConfig := &AuthServiceConfig{
		Config: *baseConfig,
		Auth: AuthConfig{
			MinPasswordLength:      getEnvInt("AUTH_MIN_PASSWORD_LENGTH", 8),
			RequireSpecialChars:    getEnvBool("AUTH_REQUIRE_SPECIAL_CHARS", true),
			RequireNumbers:         getEnvBool("AUTH_REQUIRE_NUMBERS", true),
			RequireUppercase:       getEnvBool("AUTH_REQUIRE_UPPERCASE", true),
			MaxLoginAttempts:       getEnvInt("AUTH_MAX_LOGIN_ATTEMPTS", 5),
			AccountLockoutDuration: getEnv("AUTH_ACCOUNT_LOCKOUT_DURATION", "15m"),
			SessionTimeout:         getEnv("AUTH_SESSION_TIMEOUT", "24h"),
			PasswordResetTokenTTL:  getEnv("AUTH_PASSWORD_RESET_TOKEN_TTL", "1h"),
			EmailVerificationTTL:   getEnv("AUTH_EMAIL_VERIFICATION_TTL", "24h"),
		},
	}

	return authConfig, nil
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}