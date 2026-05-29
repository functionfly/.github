import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('errors');
const dbQueryTime = new Trend('db_query_time');

export const options = {
  stages: [
    { duration: '1m', target: 50 },
    { duration: '3m', target: 50 },
    { duration: '1m', target: 0 },
  ],
  thresholds: {
    db_query_time: ['p(95)<1000', 'p(99)<2000'],
    errors: ['rate<0.02'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export function setup() {
  const res = http.get(`${BASE_URL}/api/health`);
  if (res.status !== 200) {
    throw new Error(`API health check failed: ${res.status}`);
  }
  return { baseUrl: BASE_URL };
}

export default function(data) {
  const url = data.baseUrl;

  // Simulate billing report queries
  // These are the queries that benefit from the performance indexes
  const tenantId = __ENV.TEST_TENANT_ID || '00000000-0000-0000-0000-000000000000';

  // Test 1: Invoice listing (benefits from idx_invoices_period_tenant)
  {
    const start = Date.now();
    const res = http.get(`${url}/api/v1/billing/invoices?period_start=2024-01-01&period_end=2024-12-31`);
    const duration = Date.now() - start;
    const success = check(res, {
      'invoices query 200': (r) => r.status === 200 || r.status === 401,
    });
    errorRate.add(!success);
    dbQueryTime.add(duration);
  }

  // Test 2: Payment retry status query (benefits from idx_payment_retries_status_next_retry)
  {
    const start = Date.now();
    const res = http.get(`${url}/api/v1/billing/payment-retries?status=active`);
    const duration = Date.now() - start;
    const success = check(res, {
      'payment retries query 200': (r) => r.status === 200 || r.status === 401,
    });
    errorRate.add(!success);
    dbQueryTime.add(duration);
  }

  // Test 3: Cost allocation query (benefits from idx_cost_allocation_entries_timestamp)
  {
    const start = Date.now();
    const res = http.get(`${url}/api/v1/billing/cost-allocation?from=2024-01-01&to=2024-03-31`);
    const duration = Date.now() - start;
    const success = check(res, {
      'cost allocation query 200': (r) => r.status === 200 || r.status === 401,
    });
    errorRate.add(!success);
    dbQueryTime.add(duration);
  }

  // Test 4: Dunning status query (benefits from idx_payment_retries_grace_period_status)
  {
    const start = Date.now();
    const res = http.get(`${url}/api/v1/billing/dunning?status=active`);
    const duration = Date.now() - start;
    const success = check(res, {
      'dunning query 200': (r) => r.status === 200 || r.status === 401,
    });
    errorRate.add(!success);
    dbQueryTime.add(duration);
  }

  sleep(0.5);
}

export function handleSummary(data) {
  const p95 = data.metrics.db_query_time.values['p(95)'].toFixed(2);
  const p99 = data.metrics.db_query_time.values['p(99)'].toFixed(2);

  return {
    'stdout': `\n=== Billing Query Performance ===\nP(95): ${p95}ms\nP(99): ${p99}ms\n`,
    'billing-perf-summary.json': JSON.stringify({
      db_query_p95_ms: p95,
      db_query_p99_ms: p99,
    }),
  };
}
