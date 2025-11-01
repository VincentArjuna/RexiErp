package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/VincentArjuna/RexiErp/internal/authentication/service"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService service.AuthService
	logger      *logrus.Logger
}

// NewAuthHandler creates a new instance of AuthHandler
func NewAuthHandler(authService service.AuthService, logger *logrus.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Creates a new user account with the provided details
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration request"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// Extract tenant ID from context (should be set by middleware)
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		h.respondWithError(c, http.StatusBadRequest, "Tenant ID required", "Tenant context not found")
		return
	}

	// Convert tenant ID to UUID
	tenantUUID, ok := tenantID.(uuid.UUID)
	if !ok {
		h.respondWithError(c, http.StatusBadRequest, "Invalid tenant ID", "Tenant ID format is invalid")
		return
	}

	// Create service request
	serviceReq := &service.RegisterRequest{
		Email:       req.Email,
		Password:    req.Password,
		FullName:    req.FullName,
		PhoneNumber: req.PhoneNumber,
		Role:        req.Role,
		TenantID:    tenantUUID,
	}

	// Call auth service
	response, err := h.authService.Register(c.Request.Context(), serviceReq)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"email":     req.Email,
			"tenant_id": tenantUUID,
			"error":     err,
		}).Error("User registration failed")

		if contains(err.Error(), "already exists") {
			h.respondWithError(c, http.StatusConflict, "User already exists", err.Error())
			return
		}

		h.respondWithError(c, http.StatusInternalServerError, "Registration failed", err.Error())
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":   response.User.ID,
		"email":     response.User.Email,
		"tenant_id": response.User.TenantID,
	}).Info("User registered successfully")

	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Message: "User registered successfully",
		Data:    response,
	})
}

// Login handles user authentication
// @Summary Authenticate user
// @Description Authenticates a user with email and password
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// Get client information
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Call auth service
	response, err := h.authService.Login(c.Request.Context(), &service.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}, ipAddress, userAgent)

	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"email":      req.Email,
			"ip_address": ipAddress,
			"error":      err,
		}).Warn("User login failed")

		if contains(err.Error(), "invalid credentials") || contains(err.Error(), "not found") {
			h.respondWithError(c, http.StatusUnauthorized, "Invalid credentials", "Email or password is incorrect")
			return
		}

		if contains(err.Error(), "inactive") {
			h.respondWithError(c, http.StatusForbidden, "Account inactive", "Your account is not active")
			return
		}

		h.respondWithError(c, http.StatusInternalServerError, "Login failed", err.Error())
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":    response.User.ID,
		"email":      response.User.Email,
		"tenant_id":  response.User.TenantID,
		"ip_address": ipAddress,
		"session_id": response.SessionID,
	}).Info("User logged in successfully")

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Login successful",
		Data:    response,
	})
}

// Logout handles user logout
// @Summary Logout user
// @Description Logs out a user by deactivating their session
// @Tags authentication
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get session ID from token (should be set by auth middleware)
	sessionID, exists := c.Get("session_id")
	if !exists {
		h.respondWithError(c, http.StatusBadRequest, "Session ID required", "No active session found")
		return
	}

	sessionIDStr, ok := sessionID.(string)
	if !ok {
		h.respondWithError(c, http.StatusBadRequest, "Invalid session", "Session ID format is invalid")
		return
	}

	// Call auth service
	if err := h.authService.Logout(c.Request.Context(), sessionIDStr); err != nil {
		h.logger.WithFields(logrus.Fields{
			"session_id": sessionIDStr,
			"error":      err,
		}).Error("User logout failed")

		if contains(err.Error(), "not found") {
			h.respondWithError(c, http.StatusNotFound, "Session not found", "No active session found")
			return
		}

		h.respondWithError(c, http.StatusInternalServerError, "Logout failed", err.Error())
		return
	}

	h.logger.WithField("session_id", sessionIDStr).Info("User logged out successfully")

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Logout successful",
	})
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Refreshes an access token using a refresh token
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// Call auth service
	response, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.logger.WithField("error", err).Warn("Token refresh failed")

		if contains(err.Error(), "invalid") || contains(err.Error(), "expired") {
			h.respondWithError(c, http.StatusUnauthorized, "Invalid token", "Refresh token is invalid or expired")
			return
		}

		if contains(err.Error(), "not found") {
			h.respondWithError(c, http.StatusUnauthorized, "Session not found", "No active session found")
			return
		}

		h.respondWithError(c, http.StatusInternalServerError, "Token refresh failed", err.Error())
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":    response.User.ID,
		"session_id": response.SessionID,
	}).Info("Token refreshed successfully")

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Token refreshed successfully",
		Data:    response,
	})
}

