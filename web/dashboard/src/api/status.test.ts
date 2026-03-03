import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  getPlatformStatus,
  getComponents,
  getProviders,
  getIncidents,
  getIncident,
  createIncident,
  updateIncident,
  getUptimeMetrics,
  getLatencyMetrics,
  getMaintenance,
  statusApi,
  type CreateIncidentRequest,
  type UpdateIncidentRequest,
  type GetIncidentsParams,
  type PlatformStatus,
  type ComponentHealth,
  type ProviderStatus,
  type Incident,
  type UptimeMetrics,
  type LatencyMetrics,
  type MaintenanceWindow,
} from './status';

// Mock the apiClient
vi.mock('./client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
  },
}));

import { apiClient } from './client';

const mockedApiClient = vi.mocked(apiClient);

describe('Status API', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  describe('getPlatformStatus', () => {
    it('should fetch platform status successfully', async () => {
      const mockResponse: PlatformStatus = {
        status: 'operational',
        message: 'All systems operational',
        timestamp: new Date().toISOString(),
        components: [
          {
            id: 'api',
            name: 'API',
            category: 'core',
            status: 'operational',
            latency_ms: 45,
            uptime_percent: 99.9,
            last_checked: new Date().toISOString(),
          },
        ],
      };

      mockedApiClient.get.mockResolvedValueOnce(mockResponse);

      const result = await getPlatformStatus();

      expect(mockedApiClient.get).toHaveBeenCalledWith('/v1/status');
      expect(result).toEqual(mockResponse);
      expect(result.status).toBe('operational');
      expect(result.components).toHaveLength(1);
    });

    it('should handle network errors', async () => {
      mockedApiClient.get.mockRejectedValueOnce(new Error('Network error'));

      await expect(getPlatformStatus()).rejects.toThrow('Network error');
      expect(mockedApiClient.get).toHaveBeenCalledWith('/v1/status');
    });

    it('should handle 500 server errors', async () => {
      const error = new Error('Internal Server Error');
      error.name = 'ApiError';
      mockedApiClient.get.mockRejectedValueOnce(error);

      await expect(getPlatformStatus()).rejects.toThrow();
    });
  });

  describe('getComponents', () => {
    it('should fetch component health status', async () => {
      const mockComponents: ComponentHealth[] = [
        {
          id: 'api',
          name: 'API Gateway',
          category: 'core',
          status: 'operational',
          latency_ms: 45,
          uptime_percent: 99.9,
          last_checked: new Date().toISOString(),
        },
        {
          id: 'database',
          name: 'PostgreSQL',
          category: 'infrastructure',
          status: 'operational',
          latency_ms: 12,
          uptime_percent: 99.99,
          last_checked: new Date().toISOString(),
        },
      ];

      mockedApiClient.get.mockResolvedValueOnce(mockComponents);

      const result = await getComponents();

      expect(mockedApiClient.get).toHaveBeenCalledWith('/v1/status/components');
      expect(result).toEqual(mockComponents);
      expect(result).toHaveLength(2);
    });

    it('should return empty array when no components', async () => {
      mockedApiClient.get.mockResolvedValueOnce([]);

      const result = await getComponents();

      expect(result).toEqual([]);
      expect(result).toHaveLength(0);
    });
  });

  describe('getProviders', () => {
    it('should fetch provider status', async () => {
      const mockProviders: ProviderStatus[] = [
        {
          id: 'fly',
          name: 'Fly.io',
          status: 'operational',
          regions: [
            {
              region: 'iad',
              status: 'operational',
              latency_ms: 45,
              success_rate: 99.9,
            },
          ],
          avg_latency_ms: 45,
          avg_success_rate: 99.9,
          last_updated: new Date().toISOString(),
        },
      ];

      mockedApiClient.get.mockResolvedValueOnce(mockProviders);

      const result = await getProviders();

      expect(mockedApiClient.get).toHaveBeenCalledWith('/v1/status/providers');
      expect(result).toEqual(mockProviders);
      expect(result[0].regions).toHaveLength(1);
    });
  });

  describe('getIncidents', () => {
    it('should fetch incidents with default params', async () => {
      const mockResponse = {
        incidents: [
          {
            id: 'inc-1',
            title: 'Test Incident',
            description: 'Test description',
            severity: 'high',
            status: 'investigating',
            affected_components: ['api'],
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
          },
        ] as Incident[],
        total: 1,
        limit: 20,
        offset: 0,
      };

      mockedApiClient.get.mockResolvedValueOnce(mockResponse);

      const result = await getIncidents();

      expect(mockedApiClient.get).toHaveBeenCalledWith('/v1/incidents');
      expect(result.incidents).toHaveLength(1);
      expect(result.total).toBe(1);
    });

    it('should build correct query string with filters', async () => {
      const params: GetIncidentsParams = {
        status: 'investigating',
        severity: 'critical',
        limit: 10,
        offset: 5,
      };

      mockedApiClient.get.mockResolvedValueOnce({
        incidents: [],
        total: 0,
        limit: 10,
        offset: 5,
      });

      await getIncidents(params);

      expect(mockedApiClient.get).toHaveBeenCalledWith(
        '/v1/incidents?status=investigating&severity=critical&limit=10&offset=5'
      );
    });

    it('should handle all filter values', async () => {
      const params: GetIncidentsParams = {
        status: 'all',
        severity: 'all',
      };

      mockedApiClient.get.mockResolvedValueOnce({
        incidents: [],
        total: 0,
        limit: 20,
        offset: 0,
      });

      await getIncidents(params);

      // When status and severity are 'all', they should not be included
      expect(mockedApiClient.get).toHaveBeenCalledWith('/v1/incidents');
    });

    it('should handle pagination params only', async () => {
      const params: GetIncidentsParams = {
        limit: 50,
        offset: 100,
      };

      mockedApiClient.get.mockResolvedValueOnce({
        incidents: [],
        total: 0,
        limit: 50,
        offset: 100,
      });

      await getIncidents(params);

      expect(mockedApiClient.get).toHaveBeenCalledWith(
        '/v1/incidents?limit=50&offset=100'
      );
    });
  });

  describe('getIncident', () => {
    it('should fetch single incident by ID', async () => {
      const mockIncident: Incident = {
        id: 'inc-1',
        title: 'Test Incident',
        description: 'Test description',
        severity: 'high',
        status: 'investigating',
        affected_components: ['api'],
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };

      mockedApiClient.get.mockResolvedValueOnce(mockIncident);

      const result = await getIncident('inc-1');

      expect(mockedApiClient.get).toHaveBeenCalledWith('/v1/incidents/inc-1');
      expect(result).toEqual(mockIncident);
    });

    it('should handle 404 not found', async () => {
      mockedApiClient.get.mockRejectedValueOnce(new Error('Not found'));

      await expect(getIncident('non-existent')).rejects.toThrow('Not found');
    });
  });

  describe('createIncident', () => {
    it('should create incident successfully', async () => {
      const request: CreateIncidentRequest = {
        title: 'New Incident',
        description: 'Incident description',
        severity: 'critical',
        status: 'investigating',
        affected_components: ['api', 'database'],
      };

      const mockResponse: Incident = {
        id: 'new-inc-1',
        ...request,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };

      mockedApiClient.post.mockResolvedValueOnce(mockResponse);

      const result = await createIncident(request);

      expect(mockedApiClient.post).toHaveBeenCalledWith('/v1/incidents', request);
      expect(result.id).toBe('new-inc-1');
      expect(result.title).toBe(request.title);
    });

    it('should handle validation errors', async () => {
      const request: CreateIncidentRequest = {
        title: '',
        description: '',
        severity: 'critical',
        status: 'investigating',
        affected_components: [],
      };

      const error = new Error('Validation failed');
      mockedApiClient.post.mockRejectedValueOnce(error);

      await expect(createIncident(request)).rejects.toThrow('Validation failed');
    });

    it('should handle 403 forbidden for non-admin', async () => {
      const request: CreateIncidentRequest = {
        title: 'New Incident',
        description: 'Description',
        severity: 'high',
        status: 'investigating',
        affected_components: [],
      };

      const error = new Error('Forbidden');
      mockedApiClient.post.mockRejectedValueOnce(error);

      await expect(createIncident(request)).rejects.toThrow('Forbidden');
    });
  });

  describe('updateIncident', () => {
    it('should update incident successfully', async () => {
      const request: UpdateIncidentRequest = {
        status: 'resolved',
        message: 'Issue has been resolved',
      };

      const mockResponse: Incident = {
        id: 'inc-1',
        title: 'Test Incident',
        description: 'Description',
        severity: 'high',
        status: 'resolved',
        affected_components: ['api'],
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        resolved_at: new Date().toISOString(),
      };

      mockedApiClient.patch.mockResolvedValueOnce(mockResponse);

      const result = await updateIncident('inc-1', request);

      expect(mockedApiClient.patch).toHaveBeenCalledWith('/v1/incidents/inc-1', request);
      expect(result.status).toBe('resolved');
    });

    it('should update partial fields', async () => {
      const request: UpdateIncidentRequest = {
        title: 'Updated Title',
      };

      const mockResponse: Incident = {
        id: 'inc-1',
        title: 'Updated Title',
        description: 'Description',
        severity: 'high',
        status: 'investigating',
        affected_components: ['api'],
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };

      mockedApiClient.patch.mockResolvedValueOnce(mockResponse);

      const result = await updateIncident('inc-1', request);

      expect(result.title).toBe('Updated Title');
    });
  });

  describe('getUptimeMetrics', () => {
    it('should fetch uptime metrics with default 30 days', async () => {
      const mockResponse: UptimeMetrics = {
        period_days: 30,
        overall_uptime: 99.9,
        by_component: {
          api: 99.95,
          database: 99.85,
        },
        by_provider: {
          fly: 99.9,
          vercel: 99.8,
        },
        daily_data: [
          {
            date: new Date().toISOString(),
            uptime: 99.9,
            incidents: 0,
          },
        ],
      };

      mockedApiClient.get.mockResolvedValueOnce(mockResponse);

      const result = await getUptimeMetrics();

      expect(mockedApiClient.get).toHaveBeenCalledWith('/v1/metrics/uptime?days=30');
      expect(result.period_days).toBe(30);
      expect(result.overall_uptime).toBe(99.9);
    });

    it('should fetch uptime metrics with custom period', async () => {
      const mockResponse: UptimeMetrics = {
        period_days: 90,
        overall_uptime: 99.85,
        by_component: {},
        by_provider: {},
        daily_data: [],
      };

      mockedApiClient.get.mockResolvedValueOnce(mockResponse);

      const result = await getUptimeMetrics(90);

      expect(mockedApiClient.get).toHaveBeenCalledWith('/v1/metrics/uptime?days=90');
      expect(result.period_days).toBe(90);
    });
  });

  describe('getLatencyMetrics', () => {
    it('should fetch latency metrics for provider', async () => {
      const mockResponse: LatencyMetrics = {
        provider: 'fly',
        time_range: '24h',
        avg_latency_ms: 45,
        p50_latency_ms: 40,
        p95_latency_ms: 65,
        p99_latency_ms: 85,
        data_points: [
          {
            timestamp: new Date().toISOString(),
            latency_ms: 45,
          },
        ],
      };

      mockedApiClient.get.mockResolvedValueOnce(mockResponse);

      const result = await getLatencyMetrics('fly');

      expect(mockedApiClient.get).toHaveBeenCalledWith('/v1/metrics/latency?provider=fly');
      expect(result.provider).toBe('fly');
      expect(result.avg_latency_ms).toBe(45);
    });

    it('should fetch latency metrics with region filter', async () => {
      const mockResponse: LatencyMetrics = {
        provider: 'fly',
        region: 'iad',
        time_range: '24h',
        avg_latency_ms: 45,
        p50_latency_ms: 40,
        p95_latency_ms: 65,
        p99_latency_ms: 85,
        data_points: [],
      };

      mockedApiClient.get.mockResolvedValueOnce(mockResponse);

      const result = await getLatencyMetrics('fly', 'iad');

      expect(mockedApiClient.get).toHaveBeenCalledWith(
        '/v1/metrics/latency?provider=fly&region=iad'
      );
      expect(result.region).toBe('iad');
    });
  });

  describe('getMaintenance', () => {
    it('should fetch maintenance windows', async () => {
      const mockResponse: MaintenanceWindow[] = [
        {
          id: 'maint-1',
          title: 'Scheduled Maintenance',
          description: 'Database upgrade',
          scheduled_start: new Date().toISOString(),
          scheduled_end: new Date(Date.now() + 3600000).toISOString(),
          affected_components: ['database'],
          status: 'scheduled',
          created_at: new Date().toISOString(),
        },
      ];

      mockedApiClient.get.mockResolvedValueOnce(mockResponse);

      const result = await getMaintenance();

      expect(mockedApiClient.get).toHaveBeenCalledWith('/v1/maintenance');
      expect(result).toHaveLength(1);
      expect(result[0].status).toBe('scheduled');
    });

    it('should return empty array when no maintenance', async () => {
      mockedApiClient.get.mockResolvedValueOnce([]);

      const result = await getMaintenance();

      expect(result).toEqual([]);
    });
  });

  describe('statusApi export', () => {
    it('should export all API functions', () => {
      expect(statusApi.getPlatformStatus).toBe(getPlatformStatus);
      expect(statusApi.getComponents).toBe(getComponents);
      expect(statusApi.getProviders).toBe(getProviders);
      expect(statusApi.getIncidents).toBe(getIncidents);
      expect(statusApi.getIncident).toBe(getIncident);
      expect(statusApi.createIncident).toBe(createIncident);
      expect(statusApi.updateIncident).toBe(updateIncident);
      expect(statusApi.getUptimeMetrics).toBe(getUptimeMetrics);
      expect(statusApi.getLatencyMetrics).toBe(getLatencyMetrics);
      expect(statusApi.getMaintenance).toBe(getMaintenance);
    });
  });
});

