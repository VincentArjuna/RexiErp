# Test Strategy and Standards

## Testing Philosophy

- **Approach:** Test-driven development with 80% unit coverage, integration tests for critical paths
- **Coverage Goals:** 80% unit test coverage, 100% for Indonesian tax calculations and financial operations
- **Test Pyramid:** 70% unit tests, 20% integration tests, 10% end-to-end tests

## Test Types and Organization

### Unit Tests

- **Framework:** Go's built-in testing package with testify 1.9.0
- **File Convention:** `*_test.go` files alongside source files
- **Location:** Same package as source code
- **Mocking Library:** testify/mock and gomock for external dependencies
- **Coverage Requirement:** 80% minimum, 95% for financial calculations

**AI Agent Requirements:**
- Generate tests for all public methods
- Cover edge cases and error conditions
- Follow AAA pattern (Arrange, Act, Assert)
- Mock all external dependencies (database, Redis, RabbitMQ, external APIs)
- Test Indonesian tax calculation edge cases (PPN rates, NPWP validation)
- Test multi-tenant data isolation

### Integration Tests

- **Scope:** Database operations, Redis caching, RabbitMQ messaging, external API integrations
- **Location:** `tests/integration/`
- **Test Infrastructure:**
  - **PostgreSQL:** Testcontainers PostgreSQL for integration tests
  - **Redis:** Testcontainers Redis for caching tests
  - **RabbitMQ:** Testcontainers RabbitMQ for messaging tests
  - **External APIs:** WireMock for stubbing Indonesian government APIs

### End-to-End Tests

- **Framework:** Go's testing package with HTTP client simulation
- **Scope:** Critical business workflows (sales orders with e-Faktur, payroll with BPJS)
- **Environment:** Staging environment with real Indonesian API integration
- **Test Data:** Seed test data with Indonesian business scenarios

## Test Data Management

- **Strategy:** Factory pattern with Go structs for test data generation
- **Fixtures:** `tests/fixtures/` directory with Indonesian business scenarios
- **Factories:** Data factory functions for realistic Indonesian business data
- **Cleanup:** Automatic cleanup after each test using testify/suite

## Continuous Testing

- **CI Integration:** GitHub Actions with parallel test execution
- **Performance Tests:** Go benchmarks for critical paths, k6 for load testing
- **Security Tests:** Gosec for static analysis, OWASP ZAP for dynamic analysis
