import React from "react";
import { ExecutionTimeline } from "@functionfly/ui-graph";
import { Play } from "lucide-react";

interface ExecutionPanelProps {
  events: Array<{
    id: string;
    type: string;
    nodeLabel: string;
    result: "success" | "failure" | "partial";
    timestamp: number;
    duration: number;
  }>;
  onTimeChange?: (time: number) => void;
  onRunFirstWorkflow?: () => void;
}

export function ExecutionPanel({
  events,
  onTimeChange,
  onRunFirstWorkflow,
}: ExecutionPanelProps) {
  if (events.length === 0) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center">
          <div className="w-16 h-16 rounded-full bg-bg-primary flex items-center justify-center mx-auto mb-3">
            <Play className="size-6 text-text-muted" />
          </div>
          <p className="text-sm text-text-muted mb-3">No execution events recorded yet</p>
          <button
            onClick={onRunFirstWorkflow}
            className="px-4 py-2 text-xs bg-brand-500 text-white rounded-lg hover:bg-brand-600 transition-colors"
          >
            Run First Workflow
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 min-h-0">
      <ExecutionTimeline
        events={events}
        onTimeChange={onTimeChange}
        className="flex-1 min-h-0"
      />
    </div>
  );
}