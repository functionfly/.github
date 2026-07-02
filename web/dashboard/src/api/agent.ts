import { apiClient } from './client';

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
  model?: string;
  thinking_mode?: string;
  thinking_budget?: number;
  createdAt: string;
  updatedAt: string;
  parentAgentId?: string;
  swarmRole?: string;
  maxChildAgents?: number;
  capabilities?: Record<string, unknown>;
  autonomousEnabled?: boolean;
  evolutionEnabled?: boolean;
  is_daemon_running?: boolean;
  daemon_started_at?: string;
  daemon_execution_count?: number;
  daemon_config?: Record<string, unknown>;
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

/**
 * Go handlers encode agents with snake_case JSON (`agent_id`, `tenant_id`, …).
 * The dashboard uses camelCase AgentIdentity; normalize after every fetch.
 */
export function normalizeAgentIdentity(raw: unknown): AgentIdentity {
  const r = raw as Record<string, unknown>;
  return {
    id: String(r.id ?? ''),
    agentId: String(r.agentId ?? r.agent_id ?? ''),
    name: String(r.name ?? ''),
    description: r.description != null ? String(r.description) : undefined,
    tenantId: String(r.tenantId ?? r.tenant_id ?? ''),
    status: String(r.status ?? ''),
    createdAt: String(r.createdAt ?? r.created_at ?? ''),
    updatedAt: String(r.updatedAt ?? r.updated_at ?? ''),
    parentAgentId:
      r.parentAgentId != null
        ? String(r.parentAgentId)
        : r.parent_agent_id != null
          ? String(r.parent_agent_id)
          : undefined,
    swarmRole:
      r.swarmRole != null
        ? String(r.swarmRole)
        : r.swarm_role != null
          ? String(r.swarm_role)
          : undefined,
    maxChildAgents:
      typeof r.maxChildAgents === 'number'
        ? r.maxChildAgents
        : typeof r.max_child_agents === 'number'
          ? r.max_child_agents
          : undefined,
    capabilities: r.capabilities as Record<string, unknown> | undefined,
    autonomousEnabled:
      typeof r.autonomousEnabled === 'boolean'
        ? r.autonomousEnabled
        : typeof r.autonomous_enabled === 'boolean'
          ? r.autonomous_enabled
          : undefined,
    evolutionEnabled:
      typeof r.evolutionEnabled === 'boolean'
        ? r.evolutionEnabled
        : typeof r.evolution_enabled === 'boolean'
          ? r.evolution_enabled
          : undefined,
    model: r.model != null ? String(r.model) : undefined,
    thinking_mode: r.thinking_mode != null ? String(r.thinking_mode) : r.thinkingMode != null ? String(r.thinkingMode) : undefined,
    thinking_budget: typeof r.thinking_budget === 'number' ? r.thinking_budget : typeof r.thinkingBudget === 'number' ? r.thinkingBudget : undefined,
    is_daemon_running: typeof r.is_daemon_running === 'boolean' ? r.is_daemon_running : typeof r.isDaemonRunning === 'boolean' ? r.isDaemonRunning : undefined,
    daemon_started_at: r.daemon_started_at != null ? String(r.daemon_started_at) : r.daemonStartedAt != null ? String(r.daemonStartedAt) : undefined,
    daemon_execution_count: typeof r.daemon_execution_count === 'number' ? r.daemon_execution_count : typeof r.daemonExecutionCount === 'number' ? r.daemonExecutionCount : undefined,
    daemon_config: r.daemon_config as Record<string, unknown> | undefined ?? r.daemonConfig as Record<string, unknown> | undefined,
  };
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
  functionUri?: string;
  outcome: string;
  latencyMs: number;
  costUsd: number;
  timestamp: string;
  errorCode?: string;
  model_name?: string;
  provider?: string;
  prompt_tokens?: number;
  completion_tokens?: number;
  total_tokens?: number;
  reasoning_tokens?: number;
}

export interface ExecutionListResponse {
  ok: boolean;
  executions: ExecutionRecord[];
  total: number;
  limit: number;
  offset: number;
}

