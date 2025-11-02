package metrics

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusMetrics holds all the metrics collectors
type PrometheusMetrics struct {
	// HTTP metrics
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	httpResponseSize    *prometheus.HistogramVec

	// Business metrics
	activeTenants         prometheus.Gauge
	totalTransactions     *prometheus.CounterVec
	transactionAmounts    *prometheus.HistogramVec
	userSessions          *prometheus.CounterVec
	apiCalls              *prometheus.CounterVec

	// System metrics
	uptimeCounter prometheus.Counter
	startTime     time.Time

	// Database metrics
	dbConnections     prometheus.Gauge
	dbQueryDuration   *prometheus.HistogramVec
	dbQueriesTotal    *prometheus.CounterVec
	dbConnectionErrors prometheus.Counter

	// Cache metrics
	cacheHits   *prometheus.CounterVec
	cacheMisses *prometheus.CounterVec

	// External service metrics
	externalServiceCalls    *prometheus.CounterVec
	externalServiceDuration *prometheus.HistogramVec

	// Custom metrics
	customMetrics map[string]prometheus.Collector
}

// NewPrometheusMetrics creates a new Prometheus metrics instance
func NewPrometheusMetrics(serviceName string) *PrometheusMetrics {
	pm := &PrometheusMetrics{
		startTime:     time.Now(),
		customMetrics: make(map[string]prometheus.Collector),
	}

	// Initialize HTTP metrics
	pm.httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "http_requests_total",
			Help:        "Total number of HTTP requests",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"method", "path", "status_code", "tenant_id"},
	)

	pm.httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "http_request_duration_seconds",
			Help:        "HTTP request duration in seconds",
			ConstLabels: prometheus.Labels{"service": serviceName},
			Buckets:     []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path", "status_code", "tenant_id"},
	)

	pm.httpResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "http_response_size_bytes",
			Help:        "HTTP response size in bytes",
			ConstLabels: prometheus.Labels{"service": serviceName},
			Buckets:     []float64{100, 500, 1000, 5000, 10000, 50000, 100000, 500000, 1000000},
		},
		[]string{"method", "path", "status_code"},
	)

	// Initialize business metrics
	pm.activeTenants = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "active_tenants_total",
			Help:        "Total number of active tenants",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
	)

	pm.totalTransactions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "transactions_total",
			Help:        "Total number of business transactions",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"type", "tenant_id", "status"},
	)

	pm.transactionAmounts = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "transaction_amount",
			Help:        "Transaction amounts in IDR",
			ConstLabels: prometheus.Labels{"service": serviceName, "currency": "IDR"},
			Buckets:     []float64{1000, 5000, 10000, 50000, 100000, 500000, 1000000, 5000000, 10000000, 50000000},
		},
		[]string{"type", "tenant_id"},
	)

	pm.userSessions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "user_sessions_total",
			Help:        "Total number of user sessions",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"tenant_id", "status"},
	)

	pm.apiCalls = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "api_calls_total",
			Help:        "Total number of API calls",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"endpoint", "method", "tenant_id", "status"},
	)

	// Initialize system metrics
	pm.uptimeCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name:        "uptime_seconds",
			Help:        "Service uptime in seconds",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
	)

	pm.dbConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "db_connections_active",
			Help:        "Active database connections",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
	)

	pm.dbQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "db_query_duration_seconds",
			Help:        "Database query duration in seconds",
			ConstLabels: prometheus.Labels{"service": serviceName},
			Buckets:     []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"table", "operation", "tenant_id"},
	)

	pm.dbQueriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "db_queries_total",
			Help:        "Total number of database queries",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"table", "operation", "status", "tenant_id"},
	)

	pm.dbConnectionErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name:        "db_connection_errors_total",
			Help:        "Total number of database connection errors",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
	)

	// Initialize cache metrics
	pm.cacheHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "cache_hits_total",
			Help:        "Total number of cache hits",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"cache_type", "key_prefix"},
	)

	pm.cacheMisses = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "cache_misses_total",
			Help:        "Total number of cache misses",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"cache_type", "key_prefix"},
	)

	// Initialize external service metrics
	pm.externalServiceCalls = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "external_service_calls_total",
			Help:        "Total number of external service calls",
			ConstLabels: prometheus.Labels{"caller_service": serviceName},
		},
		[]string{"service", "endpoint", "status"},
	)

	pm.externalServiceDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "external_service_duration_seconds",
			Help:        "External service call duration in seconds",
			ConstLabels: prometheus.Labels{"caller_service": serviceName},
			Buckets:     []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30},
		},
		[]string{"service", "endpoint"},
	)

	return pm
}

