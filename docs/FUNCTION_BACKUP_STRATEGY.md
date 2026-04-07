# FunctionFly Function Backup Strategy & Disaster Recovery Runbook

## Overview

This document describes the comprehensive backup strategy and disaster recovery procedures for FunctionFly's function hosting infrastructure.

**Architecture Summary:**

- **Function Code**: Stored in PostgreSQL (primary + replicas)
- **Deployment Artifacts**: Redis (fast) + R2 (durable, cross-region)
- **Backup Frequency**: Continuous (streaming replication) + Daily snapshots to R2
- **RTO (Recovery Time Objective)**: 5 minutes for region failover, 1 hour for full disaster
- **RPO (Recovery Point Objective)**: 1 hour (streaming replication lag)

## Backup Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        PRIMARY REGION (US-East)                  │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐     │
│  │  PostgreSQL  │───▶│  PostgreSQL  │───▶│  PostgreSQL  │     │
│  │   Primary    │    │ Read Replica │    │ Read Replica │     │
│  │  (Writes)    │    │  (EU West)   │    │ (APAC SE)    │     │
│  └──────┬───────┘    └──────────────┘    └──────────────┘     │
│         │                                                        │
│         ▼                                                        │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐     │
│  │    Redis     │    │     R2       │    │   R2 EU      │     │
│  │  (Hot Cache) │───▶│  (Primary)   │───▶│  (Replica)   │     │
│  └──────────────┘    └──────────────┘    └──────────────┘     │
│         │                              │                       │
│         │    Async Archival            │    Auto Replication  │
│         │                              │                       │
└─────────┼────────────────────────────┼───────────────────────┘
          │                              │
          │                         ┌────▼──────┐
          │                         │  R2 APAC  │
          │                         │ (Replica) │
          │                         └───────────┘
          │
          │   Daily pg_dump to R2
          ▼
   ┌──────────────┐
   │   R2 Backups │
   │  (Versioned) │
   │ 30-day reten.│
   └──────────────┘
```

## Backup Tiers

### Tier 1: Real-Time Streaming Replication

- **Method**: PostgreSQL streaming replication
- **Frequency**: Real-time (synchronous or asynchronous)
- **Target**: EU and APAC read replicas
- **Purpose**: Near-zero RPO, instant read scaling
- **RPO**: < 1 second (synchronous), < 30 seconds (asynchronous)

### Tier 2: Continuous Artifact Archival

- **Method**: R2 object storage with cross-region replication
- **Trigger**: On every deployment (async, non-blocking)
- **Target**: R2 Primary (US) → R2 EU → R2 APAC
- **Purpose**: Durable artifact storage with global access
- **RPO**: < 5 minutes (artifact archival)

### Tier 3: Daily Point-in-Time Snapshots

- **Method**: `pg_dump` to R2 with compression
- **Frequency**: Daily at 02:00 UTC
- **Retention**: 30 days with lifecycle rules
- **Encryption**: AES-256-GCM (optional GPG)
- **Purpose**: Point-in-time recovery, corruption protection

### Tier 4: Edge Cache Warming

- **Method**: Cross-region cache warming on deploy/popularity
- **Trigger**: Deploy events + scheduled (10 min intervals)
- **Target**: Cloudflare Workers, Vercel Edge, Fly.io Edge
- **Purpose**: Function availability even during outages

## Environment Configuration

### Required Environment Variables

#### R2 Storage Configuration

```bash
# Primary R2 credentials
R2_ACCOUNT_ID=your-account-id
R2_ACCESS_KEY_ID=your-access-key
R2_SECRET_ACCESS_KEY=your-secret-key

# Artifact storage
R2_ARTIFACT_BUCKET_PRIMARY=functionfly-artifacts-prod
R2_ARTIFACT_BUCKET_REPLICA_1=functionfly-artifacts-eu
R2_ARTIFACT_REGION_REPLICA_1=eu-west-1
R2_ARTIFACT_BUCKET_REPLICA_2=functionfly-artifacts-apac
R2_ARTIFACT_REGION_REPLICA_2=ap-southeast-1

# Enable auto-archival from Redis to R2
R2_AUTO_ARCHIVE_ENABLED=true
R2_ARCHIVE_SYNC=false  # Async for performance

