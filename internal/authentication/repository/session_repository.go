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

// SessionRepository interface defines the contract for session data operations
type SessionRepository interface {
	Create(ctx context.Context, session *model.UserSession) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.UserSession, error)
	GetBySessionID(ctx context.Context, sessionID string) (*model.UserSession, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, activeOnly bool) ([]*model.UserSession, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*model.UserSession, error)
	Update(ctx context.Context, session *model.UserSession) error
	UpdateActivity(ctx context.Context, sessionID string) error
	Deactivate(ctx context.Context, sessionID string) error
	DeactivateByUserID(ctx context.Context, userID uuid.UUID) error
	Delete(ctx context.Context, sessionID string) error
	CleanupExpiredSessions(ctx context.Context) (int64, error)
	GetActiveSessionsCount(ctx context.Context, userID uuid.UUID) (int64, error)
}

// sessionRepository implements SessionRepository interface
type sessionRepository struct {
	db     *database.Database
	logger *logrus.Logger
}

// NewSessionRepository creates a new instance of SessionRepository
func NewSessionRepository(db *database.Database, logger *logrus.Logger) SessionRepository {
	return &sessionRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new session
func (r *sessionRepository) Create(ctx context.Context, session *model.UserSession) error {
	r.logger.WithFields(logrus.Fields{
		"session_id": session.SessionID,
		"user_id":    session.UserID,
		"tenant_id":  session.TenantID,
	}).Debug("Creating new session")

	if err := r.db.DB.WithContext(ctx).Create(session).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"session_id": session.SessionID,
			"user_id":    session.UserID,
			"tenant_id":  session.TenantID,
			"error":      err,
		}).Error("Failed to create session")
		return fmt.Errorf("failed to create session: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"session_id": session.SessionID,
		"user_id":    session.UserID,
		"tenant_id":  session.TenantID,
	}).Info("Session created successfully")

	return nil
}

// GetByID retrieves a session by ID
func (r *sessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.UserSession, error) {
	r.logger.WithField("session_id", id).Debug("Getting session by ID")

	var session model.UserSession
	if err := r.db.DB.WithContext(ctx).
		Preload("User").
		Where("id = ?", id).
		First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WithField("session_id", id).Debug("Session not found")
			return nil, fmt.Errorf("session not found")
		}
		r.logger.WithFields(logrus.Fields{
			"session_id": id,
			"error":      err,
		}).Error("Failed to get session by ID")
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"session_id": session.SessionID,
		"user_id":    session.UserID,
		"tenant_id":  session.TenantID,
	}).Debug("Session retrieved successfully")

	return &session, nil
}

// GetBySessionID retrieves a session by session ID
func (r *sessionRepository) GetBySessionID(ctx context.Context, sessionID string) (*model.UserSession, error) {
	r.logger.WithField("session_id", sessionID).Debug("Getting session by session ID")

	var session model.UserSession
	if err := r.db.DB.WithContext(ctx).
		Preload("User").
		Where("session_id = ?", sessionID).
		First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WithField("session_id", sessionID).Debug("Session not found by session ID")
			return nil, fmt.Errorf("session not found")
		}
		r.logger.WithFields(logrus.Fields{
			"session_id": sessionID,
			"error":      err,
		}).Error("Failed to get session by session ID")
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"session_id": session.SessionID,
		"user_id":    session.UserID,
		"tenant_id":  session.TenantID,
	}).Debug("Session retrieved successfully by session ID")

	return &session, nil
}

// GetByUserID retrieves sessions by user ID
func (r *sessionRepository) GetByUserID(ctx context.Context, userID uuid.UUID, activeOnly bool) ([]*model.UserSession, error) {
	r.logger.WithFields(logrus.Fields{
		"user_id":      userID,
		"active_only":  activeOnly,
	}).Debug("Getting sessions by user ID")

	var sessions []*model.UserSession
	query := r.db.DB.WithContext(ctx).Where("user_id = ?", userID)

	if activeOnly {
		query = query.Where("is_active = ? AND expires_at > ?", true, time.Now())
	}

	if err := query.
		Preload("User").
		Order("created_at DESC").
		Find(&sessions).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err,
		}).Error("Failed to get sessions by user ID")
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"count":   len(sessions),
	}).Debug("Sessions retrieved successfully by user ID")

	return sessions, nil
}

// GetByTokenHash retrieves a session by token hash
func (r *sessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*model.UserSession, error) {
	r.logger.WithField("token_hash", tokenHash[:8]+"...").Debug("Getting session by token hash")

	var session model.UserSession
	if err := r.db.DB.WithContext(ctx).
		Preload("User").
		Where("token_hash = ? AND is_active = ? AND expires_at > ?", tokenHash, true, time.Now()).
		First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WithField("token_hash", tokenHash[:8]+"...").Debug("Session not found by token hash")
			return nil, fmt.Errorf("session not found")
		}
		r.logger.WithFields(logrus.Fields{
			"token_hash": tokenHash[:8] + "...",
			"error":      err,
		}).Error("Failed to get session by token hash")
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"session_id": session.SessionID,
		"user_id":    session.UserID,
		"tenant_id":  session.TenantID,
	}).Debug("Session retrieved successfully by token hash")

	return &session, nil
}

