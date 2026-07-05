import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { registryApi, type RegistryFunction, type RegistryFunctionVersion, type RegistryRatingRequest } from '@/api/registry';

// Query keys
export const registryKeys = {
  all: ['registry'] as const,
  functions: (params?: { category?: string; author?: string }) =>
    [...registryKeys.all, 'functions', params] as const,
  function: (author: string, name: string) => [...registryKeys.all, 'function', author, name] as const,
  versions: (author: string, name: string) => [...registryKeys.all, 'versions', author, name] as const,
  stats: (author: string, name: string) => [...registryKeys.all, 'stats', author, name] as const,
  search: (query: string, category?: string) => [...registryKeys.all, 'search', query, category] as const,
  reviews: (author: string, name: string, params?: { limit?: number; offset?: number }) =>
    [...registryKeys.all, 'reviews', author, name, params] as const,
  settings: (author: string, name: string) => [...registryKeys.all, 'settings', author, name] as const,
  replay: (executionId: string) => [...registryKeys.all, 'replay', executionId] as const,
};

// List/search functions
export function useRegistryFunctions(params?: {
  category?: string;
  author?: string;
  visibility?: string;
  tags?: string[];
  limit?: number;
  offset?: number;
}) {
  return useQuery({
    queryKey: registryKeys.functions(params),
    queryFn: () => registryApi.getFunctions(params),
    staleTime: 1000 * 60 * 5,
  });
}

// Get single function
export function useRegistryFunction(author: string, name: string) {
  return useQuery({
    queryKey: registryKeys.function(author, name),
    queryFn: () => registryApi.getFunction(author, name),
    enabled: !!author && !!name,
    staleTime: 1000 * 60 * 5,
  });
}

// Get function versions
export function useRegistryFunctionVersions(author: string, name: string) {
  return useQuery({
    queryKey: registryKeys.versions(author, name),
    queryFn: () => registryApi.getFunctionVersions(author, name),
    enabled: !!author && !!name,
    staleTime: 1000 * 60 * 5,
  });
}

// Get function stats
export function useRegistryFunctionStats(author: string, name: string) {
  return useQuery({
    queryKey: registryKeys.stats(author, name),
    queryFn: () => registryApi.getFunctionStats(author, name),
    enabled: !!author && !!name,
    staleTime: 1000 * 60 * 5,
  });
}

// Search functions
export function useRegistrySearch(query: string, category?: string, limit = 50) {
  return useQuery({
    queryKey: registryKeys.search(query, category),
    queryFn: () => registryApi.searchFunctions(query, category, limit),
    enabled: !!query,
    staleTime: 1000 * 60,
  });
}

// Execute function
export function useExecuteRegistryFunction() {
  return useMutation({
    mutationFn: ({
      author,
      name,
      request,
    }: {
      author: string;
      name: string;
      request: { input: unknown; version?: string };
    }) => registryApi.executeFunction(author, name, request),
    onError: (error: Error) => {
      toast.error(`Execution failed: ${error.message}`);
    },
  });
}

// Submit rating
export function useSubmitRegistryRating() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      author,
      name,
      rating,
    }: {
      author: string;
      name: string;
      rating: RegistryRatingRequest;
    }) => registryApi.submitRating(author, name, rating),
    onSuccess: (_, { author, name }) => {
      queryClient.invalidateQueries({ queryKey: registryKeys.stats(author, name) });
      toast.success('Rating submitted successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to submit rating: ${error.message}`);
    },
  });
}

// Get reviews
export function useRegistryReviews(
  author: string,
  name: string,
  params?: { limit?: number; offset?: number }
) {
  return useQuery({
    queryKey: registryKeys.reviews(author, name, params),
    queryFn: () => registryApi.listReviews(author, name, params),
    enabled: !!author && !!name,
    staleTime: 1000 * 60,
  });
}

// Submit review
export function useSubmitRegistryReview() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      author,
      name,
      stars,
      title,
      body,
    }: {
      author: string;
      name: string;
      stars: number;
      title?: string;
      body?: string;
    }) => registryApi.submitReview(author, name, { stars, title, body }),
    onSuccess: (_, { author, name }) => {
      queryClient.invalidateQueries({ queryKey: registryKeys.reviews(author, name) });
      toast.success('Review submitted successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to submit review: ${error.message}`);
    },
  });
}

// Publish function
export function usePublishRegistryFunction() {
  return useMutation({
    mutationFn: (data: Parameters<typeof registryApi.publishFunction>[0]) =>
      registryApi.publishFunction(data),
    onSuccess: () => {
      toast.success('Function published successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to publish function: ${error.message}`);
    },
  });
}

// Publish function via presigned direct upload. For large source/WASM blobs
// the dashboard uploads the bytes straight to R2 instead of sending them
// through the orchestrator. Falls back to regular publish when the artifact
// store is unavailable or below the size threshold.
export function usePublishRegistryFunctionViaPresigned() {
  return useMutation({
    mutationFn: (
      data: Parameters<typeof registryApi.publishFunctionViaPresigned>[0]
    ) => registryApi.publishFunctionViaPresigned(data),
    onSuccess: () => {
      toast.success('Function published successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to publish function: ${error.message}`);
    },
  });
}

// Get replay
export function useReplay(executionId: string) {
  return useQuery({
    queryKey: registryKeys.replay(executionId),
    queryFn: () => registryApi.getReplay(executionId),
    enabled: !!executionId,
    staleTime: 1000 * 60 * 5,
  });
}

// Get function settings
export function useFunctionSettings(author: string, name: string) {
  return useQuery({
    queryKey: registryKeys.settings(author, name),
    queryFn: () => registryApi.getFunctionSettings(author, name),
    enabled: !!author && !!name,
    staleTime: 1000 * 60,
  });
}

// Update function settings
export function useUpdateFunctionSettings() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      author,
      name,
      data,
    }: {
      author: string;
      name: string;
      data: { customDomains?: string[] };
    }) => registryApi.patchFunctionSettings(author, name, data),
    onSuccess: (_, { author, name }) => {
      queryClient.invalidateQueries({ queryKey: registryKeys.settings(author, name) });
      toast.success('Settings updated successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update settings: ${error.message}`);
    },
  });
}

// Test function
export function useTestRegistryFunction() {
  return useMutation({
    mutationFn: ({
      author,
      name,
      input,
    }: {
      author: string;
      name: string;
      input: unknown;
    }) => registryApi.testFunction(author, name, input),
    onError: (error: Error) => {
      toast.error(`Test failed: ${error.message}`);
    },
  });
}
