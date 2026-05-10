/** Function DNA — types for the living code evolution system */

// ──────────────────────────────────────────────────────────────────────────────
// Core DNA Profile
// ──────────────────────────────────────────────────────────────────────────────

export interface InputPattern {
  shape: string;
  hash: string;
  frequency: number;
  count: number;
}

export interface BottleneckEntry {
  type: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  frequency: number;
}

export interface DNAProfile {
  id: string;
  function_id: string;
  function_type: 'registry' | 'managed';
  tenant_id: string;
  generation: number;
  fitness_score: number;
  total_executions: number;
  total_mutations: number;
  avg_latency_ms: number | null;
  p99_latency_ms: number | null;
  success_rate: number;
  error_distribution: Record<string, number>;
  input_patterns: InputPattern[];
  bottleneck_signature: BottleneckEntry[];
  dna_hash: string | null;
  evolution_enabled: boolean;
  last_analyzed_at: string | null;
  created_at: string;
  updated_at: string;
}

// ──────────────────────────────────────────────────────────────────────────────
// Mutations
// ──────────────────────────────────────────────────────────────────────────────

export type MutationType =
  | 'optimize_latency'
  | 'reduce_memory'
  | 'fix_error_pattern'
  | 'improve_reliability'
  | 'refactor_hotpath';

export type MutationStatus =
  | 'proposed'
  | 'accepted'
  | 'rejected'
  | 'deploying'
  | 'deployed'
  | 'rolled_back';

export interface MutationImpact {
  latency_improvement_pct: number;
  memory_reduction_pct: number;
  reliability_improvement_pct: number;
}

export interface DNAMutation {
  id: string;
  function_id: string;
  function_type: string;
  tenant_id: string;
  generation: number;
  mutation_type: MutationType;
  status: MutationStatus;
  trigger_reason: string | null;
  original_code: string | null;
  mutated_code: string | null;
  original_hash: string | null;
  mutated_hash: string | null;
  diff: string | null;
  estimated_impact: MutationImpact;
  actual_impact: MutationImpact | null;
  confidence: number | null;
  model_used: string | null;
  analysis_window_hours: number | null;
  executions_analyzed: number | null;
  accepted_by: string | null;
  accepted_at: string | null;
  deployed_at: string | null;
  rolled_back_at: string | null;
  rejected_reason: string | null;
  created_at: string;
}

export interface MutationListResponse {
  mutations: DNAMutation[];
  total: number;
  limit: number;
  offset: number;
}

// ──────────────────────────────────────────────────────────────────────────────
// Insights
// ──────────────────────────────────────────────────────────────────────────────

export interface AggregatedMetrics {
  total_executions: number;
  avg_latency_ms: number;
  p50_latency_ms: number;
  p95_latency_ms: number;
  p99_latency_ms: number;
  success_rate: number;
  error_distribution: Record<string, number>;
  cold_start_rate: number;
  avg_memory_peak_mb: number;
}

export interface MutationOutcomes {
  total: number;
  outcomes: {
    accepted: number;
    rejected: number;
    proposed: number;
    deployed: number;
    rolled_back: number;
  };
}

export interface DNAInsights {
  function_id: string;
  period: string;
  metrics: AggregatedMetrics;
  mutation_outcomes: MutationOutcomes;
}

export interface EnterpriseInsights {
  total_functions_analyzed: number;
  total_mutations_proposed: number;
  total_mutations_accepted: number;
  avg_fitness_score: number;
  avg_latency_improvement_pct: number;
  total_cost_savings_usd: number;
  top_bottleneck_categories: { category: string; count: number }[];
  evolution_leaderboard: {
    function_id: string;
    generation: number;
    fitness_score: number;
  }[];
}

// ──────────────────────────────────────────────────────────────────────────────
// Mutation Type Metadata
// ──────────────────────────────────────────────────────────────────────────────

export const MUTATION_TYPE_META: Record<
  MutationType,
  { label: string; icon: string; color: string; description: string }
> = {
  optimize_latency: {
    label: 'Latency Optimization',
    icon: 'Zap',
    color: 'text-velocity-500',
    description: 'Reduces execution time by optimizing hot paths',
  },
  reduce_memory: {
    label: 'Memory Reduction',
    icon: 'Cpu',
    color: 'text-info',
    description: 'Lowers memory footprint and cold start overhead',
  },
  fix_error_pattern: {
    label: 'Error Pattern Fix',
    icon: 'ShieldAlert',
    color: 'text-error',
    description: 'Addresses recurring error patterns detected in production',
  },
  improve_reliability: {
    label: 'Reliability Improvement',
    icon: 'Shield',
    color: 'text-success',
    description: 'Increases success rate and handles edge cases',
  },
  refactor_hotpath: {
    label: 'Hot Path Refactor',
    icon: 'GitBranch',
    color: 'text-warning',
    description: 'Restructures frequently-executed code paths',
  },
};

export const MUTATION_STATUS_META: Record<
  MutationStatus,
  { label: string; variant: 'default' | 'secondary' | 'success' | 'warning' | 'error' | 'outline' }
> = {
  proposed: { label: 'Proposed', variant: 'default' },
  accepted: { label: 'Accepted', variant: 'success' },
  rejected: { label: 'Rejected', variant: 'error' },
  deploying: { label: 'Deploying', variant: 'warning' },
  deployed: { label: 'Deployed', variant: 'success' },
  rolled_back: { label: 'Rolled Back', variant: 'error' },
};
