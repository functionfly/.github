/**
 * Vault Types - Secret management type definitions
 * Mirrors internal/api/handlers/vault/types.go and internal/storage/vault/types.go
 */

/** Secret type enumeration */
export type SecretType = 'api_key' | 'oauth_token' | 'password' | 'certificate';

/** Secret type constants */
export const SecretType = {
  API_KEY: 'api_key' as SecretType,
  OAUTH_TOKEN: 'oauth_token' as SecretType,
  PASSWORD: 'password' as SecretType,
  CERTIFICATE: 'certificate' as SecretType,
};

/** Audit action enumeration */
export type AuditAction = 'create' | 'read' | 'update' | 'delete' | 'use' | 'revoke';

/** Actor type enumeration */
export type ActorType = 'user' | 'token' | 'system' | 'api_key';

/** Encrypted data payload - matches server-side structure */
export interface EncryptedDataPayload {
  ciphertext: string;  // base64 encoded encrypted data
  iv: string;          // base64 encoded initialization vector
  salt: string;        // base64 encoded PBKDF2 salt
  tag: string;         // base64 encoded authentication tag
  key_version: number; // encryption key version (1=passphrase, 2=KMS, 3=HSM)
}

/** Secret metadata (for list views without encrypted data) */
export interface SecretMetadata {
  id: string;
  name: string;
  description?: string;
  secret_type: SecretType;
  scopes?: string[];
  metadata?: Record<string, unknown>;
  last_accessed_at?: string;
  access_count: number;
  created_at: string;
  updated_at: string;
}

/** Full secret with encrypted data */
export interface Secret extends SecretMetadata {
  tenant_id: string;
  encrypted_data: EncryptedDataPayload;
}

/** Request to create a new secret */
export interface CreateSecretRequest {
  name: string;
  description?: string;
  secret_type: SecretType;
  encrypted_data: EncryptedDataPayload;
  scopes?: string[];
  metadata?: Record<string, unknown>;
}

/** Request to update a secret (partial update) */
export interface UpdateSecretRequest {
  name?: string;
  description?: string;
  scopes?: string[];
  metadata?: Record<string, unknown>;
}

/** Access token information (without the actual token) */
export interface AccessToken {
  id: string;
  secret_id: string;
  name?: string;
  scopes?: string[];
  expires_at: string;
  is_revoked: boolean;
  revoked_at?: string;
  revoked_reason?: string;
  last_used_at?: string;
  use_count: number;
  created_at: string;
}

/** Request to generate a new access token */
export interface GenerateTokenRequest {
  scopes?: string[];
  expires_in_hours: number; // min=1, max=8760 (1 year)
  name?: string;
}

/** Response with generated token (token shown only once) */
export interface GenerateTokenResponse {
  token_id: string;
  token: string; // plaintext token, shown once
  secret_id: string;
  name?: string;
  expires_at: string;
  scopes?: string[];
  created_at: string;
}

/** Audit log entry */
export interface AuditLogEntry {
  id: string;
  secret_id?: string;
  action: AuditAction;
  actor_id: string;
  actor_type: ActorType;
  request_id?: string;
  ip_address?: string;
  user_agent?: string;
  metadata?: Record<string, unknown>;
  success: boolean;
  error_message?: string;
  created_at: string;
}

/** List secrets response */
export interface ListSecretsResponse {
  secrets: SecretMetadata[];
  total: number;
  limit: number;
  offset: number;
}

/** List tokens response */
export interface ListTokensResponse {
  tokens: AccessToken[];
  total: number;
}

/** List audit log response */
export interface ListAuditLogResponse {
  entries: AuditLogEntry[];
  total: number;
  limit: number;
  offset: number;
}
