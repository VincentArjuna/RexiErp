package logger

import (
	"github.com/sirupsen/logrus"
)

// Standard field names used across all services for consistency
const (
	// Request and Context Fields
	FieldCorrelationID = "correlation_id"
	FieldTenantID      = "tenant_id"
	FieldUserID        = "user_id"
	FieldService       = "service"
	FieldVersion       = "version"
	FieldEnvironment   = "environment"

	// HTTP Fields
	FieldMethod       = "method"
	FieldPath         = "path"
	FieldQuery        = "query"
	FieldStatusCode   = "status_code"
	FieldClientIP     = "client_ip"
	FieldUserAgent    = "user_agent"
	FieldRequestSize  = "request_size"
	FieldResponseSize = "response_size"
	FieldLatency      = "latency"
	FieldLatencyMs    = "latency_ms"

	// Database Fields
	FieldDatabase    = "database"
	FieldTable       = "table"
	FieldOperation   = "operation"
	FieldDuration    = "duration"
	FieldRowsAffected = "rows_affected"
	FieldDBQuery     = "db_query"
	FieldQueryParams = "query_params"

	// Business Fields
	FieldTransactionID = "transaction_id"
	FieldOrderID       = "order_id"
	FieldProductID     = "product_id"
	FieldCustomerID    = "customer_id"
	FieldInvoiceID     = "invoice_id"
	FieldPaymentID     = "payment_id"
	FieldAmount        = "amount"
	FieldCurrency      = "currency"

	// System Fields
	FieldName       = "name"
	FieldType       = "type"
	FieldComponent   = "component"
	FieldStatus      = "status"
	FieldError       = "error"
	FieldStackTrace  = "stack_trace"
	FieldHost        = "host"
	FieldPort        = "port"
	FieldProcessID   = "process_id"
	FieldGoroutineID = "goroutine_id"

	// External Service Fields
	FieldExternalService = "external_service"
	FieldEndpoint        = "endpoint"
	FieldAttempt         = "attempt"
	FieldMaxAttempts     = "max_attempts"
	FieldTimeout         = "timeout"

	// Security Fields
	FieldAction        = "action"
	FieldResource      = "resource"
	FieldPermission    = "permission"
	FieldRole          = "role"
	FieldSourceIP      = "source_ip"
	FieldDestinationIP = "destination_ip"

	// Performance Fields
	FieldMemoryUsage   = "memory_usage"
	FieldCPUUsage      = "cpu_usage"
	FieldGoroutineCount = "goroutine_count"
	FieldGCGen         = "gc_gen"
	FieldGCPause       = "gc_pause"
)

// ServiceFields returns standard service context fields
func ServiceFields(serviceName, version, environment string) logrus.Fields {
	return logrus.Fields{
		FieldService:     serviceName,
		FieldVersion:     version,
		FieldEnvironment: environment,
	}
}

// RequestContextFields returns standard HTTP request context fields
func RequestContextFields(correlationID, tenantID, userID string) logrus.Fields {
	fields := logrus.Fields{
		FieldCorrelationID: correlationID,
	}

	if tenantID != "" {
		fields[FieldTenantID] = tenantID
	}
	if userID != "" {
		fields[FieldUserID] = userID
	}

	return fields
}

// HTTPRequestFields returns standard HTTP request fields
func HTTPRequestFields(method, path, query, clientIP, userAgent string) logrus.Fields {
	fields := logrus.Fields{
		FieldMethod:    method,
		FieldPath:      path,
		FieldClientIP:  clientIP,
		FieldUserAgent: userAgent,
	}

	if query != "" {
		fields[FieldQuery] = query
	}

	return fields
}

// HTTPResponseFields returns standard HTTP response fields
func HTTPResponseFields(statusCode, responseSize int, latencyMs int64) logrus.Fields {
	return logrus.Fields{
		FieldStatusCode:   statusCode,
		FieldResponseSize: responseSize,
		FieldLatencyMs:    latencyMs,
	}
}

// DatabaseFields returns standard database operation fields
func DatabaseFields(database, table, operation string, duration int64, rowsAffected int) logrus.Fields {
	fields := logrus.Fields{
		FieldDatabase:   database,
		FieldTable:      table,
		FieldOperation:  operation,
		FieldDuration:   duration,
	}

	if rowsAffected >= 0 {
		fields[FieldRowsAffected] = rowsAffected
	}

	return fields
}

// BusinessTransactionFields returns standard business transaction fields
func BusinessTransactionFields(transactionType, transactionID string) logrus.Fields {
	return logrus.Fields{
		FieldComponent:     "business",
		"transaction_type": transactionType,
		FieldTransactionID: transactionID,
	}
}

// ExternalServiceFields returns standard external service call fields
func ExternalServiceFields(service, endpoint, attempt int) logrus.Fields {
	return logrus.Fields{
		FieldExternalService: service,
		FieldEndpoint:        endpoint,
		FieldAttempt:         attempt,
		FieldComponent:       "external_service",
	}
}

// SecurityFields returns standard security event fields
func SecurityFields(action, resource, userID, sourceIP string) logrus.Fields {
	fields := logrus.Fields{
		FieldComponent: "security",
		FieldAction:    action,
		FieldResource:  resource,
		FieldSourceIP:  sourceIP,
	}

	if userID != "" {
		fields[FieldUserID] = userID
	}

	return fields
}

// PerformanceFields returns standard performance monitoring fields
func PerformanceFields(memoryUsage, cpuUsage float64, goroutineCount int) logrus.Fields {
	return logrus.Fields{
		FieldComponent:      "performance",
		FieldMemoryUsage:    memoryUsage,
		FieldCPUUsage:       cpuUsage,
		FieldGoroutineCount: goroutineCount,
	}
}

// ErrorFields returns standard error fields
func ErrorFields(err error, stackTrace string) logrus.Fields {
	fields := logrus.Fields{
		FieldError: err.Error(),
	}

	if stackTrace != "" {
		fields[FieldStackTrace] = stackTrace
	}

	return fields
}

// ComponentFields returns fields for component-specific logging
func ComponentFields(component, name, status string) logrus.Fields {
	return logrus.Fields{
		FieldComponent: component,
		FieldName:      name,
		FieldStatus:    status,
	}
}