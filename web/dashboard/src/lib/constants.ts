export const APP_NAME = "FunctionFly";
export const APP_TAGLINE = "Multi-Cloud Failover for Indie SaaS";

export const ROUTES = {
  HOME: "/",
  PRICING: "/pricing",
  LOGIN: "/login",
  SIGNUP: "/signup",
  DASHBOARD: "/dashboard",
  FUNCTIONS: "/functions",
  REGISTRY: "/registry",
  PROVIDERS: "/providers",
  ANALYTICS: "/analytics",
  STATE_FABRIC: "/state-fabric",
  SETTINGS: "/settings",
  APPS: "/apps",
  APP_DETAIL: "/apps/:appId",
  // Admin routes
  ADMIN: "/admin",
  ADMIN_TENANTS: "/admin/tenants",
  ADMIN_USERS: "/admin/users",
  ADMIN_BILLING: "/admin/billing",
  ADMIN_AUDIT: "/admin/audit",
  ADMIN_SYSTEM: "/admin/system",
  // Blog & Content Admin
  ADMIN_CONTENT: "/admin/content",
  ADMIN_REDIRECTS: "/admin/redirects",
  ADMIN_NEWSLETTER: "/admin/newsletter",
  ADMIN_CONTENT_CALENDAR: "/admin/content-calendar",
  ADMIN_FEEDBACK: "/admin/feedback",
  ADMIN_FUNCTIONS: "/admin/functions",
  ADMIN_REGISTRY: "/admin/registry",
  ADMIN_STATE_FABRIC: "/admin/state-fabric",
} as const;

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
    },
  },
  STARTER: {
    id: "starter",
    name: "Starter",
    price: 29,
    priceCents: 2900,
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
    },
  },
  PROFESSIONAL: {
    id: "professional",
    name: "Professional",
    price: 99,
    priceCents: 9900,
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
    },
  },
  ENTERPRISE: {
    id: "enterprise",
    name: "Enterprise",
    price: "Custom",
    priceCents: 0,
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
    },
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

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "/api";
