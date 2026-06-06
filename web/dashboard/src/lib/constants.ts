export const APP_NAME = 'FunctionFly';
export const APP_TAGLINE = 'The Trust Layer for AI Agents';

/**
 * Allowed docs site origins — prevents open-redirect if VITE_DOCS_SITE_URL is misconfigured.
 * In production only docs.functionfly.com is allowed.
 * In development localhost ports 4322/4323 are allowed.
 */
const ALLOWED_DOCS_ORIGINS = new Set([
  'https://docs.functionfly.com',
  'http://localhost:4322',
  'http://localhost:4323',
]);

function isAllowedDocsOrigin(origin: string): boolean {
  try {
    const url = new URL(origin);
    // In production (or when no override is set), only allow known safe origins
    if (import.meta.env.PROD) {
      return ALLOWED_DOCS_ORIGINS.has(origin);
    }
    // In dev, allow localhost on the expected ports
    return url.hostname === 'localhost' && ALLOWED_DOCS_ORIGINS.has(origin);
  } catch {
    return false;
  }
}

/**
 * Origin of the standalone Astro docs site (web/docs), e.g. https://docs.functionfly.com.
 * Dashboard /docs/* redirects here. Dev default: http://localhost:4322 (see web/docs astro.config).
 * Override with VITE_DOCS_SITE_URL.
 */
export function getPublicDocsSiteOrigin(): string {
  const raw = (import.meta.env.VITE_DOCS_SITE_URL ?? '').trim().replace(/\/$/, '');
  if (raw && isAllowedDocsOrigin(raw)) return raw;
  if (import.meta.env.PROD) return 'https://docs.functionfly.com';
  return 'http://localhost:4322';
}

/** @deprecated Prefer getPublicDocsSiteOrigin(); kept as alias for href bases (no path). */
export const DOCS_SITE_URL = getPublicDocsSiteOrigin();

export const ROUTES = {
  HOME: '/',
  LAUNCH: '/launch',
  COMING_SOON: '/coming-soon',
  PRICING: '/pricing',
  LOGIN: '/login',
  SIGNUP: '/signup',
  DASHBOARD: '/dashboard',
  /** Marketplace / discovery home (sidebar: Discover). */
  DISCOVER: '/functions/discovery',
  /** Metrics / activity home (sidebar: Overview). */
  OVERVIEW: '/overview',
  FUNCTIONS: '/functions',
  REGISTRY: '/registry',
  PROVIDERS: '/providers',
  ANALYTICS: '/analytics',
  USAGE: '/usage',
  STATE_FABRIC: '/state-fabric',
  SECRETS: '/secrets',
  API_KEYS: '/api-keys',
  CERTIFICATION: '/certification',
  CREDENTIALS: '/credentials',
  SETTINGS: '/settings',
  BILLING: '/settings', // Billing tab lives on Settings page
  TEAMS: '/teams',
  TEAM_MEMORY: '/teams/:teamId/memory',
  TEAM_DECISIONS: '/teams/:teamId/decisions',
  DECISIONS: '/decisions',
  APPS: '/apps',
  APP_DETAIL: '/apps/:slug',
  FUNCTION_DETAIL: '/functions/:id',
  // Agent routes
  AGENTS: (username: string) => `/u/${username}/agents`,
  AGENT_LIST: '/agents',
  AGENT_NEW: '/agents/new',
  AGENT_DETAIL: '/agents/:id',
  AGENT_EDIT: '/agents/:id/edit',
  AGENT_WALLET: '/agents/:id/wallet',
  AGENT_ANALYTICS: '/agents/:id/analytics',
  SDK_INTEGRATIONS: '/sdk-integrations',
  MARKETPLACE: '/marketplace',
  MARKETPLACE_AGENTS: '/marketplace',
  MARKETPLACE_FUNCTIONS: '/functions/discovery',
  EVOLUTION: '/evolution',
  STUDIO: '/studio',
  // FRG - Function Runtime Graph
  FRG: '/frg',
  FRG_NEW: '/frg/new',
  FRG_EDIT: '/frg/:id',
  // Enterprise routes
  ENTERPRISE: '/enterprise',
  ENTERPRISE_SLA: '/enterprise/sla',
  ENTERPRISE_AUDIT: '/enterprise/audit',
  ENTERPRISE_SECURITY: '/enterprise/security',
  ENTERPRISE_SUPPORT: '/enterprise/support',
  ENTERPRISE_COMPLIANCE: '/enterprise/compliance',
  PAYOUTS: '/payouts',
  PLATFORM_WALLET: '/platform-wallet',
  AGENT_MEMORIES: '/agent-memories',
  CONVERSATIONS: '/conversations',
  STATE: '/state',
  // Time Machine routes
  TIME_MACHINE: '/time-machine',
  TIME_MACHINE_NEW: '/time-machine/new',
  TIME_MACHINE_DETAIL: '/time-machine/:id',
  TIME_MACHINE_SCHEDULES: '/time-machine/schedules',
  TIME_MACHINE_AUDIT: '/time-machine/audit',
} as const;

