/**
 * Shared helpers for Settings page and tab components.
 */

export const formatDate = (dateStr: string | null): string => {
  if (!dateStr) return '-';
  return new Date(dateStr).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
};

export const formatCurrency = (amount: number, currency: string = 'usd'): string => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: currency.toUpperCase(),
  }).format(amount / 100);
};

/** Extract user-facing message from API error (axios or similar). */
export function getApiErrorMessage(
  err: unknown,
  fallbacks: { 401?: string; 403?: string; 404?: string; default?: string }
): string {
  const res =
    err && typeof err === 'object' && 'response' in err
      ? (err as { response?: { data?: unknown; status?: number } }).response
      : null;
  const data = res?.data;
  const status = res?.status;
  if (data && typeof data === 'object' && data !== null && 'error' in data) {
    const errorObj = (data as { error?: { message?: string } }).error;
    if (errorObj && typeof errorObj.message === 'string' && errorObj.message)
      return errorObj.message;
  }
  if (status === 401 && fallbacks[401]) return fallbacks[401];
  if (status === 403 && fallbacks[403]) return fallbacks[403];
  if (status === 404 && fallbacks[404]) return fallbacks[404];
  return fallbacks.default ?? 'Something went wrong.';
}

export const VALID_TABS = [
  'account',
  'billing',
  'api',
  'notifications',
  'security',
  'privacy',
] as const;
export type SettingsTabValue = (typeof VALID_TABS)[number];
