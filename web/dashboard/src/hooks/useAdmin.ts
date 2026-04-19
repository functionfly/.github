import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { tenantApi, auditApi, healthApi, billingApi, analyticsApi } from '@/api/admin';
import type { Tenant, AuditEvent, PricingTier, Coupon, SystemHealth } from '@/api/admin';

// Query keys
export const adminKeys = {
  all: ['admin'] as const,
  tenants: (params?: { limit?: number; offset?: number; status?: string }) =>
    [...adminKeys.all, 'tenants', params] as const,
  tenant: (id: string) => [...adminKeys.all, 'tenant', id] as const,
  audit: (params?: { limit?: number; offset?: number; actor_user_id?: string; tenant_id?: string }) =>
    [...adminKeys.all, 'audit', params] as const,
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
    mutationFn: ({ id, data }: { id: string; data: Partial<Tenant> }) =>
      tenantApi.update(id, data),
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
