import { apiClient } from "./client";
import type { Deployment, DeployRequest, DeployResult } from "@/types";

export const deploymentsApi = {
  list: (appId: string, limit?: number) =>
    apiClient.get<{ deployments: Deployment[] }>(`/v1/apps/${appId}/deployments`, {
      params: { limit },
    }),

  get: (deploymentId: string) =>
    apiClient.get<Deployment>(`/v1/deployments/${deploymentId}`),

  deploy: (appId: string, data: DeployRequest) =>
    apiClient.post<DeployResult>(`/v1/apps/${appId}/deploy`, data),

  rollback: (deploymentId: string) =>
    apiClient.post<DeployResult>(`/v1/deployments/${deploymentId}/rollback`),
};
