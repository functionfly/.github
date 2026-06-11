/**
 * MCP Center - useMCPFunctions Hook
 * Fetches functions with MCP settings and provides filtering/sorting
 */

import { useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import type { MCPFunction, MCPFunctionFilter, MCPFunctionSort } from '../types';
import { mcpApi } from '@/api/mcp';
import { toast } from 'sonner';

export function useMCPFunctions(
  filter: MCPFunctionFilter = 'all',
  sort: MCPFunctionSort = 'name',
  search: string = ''
) {
  // Fetch all functions with MCP settings
  const { data, isLoading, error } = useQuery({
    queryKey: ['mcp', 'functions'],
    queryFn: () => mcpApi.listFunctions(),
    staleTime: 1000 * 60,
  });

  // Transform API response to MCPFunction format
  const mcpFunctions: MCPFunction[] = useMemo(() => {
    if (!data?.functions) return [];

    return data.functions.map((fn) => ({
      id: fn.id,
      author: fn.author,
      name: fn.name,
      status: fn.status,
      enabled: fn.mcp?.enabled ?? false,
      transports: fn.mcp?.transports ?? ['streamable-http'],
      expose_input_schema: fn.mcp?.expose_input_schema ?? true,
      expose_output_schema: fn.mcp?.expose_output_schema ?? false,
      tool_name_override: fn.mcp?.tool_name_override ?? '',
      rate_limit_per_min: fn.mcp?.rate_limit_per_min ?? 60,
      allowlist_origins: fn.mcp?.allowlist_origins ?? [],
      verified_mcp: fn.mcp?.verified_mcp,
      invocation_count: fn.mcp?.invocation_count,
      last_invoked_at: fn.mcp?.last_invoked_at,
    }));
  }, [data]);

  // Apply filters
  const filteredFunctions = useMemo(() => {
    let result = [...mcpFunctions];

    // Apply search
    if (search) {
      const searchLower = search.toLowerCase();
      result = result.filter(
        (fn) =>
          fn.name.toLowerCase().includes(searchLower) ||
          fn.author.toLowerCase().includes(searchLower)
      );
    }

    // Apply filter
    switch (filter) {
      case 'enabled':
        result = result.filter((fn) => fn.enabled);
        break;
      case 'disabled':
        result = result.filter((fn) => !fn.enabled);
        break;
      case 'verified':
        result = result.filter((fn) => fn.verified_mcp);
        break;
    }

    // Apply sort
    switch (sort) {
      case 'name':
        result.sort((a, b) => `${a.author}/${a.name}`.localeCompare(`${b.author}/${b.name}`));
        break;
      case 'invocations':
        result.sort((a, b) => (b.invocation_count || 0) - (a.invocation_count || 0));
        break;
      case 'lastInvoked':
        result.sort((a, b) => {
          if (!a.last_invoked_at) return 1;
          if (!b.last_invoked_at) return -1;
          return new Date(b.last_invoked_at).getTime() - new Date(a.last_invoked_at).getTime();
        });
        break;
    }

    return result;
  }, [mcpFunctions, filter, sort, search]);

  // Computed summary stats
  const stats = useMemo(() => {
    return {
      total: mcpFunctions.length,
      enabled: mcpFunctions.filter((fn) => fn.enabled).length,
      verified: mcpFunctions.filter((fn) => fn.verified_mcp).length,
      totalInvocations: mcpFunctions.reduce((sum, fn) => sum + (fn.invocation_count || 0), 0),
      transportsCount: new Set(mcpFunctions.flatMap((fn) => fn.transports)).size,
    };
  }, [mcpFunctions]);

  return {
    functions: filteredFunctions,
    allFunctions: mcpFunctions,
    isLoading,
    error,
    stats,
  };
}

// Get MCP settings for a specific function
export function useMCPFunctionSettings(functionId: string) {
  return useQuery({
    queryKey: ['mcp', 'function', functionId, 'settings'],
    queryFn: () => mcpApi.getFunctionSettings(functionId),
    enabled: !!functionId,
    staleTime: 1000 * 60,
  });
}

// Update MCP settings for a function
export function useUpdateMCPFunctionSettings() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      functionId,
      settings,
    }: {
      functionId: string;
      settings: Record<string, unknown>;
    }) => mcpApi.updateFunctionSettings(functionId, settings),
    onSuccess: (_, { functionId }) => {
      queryClient.invalidateQueries({ queryKey: ['mcp', 'function', functionId, 'settings'] });
      queryClient.invalidateQueries({ queryKey: ['mcp', 'functions'] });
      toast.success('MCP settings updated');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update MCP settings: ${error.message}`);
    },
  });
}

// Toggle MCP enabled for a function
export function useToggleMCPEnabled() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ functionId, enabled }: { functionId: string; enabled: boolean }) =>
      mcpApi.updateFunctionSettings(functionId, { enabled }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mcp', 'functions'] });
      toast.success('MCP status updated');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update MCP status: ${error.message}`);
    },
  });
}