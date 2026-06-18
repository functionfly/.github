import { useStudioAgents, useStudioMemory } from "@/hooks/useStudio";
import type { AgentMemory } from "@/types";
import {
  AgentDock,
  AgentLifecyclePanel,
  AgentMemoryViewer,
  AgentPermissionEditor,
  AgentRuntimeInspector,
  type AgentData,
} from "@functionfly/ui-agent";
import { AgentsPanelSkeleton } from "../components/StudioPanelsSkeleton";

interface AgentsPanelProps {
  selectedAgentId: string | null;
  onAgentSelect: (agentId: string) => void;
  onAgentCreate: () => void;
  onAgentTerminate: (agentId: string) => void;
  onAgentPause: (agentId: string) => void;
  onAgentResume: (agentId: string) => void;
  onAgentRestart: (agentId: string) => void;
}

export function AgentsPanel({
  selectedAgentId,
  onAgentSelect,
  onAgentCreate,
  onAgentTerminate,
  onAgentPause,
  onAgentResume,
  onAgentRestart,
}: AgentsPanelProps) {
  const { agents: rawAgents, isLoading: isLoadingAgents } = useStudioAgents();
  const agentMemories = useStudioMemory(selectedAgentId || undefined);

  const agents: AgentData[] = rawAgents.map((agent) => ({
    id: agent.id,
    name: agent.name,
    role: agent.agentId,
    status:
      agent.status === "active"
        ? "running"
        : agent.status === "pending"
          ? "idle"
          : agent.status === "terminating" || agent.status === "terminated"
            ? "terminated"
            : (agent.status as "running" | "idle" | "paused" | "terminated" | "error"),
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
    tags: [] as string[],
  }));

  const selectedAgent = agents.find((a) => a.id === selectedAgentId) || null;

  if (isLoadingAgents) {
    return <AgentsPanelSkeleton />;
  }

  return (
    <div className="flex-1 overflow-y-auto">
      <AgentDock
        agents={agents}
        selectedAgentId={selectedAgentId || undefined}
        onAgentSelect={onAgentSelect}
        onAgentCreate={onAgentCreate}
        onAgentTerminate={onAgentTerminate}
        isLoading={isLoadingAgents}
      />

      {selectedAgent && (
        <div className="px-3 pb-3">
          <AgentLifecyclePanel
            agent={selectedAgent}
            onSpawn={onAgentCreate}
            onPause={() => onAgentPause(selectedAgent.id)}
            onResume={() => onAgentResume(selectedAgent.id)}
            onTerminate={() => onAgentTerminate(selectedAgent.id)}
            onRestart={() => onAgentRestart(selectedAgent.id)}
          />
        </div>
      )}

      {selectedAgent && (
        <div className="px-3 pb-3">
          <AgentMemoryViewer
            agentId={selectedAgent.id}
            memories={(agentMemories.memories || []).map((m) => ({
              id: m.id,
              type: m.memory_type as "working" | "longterm" | "context" | "episodic",
              content: m.content || "",
              importance: m.importance_score,
              lastAccessed: m.last_accessed_at || new Date().toISOString(),
              createdAt: m.created_at,
            }))}
            onMemoryAdd={() => {
              // Placeholder - component doesn't provide content/type on click
            }}
            onMemorySearch={(q) => agentMemories.searchMemories.mutate(q)}
            onMemoryDelete={(id) => agentMemories.deleteMemory.mutate(id)}
          />
        </div>
      )}

      {selectedAgent && selectedAgent.permissions && selectedAgent.permissions.length > 0 && (
        <div className="px-3 pb-3">
          <AgentPermissionEditor agentId={selectedAgent.id} permissions={selectedAgent.permissions} />
        </div>
      )}

      {selectedAgent && (
        <div className="px-3 pb-3">
          <AgentRuntimeInspector
            agentId={selectedAgent.id}
            sandboxState={{
              status: "idle",
              memoryUsedMB: Math.round(selectedAgent.memoryUsage / 1024 / 1024),
              memoryLimitMB: Math.round(selectedAgent.memoryLimit / 1024 / 1024),
              cpuCores: 1,
              networkEnabled: false,
              sandboxEnabled: true,
              environmentVars: 0,
              blockedModules: [],
              coldStartMs: 450,
              executionMs: selectedAgent.avgLatency,
              successRate:
                selectedAgent.tasksCompleted + selectedAgent.tasksFailed > 0
                  ? Math.round(
                      (selectedAgent.tasksCompleted /
                        (selectedAgent.tasksCompleted + selectedAgent.tasksFailed)) *
                        100
                    )
                  : 0,
            }}
          />
        </div>
      )}
    </div>
  );
}