import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { agentMemoryApi } from "@/api/agentMemory";
import type {
  AgentMemory,
  CreateAgentMemoryRequest,
  UpdateAgentMemoryRequest,
  AgentMemorySearchRequest,
  RebuildIndexRequest,
  AgentMemoryType,
} from "@/types";

// Query keys
export const agentMemoryKeys = {
  all: ["agent-memories"] as const,
  lists: () => [...agentMemoryKeys.all, "list"] as const,
  list: (filters: {
    agent_id?: string;
    memory_type?: AgentMemoryType;
    limit?: number;
    offset?: number;
  }) => [...agentMemoryKeys.lists(), filters] as const,
  details: () => [...agentMemoryKeys.all, "detail"] as const,
  detail: (id: string) => [...agentMemoryKeys.details(), id] as const,
  search: (filters: {
    agent_id?: string;
    query: string;
    memory_type?: AgentMemoryType;
  }) => [...agentMemoryKeys.all, "search", filters] as const,
  index: () => [...agentMemoryKeys.all, "index"] as const,
};

// List all agent memories with optional filters
export function useAgentMemories(params?: {
  agent_id?: string;
  memory_type?: AgentMemoryType;
  limit?: number;
  offset?: number;
}, enabled: boolean = true) {
  return useQuery({
    queryKey: agentMemoryKeys.list(params),
    queryFn: () => agentMemoryApi.list(params),
    enabled,
  });
}

// Get a single agent memory by ID
export function useAgentMemory(id: string) {
  const isValidId = !!id && id !== "new" && id !== "undefined";

  return useQuery({
    queryKey: agentMemoryKeys.detail(id),
    queryFn: () => agentMemoryApi.get(id),
    enabled: isValidId,
  });
}

// Create a new agent memory
export function useCreateAgentMemory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateAgentMemoryRequest) => agentMemoryApi.create(data),
    onSuccess: (created) => {
      queryClient.invalidateQueries({ queryKey: agentMemoryKeys.lists() });
      toast.success("Memory created successfully");
    },
    onError: (error: Error) => {
      toast.error(`Failed to create memory: ${error.message}`);
    },
  });
}

// Update an existing agent memory
export function useUpdateAgentMemory(id: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateAgentMemoryRequest) => agentMemoryApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: agentMemoryKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: agentMemoryKeys.lists() });
      toast.success("Memory updated successfully");
    },
    onError: (error: Error) => {
      toast.error(`Failed to update memory: ${error.message}`);
    },
  });
}

// Delete an agent memory
export function useDeleteAgentMemory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => agentMemoryApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: agentMemoryKeys.lists() });
      toast.success("Memory deleted successfully");
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete memory: ${error.message}`);
    },
  });
}

// Search agent memories
export function useSearchAgentMemories() {
  return useMutation({
    mutationFn: (data: AgentMemorySearchRequest) => agentMemoryApi.search(data),
    onError: (error: Error) => {
      toast.error(`Failed to search memories: ${error.message}`);
    },
  });
}

// Rebuild the search index
export function useRebuildIndex() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data?: RebuildIndexRequest) => agentMemoryApi.rebuildIndex(data),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: agentMemoryKeys.index() });
      toast.success(result.message);
    },
    onError: (error: Error) => {
      toast.error(`Failed to rebuild index: ${error.message}`);
    },
  });
}
