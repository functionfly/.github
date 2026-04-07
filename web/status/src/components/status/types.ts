import type { Component, Incident, MaintenanceSummary } from "@/lib/api";
import type { ReactNode } from "react";

// Re-export other API types
export type { IncidentUpdate, UptimeDataPoint } from "@/lib/api";

// Local types
export interface Metric {
  label: string;
  value: string;
  /** Short secondary line; omit when there is nothing meaningful to show */
  change?: string;
  trend: "up" | "down" | "neutral";
  icon: ReactNode;
}

export interface Provider {
  id: string;
  name: string;
  type:
    | "cloud"
    | "cdn"
    | "database"
    | "storage"
    | "compute"
    | "edge"
    | "ai"
    | "security";
  status: "operational" | "degraded" | "major_outage" | "partial_outage";
  region: string;
  latency: number;
  healthScore: number;
  description: string;
}

export interface StatusOrbitalProps {
  status: string;
  size?: "sm" | "md" | "lg" | "xl";
}

export interface HeroStatusProps {
  overallStatus: string;
  isLoading: boolean;
  lastUpdated: Date;
}

export interface ServiceCardProps {
  component: Component;
  index: number;
}

export interface ProviderCardProps {
  provider: Provider;
  index: number;
}

export interface MetricsSectionProps {
  components: Component[];
  isLoading: boolean;
  probeLatencyMs: number | null;
  probeLatencyLoading: boolean;
}

export interface IncidentTimelineProps {
  incidents: Incident[];
  isLoading: boolean;
}

export interface MaintenanceSectionProps {
  maintenance: MaintenanceSummary[];
  isLoading: boolean;
}

export interface UptimeHistorySectionProps {
  isLoading: boolean;
}

export interface ProviderSectionProps {
  providers: Provider[];
  isLoading: boolean;
}

export interface HeaderProps {
  onRefresh: () => void;
  isRefreshing: boolean;
  lastUpdated: Date;
  isRealtimeConnected: boolean;
}
