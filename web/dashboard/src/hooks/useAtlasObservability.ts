'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { tokenVault } from '../utils/token-vault';

const API_BASE = import.meta.env.VITE_API_URL || '';

interface Run {
  id: string;
  atlas_run_id: string;
  atlas_tenant_id: string;
  agent_id: string;
  agent_type: string;
  status: string;
  total_cost_usd: number;
  total_input_tokens: number;
  total_output_tokens: number;
  event_count: number;
  error_count: number;
  tool_call_count: number;
  started_at: string;
  ended_at?: string;
}

interface Event {
  event_id: string;
  sequence: number;
  kind: string;
  timestamp: string;
  system_id: string;
  payload: any;
  parent_id?: string;
  span_id?: string;
}

interface Config {
  sampling_rate: number;
  trace_errors_only: boolean;
  sample_head_percent: number;
  sample_tail_count: number;
  retention_days: number;
}

async function fetchJSON(url: string, options?: RequestInit) {
  await tokenVault.initialize();
  const token = await tokenVault.getAccessToken();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  if (options?.headers) {
    const customHeaders = options.headers as Record<string, string>;
    Object.assign(headers, customHeaders);
  }

  const response = await fetch(`${API_BASE}${url}`, {
    ...options,
    headers,
  });

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
  }

  return response.json();
}

export interface RunsFilter {
  agentId?: string;
  status?: string;
  startedAfter?: string;
  startedBefore?: string;
}

export interface RunsPagination {
  limit: number;
  offset: number;
  total: number;
}

export function useAtlasRuns(filter?: RunsFilter, pagination?: { limit: number; offset: number }) {
  const [runs, setRuns] = useState<Run[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [paginationInfo, setPaginationInfo] = useState<RunsPagination>({ limit: 20, offset: 0, total: 0 });

  const params = new URLSearchParams();
  if (filter?.agentId) params.set('agent_id', filter.agentId);
  if (filter?.status) params.set('status', filter.status);
  if (filter?.startedAfter) params.set('started_after', filter.startedAfter);
  if (filter?.startedBefore) params.set('started_before', filter.startedBefore);
  params.set('limit', String(pagination?.limit || 20));
  params.set('offset', String(pagination?.offset || 0));

  const fetchRuns = useCallback(async () => {
    setLoading(true);
    try {
      const data = await fetchJSON(`/v1/agent-observability/runs?${params}`);
      setRuns(data.runs || []);
      setPaginationInfo({ limit: data.limit || 20, offset: data.offset || 0, total: data.total || 0 });
      setError(null);
    } catch (e) {
      setError(e as Error);
    } finally {
      setLoading(false);
    }
  }, [filter?.agentId, filter?.status, filter?.startedAfter, filter?.startedBefore, pagination?.limit, pagination?.offset]);

  useEffect(() => {
    fetchRuns();
  }, [fetchRuns]);

  return { runs, loading, error, paginationInfo, refetch: fetchRuns };
}

export interface EventsPagination {
  limit: number;
  afterSequence: number;
  total: number;
}

export function useAtlasEvents(runId: string | undefined, pagination?: { limit: number; afterSequence: number }) {
  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [paginationInfo, setPaginationInfo] = useState<EventsPagination>({ limit: 100, afterSequence: 0, total: 0 });

  useEffect(() => {
    if (!runId) {
      setEvents([]);
      return;
    }

    const fetchEvents = async () => {
      setLoading(true);
      try {
        const params = new URLSearchParams();
        params.set('limit', String(pagination?.limit || 100));
        params.set('after_sequence', String(pagination?.afterSequence || 0));
        const data = await fetchJSON(`/v1/agent-observability/runs/${runId}/events?${params}`);
        setEvents(data.events || []);
        setPaginationInfo({ limit: data.limit || 100, afterSequence: data.after_sequence || 0, total: data.total || 0 });
        setError(null);
      } catch (e) {
        setError(e as Error);
      } finally {
        setLoading(false);
      }
    };

    fetchEvents();
  }, [runId, pagination?.limit, pagination?.afterSequence]);

  return { events, loading, error, paginationInfo };
}

export function useAtlasStream(runId: string | undefined, options?: { autoRefresh?: boolean; refreshInterval?: number }) {
  const [events, setEvents] = useState<Event[]>([]);
  const [connected, setConnected] = useState(false);
  const [reconnectKey, setReconnectKey] = useState(0);
  const wsRef = useRef<WebSocket | null>(null);

  const reconnect = useCallback(() => {
    setReconnectKey((k) => k + 1);
  }, []);

  useEffect(() => {
    if (!runId) {
      setEvents([]);
      setConnected(false);
      return;
    }

    let ws: WebSocket | null = null;

    (async () => {
      const wsUrl = `${API_BASE.replace('http', 'ws')}/v1/agent-observability/runs/${runId}/stream`;
      await tokenVault.initialize();
      const token = await tokenVault.getAccessToken();

      const headers: Record<string, string> = {};
      if (token) headers['Authorization'] = `Bearer ${token}`;

      ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        setConnected(true);
        if (token) {
          ws.send(JSON.stringify({ headers }));
        }
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          setEvents((prev) => [...prev, data]);
        } catch (e) {
          console.error('Failed to parse WebSocket message:', e);
        }
      };

      ws.onclose = () => {
        setConnected(false);
      };

      ws.onerror = () => {
        setConnected(false);
      };
    })();

    return () => {
      if (ws) ws.close();
    };
  }, [runId, reconnectKey]);

  useEffect(() => {
    if (!options?.autoRefresh || !runId) return;

    const interval = setInterval(() => {
      if (!connected && wsRef.current?.readyState !== WebSocket.OPEN) {
        setReconnectKey((k) => k + 1);
      }
    }, options.refreshInterval || 5000);

    return () => clearInterval(interval);
  }, [options?.autoRefresh, options?.refreshInterval, runId, connected]);

  return { events, connected, reconnect };
}

export function useAtlasConfig() {
  const [config, setConfig] = useState<Config | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchConfig = async () => {
      setLoading(true);
      try {
        const data = await fetchJSON(`/v1/agent-observability/config`);
        setConfig(data);
        setError(null);
      } catch (e) {
        setError(e as Error);
      } finally {
        setLoading(false);
      }
    };

    fetchConfig();
  }, []);

  const updateConfig = useCallback(async (updates: Partial<Config>) => {
    try {
      const data = await fetchJSON(`/v1/agent-observability/config`, {
        method: 'PUT',
        body: JSON.stringify(updates),
      });
      setConfig(data);
      return data;
    } catch (e) {
      setError(e as Error);
      throw e;
    }
  }, []);

  return { config, loading, error, updateConfig };
}

export function useCreateRun() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const createRun = useCallback(async (params: {
    agent_id: string;
    agent_type: string;
    span_id?: string;
    parent_span_id?: string;
    metadata?: Record<string, string>;
  }): Promise<Run> => {
    setLoading(true);
    try {
      const data = await fetchJSON(`/v1/agent-observability/runs`, {
        method: 'POST',
        body: JSON.stringify(params),
      });
      setError(null);
      return data;
    } catch (e) {
      setError(e as Error);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  return { createRun, loading, error };
}
