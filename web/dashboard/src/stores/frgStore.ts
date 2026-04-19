/**
 * FRG (Function Runtime Graph) Zustand Store
 * Manages graph editor state with persistence and real-time sync
 */

import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import type {
  FRGNode,
  FRGEdge,
  GraphDefinition,
  GraphInstance,
  GraphEvent,
  FunctionCatalogItem,
  RuntimeNodeState,
  RuntimeEdgeState,
  AISuggestion,
  TestCase,
  SmartConnection,
  NodeExecutionStatus,
  ExecutionResult,
} from '@/types/frg';
import type { Connection, Viewport, NodeChange, EdgeChange } from '@xyflow/react';

// Editor modes
export type EditorMode = 'edit' | 'debug' | 'test' | 'view';

// Sidebar visibility
export type SidebarPanel = 
  | 'library' 
  | 'inspector' 
  | 'ai' 
  | 'test' 
  | 'versions'
  | 'evolution'
  | null;

// Evolution suggestion status
export type EvolutionSuggestionStatus = 'pending' | 'approved' | 'rejected' | 'implemented';

// Evolution suggestion interface
export interface EvolutionSuggestion {
  id: string;
  type: string;
  status: EvolutionSuggestionStatus;
  data: Record<string, unknown>;
  description: string;
  expectedImpact: number;
  confidence: 'low' | 'medium' | 'high';
  createdAt: string;
  implementedAt?: string;
  approvedBy?: string;
}

// Evolution status interface
export interface EvolutionStatus {
  agentId: string;
  evolutionEnabled: boolean;
  pendingCount: number;
  implementedCount: number;
  canEvolve: boolean;
}

interface FRGState {
  // Graph definition
  definition: Partial<GraphDefinition> | null;
  setDefinition: (definition: Partial<GraphDefinition> | null) => void;
  
  // React Flow elements
  nodes: FRGNode[];
  edges: FRGEdge[];
  setNodes: (nodes: FRGNode[] | ((prev: FRGNode[]) => FRGNode[])) => void;
  setEdges: (edges: FRGEdge[] | ((prev: FRGEdge[]) => FRGEdge[])) => void;
  onNodesChange: (changes: NodeChange<FRGNode>[]) => void;
  onEdgesChange: (changes: EdgeChange<FRGEdge>[]) => void;
  
  // Selection
  selectedNodeId: string | null;
  selectedEdgeId: string | null;
  setSelectedNode: (id: string | null) => void;
  setSelectedEdge: (id: string | null) => void;
  
  // Editor state
  editorMode: EditorMode;
  setEditorMode: (mode: EditorMode) => void;
  
  // Viewport
  viewport: Viewport;
  setViewport: (viewport: Viewport) => void;
  
  // Sidebar panels
  leftPanel: SidebarPanel;
  rightPanel: SidebarPanel;
  toggleLeftPanel: (panel: SidebarPanel) => void;
  toggleRightPanel: (panel: SidebarPanel) => void;
  
  // Function library
  libraryFunctions: FunctionCatalogItem[];
  librarySearch: string;
  libraryCategory: string | null;
  setLibrarySearch: (search: string) => void;
  setLibraryCategory: (category: string | null) => void;
  
  // Execution state
  activeInstance: GraphInstance | null;
  executionStatus: 'idle' | 'running' | 'streaming' | 'paused' | 'completed' | 'failed';
  executionProgress: number;
  executionResult: ExecutionResult | null;
  
  // Runtime state for live updates
  nodeRuntimeStates: Record<string, RuntimeNodeState>;
  edgeRuntimeStates: Record<string, RuntimeEdgeState>;
  
  // Execution controls
  startExecution: (input?: unknown) => void;
  pauseExecution: () => void;
  resumeExecution: () => void;
  stopExecution: () => void;
  stepExecution: () => void;
  
  // Live events
  events: GraphEvent[];
  addEvent: (event: GraphEvent) => void;
  clearEvents: () => void;
  
  // Smart connections
  smartConnections: SmartConnection[];
  setSmartConnections: (connections: SmartConnection[]) => void;
  
  // AI suggestions
  aiSuggestions: AISuggestion[];
  setAiSuggestions: (suggestions: AISuggestion[]) => void;
  dismissAiSuggestion: (id: string) => void;
  applyAiSuggestion: (id: string) => void;
  isAiLoading: boolean;
  setAiLoading: (loading: boolean) => void;
  
