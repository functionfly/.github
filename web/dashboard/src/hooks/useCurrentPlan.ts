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
import { platformToVaultPlan } from "@/lib/vaultPlans";
import type { VaultPlan } from "@/types/vault-enterprise";

export function useCurrentPlan() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["tenant", "plan", "v2"],
    queryFn: async (): Promise<{ raw: string; vault: VaultPlan }> => {
      try {
        const res = (await apiClient.get("/v1/tenants/plan")) as unknown as {
          plan?: string;
        };
        const raw = (res?.plan ?? "free").toLowerCase();
        return { raw, vault: platformToVaultPlan(raw) };
      } catch {
        return { raw: "free", vault: "free" as VaultPlan };
      }
    },
    staleTime: 5 * 60_000,
  });

  return {
    plan: data?.vault ?? ("free" as VaultPlan),
    rawPlan: data?.raw ?? "free",
    isLoading,
    error,
  };
}