# Backup storage
R2_BACKUP_BUCKET=functionfly-backups
```

#### PostgreSQL Read Replica Configuration

```bash
# Enable read replicas
DB_READ_REPLICA_ENABLED=true
DB_HEALTH_CHECK_INTERVAL=30s
DB_HEALTH_CHECK_TIMEOUT=5s
DB_MAX_HEALTH_FAILURES=3

# EU Replica
DB_REPLICA_1_HOST=postgres-replica-eu.yourdomain.com
DB_REPLICA_1_PORT=5432
DB_REPLICA_1_WEIGHT=30
DB_REPLICA_1_PRIORITY=1
DB_REPLICA_1_REGION=eu-west-1

# APAC Replica
DB_REPLICA_2_HOST=postgres-replica-apac.yourdomain.com
DB_REPLICA_2_PORT=5432
DB_REPLICA_2_WEIGHT=20
DB_REPLICA_2_PRIORITY=2
DB_REPLICA_2_REGION=ap-southeast-1
```

#### Cross-Region Warming Configuration

```bash
# Enable cache warming
CACHE_WARMING_ENABLED=true
CACHE_WARMING_INTERVAL=10m
CACHE_WARMING_MIN_POPULARITY=100
CACHE_WARMING_TIMEOUT=30s

# Target regions
CACHE_WARMING_REGION_1=us-east-1
CACHE_WARMING_REGION_2=eu-west-1
CACHE_WARMING_REGION_3=ap-southeast-1
```

## Disaster Recovery Procedures

### Quick Decision Matrix

| Scenario | Detection | First Action | RTO | Runbook Section |
|----------|-----------|-------------|-----|-----------------|
| Single replica down | Health check alert | Restart replica service | 5 min | 3.1 |
| All replicas down | Database alert | Primary serves reads | 15 min | 3.2 |
| Primary database down | Critical DB alert | Failover to replica | 15 min | 3.3 |
| Redis cache loss | Cache miss spike | Rebuild from R2 | 10 min | 3.4 |
| Region outage (US East) | Multi-service down | Activate EU standby | 5 min | 3.5 |
| Complete infrastructure loss | Disaster detection | Restore from R2 to new region | 1 hour | 3.6 |

### 3.1 Single Replica Down

**Detection**: Prometheus alert `FunctionFlyDBReplicaUnhealthy`

**Impact**: Reduced read capacity, higher latency for affected region

**Procedure**:

1. **Verify replica status**:

   ```bash
   # Check replica lag
   psql -h $DB_REPLICA_1_HOST -c "SELECT EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp())) AS lag_seconds;"
   
   # Check replication status
   psql -h $DB_PRIMARY_HOST -c "SELECT * FROM pg_stat_replication;"
   ```

2. **Restart replica**:

   ```bash
   # Kubernetes
   kubectl rollout restart statefulset/postgres-replica-eu -n functionfly
   
   # Or fly.io machine
   fly machine restart --app your-postgres-app
   ```

3. **Verify recovery**:

   ```bash
   # Wait for lag to drop
   watch -n 5 'psql -h $DB_REPLICA_1_HOST -c "SELECT pg_is_in_recovery(), pg_last_xact_replay_timestamp();"'
   ```

4. **Verify application**:
   Check that read queries are being distributed (see monitoring dashboard)

### 3.2 All Replicas Down

**Detection**: Prometheus alert `FunctionFlyDBAllReplicasDown`

**Impact**: Primary serving all traffic, increased load, no geographic distribution

**Procedure**:

1. **Immediate**: Primary automatically serves all traffic (circuit breakers open)

2. **Check replica connectivity**:

   ```bash
   # Test from application pod
   kubectl exec -it deploy/orchestrator -- nc -zv $DB_REPLICA_1_HOST 5432
   kubectl exec -it deploy/orchestrator -- nc -zv $DB_REPLICA_2_HOST 5432
   ```

3. **Check network connectivity**:

   ```bash
   # Verify VPC peering, firewall rules
   kubectl exec -it deploy/orchestrator -- traceroute $DB_REPLICA_1_HOST
   ```

4. **Restart all replicas** (if network issue resolved):

   ```bash
   kubectl rollout restart statefulset/postgres-replica-eu postgres-replica-apac -n functionfly
   ```

5. **Monitor recovery**:
   Watch replication lag in Grafana dashboard. Expect 5-10 minutes for full sync.

### 3.3 Primary Database Down

**Detection**: Prometheus alert `FunctionFlyDBPrimaryUnhealthy` + application errors

**Impact**: Write operations failing, potential data loss if no failover

**Procedure**:

1. **Immediate actions** (first 2 minutes):
   - Page on-call database engineer
   - Announce incident in #incidents Slack channel
   - Start incident log

2. **Assess if primary will recover** (minutes 2-5):

   ```bash
   # Check primary status
   kubectl get pods -n functionfly -l app=postgres,role=primary
   kubectl describe pod postgres-primary-0 -n functionfly
   
   # Check logs
   kubectl logs postgres-primary-0 -n functionfly --tail=100
   ```

3. **Decision point** (minute 5):
   - If primary shows signs of recovery: Wait (max 10 min)
   - If primary is dead: Initiate failover

4. **Initiate failover** (if needed):

   ```bash
   # Promote EU replica to primary
   kubectl exec postgres-replica-eu-0 -n functionfly -- pg_ctl promote -D /var/lib/postgresql/data
   
   # Update application configuration
   kubectl patch configmap postgres-config -n functionfly --patch '{"data":{"DB_HOST":"postgres-replica-eu"}}'
   
   # Rolling restart of orchestrator to pick up new primary
   kubectl rollout restart deployment/orchestrator -n functionfly
   ```

5. **Post-failover**:
   - Verify writes are succeeding
   - Monitor replication lag (former primary will be behind when it comes back)
   - Plan to rebuild former primary as replica

### 3.4 Redis Cache Loss

**Detection**: Redis memory metrics + cache miss spike

**Impact**: Slower function loads (R2 fallback), but no data loss

**Procedure**:

1. **Verify R2 accessibility**:

   ```bash
   # Test R2 from application
   kubectl exec -it deploy/orchestrator -- curl -I \
     "https://${R2_ACCOUNT_ID}.r2.cloudflarestorage.com/${R2_ARTIFACT_BUCKET_PRIMARY}"
   ```

2. **Rebuild Redis from R2**:
   The application automatically falls back to R2 when Redis misses. To pre-warm:

   ```bash
   # Run warming script
   go run ./cmd/function-backup --warm-cache-only
   
   # Or manually warm popular functions
   for func in $(cat /tmp/popular_functions.txt); do
     kubectl exec -it deploy/orchestrator -- curl "http://localhost:8080/v1/admin/cache/warm/$func"
   done
   ```

3. **Monitor recovery**:
   - Cache hit ratio should recover to >90% within 30 minutes
   - R2 egress costs will spike during rebuild (expected)

### 3.5 Region Outage (US East)

**Detection**: Multi-service alerts + health check failures + potential Cloudflare notifications

**Impact**: Complete US-East unavailability

**Procedure**:

**Phase 1: Detection & Decision (0-2 minutes)**

1. Verify regional outage (not just your services):
   - Check AWS/Fly.io status pages
   - Check Cloudflare status

2. Confirm EU standby is healthy:

   ```bash
   curl https://eu.functionfly.com/healthz
   fly status --app functionfly-control-eu
   ```

**Phase 2: Failover Activation (2-5 minutes)**

1. Activate EU standby:

   ```bash
   # Scale up EU orchestrator
   fly scale count 3 --app functionfly-control-eu
   
   # Verify it comes online
   fly status --app functionfly-control-eu --watch
   ```

2. Update DNS / Load Balancer:

   ```bash
   # Option A: Cloudflare Load Balancer (automatic if health checks configured)
   # No action needed - LB detects primary failure and routes to EU
   
   # Option B: Manual DNS update
   # Update Cloudflare DNS A record:
   # api.functionfly.com → EU load balancer IP
   
   # Option C: Using fly.io's built-in geo-routing
   fly regions set iad=0 lax=0 fra=100 --app functionfly-control
   ```

3. Verify traffic flow:

   ```bash
   # Monitor EU logs
   fly logs --app functionfly-control-eu
   
   # Test function execution
   curl "https://eu.functionfly.com/v1/registry/test/hello"
   ```

**Phase 3: Recovery Preparation (ongoing)**

1. EU orchestrator is now PRIMARY:
   - Writes go to EU (until US recovers)
   - Read replicas in APAC still serve reads
   - R2 continues global artifact storage

2. When US-East recovers:
   - US primary will catch up via streaming replication
   - Consider switching back or keeping EU as new primary
   - If keeping EU primary: Update all config, restart US as replica

### 3.6 Complete Infrastructure Loss (Disaster)

**Detection**: All regions down, potentially datacenter-wide failure

**Impact**: Complete service unavailability, potential data loss

**Procedure**:

**Phase 1: Emergency Response (0-30 minutes)**

1. **Assemble response team**:
   - Incident commander
   - Database engineer
   - Infrastructure engineer
   - Communications lead

2. **Select recovery region** (pick from available):
   - EU-West (Frankfurt)
   - US-West (Los Angeles)
   - APAC-Southeast (Singapore)

   Decision criteria:
   - Which region has best connectivity?
   - Where is most recent replica?
   - Cost/performance considerations

**Phase 2: Infrastructure Rebuild (30-60 minutes)**

1. **Restore PostgreSQL from backup**:

   ```bash
   # Download latest backup from R2
   rclone copy ":s3,provider=Cloudflare:${R2_BACKUP_BUCKET}/backups/functions/$(date +%Y/%m/%d)/" /tmp/backup/
   
   # Extract and restore
   gunzip < /tmp/backup/functions-backup-*.sql.gz | psql -h new-primary-host
   ```

2. **Deploy new orchestrator**:

   ```bash
   # Deploy to selected region
   fly deploy --config deploy/fly/functionfly-control/fly.toml --region fra
   
   # Or use Terraform/Kubernetes
   terraform apply -var="primary_region=eu-west-1"
   ```

3. **Restore artifacts**:

   ```bash
   # Artifacts are in R2, will be retrieved on-demand
   # Pre-warm critical functions
   go run ./cmd/function-backup --warm-cache-only
   ```

4. **Verify restoration**:
   - All functions loadable
   - Database connections working
   - Health checks passing

**Phase 3: Service Restoration (60+ minutes)**

1. **Gradual traffic restoration**:
   - Start with 10% traffic via Cloudflare Load Balancer
   - Monitor error rates
   - Scale to 100% over 30 minutes

2. **Post-incident**:
   - Root cause analysis
   - Update runbook with learnings
   - Schedule follow-up DR test

## Testing Procedures

### Quarterly DR Test Schedule

**Test 1: Replica Failover (Month 1)**

```bash
# Simulate replica failure
kubectl exec postgres-replica-eu-0 -n functionfly -- pkill -9 postgres

