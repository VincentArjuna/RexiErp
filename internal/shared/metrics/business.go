package metrics

import (
	"context"
	"time"

	"github.com/VincentArjuna/RexiErp/internal/shared/logger"
	"github.com/prometheus/client_golang/prometheus"
)

// BusinessMetricTypes defines the types of business metrics we can track
type BusinessMetricTypes struct {
	// E-commerce metrics
	OrdersCreated       *prometheus.CounterVec
	OrderValue          *prometheus.HistogramVec
	ProductsViewed      *prometheus.CounterVec
	CartAbandoned       *prometheus.CounterVec
	PaymentTransactions *prometheus.CounterVec

	// User activity metrics
	UserLogins      *prometheus.CounterVec
	UserRegistrations *prometheus.CounterVec
	ActiveSessions  prometheus.Gauge

	// Inventory metrics
	StockLevels    prometheus.Gauge
	StockMovements *prometheus.CounterVec
	ProductChanges *prometheus.CounterVec

	// Financial metrics
	InvoiceCreated *prometheus.CounterVec
	PaymentReceived *prometheus.CounterVec
	ExpenseRecorded *prometheus.CounterVec

	// HR metrics
	EmployeeCount    prometheus.Gauge
	PayrollProcessed *prometheus.CounterVec
	LeaveRequests    *prometheus.CounterVec
}

// TenantMetrics holds tenant-specific metrics
type TenantMetrics struct {
	TenantID string
	Metrics  *BusinessMetricTypes
}

// BusinessMetricsCollector manages business metrics collection
type BusinessMetricsCollector struct {
	metrics    *PrometheusMetrics
	business   *BusinessMetricTypes
	tenants    map[string]*TenantMetrics
	logger     *logger.Logger
	startTime  time.Time
}

// NewBusinessMetricsCollector creates a new business metrics collector
func NewBusinessMetricsCollector(baseMetrics *PrometheusMetrics, log *logger.Logger) *BusinessMetricsCollector {
	bmc := &BusinessMetricsCollector{
		metrics:   baseMetrics,
		tenants:   make(map[string]*TenantMetrics),
		logger:    log,
		startTime: time.Now(),
	}

	// Initialize business metrics structure without registering them
	bmc.initializeBusinessMetricsStructure()

	return bmc
}

// initializeBusinessMetricsStructure creates business metrics without registering them
func (bmc *BusinessMetricsCollector) initializeBusinessMetricsStructure() {
	serviceName := "rexi-erp"

	bmc.business = &BusinessMetricTypes{
		// E-commerce metrics
		OrdersCreated: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "orders_created_total",
				Help:        "Total number of orders created",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
			[]string{"tenant_id", "status", "channel"},
		),
		OrderValue: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "order_value_idr",
				Help:        "Order value in IDR",
				ConstLabels: prometheus.Labels{"service": serviceName, "currency": "IDR"},
				Buckets:     []float64{10000, 50000, 100000, 500000, 1000000, 5000000, 10000000, 50000000, 100000000},
			},
			[]string{"tenant_id", "channel"},
		),
		ProductsViewed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "products_viewed_total",
				Help:        "Total number of product views",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
			[]string{"tenant_id", "category", "product_type"},
		),
		CartAbandoned: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "carts_abandoned_total",
				Help:        "Total number of abandoned carts",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
			[]string{"tenant_id", "reason"},
		),
		PaymentTransactions: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "payment_transactions_total",
				Help:        "Total number of payment transactions",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
			[]string{"tenant_id", "method", "status"},
		),

		// User activity metrics
		UserLogins: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "user_logins_total",
				Help:        "Total number of user logins",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
			[]string{"tenant_id", "status", "method"},
		),
		UserRegistrations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "user_registrations_total",
				Help:        "Total number of user registrations",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
			[]string{"tenant_id", "role", "status"},
		),
		ActiveSessions: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:        "active_sessions_total",
				Help:        "Total number of active user sessions",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
		),

		// Inventory metrics
		StockLevels: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:        "stock_levels_total",
				Help:        "Total stock levels across all products",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
		),
		StockMovements: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "stock_movements_total",
				Help:        "Total number of stock movements",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
			[]string{"tenant_id", "type", "product_category"},
		),
		ProductChanges: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "product_changes_total",
				Help:        "Total number of product changes",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
			[]string{"tenant_id", "type"},
		),

		// Financial metrics
		InvoiceCreated: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "invoices_created_total",
				Help:        "Total number of invoices created",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
			[]string{"tenant_id", "type", "status"},
		),
		PaymentReceived: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "payments_received_total",
				Help:        "Total number of payments received",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
			[]string{"tenant_id", "method", "status"},
		),
		ExpenseRecorded: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "expenses_recorded_total",
				Help:        "Total number of expenses recorded",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
			[]string{"tenant_id", "category", "status"},
		),

		// HR metrics
		EmployeeCount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:        "employees_total",
				Help:        "Total number of employees",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
		),
		PayrollProcessed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "payrolls_processed_total",
				Help:        "Total number of payrolls processed",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
			[]string{"tenant_id", "period", "status"},
		),
		LeaveRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "leave_requests_total",
				Help:        "Total number of leave requests",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
			[]string{"tenant_id", "type", "status"},
		),
	}

	// Note: We don't register business metrics here to avoid test conflicts
	// In production, call RegisterBusinessMetrics() explicitly
}

