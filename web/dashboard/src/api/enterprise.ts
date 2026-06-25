import { apiClient } from "./client";

export interface AuditLogItem {
  id: string;
  tenant_id: string;
  service_area: string;
  action: string;
  resource_type: string;
  resource_id: string | null;
  actor_type: string;
  actor_id: string;
  actor_name: string;
  request_id: string;
  ip_address: string;
  user_agent: string;
  metadata: Record<string, unknown>;
  success: boolean;
  error_message: string | null;
  created_at: string;
}

export interface AuditLogsResponse {
  logs: AuditLogItem[];
  total: number;
  limit: number;
  offset: number;
}

export interface AuditFiltersResponse {
  service_areas: string[];
  actions: string[];
}

export interface AuditExportResponse {
  body: string;
  generated_at: string;
  row_count: number;
  signature: string;
}

export const enterpriseAuditApi = {
  listLogs: async (params?: {
    limit?: number;
    offset?: number;
    service_area?: string;
    action?: string;
    resource_type?: string;
    resource_id?: string;
    actor_type?: string;
    actor_id?: string;
    success?: boolean;
    start_time?: string;
    end_time?: string;
    search?: string;
  }): Promise<AuditLogsResponse> => {
    const searchParams = new URLSearchParams();
    if (params?.limit) searchParams.set('limit', String(params.limit));
    if (params?.offset) searchParams.set('offset', String(params.offset));
    if (params?.service_area) searchParams.set('service_area', params.service_area);
    if (params?.action) searchParams.set('action', params.action);
    if (params?.resource_type) searchParams.set('resource_type', params.resource_type);
    if (params?.resource_id) searchParams.set('resource_id', params.resource_id);
    if (params?.actor_type) searchParams.set('actor_type', params.actor_type);
    if (params?.actor_id) searchParams.set('actor_id', params.actor_id);
    if (params?.success !== undefined) searchParams.set('success', String(params.success));
    if (params?.start_time) searchParams.set('start_time', params.start_time);
    if (params?.end_time) searchParams.set('end_time', params.end_time);
    if (params?.search) searchParams.set('search', params.search);

    const res = await apiClient.get<AuditLogsResponse>(
      `/v1/enterprise/audit/logs?${searchParams.toString()}`
    );
    return res;
  },

  getLog: async (id: string): Promise<{ log: AuditLogItem }> => {
    const res = await apiClient.get<{ log: AuditLogItem }>(
      `/v1/enterprise/audit/logs?id=${id}`
    );
    return res;
  },

  getFilters: async (): Promise<AuditFiltersResponse> => {
    const res = await apiClient.get<AuditFiltersResponse>(
      `/v1/enterprise/audit/filters`
    );
    return res;
  },

  exportAudit: async (params?: {
    from?: string;
    to?: string;
    format?: 'json' | 'csv' | 'cef';
    service_area?: string;
    action?: string;
  }): Promise<AuditExportResponse> => {
    const searchParams = new URLSearchParams();
    if (params?.from) searchParams.set('from', params.from);
    if (params?.to) searchParams.set('to', params.to);
    if (params?.format) searchParams.set('format', params.format);
    if (params?.service_area) searchParams.set('service_area', params.service_area);
    if (params?.action) searchParams.set('action', params.action);

    const res = await apiClient.get<AuditExportResponse>(
      `/v1/enterprise/audit/export?${searchParams.toString()}`
    );
    return res;
  },
};

export interface SLAOverviewResponse {
  current_uptime_percent: number;
  sla_target_percent: number;
  incident_count: number;
  period_days: number;
}

export interface UptimeHistoryPoint {
  date: string;
  uptime_percent: number;
  incident_count: number;
}

export interface SLAUptimeHistoryResponse {
  period_days: number;
  points: UptimeHistoryPoint[];
}

export interface SLAIncidentItem {
  id: string;
  title: string;
  severity: string;
  status: string;
  description: string;
  created_at: string;
  resolved_at: string;
  updated_at: string;
}

export interface SLAIncidentsResponse {
  incidents: SLAIncidentItem[];
  period_days: number;
}

export const enterpriseSlaApi = {
  getOverview: async (days = 30): Promise<SLAOverviewResponse> => {
    const res = await apiClient.get<SLAOverviewResponse>(
      `/v1/enterprise/sla/overview?days=${days}`
    );
    return res;
  },

  getUptimeHistory: async (days = 30): Promise<SLAUptimeHistoryResponse> => {
    const res = await apiClient.get<SLAUptimeHistoryResponse>(
      `/v1/enterprise/sla/uptime-history?days=${days}`
    );
    return res;
  },

  getIncidents: async (
    params?: { limit?: number; days?: number }
  ): Promise<SLAIncidentsResponse> => {
    const limit = params?.limit ?? 20;
    const days = params?.days ?? 30;
    const res = await apiClient.get<SLAIncidentsResponse>(
      `/v1/enterprise/sla/incidents?limit=${limit}&days=${days}`
    );
    return res;
  },
};
