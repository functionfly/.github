/**
 * @functionfly/ui-agent
 * Agent operating system components for FunctionFly Studio
 */

import * as React from "react";
import { cn } from "@functionfly/ui-core";
import {
  Cpu,
  MemoryStick,
  HardDrive,
  Shield,
  Zap,
  DollarSign,
  Clock,
  Activity,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Loader2,
  Bot,
  Settings2,
  Gauge,
  BarChart2,
  Lock,
  Unlock,
  Trash2,
  Search,
} from "lucide-react";

// --- Types ---
export interface AgentData {
  id: string;
  name: string;
  role: string;
  status: "idle" | "running" | "paused" | "error" | "terminated" | "spawning";
  memoryUsage: number;
  memoryLimit: number;
  executionBudget: number;
  executionBudgetUsed: number;
  permissions: AgentPermission[];
  tools: AgentTool[];
  runtime: string;
  model: string;
  uptime?: number;
  tasksCompleted: number;
  tasksFailed: number;
  avgLatency: number;
  lastHeartbeat: string;
  createdAt: string;
  description?: string;
  tags?: string[];
  error?: string;
  swarmRole?: string;
  skills?: string[];
  dependencies?: Array<{ nodeId: string; type: "requires" | "enhances" | "conflicts" }>;
}

export interface AgentPermission {
  id: string;
  name: string;
  description: string;
  granted: boolean;
  category: "read" | "write" | "execute" | "admin" | "network" | "storage";
  riskLevel: "low" | "medium" | "high" | "critical";
}

export interface AgentTool {
  id: string;
  name: string;
  description: string;
  type: "api" | "function" | "service" | "resource";
  enabled: boolean;
  lastUsed?: string;
  callCount: number;
}

export interface AgentCardProps {
  agent: AgentData;
  onClick?: (id: string) => void;
  onTerminate?: (id: string) => void;
  onPause?: (id: string) => void;
  onResume?: (id: string) => void;
  className?: string;
}

export interface AgentDockProps {
  agents: AgentData[];
  selectedAgentId?: string;
  onAgentSelect?: (id: string) => void;
  onAgentCreate?: () => void;
  onAgentTerminate?: (id: string) => void;
  isLoading?: boolean;
  className?: string;
}

export interface AgentLifecyclePanelProps {
  agent: AgentData;
  onSpawn: () => void;
  onPause: () => void;
  onResume: () => void;
  onTerminate: () => void;
  onRestart: () => void;
}

export interface AgentMemoryViewerProps {
  memories: Array<{
    id: string;
    type: "working" | "longterm" | "context" | "episodic";
    content: string;
    importance: number;
    lastAccessed: string;
    createdAt: string;
    embedding?: number[];
  }>;
  agentId: string;
  onMemoryAdd?: () => void;
  onMemorySearch?: (query: string) => void;
  className?: string;
}

export interface AgentPermissionEditorProps {
  permissions: AgentPermission[];
  onPermissionToggle?: (id: string, granted: boolean) => void;
  onSave?: () => void;
  className?: string;
}

export interface AgentToolchainEditorProps {
  tools: AgentTool[];
  availableTools: AgentTool[];
  onToolToggle?: (id: string, enabled: boolean) => void;
  onToolAdd?: (toolId: string) => void;
  onToolRemove?: (toolId: string) => void;
  className?: string;
}

export interface AgentBudgetMeterProps {
  budget: number;
  used: number;
  limit: number;
  period: string;
  onBudgetChange?: (amount: number) => void;
}

export interface SwarmAgent extends AgentData {
  communicationPattern?: "broadcast" | "chain" | "fan-out" | "peer";
}

export interface SwarmCoordinatorProps {
  agents: SwarmAgent[];
  selectedAgentId?: string;
  onAgentSelect?: (agentId: string) => void;
  onReassign?: (agentId: string, role: string) => void;
  onReshape?: (shape: "chain" | "star" | "mesh" | "tree") => void;
  className?: string;
}

// --- Utility functions ---
export function getStatusConfig(status: string) {
  return {
    idle: { color: "#6b7280", label: "Idle", icon: <Loader2 className="size-3" /> },
    running: { color: "#10b981", label: "Running", icon: <Activity className="size-3" /> },
    paused: { color: "#f59e0b", label: "Paused", icon: <Clock className="size-3" /> },
    error: { color: "#ef4444", label: "Error", icon: <AlertTriangle className="size-3" /> },
    terminated: { color: "#6b7280", label: "Terminated", icon: <XCircle className="size-3" /> },
    spawning: { color: "#3b82f6", label: "Spawning", icon: <Loader2 className="size-3 animate-spin" /> },
  }[status] || { color: "#6b7280", label: "Unknown", icon: <Loader2 className="size-3" /> };
}

