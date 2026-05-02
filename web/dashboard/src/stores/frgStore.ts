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
  GraphNodeRef,
  GraphEdgeDefinition,
} from '@/types/frg';
import { frgApi, type ExecuteGraphResponse } from '@/api/frg';
import { registryApi } from '@/api/registry';
import { apiClient } from '@/api/client';
import { getApiBaseUrl } from '@/lib/constants';
import { toast } from '@/components/ui';
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
  setNodeRuntimeState: (nodeId: string, state: Partial<RuntimeNodeState>) => void;
  
  // Graph metadata
  graphAuthor: string | null;
  graphName: string | null;
  graphVersion: string | null;
  
  // Graph load/save actions
  loadGraph: (author: string, name: string, version?: string) => Promise<void>;
  saveGraph: () => Promise<void>;
  createNewGraph: (name: string, executionMode?: string) => void;
  
  // Library
  fetchLibraryFunctions: () => Promise<void>;
  
  // Execution controls
  startExecution: (input?: Record<string, unknown>) => Promise<void>;
  runNode: (nodeId: string, input?: Record<string, unknown>) => Promise<void>;
  pauseExecution: () => void;
  resumeExecution: () => void;
  stopExecution: () => Promise<void>;
  pollInstanceStatus: () => Promise<void>;
  
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
  
  // Loading states (granular per-operation)
  isLoading: boolean;
  isSaving: boolean;
  isExecuting: boolean;
  setIsLoading: (loading: boolean) => void;
  setIsSaving: (saving: boolean) => void;
  setIsExecuting: (executing: boolean) => void;

  // Auto-save
  autoSaveEnabled: boolean;
  autoSaveInterval: number; // ms
  lastSavedAt: string | null;
  setAutoSaveEnabled: (enabled: boolean) => void;
  setAutoSaveInterval: (ms: number) => void;

  // Offline support
  isOffline: boolean;
  operationQueue: Array<{ id: string; type: string; payload: unknown; timestamp: string }>;
  queuedOperationCount: number;
  setIsOffline: (offline: boolean) => void;
  queueOperation: (type: string, payload: unknown) => void;
  processQueue: () => Promise<void>;
  clearQueue: () => void;
  
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
        setNodeRuntimeState: (nodeId, state) => set((prev) => ({
          nodeRuntimeStates: {
            ...prev.nodeRuntimeStates,
            [nodeId]: {
              status: 'idle',
              attemptCount: 0,
              durationMs: 0,
              isActive: false,
              ...prev.nodeRuntimeStates[nodeId],
              ...state,
            } as RuntimeNodeState,
          },
        })),
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
        isSaving: false,
        isExecuting: false,
        error: null,
        dataFlowParticles: [],
        autoSaveEnabled: true,
        autoSaveInterval: 30000, // 30 seconds
        lastSavedAt: null,
        isOffline: false,
        operationQueue: [],
        queuedOperationCount: 0,
        
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
          const { nodes, setNodes, saveToHistory } = get();
          const hasMeaningfulChange = changes.some(c =>
            c.type === 'position' || c.type === 'remove' || c.type === 'add' ||
            (c.type === 'dimensions') ||
            (c.type === 'select')
          );
          if (!hasMeaningfulChange) return;

          const applyNodeChanges = (nodes: FRGNode[], changes: NodeChange<FRGNode>[]): FRGNode[] => {
            return changes.reduce((acc, change) => {
              switch (change.type) {
                case 'position':
                  return acc.map(n => n.id === change.id ? { ...n, position: change.position } : n);
                case 'dimensions':
                  return acc.map(n => n.id === change.id ? {
                    ...n,
                    width: (change as { width?: number }).width,
                    height: (change as { height?: number }).height
                  } : n);
                case 'remove':
                  return acc.filter(n => n.id !== change.id);
                case 'select':
                  return acc.map(n => n.id === change.id ? { ...n, selected: change.selected } : n);
                default:
                  return acc;
              }
            }, nodes);
          };

          const newNodes = applyNodeChanges(nodes, changes);
          saveToHistory();
          setNodes(newNodes);
        },
        onEdgesChange: (changes) => {
          const { edges, setEdges, saveToHistory } = get();

          const applyEdgeChanges = (edges: FRGEdge[], changes: EdgeChange<FRGEdge>[]): FRGEdge[] => {
            return changes.reduce((acc, change) => {
              switch (change.type) {
                case 'remove':
                  return acc.filter(e => e.id !== change.id);
                case 'select':
                  return acc.map(e => e.id === change.id ? { ...e, selected: change.selected } : e);
                default:
                  return acc;
              }
            }, edges);
          };

          const newEdges = applyEdgeChanges(edges, changes);
          saveToHistory();
          setEdges(newEdges);
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
        setLibrarySearch: (search) => set({ librarySearch: search }),
        setLibraryCategory: (category) => set({ libraryCategory: category }),
        
        // Graph metadata
        graphAuthor: null,
        graphName: null,
        graphVersion: null,

        // Library
        fetchLibraryFunctions: async () => {
          set({ isLoading: true, error: null });
          try {
            const response = await registryApi.getFunctions({ limit: 50 });
            const functions = response.functions || [];
            
            const categoryColors: Record<string, string> = {
              api: '#6366f1',
              data: '#10b981',
              text: '#3b82f6',
              image: '#8b5cf6',
              video: '#ef4444',
              audio: '#f59e0b',
              code: '#14b8a6',
              ml: '#ec4899',
              default: '#6b7280',
            };
            
            const categoryIcons: Record<string, string> = {
              api: 'globe',
              data: 'database',
              text: 'file-text',
              image: 'image',
              video: 'video',
              audio: 'music',
              code: 'code',
              ml: 'cpu',
              default: 'code',
            };
            
            const libraryFunctions: FunctionCatalogItem[] = functions.map((fn) => {
              const category = fn.category || 'default';
              return {
                id: fn.id,
                author: fn.author,
                name: fn.name,
                version: fn.latest_version || '1.0.0',
                description: fn.description || `${fn.name} - ${fn.author}`,
                category,
                tags: fn.tags || [],
                inputSchema: { type: 'object', properties: {} },
                outputSchema: { type: 'object', properties: {} },
                trustScore: fn.trust_score ?? fn.overall_score ?? 4.0,
                usageCount: fn.popularity_score ? Math.round(fn.popularity_score * 1000) : 0,
                avgExecutionTimeMs: 0,
                icon: categoryIcons[category] || categoryIcons.default,
                color: categoryColors[category] || categoryColors.default,
              };
            });
            
            set({ libraryFunctions, isLoading: false });
            toast({ title: `Loaded ${libraryFunctions.length} functions from library` });
          } catch (err) {
            const message = err instanceof Error ? err.message : 'Failed to fetch library functions';
            set({
              error: message,
              isLoading: false,
              libraryFunctions: [],
            });
            toast({ title: message, variant: 'destructive' });
          }
        },

        // Graph load/save
        loadGraph: async (author: string, name: string, version?: string) => {
          set({ isLoading: true, error: null });
          try {
            const graph = await frgApi.getGraph(author, name, version);

            const flowNodes: FRGNode[] = graph.nodeRefs.map((ref: GraphNodeRef) => ({
              id: ref.nodeId,
              type: 'functionNode',
              position: { x: 0, y: 0 },
              data: {
                functionRef: ref,
                isSelected: false,
                isEditable: true,
              },
            }));

            const flowEdges: FRGEdge[] = graph.edges.map((edge: GraphEdgeDefinition) => ({
              id: edge.id,
              source: edge.sourceNodeId,
              target: edge.targetNodeId,
              type: 'custom',
              data: {
                mapping: edge.mapping,
                condition: edge.condition,
                retryPolicy: edge.retryPolicy,
                isValid: true,
                runtimeState: {
                  status: 'idle' as const,
                  recordsTransferred: 0,
                  bytesTransferred: 0,
                  isDataFlowing: false,
                  flowProgress: 0,
                },
              },
            }));

            set({
              definition: graph,
              nodes: flowNodes,
              edges: flowEdges,
              graphAuthor: graph.author,
              graphName: graph.name,
              graphVersion: graph.version,
              isLoading: false,
              isDirty: false,
            });
            toast({ title: `Graph "${graph.name}" loaded successfully` });
          } catch (err) {
            const message = err instanceof Error ? err.message : 'Failed to load graph';
            set({
              error: message,
              isLoading: false,
            });
            toast({ title: message, variant: 'destructive' });
          }
        },

        saveGraph: async () => {
          const { graphAuthor, graphName, definition, nodes, edges, isOffline, queueOperation } = get();
          if (!graphAuthor || !graphName) {
            toast({ title: 'No graph loaded to save', variant: 'destructive' });
            return;
          }

          if (nodes.length === 0) {
            toast({ title: 'Cannot save an empty graph. Add at least one function node.', variant: 'destructive' });
            return;
          }

          set({ isSaving: true, error: null });

          if (isOffline) {
            queueOperation('saveGraph', { graphAuthor, graphName, definition, nodes, edges });
            set({ isSaving: false });
            toast({ title: 'Save queued for when you reconnect' });
            return;
          }

          try {
            const nodeRefs: GraphNodeRef[] = nodes.map((node) => ({
              nodeId: node.id,
              author: node.data.functionRef?.author || '',
              name: node.data.functionRef?.name || '',
              version: node.data.functionRef?.version || 'latest',
              config: node.data.functionRef?.config || {},
              metadata: node.data.functionRef?.metadata || {},
            }));

            const graphEdges: GraphEdgeDefinition[] = edges.map((edge) => ({
              id: edge.id,
              sourceNodeId: edge.source,
              targetNodeId: edge.target,
              mapping: edge.data?.mapping || { sourcePath: '*', targetPath: '*' },
              condition: edge.data?.condition,
              type: 'sync',
              retryPolicy: edge.data?.retryPolicy,
            }));

            if (definition?.id) {
              const updated = await frgApi.updateGraph(graphAuthor, graphName, {
                nodes: nodeRefs,
                edges: graphEdges,
              });
              set({ definition: updated, isDirty: false, isSaving: false, lastSavedAt: new Date().toISOString() });
              toast({ title: `Graph "${graphName}" saved successfully` });
            } else {
              const created = await frgApi.createGraph({
                name: graphName,
                nodes: nodeRefs,
                edges: graphEdges,
                executionMode: definition?.executionMode || 'sync',
              });
              set({
                definition: created,
                graphAuthor: created.author,
                graphName: created.name,
                graphVersion: created.version,
                isDirty: false,
                isSaving: false,
                lastSavedAt: new Date().toISOString(),
              });
              toast({ title: `Graph "${graphName}" created successfully` });
            }
          } catch (err) {
            const message = err instanceof Error ? err.message : 'Failed to save graph';
            set({
              error: message,
              isSaving: false,
            });
            toast({ title: message, variant: 'destructive' });
          }
        },

        createNewGraph: (name: string, executionMode = 'sync') => {
          set({
            definition: {
              id: '',
              author: '',
              name,
              version: 'v1',
              fullName: name,
              nodeRefs: [],
              edges: [],
              executionMode: executionMode as GraphDefinition['executionMode'],
              visibility: 'private',
              compositionScore: 0,
              trustScore: 0,
              deterministic: false,
              pricingType: 'free',
              basePrice: 0,
              revenueShare: 80,
              createdAt: new Date().toISOString(),
              updatedAt: new Date().toISOString(),
            },
            nodes: [],
            edges: [],
            graphAuthor: 'local',
            graphName: name,
            graphVersion: 'v1',
            isDirty: true,
            isLoading: false,
            error: null,
          });
        },

        // Execution controls
        startExecution: async (input?: Record<string, unknown>) => {
          const { graphAuthor, graphName, nodes, isOffline, queueOperation } = get();
          if (!graphAuthor || !graphName) {
            toast({ title: 'No graph loaded to execute', variant: 'destructive' });
            return;
          }

          if (nodes.length === 0) {
            toast({ title: 'Cannot execute an empty graph', variant: 'destructive' });
            return;
          }

          if (isOffline) {
            queueOperation('executeGraph', { graphAuthor, graphName, input });
            toast({ title: 'Execution queued for when you reconnect' });
            return;
          }

          set({
            isExecuting: true,
            executionStatus: 'running',
            executionProgress: 0,
            events: [],
            nodeRuntimeStates: {},
            edgeRuntimeStates: {},
            error: null,
          });
          toast({ title: 'Graph execution started' });

          try {
            const result = await frgApi.executeGraph(graphAuthor, graphName, input);

            if (result.instanceId) {
              set({
                activeInstance: {
                  id: result.instanceId,
                  definitionId: '',
                  status: result.status as GraphInstance['status'],
                  inputData: input,
                  outputData: result.output,
                  errorMessage: result.error || undefined,
                  totalDurationMs: result.durationMs || 0,
                  createdAt: new Date().toISOString(),
                } as GraphInstance,
              });

              if (result.output) {
                set({
                  isExecuting: false,
                  executionStatus: 'completed',
                  executionProgress: 100,
                  executionResult: {
                    instanceId: result.instanceId,
                    status: 'completed',
                    output: result.output,
                    nodeResults: result.nodeResults as Record<string, { status: NodeExecutionStatus; output?: unknown; error?: string; durationMs: number }> || {},
                    durationMs: result.durationMs || 0,
                    computeUnits: 0,
                  },
                });
                toast({ title: 'Graph execution completed' });
              } else {
                get().pollInstanceStatus();
              }
            }
          } catch (err) {
            const message = err instanceof Error ? err.message : 'Execution failed';
            set({
              isExecuting: false,
              executionStatus: 'failed',
              error: message,
            });
            toast({ title: message, variant: 'destructive' });
          }
        },

        pollInstanceStatus: async () => {
          const { activeInstance } = get();
          if (!activeInstance?.id) return;

          try {
            const instance = await frgApi.getInstanceStatus(activeInstance.id);
            set({ activeInstance: instance });

            // Update progress based on status
            if (instance.status === 'running') {
              // Continue polling every 500ms
              setTimeout(() => get().pollInstanceStatus(), 500);
            } else if (instance.status === 'completed') {
              set({ isExecuting: false, executionStatus: 'completed', executionProgress: 100 });
            } else if (instance.status === 'failed') {
              set({
                isExecuting: false,
                executionStatus: 'failed',
                error: instance.errorMessage || 'Execution failed',
              });
            }
          } catch (err) {
            console.error('Failed to poll instance status:', err);
          }
        },

        pauseExecution: () => {
          const { activeInstance } = get();
          if (activeInstance?.id) {
            set({ executionStatus: 'paused' });
          }
        },

        runNode: async (nodeId: string, input?: Record<string, unknown>) => {
          const { nodes, setNodeRuntimeState, graphAuthor } = get();
          const node = nodes.find(n => n.id === nodeId);
          if (!node?.data?.functionRef) {
            toast({ title: 'Node not found or has no function reference', variant: 'destructive' });
            return;
          }

          const fn = node.data.functionRef;
          setNodeRuntimeState(nodeId, { status: 'executing', isActive: true });
          toast({ title: `Executing node: ${fn.name}` });

          try {
            let fnExists = false;
            try {
              await registryApi.getFunction(fn.author, fn.name);
              fnExists = true;
            } catch (err: any) {
              if (err?.response?.status === 404) {
                fnExists = false;
              } else {
                throw err;
              }
            }

            if (!fnExists) {
              console.log('[runNode] Function not found, generating code for:', fn.author, fn.name);
              let generated: Record<string, unknown>;
              try {
                generated = await apiClient.post<Record<string, unknown>>('/frg/functions/generate', {
                  author: fn.author,
                  name: fn.name,
                  description: fn.metadata?.description || fn.name,
                  runtime: 'python',
                });
              } catch (err: unknown) {
                const axiosErr = err as { response?: { data?: unknown }; message?: string };
                console.error('[runNode] Generate failed:', axiosErr?.response?.data || axiosErr?.message);
                throw new Error(`Failed to generate function code: ${axiosErr?.response?.data || axiosErr?.message}`);
              }
              console.log('[runNode] Generated function:', generated);

              const generatedCode = ((generated?.version as Record<string, unknown> | undefined)?.SourceCode as string | undefined) || (generated?.code as string | undefined);
              if (generatedCode) {
                setNodeRuntimeState(nodeId, { generatedCode });
              }
            }

            const result = await frgApi.executeNode(
              fn.author,
              fn.name,
              fn.version || 'latest',
              input
            );

            const status = result.error ? 'failed' : 'completed';
            setNodeRuntimeState(nodeId, {
              status,
              output: result.output ?? result.result ?? JSON.stringify(result),
              error: result.error,
              durationMs: result.durationMs,
              isActive: false,
            });

            if (result.error) {
              toast({ title: `Node execution failed: ${result.error}`, variant: 'destructive' });
            } else {
              toast({ title: 'Node execution completed' });
            }
          } catch (err) {
            const message = err instanceof Error ? err.message : 'Execution failed';
            setNodeRuntimeState(nodeId, {
              status: 'failed',
              error: message,
              isActive: false,
            });
            toast({ title: message, variant: 'destructive' });
          }
        },

        resumeExecution: () => {
          const { activeInstance } = get();
          if (activeInstance?.id) {
            // TODO: Call resume API if available
            set({ executionStatus: 'running' });
            get().pollInstanceStatus();
          }
        },

        stopExecution: async () => {
          const { activeInstance } = get();
          if (activeInstance?.id) {
            try {
              await frgApi.stopInstance(activeInstance.id);
            } catch (err) {
              console.error('Failed to stop instance:', err);
            }
          }
          set({
            executionStatus: 'idle',
            executionProgress: 0,
            activeInstance: null,
          });
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
        setIsSaving: (saving) => set({ isSaving: saving }),
        setIsExecuting: (executing) => set({ isExecuting: executing }),
        setError: (error) => set({ error }),

        // Auto-save
        setAutoSaveEnabled: (enabled) => set({ autoSaveEnabled: enabled }),
        setAutoSaveInterval: (ms) => set({ autoSaveInterval: ms }),

        // Offline support
        setIsOffline: (offline) => set({ isOffline: offline }),
        queueOperation: (type, payload) => set((state) => ({
          operationQueue: [
            ...state.operationQueue,
            {
              id: `op-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`,
              type,
              payload,
              timestamp: new Date().toISOString(),
            },
          ],
        })),
        processQueue: async () => {
          const { operationQueue, setIsOffline } = get();
          if (operationQueue.length === 0) return;

          setIsOffline(true);
          // Process queued operations when back online
          // Each operation will be replayed via its respective API call
        },
        clearQueue: () => set({ operationQueue: [], queuedOperationCount: 0 }),
        
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
          // Optimizations are considered "approved" - in a full impl, track this
        },

        rejectEvolutionSuggestion: (id) => {
          set((state) => ({
            evolutionSuggestions: state.evolutionSuggestions.filter(s => s.id !== id),
            evolutionHistory: [
              ...state.evolutionHistory,
              ...state.evolutionSuggestions.filter(s => s.id === id).map(s => ({
                ...s,
                status: 'rejected' as const,
              })),
            ],
          }));
        },
        
        toggleEvolutionMode: async (enabled) => {
          const definition = get().definition;
          if (!definition?.author || !definition?.name) return;

          set({ isEvolutionLoading: true, evolutionError: null });
          try {
            set((state) => ({
              evolutionStatus: state.evolutionStatus
                ? { ...state.evolutionStatus, evolutionEnabled: enabled }
                : {
                    agentId: definition.id || '',
                    evolutionEnabled: enabled,
                    pendingCount: 0,
                    implementedCount: 0,
                    canEvolve: enabled,
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
          if (!definition?.author || !definition?.name) return;

          try {
            const optimizations = await frgApi.getOptimizations(definition.author, definition.name);
            set({
              evolutionStatus: {
                agentId: definition.id || '',
                evolutionEnabled: optimizations.length > 0,
                pendingCount: optimizations.filter(o => !o.applied && !o.dismissed).length,
                implementedCount: optimizations.filter(o => o.applied).length,
                canEvolve: optimizations.length > 0,
              },
            });
          } catch (err) {
            console.error('Failed to fetch evolution status:', err);
          }
        },

        fetchEvolutionSuggestions: async () => {
          const definition = get().definition;
          if (!definition?.author || !definition?.name) return;

          set({ isEvolutionLoading: true });
          try {
            const optimizations = await frgApi.getOptimizations(definition.author, definition.name);
            set({
              evolutionSuggestions: optimizations
                .filter(o => !o.applied && !o.dismissed)
                .map(o => ({
                  id: o.id,
                  type: o.suggestionType,
                  status: 'pending' as const,
                  data: o.actionConfig || {},
                  description: o.description,
                  expectedImpact: o.estimatedImpact || 0,
                  confidence: (o.aiConfidence >= 0.8 ? 'high' : o.aiConfidence >= 0.5 ? 'medium' : 'low') as 'low' | 'medium' | 'high',
                  createdAt: o.generatedAt || new Date().toISOString(),
                })),
            });
          } catch (err) {
            set({ evolutionError: err instanceof Error ? err.message : 'Unknown error' });
          } finally {
            set({ isEvolutionLoading: false });
          }
        },

        fetchEvolutionHistory: async () => {
          const definition = get().definition;
          if (!definition?.author || !definition?.name) return;

          try {
            const optimizations = await frgApi.getOptimizations(definition.author, definition.name);
            set({
              evolutionHistory: optimizations
                .filter(o => o.applied || o.dismissed)
                .map(o => ({
                  id: o.id,
                  type: o.suggestionType,
                  status: (o.applied ? 'implemented' : 'rejected') as 'implemented' | 'rejected',
                  data: o.actionConfig || {},
                  description: o.description,
                  expectedImpact: o.estimatedImpact || 0,
                  confidence: (o.aiConfidence >= 0.8 ? 'high' : o.aiConfidence >= 0.5 ? 'medium' : 'low') as 'low' | 'medium' | 'high',
                  createdAt: o.generatedAt || new Date().toISOString(),
                  implementedAt: o.applied ? new Date().toISOString() : undefined,
                })),
            });
          } catch (err) {
            console.error('Failed to fetch evolution history:', err);
          }
        },

        triggerEvolutionAnalysis: async () => {
          set({ isEvolutionLoading: true, evolutionError: null });
          // Trigger optimization analysis via the existing optimizations endpoint
          // The backend will compute new optimizations if the graph has changed
          await get().fetchEvolutionSuggestions();
          set({ isEvolutionLoading: false });
        },
      }),
      {
        name: 'frg-editor-storage',
        partialize: (state) => {
          const persistedNodeIds = new Set(state.nodes.map(n => n.id));
          const filteredRuntimeStates: Record<string, RuntimeNodeState> = {};
          for (const [nodeId, runtimeState] of Object.entries(state.nodeRuntimeStates)) {
            if (persistedNodeIds.has(nodeId)) {
              filteredRuntimeStates[nodeId] = runtimeState;
            }
          }

          const SECRET_CONFIG_KEYS = ['apiKey', 'api_key', 'apiSecret', 'api_secret', 'password', 'passwd', 'secret', 'token', 'auth', 'credential', 'privateKey', 'private_key'];

          function isSecretKey(key: string): boolean {
            const lower = key.toLowerCase();
            return SECRET_CONFIG_KEYS.some(sk => lower.includes(sk.toLowerCase()));
          }

          function redactNodeConfigs(nodes: typeof state.nodes): typeof state.nodes {
            return nodes.map(node => {
              if (!node.data?.functionRef?.config) return node;
              const redactedConfig: Record<string, unknown> = {};
              for (const [key, value] of Object.entries(node.data.functionRef.config)) {
                if (isSecretKey(key)) {
                  redactedConfig[key] = '[REDACTED]';
                } else {
                  redactedConfig[key] = value;
                }
              }
              return {
                ...node,
                data: {
                  ...node.data,
                  functionRef: {
                    ...node.data.functionRef,
                    config: redactedConfig,
                  },
                },
              };
            });
          }

          return {
            definition: state.definition,
            nodes: redactNodeConfigs(state.nodes),
            edges: state.edges,
            viewport: state.viewport,
            testCases: state.testCases,
            nodeRuntimeStates: filteredRuntimeStates,
          };
        },
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
