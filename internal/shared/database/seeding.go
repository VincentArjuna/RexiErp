package database

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Seeder represents a data seeder
type Seeder struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Environment string                 `json:"environment"`
	Dependencies []string              `json:"dependencies"`
	Data        map[string]interface{} `json:"data"`
	Order       int                    `json:"order"`
}

// SeedManager manages database seeding
type SeedManager struct {
	db             *gorm.DB
	logger         *logrus.Logger
	seedsPath      string
	seeders        map[string]*Seeder
	appliedSeeders map[string]bool
}

// NewSeedManager creates a new seed manager
func NewSeedManager(db *gorm.DB, logger *logrus.Logger, seedsPath string) *SeedManager {
	return &SeedManager{
		db:             db,
		logger:         logger,
		seedsPath:      seedsPath,
		seeders:        make(map[string]*Seeder),
		appliedSeeders: make(map[string]bool),
	}
}

// LoadSeeders loads seeders from the seeds directory
func (sm *SeedManager) LoadSeeders(ctx context.Context) error {
	entries, err := os.ReadDir(sm.seedsPath)
	if err != nil {
		if os.IsNotExist(err) {
			sm.logger.Info("Seeds directory does not exist, skipping seed loading")
			return nil
		}
		return fmt.Errorf("failed to read seeds directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		if err := sm.loadSeederFile(filepath.Join(sm.seedsPath, entry.Name())); err != nil {
			sm.logger.WithError(err).WithField("file", entry.Name()).Warn("Failed to load seeder file")
		}
	}

	sm.logger.WithField("count", len(sm.seeders)).Info("Seeders loaded")

	return nil
}

// loadSeederFile loads a single seeder file
func (sm *SeedManager) loadSeederFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read seeder file %s: %w", filePath, err)
	}

	var seeder Seeder
	if err := json.Unmarshal(data, &seeder); err != nil {
		return fmt.Errorf("failed to unmarshal seeder file %s: %w", filePath, err)
	}

	sm.seeders[seeder.Name] = &seeder

	return nil
}

// Seed applies seeders for the specified environment
func (sm *SeedManager) Seed(ctx context.Context, environment string, seederNames ...string) error {
	if len(seederNames) == 0 {
		// Apply all seeders for the environment
		return sm.seedAll(ctx, environment)
	}

	// Apply specific seeders
	return sm.seedSpecific(ctx, environment, seederNames)
}

// seedAll applies all seeders for the environment
func (sm *SeedManager) seedAll(ctx context.Context, environment string) error {
	// Filter seeders by environment
	var applicableSeeders []*Seeder
	for _, seeder := range sm.seeders {
		if seeder.Environment == environment || seeder.Environment == "all" {
			applicableSeeders = append(applicableSeeders, seeder)
		}
	}

	// Sort by order
	for i := 0; i < len(applicableSeeders); i++ {
		for j := i + 1; j < len(applicableSeeders); j++ {
			if applicableSeeders[i].Order > applicableSeeders[j].Order {
				applicableSeeders[i], applicableSeeders[j] = applicableSeeders[j], applicableSeeders[i]
			}
		}
	}

	// Apply seeders
	for _, seeder := range applicableSeeders {
		if err := sm.applySeeder(ctx, seeder); err != nil {
			return fmt.Errorf("failed to apply seeder %s: %w", seeder.Name, err)
		}
	}

	return nil
}

// seedSpecific applies specific seeders
func (sm *SeedManager) seedSpecific(ctx context.Context, environment string, seederNames []string) error {
	for _, name := range seederNames {
		seeder, exists := sm.seeders[name]
		if !exists {
			return fmt.Errorf("seeder %s not found", name)
		}

		if seeder.Environment != environment && seeder.Environment != "all" {
			sm.logger.WithFields(logrus.Fields{
				"seeder":      name,
				"environment": environment,
			}).Warn("Seeder environment mismatch, skipping")
			continue
		}

		if err := sm.applySeeder(ctx, seeder); err != nil {
			return fmt.Errorf("failed to apply seeder %s: %w", name, err)
		}
	}

	return nil
}

// applySeeder applies a single seeder
func (sm *SeedManager) applySeeder(ctx context.Context, seeder *Seeder) error {
	if sm.appliedSeeders[seeder.Name] {
		sm.logger.WithField("seeder", seeder.Name).Debug("Seeder already applied")
		return nil
	}

	// Check dependencies
	for _, dep := range seeder.Dependencies {
		if !sm.appliedSeeders[dep] {
			return fmt.Errorf("seeder %s depends on %s which has not been applied", seeder.Name, dep)
		}
	}

	sm.logger.WithFields(logrus.Fields{
		"seeder":      seeder.Name,
		"description": seeder.Description,
	}).Info("Applying seeder")

	// Apply seed data
	for tableName, data := range seeder.Data {
		if err := sm.seedTable(ctx, tableName, data); err != nil {
			return fmt.Errorf("failed to seed table %s: %w", tableName, err)
		}
	}

	sm.appliedSeeders[seeder.Name] = true

	sm.logger.WithField("seeder", seeder.Name).Info("Seeder applied successfully")

	return nil
}

// seedTable seeds data for a specific table
func (sm *SeedManager) seedTable(ctx context.Context, tableName string, data interface{}) error {
	// Convert data to appropriate format
	records, err := sm.convertToRecords(data)
	if err != nil {
		return fmt.Errorf("failed to convert seed data: %w", err)
	}

	if len(records) == 0 {
		return nil
	}

	// Use GORM to insert records
	for _, record := range records {
		if err := sm.db.WithContext(ctx).Table(tableName).Create(record).Error; err != nil {
			return fmt.Errorf("failed to insert record into %s: %w", tableName, err)
		}
	}

	sm.logger.WithFields(logrus.Fields{
		"table":  tableName,
		"count":  len(records),
	}).Debug("Table seeded")

	return nil
}

