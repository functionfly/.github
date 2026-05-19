/**
 * @functionfly/ui-ghost
 * Ghost Mode: Autonomous building and orchestration components
 */

import * as React from "react";
import { cn } from "@functionfly/ui-core";
import {
  Ghost,
  Play,
  Pause,
  RotateCcw,
  ChevronRight,
  ChevronDown,
  CheckCircle,
  XCircle,
  Clock,
  Zap,
  Brain,
  Database,
  Globe,
  Layout,
  Monitor,
  FileCode,
  GitPullRequest,
  AlertTriangle,
  Loader2,
  Terminal,
  MessageSquare,
  BarChart2,
} from "lucide-react";

// ============================================================================
// Types
// ============================================================================

export type GhostPhase =
  | "planning"
  | "provisioning"
  | "building"
  | "deploying"
  | "monitoring"
  | "complete"
  | "error"
  | "paused";

export interface GhostTask {
  id: string;
  title: string;
  description: string;
  status: "pending" | "in_progress" | "completed" | "failed" | "skipped";
  phase: GhostPhase;
  startedAt?: string;
  completedAt?: string;
  duration?: number;
  logs: Array<{ timestamp: string; level: "info" | "warn" | "error" | "debug"; message: string }>;
  artifacts?: Array<{ name: string; type: string; path: string; size?: number }>;
  agentId?: string;
  confidence?: number;
  dependencies?: string[];
}

export interface GhostBuild {
  id: string;
  goal: string;
  description: string;
  phase: GhostPhase;
  progress: number; // 0-1
  tasks: GhostTask[];
  startedAt: string;
  updatedAt: string;
  estimatedCompletion?: string;
  currentTaskId?: string;
  humanApprovalRequired?: boolean;
  approvalType?: "schema" | "deployment" | "pr" | "infra";
  error?: string;
}

export interface AgentConversationMessage {
  id: string;
  agentId: string;
  agentName: string;
  agentRole: string;
  type: "thought" | "action" | "observation" | "decision" | "error" | "approval_request";
  content: string;
  timestamp: string;
  parentMessageId?: string;
  metadata?: Record<string, unknown>;
}

export interface AgentDecisionPoint {
  id: string;
  timestamp: string;
  agentId: string;
  agentName: string;
  decision: string;
  rationale: string;
  alternatives: string[];
  chosen: string;
  confidence: number;
  outcome?: string;
}

// ============================================================================
// GhostModeOrchestrator
// ============================================================================

export interface GhostModeOrchestratorProps {
  build: GhostBuild;
  onPause?: () => void;
  onResume?: () => void;
  onCancel?: () => void;
  onApprove?: (approvalType: string) => void;
  onTaskClick?: (taskId: string) => void;
  className?: string;
}

const phaseConfig: Record<GhostPhase, { color: string; icon: React.ReactNode; label: string }> = {
  planning: { color: "#8b5cf6", icon: <Brain className="size-4" />, label: "Planning" },
  provisioning: { color: "#3b82f6", icon: <Database className="size-4" />, label: "Provisioning" },
  building: { color: "#f97316", icon: <FileCode className="size-4" />, label: "Building" },
  deploying: { color: "#06b6d4", icon: <Globe className="size-4" />, label: "Deploying" },
  monitoring: { color: "#10b981", icon: <Monitor className="size-4" />, label: "Monitoring" },
  complete: { color: "#10b981", icon: <CheckCircle className="size-4" />, label: "Complete" },
  error: { color: "#ef4444", icon: <XCircle className="size-4" />, label: "Error" },
  paused: { color: "#f59e0b", icon: <Pause className="size-4" />, label: "Paused" },
};