/**
 * Standalone admin app (web/admin-dashboard): local dev :3002, production admin.functionfly.com.
 * Override per environment with VITE_ADMIN_DASHBOARD_URL (e.g. staging).
 */
export const ADMIN_DASHBOARD_URL =
  (import.meta.env.VITE_ADMIN_DASHBOARD_URL ?? '').trim() ||
  (import.meta.env.DEV ? 'http://localhost:3002' : 'https://admin.functionfly.com');

/**
 * Origin for user-visible app URLs (e.g. https://functionfly.com/apps/my-app).
 * Override with VITE_PUBLIC_SITE_URL for staging (no trailing slash).
 */
export const PUBLIC_SITE_ORIGIN =
  (import.meta.env.VITE_PUBLIC_SITE_URL ?? '').trim().replace(/\/$/, '') ||
  'https://functionfly.com';

/**
 * URL of the marketing site (functionfly.com). The dashboard app redirects "/" here.
 * Same as PUBLIC_SITE_ORIGIN unless overridden for multi-origin setups.
 */
export const MARKETING_SITE_URL = PUBLIC_SITE_ORIGIN;

/**
 * Blog site origin (standalone Astro blog at web/blog).
 * Production: https://blog.functionfly.com, local dev: http://localhost:4327.
 * Override with VITE_BLOG_SITE_URL.
 */
export function getBlogSiteOrigin(): string {
  const env = (import.meta.env.VITE_BLOG_SITE_URL ?? '').trim().replace(/\/$/, '');
  if (env) return env;
  if (import.meta.env.PROD) return 'https://blog.functionfly.com';
  return 'http://localhost:4327';
}

/**
 * URL of the blog site (blog.functionfly.com). The dashboard app redirects "/blog" here.
 */
export const BLOG_SITE_URL = getBlogSiteOrigin();

/**
 * Logged-out "/" and nav "Home" go to the Astro marketing site: production uses MARKETING_SITE_URL;
 * local dev defaults to http://localhost:4321 (web/site). Override with VITE_MARKETING_DEV_URL.
 */
export function getMarketingRedirectOrigin(): string {
  if (import.meta.env.PROD) {
    return MARKETING_SITE_URL;
  }
  const raw = (import.meta.env.VITE_MARKETING_DEV_URL ?? '').trim().replace(/\/$/, '');
  return raw || 'http://localhost:4321';
}

/** Absolute URL for a path on the Astro marketing site (e.g. `/privacy`, `/terms`). */
export function getMarketingPageUrl(path: string): string {
  const origin = getMarketingRedirectOrigin();
  const p = path.startsWith('/') ? path : `/${path}`;
  return `${origin}${p}`;
}

export function publicAppUrl(slug: string): string {
  return `${PUBLIC_SITE_ORIGIN}/apps/${slug}`;
}

/**
 * All sidebar main nav paths (for recent-tab tracking).
 * Sorted by path length descending so longer paths match first.
 */
export const MAIN_NAV_PATHS: string[] = [
  ROUTES.STATE_FABRIC,
  ROUTES.SDK_INTEGRATIONS,
  ROUTES.MARKETPLACE_AGENTS,
  ROUTES.FUNCTION_DETAIL,
  ROUTES.APP_DETAIL,
  ROUTES.DISCOVER,
  ROUTES.DASHBOARD,
  ROUTES.OVERVIEW,
  ROUTES.FUNCTIONS,
  ROUTES.APPS,
  ROUTES.REGISTRY,
  ROUTES.PROVIDERS,
  ROUTES.TEAMS,
  ROUTES.AGENT_LIST,
  ROUTES.ANALYTICS,
  ROUTES.USAGE,
  ROUTES.SECRETS,
  ROUTES.API_KEYS,
  ROUTES.SETTINGS,
].sort((a, b) => b.length - a.length);

