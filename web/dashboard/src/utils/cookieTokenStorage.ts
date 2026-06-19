/**
 * Cookie-based token storage utilities
 * 
 * This module provides utilities for reading tokens from HttpOnly cookies
 * set by the backend. While the TokenVault provides client-side encryption
 * for additional security, these utilities allow the app to read tokens
 * that the server has set in HttpOnly cookies.
 * 
 * HttpOnly cookies provide XSS protection because JavaScript cannot access them.
 * The browser automatically includes them with API requests.
 */

const ACCESS_TOKEN_COOKIE = 'ff_access_token';
const REFRESH_TOKEN_COOKIE = 'ff_refresh_token';
const CSRF_TOKEN_COOKIE = 'ff_csrf_token';

/**
 * Parse cookies from document.cookie string
 */
function parseCookies(cookieString: string): Record<string, string> {
  const cookies: Record<string, string> = {};
  if (!cookieString) return cookies;
  
  cookieString.split(';').forEach(cookie => {
    const [name, ...valueParts] = cookie.trim().split('=');
    if (name && valueParts.length > 0) {
      cookies[name] = valueParts.join('=');
    }
  });
  
  return cookies;
}

/**
 * Get all tokens stored in cookies
 */
export function getCookieTokens(): {
  accessToken: string | null;
  refreshToken: string | null;
  csrfToken: string | null;
} {
  if (typeof document === 'undefined') {
    return { accessToken: null, refreshToken: null, csrfToken: null };
  }
  
  const cookies = parseCookies(document.cookie);
  
  return {
    accessToken: cookies[ACCESS_TOKEN_COOKIE] || null,
    refreshToken: cookies[REFRESH_TOKEN_COOKIE] || null,
    csrfToken: cookies[CSRF_TOKEN_COOKIE] || null,
  };
}

/**
 * Check if tokens are present in cookies (indicates server-set HttpOnly auth)
 */
export function hasCookieTokens(): boolean {
  if (typeof document === 'undefined') return false;
  const cookies = parseCookies(document.cookie);
  return !!(cookies[ACCESS_TOKEN_COOKIE] || cookies[REFRESH_TOKEN_COOKIE]);
}

/**
 * Get access token from cookie (if present)
 */
export function getAccessTokenFromCookie(): string | null {
  if (typeof document === 'undefined') return null;
  const cookies = parseCookies(document.cookie);
  return cookies[ACCESS_TOKEN_COOKIE] || null;
}

/**
 * Get refresh token from cookie (if present)
 */
export function getRefreshTokenFromCookie(): string | null {
  if (typeof document === 'undefined') return null;
  const cookies = parseCookies(document.cookie);
  return cookies[REFRESH_TOKEN_COOKIE] || null;
}

/**
 * Cookie names for external reference
 */
export const COOKIE_NAMES = {
  ACCESS_TOKEN: ACCESS_TOKEN_COOKIE,
  REFRESH_TOKEN: REFRESH_TOKEN_COOKIE,
  CSRF_TOKEN: CSRF_TOKEN_COOKIE,
} as const;
