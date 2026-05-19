/**
 * @functionfly/ui-agent
 * Agent consensus and conflict resolution components
 */

import * as React from "react";
import { cn } from "@functionfly/ui-core";
import { Badge } from "@functionfly/ui-core";

// Types
export interface ConsensusVote {
  agentId: string;
  agentName: string;
  vote: "approve" | "reject" | "abstain";
  reason?: string;
  timestamp: string;
  confidence: number; // 0-1
}

export interface ConsensusDecision {
  id: string;
  topic: string;
  description: string;
  status: "pending" | "decided" | "expired" | "disputed";
  votes: ConsensusVote[];
  outcome?: "approved" | "rejected" | "compromised";
  decidedAt?: string;
  threshold: number; // 0-1, required confidence to decide
}

export interface ConflictEntry {
  id: string;
  type: "resource" | "goal" | "approach" | "priority" | "communication";
  title: string;
  description: string;
  involvedAgents: string[];
  severity: "low" | "medium" | "high" | "critical";
  status: "active" | "resolved" | "escalated";
  resolution?: string;
  startedAt: string;
  resolvedAt?: string;
}

export interface AgentConsensusViewerProps {
  decisions: ConsensusDecision[];
  onVote?: (decisionId: string, vote: ConsensusVote["vote"]) => void;
  onEscalate?: (decisionId: string) => void;
  className?: string;
}

export interface AgentConflictResolverProps {
  conflicts: ConflictEntry[];
  onResolve?: (conflictId: string, resolution: string) => void;
  onEscalate?: (conflictId: string) => void;
  onAgentMediate?: (conflictId: string, mediatorAgentId: string) => void;
  className?: string;
}

// Helper
function getDecisionSummary(decisions: ConsensusDecision[]) {
  const total = decisions.length;
  const approved = decisions.filter((d) => d.outcome === "approved").length;
  const rejected = decisions.filter((d) => d.outcome === "rejected").length;
  const pending = decisions.filter((d) => d.status === "pending").length;
  return { total, approved, rejected, pending };
}

function getConflictStats(conflicts: ConflictEntry[]) {
  const total = conflicts.length;
  const active = conflicts.filter((c) => c.status === "active").length;
  const critical = conflicts.filter((c) => c.severity === "critical" && c.status === "active").length;
  return { total, active, critical };
}

// Components

