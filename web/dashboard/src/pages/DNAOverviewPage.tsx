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
  Loader2,
} from 'lucide-react';
import {
  PageGrid,
  Chamber,
  CornerBrace,
  TrustSeal,
  SealedButton,
  FrameButton,
  StatusPill,
  GaugeStrip,
  Gauge,
  AnnotationTag,
  Card,
} from '@/components/containment';
import { useEnterpriseDNAInsights } from '@/hooks/useFunctionDNA';
import { Link } from 'react-router-dom';
import { cn } from '@/lib/utils';

import './dna-overview.css';

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
  const pendingMutations = insights?.total_mutations_proposed || 0;
  const avgFitness = insights?.avg_fitness_score || 0;
  const costSavings = insights?.total_cost_savings_usd || 0;
  const latencyImprovement = insights?.avg_latency_improvement_pct || 0;
  const acceptanceRate = totalMutations > 0 ? (acceptedMutations / totalMutations) * 100 : 0;

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
      <div className="dna-page">
        <PageGrid />
        <div className="dna-loading">
          <Loader2 className="dna-loading__spinner" />
        </div>
      </div>
    );
  }

  if (error || !insights) {
    return (
      <div className="dna-page">
        <PageGrid />

        <Chamber className="dna-hero">
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <div className="dna-hero__header">
            <div className="dna-hero__title-row">
              <TrustSeal size="lg" />
              <h1 className="dna-hero__title">Function DNA</h1>
            </div>
            <p className="dna-hero__subtitle">Living code that evolves based on real production traffic</p>
          </div>
        </Chamber>

        <Chamber className="dna-empty">
          <Dna className="dna-empty__icon" />
          <h3 className="dna-empty__title">DNA Not Available</h3>
          <p className="dna-empty__desc">Unable to load DNA insights. Please ensure you have functions with DNA enabled.</p>
          <Link to="/functions/my">
            <FrameButton iconRight={<ChevronRight className="h-4 w-4" />}>
              Go to Functions
            </FrameButton>
          </Link>
        </Chamber>
      </div>
    );
  }

  // Empty state — no functions with DNA yet
  if (totalFunctions === 0) {
    return (
      <div className="dna-page">
        <PageGrid />

        <Chamber className="dna-hero" ribs>
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <AnnotationTag primary="MODULE DNA-01" secondary="Function DNA" position="top-right" />
          <div className="dna-hero__header">
            <div className="dna-hero__title-row">
              <TrustSeal size="lg" />
              <h1 className="dna-hero__title">Function DNA</h1>
            </div>
            <p className="dna-hero__subtitle">Living code that evolves based on real production traffic</p>
          </div>
        </Chamber>

        <Chamber className="dna-empty">
          <CornerBrace position="tr" />
          <CornerBrace position="bl" />
          <div className="dna-empty__center">
            <div className="dna-empty__icon-wrap">
              <Dna className="dna-empty__icon-lg" />
            </div>
            <h2 className="dna-empty__title">Enable Function DNA</h2>
            <p className="dna-empty__desc">
              Function DNA tracks execution patterns and proposes AI-powered code optimizations.
              Enable it from Platform Settings to start analyzing your functions.
            </p>
            <div className="dna-empty__actions">
              <Link to="/settings#platform">
                <SealedButton iconLeft={<Sparkles className="h-4 w-4" />} iconRight={<ChevronRight className="h-4 w-4" />}>
                  Enable in Platform Settings
                </SealedButton>
              </Link>
              <Link to="/functions/my">
                <FrameButton iconRight={<ChevronRight className="h-4 w-4" />}>
                  Browse Functions
                </FrameButton>
              </Link>
            </div>
          </div>
        </Chamber>

        <div className="dna-info-grid">
          <Card className="dna-info-card">
            <div className="dna-info-card__header">
              <Activity className="dna-info-card__icon" />
              <h3 className="dna-info-card__title">What you'll see</h3>
            </div>
            <div className="dna-info-card__list">
              {[
                'AI-proposed code mutations based on execution patterns',
                'Latency and cost optimization recommendations',
                'Accept/reject mutations with canary deployment',
                'Evolution history and fitness scores per function',
              ].map((item, i) => (
                <div key={i} className="dna-info-card__item">
                  <CheckCircle2 className="dna-info-card__check" />
                  <span>{item}</span>
                </div>
              ))}
            </div>
          </Card>

          <Card className="dna-info-card">
            <div className="dna-info-card__header">
              <Sparkles className="dna-info-card__icon" />
              <h3 className="dna-info-card__title">Quick Start</h3>
            </div>
            <div className="dna-info-card__steps">
              {[
                { step: '01', text: 'Go to Platform Settings' },
                { step: '02', text: 'Enable Auto-Evolution' },
                { step: '03', text: 'Wait for DNA analysis' },
              ].map(({ step, text }) => (
                <div key={step} className="dna-info-card__step">
                  <span className="dna-info-card__step-num">{step}</span>
                  <span>{text}</span>
                </div>
              ))}
            </div>
            <p className="dna-info-card__hint">
              DNA analyzes function execution patterns over time and proposes optimizations automatically.
            </p>
          </Card>
        </div>
      </div>
    );
  }

  const hasPending = pendingMutations > 0;

  return (
    <div className="dna-page">
      <PageGrid />

      {/* Hero */}
      <Chamber className="dna-hero" ribs>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary="MODULE DNA-01" secondary="Function DNA" position="top-right" />

        <div className="dna-hero__header">
          <div className="dna-hero__title-row">
            <TrustSeal size="lg" />
            <h1 className="dna-hero__title">Function DNA</h1>
          </div>
          <p className="dna-hero__subtitle">Living code that evolves based on real production traffic</p>
          <div className="dna-hero__controls">
            <div className="dna-period-selector">
              {(['7d', '30d', '90d'] as Period[]).map((p) => (
                <button
                  key={p}
                  onClick={() => setPeriod(p)}
                  className={cn('dna-period-btn', period === p && 'dna-period-btn--active')}
                >
                  {p.toUpperCase()}
                </button>
              ))}
            </div>
            <StatusPill status="live" label={PERIOD_LABELS[period]} />
          </div>
        </div>

        <GaugeStrip>
          <Gauge isFirst data={{ value: totalFunctions, label: 'Functions Analyzed' }} />
          <Gauge data={{ value: totalMutations, label: 'Mutations Proposed' }} />
          <Gauge data={{ value: `${acceptanceRate.toFixed(0)}%`, label: 'Acceptance Rate' }} />
        </GaugeStrip>
      </Chamber>

      {/* Stats Grid */}
      <div className="dna-stats-grid">
        <Card className="dna-stat-card">
          <span className="dna-stat-card__label">Functions Analyzed</span>
          <span className="dna-stat-card__value">{totalFunctions}</span>
          <span className="dna-stat-card__hint">with active DNA profiles</span>
        </Card>

        <Card className="dna-stat-card">
          <span className="dna-stat-card__label">Mutations Proposed</span>
          <div className="dna-stat-card__row">
            <span className="dna-stat-card__value">{totalMutations}</span>
            {acceptedMutations > 0 && (
              <span className="dna-stat-card__badge">{acceptedMutations} accepted</span>
            )}
          </div>
          <span className="dna-stat-card__hint">across all functions</span>
        </Card>

        <Card className="dna-stat-card">
          <span className="dna-stat-card__label">Acceptance Rate</span>
          <span className="dna-stat-card__value">{acceptanceRate.toFixed(0)}%</span>
          <span className="dna-stat-card__hint">{acceptedMutations}/{totalMutations} mutations approved</span>
        </Card>

        <Card className="dna-stat-card">
          <span className="dna-stat-card__label">Avg Fitness Score</span>
          <span className="dna-stat-card__value">{avgFitness.toFixed(1)}%</span>
          <span className="dna-stat-card__hint">across all functions</span>
        </Card>

        <Card className="dna-stat-card">
          <span className="dna-stat-card__label">Latency Improvement</span>
          <div className="dna-stat-card__row">
            <span className="dna-stat-card__value dna-stat-card__value--accent">+{latencyImprovement.toFixed(1)}%</span>
            {latencyDelta && (
              <span className={cn('dna-stat-card__delta', latencyDelta.positive ? 'dna-stat-card__delta--ok' : 'dna-stat-card__delta--bad')}>
                {latencyDelta.icon} {latencyDelta.value}%
              </span>
            )}
          </div>
          <span className="dna-stat-card__hint">average improvement vs baseline</span>
        </Card>

        <Card className="dna-stat-card">
          <span className="dna-stat-card__label">Cost Savings</span>
          <div className="dna-stat-card__row">
            <span className="dna-stat-card__value dna-stat-card__value--accent">
              <DollarSign className="dna-stat-card__dollar" />{costSavings.toFixed(0)}
            </span>
            {costDelta && (
              <span className={cn('dna-stat-card__delta', costDelta.positive ? 'dna-stat-card__delta--ok' : 'dna-stat-card__delta--bad')}>
                {costDelta.icon} {costDelta.value}%
              </span>
            )}
          </div>
          <span className="dna-stat-card__hint">USD saved via optimizations</span>
        </Card>
      </div>

      {/* Bottom row */}
      <div className="dna-bottom-grid">
        {/* Top Categories */}
        <Chamber className="dna-categories">
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <h2 className="dna-section-title">Top Bottleneck Categories</h2>
          {insights.top_bottleneck_categories && insights.top_bottleneck_categories.length > 0 ? (
            <div className="dna-categories__list">
              {insights.top_bottleneck_categories.slice(0, 5).map((cat, i) => {
                const maxCount = insights.top_bottleneck_categories?.[0]?.count || 1;
                const widthPct = (cat.count / maxCount) * 100;
                return (
                  <div key={i} className="dna-categories__item">
                    <div className="dna-categories__item-header">
                      <span className="dna-categories__name">{cat.category}</span>
                      <span className="dna-categories__count">{cat.count}</span>
                    </div>
                    <div className="dna-categories__bar">
                      <div className="dna-categories__fill" style={{ width: `${widthPct}%` }} />
                    </div>
                  </div>
                );
              })}
            </div>
          ) : (
            <p className="dna-categories__empty">No bottleneck data yet</p>
          )}
        </Chamber>

        {/* Pending Review */}
        <Chamber className={cn('dna-pending', hasPending && 'dna-pending--active')}>
          <CornerBrace position="tr" />
          <CornerBrace position="bl" />
          {hasPending ? (
            <>
              <div className="dna-pending__header">
                <Clock className="dna-pending__icon" />
                <h2 className="dna-section-title">Pending Review</h2>
              </div>
              <div className="dna-pending__value">{pendingMutations}</div>
              <p className="dna-pending__hint">
                mutation{pendingMutations !== 1 ? 's' : ''} awaiting your approval
              </p>
              <Link to="/functions/my?dna_pending=true">
                <SealedButton className="dna-pending__btn" iconRight={<ChevronRight className="h-4 w-4" />}>
                  Review Mutations
                </SealedButton>
              </Link>
            </>
          ) : (
            <>
              <div className="dna-pending__header">
                <CheckCircle2 className="dna-pending__icon dna-pending__icon--ok" />
                <h2 className="dna-section-title">All Caught Up</h2>
              </div>
              <p className="dna-pending__desc">No pending mutations — all proposed changes have been reviewed.</p>
            </>
          )}
        </Chamber>
      </div>

      {/* Quick Actions */}
      <Chamber className="dna-actions">
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <div className="dna-actions__header">
          <Sparkles className="dna-actions__icon" />
          <h2 className="dna-section-title">Quick Actions</h2>
        </div>
        <div className="dna-actions-grid">
          {[
            { label: 'Browse Functions', description: 'View functions with DNA profiles', icon: Dna, href: '/functions/my', cta: 'Go to Functions' },
            { label: 'Evolution History', description: 'Track all mutation events across functions', icon: GitBranch, href: '/functions/my?tab=evolution', cta: 'View History' },
            { label: 'Platform Settings', description: 'Configure DNA auto-evolution and notifications', icon: Activity, href: '/settings#platform', cta: 'Configure' },
            { label: 'Performance Analytics', description: 'Deep-dive into latency, cost, and fitness metrics', icon: BarChart3, href: '/analytics', cta: 'View Analytics' },
          ].map(({ label, description, icon: Icon, href, cta }) => (
            <Link key={label} to={href} className="dna-action-card">
              <div className="dna-action-card__info">
                <div className="dna-action-card__icon-wrap">
                  <Icon className="dna-action-card__icon" />
                </div>
                <div>
                  <p className="dna-action-card__label">{label}</p>
                  <p className="dna-action-card__desc">{description}</p>
                </div>
              </div>
              <FrameButton size="sm" iconRight={<ChevronRight className="h-4 w-4" />}>
                {cta}
              </FrameButton>
            </Link>
          ))}
        </div>
      </Chamber>

      {/* Cost Indicator */}
      <div className="dna-cost-bar">
        <div className="dna-cost-bar__info">
          <div className="dna-cost-bar__icon-wrap">
            <Sparkles className="dna-cost-bar__icon" />
          </div>
          <div>
            <p className="dna-cost-bar__title">Mutation Cost</p>
            <p className="dna-cost-bar__desc">
              Each accepted mutation costs{' '}
              <span className="dna-cost-bar__credits">50 credits</span> from your wallet.
              Rejected mutations are free.
            </p>
          </div>
        </div>
        <span className="dna-cost-bar__badge">50 cr</span>
      </div>
    </div>
  );
}
