// Admin Dashboard Types

export interface AdminUser {
  id: string;
  email: string;
  username?: string;
  name?: string;
  role: string; // super_admin | admin | moderator | user | etc.
  permissions?: string[];
  mfa_enabled: boolean;
  created_at: string;
  updated_at: string;
  tenant_id?: string;
  tenant_name?: string;
  plan?: string; // tenant plan
}

export interface AdminSession {
  id: string;
  user_id: string;
  session_token_hash: string;
  access_token?: string;
  ip_address: string;
  user_agent: string;
  device_fingerprint?: string;
  mfa_verified_at?: string;
  created_at: string;
  last_activity_at: string;
  expires_at: string;
  revoked_at?: string;
}

export interface AdminAuthState {
  user: AdminUser | null;
  session: AdminSession | null;
  isAuthenticated: boolean;
  mfaVerified: boolean;
  lastActivity: number;
  deviceFingerprint: string | null;
  isIpAllowed: boolean;
  ipCheckReason: string | null;
}

export interface AdminIPCheckResponse {
  allowed: boolean;
  reason?: string;
  source_ip?: string;
}

export interface AuditEvent {
  id: string;
  actor_user_id?: string;
  actor_email?: string;
  action: string;
  resource_type: string;
  resource_id?: string;
  ip_address?: string;
  user_agent?: string;
  timestamp: string;
  success: boolean;
  admin_context?: Record<string, any>;
  risk_score?: number;
  session_id?: string;
  before_state?: Record<string, any>;
  after_state?: Record<string, any>;
}

export interface IPWhitelistEntry {
  id: string;
  ip_address?: string;
  ip_range?: string;
  description: string;
  enabled: boolean;
  created_by: string;
  created_at: string;
  expires_at?: string;
  last_used_at?: string;
  use_count: number;
}

export interface Tenant {
  id: string;
  name: string;
  plan: string;
  status: 'active' | 'suspended';
  created_at: string;
  updated_at: string;
}

export interface AdminAPIResponse<T> {
  data: T;
  success: boolean;
  error?: string;
  timestamp: string;
}

export interface HMACSignature {
  timestamp: string;
  signature: string;
}

export interface AdminAuthLoginResponse {
  token: string;
  refresh_token?: string;
  user: AdminUser;
}

export interface AdminSessionBootstrapResponse {
  session: AdminSession;
  user: AdminUser;
}