// RegisterBusinessMetrics registers all business metrics with Prometheus
func (bmc *BusinessMetricsCollector) RegisterBusinessMetrics() {
	// E-commerce metrics
	prometheus.Register(bmc.business.OrdersCreated)
	prometheus.Register(bmc.business.OrderValue)
	prometheus.Register(bmc.business.ProductsViewed)
	prometheus.Register(bmc.business.CartAbandoned)
	prometheus.Register(bmc.business.PaymentTransactions)

	// User activity metrics
	prometheus.Register(bmc.business.UserLogins)
	prometheus.Register(bmc.business.UserRegistrations)
	prometheus.Register(bmc.business.ActiveSessions)

	// Inventory metrics
	prometheus.Register(bmc.business.StockLevels)
	prometheus.Register(bmc.business.StockMovements)
	prometheus.Register(bmc.business.ProductChanges)

	// Financial metrics
	prometheus.Register(bmc.business.InvoiceCreated)
	prometheus.Register(bmc.business.PaymentReceived)
	prometheus.Register(bmc.business.ExpenseRecorded)

	// HR metrics
	prometheus.Register(bmc.business.EmployeeCount)
	prometheus.Register(bmc.business.PayrollProcessed)
	prometheus.Register(bmc.business.LeaveRequests)
}

// registerBusinessMetrics registers all business metrics with Prometheus (private method)
func (bmc *BusinessMetricsCollector) registerBusinessMetrics() {
	// E-commerce metrics
	prometheus.Register(bmc.business.OrdersCreated)
	prometheus.Register(bmc.business.OrderValue)
	prometheus.Register(bmc.business.ProductsViewed)
	prometheus.Register(bmc.business.CartAbandoned)
	prometheus.Register(bmc.business.PaymentTransactions)

	// User activity metrics
	prometheus.Register(bmc.business.UserLogins)
	prometheus.Register(bmc.business.UserRegistrations)
	prometheus.Register(bmc.business.ActiveSessions)

	// Inventory metrics
	prometheus.Register(bmc.business.StockLevels)
	prometheus.Register(bmc.business.StockMovements)
	prometheus.Register(bmc.business.ProductChanges)

	// Financial metrics
	prometheus.Register(bmc.business.InvoiceCreated)
	prometheus.Register(bmc.business.PaymentReceived)
	prometheus.Register(bmc.business.ExpenseRecorded)

	// HR metrics
	prometheus.Register(bmc.business.EmployeeCount)
	prometheus.Register(bmc.business.PayrollProcessed)
	prometheus.Register(bmc.business.LeaveRequests)
}

// E-commerce Metrics Methods

// RecordOrderCreated records when an order is created
func (bmc *BusinessMetricsCollector) RecordOrderCreated(ctx context.Context, tenantID, status, channel string, orderValue float64, orderID string) {
	if tenantID == "" {
		tenantID = "unknown"
	}
	if status == "" {
		status = "pending"
	}
	if channel == "" {
		channel = "unknown"
	}

	bmc.business.OrdersCreated.WithLabelValues(tenantID, status, channel).Inc()
	bmc.business.OrderValue.WithLabelValues(tenantID, channel).Observe(orderValue)

	// Log significant orders
	if orderValue > 10000000 { // > 10 million IDR
		correlationID := logger.GetCorrelationID(ctx)
		logEntry := bmc.logger.WithRequestContext(correlationID, tenantID, "")
		logEntry.WithFields(map[string]interface{}{
			"order_id":        orderID,
			"order_value":     orderValue,
			"channel":         channel,
			"business_type":   "high_value_order",
		}).Info("High value order created")
	}
}

