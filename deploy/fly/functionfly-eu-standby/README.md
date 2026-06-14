# FunctionFly EU Standby Orchestrator

This directory contains the deployment configuration for the EU (Frankfurt) standby orchestrator, providing multi-region redundancy and disaster recovery capabilities.

## Purpose

The EU standby orchestrator serves as a **warm standby** that can be activated in case of:

1. **Primary region failure** - US East (IAD) becomes unavailable
2. **Disaster recovery** - Complete primary infrastructure failure
3. **EU traffic optimization** - Route EU users to closer infrastructure
4. **Maintenance window** - Failover during primary maintenance

## Architecture

```
┌─────────────────┐         ┌─────────────────┐
│  Primary (IAD)  │         │  Standby (FRA)  │
│  ┌───────────┐  │         │  ┌───────────┐  │
│  │ Primary   │◄─┼─────────┼─┤ Read      │  │
│  │   PG      │  │Replica  │ │ Replica   │  │
│  └─────┬─────┘  │         │ └─────┬─────┘  │
│        │         │         │       │         │
│  ┌─────▼─────┐  │         │ ┌─────▼─────┐  │
│  │ Primary   │  │         │ │ Secondary │  │
│  │Orchestrator│  │         │ │Orchestrator│  │
│  └─────┬─────┘  │         │ └─────┬─────┘  │
└────────┼────────┘         └───────┼────────┘
         │                          │
         └──────────┬───────────────┘
                    │
              ┌─────▼─────┐
              │  R2 Storage │
              │(Artifacts)  │
              └─────────────┘
```

## Quick Start

### Deploy the Standby

```bash
# Deploy the EU standby orchestrator
fly deploy --config deploy/fly/functionfly-eu-standby/fly.toml

# Verify deployment
fly status --app functionfly-control-eu

# Check logs
fly logs --app functionfly-control-eu
```

### Activate Standby (Failover)

In case of primary failure:

```bash
# Option 1: Scale standby up to handle traffic
fly scale count 3 --app functionfly-control-eu

# Option 2: Update Cloudflare DNS to point to standby
# (Requires manual DNS update or automated failover)

# Option 3: Use Cloudflare Load Balancer with health checks
# (Automatically routes to healthy region)
```

### Deactivate Standby

After primary recovery:

```bash
# Scale standby back down
fly scale count 0 --app functionfly-control-eu

# Verify primary is healthy
fly status --app functionfly-control
```

## Cost Optimization

The standby is configured with `min_machines_running = 0`, meaning:

- **Idle cost**: ~$0 (machines stop after inactivity)
- **Active cost**: ~$5-10/month when serving traffic
- **Failover time**: 30-60 seconds to start machines

For faster failover (but higher cost):

```toml
min_machines_running = 1  # Always keep 1 machine running
```

## Database Connectivity

The standby uses read replicas for optimal EU performance:

1. **Read operations**: Go to EU PostgreSQL replica (low latency)
2. **Write operations**: Forwarded to US primary (acceptable latency for writes)
3. **Replication lag**: Monitored via health checks

### Read Replica Configuration

Set via environment variables in `fly.toml`:

```toml
DB_READ_REPLICA_ENABLED = "true"
DB_REPLICA_1_HOST = "postgres-replica-eu.fly.dev"
DB_REPLICA_1_WEIGHT = "100"  # High priority for EU
DB_REPLICA_1_PRIORITY = "1"
DB_REPLICA_1_REGION = "eu-west-1"
```

## R2 Cross-Region Access

Artifacts are stored in Cloudflare R2, accessible from any region:

- **No egress fees**: R2 doesn't charge for bandwidth
- **Global access**: EU, US, and APAC can access same bucket
- **Automatic replication**: Configure R2 for multi-region replication

## Monitoring

The standby exports Prometheus metrics:

```bash
# Access metrics
fly ssh console --app functionfly-control-eu
curl localhost:9090/metrics
```

Key metrics to watch:

- `functionfly_db_replica_lag_seconds` - Replication lag
- `functionfly_health_check_status` - Health check results
- `functionfly_r2_operations_total` - R2 operation counts

## Testing Failover

Test the standby without affecting production:

```bash
# 1. Deploy standby
fly deploy --config deploy/fly/functionfly-eu-standby/fly.toml

# 2. Start standby machines
fly scale count 1 --app functionfly-control-eu

# 3. Test health endpoint
STANDBY_URL=$(fly status --app functionfly-control-eu --json | jq -r '.Hostname')
curl "https://${STANDBY_URL}/healthz"

# 4. Test function execution (if functions deployed)
curl "https://${STANDBY_URL}/v1/registry/test/hello"

# 5. Scale back down
fly scale count 0 --app functionfly-control-eu
```

## Troubleshooting

### Standby Won't Start

```bash
# Check machine status
fly machines list --app functionfly-control-eu

# Check recent logs
fly logs --app functionfly-control-eu --recent

# Verify database connectivity
fly ssh console --app functionfly-control-eu
# Inside container:
psql $DATABASE_URL -c "SELECT 1"
```

### High Replication Lag

```bash
# Check replication lag from standby
fly ssh console --app functionfly-control-eu
# Inside container:
psql $DATABASE_URL -c "SELECT EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp())) AS lag_seconds;"
```

### R2 Access Issues

```bash
# Test R2 connectivity
fly ssh console --app functionfly-control-eu
# Inside container:
curl -I "https://${R2_ACCOUNT_ID}.r2.cloudflarestorage.com/${R2_BUCKET}"
```

## Security Considerations

1. **Network isolation**: Standby uses separate Fly.io app for isolation
2. **Credential rotation**: Same credentials as primary (rotate together)
3. **VPC connectivity**: Use Fly.io private networking or WireGuard
4. **Backup access**: Ensure R2 access from standby region

## Maintenance Procedures

### Updating Standby

```bash
# Deploy new version
fly deploy --config deploy/fly/functionfly-eu-standby/fly.toml

# Verify rollout
fly status --app functionfly-control-eu
```

### Database Failover

If EU replica fails:

```bash
# 1. Switch standby to use primary directly
fly secrets set DB_READ_REPLICA_ENABLED=false --app functionfly-control-eu

# 2. Restart to apply
fly deploy --config deploy/fly/functionfly-eu-standby/fly.toml

# 3. Fix or recreate replica
# (Follow PostgreSQL replica recovery procedures)
```

## See Also

- [Primary Orchestrator](../functionfly-control/fly.toml)
- [Disaster Recovery Runbook](../../docs/DISASTER_RECOVERY.md)
- [Database Multi-Region](../../deploy/kubernetes/postgres-multi-region.yaml)
