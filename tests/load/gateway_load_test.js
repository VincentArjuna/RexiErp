#!/usr/bin/env node

/**
 * API Gateway Load Testing Script
 * Tests Story 1.2 API Gateway performance under load
 *
 * Usage: node gateway_load_test.js [options]
 *
 * Options:
 *   --concurrent <number>  Number of concurrent connections (default: 10)
 *   --duration <seconds>   Test duration in seconds (default: 30)
 *   --url <url>           Base URL to test (default: http://localhost:8080)
 *   --help                Show this help message
 */

const http = require('http');
const https = require('https');
const { performance } = require('perf_hooks');

// Parse command line arguments
const args = process.argv.slice(2);
let concurrentConnections = 10;
let testDuration = 30;
let baseUrl = 'http://localhost:8080';

for (let i = 0; i < args.length; i++) {
    switch (args[i]) {
        case '--concurrent':
            concurrentConnections = parseInt(args[++i]) || 10;
            break;
        case '--duration':
            testDuration = parseInt(args[++i]) || 30;
            break;
        case '--url':
            baseUrl = args[++i] || 'http://localhost:8080';
            break;
        case '--help':
            console.log(`
API Gateway Load Testing Script

Usage: node gateway_load_test.js [options]

Options:
  --concurrent <number>  Number of concurrent connections (default: 10)
  --duration <seconds>   Test duration in seconds (default: 30)
  --url <url>           Base URL to test (default: http://localhost:8080)
  --help                Show this help message

Examples:
  node gateway_load_test.js --concurrent 20 --duration 60
  node gateway_load_test.js --url https://localhost:8443 --concurrent 50
            `);
            process.exit(0);
            break;
    }
}

// Colors for console output
const colors = {
    green: '\x1b[32m',
    red: '\x1b[31m',
    yellow: '\x1b[33m',
    blue: '\x1b[34m',
    reset: '\x1b[0m'
};

// Test scenarios
const testScenarios = [
    {
        name: 'Health Check',
        path: '/health',
        method: 'GET',
        headers: {},
        expectedStatus: 200
    },
    {
        name: 'API Documentation',
        path: '/api/docs/',
        method: 'GET',
        headers: {},
        expectedStatus: 200
    },
    {
        name: 'Auth Service Route',
        path: '/api/v1/auth/login',
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Origin': 'http://localhost:3000'
        },
        body: JSON.stringify({
            email: 'test@example.com',
            password: 'testpass123'
        }),
        expectedStatus: [502, 503, 504] // Service unavailable
    },
    {
        name: 'Inventory Service Route',
        path: '/api/v1/inventory/products',
        method: 'GET',
        headers: {
            'Authorization': 'Bearer fake-token',
            'Origin': 'http://localhost:3000'
        },
        expectedStatus: [502, 503, 504] // Service unavailable
    },
    {
        name: 'CORS Preflight',
        path: '/api/v1/auth/login',
        method: 'OPTIONS',
        headers: {
            'Origin': 'http://localhost:3000',
            'Access-Control-Request-Method': 'POST',
            'Access-Control-Request-Headers': 'Authorization,Content-Type'
        },
        expectedStatus: 204
    }
];

// Statistics tracking
class StatsCollector {
    constructor() {
        this.reset();
    }

    reset() {
        this.totalRequests = 0;
        this.successfulRequests = 0;
        this.failedRequests = 0;
        this.responseTimes = [];
        this.statusCodes = {};
        this.errors = {};
        this.startTime = null;
        this.endTime = null;
    }

    recordRequest(responseTime, statusCode, error = null) {
        this.totalRequests++;

        if (error) {
            this.failedRequests++;
            this.errors[error] = (this.errors[error] || 0) + 1;
        } else if (statusCode >= 200 && statusCode < 400) {
            this.successfulRequests++;
        } else {
            this.failedRequests++;
        }

        this.responseTimes.push(responseTime);
        this.statusCodes[statusCode] = (this.statusCodes[statusCode] || 0) + 1;
    }

    getStats() {
        const sortedTimes = this.responseTimes.sort((a, b) => a - b);
        const totalTime = this.endTime - this.startTime;

        return {
            totalRequests: this.totalRequests,
            successfulRequests: this.successfulRequests,
            failedRequests: this.failedRequests,
            successRate: ((this.successfulRequests / this.totalRequests) * 100).toFixed(2),
            requestsPerSecond: (this.totalRequests / (totalTime / 1000)).toFixed(2),
            avgResponseTime: (this.responseTimes.reduce((a, b) => a + b, 0) / this.responseTimes.length).toFixed(2),
            minResponseTime: Math.min(...this.responseTimes).toFixed(2),
            maxResponseTime: Math.max(...this.responseTimes).toFixed(2),
            p50ResponseTime: sortedTimes[Math.floor(sortedTimes.length * 0.5)].toFixed(2),
            p95ResponseTime: sortedTimes[Math.floor(sortedTimes.length * 0.95)].toFixed(2),
            p99ResponseTime: sortedTimes[Math.floor(sortedTimes.length * 0.99)].toFixed(2),
            statusCodes: this.statusCodes,
            errors: this.errors,
            duration: (totalTime / 1000).toFixed(2)
        };
    }
}

// HTTP client based on URL protocol
function getHttpClient(url) {
    return url.startsWith('https') ? https : http;
}

