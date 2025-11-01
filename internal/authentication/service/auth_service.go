package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"github.com/VincentArjuna/RexiErp/internal/authentication/model"
	"github.com/VincentArjuna/RexiErp/internal/authentication/repository"
	"github.com/VincentArjuna/RexiErp/internal/shared/cache"
)

// authService implements AuthService interface
type authService struct {
	userRepo        repository.UserRepository
	sessionRepo     repository.SessionRepository
	activityRepo    repository.ActivityRepository
	passwordResetRepo repository.PasswordResetRepository
	cache           *cache.RedisCache
	jwtService      JWTService
	logger          *logrus.Logger
	config          *AuthConfig
}

// NewAuthService creates a new instance of AuthService
func NewAuthService(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	activityRepo repository.ActivityRepository,
	passwordResetRepo repository.PasswordResetRepository,
	cache *cache.RedisCache,
	jwtService JWTService,
	logger *logrus.Logger,
	config *AuthConfig,
) AuthService {
	return &authService{
		userRepo:        userRepo,
		sessionRepo:     sessionRepo,
		activityRepo:    activityRepo,
		passwordResetRepo: passwordResetRepo,
		cache:           cache,
		jwtService:      jwtService,
		logger:          logger,
		config:          config,
	}
}

// Register registers a new user
func (s *authService) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"email":     req.Email,
		"tenant_id": req.TenantID,
	}).Debug("Registering new user")

	// Validate input
	if err := s.validateRegistrationRequest(req); err != nil {
		s.logActivity(ctx, nil, req.TenantID, "register", "user", nil, nil, nil,
			false, fmt.Sprintf("Validation failed: %v", err), "")
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if user already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email, req.TenantID)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"email":     req.Email,
			"tenant_id": req.TenantID,
			"error":     err,
		}).Error("Failed to check if user exists")
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	if exists {
		s.logActivity(ctx, nil, req.TenantID, "register", "user", nil, nil, nil,
			false, "User already exists", "")
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.WithField("error", err).Error("Failed to hash password")
		return nil, fmt.Errorf("failed to process password: %w", err)
	}

	// Parse role
	role := model.RoleViewer // default role
	if req.Role != "" {
		role = model.UserRole(req.Role)
	}

	// Create user
	user := &model.User{
		TenantID:     req.TenantID,
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		PasswordHash: string(hashedPassword),
		FullName:     strings.TrimSpace(req.FullName),
		PhoneNumber:  strings.TrimSpace(req.PhoneNumber),
		Role:         role,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.WithFields(logrus.Fields{
			"email":     req.Email,
			"tenant_id": req.TenantID,
			"error":     err,
		}).Error("Failed to create user")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create session and tokens
	authResponse, err := s.createSessionAndTokens(ctx, user, "127.0.0.1", "Registration")
	if err != nil {
		// Rollback user creation
		_ = s.userRepo.SoftDelete(ctx, user.ID)
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	s.logActivity(ctx, &user.ID, req.TenantID, "register", "user", &user.ID,
		map[string]interface{}{}, map[string]interface{}{"email": user.Email},
		true, "", authResponse.SessionID)

	s.logger.WithFields(logrus.Fields{
		"user_id":   user.ID,
		"email":     user.Email,
		"tenant_id": user.TenantID,
	}).Info("User registered successfully")

	return authResponse, nil
}

// Login authenticates a user
func (s *authService) Login(ctx context.Context, req *LoginRequest, ipAddress, userAgent string) (*AuthResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"email":      req.Email,
		"ip_address": ipAddress,
	}).Debug("User login attempt")

	// Get user by email - we need to determine tenant ID from email or have it in the request
	// For now, we'll assume tenant ID is known or we have a way to determine it
	// In a real implementation, you might have a subdomain or other way to identify tenant

	// Find user across all tenants (for login, we might not know tenant beforehand)
	// This is a simplified approach - in production, you'd have tenant identification
	user, err := s.findUserByEmail(ctx, req.Email)
	if err != nil {
		s.logActivity(ctx, nil, uuid.Nil, "login", "user", nil, nil, nil,
			false, "User not found", "")
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		s.logActivity(ctx, &user.ID, user.TenantID, "login", "user", &user.ID, nil, nil,
			false, "Invalid password", "")
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is active
	if !user.IsActiveUser() {
		s.logActivity(ctx, &user.ID, user.TenantID, "login", "user", &user.ID, nil, nil,
			false, "User account is inactive", "")
		return nil, fmt.Errorf("account is inactive")
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		s.logger.WithFields(logrus.Fields{
			"user_id": user.ID,
			"error":   err,
		}).Warn("Failed to update last login")
	}

	// Create session and tokens
	authResponse, err := s.createSessionAndTokens(ctx, user, ipAddress, userAgent)
	if err != nil {
		s.logActivity(ctx, &user.ID, user.TenantID, "login", "user", &user.ID, nil, nil,
			false, fmt.Sprintf("Failed to create session: %v", err), "")
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	s.logActivity(ctx, &user.ID, user.TenantID, "login", "user", &user.ID, nil, nil,
		true, "", authResponse.SessionID)

	s.logger.WithFields(logrus.Fields{
		"user_id":   user.ID,
		"email":     user.Email,
		"tenant_id": user.TenantID,
		"ip_address": ipAddress,
	}).Info("User logged in successfully")

	return authResponse, nil
}

// Logout logs out a user by deactivating the session
func (s *authService) Logout(ctx context.Context, sessionID string) error {
	s.logger.WithField("session_id", sessionID).Debug("User logout")

	// Get session
	session, err := s.sessionRepo.GetBySessionID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found")
	}

	// Deactivate session
	if err := s.sessionRepo.Deactivate(ctx, sessionID); err != nil {
		s.logger.WithFields(logrus.Fields{
			"session_id": sessionID,
			"error":      err,
		}).Error("Failed to deactivate session")
		return fmt.Errorf("failed to logout: %w", err)
	}

	// Add token to blacklist (if using Redis for token blacklisting)
	if err := s.cache.Set(ctx, fmt.Sprintf("blacklist:%s", sessionID), true, time.Until(session.ExpiresAt)); err != nil {
		s.logger.WithFields(logrus.Fields{
			"session_id": sessionID,
			"error":      err,
		}).Warn("Failed to add token to blacklist")
	}

	s.logActivity(ctx, &session.UserID, session.TenantID, "logout", "session", &session.ID,
		nil, nil, true, "", sessionID)

	s.logger.WithField("session_id", sessionID).Info("User logged out successfully")
	return nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	s.logger.Debug("Refreshing access token")

	// Validate refresh token
	claims, err := s.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("invalid token type")
	}

	// Get session
	session, err := s.sessionRepo.GetByTokenHash(ctx, claims.TokenHash)
	if err != nil {
		return nil, fmt.Errorf("session not found")
	}

	// Check if session is valid
	if !session.IsValid() {
		return nil, fmt.Errorf("session expired or inactive")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is active
	if !user.IsActiveUser() {
		return nil, fmt.Errorf("account is inactive")
	}

	// Generate new tokens
	serviceUser := &User{
		ID:       user.ID,
		TenantID: user.TenantID,
		Email:    user.Email,
		Role:     string(user.Role),
	}
	accessToken, refreshTokenNew, err := s.jwtService.GenerateTokenPair(serviceUser, session.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Update session with new token hash
	session.TokenHash = s.hashToken(accessToken)
	session.RefreshTokenHash = s.hashToken(refreshTokenNew)
	session.UpdateActivity()
	session.ExtendExpiration(s.config.AccessTokenTTL)

	if err := s.sessionRepo.Update(ctx, session); err != nil {
		s.logger.WithField("error", err).Error("Failed to update session")
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	s.logActivity(ctx, &user.ID, user.TenantID, "refresh_token", "session", &session.ID,
		nil, nil, true, "", session.SessionID)

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenNew,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.config.AccessTokenTTL.Seconds()),
		User:         user.SanitizeForResponse(),
		SessionID:    session.SessionID,
	}, nil
}

