package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/VincentArjuna/RexiErp/internal/authentication/model"
	"github.com/VincentArjuna/RexiErp/internal/shared/database"
)

// UserRepository interface defines the contract for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByEmail(ctx context.Context, email string, tenantID uuid.UUID) (*model.User, error)
	FindByEmailAcrossTenants(ctx context.Context, email string) (*model.User, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*model.User, error)
	Update(ctx context.Context, user *model.User) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	Delete(ctx context.Context, userID uuid.UUID) error
	SoftDelete(ctx context.Context, userID uuid.UUID) error
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error)
	ExistsByEmail(ctx context.Context, email string, tenantID uuid.UUID) (bool, error)
	SearchUsers(ctx context.Context, tenantID uuid.UUID, query string, limit, offset int) ([]*model.User, error)
}

// userRepository implements UserRepository interface
type userRepository struct {
	db     *database.Database
	logger *logrus.Logger
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository(db *database.Database, logger *logrus.Logger) UserRepository {
	return &userRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	r.logger.WithFields(logrus.Fields{
		"email":     user.Email,
		"tenant_id": user.TenantID,
		"role":      user.Role,
	}).Debug("Creating new user")

	if err := r.db.DB.WithContext(ctx).Create(user).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"email":     user.Email,
			"tenant_id": user.TenantID,
			"error":     err,
		}).Error("Failed to create user")
		return fmt.Errorf("failed to create user: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"user_id":   user.ID,
		"email":     user.Email,
		"tenant_id": user.TenantID,
	}).Info("User created successfully")

	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	r.logger.WithField("user_id", id).Debug("Getting user by ID")

	var user model.User
	if err := r.db.DB.WithContext(ctx).
		Preload("UserSessions").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WithField("user_id", id).Debug("User not found")
			return nil, fmt.Errorf("user not found")
		}
		r.logger.WithFields(logrus.Fields{
			"user_id": id,
			"error":   err,
		}).Error("Failed to get user by ID")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"user_id":   user.ID,
		"email":     user.Email,
		"tenant_id": user.TenantID,
	}).Debug("User retrieved successfully")

	return &user, nil
}

// GetByEmail retrieves a user by email and tenant ID
func (r *userRepository) GetByEmail(ctx context.Context, email string, tenantID uuid.UUID) (*model.User, error) {
	r.logger.WithFields(logrus.Fields{
		"email":     email,
		"tenant_id": tenantID,
	}).Debug("Getting user by email")

	var user model.User
	if err := r.db.DB.WithContext(ctx).
		Where("email = ? AND tenant_id = ? AND deleted_at IS NULL", email, tenantID).
		First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WithFields(logrus.Fields{
				"email":     email,
				"tenant_id": tenantID,
			}).Debug("User not found by email")
			return nil, fmt.Errorf("user not found")
		}
		r.logger.WithFields(logrus.Fields{
			"email":     email,
			"tenant_id": tenantID,
			"error":     err,
		}).Error("Failed to get user by email")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"user_id":   user.ID,
		"email":     user.Email,
		"tenant_id": user.TenantID,
	}).Debug("User retrieved successfully by email")

	return &user, nil
}

// GetByTenantID retrieves users by tenant ID with pagination
func (r *userRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*model.User, error) {
	r.logger.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"limit":     limit,
		"offset":    offset,
	}).Debug("Getting users by tenant ID")

	var users []*model.User
	if err := r.db.DB.WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"tenant_id": tenantID,
			"error":     err,
		}).Error("Failed to get users by tenant ID")
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"count":     len(users),
	}).Debug("Users retrieved successfully by tenant ID")

	return users, nil
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	r.logger.WithFields(logrus.Fields{
		"user_id":   user.ID,
		"email":     user.Email,
		"tenant_id": user.TenantID,
	}).Debug("Updating user")

	if err := r.db.DB.WithContext(ctx).Save(user).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"user_id":   user.ID,
			"email":     user.Email,
			"tenant_id": user.TenantID,
			"error":     err,
		}).Error("Failed to update user")
		return fmt.Errorf("failed to update user: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"user_id":   user.ID,
		"email":     user.Email,
		"tenant_id": user.TenantID,
	}).Info("User updated successfully")

	return nil
}