// RecordProductView records when a product is viewed
func (bmc *BusinessMetricsCollector) RecordProductView(ctx context.Context, tenantID, category, productType string) {
	if tenantID == "" {
		tenantID = "unknown"
	}
	if category == "" {
		category = "unknown"
	}
	if productType == "" {
		productType = "unknown"
	}

	bmc.business.ProductsViewed.WithLabelValues(tenantID, category, productType).Inc()
}

// RecordCartAbandoned records when a shopping cart is abandoned
func (bmc *BusinessMetricsCollector) RecordCartAbandoned(ctx context.Context, tenantID, reason string, cartValue float64) {
	if tenantID == "" {
		tenantID = "unknown"
	}
	if reason == "" {
		reason = "unknown"
	}

	bmc.business.CartAbandoned.WithLabelValues(tenantID, reason).Inc()

	// Log high-value abandoned carts
	if cartValue > 5000000 { // > 5 million IDR
		correlationID := logger.GetCorrelationID(ctx)
		logEntry := bmc.logger.WithRequestContext(correlationID, tenantID, "")
		logEntry.WithFields(map[string]interface{}{
			"cart_value":   cartValue,
			"reason":       reason,
			"business_type": "high_value_abandoned_cart",
		}).Warn("High value cart abandoned")
	}
}

// User Activity Metrics Methods

// RecordUserLogin records a user login event
func (bmc *BusinessMetricsCollector) RecordUserLogin(ctx context.Context, tenantID, status, method string, userID string) {
	if tenantID == "" {
		tenantID = "unknown"
	}
	if status == "" {
		status = "success"
	}
	if method == "" {
		method = "password"
	}

	bmc.business.UserLogins.WithLabelValues(tenantID, status, method).Inc()

	// Log failed login attempts
	if status != "success" {
		correlationID := logger.GetCorrelationID(ctx)
		logEntry := bmc.logger.WithRequestContext(correlationID, tenantID, userID)
		logEntry.WithFields(map[string]interface{}{
			"login_method": method,
			"security_type": "failed_login",
		}).Warn("Failed user login attempt")
	}
}

// RecordUserRegistration records a user registration
func (bmc *BusinessMetricsCollector) RecordUserRegistration(ctx context.Context, tenantID, role, status string, userID string) {
	if tenantID == "" {
		tenantID = "unknown"
	}
	if role == "" {
		role = "user"
	}
	if status == "" {
		status = "success"
	}

	bmc.business.UserRegistrations.WithLabelValues(tenantID, role, status).Inc()

	// Log new user registrations
	if status == "success" {
		correlationID := logger.GetCorrelationID(ctx)
		logEntry := bmc.logger.WithRequestContext(correlationID, tenantID, userID)
		logEntry.WithFields(map[string]interface{}{
			"role": role,
			"business_type": "user_registration",
		}).Info("New user registered")
	}
}

// UpdateActiveSessions updates the number of active sessions
func (bmc *BusinessMetricsCollector) UpdateActiveSessions(count float64) {
	bmc.business.ActiveSessions.Set(count)
}

// Inventory Metrics Methods

// RecordStockMovement records stock movement
func (bmc *BusinessMetricsCollector) RecordStockMovement(ctx context.Context, tenantID, movementType, productCategory string, quantity int) {
	if tenantID == "" {
		tenantID = "unknown"
	}
	if movementType == "" {
		movementType = "adjustment"
	}
	if productCategory == "" {
		productCategory = "unknown"
	}

	bmc.business.StockMovements.WithLabelValues(tenantID, movementType, productCategory).Inc()

	// Log significant stock movements
	if quantity > 1000 {
		correlationID := logger.GetCorrelationID(ctx)
		logEntry := bmc.logger.WithRequestContext(correlationID, tenantID, "")
		logEntry.WithFields(map[string]interface{}{
			"movement_type":   movementType,
			"quantity":        quantity,
			"product_category": productCategory,
			"business_type":   "large_stock_movement",
		}).Info("Large stock movement recorded")
	}
}

// UpdateStockLevels updates the total stock levels
func (bmc *BusinessMetricsCollector) UpdateStockLevels(count float64) {
	bmc.business.StockLevels.Set(count)
}

// Financial Metrics Methods

