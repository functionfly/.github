# Incident Response Procedure

**Document Version:** 1.0  
**Last Updated:** 2026-03-22  
**Owner:** Security Team  

---

## Overview

This document outlines the incident response procedures for security and operational incidents affecting the FunctionFly platform, including the admin dashboard.

## Table of Contents

1. [Incident Classification](#1-incident-classification)
2. [Response Team Contacts](#2-response-team-contacts)
3. [Response Procedures by Severity](#3-response-procedures-by-severity)
4. [Communication Procedures](#4-communication-procedures)
5. [Evidence Preservation](#5-evidence-preservation)
6. [Post-Incident Review](#6-post-incident-review)

---

## 1. Incident Classification

### 1.1 Severity Levels

| Severity | Description | Response Time | Examples |
|----------|-------------|---------------|----------|
| **P1 - Critical** | Active breach, complete service loss | Immediate (< 15 min) | Data exfiltration, complete system compromise |
| **P2 - High** | Significant security event, major outage | < 1 hour | Brute force attack, partial data breach, DDoS |
| **P3 - Medium** | Limited security event, degraded service | < 4 hours | Unauthorized access attempt, minor vulnerability |
| **P4 - Low** | Informational, minimal impact | < 24 hours | Policy violation, suspicious activity |

### 1.2 Incident Categories

| Category | Description |
|----------|-------------|
| **SEC** | Security incident (breach, attack, compromise) |
| **OPS** | Operational incident (outage, degradation) |
| **DATA** | Data-related incident (loss, corruption, exposure) |
| **INFRA** | Infrastructure incident (network, compute, storage) |

### 1.3 Classification Examples

```
P1-SEC: Admin account credentials stolen and used
P1-OPS: Complete platform outage
P1-DATA: Customer data exfiltration confirmed
P1-INFRA: Data center failure with data loss

P2-SEC: Brute force attack in progress
P2-OPS: Multiple services degraded
P2-DATA: Accidental data exposure (no theft)
P2-INFRA: Single-region outage

P3-SEC: Failed unauthorized access attempts
P3-OPS: Non-critical service slow
P3-DATA: Data integrity issue (no loss)
P3-INFRA: Non-production system affected

P4-SEC: Suspicious but benign activity
P4-OPS: Minor UI bug
P4-DATA: Test data in production
P4-INFRA: Temporary monitoring gap
```

---

## 2. Response Team Contacts

### 2.1 Primary Response Team

| Role | Name | Contact | Backup Contact |
|------|------|---------|---------------|
| Incident Commander | On-call rotation | PagerDuty | Security Lead |
| Security Lead | [NAME] | <security-lead@functionfly.com> | [BACKUP] |
| Platform Lead | [NAME] | <platform-lead@functionfly.com> | [BACKUP] |
| Engineering Lead | [NAME] | <eng-lead@functionfly.com> | [BACKUP] |

### 2.2 Escalation Contacts

| Role | Contact | When to Escalate |
|------|---------|------------------|
| CEO | [EMAIL] | P1 incidents only |
| CTO | [EMAIL] | P1/P2 incidents |
| VP Engineering | [EMAIL] | All security incidents |
| Legal Counsel | [EMAIL] | Data breaches, regulatory |
| PR/Communications | [EMAIL] | Customer-facing incidents |

### 2.3 External Contacts

| Vendor/Service | Contact | Purpose |
|---------------|---------|---------|
| AWS Support | [ACCOUNT #] | Infrastructure issues |
| Datadog | [CONTACT] | Monitoring/alerting |
| Cloudflare | [CONTACT] | CDN/DDoS issues |
| External Security Consultant | [NAME] | P1 incident support |

---

## 3. Response Procedures by Severity

### 3.1 P1 - Critical Incident Response

#### Immediate Actions (0-15 minutes)

1. **Acknowledge and Declare**

   ```
   - Acknowledge alert in PagerDuty
   - Declare P1 incident in #incident-response channel
   - Start incident timeline document
   ```

2. **Assemble Response Team**

   ```
   - Page incident commander
   - Notify security lead
   - Engage relevant domain experts
   - Begin war room (video call)
   ```

3. **Initial Assessment**

   ```
   - What systems are affected?
   - What data is at risk?
   - Is the incident ongoing?
   - What is the blast radius?
   ```

4. **Containment (Immediate)**

   ```sql
   -- Emergency: Revoke all sessions if credential compromise
   UPDATE admin_sessions SET is_revoked = TRUE, 
   revoked_at = NOW(), revocation_reason = 'P1_incident'
   WHERE is_revoked = FALSE;
   
   -- Block attacking IPs at firewall
   iptables -A INPUT -s <attacker_ip> -j DROP
   ```

#### Short-term Actions (15 minutes - 1 hour)

- [ ] Implement full containment
- [ ] Begin evidence preservation
- [ ] Notify executive team
- [ ] Prepare customer communication
- [ ] Consider law enforcement notification
- [ ] Document all actions with timestamps

#### Recovery (1+ hours)

- [ ] Verify containment is complete
- [ ] Begin system recovery
- [ ] Validate data integrity
- [ ] Restore services systematically
- [ ] Monitor for reoccurrence

### 3.2 P2 - High Severity Incident Response

#### Immediate Actions (0-1 hour)

1. **Acknowledge**

   ```
   - Acknowledge alert
   - Create incident ticket
   - Start timeline
   ```

2. **Assess**

   ```
   - Determine scope
   - Identify affected systems
   - Assess current threat status
   ```

3. **Contain**

   ```
   - Block malicious IPs
   - Revoke compromised credentials
   - Enable additional monitoring
   ```

#### Follow-up (1-4 hours)

- [ ] Complete investigation
- [ ] Notify stakeholders
- [ ] Document findings
- [ ] Implement fixes
- [ ] Plan recovery

### 3.3 P3 - Medium Severity Incident Response

#### Actions (0-4 hours)

1. **Review** - Examine alert details
2. **Investigate** - Determine cause and scope
3. **Remediate** - Fix vulnerability/issue
4. **Document** - Record findings

#### Escalation Triggers

- Escalate to P2 if:
  - Incident spreads to more systems
  - Confirmed unauthorized access
  - Service impact increases

### 3.4 P4 - Low Severity Incident Response

#### Actions (0-24 hours)

1. **Log** - Document the event
2. **Review** - Determine if investigation needed
3. **Address** - Remediate if simple
4. **Close** - Document resolution

---

## 4. Communication Procedures

### 4.1 Internal Communication

| Audience | Channel | Update Frequency |
|----------|---------|------------------|
| Response team | #incident-response (Slack) + war room | Continuous |
| Engineering | #engineering | Hourly for P1/P2 |
| All staff | #company-wide | As needed |
| Executive team | Email + direct | P1: 30 min, P2: 2 hours |

### 4.2 Customer Communication

#### Communication Triggers

- P1 incidents affecting customers: Always
- P2 incidents with > 1 hour impact: Always
- P3 incidents with data implications: Consider

#### Communication Template

```
Subject: [INCIDENT] Service Status Update - [TIME]

Incident ID: [ID]
Status: [Investigating/Identified/Resolving/Resolved]
Impact: [Description of impact to customers]
Start Time: [UTC timestamp]
Current Update: [What's happening now]

What we're doing:
- [Action 1]
- [Action 2]

Next update: [TIME] UTC
```

### 4.3 Regulatory Communication

| Regulation | Requirement | Timeline |
|------------|-------------|----------|
| GDPR | Data breach notification | 72 hours |
| SOC 2 | Security incidents | 30 days |
| PCI DSS | Cardholder data breach | 24 hours |
| HIPAA | Protected health info | 60 days |

---

## 5. Evidence Preservation

### 5.1 Collection Priorities

1. **Memory/RAM** - Volatile data, malware analysis
2. **Disk Images** - Full system state
3. **Network Logs** - Traffic patterns
4. **Application Logs** - Access patterns
5. **Authentication Logs** - Login attempts
6. **Database State** - Data at time of incident

### 5.2 Collection Procedure

```bash
# Create evidence directory
mkdir -p /incident/YYYYMMDD_[incident_id]
cd /incident/YYYYMMDD_[incident_id]

# Collect system memory (if applicable)
sudo dd if=/dev/mem of=memory.img bs=1M

# Collect disk image (if applicable)
sudo dd if=/dev/[disk] of=disk.img bs=1M

# Copy logs
cp -r /var/log ./logs
cp -r /home/*/.bash_history ./

# Export database state
pg_dump -Fc -f database_dump.sql

# Collect network captures (if available)
tcpdump -w network_capture.pcap -i [interface]

# Create chain of custody document
cat > chain_of_custody.txt << EOF
Incident ID: [ID]
Collector: [NAME]
Date/Time: [UTC]
Systems: [LIST]
Hashes: [SHA256 hashes of all collected files]
EOF
```

### 5.3 Chain of Custody

All evidence must be documented:

```
Evidence ID: [UUID]
Description: [What it is]
Collected from: [System/location]
Collected by: [Name]
Date/Time: [UTC]
Hash: [SHA256]
Storage location: [Where kept]
Access log: [Who accessed and when]
```

---

## 6. Post-Incident Review

### 6.1 Requirements by Severity

| Severity | Post-Incident Review Required | Timeline |
|----------|-------------------------------|----------|
| P1 | Yes - Mandatory | 48 hours |
| P2 | Yes - Mandatory | 1 week |
| P3 | Recommended | 2 weeks |
| P4 | Optional | As needed |

### 6.2 Post-Incident Review Template

```markdown
# Incident Post-Incident Review

## Incident Summary
- **ID**: 
- **Severity**: 
- **Date/Time**: 
- **Duration**: 
- **Status**: 

## Impact
- Systems affected:
- Users affected:
- Data at risk:
- Financial impact:

## Timeline
| Time (UTC) | Action |
|------------|--------|
| HH:MM | Event description |
| HH:MM | Action taken |
| HH:MM | Update |

## Root Cause Analysis
- What happened:
- Why it happened:
- Why detection was delayed (if applicable):

## Response Evaluation
### What went well:
-

### What could be improved:
-

### Action items:
| Item | Owner | Due Date |
|------|-------|----------|
| | | |

## Lessons Learned
1.
2.
3.

## Sign-off
- Incident Commander: [Name/Date]
- Security Lead: [Name/Date]
- Engineering Lead: [Name/Date]
```

### 6.3 Action Item Tracking

All action items from post-incident reviews must be:

- [ ] Entered into issue tracking system
- [ ] Assigned to specific owner
- [ ] Given target completion date
- [ ] Reviewed in weekly security meeting
- [ ] Closed only when verified complete

---

## Appendix A: Quick Reference

### Emergency Contacts

```
Security Lead:    [PHONE]
Platform Lead:    [PHONE]
CTO:             [PHONE]
PagerDuty:       [URL]
```

### Common Commands

```sql
-- Emergency session revocation
UPDATE admin_sessions SET is_revoked = TRUE, revoked_at = NOW() 
WHERE user_id = '<id>' AND is_revoked = FALSE;

-- Block IP
iptables -A INPUT -s <ip> -j DROP

-- Collect evidence hash
sha256sum [file] > hash.txt
```

---

## Document Control

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-03-22 | Security Team | Initial version |

**Next Review:** 2026-06-22
