# RexiERP Architecture Document

This document outlines the overall project architecture for RexiERP, including backend systems, shared services, and non-UI specific concerns. Its primary goal is to serve as the guiding architectural blueprint for AI-driven development, ensuring consistency and adherence to chosen patterns and technologies.

**Relationship to Frontend Architecture:**
If the project includes a significant user interface, a separate Frontend Architecture Document will detail the frontend-specific design and MUST be used in conjunction with this document. Core technology stack choices documented herein (see "Tech Stack") are definitive for the entire project, including any frontend components.

## Starter Template or Existing Project

Based on my analysis of the PRD, this is a greenfield Go-based microservices project. The PRD clearly states:

- **Repository Structure:** Monorepo (explicitly decided in PRD)
- **Service Architecture:** Microservices with domain-driven design
- **Technology Stack:** Go programming language with Gin framework, GORM, PostgreSQL, Redis, and container-based deployment

**Recommendation:** No starter template will be used. This is a custom Go microservices project built from scratch to meet Indonesian MSME requirements, with specific technical constraints around zero-cost licensing and cloud-agnostic deployment.

**Rationale:**
- The project requires specific Indonesian compliance features (e-Faktur, BPJS, PPn 11%) that generic templates don't support
- Zero-cost licensing requirement eliminates most commercial templates
- Cloud-agnostic deployment needs custom infrastructure as code
- Monorepo structure with microservices warrants custom organization

## Change Log

| Date | Version | Description | Author |
|------|---------|-------------|---------|
| 2025-11-01 | v1.0 | Initial backend architecture creation | Winston (Architect) |

## High Level Architecture

### Technical Summary

RexiERP is a cloud-native microservices architecture built on Go that provides comprehensive business management for Indonesian MSMEs. The system uses domain-driven design with separate services for core business functions (inventory, accounting, HR, CRM), connected via REST APIs and event-driven messaging. PostgreSQL databases provide transactional integrity, Redis enables caching and session management, and containerized deployment ensures scalability across multiple cloud providers while maintaining zero software licensing costs.

### High Level Overview

**Architectural Style:** Microservices with Domain-Driven Design (DDD)

**Repository Structure:** Monorepo (as decided in PRD) with separate service directories

**Service Architecture:**
- Core business services: Inventory, Accounting, HR, CRM
- Cross-cutting services: Authentication, Notification, Integration
- API Gateway for unified access point

**Primary User Flow:**
1. User authenticates via API Gateway
2. Requests routed to appropriate microservice
3. Services communicate via REST APIs and async messaging (RabbitMQ)
4. Data persisted in PostgreSQL with Redis caching
5. External Indonesian APIs integrated for compliance

**Key Architectural Decisions:**
- **Microservices:** Enables independent scaling and deployment of business domains
- **Go + Gin:** Zero-cost licensing, excellent performance, Indonesian developer familiarity
- **PostgreSQL:** Open-source, ACID compliance for financial data, mature ecosystem
- **Redis:** Zero-cost caching, Indonesian cloud provider support
- **Monorepo:** Simplifies dependency management and deployment coordination

### High Level Project Diagram

```mermaid
graph TB
    subgraph "External Systems"
        EF[ e-Faktur API ]
        BPJS[ BPJS API ]
        EINV[ e-Invoice API ]
        TAX[ Tax API ]
    end

    subgraph "API Gateway Layer"
        GW[ API Gateway<br/>Gin + Rate Limiting ]
    end

    subgraph "Microservices Layer"
        AUTH[ Authentication Service ]
        INV[ Inventory Service ]
        ACC[ Accounting Service ]
        HR[ HR Service ]
        CRM[ CRM Service ]
        NOTIF[ Notification Service ]
        INTEG[ Integration Service ]
    end

    subgraph "Data Layer"
        PG[( PostgreSQL<br/>Multi-tenant )]
        REDIS[( Redis<br/>Cache/Sessions )]
    end

    subgraph "Infrastructure"
        K8S[ Kubernetes Cluster ]
        MON[ Monitoring Stack ]
    end

    Users --> GW
    GW --> AUTH
    GW --> INV
    GW --> ACC
    GW --> HR
    GW --> CRM

    AUTH --> PG
    AUTH --> REDIS
    INV --> PG
    INV --> REDIS
    ACC --> PG
    ACC --> REDIS
    HR --> PG
    HR --> REDIS
    CRM --> PG
    CRM --> REDIS

    INTEG --> EF
    INTEG --> BPJS
    INTEG --> EINV
    INTEG --> TAX

    NOTIF --> REDIS

    K8S --> GW
    K8S --> AUTH
    K8S --> INV
    K8S --> ACC
    K8S --> HR
    K8S --> CRM
    K8S --> NOTIF
    K8S --> INTEG
```

### Architectural and Design Patterns

**Recommended Patterns:**

- **Domain-Driven Design (DDD):** Bounded contexts for each business domain (Inventory, Accounting, HR, CRM) - _Rationale:_ Clean separation of concerns, aligns with microservices, handles Indonesian business complexity

- **Repository Pattern:** Abstract data access with GORM interfaces - _Rationale:_ Enables testing, database migration flexibility, cleaner service logic

- **Event-Driven Architecture:** RabbitMQ for async communication - _Rationale:_ Service decoupling, better scalability, Indonesian compliance event handling

- **API Gateway Pattern:** Single entry point with authentication/routing - _Rationale:_ Simplifies client access, centralized security, Indonesian compliance enforcement

- **CQRS (Command Query Responsibility Segregation):** Separate read/write models for complex queries - _Rationale:_ Performance optimization for reporting, audit trail requirements for Indonesian taxes

- **Circuit Breaker Pattern:** Fault tolerance for external Indonesian APIs - _Rationale:_ Handles unreliable government APIs, prevents cascading failures

- **Multi-tenant Pattern:** Schema-based tenant isolation - _Rationale:_ Cost efficiency for MSME market, data isolation compliance

## Tech Stack

**CRITICAL:** This section is the DEFINITIVE technology selection for RexiERP. All architectural decisions depend on these choices. Please review carefully as these selections will guide all development.

### Cloud Infrastructure

- **Provider:** Multi-cloud capable (AWS, GCP, Azure support via Kubernetes)
- **Key Services:** Container orchestration, managed PostgreSQL, managed Redis, object storage, CDN
- **Deployment Regions:** Indonesia (Jakarta primary), Singapore secondary for DR

### Technology Stack Table

