/**
 * Vault Plan Gating
 *
 * Single source of truth for "which plan tier has which feature" and
 * the hard quotas enforced server-side (mirrored client-side so we
 * can show "X of Y used" without a round-trip).
 *
 * When the tenant's plan is upgraded, this matrix and the server's
 * quota plan matrix in internal/storage/vault/quota must move
 * together.
 */

import type { PlanLimits, VaultPlan } from "@/types/vault-enterprise";

const FREE: PlanLimits = {
  maxSecrets: 25,
  maxDynamicCreds: 100,
  tokensPerSecret: 5,
  auditExportsPerDay: 1,
  dynamicBackends: [],
  features: {
    mfa: false,
    ipAllowlist: false,
    expiration: true, // basic expiration is free
    breakGlass: false,
    escrow: false,
    rbac: false,
    namespaces: true, // namespaces themselves are free; they just lack cross-tenant features
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
};

const PRO: PlanLimits = {
  maxSecrets: 500,
  maxDynamicCreds: 5_000,
  tokensPerSecret: 25,
  auditExportsPerDay: 10,
  dynamicBackends: ["postgres"],
  features: {
    ...FREE.features,
    mfa: true,
    ipAllowlist: true,
    breakGlass: true, // basic self-service break-glass on Pro
    auditExport: true,
    cacheStats: true,
    quotaWidget: true,
    expirationDashboard: true,
    tokenMonitor: true,
  },
};

const TEAM: PlanLimits = {
  maxSecrets: 5_000,
  maxDynamicCreds: 50_000,
  tokensPerSecret: 100,
  auditExportsPerDay: 50,
  dynamicBackends: ["postgres", "mysql"],
  features: {
    ...PRO.features,
    escrow: true,
    rbac: true,
    namespaces: true,
    shares: true,
    siemWebhooks: true,
    dependencyGraph: true,
    expirationDashboard: true,
    tokenMonitor: true,
  },
};

const ENTERPRISE: PlanLimits = {
  maxSecrets: 1_000_000,
  maxDynamicCreds: 1_000_000,
  tokensPerSecret: 1_000,
  auditExportsPerDay: 1_000,
  dynamicBackends: ["postgres", "mysql"],
  features: {
    ...TEAM.features,
    sso: true,
    haStatus: true,
    shares: true,
  },
};

export const PLAN_LIMITS: Record<VaultPlan, PlanLimits> = {
  free: FREE,
  pro: PRO,
  team: TEAM,
  enterprise: ENTERPRISE,
};

/**
 * getPlanLimits returns the limits + feature matrix for a plan.
 */
export function getPlanLimits(plan: VaultPlan): PlanLimits {
  return PLAN_LIMITS[plan] ?? FREE;
}

/**
 * hasFeature reports whether a plan includes a specific feature.
 */
export function hasFeature(plan: VaultPlan, feature: keyof PlanLimits["features"]): boolean {
  return getPlanLimits(plan).features[feature] === true;
}

/**
 * Plan display metadata.
 */
export const PLAN_META: Record<VaultPlan, { name: string; price: string; tagline: string }> = {
  free: { name: "Free", price: "$0/mo", tagline: "Get started with a personal vault" },
  pro: { name: "Pro", price: "$29/mo", tagline: "For serious solo developers" },
  team: { name: "Team", price: "$99/mo", tagline: "For small teams" },
  enterprise: {
    name: "Enterprise",
    price: "Contact us",
    tagline: "SAML SSO, HA, and cross-tenant sharing",
  },
};

/**
 * Upgrades needed to access a feature.
 */
export const FEATURE_MIN_PLAN: Record<keyof PlanLimits["features"], VaultPlan> = {
  mfa: "pro",
  ipAllowlist: "pro",
  expiration: "free",
  breakGlass: "pro",
  escrow: "team",
  rbac: "team",
  namespaces: "free",
  shares: "team",
  sso: "enterprise",
  siemWebhooks: "team",
  auditExport: "pro",
  cacheStats: "pro",
  quotaWidget: "free",
  haStatus: "enterprise",
  dependencyGraph: "team",
  expirationDashboard: "free",
  tokenMonitor: "pro",
};

/**
 * minPlanForFeature returns the lowest plan that unlocks `feature`.
 */
export function minPlanForFeature(
  feature: keyof PlanLimits["features"],
): VaultPlan {
  return FEATURE_MIN_PLAN[feature];
}

/**
 * isFeatureAvailable is a convenience over hasFeature.
 */
export function isFeatureAvailable(
  plan: VaultPlan,
  feature: keyof PlanLimits["features"],
): boolean {
  return hasFeature(plan, feature);
}