/** Resolve current pathname to a canonical sidebar path, or null if not a main nav route. */
export function getCanonicalNavPath(pathname: string): string | null {
  if (
    pathname === ROUTES.MARKETPLACE_FUNCTIONS ||
    pathname.startsWith(ROUTES.MARKETPLACE_FUNCTIONS + '/')
  ) {
    return ROUTES.DASHBOARD;
  }
  for (const p of MAIN_NAV_PATHS) {
    if (
      pathname === p ||
      (p !== ROUTES.DASHBOARD && p !== ROUTES.OVERVIEW && pathname.startsWith(p + '/'))
    ) {
      return p;
    }
  }
  return null;
}

// ============================================================================
// Dynamic Route Builder Functions
// ============================================================================

/**
 * Builder functions for dynamic routes
 * Use these to generate URLs with parameters safely
 */
export const ROUTE_BUILDERS = {
  // Function routes
  function: (author: string, name: string) => `/fx/${author}/${name}`,
  functionWithVersion: (author: string, name: string, version: string) =>
    `/fx/${author}/${name}@${version}`,

  // Playground/Run routes
  playground: (author: string, name: string) => `/run/${author}/${name}`,
  playgroundWithVersion: (author: string, name: string, version: string) =>
    `/run/${author}/${name}@${version}`,

  // Execution replay
  replay: (execId: string) => `/replay/${execId}`,

  // User profile
  userProfile: (username: string) => `/u/${username}`,

  // Blog posts (public blog on standalone Astro site web/blog, not the app)
  blogPost: (slug: string) => `${getBlogSiteOrigin()}/${slug}`,

  // Documentation (standalone Astro app web/docs)
  docs: (slug?: string) =>
    slug ? `${getPublicDocsSiteOrigin()}/docs/${slug}` : `${getPublicDocsSiteOrigin()}/`,
  docsApi: (endpoint?: string) =>
    endpoint ? `${getPublicDocsSiteOrigin()}/api/${endpoint}` : `${getPublicDocsSiteOrigin()}/api`,

  // Registry search
  registrySearch: (query: string, page = 1) =>
    `/registry?q=${encodeURIComponent(query)}&page=${page}`,

  // Function details/edit
  functionEdit: (author: string, name: string) => `/functions/${author}/${name}`,
  functionSettings: (author: string, name: string) => `/functions/${author}/${name}/settings`,
  functionLogs: (author: string, name: string) => `/functions/${author}/${name}/logs`,

  // User dashboard sections
  userFunctions: (username: string) => `/dashboard/${username}/functions`,
  userSettings: (username: string) => `/u/${username}/settings`,

  // Agent dynamic routes
  agent: (slug: string) => `/agents/${slug}`,
} as const;

// Type for route builder function
export type RouteBuilder = typeof ROUTE_BUILDERS;

export const PROVIDERS = {
  CLOUDFLARE: {
    id: 'workers',
    name: 'Cloudflare Workers',
    color: '#f48120',
    icon: 'Cloud',
    regions: ['iad', 'lhr', 'sin', 'syd', 'fra', 'hkg', 'tyo'],
  },
  VERCEL: {
    id: 'vercel',
    name: 'Vercel',
    color: '#000000',
    icon: 'Triangle',
    regions: ['iad1', 'sfo1', 'lhr1', 'fra1'],
  },
  FLY: {
    id: 'fly',
    name: 'Fly.io',
    color: '#7b68ee',
    icon: 'Plane',
    regions: ['iad', 'lax', 'ord', 'lhr', 'fra', 'sin', 'syd', 'nrt'],
  },
  DENO: {
    id: 'deno',
    name: 'Deno Deploy',
    color: '#000000',
    icon: 'Dino',
    regions: ['us-east4', 'us-west2', 'europe-west1', 'asia-northeast1'],
  },
  FUNCTIONFLY_EDGE: {
    id: 'functionfly-edge',
    name: 'FunctionFly Edge',
    color: '#6366f1',
    icon: 'Zap',
    regions: ['us-east-1', 'us-west-2', 'eu-central-1'],
    description:
      "Host your edge functions on FunctionFly's infrastructure - no deployment required",
    isManaged: true,
  },
  AWS_LAMBDA: {
    id: 'aws-lambda',
    name: 'AWS Lambda',
    color: '#FF9900',
    icon: 'Aws',
    regions: ['us-east-1', 'us-east-2', 'us-west-2', 'eu-west-1', 'eu-central-1', 'ap-southeast-1', 'ap-northeast-1'],
    description:
      'Deploy functions to AWS Lambda with full lifecycle management, auto-scaling, and pay-per-use pricing',
  },
} as const;

