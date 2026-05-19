/**
 * Graph Runtime Hook
 * Provides unified access to Graph Runtime components and their state
 */

import { useCallback, useMemo } from 'react'
import { useGraphRuntimeStore } from '@/stores/graphRuntimeStore'
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
// useGraphRuntime
// ============================================================================

export function useGraphRuntime() {
  const store = useGraphRuntimeStore()

  // ---- Graph Data ----
  const setNodes = useCallback((nodes: NodeData[]) => {
    store.setNodes(nodes as any[])
  }, [store])

  const addNode = useCallback((node: NodeData) => {
    store.addNode(node as any)
  }, [store])

  const updateNode = useCallback((id: string, updates: Partial<NodeData>) => {
    store.updateNode(id, updates as any)
  }, [store])

  const removeNode = useCallback((id: string) => {
    store.removeNode(id)
  }, [store])

  const setEdges = useCallback((edges: EdgeData[]) => {
    store.setEdges(edges)
  }, [store])

  const addEdge = useCallback((edge: EdgeData) => {
    store.addEdge(edge)
  }, [store])

  const removeEdge = useCallback((id: string) => {
    store.removeEdge(id)
  }, [store])

  // ---- Viewport ----
  const setViewport = useCallback((viewport: CanvasViewport) => {
    store.setViewport(viewport)
  }, [store])

  const setViewMode = useCallback((mode: ViewMode) => {
    store.setViewMode(mode)
  }, [store])

  const zoomIn = useCallback(() => store.zoomIn(), [store])
  const zoomOut = useCallback(() => store.zoomOut(), [store])
  const fitView = useCallback(() => store.fitView(), [store])
  const resetView = useCallback(() => store.resetView(), [store])

  // ---- Selection ----
  const selectNode = useCallback((id: string, append = false) => {
    store.selectNode(id, append)
  }, [store])

  const deselectAll = useCallback(() => store.deselectAll(), [store])

  // ---- Runtime Execution ----
  const startExecution = useCallback(() => {
    store.startExecution()
  }, [store])

  const pauseExecution = useCallback(() => {
    store.pauseExecution()
  }, [store])

  const stopExecution = useCallback(() => {
    store.stopExecution()
  }, [store])

  const resetExecution = useCallback(() => {
    store.resetExecution()
  }, [store])

  const setRuntimeStatus = useCallback((status: RuntimeStatus) => {
    store.setRuntimeStatus(status)
  }, [store])

  const setRuntimeMetrics = useCallback((metrics: RuntimeMetrics | null) => {
    store.setRuntimeMetrics(metrics)
  }, [store])

  // ---- Execution Events ----
  const addExecutionEvent = useCallback((event: ExecutionEvent) => {
    store.addExecutionEvent(event)
  }, [store])

  const clearExecutionEvents = useCallback(() => {
    store.clearExecutionEvents()
  }, [store])

  // ---- Panels ----
  const openNodeInspector = useCallback((nodeId: string) => {
    store.openNodeInspector(nodeId)
  }, [store])

  const closeNodeInspector = useCallback(() => {
    store.closeNodeInspector()
  }, [store])

  const toggleMinimap = useCallback(() => {
    store.toggleMinimap()
  }, [store])

  const toggleExecutionTimeline = useCallback(() => {
    store.toggleExecutionTimeline()
  }, [store])

  // ---- Replay ----
  const setIsReplaying = useCallback((replaying: boolean) => {
    store.setIsReplaying(replaying)
  }, [store])

  const setReplaySpeed = useCallback((speed: number) => {
    store.setReplaySpeed(speed)
  }, [store])

  const seekReplay = useCallback((time: number) => {
    store.seekReplay(time)
  }, [store])

  // ---- Forks ----
  const addFork = useCallback((fork: { nodeId: string; label?: string; branchCount: number }) => {
    store.addFork(fork)
  }, [store])

  const removeFork = useCallback((id: string) => {
    store.removeFork(id)
  }, [store])

  const setActiveBranch = useCallback((forkId: string, branchId: string) => {
    store.setActiveBranch(forkId, branchId)
  }, [store])

  // ---- Loops ----
  const addLoopIteration = useCallback((iteration: Omit<ReturnType<typeof store.addLoopIteration>, 'id'> & { nodeId: string; index: number; startTime: number; status: 'running' | 'completed' | 'failed' | 'skipped' }) => {
    store.addLoopIteration({
      id: `iter-${Date.now()}`,
      ...iteration,
    })
  }, [store])

  const setMaxLoopIterations = useCallback((max: number) => {
    store.setMaxLoopIterations(max)
  }, [store])

  return {
    // State
    ...store,

    // Graph Data
    setNodes,
    addNode,
    updateNode,
    removeNode,
    setEdges,
    addEdge,
    removeEdge,

    // Viewport
    setViewport,
    setViewMode,
    zoomIn,
    zoomOut,
    fitView,
    resetView,

    // Selection
    selectNode,
    deselectAll,

    // Runtime
    startExecution,
    pauseExecution,
    stopExecution,
    resetExecution,
    setRuntimeStatus,
    setRuntimeMetrics,

    // Events
    addExecutionEvent,
    clearExecutionEvents,

    // Panels
    openNodeInspector,
    closeNodeInspector,
    toggleMinimap,
    toggleExecutionTimeline,

    // Replay
    setIsReplaying,
    setReplaySpeed,
    seekReplay,

    // Forks
    addFork,
    removeFork,
    setActiveBranch,

    // Loops
    addLoopIteration,
    setMaxLoopIterations,
  }
}

