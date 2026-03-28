/**
 * Trust Metrics Section Component
 *
 * Displays user's trust score with detailed metrics breakdown.
 */

import { getTrustColorConfig, getTrustScoreBand } from '@/components/functions/TrustScoreBadge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { cn } from '@/lib/utils';
import { BookOpen, Shield, Target, Users, Zap } from 'lucide-react';

export interface TrustMetricsSectionProps {
  trustScore: number;
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
        <span className={cn('text-xl font-bold leading-none tabular-nums', colorConfig.text)}>
          {Math.round(normalized)}
        </span>
        <span className="text-[10px] font-medium leading-tight text-text-muted">% trust</span>
      </div>
    </div>
  );
}

export function TrustMetricsSection({ trustScore }: TrustMetricsSectionProps) {
  const normalizedTrust = Math.max(0, Math.min(100, Number.isFinite(trustScore) ? trustScore : 0));
  const band = getTrustScoreBand(normalizedTrust);
  const headlineConfig = getTrustColorConfig(band);

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
        <CardTitle className="text-lg flex items-center gap-2">
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
              <span className="text-sm font-medium text-text-primary w-10 text-right">
                {Math.round(metric.score)}
              </span>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
