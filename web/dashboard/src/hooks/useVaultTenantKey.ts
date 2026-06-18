/**
 * useVaultTenantKey - React hook that manages the per-tenant DEK
 * (data encryption key) used for client-encrypted dynamic credentials.
 *
 * The DEK is generated client-side at first vault unlock, wrapped
 * under the user's passphrase-derived KEK, and stored on the server
 * as opaque ciphertext. The plaintext DEK is held in a module-level
 * Map keyed by tenantID for the lifetime of the browser tab.
 *
 * Usage:
 *   const { dek, ensureDEK, rotateDEK, shareDEK, isReady } = useVaultTenantKey();
 *
 *   // At first unlock:
 *   await ensureDEK(passphrase);
 *
 *   // Before generating a credential:
 *   if (!isReady) throw new Error("DEK not loaded");
 *   const adminPwd = await VaultCrypto.unwrapAdminPassword(...);
 */

import { useCallback, useEffect, useState } from 'react';
import { VaultCrypto, type EncryptedData } from '@/utils/vault-crypto';
import { vaultApi } from '@/api/vault';

// Module-level cache: tenantID -> DEK (Uint8Array). Cleared on tab close.
const dekCache = new Map<string, Uint8Array>();

export interface WrappedDEKResponse {
  tenant_id: string;
  user_id: string;
  wrapped_dek: string;
  dek_iv: string;
  dek_auth_tag: string;
  dek_salt: string;
  key_version: number;
  kdf_params: Record<string, unknown>;
  created_at: string;
  rotated_at?: string | null;
}

export interface UseVaultTenantKeyState {
  dek: Uint8Array | null;
  isReady: boolean;
  isLoading: boolean;
  error: string | null;
  hasServerRecord: boolean | null;
}

export interface UseVaultTenantKeyApi extends UseVaultTenantKeyState {
  ensureDEK: (passphrase: string) => Promise<Uint8Array>;
  rotateDEK: (passphrase: string) => Promise<Uint8Array>;
  shareDEK: (
    targetUserId: string,
    sourcePassphrase: string,
    newPassphraseForTarget: string,
  ) => Promise<void>;
  clearDEK: (tenantId?: string) => void;
  getDEK: (tenantId: string) => Uint8Array | null;
}

function setCache(tenantId: string, dek: Uint8Array) {
  dekCache.set(tenantId, dek);
}

function clearCache(tenantId?: string) {
  if (!tenantId) {
    dekCache.clear();
    return;
  }
  const cur = dekCache.get(tenantId);
  if (cur) {
    // Zeroize the buffer before dropping the reference.
    for (let i = 0; i < cur.length; i++) cur[i] = 0;
  }
  dekCache.delete(tenantId);
}

