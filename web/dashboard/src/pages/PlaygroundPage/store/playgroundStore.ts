import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface ExecutionResult {
  ok: boolean;
  data?: unknown;
  cached: boolean;
  duration_ms: number;
  version: string;
  execution_id?: string;
  error?: {
    code: string;
    message: string;
  };
}

export interface ExecutionHistoryItem {
  id: string;
  timestamp: number;
  input: unknown;
  result: ExecutionResult;
  inputJson: string;
}

export interface FunctionInfo {
  author: string;
  name: string;
  version: string;
  title?: string;
  description?: string;
  runtime?: string;
  manifest?: {
    input?: {
      type: string;
      schema?: Record<string, unknown>;
      example?: unknown;
      description?: string;
      title?: string;
      required?: boolean;
    };
    output?: {
      type: string;
      schema?: Record<string, unknown>;
      example?: unknown;
      description?: string;
    };
    examples?: Array<{ name: string; input: unknown; description?: string }>;
  };
  tags?: string[];
  category?: string;
  price_per_call?: number;
  reliability_score?: number;
  cache_ttl?: number;
}

export type InputMode = 'form' | 'json';
export type InputTab = 'form' | 'json' | 'examples';
export type OutputTab = 'response' | 'headers' | 'timeline' | 'diff';
export type SidebarPanel = 'history' | 'schema' | 'snippets' | 'share' | 'info';

interface PlaygroundSettings {
  autoRun: boolean;
  showTimeline: boolean;
  showHeaders: boolean;
  inputMode: InputMode;
}

interface PlaygroundStore {
  // Function info
  functionInfo: FunctionInfo | null;
  setFunctionInfo: (info: FunctionInfo | null) => void;

  // Input state
  inputMode: InputMode;
  inputValue: unknown;
  inputJson: string;
  setInputMode: (mode: InputMode) => void;
  setInputValue: (value: unknown) => void;
  setInputJson: (json: string) => void;

  // Output state
  executionResult: ExecutionResult | null;
  isExecuting: boolean;
  executionHistory: ExecutionHistoryItem[];
  setExecutionResult: (result: ExecutionResult | null) => void;
  setIsExecuting: (executing: boolean) => void;
  addToHistory: (item: ExecutionHistoryItem) => void;
  removeFromHistory: (id: string) => void;
  clearHistory: () => void;

  // UI state
  activeInputTab: InputTab;
  activeOutputTab: OutputTab;
  activeSidebarPanel: SidebarPanel;
  sidebarOpen: boolean;
  setActiveInputTab: (tab: InputTab) => void;
  setActiveOutputTab: (tab: OutputTab) => void;
  setActiveSidebarPanel: (panel: SidebarPanel) => void;
  setSidebarOpen: (open: boolean) => void;

  // Settings
  settings: PlaygroundSettings;
  updateSettings: (settings: Partial<PlaygroundSettings>) => void;

  // Diff state
  diffBaseItem: ExecutionHistoryItem | null;
  setDiffBaseItem: (item: ExecutionHistoryItem | null) => void;

  // Actions
  loadFromHistory: (item: ExecutionHistoryItem) => void;
  resetPlayground: () => void;
  formatJson: () => void;
  execute: (author: string, name: string) => Promise<void>;
}

const MAX_HISTORY = 50;

export const usePlaygroundStore = create<PlaygroundStore>()(
  persist(
    (set, get) => ({
      // Function info
      functionInfo: null,
      setFunctionInfo: (info) => set({ functionInfo: info }),

      // Input state
      inputMode: 'form',
      inputValue: {},
      inputJson: '{}',
      setInputMode: (mode) => set({ inputMode: mode }),
      setInputValue: (value) => {
        set({
          inputValue: value,
          inputJson: JSON.stringify(value, null, 2),
        });
      },
      setInputJson: (json) => {
        set({ inputJson: json });
        try {
          const parsed = JSON.parse(json);
          set({ inputValue: parsed });
        } catch {
          // Invalid JSON, keep inputValue as-is
        }
      },

      // Output state
      executionResult: null,
      isExecuting: false,
      executionHistory: [],
      setExecutionResult: (result) => set({ executionResult: result }),
      setIsExecuting: (executing) => set({ isExecuting: executing }),
      addToHistory: (item) => {
        const history = get().executionHistory;
        const updated = [item, ...history].slice(0, MAX_HISTORY);
        set({ executionHistory: updated });
      },
      removeFromHistory: (id) => {
        set({ executionHistory: get().executionHistory.filter((h) => h.id !== id) });
      },
      clearHistory: () => set({ executionHistory: [] }),

      // UI state
      activeInputTab: 'form',
      activeOutputTab: 'response',
      activeSidebarPanel: 'history',
      sidebarOpen: true,
      setActiveInputTab: (tab) => set({ activeInputTab: tab }),
      setActiveOutputTab: (tab) => set({ activeOutputTab: tab }),
      setActiveSidebarPanel: (panel) => set({ activeSidebarPanel: panel }),
      setSidebarOpen: (open) => set({ sidebarOpen: open }),

      // Settings
      settings: {
        autoRun: false,
        showTimeline: true,
        showHeaders: true,
        inputMode: 'form',
      },
      updateSettings: (newSettings) => {
        set({ settings: { ...get().settings, ...newSettings } });
      },

      // Diff state
      diffBaseItem: null,
      setDiffBaseItem: (item) => set({ diffBaseItem: item }),

      // Actions
      loadFromHistory: (item) => {
        set({
          inputValue: item.input,
          inputJson: item.inputJson,
          executionResult: item.result,
          activeOutputTab: 'response',
        });
      },

      resetPlayground: () => {
        set({
          inputValue: {},
          inputJson: '{}',
          executionResult: null,
          activeOutputTab: 'response',
        });
      },

      formatJson: () => {
        const { inputJson } = get();
        try {
          const parsed = JSON.parse(inputJson);
          const formatted = JSON.stringify(parsed, null, 2);
          set({ inputJson: formatted, inputValue: parsed });
        } catch {
          // Invalid JSON, can't format
        }
      },

      execute: async (author: string, name: string) => {
        const { inputValue, isExecuting } = get();
        if (isExecuting) return;

        set({ isExecuting: true, executionResult: null });

        const startTime = Date.now();
        try {
          const response = await fetch(`/v1/fx/${author}/${name}`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'X-Playground': '1',
            },
            body: JSON.stringify(inputValue),
          });

          const result: ExecutionResult = await response.json();
          set({ executionResult: result });

          // Add to history
          const historyItem: ExecutionHistoryItem = {
            id: crypto.randomUUID(),
            timestamp: Date.now(),
            input: inputValue,
            result,
            inputJson: JSON.stringify(inputValue, null, 2),
          };
          get().addToHistory(historyItem);
        } catch (error) {
          const errorResult: ExecutionResult = {
            ok: false,
            cached: false,
            duration_ms: Date.now() - startTime,
            version: get().functionInfo?.version || 'unknown',
            error: {
              code: 'NETWORK_ERROR',
              message: error instanceof Error ? error.message : 'Failed to execute function',
            },
          };
          set({ executionResult: errorResult });
        } finally {
          set({ isExecuting: false });
        }
      },
    }),
    {
      name: 'playground-store',
      partialize: (state) => ({
        settings: state.settings,
        sidebarOpen: state.sidebarOpen,
        // Don't persist execution results or history (too large, per-function)
      }),
    }
  )
);
