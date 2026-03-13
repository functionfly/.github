/**
 * Vault Crypto - WebCrypto utilities for client-side encryption
 * Uses PBKDF2 for key derivation and AES-256-GCM for encryption.
 *
 * Key versions:
 * - 1: 100_000 PBKDF2 iterations (legacy; existing secrets remain decryptable)
 * - 2: 600_000 PBKDF2 iterations (OWASP-style; used for new secrets)
 */

import type { EncryptedDataPayload } from "@/types/vault";

/**
 * Encrypted data structure returned from encrypt()
 */
export interface EncryptedData {
  ciphertext: string; // base64 encoded
  iv: string;         // base64 encoded
  salt: string;       // base64 encoded
  tag: string;        // base64 encoded
  keyVersion: number;
}

/**
 * Helper function to encode ArrayBuffer to base64 string
 */
export function arrayBufferToBase64(buffer: ArrayBuffer | Uint8Array): string {
  const bytes = buffer instanceof Uint8Array ? buffer : new Uint8Array(buffer);
  let binary = "";
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary);
}

/**
 * Helper function to decode base64 string to Uint8Array
 */
export function base64ToUint8Array(base64: string): Uint8Array {
  const binary = atob(base64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes;
}

/**
 * VaultCrypto - Static class for cryptographic operations
 */
export class VaultCrypto {
  // Algorithm constants
  private static readonly PBKDF2_ITERATIONS_V1 = 100_000; // legacy (existing secrets)
  private static readonly PBKDF2_ITERATIONS_V2 = 600_000; // OWASP-style for new secrets
  private static readonly SALT_LENGTH = 16; // 128 bits
  private static readonly IV_LENGTH = 12;   // 96 bits for AES-GCM
  private static readonly KEY_LENGTH = 256; // AES-256
  private static readonly KEY_VERSION = 2;  // 2 = 600k iterations; 1 = 100k (legacy)

  /**
   * Generate a random salt for PBKDF2
   * @returns Uint8Array of 16 random bytes
   */
  static generateSalt(): Uint8Array {
    return crypto.getRandomValues(new Uint8Array(this.SALT_LENGTH));
  }

  /**
   * Generate a random IV for AES-GCM
   * @returns Uint8Array of 12 random bytes (AES-GCM standard)
   */
  static generateIV(): Uint8Array {
    return crypto.getRandomValues(new Uint8Array(this.IV_LENGTH));
  }

  /**
   * Derive a cryptographic key from a passphrase using PBKDF2
   * @param passphrase - User's passphrase
   * @param salt - Random salt (should be generated via generateSalt)
   * @param iterations - PBKDF2 iterations (default: KEY_VERSION 2 = 600k)
   * @returns Promise<CryptoKey> - Derived AES-GCM key
   */
  static async deriveKey(
    passphrase: string,
    salt: Uint8Array,
    iterations: number = this.PBKDF2_ITERATIONS_V2
  ): Promise<CryptoKey> {
    const encoder = new TextEncoder();
    const passphraseData = encoder.encode(passphrase);

    const baseKey = await crypto.subtle.importKey(
      "raw",
      passphraseData,
      { name: "PBKDF2" },
      false,
      ["deriveBits", "deriveKey"]
    );

    const saltBuffer = salt.buffer.slice(salt.byteOffset, salt.byteOffset + salt.byteLength) as ArrayBuffer;

    const key = await crypto.subtle.deriveKey(
      {
        name: "PBKDF2",
        salt: saltBuffer,
        iterations,
        hash: "SHA-256",
      },
      baseKey,
      {
        name: "AES-GCM",
        length: this.KEY_LENGTH,
      },
      false,
      ["encrypt", "decrypt"]
    );

    return key;
  }

  /**
   * Encrypt plaintext using AES-256-GCM
   * @param plaintext - The string to encrypt
   * @param key - The CryptoKey for encryption (derived via deriveKey)
   * @returns Promise<EncryptedData> - Encrypted data with all necessary fields
   */
  static async encrypt(plaintext: string, key: CryptoKey): Promise<EncryptedData> {
    // Generate random IV
    const iv = this.generateIV();

    // Encode plaintext as UTF-8
    const encoder = new TextEncoder();
    const plaintextData = encoder.encode(plaintext);

    // Convert IV to ArrayBuffer
    const ivBuffer = iv.buffer.slice(iv.byteOffset, iv.byteOffset + iv.byteLength) as ArrayBuffer;

    // Encrypt the data
    const encryptedBuffer = await crypto.subtle.encrypt(
      {
        name: "AES-GCM",
        iv: ivBuffer,
      },
      key,
      plaintextData
    );

    // Extract ciphertext and authentication tag
    // AES-GCM appends the 16-byte tag to the ciphertext
    const encryptedBytes = new Uint8Array(encryptedBuffer);
    const tagLength = 16; // 128-bit authentication tag
    const ciphertextBytes = encryptedBytes.slice(0, encryptedBytes.length - tagLength);
    const tagBytes = encryptedBytes.slice(encryptedBytes.length - tagLength);

    return {
      ciphertext: arrayBufferToBase64(ciphertextBytes),
      iv: arrayBufferToBase64(iv),
      salt: "", // Salt is not known at this point; set separately if needed
      tag: arrayBufferToBase64(tagBytes),
      keyVersion: this.KEY_VERSION,
    };
  }

  /**
   * Decrypt ciphertext using AES-256-GCM
   * @param encryptedData - The encrypted data structure from encrypt()
   * @param key - The CryptoKey for decryption (derived via deriveKey)
   * @returns Promise<string> - Decrypted plaintext
   */
  static async decrypt(encryptedData: EncryptedData, key: CryptoKey): Promise<string> {
    // Decode base64 strings to Uint8Arrays
    const ciphertext = base64ToUint8Array(encryptedData.ciphertext);
    const iv = base64ToUint8Array(encryptedData.iv);
    const tag = base64ToUint8Array(encryptedData.tag);

    // Combine ciphertext and tag for decryption (AES-GCM expects this format)
    const combined = new Uint8Array(ciphertext.length + tag.length);
    combined.set(ciphertext, 0);
    combined.set(tag, ciphertext.length);

    // Convert IV to ArrayBuffer
    const ivBuffer = iv.buffer.slice(iv.byteOffset, iv.byteOffset + iv.byteLength) as ArrayBuffer;

    // Decrypt the data
    const decryptedBuffer = await crypto.subtle.decrypt(
      {
        name: "AES-GCM",
        iv: ivBuffer,
      },
      key,
      combined
    );

    // Decode UTF-8 to string
    const decoder = new TextDecoder();
    return decoder.decode(decryptedBuffer);
  }

  /**
   * Convenience method to encrypt with passphrase (combines deriveKey + encrypt)
   * @param plaintext - The string to encrypt
   * @param passphrase - User's passphrase
   * @returns Promise<EncryptedData> - Encrypted data including salt
   */
  static async encryptWithPassphrase(
    plaintext: string,
    passphrase: string
  ): Promise<EncryptedData> {
    const salt = this.generateSalt();
    const key = await this.deriveKey(passphrase, salt, this.PBKDF2_ITERATIONS_V2);
    const encrypted = await this.encrypt(plaintext, key);
    encrypted.salt = arrayBufferToBase64(salt);
    encrypted.keyVersion = this.KEY_VERSION; // 2 = 600k iterations
    return encrypted;
  }

  /**
   * Convenience method to decrypt with passphrase (combines deriveKey + decrypt)
   * @param encryptedData - The encrypted data structure
   * @param passphrase - User's passphrase
   * @returns Promise<string> - Decrypted plaintext
   */
  static async decryptWithPassphrase(
    encryptedData: EncryptedData,
    passphrase: string
  ): Promise<string> {
    const salt = base64ToUint8Array(encryptedData.salt);
    const iterations =
      encryptedData.keyVersion === 2 ? this.PBKDF2_ITERATIONS_V2 : this.PBKDF2_ITERATIONS_V1;
    const key = await this.deriveKey(passphrase, salt, iterations);
    return this.decrypt(encryptedData, key);
  }

  /**
   * Convert EncryptedData to EncryptedDataPayload for API requests
   */
  static toPayload(encryptedData: EncryptedData): EncryptedDataPayload {
    return {
      ciphertext: encryptedData.ciphertext,
      iv: encryptedData.iv,
      salt: encryptedData.salt,
      tag: encryptedData.tag,
      key_version: encryptedData.keyVersion,
    };
  }

  /**
   * Convert EncryptedDataPayload from API to EncryptedData
   */
  static fromPayload(payload: EncryptedDataPayload): EncryptedData {
    return {
      ciphertext: payload.ciphertext,
      iv: payload.iv,
      salt: payload.salt,
      tag: payload.tag,
      keyVersion: payload.key_version,
    };
  }
}
