import { create } from 'zustand';
import type { ParsedFunction, ImportConfig, AnalyzerStatus } from '../types/codePaste';

interface CodePasteStore {
  code: string;
  language: string | null;
  confidence: number;
  functions: ParsedFunction[];
  selectedIds: Set<string>;
  importConfig: ImportConfig;
  status: AnalyzerStatus;
  error: string | null;

  setCode: (code: string) => void;
  setLanguage: (language: string, confidence?: number) => void;
  setFunctions: (functions: ParsedFunction[]) => void;
  toggleSelection: (id: string) => void;
  selectAll: () => void;
  clearSelection: () => void;
  setImportConfig: (config: Partial<ImportConfig>) => void;
  setStatus: (status: AnalyzerStatus) => void;
  setError: (error: string | null) => void;
  reset: () => void;
}

const initialImportConfig: ImportConfig = {
  visibility: 'private',
  providers: ['cloud'],
  region: 'us-east-1',
};

export const useCodePasteStore = create<CodePasteStore>((set) => ({
  code: '',
  language: null,
  confidence: 0,
  functions: [],
  selectedIds: new Set(),
  importConfig: initialImportConfig,
  status: 'idle',
  error: null,

  setCode: (code) => set({
    code,
    status: code.length > 0 ? 'typing' : 'idle'
  }),

  setLanguage: (language, confidence = 0) => set({ language, confidence }),

  setFunctions: (functions) => set({
    functions,
    status: 'parsed',
    selectedIds: new Set(functions.map(f => f.id)),
    error: null,
  }),

  toggleSelection: (id) => set((state) => {
    const newIds = new Set(state.selectedIds);
    if (newIds.has(id)) {
      newIds.delete(id);
    } else {
      newIds.add(id);
    }
    return { selectedIds: newIds };
  }),

  selectAll: () => set((state) => ({
    selectedIds: new Set(state.functions.map(f => f.id))
  })),

  clearSelection: () => set({ selectedIds: new Set() }),

  setImportConfig: (config) => set((state) => ({
    importConfig: { ...state.importConfig, ...config }
  })),

  setStatus: (status) => set({ status }),

  setError: (error) => set({ error, status: error ? 'error' : 'idle' }),

  reset: () => set({
    code: '',
    language: null,
    confidence: 0,
    functions: [],
    selectedIds: new Set(),
    importConfig: initialImportConfig,
    status: 'idle',
    error: null,
  }),
}));