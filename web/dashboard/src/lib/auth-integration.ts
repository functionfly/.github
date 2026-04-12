/**
 * Auth integration with @web/auth (functionfly-auth microsite)
 *
 * This module handles the integration between the dashboard and the standalone
 * auth site (auth.functionfly.com). The auth site manages login/signup flows
 * and redirects back to the dashboard with tokens stored in sessionStorage.
 *
 * Flow:
 * 1. Dashboard detects unauthenticated user
 * 2. Redirects to auth site with ?redirect_uri=dashboard/login/callback
 * 3. User authenticates on auth site
 * 4. Auth site stores tokens in sessionStorage and redirects to /auth/callback
 * 5. Dashboard callback handler extracts tokens and stores in localStorage
 * 6. User continues to intended destination
 */

import { logger } from '@/lib/logger';

/** Auth site origin - matches web/auth/src/config.ts */
export function getAuthSiteOrigin(): string {
  const env = (import.meta.env.VITE_AUTH_SITE_URL ?? '').trim().replace(/\/$/, '');
  if (env) return env;

  // Dev default - auth site runs on port 4324
  if (typeof window !== 'undefined') {
    const h = window.location.hostname;
    if (h === 'localhost' || h === '127.0.0.1') return 'http://localhost:4324';
  }

  return 'https://auth.functionfly.com';
}

/**
 * Dashboard (app) origin - where users go after auth
 * Production: https://app.functionfly.com
 * Dev: http://localhost:3000
 *
 * Uses VITE_APP_URL if set, otherwise falls back to window.location.origin
 */
export function getDashboardOrigin(): string {
  // Allow explicit configuration via env var
  const envUrl = (import.meta.env.VITE_APP_URL ?? '').trim().replace(/\/$/, '');
  if (envUrl) return envUrl;

  // Fall back to current origin (works in dev and when deployed correctly)
  if (typeof window !== 'undefined') {
    return window.location.origin;
  }

  // Default for production SSR/build contexts
  return 'https://app.functionfly.com';
}

/**
 * Build the auth site login URL with redirect parameter
 */
export function buildAuthSiteLoginUrl(redirectPath?: string): string {
  const authOrigin = getAuthSiteOrigin();
  const dashboardOrigin = getDashboardOrigin();

  // Build the final redirect URL (where user should land after auth)
  const finalRedirect = redirectPath
    ? `${dashboardOrigin}${redirectPath.startsWith('/') ? redirectPath : `/${redirectPath}`}`
    : `${dashboardOrigin}/overview`;

  // Encode the redirect URL
  const encodedRedirect = encodeURIComponent(finalRedirect);

  return `${authOrigin}/login?redirect_uri=${encodedRedirect}`;
}

/**
 * Build the auth site signup URL with redirect parameter
 */
export function buildAuthSiteSignupUrl(redirectPath?: string, inviteCode?: string): string {
  const authOrigin = getAuthSiteOrigin();
  const dashboardOrigin = getDashboardOrigin();

  const finalRedirect = redirectPath
    ? `${dashboardOrigin}${redirectPath.startsWith('/') ? redirectPath : `/${redirectPath}`}`
    : `${dashboardOrigin}/overview`;

  const encodedRedirect = encodeURIComponent(finalRedirect);
  const inviteParam = inviteCode ? `&invite_code=${encodeURIComponent(inviteCode)}` : '';

  return `${authOrigin}/signup?redirect_uri=${encodedRedirect}${inviteParam}`;
}

/**
 * Build the auth site forgot password URL
 */
export function buildAuthSiteForgotPasswordUrl(): string {
  const authOrigin = getAuthSiteOrigin();
  return `${authOrigin}/forgot-password`;
}

/**
 * Extract and migrate tokens from sessionStorage (set by auth site) to localStorage
 * This is called by the auth callback handler.
 *
 * The auth site uses these sessionStorage keys:
 * - ff_token (access token)
 * - ff_refresh_token (refresh token)
 */
export function migrateTokensFromSessionStorage(): {
  accessToken: string | null;
  refreshToken: string | null;
  source: 'sessionStorage' | 'localStorage' | 'none';
} {
  if (typeof window === 'undefined') {
    return { accessToken: null, refreshToken: null, source: 'none' };
  }

  // Try sessionStorage first (tokens set by auth site callback)
  const sessionToken = sessionStorage.getItem('ff_token');
  const sessionRefresh = sessionStorage.getItem('ff_refresh_token');

  if (sessionToken) {
    // Migrate to localStorage for dashboard's existing token management
    localStorage.setItem('ff-access-token', sessionToken);
    if (sessionRefresh) {
      localStorage.setItem('ff-refresh-token', sessionRefresh);
    }

    // Clear from sessionStorage to prevent reuse
    sessionStorage.removeItem('ff_token');
    sessionStorage.removeItem('ff_refresh_token');

    logger.info('Migrated tokens from sessionStorage to localStorage');

    return {
      accessToken: sessionToken,
      refreshToken: sessionRefresh,
      source: 'sessionStorage',
    };
  }

  // Fall back to checking existing localStorage tokens
  const localToken = localStorage.getItem('ff-access-token');
  const localRefresh = localStorage.getItem('ff-refresh-token');

  if (localToken) {
    return {
      accessToken: localToken,
      refreshToken: localRefresh,
      source: 'localStorage',
    };
  }

  return { accessToken: null, refreshToken: null, source: 'none' };
}

/**
 * Check if tokens exist in sessionStorage (waiting to be migrated)
 * This can be used to detect if we're in a fresh auth callback flow
 */
export function hasPendingAuthTokens(): boolean {
  if (typeof window === 'undefined') return false;

  const sessionToken = sessionStorage.getItem('ff_token');
  return !!sessionToken;
}

/**
 * Clear all auth tokens from both storage types
 */
export function clearAllAuthTokens(): void {
  if (typeof window === 'undefined') return;

  // Clear localStorage (dashboard's primary storage)
  localStorage.removeItem('ff-access-token');
  localStorage.removeItem('ff-refresh-token');
  localStorage.removeItem('ff-last-wallet-agent-id');

  // Clear sessionStorage (auth site's temporary storage)
  sessionStorage.removeItem('ff_token');
  sessionStorage.removeItem('ff_refresh_token');

  logger.info('Cleared all auth tokens from storage');
}

/**
 * Redirect to auth site for authentication
 * Use this instead of <Navigate to="/login" /> when user needs to authenticate
 */
export function redirectToAuthSite(redirectPath?: string): void {
  if (typeof window === 'undefined') return;

  const loginUrl = buildAuthSiteLoginUrl(redirectPath);
  window.location.href = loginUrl;
}

/**
 * Redirect to auth site for registration
 */
export function redirectToAuthSiteSignup(redirectPath?: string, inviteCode?: string): void {
  if (typeof window === 'undefined') return;

  const signupUrl = buildAuthSiteSignupUrl(redirectPath, inviteCode);
  window.location.href = signupUrl;
}
