import { apiClient } from "./client";

// ============================================================================
// Types
// ============================================================================

export interface AgentIdentity {
  id: string;
  agentId: string;
  name: string;
  description?: string;
  tenantId: string;
  status: string;
  createdAt: string;
  updatedAt: string;
  parentAgentId?: string;
  swarmRole?: string;
  maxChildAgents?: number;
  capabilities?: Record<string, unknown>;
  autonomousEnabled?: boolean;
  evolutionEnabled?: boolean;
}

export interface RegisterAgentRequest {
  agentId: string;
  name: string;
  description?: string;
}

export interface RegisterAgentResponse {
  ok: boolean;
  agent: AgentIdentity;
  api_key: string;
}

export interface AgentListResponse {
  ok: boolean;
  agents: AgentIdentity[];
  total: number;
  limit: number;
  offset: number;
}

export interface AgentQuota {
  maxCallsPerMinute?: number;
  maxConcurrentExecutions?: number;
  maxMemoryMB?: number;
  maxExecutionTimeMs?: number;
  dailySpendCap?: number;
}

export interface AgentUsage {
  callsThisMinute: number;
  concurrentExecutions: number;
  memoryUsageMB: number;
  executionTimeMs: number;
  spendToday: number;
}

export interface BehavioralPolicy {
  agentId: string;
  maxExecutionDepth: number;
  maxRecursionDepth: number;
  maxWallTimeMs: number;
  maxMemoryGrowthMB: number;
  forbiddenFunctions?: string[];
  deterministicOnly?: boolean;
  allowedCapabilities?: string[];
}

export interface ExecutionRecord {
  id: string;
  executionId: string;
  agentId: string;
  functionAuthor: string;
  functionName: string;
  functionVersion?: string;
  outcome: string;
  latencyMs: number;
  costUsd: number;
  timestamp: string;
  errorCode?: string;
}

export interface ExecutionListResponse {
  ok: boolean;
  executions: ExecutionRecord[];
  total: number;
  limit: number;
  offset: number;
}

export interface AgentAnalytics {
  totalExecutions: number;
  successRate: number;
  avgLatencyMs: number;
  avgCostUsd: number;
  period: string;
}

export interface AgentSession {
  id: string;
  agentId: string;
  status: string;
  startedAt: string;
  endedAt?: string;
  metadata?: Record<string, unknown>;
}

export interface BillingSummary {
  agentId: string;
  currentBalance: number;
  spendThisPeriod: number;
  spendCap?: number;
  periodStart: string;
  periodEnd: string;
}

export interface CostBreakdown {
  agentId: string;
  totalCost: number;
  byFunction: Record<string, number>;
  byPeriod: Record<string, number>;
}

export interface CreditBalance {
  agentId: string;
  balance: number;
  lastPurchase?: string;
}

export interface ConcurrencyStats {
  totalAgents: number;
  activeAgents: number;
  totalConcurrent: number;
  maxConcurrent: number;
  queueDepth: number;
}

// ============================================================================
// Swarm Types
// ============================================================================

export interface ChildAgent {
  id: string;
  name: string;
  status: "active" | "suspended" | "pending";
  swarmRole: "worker" | "manager" | "infrastructure";
  trustScore: number;
  economicScore: number;
  parentAgentId?: string;
}

export interface SwarmStats {
  totalAgents: number;
  activeAgents: number;
  totalMessages: number;
  pendingMessages: number;
  walletBalance: number;
  revenueThisMonth: number;
}

export interface AgentWallet {
  agentId: string;
  balanceUSD: number;
  escrowBalanceUSD: number;
  totalEarnedUSD: number;
  totalSpentUSD: number;
}

export interface AgentMessage {
  id: string;
  fromAgentId: string;
  toAgentId: string;
  messageType: string;
  payload?: Record<string, unknown>;
  status: string;
  createdAt: string;
  deliveredAt?: string;
}

export interface MarketplaceAgent {
  id: string;
  agentId: string;
  name: string;
  description: string;
  listingType: "worker" | "manager" | "infrastructure";
  pricingModel: "free" | "per_call" | "subscription" | "revenue_share";
  pricePerCall?: number;
  subscriptionMonthlyUsd?: number;
  revenueSharePercent?: number;
  ratingScore: number;
  totalCalls: number;
  roiScore: number;
  trustScore?: number;
  deterministicVerified?: boolean;
  capabilities?: string[];
}

