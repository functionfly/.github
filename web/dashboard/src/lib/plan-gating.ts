/**
 * Plan Gating Utilities for Onboarding
 *
 * This module provides plan-based feature gating to ensure users only see
 * features and steps available to their plan tier.
 */

import type { PlanTier } from '@/stores/onboardingStore';
import { PLANS } from './constants';

export { type PlanTier };

/**
 * Get the maximum number of custom domains allowed for a plan
 */
export function getCustomDomainsLimit(plan: PlanTier): number {
  const planLimits = PLANS[plan.toUpperCase() as keyof typeof PLANS]?.limits;
  if (!planLimits) return 0;
  return planLimits.customDomains ?? 0;
}

/**
 * Get the maximum number of regions allowed for a plan
 */
export function getRegionsLimit(plan: PlanTier): number {
  const planLimits = PLANS[plan.toUpperCase() as keyof typeof PLANS]?.limits;
  if (!planLimits) return 1;
  return planLimits.providers ?? 1;
}

/**
 * Get the maximum number of API keys allowed for a plan
 */
export function getApiKeysLimit(plan: PlanTier): number {
  switch (plan) {
    case 'free':
      return 1;
    case 'starter':
      return 5;
    case 'professional':
      return 20;
    case 'enterprise':
    case 'agent_enterprise':
      return -1;
    default:
      return 1;
  }
}

/**
 * Get the maximum number of custom environments allowed for a plan
 */
export function getCustomEnvironmentsLimit(plan: PlanTier): number {
  switch (plan) {
    case 'free':
    case 'starter':
      return 3;
    case 'professional':
      return 10;
    case 'enterprise':
    case 'agent_enterprise':
      return -1;
    default:
      return 0;
  }
}

/**
 * Get the maximum number of team members allowed for a plan
 */
export function getTeamMembersLimit(plan: PlanTier): number {
  switch (plan) {
    case 'free':
      return 1;
    case 'starter':
      return 3;
    case 'professional':
      return 10;
    case 'enterprise':
    case 'agent_enterprise':
      return -1;
    default:
      return 1;
  }
}

/**
 * Check if custom domains feature is available for the plan
 */
export function hasCustomDomainsFeature(plan: PlanTier): boolean {
  return getCustomDomainsLimit(plan) > 0;
}

/**
 * Check if the plan selection step should show upgrade options
 */
export function shouldShowUpgradeOptions(currentPlan: PlanTier): boolean {
  return currentPlan === 'free';
}

/**
 * Get plan display info
 */
export function getPlanInfo(plan: PlanTier): {
  name: string;
  price: number;
  priceLabel: string;
  features: string[];
} {
  const planData = PLANS[plan.toUpperCase() as keyof typeof PLANS];
  if (!planData) {
    return { name: 'Unknown', price: 0, priceLabel: '$0', features: [] };
  }

  return {
    name: planData.name,
    price: planData.price,
    priceLabel: planData.price === 0 ? 'Free' : `$${planData.price}/mo`,
    features: [...planData.features],
  };
}

/**
 * Format the remaining limit for display
 */
export function formatLimitRemaining(current: number, max: number): string {
  if (max === -1) return 'Unlimited';
  return `${current} of ${max}`;
}

/**
 * Check if user can add more of a resource
 */
export function canAddMore(current: number, max: number): boolean {
  if (max === -1) return true;
  return current < max;
}

/**
 * Check if a specific integration type is available for the plan
 */
export function isIntegrationAvailable(
  integrationType: string,
  plan: PlanTier
): boolean {
  const basicIntegrations = ['slack', 'discord', 'github'];
  if (basicIntegrations.includes(integrationType)) {
    return true;
  }

  const advancedIntegrations = ['sentry', 'datadog', 'newrelic'];
  if (advancedIntegrations.includes(integrationType)) {
    return plan === 'professional' || plan === 'enterprise' || plan === 'agent_enterprise';
  }

  return false;
}

/**
 * Check if custom webhooks are available for the plan
 */
export function hasCustomWebhooks(plan: PlanTier): boolean {
  return plan === 'professional' || plan === 'enterprise' || plan === 'agent_enterprise';
}
