# Coding Standards

**IMPORTANT:** These standards directly control AI developer behavior. Only include critical rules needed to prevent bad code.

## Core Standards

- **Languages & Runtimes:** Go 1.23.1 with Docker containerization
- **Style & Linting:** golangci-lint with custom configuration
- **Test Organization:** Unit tests in `*_test.go` files, integration tests in `tests/integration/`

## Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Package | lowercase, short, descriptive | `inventory`, `accounting` |
| Interface | -er suffix | `ProductRepository`, `InvoiceService` |
| Struct | PascalCase, descriptive | `Product`, `SalesOrder` |
| Function | PascalCase for exported, camelCase for unexported | `CreateProduct()`, `calculateTax()` |
| Variable | camelCase, meaningful | `productService`, `taxRate` |
| Constant | UPPER_SNAKE_CASE | `MAX_RETRY_ATTEMPTS`, `DEFAULT_TAX_RATE` |

## Critical Rules

- **No hardcoded secrets:** Use environment variables or configuration service only
- **Tenant context required:** All database operations must include tenant isolation
- **Error wrapping:** Use `fmt.Errorf` with context, never use bare `errors.New`
- **Structured logging:** Use Logrus with consistent field names across services
- **Input validation:** Validate all external inputs at service boundaries
- **No console.log in production:** Use structured logger instead
- **Repository pattern:** All database access through repository interfaces
- **Context propagation:** Always pass context parameter through function chains
- **Indonesian timezone:** Use Asia/Jakarta timezone for business operations
- **Tax calculations:** All financial calculations use decimal arithmetic, never float