/** Vendor cloud dashboards opened from Providers → Configure (connected). */
export const PROVIDER_EXTERNAL_DASHBOARD_URL = {
  workers: 'https://dash.cloudflare.com/',
  vercel: 'https://vercel.com/dashboard',
  fly: 'https://fly.io/dashboard/',
  deno: 'https://dash.deno.com/',
  'aws-lambda': 'https://console.aws.amazon.com/lambda/',
} as const;

export const PLANS = {
  FREE: {
    id: 'free',
    name: 'Free',
    price: 0,
    priceCents: 0,
    priceAnnualCents: 0,
    priceId: '',
    priceIdAnnual: '',
    description: 'Perfect for getting started with FunctionFly',
    features: ['3 functions', '2 providers', '500 requests/month', 'Community support', '24h Time Machine replay'],
    overageRate: null, // Hard stop at limit
    annualDiscount: 0,
    comingSoon: false,
    limits: {
      functions: 3,
      providers: 2,
      requests: 500,
      customDomains: 0,
      stateFabrics: 0,
      agents: 3,
      apps: 0,
      secrets: 0,
      tokensPerSecret: 0,
      apiKeyBudgets: false,
      perKeyCostAttribution: false,
      replayWindowHours: 24,
      maxExecutionsPerReplay: 100,
      maxConcurrentReplays: 1,
    },
  },
  STARTER: {
    id: 'starter',
    name: 'Starter',
    price: 24,
    priceCents: 2400,
    priceAnnualCents: 24000,
    priceId: import.meta.env.VITE_STRIPE_PRICE_STARTER || 'price_starter_placeholder',
    priceIdAnnual: import.meta.env.VITE_STRIPE_PRICE_STARTER_ANNUAL || 'price_starter_annual_placeholder',
    description: 'For side projects and MVPs',
    features: [
      '5 functions',
      '3 providers',
      '100K AI calls/month',
      '$0.15 per 1K overage calls',
      '1 custom domain',
      'Email support',
      'Basic analytics',
      '10 agents included',
      '100 state writes/hour',
      '72h Time Machine replay',
    ],
    overageRate: 15, // $0.15 per 1000 calls
    annualDiscount: 0.17, // 17% off (2 months free)
    comingSoon: false,
    limits: {
      functions: 5,
      providers: 3,
      requests: 1000000,
      customDomains: 1,
      stateFabrics: 1,
      agents: 10,
      apps: 3,
      secrets: 10,
      tokensPerSecret: 5,
      apiKeyBudgets: false,
      perKeyCostAttribution: false,
      aiCallsPerMonth: 100000,
      agentConcurrency: 10,
      agentCallsPerMinute: 100,
      replayWindowHours: 72,
      maxExecutionsPerReplay: 1000,
      maxConcurrentReplays: 1,
    },
  },
  PROFESSIONAL: {
    id: 'professional',
    name: 'Professional',
    price: 79,
    priceCents: 7900,
    priceAnnualCents: 79000,
    priceId: import.meta.env.VITE_STRIPE_PRICE_PROFESSIONAL || 'price_professional_placeholder',
    priceIdAnnual: import.meta.env.VITE_STRIPE_PRICE_PROFESSIONAL_ANNUAL || 'price_professional_annual_placeholder',
    description: 'For growing businesses and SaaS applications',
    features: [
      '25 functions',
      '5 providers',
      '1M AI calls/month',
      '$0.08 per 1K overage calls',
      '5 custom domains',
      '99.9% SLA',
      'Priority support',
      'Advanced analytics',
      'Team collaboration',
      'Per-API-key cost tracking',
      'Custom routing rules',
      '100 agents included',
      '10K state writes/hour',
      '30-day Time Machine replay + reconciliation',
    ],
    overageRate: 8, // $0.08 per 1000 calls
    annualDiscount: 0.17, // 17% off
    comingSoon: false,
    limits: {
      functions: 25,
      providers: 5,
      requests: 10000000,
      customDomains: 5,
      sla: '99.9%',
      stateFabrics: 5,
      agents: 100,
      apps: 10,
      secrets: 50,
      tokensPerSecret: 20,
      apiKeyBudgets: false,
      perKeyCostAttribution: true, // Can track costs per API key
      aiCallsPerMonth: 1000000,
      agentConcurrency: 100,
      agentCallsPerMinute: 500,
      replayWindowHours: 720,
      maxExecutionsPerReplay: 10000,
      maxConcurrentReplays: 3,
    },
  },
  ENTERPRISE: {
    id: 'enterprise',
    name: 'Enterprise',
    price: 299,
    priceCents: 29900,
    priceAnnualCents: 299000,
    priceId: import.meta.env.VITE_STRIPE_PRICE_ENTERPRISE || 'price_enterprise_placeholder',
    priceIdAnnual: import.meta.env.VITE_STRIPE_PRICE_ENTERPRISE_ANNUAL || 'price_enterprise_annual_placeholder',
    description: 'For large-scale applications and enterprises',
    features: [
      'Unlimited functions',
      'All providers',
      '5M AI calls/month',
      '$0.05 per 1K overage calls',
      'Unlimited custom domains',
      '99.99% SLA',
      'Dedicated support',
      'Custom integrations',
      'Per-API-key budgets & alerts',
      'Volume discounts',
      'On-premise deployment',
      '500 agents included',
      '50K state writes/hour',
      '90-day Time Machine + live reconciliation + audit certificates',
    ],
    overageRate: 5, // $0.05 per 1000 calls
    annualDiscount: 0.17, // 17% off
    comingSoon: false,
    limits: {
      functions: Infinity,
      providers: Infinity,
      requests: Infinity,
      customDomains: Infinity,
      sla: '99.99%',
      stateFabrics: Infinity,
      agents: 500,
      apps: -1, // Unlimited
      secrets: 10000,
      tokensPerSecret: 100,
      apiKeyBudgets: true, // Full API key budget controls
      perKeyCostAttribution: true,
      highValueKeySeparation: true,
      aiCallsPerMonth: 5000000,
      agentConcurrency: 500,
      agentCallsPerMinute: 2000,
      replayWindowHours: 2160,
      maxExecutionsPerReplay: 100000,
      maxConcurrentReplays: 10,
    },
  },
} as const;

