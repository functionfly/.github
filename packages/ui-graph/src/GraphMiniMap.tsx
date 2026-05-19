/**
 * @functionfly/ui-graph
 * GraphMiniMap - bird's eye view minimap with viewport navigation
 */

import * as React from "react";
import { cn } from "./utils";
import type { GraphMiniMapProps, GraphMiniMapNode, NodeType } from "./types";

const NODE_COLORS: Record<NodeType, string> = {
  function: "#f97316",
  agent: "#8b5cf6",
  api: "#3b82f6",
  memory: "#10b981",
  database: "#ef4444",
  robot: "#f59e0b",
  browser: "#06b6d4",
  gpu: "#ec4899",
  workflow: "#a855f7",
  trigger: "#6366f1",
  condition: "#14b8a6",
  output: "#64748b",
};

const STATUS_COLORS = {
  idle: "#6b7280",
  running: "#3b82f6",
  completed: "#10b981",
  error: "#ef4444",
  waiting: "#f59e0b",
};

export function GraphMiniMap({
  nodes,
  edges,
  viewport,
  canvasWidth,
  canvasHeight,
  onNavigate,
  className,
}: GraphMiniMapProps) {
  const containerRef = React.useRef<HTMLDivElement>(null);
  const minimapW = 200;
  const minimapH = 150;
  const [bounds, setBounds] = React.useState({ x: 0, y: 0, w: 800, h: 600 });

  // Calculate bounds
  React.useEffect(() => {
    if (nodes.length === 0) {
      setBounds({ x: 0, y: 0, w: 800, h: 600 });
      return;
    }

    let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity;
    for (const n of nodes) {
      if (n.position) {
        minX = Math.min(minX, n.position.x);
        minY = Math.min(minY, n.position.y);
        maxX = Math.max(maxX, n.position.x + 200);
        maxY = Math.max(maxY, n.position.y + 120);
      }
    }
    const pad = 100;
    setBounds({ x: minX - pad, y: minY - pad, w: maxX - minX + pad * 2, h: maxY - minY + pad * 2 });
  }, [nodes]);

  const scaleX = minimapW / bounds.w;
  const scaleY = minimapH / bounds.h;

  const handleClick = (e: React.MouseEvent) => {
    const rect = (e.target as HTMLElement).getBoundingClientRect();
    const mx = e.clientX - rect.left;
    const my = e.clientY - rect.top;
    const canvasX = mx / scaleX + bounds.x;
    const canvasY = my / scaleY + bounds.y;
    const screenW = canvasWidth / viewport.zoom;
    const screenH = canvasHeight / viewport.zoom;

    onNavigate?.({
      x: -(canvasX - screenW / 2) * viewport.zoom + canvasWidth / 2,
      y: -(canvasY - screenH / 2) * viewport.zoom + canvasHeight / 2,
      zoom: viewport.zoom,
    });
  };

  return (
    <div
      className={cn(
        "absolute bottom-4 right-4 z-20 rounded-lg overflow-hidden",
        "bg-[rgba(10,10,15,0.85)] backdrop-blur-sm",
        "border border-[rgba(249,115,22,0.2)]",
        className
      )}
      style={{ width: minimapW, height: minimapH }}
    >
      <div ref={containerRef} onClick={handleClick} className="relative cursor-pointer" style={{ width: "100%", height: "100%" }}>
        {/* Grid */}
        <svg width="100%" height="100%" className="absolute inset-0">
          {nodes.map((node) =>
            node.position ? (
              <rect
                key={node.id}
                x={(node.position.x - bounds.x) * scaleX}
                y={(node.position.y - bounds.y) * scaleY}
                width={Math.max(200 * scaleX, 3)}
                height={Math.max(120 * scaleY, 2)}
                fill={NODE_COLORS[node.type] || "#6b7280"}
                fillOpacity={node.status === "running" ? 0.9 : 0.4}
                stroke={node.isSelected ? "#f97316" : "transparent"}
                strokeWidth={0.5}
                rx={1}
              />
            ) : null
          )}

          {/* Edges */}
          {edges.map((edge) => {
            const sourceNode = nodes.find((n) => n.id === edge.source);
            const targetNode = nodes.find((n) => n.id === edge.target);
            if (!sourceNode?.position || !targetNode?.position) return null;

            return (
              <line
                key={edge.id}
                x1={(sourceNode.position.x - bounds.x) * scaleX + 100 * scaleX}
                y1={(sourceNode.position.y - bounds.y) * scaleY + 60 * scaleY}
                x2={(targetNode.position.x - bounds.x) * scaleX + 100 * scaleX}
                y2={(targetNode.position.y - bounds.y) * scaleY + 60 * scaleY}
                stroke="rgba(255,255,255,0.15)"
                strokeWidth={0.5}
              />
            );
          })}
        </svg>

        {/* Viewport rectangle */}
        <div
          className="absolute border pointer-events-none"
          style={{
            left: ((-viewport.x / viewport.zoom - bounds.x) * scaleX),
            top: ((-viewport.y / viewport.zoom - bounds.y) * scaleY),
            width: (canvasWidth / viewport.zoom) * scaleX,
            height: (canvasHeight / viewport.zoom) * scaleY,
            borderColor: "#f97316",
            backgroundColor: "rgba(249,115,22,0.08)",
          }}
        >
          {/* Viewport corner accents */}
          <div className="absolute -top-0.5 -left-0.5 size-1.5 rounded-full bg-[#f97316]" />
          <div className="absolute -top-0.5 -right-0.5 size-1.5 rounded-full bg-[#f97316]" />
          <div className="absolute -bottom-0.5 -left-0.5 size-1.5 rounded-full bg-[#f97316]" />
          <div className="absolute -bottom-0.5 -right-0.5 size-1.5 rounded-full bg-[#f97316]" />
        </div>
      </div>

      {/* Zoom indicator */}
      <div className="absolute bottom-1 right-1 px-1 py-0.5 rounded bg-[rgba(0,0,0,0.5)] text-[9px] text-[#6b7280] font-mono">
        {Math.round(viewport.zoom * 100)}%
      </div>
    </div>
  );
}
