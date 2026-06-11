/**
 * @functionfly/ui-futuristic
 * SwarmMindVisualizer - Swarm intelligence visualization
 */

import React, { useState, useEffect, useRef } from "react";
import { cn } from "@functionfly/ui-core";
import { Bot } from "lucide-react";
import type {
  SwarmMindVisualizerProps,
  SwarmAgent,
  SwarmAgentState,
} from "../types";

export const SwarmMindVisualizer: React.FC<SwarmMindVisualizerProps> = ({
  agents = Array.from({ length: 8 }, (_, i) => ({
    id: String(i + 1),
    x: Math.random() * 400,
    y: Math.random() * 300,
    velocity: { dx: (Math.random() - 0.5) * 4, dy: (Math.random() - 0.5) * 4 },
    state: ["exploring", "exploiting", "returning"][
      Math.floor(Math.random() * 3)
    ] as SwarmAgentState,
  })),
  targetX = 200,
  targetY = 150,
  isActive = true,
  onAgentClick,
  className,
}) => {
  const [swarmAgents, setSwarmAgents] = useState<SwarmAgent[]>(agents);
  const [hoveredAgent, setHoveredAgent] = useState<string | null>(null);
  const animationRef = useRef<number | undefined>(undefined);

  useEffect(() => {
    if (!isActive) return;

    const animate = (_timestamp?: number) => {
      setSwarmAgents((prev) =>
        prev.map((agent) => {
          const dx = targetX - agent.x;
          const dy = targetY - agent.y;
          const dist = Math.sqrt(dx * dx + dy * dy);

          let newState: SwarmAgentState = "exploring";
          let vx = agent.velocity.dx;
          let vy = agent.velocity.dy;

          if (dist < 100) {
            newState = "exploiting";
            vx += dx * 0.002 + (Math.random() - 0.5) * 0.5;
            vy += dy * 0.002 + (Math.random() - 0.5) * 0.5;
          } else if (dist > 200) {
            newState = "returning";
            vx += dx * 0.001;
            vy += dy * 0.001;
          } else {
            newState = "exploring";
            vx += (Math.random() - 0.5) * 1;
            vy += (Math.random() - 0.5) * 1;
          }

          // Apply velocity with damping
          vx *= 0.98;
          vy *= 0.98;

          // Clamp velocity
          const maxVel = 4;
          vx = Math.max(-maxVel, Math.min(maxVel, vx));
          vy = Math.max(-maxVel, Math.min(maxVel, vy));

          return {
            ...agent,
            x: agent.x + vx,
            y: agent.y + vy,
            velocity: { dx: vx, dy: vy },
            state: newState,
          };
        }),
      );

      animationRef.current = requestAnimationFrame(animate);
    };

    animationRef.current = requestAnimationFrame(animate);
    return () => {
      if (animationRef.current) cancelAnimationFrame(animationRef.current);
    };
  }, [isActive, targetX, targetY]);

  const getAgentColor = (
    state: SwarmAgentState,
  ): { fill: string; glow: string } => {
    switch (state) {
      case "exploring":
        return { fill: "#06b6d4", glow: "rgba(6,182,212,0.5)" };
      case "exploiting":
        return { fill: "#a855f7", glow: "rgba(168,85,247,0.5)" };
      case "returning":
        return { fill: "#22c55e", glow: "rgba(34,197,94,0.5)" };
    }
  };

  return (
    <div
      className={cn(
        "relative rounded-xl overflow-hidden bg-slate-900/95",
        "border border-slate-700/50",
        "shadow-[0_0_30px_rgba(0,0,0,0.5)]",
        className,
      )}
    >
      {/* Header */}
      <div className="absolute top-3 left-3 z-10 flex items-center gap-2 px-3 py-1.5 rounded-lg bg-slate-800/90 backdrop-blur-sm border border-slate-600/50">
        <Bot className="w-4 h-4 text-cyan-400" />
        <span className="text-xs text-cyan-300 font-mono">SWARM</span>
      </div>

      {/* Target indicator */}
      <div className="absolute top-3 right-3 z-10 px-3 py-1.5 rounded-lg bg-slate-800/90 backdrop-blur-sm border border-slate-600/50">
        <span className="text-[10px] text-slate-400">TARGET</span>
      </div>

      {/* SVG Canvas */}
      <svg className="w-full h-64" viewBox="0 0 400 300">
        {/* Grid */}
        <defs>
          <pattern
            id="swarm-grid"
            width="20"
            height="20"
            patternUnits="userSpaceOnUse"
          >
            <path
              d="M 20 0 L 0 0 0 20"
              fill="none"
              stroke="rgba(100,116,139,0.2)"
              strokeWidth="0.5"
            />
          </pattern>
        </defs>
        <rect width="100%" height="100%" fill="url(#swarm-grid)" />

        {/* Target zone */}
        <circle
          cx={targetX}
          cy={targetY}
          r="35"
          fill="none"
          stroke="#06b6d4"
          strokeWidth="1"
          strokeDasharray="4 4"
        >
          <animate
            attributeName="r"
            values="35;45;35"
            dur="2s"
            repeatCount="indefinite"
          />
        </circle>
        <circle cx={targetX} cy={targetY} r="8" fill="rgba(6,182,212,0.3)" />

        {/* Agent trails */}
        {swarmAgents.map((agent) => {
          const colors = getAgentColor(agent.state);
          return (
            <line
              key={`trail-${agent.id}`}
              x1={agent.x - agent.velocity.dx * 5}
              y1={agent.y - agent.velocity.dy * 5}
              x2={agent.x}
              y2={agent.y}
              stroke={colors.fill}
              strokeWidth="1"
              strokeOpacity="0.3"
            />
          );
        })}

        {/* Agents */}
        {swarmAgents.map((agent) => {
          const colors = getAgentColor(agent.state);
          const isHovered = hoveredAgent === agent.id;

          return (
            <g
              key={agent.id}
              onClick={() => onAgentClick?.(agent)}
              onMouseEnter={() => setHoveredAgent(agent.id)}
              onMouseLeave={() => setHoveredAgent(null)}
              className="cursor-pointer"
            >
              {/* Glow */}
              <circle
                cx={agent.x}
                cy={agent.y}
                r={isHovered ? 12 : 8}
                fill={colors.glow}
                fillOpacity="0.2"
                filter="blur(2px)"
              />

              {/* Main body */}
              <circle
                cx={agent.x}
                cy={agent.y}
                r={isHovered ? 6 : 4}
                fill={colors.fill}
                stroke={isHovered ? "#fff" : "none"}
                strokeWidth="1"
              >
                {agent.state === "exploiting" && (
                  <animate
                    attributeName="r"
                    values="4;6;4"
                    dur="0.5s"
                    repeatCount="indefinite"
                  />
                )}
              </circle>

              {/* Label on hover */}
              {isHovered && (
                <g transform={`translate(${agent.x - 20}, ${agent.y - 20})`}>
                  <rect
                    x="0"
                    y="0"
                    width="40"
                    height="16"
                    rx="4"
                    fill="rgba(15,23,42,0.9)"
                    stroke={colors.fill}
                    strokeWidth="1"
                  />
                  <text
                    x="20"
                    y="11"
                    textAnchor="middle"
                    className="text-[8px] fill-white"
                  >
                    {agent.state.toUpperCase()}
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
          <span>{swarmAgents.length} Agents Active</span>
          <span>
            Target: ({targetX}, {targetY})
          </span>
        </div>
      </div>
    </div>
  );
};
