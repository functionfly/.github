/**
 * Vault Enterprise Types — Phases 1.1-5 + 6
 * Mirrors the Go types in:
 *   - internal/api/handlers/vault/types.go
 *   - internal/storage/vault/enterprise_models.go
 *   - internal/storage/vault/phase1_models.go
 *   - internal/storage/vault/dynamic_models.go
 *
 * Plan gating lives in src/lib/vaultPlans.ts.
 */

import type { SecretType } from './vault';

// ============================================================================
// Plans
// ============================================================================

export type VaultPlan = 'free' | 'pro' | 'team' | 'enterprise';

export interface PlanLimits {
  /** Lifetime cap on stored secrets */
  maxSecrets: number;
  /** Dynamic credentials minted in a rolling 30-day window */
  maxDynamicCreds: number;
  /** Active runtime tokens per secret */
  tokensPerSecret: number;
  /** Audit log exports per 24h */
  auditExportsPerDay: number;
  /** Storage backends available for dynamic creds */
  dynamicBackends: ('postgres' | 'mysql')[];
  /** Feature flags */
  features: {
    mfa: boolean;
    ipAllowlist: boolean;
    expiration: boolean;
    breakGlass: boolean;
    escrow: boolean;
    rbac: boolean;
    namespaces: boolean;
    shares: boolean;
    sso: boolean;
    siemWebhooks: boolean;
    auditExport: boolean;
    cacheStats: boolean;
    quotaWidget: boolean;
    haStatus: boolean;
    dependencyGraph: boolean;
    expirationDashboard: boolean;
    tokenMonitor: boolean;
    rotationSchedules: boolean;
  };
}

// ============================================================================
// Phase 1.1: Vault MFA config
// ============================================================================

export interface VaultMFAConfig {
  tenant_id: string;
  mfa_required: boolean;
  mfa_method: 'totp' | 'webauthn' | 'both';
  enforce_for_tokens: boolean;
  enforce_for_api: boolean;
  mfa_session_ttl_seconds: number;
  updated_at: string;
}

export interface UpdateVaultMFARequest {
  mfa_required?: boolean;
  mfa_method?: 'totp' | 'webauthn' | 'both';
  enforce_for_tokens?: boolean;
  enforce_for_api?: boolean;
  mfa_session_ttl_seconds?: number;
}

// ============================================================================
// Phase 1.2: Token IP allowlist
// ============================================================================

export interface AccessToken {
  id: string;
  secret_id: string;
  name?: string;
  expires_at: string;
  is_revoked: boolean;
  revoked_at?: string;
  revoked_reason?: string;
  last_used_at?: string;
  use_count: number;
  created_at: string;
  allowed_ips?: string[];
  denied_ips?: string[];
  ip_restriction_enabled?: boolean;
}

export interface UpdateTokenIPPolicyRequest {
  allowed_ips: string[];
  denied_ips: string[];
  enabled: boolean;
}

// ============================================================================
// Phase 1.3: Expiration
// ============================================================================

export type SecretStatus = 'active' | 'expiring_soon' | 'expired' | 'revoked';

export interface SetSecretExpirationRequest {
  expires_at?: string;
  expire_after_days?: number;
}

// ============================================================================
// Phase 1.4: Break-glass
// ============================================================================

export type BreakGlassStatus = 'pending' | 'approved' | 'denied' | 'expired' | 'revoked';

export interface BreakGlassRequest {
  id: string;
  tenant_id: string;
  requested_by: string;
  approved_by?: string;
  reason: string;
  status: BreakGlassStatus;
  duration_minutes: number;
  expires_at: string;
  approved_at?: string;
  revoked_at?: string;
  created_at: string;
}

export interface BreakGlassRequestBody {
  reason: string;
  duration_minutes?: number;
}

export interface BreakGlassConfig {
  tenant_id: string;
  max_duration_minutes: number;
  required_approver_count: number;
  enabled: boolean;
  updated_at: string;
}

