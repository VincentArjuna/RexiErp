package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"

	"github.com/VincentArjuna/RexiErp/internal/shared/config"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
)

// ComponentHealth represents the health status of a component
type ComponentHealth struct {
	Name      string                 `json:"name"`
	Status    HealthStatus            `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time               `json:"timestamp"`
}

// HealthReport represents the overall health report
type HealthReport struct {
	Status     HealthStatus                `json:"status"`
	Components map[string]ComponentHealth   `json:"components"`
	Summary    HealthSummary               `json:"summary"`
	Timestamp  time.Time                   `json:"timestamp"`
}

// HealthSummary provides a summary of the health check
type HealthSummary struct {
	Total   int `json:"total"`
	Healthy int `json:"healthy"`
	Unhealthy int `json:"unhealthy"`
	Degraded  int `json:"degraded"`
}

// HealthChecker provides comprehensive health checking for all components
type HealthChecker struct {
	multiDBManager *MultiDBManager
	logger         *logrus.Logger
	config         *config.Config
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(multiDBManager *MultiDBManager, logger *logrus.Logger, cfg *config.Config) *HealthChecker {
	return &HealthChecker{
		multiDBManager: multiDBManager,
		logger:         logger,
		config:         cfg,
	}
}

// CheckHealth performs a comprehensive health check of all components
func (hc *HealthChecker) CheckHealth(ctx context.Context) *HealthReport {
	report := &HealthReport{
		Components: make(map[string]ComponentHealth),
		Timestamp:  time.Now(),
	}

	// Check master database
	report.Components["database_master"] = hc.checkDatabase(ctx, "master", hc.multiDBManager)

	// Check replica databases
	report.Components["database_replicas"] = hc.checkDatabaseReplicas(ctx, hc.multiDBManager)

	// Check Redis
	report.Components["redis"] = hc.checkRedis(ctx, hc.multiDBManager)

	// Check MinIO
	report.Components["minio"] = hc.checkMinIO(ctx, hc.multiDBManager)

	// Calculate overall status
	report.Summary = hc.calculateSummary(report.Components)
	report.Status = hc.calculateOverallStatus(report.Summary)

	return report
}

// checkDatabase checks the health of a specific database
func (hc *HealthChecker) checkDatabase(ctx context.Context, name string, multiDB *MultiDBManager) ComponentHealth {
	db, err := multiDB.GetDatabase(name)
	if err != nil {
		return ComponentHealth{
			Name:      fmt.Sprintf("database_%s", name),
			Status:    HealthStatusUnhealthy,
			Message:   fmt.Sprintf("Database not available: %v", err),
			Timestamp: time.Now(),
		}
	}

	// Perform health check
	if err := db.HealthCheck(); err != nil {
		return ComponentHealth{
			Name:      fmt.Sprintf("database_%s", name),
			Status:    HealthStatusUnhealthy,
			Message:   fmt.Sprintf("Health check failed: %v", err),
			Timestamp: time.Now(),
		}
	}

	// Get connection metrics
	metrics := db.GetMetrics()

	// Check connection pool health
	status := HealthStatusHealthy
	var messages []string
	details := make(map[string]interface{})

	details["total_connections"] = metrics.TotalConnections
	details["active_connections"] = metrics.ActiveConnections
	details["idle_connections"] = metrics.IdleConnections
	details["wait_count"] = metrics.WaitCount
	details["wait_duration"] = metrics.WaitDuration.String()
	details["health_check_failures"] = metrics.HealthCheckFailures
	details["reconnection_attempts"] = metrics.ReconnectionAttempts

	// Check for connection pool issues
	if metrics.HealthCheckFailures > 5 {
		status = HealthStatusDegraded
		messages = append(messages, fmt.Sprintf("Multiple health check failures: %d", metrics.HealthCheckFailures))
	}

	if metrics.ReconnectionAttempts > 10 {
		status = HealthStatusDegraded
		messages = append(messages, fmt.Sprintf("Frequent reconnections: %d", metrics.ReconnectionAttempts))
	}

	// Check connection pool utilization
	if metrics.TotalConnections > 0 {
		utilization := float64(metrics.ActiveConnections) / float64(metrics.TotalConnections)
		details["connection_utilization"] = utilization

		if utilization > 0.9 {
			status = HealthStatusDegraded
			messages = append(messages, fmt.Sprintf("High connection pool utilization: %.2f%%", utilization*100))
		}
	}

	// Check wait times
	if metrics.WaitCount > 0 {
		avgWaitTime := float64(metrics.WaitDuration) / float64(metrics.WaitCount)
		details["avg_wait_time"] = avgWaitTime

		if avgWaitTime > float64(50*time.Millisecond.Nanoseconds()) {
			status = HealthStatusDegraded
			messages = append(messages, fmt.Sprintf("High average wait time: %v", time.Duration(avgWaitTime)))
		}
	}

	message := ""
	if len(messages) > 0 {
		message = strings.Join(messages, "; ")
	}

	return ComponentHealth{
		Name:      fmt.Sprintf("database_%s", name),
		Status:    status,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
	}
}

// checkDatabaseReplicas checks the health of replica databases
func (hc *HealthChecker) checkDatabaseReplicas(ctx context.Context, multiDB *MultiDBManager) ComponentHealth {
	// Get all replica databases
	replicas := make(map[string]*Database)
	multiDB.mu.RLock()
	for name, db := range multiDB.databases {
		if name != "master" && strings.HasPrefix(name, "replica_") {
			replicas[name] = db
		}
	}
	multiDB.mu.RUnlock()

	if len(replicas) == 0 {
		return ComponentHealth{
			Name:      "database_replicas",
			Status:    HealthStatusDegraded,
			Message:   "No replica databases configured",
			Timestamp: time.Now(),
		}
	}

	var healthyReplicas int
	var totalReplicas int
	var messages []string
	details := make(map[string]interface{})

	for name, db := range replicas {
		totalReplicas++
		if err := db.HealthCheck(); err != nil {
			messages = append(messages, fmt.Sprintf("%s: %v", name, err))
		} else {
			healthyReplicas++
		}
	}

	details["total_replicas"] = totalReplicas
	details["healthy_replicas"] = healthyReplicas
	details["replica_health_ratio"] = float64(healthyReplicas) / float64(totalReplicas)

	status := HealthStatusHealthy
	message := ""

	if healthyReplicas == 0 {
		status = HealthStatusUnhealthy
		message = "All replicas are unhealthy"
	} else if healthyReplicas < totalReplicas {
		status = HealthStatusDegraded
		message = fmt.Sprintf("Only %d/%d replicas are healthy", healthyReplicas, totalReplicas)
	}

	if len(messages) > 0 {
		if message != "" {
			message += "; "
		}
		message += strings.Join(messages, "; ")
	}

	return ComponentHealth{
		Name:      "database_replicas",
		Status:    status,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
	}
}

// checkRedis checks the health of Redis connection
func (hc *HealthChecker) checkRedis(ctx context.Context, multiDB *MultiDBManager) ComponentHealth {
	redisClient := multiDB.GetRedis()
	if redisClient == nil {
		return ComponentHealth{
			Name:      "redis",
			Status:    HealthStatusUnhealthy,
			Message:   "Redis client not initialized",
			Timestamp: time.Now(),
		}
	}

	// Check Redis connectivity
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return ComponentHealth{
			Name:      "redis",
			Status:    HealthStatusUnhealthy,
			Message:   fmt.Sprintf("Redis ping failed: %v", err),
			Timestamp: time.Now(),
		}
	}

	// Get Redis info and pool stats (only for single client)
	var info string
	var poolStats *redis.PoolStats

	if client, ok := redisClient.(*redis.Client); ok {
		if result, err := client.Info(ctx).Result(); err == nil {
			info = result
		}
		stats := client.PoolStats()
		poolStats = stats
	} else {
		info = "Redis cluster/sentinel (detailed info not available)"
	}
	details := make(map[string]interface{})

	details["redis_info"] = info
	if poolStats != nil {
		details["pool_hits"] = poolStats.Hits
		details["pool_misses"] = poolStats.Misses
		details["pool_timeouts"] = poolStats.Timeouts
		details["total_connections"] = poolStats.TotalConns
		details["idle_connections"] = poolStats.IdleConns
		details["stale_connections"] = poolStats.StaleConns

		// Calculate hit ratio
		if poolStats.Hits+poolStats.Misses > 0 {
			hitRatio := float64(poolStats.Hits) / float64(poolStats.Hits+poolStats.Misses)
			details["hit_ratio"] = hitRatio
		}
	}

	status := HealthStatusHealthy
	var messages []string

	// Check for connection pool issues
	if poolStats.Timeouts > 10 {
		status = HealthStatusDegraded
		messages = append(messages, fmt.Sprintf("High number of pool timeouts: %d", poolStats.Timeouts))
	}

	if poolStats.StaleConns > 5 {
		status = HealthStatusDegraded
		messages = append(messages, fmt.Sprintf("High number of stale connections: %d", poolStats.StaleConns))
	}

	// Check connection utilization
	if poolStats.TotalConns > 0 {
		utilization := float64(poolStats.TotalConns-poolStats.IdleConns) / float64(poolStats.TotalConns)
		details["connection_utilization"] = utilization

		if utilization > 0.9 {
			status = HealthStatusDegraded
			messages = append(messages, fmt.Sprintf("High connection utilization: %.2f%%", utilization*100))
		}
	}

	message := ""
	if len(messages) > 0 {
		message = strings.Join(messages, "; ")
	}

	return ComponentHealth{
		Name:      "redis",
		Status:    status,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
	}
}

// checkMinIO checks the health of MinIO connection
func (hc *HealthChecker) checkMinIO(ctx context.Context, multiDB *MultiDBManager) ComponentHealth {
	minio := multiDB.GetMinIO()
	if minio == nil {
		return ComponentHealth{
			Name:      "minio",
			Status:    HealthStatusUnhealthy,
			Message:   "MinIO client not initialized",
			Timestamp: time.Now(),
		}
	}

	// Check MinIO connectivity
	ctx, cancel := context.WithTimeout(ctx, hc.config.MinIO.Timeout)
	defer cancel()

	// Try to list buckets
	_, err := minio.ListBuckets(ctx)
	if err != nil {
		return ComponentHealth{
			Name:      "minio",
			Status:    HealthStatusUnhealthy,
			Message:   fmt.Sprintf("MinIO connection failed: %v", err),
			Timestamp: time.Now(),
		}
	}

	// Check bucket access
	bucket := hc.config.MinIO.Bucket
	exists, err := minio.BucketExists(ctx, bucket)
	if err != nil {
		return ComponentHealth{
			Name:      "minio",
			Status:    HealthStatusDegraded,
			Message:   fmt.Sprintf("Failed to check bucket existence: %v", err),
			Timestamp: time.Now(),
		}
	}

	details := make(map[string]interface{})
	details["endpoint"] = hc.config.MinIO.Endpoint
	details["bucket"] = bucket
	details["bucket_exists"] = exists
	details["use_ssl"] = hc.config.MinIO.UseSSL
	details["region"] = hc.config.MinIO.Region

	status := HealthStatusHealthy
	message := ""

	if !exists {
		status = HealthStatusDegraded
		message = fmt.Sprintf("Configured bucket '%s' does not exist", bucket)
	}

	return ComponentHealth{
		Name:      "minio",
		Status:    status,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
	}
}

// calculateSummary calculates the health summary
func (hc *HealthChecker) calculateSummary(components map[string]ComponentHealth) HealthSummary {
	summary := HealthSummary{
		Total: len(components),
	}

	for _, component := range components {
		switch component.Status {
		case HealthStatusHealthy:
			summary.Healthy++
		case HealthStatusUnhealthy:
			summary.Unhealthy++
		case HealthStatusDegraded:
			summary.Degraded++
		}
	}

	return summary
}

// calculateOverallStatus calculates the overall health status
func (hc *HealthChecker) calculateOverallStatus(summary HealthSummary) HealthStatus {
	if summary.Unhealthy > 0 {
		return HealthStatusUnhealthy
	}
	if summary.Degraded > 0 {
		return HealthStatusDegraded
	}
	return HealthStatusHealthy
}

// IsHealthy returns true if all components are healthy
func (report *HealthReport) IsHealthy() bool {
	return report.Status == HealthStatusHealthy
}

// IsUnhealthy returns true if any component is unhealthy
func (report *HealthReport) IsUnhealthy() bool {
	return report.Status == HealthStatusUnhealthy
}

// IsDegraded returns true if any component is degraded
func (report *HealthReport) IsDegraded() bool {
	return report.Status == HealthStatusDegraded
}