// UpdateLastLogin updates the last login timestamp for a user
func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	r.logger.WithField("user_id", userID).Debug("Updating user last login")

	now := time.Now()
	if err := r.db.DB.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", userID).
		Update("last_login", now).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err,
		}).Error("Failed to update user last login")
		return fmt.Errorf("failed to update last login: %w", err)
	}

	r.logger.WithField("user_id", userID).Debug("User last login updated successfully")
	return nil
}

// Delete hard deletes a user (use with caution)
func (r *userRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	r.logger.WithField("user_id", userID).Warn("Hard deleting user")

	if err := r.db.DB.WithContext(ctx).Unscoped().Delete(&model.User{}, userID).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err,
		}).Error("Failed to delete user")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	r.logger.WithField("user_id", userID).Info("User deleted successfully")
	return nil
}

// SoftDelete soft deletes a user (recommended)
func (r *userRepository) SoftDelete(ctx context.Context, userID uuid.UUID) error {
	r.logger.WithField("user_id", userID).Debug("Soft deleting user")

	if err := r.db.DB.WithContext(ctx).Delete(&model.User{}, userID).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err,
		}).Error("Failed to soft delete user")
		return fmt.Errorf("failed to soft delete user: %w", err)
	}

	r.logger.WithField("user_id", userID).Info("User soft deleted successfully")
	return nil
}

// CountByTenant counts users by tenant ID
func (r *userRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	r.logger.WithField("tenant_id", tenantID).Debug("Counting users by tenant ID")

	var count int64
	if err := r.db.DB.WithContext(ctx).
		Model(&model.User{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Count(&count).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"tenant_id": tenantID,
			"error":     err,
		}).Error("Failed to count users by tenant ID")
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"count":     count,
	}).Debug("Users counted successfully by tenant ID")

	return count, nil
}

// ExistsByEmail checks if a user exists by email and tenant ID
func (r *userRepository) ExistsByEmail(ctx context.Context, email string, tenantID uuid.UUID) (bool, error) {
	r.logger.WithFields(logrus.Fields{
		"email":     email,
		"tenant_id": tenantID,
	}).Debug("Checking if user exists by email")

	var count int64
	if err := r.db.DB.WithContext(ctx).
		Model(&model.User{}).
		Where("email = ? AND tenant_id = ? AND deleted_at IS NULL", email, tenantID).
		Count(&count).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"email":     email,
			"tenant_id": tenantID,
			"error":     err,
		}).Error("Failed to check if user exists by email")
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	exists := count > 0
	r.logger.WithFields(logrus.Fields{
		"email":     email,
		"tenant_id": tenantID,
		"exists":    exists,
	}).Debug("User existence check completed")

	return exists, nil
}

// SearchUsers searches users by query string within a tenant
func (r *userRepository) SearchUsers(ctx context.Context, tenantID uuid.UUID, query string, limit, offset int) ([]*model.User, error) {
	r.logger.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"query":     query,
		"limit":     limit,
		"offset":    offset,
	}).Debug("Searching users")

	var users []*model.User
	searchPattern := "%" + query + "%"

	if err := r.db.DB.WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL AND (email ILIKE ? OR full_name ILIKE ?)", tenantID, searchPattern, searchPattern).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"tenant_id": tenantID,
			"query":     query,
			"error":     err,
		}).Error("Failed to search users")
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"query":     query,
		"count":     len(users),
	}).Debug("Users searched successfully")

	return users, nil
}

// FindByEmailAcrossTenants finds a user by email across all tenants
func (r *userRepository) FindByEmailAcrossTenants(ctx context.Context, email string) (*model.User, error) {
	r.logger.WithField("email", email).Debug("Finding user by email across all tenants")

	var user model.User
	if err := r.db.DB.WithContext(ctx).
		Where("email = ? AND deleted_at IS NULL", strings.ToLower(strings.TrimSpace(email))).
		First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WithField("email", email).Debug("User not found by email across all tenants")
			return nil, fmt.Errorf("user not found")
		}
		r.logger.WithFields(logrus.Fields{
			"email": email,
			"error": err,
		}).Error("Failed to find user by email across all tenants")
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"user_id":   user.ID,
		"email":     user.Email,
		"tenant_id": user.TenantID,
	}).Debug("User found successfully by email across all tenants")

	return &user, nil
}