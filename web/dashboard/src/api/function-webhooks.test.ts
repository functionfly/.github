import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  functionWebhooksApi,
  type FunctionWebhook,
  type WebhookCreateRequest,
  type WebhookUpdateRequest,
} from './function-webhooks';

vi.mock('./client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
  },
}));

import { apiClient } from './client';

const mockedApiClient = vi.mocked(apiClient);

describe('Function Webhooks API', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  describe('list', () => {
    it('should fetch webhooks successfully', async () => {
      const mockResponse = {
        subscriptions: [
          {
            id: 'wh-1',
            url: 'https://example.com/webhook',
            secret: 'whsec_abc123',
            event_types: ['function.executed', 'function.failed'],
            active: true,
            created_at: '2024-01-15T10:00:00Z',
            updated_at: '2024-01-15T10:00:00Z',
            created_by: 'user-1',
          },
        ] as FunctionWebhook[],
        total_count: 1,
        page: 1,
        page_size: 20,
      };

      mockedApiClient.get.mockResolvedValueOnce(mockResponse);

      const result = await functionWebhooksApi.list();

      expect(mockedApiClient.get).toHaveBeenCalledWith('/v1/function-webhooks');
      expect(result.subscriptions).toHaveLength(1);
      expect(result.total_count).toBe(1);
      expect(result.subscriptions[0].url).toBe('https://example.com/webhook');
    });

    it('should return empty list when no webhooks', async () => {
      mockedApiClient.get.mockResolvedValueOnce({
        subscriptions: [],
        total_count: 0,
        page: 1,
        page_size: 20,
      });

      const result = await functionWebhooksApi.list();

      expect(result.subscriptions).toHaveLength(0);
      expect(result.total_count).toBe(0);
    });

    it('should handle network errors', async () => {
      mockedApiClient.get.mockRejectedValueOnce(new Error('Network error'));

      await expect(functionWebhooksApi.list()).rejects.toThrow('Network error');
    });

    it('should handle 404 as empty list (component-layer behavior)', async () => {
      const error = new Error('Not found');
      error.name = 'ApiError';
      mockedApiClient.get.mockRejectedValueOnce(error);

      await expect(functionWebhooksApi.list()).rejects.toThrow('Not found');
    });
  });

  describe('get', () => {
    it('should fetch a single webhook by id', async () => {
      const mockWebhook: FunctionWebhook = {
        id: 'wh-1',
        url: 'https://example.com/webhook',
        secret: 'whsec_abc123',
        event_types: ['function.executed'],
        active: true,
        created_at: '2024-01-15T10:00:00Z',
        updated_at: '2024-01-15T10:00:00Z',
        created_by: 'user-1',
      };

      mockedApiClient.get.mockResolvedValueOnce({ data: mockWebhook });

      const result = await functionWebhooksApi.get('wh-1');

      expect(mockedApiClient.get).toHaveBeenCalledWith('/v1/function-webhooks/wh-1');
      expect(result.data).toEqual(mockWebhook);
      expect(result.data.event_types).toContain('function.executed');
    });

    it('should handle 404 not found', async () => {
      mockedApiClient.get.mockRejectedValueOnce(new Error('Not found'));

      await expect(functionWebhooksApi.get('non-existent')).rejects.toThrow('Not found');
    });
  });

  describe('create', () => {
    it('should create a webhook successfully', async () => {
      const request: WebhookCreateRequest = {
        url: 'https://example.com/webhook',
        event_types: ['function.executed', 'function.failed'],
        secret: 'my-secret',
      };

      const mockCreatedWebhook: FunctionWebhook = {
        id: 'wh-new',
        url: 'https://example.com/webhook',
        secret: 'my-secret',
        event_types: ['function.executed', 'function.failed'],
        active: true,
        created_at: '2024-06-01T10:00:00Z',
        updated_at: '2024-06-01T10:00:00Z',
        created_by: 'user-1',
      };

      mockedApiClient.post.mockResolvedValueOnce({ data: mockCreatedWebhook });

      const result = await functionWebhooksApi.create(request);

      expect(mockedApiClient.post).toHaveBeenCalledWith('/v1/function-webhooks', request);
      expect(result.data.id).toBe('wh-new');
      expect(result.data.url).toBe('https://example.com/webhook');
    });

    it('should create webhook without optional secret (auto-generate)', async () => {
      const request: WebhookCreateRequest = {
        url: 'https://example.com/webhook',
        event_types: ['function.executed'],
      };

      mockedApiClient.post.mockResolvedValueOnce({
        data: { ...request, id: 'wh-no-secret', active: true, created_at: new Date().toISOString(), updated_at: new Date().toISOString(), created_by: 'user-1', secret: 'auto-generated' },
      });

      const result = await functionWebhooksApi.create(request);

      expect(mockedApiClient.post).toHaveBeenCalledWith('/v1/function-webhooks', request);
      expect(result.data.secret).toBe('auto-generated');
    });

    it('should create webhook with function_id', async () => {
      const request: WebhookCreateRequest = {
        url: 'https://example.com/webhook',
        event_types: ['function.executed'],
        function_id: 'func-123',
      };

      mockedApiClient.post.mockResolvedValueOnce({
        data: { ...request, id: 'wh-func', active: true, created_at: new Date().toISOString(), updated_at: new Date().toISOString(), created_by: 'user-1' },
      });

      const result = await functionWebhooksApi.create(request);

      expect(mockedApiClient.post).toHaveBeenCalledWith('/v1/function-webhooks', request);
      expect(result.data.function_id).toBe('func-123');
    });

    it('should handle invalid URL', async () => {
      const error = new Error('Invalid URL');
      mockedApiClient.post.mockRejectedValueOnce(error);

      await expect(
        functionWebhooksApi.create({ url: 'not-a-url', event_types: ['function.executed'] })
      ).rejects.toThrow('Invalid URL');
    });

    it('should handle 401 unauthorized', async () => {
      const error = new Error('Unauthorized');
      mockedApiClient.post.mockRejectedValueOnce(error);

      await expect(
        functionWebhooksApi.create({ url: 'https://example.com/wh', event_types: ['function.executed'] })
      ).rejects.toThrow('Unauthorized');
    });
  });

  describe('update', () => {
    it('should update webhook successfully', async () => {
      const request: WebhookUpdateRequest = {
        url: 'https://new-example.com/webhook',
        active: false,
      };

      const mockUpdatedWebhook: FunctionWebhook = {
        id: 'wh-1',
        url: 'https://new-example.com/webhook',
        secret: 'whsec_abc123',
        event_types: ['function.executed'],
        active: false,
        created_at: '2024-01-15T10:00:00Z',
        updated_at: '2024-06-01T12:00:00Z',
        created_by: 'user-1',
      };

      mockedApiClient.patch.mockResolvedValueOnce({ data: mockUpdatedWebhook });

      const result = await functionWebhooksApi.update('wh-1', request);

      expect(mockedApiClient.patch).toHaveBeenCalledWith('/v1/function-webhooks/wh-1', request);
      expect(result.data.active).toBe(false);
      expect(result.data.url).toBe('https://new-example.com/webhook');
    });

    it('should update only event_types', async () => {
      const request: WebhookUpdateRequest = {
        event_types: ['function.completed'],
      };

      mockedApiClient.patch.mockResolvedValueOnce({
        data: { id: 'wh-1', url: 'https://example.com/wh', secret: 'whsec_abc', event_types: ['function.completed'], active: true, created_at: '2024-01-01', updated_at: '2024-06-01', created_by: 'user-1' },
      });

      const result = await functionWebhooksApi.update('wh-1', request);

      expect(result.data.event_types).toContain('function.completed');
    });
  });

  describe('delete', () => {
    it('should delete a webhook successfully', async () => {
      mockedApiClient.delete.mockResolvedValueOnce(undefined);

      await expect(functionWebhooksApi.delete('wh-1')).resolves.toBeUndefined();
      expect(mockedApiClient.delete).toHaveBeenCalledWith('/v1/function-webhooks/wh-1');
    });

    it('should handle 404 not found on delete', async () => {
      mockedApiClient.delete.mockRejectedValueOnce(new Error('Not found'));

      await expect(functionWebhooksApi.delete('non-existent')).rejects.toThrow('Not found');
    });
  });

  describe('test', () => {
    it('should test a webhook successfully', async () => {
      mockedApiClient.post.mockResolvedValueOnce({
        success: true,
        delivery_id: 'del-123',
      });

      const result = await functionWebhooksApi.test('wh-1');

      expect(mockedApiClient.post).toHaveBeenCalledWith('/v1/function-webhooks/wh-1/test');
      expect(result.success).toBe(true);
      expect(result.delivery_id).toBe('del-123');
    });

    it('should handle test failure', async () => {
      mockedApiClient.post.mockResolvedValueOnce({
        success: false,
      });

      const result = await functionWebhooksApi.test('wh-1');

      expect(result.success).toBe(false);
    });

    it('should handle webhook test network error', async () => {
      mockedApiClient.post.mockRejectedValueOnce(new Error('Network error'));

      await expect(functionWebhooksApi.test('wh-1')).rejects.toThrow('Network error');
    });
  });

  describe('listDeliveries', () => {
    it('should fetch deliveries for a webhook', async () => {
      const mockResponse = {
        deliveries: [
          {
            id: 'del-1',
            webhook_id: 'wh-1',
            event_type: 'function.executed',
            payload: { function_id: 'func-1', result: 'success' },
            response_status: 200,
            response_body: '{"status":"ok"}',
            attempted_at: '2024-06-01T10:00:00Z',
            delivered_at: '2024-06-01T10:00:01Z',
          },
        ],
        total_count: 1,
        page: 1,
        page_size: 20,
      };

      mockedApiClient.get.mockResolvedValueOnce(mockResponse);

      const result = await functionWebhooksApi.listDeliveries('wh-1');

      expect(mockedApiClient.get).toHaveBeenCalledWith('/v1/function-webhooks/wh-1/deliveries');
      expect(result.deliveries).toHaveLength(1);
      expect(result.deliveries[0].response_status).toBe(200);
    });

    it('should return empty deliveries when none exist', async () => {
      mockedApiClient.get.mockResolvedValueOnce({
        deliveries: [],
        total_count: 0,
        page: 1,
        page_size: 20,
      });

      const result = await functionWebhooksApi.listDeliveries('wh-new');

      expect(result.deliveries).toHaveLength(0);
    });

    it('should handle failed delivery', async () => {
      const mockResponse = {
        deliveries: [
          {
            id: 'del-fail',
            webhook_id: 'wh-1',
            event_type: 'function.executed',
            payload: { function_id: 'func-1' },
            error: 'Connection timeout',
            attempted_at: '2024-06-01T10:00:00Z',
          },
        ],
        total_count: 1,
        page: 1,
        page_size: 20,
      };

      mockedApiClient.get.mockResolvedValueOnce(mockResponse);

      const result = await functionWebhooksApi.listDeliveries('wh-1');

      expect(result.deliveries[0].error).toBe('Connection timeout');
    });
  });
});

