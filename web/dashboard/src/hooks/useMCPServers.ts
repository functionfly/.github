import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/api/client';
import { toast } from 'sonner';

export interface AgentMCPServer {
  id: string;
  agent_id: string;
  tenant_id: string;
  name: string;
  url: string;
  transport: string;
  description: string;
  enabled: boolean;
  headers: Record<string, string>;
  tool_count: number;
  last_connected_at?: string;
  last_error?: string;
  created_at: string;
  updated_at: string;
}

export const mcpServerKeys = {
  all: ['mcp-servers'] as const,
  list: (agentId: string) => [...mcpServerKeys.all, agentId] as const,
};

export function useMCPServers(agentId: string) {
  return useQuery({
    queryKey: mcpServerKeys.list(agentId),
    queryFn: async () => {
      const res = await apiClient.get<{ servers: AgentMCPServer[] }>(
        `/v1/agent/${agentId}/mcp-servers`
      );
      return res.servers ?? [];
    },
    enabled: !!agentId,
  });
}

export function useAddMCPServer(agentId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: { name: string; url: string; transport: string; description?: string; headers?: Record<string, string> }) => {
      const res = await apiClient.post<{ server: AgentMCPServer }>(
        `/v1/agent/${agentId}/mcp-servers`,
        data
      );
      return res.server;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: mcpServerKeys.list(agentId) });
      toast.success('MCP server added');
    },
    onError: (err: Error) => {
      toast.error(`Failed to add MCP server: ${err.message}`);
    },
  });
}

export function useUpdateMCPServer(agentId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ serverId, data }: { serverId: string; data: Partial<{ name: string; url: string; transport: string; description: string; enabled: boolean; headers: Record<string, string> }> }) => {
      const res = await apiClient.patch<{ server: AgentMCPServer }>(
        `/v1/agent/${agentId}/mcp-servers/${serverId}`,
        data
      );
      return res.server;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: mcpServerKeys.list(agentId) });
    },
    onError: (err: Error) => {
      toast.error(`Failed to update MCP server: ${err.message}`);
    },
  });
}

export function useDeleteMCPServer(agentId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (serverId: string) => {
      await apiClient.delete(`/v1/agent/${agentId}/mcp-servers/${serverId}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: mcpServerKeys.list(agentId) });
      toast.success('MCP server removed');
    },
    onError: (err: Error) => {
      toast.error(`Failed to delete MCP server: ${err.message}`);
    },
  });
}
