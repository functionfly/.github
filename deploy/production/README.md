# FunctionFly Production Deployment

## Overview

This directory contains production-ready deployment configurations for FunctionFly, including:
- Kubernetes manifests
- Docker Compose for local development
- Terraform for AWS/GCP/Azure infrastructure
- Health checks and monitoring
- SSL/TLS configurations

## Quick Start

### Local Development (Docker Compose)

```bash
# Start all services
cd deploy/production
docker-compose -f docker-compose.yml up -d

# View logs
docker-compose logs -f orchestrator-api

# Scale workers
docker-compose up -d --scale worker=3
```

### Production Kubernetes

```bash
# Apply all manifests
kubectl apply -k k8s/

# Check deployment status
kubectl get pods -n functionfly
kubectl get svc -n functionfly
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Load Balancer                          │
│                    (nginx / AWS ALB / traefik)                │
└──────────────────────┬────────────────────────────────────────┘
                       │
       ┌───────────────┴──────────────────┐
       │                                  │
┌──────▼───────┐                ┌────────▼────────┐
│   API Pods   │                │   Worker Pods   │
│  (orchestrator)              │   (graph execution)│
└──────┬───────┘                └────────┬────────┘
       │                                │
       └──────────┬───────────────────────┘
                  │
         ┌────────▼────────┐
         │    PostgreSQL   │
         │   (State + Data) │
         └─────────────────┘
                  │
         ┌────────▼────────┐
         │     Redis       │
         │  (Cache + Queue)│
         └─────────────────┘
                  │
         ┌────────▼────────┐      ┌───────────────┐
         │      NATS       │◄────►│   Prometheus  │
         │  (Event Bus)    │      │  (Monitoring) │
         └─────────────────┘      └───────────────┘
```

## Services

### Orchestrator API

- **Image**: `functionfly/orchestrator-api:latest`
- **Port**: 8080
- **Features**:
  - REST/GraphQL API
  - Graph execution
  - DRE certificate generation
  - Auto-generated endpoints

### Workers

- **Image**: `functionfly/worker:latest`
- **Scales**: Horizontally (HPA)
- **Features**:
  - Async graph execution
  - Sandbox function execution
  - DRE verification

### PostgreSQL

- **Type**: PostgreSQL 15+ with pgvector
- **Features**:
  - State fabric
  - Graph definitions
  - DRE certificates (MEG, FXCERT)
  - Event sourcing

### Redis

- **Use**: Caching, pub/sub, rate limiting
- **Features**:
  - Graph trigger cache
  - Session storage
  - Queue management

### NATS (Optional)

- **Use**: Event streaming for reactive graphs
- **Features**:
  - JetStream for persistence
  - Pub/sub for real-time updates

## Configuration

### Environment Variables

```bash
# Core
DATABASE_URL=postgres://user:pass@localhost:5432/functionfly
REDIS_URL=redis://localhost:6379
NATS_URL=nats://localhost:4222  # Optional
API_PORT=8080

# DRE (Optional but recommended)
DRE_NODE_ID=node-1
DRE_REGION=us-east-1
DRE_NODE_KEY_PATH=/secrets/node.key
DRE_PLATFORM_KEY_PATH=/secrets/platform.key

# Security
JWT_SECRET=your-secret
ENCRYPTED_STATE_ENABLED=true
STATE_ENCRYPTION_KEY=base64-encoded-key

# AI Service
AI_SERVICE_URL=http://ai-service:8081
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...

# Observability
PROMETHEUS_ENABLED=true
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
LOG_LEVEL=info
```

## Health Checks

### API Health Endpoint

```bash
# Liveness check
curl http://localhost:8080/health/live

# Readiness check (includes DB, Redis connectivity)
curl http://localhost:8080/health/ready

# Detailed health with all dependencies
curl http://localhost:8080/health
```

### Expected Responses

```json
// /health/live
{"status": "alive", "timestamp": "2024-01-01T00:00:00Z"}

// /health/ready
{
  "status": "ready",
  "checks": {
    "database": {"status": "up", "latency_ms": 5},
    "redis": {"status": "up", "latency_ms": 2},
    "nats": {"status": "up", "latency_ms": 3}
  }
}
```

## Monitoring

### Prometheus Metrics

Available at `:8080/metrics`:

```
# Graph execution metrics
functionfly_graph_executions_total{status="success"}
functionfly_graph_execution_duration_seconds
functionfly_graph_node_executions_total
functionfly_graph_node_execution_duration_seconds

# Function metrics
functionfly_function_invocations_total
functionfly_function_execution_duration_seconds
functionfly_function_memory_bytes

# State fabric metrics
functionfly_state_fabric_operations_total
functionfly_state_fabric_operation_duration_seconds

# DRE metrics
functionfly_dre_certificates_generated_total
functionfly_dre_verifications_total
functionfly_dre_trust_score
```

### Alerts

```yaml
# Example Prometheus alerts
groups:
  - name: functionfly
    rules:
      - alert: HighErrorRate
        expr: rate(functionfly_graph_executions_total{status="error"}[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: High graph execution error rate

      - alert: DRECertificateGenerationFailure
        expr: rate(functionfly_dre_certificates_generated_total[1h]) == 0
        for: 10m
        labels:
          severity: info
        annotations:
          summary: No DRE certificates generated recently
```

## Scaling

### Horizontal Pod Autoscaler (K8s)

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: orchestrator-api
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: orchestrator-api
  minReplicas: 2
  maxReplicas: 20
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Pods
      pods:
        metric:
          name: functionfly_graph_executions_per_second
        target:
          type: AverageValue
          averageValue: "100"
```

## Security

### SSL/TLS

- Auto-generated endpoints support HTTPS via Let's Encrypt
- DRE certificates use Ed25519 signatures
- State encryption via AES-256-GCM

### Network Policies

```yaml
# Allow API → Database
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-api-db
spec:
  podSelector:
    matchLabels:
      app: orchestrator-api
  policyTypes:
    - Egress
  egress:
    - to:
        - podSelector:
            matchLabels:
              app: postgres
      ports:
        - protocol: TCP
          port: 5432
```

## Backup & Recovery

### State Fabric Snapshots

```bash
# Create snapshot
POST /api/platform/state-fabrics/{id}/snapshots

# Restore from snapshot
POST /api/platform/state-fabrics/{id}/restore

# Automated daily backups via cron
0 2 * * * pg_dump functionfly > /backups/functionfly-$(date +%Y%m%d).sql
```

### DRE Certificate Archival

- FXCERTs stored in PostgreSQL with 7-year retention
- Optional export to S3 Glacier for compliance

## Troubleshooting

### Common Issues

1. **High Memory Usage in Workers**
   - Check sandbox timeouts
   - Review function memory limits
   - Scale workers horizontally

2. **Slow Graph Execution**
   - Check for missing indexes on state queries
   - Review Redis cache hit rate
   - Optimize parallel node execution

3. **NATS Connection Issues**
   - Verify NATS_URL environment variable
   - Check network policies
   - Review NATS server logs

### Debug Endpoints

```bash
# Get runtime info
curl http://localhost:8080/debug/pprof/heap

# Get active graph instances
curl http://localhost:8080/api/admin/frg/instances

# Get DRE statistics
curl http://localhost:8080/api/admin/dre/stats
```
