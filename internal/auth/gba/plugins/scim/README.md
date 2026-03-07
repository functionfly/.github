# SCIM Provisioning Plugin for GoBetterAuth

This plugin provides SCIM 2.0 (System for Cross-domain Identity Management) support for GoBetterAuth. It enables enterprise customers to automatically provision and deprovision users and groups from their Identity Providers (IdPs) such as Okta, Azure AD, OneLogin, and others.

## Features

- **Full SCIM 2.0 compliance** (RFC 7643, RFC 7644)
- **User resource CRUD** - Create, Read, Update, Delete users
- **Group resource CRUD** - Create, Read, Update, Delete groups
- **Patch operations** - Incremental updates via PATCH
- **Filter support** - Query users and groups with SCIM filters
- **Bulk operations** - Process multiple operations in a single request
- **Bearer token authentication** - Secure API access
- **Multi-tenancy support** - Tenant-isolated provisioning

## Installation

The SCIM plugin is included in the GoBetterAuth package.

## Configuration

### Environment Variables

```bash
# Enable SCIM Provisioning
GBA_SCIM_ENABLED=true

# SCIM Base URL
SCIM_BASE_URL=https://app.functionfly.com/v1/scim
```

### Programmatic Configuration

```go
import "github.com/functionfly/functionfly/internal/auth/gba/plugins/scim"

config := &scim.SCIMPluginConfig{
    Enabled:     true,
    BaseURL:     "https://app.functionfly.com/v1/scim",
    TokenExpiry: 365 * 24 * time.Hour,
}

plugin, err := scim.New(db, config, logger)
if err != nil {
    log.Fatal(err)
}
```

## API Endpoints

### SCIM 2.0 Endpoints (RFC 7644 Compliant)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/scim/Users` | List users (with filter support) |
| POST | `/v1/scim/Users` | Create a new user |
| GET | `/v1/scim/Users/{id}` | Get a specific user |
| PUT | `/v1/scim/Users/{id}` | Replace a user |
| PATCH | `/v1/scim/Users/{id}` | Update a user (partial) |
| DELETE | `/v1/scim/Users/{id}` | Delete a user |
| GET | `/v1/scim/Groups` | List groups (with filter support) |
| POST | `/v1/scim/Groups` | Create a new group |
| GET | `/v1/scim/Groups/{id}` | Get a specific group |
| PUT | `/v1/scim/Groups/{id}` | Replace a group |
| PATCH | `/v1/scim/Groups/{id}` | Update a group (partial) |
| DELETE | `/v1/scim/Groups/{id}` | Delete a group |

### Service Provider Configuration Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/scim/ServiceProviderConfig` | SCIM service provider configuration |
| GET | `/v1/scim/ResourceTypes` | Available resource types |
| GET | `/v1/scim/Schemas` | Supported schemas |

### Admin Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/admin/tenants/{id}/scim/token` | Generate SCIM bearer token |
| DELETE | `/admin/tenants/{id}/scim/token` | Revoke SCIM token |
| GET | `/admin/tenants/{id}/scim/config` | Get SCIM configuration |
| PUT | `/admin/tenants/{id}/scim/config` | Update SCIM configuration |

## Authentication

All SCIM endpoints require Bearer token authentication:

```bash
curl https://api.functionfly.com/v1/scim/Users \
  -H "Authorization: Bearer {scim_token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

## Setting up SCIM for a Tenant

### 1. Generate SCIM Token

```bash
curl -X POST https://api.functionfly.com/admin/tenants/{tenant_id}/scim/token \
  -H "Authorization: Bearer {admin_token}"
```

Response:
```json
{
  "token": "scim_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "created_at": "2024-03-07T10:00:00Z"
}
```

**Important**: Save the token securely. It cannot be retrieved again.

### 2. Configure Your IdP

Use the following information to configure SCIM provisioning in your IdP:

- **SCIM Base URL**: `https://app.functionfly.com/v1/scim`
- **Authentication**: Bearer Token
- **Token**: The token generated in step 1
- **Tenant ID Header**: `X-Tenant-ID: {tenant_id}`

### 3. Configure SCIM Settings

```bash
curl -X PUT https://api.functionfly.com/admin/tenants/{tenant_id}/scim/config \
  -H "Authorization: Bearer {admin_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": true,
    "sync_groups": true,
    "sync_users": true
  }'
```

## SCIM Operations

### Create User

```bash
curl -X POST https://api.functionfly.com/v1/scim/Users \
  -H "Authorization: Bearer {scim_token}" \
  -H "X-Tenant-ID: {tenant_id}" \
  -H "Content-Type: application/scim+json" \
  -d '{
    "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
    "userName": "john.doe@example.com",
    "externalId": "user-12345",
    "active": true,
    "emails": [
      {
        "value": "john.doe@example.com",
        "type": "work",
        "primary": true
      }
    ],
    "name": {
      "formatted": "John Doe",
      "givenName": "John",
      "familyName": "Doe"
    }
  }'
```

