import { AGENT_ENTERPRISE, PLANS } from './constants';

export type PlanTier =
  | 'free'
  | 'starter'
  | 'professional'
  | 'enterprise'
  | 'enterprise_sla'
  | 'agent_enterprise';

export const PLAN_HIERARCHY: Record<PlanTier, number> = {
  free: 0,
  starter: 1,
  professional: 2,
  enterprise: 3,
  enterprise_sla: 4, // Enhanced Enterprise with premium SLA
  agent_enterprise: 5, // Top tier - unlimited
};

/**
 * Check if user is on enterprise plan (base or SLA)
 */
export const isEnterprise = (plan?: string): boolean => {
  const p = plan?.toLowerCase();
  return p === 'enterprise' || p === 'enterprise_sla';
};

/**
 * Check if user is on Enterprise SLA plan (premium SLA)
 */
export const isEnterpriseSLA = (plan?: string): boolean => plan?.toLowerCase() === 'enterprise_sla';

/**
 * Get the SLA target percentage for the plan
 */
export const getSLATargetPercent = (plan?: string): number => {
  if (plan?.toLowerCase() === 'enterprise_sla') {
    return 99.999;
  }
  if (plan?.toLowerCase() === 'enterprise') {
    return 99.99;
  }
  if (plan?.toLowerCase() === 'professional') {
    return 99.9;
  }
  if (plan?.toLowerCase() === 'starter') {
    return 99.5;
  }
  return 0; // No SLA for free tier
};

/**
 * Check if the plan has premium SLA features (99.999%)
 */
export const hasPremiumSLA = (plan?: string): boolean => plan?.toLowerCase() === 'enterprise_sla';

/**
 * Check if SLA credits are enabled for the plan
 */
export const hasSLACredits = (plan?: string): boolean => {
  const p = plan?.toLowerCase();
  return p === 'enterprise_sla' || p === 'enterprise' || p === 'agent_enterprise';
};

/**
 * Check if user has at least the specified plan tier
 */
export const hasMinPlan = (userPlan: string, minPlan: PlanTier): boolean => {
  const userTier = PLAN_HIERARCHY[userPlan.toLowerCase() as PlanTier] ?? 0;
  const requiredTier = PLAN_HIERARCHY[minPlan];
  return userTier >= requiredTier;
};

/**
 * Get plan display name with enterprise styling
 */
export const getPlanDisplayName = (plan?: string): string => {
  const planKey = plan?.toUpperCase() as keyof typeof PLANS;
  return PLANS[planKey]?.name || plan || 'Unknown';
};

/**
 * Get plan limits for the user's tier
 */
export const getPlanLimits = (plan?: string) => {
  const planKey = plan?.toUpperCase() as keyof typeof PLANS;
  return PLANS[planKey]?.limits;
};

/**
 * Get the maximum number of secrets allowed for the user's tier
 * Returns Infinity for enterprise, 0 for free tier (no secrets)
 */
export const getSecretsLimit = (plan?: string): number => {
  const limits = getPlanLimits(plan);
  if (!limits) return 0;
  // Handle Infinity case for enterprise
  if (limits.secrets === Infinity) return 10000;
  return limits.secrets ?? 0;
};

/**
 * Get the maximum number of tokens per secret allowed for the user's tier
 * Returns 0 for free tier (no tokens)
 */
export const getTokensPerSecretLimit = (plan?: string): number => {
  const limits = getPlanLimits(plan);
  if (!limits) return 0;
  // Handle Infinity case for enterprise
  if (limits.tokensPerSecret === Infinity) return 100;
  return limits.tokensPerSecret ?? 0;
};

/**
 * Get the maximum number of custom domains allowed for the user's tier
 * Returns 0 for free tier (no custom domains)
 */
export const getCustomDomainsLimit = (plan?: string): number => {
  const limits = getPlanLimits(plan);
  if (!limits) return 0;
  const raw = (limits as { customDomains?: number }).customDomains;
  if (raw === undefined) return 0;
  return raw === Infinity ? 10000 : raw;
};

/**
 * Check if the user can add more custom domains (has feature and under limit)
 */
export const canAddCustomDomain = (plan: string | undefined, currentCount: number): boolean => {
  if (!hasFeature(plan, 'CUSTOM_DOMAINS')) return false;
  const limit = getCustomDomainsLimit(plan);
  if (limit === 0) return false;
  return currentCount < limit;
};

/**
 * Format custom domains usage for display (e.g. "2 of 5" or "Unlimited")
 */
export const formatCustomDomainsRemaining = (currentCount: number, plan?: string): string => {
  const limit = getCustomDomainsLimit(plan);
  if (limit === 0) return 'Not available on your plan';
  if (limit >= 10000) return `${currentCount} (unlimited)`;
  return `${currentCount} of ${limit}`;
};

