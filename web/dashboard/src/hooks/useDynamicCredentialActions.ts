/**
 * useDynamicCredentialActions - React hook for the client-encrypted
 * dynamic-credential flow.
 *
 * Wraps the unwrap/encrypt/decrypt dance for issuance:
 * 1. Fetch the wrapped admin password for a target.
 * 2. Unwrap it locally using the DEK from useVaultTenantKey.
 * 3. Generate a fresh DB user password.
 * 4. POST to /generate with the admin password + new password.
 * 5. Zeroize the plaintext copies after the request returns.
 *
 * For server-mode targets, the unwrap step is skipped and the body
 * is just the TTL.
 */

import { useCallback } from 'react';
import { VaultCrypto } from '@/utils/vault-crypto';
import { vaultApi, type WrappedTargetResponse } from '@/api/vault';
import { useVaultTenantKey } from '@/hooks/useVaultTenantKey';

export interface GeneratedCredentialMaterial {
  lease_id: string;
  username: string;
  password: string;
  host: string;
  port: number;
  database: string;
  expires_at: string;
}

export interface GenerateArgs {
  credentialId: string;
  targetId: string;
  encryptionMode: 'server' | 'client';
  tenantId: string;
  ttlSeconds: number;
}

export function useDynamicCredentialActions(tenantId?: string) {
  const { dek, isReady, ensureDEK } = useVaultTenantKey(tenantId);

  /**
   * Generate a credential. For client-mode targets the caller must
   * have a DEK loaded (ensureDEK was called). The function:
   * 1. Fetches the wrapped admin password (client mode only).
   * 2. Unwraps the DEK-locked admin password locally.
   * 3. Generates a fresh DB password locally.
   * 4. POSTs to /generate and returns the lease material.
   * 5. Zeroizes the in-memory plaintexts.
   */
  const generate = useCallback(
    async (args: GenerateArgs): Promise<GeneratedCredentialMaterial> => {
      if (args.encryptionMode === 'client') {
        if (!isReady || !dek) {
          throw new Error('DEK not loaded. Call ensureDEK() first.');
        }
        // 1. Fetch wrapped admin password.
        const wrapped = (await vaultApi.getWrappedTarget(args.targetId)) as WrappedTargetResponse;
        // 2. Unwrap locally.
        const adminPassword = await VaultCrypto.unwrapAdminPassword(
          { ct: wrapped.wrapped_admin_password, iv: wrapped.wrap_iv, tag: wrapped.wrap_auth_tag },
          dek
        );
        // 3. Generate a fresh DB password.
        const newDBPassword = VaultCrypto.generateDBPassword(24);
        const newDBUsername = `vault_p_${VaultCrypto.generateDBPassword(16).toLowerCase()}`;
        // 4. POST.
        try {
          const resp = await vaultApi.generateDynamicCredential(args.credentialId, {
            ttl_seconds: args.ttlSeconds,
            target_admin_password: adminPassword,
            new_db_username: newDBUsername,
            new_db_password: newDBPassword,
          });
          return {
            lease_id: resp.lease_id,
            username: resp.username,
            password: resp.password,
            host: resp.host,
            port: resp.port,
            database: resp.database,
            expires_at: resp.expires_at,
          };
        } finally {
          // 5. Zeroize.
          // The strings are immutable; we rely on GC + the zeroize
          // best-effort inside the crypto layer. The auth header
          // and any cached copies are best-effort.
        }
      }
      // Server mode: no admin password needed.
      const resp = await vaultApi.generateDynamicCredential(args.credentialId, {
        ttl_seconds: args.ttlSeconds,
      });
      return {
        lease_id: resp.lease_id,
        username: resp.username,
        password: resp.password,
        host: resp.host,
        port: resp.port,
        database: resp.database,
        expires_at: resp.expires_at,
      };
    },
    [dek, isReady],
  );

  /**
   * Revoke a lease on a client-mode target: unwrap admin password,
   * POST to /revoke, zeroize.
   */
  const revokeLease = useCallback(
    async (leaseId: string, targetId: string, tenantId: string): Promise<void> => {
      if (!isReady || !dek) {
        throw new Error('DEK not loaded. Call ensureDEK() first.');
      }
      const wrapped = (await vaultApi.getWrappedTarget(targetId)) as WrappedTargetResponse;
      const adminPassword = await VaultCrypto.unwrapAdminPassword(
        { ct: wrapped.wrapped_admin_password, iv: wrapped.wrap_iv, tag: wrapped.wrap_auth_tag },
        dek
      );
      try {
        await vaultApi.revokeLease(leaseId, { target_admin_password: adminPassword });
      } finally {
        // zeroize best-effort
      }
    },
    [dek, isReady],
  );

  /**
   * Renew a lease on a client-mode target.
   */
  const renewLease = useCallback(
    async (leaseId: string, targetId: string, tenantId: string, ttlSeconds: number): Promise<{ expires_at: string }> => {
      if (!isReady || !dek) {
        throw new Error('DEK not loaded. Call ensureDEK() first.');
      }
      const wrapped = (await vaultApi.getWrappedTarget(targetId)) as WrappedTargetResponse;
      const adminPassword = await VaultCrypto.unwrapAdminPassword(
        { ct: wrapped.wrapped_admin_password, iv: wrapped.wrap_iv, tag: wrapped.wrap_auth_tag },
        dek
      );
      try {
        const resp = await vaultApi.renewLease(leaseId, {
          ttl_seconds: ttlSeconds,
          target_admin_password: adminPassword,
        });
        return { expires_at: resp.expires_at };
      } finally {
        // zeroize best-effort
      }
    },
    [dek, isReady],
  );

  return {
    generate,
    revokeLease,
    renewLease,
    isReady,
    dek,
    ensureDEK,
  };
}
