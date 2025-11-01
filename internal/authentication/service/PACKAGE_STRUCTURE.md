# Authentication Service Package Structure

This document outlines the organizational structure to prevent type resolution issues.

## Package: `service`

### File Organization Philosophy

1. **types.go** - ALL shared data structures and interfaces
2. **config.go** - Configuration and factory functions only
3. **auth_service.go** - Main service implementation
4. **jwt_service.go** - JWT service implementation
5. **errors.go** - Service-specific error types
6. **validators.go** - Input validation functions

### Type Definition Rules

- ✅ **DO**: Define all types in `types.go`
- ✅ **DO**: Use consistent naming conventions
- ✅ **DO**: Add comprehensive documentation
- ❌ **DON'T**: Define the same type in multiple files
- ❌ **DON'T**: Reference types before they're defined

### Import Management

- Keep imports minimal and focused
- Group imports by: stdlib, external, internal
- Remove unused imports immediately

### Documentation Standards

- All public types must have godoc comments
- Include JSON field tags with examples
- Document validation rules and constraints

## Future Prevention Checklist

1. **Before adding new types**: Check if they already exist in `types.go`
2. **When modifying types**: Update all references and tests
3. **After changes**: Run `go build ./...` to verify no breaking changes
4. **Regular cleanup**: Remove unused imports and type definitions

This structure ensures IDE compatibility and prevents circular dependencies.