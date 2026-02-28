import { apiClient } from "./client";

export interface UsageDataPoint {
  time: string;
  value: number;
}

export interface ExecutionRateDataPoint {
  time: string;
  rate: number;
}

export interface DashboardActivityItem {
  id: string;
  type: string;
  title: string;
  description?: string;
  timestamp: string;
  function_id?: string;
  function_name?: string;
}

export interface DashboardUsageResponse {
  data: UsageDataPoint[];
}

export interface DashboardExecutionRateResponse {
  data: ExecutionRateDataPoint[];
}

export interface DashboardActivityResponse {
  activities: DashboardActivityItem[];
}

export const dashboardApi = {
  getUsage: async (days = 14): Promise<DashboardUsageResponse> => {
    const res = await apiClient.get<DashboardUsageResponse>(
      `/v1/dashboard/usage?days=${days}`
    );
    return res ?? { data: [] };
  },

  getExecutionRate: async (hours = 24): Promise<DashboardExecutionRateResponse> => {
    const res = await apiClient.get<DashboardExecutionRateResponse>(
      `/v1/dashboard/execution-rate?hours=${hours}`
    );
    return res ?? { data: [] };
  },

  getActivity: async (limit = 20): Promise<DashboardActivityResponse> => {
    const res = await apiClient.get<DashboardActivityResponse>(
      `/v1/dashboard/activity?limit=${limit}`
    );
    return res ?? { activities: [] };
  },
};
