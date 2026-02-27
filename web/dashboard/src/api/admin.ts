import { apiClient } from './client';
import type {
  FunctionConfig,
  FunctionDeployment,
  FunctionLog
} from "@/types";

// Types
export interface Tenant {
  id: string;
  name: string;
  plan: string;
  status: 'active' | 'suspended';
  created_at: string;
  updated_at: string;
}

export interface AuditEvent {
  id: string;
  actor_user_id?: string;
  actor_email?: string;
  tenant_id?: string;
  action: string;
  resource_type: string;
  resource_id?: string;
  request_id?: string;
  before_state?: any;
  after_state?: any;
  ip_address?: string;
  user_agent?: string;
  timestamp: string;
  success: boolean;
}

export interface SystemHealth {
  status: 'healthy' | 'unhealthy';
  version: string;
  timestamp: string;
  checks: {
    database: { status: string; healthy: boolean; response_time_ms?: number };
    api: { status: string; healthy: boolean };
    repository: { status: string; healthy: boolean; response_time_ms?: number };
    system: { status: string; healthy: boolean; goroutines?: number };
  };
}

export interface PricingTier {
  id: string;
  name: string;
  description: string;
  price_cents: number;
  currency: string;
  features: any;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Subscription {
  id: string;
  tenant_id: string;
  pricing_tier_id: string;
  status: string;
  current_period_start: string;
  current_period_end: string;
  cancel_at_period_end: boolean;
  created_at: string;
  updated_at: string;
}

export interface Invoice {
  id: string;
  tenant_id: string;
  subscription_id?: string;
  status: string;
  amount_due_cents: number;
  amount_paid_cents: number;
  currency: string;
  period_start?: string;
  period_end?: string;
  paid_at?: string;
  created_at: string;
  updated_at: string;
}

export interface Coupon {
  id: string;
  code: string;
  name: string;
  description: string;
  discount_type: 'percent' | 'fixed';
  discount_value: number;
  max_redemptions?: number;
  times_redeemed: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

// Tenant Management API
export const tenantApi = {
  listTenants: async (): Promise<{ tenants: Tenant[] }> => {
    const response = await apiClient.get<{ tenants: Tenant[] }>('/v1/admin/tenants');
    return response;
  },

  getTenant: async (tenantId: string): Promise<Tenant> => {
    const response = await apiClient.get<Tenant>(`/v1/admin/tenants/${tenantId}`);
    return response;
  },

  createTenant: async (name: string): Promise<Tenant> => {
    const response = await apiClient.post<Tenant>('/v1/admin/tenants', { name });
    return response;
  },

  updateTenant: async (tenantId: string, updates: Partial<Tenant>): Promise<Tenant> => {
    const response = await apiClient.patch<Tenant>(`/v1/admin/tenants/${tenantId}`, updates);
    return response;
  },

  deleteTenant: async (tenantId: string): Promise<void> => {
    await apiClient.delete(`/v1/admin/tenants/${tenantId}`);
  },
};

// Audit Events API
export const auditApi = {
  listAuditEvents: async (params?: {
    limit?: number;
    offset?: number;
    actor_user_id?: string;
    actor_email?: string;
    tenant_id?: string;
    action?: string;
    resource_type?: string;
    resource_id?: string;
    success?: boolean;
    start_time?: string;
    end_time?: string;
  }): Promise<{ events: AuditEvent[]; limit: number; offset: number; filters: any }> => {
    const searchParams = new URLSearchParams();
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.offset) searchParams.set('offset', params.offset.toString());
    if (params?.actor_user_id) searchParams.set('actor_user_id', params.actor_user_id);
    if (params?.actor_email) searchParams.set('actor_email', params.actor_email);
    if (params?.tenant_id) searchParams.set('tenant_id', params.tenant_id);
    if (params?.action) searchParams.set('action', params.action);
    if (params?.resource_type) searchParams.set('resource_type', params.resource_type);
    if (params?.resource_id) searchParams.set('resource_id', params.resource_id);
    if (params?.success !== undefined) searchParams.set('success', params.success.toString());
    if (params?.start_time) searchParams.set('start_time', params.start_time);
    if (params?.end_time) searchParams.set('end_time', params.end_time);

    const response = await apiClient.get<{
      events: AuditEvent[];
      limit: number;
      offset: number;
      filters: any
    }>(`/v1/admin/audit-events?${searchParams}`);
    return response;
  },
};

// System Health API
export const healthApi = {
  getSystemHealth: async (): Promise<SystemHealth> => {
    const response = await apiClient.get<SystemHealth>('/v1/admin/health');
    return response;
  },
};

// Billing API
export const billingApi = {
  // Pricing Tiers
  listPricingTiers: async (): Promise<{ tiers: PricingTier[] }> => {
    const response = await apiClient.get<{ tiers: PricingTier[] }>('/v1/admin/billing/tiers');
    return response;
  },

  createPricingTier: async (tier: Omit<PricingTier, 'id' | 'created_at' | 'updated_at'>): Promise<PricingTier> => {
    const response = await apiClient.post<PricingTier>('/v1/admin/billing/tiers', tier);
    return response;
  },

  updatePricingTier: async (tierId: string, updates: Partial<PricingTier>): Promise<PricingTier> => {
    const response = await apiClient.patch<PricingTier>(`/v1/admin/billing/tiers/${tierId}`, updates);
    return response;
  },

  deletePricingTier: async (tierId: string): Promise<void> => {
    await apiClient.delete(`/v1/admin/billing/tiers/${tierId}`);
  },

  // Subscriptions
  listSubscriptions: async (): Promise<{ subscriptions: Subscription[] }> => {
    const response = await apiClient.get<{ subscriptions: Subscription[] }>('/v1/admin/billing/subscriptions');
    return response;
  },

  createSubscription: async (subscription: {
    tenant_id: string;
    pricing_tier_id: string;
    trial_end?: string;
  }): Promise<Subscription> => {
    const response = await apiClient.post<Subscription>('/v1/admin/billing/subscriptions', subscription);
    return response;
  },

  updateSubscription: async (subscriptionId: string, updates: Partial<Subscription>): Promise<Subscription> => {
    const response = await apiClient.patch<Subscription>(`/v1/admin/billing/subscriptions/${subscriptionId}`, updates);
    return response;
  },

  cancelSubscription: async (subscriptionId: string): Promise<{ status: string }> => {
    const response = await apiClient.post<{ status: string }>(`/v1/admin/billing/subscriptions/${subscriptionId}/cancel`);
    return response;
  },

  // Invoices
  listInvoices: async (): Promise<{ invoices: Invoice[] }> => {
    const response = await apiClient.get<{ invoices: Invoice[] }>('/v1/admin/billing/invoices');
    return response;
  },

  createInvoice: async (invoice: {
    tenant_id: string;
    subscription_id?: string;
    amount_due_cents: number;
    period_start?: string;
    period_end?: string;
    due_date?: string;
  }): Promise<Invoice> => {
    const response = await apiClient.post<Invoice>('/v1/admin/billing/invoices', invoice);
    return response;
  },

  updateInvoice: async (invoiceId: string, updates: Partial<Invoice>): Promise<Invoice> => {
    const response = await apiClient.patch<Invoice>(`/v1/admin/billing/invoices/${invoiceId}`, updates);
    return response;
  },

  // Coupons
  listCoupons: async (): Promise<{ coupons: Coupon[] }> => {
    const response = await apiClient.get<{ coupons: Coupon[] }>('/v1/admin/billing/coupons');
    return response;
  },

  createCoupon: async (coupon: {
    code: string;
    name: string;
    description: string;
    discount_type: 'percent' | 'fixed';
    discount_value: number;
    max_redemptions?: number;
    valid_from?: string;
    valid_until?: string;
  }): Promise<Coupon> => {
    const response = await apiClient.post<Coupon>('/v1/admin/billing/coupons', coupon);
    return response;
  },

  redeemCoupon: async (couponId: string, redemption: {
    tenant_id: string;
    subscription_id?: string;
  }): Promise<any> => {
    const response = await apiClient.post(`/v1/admin/billing/coupons/${couponId}/redeem`, redemption);
    return response;
  },
};

// Analytics Settings API
export const analyticsApi = {
  getSettings: async (): Promise<any> => {
    const response = await apiClient.get('/v1/admin/analytics');
    return response;
  },

  updateSettings: async (settings: any): Promise<any> => {
    const response = await apiClient.patch('/v1/admin/analytics', settings);
    return response;
  },
};

export interface Incident {
  id: string;
  title: string;
  severity: 'critical' | 'high' | 'medium' | 'low';
  status: 'resolved' | 'investigating' | 'monitoring';
  description: string;
  created_at: string;
  resolved_at?: string;
  updated_at: string;
}

export const incidentsApi = {
  listIncidents: async (params?: {
    limit?: number;
    offset?: number;
    status?: 'resolved' | 'investigating' | 'monitoring';
  }): Promise<{ incidents: Incident[] }> => {
    const queryParams = new URLSearchParams();
    if (params?.limit) queryParams.append('limit', params.limit.toString());
    if (params?.offset) queryParams.append('offset', params.offset.toString());
    if (params?.status) queryParams.append('status', params.status);

    const query = queryParams.toString();
    const url = query ? `/v1/admin/incidents?${query}` : '/v1/admin/incidents';

    return await apiClient.get<{ incidents: Incident[] }>(url);
  },

  getIncident: async (incidentId: string): Promise<Incident> => {
    return await apiClient.get<Incident>(`/v1/admin/incidents/${incidentId}`);
  },

  createIncident: async (incident: {
    title: string;
    severity: 'critical' | 'high' | 'medium' | 'low';
    description: string;
  }): Promise<Incident> => {
    return await apiClient.post<Incident>('/v1/admin/incidents', incident);
  },

  updateIncident: async (incidentId: string, updates: Partial<{
    title: string;
    severity: 'critical' | 'high' | 'medium' | 'low';
    status: 'resolved' | 'investigating' | 'monitoring';
    description: string;
  }>): Promise<Incident> => {
    return await apiClient.patch<Incident>(`/v1/admin/incidents/${incidentId}`, updates);
  },

  resolveIncident: async (incidentId: string): Promise<Incident> => {
    return await apiClient.post<Incident>(`/v1/admin/incidents/${incidentId}/resolve`);
  },
};

// Admin Users API
export const adminUsersApi = {
  // List all users across all tenants
  listUsers: async (params?: {
    limit?: number;
    offset?: number;
    tenant_id?: string;
    plan?: string;
    role?: string;
    search?: string;
    is_active?: boolean;
  }): Promise<{ users: AdminUser[]; total: number }> => {
    const searchParams = new URLSearchParams();
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.offset) searchParams.set('offset', params.offset.toString());
    if (params?.tenant_id) searchParams.set('tenant_id', params.tenant_id);
    if (params?.plan) searchParams.set('plan', params.plan);
    if (params?.role) searchParams.set('role', params.role);
    if (params?.search) searchParams.set('search', params.search);
    if (params?.is_active !== undefined) searchParams.set('is_active', params.is_active.toString());

    const query = searchParams.toString();
    const url = query ? `/v1/admin/users?${query}` : '/v1/admin/users';
    return await apiClient.get<{ users: AdminUser[]; total: number }>(url);
  },

