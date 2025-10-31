# RexiERP - Indonesian ERP System for MSMEs

![Go Version](https://img.shields.io/badge/Go-1.23.1-blue)
![License](https://img.shields.io/badge/License-MIT-green)
![Docker](https://img.shields.io/badge/Docker-27.3.1-blue)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16.6-blue)
![Redis](https://img.shields.io/badge/Redis-7.2.5-red)

**RexiERP** is a comprehensive Enterprise Resource Planning (ERP) system specifically designed for Indonesian Micro, Small, and Medium Enterprises (MSMEs). Built with Go microservices architecture, it provides robust business management capabilities compliant with Indonesian regulations and business practices.

## 🚀 Features

### Core Business Modules
- **Authentication & Authorization**: Multi-tenant user management with RBAC
- **Inventory Management**: Stock control, warehouses, and product tracking
- **Accounting**: SAK-compliant financial management with Indonesian tax support
- **Human Resources**: Employee management with BPJS integration
- **CRM**: Customer relationship management
- **Notifications**: Multi-channel notifications (SMS, Email, WhatsApp)
- **Integration**: Indonesian government API integrations (e-Faktur, e-Invoice, BPJS)

### Technical Features
- **Multi-tenant Architecture**: Support multiple companies from single deployment
- **Microservices**: Scalable, independent service deployment
- **Real-time Processing**: Event-driven architecture with RabbitMQ
- **Caching**: Redis-based performance optimization
- **Monitoring**: Prometheus + Grafana observability
- **API Gateway**: Nginx-based gateway with rate limiting
- **Security**: JWT authentication, encrypted data transmission

## 🏗️ Architecture

### Technology Stack
- **Language**: Go 1.23.1
- **Web Framework**: Gin 1.9.1
- **Database**: PostgreSQL 16.6
- **Cache**: Redis 7.2.5
- **Message Queue**: RabbitMQ 3.13.6
- **API Gateway**: Nginx 1.25.5
- **Monitoring**: Prometheus 2.53.2 + Grafana 11.1.0
- **Containerization**: Docker 27.3.1 + Docker Compose 2.29.0

### Microservices
```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  API Gateway    │    │  Authentication  │    │    Inventory    │
│     Nginx       │    │     Service      │    │     Service     │
│   (Port 8080)   │◄──►│   (Port 8001)    │◄──►│   (Port 8002)   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Accounting    │    │       HR         │    │       CRM       │
│     Service     │    │     Service      │    │     Service     │
│   (Port 8003)   │◄──►│   (Port 8004)    │◄──►│   (Port 8005)   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  Notification   │    │   Integration    │    │   Data Stores   │
│     Service     │    │     Service      │    │ PostgreSQL,     │
│   (Port 8006)   │◄──►│   (Port 8007)    │◄──►│ Redis, RabbitMQ │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## 🛠️ Quick Start

### Prerequisites
- [Go 1.23.1+](https://go.dev/dl/)
- [Docker 27.3.1+](https://docs.docker.com/get-docker/)
- [Docker Compose 2.29.0+](https://docs.docker.com/compose/install/)
- [Make](https://www.gnu.org/software/make/)

### 1. Clone Repository
```bash
git clone https://github.com/VincentArjuna/RexiErp.git
cd RexiErp
```

### 2. Environment Setup
```bash
# Copy environment template
cp .env.example .env

# Edit environment variables (optional for development)
nano .env
```

### 3. Development Environment Setup
```bash
# Complete development setup (installs tools, sets up pre-commit hooks)
make dev-setup

# Or manual setup:
make install-tools  # Install development tools
make pre-commit-install  # Install pre-commit hooks
```

### 4. Start Development Environment
```bash
# Start all services (infrastructure + applications)
make up

# Or start infrastructure only
docker-compose -f deployments/docker-compose/docker-compose.yml up -d postgres redis rabbitmq
```

### 5. Initialize Database
```bash
# Run database migrations and seed data
make migrate-up  # or: ./scripts/migrate.sh init
```

### 6. Verify Installation
```bash
# Check service health
make health

# Access services
# API Gateway: http://localhost:8080/health
# Grafana: http://localhost:3000 (admin/admin)
# Prometheus: http://localhost:9090
# RabbitMQ Management: http://localhost:15672 (guest/guest)
```

## 📋 Available Commands

### Development Commands
```bash
# Setup and tools
make dev-setup          # Complete development environment setup
make install-tools      # Install development tools
make pre-commit-install # Install pre-commit hooks

# Building and running
make build              # Build all services
make build-service SERVICE=authentication-service  # Build specific service
make run-service SERVICE=authentication-service    # Run specific service

# Testing and quality
make test               # Run all tests
make test-coverage      # Run tests with coverage report
make test-integration   # Run integration tests
make dev-check          # Run all development checks
make ci                 # Run CI pipeline locally

# Code quality
make fmt                # Format Go code
make lint               # Run linter
make security           # Run security scan
make vet                # Run go vet
```

### Docker Commands
```bash
make up                 # Start all services
make down               # Stop all services
make restart            # Restart all services
make logs               # Show logs from all services
make logs-service SERVICE=postgres  # Show logs from specific service
make status             # Show service status
make docker-build       # Build Docker images
make docker-clean       # Clean Docker resources
```

### Database Commands
```bash
make migrate-up         # Run database migrations
make db-seed            # Seed database with test data
make backup-db          # Create database backup
```

### Utility Commands
```bash
make help               # Show all available commands
make version            # Show version information
make stats              # Show project statistics
make health             # Check service health
make env-check          # Check environment configuration
```

## 🏁 Local Development

### Directory Structure
```
RexiErp/
├── cmd/                    # Service entry points
│   ├── authentication-service/
│   ├── inventory-service/
│   ├── accounting-service/
│   ├── hr-service/
│   ├── crm-service/
│   ├── notification-service/
│   └── integration-service/
├── internal/               # Private application code
│   ├── authentication/
│   ├── inventory/
│   ├── accounting/
│   ├── hr/
│   ├── crm/
│   ├── notification/
│   ├── integration/
│   └── shared/            # Shared utilities
├── pkg/                    # Public library code
├── migrations/             # Database migrations
├── configs/                # Configuration files
├── deployments/            # Deployment configurations
│   └── docker-compose/    # Local development setup
├── scripts/               # Build and utility scripts
├── tests/                 # Test files
├── docs/                  # Documentation
└── tools/                 # Development tools
```

### Configuration
Configuration is managed through environment variables and YAML files:

1. **Environment Variables**: Copy `.env.example` to `.env` and modify as needed
2. **YAML Config**: Service-specific configs in `configs/{environment}/`
3. **Runtime Config**: Loaded with precedence: CLI flags > ENV > YAML > defaults

### Database Management
```bash
# Migration management
./scripts/migrate.sh status     # Show migration status
./scripts/migrate.sh migrate    # Run pending migrations
./scripts/migrate.sh reset      # Reset database (⚠️ destructive)
./scripts/migrate.sh backup     # Create backup
./scripts/migrate.sh restore <backup_file>  # Restore from backup
```

### Testing
```bash
# Run all tests
make test

# Run specific test
go test ./internal/authentication/service/

# Run tests with coverage
make test-coverage

# Run integration tests
make test-integration
```

### Code Quality
Pre-commit hooks ensure code quality:
- **Go fmt**: Code formatting
- **Go vet**: Static analysis
- **Golangci-lint**: Comprehensive linting
- **Gosec**: Security scanning
- **Go test**: Run tests

## 🐳 Docker Deployment

### Development Environment
```bash
# Start complete development stack
make up

# View logs
make logs

# Stop services
make down
```

### Production Deployment
For production deployment, see [Deployment Guide](docs/deployment/production.md).

## 📊 Monitoring & Observability

### Metrics
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **Service Metrics**: Available at `http://localhost:{service_port}/metrics`

### Logging
Structured JSON logging with Logrus:
- **Development**: Human-readable format
- **Production**: JSON format with correlation IDs
- **Levels**: DEBUG, INFO, WARN, ERROR, FATAL

### Health Checks
All services expose health endpoints:
- **API Gateway**: `/health`
- **Services**: `/health`
- **Database**: Connection health checks
- **External Dependencies**: Health status monitoring

## 🔧 Configuration Reference

### Environment Variables
Key environment variables (see `.env.example` for complete list):

```bash
# Application
APP_ENV=development
LOG_LEVEL=debug

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=rexi_erp
DB_USER=rexi
DB_PASSWORD=password

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# RabbitMQ
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest

# JWT
JWT_SECRET=your-super-secret-jwt-key
JWT_EXPIRATION_HOURS=24

# Indonesian APIs
EFAKTUR_API_URL=https://api.efaktur.pajak.go.id
BPJS_API_URL=https://api.bpjs-kesehatan.go.id
EINVOICE_API_URL=https://api.einvoice.pajak.go.id
```

### Service Ports
- **API Gateway**: 8080 (HTTP), 8443 (HTTPS)
- **Authentication Service**: 8001
- **Inventory Service**: 8002
- **Accounting Service**: 8003
- **HR Service**: 8004
- **CRM Service**: 8005
- **Notification Service**: 8006
- **Integration Service**: 8007
- **Prometheus**: 9090
- **Grafana**: 3000
- **RabbitMQ Management**: 15672

## 🌐 API Documentation

### Authentication
```bash
# Login
POST /api/v1/auth/login
{
  "email": "admin@company.com",
  "password": "password"
}

# Register
POST /api/v1/auth/register
{
  "email": "user@company.com",
  "password": "password",
  "first_name": "John",
  "last_name": "Doe"
}
```

### API Documentation
- **Swagger UI**: Available at `/docs` when running
- **OpenAPI Spec**: Available at `/docs/swagger.json`

## 🧪 Testing

### Test Structure
```
tests/
├── integration/         # Integration tests
├── e2e/                # End-to-end tests
└── fixtures/           # Test data
```

### Running Tests
```bash
# Unit tests
make test

# Integration tests
make test-integration

# Coverage report
make test-coverage

# Test specific package
go test -v ./internal/authentication/service/
```

## 🤝 Contributing

1. **Fork** the repository
2. **Create** a feature branch: `git checkout -b feature/amazing-feature`
3. **Commit** your changes: `git commit -m 'Add amazing feature'`
4. **Push** to the branch: `git push origin feature/amazing-feature`
5. **Open** a Pull Request

### Development Guidelines
- Follow Go coding standards
- Write comprehensive tests
- Update documentation
- Ensure pre-commit hooks pass
- Use conventional commit messages

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

### Documentation
- [Architecture Guide](docs/architecture/)
- [API Documentation](docs/api/)
- [Deployment Guide](docs/deployment/)
- [Development Guide](docs/development/)

### Issues
- [GitHub Issues](https://github.com/VincentArjuna/RexiErp/issues)
- [Discussions](https://github.com/VincentArjuna/RexiErp/discussions)

### Community
- [Discord Server](https://discord.gg/rexierp)
- [Telegram Group](https://t.me/rexierp)

## 🙏 Acknowledgments

- Indonesian MSME community for requirements and feedback
- Go open-source community
- Docker and Kubernetes ecosystems
- Indonesian tax authorities for API documentation

---

**Built with ❤️ for Indonesian MSMEs**