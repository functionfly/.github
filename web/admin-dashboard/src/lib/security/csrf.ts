/**
 * CSRF token management for admin mutating requests.
 * Token is stored in memory only (not localStorage) to prevent XSS theft.
 */

import { adminApiClient } from '@/lib/api/adminClient';

// In-memory token storage (not localStorage to prevent XSS theft)
let csrfToken: string | null = null;
let csrfTokenExpiry: number | null = null;
const CSRF_TOKEN_TTL = 3600000; // 1 hour in ms

function readCookie(name: string): string | null {
  const encodedName = `${encodeURIComponent(name)}=`;
  const parts = document.cookie.split(';');

  for (const part of parts) {
    const trimmed = part.trim();
    if (trimmed.startsWith(encodedName)) {
      return decodeURIComponent(trimmed.slice(encodedName.length));
    }
  }

  return null;
}

/**
 * Get the current CSRF token from memory, meta tag, or cookie.
 * Does not trigger a fetch - use refreshCsrfToken() for that.
 */
export function getCsrfToken(): string | null {
  // Return in-memory token if valid
  if (csrfToken && csrfTokenExpiry && Date.now() < csrfTokenExpiry) {
    return csrfToken;
  }

  // Check meta tag first
  const fromMeta = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content');
  if (fromMeta) {
    csrfToken = fromMeta;
    csrfTokenExpiry = Date.now() + CSRF_TOKEN_TTL;
    return csrfToken;
  }

  // Fall back to cookie
  const fromCookie = readCookie('csrf_token') || readCookie('XSRF-TOKEN');
  if (fromCookie) {
    csrfToken = fromCookie;
    csrfTokenExpiry = Date.now() + CSRF_TOKEN_TTL;
    return csrfToken;
  }

  return null;
}

/**
 * Check if the CSRF token needs to be refreshed.
 */
export function isCsrfTokenExpired(): boolean {
  if (!csrfToken || !csrfTokenExpiry) {
    return true;
  }
  return Date.now() >= csrfTokenExpiry;
}

/**
 * Refresh CSRF token from the server.
 * Makes a GET request to /v1/admin/csrf to get a new token.
 */
export async function refreshCsrfToken(): Promise<string | null> {
  try {
    // Import adminApiClient dynamically to avoid circular dependencies
    const { adminApiClient } = await import('@/lib/api/adminClient');

    // Make a direct request to the CSRF endpoint (this will include JWT token)
    // Note: This request will be intercepted and will try to add CSRF token,
    // but since we're fetching the token, we need to temporarily skip CSRF for this request
    const response = await adminApiClient.client.get('/csrf', {
      _skipCsrf: true, // Custom flag to skip CSRF token addition
    });

    if (response.data?.token) {
      csrfToken = response.data.token;
      csrfTokenExpiry = Date.now() + CSRF_TOKEN_TTL;
      return csrfToken;
    }

    return null;
  } catch (error) {
    console.warn('Failed to refresh CSRF token:', error);
    return null;
  }
}

/**
 * Clear the in-memory CSRF token.
 * Should be called on logout.
 */
export function clearCsrfToken(): void {
  csrfToken = null;
  csrfTokenExpiry = null;
}

/**
 * Initialize CSRF token on app startup.
 * Returns the token if successfully fetched, null otherwise.
 */
export async function initializeCsrfToken(): Promise<string | null> {
  // If we have a valid token, use it
  const existing = getCsrfToken();
  if (existing && !isCsrfTokenExpired()) {
    return existing;
  }

  // Otherwise, try to refresh from server
  return refreshCsrfToken();
}
