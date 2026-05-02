/**
 * Secure Vault Key Store
 *
 * Replaces plaintext sessionStorage passphrase storage with a layered approach:
 * - A non-extractable AES-256-GCM CryptoKey is stored in IndexedDB (device-bound)
 * - The user's passphrase is encrypted with that key before being placed in sessionStorage
 * - On page load the passphrase is decrypted from sessionStorage using the IndexedDB key
 * - Clearing IndexedDB or sessionStorage invalidates the cached passphrase
 *
 * This prevents any XSS that only has access to sessionStorage/memory from
 * exfiltrating the plaintext passphrase.
 */

const DB_NAME = 'ff_vault_keys';
const DB_VERSION = 1;
const STORE_NAME = 'device_keys';
const DEVICE_KEY_ALIAS = 'device_key';
const SESSION_ENCRYPTED_KEY = 'ff_vault_enc_passphrase';
const SESSION_SALT_KEY = 'ff_vault_passphrase_salt';

function openDB(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(DB_NAME, DB_VERSION);
    request.onupgradeneeded = () => {
      const db = request.result;
      if (!db.objectStoreNames.contains(STORE_NAME)) {
        db.createObjectStore(STORE_NAME);
      }
    };
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error);
  });
}

async function getDeviceKeyFromIDB(): Promise<CryptoKey | null> {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, 'readonly');
    const store = tx.objectStore(STORE_NAME);
    const req = store.get(DEVICE_KEY_ALIAS);
    req.onsuccess = () => resolve(req.result as CryptoKey | null);
    req.onerror = () => reject(req.error);
  });
}

async function saveDeviceKeyToIDB(key: CryptoKey): Promise<void> {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, 'readwrite');
    const store = tx.objectStore(STORE_NAME);
    const req = store.put(key, DEVICE_KEY_ALIAS);
    req.onsuccess = () => resolve();
    req.onerror = () => reject(req.error);
  });
}

async function getOrCreateDeviceKey(): Promise<CryptoKey> {
  const existing = await getDeviceKeyFromIDB();
  if (existing) return existing;

  const key = await crypto.subtle.generateKey(
    { name: 'AES-GCM', length: 256 },
    false,
    ['encrypt', 'decrypt']
  );
  await saveDeviceKeyToIDB(key);
  return key;
}

function b64Encode(buf: ArrayBuffer | Uint8Array): string {
  const bytes = buf instanceof Uint8Array ? buf : new Uint8Array(buf);
  let binary = '';
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary);
}

function b64Decode(b64: string): Uint8Array {
  const binary = atob(b64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes;
}

/**
 * Store the passphrase securely: encrypt it with the device-bound CryptoKey
 * and persist the ciphertext in sessionStorage.
 */
export async function secureStorePassphrase(passphrase: string): Promise<void> {
  const deviceKey = await getOrCreateDeviceKey();
  const iv = crypto.getRandomValues(new Uint8Array(12));
  const encoder = new TextEncoder();

  const encrypted = await crypto.subtle.encrypt(
    { name: 'AES-GCM', iv },
    deviceKey,
    encoder.encode(passphrase)
  );

  sessionStorage.setItem(SESSION_ENCRYPTED_KEY, b64Encode(encrypted));
  sessionStorage.setItem(SESSION_SALT_KEY, b64Encode(iv));
}

/**
 * Retrieve and decrypt the passphrase from sessionStorage using the device key.
 * Returns null if no passphrase is stored or decryption fails.
 */
export async function secureGetPassphrase(): Promise<string | null> {
  const encB64 = sessionStorage.getItem(SESSION_ENCRYPTED_KEY);
  const ivB64 = sessionStorage.getItem(SESSION_SALT_KEY);
  if (!encB64 || !ivB64) return null;

  try {
    const deviceKey = await getDeviceKeyFromIDB();
    if (!deviceKey) return null;

    const encrypted = b64Decode(encB64);
    const iv = b64Decode(ivB64);

    const decrypted = await crypto.subtle.decrypt(
      { name: 'AES-GCM', iv: iv as BufferSource },
      deviceKey,
      encrypted as BufferSource
    );

    return new TextDecoder().decode(decrypted);
  } catch {
    return null;
  }
}

/**
 * Check if a passphrase is cached in sessionStorage.
 */
export function hasStoredPassphrase(): boolean {
  return sessionStorage.getItem(SESSION_ENCRYPTED_KEY) !== null;
}

/**
 * Clear the cached passphrase from sessionStorage.
 */
export function secureClearPassphrase(): void {
  sessionStorage.removeItem(SESSION_ENCRYPTED_KEY);
  sessionStorage.removeItem(SESSION_SALT_KEY);
}

/**
 * Clear the device key from IndexedDB (full reset).
 */
export async function clearDeviceKey(): Promise<void> {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, 'readwrite');
    const store = tx.objectStore(STORE_NAME);
    const req = store.delete(DEVICE_KEY_ALIAS);
    req.onsuccess = () => resolve();
    req.onerror = () => reject(req.error);
  });
}
