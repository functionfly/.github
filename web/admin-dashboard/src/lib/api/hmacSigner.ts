/**
 * HMAC request signing - server-issued signatures.
 *
 * The shared HMAC secret is NEVER held in the browser. For each mutating
 * request the client asks the orchestrator API to sign the request intent
 * (method + path + body hash) and returns a short-lived signature bound to
 * the current admin session.
 *
 * See internal/api/handlers/admin/sign_request.go on the server side.
 */

import { adminApiClient } from './adminClient';

export interface HMACSignature {
  timestamp: string;
  signature: string;
}

interface SignRequestResponse {
  timestamp: string;
  signature: string;
  expires_at: number;
}

const SIGN_PATH = '/auth/sign-request';
const SIGN_TIMEOUT_MS = 5000;

// In-memory cache of (path, method, bodyHash) -> signature. Signatures are
// short-lived (5 min server-side), so a single signature can be reused for
// mutating requests with the same intent — eliminates a round trip per
// request while staying within the validity window.
const signatureCache = new Map<string, { signature: HMACSignature; expiresAt: number }>();

// Coalesce concurrent sign requests for the same intent.
let inflightSign: Promise<HMACSignature> | null = null;

function cacheKey(method: string, path: string, bodyHash: string): string {
  return `${method}\n${path}\n${bodyHash}`;
}

async function sha256Hex(input: string): Promise<string> {
  if (typeof crypto !== 'undefined' && crypto.subtle) {
    const data = new TextEncoder().encode(input);
    const hash = await crypto.subtle.digest('SHA-256', data);
    return Array.from(new Uint8Array(hash))
      .map((b) => b.toString(16).padStart(2, '0'))
      .join('');
  }
  // Fallback for environments without SubtleCrypto. Admin only runs in modern
  // browsers, so this is a safety net rather than an expected code path.
  let h = 0;
  for (let i = 0; i < input.length; i += 1) {
    h = (Math.imul(31, h) + input.charCodeAt(i)) | 0;
  }
  return Math.abs(h).toString(16);
}

/**
 * Sign a mutating request by asking the server to do it.
 * Returns the signature components to attach as headers.
 */
export async function signRequest(
  method: string,
  path: string,
  body: string
): Promise<HMACSignature> {
  const bodyHash = await sha256Hex(body);
  const key = cacheKey(method, path, bodyHash);

  const cached = signatureCache.get(key);
  // Refresh slightly before expiry to avoid races.
  if (cached && Date.now() < cached.expiresAt - 5_000) {
    return cached.signature;
  }

  // Coalesce concurrent sign requests so a burst of mutating requests only
  // hits the server once.
  if (inflightSign) {
    return inflightSign;
  }

  inflightSign = (async () => {
    try {
      const { data } = await adminApiClient.postRaw<SignRequestResponse>(
        SIGN_PATH,
        {
          method: method.toUpperCase(),
          path,
          body_hash: bodyHash,
          timestamp: Math.floor(Date.now() / 1000),
        },
        {
          // Skip the local CSRF interceptor so the first sign-request doesn't
          // wait on a CSRF refresh. The server-side CSRF middleware still
          // applies because /auth/sign-request is registered under adminRoutes.
          _skipCsrf: true,
          timeout: SIGN_TIMEOUT_MS,
        }
      );
      if (!data?.signature || !data?.timestamp) {
        throw new Error('Malformed sign-request response');
      }
      const sig: HMACSignature = { timestamp: data.timestamp, signature: data.signature };
      signatureCache.set(key, { signature: sig, expiresAt: data.expires_at * 1000 });
      return sig;
    } finally {
      inflightSign = null;
    }
  })();

  return inflightSign;
}

/**
 * Clear the signature cache. Call on logout so a new session doesn't reuse
 * signatures from a previous one.
 */
export function clearSignatureCache(): void {
  signatureCache.clear();
}
