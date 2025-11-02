package logger

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Config holds the logger configuration
type Config struct {
	Level       string `json:"level" yaml:"level"`
	Format      string `json:"format" yaml:"format"`       // "json" or "text"
	ServiceName string `json:"serviceName" yaml:"serviceName"`
	Version     string `json:"version" yaml:"version"`
	Environment string `json:"environment" yaml:"environment"` // "development", "staging", "production"
}

// DefaultConfig returns a default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:       logrus.InfoLevel.String(),
		Format:      "json",
		ServiceName: "rexi-erp",
		Version:     "1.0.0",
		Environment: "development",
	}
}

// Logger wraps logrus with additional functionality
type Logger struct {
	*logrus.Logger
	config *Config
}

// NewLogger creates a new logger instance with the given configuration
func NewLogger(config *Config) *Logger {
	if config == nil {
		config = DefaultConfig()
	}

	log := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	// Disable DEBUG in production
	if strings.ToLower(config.Environment) == "production" && level == logrus.DebugLevel {
		log.SetLevel(logrus.InfoLevel)
	}

	// Set formatter
	if strings.ToLower(config.Format) == "json" {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "@timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// Set output
	log.SetOutput(os.Stdout)

	return &Logger{
		Logger: log,
		config: config,
	}
}

// WithServiceContext adds service context to the logger
func (l *Logger) WithServiceContext() *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields{
		"service":     l.config.ServiceName,
		"version":     l.config.Version,
		"environment": l.config.Environment,
	})
}

// WithRequestContext adds request context to the logger
func (l *Logger) WithRequestContext(correlationID, tenantID, userID string) *logrus.Entry {
	fields := logrus.Fields{
		"service":     l.config.ServiceName,
		"version":     l.config.Version,
		"environment": l.config.Environment,
	}

	if correlationID != "" {
		fields["correlation_id"] = correlationID
	}
	if tenantID != "" {
		fields["tenant_id"] = tenantID
	}
	if userID != "" {
		fields["user_id"] = userID
	}

	return l.Logger.WithFields(fields)
}

// WithError adds error context to the logger
func (l *Logger) WithError(err error) *logrus.Entry {
	if err == nil {
		return l.WithServiceContext()
	}
	return l.Logger.WithFields(logrus.Fields{
		"service":     l.config.ServiceName,
		"version":     l.config.Version,
		"environment": l.config.Environment,
	}).WithError(err)
}

// WithFields adds custom fields to the logger
func (l *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	// Merge service context with custom fields
	serviceFields := logrus.Fields{
		"service":     l.config.ServiceName,
		"version":     l.config.Version,
		"environment": l.config.Environment,
	}

	// Combine fields
	for k, v := range fields {
		serviceFields[k] = v
	}

	return l.Logger.WithFields(serviceFields)
}

// GetConfig returns the current logger configuration
func (l *Logger) GetConfig() *Config {
	return l.config
}

// SetLevel dynamically updates the log level
func (l *Logger) SetLevel(level string) error {
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}

	// Prevent enabling DEBUG in production
	if strings.ToLower(l.config.Environment) == "production" && logLevel == logrus.DebugLevel {
		logLevel = logrus.InfoLevel
	}

	l.Logger.SetLevel(logLevel)
	l.config.Level = logLevel.String()
	return nil
}

// IsProduction checks if the logger is in production mode
func (l *Logger) IsProduction() bool {
	return strings.ToLower(l.config.Environment) == "production"
}