/**
 * Max state fabrics for the tier (from PLANS.limits.stateFabrics).
 * Enterprise unlimited is represented as a large cap for display comparisons.
 */
export const getStateFabricsLimit = (plan?: string): number => {
  const limits = getPlanLimits(plan);
  if (!limits) return 0;
  const raw = (limits as { stateFabrics?: number }).stateFabrics;
  if (raw === undefined) return 0;
  return raw === Infinity ? 10000 : raw;
};

/**
 * Max agents for the tier (from PLANS.limits.agents).
 */
export const getAgentsLimit = (plan?: string): number => {
  // Handle agent_enterprise specially (unlimited)
  if (plan === 'agent_enterprise') {
    return AGENT_ENTERPRISE.limits.agents as number;
  }
  const limits = getPlanLimits(plan);
  if (!limits) return 0;
  const raw = (limits as { agents?: number }).agents;
  if (raw === undefined) return 0;
  return raw === Infinity ? 10000 : raw;
};

/**
 * Max apps for the tier (from PLANS.limits.apps).
 */
export const getAppsLimit = (plan?: string): number => {
  if (plan === 'agent_enterprise') {
    return -1; // Unlimited
  }
  const limits = getPlanLimits(plan);
  if (!limits) return 0;
  const raw = (limits as { apps?: number }).apps;
  if (raw === undefined) return 0;
  return raw === Infinity ? -1 : raw;
};

/**
 * Max AI calls per month for the tier (bundled agent capability).
 */
export const getAICallsLimit = (plan?: string): number => {
  if (plan === 'agent_enterprise') {
    return -1; // Unlimited
  }
  const limits = getPlanLimits(plan);
  if (!limits) return 0;
  const raw = (limits as { aiCallsPerMonth?: number }).aiCallsPerMonth;
  return raw ?? 0;
};

/**
 * Agent concurrency limit for the tier.
 */
export const getAgentConcurrencyLimit = (plan?: string): number => {
  if (plan === 'agent_enterprise') {
    return -1; // Unlimited
  }
  const limits = getPlanLimits(plan);
  if (!limits) return 0;
  const raw = (limits as { agentConcurrency?: number }).agentConcurrency;
  return raw ?? 0;
};

/**
 * Agent calls per minute limit for the tier.
 */
export const getAgentCallsPerMinuteLimit = (plan?: string): number => {
  if (plan === 'agent_enterprise') {
    return -1; // Unlimited
  }
  const limits = getPlanLimits(plan);
  if (!limits) return 0;
  const raw = (limits as { agentCallsPerMinute?: number }).agentCallsPerMinute;
  return raw ?? 0;
};

/**
 * Whether the user may create another state fabric (plan feature + under limit).
 */
export const canCreateStateFabric = (plan: string | undefined, currentCount: number): boolean => {
  if (!hasFeature(plan, 'STATE_FABRIC')) return false;
  const limit = getStateFabricsLimit(plan);
  if (limit === 0) return false;
  if (limit >= 10000) return true;
  return currentCount < limit;
};

/**
 * Whether the user may register another agent (plan feature + under limit).
 */
export const canCreateAgent = (plan: string | undefined, currentCount: number): boolean => {
  if (!hasFeature(plan, 'AGENTS')) return false;
  const limit = getAgentsLimit(plan);
  if (limit === 0) return false;
  if (limit < 0) return true; // Negative means unlimited (agent_enterprise)
  if (limit >= 10000) return true;
  return currentCount < limit;
};

/**
 * Check if the user can create secrets based on their tier
 * Returns true if they have at least some secrets allowed
 */
export const canCreateSecrets = (plan?: string): boolean => {
  return getSecretsLimit(plan) > 0;
};

/**
 * Check if the user can create tokens based on their tier
 * Returns true if they have at least some tokens allowed
 */
export const canCreateTokens = (plan?: string): boolean => {
  return getTokensPerSecretLimit(plan) > 0;
};

/**
 * Format secrets remaining for display
 * Shows "X of Y used" or "X remaining" or "Unlimited"
 */
export const formatSecretsRemaining = (currentCount: number, plan?: string): string => {
  const limit = getSecretsLimit(plan);

  if (limit === Infinity || limit >= 10000) {
    return `${currentCount} secrets (unlimited)`;
  }

  if (limit === 0) {
    return 'Secrets not available on your plan';
  }

  const remaining = limit - currentCount;
  return `${currentCount} of ${limit} secrets used (${remaining} remaining)`;
};

/**
 * Feature availability by plan tier
 */
