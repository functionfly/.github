import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { agentApi, type AgentIdentity, type BehavioralPolicy, type AgentQuota } from "@/api/agent";

// Query keys
export const agentKeys = {
  all: ["agents"] as const,
  lists: () => [...agentKeys.all, "list"] as const,
  list: (filters: { limit?: number; offset?: number }) => [...agentKeys.lists(), filters] as const,
  details: () => [...agentKeys.all, "detail"] as const,
  detail: (id: string) => [...agentKeys.details(), id] as const,
  usage: (id: string) => [...agentKeys.detail(id), "usage"] as const,
  policy: (id: string) => [...agentKeys.detail(id), "policy"] as const,
  quota: (id: string) => [...agentKeys.detail(id), "quota"] as const,
  executions: (id: string) => [...agentKeys.detail(id), "executions"] as const,
  analytics: (id: string) => [...agentKeys.detail(id), "analytics"] as const,
};

// List all agents
export function useAgents(params?: { limit?: number; offset?: number }) {
  return useQuery({
    queryKey: agentKeys.list(params ?? {}),
    queryFn: () => agentApi.listAgents(params),
  });
}

// Get a single agent by ID
export function useAgent(agentId: string) {
  return useQuery({
    queryKey: agentKeys.detail(agentId),
    queryFn: () => agentApi.getAgent(agentId),
    enabled: !!agentId,
  });
}

// Get agent usage
export function useAgentUsage(agentId: string) {
  return useQuery({
    queryKey: agentKeys.usage(agentId),
    queryFn: () => agentApi.getUsage(agentId),
    enabled: !!agentId,
  });
}

// Get agent policy
export function useAgentPolicy(agentId: string) {
  return useQuery({
    queryKey: agentKeys.policy(agentId),
    queryFn: () => agentApi.getPolicy(agentId),
    enabled: !!agentId,
  });
}

// Get agent quota
export function useAgentQuota(agentId: string) {
  return useQuery({
    queryKey: agentKeys.quota(agentId),
    queryFn: () => agentApi.getQuota(agentId),
    enabled: !!agentId,
  });
}

// Register a new agent
export function useRegisterAgent() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { agentId: string; name: string; description?: string }) =>
      agentApi.registerAgent(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: agentKeys.lists() });
      toast.success("Agent registered successfully");
    },
    onError: (error: Error) => {
      toast.error(`Failed to register agent: ${error.message}`);
    },
  });
}

// Update agent (general update via direct agent endpoint)
export function useUpdateAgent(agentId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: {
      name?: string;
      description?: string;
      capabilities?: Record<string, unknown>;
      autonomous_enabled?: boolean;
      evolution_enabled?: boolean;
      model?: string;
      thinking_mode?: string;
      thinking_budget?: number;
    }) => agentApi.updateAgent(agentId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: agentKeys.detail(agentId) });
      toast.success("Agent updated successfully");
    },
    onError: (error: Error) => {
      toast.error(`Failed to update agent: ${error.message}`);
    },
  });
}

// Update agent policy
export function useUpdateAgentPolicy(agentId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (policy: Partial<BehavioralPolicy>) => agentApi.updatePolicy(agentId, policy),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: agentKeys.policy(agentId) });
      toast.success("Policy updated successfully");
    },
    onError: (error: Error) => {
      toast.error(`Failed to update policy: ${error.message}`);
    },
  });
}

// Delete an agent
export function useDeleteAgent() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (agentId: string) => agentApi.deleteAgent(agentId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: agentKeys.lists() });
      toast.success("Agent deleted successfully");
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete agent: ${error.message}`);
    },
  });
}
