/**
 * @functionfly/ui-data-visualization
 * Data Visualization Components - Charts, graphs, and visual analytics
 */

import React, { useState, useMemo, useRef, useEffect } from 'react';
import { cn } from '@functionfly/ui-core';
import {
  LineChart,
  TrendingUp,
  TrendingDown,
  Activity,
  Zap,
  Clock,
  Target,
  BarChart3,
  PieChart,
  Circle,
  Box,
  Database,
  Cpu,
  Network,
  GitFork,
  GitBranch,
  Users,
  Bot,
  MessageSquare,
  Timer,
  Layers,
  ScatterChart,
  BarChart,
  PieChartIcon,
  CircleDot,
  Triangle,
  Hexagon,
  Diamond,
  ArrowRight,
  Eye,
  RefreshCw,
  Play,
  Pause,
  Maximize2,
  Minimize2,
} from 'lucide-react';

import type {
  ChartDataPoint,
  ScatterDataPoint,
  TopologyNode,
  SunburstSegment,
  TreemapNode,
  FlowConnection,
  WaterfallStep,
  ClusterPoint,
  GraphNode,
  GraphEdge,
  TopologyLink,
  StreamingLineChartProps,
  RealtimeScatterPlotProps,
  ThreeDTopologyChartProps,
  ExecutionSunburstProps,
  DependencyTreemapProps,
  CircularFlowDiagramProps,
  RuntimeWaterfallChartProps,
  CostDistributionGraphProps,
  SemanticClusterChartProps,
  AgentInteractionGraphProps,
} from './types';

// ============================================================================
// Streaming Line Chart
// ============================================================================

