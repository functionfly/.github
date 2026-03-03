import { apiClient } from "./client";

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