| Category | Technology | Version | Purpose | Rationale |
|----------|------------|---------|---------|-----------|
| **Language** | Go | 1.23.1 | Primary development language | Zero-cost licensing, excellent performance, strong Indonesian developer community |
| **Web Framework** | Gin | 1.9.1 | HTTP web framework | Lightweight, fast, Indonesian developer familiarity, minimal memory footprint |
| **ORM** | GORM | 1.25.10 | Database ORM | Mature Go ORM, PostgreSQL support, migration tools, Indonesian community |
| **Database** | PostgreSQL | 16.6 | Primary relational database | ACID compliance, JSON support, mature, zero-cost, Indonesian cloud support |
| **Cache/Session** | Redis | 7.2.5 | Caching and session store | Zero-cost, Indonesian cloud support, local Docker support |
| **Message Queue** | RabbitMQ | 3.13.6 | Async messaging and events | Mature, reliable Docker support, Indonesian cloud compatibility, better features than Redis pub/sub |
| **API Gateway** | **Nginx** | 1.25.5 | **API gateway and load balancer** | **PRD requirement**, battle-tested, Indonesian cloud support, excellent performance |
| **Container Runtime** | Docker | 27.3.1 | Local development and deployment | Primary local development environment, Indonesian cloud compatibility |
| **Orchestration** | Docker Compose | 2.29.0 | Local multi-service orchestration | **Priority #1** - Local development environment before Kubernetes |
| **Orchestration** | Kubernetes | 1.31.0 | Production container orchestration | **Phase 2** - After local Docker Compose setup is working |
| **Authentication** | JWT | 1.2.1 | Token-based auth | Stateless, Indonesian mobile app compatibility |
| **Monitoring** | Prometheus + Grafana | 2.53.2 + 11.1.0 | Metrics and visualization | Open-source, Docker Compose support, Indonesian cloud support |
| **Logging** | Logrus + ELK | 1.9.3 + 8.15.0 | Structured logging | JSON logging, Docker Compose support, Indonesian timezone support |
| **Infrastructure as Code** | Terraform | 1.9.8 | Infrastructure provisioning | **Phase 2** - After Docker Compose local setup works |
| **CI/CD** | GitHub Actions | Latest | Pipeline automation | Free for public repos, Indonesian developer familiarity |
| **API Documentation** | Swagger/OpenAPI | 3.0 | API documentation | Auto-generated from Go code, Indonesian developer familiarity |

### Development Phases

**Phase 1: Local Docker Development (Priority)**
- Docker Compose for all services orchestration
- **Nginx as API Gateway** (reverse proxy, load balancing, SSL termination)
- PostgreSQL, Redis, RabbitMQ in Docker containers
- All microservices running locally
- Complete local development environment

**Phase 2: Production Deployment**
- Kubernetes orchestration for production
- **Nginx Ingress Controller** for production API gateway
- Terraform for cloud infrastructure
- Managed cloud services (PostgreSQL, Redis, RabbitMQ)
- CI/CD pipeline setup

**Updated Key Decisions:**
- **Nginx API Gateway:** Following PRD requirement instead of custom Gin gateway
- **Docker Compose First:** Local development environment priority over Kubernetes
- **RabbitMQ over Redis pub/sub:** Proper message queuing with reliability features
- **Phased Approach:** Local development fully functional before cloud deployment
- **Same Stack:** Consistent technology between local and production environments

## Data Models

### Users

**Purpose:** Core user management with role-based access control for multi-tenant Indonesian MSME operations

**Key Attributes:**
- id: UUID - Primary identifier
- tenant_id: UUID - Multi-tenant isolation
- email: string - User login and notifications (Indonesian email validation)
- password_hash: string - Secure password storage
- full_name: string - User's full name (Indonesian naming conventions)
- phone_number: string - Indonesian mobile number format (+62)
- role: enum - Role-based permissions (super_admin, tenant_admin, staff, viewer)
- is_active: boolean - Account status
- last_login: timestamp - Security tracking
- created_at: timestamp - Audit trail
- updated_at: timestamp - Audit trail

**Relationships:**
-belongsTo Tenant (tenant_id)
-hasMany UserSessions
-hasMany ActivityLogs

### Tenants

**Purpose:** Multi-tenant management for Indonesian MSME customers with company-specific data isolation

**Key Attributes:**
- id: UUID - Primary identifier
- company_name: string - Legal company name (Indonesian PT/CV/Firma formats)
- company_type: enum - Business entity type (PT, CV, Firma, Udaha)
- tax_id: string - NPWP (Indonesian Taxpayer Identification Number)
- business_license: string - SIUP/NIB (Indonesian Business License)
- address: text - Full Indonesian address format
- province: string - Indonesian province code
- city: string - Indonesian city code
- postal_code: string - Indonesian postal code
- phone: string - Company phone (Indonesian format)
- email: string - Company email
- is_active: boolean - Tenant status
- subscription_plan: enum - Subscription tier (basic, professional, enterprise)
- max_users: integer - User limits per plan
- created_at: timestamp - Account creation
- updated_at: timestamp - Last modifications

**Relationships:**
-hasMany Users
-hasMany Companies
-hasMany Warehouses
-hasMany Products
-hasMany Customers
-hasMany Suppliers
-hasMany Transactions

### Products

**Purpose:** Product catalog management with Indonesian tax compliance and inventory tracking

**Key Attributes:**
- id: UUID - Primary identifier
- tenant_id: UUID - Multi-tenant isolation
- sku: string - Stock Keeping Unit (unique per tenant)
- barcode: string - Product barcode (Indonesian format)
- name: string - Product name (Indonesian/English)
- description: text - Product description
- category_id: UUID - Product categorization
- unit: string - Measurement unit (pcs, kg, liter, etc.)
- purchase_price: decimal - Cost price (IDR)
- selling_price: decimal - Retail price (IDR)
- tax_rate: decimal - PPN rate (11% standard, other rates)
- is_taxable: boolean - Tax eligibility
- stock_quantity: integer - Current stock
- min_stock: integer - Reorder point
- max_stock: integer - Maximum stock
- weight: decimal - Shipping weight (kg)
- dimensions: json - Product dimensions (LxWxH cm)
- is_active: boolean - Product status
- created_at: timestamp - Product creation
- updated_at: timestamp - Last modification

**Relationships:**
-belongsTo Tenant
-belongsTo Category
-hasMany InventoryTransactions
-hasMany OrderItems
-hasMany PurchaseOrderItems

### Customers

**Purpose:** Indonesian customer management with tax compliance and billing information

**Key Attributes:**
- id: UUID - Primary identifier
- tenant_id: UUID - Multi-tenant isolation
- customer_code: string - Customer identifier (auto-generated)
- name: string - Customer name (Indonesian naming conventions)
- email: string - Customer email
- phone: string - Indonesian phone format
- address: text - Full Indonesian address
- province: string - Indonesian province code
- city: string - Indonesian city code
- postal_code: string - Indonesian postal code
- tax_id: string - NPWP (for tax invoices)
- customer_type: enum - Individual/Company/B2B
- credit_limit: decimal - Credit limit (IDR)
- is_active: boolean - Customer status
- created_at: timestamp - Customer creation
- updated_at: timestamp - Last modification

**Relationships:**
-belongsTo Tenant
-hasMany SalesOrders
-hasMany Invoices
-hasMany Payments

### Suppliers

**Purpose:** Indonesian supplier management with procurement and tax compliance

**Key Attributes:**
- id: UUID - Primary identifier
- tenant_id: UUID - Multi-tenant isolation
- supplier_code: string - Supplier identifier
- name: string - Supplier company name
- contact_person: string - Contact person
- email: string - Supplier email
- phone: string - Indonesian phone format
- address: text - Full Indonesian address
- tax_id: string - Supplier NPWP
- payment_terms: integer - Payment terms in days
- is_active: boolean - Supplier status
- created_at: timestamp - Supplier creation
- updated_at: timestamp - Last modification

**Relationships:**
-belongsTo Tenant
-hasMany PurchaseOrders
-hasMany Bills

### Transactions

**Purpose:** Financial transaction tracking with Indonesian tax compliance and audit trails

**Key Attributes:**
- id: UUID - Primary identifier
- tenant_id: UUID - Multi-tenant isolation
- transaction_number: string - Auto-generated transaction number
- transaction_type: enum - Sale/Purchase/Return/Adjustment
- reference_id: UUID - Related document ID (order/invoice/etc.)
- customer_id: UUID - Customer reference (nullable)
- supplier_id: UUID - Supplier reference (nullable)
- total_amount: decimal - Total amount (IDR)
- tax_amount: decimal - PPN tax amount
- discount_amount: decimal - Discount amount
- final_amount: decimal - Final amount paid
- payment_method: enum - Cash/Bank Transfer/E-wallet
- payment_status: enum - Pending/Paid/Overdue/Cancelled
- transaction_date: timestamp - Transaction date
- due_date: timestamp - Payment due date
- notes: text - Transaction notes
- created_at: timestamp - Transaction creation
- updated_at: timestamp - Last modification

