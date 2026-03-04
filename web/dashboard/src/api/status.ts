import { apiClient } from './client';

// ============================================================================
// Types matching backend DTOs
// ============================================================================

export type PlatformStatusType = 'operational' | 'degraded' | 'major_outage';
export type ComponentStatusType = 'operational' | 'degraded' | 'down';
export type IncidentSeverity = 'critical' | 'high' | 'medium' | 'low';
export type IncidentStatus = 'investigating' | 'identified' | 'monitoring' | 'resolved';
export type ProviderType = 'cloudflare' | 'vercel' | 'fly' | 'deno' | 'functionfly_edge';

export interface PlatformStatus {
  status: PlatformStatusType;
  message: string;
  timestamp: string;
  components: ComponentHealth[];
}

export interface ComponentHealth {
  id: string;
  name: string;
  category: 'core' | 'provider' | 'infrastructure';
  status: ComponentStatusType;
  latency_ms: number;
  uptime_percent: number;
  last_checked: string;
  message?: string;
}

/** Raw component shape from API (type, uptime_24h, response_time_ms) */
interface ApiComponent {
  id: string;
  name: string;
  type?: string;
  status: string;
  description?: string;
  uptime_24h?: number;
  uptime_7d?: number;
  uptime_30d?: number;
  response_time_ms?: number;
  last_checked?: string;
  category?: 'core' | 'provider' | 'infrastructure';
  latency_ms?: number;
  uptime_percent?: number;
  message?: string;
}

const COMPONENT_TYPE_TO_CATEGORY: Record<string, 'core' | 'provider' | 'infrastructure'> = {
  api: 'core',
  database: 'core',
  cache: 'core',
  monitoring: 'core',
  provider: 'provider',
};
const API_STATUS_TO_UI: Record<string, ComponentStatusType> = {
  operational: 'operational',
  degraded_performance: 'degraded',
  partial_outage: 'degraded',
  degraded: 'degraded',
  major_outage: 'down',
  maintenance: 'degraded',
  down: 'down',
};

const VALID_STATUSES: ComponentStatusType[] = ['operational', 'degraded', 'down'];

function normalizeComponent(c: ApiComponent): ComponentHealth {
  const category =
    c.category ?? COMPONENT_TYPE_TO_CATEGORY[c.type ?? ''] ?? 'infrastructure';
  const rawStatus = API_STATUS_TO_UI[String(c.status ?? '')] ?? 'operational';
  const status: ComponentStatusType = VALID_STATUSES.includes(rawStatus)
    ? rawStatus
    : 'operational';
  const uptime = c.uptime_percent ?? c.uptime_24h ?? 0; // Default to 0 instead of 99.9
  const latency = c.latency_ms ?? c.response_time_ms ?? 0;
  const lastChecked =
    typeof c.last_checked === 'string'
      ? c.last_checked
      : c.last_checked != null
        ? new Date(c.last_checked).toISOString()
        : new Date().toISOString();
  return {
    id: c.id,
    name: c.name,
    category,
    status,
    latency_ms: latency,
    uptime_percent: uptime,
    last_checked: lastChecked,
    message: c.message ?? c.description,
  };
}

export interface RegionStatus {
  region: string;
  status: ComponentStatusType;
  latency_ms: number;
  success_rate: number;
}

export interface ProviderStatus {
  id: ProviderType;
  name: string;
  status: ComponentStatusType;
  regions: RegionStatus[];
  avg_latency_ms: number;
  avg_success_rate: number;
  last_updated: string;
}

export interface Incident {
  id: string;
  title: string;
  description: string;
  severity: IncidentSeverity;
  status: IncidentStatus;
  affected_components: string[];
  created_at: string;
  updated_at: string;
  resolved_at?: string;
  created_by?: string;
  updates?: IncidentUpdate[];
}

export interface IncidentUpdate {
  id: string;
  incident_id: string;
  message: string;
  status: IncidentStatus;
  created_at: string;
  created_by: string;
}

