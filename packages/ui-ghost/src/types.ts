/**
 * @functionfly/ui-ghost
 * Type definitions
 */

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
  progress: number;
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