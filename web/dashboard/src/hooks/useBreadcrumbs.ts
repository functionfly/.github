/**
 * useBreadcrumbs — hook for consuming and extending breadcrumbs.
 *
 * Usage:
 *   const { breadcrumbs } = useBreadcrumbs();
 *   const { breadcrumbs } = useBreadcrumbs([{ label: 'Extra', path: '/extra' }]);
 *   const { breadcrumbs } = useBreadcrumbs([], { prependHome: false });
 *
 * Pages can extend breadcrumbs by passing extra items — useful for detail pages
 * where the entity name is dynamic and can't be in the static registry.
 */

import { useMemo } from 'react';
import { useLocation } from 'react-router-dom';
import { generateBreadcrumbs, type BreadcrumbEntry } from '@/lib/breadcrumbs';

export interface UseBreadcrumbsOptions {
  /** Prepend the static Home crumb. Default: true. */
  prependHome?: boolean;
  /**
   * Map of path segments (e.g. 'functions') → display label.
   * Use when the route contains an entity ID that should display as a readable name.
   * Example: { 'functions': 'My Functions', 'abc-123': 'my-function-name' }
   */
  segmentLabels?: Record<string, string>;
  /**
   * Replace the auto-generated crumb for the current pathname.
   * Useful when you want to override with a dynamic entity name.
   */
  overrideCurrent?: string;
}

export interface Breadcrumb extends BreadcrumbEntry {
  isActive?: boolean;
}

/**
 * Returns breadcrumbs derived from the current pathname.
 * Optionally extended with additional items or with overrides for entity names.
 */
export function useBreadcrumbs(
  extra: BreadcrumbEntry[] = [],
  options: UseBreadcrumbsOptions = {}
): { breadcrumbs: Breadcrumb[] } {
  const { pathname } = useLocation();
  const { prependHome = true, segmentLabels = {}, overrideCurrent } = options;

  const breadcrumbs = useMemo(() => {
    const base = generateBreadcrumbs(pathname);

    // Apply segment label overrides
    const withOverrides = base.map((crumb, idx) => {
      // Last crumb (active page)
      if (idx === base.length - 1 && overrideCurrent) {
        return { ...crumb, label: overrideCurrent };
      }
      return crumb;
    });

    const filtered = prependHome
      ? withOverrides
      : withOverrides.slice(1); // drop Home

    const last = filtered.length - 1;
    return filtered.map((crumb, idx) => ({
      ...crumb,
      isActive: idx === last,
    }));
  }, [pathname, prependHome, overrideCurrent]);

  const extended = useMemo(
    () => (extra.length > 0 ? [...breadcrumbs.slice(0, -1), ...extra] : breadcrumbs),
    [breadcrumbs, extra]
  );

  return { breadcrumbs: extended };
}