# Verify primary takes over reads
# Verify lag metrics
# Restart replica and verify sync
```

**Test 2: Region Failover (Month 4)**

```bash
# Update fly.toml to simulate primary region down
fly regions set iad=0 fra=100 --app functionfly-control

# Verify EU receives traffic
# Verify R2 artifact access works
# Run smoke tests

# Restore
fly regions set iad=100 fra=0 --app functionfly-control
```

**Test 3: Backup Restore (Month 7)**

```bash
# Create test restore environment
# Download yesterday's backup
./scripts/backup-functions-to-r2.sh --restore-to=/tmp/restore-test

# Verify restored data integrity
# Run integration tests against restored DB
```

**Test 4: Full Disaster Simulation (Month 10)**

- Complete infrastructure teardown in staging
- Follow 3.6 procedure
- Measure actual RTO/RPO
- Document deviations from expected

### Automated Testing

**Daily**:

- Backup job runs and uploads to R2 (verified by checksum)
- Replication lag monitoring (< 30s alert threshold)

**Weekly**:

- Restore test from backup (staging environment)
- Artifact retrieval from R2 (random samples)

**Monthly**:

- Full DR procedure walkthrough (tabletop exercise)
- RTO/RPO measurement and reporting

## Monitoring & Alerting

### Key Metrics

| Metric | Target | Alert Threshold | Dashboard |
|--------|--------|-----------------|-----------|
| DB Replication Lag | < 5s | > 30s | PostgreSQL Overview |
| R2 Artifact Upload | 100% | < 95% | Storage Health |
| Backup Success | 100% | Failed | Backup Status |
| Cache Warming Success | > 99% | < 95% | Cache Performance |
| Cross-Region Health | 100% | < 90% | Multi-Region Status |
| Measured RTO | < 5 min | > 10 min | DR Readiness |
| Measured RPO | < 1 hour | > 2 hours | DR Readiness |

### Alert Routing

```yaml
# PagerDuty routing
Critical alerts (Primary DB down, Complete region loss):
  - Page on-call immediately
  - Escalate to infrastructure lead after 5 min

