import { create } from "zustand";
import { persist } from "zustand/middleware";
import { providersApi } from "@/api";
import type { ConnectedProvider, ConnectProviderRequest } from "@/types";

interface ProvidersState {
  providers: ConnectedProvider[];
  isLoading: boolean;
  error: string | null;

  // Actions
  fetchProviders: () => Promise<void>;
  connectProvider: (request: ConnectProviderRequest) => Promise<void>;
  disconnectProvider: (providerId: string) => Promise<void>;
  testConnection: (providerId: string) => Promise<boolean>;
  clearError: () => void;
}

export const useProvidersStore = create<ProvidersState>()(
  persist(
    (set, get) => ({
      providers: [],
      isLoading: false,
      error: null,

      fetchProviders: async () => {
        set({ isLoading: true, error: null });
        try {
          const providers = await providersApi.getConnectedProviders();
          set({ providers, isLoading: false });
        } catch (error) {
          set({
            error: error instanceof Error ? error.message : "Failed to fetch providers",
            isLoading: false,
          });
        }
      },

      connectProvider: async (request) => {
        set({ isLoading: true, error: null });
        try {
          const response = await providersApi.connectProvider(request);
          const { providers } = get();
          set({
            providers: [...providers, response.provider],
            isLoading: false,
          });
        } catch (error) {
          set({
            error: error instanceof Error ? error.message : "Failed to connect provider",
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
            error: error instanceof Error ? error.message : "Failed to disconnect provider",
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

      clearError: () => set({ error: null }),
    }),
    {
      name: "providers-storage",
      partialize: (state) => ({ providers: state.providers }),
    }
  )
);