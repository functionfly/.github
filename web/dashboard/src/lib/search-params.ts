/**
 * nuqs-based URL search params utilities.
 *
 * Replaces manual useSearchParams + URLSearchParams patterns with type-safe,
 * declarative search param state that syncs with the URL.
 *
 * Usage:
 * ```tsx
 * import { useTabParam, useRedirectParam } from '@/lib/search-params';
 *
 * function MyPage() {
 *   const [tab, setTab] = useTabParam(['overview', 'settings', 'billing']);
 *   const redirect = useRedirectParam();
 * }
 * ```
 */
import {
  parseAsBoolean,
  parseAsInteger,
  parseAsString,
  parseAsStringEnum,
  useQueryState,
  useQueryStates,
  type Options,
} from 'nuqs';

// Re-export nuqs hooks and parsers for direct use
export {
  parseAsArrayOf,
  parseAsBoolean,
  parseAsFloat,
  parseAsInteger,
  parseAsJson,
  parseAsString,
  parseAsStringEnum,
  parseAsStringLiteral,
  useQueryState,
  useQueryStates,
} from 'nuqs';

// ============================================================================
// Common Search Param Hooks
// ============================================================================

/**
 * Tab navigation state via URL search param.
 * Replaces: const [searchParams, setSearchParams] = useSearchParams(); searchParams.get('tab')
 *
 * @param validTabs - Array of valid tab values
 * @param defaultTab - Default tab (first value in array)
 * @param options - nuqs options (history, scroll, etc.)
 */
export function useTabParam<T extends string>(
  validTabs: readonly T[],
  defaultTab?: T,
  options?: Options
) {
  const parser = parseAsStringEnum<T>(validTabs as T[]);
  const defaulted = defaultTab !== undefined ? parser.withDefault(defaultTab) : parser;
  return useQueryState('tab', defaulted.withOptions(options ?? { history: 'replace' }));
}

/**
 * Post-auth redirect path from URL search param.
 * Read-only — set by auth pages, consumed after login.
 */
export function useRedirectParam() {
  const [redirect] = useQueryState('redirect', parseAsString.withDefault('/'));
  return redirect;
}

/**
 * Boolean admin flag (string '1' → true).
 * Used on login/MFA pages to redirect to admin dashboard.
 */
export function useAdminParam() {
  const [admin, setAdmin] = useQueryState(
    'admin',
    parseAsBoolean.withDefault(false)
  );
  return [admin, setAdmin] as const;
}

/**
 * Stripe callback status (success/cancel) that auto-clears after consumption.
 * Replaces the repeated pattern of reading walletTopUp/subscription/credits
 * then deleting from URL.
 *
 * @param key - The search param key (e.g., 'walletTopUp', 'subscription', 'credits')
 */
export function useStripeCallbackParam(key: string) {
  const [status, setStatus] = useQueryState(
    key,
    parseAsString.withOptions({ history: 'replace' })
  );

  const clearStatus = () => setStatus(null);

  return [status, clearStatus] as const;
}

/**
 * Page number for paginated views.
 */
export function usePageParam(defaultPage = 1) {
  return useQueryState(
    'page',
    parseAsInteger.withDefault(defaultPage).withOptions({ history: 'replace' })
  );
}

/**
 * Sort direction for sortable views.
 */
export function useSortParam(defaultSort: 'asc' | 'desc' = 'desc') {
  return useQueryState(
    'sort',
    parseAsStringEnum<'asc' | 'desc'>(['asc', 'desc']).withDefault(defaultSort)
  );
}

/**
 * Filter/search query string param.
 */
export function useSearchParam() {
  return useQueryState('q', parseAsString.withDefault(''));
}

// ============================================================================
// Multi-Param Hooks (for pages with several search params)
// ============================================================================

/**
 * Common execution explorer search params.
 */
export function useExecutionExplorerParams() {
  return useQueryStates({
    tab: parseAsStringEnum(['executions', 'certificates', 'history']).withDefault('executions'),
    certId: parseAsString,
  });
}

/**
 * Common auth flow search params.
 */
export function useAuthFlowParams() {
  return useQueryStates({
    redirect: parseAsString.withDefault('/'),
    admin: parseAsBoolean.withDefault(false),
    email: parseAsString,
  });
}

/**
 * Fly assistant search params.
 */
export function useFlyAssistantParams() {
  return useQueryStates({
    fly_open: parseAsBoolean.withDefault(false),
    fly_mode: parseAsString,
    fly_context: parseAsString,
  });
}
