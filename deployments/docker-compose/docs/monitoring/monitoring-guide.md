# Monitoring and Metrics Guide

## Overview

RexiERP includes comprehensive monitoring using Prometheus and Grafana. This guide covers setup, configuration, and usage of the monitoring stack.

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Nginx         │    │   Services      │    │   Infrastructure│
│   API Gateway   │───▶│   (Go Apps)     │───▶│   (PostgreSQL,  │
│                 │    │                 │    │    Redis,       │
│   /metrics      │    │   /metrics      │    │    RabbitMQ)    │
│   /nginx_status │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   Prometheus    │
                    │   (Scraping)    │
                    │                 │
                    │   :9090         │
                    └─────────────────┘
                                 │
                    ┌─────────────────┐
                    │   Grafana       │
                    │   (Dashboard)   │
                    │                 │
                    │   :3000         │
                    └─────────────────┘
```

## Quick Start

### 1. Access Grafana
```
URL: http://localhost:3000
Username: admin
Password: admin
```

### 2. Access Prometheus
```
URL: http://localhost:9090
```

### 3. View Pre-built Dashboards
- "RexiERP - Nginx API Gateway Metrics"
- "RexiERP - System Overview" (coming soon)

## Available Metrics

### Nginx API Gateway

#### Connection Metrics
- `nginx_connections_active` - Currently active connections
- `nginx_connections_reading` - Connections reading request header
- `nginx_connections_writing` - Connections writing response
- `nginx_connections_waiting` - Idle connections

#### Request Metrics
- `nginx_requests_total` - Total number of requests
- `api_requests_total` - API requests by method and status

#### Cache Metrics
- `nginx_cache_hits_total` - Number of cache hits
- `nginx_cache_misses_total` - Number of cache misses

#### Status Metrics
- `nginx_up` - Nginx server status (1=up, 0=down)

### Application Metrics (Go Services)

#### HTTP Metrics
- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request duration histogram
- `http_response_size_bytes` - Response size histogram

#### Database Metrics
- `db_connections_active` - Active database connections
- `db_query_duration_seconds` - Query execution time
- `db_transactions_total` - Database transaction count

#### Business Metrics
- `users_registered_total` - Total registered users
- `orders_created_total` - Total orders created
- `revenue_total` - Total revenue

### Infrastructure Metrics

#### PostgreSQL
- `pg_stat_database_numbackends` - Number of connections
- `pg_stat_database_xact_commit` - Transaction commits
- `pg_stat_database_xact_rollback` - Transaction rollbacks

#### Redis
- `redis_connected_clients` - Connected clients
- `redis_used_memory_bytes` - Memory usage
- `redis_commands_processed_total` - Commands processed

#### RabbitMQ
- `rabbitmq_queues_messages_ready` - Messages ready for delivery
- `rabbitmq_queues_messages_unacknowledged` - Unacknowledged messages

## Grafana Dashboards

### Nginx API Gateway Dashboard

**Panels:**
1. **Active Connections** - Current active connections with threshold alerts
2. **Requests per Second** - Request rate over time
3. **Response Time** - P50 and P95 response times
4. **Connections Status** - Reading, writing, waiting connections
5. **Cache Hit Rate** - Cache efficiency percentage
6. **HTTP Status Codes** - Distribution of response codes
7. **API Requests by Method** - GET, POST, PUT, DELETE distribution
8. **Server Status** - Up/down status indicator

### Creating Custom Dashboards

#### Step 1: Create New Dashboard
1. Go to Grafana → Dashboards → New Dashboard
2. Add panels with Prometheus queries

#### Step 2: Common Queries

**Request Rate:**
```
rate(nginx_requests_total[5m])
```

**Error Rate:**
```
sum(rate(api_requests_total{status=~"5.."}[5m])) / sum(rate(api_requests_total[5m]))
```

**Response Time Percentiles:**
```
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
histogram_quantile(0.50, rate(http_request_duration_seconds_bucket[5m]))
```

**Cache Hit Rate:**
```
nginx_cache_hits_total / (nginx_cache_hits_total + nginx_cache_misses_total) * 100
```

## Alerting

### Prometheus Alerts

Create alerts in `prometheus.yml`:

```yaml
rule_files:
  - "alert_rules.yml"

