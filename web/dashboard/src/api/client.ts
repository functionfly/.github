import axios, { AxiosInstance, AxiosRequestConfig } from "axios";
import { ZodSchema } from "zod";
import { getApiBaseUrl } from "@/lib/constants";
import { safeParse, ValidationResult } from "@/lib/validation-utils";

class ApiClient {
  private client: AxiosInstance;
  private token: string | null = null;

  constructor() {
    const baseURL = getApiBaseUrl();
    this.client = axios.create({
      baseURL: baseURL || window.location.origin,
      headers: {
        "Content-Type": "application/json",
      },
    });

    // Add request interceptor to include auth token
    this.client.interceptors.request.use(
      (config) => {
        // Always get the latest token from localStorage
        const storedToken = localStorage.getItem("ff-access-token");

        if (storedToken) {
          this.token = storedToken;
          config.headers.Authorization = `Bearer ${this.token}`;
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // Add response interceptor to handle auth errors
    this.client.interceptors.response.use(
      (response) => response,
      async (error) => {
        if (error.response?.status === 401) {
          // Token is invalid or expired - try to refresh first
          const refreshToken = localStorage.getItem("ff-refresh-token");
          if (refreshToken) {
            try {
              console.log('Attempting token refresh...');
              const apiUrl = getApiBaseUrl();
              const refreshResponse = await fetch(`${apiUrl}/v1/auth/refresh`, {
                method: 'POST',
                headers: {
                  'Content-Type': 'application/json',
                },
                body: JSON.stringify({ refresh_token: refreshToken }),
              });

              if (refreshResponse.ok) {
                const refreshData = await refreshResponse.json();

                // Store new tokens
                localStorage.setItem('ff-access-token', refreshData.token);
                localStorage.setItem('ff-refresh-token', refreshData.refresh_token);

                // Update the request with new token and retry
                const newToken = refreshData.token;
                this.token = newToken;
                error.config.headers.Authorization = `Bearer ${newToken}`;

                // Retry the original request
                return this.client.request(error.config);
              }
            } catch (refreshError) {
              console.warn('Token refresh failed:', refreshError);
            }
          }

          // If refresh failed or no refresh token, clear auth state
          console.log('Token refresh failed or no refresh token, logging out user');
          localStorage.removeItem("ff-access-token");
          localStorage.removeItem("ff-refresh-token");

          // Import auth store dynamically to avoid circular dependencies
          import('@/stores/authStore').then(({ useAuthStore }) => {
            useAuthStore.getState().logout();
          });
        }
        return Promise.reject(error);
      }
    );

    // Load token on initialization
    this.loadToken();
  }

  clearToken() {
    this.token = null;
    localStorage.removeItem("ff-access-token");
    localStorage.removeItem("ff-refresh-token");
  }

  loadToken() {
    const token = localStorage.getItem("ff-access-token");
    if (token) {
      this.token = token;
    }
  }

  getToken() {
    return this.token;
  }

  reloadToken() {
    this.loadToken();
  }

  checkTokenInStorage() {
    return localStorage.getItem("ff-access-token");
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
    schema: ZodSchema<T>,
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
    schema: ZodSchema<T>,
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
    schema: ZodSchema<T>,
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
    schema: ZodSchema<T>,
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
    schema: ZodSchema<T>,
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
    schema: ZodSchema<T>,
    url: string,
    config?: AxiosRequestConfig
  ): Promise<T> {
    const result = await this.getValidated(schema, url, config);
    if (!result.success || result.data === undefined) {
      throw new Error(result.error || 'Validation failed');
    }
    return result.data;
  }

  async postValidatedData<T>(
    schema: ZodSchema<T>,
    url: string,
    data?: unknown,
    config?: AxiosRequestConfig
  ): Promise<T> {
    const result = await this.postValidated(schema, url, data, config);
    if (!result.success || result.data === undefined) {
      throw new Error(result.error || 'Validation failed');
    }
    return result.data;
  }

  async putValidatedData<T>(
    schema: ZodSchema<T>,
    url: string,
    data?: unknown,
    config?: AxiosRequestConfig
  ): Promise<T> {
    const result = await this.putValidated(schema, url, data, config);
    if (!result.success || result.data === undefined) {
      throw new Error(result.error || 'Validation failed');
    }
    return result.data;
  }

  async patchValidatedData<T>(
    schema: ZodSchema<T>,
    url: string,
    data?: unknown,
    config?: AxiosRequestConfig
  ): Promise<T> {
    const result = await this.patchValidated(schema, url, data, config);
    if (!result.success || result.data === undefined) {
      throw new Error(result.error || 'Validation failed');
    }
    return result.data;
  }

  async deleteValidatedData<T>(
    schema: ZodSchema<T>,
    url: string,
    config?: AxiosRequestConfig
  ): Promise<T> {
    const result = await this.deleteValidated(schema, url, config);
    if (!result.success || result.data === undefined) {
      throw new Error(result.error || 'Validation failed');
    }
    return result.data;
  }
}

export const apiClient = new ApiClient();
