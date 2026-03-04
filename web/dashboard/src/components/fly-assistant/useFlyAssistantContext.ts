/**
 * useFlyAssistantContext.ts
 *
 * A composite hook that combines all FlyAssistant hooks into a single,
 * convenient interface. This hook provides everything needed for most
 * FlyAssistant use cases in one place.
 *
 * @module fly-assistant
 * @example
 * ```tsx
 * function MyComponent() {
 *   const {
 *     // UI State
 *     isOpen,
 *     isMinimized,
 *     isFullscreen,
 *
 *     // User
 *     userSession,
 *     currentRoute,
 *
 *     // Status
 *     trustScore,
 *     trustTier,
 *     hasError,
 *     notificationCount,
 *
 *     // Actions
 *     open,
 *     close,
 *     toggle,
 *     setFullscreen,
 *
 *     // Route
 *     route,
 *     pageName,
 *
 *     // Events
 *     track,
 *     trackError,
 *   } = useFlyAssistantContext();
 *
 *   return (
 *     <button onClick={toggle}>
 *       {isOpen ? 'Close' : 'Open'} Assistant
 *     </button>
 *   );
 * }
 * ```
 */

import {
  useFlyAssistant,
  useFlyAssistantActions,
  useFlyAssistantUser,
  useFlyAssistantStatus,
  useFlyAssistantCache,
} from "./FlyAssistantProvider";
import { useRouteTracking } from "./hooks/useRouteTracking";
import { useEventTracking } from "./hooks/useEventTracking";
import type {
  DeploymentEventData,
  TrustChangeEventData,
  MarketplaceEventData,
  AssistantUsageEventData,
  FlyAssistantState,
  TrustTier,
  RouteContext,
  UserSession,
  ExtendedRouteContext,
} from "./types";

// ============================================================================
// Return Type
// ============================================================================

/**
 * Return type for useFlyAssistantContext hook
 * Combines all FlyAssistant state and actions into a single interface
 */
export interface UseFlyAssistantContextReturn {
  // ==========================================================================
  // UI State (from FlyAssistantState)
  // ==========================================================================

  /** Whether the assistant panel is currently open */
  isOpen: boolean;
  /** Whether the assistant is currently minimized */
  isMinimized: boolean;
  /** Whether the assistant is in fullscreen mode */
  isFullscreen: boolean;

  // ==========================================================================
  // User State (from useFlyAssistantUser)
  // ==========================================================================

  /** Current user session information */
  userSession: UserSession | null;
  /** Current route context */
  currentRoute: RouteContext | null;

  // ==========================================================================
  // Status State (from useFlyAssistantStatus)
  // ==========================================================================

  /** Current trust score (0-1) */
  trustScore: number;
  /** Current trust tier based on score */
  trustTier: TrustTier;
  /** Whether an error has occurred */
  hasError: boolean;
  /** Error message if hasError is true */
  errorMessage: string | null;
  /** Whether there are insights to display */
  hasInsights: boolean;
  /** Number of pending notifications */
  notificationCount: number;

  // ==========================================================================
  // Cache State (from useFlyAssistantCache)
  // ==========================================================================

  /** Map of cached conversations */
  conversationCache: Map<string, FlyAssistantState["conversationCache"] extends Map<string, infer V> ? V : never>;
  /** ID of the current conversation */
  currentConversationId: string | null;

  // ==========================================================================
  // UI Actions (from useFlyAssistantActions)
  // ==========================================================================

  /** Open the assistant panel */
  open: () => void;
  /** Close the assistant panel */
  close: () => void;
  /** Toggle the assistant panel open/closed */
  toggle: () => void;
  /** Minimize the assistant panel */
  minimize: () => void;
  /** Expand the assistant from minimized state */
  expand: () => void;
  /** Set fullscreen mode */
  setFullscreen: (value: boolean) => void;

  // ==========================================================================
  // User Actions (from useFlyAssistantUser)
  // ==========================================================================

  /** Update the user session */
  setUserSession: (session: UserSession | null) => void;
  /** Update the current route */
  setCurrentRoute: (route: RouteContext | null) => void;

  // ==========================================================================
  // Status Actions (from useFlyAssistantStatus)
  // ==========================================================================

  /** Update the trust score */
  setTrustScore: (score: number) => void;
  /** Set an error message */
  setError: (error: string | null) => void;
  /** Set whether insights are available */
  setHasInsights: (value: boolean) => void;
  /** Set the notification count */
  setNotificationCount: (count: number) => void;

  // ==========================================================================
  // Cache Actions (from useFlyAssistantCache)
  // ==========================================================================

  /** Add a conversation entry to the cache */
  addToCache: FlyAssistantState["addToCache"];
  /** Clear all cached conversations */
  clearCache: () => void;
  /** Set the current conversation ID */
  setCurrentConversation: (id: string | null) => void;

  // ==========================================================================
  // Route Tracking (from useRouteTracking)
  // ==========================================================================

  /** Extended route context with computed properties */
  route: ExtendedRouteContext | null;
  /** Current page name */
  pageName: string;
  /** Current path */
  path: string;
  /** URL parameters */
  params: Record<string, string>;
  /** Whether route tracking is ready */
  isReady: boolean;
  /** Refresh the route context */
  refresh: () => void;

  // ==========================================================================
  // Event Tracking (from useEventTracking)
  // ==========================================================================