**Relationships:**
-belongsTo Tenant
-belongsTo Customer
-belongsTo Supplier
-hasMany TransactionItems
-hasMany Payments

## Components

Based on the architectural patterns, tech stack, and data models, here are the major logical components:

### API Gateway

**Responsibility:** Single entry point for all client requests, routing, authentication, and load balancing

**Key Interfaces:**
- HTTP/HTTPS endpoints for external access
- Authentication middleware with JWT validation
- Rate limiting and request throttling
- Request routing to appropriate microservices
- SSL/TLS termination
- CORS handling for Indonesian cross-origin requirements

**Dependencies:** Nginx, Redis (for rate limiting), Authentication Service

**Technology Stack:** Nginx 1.25.5 with custom configuration, Lua scripting for complex routing rules

### Authentication Service

**Responsibility:** User authentication, authorization, and session management for multi-tenant Indonesian MSMEs

**Key Interfaces:**
- POST /auth/login - User authentication
- POST /auth/logout - Session termination
- POST /auth/refresh - Token refresh
- GET /auth/profile - User profile retrieval
- POST /auth/register - New user registration
- Internal APIs for service-to-service authentication

**Dependencies:** PostgreSQL, Redis, RabbitMQ

**Technology Stack:** Go 1.23.1, Gin 1.9.1, GORM 1.25.10, JWT 1.2.1, bcrypt for password hashing

### Inventory Service

**Responsibility:** Product catalog management, stock tracking, and inventory operations for Indonesian MSMEs

**Key Interfaces:**
- CRUD operations for Products, Categories, Warehouses
- Stock adjustment and tracking APIs
- Low stock alerts and reordering suggestions
- Inventory reporting and analytics
- Barcode scanning support
- Multi-warehouse management

**Dependencies:** PostgreSQL, Redis (caching), RabbitMQ (events)

**Technology Stack:** Go 1.23.1, Gin 1.9.1, GORM 1.25.10, Redis 7.2.5

### Accounting Service

**Responsibility:** Financial transaction management, Indonesian tax compliance, and accounting operations

**Key Interfaces:**
- Transaction CRUD operations
- Invoice generation with Indonesian tax compliance
- Payment processing and tracking
- Financial reporting (P&L, Balance Sheet)
- Tax calculation and reporting (PPN 11%)
- Account reconciliation
- BPJS calculations and deductions

**Dependencies:** PostgreSQL, Redis, RabbitMQ, External Indonesian APIs

**Technology Stack:** Go 1.23.1, Gin 1.9.1, GORM 1.25.10, custom Indonesian tax calculation library

### HR Service

**Responsibility:** Employee management, payroll processing, and BPJS social security integration

**Key Interfaces:**
- Employee data management
- Payroll calculation and processing
- BPJS contribution calculations
- Attendance and leave management
- Employee performance tracking
- Indonesian tax calculations for payroll

**Dependencies:** PostgreSQL, Redis, RabbitMQ, BPJS API

**Technology Stack:** Go 1.23.1, Gin 1.9.1, GORM 1.25.10, BPJS API client library

### CRM Service

**Responsibility:** Customer relationship management, sales tracking, and customer analytics

**Key Interfaces:**
- Customer data management
- Sales order processing
- Customer communication tracking
- Sales pipeline management
- Customer analytics and reporting
- Marketing campaign management

**Dependencies:** PostgreSQL, Redis, RabbitMQ

**Technology Stack:** Go 1.23.1, Gin 1.9.1, GORM 1.25.10

### Integration Service

**Responsibility:** External API integrations for Indonesian government compliance and third-party services

**Key Interfaces:**
- e-Faktur API integration for tax compliance
- BPJS API integration for social security
- Bank API integrations for payments
- e-Invoice API integration
- Third-party service APIs (GoTo, e-wallets)
- Data synchronization and retry logic

**Dependencies:** RabbitMQ, External Indonesian APIs, Redis

**Technology Stack:** Go 1.23.1, custom HTTP clients with circuit breaker pattern

### Notification Service

**Responsibility:** Multi-channel notifications for Indonesian MSME business operations

**Key Interfaces:**
- Email notifications (transaction confirmations, alerts)
- SMS notifications for Indonesian mobile numbers
- In-app notifications
- Push notifications for mobile apps
- Notification templates and localization (Indonesian/English)

**Dependencies:** Redis, RabbitMQ, External SMS/Email providers

**Technology Stack:** Go 1.23.1, email/SMS client libraries, template engine

### Component Diagrams

```mermaid
graph TB
    subgraph "External Systems"
        EF[ e-Faktur API ]
        BPJS[ BPJS API ]
        EINV[ e-Invoice API ]
        BANK[ Bank APIs ]
        SMS[ SMS Gateway ]
        EMAIL[ Email Gateway ]
    end

    subgraph "Client Layer"
        WEB[ Web App ]
        MOBILE[ Mobile App ]
    end

    subgraph "API Gateway Layer"
        NGINX[ Nginx Gateway<br/>Authentication<br/>Rate Limiting<br/>Load Balancing ]
    end

    subgraph "Microservices Layer"
        AUTH[ Authentication Service ]
        INV[ Inventory Service ]
        ACC[ Accounting Service ]
        HR[ HR Service ]
        CRM[ CRM Service ]
        NOTIF[ Notification Service ]
        INTEG[ Integration Service ]
    end

    subgraph "Communication Layer"
        RMQ[ RabbitMQ<br/>Event Bus ]
    end

    subgraph "Data Layer"
        PG[( PostgreSQL<br/>Multi-tenant Schema )]
        REDIS[( Redis<br/>Cache & Sessions )]
    end

    WEB --> NGINX
    MOBILE --> NGINX

    NGINX --> AUTH
    NGINX --> INV
    NGINX --> ACC
    NGINX --> HR
    NGINX --> CRM

    AUTH -.-> RMQ
    INV -.-> RMQ
    ACC -.-> RMQ
    HR -.-> RMQ
    CRM -.-> RMQ
    NOTIF -.-> RMQ
    INTEG -.-> RMQ

    RMQ --> AUTH
    RMQ --> INV
    RMQ --> ACC
    RMQ --> HR
    RMQ --> CRM
    RMQ --> NOTIF
    RMQ --> INTEG

    AUTH --> PG
    AUTH --> REDIS
    INV --> PG
    INV --> REDIS
    ACC --> PG
    ACC --> REDIS
    HR --> PG
    HR --> REDIS
    CRM --> PG
    CRM --> REDIS

    INTEG --> EF
    INTEG --> BPJS
    INTEG --> EINV
    INTEG --> BANK

    NOTIF --> SMS
    NOTIF --> EMAIL
```

## External APIs

### e-Faktur API

**Purpose:** Indonesian tax compliance for electronic tax invoices (e-Faktur) required by DJP (Direktorat Jenderal Pajak)

- **Documentation:** https://efaktur.pajak.go.id/
- **Base URL(s):** https://api.efaktur.pajak.go.id/v1/
- **Authentication:** OAuth 2.0 with client credentials + digital certificate
- **Rate Limits:** 100 requests per minute, 5000 requests per day

