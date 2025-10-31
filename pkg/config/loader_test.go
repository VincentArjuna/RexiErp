package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type TestConfig struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Debug   bool   `yaml:"debug"`
	Timeout int    `yaml:"timeout"`
}

func TestLoaderLoadFromFile(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	t.Run("successful file load", func(t *testing.T) {
		// Create temporary config file
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		testConfig := TestConfig{
			Name:    "TestApp",
			Version: "1.0.0",
			Debug:   true,
			Timeout: 30,
		}

		data, err := yaml.Marshal(testConfig)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(configPath, data, 0644))

		loader := NewLoader(logger)
		var cfg TestConfig

		err = loader.LoadFromFile(configPath, &cfg)
		assert.NoError(t, err)
		assert.Equal(t, testConfig, cfg)
	})

	t.Run("file does not exist", func(t *testing.T) {
		loader := NewLoader(logger)
		var cfg TestConfig

		err := loader.LoadFromFile("/non/existent/file.yaml", &cfg)
		assert.NoError(t, err) // Should not error, just log warning
	})

	t.Run("invalid YAML file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "invalid.yaml")

		require.NoError(t, os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644))

		loader := NewLoader(logger)
		var cfg TestConfig

		err := loader.LoadFromFile(configPath, &cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse configuration file")
	})

	t.Run("unreadable file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		testConfig := TestConfig{Name: "Test"}
		data, err := yaml.Marshal(testConfig)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(configPath, data, 0644))

		// Make file unreadable
		require.NoError(t, os.Chmod(configPath, 0000))
		defer os.Chmod(configPath, 0644) // Restore for cleanup

		loader := NewLoader(logger)
		var cfg TestConfig

		err = loader.LoadFromFile(configPath, &cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read configuration file")
	})
}

func TestLoaderLoadFromDir(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("successful directory load with config.yaml", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		testConfig := TestConfig{
			Name:    "TestApp",
			Version: "1.0.0",
		}

		data, err := yaml.Marshal(testConfig)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(configPath, data, 0644))

		loader := NewLoader(logger)
		var cfg TestConfig

		err = loader.LoadFromDir(tmpDir, &cfg)
		assert.NoError(t, err)
		assert.Equal(t, testConfig, cfg)
	})

	t.Run("successful directory load with config.yml", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yml")

		testConfig := TestConfig{
			Name:    "TestApp",
			Version: "2.0.0",
		}

		data, err := yaml.Marshal(testConfig)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(configPath, data, 0644))

		loader := NewLoader(logger)
		var cfg TestConfig

		err = loader.LoadFromDir(tmpDir, &cfg)
		assert.NoError(t, err)
		assert.Equal(t, testConfig, cfg)
	})

	t.Run("no config file found in directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		loader := NewLoader(logger)
		var cfg TestConfig

		err := loader.LoadFromDir(tmpDir, &cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no configuration file found")
	})
}

func TestLoaderSaveToFile(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("successful file save", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "subdir", "config.yaml")

		testConfig := TestConfig{
			Name:    "TestApp",
			Version: "1.0.0",
			Debug:   true,
			Timeout: 60,
		}

		loader := NewLoader(logger)

		err := loader.SaveToFile(configPath, &testConfig)
		assert.NoError(t, err)

		// Verify file was created with correct permissions
		info, err := os.Stat(configPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

		// Verify content
		data, err := os.ReadFile(configPath)
		require.NoError(t, err)

		var savedConfig TestConfig
		err = yaml.Unmarshal(data, &savedConfig)
		require.NoError(t, err)
		assert.Equal(t, testConfig, savedConfig)
	})

	t.Run("save to directory without write permissions", func(t *testing.T) {
		// Create a directory without write permissions
		tmpDir := t.TempDir()
		readOnlyDir := filepath.Join(tmpDir, "readonly")
		require.NoError(t, os.Mkdir(readOnlyDir, 0444))
		defer os.Chmod(readOnlyDir, 0755) // Restore for cleanup

		configPath := filepath.Join(readOnlyDir, "config.yaml")
		testConfig := TestConfig{Name: "Test"}

		loader := NewLoader(logger)

		err := loader.SaveToFile(configPath, &testConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "permission denied")
	})
}

