import type { Coupon, MatchResultRequest, PricingTier, Tenant, WarMatch } from '@/api/admin';
import { analyticsApi, auditApi, billingApi, cityWarsApi, healthApi, tenantApi } from '@/api/admin';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

// Query keys
export const adminKeys = {
  all: ['admin'] as const,
  tenants: (params?: { limit?: number; offset?: number; status?: string }) =>
    [...adminKeys.all, 'tenants', params] as const,
  tenant: (id: string) => [...adminKeys.all, 'tenant', id] as const,
  audit: (params?: {
    limit?: number;
    offset?: number;
    actor_user_id?: string;
    tenant_id?: string;
  }) => [...adminKeys.all, 'audit', params] as const,
  auditEvent: (id: string) => [...adminKeys.all, 'audit-event', id] as const,
  health: () => [...adminKeys.all, 'health'] as const,
  pricingTiers: () => [...adminKeys.all, 'pricing-tiers'] as const,
  invoices: (tenantId: string) => [...adminKeys.all, 'invoices', tenantId] as const,
  coupons: () => [...adminKeys.all, 'coupons'] as const,
  analytics: (params?: { start_date?: string; end_date?: string }) =>
    [...adminKeys.all, 'analytics', params] as const,
};

// ==================== Tenant Hooks ====================

export function useTenants(params?: { limit?: number; offset?: number; status?: string }) {
  return useQuery({
    queryKey: adminKeys.tenants(params),
    queryFn: () => tenantApi.list(params),
    staleTime: 1000 * 60,
  });
}

export function useTenant(id: string) {
  return useQuery({
    queryKey: adminKeys.tenant(id),
    queryFn: () => tenantApi.get(id),
    enabled: !!id,
    staleTime: 1000 * 60,
  });
}

export function useUpdateTenant() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Tenant> }) => tenantApi.update(id, data),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: adminKeys.tenant(id) });
      queryClient.invalidateQueries({ queryKey: adminKeys.tenants() });
      toast.success('Tenant updated successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update tenant: ${error.message}`);
    },
  });
}

export function useSuspendTenant() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => tenantApi.suspend(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: adminKeys.tenant(id) });
      queryClient.invalidateQueries({ queryKey: adminKeys.tenants() });
      toast.success('Tenant suspended successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to suspend tenant: ${error.message}`);
    },
  });
}

export function useActivateTenant() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => tenantApi.activate(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: adminKeys.tenant(id) });
      queryClient.invalidateQueries({ queryKey: adminKeys.tenants() });
      toast.success('Tenant activated successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to activate tenant: ${error.message}`);
    },
  });
}

// ==================== Audit Hooks ====================

export function useAuditEvents(params?: {
  limit?: number;
  offset?: number;
  actor_user_id?: string;
  tenant_id?: string;
  action?: string;
  resource_type?: string;
}) {
  return useQuery({
    queryKey: adminKeys.audit(params),
    queryFn: () => auditApi.list(params),
    staleTime: 1000 * 30,
  });
}

export function useAuditEvent(id: string) {
  return useQuery({
    queryKey: adminKeys.auditEvent(id),
    queryFn: () => auditApi.get(id),
    enabled: !!id,
    staleTime: 1000 * 60,
  });
}

// ==================== Health Hooks ====================

export function useSystemHealth() {
  return useQuery({
    queryKey: adminKeys.health(),
    queryFn: () => healthApi.getStatus(),
    staleTime: 1000 * 30,
    refetchInterval: 30000, // Refresh every 30s
  });
}

// ==================== Billing Hooks ====================

export function usePricingTiers() {
  return useQuery({
    queryKey: adminKeys.pricingTiers(),
    queryFn: () => billingApi.listPricingTiers(),
    staleTime: 1000 * 60 * 5,
  });
}

export function useCreatePricingTier() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: Omit<PricingTier, 'id' | 'created_at' | 'updated_at'>) =>
      billingApi.createPricingTier(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: adminKeys.pricingTiers() });
      toast.success('Pricing tier created successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create pricing tier: ${error.message}`);
    },
  });
}

export function useUpdatePricingTier() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<PricingTier> }) =>
      billingApi.updatePricingTier(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: adminKeys.pricingTiers() });
      toast.success('Pricing tier updated successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update pricing tier: ${error.message}`);
    },
  });
}

