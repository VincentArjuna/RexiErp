package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// Monitoring service endpoints
	PrometheusURL   = "http://localhost:9090"
	GrafanaURL      = "http://localhost:3000"
	ElasticsearchURL = "http://localhost:9200"
	KibanaURL       = "http://localhost:5601"
	JaegerURL       = "http://localhost:16686"
	AlertManagerURL = "http://localhost:9093"
)

// TestMonitoringStackHealth tests that all monitoring services are healthy
func TestMonitoringStackHealth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Wait for services to be ready
	require.Eventually(t, func() bool {
		return isServiceHealthy(t, PrometheusURL+"/-/healthy")
	}, 2*time.Minute, 5*time.Second, "Prometheus should become healthy")

	require.Eventually(t, func() bool {
		return isServiceHealthy(t, GrafanaURL+"/api/health")
	}, 2*time.Minute, 5*time.Second, "Grafana should become healthy")

	require.Eventually(t, func() bool {
		return isServiceHealthy(t, ElasticsearchURL+"/_cluster/health")
	}, 2*time.Minute, 5*time.Second, "Elasticsearch should become healthy")

	require.Eventually(t, func() bool {
		return isServiceHealthy(t, KibanaURL+"/api/status")
	}, 2*time.Minute, 5*time.Second, "Kibana should become healthy")

	require.Eventually(t, func() bool {
		return isServiceHealthy(t, JaegerURL+"/api/services")
	}, 2*time.Minute, 5*time.Second, "Jaeger should become healthy")

	require.Eventually(t, func() bool {
		return isServiceHealthy(t, AlertManagerURL+"/-/healthy")
	}, 2*time.Minute, 5*time.Second, "AlertManager should become healthy")

	t.Log("All monitoring services are healthy")
}

// TestPrometheusMetrics tests Prometheus metrics collection
func TestPrometheusMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test Prometheus targets
	resp, err := http.Get(PrometheusURL + "/api/v1/targets")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var targetsResp struct {
		Status string `json:"status"`
		Data   struct {
			ActiveTargets []struct {
				Labels map[string]string `json:"labels"`
				Health string             `json:"health"`
			} `json:"activeTargets"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&targetsResp)
	require.NoError(t, err)

	assert.Equal(t, "success", targetsResp.Status)

	// Check that we have monitoring targets
	activeTargets := targetsResp.Data.ActiveTargets
	assert.Greater(t, len(activeTargets), 0, "Should have active targets")

	// Check for specific targets
	targetNames := make(map[string]bool)
	for _, target := range activeTargets {
		if job, exists := target.Labels["job"]; exists {
			targetNames[job] = true
		}
	}

	expectedTargets := []string{
		"prometheus",
		"alertmanager",
		"elasticsearch",
		"jaeger-collector",
		"jaeger-query",
	}

	for _, expected := range expectedTargets {
		assert.True(t, targetNames[expected], "Should have target: %s", expected)
	}

	t.Log("Prometheus targets are configured correctly")
}

// TestPrometheusQueries tests Prometheus query functionality
func TestPrometheusQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test basic query
	testQuery := "up"
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/query?query=%s", PrometheusURL, testQuery))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var queryResp struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Value  []interface{}      `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&queryResp)
	require.NoError(t, err)

	assert.Equal(t, "success", queryResp.Status)
	assert.Equal(t, "vector", queryResp.Data.ResultType)
	assert.Greater(t, len(queryResp.Data.Result), 0, "Query should return results")

	t.Log("Prometheus queries are working")
}

// TestGrafanaDatasources tests Grafana datasource configuration
func TestGrafanaDatasources(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test Grafana API with authentication
	client := &http.Client{}
	req, err := http.NewRequest("GET", GrafanaURL+"/api/datasources", nil)
	require.NoError(t, err)

	// Use default admin credentials
	req.SetBasicAuth("admin", "admin")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var datasources []struct {
		Name string `json:"name"`
		Type string `json:"type"`
		URL  string `json:"url"`
	}

	err = json.NewDecoder(resp.Body).Decode(&datasources)
	require.NoError(t, err)

	// Check for expected datasources
	expectedDatasources := map[string]string{
		"Prometheus":   "prometheus",
		"Elasticsearch": "elasticsearch",
		"Jaeger":       "jaeger",
	}

	dsNames := make(map[string]string)
	for _, ds := range datasources {
		dsNames[ds.Name] = ds.Type
	}

	for name, dsType := range expectedDatasources {
		assert.Equal(t, dsType, dsNames[name], "Should have datasource: %s of type: %s", name, dsType)
	}

	t.Log("Grafana datasources are configured correctly")
}

