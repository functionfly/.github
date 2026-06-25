import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Gauge, Counter } from 'k6/metrics';

const coldStartTimes = {
  orchestrator: new Trend('cold_start_orchestrator_ms'),
  aiService: new Trend('cold_start_ai_service_ms'),
  dashboard: new Trend('cold_start_dashboard_ms'),
  runtimeLocal: new Trend('cold_start_runtime_local_ms'),
  runtimePrism: new Trend('cold_start_runtime_prism_ms'),
  runtimeKotlin: new Trend('cold_start_runtime_kotlin_ms'),
};

const readinessGauges = {
  orchestrator: new Gauge('ready_orchestrator'),
  aiService: new Gauge('ready_ai_service'),
  dashboard: new Gauge('ready_dashboard'),
  runtimeLocal: new Gauge('ready_runtime_local'),
  runtimePrism: new Gauge('ready_runtime_prism'),
  runtimeKotlin: new Gauge('ready_runtime_kotlin'),
};

const startupAttempts = {
  orchestrator: new Counter('startup_attempts_orchestrator'),
  aiService: new Counter('startup_attempts_ai_service'),
  runtimeLocal: new Counter('startup_attempts_runtime_local'),
  runtimePrism: new Counter('startup_attempts_runtime_prism'),
  runtimeKotlin: new Counter('startup_attempts_runtime_kotlin'),
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const AI_SERVICE_URL = __ENV.AI_SERVICE_URL || 'http://localhost:18081';
const DASHBOARD_URL = __ENV.DASHBOARD_URL || 'http://localhost:3000';
const RUNTIME_LOCAL_URL = __ENV.RUNTIME_LOCAL_URL || 'http://localhost:8083';
const RUNTIME_PRISM_URL = __ENV.RUNTIME_PRISM_URL || 'http://localhost:8084';
const RUNTIME_KOTLIN_URL = __ENV.RUNTIME_KOTLIN_URL || 'http://localhost:8085';

const POLL_INTERVAL_MS = parseInt(__ENV.POLL_INTERVAL_MS || '100');
const MAX_WAIT_MS = parseInt(__ENV.MAX_WAIT_MS || '60000');
const ITERATIONS = parseInt(__ENV.ITERATIONS || '5');

function pollUntilReady(url, name, timeout = MAX_WAIT_MS) {
  const startTime = Date.now();
  startupAttempts[name].add(1);

  while (Date.now() - startTime < timeout) {
    try {
      const response = http.get(url, { timeout: '5s', tags: { name: name } });
      
      if (response.status === 200) {
        const elapsed = Date.now() - startTime;
        coldStartTimes[name].add(elapsed);
        readinessGauges[name].add(1);
        return { ready: true, elapsed, status: response.status };
      }
    } catch (e) {
      // Not ready yet
    }
    sleep(POLL_INTERVAL_MS / 1000);
  }
  
  readinessGauges[name].add(0);
  return { ready: false, elapsed: timeout, error: 'timeout' };
}

export const options = {
  scenarios: {
    cold_start_test: {
      executor: 'per-vu-iterations',
      vus: 1,
      iterations: ITERATIONS,
      maxDuration: '30m',
    },
  },
  thresholds: {
    'cold_start_orchestrator_ms': ['p(95)<30000', 'p(99)<60000'],
    'cold_start_ai_service_ms': ['p(95)<30000', 'p(99)<60000'],
    'cold_start_runtime_local_ms': ['p(95)<10000', 'p(99)<30000'],
    'cold_start_runtime_prism_ms': ['p(95)<10000', 'p(99)<30000'],
    'cold_start_runtime_kotlin_ms': ['p(95)<30000', 'p(99)<60000'],
  },
};

export default function () {
  const service = __ENV.SERVICE || 'all';
  
  switch (service) {
    case 'orchestrator':
      pollUntilReady(`${BASE_URL}/health`, 'orchestrator');
      break;
    case 'ai_service':
      pollUntilReady(`${AI_SERVICE_URL}/health`, 'aiService');
      break;
    case 'dashboard':
      pollUntilReady(`${DASHBOARD_URL}/`, 'dashboard');
      break;
    case 'runtime_local':
      pollUntilReady(`${RUNTIME_LOCAL_URL}/health`, 'runtimeLocal');
      break;
    case 'runtime_prism':
      pollUntilReady(`${RUNTIME_PRISM_URL}/health`, 'runtimePrism');
      break;
    case 'runtime_kotlin':
      pollUntilReady(`${RUNTIME_KOTLIN_URL}/health`, 'runtimeKotlin');
      break;
    case 'all':
    default:
      const results = {
        orchestrator: pollUntilReady(`${BASE_URL}/health`, 'orchestrator'),
        aiService: pollUntilReady(`${AI_SERVICE_URL}/health`, 'aiService'),
        runtimeLocal: pollUntilReady(`${RUNTIME_LOCAL_URL}/health`, 'runtimeLocal'),
        runtimePrism: pollUntilReady(`${RUNTIME_PRISM_URL}/health`, 'runtimePrism'),
        runtimeKotlin: pollUntilReady(`${RUNTIME_KOTLIN_URL}/health`, 'runtimeKotlin'),
      };
      console.log(`Cold start results: ${JSON.stringify(results)}`);
      break;
  }
}

export function handleSummary(data) {
  const results = {};
  
  for (const [name, metric] of Object.entries(data.metrics)) {
    if (name.startsWith('cold_start_') && metric.values) {
      const service = name.replace('cold_start_', '').replace('_ms', '');
      results[service] = {
        avg_ms: Math.round(metric.values.avg || 0),
        p95_ms: Math.round(metric.values['p(95)'] || 0),
        p99_ms: Math.round(metric.values['p(99)'] || 0),
        min_ms: Math.round(metric.values.min || 0),
        max_ms: Math.round(metric.values.max || 0),
        samples: metric.values.count,
      };
    }
  }

  return {
    'stdout': `
Cold Start Performance Test Results
====================================

Services Tested:
${Object.entries(results).map(([svc, r]) => 
  `  ${svc}:
    Average: ${r.avg_ms}ms
    P95: ${r.p95_ms}ms
    P99: ${r.p99_ms}ms
    Min: ${r.min_ms}ms
    Max: ${r.max_ms}ms
    Samples: ${r.samples}`
).join('\n')}

Test Configuration:
  Poll Interval: ${POLL_INTERVAL_MS}ms
  Max Wait: ${MAX_WAIT_MS}ms
  Iterations: ${ITERATIONS}
`,
    'cold-start-results.json': JSON.stringify(results, null, 2),
    'detailed-metrics.json': JSON.stringify(data, null, 2),
  };
}