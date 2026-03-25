/**
 * Public API Client (non-admin)
 * Used for routes served under /v1 (not /v1/admin)
 */

import { CACHE_KEYS, getApiBaseUrl } from '@/lib/constants';
import type { AdminAPIResponse } from '@/types';
import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';

class PublicAPIClient {
  private client: AxiosInstance;
  private sessionToken: string | null = null;

  constructor() {
    const baseURL = `${getApiBaseUrl()}/v1`;
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

    this.client.interceptors.request.use((config) => {
      if (this.sessionToken) {
        config.headers.Authorization = `Bearer ${this.sessionToken}`;
      }
      return config;
    });
  }

  async get<T>(path: string, config?: AxiosRequestConfig): Promise<AdminAPIResponse<T>> {
    const response = await this.client.get<AdminAPIResponse<T>>(path, config);
    return response.data;
  }
}

export const publicApiClient = new PublicAPIClient();
export default publicApiClient;
