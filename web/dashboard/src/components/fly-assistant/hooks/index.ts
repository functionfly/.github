/**
 * FlyAssistant Hooks
 *
 * Custom hooks for route tracking, event tracking, and permission management.
 * These hooks provide reusable logic for integrating FlyAssistant features
 * throughout your application.
 *
 * @module fly-assistant/hooks
 * @example
 * ```tsx
 * import {
 *   useRouteTracking,
 *   useEventTracking,
 *   usePermission,
 * } from '@/components/fly-assistant/hooks';
 *
 * function MyComponent() {
 *   const { route, pageName } = useRouteTracking();
 *   const { track } = useEventTracking();
 *   const { hasPermission } = usePermission('pro', user.role);
 *
 *   return <div>Current page: {pageName}</div>;
 * }
 * ```
 */

// ============================================================================
// Route Tracking
// ============================================================================

/**
 * useRouteTracking - Hook for accessing and monitoring route context
 *
 * Automatically updates when the route changes and provides computed
 * properties like isFunctionPage, isMarketplacePage, etc.
 */
export {
  useRouteTracking,
} from "./useRouteTracking";

export type {
  ExtendedRouteContext,
  UseRouteTrackingReturn,
} from "./useRouteTracking";

// ============================================================================
// Event Tracking
// ============================================================================

/**
 * useEventTracking - Hook for tracking events, errors, and analytics
 *
 * Provides methods for tracking various events. Must be used within
 * a FlyEventTracker provider.
 */
export {
  useEventTracking,
} from "./useEventTracking";

export type {
  UseEventTrackingReturn,
} from "./useEventTracking";

// ============================================================================
// Permission Management
// ============================================================================

/**
 * usePermission - Hook for checking user permissions and tier-based feature access
 *
 * Provides utilities for checking user tier and permissions.
 * Integrates with FlyPermissionGuard.
 */
export {
  usePermission,
  useIsPro,
  useIsEnterprise,
  useTierComparison,
  checkTierAccess,
  getNextTier,
  getTierIndex,
} from "./usePermission";

export type {
  PermissionResult,
} from "./usePermission";
