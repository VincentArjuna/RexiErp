# Security

## Input Validation

- **Validation Library:** Go's validator package with custom Indonesian validators
- **Validation Location:** API gateway layer and service boundaries
- **Required Rules:**
  - All external inputs MUST be validated
  - Validation at API boundary before processing
  - Whitelist approach preferred over blacklist
  - Indonesian phone number validation (+62 format)
  - NPWP (tax ID) format validation
  - Indonesian email domain validation

## Authentication & Authorization

- **Auth Method:** JWT tokens with refresh token rotation
- **Session Management:** Redis-based session store with configurable TTL
- **Required Patterns:**
  - JWT signed with RS256 keys
  - Role-based access control (RBAC) with tenant context
  - API key authentication for Indonesian government integrations
  - Multi-factor authentication for admin users

## Secrets Management

- **Development:** Environment variables with .env file (gitignored)
- **Production:** AWS Secrets Manager / Azure Key Vault / Google Secret Manager
- **Code Requirements:**
  - NEVER hardcode secrets
  - Access via configuration service only
  - No secrets in logs or error messages
  - Automatic secret rotation support

## API Security

- **Rate Limiting:** Redis-based rate limiting with configurable rules per endpoint
- **CORS Policy:** Restrictive CORS with Indonesian domain whitelist
- **Security Headers:** Strict-Transport-Security, X-Content-Type-Options, X-Frame-Options
- **HTTPS Enforcement:** Mandatory HTTPS in production, HTTP/2 support

## Data Protection

- **Encryption at Rest:** PostgreSQL Transparent Data Encryption (TDE) or application-level encryption
- **Encryption in Transit:** TLS 1.3 for all communications
- **PII Handling:** Encrypted storage for sensitive Indonesian personal data
- **Logging Restrictions:** No PII or financial data in application logs

## Dependency Security

- **Scanning Tool:** Go's built-in vulnerability scanner + Snyk
- **Update Policy:** Weekly dependency scans, immediate patching for critical vulnerabilities
- **Approval Process:** Security review for new external dependencies

## Security Testing

- **SAST Tool:** Gosec for static analysis
- **DAST Tool:** OWASP ZAP for dynamic security testing
- **Penetration Testing:** Quarterly penetration testing with focus on Indonesian compliance
