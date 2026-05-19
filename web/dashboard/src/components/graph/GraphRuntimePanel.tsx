/**
 * GraphRuntimePanel
 * Main integration component for Visual Graph Runtime components
 * Wires together all graph components with the store
 */

import * as React from 'react'
import { cn } from '@functionfly/ui-core'
import { useGraphRuntime } from '@/hooks/useGraphRuntime'
import { useNodeSelection, useExecutionReplay, useRuntimeMetrics, useNodeInspector, useResourceConsumption, useDistributedRuntime, useFailurePropagation, useLoopVisualization, useNodeVersions } from '@/hooks/useGraphRuntime'
import {
  RuntimeGraph,
  DynamicPort,
  TokenFlowRenderer,
  NodeInspector,
  GraphMiniMap,
  ExecutionReplayControls,
  ExecutionHeatmap,
  GraphViewport,
  WorkflowForkManager,
  LiveNodeTelemetry,
  GraphStateDiffViewer,
  DependencyExplorer,
  OrchestrationLayerView,
  ConditionalBranchRenderer,
  LoopVisualizer,
  FailurePropagationMap,
  ResourceConsumptionOverlay,
  NodeVersionHistory,
  ExecutionTraceExplorer,
  DistributedRuntimeMap,
  type NodeData,
  type EdgeData,
  type CanvasViewport,
  type ViewMode,
  type RuntimeStatus,
  type RuntimeMetrics,
  type ExecutionEvent,
  type NodeType,
  type NodeInspectorProps,
  type GraphMiniMapProps,
  type ExecutionReplayControlsProps,
  type ExecutionHeatmapProps,
  type WorkflowForkManagerProps,
  type LiveNodeTelemetryProps,
  type GraphStateDiffViewerProps,
  type DependencyExplorerProps,
  type OrchestrationLayerViewProps,
  type ConditionalBranchRendererProps,
  type LoopVisualizerProps,
  type FailurePropagationMapProps,
  type ResourceConsumptionOverlayProps,
  type NodeVersionHistoryProps,
  type ExecutionTraceExplorerProps,
  type DistributedRuntimeMapProps,
  type RuntimeGraphProps,
} from '@functionfly/ui-graph'

// ============================================================================
// GraphRuntimePanel
// ============================================================================

export interface GraphRuntimePanelProps {
  className?: string
  showViewport?: boolean
  showMinimap?: boolean
  showInspector?: boolean
  showReplayControls?: boolean
  showHeatmap?: boolean
  showForkManager?: boolean
  showTelemetry?: boolean
  showDependencyExplorer?: boolean
  showLayerView?: boolean
  showBranchRenderer?: boolean
  showLoopVisualizer?: boolean
  showFailureMap?: boolean
  showResourceOverlay?: boolean
  showVersionHistory?: boolean
  showTraceExplorer?: boolean
  showDistributedMap?: boolean
  initialNodes?: NodeData[]
  initialEdges?: EdgeData[]
  initialViewMode?: ViewMode
}

