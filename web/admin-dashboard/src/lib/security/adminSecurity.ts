/**
 * Admin security service helpers.
 */

import { adminApiClient } from '@/lib/api/adminClient';

export interface IPCheckResult {
  allowed: boolean;
  reason?: string;
  sourceIp?: string;
}

function isNotFoundError(err: unknown): boolean {
  const maybe = err as { response?: { status?: number } };
  return maybe?.response?.status === 404;
}

export async function checkAdminIPAccess(): Promise<IPCheckResult> {
  try {
    const resp = await adminApiClient.get<{ allowed?: boolean; reason?: string; source_ip?: string }>('/security/check-ip');
    const payload = resp?.data || {};

    return {
      allowed: payload.allowed !== false,
      reason: payload.reason,
      sourceIp: payload.source_ip,
    };
  } catch (err) {
    // Backward-compatible fallback while endpoint is rolling out.
    if (isNotFoundError(err)) {
      return {
        allowed: true,
        reason: 'ip_check_endpoint_not_available',
      };
    }

    return {
      allowed: false,
      reason: 'ip_check_failed',
    };
  }
}