// ============================================================================
// Agent Enterprise Plan - Ultimate tier with unlimited AI
// ============================================================================

export const AGENT_ENTERPRISE = {
  id: 'agent_enterprise',
  name: 'Agent Enterprise',
  price: 499,
  priceCents: 49900,
  priceAnnualCents: 499000,
  priceId: import.meta.env.VITE_STRIPE_PRICE_AGENT_ENTERPRISE || 'price_agent_enterprise_placeholder',
  priceIdAnnual: import.meta.env.VITE_STRIPE_PRICE_AGENT_ENTERPRISE_ANNUAL || 'price_agent_enterprise_annual_placeholder',
  description: 'Unlimited AI agent scale for enterprise',
  features: [
    'Unlimited AI calls/month',
    'Unlimited agents',
    'Unlimited concurrency',
    'Unlimited state writes',
    'Unlimited memory storage',
    '1-year log retention',
    'Dedicated infrastructure',
    'Custom SLA',
    '24/7 support',
    'Volume discounts',
    'On-premise deployment',
    'Unlimited Time Machine + incident insurance',
  ],
  overageRate: 0, // No overage (unlimited)
  annualDiscount: 0.17, // 17% off
  limits: {
    functions: Infinity,
    providers: Infinity,
    requests: Infinity,
    customDomains: Infinity,
    sla: '99.99%',
    stateFabrics: Infinity,
    agents: -1, // Unlimited
    apps: -1, // Unlimited
    secrets: Infinity,
    tokensPerSecret: Infinity,
    apiKeyBudgets: true,
    perKeyCostAttribution: true,
    highValueKeySeparation: true,
    aiCallsPerMonth: -1, // Unlimited
    agentConcurrency: -1, // Unlimited
    agentCallsPerMinute: -1, // Unlimited
    replayWindowHours: Infinity,
    maxExecutionsPerReplay: Infinity,
    maxConcurrentReplays: Infinity,
  },
} as const;

