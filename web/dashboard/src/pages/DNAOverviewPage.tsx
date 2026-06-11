import React, { useState } from 'react';
import { usePageTitle } from '@/hooks';
import {
  Dna,
  BarChart3,
  GitBranch,
  Sparkles,
  ChevronRight,
  DollarSign,
  TrendingUp,
  TrendingDown,
  Activity,
  CheckCircle2,
  Clock,
} from 'lucide-react';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { LoadingSpinner } from '@/components/ui/loading-spinner';
import { useEnterpriseDNAInsights } from '@/hooks/useFunctionDNA';
import { Link } from 'react-router-dom';
import { cn } from '@/lib/utils';

type Period = '7d' | '30d' | '90d';

const PERIOD_LABELS: Record<Period, string> = {
  '7d': '7 days',
  '30d': '30 days',
  '90d': '90 days',
};

export default function DNAOverviewPage() {
  usePageTitle('DNA Overview');
  const [period, setPeriod] = useState<Period>('30d');
  const { data: insights, isLoading, error } = useEnterpriseDNAInsights(period);

  const totalFunctions = insights?.total_functions_analyzed || 0;
  const totalMutations = insights?.total_mutations_proposed || 0;
  const acceptedMutations = insights?.total_mutations_accepted || 0;
  const pendingMutations = insights?.total_mutations_pending || 0;
  const avgFitness = insights?.avg_fitness_score || 0;
  const costSavings = insights?.total_cost_savings_usd || 0;
  const latencyImprovement = insights?.avg_latency_improvement_pct || 0;
  const acceptanceRate = totalMutations > 0 ? (acceptedMutations / totalMutations) * 100 : 0;

  // Previous period data from the insights response
  const prevCostSavings = (insights as any)?.prev_period_cost_savings_usd ?? null;
  const prevLatency = (insights as any)?.prev_period_latency_improvement_pct ?? null;

  const formatDelta = (current: number, prev: number | null, invert = false) => {
    if (prev === null || prev === 0) return null;
    const delta = ((current - prev) / prev) * 100;
    const positive = invert ? delta < 0 : delta > 0;
    return {
      value: Math.abs(delta).toFixed(1),
      positive,
      icon: positive ? <TrendingUp className="h-3 w-3" /> : <TrendingDown className="h-3 w-3" />,
    };
  };

  const costDelta = formatDelta(costSavings, prevCostSavings);
  const latencyDelta = formatDelta(latencyImprovement, prevLatency);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <LoadingSpinner />
      </div>
    );
  }

  if (error || !insights) {
    return (
      <div className="space-y-6 max-w-6xl mx-auto">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-xl bg-velocity-500/10">
            <Dna className="h-6 w-6 text-velocity-500" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-text-primary font-display">Function DNA</h1>
            <p className="text-sm text-text-secondary">
              Living code that evolves based on real production traffic
            </p>
          </div>
        </div>
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-16">
            <Dna className="h-12 w-12 text-text-muted mb-4" />
            <h3 className="text-lg font-semibold text-text-primary mb-2">DNA Not Available</h3>
            <p className="text-sm text-text-secondary mb-4 text-center max-w-md">
              Unable to load DNA insights. Please ensure you have functions with DNA enabled.
            </p>
            <Link to="/functions/my">
              <Button variant="outline" className="gap-1.5">
                Go to Functions
                <ChevronRight className="h-4 w-4 opacity-60" />
              </Button>
            </Link>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Empty state — no functions with DNA yet
  if (totalFunctions === 0) {
    return (
      <div className="space-y-6 max-w-6xl mx-auto">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-xl bg-velocity-500/10">
            <Dna className="h-6 w-6 text-velocity-500" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-text-primary font-display">Function DNA</h1>
            <p className="text-sm text-text-secondary">
              Living code that evolves based on real production traffic
            </p>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <Card className="md:col-span-2">
            <CardContent className="flex flex-col items-center justify-center py-16 px-8">
              <div className="p-3 rounded-2xl bg-velocity-500/10 mb-4">
                <Dna className="h-10 w-10 text-velocity-500" />
              </div>
              <h3 className="text-xl font-semibold text-text-primary mb-2">Enable Function DNA</h3>
              <p className="text-sm text-text-secondary mb-6 text-center max-w-lg">
                Function DNA tracks execution patterns and proposes AI-powered code optimizations.
                Enable it from Platform Settings to start analyzing your functions.
              </p>
              <div className="flex items-center gap-3">
                <Link to="/settings#platform">
                  <Button className="gap-1.5">
                    <Sparkles className="h-4 w-4" />
                    Enable in Platform Settings
                    <ChevronRight className="h-4 w-4 opacity-60" />
                  </Button>
                </Link>
                <Link to="/functions/my">
                  <Button variant="outline" className="gap-1.5">
                    Browse Functions
                    <ChevronRight className="h-4 w-4" />
                  </Button>
                </Link>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base">
                <Activity className="h-4 w-4 text-velocity-500" />
                What you'll see
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              {[
                'AI-proposed code mutations based on execution patterns',
                'Latency and cost optimization recommendations',
                'Accept/reject mutations with canary deployment',
                'Evolution history and fitness scores per function',
              ].map((item, i) => (
                <div key={i} className="flex items-start gap-2">
                  <CheckCircle2 className="h-4 w-4 text-velocity-500 mt-0.5 shrink-0" />
                  <span className="text-sm text-text-secondary">{item}</span>
                </div>
              ))}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base">
                <Sparkles className="h-4 w-4 text-velocity-500" />
                Quick Start
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                {[
                  { step: '1', text: 'Go to Platform Settings' },
                  { step: '2', text: 'Enable Auto-Evolution' },
                  { step: '3', text: 'Wait for DNA analysis' },
                ].map(({ step, text }) => (
                  <div key={step} className="flex items-center gap-2 text-sm">
                    <span className="flex items-center justify-center w-5 h-5 rounded-full bg-velocity-500/10 text-velocity-500 text-xs font-bold shrink-0">
                      {step}
                    </span>
                    <span className="text-text-primary">{text}</span>
                  </div>
                ))}
              </div>
              <p className="text-xs text-text-muted">
                DNA analyzes function execution patterns over time and proposes optimizations
                automatically.
              </p>
            </CardContent>
          </Card>
        </div>
      </div>
    );
  }

  const hasPending = pendingMutations > 0;

  return (
    <div className="space-y-6 max-w-6xl mx-auto">
      {/* Page header + period selector */}
      <div className="flex items-center gap-3 flex-wrap">
        <div className="p-2 rounded-xl bg-velocity-500/10">
          <Dna className="h-6 w-6 text-velocity-500" />
        </div>
        <div>
          <h1 className="text-2xl font-bold text-text-primary font-display">Function DNA</h1>
          <p className="text-sm text-text-secondary">
            Living code that evolves based on real production traffic
          </p>
        </div>
        <div className="ml-auto flex items-center gap-2">
          <div className="flex items-center rounded-lg border border-border-subtle bg-bg-secondary p-0.5">
            {(['7d', '30d', '90d'] as Period[]).map((p) => (
              <button
                key={p}
                onClick={() => setPeriod(p)}
                className={cn(
                  'px-3 py-1.5 text-xs font-medium rounded-md transition-all duration-200',
                  period === p
                    ? 'bg-velocity-500 text-white shadow-sm'
                    : 'text-text-muted hover:text-text-secondary hover:bg-surface-elevated'
                )}
              >
                {p.toUpperCase()}
              </button>
            ))}
          </div>
          <Badge variant="outline" className="text-xs">
            {PERIOD_LABELS[period]}
          </Badge>
        </div>
      </div>

      {/* Stats grid — 3 cols, 2 rows */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {/* 1. Functions Analyzed */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-secondary">
              Functions Analyzed
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-text-primary">{totalFunctions}</div>
            <p className="text-xs text-text-muted mt-1">with active DNA profiles</p>
          </CardContent>
        </Card>

        {/* 2. Mutations Proposed */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-secondary">
              Mutations Proposed
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-baseline gap-2">
              <div className="text-3xl font-bold text-text-primary">{totalMutations}</div>
              {acceptedMutations > 0 && (
                <Badge
                  variant="outline"
                  className="text-xs text-velocity-500 border-velocity-500/30"
                >
                  {acceptedMutations} accepted
                </Badge>
              )}
            </div>
            <p className="text-xs text-text-muted mt-1">across all functions</p>
          </CardContent>
        </Card>

        {/* 3. Acceptance Rate */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-secondary">
              Acceptance Rate
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-text-primary">{acceptanceRate.toFixed(0)}%</div>
            <p className="text-xs text-text-muted mt-1">
              {acceptedMutations}/{totalMutations} mutations approved
            </p>
          </CardContent>
        </Card>

        {/* 4. Avg Fitness Score */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-secondary">
              Avg Fitness Score
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-text-primary">{avgFitness.toFixed(1)}%</div>
            <p className="text-xs text-text-muted mt-1">across all functions</p>
          </CardContent>
        </Card>

        {/* 5. Latency Improvement */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-secondary">
              Latency Improvement
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-baseline gap-2">
              <div className="text-3xl font-bold text-velocity-500">
                +{latencyImprovement.toFixed(1)}%
              </div>
              {latencyDelta && (
                <Badge
                  variant="outline"
                  className={cn(
                    'text-xs gap-0.5',
                    latencyDelta.positive
                      ? 'text-velocity-500 border-velocity-500/30'
                      : 'text-red-400 border-red-400/30'
                  )}
                >
                  {latencyDelta.icon}
                  {latencyDelta.value}%
                </Badge>
              )}
            </div>
            <p className="text-xs text-text-muted mt-1">average improvement vs baseline</p>
          </CardContent>
        </Card>

        {/* 6. Cost Savings */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-secondary">Cost Savings</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-baseline gap-2">
              <div className="text-3xl font-bold text-velocity-500">
                <DollarSign className="inline h-5 w-5" />
                {costSavings.toFixed(0)}
              </div>
              {costDelta && (
                <Badge
                  variant="outline"
                  className={cn(
                    'text-xs gap-0.5',
                    costDelta.positive
                      ? 'text-velocity-500 border-velocity-500/30'
                      : 'text-red-400 border-red-400/30'
                  )}
                >
                  {costDelta.icon}
                  {costDelta.value}%
                </Badge>
              )}
            </div>
            <p className="text-xs text-text-muted mt-1">USD saved via optimizations</p>
          </CardContent>
        </Card>
      </div>

      {/* Second row: Top Categories + Pending Review */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Top Categories with bar chart */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-secondary">
              Top Bottleneck Categories
            </CardTitle>
          </CardHeader>
          <CardContent>
            {insights.top_bottleneck_categories && insights.top_bottleneck_categories.length > 0 ? (
              <div className="space-y-3">
                {insights.top_bottleneck_categories.slice(0, 5).map((cat, i) => {
                  const maxCount = insights.top_bottleneck_categories?.[0]?.count || 1;
                  const widthPct = (cat.count / maxCount) * 100;
                  return (
                    <div key={i} className="space-y-1">
                      <div className="flex items-center justify-between text-sm">
                        <span className="text-text-primary font-medium truncate">
                          {cat.category}
                        </span>
                        <Badge
                          variant="outline"
                          className="text-xs text-velocity-500 border-velocity-500/30 font-mono shrink-0"
                        >
                          {cat.count}
                        </Badge>
                      </div>
                      <div className="h-1.5 w-full rounded-full bg-surface-elevated overflow-hidden">
                        <div
                          className="h-full rounded-full bg-gradient-to-r from-velocity-500 to-velocity-400 transition-all duration-500"
                          style={{ width: `${widthPct}%` }}
                        />
                      </div>
                    </div>
                  );
                })}
              </div>
            ) : (
              <p className="text-sm text-text-muted">No bottleneck data yet</p>
            )}
          </CardContent>
        </Card>

        {/* Pending Review or All Caught Up */}
        {hasPending ? (
          <Card className="border-velocity-500/30 bg-velocity-500/5">
            <CardHeader className="pb-2">
              <CardTitle className="flex items-center gap-2 text-sm font-medium text-text-secondary">
                <Clock className="h-4 w-4 text-velocity-500" />
                Pending Review
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="text-3xl font-bold text-velocity-500">{pendingMutations}</div>
              <p className="text-xs text-text-muted">
                mutation{pendingMutations !== 1 ? 's' : ''} awaiting your approval
              </p>
              <Link to="/functions/my?dna_pending=true">
                <Button size="sm" className="w-full gap-1.5 mt-1">
                  Review Mutations
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </Link>
            </CardContent>
          </Card>
        ) : (
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="flex items-center gap-2 text-sm font-medium text-text-secondary">
                <CheckCircle2 className="h-4 w-4 text-success" />
                All Caught Up
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-text-secondary">
                No pending mutations — all proposed changes have been reviewed.
              </p>
            </CardContent>
          </Card>
        )}
      </div>

      {/* Quick Actions */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Sparkles className="h-4 w-4 text-velocity-500" />
            Quick Actions
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            {[
              {
                label: 'Browse Functions',
                description: 'View functions with DNA profiles',
                icon: Dna,
                href: '/functions/my',
                cta: 'Go to Functions',
              },
              {
                label: 'Evolution History',
                description: 'Track all mutation events across functions',
                icon: GitBranch,
                href: '/functions/my?tab=evolution',
                cta: 'View History',
              },
              {
                label: 'Platform Settings',
                description: 'Configure DNA auto-evolution and notifications',
                icon: Activity,
                href: '/settings#platform',
                cta: 'Configure',
              },
              {
                label: 'Performance Analytics',
                description: 'Deep-dive into latency, cost, and fitness metrics',
                icon: BarChart3,
                href: '/analytics',
                cta: 'View Analytics',
              },
            ].map(({ label, description, icon: Icon, href, cta }) => (
              <div
                key={label}
                className="flex items-center justify-between p-4 rounded-lg border border-border-subtle bg-card hover:border-velocity-500/30 hover:bg-velocity-500/5 transition-colors group"
              >
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-lg bg-velocity-500/10">
                    <Icon className="h-4 w-4 text-velocity-500" />
                  </div>
                  <div>
                    <p className="font-medium text-text-primary text-sm">{label}</p>
                    <p className="text-xs text-text-muted">{description}</p>
                  </div>
                </div>
                <Link to={href}>
                  <Button
                    variant="outline"
                    size="sm"
                    className="gap-1.5 shrink-0 group-hover:border-velocity-500/40"
                  >
                    {cta}
                    <ChevronRight className="h-4 w-4" />
                  </Button>
                </Link>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Cost Indicator */}
      <div className="rounded-xl border border-border-subtle bg-bg-secondary/50 p-4">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-lg bg-velocity-500/10">
            <Sparkles className="h-4 w-4 text-velocity-500" />
          </div>
          <div>
            <p className="text-sm font-medium text-text-primary">Mutation Cost</p>
            <p className="text-xs text-text-muted">
              Each accepted mutation costs{' '}
              <span className="font-mono text-velocity-500">50 credits</span> from your wallet.
              Rejected mutations are free.
            </p>
          </div>
          <Badge
            variant="outline"
            className="ml-auto font-mono text-velocity-500 border-velocity-500/30"
          >
            50 cr
          </Badge>
        </div>
      </div>
    </div>
  );
}
