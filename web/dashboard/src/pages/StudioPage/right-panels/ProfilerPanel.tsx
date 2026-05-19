import React from "react";
import { ExecutionProfiler } from "@functionfly/ui-observability";
import { BarChart2, Clock, DollarSign, AlertCircle, CheckCircle, XCircle, Timer } from "lucide-react";

interface Execution {
  id: string;
  graphId?: string;
  nodeResults?: Array<{
    nodeId: string;
    durationMs?: number;
    status: string;
    error?: string;
  }>;
  startedAt: string;
  status: "completed" | "failed" | "timeout";
}

interface ProfilerPanelProps {
  executions: Execution[];
}

export function ProfilerPanel({ executions }: ProfilerPanelProps) {
  const totalDuration = executions.reduce(
    (acc, ex) => acc + (ex.nodeResults?.[0]?.durationMs || 0),
    0
  );
  const successCount = executions.filter((ex) => ex.status === "completed").length;
  const failCount = executions.filter((ex) => ex.status === "failed").length;
  const timeoutCount = executions.filter((ex) => ex.status === "timeout").length;

  return (
    <div className="p-3 space-y-4">
      <div className="grid grid-cols-2 gap-2">
        <div className="bg-bg-primary rounded-lg border border-border-subtle p-3 text-center">
          <CheckCircle className="size-4 text-success mx-auto mb-1" />
          <div className="text-lg font-semibold">{successCount}</div>
          <div className="text-[10px] text-text-muted">Successful</div>
        </div>
        <div className="bg-bg-primary rounded-lg border border-border-subtle p-3 text-center">
          <XCircle className="size-4 text-error mx-auto mb-1" />
          <div className="text-lg font-semibold">{failCount}</div>
          <div className="text-[10px] text-text-muted">Failed</div>
        </div>
        <div className="bg-bg-primary rounded-lg border border-border-subtle p-3 text-center">
          <Timer className="size-4 text-warning mx-auto mb-1" />
          <div className="text-lg font-semibold">{timeoutCount}</div>
          <div className="text-[10px] text-text-muted">Timeouts</div>
        </div>
        <div className="bg-bg-primary rounded-lg border border-border-subtle p-3 text-center">
          <Clock className="size-4 text-brand-400 mx-auto mb-1" />
          <div className="text-lg font-semibold">{Math.round(totalDuration)}ms</div>
          <div className="text-[10px] text-text-muted">Total Time</div>
        </div>
      </div>

      <div className="border-t border-border-subtle pt-4">
        <ExecutionProfiler
          executions={executions.map((ex, i) => ({
            id: ex.id || `ex-${i}`,
            name: ex.graphId || "Workflow",
            duration: ex.nodeResults?.[0]?.durationMs || 0,
            tokens: 0,
            cost: 0,
            timestamp: ex.startedAt,
            status:
              ex.status === "completed"
                ? "success"
                : ex.status === "failed"
                  ? "error"
                  : "timeout",
            retries: 0,
          }))}
        />
      </div>
    </div>
  );
}