export interface CreateIncidentRequest {
  title: string;
  description: string;
  severity: IncidentSeverity;
  status: IncidentStatus;
  affected_components: string[];
}

export interface UpdateIncidentRequest {
  title?: string;
  description?: string;
  severity?: IncidentSeverity;
  status?: IncidentStatus;
  affected_components?: string[];
  message?: string;
}

export interface UptimeMetrics {
  period_days: number;
  overall_uptime: number;
  by_component: Record<string, number>;
  by_provider: Record<string, number>;
  daily_data: {
    date: string;
    uptime: number;
    incidents: number;
  }[];
}

export interface LatencyMetrics {
  provider: ProviderType;
  region?: string;
  time_range: string;
  avg_latency_ms: number;
  p50_latency_ms: number;
  p95_latency_ms: number;
  p99_latency_ms: number;
  data_points: {
    timestamp: string;
    latency_ms: number;
  }[];
}

export interface MaintenanceWindow {
  id: string;
  title: string;
  description: string;
  scheduled_start: string;
  scheduled_end: string;
  affected_components: string[];
  status: 'scheduled' | 'in_progress' | 'completed' | 'cancelled';
  created_at: string;
}

// ============================================================================
// WebSocket Types
// ============================================================================

export interface StatusWebSocketMessage {
  type: 'status_update' | 'incident_update' | 'maintenance_update' | 'ping';
  timestamp: string;
  data: PlatformStatus | Incident | MaintenanceWindow;
}

// ============================================================================
// API Functions
// ============================================================================

/** Raw platform status from API (updated_at, description) */
interface ApiPlatformStatus {
  status: PlatformStatusType;
  indicator?: string;
  description?: string;
  updated_at?: string;
  components?: ApiComponent[];
  incidents?: unknown[];
  maintenance?: unknown[];
}

/**
 * Get overall platform status (normalized: updated_at -> timestamp, description -> message)
 */
export async function getPlatformStatus(): Promise<PlatformStatus> {
  const raw = await apiClient.get<ApiPlatformStatus>('/v1/status');
  const timestamp =
    typeof raw.updated_at === 'string'
      ? raw.updated_at
      : raw.updated_at != null
        ? new Date(String(raw.updated_at)).toISOString()
        : new Date().toISOString();
  const components = Array.isArray(raw.components)
    ? raw.components.map(normalizeComponent)
    : [];
  return {
    status: raw.status,
    message: raw.description ?? '',
    timestamp,
    components,
  };
}

/**
 * Get all component health statuses (normalized for UI)
 */
export async function getComponents(): Promise<ComponentHealth[]> {
  const raw = await apiClient.get<ApiComponent[]>('/v1/status/components');
  return Array.isArray(raw) ? raw.map(normalizeComponent) : [];
}

/** Raw provider/region from API */
interface ApiProviderStatus {
  name: string;
  display_name?: string;
  overall_status?: string;
  regions?: { code?: string; name?: string; status?: string; latency_ms?: number; error_rate?: number }[];
  summary?: { avg_latency_ms?: number; error_rate?: number };
}

const PROVIDER_NAME_TO_ID: Record<string, ProviderType> = {
  cloudflare: 'cloudflare',
  vercel: 'vercel',
  fly: 'fly',
  deno: 'deno',
  functionfly_edge: 'functionfly_edge',
};
const OVERALL_STATUS_TO_UI: Record<string, ComponentStatusType> = {
  operational: 'operational',
  degraded: 'degraded',
  outage: 'down',
  down: 'down',
};

function normalizeProvider(p: ApiProviderStatus): ProviderStatus {
  const id = PROVIDER_NAME_TO_ID[p.name?.toLowerCase()] ?? (p.name?.toLowerCase() as ProviderType);
  const status = OVERALL_STATUS_TO_UI[p.overall_status ?? ''] ?? 'operational';
  const regions: RegionStatus[] = (p.regions ?? []).map((r) => ({
    region: r.code ?? r.name ?? '',
    status: (OVERALL_STATUS_TO_UI[r.status ?? ''] ?? 'operational') as ComponentStatusType,
    latency_ms: r.latency_ms ?? 0,
    success_rate: 100 - (r.error_rate ?? 0),
  }));
  const summary = p.summary;
  const avgLatency = summary?.avg_latency_ms ?? 0;
  const avgError = summary?.error_rate ?? 0;
  return {
    id,
    name: p.display_name ?? p.name,
    status,
    regions,
    avg_latency_ms: avgLatency,
    avg_success_rate: 100 - avgError,
    last_updated: new Date().toISOString(),
  };
}