describe('Status API Error Handling', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should handle timeout errors', async () => {
    const error = new Error('Request timeout');
    error.name = 'TimeoutError';
    mockedApiClient.get.mockRejectedValueOnce(error);

    await expect(getPlatformStatus()).rejects.toThrow('Request timeout');
  });

  it('should handle rate limiting (429)', async () => {
    const error = new Error('Too many requests');
    error.name = 'ApiError';
    mockedApiClient.get.mockRejectedValueOnce(error);

    await expect(getIncidents({ limit: 1000 })).rejects.toThrow();
  });

  it('should handle authentication errors (401)', async () => {
    const error = new Error('Unauthorized');
    mockedApiClient.post.mockRejectedValueOnce(error);

    await expect(
      createIncident({
        title: 'Test',
        description: 'Test',
        severity: 'high',
        status: 'investigating',
        affected_components: [],
      })
    ).rejects.toThrow('Unauthorized');
  });

  it('should handle malformed JSON responses', async () => {
    const error = new Error('Invalid JSON');
    mockedApiClient.get.mockRejectedValueOnce(error);

    await expect(getPlatformStatus()).rejects.toThrow('Invalid JSON');
  });
});

describe('Status API Request Validation', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedApiClient.get.mockResolvedValue({});
  });

  it('should validate incident severity values', async () => {
    // Valid severities: critical, high, medium, low
    const validSeverities = ['critical', 'high', 'medium', 'low'] as const;

    for (const severity of validSeverities) {
      vi.clearAllMocks();
      mockedApiClient.get.mockResolvedValue({ incidents: [], total: 0, limit: 20, offset: 0 });

      await getIncidents({ severity });

      expect(mockedApiClient.get).toHaveBeenCalledWith(
        expect.stringContaining(`severity=${severity}`)
      );
    }
  });

  it('should validate incident status values', async () => {
    const validStatuses = ['investigating', 'identified', 'monitoring', 'resolved'] as const;

    for (const status of validStatuses) {
      vi.clearAllMocks();
      mockedApiClient.get.mockResolvedValue({ incidents: [], total: 0, limit: 20, offset: 0 });

      await getIncidents({ status });

      expect(mockedApiClient.get).toHaveBeenCalledWith(
        expect.stringContaining(`status=${status}`)
      );
    }
  });

  it('should validate uptime period values', async () => {
    const validPeriods: Array<30 | 90 | 365> = [30, 90, 365];

    for (const days of validPeriods) {
      vi.clearAllMocks();
      mockedApiClient.get.mockResolvedValue({
        period_days: days,
        overall_uptime: 99.9,
        by_component: {},
        by_provider: {},
        daily_data: [],
      });

      await getUptimeMetrics(days);

      expect(mockedApiClient.get).toHaveBeenCalledWith(`/v1/metrics/uptime?days=${days}`);
    }
  });
});
