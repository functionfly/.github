/**
 * Trust Metrics Section Component
 *
 * Displays user's trust score with detailed metrics breakdown.
 * Shows per-component trust scores (reliability, latency, error rate, user rating, verification)
 * when available from the user trust breakdown API, plus reputation profile scores.
 */

import { getTrustColorConfig, getTrustScoreBand } from '@/components/functions/TrustScoreBadge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { cn } from '@/lib/utils';
import type { UserTrustBreakdown } from '@/types';
import { Activity, BookOpen, CheckCircle, Clock, Shield, Star, Target, Users, Zap, Hammer, Lightbulb, GraduationCap, Bot, type LucideIcon } from 'lucide-react';

export interface TrustMetricsSectionProps {
  trustScore: number;
  trustBreakdown?: UserTrustBreakdown | null;
  builderScore?: number;
  optimizerScore?: number;
  mentorScore?: number;
  agentWhispererScore?: number;
  reputationTier?: string;
  overallReputationScore?: number;
}

const RING_SIZE = 96;
const RING_STROKE = 5;

function TrustScoreRing({ score }: { score: number }) {
  const normalized = Math.max(0, Math.min(100, Number.isFinite(score) ? score : 0));
  const band = getTrustScoreBand(normalized);
  const colorConfig = getTrustColorConfig(band);
  const radius = (RING_SIZE - RING_STROKE) / 2;
  const circumference = 2 * Math.PI * radius;
  const strokeDashoffset = circumference - (normalized / 100) * circumference;

  return (
    <div
      className="relative flex h-24 w-24 shrink-0 items-center justify-center"
      aria-label={`Trust score ${Math.round(normalized)} percent`}
    >
      <svg width={RING_SIZE} height={RING_SIZE} className="-rotate-90" aria-hidden="true">
        <circle
          cx={RING_SIZE / 2}
          cy={RING_SIZE / 2}
          r={radius}
          fill="none"
          stroke="currentColor"
          strokeWidth={RING_STROKE}
          className="text-muted-foreground/25"
        />
        <circle
          cx={RING_SIZE / 2}
          cy={RING_SIZE / 2}
          r={radius}
          fill="none"
          stroke={colorConfig.primary}
          strokeWidth={RING_STROKE}
          strokeLinecap="round"
          strokeDasharray={circumference}
          strokeDashoffset={strokeDashoffset}
          className="transition-[stroke-dashoffset] duration-700 ease-out"
        />
      </svg>
      <div className="pointer-events-none absolute inset-0 flex flex-col items-center justify-center gap-0 px-1">
        <span className={cn('text-xl font-bold font-mono leading-none tabular-nums', colorConfig.text)}>
          {Math.round(normalized)}
        </span>
        <span className="text-[10px] font-medium leading-tight text-text-muted">% trust</span>
      </div>
    </div>
  );
}

function getComponentColor(score: number): string {
  if (score >= 80) return 'bg-emerald-500';
  if (score >= 60) return 'bg-blue-500';
  if (score >= 40) return 'bg-amber-500';
  return 'bg-orange-500';
}

function getComponentTextColor(score: number): string {
  if (score >= 80) return 'text-emerald-500';
  if (score >= 60) return 'text-blue-500';
  if (score >= 40) return 'text-amber-500';
  return 'text-orange-500';
}

interface TrustComponentBarProps {
  label: string;
  score: number;
  icon: LucideIcon;
  suffix?: string;
}

function TrustComponentBar({ label, score, icon: Icon, suffix = '%' }: TrustComponentBarProps) {
  const clampedScore = Math.max(0, Math.min(100, score));
  return (
    <div className="flex items-center gap-3">
      <Icon className="w-4 h-4 text-text-muted" />
      <span className="text-sm text-text-secondary w-28">{label}</span>
      <div className="flex-1">
        <div className="h-2 bg-border-subtle rounded-full overflow-hidden">
          <div
            className={cn('h-full rounded-full transition-all duration-1000', getComponentColor(clampedScore))}
            style={{ width: `${clampedScore}%` }}
          />
        </div>
      </div>
      <span className={cn('text-sm font-medium font-mono tabular-nums w-10 text-right', getComponentTextColor(clampedScore))}>
        {suffix === '%' ? Math.round(clampedScore) : Math.round(clampedScore)}{suffix}
      </span>
    </div>
  );
}

interface ReputationScoreBarProps {
  label: string;
  score: number;
  maxScore: number;
  icon: LucideIcon;
  color: string;
}

function ReputationScoreBar({ label, score, maxScore, icon: Icon, color }: ReputationScoreBarProps) {
  const percentage = maxScore > 0 ? Math.min(100, (score / maxScore) * 100) : 0;

  return (
    <div className="flex items-center gap-3">
      <Icon className={cn('w-4 h-4', color)} />
      <span className="text-sm text-text-secondary w-28">{label}</span>
      <div className="flex-1">
        <div className="h-2 bg-border-subtle rounded-full overflow-hidden">
          <div
            className={cn('h-full rounded-full transition-all duration-1000', color.replace('text-', 'bg-'))}
            style={{ width: `${percentage}%` }}
          />
        </div>
      </div>
      <span className="text-sm font-medium font-mono tabular-nums text-text-primary w-10 text-right">
        {score.toLocaleString()}
      </span>
    </div>
  );
}

