package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// jwtService implements JWTService interface
type jwtService struct {
	secret           []byte
	issuer           string
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
}

// NewJWTService creates a new instance of JWTService
func NewJWTService(secret, issuer string, accessTokenTTL, refreshTokenTTL time.Duration) JWTService {
	return &jwtService{
		secret:          []byte(secret),
		issuer:          issuer,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

// GenerateTokenPair generates both access and refresh tokens
func (j *jwtService) GenerateTokenPair(user *User, sessionID string) (accessToken, refreshToken string, err error) {
	// Generate access token
	accessToken, err = j.GenerateAccessToken(user, sessionID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err = j.GenerateRefreshToken(user, sessionID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// GenerateAccessToken generates a new access token
func (j *jwtService) GenerateAccessToken(user *User, sessionID string) (string, error) {
	now := time.Now()

	claims := &TokenClaims{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		Email:     user.Email,
		Role:      user.Role,
		SessionID: sessionID,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Issuer:    j.issuer,
			Subject:   user.ID.String(),
			Audience:  []string{"rexi-erp"},
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTokenTTL)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// GenerateRefreshToken generates a new refresh token
func (j *jwtService) GenerateRefreshToken(user *User, sessionID string) (string, error) {
	now := time.Now()

	claims := &TokenClaims{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		Email:     user.Email,
		Role:      user.Role,
		SessionID: sessionID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Issuer:    j.issuer,
			Subject:   user.ID.String(),
			Audience:  []string{"rexi-erp"},
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshTokenTTL)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// ValidateToken validates a JWT token and returns the claims
func (j *jwtService) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate issuer
	if claims.Issuer != j.issuer {
		return nil, fmt.Errorf("invalid token issuer")
	}

	// Validate audience
	if !containsAudience(claims.Audience, "rexi-erp") {
		return nil, fmt.Errorf("invalid token audience")
	}

	return claims, nil
}

// ExtractTokenFromHeader extracts the token from the Authorization header
func (j *jwtService) ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is required")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", fmt.Errorf("authorization header must be in format 'Bearer <token>'")
	}

	return authHeader[len(bearerPrefix):], nil
}

// GetTokenExpirationTime returns the expiration time for a given token type
func (j *jwtService) GetTokenExpirationTime(tokenType string) time.Duration {
	switch tokenType {
	case "access":
		return j.accessTokenTTL
	case "refresh":
		return j.refreshTokenTTL
	default:
		return j.accessTokenTTL
	}
}

// IsTokenExpired checks if a token is expired
func (j *jwtService) IsTokenExpired(claims *TokenClaims) bool {
	if claims.ExpiresAt == nil {
		return true
	}
	return time.Now().After(claims.ExpiresAt.Time)
}

// GetTokenRemainingTime returns the remaining time until token expiration
func (j *jwtService) GetTokenRemainingTime(claims *TokenClaims) time.Duration {
	if claims.ExpiresAt == nil {
		return 0
	}
	remaining := claims.ExpiresAt.Time.Sub(time.Now())
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Helper function to check if audience claim contains the expected value
func containsAudience(audience []string, expected string) bool {
	for _, aud := range audience {
		if aud == expected {
			return true
		}
	}
	return false
}