// GetProfile handles user profile retrieval
// @Summary Get user profile
// @Description Retrieves the current user's profile information
// @Tags authentication
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} UserDTO
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Get user ID from token (should be set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.respondWithError(c, http.StatusUnauthorized, "Unauthorized", "User context not found")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		h.respondWithError(c, http.StatusUnauthorized, "Invalid user", "User ID format is invalid")
		return
	}

	// Call auth service
	user, err := h.authService.GetProfile(c.Request.Context(), userUUID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"user_id": userUUID,
			"error":   err,
		}).Error("Failed to get user profile")

		if contains(err.Error(), "not found") {
			h.respondWithError(c, http.StatusNotFound, "User not found", "User profile not found")
			return
		}

		h.respondWithError(c, http.StatusInternalServerError, "Failed to get profile", err.Error())
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Profile retrieved successfully",
		Data:    UserToDTO(user),
	})
}

// UpdateProfile handles user profile update
// @Summary Update user profile
// @Description Updates the current user's profile information
// @Tags authentication
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body UpdateProfileRequest true "Update profile request"
// @Success 200 {object} UserDTO
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// Get user ID from token
	userID, exists := c.Get("user_id")
	if !exists {
		h.respondWithError(c, http.StatusUnauthorized, "Unauthorized", "User context not found")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		h.respondWithError(c, http.StatusUnauthorized, "Invalid user", "User ID format is invalid")
		return
	}

	// Create service request
	serviceReq := &service.UpdateProfileRequest{
		FullName:    req.FullName,
		PhoneNumber: req.PhoneNumber,
	}

	// Call auth service
	user, err := h.authService.UpdateProfile(c.Request.Context(), userUUID, serviceReq)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"user_id": userUUID,
			"error":   err,
		}).Error("Failed to update user profile")

		if contains(err.Error(), "not found") {
			h.respondWithError(c, http.StatusNotFound, "User not found", "User profile not found")
			return
		}

		h.respondWithError(c, http.StatusInternalServerError, "Failed to update profile", err.Error())
		return
	}

	h.logger.WithField("user_id", userUUID).Info("User profile updated successfully")

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Profile updated successfully",
		Data:    UserToDTO(user),
	})
}

// ChangePassword handles password change
// @Summary Change user password
// @Description Changes the current user's password
// @Tags authentication
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body ChangePasswordRequest true "Change password request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// Get user ID from token
	userID, exists := c.Get("user_id")
	if !exists {
		h.respondWithError(c, http.StatusUnauthorized, "Unauthorized", "User context not found")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		h.respondWithError(c, http.StatusUnauthorized, "Invalid user", "User ID format is invalid")
		return
	}

	// Create service request
	serviceReq := &service.ChangePasswordRequest{
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}

	// Call auth service
	if err := h.authService.ChangePassword(c.Request.Context(), userUUID, serviceReq); err != nil {
		h.logger.WithFields(logrus.Fields{
			"user_id": userUUID,
			"error":   err,
		}).Error("Failed to change password")

		if contains(err.Error(), "incorrect") {
			h.respondWithError(c, http.StatusUnauthorized, "Invalid current password", "Current password is incorrect")
			return
		}

		h.respondWithError(c, http.StatusInternalServerError, "Failed to change password", err.Error())
		return
	}

	h.logger.WithField("user_id", userUUID).Info("User password changed successfully")

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Password changed successfully",
	})
}

// GetSessions handles retrieval of user sessions
// @Summary Get user sessions
// @Description Retrieves all active sessions for the current user
// @Tags authentication
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {array} SessionDTO
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/sessions [get]
func (h *AuthHandler) GetSessions(c *gin.Context) {
	// Get user ID from token
	userID, exists := c.Get("user_id")
	if !exists {
		h.respondWithError(c, http.StatusUnauthorized, "Unauthorized", "User context not found")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		h.respondWithError(c, http.StatusUnauthorized, "Invalid user", "User ID format is invalid")
		return
	}

	// Call auth service
	sessions, err := h.authService.GetUserSessions(c.Request.Context(), userUUID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"user_id": userUUID,
			"error":   err,
		}).Error("Failed to get user sessions")

		h.respondWithError(c, http.StatusInternalServerError, "Failed to get sessions", err.Error())
		return
	}

	// Convert to DTOs
	sessionDTOs := make([]*SessionDTO, len(sessions))
	for i, session := range sessions {
		sessionDTOs[i] = SessionToDTO(session)
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Sessions retrieved successfully",
		Data:    sessionDTOs,
	})
}

