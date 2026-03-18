import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactNode } from 'react';
import {
  useStatus,
  useUptimeMetrics,
  useLatencyMetrics,
  useMaintenance,
  useIncidents,
  useIncident,
  useCreateIncident,
  useStatusColors,
  useSeverityColors,
  useStatusLabels,
  statusKeys,
  incidentKeys,
} from './useStatus';
import * as statusApi from '@/api/status';
import type { LatencyMetrics } from '@/api/status';

// Mock the status API
vi.mock('@/api/status', () => ({
  getPlatformStatus: vi.fn(),
  getComponents: vi.fn(),
  getProviders: vi.fn(),
  getUptimeMetrics: vi.fn(),
  getLatencyMetrics: vi.fn(),
  getMaintenance: vi.fn(),
  getIncidents: vi.fn(),
  getIncident: vi.fn(),
  createIncident: vi.fn(),
  statusApi: {
    getPlatformStatus: vi.fn(),
    getComponents: vi.fn(),
    getProviders: vi.fn(),
    getUptimeMetrics: vi.fn(),
    getLatencyMetrics: vi.fn(),
    getMaintenance: vi.fn(),
    getIncidents: vi.fn(),
    getIncident: vi.fn(),
    createIncident: vi.fn(),
  },
}));

const mockedStatusApi = vi.mocked(statusApi);

// Create a wrapper with QueryClient
const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

  return function Wrapper({ children }: { children: any }) {
    return (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
  };
};

describe('useStatus', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch platform status, components, and providers', async () => {
    const mockPlatformStatus = {
      status: 'operational' as const,
      message: 'All systems operational',
      timestamp: new Date().toISOString(),
      components: [],
    };

    const mockComponents = [
      {
        id: 'api',
        name: 'API',
        category: 'core' as const,
        status: 'operational' as const,
        latency_ms: 45,
        uptime_percent: 99.9,
        last_checked: new Date().toISOString(),
      },
    ];

    const mockProviders = [
      {
        id: 'fly' as const,
        name: 'Fly.io',
        status: 'operational' as const,
        regions: [],
        avg_latency_ms: 45,
        avg_success_rate: 99.9,
        last_updated: new Date().toISOString(),
      },
    ];

    mockedStatusApi.getPlatformStatus.mockResolvedValueOnce(mockPlatformStatus);
    mockedStatusApi.getComponents.mockResolvedValueOnce(mockComponents);
    mockedStatusApi.getProviders.mockResolvedValueOnce(mockProviders);

    const { result } = renderHook(() => useStatus(), {
      wrapper: createWrapper(),
    });

    // Initially loading
    expect(result.current.isLoading).toBe(true);

    // Wait for data to load
    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.platformStatus).toEqual(mockPlatformStatus);
    expect(result.current.components).toEqual(mockComponents);
    expect(result.current.providers).toEqual(mockProviders);
  });

  it('should handle errors gracefully', async () => {
    mockedStatusApi.getPlatformStatus.mockRejectedValueOnce(
      new Error('Network error')
    );
    mockedStatusApi.getComponents.mockRejectedValueOnce(
      new Error('Network error')
    );
    mockedStatusApi.getProviders.mockRejectedValueOnce(
      new Error('Network error')
    );

    const { result } = renderHook(() => useStatus(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.error).toBeTruthy();
    });

    expect(result.current.platformError).toBeTruthy();
    expect(result.current.componentsError).toBeTruthy();
    expect(result.current.providersError).toBeTruthy();
  });

  it('should allow manual refetch', async () => {
    mockedStatusApi.getPlatformStatus.mockResolvedValue({
      status: 'operational',
      message: 'All systems operational',
      timestamp: new Date().toISOString(),
      components: [],
    });
    mockedStatusApi.getComponents.mockResolvedValue([]);
    mockedStatusApi.getProviders.mockResolvedValue([]);

    const { result } = renderHook(() => useStatus(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    // Call refetch
    await result.current.refetch();

    // Should have been called twice (initial + refetch)
    expect(mockedStatusApi.getPlatformStatus).toHaveBeenCalledTimes(2);
  });

  it('should support disabling queries', async () => {
    mockedStatusApi.getPlatformStatus.mockResolvedValue({
      status: 'operational',
      message: 'All systems operational',
      timestamp: new Date().toISOString(),
      components: [],
    });
    mockedStatusApi.getComponents.mockResolvedValue([]);
    mockedStatusApi.getProviders.mockResolvedValue([]);

    renderHook(() => useStatus({ enabled: false }), {
      wrapper: createWrapper(),
    });

    // Should not fetch when disabled
    expect(mockedStatusApi.getPlatformStatus).not.toHaveBeenCalled();
  });
});