// ============================================================================
// useNodeSelection
// ============================================================================

export function useNodeSelection() {
  const store = useGraphRuntimeStore()

  const selectedNodes = useMemo(() =>
    store.nodes.filter(n => store.selectedNodeIds.includes(n.id))
  , [store.nodes, store.selectedNodeIds])

  const selectedNodeCount = store.selectedNodeIds.length
  const hasSelection = selectedNodeCount > 0

  return {
    selectedNodes,
    selectedNodeIds: store.selectedNodeIds,
    selectedNodeCount,
    hasSelection,
    isNodeSelected: (id: string) => store.selectedNodeIds.includes(id),
    selectNode: store.selectNode,
    selectNodes: store.selectNodes,
    deselectAll: store.deselectAll,
  }
}

// ============================================================================
// useExecutionReplay
// ============================================================================

export function useExecutionReplay() {
  const store = useGraphRuntimeStore()

  const progress = useMemo(() => {
    if (store.replayDuration === 0) return 0
    return (store.replayCurrentTime / store.replayDuration) * 100
  }, [store.replayCurrentTime, store.replayDuration])

  const formattedTime = useMemo(() => {
    const formatMs = (ms: number) => {
      const totalSeconds = Math.floor(ms / 1000)
      const minutes = Math.floor(totalSeconds / 60)
      const seconds = totalSeconds % 60
      return `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`
    }
    return {
      current: formatMs(store.replayCurrentTime),
      duration: formatMs(store.replayDuration),
    }
  }, [store.replayCurrentTime, store.replayDuration])

  return {
    isReplaying: store.isReplaying,
    isPaused: store.runtimeStatus === 'paused',
    currentTime: store.replayCurrentTime,
    duration: store.replayDuration,
    speed: store.replaySpeed,
    progress,
    formattedTime,
    setIsReplaying: store.setIsReplaying,
    setReplaySpeed: store.setReplaySpeed,
    seekReplay: store.seekReplay,
    play: () => store.setIsReplaying(true),
    pause: () => store.setIsReplaying(false),
    stop: () => {
      store.setIsReplaying(false)
      store.seekReplay(0)
    },
  }
}

// ============================================================================
// useRuntimeMetrics
// ============================================================================

export function useRuntimeMetrics() {
  const store = useGraphRuntimeStore()

  const statusColor = useMemo(() => {
    const colors: Record<RuntimeStatus, string> = {
      idle: '#6b7280',
      starting: '#3b82f6',
      running: '#10b981',
      paused: '#f59e0b',
      completed: '#22c55e',
      failed: '#ef4444',
      cancelled: '#a855f7',
    }
    return colors[store.runtimeStatus] || colors.idle
  }, [store.runtimeStatus])

  const isRunning = store.runtimeStatus === 'running' || store.runtimeStatus === 'starting'
  const isPaused = store.runtimeStatus === 'paused'
  const isFinished = ['completed', 'failed', 'cancelled'].includes(store.runtimeStatus)

  return {
    status: store.runtimeStatus,
    statusColor,
    metrics: store.runtimeMetrics,
    isRunning,
    isPaused,
    isFinished,
    isIdle: store.runtimeStatus === 'idle',
  }
}

// ============================================================================
// useNodeInspector
// ============================================================================

