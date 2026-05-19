import React from "react";
import {
  GhostModeOrchestrator,
  AgentConversationTimeline,
  MultiAgentConversationView,
} from "@functionfly/ui-ghost";
import { useStudioAgents } from "@/hooks/useStudio";
import { Activity, Plus, X } from "lucide-react";

interface GhostBuild {
  id: string;
  goal: string;
  phase: "planning" | "building" | "complete" | "error";
  progress: number;
  startedAt: string;
  updatedAt: string;
}

interface GhostTask {
  id: string;
  title: string;
  description?: string;
  status: string;
  logs?: Array<{ message: string }>;
  updatedAt: string;
  createdAt: string;
}

interface GhostPanelProps {
  build?: GhostBuild;
  tasks: GhostTask[];
  onCancelBuild: () => void;
  onCreateBuild: () => void;
}

export function GhostPanel({ build, tasks, onCancelBuild, onCreateBuild }: GhostPanelProps) {
  const { agents: rawAgents } = useStudioAgents();

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

  return (
    <div className="p-3 space-y-4">
      {build ? (
        <>
          <GhostModeOrchestrator
            build={build}
            onCancel={onCancelBuild}
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
                timestamp: task.updatedAt,
              }))}
              decisions={[]}
            />
          </div>

          <div className="border-t border-border-subtle pt-4">
            <h4 className="text-xs font-medium mb-2">Multi-Agent Conversations</h4>
            <MultiAgentConversationView
              conversations={agents.slice(0, 3).map((agent, i) => ({
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
                  timestamp: task.updatedAt,
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