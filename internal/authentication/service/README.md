# Authentication Service

This package provides comprehensive authentication and authorization functionality for RexiERP.

## 🛡️ Features

- **User Authentication**: Login, logout, and session management
- **JWT Token Management**: Access and refresh token generation/validation
- **Password Security**: Strong password policies and secure storage
- **Multi-tenant Support**: Tenant-isolated authentication
- **Account Security**: Lockout protection and activity logging
- **Profile Management**: User profile updates and password changes

## 📁 Package Structure

```
service/
├── types.go          # ALL types, interfaces, and structs (CENTRALIZED)
├── config.go         # Configuration factory functions only
├── auth_service.go   # Main authentication service implementation
├── jwt_service.go    # JWT token service implementation
├── errors.go         # Service-specific error types (future)
├── validators.go     # Input validation functions (future)
├── README.md         # This documentation
└── PACKAGE_STRUCTURE.md # Architectural guidelines
```

## 🔒 Type Resolution Prevention

### ✅ **PROBLEM SOLVED**
All types are now centralized in `types.go` to prevent IDE resolution issues and circular dependencies.

### 📋 **RULES TO FOLLOW**
1. **NEVER** define types in multiple files
2. **ALWAYS** add new types to `types.go` first
3. **UPDATE** documentation when adding new types
4. **VERIFY** with `go build ./...` after changes

## 🚀 Quick Start

```go
// Create service configuration
config := NewAuthConfig(authServiceConfig)

// Create service instances
jwtService := NewJWTService(config)
authService := NewAuthService(userRepo, sessionRepo, activityRepo, cache, jwtService, logger, config)

// Register a new user
user, err := authService.Register(ctx, &RegisterRequest{
    Email:    "user@company.com",
    Password: "SecurePass123!",
    FullName: "John Doe",
    // ... other fields
})

// Login user
response, err := authService.Login(ctx, &LoginRequest{
    Email:    "user@company.com",
    Password: "SecurePass123!",
}, clientIP, userAgent)
```

## 🔧 Configuration

See `AuthConfig` in `types.go` for all available configuration options:

- Password policy settings
- Account lockout thresholds
- Token lifetime configuration
- JWT signing settings

## 📚 API Documentation

### AuthService Interface
- `Register()` - User registration
- `Login()` - User authentication
- `Logout()` - Session termination
- `RefreshToken()` - Token renewal
- `ValidateToken()` - Token validation
- `GetProfile()` - User profile retrieval
- `UpdateProfile()` - Profile updates
- `ChangePassword()` - Password changes

### JWTService Interface
- `GenerateTokenPair()` - Create access/refresh tokens
- `GenerateAccessToken()` - Create access token only
- `GenerateRefreshToken()` - Create refresh token only
- `ValidateToken()` - Token validation and parsing
- `ExtractTokenFromHeader()` - Extract token from Authorization header

## 🔍 Type Reference

### Core Types
- `User` - Simplified user model for JWT operations
- `TokenValidationResult` - Token validation response
- `TokenClaims` - JWT token claim structure
- `AuthConfig` - Service configuration

### Request/Response Types
- `RegisterRequest` - User registration input
- `LoginRequest` - User login credentials
- `UpdateProfileRequest` - Profile update data
- `ChangePasswordRequest` - Password change data
- `AuthResponse` - Authentication response with tokens

## 🛡️ Security Features

- **Password Hashing**: bcrypt with configurable cost
- **Token Security**: JWT with RS256 signing (configurable)
- **Session Management**: Secure session IDs with TTL
- **Account Lockout**: Configurable attempt thresholds
- **Activity Logging**: Comprehensive audit trail
- **Input Validation**: Request validation and sanitization

## 🧪 Testing

Run tests for the service package:

```bash
go test ./internal/authentication/service/...
```

## 📖 Dependencies

- `github.com/golang-jwt/jwt/v5` - JWT token handling
- `github.com/google/uuid` - UUID generation
- `golang.org/x/crypto/bcrypt` - Password hashing
- `github.com/sirupsen/logrus` - Structured logging

## 🔄 Maintenance

- Regular security updates for JWT library
- Periodic password policy reviews
- Token rotation and refresh strategies
- Performance monitoring and optimization