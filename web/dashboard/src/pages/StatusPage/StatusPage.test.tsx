import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import StatusPage from './index';
import * as useStatusHooks from '@/hooks/useStatus';
import * as useStatusWebSocketHooks from '@/hooks/useStatusWebSocket';

// Mock the hooks
vi.mock('@/hooks/useStatus', () => ({
  useStatus: vi.fn(),
  useUptimeMetrics: vi.fn(),
  useMaintenance: vi.fn(),
  useIncidents: vi.fn(),
  useStatusColors: vi.fn(),
  useSeverityColors: vi.fn(),
  useStatusLabels: vi.fn(),
}));

vi.mock('@/hooks/useStatusWebSocket', () => ({
  useRealtimeStatus: vi.fn(),
}));

vi.mock('@/stores/statusStore', () => ({
  useStatusStore: vi.fn(() => ({
    setPlatformStatus: vi.fn(),
    setComponents: vi.fn(),
    setProviders: vi.fn(),
  })),
}));

const mockedUseStatus = vi.mocked(useStatusHooks.useStatus);
const mockedUseUptimeMetrics = vi.mocked(useStatusHooks.useUptimeMetrics);
const mockedUseMaintenance = vi.mocked(useStatusHooks.useMaintenance);
const mockedUseIncidents = vi.mocked(useStatusHooks.useIncidents);
const mockedUseRealtimeStatus = vi.mocked(useStatusWebSocketHooks.useRealtimeStatus);

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

  return function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    );
  };
};

