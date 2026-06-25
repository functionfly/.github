import { test, expect } from '@playwright/test';

const COLD_START_TIMEOUT = 60000;
const POLL_INTERVAL_MS = 100;

interface ColdStartResult {
  service: string;
  timeMs: number;
  success: boolean;
}

async function measureServiceColdStart(
  page: any,
  serviceName: string,
  url: string,
  maxWaitMs: number = COLD_START_TIMEOUT
): Promise<ColdStartResult> {
  const startTime = Date.now();
  
  while (Date.now() - startTime < maxWaitMs) {
    try {
      const response = await page.goto(url, { 
        waitUntil: 'domcontentloaded',
        timeout: 5000 
      });
      
      if (response && response.ok()) {
        return {
          service: serviceName,
          timeMs: Date.now() - startTime,
          success: true,
        };
      }
    } catch (e) {
      // Service not ready yet, continue polling
    }
    
    await page.waitForTimeout(POLL_INTERVAL_MS);
  }
  
  return {
    service: serviceName,
    timeMs: Date.now() - startTime,
    success: false,
  };
}

test.describe('Cold Start Performance Tests', () => {
  test.setTimeout(300000);
  
  test('Orchestrator API cold start', async ({ page }) => {
    const result = await measureServiceColdStart(
      page,
      'orchestrator',
      'http://localhost:8080/health'
    );
    
    console.log(`Orchestrator cold start: ${result.timeMs}ms (success: ${result.success})`);
    
    expect(result.success).toBe(true);
    expect(result.timeMs).toBeLessThan(COLD_START_TIMEOUT);
  });
  
  test('AI Service cold start', async ({ page }) => {
    const result = await measureServiceColdStart(
      page,
      'ai-service',
      'http://localhost:18081/health'
    );
    
    console.log(`AI Service cold start: ${result.timeMs}ms (success: ${result.success})`);
    
    expect(result.success).toBe(true);
    expect(result.timeMs).toBeLessThan(COLD_START_TIMEOUT);
  });
  
  test('Dashboard cold start (first paint)', async ({ page }) => {
    const startTime = Date.now();
    
    try {
      await page.goto('http://localhost:3000', { 
        waitUntil: 'domcontentloaded',
        timeout: COLD_START_TIMEOUT 
      });
      
      const timeToFirstPaint = Date.now() - startTime;
      console.log(`Dashboard cold start (domcontentloaded): ${timeToFirstPaint}ms`);
      
      expect(timeToFirstPaint).toBeLessThan(COLD_START_TIMEOUT);
    } catch (e) {
      console.log(`Dashboard failed to load: ${e}`);
      throw e;
    }
  });
  
  test('Dashboard cold start (full load)', async ({ page }) => {
    const startTime = Date.now();
    
    try {
      await page.goto('http://localhost:3000', { 
        waitUntil: 'networkidle',
        timeout: COLD_START_TIMEOUT 
      });
      
      const timeToFullLoad = Date.now() - startTime;
      console.log(`Dashboard cold start (networkidle): ${timeToFullLoad}ms`);
      
      expect(timeToFullLoad).toBeLessThan(COLD_START_TIMEOUT);
    } catch (e) {
      console.log(`Dashboard failed to load: ${e}`);
      throw e;
    }
  });
  
  test('All services cold start comparison', async ({ page }) => {
    const results: ColdStartResult[] = [];
    
    const services = [
      { name: 'orchestrator', url: 'http://localhost:8080/health' },
      { name: 'ai-service', url: 'http://localhost:18081/health' },
    ];
    
    for (const service of services) {
      const result = await measureServiceColdStart(
        page,
        service.name,
        service.url
      );
      results.push(result);
    }
    
    console.log('\n=== Cold Start Results ===');
    results.forEach(r => {
      console.log(`${r.service}: ${r.timeMs}ms (${r.success ? 'OK' : 'FAILED'})`);
    });
    
    const csv = [
      'service,time_ms,success',
      ...results.map(r => `${r.service},${r.timeMs},${r.success}`),
    ].join('\n');
    
    console.log('\n=== CSV Output ===');
    console.log(csv);
  });
});

test.describe('Warm Start Performance Tests', () => {
  test('Orchestrator API warm request', async ({ page }) => {
    // First, ensure service is warm
    await page.goto('http://localhost:8080/health');
    
    const startTime = Date.now();
    await page.goto('http://localhost:8080/health');
    const warmTime = Date.now() - startTime;
    
    console.log(`Orchestrator warm request: ${warmTime}ms`);
    
    expect(warmTime).toBeLessThan(1000);
  });
});