/**
 * @functionfly/ui-ai
 * Tool Invocation Feed - Real-time tool call streaming
 */

import * as React from "react";
import { cn } from "@functionfly/ui-core";
import { Badge } from "@functionfly/ui-core";
import { Wrench, ChevronDown, ChevronRight, CheckCircle2, XCircle, Loader2, Clock } from "lucide-react";

export interface ToolInvocation {
  id: string;
  toolName: string;
  args: Record<string, any>;
  result?: any;
  status: "pending" | "running" | "completed" | "failed";
  startTime: number;
  endTime?: number;
  error?: string;
  duration?: number;
}

export interface ToolInvocationFeedProps {
  invocations: ToolInvocation[];
  onInvocationClick?: (invocation: ToolInvocation) => void;
  className?: string;
  autoScroll?: boolean;
}

const statusConfig = {
  pending: { icon: Clock, color: "text-text-muted", bg: "bg-bg-tertiary" },
  running: { icon: Loader2, color: "text-brand-500", bg: "bg-brand-500/10", animate: true },
  completed: { icon: CheckCircle2, color: "text-success", bg: "bg-success/10" },
  failed: { icon: XCircle, color: "text-error", bg: "bg-error/10" },
};

export function ToolInvocationFeed({
  invocations,
  onInvocationClick,
  className,
  autoScroll = true,
}: ToolInvocationFeedProps) {
  const [expandedIds, setExpandedIds] = React.useState<Set<string>>(new Set());
  const scrollRef = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    if (autoScroll && scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [invocations, autoScroll]);

  const toggleExpand = (id: string) => {
    setExpandedIds(prev => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const runningCount = invocations.filter(i => i.status === "running").length;

  return (
    <div className={cn("flex flex-col h-full", className)}>
      {/* Header */}
      <div className="flex items-center gap-2 px-4 py-3 border-b border-border-subtle">
        <Wrench className="size-4 text-brand-500" />
        <span className="text-sm font-medium text-text-primary">Tool Calls</span>
        {runningCount > 0 && (
          <Badge variant="brand" size="sm" pulse>{runningCount} running</Badge>
        )}
        <Badge variant="ghost" size="sm" className="ml-auto">{invocations.length} total</Badge>
      </div>

      {/* Invocation List */}
      <div ref={scrollRef} className="flex-1 overflow-y-auto p-3 space-y-2">
        {invocations.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 text-text-muted">
            <Wrench className="size-12 mb-3 opacity-50" />
            <p className="text-sm">No tool invocations</p>
            <p className="text-xs mt-1">Tool calls will appear here</p>
          </div>
        ) : (
          invocations.map(invocation => {
            const config = statusConfig[invocation.status];
            const isExpanded = expandedIds.has(invocation.id);
            const Icon = config.icon;
            const duration = invocation.duration || (invocation.endTime ? invocation.endTime - invocation.startTime : undefined);

            return (
              <div
                key={invocation.id}
                onClick={() => onInvocationClick?.(invocation)}
                className="rounded-lg border border-border-subtle bg-bg-secondary hover:bg-bg-hover transition-colors cursor-pointer"
              >
                {/* Header Row */}
                <div
                  onClick={() => toggleExpand(invocation.id)}
                  className="flex items-center gap-3 p-3"
                >
                  <button className="text-text-muted">
                    {isExpanded ? <ChevronDown className="size-4" /> : <ChevronRight className="size-4" />}
                  </button>

                  <div className={cn("size-8 rounded-lg flex items-center justify-center shrink-0", config.bg)}>
                    <Icon className={cn("size-4", config.color, config.animate && "animate-spin")} />
                  </div>

                  <div className="flex-1 min-w-0">
                    <span className="text-sm font-mono font-medium text-text-primary">{invocation.toolName}</span>
                    <div className="flex items-center gap-2 mt-0.5">
                      {duration !== undefined && (
                        <span className="text-[10px] text-text-muted">{duration}ms</span>
                      )}
                      <Badge variant="outline" size="sm" className="text-[10px]">
                        {invocation.status}
                      </Badge>
                    </div>
                  </div>

                  <span className="text-[10px] text-text-muted">
                    {new Date(invocation.startTime).toLocaleTimeString()}
                  </span>
                </div>

                {/* Expanded Details */}
                {isExpanded && (
                  <div className="px-3 pb-3 space-y-2">
                    {/* Arguments */}
                    <div className="space-y-1">
                      <label className="text-[10px] font-medium text-text-muted uppercase tracking-wide">
                        Arguments
                      </label>
                      <pre className="p-2 bg-bg-tertiary/50 rounded text-xs font-mono text-text-secondary overflow-auto max-h-32">
                        {JSON.stringify(invocation.args, null, 2)}
                      </pre>
                    </div>

                    {/* Result */}
                    {invocation.result !== undefined && (
                      <div className="space-y-1">
                        <label className="text-[10px] font-medium text-text-muted uppercase tracking-wide">
                          Result
                        </label>
                        <pre className="p-2 bg-bg-tertiary/50 rounded text-xs font-mono text-text-secondary overflow-auto max-h-32">
                          {typeof invocation.result === 'string' 
                            ? invocation.result 
                            : JSON.stringify(invocation.result, null, 2)}
                        </pre>
                      </div>
                    )}

                    {/* Error */}
                    {invocation.error && (
                      <div className="p-2 bg-error/10 rounded border border-error/20">
                        <span className="text-xs text-error font-medium">Error: </span>
                        <span className="text-xs text-error">{invocation.error}</span>
                      </div>
                    )}
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
