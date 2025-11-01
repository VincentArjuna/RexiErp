package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/VincentArjuna/RexiErp/internal/shared/config"
	"github.com/VincentArjuna/RexiErp/internal/shared/database"
)

// HealthStatus represents the health status
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusDegraded  HealthStatus = "degraded"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status      HealthStatus            `json:"status"`
	Timestamp   time.Time               `json:"timestamp"`
	Version     string                  `json:"version"`
	Environment string                  `json:"environment"`
	Uptime      time.Duration           `json:"uptime"`
	Checks      map[string]CheckResult  `json:"checks"`
	System      SystemInfo              `json:"system"`
}

// CheckResult represents the result of a health check
type CheckResult struct {
	Status  HealthStatus `json:"status"`
	Message string      `json:"message,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

// SystemInfo represents system information
type SystemInfo struct {
	GoVersion   string `json:"go_version"`
	NumGoroutine int   `json:"num_goroutine"`
	MemoryUsage MemoryInfo `json:"memory"`
}

// MemoryInfo represents memory usage information
type MemoryInfo struct {
	Alloc      uint64 `json:"alloc_mb"`
	TotalAlloc uint64 `json:"total_alloc_mb"`
	Sys        uint64 `json:"sys_mb"`
	NumGC      uint32 `json:"num_gc"`
}

// HealthChecker provides health check functionality
type HealthChecker struct {
	db     *database.Database
	config *config.Config
	logger *logrus.Logger
	startTime time.Time
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(db *database.Database, cfg *config.Config, logger *logrus.Logger) *HealthChecker {
	return &HealthChecker{
		db:     db,
		config: cfg,
		logger: logger,
		startTime: time.Now(),
	}
}

// RegisterRoutes registers health check routes
func (h *HealthChecker) RegisterRoutes(router *gin.Engine) {
	health := router.Group("/health")
	{
		health.GET("/", h.BasicHealth)
		health.GET("/live", h.Liveness)
		health.GET("/ready", h.Readiness)
		health.GET("/detailed", h.DetailedHealth)
	}
}

// BasicHealth returns a simple health check
func (h *HealthChecker) BasicHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    StatusHealthy,
		"timestamp": time.Now(),
		"message":   "API Gateway is healthy",
	})
}

// Liveness checks if the service is alive
func (h *HealthChecker) Liveness(c *gin.Context) {
	// Basic liveness check - if we can respond, we're alive
	c.JSON(http.StatusOK, gin.H{
		"status":    StatusHealthy,
		"timestamp": time.Now(),
		"uptime":    time.Since(h.startTime).String(),
	})
}

// Readiness checks if the service is ready to accept traffic
func (h *HealthChecker) Readiness(c *gin.Context) {
	checks := make(map[string]CheckResult)
	overallStatus := StatusHealthy

	// Check database connection
	dbStatus := h.checkDatabase()
	checks["database"] = dbStatus
	if dbStatus.Status != StatusHealthy {
		overallStatus = StatusUnhealthy
	}

	response := HealthResponse{
		Status:      overallStatus,
		Timestamp:   time.Now(),
		Version:     h.config.App.Version,
		Environment: h.config.App.Environment,
		Uptime:      time.Since(h.startTime),
		Checks:      checks,
	}

	if overallStatus == StatusHealthy {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusServiceUnavailable, response)
	}
}

// DetailedHealth returns comprehensive health information
func (h *HealthChecker) DetailedHealth(c *gin.Context) {
	checks := make(map[string]CheckResult)
	overallStatus := StatusHealthy

	// Database check
	dbStatus := h.checkDatabase()
	checks["database"] = dbStatus
	if dbStatus.Status != StatusHealthy {
		overallStatus = StatusDegraded
	}

	// Configuration check
	configStatus := h.checkConfiguration()
	checks["configuration"] = configStatus
	if configStatus.Status != StatusHealthy {
		overallStatus = StatusDegraded
	}

	// System resources check
	systemStatus := h.checkSystemResources()
	checks["system"] = systemStatus
	if systemStatus.Status == StatusUnhealthy {
		overallStatus = StatusUnhealthy
	}

	response := HealthResponse{
		Status:      overallStatus,
		Timestamp:   time.Now(),
		Version:     h.config.App.Version,
		Environment: h.config.App.Environment,
		Uptime:      time.Since(h.startTime),
		Checks:      checks,
		System:      h.getSystemInfo(),
	}

	if overallStatus == StatusHealthy {
		c.JSON(http.StatusOK, response)
	} else if overallStatus == StatusDegraded {
		c.JSON(http.StatusOK, response) // Still 200 for degraded
	} else {
		c.JSON(http.StatusServiceUnavailable, response)
	}
}

// checkDatabase validates the database connection
func (h *HealthChecker) checkDatabase() CheckResult {
	if h.db == nil {
		return CheckResult{
			Status:  StatusUnhealthy,
			Message: "Database connection not initialized",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.db.HealthCheck(); err != nil {
		h.logger.WithError(err).Error("Database health check failed")
		return CheckResult{
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("Database connection failed: %v", err),
		}
	}

	// Get connection pool stats
	stats := h.db.GetStats()
	details := map[string]interface{}{
		"open_connections":     stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration.String(),
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}

	// Check connection pool health
	if stats.WaitCount > 100 && stats.WaitDuration > time.Second {
		return CheckResult{
			Status:  StatusDegraded,
			Message: "Database connection pool experiencing high wait times",
			Details: details,
		}
	}

	if stats.OpenConnections >= h.config.Database.MaxOpenConns*9/10 {
		return CheckResult{
			Status:  StatusDegraded,
			Message: "Database connection pool near capacity",
			Details: details,
		}
	}

	return CheckResult{
		Status:  StatusHealthy,
		Message: "Database connection healthy",
		Details: details,
	}
}

// checkConfiguration validates the application configuration
func (h *HealthChecker) checkConfiguration() CheckResult {
	// Validate critical configuration
	if h.config.JWT.Secret == "" {
		return CheckResult{
			Status:  StatusUnhealthy,
			Message: "JWT secret not configured",
		}
	}

	if h.config.JWT.Secret == "your-super-secret-jwt-key-for-development-only" && h.config.App.Environment == "production" {
		return CheckResult{
			Status:  StatusDegraded,
			Message: "Using default JWT secret in production",
		}
	}

	if len(h.config.APIKey.Keys) == 0 {
		return CheckResult{
			Status:  StatusDegraded,
			Message: "No API keys configured",
		}
	}

	return CheckResult{
		Status:  StatusHealthy,
		Message: "Configuration valid",
	}
}

// checkSystemResources checks system resource usage
func (h *HealthChecker) checkSystemResources() CheckResult {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Check memory usage (convert MB to bytes for comparison)
	allocMB := bToMb(m.Alloc)
	if allocMB > 500 { // 500MB threshold
		return CheckResult{
			Status:  StatusDegraded,
			Message: fmt.Sprintf("High memory usage: %d MB", allocMB),
			Details: map[string]interface{}{
				"alloc_mb":     allocMB,
				"sys_mb":       bToMb(m.Sys),
				"num_goroutine": runtime.NumGoroutine(),
			},
		}
	}

	// Check goroutine count
	if runtime.NumGoroutine() > 1000 {
		return CheckResult{
			Status:  StatusDegraded,
			Message: fmt.Sprintf("High goroutine count: %d", runtime.NumGoroutine()),
			Details: map[string]interface{}{
				"alloc_mb":     allocMB,
				"sys_mb":       bToMb(m.Sys),
				"num_goroutine": runtime.NumGoroutine(),
			},
		}
	}

	return CheckResult{
		Status:  StatusHealthy,
		Message: "System resources healthy",
		Details: map[string]interface{}{
			"alloc_mb":     allocMB,
			"sys_mb":       bToMb(m.Sys),
			"num_goroutine": runtime.NumGoroutine(),
		},
	}
}

// getSystemInfo returns system information
func (h *HealthChecker) getSystemInfo() SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemInfo{
		GoVersion:    runtime.Version(),
		NumGoroutine: runtime.NumGoroutine(),
		MemoryUsage: MemoryInfo{
			Alloc:      bToMb(m.Alloc),
			TotalAlloc: bToMb(m.TotalAlloc),
			Sys:        bToMb(m.Sys),
			NumGC:      m.NumGC,
		},
	}
}

// Helper function to convert bytes to megabytes
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// SetupDefaultHealthRoutes sets up default health check routes
func SetupDefaultHealthRoutes(router *gin.Engine, db *database.Database, cfg *config.Config, logger *logrus.Logger) {
	healthChecker := NewHealthChecker(db, cfg, logger)
	healthChecker.RegisterRoutes(router)
}

// Middleware to add health information to context
func HealthMiddleware(db *database.Database, cfg *config.Config, logger *logrus.Logger) gin.HandlerFunc {
	healthChecker := NewHealthChecker(db, cfg, logger)

	return func(c *gin.Context) {
		// Add basic health info to context
		c.Set("health_checker", healthChecker)
		c.Next()
	}
}