**Key Endpoints Used:**
- `POST /invoices` - Create e-Faktur invoice
- `GET /invoices/{id}` - Retrieve e-Faktur status
- `POST /invoices/{id}/cancel` - Cancel e-Faktur
- `GET /tax-rates` - Get current PPN tax rates
- `POST /validation` - Validate taxpayer data

**Integration Notes:**
- Requires digital certificate for authentication
- Indonesian timezone (WIB/UTC+7) for all timestamps
- Retry logic essential due to government API reliability issues
- 30-day retention period for invoice data

### BPJS API

**Purpose:** Indonesian social security program (BPJS Ketenagakerjaan & BPJS Kesehatan) integration for employee contributions

- **Documentation:** https://api.bpjsketenagakerjaan.go.id/ and https://api.bpjs-kesehatan.go.id/
- **Base URL(s):**
  - Employment: https://api.bpjsketenagakerjaan.go.id/v1/
  - Health: https://api.bpjs-kesehatan.go.id/v1/
- **Authentication:** Consumer Key + Consumer Secret + HMAC signature
- **Rate Limits:** 150 requests per minute per service

**Key Endpoints Used:**
- `POST /participant/verification` - Verify employee BPJS status
- `GET /contributions/{period}` - Get contribution rates
- `POST /contributions/calculate` - Calculate employee contributions
- `GET /participants/{nik}` - Get participant details
- `POST /payment/report` - Report contribution payments

**Integration Notes:**
- Separate APIs for employment and health BPJS
- Indonesian ID number (NIK) validation required
- Monthly contribution periods (1st-21st of each month)
- Complex calculation rules for different employee categories

### e-Invoice API

**Purpose:** Indonesian electronic invoice system for B2G (Business-to-Government) transactions

- **Documentation:** https://einvoice.pajak.go.id/
- **Base URL(s):** https://api.einvoice.pajak.go.id/v1/
- **Authentication:** Digital certificate + OAuth 2.0
- **Rate Limits:** 200 requests per minute

**Key Endpoints Used:**
- `POST /invoices/create` - Create electronic invoice
- `GET /invoices/status/{id}` - Check invoice status
- `POST /invoices/cancel` - Cancel electronic invoice
- `GET /partners` - Get government partner information

**Integration Notes:**
- Mandatory for B2G transactions above certain thresholds
- Requires pre-approval from tax authorities
- Integration with existing accounting system critical

### Bank API Integrations

**Purpose:** Payment processing and bank account verification for Indonesian banks

- **Documentation:** Varies by bank (BCA, Mandiri, BNI, BRI, etc.)
- **Base URL(s):** Bank-specific APIs
- **Authentication:** OAuth 2.0 + API keys
- **Rate Limits:** Varies by bank (typically 100-500 requests/minute)

**Key Endpoints Used:**
- `POST /payments/verify` - Verify payment status
- `GET /accounts/{number}/validate` - Validate bank accounts
- `POST /payments/disbursement` - Process payments
- `GET /transactions/history` - Get transaction history

**Integration Notes:**
- Multiple bank integrations required for Indonesian MSMEs
- BI (Bank Indonesia) compliance for payment processing
- Real-time verification vs batch processing options

### SMS Gateway API

**Purpose:** SMS notifications for Indonesian mobile numbers (transaction alerts, OTP, notifications)

- **Documentation:** Provider-specific (e.g., Telkom, Vonage, Twilio Indonesia)
- **Base URL(s):** Provider-specific
- **Authentication:** API key + HMAC signature
- **Rate Limits:** Provider-specific (typically 100 SMS/minute)

**Key Endpoints Used:**
- `POST /sms/send` - Send SMS notification
- `GET /sms/delivery/{id}` - Check delivery status
- `GET /sms/balance` - Check SMS balance

**Integration Notes:**
- Indonesian mobile number format validation (+62)
- Provider selection for cost optimization
- Delivery reliability considerations

## Core Workflows

### Customer Sales Transaction Flow

```mermaid
sequenceDiagram
    participant C as Customer
    participant G as Nginx Gateway
    participant A as Auth Service
    participant CRM as CRM Service
    participant INV as Inventory Service
    participant ACC as Accounting Service
    participant N as Notification Service
    participant DB as PostgreSQL
    participant R as Redis
    participant MQ as RabbitMQ

    C->>G: Login Request
    G->>A: Validate Credentials
    A->>DB: Verify User
    A-->>G: JWT Token
    G-->>C: Auth Success

    C->>G: Create Sales Order
    G->>A: Validate JWT
    A-->>G: Token Valid
    G->>CRM: Create Order Request

    CRM->>INV: Check Product Availability
    INV->>DB: Query Stock
    INV->>R: Check Cache
    R-->>INV: Cached Stock Data
    INV-->>CRM: Stock Available

    CRM->>DB: Create Sales Order
    CRM->>MQ: Publish Order Created Event

    MQ->>INV: Reserve Stock
    INV->>DB: Update Inventory
    MQ->>ACC: Create Invoice
    ACC->>DB: Generate Invoice with PPN 11%
    ACC->>DB: Apply Tax Rules
    MQ->>N: Send Order Confirmation

    N->>SMS: Send SMS to Customer
    N->>EMAIL: Send Email Invoice

    CRM-->>G: Order Created Response
    G-->>C: Order Confirmation
```

### Indonesian Tax Compliance Workflow

```mermaid
sequenceDiagram
    participant SYS as System
    participant ACC as Accounting Service
    participant INTEG as Integration Service
    participant EF as e-Faktur API
    participant EINV as e-Invoice API
    participant MQ as RabbitMQ
    participant N as Notification Service

    ACC->>DB: Get Invoice Ready for e-Faktur
    ACC->>ACC: Validate Indonesian Tax Rules
    ACC->>ACC: Calculate PPN 11%
    ACC->>INTEG: Submit to e-Faktur

    INTEG->>EF: POST /invoices
    EF-->>INTEG: e-Faktur Number
    INTEG->>DB: Save e-Faktur Reference

    par Government Processing
        EF->>EF: Validate Tax Data
        EF->>EF: Generate Tax Invoice
    end

    EF->>INTEG: e-Faktur Ready
    INTEG->>MQ: Publish e-Faktur Ready Event

    MQ->>N: Send Tax Invoice
    MQ->>ACC: Update Invoice Status

    ACC->>DB: Mark as Tax Compliant
    N->>EMAIL: Send e-Faktur to Customer
```

### BPJS Payroll Processing Workflow

```mermaid
sequenceDiagram
    participant HR as HR Service
    participant DB as PostgreSQL
    participant MQ as RabbitMQ
    participant INTEG as Integration Service
    participant BPJS as BPJS API
    participant ACC as Accounting Service

    HR->>DB: Calculate Payroll
    HR->>HR: Calculate BPJS Contributions
    HR->>HR: Apply Indonesian Tax Rules
    HR->>DB: Generate Payroll Records

    HR->>MQ: Publish Payroll Processing Event
    MQ->>INTEG: Process BPJS Payments

    INTEG->>BPJS: Verify Employee BPJS Status
    BPJS-->>INTEG: Employee Valid
    INTEG->>BPJS: Calculate Contributions
    BPJS-->>INTEG: Contribution Amounts

    INTEG->>ACC: Update Payroll with BPJS
    ACC->>DB: Save BPJS Deductions
    ACC->>DB: Generate Payroll Report

    INTEG->>BPJS: Submit Payment Report
    BPJS-->>INTEG: Payment Confirmation
    INTEG->>DB: Save BPJS Reference
```

### Multi-tenant User Registration Flow

