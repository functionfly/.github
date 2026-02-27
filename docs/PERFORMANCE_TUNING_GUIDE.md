# FunctionFly Performance Tuning Guide

## Overview

This guide provides comprehensive strategies for performance tuning and load testing FunctionFly, a virtual edge layer for routing requests to edge computing backends.

## Architecture Overview

FunctionFly consists of several performance-critical components:

- **Routing Engine**: Health-aware backend selection with EWMA latency scoring
- **StateFabric**: Multi-layer caching (LRU + Redis) with event sourcing
- **API Gateway**: HTTP request processing and metrics collection
- **Monitoring System**: Real-time performance tracking and alerting

## Performance Bottlenecks Analysis

### 1. Routing Performance

**Current Implementation:**

- EWMA (Exponentially Weighted Moving Average) for latency scoring
- Circuit breaker pattern for fault tolerance
- Priority-based backend selection for Pro plans

**Potential Bottlenecks:**

- Database queries for health checks on every routing decision
- EWMA calculations on hot path
- Circuit state lookups

**Optimization Strategies:**

- Implement caching for circuit states and health metrics
- Pre-compute EWMA scores and update asynchronously
- Use Redis for distributed circuit state management

### 2. Caching Performance

**Current Implementation:**

- Two-layer caching: LRU (Phase 1) + Redis (Phase 2)
- Metadata, snapshots, and state caching
- Rate limiting and session management

**Potential Bottlenecks:**

- Serialization/deserialization overhead
- Redis connection pool contention
- Cache key generation complexity

**Optimization Strategies:**

- Implement connection pooling improvements
- Optimize serialization with custom codecs
- Implement cache warming strategies

### 3. Database Performance

**Current Implementation:**

- PostgreSQL with comprehensive indexing
- Connection pooling with pgx
- Complex queries for analytics and monitoring

**Potential Bottlenecks:**

- N+1 query problems in health checking
- Lock contention on frequent updates
- Large result sets for analytics

**Optimization Strategies:**

- Implement query result caching
- Use database read replicas for health checks
- Optimize indexes for common query patterns

## Load Testing Infrastructure

### 1. Artillery Load Tests

Located in `load-tests/api-load-test.yml`, provides:

- Realistic HTTP request patterns
- Authentication flow simulation
- Function execution scenarios
- Error handling validation

**Usage:**

```bash
npm install -g artillery
artillery run load-tests/api-load-test.yml
```

### 2. K6 Advanced Load Tests

Located in `load-tests/k6-load-test.js`, provides:

- Custom metrics collection
- Spike and stress testing scenarios
- Distributed load simulation
- Performance threshold validation

**Usage:**

```bash
k6 run load-tests/k6-load-test.js
```

### 3. Distributed Load Testing

Located in `load-tests/distributed-load-test.py`, provides:

- Multi-region load testing
- Provider-specific testing
- Geo-routing validation
- Comparative performance analysis

**Usage:**

```python
python load-tests/distributed-load-test.py
```

### 4. Database Load Testing

Using pgbench for database-specific load testing:

```bash
# TPC-B benchmark
make load-test-tpcb

# Mixed read/write test
make load-test-mixed

# Custom application queries
make load-test-custom
```

## Performance Benchmarks

### Go Benchmarks

Run comprehensive Go benchmarks:

```bash
# All benchmarks
make bench

# Database-specific benchmarks
make bench-db

# Database benchmarks with profiling
make bench-db-profile
```

### Rust Benchmarks

Run StateFabric benchmarks:

```bash
cd statefabric
cargo bench
```

Key benchmark areas:

- LRU cache operations
- Redis cache interactions
- Event processing and serialization
- Concurrent state access patterns

## Monitoring and Metrics

### Prometheus Metrics

Comprehensive metrics collection implemented in `internal/monitoring/metrics.go`:

**HTTP Metrics:**

- Request duration by endpoint, method, backend
- Request counts and error rates
- Routing decision tracking

**Cache Metrics:**

- Hit/miss ratios
- Cache size and eviction rates
- Redis connection pool status

**Database Metrics:**

- Query execution times
- Connection pool utilization
- Transaction performance

**Business Metrics:**

- Active user sessions
- Function deployment rates
- Geographic request distribution

### Custom Dashboards

Performance metrics are exposed via Prometheus for integration with Grafana dashboards.

## Optimization Strategies

### 1. Routing Optimizations

**Circuit Breaker Caching:**

```go
// Cache circuit states in Redis to avoid DB lookups
circuitState, err := redisCache.GetCircuitState(backendID)
if err != nil {
    // Fallback to database
    circuitState, err = db.GetCircuitState(backendID)
    // Cache for future use
    redisCache.SetCircuitState(backendID, circuitState)
}
```

**EWMA Pre-computation:**

