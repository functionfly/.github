/**
 * @functionfly/ui-agent
 * Agent dependency and skill graph components
 */

import * as React from "react";
import { cn } from "@functionfly/ui-core";
import { Badge } from "@functionfly/ui-core";

interface SkillNode {
  id: string;
  name: string;
  category: "language" | "tool" | "runtime" | "domain" | "framework";
  level: "beginner" | "intermediate" | "advanced" | "expert";
  dependencies: string[];
  agentsUsing?: string[];
}

interface AgentSkillGraphProps {
  agents: Array<{
    id: string;
    name: string;
    skills: string[];
  }>;
  onSkillClick?: (skillId: string) => void;
  className?: string;
}

interface DependencyEdge {
  from: string;
  to: string;
  type: "requires" | "enhances" | "conflicts";
}

interface AgentDependencyMapProps {
  agents: Array<{
    id: string;
    name: string;
    skills: string[];
    dependencies: Array<{ targetId: string; type: "requires" | "enhances" | "conflicts" }>;
  }>;
  selectedAgentId?: string;
  onAgentClick?: (agentId: string) => void;
  onDependencyClick?: (dep: { nodeId: string; type: string }) => void;
  className?: string;
}

const categoryColors: Record<string, string> = {
  language: "#3b82f6",
  tool: "#10b981",
  runtime: "#f97316",
  domain: "#8b5cf6",
  framework: "#ec4899",
};

const levelBars: Record<string, { width: string; color: string }> = {
  beginner: { width: "25%", color: "#6b7280" },
  intermediate: { width: "50%", color: "#3b82f6" },
  advanced: { width: "75%", color: "#10b981" },
  expert: { width: "100%", color: "#f97316" },
};

