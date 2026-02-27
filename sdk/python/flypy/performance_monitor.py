"""
Performance monitoring and profiling tools for FlyPy functions.

This module provides comprehensive performance monitoring, profiling, metrics collection,
alerting, and reporting capabilities for FlyPy functions.
"""

import time
import threading
import statistics
from collections import defaultdict, deque
from typing import Dict, Any, List, Optional, Callable, Union
from datetime import datetime, timedelta
import json
import logging
import psutil
import os
from functools import wraps
import asyncio
from concurrent.futures import ThreadPoolExecutor


class PerformanceMetrics:
    """Collects and analyzes performance metrics."""

    def __init__(self, max_samples: int = 1000):
        self.max_samples = max_samples
        self.execution_times: Dict[str, deque] = defaultdict(lambda: deque(maxlen=max_samples))
        self.memory_usage: Dict[str, deque] = defaultdict(lambda: deque(maxlen=max_samples))
        self.error_counts: Dict[str, int] = defaultdict(int)
        self.call_counts: Dict[str, int] = defaultdict(int)
        self.last_execution: Dict[str, float] = {}
        self.function_metadata: Dict[str, Dict[str, Any]] = {}

    def record_execution(
        self,
        function_name: str,
        execution_time: float,
        memory_used: Optional[float] = None,
        success: bool = True,
        custom_metrics: Optional[Dict[str, Any]] = None
    ):
        """Record execution metrics for a function."""
        self.execution_times[function_name].append(execution_time)
        self.call_counts[function_name] += 1
        self.last_execution[function_name] = time.time()

        if memory_used is not None:
            self.memory_usage[function_name].append(memory_used)

        if not success:
            self.error_counts[function_name] += 1

        # Store custom metrics
        if custom_metrics:
            if function_name not in self.function_metadata:
                self.function_metadata[function_name] = {}
            self.function_metadata[function_name].setdefault('custom_metrics', []).append({
                'timestamp': time.time(),
                'metrics': custom_metrics
            })

    def get_function_stats(self, function_name: str) -> Dict[str, Any]:
        """Get performance statistics for a specific function."""
        times = list(self.execution_times[function_name])

        if not times:
            return {
                'function_name': function_name,
                'total_calls': 0,
                'error_count': self.error_counts[function_name],
                'last_execution': self.last_execution.get(function_name)
            }

        memory = list(self.memory_usage[function_name]) if self.memory_usage[function_name] else []

        stats = {
            'function_name': function_name,
            'total_calls': self.call_counts[function_name],
            'error_count': self.error_counts[function_name],
            'error_rate': self.error_counts[function_name] / self.call_counts[function_name],
            'last_execution': self.last_execution.get(function_name),
            'execution_time': {
                'min': min(times),
                'max': max(times),
                'mean': statistics.mean(times),
                'median': statistics.median(times),
                'p95': statistics.quantiles(times, n=20)[18] if len(times) >= 20 else max(times),
                'p99': statistics.quantiles(times, n=100)[98] if len(times) >= 100 else max(times),
                'std_dev': statistics.stdev(times) if len(times) > 1 else 0
            }
        }

        if memory:
            stats['memory_usage'] = {
                'min': min(memory),
                'max': max(memory),
                'mean': statistics.mean(memory),
                'median': statistics.median(memory)
            }

        return stats

    def get_all_stats(self) -> Dict[str, Any]:
        """Get performance statistics for all functions."""
        all_stats = {}
        for func_name in set(list(self.execution_times.keys()) + list(self.call_counts.keys())):
            all_stats[func_name] = self.get_function_stats(func_name)

        return {
            'functions': all_stats,
            'summary': {
                'total_functions': len(all_stats),
                'total_calls': sum(stats['total_calls'] for stats in all_stats.values()),
                'total_errors': sum(stats['error_count'] for stats in all_stats.values()),
                'avg_error_rate': statistics.mean([stats['error_rate'] for stats in all_stats.values() if 'error_rate' in stats])
            }
        }

    def clear_metrics(self, function_name: Optional[str] = None):
        """Clear metrics for a function or all functions."""
        if function_name:
            if function_name in self.execution_times:
                self.execution_times[function_name].clear()
            if function_name in self.memory_usage:
                self.memory_usage[function_name].clear()
            self.error_counts[function_name] = 0
            self.call_counts[function_name] = 0
            if function_name in self.last_execution:
                del self.last_execution[function_name]
        else:
            self.execution_times.clear()
            self.memory_usage.clear()
            self.error_counts.clear()
            self.call_counts.clear()
            self.last_execution.clear()