// Register registers all metrics with the default Prometheus registry
func (pm *PrometheusMetrics) Register() error {
	// Register HTTP metrics
	if err := prometheus.Register(pm.httpRequestsTotal); err != nil {
		return err
	}
	if err := prometheus.Register(pm.httpRequestDuration); err != nil {
		return err
	}
	if err := prometheus.Register(pm.httpResponseSize); err != nil {
		return err
	}

	// Register business metrics
	if err := prometheus.Register(pm.activeTenants); err != nil {
		return err
	}
	if err := prometheus.Register(pm.totalTransactions); err != nil {
		return err
	}
	if err := prometheus.Register(pm.transactionAmounts); err != nil {
		return err
	}
	if err := prometheus.Register(pm.userSessions); err != nil {
		return err
	}
	if err := prometheus.Register(pm.apiCalls); err != nil {
		return err
	}

	// Register system metrics
	if err := prometheus.Register(pm.uptimeCounter); err != nil {
		return err
	}
	if err := prometheus.Register(pm.dbConnections); err != nil {
		return err
	}
	if err := prometheus.Register(pm.dbQueryDuration); err != nil {
		return err
	}
	if err := prometheus.Register(pm.dbQueriesTotal); err != nil {
		return err
	}
	if err := prometheus.Register(pm.dbConnectionErrors); err != nil {
		return err
	}

	// Register cache metrics
	if err := prometheus.Register(pm.cacheHits); err != nil {
		return err
	}
	if err := prometheus.Register(pm.cacheMisses); err != nil {
		return err
	}

	// Register external service metrics
	if err := prometheus.Register(pm.externalServiceCalls); err != nil {
		return err
	}
	if err := prometheus.Register(pm.externalServiceDuration); err != nil {
		return err
	}

	// Register custom metrics
	for _, metric := range pm.customMetrics {
		if err := prometheus.Register(metric); err != nil {
			return err
		}
	}

	return nil
}

// Unregister unregisters all metrics from the default Prometheus registry
func (pm *PrometheusMetrics) Unregister() {
	prometheus.Unregister(pm.httpRequestsTotal)
	prometheus.Unregister(pm.httpRequestDuration)
	prometheus.Unregister(pm.httpResponseSize)
	prometheus.Unregister(pm.activeTenants)
	prometheus.Unregister(pm.totalTransactions)
	prometheus.Unregister(pm.transactionAmounts)
	prometheus.Unregister(pm.userSessions)
	prometheus.Unregister(pm.apiCalls)
	prometheus.Unregister(pm.uptimeCounter)
	prometheus.Unregister(pm.dbConnections)
	prometheus.Unregister(pm.dbQueryDuration)
	prometheus.Unregister(pm.dbQueriesTotal)
	prometheus.Unregister(pm.dbConnectionErrors)
	prometheus.Unregister(pm.cacheHits)
	prometheus.Unregister(pm.cacheMisses)
	prometheus.Unregister(pm.externalServiceCalls)
	prometheus.Unregister(pm.externalServiceDuration)

	for _, metric := range pm.customMetrics {
		prometheus.Unregister(metric)
	}
}

// UpdateUptime updates the uptime counter
func (pm *PrometheusMetrics) UpdateUptime() {
	pm.uptimeCounter.Add(time.Since(pm.startTime).Seconds())
}

// MetricsHandler returns the Prometheus metrics HTTP handler
func MetricsHandler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}

// StartMetricsServer starts a metrics server on the given port
func StartMetricsServer(port string) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	return server.ListenAndServe()
}

// GetMetricNames returns a list of all registered metric names
func (pm *PrometheusMetrics) GetMetricNames() []string {
	names := []string{
		"http_requests_total",
		"http_request_duration_seconds",
		"http_response_size_bytes",
		"active_tenants_total",
		"transactions_total",
		"transaction_amount",
		"user_sessions_total",
		"api_calls_total",
		"uptime_seconds",
		"db_connections_active",
		"db_query_duration_seconds",
		"db_queries_total",
		"db_connection_errors_total",
		"cache_hits_total",
		"cache_misses_total",
		"external_service_calls_total",
		"external_service_duration_seconds",
	}

	for name := range pm.customMetrics {
		names = append(names, name)
	}

	return names
}