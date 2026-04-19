import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { providersApi, type ConnectProviderRequest } from '@/api/providers';

// Query keys
export const providerKeys = {
  all: ['providers'] as const,
  connected: () => [...providerKeys.all, 'connected'] as const,
};

// Get connected providers
export function useConnectedProviders() {
  return useQuery({
    queryKey: providerKeys.connected(),
    queryFn: () => providersApi.getConnectedProviders(),
    staleTime: 1000 * 60,
  });
}

// Connect provider
export function useConnectProvider() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: ConnectProviderRequest) =>
      providersApi.connectProvider(request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: providerKeys.connected() });
      toast.success('Provider connected successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to connect provider: ${error.message}`);
    },
  });
}

// Disconnect provider
export function useDisconnectProvider() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (providerId: string) =>
      providersApi.disconnectProvider(providerId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: providerKeys.connected() });
      toast.success('Provider disconnected');
    },
    onError: (error: Error) => {
      toast.error(`Failed to disconnect provider: ${error.message}`);
    },
  });
}

// Test connection
export function useTestProviderConnection() {
  return useMutation({
    mutationFn: (providerId: string) => providersApi.testConnection(providerId),
    onSuccess: (data) => {
      if (data.success) {
        toast.success(data.message || 'Connection test successful');
      } else {
        toast.error(data.message || 'Connection test failed');
      }
    },
    onError: (error: Error) => {
      toast.error(`Connection test failed: ${error.message}`);
    },
  });
}