// LogoutAll handles logout from all sessions
// @Summary Logout from all sessions
// @Description Logs out the user from all active sessions
// @Tags authentication
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	// Get user ID from token
	userID, exists := c.Get("user_id")
	if !exists {
		h.respondWithError(c, http.StatusUnauthorized, "Unauthorized", "User context not found")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		h.respondWithError(c, http.StatusUnauthorized, "Invalid user", "User ID format is invalid")
		return
	}

	// Call auth service
	if err := h.authService.DeactivateAllSessions(c.Request.Context(), userUUID); err != nil {
		h.logger.WithFields(logrus.Fields{
			"user_id": userUUID,
			"error":   err,
		}).Error("Failed to logout from all sessions")

		h.respondWithError(c, http.StatusInternalServerError, "Failed to logout from all sessions", err.Error())
		return
	}

	h.logger.WithField("user_id", userUUID).Info("User logged out from all sessions successfully")

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Logged out from all sessions successfully",
	})
}

// RequestPasswordReset handles password reset requests
// @Summary Request password reset
// @Description Sends a password reset link to the user's email
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body PasswordResetRequest true "Password reset request"
// @Success 200 {object} PasswordResetResponse
// @Failure 400 {object} ErrorResponse
// @Failure 429 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/password-reset [post]
func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var req PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// Create service request
	serviceReq := &service.PasswordResetRequest{
		Email: req.Email,
	}

	// Call auth service
	response, err := h.authService.RequestPasswordReset(c.Request.Context(), serviceReq)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"email": req.Email,
			"error": err,
		}).Error("Failed to process password reset request")

		h.respondWithError(c, http.StatusInternalServerError, "Failed to process password reset request", err.Error())
		return
	}

	// Log the attempt (without revealing if user exists)
	h.logger.WithField("email", req.Email).Info("Password reset request processed")

	c.JSON(http.StatusOK, PasswordResetResponse{
		Message:      response.Message,
		ResetTokenID: response.ResetTokenID,
		ExpiresAt:    response.ExpiresAt,
		SentToEmail:  response.SentToEmail,
		RateLimited:  response.RateLimited,
	})
}

// ValidateResetToken handles password reset token validation
// @Summary Validate reset token
// @Description Validates if a password reset token is valid
// @Tags authentication
// @Produce json
// @Param token query string true "Reset token"
// @Success 200 {object} service.ResetTokenValidationResult
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/validate-reset-token [get]
func (h *AuthHandler) ValidateResetToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		h.respondWithError(c, http.StatusBadRequest, "Token required", "Reset token is required")
		return
	}

	// Call auth service
	result, err := h.authService.ValidateResetToken(c.Request.Context(), token)
	if err != nil {
		h.logger.WithField("token", "****").Error("Failed to validate reset token")
		h.respondWithError(c, http.StatusInternalServerError, "Failed to validate reset token", err.Error())
		return
	}

	statusCode := http.StatusOK
	if !result.IsValid {
		statusCode = http.StatusNotFound
	}

	c.JSON(statusCode, result)
}

// ResetPassword handles password reset with token
// @Summary Reset password
// @Description Resets user password using a valid reset token
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Reset password request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// Create service request
	serviceReq := &service.ResetPasswordRequest{
		Token:       req.Token,
		NewPassword: req.NewPassword,
	}

	// Call auth service
	if err := h.authService.ResetPassword(c.Request.Context(), serviceReq); err != nil {
		h.logger.WithField("token", "****").Error("Failed to reset password")

		if contains(err.Error(), "invalid reset token") {
			h.respondWithError(c, http.StatusNotFound, "Invalid or expired token", err.Error())
			return
		}

		h.respondWithError(c, http.StatusInternalServerError, "Failed to reset password", err.Error())
		return
	}

	h.logger.Info("Password reset completed successfully")

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Password reset successfully. You can now login with your new password.",
	})
}

// Helper functions

func (h *AuthHandler) respondWithError(c *gin.Context, statusCode int, errorType, message string) {
	response := ErrorResponse{
		Error:   errorType,
		Message: message,
		Code:    getErrorCode(statusCode),
	}

	c.JSON(statusCode, response)
}

func getErrorCode(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusInternalServerError:
		return "INTERNAL_SERVER_ERROR"
	default:
		return "UNKNOWN_ERROR"
	}
}

func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr ||
		(len(str) > len(substr) &&
			(str[:len(substr)] == substr ||
			 str[len(str)-len(substr):] == substr ||
			 findSubstring(str, substr))))
}

func findSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}