// GetProfile retrieves a user's profile
func (s *authService) GetProfile(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	s.logger.WithField("user_id", userID).Debug("Getting user profile")

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}

// UpdateProfile updates a user's profile
func (s *authService) UpdateProfile(ctx context.Context, userID uuid.UUID, req *UpdateProfileRequest) (*model.User, error) {
	s.logger.WithField("user_id", userID).Debug("Updating user profile")

	// Get current user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Store old values for activity log
	oldValues := map[string]interface{}{
		"full_name":    user.FullName,
		"phone_number": user.PhoneNumber,
	}

	// Update fields
	if req.FullName != nil && *req.FullName != "" {
		user.FullName = strings.TrimSpace(*req.FullName)
	}
	if req.PhoneNumber != nil {
		user.PhoneNumber = strings.TrimSpace(*req.PhoneNumber)
	}

	// Update user
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	// Log activity
	newValues := map[string]interface{}{
		"full_name":    user.FullName,
		"phone_number": user.PhoneNumber,
	}

	s.logActivity(ctx, &user.ID, user.TenantID, "update_profile", "user", &user.ID,
		oldValues, newValues, true, "", "")

	s.logger.WithField("user_id", userID).Info("User profile updated successfully")
	return user, nil
}

// ChangePassword changes a user's password
func (s *authService) ChangePassword(ctx context.Context, userID uuid.UUID, req *ChangePasswordRequest) error {
	s.logger.WithField("user_id", userID).Debug("Changing user password")

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		s.logActivity(ctx, &user.ID, user.TenantID, "change_password", "user", &user.ID,
			nil, nil, false, "Invalid current password", "")
		return fmt.Errorf("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to process new password: %w", err)
	}

	// Update password
	user.PasswordHash = string(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Deactivate all other sessions for security
	if err := s.sessionRepo.DeactivateByUserID(ctx, userID); err != nil {
		s.logger.WithField("error", err).Warn("Failed to deactivate other sessions")
	}

	s.logActivity(ctx, &user.ID, user.TenantID, "change_password", "user", &user.ID,
		nil, nil, true, "", "")

	s.logger.WithField("user_id", userID).Info("User password changed successfully")
	return nil
}