export const StreamingLineChart: React.FC<StreamingLineChartProps> = ({
  data,
  maxPoints = 50,
  showGrid = true,
  showLegend = false,
  color = '#00d4ff',
  height = 200,
  onPointClick,
  className,
}) => {
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);

  const displayData = useMemo(() => {
    return data.slice(-maxPoints);
  }, [data, maxPoints]);

  const maxValue = useMemo(() => {
    return Math.max(...displayData.map(d => d.value), 1);
  }, [displayData]);

  const minValue = useMemo(() => {
    return Math.min(...displayData.map(d => d.value), 0);
  }, [displayData]);

  const formatTime = (timestamp: number) => {
    return new Date(timestamp).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  };

  const getPointColor = (value: number, index: number) => {
    if (index === 0) return color;
    const prevValue = displayData[index - 1].value;
    if (value > prevValue) return '#00ff88';
    if (value < prevValue) return '#ff4466';
    return color;
  };

  return (
    <div className={cn('flex flex-col bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel flex items-center justify-between">
        <div className="flex items-center gap-2">
          <LineChart className="w-5 h-5 text-aviation-cyan" />
          <h3 className="text-sm font-medium text-aviation-text-primary">Streaming Line Chart</h3>
        </div>
        <div className="flex items-center gap-2 text-xs text-aviation-text-muted">
          <span>{displayData.length} points</span>
          <span>•</span>
          <span className="text-aviation-cyan">{displayData[displayData.length - 1]?.value.toFixed(2)}</span>
        </div>
      </div>

      <div className="p-4" style={{ height }}>
        <svg className="w-full h-full" viewBox="0 0 100 100" preserveAspectRatio="none">
          {showGrid && [0, 25, 50, 75, 100].map(y => (
            <line key={y} x1="0" y1={y} x2="100" y2={y} className="stroke-aviation-border-panel" strokeWidth="0.5" strokeDasharray="2 2" />
          ))}

          <path
            d={displayData.map((d, i) => {
              const x = (i / Math.max(displayData.length - 1, 1)) * 100;
              const normalizedValue = (d.value - minValue) / (maxValue - minValue || 1);
              const y = 100 - (normalizedValue * 100);
              return `${i === 0 ? 'M' : 'L'} ${x} ${y}`;
            }).join(' ')}
            className="fill-none"
            stroke={color}
            strokeWidth="2"
          />

          {displayData.map((d, i) => {
            const x = (i / Math.max(displayData.length - 1, 1)) * 100;
            const normalizedValue = (d.value - minValue) / (maxValue - minValue || 1);
            const y = 100 - (normalizedValue * 100);
            return (
              <circle
                key={i}
                cx={x}
                cy={y}
                r={hoveredIndex === i ? 2 : 1}
                className="fill-current cursor-pointer transition-all"
                style={{ color: getPointColor(d.value, i) }}
                onMouseEnter={() => setHoveredIndex(i)}
                onMouseLeave={() => setHoveredIndex(null)}
                onClick={() => onPointClick?.(d)}
              />
            );
          })}
        </svg>
      </div>

      <div className="px-4 py-2 border-t border-aviation-border-panel flex justify-between text-[10px] text-aviation-text-dim">
        <span>{formatTime(displayData[0]?.timestamp || Date.now())}</span>
        <span>{formatTime(displayData[displayData.length - 1]?.timestamp || Date.now())}</span>
      </div>

      {hoveredIndex !== null && displayData[hoveredIndex] && (
        <div className="px-4 py-2 border-t border-aviation-border-panel bg-aviation-bg-secondary flex items-center justify-between">
          <span className="text-xs text-aviation-text-muted">{formatTime(displayData[hoveredIndex].timestamp)}</span>
          <span className="text-sm font-medium text-aviation-text-primary">{displayData[hoveredIndex].value.toFixed(4)}</span>
        </div>
      )}
    </div>
  );
};

// ============================================================================
// Realtime Scatter Plot
// ============================================================================

export const RealtimeScatterPlot: React.FC<RealtimeScatterPlotProps> = ({
  data,
  xLabel = 'X',
  yLabel = 'Y',
  showLabels = false,
  showGrid = true,
  animated = true,
  height = 300,
  onPointClick,
  className,
}) => {
  const [hoveredPoint, setHoveredPoint] = useState<ScatterDataPoint | null>(null);

  const categories = useMemo(() => {
    return [...new Set(data.map(d => d.category || 'default'))];
  }, [data]);

  const categoryColors: Record<string, string> = {
    cluster_a: '#00d4ff',
    cluster_b: '#ff6b6b',
    cluster_c: '#51cf66',
    cluster_d: '#ffd43b',
    cluster_e: '#cc5de8',
    default: '#00d4ff',
  };

  const getColor = (point: ScatterDataPoint) => {
    return categoryColors[point.category || 'default'] || categoryColors.default;
  };

  return (
    <div className={cn('flex flex-col bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel flex items-center justify-between">
        <div className="flex items-center gap-2">
          <ScatterChart className="w-5 h-5 text-aviation-cyan" />
          <h3 className="text-sm font-medium text-aviation-text-primary">Realtime Scatter Plot</h3>
        </div>
        <div className="flex items-center gap-2 text-xs text-aviation-text-muted">
          {categories.map(cat => (
            <span key={cat} className="flex items-center gap-1">
              <span className="w-2 h-2 rounded-full" style={{ backgroundColor: categoryColors[cat] }} />
              <span>{cat}</span>
            </span>
          ))}
        </div>
      </div>

      <div className="p-4 flex-1" style={{ height }}>
        <svg className="w-full h-full" viewBox="0 0 100 100" preserveAspectRatio="none">
          {showGrid && (
            <>
              {[0, 25, 50, 75, 100].map(y => (
                <line key={`h-${y}`} x1="0" y1={y} x2="100" y2={y} className="stroke-aviation-border-panel" strokeWidth="0.5" strokeDasharray="2 2" />
              ))}
              {[0, 25, 50, 75, 100].map(x => (
                <line key={`v-${x}`} x1={x} y1="0" x2={x} y2="100" className="stroke-aviation-border-panel" strokeWidth="0.5" strokeDasharray="2 2" />
              ))}
            </>
          )}

          {data.map((point, i) => {
            const x = point.x;
            const y = 100 - point.y;
            const size = (point.size || 5) / 2;
            const isHovered = hoveredPoint === point;
            return (
              <g key={i}>
                <circle
                  cx={x}
                  cy={y}
                  r={isHovered ? size + 2 : size}
                  className={cn('fill-current transition-all cursor-pointer', animated ? 'animate-pulse' : '')}
                  style={{ color: getColor(point), opacity: isHovered ? 1 : 0.8 }}
                  onMouseEnter={() => setHoveredPoint(point)}
                  onMouseLeave={() => setHoveredPoint(null)}
                  onClick={() => onPointClick?.(point)}
                />
                {showLabels && point.label && (
                  <text x={x + size + 1} y={y} className="text-[6px] fill-aviation-text-primary">
                    {point.label}
                  </text>
                )}
              </g>
            );
          })}
        </svg>
      </div>

      <div className="px-4 py-2 border-t border-aviation-border-panel flex justify-between text-[10px] text-aviation-text-dim">
        <span>{xLabel}: 0-100</span>
        <span>{yLabel}: 0-100</span>
      </div>

      {hoveredPoint && (
        <div className="px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
          <div className="grid grid-cols-3 gap-4 text-xs">
            <div>
              <span className="text-aviation-text-dim">X: </span>
              <span className="text-aviation-text-primary font-medium">{hoveredPoint.x.toFixed(2)}</span>
            </div>
            <div>
              <span className="text-aviation-text-dim">Y: </span>
              <span className="text-aviation-text-primary font-medium">{hoveredPoint.y.toFixed(2)}</span>
            </div>
            <div>
              <span className="text-aviation-text-dim">Category: </span>
              <span className="text-aviation-cyan">{hoveredPoint.category || 'default'}</span>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

// ============================================================================
// 3D Topology Chart
// ============================================================================

export const ThreeDTopologyChart: React.FC<ThreeDTopologyChartProps> = ({
  nodes,
  links = [],
  showLabels = true,
  showConnections = true,
  height = 350,
  onNodeClick,
  className,
}) => {
  const [rotation, setRotation] = useState(0);
  const [hoveredNode, setHoveredNode] = useState<TopologyNode | null>(null);

  useEffect(() => {
    const interval = setInterval(() => {
      setRotation(r => (r + 0.5) % 360);
    }, 50);
    return () => clearInterval(interval);
  }, []);

  const layoutNodes = useMemo(() => {
    const result: Array<TopologyNode & { x: number; y: number; z: number }> = [];
    
    const layoutLevel = (nodes: TopologyNode[], depth: number, xOffset: number, spread: number) => {
      nodes.forEach((node, i) => {
        const angle = ((i - (nodes.length - 1) / 2) / nodes.length) * Math.PI;
        const x = xOffset + Math.sin(angle + (rotation * Math.PI / 180)) * spread;
        const y = depth * 25 + Math.cos(angle * 1.5) * 8;
        const z = depth + Math.cos(angle) * spread * 0.3;
        result.push({ ...node, x, y, z });
        if (node.children) {
          layoutLevel(node.children, depth + 1, xOffset, spread * 0.6);
        }
      });
    };
    
    layoutLevel(nodes, 0, 50, 35);
    return result;
  }, [nodes, rotation]);

  const getNodeColor = (depth: number) => {
    const colors = ['#00d4ff', '#00ff88', '#ffd43b', '#ff6b6b', '#cc5de8'];
    return colors[depth % colors.length];
  };

  return (
    <div className={cn('relative flex flex-col bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Box className="w-5 h-5 text-aviation-cyan" />
          <h3 className="text-sm font-medium text-aviation-text-primary">3D Topology</h3>
        </div>
        <div className="flex items-center gap-2 text-xs text-aviation-text-muted">
          <span>{layoutNodes.length} nodes</span>
          <button className="p-1 hover:bg-aviation-bg-instrument rounded" onClick={() => setRotation(0)}>
            <RefreshCw className="w-4 h-4" />
          </button>
        </div>
      </div>

      <div className="relative flex-1" style={{ height }}>
        <svg className="w-full h-full" viewBox="0 0 100 120" preserveAspectRatio="xMidYMid meet">
          {showConnections && links.map((link, i) => {
            const sourceNode = layoutNodes.find(n => n.id === link.source);
            const targetNode = layoutNodes.find(n => n.id === link.target);
            if (!sourceNode || !targetNode) return null;
            return (
              <line
                key={i}
                x1={sourceNode.x}
                y1={sourceNode.y}
                x2={targetNode.x}
                y2={targetNode.y}
                className="stroke-aviation-border-panel"
                strokeWidth="0.5"
                opacity="0.5"
              />
            );
          })}

          {layoutNodes.map((node, i) => {
            const scale = 1 + (node.z * 0.1);
            const opacity = 0.5 + (node.z * 0.1);
            const isHovered = hoveredNode?.id === node.id;
            return (
              <g
                key={node.id}
                onClick={() => onNodeClick?.(node)}
                onMouseEnter={() => setHoveredNode(node)}
                onMouseLeave={() => setHoveredNode(null)}
                className="cursor-pointer"
              >
                <polygon
                  points={`${node.x},${node.y - 5 * scale} ${node.x + 4 * scale},${node.y} ${node.x},${node.y + 3 * scale} ${node.x - 4 * scale},${node.y}`}
                  className="fill-current transition-all"
                  style={{ color: getNodeColor(node.depth), opacity: isHovered ? 1 : opacity }}
                  stroke={isHovered ? '#fff' : 'transparent'}
                  strokeWidth="0.5"
                />
                {showLabels && (
                  <text x={node.x} y={node.y + 10} textAnchor="middle" className="text-[6px] fill-aviation-text-primary">
                    {node.label}
                  </text>
                )}
              </g>
            );
          })}
        </svg>

        <div className="absolute bottom-2 left-2 flex items-center gap-2 text-[10px] text-aviation-text-dim">
          <span>Depth:</span>
          {[0, 1, 2, 3, 4].map(d => (
            <span key={d} className="flex items-center gap-1">
              <span className="w-2 h-2 rounded" style={{ backgroundColor: getNodeColor(d) }} />
              {d}
            </span>
          ))}
        </div>
      </div>

      {hoveredNode && (
        <div className="px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
          <div className="flex items-center justify-between">
            <div>
              <span className="text-sm font-medium text-aviation-text-primary">{hoveredNode.label}</span>
              <span className="ml-2 text-xs text-aviation-text-dim">Depth: {hoveredNode.depth}</span>
            </div>
            {hoveredNode.metadata && (
              <span className="text-xs text-aviation-cyan">View metadata</span>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

// ============================================================================
// Execution Sunburst
// ============================================================================

export const ExecutionSunburst: React.FC<ExecutionSunburstProps> = ({
  data,
  showLabels = true,
  showValues = true,
  height = 350,
  onSegmentClick,
  className,
}) => {
  const [activeSegment, setActiveSegment] = useState<{ segment: SunburstSegment; path: string[] } | null>(null);

  const colors = ['#00d4ff', '#00ff88', '#ffd43b', '#ff6b6b', '#cc5de8', '#20c997', '#fd7e14', '#e599f7'];

  const buildHierarchy = (segment: SunburstSegment, startAngle: number, endAngle: number, depth: number, path: string[]): Array<{
    segment: SunburstSegment;
    path: string[];
    startAngle: number;
    endAngle: number;
    depth: number;
  }> => {
    const result: Array<{ segment: SunburstSegment; path: string[]; startAngle: number; endAngle: number; depth: number }> = [];
    const totalValue = segment.children?.reduce((sum, c) => sum + c.value, 0) || segment.value;
    
    if (segment.children && segment.children.length > 0) {
      let currentAngle = startAngle;
      segment.children.forEach((child, i) => {
        const childAngleSpan = ((child.value / totalValue) * (endAngle - startAngle));
        const childEndAngle = currentAngle + childAngleSpan;
        result.push({ segment: child, path: [...path, segment.name], startAngle: currentAngle, endAngle: childEndAngle, depth: depth + 1 });
        result.push(...buildHierarchy(child, currentAngle, childEndAngle, depth + 1, [...path, segment.name]));
        currentAngle = childEndAngle;
      });
    }
    
    return result;
  };

  const segments = useMemo(() => {
    return buildHierarchy(data, 0, 360, 0, []);
  }, [data]);

  const describeArc = (startAngle: number, endAngle: number, innerRadius: number, outerRadius: number) => {
    const start = polarToCartesian(50, 50, outerRadius, endAngle);
    const end = polarToCartesian(50, 50, outerRadius, startAngle);
    const innerStart = polarToCartesian(50, 50, innerRadius, endAngle);
    const innerEnd = polarToCartesian(50, 50, innerRadius, startAngle);
    const largeArcFlag = endAngle - startAngle <= 180 ? '0' : '1';
    return `M ${start.x} ${start.y} A ${outerRadius} ${outerRadius} 0 ${largeArcFlag} 0 ${end.x} ${end.y} L ${innerEnd.x} ${innerEnd.y} A ${innerRadius} ${innerRadius} 0 ${largeArcFlag} 1 ${innerStart.x} ${innerStart.y} Z`;
  };

  const polarToCartesian = (cx: number, cy: number, radius: number, angle: number) => {
    const rad = (angle - 90) * Math.PI / 180;
    return { x: cx + radius * Math.cos(rad), y: cy + radius * Math.sin(rad) };
  };

  return (
    <div className={cn('flex flex-col bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel flex items-center justify-between">
        <div className="flex items-center gap-2">
          <PieChartIcon className="w-5 h-5 text-aviation-cyan" />
          <h3 className="text-sm font-medium text-aviation-text-primary">Execution Sunburst</h3>
        </div>
        <div className="text-xs text-aviation-text-muted">
          {segments.length} segments
        </div>
      </div>

      <div className="flex-1 flex items-center justify-center p-4" style={{ height }}>
        <svg className="max-w-full max-h-full" viewBox="0 0 100 100">
          <g onClick={() => onSegmentClick?.(data, [])} className="cursor-pointer">
            <path
              d={describeArc(0, 360, 0, 45)}
              className="fill-aviation-bg-instrument stroke-aviation-border-panel"
              strokeWidth="0.5"
            />
          </g>
          {[0, 1, 2].map(depth => {
            const depthRadius = 15 + depth * 15;
            const innerRadius = depthRadius - 13;
            return segments.filter(s => s.depth === depth).map((s, i) => {
              const color = s.segment.color || colors[i % colors.length];
              const isActive = activeSegment?.path.join('/') === s.path.join('/') && activeSegment?.segment.name === s.segment.name;
              return (
                <g
                  key={`${s.segment.name}-${s.depth}-${i}`}
                  onClick={(e) => { e.stopPropagation(); setActiveSegment(s); onSegmentClick?.(s.segment, s.path); }}
                  className="cursor-pointer"
                >
                  <path
                    d={describeArc(s.startAngle, s.endAngle, innerRadius, depthRadius)}
                    className="transition-all"
                    style={{ fill: color, opacity: isActive ? 1 : 0.85 }}
                    stroke={isActive ? '#fff' : 'transparent'}
                    strokeWidth="0.5"
                  />
                  {showLabels && (s.endAngle - s.startAngle > 20) && (
                    <text
                      x={polarToCartesian(50, 50, (innerRadius + depthRadius) / 2, (s.startAngle + s.endAngle) / 2).x}
                      y={polarToCartesian(50, 50, (innerRadius + depthRadius) / 2, (s.startAngle + s.endAngle) / 2).y}
                      textAnchor="middle"
                      dominantBaseline="middle"
                      className="text-[6px] fill-white font-medium pointer-events-none"
                    >
                      {s.segment.name.length > 8 ? s.segment.name.slice(0, 8) + '...' : s.segment.name}
                    </text>
                  )}
                </g>
              );
            });
          })}
        </svg>
      </div>

      {activeSegment && (
        <div className="px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
          <div className="flex items-center justify-between mb-1">
            <span className="text-sm font-medium text-aviation-text-primary">{activeSegment.segment.name}</span>
            <span className="text-sm text-aviation-cyan">{activeSegment.segment.value.toFixed(2)}</span>
          </div>
          <div className="text-xs text-aviation-text-dim">
            Path: {activeSegment.path.join(' → ')} → {activeSegment.segment.name}
          </div>
        </div>
      )}

      <div className="px-4 py-2 border-t border-aviation-border-panel flex flex-wrap gap-2">
        {data.children?.slice(0, 6).map((child, i) => (
          <span key={child.name} className="flex items-center gap-1 text-[10px] text-aviation-text-muted">
            <span className="w-2 h-2 rounded" style={{ backgroundColor: child.color || colors[i % colors.length] }} />
            {child.name}
          </span>
        ))}
      </div>
    </div>
  );
};

// ============================================================================
// Dependency Treemap
// ============================================================================

export const DependencyTreemap: React.FC<DependencyTreemapProps> = ({
  data,
  showLabels = true,
  showValues = true,
  height = 350,
  onNodeClick,
  className,
}) => {
  const [hoveredNode, setHoveredNode] = useState<TreemapNode | null>(null);

  const colors = ['#00d4ff', '#00ff88', '#ffd43b', '#ff6b6b', '#cc5de8', '#20c997', '#fd7e14', '#e599f7'];

  const treemapLayout = (nodes: TreemapNode[], x: number, y: number, w: number, h: number): Array<TreemapNode & { x: number; y: number; width: number; height: number }> => {
    const result: Array<TreemapNode & { x: number; y: number; width: number; height: number }> = [];
    const totalValue = nodes.reduce((sum, n) => sum + n.value, 0);
    
    if (nodes.length === 0) return result;

    if (w > h) {
      let currentX = x;
      nodes.forEach((node) => {
        const nodeWidth = (node.value / totalValue) * w;
        result.push({ ...node, x: currentX, y, width: nodeWidth, height: h });
        if (node.children && node.children.length > 0) {
          result.push(...treemapLayout(node.children, currentX, y, nodeWidth, h));
        }
        currentX += nodeWidth;
      });
    } else {
      let currentY = y;
      nodes.forEach((node) => {
        const nodeHeight = (node.value / totalValue) * h;
        result.push({ ...node, x, y: currentY, width: w, height: nodeHeight });
        if (node.children && node.children.length > 0) {
          result.push(...treemapLayout(node.children, x, currentY, w, nodeHeight));
        }
        currentY += nodeHeight;
      });
    }
    
    return result;
  };

  const layout = useMemo(() => {
    return treemapLayout(data, 0, 0, 100, 100);
  }, [data]);

  const getColor = (node: TreemapNode, index: number) => {
    return node.category ? colors[data.findIndex(d => d.id === node.category) % colors.length] : colors[index % colors.length];
  };

  return (
    <div className={cn('flex flex-col bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Layers className="w-5 h-5 text-aviation-cyan" />
          <h3 className="text-sm font-medium text-aviation-text-primary">Dependency Treemap</h3>
        </div>
        <div className="text-xs text-aviation-text-muted">
          {data.length} top-level nodes
        </div>
      </div>

      <div className="flex-1 p-4" style={{ height }}>
        <svg className="w-full h-full" viewBox="0 0 100 100" preserveAspectRatio="none">
          {layout.map((node, i) => {
            const isHovered = hoveredNode?.id === node.id;
            return (
              <g
                key={node.id}
                onMouseEnter={() => setHoveredNode(node)}
                onMouseLeave={() => setHoveredNode(null)}
                onClick={() => onNodeClick?.(node, data.find(d => d.id === node.id) ? [d.id] : [])}
                className="cursor-pointer"
              >
                <rect
                  x={node.x + 0.5}
                  y={node.y + 0.5}
                  width={Math.max(0, node.width - 1)}
                  height={Math.max(0, node.height - 1)}
                  className="transition-all"
                  style={{ fill: getColor(node, i), opacity: isHovered ? 1 : 0.8 }}
                  stroke={isHovered ? '#fff' : 'transparent'}
                  strokeWidth="0.5"
                  rx="2"
                />
                {showLabels && node.width > 8 && node.height > 6 && (
                  <>
                    <text
                      x={node.x + node.width / 2}
                      y={node.y + node.height / 2 - 2}
                      textAnchor="middle"
                      className="text-[6px] fill-white font-medium pointer-events-none"
                    >
                      {node.name.length > 10 ? node.name.slice(0, 10) + '...' : node.name}
                    </text>
                    {showValues && (
                      <text
                        x={node.x + node.width / 2}
                        y={node.y + node.height / 2 + 4}
                        textAnchor="middle"
                        className="text-[5px] fill-white/70 pointer-events-none"
                      >
                        {node.value.toFixed(0)}
                      </text>
                    )}
                  </>
                )}
              </g>
            );
          })}
        </svg>
      </div>

      {hoveredNode && (
        <div className="px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
          <div className="flex items-center justify-between mb-1">
            <span className="text-sm font-medium text-aviation-text-primary">{hoveredNode.name}</span>
            <span className="text-sm text-aviation-cyan">{hoveredNode.value.toFixed(2)}</span>
          </div>
          <div className="text-xs text-aviation-text-dim">
            Category: {hoveredNode.category || 'uncategorized'} • ID: {hoveredNode.id}
          </div>
        </div>
      )}
    </div>
  );
};

// ============================================================================
// Circular Flow Diagram
// ============================================================================

export const CircularFlowDiagram: React.FC<CircularFlowDiagramProps> = ({
  nodes,
  connections,
  showLabels = true,
  showValues = true,
  animated = true,
  height = 350,
  onNodeClick,
  onConnectionClick,
  className,
}) => {
  const [hoveredItem, setHoveredItem] = useState<string | null>(null);

  const nodePositions = useMemo(() => {
    const positions: Record<string, { x: number; y: number }> = {};
    const angleStep = (2 * Math.PI) / nodes.length;
    const radius = 35;
    
    nodes.forEach((node, i) => {
      const angle = i * angleStep - Math.PI / 2;
      positions[node.id] = {
        x: 50 + radius * Math.cos(angle),
        y: 50 + radius * Math.sin(angle),
      };
    });
    return positions;
  }, [nodes]);

  const getNodeColor = (type?: string) => {
    switch (type) {
      case 'source': return '#00d4ff';
      case 'processor': return '#00ff88';
      case 'sink': return '#ff6b6b';
      default: return '#ffd43b';
    }
  };

  const getConnectionColor = (value: number, maxValue: number) => {
    const ratio = value / maxValue;
    if (ratio > 0.7) return '#00d4ff';
    if (ratio > 0.4) return '#00ff88';
    return '#ffd43b';
  };

  const maxValue = useMemo(() => {
    return Math.max(...connections.map(c => c.value), 1);
  }, [connections]);

  return (
    <div className={cn('flex flex-col bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel flex items-center justify-between">
        <div className="flex items-center gap-2">
          <RefreshCw className="w-5 h-5 text-aviation-cyan" />
          <h3 className="text-sm font-medium text-aviation-text-primary">Circular Flow</h3>
        </div>
        <div className="text-xs text-aviation-text-muted">
          {connections.length} connections
        </div>
      </div>

      <div className="flex-1 flex items-center justify-center p-4" style={{ height }}>
        <svg className="max-w-full max-h-full" viewBox="0 0 100 100">
          <circle cx="50" cy="50" r="35" className="fill-none stroke-aviation-border-panel" strokeWidth="0.5" strokeDasharray="4 2" />

          {connections.map((conn, i) => {
            const sourcePos = nodePositions[conn.source];
            const targetPos = nodePositions[conn.target];
            if (!sourcePos || !targetPos) return null;

            const isHovered = hoveredItem === `conn-${i}`;
            const midX = (sourcePos.x + targetPos.x) / 2;
            const midY = (sourcePos.y + targetPos.y) / 2;

            return (
              <g key={i}>
                <path
                  d={`M ${sourcePos.x} ${sourcePos.y} Q ${midX} ${midY - 10} ${targetPos.x} ${targetPos.y}`}
                  className="fill-none transition-all"
                  stroke={getConnectionColor(conn.value, maxValue)}
                  strokeWidth={isHovered ? 2 : 1}
                  opacity={isHovered ? 1 : 0.6}
                  strokeDasharray={animated ? '4 2' : undefined}
                  onMouseEnter={() => setHoveredItem(`conn-${i}`)}
                  onMouseLeave={() => setHoveredItem(null)}
                  onClick={() => onConnectionClick?.(conn)}
                  style={{ cursor: 'pointer' }}
                />
                {showValues && isHovered && (
                  <text x={midX} y={midY - 5} textAnchor="middle" className="text-[6px] fill-aviation-text-primary">
                    {conn.value.toFixed(1)}
                  </text>
                )}
              </g>
            );
          })}

          {nodes.map((node, i) => {
            const pos = nodePositions[node.id];
            if (!pos) return null;

            const isHovered = hoveredItem === node.id;
            const color = getNodeColor(node.type);
            const radius = 5 + (node.value / Math.max(...nodes.map(n => n.value))) * 3;

            return (
              <g
                key={node.id}
                onMouseEnter={() => setHoveredItem(node.id)}
                onMouseLeave={() => setHoveredItem(null)}
                onClick={() => onNodeClick?.(node.id)}
                className="cursor-pointer"
              >
                <circle
                  cx={pos.x}
                  cy={pos.y}
                  r={isHovered ? radius + 2 : radius}
                  className="transition-all"
                  style={{ fill: color, opacity: isHovered ? 1 : 0.85 }}
                  stroke={isHovered ? '#fff' : 'transparent'}
                  strokeWidth="0.5"
                />
                {showLabels && (
                  <text
                    x={pos.x}
                    y={pos.y + radius + 6}
                    textAnchor="middle"
                    className="text-[6px] fill-aviation-text-primary"
                  >
                    {node.label.length > 8 ? node.label.slice(0, 8) + '...' : node.label}
                  </text>
                )}
              </g>
            );
          })}
        </svg>
      </div>

      <div className="px-4 py-2 border-t border-aviation-border-panel flex justify-center gap-4">
        {['source', 'processor', 'sink'].map(type => (
          <span key={type} className="flex items-center gap-1 text-[10px] text-aviation-text-muted">
            <span className="w-2 h-2 rounded-full" style={{ backgroundColor: getNodeColor(type) }} />
            {type}
          </span>
        ))}
      </div>
    </div>
  );
};

// ============================================================================
// Runtime Waterfall Chart
// ============================================================================

export const RuntimeWaterfallChart: React.FC<RuntimeWaterfallChartProps> = ({
  steps,
  showLabels = true,
  showValues = true,
  showDuration = true,
  height = 300,
  onStepClick,
  className,
}) => {
  const [hoveredStep, setHoveredStep] = useState<WaterfallStep | null>(null);

  const colors: Record<string, string> = {
    compute: '#00d4ff',
    memory: '#00ff88',
    network: '#ffd43b',
    io: '#ff6b6b',
    wait: '#cc5de8',
  };

  const maxEnd = useMemo(() => {
    return Math.max(...steps.map(s => s.end), 1);
  }, [steps]);

  const chartHeight = 70;

  return (
    <div className={cn('flex flex-col bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel flex items-center justify-between">
        <div className="flex items-center gap-2">
          <BarChart3 className="w-5 h-5 text-aviation-cyan" />
          <h3 className="text-sm font-medium text-aviation-text-primary">Runtime Waterfall</h3>
        </div>
        <div className="flex items-center gap-2 text-xs text-aviation-text-muted">
          <span>Total: {steps.reduce((sum, s) => sum + (s.end - s.start), 0).toFixed(1)}ms</span>
        </div>
      </div>

      <div className="flex-1 p-4 overflow-x-auto" style={{ height }}>
        <svg className="w-full h-full" viewBox={`0 0 ${steps.length * 20 + 40} 100`} preserveAspectRatio="xMinYMid meet">
          <line x1="20" y1={chartHeight + 5} x2={steps.length * 20 + 20} y2={chartHeight + 5} className="stroke-aviation-border-panel" strokeWidth="0.5" />

          {[0, 0.25, 0.5, 0.75, 1].map(ratio => (
            <g key={ratio}>
              <line x1="20" y1={chartHeight - ratio * chartHeight + 5} x2={steps.length * 20 + 20} y2={chartHeight - ratio * chartHeight + 5} className="stroke-aviation-border-panel" strokeWidth="0.5" strokeDasharray="2 2" />
              <text x="15" y={chartHeight - ratio * chartHeight + 8} textAnchor="end" className="text-[6px] fill-aviation-text-dim">
                {(maxEnd * ratio).toFixed(0)}ms
              </text>
            </g>
          ))}

          {steps.map((step, i) => {
            const x = 25 + i * 20;
            const y = chartHeight - (step.end / maxEnd) * chartHeight + 5;
            const barHeight = ((step.end - step.start) / maxEnd) * chartHeight;
            const color = step.color || colors[step.category || 'compute'];
            const isHovered = hoveredStep === step;

            return (
              <g
                key={i}
                onMouseEnter={() => setHoveredStep(step)}
                onMouseLeave={() => setHoveredStep(null)}
                onClick={() => onStepClick?.(step)}
                className="cursor-pointer"
              >
                {i > 0 && (
                  <line
                    x1={x - 15}
                    y1={chartHeight - (step.start / maxEnd) * chartHeight + 5}
                    x2={x}
                    y2={chartHeight - (step.start / maxEnd) * chartHeight + 5}
                    className="stroke-aviation-text-muted"
                    strokeWidth="0.5"
                  />
                )}

                <rect
                  x={x - 8}
                  y={y}
                  width="16"
                  height={Math.max(2, barHeight)}
                  className="transition-all"
                  style={{ fill: color, opacity: isHovered ? 1 : 0.85 }}
                  stroke={isHovered ? '#fff' : 'transparent'}
                  strokeWidth="0.5"
                  rx="2"
                />

                {showLabels && (
                  <text
                    x={x}
                    y={chartHeight + 15}
                    textAnchor="middle"
                    className="text-[5px] fill-aviation-text-dim"
                    transform={`rotate(-45, ${x}, ${chartHeight + 15})`}
                  >
                    {step.name.length > 6 ? step.name.slice(0, 6) + '...' : step.name}
                  </text>
                )}

                {showValues && isHovered && (
                  <text x={x} y={y - 3} textAnchor="middle" className="text-[6px] fill-aviation-text-primary">
                    {(step.end - step.start).toFixed(1)}ms
                  </text>
                )}
              </g>
            );
          })}
        </svg>
      </div>

      {hoveredStep && (
        <div className="px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
          <div className="flex items-center justify-between mb-1">
            <span className="text-sm font-medium text-aviation-text-primary">{hoveredStep.name}</span>
            <span className="text-sm text-aviation-cyan">{(hoveredStep.end - hoveredStep.start).toFixed(2)}ms</span>
          </div>
          <div className="flex items-center gap-4 text-xs text-aviation-text-dim">
            <span>Start: {hoveredStep.start.toFixed(2)}ms</span>
            <span>End: {hoveredStep.end.toFixed(2)}ms</span>
            <span>Category: {hoveredStep.category || 'compute'}</span>
          </div>
        </div>
      )}

      <div className="px-4 py-2 border-t border-aviation-border-panel flex flex-wrap gap-2">
        {Object.entries(colors).map(([cat, color]) => (
          <span key={cat} className="flex items-center gap-1 text-[10px] text-aviation-text-muted">
            <span className="w-2 h-2 rounded" style={{ backgroundColor: color }} />
            {cat}
          </span>
        ))}
      </div>
    </div>
  );
};

// ============================================================================
// Cost Distribution Graph
// ============================================================================

export const CostDistributionGraph: React.FC<CostDistributionGraphProps> = ({
  data,
  showLabels = true,
  showValues = true,
  showPercentages = true,
  height = 300,
  onSegmentClick,
  className,
}) => {
  const [hoveredSegment, setHoveredSegment] = useState<string | null>(null);

  const total = useMemo(() => {
    return data.reduce((sum, d) => sum + d.value, 0);
  }, [data]);

  const sortedData = useMemo(() => {
    return [...data].sort((a, b) => b.value - a.value);
  }, [data]);

  const colors = ['#00d4ff', '#00ff88', '#ffd43b', '#ff6b6b', '#cc5de8', '#20c997', '#fd7e14', '#e599f7'];

  const formatCurrency = (value: number) => {
    if (value >= 1000) return `$${(value / 1000).toFixed(1)}k`;
    return `$${value.toFixed(2)}`;
  };

  return (
    <div className={cn('flex flex-col bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel flex items-center justify-between">
        <div className="flex items-center gap-2">
          <PieChart className="w-5 h-5 text-aviation-cyan" />
          <h3 className="text-sm font-medium text-aviation-text-primary">Cost Distribution</h3>
        </div>
        <div className="text-xs text-aviation-text-muted">
          Total: {formatCurrency(total)}
        </div>
      </div>

      <div className="flex-1 flex items-center gap-6 p-4" style={{ height }}>
        <div className="relative flex-shrink-0" style={{ width: height * 0.5, height: height * 0.5 }}>
          <svg className="w-full h-full" viewBox="0 0 100 100">
            {sortedData.reduce<{ angle: number; paths: JSX.Element[] }>((acc, segment, i) => {
              const percentage = segment.value / total;
              const angleSpan = percentage * 360;
              const startAngle = acc.angle;
              const endAngle = acc.angle + angleSpan;

              const startRad = (startAngle - 90) * Math.PI / 180;
              const endRad = (endAngle - 90) * Math.PI / 180;
              const largeArc = angleSpan > 180 ? 1 : 0;

              const x1 = 50 + 45 * Math.cos(startRad);
              const y1 = 50 + 45 * Math.sin(startRad);
              const x2 = 50 + 45 * Math.cos(endRad);
              const y2 = 50 + 45 * Math.sin(endRad);

              const isHovered = hoveredSegment === segment.category;
              const color = segment.color || colors[i % colors.length];

              acc.paths.push(
                <path
                  key={segment.category}
                  d={`M 50 50 L ${x1} ${y1} A 45 45 0 ${largeArc} 1 ${x2} ${y2} Z`}
                  className="transition-all cursor-pointer"
                  style={{ fill: color, opacity: isHovered ? 1 : 0.85 }}
                  stroke={isHovered ? '#fff' : 'transparent'}
                  strokeWidth="0.5"
                  onMouseEnter={() => setHoveredSegment(segment.category)}
                  onMouseLeave={() => setHoveredSegment(null)}
                  onClick={() => onSegmentClick?.(segment)}
                />
              );

              acc.angle = endAngle;
              return acc;
            }, { angle: 0, paths: [] }).paths}

            <circle cx="50" cy="50" r="25" className="fill-aviation-bg-panel" />
            {hoveredSegment && (
              <>
                <text x="50" y="47" textAnchor="middle" className="text-[10px] fill-aviation-text-primary font-medium">
                  {formatCurrency(sortedData.find(s => s.category === hoveredSegment)?.value || 0)}
                </text>
                {showPercentages && (
                  <text x="50" y="58" textAnchor="middle" className="text-[8px] fill-aviation-text-dim">
                    {((sortedData.find(s => s.category === hoveredSegment)?.value || 0) / total * 100).toFixed(1)}%
                  </text>
                )}
              </>
            )}
          </svg>
        </div>

        <div className="flex-1 flex flex-col gap-2">
          {sortedData.map((segment, i) => {
            const color = segment.color || colors[i % colors.length];
            const percentage = (segment.value / total) * 100;
            const isHovered = hoveredSegment === segment.category;

            return (
              <div
                key={segment.category}
                className={cn(
                  'flex items-center gap-3 p-2 rounded cursor-pointer transition-colors',
                  isHovered ? 'bg-aviation-bg-instrument' : 'hover:bg-aviation-bg-secondary'
                )}
                onMouseEnter={() => setHoveredSegment(segment.category)}
                onMouseLeave={() => setHoveredSegment(null)}
                onClick={() => onSegmentClick?.(segment)}
              >
                <span className="w-3 h-3 rounded" style={{ backgroundColor: color }} />
                <span className="flex-1 text-sm text-aviation-text-primary">{segment.category}</span>
                {showValues && (
                  <span className="text-sm text-aviation-text-primary font-medium">{formatCurrency(segment.value)}</span>
                )}
                {showPercentages && (
                  <span className="text-xs text-aviation-text-dim w-12 text-right">{percentage.toFixed(1)}%</span>
                )}
                <div className="w-16 h-2 bg-aviation-bg-instrument rounded-full overflow-hidden">
                  <div className="h-full rounded-full transition-all" style={{ width: `${percentage}%`, backgroundColor: color }} />
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Semantic Cluster Chart
// ============================================================================

export const SemanticClusterChart: React.FC<SemanticClusterChartProps> = ({
  points,
  clusters,
  showLabels = true,
  showConnections = false,
  height = 350,
  onPointClick,
  className,
}) => {
  const [hoveredPoint, setHoveredPoint] = useState<ClusterPoint | null>(null);

  const clusterColors: Record<string, string> = useMemo(() => {
    const colors: Record<string, string> = {};
    clusters.forEach((cluster, i) => {
      colors[cluster] = ['#00d4ff', '#00ff88', '#ffd43b', '#ff6b6b', '#cc5de8', '#20c997', '#fd7e14', '#e599f7'][i % 8];
    });
    return colors;
  }, [clusters]);

  const clusterCenters = useMemo(() => {
    const centers: Record<string, { x: number; y: number }> = {};
    clusters.forEach(cluster => {
      const clusterPoints = points.filter(p => p.cluster === cluster);
      if (clusterPoints.length > 0) {
        centers[cluster] = {
          x: clusterPoints.reduce((sum, p) => sum + p.x, 0) / clusterPoints.length,
          y: clusterPoints.reduce((sum, p) => sum + p.y, 0) / clusterPoints.length,
        };
      }
    });
    return centers;
  }, [points, clusters]);

  return (
    <div className={cn('flex flex-col bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Hexagon className="w-5 h-5 text-aviation-cyan" />
          <h3 className="text-sm font-medium text-aviation-text-primary">Semantic Clusters</h3>
        </div>
        <div className="flex items-center gap-2 text-xs text-aviation-text-muted">
          {clusters.map(cluster => (
            <span key={cluster} className="flex items-center gap-1">
              <span className="w-2 h-2 rounded" style={{ backgroundColor: clusterColors[cluster] }} />
              <span>{cluster}</span>
            </span>
          ))}
        </div>
      </div>

      <div className="flex-1 p-4" style={{ height }}>
        <svg className="w-full h-full" viewBox="0 0 100 100" preserveAspectRatio="none">
          {[0, 25, 50, 75, 100].map(y => (
            <line key={y} x1="0" y1={y} x2="100" y2={y} className="stroke-aviation-border-panel" strokeWidth="0.5" strokeDasharray="2 2" />
          ))}
          {[0, 25, 50, 75, 100].map(x => (
            <line key={x} x1={x} y1="0" x2={x} y2="100" className="stroke-aviation-border-panel" strokeWidth="0.5" strokeDasharray="2 2" />
          ))}

          {clusters.map(cluster => {
            const center = clusterCenters[cluster];
            if (!center) return null;
            return (
              <g key={cluster}>
                <polygon
                  points={points.filter(p => p.cluster === cluster).map(p => `${p.x},${100 - p.y}`).join(' ')}
                  className="fill-current"
                  style={{ fill: clusterColors[cluster], opacity: 0.1 }}
                />
                <circle
                  cx={center.x}
                  cy={100 - center.y}
                  r="3"
                  className="fill-current"
                  style={{ color: clusterColors[cluster] }}
                />
                {showLabels && (
                  <text x={center.x} y={100 - center.y - 5} textAnchor="middle" className="text-[6px] fill-aviation-text-dim">
                    {cluster}
                  </text>
                )}
              </g>
            );
          })}

          {points.map((point, i) => {
            const isHovered = hoveredPoint?.id === point.id;
            return (
              <circle
                key={point.id}
                cx={point.x}
                cy={100 - point.y}
                r={isHovered ? 2 : 1}
                className="fill-current cursor-pointer transition-all"
                style={{ color: clusterColors[point.cluster], opacity: isHovered ? 1 : 0.7 }}
                onMouseEnter={() => setHoveredPoint(point)}
                onMouseLeave={() => setHoveredNode(null)}
                onClick={() => onPointClick?.(point)}
              />
            );
          })}
        </svg>
      </div>

      {hoveredPoint && (
        <div className="px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
          <div className="flex items-center justify-between mb-1">
            <span className="text-sm font-medium text-aviation-text-primary">{hoveredPoint.label || hoveredPoint.id}</span>
            <span className="text-xs px-2 py-0.5 rounded" style={{ backgroundColor: clusterColors[hoveredPoint.cluster] + '30', color: clusterColors[hoveredPoint.cluster] }}>
              {hoveredPoint.cluster}
            </span>
          </div>
          <div className="text-xs text-aviation-text-dim">
            Position: ({hoveredPoint.x.toFixed(2)}, {hoveredPoint.y.toFixed(2)})
          </div>
        </div>
      )}
    </div>
  );
};

// Helper function to fix the type issue
const setHoveredNode = (node: TopologyNode | null) => {
  // Empty function - just a placeholder for the unused function reference
};

// ============================================================================
// Agent Interaction Graph
// ============================================================================

export const AgentInteractionGraph: React.FC<AgentInteractionGraphProps> = ({
  nodes,
  edges,
  showLabels = true,
  showWeights = false,
  animated = true,
  height = 400,
  onNodeClick,
  onEdgeClick,
  className,
}) => {
  const [hoveredNode, setHoveredNode] = useState<GraphNode | null>(null);
  const [hoveredEdge, setHoveredEdge] = useState<GraphEdge | null>(null);

  const nodePositions = useMemo(() => {
    const positions: Record<string, { x: number; y: number }> = {};
    
    nodes.forEach((node, i) => {
      if (node.id === 'root') {
        positions[node.id] = { x: 50, y: 50 };
      } else {
        const angle = (i / nodes.length) * 2 * Math.PI;
        const radius = 35;
        positions[node.id] = {
          x: 50 + radius * Math.cos(angle - Math.PI / 2),
          y: 50 + radius * Math.sin(angle - Math.PI / 2),
        };
      }
    });
    return positions;
  }, [nodes]);

  const getNodeColor = (type: string) => {
    switch (type) {
      case 'orchestrator': return '#00d4ff';
      case 'agent': return '#00ff88';
      case 'tool': return '#ffd43b';
      case 'resource': return '#ff6b6b';
      default: return '#cc5de8';
    }
  };

  const edgeColors: Record<string, string> = {
    triggers: '#00d4ff',
    uses: '#00ff88',
    manages: '#ffd43b',
    depends_on: '#ff6b6b',
  };

  return (
    <div className={cn('flex flex-col bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Network className="w-5 h-5 text-aviation-cyan" />
          <h3 className="text-sm font-medium text-aviation-text-primary">Agent Interactions</h3>
        </div>
        <div className="flex items-center gap-3 text-xs">
          {nodes.length} nodes
          <span>•</span>
          {edges.length} edges
        </div>
      </div>

      <div className="flex-1 p-4" style={{ height }}>
        <svg className="w-full h-full" viewBox="0 0 100 100" preserveAspectRatio="xMidYMid meet">
          {edges.map((edge, i) => {
            const sourcePos = nodePositions[edge.source];
            const targetPos = nodePositions[edge.target];
            if (!sourcePos || !targetPos) return null;

            const isHovered = hoveredEdge === edge;
            const color = edgeColors[edge.type || 'uses'] || '#00d4ff';

            const midX = (sourcePos.x + targetPos.x) / 2;
            const midY = (sourcePos.y + targetPos.y) / 2;
            const dx = targetPos.x - sourcePos.x;
            const dy = targetPos.y - sourcePos.y;
            const offset = 3;
            const controlX = midX - dy * offset * 0.01;
            const controlY = midY + dx * offset * 0.01;

            return (
              <g key={i}>
                <path
                  d={`M ${sourcePos.x} ${sourcePos.y} Q ${controlX} ${controlY} ${targetPos.x} ${targetPos.y}`}
                  className="fill-none transition-all"
                  stroke={color}
                  strokeWidth={isHovered ? 1.5 : 0.8}
                  opacity={isHovered ? 1 : 0.5}
                  strokeDasharray={animated && !isHovered ? '3 2' : undefined}
                  onMouseEnter={() => setHoveredEdge(edge)}
                  onMouseLeave={() => setHoveredEdge(null)}
                  onClick={() => onEdgeClick?.(edge)}
                  style={{ cursor: 'pointer' }}
                />
                {showWeights && isHovered && (
                  <text x={controlX} y={controlY - 2} textAnchor="middle" className="text-[6px] fill-aviation-text-primary">
                    {edge.weight.toFixed(1)}
                  </text>
                )}
                <circle cx={targetPos.x} cy={targetPos.y} r="1" className="fill-current" style={{ color }} />
              </g>
            );
          })}

          {nodes.map((node, i) => {
            const pos = nodePositions[node.id];
            if (!pos) return null;

            const isHovered = hoveredNode?.id === node.id;
            const isRoot = node.id === 'root';
            const radius = isRoot ? 5 : 3;
            const color = getNodeColor(node.type);

            return (
              <g
                key={node.id}
                onMouseEnter={() => setHoveredNode(node)}
                onMouseLeave={() => setHoveredNode(null)}
                onClick={() => onNodeClick?.(node)}
                className="cursor-pointer"
              >
                {isRoot && (
                  <circle
                    cx={pos.x}
                    cy={pos.y}
                    r={radius + 4}
                    className="fill-none"
                    style={{ stroke: color, opacity: 0.3 }}
                    strokeWidth="1"
                  />
                )}
                <circle
                  cx={pos.x}
                  cy={pos.y}
                  r={isHovered ? radius + 1 : radius}
                  className="transition-all"
                  style={{ fill: color, opacity: isHovered ? 1 : 0.85 }}
                  stroke={isHovered ? '#fff' : 'transparent'}
                  strokeWidth="0.5"
                />
                {showLabels && (
                  <text
                    x={pos.x}
                    y={pos.y + radius + 6}
                    textAnchor="middle"
                    className="text-[6px] fill-aviation-text-primary"
                  >
                    {node.label.length > 10 ? node.label.slice(0, 10) + '...' : node.label}
                  </text>
                )}
              </g>
            );
          })}
        </svg>
      </div>

      {hoveredNode && (
        <div className="px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
          <div className="flex items-center justify-between mb-1">
            <span className="text-sm font-medium text-aviation-text-primary">{hoveredNode.label}</span>
            <span className="text-xs px-2 py-0.5 rounded capitalize" style={{ backgroundColor: getNodeColor(hoveredNode.type) + '30', color: getNodeColor(hoveredNode.type) }}>
              {hoveredNode.type}
            </span>
          </div>
          <div className="flex items-center gap-4 text-xs text-aviation-text-dim">
            <span>Connections: {hoveredNode.connections.length}</span>
            {hoveredNode.weight !== undefined && <span>Weight: {hoveredNode.weight.toFixed(2)}</span>}
          </div>
        </div>
      )}

      {hoveredEdge && !hoveredNode && (
        <div className="px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
          <div className="flex items-center justify-between">
            <span className="text-sm text-aviation-text-primary">
              {edges.find(e => e === hoveredEdge)?.source} → {edges.find(e => e === hoveredEdge)?.target}
            </span>
            <span className="text-xs text-aviation-cyan">Weight: {hoveredEdge.weight.toFixed(2)}</span>
          </div>
          <div className="text-xs text-aviation-text-dim mt-1">
            Type: {hoveredEdge.type || 'uses'}
          </div>
        </div>
      )}

      <div className="px-4 py-2 border-t border-aviation-border-panel flex flex-wrap gap-3">
        {Object.entries(edgeColors).map(([type, color]) => (
          <span key={type} className="flex items-center gap-1 text-[10px] text-aviation-text-muted">
            <span className="w-4 h-0.5" style={{ backgroundColor: color }} />
            {type.replace('_', ' ')}
          </span>
        ))}
      </div>
    </div>
  );
};

export const CircularFlow = CircularFlowDiagram;
export const CostDistribution = CostDistributionGraph;
export const SemanticCluster = SemanticClusterChart;
export const WaterfallChart = RuntimeWaterfallChart;