export interface SEBGModificationProposal {
  id: string;
  graph_id: string;
  change_type: 'add_node' | 'remove_node' | 'rewire_edge' | 'add_specialist' | 'optimize';
  target_node_id: string;
  target_node_name: string;
  expected_revenue_lift: number;
  expected_lift_pct: number;
  risk_score: number;
  status: 'pending' | 'approved' | 'rejected' | 'applied' | 'expired';
  approved_by?: string;
  created_at: string;
  updated_at?: string;
}

export interface AgentAnalytics {
  agent_id: string;
  total_calls: number;
  total_cost_usd: number;
  avg_latency_ms: number;
  p50_latency_ms: number;
  p95_latency_ms: number;
  success_count: number;
  error_count: number;
  timeout_count: number;
  policy_violation_count: number;
  success_rate: number;
  total_prompt_tokens: number;
  total_completion_tokens: number;
  total_all_tokens: number;
  total_reasoning_tokens: number;
}

export interface ModelBreakdownItem {
  model_name: string;
  provider: string;
  total_calls: number;
  total_cost_usd: number;
  avg_latency_ms: number;
  total_prompt_tokens: number;
  total_completion_tokens: number;
  total_tokens: number;
  total_reasoning_tokens: number;
  avg_tokens_per_call: number;
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
  agentId: string;
  name: string;
  status: 'active' | 'suspended' | 'pending';
  swarmRole: 'worker' | 'manager' | 'infrastructure';
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
  agent_id: string;
  balance_usd: number;
  escrow_balance_usd: number;
  total_earned_usd: number;
  total_spent_usd: number;
  last_earning_at?: string;
  last_spending_at?: string;
}

export interface AgentFinancialTransaction {
  id: string;
  tenant_id: string;
  agent_id: string;
  kind:
    | 'credit_purchase'
    | 'execution_debit'
    | 'transfer_in'
    | 'transfer_out'
    | 'adjustment'
    | 'refund';
  amount_usd: number;
  status: 'pending' | 'completed' | 'failed';
  provider?: string;
  provider_ref?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
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
  listingType: 'worker' | 'manager' | 'infrastructure';
  pricingModel: 'free' | 'per_call' | 'subscription' | 'revenue_share' | 'tiered' | 'dynamic' | 'auction';
  pricePerCall?: number;
  subscriptionMonthlyUsd?: number;
  revenueSharePercent?: number;
  ratingScore: number;
  totalCalls: number;
  roiScore: number;
  trustScore?: number;
  deterministicVerified?: boolean;
  capabilities?: string[];
  status?: string;
  isOfficial?: boolean;
  rankScore?: number;
  walletBalanceUsd?: number;
  hiringHistoryCount?: number;
}

export interface MarketplaceAgentSearchParams {
  agent_id?: string;
  pricing_model?: string;
  min_rating?: number;
  max_price_per_call?: number;
  min_roi_score?: number;
  listing_types?: string[];
  capabilities?: string[];
  sort_by?: 'rank_score' | 'rating_score' | 'price_per_call' | 'total_calls';
  limit?: number;
  offset?: number;
}

export interface EvolutionProposal {
  id: string;
  agentId: string;
  proposalType: string;
  status: 'pending' | 'approved' | 'rejected' | 'implemented';
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
   * Sends snake_case (agent_id) for API compatibility.
   */
  registerAgent: async (data: RegisterAgentRequest) => {
    const res = await apiClient.post<RegisterAgentResponse>('/v1/agent/register', {
      agent_id: data.agentId,
      name: data.name,
      ...(data.description != null && data.description !== '' && { description: data.description }),
    });
    return { ...res, agent: normalizeAgentIdentity(res.agent) };
  },

  /**
   * List all agents for the authenticated tenant.
   * GET /v1/agent
   */
  listAgents: async (params?: { limit?: number; offset?: number }) => {
    const res = await apiClient.get<AgentListResponse>('/v1/agent', { params });
    return {
      ...res,
      agents: (res.agents ?? []).map((a) => normalizeAgentIdentity(a)),
    };
  },

  /**
   * Get an agent by ID.
   * GET /v1/agent/{agent_id}
   */
  getAgent: async (agentId: string) => {
    const res = await apiClient.get<{ ok: boolean; agent: unknown }>(`/v1/agent/${agentId}`);
    return { ...res, agent: normalizeAgentIdentity(res.agent) };
  },

