import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { cn } from '@/lib/utils';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { StatCard, fitnessColor, formatNumber, fitnessBg } from './DNAHelix';
import type { EnterpriseInsights, DNAInsights, AggregatedMetrics } from '@/types/dna';
import {
  Dna,
  TrendingUp,
  DollarSign,
  Activity,
  BarChart3,
  Trophy,
  Zap,
  AlertTriangle,
} from 'lucide-react';

// ──────────────────────────────────────────────────────────────────────────────
// Period Selector
// ──────────────────────────────────────────────────────────────────────────────

const PERIODS = [
  { value: '7d', label: '7 Days' },
  { value: '30d', label: '30 Days' },
  { value: '90d', label: '90 Days' },
] as const;

function PeriodSelector({
  value,
  onChange,
}: {
  value: string;
  onChange: (v: string) => void;
}) {
  return (
    <div className="flex items-center gap-1 rounded-lg bg-bg-tertiary p-0.5">
      {PERIODS.map((p) => (
        <button
          key={p.value}
          onClick={() => onChange(p.value)}
          className={cn(
            'px-3 py-1 text-xs font-medium rounded-md transition-all',
            value === p.value
              ? 'bg-card text-text-primary shadow-sm'
              : 'text-text-muted hover:text-text-secondary'
          )}
        >
          {p.label}
        </button>
      ))}
    </div>
  );
}

// ──────────────────────────────────────────────────────────────────────────────
// Bottleneck Chart — horizontal bar chart
// ──────────────────────────────────────────────────────────────────────────────

function BottleneckChart({
  data,
}: {
  data: { category: string; count: number }[];
}) {
  if (!data || data.length === 0) return null;
  const max = Math.max(...data.map((d) => d.count));

  const categoryIcons: Record<string, string> = {
    cold_start: '🥶',
    db_query: '🗄️',
    memory_allocation: '🧠',
    network: '🌐',
    timeout: '⏱️',
    runtime: '⚙️',
  };

  return (
    <div className="space-y-2">
      {data.map((item, i) => (
        <motion.div
          key={item.category}
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: i * 0.05 }}
          className="flex items-center gap-3"
        >
          <span className="text-sm w-6 text-center">
            {categoryIcons[item.category] || '📊'}
          </span>
          <span className="text-xs text-text-secondary w-28 truncate capitalize">
            {item.category.replace(/_/g, ' ')}
          </span>
          <div className="flex-1 h-5 rounded-full bg-bg-tertiary overflow-hidden">
            <motion.div
              className={cn('h-full rounded-full', fitnessBg(Math.max(20, 100 - (item.count / max) * 80)))}
              initial={{ width: 0 }}
              animate={{ width: `${(item.count / max) * 100}%` }}
              transition={{ duration: 0.6, ease: 'easeOut', delay: i * 0.05 }}
            />
          </div>
          <span className="text-xs font-mono text-text-muted w-8 text-right">
            {item.count}
          </span>
        </motion.div>
      ))}
    </div>
  );
}

// ──────────────────────────────────────────────────────────────────────────────
// Evolution Leaderboard
// ──────────────────────────────────────────────────────────────────────────────

function EvolutionLeaderboard({
  data,
}: {
  data: { function_id: string; generation: number; fitness_score: number }[];
}) {
  if (!data || data.length === 0) return null;

  const medals = ['🥇', '🥈', '🥉'];

  return (
    <div className="space-y-2">
      {data.map((item, i) => (
        <motion.div
          key={item.function_id}
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: i * 0.04 }}
          className={cn(
            'flex items-center gap-3 rounded-lg px-3 py-2 transition-colors',
            i < 3 ? 'bg-velocity-500/5 border border-velocity-500/10' : 'hover:bg-bg-hover'
          )}
        >
          <span className="text-sm w-6 text-center">
            {medals[i] || <span className="text-xs text-text-muted font-mono">{i + 1}</span>}
          </span>
          <span className="flex-1 text-xs text-text-primary font-mono truncate">
            {item.function_id}
          </span>
          <span className="text-[10px] text-velocity-500 font-mono">
            Gen {item.generation}
          </span>
          <span className={cn('text-xs font-mono font-semibold', fitnessColor(item.fitness_score))}>
            {Math.round(item.fitness_score)}
          </span>
        </motion.div>
      ))}
    </div>
  );
}

// ──────────────────────────────────────────────────────────────────────────────
// Mutation Funnel — visual accept/reject ratio
// ──────────────────────────────────────────────────────────────────────────────

function MutationFunnel({
  proposed,
  accepted,
  rejected,
  deployed,
}: {
  proposed: number;
  accepted: number;
  rejected: number;
  deployed: number;
}) {
  const total = proposed + accepted + rejected || 1;

  const stages = [
    { label: 'Proposed', value: proposed, color: 'bg-velocity-500' },
    { label: 'Accepted', value: accepted, color: 'bg-success' },
    { label: 'Deployed', value: deployed, color: 'bg-info' },
    { label: 'Rejected', value: rejected, color: 'bg-error' },
  ];

  return (
    <div className="space-y-2">
      {stages.map((stage) => (
        <div key={stage.label} className="flex items-center gap-3">
          <span className="text-xs text-text-secondary w-16">{stage.label}</span>
          <div className="flex-1 h-4 rounded-full bg-bg-tertiary overflow-hidden">
            <motion.div
              className={cn('h-full rounded-full', stage.color)}
              initial={{ width: 0 }}
              animate={{ width: `${(stage.value / total) * 100}%` }}
              transition={{ duration: 0.6, ease: 'easeOut' }}
            />
          </div>
          <span className="text-xs font-mono text-text-muted w-8 text-right">
            {stage.value}
          </span>
        </div>
      ))}
    </div>
  );
}

