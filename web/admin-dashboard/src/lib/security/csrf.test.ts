/**
 * Tests for the CSRF token helper.
 *
 * Pin the security-critical behavior:
 *  - the token is read in this order: in-memory → meta tag → cookie.
 *  - the in-memory token is only valid until its TTL expires.
 *  - refreshCsrfToken() hits /csrf and caches the response in memory.
 *  - the token is never persisted to localStorage / sessionStorage.
 *  - clearCsrfToken() resets the in-memory cache.
 */
// @vitest-environment happy-dom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

const getRaw = vi.fn();
const metaCsrfToken = vi.fn();

vi.mock('@/lib/api/adminClient', () => ({
  adminApiClient: {
    getRaw: (...args: unknown[]) => getRaw(...args),
  },
}));

// We import dynamically so the test can reset the in-memory cache between
// tests by calling clearCsrfToken() on the same module instance.
async function getModule() {
  return await import('./csrf');
}

beforeEach(() => {
  getRaw.mockReset();
  document.head.innerHTML = '';
  document.cookie.split(';').forEach((c) => {
    const name = c.split('=')[0].trim();
    if (name) {
      document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 GMT; path=/`;
    }
  });
  metaCsrfToken.mockReset();
});

afterEach(async () => {
  const mod = await getModule();
  mod.clearCsrfToken();
  vi.useRealTimers();
});

describe('getCsrfToken()', () => {
  it('returns the value of the <meta name="csrf-token"> tag when no in-memory token', async () => {
    document.head.innerHTML = '<meta name="csrf-token" content="meta-token-123" />';
    const mod = await getModule();
    expect(mod.getCsrfToken()).toBe('meta-token-123');
  });

  it('falls back to the csrf_token cookie when no meta tag', async () => {
    document.cookie = 'csrf_token=cookie-token-456; path=/';
    const mod = await getModule();
    expect(mod.getCsrfToken()).toBe('cookie-token-456');
  });

  it('falls back to the XSRF-TOKEN cookie when no csrf_token cookie', async () => {
    document.cookie = 'XSRF-TOKEN=xsrf-token-789; path=/';
    const mod = await getModule();
    expect(mod.getCsrfToken()).toBe('xsrf-token-789');
  });

  it('returns null when no token source is available', async () => {
    const mod = await getModule();
    expect(mod.getCsrfToken()).toBeNull();
  });

  it('caches the token in memory and does not re-read sources on subsequent calls', async () => {
    document.head.innerHTML = '<meta name="csrf-token" content="once" />';
    const mod = await getModule();
    expect(mod.getCsrfToken()).toBe('once');
    // Remove the meta tag — the cached token must still be returned.
    document.head.innerHTML = '';
    expect(mod.getCsrfToken()).toBe('once');
  });

  it('expires the in-memory token after the TTL', async () => {
    vi.useFakeTimers();
    document.cookie = 'csrf_token=shortlived; path=/';
    const mod = await getModule();
    expect(mod.getCsrfToken()).toBe('shortlived');
    // Advance past the 1-hour TTL.
    vi.setSystemTime(Date.now() + 3600_001);
    document.cookie = 'csrf_token=refreshed; path=/';
    expect(mod.getCsrfToken()).toBe('refreshed');
  });
});

describe('refreshCsrfToken()', () => {
  it('fetches /csrf with _skipCsrf and caches the returned token', async () => {
    getRaw.mockResolvedValueOnce({ data: { token: 'fresh-from-server' }, status: 200 });
    const mod = await getModule();
    const tok = await mod.refreshCsrfToken();
    expect(tok).toBe('fresh-from-server');
    expect(getRaw).toHaveBeenCalledTimes(1);
    const [path, config] = getRaw.mock.calls[0];
    expect(path).toBe('/csrf');
    expect(config._skipCsrf).toBe(true);
    expect(mod.getCsrfToken()).toBe('fresh-from-server');
  });

  it('returns null and does not cache on a malformed response', async () => {
    getRaw.mockResolvedValueOnce({ data: {}, status: 200 });
    const mod = await getModule();
    expect(await mod.refreshCsrfToken()).toBeNull();
    // Next getCsrfToken should also return null (no other source present).
    expect(mod.getCsrfToken()).toBeNull();
  });

  it('returns null and never throws when the network fails', async () => {
    getRaw.mockRejectedValueOnce(new Error('network down'));
    const mod = await getModule();
    await expect(mod.refreshCsrfToken()).resolves.toBeNull();
  });
});

describe('clearCsrfToken()', () => {
  it('forces the next getCsrfToken() to re-read the meta tag', async () => {
    document.head.innerHTML = '<meta name="csrf-token" content="first" />';
    const mod = await getModule();
    expect(mod.getCsrfToken()).toBe('first');
    document.head.innerHTML = '<meta name="csrf-token" content="second" />';
    mod.clearCsrfToken();
    expect(mod.getCsrfToken()).toBe('second');
  });
});

describe('isCsrfTokenExpired()', () => {
  it('returns true before any token has been set', async () => {
    const mod = await getModule();
    expect(mod.isCsrfTokenExpired()).toBe(true);
  });

  it('returns false while a cached token is fresh', async () => {
    document.cookie = 'csrf_token=fresh; path=/';
    const mod = await getModule();
    mod.getCsrfToken();
    expect(mod.isCsrfTokenExpired()).toBe(false);
  });

  it('returns true after the TTL elapses', async () => {
    vi.useFakeTimers();
    document.cookie = 'csrf_token=fresh; path=/';
    const mod = await getModule();
    mod.getCsrfToken();
    expect(mod.isCsrfTokenExpired()).toBe(false);
    vi.setSystemTime(Date.now() + 3600_001);
    expect(mod.isCsrfTokenExpired()).toBe(true);
  });
});
