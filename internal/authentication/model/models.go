package model

import (
	"github.com/VincentArjuna/RexiErp/internal/shared/database"
)

// AutoMigrate runs auto migration for all authentication models
func AutoMigrate(db *database.Database) error {
	return db.DB.AutoMigrate(
		&User{},
		&UserSession{},
		&ActivityLog{},
		&PasswordResetToken{},
	)
}

// ModelValidationErrors represents collection of validation errors
type ModelValidationErrors struct {
	Errors map[string]string `json:"errors"`
}

// AddError adds a validation error
func (m *ModelValidationErrors) AddError(field, message string) {
	if m.Errors == nil {
		m.Errors = make(map[string]string)
	}
	m.Errors[field] = message
}

// HasErrors returns true if there are validation errors
func (m *ModelValidationErrors) HasErrors() bool {
	return len(m.Errors) > 0
}

// ToError converts validation errors to a single error message
func (m *ModelValidationErrors) ToError() error {
	if !m.HasErrors() {
		return nil
	}
	return &ModelValidationErrors{Errors: m.Errors}
}

// Error implements the error interface
func (m *ModelValidationErrors) Error() string {
	// Return a formatted string of all validation errors
	return "model validation failed"
}