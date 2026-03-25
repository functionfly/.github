# FunctionFly Disaster Recovery Runbook

## Overview

This runbook provides step-by-step procedures for recovering from various failure scenarios in the FunctionFly platform.

## Table of Contents

1. [Database Failure](#database-failure)
2. [Redis Failure](#redis-failure)
3. [Orchestrator API Failure](#orchestrator-api-failure)
4. [Health Monitor Failure](#health-monitor-failure)
5. [Network Partition](#network-partition)
6. [Data Corruption](#data-corruption)
7. [Security Incident](#security-incident)

---

## Database Failure

### Symptoms

- API returns 500 errors
- Health checks fail
- Logs show database connection errors

### Immediate Actions

1. **Check database status**

   ```bash
   # Check if database is running
   sudo systemctl status postgresql

   # Check database connections
   sudo -u postgres psql -c "SELECT count(*) FROM pg_stat_activity;"
   ```

2. **Check database logs**

   ```bash
   sudo tail -f /var/log/postgresql/postgresql-*.log
   ```

3. **Restart database if needed**

   ```bash
   sudo systemctl restart postgresql
   ```

### Recovery Steps

1. **If database is corrupted**

   ```bash
   # Stop the database
   sudo systemctl stop postgresql

   # Restore from backup
   sudo -u postgres pg_restore -d functionfly /backup/functionfly_$(date +%Y%m%d).dump

   # Start the database
   sudo systemctl start postgresql
   ```

2. **If database is slow**

   ```bash
   # Check for long-running queries
   sudo -u postgres psql -c "SELECT pid, now() - pg_stat_activity.query_start AS duration, query FROM pg_stat_activity WHERE (now() - pg_stat_activity.query_start) > interval '5 minutes';"

   # Kill long-running queries
   sudo -u postgres psql -c "SELECT pg_terminate_backend(pid);"
   ```

3. **Verify recovery**

   ```bash
   # Test database connection
   curl http://localhost:8080/health

   # Check API functionality
   curl http://localhost:8080/v1/apps
   ```

---

## Redis Failure

### Symptoms

- Cache misses increase
- API response times increase
- Logs show Redis connection errors

### Immediate Actions

1. **Check Redis status**

   ```bash
   redis-cli ping
   redis-cli info
   ```

2. **Check Redis logs**

   ```bash
   sudo tail -f /var/log/redis/redis-server.log
   ```

3. **Restart Redis if needed**

   ```bash
   sudo systemctl restart redis
   ```

### Recovery Steps

1. **If Redis is corrupted**

   ```bash
   # Stop Redis
   sudo systemctl stop redis

   # Clear Redis data
   sudo rm -f /var/lib/redis/dump.rdb

   # Start Redis
   sudo systemctl start redis
   ```

2. **If Redis is slow**

   ```bash
   # Check memory usage
   redis-cli info memory

   # Clear old keys
   redis-cli --scan --pattern "registry:*" | xargs redis-cli del
   ```

3. **Verify recovery**

   ```bash
   # Test Redis connection
   redis-cli ping

   # Test cache functionality
   curl http://localhost:8080/v1/apps
   ```

---

## Orchestrator API Failure

### Symptoms

- API returns 502/503 errors
- Health checks fail
- Logs show application errors

### Immediate Actions

1. **Check API status**

   ```bash
   # Check if API is running
   ps aux | grep orchestrator-api

   # Check API logs
   journalctl -u orchestrator-api -f
   ```

2. **Restart API if needed**

   ```bash
   # Restart API
   sudo systemctl restart orchestrator-api

   # Or restart Docker container
   docker restart orchestrator-api
   ```

### Recovery Steps

1. **If API is crashing**

   ```bash
   # Check for errors in logs
   journalctl -u orchestrator-api --since "1 hour ago" | grep -i error

   # Check system resources
   free -h
   df -h
   top
   ```

2. **If API is slow**

   ```bash
   # Check database connections
   sudo -u postgres psql -c "SELECT count(*) FROM pg_stat_activity;"

   # Check Redis connections
   redis-cli info clients

   # Check system load
   uptime
   ```

3. **Verify recovery**

   ```bash
   # Test API health
   curl http://localhost:8080/health

   # Test API functionality
   curl http://localhost:8080/v1/apps
   ```

---

## Health Monitor Failure

### Symptoms

- Backend health status not updating
- Circuit breakers not working
- Logs show health check errors

### Immediate Actions

1. **Check health monitor status**

   ```bash
   # Check if health monitor is running
   ps aux | grep health-monitor

   # Check health monitor logs
   journalctl -u health-monitor -f
   ```

2. **Restart health monitor if needed**

   ```bash
   sudo systemctl restart health-monitor
   ```

### Recovery Steps

1. **If health monitor is crashing**

   ```bash
   # Check for errors in logs
   journalctl -u health-monitor --since "1 hour ago" | grep -i error

   # Check database connectivity
   sudo -u postgres psql -c "SELECT 1;"
   ```

2. **If health checks are failing**

   ```bash
   # Check backend URLs
   curl -I https://example.edge-target.dev/healthz

   # Check network connectivity
   ping example.edge-target.dev
   ```

3. **Verify recovery**

   ```bash
   # Check health monitor logs
   journalctl -u health-monitor -n 50

   # Check backend health status
   curl http://localhost:8080/v1/apps/{appId}/backends
   ```

---

## Network Partition

### Symptoms

- Some backends unreachable
- Health checks failing for specific regions
- Logs show connection timeouts

### Immediate Actions

1. **Check network connectivity**

   ```bash
   # Test connectivity to backends
   ping example.edge-target.dev
   curl -I https://example.edge-target.dev/healthz

   # Check DNS resolution
   nslookup example.edge-target.dev
   ```

2. **Check firewall rules**

   ```bash
   # Check iptables rules
   sudo iptables -L -n

   # Check if ports are open
   netstat -tlnp | grep 8080
   ```

### Recovery Steps

1. **If network is partitioned**

   ```bash
   # Check routing
   traceroute example.edge-target.dev

   # Check if VPN is up
   ip addr show
   ```

2. **If DNS is failing**

   ```bash
   # Check DNS configuration
   cat /etc/resolv.conf

   # Test DNS resolution
   dig example.edge-target.dev
   ```

3. **Verify recovery**

   ```bash
   # Test connectivity
   curl -I https://example.edge-target.dev/healthz

   # Check health monitor
   journalctl -u health-monitor -n 20
   ```

---

## Data Corruption

### Symptoms

- API returns unexpected data
- Database queries fail
- Logs show data integrity errors

### Immediate Actions

1. **Stop writes to database**

   ```bash
   # Put API in read-only mode
   export READ_ONLY_MODE=true
   sudo systemctl restart orchestrator-api
   ```

2. **Assess corruption**

   ```bash
   # Check database integrity
   sudo -u postgres psql -c "SELECT * FROM pg_catalog.pg_tables;"

   # Check for errors
   sudo -u postgres psql -c "SELECT * FROM pg_stat_database WHERE datname = 'functionfly';"
   ```

### Recovery Steps

1. **If data is corrupted**

   ```bash
   # Restore from backup
   sudo -u postgres pg_restore -d functionfly /backup/functionfly_$(date +%Y%m%d).dump

   # Verify data integrity
   sudo -u postgres psql -c "SELECT count(*) FROM apps;"
   ```

2. **If specific tables are corrupted**

   ```bash
   # Rebuild specific tables
   sudo -u postgres psql -c "REINDEX TABLE apps;"
   sudo -u postgres psql -c "REINDEX TABLE backends;"
   ```

3. **Verify recovery**

   ```bash
   # Test API functionality
   curl http://localhost:8080/v1/apps

   # Check data integrity
   sudo -u postgres psql -c "SELECT count(*) FROM apps;"
   ```

---

## Security Incident

### Symptoms

- Unauthorized access detected
- Suspicious API calls
- Logs show security violations

### Immediate Actions

1. **Isolate affected systems**

   ```bash
   # Block suspicious IPs
   sudo iptables -A INPUT -s suspicious_ip -j DROP

   # Revoke compromised API keys
   curl -X DELETE http://localhost:8080/v1/api-keys/{keyId}
   ```

2. **Preserve evidence**

   ```bash
   # Copy logs
   cp -r /var/log/functionfly /backup/incident_$(date +%Y%m%d_%H%M%S)

   # Export database
   sudo -u postgres pg_dump functionfly > /backup/incident_db_$(date +%Y%m%d).dump
   ```

### Recovery Steps

1. **If API keys are compromised**

   ```bash
   # Rotate all API keys
   curl -X POST http://localhost:8080/v1/api-keys/rotate

   # Update secrets
   export JWT_SECRET=$(openssl rand -base64 32)
   sudo systemctl restart orchestrator-api
   ```

2. **If database is compromised**

   ```bash
   # Restore from clean backup
   sudo -u postgres pg_restore -d functionfly /backup/functionfly_clean.dump

   # Change database passwords
   sudo -u postgres psql -c "ALTER USER postgres PASSWORD 'new_password';"
   ```

3. **Verify recovery**

   ```bash
   # Test API functionality
   curl http://localhost:8080/health

   # Check for suspicious activity
   journalctl -u orchestrator-api --since "1 hour ago" | grep -i "unauthorized"
   ```

---

## Backup Procedures

### Daily Backup

```bash
#!/bin/bash
# daily-backup.sh

DATE=$(date +%Y%m%d)
BACKUP_DIR="/backup"

# Backup database
sudo -u postgres pg_dump functionfly > $BACKUP_DIR/functionfly_$DATE.dump

# Backup Redis
redis-cli BGSAVE
cp /var/lib/redis/dump.rdb $BACKUP_DIR/redis_$DATE.rdb

# Backup configuration
cp -r /etc/functionfly $BACKUP_DIR/config_$DATE

# Clean old backups (keep 30 days)
find $BACKUP_DIR -name "*.dump" -mtime +30 -delete
find $BACKUP_DIR -name "*.rdb" -mtime +30 -delete
find $BACKUP_DIR -name "config_*" -mtime +30 -delete
```

### Weekly Backup Verification

```bash
#!/bin/bash
# weekly-backup-verify.sh

# Test database restore
sudo -u postgres pg_restore -d functionfly_test /backup/functionfly_$(date +%Y%m%d).dump

# Test Redis restore
redis-cli FLUSHALL
cp /backup/redis_$(date +%Y%m%d).rdb /var/lib/redis/dump.rdb
sudo systemctl restart redis

# Verify data
curl http://localhost:8080/health
```

---

## Monitoring and Alerting

### Key Metrics to Monitor

1. **Database**
   - Connection count
   - Query duration
   - Replication lag

2. **Redis**
   - Memory usage
   - Connection count
   - Hit rate

3. **API**
   - Response time
   - Error rate
   - Request rate

4. **System**
   - CPU usage
   - Memory usage
   - Disk usage

### Alert Thresholds

| Metric | Warning | Critical |
|--------|---------|----------|
| Database connections | > 50 | > 100 |
| Redis memory | > 80% | > 90% |
| API response time | > 500ms | > 1000ms |
| Error rate | > 1% | > 5% |
| CPU usage | > 70% | > 90% |
| Memory usage | > 80% | > 90% |
| Disk usage | > 80% | > 90% |

---

## Contact Information

### Escalation Path

1. **Level 1**: On-call engineer
2. **Level 2**: Senior engineer
3. **Level 3**: Engineering manager
4. **Level 4**: CTO

### Communication Channels

- **Slack**: #functionfly-alerts
- **Email**: <alerts@functionfly.com>
- **Phone**: +1-555-0123

---

## Post-Incident Review

After any incident, conduct a post-incident review:

1. **Timeline**: Document what happened and when
2. **Root Cause**: Identify the root cause
3. **Impact**: Assess the impact on users
4. **Action Items**: Create action items to prevent recurrence
5. **Lessons Learned**: Document lessons learned

---

*Last Updated: 2026-03-19*
*Version: 1.0*