  // Get user statistics
  getUserStats: async (): Promise<AdminUserStats> => {
    return await apiClient.get<AdminUserStats>('/v1/admin/users/stats');
  },

  // Get a specific user
  getUser: async (userId: string): Promise<AdminUser> => {
    return await apiClient.get<AdminUser>(`/v1/admin/users/${userId}`);
  },

  // Update a user
  updateUser: async (userId: string, updates: Partial<AdminUser>): Promise<AdminUser> => {
    return await apiClient.patch<AdminUser>(`/v1/admin/users/${userId}`, updates);
  },

  // Delete/Deactivate a user
  deactivateUser: async (userId: string): Promise<{ success: boolean; message: string }> => {
    return await apiClient.post<{ success: boolean; message: string }>(`/v1/admin/users/${userId}/deactivate`);
  },

  // Reactivate a user
  activateUser: async (userId: string): Promise<{ success: boolean; message: string }> => {
    return await apiClient.post<{ success: boolean; message: string }>(`/v1/admin/users/${userId}/activate`);
  },

  // Reset user password (sends reset email)
  resetUserPassword: async (userId: string): Promise<{ success: boolean; message: string }> => {
    return await apiClient.post<{ success: boolean; message: string }>(`/v1/admin/users/${userId}/reset-password`);
  },

