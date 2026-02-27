# Disaster Recovery Runbook

## Overview

This runbook documents the procedures for handling various disaster scenarios in the FunctionFly multi-region deployment.

## Emergency Contacts

| Role | Contact | Phone | Email |
|------|---------|-------|-------|
| Primary On-Call | [Name] | [Phone] | [Email] |
| Secondary On-Call | [Name] | [Phone] | [Email] |
| Infrastructure Team | Team | [Phone] | infrastructure@functionfly.com |
| Security Team | Team | [Phone] | security@functionfly.com |

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Global Load Balancer                     │
│                      (Cloudflare)                           │
└─────────────────────────┬───────────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
        ▼                 ▼                 ▼
┌───────────────┐  ┌───────────────┐  ┌───────────────┐
│   iad (US)    │  │   lax (US)    │  │   fra (EU)    │
│   PRIMARY     │  │  SECONDARY    │  │  TERTIARY     │
│               │  │               │  │               │
│  ┌─────────┐  │  │  ┌─────────┐  │  │  ┌─────────┐  │
│  │API + DB │  │  │  │API + DB │  │  │  │API + DB │  │
│  └─────────┘  │  │  └─────────┘  │  │  └─────────┘  │
└───────────────┘  └───────────────┘  └───────────────┘
        │                 │                 │
        └─────────────────┼─────────────────┘
                          │
                    Async Replication
```

## Recovery Objectives

| Metric | Target | Description |
|--------|--------|-------------|
| RTO | 5 minutes | Recovery Time Objective |
| RPO | 1 hour | Recovery Point Objective (backup frequency) |
| Availability | 99.9% | Target uptime |

---

## Scenario 1: Complete Region Outage

### Detection
- Health checks failing for all instances in region
- Increased error rates from monitoring
- Customer complaints via support channels
- Alert from automated monitoring

### Steps

1. **Acknowledge the alert**
   ```bash
   # Check current status
   ./scripts/failover-control-plane.sh health iad
   ```

2. **Verify the outage**
   ```bash
   # Check all regions
   for region in iad lax fra; do
       echo "Checking $region..."
       ./scripts/failover-control-plane.sh health $region
   done
   ```

3. **Run automated failover** (if auto-failover is not enabled)
   ```bash
   # Dry run first (optional)
   DRY_RUN=true ./scripts/failover-control-plane.sh check
   
   # Actual failover
   ./scripts/failover-control-plane.sh check
   ```

4. **Verify DNS propagation**
   ```bash
   dig api.functionfly.com
   # or
   nslookup api.functionfly.com
   ```

5. **Check application health**
   ```bash
   curl https://api.functionfly.com/healthz
   ```

6. **Update status page**
   ```bash
   # Post to status page (configured via STATUS_PAGE_URL)
   ```

### Recovery Time Objective: **5 minutes**

### Post-Failover Actions
- [ ] Monitor the new primary region for issues
- [ ] Investigate the failed region
- [ ] Plan for region recovery
- [ ] Update documentation with lessons learned

---

## Scenario 2: Database Corruption

### Detection
- Database health checks failing
- Query errors in logs (`ERROR: could not read block`)
- Replication lag increasing dramatically
- Data integrity check failures

### Steps

1. **Stop application traffic**
   ```bash
   # Scale down the orchestrator
   flyctl scale count 0 --region iad --app functionfly-control
   ```

2. **Identify last good backup**
   ```bash
   # List available backups
   aws s3 ls s3://functionfly-backups/iad/backups/
   
   # Verify latest backup
   ./scripts/multi-region-backup.sh verify iad latest.sql.gz
   ```

3. **Restore from backup**
   ```bash
   # Restore to point-in-time (if using WAL archiving)
   ./scripts/restore-database-pitr.sh --timestamp "2024-01-15 14:30:00" --region iad
   
   # Or restore from latest backup
   ./scripts/restore-database.sh --backup-file latest.sql.gz --region iad
   ```

4. **Verify data integrity**
   ```bash
   # Run integrity checks
   psql -c "SELECT COUNT(*) FROM users;" functionfly
   psql -c "SELECT COUNT(*) FROM functions;" functionfly
   ```

5. **Resume traffic**
   ```bash
   # Scale back up
   flyctl scale count 2 --region iad --app functionfly-control
   ```

### Recovery Time Objective: **30 minutes**

### Prevention
- Enable WAL archiving for point-in-time recovery
- Set up automated integrity checks
- Monitor replication lag

---

## Scenario 3: Data Loss (Accidental Delete)

### Detection
- Missing data in application
- Backup verification failures
- Customer reports of missing content

### Steps

1. **Immediately stop all write operations**
   ```bash
   # Enable read-only mode
   # This requires implementation in your application
   ```

2. **Identify scope of data loss**
   ```bash
   # Check logs for delete operations
   grep "DELETE" /var/log/functionfly/api.log | tail -100
   
   # Identify affected tables
   ```

3. **Restore from backup**
   ```bash
   # Find backup before the deletion
   aws s3 ls s3://functionfly-backups/iad/backups/ | grep "2024"
   
   # Restore to a new instance first
   ./scripts/restore-database.sh \
       --backup-file functionfly_iad_20240115_120000.sql.gz \
       --new-instance functionfly-restore
   ```

4. **Compare restored data with logs**
   ```bash
   # Extract data that needs to be recovered
   # Apply any transactions that occurred after the backup
   ```

5. **Migrate recovered data to production**
   ```bash
   # Carefully migrate only the lost data
   ```

### Recovery Time Objective: **1 hour**

### Prevention
- Implement soft delete (mark as deleted rather than actual delete)
- Set up audit logging for all delete operations
- Regular backup verification

---

## Scenario 4: Security Breach

### Detection
- Unauthorized access alerts
- Suspicious API activity patterns
- Unexpected data exfiltration
- Malware detection
- Unusual resource usage

### Steps

1. **Isolate affected systems**
   ```bash
   # Block network access to compromised instance
   flyctl ips allocate-v4 --region iad --app functionfly-control --yes
   flyctl ips release <compromised-ip>
   
   # Or scale to zero temporarily
   flyctl scale count 0 --region iad --app functionfly-control
   ```

2. **Rotate all credentials**
   ```bash
   # Rotate database passwords
   # Update secrets in Infisical/vault
   
   # Rotate API keys
   # Generate new JWT secrets
   ```

3. **Enable enhanced audit logging**
   ```bash
   # Enable verbose logging
   export LOG_LEVEL=debug
   ```

4. **Restore from known-good backup**
   ```bash
   # Only if breach involved data tampering
   # Choose a backup from before the breach
   ./scripts/restore-database.sh --backup-file <pre-breach-backup>
   ```

5. **Review access logs**
   ```bash
   # Analyze what was accessed
   grep "unauthorized" /var/log/functionfly/auth.log
   grep "api.functionfly.com" access.log | grep "200" | wc -l
   ```

6. **Report to compliance team**
   ```bash
   # Document everything
   # Follow GDPR/CCPA notification requirements
   ```

### Recovery Time Objective: **2 hours**

### Prevention
- Regular security audits
- Implement rate limiting
- Enable WAF rules
- Network segmentation

---

## Scenario 5: Network Partition

### Detection
- Intermittent connectivity between regions
- High latency between data centers
- DNS resolution issues

### Steps

1. **Check network status**
   ```bash
   # Test connectivity
   ping iad.functionfly.internal
   ping lax.functionfly.internal
   
   # Check DNS
   dig +short functionfly-control.iad.internal
   ```

2. **Enable fallback mode**
   ```bash
   # Switch to regional operation mode
   export OPERATION_MODE=regional
   ```

3. **Test cross-region connectivity**
   ```bash
   # Verify replication status
   psql -c "SELECT * FROM pg_stat_replication;" functionfly
   ```

4. **Gradually restore connectivity**
   ```bash
   # Monitor as you bring services back online
   ```

---

## Verification Checklist

After any failover or recovery, verify:

- [ ] Health endpoints responding (`curl /healthz`)
- [ ] Database connections working
- [ ] DNS resolving correctly (`dig api.functionfly.com`)
- [ ] Edge targets accessible
- [ ] Customer authentication working
- [ ] Monitoring showing green status
- [ ] Recent backups verified
- [ ] Team notified of recovery
- [ ] Incident report filed

---

## Rollback Procedures

If a failover causes issues and you need to rollback:

```bash
# Rollback from lax back to iad
./scripts/failover-control-plane.sh rollback lax iad

