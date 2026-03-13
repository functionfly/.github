import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { followApi, type FollowUserRequest, type FollowFunctionRequest } from "@/api/follow";
import { useApiReachableStore } from "@/stores/apiReachableStore";
import { toast } from "sonner";

const FOLLOW_QUERY_KEYS = {
  userStatus: (username: string) => ["follow", "user", username, "status"] as const,
  userFollowers: (username: string, page: number) => ["follow", "user", username, "followers", page] as const,
  userFollowing: (username: string, page: number) => ["follow", "user", username, "following", page] as const,
  functionStatus: (functionId: string) => ["follow", "function", functionId, "status"] as const,
  functionFollowers: (functionId: string, page: number) => ["follow", "function", functionId, "followers", page] as const,
  myFollowedFunctions: (page: number) => ["follow", "me", "functions", page] as const,
  myStats: () => ["follow", "me", "stats"] as const,
};

export function useUserFollowStatus(username: string) {
  return useQuery({
    queryKey: FOLLOW_QUERY_KEYS.userStatus(username),
    queryFn: () => followApi.getUserFollowStatus(username),
    enabled: !!username,
    staleTime: 30 * 1000, // 30 seconds
  });
}

export function useFollowUser(username: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data?: FollowUserRequest) => followApi.followUser(username, data),
    onSuccess: () => {
      // Invalidate relevant queries
      queryClient.invalidateQueries({ queryKey: FOLLOW_QUERY_KEYS.userStatus(username) });
      queryClient.invalidateQueries({ queryKey: ["follow", "user", username, "followers"] });
      queryClient.invalidateQueries({ queryKey: FOLLOW_QUERY_KEYS.myStats() });
      
      toast.success(`You are now following @${username}`);
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to follow user");
    },
  });
}

export function useUnfollowUser(username: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => followApi.unfollowUser(username),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: FOLLOW_QUERY_KEYS.userStatus(username) });
      queryClient.invalidateQueries({ queryKey: ["follow", "user", username, "followers"] });
      queryClient.invalidateQueries({ queryKey: FOLLOW_QUERY_KEYS.myStats() });
      
      toast.success(`You have unfollowed @${username}`);
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to unfollow user");
    },
  });
}

export function useUserFollowers(username: string, page = 1, pageSize = 20) {
  return useQuery({
    queryKey: FOLLOW_QUERY_KEYS.userFollowers(username, page),
    queryFn: () => followApi.getUserFollowers(username, page, pageSize),
    enabled: !!username,
    staleTime: 60 * 1000, // 1 minute
  });
}

export function useUserFollowing(username: string, page = 1, pageSize = 20) {
  return useQuery({
    queryKey: FOLLOW_QUERY_KEYS.userFollowing(username, page),
    queryFn: () => followApi.getUserFollowing(username, page, pageSize),
    enabled: !!username,
    staleTime: 60 * 1000,
  });
}

export function useFunctionFollowStatus(functionId: string) {
  return useQuery({
    queryKey: FOLLOW_QUERY_KEYS.functionStatus(functionId),
    queryFn: () => followApi.getFunctionFollowStatus(functionId),
    enabled: !!functionId,
    staleTime: 30 * 1000,
  });
}

export function useFollowFunction(functionId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data?: FollowFunctionRequest) => followApi.followFunction(functionId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: FOLLOW_QUERY_KEYS.functionStatus(functionId) });
      queryClient.invalidateQueries({ queryKey: ["follow", "function", functionId, "followers"] });
      queryClient.invalidateQueries({ queryKey: FOLLOW_QUERY_KEYS.myStats() });
      
      toast.success("You are now following this function");
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to follow function");
    },
  });
}

export function useUnfollowFunction(functionId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => followApi.unfollowFunction(functionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: FOLLOW_QUERY_KEYS.functionStatus(functionId) });
      queryClient.invalidateQueries({ queryKey: ["follow", "function", functionId, "followers"] });
      queryClient.invalidateQueries({ queryKey: FOLLOW_QUERY_KEYS.myStats() });
      
      toast.success("You have unfollowed this function");
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to unfollow function");
    },
  });
}

export function useFunctionFollowers(functionId: string, page = 1, pageSize = 20) {
  return useQuery({
    queryKey: FOLLOW_QUERY_KEYS.functionFollowers(functionId, page),
    queryFn: () => followApi.getFunctionFollowers(functionId, page, pageSize),
    enabled: !!functionId,
    staleTime: 60 * 1000,
  });
}

export function useMyFollowedFunctions(page = 1, pageSize = 20) {
  return useQuery({
    queryKey: FOLLOW_QUERY_KEYS.myFollowedFunctions(page),
    queryFn: () => followApi.getMyFollowedFunctions(page, pageSize),
    staleTime: 60 * 1000,
  });
}

const EMPTY_FOLLOW_STATS = { followers: 0, following: 0, functions_followed: 0 };

export function useMyFollowStats() {
  const apiReachable = useApiReachableStore((s) => s.apiReachable);
  return useQuery({
    queryKey: FOLLOW_QUERY_KEYS.myStats(),
    queryFn: async () => {
      try {
        return await followApi.getMyFollowStats();
      } catch (e: unknown) {
        const status = (e as { response?: { status?: number } })?.response?.status;
        if (status === 404) return EMPTY_FOLLOW_STATS;
        throw e;
      }
    },
    enabled: apiReachable === true,
    staleTime: 30 * 1000,
    retry: false,
  });
}
