/**
 * State Fabric pricing plans – tiered strategy (Sandbox → Enterprise).
 * Edit here to keep pricing page and marketing in sync.
 * See strategic pricing doc for add-on pricing (storage, ops, hot cache, etc.).
 */

export interface StateFabricPlan {
  id: string;
  name: string;
  tagline?: string;
  price: string;
  period: string;
  description: string;
  features: string[];
  addOns?: string;
  highlighted: boolean;
  cta: string;
  href: string;
}

export const STATE_FABRIC_PLANS: StateFabricPlan[] = [
  {
    id: "sandbox",
    name: "Sandbox",
    tagline: "Free",
    price: "$0",
    period: "month",
    description: "Great for experimentation & onboarding.",
    features: [
      "Stateless usage",
      "1 state object",
      "1 GB event storage",
      "10,000 state operations",
      "7-day snapshot retention",
      "Community support",
    ],
    highlighted: false,
    cta: "Get Started",
    href: "/signup",
  },
  {
    id: "starter",
    name: "Starter",
    tagline: "Dev State",
    price: "$19",
    period: "month",
    description: "Internal dev & small side projects.",
    features: [
      "Up to 5 state objects",
      "10 GB event storage",
      "100,000 state ops",
      "30-day snapshot retention",
      "Basic snapshot scheduling",
      "Replay engine (limited)",
    ],
    addOns: "Storage: $0.10/GB · Ops: $0.50/100k",
    highlighted: false,
    cta: "Start Free Trial",
    href: "/signup",
  },
  {
    id: "pro",
    name: "Pro",
    tagline: "Persistent Apps",
    price: "$99",
    period: "month",
    description: "Production apps & team projects.",
    features: [
      "Up to 50 state objects",
      "100 GB event storage",
      "1M state ops",
      "90-day snapshot retention",
      "Hot cache tier",
      "Fast deterministic replay",
      "Replay API + billing analytics",
    ],
    addOns: "Storage: $0.08/GB · Ops: $0.35/100k",
    highlighted: true,
    cta: "Start Free Trial",
    href: "/signup",
  },
  {
    id: "business",
    name: "Business",
    tagline: "Scale State",
    price: "$499",
    period: "month",
    description: "Business-critical systems.",
    features: [
      "Up to 500 state objects",
      "1 TB event storage",
      "10M state ops",
      "180-day retention",
      "Multi-region replication",
      "Dedicated hot cache",
      "SLA + advanced analytics",
      "Event subscription streams",
    ],
    addOns: "Storage: $0.06/GB · Archive: $30/TB/mo",
    highlighted: false,
    cta: "Start Free Trial",
    href: "/signup",
  },
  {
    id: "enterprise",
    name: "Enterprise",
    tagline: "StateFabric Enterprise",
    price: "$1,999",
    period: "month+",
    description: "Large enterprises & regulated industries.",
    features: [
      "Unlimited state objects",
      "Unlimited event storage",
      "Unlimited ops",
      "365-day retention",
      "Multi-region + edge replication",
      "Enterprise key management (BYOK)",
      "Immutable audit logs",
      "Replay export · Dedicated support",
    ],
    highlighted: false,
    cta: "Contact Sales",
    href: "/contact",
  },
];

/** Comparison table rows for State Fabric tiers */
export const STATE_FABRIC_COMPARISON_ROWS = [
  { feature: "State objects", sandbox: "1", starter: "5", pro: "50", business: "500", enterprise: "Unlimited" },
  { feature: "Event storage", sandbox: "1 GB", starter: "10 GB", pro: "100 GB", business: "1 TB", enterprise: "Unlimited" },
  { feature: "State ops", sandbox: "10K", starter: "100K", pro: "1M", business: "10M", enterprise: "Unlimited" },
  { feature: "Snapshot retention", sandbox: "7 days", starter: "30 days", pro: "90 days", business: "180 days", enterprise: "365 days" },
  { feature: "Replay", sandbox: "—", starter: "Limited", pro: "Fast + API", business: "Full", enterprise: "Export" },
  { feature: "Multi-region", sandbox: "—", starter: "—", pro: "—", business: "Yes", enterprise: "Yes + edge" },
  { feature: "Hot cache", sandbox: "—", starter: "—", pro: "Yes", business: "Dedicated", enterprise: "Dedicated" },
  { feature: "Support", sandbox: "Community", starter: "Email", pro: "Priority", business: "SLA", enterprise: "Dedicated" },
];

/** Optional add-ons (available across tiers) */
export const STATE_FABRIC_ADDONS = [
  { name: "Hot Cache Booster", price: "$49", period: "/mo per 5GB", description: "Reduces replay and read costs" },
  { name: "Advanced Security Pack", price: "$99", period: "/mo", description: "SOC2-friendly logs, key rotation, audit streams" },
  { name: "AI Memory Pack", price: "$149", period: "/mo", description: "Vector index, embeddings storage, fast read engine" },
  { name: "Advanced Insights", price: "$79", period: "/mo", description: "Cost forecasting, anomaly detection, hot path alerts" },
];
