/**
 * FunctionNode Component
 * Visual block representing a function in the graph
 * Shows: name, inputs (left), outputs (right), status
 */

import { memo, useCallback, useState } from 'react';
import { Handle, Position, type NodeProps } from '@xyflow/react';
import { 
  Play, 
  Square, 
  AlertCircle, 
  Loader2, 
  CheckCircle, 
  Clock, 
  MoreHorizontal,
  Settings,
  Bug,
  Activity,
  Zap,
} from 'lucide-react';

import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { 
  Tooltip, 
  TooltipContent, 
  TooltipProvider, 
  TooltipTrigger 
} from '@/components/ui/tooltip';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

import { useFRGStore } from '@/stores/frgStore';
import type { FunctionNodeData, NodeExecutionStatus } from '@/types/frg';

interface FunctionNodeProps extends NodeProps {
  data: FunctionNodeData;
}

// Status configuration
const statusConfig: Record<NodeExecutionStatus, {
  icon: React.ReactNode;
  color: string;
  bgColor: string;
  borderColor: string;
  label: string;
}> = {
  idle: {
    icon: <Zap className="w-3 h-3" />,
    color: 'text-gray-400',
    bgColor: 'bg-gray-500/10',
    borderColor: 'border-gray-500/30',
    label: 'Idle',
  },
  pending: {
    icon: <Clock className="w-3 h-3" />,
    color: 'text-yellow-400',
    bgColor: 'bg-yellow-500/10',
    borderColor: 'border-yellow-500/30',
    label: 'Pending',
  },
  executing: {
    icon: <Loader2 className="w-3 h-3 animate-spin" />,
    color: 'text-blue-400',
    bgColor: 'bg-blue-500/10',
    borderColor: 'border-blue-500/30',
    label: 'Running',
  },
  waiting: {
    icon: <Clock className="w-3 h-3" />,
    color: 'text-orange-400',
    bgColor: 'bg-orange-500/10',
    borderColor: 'border-orange-500/30',
    label: 'Waiting',
  },
  completed: {
    icon: <CheckCircle className="w-3 h-3" />,
    color: 'text-green-400',
    bgColor: 'bg-green-500/10',
    borderColor: 'border-green-500/30',
    label: 'Completed',
  },
  failed: {
    icon: <AlertCircle className="w-3 h-3" />,
    color: 'text-red-400',
    bgColor: 'bg-red-500/10',
    borderColor: 'border-red-500/30',
    label: 'Failed',
  },
  retrying: {
    icon: <Activity className="w-3 h-3" />,
    color: 'text-purple-400',
    bgColor: 'bg-purple-500/10',
    borderColor: 'border-purple-500/30',
    label: 'Retrying',
  },
};

// Generate input/output handles from function schema
function generateHandles(type: 'input' | 'output', count: number, nodeId: string) {
  const handles = [];
  const spacing = 100 / (count + 1);
  
  for (let i = 0; i < count; i++) {
    const position = ((i + 1) * spacing);
    handles.push(
      <Handle
        key={`${type}-${i}`}
        type={type === 'input' ? 'target' : 'source'}
        position={type === 'input' ? Position.Left : Position.Right}
        id={`${type}-${nodeId}-${i}`}
        style={{
          top: `${position}%`,
          width: '12px',
          height: '12px',
          background: 'var(--bg-secondary)',
          border: '2px solid var(--border-default)',
        }}
        className="!border-brand-500 hover:!bg-brand-500 hover:!border-brand-500 transition-colors"
      />
    );
  }
  return handles;
}

