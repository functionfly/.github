import { CACHE_KEYS } from '@/lib/constants';
import { useCallback, useEffect, useRef, useState } from 'react';

export interface DeploymentUpdate {
  id: string;
  function_id: string;
  function_name: string;
  status: 'pending' | 'building' | 'deploying' | 'ready' | 'failed';
  version: number;
  started_at: string;
  completed_at?: string;
  error?: string;
}

export interface FunctionMetrics {
  function_id: string;
  function_name: string;
  invocations: number;
  errors: number;
  latency_p50_ms: number;
  latency_p99_ms: number;
  memory_usage_mb: number;
  cold_starts: number;
  timestamp: string;
}

export interface UseRealtimeDeploymentsOptions {
  enabled?: boolean;
  onDeploymentUpdate?: (deployment: DeploymentUpdate) => void;
  onFunctionMetrics?: (metrics: FunctionMetrics) => void;
}

export interface UseRealtimeDeploymentsResult {
  isConnected: boolean;
  error: string | null;
}

/**
 * Hook for real-time deployment and function metrics updates
 * Connects to WebSocket for live deployment status and metrics
 */
export function useRealtimeDeployments(
  options: UseRealtimeDeploymentsOptions = {}
): UseRealtimeDeploymentsResult {
  const { enabled = true, onDeploymentUpdate, onFunctionMetrics } = options;

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectRef = useRef<number>(0);
  const reconnectTimeoutRef = useRef<number | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const MAX_RECONNECT_ATTEMPTS = 5;

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    const token = sessionStorage.getItem(CACHE_KEYS.ADMIN_ACCESS_TOKEN);
    if (!token) {
      setError('Missing admin access token for realtime connection');
      return;
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const baseUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';
    const wsBase = baseUrl.replace(/^http/, 'ws');
    const wsUrl = new URL(`${protocol}//${new URL(wsBase).host}/v1/admin/realtime/deployments`);
    wsUrl.searchParams.set('token', token);

    const ws = new WebSocket(wsUrl.toString());

    ws.onopen = () => {
      setIsConnected(true);
      setError(null);
      reconnectRef.current = 0;

      // Subscribe to deployment updates
      ws.send(JSON.stringify({ type: 'subscribe', channel: 'deployments' }));
      // Subscribe to function metrics
      ws.send(JSON.stringify({ type: 'subscribe', channel: 'function_metrics' }));
    };

    ws.onmessage = (message) => {
      try {
        const data = JSON.parse(message.data as string);

        if (data.type === 'deployment_update' && onDeploymentUpdate) {
          onDeploymentUpdate(data.payload as DeploymentUpdate);
        } else if (data.type === 'function_metrics' && onFunctionMetrics) {
          onFunctionMetrics(data.payload as FunctionMetrics);
        }
      } catch {
        setError('Failed to parse realtime message');
      }
    };

    ws.onclose = () => {
      setIsConnected(false);
      if (reconnectRef.current < MAX_RECONNECT_ATTEMPTS) {
        reconnectRef.current += 1;
        const delay = 1000 * Math.pow(2, reconnectRef.current - 1);
        reconnectTimeoutRef.current = window.setTimeout(connect, delay);
      }
    };

    ws.onerror = () => {
      setError('Realtime WebSocket error');
    };

    wsRef.current = ws;
  }, [onDeploymentUpdate, onFunctionMetrics]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      window.clearTimeout(reconnectTimeoutRef.current);
    }
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    setIsConnected(false);
  }, []);

  useEffect(() => {
    if (enabled) {
      connect();
    }
    return () => disconnect();
  }, [enabled, connect, disconnect]);

  return { isConnected, error };
}

/**
 * Hook specifically for deployment status updates
 * Provides deployment progress and status changes
 */
export function useDeploymentUpdates(onUpdate?: (deployment: DeploymentUpdate) => void) {
  const callbackRef = useRef(onUpdate);
  callbackRef.current = onUpdate;

  return useRealtimeDeployments({
    enabled: true,
    onDeploymentUpdate: (deployment) => {
      if (callbackRef.current) {
        callbackRef.current(deployment);
      }
    },
  });
}

/**
 * Hook specifically for function metrics updates
 * Provides real-time function performance data
 */
export function useFunctionMetricsUpdates(onMetrics?: (metrics: FunctionMetrics) => void) {
  const callbackRef = useRef(onMetrics);
  callbackRef.current = onMetrics;

  return useRealtimeDeployments({
    enabled: true,
    onFunctionMetrics: (metrics) => {
      if (callbackRef.current) {
        callbackRef.current(metrics);
      }
    },
  });
}
