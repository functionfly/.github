/**
 * @functionfly/ui-graph
 * ExecutionEdge - connection between nodes
 */

import * as React from "react";
import { cn } from "./utils";
import { getEdgePath } from "./utils";
import type { EdgeData } from "./types";

export interface ExecutionEdgeProps extends EdgeData {
  animated?: boolean;
  tokenFlow?: boolean;
  tokenPosition?: number;
  onClick?: (id: string) => void;
  onSelect?: (id: string) => void;
  onRemove?: (id: string) => void;
  className?: string;
}

export function ExecutionEdge({
  id,
  source,
  target,
  sourcePort,
  targetPort,
  label,
  status = "idle",
  animated = false,
  tokenFlow = false,
  tokenPosition = 0,
  isSelected = false,
  onClick,
  onSelect,
  className,
}: ExecutionEdgeProps) {
  const [sourcePos, setSourcePos] = React.useState({ x: 0, y: 0 });
  const [targetPos, setTargetPos] = React.useState({ x: 0, y: 0 });

  React.useEffect(() => {
    const updatePositions = () => {
      const sourceEl = document.querySelector(`[data-node-id="${source}"]`);
      const targetEl = document.querySelector(`[data-node-id="${target}"]`);

      if (sourceEl && targetEl) {
        const sRect = sourceEl.getBoundingClientRect();
        const tRect = targetEl.getBoundingClientRect();
        setSourcePos({ x: sRect.right, y: sRect.top + sRect.height / 3 });
        setTargetPos({ x: tRect.left, y: tRect.top + tRect.height / 3 });
      }
    };

    updatePositions();
    window.addEventListener("resize", updatePositions);
    window.addEventListener("scroll", updatePositions, true);

    const observer = new MutationObserver(updatePositions);
    const sourceParent = document.querySelector(`[data-node-id="${source}"]`);
    const targetParent = document.querySelector(`[data-node-id="${target}"]`);
    if (sourceParent) observer.observe(sourceParent, { attributes: true, attributeFilter: ["style"] });
    if (targetParent) observer.observe(targetParent, { attributes: true, attributeFilter: ["style"] });

    return () => {
      window.removeEventListener("resize", updatePositions);
      window.removeEventListener("scroll", updatePositions, true);
      observer.disconnect();
    };
  }, [source, target]);

  const statusColor = {
    idle: "rgba(255, 255, 255, 0.15)",
    active: "#f97316",
    error: "#ef4444",
  }[status] || "rgba(255, 255, 255, 0.15)";

  const path = getEdgePath(sourcePos.x, sourcePos.y, targetPos.x, targetPos.y);

  return (
    <g className={cn("execution-edge", className)}>
      {/* Shadow line */}
      <path
        d={path}
        fill="none"
        stroke={statusColor}
        strokeWidth={isSelected ? 3 : 2}
        className="edge-shadow"
        opacity={0.3}
        filter="blur(2px)"
      />

      {/* Main edge */}
      <path
        d={path}
        fill="none"
        stroke={statusColor}
        strokeWidth={isSelected ? 2.5 : 1.5}
        strokeOpacity={isSelected ? 1 : 0.5}
        className="edge-path"
        strokeLinecap="round"
        onClick={() => onClick?.(id)}
        style={{ cursor: "pointer" }}
      />

      {/* Animated token dot */}
      {tokenFlow && (
        <>
          <defs>
            <radialGradient id={`glow-${id}`}>
              <stop offset="0%" stopColor="#f97316" stopOpacity="0.8" />
              <stop offset="100%" stopColor="#f97316" stopOpacity="0" />
            </radialGradient>
          </defs>
          <circle
            r={10}
            fill={`url(#glow-${id})`}
            opacity={0.3}
            style={{
              offsetPath: `path('${path}')`,
              offsetDistance: `${((tokenPosition * 3) % 1) * 100}%`,
            }}
          >
            <animate
              attributeName="offset-distance"
              values="0%;100%"
              dur={animated ? "1.5s" : "0.1s"}
              repeatCount="indefinite"
            />
          </circle>
          <circle
            r={3}
            fill="#f97316"
            className="token-dot"
            filter={`url(#glow-${id})`}
            style={{
              offsetPath: `path('${path}')`,
              offsetDistance: `${((tokenPosition * 3 + 0.33) % 1) * 100}%`,
            }}
          >
            <animate
              attributeName="offset-distance"
              values="0%;100%"
              dur={animated ? "1.5s" : "0.1s"}
              repeatCount="indefinite"
            />
          </circle>
          <circle
            r={3}
            fill="#fbbf24"
            className="token-dot"
            style={{
              offsetPath: `path('${path}')`,
              offsetDistance: `${((tokenPosition * 3 + 0.66) % 1) * 100}%`,
            }}
          >
            <animate
              attributeName="offset-distance"
              values="0%;100%"
              dur={animated ? "1.5s" : "0.1s"}
              repeatCount="indefinite"
            />
          </circle>
        </>
      )}

      {/* Edge label */}
      {label && (
        <text
          x={(sourcePos.x + targetPos.x) / 2}
          y={(sourcePos.y + targetPos.y) / 2 - 8}
          textAnchor="middle"
          fontSize={10}
          fill="#6b6b7b"
          className="edge-label"
        >
          {label}
        </text>
      )}
    </g>
  );
}