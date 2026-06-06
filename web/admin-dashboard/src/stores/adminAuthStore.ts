/**
 * Admin Authentication Store
 * Manages admin user session, authentication state, and security checks.
 * Access token is persisted in sessionStorage so the session survives page refresh.
 */

import { extendAdminSession } from '@/lib/api/adminAuth';
import { adminApiClient } from '@/lib/api/adminClient';
import { CACHE_KEYS } from '@/lib/constants';
import { logger } from '@/lib/monitoring/logger';
import { trackSecurityEvent } from '@/lib/monitoring/securityEvents';
import { clearCsrfToken, refreshCsrfToken } from '@/lib/security/csrf';
import type { AdminSession, AdminUser } from '@/types';
import { create } from 'zustand';

interface LastLoginInfo {
  ip_address: string;
  device_name: string;
  timestamp: string;
  suspicious: boolean;
}

interface AdminAuthState {
  user: AdminUser | null;
  session: AdminSession | null;
  isAuthenticated: boolean;
  mfaVerified: boolean;
  lastActivity: number;
  deviceFingerprint: string | null;
  isIpAllowed: boolean;
  ipCheckReason: string | null;
  lastLoginInfo: LastLoginInfo | null;
  sessionIpAddress: string | null;
  activityLog: Array<{ timestamp: number; action: string }>;

  // Actions
  login: (session: AdminSession, user: AdminUser, lastLoginInfo?: LastLoginInfo) => void;
  logout: () => void;
  logoutAllSessions: () => Promise<void>;
  /**
   * Submit a 6-digit TOTP code to the server. Resolves on success; rejects
   * with an Error on failure (invalid code, network, etc.).
   */
  verifyMFA: (code: string) => Promise<void>;
  updateActivity: () => void;
  /** Extend session: calls backend for new JWT and updates store. Returns a promise that rejects on failure. */
  extendSession: () => Promise<void>;
  checkSession: () => boolean;
  setUser: (user: AdminUser) => void;
  setSession: (session: AdminSession) => void;
  setDeviceFingerprint: (fingerprint: string) => void;
  setIpAllowed: (allowed: boolean, reason?: string) => void;
  setLastLoginInfo: (info: LastLoginInfo | null) => void;
  reportSuspiciousLogin: () => Promise<void>;
  initialize: () => Promise<void>;
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
  lastLoginInfo: null,
  sessionIpAddress: null,
  activityLog: [],

  login: (session, user, lastLoginInfo) => {
    if (session.access_token) {
      adminApiClient.setSessionToken(session.access_token);
      try {
        sessionStorage.setItem(CACHE_KEYS.ADMIN_ACCESS_TOKEN, session.access_token);
      } catch {
        // sessionStorage may be unavailable (private mode, etc.)
      }
    }

    // Track initial IP address for session
    const clientIp =
      session.ip_address ||
      (typeof window !== 'undefined'
        ? (window as Window & { _clientIp?: string })._clientIp
        : null);

    set({
      user,
      session,
      isAuthenticated: true,
      lastActivity: Date.now(),
      isIpAllowed: true,
      ipCheckReason: null,
      lastLoginInfo: lastLoginInfo || null,
      sessionIpAddress: clientIp || session.ip_address,
      activityLog: [{ timestamp: Date.now(), action: 'login' }],
    });

    // Pre-fetch CSRF token so mutating requests don't fail before it's fetched on-demand
    refreshCsrfToken().catch(() => {
      /* silent — first mutating request will retry */
    });

    // Track security event
    trackSecurityEvent('login_success', {
      ip_address: session.ip_address,
      device_fingerprint: session.device_fingerprint,
    });
  },

  logout: () => {
    const state = get();

    // Track logout activity
    trackSecurityEvent('logout', {
      session_id: state.session?.id,
    });

    adminApiClient.clearSessionToken();
    adminApiClient.clearDeviceFingerprint();
    clearCsrfToken();
    try {
      sessionStorage.removeItem(CACHE_KEYS.ADMIN_ACCESS_TOKEN);
    } catch {
      /* ignore */
    }

    // Clear all state on logout.
    // Do NOT use window.location.href — that causes a hard reload which re-runs
    // AdminAuthRestore and can bounce the user back to / if a stale token remains.
    // ProtectedRoute reads isAuthenticated from this store and will redirect to
    // /auth/login via React Router as soon as isAuthenticated becomes false.
    set({
      user: null,
      session: null,
      isAuthenticated: false,
      mfaVerified: false,
      deviceFingerprint: null,
      isIpAllowed: true,
      ipCheckReason: null,
      lastLoginInfo: null,
      sessionIpAddress: null,
      activityLog: [],
    });
  },

  logoutAllSessions: async () => {
    const state = get();
    if (!state.session?.id) return;

    try {
      await adminApiClient.post('/auth/logout-all-sessions', {
        current_session_id: state.session.id,
      });
      get().logout();
    } catch (error) {
      logger.error('Failed to logout all sessions', { error });
      // Still logout locally even if remote fails
      get().logout();
    }
  },

