# SAML SSO Plugin for GoBetterAuth

This plugin provides SAML 2.0 Single Sign-On authentication support for GoBetterAuth. It enables enterprise customers to authenticate users through their existing Identity Providers (IdPs) such as Okta, Azure AD, OneLogin, and others.

## Features

- **SAML 2.0 Service Provider (SP)** implementation
- **Identity Provider (IdP) initiated SSO** support
- **Service Provider initiated SSO** support
- **Assertion Consumer Service (ACS)** endpoint
- **SP Metadata** endpoint for easy IdP configuration
- **Single Logout (SLO)** support
- **Attribute mapping** for user profile synchronization
- **Auto-provisioning** of users on first login
- **Tenant-specific IdP configurations** for multi-tenancy

## Installation

The SAML plugin is included in the GoBetterAuth package. Ensure the `github.com/crewjam/saml` library is in your go.mod (or use the built-in XML handling).

## Configuration

### Environment Variables

```bash
# Enable SAML SSO
GBA_SAML_ENABLED=true

# Service Provider Configuration
SAML_SP_ENTITY_ID=https://app.functionfly.com
SAML_ACS_URL=https://app.functionfly.com/v1/auth/saml/acs
```

### Programmatic Configuration

```go
import "github.com/functionfly/functionfly/internal/auth/gba/plugins/saml"

config := &saml.SAMLPluginConfig{
    Enabled:        true,
    EntityID:       "https://app.functionfly.com",
    ACSURL:         "https://app.functionfly.com/v1/auth/saml/acs",
    AutoProvision:  true,
    SyncAttributes: true,
    AttributeMapping: &saml.SAMLAttributeMapping{
        Email:      "email",
        FirstName:  "firstName",
        LastName:   "lastName",
        Groups:     "groups",
        Department: "department",
    },
}

plugin, err := saml.New(db, config, logger)
if err != nil {
    log.Fatal(err)
}
```

## API Endpoints

### Public Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/auth/saml/metadata/{tenant_id}` | SP Metadata XML for IdP configuration |
| GET | `/v1/auth/saml/login/{tenant_id}` | Initiate SAML SSO |
| POST | `/v1/auth/saml/acs/{tenant_id}` | Assertion Consumer Service (ACS) |
| POST/GET | `/v1/auth/saml/slo/{tenant_id}` | Single Logout Service |
| GET | `/v1/auth/saml/status/{tenant_id}` | SAML configuration status |

### Admin Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/auth/saml/admin/config/{tenant_id}` | Get SAML configuration |
| PUT | `/v1/auth/saml/admin/config/{tenant_id}` | Create/Update SAML configuration |
| DELETE | `/v1/auth/saml/admin/config/{tenant_id}` | Delete SAML configuration |

## Setting up SAML for a Tenant

### 1. Configure IdP in Your Application

```bash
curl -X PUT https://api.functionfly.com/v1/auth/saml/admin/config/{tenant_id} \
  -H "Authorization: Bearer {admin_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "idp_entity_id": "https://login.microsoftonline.com/{tenant-id}/v2.0",
    "idp_sso_url": "https://login.microsoftonline.com/{tenant-id}/saml2",
    "idp_certificate": "-----BEGIN CERTIFICATE-----\nMIID...\n-----END CERTIFICATE-----",
    "sp_entity_id": "https://app.functionfly.com",
    "acs_url": "https://app.functionfly.com/v1/auth/saml/acs",
    "name_id_format": "emailAddress"
  }'
```

### 2. Get SP Metadata for IdP Configuration

```bash
curl https://api.functionfly.com/v1/auth/saml/metadata/{tenant_id}
```

### 3. Configure Your IdP

Use the SP metadata XML to configure your IdP (Okta, Azure AD, OneLogin, etc.). Configure the following attributes to be sent in the SAML assertion:

- `email` - User's email address (required)
- `firstName` - User's first name (optional)
- `lastName` - User's last name (optional)
- `groups` - User's group memberships (optional)

### 4. Test SSO

```bash
# Initiate SSO (redirects to IdP)
curl -L https://api.functionfly.com/v1/auth/saml/login/{tenant_id}
```

## Database Schema

### gba_saml_configs

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| tenant_id | UUID | Tenant reference |
| enabled | boolean | Whether SAML is enabled |
| idp_entity_id | varchar(500) | IdP Entity ID |
| idp_sso_url | varchar(500) | IdP SSO URL |
| idp_certificate | text | IdP X.509 certificate (PEM) |
| sp_entity_id | varchar(500) | SP Entity ID |
| acs_url | varchar(500) | ACS URL |
| name_id_format | varchar(100) | NameID format |

### gba_saml_sessions

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| tenant_id | UUID | Tenant reference |
| user_id | UUID | User reference |
| name_id | varchar(255) | SAML NameID |
| session_index | varchar(255) | SAML SessionIndex |
| expires_at | timestamp | Session expiration |

## Supported Identity Providers

- **Microsoft Azure AD / Entra ID**
- **Okta**
- **OneLogin**
- **Google Workspace**
- **Ping Identity**
- **Any SAML 2.0 compliant IdP**

## Security Considerations

1. **Certificate Validation**: Always validate IdP certificates
2. **Assertion Signing**: Require signed assertions
3. **Clock Skew**: Allow for minor clock differences between SP and IdP
4. **Secure Cookies**: Use secure, httpOnly cookies for sessions
5. **Token Expiry**: Set appropriate token expiration times
6. **Audit Logging**: All SAML operations are logged for security review

## Troubleshooting

### Common Issues

1. **"SAML not configured for tenant"**
   - Verify SAML configuration exists for the tenant
   - Check that `enabled` is set to `true`

2. **"Invalid SAML response"**
   - Verify IdP certificate is correct and not expired
   - Check that SAML response is properly encoded
   - Ensure clock synchronization between SP and IdP

3. **"No email found in SAML assertion"**
   - Verify IdP is configured to send email attribute
   - Check attribute mapping configuration

4. **"User account not found"**
   - Enable auto-provisioning or manually create user accounts
   - Verify email addresses match between IdP and application

## License

This plugin is part of the FunctionFly platform and follows the same licensing terms.
