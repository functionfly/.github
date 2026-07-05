# FunctionFly Runtimes — Production Deployment Guide

This guide covers deploying the FunctionFly runtime fleet (`runtimes/bun`,
`runtimes/deno`, `runtimes/nodejs`, `runtimes/kotlin`, `runtimes/ruby`,
`runtimes/wasmedge`, `runtimes/prism`, `runtimes/microvm`, `runtimes/sar-local`)
for production traffic.

---

## 1. Standard port allocation

All runtimes bind to loopback (`127.0.0.1`) by default. The Go orchestrator
runs on the same host and proxies requests via NATS or local HTTP. If you
need a runtime to be reachable from outside the host, bind it explicitly via
the runtime's `--addr` / `--bind-addr` flag.

| Runtime           | Default port       | CLI override            |
| ----------------- | ------------------ | ----------------------- |
| `bun`             | 8091               | `--port`                |
| `deno`            | 8090               | `--port`                |
| `nodejs` (daemon) | 9091               | `--port`                |
| `kotlin`          | 8091               | `--addr 127.0.0.1:PORT` |
| `ruby`            | 8094               | `--port`                |
| `wasmedge`        | 8093               | `--port`                |
| `prism`           | 8084 (CLI default) | `start --address`       |
| `microvm`         | 9091               | `--port`                |
| `sar-local`       | 8082               | `--api-port`            |

The `microvm` orchestrator binds to `0.0.0.0` by default because it must
accept traffic from the Go orchestrator. Override with
`FUNCTIONFLY_MICROVM_BIND_ADDR=127.0.0.1` if the host is not network-isolated.

---

## 2. Required environment variables

Every runtime requires the following in production:

```bash
ENVIRONMENT=production
RUNTIME_API_TOKEN=$(openssl rand -hex 32)   # bearer token (≥ 32 chars)
```

Production-mode startup is refused if `RUNTIME_API_TOKEN` is missing or
shorter than 32 characters. For SAR, the analogous variables are
`SAR_API_KEY` and `SAR_ADMIN_API_KEY`.

### NATS configuration

```bash
NATS_URL=nats://nats.functionfly.internal:4222
NATS_TLS_ENABLED=true
NATS_TLS_CA_CERT=/etc/functionfly/nats/ca.pem
```

By default, runtimes fall back to `nats://localhost:4222` (plaintext). For
production, **always set `NATS_TLS_ENABLED=true` and supply the CA
certificate**. Do not enable NATS without TLS in a multi-tenant environment.

### CORS

`CORS_ALLOWED_ORIGINS` (microvm only): comma-separated list. **If unset,
CORS is disabled entirely** — browsers will reject cross-origin requests,
leaving the service API-to-API only. We do not default to `*`.

### Bind address

```bash
FUNCTIONFLY_MICROVM_BIND_ADDR=127.0.0.1   # only if you don't need network access
```

---

## 3. Container build

Each runtime has a multi-stage `Dockerfile`:

```bash
docker build -t functionfly/bun-runtime:1.0.0      -f runtimes/bun/Dockerfile .
docker build -t functionfly/deno-runtime:1.0.0     -f runtimes/deno/Dockerfile .
docker build -t functionfly/nodejs-runtime:1.0.0   -f runtimes/nodejs/Dockerfile .
docker build -t functionfly/kotlin-runtime:1.0.0   -f runtimes/kotlin/Dockerfile .
docker build -t functionfly/ruby-runtime:1.0.0     -f runtimes/ruby/Dockerfile .
docker build -t functionfly/wasmedge-runtime:1.0.0 -f runtimes/wasmedge/Dockerfile .
docker build -t functionfly/microvm:1.0.0          -f runtimes/microvm/Dockerfile .
docker build -t functionfly/sar:1.0.0              -f runtimes/sar-local/Dockerfile .
```

All Dockerfiles:

- Use multi-stage builds with BuildKit cache mounts for fast rebuilds.
- Run as a non-root user with a fixed UID (10001–10008) to avoid collisions
  with system users.
- Use `tini` as PID 1 so SIGTERM is forwarded correctly.
- Pin the OS base image to `debian:bookworm-slim` for reproducibility.
- Install only runtime libraries — no compilers, headers, or build tools.

---

## 4. Security checklist

Before going to production, verify each runtime passes:

- [ ] `ENVIRONMENT=production` is set
- [ ] `RUNTIME_API_TOKEN` is set and ≥ 32 chars
- [ ] `RUNTIME_API_TOKEN` is stored in a secret manager (Vault, k8s Secret,
      AWS Secrets Manager, etc.), **never** in the image or a config file
