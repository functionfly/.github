# FunctionFly Registry Disaster Recovery Runbook

This runbook provides step-by-step procedures for recovering from various disaster scenarios affecting the FunctionFly Registry.

## Table of Contents

1. [Overview](#overview)
2. [Emergency Contacts](#emergency-contacts)
3. [Recovery Objectives](#recovery-objectives)
4. [Scenario 1: Application Failure](#scenario-1-application-failure)
5. [Scenario 2: Database Corruption](#scenario-2-database-corruption)
6. [Scenario 3: Neon Region Outage](#scenario-3-neon-region-outage)
7. [Scenario 4: Complete Infrastructure Loss](#scenario-4-complete-infrastructure-loss)
8. [Scenario 5: Security Breach](#scenario-5-security-breach)
9. [Post-Recovery Verification](#post-recovery-verification)
10. [Prevention Checklist](#prevention-checklist)

## Overview

### Infrastructure

| Component | Primary | Backup | Recovery Method |
|-----------|---------|--------|-----------------|
| Application | Fly.io (iad) | Fly.io (ams) | Automated redeploy |
| Database | Neon (us-east-1) | Read replica / PITR | Restore from backup |
| Storage | R2 (auto) | B2 (cross-region) | Download & restore |
| Cache | Upstash Redis | Rebuild from DB | Reconnect |

### Recovery Definitions

- **RTO (Recovery Time Objective)**: Maximum acceptable downtime
  - Application: 5 minutes
  - Database: 15 minutes
  - Complete system: 1 hour

- **RPO (Recovery Point Objective)**: Maximum acceptable data loss
  - Database: < 5 minutes (WAL archiving)
  - Function artifacts: 0 (immutable storage)
  - Session data: Acceptable loss (ephemeral)

## Emergency Contacts

| Role | Contact | Escalation |
|------|---------|------------|
| On-call Engineer | [Your PagerDuty/Slack] | Immediate |
| Fly.io Support | support@fly.io | After 30 min |
| Neon Support | support@neon.tech | After 30 min |
| Cloudflare Support | support.cloudflare.com | After 1 hour |

## Scenario 1: Application Failure

**Symptoms**: App crashes, health checks failing, 5xx errors

**RTO**: 5 minutes | **RPO**: 0

### Diagnosis

```bash
# Check app status
fly status --app functionfly-orchestrator

# View recent logs
fly logs --app functionfly-orchestrator --recent

# Check for panics
curl -s https://functionfly-orchestrator.fly.dev/health/detailed | jq .
```

### Recovery Steps

#### Option A: Automatic Recovery (Preferred)

```bash
# Fly.io auto-restart should handle this
# If not, manually restart
fly apps restart functionfly-orchestrator
```

#### Option B: Rollback to Previous Version

```bash
# List recent releases
fly releases list --app functionfly-orchestrator

# Rollback to previous working version
fly deploy --app functionfly-orchestrator \
  --image $(fly releases list --app functionfly-orchestrator | awk 'NR==3{print $2}')
```

#### Option C: Full Redeploy

```bash
# Redeploy from latest main
./scripts/deploy-fly-production.sh --skip-staging --force
```

### Verification

```bash
# Wait for health check
for i in {1..12}; do
  if curl -sf https://functionfly-orchestrator.fly.dev/health; then
    echo "✅ App is healthy"
    break
  fi
  echo "Waiting... ($i/12)"
  sleep 10
done

# Check metrics
fly metrics --app functionfly-orchestrator
```

## Scenario 2: Database Corruption

**Symptoms**: Query errors, data inconsistencies, neon errors in logs

**RTO**: 30 minutes | **RPO**: < 5 minutes

### Diagnosis

```bash
# Check database connectivity
fly ssh console --app functionfly-orchestrator --command "psql \$DATABASE_URL -c 'SELECT 1;'"

# Check for data corruption
psql "${DATABASE_URL}" -c "
  SELECT schemaname, tablename, attname, n_tup_ins, n_tup_upd, n_tup_del
  FROM pg_stat_user_tables
  ORDER BY n_tup_ins DESC;
"
```

### Recovery Steps

#### Step 1: Stop Application Writes

```bash
# Pause the app to prevent further corruption
fly scale count 0 --app functionfly-orchestrator
```

#### Step 2: Identify Last Good Backup

```bash
# List available backups
./scripts/backup-registry-fly.sh --list

# Or check R2 directly
aws s3 ls s3://functionfly-prod/backups/ \
  --endpoint-url https://YOUR_ACCOUNT_ID.r2.cloudflarestorage.com | tail -20
```

#### Step 3: Restore from Backup

```bash
# Download the backup
BACKUP_DATE="20240115_020000"
aws s3 cp \
  s3://functionfly-prod/backups/registry_${BACKUP_DATE}_functions.sql.gz \
  /tmp/restore.sql.gz \
  --endpoint-url https://YOUR_ACCOUNT_ID.r2.cloudflarestorage.com

# Verify checksum
sha256sum -c /tmp/restore.sql.gz.sha256

# Extract
gunzip /tmp/restore.sql.gz

# Create new database (if needed)
# Via Neon console or API
NEW_DB_URL="postgresql://...neon.tech/functionfly_restored"

# Restore
psql "${NEW_DB_URL}" < /tmp/restore.sql
```

#### Step 4: Point-in-Time Recovery (PITR) - Neon

If using Neon Pro with PITR:

```bash
# Neon supports point-in-time recovery via console or API
# Go to Neon Dashboard → Project → Branches → Create Branch
# Select "Point in Time" and choose recovery point

# Update DATABASE_URL after branch creation
fly secrets set DATABASE_URL="postgresql://...neon.tech/functionfly_restored" \
  --app functionfly-orchestrator
```

#### Step 5: Resume Application

```bash
# Update DATABASE_URL if using new database
fly secrets set DATABASE_URL="${NEW_DB_URL}" --app functionfly-orchestrator

# Scale back up
fly scale count 2 --app functionfly-orchestrator

# Verify health
curl -s https://functionfly-orchestrator.fly.dev/health | jq .
```

## Scenario 3: Neon Region Outage

**Symptoms**: Database connection timeouts, Neon status page shows outage

**RTO**: 15 minutes | **RPO**: < 1 minute

### Recovery Steps

#### Step 1: Failover to Read Replica

```bash
# Check if read replica is available
fly secrets list --app functionfly-orchestrator | grep READ_REPLICA

# If configured, switch to read replica
fly secrets set DATABASE_URL="${READ_REPLICA_URL}" \
  --app functionfly-orchestrator
```

#### Step 2: Promote Read Replica (If Needed)

```bash
# Neon read replicas can be promoted to primary
# Via Neon Dashboard or API:
# Branches → [Replica Branch] → Promote to Primary

# Update connection string
fly secrets set DATABASE_URL="${NEW_PRIMARY_URL}" \
  --app functionfly-orchestrator
```

#### Step 3: Deploy to New Region

```bash
# If entire region is affected, deploy to standby region
fly regions add ams --app functionfly-orchestrator
fly scale count 0 --region iad
fly scale count 2 --region ams
```

### Verification

```bash
# Verify database connectivity
fly ssh console --app functionfly-orchestrator --command "pg_isready -d \$DATABASE_URL"

# Check data consistency
psql "${DATABASE_URL}" -c "SELECT COUNT(*) FROM registry_functions;"
```

## Scenario 4: Complete Infrastructure Loss

**Symptoms**: Multiple component failures, complete unavailability

**RTO**: 1 hour | **RPO**: < 1 hour

### Phase 1: Emergency Assessment (0-10 min)

```bash
# Check status of all components
echo "=== Fly.io Status ==="
fly status --app functionfly-orchestrator

echo "=== Neon Status ==="
curl -s https://status.neon.tech/api/v2/status.json | jq '.status.description'

echo "=== Cloudflare Status ==="
curl -s https://www.cloudflarestatus.com/api/v2/status.json | jq '.status.description'
```

### Phase 2: Database Recovery (10-30 min)

#### Option A: Restore from R2 Backup

```bash
# 1. Create new Neon database (via console)
# 2. Download latest backup
LATEST_BACKUP=$(aws s3 ls s3://functionfly-prod/backups/ \
  --endpoint-url https://YOUR_ACCOUNT_ID.r2.cloudflarestorage.com |
  grep "registry_.*functions.sql.gz" | sort | tail -1 | awk '{print $4}')

aws s3 cp "s3://functionfly-prod/backups/${LATEST_BACKUP}" \
  /tmp/recovery.sql.gz \
  --endpoint-url https://YOUR_ACCOUNT_ID.r2.cloudflarestorage.com

# 3. Restore
gunzip /tmp/recovery.sql.gz
psql "${NEW_DATABASE_URL}" < /tmp/recovery.sql
```

#### Option B: Restore from B2 Cross-Region Replica

```bash
# If R2 is unavailable, use B2
b2 authorize "${B2_KEY_ID}" "${B2_KEY}"
b2 download-file functionfly-backup-replica "${LATEST_BACKUP}" /tmp/recovery.sql.gz
```

### Phase 3: Application Recovery (30-45 min)

```bash
# 1. Create new Fly app
fly apps create functionfly-orchestrator-recovery

# 2. Set all secrets
fly secrets set --app functionfly-orchestrator-recovery \
  DATABASE_URL="${NEW_DATABASE_URL}" \
  REDIS_URL="${REDIS_URL}" \
  R2_ACCESS_KEY_ID="${R2_ACCESS_KEY_ID}" \
  R2_SECRET_ACCESS_KEY="${R2_SECRET_ACCESS_KEY}" \
  R2_ENDPOINT="${R2_ENDPOINT}" \
  JWT_SECRET="${JWT_SECRET}"

# 3. Deploy
fly deploy --config fly.toml --app functionfly-orchestrator-recovery --yes

# 4. Verify
fly status --app functionfly-orchestrator-recovery
```

### Phase 4: DNS Update (45-60 min)

```bash
# Update DNS to point to new app
# Cloudflare DNS:
# A record: functionfly-orchestrator.fly.dev → functionfly-orchestrator-recovery.fly.dev
# Or update CNAME if using custom domain
```

### Phase 5: Data Reconciliation (Post-Recovery)

```bash
# Check for any missing data since last backup
psql "${NEW_DATABASE_URL}" -c "
  SELECT 
    'registry_functions' as table_name,
    COUNT(*) as row_count,
    MAX(updated_at) as latest_update
  FROM registry_functions;
"

# If WAL archiving was enabled, replay any missing transactions
# (Neon handles this automatically if using PITR)
```

## Scenario 5: Security Breach

**Symptoms**: Unauthorized access, suspicious activity, leaked credentials

**RTO**: 30 minutes | **RPO**: 0 (intentional)

### Immediate Response

```bash
# 1. Revoke all sessions (if using Redis for sessions)
redis-cli -u "${REDIS_URL}" FLUSHDB

# 2. Rotate all secrets
fly secrets set --app functionfly-orchestrator \
  JWT_SECRET="$(openssl rand -hex 32)" \
  API_KEY_SECRET="$(openssl rand -hex 32)"

# 3. Update R2 credentials
# Via Cloudflare Dashboard → R2 → Manage API Tokens

# 4. Rotate database credentials
# Via Neon Dashboard → Connection → Reset Password
fly secrets set DATABASE_URL="${NEW_DATABASE_URL}" --app functionfly-orchestrator

# 5. Restart app
fly apps restart functionfly-orchestrator
```

### Audit and Investigation

```bash
# Review access logs
fly logs --app functionfly-orchestrator | grep -i "auth\|login\|unauthorized"

# Check for compromised functions
psql "${DATABASE_URL}" -c "
  SELECT id, author, name, updated_at
  FROM registry_functions
  WHERE updated_at > NOW() - INTERVAL '24 hours'
  ORDER BY updated_at DESC;
"

# Review function executions
psql "${DATABASE_URL}" -c "
  SELECT 
    function_id,
    COUNT(*) as execution_count,
    MAX(created_at) as last_execution
  FROM registry_function_executions
  WHERE created_at > NOW() - INTERVAL '24 hours'
  GROUP BY function_id
  ORDER BY execution_count DESC;
"
```

### Verification

```bash
# Ensure all secrets rotated
fly secrets list --app functionfly-orchestrator

# Verify no unauthorized access
curl -s https://functionfly-orchestrator.fly.dev/metrics | grep -i "auth_error"
```

## Post-Recovery Verification

### Health Checks

```bash
#!/bin/bash
# post-recovery-check.sh

BASE_URL="https://functionfly-orchestrator.fly.dev"

echo "=== Post-Recovery Verification ==="

# 1. Basic health
curl -sf "${BASE_URL}/health" && echo "✅ Health check passed" || echo "❌ Health check failed"

# 2. Detailed health
curl -sf "${BASE_URL}/health/detailed" | jq '.status'

# 3. Registry API
curl -sf "${BASE_URL}/api/v1/registry/functions?limit=1" && echo "✅ Registry API working" || echo "❌ Registry API failed"

# 4. Database connectivity
fly ssh console --app functionfly-orchestrator --command "psql \$DATABASE_URL -c 'SELECT 1;'"

# 5. Storage connectivity
fly ssh console --app functionfly-orchestrator --command "echo 'test' | aws s3 cp - s3://\$R2_BUCKET/test.txt --endpoint-url=\$R2_ENDPOINT"

echo "=== Verification Complete ==="
```

### Data Integrity Checks

```bash
# Verify function counts
psql "${DATABASE_URL}" -c "
  SELECT 
    (SELECT COUNT(*) FROM registry_functions) as functions,
    (SELECT COUNT(*) FROM registry_function_versions) as versions,
    (SELECT COUNT(*) FROM registry_function_executions WHERE created_at > NOW() - INTERVAL '1 hour') as recent_executions;
"

# Check for orphaned records
psql "${DATABASE_URL}" -c "
  SELECT 'orphaned_versions' as check_name,
    COUNT(*) as count
  FROM registry_function_versions rfv
  LEFT JOIN registry_functions rf ON rfv.function_id = rf.id
  WHERE rf.id IS NULL;
"
```

## Prevention Checklist

### Weekly

- [ ] Review backup logs (GitHub Actions)
- [ ] Verify backup restoration (monthly test restore)
- [ ] Check Fly.io app health
- [ ] Review security advisories

### Monthly

- [ ] Rotate non-critical secrets
- [ ] Review access logs for anomalies
- [ ] Test failover procedures (staging)
- [ ] Update runbook with lessons learned

### Quarterly

- [ ] Full DR drill (staging environment)
- [ ] Review and update RTO/RPO targets
- [ ] Audit all secrets and access controls
- [ ] Review and update monitoring thresholds

### Annually

- [ ] Complete infrastructure review
- [ ] Update emergency contacts
- [ ] Review vendor SLAs and support contracts
- [ ] Document lessons learned from any incidents

## Appendix A: Recovery Commands Quick Reference

```bash
# Emergency Stop
fly scale count 0 --app functionfly-orchestrator

# Emergency Restart
fly apps restart functionfly-orchestrator

# View Logs
fly logs --app functionfly-orchestrator --recent

# SSH Debug
fly ssh console --app functionfly-orchestrator

# Database Connect
fly ssh console --app functionfly-orchestrator --command "psql \$DATABASE_URL"

# Check Secrets (names only)
fly secrets list --app functionfly-orchestrator

# Force Redeploy
fly deploy --app functionfly-orchestrator --yes

# Scale Up
fly scale count 2 --app functionfly-orchestrator

# List Releases
fly releases list --app functionfly-orchestrator
```

## Appendix B: Backup Restoration Script

```bash
#!/bin/bash
# restore-from-backup.sh

set -euo pipefail

BACKUP_FILE="${1:-}"
TARGET_DB_URL="${2:-${DATABASE_URL}}"

if [[ -z "$BACKUP_FILE" ]]; then
  echo "Usage: $0 <backup-file> [target-db-url]"
  echo ""
  echo "Example:"
  echo "  $0 registry_20240115_020000_functions.sql.gz"
  exit 1
fi

echo "Restoring from: $BACKUP_FILE"
echo "Target database: $(echo "$TARGET_DB_URL" | sed 's/:[^:@]*@/@/g')"

# Confirm
read -p "Are you sure? This will OVERWRITE existing data! [y/N] " confirm
if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
  echo "Cancelled"
  exit 1
fi

# Extract if compressed
if [[ "$BACKUP_FILE" == *.gz ]]; then
  echo "Extracting..."
  gunzip -c "$BACKUP_FILE" > /tmp/restore.sql
  BACKUP_FILE="/tmp/restore.sql"
fi

# Restore
echo "Restoring database..."
psql "$TARGET_DB_URL" < "$BACKUP_FILE"

echo "✅ Restore complete!"
```

---

**Last Updated**: $(date +%Y-%m-%d)  
**Owner**: Platform Team  
**Review Cycle**: Quarterly