```go
// Update EWMA scores asynchronously
func (r *Router) updateEWMAScores() {
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        backends := r.getAllBackends()
        for _, backend := range backends {
            score := r.calculateEWMAScore(backend.ID)
            r.cache.SetEWMAScore(backend.ID, score)
        }
    }
}
```

### 2. Cache Optimizations

**Connection Pool Tuning:**

```rust
let redis_config = RedisConfig {
    urls: vec!["redis://localhost:6379".to_string()],
    key_prefix: "functionfly".to_string(),
    default_ttl: 3600,
    max_connections: 50, // Increased from default
    connection_timeout: Duration::from_millis(100),
    operation_timeout: Duration::from_millis(500),
};
```

**Serialization Optimization:**

```rust
// Use bincode for faster serialization
let serialized = bincode::serialize(&data)?;
let deserialized: Value = bincode::deserialize(&serialized)?;
```

### 3. Database Optimizations

**Read Replica Usage:**

```go
// Use read replicas for health checks
func (r *Router) GetRecentHealthChecks(backendID uuid.UUID, limit int) ([]*HealthCheck, error) {
    // Route to read replica
    return r.readReplicaDB.GetRecentHealthChecks(backendID, limit)
}
```

**Query Optimization:**

```sql
-- Add composite indexes for common queries
CREATE INDEX CONCURRENTLY idx_health_checks_backend_time
ON health_checks (backend_id, checked_at DESC);

CREATE INDEX CONCURRENTLY idx_routing_events_app_backend
ON routing_events (app_id, backend_id, created_at DESC);
```

## Performance Testing Best Practices

### 1. Test Environment Setup

- Use staging environment that mirrors production
- Ensure sufficient resources (CPU, memory, network)
- Pre-warm caches and connection pools
- Disable non-essential services during testing

### 2. Load Test Scenarios

**Realistic Traffic Patterns:**

- Authentication flows (login, token validation)
- Function execution with varying payloads
- Health monitoring and status checks
- Error scenarios and failure handling

**Stress Testing:**

- Sudden traffic spikes
- Sustained high load
- Resource exhaustion scenarios
- Network partition simulation

### 3. Performance Baselines

Establish performance baselines for:

- Response times (p50, p95, p99)
- Throughput (requests/second)
- Error rates
- Resource utilization
- Cache hit rates

### 4. Continuous Performance Testing

Integrate performance testing into CI/CD:

```yaml
# .github/workflows/performance.yml
name: Performance Tests
on:
  pull_request:
    paths:
      - 'internal/routing/**'
      - 'statefabric/**'
      - 'internal/api/**'

jobs:
  performance-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run benchmarks
        run: make bench
      - name: Run load tests
        run: |
          artillery run load-tests/api-load-test.yml --output report.json
          k6 run load-tests/k6-load-test.js
```

## Troubleshooting Performance Issues

### High Latency Issues

1. **Check routing performance:**

   ```bash
   # Monitor routing latency
   curl http://localhost:9090/api/v1/query?query=routing_latency_seconds
   ```

2. **Analyze database queries:**

   ```bash
   # Check slow queries
   heroku pg:diagnose --app functionfly
   ```

3. **Cache performance:**

   ```bash
   # Monitor cache hit rates
   curl http://localhost:9090/api/v1/query?query=cache_hit_rate
   ```

### High Error Rates

1. **Circuit breaker status:**

   ```bash
   # Check circuit breaker states
   curl http://localhost:8080/api/circuit-breakers
   ```

2. **Backend health:**

   ```bash
   # Monitor backend health
   curl http://localhost:8080/api/backends/health
   ```

### Resource Contention

1. **Connection pool monitoring:**

   ```bash
   # Check database connections
   curl http://localhost:8080/api/metrics | grep db_connections
   ```

2. **Redis performance:**

   ```bash
   # Monitor Redis stats
   redis-cli info stats
   ```

## Scaling Strategies

### Horizontal Scaling

- **API Gateway:** Deploy multiple instances behind load balancer
- **StateFabric:** Scale Redis cluster for distributed caching
- **Database:** Use read replicas for health check queries

### Vertical Scaling

- **Memory:** Increase for larger caches
- **CPU:** Additional cores for parallel request processing
- **Network:** Higher bandwidth for inter-region communication

### Geographic Distribution

- Deploy edge instances closer to users
- Implement geo-aware routing decisions
- Use CDN for static asset delivery

## Conclusion

FunctionFly's performance tuning focuses on optimizing the critical path of request routing while maintaining high availability and fault tolerance. The combination of comprehensive load testing, detailed metrics collection, and strategic optimizations ensures the system can handle production traffic at scale.

Regular performance testing and monitoring are essential for maintaining optimal performance as the system evolves and traffic patterns change.