  /**
   * Submit an MFA TOTP code to the server and, on success, mark the session
   * as MFA-verified. Rejects on backend error so the UI can show a real
   * failure message instead of dismissing the prompt.
   */
   verifyMFA: async (code: string) => {
     try {
       const response = await adminApiClient.post<{ verified?: boolean; error?: string }>('/auth/mfa/verify', {
         code,
         method: 'totp',
       }, { _skipCsrf: true } as any);
       const data = response.data;
       if (!data?.verified) {
         trackSecurityEvent('mfa_verify_failed', { reason: data?.error || 'invalid_code' });
         throw new Error(data?.error || 'MFA verification failed');
       }
      // Refresh the session bootstrap so session.mfa_verified_at picks up the
      // server's new mfa_last_used and the re-prompt timer resets.
      const token =
        (typeof sessionStorage !== 'undefined' &&
          sessionStorage.getItem(CACHE_KEYS.ADMIN_ACCESS_TOKEN)) ||
        get().session?.access_token ||
        null;
      if (token) {
        const bootstrap = await extendAdminSession(token);
        if (bootstrap.session) {
          adminApiClient.setSessionToken(bootstrap.session.access_token ?? token);
          try {
            sessionStorage.setItem(
              CACHE_KEYS.ADMIN_ACCESS_TOKEN,
              bootstrap.session.access_token ?? token
            );
          } catch {
            /* ignore */
          }
          set({
            session: bootstrap.session as AdminSession,
            user: (bootstrap.user as AdminUser) ?? get().user,
            mfaVerified: true,
            lastActivity: Date.now(),
          });
        }
      } else {
        set({ mfaVerified: true, lastActivity: Date.now() });
      }
      trackSecurityEvent('mfa_verified');
    } catch (error) {
      trackSecurityEvent('mfa_verify_failed', { reason: 'request_error' });
      throw error;
    }
  },

  updateActivity: () => {
    const state = get();
    const now = Date.now();
    const idleTime = now - state.lastActivity;

    // Check for idle timeout (configurable)
    const idleTimeout = parseInt(import.meta.env.VITE_IDLE_TIMEOUT || '900000', 10);
    if (idleTime > idleTimeout) {
      trackSecurityEvent('session_timeout_warning');
      get().logout();
      return;
    }

    // Log activity periodically (every 5 minutes)
    const lastLog = state.activityLog[state.activityLog.length - 1];
    if (!lastLog || now - lastLog.timestamp > 300000) {
      set({
        lastActivity: now,
        activityLog: [...state.activityLog, { timestamp: now, action: 'activity' }],
      });
    } else {
      set({ lastActivity: now });
    }
  },

  extendSession: async () => {
    const state = get();
    const token =
      state.session?.access_token ??
      (typeof sessionStorage !== 'undefined'
        ? sessionStorage.getItem(CACHE_KEYS.ADMIN_ACCESS_TOKEN)
        : null);
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
        activityLog: [...state.activityLog, { timestamp: Date.now(), action: 'session_extended' }],
      });
    }
  },

  checkSession: () => {
    const state = get();
    if (!state.session) return false;

    const now = Date.now();

    // Check session expiry
    if (now > new Date(state.session.expires_at).getTime()) {
      trackSecurityEvent('session_expired');
      get().logout();
      return false;
    }

    // Check idle timeout
    const idleTimeout = parseInt(import.meta.env.VITE_IDLE_TIMEOUT || '900000', 10);
    const idleTime = now - state.lastActivity;
    if (idleTime > idleTimeout) {
      trackSecurityEvent('session_timeout_warning');
      get().logout();
      return false;
    }

    // Check MFA re-verification requirement
    if (state.session.mfa_verified_at) {
      const mfaReverifyInterval = parseInt(
        import.meta.env.VITE_MFA_REVERIFY_INTERVAL || '14400000',
        10
      );
      const timeSinceMFAVerification = now - new Date(state.session.mfa_verified_at).getTime();
      if (timeSinceMFAVerification > mfaReverifyInterval) {
        set({ mfaVerified: false });
        trackSecurityEvent('mfa_required');
      }
    }

    // Check IP address match (if whitelisting enabled)
    if (state.sessionIpAddress && state.session.ip_address !== state.sessionIpAddress) {
      trackSecurityEvent('suspicious_activity', {
        reason: 'ip_mismatch',
        original_ip: state.session.ip_address,
        current_ip: state.sessionIpAddress,
      });
      // Don't logout but flag as suspicious
      set({ ipCheckReason: 'IP address mismatch detected' });
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
    adminApiClient.setDeviceFingerprint(fingerprint);
  },

  setIpAllowed: (allowed, reason) => {
    set({
      isIpAllowed: allowed,
      ipCheckReason: reason || null,
    });
  },

  setLastLoginInfo: (info) => {
    set({ lastLoginInfo: info });
  },

  reportSuspiciousLogin: async () => {
    const state = get();
    if (!state.session?.id || !state.lastLoginInfo) return;

    try {
      await adminApiClient.post('/security/report-suspicious-login', {
        session_id: state.session.id,
        ip_address: state.lastLoginInfo.ip_address,
        device_name: state.lastLoginInfo.device_name,
        timestamp: state.lastLoginInfo.timestamp,
        reason: 'user_reported_suspicious',
      });
      trackSecurityEvent('suspicious_activity', {
        reason: 'user_reported',
        last_login: state.lastLoginInfo,
      });
    } catch {
      // Silently fail
    }
  },

  initialize: async () => {
    try {
      // Check for existing token in sessionStorage
      const token =
        typeof sessionStorage !== 'undefined'
          ? sessionStorage.getItem(CACHE_KEYS.ADMIN_ACCESS_TOKEN)
          : null;

      if (token) {
        // Try to bootstrap session with existing token
        try {
          const bootstrap = await extendAdminSession(token);
          if (bootstrap.session && bootstrap.user) {
            get().login(bootstrap.session as AdminSession, bootstrap.user as AdminUser);
            return;
          }
        } catch (error) {
          // Token might be expired, clear it
          logger.warn('Failed to restore admin session', { error });
          try {
            sessionStorage.removeItem(CACHE_KEYS.ADMIN_ACCESS_TOKEN);
          } catch {
            /* ignore */
          }
        }
      }
    } catch (error) {
      logger.warn('Failed to initialize admin auth', { error });
    }
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
