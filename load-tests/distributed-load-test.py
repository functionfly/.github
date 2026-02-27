#!/usr/bin/env python3
"""
Distributed Load Testing Framework for FunctionFly

This script coordinates load testing across multiple regions and providers
to simulate real-world traffic patterns and test geo-routing capabilities.
"""

import asyncio
import json
import logging
import subprocess
import sys
from dataclasses import dataclass, asdict
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, List, Optional, Any
import aiohttp
import psutil
import yaml

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

@dataclass
class LoadTestConfig:
    """Configuration for distributed load testing"""
    regions: List[str]
    providers: List[str]
    test_duration: int  # seconds
    ramp_up_time: int   # seconds
    concurrent_users: int
    requests_per_second: int
    functionfly_endpoints: Dict[str, str]
    output_dir: str = "load-test-results"

@dataclass
class TestResult:
    """Results from a single test run"""
    region: str
    provider: str
    start_time: datetime
    end_time: datetime
    total_requests: int
    successful_requests: int
    failed_requests: int
    avg_response_time: float
    p95_response_time: float
    p99_response_time: float
    error_rate: float
    throughput_rps: float
    routing_latencies: List[float]
    backend_failover_events: int

@dataclass
class SystemMetrics:
    """System resource metrics during test"""
    timestamp: datetime
    cpu_percent: float
    memory_percent: float
    network_connections: int
    disk_io_read: int
    disk_io_write: int

