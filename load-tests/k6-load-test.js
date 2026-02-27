import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const routingLatency = new Trend('routing_latency');
const backendSelectionTime = new Trend('backend_selection_time');
const cacheHitRate = new Rate('cache_hit_rate');
const circuitBreakerTrips = new Rate('circuit_breaker_trips');

// Test configuration
export const options = {
  scenarios: {
    // Steady load test
    steady_load: {
      executor: 'constant-arrival-rate',
      rate: 100, // requests per second
      timeUnit: '1s',
      duration: '5m',
      preAllocatedVUs: 50,
      maxVUs: 200,
    },

    // Spike test
    spike_test: {
      executor: 'ramping-arrival-rate',
      startRate: 10,
      stages: [
        { duration: '30s', target: 10 },   // Warm up
        { duration: '30s', target: 100 },  // Normal load
        { duration: '30s', target: 1000 }, // Spike
        { duration: '30s', target: 100 },  // Recovery
        { duration: '30s', target: 10 },   // Cool down
      ],
      preAllocatedVUs: 100,
      maxVUs: 1000,
    },

    // Stress test
    stress_test: {
      executor: 'constant-arrival-rate',
      rate: 500,
      timeUnit: '1s',
      duration: '2m',
      preAllocatedVUs: 200,
      maxVUs: 500,
    },

    // Routing performance test
    routing_performance: {
      executor: 'per-vu-iterations',
      vus: 10,
      iterations: 1000,
      maxDuration: '10m',
    },
  },

  thresholds: {
    // Overall request performance
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    http_req_failed: ['rate<0.05'], // Less than 5% failure rate

    // Custom routing metrics
    routing_latency: ['p(95)<100', 'p(99)<200'],
    backend_selection_time: ['p(95)<50'],

    // Cache performance
    cache_hit_rate: ['rate>0.8'], // At least 80% cache hit rate

    // Circuit breaker
    circuit_breaker_trips: ['rate<0.01'], // Less than 1% circuit breaker trips
  },
};

// Test data
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const TEST_USERS = [
  { email: 'user1@example.com', token: 'user_token_1' },
  { email: 'user2@example.com', token: 'user_token_2' },
  { email: 'user3@example.com', token: 'user_token_3' },
];

const FUNCTIONS = [
  '550e8400-e29b-41d4-a716-446655440000',
  '550e8400-e29b-41d4-a716-446655440001',
  '550e8400-e29b-41d4-a716-446655440002',
];

// Helper functions
function getRandomUser() {
  return TEST_USERS[Math.floor(Math.random() * TEST_USERS.length)];
}

function getRandomFunction() {
  return FUNCTIONS[Math.floor(Math.random() * FUNCTIONS.length)];
}

function generatePayload() {
  return {
    action: 'process',
    data: {
      id: Math.floor(Math.random() * 10000),
      items: Array.from({length: Math.floor(Math.random() * 10) + 1}, () =>
        Math.random().toString(36).substring(7)
      ),
      metadata: {
        userId: Math.floor(Math.random() * 1000),
        timestamp: new Date().toISOString(),
        source: 'k6_load_test',
      },
    },
  };
}