export interface MarketplaceAgentSearchParams {
  pricing_model?: string;
  min_rating?: number;
  max_price_per_call?: number;
  min_roi_score?: number;
  listing_types?: string[];
  limit?: number;
  offset?: number;
}

export interface EvolutionProposal {
  id: string;
  agentId: string;
  proposalType: string;
  status: "pending" | "approved" | "rejected" | "implemented";
  proposalData?: Record<string, unknown>;
  createdAt: string;
  updatedAt: string;
}

export interface AutonomySchedule {
  id: string;
  agentId: string;
  scheduleType: string;
  actionType: string;
  cronExpression?: string;
  nextRunAt?: string;
  isActive: boolean;
  createdAt: string;
}

// ============================================================================
// Agent API
// ============================================================================

export const agentApi = {
  // ---------------------------------------------------------------------------
  // Agent Management
  // ---------------------------------------------------------------------------

  /**
   * Register a new agent.
   * POST /v1/agent/register
   */
  registerAgent: (data: RegisterAgentRequest) =>
    apiClient.post<RegisterAgentResponse>("/v1/agent/register", data),

  /**
   * List all agents for the authenticated tenant.
   * GET /v1/agent
   */
  listAgents: (params?: { limit?: number; offset?: number }) =>
    apiClient.get<AgentListResponse>("/v1/agent", { params }),

  /**
   * Get an agent by ID.
   * GET /v1/agent/{agent_id}
   */
  getAgent: (agentId: string) =>
    apiClient.get<{ ok: boolean; agent: AgentIdentity }>(`/v1/agent/${agentId}`),

  /**
   * Delete an agent.
   * DELETE /v1/agent/{agent_id}
   */
  deleteAgent: (agentId: string) =>
    apiClient.delete<{ ok: boolean; message: string }>(`/v1/agent/${agentId}`),

  // ---------------------------------------------------------------------------
  // Quota Management
  // ---------------------------------------------------------------------------

  /**
   * Update agent quota configuration.
   * PUT /v1/agent/{agent_id}/quota
   */
  updateQuota: (agentId: string, quota: AgentQuota) =>
    apiClient.put<{ ok: boolean; quota: AgentQuota }>(
      `/v1/agent/${agentId}/quota`,
      quota
    ),

  /**
   * Get current agent usage.
   * GET /v1/agent/{agent_id}/usage
   */
  getUsage: (agentId: string) =>
    apiClient.get<{ ok: boolean; usage: AgentUsage }>(
      `/v1/agent/${agentId}/usage`
    ),

  // ---------------------------------------------------------------------------
  // Policy Management
  // ---------------------------------------------------------------------------

  /**
   * Get agent behavioral policy.
   * GET /v1/agent/{agent_id}/policy
   */
  getPolicy: (agentId: string) =>
    apiClient.get<{ ok: boolean; policy: BehavioralPolicy }>(
      `/v1/agent/${agentId}/policy`
    ),

  /**
   * Update agent behavioral policy.
   * PUT /v1/agent/{agent_id}/policy
   */
  updatePolicy: (agentId: string, policy: Partial<BehavioralPolicy>) =>
    apiClient.put<{ ok: boolean; policy: BehavioralPolicy }>(
      `/v1/agent/${agentId}/policy`,
      policy
    ),

  // ---------------------------------------------------------------------------
  // Execution History
  // ---------------------------------------------------------------------------

  /**
   * List agent execution records.
   * GET /v1/agent/{agent_id}/executions
   */
  listExecutions: (agentId: string, params?: { limit?: number; offset?: number }) =>
    apiClient.get<ExecutionListResponse>(
      `/v1/agent/${agentId}/executions`,
      { params }
    ),

  /**
   * Get a specific execution record.
   * GET /v1/agent/{agent_id}/executions/{exec_id}
   */
  getExecution: (agentId: string, execId: string) =>
    apiClient.get<{ ok: boolean; execution: ExecutionRecord }>(
      `/v1/agent/${agentId}/executions/${execId}`
    ),

  // ---------------------------------------------------------------------------
  // Analytics
  // ---------------------------------------------------------------------------

  /**
   * Get agent analytics.
   * GET /v1/agent/{agent_id}/analytics
   */
  getAnalytics: (agentId: string, params?: { since?: string }) =>
    apiClient.get<{ ok: boolean; analytics: AgentAnalytics }>(
      `/v1/agent/${agentId}/analytics`,
      { params }
    ),

  // ---------------------------------------------------------------------------
  // Session Management
  // ---------------------------------------------------------------------------

  /**
   * Start an agent session.
   * POST /v1/agent/{agent_id}/session/start
   */
  startSession: (agentId: string, metadata?: Record<string, unknown>) =>
    apiClient.post<{ ok: boolean; session: AgentSession }>(
      `/v1/agent/${agentId}/session/start`,
      metadata
    ),

  /**
   * End an agent session.
   * POST /v1/agent/{agent_id}/session/{session_id}/end
   */
  endSession: (agentId: string, sessionId: string) =>
    apiClient.post<{ ok: boolean; session: AgentSession }>(
      `/v1/agent/${agentId}/session/${sessionId}/end`
    ),

  /**
   * Get session details.
   * GET /v1/agent/{agent_id}/session/{session_id}
   */
  getSession: (agentId: string, sessionId: string) =>
    apiClient.get<{ ok: boolean; session: AgentSession }>(
      `/v1/agent/${agentId}/session/${sessionId}`
    ),

  // ---------------------------------------------------------------------------
  // Billing & Credits
  // ---------------------------------------------------------------------------

  /**
   * Get billing summary.
   * GET /v1/agent/{agent_id}/billing/summary
   */
  getBillingSummary: (agentId: string) =>
    apiClient.get<{ ok: boolean; summary: BillingSummary }>(
      `/v1/agent/${agentId}/billing/summary`
    ),

  /**
   * Update spend cap.
   * PUT /v1/agent/{agent_id}/billing/spend-cap
   */
  updateSpendCap: (agentId: string, spendCap: number) =>
    apiClient.put<{ ok: boolean }>(
      `/v1/agent/${agentId}/billing/spend-cap`,
      { spend_cap: spendCap }
    ),

  /**
   * Get cost breakdown.
   * GET /v1/agent/{agent_id}/cost-breakdown
   */
  getCostBreakdown: (agentId: string) =>
    apiClient.get<{ ok: boolean; breakdown: CostBreakdown }>(
      `/v1/agent/${agentId}/cost-breakdown`
    ),

  /**
   * Get credit balance.
   * GET /v1/agent/{agent_id}/credits/balance
   */
  getCreditBalance: (agentId: string) =>
    apiClient.get<{ ok: boolean; balance: CreditBalance }>(
      `/v1/agent/${agentId}/credits/balance`
    ),

  /**
   * Purchase credits.
   * POST /v1/agent/{agent_id}/credits/purchase
   */
  purchaseCredits: (agentId: string, amount: number) =>
    apiClient.post<{ ok: boolean; balance: CreditBalance }>(
      `/v1/agent/${agentId}/credits/purchase`,
      { amount }
    ),

  // ---------------------------------------------------------------------------
  // Concurrency
  // ---------------------------------------------------------------------------

  /**
   * Get concurrency statistics.
   * GET /v1/agent/concurrency/stats
   */
  getConcurrencyStats: () =>
    apiClient.get<{ ok: boolean; stats: ConcurrencyStats }>(
      "/v1/agent/concurrency/stats"
    ),

  // ---------------------------------------------------------------------------
  // Discovery & Execution
  // ---------------------------------------------------------------------------

  /**
   * Discover available functions.
   * GET /v1/agent/discover
   */
  discover: (params?: { query?: string; limit?: number }) =>
    apiClient.get<{ ok: boolean; functions: unknown[] }>("/v1/agent/discover", {
      params,
    }),

  /**
   * Discover a specific function.
   * GET /v1/agent/discover/{author}/{name}
   */
  discoverFunction: (author: string, name: string) =>
    apiClient.get<{ ok: boolean; function: unknown }>(
      `/v1/agent/discover/${author}/${name}`
    ),

  /**
   * Execute a function using an agent.
   * POST /v1/agent/execute/{author}/{name}
   */
  execute: (
    author: string,
    name: string,
    data: { input?: unknown; version?: string; apiKey?: string }
  ) =>
    apiClient.post<{ ok: boolean; result: unknown; executionId: string }>(
      `/v1/agent/execute/${author}/${name}`,
      data
    ),

  // ---------------------------------------------------------------------------
  // Swarm Operations
  // ---------------------------------------------------------------------------

  /**
   * Spawn a child agent.
   * POST /v1/agent/{id}/spawn
   */
  spawnChild: (
    agentId: string,
    data: {
      child_agent_id: string;
      child_name: string;
      child_description?: string;
      swarm_role?: string;
      max_child_agents?: number;
      capabilities?: Record<string, unknown>;
      initial_budget_usd?: number;
    }
  ) =>
    apiClient.post<{ ok: boolean; agent: AgentIdentity; api_key: string }>(
      `/v1/agent/${agentId}/spawn`,
      data
    ),

  /**
   * Get child agents.
   * GET /v1/agent/{id}/children
   */
  getChildren: (agentId: string) =>
    apiClient.get<{ ok: boolean; children: ChildAgent[] }>(
      `/v1/agent/${agentId}/children`
    ),

  /**
   * Get parent agent.
   * GET /v1/agent/{id}/parent
   */
  getParent: (agentId: string) =>
    apiClient.get<{ ok: boolean; parent: AgentIdentity }>(
      `/v1/agent/${agentId}/parent`
    ),

  // ---------------------------------------------------------------------------
  // Messaging
  // ---------------------------------------------------------------------------

  /**
   * Send a message to another agent.
   * POST /v1/agent/{id}/message
   */
  sendMessage: (
    agentId: string,
    data: {
      to_agent_id: string;
      message_type: string;
      payload?: Record<string, unknown>;
    }
  ) =>
    apiClient.post<{ ok: boolean; message: string; msg_id: string }>(
      `/v1/agent/${agentId}/message`,
      data
    ),

  /**
   * Get agent inbox.
   * GET /v1/agent/{id}/inbox
   */
  getInbox: (agentId: string) =>
    apiClient.get<{ ok: boolean; messages: AgentMessage[] }>(
      `/v1/agent/${agentId}/inbox`
    ),

  // ---------------------------------------------------------------------------
  // Wallet
  // ---------------------------------------------------------------------------

  /**
   * Get agent wallet.
   * GET /v1/agent/{id}/wallet
   */
  getWallet: (agentId: string) =>
    apiClient.get<{ ok: boolean; wallet: AgentWallet }>(
      `/v1/agent/${agentId}/wallet`
    ),

  // ---------------------------------------------------------------------------
  // Marketplace
  // ---------------------------------------------------------------------------

  /**
   * Search marketplace agents.
   * GET /v1/marketplace/agents
   */
  searchMarketplaceAgents: (params?: MarketplaceAgentSearchParams) =>
    apiClient.get<{ ok: boolean; agents: MarketplaceAgent[]; total: number }>(
      "/v1/marketplace/agents",
      { params }
    ),

  /**
   * Create a marketplace listing for an agent.
   * POST /v1/marketplace/agent/list
   */
  createMarketplaceListing: (data: {
    agent_id: string;
    listing_type: string;
    pricing_model: string;
    price_per_call?: number;
    subscription_monthly_usd?: number;
    is_active: boolean;
  }) =>
    apiClient.post<{ ok: boolean; listing: MarketplaceAgent }>(
      "/v1/marketplace/agent/list",
      data
    ),

  // ---------------------------------------------------------------------------
  // Evolution
  // ---------------------------------------------------------------------------

  /**
   * Propose agent evolution.
   * POST /v1/agent/{id}/evolve
   */
  proposeEvolution: (agentId: string) =>
    apiClient.post<{
      ok: boolean;
      proposal: EvolutionProposal;
      analysis: unknown;
    }>(`/v1/agent/${agentId}/evolve`),

  // ---------------------------------------------------------------------------
  // Schedules
  // ---------------------------------------------------------------------------

  /**
   * Create an autonomy schedule.
   * POST /v1/agent/{id}/schedule
   */
  createSchedule: (
    agentId: string,
    data: {
      schedule_type: string;
      action_type: string;
      cron_expression?: string;
      next_run_at?: string;
      action_payload?: Record<string, unknown>;
    }
  ) =>
    apiClient.post<{ ok: boolean; schedule: AutonomySchedule }>(
      `/v1/agent/${agentId}/schedule`,
      data
    ),

  /**
   * Get agent schedules.
   * GET /v1/agent/{id}/schedules
   */
  getSchedules: (agentId: string) =>
    apiClient.get<{ ok: boolean; schedules: AutonomySchedule[] }>(
      `/v1/agent/${agentId}/schedules`
    ),
};