  // Create a new user
  createUser: async (user: {
    email: string;
    name?: string;
    tenant_id: string;
    plan: string;
    role?: string;
  }): Promise<AdminUser> => {
    return await apiClient.post<AdminUser>('/v1/admin/users', user);
  },
};

// Admin Function Types
export interface AdminFunction {
  id: string;
  name: string;
  tenant_id: string;
  tenant_name?: string;
  providers: string[];
  region: string;
  code: string;
  env_vars: Array<{ key: string; value: string; is_secret: boolean }>;
  version: string;
  status: 'draft' | 'deploying' | 'deployed' | 'failed';
  created_at: string;
  updated_at: string;
}

export interface AdminFunctionDeployment {
  id: string;
  function_id: string;
  version: string;
  status: 'pending' | 'deploying' | 'success' | 'failed';
  provider: string;
  region: string;
  deployed_url?: string;
  error_message?: string;
  created_at: string;
  updated_at: string;
}

export interface AdminFunctionLog {
  id: string;
  function_id: string;
  deployment_id?: string;
  level: 'info' | 'warn' | 'error' | 'debug';
  message: string;
  timestamp: string;
  source: string;
  metadata?: Record<string, any>;
}

// Admin User Types
export interface AdminUser {
  id: string;
  email: string;
  name?: string;
  tenant_id: string;
  plan: string;
  role?: string;
  created_at: string;
  updated_at?: string;
  last_login_at?: string;
  is_active: boolean;
}

