import { apiClient } from './client';

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
  status: 'active' | 'sync_error' | 'revoked' | 'disabled';
  encrypted_credentials: Record<string, unknown>;
  last_sync_at?: string;
  sync_error?: string;
  created_at: string;
  updated_at: string;
  sync_frequency?: string;
  auto_sync?: boolean;
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

export interface OAuthUrlResponse {
  oauth_url: string;
  state: string;
}

export interface OAuthCallbackResponse {
  status: string;
  tokens: {
    access_token: string;
    refresh_token?: string;
    expires_in: number;
    token_type: string;
    raw?: Record<string, unknown>;
  };
  server_encrypted: {
    ciphertext: string;
    iv: string;
    salt: string;
    tag: string;
    key_version: number;
  };
  connector_slug: string;
}

export interface UpdateConnectorRequest {
  enabled?: boolean;
  display_name?: string;
  sync_frequency?: string;
  auto_sync?: boolean;
}

export const connectorsApi = {
  listCatalog: async (): Promise<Connector[]> => {
    const res = await apiClient.get<{ connectors: Connector[] }>('/v1/connectors/catalog');
    return res.connectors;
  },

  listUserConnectors: async (): Promise<UserConnector[]> => {
    const res = await apiClient.get<{ connectors: UserConnector[] }>('/v1/connectors');
    return res.connectors;
  },

  /**
   * Get OAuth URL for a connector — initiates the OAuth flow.
   * Returns { oauth_url, state }.
   */
  getConnectorOAuthUrl: async (slug: string): Promise<OAuthUrlResponse> => {
    return apiClient.get<OAuthUrlResponse>(
      `/v1/connectors/oauth-url?slug=${encodeURIComponent(slug)}`
    );
  },

  /**
   * Link a connector after OAuth callback (called with connector_slug after
   * the user has authenticated via the popup). The backend creates the
   * user_connector record.
   */
  linkConnector: async (data: LinkConnectorRequest): Promise<UserConnector> => {
    const res = await apiClient.post<{ connector: UserConnector }>('/v1/connectors/link', data);
    return res.connector;
  },

  /**
   * Update a linked connector's settings: enable/disable, display name,
   * sync frequency, auto-sync.
   */
  updateConnector: async (
    connectorId: string,
    data: UpdateConnectorRequest
  ): Promise<UserConnector> => {
    const res = await apiClient.patch<{ connector: UserConnector }>(
      `/v1/connectors/${connectorId}`,
      data
    );
    return res.connector;
  },

  oauthCallback: async (code: string, state: string): Promise<OAuthCallbackResponse> => {
    return apiClient.post<OAuthCallbackResponse>('/v1/connectors/callback', { code, state });
  },

  unlinkConnector: async (connectorId: string): Promise<void> => {
    await apiClient.delete(`/v1/connectors/${connectorId}`);
  },

  triggerSync: async (connectorId: string): Promise<SyncTriggerResponse> => {
    return apiClient.post<SyncTriggerResponse>(`/v1/connectors/${connectorId}/sync`);
  },
};
