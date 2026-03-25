/**
 * useRouteTracking.ts
 *
 * Hook for accessing and monitoring route context changes.
 * Automatically updates when the route changes.
 *
 * @module fly-assistant/hooks
 */

import { useCallback, useEffect, useState } from "react";
import {
  useFlyAssistant,
  RouteContext,
} from "../FlyAssistantProvider";

// ============================================================================
// Types
// ============================================================================

/**
 * Extended route context with parsed information
 */
export interface ExtendedRouteContext extends RouteContext {
  /** Whether currently on a function page */
  isFunctionPage: boolean;
  /** Whether currently on the marketplace */
  isMarketplacePage: boolean;
  /** Whether currently on dashboard home */
  isDashboard: boolean;
  /** Function ID if on function page */
  functionId?: string;
  /** Category if on marketplace */
  category?: string;
}

/**
 * Return type for useRouteTracking hook
 */
export interface UseRouteTrackingReturn {
  /** Current route context */
  route: ExtendedRouteContext | null;
  /** Current page name */
  pageName: string;
  /** Current path */
  path: string;
  /** URL parameters */
  params: Record<string, string>;
  /** Whether route is ready */
  isReady: boolean;
  /** Refetch route context */
  refresh: () => void;
}

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Extend route context with computed properties
 */
function extendRouteContext(
  route: RouteContext | null
): ExtendedRouteContext | null {
  if (!route) return null;

  const path = route.path;
  const isFunctionPage = /\/functions\/[^\/]+/.test(path);
  const isMarketplacePage = path.startsWith("/marketplace") || path === "/dashboard";
  const isDashboard = path === "/" || path === "/dashboard" || path === "/overview";

  return {
    ...route,
    isFunctionPage,
    isMarketplacePage,
    isDashboard,
    functionId: isFunctionPage
      ? route.params?.id || route.params?.functionId
      : undefined,
    category: isMarketplacePage ? route.params?.category : undefined,
  };
}

// ============================================================================
// Hook
// ============================================================================

/**
 * useRouteTracking - Hook for route context access
 *
 * Returns current route context with computed properties.
 * Automatically re-renders when route changes.
 *
 * @example
 * ```tsx
 * const { route, pageName, isFunctionPage } = useRouteTracking();
 *
 * if (isFunctionPage) {
 *   console.log("Current function:", route.functionId);
 * }
 * ```
 */
export function useRouteTracking(): UseRouteTrackingReturn {
  const currentRoute = useFlyAssistant((state) => state.currentRoute);
  const setCurrentRoute = useFlyAssistant((state) => state.setCurrentRoute);

  const [extendedRoute, setExtendedRoute] = useState<ExtendedRouteContext | null>(
    () => extendRouteContext(currentRoute)
  );

  // Update extended route when current route changes
  useEffect(() => {
    setExtendedRoute(extendRouteContext(currentRoute));
  }, [currentRoute]);

  // Listen for route change events
  useEffect(() => {
    const handleRouteChange = (event: Event) => {
      const customEvent = event as CustomEvent;
      if (customEvent.detail?.route) {
        setExtendedRoute(extendRouteContext(customEvent.detail.route));
      }
    };

    window.addEventListener("fly:routeChange", handleRouteChange);
    return () => window.removeEventListener("fly:routeChange", handleRouteChange);
  }, []);

  /**
   * Refresh route context
   */
  const refresh = useCallback(() => {
    // Re-trigger route parsing by dispatching a custom event
    window.dispatchEvent(
      new CustomEvent("fly:routeRefresh", {
        detail: { timestamp: Date.now() },
      })
    );
  }, []);

  return {
    route: extendedRoute,
    pageName: extendedRoute?.name || "Unknown",
    path: extendedRoute?.path || "",
    params: extendedRoute?.params || {},
    isReady: extendedRoute !== null,
    refresh,
  };
}

export default useRouteTracking;
