import { apiClient } from './client';

/** Valid environment values for the environment selector */
export type Environment = 'production' | 'staging' | 'development';

export interface ActiveEnvironmentResponse {
  environment: Environment;
  available: Environment[];
}

export interface SetEnvironmentRequest {
  environment: Environment;
}

export interface SetEnvironmentResponse {
  message: string;
  environment: Environment;
}

/**
 * Service for managing the user's active environment preference.
 * The active environment is sent with every API request via the X-Environment header
 * to scope data to the selected environment (production, staging, or development).
 */
export const environmentService = {
  /**
   * Get the current active environment from the backend.
   * This is the source of truth for which environment the user is working in.
   */
  getActiveEnvironment: async (): Promise<ActiveEnvironmentResponse> => {
    return apiClient.get<ActiveEnvironmentResponse>('/v1/users/me/environment');
  },

  /**
   * Set the active environment preference on the backend.
   * This updates the user's settings and affects all subsequent API requests.
   */
  setActiveEnvironment: async (environment: Environment): Promise<SetEnvironmentResponse> => {
    return apiClient.patch<SetEnvironmentResponse>('/v1/users/me/environment', { environment });
  },
};
