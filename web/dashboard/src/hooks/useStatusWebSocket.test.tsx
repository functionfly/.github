import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useRealtimeStatus, useStatusWebSocket } from './useStatusWebSocket';

// Mock the toast notifications
vi.mock('sonner', () => ({
  toast: {
    warning: vi.fn(),
    info: vi.fn(),
  },
}));

import { toast } from 'sonner';

const mockedToast = vi.mocked(toast);

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
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
  };
};

describe('useStatusWebSocket', () => {
  let mockWebSocket: any;
  let mockSend: any;
  let mockClose: any;

  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();

    // Create mock WebSocket
    mockSend = vi.fn();
    mockClose = vi.fn();
    mockWebSocket = {
      send: mockSend,
      close: mockClose,
      readyState: WebSocket.CONNECTING,
      onopen: null as Function | null,
      onmessage: null as Function | null,
      onerror: null as Function | null,
      onclose: null as Function | null,
    };

    // Mock WebSocket constructor
    global.WebSocket = vi.fn(() => mockWebSocket) as any;

    // Mock window.location
    Object.defineProperty(window, 'location', {
      value: {
        protocol: 'https:',
        host: 'test.functionfly.io',
      },
      writable: true,
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  it('should create WebSocket connection on mount when enabled', () => {
    renderHook(() => useStatusWebSocket({ enabled: true }), {
      wrapper: createWrapper(),
    });

    expect(global.WebSocket).toHaveBeenCalledWith('wss://test.functionfly.io/ws/v1/status');
  });

  it('should not connect when disabled', () => {
    renderHook(() => useStatusWebSocket({ enabled: false }), {
      wrapper: createWrapper(),
    });

    expect(global.WebSocket).not.toHaveBeenCalled();
  });

  it('should handle successful connection', () => {
    const { result } = renderHook(() => useStatusWebSocket({ enabled: true }), {
      wrapper: createWrapper(),
    });

    // Initially connecting
    expect(result.current.isConnecting).toBe(true);
    expect(result.current.isConnected).toBe(false);

    // Simulate connection open
    act(() => {
      mockWebSocket.readyState = WebSocket.OPEN;
      mockWebSocket.onopen?.();
    });

    expect(result.current.isConnected).toBe(true);
    expect(result.current.isConnecting).toBe(false);
    expect(result.current.error).toBeNull();
  });

  it('should handle connection error', () => {
    const { result } = renderHook(() => useStatusWebSocket({ enabled: true }), {
      wrapper: createWrapper(),
    });

    act(() => {
      mockWebSocket.onerror?.(new Event('error'));
    });

    expect(result.current.error).toBeTruthy();
    expect(result.current.isConnected).toBe(false);
  });

  it('should handle connection close and reconnect', () => {
    const { result } = renderHook(() => useStatusWebSocket({ enabled: true }), {
      wrapper: createWrapper(),
    });

    // Connect first
    act(() => {
      mockWebSocket.readyState = WebSocket.OPEN;
      mockWebSocket.onopen?.();
    });

    expect(result.current.isConnected).toBe(true);

    // Close connection
    act(() => {
      mockWebSocket.onclose?.({ code: 1000, reason: 'Normal closure' });
    });

    expect(result.current.isConnected).toBe(false);

    // Fast-forward past reconnection delay
    act(() => {
      vi.advanceTimersByTime(2000);
    });

    // Should attempt reconnection
    expect(global.WebSocket).toHaveBeenCalledTimes(2);
  });

  it('should handle status_update message', () => {
    const onStatusUpdate = vi.fn();
    const queryClient = new QueryClient();

    const wrapper = ({ children }: { children: any }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );

    renderHook(() => useStatusWebSocket({ enabled: true, onStatusUpdate }), { wrapper });

    // Connect
    act(() => {
      mockWebSocket.readyState = WebSocket.OPEN;
      mockWebSocket.onopen?.();
    });

    // Simulate status update message
    const statusUpdate = {
      status: 'operational',
      message: 'All systems operational',
      timestamp: new Date().toISOString(),
      components: [],
    };

    act(() => {
      mockWebSocket.onmessage?.({
        data: JSON.stringify({
          type: 'status_update',
          timestamp: new Date().toISOString(),
          data: statusUpdate,
        }),
      } as MessageEvent);
    });

    expect(onStatusUpdate).toHaveBeenCalledWith(statusUpdate);
  });

  it('should handle incident_update message', () => {
    const onIncidentUpdate = vi.fn();

    const { result } = renderHook(
      () => useStatusWebSocket({ enabled: true, onIncidentUpdate, showNotifications: true }),
      { wrapper: createWrapper() }
    );

    // Connect
    act(() => {
      mockWebSocket.readyState = WebSocket.OPEN;
      mockWebSocket.onopen?.();
    });

    // Simulate incident update message
    const incidentUpdate = {
      id: 'inc-1',
      title: 'New Incident',
      description: 'System is experiencing issues',
      severity: 'high',
      status: 'investigating',
      affected_components: ['api'],
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    act(() => {
      mockWebSocket.onmessage?.({
        data: JSON.stringify({
          type: 'incident_update',
          timestamp: new Date().toISOString(),
          data: incidentUpdate,
        }),
      } as MessageEvent);
    });

    expect(onIncidentUpdate).toHaveBeenCalledWith(incidentUpdate);
    expect(mockedToast.warning).toHaveBeenCalled();
  });

  it('should handle maintenance_update message', () => {
    const onMaintenanceUpdate = vi.fn();

    const { result } = renderHook(
      () => useStatusWebSocket({ enabled: true, onMaintenanceUpdate, showNotifications: true }),
      { wrapper: createWrapper() }
    );

    // Connect
    act(() => {
      mockWebSocket.readyState = WebSocket.OPEN;
      mockWebSocket.onopen?.();
    });

    // Simulate maintenance update message
    const maintenanceUpdate = {
      id: 'maint-1',
      title: 'Scheduled Maintenance',
      description: 'Database upgrade',
      scheduled_start: new Date(Date.now() + 86400000).toISOString(),
      scheduled_end: new Date(Date.now() + 90000000).toISOString(),
      affected_components: ['database'],
      status: 'scheduled',
      created_at: new Date().toISOString(),
    };

    act(() => {
      mockWebSocket.onmessage?.({
        data: JSON.stringify({
          type: 'maintenance_update',
          timestamp: new Date().toISOString(),
          data: maintenanceUpdate,
        }),
      } as MessageEvent);
    });

    expect(onMaintenanceUpdate).toHaveBeenCalledWith(maintenanceUpdate);
    expect(mockedToast.info).toHaveBeenCalled();
  });

  it('should handle ping message', () => {
    const { result } = renderHook(() => useStatusWebSocket({ enabled: true }), {
      wrapper: createWrapper(),
    });

    // Connect
    act(() => {
      mockWebSocket.readyState = WebSocket.OPEN;
      mockWebSocket.onopen?.();
    });

    act(() => {
      mockWebSocket.onmessage?.({
        data: JSON.stringify({
          type: 'ping',
          timestamp: new Date().toISOString(),
        }),
      } as MessageEvent);
    });

    expect(mockSend).toHaveBeenCalledWith(JSON.stringify({ type: 'pong' }));
  });

  it('should handle invalid JSON messages gracefully', () => {
    const { result } = renderHook(() => useStatusWebSocket({ enabled: true }), {
      wrapper: createWrapper(),
    });

    // Connect
    act(() => {
      mockWebSocket.readyState = WebSocket.OPEN;
      mockWebSocket.onopen?.();
    });

    // Should not throw
    act(() => {
      mockWebSocket.onmessage?.({
        data: 'invalid json',
      } as MessageEvent);
    });

    expect(result.current.isConnected).toBe(true);
  });

  it('should disconnect on unmount', () => {
    const { unmount } = renderHook(() => useStatusWebSocket({ enabled: true }), {
      wrapper: createWrapper(),
    });

    // Connect
    act(() => {
      mockWebSocket.readyState = WebSocket.OPEN;
      mockWebSocket.onopen?.();
    });

    unmount();

    expect(mockClose).toHaveBeenCalled();
  });

  it('should allow manual disconnect', () => {
    const { result } = renderHook(() => useStatusWebSocket({ enabled: true }), {
      wrapper: createWrapper(),
    });

    // Connect
    act(() => {
      mockWebSocket.readyState = WebSocket.OPEN;
      mockWebSocket.onopen?.();
    });

    expect(result.current.isConnected).toBe(true);

    act(() => {
      result.current.disconnect();
    });

    expect(mockClose).toHaveBeenCalled();
    expect(result.current.isConnected).toBe(false);
  });

  it('should allow manual reconnect', () => {
    const { result } = renderHook(() => useStatusWebSocket({ enabled: true }), {
      wrapper: createWrapper(),
    });

    // Connect
    act(() => {
      mockWebSocket.readyState = WebSocket.OPEN;
      mockWebSocket.onopen?.();
    });

    act(() => {
      result.current.reconnect();
    });

    // Should close and reopen connection
    expect(mockClose).toHaveBeenCalled();
    expect(global.WebSocket).toHaveBeenCalledTimes(2);
  });

  it('should track reconnection attempts', () => {
    const { result } = renderHook(() => useStatusWebSocket({ enabled: true }), {
      wrapper: createWrapper(),
    });

    // Multiple connection failures
    for (let i = 0; i < 3; i++) {
      act(() => {
        mockWebSocket.onclose?.({ code: 1006, reason: 'Connection failed' });
      });

      act(() => {
        vi.advanceTimersByTime(Math.min(1000 * Math.pow(2, i), 30000));
      });
    }

    expect(result.current.reconnectAttempt).toBeGreaterThan(0);
  });

  it('should stop reconnecting after max attempts', () => {
    const { result } = renderHook(() => useStatusWebSocket({ enabled: true }), {
      wrapper: createWrapper(),
    });

    // Exceed max reconnection attempts (10)
    for (let i = 0; i < 12; i++) {
      act(() => {
        mockWebSocket.onclose?.({ code: 1006, reason: 'Connection failed' });
      });

      act(() => {
        vi.advanceTimersByTime(30000);
      });
    }

    expect(result.current.error?.message).toContain('Max reconnection attempts');
  });

  it('should not reconnect when intentionally closed', () => {
    const { result } = renderHook(() => useStatusWebSocket({ enabled: true }), {
      wrapper: createWrapper(),
    });

    // Connect
    act(() => {
      mockWebSocket.readyState = WebSocket.OPEN;
      mockWebSocket.onopen?.();
    });

    // Disconnect intentionally
    act(() => {
      result.current.disconnect();
    });

    const connectionCount = (global.WebSocket as any).mock.calls.length;

    // Advance time
    act(() => {
      vi.advanceTimersByTime(10000);
    });

    // Should not have attempted reconnection
    expect((global.WebSocket as any).mock.calls.length).toBe(connectionCount);
  });

  it('should use http protocol for ws when not https', () => {
    Object.defineProperty(window, 'location', {
      value: {
        protocol: 'http:',
        host: 'localhost:3000',
      },
      writable: true,
    });

    renderHook(() => useStatusWebSocket({ enabled: true }), {
      wrapper: createWrapper(),
    });

    expect(global.WebSocket).toHaveBeenCalledWith('ws://localhost:3000/ws/v1/status');
  });
});

describe('useRealtimeStatus', () => {
  let mockWebSocket: any;

  beforeEach(() => {
    vi.clearAllMocks();

    mockWebSocket = {
      send: vi.fn(),
      close: vi.fn(),
      readyState: WebSocket.CONNECTING,
      onopen: null as Function | null,
      onmessage: null as Function | null,
      onerror: null as Function | null,
      onclose: null as Function | null,
    };

    global.WebSocket = vi.fn(() => mockWebSocket) as any;

    Object.defineProperty(window, 'location', {
      value: {
        protocol: 'https:',
        host: 'test.functionfly.io',
      },
      writable: true,
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should expose realtime status', () => {
    const { result } = renderHook(() => useRealtimeStatus({ enabled: true }), {
      wrapper: createWrapper(),
    });

    expect(typeof result.current.isRealtime).toBe('boolean');
    expect(typeof result.current.isConnecting).toBe('boolean');
    expect(typeof result.current.reconnect).toBe('function');
    expect(typeof result.current.disconnect).toBe('function');
  });

  it('should report realtime when connected', () => {
    const { result } = renderHook(() => useRealtimeStatus({ enabled: true }), {
      wrapper: createWrapper(),
    });

    act(() => {
      mockWebSocket.readyState = WebSocket.OPEN;
      mockWebSocket.onopen?.();
    });

    expect(result.current.isRealtime).toBe(true);
  });

  it('should expose wsError when connection fails', () => {
    const { result } = renderHook(() => useRealtimeStatus({ enabled: true }), {
      wrapper: createWrapper(),
    });

    act(() => {
      mockWebSocket.onerror?.(new Event('error'));
    });

    expect(result.current.wsError).toBeTruthy();
  });
});

describe('WebSocket Reconnection Delays', () => {
  let mockWebSocket: any;

  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();

    mockWebSocket = {
      send: vi.fn(),
      close: vi.fn(),
      readyState: WebSocket.CONNECTING,
      onopen: null as Function | null,
      onclose: null as Function | null,
    };

    global.WebSocket = vi.fn(() => mockWebSocket) as any;

    Object.defineProperty(window, 'location', {
      value: {
        protocol: 'https:',
        host: 'test.functionfly.io',
      },
      writable: true,
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  it('should use exponential backoff for reconnection', () => {
    renderHook(() => useStatusWebSocket({ enabled: true }), {
      wrapper: createWrapper(),
    });

    const expectedDelays = [1000, 2000, 4000, 8000, 16000];

    expectedDelays.forEach((expectedDelay, index) => {
      act(() => {
        mockWebSocket.onclose?.({ code: 1006, reason: 'Connection failed' });
      });

      // Check that WebSocket was not created yet (waiting for delay)
      const callCountBefore = (global.WebSocket as any).mock.calls.length;

      // Advance time just before expected delay
      act(() => {
        vi.advanceTimersByTime(expectedDelay - 100);
      });

      expect((global.WebSocket as any).mock.calls.length).toBe(callCountBefore);

      // Advance past expected delay
      act(() => {
        vi.advanceTimersByTime(200);
      });

      expect((global.WebSocket as any).mock.calls.length).toBe(callCountBefore + 1);
    });
  });

  it('should cap reconnection delay at 30 seconds', () => {
    renderHook(() => useStatusWebSocket({ enabled: true }), {
      wrapper: createWrapper(),
    });

    // Fail connection many times
    for (let i = 0; i < 10; i++) {
      act(() => {
        mockWebSocket.onclose?.({ code: 1006, reason: 'Connection failed' });
      });

      act(() => {
        vi.advanceTimersByTime(30000);
      });
    }

    // After many attempts, delay should be capped at 30 seconds
    const lastDelay = 30000;

    act(() => {
      mockWebSocket.onclose?.({ code: 1006, reason: 'Connection failed' });
    });

    const callCount = (global.WebSocket as any).mock.calls.length;

    act(() => {
      vi.advanceTimersByTime(29000);
    });

    // Should not have reconnected yet
    expect((global.WebSocket as any).mock.calls.length).toBe(callCount);

    act(() => {
      vi.advanceTimersByTime(2000);
    });

    // Should have reconnected
    expect((global.WebSocket as any).mock.calls.length).toBe(callCount + 1);
  });
});
