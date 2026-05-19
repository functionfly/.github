/**
 * @functionfly/ui-graph
 * Visual graph runtime components for FunctionFly Studio
 */

import * as React from "react";

// --- Node Types ---
export type NodeType =
  | "function"
  | "agent"
  | "api"
  | "memory"
  | "database"
  | "robot"
  | "browser"
  | "gpu"
  | "workflow"
  | "trigger"
  | "condition"
  | "output";

export interface NodeData {
  id: string;
  type: NodeType;
  label: string;
  description?: string;
  status?: "idle" | "running" | "completed" | "error" | "waiting";
  metadata?: Record<string, unknown>;
  inputs?: PortData[];
  outputs?: PortData[];
  position?: { x: number; y: number };
  executionTime?: number;
  tokenCost?: number;
  costPerExecution?: number;
  errorRate?: number;
  reliabilityScore?: number;
  lastExecutedAt?: string;
  version?: string;
  icon?: React.ReactNode;
  color?: string;
  isHovered?: boolean;
  isSelected?: boolean;
}

export interface PortData {
  id: string;
  type: "input" | "output";
  label: string;
  dataType?: string;
  multiple?: boolean;
}

// --- Edge Types ---
export interface EdgeData {
  id: string;
  source: string;
  target: string;
  sourcePort: string;
  targetPort: string;
  label?: string;
  status?: "idle" | "active" | "error";
  throughput?: number; // tokens/sec
  latency?: number;
  activeTokens?: number; // for token flow animation
  isSelected?: boolean;
}

// --- Canvas Types ---
export type ViewMode = "design" | "execute" | "debug" | "monitor" | "simulate";
export type CanvasZoomLevel = "overview" | "normal" | "detail";

export interface CanvasViewport {
  x: number;
  y: number;
  zoom: number;
  rotation?: number;
}

export interface CanvasProps {
  nodes: NodeData[];
  edges: EdgeData[];
  viewMode?: ViewMode;
  zoomLevel?: CanvasZoomLevel;
  viewport?: CanvasViewport;
  onNodeSelect?: (node: NodeData) => void;
  onNodeDoubleClick?: (node: NodeData) => void;
  onEdgeSelect?: (edge: EdgeData) => void;
  onCanvasClick?: (position: { x: number; y: number }) => void;
  onNodeDrag?: (nodeId: string, position: { x: number; y: number }) => void;
  onNodeAdd?: (type: NodeType, position: { x: number; y: number }) => void;
  onEdgeAdd?: (source: string, target: string) => void;
  onEdgeRemove?: (edgeId: string) => void;
  children?: React.ReactNode;
  className?: string;
  readOnly?: boolean;
  showGrid?: boolean;
  showMinimap?: boolean;
  enablePan?: boolean;
  enableZoom?: boolean;
  enableConnect?: boolean;
  enableDrag?: boolean;
  animateTokenFlow?: boolean;
  tokenFlowSpeed?: number;
}

export interface ExecutionNodeProps extends Omit<NodeData, "position"> {
  position?: { x: number; y: number };
  isSelected?: boolean;
  isHovered?: boolean;
  onSelect?: (id: string) => void;
  onDragStart?: (id: string, pos: { x: number; y: number }) => void;
  onDragEnd?: (id: string, pos: { x: number; y: number }) => void;
  onAddInput?: (nodeId: string) => void;
  onAddOutput?: (nodeId: string) => void;
  onRemove?: (nodeId: string) => void;
}

export interface ExecutionEdgeProps extends Omit<EdgeData, "activeTokens"> {
  animated?: boolean;
  tokenFlow?: boolean;
  tokenPosition?: number; // 0-1 animation progress
  onSelect?: (id: string) => void;
  onRemove?: (id: string) => void;
}

export interface NodeInspectorProps {
  node: NodeData;
  onClose?: () => void;
  onUpdate?: (nodeId: string, data: Partial<NodeData>) => void;
}


export interface GraphMiniMapProps {
  nodes: NodeData[];
  edges?: EdgeData[];
  viewport: CanvasViewport;
  canvasWidth: number;
  canvasHeight: number;
  onNavigate?: (viewport: CanvasViewport) => void;
  className?: string;
}
export interface ExecutionTimelineProps {
  events: ExecutionEvent[];
  currentTime?: number;
  onTimeChange?: (time: number) => void;
  onEventClick?: (event: ExecutionEvent) => void;
  zoom?: number;
}