describe('Function Webhooks API Response Shapes', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should return correct webhook structure', async () => {
    mockedApiClient.get.mockResolvedValueOnce({
      subscriptions: [],
      total_count: 0,
      page: 1,
      page_size: 20,
    });

    const result = await functionWebhooksApi.list();

    expect(result).toHaveProperty('subscriptions');
    expect(result).toHaveProperty('total_count');
    expect(result).toHaveProperty('page');
    expect(result).toHaveProperty('page_size');
  });

  it('should return correct single webhook data structure', async () => {
    mockedApiClient.get.mockResolvedValueOnce({
      data: {
        id: 'wh-1',
        url: 'https://example.com/wh',
        secret: 'secret',
        event_types: ['function.executed'],
        active: true,
        created_at: '2024-01-01',
        updated_at: '2024-01-01',
        created_by: 'user-1',
      },
    });

    const result = await functionWebhooksApi.get('wh-1');

    expect(result.data).toHaveProperty('id');
    expect(result.data).toHaveProperty('url');
    expect(result.data).toHaveProperty('secret');
    expect(result.data).toHaveProperty('event_types');
    expect(result.data).toHaveProperty('active');
  });

  it('should include optional function_id when present', async () => {
    mockedApiClient.get.mockResolvedValueOnce({
      data: {
        id: 'wh-1',
        function_id: 'func-123',
        url: 'https://example.com/wh',
        secret: 'secret',
        event_types: ['function.executed'],
        active: true,
        created_at: '2024-01-01',
        updated_at: '2024-01-01',
        created_by: 'user-1',
      },
    });

    const result = await functionWebhooksApi.get('wh-1');

    expect(result.data.function_id).toBe('func-123');
  });
});

describe('Function Webhooks API Error Handling', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should handle network timeout', async () => {
    const error = new Error('Request timeout');
    error.name = 'TimeoutError';
    mockedApiClient.get.mockRejectedValueOnce(error);

    await expect(functionWebhooksApi.list()).rejects.toThrow('Request timeout');
  });

  it('should handle rate limiting', async () => {
    const error = new Error('Too many requests');
    error.name = 'ApiError';
    mockedApiClient.post.mockRejectedValueOnce(error);

    await expect(
      functionWebhooksApi.create({ url: 'https://example.com/wh', event_types: ['function.executed'] })
    ).rejects.toThrow('Too many requests');
  });

  it('should handle malformed JSON', async () => {
    const error = new Error('Invalid JSON');
    mockedApiClient.delete.mockRejectedValueOnce(error);

    await expect(functionWebhooksApi.delete('wh-1')).rejects.toThrow('Invalid JSON');
  });
});