/**
 * ExecutionOverlay Component
 * Shows: Data flowing between nodes, Highlight active nodes, Real-time updates
 * This makes the product feel alive
 */

import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useStoreApi, type Node } from '@xyflow/react';
import { motion, AnimatePresence } from 'framer-motion';
import { 
  Activity, 
  Zap, 
  ArrowRight, 
  Play, 
  CheckCircle, 
  AlertCircle,
  Database,
  Network,
  Gauge,
  TrendingUp,
} from 'lucide-react';

import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';

import { useFRGStore } from '@/stores/frgStore';
import type { InstanceStatus } from '@/types/frg';

// Data packet animation component
interface DataPacketProps {
  sourceX: number;
  sourceY: number;
  targetX: number;
  targetY: number;
  data?: unknown;
  duration?: number;
  onComplete?: () => void;
}

function DataPacket({ 
  sourceX, 
  sourceY, 
  targetX, 
  targetY, 
  data,
  duration = 1000,
  onComplete 
}: DataPacketProps) {
  const [progress, setProgress] = useState(0);
  
  useEffect(() => {
    const startTime = performance.now();
    let animationFrame: number;
    
    const animate = (currentTime: number) => {
      const elapsed = currentTime - startTime;
      const newProgress = Math.min(elapsed / duration, 1);
      setProgress(newProgress);
      
      if (newProgress < 1) {
        animationFrame = requestAnimationFrame(animate);
      } else {
        onComplete?.();
      }
    };
    
    animationFrame = requestAnimationFrame(animate);
    return () => cancelAnimationFrame(animationFrame);
  }, [duration, onComplete]);

  const x = sourceX + (targetX - sourceX) * progress;
  const y = sourceY + (targetY - sourceY) * progress;

  return (
    <motion.div
      className="absolute pointer-events-none z-50"
      style={{ 
        left: x, 
        top: y,
        transform: 'translate(-50%, -50%)'
      }}
      initial={{ scale: 0, opacity: 0 }}
      animate={{ scale: 1, opacity: 1 }}
      exit={{ scale: 0, opacity: 0 }}
    >
      <div className="relative">
        <div className="w-6 h-6 rounded-full bg-gradient-to-r from-brand-500 to-purple-500 shadow-lg shadow-brand-500/50 flex items-center justify-center">
          <Database className="w-3 h-3 text-white" />
        </div>
        {/* Trail effect */}
        <div 
          className="absolute inset-0 rounded-full bg-brand-500/30 animate-ping"
          style={{ animationDuration: '0.5s' }}
        />
      </div>
    </motion.div>
  );
}

// Active node pulse effect
function NodePulse({ 
  nodeId, 
  x, 
  y, 
  width, 
  height, 
  status 
}: { 
  nodeId: string;
  x: number;
  y: number;
  width: number;
  height: number;
  status: 'executing' | 'completed' | 'failed';
}) {
  const color = status === 'completed' ? 'border-green-500' : 
                status === 'failed' ? 'border-red-500' : 
                'border-brand-500';
  
  return (
    <div
      className={cn(
        "absolute pointer-events-none rounded-xl border-2",
        color,
        "animate-[pulse-ring_1.5s_ease-out_infinite]"
      )}
      style={{
        left: x - 4,
        top: y - 4,
        width: width + 8,
        height: height + 8,
      }}
    />
  );
}