### Update User

```bash
curl -X PUT https://api.functionfly.com/v1/scim/Users/{user_id} \
  -H "Authorization: Bearer {scim_token}" \
  -H "X-Tenant-ID: {tenant_id}" \
  -H "Content-Type: application/scim+json" \
  -d '{
    "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
    "id": "{user_id}",
    "userName": "john.doe@example.com",
    "active": false
  }'
```

### Patch User (Partial Update)

```bash
curl -X PATCH https://api.functionfly.com/v1/scim/Users/{user_id} \
  -H "Authorization: Bearer {scim_token}" \
  -H "X-Tenant-ID: {tenant_id}" \
  -H "Content-Type: application/scim+json" \
  -d '{
    "schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
    "Operations": [
      {
        "op": "replace",
        "path": "active",
        "value": false
      }
    ]
  }'
```

### List Users with Filter

```bash
curl "https://api.functionfly.com/v1/scim/Users?filter=userName%20eq%20%22john.doe@example.com%22" \
  -H "Authorization: Bearer {scim_token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

### Create Group

```bash
curl -X POST https://api.functionfly.com/v1/scim/Groups \
  -H "Authorization: Bearer {scim_token}" \
  -H "X-Tenant-ID: {tenant_id}" \
  -H "Content-Type: application/scim+json" \
  -d '{
    "schemas": ["urn:ietf:params:scim:schemas:core:2.0:Group"],
    "displayName": "Engineering",
    "externalId": "group-12345",
    "members": [
      {
        "value": "{user_id}",
        "display": "John Doe"
      }
    ]
  }'
```

### Add Member to Group (Patch)

```bash
curl -X PATCH https://api.functionfly.com/v1/scim/Groups/{group_id} \
  -H "Authorization: Bearer {scim_token}" \
  -H "X-Tenant-ID: {tenant_id}" \
  -H "Content-Type: application/scim+json" \
  -d '{
    "schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
    "Operations": [
      {
        "op": "add",
        "path": "members",
        "value": [
          {
            "value": "{user_id}",
            "display": "John Doe"
          }
        ]
      }
    ]
  }'
```

## Database Schema

### gba_scim_configs

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| tenant_id | UUID | Tenant reference |
| enabled | boolean | Whether SCIM is enabled |
| token_hash | text | Bearer token hash (bcrypt) |
| sync_groups | boolean | Enable group sync |
| sync_users | boolean | Enable user sync |
| last_sync_at | timestamp | Last sync timestamp |

### gba_scim_users

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| tenant_id | UUID | Tenant reference |
| external_id | varchar(255) | External IdP user ID |
| user_name | varchar(255) | Username (typically email) |
| display_name | varchar(255) | Display name |
| emails | jsonb | Email addresses |
| active | boolean | User active status |
| groups | jsonb | Group memberships |
| raw | jsonb | Full SCIM JSON |

### gba_scim_groups

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| tenant_id | UUID | Tenant reference |
| external_id | varchar(255) | External IdP group ID |
| display_name | varchar(255) | Group name |
| members | jsonb | Group members |
| raw | jsonb | Full SCIM JSON |

## Supported Identity Providers

- **Microsoft Azure AD / Entra ID**
- **Okta**
- **OneLogin**
- **Google Workspace**
- **Ping Identity**
- **Any SCIM 2.0 compliant IdP**

## Filter Support

The plugin supports basic SCIM filter queries:

| Filter | Example | Description |
|--------|---------|-------------|
| eq | `userName eq "john"` | Equal |
| sw | `userName sw "john"` | Starts with |
| co | `displayName co "john"` | Contains |

## Security Considerations

1. **Token Security**: Store SCIM tokens securely and rotate them regularly
2. **HTTPS Only**: SCIM endpoints should only be accessible over HTTPS
3. **Rate Limiting**: Implement rate limiting to prevent abuse
4. **Audit Logging**: All SCIM operations are logged for security review
5. **Tenant Isolation**: Users and groups are strictly isolated by tenant

## Troubleshooting

### Common Issues

1. **"Invalid token"**
   - Verify the Bearer token is correct
   - Check that the token hasn't been revoked
   - Ensure the `X-Tenant-ID` header matches the token's tenant

2. **"User already exists"**
   - The username must be unique within the tenant
   - Check if the user was already provisioned

3. **"Group not found"**
   - Verify the group ID is correct
   - Ensure the group belongs to the correct tenant

4. **Filter not working**
   - Only basic filters (eq, sw, co) are supported
   - Complex nested filters may not be supported

## References

- [SCIM 2.0 Protocol (RFC 7644)](https://tools.ietf.org/html/rfc7644)
- [SCIM 2.0 Core Schema (RFC 7643)](https://tools.ietf.org/html/rfc7643)
- [SCIM 2.0 REST API](http://www.simplecloud.info/)

## License

This plugin is part of the FunctionFly platform and follows the same licensing terms.
