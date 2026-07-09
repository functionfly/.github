import { format, subDays } from "date-fns";

const envApi = import.meta.env.VITE_API_URL?.trim();
/** Dev default uses Vite proxy (same origin) so API + WS work without CORS. */
export const STATUS_API_BASE_URL =
  envApi || (import.meta.env.DEV ? "/api" : "https://api.functionfly.com");

// ============================================================================
// Types matching backend models
// ============================================================================

export type ComponentStatus =
  | "operational"
  | "degraded"
  | "major_outage"
  | "maintenance"
  | "partial_outage";

export interface Component {
  id: string;
  name: string;
  status: ComponentStatus;
  type: string;
  description?: string;
  uptime_24h: number | null;
  uptime_7d: number | null;
  uptime_30d: number | null;
  response_time_ms: number;
  last_checked?: string;
  history?: StatusHistoryPoint[];
}

export interface StatusHistoryPoint {
  timestamp: string;
  status: ComponentStatus;
  response_time_ms: number;
}

export type IncidentStatus =
  | "investigating"
  | "identified"
  | "monitoring"
  | "resolved";
export type IncidentSeverity = "critical" | "high" | "medium" | "low";

export interface Incident {
  id: string;
  title: string;
  severity: IncidentSeverity;
  status: IncidentStatus;
  description: string;
  affected_components: string[];
  created_at: string;
  resolved_at?: string;
  updated_at: string;
  duration_minutes?: number;
  updates: IncidentUpdate[];
}

export interface IncidentUpdate {
  id: string;
  status: IncidentStatus;
  message: string;
  created_at: string;
  created_by?: UserRef;
}

export interface UserRef {
  id: string;
  name: string;
}

export interface MaintenanceWindow {
  id: string;
  title: string;
  description: string;
  status: "scheduled" | "in_progress" | "completed" | "cancelled";
  scheduled_start: string;
  scheduled_end: string;
  actual_start?: string;
  actual_end?: string;
  affected_components: string[];
  affected_providers: string[];
  created_at: string;
  updated_at: string;
}

export interface MaintenanceSummary {
  id: string;
  title: string;
  status: string;
  scheduled_start: string;
  scheduled_end: string;
}

export interface PlatformStatus {
  status: ComponentStatus;
  indicator: "none" | "minor" | "major" | "critical";
  description: string;
  updated_at: string;
  components: Component[];
  incidents: Incident[];
  maintenance: MaintenanceSummary[];
}

export interface ComponentStatusResponse {
  components: Component[];
  generated_at: string;
}

export interface UptimeMetricsResponse {
  period: string;
  resolution: string;
  overall_uptime: number;
  data_points: UptimeDataPoint[];
}

export interface UptimeDataPoint {
  timestamp: string;
  uptime_percent: number;
  total_checks: number;
  failed_checks: number;
  component_breakdown?: Record<string, number>;
}

export interface LatencyMetricsResponse {
  period: string;
  percentile: string;
  overall_avg_ms: number;
  data_points: LatencyDataPoint[];
  by_provider?: Record<string, LatencyStats>;
}

export interface LatencyDataPoint {
  timestamp: string;
  value_ms: number;
  provider?: string;
}

export interface LatencyStats {
  avg_ms: number;
  min_ms: number;
  max_ms: number;
  p50_ms: number;
  p95_ms: number;
  p99_ms: number;
}

export interface IncidentsListResponse {
  incidents: Incident[];
  pagination: {
    total: number;
    limit: number;
    offset: number;
    has_more: boolean;
  };
}

export interface MaintenanceListResponse {
  maintenance_windows: MaintenanceWindow[];
}

// ============================================================================
// API Client
// ============================================================================

class StatusAPI {
  private baseURL: string;

  constructor(baseURL: string = STATUS_API_BASE_URL) {
    this.baseURL = baseURL;
  }

  private async fetch<T>(
    endpoint: string,
    options: RequestInit = {},
  ): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;
    const response = await fetch(url, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...options.headers,
      },
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`API Error ${response.status}: ${error}`);
    }

    return response.json() as Promise<T>;
  }

  // Status endpoints
  async getPlatformStatus(): Promise<PlatformStatus> {
    return this.fetch<PlatformStatus>("/v1/status");
  }

  async getComponents(
    includeHistory: boolean = false,
  ): Promise<ComponentStatusResponse> {
    const params = new URLSearchParams();
    if (includeHistory) params.set("include_history", "true");
    return this.fetch<ComponentStatusResponse>(
      `/v1/status/components?${params}`,
    );
  }

  async getProviders(
    provider?: string,
    region?: string,
    detailed: boolean = false,
  ): Promise<unknown> {
    const params = new URLSearchParams();
    if (provider) params.set("provider", provider);
    if (region) params.set("region", region);
    if (detailed) params.set("detailed", "true");
    return this.fetch(`/v1/status/providers?${params}`);
  }

  // Incident endpoints
  async listIncidents(
    options: {
      status?: string;
      severity?: string;
      startDate?: Date;
      endDate?: Date;
      limit?: number;
      offset?: number;
    } = {},
  ): Promise<IncidentsListResponse> {
    const params = new URLSearchParams();
    if (options.status) params.set("status", options.status);
    if (options.severity) params.set("severity", options.severity);
    if (options.startDate)
      params.set("start_date", options.startDate.toISOString());
    if (options.endDate) params.set("end_date", options.endDate.toISOString());
    if (options.limit) params.set("limit", options.limit.toString());
    if (options.offset !== undefined)
      params.set("offset", options.offset.toString());
    return this.fetch<IncidentsListResponse>(`/v1/incidents?${params}`);
  }

  async getIncident(id: string): Promise<Incident> {
    return this.fetch<Incident>(`/v1/incidents/${id}`);
  }

  // Metrics endpoints
  async getUptimeMetrics(
    component: string = "all",
    period: string = "24h",
    resolution?: string,
  ): Promise<UptimeMetricsResponse> {
    const params = new URLSearchParams({ component, period });
    if (resolution) params.set("resolution", resolution);
    return this.fetch<UptimeMetricsResponse>(`/v1/metrics/uptime?${params}`);
  }

  async getLatencyMetrics(
    provider: string = "all",
    period: string = "24h",
    percentile: string = "p95",
  ): Promise<LatencyMetricsResponse> {
    const params = new URLSearchParams({ provider, period, percentile });
    return this.fetch<LatencyMetricsResponse>(`/v1/metrics/latency?${params}`);
  }

  // Maintenance endpoints
  async listMaintenance(
    upcoming: boolean = true,
  ): Promise<MaintenanceListResponse> {
    const params = new URLSearchParams();
    if (upcoming) params.set("upcoming", "true");
    return this.fetch<MaintenanceListResponse>(`/v1/maintenance?${params}`);
  }

  // WebSocket connection
  connectWebSocket(): WebSocket {
    if (this.baseURL.startsWith("/")) {
      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      return new WebSocket(`${protocol}//${window.location.host}/ws/v1/status`);
    }
    const wsURL = this.baseURL.replace(/^http/, "ws");
    return new WebSocket(`${wsURL}/ws/v1/status`);
  }
}

