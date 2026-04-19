import type { App, CreateAppRequest } from '@/types';
import { useApps as useAppsQuery, useCreateApp } from '@/hooks';
import { useCallback, useState } from 'react';
import { toast } from 'sonner';

export type SortOption = 'name-asc' | 'name-desc' | 'created-asc' | 'created-desc';

export interface UseAppsOptions {
  initialSearch?: string;
  initialSort?: SortOption;
}

export function useApps({ initialSearch = '', initialSort = 'created-desc' }: UseAppsOptions = {}) {
  const [searchQuery, setSearchQuery] = useState(initialSearch);
  const [sortOption, setSortOption] = useState<SortOption>(initialSort);

  // Use the main hooks instead of raw queries
  const { data, isLoading, error, refetch } = useAppsQuery();
  const createMutation = useCreateApp();

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

  const apps = data?.apps ?? [];

  const filteredApps = apps.length > 0
    ? sortApps(
        apps.filter(
          (app) =>
            app.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
            app.slug.toLowerCase().includes(searchQuery.toLowerCase())
        )
      )
    : [];

  const createApp = useCallback(
    (data: CreateAppRequest) => {
      createMutation.mutate(data, {
        onError: (err: Error) => {
          toast.error(err.message || 'Failed to create app');
        },
      });
    },
    [createMutation]
  );

  return {
    apps: filteredApps,
    allApps: apps,
    isLoading,
    error,
    refetch,
    searchQuery,
    setSearchQuery,
    sortOption,
    setSortOption,
    createApp,
    createAppAsync: createMutation.mutateAsync,
    isCreating: createMutation.isPending,
    // Note: delete endpoint not yet implemented in the API
    deleteApp: (_appId: string) => {
      toast.error('Delete not yet implemented');
      throw new Error('Delete not yet implemented');
    },
    isDeleting: false,
  };
}
