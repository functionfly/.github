# Admin Access Policy

**Document Version:** 1.0  
**Last Updated:** 2026-03-22  
**Owner:** Security Team  
**Review Frequency:** Quarterly  

---

## 1. Purpose

This policy defines the requirements and procedures for requesting, granting, reviewing, and revoking administrative access to the FunctionFly Admin Dashboard. The goal is to ensure proper access control while maintaining security and compliance.

## 2. Scope

This policy applies to all users with administrative privileges including:

- Super Administrators
- Support Administrators
- Billing Administrators
- Developer Administrators
- Any user with elevated permissions

## 3. Admin Roles and Permissions

### 3.1 Role Definitions

| Role | Description | Key Permissions |
|------|-------------|-----------------|
| **super_admin** | Full system access | All permissions, user management, system configuration |
| **support** | Customer support | Read-only tenant/user access, no billing |
| **billing_admin** | Billing management | Billing operations, subscription management |
| **developer_admin** | Developer support | Function registry, deployment access |

### 3.2 Permission Matrix

| Permission | super_admin | support | billing_admin | developer_admin |
|------------|-------------|---------|---------------|------------------|
| View all tenants | ✓ | ✓ | ✓ | ✓ |
| Create/delete tenants | ✓ | ✗ | ✗ | ✗ |
| View all users | ✓ | ✓ | ✓ | ✓ |
| Manage users | ✓ | Limited | ✗ | ✗ |
| View audit logs | ✓ | ✓ | ✓ | ✓ |
| Manage billing | ✓ | ✗ | ✓ | ✗ |
| Manage functions | ✓ | ✗ | ✗ | ✓ |
| System configuration | ✓ | ✗ | ✗ | ✗ |
| IP allowlist | ✓ | ✗ | ✗ | ✗ |

## 4. Access Request Workflow

### 4.1 Standard Access Request

1. **Request Submission**
   - User submits access request via [Access Request Form]
   - Request must include:
     - Requested role/permissions
     - Business justification
     - Duration (if temporary)
     - Manager approval

2. **Review Process**
   - Security team reviews within 2 business days
   - Required approvals:
     - Direct manager
     - Security team (for elevated roles)
     - VP of department (for super_admin)

3. **Provisioning**
   - Access granted within 1 business day of approval
   - User notified via email
   - Access logged in audit system

### 4.2 Emergency Access

For urgent operational needs:

1. Request via Slack #security-requests channel
2. Requires VP-level approval (email confirmation)
3. Temporary access granted (max 24 hours)
4. Full request process completed retroactively

### 4.3 Temporary Access

| Duration | Approval Requirement | Auto-Revoke |
|----------|---------------------|-------------|
| < 1 week | Manager only | Yes |
| 1-4 weeks | Manager + Security | Yes |
| > 4 weeks | Full approval process | Yes |

## 5. Access Review Requirements

### 5.1 Quarterly Access Review

All admin access is reviewed quarterly (every 90 days) by:

- **Security Team**: Reviews all super_admin and support accounts
- **Department Heads**: Review access within their department
- **Compliance Team**: Validates adherence to this policy

### 5.2 Review Process

1. Automated report generated listing:
   - All admin users and their roles
   - Last activity date
   - Access granted date
   - Approval status

2. Managers receive their team's access list

3. Users must confirm:
   - Access is still needed
   - Use was appropriate
   - No unauthorized access occurred

4. Non-responsive accounts are suspended after 2 follow-ups

### 5.3 Access Modification

- Role upgrades require full approval process
- Role downgrades can be requested by user or manager
- All modifications logged in audit trail

## 6. Separation of Duties

### 6.1 Requirements

| Operation | Required Roles |
|-----------|----------------|
| User deactivation | super_admin OR manager |
| Billing changes | billing_admin (not self) |
| IP allowlist changes | super_admin (not self) |
| Audit log review | Not user's own actions |
| System configuration | super_admin only |

### 6.2 Conflict of Interest

Users cannot have:

- Both billing and user management access
- Both support and developer access for same tenant
- Access to audit their own actions

## 7. Access Termination

### 7.1 Immediate Termination Triggers

- Employee termination
- Role change removing admin need
- Security incident indication
- Policy violation
- Unauthorized access suspected

### 7.2 Termination Procedure

1. **Notification**
   - HR/manager notifies Security team
   - Ticket created for tracking

2. **Immediate Actions** (< 1 hour)

   ```sql
   -- Revoke all sessions
   UPDATE admin_sessions 
   SET is_revoked = TRUE, revoked_at = NOW()
   WHERE user_id = '<user_id>';
   
   -- Deactivate account
   UPDATE admin_users 
   SET deactivated_at = NOW() 
   WHERE id = '<user_id>';
   ```

3. **Access Removal**
   - Remove from all roles
   - Remove IP allowlist entries
   - Revoke API keys
   - Remove from identity provider groups

4. **Documentation**
   - Record termination in audit log
   - Document reason for termination
   - Note any security concerns

### 7.3 Post-Termination Review

Within 7 days:

- Review actions taken by terminated user
- Check for unauthorized data access
- Verify proper access removal
- Update compliance records

## 8. Access for Service Accounts

### 8.1 Service Account Requirements

- Must have specific purpose documented
- Shared credentials prohibited
- Key rotation every 90 days
- Monitoring enabled

### 8.2 Service Account Approval

| Type | Approval |
|------|----------|
| Internal tools | Security team |
| External integrations | Security + VP |
| Production access | Security + CTO |

## 9. Compliance and Auditing

### 9.1 Required Logs

All admin actions must be logged:

- Timestamp
- User ID
- Action type
- Resource affected
- IP address
- Result (success/failure)

### 9.2 Log Retention

| Log Type | Retention Period |
|----------|-----------------|
| Authentication | 2 years |
| Admin actions | 2 years |
| Audit reviews | 5 years |
| Incident records | 7 years |

### 9.3 Compliance Reporting

Quarterly reports include:

- Total admin users by role
- Access requests processed
- Terminations completed
- Access reviews conducted
- Incidents by severity

## 10. Training Requirements

### 10.1 Required Training

| Role | Initial | Annual |
|------|---------|--------|
| All admin users | Security awareness + Admin specific | Refresher |
| super_admin | Additional security training | Yes |
| billing_admin | Payment security (PCI) | Yes |

### 10.2 Training Topics

- Security policies and procedures
- Data handling requirements
- Incident reporting
- Access control best practices
- Phishing awareness

## 11. Exceptions

### 11.1 Exception Process

1. Submit exception request with:
   - Specific policy requirement
   - Business justification
   - Alternative controls
   - Duration

2. Exceptions approved by:
   - Security team lead (low risk)
   - CISO (medium risk)
   - Executive team (high risk)

3. Documented and reviewed quarterly

### 11.2 Common Exceptions

- Time-limited elevated access
- Cross-department access for projects
- Temporary staff access

## 12. Document Control

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-03-22 | Security Team | Initial version |

### Review Schedule

- **Next Review**: 2026-06-22
- **Review Owner**: Security Team Lead
- **Approval Required**: CTO, CISO

---

## Appendix A: Access Request Form Fields

```
- Full Name:
- Email:
- Department:
- Manager:
- Requested Role:
- Business Justification:
- Duration:
- Additional Notes:
```

## Appendix B: Emergency Contact Information

| Role | Contact | Response Time |
|------|---------|---------------|
| Security Lead | <security-lead@functionfly.com> | < 15 min |
| Platform Admin | <platform-admin@functionfly.com> | < 1 hour |
| HR Representative | <hr@functionfly.com> | Business hours |
