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

/**
 * Get overall platform status
 */
export async function getPlatformStatus(): Promise<PlatformStatus> {
  return apiClient.get<PlatformStatus>('/v1/status');
}

/**
 * Get all component health statuses
 */
export async function getComponents(): Promise<ComponentHealth[]> {
  return apiClient.get<ComponentHealth[]>('/v1/status/components');
}

/**
 * Get all provider statuses by region
 */
export async function getProviders(): Promise<ProviderStatus[]> {
  return apiClient.get<ProviderStatus[]>('/v1/status/providers');
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
  return apiClient.get<MaintenanceWindow[]>('/v1/maintenance');
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
