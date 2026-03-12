/**
 * Vault React Query Hooks
 * Provides hooks for secrets management with caching and optimistic updates
 * Pattern follows useStateFabric.ts
 */

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { vaultApi } from "@/api/vault";
import type {
  Secret,
  SecretMetadata,
  AccessToken,
  AuditLogEntry,
  CreateSecretRequest,
  UpdateSecretRequest,
  GenerateTokenRequest,
  GenerateTokenResponse,
} from "@/types/vault";

// ==================== Query Keys ====================

export const vaultKeys = {
  all: ["vault"] as const,
  lists: () => [...vaultKeys.all, "list"] as const,
  list: (filters: string) => [...vaultKeys.lists(), { filters }] as const,
  details: () => [...vaultKeys.all, "detail"] as const,
  detail: (id: string) => [...vaultKeys.details(), id] as const,
  tokens: (secretId: string) => [...vaultKeys.detail(secretId), "tokens"] as const,
  audit: () => [...vaultKeys.all, "audit"] as const,
  secretAudit: (secretId: string) => [...vaultKeys.detail(secretId), "audit"] as const,
};

// ==================== Helper Functions ====================

/**
 * Check if secret ID is valid for fetching
 */
function isSecretIdValidForFetch(id: string): boolean {
  return !!id && id !== "new" && id.length > 0;
}

// ==================== Query Hooks ====================

/**
 * Hook to list all secrets (metadata only)
 */
export function useVaultSecrets() {
  return useQuery({
    queryKey: vaultKeys.lists(),
    queryFn: () => vaultApi.listSecrets(),
  });
}

/**
 * Hook to get a single secret by ID
 */
export function useVaultSecret(id: string) {
  return useQuery({
    queryKey: vaultKeys.detail(id),
    queryFn: () => vaultApi.getSecret(id),
    enabled: isSecretIdValidForFetch(id),
  });
}

/**
 * Hook to get audit log entries
 */
export function useVaultAuditLog(limit?: number) {
  return useQuery({
    queryKey: [...vaultKeys.audit(), { limit }],
    queryFn: () => vaultApi.getAuditLog(limit),
  });
}

/**
 * Hook to get audit log for a specific secret
 */
export function useSecretAuditLog(secretId: string, limit?: number) {
  return useQuery({
    queryKey: [...vaultKeys.secretAudit(secretId), { limit }],
    queryFn: () => vaultApi.getSecretAuditLog(secretId, limit),
    enabled: isSecretIdValidForFetch(secretId),
  });
}

/**
 * Hook to list access tokens for a secret
 */
export function useSecretTokens(secretId: string) {
  return useQuery({
    queryKey: vaultKeys.tokens(secretId),
    queryFn: () => vaultApi.listTokens(secretId),
    enabled: isSecretIdValidForFetch(secretId),
  });
}

// ==================== Mutation Hooks ====================

/**
 * Hook to create a new secret
 */
export function useCreateSecret() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateSecretRequest) => vaultApi.createSecret(data),
    onSuccess: (created) => {
      queryClient.invalidateQueries({ queryKey: vaultKeys.lists() });
      toast.success(`Secret "${created.name}" created successfully`);
    },
    onError: (error: Error) => {
      // Check for secret limit exceeded error
      const errorMessage = error.message || '';
      if (
        errorMessage.includes('SECRET_LIMIT_EXCEEDED') ||
        errorMessage.includes('403') ||
        errorMessage.toLowerCase().includes('limit')
      ) {
        toast.error(
          "You've reached your secrets limit. Upgrade your plan to create more secrets.",
          {
            duration: 5000,
            action: {
              label: 'View Plans',
              onClick: () => window.location.href = '/pricing'
            }
          }
        );
      } else {
        toast.error(`Failed to create secret: ${error.message}`);
      }
    },
  });
}

/**
 * Hook to update an existing secret
 */
export function useUpdateSecret(id: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateSecretRequest) => vaultApi.updateSecret(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: vaultKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: vaultKeys.lists() });
      toast.success("Secret updated successfully");
    },
    onError: (error: Error) => {
      toast.error(`Failed to update secret: ${error.message}`);
    },
  });
}

/**
 * Hook to delete a secret
 */
export function useDeleteSecret() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => vaultApi.deleteSecret(id),
    onSuccess: (_, deletedId) => {
      queryClient.invalidateQueries({ queryKey: vaultKeys.lists() });
      queryClient.removeQueries({ queryKey: vaultKeys.detail(deletedId) });
      toast.success("Secret deleted successfully");
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete secret: ${error.message}`);
    },
  });
}

/**
 * Hook to generate an access token for a secret
 */
export function useGenerateToken(secretId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: GenerateTokenRequest) => vaultApi.generateToken(secretId, data),
    onSuccess: (response) => {
      queryClient.invalidateQueries({ queryKey: vaultKeys.tokens(secretId) });
      toast.success(`Token "${response.name || response.token_id}" generated successfully`);
      // Important: The token is only shown once, handled by the component
    },
    onError: (error: Error) => {
      // Check for token limit exceeded error
      const errorMessage = error.message || '';
      if (
        errorMessage.includes('TOKEN_LIMIT_EXCEEDED') ||
        errorMessage.includes('403') ||
        errorMessage.toLowerCase().includes('token') && errorMessage.toLowerCase().includes('limit')
      ) {
        toast.error(
          "You've reached your token limit for this secret. Upgrade your plan to create more tokens.",
          {
            duration: 5000,
            action: {
              label: 'View Plans',
              onClick: () => window.location.href = '/pricing'
            }
          }
        );
      } else {
        toast.error(`Failed to generate token: ${error.message}`);
      }
    },
  });
}

/**
 * Hook to revoke an access token
 */
export function useRevokeToken() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (tokenId: string) => vaultApi.revokeToken(tokenId),
    onSuccess: () => {
      // Invalidate all token queries since we don't know which secret this belongs to
      queryClient.invalidateQueries({ queryKey: vaultKeys.all });
      toast.success("Token revoked successfully");
    },
    onError: (error: Error) => {
      toast.error(`Failed to revoke token: ${error.message}`);
    },
  });
}

// ==================== Utility Hooks ====================

/**
 * Hook to decrypt a secret value (for server-side decryption with KMS/HSM)
 * For client-side encryption, use VaultCrypto utilities directly
 */
export function useDecryptSecret() {
  return useMutation({
    mutationFn: (id: string) => vaultApi.decryptSecret(id),
    onError: (error: Error) => {
      toast.error(`Failed to decrypt secret: ${error.message}`);
    },
  });
}