export interface AdminUserStats {
  total_users: number;
  active_users: number;
  admin_users: number;
  inactive_users: number;
  users_by_plan: Record<string, number>;
  users_by_role: Record<string, number>;
}

// Admin Registry Types
export interface AdminRegistryFunction {
  id: string;
  author: string;
  name: string;
  title?: string;
  description?: string;
  category?: string;
  tags: string[];
  visibility: 'public' | 'private' | 'unlisted';
  price_per_call: number;
  popularity_score: number;
  reliability_score: number;
  deterministic_score: number;
  latest_version?: string;
  total_ratings: number;
  overall_score: number;
  created_at: string;
  updated_at: string;
  is_flagged: boolean;
  flag_reason?: string;
}

export interface AdminRegistryFunctionVersion {
  id: string;
  function_id: string;
  version: string;
  runtime: string;
  timeout_ms: number;
  memory_mb: number;
  deterministic: boolean;
  cache_ttl: number;
  published_at: string;
  is_active: boolean;
}

export interface AdminRegistryStats {
  total_functions: number;
  public_functions: number;
  private_functions: number;
  unlisted_functions: number;
  flagged_functions: number;
  total_calls: number;
  total_revenue: number;
  avg_rating: number;
}

// Admin Functions API
export const adminFunctionsApi = {
  // List all functions across all tenants
  listFunctions: async (params?: {
    limit?: number;
    offset?: number;
    tenant_id?: string;
    status?: string;
    search?: string;
  }): Promise<{ functions: AdminFunction[]; total: number }> => {
    const searchParams = new URLSearchParams();
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.offset) searchParams.set('offset', params.offset.toString());
    if (params?.tenant_id) searchParams.set('tenant_id', params.tenant_id);
    if (params?.status) searchParams.set('status', params.status);
    if (params?.search) searchParams.set('search', params.search);

    const query = searchParams.toString();
    const url = query ? `/v1/admin/functions?${query}` : '/v1/admin/functions';
    return await apiClient.get<{ functions: AdminFunction[]; total: number }>(url);
  },

  // Get a specific function
  getFunction: async (functionId: string): Promise<AdminFunction> => {
    return await apiClient.get<AdminFunction>(`/v1/admin/functions/${functionId}`);
  },

  // Update a function (admin can modify any function)
  updateFunction: async (functionId: string, updates: Partial<AdminFunction>): Promise<AdminFunction> => {
    return await apiClient.patch<AdminFunction>(`/v1/admin/functions/${functionId}`, updates);
  },

  // Delete a function
  deleteFunction: async (functionId: string): Promise<{ ok: boolean; message: string }> => {
    return await apiClient.delete<{ ok: boolean; message: string }>(`/v1/admin/functions/${functionId}`);
  },

  // Enable/Disable a function
  toggleFunctionStatus: async (functionId: string, enabled: boolean): Promise<AdminFunction> => {
    return await apiClient.post<AdminFunction>(`/v1/admin/functions/${functionId}/toggle`, { enabled });
  },

  // Get function deployments
  getFunctionDeployments: async (functionId: string): Promise<{ deployments: AdminFunctionDeployment[] }> => {
    return await apiClient.get<{ deployments: AdminFunctionDeployment[] }>(`/v1/admin/functions/${functionId}/deployments`);
  },

  // Get function logs
  getFunctionLogs: async (functionId: string, params?: {
    limit?: number;
    level?: string;
    since?: string;
  }): Promise<{ logs: AdminFunctionLog[] }> => {
    const searchParams = new URLSearchParams();
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.level) searchParams.set('level', params.level);
    if (params?.since) searchParams.set('since', params.since);

    const query = searchParams.toString();
    const url = query ? `/v1/admin/functions/${functionId}/logs?${query}` : `/v1/admin/functions/${functionId}/logs`;
    return await apiClient.get<{ logs: AdminFunctionLog[] }>(url);
  },

  // Get function metrics
  getFunctionMetrics: async (functionId: string, params?: {
    period?: string;
    from?: string;
    to?: string;
  }): Promise<any> => {
    const searchParams = new URLSearchParams();
    if (params?.period) searchParams.set('period', params.period);
    if (params?.from) searchParams.set('from', params.from);
    if (params?.to) searchParams.set('to', params.to);

    const query = searchParams.toString();
    const url = query ? `/v1/admin/functions/${functionId}/metrics?${query}` : `/v1/admin/functions/${functionId}/metrics`;
    return await apiClient.get(url);
  },
};