// Main test scenarios
export default function () {
  const user = getRandomUser();
  const functionId = getRandomFunction();

  // Scenario 1: Function execution (most common)
  const startTime = new Date().getTime();
  const response = http.post(
    `${BASE_URL}/api/functions/${functionId}/invoke`,
    JSON.stringify(generatePayload()),
    {
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${user.token}`,
        'x-request-id': `k6-${__VU}-${__ITER}`,
      },
    }
  );

  const endTime = new Date().getTime();
  const totalLatency = endTime - startTime;

  // Check response
  const result = check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
    'has valid response': (r) => r.json() && r.json().message,
    'cache hit': (r) => r.headers['x-cache-status'] === 'HIT',
  });

  // Record custom metrics
  if (response.headers['x-routing-latency']) {
    routingLatency.add(parseInt(response.headers['x-routing-latency']));
  }

  if (response.headers['x-backend-selection-time']) {
    backendSelectionTime.add(parseInt(response.headers['x-backend-selection-time']));
  }

  if (response.headers['x-cache-status']) {
    cacheHitRate.add(response.headers['x-cache-status'] === 'HIT' ? 1 : 0);
  }

  if (response.headers['x-circuit-breaker-tripped'] === 'true') {
    circuitBreakerTrips.add(1);
  }

  // Scenario 2: Health check (periodic)
  if (__ITER % 10 === 0) {
    const healthResponse = http.get(`${BASE_URL}/health`);
    check(healthResponse, {
      'health check passes': (r) => r.status === 200,
    });
  }

  // Scenario 3: Error simulation (rare)
  if (__ITER % 100 === 0) {
    const errorResponse = http.get(`${BASE_URL}/api/functions/nonexistent/invoke`);
    check(errorResponse, {
      'error response correct': (r) => r.status === 404,
    });
  }

  // Think time between requests
  sleep(Math.random() * 0.1 + 0.05); // 50-150ms random think time
}

// Setup function - run once before the test
export function setup() {
  console.log('Starting FunctionFly load test setup...');

  // Warm up the system
  const warmupResponse = http.get(`${BASE_URL}/health`);
  if (warmupResponse.status !== 200) {
    console.error('Warmup failed - system may not be ready');
  }

  // Pre-populate some data if needed
  const setupData = {
    baseUrl: BASE_URL,
    testUsers: TEST_USERS.length,
    functions: FUNCTIONS.length,
  };

  console.log(`Load test setup complete: ${JSON.stringify(setupData)}`);
  return setupData;
}

// Teardown function - run once after the test
export function teardown(data) {
  console.log('Load test teardown...');
  console.log(`Test completed with data: ${JSON.stringify(data)}`);
}

// Handle summary - custom reporting
export function handleSummary(data) {
  const summary = {
    'total-requests': data.metrics.http_reqs.values.count,
    'avg-response-time': data.metrics.http_req_duration.values.avg,
    '95th-percentile': data.metrics.http_req_duration.values['p(95)'],
    '99th-percentile': data.metrics.http_req_duration.values['p(99)'],
    'error-rate': data.metrics.http_req_failed.values.rate,
    'routing-latency-avg': data.metrics.routing_latency?.values.avg || 'N/A',
    'cache-hit-rate': data.metrics.cache_hit_rate?.values.rate || 'N/A',
    'circuit-breaker-trips': data.metrics.circuit_breaker_trips?.values.rate || 'N/A',
  };

  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'performance-summary.json': JSON.stringify(summary, null, 2),
    'detailed-results.json': JSON.stringify(data, null, 2),
  };
}

function textSummary(data, options) {
  return `
FunctionFly Load Test Results
============================

Test Duration: ${data.metrics.iteration_duration.values.avg}ms avg iteration
Total Requests: ${data.metrics.http_reqs.values.count}
Failed Requests: ${data.metrics.http_req_failed.values.rate * 100}%

Response Times:
  Average: ${Math.round(data.metrics.http_req_duration.values.avg)}ms
  95th percentile: ${Math.round(data.metrics.http_req_duration.values['p(95)'])}ms
  99th percentile: ${Math.round(data.metrics.http_req_duration.values['p(99)'])}ms

Routing Performance:
  Average routing latency: ${Math.round(data.metrics.routing_latency?.values.avg || 0)}ms
  Cache hit rate: ${Math.round((data.metrics.cache_hit_rate?.values.rate || 0) * 100)}%
  Circuit breaker trips: ${Math.round((data.metrics.circuit_breaker_trips?.values.rate || 0) * 100)}%

Threshold Results:
${Object.entries(data.metrics)
  .filter(([name]) => name.includes('threshold'))
  .map(([name, metric]) => `  ${name}: ${metric.values.passes}/${metric.values.count}`)
  .join('\n') || 'No thresholds defined'}
`;
}