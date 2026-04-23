import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  deployKeysApi,
  type DeployKey,
  type DeployKeyCreateRequest,
} from './deploy-keys';

vi.mock('./client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  },
}));

import { apiClient } from './client';

const mockedApiClient = vi.mocked(apiClient);

describe('Deploy Keys API', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  describe('list', () => {
    it('should fetch deploy keys successfully', async () => {
      const mockResponse = {
        deploy_keys: [
          {
            id: 'key-1',
            name: 'Production Deploy',
            public_key: 'ssh-rsa AAAAB3...',
            fingerprint: 'SHA256:abc123...',
            created_at: '2024-01-15T10:00:00Z',
            created_by: 'user-1',
          },
        ] as DeployKey[],
        total_count: 1,
        page: 1,
        page_size: 20,
      };

      mockedApiClient.get.mockResolvedValueOnce(mockResponse);

      const result = await deployKeysApi.list();

      expect(mockedApiClient.get).toHaveBeenCalledWith('/v1/deploy-keys');
      expect(result.deploy_keys).toHaveLength(1);
      expect(result.total_count).toBe(1);
      expect(result.deploy_keys[0].name).toBe('Production Deploy');
    });

    it('should return empty list when no deploy keys', async () => {
      mockedApiClient.get.mockResolvedValueOnce({
        deploy_keys: [],
        total_count: 0,
        page: 1,
        page_size: 20,
      });

      const result = await deployKeysApi.list();

      expect(result.deploy_keys).toHaveLength(0);
      expect(result.total_count).toBe(0);
    });

    it('should handle network errors', async () => {
      mockedApiClient.get.mockRejectedValueOnce(new Error('Network error'));

      await expect(deployKeysApi.list()).rejects.toThrow('Network error');
    });

    it('should handle 404 as empty list (component-layer behavior)', async () => {
      const error = new Error('Not found');
      error.name = 'ApiError';
      mockedApiClient.get.mockRejectedValueOnce(error);

      await expect(deployKeysApi.list()).rejects.toThrow('Not found');
    });
  });

  describe('get', () => {
    it('should fetch a single deploy key by id', async () => {
      const mockKey: DeployKey = {
        id: 'key-1',
        name: 'CI Server',
        public_key: 'ssh-rsa AAAAB3...',
        fingerprint: 'SHA256:abc123...',
        created_at: '2024-01-15T10:00:00Z',
        expires_at: '2025-01-15T10:00:00Z',
        last_used_at: '2024-06-01T08:30:00Z',
        created_by: 'user-1',
      };

      mockedApiClient.get.mockResolvedValueOnce({ data: mockKey });

      const result = await deployKeysApi.get('key-1');

      expect(mockedApiClient.get).toHaveBeenCalledWith('/v1/deploy-keys/key-1');
      expect(result.data).toEqual(mockKey);
      expect(result.data.fingerprint).toBe('SHA256:abc123...');
    });

    it('should handle 404 not found', async () => {
      mockedApiClient.get.mockRejectedValueOnce(new Error('Not found'));

      await expect(deployKeysApi.get('non-existent')).rejects.toThrow('Not found');
    });
  });

  describe('create', () => {
    it('should create a deploy key successfully', async () => {
      const request: DeployKeyCreateRequest = {
        name: 'New Deploy Key',
        public_key: 'ssh-rsa AAAAB3...',
      };

      const mockCreatedKey: DeployKey = {
        id: 'key-new',
        name: 'New Deploy Key',
        public_key: 'ssh-rsa AAAAB3...',
        fingerprint: 'SHA256:new123...',
        created_at: '2024-06-01T10:00:00Z',
        created_by: 'user-1',
      };

      mockedApiClient.post.mockResolvedValueOnce({ data: mockCreatedKey });

      const result = await deployKeysApi.create(request);

      expect(mockedApiClient.post).toHaveBeenCalledWith('/v1/deploy-keys', request);
      expect(result.data.id).toBe('key-new');
      expect(result.data.name).toBe('New Deploy Key');
    });

    it('should create deploy key with only required fields', async () => {
      const request: DeployKeyCreateRequest = {
        name: 'Minimal Key',
        public_key: 'ssh-ed25519 AAAAC3...',
      };

      mockedApiClient.post.mockResolvedValueOnce({
        data: { ...request, id: 'key-min', fingerprint: 'SHA256:min...', created_at: new Date().toISOString(), created_by: 'user-1' },
      });

      const result = await deployKeysApi.create(request);

      expect(mockedApiClient.post).toHaveBeenCalledWith('/v1/deploy-keys', request);
      expect(result.data.name).toBe('Minimal Key');
    });

    it('should handle validation errors', async () => {
      const error = new Error('Validation failed');
      mockedApiClient.post.mockRejectedValueOnce(error);

      await expect(
        deployKeysApi.create({ name: '', public_key: '' })
      ).rejects.toThrow('Validation failed');
    });

    it('should handle 401 unauthorized', async () => {
      const error = new Error('Unauthorized');
      mockedApiClient.post.mockRejectedValueOnce(error);

      await expect(
        deployKeysApi.create({ name: 'Test', public_key: 'ssh-rsa AAA...' })
      ).rejects.toThrow('Unauthorized');
    });
  });

  describe('delete', () => {
    it('should delete a deploy key successfully', async () => {
      mockedApiClient.delete.mockResolvedValueOnce(undefined);

      await expect(deployKeysApi.delete('key-1')).resolves.toBeUndefined();
      expect(mockedApiClient.delete).toHaveBeenCalledWith('/v1/deploy-keys/key-1');
    });

    it('should handle 404 not found on delete', async () => {
      mockedApiClient.delete.mockRejectedValueOnce(new Error('Not found'));

      await expect(deployKeysApi.delete('non-existent')).rejects.toThrow('Not found');
    });

    it('should handle 403 forbidden on delete', async () => {
      const error = new Error('Forbidden');
      mockedApiClient.delete.mockRejectedValueOnce(error);

      await expect(deployKeysApi.delete('key-1')).rejects.toThrow('Forbidden');
    });
  });

  describe('verify', () => {
    it('should verify a deploy key successfully', async () => {
      mockedApiClient.post.mockResolvedValueOnce({
        valid: true,
        fingerprint: 'SHA256:verified...',
      });

      const result = await deployKeysApi.verify('key-1');

      expect(mockedApiClient.post).toHaveBeenCalledWith('/v1/deploy-keys/key-1/verify');
      expect(result.valid).toBe(true);
      expect(result.fingerprint).toBe('SHA256:verified...');
    });

    it('should handle verification failure', async () => {
      mockedApiClient.post.mockRejectedValueOnce(new Error('Verification failed'));

      await expect(deployKeysApi.verify('key-1')).rejects.toThrow('Verification failed');
    });

    it('should return invalid for expired key', async () => {
      mockedApiClient.post.mockResolvedValueOnce({
        valid: false,
        fingerprint: 'SHA256:expired...',
      });

      const result = await deployKeysApi.verify('expired-key');

      expect(result.valid).toBe(false);
    });
  });
});