// ============================================================================
// Agent Execution Plans (AEP) - LEGACY, now bundled into main PLANS
// Kept for backward compatibility - agents now get capabilities from main plan
// These are now aliased to equivalent main plan limits
// ============================================================================

export const AGENT_PLANS = {
  STARTER: {
    id: 'agent_starter',
    name: 'Agent Starter',
    price: 24,
    priceCents: 2400,
    priceAnnualCents: 24000,
    priceId: import.meta.env.VITE_STRIPE_PRICE_AGENT_STARTER || 'price_agent_starter_placeholder',
    description: 'For small AI agent projects and prototyping',
    features: [
      '100K tool calls/month',
      '10 concurrent agents',
      '100 calls/second',
      '1K state writes/hour',
      '10GB memory storage',
      '30-day log retention',
      'Per-agent cost tracking',
      'Email support',
    ],
    limits: {
      callsPerMonth: 100000,
      concurrency: 10,
      burst: 100,
      dailySpendCap: 5,
      stateWritesPerHour: 1000,
      memoryGB: 10,
      logRetentionDays: 30,
    },
    // DEPRECATED: Use PLANS.STARTER instead - same pricing and limits
    deprecated: true,
    aliasFor: 'starter',
  },
  SCALE: {
    id: 'agent_scale',
    name: 'Agent Scale',
    price: 79,
    priceCents: 7900,
    priceAnnualCents: 79000,
    priceId: import.meta.env.VITE_STRIPE_PRICE_AGENT_SCALE || 'price_agent_scale_placeholder',
    description: 'For growing AI agent applications',
    features: [
      '1M tool calls/month',
      '100 concurrent agents',
      '500 calls/second',
      '10K state writes/hour',
      '100GB memory storage',
      '90-day log retention',
      'Per-agent cost tracking',
      'Budget enforcement',
      'Priority support',
    ],
    limits: {
      callsPerMonth: 1000000,
      concurrency: 100,
      burst: 500,
      dailySpendCap: 30,
      stateWritesPerHour: 10000,
      memoryGB: 100,
      logRetentionDays: 90,
    },
    // DEPRECATED: Use PLANS.PROFESSIONAL instead - same pricing and limits
    deprecated: true,
    aliasFor: 'professional',
  },
  PRO: {
    id: 'agent_pro',
    name: 'Agent Pro',
    price: 299,
    priceCents: 29900,
    priceAnnualCents: 299000,
    priceId: import.meta.env.VITE_STRIPE_PRICE_AGENT_PRO || 'price_agent_pro_placeholder',
    description: 'For production AI agent systems',
    features: [
      '5M tool calls/month',
      '500 concurrent agents',
      '2000 calls/second',
      '50K state writes/hour',
      '500GB memory storage',
      '1-year log retention',
      'Per-agent cost tracking',
      'Budget enforcement',
      'Multi-agent coordination',
      'Dedicated support',
    ],
    limits: {
      callsPerMonth: 5000000,
      concurrency: 500,
      burst: 2000,
      dailySpendCap: 100,
      stateWritesPerHour: 50000,
      memoryGB: 500,
      logRetentionDays: 365,
    },
    // DEPRECATED: Use PLANS.ENTERPRISE instead - same pricing and limits
    deprecated: true,
    aliasFor: 'enterprise',
  },
  ENTERPRISE: {
    id: 'agent_enterprise',
    name: 'Agent Enterprise',
    price: 499,
    priceCents: 49900,
    priceAnnualCents: 499000,
    priceId: import.meta.env.VITE_STRIPE_PRICE_AGENT_ENTERPRISE || 'price_agent_enterprise_placeholder',
    description: 'For unlimited AI agent scale',
    features: [
      'Unlimited tool calls',
      'Unlimited concurrent agents',
      'Unlimited burst',
      'Unlimited state writes',
      'Unlimited memory storage',
      '1-year log retention',
      'Per-agent cost attribution',
      'Custom budget controls',
      'Multi-region deployment',
      'Dedicated infrastructure',
      '24/7 phone support',
    ],
    limits: {
      callsPerMonth: -1,
      concurrency: -1,
      burst: -1,
      dailySpendCap: -1,
      stateWritesPerHour: -1,
      memoryGB: -1,
      logRetentionDays: -1,
    },
    // ACTIVE: This is the max tier for unlimited agents
    deprecated: false,
    aliasFor: null,
  },
} as const;

