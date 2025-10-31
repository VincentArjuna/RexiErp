# Error Handling Strategy

## General Approach

- **Error Model:** Structured error response with consistent format across all microservices
- **Exception Hierarchy:** Custom error types wrapped in standard HTTP response format
- **Error Propagation:** Errors propagated through RabbitMQ events with correlation IDs

## Logging Standards

- **Library:** Logrus 1.9.3 with structured JSON logging
- **Format:** JSON with correlation ID, service name, timestamp, and error context
- **Levels:** ERROR, WARN, INFO, DEBUG (DEBUG disabled in production)
- **Required Context:**
  - Correlation ID: UUID format for request tracing
  - Service Context: Service name and version
  - User Context: Tenant ID and User ID (when available)

## Error Handling Patterns

### External API Errors

- **Retry Policy:** Exponential backoff with jitter (100ms base, max 30s, 3 retries)
- **Circuit Breaker:** Opens after 5 consecutive failures, closes after 60s
- **Timeout Configuration:** 30s for Indonesian APIs, 10s for internal services
- **Error Translation:** Map Indonesian API errors to standardized internal error codes

### Business Logic Errors

- **Custom Exceptions:** ValidationError, BusinessRuleError, ComplianceError
- **User-Facing Errors:** Indonesian and English error messages
- **Error Codes:** Standardized error codes for mobile app handling

### Data Consistency

- **Transaction Strategy:** Database transactions with rollback on errors
- **Compensation Logic:** SAGA pattern for distributed transactions
- **Idempotency:** Idempotent operation keys for retry safety
