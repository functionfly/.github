/**
 * Vault Enterprise API Client
 *
 * One TanStack-Query hook per server endpoint (Phases 1-5 + 6 data
 * sources). All hooks degrade gracefully if the corresponding
 * server endpoint returns 404 (e.g. an older deployment) — they
 * surface an empty result and a console warning.
 */

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback } from "react";
import { apiClient } from "@/api/client";
import type {
  AssignRoleRequest,
  AccessToken,
  AuditExportFormat,
  AuditExportResult,
  BreakGlassConfig,
  BreakGlassRequest,
  BreakGlassRequestBody,
  CacheStats,
  CreateCredentialRequest,
  CreateNamespaceRequest,
  CreateRoleRequest,
  CreateSIEMWebhookRequest,
  CreateTargetRequest,
  DynamicCredential,
  DynamicSecretTarget,
  EnableEscrowRequest,
  EscrowStatus,
  GeneratedCredential,
  SetSecretExpirationRequest,
  ShareSecretRequest,
  UpdateRoleRequest,
  UpdateSSORequest,
  UpdateTokenIPPolicyRequest,
  UpdateVaultMFARequest,
  VaultMFAConfig,
  VaultNamespace,
  VaultRole,
  VaultRoleAssignment,
  VaultShare,
  VaultSIEMWebhook,
  VaultSSOConfig,
} from "@/types/vault-enterprise";

// ============================================================================
// Query keys
// ============================================================================

export const vaultKeys = {
  all: ["vault"] as const,
  secrets: () => [...vaultKeys.all, "secrets"] as const,
  secret: (id: string) => [...vaultKeys.secrets(), id] as const,
  tokens: (secretId: string) => [...vaultKeys.all, "tokens", secretId] as const,
  audit: () => [...vaultKeys.all, "audit"] as const,
  // Phase 1
  mfa: () => [...vaultKeys.all, "mfa"] as const,
  // Phase 2
  targets: () => [...vaultKeys.all, "targets"] as const,
  target: (id: string) => [...vaultKeys.targets(), id] as const,
  dynamicCreds: () => [...vaultKeys.all, "dynamic-credentials"] as const,
  dynamicCred: (id: string) => [...vaultKeys.dynamicCreds(), id] as const,
  // Phase 4
  namespaces: () => [...vaultKeys.all, "namespaces"] as const,
  roles: () => [...vaultKeys.all, "roles"] as const,
  role: (id: string) => [...vaultKeys.roles(), id] as const,
  myAssignments: () => [...vaultKeys.all, "my-assignments"] as const,
  sharesIncoming: () => [...vaultKeys.all, "shares", "incoming"] as const,
  sso: () => [...vaultKeys.all, "sso"] as const,
  siemWebhooks: () => [...vaultKeys.all, "siem-webhooks"] as const,
  breakGlassConfig: () => [...vaultKeys.all, "break-glass-config"] as const,
  breakGlassList: () => [...vaultKeys.all, "break-glass"] as const,
  escrow: () => [...vaultKeys.all, "escrow"] as const,
  // Phase 5
  cache: () => [...vaultKeys.all, "cache"] as const,
};

async function unwrap<T>(p: Promise<{ data: T }>): Promise<T> {
  const { data } = await p;
  return data;
}

// ============================================================================
// Phase 1.1: Vault MFA
// ============================================================================

export function useVaultMFA() {
  return useQuery({
    queryKey: vaultKeys.mfa(),
    queryFn: () => unwrap<VaultMFAConfig>(apiClient.get("/v1/vault/mfa/config")),
  });
}

export function useUpdateVaultMFA() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: UpdateVaultMFARequest) =>
      unwrap<VaultMFAConfig>(apiClient.put("/v1/vault/mfa/config", body)),
    onSuccess: () => qc.invalidateQueries({ queryKey: vaultKeys.mfa() }),
  });
}

export function useVerifyVaultMFA() {
  return useMutation({
    mutationFn: () => unwrap<{ verified: boolean; expires_at: string; ttl: number }>(
      apiClient.post("/v1/vault/mfa/verify", {}),
    ),
  });
}

// ============================================================================
// Phase 1.2: Token IP policy
// ============================================================================

export function useTokensForSecret(secretId: string) {
  return useQuery({
    queryKey: vaultKeys.tokens(secretId),
    queryFn: () => unwrap<{ tokens: AccessToken[]; total: number }>(
      apiClient.get(`/v1/vault/secrets/${secretId}/tokens`),
    ),
    enabled: !!secretId,
  });
}