func TestLoaderValidateEnvironment(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("all required variables present", func(t *testing.T) {
		os.Setenv("TEST_VAR1", "value1")
		os.Setenv("TEST_VAR2", "value2")
		defer func() {
			os.Unsetenv("TEST_VAR1")
			os.Unsetenv("TEST_VAR2")
		}()

		loader := NewLoader(logger)
		requiredVars := []string{"TEST_VAR1", "TEST_VAR2"}

		err := loader.ValidateEnvironment(requiredVars)
		assert.NoError(t, err)
	})

	t.Run("missing required variables", func(t *testing.T) {
		os.Setenv("TEST_VAR1", "value1")
		defer os.Unsetenv("TEST_VAR1")

		loader := NewLoader(logger)
		requiredVars := []string{"TEST_VAR1", "TEST_VAR2", "TEST_VAR3"}

		err := loader.ValidateEnvironment(requiredVars)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required environment variables")
		assert.Contains(t, err.Error(), "TEST_VAR2")
		assert.Contains(t, err.Error(), "TEST_VAR3")
	})

	t.Run("empty required variables", func(t *testing.T) {
		os.Setenv("TEST_VAR1", "value1")
		os.Setenv("TEST_VAR2", "") // Empty string
		defer func() {
			os.Unsetenv("TEST_VAR1")
			os.Unsetenv("TEST_VAR2")
		}()

		loader := NewLoader(logger)
		requiredVars := []string{"TEST_VAR1", "TEST_VAR2"}

		err := loader.ValidateEnvironment(requiredVars)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "TEST_VAR2")
	})
}

func TestGetConfigPaths(t *testing.T) {
	// Save current working directory
	originalWD, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWD)

	// Change to temp directory for testing
	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))

	// Set environment variable
	os.Setenv("APP_ENV", "test")
	defer os.Unsetenv("APP_ENV")

	paths := GetConfigPaths()

	// Should include current directory configs
	assert.Contains(t, paths, filepath.Join(tmpDir, "config.yaml"))
	assert.Contains(t, paths, filepath.Join(tmpDir, "config.yml"))

	// Should include config directory paths
	assert.Contains(t, paths, "configs/local/config.yaml")
	assert.Contains(t, paths, "configs/local/config.yml")

	// Should include environment-specific paths
	assert.Contains(t, paths, "configs/test/config.yaml")
	assert.Contains(t, paths, "configs/test/config.yml")

	// Should include home directory paths
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	assert.Contains(t, paths, filepath.Join(home, ".rexi-erp", "config.yaml"))
	assert.Contains(t, paths, filepath.Join(home, ".rexi-erp", "config.yml"))

	// Should include system paths
	assert.Contains(t, paths, "/etc/rexi-erp/config.yaml")
	assert.Contains(t, paths, "/etc/rexi-erp/config.yml")
}

func TestNewLoader(t *testing.T) {
	logger := logrus.New()
	loader := NewLoader(logger)

	assert.NotNil(t, loader)
	assert.Equal(t, logger, loader.logger)
}

// Test configuration loading with time fields
func TestConfigWithTimeFields(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	type ConfigWithTime struct {
		Name        string        `yaml:"name"`
		Timeout     time.Duration `yaml:"timeout"`
		CreatedAt   time.Time     `yaml:"created_at"`
	}

	t.Run("load config with time fields", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		testConfig := ConfigWithTime{
			Name:      "TimeTest",
			Timeout:   5 * time.Minute,
			CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		}

		data, err := yaml.Marshal(testConfig)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(configPath, data, 0644))

		loader := NewLoader(logger)
		var cfg ConfigWithTime

		err = loader.LoadFromFile(configPath, &cfg)
		assert.NoError(t, err)
		assert.Equal(t, testConfig, cfg)
	})
}