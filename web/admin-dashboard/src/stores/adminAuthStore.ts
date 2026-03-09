/**
 * Admin Authentication Store
 * Manages admin user session, authentication state, and security checks.
 * Access token is persisted in sessionStorage so the session survives page refresh.
 */

import { create } from 'zustand';
import type { AdminUser, AdminSession } from '@/types';
import { adminApiClient } from '@/lib/api/adminClient';
import { extendAdminSession } from '@/lib/api/adminAuth';
import { CACHE_KEYS } from '@/lib/constants';

interface AdminAuthState {
  user: AdminUser | null;
  session: AdminSession | null;
  isAuthenticated: boolean;
  mfaVerified: boolean;
  lastActivity: number;
  deviceFingerprint: string | null;
  isIpAllowed: boolean;
  ipCheckReason: string | null;

  // Actions
  login: (session: AdminSession, user: AdminUser) => void;
  logout: () => void;
  verifyMFA: () => void;
  updateActivity: () => void;
  /** Extend session: calls backend for new JWT and updates store. Returns a promise that rejects on failure. */
  extendSession: () => Promise<void>;
  checkSession: () => boolean;
  setUser: (user: AdminUser) => void;
  setSession: (session: AdminSession) => void;
  setDeviceFingerprint: (fingerprint: string) => void;
  setIpAllowed: (allowed: boolean, reason?: string) => void;
}

export const useAdminAuthStore = create<AdminAuthState>((set, get) => ({
  user: null,
  session: null,
  isAuthenticated: false,
  mfaVerified: false,
  lastActivity: Date.now(),
  deviceFingerprint: null,
  isIpAllowed: true,
  ipCheckReason: null,

  login: (session, user) => {
    if (session.access_token) {
      adminApiClient.setSessionToken(session.access_token);
      try {
        sessionStorage.setItem(CACHE_KEYS.ADMIN_ACCESS_TOKEN, session.access_token);
      } catch {
        // sessionStorage may be unavailable (private mode, etc.)
      }
    }

    set({
      user,
      session,
      isAuthenticated: true,
      lastActivity: Date.now(),
      isIpAllowed: true,
      ipCheckReason: null,
    });
  },

  logout: () => {
    adminApiClient.clearSessionToken();
    adminApiClient.clearDeviceFingerprint();
    try {
      sessionStorage.removeItem(CACHE_KEYS.ADMIN_ACCESS_TOKEN);
    } catch {
      /* ignore */
    }

    // Clear all state on logout
    set({
      user: null,
      session: null,
      isAuthenticated: false,
      mfaVerified: false,
      deviceFingerprint: null,
      isIpAllowed: true,
      ipCheckReason: null,
    });

    // Redirect to auth page
    window.location.href = '/auth/login';
  },

  verifyMFA: () => {
    set({ mfaVerified: true, lastActivity: Date.now() });
  },

  updateActivity: () => {
    const state = get();
    const now = Date.now();
    const idleTime = now - state.lastActivity;

    // Check for idle timeout (configurable)
    const idleTimeout = parseInt(
      import.meta.env.VITE_IDLE_TIMEOUT || '900000',
      10
    );
    if (idleTime > idleTimeout) {
      get().logout();
      return;
    }

    set({ lastActivity: now });
  },

  extendSession: async () => {
    const state = get();
    const token =
      state.session?.access_token ??
      (typeof sessionStorage !== 'undefined' ? sessionStorage.getItem(CACHE_KEYS.ADMIN_ACCESS_TOKEN) : null);
    if (!token) {
      get().logout();
      return;
    }
    const bootstrap = await extendAdminSession(token);
    const newToken = bootstrap.session?.access_token ?? token;
    if (bootstrap.session) {
      adminApiClient.setSessionToken(newToken);
      try {
        sessionStorage.setItem(CACHE_KEYS.ADMIN_ACCESS_TOKEN, newToken);
      } catch {
        /* ignore */
      }
      set({
        session: bootstrap.session as AdminSession,
        user: bootstrap.user as AdminUser,
        lastActivity: Date.now(),
      });
    }
  },

  checkSession: () => {
    const state = get();
    if (!state.session) return false;

    const now = Date.now();

    // Check session expiry
    if (now > new Date(state.session.expires_at).getTime()) {
      get().logout();
      return false;
    }

    // Check idle timeout
    const idleTimeout = parseInt(
      import.meta.env.VITE_IDLE_TIMEOUT || '900000',
      10
    );
    const idleTime = now - state.lastActivity;
    if (idleTime > idleTimeout) {
      get().logout();
      return false;
    }

    // Check MFA re-verification requirement
    if (state.session.mfa_verified_at) {
      const mfaReverifyInterval = parseInt(
        import.meta.env.VITE_MFA_REVERIFY_INTERVAL || '14400000',
        10
      );
      const timeSinceMFAVerification =
        now - new Date(state.session.mfa_verified_at).getTime();
      if (timeSinceMFAVerification > mfaReverifyInterval) {
        set({ mfaVerified: false });
      }
    }

    return true;
  },

  setUser: (user) => {
    set({ user });
  },

  setSession: (session) => {
    set({ session });
  },

  setDeviceFingerprint: (fingerprint) => {
    set({ deviceFingerprint: fingerprint });
  },

  setIpAllowed: (allowed, reason) => {
    set({
      isIpAllowed: allowed,
      ipCheckReason: reason || null,
    });
  },
}));

/**
 * Session monitoring hook
 * Automatically checks and updates session state
 */
export function useSessionMonitor() {
  const checkSession = useAdminAuthStore((state) => state.checkSession);
  const updateActivity = useAdminAuthStore((state) => state.updateActivity);

  // Effect is handled in AdminSessionMonitor component
  return { checkSession, updateActivity };
}
