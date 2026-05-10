import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  provisionBundle,
  getProvisioningStatus,
  retryProvisioning,
  type ProvisioningResult,
  type NotProvisionedResponse,
} from '@/api/provisioning';

// ─── Query Keys ──────────────────────────────────────────────────────────────

export const provisioningKeys = {
  all: ['provisioning'] as const,
  status: () => [...provisioningKeys.all, 'status'] as const,
};

// ─── Hooks ───────────────────────────────────────────────────────────────────

/**
 * Fetch current provisioning status. Polls every 3s while provisioning is in progress.
 */
export function useProvisioningStatus() {
  return useQuery<ProvisioningResult | NotProvisionedResponse>({
    queryKey: provisioningKeys.status(),
    queryFn: getProvisioningStatus,
    refetchInterval: (query) => {
      const data = query.state.data;
      if (!data || !('components' in data)) return false;
      // Poll while provisioning
      return data.status === 'provisioning' ? 3000 : false;
    },
    retry: false,
    staleTime: 10_000,
  });
}

/**
 * Trigger one-click bundle provisioning.
 */
export function useProvisionBundle() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (bundleSlug: string) => provisionBundle(bundleSlug),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: provisioningKeys.all });
    },
  });
}

/**
 * Retry failed provisioning components.
 */
export function useRetryProvisioning() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: retryProvisioning,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: provisioningKeys.all });
    },
  });
}

/**
 * Returns true if the tenant has an active provisioning state.
 */
export function useIsProvisioned(): boolean {
  const { data } = useProvisioningStatus();
  return !!data && 'status' in data && data.status === 'active';
}

/**
 * Returns the bundle slug if provisioned, null otherwise.
 */
export function useProvisionedBundle(): string | null {
  const { data } = useProvisioningStatus();
  if (!data || !('bundle_slug' in data)) return null;
  return data.bundle_slug;
}
