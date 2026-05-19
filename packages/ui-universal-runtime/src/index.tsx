/**
 * @functionfly/ui-universal-runtime
 * Universal Runtime Components - Multi-runtime execution visualization
 */

import React, { useState, useMemo, useEffect } from 'react';
import { cn } from '@functionfly/ui-core';
import {
  Server,
  Cpu,
  Monitor,
  Globe,
  Cloud,
  Zap,
  Activity,
  Clock,
  CheckCircle,
  XCircle,
  AlertTriangle,
  Play,
  Pause,
  Square,
  RefreshCw,
  ChevronRight,
  GitBranch,
  Layers,
  Network,
  Database,
  Bot,
  Brain,
  Box,
  Radio,
  Broadcast,
  Gauge,
  Timer,
  TrendingUp,
  TrendingDown,
  ArrowRight,
  Settings,
  Eye,
  BarChart3,
  PieChart,
  LineChart,
  SquareDot,
  CircleDot,
  Hexagon,
} from 'lucide-react';

// ============================================================================
// Runtime Abstraction UI
// ============================================================================

export const RuntimeAbstractionUI: React.FC<RuntimeAbstractionUIProps> = ({
  instances,
  selectedInstanceId = null,
  onInstanceSelect,
  onInstanceHover,
  className,
}) => {
  const [hoveredId, setHoveredId] = useState<string | null>(null);

  const getStatusColor = (status: RuntimeStatus) => {
    switch (status) {
      case 'ready': return 'text-green-400';
      case 'busy': return 'text-amber-400';
      case 'error': return 'text-red-400';
      default: return 'text-aviation-text-muted';
    }
  };

  const getRuntimeIcon = (type: RuntimeType) => {
    switch (type) {
      case 'wasm': return <Box className="w-4 h-4" />;
      case 'native': return <Cpu className="w-4 h-4" />;
      case 'serverless': return <Cloud className="w-4 h-4" />;
      case 'browser': return <Monitor className="w-4 h-4" />;
      case 'edge': return <Globe className="w-4 h-4" />;
      case 'gpu': return <Zap className="w-4 h-4" />;
    }
  };

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Server className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Runtime Abstraction</h3>
          </div>
          <span className="text-xs text-aviation-text-dim">{instances.length} instances</span>
        </div>
      </div>

      <div className="flex-1 overflow-auto p-4">
        <div className="grid grid-cols-2 gap-3">
          {instances.map(instance => {
            const isSelected = selectedInstanceId === instance.id;
            const isHovered = hoveredId === instance.id;

            return (
              <div
                key={instance.id}
                onClick={() => onInstanceSelect?.(instance)}
                onMouseEnter={() => { setHoveredId(instance.id); onInstanceHover?.(instance); }}
                onMouseLeave={() => { setHoveredId(null); onInstanceHover?.(null); }}
                className={cn(
                  'p-4 rounded-lg border cursor-pointer transition-all',
                  isSelected ? 'bg-aviation-bg-instrument border-aviation-cyan' : 'bg-aviation-bg-secondary border-aviation-border-panel hover:border-aviation-text-muted'
                )}
              >
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-2">
                    <span className="text-aviation-cyan">{getRuntimeIcon(instance.type)}</span>
                    <span className="text-sm font-medium text-aviation-text-primary">{instance.name}</span>
                  </div>
                  <span className={cn('text-xs uppercase font-medium', getStatusColor(instance.status))}>
                    {instance.status}
                  </span>
                </div>

                <div className="grid grid-cols-2 gap-2 text-[10px]">
                  <div>
                    <span className="text-aviation-text-dim">CPU</span>
                    <div className="w-full h-1 bg-aviation-bg-instrument rounded-full mt-1">
                      <div className="h-full bg-aviation-cyan rounded-full" style={{ width: `${instance.metrics.cpuUsage}%` }} />
                    </div>
                    <span className="text-aviation-text-primary">{instance.metrics.cpuUsage.toFixed(0)}%</span>
                  </div>
                  <div>
                    <span className="text-aviation-text-dim">Memory</span>
                    <div className="w-full h-1 bg-aviation-bg-instrument rounded-full mt-1">
                      <div className="h-full bg-purple-500 rounded-full" style={{ width: `${instance.metrics.memoryUsage}%` }} />
                    </div>
                    <span className="text-aviation-text-primary">{instance.metrics.memoryUsage.toFixed(0)}%</span>
                  </div>
                </div>

                <div className="flex items-center justify-between mt-3 text-[10px]">
                  <span className="text-aviation-text-dim">{instance.region || 'Local'}</span>
                  <span className="text-aviation-text-primary">{instance.metrics.executionCount} exec</span>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {hoveredId && instances.find(i => i.id === hoveredId) && (
        <div className="px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
          <div className="grid grid-cols-4 gap-3 text-xs">
            <div>
              <span className="text-aviation-text-dim">Avg Latency</span>
              <span className="ml-2 text-aviation-text-primary">{instances.find(i => i.id === hoveredId)!.metrics.avgLatency.toFixed(1)}ms</span>
            </div>
            <div>
              <span className="text-aviation-text-dim">Error Rate</span>
              <span className="ml-2 text-aviation-text-primary">{instances.find(i => i.id === hoveredId)!.metrics.errorRate.toFixed(2)}%</span>
            </div>
            {instances.find(i => i.id === hoveredId)!.metrics.gpuUsage !== undefined && (
              <div>
                <span className="text-aviation-text-dim">GPU</span>
                <span className="ml-2 text-aviation-text-primary">{instances.find(i => i.id === hoveredId)!.metrics.gpuUsage!.toFixed(0)}%</span>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

// ============================================================================
// WebAssembly Execution Panel
// ============================================================================

export const WasmExecutionPanel: React.FC<WasmExecutionPanelProps> = ({
  executions,
  selectedExecutionId = null,
  onExecutionSelect,
  onRefresh,
  className,
}) => {
  const formatBytes = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Box className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">WebAssembly Execution</h3>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-xs text-aviation-text-dim">{executions.length} executions</span>
            <button onClick={onRefresh} className="p-1.5 hover:bg-aviation-bg-instrument rounded transition-colors">
              <RefreshCw className="w-4 h-4 text-aviation-text-muted" />
            </button>
          </div>
        </div>
      </div>

      <div className="flex-1 overflow-auto">
        {executions.map(exec => {
          const isSelected = selectedExecutionId === exec.id;
          return (
            <div
              key={exec.id}
              onClick={() => onExecutionSelect?.(exec)}
              className={cn(
                'p-4 border-b border-aviation-border-panel cursor-pointer transition-colors',
                isSelected ? 'bg-aviation-bg-instrument' : 'hover:bg-aviation-bg-secondary'
              )}
            >
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-aviation-text-primary">{exec.moduleName}</span>
                  <ChevronRight className="w-3 h-3 text-aviation-text-dim" />
                  <span className="text-xs text-aviation-cyan">{exec.functionName}</span>
                </div>
                <span className="text-xs text-aviation-text-dim">
                  {new Date(exec.timestamp).toLocaleTimeString()}
                </span>
              </div>

              <div className="grid grid-cols-4 gap-3 text-[10px]">
                <div>
                  <span className="text-aviation-text-dim">Input</span>
                  <span className="ml-1 text-aviation-text-primary">{formatBytes(exec.inputSize)}</span>
                </div>
                <div>
                  <span className="text-aviation-text-dim">Output</span>
                  <span className="ml-1 text-aviation-text-primary">{formatBytes(exec.outputSize)}</span>
                </div>
                <div>
                  <span className="text-aviation-text-dim">Time</span>
                  <span className="ml-1 text-aviation-text-primary">{exec.executionTime.toFixed(2)}ms</span>
                </div>
                <div>
                  <span className="text-aviation-text-dim">Memory</span>
                  <span className="ml-1 text-aviation-text-primary">{formatBytes(exec.memoryUsed)}</span>
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

// ============================================================================
// GPU Kernel Inspector
// ============================================================================

export const GPUKernelInspector: React.FC<GPUKernelInspectorProps> = ({
  kernels,
  selectedKernelId = null,
  onKernelSelect,
  onKernelLaunch,
  className,
}) => {
  const getStatusColor = (status: KernelStatus) => {
    switch (status) {
      case 'running': return 'text-green-400';
      case 'queued': return 'text-amber-400';
      case 'completed': return 'text-aviation-cyan';
      default: return 'text-aviation-text-muted';
    }
  };

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Zap className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">GPU Kernel Inspector</h3>
          </div>
          <span className="text-xs text-aviation-text-dim">{kernels.length} kernels</span>
        </div>
      </div>

      <div className="flex-1 overflow-auto">
        {kernels.map(kernel => {
          const isSelected = selectedKernelId === kernel.id;
          return (
            <div
              key={kernel.id}
              onClick={() => onKernelSelect?.(kernel)}
              className={cn(
                'p-4 border-b border-aviation-border-panel cursor-pointer transition-colors',
                isSelected ? 'bg-aviation-bg-instrument border-l-2 border-l-aviation-cyan' : 'hover:bg-aviation-bg-secondary'
              )}
            >
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-aviation-text-primary">{kernel.name}</span>
                  <span className={cn('text-[10px] uppercase font-medium', getStatusColor(kernel.status))}>
                    {kernel.status}
                  </span>
                </div>
                <button
                  onClick={(e) => { e.stopPropagation(); onKernelLaunch?.(kernel.id); }}
                  className="p-1.5 bg-aviation-cyan/20 hover:bg-aviation-cyan/30 rounded transition-colors"
                >
                  <Play className="w-3 h-3 text-aviation-cyan" />
                </button>
              </div>

              <div className="grid grid-cols-2 gap-2 text-[10px] mb-2">
                <div className="flex items-center gap-2">
                  <span className="text-aviation-text-dim">Grid:</span>
                  <span className="text-aviation-text-primary">[{kernel.gridSize.join(', ')}]</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-aviation-text-dim">Block:</span>
                  <span className="text-aviation-text-primary">[{kernel.blockSize.join(', ')}]</span>
                </div>
              </div>

              <div className="grid grid-cols-3 gap-3 text-[10px]">
                <div>
                  <span className="text-aviation-text-dim">Execution</span>
                  <span className="ml-1 text-aviation-text-primary">{kernel.executionTime.toFixed(2)}ms</span>
                </div>
                <div>
                  <span className="text-aviation-text-dim">Memory</span>
                  <span className="ml-1 text-aviation-text-primary">{kernel.memoryUsage}KB</span>
                </div>
                {kernel.registersPerThread && (
                  <div>
                    <span className="text-aviation-text-dim">Registers</span>
                    <span className="ml-1 text-aviation-text-primary">{kernel.registersPerThread}</span>
                  </div>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

// ============================================================================
// Serverless Runtime Viewer
// ============================================================================

export const ServerlessRuntimeViewer: React.FC<ServerlessRuntimeViewerProps> = ({
  functions,
  selectedFunctionId = null,
  onFunctionSelect,
  className,
}) => {
  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Cloud className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Serverless Runtime</h3>
          </div>
          <span className="text-xs text-aviation-text-dim">{functions.length} functions</span>
        </div>
      </div>

      <div className="flex-1 overflow-auto p-4">
        <div className="space-y-3">
          {functions.map(fn => {
            const isSelected = selectedFunctionId === fn.id;
            return (
              <div
                key={fn.id}
                onClick={() => onFunctionSelect?.(fn)}
                className={cn(
                  'p-4 rounded-lg border cursor-pointer transition-all',
                  isSelected ? 'bg-aviation-bg-instrument border-aviation-cyan' : 'bg-aviation-bg-secondary border-aviation-border-panel hover:border-aviation-text-muted'
                )}
              >
                <div className="flex items-center justify-between mb-3">
                  <span className="text-sm font-medium text-aviation-text-primary">{fn.name}</span>
                  <div className="flex items-center gap-2">
                    <span className="px-1.5 py-0.5 bg-aviation-bg-instrument rounded text-[10px] text-aviation-text-muted">
                      {fn.runtime}
                    </span>
                    <span className="text-[10px] text-aviation-text-dim">
                      {fn.memory}MB
                    </span>
                  </div>
                </div>

                <div className="grid grid-cols-3 gap-3 text-[10px]">
                  <div>
                    <span className="text-aviation-text-dim">Invocations</span>
                    <div className="text-aviation-text-primary font-medium">{fn.invocationCount.toLocaleString()}</div>
                  </div>
                  <div>
                    <span className="text-aviation-text-dim">Avg Duration</span>
                    <div className="text-aviation-text-primary font-medium">{fn.avgDuration.toFixed(0)}ms</div>
                  </div>
                  <div>
                    <span className="text-aviation-text-dim">Error Rate</span>
                    <div className={cn('font-medium', fn.errorRate > 1 ? 'text-red-400' : 'text-aviation-text-primary')}>
                      {fn.errorRate.toFixed(2)}%
                    </div>
                  </div>
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
// Browser Agent Session
// ============================================================================

export const BrowserAgentSession: React.FC<BrowserAgentSessionProps> = ({
  session,
  onSessionPause,
  onSessionResume,
  onSessionStop,
  className,
}) => {
  const [expandedAction, setExpandedAction] = useState<number | null>(null);

  if (!session) {
    return (
      <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
        <div className="px-4 py-3 border-b border-aviation-border-panel">
          <div className="flex items-center gap-2">
            <Bot className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Browser Agent Session</h3>
          </div>
        </div>
        <div className="flex-1 flex items-center justify-center text-aviation-text-muted">
          <p className="text-sm">No active session</p>
        </div>
      </div>
    );
  }

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Bot className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Browser Agent Session</h3>
          </div>
          <div className="flex items-center gap-2">
            {session.status === 'active' && (
              <div className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />
            )}
            <span className="text-xs uppercase text-aviation-text-dim">{session.status}</span>
          </div>
        </div>
      </div>

      <div className="px-4 py-2 border-b border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="flex items-center justify-between text-xs">
          <div className="flex items-center gap-2">
            <span className="text-aviation-text-dim">Session:</span>
            <span className="text-aviation-text-primary font-mono">{session.id.slice(0, 8)}</span>
          </div>
          {session.currentUrl && (
            <span className="text-aviation-text-dim truncate ml-2">{session.currentUrl}</span>
          )}
        </div>
      </div>

      <div className="flex-1 overflow-auto">
        {session.actions.map((action, i) => (
          <div
            key={i}
            onClick={() => setExpandedAction(expandedAction === i ? null : i)}
            className="p-3 border-b border-aviation-border-panel hover:bg-aviation-bg-secondary cursor-pointer transition-colors"
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                {action.success ? (
                  <CheckCircle className="w-3 h-3 text-green-400" />
                ) : (
                  <XCircle className="w-3 h-3 text-red-400" />
                )}
                <span className="text-xs text-aviation-text-primary">{action.type}</span>
              </div>
              <div className="flex items-center gap-3 text-[10px] text-aviation-text-dim">
                <span>{new Date(action.timestamp).toLocaleTimeString()}</span>
                {action.duration && <span>{action.duration.toFixed(0)}ms</span>}
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className="px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="flex items-center justify-center gap-3">
          {session.status === 'active' && (
            <button
              onClick={onSessionPause}
              className="flex items-center gap-2 px-3 py-1.5 bg-aviation-amber text-aviation-bg-primary rounded-lg hover:bg-aviation-amber/90 transition-colors text-xs"
            >
              <Pause className="w-3 h-3" />
              Pause
            </button>
          )}
          {session.status === 'paused' && (
            <button
              onClick={onSessionResume}
              className="flex items-center gap-2 px-3 py-1.5 bg-aviation-cyan text-aviation-bg-primary rounded-lg hover:bg-aviation-cyan/90 transition-colors text-xs"
            >
              <Play className="w-3 h-3" />
              Resume
            </button>
          )}
          {(session.status === 'active' || session.status === 'paused') && (
            <button
              onClick={onSessionStop}
              className="flex items-center gap-2 px-3 py-1.5 bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors text-xs"
            >
              <Square className="w-3 h-3" />
              Stop
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Edge Runtime Map
// ============================================================================

export const EdgeRuntimeMap: React.FC<EdgeRuntimeMapProps> = ({
  nodes,
  selectedNodeId = null,
  onNodeSelect,
  onNodeHover,
  className,
}) => {
  const nodePositions = useMemo(() => {
    const positions: Record<string, { x: number; y: number }> = {};
    const centerX = 250;
    const centerY = 150;
    nodes.forEach((node, i) => {
      const angle = (2 * Math.PI * i) / nodes.length;
      const radius = 100;
      positions[node.id] = {
        x: centerX + radius * Math.cos(angle),
        y: centerY + radius * Math.sin(angle),
      };
    });
    return positions;
  }, [nodes]);

  const getStatusColor = (status: RuntimeStatus) => {
    switch (status) {
      case 'ready': return { fill: 'fill-green-500/20', stroke: 'stroke-green-500', text: 'text-green-400' };
      case 'busy': return { fill: 'fill-amber-500/20', stroke: 'stroke-amber-500', text: 'text-amber-400' };
      case 'error': return { fill: 'fill-red-500/20', stroke: 'stroke-red-500', text: 'text-red-400' };
      default: return { fill: 'fill-aviation-bg-instrument', stroke: 'stroke-aviation-text-muted', text: 'text-aviation-text-muted' };
    }
  };

  return (
    <div className={cn('relative h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="absolute top-3 left-3 flex items-center gap-2 z-10">
        <Globe className="w-5 h-5 text-aviation-cyan" />
        <span className="text-sm font-medium text-aviation-text-primary">Edge Runtime Topology</span>
      </div>

      <svg className="w-full h-full" viewBox="0 0 500 300">
        {nodes.map((node, i) => {
          const pos = nodePositions[node.id];
          if (!pos) return null;
          const colors = getStatusColor(node.status);
          const isSelected = selectedNodeId === node.id;

          return (
            <g
              key={node.id}
              onClick={() => onNodeSelect?.(node)}
              onMouseEnter={() => onNodeHover?.(node)}
              onMouseLeave={() => onNodeHover?.(null)}
              className="cursor-pointer"
            >
              <circle
                cx={pos.x}
                cy={pos.y}
                r={isSelected ? 35 : 30}
                className={cn('transition-all', colors.fill, colors.stroke)}
                strokeWidth={isSelected ? 3 : 2}
              />
              <text x={pos.x} y={pos.y - 40} textAnchor="middle" className="text-[10px] fill-aviation-text-primary">
                {node.name}
              </text>
              <text x={pos.x} y={pos.y - 28} textAnchor="middle" className="text-[8px] fill-aviation-text-dim">
                {node.location}
              </text>
              <text x={pos.x} y={pos.y + 5} textAnchor="middle" className={cn('text-[10px] font-bold', colors.text)}>
                {node.latency}ms
              </text>
            </g>
          );
        })}

        {nodes.map((node, i) => {
          const nextNode = nodes[(i + 1) % nodes.length];
          const posA = nodePositions[node.id];
          const posB = nodePositions[nextNode.id];
          if (!posA || !posB) return null;

          return (
            <line
              key={`${node.id}-${nextNode.id}`}
              x1={posA.x}
              y1={posA.y}
              x2={posB.x}
              y2={posB.y}
              className="stroke-aviation-border-panel"
              strokeWidth={1}
              strokeDasharray="4 4"
            />
          );
        })}
      </svg>

      <div className="absolute bottom-3 right-3 flex items-center gap-3 px-3 py-2 bg-aviation-bg-secondary/90 rounded-lg border border-aviation-border-panel">
        {['ready', 'busy', 'error', 'offline'].map(status => (
          <div key={status} className="flex items-center gap-1 text-[10px]">
            <div className={cn('w-2 h-2 rounded-full', getStatusColor(status as RuntimeStatus).stroke.replace('stroke-', 'bg-'))} />
            <span className="text-aviation-text-dim capitalize">{status}</span>
          </div>
        ))}
      </div>
    </div>
  );
};

// ============================================================================
// Hybrid Execution Orchestrator
// ============================================================================

export const HybridExecutionOrchestrator: React.FC<HybridExecutionOrchestratorProps> = ({
  tasks,
  selectedTaskId = null,
  onTaskSelect,
  onTaskCancel,
  className,
}) => {
  const getRuntimeIcon = (type: RuntimeType) => {
    switch (type) {
      case 'wasm': return <Box className="w-3 h-3" />;
      case 'native': return <Cpu className="w-3 h-3" />;
      case 'serverless': return <Cloud className="w-3 h-3" />;
      case 'browser': return <Monitor className="w-3 h-3" />;
      case 'edge': return <Globe className="w-3 h-3" />;
      case 'gpu': return <Zap className="w-3 h-3" />;
    }
  };

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Layers className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Hybrid Execution</h3>
          </div>
          <span className="text-xs text-aviation-text-dim">{tasks.length} tasks</span>
        </div>
      </div>

      <div className="flex-1 overflow-auto">
        {tasks.map(task => {
          const isSelected = selectedTaskId === task.id;
          return (
            <div
              key={task.id}
              onClick={() => onTaskSelect?.(task)}
              className={cn(
                'p-4 border-b border-aviation-border-panel cursor-pointer transition-colors',
                isSelected ? 'bg-aviation-bg-instrument' : 'hover:bg-aviation-bg-secondary'
              )}
            >
              <div className="flex items-center justify-between mb-3">
                <span className="text-sm font-medium text-aviation-text-primary">{task.name}</span>
                <div className="flex items-center gap-2">
                  <span className="text-xs text-aviation-text-dim">
                    Stage {task.currentStage + 1}/{task.stages.length}
                  </span>
                  <button
                    onClick={(e) => { e.stopPropagation(); onTaskCancel?.(task.id); }}
                    className="p-1 hover:bg-aviation-bg-panel rounded"
                  >
                    <Square className="w-3 h-3 text-aviation-text-muted" />
                  </button>
                </div>
              </div>

              <div className="flex items-center gap-1">
                {task.stages.map((stage, i) => (
                  <React.Fragment key={i}>
                    <div className={cn(
                      'flex items-center justify-center w-8 h-8 rounded border',
                      stage.status === 'completed' ? 'bg-green-500/20 border-green-500 text-green-400' :
                      stage.status === 'running' ? 'bg-aviation-cyan/20 border-aviation-cyan text-aviation-cyan' :
                      stage.status === 'queued' ? 'bg-amber-500/20 border-amber-500 text-amber-400' :
                      'bg-aviation-bg-instrument border-aviation-border-panel text-aviation-text-muted'
                    )}>
                      {getRuntimeIcon(stage.runtimeType)}
                    </div>
                    {i < task.stages.length - 1 && (
                      <ArrowRight className="w-3 h-3 text-aviation-text-dim" />
                    )}
                  </React.Fragment>
                ))}
              </div>

              <div className="flex items-center justify-between mt-2 text-[10px] text-aviation-text-dim">
                <span>Total: {task.totalDuration.toFixed(0)}ms</span>
                <span className="capitalize">{task.stages[task.currentStage]?.runtimeType}</span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

// ============================================================================
// Cross-Cloud Topology Map
// ============================================================================

export const CrossCloudTopologyMap: React.FC<CrossCloudTopologyMapProps> = ({
  nodes,
  selectedNodeId = null,
  onNodeSelect,
  className,
}) => {
  const getProviderColor = (provider: CloudProvider) => {
    switch (provider) {
      case 'aws': return 'fill-orange-500/20 stroke-orange-500';
      case 'gcp': return 'fill-blue-500/20 stroke-blue-500';
      case 'azure': return 'fill-blue-600/20 stroke-blue-600';
      case 'cloudflare': return 'fill-orange-400/20 stroke-orange-400';
      case 'local': return 'fill-green-500/20 stroke-green-500';
    }
  };

  const nodePositions = useMemo(() => {
    const positions: Record<string, { x: number; y: number }> = {};
    const centerX = 300;
    const centerY = 200;
    const innerRadius = 80;
    const outerRadius = 150;

    nodes.forEach((node, i) => {
      const isOuter = i > Math.floor(nodes.length / 2);
      const radius = isOuter ? outerRadius : innerRadius;
      const offset = isOuter ? Math.PI / nodes.length : 0;
      const angle = (2 * Math.PI * i) / nodes.length + offset;
      positions[node.id] = {
        x: centerX + radius * Math.cos(angle),
        y: centerY + radius * Math.sin(angle),
      };
    });
    return positions;
  }, [nodes]);

  return (
    <div className={cn('relative h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="absolute top-3 left-3 flex items-center gap-2 z-10">
        <Network className="w-5 h-5 text-aviation-cyan" />
        <span className="text-sm font-medium text-aviation-text-primary">Cross-Cloud Topology</span>
      </div>

      <svg className="w-full h-full" viewBox="0 0 600 400">
        {nodes.filter(n => n.connections.length > 0).map(node => {
          const sourcePos = nodePositions[node.id];
          if (!sourcePos) return null;

          return node.connections.map(connId => {
            const targetNode = nodes.find(n => n.id === connId);
            if (!targetNode) return null;
            const targetPos = nodePositions[targetNode.id];
            if (!targetPos) return null;

            return (
              <line
                key={`${node.id}-${connId}`}
                x1={sourcePos.x}
                y1={sourcePos.y}
                x2={targetPos.x}
                y2={targetPos.y}
                className="stroke-aviation-cyan/30"
                strokeWidth={2}
              />
            );
          });
        })}

        {nodes.map(node => {
          const pos = nodePositions[node.id];
          if (!pos) return null;
          const isSelected = selectedNodeId === node.id;
          const colors = getProviderColor(node.provider);
          const statusColors = node.status === 'ready' ? 'fill-green-500' : node.status === 'busy' ? 'fill-amber-500' : 'fill-red-500';

          return (
            <g key={node.id} onClick={() => onNodeSelect?.(node)} className="cursor-pointer">
              <circle
                cx={pos.x}
                cy={pos.y}
                r={isSelected ? 45 : 40}
                className={cn(colors, 'transition-all')}
                strokeWidth={isSelected ? 3 : 2}
              />
              <circle cx={pos.x + 30} cy={pos.y - 30} r={6} className={statusColors} />
              <text x={pos.x} y={pos.y + 5} textAnchor="middle" className="text-[11px] fill-aviation-text-primary font-medium">
                {node.provider.toUpperCase()}
              </text>
              <text x={pos.x} y={pos.y + 20} textAnchor="middle" className="text-[9px] fill-aviation-text-dim">
                {node.region}
              </text>
            </g>
          );
        })}
      </svg>

      <div className="absolute bottom-3 right-3 flex items-center gap-3 px-3 py-2 bg-aviation-bg-secondary/90 rounded-lg border border-aviation-border-panel">
        {(['aws', 'gcp', 'azure', 'cloudflare', 'local'] as CloudProvider[]).map(provider => (
          <div key={provider} className="flex items-center gap-1 text-[10px]">
            <div className={cn('w-3 h-3 rounded', getProviderColor(provider).split(' ')[0].replace('fill-', 'bg-'))} />
            <span className="text-aviation-text-dim uppercase">{provider}</span>
          </div>
        ))}
      </div>
    </div>
  );
};

// ============================================================================
// Model Routing Visualizer
// ============================================================================

export const ModelRoutingVisualizer: React.FC<ModelRoutingVisualizerProps> = ({
  routes,
  selectedRouteId = null,
  onRouteSelect,
  className,
}) => {
  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Brain className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Model Routing</h3>
          </div>
          <span className="text-xs text-aviation-text-dim">{routes.length} routes</span>
        </div>
      </div>

      <div className="flex-1 overflow-auto p-4">
        <svg className="w-full h-full" viewBox="0 0 100 100" preserveAspectRatio="none">
          {[0, 25, 50, 75, 100].map(y => (
            <line key={y} x1="0" y1={y} x2="100" y2={y} className="stroke-aviation-border-panel" strokeWidth="0.5" strokeDasharray="2 2" />
          ))}

          {routes.map((route, i) => {
            const y = (i / (routes.length - 1)) * 80 + 10;
            const width = (route.requestCount / Math.max(...routes.map(r => r.requestCount))) * 80;
            return (
              <g key={route.id} onClick={() => onRouteSelect?.(route)} className="cursor-pointer">
                <rect x="10" y={y} width={width} height="8" rx="2" className="fill-aviation-cyan/30" />
                <rect x="10" y={y} width={width * route.successRate / 100} height="8" rx="2" className="fill-aviation-cyan" />
                <text x="10" y={y - 2} className="text-[6px] fill-aviation-text-primary">{route.modelName}</text>
                <text x={width + 12} y={y + 6} className="text-[5px] fill-aviation-text-dim">{route.provider}</text>
              </g>
            );
          })}
        </svg>

        <div className="mt-4 space-y-2">
          {routes.map(route => (
            <div
              key={route.id}
              onClick={() => onRouteSelect?.(route)}
              className={cn(
                'p-3 rounded-lg border cursor-pointer transition-all',
                selectedRouteId === route.id ? 'bg-aviation-bg-instrument border-aviation-cyan' : 'bg-aviation-bg-secondary border-aviation-border-panel'
              )}
            >
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm text-aviation-text-primary">{route.modelName}</span>
                <span className="px-1.5 py-0.5 bg-aviation-bg-instrument rounded text-[10px] text-aviation-text-muted uppercase">
                  {route.provider}
                </span>
              </div>
              <div className="grid grid-cols-3 gap-2 text-[10px]">
                <div>
                  <span className="text-aviation-text-dim">Requests</span>
                  <span className="ml-1 text-aviation-text-primary">{route.requestCount.toLocaleString()}</span>
                </div>
                <div>
                  <span className="text-aviation-text-dim">Latency</span>
                  <span className="ml-1 text-aviation-text-primary">{route.avgLatency.toFixed(0)}ms</span>
                </div>
                <div>
                  <span className="text-aviation-text-dim">Success</span>
                  <span className={cn('ml-1 font-medium', route.successRate > 95 ? 'text-green-400' : 'text-aviation-text-primary')}>
                    {route.successRate.toFixed(1)}%
                  </span>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Inference Provider Selector
// ============================================================================

export const InferenceProviderSelector: React.FC<InferenceProviderSelectorProps> = ({
  options,
  selectedProvider = null,
  onProviderSelect,
  onCompare,
  className,
}) => {
  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Cpu className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Inference Provider</h3>
          </div>
          <button
            onClick={onCompare}
            className="px-2 py-1 text-xs text-aviation-cyan hover:bg-aviation-cyan/10 rounded transition-colors"
          >
            Compare
          </button>
        </div>
      </div>

      <div className="flex-1 overflow-auto p-4">
        <div className="space-y-3">
          {options.map(option => {
            const isSelected = selectedProvider === option.provider;
            return (
              <div
                key={option.provider}
                onClick={() => option.available && onProviderSelect?.(option.provider)}
                className={cn(
                  'p-4 rounded-lg border cursor-pointer transition-all',
                  isSelected ? 'bg-aviation-bg-instrument border-aviation-cyan' : 
                  option.available ? 'bg-aviation-bg-secondary border-aviation-border-panel hover:border-aviation-text-muted' :
                  'bg-aviation-bg-instrument border-aviation-border-panel opacity-50'
                )}
              >
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-aviation-text-primary">{option.name}</span>
                    {!option.available && (
                      <span className="px-1.5 py-0.5 bg-red-500/20 rounded text-[10px] text-red-400">Unavailable</span>
                    )}
                  </div>
                  <span className="text-xs text-aviation-text-muted uppercase">{option.provider}</span>
                </div>

                <div className="mb-3">
                  <span className="text-[10px] text-aviation-text-dim">Models</span>
                  <div className="flex flex-wrap gap-1 mt-1">
                    {option.models.map(model => (
                      <span key={model} className="px-1.5 py-0.5 bg-aviation-bg-instrument rounded text-[10px] text-aviation-text-muted">
                        {model}
                      </span>
                    ))}
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-3 text-[10px]">
                  <div>
                    <span className="text-aviation-text-dim">Latency</span>
                    <span className="ml-1 text-aviation-text-primary">{option.latency.toFixed(0)}ms</span>
                  </div>
                  <div>
                    <span className="text-aviation-text-dim">Cost</span>
                    <span className="ml-1 text-aviation-text-primary">${option.costPer1kTokens.toFixed(3)}/1k</span>
                  </div>
                </div>

                <div className="mt-3 flex flex-wrap gap-1">
                  {option.capabilities.map(cap => (
                    <span key={cap} className="px-1.5 py-0.5 bg-aviation-cyan/10 rounded text-[10px] text-aviation-cyan">
                      {cap}
                    </span>
                  ))}
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
// Runtime Capability Matrix
// ============================================================================

export const RuntimeCapabilityMatrix: React.FC<RuntimeCapabilityMatrixProps> = ({
  capabilities,
  selectedRuntimeType = null,
  onRuntimeSelect,
  className,
}) => {
  const [hoveredRuntime, setHoveredRuntime] = useState<RuntimeType | null>(null);

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Gauge className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Capability Matrix</h3>
          </div>
        </div>
      </div>

      <div className="flex-1 overflow-auto p-4">
        <div className="mb-4 flex items-center gap-4 text-[10px]">
          <span className="text-aviation-text-dim">Supports:</span>
          <span className="flex items-center gap-1"><div className="w-3 h-3 rounded bg-aviation-cyan/30" /> <span className="text-aviation-text-primary">Streaming</span></span>
          <span className="flex items-center gap-1"><div className="w-3 h-3 rounded bg-purple-500/30" /> <span className="text-aviation-text-primary">Concurrency</span></span>
          <span className="flex items-center gap-1"><div className="w-3 h-3 rounded bg-amber-500/30" /> <span className="text-aviation-text-primary">GPU</span></span>
        </div>

        <svg className="w-full" viewBox="0 0 400 200" preserveAspectRatio="xMidYMid meet">
          <rect x="0" y="0" width="400" height="200" className="fill-aviation-bg-panel" />

          <text x="10" y="20" className="text-[10px] fill-aviation-text-dim">Runtime</text>
          <text x="100" y="20" className="text-[10px] fill-aviation-text-dim">Memory</text>
          <text x="170" y="20" className="text-[10px] fill-aviation-text-dim">Streaming</text>
          <text x="240" y="20" className="text-[10px] fill-aviation-text-dim">Concurrent</text>
          <text x="320" y="20" className="text-[10px] fill-aviation-text-dim">GPU</text>

          {capabilities.map((cap, i) => {
            const y = 35 + i * 22;
            const isSelected = selectedRuntimeType === cap.runtimeType;
            const isHovered = hoveredRuntime === cap.runtimeType;

            return (
              <g
                key={cap.runtimeType}
                onClick={() => onRuntimeSelect?.(cap.runtimeType)}
                onMouseEnter={() => setHoveredRuntime(cap.runtimeType)}
                onMouseLeave={() => setHoveredRuntime(null)}
                className="cursor-pointer"
              >
                <rect x="0" y={y - 12} width="400" height="20" rx="2" className={cn(
                  'transition-colors',
                  isSelected ? 'fill-aviation-cyan/20' : isHovered ? 'fill-aviation-bg-instrument' : 'fill-transparent'
                )} />

                <text x="10" y={y} className={cn('text-[11px] font-medium', isSelected ? 'fill-aviation-cyan' : 'fill-aviation-text-primary')}>
                  {cap.runtimeType.toUpperCase()}
                </text>
                <text x="100" y={y} className="text-[10px] fill-aviation-text-dim">{cap.maxMemory}GB</text>

                <circle cx="190" cy={y - 3} r="4" className={cap.supportsStreaming ? 'fill-aviation-cyan' : 'fill-aviation-bg-instrument stroke-aviation-border-panel'} />
                <circle cx="250" cy={y - 3} r="4" className={cap.supportsConcurrency ? 'fill-purple-500' : 'fill-aviation-bg-instrument stroke-aviation-border-panel'} />
                <circle cx="320" cy={y - 3} r="4" className={cap.supportsGpu ? 'fill-amber-500' : 'fill-aviation-bg-instrument stroke-aviation-border-panel'} />

                <text x="340" y={y} className="text-[9px] fill-aviation-text-dim">${cap.estimatedCostPerHour}/hr</text>
              </g>
            );
          })}
        </svg>

        <div className="mt-4 grid grid-cols-2 gap-3">
          {capabilities.map(cap => {
            const isSelected = selectedRuntimeType === cap.runtimeType;
            return (
              <div
                key={cap.runtimeType}
                onClick={() => onRuntimeSelect?.(cap.runtimeType)}
                className={cn(
                  'p-3 rounded-lg border cursor-pointer transition-all',
                  isSelected ? 'bg-aviation-bg-instrument border-aviation-cyan' : 'bg-aviation-bg-secondary border-aviation-border-panel'
                )}
              >
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm font-medium text-aviation-text-primary uppercase">{cap.runtimeType}</span>
                  <span className="text-xs text-aviation-text-dim">${cap.estimatedCostPerHour}/hr</span>
                </div>
                <div className="text-[10px] text-aviation-text-dim">
                  Max Memory: {cap.maxMemory}GB | Max Compute: {cap.maxCompute}
                </div>
                <div className="flex flex-wrap gap-1 mt-2">
                  {cap.supportedLanguages.slice(0, 3).map(lang => (
                    <span key={lang} className="px-1.5 py-0.5 bg-aviation-bg-instrument rounded text-[10px] text-aviation-text-muted">
                      {lang}
                    </span>
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};