export const FEATURES: Record<string, readonly PlanTier[]> = {
  ADVANCED_ANALYTICS: ['professional', 'enterprise', 'enterprise_sla'],
  CUSTOM_DASHBOARDS: ['enterprise', 'enterprise_sla'],
  AUDIT_LOGS: ['enterprise', 'enterprise_sla'],
  SLA_DASHBOARD: ['enterprise', 'enterprise_sla'],
  /** Premium SLA with 99.999% uptime - Enterprise SLA only */
  SLA_PREMIUM: ['enterprise_sla'],
  /** SLA credits for violations - Enterprise and above */
  SLA_CREDITS: ['enterprise', 'enterprise_sla', 'agent_enterprise'],
  DEDICATED_SUPPORT: ['enterprise', 'enterprise_sla'],
  EXPORT_REPORTS: ['enterprise', 'enterprise_sla'],
  TEAM_MANAGEMENT: ['professional', 'enterprise', 'enterprise_sla'],
  API_ACCESS: ['starter', 'professional', 'enterprise', 'enterprise_sla'],
  WEBHOOKS: ['professional', 'enterprise', 'enterprise_sla'],
  CUSTOM_DOMAINS: ['starter', 'professional', 'enterprise', 'enterprise_sla'],
  /** Stateful fabrics: Free has 1 (Sandbox), paid tiers per PLANS.limits.stateFabrics */
  STATE_FABRIC: [
    'free',
    'starter',
    'professional',
    'enterprise',
    'enterprise_sla',
    'agent_enterprise',
  ],
  /** AI agents: Free has 0 quota; paid tiers per PLANS.limits.agents */
  AGENTS: ['free', 'starter', 'professional', 'enterprise', 'enterprise_sla', 'agent_enterprise'],
  UNLIMITED_FUNCTIONS: ['enterprise', 'enterprise_sla'],
  UNLIMITED_PROVIDERS: ['enterprise', 'enterprise_sla'],
  PRIORITY_SUPPORT: ['professional', 'enterprise', 'enterprise_sla'],
  /** Enterprise sidebar section with SLA, Audit, Support */
  ENTERPRISE_SECTION: ['professional', 'enterprise', 'enterprise_sla'],
  /** Time Machine: 24h replay, basic diff — available on all plans */
  TIME_MACHINE_BASIC: [
    'free',
    'starter',
    'professional',
    'enterprise',
    'enterprise_sla',
    'agent_enterprise',
  ],
  /** Time Machine: 72h replay window */
  TIME_MACHINE_EXTENDED: [
    'starter',
    'professional',
    'enterprise',
    'enterprise_sla',
    'agent_enterprise',
  ],
  /** Time Machine: 30-day replay, full diffs, dry-run reconciliation */
  TIME_MACHINE_PRO: ['professional', 'enterprise', 'enterprise_sla', 'agent_enterprise'],
  /** Time Machine: 90-day replay, live reconciliation, audit certificates */
  TIME_MACHINE_ENTERPRISE: ['enterprise', 'enterprise_sla', 'agent_enterprise'],
  /** Time Machine: unlimited replay, custom rules, legal-grade certs */
  TIME_MACHINE_UNLIMITED: ['agent_enterprise'],
  /** Incident Insurance: dedicated engineer during critical incidents */
  TIME_MACHINE_INSURANCE: ['agent_enterprise'],
  /** Function DNA: AI-powered code evolution based on execution patterns */
  FUNCTION_DNA: ['starter', 'professional', 'enterprise', 'enterprise_sla', 'agent_enterprise'],
  /** Studio: AI-powered code & function studio */
  STUDIO: ['agent_enterprise'],
  /** State Fabric Add-ons: available for purchase on paid plans (Starter+) */
  SF_ADDON_HOT_CACHE: [
    'starter',
    'professional',
    'enterprise',
    'enterprise_sla',
    'agent_enterprise',
  ],
  SF_ADDON_MULTI_REGION: [
    'starter',
    'professional',
    'enterprise',
    'enterprise_sla',
    'agent_enterprise',
  ],
  SF_ADDON_AI_RECALL: [
    'starter',
    'professional',
    'enterprise',
    'enterprise_sla',
    'agent_enterprise',
  ],
  SF_ADDON_ADVANCED_INSIGHTS: [
    'starter',
    'professional',
    'enterprise',
    'enterprise_sla',
    'agent_enterprise',
  ],
  SF_ADDON_ADVANCED_SECURITY: [
    'starter',
    'professional',
    'enterprise',
    'enterprise_sla',
    'agent_enterprise',
  ],
} as const;

export type FeatureKey = keyof typeof FEATURES;

