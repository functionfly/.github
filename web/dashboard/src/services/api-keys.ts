import { apiClient } from "@/api/client";
import type {
  APIKey,
  APIKeyCreateResponse,
  APIKeyListResponse,
  APIKeyPermission,
  APIKeyEnvironment,
  APIKeyRotation,
  CreateAPIKeyRequest,
  UpdateAPIKeyRequest,
  RotateAPIKeyRequest,
  AddPermissionRequest,
  AddEnvironmentRequest,
  APIKeyFilters,
  PaginatedResponse,
} from "@/types/api-key";

const BASE_URL = "/v1/api-keys";

/**
 * API Key Service
 * Handles all API key CRUD operations and related functionality
 */
export const apiKeysService = {
  /**
   * Create a new API key
   */
  createKey: async (data: CreateAPIKeyRequest): Promise<APIKeyCreateResponse> => {
    const response = await apiClient.post<APIKeyCreateResponse>(BASE_URL, data);
    return response;
  },

  /**
   * List all API keys with optional filtering and pagination
   */
  listKeys: async (
    filters?: APIKeyFilters,
    page = 1,
    pageSize = 10
  ): Promise<APIKeyListResponse> => {
    const params = new URLSearchParams();
    params.append("page", String(page));
    params.append("page_size", String(pageSize));

    if (filters?.key_type) {
      params.append("key_type", filters.key_type);
    }
    if (filters?.is_active !== undefined) {
      params.append("is_active", String(filters.is_active));
    }
    if (filters?.expires_before) {
      params.append("expires_before", filters.expires_before);
    }
    if (filters?.expires_after) {
      params.append("expires_after", filters.expires_after);
    }
    if (filters?.search) {
      params.append("search", filters.search);
    }

    const response = await apiClient.get<APIKeyListResponse>(
      `${BASE_URL}?${params.toString()}`
    );
    return response;
  },

  /**
   * Get a single API key by ID
   */
  getKey: async (id: string): Promise<APIKey> => {
    const response = await apiClient.get<APIKey>(`${BASE_URL}/${id}`);
    return response;
  },

  /**
   * Update an existing API key
   */
  updateKey: async (id: string, data: UpdateAPIKeyRequest): Promise<APIKey> => {
    const response = await apiClient.patch<APIKey>(`${BASE_URL}/${id}`, data);
    return response;
  },

  /**
   * Delete (deactivate) an API key
   */
  deleteKey: async (id: string): Promise<void> => {
    await apiClient.delete(`${BASE_URL}/${id}`);
  },

  /**
   * Rotate an API key
   * Returns the new key with plaintext
   */
  rotateKey: async (id: string, data?: RotateAPIKeyRequest): Promise<APIKeyCreateResponse> => {
    const response = await apiClient.post<APIKeyCreateResponse>(
      `${BASE_URL}/${id}/rotate`,
      data || {}
    );
    return response;
  },

  // Permissions

  /**
   * Get permissions for an API key
   */
  getPermissions: async (keyId: string): Promise<APIKeyPermission[]> => {
    const response = await apiClient.get<APIKeyPermission[]>(
      `${BASE_URL}/${keyId}/permissions`
    );
    return response;
  },

  /**
   * Add a permission to an API key
   */
  addPermission: async (keyId: string, data: AddPermissionRequest): Promise<APIKeyPermission> => {
    const response = await apiClient.post<APIKeyPermission>(
      `${BASE_URL}/${keyId}/permissions`,
      data
    );
    return response;
  },

  /**
   * Remove a permission from an API key
   */
  removePermission: async (keyId: string, permissionId: string): Promise<void> => {
    await apiClient.delete(`${BASE_URL}/${keyId}/permissions/${permissionId}`);
  },

  // Environments

  /**
   * Get environments linked to an API key
   */
  getEnvironments: async (keyId: string): Promise<APIKeyEnvironment[]> => {
    const response = await apiClient.get<APIKeyEnvironment[]>(
      `${BASE_URL}/${keyId}/environments`
    );
    return response;
  },

  /**
   * Link an environment to an API key
   */
  linkEnvironment: async (keyId: string, data: AddEnvironmentRequest): Promise<APIKeyEnvironment> => {
    const response = await apiClient.post<APIKeyEnvironment>(
      `${BASE_URL}/${keyId}/environments`,
      data
    );
    return response;
  },

  /**
   * Unlink an environment from an API key
   */
  unlinkEnvironment: async (keyId: string, environmentId: string): Promise<void> => {
    await apiClient.delete(`${BASE_URL}/${keyId}/environments/${environmentId}`);
  },

  // Rotation History

  /**
   * Get rotation history for an API key
   */
  getRotationHistory: async (keyId: string): Promise<APIKeyRotation[]> => {
    const response = await apiClient.get<APIKeyRotation[]>(
      `${BASE_URL}/${keyId}/rotations`
    );
    return response;
  },
};

/**
 * Authentication with API Key
 * Used for authenticating requests using an API key
 */
export const authenticateWithKey = async (apiKey: string): Promise<{ token: string }> => {
  const response = await apiClient.post<{ token: string }>("/v1/auth/api-key", {
    api_key: apiKey,
  });
  return response;
};

/**
 * Local storage helpers for newly created API keys
 */
export const API_KEY_STORAGE_KEY = "ff_new_api_key";

export const storeNewApiKey: (key: APIKeyCreateResponse) => void = (key) => {
  try {
    localStorage.setItem(API_KEY_STORAGE_KEY, JSON.stringify({
      key,
      createdAt: new Date().toISOString(),
    }));
  } catch (error) {
    console.error("Failed to store API key:", error);
  }
};

export const getStoredApiKey: () => { key: APIKeyCreateResponse; createdAt: string } | null = () => {
  try {
    const stored = localStorage.getItem(API_KEY_STORAGE_KEY);
    if (!stored) return null;
    const parsed = JSON.parse(stored);

    // Clear after retrieving (one-time display)
    localStorage.removeItem(API_KEY_STORAGE_KEY);

    return parsed;
  } catch (error) {
    console.error("Failed to retrieve API key:", error);
    return null;
  }
};

export const clearStoredApiKey: () => void = () => {
  try {
    localStorage.removeItem(API_KEY_STORAGE_KEY);
  } catch (error) {
    console.error("Failed to clear API key:", error);
  }
};