Warning alerts (Replica lag, Backup delays):
  - Slack notification #infrastructure
  - Create Jira ticket
  - Review in next standup
```

## Cost Optimization

### Backup Storage Costs (Estimated)

| Component | Size | Cost/Month |
|-----------|------|------------|
| R2 Artifact Storage | 100 GB | ~$5 |
| R2 Backups (30 days) | 50 GB | ~$2.50 |
| R2 Operations | 1M ops | ~$10 |
| Cross-region DB replica | - | ~$50-100 |
| EU standby (scaled 0) | - | ~$0-5 |
| EU standby (active) | - | ~$20-40 |
| **Total standby cost** | - | **~$85-155/month** |

### Cost Reduction Strategies

1. **EU Standby scaling**:
   - Scale to 0 when not needed (configured in fly.toml)
   - Activates in 30-60 seconds on first request

2. **R2 lifecycle rules**:
   - Delete artifacts older than 90 days (or archive to Glacier)
   - Compress backups more aggressively

3. **Selective replication**:
   - Only replicate "production" tier artifacts
   - Development artifacts stay in single region

## Security Considerations

### Data Protection

1. **Encryption at rest**:
   - R2: AES-256 (Cloudflare managed)
   - PostgreSQL: Disk encryption enabled
   - Backups: Optional GPG encryption

2. **Encryption in transit**:
   - All replication: TLS 1.3
   - R2 access: HTTPS only
   - Inter-service: mTLS where supported

3. **Access control**:
   - R2: IAM policies limiting to specific buckets
   - PostgreSQL: Role-based access
   - Backups: Dedicated backup service account

### Compliance

- **GDPR**: Backups encrypted, EU region for EU data
- **SOC 2**: Quarterly DR tests, documented procedures
- **HIPAA** (if applicable): BAAs with Cloudflare, encrypted backups

## Troubleshooting

### Common Issues

**Issue: R2 upload failures**

```bash
# Check R2 credentials
env | grep R2_

