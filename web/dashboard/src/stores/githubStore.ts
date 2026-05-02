import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { GitHubConnection, GitHubRepo, ScanResult } from '@/types/github';

export interface ImportConfig {
  selectedFunctions: string[];
  visibilityOverrides: Record<string, 'public' | 'private' | 'unlisted'>;
  globalVisibility: 'public' | 'private' | 'unlisted';
  autoSync: boolean;
  syncBranches: string[];
  environmentMappings: Record<string, string>;
  templateId: string | null;
}

interface GitHubState {
  connection: GitHubConnection | null;
  isConnected: boolean;
  selectedRepo: GitHubRepo | null;
  scanResult: ScanResult | null;
  importConfig: ImportConfig;
  activeImportId: string | null;

  setConnection: (connection: GitHubConnection | null) => void;
  setIsConnected: (connected: boolean) => void;
  setSelectedRepo: (repo: GitHubRepo | null) => void;
  setScanResult: (result: ScanResult | null) => void;
  setActiveImportId: (id: string | null) => void;

  setSelectedFunctions: (functions: string[]) => void;
  toggleFunction: (functionName: string) => void;
  setVisibilityOverride: (functionName: string, visibility: 'public' | 'private' | 'unlisted') => void;
  setGlobalVisibility: (visibility: 'public' | 'private' | 'unlisted') => void;
  setAutoSync: (enabled: boolean) => void;
  setSyncBranches: (branches: string[]) => void;
  setEnvironmentMappings: (mappings: Record<string, string>) => void;
  setTemplateId: (id: string | null) => void;
  resetImportConfig: () => void;
  resetAll: () => void;
}

const defaultImportConfig: ImportConfig = {
  selectedFunctions: [],
  visibilityOverrides: {},
  globalVisibility: 'private',
  autoSync: false,
  syncBranches: [],
  environmentMappings: {},
  templateId: null,
};

export const useGitHubStore = create<GitHubState>()(
  persist(
    (set) => ({
      connection: null,
      isConnected: false,
      selectedRepo: null,
      scanResult: null,
      importConfig: { ...defaultImportConfig },
      activeImportId: null,

      setConnection: (connection) =>
        set({
          connection,
          isConnected: !!connection && connection.status === 'active',
        }),

      setIsConnected: (connected) => set({ isConnected: connected }),

      setSelectedRepo: (repo) =>
        set({
          selectedRepo: repo,
          scanResult: null,
          importConfig: { ...defaultImportConfig },
        }),

      setScanResult: (result) =>
        set((state) => ({
          scanResult: result,
          importConfig: {
            ...state.importConfig,
            selectedFunctions: result?.functions.map((f) => f.name) ?? [],
          },
        })),

      setActiveImportId: (id) => set({ activeImportId: id }),

      setSelectedFunctions: (functions) =>
        set((state) => ({
          importConfig: { ...state.importConfig, selectedFunctions: functions },
        })),

      toggleFunction: (functionName) =>
        set((state) => {
          const current = state.importConfig.selectedFunctions;
          const updated = current.includes(functionName)
            ? current.filter((f) => f !== functionName)
            : [...current, functionName];
          return {
            importConfig: { ...state.importConfig, selectedFunctions: updated },
          };
        }),

      setVisibilityOverride: (functionName, visibility) =>
        set((state) => ({
          importConfig: {
            ...state.importConfig,
            visibilityOverrides: {
              ...state.importConfig.visibilityOverrides,
              [functionName]: visibility,
            },
          },
        })),

      setGlobalVisibility: (visibility) =>
        set((state) => ({
          importConfig: { ...state.importConfig, globalVisibility: visibility },
        })),

      setAutoSync: (enabled) =>
        set((state) => ({
          importConfig: { ...state.importConfig, autoSync: enabled },
        })),

      setSyncBranches: (branches) =>
        set((state) => ({
          importConfig: { ...state.importConfig, syncBranches: branches },
        })),

      setEnvironmentMappings: (mappings) =>
        set((state) => ({
          importConfig: { ...state.importConfig, environmentMappings: mappings },
        })),

      setTemplateId: (id) =>
        set((state) => ({
          importConfig: { ...state.importConfig, templateId: id },
        })),

      resetImportConfig: () =>
        set({ importConfig: { ...defaultImportConfig } }),

      resetAll: () =>
        set({
          connection: null,
          isConnected: false,
          selectedRepo: null,
          scanResult: null,
          importConfig: { ...defaultImportConfig },
          activeImportId: null,
        }),
    }),
    {
      name: 'github-storage',
      partialize: (state) => ({
        importConfig: state.importConfig,
        activeImportId: state.activeImportId,
      }),
    }
  )
);
