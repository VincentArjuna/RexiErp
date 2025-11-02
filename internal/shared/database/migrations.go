package database

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Migration represents a database migration
type Migration struct {
	Version      string                 `gorm:"primaryKey;column:version"`
	Description  string                 `gorm:"column:description"`
	Applied      bool                   `gorm:"column:applied"`
	AppliedAt    *time.Time             `gorm:"column:applied_at"`
	Checksum     string                 `gorm:"column:checksum"`
	Dependencies []string               `gorm:"column:dependencies;serializer:json"`
	Metadata     map[string]interface{} `gorm:"column:metadata;serializer:json"`
	CreatedAt    time.Time              `gorm:"column:created_at"`
	UpdatedAt    time.Time              `gorm:"column:updated_at"`
}

// MigrationFile represents a migration file on disk
type MigrationFile struct {
	Version     string
	Description string
	UpSQL       string
	DownSQL     string
	Checksum    string
	Path        string
	Dependencies []string
}

// MigrationHook represents a pre or post migration hook
type MigrationHook func(ctx context.Context, db *gorm.DB, migration *Migration) error

// MigrationManager manages database migrations
type MigrationManager struct {
	db              *gorm.DB
	logger          *logrus.Logger
	migrationsPath  string
	migrations      map[string]*MigrationFile
	preHooks        map[string][]MigrationHook
	postHooks       map[string][]MigrationHook
	appliedVersions map[string]bool
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *gorm.DB, logger *logrus.Logger, migrationsPath string) *MigrationManager {
	return &MigrationManager{
		db:              db,
		logger:          logger,
		migrationsPath:  migrationsPath,
		migrations:      make(map[string]*MigrationFile),
		preHooks:        make(map[string][]MigrationHook),
		postHooks:       make(map[string][]MigrationHook),
		appliedVersions: make(map[string]bool),
	}
}

// Initialize creates the migrations table and loads available migrations
func (mm *MigrationManager) Initialize(ctx context.Context) error {
	// Create migrations table
	if err := mm.db.WithContext(ctx).AutoMigrate(&Migration{}); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Load applied migrations
	if err := mm.loadAppliedMigrations(ctx); err != nil {
		return fmt.Errorf("failed to load applied migrations: %w", err)
	}

	// Scan for migration files
	if err := mm.scanMigrationFiles(); err != nil {
		return fmt.Errorf("failed to scan migration files: %w", err)
	}

	mm.logger.WithField("count", len(mm.migrations)).Info("Migration manager initialized")

	return nil
}

// loadAppliedMigrations loads migrations that have already been applied
func (mm *MigrationManager) loadAppliedMigrations(ctx context.Context) error {
	var migrations []Migration
	if err := mm.db.WithContext(ctx).Where("applied = ?", true).Find(&migrations).Error; err != nil {
		return fmt.Errorf("failed to query applied migrations: %w", err)
	}

	for _, migration := range migrations {
		mm.appliedVersions[migration.Version] = true
	}

	return nil
}

// scanMigrationFiles scans the migrations directory for migration files
func (mm *MigrationManager) scanMigrationFiles() error {
	entries, err := os.ReadDir(mm.migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Scan subdirectories (e.g., master/, tenants/)
			if err := mm.scanMigrationSubdirectory(filepath.Join(mm.migrationsPath, entry.Name())); err != nil {
				mm.logger.WithError(err).WithField("subdir", entry.Name()).Warn("Failed to scan migration subdirectory")
			}
		} else {
			// Scan root directory files
			if err := mm.processMigrationFile(filepath.Join(mm.migrationsPath, entry.Name())); err != nil {
				mm.logger.WithError(err).WithField("file", entry.Name()).Warn("Failed to process migration file")
			}
		}
	}

	return nil
}

// scanMigrationSubdirectory scans a subdirectory for migration files
func (mm *MigrationManager) scanMigrationSubdirectory(subdir string) error {
	entries, err := os.ReadDir(subdir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			if err := mm.processMigrationFile(filepath.Join(subdir, entry.Name())); err != nil {
				mm.logger.WithError(err).WithField("file", entry.Name()).Warn("Failed to process migration file")
			}
		}
	}

	return nil
}

