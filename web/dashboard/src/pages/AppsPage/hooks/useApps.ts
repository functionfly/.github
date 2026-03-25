import { appsApi } from '@/api/apps';
import type { App, CreateAppRequest } from '@/types';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useCallback, useState } from 'react';
import { toast } from 'sonner';

export type SortOption = 'name-asc' | 'name-desc' | 'created-asc' | 'created-desc';

export interface UseAppsOptions {
  initialSearch?: string;
  initialSort?: SortOption;
}

export function useApps({ initialSearch = '', initialSort = 'created-desc' }: UseAppsOptions = {}) {
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState(initialSearch);
  const [sortOption, setSortOption] = useState<SortOption>(initialSort);

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['apps'],
    queryFn: async () => {
      const response = await appsApi.list();
      return response.apps;
    },
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateAppRequest) => appsApi.create(data),
    onSuccess: (app: App) => {
      queryClient.invalidateQueries({ queryKey: ['apps'] });
      toast.success(`App "${app.name}" created successfully`);
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Failed to create app');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (_appId: string) => {
      // Note: delete endpoint not yet implemented in the API
      throw new Error('Delete not yet implemented');
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['apps'] });
      toast.success('App deleted successfully');
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Failed to delete app');
    },
  });

  const sortApps = useCallback(
    (apps: App[]): App[] => {
      return [...apps].sort((a, b) => {
        switch (sortOption) {
          case 'name-asc':
            return a.name.localeCompare(b.name);
          case 'name-desc':
            return b.name.localeCompare(a.name);
          case 'created-asc':
            return new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime();
          case 'created-desc':
            return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime();
          default:
            return 0;
        }
      });
    },
    [sortOption]
  );

  const filteredApps = data
    ? sortApps(
        data.filter(
          (app) =>
            app.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
            app.slug.toLowerCase().includes(searchQuery.toLowerCase())
        )
      )
    : [];

  return {
    apps: filteredApps,
    allApps: data ?? [],
    isLoading,
    error,
    refetch,
    searchQuery,
    setSearchQuery,
    sortOption,
    setSortOption,
    createApp: createMutation.mutate,
    createAppAsync: createMutation.mutateAsync,
    isCreating: createMutation.isPending,
    deleteApp: deleteMutation.mutate,
    isDeleting: deleteMutation.isPending,
  };
}
