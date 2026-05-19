/**
 * @functionfly/ui-graph
 * ExecutionNode - visual node on the FunctionCanvas
 */

import * as React from "react";
import { cn } from "./utils";
import { type NodeType, type NodeData } from "./types";
import { NODE_CONFIG } from "./utils";

export interface ExecutionNodeProps extends Omit<NodeData, "position"> {
  position?: { x: number; y: number };
  isSelected?: boolean;
  isHovered?: boolean;
  children?: React.ReactNode;
  className?: string;
  onSelect?: (id: string) => void;
  onDragStart?: (id: string, pos: { x: number; y: number }) => void;
  onDrag?: (id: string, pos: { x: number; y: number }) => void;
  onDragEnd?: (id: string, pos: { x: number; y: number }) => void;
  onAddInput?: (nodeId: string) => void;
  onAddOutput?: (nodeId: string) => void;
  onRemove?: (nodeId: string) => void;
  onPortMouseDown?: (nodeId: string, portId: string, type: "input" | "output", e: React.MouseEvent) => void;
}

export function ExecutionNode({
  id,
  type,
  label,
  description,
  status = "idle",
  isSelected = false,
  isHovered = false,
  onSelect,
  onDragStart,
  onDrag,
  onDragEnd,
  onAddInput,
  onAddOutput,
  onRemove,
  onPortMouseDown,
  children,
  className,
  inputs = [],
  outputs = [],
  executionTime,
  tokenCost,
  errorRate,
  reliabilityScore,
  icon,
  color,
  version,
  position,
}: ExecutionNodeProps) {
  const config = NODE_CONFIG[type] || NODE_CONFIG.function;
  const nodeColor = color || config.color;
  const ref = React.useRef<HTMLDivElement>(null);
  const [dragging, setDragging] = React.useState(false);
  const dragOffset = React.useRef({ x: 0, y: 0 });

  const handleMouseDown = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (ref.current) {
      const rect = ref.current.getBoundingClientRect();
      dragOffset.current = {
        x: e.clientX - rect.left,
        y: e.clientY - rect.top,
      };
    }
    setDragging(true);
    onSelect?.(id);
    onDragStart?.(id, { x: e.clientX, y: e.clientY });
  };

  const handleMouseMove = React.useCallback(
    (e: MouseEvent) => {
      if (!dragging) return;
      onDrag?.(id, { x: e.clientX - dragOffset.current.x, y: e.clientY - dragOffset.current.y });
    },
    [dragging, id, onDrag]
  );

  const handleMouseUp = React.useCallback(() => {
    if (dragging) {
      setDragging(false);
    }
  }, [dragging]);

  React.useEffect(() => {
    if (dragging) {
      document.addEventListener("mousemove", handleMouseMove);
      document.addEventListener("mouseup", handleMouseUp);
      return () => {
        document.removeEventListener("mousemove", handleMouseMove);
        document.removeEventListener("mouseup", handleMouseUp);
      };
    }
  }, [dragging, handleMouseMove, handleMouseUp]);

  const statusColor = {
    idle: "#6b7280",
    running: "#3b82f6",
    completed: "#10b981",
    error: "#ef4444",
    waiting: "#f59e0b",
  }[status] || "#6b7280";

  return (
    <div
      ref={ref}
      className={cn(
        "absolute min-w-[160px] max-w-[240px] rounded-xl border transition-all duration-150",
        "select-none cursor-grab active:cursor-grabbing",
        "bg-[rgba(26,26,37,0.95)] backdrop-blur-sm",
        "border-[rgba(255,255,255,0.12)]",
        isSelected && "border-[rgba(249,115,22,0.5)] shadow-[0_0_20px_rgba(249,115,22,0.15)]",
        isHovered && !isSelected && "border-[rgba(255,255,255,0.2)] hover:border-[rgba(249,115,22,0.3)]",
        dragging && "opacity-80 scale-[1.02]",
        className
      )}
      style={{
        left: 0,
        top: 0,
        transform: `translate(${position?.x ?? 0}px, ${position?.y ?? 0}px)`,
        boxShadow: isSelected
          ? `0 0 24px ${nodeColor}22, 0 4px 20px rgba(0,0,0,0.3)`
          : `0 4px 16px ${nodeColor}11, 0 1px 3px rgba(0,0,0,0.2)`,
      }}
      onClick={(e) => {
        e.stopPropagation();
        onSelect?.(id);
      }}
    >
      {/* Node Header */}
      <div className="flex items-center gap-2 px-3 pt-3 pb-2">
        <span className="text-base leading-none select-none">{icon || config.icon}</span>
        <div className="flex-1 min-w-0">
          <div className="text-xs font-bold text-text-primary truncate" style={{ color: nodeColor }}>
            {formatLabel(label)}
          </div>
          {description && (
            <div className="text-[10px] text-text-muted truncate mt-0.5">{description}</div>
          )}
        </div>
        {/* Status dot */}
        <div className="relative shrink-0">
          <div
            className="size-2 rounded-full"
            style={{
              backgroundColor: statusColor,
              boxShadow: status === "running" ? `0 0 6px ${statusColor}` : "none",
            }}
          />
        </div>
        {/* Remove button */}
        {onRemove && (
          <button
            className="opacity-0 group-hover/node:opacity-60 hover:opacity-100 hover:text-error transition-opacity text-xs size-5 flex items-center justify-center rounded hover:bg-error/10"
            onClick={(e) => {
              e.stopPropagation();
              onRemove(id);
            }}
          >
            ✕
          </button>
        )}
      </div>

      {/* Node Body */}
      {(executionTime != null || tokenCost != null || errorRate != null || reliabilityScore != null) && (
        <div className="px-3 py-1.5 space-y-1 border-t border-[rgba(255,255,255,0.06)]">
          {executionTime != null && (
            <div className="flex justify-between text-[10px] text-text-muted">
              <span>Latency</span>
              <span className="text-text-primary font-mono">{formatDuration(executionTime)}</span>
            </div>
          )}
          {tokenCost != null && (
            <div className="flex justify-between text-[10px] text-text-muted">
              <span>Cost</span>
              <span className="text-text-primary font-mono">${tokenCost.toFixed(4)}</span>
            </div>
          )}
          {reliabilityScore != null && (
            <div className="flex items-center gap-2 text-[10px]">
              <span className="text-text-muted">Reliability</span>
              <div className="flex-1 h-1.5 bg-bg-tertiary rounded-full overflow-hidden">
                <div
                  className="h-full rounded-full transition-all duration-300"
                  style={{
                    width: `${reliabilityScore}%`,
                    backgroundColor: reliabilityScore > 80 ? "#10b981" : reliabilityScore > 50 ? "#f59e0b" : "#ef4444",
                  }}
                />
              </div>
              <span className="text-text-primary font-mono w-8 text-right">{reliabilityScore}%</span>
            </div>
          )}
          {errorRate != null && (
            <div className="flex justify-between text-[10px] text-text-muted">
              <span>Error rate</span>
              <span className="text-text-primary font-mono">{errorRate.toFixed(1)}%</span>
            </div>
          )}
        </div>
      )}

      {/* Input Ports */}
      <div className="flex justify-start mt-1 px-1">
        {inputs.map((port) => (
          <div
            key={port.id}
            className={cn(
              "port port-input size-3 rounded-full -ml-0.5 border-2 cursor-crosshair transition-all duration-150",
              "bg-bg-primary border-border-subtle hover:border-brand-500 hover:scale-150"
            )}
            onMouseDown={(e) => onPortMouseDown?.(id, port.id, "input", e)}
            title={`${port.label} (${port.dataType || "any"})`}
          />
        ))}
        {inputs.length === 0 && (
          <div
            className="size-3 rounded-full border-2 border-dashed border-border-subtle/30 cursor-crosshair hover:border-brand-500/50 transition-colors"
            onClick={() => onAddInput?.(id)}
            title="Add input port"
          />
        )}
      </div>

      {/* Output Ports */}
      <div className="flex justify-end mt-1 px-1 mb-2">
        {outputs.map((port) => (
          <div
            key={port.id}
            className={cn(
              "port port-output size-3 rounded-full -mr-0.5 border-2 cursor-crosshair transition-all duration-150",
              "bg-bg-primary border-border-subtle hover:border-brand-500 hover:scale-150"
            )}
            onMouseDown={(e) => onPortMouseDown?.(id, port.id, "output", e)}
            title={`${port.label} (${port.dataType || "any"})`}
          />
        ))}
        {outputs.length === 0 && (
          <div
            className="size-3 rounded-full border-2 border-dashed border-border-subtle/30 cursor-crosshair hover:border-brand-500/50 transition-colors"
            onClick={() => onAddOutput?.(id)}
            title="Add output port"
          />
        )}
      </div>

      {/* Version badge */}
      {version && (
        <div className="absolute -top-1 -right-1 px-1 text-[8px] bg-brand-500 text-white rounded font-mono">
          v{version}
        </div>
      )}
    </div>
  );
}

function formatLabel(label: string): string {
  return label.length > 20 ? label.slice(0, 17) + "…" : label;
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  return `${(ms / 60000).toFixed(1)}m`;
}

// Global styles for ports and drag state
const PORT_STYLES = `
.port {
  position: absolute;
  z-index: 10;
}
.port-input {
  /* Positioned by parent layout */
}
.port-output {
  /* Positioned by parent layout */
}
.port:hover {
  transform: scale(1.5);
  z-index: 20;
}
`;