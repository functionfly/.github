/**
 * @functionfly/vault
 *
 * Official Node.js SDK for the FunctionFly zero-knowledge secrets vault.
 *
 * Quick start:
 *
 *   import { VaultClient, SecretType } from "@functionfly/vault";
 *
 *   const client = new VaultClient({
 *     baseUrl: "https://api.functionfly.com",
 *     token: "fnly_xxx",
 *   });
 *
 *   // The caller performs encryption (zero-knowledge); the SDK ships
 *   // ciphertext + IV + salt + auth tag verbatim.
 *   const secret = await client.secrets.create({
 *     name: "STRIPE_API_KEY",
 *     secretType: SecretType.APIKey,
 *     encryptedData: {
 *       ciphertext: ctBase64,
 *       iv: ivBase64,
 *       salt: saltBase64,
 *       tag: tagBase64,
 *       keyVersion: 2, // 1=PBKDF2, 2=Argon2id
 *     },
 *   });
 */

export { VaultClient } from "./client";
export { VaultAPIError } from "./errors";
export {
  SecretType,
  type EncryptedData,
  type Secret,
  type SecretCreate,
  type SecretUpdate,
  type SecretRotate,
  type SecretList,
  type SecretListOptions,
} from "./types/secrets";
export {
  type Token,
  type TokenInfo,
  type TokenCreate,
  type TokenList,
} from "./types/tokens";
export {
  DynamicDBType,
  type DynamicTarget,
  type DynamicTargetCreate,
  type DynamicTargetsList,
  type DynamicCredential,
  type DynamicCredentialCreate,
  type GeneratedCredential,
  type GenerateOptions,
  type RenewOptions,
} from "./types/dynamic";
export {
  type AuditEntry,
  type AuditList,
  type AuditListOptions,
} from "./types/audit";
export type { VaultClientOptions } from "./client";