export const FunctionNode = memo(function FunctionNode({ 
  id, 
  data, 
  selected, 
  isConnectable,
}: FunctionNodeProps) {
  const store = useFRGStore();
  const { removeNode, setSelectedNode, startExecution } = store;
  
  const { functionRef, runtimeState, isEditable } = data;
  const status = runtimeState?.status || 'idle';
  const statusStyle = statusConfig[status];
  
  const [isHovered, setIsHovered] = useState(false);

  const handleConfigChange = useCallback(() => {
    setSelectedNode(id);
  }, [id, setSelectedNode]);

  const handleRun = useCallback(() => {
    // Run this node - in a real implementation this would trigger execution
    console.log('Run node', id);
  }, [id]);

  const handleTest = useCallback(() => {
    // Test this node
    console.log('Test node', id);
  }, [id]);

  const handleDebug = useCallback(() => {
    // Debug this node
    console.log('Debug node', id);
  }, [id]);

  const handleDelete = useCallback(() => {
    removeNode(id);
  }, [id, removeNode]);

  // Mock input/output counts - in real app, derive from function schema
  const inputCount = 2;
  const outputCount = 2;

  return (
    <TooltipProvider>
      <div
        className={cn(
          "relative min-w-[180px] max-w-[280px] rounded-xl border-2 transition-all duration-200",
          "bg-[var(--bg-secondary)] shadow-lg",
          selected 
            ? "border-brand-500 ring-2 ring-brand-500/20 shadow-brand-500/20" 
            : "border-[var(--border-subtle)]",
          isHovered && "shadow-xl scale-[1.02]",
          status !== 'idle' && statusStyle.borderColor
        )}
        onMouseEnter={() => setIsHovered(true)}
        onMouseLeave={() => setIsHovered(false)}
      >
        {/* Status Indicator Strip */}
        <div className={cn(
          "absolute top-0 left-0 right-0 h-1 rounded-t-xl",
          status !== 'idle' && statusStyle.bgColor.replace('/10', '/50')
        )} />

        {/* Input Handles */}
        <div className="absolute left-0 top-0 bottom-0 w-0">
          {generateHandles('input', inputCount, id)}
        </div>

        {/* Output Handles */}
        <div className="absolute right-0 top-0 bottom-0 w-0">
          {generateHandles('output', outputCount, id)}
        </div>

        {/* Header */}
        <div className={cn(
          "px-3 py-2 border-b border-[var(--border-subtle)] rounded-t-xl",
          "flex items-center gap-2"
        )}>
          {/* Status Icon */}
          <div className={cn(
            "flex items-center justify-center w-6 h-6 rounded-full",
            statusStyle.bgColor,
            statusStyle.color
          )}>
            {statusStyle.icon}
          </div>

          {/* Function Name */}
          <div className="flex-1 min-w-0">
            <div className="font-semibold text-sm text-[var(--text-primary)] truncate">
              {functionRef.name}
            </div>
            <div className="text-[10px] text-[var(--text-secondary)] truncate">
              {functionRef.author}/{functionRef.version}
            </div>
          </div>

          {/* Quick Actions */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button 
                variant="ghost" 
                size="icon" 
                className="h-6 w-6"
                style={{ opacity: isHovered ? 1 : 0.6 }}
              >
                <MoreHorizontal className="w-3 h-3" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-40">
              <DropdownMenuItem onClick={handleConfigChange}>
                <Settings className="w-4 h-4 mr-2" />
                Configure
              </DropdownMenuItem>
              <DropdownMenuItem onClick={handleRun}>
                <Play className="w-4 h-4 mr-2" />
                Run Node
              </DropdownMenuItem>
              <DropdownMenuItem onClick={handleTest}>
                <Zap className="w-4 h-4 mr-2" />
                Test
              </DropdownMenuItem>
              <DropdownMenuItem onClick={handleDebug}>
                <Bug className="w-4 h-4 mr-2" />
                Debug
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem 
                onClick={handleDelete}
                className="text-red-500 focus:text-red-500"
              >
                <Square className="w-4 h-4 mr-2" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        {/* Body */}
        <div className="p-3 space-y-2">
          {/* Execution Info */}
          {runtimeState && (
            <div className="flex items-center gap-2 text-xs">
              <Badge 
                variant="secondary" 
                className={cn(
                  "text-[10px] px-1.5 py-0.5",
                  statusStyle.bgColor,
                  statusStyle.color
                )}
              >
                {statusStyle.label}
              </Badge>
              
              {runtimeState.durationMs > 0 && (
                <span className="text-[var(--text-secondary)]">
                  {runtimeState.durationMs}ms
                </span>
              )}

              {runtimeState.attemptCount > 0 && (
                <Tooltip>
                  <TooltipTrigger>
                    <Badge variant="outline" className="text-[10px]">
                      {runtimeState.attemptCount}x
                    </Badge>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>Attempt count</p>
                  </TooltipContent>
                </Tooltip>
              )}
            </div>
          )}

          {/* Progress Bar (when running) */}
          {status === 'executing' && runtimeState?.progress !== undefined && (
            <div className="space-y-1">
              <div className="h-1.5 w-full bg-[var(--border-subtle)] rounded-full overflow-hidden">
                <div 
                  className={cn(
                    "h-full rounded-full transition-all duration-300",
                    "bg-gradient-to-r from-brand-500 to-purple-500"
                  )}
                  style={{ width: `${runtimeState.progress}%` }}
                />
              </div>
              <div className="text-[10px] text-[var(--text-secondary)] text-right">
                {runtimeState.progress}%
              </div>
            </div>
          )}

          {/* Error Display */}
          {runtimeState?.error && (
            <div className="text-xs text-red-500 bg-red-500/10 p-2 rounded border border-red-500/20 truncate">
              {runtimeState.error}
            </div>
          )}

          {/* Input/Output Preview */}
          <div className="grid grid-cols-2 gap-2 text-[10px] text-[var(--text-secondary)]">
            <div className="space-y-1">
              <span className="font-medium">Inputs</span>
              <div className="text-[var(--text-muted)]">
                {inputCount} params
              </div>
            </div>
            <div className="space-y-1 text-right">
              <span className="font-medium">Outputs</span>
              <div className="text-[var(--text-muted)]">
                {outputCount} results
              </div>
            </div>
          </div>
        </div>

        {/* Footer - mini toolbar */}
        {isHovered && (
          <div className="absolute -bottom-10 left-1/2 -translate-x-1/2 flex items-center gap-1 p-1 bg-[var(--bg-secondary)] rounded-lg border border-[var(--border-subtle)] shadow-xl z-10">
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="ghost" size="icon" className="h-6 w-6" onClick={handleRun}>
                  <Play className="w-3 h-3" />
                </Button>
              </TooltipTrigger>
              <TooltipContent side="top">Run</TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="ghost" size="icon" className="h-6 w-6" onClick={handleTest}>
                  <Zap className="w-3 h-3" />
                </Button>
              </TooltipTrigger>
              <TooltipContent side="top">Quick Test</TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="ghost" size="icon" className="h-6 w-6" onClick={handleDebug}>
                  <Bug className="w-3 h-3" />
                </Button>
              </TooltipTrigger>
              <TooltipContent side="top">Debug</TooltipContent>
            </Tooltip>
          </div>
        )}
      </div>
    </TooltipProvider>
  );
});

export default FunctionNode;
