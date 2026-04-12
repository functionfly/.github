import { apiClient } from '@/api/client';

export interface TeamMemory {
  id: string;
  team_id: string;
  memory_type: 'decision' | 'preference' | 'process' | 'client_context';
  category?: string;
  summary?: string;
  content?: Record<string, any>;
  is_encrypted: boolean;
  confidence_score: number;
  is_validated: boolean;
  validated_by?: string;
  validated_at?: string;
  importance_score: number;
  access_count: number;
  last_accessed_at?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface EncryptedMemoryData {
  ciphertext: string;
  iv: string;
  tag: string;
}

export interface CreateMemoryRequest {
  memory_type: string;
  category?: string;
  summary: string;
  content: Record<string, any>;
  confidence?: number;
  ttl_days?: number;
  is_encrypted?: boolean;
  encrypted_data?: EncryptedMemoryData;
}

export interface SearchResult {
  memories: TeamMemory[];
  total: number;
}

export interface MemoryExtraction {
  id: string;
  team_id: string;
  conversation_id: string;
  memory_type: string;
  category?: string;
  content: Record<string, any>;
  summary: string;
  confidence: number;
  rationale: string;
  status: 'pending' | 'approved' | 'rejected' | 'auto_applied';
  created_at: string;
}

const getTeamMemoriesPath = (teamId: string) => `/v1/teams/${teamId}/memories`;

export const teamMemoryService = {
  // List memories
  async listMemories(
    teamId: string,
    filters?: {
      memory_type?: string;
      category?: string;
      validated?: boolean;
      limit?: number;
      offset?: number;
    }
  ): Promise<{ memories: TeamMemory[]; total: number }> {
    const params = new URLSearchParams();
    if (filters?.memory_type) params.append('memory_type', filters.memory_type);
    if (filters?.category) params.append('category', filters.category);
    if (filters?.validated !== undefined) params.append('validated', String(filters.validated));
    if (filters?.limit) params.append('limit', String(filters.limit));
    if (filters?.offset) params.append('offset', String(filters.offset));

    const query = params.toString() ? `?${params.toString()}` : '';
    return apiClient.get(`${getTeamMemoriesPath(teamId)}${query}`);
  },

  // Get single memory
  async getMemory(teamId: string, memoryId: string): Promise<TeamMemory> {
    return apiClient.get(`${getTeamMemoriesPath(teamId)}/${memoryId}`);
  },

  // Create memory
  async createMemory(teamId: string, data: CreateMemoryRequest): Promise<TeamMemory> {
    return apiClient.post(getTeamMemoriesPath(teamId), data);
  },

  // Update memory
  async updateMemory(
    teamId: string,
    memoryId: string,
    data: Partial<CreateMemoryRequest>
  ): Promise<TeamMemory> {
    return apiClient.patch(`${getTeamMemoriesPath(teamId)}/${memoryId}`, data);
  },

  // Delete memory
  async deleteMemory(teamId: string, memoryId: string): Promise<void> {
    return apiClient.delete(`${getTeamMemoriesPath(teamId)}/${memoryId}`);
  },

  // Search memories
  async searchMemories(
    teamId: string,
    query: string,
    options?: {
      memory_type?: string;
      category?: string;
      limit?: number;
    }
  ): Promise<SearchResult> {
    return apiClient.post(`${getTeamMemoriesPath(teamId)}/search`, {
      query,
      ...options,
    });
  },

  // Natural language query (for agents)
  async queryMemories(
    teamId: string,
    query: string,
    categories?: string[]
  ): Promise<{ context: string; sources: TeamMemory[] }> {
    return apiClient.post(`${getTeamMemoriesPath(teamId)}/query`, {
      query,
      categories,
    });
  },

  // Validate/unvalidate memory
  async validateMemory(
    teamId: string,
    memoryId: string,
    validated: boolean
  ): Promise<void> {
    return apiClient.post(
      `${getTeamMemoriesPath(teamId)}/${memoryId}/validate`,
      { validated }
    );
  },

  // List pending extractions
  async listExtractions(
    teamId: string,
    status: string = 'pending',
    limit: number = 20
  ): Promise<MemoryExtraction[]> {
    const params = new URLSearchParams({ status, limit: String(limit) });
    return apiClient.get(
      `${getTeamMemoriesPath(teamId)}/extractions?${params.toString()}`
    );
  },

  // Approve extraction
  async approveExtraction(teamId: string, extractionId: string): Promise<TeamMemory> {
    return apiClient.post(
      `${getTeamMemoriesPath(teamId)}/extractions/${extractionId}/approve`,
      {}
    );
  },

  // Reject extraction
  async rejectExtraction(
    teamId: string,
    extractionId: string,
    reason?: string
  ): Promise<void> {
    return apiClient.post(
      `${getTeamMemoriesPath(teamId)}/extractions/${extractionId}/reject`,
      { reason }
    );
  },
};

export default teamMemoryService;