/**
 * Check if a feature is available for the plan
 */
export const hasFeature = (plan: string | undefined, feature: FeatureKey): boolean => {
  if (!plan) return false;
  const allowedPlans = FEATURES[feature];
  return allowedPlans.includes(plan.toLowerCase() as PlanTier);
};

/**
 * Get all features available for a plan
 */
export const getAvailableFeatures = (plan?: string): FeatureKey[] => {
  if (!plan) return [];
  return (Object.keys(FEATURES) as FeatureKey[]).filter((feature) => hasFeature(plan, feature));
};

/**
 * Check if plan has unlimited resources
 */
export const hasUnlimitedResources = (plan?: string): boolean => isEnterprise(plan);

/**
 * Get plan color for UI theming
 */
export const getPlanColor = (plan?: string): string => {
  switch (plan?.toLowerCase()) {
    case 'enterprise':
    case 'enterprise_sla':
      return 'amber';
    case 'professional':
      return 'indigo';
    case 'starter':
      return 'emerald';
    case 'free':
    default:
      return 'slate';
  }
};

/**
 * Format plan name for display with proper capitalization
 */
export const formatPlanName = (plan?: string): string => {
  if (!plan) return 'Unknown';
  return plan.charAt(0).toUpperCase() + plan.slice(1).toLowerCase();
};

/**
 * Get overage rate for a plan (cents per 1000 requests)
 * Returns null if the plan has a hard stop (no overage)
 */
export const getOverageRate = (plan?: string): number | null => {
  const planKey = plan?.toUpperCase() as keyof typeof PLANS;
  return PLANS[planKey]?.overageRate ?? null;
};

/**
 * Get annual discount for a plan (0.0 - 1.0)
 */
export const getAnnualDiscount = (plan?: string): number => {
  const planKey = plan?.toUpperCase() as keyof typeof PLANS;
  return PLANS[planKey]?.annualDiscount ?? 0;
};

/**
 * Get annual price for a plan
 */
export const getAnnualPrice = (plan?: string): number | null => {
  const planKey = plan?.toUpperCase() as keyof typeof PLANS;
  const planData = PLANS[planKey];
  if (!planData || typeof planData.price !== 'number' || planData.price === 0) return null;
  const monthly = planData.price * (1 - (planData.annualDiscount || 0));
  return Math.round(monthly * 12);
};

/**
 * Check if a plan supports per-API-key cost attribution
 */
export const hasPerKeyCostAttribution = (plan?: string): boolean => {
  const limits = getPlanLimits(plan);
  if (!limits) return false;
  return (limits as { perKeyCostAttribution?: boolean }).perKeyCostAttribution === true;
};

/**
 * Check if a plan supports API key budgets and alerts
 */
export const hasAPIKeyBudgets = (plan?: string): boolean => {
  const limits = getPlanLimits(plan);
  if (!limits) return false;
  return (limits as { apiKeyBudgets?: boolean }).apiKeyBudgets === true;
};

/**
 * Check if a plan supports high-value key separation
 */
export const hasHighValueKeySeparation = (plan?: string): boolean => {
  const limits = getPlanLimits(plan);
  if (!limits) return false;
  return (limits as { highValueKeySeparation?: boolean }).highValueKeySeparation === true;
};

/**
 * Format overage rate for display
 */
export const formatOverageRate = (plan?: string): string => {
  const rate = getOverageRate(plan);
  if (rate === null) return 'Hard stop';
  if (rate === 0) return 'Included';
  return `$${(rate / 1000).toFixed(4)}/req`;
};

/**
 * Check if the plan has Function Consciousness (basic insights, daily digest)
 * Available on Professional and above
 */
export const hasConsciousness = (plan?: string): boolean => {
  return hasMinPlan(plan as PlanTier, 'professional');
};

/**
 * Check if the plan has Advanced Consciousness (real-time, predictive, auto-fix)
 * Available on Enterprise and above
 */
export const hasAdvancedConsciousness = (plan?: string): boolean => {
  return hasMinPlan(plan as PlanTier, 'enterprise');
};

/**
 * Check if the plan has Autonomous Consciousness (self-healing, unlimited lookback)
 * Available on Agent Enterprise only
 */
export const hasAutonomousConsciousness = (plan?: string): boolean => {
  return plan?.toLowerCase() === 'agent_enterprise';
};

/**
 * Get consciousness lookback days for the plan
 */
export const getConsciousnessLookbackDays = (plan?: string): number => {
  if (hasAutonomousConsciousness(plan)) return -1; // Unlimited
  if (hasAdvancedConsciousness(plan)) return 30;
  if (hasConsciousness(plan)) return 7;
  return 0;
};