  // Test runner
  testCases: TestCase[];
  activeTestCase: string | null;
  testResults: Record<string, TestCase>;
  setTestCases: (tests: TestCase[]) => void;
  addTestCase: (test: TestCase) => void;
  updateTestCase: (id: string, updates: Partial<TestCase>) => void;
  removeTestCase: (id: string) => void;
  runTest: (id: string) => void;
  runAllTests: () => void;
  
  // Versions
  versions: { version: string; createdAt: string; isPublished: boolean }[];
  selectedVersion: string | null;
  compareVersions: [string, string] | null;
  setSelectedVersion: (version: string | null) => void;
  setCompareVersions: (versions: [string, string] | null) => void;
  
  // Undo/Redo stack
  history: { nodes: FRGNode[]; edges: FRGEdge[] }[];
  historyIndex: number;
  canUndo: boolean;
  canRedo: boolean;
  saveToHistory: () => void;
  undo: () => void;
  redo: () => void;
  
  // Dirty state
  isDirty: boolean;
  markDirty: () => void;
  markClean: () => void;
  
  // Loading states
  isLoading: boolean;
  setIsLoading: (loading: boolean) => void;
  
  // Error handling
  error: string | null;
  setError: (error: string | null) => void;
  
  // Actions
  addNode: (node: FRGNode) => void;
  removeNode: (id: string) => void;
  updateNode: (id: string, data: Partial<FRGNode>) => void;
  addEdge: (edge: FRGEdge) => void;
  removeEdge: (id: string) => void;
  updateEdge: (id: string, data: Partial<FRGEdge>) => void;
  onConnect: (connection: Connection) => void;
  
  // Data flow animation
  dataFlowParticles: Array<{
    id: string;
    edgeId: string;
    progress: number;
    data: unknown;
  }>;
  addDataFlowParticle: (edgeId: string, data: unknown) => void;
  removeDataFlowParticle: (id: string) => void;
  updateParticleProgress: (id: string, progress: number) => void;
  
  // Evolution state (Phase 1: Agent Evolution Mode)
  evolutionStatus: EvolutionStatus | null;
  evolutionSuggestions: EvolutionSuggestion[];
  evolutionHistory: EvolutionSuggestion[];
  isEvolutionLoading: boolean;
  evolutionError: string | null;
  
  // Evolution actions
  setEvolutionStatus: (status: EvolutionStatus | null) => void;
  setEvolutionSuggestions: (suggestions: EvolutionSuggestion[]) => void;
  setEvolutionHistory: (history: EvolutionSuggestion[]) => void;
  approveEvolutionSuggestion: (id: string) => void;
  rejectEvolutionSuggestion: (id: string) => void;
  toggleEvolutionMode: (enabled: boolean) => Promise<void>;
  fetchEvolutionStatus: () => Promise<void>;
  fetchEvolutionSuggestions: () => Promise<void>;
  fetchEvolutionHistory: () => Promise<void>;
  triggerEvolutionAnalysis: () => Promise<void>;
}