```mermaid
sequenceDiagram
    participant USER as New User
    participant G as Nginx Gateway
    participant AUTH as Auth Service
    participant DB as PostgreSQL
    participant R as Redis
    participant MQ as RabbitMQ
    participant N as Notification Service

    USER->>G: Register Account
    G->>AUTH: Registration Request
    AUTH->>DB: Check Tenant Exists
    AUTH->>DB: Check Email Available

    DB-->>AUTH: Tenant Valid
    DB-->>AUTH: Email Available
    AUTH->>AUTH: Hash Password
    AUTH->>DB: Create User Account
    AUTH->>R: Create Session

    AUTH->>MQ: Publish User Registered Event
    MQ->>N: Send Welcome Email

    N->>EMAIL: Send Welcome Message
    AUTH->>R: Set User Permissions

    AUTH-->>G: Registration Success
    G-->>USER: Account Created
```

## REST API Spec

```yaml
openapi: 3.0.0
info:
  title: RexiERP Backend API
  version: v1.0.0
  description: Comprehensive ERP system for Indonesian MSMEs with tax compliance and government integrations
servers:
  - url: https://api.rexi-erp.id/v1
    description: Production server (Indonesia)
  - url: https://staging-api.rexi-erp.id/v1
    description: Staging server
  - url: http://localhost:8080/v1
    description: Local development server

paths:
  /auth/login:
    post:
      tags:
        - Authentication
      summary: User login
      description: Authenticate user and return JWT token
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - email
                - password
              properties:
                email:
                  type: string
                  format: email
                  example: "user@company.co.id"
                password:
                  type: string
                  format: password
                  example: "SecurePass123!"
      responses:
        '200':
          description: Login successful
          content:
            application/json:
              schema:
                type: object
                properties:
                  success:
                    type: boolean
                    example: true
                  data:
                    type: object
                    properties:
                      token:
                        type: string
                        example: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
                      user:
                        $ref: '#/components/schemas/User'
                      expires_in:
                        type: integer
                        example: 3600
        '401':
          description: Invalid credentials
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

  /products:
    get:
      tags:
        - Inventory
      summary: List products
      description: Get paginated list of products with filters
      parameters:
        - name: page
          in: query
          schema:
            type: integer
            default: 1
        - name: limit
          in: query
          schema:
            type: integer
            default: 20
        - name: category
          in: query
          schema:
            type: string
        - name: search
          in: query
          schema:
            type: string
      responses:
        '200':
          description: Products retrieved successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  success:
                    type: boolean
                    example: true
                  data:
                    type: object
                    properties:
                      products:
                        type: array
                        items:
                          $ref: '#/components/schemas/Product'
                      pagination:
                        $ref: '#/components/schemas/Pagination'

    post:
      tags:
        - Inventory
      summary: Create product
      description: Create new product with automatic SKU generation
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateProductRequest'
      responses:
        '201':
          description: Product created successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  success:
                    type: boolean
                    example: true
                  data:
                    $ref: '#/components/schemas/Product'

  /products/{id}:
    get:
      tags:
        - Inventory
      summary: Get product by ID
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: Product retrieved successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  success:
                    type: boolean
                    example: true
                  data:
                    $ref: '#/components/schemas/Product'
        '404':
          description: Product not found

  /invoices:
    post:
      tags:
        - Accounting
      summary: Create invoice
      description: Create sales invoice with Indonesian tax compliance
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateInvoiceRequest'
      responses:
        '201':
          description: Invoice created successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  success:
                    type: boolean
                    example: true
                  data:
                    $ref: '#/components/schemas/Invoice'

  /invoices/{id}/e-faktur:
    post:
      tags:
        - Accounting
      summary: Submit to e-Faktur
      description: Submit invoice to Indonesian e-Faktur system
      security:
        - bearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: e-Faktur submitted successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  success:
                    type: boolean
                    example: true
                  data:
                    type: object
                    properties:
                      efaktur_number:
                        type: string
                        example: "010.000-25.12345678"
                      status:
                        type: string
                        example: "submitted"

components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
          format: uuid
        email:
          type: string
          format: email
        full_name:
          type: string
        role:
          type: string
          enum: [super_admin, tenant_admin, staff, viewer]
        tenant_id:
          type: string
          format: uuid
        is_active:
          type: boolean
        created_at:
          type: string
          format: date-time

    Product:
      type: object
      properties:
        id:
          type: string
          format: uuid
        tenant_id:
          type: string
          format: uuid
        sku:
          type: string
        name:
          type: string
        description:
          type: string
        category_id:
          type: string
          format: uuid
        unit:
          type: string
        purchase_price:
          type: number
          format: decimal
        selling_price:
          type: number
          format: decimal
        tax_rate:
          type: number
          format: decimal
          example: 0.11
        is_taxable:
          type: boolean
        stock_quantity:
          type: integer
        min_stock:
          type: integer
        created_at:
          type: string
          format: date-time

    CreateProductRequest:
      type: object
      required:
        - name
        - category_id
        - unit
        - purchase_price
        - selling_price
      properties:
        name:
          type: string
        description:
          type: string
        category_id:
          type: string
          format: uuid
        unit:
          type: string
        purchase_price:
          type: number
          format: decimal
        selling_price:
          type: number
          format: decimal
        is_taxable:
          type: boolean
          default: true
        min_stock:
          type: integer
          default: 0

    Invoice:
      type: object
      properties:
        id:
          type: string
          format: uuid
        tenant_id:
          type: string
          format: uuid
        invoice_number:
          type: string
        customer_id:
          type: string
          format: uuid
        invoice_date:
          type: string
          format: date
        due_date:
          type: string
          format: date
        status:
          type: string
          enum: [draft, sent, paid, overdue, cancelled]
        subtotal:
          type: number
          format: decimal
        tax_amount:
          type: number
          format: decimal
        discount_amount:
          type: number
          format: decimal
        total_amount:
          type: number
          format: decimal
        paid_amount:
          type: number
          format: decimal
        efaktur_number:
          type: string
        efaktur_status:
          type: string
          enum: [pending, submitted, approved, rejected]
        efaktur_url:
          type: string
        notes:
          type: string
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time

    CreateInvoiceRequest:
      type: object
      required:
        - customer_id
        - items
      properties:
        customer_id:
          type: string
          format: uuid
        items:
          type: array
          items:
            type: object
            required:
              - product_id
              - quantity
              - unit_price
            properties:
              product_id:
                type: string
                format: uuid
              quantity:
                type: integer
                minimum: 1
              unit_price:
                type: number
                format: decimal
                minimum: 0

    Pagination:
      type: object
      properties:
        current_page:
          type: integer
        total_pages:
          type: integer
        total_items:
          type: integer
        items_per_page:
          type: integer

    Error:
      type: object
      properties:
        success:
          type: boolean
          example: false
        error:
          type: object
          properties:
            code:
              type: string
            message:
              type: string
            details:
              type: object

  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

security:
  - bearerAuth: []
```

## Database Schema

### Multi-tenant PostgreSQL Schema Design

**Architecture Approach:** Schema-based multi-tenancy for data isolation while maintaining cost efficiency for Indonesian MSME market

### Master Database Schema

