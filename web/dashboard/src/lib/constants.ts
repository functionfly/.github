export const APP_NAME = "FunctionFly";
export const APP_TAGLINE = "Serverless Functions & AI Agent Infrastructure";

/** Base URL for the docs site (web/docs app). */
export const DOCS_SITE_URL = "/docs";

export const ROUTES = {
  HOME: "/",
  LAUNCH: "/launch",
  COMING_SOON: "/coming-soon",
  PRICING: "/pricing",
  LOGIN: "/login",
  SIGNUP: "/signup",
  DASHBOARD: "/dashboard",
  FUNCTIONS: "/functions",
  REGISTRY: "/registry",
  PROVIDERS: "/providers",
  ANALYTICS: "/analytics",
  USAGE: "/usage",
  STATE_FABRIC: "/state-fabric",
  SECRETS: "/secrets",
  API_KEYS: "/api-keys",
  SETTINGS: "/settings",
  BILLING: "/settings", // Billing tab lives on Settings page
  TEAMS: "/teams",
  APPS: "/apps",
  APP_DETAIL: "/apps/:appId",
  FUNCTION_DETAIL: "/functions/:id",
  // Agent routes
  AGENTS: "/agents",
  AGENT_DETAIL: "/agents/:agentId",
  MARKETPLACE_AGENTS: "/marketplace/agents",
  MARKETPLACE_FUNCTIONS: "/marketplace/functions",
  EVOLUTION: "/evolution",
  // Enterprise routes
  ENTERPRISE: "/enterprise",
  ENTERPRISE_SLA: "/enterprise/sla",
  ENTERPRISE_AUDIT: "/enterprise/audit",
  ENTERPRISE_SECURITY: "/enterprise/security",
  ENTERPRISE_SUPPORT: "/enterprise/support",
  ENTERPRISE_COMPLIANCE: "/enterprise/compliance",
} as const;

/**
 * All sidebar main nav paths (for recent-tab tracking).
 * Sorted by path length descending so longer paths match first.
 */
export const MAIN_NAV_PATHS: string[] = [
  ROUTES.STATE_FABRIC,
  ROUTES.MARKETPLACE_AGENTS,
  ROUTES.MARKETPLACE_FUNCTIONS,
  ROUTES.FUNCTION_DETAIL,
  ROUTES.APP_DETAIL,
  ROUTES.DASHBOARD,
  ROUTES.FUNCTIONS,
  ROUTES.APPS,
  ROUTES.REGISTRY,
  ROUTES.PROVIDERS,
  ROUTES.TEAMS,
  ROUTES.AGENTS,
  ROUTES.ANALYTICS,
  ROUTES.USAGE,
  ROUTES.SECRETS,
  ROUTES.API_KEYS,
  ROUTES.SETTINGS,
].sort((a, b) => b.length - a.length);

