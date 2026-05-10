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
  available_versions?: string[];
}

export interface PlaygroundVariable {
  id: string;
  name: string;
  value: string;
  description?: string;
}

export type InputMode = 'form' | 'json';
export type InputTab = 'form' | 'json' | 'examples';
export type OutputTab = 'response' | 'headers' | 'timeline' | 'diff';
export type SidebarPanel = 'history' | 'schema' | 'snippets' | 'share' | 'info' | 'variables';
export type DiffViewMode = 'output' | 'input-output';

interface PlaygroundSettings {
  autoRun: boolean;
  showTimeline: boolean;
  showHeaders: boolean;
  inputMode: InputMode;
  diffViewMode: DiffViewMode;
}

interface StreamingState {
  isStreaming: boolean;
  chunks: string[];
  partialData: unknown;
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

  // Streaming state
  streaming: StreamingState;
  startStreaming: () => void;
  addStreamChunk: (chunk: string) => void;
  setStreamingPartialData: (data: unknown) => void;
  stopStreaming: () => void;

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

  // Variables
  variables: PlaygroundVariable[];
  addVariable: (variable: PlaygroundVariable) => void;
  updateVariable: (id: string, updates: Partial<PlaygroundVariable>) => void;
  removeVariable: (id: string) => void;
  clearVariables: () => void;

  // Version selection
  selectedVersion: string | null;
  setSelectedVersion: (version: string | null) => void;

  // Actions
  loadFromHistory: (item: ExecutionHistoryItem) => void;
  resetPlayground: () => void;
  formatJson: () => void;
  execute: (author: string, name: string) => Promise<void>;
  substituteVariables: (input: unknown) => unknown;
}

// API Response shapes (for mapping)
// The execute endpoint returns ExecutionResponse format: {ok, data, duration_ms, version, ...}
interface PlaygroundExecuteResponse {
  success: boolean;
  output?: unknown;
  error?: string;
  latency_ms: number;
  status_code: number;
  version?: string;
  execution_id?: string;
}

function mapApiResponseToExecutionResult(response: PlaygroundExecuteResponse): ExecutionResult {
  return {
    ok: response.success,
    data: response.output,
    cached: false,
    duration_ms: response.latency_ms,
    version: response.version || 'unknown',
    execution_id: response.execution_id,
    error: response.error ? { code: 'EXECUTION_ERROR', message: response.error } : undefined,
  };
}

