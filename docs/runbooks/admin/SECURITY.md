# Admin Dashboard Security Runbook

**Document Version:** 1.0  
**Last Updated:** 2026-03-22  
**Owner:** Security Team  

---

## Overview

This runbook documents procedures for responding to security incidents affecting the FunctionFly Admin Dashboard. All security personnel must be familiar with these procedures.

## Table of Contents

1. [Brute Force Attack Response](#1-brute-force-attack-response)
2. [Compromised Admin Account Response](#2-compromised-admin-account-response)
3. [IP Allowlist Management](#3-ip-allowlist-management)
4. [Emergency Session Revocation](#4-emergency-session-revocation)
5. [Security Incident Classification](#5-security-incident-classification)

---

## 1. Brute Force Attack Response

### Symptoms

- Multiple failed login attempts from same IP
- Login attempts with common usernames
- Unusual traffic patterns on admin login endpoint
- Alerts from `failed_login_threshold` security rule

### Response Procedure

#### Step 1: Verify the Attack

```bash
# Check recent failed login attempts
SELECT * FROM auth_events 
WHERE event_type IN ('login_failed', 'login_blocked')
AND created_at >= NOW() - INTERVAL '1 hour'
ORDER BY created_at DESC;

# Check for blocked IPs
SELECT DISTINCT ip_address, COUNT(*) as attempts
FROM auth_events
WHERE event_type = 'login_blocked'
AND created_at >= NOW() - INTERVAL '1 hour'
GROUP BY ip_address;
```

#### Step 2: Block Attacking IPs

```bash
# Block IP at application level (already done by system)
# Verify IP is blocked in Redis
redis-cli GET "security:blocked:<attacking_ip>"

# If not blocked, manually block
redis-cli SET "security:blocked:<attacking_ip>" "1" EX 3600
```

#### Step 3: Review and Harden

- [ ] Review login attempts to identify pattern
- [ ] Check if any accounts were compromised
- [ ] Verify rate limiting is working correctly
- [ ] Consider implementing temporary IP range block
- [ ] Update WAF rules if attack continues

#### Step 4: Document Incident

- [ ] Record IP addresses involved
- [ ] Record timestamp of attack
- [ ] Record number of attempts
- [ ] Note any successful logins
- [ ] Submit security incident report

### Escalation Criteria

- Attack persists > 30 minutes
- More than 10 unique IPs involved
- Any successful unauthorized access
- Attack targets multiple user accounts

---

## 2. Compromised Admin Account Response

### Symptoms

- Login from unusual location/device
- Unexpected admin actions
- Session anomaly alerts
- User reports suspicious activity

### Response Procedure

#### Step 1: Immediate Session Revocation

```bash
# Revoke all sessions for user
UPDATE admin_sessions 
SET is_revoked = TRUE, 
    revoked_at = NOW(), 
    revocation_reason = 'security_incident'
WHERE user_id = '<compromised_user_id>'
AND is_revoked = FALSE;

# Force password reset
UPDATE admin_users 
SET password_changed_at = NULL 
WHERE id = '<compromised_user_id>';
```

#### Step 2: Lock Account

```bash
# Disable the account
UPDATE admin_users 
SET deactivated_at = NOW() 
WHERE id = '<compromised_user_id>';
```

#### Step 3: Investigate

- [ ] Review audit logs for user's actions
- [ ] Check what data was accessed
- [ ] Identify any unauthorized changes
- [ ] Review session history for anomalies

#### Step 4: Notify and Restore

- [ ] Notify the account owner
- [ ] Verify identity before restoring
- [ ] Require MFA reset
- [ ] Change password in presence of admin
- [ ] Restore account only after full investigation

### Escalation Criteria

- Unauthorized access to sensitive data
- Unauthorized admin actions performed
- Account used to attack other systems

---

## 3. IP Allowlist Management

### Overview

The IP allowlist restricts admin access to specific IP addresses or CIDR ranges. Proper management is critical for security.

### Adding IP Addresses

#### Via Admin Dashboard

1. Navigate to Admin Dashboard → Security → IP Allowlist
2. Click "Add IP Rule"
3. Enter CIDR notation (e.g., `192.168.1.0/24`)
4. Select applicable role or user
5. Add description for reference
6. Submit

#### Via Database (Emergency)

```sql
-- Add IP for user
INSERT INTO ip_allowlist (user_id, cidr, description, created_by, created_at)
VALUES (
    '<user_uuid>',
    '192.168.1.100/32'::inet,
    'Office IP - emergency add',
    '<admin_uuid>',
    NOW()
);

-- Add IP for role
INSERT INTO ip_allowlist (role, cidr, description, created_by, created_at)
VALUES (
    'support',
    '10.0.0.0/8'::inet,
    'Internal network range',
    '<admin_uuid>',
    NOW()
);
```

### Removing IP Addresses

#### Via Admin Dashboard

1. Navigate to Admin Dashboard → Security → IP Allowlist
2. Find the rule to remove
3. Click "Disable" or "Delete"
4. Confirm action

#### Emergency Removal

```sql
-- Disable without deleting (preserves audit trail)
UPDATE ip_allowlist 
SET is_active = FALSE 
WHERE id = '<rule_id>';

-- Delete completely (use with caution)
DELETE FROM ip_allowlist WHERE id = '<rule_id>';
```

### Verification

```bash
# Check if IP is allowed
SELECT is_ip_allowed(
    '<user_uuid>'::uuid,
    '<role>'::varchar,
    '<ip_to_check>'::inet
);
```

---

## 4. Emergency Session Revocation

### Single User Session Revocation

```sql
-- Revoke all sessions for a user
UPDATE admin_sessions 
SET is_revoked = TRUE,
    revoked_at = NOW(),
    revocation_reason = 'emergency_revocation'
WHERE user_id = '<target_user_id>'
AND is_revoked = FALSE;

-- Record in audit log
INSERT INTO admin_audit_log (
    user_id, action, resource_type, details, ip_address, created_at
) VALUES (
    '<admin_user_id>',
    'session.revoke_all',
    'admin_session',
    '{"target_user": "<target_user_id>", "reason": "emergency"}',
    '<admin_ip>',
    NOW()
);
```

### All Sessions Revocation (Mass Incident)

```sql
-- Revoke ALL admin sessions
UPDATE admin_sessions 
SET is_revoked = TRUE,
    revoked_at = NOW(),
    revocation_reason = 'mass_emergency_revocation'
WHERE is_revoked = FALSE;

-- Notify all users
-- (This should trigger email notifications to all admin users)
```

### Redis Session Cleanup

```bash
# Clear session-related Redis keys
redis-cli KEYS "session:*" | xargs redis-cli DEL
redis-cli KEYS "admin:session:*" | xargs redis-cli DEL
```

---

## 5. Security Incident Classification

### Classification Framework

| Severity | Description | Response Time | Examples |
|----------|-------------|---------------|----------|
| **P1 - Critical** | Active breach, data exfiltration | Immediate | Account compromise, data breach, service down |
| **P2 - High** | Confirmed attack attempt | < 1 hour | Brute force, injection attack, unauthorized access |
| **P3 - Medium** | Suspected attack, anomaly | < 4 hours | Unusual patterns, multiple failed logins |
| **P4 - Low** | Informational, monitoring | < 24 hours | Policy violations, suspicious but benign activity |

### Response Requirements by Severity

#### P1 - Critical

- [ ] Immediate escalation to security team lead
- [ ] Activate incident response team
- [ ] Begin evidence preservation
- [ ] Notify executive team
- [ ] Consider law enforcement involvement
- [ ] Document all actions with timestamps

#### P2 - High

- [ ] Escalate to security team within 15 minutes
- [ ] Begin investigation
- [ ] Prepare stakeholder notification
- [ ] Document initial findings
- [ ] Engage relevant domain experts

#### P3 - Medium

- [ ] Security team investigates within 4 hours
- [ ] Document incident details
- [ ] Monitor for escalation
- [ ] Update if situation changes

#### P4 - Low

- [ ] Log for tracking and review
- [ ] Investigate within 24 hours
- [ ] Document findings
- [ ] Implement preventive measures if needed

---

## Emergency Contacts

| Role | Contact | Availability |
|------|---------|---------------|
| Security Lead | <security-lead@functionfly.com> | 24/7 |
| Platform Admin | <platform-admin@functionfly.com> | Business hours |
| On-call Engineer | PagerDuty escalation | 24/7 |

---

## Document Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-03-22 | Security Team | Initial version |

---

**Next Review Date:** 2026-06-22