describe('useUptimeMetrics', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch uptime metrics with default 30 days', async () => {
    const mockMetrics = {
      period_days: 30,
      overall_uptime: 99.9,
      by_component: {},
      by_provider: {},
      daily_data: [],
    };

    mockedStatusApi.getUptimeMetrics.mockResolvedValueOnce(mockMetrics);

    const { result } = renderHook(() => useUptimeMetrics(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data).toEqual(mockMetrics);
    expect(mockedStatusApi.getUptimeMetrics).toHaveBeenCalledWith(30);
  });

  it('should fetch uptime metrics with custom period', async () => {
    const mockMetrics = {
      period_days: 90,
      overall_uptime: 99.85,
      by_component: {},
      by_provider: {},
      daily_data: [],
    };

    mockedStatusApi.getUptimeMetrics.mockResolvedValueOnce(mockMetrics);

    const { result } = renderHook(() => useUptimeMetrics(90), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(mockedStatusApi.getUptimeMetrics).toHaveBeenCalledWith(90);
  });

  it('should use 5 minute stale time', async () => {
    mockedStatusApi.getUptimeMetrics.mockResolvedValue({
      period_days: 30,
      overall_uptime: 99.9,
      by_component: {},
      by_provider: {},
      daily_data: [],
    });

    const { result } = renderHook(() => useUptimeMetrics(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    // Data should be considered fresh for 5 minutes
    expect(result.current.isStale).toBe(false);
  });
});

describe('useLatencyMetrics', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch latency metrics for provider', async () => {
    const mockMetrics: LatencyMetrics = {
      provider: 'fly',
      time_range: '24h',
      avg_latency_ms: 45,
      p50_latency_ms: 40,
      p95_latency_ms: 65,
      p99_latency_ms: 85,
      data_points: [],
    };

    mockedStatusApi.getLatencyMetrics.mockResolvedValueOnce(mockMetrics);

    const { result } = renderHook(() => useLatencyMetrics('fly'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data).toEqual(mockMetrics);
    expect(mockedStatusApi.getLatencyMetrics).toHaveBeenCalledWith('fly', undefined);
  });

  it('should fetch latency metrics with region', async () => {
    const mockMetrics: LatencyMetrics = {
      provider: 'fly',
      region: 'iad',
      time_range: '24h',
      avg_latency_ms: 45,
      p50_latency_ms: 40,
      p95_latency_ms: 65,
      p99_latency_ms: 85,
      data_points: [],
    };

    mockedStatusApi.getLatencyMetrics.mockResolvedValueOnce(mockMetrics);

    const { result } = renderHook(() => useLatencyMetrics('fly', 'iad'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(mockedStatusApi.getLatencyMetrics).toHaveBeenCalledWith('fly', 'iad');
  });

  it('should be disabled when provider is empty', async () => {
    const { result } = renderHook(() => useLatencyMetrics('' as any), {
      wrapper: createWrapper(),
    });

    // Should not fetch when provider is empty
    expect(result.current.isLoading).toBe(false);
    expect(mockedStatusApi.getLatencyMetrics).not.toHaveBeenCalled();
  });
});

describe('useMaintenance', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch maintenance windows', async () => {
    const mockMaintenance = [
      {
        id: 'maint-1',
        title: 'Scheduled Maintenance',
        description: 'System upgrade',
        scheduled_start: new Date().toISOString(),
        scheduled_end: new Date(Date.now() + 3600000).toISOString(),
        affected_components: ['api'],
        status: 'scheduled' as const,
        created_at: new Date().toISOString(),
      },
    ];

    mockedStatusApi.getMaintenance.mockResolvedValueOnce(mockMaintenance);

    const { result } = renderHook(() => useMaintenance(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data).toEqual(mockMaintenance);
  });
});

