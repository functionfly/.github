import { useState, useCallback, useEffect } from 'react';
import { agentApi } from '@/api/agent';

export interface WorkspaceEntry {
  name: string;
  path: string;
  is_dir: boolean;
  size: number;
  mod_time: string;
}

export interface WorkspaceManifest {
  agent_id: string;
  tenant_id: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
  file_count: number;
  total_bytes: number;
  structure: string[];
}

export interface HistoryEntry {
  timestamp: string;
  action: string;
  tool: string;
  path?: string;
  description: string;
  agent_id: string;
  session_id?: string;
  result?: string;
}

export function useWorkspace(agentId: string) {
  const [entries, setEntries] = useState<WorkspaceEntry[]>([]);
  const [currentPath, setCurrentPath] = useState('/');
  const [manifest, setManifest] = useState<WorkspaceManifest | null>(null);
  const [history, setHistory] = useState<HistoryEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchEntries = useCallback(async (path = '') => {
    setLoading(true);
    setError(null);
    try {
      const res = await agentApi.browseWorkspace(agentId, path);
      setEntries(res.entries || []);
      setCurrentPath(res.path || '/');
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to load workspace');
    } finally {
      setLoading(false);
    }
  }, [agentId]);

  const fetchManifest = useCallback(async () => {
    try {
      const res = await agentApi.getWorkspaceManifest(agentId);
      setManifest(res.manifest || null);
    } catch {
      // manifest may not exist yet
    }
  }, [agentId]);

  const fetchHistory = useCallback(async () => {
    try {
      const res = await agentApi.getWorkspaceHistory(agentId);
      setHistory(res.history || []);
    } catch {
      // history may not exist yet
    }
  }, [agentId]);

  const navigate = useCallback((path: string) => {
    fetchEntries(path);
  }, [fetchEntries]);

  const refresh = useCallback(() => {
    fetchEntries(currentPath);
    fetchManifest();
    fetchHistory();
  }, [fetchEntries, fetchManifest, fetchHistory, currentPath]);

  useEffect(() => {
    fetchEntries();
    fetchManifest();
    fetchHistory();
  }, [fetchEntries, fetchManifest, fetchHistory]);

  return {
    entries,
    currentPath,
    manifest,
    history,
    loading,
    error,
    navigate,
    refresh,
  };
}