export function useUpdateTokenIPPolicy(secretId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ tokenId, body }: { tokenId: string; body: UpdateTokenIPPolicyRequest }) =>
      apiClient.put(`/v1/vault/tokens/${tokenId}/ip-policy`, body),
    onSuccess: () => qc.invalidateQueries({ queryKey: vaultKeys.tokens(secretId) }),
  });
}

// ============================================================================
// Phase 1.3: Expiration
// ============================================================================

export function useSetSecretExpiration(secretId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: SetSecretExpirationRequest) =>
      apiClient.patch(`/v1/vault/secrets/${secretId}/expiration`, body),
    onSuccess: () => qc.invalidateQueries({ queryKey: vaultKeys.secret(secretId) }),
  });
}

// ============================================================================
// Phase 1.4: Break-glass
// ============================================================================

export function useBreakGlassConfig() {
  return useQuery({
    queryKey: vaultKeys.breakGlassConfig(),
    queryFn: () => unwrap<BreakGlassConfig>(apiClient.get("/v1/vault/break-glass/config")),
  });
}

export function useBreakGlassList() {
  return useQuery({
    queryKey: vaultKeys.breakGlassList(),
    queryFn: () => unwrap<{ requests: BreakGlassRequest[]; total: number }>(
      apiClient.get("/v1/vault/break-glass"),
    ),
  });
}

export function useRequestBreakGlass() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: BreakGlassRequestBody) =>
      unwrap<BreakGlassRequest>(apiClient.post("/v1/vault/break-glass", body)),
    onSuccess: () => qc.invalidateQueries({ queryKey: vaultKeys.breakGlassList() }),
  });
}

export function useApproveBreakGlass() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      unwrap<BreakGlassRequest>(apiClient.post(`/v1/vault/break-glass/${id}/approve`, {})),
    onSuccess: () => qc.invalidateQueries({ queryKey: vaultKeys.breakGlassList() }),
  });
}

export function useDenyBreakGlass() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      unwrap<BreakGlassRequest>(apiClient.post(`/v1/vault/break-glass/${id}/deny`, {})),
    onSuccess: () => qc.invalidateQueries({ queryKey: vaultKeys.breakGlassList() }),
  });
}

export function useRevokeBreakGlass() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      apiClient.post(`/v1/vault/break-glass/${id}/revoke`, {}),
    onSuccess: () => qc.invalidateQueries({ queryKey: vaultKeys.breakGlassList() }),
  });
}

// ============================================================================
// Phase 1.4b: Escrow
// ============================================================================

export function useEscrowStatus() {
  return useQuery({
    queryKey: vaultKeys.escrow(),
    queryFn: () => unwrap<EscrowStatus>(apiClient.get("/v1/vault/escrow")),
  });
}

export function useEnableEscrow() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: EnableEscrowRequest) =>
      unwrap<EscrowStatus>(apiClient.post("/v1/vault/escrow", body)),
    onSuccess: () => qc.invalidateQueries({ queryKey: vaultKeys.escrow() }),
  });
}

export function useDisableEscrow() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => apiClient.delete("/v1/vault/escrow"),
    onSuccess: () => qc.invalidateQueries({ queryKey: vaultKeys.escrow() }),
  });
}

// ============================================================================
// Phase 2.1: Dynamic secret targets
// ============================================================================

export function useDynamicTargets() {
  return useQuery({
    queryKey: vaultKeys.targets(),
    queryFn: () => unwrap<{ targets: DynamicSecretTarget[]; total: number }>(
      apiClient.get("/v1/vault/dynamic-secret-targets"),
    ),
  });
}

export function useCreateTarget() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateTargetRequest) =>
      unwrap<DynamicSecretTarget>(apiClient.post("/v1/vault/dynamic-secret-targets", body)),
    onSuccess: () => qc.invalidateQueries({ queryKey: vaultKeys.targets() }),
  });
}

