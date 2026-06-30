import { apiClient } from '@/api/client';
import type {
  AIProviderKey,
  ConnectAIKeyRequest,
  ConnectAIKeyResponse,
  ListAIKeysResponse,
  ListSupportedProvidersResponse,
  RotateAIKeyRequest,
  RotateAIKeyResponse,
  SupportedProvider,
  TestAIKeyResponse,
} from '@/types/ai-keys';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

export const aiKeysQueryKeys = {
  all: ['ai-keys'] as const,
  list: () => [...aiKeysQueryKeys.all, 'list'] as const,
  providers: () => [...aiKeysQueryKeys.all, 'providers'] as const,
};

export const aiKeysApi = {
  async listKeys(): Promise<AIProviderKey[]> {
    const resp = await apiClient.get<ListAIKeysResponse>('/v1/ai-keys');
    return resp.keys ?? [];
  },

  async connectKey(request: ConnectAIKeyRequest): Promise<AIProviderKey> {
    const resp = await apiClient.post<ConnectAIKeyResponse>('/v1/ai-keys/connect', request);
    return resp.key;
  },

  async disconnectKey(provider: string): Promise<void> {
    await apiClient.delete(`/v1/ai-keys/${encodeURIComponent(provider)}`);
  },

  async testKey(provider: string): Promise<TestAIKeyResponse> {
    return await apiClient.post<TestAIKeyResponse>(
      `/v1/ai-keys/${encodeURIComponent(provider)}/test`
    );
  },

  async rotateKey(provider: string, request: RotateAIKeyRequest): Promise<AIProviderKey> {
    const resp = await apiClient.post<RotateAIKeyResponse>(
      `/v1/ai-keys/${encodeURIComponent(provider)}/rotate`,
      request
    );
    return resp.key;
  },

  async listSupportedProviders(): Promise<SupportedProvider[]> {
    const resp = await apiClient.get<ListSupportedProvidersResponse>('/v1/ai-keys/providers');
    return resp.providers ?? [];
  },
};

export function useAIKeys() {
  return useQuery({
    queryKey: aiKeysQueryKeys.list(),
    queryFn: aiKeysApi.listKeys,
    staleTime: 30_000,
  });
}

export function useSupportedProviders() {
  return useQuery({
    queryKey: aiKeysQueryKeys.providers(),
    queryFn: aiKeysApi.listSupportedProviders,
    staleTime: 5 * 60_000,
  });
}

export function useConnectAIKey() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: aiKeysApi.connectKey,
    onSuccess: (key) => {
      queryClient.invalidateQueries({ queryKey: aiKeysQueryKeys.list() });
      toast.success(`Connected ${key.provider} key (...${key.key_last4})`);
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Failed to connect key');
    },
  });
}

export function useDisconnectAIKey() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: aiKeysApi.disconnectKey,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: aiKeysQueryKeys.list() });
      toast.success('Key disconnected');
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Failed to disconnect key');
    },
  });
}

export function useTestAIKey() {
  return useMutation({
    mutationFn: aiKeysApi.testKey,
  });
}

export function useRotateAIKey() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ provider, apiKey }: { provider: string; apiKey: string }) =>
      aiKeysApi.rotateKey(provider, { apiKey }),
    onSuccess: (key) => {
      queryClient.invalidateQueries({ queryKey: aiKeysQueryKeys.list() });
      toast.success(`Rotated ${key.provider} key (...${key.key_last4})`);
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Failed to rotate key');
    },
  });
}
