import { apiClient } from '@/api/client';
import { getApiBaseUrl } from '@/lib/constants';
import { tokenVault } from '@/utils/token-vault';
import type { LoginRequest, Session, SignupRequest, SignupResponse, User } from '@/types';
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

// Extend window interface
declare global {
  interface Window {
    hasAuthSyncListener?: boolean;
  }
}

const AUTH_CHANNEL_NAME = 'ff-auth-sync';
let authChannel: BroadcastChannel | null = null;

function getAuthChannel(): BroadcastChannel | null {
  if (typeof BroadcastChannel === 'undefined') return null;
  if (!authChannel) {
    authChannel = new BroadcastChannel(AUTH_CHANNEL_NAME);
  }
  return authChannel;
}

type AuthSyncEvent = {
  type: 'login' | 'logout' | 'token_updated';
  timestamp: number;
  userId?: string;
};

function broadcastAuthEvent(event: AuthSyncEvent) {
  const channel = getAuthChannel();
  if (channel) {
    channel.postMessage(event);
  }
  localStorage.setItem('ff-auth-event', JSON.stringify(event));
}

function setupAuthSyncListener(store: { getState: () => { logout: (redirect?: boolean) => void; initialize: () => void } }) {
  if (typeof window === 'undefined' || window.hasAuthSyncListener) return;
  window.hasAuthSyncListener = true;

  const handleAuthEvent = async (event: AuthSyncEvent) => {
    if (event.type === 'logout') {
      // Check if we already processed this logout event (prevent double-processing)
      const lastLogoutTime = localStorage.getItem('ff-last-logout-time');
      const now = Date.now();
      if (lastLogoutTime && now - parseInt(lastLogoutTime, 10) < 1000) {
        // Processed a logout within the last second, skip duplicate
        return;
      }
      localStorage.setItem('ff-last-logout-time', now.toString());

      // Check if token is still valid (within 60 seconds of expiry)
      const token = await tokenVault.getAccessToken();
      if (token) {
        const payload = safeDecodeJwtPayload(token);
        const currentTime = Math.floor(Date.now() / 1000);
        if (payload?.exp && payload.exp > currentTime - 60) {
          return;
        }
      }
      store.getState().logout(false);
      return;
    }
    if (event.type === 'login' || event.type === 'token_updated') {
      store.getState().initialize();
    }
  };

  const broadcastChannel = getAuthChannel();
  if (broadcastChannel) {
    broadcastChannel.onmessage = (e: MessageEvent<AuthSyncEvent>) => {
      handleAuthEvent(e.data);
    };
  }

  window.addEventListener('storage', (e: StorageEvent) => {
    if (e.key === 'ff-auth-event' && e.newValue) {
      try {
        const event: AuthSyncEvent = JSON.parse(e.newValue);
        handleAuthEvent(event);
      } catch {
        // Ignore parse errors
      }
    }
  });
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
        setupAuthSyncListener(authStore);

        // Initialize TokenVault and get tokens from secure storage
        await tokenVault.initialize();
        const jwtToken = await tokenVault.getAccessToken();
        const refreshToken = await tokenVault.getRefreshToken();

        if (!jwtToken) {
          set({ user: null, session: null, isAuthenticated: false, authChecked: true });
          return;
        }

        let cleared = false;
        const clearAuth = async () => {
          if (cleared) return;
          cleared = true;
          await tokenVault.clearTokens();
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
            const payload = safeDecodeJwtPayload(jwtToken);
            if (!payload) {
              if (!refreshToken) {
                await clearAuth();
                return;
              }
            }

            const currentTime = Math.floor(Date.now() / 1000);
            const expiresAt = payload?.exp ? payload.exp : 0;

            let response: Response;
            if (expiresAt === 0 || expiresAt > currentTime) {
              const apiUrl = getApiBaseUrl();
              response = await fetch(`${apiUrl}/v1/auth/validate`, {
                method: 'GET',
                headers: {
                  Authorization: `Bearer ${jwtToken}`,
                },
              });

              if (!response.ok) {
                if (!refreshToken) {
                  await clearAuth();
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
                broadcastAuthEvent({ type: 'token_updated', timestamp: Date.now(), userId: user.id });
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

                // Store new tokens in encrypted storage
                await tokenVault.setAccessToken(refreshData.token);
                await tokenVault.setRefreshToken(refreshData.refresh_token);

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
                broadcastAuthEvent({ type: 'token_updated', timestamp: Date.now(), userId: user.id });
                return;
              }

              // Refresh failed with 4xx (not 429) - token is invalid, don't retry
              if (refreshResponse.status >= 400 && refreshResponse.status < 500 && refreshResponse.status !== 429) {
                await clearAuth();
                return;
              }

              // 5xx or 429 - retry with backoff
              if (attempt < maxRetries - 1) {
                const delayMs = 500 * Math.pow(2, attempt);
                await new Promise((resolve) => setTimeout(resolve, delayMs));
                continue;
              }
            }

            // No refresh token or all retries failed
            await clearAuth();
            return;
          } catch (error) {
            lastError = error instanceof Error ? error : new Error(String(error));
            console.warn(`Auth initialization attempt ${attempt + 1} failed:`, error);

            // Network error or exception - retry with backoff
            if (attempt < maxRetries - 1) {
              const delayMs = 500 * Math.pow(2, attempt);
              await new Promise((resolve) => setTimeout(resolve, delayMs));
            }
          }
        }

        // All retries exhausted due to network errors - preserve session if we have a valid-looking token
        const finalPayload = safeDecodeJwtPayload(jwtToken);
        if (finalPayload && finalPayload.exp && finalPayload.exp > Math.floor(Date.now() / 1000) - 60) {
          // Token looks valid (has future expiration), don't clear auth due to network issues
          console.warn('Auth initialization failed due to network errors, but token appears valid - keeping session');
          set({
            user: null,
            session: null,
            isAuthenticated: false,
            error: 'Network error - please refresh',
            authChecked: true,
          });
          return;
        }

        console.error('Auth initialization failed after all retries:', lastError);
        await clearAuth();
      },

      login: async (data: LoginRequest) => {
        set({ isLoading: true, error: null });
        try {
          const response = await apiClient.post<{
            user: User;
            session: Session;
          }>('/auth/login', data);

          const { user, session } = response;

          // Store tokens in encrypted storage
          await tokenVault.setAccessToken(session.access_token);
          if (session.refresh_token) {
            await tokenVault.setRefreshToken(session.refresh_token);
          }

          set({
            user: { ...user, isOnline: true },
            session,
            isAuthenticated: true,
            isLoading: false,
            authChecked: true,
            mfaRequired: false,
          });
          broadcastAuthEvent({ type: 'login', timestamp: Date.now(), userId: user.id });
        } catch (error) {
          const axiosData =
            error && typeof error === 'object' && 'response' in error
              ? (error as { response?: { data?: { message?: string } } }).response?.data
              : null;
          const message =
            axiosData?.message ||
            (error instanceof Error ? error.message : 'Login failed');
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
          await tokenVault.clearTokens();
          await tokenVault.clearSessionKey();
          localStorage.removeItem('ff-last-wallet-agent-id');
          broadcastAuthEvent({ type: 'logout', timestamp: Date.now() });
          set({
            user: null,
            session: null,
            isAuthenticated: false,
            error: null,
            mfaRequired: false,
            authChecked: true,
          });
          // Redirect to login if shouldRedirect is true and not already there
          if (shouldRedirect && window.location.pathname !== '/login') {
            const currentPath = window.location.pathname + window.location.search;
            window.location.href = `/login?redirect=${encodeURIComponent(currentPath)}`;
          }
        }
      },

      clearError: () => set({ error: null }),

      refreshSession: async () => {
        const refreshToken = await tokenVault.getRefreshToken();
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
          await tokenVault.setAccessToken(data.token);
          await tokenVault.setRefreshToken(data.refresh_token);

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
          await tokenVault.clearTokens();
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

          // Store tokens in encrypted storage
          await tokenVault.setAccessToken(session.access_token);
          if (session.refresh_token) {
            await tokenVault.setRefreshToken(session.refresh_token);
          }

          set({
            user,
            session,
            isAuthenticated: true,
            isLoading: false,
            mfaRequired: false,
            authChecked: true,
          });
          broadcastAuthEvent({ type: 'login', timestamp: Date.now(), userId: user.id });
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
        isAuthenticated: state.isAuthenticated,
        authChecked: state.authChecked,
      }),
    }
  )
);

export const useAuthStore = authStore;
