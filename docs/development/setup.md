# Development Setup Guide

This guide provides detailed instructions for setting up a development environment for RexiERP.

## Prerequisites

### Required Software
- **Go 1.23.1+**: [Download Go](https://go.dev/dl/)
- **Docker 27.3.1+**: [Download Docker](https://docs.docker.com/get-docker/)
- **Docker Compose 2.29.0+**: [Install Docker Compose](https://docs.docker.com/compose/install/)
- **Make**: Usually pre-installed on Linux/macOS
- **Git**: [Download Git](https://git-scm.com/)

### Development Tools (Optional)
- **VS Code**: Recommended IDE with Go extension
- **Postman**: API testing
- **DBeaver**: Database management
- **Redis Desktop Manager**: Redis GUI

## Quick Setup

### 1. Clone Repository
```bash
git clone https://github.com/VincentArjuna/RexiErp.git
cd RexiErp
```

### 2. One-Command Setup
```bash
# Complete development environment setup
make dev-setup
```

This command will:
- Install Go development tools
- Set up pre-commit hooks
- Verify environment configuration
- Start infrastructure services
- Run database migrations

### 3. Start Development
```bash
# Start all services
make up

# Verify installation
make health
```

## Manual Setup

### 1. Environment Configuration
```bash
# Copy environment template
cp .env.example .env

# Edit configuration
nano .env
```

Key environment variables to configure:
```bash
# Development environment
APP_ENV=development
LOG_LEVEL=debug

# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=rexi_erp
DB_USER=rexi
DB_PASSWORD=password

# Redis configuration
REDIS_HOST=localhost
REDIS_PORT=6379

# RabbitMQ configuration
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest

# JWT configuration
JWT_SECRET=your-super-secret-jwt-key-for-development-only
JWT_EXPIRATION_HOURS=24
```

### 2. Install Development Tools
```bash
# Install Go development tools
make install-tools

# Install pre-commit hooks
make pre-commit-install
```

### 3. Start Infrastructure Services
```bash
# Start databases and message queue
docker-compose -f deployments/docker-compose/docker-compose.yml up -d postgres redis rabbitmq

# Wait for services to be ready
sleep 10
```

### 4. Database Setup
```bash
# Run migrations and seed data
./scripts/migrate.sh init

# Verify database connection
./scripts/migrate.sh status
```

### 5. Build and Run Services
```bash
# Build all services
make build

# Run individual service
make run-service SERVICE=authentication-service
```

## IDE Configuration

### VS Code Setup

#### Extensions
```json
{
  "recommendations": [
    "golang.go",
    "ms-vscode.vscode-json",
    "redhat.vscode-yaml",
    "ms-vscode.docker",
    "bradlc.vscode-tailwindcss",
    "esbenp.prettier-vscode"
  ]
}
```

#### Settings
```json
{
  "go.toolsManagement.checkForUpdates": "local",
  "go.useLanguageServer": true,
  "go.gopath": "",
  "go.goroot": "",
  "go.formatTool": "goimports",
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "go.testOnSave": false,
  "go.coverOnSave": false,
  "go.coverageDecorator": {
    "type": "gutter",
    "coveredHighlightColor": "rgba(64,128,64,0.5)",
    "uncoveredHighlightColor": "rgba(128,64,64,0.25)"
  },
  "files.exclude": {
    "**/bin": true,
    "**/vendor": true,
    "**/*.exe": true,
    "**/coverage.out": true,
    "**/coverage.html": true
  }
}
```

### GoLand Setup
1. Open project directory
2. Go to Settings → Go → GOROOT → Set to Go 1.23.1
3. Go to Settings → Go → GOPATH → Set to project directory
4. Enable Go Modules integration
5. Configure code formatting: goimports + gofmt

## Development Workflow

### 1. Daily Development
```bash
# Start services
make up

# Check status
make status

# View logs
make logs

# Run tests
make test

# Format code
make fmt

# Run linting
make lint
```

### 2. Making Changes
```bash
# Create feature branch
git checkout -b feature/new-feature

# Make changes
# ... (edit files)

# Run quality checks
make dev-check

# Run tests
make test-coverage

# Commit changes
git add .
git commit -m "feat: add new feature"

# Push and create PR
git push origin feature/new-feature
```

### 3. Testing Changes
```bash
# Run unit tests
make test

# Run integration tests
make test-integration

# Run specific test
go test -v ./internal/authentication/service/

# Run tests with coverage
make test-coverage
```

## Database Development

### Migration Management
```bash
# Create new migration
echo "-- Add your migration SQL here" > migrations/master/003_new_feature.sql

# Run migration
./scripts/migrate.sh migrate

# Rollback migration
./scripts/migrate.sh reset  # ⚠️ destructive
```

### Database Access
```bash
# Connect to PostgreSQL
docker exec -it rexi-postgres psql -U rexi -d rexi_erp

# Connect to Redis
docker exec -it rexi-redis redis-cli

# Access RabbitMQ Management
open http://localhost:15672
```

### Development Data
The project includes Indonesian business scenario seed data:
- **Companies**: MSMEs from major Indonesian cities
- **Products**: Indonesian business products with tax configuration
- **Customers/Suppliers**: Indonesian addresses and tax numbers
- **Chart of Accounts**: SAK-compliant accounting structure

## API Development

### Running Services
```bash
# Run authentication service
make run-service SERVICE=authentication-service

# Run with environment variables
APP_PORT=8001 go run ./cmd/authentication-service/
```

### Testing APIs
```bash
# Test health endpoint
curl http://localhost:8001/health

# Test authentication
curl -X POST http://localhost:8001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@majujaya.com","password":"password"}'
```

### API Documentation
- **Swagger UI**: http://localhost:8001/docs (when service is running)
- **OpenAPI Spec**: http://localhost:8001/docs/swagger.json

## Debugging

### Service Debugging
```bash
# Run with debug flags
go run -tags debug ./cmd/authentication-service/

# Enable debug logging
LOG_LEVEL=debug make run-service SERVICE=authentication-service
```

### Docker Debugging
```bash
# View container logs
docker logs rexi-authentication-service

# Access container shell
docker exec -it rexi-authentication-service sh

# Inspect container
docker inspect rexi-postgres
```

### Database Debugging
```bash
# Check database connections
./scripts/migrate.sh status

# View query logs
docker logs rexi-postgres | grep "statement\|ERROR"

# Connect to database
docker exec -it rexi-postgres psql -U rexi -d rexi_erp
```

## Performance Monitoring

### Local Monitoring
```bash
# Access monitoring tools
# Grafana: http://localhost:3000 (admin/admin)
# Prometheus: http://localhost:9090

# View service metrics
curl http://localhost:8001/metrics
```

### Performance Testing
```bash
# Install hey (HTTP load testing)
go install github.com/rakyll/hey@latest

# Load test API endpoint
hey -n 1000 -c 10 http://localhost:8001/health
```

## Troubleshooting

### Common Issues

#### Port Conflicts
```bash
# Check what's using ports
netstat -tulpn | grep :8080

# Kill processes
sudo kill -9 <PID>
```

#### Docker Issues
```bash
# Reset Docker environment
make docker-clean
docker system prune -a

# Rebuild containers
make docker-build
make up
```

#### Database Issues
```bash
# Reset database
./scripts/migrate.sh reset

# Check PostgreSQL logs
docker logs rexi-postgres

# Recreate database
docker-compose -f deployments/docker-compose/docker-compose.yml down postgres
docker-compose -f deployments/docker-compose/docker-compose.yml up -d postgres
```

#### Go Build Issues
```bash
# Clean module cache
go clean -modcache

# Update dependencies
make update-deps

# Verify module
go mod verify
```

### Getting Help
- Check logs: `make logs`
- Check service status: `make status`
- Run health checks: `make health`
- Review environment: `make env-check`

## Development Best Practices

### Code Quality
- Always run `make dev-check` before committing
- Write comprehensive tests
- Follow Go naming conventions
- Use structured logging with correlation IDs

### Git Workflow
- Use feature branches
- Write conventional commit messages
- Keep PRs focused and small
- Update documentation with changes

### Security
- Never commit secrets
- Use environment variables for configuration
- Validate all inputs
- Follow OWASP guidelines

### Performance
- Use connection pooling
- Implement proper caching
- Monitor resource usage
- Optimize database queries

## Next Steps

After setting up development:

1. **Explore the codebase**: Read through the architecture documentation
2. **Run the tests**: `make test`
3. **Try the APIs**: Use Postman or curl to test endpoints
4. **Add a feature**: Follow the contribution guidelines
5. **Read the documentation**: Explore the docs directory

Happy coding! 🚀