export interface ExecutionEvent {
  id: string;
  timestamp: number;
  type: "node_start" | "node_end" | "edge_traversal" | "error" | "fork" | "join";
  nodeId?: string;
  edgeId?: string;
  data?: Record<string, unknown>;
  duration?: number;
  result?: "success" | "failure" | "partial";
}

// Port connection types
export interface Connection {
  id: string;
  from: { nodeId: string; portId: string };
  to: { nodeId: string; portId: string };
  status?: "idle" | "active" | "error";
}

// Canvas state for hooks
export interface CanvasState {
  nodes: NodeData[];
  edges: EdgeData[];
  selectedNodes: string[];
  selectedEdges: string[];
  viewport: CanvasViewport;
  viewMode: ViewMode;
  isPanning: boolean;
  isConnecting: boolean;
  connectionStart?: { nodeId: string; portId: string; type: "input" | "output" };
  hoveredNode: string | null;
  hoveredEdge: string | null;
}

// Context
export interface GraphContextType {
  nodes: NodeData[];
  edges: EdgeData[];
  viewport: CanvasViewport;
  selectedNodes: string[];
  selectedEdges: string[];
  viewMode: ViewMode;
  setViewMode: (mode: ViewMode) => void;
  selectNode: (id: string, append?: boolean) => void;
  selectEdge: (id: string) => void;
  deselectAll: () => void;
  updateNode: (id: string, data: Partial<NodeData>) => void;
  updateEdge: (id: string, data: Partial<EdgeData>) => void;
  addNode: (type: NodeType, position: { x: number; y: number }) => string;
  removeNode: (id: string) => void;
  connect: (from: string, fromPort: string, to: string, toPort: string) => void;
  disconnect: (edgeId: string) => void;
  setViewport: (viewport: CanvasViewport) => void;
  zoomIn: () => void;
  zoomOut: () => void;
  fitView: () => void;
  resetView: () => void;
  forkWorkflow: () => CanvasState;
}
// ============================================================================
// Visual Graph Runtime Component Types
// ============================================================================

// --- Runtime Graph ---
export type RuntimeStatus = "idle" | "starting" | "running" | "paused" | "completed" | "failed" | "cancelled";

export interface RuntimeMetrics {
  totalExecutions: number;
  successfulExecutions: number;
  failedExecutions: number;
  averageLatency: number;
  tokensProcessed: number;
  totalCost: number;
  uptime: number;
}

export interface RuntimeGraphProps {
  nodes: NodeData[];
  edges: EdgeData[];
  runtimeStatus?: RuntimeStatus;
  metrics?: RuntimeMetrics;
  onNodeSelect?: (node: NodeData) => void;
  onNodeDoubleClick?: (node: NodeData) => void;
  onEdgeSelect?: (edge: EdgeData) => void;
  onCanvasClick?: (position: { x: number; y: number }) => void;
  onStart?: () => void;
  onPause?: () => void;
  onStop?: () => void;
  onReset?: () => void;
  className?: string;
}

// --- DynamicPort ---
export interface DynamicPortProps {
  nodeId: string;
  port: PortData;
  position: "left" | "right" | "top" | "bottom";
  isConnected?: boolean;
  isHovered?: boolean;
  isValidConnection?: boolean;
  onMouseDown?: (nodeId: string, portId: string, type: "input" | "output", e: React.MouseEvent) => void;
  onMouseUp?: (nodeId: string, portId: string, type: "input" | "output", e: React.MouseEvent) => void;
  onConnectionStart?: (nodeId: string, portId: string, type: "input" | "output") => void;
  onConnectionEnd?: (nodeId: string, portId: string, type: "input" | "output") => void;
  className?: string;
}

// --- TokenFlowRenderer ---
export interface TokenFlowParticle {
  id: string;
  edgeId: string;
  progress: number;
  speed: number;
  color: string;
  size: number;
}

export interface TokenFlowRendererProps {
  edges: EdgeData[];
  activeTokens?: number;
  tokenFlowSpeed?: number;
  showTokenPath?: boolean;
  onTokenClick?: (particle: TokenFlowParticle) => void;
  className?: string;
}

