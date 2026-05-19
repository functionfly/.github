/**
 * @functionfly/ui-data-visualization
 * Data Visualization Components - Charts, graphs, and visual analytics
 */

// ============================================================================
// Chart Data Types
// ============================================================================

export type ChartDataPoint = {
  timestamp: number;
  value: number;
  label?: string;
};

export type ScatterDataPoint = {
  x: number;
  y: number;
  size?: number;
  category?: string;
  label?: string;
};

export type TopologyNode = {
  id: string;
  label: string;
  depth: number;
  children?: TopologyNode[];
  metadata?: Record<string, unknown>;
};

export type SunburstSegment = {
  name: string;
  value: number;
  children?: SunburstSegment[];
  color?: string;
};

export type TreemapNode = {
  id: string;
  name: string;
  value: number;
  children?: TreemapNode[];
  category?: string;
};

export type FlowConnection = {
  source: string;
  target: string;
  value: number;
  label?: string;
};

export type WaterfallStep = {
  name: string;
  start: number;
  end: number;
  category?: string;
  color?: string;
};

export type ClusterPoint = {
  id: string;
  x: number;
  y: number;
  cluster: string;
  label?: string;
  metadata?: Record<string, unknown>;
};

export type GraphNode = {
  id: string;
  label: string;
  type: string;
  connections: string[];
  weight?: number;
};

export type GraphEdge = {
  source: string;
  target: string;
  weight: number;
  type?: string;
};

// ============================================================================
// Streaming Line Chart
// ============================================================================

export interface StreamingLineChartProps {
  data: ChartDataPoint[];
  maxPoints?: number;
  streaming?: boolean;
  showGrid?: boolean;
  showLegend?: boolean;
  color?: string;
  height?: number;
  onPointClick?: (point: ChartDataPoint) => void;
  className?: string;
}

// ============================================================================
// Realtime Scatter Plot
// ============================================================================

export interface RealtimeScatterPlotProps {
  data: ScatterDataPoint[];
  xLabel?: string;
  yLabel?: string;
  showLabels?: boolean;
  showGrid?: boolean;
  animated?: boolean;
  height?: number;
  onPointClick?: (point: ScatterDataPoint) => void;
  className?: string;
}

// ============================================================================
// 3D Topology Chart
// ============================================================================

export interface TopologyLink {
  source: string;
  target: string;
  weight?: number;
}

export interface ThreeDTopologyChartProps {
  nodes: TopologyNode[];
  links?: TopologyLink[];
  showLabels?: boolean;
  showConnections?: boolean;
  rotationSpeed?: number;
  height?: number;
  onNodeClick?: (node: TopologyNode) => void;
  className?: string;
}

// ============================================================================
// Execution Sunburst
// ============================================================================

export interface ExecutionSunburstProps {
  data: SunburstSegment;
  showLabels?: boolean;
  showValues?: boolean;
  height?: number;
  onSegmentClick?: (segment: SunburstSegment, path: string[]) => void;
  className?: string;
}

// ============================================================================
// Dependency Treemap
// ============================================================================

export interface DependencyTreemapProps {
  data: TreemapNode[];
  showLabels?: boolean;
  showValues?: boolean;
  height?: number;
  onNodeClick?: (node: TreemapNode, path: string[]) => void;
  className?: string;
}

// ============================================================================
// Circular Flow Diagram
// ============================================================================

export interface CircularFlowDiagramProps {
  nodes: Array<{ id: string; label: string; value: number; type?: string }>;
  connections: FlowConnection[];
  showLabels?: boolean;
  showValues?: boolean;
  animated?: boolean;
  height?: number;
  onNodeClick?: (nodeId: string) => void;
  onConnectionClick?: (connection: FlowConnection) => void;
  className?: string;
}

// ============================================================================
// Runtime Waterfall Chart
// ============================================================================

export interface RuntimeWaterfallChartProps {
  steps: WaterfallStep[];
  showLabels?: boolean;
  showValues?: boolean;
  showDuration?: boolean;
  height?: number;
  onStepClick?: (step: WaterfallStep) => void;
  className?: string;
}

// ============================================================================
// Cost Distribution Graph
// ============================================================================

export interface CostDistributionGraphProps {
  data: Array<{
    category: string;
    value: number;
    color?: string;
  }>;
  showLabels?: boolean;
  showValues?: boolean;
  showPercentages?: boolean;
  height?: number;
  onSegmentClick?: (segment: { category: string; value: number }) => void;
  className?: string;
}

// ============================================================================
// Semantic Cluster Chart
// ============================================================================

export interface SemanticClusterChartProps {
  points: ClusterPoint[];
  clusters: string[];
  showLabels?: boolean;
  showConnections?: boolean;
  height?: number;
  onPointClick?: (point: ClusterPoint) => void;
  className?: string;
}

// ============================================================================
// Agent Interaction Graph
// ============================================================================

export interface AgentInteractionGraphProps {
  nodes: GraphNode[];
  edges: GraphEdge[];
  showLabels?: boolean;
  showWeights?: boolean;
  animated?: boolean;
  height?: number;
  onNodeClick?: (node: GraphNode) => void;
  onEdgeClick?: (edge: GraphEdge) => void;
  className?: string;
}