export function getRiskColor(risk: string) {
  return {
    low: "#10b981",
    medium: "#f59e0b",
    high: "#f97316",
    critical: "#ef4444",
  }[risk] || "#6b7280";
}

// --- AgentCard ---
export function AgentCard({
  agent,
  onClick,
  onTerminate,
  onPause,
  onResume,
  className,
}: AgentCardProps) {
  const statusConfig = getStatusConfig(agent.status);
  const memoryPercent = (agent.memoryUsage / agent.memoryLimit) * 100;
  const budgetPercent = (agent.executionBudgetUsed / agent.executionBudget) * 100;
  const successRate = agent.tasksCompleted > 0
    ? ((agent.tasksCompleted / (agent.tasksCompleted + agent.tasksFailed)) * 100).toFixed(1)
    : "—";

  return (
    <div
      className={cn(
        "group relative bg-bg-primary border border-border-subtle rounded-xl p-4 transition-all duration-200 overflow-hidden",
        "hover:border-border-default hover:shadow-lg hover:shadow-black/20",
        agent.status === "running" && "border-brand-500/30 shadow-brand-500/5",
        className
      )}
      style={{
        boxShadow: agent.status === "running" ? `0 0 12px ${statusConfig.color}15` : undefined,
      }}
    >
      {/* Status indicator bar */}
      <div
        className="absolute top-0 left-0 right-0 h-[2px]"
        style={{ backgroundColor: statusConfig.color }}
      />

      {/* Header */}
      <div className="flex items-start gap-3 mb-3">
        <div
          className={cn(
            "size-10 rounded-lg flex items-center justify-center text-sm font-bold",
            "bg-gradient-to-br from-brand-500/20 to-brand-500/5 border border-brand-500/20 text-brand-400"
          )}
        >
          <Bot className="size-5" />
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <h3 className="text-sm font-semibold text-text-primary truncate">{agent.name}</h3>
            <Badge
              variant="ghost"
              size="sm"
              style={{ color: statusConfig.color, borderColor: statusConfig.color + "40" }}
            >
              {statusConfig.icon}
              <span>{statusConfig.label}</span>
            </Badge>
          </div>
          <p className="text-xs text-text-muted truncate mt-0.5">{agent.role}</p>
        </div>
        {/* Runtime badge */}
        <span className="text-[10px] font-mono bg-bg-tertiary px-1.5 py-0.5 rounded text-text-muted">
          {agent.runtime}
        </span>
      </div>

      {/* Description */}
      {agent.description && (
        <p className="text-xs text-text-muted mb-3 line-clamp-2">{agent.description}</p>
      )}

      {/* Metrics row */}
      <div className="grid grid-cols-3 gap-2 mb-3">
        <div className="bg-bg-secondary rounded-lg p-2 text-center">
          <div className="text-[10px] text-text-muted mb-0.5">Tasks</div>
          <div className="text-sm font-bold text-text-primary">
            {agent.tasksCompleted}/{agent.tasksCompleted + agent.tasksFailed}
          </div>
          <div className="text-[10px] text-success">{successRate}% success</div>
        </div>
        <div className="bg-bg-secondary rounded-lg p-2 text-center">
          <div className="text-[10px] text-text-muted mb-0.5">Mem</div>
          <div className="text-sm font-bold text-text-primary">
            {formatBytes(agent.memoryUsage)}
          </div>
          <div className="text-[10px] text-text-muted">/ {formatBytes(agent.memoryLimit)}</div>
        </div>
        <div className="bg-bg-secondary rounded-lg p-2 text-center">
          <div className="text-[10px] text-text-muted mb-0.5">Avg Latency</div>
          <div className="text-sm font-bold text-text-primary">
            {agent.avgLatency}ms
          </div>
          <div className="text-[10px] text-text-muted">
            {new Date(agent.lastHeartbeat).toLocaleTimeString()}
          </div>
        </div>
      </div>

      {/* Memory progress bar */}
      <div className="mb-2">
        <div className="flex justify-between text-[10px] text-text-muted mb-0.5">
          <span>Memory</span>
          <span>{memoryPercent.toFixed(1)}%</span>
        </div>
        <div className="h-1.5 bg-bg-tertiary rounded-full overflow-hidden">
          <div
            className="h-full rounded-full transition-all duration-500"
            style={{
              width: `${Math.min(memoryPercent, 100)}%`,
              backgroundColor: memoryPercent > 80 ? "#ef4444" : memoryPercent > 60 ? "#f59e0b" : "#10b981",
            }}
          />
        </div>
      </div>

      {/* Budget progress bar */}
      <div className="mb-3">
        <div className="flex justify-between text-[10px] text-text-muted mb-0.5">
          <span>Budget</span>
          <span>${agent.executionBudgetUsed.toFixed(2)} / ${agent.executionBudget.toFixed(2)}</span>
        </div>
        <div className="h-1.5 bg-bg-tertiary rounded-full overflow-hidden">
          <div
            className="h-full rounded-full transition-all duration-500"
            style={{
              width: `${Math.min(budgetPercent, 100)}%`,
              backgroundColor: budgetPercent > 80 ? "#ef4444" : "#f97316",
            }}
          />
        </div>
      </div>

      {/* Tags */}
      {agent.tags && agent.tags.length > 0 && (
        <div className="flex flex-wrap gap-1 mb-3">
          {agent.tags.map((tag) => (
            <span
              key={tag}
              className="px-1.5 py-0.5 text-[10px] bg-bg-tertiary text-text-muted rounded"
            >
              {tag}
            </span>
          ))}
        </div>
      )}

      {/* Actions */}
      <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
        {agent.status === "running" && (
          <button
            onClick={() => onPause?.(agent.id)}
            className="flex-1 py-1.5 text-[11px] bg-bg-tertiary hover:bg-bg-hover text-text-secondary rounded-lg transition-colors flex items-center justify-center gap-1"
          >
            <Clock className="size-3" /> Pause
          </button>
        )}
        {agent.status === "paused" && (
          <button
            onClick={() => onResume?.(agent.id)}
            className="flex-1 py-1.5 text-[11px] bg-brand-500/20 hover:bg-brand-500/30 text-brand-400 rounded-lg transition-colors flex items-center justify-center gap-1"
          >
            <Activity className="size-3" /> Resume
          </button>
        )}
        {(agent.status === "idle" || agent.status === "error") && (
          <button
            onClick={() => onClick?.(agent.id)}
            className="flex-1 py-1.5 text-[11px] bg-brand-500/20 hover:bg-brand-500/30 text-brand-400 rounded-lg transition-colors flex items-center justify-center gap-1"
          >
            <Zap className="size-3" /> Activate
          </button>
        )}
        {onTerminate && (agent.status === "running" || agent.status === "paused") && (
          <button
            onClick={() => onTerminate(agent.id)}
            className="p-1.5 text-text-muted hover:text-error hover:bg-error/10 rounded-lg transition-colors"
            title="Terminate agent"
          >
            <Trash2 className="size-3.5" />
          </button>
        )}
      </div>
    </div>
  );
}