export function useDeletePricingTier() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => billingApi.deletePricingTier(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: adminKeys.pricingTiers() });
      toast.success('Pricing tier deleted successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete pricing tier: ${error.message}`);
    },
  });
}

export function useAdminInvoices(tenantId: string) {
  return useQuery({
    queryKey: adminKeys.invoices(tenantId),
    queryFn: () => billingApi.listInvoices(tenantId),
    enabled: !!tenantId,
    staleTime: 1000 * 60,
  });
}

export function useCoupons() {
  return useQuery({
    queryKey: adminKeys.coupons(),
    queryFn: () => billingApi.listCoupons(),
    staleTime: 1000 * 60 * 5,
  });
}

export function useCreateCoupon() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: Omit<Coupon, 'id' | 'current_uses' | 'created_at' | 'updated_at'>) =>
      billingApi.createCoupon(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: adminKeys.coupons() });
      toast.success('Coupon created successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create coupon: ${error.message}`);
    },
  });
}

// ==================== Analytics Hooks ====================

export function useAdminAnalytics(params?: { start_date?: string; end_date?: string }) {
  return useQuery({
    queryKey: adminKeys.analytics(params),
    queryFn: () => analyticsApi.getOverview(params),
    staleTime: 1000 * 60,
  });
}

// ==================== City Wars Admin Hooks ====================

export const cityWarsKeys = {
  all: ['cityWars'] as const,
  wars: (params?: { status?: string }) => [...cityWarsKeys.all, 'wars', params] as const,
  war: (id: number) => [...cityWarsKeys.all, 'war', id] as const,
  eligibleMetros: () => [...cityWarsKeys.all, 'eligible-metros'] as const,
  bracket: (id: number) => [...cityWarsKeys.all, 'bracket', id] as const,
};

export function useCityWars(params?: { status?: string }) {
  return useQuery({
    queryKey: cityWarsKeys.wars(params),
    queryFn: () => cityWarsApi.listWars(params),
    staleTime: 1000 * 30,
  });
}

export function useCityWar(id: number) {
  return useQuery({
    queryKey: cityWarsKeys.war(id),
    queryFn: () => cityWarsApi.getWar(id),
    enabled: !!id,
    staleTime: 1000 * 30,
  });
}

export function useEligibleMetros() {
  return useQuery({
    queryKey: cityWarsKeys.eligibleMetros(),
    queryFn: () => cityWarsApi.getEligibleMetros(),
    staleTime: 1000 * 60 * 5,
  });
}

export function useCreateWar() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: {
      name: string;
      season: string;
      slug: string;
      starts_at: string;
      ends_at: string;
    }) => cityWarsApi.createWar(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: cityWarsKeys.wars() });
      toast.success('War created successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create war: ${error.message}`);
    },
  });
}

export function useActivateWar() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => cityWarsApi.activateWar(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: cityWarsKeys.wars() });
      toast.success('War activated successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to activate war: ${error.message}`);
    },
  });
}

export function useCancelWar() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => cityWarsApi.cancelWar(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: cityWarsKeys.wars() });
      toast.success('War cancelled successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to cancel war: ${error.message}`);
    },
  });
}

export function useGenerateBracket() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => cityWarsApi.generateBracket(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: cityWarsKeys.wars() });
      toast.success('Bracket generated successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to generate bracket: ${error.message}`);
    },
  });
}

export function useSetQuarterfinals() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      matches,
    }: {
      id: number;
      matches: Omit<WarMatch, 'id' | 'war_id' | 'status' | 'completed_at'>[];
    }) => cityWarsApi.setQuarterfinals(id, matches),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: cityWarsKeys.wars() });
      toast.success('Quarterfinals set successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to set quarterfinals: ${error.message}`);
    },
  });
}

export function useAdvanceWar() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, round }: { id: number; round: 'semifinal' | 'final' }) =>
      cityWarsApi.advanceWar(id, round),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: cityWarsKeys.wars() });
      toast.success('War advanced successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to advance war: ${error.message}`);
    },
  });
}

export function useOverrideMatch() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: MatchResultRequest }) =>
      cityWarsApi.overrideMatch(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: cityWarsKeys.wars() });
      toast.success('Match overridden successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to override match: ${error.message}`);
    },
  });
}

export function useRecordMatch() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: MatchResultRequest }) =>
      cityWarsApi.recordMatch(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: cityWarsKeys.wars() });
      toast.success('Match recorded successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to record match: ${error.message}`);
    },
  });
}