// RecordInvoiceCreated records invoice creation
func (bmc *BusinessMetricsCollector) RecordInvoiceCreated(ctx context.Context, tenantID, invoiceType, status string, amount float64, invoiceID string) {
	if tenantID == "" {
		tenantID = "unknown"
	}
	if invoiceType == "" {
		invoiceType = "sales"
	}
	if status == "" {
		status = "draft"
	}

	bmc.business.InvoiceCreated.WithLabelValues(tenantID, invoiceType, status).Inc()

	// Log high-value invoices
	if amount > 50000000 { // > 50 million IDR
		correlationID := logger.GetCorrelationID(ctx)
		logEntry := bmc.logger.WithRequestContext(correlationID, tenantID, "")
		logEntry.WithFields(map[string]interface{}{
			"invoice_id":   invoiceID,
			"invoice_type": invoiceType,
			"amount":       amount,
			"business_type": "high_value_invoice",
		}).Info("High value invoice created")
	}
}

// RecordPaymentReceived records payment receipt
func (bmc *BusinessMetricsCollector) RecordPaymentReceived(ctx context.Context, tenantID, method, status string, amount float64, paymentID string) {
	if tenantID == "" {
		tenantID = "unknown"
	}
	if method == "" {
		method = "unknown"
	}
	if status == "" {
		status = "pending"
	}

	bmc.business.PaymentReceived.WithLabelValues(tenantID, method, status).Inc()

	// Log payment events
	correlationID := logger.GetCorrelationID(ctx)
	logEntry := bmc.logger.WithRequestContext(correlationID, tenantID, "")
	logEntry.WithFields(map[string]interface{}{
		"payment_id":   paymentID,
		"method":       method,
		"amount":       amount,
		"business_type": "payment_received",
	}).Info("Payment received")
}

// HR Metrics Methods

// UpdateEmployeeCount updates the total employee count
func (bmc *BusinessMetricsCollector) UpdateEmployeeCount(count float64) {
	bmc.business.EmployeeCount.Set(count)
}

// RecordPayrollProcessed records payroll processing
func (bmc *BusinessMetricsCollector) RecordPayrollProcessed(ctx context.Context, tenantID, period, status string, employeeCount int) {
	if tenantID == "" {
		tenantID = "unknown"
	}
	if period == "" {
		period = "monthly"
	}
	if status == "" {
		status = "success"
	}

	bmc.business.PayrollProcessed.WithLabelValues(tenantID, period, status).Inc()

	// Log payroll processing
	correlationID := logger.GetCorrelationID(ctx)
	logEntry := bmc.logger.WithRequestContext(correlationID, tenantID, "")
	logEntry.WithFields(map[string]interface{}{
		"period":         period,
		"employee_count": employeeCount,
		"business_type":  "payroll_processed",
	}).Info("Payroll processed")
}

// RecordLeaveRequest records leave requests
func (bmc *BusinessMetricsCollector) RecordLeaveRequest(ctx context.Context, tenantID, leaveType, status string) {
	if tenantID == "" {
		tenantID = "unknown"
	}
	if leaveType == "" {
		leaveType = "annual"
	}
	if status == "" {
		status = "pending"
	}

	bmc.business.LeaveRequests.WithLabelValues(tenantID, leaveType, status).Inc()

	// Log leave requests
	correlationID := logger.GetCorrelationID(ctx)
	logEntry := bmc.logger.WithRequestContext(correlationID, tenantID, "")
	logEntry.WithFields(map[string]interface{}{
		"leave_type":    leaveType,
		"business_type": "leave_request",
	}).Info("Leave request recorded")
}

// GetTenantMetrics creates or returns tenant-specific metrics
func (bmc *BusinessMetricsCollector) GetTenantMetrics(tenantID string) *TenantMetrics {
	if tenantMetrics, exists := bmc.tenants[tenantID]; exists {
		return tenantMetrics
	}

	tenantMetrics := &TenantMetrics{
		TenantID: tenantID,
		Metrics:  bmc.business,
	}
	bmc.tenants[tenantID] = tenantMetrics
	return tenantMetrics
}

// GetMetricsSummary returns a summary of all business metrics
func (bmc *BusinessMetricsCollector) GetMetricsSummary() map[string]interface{} {
	uptime := time.Since(bmc.startTime).Seconds()

	return map[string]interface{}{
		"uptime_seconds":        uptime,
		"active_tenants":        len(bmc.tenants),
		"registered_tenant_ids": func() []string {
			tenantIDs := make([]string, 0, len(bmc.tenants))
			for tenantID := range bmc.tenants {
				tenantIDs = append(tenantIDs, tenantID)
			}
			return tenantIDs
		}(),
		"last_updated": time.Now().Format(time.RFC3339),
	}
}