// --- AgentDock ---
export function AgentDock({
  agents,
  selectedAgentId,
  onAgentSelect,
  onAgentCreate,
  onAgentTerminate,
  isLoading,
  className,
}: AgentDockProps) {
  const statusGroups = {
    running: agents.filter((a) => a.status === "running"),
    spawning: agents.filter((a) => a.status === "spawning"),
    idle: agents.filter((a) => a.status === "idle"),
    paused: agents.filter((a) => a.status === "paused"),
    error: agents.filter((a) => a.status === "error"),
    terminated: agents.filter((a) => a.status === "terminated"),
  };

  return (
    <div className={cn("flex flex-col h-full", className)}>
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-2.5 border-b border-border-subtle bg-bg-secondary gap-2">
        <div className="flex items-center gap-2 min-w-0">
          <Bot className="size-4 text-brand-400 shrink-0" />
          <span className="text-sm font-semibold text-text-primary whitespace-nowrap truncate">Agent Dock</span>
          <Badge variant="outline" size="sm" className="shrink-0 whitespace-nowrap">
            {agents.filter((a) => a.status !== "terminated").length} Active
          </Badge>
          {isLoading && (
            <Loader2 className="size-3.5 text-brand-400 animate-spin shrink-0" />
          )}
        </div>
        {onAgentCreate && (
          <button
            onClick={onAgentCreate}
            className="shrink-0 px-2 py-1 text-[10px] bg-brand-500 hover:bg-brand-600 text-white rounded-md transition-colors flex items-center gap-1 whitespace-nowrap"
          >
            <Zap className="size-3" /> Spawn
          </button>
        )}
      </div>

      {/* Scrollable agent list */}
      <div className="flex-1 overflow-y-auto p-2 space-y-2">
        {Object.entries(statusGroups).map(
          ([status, groupAgents]) =>
            groupAgents.length > 0 && (
              <div key={status}>
                <div className="px-2 py-1.5 text-[10px] font-medium text-text-muted uppercase tracking-wider">
                  {status} ({groupAgents.length})
                </div>
                <div className="space-y-2">
                  {groupAgents.map((agent) => (
                    <div
                      key={agent.id}
                      className={cn(
                        "rounded-lg transition-colors duration-150",
                        selectedAgentId === agent.id
                          ? "ring-2 ring-brand-500/50"
                          : "hover:ring-1 hover:ring-border-subtle"
                      )}
                    >
                      <AgentCard
                        agent={agent}
                        onClick={() => onAgentSelect?.(agent.id)}
                        onTerminate={onAgentTerminate}
                      />
                    </div>
                  ))}
                </div>
              </div>
            )
        )}

        {agents.length === 0 && (
          <div className="flex flex-col items-center justify-center py-12 text-text-muted">
            <Bot className="size-12 mb-3 opacity-30" />
            <p className="text-sm font-medium">No agents yet</p>
            <p className="text-xs mt-1">Spawn your first agent to begin</p>
          </div>
        )}
      </div>

      {/* Summary bar */}
      <div className="flex items-center justify-between px-4 py-2 border-t border-border-subtle bg-bg-tertiary text-[11px] text-text-muted">
        <span>
          Total: {agents.length} agents ·{" "}
          {agents.filter((a) => a.status === "running").length} running
        </span>
      </div>
    </div>
  );
}

