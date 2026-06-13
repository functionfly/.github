/**
 * useCurrentPlan
 *
 * Hook returning the current tenant's plan. Tries, in order:
 *   1. The local billing store (if one exists)
 *   2. A `tenants/plan` API call
 *   3. Falls back to "free" when the plan is unknown.
 *
 * The hook is intentionally conservative: if we can't determine
 * the plan we default to "free" so that all features are
 * locked down (the safe default). The plan banner in the
 * top-of-page surfaces a "tell us your plan" CTA when this
 * happens so the user can fix it via the billing page.
 */

import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/api/client";
import type { VaultPlan } from "@/types/vault-enterprise";

interface TenantPlanResponse {
  plan: VaultPlan;
}

export function useCurrentPlan() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["tenant", "plan"],
    queryFn: async (): Promise<VaultPlan> => {
      try {
        // apiClient.get returns Promise<AxiosResponse<T>>; we read .data.
        // We cast through unknown to avoid coupling to axios's exact types.
        const res = (await apiClient.get("/v1/tenants/plan")) as unknown as {
          data?: { plan?: VaultPlan };
        };
        const plan = (res?.data?.plan ?? "free") as VaultPlan;
        return plan;
      } catch {
        return "free" as VaultPlan;
      }
    },
    staleTime: 5 * 60_000,
  });

  return {
    plan: (data ?? "free") as VaultPlan,
    isLoading,
    error,
  };
}