export const useFRGStore = create<FRGState>()(
  immer(
    persist(
      (set, get) => ({
        // Initial state
        definition: null,
        nodes: [],
        edges: [],
        selectedNodeId: null,
        selectedEdgeId: null,
        editorMode: 'edit',
        viewport: { x: 0, y: 0, zoom: 1 },
        leftPanel: 'library',
        rightPanel: null,
        libraryFunctions: [],
        librarySearch: '',
        libraryCategory: null,
        activeInstance: null,
        executionStatus: 'idle',
        executionProgress: 0,
        executionResult: null,
        nodeRuntimeStates: {},
        edgeRuntimeStates: {},
        events: [],
        smartConnections: [],
        aiSuggestions: [],
        isAiLoading: false,
        testCases: [],
        activeTestCase: null,
        testResults: {},
        versions: [],
        selectedVersion: null,
        compareVersions: null,
        history: [],
        historyIndex: -1,
        canUndo: false,
        canRedo: false,
        isDirty: false,
        isLoading: false,
        error: null,
        dataFlowParticles: [],
        
        // Evolution initial state
        evolutionStatus: null,
        evolutionSuggestions: [],
        evolutionHistory: [],
        isEvolutionLoading: false,
        evolutionError: null,
        
        // Definition actions
        setDefinition: (definition) => set({ definition }),
        
        // Node/Edge setters
        setNodes: (nodes) => {
          const newNodes = typeof nodes === 'function' ? nodes(get().nodes) : nodes;
          set({ nodes: newNodes, isDirty: true });
        },
        setEdges: (edges) => {
          const newEdges = typeof edges === 'function' ? edges(get().edges) : edges;
          set({ edges: newEdges, isDirty: true });
        },
        
        onNodesChange: (changes) => {
          // This would use applyNodeChanges from React Flow
          // For now, simplified implementation
          get().saveToHistory();
        },
        onEdgesChange: (changes) => {
          // This would use applyEdgeChanges from React Flow
          get().saveToHistory();
        },
        
        // Selection
        setSelectedNode: (id) => {
          set({ 
            selectedNodeId: id, 
            selectedEdgeId: null,
            rightPanel: id ? 'inspector' : null 
          });
        },
        setSelectedEdge: (id) => {
          set({ 
            selectedEdgeId: id, 
            selectedNodeId: null 
          });
        },
        
        // Editor mode
        setEditorMode: (mode) => set({ editorMode: mode }),
        
        // Viewport
        setViewport: (viewport) => set({ viewport }),
        
        // Sidebars
        toggleLeftPanel: (panel) => set((state) => ({
          leftPanel: state.leftPanel === panel ? null : panel,
        })),
        toggleRightPanel: (panel) => set((state) => ({
          rightPanel: state.rightPanel === panel ? null : panel,
        })),
        
        // Library
        setLibrarySearch: (search) => set({ librarySearch: search }),
        setLibraryCategory: (category) => set({ libraryCategory: category }),
        
        // Execution controls
        startExecution: (input) => {
          set({ 
            executionStatus: 'running',
            executionProgress: 0,
            events: [],
            nodeRuntimeStates: {},
            edgeRuntimeStates: {},
          });
        },
        pauseExecution: () => set({ executionStatus: 'paused' }),
        resumeExecution: () => set({ executionStatus: 'running' }),
        stopExecution: () => {
          set({ 
            executionStatus: 'idle',
            executionProgress: 0,
            activeInstance: null,
          });
        },
        stepExecution: () => {
          // Step execution - advance to next node
          const { executionStatus, executionProgress } = get();
          if (executionStatus === "paused" || executionStatus === "running") {
            const newProgress = Math.min(executionProgress + 10, 100);
            set({ executionProgress: newProgress });
            if (newProgress >= 100) {
              set({ executionStatus: "completed" });
            }
          }
        },
        
        // Events
        addEvent: (event) => set((state) => ({
          events: [...state.events.slice(-999), event],
        })),
        clearEvents: () => set({ events: [] }),
        
        // Smart connections
        setSmartConnections: (connections) => set({ smartConnections: connections }),
        
        // AI suggestions
        setAiSuggestions: (suggestions) => set({ aiSuggestions: suggestions }),
        dismissAiSuggestion: (id) => set((state) => ({
          aiSuggestions: state.aiSuggestions.filter(s => s.id !== id),
        })),
        applyAiSuggestion: (id) => {
          const suggestion = get().aiSuggestions.find(s => s.id === id);
          if (suggestion) {
            // Apply the suggestion to the graph
            get().saveToHistory();
            get().dismissAiSuggestion(id);
          }
        },
        setAiLoading: (loading) => set({ isAiLoading: loading }),
        
        // Test runner
        setTestCases: (tests) => set({ testCases: tests }),
        addTestCase: (test) => set((state) => ({
          testCases: [...state.testCases, test],
          activeTestCase: test.id,
        })),
        updateTestCase: (id, updates) => set((state) => ({
          testCases: state.testCases.map(t => 
            t.id === id ? { ...t, ...updates } : t
          ),
        })),
        removeTestCase: (id) => set((state) => ({
          testCases: state.testCases.filter(t => t.id !== id),
          activeTestCase: state.activeTestCase === id ? null : state.activeTestCase,
        })),
        runTest: (id) => {
          get().updateTestCase(id, { status: 'running' });
        },
        runAllTests: () => {
          get().testCases.forEach(t => get().runTest(t.id));
        },
        
        // Versions
        setSelectedVersion: (version) => set({ selectedVersion: version }),
        setCompareVersions: (versions) => set({ compareVersions: versions }),
        
        // History
        saveToHistory: () => {
          const { nodes, edges, history, historyIndex } = get();
          const newHistory = history.slice(0, historyIndex + 1);
          newHistory.push({ nodes: [...nodes], edges: [...edges] });
          set({
            history: newHistory.slice(-50), // Keep last 50 states
            historyIndex: Math.min(historyIndex + 1, 49),
            canUndo: true,
            canRedo: false,
          });
        },
        undo: () => {
          const { history, historyIndex } = get();
          if (historyIndex > 0) {
            const newIndex = historyIndex - 1;
            const { nodes, edges } = history[newIndex];
            set({
              nodes: [...nodes],
              edges: [...edges],
              historyIndex: newIndex,
              canUndo: newIndex > 0,
              canRedo: true,
            });
          }
        },
        redo: () => {
          const { history, historyIndex } = get();
          if (historyIndex < history.length - 1) {
            const newIndex = historyIndex + 1;
            const { nodes, edges } = history[newIndex];
            set({
              nodes: [...nodes],
              edges: [...edges],
              historyIndex: newIndex,
              canUndo: true,
              canRedo: newIndex < history.length - 1,
            });
          }
        },
        
        // Dirty state
        markDirty: () => set({ isDirty: true }),
        markClean: () => set({ isDirty: false }),
        
        // Loading
        setIsLoading: (loading) => set({ isLoading: loading }),
        setError: (error) => set({ error }),
        
        // CRUD operations
        addNode: (node) => {
          get().saveToHistory();
          set((state) => ({ nodes: [...state.nodes, node], isDirty: true }));
        },
        removeNode: (id) => {
          get().saveToHistory();
          set((state) => ({
            nodes: state.nodes.filter(n => n.id !== id),
            edges: state.edges.filter(e => e.source !== id && e.target !== id),
            isDirty: true,
          }));
        },
        updateNode: (id, data) => {
          set((state) => ({
            nodes: state.nodes.map(n => 
              n.id === id ? { ...n, ...data } : n
            ),
            isDirty: true,
          }));
        },
        addEdge: (edge) => {
          get().saveToHistory();
          set((state) => ({ edges: [...state.edges, edge], isDirty: true }));
        },
        removeEdge: (id) => {
          get().saveToHistory();
          set((state) => ({
            edges: state.edges.filter(e => e.id !== id),
            isDirty: true,
          }));
        },
        updateEdge: (id, data) => {
          set((state) => ({
            edges: state.edges.map(e => 
              e.id === id ? { ...e, ...data } : e
            ),
            isDirty: true,
          }));
        },
        onConnect: (connection) => {
          get().saveToHistory();
          const newEdge: FRGEdge = {
            id: `e-${connection.source}-${connection.target}`,
            source: connection.source!,
            target: connection.target!,
            sourceHandle: connection.sourceHandle,
            targetHandle: connection.targetHandle,
            type: 'smoothstep',
            animated: false,
            data: {
              mapping: { sourcePath: '*', targetPath: '*' },
              isValid: true,
              runtimeState: { 
                status: 'idle', 
                recordsTransferred: 0, 
                bytesTransferred: 0,
                isDataFlowing: false,
                flowProgress: 0,
              },
            },
          };
          set((state) => ({ edges: [...state.edges, newEdge], isDirty: true }));
        },
        
        // Data flow animation
        addDataFlowParticle: (edgeId, data) => set((state) => ({
          dataFlowParticles: [
            ...state.dataFlowParticles,
            {
              id: `particle-${Date.now()}-${Math.random()}`,
              edgeId,
              progress: 0,
              data,
            },
          ],
        })),
        removeDataFlowParticle: (id) => set((state) => ({
          dataFlowParticles: state.dataFlowParticles.filter(p => p.id !== id),
        })),
        updateParticleProgress: (id, progress) => set((state) => ({
          dataFlowParticles: state.dataFlowParticles.map(p =>
            p.id === id ? { ...p, progress } : p
          ),
        })),
        
        // Evolution actions (Phase 1: Agent Evolution Mode)
        setEvolutionStatus: (status) => set({ evolutionStatus: status }),
        setEvolutionSuggestions: (suggestions) => set({ evolutionSuggestions: suggestions }),
        setEvolutionHistory: (history) => set({ evolutionHistory: history }),
        
        approveEvolutionSuggestion: (id) => {
          set((state) => ({
            evolutionSuggestions: state.evolutionSuggestions.map(s =>
              s.id === id ? { ...s, status: 'approved' as const } : s
            ),
          }));
        },
        
        rejectEvolutionSuggestion: (id) => {
          set((state) => ({
            evolutionSuggestions: state.evolutionSuggestions.map(s =>
              s.id === id ? { ...s, status: 'rejected' as const } : s
            ),
          }));
        },
        
        toggleEvolutionMode: async (enabled) => {
          set({ isEvolutionLoading: true, evolutionError: null });
          try {
            const definition = get().definition;
            if (!definition?.id) throw new Error('No graph loaded');
            
            const response = await fetch(`/api/agents/${definition.id}/evolution/auto-enable`, {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ enabled }),
            });
            
            if (!response.ok) throw new Error('Failed to toggle evolution mode');
            
            // Update local state
            set((state) => ({
              evolutionStatus: state.evolutionStatus 
                ? { ...state.evolutionStatus, evolutionEnabled: enabled }
                : { 
                    agentId: definition.id!, 
                    evolutionEnabled: enabled, 
                    pendingCount: 0, 
                    implementedCount: 0,
                    canEvolve: enabled
                  },
            }));
          } catch (err) {
            set({ evolutionError: err instanceof Error ? err.message : 'Unknown error' });
          } finally {
            set({ isEvolutionLoading: false });
          }
        },
        
        fetchEvolutionStatus: async () => {
          const definition = get().definition;
          if (!definition?.id) return;
          
          try {
            const response = await fetch(`/api/agents/${definition.id}/evolution/status`);
            if (!response.ok) throw new Error('Failed to fetch evolution status');
            
            const data = await response.json();
            if (data.ok) {
              set({ evolutionStatus: data.status });
            }
          } catch (err) {
            console.error('Failed to fetch evolution status:', err);
          }
        },
        
        fetchEvolutionSuggestions: async () => {
          const definition = get().definition;
          if (!definition?.id) return;
          
          set({ isEvolutionLoading: true });
          try {
            const response = await fetch(`/api/agents/${definition.id}/evolution/suggestions`);
            if (!response.ok) throw new Error('Failed to fetch suggestions');
            
            const data = await response.json();
            if (data.ok) {
              set({ evolutionSuggestions: data.suggestions });
            }
          } catch (err) {
            set({ evolutionError: err instanceof Error ? err.message : 'Unknown error' });
          } finally {
            set({ isEvolutionLoading: false });
          }
        },
        
        fetchEvolutionHistory: async () => {
          const definition = get().definition;
          if (!definition?.id) return;
          
          try {
            const response = await fetch(`/api/agents/${definition.id}/evolution/history`);
            if (!response.ok) throw new Error('Failed to fetch history');
            
            const data = await response.json();
            if (data.ok) {
              set({ evolutionHistory: data.history });
            }
          } catch (err) {
            console.error('Failed to fetch evolution history:', err);
          }
        },
        
        triggerEvolutionAnalysis: async () => {
          const definition = get().definition;
          if (!definition?.id) return;
          
          set({ isEvolutionLoading: true, evolutionError: null });
          try {
            const response = await fetch(`/api/agents/${definition.id}/evolution/analyze`, {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
            });
            
            if (!response.ok) throw new Error('Failed to trigger analysis');
            
            const data = await response.json();
            if (data.ok && data.proposal_created) {
              // Refresh suggestions
              await get().fetchEvolutionSuggestions();
            }
          } catch (err) {
            set({ evolutionError: err instanceof Error ? err.message : 'Unknown error' });
          } finally {
            set({ isEvolutionLoading: false });
          }
        },
      }),
      {
        name: 'frg-editor-storage',
        partialize: (state) => ({
          // Only persist these fields
          definition: state.definition,
          nodes: state.nodes,
          edges: state.edges,
          viewport: state.viewport,
          testCases: state.testCases,
        }),
      }
    )
  )
);