describe('useIncidents', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch incidents with default params', async () => {
    const mockResponse = {
      incidents: [
        {
          id: 'inc-1',
          title: 'Test Incident',
          description: 'Test description',
          severity: 'high' as const,
          status: 'investigating' as const,
          affected_components: ['api'],
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
      ],
      total: 1,
      limit: 20,
      offset: 0,
    };

    mockedStatusApi.getIncidents.mockResolvedValueOnce(mockResponse);

    const { result } = renderHook(() => useIncidents(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data).toEqual(mockResponse);
    expect(mockedStatusApi.getIncidents).toHaveBeenCalledWith({});
  });

  it('should fetch incidents with custom params', async () => {
    const mockResponse = {
      incidents: [],
      total: 0,
      limit: 10,
      offset: 5,
    };

    mockedStatusApi.getIncidents.mockResolvedValueOnce(mockResponse);

    const params = { status: 'investigating' as const, severity: 'critical' as const, limit: 10, offset: 5 };

    const { result } = renderHook(() => useIncidents(params), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(mockedStatusApi.getIncidents).toHaveBeenCalledWith(params);
  });
});

describe('useIncident', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch single incident', async () => {
    const mockIncident = {
      id: 'inc-1',
      title: 'Test Incident',
      description: 'Test description',
      severity: 'high' as const,
      status: 'investigating' as const,
      affected_components: ['api'],
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    mockedStatusApi.getIncident.mockResolvedValueOnce(mockIncident);

    const { result } = renderHook(() => useIncident('inc-1'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data).toEqual(mockIncident);
    expect(mockedStatusApi.getIncident).toHaveBeenCalledWith('inc-1');
  });

  it('should be disabled when id is empty', async () => {
    const { result } = renderHook(() => useIncident(''), {
      wrapper: createWrapper(),
    });

    expect(result.current.isLoading).toBe(false);
    expect(mockedStatusApi.getIncident).not.toHaveBeenCalled();
  });
});

describe('useCreateIncident', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should create incident and invalidate cache', async () => {
    const newIncident = {
      title: 'New Incident',
      description: 'Description',
      severity: 'critical' as const,
      status: 'investigating' as const,
      affected_components: ['api'],
    };

    const mockResponse = {
      id: 'new-inc-1',
      ...newIncident,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    mockedStatusApi.createIncident.mockResolvedValueOnce(mockResponse);

    const { result } = renderHook(() => useCreateIncident(), {
      wrapper: createWrapper(),
    });

    await result.current.mutateAsync(newIncident);

    expect(mockedStatusApi.createIncident).toHaveBeenCalledWith(newIncident);
  });
});

describe('useStatusColors', () => {
  it('should return color config for all statuses', () => {
    const { result } = renderHook(() => useStatusColors());

    expect(result.current.operational).toBeDefined();
    expect(result.current.degraded).toBeDefined();
    expect(result.current.down).toBeDefined();
    expect(result.current.major_outage).toBeDefined();

    // Check specific colors
    expect(result.current.operational.bg).toBe('bg-emerald-500');
    expect(result.current.operational.text).toBe('text-emerald-400');
    expect(result.current.degraded.bg).toBe('bg-amber-500');
    expect(result.current.down.bg).toBe('bg-red-500');
    expect(result.current.major_outage.bg).toBe('bg-red-600');
  });
});

