import { useReassignAgentRole, useReshapeSwarm } from "@/hooks/useAgentSwarm";
import { useStudioAgents } from "@/hooks/useStudio";
import type { AgentData } from "@functionfly/ui-agent";
import { SwarmCoordinator } from "@functionfly/ui-agent";
import { GitBranch, Network, Users } from "lucide-react";

export function SwarmPanel() {
  const { agents: rawAgents } = useStudioAgents();
  const reassignRole = useReassignAgentRole();
  const reshapeSwarm = useReshapeSwarm();

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

  return (
    <div className="p-3 space-y-4">
      <div className="border-b border-border-subtle pb-3">
        <h3 className="text-sm font-medium mb-1">Swarm Coordination</h3>
        <p className="text-xs text-text-muted">Coordinate multiple agents working together</p>
      </div>

      <div className="grid grid-cols-2 gap-2 mb-4">
        <div className="bg-bg-primary rounded-lg border border-border-subtle p-3 text-center">
          <Users className="size-5 text-brand-400 mx-auto mb-1" />
          <div className="text-lg font-semibold">{agents.length}</div>
          <div className="text-[10px] text-text-muted">Active Agents</div>
        </div>
        <div className="bg-bg-primary rounded-lg border border-border-subtle p-3 text-center">
          <Network className="size-5 text-success mx-auto mb-1" />
          <div className="text-lg font-semibold">Mesh</div>
          <div className="text-[10px] text-text-muted">Topology</div>
        </div>
      </div>

      <SwarmCoordinator
        agents={agents}
        onReassign={(agentId, role) => reassignRole.mutate({ agentId, swarmRole: role })}
        onReshape={(topology) => reshapeSwarm.mutate({ agentId: agents[0]?.id || "", topology })}
        className="mb-4"
      />

      <div className="border-t border-border-subtle pt-4">
        <h4 className="text-xs font-medium mb-2">Topology Views</h4>
        <div className="grid grid-cols-3 gap-2">
          {["mesh", "star", "chain", "ring", "tree", "custom"].map((topology) => (
            <button
              key={topology}
              className="px-3 py-2 text-xs bg-bg-primary border border-border-subtle rounded-lg hover:border-brand-500/50 hover:text-brand-400 transition-colors capitalize flex items-center justify-center gap-1"
            >
              <GitBranch className="size-3" />
              {topology}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}