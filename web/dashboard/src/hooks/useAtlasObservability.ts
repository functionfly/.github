'use client';

import { useState, useEffect, useCallback } from 'react';

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
  const token = localStorage.getItem('token');
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...options?.headers,
  };

  const response = await fetch(`${API_BASE}${url}`, {
    ...options,
    headers,
  });

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
  }

  return response.json();
}

export function useAtlasRuns(agentId?: string, status?: string) {
  const [runs, setRuns] = useState<Run[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const params = new URLSearchParams();
  if (agentId) params.set('agent_id', agentId);
  if (status) params.set('status', status);

  useEffect(() => {
    const fetchRuns = async () => {
      setLoading(true);
      try {
        const data = await fetchJSON(`/v1/agent-observability/runs?${params}`);
        setRuns(data.runs || []);
        setError(null);
      } catch (e) {
        setError(e as Error);
      } finally {
        setLoading(false);
      }
    };

    fetchRuns();
  }, [agentId, status]);

  return { runs, loading, error };
}

export function useAtlasEvents(runId: string | undefined) {
  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!runId) {
      setEvents([]);
      return;
    }

    const fetchEvents = async () => {
      setLoading(true);
      try {
        const data = await fetchJSON(`/v1/agent-observability/runs/${runId}/events`);
        setEvents(data.events || []);
        setError(null);
      } catch (e) {
        setError(e as Error);
      } finally {
        setLoading(false);
      }
    };

    fetchEvents();
  }, [runId]);

  return { events, loading, error };
}

export function useAtlasStream(runId: string | undefined) {
  const [events, setEvents] = useState<Event[]>([]);
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    if (!runId) {
      setEvents([]);
      setConnected(false);
      return;
    }

    const wsUrl = `${API_BASE.replace('http', 'ws')}/v1/agent-observability/runs/${runId}/stream`;
    const token = localStorage.getItem('token');

    const headers: Record<string, string> = {};
    if (token) headers['Authorization'] = `Bearer ${token}`;

    const ws = new WebSocket(wsUrl);

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

    return () => {
      ws.close();
    };
  }, [runId]);

  return { events, connected };
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