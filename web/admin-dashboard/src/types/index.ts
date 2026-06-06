export interface AdminUser {
  id: string;
  email: string;
  username?: string;
  name?: string;
  role: string;
  permissions?: string[];
  mfa_enabled: boolean;
  mfa_verified_at?: string;
  created_at: string;
  updated_at: string;
  tenant_id?: string;
  tenant_name?: string;
  plan?: string;
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
  user_email?: string;
  event_type?: string;
  action: string;
  resource_type: string;
  resource_id?: string;
  ip_address?: string;
  user_agent?: string;
  timestamp: string;
  created_at: string;
  success: boolean;
  admin_context?: Record<string, any>;
  risk_score?: number;
  session_id?: string;
  before_state?: Record<string, any>;
  after_state?: Record<string, any>;
  tenant_name?: string;
  metadata?: Record<string, any>;
}

/**
 * Auth audit endpoints return a slightly different shape than the
 * general audit log. This is the row type the /admin/auth-audit endpoint
 * returns. The resource/state fields are optional because not every
 * auth event has them (e.g. a login has no resource_id).
 */
export interface AuthAuditEvent {
  id: string;
  event_type: string;
  action: string;
  user_email?: string;
  actor_email?: string;
  ip_address?: string;
  user_agent?: string;
  tenant_id?: string;
  tenant_name?: string;
  resource_type?: string;
  resource_id?: string;
  success: boolean;
  risk_score?: number;
  created_at: string;
  timestamp?: string;
  before_state?: Record<string, unknown>;
  after_state?: Record<string, unknown>;
  metadata?: Record<string, unknown>;
}

export interface AuditLogFilters {
  event_type?: string;
  success?: boolean;
  start_date?: string;
  end_date?: string;
  search?: string;
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
  cidr?: string | number;
}

export type IPAllowlistEntry = IPWhitelistEntry;

export interface IPAllowlist {
  id: string;
  tenant_id: string;
  name: string;
  description: string;
  default_policy: 'allow' | 'deny';
  mfa_bypass: boolean;
  enabled: boolean;
  created_at: string;
  updated_at: string;
  entries: IPWhitelistEntry[];
}

export interface IPAllowlistCreateInput {
  tenant_id: string;
  name: string;
  description: string;
  default_policy: 'allow' | 'deny';
  mfa_bypass: boolean;
}

export interface IPEntryCreateInput {
  ip_address: string;
  cidr?: string;
  description: string;
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

export interface SIEMConfig {
  id: string;
  tenant_id: string;
  name: string;
  destination_type: 'splunk' | 'elastic' | 'qradar' | 'syslog' | 'webhook' | 'cloudwatch' | 'azure' | 'gcp';
  format: 'json' | 'cef' | 'leef';
  endpoint: string;
  enabled: boolean;
  api_key?: string;
  config: Record<string, string>;
  last_export_at?: string;
  created_at: string;
  updated_at: string;
}

export interface SIEMConfigCreateInput {
  tenant_id: string;
  name: string;
  destination_type: string;
  format: string;
  endpoint: string;
  api_key?: string;
  config: Record<string, string>;
}

export type SIEMDestination = string;

export type SIEMFormat = string;

export interface SIEMTestResult {
  success: boolean;
  message: string;
  latency_ms?: number;
}

export interface SAMLConfig {
  id: string;
  tenant_id?: string;
  entity_id: string;
  sso_url: string;
  slo_url?: string;
  x509_cert?: string;
  certificate?: string;
  enabled: boolean;
}

export interface SAMLMetadata {
  entity_id: string;
  sso_url: string;
  slo_url?: string;
  x509_cert: string;
  metadata_xml?: string;
  acs_url?: string;
}

export type MFAPolicy = 'optional' | 'required';

export interface SessionPolicy {
  timeout_minutes: number;
  idle_timeout_minutes: number;
  max_concurrent_sessions: number;
  enforce_device_fingerprint: boolean;
  max_duration_minutes?: number;
}

export interface ActiveSession {
  id: string;
  user_id: string;
  user_email?: string;
  ip_address: string;
  user_agent: string;
  created_at: string;
  last_activity_at: string;
  expires_at: string;
  mfa_verified: boolean;
}