// --- Agent Lifecycle Panel ---
export function AgentLifecyclePanel({
  agent,
  onSpawn,
  onPause,
  onResume,
  onTerminate,
  onRestart,
}: AgentLifecyclePanelProps) {
  const statusConfig = getStatusConfig(agent.status);

  const actions = {
    idle: [{ label: "Spawn", icon: <Zap />, action: onSpawn, variant: "brand" as const }],
    spawning: [
      { label: "Cancel", icon: <XCircle />, action: onTerminate, variant: "destructive" as const },
    ],
    running: [
      { label: "Pause", icon: <Clock />, action: onPause, variant: "default" as const },
      { label: "Restart", icon: <Settings2 />, action: onRestart, variant: "default" as const },
      { label: "Terminate", icon: <Trash2 />, action: onTerminate, variant: "destructive" as const },
    ],
    paused: [
      { label: "Resume", icon: <Activity />, action: onResume, variant: "brand" as const },
      { label: "Terminate", icon: <Trash2 />, action: onTerminate, variant: "destructive" as const },
    ],
    error: [{ label: "Restart", icon: <Settings2 />, action: onRestart, variant: "brand" as const }],
  };

  const availableActions = actions[agent.status] || [];

  return (
    <div className="space-y-4">
      {/* Status Display */}
      <div className="flex items-center gap-4 p-4 rounded-xl bg-bg-secondary border border-border-subtle">
        <div
          className="size-12 rounded-xl flex items-center justify-center text-2xl"
          style={{
            background: `linear-gradient(135deg, ${statusConfig.color}20, ${statusConfig.color}08)`,
            color: statusConfig.color,
          }}
        >
          {statusConfig.icon}
        </div>
        <div>
          <h3 className="text-sm font-semibold text-text-primary">{statusConfig.label}</h3>
          <p className="text-xs text-text-muted">Last heartbeat: {agent.lastHeartbeat}</p>
        </div>
      </div>

      {/* Runtime Info */}
      <div className="grid grid-cols-2 gap-2 p-3 bg-bg-secondary rounded-xl border border-border-subtle text-sm">
        <div className="flex items-center gap-2">
          <Cpu className="size-4 text-text-muted" />
          <span className="text-text-muted">Runtime</span>
          <span className="font-medium text-text-primary ml-auto">{agent.runtime}</span>
        </div>
        <div className="flex items-center gap-2">
          <MemoryStick className="size-4 text-text-muted" />
          <span className="text-text-muted">Model</span>
          <span className="font-medium text-text-primary ml-auto">{agent.model}</span>
        </div>
        <div className="flex items-center gap-2">
          <HardDrive className="size-4 text-text-muted" />
          <span className="text-text-muted">Memory</span>
          <span className="font-medium text-text-primary ml-auto">
            {formatBytes(agent.memoryUsage)} / {formatBytes(agent.memoryLimit)}
          </span>
        </div>
        <div className="flex items-center gap-2">
          <DollarSign className="size-4 text-text-muted" />
          <span className="text-text-muted">Budget</span>
          <span className="font-medium text-text-primary ml-auto">
            ${agent.executionBudgetUsed.toFixed(2)} / ${agent.executionBudget.toFixed(2)}
          </span>
        </div>
      </div>

      {/* Actions */}
      <div className="space-y-2">
        {availableActions.map((action) => (
          <button
            key={action.label}
            onClick={action.action}
            className={cn(
              "w-full flex items-center justify-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium transition-all duration-200",
              action.variant === "brand"
                ? "bg-brand-500 hover:bg-brand-600 text-white shadow-md shadow-brand-500/20"
                : action.variant === "destructive"
                ? "bg-error/10 hover:bg-error/20 text-error border border-error/30"
                : "bg-bg-tertiary hover:bg-bg-hover text-text-primary border border-border-subtle"
            )}
          >
            {action.icon}
            {action.label}
          </button>
        ))}
      </div>
    </div>
  );
}

