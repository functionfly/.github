/**
 * Vault API Key Storage Service
 * Provides secure vault-based storage for API keys using client-side encryption.
 *
 * Security:
 * - API keys are encrypted client-side using AES-256-GCM with PBKDF2 key derivation
 * - The encryption passphrase is encrypted with a device-bound CryptoKey (IndexedDB)
 *   before being cached in sessionStorage — never stored in plaintext
 * - Encrypted keys are stored in the vault via vaultApi
 * - Zero-knowledge: server never sees plaintext or passphrase
 */

import { VaultCrypto } from '@/utils/vault-crypto';
import { vaultApi } from '@/api/vault';
import type { CreateSecretRequest, SecretMetadata } from '@/types/vault';
import type { APIKeyCreateResponse } from '@/types/api-key';
import {
  secureStorePassphrase,
  secureGetPassphrase,
  hasStoredPassphrase,
  secureClearPassphrase,
} from '@/services/secure-vault-key-store';

const VAULT_API_KEY_PREFIX = 'ff_api_key:';

export function isVaultPassphraseSet(): boolean {
  return hasStoredPassphrase();
}

export async function setVaultPassphrase(passphrase: string): Promise<void> {
  await secureStorePassphrase(passphrase);
}

export async function getVaultPassphrase(): Promise<string | null> {
  return secureGetPassphrase();
}

export function clearVaultPassphrase(): void {
  secureClearPassphrase();
}

export async function encryptApiKeyForVault(
  apiKey: APIKeyCreateResponse,
  passphrase: string
): Promise<CreateSecretRequest> {
  const plaintext = apiKey.plaintext ?? '';
  const encrypted = await VaultCrypto.encryptWithPassphrase(plaintext, passphrase);

  const secretName = `${VAULT_API_KEY_PREFIX}${apiKey.name || apiKey.id}`;

  return {
    name: secretName,
    description: `Encrypted API key: ${apiKey.name || 'Unnamed'}`,
    secret_type: 'api_key',
    encrypted_data: VaultCrypto.toPayload(encrypted),
    scopes: ['read', 'write'],
    metadata: {
      apiKeyId: apiKey.id,
      keyType: apiKey.key_type,
      createdAt: apiKey.created_at,
    },
  };
}

export async function storeApiKeyInVault(
  apiKey: APIKeyCreateResponse,
  passphrase: string
): Promise<SecretMetadata> {
  const secretRequest = await encryptApiKeyForVault(apiKey, passphrase);
  const created = await vaultApi.createSecret(secretRequest);
  return created;
}

export async function decryptApiKeyFromVault(
  secretId: string,
  passphrase: string
): Promise<string> {
  const result = await vaultApi.decryptSecret(secretId, passphrase);
  return result.value;
}

export async function findApiKeySecretInVault(
  apiKeyId: string
): Promise<SecretMetadata | null> {
  try {
    const secrets = await vaultApi.listSecrets();
    return secrets.find((s) => {
      const meta = s.metadata as Record<string, unknown> | undefined;
      return meta?.apiKeyId === apiKeyId;
    }) || null;
  } catch {
    return null;
  }
}

export async function deleteApiKeyFromVault(apiKeyId: string): Promise<boolean> {
  try {
    const secret = await findApiKeySecretInVault(apiKeyId);
    if (secret) {
      await vaultApi.deleteSecret(secret.id);
      return true;
    }
    return false;
  } catch {
    return false;
  }
}

export interface VaultSetupResult {
  success: boolean;
  error?: string;
}

export async function setupVaultPassphrase(
  passphrase: string
): Promise<VaultSetupResult> {
  try {
    if (passphrase.length < 12) {
      return { success: false, error: 'Passphrase must be at least 12 characters' };
    }

    const testPayload = await VaultCrypto.encryptWithPassphrase('vault-setup-test', passphrase);
    const decrypted = await VaultCrypto.decryptWithPassphrase(
      VaultCrypto.fromPayload(VaultCrypto.toPayload(testPayload)),
      passphrase
    );

    if (decrypted !== 'vault-setup-test') {
      return { success: false, error: 'Passphrase verification failed' };
    }

    await setVaultPassphrase(passphrase);
    return { success: true };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Failed to setup vault passphrase',
    };
  }
}

/**
 * No-op: vault API is the source of truth for which keys are stored.
 * Previously stored metadata in sessionStorage which leaked key IDs to same-origin scripts.
 */
export function markVaultApiKeyStored(_apiKeyId: string): void {
  // intentionally empty — use vaultApi.findApiKeySecretInVault() instead
}
