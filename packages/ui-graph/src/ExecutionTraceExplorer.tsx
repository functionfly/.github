/**
 * @functionfly/ui-graph
 * ExecutionTraceExplorer - explore distributed trace spans
 */

import * as React from "react";
import { cn } from "./utils";
import type { ExecutionTraceExplorerProps, TraceSpan } from "./types";

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  return `${(ms / 60000).toFixed(1)}m`;
}

function formatTimestamp(ts: number): string {
  return new Date(ts).toLocaleTimeString();
}

const STATUS_ICONS = {
  ok: "✓",
  error: "✗",
};

export function ExecutionTraceExplorer({
  traceId,
  spans,
  onSpanClick,
  onExpand,
  className,
}: ExecutionTraceExplorerProps) {
  const [expandedSpans, setExpandedSpans] = React.useState<Set<string>>(new Set());
  const [selectedSpanId, setSelectedSpanId] = React.useState<string | null>(null);

  const totalDuration = React.useMemo(() => {
    const sorted = [...spans].sort((a, b) => (a.startTime > b.startTime ? 1 : -1));
    if (sorted.length === 0) return 0;
    const last = sorted[sorted.length - 1];
    return (last.endTime || last.startTime) - sorted[0].startTime;
  }, [spans]);

  const toggleExpand = (spanId: string) => {
    setExpandedSpans((prev) => {
      const next = new Set(prev);
      if (next.has(spanId)) next.delete(spanId);
      else next.add(spanId);
      return next;
    });
    onExpand?.(spanId);
  };

  return (
    <div className={cn("flex flex-col gap-3 p-4 bg-[#0d0d14] rounded-xl border border-[rgba(255,255,255,0.08)]", className)}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-base">🔍</span>
          <div>
            <div className="text-xs text-[#e8e8f0]">Trace Explorer</div>
            <div className="text-[10px] text-[#6b7280] font-mono">{traceId.slice(0, 12)}...</div>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-[10px] text-[#6b7280]">{spans.length} spans</span>
          <span className="text-[10px] text-[#f97316] font-mono">{formatDuration(totalDuration)}</span>
        </div>
      </div>

      {/* Trace spans */}
      {spans.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-6 text-center">
          <div className="text-2xl opacity-30 mb-1">📊</div>
          <p className="text-xs text-[#6b7280]">No spans in trace</p>
        </div>
      ) : (
        <div className="space-y-0.5 max-h-80 overflow-y-auto">
          {spans.map((span) => {
            const isExpanded = expandedSpans.has(span.id);
            const isSelected = selectedSpanId === span.id;
            const duration = span.duration || ((span.endTime || span.startTime) - span.startTime);
            const startOffset = span.startTime - spans[0].startTime;

            return (
              <div key={span.id}>
                <div
                  onClick={() => {
                    setSelectedSpanId(span.id);
                    onSpanClick?.(span);
                  }}
                  className={cn(
                    "flex items-center gap-2 px-2 py-1.5 rounded cursor-pointer transition-colors",
                    isSelected
                      ? "bg-[rgba(249,115,22,0.1)] border border-[rgba(249,115,22,0.2)]"
                      : "hover:bg-[rgba(255,255,255,0.04)] border border-transparent"
                  )}
                >
                  {/* Expand toggle */}
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      toggleExpand(span.id);
                    }}
                    className="size-4 flex items-center justify-center text-[#6b7280] hover:text-[#e8e8f0]"
                  >
                    <span className={`text-[8px] transition-transform ${isExpanded ? "rotate-90" : ""}`}>▶</span>
                  </button>

                  {/* Status icon */}
                  <span
                    className={cn(
                      "text-[10px]",
                      span.status === "error" ? "text-[#ef4444]" : "text-[#10b981]"
                    )}
                  >
                    {STATUS_ICONS[span.status || "ok"]}
                  </span>

                  {/* Operation name */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-xs text-[#e8e8f0] truncate">{span.operation}</span>
                      {span.nodeId && (
                        <span className="text-[9px] px-1 py-0.5 rounded bg-[#14141f] text-[#6b7280] font-mono">
                          {span.nodeId.slice(0, 6)}
                        </span>
                      )}
                    </div>
                  </div>

                  {/* Duration */}
                  <span className="text-[10px] text-[#6b7280] font-mono">{formatDuration(duration)}</span>
                </div>

                {/* Expanded details */}
                {isExpanded && (
                  <div className="ml-6 mt-1 p-2 rounded-lg bg-[#14141f] border border-[rgba(255,255,255,0.06)]">
                    {/* Span info */}
                    <div className="grid grid-cols-2 gap-2 text-[10px]">
                      <div>
                        <span className="text-[#6b7280]">Span ID: </span>
                        <span className="text-[#e8e8f0] font-mono">{span.spanId}</span>
                      </div>
                      {span.parentSpanId && (
                        <div>
                          <span className="text-[#6b7280]">Parent: </span>
                          <span className="text-[#e8e8f0] font-mono">{span.parentSpanId.slice(0, 8)}</span>
                        </div>
                      )}
                      <div>
                        <span className="text-[#6b7280]">Start: </span>
                        <span className="text-[#e8e8f0] font-mono">{formatTimestamp(span.startTime)}</span>
                      </div>
                      {span.endTime && (
                        <div>
                          <span className="text-[#6b7280]">End: </span>
                          <span className="text-[#e8e8f0] font-mono">{formatTimestamp(span.endTime)}</span>
                        </div>
                      )}
                    </div>

                    {/* Tags */}
                    {span.tags && Object.keys(span.tags).length > 0 && (
                      <div className="mt-2 pt-2 border-t border-[rgba(255,255,255,0.06)]">
                        <div className="text-[10px] text-[#6b7280] mb-1">Tags</div>
                        <div className="flex flex-wrap gap-1">
                          {Object.entries(span.tags).map(([key, value]) => (
                            <span
                              key={key}
                              className="px-1.5 py-0.5 rounded text-[9px] bg-[#1a1a28] text-[#9ca3af]"
                            >
                              {key}={value}
                            </span>
                          ))}
                        </div>
                      </div>
                    )}

                    {/* Logs */}
                    {span.logs && span.logs.length > 0 && (
                      <div className="mt-2 pt-2 border-t border-[rgba(255,255,255,0.06)]">
                        <div className="text-[10px] text-[#6b7280] mb-1">Logs</div>
                        <div className="space-y-0.5">
                          {span.logs.map((log, i) => (
                            <div key={i} className="text-[10px] font-mono text-[#6b7280]">
                              <span className="text-[#4b5563]">[{formatTimestamp(log.timestamp)}]</span>{" "}
                              {log.message}
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
