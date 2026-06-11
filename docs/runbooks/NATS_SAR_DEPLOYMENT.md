# NATS & SAR Runtime — Production Runbook

**Document Version:** 1.1
**Last Updated:** 2026-06-07
**Owner:** Platform Team

> **Note:** SAR source code has moved to its own repo: [functionfly/sar](https://github.com/functionfly/sar). See that repo for build instructions, CI, and release binaries.

---

## Overview

NATS JetStream (port 4222) is the event bus for FRG async flows and inter-service pub/sub. The SAR (Stateful Agent Runtime, port 8082) registers agents with the orchestrator via NATS and handles lifecycle events. Both are required for full platform functionality — without NATS, FRG async flows fail silently; without SAR, agent registration and agent-based execution are unavailable.

### Components

| Component | Port | Role |
|-----------|------|------|
| NATS server | 4222 | Message broker, JetStream persistence |
| SAR (Stateful Agent Runtime) | 8082 | Agent registry, lifecycle management, graph execution |
| Orchestrator API | 8080 | Publishes FRG events to NATS; subscribes to agent results |

### When NATS or SAR is Down

| Scenario | Impact | Workaround |
|----------|--------|------------|
| NATS down | FRG async flows fail; agent results not delivered | Restart NATS; pending FRG executions retry from orchestrator queue |
| SAR down | New agents cannot register; existing agent calls return 503 | Restart SAR; it reconnects automatically |
| NATS down during SAR run | SAR logs connection errors; recovers when NATS returns | No manual action needed; SAR implements exponential backoff reconnect |

---

## Prerequisites

- NATS server **2.10+** installed (or Docker image `nats:latest`)
- Rust toolchain **1.75+** for SAR (or pre-built binary at `bin/functionfly-sar`)
- Ports **4222** (NATS client), **8222** (NATS monitoring), and **8082** (SAR) open
- `NATS_URL`, `DATABASE_URL`, `REDIS_ADDR` environment variables set

---

## Starting NATS (Idempotent)

### Standalone Binary

```bash
# Start on port 4222 (idempotent — skips if already running)
make nats

# Verify it started
pgrep -x nats-server && echo "NATS running on port 4222"

# Stop (rarely needed)
make nats-stop
```

### Docker

```bash
# Simple start
docker run -d --name nats -p 4222:4222 nats:latest

# With JetStream persistence
docker run -d \
  --name nats \
  -p 4222:4222 \
  -p 8222:8222 \
  -v nats_data:/data \
  nats:latest \
  --jetstream --store_dir=/data/jetstream
```

### Production Cluster (3+ nodes)

For HA, deploy a NATS cluster with JetStream. Minimal production config (`nats-server.conf`):

```
server_name: nats-1
listen: 0.0.0.0:4222

jetstream {
  store_dir: /data/jetstream
  max_mem: 1G
  max_file: 10G
}

cluster {
  name: functionfly
  port: 6222
  routes: [
    nats-route://nats-2:6222
    nats-route://nats-3:6222
  ]
}

logging {
  debug: false
  trace: false
}
```

Key points:
- JetStream streams are **raft-based** and survive node failures as long as a quorum is available
- Minimum 3 nodes for HA; do not run single-node JetStream in production
- Use a shared mount (NFS, EFS, etc.) for `store_dir` in cloud deployments
- Monitor `consumers` and `streams` health via `nats-server -v` or the monitoring port

---

## Starting the SAR Runtime

### Prerequisites

```bash
# Build the SAR binary once (requires Rust)
make build-sar

# Verify binary exists
ls bin/functionfly-sar
```

### On the Host (Development / Single-Node)

```bash
# Requires NATS to be running first
make dev-sar

# Or manually with env vars
NATS_URL=nats://localhost:4222 \
REDIS_URL=redis://localhost:6379 \
./bin/functionfly-sar --api-port 8082

# Override port if 8082 is in use
NATS_URL=nats://localhost:4222 ./bin/functionfly-sar --api-port 9092
```

### Via Docker (Infrastructure Only)

SAR cannot be containerized easily because it requires the Rust toolchain to build. Use `docker-compose.dev.yml` for infrastructure (Postgres, Redis, NATS) and run SAR as a host binary.

For a containerized SAR in production, build a multi-stage Dockerfile:

```dockerfile
# See https://github.com/functionfly/sar for the official Dockerfile
# Or use a pre-built release binary from https://github.com/functionfly/sar/releases
FROM rust:1.75 AS builder
WORKDIR /app
# Clone from the SAR repo or copy from a release
RUN cargo build --release

FROM debian:bookworm-slim
COPY --from=builder /app/target/release/functionfly-sar /app/bin/
EXPOSE 8082
CMD ["/app/bin/functionfly-sar"]
```

Then deploy as a Docker container or Kubernetes Deployment with the appropriate `NATS_URL`, `REDIS_URL`, and `DATABASE_URL` env vars.

---

## Health Checks

### NATS

```bash
# Version check
nats-server -V

# Monitor endpoint (port 8222)
curl localhost:8222/varz | jq .

# Check JetStream streams
nats stream list

# Check consumers
nats consumer list

# Verify JetStream is enabled
curl localhost:8222/varz | jq '.jetstream'
```

**Healthy indicators:**
- `state: 1` (leader or active)
- `leader: true` or `offline: false`
- Stream `messages` count is stable or growing (not dropping unexpectedly)

### SAR

```bash
# Basic health
curl localhost:8082/health

# Detailed health (includes NATS connection status)
curl localhost:8082/health/detailed

# Check agent registry
curl localhost:8082/api/v1/agents | jq '.'
```

### Orchestrator Runtime Health

```bash
# Check if SAR is reachable from orchestrator
curl http://localhost:8080/api/health/detailed | jq '.services.runtimes'

# Expected response: sar runtime status is "connected"
```

---

## Alerting

Set up alerts on the following conditions:

### Critical (Page Immediately)

| Alert | Condition | Action |
|-------|-----------|--------|
| NATS down | `pgrep -x nats-server` fails | Restart NATS; check system memory/disk |
| SAR unreachable | `curl localhost:8082/health` returns non-200 | Check NATS connectivity; restart SAR |
| JetStream stream lag | consumer `pending` > 10000 for > 5 min | Investigate consumer group; check SAR is processing |
| NATS cannot connect to cluster | `nats-server -v` shows repeated reconnect attempts | Check network between nodes; verify quorum |

### Warning (Notify on Next Business Day)

| Alert | Condition | Action |
|-------|-----------|--------|
| SAR high memory | SAR RSS > 500MB | Check for memory leaks; restart SAR |
| Consumer lag | consumer `pending` > 1000 for > 10 min | Monitor; may indicate slow processing |
| NATS file descriptor usage | `varz` shows fds > 80% of limit | Increase `max_fds` in nats-server.conf |
| Repeated SAR reconnects | > 5 reconnect events in 1 hour | Check NATS stability; investigate network |

### Recommended Monitoring Setup

If using Prometheus (via `docker-compose.dev.yml` or the monitoring stack):

```yaml
# prometheus rules for NATS/SAR
groups:
  - name: nats_sar_alerts
    rules:
      - alert: NATSServerDown
        expr: up{job="nats"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "NATS server is down"

      - alert: SARRuntimeDown
        expr: up{job="sar"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "SAR runtime is unreachable"
```

---

## Restart Procedures

### Restarting NATS Without Losing JetStream Streams

JetStream streams are persisted to `store_dir`. A clean restart preserves all streams and consumers:

```bash
# 1. Check stream health before restart
nats stream list
# Note stream names and their states

# 2. Stop NATS gracefully (allow in-flight messages to drain)
nats-server --signal SIGINT
# or
pkill -INT nats-server

# 3. Wait 5 seconds for graceful shutdown
sleep 5

# 4. Restart NATS
make nats
# or
nats-server -c /path/to/nats-server.conf

# 5. Verify streams restored
nats stream list
nats consumer list
```

**Critical: Do not use `SIGKILL` on NATS.** This can corrupt JetStream WAL (write-ahead log) and require stream recovery. Always use `SIGINT` or `SIGTERM` for graceful shutdown.

If JetStream data appears corrupted after restart:

```bash
# Verify JetStream store
nats server check /data/jetstream

# Recover stream (if raft quorum is lost, you may need to)
nats stream raft recover STREAM_NAME
```

### Restarting SAR

SAR reconnect behavior:
- SAR implements **automatic exponential backoff reconnect** to NATS (default: 1s, 2s, 4s, 8s, max 30s)
- On reconnect, SAR re-registers all agents — no data loss
- In-flight agent executions that were interrupted by a NATS outage will be retried by the orchestrator (FRG retry logic)
- SAR does **not** lose agent graph state — it is stored in Redis and reloaded on restart

```bash
# 1. Stop SAR (SIGINT for graceful shutdown)
pkill -INT functionfly-sar

# 2. Wait for graceful shutdown (in-flight requests will complete, up to 30s)
sleep 5

# 3. Restart
make dev-sar

# 4. Verify
curl localhost:8082/health
```

### Restarting After NATS Outage (Mid-Execution)

When NATS goes down and comes back:

1. **Orchestrator** (Go API) will automatically reconnect to NATS and resume publishing FRG events
2. **SAR** will automatically reconnect and resume agent processing
3. **In-flight FRG executions**: The orchestrator implements a retry queue. If an async FRG execution was in progress when NATS went down, the orchestrator will retry delivery on reconnect. No manual intervention required if the orchestrator is configured with retry settings.
4. **SAR agent calls in flight**: If SAR was processing a request and NATS failed, the agent execution is retried by the orchestrator's FRG retry mechanism. SAR itself does not retry — it waits for a new request.

---

## Verifying JetStream Stream Health

```bash
# List all streams
nats stream list

# Inspect a specific stream
nats stream info FRG_ASYNC_RESULTS

# Check consumer lag
nats consumer info FRG_ASYNC_RESULTS my-consumer

# Monitor message throughput (live)
watch -n 2 'nats stream list'
```

If a stream has `offline: true` consumers, the consumer group has lost quorum. Restart the affected consumer (or SAR in this case).

---

## Configuration Reference

### NATS Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `NATS_URL` | NATS server URL | `nats://localhost:4222` |
| `NATS_CREDS` | NATS credentials file (for auth) | (none) |
| `NATS_TLS_CA` | TLS CA file | (none) |
| `NATS_TLS_CERT` | TLS certificate file | (none) |
| `NATS_TLS_KEY` | TLS key file | (none) |

### SAR Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `NATS_URL` | NATS server URL | `nats://localhost:4222` |
| `REDIS_URL` | Redis URL for agent registry cache | `redis://localhost:6379` |
| `DATABASE_URL` | Postgres connection (optional for local dev) | (none) |
| `SAR_API_PORT` | HTTP port for SAR API | `8082` |
| `SAR_MAX_CONCURRENT` | Max concurrent agent executions | `10000` |

### Orchestrator Environment Variables (Related)

| Variable | Description |
|----------|-------------|
| `NATS_URL` | Must match the NATS instance used by SAR |
| `REDIS_ADDR` | Redis for StateFabric caching; shared with SAR |
| `SKIP_MIGRATION_VALIDATION` | `true` for local dev (required) |

---

## Common Issues

### "NATS connection refused" in SAR logs

- Verify NATS is running: `pgrep -x nats-server`
- Check port: `curl localhost:4222` (should return OK)
- Check `NATS_URL` env var in SAR process

### SAR returns 503 on agent calls

- SAR is running but NATS connection is down — check `curl localhost:8082/health/detailed`
- Agent not registered — restart SAR so it re-registers

### FRG async executions not completing

- Check NATS streams exist: `nats stream list`
- Check orchestrator logs for publish errors
- Verify SAR is processing: `curl localhost:8082/health/detailed | jq '.nats_connected'`

### JetStream stream is offline

- Node carrying the RAFT leader is down — wait for failover (automatic in cluster)
- If single-node and stream is offline, restart NATS after ensuring graceful shutdown
- Check disk space on the NATS host — JetStream requires free space to operate

---

## Document Control

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-05-27 | Platform Team | Initial version |

**Next Review:** 2026-11-27