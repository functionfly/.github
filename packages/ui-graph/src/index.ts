/**
 * @functionfly/ui-graph
 * Visual graph runtime components for FunctionFly Studio
 * Index and exports
 */

// ============================================================================
// Original Types
// ============================================================================

export type {
  NodeData,
  EdgeData,
  CanvasViewport,
  ViewMode,
  CanvasZoomLevel,
  CanvasProps,
  ExecutionNodeProps,
  ExecutionEdgeProps,
  NodeInspectorProps,
  GraphMiniMapProps,
  ExecutionTimelineProps,
  ExecutionEvent,
  PortData,
  Connection,
  CanvasState,
  GraphContextType,
  NodeType,
} from "./types";

// ============================================================================
// New Runtime Graph Types
// ============================================================================

export type {
  RuntimeStatus,
  RuntimeMetrics,
  RuntimeGraphProps,
  DynamicPortProps,
  TokenFlowRendererProps,
  TokenFlowParticle,
  NodeInspectorTab,
  GraphMiniMapNode,
  ExecutionReplayControlsProps,
  ReplaySpeed,
  ExecutionHeatmapProps,
  HeatmapCell,
  GraphViewportProps,
  WorkflowForkManagerProps,
  ForkPoint,
  LiveNodeTelemetryProps,
  TelemetryMetric,
  GraphStateDiffViewerProps,
  StateDiff,
  DependencyExplorerProps,
  DependencyNode,
  OrchestrationLayerViewProps,
  OrchestrationLayer,
  ConditionalBranchRendererProps,
  ConditionalBranch,
  LoopVisualizerProps,
  LoopIteration,
  FailurePropagationMapProps,
  FailureNode,
  ResourceConsumptionOverlayProps,
  ResourceMetric,
  NodeVersionHistoryProps,
  NodeVersion,
  ExecutionTraceExplorerProps,
  TraceSpan,
  DistributedRuntimeMapProps,
  RuntimeRegion,
} from "./types";

// ============================================================================
// Original Components
// ============================================================================

export { FunctionCanvas } from "./FunctionCanvas";
export { ExecutionNode } from "./ExecutionNode";
export { ExecutionEdge } from "./ExecutionEdge";
export { ExecutionTimeline } from "./ExecutionTimeline";
export { GraphContext } from "./FunctionCanvas";

// ============================================================================
// New Visual Graph Runtime Components
// ============================================================================

export { RuntimeGraph } from "./RuntimeGraph";
export { DynamicPort } from "./DynamicPort";
export { TokenFlowRenderer } from "./TokenFlowRenderer";
export { NodeInspector } from "./NodeInspector";
export { GraphMiniMap } from "./GraphMiniMap";
export { ExecutionReplayControls } from "./ExecutionReplayControls";
export { ExecutionHeatmap } from "./ExecutionHeatmap";
export { GraphViewport } from "./GraphViewport";
export { WorkflowForkManager } from "./WorkflowForkManager";
export { LiveNodeTelemetry } from "./LiveNodeTelemetry";
export { GraphStateDiffViewer } from "./GraphStateDiffViewer";
export { DependencyExplorer } from "./DependencyExplorer";
export { OrchestrationLayerView } from "./OrchestrationLayerView";
export { ConditionalBranchRenderer } from "./ConditionalBranchRenderer";
export { LoopVisualizer } from "./LoopVisualizer";
export { FailurePropagationMap } from "./FailurePropagationMap";
export { ResourceConsumptionOverlay } from "./ResourceConsumptionOverlay";
export { NodeVersionHistory } from "./NodeVersionHistory";
export { ExecutionTraceExplorer } from "./ExecutionTraceExplorer";
export { DistributedRuntimeMap } from "./DistributedRuntimeMap";

// ============================================================================
// Utilities
// ============================================================================

export { autoLayoutNodes } from "./utils";
export { getCanvasBounds } from "./utils";
export { formatNodeLabel } from "./utils";
export { getEdgePath } from "./utils";
export { getStatusColor } from "./utils";
export { VIEW_MODE_CONFIG } from "./utils";
export { generateSampleGraph } from "./utils";
export { GRAPH_THEME } from "./utils";