// ValidateToken validates a JWT token
func (s *authService) ValidateToken(ctx context.Context, tokenString string) (*TokenValidationResult, error) {
	s.logger.Debug("Validating token")

	// Check if token is blacklisted
	tokenHash := s.hashToken(tokenString)
	isBlacklisted, err := s.cache.Exists(ctx, fmt.Sprintf("blacklist:%s", tokenHash))
	if err != nil {
		s.logger.WithField("error", err).Warn("Failed to check token blacklist")
	}

	if isBlacklisted {
		return &TokenValidationResult{IsValid: false}, nil
	}

	// Validate JWT token
	claims, err := s.jwtService.ValidateToken(tokenString)
	if err != nil {
		return &TokenValidationResult{IsValid: false}, nil
	}

	if claims.TokenType != "access" {
		return &TokenValidationResult{IsValid: false}, nil
	}

	// Get session to validate
	session, err := s.sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil || !session.IsValid() {
		return &TokenValidationResult{IsValid: false}, nil
	}

	return &TokenValidationResult{
		IsValid:   true,
		UserID:    claims.UserID,
		TenantID:  claims.TenantID,
		Role:      claims.Role,
		SessionID: claims.SessionID,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}

// DeactivateAllSessions deactivates all sessions for a user
func (s *authService) DeactivateAllSessions(ctx context.Context, userID uuid.UUID) error {
	s.logger.WithField("user_id", userID).Debug("Deactivating all user sessions")

	return s.sessionRepo.DeactivateByUserID(ctx, userID)
}

// GetUserSessions retrieves all active sessions for a user
func (s *authService) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*model.UserSession, error) {
	s.logger.WithField("user_id", userID).Debug("Getting user sessions")

	return s.sessionRepo.GetByUserID(ctx, userID, true)
}

// Helper functions