// ============================================================================
// State Fabric Plans - pricing for stateful serverless capabilities
// ============================================================================

export const STATE_FABRIC_PLANS = {
  SANDBOX: {
    id: 'sf_sandbox',
    name: 'Sandbox',
    price: 0,
    priceCents: 0,
    priceId: '',
    description: 'Great for experimentation & onboarding.',
  },
  STARTER: {
    id: 'sf_starter',
    name: 'Starter',
    price: 19,
    priceCents: 1900,
    priceId: import.meta.env.VITE_STRIPE_PRICE_SF_STARTER || 'price_sf_starter_placeholder',
    description: 'Internal dev & small side projects.',
  },
  PRO: {
    id: 'sf_pro',
    name: 'Pro',
    price: 99,
    priceCents: 9900,
    priceId: import.meta.env.VITE_STRIPE_PRICE_SF_PRO || 'price_sf_pro_placeholder',
    description: 'Production apps & team projects.',
  },
  BUSINESS: {
    id: 'sf_business',
    name: 'Business',
    price: 499,
    priceCents: 49900,
    priceId: import.meta.env.VITE_STRIPE_PRICE_SF_BUSINESS || 'price_sf_business_placeholder',
    description: 'Business-critical systems.',
  },
  ENTERPRISE: {
    id: 'sf_enterprise',
    name: 'Enterprise',
    price: 'Custom',
    priceCents: 0,
    priceId: '',
    description: 'Large enterprises & regulated industries.',
  },
} as const;

export const STATUS_COLORS = {
  online: {
    bg: 'bg-emerald-500',
    text: 'text-emerald-500',
    border: 'border-emerald-500',
    label: 'Online',
  },
  offline: {
    bg: 'bg-red-500',
    text: 'text-red-500',
    border: 'border-red-500',
    label: 'Offline',
  },
  degraded: {
    bg: 'bg-amber-500',
    text: 'text-amber-500',
    border: 'border-amber-500',
    label: 'Degraded',
  },
  pending: {
    bg: 'bg-gray-500',
    text: 'text-gray-500',
    border: 'border-gray-500',
    label: 'Pending',
  },
} as const;

/**
 * Launch-only mode: every route shows the coming-soon page.
 * - Production (`vite build`): true when VITE_COMING_SOON_ONLY=true (pre-launch deploys).
 * - Dev (`vite`): ignores VITE_COMING_SOON_ONLY unless VITE_COMING_SOON_IN_DEV=true, so a copied
 *   production .env does not hide /login on localhost. To preview launch mode locally, set both to "true".
 */
export const COMING_SOON_ONLY =
  import.meta.env.VITE_COMING_SOON_ONLY === 'true' &&
  (import.meta.env.PROD || import.meta.env.VITE_COMING_SOON_IN_DEV === 'true');

/**
 * Canonical API base URL (no trailing slash). Use for all API and WebSocket calls.
 * - Production: set VITE_API_URL (e.g. https://api.functionfly.com).
 * - Local dev: set VITE_API_URL=http://localhost:8080 to hit the API directly (recommended), or VITE_API_URL=/api to use the Vite proxy (backend must be running on 8080). If you see 404 on /api/v1/api-keys, use http://localhost:8080 or start the backend.
 * - Flywheel production: http://localhost:3000/flywheel
 */
export function getApiBaseUrl(): string {
  const env = (import.meta.env.VITE_API_URL ?? '').trim();
  if (import.meta.env.DEV) {
    // In dev, use /api so requests go through Vite proxy to the backend (avoids "Cannot POST /auth/login" when the request would otherwise hit the dev server).
    if (!env) return '/api';
    const base = env.replace(/\/$/, '');
    try {
      const origin = typeof window !== 'undefined' ? window.location.origin : '';
      if (origin && (base === origin || base === '')) return '/api';
    } catch {
      /* ignore */
    }
    return base;
  }
  // Production: Use configured URL or fail - no localhost fallback
  if (env) return env.replace(/\/$/, '');
  throw new Error('VITE_API_URL environment variable is required in production');
}