  /** Track a generic event */
  track: (eventName: string, data?: Record<string, unknown>) => void;
  /** Track an error */
  trackError: (error: Error, context?: Record<string, unknown>) => void;
  /** Track a deployment event */
  trackDeployment: (data: DeploymentEventData) => void;
  /** Track a trust score change */
  trackTrustChange: (data: TrustChangeEventData) => void;
  /** Track marketplace view */
  trackMarketplaceView: (data: MarketplaceEventData) => void;
  /** Track marketplace install */
  trackMarketplaceInstall: (data: MarketplaceEventData) => void;
  /** Track marketplace rating */
  trackMarketplaceRate: (data: MarketplaceEventData) => void;
  /** Track assistant message */
  trackAssistantMessage: (data?: AssistantUsageEventData) => void;
  /** Track assistant action */
  trackAssistantAction: (actionType: string, data?: Record<string, unknown>) => void;
  /** Track assistant open */
  trackAssistantOpen: (mode?: string) => void;
  /** Track assistant close */
  trackAssistantClose: () => void;
  /** Track page view */
  trackPageView: (pageName: string, data?: Record<string, unknown>) => void;
  /** Manually flush pending events */
  flush: () => void;
  /** Whether tracking is enabled */
  isTrackingEnabled: boolean;
  /** Number of pending events */
  pendingEventCount: number;
}

// ============================================================================
// Hook
// ============================================================================

/**
 * useFlyAssistantContext - Composite hook for FlyAssistant
 *
 * Combines all individual FlyAssistant hooks into a single interface.
 * This is the recommended hook for most use cases as it provides
 * complete access to the FlyAssistant state and actions.
 *
 * @example
 * ```tsx
 * // Basic usage
 * function AssistantToggle() {
 *   const { isOpen, toggle, notificationCount } = useFlyAssistantContext();
 *
 *   return (
 *     <button onClick={toggle}>
 *       Assistant {isOpen ? 'Open' : 'Closed'}
 *       {notificationCount > 0 && <span>({notificationCount})</span>}
 *     </button>
 *   );
 * }
 *
 * // With route tracking
 * function ContextAwareComponent() {
 *   const { route, trackPageView } = useFlyAssistantContext();
 *
 *   useEffect(() => {
 *     if (route?.isFunctionPage) {
 *       trackPageView('function_detail', { functionId: route.functionId });
 *     }
 *   }, [route]);
 *
 *   return <div>Current page: {route?.pageName}</div>;
 * }
 *
 * // With error tracking
 * function ErrorProneComponent() {
 *   const { trackError, setError } = useFlyAssistantContext();
 *
 *   const handleOperation = async () => {
 *     try {
 *       await riskyOperation();
 *     } catch (error) {
 *       trackError(error as Error, { component: 'ErrorProneComponent' });
 *       setError('Operation failed');
 *     }
 *   };
 *
 *   return <button onClick={handleOperation}>Do Risky Thing</button>;
 * }
 * ```
 *
 * @returns Combined FlyAssistant context with all state and actions
 * @throws Error if used outside of FlyAssistantProvider
 */
export function useFlyAssistantContext(): UseFlyAssistantContextReturn {
  // Get full state for UI properties
  const state = useFlyAssistant((s) => s);

  // Get individual hook results
  const actions = useFlyAssistantActions();
  const user = useFlyAssistantUser();
  const status = useFlyAssistantStatus();
  const cache = useFlyAssistantCache();
  const route = useRouteTracking();
  const events = useEventTracking();

  return {
    // UI State
    isOpen: state.isOpen,
    isMinimized: state.isMinimized,
    isFullscreen: state.isFullscreen,

    // User State
    userSession: user.userSession,
    currentRoute: user.currentRoute,

    // Status State
    trustScore: status.trustScore,
    trustTier: status.trustTier,
    hasError: status.hasError,
    errorMessage: status.errorMessage,
    hasInsights: status.hasInsights,
    notificationCount: status.notificationCount,

    // Cache State
    conversationCache: cache.conversationCache,
    currentConversationId: cache.currentConversationId,

    // UI Actions
    open: actions.open,
    close: actions.close,
    toggle: actions.toggle,
    minimize: actions.minimize,
    expand: actions.expand,
    setFullscreen: actions.setFullscreen,

    // User Actions
    setUserSession: user.setUserSession,
    setCurrentRoute: user.setCurrentRoute,

    // Status Actions
    setTrustScore: status.setTrustScore,
    setError: status.setError,
    setHasInsights: status.setHasInsights,
    setNotificationCount: status.setNotificationCount,

    // Cache Actions
    addToCache: cache.addToCache,
    clearCache: cache.clearCache,
    setCurrentConversation: cache.setCurrentConversation,

    // Route Tracking
    route: route.route,
    pageName: route.pageName,
    path: route.path,
    params: route.params,
    isReady: route.isReady,
    refresh: route.refresh,

    // Event Tracking
    track: events.track,
    trackError: events.trackError,
    trackDeployment: events.trackDeployment,
    trackTrustChange: events.trackTrustChange,
    trackMarketplaceView: events.trackMarketplaceView,
    trackMarketplaceInstall: events.trackMarketplaceInstall,
    trackMarketplaceRate: events.trackMarketplaceRate,
    trackAssistantMessage: events.trackAssistantMessage,
    trackAssistantAction: events.trackAssistantAction,
    trackAssistantOpen: events.trackAssistantOpen,
    trackAssistantClose: events.trackAssistantClose,
    trackPageView: events.trackPageView,
    flush: events.flush,
    isTrackingEnabled: events.isEnabled,
    pendingEventCount: events.pendingCount,
  };
}

export default useFlyAssistantContext;
