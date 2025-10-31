# Source Tree

## Monorepo Structure for RexiERP Go Microservices

```
rexi-erp/
├── cmd/                           # Application entry points
│   ├── api-gateway/              # Nginx configuration
│   │   ├── nginx.conf
│   │   ├── Dockerfile
│   │   └── docker-compose.yml
│   ├── authentication-service/
│   │   └── main.go
│   ├── inventory-service/
│   │   └── main.go
│   ├── accounting-service/
│   │   └── main.go
│   ├── hr-service/
│   │   └── main.go
│   ├── crm-service/
│   │   └── main.go
│   ├── notification-service/
│   │   └── main.go
│   └── integration-service/
│       └── main.go
├── internal/                     # Private application code
│   ├── authentication/
│   │   ├── handler/             # HTTP handlers
│   │   ├── service/             # Business logic
│   │   ├── repository/          # Data access
│   │   ├── model/               # Data models
│   │   └── config/              # Configuration
│   ├── inventory/
│   │   ├── handler/
│   │   ├── service/
│   │   ├── repository/
│   │   ├── model/
│   │   └── config/
│   ├── accounting/
│   │   ├── handler/
│   │   ├── service/
│   │   ├── repository/
│   │   ├── model/
│   │   ├── tax/                 # Indonesian tax calculations
│   │   └── config/
│   ├── hr/
│   │   ├── handler/
│   │   ├── service/
│   │   ├── repository/
│   │   ├── model/
│   │   ├── payroll/             # BPJS calculations
│   │   └── config/
│   ├── crm/
│   │   ├── handler/
│   │   ├── service/
│   │   ├── repository/
│   │   ├── model/
│   │   └── config/
│   ├── notification/
│   │   ├── handler/
│   │   ├── service/
│   │   ├── repository/
│   │   ├── model/
│   │   └── channels/            # SMS/Email/WhatsApp
│   ├── integration/
│   │   ├── handler/
│   │   ├── service/
│   │   ├── repository/
│   │   ├── indonesia/           # Indonesian government APIs
│   │   │   ├── efaktur/
│   │   │   ├── bpjs/
│   │   │   ├── einvoice/
│   │   │   └── banks/
│   │   └── config/
│   └── shared/                  # Shared utilities
│       ├── middleware/          # Authentication, logging, rate limiting
│       ├── database/            # Database connections, migrations
│       ├── cache/               # Redis operations
│       ├── messaging/           # RabbitMQ operations
│       ├── validation/          # Input validation
│       ├── errors/              # Error handling
│       ├── logger/              # Structured logging
│       ├── auth/                # JWT handling
│       ├── tenant/              # Multi-tenant utilities
│       └── utils/               # General utilities
├── pkg/                         # Public library code
│   ├── api/                     # API contracts
│   │   ├── authentication/
│   │   ├── inventory/
│   │   ├── accounting/
│   │   ├── hr/
│   │   ├── crm/
│   │   ├── notification/
│   │   └── integration/
│   ├── database/                # Database interfaces
│   ├── cache/                   # Cache interfaces
│   ├── messaging/               # Messaging interfaces
│   ├── logger/                  # Logger interfaces
│   └── config/                  # Configuration management
├── migrations/                  # Database migrations
│   ├── master/                  # Master database migrations
│   └── tenants/                 # Tenant schema migrations
├── configs/                     # Configuration files
│   ├── local/
│   ├── staging/
│   └── production/
├── deployments/                 # Deployment configurations
│   ├── docker/                  # Docker files
│   ├── docker-compose/          # Local development
│   │   ├── docker-compose.yml
│   │   ├── docker-compose.override.yml
│   │   └── .env.example
│   └── kubernetes/              # K8s manifests (Phase 2)
│       ├── namespaces/
│       ├── services/
│       ├── deployments/
│       ├── ingress/
│       └── configmaps/
├── infrastructure/              # Infrastructure as Code (Phase 2)
│   ├── terraform/
│   │   ├── modules/
│   │   ├── environments/
│   │   └── scripts/
│   └── ansible/
├── scripts/                     # Build and utility scripts
│   ├── build.sh
│   ├── test.sh
│   ├── migrate.sh
│   ├── deploy.sh
│   └── setup-dev.sh
├── tests/                       # Test files
│   ├── integration/
│   ├── e2e/
│   └── fixtures/
├── docs/                        # Documentation
│   ├── api/                     # API documentation
│   ├── deployment/              # Deployment guides
│   └── architecture/
├── tools/                       # Development tools
│   ├── linter-configs/
│   └── generator/
├── .github/                     # GitHub workflows
│   └── workflows/
│       ├── ci.yml
│       ├── cd.yml
│       └── security-scan.yml
├── go.mod                       # Go module file
├── go.sum                       # Go dependencies
├── Makefile                     # Build commands
├── README.md                    # Project documentation
├── .gitignore
├── .env.example                 # Environment variables template
└── LICENSE
```