describe('Deploy Keys API Response Shapes', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should return correct list response structure', async () => {
    mockedApiClient.get.mockResolvedValueOnce({
      deploy_keys: [],
      total_count: 0,
      page: 1,
      page_size: 20,
    });

    const result = await deployKeysApi.list();

    expect(result).toHaveProperty('deploy_keys');
    expect(result).toHaveProperty('total_count');
    expect(result).toHaveProperty('page');
    expect(result).toHaveProperty('page_size');
  });

  it('should return correct single key response structure', async () => {
    mockedApiClient.get.mockResolvedValueOnce({
      data: {
        id: 'key-1',
        name: 'Test Key',
        public_key: 'ssh-rsa AAA...',
        fingerprint: 'SHA256:test...',
        created_at: '2024-01-01T00:00:00Z',
        created_by: 'user-1',
      },
    });

    const result = await deployKeysApi.get('key-1');

    expect(result.data).toHaveProperty('id');
    expect(result.data).toHaveProperty('name');
    expect(result.data).toHaveProperty('public_key');
    expect(result.data).toHaveProperty('fingerprint');
    expect(result.data).toHaveProperty('created_at');
    expect(result.data).toHaveProperty('created_by');
  });

  it('should include optional fields when present', async () => {
    mockedApiClient.get.mockResolvedValueOnce({
      data: {
        id: 'key-1',
        name: 'Key with optional fields',
        public_key: 'ssh-rsa AAA...',
        fingerprint: 'SHA256:opt...',
        created_at: '2024-01-01T00:00:00Z',
        expires_at: '2025-01-01T00:00:00Z',
        last_used_at: '2024-06-01T00:00:00Z',
        created_by: 'user-1',
      },
    });

    const result = await deployKeysApi.get('key-1');

    expect(result.data.expires_at).toBe('2025-01-01T00:00:00Z');
    expect(result.data.last_used_at).toBe('2024-06-01T00:00:00Z');
  });
});

describe('Deploy Keys API Error Handling', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should handle network timeout', async () => {
    const error = new Error('Request timeout');
    error.name = 'TimeoutError';
    mockedApiClient.get.mockRejectedValueOnce(error);

    await expect(deployKeysApi.list()).rejects.toThrow('Request timeout');
  });

  it('should handle rate limiting', async () => {
    const error = new Error('Too many requests');
    error.name = 'ApiError';
    mockedApiClient.post.mockRejectedValueOnce(error);

    await expect(
      deployKeysApi.create({ name: 'Test', public_key: 'ssh-rsa AAA...' })
    ).rejects.toThrow('Too many requests');
  });

  it('should handle malformed JSON', async () => {
    const error = new Error('Invalid JSON');
    mockedApiClient.delete.mockRejectedValueOnce(error);

    await expect(deployKeysApi.delete('key-1')).rejects.toThrow('Invalid JSON');
  });
});