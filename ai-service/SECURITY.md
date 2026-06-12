# Security Policy for FlyMind AI Service

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability, please report it responsibly:

1. **DO NOT** create a public GitHub issue for security vulnerabilities
2. Email security concerns to `security@functionfly.com` with:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Any suggested mitigations (if known)
3. Expected response time: 48 hours (acknowledgment), 7 days (initial assessment)
4. We appreciate responsible disclosure and will credit researchers who report valid issues (with permission)

## Security Best Practices

### API Keys

- Never commit API keys to version control
- Use environment variables or secrets managers (Fly.io secrets, AWS Secrets Manager, HashiCorp Vault)
- Rotate API keys regularly
- Use separate keys for development and production

### Database Connections

- Always use SSL/TLS for database connections
- Use individual connection parameters instead of connection URLs with embedded credentials
- Implement connection pooling with appropriate limits

### Redis Security

- Bind Redis to localhost only (unless using managed service with TLS)
- Use password authentication
- Use TLS for managed Redis (Upstash, ElastiCache, etc.)

### gRPC Security

- Always use TLS in production
- Enable authentication interceptor
- Validate API keys on every request

### Input Validation

- RAG paths are validated to prevent path traversal
- All user inputs are validated using Pydantic models
- File paths in RAG are checked against allowed directories

### Dependency Management

- Use `uv sync --frozen` for reproducible builds
- Run regular CVE scans with `pip-audit`
- Generate and track SBOM (Software Bill of Materials)
- Keep dependencies up to date

### Container Security

- Run as non-root user
- Use read-only filesystem where possible
- Drop all capabilities (`cap_drop: ALL`)
- Disable new privileges (`no-new-privileges: true`)
- Use minimal base images (slim or distroless)

### Monitoring

- Monitor for authentication failures
- Set up alerts for anomalous API usage
- Track cost anomalies (unexpected high spending)
- Monitor for rate limit violations

## Security Architecture

### Multi-Tenant Isolation

- Tenant context isolated via `contextvars.ContextVar`
- Resource access checked against tenant ID
- Admin scope required for cross-tenant operations

### API Key Authentication

- API keys validated via Go orchestrator backend
- Keys cached for 60 seconds to reduce latency
- Key status checked (ACTIVE, REVOKED, EXPIRED, SUSPENDED)
- Scope-based authorization (EMBED_READ, EMBED_WRITE, CHAT_WRITE, etc.)

### Audit Logging

- All embedding operations logged with SHA256 hash of input (not plaintext)
- RAG retrieval logged with query hash
- Events buffered and batch-written to database
- Buffer size limited to prevent memory exhaustion

### Rate Limiting

- Per-tenant rate limits enforced
- Redis-backed for distributed deployments
- Local fallback for single-instance deployments
- Token budget tracking for embedding operations

### Security Headers

All responses include security headers:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security: max-age=31536000; includeSubDomains`
- `Content-Security-Policy: default-src 'self'`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Permissions-Policy: accelerometer=(), camera=(), microphone=(), geolocation=()`

## Compliance Notes

### Data Retention

- Audit logs retained per data retention policy
- Cost allocation entries: 90 days for detailed, 7 years for financial aggregates
- Legal holds block deletion of protected data

### PII Protection

- Input text hashed (SHA256) for audit logs, not stored in plaintext
- API key hashes used for cache keys
- RAG query hashes stored, not actual queries

## Changelog

### 1.1.0 (Security Hardening)

- Added path validation for RAG docs directory
- Added security headers middleware
- Added TLS and auth interceptors for gRPC
- Fixed audit buffer unbounded growth vulnerability
- Fixed API key validator async issues
- Fixed token budget reset logic
- Hardened Docker container configuration
- Added CVE scanning CI workflow
- Added SBOM generation to CI pipeline

### 1.0.0 (Initial Release)

- Basic API key authentication
- Multi-provider support (OpenAI, Anthropic, Ollama, etc.)
- Rate limiting
- Audit logging
- RAG support