func (s *authService) validateRegistrationRequest(req *RegisterRequest) error {
	// Validate email
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		return fmt.Errorf("invalid email format")
	}

	// Validate password
	if len(req.Password) < s.config.MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters long", s.config.MinPasswordLength)
	}

	if s.config.RequireUppercase {
		hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(req.Password)
		if !hasUpper {
			return fmt.Errorf("password must contain at least one uppercase letter")
		}
	}

	if s.config.RequireNumbers {
		hasNumber := regexp.MustCompile(`[0-9]`).MatchString(req.Password)
		if !hasNumber {
			return fmt.Errorf("password must contain at least one number")
		}
	}

	if s.config.RequireSpecialChars {
		hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(req.Password)
		if !hasSpecial {
			return fmt.Errorf("password must contain at least one special character")
		}
	}

	// Validate full name
	if strings.TrimSpace(req.FullName) == "" {
		return fmt.Errorf("full name is required")
	}

	// Validate role
	if req.Role != "" {
		validRoles := map[string]bool{
			"super_admin":  true,
			"tenant_admin": true,
			"staff":        true,
			"viewer":       true,
		}
		if !validRoles[req.Role] {
			return fmt.Errorf("invalid role: %s", req.Role)
		}
	}

	return nil
}

func (s *authService) createSessionAndTokens(ctx context.Context, user *model.User, ipAddress, userAgent string) (*AuthResponse, error) {
	// Generate JWT tokens
	serviceUser := &User{
		ID:       user.ID,
		TenantID: user.TenantID,
		Email:    user.Email,
		Role:     string(user.Role),
	}
	accessToken, refreshToken, err := s.jwtService.GenerateTokenPair(serviceUser, "")
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Create session
	sessionID := uuid.New().String()
	session := &model.UserSession{
		UserID:           user.ID,
		TenantID:         user.TenantID,
		SessionID:        sessionID,
		TokenHash:        s.hashToken(accessToken),
		RefreshTokenHash: s.hashToken(refreshToken),
		IPAddress:        ipAddress,
		UserAgent:        userAgent,
		ExpiresAt:        time.Now().Add(s.config.AccessTokenTTL),
		LastActivity:     time.Now(),
		IsActive:         true,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.config.AccessTokenTTL.Seconds()),
		User:         user.SanitizeForResponse(),
		SessionID:    sessionID,
	}, nil
}

func (s *authService) findUserByEmail(ctx context.Context, email string) (*model.User, error) {
	// For login, we need to find users across all tenants since the tenant might not be known beforehand
	// In a production environment, you might have subdomain, header, or other tenant identification methods
	// For now, we'll search across all tenants and validate the user exists

	s.logger.WithField("email", email).Debug("Finding user by email across all tenants for login")

	user, err := s.userRepo.FindByEmailAcrossTenants(ctx, email)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"email": email,
			"error": err,
		}).Debug("User not found during login attempt")
		return nil, fmt.Errorf("invalid credentials")
	}

	// Additional validation can be added here if needed
	// For example, checking if the tenant is active, etc.

	return user, nil
}

func (s *authService) hashToken(token string) string {
	// Implement proper SHA-256 hashing for token security
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}

