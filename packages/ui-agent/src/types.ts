/**
 * @functionfly/ui-agent
 * Type definitions
 * Note: Most types are re-exported from AgentDock.tsx
 * This file contains additional types not in AgentDock.tsx
 */

export interface AgentMemory {
  id: string;
  agent_id: string;
  memory_type: "working" | "longterm" | "context" | "episodic";
  content: string;
  structured_data?: Record<string, unknown>;
  importance_score: number;
  last_accessed: string;
  created_at: string;
  updated_at: string;
}

export interface ListMemoriesResponse {
  memories: AgentMemory[];
  total: number;
  limit: number;
  offset: number;
}

export interface Task {
  id: string;
  title: string;
  description?: string;
  status: "todo" | "in-progress" | "done" | "blocked";
  priority: "low" | "medium" | "high";
  assignedTo?: string;
  createdAt: string;
  updatedAt: string;
}

export interface AutonomousTaskBoardProps {
  tasks: Task[];
  agents?: Array<{ id: string; name: string; isAI?: boolean }>;
  onTaskCreate?: (task: Partial<Task>) => void;
  onTaskUpdate?: (task: Partial<Task> & { id: string }) => void;
  onTaskDelete?: (id: string) => void;
  onTaskAssign?: (taskId: string, agentId?: string) => void;
}

export interface ConsensusVote {
  id: string;
  agentId: string;
  agentName: string;
  decision: "approve" | "reject" | "abstain";
  rationale?: string;
  timestamp: string;
}

export interface ConsensusDecision {
  id: string;
  title: string;
  description: string;
  votes: ConsensusVote[];
  outcome: "approved" | "rejected" | "pending";
  createdAt: string;
  resolvedAt?: string;
}

export interface ConflictEntry {
  id: string;
  field: string;
  current: { user: string; value: string };
  incoming: { user: string; value: string };
  timestamp: string;
}

export interface AgentConsensusViewerProps {
  decisions: ConsensusDecision[];
  currentDecision?: string;
  onDecisionSelect?: (id: string) => void;
}

export interface AgentConflictResolverProps {
  conflicts: ConflictEntry[];
  onResolve?: (id: string, resolution: "current" | "incoming") => void;
}

export interface AgentRuntimeInspectorProps {
  agentId: string;
  sandboxState: {
    status: "idle" | "active" | "error";
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
  className?: string;
}