// --- NodeInspector ---
export interface NodeInspectorTab {
  id: string;
  label: string;
  icon?: React.ReactNode;
}

export interface NodeInspectorProps {
  node: NodeData;
  tabs?: NodeInspectorTab[];
  activeTab?: string;
  onTabChange?: (tabId: string) => void;
  onClose?: () => void;
  onUpdate?: (nodeId: string, data: Partial<NodeData>) => void;
  onDelete?: (nodeId: string) => void;
  className?: string;
}

// --- GraphMiniMap ---
export interface GraphMiniMapNode {
  id: string;
  type: NodeType;
  position: { x: number; y: number };
  status?: "idle" | "running" | "completed" | "error" | "waiting";
  isSelected?: boolean;
}

export interface RuntimeGraphMiniMapProps {
  nodes: GraphMiniMapNode[];
  edges: EdgeData[];
  viewport: CanvasViewport;
  canvasWidth: number;
  canvasHeight: number;
  onNavigate?: (viewport: CanvasViewport) => void;
  className?: string;
}

// --- ExecutionReplayControls ---
export type ReplaySpeed = 0.5 | 1 | 1.5 | 2 | 4 | 8;

export interface ExecutionReplayControlsProps {
  isPlaying: boolean;
  currentTime: number;
  duration: number;
  speed?: ReplaySpeed;
  onPlayPause: () => void;
  onStop: () => void;
  onSeek: (time: number) => void;
  onSpeedChange?: (speed: ReplaySpeed) => void;
  onStepForward?: () => void;
  onStepBackward?: () => void;
  className?: string;
}

// --- ExecutionHeatmap ---
export interface HeatmapCell {
  nodeId: string;
  timestamp: number;
  intensity: number; // 0-1
  status?: "idle" | "running" | "completed" | "error" | "waiting";
}

export interface ExecutionHeatmapProps {
  cells: HeatmapCell[];
  timeRange: { start: number; end: number };
  columns?: number;
  cellSize?: number;
  colorScale?: "hot" | "cool" | "viridis" | "plasma";
  onCellClick?: (cell: HeatmapCell) => void;
  className?: string;
}

// --- GraphViewport ---
export interface GraphViewportProps {
  viewport: CanvasViewport;
  zoomRange?: { min: number; max: number };
  showZoomIndicator?: boolean;
  showGrid?: boolean;
  gridSize?: number;
  onViewportChange?: (viewport: CanvasViewport) => void;
  onZoomIn?: () => void;
  onZoomOut?: () => void;
  onFitView?: () => void;
  onResetView?: () => void;
  children?: React.ReactNode;
  className?: string;
}

// --- WorkflowForkManager ---
export interface ForkPoint {
  id: string;
  nodeId: string;
  label?: string;
  branchCount: number;
  activeBranch?: string;
}

export interface WorkflowForkManagerProps {
  forks: ForkPoint[];
  onForkSelect?: (forkId: string, branchId: string) => void;
  onForkCreate?: (nodeId: string, branchCount: number) => void;
  onForkDelete?: (forkId: string) => void;
  className?: string;
}

// --- LiveNodeTelemetry ---
export interface TelemetryMetric {
  label: string;
  value: number | string;
  unit?: string;
  trend?: "up" | "down" | "stable";
  color?: string;
}

export interface LiveNodeTelemetryProps {
  nodeId: string;
  nodeLabel: string;
  metrics: TelemetryMetric[];
  isExpanded?: boolean;
  refreshInterval?: number;
  onToggleExpand?: () => void;
  className?: string;
}

// --- GraphStateDiffViewer ---
export interface StateDiff {
  type: "added" | "removed" | "modified";
  path: string;
  oldValue?: unknown;
  newValue?: unknown;
}

export interface GraphStateDiffViewerProps {
  leftState: CanvasState;
  rightState: CanvasState;
  diffType?: "full" | "nodes" | "edges" | "viewport";
  onRestore?: (diff: StateDiff) => void;
  className?: string;
}

// --- DependencyExplorer ---
export interface DependencyNode {
  id: string;
  label: string;
  type: NodeType;
  depth: number;
  path: string[];
  isCyclic?: boolean;
}

export interface DependencyExplorerProps {
  nodes: NodeData[];
  edges: EdgeData[];
  rootNodeId?: string;
  direction?: "downstream" | "upstream" | "both";
  onNodeClick?: (nodeId: string) => void;
  className?: string;
}

