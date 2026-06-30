import { API_URLS } from '@/lib/api-urls';
import { useCallback, useEffect, useRef, useState } from 'react';

export interface PlatformMetrics {
  uptime: number;
  latency: number;
  failoverRate: number;
  status: 'operational' | 'degraded' | 'outage';
  timestamp: string;
}

const FALLBACK_POLL_INTERVAL = 30_000;

export function usePlatformStream(): PlatformMetrics {
  const [metrics, setMetrics] = useState<PlatformMetrics>({
    uptime: 0,
    latency: 0,
    failoverRate: 0,
    status: 'operational',
    timestamp: '',
  });

  const eventSourceRef = useRef<EventSource | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const reconnectAttemptRef = useRef(0);

  const connectSSE = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    const es = new EventSource(API_URLS.platform.metricsStream);
    eventSourceRef.current = es;

    es.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as PlatformMetrics;
        setMetrics(data);
        reconnectAttemptRef.current = 0;
      } catch {
        // ignore parse errors
      }
    };

    es.onerror = () => {
      es.close();
      eventSourceRef.current = null;
      const attempt = reconnectAttemptRef.current;
      const delay = Math.min(1000 * Math.pow(2, attempt), 30_000);
      reconnectAttemptRef.current = attempt + 1;
      reconnectTimerRef.current = setTimeout(connectSSE, delay);
    };
  }, []);

  const pollFallback = useCallback(async () => {
    try {
      const res = await fetch(API_URLS.platform.metricsGlobal);
      if (res.ok) {
        const data = (await res.json()) as PlatformMetrics;
        setMetrics(data);
      }
    } catch {
      // silent
    }
  }, []);

  useEffect(() => {
    connectSSE();
    const pollInterval = setInterval(pollFallback, FALLBACK_POLL_INTERVAL);

    return () => {
      eventSourceRef.current?.close();
      clearTimeout(reconnectTimerRef.current);
      clearInterval(pollInterval);
    };
  }, [connectSSE, pollFallback]);

  return metrics;
}
