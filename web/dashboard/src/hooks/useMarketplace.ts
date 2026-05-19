import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { marketplaceApi, type Extension, type MarketplaceFilters, type CreateExtensionRequest, type UpdateExtensionRequest } from '@/api/marketplace';

export { type Extension, type MarketplaceFilters, type CreateExtensionRequest, type UpdateExtensionRequest };

export const marketplaceKeys = {
  all: ['marketplace'] as const,
  lists: () => [...marketplaceKeys.all, 'list'] as const,
  list: (filters: MarketplaceFilters) => [...marketplaceKeys.lists(), filters] as const,
  details: () => [...marketplaceKeys.all, 'detail'] as const,
  detail: (id: string) => [...marketplaceKeys.details(), id] as const,
  categories: () => [...marketplaceKeys.all, 'categories'] as const,
};

export function useExtensions(filters?: MarketplaceFilters) {
  return useQuery({
    queryKey: marketplaceKeys.list(filters || {}),
    queryFn: () => marketplaceApi.list(filters),
    staleTime: 1000 * 60,
  });
}

export function useExtension(extensionId: string) {
  return useQuery({
    queryKey: marketplaceKeys.detail(extensionId),
    queryFn: () => marketplaceApi.get(extensionId),
    enabled: !!extensionId,
  });
}

export function useCategories() {
  return useQuery({
    queryKey: marketplaceKeys.categories(),
    queryFn: () => marketplaceApi.getCategories(),
    staleTime: 1000 * 60 * 5,
  });
}

export function useCreateExtension() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateExtensionRequest) => marketplaceApi.create(data),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: marketplaceKeys.all });
      toast.success(`Extension "${result.extension.name}" created`);
    },
    onError: (error: Error) => {
      toast.error(`Failed to create extension: ${error.message}`);
    },
  });
}

export function useUpdateExtension() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateExtensionRequest }) =>
      marketplaceApi.update(id, data),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: marketplaceKeys.all });
      queryClient.invalidateQueries({ queryKey: marketplaceKeys.detail(id) });
      toast.success('Extension updated');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update extension: ${error.message}`);
    },
  });
}

export function useDeleteExtension() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (extensionId: string) => marketplaceApi.delete(extensionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: marketplaceKeys.all });
      toast.success('Extension deleted');
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete extension: ${error.message}`);
    },
  });
}

export function useInstallExtension() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (extensionId: string) => marketplaceApi.install(extensionId),
    onSuccess: (_, extensionId) => {
      queryClient.invalidateQueries({ queryKey: marketplaceKeys.all });
      toast.success('Extension installed');
    },
    onError: (error: Error) => {
      toast.error(`Failed to install extension: ${error.message}`);
    },
  });
}