// --- Agent Memory Viewer ---
export function AgentMemoryViewer({
  memories,
  agentId,
  onMemoryAdd,
  onMemorySearch,
  className,
}: AgentMemoryViewerProps) {
  const [searchQuery, setSearchQuery] = React.useState("");

  const typeConfig = {
    working: { color: "#3b82f6", label: "Working" },
    longterm: { color: "#10b981", label: "Long-term" },
    context: { color: "#f59e0b", label: "Context" },
    episodic: { color: "#8b5cf6", label: "Episodic" },
  };

  const filteredMemories = memories.filter((m) =>
    searchQuery
      ? m.content.toLowerCase().includes(searchQuery.toLowerCase())
      : true
  );

  return (
    <div className={cn("space-y-4", className)}>
      {/* Search */}
      <div className="relative">
        <input
          type="text"
          value={searchQuery}
          onChange={(e) => {
            setSearchQuery(e.target.value);
            onMemorySearch?.(e.target.value);
          }}
          placeholder="Search memories..."
          className="w-full px-4 py-2 text-sm bg-bg-primary border border-border-subtle rounded-lg text-text-primary focus:outline-none focus:border-brand-500 transition-colors pl-10"
        />
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-text-muted" />
      </div>

      {/* Memory Stats */}
      <div className="grid grid-cols-4 gap-2">
        {Object.entries(
          memories.reduce(
            (acc, m) => {
              acc[m.type] = (acc[m.type] || 0) + 1;
              return acc;
            },
            {} as Record<string, number>
          )
        ).map(([type, count]) => {
          const config = typeConfig[type as keyof typeof typeConfig];
          return (
            <div
              key={type}
              className="p-2 bg-bg-secondary rounded-lg text-center border border-border-subtle"
            >
              <div
                className="text-xs font-bold mb-0.5"
                style={{ color: config.color }}
              >
                {count}
              </div>
              <div className="text-[9px] text-text-muted capitalize">{config.label}</div>
            </div>
          );
        })}
      </div>

      {/* Memory List */}
      <div className="space-y-2 max-h-96 overflow-y-auto">
        {filteredMemories.map((memory) => {
          const config = typeConfig[memory.type];
          return (
            <div
              key={memory.id}
              className="flex items-start gap-3 p-3 bg-bg-secondary rounded-lg border border-border-subtle hover:border-border-default transition-colors group"
            >
              <div
                className="size-2.5 rounded-full shrink-0 mt-1.5"
                style={{ backgroundColor: config.color }}
              />
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-0.5">
                  <span className="text-[10px] font-medium capitalize" style={{ color: config.color }}>
                    {config.label}
                  </span>
                  <span className="text-[10px] text-text-muted">
                    {new Date(memory.lastAccessed).toLocaleTimeString()}
                  </span>
                </div>
                <p className="text-sm text-text-secondary line-clamp-2">{memory.content}</p>
              </div>
              <Badge variant="ghost" size="sm" className="opacity-0 group-hover:opacity-100">
                {memory.importance}
              </Badge>
            </div>
          );
        })}
      </div>

      {/* Add Memory */}
      {onMemoryAdd && (
        <button
          onClick={onMemoryAdd}
          className="w-full py-2 border-2 border-dashed border-border-subtle rounded-lg text-text-muted hover:text-brand-500 hover:border-brand-500/50 transition-colors flex items-center justify-center gap-2 text-sm"
        >
          <Zap className="size-4" /> Add Memory
        </button>
      )}
    </div>
  );
}

