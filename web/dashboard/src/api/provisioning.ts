import { apiClient } from './client';

// ─── Types ───────────────────────────────────────────────────────────────────

export interface ComponentState {
  status: 'pending' | 'provisioning' | 'active' | 'failed' | 'rolled_back';
  timestamp: string;
  error?: string;
  resource_id?: string;
}

export interface ProvisioningResult {
  tenant_id: string;
  bundle_slug: string;
  status: 'pending' | 'provisioning' | 'active' | 'failed';
  components: Record<string, ComponentState>;
  started_at: string;
  finished_at: string;
  duration_ms: number;
  error_log?: string[];
}

export interface NotProvisionedResponse {
  status: 'not_provisioned';
  tenant_id: string;
}

// ─── API Functions ───────────────────────────────────────────────────────────

/**
 * Trigger one-click provisioning for a bundle.
 * Creates an isolated database with Auth, Payments, Email, and Analytics.
 */
export async function provisionBundle(bundleSlug: string): Promise<ProvisioningResult> {
  return apiClient.post<ProvisioningResult>('/v1/provisioning/bundle', {
    bundle_slug: bundleSlug,
  });
}

/**
 * Get the current provisioning status for the authenticated tenant.
 */
export async function getProvisioningStatus(): Promise<ProvisioningResult | NotProvisionedResponse> {
  return apiClient.get<ProvisioningResult | NotProvisionedResponse>('/v1/provisioning/status');
}

/**
 * Retry failed provisioning components.
 * Idempotent — active components are skipped.
 */
export async function retryProvisioning(): Promise<{ status: string; message: string }> {
  return apiClient.post<{ status: string; message: string }>('/v1/provisioning/retry');
}
