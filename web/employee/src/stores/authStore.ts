import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface User {
  id: string;
  email: string;
  name: string;
  username?: string;
  avatar_url?: string;
  role?: string;
}

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  setUser: (user: User | null) => void;
  setToken: (token: string | null) => void;
  logout: () => void;
  initialize: () => Promise<void>;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: true,

      setUser: (user) => set({ user, isAuthenticated: !!user }),

      setToken: (token) => {
        if (token) {
          localStorage.setItem('ff_token', token);
        } else {
          localStorage.removeItem('ff_token');
        }
        set({ token });
      },

      logout: () => {
        localStorage.removeItem('ff_token');
        set({ user: null, token: null, isAuthenticated: false });
        window.location.href = '/api/auth/logout';
      },

      initialize: async () => {
        const token = localStorage.getItem('ff_token');
        if (!token) {
          set({ isLoading: false });
          return;
        }

        try {
          const res = await fetch('/v1/auth/validate', {
            headers: { Authorization: `Bearer ${token}` },
          });

          if (res.ok) {
            const data = await res.json();
            set({
              user: data.user,
              token,
              isAuthenticated: true,
              isLoading: false,
            });
          } else {
            localStorage.removeItem('ff_token');
            set({ user: null, token: null, isAuthenticated: false, isLoading: false });
          }
        } catch {
          set({ isLoading: false });
        }
      },
    }),
    {
      name: 'fwos-auth',
      partialize: (state) => ({ user: state.user, token: state.token, isAuthenticated: state.isAuthenticated }),
    },
  ),
);