// Update updates a session
func (r *sessionRepository) Update(ctx context.Context, session *model.UserSession) error {
	r.logger.WithFields(logrus.Fields{
		"session_id": session.SessionID,
		"user_id":    session.UserID,
	}).Debug("Updating session")

	if err := r.db.DB.WithContext(ctx).Save(session).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"session_id": session.SessionID,
			"user_id":    session.UserID,
			"error":      err,
		}).Error("Failed to update session")
		return fmt.Errorf("failed to update session: %w", err)
	}

	r.logger.WithField("session_id", session.SessionID).Debug("Session updated successfully")
	return nil
}

// UpdateActivity updates the last activity timestamp for a session
func (r *sessionRepository) UpdateActivity(ctx context.Context, sessionID string) error {
	r.logger.WithField("session_id", sessionID).Debug("Updating session activity")

	if err := r.db.DB.WithContext(ctx).
		Model(&model.UserSession{}).
		Where("session_id = ?", sessionID).
		Update("last_activity", time.Now()).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"session_id": sessionID,
			"error":      err,
		}).Error("Failed to update session activity")
		return fmt.Errorf("failed to update session activity: %w", err)
	}

	r.logger.WithField("session_id", sessionID).Debug("Session activity updated successfully")
	return nil
}

// Deactivate deactivates a session
func (r *sessionRepository) Deactivate(ctx context.Context, sessionID string) error {
	r.logger.WithField("session_id", sessionID).Debug("Deactivating session")

	if err := r.db.DB.WithContext(ctx).
		Model(&model.UserSession{}).
		Where("session_id = ?", sessionID).
		Update("is_active", false).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"session_id": sessionID,
			"error":      err,
		}).Error("Failed to deactivate session")
		return fmt.Errorf("failed to deactivate session: %w", err)
	}

	r.logger.WithField("session_id", sessionID).Info("Session deactivated successfully")
	return nil
}

// DeactivateByUserID deactivates all sessions for a user
func (r *sessionRepository) DeactivateByUserID(ctx context.Context, userID uuid.UUID) error {
	r.logger.WithField("user_id", userID).Debug("Deactivating all sessions for user")

	result := r.db.DB.WithContext(ctx).
		Model(&model.UserSession{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Update("is_active", false)

	if result.Error != nil {
		r.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   result.Error,
		}).Error("Failed to deactivate sessions by user ID")
		return fmt.Errorf("failed to deactivate sessions: %w", result.Error)
	}

	r.logger.WithFields(logrus.Fields{
		"user_id":     userID,
		"deactivated": result.RowsAffected,
	}).Info("Sessions deactivated successfully by user ID")

	return nil
}

// Delete deletes a session
func (r *sessionRepository) Delete(ctx context.Context, sessionID string) error {
	r.logger.WithField("session_id", sessionID).Debug("Deleting session")

	if err := r.db.DB.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Delete(&model.UserSession{}).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"session_id": sessionID,
			"error":      err,
		}).Error("Failed to delete session")
		return fmt.Errorf("failed to delete session: %w", err)
	}

	r.logger.WithField("session_id", sessionID).Info("Session deleted successfully")
	return nil
}

// CleanupExpiredSessions removes expired sessions from the database
func (r *sessionRepository) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	r.logger.Debug("Cleaning up expired sessions")

	result := r.db.DB.WithContext(ctx).
		Where("expires_at < ? OR (is_active = ? AND last_activity < ?)",
			time.Now(), false, time.Now().Add(-7*24*time.Hour)).
		Delete(&model.UserSession{})

	if result.Error != nil {
		r.logger.WithField("error", result.Error).Error("Failed to cleanup expired sessions")
		return 0, fmt.Errorf("failed to cleanup expired sessions: %w", result.Error)
	}

	r.logger.WithField("cleaned_count", result.RowsAffected).Info("Expired sessions cleaned up successfully")
	return result.RowsAffected, nil
}

// GetActiveSessionsCount counts active sessions for a user
func (r *sessionRepository) GetActiveSessionsCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	r.logger.WithField("user_id", userID).Debug("Counting active sessions for user")

	var count int64
	if err := r.db.DB.WithContext(ctx).
		Model(&model.UserSession{}).
		Where("user_id = ? AND is_active = ? AND expires_at > ?", userID, true, time.Now()).
		Count(&count).Error; err != nil {
		r.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err,
		}).Error("Failed to count active sessions")
		return 0, fmt.Errorf("failed to count active sessions: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"count":   count,
	}).Debug("Active sessions counted successfully")

	return count, nil
}