export function GraphRuntimePanel({
  className,
  showViewport = true,
  showMinimap = true,
  showInspector = true,
  showReplayControls = false,
  showHeatmap = false,
  showForkManager = false,
  showTelemetry = false,
  showDependencyExplorer = false,
  showLayerView = false,
  showBranchRenderer = false,
  showLoopVisualizer = false,
  showFailureMap = false,
  showResourceOverlay = false,
  showVersionHistory = false,
  showTraceExplorer = false,
  showDistributedMap = false,
  initialNodes = [],
  initialEdges = [],
  initialViewMode = 'design',
}: GraphRuntimePanelProps) {
  const runtime = useGraphRuntime()

  // Initialize with provided nodes/edges
  React.useEffect(() => {
    if (initialNodes.length > 0) runtime.setNodes(initialNodes)
  }, [initialNodes])

  React.useEffect(() => {
    if (initialEdges.length > 0) runtime.setEdges(initialEdges)
  }, [initialEdges])

  React.useEffect(() => {
    runtime.setViewMode(initialViewMode)
  }, [initialViewMode])

  const { isRunning, isPaused, isFinished, status, statusColor, metrics } = useRuntimeMetrics()
  const { isOpen: isInspectorOpen, node: inspectedNode, openNodeInspector, closeNodeInspector } = useNodeInspector()
  const { selectedNodes, selectedNodeIds, isNodeSelected, deselectAll } = useNodeSelection()
  const { isReplaying, currentTime, duration, speed, progress, formattedTime, play, pause, stop, seekReplay, setReplaySpeed } = useExecutionReplay()
  const { metrics: resourceMetrics, aggregatedMetrics } = useResourceConsumption()
  const { regions, selectedRegion, regionStats, selectRegion } = useDistributedRuntime()
  const { rootFailure, propagation: failurePropagation, setFailurePropagation, clearFailurePropagation } = useFailurePropagation()
  const { iterations: loopIterations, maxIterations, completedIterations, progress: loopProgress, addIteration, setMaxIterations } = useLoopVisualization()

  const inspectorNode = inspectedNode as NodeData | null

  return (
    <div className={cn("flex h-full bg-[#08080f]", className)}>
      {/* Main Canvas Area */}
      <div className="flex-1 relative">
        {showViewport ? (
          <GraphViewport
            viewport={runtime.viewport}
            showGrid={true}
            gridSize={64}
            showZoomIndicator={true}
            onViewportChange={runtime.setViewport}
            onZoomIn={runtime.zoomIn}
            onZoomOut={runtime.zoomOut}
            onFitView={runtime.fitView}
            onResetView={runtime.resetView}
          >
            {/* Runtime Graph Canvas */}
            <RuntimeGraph
              nodes={runtime.nodes as RuntimeGraphProps['nodes']}
              edges={runtime.edges as RuntimeGraphProps['edges']}
              runtimeStatus={runtime.runtimeStatus}
              metrics={runtime.runtimeMetrics as RuntimeGraphProps['metrics']}
              onNodeSelect={(node) => runtime.selectNode(node.id)}
              onNodeDoubleClick={(node) => runtime.openNodeInspector(node.id)}
              onStart={runtime.startExecution}
              onPause={runtime.pauseExecution}
              onStop={runtime.stopExecution}
              onReset={runtime.resetExecution}
            />

            {/* Token Flow Overlay */}
            <div className="absolute inset-0 pointer-events-none">
              <TokenFlowRenderer
                edges={runtime.edges}
                activeTokens={3}
                tokenFlowSpeed={1}
              />
            </div>
          </GraphViewport>
        ) : (
          <RuntimeGraph
            nodes={runtime.nodes as RuntimeGraphProps['nodes']}
            edges={runtime.edges as RuntimeGraphProps['edges']}
            runtimeStatus={runtime.runtimeStatus}
            metrics={runtime.runtimeMetrics as RuntimeGraphProps['metrics']}
            onNodeSelect={(node) => runtime.selectNode(node.id)}
            onNodeDoubleClick={(node) => runtime.openNodeInspector(node.id)}
            onStart={runtime.startExecution}
            onPause={runtime.pauseExecution}
            onStop={runtime.stopExecution}
            onReset={runtime.resetExecution}
          />
        )}

        {/* Minimap */}
        {showMinimap && runtime.isMinimapVisible && (
          <GraphMiniMap
            nodes={runtime.nodes}
            edges={runtime.edges}
            viewport={runtime.viewport}
            canvasWidth={800}
            canvasHeight={600}
            onNavigate={runtime.setViewport}
          />
        )}
      </div>

      {/* Right Sidebar Panels */}
      <div className="w-80 flex flex-col gap-2 p-2 bg-[#0d0d14] border-l border-[rgba(255,255,255,0.08)] overflow-y-auto">
        {/* Replay Controls */}
        {showReplayControls && (
          <ExecutionReplayControls
            isPlaying={isReplaying}
            currentTime={runtime.replayCurrentTime}
            duration={runtime.replayDuration}
            speed={speed as ExecutionReplayControlsProps['speed']}
            onPlayPause={isReplaying ? pause : play}
            onStop={stop}
            onSeek={seekReplay}
            onSpeedChange={setReplaySpeed}
          />
        )}

        {/* Execution Heatmap */}
        {showHeatmap && (
          <ExecutionHeatmap
            cells={runtime.executionEvents.map((e) => ({
              nodeId: e.nodeId || '',
              timestamp: e.timestamp,
              intensity: 0.5,
              status: e.result === 'success' ? 'completed' : e.result === 'failure' ? 'error' : 'running',
            }))}
            timeRange={{ start: Date.now() - 60000, end: Date.now() }}
            columns={24}
            cellSize={12}
            colorScale="hot"
          />
        )}

        {/* Node Inspector */}
        {showInspector && inspectorNode && (
          <NodeInspector
            node={inspectedNode as NodeInspectorProps['node']}
            onClose={closeNodeInspector}
            onUpdate={(nodeId, data) => runtime.updateNode(nodeId, data)}
            onDelete={runtime.removeNode}
          />
        )}

        {/* Version History */}
        {showVersionHistory && inspectedNode && (
          <NodeVersionHistory
            nodeId={inspectedNode.id}
            nodeLabel={inspectedNode.label}
            versions={runtime.nodeVersions[inspectedNode.id] || []}
            onVersionSelect={(version) => {
              console.log('Select version:', version)
            }}
            onVersionRestore={(version) => {
              if (inspectedNode) runtime.restoreNodeVersion?.(inspectedNode.id, version.version)
            }}
          />
        )}

        {/* Live Node Telemetry */}
        {showTelemetry && selectedNodes.length === 1 && (
          <LiveNodeTelemetry
            nodeId={selectedNodes[0].id}
            nodeLabel={selectedNodes[0].label}
            metrics={[
              { label: 'CPU', value: 45.2, unit: '%', trend: 'up' },
              { label: 'Memory', value: 128, unit: 'MB', trend: 'stable' },
              { label: 'Latency', value: 23, unit: 'ms', trend: 'down' },
            ]}
            isExpanded={true}
            onToggleExpand={() => {}}
          />
        )}

        {/* Workflow Fork Manager */}
        {showForkManager && runtime.forks.length > 0 && (
          <WorkflowForkManager
            forks={runtime.forks as WorkflowForkManagerProps['forks']}
            onForkSelect={(forkId, branchId) => runtime.setActiveBranch(forkId, branchId)}
            onForkDelete={runtime.removeFork}
          />
        )}

        {/* Dependency Explorer */}
        {showDependencyExplorer && selectedNodes.length === 1 && (
          <DependencyExplorer
            nodes={runtime.nodes as DependencyExplorerProps['nodes']}
            edges={runtime.edges as DependencyExplorerProps['edges']}
            rootNodeId={selectedNodes[0].id}
            direction="downstream"
            onNodeClick={(nodeId) => runtime.selectNode(nodeId)}
          />
        )}

        {/* Orchestration Layer View */}
        {showLayerView && (
          <OrchestrationLayerView
            layers={[
              {
                id: 'layer-1',
                label: 'Input Processing',
                depth: 0,
                nodes: runtime.nodes.filter(n => n.type === 'trigger'),
                isParallel: false,
                hasError: false,
              },
              {
                id: 'layer-2',
                label: 'Agent Processing',
                depth: 1,
                nodes: runtime.nodes.filter(n => n.type === 'agent'),
                isParallel: true,
                hasError: false,
              },
              {
                id: 'layer-3',
                label: 'Output',
                depth: 2,
                nodes: runtime.nodes.filter(n => n.type === 'output'),
                isParallel: false,
                hasError: false,
              },
            ]}
            onLayerClick={(layer) => console.log('Layer click:', layer)}
            onLayerExpand={(layerId) => console.log('Expand layer:', layerId)}
          />
        )}

        {/* Conditional Branch Renderer */}
        {showBranchRenderer && (
          <ConditionalBranchRenderer
            branches={[
              {
                id: 'branch-1',
                label: 'Success Path',
                condition: 'status === 200',
                nodes: runtime.nodes.filter(n => n.type === 'function').slice(0, 2),
                edges: [],
                isActive: true,
                path: 'A → B',
              },
              {
                id: 'branch-2',
                label: 'Error Path',
                condition: 'status !== 200',
                nodes: runtime.nodes.filter(n => n.type === 'function').slice(2, 4),
                edges: [],
                isActive: false,
                path: 'A → C',
              },
            ]}
            activeBranchId={runtime.activeBranchId || undefined}
            onBranchSelect={(branchId) => runtime.setActiveBranch('fork-1', branchId)}
            showCondition={true}
          />
        )}

        {/* Loop Visualizer */}
        {showLoopVisualizer && selectedNodes.length === 1 && (
          <LoopVisualizer
            loopNodeId={selectedNodes[0].id}
            iterations={loopIterations as LoopVisualizerProps['iterations']}
            maxIterations={maxIterations}
            isExpanded={true}
            currentIteration={loopIterations.length}
            onIterationClick={(iteration) => console.log('Iteration click:', iteration)}
            onToggleExpand={() => {}}
          />
        )}

        {/* Failure Propagation Map */}
        {showFailureMap && rootFailure && (
          <FailurePropagationMap
            rootFailure={rootFailure as FailurePropagationMapProps['rootFailure']}
            propagation={failurePropagation as FailurePropagationMapProps['propagation']}
            onNodeClick={(nodeId) => runtime.selectNode(nodeId)}
          />
        )}

        {/* Resource Consumption Overlay */}
        {showResourceOverlay && (
          <ResourceConsumptionOverlay
            metrics={resourceMetrics as ResourceConsumptionOverlayProps['metrics']}
            aggregatedMetrics={aggregatedMetrics as ResourceConsumptionOverlayProps['aggregatedMetrics']}
            showLabels={true}
          />
        )}

        {/* Distributed Runtime Map */}
        {showDistributedMap && (
          <DistributedRuntimeMap
            regions={regions as DistributedRuntimeMapProps['regions']}
            selectedRegionId={runtime.selectedRegionId || undefined}
            onRegionSelect={(regionId) => selectRegion(regionId)}
            onRegionExpand={(regionId) => selectRegion(regionId)}
          />
        )}
      </div>

      {/* Bottom Timeline */}
      {runtime.isExecutionTimelineVisible && (
        <div className="absolute bottom-0 left-0 right-0 h-32 bg-[#0d0d14] border-t border-[rgba(255,255,255,0.08)]">
          <ExecutionTimelineMini
            events={runtime.executionEvents}
            currentTime={runtime.replayCurrentTime}
            onSeek={seekReplay}
          />
        </div>
      )}
    </div>
  )
}

// ============================================================================
// ExecutionTimelineMini
// ============================================================================

function ExecutionTimelineMini({
  events,
  currentTime,
  onSeek,
}: {
  events: ExecutionEvent[]
  currentTime: number
  onSeek: (time: number) => void
}) {
  if (events.length === 0) {
    return (
      <div className="flex items-center justify-center h-full text-[#6b7280] text-xs">
        No execution events yet
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full p-2">
      <div className="flex items-center justify-between text-[10px] text-[#6b7280] mb-1">
        <span>Execution Timeline</span>
        <span>{events.length} events</span>
      </div>
      <div className="flex-1 flex items-center gap-1">
        {events.map((event, index) => {
          const color = event.result === 'success' ? '#10b981' : event.result === 'failure' ? '#ef4444' : '#3b82f6'
          return (
            <div
              key={event.id}
              className="flex-1 h-6 rounded-sm cursor-pointer transition-all hover:scale-y-125"
              style={{ backgroundColor: color, opacity: 0.6 }}
              onClick={() => onSeek(event.timestamp)}
              title={`${event.type} at ${new Date(event.timestamp).toLocaleTimeString()}`}
            />
          )
        })}
      </div>
    </div>
  )
}
