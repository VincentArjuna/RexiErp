package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ActivityLog represents an activity log entry
type ActivityLog struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       *uuid.UUID `gorm:"type:uuid;index:idx_activity_user" json:"user_id"`
	TenantID     uuid.UUID  `gorm:"type:uuid;not null;index:idx_activity_tenant" json:"tenant_id"`
	Action       string     `gorm:"type:varchar(100);not null;index" json:"action"`
	ResourceType string     `gorm:"type:varchar(100);not null;index" json:"resource_type"`
	ResourceID   *uuid.UUID `gorm:"type:uuid;index:idx_activity_resource" json:"resource_id"`
	OldValues    string     `gorm:"type:json" json:"old_values"`
	NewValues    string     `gorm:"type:json" json:"new_values"`
	IPAddress    string     `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent    string     `gorm:"type:text" json:"user_agent"`
	SessionID    string     `gorm:"type:varchar(255);index" json:"session_id"`
	Success      bool       `gorm:"not null;default:true;index" json:"success"`
	ErrorMessage string     `gorm:"type:text" json:"error_message"`
	Context      string     `gorm:"type:json" json:"context"`
	CreatedAt    time.Time  `gorm:"not null;index" json:"created_at"`

	// Relationships
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL" json:"user,omitempty"`
}

// ActivityContext represents additional context data for an activity
type ActivityContext struct {
	RequestID     string                 `json:"request_id"`
	IPAddress     string                 `json:"ip_address"`
	UserAgent     string                 `json:"user_agent"`
	Metadata      map[string]interface{} `json:"metadata"`
	Component     string                 `json:"component"`
	TraceID       string                 `json:"trace_id"`
	SpanID        string                 `json:"span_id"`
}

// TableName returns the table name for the ActivityLog model
func (ActivityLog) TableName() string {
	return "activity_logs"
}

// BeforeCreate is a GORM hook that runs before creating an activity log
func (al *ActivityLog) BeforeCreate(tx *gorm.DB) error {
	if al.ID == uuid.Nil {
		al.ID = uuid.New()
	}
	return nil
}

// SetOldValues sets the old values as JSON
func (al *ActivityLog) SetOldValues(values interface{}) error {
	if values == nil {
		al.OldValues = ""
		return nil
	}

	jsonData, err := json.Marshal(values)
	if err != nil {
		return err
	}
	al.OldValues = string(jsonData)
	return nil
}

// SetNewValues sets the new values as JSON
func (al *ActivityLog) SetNewValues(values interface{}) error {
	if values == nil {
		al.NewValues = ""
		return nil
	}

	jsonData, err := json.Marshal(values)
	if err != nil {
		return err
	}
	al.NewValues = string(jsonData)
	return nil
}

// GetOldValues returns the old values as a map
func (al *ActivityLog) GetOldValues() (map[string]interface{}, error) {
	if al.OldValues == "" {
		return nil, nil
	}

	var values map[string]interface{}
	err := json.Unmarshal([]byte(al.OldValues), &values)
	if err != nil {
		return nil, err
	}
	return values, nil
}

// GetNewValues returns the new values as a map
func (al *ActivityLog) GetNewValues() (map[string]interface{}, error) {
	if al.NewValues == "" {
		return nil, nil
	}

	var values map[string]interface{}
	err := json.Unmarshal([]byte(al.NewValues), &values)
	if err != nil {
		return nil, err
	}
	return values, nil
}

// SetContext sets the context as JSON
func (al *ActivityLog) SetContext(context ActivityContext) error {
	jsonData, err := json.Marshal(context)
	if err != nil {
		return err
	}
	al.Context = string(jsonData)
	return nil
}

// GetContext returns the context as a struct
func (al *ActivityLog) GetContext() (*ActivityContext, error) {
	if al.Context == "" {
		return nil, nil
	}

	var context ActivityContext
	err := json.Unmarshal([]byte(al.Context), &context)
	if err != nil {
		return nil, err
	}
	return &context, nil
}

// MarkAsFailed marks the activity as failed with an error message
func (al *ActivityLog) MarkAsFailed(errorMessage string) {
	al.Success = false
	al.ErrorMessage = errorMessage
}

// SanitizeForResponse returns an activity log object with sensitive data filtered
func (al *ActivityLog) SanitizeForResponse() *ActivityLog {
	return &ActivityLog{
		ID:           al.ID,
		UserID:       al.UserID,
		TenantID:     al.TenantID,
		Action:       al.Action,
		ResourceType: al.ResourceType,
		ResourceID:   al.ResourceID,
		IPAddress:    al.IPAddress,
		UserAgent:    al.UserAgent,
		SessionID:    al.SessionID,
		Success:      al.Success,
		ErrorMessage: al.ErrorMessage,
		CreatedAt:    al.CreatedAt,
	}
}