```sql
-- Master database for tenant management
CREATE DATABASE rexi_erp_master;

-- Tenants table for multi-tenant management
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_name VARCHAR(255) NOT NULL,
    company_type VARCHAR(50) NOT NULL CHECK (company_type IN ('PT', 'CV', 'Firma', 'Udaha')),
    tax_id VARCHAR(25) UNIQUE, -- NPWP format: XX.XXX.XXX.X-XXX.XXX
    business_license VARCHAR(50), -- SIUP/NIB format
    address TEXT,
    province VARCHAR(100),
    city VARCHAR(100),
    postal_code VARCHAR(10),
    phone VARCHAR(20),
    email VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    subscription_plan VARCHAR(20) DEFAULT 'basic' CHECK (subscription_plan IN ('basic', 'professional', 'enterprise')),
    max_users INTEGER DEFAULT 5,
    database_name VARCHAR(100) UNIQUE NOT NULL,
    schema_name VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX idx_tenants_tax_id ON tenants(tax_id);
CREATE INDEX idx_tenants_is_active ON tenants(is_active);
CREATE INDEX idx_tenants_subscription_plan ON tenants(subscription_plan);
```

### Tenant-Specific Schema Template

```sql
-- Template for each tenant's database schema
-- Each tenant gets their own schema within a shared database

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    phone_number VARCHAR(20), -- Indonesian format: +62xxx
    role VARCHAR(20) NOT NULL CHECK (role IN ('super_admin', 'tenant_admin', 'staff', 'viewer')),
    is_active BOOLEAN DEFAULT true,
    last_login TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Product categories
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    parent_id UUID REFERENCES categories(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Products with Indonesian tax compliance
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    sku VARCHAR(100) UNIQUE NOT NULL,
    barcode VARCHAR(100),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category_id UUID REFERENCES categories(id),
    unit VARCHAR(50) NOT NULL,
    purchase_price DECIMAL(15,2) NOT NULL,
    selling_price DECIMAL(15,2) NOT NULL,
    tax_rate DECIMAL(5,4) DEFAULT 0.1100, -- 11% PPN standard rate
    is_taxable BOOLEAN DEFAULT true,
    stock_quantity INTEGER DEFAULT 0,
    min_stock INTEGER DEFAULT 0,
    max_stock INTEGER,
    weight DECIMAL(8,3), -- kg
    dimensions JSONB, -- {"length": 10.5, "width": 5.2, "height": 3.1}
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Customers with Indonesian tax info
CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    customer_code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(20),
    address TEXT,
    province VARCHAR(100),
    city VARCHAR(100),
    postal_code VARCHAR(10),
    tax_id VARCHAR(25), -- Customer NPWP
    customer_type VARCHAR(20) DEFAULT 'individual' CHECK (customer_type IN ('individual', 'company', 'b2b')),
    credit_limit DECIMAL(15,2) DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Suppliers for procurement
CREATE TABLE suppliers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    supplier_code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    contact_person VARCHAR(255),
    email VARCHAR(255),
    phone VARCHAR(20),
    address TEXT,
    tax_id VARCHAR(25), -- Supplier NPWP
    payment_terms INTEGER DEFAULT 30, -- days
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Sales orders
CREATE TABLE sales_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    order_number VARCHAR(100) UNIQUE NOT NULL,
    customer_id UUID NOT NULL REFERENCES customers(id),
    order_date DATE NOT NULL,
    due_date DATE,
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'confirmed', 'shipped', 'delivered', 'cancelled')),
    subtotal DECIMAL(15,2) NOT NULL,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    discount_amount DECIMAL(15,2) DEFAULT 0,
    total_amount DECIMAL(15,2) NOT NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Sales order items
CREATE TABLE sales_order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sales_order_id UUID NOT NULL REFERENCES sales_orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(15,2) NOT NULL,
    discount_percent DECIMAL(5,2) DEFAULT 0,
    tax_rate DECIMAL(5,4) DEFAULT 0.1100,
    line_total DECIMAL(15,2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Invoices with Indonesian compliance
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    invoice_number VARCHAR(100) UNIQUE NOT NULL,
    sales_order_id UUID REFERENCES sales_orders(id),
    customer_id UUID NOT NULL REFERENCES customers(id),
    invoice_date DATE NOT NULL,
    due_date DATE,
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'sent', 'paid', 'overdue', 'cancelled')),
    subtotal DECIMAL(15,2) NOT NULL,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    discount_amount DECIMAL(15,2) DEFAULT 0,
    total_amount DECIMAL(15,2) NOT NULL,
    paid_amount DECIMAL(15,2) DEFAULT 0,
    efaktur_number VARCHAR(50), -- e-Faktur reference
    efaktur_status VARCHAR(20) CHECK (efaktur_status IN ('pending', 'submitted', 'approved', 'rejected')),
    efaktur_url VARCHAR(500),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Inventory transactions for audit trail
CREATE TABLE inventory_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    product_id UUID NOT NULL REFERENCES products(id),
    transaction_type VARCHAR(20) NOT NULL CHECK (transaction_type IN ('purchase', 'sale', 'adjustment', 'return', 'transfer')),
    quantity_change INTEGER NOT NULL,
    quantity_before INTEGER NOT NULL,
    quantity_after INTEGER NOT NULL,
    reference_type VARCHAR(20), -- 'sales_order', 'purchase_order', etc.
    reference_id UUID,
    notes TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX idx_users_tenant_email ON users(tenant_id, email);
CREATE INDEX idx_products_tenant_sku ON products(tenant_id, sku);
CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_customers_tenant_code ON customers(tenant_id, customer_code);
CREATE INDEX idx_suppliers_tenant_code ON suppliers(tenant_id, supplier_code);
CREATE INDEX idx_sales_orders_tenant ON sales_orders(tenant_id);
CREATE INDEX idx_sales_orders_customer ON sales_orders(customer_id);
CREATE INDEX idx_sales_orders_date ON sales_orders(order_date);
CREATE INDEX idx_invoices_tenant ON invoices(tenant_id);
CREATE INDEX idx_invoices_customer ON invoices(customer_id);
CREATE INDEX idx_invoices_efaktur ON invoices(efaktur_number);
CREATE INDEX idx_inventory_transactions_product ON inventory_transactions(product_id);
CREATE INDEX idx_inventory_transactions_date ON inventory_transactions(created_at);

-- Row Level Security (RLS) for tenant isolation
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE products ENABLE ROW LEVEL SECURITY;
ALTER TABLE customers ENABLE ROW LEVEL SECURITY;
ALTER TABLE suppliers ENABLE ROW LEVEL SECURITY;
ALTER TABLE sales_orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE invoices ENABLE ROW LEVEL SECURITY;

-- RLS policies to ensure tenant isolation
CREATE POLICY tenant_isolation_users ON users FOR ALL TO application_role USING (tenant_id = current_setting('app.current_tenant_id')::UUID);
CREATE POLICY tenant_isolation_products ON products FOR ALL TO application_role USING (tenant_id = current_setting('app.current_tenant_id')::UUID);
CREATE POLICY tenant_isolation_customers ON customers FOR ALL TO application_role USING (tenant_id = current_setting('app.current_tenant_id')::UUID);
CREATE POLICY tenant_isolation_suppliers ON suppliers FOR ALL TO application_role USING (tenant_id = current_setting('app.current_tenant_id')::UUID);
CREATE POLICY tenant_isolation_sales_orders ON sales_orders FOR ALL TO application_role USING (tenant_id = current_setting('app.current_tenant_id')::UUID);
CREATE POLICY tenant_isolation_invoices ON invoices FOR ALL TO application_role USING (tenant_id = current_setting('app.current_tenant_id')::UUID);
```

## Source Tree

### Monorepo Structure for RexiERP Go Microservices

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

## Infrastructure and Deployment

### Infrastructure as Code

