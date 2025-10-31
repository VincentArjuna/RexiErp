# Infrastructure and Deployment

## Infrastructure as Code

- **Tool:** Docker Compose 2.29.0 (Phase 1), Terraform 1.9.8 (Phase 2)
- **Location:** `deployments/docker-compose/` (Phase 1), `infrastructure/terraform/` (Phase 2)
- **Approach:** Local-first development with containerized services, progressing to cloud infrastructure

## Deployment Strategy

- **Strategy:** Blue-Green Deployment with Canary releases
- **CI/CD Platform:** GitHub Actions
- **Pipeline Configuration:** `.github/workflows/`

## Environments

- **Local Development:** Docker Compose with hot reload
- **Staging:** Production-like environment for testing Indonesian integrations
- **Production:** Multi-region deployment (Indonesia primary, Singapore DR)

## Environment Promotion Flow

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

## Local Docker Compose Configuration (Phase 1)

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

## Rollback Strategy

- **Primary Method:** Blue-Green deployment with instant traffic switching
- **Trigger Conditions:** Error rate >5%, response time >2s, Indonesian API failures >10%
- **Recovery Time Objective:** 5 minutes for full rollback
