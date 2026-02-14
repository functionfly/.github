# Security Model (MVP1)

## Tenancy

- Tenant owns apps.
- Apps own backends.

## Auth

- Dashboard uses JWT.
- Programmatic use uses app keys.

## Request signing

- HMAC signing between orchestrator and edge target endpoints.
- Prevents unauthorized traffic injection.

## Rate limiting

- Edge entry: per appSlug at Caddy.
- API: per token and per IP.

## Secrets

- MVP1 does not store provider tokens.
- Store only shared secrets for edge target signing.
- Later: encrypted provider tokens as opt-in.

