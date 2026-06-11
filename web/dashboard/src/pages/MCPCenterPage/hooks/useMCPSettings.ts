/**
 * MCP Center - useMCPSettings Hook
 * Global MCP settings management
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import type { MCPSettingsGlobal } from '../types';
import { mcpApi } from '@/api/mcp';
import { toast } from 'sonner';

export function useMCPSettings() {
  const queryClient = useQueryClient();

  const { data, isLoading, error } = useQuery({
    queryKey: ['mcp', 'settings'],
    queryFn: () => mcpApi.getSettings(),
    staleTime: 1000 * 60 * 10, // 10 minutes
  });

  const mutation = useMutation({
    mutationFn: (settings: Partial<MCPSettingsGlobal>) => mcpApi.updateSettings(settings),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mcp', 'settings'] });
      toast.success('MCP settings updated');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update MCP settings: ${error.message}`);
    },
  });

  return {
    settings: data,
    isLoading,
    error,
    updateSettings: mutation.mutate,
    isUpdating: mutation.isPending,
  };
}