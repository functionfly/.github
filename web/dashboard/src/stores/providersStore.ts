import { dedupeConnectedProvidersBySlug, providersApi } from '@/api';
import type { ConnectedProvider, ConnectProviderRequest } from '@/types';
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

// Exponential backoff helper for retries
async function retryWithBackoff<T>(
  operation: () => Promise<T>,
  maxRetries = 3,
  baseDelay = 1000
): Promise<T> {
  let lastError: Error | undefined;

  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      return await operation();
    } catch (error) {
      lastError = error instanceof Error ? error : new Error(String(error));

      // Don't retry on 4xx errors (client errors)
      if (error && typeof error === 'object' && 'response' in error) {
        const axiosError = error as { response?: { status?: number } };
        const status = axiosError.response?.status;
        if (status && status >= 400 && status < 500 && status !== 429) {
          throw lastError;
        }
      }

      if (attempt < maxRetries - 1) {
        const delay = Math.min(baseDelay * Math.pow(2, attempt), 10000);
        await new Promise((resolve) => setTimeout(resolve, delay));
      }
    }
  }

  throw lastError;
}

interface ConnectProviderResult {
  success: boolean;
  apiKey?: string;
  apiKeyId?: string;
}

interface ProvidersState {
  providers: ConnectedProvider[];
  isLoading: boolean;
  error: string | null;
  lastFetchTime: number | null;
  isPolling: boolean;

  // Actions
  fetchProviders: (forceRefresh?: boolean) => Promise<void>;
  connectProvider: (request: ConnectProviderRequest) => Promise<ConnectProviderResult>;
  disconnectProvider: (providerId: string) => Promise<void>;
  testConnection: (providerId: string) => Promise<boolean>;
  clearError: () => void;
  refreshProviderStatus: (providerId: string) => Promise<void>;
  startHealthCheckPolling: (intervalMs?: number) => () => void;
  stopHealthCheckPolling: () => void;
}

// Cache duration: 30 seconds for provider list
const CACHE_DURATION = 30000;

// Health check polling interval: 5 minutes by default
const DEFAULT_POLLING_INTERVAL = 5 * 60 * 1000;

// Global state for polling interval ID (outside the store to persist across component unmounts)
let pollingIntervalId: ReturnType<typeof setInterval> | null = null;

export const useProvidersStore = create<ProvidersState>()(
  persist(
    (set, get) => ({
      providers: [],
      isLoading: false,
      error: null,
      lastFetchTime: null,
      isPolling: false,

      fetchProviders: async (forceRefresh = false) => {
        const { lastFetchTime } = get();
        const now = Date.now();

        // Use cache if available and not forced refresh
        if (!forceRefresh && lastFetchTime && now - lastFetchTime < CACHE_DURATION) {
          return;
        }

        set({ isLoading: true, error: null });
        try {
          const providers = await retryWithBackoff(() => providersApi.getConnectedProviders());
          set({ providers, isLoading: false, lastFetchTime: now });
        } catch (error) {
          set({
            error: error instanceof Error ? error.message : 'Failed to fetch providers',
            isLoading: false,
          });
        }
      },

      connectProvider: async (request) => {
        set({ isLoading: true, error: null });
        try {
          const response = await retryWithBackoff(() => providersApi.connectProvider(request));
          const { providers } = get();
          // Avoid duplicates if already connected (e.g. re-enabling)
          const filtered = providers.filter((p) => p.id !== response.provider.id);
          set({
            providers: dedupeConnectedProvidersBySlug([...filtered, response.provider]),
            isLoading: false,
          });
          return {
            success: true,
            apiKey: response.apiKey,
            apiKeyId: response.apiKeyId,
          };
        } catch (error) {
          set({
            error: error instanceof Error ? error.message : 'Failed to connect provider',
            isLoading: false,
          });
          throw error;
        }
      },

      disconnectProvider: async (providerId) => {
        set({ isLoading: true, error: null });
        try {
          await providersApi.disconnectProvider(providerId);
          const { providers } = get();
          set({
            providers: providers.filter((p) => p.id !== providerId),
            isLoading: false,
          });
        } catch (error) {
          set({
            error: error instanceof Error ? error.message : 'Failed to disconnect provider',
            isLoading: false,
          });
          throw error;
        }
      },

      testConnection: async (providerId) => {
        try {
          const result = await providersApi.testConnection(providerId);
          return result.success;
        } catch (error) {
          return false;
        }
      },

      refreshProviderStatus: async (providerId: string) => {
        try {
          const isHealthy = await providersApi.testConnection(providerId);
          const { providers } = get();

          // Update the provider status in the list
          const updatedProviders = providers.map((p) =>
            p.name === providerId
              ? { ...p, status: (isHealthy ? 'online' : 'degraded') as ConnectedProvider['status'] }
              : p
          );

          set({ providers: updatedProviders });
        } catch (error) {
          console.warn('Failed to refresh provider status:', error);
        }
      },

      startHealthCheckPolling: (intervalMs = DEFAULT_POLLING_INTERVAL) => {
        const { isPolling } = get();

        // Don't start if already polling
        if (isPolling || pollingIntervalId !== null) {
          return () => {};
        }

        set({ isPolling: true });
        console.log('[ProvidersStore] Starting health check polling');

        // Immediately check all providers on start
        const { providers } = get();
        providers.forEach((provider) => {
          get().refreshProviderStatus(provider.name);
        });

        // Set up interval for periodic health checks
        pollingIntervalId = setInterval(() => {
          const { providers: currentProviders, isPolling: stillPolling } = get();

          if (!stillPolling) {
            if (pollingIntervalId) {
              clearInterval(pollingIntervalId);
              pollingIntervalId = null;
            }
            return;
          }

          currentProviders.forEach((provider) => {
            get().refreshProviderStatus(provider.name);
          });
        }, intervalMs);

        // Return cleanup function
        return () => {
          get().stopHealthCheckPolling();
        };
      },

      stopHealthCheckPolling: () => {
        if (pollingIntervalId) {
          clearInterval(pollingIntervalId);
          pollingIntervalId = null;
          set({ isPolling: false });
          console.log('[ProvidersStore] Stopped health check polling');
        }
      },

      clearError: () => set({ error: null }),
    }),
    {
      name: 'providers-storage',
      partialize: (state) => ({ providers: state.providers }),
    }
  )
);
