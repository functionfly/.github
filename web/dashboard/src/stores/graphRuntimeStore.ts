/**
 * Graph Runtime Store
 * Global state management for Visual Graph Runtime components
 */

import { create } from 'zustand'
import { immer } from 'zustand/middleware/immer'
import type {
  NodeData,
  EdgeData,
  CanvasViewport,
  ViewMode,
  RuntimeStatus,
  RuntimeMetrics,
  ExecutionEvent,
  NodeType,
} from '@functionfly/ui-graph'

// ============================================================================
// Types
// ============================================================================

export interface GraphRuntimeNode extends NodeData {
  // Extended with runtime-specific fields
  executionCount?: number
  averageExecutionTime?: number
  lastError?: string
}

export interface GraphRuntimeState {
  // Graph data
  nodes: GraphRuntimeNode[]
  edges: EdgeData[]
  
  // Viewport
  viewport: CanvasViewport
  viewMode: ViewMode
  
  // Selection
  selectedNodeIds: string[]
  selectedEdgeIds: string[]
  
  // Runtime execution
  runtimeStatus: RuntimeStatus
  runtimeMetrics: RuntimeMetrics | null
  executionEvents: ExecutionEvent[]
  
  // Panels
  isNodeInspectorOpen: boolean
  inspectorNodeId: string | null
  isMinimapVisible: boolean
  isExecutionTimelineVisible: boolean
  
  // Replay
  isReplaying: boolean
  replaySpeed: number
  replayCurrentTime: number
  replayDuration: number
  
  // Fork management
  forks: Array<{
    id: string
    nodeId: string
    label?: string
    branchCount: number
    activeBranch?: string
  }>
  
  // Active branch (for conditional rendering)
  activeBranchId: string | null
  
  // Loop state
  loopIterations: Array<{
    id: string
    index: number
    startTime: number
    endTime?: number
    duration?: number
    status: 'running' | 'completed' | 'failed' | 'skipped'
    nodeId: string
  }>
  maxLoopIterations: number
  
  // Version history
  nodeVersions: Record<string, Array<{
    version: string
    timestamp: number
    author?: string
    changes?: string
    isActive?: boolean
  }>>
  
  // Resource consumption
  resourceMetrics: Array<{
    nodeId: string
    cpuPercent?: number
    memoryMB?: number
    networkMB?: number
    storageMB?: number
    gpuPercent?: number
  }>
  
  // Distributed runtime
  runtimeRegions: Array<{
    id: string
    label: string
    location: string
    nodeCount: number
    healthyNodes: number
    unhealthyNodes: number
    latency?: number
    status: 'healthy' | 'degraded' | 'unhealthy'
    nodes: NodeData[]
  }>
  selectedRegionId: string | null
  
  // Failure propagation
  failurePropagation: Array<{
    id: string
    nodeId: string
    label: string
    type: NodeType
    failureReason?: string
    affectedNodes: string[]
    depth: number
    isRootCause?: boolean
    propagationPath: string[]
  }>
  rootFailure: { nodeId: string; label: string; reason: string } | null
  
  // Actions - Graph Data
  setNodes: (nodes: GraphRuntimeNode[]) => void
  addNode: (node: GraphRuntimeNode) => void
  updateNode: (id: string, updates: Partial<GraphRuntimeNode>) => void
  removeNode: (id: string) => void
  setEdges: (edges: EdgeData[]) => void
  addEdge: (edge: EdgeData) => void
  removeEdge: (id: string) => void
  
  // Actions - Viewport
  setViewport: (viewport: CanvasViewport) => void
  setViewMode: (mode: ViewMode) => void
  zoomIn: () => void
  zoomOut: () => void
  fitView: () => void
  resetView: () => void
  
  // Actions - Selection
  selectNode: (id: string, append?: boolean) => void
  selectNodes: (ids: string[]) => void
  deselectAll: () => void
  selectEdge: (id: string) => void
  
  // Actions - Runtime
  setRuntimeStatus: (status: RuntimeStatus) => void
  setRuntimeMetrics: (metrics: RuntimeMetrics | null) => void
  startExecution: () => void
  pauseExecution: () => void
  stopExecution: () => void
  resetExecution: () => void
  
  // Actions - Execution Events
  addExecutionEvent: (event: ExecutionEvent) => void
  setExecutionEvents: (events: ExecutionEvent[]) => void
  clearExecutionEvents: () => void
  
  // Actions - Panels
  openNodeInspector: (nodeId: string) => void
  closeNodeInspector: () => void
  toggleMinimap: () => void
  toggleExecutionTimeline: () => void
  
  // Actions - Replay
  setIsReplaying: (replaying: boolean) => void
  setReplaySpeed: (speed: number) => void
  setReplayCurrentTime: (time: number) => void
  setReplayDuration: (duration: number) => void
  seekReplay: (time: number) => void
  
  // Actions - Forks
  addFork: (fork: { nodeId: string; label?: string; branchCount: number }) => void
  removeFork: (id: string) => void
  setActiveBranch: (forkId: string, branchId: string) => void
  
