import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { User, Session, LoginRequest, SignupRequest, SignupResponse } from "@/types";
import { apiClient } from "@/api/client";

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
  error: string | null;
  mfaRequired: boolean;

  // Actions
  login: (data: LoginRequest) => Promise<void>;
  signup: (data: SignupRequest) => Promise<SignupResponse>;
  logout: () => Promise<void>;
  clearError: () => void;
  initialize: () => Promise<void>;
  refreshSession: () => Promise<void>;
  verifyMFA: (code: string) => Promise<void>;
  /** Sync plan from an authoritative source (e.g. GET /users/me) so UI shows correct plan. */
  setUserPlan: (plan: string) => void;
}

// Create the store
const authStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      session: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,
      mfaRequired: false,

      initialize: async () => {
        const jwtToken = localStorage.getItem("ff-access-token");
        const refreshToken = localStorage.getItem("ff-refresh-token");

        if (!jwtToken) {
          set({ user: null, session: null, isAuthenticated: false });
          return;
        }

        try {
          // Check if token is expired locally first
          const payload = JSON.parse(atob(jwtToken.split('.')[1]));
          const currentTime = Math.floor(Date.now() / 1000);
          const expiresAt = payload.exp || 0;

          // If token is still valid, validate with backend
          if (expiresAt > currentTime) {
            const response = await fetch(`${import.meta.env.VITE_API_URL}/auth/validate`, {
              method: "GET",
              headers: {
                "Authorization": `Bearer ${jwtToken}`,
              },
            });

            if (response.ok) {
              const userData = await response.json();
              const user: User = {
                id: userData.user.id,
                email: userData.user.email,
                username: userData.user.username,
                companyName: userData.user.company_name,
                name: userData.user.name || '',
                avatar: userData.user.avatar || '',
                tenantId: userData.user.tenant_id || 'default',
                plan: userData.user.plan ?? 'starter',
                role: userData.user.role,
                createdAt: userData.user.created_at,
                updatedAt: userData.user.updated_at,
              };

              const session: Session = {
                access_token: jwtToken,
                refresh_token: refreshToken || "",
                expires_at: expiresAt,
                token_type: "bearer",
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
              });
              return;
            }
          }

          // Token is expired or invalid, try to refresh if we have a refresh token
          if (refreshToken) {
            try {
              const refreshResponse = await fetch(`${import.meta.env.VITE_API_URL}/auth/refresh`, {
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
                  plan: refreshData.user.plan ?? 'starter',
                  role: refreshData.user.role,
                  createdAt: refreshData.user.created_at,
                  updatedAt: refreshData.user.updated_at,
                };

                // Decode new token to get expiration
                const newPayload = JSON.parse(atob(refreshData.token.split('.')[1]));
                const newExpiresAt = newPayload.exp || (Math.floor(Date.now() / 1000) + (30 * 60)); // 30 minutes fallback

                const session: Session = {
                  access_token: refreshData.token,
                  refresh_token: refreshData.refresh_token,
                  expires_at: newExpiresAt,
                  token_type: "bearer",
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
                });
                return;
              }
            } catch (refreshError) {
              console.warn('Token refresh failed:', refreshError);
            }
          }

          // If we get here, both token validation and refresh failed
          localStorage.removeItem("ff-access-token");
          localStorage.removeItem("ff-refresh-token");
          set({
            user: null,
            session: null,
            isAuthenticated: false,
            error: null,
          });
        } catch {
          // Network or parse error during validation — clear auth state
          localStorage.removeItem("ff-access-token");
          localStorage.removeItem("ff-refresh-token");
          set({
            user: null,
            session: null,
            isAuthenticated: false,
            error: null,
          });
        }
      },

      refreshSession: async () => {
        // JWT tokens are stateless; re-run initialize to re-validate the stored token
        // and refresh in-memory state from the backend.
        await authStore.getState().initialize();
      },

      login: async (data) => {
        set({ isLoading: true, error: null });
        try {
          const response = await fetch(`${import.meta.env.VITE_API_URL}/auth/login`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({
              email: data.email,
              password: data.password,
            }),
          });

          if (!response.ok) {
            const text = await response.text();
            let msg = 'Login failed';
            let detail: string | undefined;
            try {
              const errorData = text ? JSON.parse(text) : {};
              msg = errorData.message || msg;
              // Only expose detail in development mode
              if (import.meta.env.DEV) {
                detail = errorData.detail;
              }
            } catch {
              // Non-JSON body (proxy error, network issue, etc.)
              if (response.status >= 500) {
                msg = 'Cannot reach the API. Please try again later.';
              }
            }
            const fullMessage = detail ? `${msg} ${detail}` : msg;
            throw new Error(fullMessage);
          }

          const authData = await response.json();

          // Check if MFA is required
          if (authData.mfaRequired) {
            // Store temp token for MFA verification
            if (authData.tempToken) {
              localStorage.setItem('sb-mfa-temp-token', authData.tempToken);
            }
            set({
              mfaRequired: true,
              isLoading: false,
              error: null,
            });
            throw new Error('MFA_REQUIRED');
          }

          if (!authData.token) {
            throw new Error('Authentication response missing token');
          }

          // Store JWT token and refresh token for future requests
          localStorage.setItem('ff-access-token', authData.token);
          if (authData.refresh_token) {
            localStorage.setItem('ff-refresh-token', authData.refresh_token);
          }

          // Create user object from response
          const user: User = {
            id: authData.user.id,
            email: authData.user.email || '',
            username: authData.user.username,
            companyName: authData.user.company_name,
            name: authData.user.name || '',
            avatar: authData.user.avatar || '',
            tenantId: authData.user.tenant_id || 'default',
            plan: authData.user.plan ?? 'starter',
            role: authData.user.role,
            createdAt: authData.user.created_at,
            updatedAt: authData.user.updated_at,
          };

          // Decode JWT to get actual expiration time
          const payload = JSON.parse(atob(authData.token.split('.')[1]));
          const expiresAt = payload.exp || (Math.floor(Date.now() / 1000) + (24 * 60 * 60));

          const loginSession: Session = {
            access_token: authData.token,
            refresh_token: '',
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
            session: loginSession,
            isAuthenticated: true,
            isLoading: false,
          });

          // Reload API client token cache
          apiClient.reloadToken();
        } catch (error) {
          let errorMessage = "Login failed";
          if (error instanceof Error) {
            errorMessage = error.message;
            if (error.name === "TypeError" && (error.message === "Failed to fetch" || error.message.includes("NetworkError"))) {
              errorMessage = "Cannot reach the server. Check that the API is running and try again.";
            }
          }
          set({
            error: errorMessage,
            isLoading: false,
            isAuthenticated: false,
            user: null,
          });
          throw error;
        }
      },

      signup: async (data) => {
        set({ isLoading: true, error: null });
        try {
          const response = await fetch(`${import.meta.env.VITE_API_URL}/auth/signup`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({
              email: data.email,
              password: data.password,
              confirmPassword: data.password,
              name: data.name,
              username: data.username || undefined,
              companyName: data.companyName || undefined,
              termsAccepted: true,
            }),
          });

          if (!response.ok) {
            const errorData = await response.json().catch(() => ({ message: 'Signup failed' }));
            throw new Error(errorData.message || 'Signup failed');
          }

          const authData = await response.json();

          set({ isLoading: false });

          return {
            message: authData.message || 'Account created successfully',
            emailSent: authData.emailSent || false,
            requiresVerification: authData.requiresVerification || false,
          };
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : "Signup failed";
          set({
            error: errorMessage,
            isLoading: false,
          });
          throw error;
        }
      },

      logout: async () => {
        const token = localStorage.getItem('ff-access-token');

        // Invalidate the session server-side (best-effort — don't block local clear on failure)
        if (token) {
          try {
            await fetch(`${import.meta.env.VITE_API_URL}/auth/logout`, {
              method: 'POST',
              headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json',
              },
            });
          } catch {
            // Network error during logout — proceed with local cleanup regardless
          }
        }

        // Clear all local auth state
        localStorage.removeItem('ff-access-token');
        localStorage.removeItem('ff-refresh-token');
        apiClient.clearToken();

        set({
          user: null,
          session: null,
          isAuthenticated: false,
          error: null,
        });
      },

      clearError: () => set({ error: null }),

      setUserPlan: (plan: string) =>
        set((state) =>
          state.user ? { user: { ...state.user, plan } } : {}
        ),

      verifyMFA: async (code: string) => {
        set({ isLoading: true, error: null });
        try {
          const token = localStorage.getItem('ff-access-token');
          const response = await fetch(`${import.meta.env.VITE_API_URL}/auth/mfa/verify`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${token}`,
            },
            body: JSON.stringify({ code }),
          });

          if (!response.ok) {
            const errorData = await response.json().catch(() => ({ message: 'MFA verification failed' }));
            throw new Error(errorData.message || 'Invalid code. Please try again.');
          }

          const authData = await response.json();

          // Update session with new token if provided
          if (authData.token) {
            localStorage.setItem('ff-access-token', authData.token);
          }

          set({
            mfaRequired: false,
            isLoading: false,
          });

          // Reload API client token cache
          apiClient.reloadToken();
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : "MFA verification failed";
          set({
            error: errorMessage,
            isLoading: false,
          });
          throw error;
        }
      },
    }),
    {
      name: "auth-storage",
      partialize: (state) => ({
        user: state.user,
        session: state.session,
        isAuthenticated: state.isAuthenticated
      }),
    }
  )
);

export const useAuthStore = authStore;
