/**
 * usePermission.ts
 *
 * Hook for checking user permissions and tier-based feature access.
 * Integrates with FlyPermissionGuard for consistent permission handling.
 *
 * @module fly-assistant/hooks
 */

import { useCallback, useMemo } from "react";
import { UserRole } from "../FlyAssistantProvider";

// ============================================================================
// Types
// ============================================================================

/**
 * Permission check result
 */
export interface PermissionResult {
  /** Whether user has the required permission */
  hasPermission: boolean;
  /** Whether user is on Pro tier or higher */
  isPro: boolean;
  /** Whether user is on Enterprise tier */
  isEnterprise: boolean;
  /** Whether user is on Free tier */
  isFree: boolean;
  /** Current user tier */
  currentTier: UserRole;
  /** Next tier for upgrade (null if already at highest) */
  nextTier: UserRole | null;
  /** Show upgrade prompt callback */
  showUpgradePrompt: () => void;
  /** Get missing tier message */
  getMissingTierMessage: (featureName: string) => string;
}

// ============================================================================
// Constants
// ============================================================================

/** Tier hierarchy from lowest to highest */
const TIER_ORDER: UserRole[] = ["free", "pro", "enterprise"];

/** Tier display names */
const TIER_NAMES: Record<UserRole, string> = {
  free: "Free",
  pro: "Pro",
  enterprise: "Enterprise",
};

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Check if current tier meets required tier
 */
export function checkTierAccess(
  currentTier: UserRole,
  requiredTier: UserRole
): boolean {
  const currentIndex = TIER_ORDER.indexOf(currentTier);
  const requiredIndex = TIER_ORDER.indexOf(requiredTier);
  return currentIndex >= requiredIndex;
}

/**
 * Get next tier in hierarchy
 */
export function getNextTier(currentTier: UserRole): UserRole | null {
  const currentIndex = TIER_ORDER.indexOf(currentTier);
  if (currentIndex < TIER_ORDER.length - 1) {
    return TIER_ORDER[currentIndex + 1];
  }
  return null;
}

/**
 * Get tier index (for comparisons)
 */
export function getTierIndex(tier: UserRole): number {
  return TIER_ORDER.indexOf(tier);
}

// ============================================================================
// Hook
// ============================================================================

/**
 * usePermission - Hook for permission checking
 *
 * Provides utilities for checking user tier and permissions.
 * Integrates with FlyPermissionGuard.
 *
 * @param requiredTier - Minimum tier required for access
 * @param currentTier - Current user's tier
 * @returns Permission check result
 *
 * @example
 * ```tsx
 * const { hasPermission, isPro, showUpgradePrompt } = usePermission("pro", user.role);
 *
 * if (!hasPermission) {
 *   return <UpgradePrompt onClick={showUpgradePrompt} />;
 * }
 *
 * return <PremiumFeature />;
 * ```
 */
export function usePermission(
  requiredTier: UserRole,
  currentTier: UserRole
): PermissionResult {
  // Memoize permission checks
  const permissionState = useMemo(() => {
    const hasPermission = checkTierAccess(currentTier, requiredTier);
    const isPro = checkTierAccess(currentTier, "pro");
    const isEnterprise = currentTier === "enterprise";
    const isFree = currentTier === "free";
    const nextTier = getNextTier(currentTier);

    return {
      hasPermission,
      isPro,
      isEnterprise,
      isFree,
      nextTier,
    };
  }, [requiredTier, currentTier]);

  /**
   * Show upgrade prompt
   * Dispatches a custom event that can be listened to by UI components
   */
  const showUpgradePrompt = useCallback(() => {
    const targetTier = permissionState.nextTier || requiredTier;

    window.dispatchEvent(
      new CustomEvent("fly:showUpgradePrompt", {
        detail: {
          currentTier,
          requiredTier,
          targetTier,
          timestamp: Date.now(),
        },
      })
    );
  }, [currentTier, requiredTier, permissionState.nextTier]);

  /**
   * Get message about missing tier requirement
   */
  const getMissingTierMessage = useCallback(
    (featureName: string): string => {
      if (permissionState.hasPermission) {
        return "";
      }

      const targetTier = permissionState.nextTier || requiredTier;
      return `${featureName} requires a ${TIER_NAMES[targetTier]} plan or higher. You are currently on the ${TIER_NAMES[currentTier]} plan.`;
    },
    [permissionState.hasPermission, permissionState.nextTier, requiredTier, currentTier]
  );

  return {
    ...permissionState,
    currentTier,
    showUpgradePrompt,
    getMissingTierMessage,
  };
}

// ============================================================================
// Convenience Hooks
// ============================================================================

/**
 * Hook to check if user is Pro or higher
 */
export function useIsPro(currentTier: UserRole): boolean {
  return useMemo(() => checkTierAccess(currentTier, "pro"), [currentTier]);
}

/**
 * Hook to check if user is Enterprise
 */
export function useIsEnterprise(currentTier: UserRole): boolean {
  return useMemo(() => currentTier === "enterprise", [currentTier]);
}

/**
 * Hook to get tier comparison utilities
 */
export function useTierComparison(currentTier: UserRole) {
  return useMemo(
    () => ({
      isAtLeast: (tier: UserRole) => checkTierAccess(currentTier, tier),
      isExactly: (tier: UserRole) => currentTier === tier,
      isBelow: (tier: UserRole) => !checkTierAccess(currentTier, tier),
      nextTier: getNextTier(currentTier),
      tierIndex: getTierIndex(currentTier),
      compareTo: (tier: UserRole) => {
        const current = getTierIndex(currentTier);
        const other = getTierIndex(tier);
        return current - other;
      },
    }),
    [currentTier]
  );
}

export default usePermission;