export interface EscrowStatus {
  tenant_id: string;
  enabled: boolean;
  kdf_method: string;
  blob_key_version: number;
  created_at?: string;
  updated_at: string;
  last_recovered_at?: string;
}

export interface EnableEscrowRequest {
  security_question_hashes: string[];
  kdf_salt: string; // base64
  encrypted_blob: string; // base64
  blob_iv: string; // base64
  blob_auth_tag: string; // base64
}

// ============================================================================
// Phase 2: Dynamic credentials
// ============================================================================

export type DynamicDBType = 'postgres' | 'mysql';

export interface DynamicSecretTarget {
  id: string;
  tenant_id: string;
  name: string;
  description?: string;
  db_type: DynamicDBType;
  host: string;
  port: number;
  database_name: string;
  admin_username: string;
  ssl_mode: string;
  allowed_roles?: string[];
  default_ttl_seconds: number;
  max_ttl_seconds: number;
  status: string;
  created_at: string;
  updated_at: string;
  last_used_at?: string;
}

export interface CreateTargetRequest {
  name: string;
  description?: string;
  db_type: DynamicDBType;
  host: string;
  port: number;
  database_name: string;
  admin_username: string;
  admin_password: string;
  ssl_mode?: string;
  allowed_roles?: string[];
  default_ttl_seconds?: number;
  max_ttl_seconds?: number;
}