// Mini metrics card
function LiveMetrics() {
  const { nodeRuntimeStates, executionStatus, executionProgress } = useFRGStore();

  const metrics = useMemo(() => {
    const states = Object.values(nodeRuntimeStates);
    const totalNodes = states.length;
    const executingNodes = states.filter(s => s.status === 'executing' || s.status === 'retrying').length;
    const completedNodes = states.filter(s => s.status === 'completed').length;
    const failedNodes = states.filter(s => s.status === 'failed').length;

    const totalDurationMs = states.reduce((sum, s) => sum + (s.durationMs || 0), 0);
    const avgLatency = totalNodes > 0 ? Math.round(totalDurationMs / totalNodes) : 0;

    const activeCount = executingNodes + completedNodes + failedNodes;
    const throughput = activeCount > 0 ? Math.round((completedNodes / Math.max(activeCount, 1)) * 1000) : 0;
    const cpu = executionStatus === 'running' ? Math.min(100, Math.round((executingNodes / Math.max(totalNodes, 1)) * 100)) : 0;
    const memory = Math.round(totalDurationMs / 10000);

    return {
      throughput,
      latency: avgLatency,
      cpu,
      memory: Math.max(memory, 10),
      progress: executionProgress,
      executing: executingNodes,
      completed: completedNodes,
      failed: failedNodes,
    };
  }, [nodeRuntimeStates, executionStatus, executionProgress]);

  return (
    <div className="bg-[var(--bg-secondary)] border border-[var(--border-subtle)] rounded-lg p-3 shadow-lg">
      <div className="flex items-center gap-2 mb-2">
        <Activity className="w-4 h-4 text-green-500 animate-pulse" />
        <span className="text-xs font-medium text-[var(--text-primary)]">Live Metrics</span>
      </div>
      <div className="grid grid-cols-2 gap-2 text-xs">
        <div>
          <span className="text-[var(--text-muted)]">Progress</span>
          <div className="text-[var(--text-primary)] font-mono">
            {metrics.progress}%
          </div>
        </div>
        <div>
          <span className="text-[var(--text-muted)]">Latency</span>
          <div className="text-[var(--text-primary)] font-mono">
            {metrics.latency}ms
          </div>
        </div>
        <div>
          <span className="text-[var(--text-muted)]">CPU</span>
          <div className="flex items-center gap-1">
            <div className="flex-1 h-1 bg-[var(--bg-tertiary)] rounded-full overflow-hidden">
              <div
                className="h-full bg-blue-500 rounded-full transition-all"
                style={{ width: `${metrics.cpu}%` }}
              />
            </div>
            <span className="text-[var(--text-primary)] font-mono">{metrics.cpu}%</span>
          </div>
        </div>
        <div>
          <span className="text-[var(--text-muted)]">Memory</span>
          <div className="text-[var(--text-primary)] font-mono">
            {metrics.memory}MB
          </div>
        </div>
        <div className="col-span-2 text-xs text-[var(--text-muted)] mt-1">
          Executing: {metrics.executing} | Completed: {metrics.completed} | Failed: {metrics.failed}
        </div>
      </div>
    </div>
  );
}

// Data flow visualization along edge
function EdgeFlow({ 
  sourceX, 
  sourceY, 
  targetX, 
  targetY,
  isActive,
}: { 
  sourceX: number;
  sourceY: number;
  targetX: number;
  targetY: number;
  isActive: boolean;
}) {
  if (!isActive) return null;

  return (
    <svg
      className="absolute pointer-events-none"
      style={{
        left: Math.min(sourceX, targetX),
        top: Math.min(sourceY, targetY),
        width: Math.abs(targetX - sourceX) + 50,
        height: Math.abs(targetY - sourceY) + 50,
      }}
    >
      <defs>
        <marker
          id="flow-arrow"
          viewBox="0 0 10 10"
          refX="5"
          refY="5"
          markerWidth="4"
          markerHeight="4"
          orient="auto-start-reverse"
        >
          <path d="M 0 0 L 10 5 L 0 10 z" fill="#8b5cf6" />
        </marker>
        <linearGradient id="flow-gradient" x1="0%" y1="0%" x2="100%" y2="0%">
          <stop offset="0%" stopColor="transparent" />
          <stop offset="50%" stopColor="#8b5cf6" />
          <stop offset="100%" stopColor="transparent" />
        </linearGradient>
      </defs>
      <line
        x1={sourceX - Math.min(sourceX, targetX) + 25}
        y1={sourceY - Math.min(sourceY, targetY) + 25}
        x2={targetX - Math.min(sourceX, targetX) + 25}
        y2={targetY - Math.min(sourceY, targetY) + 25}
        stroke="url(#flow-gradient)"
        strokeWidth="3"
        strokeDasharray="10 5"
        className="animate-[dash_0.5s_linear_infinite]"
        markerEnd="url(#flow-arrow)"
      />
    </svg>
  );
}