# Verify rollback
./scripts/failover-control-plane.sh health iad
```

---

## Escalation Procedures

| Time | Action |
|------|--------|
| 0 min | On-call engineer acknowledges |
| 5 min | If not resolved, notify infrastructure lead |
| 15 min | Notify CTO, begin customer communication |
| 30 min | Consider engaging external support |
| 60 min | Executive notification required |

---

## Useful Commands

```bash
# Check region health
./scripts/failover-control-plane.sh health iad

# Run failover check
./scripts/failover-control-plane.sh check

# Start auto-failover monitor
./scripts/failover-control-plane.sh auto

# Verify backup
./scripts/multi-region-backup.sh verify iad latest.sql.gz

# List backups
aws s3 ls s3://functionfly-backups/iad/backups/

# Check Fly.io status
flyctl status --app functionfly-control

# View logs
flyctl logs --region iad --app functionfly-control
```

---

## Post-Incident Review

After any major incident, conduct a post-incident review:

1. **Timeline**: Document what happened and when
2. **Root Cause**: Identify the underlying cause
3. **Impact**: Document customer and business impact
4. **Response**: Evaluate the response effectiveness
5. **Prevention**: Identify improvements to prevent recurrence

Template: [Post-Incident Review Template](link-to-template)

---

## Testing

### Regular Testing Schedule

| Test | Frequency | Owner |
|------|-----------|-------|
| Failover drill | Monthly | Infrastructure |
| Backup restoration | Weekly | DBA |
| Security incident response | Quarterly | Security |
| Full DR test | Annually | CTO |

### Test Commands

```bash
# Dry run failover
DRY_RUN=true ./scripts/failover-control-plane.sh check

# Restore test
./scripts/restore-database.sh --backup-file <test-backup> --new-instance functionfly-test
```
