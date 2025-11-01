# Local Development Setup

## Prerequisites

- Docker 27.3.1 or later
- Docker Compose 2.29.0 or later
- Go 1.23.1 (for local development)
- Git

## Quick Start

### 1. Clone Repository
```bash
git clone https://github.com/VincentArjuna/RexiErp.git
cd RexiErp/deployments/docker-compose
```

### 2. Start Services
```bash
docker-compose up -d
```

### 3. Verify Services
```bash
# Check all services are running
docker-compose ps

# Test API Gateway
curl http://localhost:8080/health

# Test with API key
curl -H "X-API-Key: rexierp-api-key-2024-dev" \
     http://localhost:8080/api/v1/health
```

## Service URLs

| Service | URL | Description |
|---------|-----|-------------|
| API Gateway | http://localhost:8080 | Main API entry point |
| Prometheus | http://localhost:9090 | Metrics and monitoring |
| Grafana | http://localhost:3000 | Visualization dashboards |
| RabbitMQ Management | http://localhost:15672 | Message queue management |
| PostgreSQL | localhost:5432 | Primary database |
| Redis | localhost:6379 | Cache and session store |

## Default Credentials

### Grafana
- **Username**: admin
- **Password**: admin

### RabbitMQ Management
- **Username**: guest
- **Password**: guest

### PostgreSQL
- **Database**: rexi_erp
- **Username**: rexi
- **Password**: password

## Development Workflow

### 1. Environment Configuration

Copy the environment template:
```bash
cp .env.example .env
```

Edit `.env` file as needed:
```bash
# API Keys
API_KEYS=rexierp-api-key-2024-dev,your-production-key
API_KEY_AUTH_ENABLED=true

# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_NAME=rexi_erp
DB_USER=rexi
DB_PASSWORD=password

# Redis Configuration
REDIS_HOST=redis
REDIS_PORT=6379

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-for-development-only
JWT_EXPIRATION_HOURS=24
```

### 2. Service Development

#### Running Individual Services
```bash
# Run authentication service
go run cmd/authentication-service/main.go

# Run inventory service
go run cmd/inventory-service/main.go
```

#### Testing Services
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test ./internal/authentication/handler/
```

### 3. Database Migrations

```bash
# Run migrations
go run migrations/migrate.go up

# Create new migration
go run migrations/migrate.go create migration_name

# Rollback migration
go run migrations/migrate.go down
```

## API Testing

### Using curl
```bash
# Health check
curl http://localhost:8080/health

# API with authentication
curl -H "X-API-Key: rexierp-api-key-2024-dev" \
     http://localhost:8080/api/v1/inventory/products

# Post data
curl -X POST \
     -H "Content-Type: application/json" \
     -H "X-API-Key: rexierp-api-key-2024-dev" \
     -d '{"name":"Test Product"}' \
     http://localhost:8080/api/v1/inventory/products
```

### Using HTTPie
```bash
# Install: pip install httpie

http GET localhost:8080/health
http GET localhost:8080/api/v1/inventory/products X-API-Key:rexierp-api-key-2024-dev
```

### Using Postman
1. Import Postman collection: `postman/RexiERP-API.postman_collection.json`
2. Set environment variables:
   - `base_url`: http://localhost:8080
   - `api_key`: rexierp-api-key-2024-dev

## Monitoring and Debugging

### Logs
```bash
# View all service logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f api-gateway
docker-compose logs -f authentication-service

# View last 100 lines
docker-compose logs --tail=100 -f
```

### Metrics
1. Open Grafana: http://localhost:3000
2. Login with admin/admin
3. View pre-configured dashboards:
   - "RexiERP - Nginx API Gateway Metrics"
   - "RexiERP - System Overview"

### Health Checks
```bash
# API Gateway Health
curl http://localhost:8080/health

# Individual Service Health
curl http://localhost:8080/api/v1/health

# Docker Health Status
docker-compose ps
```

## Common Issues

### Port Conflicts
If ports are already in use, modify `docker-compose.yml`:
```yaml
ports:
  - "8081:8080"  # Changed from 8080
```

### Permission Issues
```bash
# Fix Docker permission issues
sudo chown -R $USER:$USER .

# Reset Docker permissions
sudo usermod -aG docker $USER
```

### Database Connection Issues
```bash
# Check database connectivity
docker-compose exec postgres psql -U rexi -d rexi_erp -c "SELECT 1;"

# Reset database
docker-compose down -v
docker-compose up -d postgres
```

### Cache Issues
```bash
# Clear Redis cache
docker-compose exec redis redis-cli FLUSHALL

# Restart Redis
docker-compose restart redis
```

## Performance Tuning

### API Gateway Optimization
- Enable response caching for static content
- Adjust worker connections in nginx.conf
- Tune rate limits based on usage patterns

### Database Optimization
- Monitor connection pool usage
- Add appropriate indexes
- Use connection pooling in application code

### Monitoring Setup
- Configure alerts in Prometheus/Grafana
- Set up log aggregation
- Monitor resource usage

## Security Considerations

### API Keys
- Use different keys for development/staging/production
- Rotate keys regularly
- Monitor API key usage

### Network Security
- API Gateway only exposes necessary ports
- Internal services communicate within Docker network
- Use HTTPS in production

### Data Security
- Sensitive data in environment variables
- Database credentials managed securely
- Enable audit logging

## Production Deployment Notes

This setup is optimized for development. For production:

1. **Security**
   - Use proper SSL certificates
   - Implement firewall rules
   - Use secret management system

2. **Performance**
   - Optimize Docker image sizes
   - Implement horizontal scaling
   - Use load balancers

3. **Monitoring**
   - Set up comprehensive logging
   - Configure alerting
   - Implement health checks

4. **Backup**
   - Database backups
   - Configuration backups
   - Disaster recovery plan

## Troubleshooting

### Service Won't Start
```bash
# Check Docker logs
docker-compose logs service-name

# Check resource usage
docker stats

# Restart specific service
docker-compose restart service-name
```

### API Returns 401 Unauthorized
1. Verify API key is correct
2. Check if API key authentication is enabled
3. Verify header format: `X-API-Key: your-key`

### High Response Times
1. Check resource usage: `docker stats`
2. Review logs for errors
3. Check database performance
4. Monitor metrics in Grafana

### Database Connection Failures
1. Verify database is running: `docker-compose ps`
2. Check database logs: `docker-compose logs postgres`
3. Test connection manually
4. Verify configuration in `.env`

## Support

For development support:
1. Check existing issues in the repository
2. Review logs for error messages
3. Consult the API documentation
4. Contact the development team