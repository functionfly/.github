import { apiClient } from '@/api/client';
import { getApiBaseUrl } from '@/lib/constants';
import type { LoginRequest, Session, SignupRequest, SignupResponse, User } from '@/types';
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

// Extend window interface
declare global {
  interface Window {
    hasAuthLogoutListener?: boolean;
  }
}

interface AuthState {
  user: User | null;
  session: Session | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  /** True after the first initialize() has completed (so we know if currentUser is set). */
  authChecked: boolean;
  error: string | null;
  mfaRequired: boolean;

  // Actions
  login: (data: LoginRequest) => Promise<void>;
  signup: (data: SignupRequest) => Promise<SignupResponse>;
  logout: (shouldRedirect?: boolean) => Promise<void>;
  clearError: () => void;
  initialize: () => Promise<void>;
  refreshSession: () => Promise<void>;
  verifyMFA: (code: string) => Promise<void>;
  /** Sync plan from an authoritative source (e.g. GET /users/me) so UI shows correct plan. */
  setUserPlan: (plan: string) => void;
}

/**
 * Safely decode a JWT token payload
 * Handles base64url encoding properly and returns null if decoding fails
 */
function safeDecodeJwtPayload(token: string): { exp?: number; [key: string]: unknown } | null {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) {
      console.error('Invalid JWT format: expected 3 parts');
      return null;
    }

    const base64Url = parts[1];
    if (!base64Url) {
      return null;
    }

    // Convert base64url to base64
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');

    // Add padding if needed
    const padding = 4 - (base64.length % 4);
    const paddedBase64 = padding !== 4 ? base64 + '='.repeat(padding) : base64;

    // Decode base64
    const jsonPayload = atob(paddedBase64);

    return JSON.parse(jsonPayload);
  } catch (error) {
    console.error('Failed to decode JWT token:', error);
    return null;
  }
}

