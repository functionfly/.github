/**
 * FlyRouteListener.tsx
 *
 * Detects route changes using React Router and updates assistant context automatically.
 * Supports deep linking and debounced updates for optimal performance.
 *
 * @module fly-assistant
 */

import { useEffect, useRef, useCallback, useMemo } from "react";
import { useLocation, useSearchParams } from "react-router-dom";
import {
  useFlyAssistant,
  RouteContext,
  UserRole,
} from "./FlyAssistantProvider";

// ============================================================================
// Types & Interfaces
// ============================================================================

/**
 * Route pattern configuration for parsing URLs
 */
export interface RoutePattern {
  /** Regular expression to match the route */
  pattern: RegExp;
  /** Function to extract context from regex match */
  extract: (match: RegExpMatchArray) => Partial<ParsedRouteContext>;
  /** Human-readable route name */
  name: string;
}

/**
 * Extended route context with parsed parameters
 */
export interface ParsedRouteContext extends RouteContext {
  /** Current function ID if on function page */
  currentFunction?: string;
  /** Current marketplace category if on marketplace */
  category?: string;
  /** User role context */
  userRole?: UserRole;
  /** Query parameters from URL */
  queryParams?: Record<string, string>;
}

/**
 * Props for FlyRouteListener component
 */
export interface FlyRouteListenerProps {
  /** Callback fired when route changes */
  onRouteChange?: (route: ParsedRouteContext) => void;
  /** Enable deep linking support for assistant state */
  enableDeepLink?: boolean;
  /** Debounce delay for route updates in ms */
  debounceMs?: number;
  /** Custom route patterns to parse */
  customPatterns?: RoutePattern[];
}

/**
 * Event detail for route change events
 */
export interface RouteChangeEventDetail {
  route: ParsedRouteContext;
  previousRoute: ParsedRouteContext | null;
  timestamp: number;
}

// ============================================================================
// Default Route Patterns
// ============================================================================

const defaultRoutePatterns: RoutePattern[] = [
  // Function detail page
  {
    pattern: /\/functions\/([^\/]+)/,
    name: "Function Detail",
    extract: (match) => ({
      currentFunction: match[1],
    }),
  },
  // Function edit page
  {
    pattern: /\/functions\/([^\/]+)\/edit/,
    name: "Function Editor",
    extract: (match) => ({
      currentFunction: match[1],
    }),
  },
  // Marketplace with category
  {
    pattern: /\/marketplace(?:\/([^\/]+))?/,
    name: "Marketplace",
    extract: (match) => ({
      category: match[1] || "all",
    }),
  },
  // Dashboard home
  {
    pattern: /^\/$/,
    name: "Dashboard",
    extract: () => ({}),
  },
  // Functions list
  {
    pattern: /\/functions\/?$/,
    name: "Functions List",
    extract: () => ({}),
  },
  // Deployments
  {
    pattern: /\/deployments/,
    name: "Deployments",
    extract: () => ({}),
  },
  // Analytics
  {
    pattern: /\/analytics/,
    name: "Analytics",
    extract: () => ({}),
  },
  // Settings
  {
    pattern: /\/settings/,
    name: "Settings",
    extract: () => ({}),
  },
  // Organization
  {
    pattern: /\/org/,
    name: "Organization",
    extract: () => ({}),
  },
];

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Parse URL to extract route context
 */
function parseRouteContext(
  pathname: string,
  searchParams: URLSearchParams,
  patterns: RoutePattern[]
): ParsedRouteContext {
  // Find matching pattern
  for (const routePattern of patterns) {
    const match = pathname.match(routePattern.pattern);
    if (match) {
      const extracted = routePattern.extract(match);
      const queryParams: Record<string, string> = {};

      searchParams.forEach((value, key) => {
        queryParams[key] = value;
      });

      return {
        path: pathname,
        name: routePattern.name,
        params: Object.fromEntries(searchParams.entries()),
        ...extracted,
        queryParams,
      };
    }
  }

  // Default fallback
  const queryParams: Record<string, string> = {};
  searchParams.forEach((value, key) => {
    queryParams[key] = value;
  });

  return {
    path: pathname,
    name: "Unknown",
    params: Object.fromEntries(searchParams.entries()),
    queryParams,
  };
}

/**
 * Deep link parameter keys
 */
const DEEP_LINK_PARAMS = {
  open: "fly_open",
  mode: "fly_mode",
  context: "fly_context",
};

// ============================================================================
// Component
// ============================================================================