// --- AgentPermissionEditor ---
export function AgentPermissionEditor({
  permissions,
  onPermissionToggle,
  onSave,
  className,
}: AgentPermissionEditorProps) {
  const categoryLabels: Record<string, string> = {
    read: "Read Access",
    write: "Write Access",
    execute: "Execute",
    admin: "Administrative",
    network: "Network",
    storage: "Storage",
  };

  const grouped = permissions.reduce(
    (acc, p) => {
      if (!acc[p.category]) acc[p.category] = [];
      acc[p.category].push(p);
      return acc;
    },
    {} as Record<string, AgentPermission[]>
  );

  return (
    <div className={cn("space-y-4", className)}>
      {Object.entries(grouped).map(([category, perms]) => (
        <div key={category} className="space-y-2">
          <h4 className="text-xs font-semibold text-text-muted uppercase tracking-wider">
            {categoryLabels[category] || category}
          </h4>
          <div className="space-y-1">
            {perms.map((perm) => (
              <div
                key={perm.id}
                className={cn(
                  "flex items-center justify-between p-3 rounded-lg transition-all duration-200",
                  perm.granted
                    ? "bg-brand-500/5 border border-brand-500/20"
                    : "bg-bg-secondary border border-border-subtle"
                )}
              >
                <div className="flex items-center gap-3">
                  <div
                    className="size-2 rounded-full"
                    style={{
                      backgroundColor: perm.granted
                        ? getRiskColor(perm.riskLevel)
                        : "transparent",
                      border: `1px solid ${getRiskColor(perm.riskLevel)}`,
                    }}
                  />
                  <div>
                    <div className="text-sm font-medium text-text-primary">{perm.name}</div>
                    <div className="text-[11px] text-text-muted">{perm.description}</div>
                  </div>
                </div>
                <button
                  onClick={() => onPermissionToggle?.(perm.id, !perm.granted)}
                  className={cn(
                    "p-1.5 rounded-lg transition-colors",
                    perm.granted
                      ? "text-brand-500 hover:bg-brand-500/10"
                      : "text-text-muted hover:bg-bg-hover"
                  )}
                >
                  {perm.granted ? (
                    <Lock className="size-4" />
                  ) : (
                    <Unlock className="size-4" />
                  )}
                </button>
              </div>
            ))}
          </div>
        </div>
      ))}

      {onSave && (
        <button
          onClick={onSave}
          className="w-full mt-4 py-2.5 bg-brand-500 hover:bg-brand-600 text-white font-medium rounded-lg transition-colors text-sm"
        >
          Save Permissions
        </button>
      )}
    </div>
  );
}

