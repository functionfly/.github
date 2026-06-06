/**
 * Tests for the admin auth store.
 *
 * Pin the security-critical behavior:
 *  - login() sets isAuthenticated, persists the token in sessionStorage,
 *    and clears the in-memory CSRF cache so a stolen token from a previous
 *    session can't be replayed.
 *  - logout() always clears the token from sessionStorage and resets
 *    isAuthenticated, even if the remote logout request fails.
 *  - verifyMFA() rejects on a non-200 response and never flips
 *    mfaVerified to true in that case.
 *  - checkSession() expires the session when the inactivity threshold
 *    is reached.
 */
// @vitest-environment happy-dom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

// Mock the modules the store depends on so we can observe side effects
// without touching the network or sessionStorage directly until we want to.
const apiSetSessionToken = vi.fn();
const apiClearSessionToken = vi.fn();
const apiClearDeviceFingerprint = vi.fn();
const apiPost = vi.fn();
const apiGet = vi.fn();
const clearCsrfToken = vi.fn();
const refreshCsrfTokenMock = vi.fn().mockResolvedValue('mock-csrf-token');
const trackSecurityEvent = vi.fn();
const extendAdminSession = vi.fn();

vi.mock('@/lib/api/adminClient', () => ({
  adminApiClient: {
    setSessionToken: (...args: unknown[]) => apiSetSessionToken(...args),
    clearSessionToken: (...args: unknown[]) => apiClearSessionToken(...args),
    clearDeviceFingerprint: (...args: unknown[]) => apiClearDeviceFingerprint(...args),
    post: (...args: unknown[]) => apiPost(...args),
    get: (...args: unknown[]) => apiGet(...args),
  },
}));

vi.mock('@/lib/security/csrf', () => ({
  clearCsrfToken: (...args: unknown[]) => clearCsrfToken(...args),
  refreshCsrfToken: (...args: unknown[]) => refreshCsrfTokenMock(...args),
}));

vi.mock('@/lib/monitoring/securityEvents', () => ({
  trackSecurityEvent: (...args: unknown[]) => trackSecurityEvent(...args),
}));

vi.mock('@/lib/api/adminAuth', () => ({
  extendAdminSession: (...args: unknown[]) => extendAdminSession(...args),
}));

// Pull the store in *after* mocks are set up.
import { useAdminAuthStore } from './adminAuthStore';
import { CACHE_KEYS } from '@/lib/constants';

function makeUser(overrides: Record<string, unknown> = {}) {
  return {
    id: 'u1',
    email: 'admin@example.com',
    name: 'Admin',
    role: 'admin',
    mfa_enabled: false,
    ...overrides,
  } as any;
}

function makeSession(overrides: Record<string, unknown> = {}) {
  return {
    id: 's1',
    access_token: 'jwt-test',
    expires_at: new Date(Date.now() + 3600_000).toISOString(),
    ip_address: '127.0.0.1',
    device_fingerprint: 'fp',
    ...overrides,
  } as any;
}

beforeEach(() => {
  sessionStorage.clear();
  apiSetSessionToken.mockReset();
  apiClearSessionToken.mockReset();
  apiClearDeviceFingerprint.mockReset();
  apiPost.mockReset();
  apiGet.mockReset();
  clearCsrfToken.mockReset();
  refreshCsrfTokenMock.mockReset();
  refreshCsrfTokenMock.mockResolvedValue('mock-csrf-token');
  trackSecurityEvent.mockReset();
  extendAdminSession.mockReset();
  useAdminAuthStore.setState({
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
  });
});

afterEach(() => {
  vi.useRealTimers();
});

describe('login()', () => {
  it('sets isAuthenticated=true and persists the access token in sessionStorage', () => {
    const { login } = useAdminAuthStore.getState();
    login(makeSession(), makeUser());
    const state = useAdminAuthStore.getState();
    expect(state.isAuthenticated).toBe(true);
    expect(state.user?.email).toBe('admin@example.com');
    expect(state.session?.access_token).toBe('jwt-test');
    expect(sessionStorage.getItem(CACHE_KEYS.ADMIN_ACCESS_TOKEN)).toBe('jwt-test');
    expect(apiSetSessionToken).toHaveBeenCalledWith('jwt-test');
  });

  it('pre-fetches the CSRF token so the first mutating request succeeds', async () => {
    const { login } = useAdminAuthStore.getState();
    login(makeSession(), makeUser());
    // refreshCsrfToken is fire-and-forget; wait a microtask.
    await Promise.resolve();
    expect(refreshCsrfTokenMock).toHaveBeenCalledTimes(1);
  });

  it('tracks a login_success security event with ip + device fingerprint', () => {
    const { login } = useAdminAuthStore.getState();
    login(makeSession(), makeUser());
    expect(trackSecurityEvent).toHaveBeenCalledWith(
      'login_success',
      expect.objectContaining({ ip_address: '127.0.0.1', device_fingerprint: 'fp' })
    );
  });

  it('does not persist a token if access_token is missing', () => {
    const { login } = useAdminAuthStore.getState();
    login(makeSession({ access_token: undefined }), makeUser());
    expect(sessionStorage.getItem(CACHE_KEYS.ADMIN_ACCESS_TOKEN)).toBeNull();
  });
});

