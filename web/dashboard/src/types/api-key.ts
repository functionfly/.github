// API Key types matching backend Go types

export type APIKeyType = 'platform' | 'function' | 'agent' | 'environment' | 'oauth' | 'trust' | 'micropython' | 'runtime';

export type Permission = 'read' | 'write' | 'execute' | 'admin';

export type ResourceType = 'function' | 'app' | 'tenant' | 'registry' | 'deployment' | 'secret';

export type RotationReason = 'manual' | 'automatic' | 'compromised';

export interface RateLimitConfig {
  rpm: number;
  rph: number;
  rpd: number;
}

export interface APIKeyPermission {
  id: string;
  api_key_id: string;
  permission: Permission;
  resource_type: ResourceType;
  resource_id: string;
  created_at: string;
}

export interface APIKeyEnvironment {
  id: string;
  api_key_id: string;
  environment_id: string;
  environment_name: string;
  created_at: string;
}

/** Platform environment that can be linked to an API key (from GET /api-keys/environments/available) */
export interface AvailableEnvironment {
  id: string;
  name: string;
}

export interface APIKeyRotation {
  id: string;
  api_key_id: string;
  rotated_at: string;
  expires_at?: string;
  created_by?: string;
  key_hash: string;
  rotation_reason: RotationReason;
  metadata?: Record<string, unknown>;
}

export interface APIKey {
  id: string;
  name: string;
  description?: string;
  key_type: APIKeyType;
  key_id?: string; // Public key identifier (used by Trust API keys)
  key_prefix: string;
  // Alias for backwards compatibility
  prefix?: string;
  expires_at?: string;
  last_rotated_at: string;
  rotation_frequency_days: number;
  // Flat rate limit fields from backend
  rate_limit_rpm: number;
  rate_limit_rph: number;
  rate_limit_rpd: number;
  // Nested rate limit for backwards compatibility
  rate_limit?: RateLimitConfig;
  is_active: boolean;
  created_at: string;
  updated_at?: string;
  last_used_at?: string;
  permissions?: APIKeyPermission[];
  environments?: APIKeyEnvironment[];
  // Trust API specific fields
  partner_id?: string;
  scopes?: Record<string, boolean>;
  is_revoked?: boolean;
  revoked_at?: string;
  revoked_reason?: string;
  use_count?: number;
  billing_budget_cents?: number;
  is_high_value?: boolean;
  cost_center?: string;
  created_by?: string;
}

export interface APIKeyCreateResponse extends APIKey {
  plaintext: string; // Only returned on creation
}

// Request types
export interface PermissionGrant {
  permission: Permission;
  resource_type: ResourceType;
  resource_id: string;
}

// CreateAPIKeyRequest: matches backend apikey.CreateAPIKeyRequest shape.
// IMPORTANT: rate_limit must be sent as a NESTED object (not flat fields) so
// that the backend's req.RateLimit pointer is non-nil and the per-key limits
// are persisted. Sending flat `rate_limit_rpm` etc. was silently dropped.
export interface CreateAPIKeyRequest {
  name: string;
  description?: string;
  key_type: APIKeyType;
  permissions?: PermissionGrant[];
  environments?: string[];
  expires_at?: string;
  rotation_frequency_days?: number;
  rate_limit?: RateLimitConfig;
  metadata?: Record<string, unknown>;
}

// UpdateAPIKeyRequest: backend uses flat rate_limit_* fields on the update path
// (see internal/api/handlers/apikeys/update.go UpdateAPIKeyRequest). Keep flat
// fields here for update; create uses the nested object above.
export interface UpdateAPIKeyRequest {
  name?: string;
  description?: string;
  expires_at?: string;
  rotation_frequency_days?: number;
  rate_limit_rpm?: number;
  rate_limit_rph?: number;
  rate_limit_rpd?: number;
  is_active?: boolean;
  metadata?: Record<string, unknown>;
}

export interface RotateAPIKeyRequest {
  reason?: RotationReason;
  expires_in_days?: number;
  metadata?: Record<string, unknown>;
}

export interface AddPermissionRequest {
  permission: Permission;
  resource_type: ResourceType;
  resource_id: string;
}

export interface AddEnvironmentRequest {
  environment_id: string;
  environment_name?: string;
}

// Filter types
export interface APIKeyFilters {
  key_type?: APIKeyType;
  is_active?: boolean;
  expires_before?: string;
  expires_after?: string;
  search?: string;
  page?: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface APIKeyListResponse extends PaginatedResponse<APIKey> {}

// Helper functions
export const API_KEY_TYPE_LABELS: Record<APIKeyType, string> = {
  platform: 'Platform',
  function: 'Function',
  agent: 'Agent',
  environment: 'Environment',
  oauth: 'OAuth',
  trust: 'Trust API',
  micropython: 'MicroPython Runtime',
  runtime: 'Runtime',
};

export const API_KEY_TYPE_DESCRIPTIONS: Record<APIKeyType, string> = {
  platform: 'Full access to platform APIs',
  function: 'Access to function execution',
  agent: 'Access for AI agents',
  environment: 'Environment-specific access',
  oauth: 'OAuth-based access',
  trust: 'Trust API partner access for trust scores, verification, and reports',
  micropython: 'MicroPython runtime access for Enterprise accounts (Firecracker isolation)',
  runtime: 'Authenticate runtime /execute endpoints (bun, deno, kotlin, ruby, nodejs, wasmedge, prism)',
};

export const PERMISSION_LABELS: Record<Permission, string> = {
  read: 'Read',
  write: 'Write',
  execute: 'Execute',
  admin: 'Admin',
};

export const RESOURCE_TYPE_LABELS: Record<ResourceType, string> = {
  function: 'Function',
  app: 'App',
  tenant: 'Tenant',
  registry: 'Registry',
  deployment: 'Deployment',
  secret: 'Secret',
};

export const ROTATION_REASON_LABELS: Record<RotationReason, string> = {
  manual: 'Manual',
  automatic: 'Automatic',
  compromised: 'Compromised',
};

export const DEFAULT_RATE_LIMIT: RateLimitConfig = {
  rpm: 1000,
  rph: 60000,
  rpd: 1000000,
};

export const DEFAULT_ROTATION_DAYS = 90;