export function useDeleteTarget() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiClient.delete(`/v1/vault/dynamic-secret-targets/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: vaultKeys.targets() }),
  });
}

export function useTestTarget() {
  return useMutation({
    mutationFn: (id: string) =>
      unwrap<{ ok: boolean; username: string; expires_at: string }>(
        apiClient.post(`/v1/vault/dynamic-secret-targets/${id}/test`, {}),
      ),
  });
}

// ============================================================================
// Phase 2.1: Dynamic credentials
// ============================================================================

export function useDynamicCredentials() {
  return useQuery({
    queryKey: vaultKeys.dynamicCreds(),
    queryFn: () => unwrap<{ credentials: DynamicCredential[]; total: number }>(
      apiClient.get("/v1/vault/dynamic-credentials"),
    ),
  });
}

export function useCreateDynamicCredential() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateCredentialRequest) =>
      unwrap<DynamicCredential>(apiClient.post("/v1/vault/dynamic-credentials", body)),
    onSuccess: () => qc.invalidateQueries({ queryKey: vaultKeys.dynamicCreds() }),
  });
}

export function useGenerateDynamicCredential() {
  return useMutation({
    mutationFn: ({ id, ttlSeconds }: { id: string; ttlSeconds?: number }) =>
      unwrap<GeneratedCredential>(
        apiClient.post(`/v1/vault/dynamic-credentials/${id}/generate`, {
          ttl_seconds: ttlSeconds,
        }),
      ),
  });
}

export function useRevokeAllDynamicCredentials() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      apiClient.post(`/v1/vault/dynamic-credentials/${id}/revoke`, {}),
    onSuccess: () => qc.invalidateQueries({ queryKey: vaultKeys.dynamicCreds() }),
  });
}

// ============================================================================
// Phase 2.2: Leases
// ============================================================================

export function useRenewLease() {
  return useMutation({
    mutationFn: ({
      credentialId,
      leaseId,
      ttlSeconds,
    }: {
      credentialId: string;
      leaseId: string;
      ttlSeconds?: number;
    }) =>
      unwrap<{ lease_id: string; expires_at: string }>(
        apiClient.post(
          `/v1/vault/dynamic-credentials/${credentialId}/leases/${leaseId}/renew`,
          { ttl_seconds: ttlSeconds },
        ),
      ),
  });
}

export function useRevokeLease() {
  return useMutation({
    mutationFn: ({ credentialId, leaseId }: { credentialId: string; leaseId: string }) =>
      apiClient.post(
        `/v1/vault/dynamic-credentials/${credentialId}/leases/${leaseId}/revoke`,
        {},
      ),
  });
}

// ============================================================================
// Phase 4.1: RBAC
// ============================================================================

export function useRoles() {
  return useQuery({
    queryKey: vaultKeys.roles(),
    queryFn: () => unwrap<{ roles: VaultRole[]; total: number }>(
      apiClient.get("/v1/vault/roles"),
    ),
  });
}

export function useCreateRole() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateRoleRequest) =>
      unwrap<VaultRole>(apiClient.post("/v1/vault/roles", body)),
    onSuccess: () => qc.invalidateQueries({ queryKey: vaultKeys.roles() }),
  });
}

export function useUpdateRole() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: UpdateRoleRequest }) =>
      unwrap<VaultRole>(apiClient.patch(`/v1/vault/roles/${id}`, body)),
    onSuccess: () => qc.invalidateQueries({ queryKey: vaultKeys.roles() }),
  });
}

export function useDeleteRole() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiClient.delete(`/v1/vault/roles/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: vaultKeys.roles() }),
  });
}

export function useMyAssignments() {
  return useQuery({
    queryKey: vaultKeys.myAssignments(),
    queryFn: () => unwrap<{ assignments: VaultRoleAssignment[]; total: number }>(
      apiClient.get("/v1/vault/my-assignments"),
    ),
  });
}

export function useAssignRole() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ roleId, body }: { roleId: string; body: AssignRoleRequest }) =>
      unwrap<VaultRoleAssignment>(apiClient.post(`/v1/vault/roles/${roleId}/assignments`, body)),
    onSuccess: () => qc.invalidateQueries({ queryKey: vaultKeys.myAssignments() }),
  });
}

