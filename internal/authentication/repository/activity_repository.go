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

// ActivityRepository interface defines the contract for activity log data operations
type ActivityRepository interface {
	Create(ctx context.Context, activity *model.ActivityLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.ActivityLog, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.ActivityLog, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*model.ActivityLog, error)
	GetByResource(ctx context.Context, resourceType string, resourceID uuid.UUID, limit, offset int) ([]*model.ActivityLog, error)
	SearchActivities(ctx context.Context, tenantID uuid.UUID, filters ActivityFilters, limit, offset int) ([]*model.ActivityLog, error)
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error)
	CountByAction(ctx context.Context, tenantID uuid.UUID, action string) (int64, error)
	DeleteOldActivities(ctx context.Context, olderThan time.Time) (int64, error)
}

// ActivityFilters represents filters for searching activities
type ActivityFilters struct {
	Action       string     `json:"action"`
	ResourceType string     `json:"resource_type"`
	UserID       *uuid.UUID `json:"user_id"`
	Success      *bool      `json:"success"`
	StartDate    *time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	IPAddress    string     `json:"ip_address"`
	SessionID    string     `json:"session_id"`
}

// activityRepository implements ActivityRepository interface
type activityRepository struct {
	db     *database.Database
	logger *logrus.Logger
}

// NewActivityRepository creates a new instance of ActivityRepository
func NewActivityRepository(db *database.Database, logger *logrus.Logger) ActivityRepository {
	return &activityRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new activity log entry
func (r *activityRepository) Create(ctx context.Context, activity *model.ActivityLog) error {
	r.logger.WithFields(logrus.Fields{
		"action":       activity.Action,
		"resource_type": activity.ResourceType,
		"user_id":      activity.UserID,
		"tenant_id":    activity.TenantID,
	}).Debug("Creating activity log")

	if err := r.db.DB.WithContext(ctx).Create(activity).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"action":       activity.Action,
			"resource_type": activity.ResourceType,
			"user_id":      activity.UserID,
			"tenant_id":    activity.TenantID,
			"error":        err,
		}).Error("Failed to create activity log")
		return fmt.Errorf("failed to create activity log: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"activity_id":  activity.ID,
		"action":       activity.Action,
		"resource_type": activity.ResourceType,
		"tenant_id":    activity.TenantID,
	}).Debug("Activity log created successfully")

	return nil
}

// GetByID retrieves an activity log entry by ID
func (r *activityRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.ActivityLog, error) {
	r.logger.WithField("activity_id", id).Debug("Getting activity log by ID")

	var activity model.ActivityLog
	if err := r.db.DB.WithContext(ctx).
		Preload("User").
		Where("id = ?", id).
		First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WithField("activity_id", id).Debug("Activity log not found")
			return nil, fmt.Errorf("activity log not found")
		}
		r.logger.WithFields(logrus.Fields{
			"activity_id": id,
			"error":       err,
		}).Error("Failed to get activity log by ID")
		return nil, fmt.Errorf("failed to get activity log: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"activity_id":  activity.ID,
		"action":       activity.Action,
		"resource_type": activity.ResourceType,
	}).Debug("Activity log retrieved successfully")

	return &activity, nil
}

// GetByUserID retrieves activity log entries by user ID with pagination
func (r *activityRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.ActivityLog, error) {
	r.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"limit":   limit,
		"offset":  offset,
	}).Debug("Getting activity logs by user ID")

	var activities []*model.ActivityLog
	if err := r.db.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&activities).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err,
		}).Error("Failed to get activity logs by user ID")
		return nil, fmt.Errorf("failed to get activity logs: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"count":   len(activities),
	}).Debug("Activity logs retrieved successfully by user ID")

	return activities, nil
}

// GetByTenantID retrieves activity log entries by tenant ID with pagination
func (r *activityRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*model.ActivityLog, error) {
	r.logger.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"limit":     limit,
		"offset":    offset,
	}).Debug("Getting activity logs by tenant ID")

	var activities []*model.ActivityLog
	if err := r.db.DB.WithContext(ctx).
		Preload("User").
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&activities).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"tenant_id": tenantID,
			"error":     err,
		}).Error("Failed to get activity logs by tenant ID")
		return nil, fmt.Errorf("failed to get activity logs: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"count":     len(activities),
	}).Debug("Activity logs retrieved successfully by tenant ID")

	return activities, nil
}

