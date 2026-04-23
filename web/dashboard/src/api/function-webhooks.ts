import { apiClient } from "./client";

export interface FunctionWebhook {
  id: string;
  function_id?: string;
  url: string;
  secret: string;
  event_types: string[];
  active: boolean;
  created_at: string;
  updated_at: string;
  created_by: string;
}

export interface WebhookDelivery {
  id: string;
  webhook_id: string;
  event_type: string;
  payload: Record<string, unknown>;
  response_status?: number;
  response_body?: string;
  error?: string;
  attempted_at: string;
  delivered_at?: string;
}

export interface WebhookListResponse {
  subscriptions: FunctionWebhook[];
  total_count: number;
  page: number;
  page_size: number;
}

export interface WebhookDeliveryListResponse {
  deliveries: WebhookDelivery[];
  total_count: number;
  page: number;
  page_size: number;
}

export interface WebhookCreateRequest {
  url: string;
  function_id?: string;
  event_types: string[];
  secret?: string;
}

export interface WebhookUpdateRequest {
  url?: string;
  event_types?: string[];
  active?: boolean;
}

export const functionWebhooksApi = {
  list: async (): Promise<WebhookListResponse> => {
    const response = await apiClient.get<WebhookListResponse>("/v1/function-webhooks");
    return response;
  },

  get: async (id: string): Promise<{ data: FunctionWebhook }> => {
    const response = await apiClient.get<{ data: FunctionWebhook }>(`/v1/function-webhooks/${id}`);
    return response;
  },

  create: async (data: WebhookCreateRequest): Promise<{ data: FunctionWebhook }> => {
    const response = await apiClient.post<{ data: FunctionWebhook }>("/v1/function-webhooks", data);
    return response;
  },

  update: async (id: string, data: WebhookUpdateRequest): Promise<{ data: FunctionWebhook }> => {
    const response = await apiClient.patch<{ data: FunctionWebhook }>(`/v1/function-webhooks/${id}`, data);
    return response;
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/v1/function-webhooks/${id}`);
  },

  test: async (id: string): Promise<{ success: boolean; delivery_id?: string }> => {
    const response = await apiClient.post<{ success: boolean; delivery_id?: string }>(`/v1/function-webhooks/${id}/test`);
    return response;
  },

  listDeliveries: async (id: string): Promise<WebhookDeliveryListResponse> => {
    const response = await apiClient.get<WebhookDeliveryListResponse>(`/v1/function-webhooks/${id}/deliveries`);
    return response;
  },
};