export interface DynamicCredential {
  id: string;
  tenant_id: string;
  target_id: string;
  name: string;
  description?: string;
  role_template?: string;
  ttl_seconds: number;
  max_ttl_seconds: number;
  status: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface CreateCredentialRequest {
  target_id: string;
  name: string;
  description?: string;
  role_template?: string;
  ttl_seconds?: number;
  max_ttl_seconds?: number;
}

export interface GeneratedCredential {
  lease_id: string;
  username: string;
  password: string;
  host: string;
  port: number;
  database: string;
  expires_at: string;
  credential: DynamicCredential;
  target: DynamicSecretTarget;
}

export interface DynamicCredentialLease {
  id: string;
  lease_id: string;
  credential_id: string;
  target_id: string;
  tenant_id: string;
  db_username: string;
  expires_at: string;
  renewed_at?: string;
  revoked_at?: string;
  revocation_reason?: string;
  last_used_at?: string;
  use_count: number;
  issued_to?: string;
  issued_ip?: string;
  created_at: string;
}

// ============================================================================
// Phase 4.1: RBAC
// ============================================================================

export interface VaultRole {
  id: string;
  tenant_id: string;
  name: string;
  description?: string;
  permissions: Record<string, boolean | string | string[]>;
  is_builtin: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateRoleRequest {
  name: string;
  description?: string;
  permissions: Record<string, boolean | string | string[]>;
}

export interface UpdateRoleRequest {
  description?: string;
  permissions?: Record<string, boolean | string | string[]>;
}

export interface VaultRoleAssignment {
  id: string;
  tenant_id: string;
  role_id: string;
  user_id?: string;
  scope: string;
  created_by: string;
  created_at: string;
}

export interface AssignRoleRequest {
  user_id: string;
  scope?: string;
}

// ============================================================================
// Phase 4.2: Audit + SIEM
// ============================================================================

export type AuditExportFormat = 'json' | 'csv' | 'cef';

export interface AuditExportResult {
  format: AuditExportFormat;
  row_count: number;
  generated_at: string;
  hmac_sha256: string;
  /** Body is the raw export blob; provided as a Blob on the wire. */
  body: Blob;
}

export interface VaultSIEMWebhook {
  id: string;
  tenant_id: string;
  name: string;
  url: string;
  format: 'json' | 'cef';
  enabled: boolean;
  last_delivery_at?: string;
  last_delivery_status?: number;
  last_delivery_error?: string;
  created_at: string;
  /** Returned once on create — used to verify X-Signature. */
  secret_hmac?: string;
}

export interface CreateSIEMWebhookRequest {
  name: string;
  url: string;
  format?: 'json' | 'cef';
}

// ============================================================================
// Phase 4.3: Namespaces
// ============================================================================

export interface VaultNamespace {
  id: string;
  tenant_id: string;
  path: string;
  description?: string;
  parent_id?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface CreateNamespaceRequest {
  path: string;
  description?: string;
  parent_id?: string;
}

// ============================================================================
// Phase 4.4: Cross-tenant shares
// ============================================================================

export type SharePermission = 'read' | 'read-write';

export interface VaultShare {
  id: string;
  secret_id: string;
  source_tenant_id: string;
  granted_to_tenant_id: string;
  granted_by_user: string;
  permissions: SharePermission;
  expires_at?: string;
  revoked_at?: string;
  created_at: string;
}

export interface ShareSecretRequest {
  grantee_tenant_id: string;
  permissions?: SharePermission;
  expires_at?: string;
}

// ============================================================================
// Phase 4.5: SSO
// ============================================================================

export interface VaultSSOConfig {
  tenant_id: string;
  enabled: boolean;
  saml_metadata_url?: string;
  saml_entity_id?: string;
  saml_sso_url?: string;
  saml_slo_url?: string;
  jit_provisioning_enabled: boolean;
  attribute_role_mapping: Record<string, string>;
  updated_at: string;
}

export interface UpdateSSORequest {
  enabled?: boolean;
  saml_metadata_url?: string;
  saml_entity_id?: string;
  saml_sso_url?: string;
  saml_slo_url?: string;
  saml_x509_cert?: string;
  jit_provisioning_enabled?: boolean;
  attribute_role_mapping?: Record<string, string>;
}

// ============================================================================
// Phase 5: Cache + Leader
// ============================================================================

export interface CacheStats {
  meta_keys: number;
  token_keys: number;
  enabled: boolean;
}

export interface LeaderStatus {
  namespace: string;
  is_leader: boolean;
  holds: boolean;
  last_renew_at?: string;
  ttl: number;
  renew_interval: number;
}

// ============================================================================
// Phase 5.2: Quota
// ============================================================================

export interface QuotaDecision {
  allowed: boolean;
  limit: number;
  current: number;
  remaining: number;
  reset?: string;
  headers?: Record<string, string>;
}

export interface QuotaUsage {
  resource: string;
  limit: number;
  current: number;
  remaining: number;
  percentage_used: number;
  window?: string;
  resets_at?: string;
}

// ============================================================================
// Phase 6: Secret Rotation Schedules
// ============================================================================

export interface RotationSchedule {
  id: string;
  tenant_id: string;
  secret_id: string;
  secret_name?: string;
  rotation_type: 'automatic' | 'scheduled' | 'manual';
  enabled: boolean;
  auto_rotate_interval?: number; // days
  scheduled_at?: string; // ISO date for scheduled rotation
  next_rotation_at?: string;
  last_rotated_at?: string;
  grace_period_hours: number;
  notify_stakeholders: boolean;
  require_approval: boolean;
  status: 'active' | 'paused' | 'pending' | 'cancelled' | 'failed';
  created_at: string;
  updated_at: string;
  created_by: string;
}

export interface RotationSchedulesResponse {
  schedules: RotationSchedule[];
  total: number;
  limit: number;
  offset: number;
}

export interface SetAutoRotationRequest {
  secret_id: string;
  enabled: boolean;
  auto_rotate_interval?: number;
  grace_period_hours?: number;
  notify_stakeholders?: boolean;
  require_approval?: boolean;
}

export interface CreateScheduledRotationRequest {
  secret_id: string;
  scheduled_at: string;
  grace_period_hours?: number;
  notify_stakeholders?: boolean;
  require_approval?: boolean;
}

export interface CancelRotationRequest {
  schedule_id: string;
  reason: string;
}

// ============================================================================
// Helper re-exports
// ============================================================================

export type { SecretType };