export const API_BASE_URL = getApiBaseUrl();

/** Path for the AI/LLM discovery manifest (GET). Full URL: ${API_BASE_URL}/.well-known/functionfly.json */
export const WELL_KNOWN_DISCOVERY_PATH = '/.well-known/functionfly.json';

/**
 * AI service base URL (FlyMind / ai-service). When set, dashboard AI features and other
 * AI features call the completion API. Leave unset for simulated responses or when
 * the orchestrator proxies AI requests.
 */
export function getAiServiceBaseUrl(): string {
  const env = (import.meta.env.VITE_AI_SERVICE_URL ?? '').trim();
  if (env) return env.replace(/\/$/, '');
  return '';
}

export const AI_SERVICE_BASE_URL = getAiServiceBaseUrl();

/**
 * Validate that Stripe price IDs are set in production.
 * Throws an error if any price ID is a placeholder value or uses invalid ID format.
 */
const STRIPE_PRICE_ID_PATTERN = /^(price_[\w]+_placeholder|)$/;
const STRIPE_PRICE_ID_VALID_PATTERN = /^price_[\w]+$/;

function validateStripePriceIds() {
  if (!import.meta.env.PROD) return;

  const priceIds = {
    VITE_STRIPE_PRICE_STARTER: import.meta.env.VITE_STRIPE_PRICE_STARTER,
    VITE_STRIPE_PRICE_PROFESSIONAL: import.meta.env.VITE_STRIPE_PRICE_PROFESSIONAL,
    VITE_STRIPE_PRICE_AGENT_STARTER: import.meta.env.VITE_STRIPE_PRICE_AGENT_STARTER,
    VITE_STRIPE_PRICE_AGENT_SCALE: import.meta.env.VITE_STRIPE_PRICE_AGENT_SCALE,
    VITE_STRIPE_PRICE_AGENT_PRO: import.meta.env.VITE_STRIPE_PRICE_AGENT_PRO,
    VITE_STRIPE_PRICE_SF_STARTER: import.meta.env.VITE_STRIPE_PRICE_SF_STARTER,
    VITE_STRIPE_PRICE_SF_PRO: import.meta.env.VITE_STRIPE_PRICE_SF_PRO,
    VITE_STRIPE_PRICE_SF_BUSINESS: import.meta.env.VITE_STRIPE_PRICE_SF_BUSINESS,
  };

  const missing: string[] = [];
  const placeholders: string[] = [];
  const productIds: string[] = [];
  const invalidIds: string[] = [];

  for (const [envVar, value] of Object.entries(priceIds)) {
    if (!value) {
      missing.push(envVar);
    } else if (STRIPE_PRICE_ID_PATTERN.test(value)) {
      placeholders.push(envVar);
    } else if (value.startsWith('prod_')) {
      productIds.push(`${envVar}=${value}`);
    } else if (!STRIPE_PRICE_ID_VALID_PATTERN.test(value)) {
      invalidIds.push(`${envVar}=${value}`);
    }
  }

  if (
    missing.length > 0 ||
    placeholders.length > 0 ||
    productIds.length > 0 ||
    invalidIds.length > 0
  ) {
    const errors: string[] = [];
    if (missing.length > 0) {
      errors.push(`Missing Stripe price IDs: ${missing.join(', ')}`);
    }
    if (placeholders.length > 0) {
      errors.push(`Placeholder Stripe price IDs detected: ${placeholders.join(', ')}`);
    }
    if (productIds.length > 0) {
      errors.push(`Product IDs (prod_*) used instead of price IDs: ${productIds.join(', ')}`);
    }
    if (invalidIds.length > 0) {
      errors.push(`Invalid price ID format: ${invalidIds.join(', ')}`);
    }
    throw new Error(
      `Production build failed: ${errors.join('. ')}. Set actual Stripe price IDs (price_*) in your production environment.`
    );
  }
}

// Run validation at module load time
validateStripePriceIds()