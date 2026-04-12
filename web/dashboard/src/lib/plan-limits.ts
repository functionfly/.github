import { PLANS } from './constants';

/**
 * Get functions limit for a plan
 */
export function getFunctionsLimit(plan?: string): number {
  const planKey = plan?.toUpperCase() as keyof typeof PLANS;
  const limits = PLANS[planKey]?.limits;
  if (!limits) return 0;
  const raw = limits.functions;
  return raw === Infinity ? 10000 : raw ?? 0;
}

/**
 * Get providers limit for a plan
 */
export function getProvidersLimit(plan?: string): number {
  const planKey = plan?.toUpperCase() as keyof typeof PLANS;
  const limits = PLANS[planKey]?.limits;
  if (!limits) return 0;
  const raw = limits.providers;
  return raw === Infinity ? 10000 : raw ?? 0;
}

/**
 * Get requests limit for a plan
 */
export function getRequestsLimit(plan?: string): number {
  const planKey = plan?.toUpperCase() as keyof typeof PLANS;
  const limits = PLANS[planKey]?.limits;
  if (!limits) return 0;
  const raw = limits.requests;
  return raw === Infinity ? 10000000 : raw ?? 0;
}

/**
 * Format a limit for display
 * Shows "Unlimited" for large numbers, otherwise the actual number
 */
export function formatLimit(limit: number): string {
  if (limit === Infinity || limit >= 10000) {
    return 'Unlimited';
  }
  if (limit >= 1000000) {
    return `${(limit / 1000000).toFixed(0)}M`;
  }
  if (limit >= 1000) {
    return `${(limit / 1000).toFixed(0)}k`;
  }
  return limit.toString();
}

/**
 * Check if user can create more functions
 */
export function canCreateFunction(plan: string | undefined, currentCount: number): boolean {
  const limit = getFunctionsLimit(plan);
  if (limit >= 10000) return true;
  return currentCount < limit;
}

/**
 * Check if user can add more providers
 */
export function canAddProvider(plan: string | undefined, currentCount: number): boolean {
  const limit = getProvidersLimit(plan);
  if (limit >= 10000) return true;
  return currentCount < limit;
}

/**
 * Format remaining count for display
 */
export function formatRemaining(current: number, limit: number): string {
  if (limit === Infinity || limit >= 10000) {
    return `${current} (unlimited)`;
  }
  const remaining = Math.max(0, limit - current);
  return `${remaining} remaining`;
}
