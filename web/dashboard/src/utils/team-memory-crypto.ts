/**
 * Team Memory Encryption Module
 * 
 * Client-side AES-256-GCM encryption for team memories.
 * Server never sees plaintext - zero-knowledge design.
 */

const PBKDF2_ITERATIONS = 100000;
const AES_KEY_SIZE = 256;
const IV_SIZE = 12;
const TAG_SIZE = 16;

export interface EncryptedPayload {
  ciphertext: Uint8Array;
  iv: Uint8Array;
  tag: Uint8Array;
}

export interface EncryptedMemoryData {
  ciphertext: string; // base64
  iv: string;         // base64
  tag: string;        // base64
}

export class TeamMemoryCrypto {
  private key: CryptoKey | null = null;
  private salt: Uint8Array | null = null;

  /**
   * Initialize with team passphrase
   * Derives AES-256 key using PBKDF2
   */
  async initialize(passphrase: string, salt?: Uint8Array): Promise<void> {
    this.salt = salt || crypto.getRandomValues(new Uint8Array(16));
    this.key = await this.deriveKey(passphrase, this.salt);
  }

  /**
   * Derive key from passphrase using PBKDF2
   */
  private async deriveKey(passphrase: string, salt: Uint8Array): Promise<CryptoKey> {
    const encoder = new TextEncoder();
    const keyMaterial = await crypto.subtle.importKey(
      'raw',
      encoder.encode(passphrase),
      'PBKDF2',
      false,
      ['deriveKey']
    );

    return crypto.subtle.deriveKey(
      {
        name: 'PBKDF2',
        salt: salt as BufferSource,
        iterations: PBKDF2_ITERATIONS,
        hash: 'SHA-256'
      },
      keyMaterial,
      { name: 'AES-GCM', length: AES_KEY_SIZE },
      false,
      ['encrypt', 'decrypt']
    );
  }

  /**
   * Encrypt memory content before sending to API
   */
  async encrypt(content: object): Promise<EncryptedPayload> {
    if (!this.key) {
      throw new Error('Crypto not initialized. Call initialize() first.');
    }

    // Generate random IV
    const iv = crypto.getRandomValues(new Uint8Array(IV_SIZE));

    // Encode content as JSON
    const encoder = new TextEncoder();
    const plaintext = encoder.encode(JSON.stringify(content));

    // Encrypt using AES-GCM
    const encrypted = await crypto.subtle.encrypt(
      { name: 'AES-GCM', iv, tagLength: TAG_SIZE * 8 },
      this.key,
      plaintext
    );

    // Split ciphertext and auth tag (last 16 bytes)
    const encryptedBytes = new Uint8Array(encrypted);
    const ciphertext = encryptedBytes.slice(0, -TAG_SIZE);
    const tag = encryptedBytes.slice(-TAG_SIZE);

    return { ciphertext, iv, tag };
  }

  /**
   * Decrypt memory content received from API
   */
  async decrypt(ciphertext: Uint8Array, iv: Uint8Array, tag: Uint8Array): Promise<object> {
    if (!this.key) {
      throw new Error('Crypto not initialized. Call initialize() first.');
    }

    // Concatenate ciphertext + tag for WebCrypto
    const combined = new Uint8Array(ciphertext.length + tag.length);
    combined.set(ciphertext);
    combined.set(tag, ciphertext.length);

    // Decrypt
    const decrypted = await crypto.subtle.decrypt(
      { name: 'AES-GCM', iv: iv as BufferSource, tagLength: TAG_SIZE * 8 },
      this.key,
      combined as BufferSource
    );

    // Decode JSON
    const decoder = new TextDecoder();
    const jsonStr = decoder.decode(decrypted);
    return JSON.parse(jsonStr);
  }

  /**
   * Encrypt and format for API submission
   */
  async encryptForAPI(content: object): Promise<EncryptedMemoryData> {
    const { ciphertext, iv, tag } = await this.encrypt(content);

    return {
      ciphertext: arrayBufferToBase64(ciphertext),
      iv: arrayBufferToBase64(iv),
      tag: arrayBufferToBase64(tag)
    };
  }

