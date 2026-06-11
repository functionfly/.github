/**
 * Vault API Client - Secrets management API
 * Mirrors internal/api/handlers/vault endpoints.
 *
 * Security: Decryption is client-side only. The server never sees plaintext.
 * Use decryptSecret(id, passphrase) to fetch encrypted data and decrypt locally.
 */

import { apiClient } from "./client";
import { VaultCrypto } from "@/utils/vault-crypto";
import type {
  Secret,
  SecretMetadata,
  EncryptedDataPayload,
  AccessToken,
  AuditLogEntry,
  CreateSecretRequest,
  UpdateSecretRequest,
  GenerateTokenRequest,
  GenerateTokenResponse,
  ListSecretsResponse,
  ListTokensResponse,
  ListAuditLogResponse,
  SecretVersion,
  SecretVersionMetadata,
  SecretVersionDiff,
  RollbackSecretRequest,
  RollbackSecretResponse,
  ListSecretVersionsResponse,
} from "@/types/vault";

/**
 * Vault API methods for secret management
 */
export const vaultApi = {
  /**
   * List all secrets (metadata only, no encrypted data)
   */
  listSecrets: async (): Promise<SecretMetadata[]> => {
    const response = await apiClient.get<ListSecretsResponse>("/v1/vault/secrets");
    return response.secrets;
  },

  /**
   * Get a single secret with encrypted data
   */
  getSecret: async (id: string): Promise<Secret> => {
    return apiClient.get<Secret>(`/v1/vault/secrets/${id}`);
  },

  /**
   * Create a new secret
   */
  createSecret: async (data: CreateSecretRequest): Promise<Secret> => {
    return apiClient.post<Secret>("/v1/vault/secrets", data);
  },

  /**
   * Update an existing secret
   */
  updateSecret: async (id: string, data: UpdateSecretRequest): Promise<Secret> => {
    return apiClient.patch<Secret>(`/v1/vault/secrets/${id}`, data);
  },

  /**
   * Delete a secret
   */
  deleteSecret: async (id: string): Promise<void> => {
    await apiClient.delete<void>(`/v1/vault/secrets/${id}`);
  },

  /**
   * Decrypt secret value client-side (no server decrypt; zero-knowledge).
   * Fetches the secret's encrypted payload then decrypts with VaultCrypto + passphrase.
   * The server never receives or sees the passphrase or plaintext.
   */
  decryptSecret: async (
    id: string,
    passphrase: string
  ): Promise<{ value: string; secret_type: string }> => {
    const secret = await apiClient.get<Secret>(`/v1/vault/secrets/${id}`);
    const encryptedData = VaultCrypto.fromPayload(secret.encrypted_data);
    const value = await VaultCrypto.decryptWithPassphrase(encryptedData, passphrase);
    return { value, secret_type: secret.secret_type };
  },

  // ==================== Access Tokens ====================

  /**
   * List access tokens for a secret
   */
  listTokens: async (secretId: string): Promise<AccessToken[]> => {
    const response = await apiClient.get<ListTokensResponse>(`/v1/vault/secrets/${secretId}/tokens`);
    return response.tokens;
  },

  /**
   * Generate a new access token for a secret
   */
  generateToken: async (
    secretId: string,
    data: GenerateTokenRequest
  ): Promise<GenerateTokenResponse> => {
    return apiClient.post<GenerateTokenResponse>(`/v1/vault/secrets/${secretId}/tokens`, data);
  },

  /**
   * Revoke an access token (DELETE /v1/vault/tokens/{id})
   */
  revokeToken: async (tokenId: string): Promise<void> => {
    await apiClient.delete<void>(`/v1/vault/tokens/${tokenId}`);
  },

  // ==================== Audit Log ====================

  /**
   * Get vault audit log
   */
  getAuditLog: async (limit: number = 100): Promise<AuditLogEntry[]> => {
    const response = await apiClient.get<ListAuditLogResponse>(`/v1/vault/audit?limit=${limit}`);
    return response.entries;
  },

  /**
   * Get audit log for a specific secret
   */
  getSecretAuditLog: async (secretId: string, limit: number = 50): Promise<AuditLogEntry[]> => {
    const response = await apiClient.get<ListAuditLogResponse>(
      `/v1/vault/secrets/${secretId}/audit?limit=${limit}`
    );
    return response.entries;
  },

  // ==================== Secret Versions ====================

  /**
   * List all versions for a secret
   */
  listSecretVersions: async (secretId: string, limit: number = 50, offset: number = 0): Promise<SecretVersionMetadata[]> => {
    const response = await apiClient.get<ListSecretVersionsResponse>(
      `/v1/vault/secrets/${secretId}/versions?limit=${limit}&offset=${offset}`
    );
    return response.versions;
  },

  /**
   * Get a specific version of a secret
   */
  getSecretVersion: async (secretId: string, versionNumber: number, includeEncrypted: boolean = false): Promise<SecretVersion> => {
    return apiClient.get<SecretVersion>(
      `/v1/vault/secrets/${secretId}/versions/${versionNumber}?include_encrypted=${includeEncrypted}`
    );
  },

  /**
   * Compare (diff) two versions of a secret
   */
  diffSecretVersions: async (
    secretId: string,
    fromVersion: number,
    toVersion?: number
  ): Promise<SecretVersionDiff> => {
    const params = new URLSearchParams();
    params.append('from_version', fromVersion.toString());
    if (toVersion !== undefined) {
      params.append('to_version', toVersion.toString());
    }
    return apiClient.get<SecretVersionDiff>(
      `/v1/vault/secrets/${secretId}/versions/diff?${params.toString()}`
    );
  },

  /**
   * Rollback a secret to a previous version
   */
  rollbackSecret: async (secretId: string, request: RollbackSecretRequest): Promise<RollbackSecretResponse> => {
    return apiClient.post<RollbackSecretResponse>(
      `/v1/vault/secrets/${secretId}/rollback`,
      request
    );
  },

  /**
   * Rotate a secret's encrypted value (re-encrypt with new ciphertext)
   */
  rotateSecret: async (secretId: string, data: { encrypted_data: EncryptedDataPayload; reason?: string }): Promise<Secret> => {
    return apiClient.patch<Secret>(
      `/v1/vault/secrets/${secretId}/rotate`,
      data
    );
  },

  // ==================== Bulk Operations ====================

  /**
   * Bulk delete multiple secrets
   */
  bulkDeleteSecrets: async (secretIds: string[], dryRun: boolean = false): Promise<{
    deleted: number;
    failed: number;
    errors?: { secret_id: string; error: string }[];
    previews?: Record<string, {
      secret_id: string;
      secret_name: string;
      found: boolean;
      tokens_count: number;
      dependencies: { id: string; type: string; name: string; criticality: string }[];
    }>;
  }> => {
    return apiClient.delete<{
      deleted: number;
      failed: number;
      errors?: { secret_id: string; error: string }[];
      previews?: Record<string, {
        secret_id: string;
        secret_name: string;
        found: boolean;
        tokens_count: number;
        dependencies: { id: string; type: string; name: string; criticality: string }[];
      }>;
    }>(`/v1/vault/secrets/bulk-delete`, {
      data: { secret_ids: secretIds, dry_run: dryRun },
    });
  },

  /**
   * Export all secrets for the tenant (metadata only, no encrypted values)
   */
  exportSecrets: async (): Promise<{
    secrets: {
      id: string;
      name: string;
      description?: string;
      secret_type: string;
      key_version: number;
      scopes?: string[];
      metadata?: Record<string, unknown>;
      created_at: string;
      updated_at: string;
    }[];
    total: number;
    exported_at: string;
  }> => {
    return apiClient.get<{
      secrets: {
        id: string;
        name: string;
        description?: string;
        secret_type: string;
        key_version: number;
        scopes?: string[];
        metadata?: Record<string, unknown>;
        created_at: string;
        updated_at: string;
      }[];
      total: number;
      exported_at: string;
    }>(`/v1/vault/secrets/export`);
  },

  // ==================== Secret Dependencies ====================

  /**
   * Get dependencies for a secret (services/functions that depend on it)
   */
  getSecretDependencies: async (secretId: string): Promise<{
    secret_id: string;
    dependencies: {
      id: string;
      dependent_id: string;
      dependent_type: string;
      dependent_name: string;
      criticality: string;
      metadata?: Record<string, unknown>;
      created_at: string;
    }[];
    total: number;
  }> => {
    return apiClient.get<{
      secret_id: string;
      dependencies: {
        id: string;
        dependent_id: string;
        dependent_type: string;
        dependent_name: string;
        criticality: string;
        metadata?: Record<string, unknown>;
        created_at: string;
      }[];
      total: number;
    }>(`/v1/vault/secrets/${secretId}/dependencies`);
  },

  /**
   * Create a secret dependency record
   */
  createSecretDependency: async (
    secretId: string,
    data: {
      dependent_id: string;
      dependent_type: string;
      dependent_name: string;
      criticality?: string;
      metadata?: Record<string, unknown>;
    }
  ): Promise<{
    id: string;
    dependent_id: string;
    dependent_type: string;
    dependent_name: string;
    criticality: string;
    metadata?: Record<string, unknown>;
    created_at: string;
  }> => {
    return apiClient.post<{
      id: string;
      dependent_id: string;
      dependent_type: string;
      dependent_name: string;
      criticality: string;
      metadata?: Record<string, unknown>;
      created_at: string;
    }>(`/v1/vault/secrets/${secretId}/dependencies`, data);
  },

  /**
   * Delete a secret dependency record
   */
  deleteSecretDependency: async (secretId: string, dependencyId: string): Promise<void> => {
    await apiClient.delete<void>(`/v1/vault/secrets/${secretId}/dependencies/${dependencyId}`);
  },
};