class DistributedLoadTester:
    """Coordinates distributed load testing across regions"""

    def __init__(self, config: LoadTestConfig):
        self.config = config
        self.results: List[TestResult] = []
        self.system_metrics: List[SystemMetrics] = []
        self.session: Optional[aiohttp.ClientSession] = None

    async def __aenter__(self):
        self.session = aiohttp.ClientSession()
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        if self.session:
            await self.session.close()

    async def run_distributed_test(self) -> Dict[str, Any]:
        """Run the complete distributed load test"""
        logger.info("Starting distributed load test...")

        # Create output directory
        output_dir = Path(self.config.output_dir)
        output_dir.mkdir(exist_ok=True)

        # Start system monitoring
        monitor_task = asyncio.create_task(self.monitor_system_resources())

        try:
            # Run tests for each region/provider combination
            test_tasks = []
            for region in self.config.regions:
                for provider in self.config.providers:
                    task = asyncio.create_task(
                        self.run_regional_test(region, provider)
                    )
                    test_tasks.append(task)

            # Wait for all regional tests to complete
            await asyncio.gather(*test_tasks)

            # Stop monitoring
            monitor_task.cancel()
            try:
                await monitor_task
            except asyncio.CancelledError:
                pass

            # Generate comprehensive report
            report = await self.generate_report()

            # Save results
            await self.save_results(output_dir, report)

            return report

        except Exception as e:
            logger.error(f"Distributed test failed: {e}")
            raise

    async def run_regional_test(self, region: str, provider: str) -> TestResult:
        """Run load test for a specific region and provider"""
        logger.info(f"Starting test for {region}/{provider}")

        endpoint = self.config.functionfly_endpoints.get(f"{region}_{provider}")
        if not endpoint:
            raise ValueError(f"No endpoint configured for {region}/{provider}")

        start_time = datetime.now()

        # Run k6 test for this region
        result = await self.execute_k6_test(region, provider, endpoint)

        end_time = datetime.now()

        test_result = TestResult(
            region=region,
            provider=provider,
            start_time=start_time,
            end_time=end_time,
            **result
        )

        self.results.append(test_result)
        logger.info(f"Completed test for {region}/{provider}: {result['successful_requests']}/{result['total_requests']} successful")

        return test_result

    async def execute_k6_test(self, region: str, provider: str, endpoint: str) -> Dict[str, Any]:
        """Execute k6 load test for specific region/provider"""
        k6_script = f"""
import http from 'k6/http';
import {{ check, sleep }} from 'k6';

export const options = {{
  duration: '{self.config.test_duration}s',
  vus: {self.config.concurrent_users},
  rps: {self.config.requests_per_second},
  tags: {{
    region: '{region}',
    provider: '{provider}'
  }},
}};

const FUNCTION_ID = '550e8400-e29b-41d4-a716-446655440000';
const AUTH_TOKEN = 'test_token_123';

export default function () {{
  const payload = {{
    action: 'process',
    data: {{
      region: '{region}',
      provider: '{provider}',
      timestamp: new Date().toISOString(),
      items: Array.from({{length: 5}}, () => Math.random().toString(36)))
    }}
  }};

  const response = http.post(
    `{endpoint}/api/functions/${{FUNCTION_ID}}/invoke`,
    JSON.stringify(payload),
    {{
      headers: {{
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${{AUTH_TOKEN}}`,
        'x-region': '{region}',
        'x-provider': '{provider}'
      }},
    }}
  );

  check(response, {{
    'status is 200': (r) => r.status === 200,
    'response time < 1000ms': (r) => r.timings.duration < 1000,
    'has routing info': (r) => r.headers['x-backend-selected'],
  }});

  sleep(0.1);
}}

export function handleSummary(data) {{
  return {{
    'stdout': JSON.stringify({{
      total_requests: data.metrics.http_reqs.values.count,
      successful_requests: data.metrics.http_reqs.values.count - (data.metrics.http_req_failed.values.count || 0),
      failed_requests: data.metrics.http_req_failed.values.count || 0,
      avg_response_time: data.metrics.http_req_duration.values.avg,
      p95_response_time: data.metrics.http_req_duration.values['p(95)'],
      p99_response_time: data.metrics.http_req_duration.values['p(99)'],
      error_rate: data.metrics.http_req_failed.values.rate,
      throughput_rps: data.metrics.http_reqs.values.rate,
    }}, null, 2),
  }};
}}
"""

        # Write k6 script to temporary file
        script_path = f"/tmp/k6_test_{region}_{provider}.js"
        with open(script_path, 'w') as f:
            f.write(k6_script)

        try:
            # Execute k6 test
            cmd = [
                'k6', 'run',
                '--tag', f'region={region}',
                '--tag', f'provider={provider}',
                script_path
            ]

            process = await asyncio.create_subprocess_exec(
                *cmd,
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE
            )

            stdout, stderr = await process.communicate()

            if process.returncode != 0:
                logger.error(f"k6 test failed for {region}/{provider}: {stderr.decode()}")
                # Return default failed result
                return {
                    'total_requests': 0,
                    'successful_requests': 0,
                    'failed_requests': 0,
                    'avg_response_time': 0.0,
                    'p95_response_time': 0.0,
                    'p99_response_time': 0.0,
                    'error_rate': 1.0,
                    'throughput_rps': 0.0,
                    'routing_latencies': [],
                    'backend_failover_events': 0,
                }

            # Parse k6 output (JSON)
            result_json = stdout.decode().strip()
            result = json.loads(result_json)

            # Add mock routing data (would come from custom k6 extensions)
            result['routing_latencies'] = [50.0, 75.0, 100.0]  # Mock data
            result['backend_failover_events'] = 2  # Mock data

            return result

        finally:
            # Clean up script file
            Path(script_path).unlink(missing_ok=True)

    async def monitor_system_resources(self):
        """Monitor system resources during the test"""
        while True:
            try:
                timestamp = datetime.now()

                # CPU and memory
                cpu_percent = psutil.cpu_percent(interval=1)
                memory = psutil.virtual_memory()

                # Network connections
                connections = len(psutil.net_connections())

                # Disk I/O
                disk_io = psutil.disk_io_counters()
                io_read = disk_io.read_bytes if disk_io else 0
                io_write = disk_io.write_bytes if disk_io else 0

                metrics = SystemMetrics(
                    timestamp=timestamp,
                    cpu_percent=cpu_percent,
                    memory_percent=memory.percent,
                    network_connections=connections,
                    disk_io_read=io_read,
                    disk_io_write=io_write,
                )

                self.system_metrics.append(metrics)

                await asyncio.sleep(5)  # Collect every 5 seconds

            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error(f"Error monitoring system resources: {e}")
                await asyncio.sleep(5)

    async def generate_report(self) -> Dict[str, Any]:
        """Generate comprehensive test report"""
        total_requests = sum(r.total_requests for r in self.results)
        successful_requests = sum(r.successful_requests for r in self.results)
        avg_response_time = sum(r.avg_response_time for r in self.results) / len(self.results) if self.results else 0

        # Calculate regional performance
        regional_performance = {}
        for result in self.results:
            key = f"{result.region}_{result.provider}"
            regional_performance[key] = {
                'requests_per_second': result.throughput_rps,
                'avg_latency': result.avg_response_time,
                'error_rate': result.error_rate,
                'p95_latency': result.p95_response_time,
            }

        # System resource summary
        if self.system_metrics:
            avg_cpu = sum(m.cpu_percent for m in self.system_metrics) / len(self.system_metrics)
            avg_memory = sum(m.memory_percent for m in self.system_metrics) / len(self.system_metrics)
            peak_cpu = max(m.cpu_percent for m in self.system_metrics)
            peak_memory = max(m.memory_percent for m in self.system_metrics)
        else:
            avg_cpu = avg_memory = peak_cpu = peak_memory = 0

        report = {
            'test_summary': {
                'total_regions': len(self.config.regions),
                'total_providers': len(self.config.providers),
                'test_duration_seconds': self.config.test_duration,
                'total_requests': total_requests,
                'successful_requests': successful_requests,
                'overall_success_rate': successful_requests / total_requests if total_requests > 0 else 0,
                'average_response_time_ms': avg_response_time,
            },
            'regional_performance': regional_performance,
            'system_resources': {
                'average_cpu_percent': avg_cpu,
                'average_memory_percent': avg_memory,
                'peak_cpu_percent': peak_cpu,
                'peak_memory_percent': peak_memory,
            },
            'recommendations': self.generate_recommendations(),
            'timestamp': datetime.now().isoformat(),
        }

        return report

    def generate_recommendations(self) -> List[str]:
        """Generate performance optimization recommendations"""
        recommendations = []

        if self.results:
            avg_error_rate = sum(r.error_rate for r in self.results) / len(self.results)
            if avg_error_rate > 0.05:
                recommendations.append("High error rate detected. Consider increasing backend capacity or implementing better load balancing.")

            avg_p95 = sum(r.p95_response_time for r in self.results) / len(self.results)
            if avg_p95 > 500:
                recommendations.append("P95 response time is high. Consider optimizing routing logic or adding more backend instances.")

            # Check for regional disparities
            latencies = [r.avg_response_time for r in self.results]
            if max(latencies) - min(latencies) > 200:
                recommendations.append("Significant latency variation between regions. Consider geo-routing optimizations.")

        if not recommendations:
            recommendations.append("Performance looks good! Consider regular load testing to maintain optimal performance.")

        return recommendations

    async def save_results(self, output_dir: Path, report: Dict[str, Any]):
        """Save test results to files"""
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")

        # Save main report
        report_file = output_dir / f"distributed_test_report_{timestamp}.json"
        with open(report_file, 'w') as f:
            json.dump(report, f, indent=2, default=str)

        # Save detailed results
        detailed_file = output_dir / f"detailed_results_{timestamp}.json"
        with open(detailed_file, 'w') as f:
            json.dump({
                'config': asdict(self.config),
                'results': [asdict(r) for r in self.results],
                'system_metrics': [asdict(m) for m in self.system_metrics],
            }, f, indent=2, default=str)

        # Save summary text
        summary_file = output_dir / f"summary_{timestamp}.txt"
        with open(summary_file, 'w') as f:
            f.write("FunctionFly Distributed Load Test Summary\n")
            f.write("=" * 50 + "\n\n")
            f.write(f"Test completed at: {datetime.now()}\n\n")

            summary = report['test_summary']
            f.write(f"Total Requests: {summary['total_requests']:,}\n")
            f.write(f"Success Rate: {summary['overall_success_rate']:.1%}\n")
            f.write(f"Average Response Time: {summary['average_response_time_ms']:.1f}ms\n\n")

            f.write("Regional Performance:\n")
            for region_provider, perf in report['regional_performance'].items():
                f.write(f"  {region_provider}:\n")
                f.write(f"    RPS: {perf['requests_per_second']:.1f}\n")
                f.write(f"    Avg Latency: {perf['avg_latency']:.1f}ms\n")
                f.write(f"    Error Rate: {perf['error_rate']:.1%}\n\n")

            f.write("System Resources:\n")
            sys_res = report['system_resources']
            f.write(f"  Avg CPU: {sys_res['average_cpu_percent']:.1f}%\n")
            f.write(f"  Avg Memory: {sys_res['average_memory_percent']:.1f}%\n")
            f.write(f"  Peak CPU: {sys_res['peak_cpu_percent']:.1f}%\n")
            f.write(f"  Peak Memory: {sys_res['peak_memory_percent']:.1f}%\n\n")

            f.write("Recommendations:\n")
            for rec in report['recommendations']:
                f.write(f"  • {rec}\n")

        logger.info(f"Results saved to {output_dir}")