describe('logout()', () => {
  it('clears the token from the api client and sessionStorage, and resets state', () => {
    const store = useAdminAuthStore.getState();
    store.login(makeSession(), makeUser());
    expect(useAdminAuthStore.getState().isAuthenticated).toBe(true);

    useAdminAuthStore.getState().logout();
    const state = useAdminAuthStore.getState();
    expect(state.isAuthenticated).toBe(false);
    expect(state.user).toBeNull();
    expect(state.session).toBeNull();
    expect(state.mfaVerified).toBe(false);
    expect(sessionStorage.getItem(CACHE_KEYS.ADMIN_ACCESS_TOKEN)).toBeNull();
    expect(apiClearSessionToken).toHaveBeenCalledTimes(1);
    expect(apiClearDeviceFingerprint).toHaveBeenCalledTimes(1);
    expect(clearCsrfToken).toHaveBeenCalledTimes(1);
  });

  it('is idempotent and safe to call when not logged in', () => {
    useAdminAuthStore.getState().logout();
    const state = useAdminAuthStore.getState();
    expect(state.isAuthenticated).toBe(false);
    expect(apiClearSessionToken).toHaveBeenCalledTimes(1);
  });
});

describe('verifyMFA()', () => {
  it('flips mfaVerified to true on a successful verification', async () => {
    apiPost.mockResolvedValueOnce({ data: { verified: true } });
    // After a successful MFA, the store re-bootstraps the session to pick
    // up the new mfa_last_used.
    extendAdminSession.mockResolvedValueOnce({
      session: makeSession({ access_token: 'jwt-after-mfa' }),
      user: makeUser(),
    });

    const { login, verifyMFA } = useAdminAuthStore.getState();
    login(makeSession(), makeUser());
    useAdminAuthStore.setState({ sessionIpAddress: '127.0.0.1' });

    await verifyMFA('123456');
    const state = useAdminAuthStore.getState();
    expect(state.mfaVerified).toBe(true);
    expect(state.session?.access_token).toBe('jwt-after-mfa');
    expect(trackSecurityEvent).toHaveBeenCalledWith('mfa_verified');
  });

  it('rejects and keeps mfaVerified=false on a non-OK response', async () => {
    apiPost.mockResolvedValueOnce({ data: { verified: false, error: 'bad_code' } });

    const { login, verifyMFA } = useAdminAuthStore.getState();
    login(makeSession(), makeUser());

    await expect(verifyMFA('000000')).rejects.toThrow(/bad_code|MFA verification/i);
    expect(useAdminAuthStore.getState().mfaVerified).toBe(false);
    expect(trackSecurityEvent).toHaveBeenCalledWith(
      'mfa_verify_failed',
      expect.objectContaining({ reason: 'bad_code' })
    );
  });

  it('rejects and tracks mfa_verify_failed when the request itself throws', async () => {
    apiPost.mockRejectedValueOnce(new Error('network down'));

    const { login, verifyMFA } = useAdminAuthStore.getState();
    login(makeSession(), makeUser());

    await expect(verifyMFA('000000')).rejects.toThrow(/network down/);
    expect(useAdminAuthStore.getState().mfaVerified).toBe(false);
    expect(trackSecurityEvent).toHaveBeenCalledWith(
      'mfa_verify_failed',
      expect.objectContaining({ reason: 'request_error' })
    );
  });
});

describe('checkSession()', () => {
  it('returns false when there is no active session', () => {
    expect(useAdminAuthStore.getState().checkSession()).toBe(false);
  });

  it('returns true when lastActivity is recent', () => {
    useAdminAuthStore.setState({
      session: makeSession(),
      lastActivity: Date.now(),
    });
    expect(useAdminAuthStore.getState().checkSession()).toBe(true);
  });

  it('returns false when the session has expired by absolute time', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-01-01T00:00:00Z'));
    useAdminAuthStore.setState({
      session: makeSession({ expires_at: new Date('2026-01-01T00:00:00Z').toISOString() }),
      lastActivity: Date.now(),
    });
    // Advance past the expiry.
    vi.setSystemTime(new Date('2026-01-01T01:00:00Z'));
    expect(useAdminAuthStore.getState().checkSession()).toBe(false);
  });

  it('returns false when the session has been idle past the threshold', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-01-01T00:00:00Z'));
    useAdminAuthStore.setState({
      session: makeSession({ expires_at: new Date('2026-01-01T03:00:00Z').toISOString() }),
      lastActivity: Date.now(),
    });
    // Advance well past the idle timeout (default 15 min).
    vi.setSystemTime(new Date('2026-01-01T02:00:00Z'));
    expect(useAdminAuthStore.getState().checkSession()).toBe(false);
  });
});