export function AgentSkillGraph({ agents, onSkillClick, className }: AgentSkillGraphProps) {
  const [hoveredSkill, setHoveredSkill] = React.useState<string | null>(null);

  const allSkills = React.useMemo(() => {
    const skillMap = new Map<string, { id: string; name: string; agentsUsing: string[] }>();
    for (const agent of agents) {
      for (const skill of agent.skills) {
        if (!skillMap.has(skill)) {
          skillMap.set(skill, { id: skill, name: skill, agentsUsing: [] });
        }
        skillMap.get(skill)!.agentsUsing.push(agent.name);
      }
    }
    return Array.from(skillMap.values());
  }, [agents]);

  return (
    <div className={cn("space-y-4", className)}>
      {/* Legend */}
      <div className="flex flex-wrap gap-3 text-[10px]">
        {Object.entries(categoryColors).map(([cat, color]) => (
          <div key={cat} className="flex items-center gap-1">
            <div className="size-2 rounded-full" style={{ backgroundColor: color }} />
            <span className="capitalize text-text-muted">{cat}</span>
          </div>
        ))}
      </div>

      {/* Skill nodes */}
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-2">
        {allSkills.map((skill) => (
          <div
            key={skill.id}
            className={cn(
              "p-3 bg-bg-secondary border border-border-subtle rounded-lg cursor-pointer transition-all",
              "hover:border-border-default hover:shadow-md",
              hoveredSkill === skill.id && "ring-2 ring-brand-500/30"
            )}
            onClick={() => onSkillClick?.(skill.id)}
            onMouseEnter={() => setHoveredSkill(skill.id)}
            onMouseLeave={() => setHoveredSkill(null)}
          >
            <div className="flex items-center gap-2 mb-1.5">
              <div
                className="size-2 rounded-full"
                style={{ backgroundColor: "#3b82f6" }}
              />
              <span className="text-xs font-medium text-text-primary truncate">{skill.name}</span>
            </div>

            {/* Agents using */}
            {skill.agentsUsing && skill.agentsUsing.length > 0 && (
              <div className="flex items-center gap-1 mt-1.5">
                <span className="text-[9px] text-text-muted">{skill.agentsUsing.length} agents</span>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}

export function AgentDependencyMap({
  agents,
  selectedAgentId,
  onAgentClick,
  onDependencyClick,
  className,
}: AgentDependencyMapProps) {
  const [internalSelectedAgent, setInternalSelectedAgent] = React.useState<string | null>(null);
  const activeSelectedAgent = selectedAgentId ?? internalSelectedAgent;

  // Calculate positions in a circular layout
  const positions = React.useMemo(() => {
    const pos: Record<string, { x: number; y: number }> = {};
    const center = { x: 150, y: 150 };
    const radius = 120;

    agents.forEach((agent, i) => {
      const angle = (2 * Math.PI * i) / agents.length - Math.PI / 2;
      pos[agent.id] = {
        x: center.x + radius * Math.cos(angle),
        y: center.y + radius * Math.sin(angle),
      };
    });
    return pos;
  }, [agents]);

  const edgeTypeStyles = {
    requires: { stroke: "#ef4444", dasharray: "0" },
    enhances: { stroke: "#10b981", dasharray: "0" },
    conflicts: { stroke: "#f59e0b", dasharray: "5,5" },
  };

  return (
    <div className={cn("relative", className)}>
      <svg width="300" height="300" className="w-full h-auto">
        {/* Edges */}
        <g>
          {agents.map((agent) =>
            agent.dependencies.map((dep) => {
              const from = positions[agent.id];
              const to = positions[dep.targetId];
              if (!from || !to) return null;

              const style = edgeTypeStyles[dep.type as keyof typeof edgeTypeStyles];
              const midX = (from.x + to.x) / 2;
              const midY = (from.y + to.y) / 2;

              return (
                <g key={`${agent.id}-${dep.targetId}`} onClick={() => onDependencyClick?.({ nodeId: dep.targetId, type: dep.type })}>
                  <path
                    d={`M ${from.x} ${from.y} Q ${midX - 20} ${midY} ${to.x} ${to.y}`}
                    fill="none"
                    stroke={style.stroke}
                    strokeWidth={1.5}
                    strokeDasharray={style.dasharray}
                    opacity={0.6}
                  />
                  {/* Arrow head */}
                  <circle cx={to.x} cy={to.y} r={3} fill={style.stroke} />
                </g>
              );
            })
          )}
        </g>

        {/* Agent nodes */}
        {agents.map((agent) => {
          const pos = positions[agent.id];
          if (!pos) return null;

          return (
            <g
              key={agent.id}
              className="cursor-pointer"
              onClick={() => {
                setInternalSelectedAgent(agent.id);
                onAgentClick?.(agent.id);
              }}
            >
              <circle
                cx={pos.x}
                cy={pos.y}
                r={activeSelectedAgent === agent.id ? 28 : 24}
                fill="var(--bg-secondary)"
                stroke={activeSelectedAgent === agent.id ? "#f97316" : "var(--border-subtle)"}
                strokeWidth={activeSelectedAgent === agent.id ? 2 : 1}
                className="transition-all"
              />
              <text
                x={pos.x}
                y={pos.y + 4}
                textAnchor="middle"
                className="text-[10px] fill-text-primary pointer-events-none"
              >
                {agent.name.slice(0, 8)}
              </text>
            </g>
          );
        })}
      </svg>

      {/* Selected agent details */}
      {activeSelectedAgent && (
        <div className="absolute bottom-0 left-0 right-0 p-3 bg-bg-secondary border-t border-border-subtle rounded-b-lg">
          {(() => {
            const agent = agents.find((a) => a.id === activeSelectedAgent);
            if (!agent) return null;
            return (
              <div className="space-y-1">
                <div className="text-xs font-medium text-text-primary">{agent.name}</div>
                <div className="flex flex-wrap gap-1">
                  {agent.skills.map((skill) => (
                    <span key={skill} className="px-1.5 py-0.5 text-[10px] bg-bg-tertiary text-text-muted rounded">
                      {skill}
                    </span>
                  ))}
                </div>
                <div className="text-[10px] text-text-muted">
                  {agent.dependencies.length} dependencies
                </div>
              </div>
            );
          })()}
        </div>
      )}

      {/* Legend */}
      <div className="flex justify-center gap-4 mt-2">
        {Object.entries(edgeTypeStyles).map(([type, style]) => (
          <div key={type} className="flex items-center gap-1.5 text-[10px] text-text-muted">
            <svg width="20" height="10">
              <line
                x1="0"
                y1="5"
                x2="20"
                y2="5"
                stroke={style.stroke}
                strokeWidth={1.5}
                strokeDasharray={style.dasharray}
              />
            </svg>
            <span className="capitalize">{type}</span>
          </div>
        ))}
      </div>
    </div>
  );
}