// processMigrationFile processes a single migration file
func (mm *MigrationManager) processMigrationFile(filePath string) error {
	filename := filepath.Base(filePath)

	// Check if it's an up migration file
	if !strings.HasSuffix(filename, ".up.sql") {
		return nil
	}

	// Extract version and description from filename
	version, description, err := mm.parseMigrationFilename(filename)
	if err != nil {
		return fmt.Errorf("invalid migration filename %s: %w", filename, err)
	}

	// Check if migration already loaded
	if _, exists := mm.migrations[version]; exists {
		return nil
	}

	// Read up migration file
	upSQL, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read up migration file %s: %w", filePath, err)
	}

	// Read down migration file
	downFilePath := strings.Replace(filePath, ".up.sql", ".down.sql", 1)
	var downSQL []byte
	if _, err := os.Stat(downFilePath); err == nil {
		downSQL, err = os.ReadFile(downFilePath)
		if err != nil {
			return fmt.Errorf("failed to read down migration file %s: %w", downFilePath, err)
		}
	}

	// Calculate checksum
	checksum := mm.calculateChecksum(string(upSQL), string(downSQL))

	// Parse dependencies from SQL comments
	dependencies := mm.parseDependencies(string(upSQL))

	migration := &MigrationFile{
		Version:      version,
		Description:  description,
		UpSQL:        string(upSQL),
		DownSQL:      string(downSQL),
		Checksum:     checksum,
		Path:         filePath,
		Dependencies: dependencies,
	}

	mm.migrations[version] = migration

	return nil
}

// parseMigrationFilename extracts version and description from filename
func (mm *MigrationManager) parseMigrationFilename(filename string) (string, string, error) {
	// Remove .up.sql extension
	name := strings.TrimSuffix(filename, ".up.sql")

	// Split by underscore to extract version
	parts := strings.SplitN(name, "_", 2)
	if len(parts) < 2 {
		return "", "", fmt.Errorf("migration filename must follow pattern: version_description.up.sql")
	}

	version := parts[0]
	description := strings.ReplaceAll(parts[1], "_", " ")

	return version, description, nil
}

// calculateChecksum calculates a SHA-256 checksum for migration files
func (mm *MigrationManager) calculateChecksum(upSQL, downSQL string) string {
	// Use SHA-256 for proper migration integrity validation
	content := upSQL + "|" + downSQL
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}

// parseDependencies extracts dependencies from SQL comments
func (mm *MigrationManager) parseDependencies(sql string) []string {
	var dependencies []string
	lines := strings.Split(sql, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "-- depends:") {
			depStr := strings.TrimPrefix(line, "-- depends:")
			depStr = strings.TrimSpace(depStr)
			if depStr != "" {
				deps := strings.Split(depStr, ",")
				for _, dep := range deps {
					dependencies = append(dependencies, strings.TrimSpace(dep))
				}
			}
		}
	}

	return dependencies
}

// GetPendingMigrations returns migrations that haven't been applied yet
func (mm *MigrationManager) GetPendingMigrations() ([]*MigrationFile, error) {
	var pending []*MigrationFile

	for version, migration := range mm.migrations {
		if !mm.appliedVersions[version] {
			pending = append(pending, migration)
		}
	}

	// Sort by version
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].Version < pending[j].Version
	})

	return pending, nil
}

// GetAppliedMigrations returns migrations that have been applied
func (mm *MigrationManager) GetAppliedMigrations() ([]*Migration, error) {
	var applied []*Migration

	for version := range mm.appliedVersions {
		var migration Migration
		if err := mm.db.Where("version = ?", version).First(&migration).Error; err == nil {
			applied = append(applied, &migration)
		}
	}

	// Sort by version
	sort.Slice(applied, func(i, j int) bool {
		return applied[i].Version < applied[j].Version
	})

	return applied, nil
}

// MigrateUp applies all pending migrations
func (mm *MigrationManager) MigrateUp(ctx context.Context) error {
	pending, err := mm.GetPendingMigrations()
	if err != nil {
		return fmt.Errorf("failed to get pending migrations: %w", err)
	}

	if len(pending) == 0 {
		mm.logger.Info("No pending migrations to apply")
		return nil
	}

	mm.logger.WithField("count", len(pending)).Info("Starting migration up")

	for _, migration := range pending {
		if err := mm.applyMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
		}
	}

	mm.logger.Info("All migrations applied successfully")
	return nil
}

// MigrateDown rolls back the last applied migration
func (mm *MigrationManager) MigrateDown(ctx context.Context, targetVersion string) error {
	if targetVersion == "" {
		// Get the last applied migration
		applied, err := mm.GetAppliedMigrations()
		if err != nil {
			return fmt.Errorf("failed to get applied migrations: %w", err)
		}

		if len(applied) == 0 {
			mm.logger.Info("No migrations to rollback")
			return nil
		}

		targetVersion = applied[len(applied)-1].Version
	}

	migration, exists := mm.migrations[targetVersion]
	if !exists {
		return fmt.Errorf("migration %s not found", targetVersion)
	}

	if !mm.appliedVersions[targetVersion] {
		return fmt.Errorf("migration %s has not been applied", targetVersion)
	}

	mm.logger.WithField("version", targetVersion).Info("Starting migration down")

	if err := mm.rollbackMigration(ctx, migration); err != nil {
		return fmt.Errorf("failed to rollback migration %s: %w", targetVersion, err)
	}

	mm.logger.WithField("version", targetVersion).Info("Migration rolled back successfully")
	return nil
}