// RequestPasswordReset handles password reset requests
func (s *authService) RequestPasswordReset(ctx context.Context, req *PasswordResetRequest) (*PasswordResetResponse, error) {
	s.logger.WithField("email", req.Email).Debug("Processing password reset request")

	// Validate email format
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return nil, fmt.Errorf("invalid email format")
	}

	// Find user by email across all tenants
	user, err := s.findUserByEmail(ctx, email)
	if err != nil {
		// Don't reveal if user exists or not for security
		s.logger.WithField("email", email).Debug("Password reset requested for non-existent user")
		return &PasswordResetResponse{
			Message:     "If an account with this email exists, a password reset link has been sent",
			ExpiresAt:   time.Now().Add(s.config.PasswordResetTokenTTL),
			SentToEmail: s.maskEmail(email),
			RateLimited: false,
		}, nil
	}

	// Check rate limiting - prevent too many reset requests
	since := time.Now().Add(-1 * time.Hour) // Allow max 3 requests per hour
	count, err := s.passwordResetRepo.CountActiveByUserID(ctx, user.ID, since)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"user_id": user.ID,
			"error":   err,
		}).Error("Failed to check password reset rate limit")
	}

	if count >= 3 {
		s.logger.WithFields(logrus.Fields{
			"user_id": user.ID,
			"email":   email,
			"count":   count,
		}).Warn("Password reset rate limit exceeded")
		return &PasswordResetResponse{
			Message:     "Too many password reset requests. Please try again later",
			ExpiresAt:   time.Now().Add(s.config.PasswordResetTokenTTL),
			SentToEmail: s.maskEmail(email),
			RateLimited: true,
		}, nil
	}

	// Deactivate any existing active tokens for this user
	if err := s.passwordResetRepo.DeactivateByUserID(ctx, user.ID); err != nil {
		s.logger.WithFields(logrus.Fields{
			"user_id": user.ID,
			"error":   err,
		}).Warn("Failed to deactivate existing password reset tokens")
	}

	// Generate reset token
	resetToken := uuid.New().String()
	tokenHash := s.hashToken(resetToken)

	// Create password reset token record
	passwordResetToken := &model.PasswordResetToken{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		Token:     resetToken,
		TokenHash: tokenHash,
		Email:     email,
		ExpiresAt: time.Now().Add(s.config.PasswordResetTokenTTL),
		IsActive:  true,
	}

	if err := s.passwordResetRepo.Create(ctx, passwordResetToken); err != nil {
		s.logger.WithFields(logrus.Fields{
			"user_id": user.ID,
			"email":   email,
			"error":   err,
		}).Error("Failed to create password reset token")
		return nil, fmt.Errorf("failed to create password reset token")
	}

	// TODO: Send email with reset token
	// In a real implementation, you would integrate with an email service
	// For now, we'll just log the token (IN PRODUCTION, NEVER LOG TOKENS)
	s.logger.WithFields(logrus.Fields{
		"user_id":     user.ID,
		"email":       email,
		"reset_token": resetToken, // REMOVE IN PRODUCTION
	}).Info("Password reset token generated (email not implemented)")

	// Log activity
	s.logActivity(ctx, &user.ID, user.TenantID, "password_reset_request", "user", &user.ID,
		nil, map[string]interface{}{"email": email}, true, "", "")

	s.logger.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"email":    email,
		"token_id": passwordResetToken.ID,
	}).Info("Password reset request processed successfully")

	return &PasswordResetResponse{
		Message:      "If an account with this email exists, a password reset link has been sent",
		ResetTokenID: passwordResetToken.ID.String(),
		ExpiresAt:    passwordResetToken.ExpiresAt,
		SentToEmail:  s.maskEmail(email),
		RateLimited:  false,
	}, nil
}

// ValidateResetToken validates a password reset token
func (s *authService) ValidateResetToken(ctx context.Context, token string) (*ResetTokenValidationResult, error) {
	s.logger.Debug("Validating password reset token")

	if token == "" {
		return &ResetTokenValidationResult{
			IsValid:     false,
			ErrorMessage: "Token is required",
		}, nil
	}

	// Hash the token and look it up
	tokenHash := s.hashToken(token)
	passwordResetToken, err := s.passwordResetRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		s.logger.WithField("token_hash", tokenHash).Debug("Password reset token not found")
		return &ResetTokenValidationResult{
			IsValid:     false,
			ErrorMessage: "Invalid or expired token",
		}, nil
	}

	// Check if token is valid
	if !passwordResetToken.IsValid() {
		var errorMsg string
		if passwordResetToken.IsExpired() {
			errorMsg = "Token has expired"
		} else if passwordResetToken.IsUsed() {
			errorMsg = "Token has already been used"
		} else {
			errorMsg = "Token is inactive"
		}

		s.logger.WithFields(logrus.Fields{
			"token_id": passwordResetToken.ID,
			"expired":  passwordResetToken.IsExpired(),
			"used":     passwordResetToken.IsUsed(),
			"active":   passwordResetToken.IsActive,
		}).Debug("Password reset token validation failed")

		return &ResetTokenValidationResult{
			IsValid:     false,
			ErrorMessage: errorMsg,
		}, nil
	}

	s.logger.WithFields(logrus.Fields{
		"token_id": passwordResetToken.ID,
		"user_id":  passwordResetToken.UserID,
		"email":    passwordResetToken.Email,
	}).Debug("Password reset token validation successful")

	return &ResetTokenValidationResult{
		IsValid:   true,
		UserID:    passwordResetToken.UserID,
		TenantID:  passwordResetToken.TenantID,
		Email:     passwordResetToken.Email,
		ExpiresAt: passwordResetToken.ExpiresAt,
	}, nil
}