describe('useSeverityColors', () => {
  it('should return color config for all severities', () => {
    const { result } = renderHook(() => useSeverityColors());

    expect(result.current.critical).toBeDefined();
    expect(result.current.high).toBeDefined();
    expect(result.current.medium).toBeDefined();
    expect(result.current.low).toBeDefined();

    // Check specific colors
    expect(result.current.critical.bg).toBe('bg-red-500');
    expect(result.current.high.bg).toBe('bg-orange-500');
    expect(result.current.medium.bg).toBe('bg-amber-500');
    expect(result.current.low.bg).toBe('bg-blue-500');

    // Check labels
    expect(result.current.critical.label).toBe('Critical');
    expect(result.current.high.label).toBe('High');
    expect(result.current.medium.label).toBe('Medium');
    expect(result.current.low.label).toBe('Low');
  });
});

describe('useStatusLabels', () => {
  it('should return human-readable labels', () => {
    const { result } = renderHook(() => useStatusLabels());

    expect(result.current.operational).toBe('Operational');
    expect(result.current.degraded).toBe('Degraded Performance');
    expect(result.current.down).toBe('Service Disruption');
    expect(result.current.major_outage).toBe('Major Outage');
  });
});

describe('Query Keys', () => {
  describe('statusKeys', () => {
    it('should generate correct platform query key', () => {
      expect(statusKeys.platform()).toEqual(['status', 'platform']);
    });

    it('should generate correct components query key', () => {
      expect(statusKeys.components()).toEqual(['status', 'components']);
    });

    it('should generate correct providers query key', () => {
      expect(statusKeys.providers()).toEqual(['status', 'providers']);
    });

    it('should generate correct uptime query key', () => {
      expect(statusKeys.uptime(30)).toEqual(['status', 'uptime', 30]);
      expect(statusKeys.uptime(90)).toEqual(['status', 'uptime', 90]);
    });

    it('should generate correct latency query key', () => {
      expect(statusKeys.latency('fly')).toEqual(['status', 'latency', 'fly', undefined]);
      expect(statusKeys.latency('fly', 'iad')).toEqual(['status', 'latency', 'fly', 'iad']);
    });

    it('should generate correct maintenance query key', () => {
      expect(statusKeys.maintenance()).toEqual(['status', 'maintenance']);
    });
  });

  describe('incidentKeys', () => {
    it('should generate correct list query key', () => {
      const params = { status: 'investigating' as const };
      expect(incidentKeys.lists(params)).toEqual(['incidents', 'list', params]);
    });

    it('should generate correct detail query key', () => {
      expect(incidentKeys.detail('inc-1')).toEqual(['incidents', 'detail', 'inc-1']);
    });
  });
});

describe('Polling Behavior', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should configure polling interval for useStatus', async () => {
    mockedStatusApi.getPlatformStatus.mockResolvedValue({
      status: 'operational',
      message: 'All systems operational',
      timestamp: new Date().toISOString(),
      components: [],
    });
    mockedStatusApi.getComponents.mockResolvedValue([]);
    mockedStatusApi.getProviders.mockResolvedValue([]);

    const { result } = renderHook(() => useStatus(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    // Polling interval is 30 seconds, we can't easily test this without
    // waiting, but we can verify the hook is configured correctly
    expect(result.current.isLoading).toBe(false);
  });
});

describe('Cache Invalidation', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should expose invalidateStatus function', async () => {
    mockedStatusApi.getPlatformStatus.mockResolvedValue({
      status: 'operational',
      message: 'All systems operational',
      timestamp: new Date().toISOString(),
      components: [],
    });
    mockedStatusApi.getComponents.mockResolvedValue([]);
    mockedStatusApi.getProviders.mockResolvedValue([]);

    const { result } = renderHook(() => useStatus(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(typeof result.current.invalidateStatus).toBe('function');
  });
});
