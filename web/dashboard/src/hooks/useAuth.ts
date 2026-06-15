/**
 * React hooks for FunctionFly authentication
 * 
 * Note: FunctionFlyAuth from '../lib/auth' is deprecated.
 * Use apiClient from '@/api/client' for all API requests.
 */

import { useEffect, useCallback } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { authApi } from '@/api/auth';
import { apiClient } from '@/api/client';
import { useAuthStore } from '@/stores/authStore';
import type { LoginRequest, SignupRequest } from '@/types';

/**
 * @deprecated Use authStore directly instead. This hook is kept for
 * backward compatibility.
 */
export function useAuth() {
  const user = useAuthStore((state) => state.user);
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const isLoading = useAuthStore((state) => state.isLoading);
  const error = useAuthStore((state) => state.error);
  const initialize = useAuthStore((state) => state.initialize);
  const login = useAuthStore((state) => state.login);
  const logout = useAuthStore((state) => state.logout);

  useEffect(() => {
    initialize();
  }, [initialize]);

  return {
    user,
    isAuthenticated,
    isLoading,
    error,
    login,
    logout,
    checkAuthStatus: initialize,
  };
}

/**
 * Hook for making authenticated API requests with automatic token refresh
 * Uses apiClient (Axios-based) for consistent behavior with other API calls.
 * Token refresh is handled automatically by apiClient interceptors.
 * 
 * @deprecated Use direct apiClient calls instead. This hook is kept for
 * backward compatibility during migration.
 */
export function useAuthenticatedRequest() {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const makeRequest = useCallback(async (
    method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE',
    path: string,
    body?: unknown
  ) => {
    try {
      setIsLoading(true);
      setError(null);

      let response: unknown;
      switch (method) {
        case 'GET':
          response = await apiClient.get(path);
          break;
        case 'POST':
          response = await apiClient.post(path, body);
          break;
        case 'PUT':
          response = await apiClient.put(path, body);
          break;
        case 'PATCH':
          response = await apiClient.patch(path, body);
          break;
        case 'DELETE':
          response = await apiClient.delete(path);
          break;
        default:
          throw new Error(`Unsupported method: ${method}`);
      }

      return response;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Request failed';
      setError(errorMessage);
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, []);

  return {
    makeRequest,
    isLoading,
    error,
  };
}

/**
 * Hook for admin operations that automatically handles CSRF and HMAC
 * Uses apiClient for consistent request handling.
 * 
 * @deprecated Admin operations should use specific authApi methods or
 * direct apiClient calls. This hook is kept for backward compatibility.
 */
export function useAdminAPI() {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const adminRequest = useCallback(async (
    method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE',
    path: string,
    body?: unknown
  ) => {
    try {
      setIsLoading(true);
      setError(null);
      
      // Ensure path starts with /v1/admin/
      const fullPath = path.startsWith('/v1/admin/') ? path : `/v1/admin/${path}`;

      let response: unknown;
      switch (method) {
        case 'GET':
          response = await apiClient.get(fullPath);
          break;
        case 'POST':
          response = await apiClient.post(fullPath, body);
          break;
        case 'PUT':
          response = await apiClient.put(fullPath, body);
          break;
        case 'PATCH':
          response = await apiClient.patch(fullPath, body);
          break;
        case 'DELETE':
          response = await apiClient.delete(fullPath);
          break;
        default:
          throw new Error(`Unsupported method: ${method}`);
      }

      return response;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Request failed';
      setError(errorMessage);
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, []);

  return {
    // Convenience methods for common admin operations
    getSignupInvites: () => adminRequest('GET', '/signup-invites'),
    createSignupInvite: (data: { label: string; max_uses?: number }) =>
      adminRequest('POST', '/signup-invites', data),
    revokeSignupInvite: (id: string) =>
      adminRequest('POST', `/signup-invites/${id}/revoke`),

    // Generic admin request method
    adminRequest,
    isLoading,
    error,
  };
}

// ============================================================================
// TanStack Query-based Auth Hooks
// ============================================================================

// Query keys
export const authKeys = {
  all: ['auth'] as const,
  user: () => [...authKeys.all, 'user'] as const,
  mfaStatus: () => [...authKeys.all, 'mfa-status'] as const,
  oauthProviders: () => [...authKeys.all, 'oauth-providers'] as const,
};

/** Hook for user login with TanStack Query */
export function useSignIn() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: LoginRequest) => authApi.login(data),
    onSuccess: (data) => {
      // Store token if returned
      if (data.token) {
        // Token is handled by apiClient
      }
      queryClient.invalidateQueries({ queryKey: authKeys.user() });
      toast.success('Signed in successfully');
    },
    onError: (error: Error) => {
      toast.error(`Sign in failed: ${error.message}`);
    },
  });
}

/** Hook for user signup with TanStack Query */
export function useSignUp() {
  return useMutation({
    mutationFn: (data: SignupRequest) => authApi.signup(data),
    onSuccess: () => {
      toast.success('Account created successfully. Please verify your email.');
    },
    onError: (error: Error) => {
      toast.error(`Sign up failed: ${error.message}`);
    },
  });
}

