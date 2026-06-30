/**
 * MCP Center - useMCPConnections Hook
 * Fetches MCP client connection data with toggle and test mutations
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import type { MCPConnection } from '../types';
import { mcpApi } from '@/api/mcp';
import { toast } from 'sonner';

export function useMCPConnections() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['mcp', 'connections'],
    queryFn: () => mcpApi.getConnections(),
    staleTime: 1000 * 60 * 2, // 2 minutes
  });

  return {
    connections: data?.connections ?? [],
    isLoading,
    error,
  };
}

export function useRefreshMCPConnections() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => mcpApi.getConnections(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mcp', 'connections'] });
    },
    onError: (error: Error) => {
      toast.error(`Failed to refresh connections: ${error.message}`);
    },
  });
}

export function useToggleMCPClient() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ clientType, enabled }: { clientType: string; enabled: boolean }) =>
      mcpApi.toggleConnection(clientType, enabled),
    onSuccess: (_, { clientType, enabled }) => {
      queryClient.invalidateQueries({ queryKey: ['mcp', 'connections'] });
      toast.success(
        enabled
          ? `${clientType} MCP connection enabled`
          : `${clientType} MCP connection disabled`
      );
    },
    onError: (error: Error) => {
      toast.error(`Failed to update connection: ${error.message}`);
    },
  });
}

export function useTestMCPConnection() {
  return useMutation({
    mutationFn: (clientType: string) => mcpApi.testConnection(clientType),
    onSuccess: (data) => {
      if (data.success) {
        toast.success(`${data.message} (${data.latency_ms}ms)`);
      } else {
        toast.error(data.message || 'Connection test failed');
      }
    },
    onError: (error: Error) => {
      toast.error(`Connection test failed: ${error.message}`);
    },
  });
}

export function useBulkToggleMCP(enabled: boolean) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (functionIds: string[]) =>
      Promise.all(
        functionIds.map((id) =>
          mcpApi.updateFunctionSettings(id, { enabled })
        )
      ),
    onSuccess: (_, functionIds) => {
      queryClient.invalidateQueries({ queryKey: ['mcp', 'functions'] });
      toast.success(
        `${functionIds.length} function(s) ${enabled ? 'enabled' : 'disabled'}`
      );
    },
    onError: (error: Error) => {
      toast.error(`Failed to update functions: ${error.message}`);
    },
  });
}