/**
 * FlyRouteListener - Route detection and context update component
 *
 * Automatically tracks route changes and updates the FlyAssistant store
 * with current page context. Supports deep linking and custom route patterns.
 *
 * @example
 * ```tsx
 * <FlyRouteListener
 *   onRouteChange={(route) => console.log("Route:", route.name)}
 *   enableDeepLink={true}
 *   debounceMs={100}
 * />
 * ```
 */
export function FlyRouteListener({
  onRouteChange,
  enableDeepLink = false,
  debounceMs = 100,
  customPatterns = [],
}: FlyRouteListenerProps) {
  const location = useLocation();
  const [searchParams, setSearchParams] = useSearchParams();
  const setCurrentRoute = useFlyAssistant((state) => state.setCurrentRoute);
  const open = useFlyAssistant((state) => state.open);
  const setFullscreen = useFlyAssistant((state) => state.setFullscreen);

  const timeoutRef = useRef<NodeJS.Timeout | null>(null);
  const previousRouteRef = useRef<ParsedRouteContext | null>(null);

  // Merge default and custom patterns
  const patterns = useMemo(
    () => [...defaultRoutePatterns, ...customPatterns],
    [customPatterns]
  );

  /**
   * Handle deep linking from URL parameters
   */
  const handleDeepLink = useCallback(
    (routeContext: ParsedRouteContext) => {
      if (!enableDeepLink) return;

      const shouldOpen = searchParams.get(DEEP_LINK_PARAMS.open) === "true";
      const mode = searchParams.get(DEEP_LINK_PARAMS.mode);
      const context = searchParams.get(DEEP_LINK_PARAMS.context);

      if (shouldOpen) {
        open();

        if (mode === "fullscreen") {
          setFullscreen(true);
        }

        // Update route context with deep link context if provided
        if (context) {
          try {
            const parsedContext = JSON.parse(decodeURIComponent(context));
            setCurrentRoute({
              ...routeContext,
              ...parsedContext,
            });
          } catch {
            // Invalid context JSON, ignore
          }
        }

        // Clean up deep link params from URL
        const newParams = new URLSearchParams(searchParams);
        newParams.delete(DEEP_LINK_PARAMS.open);
        newParams.delete(DEEP_LINK_PARAMS.mode);
        newParams.delete(DEEP_LINK_PARAMS.context);
        setSearchParams(newParams, { replace: true });
      }
    },
    [enableDeepLink, searchParams, open, setFullscreen, setCurrentRoute, setSearchParams]
  );

  /**
   * Process route change with debouncing
   */
  const processRouteChange = useCallback(
    (pathname: string, searchParams: URLSearchParams) => {
      // Clear existing timeout
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }

      timeoutRef.current = setTimeout(() => {
        const routeContext = parseRouteContext(pathname, searchParams, patterns);

        // Update store
        setCurrentRoute(routeContext);

        // Handle deep linking
        handleDeepLink(routeContext);

        // Dispatch custom event for analytics
        const eventDetail: RouteChangeEventDetail = {
          route: routeContext,
          previousRoute: previousRouteRef.current,
          timestamp: Date.now(),
        };

        window.dispatchEvent(
          new CustomEvent("fly:routeChange", {
            detail: eventDetail,
          })
        );

        // Call callback if provided
        if (onRouteChange) {
          onRouteChange(routeContext);
        }

        // Update previous route ref
        previousRouteRef.current = routeContext;
      }, debounceMs);
    },
    [patterns, setCurrentRoute, handleDeepLink, onRouteChange, debounceMs]
  );

  // Listen for route changes
  useEffect(() => {
    processRouteChange(location.pathname, searchParams);

    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, [location.pathname, searchParams, processRouteChange]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  // This is a logic-only component, no visual output
  return null;
}

/**
 * Generate deep link URL with assistant state
 */
export function generateDeepLink(
  basePath: string,
  options: {
    open?: boolean;
    mode?: "default" | "fullscreen";
    context?: Record<string, unknown>;
  }
): string {
  const params = new URLSearchParams();

  if (options.open) {
    params.set(DEEP_LINK_PARAMS.open, "true");
  }

  if (options.mode && options.mode !== "default") {
    params.set(DEEP_LINK_PARAMS.mode, options.mode);
  }

  if (options.context) {
    params.set(
      DEEP_LINK_PARAMS.context,
      encodeURIComponent(JSON.stringify(options.context))
    );
  }

  const queryString = params.toString();
  return queryString ? `${basePath}?${queryString}` : basePath;
}

export default FlyRouteListener;
