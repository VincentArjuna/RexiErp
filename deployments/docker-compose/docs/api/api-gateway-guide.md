# RexiERP API Gateway Guide

## Overview

The RexiERP API Gateway is built using Nginx and provides a unified entry point for all microservices. It handles authentication, rate limiting, caching, and request routing.

## Base URLs

- **Development**: `http://localhost:8080`
- **Production**: `https://api.rexierp.com`

## Authentication

### API Key Authentication

All API endpoints (except health checks and documentation) require API key authentication.

**Header**: `X-API-Key`
**Value**: Your API key (e.g., `rexierp-api-key-2024-dev`)

**Example Request**:
```bash
curl -H "X-API-Key: rexierp-api-key-2024-dev" \
     http://localhost:8080/api/v1/auth/login
```

### Bearer Token Alternative

API keys can also be provided via the Authorization header:

```bash
curl -H "Authorization: Bearer rexierp-api-key-2024-dev" \
     http://localhost:8080/api/v1/auth/login
```

## Rate Limiting

The API Gateway implements different rate limits for various endpoint types:

| Endpoint Type | Rate Limit | Burst |
|---------------|------------|-------|
| Global | 100 req/s | 200 |
| Authentication | 10 req/s | 20 |
| General API | 50 req/s | 50 |
| File Upload | 5 req/s | - |
| Integration | 20 req/s | 20 |

## Response Headers

### Security Headers

- `X-Frame-Options: SAMEORIGIN`
- `X-XSS-Protection: 1; mode=block`
- `X-Content-Type-Options: nosniff`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Content-Security-Policy: default-src 'self'; ...`

### Caching Headers

Cached endpoints include:
- `X-Cache-Status: HIT/MISS/EXPIRED`
- `Cache-Control: public, max-age=...`

## Error Responses

All errors follow a consistent format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "request_id": "unique-request-id",
    "timestamp": "2024-01-01T12:00:00Z"
  }
}
```

### Common HTTP Status Codes

- **200 OK**: Request successful
- **201 Created**: Resource created successfully
- **400 Bad Request**: Invalid request parameters
- **401 Unauthorized**: Missing or invalid API key
- **403 Forbidden**: Access denied
- **404 Not Found**: Resource not found
- **429 Too Many Requests**: Rate limit exceeded
- **500 Internal Server Error**: Server error

## API Endpoints

### Health Checks

#### Basic Health Check
```bash
GET /health
```

#### Service Health Check
```bash
GET /api/v1/health
GET /api/v1/health/ready
GET /api/v1/health/live
```

### Documentation

#### Swagger UI
```bash
GET /api/docs
```

#### OpenAPI Specification
```bash
GET /api/docs/openapi.yaml
```

### Metrics

#### Prometheus Metrics
```bash
GET /metrics
```

#### Nginx Status
```bash
GET /nginx_status
```
*Restricted access (internal networks only)*

## Service Routes

| Service | Base Path | Description |
|---------|-----------|-------------|
| Authentication | `/api/v1/auth/` | User authentication and authorization |
| Inventory | `/api/v1/inventory/` | Product and inventory management |
| Accounting | `/api/v1/accounting/` | Financial accounting and reporting |
| HR | `/api/v1/hr/` | Human resources and payroll |
| CRM | `/api/v1/crm/` | Customer relationship management |
| Notifications | `/api/v1/notifications/` | Email, SMS, and push notifications |
| Integrations | `/api/v1/integrations/` | Third-party service integrations |

## Request/Response Examples

### Authentication Service

#### Login
```bash
curl -X POST \
     -H "Content-Type: application/json" \
     -H "X-API-Key: rexierp-api-key-2024-dev" \
     -d '{"email":"user@example.com","password":"password"}' \
     http://localhost:8080/api/v1/auth/login
```

### Inventory Service

#### Get Products
```bash
curl -H "X-API-Key: rexierp-api-key-2024-dev" \
     http://localhost:8080/api/v1/inventory/products
```

#### Create Product
```bash
curl -X POST \
     -H "Content-Type: application/json" \
     -H "X-API-Key: rexierp-api-key-2024-dev" \
     -d '{"name":"Product Name","price":99.99,"stock":100}' \
     http://localhost:8080/api/v1/inventory/products
```

## SDK Examples

### Go
```go
import "net/http"

func callAPI() {
    req, _ := http.NewRequest("GET", "http://localhost:8080/api/v1/inventory/products", nil)
    req.Header.Set("X-API-Key", "rexierp-api-key-2024-dev")

    client := &http.Client{}
    resp, err := client.Do(req)
    // Handle response
}
```

### JavaScript
```javascript
const apiKey = 'rexierp-api-key-2024-dev';

fetch('http://localhost:8080/api/v1/inventory/products', {
    headers: {
        'X-API-Key': apiKey
    }
})
.then(response => response.json())
.then(data => console.log(data));
```

### Python
```python
import requests

headers = {'X-API-Key': 'rexierp-api-key-2024-dev'}
response = requests.get(
    'http://localhost:8080/api/v1/inventory/products',
    headers=headers
)
data = response.json()
```

## Best Practices

1. **Always include API key** in requests to protected endpoints
2. **Handle rate limiting** gracefully with exponential backoff
3. **Use appropriate HTTP methods** (GET, POST, PUT, DELETE)
4. **Validate input** before sending requests
5. **Monitor request IDs** for debugging support requests
6. **Cache responses** when appropriate to reduce load

## Support

For API support, include the `request_id` from error responses in your support requests.

**Email**: api-support@rexierp.com
**Documentation**: https://docs.rexierp.com/api