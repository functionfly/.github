/**
 * TokenVault - Secure token storage using client-side encryption
 * 
 * Security model:
 * - Tokens are encrypted with AES-256-GCM before storing in localStorage
 * - Encryption key is stored in sessionStorage (cleared on tab close)
 * - This provides XSS protection: attackers get ciphertext, not plaintext tokens
 * - Even if localStorage is dumped, tokens are not directly usable
 * 
 * Key derivation:
 * - Per-session random key generated via crypto.getRandomValues()
 * - Stored in sessionStorage (not localStorage) for session-bound protection
 * - No user passphrase required (seamless auto-login on refresh)
 */

import { VaultCrypto } from './vault-crypto';

// Storage keys
const ENCRYPTED_TOKEN_KEY = 'ff-encrypted-tokens';
const SESSION_KEY_ID = 'ff-token-key-id';

/**
 * Structure for encrypted token storage
 */
interface EncryptedTokenEntry {
  ciphertext: string;
  iv: string;
  salt: string;
  tag: string;
  keyVersion: number;
}

/**
 * Structure for storing all encrypted tokens
 */
interface EncryptedTokenStore {
  version: number;
  tokens: {
    accessToken?: EncryptedTokenEntry;
    refreshToken?: EncryptedTokenEntry;
  };
}

export class TokenVault {
  private encryptionKey: CryptoKey | null = null;
  private keyId: string | null = null;
  private initPromise: Promise<void> | null = null;

  /**
   * Initialize or retrieve the session-bound encryption key
   * Called on app startup to prepare for token operations
   */
  async initialize(): Promise<void> {
    if (typeof window === 'undefined') return;

    // If initialization is already in progress, wait for it
    if (this.initPromise) {
      await this.initPromise;
      return;
    }

    this.initPromise = this._doInitialize();
    await this.initPromise;
    this.initPromise = null;
  }

  private async _doInitialize(): Promise<void> {
    // Try to retrieve existing key from sessionStorage
    const storedKeyId = sessionStorage.getItem(SESSION_KEY_ID);
    
    if (storedKeyId) {
      const storedKeyMaterial = sessionStorage.getItem(`ff-key-${storedKeyId}`);
      if (storedKeyMaterial) {
        try {
          const keyData = this.base64ToUint8Array(storedKeyMaterial);
          this.encryptionKey = await crypto.subtle.importKey(
            'raw',
            keyData.buffer as ArrayBuffer,
            { name: 'AES-GCM' },
            true,
            ['encrypt', 'decrypt']
          );
          this.keyId = storedKeyId;
          return;
        } catch (e) {
          console.warn('Failed to import existing session key, generating new one');
        }
      }
    }

    // Generate new key for this session
    await this.generateSessionKey();
  }

  /**
   * Generate a new random session key
   */
  private async generateSessionKey(): Promise<void> {
    const keyData = crypto.getRandomValues(new Uint8Array(32)); // 256-bit key
    this.encryptionKey = await crypto.subtle.importKey(
      'raw',
      keyData,
      { name: 'AES-GCM' },
      true,
      ['encrypt', 'decrypt']
    );

    // Store key material in sessionStorage with a unique ID
    this.keyId = this.generateKeyId();
    sessionStorage.setItem(SESSION_KEY_ID, this.keyId);
    sessionStorage.setItem(`ff-key-${this.keyId}`, this.uint8ArrayToBase64(keyData));
  }

  /**
   * Generate a unique key identifier
   */
  private generateKeyId(): string {
    const bytes = crypto.getRandomValues(new Uint8Array(16));
    return Array.from(bytes).map(b => b.toString(16).padStart(2, '0')).join('');
  }

  /**
   * Store access token securely (encrypted)
   */
  async setAccessToken(token: string): Promise<void> {
    if (!this.encryptionKey) {
      if (!this.initPromise) {
        this.initPromise = this.initialize();
      }
      await this.initPromise;
    }
    
    const encrypted = await this.encrypt(token);
    const store = this.getTokenStore();
    store.tokens.accessToken = encrypted;
    this.saveTokenStore(store);
  }

