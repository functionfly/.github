import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('errors');
const responseTime = new Trend('response_time');

export const options = {
  stages: [
    { duration: '2m', target: 100 },  // Ramp up to 100 users
    { duration: '5m', target: 100 },  // Stay at 100 users
    { duration: '2m', target: 200 },  // Ramp up to 200 users
    { duration: '5m', target: 200 },  // Stay at 200 users
    { duration: '2m', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],  // 95% under 500ms, 99% under 1s
    http_req_failed: ['rate<0.05'],  // Error rate under 5%
    errors: ['rate<0.05'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export function setup() {
  // Verify API is running
  const res = http.get(`${BASE_URL}/api/health`);
  if (res.status !== 200) {
    throw new Error(`API health check failed: ${res.status}`);
  }
  return { baseUrl: BASE_URL };
}

export default function(data) {
  const url = data.baseUrl;

  // Test 1: Health endpoint
  {
    const res = http.get(`${url}/api/health`);
    const success = check(res, {
      'health status 200': (r) => r.status === 200,
    });
    errorRate.add(!success);
    responseTime.add(res.timings.duration);
  }

  // Test 2: List functions (public registry)
  {
    const res = http.get(`${url}/api/v1/functions?limit=20`);
    const success = check(res, {
      'functions list 200': (r) => r.status === 200,
      'functions returned': (r) => JSON.parse(r.body).data !== undefined,
    });
    errorRate.add(!success);
    responseTime.add(res.timings.duration);
  }

  // Test 3: Search functions
  {
    const res = http.get(`${url}/api/v1/functions/search?q=test&limit=10`);
    const success = check(res, {
      'search 200': (r) => r.status === 200,
    });
    errorRate.add(!success);
    responseTime.add(res.timings.duration);
  }

  // Test 4: Tenant user list (requires auth - may fail, that's ok)
  {
    const res = http.get(`${url}/api/v1/tenants`);
    const success = check(res, {
      'tenant endpoint accessible': (r) => r.status === 200 || r.status === 401,
    });
    errorRate.add(!success && res.status !== 401);
    responseTime.add(res.timings.duration);
  }

  sleep(1);
}

export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'performance-summary.json': JSON.stringify({
      http_req_duration_p95: data.metrics.http_req_duration.values['p(95)'],
      http_req_duration_p99: data.metrics.http_req_duration.values['p(99)'],
      http_req_failed_rate: data.metrics.http_req_failed.values.rate,
      errors_rate: data.metrics.errors.values.rate,
      total_requests: data.metrics.http_reqs.values.count,
      test_duration: data.state.testDuration,
    }),
  };
}

function textSummary(data, opts) {
  const indent = opts.indent || '';
  let summary = '\n' + indent + '='.repeat(60) + '\n';
  summary += indent + '  PERFORMANCE TEST SUMMARY\n';
  summary += indent + '='.repeat(60) + '\n\n';

  const duration = (data.state.testDuration / 1000).toFixed(1);
  const totalReqs = data.metrics.http_reqs.values.count;
  const failedRate = (data.metrics.http_req_failed.values.rate * 100).toFixed(2);
  const p95 = data.metrics.http_req_duration.values['p(95)'].toFixed(2);
  const p99 = data.metrics.http_req_duration.values['p(99)'].toFixed(2);
  const avg = data.metrics.http_req_duration.values.avg.toFixed(2);

  summary += `${indent}Duration: ${duration}s\n`;
  summary += `${indent}Total Requests: ${totalReqs}\n`;
  summary += `${indent}Failed Rate: ${failedRate}%\n`;
  summary += `${indent}Response Times:\n`;
  summary += `${indent}  Avg: ${avg}ms\n`;
  summary += `${indent}  p(95): ${p95}ms\n`;
  summary += `${indent}  p(99): ${p99}ms\n`;
  summary += '\n' + indent + '='.repeat(60) + '\n';

  return summary;
}
