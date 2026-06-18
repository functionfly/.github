/**
 * Vault Plan Gating - Unified with Main Platform Plans
 *
 * This module maps vault-specific features to the main platform plan tiers.
 * The single source of truth for plan limits is web/dashboard/src/lib/constants.ts.
 * This module adds vault-specific feature flags on top.
 *
 * Plan mapping:
 * - platform FREE → vault free tier
 * - platform STARTER → vault free tier (no vault features)
 * - platform PROFESSIONAL → vault pro tier (MFA, IP allowlist, break-glass, etc.)
 * - platform ENTERPRISE → vault team tier (RBAC, shares, SIEM, etc.)
 * - platform AGENT_ENTERPRISE → vault enterprise tier (SSO, HA status)
 */

import { PLANS } from './constants';
import type { PlanLimits, VaultPlan } from "@/types/vault-enterprise";

type PlatformPlan = keyof typeof PLANS;

/**
 * Maps platform plan to vault plan tier
 */
export function platformToVaultPlan(platformPlan: string): VaultPlan {
  switch (platformPlan.toLowerCase()) {
    case 'free':
      return 'free';
    case 'starter':
      return 'free';
    case 'professional':
      return 'pro';
    case 'enterprise':
      return 'team';
    case 'agent_enterprise':
      return 'enterprise';
    default:
      return 'free';
  }
}

/**
 * Vault feature flags per plan
 */
const VAULT_FEATURES: Record<VaultPlan, PlanLimits['features']> = {
  free: {
    mfa: false,
    ipAllowlist: false,
    expiration: true,
    breakGlass: false,
    escrow: false,
    rbac: false,
    namespaces: true,
    shares: false,
    sso: false,
    siemWebhooks: false,
    auditExport: false,
    cacheStats: false,
    quotaWidget: true,
    haStatus: false,
    dependencyGraph: false,
    expirationDashboard: true,
    tokenMonitor: false,
  },
  pro: {
    mfa: true,
    ipAllowlist: true,
    expiration: true,
    breakGlass: true,
    escrow: false,
    rbac: false,
    namespaces: true,
    shares: false,
    sso: false,
    siemWebhooks: false,
    auditExport: true,
    cacheStats: true,
    quotaWidget: true,
    haStatus: false,
    dependencyGraph: false,
    expirationDashboard: true,
    tokenMonitor: true,
  },
  team: {
    mfa: true,
    ipAllowlist: true,
    expiration: true,
    breakGlass: true,
    escrow: true,
    rbac: true,
    namespaces: true,
    shares: true,
    sso: false,
    siemWebhooks: true,
    auditExport: true,
    cacheStats: true,
    quotaWidget: true,
    haStatus: false,
    dependencyGraph: true,
    expirationDashboard: true,
    tokenMonitor: true,
  },
  enterprise: {
    mfa: true,
    ipAllowlist: true,
    expiration: true,
    breakGlass: true,
    escrow: true,
    rbac: true,
    namespaces: true,
    shares: true,
    sso: true,
    siemWebhooks: true,
    auditExport: true,
    cacheStats: true,
    quotaWidget: true,
    haStatus: true,
    dependencyGraph: true,
    expirationDashboard: true,
    tokenMonitor: true,
  },
};

/**
 * Vault-specific limits per plan (beyond what constants.ts provides)
 */
const VAULT_LIMITS: Record<VaultPlan, Omit<PlanLimits, 'features'>> = {
  free: {
    maxSecrets: 25,
    maxDynamicCreds: 100,
    tokensPerSecret: 5,
    auditExportsPerDay: 1,
    dynamicBackends: [],
  },
  pro: {
    maxSecrets: 500,
    maxDynamicCreds: 5000,
    tokensPerSecret: 25,
    auditExportsPerDay: 10,
    dynamicBackends: ['postgres'],
  },
  team: {
    maxSecrets: 5000,
    maxDynamicCreds: 50000,
    tokensPerSecret: 100,
    auditExportsPerDay: 50,
    dynamicBackends: ['postgres', 'mysql'],
  },
  enterprise: {
    maxSecrets: 1000000,
    maxDynamicCreds: 1000000,
    tokensPerSecret: 1000,
    auditExportsPerDay: 1000,
    dynamicBackends: ['postgres', 'mysql'],
  },
};

/**
 * Get vault plan limits for a platform plan
 */
export function getVaultPlanLimits(platformPlan: string): PlanLimits {
  const vaultPlan = platformToVaultPlan(platformPlan);
  return {
    ...VAULT_LIMITS[vaultPlan],
    features: VAULT_FEATURES[vaultPlan],
  };
}

/**
 * getPlanLimits returns the limits + feature matrix for a vault plan.
 * @deprecated Use getVaultPlanLimits(platformPlan) instead for unified plan system
 */
export function getPlanLimits(plan: VaultPlan): PlanLimits {
  return ConstLike(plan);
}

function ConstLike(plan: VaultPlan): PlanLimits {
  return {
    ...VAULT_LIMITS[plan],
    features: VAULT_FEATURES[plan],
  };
}

/**
 * hasFeature reports whether a plan includes a specific vault feature.
 */
export function hasFeature(plan: VaultPlan, feature: keyof PlanLimits['features']): boolean {
  return getPlanLimits(plan).features[feature] === true;
}

/**
 * Plan display metadata.
 */
export const PLAN_META: Record<VaultPlan, { name: string; price: string; tagline: string }> = {
  free: { name: 'Free', price: '$0/mo', tagline: 'Get started with a personal vault' },
  pro: { name: 'Pro', price: '$79/mo', tagline: 'For serious solo developers' },
  team: { name: 'Team', price: '$299/mo', tagline: 'For small teams' },
  enterprise: {
    name: 'Enterprise',
    price: '$499/mo',
    tagline: 'SAML SSO, HA, and cross-tenant sharing',
  },
};

/**
 * Upgrades needed to access a vault feature.
 */
export const FEATURE_MIN_PLAN: Record<keyof PlanLimits['features'], VaultPlan> = {
  mfa: 'pro',
  ipAllowlist: 'pro',
  expiration: 'free',
  breakGlass: 'pro',
  escrow: 'team',
  rbac: 'team',
  namespaces: 'free',
  shares: 'team',
  sso: 'enterprise',
  siemWebhooks: 'team',
  auditExport: 'pro',
  cacheStats: 'pro',
  quotaWidget: 'free',
  haStatus: 'enterprise',
  dependencyGraph: 'team',
  expirationDashboard: 'free',
  tokenMonitor: 'pro',
};

/**
 * minPlanForFeature returns the lowest plan that unlocks a vault feature.
 */
export function minPlanForFeature(
  feature: keyof PlanLimits['features'],
): VaultPlan {
  return FEATURE_MIN_PLAN[feature];
}

/**
 * PLAN_LIMITS combines VAULT_LIMITS and VAULT_FEATURES for direct access.
 * @deprecated Use getPlanLimits() instead
 */
export const PLAN_LIMITS: Record<VaultPlan, PlanLimits> = {
  free: { ...VAULT_LIMITS.free, features: VAULT_FEATURES.free },
  pro: { ...VAULT_LIMITS.pro, features: VAULT_FEATURES.pro },
  team: { ...VAULT_LIMITS.team, features: VAULT_FEATURES.team },
  enterprise: { ...VAULT_LIMITS.enterprise, features: VAULT_FEATURES.enterprise },
};

/**
 * isFeatureAvailable is a convenience over hasFeature.
 */
export function isFeatureAvailable(
  plan: VaultPlan,
  feature: keyof PlanLimits['features'],
): boolean {
  return hasFeature(plan, feature);
}
