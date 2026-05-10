import React, { useMemo } from 'react';
import { motion } from 'framer-motion';
import { cn } from '@/lib/utils';
import type { DNAProfile, BottleneckEntry } from '@/types/dna';
import {
  Activity,
  Zap,
  Shield,
  TrendingUp,
  TrendingDown,
  Dna,
  Pause,
  Play,
} from 'lucide-react';

// ──────────────────────────────────────────────────────────────────────────────
// Shared helpers (used across DNA components — DRY)
// ──────────────────────────────────────────────────────────────────────────────

export function fitnessColor(score: number): string {
  if (score >= 85) return 'text-success';
  if (score >= 65) return 'text-velocity-500';
  if (score >= 40) return 'text-warning';
  return 'text-error';
}

export function fitnessBg(score: number): string {
  if (score >= 85) return 'bg-success';
  if (score >= 65) return 'bg-velocity-500';
  if (score >= 40) return 'bg-warning';
  return 'bg-error';
}

export function fitnessGlow(score: number): string {
  if (score >= 85) return 'shadow-[0_0_20px_rgba(16,185,129,0.3)]';
  if (score >= 65) return 'shadow-[0_0_20px_rgba(249,115,22,0.3)]';
  if (score >= 40) return 'shadow-[0_0_20px_rgba(245,158,11,0.3)]';
  return 'shadow-[0_0_20px_rgba(239,68,68,0.3)]';
}

export function fitnessLabel(score: number): string {
  if (score >= 85) return 'Excellent';
  if (score >= 65) return 'Healthy';
  if (score >= 40) return 'Needs Work';
  return 'Critical';
}

export function formatNumber(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
  return n.toString();
}

// ──────────────────────────────────────────────────────────────────────────────
// Stat Card — reused in DNAHelix + Insights
// ──────────────────────────────────────────────────────────────────────────────

interface StatCardProps {
  icon: React.ReactNode;
  label: string;
  value: string | number;
  sublabel?: string;
  trend?: 'up' | 'down' | 'neutral';
  className?: string;
}

export function StatCard({ icon, label, value, sublabel, trend, className }: StatCardProps) {
  return (
    <div
      className={cn(
        'rounded-xl border border-border-subtle bg-card p-4 transition-all duration-200',
        'hover:border-border-default hover:shadow-sm',
        className
      )}
    >
      <div className="flex items-center gap-2 text-text-muted mb-2">
        {icon}
        <span className="text-xs font-medium uppercase tracking-wider">{label}</span>
      </div>
      <div className="flex items-end gap-2">
        <span className="text-2xl font-semibold text-text-primary font-mono">{value}</span>
        {trend && (
          <span className={cn('mb-1', trend === 'up' ? 'text-success' : trend === 'down' ? 'text-error' : 'text-text-muted')}>
            {trend === 'up' ? <TrendingUp className="h-4 w-4" /> : <TrendingDown className="h-4 w-4" />}
          </span>
        )}
      </div>
      {sublabel && <p className="text-xs text-text-muted mt-1">{sublabel}</p>}
    </div>
  );
}

// ──────────────────────────────────────────────────────────────────────────────
// Fitness Ring — circular progress for fitness score
// ──────────────────────────────────────────────────────────────────────────────

function FitnessRing({ score, size = 120 }: { score: number; size?: number }) {
  const radius = (size - 12) / 2;
  const circumference = 2 * Math.PI * radius;
  const offset = circumference - (score / 100) * circumference;

  const ringColor = useMemo(() => {
    if (score >= 85) return '#10b981';
    if (score >= 65) return '#f97316';
    if (score >= 40) return '#f59e0b';
    return '#ef4444';
  }, [score]);

  return (
    <div className="relative inline-flex items-center justify-center" style={{ width: size, height: size }}>
      <svg width={size} height={size} className="-rotate-90">
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke="var(--border-subtle)"
          strokeWidth="6"
        />
        <motion.circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke={ringColor}
          strokeWidth="6"
          strokeLinecap="round"
          strokeDasharray={circumference}
          initial={{ strokeDashoffset: circumference }}
          animate={{ strokeDashoffset: offset }}
          transition={{ duration: 1.2, ease: 'easeOut' }}
        />
      </svg>
      <div className="absolute flex flex-col items-center">
        <span className={cn('text-2xl font-bold font-mono', fitnessColor(score))}>
          {Math.round(score)}
        </span>
        <span className="text-[10px] text-text-muted uppercase tracking-widest">fitness</span>
      </div>
    </div>
  );
}

// ──────────────────────────────────────────────────────────────────────────────
// DNA Strand — animated helix visualization
// ──────────────────────────────────────────────────────────────────────────────

function DNAStrand({ generation, fitness }: { generation: number; fitness: number }) {
  const strandCount = Math.min(generation, 12);
  const color = useMemo(() => {
    if (fitness >= 85) return '#10b981';
    if (fitness >= 65) return '#f97316';
    if (fitness >= 40) return '#f59e0b';
    return '#ef4444';
  }, [fitness]);

  return (
    <div className="flex items-center gap-0.5 h-8">
      {Array.from({ length: strandCount }, (_, i) => (
        <motion.div
          key={i}
          className="rounded-full"
          style={{
            width: 4,
            height: 16 + Math.sin(i * 0.8) * 8,
            backgroundColor: color,
            opacity: 0.4 + (i / strandCount) * 0.6,
          }}
          initial={{ scaleY: 0 }}
          animate={{ scaleY: 1 }}
          transition={{ delay: i * 0.05, duration: 0.3 }}
        />
      ))}
      {generation > 12 && (
        <span className="text-[10px] text-text-muted ml-1 font-mono">+{generation - 12}</span>
      )}
    </div>
  );
}