// Make HTTP request
function makeRequest(url, options) {
    return new Promise((resolve) => {
        const startTime = performance.now();
        const client = getHttpClient(url);

        const req = client.request(url, options, (res) => {
            let data = '';

            res.on('data', (chunk) => {
                data += chunk;
            });

            res.on('end', () => {
                const endTime = performance.now();
                const responseTime = endTime - startTime;
                resolve({
                    statusCode: res.statusCode,
                    responseTime,
                    headers: res.headers,
                    error: null
                });
            });
        });

        req.on('error', (error) => {
            const endTime = performance.now();
            const responseTime = endTime - startTime;
            resolve({
                statusCode: 0,
                responseTime,
                headers: {},
                error: error.message
            });
        });

        if (options.body) {
            req.write(options.body);
        }

        req.setTimeout(10000, () => {
            req.destroy();
        });

        req.end();
    });
}

// Load test worker
async function loadTestWorker(scenario, duration, stats) {
    const url = baseUrl + scenario.path;
    const options = {
        method: scenario.method,
        headers: scenario.headers,
        timeout: 10000
    };

    if (scenario.body) {
        options.body = scenario.body;
    }

    const startTime = performance.now();
    const endTime = startTime + (duration * 1000);

    while (performance.now() < endTime) {
        const result = await makeRequest(url, options);

        // Check if response status is expected
        const expectedStatuses = Array.isArray(scenario.expectedStatus)
            ? scenario.expectedStatus
            : [scenario.expectedStatus];

        const isExpectedStatus = expectedStatuses.includes(result.statusCode) ||
                               (scenario.expectedStatus.includes && scenario.expectedStatus.includes(result.statusCode));

        if (!isExpectedStatus && result.error === null) {
            console.warn(`${colors.yellow}Warning: Unexpected status ${result.statusCode} for ${scenario.name}${colors.reset}`);
        }

        stats.recordRequest(result.responseTime, result.statusCode, result.error);
    }
}

// Run load test for a scenario
async function runScenarioTest(scenario, concurrentConnections, duration) {
    console.log(`\n${colors.blue}🔄 Testing: ${scenario.name}${colors.reset}`);
    console.log(`   URL: ${baseUrl}${scenario.path}`);
    console.log(`   Method: ${scenario.method}`);
    console.log(`   Concurrent connections: ${concurrentConnections}`);
    console.log(`   Duration: ${duration}s`);

    const stats = new StatsCollector();
    const workers = [];

    stats.startTime = performance.now();

    // Start concurrent workers
    for (let i = 0; i < concurrentConnections; i++) {
        workers.push(loadTestWorker(scenario, duration, stats));
    }

    // Wait for all workers to complete
    await Promise.all(workers);

    stats.endTime = performance.now();

    return stats.getStats();
}

// Display results
function displayResults(scenarioName, stats) {
    console.log(`\n${colors.green}📊 Results for: ${scenarioName}${colors.reset}`);
    console.log(`   Total requests: ${stats.totalRequests}`);
    console.log(`   Successful: ${stats.successfulRequests} (${stats.successRate}%)`);
    console.log(`   Failed: ${stats.failedRequests}`);
    console.log(`   Requests/sec: ${stats.requestsPerSecond}`);
    console.log(`   Response time (ms):`);
    console.log(`     Average: ${stats.avgResponseTime}`);
    console.log(`     Min: ${stats.minResponseTime}`);
    console.log(`     Max: ${stats.maxResponseTime}`);
    console.log(`     50th percentile: ${stats.p50ResponseTime}`);
    console.log(`     95th percentile: ${stats.p95ResponseTime}`);
    console.log(`     99th percentile: ${stats.p99ResponseTime}`);

    if (Object.keys(stats.statusCodes).length > 0) {
        console.log(`   Status codes:`);
        Object.entries(stats.statusCodes).forEach(([code, count]) => {
            const percentage = ((count / stats.totalRequests) * 100).toFixed(1);
            console.log(`     ${code}: ${count} (${percentage}%)`);
        });
    }

    if (Object.keys(stats.errors).length > 0) {
        console.log(`   Errors:`);
        Object.entries(stats.errors).forEach(([error, count]) => {
            console.log(`     ${error}: ${count}`);
        });
    }
}

// Main execution
async function main() {
    console.log(`${colors.blue}🚀 API Gateway Load Testing${colors.reset}`);
    console.log(`Base URL: ${baseUrl}`);
    console.log(`Concurrent connections: ${concurrentConnections}`);
    console.log(`Test duration: ${testDuration}s per scenario`);

    const allResults = [];

    for (const scenario of testScenarios) {
        try {
            const stats = await runScenarioTest(scenario, concurrentConnections, testDuration);
            displayResults(scenario.name, stats);
            allResults.push({ scenario: scenario.name, stats });
        } catch (error) {
            console.error(`${colors.red}❌ Error testing ${scenario.name}: ${error.message}${colors.reset}`);
        }
    }

    // Summary
    console.log(`\n${colors.blue}📈 Overall Summary${colors.reset}`);
    console.log(`${colors.green}✅ All tests completed${colors.reset}`);

    totalRequests = allResults.reduce((sum, result) => sum + result.stats.totalRequests, 0);
    totalSuccessful = allResults.reduce((sum, result) => sum + result.stats.successfulRequests, 0);

    console.log(`Total requests across all scenarios: ${totalRequests}`);
    console.log(`Total successful requests: ${totalSuccessful}`);
    console.log(`Overall success rate: ${((totalSuccessful / totalRequests) * 100).toFixed(2)}%`);

    console.log(`\n${colors.green}🎉 Load testing completed!${colors.reset}`);
}

// Handle uncaught errors
process.on('unhandledRejection', (reason, promise) => {
    console.error(`${colors.red}Unhandled Rejection at: ${promise}${colors.reset}`, reason);
    process.exit(1);
});

process.on('uncaughtException', (error) => {
    console.error(`${colors.red}Uncaught Exception:${colors.reset}`, error);
    process.exit(1);
});

// Run the tests
if (require.main === module) {
    main().catch(console.error);
}

module.exports = { runScenarioTest, StatsCollector };