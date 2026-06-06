import { useStudioAgents } from "@/hooks/useStudio";
import type { AgentData } from "@functionfly/ui-agent";
import { AgentDependencyMap, AgentSkillGraph } from "@functionfly/ui-agent";
import { Link2, Sparkles, Wand2, Zap } from "lucide-react";

export function SkillsPanel() {
  const { agents: rawAgents } = useStudioAgents();

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
      <div className="border-b border-border-subtle pb-3">
        <h3 className="text-sm font-medium mb-1">Skills & Dependencies</h3>
        <p className="text-xs text-text-muted">View agent capabilities and relationships</p>
      </div>

      <div className="grid grid-cols-2 gap-2 mb-4">
        <div className="bg-bg-primary rounded-lg border border-border-subtle p-3 text-center">
          <Sparkles className="size-5 text-warning mx-auto mb-1" />
          <div className="text-lg font-semibold">
            {agents.reduce((acc, a) => acc + (a.tools?.length || 0), 0)}
          </div>
          <div className="text-[10px] text-text-muted">Total Tools</div>
        </div>
        <div className="bg-bg-primary rounded-lg border border-border-subtle p-3 text-center">
          <Zap className="size-5 text-brand-400 mx-auto mb-1" />
          <div className="text-lg font-semibold">{agents.length}</div>
          <div className="text-[10px] text-text-muted">Skills</div>
        </div>
      </div>

      <div>
        <div className="flex items-center gap-2 mb-2">
          <Wand2 className="size-4 text-brand-400" />
          <span className="text-xs font-medium">Skill Graph</span>
        </div>
        <AgentSkillGraph agents={agents} onSkillClick={() => {}} />
      </div>

      <div className="border-t border-border-subtle pt-4">
        <div className="flex items-center gap-2 mb-2">
          <Link2 className="size-4 text-success" />
          <span className="text-xs font-medium">Dependency Map</span>
        </div>
        <AgentDependencyMap agents={agents} onDependencyClick={() => {}} />
      </div>
    </div>
  );
}