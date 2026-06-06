import { apiClient } from "./client";

export interface Connector {
  id: string;
  slug: string;
  name: string;
  icon_url: string;
  oauth_url: string;
  scopes: string;
  is_active: boolean;
  created_at: string;
}

export interface UserConnector {
  id: string;
  tenant_id: string;
  connector_id: string;
  connector_slug: string;
  connector_name: string;
  connector_icon_url: string;
  display_name: string;
  status: "active" | "sync_error" | "revoked";
  encrypted_credentials: Record<string, unknown>;
  last_sync_at?: string;
  sync_error?: string;
  created_at: string;
  updated_at: string;
}

export interface LinkConnectorRequest {
  connector_slug: string;
  display_name?: string;
  encrypted_credentials?: Record<string, unknown>;
  oauth_code?: string;
  redirect_uri?: string;
}

export interface SyncTriggerResponse {
  status: string;
  message: string;
  started_at: string;
}

export const connectorsApi = {
  listCatalog: async (): Promise<Connector[]> => {
    const res = await apiClient.get<{ connectors: Connector[] }>(
      "/v1/connectors/catalog"
    );
    return res.connectors;
  },

  listUserConnectors: async (): Promise<UserConnector[]> => {
    const res = await apiClient.get<{ connectors: UserConnector[] }>(
      "/v1/connectors"
    );
    return res.connectors;
  },

  linkConnector: async (data: LinkConnectorRequest): Promise<UserConnector> => {
    const res = await apiClient.post<{ connector: UserConnector }>(
      "/v1/connectors/link",
      data
    );
    return res.connector;
  },

  oauthCallback: async (code: string, state: string): Promise<void> => {
    await apiClient.post("/v1/connectors/callback", { code, state });
  },

  unlinkConnector: async (connectorId: string): Promise<void> => {
    await apiClient.delete(`/v1/connectors/${connectorId}`);
  },

  triggerSync: async (connectorId: string): Promise<SyncTriggerResponse> => {
    return apiClient.post<SyncTriggerResponse>(
      `/v1/connectors/${connectorId}/sync`
    );
  },
};