// --- OrchestrationLayerView ---
export interface OrchestrationLayer {
  id: string;
  label: string;
  description?: string;
  nodes: NodeData[];
  depth: number;
  isParallel?: boolean;
  hasError?: boolean;
}

export interface OrchestrationLayerViewProps {
  layers: OrchestrationLayer[];
  onLayerClick?: (layer: OrchestrationLayer) => void;
  onLayerExpand?: (layerId: string) => void;
  className?: string;
}

// --- ConditionalBranchRenderer ---
export interface ConditionalBranch {
  id: string;
  label: string;
  condition?: string;
  nodes: NodeData[];
  edges: EdgeData[];
  isActive?: boolean;
  path: string;
}

export interface ConditionalBranchRendererProps {
  branches: ConditionalBranch[];
  activeBranchId?: string;
  onBranchSelect?: (branchId: string) => void;
  showCondition?: boolean;
  className?: string;
}

// --- LoopVisualizer ---
export interface LoopIteration {
  id: string;
  index: number;
  startTime: number;
  endTime?: number;
  duration?: number;
  status?: "running" | "completed" | "failed" | "skipped";
  nodeId: string;
}

export interface LoopVisualizerProps {
  loopNodeId: string;
  iterations: LoopIteration[];
  maxIterations?: number;
  isExpanded?: boolean;
  currentIteration?: number;
  onIterationClick?: (iteration: LoopIteration) => void;
  onToggleExpand?: () => void;
  className?: string;
}

// --- FailurePropagationMap ---
export interface FailureNode {
  id: string;
  nodeId: string;
  label: string;
  type: NodeType;
  failureReason?: string;
  affectedNodes: string[];
  depth: number;
  isRootCause?: boolean;
  propagationPath: string[];
}

export interface FailurePropagationMapProps {
  rootFailure: {
    nodeId: string;
    label: string;
    reason: string;
  };
  propagation: FailureNode[];
  onNodeClick?: (nodeId: string) => void;
  className?: string;
}

// --- ResourceConsumptionOverlay ---
export interface ResourceMetric {
  nodeId: string;
  cpuPercent?: number;
  memoryMB?: number;
  networkMB?: number;
  storageMB?: number;
  gpuPercent?: number;
}

export interface ResourceConsumptionOverlayProps {
  metrics: ResourceMetric[];
  aggregatedMetrics?: {
    totalCpu: number;
    totalMemory: number;
    totalNetwork: number;
    totalStorage: number;
    totalGpu: number;
  };
  showLabels?: boolean;
  aggregationWindow?: number;
  className?: string;
}

// --- NodeVersionHistory ---
export interface NodeVersion {
  version: string;
  timestamp: number;
  author?: string;
  changes?: string;
  isActive?: boolean;
}

export interface NodeVersionHistoryProps {
  nodeId: string;
  nodeLabel: string;
  versions: NodeVersion[];
  onVersionSelect?: (version: NodeVersion) => void;
  onVersionRestore?: (version: NodeVersion) => void;
  className?: string;
}

// --- ExecutionTraceExplorer ---
export interface TraceSpan {
  id: string;
  traceId: string;
  spanId: string;
  parentSpanId?: string;
  nodeId?: string;
  operation: string;
  startTime: number;
  endTime?: number;
  duration?: number;
  status?: "ok" | "error";
  tags?: Record<string, string>;
  logs?: Array<{ timestamp: number; message: string }>;
}

export interface ExecutionTraceExplorerProps {
  traceId: string;
  spans: TraceSpan[];
  onSpanClick?: (span: TraceSpan) => void;
  onExpand?: (spanId: string) => void;
  className?: string;
}

// --- DistributedRuntimeMap ---
export interface RuntimeRegion {
  id: string;
  label: string;
  location: string;
  nodeCount: number;
  healthyNodes: number;
  unhealthyNodes: number;
  latency?: number;
  status?: "healthy" | "degraded" | "unhealthy";
  nodes: NodeData[];
}

export interface DistributedRuntimeMapProps {
  regions: RuntimeRegion[];
  selectedRegionId?: string;
  onRegionSelect?: (regionId: string) => void;
  onRegionExpand?: (regionId: string) => void;
  className?: string;
}