- **Tool:** Docker Compose 2.29.0 (Phase 1), Terraform 1.9.8 (Phase 2)
- **Location:** `deployments/docker-compose/` (Phase 1), `infrastructure/terraform/` (Phase 2)
- **Approach:** Local-first development with containerized services, progressing to cloud infrastructure

### Deployment Strategy

- **Strategy:** Blue-Green Deployment with Canary releases
- **CI/CD Platform:** GitHub Actions
- **Pipeline Configuration:** `.github/workflows/`

### Environments

- **Local Development:** Docker Compose with hot reload
- **Staging:** Production-like environment for testing Indonesian integrations
- **Production:** Multi-region deployment (Indonesia primary, Singapore DR)

### Environment Promotion Flow

```text
Local Development
       ↓
    Unit Tests
       ↓
Integration Tests
       ↓
Docker Compose Deploy
       ↓
Staging Environment
       ↓
Indonesian API Integration Tests
       ↓
Production Deploy (Blue-Green)
       ↓
Monitoring & Rollback if needed
```

### Local Docker Compose Configuration (Phase 1)

```yaml
# deployments/docker-compose/docker-compose.yml
version: '3.8'

services:
  # API Gateway
  nginx:
    image: nginx:1.25.5-alpine
    ports:
      - "8080:80"
      - "8443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/ssl:/etc/nginx/ssl
    depends_on:
      - authentication-service
      - inventory-service
      - accounting-service
    networks:
      - rexi-network

  # Core Services
  authentication-service:
    build:
      context: ../../
      dockerfile: cmd/authentication-service/Dockerfile
    environment:
      - DATABASE_URL=postgresql://rexi:password@postgres:5432/rexi_erp
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=${JWT_SECRET}
      - APP_ENV=development
    depends_on:
      - postgres
      - redis
    networks:
      - rexi-network

  inventory-service:
    build:
      context: ../../
      dockerfile: cmd/inventory-service/Dockerfile
    environment:
      - DATABASE_URL=postgresql://rexi:password@postgres:5432/rexi_erp
      - REDIS_URL=redis://redis:6379
      - RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672
      - APP_ENV=development
    depends_on:
      - postgres
      - redis
      - rabbitmq
    networks:
      - rexi-network

  accounting-service:
    build:
      context: ../../
      dockerfile: cmd/accounting-service/Dockerfile
    environment:
      - DATABASE_URL=postgresql://rexi:password@postgres:5432/rexi_erp
      - REDIS_URL=redis://redis:6379
      - RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672
      - EFAKTUR_API_URL=${EFAKTUR_API_URL}
      - APP_ENV=development
    depends_on:
      - postgres
      - redis
      - rabbitmq
    networks:
      - rexi-network

  hr-service:
    build:
      context: ../../
      dockerfile: cmd/hr-service/Dockerfile
    environment:
      - DATABASE_URL=postgresql://rexi:password@postgres:5432/rexi_erp
      - REDIS_URL=redis://redis:6379
      - RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672
      - BPJS_API_URL=${BPJS_API_URL}
      - APP_ENV=development
    depends_on:
      - postgres
      - redis
      - rabbitmq
    networks:
      - rexi-network

  crm-service:
    build:
      context: ../../
      dockerfile: cmd/crm-service/Dockerfile
    environment:
      - DATABASE_URL=postgresql://rexi:password@postgres:5432/rexi_erp
      - REDIS_URL=redis://redis:6379
      - RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672
      - APP_ENV=development
    depends_on:
      - postgres
      - redis
      - rabbitmq
    networks:
      - rexi-network

  notification-service:
    build:
      context: ../../
      dockerfile: cmd/notification-service/Dockerfile
    environment:
      - REDIS_URL=redis://redis:6379
      - RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672
      - SMS_GATEWAY_URL=${SMS_GATEWAY_URL}
      - EMAIL_GATEWAY_URL=${EMAIL_GATEWAY_URL}
      - APP_ENV=development
    depends_on:
      - redis
      - rabbitmq
    networks:
      - rexi-network

  integration-service:
    build:
      context: ../../
      dockerfile: cmd/integration-service/Dockerfile
    environment:
      - RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672
      - EFAKTUR_API_URL=${EFAKTUR_API_URL}
      - BPJS_API_URL=${BPJS_API_URL}
      - EINVOICE_API_URL=${EINVOICE_API_URL}
      - APP_ENV=development
    depends_on:
      - rabbitmq
    networks:
      - rexi-network

  # Data Services
  postgres:
    image: postgres:16.6-alpine
    environment:
      - POSTGRES_DB=rexi_erp
      - POSTGRES_USER=rexi
      - POSTGRES_PASSWORD=password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ../../migrations:/docker-entrypoint-initdb.d
    networks:
      - rexi-network

  redis:
    image: redis:7.2.5-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - rexi-network

  rabbitmq:
    image: rabbitmq:3.13.6-management-alpine
    environment:
      - RABBITMQ_DEFAULT_USER=guest
      - RABBITMQ_DEFAULT_PASS=guest
    ports:
      - "5672:5672"
      - "15672:15672"
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq
    networks:
      - rexi-network

  # Monitoring (Development)
  prometheus:
    image: prom/prometheus:v2.53.2
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
    networks:
      - rexi-network

  grafana:
    image: grafana/grafana:11.1.0
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_data:/var/lib/grafana
    networks:
      - rexi-network

volumes:
  postgres_data:
  redis_data:
  rabbitmq_data:
  grafana_data:

networks:
  rexi-network:
    driver: bridge
```

### Rollback Strategy

- **Primary Method:** Blue-Green deployment with instant traffic switching
- **Trigger Conditions:** Error rate >5%, response time >2s, Indonesian API failures >10%
- **Recovery Time Objective:** 5 minutes for full rollback

## Error Handling Strategy

### General Approach

- **Error Model:** Structured error response with consistent format across all microservices
- **Exception Hierarchy:** Custom error types wrapped in standard HTTP response format
- **Error Propagation:** Errors propagated through RabbitMQ events with correlation IDs

### Logging Standards

- **Library:** Logrus 1.9.3 with structured JSON logging
- **Format:** JSON with correlation ID, service name, timestamp, and error context
- **Levels:** ERROR, WARN, INFO, DEBUG (DEBUG disabled in production)
- **Required Context:**
  - Correlation ID: UUID format for request tracing
  - Service Context: Service name and version
  - User Context: Tenant ID and User ID (when available)

### Error Handling Patterns

#### External API Errors

- **Retry Policy:** Exponential backoff with jitter (100ms base, max 30s, 3 retries)
- **Circuit Breaker:** Opens after 5 consecutive failures, closes after 60s
- **Timeout Configuration:** 30s for Indonesian APIs, 10s for internal services
- **Error Translation:** Map Indonesian API errors to standardized internal error codes

#### Business Logic Errors

- **Custom Exceptions:** ValidationError, BusinessRuleError, ComplianceError
- **User-Facing Errors:** Indonesian and English error messages
- **Error Codes:** Standardized error codes for mobile app handling

#### Data Consistency

- **Transaction Strategy:** Database transactions with rollback on errors
- **Compensation Logic:** SAGA pattern for distributed transactions
- **Idempotency:** Idempotent operation keys for retry safety

## Coding Standards

**IMPORTANT:** These standards directly control AI developer behavior. Only include critical rules needed to prevent bad code.

### Core Standards

- **Languages & Runtimes:** Go 1.23.1 with Docker containerization
- **Style & Linting:** golangci-lint with custom configuration
- **Test Organization:** Unit tests in `*_test.go` files, integration tests in `tests/integration/`

### Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Package | lowercase, short, descriptive | `inventory`, `accounting` |
| Interface | -er suffix | `ProductRepository`, `InvoiceService` |
| Struct | PascalCase, descriptive | `Product`, `SalesOrder` |
| Function | PascalCase for exported, camelCase for unexported | `CreateProduct()`, `calculateTax()` |
| Variable | camelCase, meaningful | `productService`, `taxRate` |
| Constant | UPPER_SNAKE_CASE | `MAX_RETRY_ATTEMPTS`, `DEFAULT_TAX_RATE` |

### Critical Rules

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

## Test Strategy and Standards

### Testing Philosophy

- **Approach:** Test-driven development with 80% unit coverage, integration tests for critical paths
- **Coverage Goals:** 80% unit test coverage, 100% for Indonesian tax calculations and financial operations
- **Test Pyramid:** 70% unit tests, 20% integration tests, 10% end-to-end tests

### Test Types and Organization

#### Unit Tests

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

#### Integration Tests

- **Scope:** Database operations, Redis caching, RabbitMQ messaging, external API integrations
- **Location:** `tests/integration/`
- **Test Infrastructure:**
  - **PostgreSQL:** Testcontainers PostgreSQL for integration tests
  - **Redis:** Testcontainers Redis for caching tests
  - **RabbitMQ:** Testcontainers RabbitMQ for messaging tests
  - **External APIs:** WireMock for stubbing Indonesian government APIs

#### End-to-End Tests

- **Framework:** Go's testing package with HTTP client simulation
- **Scope:** Critical business workflows (sales orders with e-Faktur, payroll with BPJS)
- **Environment:** Staging environment with real Indonesian API integration
- **Test Data:** Seed test data with Indonesian business scenarios

### Test Data Management

- **Strategy:** Factory pattern with Go structs for test data generation
- **Fixtures:** `tests/fixtures/` directory with Indonesian business scenarios
- **Factories:** Data factory functions for realistic Indonesian business data
- **Cleanup:** Automatic cleanup after each test using testify/suite

### Continuous Testing

- **CI Integration:** GitHub Actions with parallel test execution
- **Performance Tests:** Go benchmarks for critical paths, k6 for load testing
- **Security Tests:** Gosec for static analysis, OWASP ZAP for dynamic analysis

## Security

### Input Validation

- **Validation Library:** Go's validator package with custom Indonesian validators
- **Validation Location:** API gateway layer and service boundaries
- **Required Rules:**
  - All external inputs MUST be validated
  - Validation at API boundary before processing
  - Whitelist approach preferred over blacklist
  - Indonesian phone number validation (+62 format)
  - NPWP (tax ID) format validation
  - Indonesian email domain validation

### Authentication & Authorization

- **Auth Method:** JWT tokens with refresh token rotation
- **Session Management:** Redis-based session store with configurable TTL
- **Required Patterns:**
  - JWT signed with RS256 keys
  - Role-based access control (RBAC) with tenant context
  - API key authentication for Indonesian government integrations
  - Multi-factor authentication for admin users

### Secrets Management

- **Development:** Environment variables with .env file (gitignored)
- **Production:** AWS Secrets Manager / Azure Key Vault / Google Secret Manager
- **Code Requirements:**
  - NEVER hardcode secrets
  - Access via configuration service only
  - No secrets in logs or error messages
  - Automatic secret rotation support

### API Security

- **Rate Limiting:** Redis-based rate limiting with configurable rules per endpoint
- **CORS Policy:** Restrictive CORS with Indonesian domain whitelist
- **Security Headers:** Strict-Transport-Security, X-Content-Type-Options, X-Frame-Options
- **HTTPS Enforcement:** Mandatory HTTPS in production, HTTP/2 support

### Data Protection

- **Encryption at Rest:** PostgreSQL Transparent Data Encryption (TDE) or application-level encryption
- **Encryption in Transit:** TLS 1.3 for all communications
- **PII Handling:** Encrypted storage for sensitive Indonesian personal data
- **Logging Restrictions:** No PII or financial data in application logs

### Dependency Security

- **Scanning Tool:** Go's built-in vulnerability scanner + Snyk
- **Update Policy:** Weekly dependency scans, immediate patching for critical vulnerabilities
- **Approval Process:** Security review for new external dependencies

### Security Testing

- **SAST Tool:** Gosec for static analysis
- **DAST Tool:** OWASP ZAP for dynamic security testing
- **Penetration Testing:** Quarterly penetration testing with focus on Indonesian compliance

## Checklist Results Report

**Status:** Architecture document creation completed. Ready for execution of architect checklist to validate architectural decisions against Indonesian MSME requirements and technical constraints.

## Next Steps

### Immediate Actions

1. **Review Architecture with Product Owner:**
   - Validate Indonesian compliance requirements
   - Confirm technology stack alignment with business goals
   - Review cost implications of chosen architecture

2. **Begin Story Implementation with Dev Agent:**
   - Set up local Docker Compose environment
   - Implement Authentication Service first (critical for all other services)
   - Create core data models and migrations
   - Implement basic CRUD operations for Products and Customers

3. **Set Up Infrastructure with DevOps Agent:**
   - Configure GitHub Actions CI/CD pipeline
   - Set up local development environment
   - Prepare staging environment for Indonesian API testing

### Development Sequence

**Phase 1: Foundation (Week 1-2)**
- Docker Compose local environment
- Authentication Service with JWT
- Basic multi-tenant database setup
- Nginx API Gateway configuration

**Phase 2: Core Services (Week 3-6)**
- Inventory Service (Products, Categories, Stock)
- Accounting Service (Invoices, Basic Tax)
- CRM Service (Customers, Basic Sales Orders)

**Phase 3: Advanced Features (Week 7-10)**
- Indonesian API Integrations (e-Faktur, BPJS)
- HR Service with Payroll
- Advanced Accounting Features
- Notification Service

**Phase 4: Production Ready (Week 11-12)**
- Comprehensive Testing
- Security Hardening
- Performance Optimization
- Production Deployment

### Architect Prompt for Frontend Architecture

*If this project includes significant UI components, provide this prompt to Architect for Frontend Architecture creation:*

"Create detailed frontend architecture for RexiERP Indonesian MSME system based on the completed backend architecture document. Key requirements:

- Mobile-first responsive design for Indonesian MSME users
- Progressive Web App (PWA) capabilities
- Indonesian language support with proper localization
- Offline functionality for areas with poor internet connectivity
- Integration with microservices backend through documented REST APIs
- Authentication integration with JWT tokens
- Real-time notifications for business operations
- Compliance with Indonesian accessibility standards

Technology stack should align with backend choices and prioritize performance on Indonesian mobile devices. Consider Progressive Web App approach for better reach in Indonesian MSME market."

---

## Architecture Document Completion

**CRITICAL WORKFLOW COMPLETED:** This comprehensive backend architecture document provides the complete blueprint for RexiERP development. All major architectural decisions have been documented with Indonesian compliance requirements, multi-tenancy considerations, and phased deployment approach prioritizing local Docker development before cloud deployment.

**Key Architectural Achievements:**
- ✅ Microservices architecture designed for Indonesian MSME scalability
- ✅ Multi-tenant PostgreSQL schema with Row Level Security
- ✅ Comprehensive Indonesian government API integration strategy
- ✅ Zero-cost technology stack aligned with Indonesian market constraints
- ✅ Phased deployment approach prioritizing local development
- ✅ Security architecture designed for Indonesian data protection compliance
- ✅ Comprehensive testing strategy for financial accuracy

**Ready for Development:** The architecture is now ready for story implementation and infrastructure setup.