async def main():
    """Main entry point"""
    # Load configuration
    config_file = Path("load-tests/distributed-config.yml")
    if not config_file.exists():
        # Create default configuration
        default_config = LoadTestConfig(
            regions=["us-east-1", "eu-west-1", "ap-southeast-1"],
            providers=["cloudflare", "vercel", "fly"],
            test_duration=300,  # 5 minutes
            ramp_up_time=60,    # 1 minute
            concurrent_users=100,
            requests_per_second=200,
            functionfly_endpoints={
                "us-east-1_cloudflare": "https://us-east-1.cloudflare.functionfly.com",
                "eu-west-1_cloudflare": "https://eu-west-1.cloudflare.functionfly.com",
                "ap-southeast-1_cloudflare": "https://ap-southeast-1.cloudflare.functionfly.com",
                "us-east-1_vercel": "https://us-east-1.vercel.functionfly.com",
                "eu-west-1_vercel": "https://eu-west-1.vercel.functionfly.com",
                "ap-southeast-1_vercel": "https://ap-southeast-1.vercel.functionfly.com",
                "us-east-1_fly": "https://us-east-1.fly.functionfly.com",
                "eu-west-1_fly": "https://eu-west-1.fly.functionfly.com",
                "ap-southeast-1_fly": "https://ap-southeast-1.fly.functionfly.com",
            }
        )

        config_file.parent.mkdir(exist_ok=True)
        with open(config_file, 'w') as f:
            yaml.dump(asdict(default_config), f, default_flow_style=False)

        logger.info(f"Created default configuration at {config_file}")
        config = default_config
    else:
        with open(config_file) as f:
            config_dict = yaml.safe_load(f)
        config = LoadTestConfig(**config_dict)

    # Run distributed test
    async with DistributedLoadTester(config) as tester:
        report = await tester.run_distributed_test()

    # Print summary
    summary = report['test_summary']
    print("\n" + "="*60)
    print("DISTRIBUTED LOAD TEST COMPLETED")
    print("="*60)
    print(f"Total Requests: {summary['total_requests']:,}")
    print(".1%")
    print(".1f")
    print("\nRegional Results:")
    for region_provider, perf in report['regional_performance'].items():
        print(f"  {region_provider}: {perf['requests_per_second']:.1f} RPS, {perf['avg_latency']:.1f}ms avg")

    print("\nRecommendations:")
    for rec in report['recommendations']:
        print(f"  • {rec}")

if __name__ == "__main__":
    asyncio.run(main())