export function ExecutionOverlay() {
  const store = useFRGStore();
  const reactFlowStore = useStoreApi();
  const { 
    executionStatus,
    nodes,
    edges,
    nodeRuntimeStates,
    dataFlowParticles,
    addDataFlowParticle,
    removeDataFlowParticle,
  } = store;

  const [showParticles, setShowParticles] = useState(true);
  const [showMetrics, setShowMetrics] = useState(true);

  // Get active nodes
  const activeNodes = useMemo(() => {
    return nodes.filter(n => {
      const state = nodeRuntimeStates[n.id];
      return state?.status === 'executing' || state?.status === 'retrying';
    });
  }, [nodes, nodeRuntimeStates]);

  // Is execution active - properly typed
  const isExecutionActive = executionStatus === 'running' || executionStatus === 'streaming';

  // Get edge positions for flow visualization - use actual node positions from React Flow
  const edgeFlows = useMemo(() => {
    if (!isExecutionActive) return [];
    
    // Get current node positions from React Flow store
    const nodePositions = (reactFlowStore.getState() as any).nodeLookup;
    
    return edges
      .filter(e => e.data?.runtimeState?.isDataFlowing)
      .map(e => {
        const sourceNode = nodePositions.get(e.source) as Node | undefined;
        const targetNode = nodePositions.get(e.target) as Node | undefined;
        return {
          id: e.id,
          sourceX: sourceNode?.position?.x || 0,
          sourceY: sourceNode?.position?.y || 0,
          targetX: targetNode?.position?.x || 0,
          targetY: targetNode?.position?.y || 0,
        };
      });
  }, [edges, reactFlowStore, isExecutionActive]);

  // Get particle positions based on actual edge connections
  const particlePositions = useMemo(() => {
    const nodePositions = (reactFlowStore.getState() as any).nodeLookup;
    return dataFlowParticles.map(particle => {
      const edge = edges.find(e => e.id === particle.edgeId);
      if (!edge) return null;
      
      const sourceNode = nodePositions.get(edge.source) as Node | undefined;
      const targetNode = nodePositions.get(edge.target) as Node | undefined;
      
      if (!sourceNode || !targetNode) return null;
      
      return {
        ...particle,
        sourceX: sourceNode.position.x + (sourceNode.width || 180) / 2,
        sourceY: sourceNode.position.y + (sourceNode.height || 100) / 2,
        targetX: targetNode.position.x + (targetNode.width || 180) / 2,
        targetY: targetNode.position.y + (targetNode.height || 100) / 2,
      };
    }).filter(Boolean);
  }, [dataFlowParticles, edges, reactFlowStore]);

  if (!isExecutionActive && dataFlowParticles.length === 0 && activeNodes.length === 0) {
    return null;
  }

  return (
    <div className="absolute inset-0 pointer-events-none z-40 overflow-hidden">
      {/* Active Node Pulses */}
      <AnimatePresence>
        {activeNodes.map((node) => (
          <NodePulse
            key={node.id}
            nodeId={node.id}
            x={node.position.x}
            y={node.position.y}
            width={node.width || 180}
            height={node.height || 100}
            status={nodeRuntimeStates[node.id]?.status === 'retrying' ? 'failed' : 'executing'}
          />
        ))}
      </AnimatePresence>

      {/* Edge Flows */}
      {edgeFlows.map((flow) => (
        <EdgeFlow
          key={flow.id}
          sourceX={flow.sourceX}
          sourceY={flow.sourceY}
          targetX={flow.targetX}
          targetY={flow.targetY}
          isActive={true}
        />
      ))}

      {/* Data Particles - use actual calculated positions */}
      <AnimatePresence>
        {showParticles && particlePositions.map((particle) => particle && (
          <DataPacket
            key={particle.id}
            sourceX={particle.sourceX}
            sourceY={particle.sourceY}
            targetX={particle.targetX}
            targetY={particle.targetY}
            data={particle.data}
            onComplete={() => removeDataFlowParticle(particle.id)}
          />
        ))}
      </AnimatePresence>

      {/* Status Badge */}
      <div className="absolute top-4 left-1/2 -translate-x-1/2 pointer-events-auto">
        <Badge 
          variant={executionStatus === 'failed' ? 'destructive' : 'default'}
          className={cn(
            "px-3 py-1.5 text-sm shadow-lg",
            executionStatus === 'running' && "bg-brand-500 animate-pulse"
          )}
        >
          <div className="flex items-center gap-2">
            {executionStatus === 'running' && <Activity className="w-4 h-4 animate-spin" />}
            {executionStatus === 'streaming' && <Zap className="w-4 h-4 animate-pulse" />}
            {executionStatus === 'completed' && <CheckCircle className="w-4 h-4" />}
            {executionStatus === 'failed' && <AlertCircle className="w-4 h-4" />}
            <span className="uppercase">{executionStatus}</span>
          </div>
        </Badge>
      </div>

      {/* Live Metrics Panel */}
      {showMetrics && (
        <div className="absolute bottom-4 right-4 pointer-events-auto">
          <LiveMetrics />
        </div>
      )}

      {/* Toggle Controls */}
      <div className="absolute top-4 right-4 flex flex-col gap-2 pointer-events-auto">
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="secondary"
                size="icon"
                className="w-8 h-8 shadow-lg"
                onClick={() => setShowParticles(!showParticles)}
              >
                <Zap className={cn("w-4 h-4", showParticles && "text-brand-500")} />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Toggle data flow visualization</TooltipContent>
          </Tooltip>
        </TooltipProvider>
        
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="secondary"
                size="icon"
                className="w-8 h-8 shadow-lg"
                onClick={() => setShowMetrics(!showMetrics)}
              >
                <Gauge className={cn("w-4 h-4", showMetrics && "text-brand-500")} />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Toggle live metrics</TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </div>

      <style>{`
        @keyframes pulse-ring {
          0% {
            transform: scale(1);
            opacity: 1;
          }
          100% {
            transform: scale(1.1);
            opacity: 0;
          }
        }
        
        @keyframes dash {
          to {
            stroke-dashoffset: -15;
          }
        }
      `}</style>
    </div>
  );
}

export default ExecutionOverlay;