// ──────────────────────────────────────────────────────────────────────────────
// Bottleneck Tags
// ──────────────────────────────────────────────────────────────────────────────

function BottleneckTag({ entry }: { entry: BottleneckEntry }) {
  const severityColors = {
    low: 'bg-info/10 text-info border-info/20',
    medium: 'bg-warning/10 text-warning border-warning/20',
    high: 'bg-error/10 text-error border-error/20',
    critical: 'bg-error/20 text-error border-error/30',
  };

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-medium border',
        severityColors[entry.severity]
      )}
    >
      {entry.type}
      <span className="opacity-60">{Math.round(entry.frequency * 100)}%</span>
    </span>
  );
}

// ──────────────────────────────────────────────────────────────────────────────
// Main DNAHelix Component
// ──────────────────────────────────────────────────────────────────────────────

export interface DNAHelixProps {
  profile: DNAProfile;
  onToggleEvolution?: (enabled: boolean) => void;
  onTriggerAnalysis?: () => void;
  isToggling?: boolean;
  isAnalyzing?: boolean;
  className?: string;
}

export function DNAHelix({
  profile,
  onToggleEvolution,
  onTriggerAnalysis,
  isToggling,
  isAnalyzing,
  className,
}: DNAHelixProps) {
  return (
    <div className={cn('space-y-6', className)}>
      {/* Header: fitness ring + generation + controls */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center gap-6">
        <FitnessRing score={profile.fitness_score} />

        <div className="flex-1 space-y-3">
          <div className="flex items-center gap-3">
            <Dna className="h-5 w-5 text-velocity-500" />
            <h3 className="text-lg font-semibold text-text-primary">
              Generation {profile.generation}
            </h3>
            <DNAStrand generation={profile.generation} fitness={profile.fitness_score} />
          </div>

          <p className={cn('text-sm font-medium', fitnessColor(profile.fitness_score))}>
            {fitnessLabel(profile.fitness_score)} — {Math.round(profile.fitness_score)}/100
          </p>

          {profile.dna_hash && (
            <p className="text-xs text-text-muted font-mono truncate max-w-md">
              {profile.dna_hash}
            </p>
          )}

          <div className="flex items-center gap-2">
            {onToggleEvolution && (
              <button
                onClick={() => onToggleEvolution(!profile.evolution_enabled)}
                disabled={isToggling}
                className={cn(
                  'inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs font-medium transition-all',
                  'border border-border-subtle hover:border-border-default',
                  profile.evolution_enabled
                    ? 'bg-success/10 text-success hover:bg-success/20'
                    : 'bg-bg-tertiary text-text-muted hover:text-text-secondary'
                )}
              >
                {profile.evolution_enabled ? (
                  <>
                    <Play className="h-3 w-3" /> Evolving
                  </>
                ) : (
                  <>
                    <Pause className="h-3 w-3" /> Paused
                  </>
                )}
              </button>
            )}
            {onTriggerAnalysis && (
              <button
                onClick={onTriggerAnalysis}
                disabled={isAnalyzing}
                className="inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs font-medium border border-border-subtle bg-bg-tertiary text-text-secondary hover:text-text-primary hover:border-border-default transition-all disabled:opacity-50"
              >
                <Activity className="h-3 w-3" />
                {isAnalyzing ? 'Analyzing...' : 'Analyze Now'}
              </button>
            )}
          </div>
        </div>
      </div>

      {/* Stats grid */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
        <StatCard
          icon={<Activity className="h-4 w-4" />}
          label="Executions"
          value={formatNumber(profile.total_executions)}
        />
        <StatCard
          icon={<Zap className="h-4 w-4" />}
          label="Avg Latency"
          value={profile.avg_latency_ms != null ? `${Math.round(profile.avg_latency_ms)}ms` : '—'}
          sublabel={profile.p99_latency_ms != null ? `P99: ${Math.round(profile.p99_latency_ms)}ms` : undefined}
        />
        <StatCard
          icon={<Shield className="h-4 w-4" />}
          label="Success Rate"
          value={`${(profile.success_rate * 100).toFixed(1)}%`}
        />
        <StatCard
          icon={<Dna className="h-4 w-4" />}
          label="Mutations"
          value={profile.total_mutations}
          sublabel={`Gen ${profile.generation}`}
        />
      </div>

      {/* Bottleneck tags */}
      {profile.bottleneck_signature.length > 0 && (
        <div className="space-y-2">
          <h4 className="text-xs font-medium text-text-muted uppercase tracking-wider">
            Detected Bottlenecks
          </h4>
          <div className="flex flex-wrap gap-2">
            {profile.bottleneck_signature.map((b, i) => (
              <BottleneckTag key={i} entry={b} />
            ))}
          </div>
        </div>
      )}

      {/* Input patterns */}
      {profile.input_patterns.length > 0 && (
        <div className="space-y-2">
          <h4 className="text-xs font-medium text-text-muted uppercase tracking-wider">
            Input Patterns
          </h4>
          <div className="space-y-1">
            {profile.input_patterns.slice(0, 5).map((p, i) => (
              <div key={i} className="flex items-center gap-2 text-xs">
                <div className="h-1.5 rounded-full bg-velocity-500/30" style={{ width: `${p.frequency * 100}%`, minWidth: 4 }}>
                  <div className="h-full rounded-full bg-velocity-500" style={{ width: '100%' }} />
                </div>
                <span className="text-text-muted font-mono truncate max-w-xs">{p.shape || p.hash}</span>
                <span className="text-text-muted">{Math.round(p.frequency * 100)}%</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
