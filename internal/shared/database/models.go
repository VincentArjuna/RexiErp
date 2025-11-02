package database

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// BaseModel provides common fields for all models
type BaseModel struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID  uuid.UUID  `gorm:"type:uuid;not null;index;default:gen_random_uuid()" json:"tenant_id"`
	CreatedAt time.Time `gorm:"not null;index" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate sets the created_at and updated_at timestamps
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	if b.CreatedAt.IsZero() {
		b.CreatedAt = now
	}
	if b.UpdatedAt.IsZero() {
		b.UpdatedAt = now
	}
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// BeforeUpdate updates the updated_at timestamp
func (b *BaseModel) BeforeUpdate(tx *gorm.DB) error {
	b.UpdatedAt = time.Now()
	return nil
}

// JSONB is a custom type for handling JSONB fields in PostgreSQL
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, &j)
	case string:
		return json.Unmarshal([]byte(v), &j)
	default:
		return fmt.Errorf("cannot scan %T into JSONB", value)
	}
}

// StringArray is a custom type for handling string arrays
type StringArray []string

// Value implements the driver.Valuer interface
func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return s, nil
}

// Scan implements the sql.Scanner interface
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, &s)
	case string:
		return json.Unmarshal([]byte(v), &s)
	default:
		return fmt.Errorf("cannot scan %T into StringArray", value)
	}
}

// Status represents the status of a record
type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusPending  Status = "pending"
	StatusSuspended Status = "suspended"
	StatusArchived Status = "archived"
)

// AuditLog represents an audit trail entry
type AuditLog struct {
	BaseModel
	Action      string                 `gorm:"not null;index" json:"action"`
	TableName   string                 `gorm:"not null;index" json:"table_name"`
	RecordID    uuid.UUID              `gorm:"type:uuid;not null;index" json:"record_id"`
	OldValues   JSONB                  `gorm:"type:jsonb" json:"old_values,omitempty"`
	NewValues   JSONB                  `gorm:"type:jsonb" json:"new_values,omitempty"`
	UserID      uuid.UUID              `gorm:"type:uuid;index" json:"user_id,omitempty"`
	UserAgent   string                 `gorm:"size:500" json:"user_agent,omitempty"`
	IPAddress   string                 `gorm:"size:45" json:"ip_address,omitempty"`
	Metadata    JSONB                  `gorm:"type:jsonb" json:"metadata,omitempty"`
}

// Config represents system configuration
type Config struct {
	BaseModel
	Key         string `gorm:"not null;uniqueIndex" json:"key"`
	Value       string `gorm:"type:text" json:"value"`
	Description string `gorm:"type:text" json:"description"`
	Category    string `gorm:"index" json:"category"`
	IsPublic    bool   `gorm:"default:false" json:"is_public"`
}

// Notification represents a system notification
type Notification struct {
	BaseModel
	Title       string    `gorm:"not null" json:"title"`
	Message     string    `gorm:"type:text;not null" json:"message"`
	Type        string    `gorm:"not null;index" json:"type"`
	Priority    string    `gorm:"default:'medium';index" json:"priority"`
	TargetType  string    `gorm:"not null;index" json:"target_type"`
	TargetID    uuid.UUID `gorm:"type:uuid;index" json:"target_id,omitempty"`
	UserID      uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	IsRead      bool      `gorm:"default:false;index" json:"is_read"`
	ReadAt      *time.Time `json:"read_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// File represents a file record