// applyMigration applies a single migration
func (mm *MigrationManager) applyMigration(ctx context.Context, migration *MigrationFile) error {
	// Check dependencies
	if err := mm.checkDependencies(ctx, migration); err != nil {
		return fmt.Errorf("dependency check failed: %w", err)
	}

	// Execute pre-migration hooks
	if err := mm.executeHooks(ctx, migration, mm.preHooks); err != nil {
		return fmt.Errorf("pre-migration hooks failed: %w", err)
	}

	// Start transaction
	tx := mm.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin migration transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// Execute migration SQL
	if err := tx.Exec(migration.UpSQL).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration
	now := time.Now()
	migrationRecord := Migration{
		Version:     migration.Version,
		Description: migration.Description,
		Applied:     true,
		AppliedAt:   &now,
		Checksum:    migration.Checksum,
		Dependencies: migration.Dependencies,
		Metadata:    make(map[string]interface{}),
	}

	if err := tx.Create(&migrationRecord).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit migration transaction: %w", err)
	}

	// Update applied versions
	mm.appliedVersions[migration.Version] = true

	// Execute post-migration hooks
	if err := mm.executeHooks(ctx, migration, mm.postHooks); err != nil {
		mm.logger.WithError(err).Warn("Post-migration hooks failed")
	}

	mm.logger.WithFields(logrus.Fields{
		"version":    migration.Version,
		"description": migration.Description,
	}).Info("Migration applied successfully")

	return nil
}

// rollbackMigration rolls back a single migration
func (mm *MigrationManager) rollbackMigration(ctx context.Context, migration *MigrationFile) error {
	if migration.DownSQL == "" {
		return fmt.Errorf("migration %s does not have a rollback script", migration.Version)
	}

	// Execute pre-migration hooks
	if err := mm.executeHooks(ctx, migration, mm.preHooks); err != nil {
		return fmt.Errorf("pre-rollback hooks failed: %w", err)
	}

	// Start transaction
	tx := mm.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin rollback transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// Execute rollback SQL
	if err := tx.Exec(migration.DownSQL).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to execute rollback SQL: %w", err)
	}

	// Remove migration record
	if err := tx.Where("version = ?", migration.Version).Delete(&Migration{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit rollback transaction: %w", err)
	}

	// Update applied versions
	delete(mm.appliedVersions, migration.Version)

	// Execute post-migration hooks
	if err := mm.executeHooks(ctx, migration, mm.postHooks); err != nil {
		mm.logger.WithError(err).Warn("Post-rollback hooks failed")
	}

	mm.logger.WithFields(logrus.Fields{
		"version":    migration.Version,
		"description": migration.Description,
	}).Info("Migration rolled back successfully")

	return nil
}

// checkDependencies checks if all dependencies for a migration are satisfied
func (mm *MigrationManager) checkDependencies(ctx context.Context, migration *MigrationFile) error {
	for _, dep := range migration.Dependencies {
		if !mm.appliedVersions[dep] {
			return fmt.Errorf("dependency %s not satisfied for migration %s", dep, migration.Version)
		}
	}
	return nil
}

// executeHooks executes pre or post migration hooks
func (mm *MigrationManager) executeHooks(ctx context.Context, migration *MigrationFile, hooks map[string][]MigrationHook) error {
	migrationHooks, exists := hooks[migration.Version]
	if !exists {
		return nil
	}

	for _, hook := range migrationHooks {
		migrationRecord := &Migration{
			Version:     migration.Version,
			Description: migration.Description,
		}

		if err := hook(ctx, mm.db, migrationRecord); err != nil {
			return fmt.Errorf("hook failed for migration %s: %w", migration.Version, err)
		}
	}

	return nil
}

// AddPreHook adds a pre-migration hook
func (mm *MigrationManager) AddPreHook(version string, hook MigrationHook) {
	mm.preHooks[version] = append(mm.preHooks[version], hook)
}

// AddPostHook adds a post-migration hook
func (mm *MigrationManager) AddPostHook(version string, hook MigrationHook) {
	mm.postHooks[version] = append(mm.postHooks[version], hook)
}

// GetStatus returns the current migration status
func (mm *MigrationManager) GetStatus(ctx context.Context) (*MigrationStatus, error) {
	pending, err := mm.GetPendingMigrations()
	if err != nil {
		return nil, err
	}

	applied, err := mm.GetAppliedMigrations()
	if err != nil {
		return nil, err
	}

	return &MigrationStatus{
		TotalMigrations: len(mm.migrations),
		AppliedMigrations: len(applied),
		PendingMigrations: len(pending),
		Applied: applied,
		Pending: pending,
	}, nil
}

// MigrationStatus represents the current migration status
type MigrationStatus struct {
	TotalMigrations   int            `json:"total_migrations"`
	AppliedMigrations int            `json:"applied_migrations"`
	PendingMigrations int            `json:"pending_migrations"`
	Applied          []*Migration   `json:"applied"`
	Pending          []*MigrationFile `json:"pending"`
}