/**
 * @functionfly/ui-futuristic
 * AIThoughtWave - AI thought wave visualization
 */

import React, { useState, useEffect, useRef } from "react";
import { cn } from "@functionfly/ui-core";
import { Brain } from "lucide-react";
import type { AIThoughtWaveProps, ThoughtWavePoint } from "../types";

export const AIThoughtWave: React.FC<AIThoughtWaveProps> = ({
  points = Array.from({ length: 20 }, (_, i) => ({
    timestamp: Date.now() - (20 - i) * 500,
    amplitude: Math.sin(i * 0.5) * 50 + Math.random() * 20 + 30,
    frequency: 0.5 + Math.random() * 0.5,
  })),
  isActive = true,
  color = "#06b6d4",
  showGrid = true,
  onPointHover,
  className,
}) => {
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);
  const [animatedPoints, setAnimatedPoints] =
    useState<ThoughtWavePoint[]>(points);
  const animationRef = useRef<number | undefined>(undefined);

  useEffect(() => {
    if (!isActive) return;

    const animate = (_timestamp?: number) => {
      setAnimatedPoints((prev) => {
        const newPoints = [...prev];
        newPoints.shift();
        newPoints.push({
          timestamp: Date.now(),
          amplitude:
            Math.sin(Date.now() * 0.005) * 50 + Math.random() * 20 + 30,
          frequency: 0.5 + Math.random() * 0.5,
        });
        return newPoints;
      });

      animationRef.current = requestAnimationFrame(animate);
    };

    animationRef.current = requestAnimationFrame(animate);
    return () => {
      if (animationRef.current) cancelAnimationFrame(animationRef.current);
    };
  }, [isActive]);

  const maxAmplitude = Math.max(...animatedPoints.map((p) => p.amplitude), 100);
  const svgHeight = 120;
  const svgWidth = 100;

  const pathD = animatedPoints
    .map((p, i) => {
      const x = (i / (animatedPoints.length - 1)) * svgWidth;
      const y = svgHeight - (p.amplitude / maxAmplitude) * svgHeight;
      return `${i === 0 ? "M" : "L"} ${x} ${y}`;
    })
    .join(" ");

  const areaD = pathD + ` L ${svgWidth} ${svgHeight} L 0 ${svgHeight} Z`;

  return (
    <div
      className={cn(
        "relative p-4 rounded-xl bg-slate-900/95",
        "border border-cyan-500/30",
        "shadow-[0_0_20px_rgba(6,182,212,0.2)]",
        className,
      )}
    >
      {/* Header */}
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <Brain className="w-4 h-4 text-cyan-400" />
          <span className="text-xs text-cyan-300 font-mono">
            NEURAL ACTIVITY
          </span>
        </div>
        <div
          className={cn(
            "flex items-center gap-1.5 px-2 py-1 rounded-full bg-slate-800/80 border border-cyan-500/30",
          )}
        >
          <div
            className={cn(
              "w-2 h-2 rounded-full",
              isActive ? "bg-green-400 animate-pulse" : "bg-slate-500",
            )}
          />
          <span className="text-[10px] text-cyan-300">
            {isActive ? "ACTIVE" : "IDLE"}
          </span>
        </div>
      </div>

      {/* Chart */}
      <div className="relative">
        {/* Grid */}
        {showGrid && (
          <svg
            className="absolute inset-0 w-full h-full"
            viewBox={`0 0 ${svgWidth} ${svgHeight}`}
            preserveAspectRatio="none"
          >
            {[0, 25, 50, 75, 100].map((y) => (
              <line
                key={y}
                x1="0"
                y1={(y / 100) * svgHeight}
                x2={svgWidth}
                y2={(y / 100) * svgHeight}
                stroke="rgba(100,116,139,0.2)"
                strokeWidth="0.5"
                strokeDasharray="2 2"
              />
            ))}
            {[0, 25, 50, 75, 100].map((x) => (
              <line
                key={x}
                x1={(x / 100) * svgWidth}
                y1="0"
                x2={(x / 100) * svgWidth}
                y2={svgHeight}
                stroke="rgba(100,116,139,0.2)"
                strokeWidth="0.5"
                strokeDasharray="2 2"
              />
            ))}
          </svg>
        )}

        {/* Wave area */}
        <svg
          className="w-full h-28"
          viewBox={`0 0 ${svgWidth} ${svgHeight}`}
          preserveAspectRatio="none"
        >
          <defs>
            <linearGradient id="waveGradient" x1="0%" y1="0%" x2="0%" y2="100%">
              <stop offset="0%" stopColor={color} stopOpacity="0.4" />
              <stop offset="100%" stopColor={color} stopOpacity="0" />
            </linearGradient>
          </defs>

          {/* Area fill */}
          <path d={areaD} fill="url(#waveGradient)" />

          {/* Wave line */}
          <path
            d={pathD}
            fill="none"
            stroke={color}
            strokeWidth="2"
            className="drop-shadow-[0_0_4px_rgba(6,182,212,0.8)]"
          />

          {/* Active point indicator */}
          {hoveredIndex !== null && animatedPoints[hoveredIndex] && (
            <g>
              <circle
                cx={(hoveredIndex / (animatedPoints.length - 1)) * svgWidth}
                cy={
                  svgHeight -
                  (animatedPoints[hoveredIndex].amplitude / maxAmplitude) *
                    svgHeight
                }
                r="3"
                fill={color}
                className="animate-ping"
              />
              <circle
                cx={(hoveredIndex / (animatedPoints.length - 1)) * svgWidth}
                cy={
                  svgHeight -
                  (animatedPoints[hoveredIndex].amplitude / maxAmplitude) *
                    svgHeight
                }
                r="2"
                fill={color}
              />
            </g>
          )}
        </svg>
      </div>

      {/* Stats */}
      <div className="flex items-center justify-between mt-3 text-[10px] text-slate-500">
        <div>
          <div className="text-[10px] text-slate-500">AVG AMP</div>
          <div className="text-xs text-cyan-300">
            {(
              animatedPoints.reduce((a, b) => a + b.amplitude, 0) /
              animatedPoints.length
            ).toFixed(1)}
          </div>
        </div>
        <div>
          <div className="text-[10px] text-slate-500">PEAK</div>
          <div className="text-xs text-cyan-300">
            {Math.max(...animatedPoints.map((p) => p.amplitude)).toFixed(1)}
          </div>
        </div>
        <div>
          <div className="text-[10px] text-slate-500">FREQ</div>
          <div className="text-xs text-cyan-300">
            {(
              animatedPoints.reduce((a, b) => a + b.frequency, 0) /
              animatedPoints.length
            ).toFixed(2)}
            Hz
          </div>
        </div>
      </div>
    </div>
  );
};
