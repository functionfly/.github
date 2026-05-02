import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { agentApi, type ChildAgent, type SwarmStats, type AgentMessage } from "@/api/agent";

// Query keys for swarm-related hooks
export const swarmKeys = {
  all: ["swarm"] as const,
  children: (agentId: string) => [...swarmKeys.all, "children", agentId] as const,
  parent: (agentId: string) => [...swarmKeys.all, "parent", agentId] as const,
  health: (agentId: string) => [...swarmKeys.all, "health", agentId] as const,
  inbox: (agentId: string) => [...swarmKeys.all, "inbox", agentId] as const,
  stats: (agentId: string) => [...swarmKeys.all, "stats", agentId] as const,
};

/**
 * Get child agents of an agent
 * GET /v1/agent/{agent_id}/children
 */
export function useAgentChildren(agentId: string) {
  return useQuery({
    queryKey: swarmKeys.children(agentId),
    queryFn: () => agentApi.getChildren(agentId),
    enabled: !!agentId,
    staleTime: 1000 * 30, // 30 seconds
  });
}

/**
 * Get parent agent of an agent
 * GET /v1/agent/{agent_id}/parent
 */
export function useAgentParent(agentId: string) {
  return useQuery({
    queryKey: swarmKeys.parent(agentId),
    queryFn: () => agentApi.getParent(agentId),
    enabled: !!agentId,
    staleTime: 1000 * 60, // 1 minute
  });
}

/**
 * Check swarm health
 * GET /v1/agent/{agent_id}/swarm/health
 */
export function useSwarmHealth(agentId: string, options?: { refetchInterval?: number }) {
  return useQuery({
    queryKey: swarmKeys.health(agentId),
    queryFn: () => agentApi.checkSwarmHealth(agentId, { hours: 1 }),
    enabled: !!agentId,
    refetchInterval: options?.refetchInterval,
    staleTime: 1000 * 10, // 10 seconds
  });
}

/**
 * Get agent inbox messages
 * GET /v1/agent/{agent_id}/inbox
 */
export function useAgentInbox(agentId: string) {
  return useQuery({
    queryKey: swarmKeys.inbox(agentId),
    queryFn: () => agentApi.getInbox(agentId),
    enabled: !!agentId,
    staleTime: 1000 * 15, // 15 seconds
  });
}

/**
 * Get swarm statistics
 * GET /v1/agent/{agent_id}/swarm/stats
 */
export function useSwarmStats(agentId: string) {
  return useQuery({
    queryKey: swarmKeys.stats(agentId),
    queryFn: () => agentApi.getSwarmStats(agentId),
    enabled: !!agentId,
    staleTime: 1000 * 30, // 30 seconds
  });
}

/**
 * Spawn a child agent
 * POST /v1/agent/{agent_id}/spawn
 */
export function useSpawnChildAgent() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: { agentId: string; name: string; swarmRole?: string }) =>
      agentApi.spawnChild(params.agentId, {
        child_name: params.name,
        swarm_role: params.swarmRole,
      }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: swarmKeys.children(variables.agentId) });
      queryClient.invalidateQueries({ queryKey: swarmKeys.stats(variables.agentId) });
      toast.success("Child agent spawned successfully");
    },
    onError: (error: Error) => {
      toast.error(`Failed to spawn child agent: ${error.message}`);
    },
  });
}

/**
 * Send message to an agent
 * POST /v1/agent/{agent_id}/message
 */
export function useSendAgentMessage() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: {
      fromAgentId: string;
      toAgentId: string;
      messageType: string;
      payload?: Record<string, unknown>;
    }) =>
      agentApi.sendMessage(params.fromAgentId, {
        to_agent_id: params.toAgentId,
        message_type: params.messageType,
        payload: params.payload,
      }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: swarmKeys.inbox(variables.toAgentId) });
      toast.success("Message sent successfully");
    },
    onError: (error: Error) => {
      toast.error(`Failed to send message: ${error.message}`);
    },
  });
}