// ──────────────────────────────────────────────────────────────────────────────
// Main DNAInsightsDashboard
// ──────────────────────────────────────────────────────────────────────────────

export interface DNAInsightsDashboardProps {
  /** Per-function insights */
  functionInsights?: DNAInsights | null;
  /** Enterprise-wide insights */
  enterpriseInsights?: EnterpriseInsights | null;
  /** Loading state */
  loading?: boolean;
  /** Called when period changes */
  onPeriodChange?: (period: string) => void;
  className?: string;
}

export function DNAInsightsDashboard({
  functionInsights,
  enterpriseInsights,
  loading,
  onPeriodChange,
  className,
}: DNAInsightsDashboardProps) {
  const [period, setPeriod] = useState('30d');

  const handlePeriodChange = (p: string) => {
    setPeriod(p);
    onPeriodChange?.(p);
  };

  if (loading) {
    return (
      <div className={cn('space-y-6', className)}>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="h-24 rounded-xl bg-bg-tertiary animate-pulse" />
          ))}
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {[1, 2].map((i) => (
            <div key={i} className="h-64 rounded-xl bg-bg-tertiary animate-pulse" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className={cn('space-y-6', className)}>
      {/* Period selector */}
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-text-primary">
          {enterpriseInsights ? 'Enterprise DNA Insights' : 'Function DNA Insights'}
        </h2>
        <PeriodSelector value={period} onChange={handlePeriodChange} />
      </div>

      {/* Enterprise stats */}
      {enterpriseInsights && (
        <>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            <StatCard
              icon={<Dna className="h-4 w-4" />}
              label="Functions Analyzed"
              value={enterpriseInsights.total_functions_analyzed}
            />
            <StatCard
              icon={<Activity className="h-4 w-4" />}
              label="Mutations Proposed"
              value={enterpriseInsights.total_mutations_proposed}
              sublabel={`${enterpriseInsights.total_mutations_accepted} accepted`}
            />
            <StatCard
              icon={<Zap className="h-4 w-4" />}
              label="Avg Latency Gain"
              value={`${enterpriseInsights.avg_latency_improvement_pct.toFixed(1)}%`}
              trend="up"
            />
            <StatCard
              icon={<DollarSign className="h-4 w-4" />}
              label="Est. Savings"
              value={`$${enterpriseInsights.total_cost_savings_usd.toFixed(0)}`}
              trend="up"
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Top bottlenecks */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-base">
                  <AlertTriangle className="h-4 w-4 text-warning" />
                  Top Bottleneck Categories
                </CardTitle>
              </CardHeader>
              <CardContent>
                <BottleneckChart data={enterpriseInsights.top_bottleneck_categories} />
              </CardContent>
            </Card>

            {/* Evolution leaderboard */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-base">
                  <Trophy className="h-4 w-4 text-velocity-500" />
                  Evolution Leaderboard
                </CardTitle>
              </CardHeader>
              <CardContent>
                <EvolutionLeaderboard data={enterpriseInsights.evolution_leaderboard} />
              </CardContent>
            </Card>
          </div>
        </>
      )}

      {/* Per-function insights */}
      {functionInsights && (
        <>
          {/* Metrics summary */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            <StatCard
              icon={<Activity className="h-4 w-4" />}
              label="Executions"
              value={formatNumber(functionInsights.metrics.total_executions)}
            />
            <StatCard
              icon={<Zap className="h-4 w-4" />}
              label="P50 Latency"
              value={`${Math.round(functionInsights.metrics.p50_latency_ms)}ms`}
              sublabel={`P99: ${Math.round(functionInsights.metrics.p99_latency_ms)}ms`}
            />
            <StatCard
              icon={<BarChart3 className="h-4 w-4" />}
              label="Success Rate"
              value={`${(functionInsights.metrics.success_rate * 100).toFixed(1)}%`}
            />
            <StatCard
              icon={<Dna className="h-4 w-4" />}
              label="Cold Start Rate"
              value={`${(functionInsights.metrics.cold_start_rate * 100).toFixed(1)}%`}
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Mutation funnel */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-base">
                  <TrendingUp className="h-4 w-4 text-velocity-500" />
                  Mutation Outcomes
                </CardTitle>
              </CardHeader>
              <CardContent>
                <MutationFunnel
                  proposed={functionInsights.mutation_outcomes.outcomes.proposed}
                  accepted={functionInsights.mutation_outcomes.outcomes.accepted}
                  rejected={functionInsights.mutation_outcomes.outcomes.rejected}
                  deployed={functionInsights.mutation_outcomes.outcomes.deployed}
                />
              </CardContent>
            </Card>

            {/* Error distribution */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-base">
                  <AlertTriangle className="h-4 w-4 text-error" />
                  Error Distribution
                </CardTitle>
              </CardHeader>
              <CardContent>
                {Object.keys(functionInsights.metrics.error_distribution).length > 0 ? (
                  <BottleneckChart
                    data={Object.entries(functionInsights.metrics.error_distribution).map(
                      ([category, count]) => ({ category, count })
                    )}
                  />
                ) : (
                  <div className="text-center py-8">
                    <span className="text-2xl">✨</span>
                    <p className="text-sm text-text-muted mt-2">No errors detected</p>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </>
      )}
    </div>
  );
}