// convertToRecords converts interface{} data to a slice of records
func (sm *SeedManager) convertToRecords(data interface{}) ([]interface{}, error) {
	switch v := data.(type) {
	case []interface{}:
		return v, nil
	case map[string]interface{}:
		// Single record as map
		return []interface{}{v}, nil
	default:
		// Try to marshal and unmarshal to get proper structure
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal seed data: %w", err)
		}

		var records []interface{}
		if err := json.Unmarshal(jsonData, &records); err != nil {
			// Try unmarshaling as single object
			var singleRecord interface{}
			if err := json.Unmarshal(jsonData, &singleRecord); err != nil {
				return nil, fmt.Errorf("failed to unmarshal seed data: %w", err)
			}
			return []interface{}{singleRecord}, nil
		}

		return records, nil
	}
}

// CreateDefaultSeeders creates default seeder files if they don't exist
func (sm *SeedManager) CreateDefaultSeeders() error {
	if err := os.MkdirAll(sm.seedsPath, 0755); err != nil {
		return fmt.Errorf("failed to create seeds directory: %w", err)
	}

	// Create development seeder
	devSeeder := Seeder{
		Name:        "development_data",
		Description: "Development environment seed data",
		Environment: "development",
		Dependencies: []string{},
		Order:       1,
		Data: map[string]interface{}{
			"users": []map[string]interface{}{
				{
					"id":         "550e8400-e29b-41d4-a716-446655440001",
					"email":      "admin@example.com",
					"first_name": "Admin",
					"last_name":  "User",
					"role":       "admin",
					"is_active":  true,
					"tenant_id":  "550e8400-e29b-41d4-a716-446655440000",
					"created_at": time.Now(),
					"updated_at": time.Now(),
				},
				{
					"id":         "550e8400-e29b-41d4-a716-446655440002",
					"email":      "user@example.com",
					"first_name": "Test",
					"last_name":  "User",
					"role":       "user",
					"is_active":  true,
					"tenant_id":  "550e8400-e29b-41d4-a716-446655440000",
					"created_at": time.Now(),
					"updated_at": time.Now(),
				},
			},
		},
	}

	if err := sm.writeSeederFile("development_data.json", devSeeder); err != nil {
		return fmt.Errorf("failed to write development seeder: %w", err)
	}

	// Create test seeder
	testSeeder := Seeder{
		Name:        "test_data",
		Description: "Test environment seed data",
		Environment: "test",
		Dependencies: []string{},
		Order:       1,
		Data: map[string]interface{}{
			"users": []map[string]interface{}{
				{
					"id":         "550e8400-e29b-41d4-a716-446655440001",
					"email":      "test@example.com",
					"first_name": "Test",
					"last_name":  "User",
					"role":       "user",
					"is_active":  true,
					"tenant_id":  "550e8400-e29b-41d4-a716-446655440000",
					"created_at": time.Now(),
					"updated_at": time.Now(),
				},
			},
		},
	}

	if err := sm.writeSeederFile("test_data.json", testSeeder); err != nil {
		return fmt.Errorf("failed to write test seeder: %w", err)
	}

	sm.logger.Info("Default seeders created successfully")

	return nil
}

// writeSeederFile writes a seeder to a JSON file
func (sm *SeedManager) writeSeederFile(filename string, seeder Seeder) error {
	filePath := filepath.Join(sm.seedsPath, filename)

	data, err := json.MarshalIndent(seeder, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal seeder: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write seeder file: %w", err)
	}

	return nil
}

// ClearSeeds clears all seeded data from the database
func (sm *SeedManager) ClearSeeds(ctx context.Context) error {
	// This would typically be used in test environments
	// Be careful with this in production!

	tables := []string{"users", "user_sessions", "activity_logs"} // Add your tables here

	for _, table := range tables {
		if err := sm.db.WithContext(ctx).Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
			return fmt.Errorf("failed to clear table %s: %w", table, err)
		}
	}

	// Reset applied seeders
	sm.appliedSeeders = make(map[string]bool)

	sm.logger.Info("All seeded data cleared")

	return nil
}

// GetSeederStatus returns the current seeder status
func (sm *SeedManager) GetSeederStatus() *SeederStatus {
	appliedCount := len(sm.appliedSeeders)
	totalCount := len(sm.seeders)

	return &SeederStatus{
		TotalSeeders:   totalCount,
		AppliedSeeders: appliedCount,
		PendingSeeders: totalCount - appliedCount,
		Applied:        sm.getAppliedSeederNames(),
		Pending:        sm.getPendingSeederNames(),
	}
}

// SeederStatus represents the current seeder status
type SeederStatus struct {
	TotalSeeders   int      `json:"total_seeders"`
	AppliedSeeders int      `json:"applied_seeders"`
	PendingSeeders int      `json:"pending_seeders"`
	Applied        []string `json:"applied"`
	Pending        []string `json:"pending"`
}

// getAppliedSeederNames returns the names of applied seeders
func (sm *SeedManager) getAppliedSeederNames() []string {
	var names []string
	for name := range sm.appliedSeeders {
		names = append(names, name)
	}
	return names
}

// getPendingSeederNames returns the names of pending seeders
func (sm *SeedManager) getPendingSeederNames() []string {
	var names []string
	for name := range sm.seeders {
		if !sm.appliedSeeders[name] {
			names = append(names, name)
		}
	}
	return names
}