/** API response shape for GET /v1/status/providers */
interface ProvidersResponse {
  providers?: ApiProviderStatus[];
  generated_at?: string;
}

/**
 * Get all provider statuses by region (normalized for UI)
 */
export async function getProviders(): Promise<ProviderStatus[]> {
  const raw = await apiClient.get<ProvidersResponse | ApiProviderStatus[]>('/v1/status/providers');
  const list = Array.isArray(raw) ? raw : raw?.providers ?? [];
  return list.map(normalizeProvider);
}

// ============================================================================
// Incident API
// ============================================================================

export interface GetIncidentsParams {
  status?: IncidentStatus | 'all';
  severity?: IncidentSeverity | 'all';
  limit?: number;
  offset?: number;
}

/**
 * List incidents with optional filtering
 */
export async function getIncidents(params: GetIncidentsParams = {}): Promise<{
  incidents: Incident[];
  total: number;
  limit: number;
  offset: number;
}> {
  const queryParams = new URLSearchParams();
  if (params.status && params.status !== 'all') {
    queryParams.append('status', params.status);
  }
  if (params.severity && params.severity !== 'all') {
    queryParams.append('severity', params.severity);
  }
  if (params.limit) {
    queryParams.append('limit', params.limit.toString());
  }
  if (params.offset) {
    queryParams.append('offset', params.offset.toString());
  }

  const query = queryParams.toString();
  const url = `/v1/incidents${query ? `?${query}` : ''}`;
  return apiClient.get(url);
}

/**
 * Get a single incident by ID
 */
export async function getIncident(id: string): Promise<Incident> {
  return apiClient.get<Incident>(`/v1/incidents/${id}`);
}

/**
 * Create a new incident (admin only)
 */
export async function createIncident(data: CreateIncidentRequest): Promise<Incident> {
  return apiClient.post<Incident>('/v1/incidents', data);
}

/**
 * Update an incident (admin only)
 */
export async function updateIncident(
  id: string,
  data: UpdateIncidentRequest
): Promise<Incident> {
  return apiClient.patch<Incident>(`/v1/incidents/${id}`, data);
}

// ============================================================================
// Metrics API
// ============================================================================

export type UptimePeriod = 30 | 90 | 365;

/**
 * Get uptime percentages for a given period
 */
export async function getUptimeMetrics(days: UptimePeriod = 30): Promise<UptimeMetrics> {
  return apiClient.get<UptimeMetrics>(`/v1/metrics/uptime?days=${days}`);
}

/**
 * Get latency trends for a specific provider and optional region
 */
export async function getLatencyMetrics(
  provider: ProviderType,
  region?: string
): Promise<LatencyMetrics> {
  const queryParams = new URLSearchParams();
  queryParams.append('provider', provider);
  if (region) {
    queryParams.append('region', region);
  }
  return apiClient.get<LatencyMetrics>(`/v1/metrics/latency?${queryParams.toString()}`);
}

// ============================================================================
// Maintenance API
// ============================================================================

/**
 * Get scheduled maintenance windows
 */
export async function getMaintenance(): Promise<MaintenanceWindow[]> {
  const response = await apiClient.get<{ maintenance_windows: MaintenanceWindow[] }>('/v1/maintenance');
  return response.maintenance_windows || [];
}

// ============================================================================
// Status API Export
// ============================================================================

export const statusApi = {
  getPlatformStatus,
  getComponents,
  getProviders,
  getIncidents,
  getIncident,
  createIncident,
  updateIncident,
  getUptimeMetrics,
  getLatencyMetrics,
  getMaintenance,
};