// --- AgentToolchainEditor ---
export function AgentToolchainEditor({
  tools,
  availableTools,
  onToolToggle,
  onToolAdd,
  onToolRemove,
  className,
}: AgentToolchainEditorProps) {
  return (
    <div className={cn("space-y-4", className)}>
      {/* Active Tools */}
      <div className="space-y-2">
        <h4 className="text-xs font-semibold text-text-muted uppercase tracking-wider">Active Tools</h4>
        <div className="space-y-1">
          {tools.map((tool) => (
            <div
              key={tool.id}
              className="flex items-center justify-between p-3 bg-bg-secondary rounded-lg border border-border-subtle hover:border-border-default transition-colors"
            >
              <div className="flex items-center gap-3">
                <Badge variant={tool.enabled ? "success" : "ghost"} size="sm" />
                <div>
                  <div className="text-sm font-medium text-text-primary">{tool.name}</div>
                  <div className="text-[11px] text-text-muted">{tool.description}</div>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-[10px] text-text-muted">{tool.callCount} calls</span>
                {onToolToggle && (
                  <button
                    onClick={() => onToolToggle(tool.id, !tool.enabled)}
                    className="p-1 text-text-muted hover:text-brand-500 transition-colors"
                  >
                    {tool.enabled ? <CheckCircle className="size-4" /> : <XCircle className="size-4" />}
                  </button>
                )}
                {onToolRemove && (
                  <button
                    onClick={() => onToolRemove(tool.id)}
                    className="p-1 text-text-muted hover:text-error transition-colors"
                  >
                    <Trash2 className="size-3.5" />
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Available Tools */}
      {availableTools.length > 0 && (
        <div className="space-y-2">
          <h4 className="text-xs font-semibold text-text-muted uppercase tracking-wider">Available Tools</h4>
          <div className="space-y-1">
            {availableTools.map((tool) => (
              <div
                key={tool.id}
                className="flex items-center justify-between p-3 bg-bg-primary rounded-lg border border-border-subtle cursor-pointer hover:border-brand-500/30 transition-colors"
                onClick={() => onToolAdd?.(tool.id)}
              >
                <div>
                  <div className="text-sm font-medium text-text-primary">{tool.name}</div>
                  <div className="text-[11px] text-text-muted">{tool.description}</div>
                </div>
                <Badge variant="brand" size="sm">
                  + Add
                </Badge>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

// --- AgentBudgetMeter ---
export function AgentBudgetMeter({
  budget,
  used,
  limit,
  period,
  onBudgetChange,
}: AgentBudgetMeterProps) {
  const remaining = budget - used;
  const percentUsed = (used / limit) * 100;
  const daysLeft = Math.ceil((limit - used) / (used / 30 || 1)); // rough estimate

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between text-sm">
        <span className="text-text-secondary">Monthly Budget</span>
        <span className="font-mono text-text-primary">
          ${used.toFixed(2)} / ${budget.toFixed(2)}
        </span>
      </div>
      <div className="h-3 bg-bg-tertiary rounded-full overflow-hidden">
        <div
          className="h-full rounded-full transition-all duration-500"
          style={{
            width: `${Math.min(percentUsed, 100)}%`,
            background: `linear-gradient(90deg, #f97316, ${percentUsed > 80 ? "#ef4444" : "#10b981"})`,
          }}
        />
      </div>
      <div className="flex items-center justify-between text-[11px] text-text-muted">
        <span>{remaining > 0 ? `${daysLeft} days estimated` : "Budget exhausted"}</span>
        <span className="font-medium">
          {remaining > 0 ? `$${remaining.toFixed(2)} remaining` : "Over budget"}
        </span>
      </div>
    </div>
  );
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i];
}

function Badge({ variant = "default", size = "md", className, children, ...props }: any) {
  const sizeClasses = { sm: "px-1.5 py-0.5 text-[10px]", md: "px-2 py-0.5 text-[11px]", lg: "px-2.5 py-1 text-xs" };
  const variantClasses = {
    default: "bg-bg-tertiary text-text-secondary",
    brand: "bg-brand-500/20 text-brand-400 border border-brand-500/30",
    success: "bg-success/20 text-success",
    error: "bg-error/20 text-error",
    warning: "bg-warning/20 text-warning",
    outline: "bg-transparent text-text-muted border border-border-subtle",
    ghost: "bg-transparent text-text-secondary",
  };
  return (
    <span className={`inline-flex items-center gap-1 rounded-full font-medium ${sizeClasses[size]} ${variantClasses[variant]} ${className || ""}`} {...props}>
      {children}
    </span>
  );
}

// --- SwarmCoordinator ---
export function SwarmCoordinator({
  agents,
  selectedAgentId,
  onAgentSelect,
  onReassign,
  onReshape,
  className,
}: SwarmCoordinatorProps) {
  const topoShapes = ["chain", "star", "mesh", "tree"] as const;

  return (
    <div className={cn("space-y-4", className)}>
      {/* Topology selector */}
      <div className="flex items-center gap-2">
        <span className="text-sm font-medium text-text-primary">Swarm Topology</span>
        <div className="flex gap-1">
          {topoShapes.map((shape) => (
            <button
              key={shape}
              onClick={() => onReshape?.(shape)}
              className="px-2 py-1 text-[10px] rounded-md bg-bg-secondary border border-border-subtle text-text-muted hover:border-brand-500/30 hover:text-brand-400 transition-all capitalize"
            >
              {shape}
            </button>
          ))}
        </div>
      </div>

      {/* Agents */}
      <div className="space-y-2">
        {agents.map((agent) => (
          <div
            key={agent.id}
            onClick={() => onAgentSelect?.(agent.id)}
            className={cn(
              "flex items-center justify-between p-3 bg-bg-secondary rounded-lg border border-border-subtle cursor-pointer transition-all",
              selectedAgentId === agent.id && "border-brand-500/50 bg-brand-500/5"
            )}
          >
            <div className="flex items-center gap-3">
              <div className="size-8 rounded-lg bg-brand-500/10 border border-brand-500/20 flex items-center justify-center text-brand-400 text-xs font-bold">
                {agent.name[0]}
              </div>
              <div>
                <div className="text-sm font-medium text-text-primary">{agent.name}</div>
                <div className="text-[10px] text-text-muted">
                  Role: <span className="text-text-primary capitalize">{(agent as any).swarmRole || "unassigned"}</span>
                  {agent.dependencies.length > 0 && ` · Depends: ${agent.dependencies.length}`}
                </div>
              </div>
            </div>
            <select
              value={(agent as any).swarmRole || "unassigned"}
              onChange={(e) => {
                e.stopPropagation();
                onReassign?.(agent.id, e.target.value);
              }}
              onClick={(e) => e.stopPropagation()}
              className="text-[10px] bg-bg-primary border border-border-subtle rounded px-1.5 py-0.5 text-text-primary focus:outline-none focus:border-brand-500"
            >
              <option value="worker">Worker</option>
              <option value="manager">Manager</option>
              <option value="infrastructure">Infrastructure</option>
            </select>
          </div>
        ))}
      </div>
    </div>
  );
}


// --- AgentRuntimeInspector ---
export interface AgentRuntimeInspectorProps {
  agentId: string;
  sandboxState?: {
    status: "active" | "idle" | "error" | "unavailable";
    memoryUsedMB: number;
    memoryLimitMB: number;
    cpuCores: number;
    networkEnabled: boolean;
    sandboxEnabled: boolean;
    environmentVars: number;
    blockedModules: string[];
    coldStartMs: number;
    executionMs: number;
    successRate: number;
  };
  onRefresh?: () => void;
  className?: string;
}

export function AgentRuntimeInspector({
  agentId,
  sandboxState,
  onRefresh,
  className,
}: AgentRuntimeInspectorProps) {
  const state = sandboxState || {
    status: "idle" as const,
    memoryUsedMB: 0,
    memoryLimitMB: 512,
    cpuCores: 1,
    networkEnabled: false,
    sandboxEnabled: true,
    environmentVars: 0,
    blockedModules: [],
    coldStartMs: 0,
    executionMs: 0,
    successRate: 0,
  };

  const memoryPercent = (state.memoryUsedMB / state.memoryLimitMB) * 100;
  const statusColors = {
    active: "#10b981",
    idle: "#6b7280",
    error: "#ef4444",
    unavailable: "#f59e0b",
  };

  return (
    <div className={cn("space-y-4", className)}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Shield className="size-4 text-brand-400" />
          <span className="text-sm font-medium text-text-primary">Runtime Inspector</span>
        </div>
        {onRefresh && (
          <button
            onClick={onRefresh}
            className="p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-bg-hover transition-colors"
          >
            <Loader2 className="size-3.5" />
          </button>
        )}
      </div>

      {/* Status Badge */}
      <div className="flex items-center gap-2">
        <div
          className="size-2.5 rounded-full animate-pulse"
          style={{ backgroundColor: statusColors[state.status] }}
        />
        <span className="text-xs text-text-muted capitalize">{state.status}</span>
        {state.sandboxEnabled && (
          <Badge variant="success" size="sm">Sandboxed</Badge>
        )}
      </div>

      {/* Resource Usage */}
      <div className="grid grid-cols-2 gap-3">
        <div className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
          <div className="flex items-center gap-2 mb-2">
            <MemoryStick className="size-3 text-text-muted" />
            <span className="text-[10px] text-text-muted">Memory</span>
          </div>
          <div className="text-lg font-bold text-text-primary">
            {state.memoryUsedMB} <span className="text-xs font-normal text-text-muted">/ {state.memoryLimitMB} MB</span>
          </div>
          <div className="h-1.5 bg-bg-tertiary rounded-full overflow-hidden mt-2">
            <div
              className="h-full rounded-full transition-all"
              style={{
                width: `${Math.min(memoryPercent, 100)}%`,
                backgroundColor: memoryPercent > 80 ? "#ef4444" : memoryPercent > 60 ? "#f59e0b" : "#10b981",
              }}
            />
          </div>
        </div>

        <div className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
          <div className="flex items-center gap-2 mb-2">
            <Cpu className="size-3 text-text-muted" />
            <span className="text-[10px] text-text-muted">CPU Cores</span>
          </div>
          <div className="text-lg font-bold text-text-primary">{state.cpuCores}</div>
        </div>
      </div>

      {/* Network & Security */}
      <div className="p-3 bg-bg-secondary rounded-lg border border-border-subtle space-y-2">
        <div className="flex items-center justify-between">
          <span className="text-xs text-text-muted">Network Access</span>
          <span className={cn("text-xs font-medium", state.networkEnabled ? "text-amber-400" : "text-text-secondary")}>
            {state.networkEnabled ? "Enabled" : "Disabled"}
          </span>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-xs text-text-muted">Environment Variables</span>
          <span className="text-xs font-medium text-text-primary">{state.environmentVars}</span>
        </div>
        {state.blockedModules.length > 0 && (
          <div>
            <span className="text-[10px] text-text-muted">Blocked Modules</span>
            <div className="flex flex-wrap gap-1 mt-1">
              {state.blockedModules.slice(0, 4).map((mod) => (
                <span key={mod} className="px-1.5 py-0.5 text-[9px] bg-error/10 text-error rounded">
                  {mod}
                </span>
              ))}
              {state.blockedModules.length > 4 && (
                <span className="text-[9px] text-text-muted">+{state.blockedModules.length - 4}</span>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Performance */}
      <div className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
        <div className="text-[10px] text-text-muted mb-2 uppercase tracking-wider">Performance</div>
        <div className="grid grid-cols-3 gap-2">
          <div className="text-center">
            <div className="text-sm font-bold text-text-primary">{state.coldStartMs}ms</div>
            <div className="text-[9px] text-text-muted">Cold Start</div>
          </div>
          <div className="text-center">
            <div className="text-sm font-bold text-text-primary">{state.executionMs}ms</div>
            <div className="text-[9px] text-text-muted">Avg Exec</div>
          </div>
          <div className="text-center">
            <div className={cn(
              "text-sm font-bold",
              state.successRate >= 95 ? "text-success" : state.successRate >= 80 ? "text-warning" : "text-error"
            )}>
              {state.successRate}%
            </div>
            <div className="text-[9px] text-text-muted">Success</div>
          </div>
        </div>
      </div>
    </div>
  );
}
