package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Loader handles loading configuration from various sources
type Loader struct {
	logger *logrus.Logger
}

// NewLoader creates a new configuration loader
func NewLoader(logger *logrus.Logger) *Loader {
	return &Loader{
		logger: logger,
	}
}

// LoadFromFile loads configuration from a YAML file
func (l *Loader) LoadFromFile(filename string, cfg interface{}) error {
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		l.logger.Warnf("Configuration file not found: %s", filename)
		return nil
	}

	// Read file
	data, err := os.ReadFile(filename) // #nosec G304 -- filename is controlled internally, not user input
	if err != nil {
		return fmt.Errorf("failed to read configuration file %s: %w", filename, err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to parse configuration file %s: %w", filename, err)
	}

	l.logger.Infof("Configuration loaded from: %s", filename)
	return nil
}

// LoadFromDir loads configuration from a directory
func (l *Loader) LoadFromDir(dir string, cfg interface{}) error {
	// Try different file names
	possibleFiles := []string{
		"config.yaml",
		"config.yml",
		"application.yaml",
		"application.yml",
	}

	for _, filename := range possibleFiles {
		fullPath := filepath.Join(dir, filename)
		// Check if file exists first
		if _, err := os.Stat(fullPath); err == nil {
			// File exists, try to load it
			if loadErr := l.LoadFromFile(fullPath, cfg); loadErr == nil {
				return nil
			}
		}
	}

	return fmt.Errorf("no configuration file found in directory: %s", dir)
}

// SaveToFile saves configuration to a YAML file
func (l *Loader) SaveToFile(filename string, cfg interface{}) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Write file with restricted permissions
	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write configuration file %s: %w", filename, err)
	}

	l.logger.Infof("Configuration saved to: %s", filename)
	return nil
}

// ValidateEnvironment checks if required environment variables are set
func (l *Loader) ValidateEnvironment(requiredVars []string) error {
	var missing []string

	for _, varName := range requiredVars {
		if os.Getenv(varName) == "" {
			missing = append(missing, varName)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missing)
	}

	return nil
}

// GetConfigPaths returns possible configuration file paths based on environment
func GetConfigPaths() []string {
	var paths []string

	// Current directory
	if pwd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(pwd, "config.yaml"))
		paths = append(paths, filepath.Join(pwd, "config.yml"))
	}

	// Config directory
	paths = append(paths, filepath.Join("configs", "local", "config.yaml"))
	paths = append(paths, filepath.Join("configs", "local", "config.yml"))

	// Environment-based paths
	if env := os.Getenv("APP_ENV"); env != "" {
		paths = append(paths, filepath.Join("configs", env, "config.yaml"))
		paths = append(paths, filepath.Join("configs", env, "config.yml"))
	}

	// Home directory
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".rexi-erp", "config.yaml"))
		paths = append(paths, filepath.Join(home, ".rexi-erp", "config.yml"))
	}

	// /etc directory
	paths = append(paths, "/etc/rexi-erp/config.yaml")
	paths = append(paths, "/etc/rexi-erp/config.yml")

	return paths
}