// Computed selectors
export const selectFilteredLibrary = (state: ReturnType<typeof useFRGStore.getState>) => {
  const { libraryFunctions, librarySearch, libraryCategory } = state;
  return libraryFunctions.filter(fn => {
    const matchesSearch = !librarySearch || 
      fn.name.toLowerCase().includes(librarySearch.toLowerCase()) ||
      fn.description.toLowerCase().includes(librarySearch.toLowerCase()) ||
      fn.tags.some(t => t.toLowerCase().includes(librarySearch.toLowerCase()));
    const matchesCategory = !libraryCategory || fn.category === libraryCategory;
    return matchesSearch && matchesCategory;
  });
};

export const selectSelectedNode = (state: ReturnType<typeof useFRGStore.getState>) => {
  return state.nodes.find(n => n.id === state.selectedNodeId);
};

export const selectSelectedEdge = (state: ReturnType<typeof useFRGStore.getState>) => {
  return state.edges.find(e => e.id === state.selectedEdgeId);
};

export const selectExecutionStats = (state: ReturnType<typeof useFRGStore.getState>) => {
  const { nodeRuntimeStates, executionStatus, executionProgress } = state;
  const totalNodes = Object.keys(nodeRuntimeStates).length;
  const completedNodes = Object.values(nodeRuntimeStates).filter(
    n => n.status === 'completed' || n.status === 'failed'
  ).length;
  return {
    totalNodes,
    completedNodes,
    progress: executionProgress,
    status: executionStatus,
  };
};
