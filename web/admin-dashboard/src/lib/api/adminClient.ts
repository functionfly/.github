/**
 * Admin API Client
 * Handles all API communication with HMAC signing for sensitive operations
 */

import { CACHE_KEYS, getAdminApiBaseUrl } from '@/lib/constants';
import { getCsrfToken, isCsrfTokenExpired, refreshCsrfToken } from '@/lib/security/csrf';
import type { AdminAPIResponse } from '@/types';
import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';
import { HMACRequestSigner } from './hmacSigner';

class AdminAPIClient {
  private client: AxiosInstance;
  private signer: HMACRequestSigner;
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

    // Initialize HMAC signer
    const sharedSecret = import.meta.env.VITE_ADMIN_SHARED_SECRET || '';
    this.signer = new HMACRequestSigner(sharedSecret);

    // Add request interceptor
    this.client.interceptors.request.use(async (config) => {
      if (this.sessionToken) {
        config.headers.Authorization = `Bearer ${this.sessionToken}`;
      }

      if (this.deviceFingerprint) {
        config.headers['X-Device-Fingerprint'] = this.deviceFingerprint;
      }

      // Add CSRF token to mutating requests
      const method = config.method?.toUpperCase();
      if (method === 'POST' || method === 'PUT' || method === 'PATCH' || method === 'DELETE') {
        // Refresh token if expired before adding to request
        if (isCsrfTokenExpired()) {
          await this.refreshCsrfTokenSafely();
        }
        const csrf = getCsrfToken();
        if (csrf) {
          config.headers['X-CSRF-Token'] = csrf;
        }
      }

      return config;
    });

    // Add response interceptor for error handling and token refresh
    this.client.interceptors.response.use(
      (response) => response,
      async (error) => {
        const originalRequest = error.config;

        // Handle CSRF token expiration (401 with csrf_token_invalid error)
        if (
          error.response?.status === 401 &&
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

        // Handle session expiration
        if (error.response?.status === 401) {
          // Session expired, redirect to login
          window.location.href = '/auth/login?reason=session_expired';
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
   * Clear session token
   */
  clearSessionToken() {
    this.sessionToken = null;
  }

  setDeviceFingerprint(fingerprint: string) {
    this.deviceFingerprint = fingerprint;
  }

  clearDeviceFingerprint() {
    this.deviceFingerprint = null;
  }

  /**
   * GET request
   */
  async get<T>(path: string, config?: AxiosRequestConfig): Promise<AdminAPIResponse<T>> {
    const response = await this.client.get<AdminAPIResponse<T>>(path, config);
    return response.data;
  }

  /**
   * POST request with HMAC signing for sensitive operations
   */
  async post<T>(
    path: string,
    data?: any,
    config?: AxiosRequestConfig
  ): Promise<AdminAPIResponse<T>> {
    const signedConfig = this.signRequest('POST', path, data, config);
    const response = await this.client.post<AdminAPIResponse<T>>(path, data, signedConfig);
    return response.data;
  }

  /**
   * PATCH request with HMAC signing
   */
  async patch<T>(
    path: string,
    data?: any,
    config?: AxiosRequestConfig
  ): Promise<AdminAPIResponse<T>> {
    const signedConfig = this.signRequest('PATCH', path, data, config);
    const response = await this.client.patch<AdminAPIResponse<T>>(path, data, signedConfig);
    return response.data;
  }

  /**
   * DELETE request with HMAC signing
   */
  async delete<T>(path: string, config?: AxiosRequestConfig): Promise<AdminAPIResponse<T>> {
    const signedConfig = this.signRequest('DELETE', path, undefined, config);
    const response = await this.client.delete<AdminAPIResponse<T>>(path, signedConfig);
    return response.data;
  }

  /**
   * Sign request with HMAC
   */
  private signRequest(
    method: string,
    path: string,
    data?: any,
    config?: AxiosRequestConfig
  ): AxiosRequestConfig {
    const body = data ? JSON.stringify(data) : '';
    const { timestamp, signature } = this.signer.sign(method, path, body);
    // Note: CSRF token is added in the request interceptor
    // to ensure proper handling of token refresh

    return {
      ...config,
      headers: {
        ...config?.headers,
        'X-FFLY-Timestamp': timestamp,
        'X-FFLY-Signature': signature,
      },
    };
  }
}

// Singleton instance
export const adminApiClient = new AdminAPIClient();

export default adminApiClient;
