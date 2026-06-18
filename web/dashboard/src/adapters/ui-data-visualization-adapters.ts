/**
 * Data Visualization Adapters
 * Transform dashboard data formats to match UI package component expectations
 */

/**
 * Target types for UI package compatibility (derived from ui-data-visualization types.d.ts)
 */

interface UIChartDataPoint {
  timestamp: number;
  value: number;
  label?: string;
}

interface UIScatterDataPoint {
  x: number;
  y: number;
  size?: number;
  category?: string;
  label?: string;
}

interface UITopologyNode {
  id: string;
  label: string;
  depth: number;
  children?: UITopologyNode[];
  metadata?: Record<string, unknown>;
}

interface UISunburstSegment {
  name: string;
  value: number;
  children?: UISunburstSegment[];
  color?: string;
}

interface UITreemapNode {
  id: string;
  name: string;
  value: number;
  children?: UITreemapNode[];
  category?: string;
}

interface UIFlowConnection {
  source: string;
  target: string;
  value: number;
  label?: string;
}

interface UIWaterfallStep {
  name: string;
  start: number;
  end: number;
  category?: string;
  color?: string;
}

interface UIClusterPoint {
  id: string;
  x: number;
  y: number;
  cluster: string;
  label?: string;
  metadata?: Record<string, unknown>;
}

interface UIVisGraphNode {
  id: string;
  label: string;
  type: string;
  connections: string[];
  weight?: number;
}

interface UIVisGraphEdge {
  source: string;
  target: string;
  weight: number;
  type?: string;
}

export interface StreamingLineData {
  timestamp?: number;
  value?: number;
  x?: number;
  y?: number;
  label?: string;
}

export interface ScatterDataInput {
  x: number;
  y: number;
  size?: number;
  z?: number;
  category?: string;
  label?: string;
}

export interface TopologyNodeInput {
  id: string;
  label: string;
  depth?: number;
  x?: number;
  y?: number;
  z?: number;
  children?: TopologyNodeInput[];
  metadata?: Record<string, unknown>;
}

export interface SunburstDataInput {
  name: string;
  value: number;
  children?: SunburstDataInput[];
}

export interface TreemapDataInput {
  id?: string;
  name: string;
  value: number;
  children?: TreemapDataInput[];
  category?: string;
}

export interface WaterfallDataInput {
  name?: string;
  label?: string;
  start?: number;
  end?: number;
  value?: number;
  category?: string;
  isTotal?: boolean;
}

export interface CostDataInput {
  category?: string;
  label?: string;
  value: number;
  color?: string;
}

export interface ClusterDataInput {
  id?: string;
  x: number;
  y: number;
  cluster: string;
  label?: string;
}

export interface AgentNodeInput {
  id: string;
  label: string;
  type: string;
  connections?: string[];
  position?: { x: number; y: number };
}

export interface AgentEdgeInput {
  source: string;
  target: string;
  weight?: number;
  type?: string;
  strength?: number;
}

/**
 * Adapt streaming line chart data
 */
export function adaptStreamingLineData(data: StreamingLineData[]): UIChartDataPoint[] {
  return data.map((d, i) => ({
    timestamp: d.timestamp ?? Date.now() - (data.length - i) * 1000,
    value: d.value ?? d.y ?? 0,
    label: d.label,
  }));
}

/**
 * Adapt scatter plot data
 */
export function adaptScatterData(data: ScatterDataInput[]): UIScatterDataPoint[] {
  return data.map(d => ({
    x: d.x,
    y: d.y,
    size: d.size ?? d.z,
    category: d.category,
    label: d.label,
  }));
}

/**
 * Adapt topology chart nodes
 */
export function adaptTopologyNodes(nodes: TopologyNodeInput[]): UITopologyNode[] {
  return nodes.map((n, i) => ({
    id: n.id,
    label: n.label,
    depth: n.depth ?? i,
    children: n.children ? adaptTopologyNodes(n.children) : undefined,
    metadata: n.metadata,
  }));
}

/**
 * Adapt sunburst data
 */
export function adaptSunburstData(data: SunburstDataInput): UISunburstSegment {
  return {
    name: data.name,
    value: data.value,
    children: data.children?.map(child => adaptSunburstData(child)),
    color: undefined,
  };
}

/**
 * Adapt treemap data
 */
export function adaptTreemapData(data: TreemapDataInput[]): UITreemapNode[] {
  return data.map((d, i) => ({
    id: d.id || `treemap-${i}`,
    name: d.name,
    value: d.value,
    children: d.children ? adaptTreemapData(d.children) : undefined,
    category: d.category,
  }));
}

/**
 * Adapt waterfall chart data
 */
export function adaptWaterfallData(data: WaterfallDataInput[]): UIWaterfallStep[] {
  let cumulative = 0;
  return data.map(d => {
    const start = d.isTotal ? 0 : cumulative;
    const end = d.isTotal ? d.value ?? d.end ?? 0 : cumulative + (d.value ?? 0);
    cumulative = end;
    return {
      name: d.name || d.label || '',
      start,
      end,
      category: d.category || 'compute',
      color: undefined,
    };
  });
}

/**
 * Adapt cost distribution data
 */
export function adaptCostData(data: CostDataInput[]): Array<{ category: string; value: number; color?: string }> {
  return data.map(d => ({
    category: d.category || d.label || '',
    value: d.value,
    color: d.color,
  }));
}

/**
 * Adapt semantic cluster data
 */
export function adaptClusterData(data: ClusterDataInput[]): UIClusterPoint[] {
  return data.map((d, i) => ({
    id: d.id || `cluster-${i}`,
    x: d.x,
    y: d.y,
    cluster: d.cluster,
    label: d.label,
    metadata: undefined,
  }));
}

/**
 * Extract unique clusters from cluster data
 */
export function extractClusters(data: ClusterDataInput[]): string[] {
  const clusterSet = new Set<string>();
  data.forEach(d => clusterSet.add(d.cluster));
  return Array.from(clusterSet);
}

/**
 * Adapt agent interaction graph nodes
 */
export function adaptAgentNodes(nodes: AgentNodeInput[]): UIVisGraphNode[] {
  return nodes.map(n => ({
    id: n.id,
    label: n.label,
    type: n.type,
    connections: n.connections || [],
    weight: undefined,
  }));
}

/**
 * Adapt agent interaction graph edges
 */
export function adaptAgentEdges(edges: AgentEdgeInput[]): UIVisGraphEdge[] {
  return edges.map(e => ({
    source: e.source,
    target: e.target,
    weight: e.weight ?? e.strength ?? 1,
    type: e.type,
  }));
}

/**
 * Adapt circular flow connections
 */
export function adaptCircularFlowConnections(
  nodes: Array<{ id: string; label: string; value?: number }>,
  connections: Array<{ source: string; target: string; value?: number }>
): UIFlowConnection[] {
  return connections.map(c => ({
    source: c.source,
    target: c.target,
    value: c.value ?? 1,
    label: undefined,
  }));
}
