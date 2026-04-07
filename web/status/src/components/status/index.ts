// Import types for local usage
import type { Provider } from "./types";

// Types
export type {
  HeaderProps,
  HeroStatusProps,
  IncidentTimelineProps,
  MaintenanceSectionProps,
  Metric,
  MetricsSectionProps,
  Provider,
  ProviderCardProps,
  ProviderSectionProps,
  ServiceCardProps,
  StatusOrbitalProps,
  UptimeHistorySectionProps,
} from "./types";

// Re-export API types directly from lib/api
export type {
  Component,
  Incident,
  IncidentUpdate,
  MaintenanceSummary,
  UptimeDataPoint,
} from "@/lib/api";

// Skeletons
export {
  HeroSkeleton,
  IncidentSkeleton,
  MetricsSectionSkeleton,
  ServiceCardSkeleton,
} from "./skeletons";

// Backgrounds & Effects
export { AnimatedBackground, StatusOrbital } from "./backgrounds";

// Cards
export { ProviderCard, ServiceCard } from "./cards";

// Sections
export {
  IncidentTimeline,
  MaintenanceSection,
  MetricsSection,
  ProviderSection,
  UptimeHistorySection,
} from "./sections";

// Layout
export { Footer, Header, HeroStatus, SubscribeSection } from "./layout";

// Default providers — user-selectable deployment targets matching GetAdapterForProvider
export const defaultProviders: Provider[] = [
  {
    id: "workers",
    name: "Cloudflare Workers",
    type: "edge",
    status: "operational",
    region: "300+ Locations",
    latency: 12,
    healthScore: 99.99,
    description: "Edge compute on Cloudflare's global network with V8 isolates",
  },
  {
    id: "vercel",
    name: "Vercel Functions",
    type: "edge",
    status: "operational",
    region: "Edge Network",
    latency: 15,
    healthScore: 99.98,
    description: "Serverless edge functions on Vercel's global infrastructure",
  },
  {
    id: "fly",
    name: "Fly.io",
    type: "compute",
    status: "operational",
    region: "Global",
    latency: 45,
    healthScore: 99.97,
    description: "Containerized compute deployed close to users worldwide",
  },
  {
    id: "deno-deploy",
    name: "Deno Deploy",
    type: "edge",
    status: "operational",
    region: "Edge Network",
    latency: 18,
    healthScore: 99.95,
    description: "Edge runtime on Deno's distributed JavaScript cloud",
  },
  {
    id: "functionfly-edge",
    name: "FunctionFly Edge",
    type: "edge",
    status: "operational",
    region: "Global Edge",
    latency: 8,
    healthScore: 99.98,
    description: "FunctionFly's managed edge infrastructure with smart routing",
  },
];
