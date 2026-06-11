/**
 * @functionfly/ui-futuristic
 * DigitalTwinViewport - Digital twin viewport
 */

import React, { useState, useEffect, useRef } from "react";
import { cn } from "@functionfly/ui-core";
import {
  Atom,
  Server,
  Database,
  Cpu,
  Layers,
  Network,
  Bot,
  Hexagon,
} from "lucide-react";
import type {
  DigitalTwinViewportProps,
  TwinEntity,
  TwinConnection,
  TwinState,
} from "../types";

export const DigitalTwinViewport: React.FC<DigitalTwinViewportProps> = ({
  state = {
    entities: [
      {
        id: "1",
        type: "Server",
        position: { x: 100, y: 100, z: 0 },
        status: "healthy",
      },
      {
        id: "2",
        type: "Database",
        position: { x: 200, y: 80, z: 0 },
        status: "healthy",
      },
      {
        id: "3",
        type: "Cache",
        position: { x: 150, y: 150, z: 0 },
        status: "warning",
      },
      {
        id: "4",
        type: "Queue",
        position: { x: 250, y: 120, z: 0 },
        status: "healthy",
      },
      {
        id: "5",
        type: "API",
        position: { x: 175, y: 200, z: 0 },
        status: "healthy",
      },
      {
        id: "6",
        type: "Worker",
        position: { x: 300, y: 100, z: 0 },
        status: "healthy",
      },
    ],
    connections: [
      { source: "1", target: "2", strength: 0.9 },
      { source: "1", target: "3", strength: 0.7 },
      { source: "2", target: "4", strength: 0.8 },
      { source: "3", target: "5", strength: 0.6 },
      { source: "1", target: "5", strength: 0.85 },
      { source: "5", target: "6", strength: 0.75 },
      { source: "4", target: "6", strength: 0.65 },
    ],
    timestamp: Date.now(),
  },
  selectedEntityId = null,
  isAnimating = true,
  cameraRotation = { x: 0, y: 0 },
  onEntitySelect,
  onConnectionHover,
  className,
}) => {
  const [rotation, setRotation] = useState(cameraRotation);
  const [hoveredEntity, setHoveredEntity] = useState<string | null>(null);
  const [hoveredConnection, setHoveredConnection] = useState<string | null>(
    null,
  );
  const animationRef = useRef<number | undefined>(undefined);

  useEffect(() => {
    if (!isAnimating) return;

    const animate = (_timestamp?: number) => {
      setRotation((prev) => ({
        x: prev.x + 0.1,
        y: prev.y + 0.15,
      }));

      animationRef.current = requestAnimationFrame(animate);
    };

    animationRef.current = requestAnimationFrame(animate);
    return () => {
      if (animationRef.current) cancelAnimationFrame(animationRef.current);
    };
  }, [isAnimating]);

  const getEntityColor = (
    status: string,
  ): { primary: string; glow: string } => {
    switch (status) {
      case "healthy":
        return { primary: "#22c55e", glow: "rgba(34,197,94,0.5)" };
      case "warning":
        return { primary: "#f59e0b", glow: "rgba(245,158,11,0.5)" };
      case "critical":
        return { primary: "#ef4444", glow: "rgba(239,68,68,0.5)" };
      default:
        return { primary: "#6b7280", glow: "rgba(107,114,128,0.5)" };
    }
  };

  const getEntityIcon = (type: string) => {
    switch (type) {
      case "Server":
        return <Server className="w-4 h-4" />;
      case "Database":
        return <Database className="w-4 h-4" />;
      case "Cache":
        return <Cpu className="w-4 h-4" />;
      case "Queue":
        return <Layers className="w-4 h-4" />;
      case "API":
        return <Network className="w-4 h-4" />;
      case "Worker":
        return <Bot className="w-4 h-4" />;
      default:
        return <Hexagon className="w-4 h-4" />;
    }
  };

  const getConnectionId = (source: string, target: string) =>
    `${source}-${target}`;

  return (
    <div
      className={cn(
        "relative rounded-xl overflow-hidden bg-slate-900/95",
        "border border-slate-700/50",
        "shadow-[0_0_40px_rgba(0,0,0,0.5)]",
        className,
      )}
    >
      {/* Header */}
      <div className="absolute top-3 left-3 z-20 flex items-center gap-2 px-3 py-1.5 rounded-lg bg-slate-800/90 backdrop-blur-sm border border-slate-600/50">
        <Atom className="w-4 h-4 text-cyan-400" />
        <span className="text-xs text-cyan-300 font-mono">DIGITAL TWIN</span>
      </div>

      {/* Status badges */}
      <div className="absolute top-3 right-3 z-20 flex items-center gap-2">
        <div className="flex items-center gap-1.5 px-2 py-1 rounded-lg bg-slate-800/90 backdrop-blur-sm border border-slate-600/50">
          <div className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />
          <span className="text-[10px] text-slate-300">
            {state.entities.filter((e) => e.status === "healthy").length}{" "}
            Healthy
          </span>
        </div>
        {state.entities.some((e) => e.status === "warning") && (
          <div className="flex items-center gap-1.5 px-2 py-1 rounded-lg bg-slate-800/90 backdrop-blur-sm border border-amber-500/50">
            <div className="w-2 h-2 rounded-full bg-amber-400 animate-pulse" />
            <span className="text-[10px] text-amber-300">
              {state.entities.filter((e) => e.status === "warning").length}{" "}
              Warning
            </span>
          </div>
        )}
      </div>

      {/* Main SVG */}
      <svg
        className="w-full h-64"
        viewBox="0 0 400 300"
        style={{
          transform: `perspective(800px) rotateX(${rotation.x}deg) rotateY(${rotation.y}deg)`,
          transformStyle: "preserve-3d",
        }}
      >
        <defs>
          <pattern
            id="twin-grid"
            width="20"
            height="20"
            patternUnits="userSpaceOnUse"
          >
            <path
              d="M 20 0 L 0 0 0 20"
              fill="none"
              stroke="rgba(100,116,139,0.15)"
              strokeWidth="0.5"
            />
          </pattern>

          <filter id="twin-glow" x="-50%" y="-50%" width="200%" height="200%">
            <feGaussianBlur stdDeviation="3" result="coloredBlur" />
            <feMerge>
              <feMergeNode in="coloredBlur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
        </defs>
        <rect width="100%" height="100%" fill="url(#twin-grid)" />

        {/* Connections */}
        {state.connections.map((conn) => {
          const source = state.entities.find((e) => e.id === conn.source);
          const target = state.entities.find((e) => e.id === conn.target);
          if (!source || !target) return null;

          const connId = getConnectionId(conn.source, conn.target);
          const isHovered = hoveredConnection === connId;
          const isSelected =
            selectedEntityId === conn.source ||
            selectedEntityId === conn.target;

          return (
            <g
              key={connId}
              onMouseEnter={() => setHoveredConnection(connId)}
              onMouseLeave={() => setHoveredConnection(null)}
              className="cursor-pointer"
            >
              <line
                x1={source.position.x}
                y1={source.position.y}
                x2={target.position.x}
                y2={target.position.y}
                stroke={
                  isHovered || isSelected ? "#06b6d4" : "rgba(100,116,139,0.5)"
                }
                strokeWidth={isHovered ? 3 : conn.strength * 2}
                strokeOpacity={conn.strength}
                filter={isHovered ? "url(#twin-glow)" : undefined}
              />
            </g>
          );
        })}

        {/* Entities */}
        {state.entities.map((entity) => {
          const colors = getEntityColor(entity.status);
          const isSelected = selectedEntityId === entity.id;
          const isHovered = hoveredEntity === entity.id;

          return (
            <g
              key={entity.id}
              onClick={() => onEntitySelect?.(entity)}
              onMouseEnter={() => setHoveredEntity(entity.id)}
              onMouseLeave={() => setHoveredEntity(null)}
              className="cursor-pointer"
            >
              {/* Glow */}
              <circle
                cx={entity.position.x}
                cy={entity.position.y}
                r="20"
                fill={colors.glow}
                fillOpacity="0.2"
              />

              {/* Main shape */}
              <g
                transform={`translate(${entity.position.x - 12}, ${entity.position.y - 12})`}
              >
                <polygon
                  points="12,0 24,6 24,18 12,24 0,18 0,6"
                  fill="rgba(15,23,42,0.9)"
                  stroke={isSelected ? "#06b6d4" : colors.primary}
                  strokeWidth={isSelected || isHovered ? 2 : 1}
                  className="transition-all duration-200"
                />

                <g
                  transform="translate(4, 4)"
                  className="text-slate-300"
                  style={{ color: colors.primary }}
                >
                  {getEntityIcon(entity.type)}
                </g>
              </g>

              {/* Label on hover */}
              {(isHovered || isSelected) && (
                <g
                  transform={`translate(${entity.position.x - 30}, ${entity.position.y + 30})`}
                >
                  <rect
                    x="0"
                    y="0"
                    width="60"
                    height="20"
                    rx="4"
                    fill="rgba(15,23,42,0.95)"
                    stroke={colors.primary}
                    strokeWidth="1"
                  />
                  <text
                    x="30"
                    y="14"
                    textAnchor="middle"
                    className="text-[10px] fill-white"
                  >
                    {entity.type}
                  </text>
                </g>
              )}
            </g>
          );
        })}
      </svg>

      {/* Footer */}
      <div className="absolute bottom-0 left-0 right-0 px-4 py-2 bg-slate-800/90 border-t border-slate-700/50">
        <div className="flex items-center justify-between text-[10px] text-slate-400">
          <div className="flex gap-4">
            <span>{state.entities.length} Entities</span>
            <span>{state.connections.length} Connections</span>
          </div>
          <span>
            Last Update: {new Date(state.timestamp).toLocaleTimeString()}
          </span>
        </div>
      </div>
    </div>
  );
};
