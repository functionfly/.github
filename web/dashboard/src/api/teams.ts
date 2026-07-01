import { apiClient } from "./client";

export interface TeamMember {
  id: string;
  user_id: string;
  team_id: string;
  role: "owner" | "admin" | "member" | "viewer";
  added_at: string;
  user?: {
    id: string;
    email: string;
    username?: string;
    name?: string;
    avatar?: string;
  };
}

export interface Team {
  id: string;
  tenant_id: string;
  name: string;
  description?: string;
  visibility: "private" | "public";
  default_invite_role: string;
  owner_id: string;
  created_by: string;
  created_at: string;
  members?: TeamMember[];
}

export interface TeamInvite {
  id: string;
  team_id: string;
  email: string;
  role: string;
  token?: string;
  status: "pending" | "accepted" | "expired";
  expires_at: string;
  created_at: string;
}

export interface TeamAuditLog {
  id: string;
  team_id: string;
  actor_id: string;
  actor?: { id: string; email: string; name?: string; username?: string };
  action: string;
  target_type?: string;
  target_id?: string;
  details?: Record<string, unknown>;
  created_at: string;
}

export interface TeamQuota {
  id: string;
  team_id: string;
  resource_type: string;
  max_count: number;
  current_count: number;
}

export interface CreateTeamRequest {
  name: string;
}

export interface UpdateTeamRequest {
  name?: string;
  description?: string;
  visibility?: "private" | "public";
  default_invite_role?: string;
}

export interface InviteMemberRequest {
  email: string;
  role: "admin" | "member" | "viewer";
}

export interface UpdateMemberRequest {
  role: "admin" | "member" | "viewer";
}

export interface TeamListResponse {
  teams: Team[];
}

export interface TeamResponse {
  team: Team;
}

export interface TeamMembersResponse {
  members: TeamMember[];
}

export interface TeamInvitesResponse {
  invites: TeamInvite[];
}

export interface TeamAuditLogsResponse {
  logs: TeamAuditLog[];
}

export interface TeamQuotasResponse {
  quotas: TeamQuota[];
}

export const teamsApi = {
  list: async (): Promise<TeamListResponse> => {
    const response = await apiClient.get<TeamListResponse>("/v1/teams");
    return response ?? { teams: [] };
  },

  get: async (teamId: string): Promise<TeamResponse> => {
    const response = await apiClient.get<TeamResponse>(`/v1/teams/${teamId}`);
    return response;
  },

  create: async (data: CreateTeamRequest): Promise<TeamResponse> => {
    const response = await apiClient.post<TeamResponse>("/v1/teams", data);
    return response;
  },

  update: async (teamId: string, data: UpdateTeamRequest): Promise<TeamResponse> => {
    const response = await apiClient.patch<TeamResponse>(`/v1/teams/${teamId}`, data);
    return response;
  },

  delete: async (teamId: string): Promise<void> => {
    await apiClient.delete(`/v1/teams/${teamId}`);
  },

  listMembers: async (teamId: string): Promise<TeamMembersResponse> => {
    const response = await apiClient.get<TeamMembersResponse>(`/v1/teams/${teamId}/members`);
    return response ?? { members: [] };
  },

  addMember: async (teamId: string, data: InviteMemberRequest): Promise<{ invite: TeamInvite }> => {
    const response = await apiClient.post<{ invite: TeamInvite }>(`/v1/teams/${teamId}/members`, data);
    return response;
  },

  updateMember: async (teamId: string, memberId: string, data: UpdateMemberRequest): Promise<{ member: TeamMember }> => {
    const response = await apiClient.patch<{ member: TeamMember }>(`/v1/teams/${teamId}/members/${memberId}`, data);
    return response;
  },

  removeMember: async (teamId: string, memberId: string): Promise<void> => {
    await apiClient.delete(`/v1/teams/${teamId}/members/${memberId}`);
  },

  leaveTeam: async (teamId: string): Promise<void> => {
    await apiClient.post(`/v1/teams/${teamId}/members/leave`, {});
  },

  transferOwnership: async (teamId: string, newOwnerId: string): Promise<void> => {
    await apiClient.post(`/v1/teams/${teamId}/transfer-ownership`, { new_owner_id: newOwnerId });
  },

  listInvites: async (teamId: string): Promise<TeamInvitesResponse> => {
    const response = await apiClient.get<TeamInvitesResponse>(`/v1/teams/${teamId}/invites`);
    return response ?? { invites: [] };
  },

  resendInvite: async (teamId: string, inviteId: string): Promise<{ invite: TeamInvite }> => {
    const response = await apiClient.post<{ invite: TeamInvite }>(`/v1/teams/${teamId}/invites/${inviteId}/resend`, {});
    return response;
  },

  cancelInvite: async (teamId: string, inviteId: string): Promise<void> => {
    await apiClient.delete(`/v1/teams/${teamId}/invites/${inviteId}`);
  },

  acceptInvite: async (token: string): Promise<{ team: Team }> => {
    const response = await apiClient.post<{ team: Team }>(`/v1/teams/invites/${token}/accept`, {});
    return response;
  },

  getAuditLogs: async (teamId: string, limit = 50, offset = 0): Promise<TeamAuditLogsResponse> => {
    const response = await apiClient.get<TeamAuditLogsResponse>(`/v1/teams/${teamId}/audit-logs?limit=${limit}&offset=${offset}`);
    return response ?? { logs: [] };
  },

  getQuotas: async (teamId: string): Promise<TeamQuotasResponse> => {
    const response = await apiClient.get<TeamQuotasResponse>(`/v1/teams/${teamId}/quotas`);
    return response ?? { quotas: [] };
  },
};
