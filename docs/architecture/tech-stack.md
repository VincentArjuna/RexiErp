# Tech Stack

**CRITICAL:** This section is the DEFINITIVE technology selection for RexiERP. All architectural decisions depend on these choices. Please review carefully as these selections will guide all development.

## Cloud Infrastructure

- **Provider:** Multi-cloud capable (AWS, GCP, Azure support via Kubernetes)
- **Key Services:** Container orchestration, managed PostgreSQL, managed Redis, object storage, CDN
- **Deployment Regions:** Indonesia (Jakarta primary), Singapore secondary for DR

## Technology Stack Table

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

## Development Phases

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