class FunctionProfiler:
    """Profiles function execution with detailed timing."""

    def __init__(self):
        self.active_profiles: Dict[str, Dict[str, Any]] = {}
        self.completed_profiles: Dict[str, List[Dict[str, Any]]] = defaultdict(list)

    def start_profiling(self, function_name: str, call_id: Optional[str] = None) -> str:
        """Start profiling a function execution."""
        profile_id = call_id or f"{function_name}_{int(time.time() * 1000000)}"

        self.active_profiles[profile_id] = {
            'function_name': function_name,
            'start_time': time.perf_counter(),
            'start_memory': self._get_memory_usage(),
            'checkpoints': [],
            'call_stack': []
        }

        return profile_id

    def add_checkpoint(self, profile_id: str, checkpoint_name: str, metadata: Optional[Dict[str, Any]] = None):
        """Add a profiling checkpoint."""
        if profile_id not in self.active_profiles:
            return

        profile = self.active_profiles[profile_id]
        checkpoint_time = time.perf_counter() - profile['start_time']

        checkpoint = {
            'name': checkpoint_name,
            'time': checkpoint_time,
            'memory': self._get_memory_usage(),
            'metadata': metadata or {}
        }

        profile['checkpoints'].append(checkpoint)

    def end_profiling(self, profile_id: str, result_metadata: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """End profiling and return profile data."""
        if profile_id not in self.active_profiles:
            return {}

        profile = self.active_profiles[profile_id]
        end_time = time.perf_counter()
        end_memory = self._get_memory_usage()

        profile_data = {
            'profile_id': profile_id,
            'function_name': profile['function_name'],
            'total_time': end_time - profile['start_time'],
            'memory_delta': end_memory - profile['start_memory'],
            'start_memory': profile['start_memory'],
            'end_memory': end_memory,
            'checkpoints': profile['checkpoints'],
            'result_metadata': result_metadata or {},
            'timestamp': time.time()
        }

        # Store completed profile
        func_name = profile['function_name']
        self.completed_profiles[func_name].append(profile_data)

        # Clean up active profile
        del self.active_profiles[profile_id]

        return profile_data

    def get_function_profiles(self, function_name: str, limit: int = 10) -> List[Dict[str, Any]]:
        """Get recent profiles for a function."""
        profiles = self.completed_profiles[function_name]
        return sorted(profiles[-limit:], key=lambda x: x['timestamp'], reverse=True)

    def _get_memory_usage(self) -> float:
        """Get current memory usage in MB."""
        try:
            process = psutil.Process(os.getpid())
            return process.memory_info().rss / 1024 / 1024  # MB
        except Exception:
            return 0.0


class PerformanceMonitor:
    """Main performance monitoring system."""

    def __init__(self):
        self.metrics = PerformanceMetrics()
        self.profiler = FunctionProfiler()
        self.alerts: List[Dict[str, Any]] = []
        self.alert_handlers: List[Callable] = []
        self.monitoring_enabled = True
        self._monitor_thread: Optional[threading.Thread] = None
        self._stop_monitoring = False

    def monitor_function(self, enable_profiling: bool = False):
        """Decorator to monitor function performance."""
        def decorator(func: Callable) -> Callable:
            @wraps(func)
            def wrapper(*args, **kwargs):
                if not self.monitoring_enabled:
                    return func(*args, **kwargs)

                func_name = getattr(func, '_flypy_metadata', {}).get('name', func.__name__)
                start_time = time.perf_counter()
                start_memory = self._get_memory_usage()

                profile_id = None
                if enable_profiling:
                    profile_id = self.profiler.start_profiling(func_name)

                try:
                    # Execute function
                    if profile_id:
                        self.profiler.add_checkpoint(profile_id, "function_start")

                    result = func(*args, **kwargs)

                    if profile_id:
                        self.profiler.add_checkpoint(profile_id, "function_end")

                    # Record successful execution
                    execution_time = time.perf_counter() - start_time
                    memory_used = self._get_memory_usage() - start_memory

                    self.metrics.record_execution(
                        func_name,
                        execution_time,
                        memory_used,
                        success=True
                    )

                    # End profiling
                    if profile_id:
                        self.profiler.end_profiling(profile_id, {'success': True})

                    return result

                except Exception as e:
                    # Record failed execution
                    execution_time = time.perf_counter() - start_time
                    memory_used = self._get_memory_usage() - start_memory

                    self.metrics.record_execution(
                        func_name,
                        execution_time,
                        memory_used,
                        success=False
                    )

                    # End profiling
                    if profile_id:
                        self.profiler.end_profiling(profile_id, {'success': False, 'error': str(e)})

                    raise e

            return wrapper
        return decorator

    def start_background_monitoring(self, interval_seconds: int = 60):
        """Start background monitoring thread."""
        if self._monitor_thread and self._monitor_thread.is_alive():
            return

        self._stop_monitoring = False
        self._monitor_thread = threading.Thread(
            target=self._background_monitor,
            args=(interval_seconds,),
            daemon=True
        )
        self._monitor_thread.start()

    def stop_background_monitoring(self):
        """Stop background monitoring."""
        self._stop_monitoring = True
        if self._monitor_thread:
            self._monitor_thread.join(timeout=5)

    def add_alert_handler(self, handler: Callable):
        """Add an alert handler function."""
        self.alert_handlers.append(handler)

    def check_alerts(self) -> List[Dict[str, Any]]:
        """Check for performance alerts."""
        alerts = []

        stats = self.metrics.get_all_stats()

        for func_name, func_stats in stats['functions'].items():
            # Check error rate
            if func_stats.get('error_rate', 0) > 0.1:  # 10% error rate
                alert = {
                    'type': 'error_rate',
                    'severity': 'high',
                    'function': func_name,
                    'message': f'High error rate: {func_stats["error_rate"]:.2%}',
                    'details': func_stats
                }
                alerts.append(alert)

            # Check execution time
            exec_time = func_stats.get('execution_time', {})
            if exec_time.get('p95', 0) > 30.0:  # 30 seconds
                alert = {
                    'type': 'slow_execution',
                    'severity': 'medium',
                    'function': func_name,
                    'message': f'Slow execution (P95): {exec_time["p95"]:.2f}s',
                    'details': func_stats
                }
                alerts.append(alert)

        # Check system resources
        system_alerts = self._check_system_alerts()
        alerts.extend(system_alerts)

        # Store alerts
        self.alerts.extend(alerts)

        # Trigger alert handlers
        for alert in alerts:
            for handler in self.alert_handlers:
                try:
                    handler(alert)
                except Exception:
                    pass  # Ignore handler errors

        return alerts

    def generate_report(self, format: str = "json") -> str:
        """Generate a performance report."""
        stats = self.metrics.get_all_stats()
        report = {
            'timestamp': datetime.utcnow().isoformat(),
            'monitoring_period': 'current_session',
            'summary': stats['summary'],
            'function_details': stats['functions'],
            'alerts': self.alerts[-10:],  # Last 10 alerts
            'system_info': self._get_system_info()
        }

        if format == "json":
            return json.dumps(report, indent=2, default=str)
        else:
            # Simple text format
            lines = [f"Performance Report - {report['timestamp']}"]
            lines.append(f"Total Functions: {report['summary']['total_functions']}")
            lines.append(f"Total Calls: {report['summary']['total_calls']}")
            lines.append(f"Total Errors: {report['summary']['total_errors']}")
            lines.append("")

            for func_name, details in report['function_details'].items():
                lines.append(f"Function: {func_name}")
                lines.append(f"  Calls: {details['total_calls']}")
                if 'execution_time' in details:
                    exec_time = details['execution_time']
                    lines.append(f"  Avg Time: {exec_time['mean']:.3f}s")
                    lines.append(f"  P95 Time: {exec_time['p95']:.3f}s")
                lines.append("")

            return "\n".join(lines)

    def _background_monitor(self, interval_seconds: int):
        """Background monitoring loop."""
        while not self._stop_monitoring:
            try:
                self.check_alerts()
                time.sleep(interval_seconds)
            except Exception:
                time.sleep(interval_seconds)  # Continue monitoring even if checks fail

    def _check_system_alerts(self) -> List[Dict[str, Any]]:
        """Check system resource alerts."""
        alerts = []

        try:
            # CPU usage
            cpu_percent = psutil.cpu_percent(interval=1)
            if cpu_percent > 90:
                alerts.append({
                    'type': 'system_cpu',
                    'severity': 'high',
                    'message': f'High CPU usage: {cpu_percent:.1f}%',
                    'details': {'cpu_percent': cpu_percent}
                })

            # Memory usage
            memory = psutil.virtual_memory()
            if memory.percent > 90:
                alerts.append({
                    'type': 'system_memory',
                    'severity': 'high',
                    'message': f'High memory usage: {memory.percent:.1f}%',
                    'details': {'memory_percent': memory.percent}
                })

            # Disk usage
            disk = psutil.disk_usage('/')
            if disk.percent > 95:
                alerts.append({
                    'type': 'system_disk',
                    'severity': 'medium',
                    'message': f'High disk usage: {disk.percent:.1f}%',
                    'details': {'disk_percent': disk.percent}
                })

        except Exception:
            pass  # Ignore system monitoring errors

        return alerts

    def _get_memory_usage(self) -> float:
        """Get current memory usage in MB."""
        try:
            process = psutil.Process(os.getpid())
            return process.memory_info().rss / 1024 / 1024
        except Exception:
            return 0.0

    def _get_system_info(self) -> Dict[str, Any]:
        """Get system information."""
        try:
            return {
                'cpu_count': psutil.cpu_count(),
                'memory_total': psutil.virtual_memory().total / 1024 / 1024 / 1024,  # GB
                'disk_total': psutil.disk_usage('/').total / 1024 / 1024 / 1024,  # GB
                'platform': os.uname().sysname if hasattr(os, 'uname') else 'Unknown'
            }
        except Exception:
            return {}


class PerformanceDashboard:
    """Web-based performance dashboard."""

    def __init__(self, monitor: PerformanceMonitor, host: str = "localhost", port: int = 8081):
        self.monitor = monitor
        self.host = host
        self.port = port
        self.server = None

    def start_dashboard(self):
        """Start the performance dashboard server."""
        try:
            from flask import Flask, jsonify, render_template_string

            app = Flask(__name__)

            @app.route('/')
            def dashboard():
                return render_template_string(self._get_dashboard_html())

            @app.route('/api/stats')
            def api_stats():
                return jsonify(self.monitor.metrics.get_all_stats())

            @app.route('/api/alerts')
            def api_alerts():
                return jsonify({'alerts': self.monitor.alerts[-50:]})  # Last 50 alerts

            @app.route('/api/profiles/<function_name>')
            def api_profiles(function_name):
                profiles = self.monitor.profiler.get_function_profiles(function_name, 20)
                return jsonify({'profiles': profiles})

            print(f"🚀 Performance dashboard started at http://{self.host}:{self.port}")
            app.run(host=self.host, port=self.port, debug=False)

        except ImportError:
            print("❌ Flask not installed. Install with: pip install flask")
        except Exception as e:
            print(f"❌ Failed to start dashboard: {e}")

    def _get_dashboard_html(self) -> str:
        """Get the dashboard HTML template."""
        return """
<!DOCTYPE html>
<html>
<head>
    <title>FlyPy Performance Dashboard</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .metric { background: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 5px; }
        .alert { background: #ffebee; border: 1px solid #f44336; padding: 10px; margin: 10px 0; }
        .chart { width: 400px; height: 200px; }
    </style>
</head>
<body>
    <h1>🚀 FlyPy Performance Dashboard</h1>

    <div id="stats" class="metric">
        <h2>Performance Statistics</h2>
        <p>Loading...</p>
    </div>

    <div id="alerts">
        <h2>Recent Alerts</h2>
        <p>Loading...</p>
    </div>

    <script>
        async function updateStats() {
            try {
                const response = await fetch('/api/stats');
                const data = await response.json();

                const statsDiv = document.getElementById('stats');
                statsDiv.innerHTML = `
                    <h2>Performance Statistics</h2>
                    <p>Total Functions: ${data.summary.total_functions}</p>
                    <p>Total Calls: ${data.summary.total_calls}</p>
                    <p>Total Errors: ${data.summary.total_errors}</p>
                    <p>Average Error Rate: ${(data.summary.avg_error_rate * 100).toFixed(2)}%</p>
                `;
            } catch (error) {
                console.error('Failed to fetch stats:', error);
            }
        }

        async function updateAlerts() {
            try {
                const response = await fetch('/api/alerts');
                const data = await response.json();

                const alertsDiv = document.getElementById('alerts');
                if (data.alerts.length === 0) {
                    alertsDiv.innerHTML = '<h2>Recent Alerts</h2><p>No alerts</p>';
                } else {
                    alertsDiv.innerHTML = '<h2>Recent Alerts</h2>' +
                        data.alerts.map(alert =>
                            `<div class="alert">
                                <strong>${alert.type.toUpperCase()}</strong> (${alert.severity}): ${alert.message}
                            </div>`
                        ).join('');
                }
            } catch (error) {
                console.error('Failed to fetch alerts:', error);
            }
        }

        // Update every 5 seconds
        updateStats();
        updateAlerts();
        setInterval(() => {
            updateStats();
            updateAlerts();
        }, 5000);
    </script>
</body>
</html>
        """


# Global performance monitor instance
performance_monitor = PerformanceMonitor()


def monitor_performance(enable_profiling: bool = False):
    """Decorator to monitor function performance."""
    return performance_monitor.monitor_function(enable_profiling=enable_profiling)


def get_performance_stats() -> Dict[str, Any]:
    """Get current performance statistics."""
    return performance_monitor.metrics.get_all_stats()


def get_performance_report(format: str = "json") -> str:
    """Generate a performance report."""
    return performance_monitor.generate_report(format=format)


def start_performance_monitoring(interval_seconds: int = 60):
    """Start background performance monitoring."""
    performance_monitor.start_background_monitoring(interval_seconds)


def stop_performance_monitoring():
    """Stop background performance monitoring."""
    performance_monitor.stop_background_monitoring()


def check_performance_alerts() -> List[Dict[str, Any]]:
    """Check for performance alerts."""
    return performance_monitor.check_alerts()


def start_performance_dashboard(host: str = "localhost", port: int = 8081):
    """Start the performance dashboard."""
    dashboard = PerformanceDashboard(performance_monitor, host, port)

    # Start dashboard in a separate thread
    import threading
    dashboard_thread = threading.Thread(target=dashboard.start_dashboard, daemon=True)
    dashboard_thread.start()

    print(f"📊 Performance dashboard starting at http://{host}:{port}")
    return dashboard_thread