describe('StatusPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();

    // Default mock values
    mockedUseStatus.mockReturnValue({
      platformStatus: {
        status: 'operational',
        message: 'All systems operational',
        timestamp: new Date().toISOString(),
        components: [
          { id: 'api', name: 'API', category: 'core', status: 'operational', latency_ms: 45, uptime_percent: 99.9, last_checked: new Date().toISOString() },
          { id: 'database', name: 'Database', category: 'infrastructure', status: 'operational', latency_ms: 12, uptime_percent: 99.99, last_checked: new Date().toISOString() },
        ],
      },
      components: [
        { id: 'api', name: 'API', category: 'core', status: 'operational', latency_ms: 45, uptime_percent: 99.9, last_checked: new Date().toISOString() },
        { id: 'database', name: 'Database', category: 'infrastructure', status: 'operational', latency_ms: 12, uptime_percent: 99.99, last_checked: new Date().toISOString() },
      ],
      providers: [
        { id: 'fly', name: 'Fly.io', status: 'operational', regions: [], avg_latency_ms: 45, avg_success_rate: 99.9, last_updated: new Date().toISOString() },
      ],
      isLoading: false,
      isPlatformLoading: false,
      isComponentsLoading: false,
      isProvidersLoading: false,
      error: null,
      platformError: null,
      componentsError: null,
      providersError: null,
      refetch: vi.fn(),
      invalidateStatus: vi.fn(),
    });

    mockedUseUptimeMetrics.mockReturnValue({
      data: {
        period_days: 30,
        overall_uptime: 99.9,
        by_component: {},
        by_provider: {},
        daily_data: [],
      },
      isLoading: false,
      isSuccess: true,
      isError: false,
      error: null,
      isPending: false,
      isLoadingError: false,
      isRefetchError: false,
      isStale: false,
      isPlaceholderData: false,
      isInitialLoading: false,
      status: 'success',
      fetchStatus: 'idle',
      dataUpdatedAt: Date.now(),
      errorUpdatedAt: 0,
      failureCount: 0,
      failureReason: null,
      errorUpdateCount: 0,
      isFetched: true,
      isFetchedAfterMount: true,
      isFetching: false,
      isPaused: false,
      isRefetching: false,
      fetchNextPage: vi.fn(),
      fetchPreviousPage: vi.fn(),
      hasNextPage: false,
      hasPreviousPage: false,
      isFetchNextPageError: false,
      isFetchingNextPage: false,
      isFetchingPreviousPage: false,
      promise: Promise.resolve(),
      refetch: vi.fn(),
      remove: vi.fn(),
    } as any);

    mockedUseMaintenance.mockReturnValue({
      data: [],
      isLoading: false,
    } as any);

    mockedUseIncidents.mockReturnValue({
      data: {
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
        ],
        total: 1,
        limit: 20,
        offset: 0,
      },
      isLoading: false,
    } as any);

    mockedUseRealtimeStatus.mockReturnValue({
      isRealtime: true,
      isConnecting: false,
      wsError: null,
      reconnectAttempt: 0,
      reconnect: vi.fn(),
      disconnect: vi.fn(),
    });
  });

  it('renders status page with all sections', () => {
    render(<StatusPage />, { wrapper: createWrapper() });

    expect(screen.getByText('System Status')).toBeInTheDocument();
    expect(screen.getByText('Real-time platform health and incident tracking')).toBeInTheDocument();
  });

  it('displays live connection badge when connected', () => {
    render(<StatusPage />, { wrapper: createWrapper() });

    expect(screen.getByText('Live')).toBeInTheDocument();
  });

  it('displays loading state initially', () => {
    mockedUseStatus.mockReturnValue({
      platformStatus: null,
      components: null,
      providers: null,
      isLoading: true,
      isPlatformLoading: true,
      isComponentsLoading: true,
      isProvidersLoading: true,
      error: null,
      platformError: null,
      componentsError: null,
      providersError: null,
      refetch: vi.fn(),
      invalidateStatus: vi.fn(),
    });

    render(<StatusPage />, { wrapper: createWrapper() });

    // Should show loading skeletons or spinner
    expect(screen.getByText('System Status')).toBeInTheDocument();
  });

  it('displays operational status correctly', () => {
    render(<StatusPage />, { wrapper: createWrapper() });

    expect(screen.getByText('All Systems Operational')).toBeInTheDocument();
  });

  it('displays degraded status when components are degraded', () => {
    mockedUseStatus.mockReturnValue({
      platformStatus: {
        status: 'degraded',
        message: 'Some services are experiencing issues',
        timestamp: new Date().toISOString(),
        components: [
          { id: 'api', name: 'API', category: 'core', status: 'operational', latency_ms: 45, uptime_percent: 99.9, last_checked: new Date().toISOString() },
          { id: 'database', name: 'Database', category: 'infrastructure', status: 'degraded', latency_ms: 120, uptime_percent: 95.0, last_checked: new Date().toISOString() },
        ],
      },
      components: [
        { id: 'api', name: 'API', category: 'core', status: 'operational', latency_ms: 45, uptime_percent: 99.9, last_checked: new Date().toISOString() },
        { id: 'database', name: 'Database', category: 'infrastructure', status: 'degraded', latency_ms: 120, uptime_percent: 95.0, last_checked: new Date().toISOString() },
      ],
      providers: [],
      isLoading: false,
      isPlatformLoading: false,
      isComponentsLoading: false,
      isProvidersLoading: false,
      error: null,
      platformError: null,
      componentsError: null,
      providersError: null,
      refetch: vi.fn(),
      invalidateStatus: vi.fn(),
    });

    render(<StatusPage />, { wrapper: createWrapper() });

    expect(screen.getByText('Degraded Performance')).toBeInTheDocument();
  });

  it('displays incidents section', () => {
    render(<StatusPage />, { wrapper: createWrapper() });

    expect(screen.getByText('Test Incident')).toBeInTheDocument();
  });

  it('handles refresh action', () => {
    const refetchMock = vi.fn();
    mockedUseStatus.mockReturnValue({
      platformStatus: {
        status: 'operational',
        message: 'All systems operational',
        timestamp: new Date().toISOString(),
        components: [],
      },
      components: [],
      providers: [],
      isLoading: false,
      isPlatformLoading: false,
      isComponentsLoading: false,
      isProvidersLoading: false,
      error: null,
      platformError: null,
      componentsError: null,
      providersError: null,
      refetch: refetchMock,
      invalidateStatus: vi.fn(),
    });

    render(<StatusPage />, { wrapper: createWrapper() });

    // Find and click refresh button
    const refreshButton = screen.getByLabelText('Refresh status');
    fireEvent.click(refreshButton);

    expect(refetchMock).toHaveBeenCalled();
  });

  it('displays maintenance banner when upcoming maintenance exists', () => {
    mockedUseMaintenance.mockReturnValue({
      data: [
        {
          id: 'maint-1',
          title: 'Database Upgrade',
          description: 'Scheduled database maintenance',
          scheduled_start: new Date(Date.now() + 3600000).toISOString(),
          scheduled_end: new Date(Date.now() + 7200000).toISOString(),
          affected_components: ['database'],
          status: 'scheduled',
          created_at: new Date().toISOString(),
        },
      ],
      isLoading: false,
    } as any);

    render(<StatusPage />, { wrapper: createWrapper() });

    expect(screen.getByText('Scheduled Maintenance: Database Upgrade')).toBeInTheDocument();
  });

  it('displays connection status indicator', () => {
    mockedUseRealtimeStatus.mockReturnValue({
      isRealtime: false,
      isConnecting: true,
      wsError: null,
      reconnectAttempt: 1,
      reconnect: vi.fn(),
      disconnect: vi.fn(),
    });

    render(<StatusPage />, { wrapper: createWrapper() });

    expect(screen.getByText(/Connecting/)).toBeInTheDocument();
  });

  it('displays offline status on WebSocket error', () => {
    mockedUseRealtimeStatus.mockReturnValue({
      isRealtime: false,
      isConnecting: false,
      wsError: new Error('Connection failed'),
      reconnectAttempt: 0,
      reconnect: vi.fn(),
      disconnect: vi.fn(),
    });

    render(<StatusPage />, { wrapper: createWrapper() });

    expect(screen.getByText('Offline')).toBeInTheDocument();
  });

  it('displays polling status when WebSocket unavailable', () => {
    mockedUseRealtimeStatus.mockReturnValue({
      isRealtime: false,
      isConnecting: false,
      wsError: null,
      reconnectAttempt: 0,
      reconnect: vi.fn(),
      disconnect: vi.fn(),
    });

    render(<StatusPage />, { wrapper: createWrapper() });

    expect(screen.getByText('Polling')).toBeInTheDocument();
  });
});

