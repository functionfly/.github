import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { usersApi } from '@/api/users';
import type { PublicUserProfile } from '@/types';
import type { UpdateProfileRequest, MeResponse, AddSkillRequest } from '@/api/users';

// Query keys
export const userKeys = {
  all: ['users'] as const,
  me: () => [...userKeys.all, 'me'] as const,
  profile: (username: string) => [...userKeys.all, 'profile', username] as const,
  sessions: () => [...userKeys.all, 'sessions'] as const,
  loginHistory: () => [...userKeys.all, 'login-history'] as const,
  settings: () => [...userKeys.all, 'settings'] as const,
  analytics: (username: string) => [...userKeys.all, 'analytics', username] as const,
  achievements: (username: string) => [...userKeys.all, 'achievements', username] as const,
  activity: (username: string) => [...userKeys.all, 'activity', username] as const,
  contributions: (username: string) => [...userKeys.all, 'contributions', username] as const,
  skills: (username: string) => [...userKeys.all, 'skills', username] as const,
  usernameEligibility: () => [...userKeys.all, 'username-eligibility'] as const,
  usernameHistory: () => [...userKeys.all, 'username-history'] as const,
};

// Get current user (me)
export function useMe() {
  return useQuery({
    queryKey: userKeys.me(),
    queryFn: () => usersApi.getMe(),
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}

// Get public user profile
export function usePublicProfile(username: string) {
  return useQuery({
    queryKey: userKeys.profile(username),
    queryFn: () => usersApi.getPublicProfile(username),
    enabled: !!username,
    staleTime: 1000 * 60 * 5,
  });
}

// Update current user profile
export function useUpdateMe() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateProfileRequest) => usersApi.updateMe(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: userKeys.me() });
      toast.success('Profile updated successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update profile: ${error.message}`);
    },
  });
}

// Get user sessions
export function useUserSessions() {
  return useQuery({
    queryKey: userKeys.sessions(),
    queryFn: () => usersApi.listSessions(),
    staleTime: 1000 * 60,
  });
}

// Revoke session
export function useRevokeSession() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (sessionId: string) => usersApi.revokeSession(sessionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: userKeys.sessions() });
      toast.success('Session revoked');
    },
    onError: (error: Error) => {
      toast.error(`Failed to revoke session: ${error.message}`);
    },
  });
}

// Revoke other sessions
export function useRevokeOtherSessions() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => usersApi.revokeOtherSessions(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: userKeys.sessions() });
      toast.success('All other sessions revoked');
    },
    onError: (error: Error) => {
      toast.error(`Failed to revoke sessions: ${error.message}`);
    },
  });
}

// Get user login history
export function useLoginHistory(params?: { limit?: number; offset?: number }) {
  return useQuery({
    queryKey: userKeys.loginHistory(),
    queryFn: () => usersApi.listLoginHistory(params),
    staleTime: 1000 * 60 * 2, // 2 minutes
  });
}

// Get user analytics
export function useUserAnalytics(username: string) {
  return useQuery({
    queryKey: userKeys.analytics(username),
    queryFn: () => usersApi.getUserAnalytics(username),
    enabled: !!username,
    staleTime: 1000 * 60 * 5,
  });
}

// Get user achievements
export function useUserAchievements(username: string) {
  return useQuery({
    queryKey: userKeys.achievements(username),
    queryFn: () => usersApi.getUserAchievements(username),
    enabled: !!username,
    staleTime: 1000 * 60 * 10,
  });
}

// Get user activity feed
export function useUserActivity(
  username: string,
  params?: { limit?: number; offset?: number }
) {
  return useQuery({
    queryKey: [...userKeys.activity(username), params],
    queryFn: () => usersApi.getUserActivity(username, params),
    enabled: !!username,
    staleTime: 1000 * 60,
  });
}

// Get user contributions (heatmap)
export function useUserContributions(username: string) {
  return useQuery({
    queryKey: userKeys.contributions(username),
    queryFn: () => usersApi.getUserContributions(username),
    enabled: !!username,
    staleTime: 1000 * 60 * 5,
  });
}

// Get user skills
export function useUserSkills(username: string) {
  return useQuery({
    queryKey: userKeys.skills(username),
    queryFn: () => usersApi.getUserSkills(username),
    enabled: !!username,
    staleTime: 1000 * 60 * 10,
  });
}

// Add skill
export function useAddSkill() {
  const queryClient = useQueryClient();

return useMutation({
    mutationFn: (data: { name: string; level?: 'beginner' | 'intermediate' | 'advanced' | 'expert'; category?: string }) => usersApi.addSkill(data as AddSkillRequest),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: userKeys.all });
      toast.success('Skill added successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to add skill: ${error.message}`);
    },
  });
}

// Remove skill
export function useRemoveSkill() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (skillId: string) => usersApi.removeSkill(skillId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: userKeys.all });
      toast.success('Skill removed successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to remove skill: ${error.message}`);
    },
  });
}

// Username change eligibility
export function useUsernameChangeEligibility() {
  return useQuery({
    queryKey: userKeys.usernameEligibility(),
    queryFn: () => usersApi.getUsernameChangeEligibility(),
    staleTime: 1000 * 60,
  });
}

// Change username
export function useChangeUsername() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { new_username: string; pay_early_fee?: boolean; stripe_payment_id?: string }) =>
      usersApi.changeUsername(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: userKeys.all });
      toast.success('Username changed successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to change username: ${error.message}`);
    },
  });
}

// Get username change history
export function useUsernameChangeHistory() {
  return useQuery({
    queryKey: userKeys.usernameHistory(),
    queryFn: () => usersApi.getUsernameChangeHistory(),
    staleTime: 1000 * 60 * 5,
  });
}

// Change password
export function useChangePassword() {
  return useMutation({
    mutationFn: (data: { currentPassword: string; newPassword: string }) =>
      usersApi.changePassword(data),
    onSuccess: () => {
      toast.success('Password changed successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to change password: ${error.message}`);
    },
  });
}

// Report profile
export function useReportProfile() {
  return useMutation({
    mutationFn: ({ username, reason, details }: { username: string; reason: string; details: string; acknowledged_accuracy: boolean }) =>
      usersApi.reportProfile(username, { reason, details, acknowledged_accuracy: true }),
    onSuccess: () => {
      toast.success('Report submitted successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to submit report: ${error.message}`);
    },
  });
}

// Delete account
export function useDeleteAccount() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => usersApi.deleteMe(),
    onSuccess: () => {
      queryClient.clear();
      toast.success('Account deleted successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete account: ${error.message}`);
    },
  });
}