/**
 * Admin Vault API methods for administrative operations.
 * NOTE: These endpoints require admin/vault routes to be implemented on the backend.
 * Currently stubbed — call only after confirming the admin vault routes exist.
 */
export const adminVaultApi = {
  /**
   * List all secrets across tenants (admin only)
   * @warning Requires /v1/admin/vault/secrets route on backend
   */
  listAllSecrets: async (params?: {
    tenantId?: string;
    limit?: number;
    offset?: number;
  }): Promise<ListSecretsResponse> => {
    const queryParams = new URLSearchParams();
    if (params?.tenantId) queryParams.append("tenantId", params.tenantId);
    if (params?.limit) queryParams.append("limit", params.limit.toString());
    if (params?.offset) queryParams.append("offset", params.offset.toString());

    return apiClient.get<ListSecretsResponse>(`/v1/admin/vault/secrets?${queryParams.toString()}`);
  },

  /**
   * Get vault statistics (admin only)
   * @warning Requires /v1/admin/vault/stats route on backend
   */
  getStats: async (): Promise<{
    totalSecrets: number;
    totalAccessTokens: number;
    activeAccessTokens: number;
    revokedAccessTokens: number;
    totalAuditEntries: number;
  }> => {
    return apiClient.get<{
      totalSecrets: number;
      totalAccessTokens: number;
      activeAccessTokens: number;
      revokedAccessTokens: number;
      totalAuditEntries: number;
    }>("/v1/admin/vault/stats");
  },

  /**
   * Rotate encryption keys (admin only)
   * @warning Requires /v1/admin/vault/rotate-keys route on backend
   */
  rotateKeys: async (): Promise<{ success: boolean; message: string }> => {
    return apiClient.post<{ success: boolean; message: string }>("/v1/admin/vault/rotate-keys", {});
  },
};
