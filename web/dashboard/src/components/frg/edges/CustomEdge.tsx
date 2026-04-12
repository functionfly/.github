/**
 * CustomEdge Component (EdgeConnector)
 * Handles connections with type validation and animated data flow
 */

import { memo, useCallback, useState } from 'react';
import { EdgeLabelRenderer, BaseEdge, getSmoothStepPath, type EdgeProps } from '@xyflow/react';
import { CheckCircle, AlertCircle, Activity, Unlink, Zap } from 'lucide-react';

import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';

import { useFRGStore } from '@/stores/frgStore';
import type { FlowEdgeData } from '@/types/frg';

interface CustomEdgeProps extends EdgeProps {
  data?: FlowEdgeData;
}

// Default edge data to prevent undefined errors
const defaultEdgeData: FlowEdgeData = {
  mapping: { sourcePath: '*', targetPath: '*' },
  isValid: true,
  runtimeState: {
    status: 'idle',
    recordsTransferred: 0,
    bytesTransferred: 0,
    isDataFlowing: false,
    flowProgress: 0,
  },
};

export const CustomEdge = memo(function CustomEdge({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  markerEnd,
  style = {},
  selected,
  data = defaultEdgeData,
}: CustomEdgeProps) {
  const store = useFRGStore();
  const { removeEdge, setSelectedEdge } = store;
  
  // Safely destructure data with fallback to default
  const { 
    mapping, 
    condition, 
    retryPolicy, 
    runtimeState, 
    isValid = true, 
    validationError 
  } = data || defaultEdgeData;
  
  const [isHovered, setIsHovered] = useState(false);

  // Calculate path
  const [edgePath, labelX, labelY] = getSmoothStepPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
    borderRadius: 16,
    offset: 20,
  });

  const handleDelete = useCallback((e: React.MouseEvent) => {
    e.stopPropagation();
    if (id) {
      removeEdge(id);
    }
  }, [id, removeEdge]);

  const handleConfigure = useCallback(() => {
    setSelectedEdge(id);
  }, [id, setSelectedEdge]);

  // Determine edge styling based on state
  const isDataFlowing = runtimeState?.isDataFlowing ?? false;
  const hasError = runtimeState?.status === 'error' || !isValid;
  const isActive = runtimeState?.status === 'active' || isDataFlowing;

  const strokeColor = hasError 
    ? '#ef4444' 
    : isActive 
    ? '#8b5cf6' 
    : selected 
    ? '#6366f1' 
    : 'var(--border-default)';

  const strokeWidth = selected ? 3 : hasError ? 2.5 : 2;

  return (
    <TooltipProvider>
      <g
        className="group"
        onMouseEnter={() => setIsHovered(true)}
        onMouseLeave={() => setIsHovered(false)}
      >
        {/* Invisible wider path for easier selection */}
        <BaseEdge
          path={edgePath}
          style={{
            stroke: 'transparent',
            strokeWidth: 20,
            fill: 'none',
          }}
          className="cursor-pointer"
        />

        {/* Main edge path */}
        <BaseEdge
          path={edgePath}
          markerEnd={markerEnd}
          style={{
            ...style,
            stroke: strokeColor,
            strokeWidth,
            transition: 'all 0.2s ease',
          }}
          className={cn(
            "transition-all duration-200",
            isDataFlowing && "animate-pulse"
          )}
          interactionWidth={20}
        />

        {/* Data Flow Animation */}
        {isDataFlowing && (
          <>
            <path
              d={edgePath}
              fill="none"
              stroke="rgba(139, 92, 246, 0.5)"
              strokeWidth={4}
              strokeDasharray="10 10"
              className="animate-[dash_1s_linear_infinite]"
              style={{
                animation: 'dash 1s linear infinite',
              }}
            />
            <style>{`
              @keyframes dash {
                to {
                  stroke-dashoffset: -20;
                }
              }
            `}</style>
          </>
        )}

        {/* Edge Label */}
        <EdgeLabelRenderer>
          <div
            style={{
              position: 'absolute',
              transform: `translate(-50%, -50%) translate(${labelX}px, ${labelY}px)`,
              pointerEvents: 'all',
            }}
            className={cn(
              "flex items-center gap-1 transition-opacity duration-200",
              (isHovered || selected) ? "opacity-100" : "opacity-0"
            )}
          >
            {/* Status Badge */}
            <Badge 
              variant={hasError ? "destructive" : isActive ? "default" : "secondary"}
              className="text-[10px] px-1.5 py-0 cursor-pointer"
              onClick={handleConfigure}
            >
              {hasError ? (
                <AlertCircle className="w-3 h-3 mr-1" />
              ) : isActive ? (
                <Activity className="w-3 h-3 mr-1 animate-pulse" />
              ) : (
                <CheckCircle className="w-3 h-3 mr-1" />
              )}
              {hasError ? 'Error' : isActive ? 'Active' : 'Ready'}
            </Badge>

            {/* Delete Button */}
            <Button
              variant="ghost"
              size="icon"
              className="h-5 w-5 bg-red-500/10 hover:bg-red-500/20"
              onClick={handleDelete}
            >
              <Unlink className="w-3 h-3 text-red-500" />
            </Button>
          </div>

          {/* Data Transfer Info */}
          {runtimeState && (runtimeState.recordsTransferred > 0 || runtimeState.bytesTransferred > 0) && (
            <Tooltip>
              <TooltipTrigger asChild>
                <div
                  style={{
                    position: 'absolute',
                    transform: `translate(-50%, -50%) translate(${labelX}px, ${labelY + 20}px)`,
                  }}
                  className="text-[10px] text-[var(--text-secondary)] bg-[var(--bg-secondary)] px-2 py-1 rounded border border-[var(--border-subtle)]"
                >
                  <Zap className="w-3 h-3 inline mr-1" />
                  {runtimeState.recordsTransferred.toLocaleString()} records
                  {runtimeState.bytesTransferred > 0 && (
                    <span className="ml-1">
                      ({(runtimeState.bytesTransferred / 1024 / 1024).toFixed(2)} MB)
                    </span>
                  )}
                </div>
              </TooltipTrigger>
              <TooltipContent>
                <div className="space-y-1 text-xs">
                  <p>Records: {runtimeState.recordsTransferred.toLocaleString()}</p>
                  <p>Bytes: {(runtimeState.bytesTransferred / 1024).toFixed(2)} KB</p>
                </div>
              </TooltipContent>
            </Tooltip>
          )}

          {/* Validation Error */}
          {hasError && validationError && (
            <div
              style={{
                position: 'absolute',
                transform: `translate(-50%, -50%) translate(${labelX}px, ${labelY - 25}px)`,
              }}
              className="text-[10px] text-red-500 bg-red-500/10 px-2 py-1 rounded border border-red-500/20 max-w-[150px] truncate"
            >
              {validationError}
            </div>
          )}

          {/* Transform Badge */}
          {mapping?.transform && (
            <div
              style={{
                position: 'absolute',
                transform: `translate(-50%, -50%) translate(${labelX}px, ${labelY - 20}px)`,
              }}
              className="text-[10px] text-purple-400 bg-purple-500/10 px-2 py-0.5 rounded border border-purple-500/20"
            >
              {mapping.transform}
            </div>
          )}

          {/* Condition Badge */}
          {condition && (
            <div
              style={{
                position: 'absolute',
                transform: `translate(-50%, -50%) translate(${labelX}px, ${labelY - 35}px)`,
              }}
              className="text-[10px] text-blue-400 bg-blue-500/10 px-2 py-0.5 rounded border border-blue-500/20"
            >
              if {condition.field} {condition.operator}
            </div>
          )}
        </EdgeLabelRenderer>
      </g>
    </TooltipProvider>
  );
});

export default CustomEdge;
