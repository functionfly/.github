/**
 * @functionfly/ui-graph
 * ExecutionTimeline - timeline-based replay UI
 */

import * as React from "react";
import { cn } from "./utils";

export interface ExecutionTimelineProps {
  events: Array<{
    id: string;
    timestamp: number;
    type: string;
    nodeId?: string;
    nodeLabel?: string;
    result?: "success" | "failure" | "partial";
    duration?: number;
    data?: Record<string, unknown>;
  }>;
  currentTime?: number;
  onTimeChange?: (time: number) => void;
  zoom?: number;
  height?: number;
  className?: string;
}

export function ExecutionTimeline({
  events,
  currentTime,
  onTimeChange,
  zoom = 1,
  height = 120,
  className,
}: ExecutionTimelineProps) {
  const containerRef = React.useRef<HTMLDivElement>(null);
  const [hoverX, setHoverX] = React.useState<number | null>(null);

  if (events.length === 0) {
    return (
      <div className={cn("flex items-center justify-center text-text-muted h-full", className)}>
        <p>No execution events recorded yet</p>
      </div>
    );
  }

  const minTime = events[0]!.timestamp;
  const maxTime = events[events.length - 1]!.timestamp;
  const timeRange = maxTime - minTime || 1;

  const getPosition = (t: number) => ((t - minTime) / timeRange) * 100;

  const groupedEvents = events.reduce(
    (acc, event) => {
      const pct = getPosition(event.timestamp);
      const bucket = Math.floor(pct / 10);
      if (!acc[bucket]) acc[bucket] = [];
      acc[bucket].push(event);
      return acc;
    },
    {} as Record<number, typeof events>
  );

  const maxBucketCount = Math.max(...Object.values(groupedEvents).map((e) => e.length), 1);

  const fillHeight = className?.includes("h-full") || className?.includes("flex-1");

  return (
    <div className={cn("relative flex flex-col", className)}>
      {/* Timeline header */}
      <div className="flex items-center justify-between mb-2 text-xs text-text-muted shrink-0">
        <span>{new Date(minTime).toLocaleTimeString()}</span>
        <span>
          {events.length} events ·{" "}
          {Math.round((timeRange / 1000) * 10) / 10}s duration
        </span>
        <span>{new Date(maxTime).toLocaleTimeString()}</span>
      </div>

      {/* Timeline track */}
      <div
        ref={containerRef}
        className={cn(
          "relative bg-bg-secondary rounded-lg overflow-hidden cursor-crosshair min-h-[64px]",
          fillHeight && "flex-1"
        )}
        style={fillHeight ? undefined : { height }}
        onMouseMove={(e) => {
          const rect = containerRef.current?.getBoundingClientRect();
          if (rect) {
            const x = e.clientX - rect.left;
            const pct = (x / rect.width) * 100;
            const time = minTime + (pct / 100) * timeRange;
            setHoverX(x);
            onTimeChange?.(time);
          }
        }}
        onMouseLeave={() => {
          setHoverX(null);
          onTimeChange?.(maxTime);
        }}
      >
        {/* Background heatmap */}
        {Object.entries(groupedEvents).map(([bucket, bucketEvents]) => (
          <div
            key={bucket}
            className="absolute top-0 bottom-0 opacity-20"
            style={{
              left: `${Number(bucket) * 10}%`,
              width: "10%",
              backgroundColor:
                bucketEvents.some((e) => e.result === "failure")
                  ? "#ef4444"
                  : bucketEvents.some((e) => e.result === "partial")
                  ? "#f59e0b"
                  : "#10b981",
              height: `${4 + (bucketEvents.length / maxBucketCount) * 12}px`,
              top: `${8 - (bucketEvents.length / maxBucketCount) * 6}px`,
            }}
          />
        ))}

        {/* Individual events */}
        {events.map((event) => {
          const pct = getPosition(event.timestamp);
          const color =
            event.result === "failure" ? "#ef4444" : event.result === "partial" ? "#f59e0b" : "#10b981";
          const isCurrent =
            currentTime != null && Math.abs(event.timestamp - currentTime) < timeRange * 0.01;

          return (
            <div
              key={event.id}
              className="absolute top-1/2 -translate-y-1/2 cursor-pointer group"
              style={{ left: `${pct}%` }}
              onClick={() => onTimeChange?.(event.timestamp)}
            >
              <div
                className={cn(
                  "size-2.5 rounded-full transition-all duration-150",
                  isCurrent ? "ring-2 ring-brand-500" : "hover:scale-150"
                )}
                style={{ backgroundColor: color }}
              />
              {/* Tooltip */}
              <div className="absolute bottom-full mb-2 left-1/2 -translate-x-1/2 hidden group-hover:block">
                <div className="bg-bg-tertiary border border-border-subtle px-2 py-1 rounded text-[10px] text-text-primary whitespace-nowrap">
                  {event.nodeLabel || event.type} · {Math.round((event.duration || 0) * 10) / 10}ms
                </div>
              </div>
            </div>
          );
        })}

        {/* Current time cursor */}
        {currentTime != null && (
          <div
            className="absolute top-0 bottom-0 w-[2px] bg-brand-500 shadow-[0_0_8px_rgba(249,115,22,0.5)]"
            style={{ left: `${getPosition(currentTime)}%` }}
          >
            <div className="absolute -top-1 -left-1.5 size-3 rounded-full bg-brand-500 border-2 border-bg-primary" />
          </div>
        )}

        {/* Hover cursor */}
        {hoverX != null && (
          <div
            className="absolute top-0 bottom-0 w-[1px] bg-border-strong pointer-events-none"
            style={{ left: hoverX }}
          />
        )}
      </div>
    </div>
  );
}