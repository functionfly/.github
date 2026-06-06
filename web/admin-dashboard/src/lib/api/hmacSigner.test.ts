/**
 * Tests for the server-issued HMAC signer.
 *
 * The signer must:
 *  1. never hold a shared secret in the browser (we don't import one)
 *  2. ask the backend for a short-lived signature per (method, path, body)
 *  3. reuse a cached signature until it's about to expire
 *  4. coalesce concurrent sign calls for the same intent
 *  5. clear the cache on logout
 */
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

// Hoisted mock so the import below picks it up.
const postRaw = vi.fn();
vi.mock('@/lib/api/adminClient', () => ({
  adminApiClient: {
    postRaw,
  },
}));

beforeEach(() => {
  postRaw.mockReset();
});

afterEach(async () => {
  // Reset module-level state between tests so cache + inflight don't leak.
  const mod = await import('./hmacSigner');
  mod.clearSignatureCache();
  vi.useRealTimers();
});

describe('hmacSigner.signRequest', () => {
  it('posts the request intent (method, path, body_hash) to /auth/sign-request', async () => {
    postRaw.mockResolvedValueOnce({
      data: {
        timestamp: '1700000000',
        signature: 'abc123',
        expires_at: Math.floor(Date.now() / 1000) + 300,
      },
      status: 200,
    });

    const { signRequest } = await import('./hmacSigner');
    const sig = await signRequest('POST', '/v1/admin/users', '{"name":"x"}');

    expect(sig).toEqual({ timestamp: '1700000000', signature: 'abc123' });
    expect(postRaw).toHaveBeenCalledTimes(1);
    const [path, body, config] = postRaw.mock.calls[0];
    expect(path).toBe('/auth/sign-request');
    expect(body).toMatchObject({
      method: 'POST',
      path: '/v1/admin/users',
    });
    expect(body.body_hash).toMatch(/^[0-9a-f]{64}$/); // sha256 hex
    expect(config._skipCsrf).toBe(true);
  });

  it('uppercases the method before posting', async () => {
    postRaw.mockResolvedValueOnce({
      data: { timestamp: '1', signature: 'x', expires_at: Math.floor(Date.now() / 1000) + 300 },
      status: 200,
    });
    const { signRequest } = await import('./hmacSigner');
    await signRequest('post', '/v1/admin/users', '');
    expect(postRaw.mock.calls[0][1].method).toBe('POST');
  });

  it('reuses a cached signature for the same intent', async () => {
    const farFuture = Math.floor(Date.now() / 1000) + 600;
    postRaw.mockResolvedValue({
      data: { timestamp: '1', signature: 'cached-sig', expires_at: farFuture },
      status: 200,
    });

    const { signRequest } = await import('./hmacSigner');
    const a = await signRequest('POST', '/v1/admin/users', '{}');
    const b = await signRequest('POST', '/v1/admin/users', '{}');
    const c = await signRequest('POST', '/v1/admin/users', '{}');
    expect(a.signature).toBe('cached-sig');
    expect(b.signature).toBe('cached-sig');
    expect(c.signature).toBe('cached-sig');
    expect(postRaw).toHaveBeenCalledTimes(1);
  });

  it('does not reuse a signature that is about to expire (within 5s)', async () => {
    // First call returns a signature that's about to expire.
    const almostExpired = Math.floor(Date.now() / 1000) + 2;
    postRaw.mockResolvedValueOnce({
      data: { timestamp: '1', signature: 'old', expires_at: almostExpired },
      status: 200,
    });
    // Second call returns a fresh one.
    const fresh = Math.floor(Date.now() / 1000) + 600;
    postRaw.mockResolvedValueOnce({
      data: { timestamp: '2', signature: 'new', expires_at: fresh },
      status: 200,
    });

    const { signRequest } = await import('./hmacSigner');
    const a = await signRequest('POST', '/v1/admin/users', '{}');
    const b = await signRequest('POST', '/v1/admin/users', '{}');
    expect(a.signature).toBe('old');
    expect(b.signature).toBe('new');
    expect(postRaw).toHaveBeenCalledTimes(2);
  });

  it('uses a different cache key for a different body', async () => {
    postRaw.mockResolvedValue({
      data: { timestamp: '1', signature: 'x', expires_at: Math.floor(Date.now() / 1000) + 600 },
      status: 200,
    });
    const { signRequest } = await import('./hmacSigner');
    await signRequest('POST', '/v1/admin/users', '{"a":1}');
    await signRequest('POST', '/v1/admin/users', '{"a":2}');
    expect(postRaw).toHaveBeenCalledTimes(2);
  });

  it('uses a different cache key for a different path', async () => {
    postRaw.mockResolvedValue({
      data: { timestamp: '1', signature: 'x', expires_at: Math.floor(Date.now() / 1000) + 600 },
      status: 200,
    });
    const { signRequest } = await import('./hmacSigner');
    await signRequest('POST', '/v1/admin/users', '{}');
    await signRequest('POST', '/v1/admin/tenants', '{}');
    expect(postRaw).toHaveBeenCalledTimes(2);
  });

  it('coalesces concurrent calls for the same intent into one server request', async () => {
    let resolvePost: (value: unknown) => void = () => {};
    const inFlight = new Promise((r) => {
      resolvePost = r;
    });
    postRaw.mockReturnValueOnce(inFlight as any);
    postRaw.mockResolvedValueOnce({
      data: { timestamp: '1', signature: 'coalesced', expires_at: Math.floor(Date.now() / 1000) + 600 },
      status: 200,
    });

    const { signRequest } = await import('./hmacSigner');
    const p1 = signRequest('POST', '/v1/admin/users', '{}');
    const p2 = signRequest('POST', '/v1/admin/users', '{}');
    const p3 = signRequest('POST', '/v1/admin/users', '{}');

    resolvePost({
      data: { timestamp: '1', signature: 'coalesced', expires_at: Math.floor(Date.now() / 1000) + 600 },
      status: 200,
    });

    const [s1, s2, s3] = await Promise.all([p1, p2, p3]);
    expect(s1.signature).toBe('coalesced');
    expect(s2.signature).toBe('coalesced');
    expect(s3.signature).toBe('coalesced');
    // Only the first call should have reached the API; the rest were
    // returned from the inflight promise.
    expect(postRaw).toHaveBeenCalledTimes(1);
  });

  it('rejects on a malformed sign-request response', async () => {
    postRaw.mockResolvedValueOnce({ data: { timestamp: '1' }, status: 200 });
    const { signRequest } = await import('./hmacSigner');
    await expect(signRequest('POST', '/v1/admin/users', '{}')).rejects.toThrow(
      /malformed/i
    );
  });

  it('clearSignatureCache forces a new sign on the next call', async () => {
    const farFuture = Math.floor(Date.now() / 1000) + 600;
    postRaw.mockResolvedValue({
      data: { timestamp: '1', signature: 'fresh', expires_at: farFuture },
      status: 200,
    });
    const { signRequest, clearSignatureCache } = await import('./hmacSigner');
    await signRequest('POST', '/v1/admin/users', '{}');
    clearSignatureCache();
    await signRequest('POST', '/v1/admin/users', '{}');
    expect(postRaw).toHaveBeenCalledTimes(2);
  });
});
