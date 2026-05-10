---
title: Organizations & Teams
description: Manage teams, roles, and permissions in FunctionFly.
sidebar:
  order: 13
---



This guide covers how to create and manage organizations, invite team members, and control access with roles.

## Organizations Overview

An **Organization** is the top-level container for your FunctionFly resources. All your functions, agents, configurations, and billing live within an organization.

### Default Organization

When you sign up, you automatically get a **personal organization** named after your username. This is your default workspace.

### Creating an Organization

1. Click your avatar in the dashboard
2. Select **Create Organization**
3. Enter organization name and slug
4. Choose plan (can be changed later)
5. Click **Create**

---

## Roles & Permissions

### Available Roles

| Role | Description | Use Case |
|------|-------------|----------|
| **Owner** | Full access including billing and deletion | Founders, team leads |
| **Admin** | Manage members, settings, and all resources | Team managers |
| **Member** | Create and manage functions, agents, state | Developers |
| **Viewer** | Read-only access to all resources | Stakeholders, auditors |

### Permission Matrix

| Action | Owner | Admin | Member | Viewer |
|---------|-------|-------|--------|--------|
| View resources | ✓ | ✓ | ✓ | ✓ |
| Create functions | ✓ | ✓ | ✓ | — |
| Deploy functions | ✓ | ✓ | ✓ | — |
| Manage secrets | ✓ | ✓ | ✓ | — |
| Create agents | ✓ | ✓ | ✓ | — |
| Manage team | ✓ | ✓ | — | — |
| Change billing | ✓ | — | — | — |
| Delete organization | ✓ | — | — | — |
| Manage SSO/SAML | ✓ | — | — | — |

---

## Inviting Team Members

### Sending Invites

1. Go to **Settings → Team** (or **Organization Settings → Team**)
2. Click **Invite Member**
3. Enter email address
4. Select role
5. Click **Send Invite**

The invite expires after **7 days**.

### Accepting Invites

Invitees receive an email with:
- Organization name
- Assigned role
- **Accept** button (creates account if needed)
- **Decline** option

### Bulk Invites

For larger teams, upload a CSV:

```csv
email,role
alice@example.com,member
bob@example.com,viewer
charlie@example.com,admin
```

1. Go to **Settings → Team**
2. Click **Bulk Invite**
3. Upload CSV
4. Review and confirm

---

## Managing Members

### Changing Roles

1. Go to **Settings → Team**
2. Find the member
3. Click the role dropdown
4. Select new role
5. Changes apply immediately

### Removing Members

1. Go to **Settings → Team**
2. Find the member
3. Click **Remove**
4. Confirm removal

**Note:** Removed members lose access immediately. Their functions remain but are reassigned to the organization.

### Resending Invites

Invites that haven't been accepted can be resent:

1. Go to **Settings → Team**
2. Find pending invite
3. Click **Resend**

---

## Organization Settings

### General Settings

| Setting | Description |
|---------|-------------|
| **Name** | Display name for the organization |
| **Slug** | URL-friendly identifier (cannot be changed) |
| **Logo** | Organization logo (shown in dashboard) |
| **Website** | Optional website URL |

### Security Settings

#### Multi-Factor Authentication (MFA)

**Enforce MFA for all members (Professional+):**

1. Go to **Settings → Security**
2. Enable **Require MFA**
3. Members have 7 days to enroll
4. After grace period, MFA is required to access

#### Session Management

| Setting | Default | Description |
|---------|---------|-------------|
| Session timeout | 24 hours | Auto-logout after inactivity |
| Max sessions | 5 | Concurrent sessions per user |
| Session logging | Enabled | Track all session activity |

### API Access

Generate API keys scoped to your organization:

1. Go to **Settings → API Keys**
2. Click **Create Organization API Key**
3. Name the key
4. Set permissions (read-only, read-write, admin)
5. Copy and store securely

---

## SSO & Enterprise Identity (Enterprise)

### SAML Single Sign-On