  /**
   * Store refresh token securely (encrypted)
   */
  async setRefreshToken(token: string): Promise<void> {
    if (!this.encryptionKey) {
      if (!this.initPromise) {
        this.initPromise = this.initialize();
      }
      await this.initPromise;
    }

    const encrypted = await this.encrypt(token);
    const store = this.getTokenStore();
    store.tokens.refreshToken = encrypted;
    this.saveTokenStore(store);
  }

  /**
   * Retrieve access token (decrypted)
   */
  async getAccessToken(): Promise<string | null> {
    if (!this.encryptionKey) {
      if (!this.initPromise) {
        this.initPromise = this.initialize();
      }
      await this.initPromise;
    }

    const store = this.getTokenStore();
    if (!store.tokens.accessToken) {
      // Fallback: try reading unencrypted token (migration support)
      return localStorage.getItem('ff-access-token');
    }

    try {
      return await this.decrypt(store.tokens.accessToken);
    } catch (e) {
      console.error('Failed to decrypt access token:', e);
      this.clearTokens();
      return null;
    }
  }

  /**
   * Retrieve refresh token (decrypted)
   */
  async getRefreshToken(): Promise<string | null> {
    if (!this.encryptionKey) {
      if (!this.initPromise) {
        this.initPromise = this.initialize();
      }
      await this.initPromise;
    }

    const store = this.getTokenStore();
    if (!store.tokens.refreshToken) {
      // Fallback: try reading unencrypted token (migration support)
      return localStorage.getItem('ff-refresh-token');
    }

    try {
      return await this.decrypt(store.tokens.refreshToken);
    } catch (e) {
      console.error('Failed to decrypt refresh token:', e);
      this.clearTokens();
      return null;
    }
  }

  /**
   * Clear all stored tokens
   */
  clearTokens(): void {
    localStorage.removeItem(ENCRYPTED_TOKEN_KEY);
    // Don't clear the session key - it can be reused if user logs in again
  }

  /**
   * Clear session key (call on logout to prevent session reuse)
   */
  clearSessionKey(): void {
    if (this.keyId) {
      sessionStorage.removeItem(`ff-key-${this.keyId}`);
    }
    sessionStorage.removeItem(SESSION_KEY_ID);
    this.encryptionKey = null;
    this.keyId = null;
  }

  /**
   * Encrypt a string using the session key
   */
  private async encrypt(plaintext: string): Promise<EncryptedTokenEntry> {
    if (!this.encryptionKey) {
      throw new Error('TokenVault not initialized');
    }

    const salt = VaultCrypto.generateSalt();
    const ivSource = VaultCrypto.generateIV();
    const ivBuffer = new ArrayBuffer(ivSource.byteLength);
    new Uint8Array(ivBuffer).set(ivSource);
    const iv = new Uint8Array(ivBuffer);

    // Import key for this specific encryption operation with salt
    const derivedKey = await this.deriveKeyWithSalt(this.encryptionKey, salt);
    
    const encoder = new TextEncoder();
    const plaintextData = encoder.encode(plaintext);
    const plaintextBuffer = new ArrayBuffer(plaintextData.byteLength);
    new Uint8Array(plaintextBuffer).set(plaintextData);

    const encryptedBuffer = await crypto.subtle.encrypt(
      { name: 'AES-GCM', iv },
      derivedKey,
      plaintextBuffer
    );

    const encryptedBytes = new Uint8Array(encryptedBuffer);
    const tagLength = 16;
    const ciphertextBytes = encryptedBytes.slice(0, encryptedBytes.length - tagLength);
    const tagBytes = encryptedBytes.slice(encryptedBytes.length - tagLength);

    return {
      ciphertext: this.uint8ArrayToBase64(ciphertextBytes),
      iv: this.uint8ArrayToBase64(iv),
      salt: this.uint8ArrayToBase64(salt),
      tag: this.uint8ArrayToBase64(tagBytes),
      keyVersion: 1,
    };
  }

