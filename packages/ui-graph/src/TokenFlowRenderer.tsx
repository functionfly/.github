/**
 * @functionfly/ui-graph
 * TokenFlowRenderer - animated token particles flowing along edges
 */

import * as React from "react";
import { cn } from "./utils";
import { getEdgePath } from "./utils";
import type { TokenFlowRendererProps, EdgeData, TokenFlowParticle } from "./types";

const TOKEN_COLORS = {
  idle: "#6b7280",
  active: "#f97316",
  success: "#10b981",
  error: "#ef4444",
  processing: "#3b82f6",
};

export function TokenFlowRenderer({
  edges,
  activeTokens = 3,
  tokenFlowSpeed = 1,
  showTokenPath = false,
  onTokenClick,
  className,
}: TokenFlowRendererProps) {
  const [tokenPositions, setTokenPositions] = React.useState<Record<string, number>>({});
  const frameRef = React.useRef<number | undefined>(undefined);

  // Initialize token positions
  React.useEffect(() => {
    const initial: Record<string, number> = {};
    edges.forEach((edge, i) => {
      initial[edge.id] = (i / edges.length);
    });
    setTokenPositions(initial);
  }, [edges]);

  // Animation loop
  React.useEffect(() => {
    let lastTime = 0;
    const animate = (time: number) => {
      if (lastTime === 0) lastTime = time;
      const delta = (time - lastTime) * 0.001 * tokenFlowSpeed;
      lastTime = time;

      setTokenPositions((prev) => {
        const next = { ...prev };
        edges.forEach((edge) => {
          next[edge.id] = (prev[edge.id] + delta) % 1;
        });
        return next;
      });

      frameRef.current = requestAnimationFrame(animate);
    };

    frameRef.current = requestAnimationFrame(animate);
    return () => {
      if (frameRef.current) cancelAnimationFrame(frameRef.current);
    };
  }, [edges, tokenFlowSpeed]);

  const getEdgePositions = (edge: EdgeData) => {
    const sourceEl = document.querySelector(`[data-node-id="${edge.source}"]`);
    const targetEl = document.querySelector(`[data-node-id="${edge.target}"]`);

    if (!sourceEl || !targetEl) {
      return { sourceX: 0, sourceY: 0, targetX: 100, targetY: 0 };
    }

    const sRect = sourceEl.getBoundingClientRect();
    const tRect = targetEl.getBoundingClientRect();
    const canvasRect = document.querySelector(".runtime-graph-canvas")?.getBoundingClientRect();

    const canvasX = canvasRect?.left ?? 0;
    const canvasY = canvasRect?.top ?? 0;

    return {
      sourceX: (sRect.right - canvasX),
      sourceY: (sRect.top - canvasY) + sRect.height / 3,
      targetX: (tRect.left - canvasX),
      targetY: (tRect.top - canvasY) + tRect.height / 3,
    };
  };

  return (
    <svg className={cn("absolute inset-0 w-full h-full pointer-events-none", className)}>
      <defs>
        {/* Glow filter */}
        <filter id="token-glow" x="-100%" y="-100%" width="300%" height="300%">
          <feGaussianBlur in="SourceGraphic" stdDeviation="3" result="blur" />
          <feFlood floodColor="#f97316" floodOpacity="0.5" result="color" />
          <feComposite in="color" in2="blur" operator="in" result="glow" />
          <feMerge>
            <feMergeNode in="glow" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>

        {/* Radial gradient for token glow */}
        <radialGradient id="token-gradient">
          <stop offset="0%" stopColor="#f97316" stopOpacity="0.8" />
          <stop offset="100%" stopColor="#f97316" stopOpacity="0" />
        </radialGradient>
      </defs>

      {/* Edge paths */}
      {edges.map((edge) => {
        const pos = getEdgePositions(edge);
        const path = getEdgePath(pos.sourceX, pos.sourceY, pos.targetX, pos.targetY);
        const statusColor = edge.status ? TOKEN_COLORS[edge.status] : TOKEN_COLORS.idle;
        const progress = tokenPositions[edge.id] ?? 0;

        return (
          <g key={edge.id}>
            {/* Background path */}
            <path
              d={path}
              fill="none"
              stroke={statusColor}
              strokeWidth={2}
              strokeOpacity={0.15}
              strokeLinecap="round"
            />

            {/* Active path (highlighted by token position) */}
            <path
              d={path}
              fill="none"
              stroke={statusColor}
              strokeWidth={2}
              strokeOpacity={0.3}
              strokeLinecap="round"
              strokeDasharray={`${progress * 100} 100`}
            />

            {/* Token glow */}
            <circle
              r={12}
              fill="url(#token-gradient)"
              opacity={0.3}
              style={{
                offsetPath: `path('${path}')`,
                offsetDistance: `${progress * 100}%`,
              }}
            />

            {/* Token dot */}
            <circle
              r={4}
              fill={statusColor}
              filter="url(#token-glow)"
              style={{
                offsetPath: `path('${path}')`,
                offsetDistance: `${progress * 100}%`,
              }}
            />

            {/* Secondary token for depth effect */}
            <circle
              r={2}
              fill="#fbbf24"
              opacity={0.8}
              style={{
                offsetPath: `path('${path}')`,
                offsetDistance: `${((progress + 0.3) % 1) * 100}%`,
              }}
            />
          </g>
        );
      })}
    </svg>
  );
}