// Admin Registry API
export const adminRegistryApi = {
  // Get registry statistics
  getStats: async (): Promise<AdminRegistryStats> => {
    return await apiClient.get<AdminRegistryStats>('/v1/admin/registry/stats');
  },

  // List all registry functions
  listFunctions: async (params?: {
    limit?: number;
    offset?: number;
    author?: string;
    category?: string;
    visibility?: string;
    flagged?: boolean;
    search?: string;
  }): Promise<{ functions: AdminRegistryFunction[]; total: number }> => {
    const searchParams = new URLSearchParams();
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.offset) searchParams.set('offset', params.offset.toString());
    if (params?.author) searchParams.set('author', params.author);
    if (params?.category) searchParams.set('category', params.category);
    if (params?.visibility) searchParams.set('visibility', params.visibility);
    if (params?.flagged !== undefined) searchParams.set('flagged', params.flagged.toString());
    if (params?.search) searchParams.set('search', params.search);

    const query = searchParams.toString();
    const url = query ? `/v1/admin/registry/functions?${query}` : '/v1/admin/registry/functions';
    return await apiClient.get<{ functions: AdminRegistryFunction[]; total: number }>(url);
  },

  // Get a specific registry function
  getFunction: async (functionId: string): Promise<{
    function: AdminRegistryFunction;
    versions: AdminRegistryFunctionVersion[];
  }> => {
    return await apiClient.get<{ function: AdminRegistryFunction; versions: AdminRegistryFunctionVersion[] }>(
      `/v1/admin/registry/functions/${functionId}`
    );
  },

  // Update registry function
  updateFunction: async (functionId: string, updates: Partial<AdminRegistryFunction>): Promise<AdminRegistryFunction> => {
    return await apiClient.patch<AdminRegistryFunction>(`/v1/admin/registry/functions/${functionId}`, updates);
  },

  // Delete registry function
  deleteFunction: async (functionId: string): Promise<{ ok: boolean; message: string }> => {
    return await apiClient.delete<{ ok: boolean; message: string }>(`/v1/admin/registry/functions/${functionId}`);
  },

  // Update function visibility
  updateVisibility: async (functionId: string, visibility: 'public' | 'private' | 'unlisted'): Promise<AdminRegistryFunction> => {
    return await apiClient.patch<AdminRegistryFunction>(`/v1/admin/registry/functions/${functionId}/visibility`, { visibility });
  },

  // Update function pricing
  updatePricing: async (functionId: string, pricePerCall: number): Promise<AdminRegistryFunction> => {
    return await apiClient.patch<AdminRegistryFunction>(`/v1/admin/registry/functions/${functionId}/pricing`, { price_per_call: pricePerCall });
  },

  // Flag/Unflag a function
  toggleFlag: async (functionId: string, flagged: boolean, reason?: string): Promise<AdminRegistryFunction> => {
    return await apiClient.post<AdminRegistryFunction>(`/v1/admin/registry/functions/${functionId}/flag`, { flagged, reason });
  },

  // Get function versions
  getVersions: async (functionId: string): Promise<{ versions: AdminRegistryFunctionVersion[] }> => {
    return await apiClient.get<{ versions: AdminRegistryFunctionVersion[] }>(`/v1/admin/registry/functions/${functionId}/versions`);
  },

  // Deactivate a specific version
  deactivateVersion: async (functionId: string, versionId: string): Promise<AdminRegistryFunctionVersion> => {
    return await apiClient.post<AdminRegistryFunctionVersion>(`/v1/admin/registry/functions/${functionId}/versions/${versionId}/deactivate`);
  },

  // Get registry function metrics
  getMetrics: async (functionId: string, params?: {
    period?: string;
    from?: string;
    to?: string;
  }): Promise<any> => {
    const searchParams = new URLSearchParams();
    if (params?.period) searchParams.set('period', params.period);
    if (params?.from) searchParams.set('from', params.from);
    if (params?.to) searchParams.set('to', params.to);

    const query = searchParams.toString();
    const url = query ? `/v1/admin/registry/functions/${functionId}/metrics?${query}` : `/v1/admin/registry/functions/${functionId}/metrics`;
    return await apiClient.get(url);
  },
};
