export interface AIProviderKey {
  id: string;
  provider: string;
  key_last4: string;
  status: 'active' | 'degraded' | 'expired' | 'revoked';
  health_message?: string;
  last_health_check?: string;
  last_used_at?: string;
  created_at: string;
}

export interface SupportedProvider {
  id: string;
  name: string;
  description: string;
  key_format: string;
  key_prefix?: string;
  is_meta_provider?: boolean;
}

export interface ConnectAIKeyRequest {
  provider: string;
  apiKey: string;
}

export interface ConnectAIKeyResponse {
  key: AIProviderKey;
}

export interface ListAIKeysResponse {
  keys: AIProviderKey[];
}

export interface ListSupportedProvidersResponse {
  providers: SupportedProvider[];
}

export interface TestAIKeyResponse {
  success: boolean;
  message?: string;
}

export interface RotateAIKeyRequest {
  apiKey: string;
}

export interface RotateAIKeyResponse {
  key: AIProviderKey;
}