type File struct {
	BaseModel
	Name        string    `gorm:"not null" json:"name"`
	OriginalName string    `gorm:"not null" json:"original_name"`
	ContentType string    `gorm:"not null" json:"content_type"`
	Size        int64     `gorm:"not null" json:"size"`
	Path        string    `gorm:"not null" json:"path"`
	Hash        string    `gorm:"size:64;index" json:"hash"`
	StorageType string    `gorm:"not null" json:"storage_type"`
	IsPublic    bool      `gorm:"default:false" json:"is_public"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// Common validation constants
const (
	// Email validation
	EmailMaxLength = 255

	// Phone validation
	PhoneMaxLength = 20

	// Name validation
	NameMinLength = 1
	NameMaxLength = 100

	// Description validation
	DescriptionMaxLength = 1000

	// Code validation
	CodeMaxLength = 50

	// URL validation
	URLMaxLength = 2048

	// Address validation
	AddressMaxLength = 500
)

// Common business rules
const (
	// Pagination
	DefaultPageSize = 20
	MaxPageSize     = 100

	// File upload
	MaxFileSize = 50 * 1024 * 1024 // 50MB

	// Password requirements
	MinPasswordLength = 8
	MaxPasswordLength = 128

	// Session management
	DefaultSessionDuration = 24 * time.Hour
	MaxSessionDuration     = 30 * 24 * time.Hour // 30 days

	// Rate limiting
	MaxLoginAttempts      = 5
	LoginAttemptWindow    = 15 * time.Minute
	PasswordResetDuration = 1 * time.Hour

	// Data retention
	AuditLogRetentionDays = 365
	NotificationRetentionDays = 30
	TemporaryFileRetentionHours = 24
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// Error implements the error interface
func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation error on field %s: %s", ve.Field, ve.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (ves ValidationErrors) Error() string {
	if len(ves) == 0 {
		return "no validation errors"
	}
	if len(ves) == 1 {
		return ves[0].Error()
	}
	return fmt.Sprintf("multiple validation errors: %d errors", len(ves))
}

// Add adds a validation error
func (ves *ValidationErrors) Add(field, message string, value interface{}) {
	*ves = append(*ves, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// HasErrors returns true if there are validation errors
func (ves ValidationErrors) HasErrors() bool {
	return len(ves) > 0
}

// Validator provides common validation functions
type Validator struct {
	errors ValidationErrors
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		errors: make(ValidationErrors, 0),
	}
}

// Required validates that a field is not empty
func (v *Validator) Required(field string, value interface{}) {
	if value == nil || value == "" {
		v.errors.Add(field, "is required", value)
	}
}

// MinLength validates minimum length
func (v *Validator) MinLength(field string, value string, min int) {
	if len(value) < min {
		v.errors.Add(field, fmt.Sprintf("must be at least %d characters", min), value)
	}
}

// MaxLength validates maximum length
func (v *Validator) MaxLength(field string, value string, max int) {
	if len(value) > max {
		v.errors.Add(field, fmt.Sprintf("must be at most %d characters", max), value)
	}
}

// Email validates email format
func (v *Validator) Email(field string, value string) {
	// Simple email validation - in production, use a proper email validation library
	if value == "" {
		return
	}

	if len(value) > EmailMaxLength || !strings.Contains(value, "@") || !strings.Contains(value, ".") {
		v.errors.Add(field, "must be a valid email address", value)
	}
}

// UUID validates UUID format
func (v *Validator) UUID(field string, value string) {
	if value == "" {
		return
	}

	if _, err := uuid.Parse(value); err != nil {
		v.errors.Add(field, "must be a valid UUID", value)
	}
}

// OneOf validates that value is one of the allowed values
func (v *Validator) OneOf(field string, value string, allowed []string) {
	if value == "" {
		return
	}

	for _, allowedValue := range allowed {
		if value == allowedValue {
			return
		}
	}

	v.errors.Add(field, fmt.Sprintf("must be one of: %v", allowed), value)
}

// Range validates numeric range
func (v *Validator) Range(field string, value int64, min, max int64) {
	if value < min || value > max {
		v.errors.Add(field, fmt.Sprintf("must be between %d and %d", min, max), value)
	}
}

// Positive validates that a number is positive
func (v *Validator) Positive(field string, value int64) {
	if value <= 0 {
		v.errors.Add(field, "must be positive", value)
	}
}

// NonNegative validates that a number is non-negative
func (v *Validator) NonNegative(field string, value int64) {
	if value < 0 {
		v.errors.Add(field, "must be non-negative", value)
	}
}

// FutureDate validates that a date is in the future
func (v *Validator) FutureDate(field string, value time.Time) {
	if !value.IsZero() && value.Before(time.Now()) {
		v.errors.Add(field, "must be in the future", value)
	}
}

// PastDate validates that a date is in the past
func (v *Validator) PastDate(field string, value time.Time) {
	if !value.IsZero() && value.After(time.Now()) {
		v.errors.Add(field, "must be in the past", value)
	}
}

// Errors returns all validation errors
func (v *Validator) Errors() ValidationErrors {
	return v.errors
}

// HasErrors returns true if there are validation errors
func (v *Validator) HasErrors() bool {
	return v.errors.HasErrors()
}

// ToError converts validation errors to a single error
func (v *Validator) ToError() error {
	if !v.HasErrors() {
		return nil
	}
	return v.errors
}

// Database utility functions

// GenerateUUID generates a new UUID
func GenerateUUID() uuid.UUID {
	return uuid.New()
}

// ParseUUID parses a UUID string
func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// IsValidUUID checks if a string is a valid UUID
func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// NowUTC returns current time in UTC
func NowUTC() time.Time {
	return time.Now().UTC()
}

// FormatTimestamp formats a timestamp for database storage
func FormatTimestamp(t time.Time) time.Time {
	return t.UTC()
}

// ParseTimestamp parses a timestamp from database storage
func ParseTimestamp(t time.Time) time.Time {
	return t.UTC()
}

// CalculateHash calculates a simple hash for file integrity
func CalculateHash(content string) string {
	// Simple hash implementation - in production, use SHA-256
	return fmt.Sprintf("%x", len(content))
}

// SanitizeInput sanitizes user input for database storage
func SanitizeInput(input string) string {
	// Simple sanitization - in production, use proper input sanitization
	if input == "" {
		return ""
	}

	// Trim whitespace
	input = strings.TrimSpace(input)

	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	return input
}

// IsValidJSON checks if a string is valid JSON
func IsValidJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

// MaskSensitiveData masks sensitive data for logging
func MaskSensitiveData(data string) string {
	if len(data) <= 4 {
		return "****"
	}
	return data[:2] + "****" + data[len(data)-2:]
}

// LogDatabaseError logs database errors with context
func LogDatabaseError(logger *logrus.Logger, operation string, err error, context map[string]interface{}) {
	if logger == nil {
		return
	}

	fields := logrus.Fields{
		"operation": operation,
		"error":     err.Error(),
	}

	for k, v := range context {
		fields[k] = v
	}

	logger.WithFields(fields).Error("Database operation failed")
}