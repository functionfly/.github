/**
 * Admin API Client
 * Handles all API communication with HMAC signing for sensitive operations
 */

import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';
import { HMACRequestSigner } from './hmacSigner';
import type { AdminAPIResponse } from '@/types';
import { getAdminApiBaseUrl } from '@/lib/constants';
import { getCsrfToken } from '@/lib/security/csrf';

class AdminAPIClient {
  private client: AxiosInstance;
  private signer: HMACRequestSigner;
  private sessionToken: string | null = null;
  private deviceFingerprint: string | null = null;

  constructor() {
    const baseURL = getAdminApiBaseUrl();

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
    this.client.interceptors.request.use((config) => {
      if (this.sessionToken) {
        config.headers.Authorization = `Bearer ${this.sessionToken}`;
      }

      if (this.deviceFingerprint) {
        config.headers['X-Device-Fingerprint'] = this.deviceFingerprint;
      }

      return config;
    });

    // Add response interceptor for error handling
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          // Session expired, redirect to login
          window.location.href = '/auth/login?reason=session_expired';
        }
        return Promise.reject(error);
      }
    );
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
  async get<T>(
    path: string,
    config?: AxiosRequestConfig
  ): Promise<AdminAPIResponse<T>> {
    const response = await this.client.get<AdminAPIResponse<T>>(
      path,
      config
    );
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
    const response = await this.client.post<AdminAPIResponse<T>>(
      path,
      data,
      signedConfig
    );
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
    const response = await this.client.patch<AdminAPIResponse<T>>(
      path,
      data,
      signedConfig
    );
    return response.data;
  }

  /**
   * DELETE request with HMAC signing
   */
  async delete<T>(
    path: string,
    config?: AxiosRequestConfig
  ): Promise<AdminAPIResponse<T>> {
    const signedConfig = this.signRequest('DELETE', path, undefined, config);
    const response = await this.client.delete<AdminAPIResponse<T>>(
      path,
      signedConfig
    );
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
    const csrfToken = getCsrfToken();

    return {
      ...config,
      headers: {
        ...config?.headers,
        'X-FFLY-Timestamp': timestamp,
        'X-FFLY-Signature': signature,
        ...(csrfToken ? { 'X-CSRF-Token': csrfToken } : {}),
      },
    };
  }
}

// Singleton instance
export const adminApiClient = new AdminAPIClient();

export default adminApiClient;