// TestElasticsearchIndex tests Elasticsearch index creation and searching
func TestElasticsearchIndex(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Wait for Elasticsearch to be ready
	require.Eventually(t, func() bool {
		resp, err := http.Get(ElasticsearchURL + "/_cluster/health?wait_for_status=yellow&timeout=50s")
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 2*time.Minute, 5*time.Second, "Elasticsearch should become ready")

	// Create a test index
	indexName := "rexierp-test-logs-" + time.Now().Format("2006.01.02")
	indexDoc := map[string]interface{}{
		"message":     "Test log message",
		"service":     "test-service",
		"level":       "info",
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"correlation_id": "test-123",
	}

	docBytes, err := json.Marshal(indexDoc)
	require.NoError(t, err)

	// Index document
	req, err := http.NewRequest("PUT", ElasticsearchURL+"/"+indexName+"/_doc/1", bytes.NewBuffer(docBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Refresh index to make document searchable
	req, err = http.NewRequest("POST", ElasticsearchURL+"/"+indexName+"/_refresh", nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Search for document
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"message": "test log message",
			},
		},
	}

	queryBytes, err := json.Marshal(searchQuery)
	require.NoError(t, err)

	req, err = http.NewRequest("GET", ElasticsearchURL+"/"+indexName+"/_search", bytes.NewBuffer(queryBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var searchResp struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	err = json.NewDecoder(resp.Body).Decode(&searchResp)
	require.NoError(t, err)

	assert.Greater(t, searchResp.Hits.Total.Value, 0, "Search should find documents")
	assert.Len(t, searchResp.Hits.Hits, 1, "Should find exactly one document")

	// Verify document content
	source := searchResp.Hits.Hits[0].Source
	assert.Equal(t, "Test log message", source["message"])
	assert.Equal(t, "test-service", source["service"])
	assert.Equal(t, "test-123", source["correlation_id"])

	t.Log("Elasticsearch indexing and search are working")
}

// TestJaegerTracing tests Jaeger tracing functionality
func TestJaegerTracing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test Jaeger services endpoint
	resp, err := http.Get(JaegerURL + "/api/services")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var services []string
	err = json.NewDecoder(resp.Body).Decode(&services)
	require.NoError(t, err)

	t.Logf("Jaeger services: %v", services)

	// Test Jaeger traces endpoint (even if empty, should respond)
	resp, err = http.Get(JaegerURL + "/api/traces?service=jaeger-query&limit=10")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var tracesResp struct {
		Data []interface{} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&tracesResp)
	require.NoError(t, err)

	// Should at least have the data structure
	assert.NotNil(t, tracesResp.Data)

	t.Log("Jaeger tracing endpoints are working")
}

// TestAlertManagerRules tests AlertManager rule configuration
func TestAlertManagerRules(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test AlertManager API
	resp, err := http.Get(AlertManagerURL + "/api/v1/rules")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var rulesResp struct {
		Status string `json:"status"`
		Data   struct {
			Groups []struct {
				Name  string `json:"name"`
				Rules []struct {
					Name   string `json:"name"`
					State  string `json:"state"`
					Health string `json:"health"`
				} `json:"rules"`
			} `json:"groups"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&rulesResp)
	require.NoError(t, err)

	assert.Equal(t, "success", rulesResp.Status)
	assert.Greater(t, len(rulesResp.Data.Groups), 0, "Should have rule groups")

	// Check for expected rule groups
	groupNames := make(map[string]bool)
	for _, group := range rulesResp.Data.Groups {
		groupNames[group.Name] = true
		assert.Greater(t, len(group.Rules), 0, "Group %s should have rules", group.Name)
	}

	expectedGroups := []string{"rexierp.rules"}
	for _, expected := range expectedGroups {
		assert.True(t, groupNames[expected], "Should have rule group: %s", expected)
	}

	t.Log("AlertManager rules are configured correctly")
}

// TestLogFlow tests end-to-end log flow through ELK stack
func TestLogFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test would require Filebeat or similar log shipping agent
	// For now, we'll test the basic ELK stack connectivity

	// Test that Logstash can process logs
	resp, err := http.Get("http://localhost:9600/_node/stats?pretty")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var statsResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&statsResp)
	require.NoError(t, err)

	// Should have node stats
	assert.Contains(t, statsResp, "node_id")

	t.Log("Logstash is running and collecting stats")
}

// TestGrafanaDashboards tests Grafana dashboard provisioning
func TestGrafanaDashboards(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test Grafana dashboards API
	client := &http.Client{}
	req, err := http.NewRequest("GET", GrafanaURL+"/api/search?type=dash-db", nil)
	require.NoError(t, err)

	// Use default admin credentials
	req.SetBasicAuth("admin", "admin")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var dashboards []struct {
		Title string `json:"title"`
		UID   string `json:"uid"`
		Tags  []string `json:"tags"`
	}

	err = json.NewDecoder(resp.Body).Decode(&dashboards)
	require.NoError(t, err)

	// Should have provisioned dashboards
	assert.Greater(t, len(dashboards), 0, "Should have provisioned dashboards")

	// Check for expected dashboards
	expectedDashboards := []string{
		"RexiERP - System Health",
		"RexiERP - Application Metrics",
		"RexiERP - Business Metrics",
	}

	dashboardTitles := make(map[string]bool)
	for _, dashboard := range dashboards {
		dashboardTitles[dashboard.Title] = true
	}

	for _, expected := range expectedDashboards {
		assert.True(t, dashboardTitles[expected], "Should have dashboard: %s", expected)
	}

	t.Log("Grafana dashboards are provisioned correctly")
}

// Helper functions

func isServiceHealthy(t *testing.T, url string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func waitForService(t *testing.T, url string, timeout time.Duration) {
	require.Eventually(t, func() bool {
		return isServiceHealthy(t, url)
	}, timeout, 5*time.Second, fmt.Sprintf("Service at %s should become healthy", url))
}