// ============================================================================
// External Services
// ============================================================================

export interface StateFabricHealth {
  status: "operational" | "degraded" | "down" | "maintenance";
  latency_ms?: number;
  checked_at: string;
}

export interface DedicatedServerHealth {
  status: "operational" | "degraded" | "down" | "maintenance";
  latency_ms?: number;
  checked_at: string;
}

export async function fetchStateFabricHealth(): Promise<StateFabricHealth> {
  const start = Date.now();
  try {
    const res = await fetch(`${STATUS_API_BASE_URL}/v1/status/statefabric`, {
      signal: AbortSignal.timeout(5000),
    });
    const latency_ms = Date.now() - start;
    if (!res.ok) {
      return { status: "down", latency_ms, checked_at: new Date().toISOString() };
    }
    const data = await res.json();
    return {
      status: data.status === "operational" ? "operational" :
              data.status === "degraded" ? "degraded" :
              data.status === "maintenance" ? "maintenance" : "down",
      latency_ms,
      checked_at: new Date().toISOString(),
    };
  } catch {
    return { status: "down", latency_ms: Date.now() - start, checked_at: new Date().toISOString() };
  }
}

export async function fetchDedicatedServerHealth(): Promise<DedicatedServerHealth> {
  const start = Date.now();
  try {
    const res = await fetch(`${STATUS_API_BASE_URL}/health/dedicated`, {
      signal: AbortSignal.timeout(5000),
    });
    const latency_ms = Date.now() - start;
    if (!res.ok) {
      return { status: "down", latency_ms, checked_at: new Date().toISOString() };
    }
    const data = await res.json();
    return {
      status: data.status === "healthy" ? "operational" :
              data.status === "degraded" ? "degraded" :
              data.status === "maintenance" ? "maintenance" : "down",
      latency_ms,
      checked_at: new Date().toISOString(),
    };
  } catch {
    return { status: "down", latency_ms: Date.now() - start, checked_at: new Date().toISOString() };
  }
}

// Singleton instance
export const statusAPI = new StatusAPI();

// ============================================================================
// Helper Functions
// ============================================================================

export function formatIncidentDate(date: string): string {
  return format(new Date(date), "MMM d, yyyy, h:mm a");
}

export function getIncidentDuration(
  startedAt: string,
  resolvedAt?: string,
): string {
  const start = new Date(startedAt);
  const end = resolvedAt ? new Date(resolvedAt) : new Date();
  const diff = end.getTime() - start.getTime();

  const hours = Math.floor(diff / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

  if (hours > 0) {
    return `${hours}h ${minutes}m`;
  }
  return `${minutes}m`;
}

export function getSeverityColor(severity: IncidentSeverity): string {
  switch (severity) {
    case "critical":
      return "text-red-400 bg-red-500/10 border-red-500/30";
    case "high":
      return "text-red-400 bg-red-500/10 border-red-500/30";
    case "medium":
      return "text-amber-400 bg-amber-500/10 border-amber-500/30";
    case "low":
      return "text-emerald-400 bg-emerald-500/10 border-emerald-500/30";
    default:
      return "text-gray-400 bg-gray-500/10 border-gray-500/30";
  }
}

export function getStatusColor(
  status: ComponentStatus | IncidentStatus,
): string {
  switch (status) {
    case "operational":
    case "resolved":
      return "text-emerald-400 bg-emerald-500/10 border-emerald-500/30";
    case "degraded":
    case "monitoring":
      return "text-amber-400 bg-amber-500/10 border-amber-500/30";
    case "major_outage":
    case "investigating":
    case "identified":
      return "text-red-400 bg-red-500/10 border-red-500/30";
    case "maintenance":
      return "text-purple-400 bg-purple-500/10 border-purple-500/30";
    case "partial_outage":
      return "text-orange-400 bg-orange-500/10 border-orange-500/30";
    default:
      return "text-gray-400 bg-gray-500/10 border-gray-500/30";
  }
}

// Helper to get date range for incident queries
export function getDateRange(days: number): { startDate: Date; endDate: Date } {
  const endDate = new Date();
  const startDate = subDays(endDate, days);
  return { startDate, endDate };
}