  // Actions - Loops
  addLoopIteration: (iteration: Omit<GraphRuntimeState['loopIterations'][0], 'id'>) => void
  updateLoopIteration: (id: string, updates: Partial<GraphRuntimeState['loopIterations'][0]>) => void
  setMaxLoopIterations: (max: number) => void
  
  // Actions - Versions
  addNodeVersion: (nodeId: string, version: Omit<GraphRuntimeState['nodeVersions'][string][0], 'id'>) => void
  restoreNodeVersion: (nodeId: string, version: string) => void
  
  // Actions - Resources
  updateResourceMetric: (nodeId: string, metric: Partial<GraphRuntimeState['resourceMetrics'][0]>) => void
  setResourceMetrics: (metrics: GraphRuntimeState['resourceMetrics']) => void
  
  // Actions - Regions
  setRuntimeRegions: (regions: GraphRuntimeState['runtimeRegions']) => void
  selectRuntimeRegion: (id: string | null) => void
  
  // Actions - Failures
  setFailurePropagation: (propagation: GraphRuntimeState['failurePropagation'], rootFailure: GraphRuntimeState['rootFailure']) => void
  clearFailurePropagation: () => void
}

// ============================================================================
// Initial State & Store
// ============================================================================

export const useGraphRuntimeStore = create<GraphRuntimeState>()(
  immer((set, get) => ({
    // Initial state
    nodes: [],
    edges: [],
    viewport: { x: 0, y: 0, zoom: 1 },
    viewMode: 'design',
    selectedNodeIds: [],
    selectedEdgeIds: [],
    runtimeStatus: 'idle',
    runtimeMetrics: null,
    executionEvents: [],
    isNodeInspectorOpen: false,
    inspectorNodeId: null,
    isMinimapVisible: true,
    isExecutionTimelineVisible: true,
    isReplaying: false,
    replaySpeed: 1,
    replayCurrentTime: 0,
    replayDuration: 0,
    forks: [],
    activeBranchId: null,
    loopIterations: [],
    maxLoopIterations: 10,
    nodeVersions: {},
    resourceMetrics: [],
    runtimeRegions: [],
    selectedRegionId: null,
    failurePropagation: [],
    rootFailure: null,

    // Actions - Graph Data
    setNodes: (nodes) => set((state) => { state.nodes = nodes }),
    addNode: (node) => set((state) => { state.nodes.push(node) }),
    updateNode: (id, updates) => set((state) => {
      const idx = state.nodes.findIndex(n => n.id === id)
      if (idx !== -1) Object.assign(state.nodes[idx], updates)
    }),
    removeNode: (id) => set((state) => {
      state.nodes = state.nodes.filter(n => n.id !== id)
      state.edges = state.edges.filter(e => e.source !== id && e.target !== id)
      state.selectedNodeIds = state.selectedNodeIds.filter(nid => nid !== id)
    }),
    setEdges: (edges) => set((state) => { state.edges = edges }),
    addEdge: (edge) => set((state) => { state.edges.push(edge) }),
    removeEdge: (id) => set((state) => {
      state.edges = state.edges.filter(e => e.id !== id)
      state.selectedEdgeIds = state.selectedEdgeIds.filter(eid => eid !== id)
    }),

    // Actions - Viewport
    setViewport: (viewport) => set((state) => { state.viewport = viewport }),
    setViewMode: (mode) => set((state) => { state.viewMode = mode }),
    zoomIn: () => set((state) => {
      state.viewport.zoom = Math.min(state.viewport.zoom * 1.2, 4)
    }),
    zoomOut: () => set((state) => {
      state.viewport.zoom = Math.max(state.viewport.zoom / 1.2, 0.2)
    }),
    fitView: () => {
      const { nodes } = get()
      if (nodes.length === 0) return
      const bounds = nodes.reduce((acc, n) => {
        if (!n.position) return acc
        return {
          minX: Math.min(acc.minX, n.position.x),
          minY: Math.min(acc.minY, n.position.y),
          maxX: Math.max(acc.maxX, n.position.x + 200),
          maxY: Math.max(acc.maxY, n.position.y + 120),
        }
      }, { minX: Infinity, minY: Infinity, maxX: -Infinity, maxY: -Infinity })
      set((state) => {
        state.viewport = {
          x: -(bounds.minX + (bounds.maxX - bounds.minX) / 2),
          y: -(bounds.minY + (bounds.maxY - bounds.minY) / 2),
          zoom: 1,
        }
      })
    },
    resetView: () => set((state) => {
      state.viewport = { x: 0, y: 0, zoom: 1 }
    }),

    // Actions - Selection
    selectNode: (id, append = false) => set((state) => {
      if (append) {
        if (!state.selectedNodeIds.includes(id)) {
          state.selectedNodeIds.push(id)
        }
      } else {
        state.selectedNodeIds = [id]
        state.selectedEdgeIds = []
      }
    }),
    selectNodes: (ids) => set((state) => {
      state.selectedNodeIds = ids
      state.selectedEdgeIds = []
    }),
    deselectAll: () => set((state) => {
      state.selectedNodeIds = []
      state.selectedEdgeIds = []
    }),
    selectEdge: (id) => set((state) => {
      state.selectedEdgeIds = [id]
      state.selectedNodeIds = []
    }),

    // Actions - Runtime
    setRuntimeStatus: (status) => set((state) => { state.runtimeStatus = status }),
    setRuntimeMetrics: (metrics) => set((state) => { state.runtimeMetrics = metrics }),
    startExecution: () => set((state) => {
      state.runtimeStatus = 'running'
      state.executionEvents = []
      state.loopIterations = []
    }),
    pauseExecution: () => set((state) => {
      if (state.runtimeStatus === 'running') {
        state.runtimeStatus = 'paused'
      }
    }),
    stopExecution: () => set((state) => {
      state.runtimeStatus = 'idle'
    }),
    resetExecution: () => set((state) => {
      state.runtimeStatus = 'idle'
      state.executionEvents = []
      state.loopIterations = []
      state.replayCurrentTime = 0
      state.replayDuration = 0
    }),

    // Actions - Execution Events
    addExecutionEvent: (event) => set((state) => {
      state.executionEvents.push(event)
    }),
    setExecutionEvents: (events) => set((state) => {
      state.executionEvents = events
    }),
    clearExecutionEvents: () => set((state) => {
      state.executionEvents = []
    }),

    // Actions - Panels
    openNodeInspector: (nodeId) => set((state) => {
      state.isNodeInspectorOpen = true
      state.inspectorNodeId = nodeId
    }),
    closeNodeInspector: () => set((state) => {
      state.isNodeInspectorOpen = false
      state.inspectorNodeId = null
    }),
    toggleMinimap: () => set((state) => {
      state.isMinimapVisible = !state.isMinimapVisible
    }),
    toggleExecutionTimeline: () => set((state) => {
      state.isExecutionTimelineVisible = !state.isExecutionTimelineVisible
    }),

    // Actions - Replay
    setIsReplaying: (replaying) => set((state) => {
      state.isReplaying = replaying
    }),
    setReplaySpeed: (speed) => set((state) => {
      state.replaySpeed = speed
    }),
    setReplayCurrentTime: (time) => set((state) => {
      state.replayCurrentTime = time
    }),
    setReplayDuration: (duration) => set((state) => {
      state.replayDuration = duration
    }),
    seekReplay: (time) => set((state) => {
      state.replayCurrentTime = Math.max(0, Math.min(time, state.replayDuration))
    }),

    // Actions - Forks
    addFork: (fork) => set((state) => {
      state.forks.push({
        id: `fork-${Date.now()}`,
        nodeId: fork.nodeId,
        label: fork.label,
        branchCount: fork.branchCount,
      })
    }),
    removeFork: (id) => set((state) => {
      state.forks = state.forks.filter(f => f.id !== id)
    }),
    setActiveBranch: (forkId, branchId) => set((state) => {
      const fork = state.forks.find(f => f.id === forkId)
      if (fork) fork.activeBranch = branchId
      state.activeBranchId = branchId
    }),

    // Actions - Loops
    addLoopIteration: (iteration) => set((state) => {
      state.loopIterations.push({
        ...iteration,
        id: `iter-${Date.now()}`,
      })
    }),
    updateLoopIteration: (id, updates) => set((state) => {
      const iter = state.loopIterations.find(i => i.id === id)
      if (iter) Object.assign(iter, updates)
    }),
    setMaxLoopIterations: (max) => set((state) => {
      state.maxLoopIterations = max
    }),

    // Actions - Versions
    addNodeVersion: (nodeId, version) => set((state) => {
      if (!state.nodeVersions[nodeId]) {
        state.nodeVersions[nodeId] = []
      }
      state.nodeVersions[nodeId].push({
        ...version,
        id: `v-${Date.now()}`,
      } as any)
    }),
    restoreNodeVersion: (nodeId, version) => set((state) => {
      const versions = state.nodeVersions[nodeId]
      if (!versions) return
      const idx = versions.findIndex(v => v.version === version)
      if (idx !== -1) {
        versions.forEach((v, i) => { v.isActive = i === idx })
      }
    }),

    // Actions - Resources
    updateResourceMetric: (nodeId, metric) => set((state) => {
      const idx = state.resourceMetrics.findIndex(m => m.nodeId === nodeId)
      if (idx !== -1) {
        Object.assign(state.resourceMetrics[idx], metric)
      } else {
        state.resourceMetrics.push({ nodeId, ...metric })
      }
    }),
    setResourceMetrics: (metrics) => set((state) => {
      state.resourceMetrics = metrics
    }),

    // Actions - Regions
    setRuntimeRegions: (regions) => set((state) => {
      state.runtimeRegions = regions
    }),
    selectRuntimeRegion: (id) => set((state) => {
      state.selectedRegionId = id
    }),

    // Actions - Failures
    setFailurePropagation: (propagation, rootFailure) => set((state) => {
      state.failurePropagation = propagation
      state.rootFailure = rootFailure
    }),
    clearFailurePropagation: () => set((state) => {
      state.failurePropagation = []
      state.rootFailure = null
    }),
  }))
)