// Create the store
const authStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      session: null,
      isAuthenticated: false,
      isLoading: false,
      authChecked: false,
      error: null,
      mfaRequired: false,

      initialize: async () => {
        const jwtToken = localStorage.getItem('ff-access-token');
        const refreshToken = localStorage.getItem('ff-refresh-token');

        if (!jwtToken) {
          set({ user: null, session: null, isAuthenticated: false, authChecked: true });
          return;
        }

        let cleared = false;
        const clearAuth = () => {
          if (cleared) return;
          cleared = true;
          localStorage.removeItem('ff-access-token');
          localStorage.removeItem('ff-refresh-token');
          localStorage.removeItem('ff-last-wallet-agent-id');
          set({
            user: null,
            session: null,
            isAuthenticated: false,
            error: null,
            authChecked: true,
          });
        };

        // Attempt to validate or refresh the token with retry logic
        const maxRetries = 3;
        let lastError: Error | null = null;

        for (let attempt = 0; attempt < maxRetries; attempt++) {
          try {
            // Check if token is expired locally first with safe decoding
            const payload = safeDecodeJwtPayload(jwtToken);
            if (!payload) {
              clearAuth();
              return;
            }

            const currentTime = Math.floor(Date.now() / 1000);
            const expiresAt = payload.exp || 0;

            let response: Response;
            if (expiresAt > currentTime) {
              // Token is still valid, validate with backend
              const apiUrl = getApiBaseUrl();
              response = await fetch(`${apiUrl}/v1/auth/validate`, {
                method: 'GET',
                headers: {
                  Authorization: `Bearer ${jwtToken}`,
                },
              });

              if (!response.ok) {
                // Token invalid or expired, try to refresh
                if (!refreshToken) {
                  clearAuth();
                  return;
                }
              } else {
                // Token valid, update session
                const userData = await response.json();
                const user: User = {
                  id: userData.user.id,
                  email: userData.user.email,
                  username: userData.user.username,
                  companyName: userData.user.company_name,
                  name: userData.user.name || '',
                  avatar: userData.user.avatar || '',
                  tenantId: userData.user.tenant_id || 'default',
                  plan: userData.user.plan || 'starter',
                  role: userData.user.role,
                  createdAt: userData.user.created_at,
                  updatedAt: userData.user.updated_at,
                  isOnline: true,
                };

                const session: Session = {
                  access_token: jwtToken,
                  refresh_token: refreshToken || '',
                  expires_at: expiresAt,
                  token_type: 'bearer',
                  user: {
                    id: user.id,
                    email: user.email,
                    user_metadata: {
                      name: user.name,
                      avatar_url: user.avatar,
                    },
                    created_at: user.createdAt,
                    updated_at: user.updatedAt,
                  },
                };

                set({
                  user,
                  session,
                  isAuthenticated: true,
                  authChecked: true,
                });
                return;
              }
            }

            // Token is expired or response was not ok, try to refresh
            if (refreshToken) {
              const apiUrl = getApiBaseUrl();
              const refreshResponse = await fetch(`${apiUrl}/v1/auth/refresh`, {
                method: 'POST',
                headers: {
                  'Content-Type': 'application/json',
                },
                body: JSON.stringify({ refresh_token: refreshToken }),
              });

              if (refreshResponse.ok) {
                const refreshData = await refreshResponse.json();

                // Store new tokens
                localStorage.setItem('ff-access-token', refreshData.token);
                localStorage.setItem('ff-refresh-token', refreshData.refresh_token);

                // Create user object from refresh response
                const user: User = {
                  id: refreshData.user.id,
                  email: refreshData.user.email,
                  username: refreshData.user.username,
                  companyName: refreshData.user.company_name,
                  name: refreshData.user.name || '',
                  avatar: refreshData.user.avatar || '',
                  tenantId: refreshData.user.tenant_id || 'default',
                  plan: refreshData.user.plan || 'starter',
                  role: refreshData.user.role,
                  createdAt: refreshData.user.created_at,
                  updatedAt: refreshData.user.updated_at,
                  isOnline: true,
                };

                // Decode new token to get expiration with safe decoding
                const newPayload = safeDecodeJwtPayload(refreshData.token);
                if (!newPayload) {
                  throw new Error('Invalid token format from refresh');
                }
                const newExpiresAt = newPayload.exp || Math.floor(Date.now() / 1000) + 30 * 60; // 30 minutes fallback

                const session: Session = {
                  access_token: refreshData.token,
                  refresh_token: refreshData.refresh_token,
                  expires_at: newExpiresAt,
                  token_type: 'bearer',
                  user: {
                    id: user.id,
                    email: user.email,
                    user_metadata: {
                      name: user.name,
                      avatar_url: user.avatar,
                    },
                    created_at: user.createdAt,
                    updated_at: user.updatedAt,
                  },
                };

                set({
                  user,
                  session,
                  isAuthenticated: true,
                  authChecked: true,
                });
                return;
              }
            }

            // If we get here, refresh failed or wasn't possible
            clearAuth();
            return;
          } catch (error) {
            lastError = error instanceof Error ? error : new Error(String(error));
            console.warn(`Auth initialization attempt ${attempt + 1} failed:`, error);

            // Don't retry on the last attempt
            if (attempt < maxRetries - 1) {
              // Exponential backoff: 500ms, 1000ms, 2000ms
              const delayMs = 500 * Math.pow(2, attempt);
              await new Promise((resolve) => setTimeout(resolve, delayMs));
            }
          }
        }

        // All retries exhausted
        console.error('Auth initialization failed after all retries:', lastError);
        clearAuth();
      },

      login: async (data: LoginRequest) => {
        set({ isLoading: true, error: null });
        try {
          const response = await apiClient.post<{
            user: User;
            session: Session;
          }>('/auth/login', data);

          const { user, session } = response;

          localStorage.setItem('ff-access-token', session.access_token);
          if (session.refresh_token) {
            localStorage.setItem('ff-refresh-token', session.refresh_token);
          }

          set({
            user: { ...user, isOnline: true },
            session,
            isAuthenticated: true,
            isLoading: false,
            authChecked: true,
            mfaRequired: false,
          });
        } catch (error) {
          const message = error instanceof Error ? error.message : 'Login failed';
          set({ error: message, isLoading: false });
          throw error;
        }
      },

      signup: async (data: SignupRequest) => {
        set({ isLoading: true, error: null });
        try {
          const response = await apiClient.post<SignupResponse>('/auth/signup', data);
          set({ isLoading: false });
          return response;
        } catch (error) {
          const message = error instanceof Error ? error.message : 'Signup failed';
          set({ error: message, isLoading: false });
          throw error;
        }
      },

      logout: async (shouldRedirect: boolean = true) => {
        try {
          await apiClient.post('/auth/logout');
        } catch (error) {
          console.warn('Logout API call failed:', error);
        } finally {
          localStorage.removeItem('ff-access-token');
          localStorage.removeItem('ff-refresh-token');
          localStorage.removeItem('ff-last-wallet-agent-id');
          set({
            user: null,
            session: null,
            isAuthenticated: false,
            error: null,
            mfaRequired: false,
            authChecked: true,
          });
          // Redirect to login if not already there
          if (shouldRedirect && window.location.pathname !== '/login') {
            const currentPath = window.location.pathname + window.location.search;
            window.location.href = `/login?redirect=${encodeURIComponent(currentPath)}`;
          }
        }
      },

      clearError: () => set({ error: null }),

      refreshSession: async () => {
        const refreshToken = localStorage.getItem('ff-refresh-token');
        if (!refreshToken) {
          set({ isAuthenticated: false, authChecked: true });
          return;
        }

        try {
          const apiUrl = getApiBaseUrl();
          const response = await fetch(`${apiUrl}/v1/auth/refresh`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({ refresh_token: refreshToken }),
          });

          if (!response.ok) {
            throw new Error('Session refresh failed');
          }

          const data = await response.json();
          localStorage.setItem('ff-access-token', data.token);
          localStorage.setItem('ff-refresh-token', data.refresh_token);

          const user: User = {
            id: data.user.id,
            email: data.user.email,
            username: data.user.username,
            companyName: data.user.company_name,
            name: data.user.name || '',
            avatar: data.user.avatar || '',
            tenantId: data.user.tenant_id || 'default',
            plan: data.user.plan || 'starter',
            role: data.user.role,
            createdAt: data.user.created_at,
            updatedAt: data.user.updated_at,
            isOnline: true,
          };

          const payload = safeDecodeJwtPayload(data.token);
          const expiresAt = payload?.exp || Math.floor(Date.now() / 1000) + 30 * 60;

          const session: Session = {
            access_token: data.token,
            refresh_token: data.refresh_token,
            expires_at: expiresAt,
            token_type: 'bearer',
            user: {
              id: user.id,
              email: user.email,
              user_metadata: {
                name: user.name,
                avatar_url: user.avatar,
              },
              created_at: user.createdAt,
              updated_at: user.updatedAt,
            },
          };

          set({
            user,
            session,
            isAuthenticated: true,
            authChecked: true,
          });
        } catch (error) {
          console.error('Session refresh failed:', error);
          localStorage.removeItem('ff-access-token');
          localStorage.removeItem('ff-refresh-token');
          set({
            user: null,
            session: null,
            isAuthenticated: false,
            authChecked: true,
          });
        }
      },

      verifyMFA: async (code: string) => {
        set({ isLoading: true, error: null });
        try {
          const response = await apiClient.post<{
            user: User;
            session: Session;
          }>('/auth/mfa/verify', { code });

          const { user, session } = response;

          localStorage.setItem('ff-access-token', session.access_token);
          if (session.refresh_token) {
            localStorage.setItem('ff-refresh-token', session.refresh_token);
          }

          set({
            user,
            session,
            isAuthenticated: true,
            isLoading: false,
            mfaRequired: false,
            authChecked: true,
          });
        } catch (error) {
          const message = error instanceof Error ? error.message : 'MFA verification failed';
          set({ error: message, isLoading: false });
          throw error;
        }
      },

      setUserPlan: (plan: string) => {
        set((state) => ({
          user: state.user ? { ...state.user, plan } : null,
        }));
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        user: state.user,
        session: state.session,
        isAuthenticated: state.isAuthenticated,
        authChecked: state.authChecked,
      }),
    }
  )
);

export const useAuthStore = authStore;