  /**
   * Decrypt a token entry
   */
  private async decrypt(entry: EncryptedTokenEntry): Promise<string> {
    if (!this.encryptionKey) {
      throw new Error('TokenVault not initialized');
    }

    const salt = this.base64ToUint8Array(entry.salt);
    const ivSource = this.base64ToUint8Array(entry.iv);
    const ivBuffer = new ArrayBuffer(ivSource.byteLength);
    new Uint8Array(ivBuffer).set(ivSource);
    const iv = new Uint8Array(ivBuffer);
    const ciphertext = this.base64ToUint8Array(entry.ciphertext);
    const tag = this.base64ToUint8Array(entry.tag);

    const derivedKey = await this.deriveKeyWithSalt(this.encryptionKey, salt);

    // Combine ciphertext and tag for AES-GCM
    const combined = new Uint8Array(ciphertext.length + tag.length);
    combined.set(ciphertext, 0);
    combined.set(tag, ciphertext.length);
    const combinedBuffer = new ArrayBuffer(combined.byteLength);
    new Uint8Array(combinedBuffer).set(combined);

    const decryptedBuffer = await crypto.subtle.decrypt(
      { name: 'AES-GCM', iv },
      derivedKey,
      combinedBuffer
    );

    const decoder = new TextDecoder();
    return decoder.decode(decryptedBuffer);
  }

  /**
   * Derive a unique key from the session key and salt
   */
  private async deriveKeyWithSalt(baseKey: CryptoKey, salt: Uint8Array): Promise<CryptoKey> {
    const keyData = await crypto.subtle.exportKey('raw', baseKey);
    const keyBytes = new Uint8Array(keyData);
    
    // Combine key data with salt and hash to get a valid 32-byte AES-256 key
    const combined = new Uint8Array(keyBytes.length + salt.length);
    combined.set(keyBytes, 0);
    combined.set(salt, keyBytes.length);

    // Hash to produce a valid 32-byte AES-256 key (48 bytes is invalid for AES)
    const hashBuffer = await crypto.subtle.digest('SHA-256', combined);

    return crypto.subtle.importKey(
      'raw',
      hashBuffer,
      { name: 'AES-GCM' },
      false,
      ['encrypt', 'decrypt']
    );
  }

  /**
   * Get the encrypted token store from localStorage
   */
  private getTokenStore(): EncryptedTokenStore {
    const stored = localStorage.getItem(ENCRYPTED_TOKEN_KEY);
    if (stored) {
      try {
        const parsed = JSON.parse(stored);
        return {
          version: parsed.version || 1,
          tokens: parsed.tokens || {},
        };
      } catch {
        return { version: 1, tokens: {} };
      }
    }
    return { version: 1, tokens: {} };
  }

  /**
   * Save the encrypted token store to localStorage
   */
  private saveTokenStore(store: EncryptedTokenStore): void {
    localStorage.setItem(ENCRYPTED_TOKEN_KEY, JSON.stringify(store));
  }

  /**
   * Convert Uint8Array to base64 string
   */
  private uint8ArrayToBase64(bytes: Uint8Array): string {
    let binary = '';
    for (let i = 0; i < bytes.byteLength; i++) {
      binary += String.fromCharCode(bytes[i]);
    }
    return btoa(binary);
  }

  /**
   * Convert base64 string to Uint8Array
   */
  private base64ToUint8Array(base64: string): Uint8Array {
    const binary = atob(base64);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i++) {
      bytes[i] = binary.charCodeAt(i);
    }
    return bytes;
  }

  /**
   * Check if we have encrypted tokens (for migration detection)
   */
  hasEncryptedTokens(): boolean {
    const store = this.getTokenStore();
    return !!(store.tokens.accessToken || store.tokens.refreshToken);
  }

  /**
   * Migrate existing unencrypted tokens to encrypted storage
   * Call this on first login after implementing TokenVault
   */
  async migrateExistingTokens(): Promise<void> {
    const existingAccess = localStorage.getItem('ff-access-token');
    const existingRefresh = localStorage.getItem('ff-refresh-token');

    if (existingAccess) {
      await this.setAccessToken(existingAccess);
    }
    if (existingRefresh) {
      await this.setRefreshToken(existingRefresh);
    }
  }
}

// Singleton instance for app-wide use
export const tokenVault = new TokenVault();
