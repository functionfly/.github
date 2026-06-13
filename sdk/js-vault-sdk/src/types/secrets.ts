/** Secret kinds. */
export enum SecretType {
  APIKey = "api_key",
  OAuthToken = "oauth_token",
  Password = "password",
  Certificate = "certificate",
}

/**
 * EncryptedData is the zero-knowledge ciphertext payload. All fields
 * are base64-encoded bytes produced by the caller's encryption layer.
 */
export interface EncryptedData {
  ciphertext: string;
  iv: string;
  salt: string;
  tag: string;
  /** 1 = PBKDF2 (legacy), 2 = Argon2id (default for new secrets). */
  keyVersion: number;
}

export interface Secret {
  id: string;
  tenantId: string;
  name: string;
  description?: string;
  secretType: SecretType;
  encryptedData: EncryptedData;
  accessCount: number;
  createdAt: string;
  updatedAt: string;
  expiresAt?: string;
}

export interface SecretCreate {
  name: string;
  description?: string;
  secretType: SecretType;
  encryptedData: EncryptedData;
  scopes?: string[];
  metadata?: Record<string, unknown>;
}

export interface SecretUpdate {
  name?: string;
  description?: string;
  scopes?: string[];
}

export interface SecretRotate {
  encryptedData: EncryptedData;
  reason?: string;
}

export interface SecretListOptions {
  limit?: number;
  offset?: number;
  secretType?: SecretType;
}

export interface SecretList {
  secrets: Secret[];
  total: number;
  limit: number;
  offset: number;
}