  /**
   * Delete an agent.
   * DELETE /v1/agent/{agent_id}
   */
  deleteAgent: (agentId: string) =>
    apiClient.delete<{ ok: boolean; message: string }>(`/v1/agent/${agentId}`),

  /**
   * Update an agent's mutable fields (name, description, capabilities, etc.).
   * PUT /v1/agent/{agent_id}
   */
  updateAgent: (
    agentId: string,
    data: {
      name?: string;
      description?: string;
      capabilities?: Record<string, unknown>;
      autonomous_enabled?: boolean;
      evolution_enabled?: boolean;
      model?: string;
    }
  ) =>
    apiClient.put<{ ok: boolean; agent: AgentIdentity }>(`/v1/agent/${agentId}`, data).then((res) => ({ ...res, agent: normalizeAgentIdentity(res.agent) })),

  // ---------------------------------------------------------------------------
  // Quota Management
  // ---------------------------------------------------------------------------

  /**
   * Update agent quota configuration.
   * PUT /v1/agent/{agent_id}/quota
   */
  updateQuota: (agentId: string, quota: AgentQuota) =>
    apiClient.put<{ ok: boolean; quota: AgentQuota }>(`/v1/agent/${agentId}/quota`, quota),

  /**
   * Get current agent usage.
   * GET /v1/agent/{agent_id}/usage
   */
  getUsage: (agentId: string) =>
    apiClient.get<{ ok: boolean; usage: AgentUsage }>(`/v1/agent/${agentId}/usage`),

  // ---------------------------------------------------------------------------
  // Policy Management
  // ---------------------------------------------------------------------------

  /**
   * Get agent behavioral policy.
   * GET /v1/agent/{agent_id}/policy
   */
  getPolicy: (agentId: string) =>
    apiClient.get<{ ok: boolean; policy: BehavioralPolicy }>(`/v1/agent/${agentId}/policy`),

  /**
   * Update agent behavioral policy.
   * PUT /v1/agent/{agent_id}/policy
   */
  updatePolicy: (agentId: string, policy: Partial<BehavioralPolicy>) =>
    apiClient.put<{ ok: boolean; policy: BehavioralPolicy }>(`/v1/agent/${agentId}/policy`, policy),

  // ---------------------------------------------------------------------------
  // Execution History
  // ---------------------------------------------------------------------------

  /**
   * List agent execution records.
   * GET /v1/agent/{agent_id}/executions
   */
  listExecutions: (agentId: string, params?: { limit?: number; offset?: number }) =>
    apiClient.get<ExecutionListResponse>(`/v1/agent/${agentId}/executions`, { params }),

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
    apiClient.get<{ ok: boolean; analytics: AgentAnalytics }>(`/v1/agent/${agentId}/analytics`, {
      params,
    }),

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
    apiClient.get<{ ok: boolean; summary: BillingSummary }>(`/v1/agent/${agentId}/billing/summary`),

  /**
   * Update spend cap.
   * PUT /v1/agent/{agent_id}/billing/spend-cap
   */
  updateSpendCap: (agentId: string, caps: { spend_cap_daily_usd?: number; spend_cap_weekly_usd?: number; spend_cap_monthly_usd?: number }) =>
    apiClient.put<{ ok: boolean }>(`/v1/agent/${agentId}/billing/spend-cap`, caps),

  /**
   * Get cost breakdown.
   * GET /v1/agent/{agent_id}/cost-breakdown
   */
  getCostBreakdown: (agentId: string) =>
    apiClient.get<{ ok: boolean; breakdown: CostBreakdown }>(`/v1/agent/${agentId}/cost-breakdown`),

  /**
   * Get cost and token breakdown by model.
   * GET /v1/agent/{agent_id}/model-breakdown
   */
  getModelBreakdown: (agentId: string) =>
    apiClient.get<{ ok: boolean; models: ModelBreakdownItem[] }>(`/v1/agent/${agentId}/model-breakdown`),