export function TrustMetricsSection({
  trustScore,
  trustBreakdown,
  builderScore,
  optimizerScore,
  mentorScore,
  agentWhispererScore,
  reputationTier,
  overallReputationScore,
}: TrustMetricsSectionProps) {
  const normalizedTrust = Math.max(0, Math.min(100, Number.isFinite(trustScore) ? trustScore : 0));
  const band = getTrustScoreBand(normalizedTrust);
  const headlineConfig = getTrustColorConfig(band);

  const hasReputationProfile =
    builderScore !== undefined ||
    optimizerScore !== undefined ||
    mentorScore !== undefined ||
    agentWhispererScore !== undefined;

  const components = trustBreakdown?.components;

  return (
    <Card className="border-border-subtle">
      <CardHeader className="pb-3">
        <CardTitle className="text-lg font-display flex items-center gap-2">
          <Target className="w-5 h-5 text-brand-500" />
          Trust Metrics
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex items-center gap-4 mb-6">
          <TrustScoreRing score={normalizedTrust} />
          <div>
            <h4 className={cn('font-medium', headlineConfig.text)}>
              {headlineConfig.label} Reputation
            </h4>
            <p className="text-sm text-text-muted">
              {trustBreakdown
                ? `Based on ${trustBreakdown.functions_with_trust} function${trustBreakdown.functions_with_trust !== 1 ? 's' : ''}`
                : 'Based on function quality, community engagement, and execution reliability'}
            </p>
          </div>
        </div>

        {/* Component trust scores — real data from API when available */}
        {components ? (
          <div className="space-y-3">
            <TrustComponentBar
              label="Reliability"
              score={components.reliability}
              icon={Activity}
            />
            <TrustComponentBar
              label="Latency"
              score={components.latency}
              icon={Zap}
            />
            <TrustComponentBar
              label="Error Rate"
              score={components.error_rate}
              icon={Shield}
            />
            <TrustComponentBar
              label="User Rating"
              score={components.user_rating}
              icon={Star}
            />
            <TrustComponentBar
              label="Verification"
              score={components.verification}
              icon={CheckCircle}
            />
          </div>
        ) : (
          <div className="space-y-3">
            <div className="flex items-center gap-3">
              <Shield className="w-4 h-4 text-text-muted" />
              <span className="text-sm text-text-secondary w-28">Overall Trust</span>
              <div className="flex-1">
                <div className="h-2 bg-border-subtle rounded-full overflow-hidden">
                  <div
                    className={cn('h-full rounded-full transition-all duration-1000', getComponentColor(normalizedTrust))}
                    style={{ width: `${Math.max(0, Math.min(100, normalizedTrust))}%` }}
                  />
                </div>
              </div>
              <span className="text-sm font-medium font-mono tabular-nums text-text-primary w-10 text-right">
                {Math.round(normalizedTrust)}
              </span>
            </div>
          </div>
        )}

        {/* Execution metrics summary when breakdown is available */}
        {trustBreakdown?.metrics && trustBreakdown.metrics.total_calls > 0 && (
          <div className="mt-4 pt-4 border-t border-border-subtle">
            <div className="grid grid-cols-2 gap-3 text-xs">
              <div className="flex items-center gap-1.5 text-text-muted">
                <Activity className="w-3.5 h-3.5" />
                <span>{trustBreakdown.metrics.total_calls.toLocaleString()} calls</span>
              </div>
              <div className="flex items-center gap-1.5 text-text-muted">
                <CheckCircle className="w-3.5 h-3.5" />
                <span>{Math.round(trustBreakdown.metrics.success_rate)}% success</span>
              </div>
              {trustBreakdown.metrics.avg_p50_latency_ms > 0 && (
                <div className="flex items-center gap-1.5 text-text-muted">
                  <Clock className="w-3.5 h-3.5" />
                  <span>p50 {Math.round(trustBreakdown.metrics.avg_p50_latency_ms)}ms</span>
                </div>
              )}
              {trustBreakdown.metrics.avg_p95_latency_ms > 0 && (
                <div className="flex items-center gap-1.5 text-text-muted">
                  <Zap className="w-3.5 h-3.5" />
                  <span>p95 {Math.round(trustBreakdown.metrics.avg_p95_latency_ms)}ms</span>
                </div>
              )}
            </div>
          </div>
        )}

        {/* User Reputation Profile Scores */}
        {hasReputationProfile && (
          <>
            <div className="mt-6 pt-6 border-t border-border-subtle">
              <div className="flex items-center justify-between mb-4">
                <h4 className="text-sm font-medium text-text-primary">Reputation Profile</h4>
                {reputationTier && (
                  <span className="text-xs px-2 py-1 rounded-full bg-brand-500/10 text-brand-400 capitalize">
                    {reputationTier.replace('_', ' ')}
                  </span>
                )}
              </div>
              <div className="space-y-3">
                {builderScore !== undefined && (
                  <ReputationScoreBar
                    label="Builder"
                    score={builderScore}
                    maxScore={10000}
                    icon={Hammer}
                    color="text-blue-500"
                  />
                )}
                {optimizerScore !== undefined && (
                  <ReputationScoreBar
                    label="Optimizer"
                    score={optimizerScore}
                    maxScore={10000}
                    icon={Lightbulb}
                    color="text-yellow-500"
                  />
                )}
                {mentorScore !== undefined && (
                  <ReputationScoreBar
                    label="Mentor"
                    score={mentorScore}
                    maxScore={10000}
                    icon={GraduationCap}
                    color="text-green-500"
                  />
                )}
                {agentWhispererScore !== undefined && (
                  <ReputationScoreBar
                    label="Agent Whisperer"
                    score={agentWhispererScore}
                    maxScore={10000}
                    icon={Bot}
                    color="text-purple-500"
                  />
                )}
              </div>
              {overallReputationScore !== undefined && (
                <div className="mt-4 pt-4 border-t border-border-subtle">
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium text-text-primary">Overall Score</span>
                    <span className="text-lg font-bold text-brand-500">
                      {overallReputationScore.toLocaleString()}
                    </span>
                  </div>
                </div>
              )}
            </div>
          </>
        )}
      </CardContent>
    </Card>
  );
}
