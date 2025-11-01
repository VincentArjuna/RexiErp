#!/bin/bash

# Nginx Configuration Test Script
# Tests Story 1.2 API Gateway configuration

set -e

echo "🔧 Testing Nginx API Gateway Configuration..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test variables
NGINX_CONTAINER="rexi-api-gateway"
TEST_PASSED=0
TEST_FAILED=0

# Function to print test results
test_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ PASS${NC}: $2"
        ((TEST_PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC}: $2"
        ((TEST_FAILED++))
    fi
}

# Function to run command with retry
run_with_retry() {
    local retries=$1
    shift
    local count=0

    until "$@"; do
        exit_code=$?
        count=$((count + 1))
        if [ $count -lt $retries ]; then
            echo "Retrying... (attempt $count/$retries)"
            sleep 2
        else
            return $exit_code
        fi
    done
}

echo -e "\n${YELLOW}🚀 Starting Nginx Configuration Tests...${NC}"

# Test 1: Validate Nginx configuration syntax
echo -e "\n${YELLOW}📋 Test 1: Nginx Configuration Syntax${NC}"
if docker exec $NGINX_CONTAINER nginx -t > /dev/null 2>&1; then
    test_result 0 "Nginx configuration syntax is valid"
else
    test_result 1 "Nginx configuration syntax has errors"
fi

# Test 2: Check if Nginx is running
echo -e "\n${YELLOW}📋 Test 2: Nginx Process Status${NC}"
if docker ps --format "table {{.Names}}\t{{.Status}}" | grep -q "$NGINX_CONTAINER.*Up"; then
    test_result 0 "Nginx container is running"
else
    test_result 1 "Nginx container is not running"
fi

# Test 3: Basic health check endpoint
echo -e "\n${YELLOW}📋 Test 3: Basic Health Check${NC}"
if curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health | grep -q "200"; then
    test_result 0 "Basic health check endpoint responds with 200"
else
    test_result 1 "Basic health check endpoint failed"
fi

# Test 4: API documentation endpoints
echo -e "\n${YELLOW}📋 Test 4: API Documentation${NC}"
if curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/docs/ | grep -q "200"; then
    test_result 0 "Swagger UI is accessible"
else
    test_result 1 "Swagger UI is not accessible"
fi

if curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/docs/openapi.yaml | grep -q "200"; then
    test_result 0 "OpenAPI specification is accessible"
else
    test_result 1 "OpenAPI specification is not accessible"
fi

# Test 5: API versioning routing
echo -e "\n${YELLOW}📋 Test 5: API Versioning${NC}"
# Test that /api/v1/ routes are properly configured (should get 404 since services aren't running yet)
if curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/v1/auth/login | grep -q "404\|502\|503"; then
    test_result 0 "API v1 routes are configured (expected service unavailable response)"
else
    test_result 1 "API v1 routing is not working"
fi

# Test 6: CORS headers
echo -e "\n${YELLOW}📋 Test 6: CORS Configuration${NC}"
cors_headers=$(curl -s -I -X OPTIONS http://localhost:8080/api/v1/auth/login \
    -H "Origin: http://localhost:3000" \
    -H "Access-Control-Request-Method: POST" \
    -H "Access-Control-Request-Headers: Authorization,Content-Type")

if echo "$cors_headers" | grep -q "Access-Control-Allow-Origin"; then
    test_result 0 "CORS headers are present"
else
    test_result 1 "CORS headers are missing"
fi

# Test 7: Security headers
echo -e "\n${YELLOW}📋 Test 7: Security Headers${NC}"
security_headers=$(curl -s -I http://localhost:8080/health)

for header in "X-Frame-Options" "X-Content-Type-Options" "X-XSS-Protection" "Referrer-Policy"; do
    if echo "$security_headers" | grep -qi "$header"; then
        test_result 0 "$header security header is present"
    else
        test_result 1 "$header security header is missing"
    fi
done

# Test 8: Rate limiting configuration
echo -e "\n${YELLOW}📋 Test 8: Rate Limiting${NC}"
# This test checks if rate limiting zones are defined in the configuration
if docker exec $NGINX_CONTAINER nginx -T 2>/dev/null | grep -q "limit_req_zone"; then
    test_result 0 "Rate limiting zones are configured"
else
    test_result 1 "Rate limiting zones are not configured"
fi

# Test 9: SSL/TLS configuration
echo -e "\n${YELLOW}📋 Test 9: SSL/TLS Configuration${NC}"
if curl -s -o /dev/null -w "%{http_code}" -k https://localhost:8443/health | grep -q "200"; then
    test_result 0 "HTTPS endpoint is accessible"
else
    test_result 1 "HTTPS endpoint is not accessible"
fi

# Test 10: JSON structured logging
echo -e "\n${YELLOW}📋 Test 10: Structured Logging${NC}"
if docker exec $NGINX_CONTAINER nginx -T 2>/dev/null | grep -q "log_format.*json"; then
    test_result 0 "JSON log format is configured"
else
    test_result 1 "JSON log format is not configured"
fi

# Test 11: Error page configuration
echo -e "\n${YELLOW}📋 Test 11: Error Page Configuration${NC}"
if docker exec $NGINX_CONTAINER nginx -T 2>/dev/null | grep -q "error_page.*error.json"; then
    test_result 0 "Custom error pages are configured"
else
    test_result 1 "Custom error pages are not configured"
fi

# Test 12: Upstream service definitions
echo -e "\n${YELLOW}📋 Test 12: Upstream Services${NC}"
expected_services=("authentication_service" "inventory_service" "accounting_service" "hr_service" "crm_service" "notification_service" "integration_service")

for service in "${expected_services[@]}"; do
    if docker exec $NGINX_CONTAINER nginx -T 2>/dev/null | grep -q "upstream $service"; then
        test_result 0 "Upstream service $service is defined"
    else
        test_result 1 "Upstream service $service is not defined"
    fi
done

# Test 13: Request ID generation
echo -e "\n${YELLOW}📋 Test 13: Request ID Generation${NC}"
if docker exec $NGINX_CONTAINER nginx -T 2>/dev/null | grep -q "request_id"; then
    test_result 0 "Request ID generation is configured"
else
    test_result 1 "Request ID generation is not configured"
fi

# Test 14: Gzip compression
echo -e "\n${YELLOW}📋 Test 14: Gzip Compression${NC}"
if docker exec $NGINX_CONTAINER nginx -T 2>/dev/null | grep -q "gzip on"; then
    test_result 0 "Gzip compression is enabled"
else
    test_result 1 "Gzip compression is not enabled"
fi

# Final results
echo -e "\n${YELLOW}📊 Test Results Summary${NC}"
echo -e "Passed: ${GREEN}$TEST_PASSED${NC}"
echo -e "Failed: ${RED}$TEST_FAILED${NC}"
echo -e "Total:  $((TEST_PASSED + TEST_FAILED))"

if [ $TEST_FAILED -eq 0 ]; then
    echo -e "\n${GREEN}🎉 All tests passed! Nginx API Gateway is properly configured.${NC}"
    exit 0
else
    echo -e "\n${RED}❌ $TEST_FAILED tests failed. Please check the configuration.${NC}"
    exit 1
fi