  /**
   * Get credit balance.
   * GET /v1/agent/{agent_id}/credits/balance
   */
  getCreditBalance: (agentId: string) =>
    apiClient.get<{ ok: boolean; balance: CreditBalance }>(`/v1/agent/${agentId}/credits/balance`),

  /**
   * Purchase credits (direct charge - requires payment_method_id).
   * POST /v1/agent/{agent_id}/credits/purchase
   */
  purchaseCredits: (agentId: string, amountUsd: number, paymentMethodId?: string) =>
    apiClient.post<{
      ok: boolean;
      agent_id: string;
      credits_added_usd: number;
      new_balance_usd: number;
    }>(`/v1/agent/${agentId}/credits/purchase`, {
      amount_usd: amountUsd,
      payment_method_id: paymentMethodId,
    }),

  /**
   * Create Stripe Checkout session for purchasing credits.
   * POST /v1/agent/{agent_id}/credits/checkout
   */
  createCreditsCheckout: (
    agentId: string,
    amountUsd: number,
    successUrl?: string,
    cancelUrl?: string
  ) =>
    apiClient.post<{
      ok: boolean;
      session_id: string;
      url: string;
    }>(`/v1/agent/${agentId}/credits/checkout`, {
      amount_usd: amountUsd,
      success_url: successUrl,
      cancel_url: cancelUrl,
    }),

  /**
   * List wallet transactions (financial ledger).
   * GET /v1/agent/{agent_id}/transactions
   */
  listWalletTransactions: (agentId: string, params?: { limit?: number; offset?: number }) =>
    apiClient.get<{
      ok: boolean;
      agent_id: string;
      transactions: AgentFinancialTransaction[];
      total: number;
      limit: number;
      offset: number;
    }>(`/v1/agent/${agentId}/transactions`, { params }),

  /**
   * Export wallet transactions as CSV.
   * GET /v1/agent/{agent_id}/transactions?format=csv (handled via blob response)
   */
  exportWalletTransactions: (agentId: string) =>
    apiClient.get<string>(`/v1/agent/${agentId}/transactions`, {
      params: { format: 'csv', limit: 1000 },
      responseType: 'text',
    }),

  // ---------------------------------------------------------------------------
  // Concurrency
  // ---------------------------------------------------------------------------

  /**
   * Get concurrency statistics.
   * GET /v1/agent/concurrency/stats
   */
  getConcurrencyStats: () =>
    apiClient.get<{ ok: boolean; stats: ConcurrencyStats }>('/v1/agent/concurrency/stats'),

  // ---------------------------------------------------------------------------
  // Discovery & Execution
  // ---------------------------------------------------------------------------

  /**
   * Discover available functions.
   * GET /v1/agent/discover
   */
  discover: (params?: { query?: string; limit?: number }) =>
    apiClient.get<{ ok: boolean; functions: unknown[] }>('/v1/agent/discover', {
      params,
    }),

  /**
   * Discover a specific function.
   * GET /v1/agent/discover/{author}/{name}
   */
  discoverFunction: (author: string, name: string) =>
    apiClient.get<{ ok: boolean; function: unknown }>(`/v1/agent/discover/${author}/${name}`),

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
      child_agent_id?: string;
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
    apiClient.get<{ ok: boolean; children: ChildAgent[] }>(`/v1/agent/${agentId}/children`),

  /**
   * Get parent agent.
   * GET /v1/agent/{id}/parent
   */
  getParent: (agentId: string) =>
    apiClient.get<{ ok: boolean; parent: AgentIdentity }>(`/v1/agent/${agentId}/parent`),

  /**
   * Reassign an agent's swarm role.
   * PUT /v1/agent/{id}/role
   */
  reassignRole: (agentId: string, swarmRole: string) =>
    apiClient.put<{ ok: boolean; agent_id: string; swarm_role: string }>(
      `/v1/agent/${agentId}/role`,
      { swarm_role: swarmRole }
    ),

