import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { User, LoginRequest, SignupRequest, SignupResponse } from "@/types";
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
  logout: () => void;
  clearError: () => void;
  initialize: () => void;
  refreshSession: () => Promise<void>;
}

// Create the store
const authStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      session: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      initialize: async () => {
        console.log("AuthStore: Initializing auth state");

        // First, check for JWT token from OAuth flow
        const jwtToken = localStorage.getItem("sb-access-token");
        if (jwtToken) {
          try {
            // Validate JWT token with backend
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
                name: userData.user.name || '',
                avatar: userData.user.avatar || '',
                tenantId: userData.user.tenant_id || 'default',
                plan: 'starter', // Default plan
                role: userData.user.role,
                createdAt: userData.user.created_at,
                updatedAt: userData.user.updated_at,
              };

              // Create mock session for compatibility
              const mockSession = {
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
                session: mockSession as any,
                isAuthenticated: true,
              });

              console.log("AuthStore: JWT token validated", { user: user.email });
              return;
            } else {
              console.error("AuthStore: JWT validation failed");
              // Clear invalid tokens
              localStorage.removeItem("sb-access-token");
              localStorage.removeItem("sb-refresh-token");
              set({
                user: null,
                session: null,
                isAuthenticated: false,
                error: "Invalid or expired token",
              });
              return;
            }
          } catch (error) {
            console.error("AuthStore: Error validating JWT token", error);
            // Clear invalid tokens on validation error
            localStorage.removeItem("sb-access-token");
            localStorage.removeItem("sb-refresh-token");
            set({
              user: null,
              session: null,
              isAuthenticated: false,
              error: "Token validation failed",
            });
          }
        }

        // No fallback session - FunctionFly uses JWT tokens only
        // If no valid JWT token, user remains unauthenticated
        console.log("AuthStore: No valid authentication found");

        // No auth state change listener needed for JWT-based auth
      },

      refreshSession: async () => {
        // FunctionFly doesn't use refresh tokens in the same way
        // JWT tokens are validated on each request
        console.log("AuthStore: Refresh session not needed for JWT tokens");
      },

      login: async (data) => {
        console.log("AuthStore: Starting login process with email:", data.email);
        set({ isLoading: true, error: null });
        try {
          // Call FunctionFly auth endpoint
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
            // #region agent log — capture 500 response for debugging
            if (response.status >= 500) {
              console.warn('AuthStore: 500 response body (first 500 chars):', text.slice(0, 500));
            }
            // #endregion
            let msg = 'Login failed';
            let detail: string | undefined;
            try {
              const errorData = text ? JSON.parse(text) : {};
              msg = errorData.message || msg;
              detail = errorData.detail; // Server sends root cause when DEVELOPMENT=true
            } catch {
              // Non-JSON body (e.g. proxy could not reach API — 500 from Vite proxy, not from backend)
              if (response.status >= 500) {
                msg = 'Cannot reach the API.';
                // Show proxy error (e.g. "Proxy error: connect ECONNREFUSED 172.18.0.7:8080") when present
                if (text && text.trim().startsWith('Proxy error:')) {
                  detail = text.trim() + ' — orchestrator-api may have exited; run: docker compose ps && docker compose logs orchestrator-api --tail=30';
                } else {
                  detail = 'Run both in Docker: docker compose --profile dashboard up -d. If the API container exited, check: docker compose logs orchestrator-api --tail=30';
                }
              }
            }
            const fullMessage = detail ? `${msg} ${detail}` : msg;
            throw new Error(fullMessage);
          }

          const authData = await response.json();
          console.log("AuthStore: Login response received");
          console.log("AuthStore: Response data keys:", Object.keys(authData));
          console.log("AuthStore: Has token field:", 'token' in authData);
          console.log("AuthStore: Token value:", authData.token ? "present" : "missing");

          // Store JWT token for future requests
          localStorage.setItem('sb-access-token', authData.token);
          console.log("AuthStore: Token saved to localStorage:", authData.token?.substring(0, 20) + "..." || "null");
          console.log("AuthStore: Verifying token storage:", localStorage.getItem('sb-access-token') ? "SUCCESS" : "FAILED");

          // Create user object from response
          const user: User = {
            id: authData.user.id,
            email: authData.user.email || '',
            name: authData.user.name || '',
            avatar: authData.user.avatar || '',
            tenantId: authData.user.tenant_id || 'default',
            plan: 'starter', // Default plan
            role: authData.user.role,
            createdAt: authData.user.created_at,
            updatedAt: authData.user.updated_at,
          };

          // Create mock session for compatibility
          const mockSession = {
            access_token: authData.token,
            refresh_token: '',
            expires_at: Math.floor(Date.now() / 1000) + (24 * 60 * 60), // 24 hours
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
            session: mockSession as any,
            isAuthenticated: true,
            isLoading: false,
          });

          // Reload API client token
          console.log("AuthStore: Calling apiClient.reloadToken()");
          apiClient.reloadToken();

          console.log("AuthStore: Authentication state updated", { isAuthenticated: true, user: user.email });
        } catch (error) {
          console.error("AuthStore: Login failed", error);
          let errorMessage = "Login failed";
          if (error instanceof Error) {
            errorMessage = error.message;
            // Network/connection errors (fetch throws TypeError)
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
          // Call FunctionFly auth endpoint
          const response = await fetch(`${import.meta.env.VITE_NEON_AUTH_URL}/signup`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({
              email: data.email,
              password: data.password,
              confirmPassword: data.password, // Assuming password confirmation
              name: data.name,
              termsAccepted: true, // Assuming terms are accepted in signup form
            }),
          });

          if (!response.ok) {
            const errorData = await response.json().catch(() => ({ message: 'Signup failed' }));
            throw new Error(errorData.message || 'Signup failed');
          }

          const authData = await response.json();

          // For signup, we don't expect user data since verification is required first
          set({ isLoading: false });

          return {
            message: authData.message || 'Account created successfully',
            emailSent: authData.emailSent || false,
            requiresVerification: authData.requiresVerification || false,
          };
        } catch (error) {
          console.error("AuthStore: Signup failed", error);
          const errorMessage = error instanceof Error ? error.message : "Signup failed";
          set({
            error: errorMessage,
            isLoading: false,
          });
          throw error;
        }
      },

      logout: async () => {
        console.log("AuthStore: Logging out user");
        try {
          // Clear JWT token
          localStorage.removeItem('sb-access-token');
          localStorage.removeItem('sb-provider-token');

          set({
            user: null,
            session: null,
            isAuthenticated: false,
            error: null,
          });
          console.log("AuthStore: User logged out, state cleared");
        } catch (error) {
          console.error("AuthStore: Logout failed", error);
          // Even if logout fails, clear local state
          set({
            user: null,
            session: null,
            isAuthenticated: false,
            error: null,
          });
        }
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
