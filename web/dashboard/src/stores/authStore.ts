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

  // Actions
  login: (data: LoginRequest) => Promise<void>;
  signup: (data: SignupRequest) => Promise<SignupResponse>;
  logout: () => Promise<void>;
  clearError: () => void;
  initialize: () => Promise<void>;
  refreshSession: () => Promise<void>;
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

      initialize: async () => {
        const jwtToken = localStorage.getItem("sb-access-token");
        if (!jwtToken) {
          set({ user: null, session: null, isAuthenticated: false });
          return;
        }

        try {
          // Validate JWT token with backend and retrieve safe user data
          const response = await fetch(`${import.meta.env.VITE_NEON_AUTH_URL}/validate`, {
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
              plan: 'starter',
              role: userData.user.role,
              createdAt: userData.user.created_at,
              updatedAt: userData.user.updated_at,
            };

            const session: Session = {
              access_token: jwtToken,
              refresh_token: "",
              expires_at: Math.floor(Date.now() / 1000) + (24 * 60 * 60),
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
          } else {
            // Token invalid or expired — clear all auth state
            localStorage.removeItem("sb-access-token");
            localStorage.removeItem("sb-refresh-token");
            set({
              user: null,
              session: null,
              isAuthenticated: false,
              error: null,
            });
          }
        } catch {
          // Network or parse error during validation — clear auth state
          localStorage.removeItem("sb-access-token");
          localStorage.removeItem("sb-refresh-token");
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
          const response = await fetch(`${import.meta.env.VITE_NEON_AUTH_URL}/login`, {
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

          if (!authData.token) {
            throw new Error('Authentication response missing token');
          }

          // Store JWT token for future requests
          localStorage.setItem('sb-access-token', authData.token);

          // Create user object from response
          const user: User = {
            id: authData.user.id,
            email: authData.user.email || '',
            username: authData.user.username,
            companyName: authData.user.company_name,
            name: authData.user.name || '',
            avatar: authData.user.avatar || '',
            tenantId: authData.user.tenant_id || 'default',
            plan: 'starter',
            role: authData.user.role,
            createdAt: authData.user.created_at,
            updatedAt: authData.user.updated_at,
          };

          const loginSession: Session = {
            access_token: authData.token,
            refresh_token: '',
            expires_at: Math.floor(Date.now() / 1000) + (24 * 60 * 60),
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
          const response = await fetch(`${import.meta.env.VITE_NEON_AUTH_URL}/signup`, {
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
        const token = localStorage.getItem('sb-access-token');

        // Invalidate the session server-side (best-effort — don't block local clear on failure)
        if (token) {
          try {
            await fetch(`${import.meta.env.VITE_NEON_AUTH_URL}/logout`, {
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
        localStorage.removeItem('sb-access-token');
        localStorage.removeItem('sb-provider-token');
        apiClient.clearToken();

        set({
          user: null,
          session: null,
          isAuthenticated: false,
          error: null,
        });
      },

      clearError: () => set({ error: null }),
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