- [ ] The container runs as non-root (UID ≥ 10000)
- [ ] Read-only root filesystem (`--read-only` flag for `docker run`,
      `securityContext.readOnlyRootFilesystem: true` for k8s)
- [ ] All capabilities dropped except `NET_BIND_SERVICE` (only if binding
      to ports < 1024)
- [ ] `seccomp` profile is applied (use the Docker default or
      RuntimeDefault)
- [ ] `apparmor` profile applied if running on a supported host
- [ ] Resource limits are set (CPU, memory)
- [ ] The runtime is bound to loopback unless it needs to be reachable
      from outside (e.g. microvm)
- [ ] NATS connection uses TLS with a verified CA cert
- [ ] Logs are shipped to a central aggregator (Datadog, Splunk, ELK)
- [ ] Metrics are scraped by Prometheus (`/metrics` endpoint)
- [ ] Health checks are wired into the orchestrator (`/health` for liveness,
      `/ready` for readiness)

---

## 5. Kubernetes deployment

Example Deployment for a simple runtime (bun):

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: functionfly-bun
spec:
  replicas: 3
  selector:
    matchLabels:
      app: functionfly-bun
  template:
    metadata:
      labels:
        app: functionfly-bun
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 10001
        runAsGroup: 10001
        seccompProfile:
          type: RuntimeDefault
      containers:
        - name: runtime
          image: functionfly/bun-runtime:1.0.0
          args: ["--port", "8091"]
          env:
            - name: ENVIRONMENT
              value: "production"
            - name: RUNTIME_API_TOKEN
              valueFrom:
                secretKeyRef:
                  name: functionfly-secrets
                  key: bun-api-token
            - name: NATS_URL
              value: "nats://nats.functionfly:4222"
          ports:
            - containerPort: 8091
          resources:
            requests:
              cpu: "500m"
              memory: "512Mi"
            limits:
              cpu: "2"
              memory: "1Gi"
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop: ["ALL"]
          livenessProbe:
            httpGet:
              path: /health
              port: 8091
            initialDelaySeconds: 10
            periodSeconds: 30
          readinessProbe:
            httpGet:
              path: /ready
              port: 8091
            initialDelaySeconds: 5
            periodSeconds: 10
          volumeMounts:
            - name: tmp
              mountPath: /tmp
      volumes:
        - name: tmp
          emptyDir: {}
```

For `microvm`, add:

```yaml
securityContext:
  privileged: false # we use specific caps instead
  capabilities:
    add: ["KVM", "NET_ADMIN"]
volumes:
  - name: kvm
    hostPath:
      path: /dev/kvm
  - name: vmimages
    hostPath:
      path: /var/lib/functionfly/vmimages
volumeMounts:
  - name: kvm
    mountPath: /dev/kvm
  - name: vmimages
    mountPath: /var/lib/functionfly/vmimages
```

---

## 6. Observability

All runtimes expose:

- `GET /health` — liveness probe (always 200 if process is up)
- `GET /ready` — readiness probe (returns orchestrator registration status
  and NATS connectivity)
- `GET /metrics` — Prometheus text exposition format
- Structured logs via `tracing` (JSON output with `RUST_LOG=info`)

Recommended alerts:

- `up == 0` for any runtime's `/health` endpoint
- `rate(http_requests_total{status="500"}[5m]) > 0.1`
- `container_memory_usage_bytes / container_spec_memory_limit_bytes > 0.85`
- NATS heartbeat failures (`sar_local_heartbeat_failures_total > 0`)

---

## 7. Incident response

If a runtime is compromised:

1. **Revoke the API token** in the orchestrator immediately.
2. **Stop the container**: `docker stop` / `kubectl delete pod` — the runtime
   shuts down within 10s (tini forwards SIGTERM, axum graceful shutdown).
3. **Capture forensics**: if you have pcap or audit logging, isolate the
   runtime's logs and memory dump.
4. **Rotate secrets**: NATS credentials, NATS TLS, RUNTIME_API_TOKEN.
5. **Re-deploy** with the patched image.
6. **Audit access logs**: every `/execute` request is logged with tenant_id
   and execution_id; cross-reference against the orchestrator's request log.

For security issues, email security@functionfly.io (per
[runtimes/sar-local/SECURITY.md](sar-local/SECURITY.md)). We acknowledge
within 48 hours and provide an initial assessment within 7 days.