describe('StatusPage Error Handling', () => {
  it('handles API errors gracefully', () => {
    mockedUseStatus.mockReturnValue({
      platformStatus: null,
      components: null,
      providers: null,
      isLoading: false,
      isPlatformLoading: false,
      isComponentsLoading: false,
      isProvidersLoading: false,
      error: new Error('Failed to fetch'),
      platformError: new Error('Failed to fetch platform status'),
      componentsError: null,
      providersError: null,
      refetch: vi.fn(),
      invalidateStatus: vi.fn(),
    });

    mockedUseIncidents.mockReturnValue({
      data: null,
      isLoading: false,
      error: new Error('Failed to fetch incidents'),
    } as any);

    render(<StatusPage />, { wrapper: createWrapper() });

    // Should still render without crashing
    expect(screen.getByText('System Status')).toBeInTheDocument();
  });
});

describe('StatusPage Accessibility', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('has proper heading hierarchy', () => {
    render(<StatusPage />, { wrapper: createWrapper() });

    const mainHeading = screen.getByRole('heading', { level: 1 });
    expect(mainHeading).toHaveTextContent('All Systems Operational');
  });

  it('has accessible status indicators', () => {
    render(<StatusPage />, { wrapper: createWrapper() });

    // Status section should have aria-label
    const statusSection = screen.getByLabelText('Platform Status');
    expect(statusSection).toBeInTheDocument();
  });

  it('has accessible refresh button', () => {
    render(<StatusPage />, { wrapper: createWrapper() });

    const refreshButton = screen.getByLabelText('Refresh status');
    expect(refreshButton).toBeInTheDocument();
  });
});