# alert_rules.yml
groups:
  - name: rexi_erp_alerts
    rules:
      - alert: HighErrorRate
        expr: sum(rate(api_requests_total{status=~"5.."}[5m])) / sum(rate(api_requests_total[5m])) > 0.05
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value | humanizePercentage }}"

      - alert: HighResponseTime
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 2
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High response time detected"
          description: "95th percentile response time is {{ $value }}s"
```

### Grafana Alerts

1. **Create Alert Rule:**
   - Go to Dashboard → Panel → Configure
   - Click "Alert" tab
   - Set conditions and notifications

2. **Alert Channels:**
   - Email
   - Slack
   - PagerDuty
   - Webhook

## Query Examples

### System Health
```promql
# Overall system health
up{job="nginx-api-gateway"}

# Service health
up{job=~"authentication-service|inventory-service|accounting-service"}
```

### Performance Analysis
```promql
# Request rate by service
sum by (job) (rate(http_requests_total[5m]))

# Response time by endpoint
histogram_quantile(0.95, sum by (le, endpoint) (rate(http_request_duration_seconds_bucket[5m])))

# Error rate by service
sum by (job) (rate(http_requests_total{status=~"5.."}[5m]))
```

### Resource Usage
```promql
# Memory usage
process_resident_memory_bytes / 1024 / 1024

# CPU usage
rate(process_cpu_seconds_total[5m])

# Database connections
pg_stat_database_numbackends
```

## Troubleshooting Monitoring

### Prometheus Issues

**Prometheus not scraping:**
```bash
# Check Prometheus logs
docker-compose logs prometheus

# Test scrape endpoint
curl http://localhost:8080/metrics

# Check Prometheus configuration
docker-compose exec prometheus cat /etc/prometheus/prometheus.yml
```

**Missing metrics:**
1. Verify service is exposing `/metrics` endpoint
2. Check network connectivity between services
3. Verify service discovery configuration

### Grafana Issues

**Dashboard not loading:**
1. Check Prometheus data source connection
2. Verify metric names exist in Prometheus
3. Check query syntax

**Incorrect data:**
1. Verify time range is appropriate
2. Check for label mismatches
3. Confirm aggregation functions

## Best Practices

### Metric Naming
- Use consistent naming conventions
- Include units in metric names
- Use appropriate metric types (Counter, Gauge, Histogram)

### Labeling
- Use meaningful labels
- Avoid high cardinality labels
- Document label purposes

### Performance
- Set appropriate scrape intervals
- Use recording rules for complex queries
- Monitor Prometheus performance

### Retention
- Configure appropriate data retention periods
- Use remote storage for long-term data
- Implement data archiving policies

## Advanced Configuration

### Custom Metrics in Go Services

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    requestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )
)

func init() {
    prometheus.MustRegister(requestsTotal)
    prometheus.MustRegister(requestDuration)
}

// Middleware to record metrics
func metricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Custom response writer to capture status code
        rw := &responseWriter{ResponseWriter: w, statusCode: 200}

        next.ServeHTTP(rw, r)

        duration := time.Since(start).Seconds()

        requestsTotal.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(rw.statusCode)).Inc()
        requestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
    })
}
```

### Recording Rules

Create `recording_rules.yml`:
```yaml
groups:
  - name: rexi_erp_recording_rules
    interval: 30s
    rules:
      - record: instance:nginx_requests:rate5m
        expr: sum by (instance) (rate(nginx_requests_total[5m]))

      - record: instance:http_request_duration_seconds:rate5m
        expr: histogram_quantile(0.95, sum by (le, instance) (rate(http_request_duration_seconds_bucket[5m])))
```

## Scaling Considerations

### Prometheus Scaling
- Use federation for multi-environment setups
- Implement remote storage (Thanos, Cortex)
- Consider sharding for large deployments

### Grafana Scaling
- Use Grafana Cloud for managed service
- Implement dashboard versioning
- Set up user management and permissions

## Security

### Access Control
- Enable authentication in Grafana
- Use network policies for service communication
- Implement RBAC for dashboard access

### Data Protection
- Encrypt sensitive metrics
- Use TLS for all communications
- Audit metric access

## Support and Resources

### Documentation
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Nginx Monitoring Guide](https://www.nginx.com/blog/monitoring-nginx-plus-metrics-prometheus-grafana/)

### Community
- Prometheus Slack Channel
- Grafana Community Forums
- Nginx Community Mailing List

### Getting Help
1. Check logs for error messages
2. Verify configuration files
3. Test metrics endpoints manually
4. Consult documentation
5. Contact monitoring team