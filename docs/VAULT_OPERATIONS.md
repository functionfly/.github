# Secrets Vault – Operations and Retention

This document covers operational aspects of the Secrets Vault: audit log retention and token cleanup.

## Audit log retention

- **Table:** `secrets_audit_log`
- **Recommendation:** Retain audit entries for at least **90 days** (compliance); **1 year** for stricter requirements.
- **Implementation options:**
  1. **Scheduled job:** Run a cron or background job (e.g. daily) that deletes rows where `created_at < NOW() - INTERVAL '90 days'` (or your chosen retention). Use a single `DELETE` with a limit/batch if the table is large.
  2. **Partitioning:** If using PostgreSQL 10+, partition `secrets_audit_log` by month on `created_at` and drop old partitions instead of deleting rows.
- **Immutability:** The application only inserts into the audit log. For stricter compliance, consider a DB trigger that blocks `UPDATE`/`DELETE` on `secrets_audit_log`, or an append-only role.

## Token cleanup

- **Table:** `secret_access_tokens`
- **Repository method:** `internal/storage/vault.Repository.CleanupExpiredTokens(ctx, olderThan)`
- **Recommendation:** Run a scheduled job (e.g. daily) that calls `CleanupExpiredTokens` with `olderThan = 30 * 24 * time.Hour` (30 days after expiry) so expired and revoked tokens are pruned.
- The API startup or a separate worker should invoke this; see “Token cleanup job” in the codebase for where to hook it.

## Key versions (client-side crypto)

- **key_version 1:** PBKDF2 with 100,000 iterations (legacy). Existing secrets remain decryptable.
- **key_version 2:** PBKDF2 with 600,000 iterations (OWASP-style). Used for new secrets created by the dashboard (`web/dashboard/src/utils/vault-crypto.ts`). The server stores `key_version` and does not decrypt; the client chooses iterations based on it when decrypting.

## Related

- **Vault security (zero-knowledge):** See AGENTS.md “Vault security” and `internal/api/handlers/vault` — the server never decrypts secrets; encryption/decryption is client-side only.