/** Hook for user logout with TanStack Query */
export function useSignOut() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      authApi.logout();
      return { success: true };
    },
    onSuccess: () => {
      queryClient.clear();
      toast.success('Signed out successfully');
    },
    onError: (error: Error) => {
      toast.error(`Sign out failed: ${error.message}`);
    },
  });
}

/** Hook for magic link login request */
export function useMagicLink() {
  return useMutation({
    mutationFn: (email: string) =>
      // Uses usersApi but accessed through the auth flow
      import('@/api/users').then((m) => m.usersApi.requestPasswordReset(email)),
    onSuccess: () => {
      toast.success('Magic link sent to your email');
    },
    onError: (error: Error) => {
      toast.error(`Failed to send magic link: ${error.message}`);
    },
  });
}

/** Hook for verifying magic link token */
export function useVerifyMagicLink() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ token, newPassword }: { token: string; newPassword?: string }) =>
      // This is actually password reset confirmation
      import('@/api/users').then((m) =>
        m.usersApi.confirmPasswordReset(token, newPassword || '')
      ),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: authKeys.user() });
      toast.success('Magic link verified successfully');
    },
    onError: (error: Error) => {
      toast.error(`Magic link verification failed: ${error.message}`);
    },
  });
}

/** Hook for fetching OAuth providers */
export function useOAuthProviders() {
  return useQuery({
    queryKey: authKeys.oauthProviders(),
    queryFn: () =>
      fetch(`${import.meta.env.VITE_API_URL || ''}/v1/auth/oauth/providers`).then((r) => r.json()),
    staleTime: 1000 * 60 * 5,
  });
}

/** Hook for getting OAuth URL */
export function useOAuthUrl() {
  return useMutation({
    mutationFn: (provider: string) =>
      fetch(
        `${import.meta.env.VITE_API_URL || ''}/v1/auth/oauth/url?provider=${encodeURIComponent(provider)}`
      ).then((r) => r.json() as Promise<{ url: string }>),
  });
}

/** Hook for password reset request */
export function usePasswordResetRequest() {
  return useMutation({
    mutationFn: (email: string) =>
      import('@/api/users').then((m) => m.usersApi.requestPasswordReset(email)),
    onSuccess: () => {
      toast.success('Password reset link sent to your email');
    },
    onError: (error: Error) => {
      toast.error(`Failed to send reset link: ${error.message}`);
    },
  });
}

/** Hook for password reset confirmation */
export function usePasswordResetConfirm() {
  return useMutation({
    mutationFn: ({ token, newPassword }: { token: string; newPassword: string }) =>
      import('@/api/users').then((m) => m.usersApi.confirmPasswordReset(token, newPassword)),
    onSuccess: () => {
      toast.success('Password reset successfully');
    },
    onError: (error: Error) => {
      toast.error(`Password reset failed: ${error.message}`);
    },
  });
}

/** Hook for checking username availability */
export function useCheckUsernameAvailability(username: string) {
  return useQuery({
    queryKey: [...authKeys.all, 'username-check', username],
    queryFn: () => authApi.checkUsernameAvailability(username),
    enabled: !!username && username.length >= 3,
    staleTime: 1000 * 60,
  });
}

/** Hook for resending verification email */
export function useResendVerification() {
  return useMutation({
    mutationFn: (email: string) => authApi.resendVerification(email),
    onSuccess: () => {
      toast.success('Verification email sent');
    },
    onError: (error: Error) => {
      toast.error(`Failed to resend verification: ${error.message}`);
    },
  });
}

/** Hook for MFA status */
export function useMFAStatus() {
  return useQuery({
    queryKey: authKeys.mfaStatus(),
    queryFn: () => authApi.getMFAStatus(),
    staleTime: 1000 * 60,
  });
}

/** Hook for setting up MFA */
export function useSetupMFA() {
  return useMutation({
    mutationFn: () => authApi.setupMFA(),
    onSuccess: (data) => {
      toast.success('MFA setup initiated');
      return data;
    },
    onError: (error: Error) => {
      toast.error(`MFA setup failed: ${error.message}`);
    },
  });
}

/** Hook for verifying MFA setup code */
export function useVerifyMFASetup() {
  return useMutation({
    mutationFn: (code: string) => authApi.verifyMFASetupCode(code),
    onSuccess: () => {
      toast.success('MFA verified successfully');
    },
    onError: (error: Error) => {
      toast.error(`MFA verification failed: ${error.message}`);
    },
  });
}

/** Hook for enabling MFA */
export function useEnableMFA() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => authApi.enableMFA(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: authKeys.mfaStatus() });
      toast.success('MFA enabled successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to enable MFA: ${error.message}`);
    },
  });
}

/** Hook for disabling MFA */
export function useDisableMFA() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ password, code }: { password: string; code: string }) =>
      authApi.disableMFA(password, code),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: authKeys.mfaStatus() });
      toast.success('MFA disabled successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to disable MFA: ${error.message}`);
    },
  });
}