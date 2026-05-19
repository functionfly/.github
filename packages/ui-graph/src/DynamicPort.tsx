/**
 * @functionfly/ui-graph
 * DynamicPort - interactive port component for connections
 */

import * as React from "react";
import { cn } from "./utils";
import type { DynamicPortProps, PortData } from "./types";

const POSITION_OFFSETS = {
  left: { x: -6, y: 0 },
  right: { x: 6, y: 0 },
  top: { x: 0, y: -6 },
  bottom: { x: 0, y: 6 },
};

const DATA_TYPE_COLORS: Record<string, string> = {
  string: "#10b981",
  number: "#3b82f6",
  boolean: "#f59e0b",
  object: "#8b5cf6",
  array: "#ec4899",
  json: "#06b6d4",
  text: "#a855f7",
  any: "#6b7280",
};

export function DynamicPort({
  nodeId,
  port,
  position = "right",
  isConnected = false,
  isHovered = false,
  isValidConnection = true,
  onMouseDown,
  onMouseUp,
  onConnectionStart,
  onConnectionEnd,
  className,
}: DynamicPortProps) {
  const offset = POSITION_OFFSETS[position];
  const dataTypeColor = DATA_TYPE_COLORS[port.dataType || "any"];

  const handleMouseDown = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    onConnectionStart?.(nodeId, port.id, port.type);
    onMouseDown?.(nodeId, port.id, port.type, e);
  };

  const handleMouseUp = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    onConnectionEnd?.(nodeId, port.id, port.type);
    onMouseUp?.(nodeId, port.id, port.type, e);
  };

  return (
    <div
      className={cn("absolute flex items-center justify-center", className)}
      style={{
        left: "50%",
        top: "50%",
        transform: `translate(calc(-50% + ${offset.x}px), calc(-50% + ${offset.y}px))`,
      }}
    >
      {/* Outer ring for hover/connection */}
      <div
        className={cn(
          "absolute rounded-full transition-all duration-150",
          isHovered && !isConnected && "scale-150 opacity-60",
          !isValidConnection && "opacity-30"
        )}
        style={{
          width: 16,
          height: 16,
          backgroundColor: "transparent",
          border: `2px solid ${isValidConnection ? dataTypeColor : "#ef4444"}`,
          opacity: isHovered ? 0.6 : 0,
        }}
      />

      {/* Port dot */}
      <div
        className={cn(
          "size-3 rounded-full border-2 transition-all duration-150 cursor-crosshair",
          "relative z-10",
          isConnected
            ? "bg-[#10b981] border-[#10b981]"
            : isValidConnection
            ? "bg-[#0d0d14] border-[#6b7280] hover:border-brand-500 hover:scale-125"
            : "bg-[#0d0d14] border-[#ef4444]",
          isHovered && "scale-125"
        )}
        style={{
          borderColor: isConnected ? "#10b981" : dataTypeColor,
          boxShadow: isConnected
            ? `0 0 8px rgba(16,185,129,0.4)`
            : isHovered
            ? `0 0 8px ${dataTypeColor}`
            : "none",
        }}
        onMouseDown={handleMouseDown}
        onMouseUp={handleMouseUp}
        title={`${port.label} (${port.dataType || "any"})${port.multiple ? " • multiple" : ""}`}
      />

      {/* Connection indicator */}
      {isConnected && (
        <div
          className="absolute -inset-1 rounded-full animate-pulse"
          style={{
            background: `radial-gradient(circle, rgba(16,185,129,0.15) 0%, transparent 70%)`,
          }}
        />
      )}

      {/* Data type label on hover */}
      {isHovered && (
        <div
          className="absolute top-full mt-1 px-1.5 py-0.5 rounded text-[9px] whitespace-nowrap z-20"
          style={{
            backgroundColor: "rgba(13,13,20,0.95)",
            border: `1px solid ${dataTypeColor}40`,
            color: dataTypeColor,
          }}
        >
          {port.dataType || "any"}
        </div>
      )}
    </div>
  );
}