  /**
   * Update an agent's swarm topology.
   * PUT /v1/agent/{id}/topology
   */
  updateTopology: (agentId: string, topology: string) =>
    apiClient.put<{ ok: boolean; agent_id: string; topology: string }>(
      `/v1/agent/${agentId}/topology`,
      { topology }
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
    apiClient.get<{ ok: boolean; messages: AgentMessage[] }>(`/v1/agent/${agentId}/inbox`),

  // ---------------------------------------------------------------------------
  // Wallet
  // ---------------------------------------------------------------------------

  /**
   * Get agent wallet.
   * GET /v1/agent/{id}/wallet
   */
  getWallet: (agentId: string) =>
    apiClient.get<{ ok: boolean; wallet: AgentWallet }>(`/v1/agent/${agentId}/wallet`),

  // ---------------------------------------------------------------------------
  // Marketplace
  // ---------------------------------------------------------------------------

  /**
   * Search marketplace agents.
   * GET /v1/marketplace/agents
   */
  searchMarketplaceAgents: (params?: MarketplaceAgentSearchParams) =>
    apiClient.get<{
      ok: boolean;
      agents: MarketplaceAgent[];
      total: number;
      limit: number;
      offset: number;
      has_more: boolean;
    }>(
      '/v1/marketplace/agents',
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
    apiClient.post<{ ok: boolean; listing: MarketplaceAgent }>('/v1/marketplace/agent/list', data),

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
    apiClient.get<{ ok: boolean; schedules: AutonomySchedule[] }>(`/v1/agent/${agentId}/schedules`),

  /**
   * Get agent quota.
   * GET /v1/agent/{agent_id}/quota
   */
  getQuota: (agentId: string) =>
    apiClient.get<{ ok: boolean; quota: AgentQuota }>(`/v1/agent/${agentId}/quota`),

  // ---------------------------------------------------------------------------
  // Learning & Analysis (New)
  // ---------------------------------------------------------------------------

  /**
   * Analyze agent execution patterns.
   * GET /v1/agent/{id}/analyze
   */
  analyzeAgent: (agentId: string, params?: { days?: number }) =>
    apiClient.get<{
      ok: boolean;
      analysis: {
        agent_id: string;
        total_executions: number;
        patterns: Array<{
          id: string;
          pattern_type: string;
          confidence: number;
          recommendations: string[];
        }>;
        insights: string[];
        success_rate: number;
        avg_latency_ms: number;
        avg_cost_usd: number;
      };
    }>(`/v1/agent/${agentId}/analyze`, { params }),

  /**
   * Optimize agent - generate optimization recommendations.
   * POST /v1/agent/{id}/optimize
   */
  optimizeAgent: (agentId: string) =>
    apiClient.post<{
      ok: boolean;
      result: {
        agent_id: string;
        patterns_found: number;
        optimizations: Array<{
          id: string;
          optimization_type: string;
          description: string;
          expected_impact: Record<string, number>;
          implementation: 'low' | 'medium' | 'high';
          status: 'pending' | 'approved' | 'rejected' | 'applied';
        }>;
      };
    }>(`/v1/agent/${agentId}/optimize`),

  /**
   * Get agent insights (patterns + optimizations + memories).
   * GET /v1/agent/{id}/insights
   */
  getInsights: (agentId: string) =>
    apiClient.get<{
      ok: boolean;
      patterns: unknown[];
      optimizations: unknown[];
      memories: unknown[];
      memory_count: number;
    }>(`/v1/agent/${agentId}/insights`),

  /**
   * Search agent memories.
   * GET /v1/agent/{id}/memories
   */
  searchMemories: (agentId: string, params?: { q?: string; limit?: number }) =>
    apiClient.get<{
      ok: boolean;
      query: string;
      memories: Array<{
        id: string;
        agent_id: string;
        memory_type: string;
        content: Record<string, unknown>;
        importance: number;
        is_learned: boolean;
        created_at: string;
      }>;
      count: number;
    }>(`/v1/agent/${agentId}/memories`, { params }),

  // ---------------------------------------------------------------------------
  // Code Generation & Deployment (New)
  // ---------------------------------------------------------------------------

  /**
   * Generate code from specification.
   * POST /v1/agent/{id}/generate
   */
  generateCode: (
    agentId: string,
    data: {
      function_spec: {
        name: string;
        title?: string;
        description: string;
        prompt?: string;
        input_schema?: Record<string, unknown>;
        output_schema?: Record<string, unknown>;
        category?: string;
        tags?: string[];
        examples?: Record<string, unknown>[];
      };
      language: 'python' | 'javascript';
      runtime: string;
    }
  ) =>
    apiClient.post<{
      ok: boolean;
      code: {
        id: string;
        generated_code: string;
        language: string;
        runtime: string;
        model_used: string;
        generation_time_ms: number;
        status: string;
        created_at: string;
      };
    }>(`/v1/agent/${agentId}/generate`, data),

  /**
   * Get generated code history.
   * GET /v1/agent/{id}/generations
   */
  getGenerations: (agentId: string, params?: { limit?: number; offset?: number }) =>
    apiClient.get<{
      ok: boolean;
      generations: unknown[];
      total: number;
      limit: number;
      offset: number;
    }>(`/v1/agent/${agentId}/generations`, { params }),

  /**
   * Publish generated function to registry.
   * POST /v1/agent/{id}/publish
   */
  publishFunction: (
    agentId: string,
    data: {
      generated_code_id: string;
      author: string;
      name: string;
      title?: string;
      description?: string;
      category?: string;
      tags?: string[];
      is_public?: boolean;
    }
  ) =>
    apiClient.post<{
      ok: boolean;
      published: {
        id: string;
        agent_id: string;
        function_id: string;
        author: string;
        name: string;
        version: string;
        status: string;
        published_at?: string;
      };
    }>(`/v1/agent/${agentId}/publish`, data),

  /**
   * Get published functions for an agent.
   * GET /v1/agent/{id}/functions
   */
  getAgentFunctions: (agentId: string, params?: { limit?: number; offset?: number }) =>
    apiClient.get<{
      ok: boolean;
      functions: unknown[];
      total: number;
      limit: number;
      offset: number;
    }>(`/v1/agent/${agentId}/functions`, { params }),

  // ---------------------------------------------------------------------------
  // Security & Health (New)
  // ---------------------------------------------------------------------------

  /**
   * Trigger kill switch on agent.
   * POST /v1/agent/{id}/kill-switch
   */
  triggerKillSwitch: (agentId: string, reason?: string) =>
    apiClient.post<{
      ok: boolean;
      agents_killed: number;
      message: string;
    }>(`/v1/agent/${agentId}/kill-switch`, { reason }),

  /**
   * Check swarm health and detect anomalies.
   * GET /v1/agent/{id}/health
   */
  checkSwarmHealth: (agentId: string, params?: { hours?: number }) =>
    apiClient.get<{
      ok: boolean;
      status: 'healthy' | 'degraded' | 'critical';
      health_score: number;
      anomalies: Array<{
        type: string;
        severity: 'low' | 'medium' | 'high';
        description: string;
        timestamp: string;
      }>;
      children: number;
    }>(`/v1/agent/${agentId}/health`, { params }),

  /**
   * Get swarm statistics.
   * GET /v1/agent/{id}/swarm/stats
   */
  getSwarmStats: (agentId: string) =>
    apiClient.get<{
      ok: boolean;
      stats: SwarmStats;
    }>(`/v1/agent/${agentId}/swarm/stats`),

  /**
   * Get security alerts for an agent.
   * GET /v1/agent/{id}/security/alerts
   */
  getSecurityAlerts: (agentId: string, filters?: { severity?: string; status?: string }) =>
    apiClient.get<{
      ok: boolean;
      alerts: Array<{
        id: string;
        agent_id: string;
        alert_type: string;
        severity: 'low' | 'medium' | 'high' | 'critical';
        status: 'active' | 'acknowledged' | 'resolved';
        message: string;
        created_at: string;
        acknowledged_at?: string;
        resolved_at?: string;
      }>;
      total: number;
    }>(`/v1/agent/${agentId}/security/alerts`, { params: filters }),

  /**
   * Acknowledge a security alert.
   * POST /v1/agent/{id}/security/alerts/{alert_id}/acknowledge
   */
  acknowledgeSecurityAlert: (agentId: string, alertId: string) =>
    apiClient.post<{ ok: boolean; message: string }>(
      `/v1/agent/${agentId}/security/alerts/${alertId}/acknowledge`
    ),

  /**
   * Resolve a security alert.
   * POST /v1/agent/{id}/security/alerts/{alert_id}/resolve
   */
  resolveSecurityAlert: (agentId: string, alertId: string, resolution?: string) =>
    apiClient.post<{ ok: boolean; message: string }>(
      `/v1/agent/${agentId}/security/alerts/${alertId}/resolve`,
      resolution ? { resolution } : {}
    ),

  // ---------------------------------------------------------------------------
  // Marketplace Hiring & Purchasing (New)
  // ---------------------------------------------------------------------------

  /**
   * Hire an agent for a task.
   * POST /v1/marketplace/hire
   */
  hireAgent: (data: {
    agent_id: string;
    task_type: string;
    task_payload?: Record<string, unknown>;
    budget_usd?: number;
  }) =>
    apiClient.post<{
      ok: boolean;
      hiring: {
        id: string;
        agent_id: string;
        hirer_id: string;
        task_type: string;
        budget_usd: number;
        status: string;
        created_at: string;
      };
      message: string;
    }>('/v1/marketplace/hire', data),

  /**
   * Purchase a function from marketplace.
   * POST /v1/marketplace/purchase
   */
  purchaseFunction: (data: {
    function_author: string;
    function_name: string;
    agent_id: string;
    max_price_usd?: number;
  }) =>
    apiClient.post<{
      ok: boolean;
      purchase: {
        id: string;
        agent_id: string;
        function_author: string;
        function_name: string;
        price_paid_usd: number;
        status: string;
        created_at: string;
      };
    }>('/v1/marketplace/purchase', data),

  // ---------------------------------------------------------------------------
  // SEBG — Self-Evolving Backend Graph
  // ---------------------------------------------------------------------------

  /**
   * Get SEBG tenant configuration.
   * GET /v1/agent/{agent_id}/sebg/config
   */
  getSEBGConfig: (agentId: string) =>
    apiClient.get<{
      ok: boolean;
      config: {
        id: string;
        tenant_id: string;
        autonomy_tier: 'manual' | 'assisted' | 'fully_autonomous';
        revenue_share_fee_pct: number;
        max_risk_score_auto_apply: number;
        is_active: boolean;
        created_at: string;
        updated_at: string;
      };
    }>(`/v1/agent/${agentId}/sebg/config`),

  /**
   * Update SEBG autonomy tier.
   * PUT /v1/agent/{agent_id}/sebg/tier
   */
  updateSEBGTier: (
    agentId: string,
    tier: 'manual' | 'assisted' | 'fully_autonomous'
  ) =>
    apiClient.put<{
      ok: boolean;
      config: {
        autonomy_tier: string;
        max_risk_score_auto_apply: number;
      };
    }>(`/v1/agent/${agentId}/sebg/tier`, { tier }),

  /**
   * List SEBG modification proposals for an agent.
   * GET /v1/agent/{agent_id}/sebg/proposals
   */
  listSEBGProposals: (
    agentId: string,
    params?: { status?: string; limit?: number; offset?: number }
  ) =>
    apiClient.get<{
      ok: boolean;
      proposals: SEBGModificationProposal[];
      total: number;
      limit: number;
      offset: number;
    }>(`/v1/agent/${agentId}/sebg/proposals`, { params }),

  /**
   * Approve or reject a SEBG modification proposal.
   * POST /v1/agent/{agent_id}/sebg/proposals/{proposal_id}/decide
   */
  decideSEBGProposal: (
    agentId: string,
    proposalId: string,
    decision: 'approved' | 'rejected'
  ) =>
    apiClient.post<{
      ok: boolean;
      proposal_id: string;
      decision: string;
    }>(`/v1/agent/${agentId}/sebg/proposals/${proposalId}/decide`, {
      decision,
    }),

  /**
   * Trigger the full SEBG evolve pipeline.
   * POST /v1/agent/{agent_id}/sebg/evolve
   */
  triggerSEBGEvolve: (agentId: string) =>
    apiClient.post<{
      ok: boolean;
      result: {
        status: string;
        agent_id: string;
        graphs_analyzed: number;
        applied: number;
        pending: number;
      };
    }>(`/v1/agent/${agentId}/sebg/evolve`),

  /**
   * Get SEBG ROI summary.
   * GET /v1/agent/{agent_id}/sebg/roi
   */
  getSEBGROI: (agentId: string) =>
    apiClient.get<{
      ok: boolean;
      roi: {
        tenant_id: string;
        applied_count: number;
        pending_count: number;
        revenue_lift_cents: number;
      };
    }>(`/v1/agent/${agentId}/sebg/roi`),

  // ---------------------------------------------------------------------------
  // AI Models
  // ---------------------------------------------------------------------------

  /**
   * Get the curated list of AI models available for agents.
   * GET /v1/agent/models
   */
  getModels: () =>
    apiClient.get<{ ok: boolean; models: Array<{ id: string; name: string; provider: string; provider_label?: string; key_source?: string; tier: string; cost: string }> }>('/v1/ai/models'),

  // ---------------------------------------------------------------------------
  // Agent Console (Chat)
  // ---------------------------------------------------------------------------

  /**
   * Send a message to an agent via FlyMind using the agent's configured model.
   * POST /v1/agent/{agent_id}/chat
   */
  agentChat: (agentId: string, message: string, sessionId?: string) =>
    apiClient.post<{ ok: boolean; message: string; model: string; thinking?: { content: string; tokens: number }; session_id?: string }>(`/v1/agent/${agentId}/chat`, { message, session_id: sessionId }),

  getChatHistory: (agentId: string, limit = 50, offset = 0, sessionId?: string) => {
    const params = new URLSearchParams({ limit: String(limit), offset: String(offset) });
    if (sessionId) params.set('session_id', sessionId);
    return apiClient.get<{ ok: boolean; messages: Array<{ role: string; content: string; model?: string; metadata?: { thinking?: { content: string; tokens: number } }; created_at: string }> }>(`/v1/agent/${agentId}/chat/history?${params}`);
  },

  clearChat: (agentId: string, sessionId?: string) => {
    const params = sessionId ? `?session_id=${sessionId}` : '';
    return apiClient.delete<{ ok: boolean }>(`/v1/agent/${agentId}/chat${params}`);
  },

  // Chat Sessions
  listChatSessions: (agentId: string, limit = 50, offset = 0) =>
    apiClient.get<{ ok: boolean; sessions: Array<{ id: string; title: string; agent_id: string; message_count: number; last_message_at?: string; created_at: string }> }>(`/v1/agent/${agentId}/chat/sessions?limit=${limit}&offset=${offset}`),

  createChatSession: (agentId: string, title?: string) =>
    apiClient.post<{ ok: boolean; session_id: string; title: string }>(`/v1/agent/${agentId}/chat/sessions`, { title }),

  deleteChatSession: (agentId: string, sessionId: string) =>
    apiClient.delete<{ ok: boolean }>(`/v1/agent/${agentId}/chat/sessions/${sessionId}`),

  // ---------------------------------------------------------------------------
  // Agent Workspace
  // ---------------------------------------------------------------------------

  /** Browse workspace files. GET /v1/agent/{agent_id}/workspace?path= */
  browseWorkspace: (agentId: string, path = '') =>
    apiClient.get<{ ok: boolean; path: string; workspace: string; entries: Array<{ name: string; path: string; is_dir: boolean; size: number; mod_time: string }> }>(`/v1/agent/${agentId}/workspace?path=${encodeURIComponent(path)}`),

  /** Get workspace manifest. GET /v1/agent/{agent_id}/workspace/manifest */
  getWorkspaceManifest: (agentId: string) =>
    apiClient.get<{ ok: boolean; manifest: { agent_id: string; tenant_id: string; name: string; description: string; created_at: string; updated_at: string; file_count: number; total_bytes: number; structure: string[] } }>(`/v1/agent/${agentId}/workspace/manifest`),

  /** Get workspace execution history. GET /v1/agent/{agent_id}/workspace/history */
  getWorkspaceHistory: (agentId: string) =>
    apiClient.get<{ ok: boolean; history: Array<{ timestamp: string; action: string; tool: string; path?: string; description: string; agent_id: string; session_id?: string; result?: string }> }>(`/v1/agent/${agentId}/workspace/history`),
};
