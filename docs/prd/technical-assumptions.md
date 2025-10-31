# Technical Assumptions

## Repository Structure: Monorepo

**Decision:** Monorepo structure containing all microservices, shared libraries, and deployment configurations. This approach simplifies dependency management, ensures consistent versioning across services, and enables atomic commits across service boundaries.

**Rationale:** Chose monorepo over polyrepo to simplify development workflows for a small team, enable shared type definitions and utilities, and make it easier to maintain consistency across all microservices.

## Service Architecture: Microservices Architecture

**Decision:** Microservices architecture with domain-driven design principles. Each business domain (accounting, inventory, HR/payroll, tax, reporting) will be implemented as independent services communicating via REST APIs and event-driven messaging.

**Critical Decision:** This architecture enables independent scaling, deployment flexibility, and technology diversity across services while maintaining clear domain boundaries.

**Rationale:** Microservices approach supports the cloud-agnostic deployment requirement and allows for gradual migration/integration of existing systems. It also aligns with the API-first philosophy mentioned in your Project Brief.

## Testing Requirements: Full Testing Pyramid

**Critical Decision:** Implement comprehensive testing strategy including unit tests, integration tests, end-to-end API tests, and manual testing convenience methods.

**Rationale:** For financial systems handling sensitive data and regulatory compliance, comprehensive testing is non-negotiable. The complexity of Indonesian tax calculations and BPJS integrations demands thorough verification.

## Additional Technical Assumptions and Requests

**Programming Language & Framework:**
- **Go** for all microservices to achieve zero licensing costs, high performance, and deployment simplicity
- **Gin framework** for REST API development with Express.js-inspired API and optimal performance
- **GORM** for database ORM with support for PostgreSQL, MySQL, and MongoDB

**Database & Storage:**
- **PostgreSQL** as primary database for financial data (ACID compliance)
- **Redis** for caching and session management
- **MinIO/S3 compatible** for file storage and document management

**API Documentation & Standards:**
- **OpenAPI 3.0** specification for all REST endpoints
- **Swagger UI** for interactive API documentation
- **Postman collections** for API testing and client integration
- **API versioning** using URL path versioning (/api/v1/)

**Authentication & Security:**
- **JWT tokens** for service-to-service authentication
- **OAuth 2.0** for external integrations (GoTo, e-wallets)
- **RBAC implementation** with role-based permissions
- **Encrypted sensitive data** (financial information, personal data)

**Deployment & Infrastructure:**
- **Docker containers** for all services
- **Docker Compose** for local development
- **Kubernetes manifests** for production deployment
- **Terraform** for infrastructure as code across AWS, Azure, GCP
- **Nginx/Traefik** as API gateway and load balancer

**Monitoring & Observability:**
- **Prometheus + Grafana** for metrics collection and visualization
- **ELK Stack** (Elasticsearch, Logstash, Kibana) for centralized logging
- **Jaeger** for distributed tracing
- **Health check endpoints** for all services

**Indonesian Compliance & Integration:**
- **DJP e-Faktur API integration** for tax compliance
- **BPJS API integration** for social security calculations
- **Bank API integrations** for payment processing
- **Multi-language support** (Indonesian primary, English secondary)
