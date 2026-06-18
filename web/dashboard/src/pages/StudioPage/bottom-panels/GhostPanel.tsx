import { useStudioAgents, useStudioGhost } from "@/hooks/useStudio";
import {
  AgentConversationTimeline,
  GhostModeOrchestrator,
  MultiAgentConversationView,
} from "@functionfly/ui-ghost";
import { Activity, Plus } from "lucide-react";
import { useCallback } from "react";

interface GhostBuild {
  id: string;
  goal?: string;
  description?: string;
  phase: "planning" | "provisioning" | "building" | "deploying" | "monitoring" | "complete" | "error" | "paused";
  progress: number;
  started_at?: string;
  updated_at?: string;
  current_task_id?: string;
  human_approval_required?: boolean;
  approval_type?: "schema" | "deployment" | "pr" | "infra";
  tasks?: GhostTask[];
}

interface GhostTask {
  id: string;
  title: string;
  description?: string;
  status: "pending" | "in_progress" | "awaiting_approval" | "approved" | "rejected" | "completed" | "failed";
  phase?: string;
  logs?: GhostLogEntry[];
  started_at?: string;
  createdAt?: string;
  updated_at?: string;
  updatedAt?: string;
  completed_at?: string;
}

interface GhostLogEntry {
  timestamp: string;
  level: "info" | "warn" | "error" | "debug";
  message: string;
  metadata?: Record<string, unknown>;
}

interface GhostPanelProps {
  build?: GhostBuild;
  tasks: GhostTask[];
  onCancelBuild: () => void;
  onCreateBuild: () => void;
}

export function GhostPanel({ build, tasks, onCancelBuild, onCreateBuild }: GhostPanelProps) {
  const { agents: rawAgents } = useStudioAgents();
  const { cancelBuild, approveTask, rejectTask, taskLogs } = useStudioGhost();

  const agents = rawAgents.map((agent) => ({
    id: agent.id,
    name: agent.name,
    role: agent.agentId,
    status:
      agent.status === "active"
        ? "running"
        : agent.status === "pending"
          ? "idle"
          : agent.status === "terminating" || agent.status === "terminated"
            ? "stopped"
            : (agent.status as "running" | "idle" | "paused" | "stopped" | "error"),
    memoryUsage: 0,
    memoryLimit: 512 * 1024 * 1024,
    executionBudget: 10.0,
    executionBudgetUsed: 0,
    permissions: [],
    tools: [],
    runtime: "wasm",
    model: "gpt-4o",
    uptime: 0,
    tasksCompleted: 0,
    tasksFailed: 0,
    avgLatency: 0,
    lastHeartbeat: agent.lastActivity || new Date().toISOString(),
    createdAt: new Date().toISOString(),
    description: `Agent ${agent.name}`,
    tags: [],
  }));

  const handleCancel = useCallback(() => {
    if (build?.id) {
      cancelBuild.mutate(build.id);
    }
    onCancelBuild();
  }, [build?.id, cancelBuild, onCancelBuild]);

  const handleApproval = useCallback((type: "approve" | "reject", notes?: string) => {
    if (!build?.id || !build.current_task_id) return;

    if (type === "approve") {
      approveTask.mutate({ buildId: build.id, taskId: build.current_task_id });
    } else {
      rejectTask.mutate({ buildId: build.id, taskId: build.current_task_id, reason: notes });
    }
  }, [build?.id, build?.current_task_id, approveTask, rejectTask]);

  return (
    <div className="p-3 space-y-4">
      {build ? (
        <>
          <GhostModeOrchestrator
            build={{
              id: build.id,
              goal: build.goal || "",
              description: build.description || "",
              phase: build.phase,
              progress: build.progress,
              tasks: (build.tasks || []) as any,
              startedAt: build.started_at || new Date().toISOString(),
              updatedAt: build.updated_at || new Date().toISOString(),
            } as any}
            onCancel={handleCancel}
            onApprove={(approvalType) => handleApproval(approvalType as "approve" | "reject")}
          />

          <div className="border-t border-border-subtle pt-4">
            <h4 className="text-xs font-medium mb-2 flex items-center gap-2">
              <Activity className="size-3 text-brand-400" />
              Conversation Timeline
            </h4>
            <AgentConversationTimeline
              messages={tasks.slice(0, 5).map((task) => ({
                id: task.id,
                agentId: task.id,
                agentName: task.title,
                agentRole: task.description || "Ghost Task",
                type: task.status === "in_progress" ? "action" : "thought",
                content: task.logs?.map((l) => l.message).join("\n") || task.title,
                timestamp: task.updated_at,
              }))}
              decisions={tasks
                .filter((t) => t.status === "awaiting_approval")
                .map((t) => ({
                  id: t.id,
                  timestamp: t.updated_at || new Date().toISOString(),
                  agentId: t.id,
                  agentName: t.title,
                  decision: t.description || "pending decision",
                  rationale: t.description || "awaiting approval",
                  alternatives: [],
                  chosen: "pending",
                  confidence: 0.5,
                })) as any}
            />
          </div>

          <div className="border-t border-border-subtle pt-4">
            <h4 className="text-xs font-medium mb-2">Multi-Agent Conversations</h4>
            <MultiAgentConversationView
              conversations={agents.slice(0, 3).map((agent) => ({
                agentId: agent.id,
                agentName: agent.name,
                agentRole: agent.role || "Agent",
                messages: tasks.slice(0, 3).map((task) => ({
                  id: task.id,
                  agentId: agent.id,
                  agentName: agent.name,
                  agentRole: task.description || "Task",
                  type: "thought" as const,
                  content: task.logs?.map((l) => l.message).join("\n") || task.title,
                  timestamp: task.updated_at,
                })),
              }))}
            />
          </div>
        </>
      ) : (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <div className="w-16 h-16 rounded-full bg-warning/10 flex items-center justify-center mb-4">
            <Activity className="size-6 text-warning" />
          </div>
          <h3 className="text-sm font-medium mb-2">Ghost Mode Inactive</h3>
          <p className="text-xs text-text-muted mb-4 max-w-[240px]">
            Enable autonomous building mode to let AI agents construct and iterate on your workflows
          </p>
          <button
            onClick={onCreateBuild}
            className="px-4 py-2 text-xs bg-warning text-white rounded-lg hover:bg-warning/90 transition-colors flex items-center gap-2"
          >
            <Plus className="size-3" />
            Start Ghost Mode
          </button>
        </div>
      )}
    </div>
  );
}