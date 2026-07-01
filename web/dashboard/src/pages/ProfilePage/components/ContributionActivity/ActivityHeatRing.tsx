/**
 * Activity Heat Ring Component
 *
 * A circular ring-based contribution visualization that replaces the
 * standard GitHub-style grid. Segments represent days arranged in
 * concentric rings (inner = oldest, outer = most recent).
 */

import { useMemo, useState } from "react";
import { arc as d3arc } from "d3-shape";
import { format } from "date-fns";
import { motion } from "framer-motion";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";
import type { UserProfile } from "@/types";

export interface ActivityHeatRingProps {
  data: UserProfile["stats"]["contributionGraph"];
  size?: number;
}

const LEVEL_COLORS = [
  "fill-panel-edge/40",
  "fill-status-ok/25",
  "fill-status-ok/45",
  "fill-status-ok/70",
  "fill-status-ok",
];

const LEVEL_STROKE = [
  "stroke-panel-edge/30",
  "stroke-status-ok/30",
  "stroke-status-ok/50",
  "stroke-status-ok/75",
  "stroke-status-ok",
];

const LEGEND_COLORS = [
  "bg-panel-edge",
  "bg-status-ok/20",
  "bg-status-ok/40",
  "bg-status-ok/60",
  "bg-status-ok",
];

const arcGenerator = d3arc();

export function ActivityHeatRing({ data, size = 220 }: ActivityHeatRingProps) {
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);

  const cx = size / 2;
  const cy = size / 2;
  const ringCount = 4;
  const ringGap = 4;
  const availableRadius = size / 2 - 16;
  const segmentWidth = (availableRadius - ringGap * (ringCount - 1)) / ringCount;
  const totalSegments = data.length;
  const segmentsPerRing = Math.ceil(totalSegments / ringCount);

  const segments = useMemo(() => {
    const result: { path: string; index: number; level: number; count: number; date: string }[] =
      [];

    for (let i = 0; i < data.length; i++) {
      const ring = Math.floor(i / segmentsPerRing);
      const posInRing = i % segmentsPerRing;
      const angleStep = (2 * Math.PI) / segmentsPerRing;
      const startAngle = posInRing * angleStep - Math.PI / 2;
      const endAngle = startAngle + angleStep - 0.02;

      const innerR = 20 + ring * (segmentWidth + ringGap);
      const outerR = innerR + segmentWidth;

      const path =
        arcGenerator({
          innerRadius: innerR,
          outerRadius: outerR,
          startAngle,
          endAngle,
        }) ?? "";

      result.push({
        path,
        index: i,
        level: data[i].level,
        count: data[i].count,
        date: data[i].date,
      });
    }

    return result;
  }, [data, segmentWidth, segmentsPerRing, ringGap]);

  return (
    <TooltipProvider delayDuration={100}>
      <div className="relative inline-flex items-center justify-center">
        <svg
          width={size}
          height={size}
          viewBox={`0 0 ${size} ${size}`}
          className="overflow-visible"
        >
          <g transform={`translate(${cx},${cy})`}>
            {segments.map((seg, i) => (
              <Tooltip key={seg.index}>
                <TooltipTrigger asChild>
                  <motion.path
                    d={seg.path}
                    className={cn(
                      LEVEL_COLORS[seg.level],
                      LEVEL_STROKE[seg.level],
                      `ca-heat-fill-${seg.level}`,
                      `ca-heat-stroke-${seg.level}`,
                      "cursor-pointer transition-all duration-150",
                      hoveredIndex === seg.index && "opacity-80 brightness-125"
                    )}
                    strokeWidth={0.5}
                    initial={{ opacity: 0, scale: 0.8 }}
                    animate={{ opacity: 1, scale: 1 }}
                    transition={{ delay: i * 0.003, duration: 0.3 }}
                    onMouseEnter={() => setHoveredIndex(seg.index)}
                    onMouseLeave={() => setHoveredIndex(null)}
                  />
                </TooltipTrigger>
                <TooltipContent
                  side="top"
                  sideOffset={8}
                  className="z-50 px-3 py-1.5 rounded-[var(--radius)] text-xs pointer-events-none whitespace-nowrap"
                  style={{ background: 'var(--panel)', border: '1px solid var(--panel-edge)', boxShadow: 'var(--shadow-chamber)', color: 'var(--text)' }}
                >
                  {seg.count} contributions on {format(new Date(seg.date), "MMM d, yyyy")}
                </TooltipContent>
              </Tooltip>
            ))}
          </g>

          {/* Center label */}
          <text
            x={cx}
            y={cy - 6}
            textAnchor="middle"
            className="font-mono"
            style={{ fontSize: "14px", fontWeight: 700, fill: 'var(--text)' }}
          >
            {data.reduce((s, d) => s + d.count, 0).toLocaleString()}
          </text>
          <text
            x={cx}
            y={cy + 10}
            textAnchor="middle"
            style={{ fontSize: "9px", fill: 'var(--text-faint)' }}
          >
            contributions
          </text>
        </svg>

        {/* Legend */}
        <div className="flex items-center gap-2 mt-1 text-xs" style={{ color: 'var(--text-faint)' }}>
          <span>Less</span>
          {[0, 1, 2, 3, 4].map((level) => (
            <div
              key={level}
              className={cn("w-3 h-3 rounded-sm", LEGEND_COLORS[level], `ca-legend-${level}`)}
            />
          ))}
          <span>More</span>
        </div>
      </div>
    </TooltipProvider>
  );
}
