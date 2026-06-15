import { getApiBaseUrl } from '@/lib/constants';
import { safeParse, ValidationResult } from '@/lib/validation-utils';
import { useApiReachableStore } from '@/stores/apiReachableStore';
import { tokenVault } from '@/utils/token-vault';
import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';
import { z } from 'zod';

class ApiClient {
  private client: AxiosInstance;
  private token: string | null = null;
  /** Prevents concurrent refresh attempts — the API uses token rotation,
   *  so a second refresh with the old token would 401 and trigger logout. */
  private refreshPromise: Promise<string | null> | null = null;
  /** Ensures initialization completes before any requests are sent */
  private initPromise: Promise<void> | null = null;

  constructor() {
    const baseURL = getApiBaseUrl();
    this.client = axios.create({
      baseURL: baseURL || window.location.origin,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Initialize token vault and load token — block subsequent requests until ready
    this.initPromise = this.initializeTokenVault();

    // Add request interceptor to include auth token and environment header
    this.client.interceptors.request.use(
      async (config) => {
        // Wait for token vault initialization before sending any request
        if (this.initPromise) {
          await this.initPromise;
        }

        // Get token from secure storage (TokenVault encrypted)
        const storedToken = await tokenVault.getAccessToken();

        if (storedToken) {
          this.token = storedToken;
          config.headers.Authorization = `Bearer ${this.token}`;
        }

        // Add X-Environment header from sidebar store (for API scoping)
        const environment = localStorage.getItem('ff-current-environment');
        if (environment) {
          config.headers['X-Environment'] = environment;
        }

        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // Add response interceptor to handle auth errors and track API reachability
    this.client.interceptors.response.use(
      (response) => {
        useApiReachableStore.getState().setApiReachable(true);
        return response;
      },
      async (error) => {
        const originalRequest = error.config;
        const status = error.response?.status;

        // Track API reachability - only mark as unreachable for network errors or 5xx
        if (!error.response || status >= 500) {
          useApiReachableStore.getState().setApiReachable(false);
        } else if (status < 500) {
          useApiReachableStore.getState().setApiReachable(true);
        }

        // Only attempt refresh once per request — _retry flag prevents infinite loops
        // when the retried request itself returns 401 (e.g. wrong tenant, revoked token).
        if (status === 401 && !originalRequest._retry) {
          const refreshToken = await tokenVault.getRefreshToken();
          if (refreshToken) {
            try {
              // Prevent concurrent refresh attempts — token rotation means
              // the second request with the old refresh token would 401.
              if (!this.refreshPromise) {
                this.refreshPromise = this._doRefresh(refreshToken);
              }
              const newToken = await this.refreshPromise;
              this.refreshPromise = null;

              if (newToken) {
                originalRequest._retry = true;
                originalRequest.headers.Authorization = `Bearer ${newToken}`;
                return this.client.request(originalRequest);
              } else {
                // Refresh returned null - token is invalid, clear session
                this._handleAuthFailure();
              }
            } catch (refreshError) {
              this.refreshPromise = null;
              console.warn('Token refresh failed after retries:', refreshError);
              // If refresh fails after all retries, clear session and log out
              this._handleAuthFailure();
            }
          } else {
            // No refresh token, clear session
            this._handleAuthFailure();
          }
        }
        return Promise.reject(error);
      }
    );
  }

  /**
   * Initialize the TokenVault and load existing tokens
   */
  private async initializeTokenVault(): Promise<void> {
    await tokenVault.initialize();
    const accessToken = await tokenVault.getAccessToken();
    if (accessToken) {
      this.token = accessToken;
    }
    // Clear initPromise so subsequent requests don't block unnecessarily
    this.initPromise = null;
  }

  // Helper method to handle auth failures
  private async _handleAuthFailure() {
    await this.clearToken();
    import('@/stores/authStore').then(({ useAuthStore }) => {
      useAuthStore.getState().logout(true);
    });
  }

  clearToken() {
    this.token = null;
    tokenVault.clearTokens();
    localStorage.removeItem('ff-last-wallet-agent-id');
  }

  clearCSRFToken() {
    localStorage.removeItem('ff-csrf-token');
  }

  private _csrfTokenPromise: Promise<string | null> | null = null;

  /**
   * Fetch a CSRF token for protected routes (billing, admin, etc.)
   * The backend stores CSRF tokens in Redis keyed by session ID.
   */
  async fetchCSRFToken(): Promise<string | null> {
    if (this._csrfTokenPromise) {
      return this._csrfTokenPromise;
    }
    this._csrfTokenPromise = this._doFetchCSRFToken();
    try {
      const token = await this._csrfTokenPromise;
      return token;
    } finally {
      this._csrfTokenPromise = null;
    }
  }

  private async _doFetchCSRFToken(): Promise<string | null> {
    try {
      const response = await this.client.get<{ token: string; expires_at: string }>('/v1/csrf');
      return response.data.token;
    } catch (error) {
      console.warn('Failed to fetch CSRF token:', error);
      return null;
    }
  }

  /**
   * Fetch CSRF token with automatic retry on auth failure.
   * If the first attempt fails with 401 (e.g. token was invalidated),
   * we retry once after token refresh.
   */
  async fetchCSRFTokenWithRetry(): Promise<string | null> {
    const refreshToken = await tokenVault.getRefreshToken();
    if (!refreshToken) {
      return this.fetchCSRFToken();
    }

    try {
      const token = await this._doFetchCSRFToken();
      if (token) {
        return token;
      }

      // Token might be invalid (401 from token refresh) - try refreshing
      const newToken = await this._doRefresh(refreshToken);
      if (newToken) {
        // Token refreshed, retry fetching CSRF
        return this._doFetchCSRFToken();
      }
    } catch (e) {
      console.warn('CSRF token fetch with retry failed:', e);
    }
    return null;
  }

  loadToken() {
    // Token is loaded async via TokenVault during initialization
    // This method is kept for compatibility but is a no-op
  }

  getToken() {
    return this.token;
  }

  async reloadToken() {
    await tokenVault.initialize();
    const accessToken = await tokenVault.getAccessToken();
    if (accessToken) {
      this.token = accessToken;
    }
  }

  async checkTokenInStorage() {
    return await tokenVault.getAccessToken();
  }

  private async _doRefresh(refreshToken: string): Promise<string | null> {
    const apiUrl = getApiBaseUrl();
    const maxRetries = 3;

    for (let attempt = 0; attempt < maxRetries; attempt++) {
      try {
        const refreshResponse = await fetch(`${apiUrl}/v1/auth/refresh`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ refresh_token: refreshToken }),
        });

        if (refreshResponse.ok) {
          const refreshData = await refreshResponse.json();
          
          // Store new tokens in encrypted storage
          await tokenVault.setAccessToken(refreshData.token);
          await tokenVault.setRefreshToken(refreshData.refresh_token);
          
          this.token = refreshData.token;
          this.clearCSRFToken();
          return refreshData.token;
        }

        // If we got a 4xx error (except 429 rate limit), token is invalid - don't retry
        if (
          refreshResponse.status >= 400 &&
          refreshResponse.status < 500 &&
          refreshResponse.status !== 429
        ) {
          console.warn(`Token refresh failed with status ${refreshResponse.status}, clearing auth`);
          return null;
        }

        // For 5xx errors or 429 rate limit, retry with backoff
        if (attempt < maxRetries - 1) {
          const delayMs = Math.min(1000 * Math.pow(2, attempt), 10000); // Cap at 10 seconds
          await new Promise((resolve) => setTimeout(resolve, delayMs));
        }
      } catch (error) {
        // Network error or other exception - retry with backoff
        if (attempt < maxRetries - 1) {
          const delayMs = Math.min(1000 * Math.pow(2, attempt), 10000);
          console.warn(
            `Token refresh attempt ${attempt + 1} failed, retrying in ${delayMs}ms:`,
            error
          );
          await new Promise((resolve) => setTimeout(resolve, delayMs));
        } else {
          console.error('Token refresh failed after all retries:', error);
          return null;
        }
      }
    }

    return null;
  }

  async get<T>(url: string, config?: AxiosRequestConfig) {
    const response = await this.client.get<T>(url, config);
    return response.data;
  }

  async post<T>(url: string, data?: unknown, config?: AxiosRequestConfig) {
    const response = await this.client.post<T>(url, data, config);
    return response.data;
  }

  async put<T>(url: string, data?: unknown, config?: AxiosRequestConfig) {
    const response = await this.client.put<T>(url, data, config);
    return response.data;
  }

  async patch<T>(url: string, data?: unknown, config?: AxiosRequestConfig) {
    const response = await this.client.patch<T>(url, data, config);
    return response.data;
  }

  async delete<T>(url: string, config?: AxiosRequestConfig) {
    const response = await this.client.delete<T>(url, config);
    return response.data;
  }

  // Validated methods that parse responses with Zod schemas
  async getValidated<T>(
    schema: z.ZodType<T>,
    url: string,
    config?: AxiosRequestConfig,
    fallback?: T
  ): Promise<ValidationResult<T>> {
    try {
      const response = await this.client.get(url, config);
      return safeParse(schema, response.data, fallback, `GET ${url}`);
    } catch (error) {
      console.error(`API validation error for GET ${url}:`, error);
      // Return fallback if provided, otherwise return error result
      if (fallback !== undefined) {
        return { success: false, data: fallback, fallbackUsed: true, error: 'API request failed' };
      }
      return { success: false, error: 'API request failed' };
    }
  }

  async postValidated<T>(
    schema: z.ZodType<T>,
    url: string,
    data?: unknown,
    config?: AxiosRequestConfig,
    fallback?: T
  ): Promise<ValidationResult<T>> {
    try {
      const response = await this.client.post(url, data, config);
      return safeParse(schema, response.data, fallback, `POST ${url}`);
    } catch (error) {
      console.error(`API validation error for POST ${url}:`, error);
      if (fallback !== undefined) {
        return { success: false, data: fallback, fallbackUsed: true, error: 'API request failed' };
      }
      return { success: false, error: 'API request failed' };
    }
  }

  async putValidated<T>(
    schema: z.ZodType<T>,
    url: string,
    data?: unknown,
    config?: AxiosRequestConfig,
    fallback?: T
  ): Promise<ValidationResult<T>> {
    try {
      const response = await this.client.put(url, data, config);
      return safeParse(schema, response.data, fallback, `PUT ${url}`);
    } catch (error) {
      console.error(`API validation error for PUT ${url}:`, error);
      if (fallback !== undefined) {
        return { success: false, data: fallback, fallbackUsed: true, error: 'API request failed' };
      }
      return { success: false, error: 'API request failed' };
    }
  }

  async patchValidated<T>(
    schema: z.ZodType<T>,
    url: string,
    data?: unknown,
    config?: AxiosRequestConfig,
    fallback?: T
  ): Promise<ValidationResult<T>> {
    try {
      const response = await this.client.patch(url, data, config);
      return safeParse(schema, response.data, fallback, `PATCH ${url}`);
    } catch (error) {
      console.error(`API validation error for PATCH ${url}:`, error);
      if (fallback !== undefined) {
        return { success: false, data: fallback, fallbackUsed: true, error: 'API request failed' };
      }
      return { success: false, error: 'API request failed' };
    }
  }

  async deleteValidated<T>(
    schema: z.ZodType<T>,
    url: string,
    config?: AxiosRequestConfig,
    fallback?: T
  ): Promise<ValidationResult<T>> {
    try {
      const response = await this.client.delete(url, config);
      return safeParse(schema, response.data, fallback, `DELETE ${url}`);
    } catch (error) {
      console.error(`API validation error for DELETE ${url}:`, error);
      if (fallback !== undefined) {
        return { success: false, data: fallback, fallbackUsed: true, error: 'API request failed' };
      }
      return { success: false, error: 'API request failed' };
    }
  }

  // Convenience methods that return validated data directly (throw on validation failure)
  async getValidatedData<T>(
    schema: z.ZodType<T>,
    url: string,
    config?: AxiosRequestConfig
  ): Promise<T> {
    const result = await this.getValidated(schema, url, config);
    if (!result.success || result.data === undefined) {
      throw new Error(result.error || 'Validation failed');
    }
    return result.data as T;
  }

  async postValidatedData<T>(
    schema: z.ZodType<T>,
    url: string,
    data?: unknown,
    config?: AxiosRequestConfig
  ): Promise<T> {
    const result = await this.postValidated(schema, url, data, config);
    if (!result.success || result.data === undefined) {
      throw new Error(result.error || 'Validation failed');
    }
    return result.data as T;
  }

  async putValidatedData<T>(
    schema: z.ZodType<T>,
    url: string,
    data?: unknown,
    config?: AxiosRequestConfig
  ): Promise<T> {
    const result = await this.putValidated(schema, url, data, config);
    if (!result.success || result.data === undefined) {
      throw new Error(result.error || 'Validation failed');
    }
    return result.data as T;
  }

  async patchValidatedData<T>(
    schema: z.ZodType<T>,
    url: string,
    data?: unknown,
    config?: AxiosRequestConfig
  ): Promise<T> {
    const result = await this.patchValidated(schema, url, data, config);
    if (!result.success || result.data === undefined) {
      throw new Error(result.error || 'Validation failed');
    }
    return result.data as T;
  }

  async deleteValidatedData<T>(
    schema: z.ZodType<T>,
    url: string,
    config?: AxiosRequestConfig
  ): Promise<T> {
    const result = await this.deleteValidated(schema, url, config);
    if (!result.success || result.data === undefined) {
      throw new Error(result.error || 'Validation failed');
    }
    return result.data as T;
  }

  // Get the base URL for the API (used for EventSource streaming)
  getBaseUrl(): string {
    return this.client.defaults.baseURL || window.location.origin;
  }
}

export const apiClient = new ApiClient();
