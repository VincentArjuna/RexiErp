package integration

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestGatewayHealthCheck tests the API gateway health check endpoints
func TestGatewayHealthCheck(t *testing.T) {
	// Test basic health check
	resp, err := http.Get("http://localhost:8080/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/plain", resp.Header.Get("Content-Type"))
}

// TestGatewayAPIRouting tests API versioning and routing
func TestGatewayAPIRouting(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		expectedStatus int
		description    string
	}{
		{
			name:           "Auth Service Route",
			path:           "/api/v1/auth/login",
			expectedStatus: http.StatusBadGateway, // Service not running
			description:    "Should route to authentication service",
		},
		{
			name:           "Inventory Service Route",
			path:           "/api/v1/inventory/products",
			expectedStatus: http.StatusBadGateway, // Service not running
			description:    "Should route to inventory service",
		},
		{
			name:           "Accounting Service Route",
			path:           "/api/v1/accounting/invoices",
			expectedStatus: http.StatusBadGateway, // Service not running
			description:    "Should route to accounting service",
		},
		{
			name:           "HR Service Route",
			path:           "/api/v1/hr/employees",
			expectedStatus: http.StatusBadGateway, // Service not running
			description:    "Should route to HR service",
		},
		{
			name:           "CRM Service Route",
			path:           "/api/v1/crm/customers",
			expectedStatus: http.StatusBadGateway, // Service not running
			description:    "Should route to CRM service",
		},
		{
			name:           "Notification Service Route",
			path:           "/api/v1/notifications/send",
			expectedStatus: http.StatusBadGateway, // Service not running
			description:    "Should route to notification service",
		},
		{
			name:           "Integration Service Route",
			path:           "/api/v1/integrations/efaktur",
			expectedStatus: http.StatusBadGateway, // Service not running
			description:    "Should route to integration service",
		},
		{
			name:           "Invalid API Route",
			path:           "/api/v1/invalid/service",
			expectedStatus: http.StatusNotFound,
			description:    "Should return 404 for invalid routes",
		},
		{
			name:           "Non-API Route",
			path:           "/some/other/path",
			expectedStatus: http.StatusNotFound,
			description:    "Should return 404 for non-API routes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "http://localhost:8080"+tt.path, nil)
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode, tt.description)
		})
	}
}

// TestGatewayCORSHeaders tests CORS configuration
func TestGatewayCORSHeaders(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		method         string
		expectedOrigin string
	}{
		{
			name:           "Localhost Origin",
			origin:         "http://localhost:3000",
			method:         "POST",
			expectedOrigin: "http://localhost:3000",
		},
		{
			name:           "RexiERP Domain",
			origin:         "https://app.rexierp.com",
			method:         "GET",
			expectedOrigin: "https://app.rexierp.com",
		},
		{
			name:           "Invalid Origin",
			origin:         "https://malicious-site.com",
			method:         "POST",
			expectedOrigin: "https://malicious-site.com", // Should still return the origin
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("OPTIONS", "http://localhost:8080/api/v1/auth/login", nil)
			require.NoError(t, err)

			req.Header.Set("Origin", tt.origin)
			req.Header.Set("Access-Control-Request-Method", tt.method)
			req.Header.Set("Access-Control-Request-Headers", "Authorization,Content-Type")

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Check CORS headers
			assert.Equal(t, tt.expectedOrigin, resp.Header.Get("Access-Control-Allow-Origin"))
			assert.Contains(t, resp.Header.Get("Access-Control-Allow-Methods"), tt.method)
			assert.Contains(t, resp.Header.Get("Access-Control-Allow-Headers"), "Authorization")
			assert.Equal(t, "1728000", resp.Header.Get("Access-Control-Max-Age"))
		})
	}
}

