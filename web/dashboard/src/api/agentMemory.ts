import { apiClient } from "./client";
import type {
  AgentMemory,
  CreateAgentMemoryRequest,
  UpdateAgentMemoryRequest,
  AgentMemorySearchRequest,
  AgentMemorySearchResponse,
  RebuildIndexRequest,
  RebuildIndexResponse,
  ListAgentMemoriesResponse,
} from "@/types";

export const agentMemoryApi = {
  // List memories with optional filters
  list: async (params?: {
    agent_id?: string;
    memory_type?: string;
    limit?: number;
    offset?: number;
  }): Promise<ListAgentMemoriesResponse> => {
    const queryParams = new URLSearchParams();
    if (params?.agent_id) queryParams.append("agent_id", params.agent_id);
    if (params?.memory_type) queryParams.append("memory_type", params.memory_type);
    if (params?.limit) queryParams.append("limit", params.limit.toString());
    if (params?.offset) queryParams.append("offset", params.offset.toString());

    return apiClient.get<ListAgentMemoriesResponse>(
      `/v1/agent-memories?${queryParams.toString()}`
    );
  },

  // Get a single memory by ID
  get: async (id: string): Promise<AgentMemory> => {
    return apiClient.get<AgentMemory>(`/v1/agent-memories/${id}`);
  },

  // Create a new memory
  create: async (data: CreateAgentMemoryRequest): Promise<AgentMemory> => {
    return apiClient.post<AgentMemory>("/v1/agent-memories", data);
  },

  // Update an existing memory
  update: async (id: string, data: UpdateAgentMemoryRequest): Promise<AgentMemory> => {
    return apiClient.patch<AgentMemory>(`/v1/agent-memories/${id}`, data);
  },

  // Delete a memory
  delete: async (id: string): Promise<void> => {
    await apiClient.delete<void>(`/v1/agent-memories/${id}`);
  },

  // Search memories using vector similarity or filters
  search: async (data: AgentMemorySearchRequest): Promise<AgentMemorySearchResponse> => {
    return apiClient.post<AgentMemorySearchResponse>("/v1/agent-memories/search", data);
  },

  // Rebuild the search index
  rebuildIndex: async (data?: RebuildIndexRequest): Promise<RebuildIndexResponse> => {
    return apiClient.post<RebuildIndexResponse>("/v1/agent-memories/index", data || {});
  },
};
