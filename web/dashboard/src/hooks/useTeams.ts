import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { teamsApi, type Team, type TeamMember, type TeamInvite } from '@/api/teams';

// Query keys
export const teamKeys = {
  all: ['teams'] as const,
  lists: () => [...teamKeys.all, 'list'] as const,
  detail: (id: string) => [...teamKeys.all, 'detail', id] as const,
  members: (teamId: string) => [...teamKeys.all, 'members', teamId] as const,
  invites: (teamId: string) => [...teamKeys.all, 'invites', teamId] as const,
};

// List teams
export function useTeams() {
  return useQuery({
    queryKey: teamKeys.lists(),
    queryFn: () => teamsApi.list(),
    staleTime: 1000 * 60,
  });
}

// Get single team
export function useTeam(teamId: string) {
  return useQuery({
    queryKey: teamKeys.detail(teamId),
    queryFn: () => teamsApi.get(teamId),
    enabled: !!teamId,
    staleTime: 1000 * 60,
  });
}

// Create team
export function useCreateTeam() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (name: string) => teamsApi.create({ name }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: teamKeys.lists() });
      toast.success('Team created successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create team: ${error.message}`);
    },
  });
}

// Update team
export function useUpdateTeam() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ teamId, name }: { teamId: string; name: string }) =>
      teamsApi.update(teamId, { name }),
    onSuccess: (_, { teamId }) => {
      queryClient.invalidateQueries({ queryKey: teamKeys.detail(teamId) });
      queryClient.invalidateQueries({ queryKey: teamKeys.lists() });
      toast.success('Team updated successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update team: ${error.message}`);
    },
  });
}

// Delete team
export function useDeleteTeam() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (teamId: string) => teamsApi.delete(teamId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: teamKeys.lists() });
      toast.success('Team deleted successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete team: ${error.message}`);
    },
  });
}

// List team members
export function useTeamMembers(teamId: string) {
  return useQuery({
    queryKey: teamKeys.members(teamId),
    queryFn: () => teamsApi.listMembers(teamId),
    enabled: !!teamId,
    staleTime: 1000 * 60,
  });
}

// Invite member
export function useInviteMember() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      teamId,
      email,
      role,
    }: {
      teamId: string;
      email: string;
      role: 'admin' | 'member' | 'viewer';
    }) => teamsApi.addMember(teamId, { email, role }),
    onSuccess: (_, { teamId }) => {
      queryClient.invalidateQueries({ queryKey: teamKeys.invites(teamId) });
      toast.success('Invitation sent successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to invite member: ${error.message}`);
    },
  });
}

// Update member role
export function useUpdateMember() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      teamId,
      memberId,
      role,
    }: {
      teamId: string;
      memberId: string;
      role: 'admin' | 'member' | 'viewer';
    }) => teamsApi.updateMember(teamId, memberId, { role }),
    onSuccess: (_, { teamId }) => {
      queryClient.invalidateQueries({ queryKey: teamKeys.members(teamId) });
      toast.success('Member role updated');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update member: ${error.message}`);
    },
  });
}

// Remove member
export function useRemoveMember() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ teamId, memberId }: { teamId: string; memberId: string }) =>
      teamsApi.removeMember(teamId, memberId),
    onSuccess: (_, { teamId }) => {
      queryClient.invalidateQueries({ queryKey: teamKeys.members(teamId) });
      toast.success('Member removed successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to remove member: ${error.message}`);
    },
  });
}

// List invites
export function useTeamInvites(teamId: string) {
  return useQuery({
    queryKey: teamKeys.invites(teamId),
    queryFn: () => teamsApi.listInvites(teamId),
    enabled: !!teamId,
    staleTime: 1000 * 60,
  });
}

// Resend invite
export function useResendInvite() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ teamId, inviteId }: { teamId: string; inviteId: string }) =>
      teamsApi.resendInvite(teamId, inviteId),
    onSuccess: (_, { teamId }) => {
      queryClient.invalidateQueries({ queryKey: teamKeys.invites(teamId) });
      toast.success('Invitation resent');
    },
    onError: (error: Error) => {
      toast.error(`Failed to resend invite: ${error.message}`);
    },
  });
}

// Cancel invite
export function useCancelInvite() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ teamId, inviteId }: { teamId: string; inviteId: string }) =>
      teamsApi.cancelInvite(teamId, inviteId),
    onSuccess: (_, { teamId }) => {
      queryClient.invalidateQueries({ queryKey: teamKeys.invites(teamId) });
      toast.success('Invitation cancelled');
    },
    onError: (error: Error) => {
      toast.error(`Failed to cancel invite: ${error.message}`);
    },
  });
}

// Accept invite
export function useAcceptInvite() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (token: string) => teamsApi.acceptInvite(token),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: teamKeys.lists() });
      toast.success('Invitation accepted successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to accept invite: ${error.message}`);
    },
  });
}
