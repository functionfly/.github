/**
 * Sidebar utilities — pure functions and hooks that carry no JSX.
 * Kept separate so they can be imported without pulling in the component tree.
 */
import { useCallback, useRef } from 'react';
import { NAV_LABEL_KEYS } from './navigation';

// ============================================================================
// Pure helpers (no React)
// ============================================================================

/** HTML-escape a string — defense-in-depth for future refactors. */
export function escapeHtml(str: string): string {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

/**
 * Rate-limited keyboard shortcut handler.
 * Prevents rapid-fire shortcut execution that could bypass navigation guards.
 * Minimum interval: 100ms between same shortcut executions.
 *
 * Uses a ref for lastExecution so the rate limit persists across
 * handler recreations caused by dependency changes.
 */
export function createRateLimitedHandler(
  handler: (e: KeyboardEvent) => void,
  minIntervalMs = 100
): (e: KeyboardEvent) => void {
const lastExecutionRef = { current: 0 };
  return (e: KeyboardEvent) => {
    const now = Date.now();
    if (now - lastExecutionRef.current >= minIntervalMs) {
      lastExecutionRef.current = now;
      handler(e);
    }
  };
}

/**
 * Check if a nav path matches the current location.
 * - Root paths (/, /dashboard, /overview) require exact match.
 * - All other paths match exactly OR when the current pathname starts with
 *   the path followed by a slash (child routes), preventing /settings
 *   from matching /settings-something.
 */
export function isItemActive(path: string, pathname: string): boolean {
  const EXACT_MATCH_ROOTS = ['/', '/dashboard', '/overview'];
  if (EXACT_MATCH_ROOTS.includes(path)) return pathname === path;
  return pathname === path || (pathname.startsWith(path + '/') && path !== '/');
}

// ============================================================================
// i18n label translation
// ============================================================================

/** Translate a nav label using i18n. i18n system is trusted to produce plain text. */
export function translateLabel(t: (key: string) => string, label: string): string {
  const translated = NAV_LABEL_KEYS[label] ? t(NAV_LABEL_KEYS[label]) : label;
  return translated;
}

// ============================================================================
// Debounce hook
// ============================================================================

/**
 * Debounce a callback (for search input).
 * Prevents excessive re-renders on every keystroke.
 */
export function useDebounce<T extends (...args: Parameters<T>) => void>(
  fn: T,
  delayMs: number
): (...args: Parameters<T>) => void {
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  return useCallback(
    (...args: Parameters<T>) => {
      if (timerRef.current) clearTimeout(timerRef.current);
      timerRef.current = setTimeout(() => fn(...args), delayMs);
    },
    [fn, delayMs]
  );
}