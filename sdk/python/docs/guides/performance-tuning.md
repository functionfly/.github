# Performance Tuning Guide

This guide covers advanced performance optimization techniques for FlyPy functions, including profiling, optimization strategies, and monitoring.

## Table of Contents

- [Performance Profiling](#performance-profiling)
- [Function Optimization](#function-optimization)
- [Memory Management](#memory-management)
- [I/O Optimization](#io-optimization)
- [Caching Strategies](#caching-strategies)
- [Build Optimizations](#build-optimizations)
- [Monitoring and Alerting](#monitoring-and-alerting)

## Performance Profiling

### Using FlyPy Profiler

FlyPy includes built-in profiling tools:

```python
import flypy
import time

@flypy.function(
    name="profiled-function",
    description="Function with performance profiling"
)
def profiled_function(data: dict) -> dict:
    """Function that demonstrates profiling techniques."""

    # Start profiling
    start_time = time.perf_counter()

    # Simulate work
    result = process_data(data)

    # End profiling
    end_time = time.perf_counter()
    execution_time = end_time - start_time

    return {
        "result": result,
        "execution_time_seconds": execution_time,
        "performance_metrics": {
            "cpu_time": execution_time,
            "memory_usage": "N/A",  # Would need external monitoring
        }
    }
```

### Profiling with External Tools

Use Python's built-in profiler:

```python
import cProfile
import pstats
from io import StringIO

@flypy.function(name="profile-external")
def profile_external(data: dict) -> dict:
    """Function that can be profiled externally."""

    # Enable profiling for this function
    pr = cProfile.Profile()
    pr.enable()

    try:
        result = expensive_operation(data)
        return result
    finally:
        pr.disable()

        # Generate profile report
        s = StringIO()
        ps = pstats.Stats(pr, stream=s).sort_stats('cumulative')
        ps.print_stats()

        # In a real application, you might log this or return it
        profile_data = s.getvalue()
        print(f"Profile data: {profile_data[:500]}...")  # First 500 chars
```

### Custom Performance Monitoring

```python
class PerformanceMonitor:
    """Custom performance monitoring utility."""

    def __init__(self):
        self.metrics = {}

    def start_operation(self, operation_name: str):
        """Start timing an operation."""
        self.metrics[operation_name] = {
            'start_time': time.perf_counter(),
            'start_memory': self._get_memory_usage()
        }

    def end_operation(self, operation_name: str) -> dict:
        """End timing an operation and return metrics."""
        if operation_name not in self.metrics:
            return {}

        start_data = self.metrics[operation_name]
        end_time = time.perf_counter()
        end_memory = self._get_memory_usage()

        metrics = {
            'duration_seconds': end_time - start_data['start_time'],
            'memory_delta_kb': (end_memory - start_data['start_memory']) / 1024,
            'end_memory_kb': end_memory / 1024
        }

        del self.metrics[operation_name]
        return metrics

    def _get_memory_usage(self) -> float:
        """Get current memory usage in bytes."""
        try:
            import psutil
            process = psutil.Process()
            return process.memory_info().rss
        except ImportError:
            return 0  # Fallback if psutil not available

# Usage
monitor = PerformanceMonitor()

@flypy.function(name="monitored-function")
def monitored_function(data: dict) -> dict:
    monitor.start_operation("total_execution")

    monitor.start_operation("data_processing")
    processed_data = process_data(data)
    processing_metrics = monitor.end_operation("data_processing")

    monitor.start_operation("result_formatting")
    result = format_result(processed_data)
    formatting_metrics = monitor.end_operation("result_formatting")

    total_metrics = monitor.end_operation("total_execution")

    return {
        "result": result,
        "performance": {
            "total": total_metrics,
            "processing": processing_metrics,
            "formatting": formatting_metrics
        }
    }
```

## Function Optimization

### Algorithm Optimization

Choose efficient algorithms and data structures:

```python
from typing import List, Dict, Any
import flypy

# Inefficient approach
@flypy.function(name="inefficient-search")
def inefficient_search(items: List[Dict[str, Any]], target_id: str) -> Dict[str, Any]:
    """O(n) search - avoid for large datasets."""
    for item in items:
        if item.get("id") == target_id:
            return item
    return {"error": "Not found"}

# Efficient approach
@flypy.function(name="efficient-search")
def efficient_search(items: List[Dict[str, Any]], target_id: str) -> Dict[str, Any]:
    """O(1) search using dictionary lookup."""
    # Pre-process into lookup table
    lookup = {item["id"]: item for item in items}

    result = lookup.get(target_id)
    if result:
        return result
    return {"error": "Not found"}

# Optimized sorting
@flypy.function(name="optimized-sort")
def optimized_sort(items: List[Dict[str, Any]], sort_by: str = "name") -> List[Dict[str, Any]]:
    """Use Timsort (Python's built-in sort) which is efficient."""
    # For small datasets (< 1000 items), Timsort is fastest
    # For larger datasets, consider external sorting

    if len(items) < 1000:
        return sorted(items, key=lambda x: x.get(sort_by, ""))
    else:
        # For large datasets, consider pagination or streaming
        raise ValueError("Dataset too large for in-memory sorting")
```

### Loop Optimization

```python
import flypy

# Inefficient loops
@flypy.function(name="inefficient-loops")
def inefficient_loops(data: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
    """Multiple passes over data - inefficient."""
    # Pass 1: Filter
    filtered = []
    for item in data:
        if item.get("active", False):
            filtered.append(item)

    # Pass 2: Transform
    transformed = []
    for item in filtered:
        transformed.append({
            "id": item["id"],
            "name": item["name"].upper(),
            "value": item.get("value", 0) * 2
        })

    # Pass 3: Sort
    sorted_items = sorted(transformed, key=lambda x: x["value"], reverse=True)

    return sorted_items

# Efficient single-pass processing
@flypy.function(name="efficient-processing")
def efficient_processing(data: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
    """Single pass with list comprehension - efficient."""
    # Single pass: filter, transform, and prepare for sorting
    processed = [
        {
            "id": item["id"],
            "name": item["name"].upper(),
            "value": item.get("value", 0) * 2
        }
        for item in data
        if item.get("active", False)
    ]

    # Sort the results
    processed.sort(key=lambda x: x["value"], reverse=True)

    return processed
```

### Memory-Efficient Processing

```python
from typing import Iterator, Generator
import flypy

# Memory-efficient generator approach
@flypy.function(name="memory-efficient-generator")
def memory_efficient_generator(data: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
    """Process large datasets using generators."""

    def process_items() -> Iterator[Dict[str, Any]]:
        """Generator that processes items one at a time."""
        for item in data:
            if item.get("active", False):
                # Process item without storing all in memory
                yield {
                    "id": item["id"],
                    "name": item["name"].upper(),
                    "processed": True
                }

    # Collect results (in real app, might stream them)
    results = list(process_items())

    # Sort only the final results
    results.sort(key=lambda x: x["name"])

    return results

# Streaming for very large datasets
@flypy.function(
    name="streaming-processor",
    capabilities=["database", "network"]
)
def streaming_processor(config: Dict[str, Any]) -> Dict[str, Any]:
    """Process data in streaming fashion."""

    batch_size = config.get("batch_size", 1000)
    total_processed = 0
    total_filtered = 0

    # In a real implementation, this would stream from a database
    # or external source rather than loading everything at once

    # Simulate streaming processing
    for batch in get_data_batches(batch_size):
        processed_batch = process_batch(batch)
        total_processed += len(processed_batch)

        # Filter results
        filtered_batch = [item for item in processed_batch if item.get("score", 0) > 0.8]
        total_filtered += len(filtered_batch)

        # In real streaming, you'd write results to output stream here
        # rather than accumulating them

    return {
        "total_processed": total_processed,
        "total_filtered": total_filtered,
        "efficiency_ratio": total_filtered / total_processed if total_processed > 0 else 0
    }

def get_data_batches(batch_size: int) -> Iterator[List[Dict[str, Any]]]:
    """Simulate getting data in batches."""
    # In real implementation, this would query database with LIMIT/OFFSET
    # or use a streaming API
    for i in range(10):  # Simulate 10 batches
        yield [{"id": f"item_{i}_{j}", "data": f"value_{j}"} for j in range(batch_size)]
```

## Memory Management

### Object Reuse

```python
import flypy

class ReusableObjects:
    """Pool of reusable objects to reduce allocations."""

    def __init__(self):
        self.string_builder = []
        self.temp_dict = {}

    def get_string_builder(self):
        """Get reusable string builder."""
        self.string_builder.clear()
        return self.string_builder

    def get_temp_dict(self):
        """Get reusable temporary dictionary."""
        self.temp_dict.clear()
        return self.temp_dict

# Global instance for reuse across function calls
reusable_objects = ReusableObjects()

@flypy.function(name="memory-efficient-function")
def memory_efficient_function(items: List[Dict[str, Any]]) -> str:
    """Build result string efficiently."""

    builder = reusable_objects.get_string_builder()

    for item in items:
        builder.append(f"Item: {item['id']}, Value: {item.get('value', 'N/A')}\n")

    result = ''.join(builder)
    return result
```

### Large Object Handling

```python
import flypy

@flypy.function(
    name="large-object-handler",
    max_execution_time=60000  # 60 seconds for large operations
)
def large_object_handler(data: Dict[str, Any]) -> Dict[str, Any]:
    """Handle large objects efficiently."""

    # For large objects, process in chunks
    large_data = data.get("large_array", [])

    if len(large_data) > 10000:
        # Process in chunks to avoid memory spikes
        chunk_size = 1000
        results = []

        for i in range(0, len(large_data), chunk_size):
            chunk = large_data[i:i + chunk_size]
            chunk_result = process_chunk(chunk)
            results.extend(chunk_result)

        return {
            "processed": True,
            "total_items": len(large_data),
            "results": results[:1000]  # Limit output size
        }
    else:
        # Process normally for smaller datasets
        return process_small_data(data)

def process_chunk(chunk: List[Any]) -> List[Dict[str, Any]]:
    """Process a chunk of data."""
    return [{"id": item.get("id"), "processed": True} for item in chunk]

def process_small_data(data: Dict[str, Any]) -> Dict[str, Any]:
    """Process smaller datasets normally."""
    return {"processed": True, "items": len(data.get("large_array", []))}
```

## I/O Optimization

### Batch Operations

```python
import flypy

@flypy.function(
    name="batch-io-operations",
    capabilities=["database", "network"]
)
def batch_io_operations(operations: List[Dict[str, Any]]) -> Dict[str, Any]:
    """Perform batched I/O operations efficiently."""

    # Group operations by type for batching
    reads = []
    writes = []
    deletes = []

    for op in operations:
        if op["type"] == "read":
            reads.append(op)
        elif op["type"] == "write":
            writes.append(op)
        elif op["type"] == "delete":
            deletes.append(op)

    results = {
        "reads": {"count": len(reads), "success": 0},
        "writes": {"count": len(writes), "success": 0},
        "deletes": {"count": len(deletes), "success": 0}
    }

    # Batch reads
    if reads:
        read_results = batch_read([op["key"] for op in reads])
        results["reads"]["success"] = len(read_results)

    # Batch writes
    if writes:
        write_success = batch_write({op["key"]: op["value"] for op in writes})
        results["writes"]["success"] = write_success

    # Batch deletes
    if deletes:
        delete_success = batch_delete([op["key"] for op in deletes])
        results["deletes"]["success"] = delete_success

    return results

def batch_read(keys: List[str]) -> List[Dict[str, Any]]:
    """Simulate batched read operation."""
    # In real implementation, this would be a single database query
    return [{"key": key, "value": f"data_for_{key}"} for key in keys]

def batch_write(data: Dict[str, Any]) -> int:
    """Simulate batched write operation."""
    # In real implementation, this would be a single database transaction
    return len(data)

def batch_delete(keys: List[str]) -> int:
    """Simulate batched delete operation."""
    # In real implementation, this would be a single database operation
    return len(keys)
```

### Connection Pooling

```python
import flypy

class ConnectionPool:
    """Simple connection pool for efficient resource usage."""

    def __init__(self, max_connections: int = 10):
        self.max_connections = max_connections
        self.available_connections = []
        self.used_connections = set()

    def get_connection(self):
        """Get a connection from the pool."""
        if self.available_connections:
            conn = self.available_connections.pop()
            self.used_connections.add(conn)
            return conn
        elif len(self.used_connections) < self.max_connections:
            # Create new connection
            conn = self._create_connection()
            self.used_connections.add(conn)
            return conn
        else:
            raise RuntimeError("No available connections")

    def return_connection(self, conn):
        """Return a connection to the pool."""
        if conn in self.used_connections:
            self.used_connections.remove(conn)
            self.available_connections.append(conn)

    def _create_connection(self):
        """Create a new connection."""
        # In real implementation, this would create database/network connections
        return f"connection_{len(self.used_connections) + len(self.available_connections)}"

# Global connection pool
connection_pool = ConnectionPool(max_connections=5)

@flypy.function(
    name="connection-pooled-operation",
    capabilities=["database"]
)
def connection_pooled_operation(query: str) -> Dict[str, Any]:
    """Perform database operation with connection pooling."""

    conn = None
    try:
        conn = connection_pool.get_connection()
        # Simulate database operation
        result = perform_query(conn, query)
        return {"result": result, "connection": conn}
    finally:
        if conn:
            connection_pool.return_connection(conn)
```

## Caching Strategies

### Function-Level Caching

```python
import flypy
from functools import lru_cache
import hashlib
import json

# In-memory cache for deterministic functions
@lru_cache(maxsize=1000)
def cached_expensive_operation(data_hash: str, operation: str) -> Dict[str, Any]:
    """Cache expensive operations based on input hash."""

    # In real implementation, you'd perform the actual expensive operation
    # This is just a simulation
    return {
        "operation": operation,
        "data_hash": data_hash,
        "result": f"computed_result_for_{data_hash}_{operation}",
        "cached": False  # Would be True if retrieved from cache
    }

@flypy.function(
    name="cached-function",
    deterministic=True,
    cache_ttl=3600  # 1 hour
)
def cached_function(data: Dict[str, Any], operation: str) -> Dict[str, Any]:
    """Function that uses caching for expensive operations."""

    # Create hash of input data for caching key
    data_str = json.dumps(data, sort_keys=True)
    data_hash = hashlib.md5(data_str.encode()).hexdigest()

    # Use cached operation
    result = cached_expensive_operation(data_hash, operation)

    return {
        "result": result,
        "input_hash": data_hash,
        "cache_hit": result.get("cached", False)
    }
```

### Multi-Level Caching

```python
import flypy

class CacheManager:
    """Multi-level caching manager."""

    def __init__(self):
        self.l1_cache = {}  # Fast in-memory cache
        self.l2_cache = {}  # Slower but larger cache
        self.max_l1_size = 100
        self.max_l2_size = 1000

    def get(self, key: str) -> Any:
        """Get from multi-level cache."""
        # Check L1 cache first
        if key in self.l1_cache:
            return self.l1_cache[key]

        # Check L2 cache
        if key in self.l2_cache:
            # Promote to L1
            value = self.l2_cache[key]
            self._add_to_l1(key, value)
            return value

        return None

    def set(self, key: str, value: Any):
        """Set in multi-level cache."""
        self._add_to_l1(key, value)
        self._add_to_l2(key, value)

    def _add_to_l1(self, key: str, value: Any):
        """Add to L1 cache with LRU eviction."""
        if len(self.l1_cache) >= self.max_l1_size:
            # Remove oldest item (simple FIFO for demo)
            oldest_key = next(iter(self.l1_cache))
            del self.l1_cache[oldest_key]

        self.l1_cache[key] = value

    def _add_to_l2(self, key: str, value: Any):
        """Add to L2 cache."""
        if len(self.l2_cache) >= self.max_l2_size:
            oldest_key = next(iter(self.l2_cache))
            del self.l2_cache[oldest_key]

        self.l2_cache[key] = value

# Global cache manager
cache_manager = CacheManager()

@flypy.function(
    name="multi-level-cached-function",
    capabilities=["cache"]
)
def multi_level_cached_function(query: str) -> Dict[str, Any]:
    """Function with multi-level caching."""

    cache_key = f"query_{hash(query)}"

    # Check cache first
    cached_result = cache_manager.get(cache_key)
    if cached_result:
        return {
            "result": cached_result,
            "source": "cache",
            "cache_level": "multi-level"
        }

    # Compute result
    result = perform_expensive_query(query)

    # Cache result
    cache_manager.set(cache_key, result)

    return {
        "result": result,
        "source": "computed",
        "cached": True
    }

def perform_expensive_query(query: str) -> Dict[str, Any]:
    """Simulate expensive query operation."""
    # In real implementation, this would be a database query
    return {"query": query, "result": f"data_for_{query}"}
```

## Build Optimizations

### Parallel Compilation

```bash
# Build multiple functions in parallel
flypy build functions.py utils.py handlers.py --parallel

# Build with optimized settings
flypy build --optimize-size --optimize-speed
```

### Incremental Builds

```python
# Use build cache for incremental builds
flypy build --cache-dir .flypy-cache --incremental
```

## Monitoring and Alerting

### Performance Metrics Collection

```python
import flypy
import time
import statistics

class PerformanceMetrics:
    """Collect and analyze performance metrics."""

    def __init__(self):
        self.execution_times = []
        self.memory_usage = []
        self.error_counts = {}

    def record_execution(self, function_name: str, execution_time: float, memory_used: int, success: bool):
        """Record execution metrics."""
        self.execution_times.append({
            'function': function_name,
            'time': execution_time,
            'memory': memory_used,
            'timestamp': time.time(),
            'success': success
        })

        if not success:
            self.error_counts[function_name] = self.error_counts.get(function_name, 0) + 1

    def get_stats(self) -> Dict[str, Any]:
        """Get performance statistics."""
        if not self.execution_times:
            return {}

        times = [entry['time'] for entry in self.execution_times]
        memory = [entry['memory'] for entry in self.execution_times]

        return {
            'total_executions': len(self.execution_times),
            'avg_execution_time': statistics.mean(times),
            'median_execution_time': statistics.median(times),
            'p95_execution_time': statistics.quantiles(times, n=20)[18],  # 95th percentile
            'max_execution_time': max(times),
            'avg_memory_usage': statistics.mean(memory),
            'total_errors': sum(self.error_counts.values()),
            'error_rate': sum(self.error_counts.values()) / len(self.execution_times)
        }

# Global metrics collector
metrics = PerformanceMetrics()

@flypy.function(name="monitored-function")
def monitored_function(data: dict) -> dict:
    """Function with performance monitoring."""

    start_time = time.perf_counter()
    start_memory = 0  # Would get actual memory usage

    try:
        result = perform_work(data)

        execution_time = time.perf_counter() - start_time
        memory_used = 0  # Would calculate actual memory usage

        metrics.record_execution("monitored-function", execution_time, memory_used, True)

        return result

    except Exception as e:
        execution_time = time.perf_counter() - start_time
        metrics.record_execution("monitored-function", execution_time, 0, False)

        raise e

@flypy.function(name="get-performance-stats")
def get_performance_stats() -> dict:
    """Get current performance statistics."""
    return metrics.get_stats()
```

### Automated Alerting

```python
import flypy

class AlertManager:
    """Manage performance alerts."""

    def __init__(self):
        self.alerts = []
        self.thresholds = {
            'max_execution_time': 30.0,  # seconds
            'error_rate_threshold': 0.05,  # 5%
            'memory_threshold': 100 * 1024 * 1024  # 100MB
        }

    def check_alerts(self, metrics: dict) -> List[dict]:
        """Check metrics against thresholds and generate alerts."""
        alerts = []

        if metrics.get('avg_execution_time', 0) > self.thresholds['max_execution_time']:
            alerts.append({
                'type': 'performance',
                'severity': 'high',
                'message': f'Average execution time {metrics["avg_execution_time"]:.2f}s exceeds threshold {self.thresholds["max_execution_time"]}s'
            })

        if metrics.get('error_rate', 0) > self.thresholds['error_rate_threshold']:
            alerts.append({
                'type': 'reliability',
                'severity': 'critical',
                'message': f'Error rate {metrics["error_rate"]:.2%} exceeds threshold {self.thresholds["error_rate_threshold"]:.2%}'
            })

        if metrics.get('avg_memory_usage', 0) > self.thresholds['memory_threshold']:
            alerts.append({
                'type': 'resource',
                'severity': 'medium',
                'message': f'Average memory usage {metrics["avg_memory_usage"]/1024/1024:.1f}MB exceeds threshold {self.thresholds["memory_threshold"]/1024/1024}MB'
            })

        self.alerts.extend(alerts)
        return alerts

# Global alert manager
alert_manager = AlertManager()

@flypy.function(name="check-system-health")
def check_system_health() -> dict:
    """Check system health and generate alerts."""

    # Get current metrics
    current_metrics = metrics.get_stats()

    # Check for alerts
    alerts = alert_manager.check_alerts(current_metrics)

    return {
        'status': 'healthy' if not alerts else 'degraded',
        'metrics': current_metrics,
        'alerts': alerts,
        'timestamp': time.time()
    }
```

This comprehensive performance tuning guide provides practical techniques for optimizing FlyPy functions across all aspects of performance: profiling, algorithm optimization, memory management, I/O efficiency, caching, and monitoring.