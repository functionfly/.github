import { apiClient } from "./client";

export interface TeamMember {
  id: string;
  user_id: string;
  team_id: string;
  role: "owner" | "admin" | "member" | "viewer";
  created_at: string;
  user?: {
    id: string;
    email: string;
    username?: string;
    name?: string;
  };
}

export interface Team {
  id: string;
  tenant_id: string;
  name: string;
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

export interface CreateTeamRequest {
  name: string;
}

export interface UpdateTeamRequest {
  name?: string;
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

export const teamsApi = {
  // List all teams for the current tenant
  list: async (): Promise<TeamListResponse> => {
    const response = await apiClient.get<TeamListResponse>("/v1/teams");
    return response ?? { teams: [] };
  },

  // Get a specific team
  get: async (teamId: string): Promise<TeamResponse> => {
    const response = await apiClient.get<TeamResponse>(`/v1/teams/${teamId}`);
    return response;
  },

  // Create a new team
  create: async (data: CreateTeamRequest): Promise<TeamResponse> => {
    const response = await apiClient.post<TeamResponse>("/v1/teams", data);
    return response;
  },

  // Update a team
  update: async (teamId: string, data: UpdateTeamRequest): Promise<TeamResponse> => {
    const response = await apiClient.patch<TeamResponse>(`/v1/teams/${teamId}`, data);
    return response;
  },

  // Delete a team
  delete: async (teamId: string): Promise<void> => {
    await apiClient.delete(`/v1/teams/${teamId}`);
  },

  // List team members
  listMembers: async (teamId: string): Promise<TeamMembersResponse> => {
    const response = await apiClient.get<TeamMembersResponse>(`/v1/teams/${teamId}/members`);
    return response ?? { members: [] };
  },

  // Add a member to a team
  addMember: async (teamId: string, data: InviteMemberRequest): Promise<{ invite: TeamInvite }> => {
    const response = await apiClient.post<{ invite: TeamInvite }>(`/v1/teams/${teamId}/members`, data);
    return response;
  },

  // Update a team member's role
  updateMember: async (teamId: string, memberId: string, data: UpdateMemberRequest): Promise<{ member: TeamMember }> => {
    const response = await apiClient.patch<{ member: TeamMember }>(`/v1/teams/${teamId}/members/${memberId}`, data);
    return response;
  },

  // Remove a member from a team
  removeMember: async (teamId: string, memberId: string): Promise<void> => {
    await apiClient.delete(`/v1/teams/${teamId}/members/${memberId}`);
  },

  // List pending invites
  listInvites: async (teamId: string): Promise<TeamInvitesResponse> => {
    const response = await apiClient.get<TeamInvitesResponse>(`/v1/teams/${teamId}/invites`);
    return response ?? { invites: [] };
  },

  // Resend invite
  resendInvite: async (teamId: string, inviteId: string): Promise<{ invite: TeamInvite }> => {
    const response = await apiClient.post<{ invite: TeamInvite }>(`/v1/teams/${teamId}/invites/${inviteId}/resend`, {});
    return response;
  },

  // Cancel invite
  cancelInvite: async (teamId: string, inviteId: string): Promise<void> => {
    await apiClient.delete(`/v1/teams/${teamId}/invites/${inviteId}`);
  },

  // Accept invite (for receiving users)
  acceptInvite: async (token: string): Promise<{ team: Team }> => {
    const response = await apiClient.post<{ team: Team }>(`/v1/teams/invites/${token}/accept`, {});
    return response;
  },
};