# Test connectivity
curl -I "https://${R2_ACCOUNT_ID}.r2.cloudflarestorage.com"

# Check bucket exists
aws s3 ls s3://${R2_ARTIFACT_BUCKET_PRIMARY} --endpoint-url=https://${R2_ACCOUNT_ID}.r2.cloudflarestorage.com
```

**Issue: Replication lag > 60s**

```bash
# Check network
ping $DB_REPLICA_1_HOST

# Check replication slot
psql -h $DB_PRIMARY_HOST -c "SELECT * FROM pg_replication_slots;"

# Check lag on replica
psql -h $DB_REPLICA_1_HOST -c "SELECT * FROM pg_stat_wal_receiver;"
```

**Issue: Cache warming queue backlog**

```bash
# Check warming workers
redis-cli LLEN cache:warming:queue

# Restart warming service
kubectl rollout restart deployment/orchestrator

# Manual warm for critical functions
curl -X POST "https://api.functionfly.com/v1/admin/cache/warm" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"functions":["critical-function-1","critical-function-2"]}'
```

## Related Documentation

- [Disaster Recovery](DISASTER_RECOVERY.md) - General DR procedures
- [Multi-Region Kubernetes](deploy/kubernetes/postgres-multi-region.yaml) - K8s manifests
- [EU Standby Deployment](deploy/fly/functionfly-eu-standby/README.md) - Standby orchestrator
- [Monitoring Setup](deploy/monitoring/) - Prometheus/Grafana configuration
- [R2 Artifact Store](internal/deployment/r2_store.go) - Implementation details
- [Cross-Region Warming](internal/cache/cross_region_warming.go) - Cache warming code

## Contact

- **Infrastructure Team**: <infrastructure@functionfly.io>
- **On-Call Escalation**: PagerDuty rotation
- **Emergency Hotline**: +1-XXX-XXX-XXXX

---

*Last updated: 2024-XX-XX*
*Next review: Quarterly*
*Tested in: Staging environment, 2024-XX-XX*
