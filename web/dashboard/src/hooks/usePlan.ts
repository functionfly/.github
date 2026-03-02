import { useAuthStore } from '@/stores/authStore';
import {
  isEnterprise,
  hasMinPlan,
  getPlanLimits,
  hasFeature,
  getAvailableFeatures,
  hasUnlimitedResources,
  getPlanColor,
  formatPlanName,
  type PlanTier,
  type FeatureKey,
} from '@/lib/plan-utils';

interface UsePlanReturn {
  /** Current plan identifier */
  plan: string | undefined;
  /** Whether user is on enterprise plan */
  isEnterprise: boolean;
  /** Whether user has any paid plan */
  isPaid: boolean;
  /** Plan limits object */
  limits: ReturnType<typeof getPlanLimits>;
  /** Check if user has a specific feature */
  hasFeature: (feature: FeatureKey) => boolean;
  /** Check if user has at least the specified plan tier */
  hasMinPlan: (minPlan: PlanTier) => boolean;
  /** Get list of all available features */
  availableFeatures: FeatureKey[];
  /** Whether plan has unlimited resources */
  hasUnlimited: boolean;
  /** Color theme for the plan */
  planColor: string;
  /** Formatted plan name for display */
  displayName: string;
}

/**
 * Hook to access plan-related functionality
 * Provides plan checking utilities and feature gating
 */
export const usePlan = (): UsePlanReturn => {
  const user = useAuthStore((state) => state.user);
  const plan = user?.plan;

  return {
    plan,
    isEnterprise: isEnterprise(plan),
    isPaid: hasMinPlan(plan ?? '', 'starter'),
    limits: getPlanLimits(plan),
    hasFeature: (feature: FeatureKey) => hasFeature(plan, feature),
    hasMinPlan: (minPlan: PlanTier) => hasMinPlan(plan ?? '', minPlan),
    availableFeatures: getAvailableFeatures(plan),
    hasUnlimited: hasUnlimitedResources(plan),
    planColor: getPlanColor(plan),
    displayName: formatPlanName(plan),
  };
};

export type { PlanTier, FeatureKey };