// Handle the actual API response format: {ok, data, cached, duration_ms, version, execution_id}
function mapExecutionResponseToExecutionResult(response: {
  ok: boolean;
  data: unknown;
  cached: boolean;
  duration_ms: number;
  version: string;
  execution_id?: string;
}): ExecutionResult {
  return {
    ok: response.ok,
    data: response.data,
    cached: response.cached,
    duration_ms: response.duration_ms,
    version: response.version || 'unknown',
    execution_id: response.execution_id,
    error: undefined,
  };
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

      // Streaming state
      streaming: {
        isStreaming: false,
        chunks: [],
        partialData: null,
      },
      startStreaming: () => {
        set({
          streaming: {
            isStreaming: true,
            chunks: [],
            partialData: null,
          },
        });
      },
      addStreamChunk: (chunk) => {
        const { streaming } = get();
        set({
          streaming: {
            ...streaming,
            chunks: [...streaming.chunks, chunk],
          },
        });
      },
      setStreamingPartialData: (data) => {
        const { streaming } = get();
        set({
          streaming: {
            ...streaming,
            partialData: data,
          },
        });
      },
      stopStreaming: () => {
        const { streaming } = get();
        set({
          streaming: {
            ...streaming,
            isStreaming: false,
          },
        });
      },

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
        diffViewMode: 'output',
      },
      updateSettings: (newSettings) => {
        set({ settings: { ...get().settings, ...newSettings } });
      },

      // Diff state
      diffBaseItem: null,
      setDiffBaseItem: (item) => set({ diffBaseItem: item }),

      // Variables
      variables: [],
      addVariable: (variable) => {
        set({ variables: [...get().variables, variable] });
      },
      updateVariable: (id, updates) => {
        set({
          variables: get().variables.map((v) =>
            v.id === id ? { ...v, ...updates } : v
          ),
        });
      },
      removeVariable: (id) => {
        set({ variables: get().variables.filter((v) => v.id !== id) });
      },
      clearVariables: () => set({ variables: [] }),

      // Version selection
      selectedVersion: null,
      setSelectedVersion: (version) => set({ selectedVersion: version }),

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

      substituteVariables: (input) => {
        const { variables } = get();
        if (variables.length === 0) return input;

        const inputStr = JSON.stringify(input);
        let result = inputStr;

        for (const variable of variables) {
          const placeholder = `{{${variable.name}}}`;
          result = result.split(placeholder).join(variable.value);
        }

        try {
          return JSON.parse(result);
        } catch {
          return input;
        }
      },

      execute: async (author: string, name: string) => {
        const { inputValue, isExecuting, substituteVariables, selectedVersion } = get();
        if (isExecuting) return;

        set({ isExecuting: true, executionResult: null });

        const startTime = Date.now();
        try {
          // Substitute variables in input before execution
          const resolvedInput = substituteVariables(inputValue);
          const version = selectedVersion || undefined;

          const url = version
            ? `/v1/fx/${author}/${name}?version=${version}`
            : `/v1/fx/${author}/${name}`;

          const response = await fetch(url, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'X-Playground': '1',
            },
            body: JSON.stringify({ input: resolvedInput }),
          });

          // Check if response is streaming
          const contentType = response.headers.get('content-type') || '';
          if (contentType.includes('text/event-stream') || contentType.includes('application/x-ndjson')) {
            // Handle streaming response
            get().startStreaming();
            const reader = response.body?.getReader();
            const decoder = new TextDecoder();

            if (reader) {
              try {
                while (true) {
                  const { done, value } = await reader.read();
                  if (done) break;

                  const chunk = decoder.decode(value, { stream: true });
                  get().addStreamChunk(chunk);

                  // Try to parse partial data from SSE format
                  const lines = chunk.split('\n');
                  for (const line of lines) {
                    if (line.startsWith('data: ')) {
                      try {
                        const data = JSON.parse(line.slice(6));
                        get().setStreamingPartialData(data);
                      } catch {
                        // ignore parse errors for partial data
                      }
                    }
                  }
                }
              } finally {
                reader.releaseLock();
              }
            }

            get().stopStreaming();

            // Build final result from chunks
            const { streaming } = get();
            let finalData = streaming.partialData;

            if (!finalData && streaming.chunks.length > 0) {
              // Try to parse accumulated chunks
              const fullText = streaming.chunks.join('');
              try {
                finalData = JSON.parse(fullText);
              } catch {
                finalData = streaming.chunks.join('');
              }
            }

            const streamingResult: ExecutionResult = {
              ok: true,
              data: finalData,
              cached: false,
              duration_ms: Date.now() - startTime,
              version: selectedVersion || get().functionInfo?.version || 'unknown',
            };
            set({ executionResult: streamingResult });

            const historyItem: ExecutionHistoryItem = {
              id: crypto.randomUUID(),
              timestamp: Date.now(),
              input: resolvedInput,
              result: streamingResult,
              inputJson: JSON.stringify(resolvedInput, null, 2),
            };
            get().addToHistory(historyItem);
          } else {
            // Normal response - parse as API response and map to ExecutionResult
            const apiResponse = await response.json();

            // Detect response format: {success, output, ...} (PlaygroundExecuteResponse) vs {ok, data, ...} (ExecutionResponse)
            let result: ExecutionResult;
            if (typeof (apiResponse as any).success === 'boolean') {
              // PlaygroundExecuteResponse format (from playground proxy)
              result = mapApiResponseToExecutionResult(apiResponse as PlaygroundExecuteResponse);
            } else if (typeof (apiResponse as any).ok === 'boolean') {
              // ExecutionResponse format (direct from API)
              result = mapExecutionResponseToExecutionResult(apiResponse as {
                ok: boolean;
                data: unknown;
                cached: boolean;
                duration_ms: number;
                version: string;
                execution_id?: string;
              });
            } else {
              // Unknown format - create error result
              result = {
                ok: false,
                cached: false,
                duration_ms: Date.now() - startTime,
                version: selectedVersion || get().functionInfo?.version || 'unknown',
                error: { code: 'INVALID_RESPONSE', message: 'Invalid response format from API' },
              };
            }

            set({ executionResult: result });

            // Add to history
            const historyItem: ExecutionHistoryItem = {
              id: crypto.randomUUID(),
              timestamp: Date.now(),
              input: resolvedInput,
              result,
              inputJson: JSON.stringify(resolvedInput, null, 2),
            };
            get().addToHistory(historyItem);
          }
        } catch (error) {
          const errorResult: ExecutionResult = {
            ok: false,
            cached: false,
            duration_ms: Date.now() - startTime,
            version: selectedVersion || get().functionInfo?.version || 'unknown',
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
        variables: state.variables,
        selectedVersion: state.selectedVersion,
      }),
    }
  )
);