export function useVaultTenantKey(tenantId?: string): UseVaultTenantKeyApi {
  const [state, setState] = useState<UseVaultTenantKeyState>({
    dek: dekCache.get(tenantId || '') || null,
    isReady: false,
    isLoading: false,
    error: null,
    hasServerRecord: null,
  });

  // Refresh isReady whenever the underlying cache changes.
  useEffect(() => {
    if (!tenantId) return;
    const cur = dekCache.get(tenantId);
    setState((s) => ({
      ...s,
      dek: cur || null,
      isReady: !!cur,
    }));
  }, [tenantId]);

  const ensureDEK = useCallback(
    async (passphrase: string): Promise<Uint8Array> => {
      if (!tenantId) throw new Error('useVaultTenantKey: tenantId is required');
      setState((s) => ({ ...s, isLoading: true, error: null }));
      try {
        let dek: Uint8Array;
        try {
          const wrapped = (await vaultApi.getWrappedDEK()) as WrappedDEKResponse | null;
          if (!wrapped) {
            // No server record — generate, wrap, upload.
            dek = VaultCrypto.generateDEK();
            const enc = await VaultCrypto.wrapDEK(dek, passphrase);
            await vaultApi.upsertWrappedDEK({
              wrapped_dek: enc.ciphertext,
              dek_iv: enc.iv,
              dek_auth_tag: enc.tag,
              dek_salt: enc.salt,
              key_version: enc.keyVersion,
              kdf_params: { t: 3, m: 65536, p: 4 },
            });
            setCache(tenantId, dek);
            setState((s) => ({
              ...s,
              dek,
              isReady: true,
              isLoading: false,
              hasServerRecord: true,
            }));
            return dek;
          }
          // Server has a record — unwrap.
          const wrappedData: EncryptedData = {
            ciphertext: wrapped.wrapped_dek,
            iv: wrapped.dek_iv,
            tag: wrapped.dek_auth_tag,
            salt: wrapped.dek_salt,
            keyVersion: wrapped.key_version,
          };
          dek = await VaultCrypto.unwrapDEK(wrappedData, passphrase);
          setCache(tenantId, dek);
          setState((s) => ({
            ...s,
            dek,
            isReady: true,
            isLoading: false,
            hasServerRecord: true,
          }));
          return dek;
        } catch (err) {
          setState((s) => ({
            ...s,
            isLoading: false,
            error: err instanceof Error ? err.message : String(err),
          }));
          throw err;
        }
      } catch (err) {
        throw err;
      }
    },
    [tenantId],
  );

  const rotateDEK = useCallback(
    async (passphrase: string): Promise<Uint8Array> => {
      if (!tenantId) throw new Error('useVaultTenantKey: tenantId is required');
      setState((s) => ({ ...s, isLoading: true, error: null }));
      try {
        const dek = VaultCrypto.generateDEK();
        const enc = await VaultCrypto.wrapDEK(dek, passphrase);
        await vaultApi.upsertWrappedDEK({
          wrapped_dek: enc.ciphertext,
          dek_iv: enc.iv,
          dek_auth_tag: enc.tag,
          dek_salt: enc.salt,
          key_version: enc.keyVersion,
          kdf_params: { t: 3, m: 65536, p: 4 },
        });
        setCache(tenantId, dek);
        setState((s) => ({ ...s, dek, isReady: true, isLoading: false }));
        return dek;
      } catch (err) {
        setState((s) => ({
          ...s,
          isLoading: false,
          error: err instanceof Error ? err.message : String(err),
        }));
        throw err;
      }
    },
    [tenantId],
  );

  const shareDEK = useCallback(
    async (
      targetUserId: string,
      sourcePassphrase: string,
      newPassphraseForTarget: string,
    ): Promise<void> => {
      if (!tenantId) throw new Error('useVaultTenantKey: tenantId is required');
      const cur = dekCache.get(tenantId);
      let dek: Uint8Array;
      if (cur) {
        dek = cur;
      } else {
        const wrapped = (await vaultApi.getWrappedDEK()) as WrappedDEKResponse | null;
        if (!wrapped) {
          throw new Error('No DEK to share: source user has not set up the vault');
        }
        const wrappedData: EncryptedData = {
          ciphertext: wrapped.wrapped_dek,
          iv: wrapped.dek_iv,
          tag: wrapped.dek_auth_tag,
          salt: wrapped.dek_salt,
          keyVersion: wrapped.key_version,
        };
        dek = await VaultCrypto.unwrapDEK(wrappedData, sourcePassphrase);
        setCache(tenantId, dek);
      }
      // Re-wrap the DEK under the target user's passphrase.
      const rewrapped = await VaultCrypto.wrapDEK(dek, newPassphraseForTarget);
      await vaultApi.shareDEK({
        target_user_id: targetUserId,
        wrapped_dek: rewrapped.ciphertext,
        dek_iv: rewrapped.iv,
        dek_auth_tag: rewrapped.tag,
        dek_salt: rewrapped.salt,
        key_version: rewrapped.keyVersion,
      } as unknown as Parameters<typeof vaultApi.shareDEK>[0]);
    },
    [tenantId],
  );

  const clearDEK = useCallback(
    (t?: string) => {
      clearCache(t || tenantId);
      setState((s) => ({ ...s, dek: null, isReady: false }));
    },
    [tenantId],
  );

  const getDEK = useCallback(
    (t: string) => dekCache.get(t) || null,
    [],
  );

  return {
    ...state,
    ensureDEK,
    rotateDEK,
    shareDEK,
    clearDEK,
    getDEK,
  };
}

export function _internalForTests() {
  return { dekCache, clearCache };
}