  /**
   * Decrypt from API response format
   */
  async decryptFromAPI(data: EncryptedMemoryData): Promise<object> {
    const ciphertext = base64ToArrayBuffer(data.ciphertext);
    const iv = base64ToArrayBuffer(data.iv);
    const tag = base64ToArrayBuffer(data.tag);

    return this.decrypt(ciphertext, iv, tag);
  }

  /**
   * Get salt for storage (so key can be re-derived)
   */
  getSalt(): Uint8Array | null {
    return this.salt;
  }

  /**
   * Get salt as base64 for storage
   */
  getSaltBase64(): string | null {
    return this.salt ? arrayBufferToBase64(this.salt) : null;
  }

  /**
   * Check if initialized
   */
  isInitialized(): boolean {
    return this.key !== null;
  }

  /**
   * Clear key from memory
   */
  clear(): void {
    this.key = null;
    this.salt = null;
  }
}

/**
 * Generate a secure team passphrase
 */
export function generateSecurePassphrase(length: number = 16): string {
  const charset = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*';
  const values = crypto.getRandomValues(new Uint8Array(length));
  let result = '';
  for (let i = 0; i < length; i++) {
    result += charset[values[i] % charset.length];
  }
  return result;
}

/**
 * Store passphrase in sessionStorage (more secure than localStorage)
 */
export function storePassphrase(teamId: string, passphrase: string): void {
  sessionStorage.setItem(`team_memory_passphrase_${teamId}`, passphrase);
}

/**
 * Retrieve passphrase from sessionStorage
 */
export function getPassphrase(teamId: string): string | null {
  return sessionStorage.getItem(`team_memory_passphrase_${teamId}`);
}

/**
 * Clear passphrase from sessionStorage
 */
export function clearPassphrase(teamId: string): void {
  sessionStorage.removeItem(`team_memory_passphrase_${teamId}`);
}

/**
 * Check if passphrase exists for team
 */
export function hasPassphrase(teamId: string): boolean {
  return getPassphrase(teamId) !== null;
}

/**
 * Convert ArrayBuffer to base64
 */
function arrayBufferToBase64(buffer: Uint8Array): string {
  let binary = '';
  for (let i = 0; i < buffer.byteLength; i++) {
    binary += String.fromCharCode(buffer[i]);
  }
  return btoa(binary);
}

/**
 * Convert base64 to ArrayBuffer
 */
function base64ToArrayBuffer(base64: string): Uint8Array {
  const binary = atob(base64);
  const buffer = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    buffer[i] = binary.charCodeAt(i);
  }
  return buffer;
}

/**
 * Singleton instance for app-wide use
 */
let cryptoInstance: TeamMemoryCrypto | null = null;

export function getTeamMemoryCrypto(): TeamMemoryCrypto {
  if (!cryptoInstance) {
    cryptoInstance = new TeamMemoryCrypto();
  }
  return cryptoInstance;
}

/**
 * Initialize crypto with team passphrase
 * Attempts to retrieve passphrase from sessionStorage first
 */
export async function initTeamMemoryCrypto(
  teamId: string,
  passphrase?: string
): Promise<TeamMemoryCrypto> {
  const crypto = getTeamMemoryCrypto();

  if (crypto.isInitialized()) {
    return crypto;
  }

  // Try to get passphrase from sessionStorage
  const storedPassphrase = passphrase || getPassphrase(teamId);

  if (!storedPassphrase) {
    throw new Error(
      `No passphrase found for team ${teamId}. ` +
      `Please provide the team passphrase to view encrypted memories.`
    );
  }

  await crypto.initialize(storedPassphrase);

  // Store in sessionStorage for this session if not already there
  if (!hasPassphrase(teamId)) {
    storePassphrase(teamId, storedPassphrase);
  }

  return crypto;
}

/**
 * Check if team has encrypted memories and passphrase is available
 */
export function canDecryptMemories(teamId: string): boolean {
  return hasPassphrase(teamId);
}

/**
 * Prompt user for passphrase
 */
export async function promptForPassphrase(
  teamId: string,
  promptMessage?: string
): Promise<string | null> {
  const message = promptMessage ||
    `Enter team passphrase to access encrypted memories:\n` +
    `(Team ID: ${teamId})\n\n` +
    `This is required only once per session.`;

  const passphrase = window.prompt(message);

  if (passphrase) {
    storePassphrase(teamId, passphrase);
    return passphrase;
  }

  return null;
}

export default TeamMemoryCrypto;
