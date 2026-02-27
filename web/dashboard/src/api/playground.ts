import { apiClient } from "./client";

// Types based on the backend handler
export interface PlaygroundInfo {
  function_id: string;
  function_name: string;
  app_slug: string;
  version: string;
  status: string;
  playground_enabled: boolean;
  playground_config?: Record<string, unknown>;
  provider?: string;
  region?: string;
  deployed_url?: string;
}

export interface ExecuteRequest {
  input: unknown;
}

export interface ExecuteResponse {
  success: boolean;
  output?: unknown;
  error?: string;
  latency_ms: number;
  status_code: number;
}

export interface ExecutionHistoryItem {
  id: string;
  timestamp: number;
  input: unknown;
  output?: unknown;
  error?: string;
  latency_ms: number;
  status_code: number;
  success: boolean;
}

const HISTORY_KEY = "functionfly_playground_history";
const MAX_HISTORY_ITEMS = 10;

export const playgroundApi = {
  // Get function info for the playground
  getInfo: (appSlug: string, functionName: string) =>
    apiClient.get<PlaygroundInfo>(`/v1/run/${appSlug}/${functionName}/info`),

  // Execute the function with input
  execute: (appSlug: string, functionName: string, input: unknown) =>
    apiClient.post<ExecuteResponse>(`/v1/run/${appSlug}/${functionName}/execute`, {
      input,
    }),

  // Get execution history from localStorage
  getHistory: (appSlug: string, functionName: string): ExecutionHistoryItem[] => {
    try {
      const key = `${HISTORY_KEY}_${appSlug}_${functionName}`;
      const data = localStorage.getItem(key);
      return data ? JSON.parse(data) : [];
    } catch {
      return [];
    }
  },

  // Save execution to history
  saveToHistory: (
    appSlug: string,
    functionName: string,
    item: Omit<ExecutionHistoryItem, "id" | "timestamp">
  ): void => {
    try {
      const key = `${HISTORY_KEY}_${appSlug}_${functionName}`;
      const history = playgroundApi.getHistory(appSlug, functionName);

      const newItem: ExecutionHistoryItem = {
        ...item,
        id: crypto.randomUUID(),
        timestamp: Date.now(),
      };

      const updatedHistory = [newItem, ...history].slice(0, MAX_HISTORY_ITEMS);
      localStorage.setItem(key, JSON.stringify(updatedHistory));
    } catch (error) {
      console.error("Failed to save to history:", error);
    }
  },

  // Clear execution history
  clearHistory: (appSlug: string, functionName: string): void => {
    try {
      const key = `${HISTORY_KEY}_${appSlug}_${functionName}`;
      localStorage.removeItem(key);
    } catch (error) {
      console.error("Failed to clear history:", error);
    }
  },

  // Generate shareable URL with input pre-filled
  createShareableUrl: (appSlug: string, functionName: string, input: unknown): string => {
    const baseUrl = `${window.location.origin}/run/${appSlug}/${functionName}`;
    const encodedInput = btoa(JSON.stringify(input)).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
    return `${baseUrl}?input=${encodedInput}`;
  },

  // Parse input from URL params
  parseUrlInput: (): unknown | null => {
    try {
      const params = new URLSearchParams(window.location.search);
      const inputParam = params.get('input');
      if (!inputParam) return null;

      // Decode from base64url
      const base64 = inputParam.replace(/-/g, '+').replace(/_/g, '/');
      const decoded = atob(base64);
      return JSON.parse(decoded);
    } catch {
      return null;
    }
  },
};
