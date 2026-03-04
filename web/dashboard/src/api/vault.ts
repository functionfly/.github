/**
 * Vault API Client - Secrets management API
 * Mirrors internal/api/handlers/vault endpoints
 */

import { apiClient } from "./client";
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
   * Decrypt and retrieve secret value
   * Note: This is a convenience endpoint for server-side decryption (if KMS/HSM is used)
   * For client-side encryption, use vault-crypto.ts utilities
   */
  decryptSecret: async (id: string): Promise<{ value: string; secret_type: string }> => {
    return apiClient.post<{ value: string; secret_type: string }>(`/v1/vault/secrets/${id}/decrypt`, {});
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
   * Revoke an access token
   */
  revokeToken: async (tokenId: string): Promise<void> => {
    await apiClient.post<void>(`/v1/vault/tokens/${tokenId}/revoke`, {});
  },

  /**
   * Get token details
   */
  getToken: async (tokenId: string): Promise<AccessToken> => {
    return apiClient.get<AccessToken>(`/v1/vault/tokens/${tokenId}`);
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
};

/**
 * Admin Vault API methods for administrative operations
 */
export const adminVaultApi = {
  /**
   * List all secrets across tenants (admin only)
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
   */
  rotateKeys: async (): Promise<{ success: boolean; message: string }> => {
    return apiClient.post<{ success: boolean; message: string }>("/v1/admin/vault/rotate-keys", {});
  },
};
