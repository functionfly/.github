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
  'developer',
  'notifications',
  'security',
  'privacy',
  'platform',
  'integrations',
  'github',
] as const;
export type SettingsTabValue = (typeof VALID_TABS)[number];

/**
 * Generate a shareable settings URL with hash-based tab routing.
 * Long-term URL structure for maximum compatibility and shareability.
 *
 * @example
 *   getSettingsUrl('traseputallaz', 'billing') // "/u/traseputallaz/settings#billing"
 *   getSettingsUrl('traseputallaz')            // "/u/traseputallaz/settings#account"
 */
export function getSettingsUrl(
  username: string,
  tab?: SettingsTabValue
): string {
  const base = `/u/${username}/settings`;
  return tab && tab !== 'account' ? `${base}#${tab}` : base;
}

/**
 * Generate legacy path-based settings URL for backwards compatibility.
 * Use only when you need the old-style /u/:username/settings/billing paths.
 *
 * @deprecated Prefer getSettingsUrl() for new code
 */
export function getSettingsUrlLegacy(
  username: string,
  tab?: SettingsTabValue
): string {
  if (tab && tab !== 'account') {
    return `/u/${username}/settings/${tab}`;
  }
  return `/u/${username}/settings`;
}

/**
 * Parse tab from full URL (handles both hash and legacy path formats).
 * Priority: hash > path > query param > default
 */
export function parseTabFromUrl(url: string): SettingsTabValue {
  try {
    const parsed = new URL(url, window.location.origin);
    const hashTab = parsed.hash.replace('#', '');
    if (hashTab && VALID_TABS.includes(hashTab as SettingsTabValue)) {
      return hashTab as SettingsTabValue;
    }
    const pathMatch = parsed.pathname.match(/\/settings\/(\w+)$/);
    if (pathMatch && VALID_TABS.includes(pathMatch[1] as SettingsTabValue)) {
      return pathMatch[1] as SettingsTabValue;
    }
    const queryTab = parsed.searchParams.get('subtab');
    if (queryTab && VALID_TABS.includes(queryTab as SettingsTabValue)) {
      return queryTab as SettingsTabValue;
    }
  } catch {
    // Invalid URL, return default
  }
  return 'account';
}
