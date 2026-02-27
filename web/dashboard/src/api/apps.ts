import { apiClient } from "./client";
import type { App, AppStatus, CreateAppRequest, CreateBackendRequest, Backend } from "@/types";

export const appsApi = {
  list: () => apiClient.get<{ apps: App[] }>("/v1/apps"),

  get: (appId: string) => apiClient.get<App>(`/v1/apps/${appId}`),

  create: (data: CreateAppRequest) =>
    apiClient.post<App>("/v1/apps", data),

  getStatus: (appId: string) =>
    apiClient.get<AppStatus>(`/v1/apps/${appId}/status`),

  listBackends: (appId: string) =>
    apiClient.get<{ backends: Backend[] }>(`/v1/apps/${appId}/backends`),

  createBackend: (appId: string, data: CreateBackendRequest) =>
    apiClient.post<Backend>(`/v1/apps/${appId}/backends`, data),

  getRoute: (appId: string, params?: { clientRegion?: string; method?: string }) =>
    apiClient.get(`/v1/apps/${appId}/route`, { params }),
};