Enterprise plans support SAML 2.0 for SSO:

**Setup steps:**

1. Go to **Settings → Security → SSO**
2. Click **Configure SAML**
3. Download Service Provider metadata
4. Configure your Identity Provider (IdP):
   - Entity ID: `https://functionfly.com/saml/{org-slug}`
   - ACS URL: `https://functionfly.com/saml/{org-slug}/acs`
   - Attribute mapping:
     - `email` → Email address
     - `firstName` → First name
     - `lastName` → Last name
5. Upload IdP metadata XML
6. Test connection
7. Enable for all members

### Supported Identity Providers

| Provider | SAML Support | OIDC Support |
|----------|-------------|---------------|
| Okta | ✓ | ✓ |
| Azure AD | ✓ | ✓ |
| Google Workspace | ✓ | ✓ |
| OneLogin | ✓ | — |
| PingFederate | ✓ | — |
| Custom IdP | ✓ | — |

### Just-in-Time Provisioning

When enabled, new users are automatically created when they first authenticate via SSO — no invite needed.

---

## Audit Logs

All organization activity is logged:

| Event Type | Logged |
|------------|--------|
| Member login | ✓ |
| Role changes | ✓ |
| Resource created/modified/deleted | ✓ |
| API key usage | ✓ |
| Billing changes | ✓ |
| SSO configuration changes | ✓ |
| Member invited/removed | ✓ |

### Viewing Audit Logs

1. Go to **Settings → Audit Logs**
2. Filter by:
   - Date range
   - Event type
   - User
   - Resource
3. Export logs (CSV or JSON)

### Log Retention

| Plan | Retention |
|------|-----------|
| Free | 24 hours |
| Starter | 7 days |
| Professional | 90 days |
| Enterprise | 1 year+ (customizable) |

---

## Transferring Ownership

Transfer organization ownership to another member:

1. Go to **Settings → General**
2. Click **Transfer Ownership**
3. Select new owner
4. Confirm with password
5. New owner accepts transfer

**Important:**
- You become an Admin after transfer
- Cannot be undone for 30 days
- All billing responsibility transfers

---

## Organization Deletion

Deleting an organization is **permanent** and irreversible.

**Before deletion:**
1. Download any data you need
2. Cancel all active subscriptions
3. Remove all team members except yourself

**Deletion steps:**
1. Go to **Settings → Danger Zone**
2. Click **Delete Organization**
3. Type organization name to confirm
4. All functions, agents, data are permanently deleted

---

## API Access for Teams

### Organization API Endpoints

```bash
# List organization members
GET /v1/org/members

# Invite a member
POST /v1/org/members/invite

# Update member role
PATCH /v1/org/members/{memberId}

# Remove a member
DELETE /v1/org/members/{memberId}

# Get audit logs
GET /v1/org/audit-logs
```

### Scoped API Keys

Generate keys with limited scope:

```json
{
    "name": "CI/CD Deploy Key",
    "scopes": ["functions:write", "deployments:write"],
    "expires_at": "2027-01-01T00:00:00Z"
}
```

---

## Troubleshooting

### Can't invite a member

**Cause:** You may not have permission (Viewer role)

**Solution:** Ask an Owner or Admin to send the invite.

### Invitation expired

**Solution:** Ask an Admin to resend the invite.

### SSO login not working

**Causes:**
- Misconfigured IdP
- User not in allowed group
- Expired certificate

**Solution:** Check IdP configuration in Settings → Security → SSO.

### Two-factor authentication locked out

**Solution:** Use backup codes sent when MFA was enabled, or contact support for manual reset.

---

## Best Practices

1. **Follow least privilege** — Give members only the permissions they need
2. **Enable MFA** — Require MFA for all organization members
3. **Use SSO** — Centralize identity management for Enterprise
4. **Regular audits** — Review team members and remove inactive users
5. **Scoped API keys** — Use limited-scope keys for CI/CD and automation
6. **Document roles** — Keep a record of who has what role and why