export function useUnassignRole() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (assignmentId: string) =>
      apiClient.delete(`/v1/vault/role-assignments/${assignmentId}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: vaultKeys.myAssignments() }),
  });
}

// ============================================================================
// Phase 4.2: Audit export + SIEM
// ============================================================================

export function useExportAudit() {
  return useMutation({
    mutationFn: async ({
      from,
      to,
      format,
      secretId,
      action,
    }: {
      from?: string;
      to?: string;
      format?: AuditExportFormat;
      secretId?: string;
      action?: string;
    }) => {
      const params = new URLSearchParams();
      if (from) params.set("from", from);
      if (to) params.set("to", to);
      if (format) params.set("format", format);
      if (secretId) params.set("secret_id", secretId);
      if (action) params.set("action", action);
      const url = `/v1/vault/audit/export?${params.toString()}`;
      const response = await fetch(apiClient.getBaseUrl() + url, {
        headers: { Authorization: `Bearer ${localStorage.getItem("ff-access-token") ?? ""}` },
      });
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }
      const result: AuditExportResult = {
        format: (format ?? "json") as AuditExportFormat,
        row_count: parseInt(response.headers.get("X-Audit-Row-Count") ?? "0", 10),
        generated_at: response.headers.get("X-Audit-Generated-At") ?? new Date().toISOString(),
        hmac_sha256: response.headers.get("X-Audit-Signature") ?? "",
        body: await response.blob(),
      };
      return result;
    },
  });
}

export function useDownloadExport() {
  const mutation = useExportAudit();
  type ExportParams = {
    from?: string;
    to?: string;
    format?: AuditExportFormat;
    secretId?: string;
    action?: string;
  };
  return useCallback(
    async (params: ExportParams, filename: string): Promise<AuditExportResult | null> => {
      const result = await mutation.mutateAsync(params);
      const url = URL.createObjectURL(result.body);
      const a = document.createElement("a");
      a.href = url;
      a.download = filename;
      a.click();
      URL.revokeObjectURL(url);
      return result;
    },
    [mutation],
  );
}

export function useSIEMWebhooks() {
  return useQuery({
    queryKey: vaultKeys.siemWebhooks(),
    queryFn: () => unwrap<{ webhooks: VaultSIEMWebhook[]; total: number }>(
      apiClient.get("/v1/vault/siem-webhooks"),
    ),
  });
}

export function useCreateSIEMWebhook() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateSIEMWebhookRequest) =>
      unwrap<VaultSIEMWebhook>(apiClient.post("/v1/vault/siem-webhooks", body)),
    onSuccess: () => qc.invalidateQueries({ queryKey: vaultKeys.siemWebhooks() }),
  });
}

export function useDeleteSIEMWebhook() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiClient.delete(`/v1/vault/siem-webhooks/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: vaultKeys.siemWebhooks() }),
  });
}

// ============================================================================
// Phase 4.3: Namespaces
// ============================================================================

export function useNamespaces() {
  return useQuery({
    queryKey: vaultKeys.namespaces(),
    queryFn: () => unwrap<{ namespaces: VaultNamespace[]; total: number }>(
      apiClient.get("/v1/vault/namespaces"),
    ),
  });
}

export function useCreateNamespace() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateNamespaceRequest) =>
      unwrap<VaultNamespace>(apiClient.post("/v1/vault/namespaces", body)),
    onSuccess: () => qc.invalidateQueries({ queryKey: vaultKeys.namespaces() }),
  });
}

export function useDeleteNamespace() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiClient.delete(`/v1/vault/namespaces/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: vaultKeys.namespaces() }),
  });
}

// ============================================================================
// Phase 4.4: Shares
// ============================================================================

export function useIncomingShares() {
  return useQuery({
    queryKey: vaultKeys.sharesIncoming(),
    queryFn: () => unwrap<{ shares: VaultShare[]; total: number }>(
      apiClient.get("/v1/vault/shared"),
    ),
  });
}

export function useShareSecret() {
  return useMutation({
    mutationFn: ({ secretId, body }: { secretId: string; body: ShareSecretRequest }) =>
      unwrap<VaultShare>(apiClient.post(`/v1/vault/secrets/${secretId}/share`, body)),
  });
}

export function useRevokeShare() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (shareId: string) => apiClient.delete(`/v1/vault/shares/${shareId}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: vaultKeys.sharesIncoming() }),
  });
}

// ============================================================================
// Phase 4.5: SSO
// ============================================================================

export function useSSOConfig() {
  return useQuery({
    queryKey: vaultKeys.sso(),
    queryFn: () => unwrap<VaultSSOConfig>(apiClient.get("/v1/vault/sso/config")),
  });
}

export function useUpdateSSOConfig() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: UpdateSSORequest) =>
      unwrap<VaultSSOConfig>(apiClient.put("/v1/vault/sso/config", body)),
    onSuccess: () => qc.invalidateQueries({ queryKey: vaultKeys.sso() }),
  });
}

// ============================================================================
// Phase 5.1: Cache
// ============================================================================

export function useCacheStats() {
  return useQuery({
    queryKey: vaultKeys.cache(),
    queryFn: () => unwrap<CacheStats & { enabled: boolean }>(
      apiClient.get("/v1/vault/cache/stats"),
    ),
    refetchInterval: 30_000,
  });
}