export function AgentConsensusViewer({
  decisions,
  onVote,
  onEscalate,
  className,
}: AgentConsensusViewerProps) {
  const [selectedDecision, setSelectedDecision] = React.useState<string | null>(null);
  const [filterStatus, setFilterStatus] = React.useState<ConsensusDecision["status"] | "all">("all");
  const summary = getDecisionSummary(decisions);

  const filteredDecisions = filterStatus === "all"
    ? decisions
    : decisions.filter((d) => d.status === filterStatus);

  return (
    <div className={cn("space-y-4", className)}>
      {/* Header stats */}
      <div className="grid grid-cols-4 gap-3">
        <div className="p-3 bg-bg-secondary border border-border-subtle rounded-lg">
          <div className="text-2xl font-bold text-text-primary">{summary.total}</div>
          <div className="text-xs text-text-muted">Total Decisions</div>
        </div>
        <div className="p-3 bg-bg-secondary border border-border-subtle rounded-lg">
          <div className="text-2xl font-bold text-emerald-400">{summary.approved}</div>
          <div className="text-xs text-text-muted">Approved</div>
        </div>
        <div className="p-3 bg-bg-secondary border border-border-subtle rounded-lg">
          <div className="text-2xl font-bold text-red-400">{summary.rejected}</div>
          <div className="text-xs text-text-muted">Rejected</div>
        </div>
        <div className="p-3 bg-bg-secondary border border-border-subtle rounded-lg">
          <div className="text-2xl font-bold text-amber-400">{summary.pending}</div>
          <div className="text-xs text-text-muted">Pending</div>
        </div>
      </div>

      {/* Filter tabs */}
      <div className="flex gap-1 border-b border-border-subtle">
        {(["all", "pending", "decided", "expired", "disputed"] as const).map((status) => (
          <button
            key={status}
            onClick={() => setFilterStatus(status)}
            className={cn(
              "px-3 py-1.5 text-xs font-medium capitalize transition-colors",
              filterStatus === status
                ? "text-brand-500 border-b-2 border-brand-500"
                : "text-text-muted hover:text-text-primary"
            )}
          >
            {status}
          </button>
        ))}
      </div>

      {/* Decision list */}
      <div className="space-y-2 max-h-80 overflow-y-auto">
        {filteredDecisions.map((decision) => {
          const hasVoted = decision.votes.length > 0;
          const approvalRate = decision.votes.length
            ? decision.votes.filter((v) => v.vote === "approve").length / decision.votes.length
            : 0;
          const avgConfidence = decision.votes.length
            ? decision.votes.reduce((sum, v) => sum + v.confidence, 0) / decision.votes.length
            : 0;

          return (
            <div
              key={decision.id}
              className={cn(
                "p-3 bg-bg-secondary border rounded-lg cursor-pointer transition-all",
                selectedDecision === decision.id
                  ? "border-brand-500/50"
                  : "border-border-subtle hover:border-border-default"
              )}
              onClick={() => setSelectedDecision(decision.id === selectedDecision ? null : decision.id)}
            >
              <div className="flex items-start justify-between gap-2">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="text-xs font-medium text-text-primary truncate">{decision.topic}</span>
                    <Badge
                      variant={decision.status === "decided" ? "success" : decision.status === "pending" ? "warning" : "error"}
                      size="sm"
                    >
                      {decision.status}
                    </Badge>
                  </div>
                  <p className="text-[11px] text-text-muted line-clamp-2">{decision.description}</p>
                </div>
                <div className="flex flex-col items-end gap-1">
                  <span className="text-xs font-medium text-text-primary">
                    {Math.round(approvalRate * 100)}%
                  </span>
                  <span className="text-[10px] text-text-muted">
                    {decision.votes.length} votes
                  </span>
                </div>
              </div>

              {/* Progress bar */}
              <div className="mt-2 h-1 bg-bg-tertiary rounded-full overflow-hidden">
                <div
                  className="h-full rounded-full transition-all"
                  style={{
                    width: `${approvalRate * 100}%`,
                    backgroundColor: approvalRate >= decision.threshold ? "#10b981" : "#f59e0b",
                  }}
                />
              </div>

              {/* Expanded details */}
              {selectedDecision === decision.id && (
                <div className="mt-3 pt-3 border-t border-border-subtle">
                  {/* Vote breakdown */}
                  <div className="space-y-2 mb-3">
                    <div className="text-[11px] font-medium text-text-muted">Votes</div>
                    {decision.votes.map((vote) => (
                      <div key={vote.agentId} className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <div
                            className={cn(
                              "size-2 rounded-full",
                              vote.vote === "approve" ? "bg-emerald-400" :
                              vote.vote === "reject" ? "bg-red-400" : "bg-gray-400"
                            )}
                          />
                          <span className="text-xs text-text-primary">{vote.agentName}</span>
                        </div>
                        <span className="text-[10px] text-text-muted capitalize">{vote.vote}</span>
                        <span className="text-[10px] text-text-muted">
                          {Math.round(vote.confidence * 100)}% confidence
                        </span>
                      </div>
                    ))}
                  </div>

                  {/* Actions */}
                  {decision.status === "pending" && (
                    <div className="flex gap-2">
                      <button
                        onClick={(e) => { e.stopPropagation(); onVote?.(decision.id, "approve"); }}
                        className="flex-1 px-3 py-1.5 text-xs bg-emerald-500/20 text-emerald-400 rounded hover:bg-emerald-500/30 transition-colors"
                      >
                        Approve
                      </button>
                      <button
                        onClick={(e) => { e.stopPropagation(); onVote?.(decision.id, "reject"); }}
                        className="flex-1 px-3 py-1.5 text-xs bg-red-500/20 text-red-400 rounded hover:bg-red-500/30 transition-colors"
                      >
                        Reject
                      </button>
                      <button
                        onClick={(e) => { e.stopPropagation(); onEscalate?.(decision.id); }}
                        className="flex-1 px-3 py-1.5 text-xs bg-amber-500/20 text-amber-400 rounded hover:bg-amber-500/30 transition-colors"
                      >
                        Escalate
                      </button>
                    </div>
                  )}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}

export function AgentConflictResolver({
  conflicts,
  onResolve,
  onEscalate,
  onAgentMediate,
  className,
}: AgentConflictResolverProps) {
  const [selectedConflict, setSelectedConflict] = React.useState<string | null>(null);
  const [filterSeverity, setFilterSeverity] = React.useState<ConflictEntry["severity"] | "all">("all");
  const [filterStatus, setFilterStatus] = React.useState<ConflictEntry["status"] | "all">("all");
  const [resolutionInput, setResolutionInput] = React.useState("");
  const stats = getConflictStats(conflicts);

  const filteredConflicts = conflicts.filter((c) => {
    if (filterSeverity !== "all" && c.severity !== filterSeverity) return false;
    if (filterStatus !== "all" && c.status !== filterStatus) return false;
    return true;
  });

  return (
    <div className={cn("space-y-4", className)}>
      {/* Header stats */}
      <div className="grid grid-cols-3 gap-3">
        <div className="p-3 bg-bg-secondary border border-border-subtle rounded-lg">
          <div className="text-2xl font-bold text-text-primary">{stats.total}</div>
          <div className="text-xs text-text-muted">Total Conflicts</div>
        </div>
        <div className="p-3 bg-bg-secondary border border-border-subtle rounded-lg">
          <div className="text-2xl font-bold text-amber-400">{stats.active}</div>
          <div className="text-xs text-text-muted">Active</div>
        </div>
        <div className="p-3 bg-bg-secondary border border-border-subtle rounded-lg">
          <div className="text-2xl font-bold text-red-400">{stats.critical}</div>
          <div className="text-xs text-text-muted">Critical</div>
        </div>
      </div>

      {/* Filters */}
      <div className="flex gap-4">
        <div className="flex gap-1 border-b border-border-subtle">
          {(["all", "active", "resolved", "escalated"] as const).map((status) => (
            <button
              key={status}
              onClick={() => setFilterStatus(status)}
              className={cn(
                "px-3 py-1.5 text-xs font-medium capitalize transition-colors",
                filterStatus === status
                  ? "text-brand-500 border-b-2 border-brand-500"
                  : "text-text-muted hover:text-text-primary"
              )}
            >
              {status}
            </button>
          ))}
        </div>
        <div className="flex gap-1 border-b border-border-subtle">
          {(["all", "low", "medium", "high", "critical"] as const).map((severity) => (
            <button
              key={severity}
              onClick={() => setFilterSeverity(severity)}
              className={cn(
                "px-3 py-1.5 text-xs font-medium capitalize transition-colors",
                filterSeverity === severity
                  ? "text-brand-500 border-b-2 border-brand-500"
                  : "text-text-muted hover:text-text-primary"
              )}
            >
              {severity}
            </button>
          ))}
        </div>
      </div>

      {/* Conflict list */}
      <div className="space-y-2 max-h-96 overflow-y-auto">
        {filteredConflicts.map((conflict) => (
          <div
            key={conflict.id}
            className={cn(
              "p-3 bg-bg-secondary border rounded-lg cursor-pointer transition-all",
              selectedConflict === conflict.id
                ? "border-brand-500/50"
                : "border-border-subtle hover:border-border-default",
              conflict.severity === "critical" && "border-red-500/30 bg-red-500/5",
              conflict.severity === "high" && "border-amber-500/30"
            )}
            onClick={() => setSelectedConflict(conflict.id === selectedConflict ? null : conflict.id)}
          >
            <div className="flex items-start justify-between gap-2">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <span className="text-xs font-medium text-text-primary truncate">{conflict.title}</span>
                  <Badge
                    variant={conflict.severity === "critical" ? "error" : conflict.severity === "high" ? "warning" : "default"}
                    size="sm"
                  >
                    {conflict.severity}
                  </Badge>
                  <Badge
                    variant={conflict.status === "active" ? "info" : conflict.status === "resolved" ? "success" : "warning"}
                    size="sm"
                  >
                    {conflict.status}
                  </Badge>
                </div>
                <p className="text-[11px] text-text-muted line-clamp-2">{conflict.description}</p>
                <div className="flex items-center gap-2 mt-1.5">
                  <span className="text-[10px] text-text-muted">Involved:</span>
                  {conflict.involvedAgents.slice(0, 3).map((agentId) => (
                    <span key={agentId} className="px-1.5 py-0.5 text-[10px] bg-bg-tertiary text-text-muted rounded">
                      {agentId}
                    </span>
                  ))}
                  {conflict.involvedAgents.length > 3 && (
                    <span className="text-[10px] text-text-muted">+{conflict.involvedAgents.length - 3}</span>
                  )}
                </div>
              </div>
              <span className="text-[10px] text-text-muted whitespace-nowrap">
                {new Date(conflict.startedAt).toLocaleDateString()}
              </span>
            </div>

            {/* Expanded resolution panel */}
            {selectedConflict === conflict.id && (
              <div className="mt-3 pt-3 border-t border-border-subtle">
                {conflict.status === "resolved" && conflict.resolution ? (
                  <div className="p-2 bg-emerald-500/10 border border-emerald-500/20 rounded">
                    <div className="text-[11px] font-medium text-emerald-400 mb-1">Resolution</div>
                    <div className="text-[11px] text-text-primary">{conflict.resolution}</div>
                  </div>
                ) : conflict.status === "active" ? (
                  <div className="space-y-3">
                    <div>
                      <div className="text-[11px] font-medium text-text-muted mb-1">Type</div>
                      <Badge size="sm" variant="default">{conflict.type}</Badge>
                    </div>
                    <div>
                      <div className="text-[11px] font-medium text-text-muted mb-1">Suggest Resolution</div>
                      <textarea
                        value={resolutionInput}
                        onChange={(e) => setResolutionInput(e.target.value)}
                        placeholder="Enter resolution details..."
                        className="w-full p-2 bg-bg-tertiary border border-border-subtle rounded text-xs text-text-primary placeholder-text-muted resize-none"
                        rows={2}
                        onClick={(e) => e.stopPropagation()}
                      />
                    </div>
                    <div className="flex gap-2">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          if (resolutionInput.trim()) {
                            onResolve?.(conflict.id, resolutionInput);
                            setResolutionInput("");
                          }
                        }}
                        className="flex-1 px-3 py-1.5 text-xs bg-emerald-500/20 text-emerald-400 rounded hover:bg-emerald-500/30 transition-colors"
                      >
                        Resolve
                      </button>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          onEscalate?.(conflict.id);
                        }}
                        className="flex-1 px-3 py-1.5 text-xs bg-amber-500/20 text-amber-400 rounded hover:bg-amber-500/30 transition-colors"
                      >
                        Escalate
                      </button>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          onAgentMediate?.(conflict.id, "mediator-agent-id");
                        }}
                        className="flex-1 px-3 py-1.5 text-xs bg-blue-500/20 text-blue-400 rounded hover:bg-blue-500/30 transition-colors"
                      >
                        Request Mediation
                      </button>
                    </div>
                  </div>
                ) : (
                  <div className="text-[11px] text-text-muted">
                    {conflict.status === "escalated" ? "Escalated for review" : "Closed"}
                  </div>
                )}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}