// GetByResource retrieves activity log entries by resource type and ID with pagination
func (r *activityRepository) GetByResource(ctx context.Context, resourceType string, resourceID uuid.UUID, limit, offset int) ([]*model.ActivityLog, error) {
	r.logger.WithFields(logrus.Fields{
		"resource_type": resourceType,
		"resource_id":   resourceID,
		"limit":         limit,
		"offset":        offset,
	}).Debug("Getting activity logs by resource")

	var activities []*model.ActivityLog
	if err := r.db.DB.WithContext(ctx).
		Preload("User").
		Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&activities).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"resource_type": resourceType,
			"resource_id":   resourceID,
			"error":         err,
		}).Error("Failed to get activity logs by resource")
		return nil, fmt.Errorf("failed to get activity logs: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"resource_type": resourceType,
		"resource_id":   resourceID,
		"count":         len(activities),
	}).Debug("Activity logs retrieved successfully by resource")

	return activities, nil
}

// SearchActivities searches activity log entries with filters
func (r *activityRepository) SearchActivities(ctx context.Context, tenantID uuid.UUID, filters ActivityFilters, limit, offset int) ([]*model.ActivityLog, error) {
	r.logger.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"filters":   filters,
		"limit":     limit,
		"offset":    offset,
	}).Debug("Searching activity logs")

	query := r.db.DB.WithContext(ctx).Where("tenant_id = ?", tenantID)

	// Apply filters
	if filters.Action != "" {
		query = query.Where("action = ?", filters.Action)
	}
	if filters.ResourceType != "" {
		query = query.Where("resource_type = ?", filters.ResourceType)
	}
	if filters.UserID != nil {
		query = query.Where("user_id = ?", *filters.UserID)
	}
	if filters.Success != nil {
		query = query.Where("success = ?", *filters.Success)
	}
	if filters.StartDate != nil {
		query = query.Where("created_at >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("created_at <= ?", *filters.EndDate)
	}
	if filters.IPAddress != "" {
		query = query.Where("ip_address = ?", filters.IPAddress)
	}
	if filters.SessionID != "" {
		query = query.Where("session_id = ?", filters.SessionID)
	}

	var activities []*model.ActivityLog
	if err := query.
		Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&activities).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"tenant_id": tenantID,
			"filters":   filters,
			"error":     err,
		}).Error("Failed to search activity logs")
		return nil, fmt.Errorf("failed to search activity logs: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"filters":   filters,
		"count":     len(activities),
	}).Debug("Activity logs searched successfully")

	return activities, nil
}

// CountByTenant counts activity log entries by tenant ID
func (r *activityRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	r.logger.WithField("tenant_id", tenantID).Debug("Counting activity logs by tenant ID")

	var count int64
	if err := r.db.DB.WithContext(ctx).
		Model(&model.ActivityLog{}).
		Where("tenant_id = ?", tenantID).
		Count(&count).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"tenant_id": tenantID,
			"error":     err,
		}).Error("Failed to count activity logs by tenant ID")
		return 0, fmt.Errorf("failed to count activity logs: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"count":     count,
	}).Debug("Activity logs counted successfully by tenant ID")

	return count, nil
}

// CountByAction counts activity log entries by tenant ID and action
func (r *activityRepository) CountByAction(ctx context.Context, tenantID uuid.UUID, action string) (int64, error) {
	r.logger.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"action":    action,
	}).Debug("Counting activity logs by tenant ID and action")

	var count int64
	if err := r.db.DB.WithContext(ctx).
		Model(&model.ActivityLog{}).
		Where("tenant_id = ? AND action = ?", tenantID, action).
		Count(&count).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"tenant_id": tenantID,
			"action":    action,
			"error":     err,
		}).Error("Failed to count activity logs by tenant ID and action")
		return 0, fmt.Errorf("failed to count activity logs: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"action":    action,
		"count":     count,
	}).Debug("Activity logs counted successfully by tenant ID and action")

	return count, nil
}

// DeleteOldActivities deletes old activity log entries
func (r *activityRepository) DeleteOldActivities(ctx context.Context, olderThan time.Time) (int64, error) {
	r.logger.WithField("older_than", olderThan).Debug("Deleting old activity logs")

	result := r.db.DB.WithContext(ctx).
		Where("created_at < ?", olderThan).
		Delete(&model.ActivityLog{})

	if result.Error != nil {
		r.logger.WithFields(logrus.Fields{
			"older_than": olderThan,
			"error":      result.Error,
		}).Error("Failed to delete old activity logs")
		return 0, fmt.Errorf("failed to delete old activity logs: %w", result.Error)
	}

	r.logger.WithFields(logrus.Fields{
		"older_than":    olderThan,
		"deleted_count": result.RowsAffected,
	}).Info("Old activity logs deleted successfully")

	return result.RowsAffected, nil
}