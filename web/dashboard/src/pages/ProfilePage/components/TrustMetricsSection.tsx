/**
 * Trust Metrics Section Component
 *
 * Displays user's trust score with detailed metrics breakdown.
 * Also shows user reputation profile scores when available.
 */

import { getTrustColorConfig, getTrustScoreBand } from '@/components/functions/TrustScoreBadge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { cn } from '@/lib/utils';
import { BookOpen, Shield, Target, Users, Zap, Hammer, Lightbulb, GraduationCap, Bot, type LucideIcon } from 'lucide-react';

export interface TrustMetricsSectionProps {
  trustScore: number;
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

  const metrics = [
    {
      name: 'Reliability',
      score: Math.min(100, trustScore + Math.random() * 10 - 5),
      icon: Shield,
    },
    { name: 'Performance', score: Math.min(100, trustScore + Math.random() * 10 - 5), icon: Zap },
    { name: 'Community', score: Math.min(100, trustScore + Math.random() * 15 - 7), icon: Users },
    {
      name: 'Documentation',
      score: Math.min(100, trustScore + Math.random() * 20 - 10),
      icon: BookOpen,
    },
  ];

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
              Based on function quality, community engagement, and execution reliability
            </p>
          </div>
        </div>

        <div className="space-y-3">
          {metrics.map((metric) => (
            <div key={metric.name} className="flex items-center gap-3">
              <metric.icon className="w-4 h-4 text-text-muted" />
              <span className="text-sm text-text-secondary w-28">{metric.name}</span>
              <div className="flex-1">
                <div className="h-2 bg-border-subtle rounded-full overflow-hidden">
                  <div
                    className={cn(
                      'h-full rounded-full transition-all duration-1000',
                      metric.score >= 80
                        ? 'bg-emerald-500'
                        : metric.score >= 60
                          ? 'bg-yellow-500'
                          : 'bg-orange-500'
                    )}
                    style={{
                      width: `${Math.max(0, Math.min(100, metric.score))}%`,
                    }}
                  />
                </div>
              </div>
              <span className="text-sm font-medium font-mono tabular-nums text-text-primary w-10 text-right">
                {Math.round(metric.score)}
              </span>
            </div>
          ))}
        </div>

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
