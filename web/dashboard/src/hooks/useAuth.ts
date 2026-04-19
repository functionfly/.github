/**
 * React hooks for FunctionFly authentication
 */

import { useState, useEffect, useCallback } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { auth, FunctionFlyAuth } from '../lib/auth';
import { authApi } from '@/api/auth';
import type { LoginRequest, SignupRequest } from '@/types';

interface User {
  id: string;
  email: string;
  username: string;
  role: string;
  plan: string;
  email_verified: boolean;
}

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
}

export function useAuth() {
  const [state, setState] = useState<AuthState>({
    user: null,
    isAuthenticated: false,
    isLoading: true,
    error: null,
  });

  // Initialize auth state on mount
  useEffect(() => {
    checkAuthStatus();
  }, []);

  const checkAuthStatus = useCallback(async () => {
    try {
      setState(prev => ({ ...prev, isLoading: true, error: null }));

      if (!auth.isAuthenticated()) {
        setState({
          user: null,
          isAuthenticated: false,
          isLoading: false,
          error: null,
        });
        return;
      }

      // Try to get user profile to validate token
      const response = await auth.get('/users/me');
      if (response.ok) {
        const userData = await response.json();
        setState({
          user: userData,
          isAuthenticated: true,
          isLoading: false,
          error: null,
        });
      } else if (response.status === 401) {
        // Token expired, try refresh
        try {
          await auth.refreshAuthToken();
          const retryResponse = await auth.get('/users/me');
          if (retryResponse.ok) {
            const userData = await retryResponse.json();
            setState({
              user: userData,
              isAuthenticated: true,
              isLoading: false,
              error: null,
            });
          } else {
            throw new Error('Failed to refresh token');
          }
        } catch (refreshError) {
          await auth.logout();
          setState({
            user: null,
            isAuthenticated: false,
            isLoading: false,
            error: 'Session expired',
          });
        }
      } else {
        throw new Error(`Failed to get user profile: ${response.statusText}`);
      }
    } catch (error) {
      setState({
        user: null,
        isAuthenticated: false,
        isLoading: false,
        error: error instanceof Error ? error.message : 'Authentication check failed',
      });
    }
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    try {
      setState(prev => ({ ...prev, isLoading: true, error: null }));

      await auth.login(email, password);

      // Get user profile after login
      const response = await auth.get('/users/me');
      if (response.ok) {
        const userData = await response.json();
        setState({
          user: userData,
          isAuthenticated: true,
          isLoading: false,
          error: null,
        });
      } else {
        throw new Error('Failed to get user profile after login');
      }
    } catch (error) {
      setState(prev => ({
        ...prev,
        isLoading: false,
        error: error instanceof Error ? error.message : 'Login failed',
      }));
      throw error;
    }
  }, []);

  const logout = useCallback(async () => {
    try {
      await auth.logout();
      setState({
        user: null,
        isAuthenticated: false,
        isLoading: false,
        error: null,
      });
    } catch (error) {
      // Even if logout fails, clear local state
      setState({
        user: null,
        isAuthenticated: false,
        isLoading: false,
        error: null,
      });
    }
  }, []);

  return {
    user: state.user,
    isAuthenticated: state.isAuthenticated,
    isLoading: state.isLoading,
    error: state.error,
    login,
    logout,
    checkAuthStatus,
  };
}

/**
 * Hook for making authenticated API requests with automatic token refresh
 */
export function useAuthenticatedRequest() {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const makeRequest = useCallback(async (
    method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE',
    path: string,
    body?: any,
    headers?: Record<string, string>
  ) => {
    try {
      setIsLoading(true);
      setError(null);

      const request = { method, path, body, headers };
      const response = await auth.request(request);

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Request failed: ${response.status} ${response.statusText} - ${errorText}`);
      }

      // Try to parse JSON response
      const contentType = response.headers.get('content-type');
      if (contentType && contentType.includes('application/json')) {
        return await response.json();
      } else {
        return await response.text();
      }
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
 */
export function useAdminAPI() {
  const { makeRequest, isLoading, error } = useAuthenticatedRequest();

  const adminRequest = useCallback(async (
    method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE',
    path: string,
    body?: any
  ) => {
    // Ensure path starts with /v1/admin/
    const fullPath = path.startsWith('/v1/admin/') ? path : `/v1/admin/${path}`;

    return makeRequest(method, fullPath, body);
  }, [makeRequest]);

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