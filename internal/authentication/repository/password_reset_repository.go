package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/VincentArjuna/RexiErp/internal/authentication/model"
	"github.com/VincentArjuna/RexiErp/internal/shared/database"
)

// PasswordResetRepository interface defines the contract for password reset token operations
type PasswordResetRepository interface {
	Create(ctx context.Context, token *model.PasswordResetToken) error
	GetByToken(ctx context.Context, token string) (*model.PasswordResetToken, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*model.PasswordResetToken, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, activeOnly bool) ([]*model.PasswordResetToken, error)
	Update(ctx context.Context, token *model.PasswordResetToken) error
	DeactivateByUserID(ctx context.Context, userID uuid.UUID) error
	DeactivateExpiredTokens(ctx context.Context) error
	Delete(ctx context.Context, tokenID uuid.UUID) error
	SoftDelete(ctx context.Context, tokenID uuid.UUID) error
	CountActiveByUserID(ctx context.Context, userID uuid.UUID, since time.Time) (int64, error)
}

// passwordResetRepository implements PasswordResetRepository interface
type passwordResetRepository struct {
	db     *database.Database
	logger *logrus.Logger
}

// NewPasswordResetRepository creates a new instance of PasswordResetRepository
func NewPasswordResetRepository(db *database.Database, logger *logrus.Logger) PasswordResetRepository {
	return &passwordResetRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new password reset token
func (r *passwordResetRepository) Create(ctx context.Context, token *model.PasswordResetToken) error {
	r.logger.WithFields(logrus.Fields{
		"user_id":   token.UserID,
		"tenant_id": token.TenantID,
		"email":     token.Email,
	}).Debug("Creating password reset token")

	if err := r.db.DB.WithContext(ctx).Create(token).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"user_id":   token.UserID,
			"tenant_id": token.TenantID,
			"email":     token.Email,
			"error":     err,
		}).Error("Failed to create password reset token")
		return fmt.Errorf("failed to create password reset token: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"token_id":  token.ID,
		"user_id":   token.UserID,
		"tenant_id": token.TenantID,
		"email":     token.Email,
	}).Info("Password reset token created successfully")

	return nil
}

// GetByToken retrieves a password reset token by token value
func (r *passwordResetRepository) GetByToken(ctx context.Context, token string) (*model.PasswordResetToken, error) {
	r.logger.WithField("token_id", "hash").Debug("Getting password reset token by token")

	var resetToken model.PasswordResetToken
	if err := r.db.DB.WithContext(ctx).
		Preload("User").
		Where("token = ? AND deleted_at IS NULL", token).
		First(&resetToken).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("Password reset token not found")
			return nil, fmt.Errorf("token not found")
		}
		r.logger.WithField("error", err).Error("Failed to get password reset token")
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"token_id": resetToken.ID,
		"user_id":  resetToken.UserID,
		"email":    resetToken.Email,
	}).Debug("Password reset token retrieved successfully")

	return &resetToken, nil
}

// GetByTokenHash retrieves a password reset token by token hash
func (r *passwordResetRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*model.PasswordResetToken, error) {
	r.logger.WithField("token_hash", tokenHash).Debug("Getting password reset token by hash")

	var resetToken model.PasswordResetToken
	if err := r.db.DB.WithContext(ctx).
		Preload("User").
		Where("token_hash = ? AND deleted_at IS NULL", tokenHash).
		First(&resetToken).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("Password reset token not found by hash")
			return nil, fmt.Errorf("token not found")
		}
		r.logger.WithFields(logrus.Fields{
			"token_hash": tokenHash,
			"error":      err,
		}).Error("Failed to get password reset token by hash")
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"token_id": resetToken.ID,
		"user_id":  resetToken.UserID,
		"email":    resetToken.Email,
	}).Debug("Password reset token retrieved successfully by hash")

	return &resetToken, nil
}

// GetByUserID retrieves all password reset tokens for a user
func (r *passwordResetRepository) GetByUserID(ctx context.Context, userID uuid.UUID, activeOnly bool) ([]*model.PasswordResetToken, error) {
	r.logger.WithField("user_id", userID).Debug("Getting password reset tokens by user ID")

	var tokens []*model.PasswordResetToken
	query := r.db.DB.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", userID)

	if activeOnly {
		query = query.Where("is_active = ? AND expires_at > ?", true, time.Now())
	}

	if err := query.Order("created_at DESC").Find(&tokens).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err,
		}).Error("Failed to get password reset tokens by user ID")
		return nil, fmt.Errorf("failed to get tokens: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"count":   len(tokens),
	}).Debug("Password reset tokens retrieved successfully by user ID")

	return tokens, nil
}