const taskStatusIcon = (status: GhostTask["status"]) => {
  switch (status) {
    case "completed": return <CheckCircle className="size-3 text-emerald-400" />;
    case "failed": return <XCircle className="size-3 text-red-400" />;
    case "in_progress": return <Loader2 className="size-3 text-brand-400 animate-spin" />;
    case "skipped": return <Ghost className="size-3 text-gray-400" />;
    default: return <Clock className="size-3 text-text-muted" />;
  }
};

export function GhostModeOrchestrator({
  build,
  onPause,
  onResume,
  onCancel,
  onApprove,
  onTaskClick,
  className,
}: GhostModeOrchestratorProps) {
  const [expandedTasks, setExpandedTasks] = React.useState<Set<string>>(new Set());
  const [showLogPanel, setShowLogPanel] = React.useState(false);
  const [selectedTaskId, setSelectedTaskId] = React.useState<string | null>(null);

  if (!build) return null;

  const phase = phaseConfig[build.phase];
  const currentTask = build.tasks.find((t) => t.id === build.currentTaskId);

  const toggleTask = (taskId: string) => {
    setExpandedTasks((prev) => {
      const next = new Set(prev);
      if (next.has(taskId)) next.delete(taskId);
      else next.add(taskId);
      return next;
    });
  };

  return (
    <div className={cn("space-y-4", className)}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="size-8 rounded-lg flex items-center justify-center" style={{ backgroundColor: `${phase.color}20`, color: phase.color }}>
            {phase.icon}
          </div>
          <div>
            <div className="text-sm font-semibold text-text-primary flex items-center gap-2">
              Ghost Mode
              <span className="text-[10px] px-1.5 py-0.5 rounded-full font-medium" style={{ backgroundColor: `${phase.color}20`, color: phase.color }}>
                {phase.label}
              </span>
            </div>
            <p className="text-[11px] text-text-muted line-clamp-1">{build.goal}</p>
          </div>
        </div>

        {/* Controls */}
        <div className="flex items-center gap-2">
          {build.phase === "paused" ? (
            <button
              onClick={onResume}
              className="flex items-center gap-1.5 px-3 py-1.5 text-xs bg-emerald-500 hover:bg-emerald-600 text-white rounded-lg transition-colors"
            >
              <Play className="size-3" /> Resume
            </button>
          ) : build.phase === "building" || build.phase === "planning" ? (
            <button
              onClick={onPause}
              className="flex items-center gap-1.5 px-3 py-1.5 text-xs bg-amber-500/20 hover:bg-amber-500/30 text-amber-400 rounded-lg transition-colors"
            >
              <Pause className="size-3" /> Pause
            </button>
          ) : null}
          <button
            onClick={onCancel}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs bg-red-500/20 hover:bg-red-500/30 text-red-400 rounded-lg transition-colors"
          >
            <XCircle className="size-3" /> Cancel
          </button>
        </div>
      </div>

      {/* Human approval banner */}
      {build.humanApprovalRequired && (
        <div className="flex items-center justify-between p-4 bg-amber-500/10 border border-amber-500/30 rounded-lg">
          <div className="flex items-center gap-3">
            <AlertTriangle className="size-5 text-amber-400" />
            <div>
              <div className="text-sm font-medium text-amber-400">Human Approval Required</div>
              <div className="text-[11px] text-text-muted">
                {build.approvalType === "schema" && "Review and approve the generated database schema"}
                {build.approvalType === "deployment" && "Confirm production deployment"}
                {build.approvalType === "pr" && "Review the auto-created pull request"}
                {build.approvalType === "infra" && "Approve infrastructure changes"}
              </div>
            </div>
          </div>
          <div className="flex gap-2">
            <button
              onClick={() => onApprove?.(build.approvalType || "")}
              className="px-4 py-2 text-xs bg-emerald-500 hover:bg-emerald-600 text-white rounded-lg transition-colors"
            >
              Approve
            </button>
            <button
              onClick={onPause}
              className="px-4 py-2 text-xs bg-bg-secondary hover:bg-bg-hover text-text-secondary rounded-lg transition-colors"
            >
              Review Later
            </button>
          </div>
        </div>
      )}

      {/* Progress bar */}
      <div className="space-y-1.5">
        <div className="flex justify-between text-[11px] text-text-muted">
          <span>Phase: {phase.label}</span>
          <span>{Math.round(build.progress * 100)}%</span>
        </div>
        <div className="h-2 bg-bg-tertiary rounded-full overflow-hidden">
          <div
            className="h-full rounded-full transition-all duration-500"
            style={{ width: `${build.progress * 100}%`, backgroundColor: phase.color }}
          />
        </div>
      </div>

      {/* Current task indicator */}
      {currentTask && (
        <div className="p-3 bg-brand-500/10 border border-brand-500/20 rounded-lg">
          <div className="flex items-center gap-2 mb-1">
            <Loader2 className="size-3 text-brand-400 animate-spin" />
            <span className="text-xs font-medium text-brand-400">Currently working on:</span>
          </div>
          <div className="text-sm text-text-primary">{currentTask.title}</div>
        </div>
      )}

      {/* Task list */}
      <div className="space-y-1 max-h-80 overflow-y-auto">
        {build.tasks.map((task) => {
          const isExpanded = expandedTasks.has(task.id);
          const isSelected = selectedTaskId === task.id;
          const taskPhase = phaseConfig[task.status === "in_progress" ? build.phase : task.status === "completed" ? "complete" : task.status === "failed" ? "error" : "planning"];

          return (
            <div key={task.id}>
              <div
                className={cn(
                  "flex items-center gap-2 p-2.5 rounded-lg border cursor-pointer transition-all",
                  isSelected ? "border-brand-500 bg-brand-500/5" : "border-border-subtle hover:border-border-default",
                  task.status === "completed" && "opacity-60",
                  task.status === "failed" && "border-red-500/30 bg-red-500/5"
                )}
                onClick={() => {
                  setSelectedTaskId(task.id);
                  toggleTask(task.id);
                  onTaskClick?.(task.id);
                }}
              >
                {taskStatusIcon(task.status)}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-medium text-text-primary truncate">{task.title}</span>
                    {task.duration && (
                      <span className="text-[9px] text-text-muted">{task.duration}s</span>
                    )}
                  </div>
                  <p className="text-[10px] text-text-muted line-clamp-1">{task.description}</p>
                </div>
                {task.logs.length > 0 && (
                  <button
                    onClick={(e) => { e.stopPropagation(); setShowLogPanel(true); setSelectedTaskId(task.id); }}
                    className="text-[10px] text-brand-400 hover:text-brand-300"
                  >
                    {task.logs.length} logs
                  </button>
                )}
                {task.artifacts && task.artifacts.length > 0 && (
                  <span className="text-[10px] text-text-muted">{task.artifacts.length} artifacts</span>
                )}
                {isExpanded ? <ChevronDown className="size-3 text-text-muted" /> : <ChevronRight className="size-3 text-text-muted" />}
              </div>

              {/* Expanded task detail */}
              {isExpanded && (
                <div className="ml-6 mt-1 p-3 bg-bg-secondary rounded-lg border border-border-subtle space-y-2">
                  {task.description && (
                    <div>
                      <span className="text-[10px] font-medium text-text-muted">Description</span>
                      <p className="text-xs text-text-secondary">{task.description}</p>
                    </div>
                  )}

                  {/* Logs */}
                  {task.logs.length > 0 && (
                    <div>
                      <span className="text-[10px] font-medium text-text-muted mb-1 block">Recent Logs</span>
                      <div className="bg-bg-primary rounded p-2 font-mono text-[10px] max-h-32 overflow-y-auto">
                        {task.logs.slice(-5).map((log, i) => (
                          <div key={i} className={cn(
                            "flex gap-2 py-0.5",
                            log.level === "error" && "text-red-400",
                            log.level === "warn" && "text-amber-400"
                          )}>
                            <span className="text-text-muted shrink-0">{log.timestamp.split("T")[1]?.slice(0, 8) || ""}</span>
                            <span className="uppercase font-bold shrink-0 w-12">{log.level}</span>
                            <span className="text-text-secondary">{log.message}</span>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Artifacts */}
                  {task.artifacts && task.artifacts.length > 0 && (
                    <div>
                      <span className="text-[10px] font-medium text-text-muted mb-1 block">Artifacts</span>
                      <div className="flex flex-wrap gap-2">
                        {task.artifacts.map((artifact, i) => (
                          <div key={i} className="flex items-center gap-1.5 px-2 py-1 bg-bg-tertiary rounded text-[10px] text-text-secondary">
                            <FileCode className="size-3" />
                            <span>{artifact.name}</span>
                            {artifact.size && <span className="text-text-muted">({artifact.size}KB)</span>}
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Agent */}
                  {task.agentId && (
                    <div className="flex items-center gap-2 text-[10px] text-text-muted">
                      <Ghost className="size-3" />
                      <span>Agent: {task.agentId}</span>
                      {task.confidence && <span>Confidence: {Math.round(task.confidence * 100)}%</span>}
                    </div>
                  )}
                </div>
              )}
            </div>
          );
        })}
      </div>

      {/* Est. completion */}
      {build.estimatedCompletion && (
        <div className="text-center text-[10px] text-text-muted">
          Estimated completion: {new Date(build.estimatedCompletion).toLocaleTimeString()}
        </div>
      )}
    </div>
  );
}

// ============================================================================
// AgentConversationTimeline
// ============================================================================

export interface AgentConversationTimelineProps {
  messages: AgentConversationMessage[];
  decisions: AgentDecisionPoint[];
  onMessageClick?: (messageId: string) => void;
  onDecisionClick?: (decisionId: string) => void;
  className?: string;
}

const messageTypeConfig: Record<string, { color: string; icon: React.ReactNode; bg: string }> = {
  thought: { color: "#8b5cf6", icon: <Brain className="size-3" />, bg: "bg-purple-500/10" },
  action: { color: "#3b82f6", icon: <Zap className="size-3" />, bg: "bg-blue-500/10" },
  observation: { color: "#10b981", icon: <CheckCircle className="size-3" />, bg: "bg-emerald-500/10" },
  decision: { color: "#f97316", icon: <GitPullRequest className="size-3" />, bg: "bg-orange-500/10" },
  error: { color: "#ef4444", icon: <AlertTriangle className="size-3" />, bg: "bg-red-500/10" },
  approval_request: { color: "#f59e0b", icon: <AlertTriangle className="size-3" />, bg: "bg-amber-500/10" },
};

export function AgentConversationTimeline({
  messages = [],
  decisions = [],
  onMessageClick,
  onDecisionClick,
  className,
}: AgentConversationTimelineProps) {
  const [filterType, setFilterType] = React.useState<string | null>(null);
  const [selectedMessage, setSelectedMessage] = React.useState<string | null>(null);

  const filteredMessages = filterType
    ? messages.filter((m) => m.type === filterType)
    : messages;

  const timelineMessages = filteredMessages.sort(
    (a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
  );

  return (
    <div className={cn("space-y-4", className)}>
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
          <MessageSquare className="size-4 text-brand-400" /> Agent Conversation
        </h4>

        {/* Filter tabs */}
        <div className="flex gap-1">
          {["thought", "action", "decision"].map((type) => {
            const config = messageTypeConfig[type];
            return (
              <button
                key={type}
                onClick={() => setFilterType(filterType === type ? null : type)}
                className={cn(
                  "flex items-center gap-1 px-2 py-0.5 text-[10px] rounded capitalize transition-colors",
                  filterType === type ? "font-medium" : "text-text-muted hover:text-text-secondary"
                )}
                style={filterType === type ? { color: config.color, backgroundColor: config.bg } : {}}
              >
                {config.icon}
                {type}
              </button>
            );
          })}
          {filterType && (
            <button
              onClick={() => setFilterType(null)}
              className="text-[10px] text-text-muted hover:text-text-secondary ml-1"
            >
              Clear
            </button>
          )}
        </div>
      </div>

      {/* Decision summary */}
      {decisions.length > 0 && (
        <div className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
          <div className="text-[10px] font-medium text-text-muted mb-2">Decision Summary</div>
          <div className="flex gap-4">
            <div className="text-center">
              <div className="text-lg font-bold text-text-primary">{decisions.length}</div>
              <div className="text-[9px] text-text-muted">Decisions</div>
            </div>
            <div className="text-center">
              <div className="text-lg font-bold text-brand-500">
                {Math.round(decisions.reduce((s, d) => s + d.confidence, 0) / decisions.length * 100)}%
              </div>
              <div className="text-[9px] text-text-muted">Avg Confidence</div>
            </div>
            <div className="text-center">
              <div className="text-lg font-bold text-emerald-400">
                {decisions.filter((d) => d.outcome === "success").length}
              </div>
              <div className="text-[9px] text-text-muted">Successful</div>
            </div>
          </div>
        </div>
      )}

      {/* Timeline */}
      <div className="relative">
        {/* Vertical line */}
        <div className="absolute left-4 top-0 bottom-0 w-px bg-border-subtle" />

        <div className="space-y-3 pl-4">
          {timelineMessages.map((message) => {
            const config = messageTypeConfig[message.type] || messageTypeConfig.thought;
            const isSelected = selectedMessage === message.id;

            return (
              <div
                key={message.id}
                className={cn(
                  "relative p-3 rounded-lg border cursor-pointer transition-all",
                  isSelected ? "border-brand-500 bg-brand-500/5" : "border-border-subtle hover:border-border-default",
                  config.bg
                )}
                onClick={() => {
                  setSelectedMessage(isSelected ? null : message.id);
                  onMessageClick?.(message.id);
                }}
              >
                {/* Timeline dot */}
                <div
                  className="absolute -left-4 top-4 size-2 rounded-full border-2 border-bg-primary"
                  style={{ backgroundColor: config.color }}
                />

                {/* Header */}
                <div className="flex items-center gap-2 mb-1.5">
                  <div
                    className="size-5 rounded flex items-center justify-center"
                    style={{ backgroundColor: `${config.color}20`, color: config.color }}
                  >
                    {config.icon}
                  </div>
                  <span className="text-xs font-medium text-text-primary">{message.agentName}</span>
                  <span className="text-[9px] px-1.5 py-0.5 rounded bg-bg-tertiary text-text-muted capitalize">
                    {message.agentRole}
                  </span>
                  <span className="text-[9px] text-text-muted ml-auto">
                    {new Date(message.timestamp).toLocaleTimeString()}
                  </span>
                </div>

                {/* Content */}
                <p className="text-[11px] text-text-secondary leading-relaxed">{message.content}</p>

                {/* Expanded detail */}
                {isSelected && message.metadata && Object.keys(message.metadata).length > 0 && (
                  <div className="mt-2 pt-2 border-t border-border-subtle">
                    <div className="text-[9px] text-text-muted mb-1">Metadata</div>
                    <div className="grid grid-cols-2 gap-1 text-[10px]">
                      {Object.entries(message.metadata).map(([key, value]) => (
                        <div key={key} className="flex justify-between">
                          <span className="text-text-muted">{key}</span>
                          <span className="text-text-primary font-mono">{String(value).slice(0, 30)}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>

      {/* Decision detail panel */}
      {decisions.length > 0 && (
        <div className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
          <div className="text-[10px] font-medium text-text-muted mb-2">Recent Decisions</div>
          <div className="space-y-2">
            {decisions.slice(0, 3).map((decision) => {
              const confColor = decision.confidence > 0.8 ? "#10b981" : decision.confidence > 0.5 ? "#f59e0b" : "#ef4444";
              return (
                <div
                  key={decision.id}
                  className="p-2 bg-bg-primary rounded border border-border-subtle cursor-pointer hover:border-border-default"
                  onClick={() => onDecisionClick?.(decision.id)}
                >
                  <div className="flex items-center justify-between mb-1">
                    <span className="text-xs font-medium text-text-primary">{decision.agentName}</span>
                    <span className="text-[10px]" style={{ color: confColor }}>
                      {Math.round(decision.confidence * 100)}% confidence
                    </span>
                  </div>
                  <p className="text-[10px] text-text-muted line-clamp-1">{decision.decision}</p>
                  <p className="text-[9px] text-text-muted mt-0.5">Chose: {decision.chosen}</p>
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}

// ============================================================================
// MultiAgentConversationView
// ============================================================================

export interface MultiAgentConversationViewProps {
  conversations: Array<{
    agentId: string;
    agentName: string;
    agentRole: string;
    messages: AgentConversationMessage[];
  }>;
  onAgentClick?: (agentId: string) => void;
  className?: string;
}

export function MultiAgentConversationView({
  conversations = [],
  onAgentClick,
  className,
}: MultiAgentConversationViewProps) {
  const [selectedAgent, setSelectedAgent] = React.useState<string | null>(
    conversations[0]?.agentId || null
  );
  const activeConversation = conversations.find((c) => c.agentId === selectedAgent);

  return (
    <div className={cn("space-y-3", className)}>
      <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
        <Users className="size-4 text-brand-400" /> Multi-Agent Conversation
      </h4>

      {/* Agent tabs */}
      <div className="flex gap-1 border-b border-border-subtle overflow-x-auto">
        {conversations.map((conv) => (
          <button
            key={conv.agentId}
            onClick={() => setSelectedAgent(conv.agentId)}
            className={cn(
              "flex items-center gap-2 px-3 py-1.5 text-xs font-medium whitespace-nowrap transition-colors",
              selectedAgent === conv.agentId
                ? "text-brand-500 border-b-2 border-brand-500"
                : "text-text-muted hover:text-text-primary"
            )}
            onDoubleClick={() => onAgentClick?.(conv.agentId)}
          >
            <Ghost className="size-3" />
            {conv.agentName}
            <span className="text-[9px] bg-bg-tertiary px-1 rounded">{(conv.messages || []).length}</span>
          </button>
        ))}
      </div>

      {/* Messages */}
      {activeConversation && (
        <div className="space-y-2 max-h-96 overflow-y-auto">
          {activeConversation.messages.map((message) => {
            const config = messageTypeConfig[message.type] || messageTypeConfig.thought;
            return (
              <div key={message.id} className={cn("p-3 rounded-lg border border-border-subtle", config.bg)}>
                <div className="flex items-center gap-2 mb-1">
                  <span className="text-[10px] font-medium text-text-primary">{message.agentName}</span>
                  <span className="text-[9px] text-text-muted capitalize">{message.type}</span>
                  <span className="text-[9px] text-text-muted ml-auto">
                    {new Date(message.timestamp).toLocaleTimeString()}
                  </span>
                </div>
                <p className="text-[11px] text-text-secondary">{message.content}</p>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

// Helper
function Users({ className }: { className?: string }) {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={className}>
      <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
      <circle cx="9" cy="7" r="4" />
      <path d="M22 21v-2a4 4 0 0 0-3-3.87" />
      <path d="M16 3.13a4 4 0 0 1 0 7.75" />
    </svg>
  );
}

// ============================================================================
// Index
// ============================================================================

export type {
  GhostPhase,
  GhostTask,
  GhostBuild,
  AgentConversationMessage,
  AgentDecisionPoint,
} from "./types";