/** Resolve current pathname to a canonical sidebar path, or null if not a main nav route. */
export function getCanonicalNavPath(pathname: string): string | null {
  for (const p of MAIN_NAV_PATHS) {
    if (pathname === p || (p !== ROUTES.DASHBOARD && pathname.startsWith(p + "/")))
      return p;
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

  // Blog posts
  blogPost: (slug: string) => `/blog/${slug}`,

  // Documentation (main docs site = web/docs app)
  docs: (slug?: string) => (slug ? `${DOCS_SITE_URL}/${slug}` : DOCS_SITE_URL),
  docsApi: (endpoint?: string) =>
    endpoint ? `${DOCS_SITE_URL}/api/${endpoint}` : `${DOCS_SITE_URL}/api`,

  // Registry search
  registrySearch: (query: string, page = 1) =>
    `/registry?q=${encodeURIComponent(query)}&page=${page}`,

  // Function details/edit
  functionEdit: (author: string, name: string) =>
    `/functions/${author}/${name}`,
  functionSettings: (author: string, name: string) =>
    `/functions/${author}/${name}/settings`,
  functionLogs: (author: string, name: string) =>
    `/functions/${author}/${name}/logs`,

  // User dashboard sections
  userFunctions: (username: string) => `/dashboard/${username}/functions`,
  userSettings: (username: string) => `/u/${username}/settings`,

  // Agent dynamic routes
  agent: (agentId: string) => `/agents/${agentId}`,
} as const;

// Type for route builder function
export type RouteBuilder = typeof ROUTE_BUILDERS;

export const PROVIDERS = {
  CLOUDFLARE: {
    id: "workers",
    name: "Cloudflare Workers",
    color: "#f48120",
    icon: "Cloud",
    regions: ["iad", "lhr", "sin", "syd", "fra", "hkg", "tyo"],
  },
  VERCEL: {
    id: "vercel",
    name: "Vercel",
    color: "#000000",
    icon: "Triangle",
    regions: ["iad1", "sfo1", "lhr1", "fra1"],
  },
  FLY: {
    id: "fly",
    name: "Fly.io",
    color: "#7b68ee",
    icon: "Plane",
    regions: ["iad", "lax", "ord", "lhr", "fra", "sin", "syd", "nrt"],
  },
  DENO: {
    id: "deno",
    name: "Deno Deploy",
    color: "#000000",
    icon: "Dino",
    regions: ["us-east4", "us-west2", "europe-west1", "asia-northeast1"],
  },
  FUNCTIONFLY_EDGE: {
    id: "functionfly-edge",
    name: "FunctionFly Edge",
    color: "#6366f1",
    icon: "Zap",
    regions: ["auto", "us-east-1", "us-west-1", "eu-west-1", "eu-central-1", "ap-southeast-1", "ap-northeast-1", "ap-south-1"],
    description: "Host your edge functions on FunctionFly's infrastructure - no deployment required",
    isManaged: true,
  },
} as const;

export const PLANS = {
  FREE: {
    id: "free",
    name: "Free",
    price: 0,
    priceCents: 0,
    priceId: "",
    description: "Perfect for getting started with FunctionFly",
    features: [
      "1 function",
      "2 providers",
      "100,000 requests/month",
      "Community support",
    ],
    limits: {
      functions: 1,
      providers: 2,
      requests: 100000,
      stateFabrics: 0,
      agents: 0,
      secrets: 0,
      tokensPerSecret: 0,
    },
  },
  STARTER: {
    id: "starter",
    name: "Starter",
    price: 29,
    priceCents: 2900,
    priceId: import.meta.env.VITE_STRIPE_PRICE_STARTER || "price_starter_placeholder",
    description: "For side projects and MVPs",
    features: [
      "5 functions",
      "3 providers",
      "1M requests/month",
      "1 custom domain",
      "Email support",
      "Basic analytics",
    ],
    limits: {
      functions: 5,
      providers: 3,
      requests: 1000000,
      customDomains: 1,
      stateFabrics: 1,
      agents: 2,
      secrets: 10,
      tokensPerSecret: 5,
    },
  },
  PROFESSIONAL: {
    id: "professional",
    name: "Professional",
    price: 99,
    priceCents: 9900,
    priceId: import.meta.env.VITE_STRIPE_PRICE_PROFESSIONAL || "price_professional_placeholder",
    description: "For growing businesses and SaaS applications",
    features: [
      "25 functions",
      "5 providers",
      "10M requests/month",
      "5 custom domains",
      "99.9% SLA",
      "Priority support",
      "Advanced analytics",
      "Team collaboration",
      "Custom routing rules",
    ],
    limits: {
      functions: 25,
      providers: 5,
      requests: 10000000,
      customDomains: 5,
      sla: "99.9%",
      stateFabrics: 5,
      agents: 10,
      secrets: 50,
      tokensPerSecret: 20,
    },
  },
  ENTERPRISE: {
    id: "enterprise",
    name: "Enterprise",
    price: "Custom",
    priceCents: 0,
    priceId: "",
    description: "For large-scale applications and enterprises",
    features: [
      "Unlimited functions",
      "All providers",
      "Unlimited requests",
      "Unlimited custom domains",
      "99.99% SLA",
      "Dedicated support",
      "Custom integrations",
      "SLA guarantees",
      "On-premise deployment",
    ],
    limits: {
      functions: Infinity,
      providers: Infinity,
      requests: Infinity,
      customDomains: Infinity,
      sla: "99.99%",
      stateFabrics: Infinity,
      agents: Infinity,
      secrets: 10000,
      tokensPerSecret: 100,
    },
  },
} as const;

// ============================================================================
// Agent Execution Plans (AEP) - AI Agent Infrastructure
// ============================================================================

export const AGENT_PLANS = {
  STARTER: {
    id: "agent_starter",
    name: "Agent Starter",
    price: 49,
    priceCents: 4900,
    priceId: import.meta.env.VITE_STRIPE_PRICE_AGENT_STARTER || "price_agent_starter_placeholder",
    description: "For small AI agent projects and prototyping",
    features: [
      "500K tool calls/month",
      "10 concurrent agents",
      "50 calls/second burst",
      "$5 daily spend cap",
      "1K state writes/hour",
      "10GB memory storage",
      "30-day log retention",
      "Per-agent cost tracking",
      "Email support",
    ],
    limits: {
      callsPerMonth: 500000,
      concurrency: 10,
      burst: 50,
      dailySpendCap: 5,
      stateWritesPerHour: 1000,
      memoryGB: 10,
      logRetentionDays: 30,
    },
  },
  SCALE: {
    id: "agent_scale",
    name: "Agent Scale",
    price: 299,
    priceCents: 29900,
    priceId: import.meta.env.VITE_STRIPE_PRICE_AGENT_SCALE || "price_agent_scale_placeholder",
    description: "For growing AI agent applications",
    features: [
      "5M tool calls/month",
      "100 concurrent agents",
      "500 calls/second burst",
      "$30 daily spend cap",
      "10K state writes/hour",
      "100GB memory storage",
      "90-day log retention",
      "Per-agent cost tracking",
      "Budget enforcement",
      "Priority support",
    ],
    limits: {
      callsPerMonth: 5000000,
      concurrency: 100,
      burst: 500,
      dailySpendCap: 30,
      stateWritesPerHour: 10000,
      memoryGB: 100,
      logRetentionDays: 90,
    },
  },
  PRO: {
    id: "agent_pro",
    name: "Agent Pro",
    price: 999,
    priceCents: 99900,
    priceId: import.meta.env.VITE_STRIPE_PRICE_AGENT_PRO || "price_agent_pro_placeholder",
    description: "For production AI agent systems",
    features: [
      "25M tool calls/month",
      "500 concurrent agents",
      "2000 calls/second burst",
      "$100 daily spend cap",
      "50K state writes/hour",
      "500GB memory storage",
      "1-year log retention",
      "Per-agent cost tracking",
      "Budget enforcement",
      "Multi-agent coordination",
      "Dedicated support",
    ],
    limits: {
      callsPerMonth: 25000000,
      concurrency: 500,
      burst: 2000,
      dailySpendCap: 100,
      stateWritesPerHour: 50000,
      memoryGB: 500,
      logRetentionDays: 365,
    },
  },
  ENTERPRISE: {
    id: "agent_enterprise",
    name: "Agent Enterprise",
    price: "Custom",
    priceCents: 250000,
    priceId: "",
    description: "For large-scale AI agent deployments",
    features: [
      "Unlimited tool calls",
      "Unlimited concurrent agents",
      "Unlimited burst",
      "Custom spend caps",
      "Unlimited state writes",
      "Unlimited memory storage",
      "Unlimited log retention",
      "Per-agent cost attribution",
      "Custom budget controls",
      "Multi-region deployment",
      "Dedicated infrastructure",
      "24/7 phone support",
    ],
    limits: {
      callsPerMonth: Infinity,
      concurrency: Infinity,
      burst: Infinity,
      dailySpendCap: -1,
      stateWritesPerHour: Infinity,
      memoryGB: Infinity,
      logRetentionDays: -1,
    },
  },
} as const;

// ============================================================================
// State Fabric Plans - pricing for stateful serverless capabilities
// ============================================================================

export const STATE_FABRIC_PLANS = {
  SANDBOX: {
    id: "sf_sandbox",
    name: "Sandbox",
    price: 0,
    priceCents: 0,
    priceId: "",
    description: "Great for experimentation & onboarding.",
  },
  STARTER: {
    id: "sf_starter",
    name: "Starter",
    price: 19,
    priceCents: 1900,
    priceId: import.meta.env.VITE_STRIPE_PRICE_SF_STARTER || "price_sf_starter_placeholder",
    description: "Internal dev & small side projects.",
  },
  PRO: {
    id: "sf_pro",
    name: "Pro",
    price: 99,
    priceCents: 9900,
    priceId: import.meta.env.VITE_STRIPE_PRICE_SF_PRO || "price_sf_pro_placeholder",
    description: "Production apps & team projects.",
  },
  BUSINESS: {
    id: "sf_business",
    name: "Business",
    price: 499,
    priceCents: 49900,
    priceId: import.meta.env.VITE_STRIPE_PRICE_SF_BUSINESS || "price_sf_business_placeholder",
    description: "Business-critical systems.",
  },
  ENTERPRISE: {
    id: "sf_enterprise",
    name: "Enterprise",
    price: "Custom",
    priceCents: 0,
    priceId: "",
    description: "Large enterprises & regulated industries.",
  },
} as const;

export const STATUS_COLORS = {
  online: {
    bg: "bg-emerald-500",
    text: "text-emerald-500",
    border: "border-emerald-500",
    label: "Online",
  },
  offline: {
    bg: "bg-red-500",
    text: "text-red-500",
    border: "border-red-500",
    label: "Offline",
  },
  degraded: {
    bg: "bg-amber-500",
    text: "text-amber-500",
    border: "border-amber-500",
    label: "Degraded",
  },
  pending: {
    bg: "bg-gray-500",
    text: "text-gray-500",
    border: "border-gray-500",
    label: "Pending",
  },
} as const;

/**
 * Canonical API base URL (no trailing slash). Use for all API and WebSocket calls.
 * - Production: set VITE_API_URL (e.g. https://api.functionfly.com).
 * - Local dev: set VITE_API_URL=http://localhost:8080 to hit the API directly (recommended), or VITE_API_URL=/api to use the Vite proxy (backend must be running on 8080). If you see 404 on /api/v1/api-keys, use http://localhost:8080 or start the backend.
 */
export function getApiBaseUrl(): string {
  const env = (import.meta.env.VITE_API_URL ?? "").trim();
  if (import.meta.env.DEV) {
    // In dev, use /api so requests go through Vite proxy to the backend (avoids "Cannot POST /auth/login" when the request would otherwise hit the dev server).
    if (!env) return "/api";
    const base = env.replace(/\/$/, "");
    try {
      const origin = typeof window !== "undefined" ? window.location.origin : "";
      if (origin && (base === origin || base === "")) return "/api";
    } catch {
      /* ignore */
    }
    return base;
  }
  if (env) return env.replace(/\/$/, "");
  return "https://api.functionfly.com";
}

export const API_BASE_URL = getApiBaseUrl();

/** Path for the AI/LLM discovery manifest (GET). Full URL: ${API_BASE_URL}/.well-known/functionfly.json */
export const WELL_KNOWN_DISCOVERY_PATH = "/.well-known/functionfly.json";