// Update updates a password reset token
func (r *passwordResetRepository) Update(ctx context.Context, token *model.PasswordResetToken) error {
	r.logger.WithFields(logrus.Fields{
		"token_id": token.ID,
		"user_id":  token.UserID,
	}).Debug("Updating password reset token")

	if err := r.db.DB.WithContext(ctx).Save(token).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"token_id": token.ID,
			"user_id":  token.UserID,
			"error":    err,
		}).Error("Failed to update password reset token")
		return fmt.Errorf("failed to update token: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"token_id": token.ID,
		"user_id":  token.UserID,
	}).Info("Password reset token updated successfully")

	return nil
}

// DeactivateByUserID deactivates all active password reset tokens for a user
func (r *passwordResetRepository) DeactivateByUserID(ctx context.Context, userID uuid.UUID) error {
	r.logger.WithField("user_id", userID).Debug("Deactivating all password reset tokens for user")

	result := r.db.DB.WithContext(ctx).
		Model(&model.PasswordResetToken{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Update("is_active", false)

	if result.Error != nil {
		r.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   result.Error,
		}).Error("Failed to deactivate password reset tokens")
		return fmt.Errorf("failed to deactivate tokens: %w", result.Error)
	}

	r.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"deactivated": result.RowsAffected,
	}).Info("Password reset tokens deactivated successfully")

	return nil
}

// DeactivateExpiredTokens deactivates all expired password reset tokens
func (r *passwordResetRepository) DeactivateExpiredTokens(ctx context.Context) error {
	r.logger.Debug("Deactivating expired password reset tokens")

	result := r.db.DB.WithContext(ctx).
		Model(&model.PasswordResetToken{}).
		Where("is_active = ? AND expires_at < ?", true, time.Now()).
		Update("is_active", false)

	if result.Error != nil {
		r.logger.WithField("error", result.Error).Error("Failed to deactivate expired password reset tokens")
		return fmt.Errorf("failed to deactivate expired tokens: %w", result.Error)
	}

	r.logger.WithField("deactivated", result.RowsAffected).Info("Expired password reset tokens deactivated successfully")

	return nil
}

// Delete hard deletes a password reset token (use with caution)
func (r *passwordResetRepository) Delete(ctx context.Context, tokenID uuid.UUID) error {
	r.logger.WithField("token_id", tokenID).Warn("Hard deleting password reset token")

	if err := r.db.DB.WithContext(ctx).Unscoped().Delete(&model.PasswordResetToken{}, tokenID).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"token_id": tokenID,
			"error":    err,
		}).Error("Failed to delete password reset token")
		return fmt.Errorf("failed to delete token: %w", err)
	}

	r.logger.WithField("token_id", tokenID).Info("Password reset token deleted successfully")
	return nil
}

// SoftDelete soft deletes a password reset token (recommended)
func (r *passwordResetRepository) SoftDelete(ctx context.Context, tokenID uuid.UUID) error {
	r.logger.WithField("token_id", tokenID).Debug("Soft deleting password reset token")

	if err := r.db.DB.WithContext(ctx).Delete(&model.PasswordResetToken{}, tokenID).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"token_id": tokenID,
			"error":    err,
		}).Error("Failed to soft delete password reset token")
		return fmt.Errorf("failed to soft delete token: %w", err)
	}

	r.logger.WithField("token_id", tokenID).Info("Password reset token soft deleted successfully")
	return nil
}

// CountActiveByUserID counts active password reset tokens for a user since a given time
func (r *passwordResetRepository) CountActiveByUserID(ctx context.Context, userID uuid.UUID, since time.Time) (int64, error) {
	r.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"since":   since,
	}).Debug("Counting active password reset tokens for user")

	var count int64
	if err := r.db.DB.WithContext(ctx).
		Model(&model.PasswordResetToken{}).
		Where("user_id = ? AND is_active = ? AND created_at >= ?", userID, true, since).
		Count(&count).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"since":   since,
			"error":   err,
		}).Error("Failed to count active password reset tokens")
		return 0, fmt.Errorf("failed to count tokens: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"since":   since,
		"count":   count,
	}).Debug("Active password reset tokens counted successfully")

	return count, nil
}