// ResetPassword resets a user's password using a valid token
func (s *authService) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	s.logger.Debug("Processing password reset")

	if req.Token == "" {
		return fmt.Errorf("reset token is required")
	}
	if req.NewPassword == "" {
		return fmt.Errorf("new password is required")
	}

	// Validate password strength
	if len(req.NewPassword) < s.config.MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters long", s.config.MinPasswordLength)
	}

	// Validate the reset token first
	validation, err := s.ValidateResetToken(ctx, req.Token)
	if err != nil {
		return fmt.Errorf("failed to validate reset token: %w", err)
	}

	if !validation.IsValid {
		return fmt.Errorf("invalid reset token: %s", validation.ErrorMessage)
	}

	// Get the user
	user, err := s.userRepo.GetByID(ctx, validation.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to process new password: %w", err)
	}

	// Store old password hash for activity log
	oldValues := map[string]interface{}{
		"password_hash": user.PasswordHash,
	}

	// Update user password
	user.PasswordHash = string(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Get and mark the reset token as used
	tokenHash := s.hashToken(req.Token)
	passwordResetToken, err := s.passwordResetRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		s.logger.WithField("token_hash", tokenHash).Warn("Failed to find reset token after validation")
	} else {
		passwordResetToken.MarkAsUsed()
		if err := s.passwordResetRepo.Update(ctx, passwordResetToken); err != nil {
			s.logger.WithFields(logrus.Fields{
				"token_id": passwordResetToken.ID,
				"error":    err,
			}).Warn("Failed to mark reset token as used")
		}
	}

	// Deactivate all other sessions for security
	if err := s.sessionRepo.DeactivateByUserID(ctx, user.ID); err != nil {
		s.logger.WithFields(logrus.Fields{
			"user_id": user.ID,
			"error":   err,
		}).Warn("Failed to deactivate user sessions after password reset")
	}

	// Deactivate any other active reset tokens
	if err := s.passwordResetRepo.DeactivateByUserID(ctx, user.ID); err != nil {
		s.logger.WithFields(logrus.Fields{
			"user_id": user.ID,
			"error":   err,
		}).Warn("Failed to deactivate other password reset tokens")
	}

	// Log activity
	s.logActivity(ctx, &user.ID, user.TenantID, "password_reset", "user", &user.ID,
		oldValues, map[string]interface{}{"password_reset": true}, true, "", "")

	s.logger.WithFields(logrus.Fields{
		"user_id": user.ID,
		"email":   user.Email,
	}).Info("Password reset completed successfully")

	return nil
}

// maskEmail masks an email address for security in responses
func (s *authService) maskEmail(email string) string {
	if len(email) < 4 {
		return "****"
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "****"
	}

	username := parts[0]
	domain := parts[1]

	if len(username) <= 2 {
		return "****@" + domain
	}

	masked := username[:2] + strings.Repeat("*", len(username)-2)
	return masked + "@" + domain
}

func (s *authService) logActivity(ctx context.Context, userID *uuid.UUID, tenantID uuid.UUID,
	action, resourceType string, resourceID *uuid.UUID, oldValues, newValues map[string]interface{},
	success bool, errorMessage, sessionID string) {

	// Create activity log
	activity := &model.ActivityLog{
		UserID:       userID,
		TenantID:     tenantID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Success:      success,
		ErrorMessage: errorMessage,
		SessionID:    sessionID,
	}

	// Set old and new values
	if oldValues != nil {
		_ = activity.SetOldValues(oldValues)
	}
	if newValues != nil {
		_ = activity.SetNewValues(newValues)
	}

	// Create activity log asynchronously
	go func() {
		if err := s.activityRepo.Create(context.Background(), activity); err != nil {
			s.logger.WithFields(logrus.Fields{
				"action": action,
				"error":  err,
			}).Error("Failed to create activity log")
		}
	}()
}