export function useNodeInspector() {
  const store = useGraphRuntimeStore()

  const inspectedNode = useMemo(() =>
    store.nodes.find(n => n.id === store.inspectorNodeId) || null
  , [store.nodes, store.inspectorNodeId])

  return {
    isOpen: store.isNodeInspectorOpen,
    node: inspectedNode,
    openNodeInspector: store.openNodeInspector,
    closeNodeInspector: store.closeNodeInspector,
  }
}

// ============================================================================
// useResourceConsumption
// ============================================================================

export function useResourceConsumption() {
  const store = useGraphRuntimeStore()

  const aggregatedMetrics = useMemo(() => {
    if (store.resourceMetrics.length === 0) return null
    return store.resourceMetrics.reduce((acc, m) => ({
      totalCpu: acc.totalCpu + (m.cpuPercent || 0),
      totalMemory: acc.totalMemory + (m.memoryMB || 0),
      totalNetwork: acc.totalNetwork + (m.networkMB || 0),
      totalStorage: acc.totalStorage + (m.storageMB || 0),
      totalGpu: acc.totalGpu + (m.gpuPercent || 0),
    }), { totalCpu: 0, totalMemory: 0, totalNetwork: 0, totalStorage: 0, totalGpu: 0 })
  }, [store.resourceMetrics])

  return {
    metrics: store.resourceMetrics,
    aggregatedMetrics,
    updateResourceMetric: store.updateResourceMetric,
    setResourceMetrics: store.setResourceMetrics,
  }
}

// ============================================================================
// useDistributedRuntime
// ============================================================================

export function useDistributedRuntime() {
  const store = useGraphRuntimeStore()

  const selectedRegion = useMemo(() =>
    store.runtimeRegions.find(r => r.id === store.selectedRegionId) || null
  , [store.runtimeRegions, store.selectedRegionId])

  const regionStats = useMemo(() => {
    const totalNodes = store.runtimeRegions.reduce((sum, r) => sum + r.nodeCount, 0)
    const healthyNodes = store.runtimeRegions.reduce((sum, r) => sum + r.healthyNodes, 0)
    const healthyRegions = store.runtimeRegions.filter(r => r.status === 'healthy').length
    return { totalNodes, healthyNodes, healthyRegions, totalRegions: store.runtimeRegions.length }
  }, [store.runtimeRegions])

  return {
    regions: store.runtimeRegions,
    selectedRegion,
    selectedRegionId: store.selectedRegionId,
    regionStats,
    selectRegion: store.selectRuntimeRegion,
    setRegions: store.setRuntimeRegions,
  }
}

// ============================================================================
// useFailurePropagation
// ============================================================================

export function useFailurePropagation() {
  const store = useGraphRuntimeStore()

  const affectedNodeCount = store.failurePropagation.reduce(
    (sum, f) => sum + f.affectedNodes.length, 0
  )

  return {
    rootFailure: store.rootFailure,
    propagation: store.failurePropagation,
    affectedNodeCount,
    setFailurePropagation: store.setFailurePropagation,
    clearFailurePropagation: store.clearFailurePropagation,
  }
}

// ============================================================================
// useNodeVersions
// ============================================================================

export function useNodeVersions(nodeId: string | null) {
  const store = useGraphRuntimeStore()

  const versions = useMemo(() =>
    nodeId ? (store.nodeVersions[nodeId] || []) : []
  , [store.nodeVersions, nodeId])

  return {
    versions,
    addVersion: (version: { version: string; timestamp: number; author?: string; changes?: string; isActive?: boolean }) => {
      if (nodeId) store.addNodeVersion(nodeId, version)
    },
    restoreVersion: (version: string) => {
      if (nodeId) store.restoreNodeVersion(nodeId, version)
    },
  }
}

// ============================================================================
// useLoopVisualization
// ============================================================================

export function useLoopVisualization() {
  const store = useGraphRuntimeStore()

  const completedIterations = useMemo(() =>
    store.loopIterations.filter(i => i.status === 'completed').length
  , [store.loopIterations])

  const failedIterations = useMemo(() =>
    store.loopIterations.filter(i => i.status === 'failed').length
  , [store.loopIterations])

  const progress = useMemo(() => {
    if (store.maxLoopIterations === 0) return 0
    return (completedIterations / store.maxLoopIterations) * 100
  }, [completedIterations, store.maxLoopIterations])

  return {
    iterations: store.loopIterations,
    maxIterations: store.maxLoopIterations,
    completedIterations,
    failedIterations,
    progress,
    isRunning: store.loopIterations.some(i => i.status === 'running'),
    addIteration: store.addLoopIteration,
    updateIteration: store.updateLoopIteration,
    setMaxIterations: store.setMaxLoopIterations,
  }
}
