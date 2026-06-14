/**
 * Admin API Client
 * Handles all API communication with server-issued HMAC signatures for
 * sensitive mutating operations. The shared secret is never held in the
 * browser.
 */

import { CACHE_KEYS, getAdminApiBaseUrl } from '@/lib/constants';
import { getCsrfToken, isCsrfTokenExpired, refreshCsrfToken } from '@/lib/security/csrf';
import type { AdminAPIResponse } from '@/types';
import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';
import { clearSignatureCache, signRequest } from './hmacSigner';

// Extend Axios config with custom flags
declare module 'axios' {
  interface InternalAxiosRequestConfig {
    _skipCsrf?: boolean;
    _skipSignature?: boolean;
  }
  interface AxiosRequestConfig {
    _skipCsrf?: boolean;
    _skipSignature?: boolean;
  }
}

class AdminAPIClient {
  private client: AxiosInstance;
  private sessionToken: string | null = null;
  private deviceFingerprint: string | null = null;
  private isRefreshingToken = false;
  private refreshSubscribers: Array<(token: string | null) => void> = [];

  constructor() {
    const baseURL = getAdminApiBaseUrl();
    try {
      const stored = sessionStorage.getItem(CACHE_KEYS.ADMIN_ACCESS_TOKEN);
      if (stored) this.sessionToken = stored;
    } catch {
      /* sessionStorage unavailable */
    }

    this.client = axios.create({
      baseURL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Add request interceptor
    this.client.interceptors.request.use(async (config) => {
      if (this.sessionToken) {
        config.headers.Authorization = `Bearer ${this.sessionToken}`;
      }

      if (this.deviceFingerprint) {
        config.headers['X-Device-Fingerprint'] = this.deviceFingerprint;
      }

      // Add CSRF token and server-issued signature to mutating requests
      const method = config.method?.toUpperCase();
      if (method === 'POST' || method === 'PUT' || method === 'PATCH' || method === 'DELETE') {
        // Skip CSRF and signature if requested (e.g. for fetching CSRF or sign-request itself)
        if (config._skipCsrf) {
          return config;
        }

        // Refresh CSRF token if expired before adding to request
        if (isCsrfTokenExpired()) {
          await this.refreshCsrfTokenSafely();
        }
        const csrf = getCsrfToken();
        if (csrf) {
          config.headers['X-CSRF-Token'] = csrf;
        }

        // Ask the server to sign this request intent. The shared HMAC secret
        // is never held in the browser; the server issues a short-lived
        // signature bound to (method, path, body hash) and the current session.
        if (!config._skipSignature) {
          const body = config.data ? JSON.stringify(config.data) : '';
          const fullPath = this.fullPath((config.url || '').replace(/^([^/])/, '/$1'));
          const signature = await signRequest(method, fullPath, body);
          config.headers['X-FFLY-Timestamp'] = signature.timestamp;
          config.headers['X-FFLY-Signature'] = signature.signature;
        }
      }

      return config;
    });

    // Add response interceptor for error handling and token refresh
    this.client.interceptors.response.use(
      (response) => response,
      async (error) => {
        const originalRequest = error.config;

        // Skip redirect if the request is marked to avoid auth redirects
        // (used by the login page to check last-login without triggering a redirect loop)
        if (originalRequest._skipAuthRedirect) {
          return Promise.reject(error);
        }

        // Handle CSRF token expiration (401 or 403 with csrf_token_invalid error)
        if (
          (error.response?.status === 401 || error.response?.status === 403) &&
          error.response?.data?.error === 'csrf_token_invalid' &&
          !originalRequest._retry
        ) {
          originalRequest._retry = true;

          const newToken = await this.refreshCsrfTokenSafely();
          if (newToken) {
            originalRequest.headers['X-CSRF-Token'] = newToken;
            return this.client(originalRequest);
          }
        }

        // Handle session expiration — clear store state (which triggers ProtectedRoute
        // to redirect to /auth/login via React Router) instead of a hard navigation.
        // Hard navigation (window.location.href) causes a full reload, which re-runs
        // AdminAuthRestore and can restore a still-valid session token, bouncing the
        // user back to / before they can even see the login page.
        // Only treat as a session expiry if the backend signals it explicitly.
        // HMAC failures, permission denials, etc. also return 401 but should NOT
        // log the user out — they should surface as errors to the calling code.
        // Session-expiry responses have a JSON body with error: "session_expired"
        // or "token_expired" (or similar), set by the auth middleware.
        if (error.response?.status === 401) {
          const errorCode = error.response?.data?.error || error.response?.data?.code || '';
          const isSessionExpiry =
            errorCode === 'session_expired' ||
            errorCode === 'token_expired' ||
            errorCode === 'invalid_token' ||
            errorCode === 'unauthorized';

          if (isSessionExpiry) {
            import('@/stores/adminAuthStore').then(({ useAdminAuthStore }) => {
              useAdminAuthStore.getState().logout();
            });
          }
        }
        return Promise.reject(error);
      }
    );
  }

  /**
   * Safely refresh CSRF token with request coalescing
   */
  private async refreshCsrfTokenSafely(): Promise<string | null> {
    if (this.isRefreshingToken) {
      // Wait for ongoing refresh
      return new Promise((resolve) => {
        this.refreshSubscribers.push(resolve);
      });
    }

    this.isRefreshingToken = true;
    try {
      const token = await refreshCsrfToken();
      this.refreshSubscribers.forEach((cb) => cb(token));
      return token;
    } catch {
      this.refreshSubscribers.forEach((cb) => cb(null));
      return null;
    } finally {
      this.isRefreshingToken = false;
      this.refreshSubscribers = [];
    }
  }

  /**
   * Set session token for authenticated requests
   */
  setSessionToken(token: string) {
    this.sessionToken = token;
  }

  /**
   * Clear session token. Also drops the in-memory signature cache so a new
   * session never reuses a signature issued under a previous one.
   */
  clearSessionToken() {
    this.sessionToken = null;
    clearSignatureCache();
  }

  setDeviceFingerprint(fingerprint: string) {
    this.deviceFingerprint = fingerprint;
  }

  clearDeviceFingerprint() {
    this.deviceFingerprint = null;
  }

  /**
   * Returns true if the client has a session token stored.
   */
  isAuthenticated(): boolean {
    return this.sessionToken !== null;
  }

  /**
   * GET request
   */
  async get<T>(path: string, config?: AxiosRequestConfig): Promise<AdminAPIResponse<T>> {
    const response = await this.client.get<AdminAPIResponse<T>>(path, config);
    return response.data;
  }

  /**
   * GET request that returns the raw Axios response (no envelope unwrapping).
   * Used by internal helpers like the CSRF refresher that need the unparsed body.
   */
  async getRaw<T = unknown>(
    path: string,
    config?: AxiosRequestConfig
  ): Promise<{ data: T; status: number }> {
    const response = await this.client.get<T>(path, config);
    return { data: response.data, status: response.status };
  }

  /**
   * GET request that does NOT trigger a 401 redirect.
   * Use this for requests on public pages (e.g. login page last-login check)
   * where a 401 is an expected "not authenticated" state, not a session expiry.
   */
  async getNoAuth<T>(path: string): Promise<T | null> {
    try {
      const response = await this.client.get<T>(path, {
        _skipAuthRedirect: true,
      } as AxiosRequestConfig);
      return response.data;
    } catch (error: any) {
      if (error?.response?.status === 401) {
        return null; // Not authenticated — expected on login page
      }
      throw error;
    }
  }

  /**
   * POST request that does NOT trigger a 401 redirect or HMAC signing.
   * Use for unauthenticated endpoints on the login page (forgot password,
   * MFA challenge exchange, etc.) where we deliberately have no session yet.
   */
  async postNoAuth<T = unknown>(path: string, data?: unknown): Promise<T | null> {
    try {
      const response = await this.client.post<T>(path, data, {
        _skipAuthRedirect: true,
        _skipSignature: true,
      } as AxiosRequestConfig);
      return response.data;
    } catch (error: any) {
      if (error?.response?.status === 401 || error?.response?.status === 403) {
        return null;
      }
      throw error;
    }
  }

  /**
   * POST request that returns the raw Axios response (no envelope unwrapping).
   * Used by internal helpers like the server-issued HMAC signer that need
   * the unparsed body.
   */
  async postRaw<T = unknown>(
    path: string,
    data?: any,
    config?: AxiosRequestConfig
  ): Promise<{ data: T; status: number }> {
    const response = await this.client.post<T>(path, data, config);
    return { data: response.data, status: response.status };
  }

  /**
   * POST request. The request interceptor attaches the server-issued
   * signature and CSRF token automatically.
   */
  async post<T>(
    path: string,
    data?: any,
    config?: AxiosRequestConfig
  ): Promise<AdminAPIResponse<T>> {
    const response = await this.client.post<AdminAPIResponse<T>>(path, data, config);
    return response.data;
  }

  /**
   * PUT request. The request interceptor attaches the server-issued signature.
   */
  async put<T>(
    path: string,
    data?: any,
    config?: AxiosRequestConfig
  ): Promise<AdminAPIResponse<T>> {
    const response = await this.client.put<AdminAPIResponse<T>>(path, data, config);
    return response.data;
  }

  /**
   * PATCH request. The request interceptor attaches the server-issued signature.
   */
  async patch<T>(
    path: string,
    data?: any,
    config?: AxiosRequestConfig
  ): Promise<AdminAPIResponse<T>> {
    const response = await this.client.patch<AdminAPIResponse<T>>(path, data, config);
    return response.data;
  }

  /**
   * DELETE request. The request interceptor attaches the server-issued signature.
   */
  async delete<T>(path: string, config?: AxiosRequestConfig): Promise<AdminAPIResponse<T>> {
    const response = await this.client.delete<AdminAPIResponse<T>>(path, config);
    return response.data;
  }

  /**
   * Build the full server-visible path for HMAC signing.
   * baseURL is e.g. "https://api.functionfly.com/v1/admin" and reqPath is
   * a relative path like "/signup-invites".  The result is "/v1/admin/signup-invites".
   */
  private fullPath(reqPath: string): string {
    const baseUrl = new URL(this.client.defaults.baseURL || '');
    const basePath = baseUrl.pathname.replace(/\/$/, '');
    const normalised = reqPath.replace(/^([^/])/, '/$1');
    return basePath + normalised;
  }
}

// Singleton instance
export const adminApiClient = new AdminAPIClient();

export default adminApiClient;

// Provider Settings types and API methods
export interface ProviderSettings {
  provider: string;
  disabled: boolean;
  disabled_reason?: string;
  disabled_at?: string;
  disabled_by?: string;
}

export interface ProviderSettingsResponse {
  settings: ProviderSettings[];
}

export interface UpdateProviderSettingsRequest {
  disabled: boolean;
  reason?: string;
}

export async function getProviderSettings(): Promise<ProviderSettings[]> {
  const response = await adminApiClient.get<ProviderSettingsResponse>('/providers/settings');
  return response.settings ?? [];
}

export async function updateProviderSettings(
  provider: string,
  data: UpdateProviderSettingsRequest
): Promise<ProviderSettings> {
  const response = await adminApiClient.patch<{ settings: ProviderSettings }>(
    `/providers/settings/${provider}`,
    data
  );
  return response.settings;
}