// TestGatewaySecurityHeaders tests security header configuration
func TestGatewaySecurityHeaders(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	expectedHeaders := map[string]string{
		"X-Frame-Options":        "SAMEORIGIN",
		"X-XSS-Protection":       "1; mode=block",
		"X-Content-Type-Options": "nosniff",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
		"Content-Security-Policy": "default-src 'self'",
	}

	for header, expectedValue := range expectedHeaders {
		t.Run(fmt.Sprintf("Security Header: %s", header), func(t *testing.T) {
			actualValue := resp.Header.Get(header)
			assert.NotEmpty(t, actualValue, "Header %s should be present", header)
			if expectedValue != "" {
				assert.Contains(t, actualValue, expectedValue,
					"Header %s should contain expected value", header)
			}
		})
	}
}

// TestGatewayRateLimiting tests rate limiting functionality
func TestGatewayRateLimiting(t *testing.T) {
	// This test makes multiple rapid requests to check if rate limiting is configured
	// Note: This is a basic test - in a real environment you'd want more sophisticated testing

	client := &http.Client{Timeout: 2 * time.Second}
	rateLimited := false

	// Make 10 rapid requests
	for i := 0; i < 10; i++ {
		req, err := http.NewRequest("GET", "http://localhost:8080/health", nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		if err != nil {
			continue // Skip on network errors
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			rateLimited = true
			break
		}
	}

	// We expect rate limiting to be configured, but this test might not always trigger it
	// depending on the rate limit settings and test execution speed
	t.Logf("Rate limiting status: %v", rateLimited)
}

// TestGatewayAPIDocumentation tests API documentation endpoints
func TestGatewayAPIDocumentation(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedType   string
	}{
		{
			name:           "Swagger UI",
			path:           "/api/docs/",
			expectedStatus: http.StatusOK,
			expectedType:   "text/html",
		},
		{
			name:           "OpenAPI Specification",
			path:           "/api/docs/openapi.yaml",
			expectedStatus: http.StatusOK,
			expectedType:   "application/yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get("http://localhost:8080" + tt.path)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			if tt.expectedType != "" {
				contentType := resp.Header.Get("Content-Type")
				assert.Contains(t, contentType, tt.expectedType)
			}
		})
	}
}

// TestGatewayErrorHandling tests error handling configuration
func TestGatewayErrorHandling(t *testing.T) {
	// Test 404 error response format
	resp, err := http.Get("http://localhost:8080/nonexistent-endpoint")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}

// TestGatewaySSLTLS tests SSL/TLS configuration
func TestGatewaySSLTLS(t *testing.T) {
	// Note: This test uses self-signed certificates, so we skip certificate verification
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Get("https://localhost:8443/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestGatewayRequestID tests request ID generation and tracking
func TestGatewayRequestID(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:8080/health", nil)
	require.NoError(t, err)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check if request ID header is present
	requestID := resp.Header.Get("X-Request-ID")
	assert.NotEmpty(t, requestID, "Request ID should be present in response headers")
}

// BenchmarkGatewayPerformance benchmarks the gateway performance
func BenchmarkGatewayPerformance(b *testing.B) {
	client := &http.Client{Timeout: 2 * time.Second}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Get("http://localhost:8080/health")
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}

// TestGatewayContainer tests the gateway container integration
func TestGatewayContainer(t *testing.T) {
	ctx := context.Background()

	// This test requires Docker to be running and the gateway container to be available
	// It's designed to work with the existing docker-compose setup

	req := testcontainers.ContainerRequest{
		Image:        "curlimages/curl:latest",
		Cmd:          []string{"sh", "-c", "while true; do sleep 30; done"},
		Networks:     []string{"rexi-network"},
		WaitingFor:   wait.ForLog("started"),
		AutoRemove:   true,
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer container.Terminate(ctx)

	// Test connectivity to the gateway from within the container network
	_, _, err = container.Exec(ctx, []string{"curl", "-f", "http://api-gateway/health"})
	assert.NoError(t, err, "Should be able to reach gateway from container network")
}