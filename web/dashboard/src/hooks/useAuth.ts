/**
 * React hooks for FunctionFly authentication
 */

import { useState, useEffect, useCallback } from 'react';
import { auth, FunctionFlyAuth } from '../lib/auth';

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