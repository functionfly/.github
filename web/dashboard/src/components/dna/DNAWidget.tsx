import React, { useMemo } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  fitnessColor,
  fitnessBg,
  fitnessGlow,
  fitnessLabel,
  formatNumber,
} from './DNAHelix';
import type { DNAProfile, DNAMutation } from '@/types/dna';
import { MUTATION_TYPE_META, MUTATION_STATUS_META } from '@/types/dna';
import {
  Dna,
  Activity,
  Zap,
  Shield,
  GitBranch,
  ChevronRight,
  Loader2,
  Sparkles,
  TrendingUp,
  TrendingDown,
  AlertTriangle,
  CheckCircle2,
  Clock,
  Lock,
  ExternalLink,
} from 'lucide-react';
import { Link } from 'react-router-dom';
import { useAuthStore } from '@/stores/authStore';

// ──────────────────────────────────────────────────────────────────────────────
// Animated fitness orb — glowing SVG ring with particle effects
// ──────────────────────────────────────────────────────────────────────────────

function FitnessOrb({ score, size = 88 }: { score: number; size?: number }) {
  const radius = (size - 10) / 2;
  const circumference = 2 * Math.PI * radius;
  const offset = circumference - (score / 100) * circumference;

  const ringColor = useMemo(() => {
    if (score >= 85) return '#10b981';
    if (score >= 65) return '#f97316';
    if (score >= 40) return '#f59e0b';
    return '#ef4444';
  }, [score]);

  const glowColor = useMemo(() => {
    if (score >= 85) return 'rgba(16,185,129,0.15)';
    if (score >= 65) return 'rgba(249,115,22,0.15)';
    if (score >= 40) return 'rgba(245,158,11,0.15)';
    return 'rgba(239,68,68,0.15)';
  }, [score]);

  return (
    <div className="relative inline-flex items-center justify-center" style={{ width: size, height: size }}>
      {/* Glow backdrop */}
      <div
        className="absolute inset-0 rounded-full blur-xl"
        style={{ backgroundColor: glowColor }}
      />

      {/* SVG ring */}
      <svg width={size} height={size} className="relative -rotate-90">
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke="var(--border-subtle)"
          strokeWidth="4"
        />
        <motion.circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke={ringColor}
          strokeWidth="4"
          strokeLinecap="round"
          strokeDasharray={circumference}
          initial={{ strokeDashoffset: circumference }}
          animate={{ strokeDashoffset: offset }}
          transition={{ duration: 1.4, ease: [0.16, 1, 0.3, 1] }}
        />
      </svg>

      {/* Center content */}
      <div className="absolute flex flex-col items-center">
        <motion.span
          className={cn('text-xl font-bold font-mono leading-none', fitnessColor(score))}
          initial={{ opacity: 0, scale: 0.5 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ delay: 0.3, duration: 0.4 }}
        >
          {Math.round(score)}
        </motion.span>
        <span className="text-[8px] text-text-muted uppercase tracking-[0.15em] mt-0.5">
          fitness
        </span>
      </div>
    </div>
  );
}

// ──────────────────────────────────────────────────────────────────────────────
// DNA strand bar — animated bar showing generation progression
// ──────────────────────────────────────────────────────────────────────────────

function GenerationBar({ generation, fitness }: { generation: number; fitness: number }) {
  const maxDisplay = 10;
  const barCount = Math.min(generation, maxDisplay);

  return (
    <div className="flex items-center gap-1">
      {Array.from({ length: barCount }, (_, i) => (
        <motion.div
          key={i}
          className="rounded-sm"
          style={{
            width: 4,
            height: 12 + Math.sin(i * 0.9) * 6,
            backgroundColor:
              fitness >= 85 ? '#10b981' :
              fitness >= 65 ? '#f97316' :
              fitness >= 40 ? '#f59e0b' : '#ef4444',
            opacity: 0.3 + (i / barCount) * 0.7,
          }}
          initial={{ scaleY: 0 }}
          animate={{ scaleY: 1 }}
          transition={{ delay: 0.6 + i * 0.04, duration: 0.3 }}
        />
      ))}
      {generation > maxDisplay && (
        <span className="text-[10px] text-text-muted ml-0.5 font-mono">+{generation - maxDisplay}</span>
      )}
    </div>
  );
}

// ──────────────────────────────────────────────────────────────────────────────
// Micro metric — inline metric display
// ──────────────────────────────────────────────────────────────────────────────

function MicroMetric({
  icon,
  label,
  value,
  trend,
  className,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
  trend?: 'up' | 'down';
  className?: string;
}) {
  return (
    <div className={cn('flex items-center gap-2', className)}>
      <span className="text-text-muted">{icon}</span>
      <div className="min-w-0 flex-1">
        <p className="text-[10px] text-text-muted uppercase tracking-wider truncate">{label}</p>
        <div className="flex items-center gap-1">
          <span className="text-sm font-semibold font-mono text-text-primary">{value}</span>
          {trend === 'up' && <TrendingUp className="h-3 w-3 text-success" />}
          {trend === 'down' && <TrendingDown className="h-3 w-3 text-error" />}
        </div>
      </div>
    </div>
  );
}

// ──────────────────────────────────────────────────────────────────────────────
// Recent mutation pill — compact display of latest mutation
// ──────────────────────────────────────────────────────────────────────────────

function RecentMutationPill({ mutation }: { mutation: DNAMutation }) {
  const typeMeta = MUTATION_TYPE_META[mutation.mutation_type];
  const statusMeta = MUTATION_STATUS_META[mutation.status];

  const statusIcon = useMemo(() => {
    switch (mutation.status) {
      case 'proposed': return <Clock className="h-3 w-3" />;
      case 'accepted': return <CheckCircle2 className="h-3 w-3 text-success" />;
      case 'deployed': return <CheckCircle2 className="h-3 w-3 text-success" />;
      case 'rejected': return <AlertTriangle className="h-3 w-3 text-error" />;
      case 'rolled_back': return <AlertTriangle className="h-3 w-3 text-error" />;
      default: return <Clock className="h-3 w-3 text-warning" />;
    }
  }, [mutation.status]);

  return (
    <div className="flex items-center gap-2 rounded-lg bg-bg-tertiary/60 px-2.5 py-2 border border-border-subtle">
      {statusIcon}
      <div className="min-w-0 flex-1">
        <p className="text-[10px] font-medium text-text-primary truncate">
          {typeMeta.label}
        </p>
        <p className="text-[9px] text-text-muted truncate">
          {mutation.trigger_reason || 'AI-generated optimization'}
        </p>
      </div>
      <Badge
        variant={statusMeta.variant}
        className="text-[9px] px-1.5 py-0 shrink-0"
      >
        {statusMeta.label}
      </Badge>
    </div>
  );
}

// ──────────────────────────────────────────────────────────────────────────────
// Main DNAWidget — compact sidebar widget for function detail pages
// ──────────────────────────────────────────────────────────────────────────────

export interface DNAWidgetProps {
  /** Function ID for linking to full DNA page */
  functionId: string;
  /** Function author/name for routing */
  functionSlug?: string;
  /** DNA profile data — null means DNA not enabled */
  profile: DNAProfile | null;
  /** Latest mutations (up to 3) */
  recentMutations?: DNAMutation[];
  /** Loading state */
  isLoading?: boolean;
  /** Optional className override */
  className?: string;
}

export function DNAWidget({
  functionId,
  functionSlug,
  profile,
  recentMutations = [],
  isLoading = false,
  className,
}: DNAWidgetProps) {
  const username = useAuthStore((s) => s.user?.username) ?? '';

  // Loading skeleton
  if (isLoading) {
    return (
      <motion.div
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        className={cn(
          'rounded-xl border border-border-subtle bg-card overflow-hidden',
          className
        )}
      >
        <div className="p-4 space-y-3">
          <div className="flex items-center gap-2">
            <div className="h-8 w-8 rounded-lg bg-bg-tertiary animate-pulse" />
            <div className="space-y-1.5 flex-1">
              <div className="h-3 w-24 bg-bg-tertiary rounded animate-pulse" />
              <div className="h-2.5 w-16 bg-bg-tertiary rounded animate-pulse" />
            </div>
          </div>
          <div className="flex items-center justify-center py-4">
            <Loader2 className="h-5 w-5 animate-spin text-text-muted" />
          </div>
        </div>
      </motion.div>
    );
  }

  // DNA not enabled — show anatomy explainer showing what DNA is made of
  if (!profile) {
    return (
      <motion.div
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className={cn(
          'rounded-xl border border-border-subtle bg-card overflow-hidden',
          className
        )}
      >
        {/* Gradient header strip */}
        <div className="relative h-1 bg-gradient-to-r from-velocity-500 via-brand-500 to-velocity-400" />

        <div className="p-4 space-y-4">
          {/* Title */}
          <div className="flex items-center gap-2.5">
            <div className="relative">
              <div className="p-1.5 rounded-lg bg-velocity-500/10">
                <Dna className="h-4 w-4 text-velocity-500" />
              </div>
              <motion.div
                className="absolute -top-0.5 -right-0.5 h-2 w-2 rounded-full bg-velocity-500"
                animate={{ scale: [1, 1.4, 1], opacity: [1, 0.5, 1] }}
                transition={{ repeat: Infinity, duration: 2 }}
              />
            </div>
            <div>
              <p className="text-xs font-semibold text-text-primary">Function DNA</p>
              <p className="text-[10px] text-text-muted">Living code evolution system</p>
            </div>
          </div>

          {/* DNA Helix SVG — animated double helix */}
          <div className="flex justify-center py-1">
            <svg width="160" height="64" viewBox="0 0 160 64" className="overflow-visible">
              {/* Strand 1 */}
              <motion.path
                d="M 8 32 Q 28 8, 48 32 Q 68 56, 88 32 Q 108 8, 128 32 Q 148 56, 160 40"
                fill="none"
                stroke="var(--velocity-500, #8b5cf6)"
                strokeWidth="2"
                strokeLinecap="round"
                initial={{ pathLength: 0, opacity: 0 }}
                animate={{ pathLength: 1, opacity: 0.8 }}
                transition={{ duration: 1.6, ease: [0.16, 1, 0.3, 1] }}
              />
              {/* Strand 2 */}
              <motion.path
                d="M 8 32 Q 28 56, 48 32 Q 68 8, 88 32 Q 108 56, 128 32 Q 148 8, 160 24"
                fill="none"
                stroke="var(--brand-500, #3b82f6)"
                strokeWidth="2"
                strokeLinecap="round"
                initial={{ pathLength: 0, opacity: 0 }}
                animate={{ pathLength: 1, opacity: 0.6 }}
                transition={{ duration: 1.6, delay: 0.3, ease: [0.16, 1, 0.3, 1] }}
              />
              {/* Rungs — connecting bars between strands */}
              {[28, 48, 68, 88, 108, 128].map((x, i) => {
                const y1 = 32 + (i % 2 === 0 ? -16 : 16);
                const y2 = 32 + (i % 2 === 0 ? 16 : -16);
                return (
                  <motion.line
                    key={i}
                    x1={x}
                    y1={y1}
                    x2={x}
                    y2={y2}
                    stroke="var(--border-subtle)"
                    strokeWidth="1"
                    strokeDasharray="2 2"
                    initial={{ opacity: 0, scaleY: 0 }}
                    animate={{ opacity: 0.4, scaleY: 1 }}
                    transition={{ delay: 0.8 + i * 0.08, duration: 0.3 }}
                  />
                );
              })}
              {/* Node dots at strand intersections */}
              {[48, 88, 128].map((x, i) => (
                <motion.circle
                  key={i}
                  cx={x}
                  cy={32}
                  r="3"
                  fill="var(--velocity-500, #8b5cf6)"
                  initial={{ scale: 0 }}
                  animate={{ scale: 1 }}
                  transition={{ delay: 1.0 + i * 0.12, type: 'spring', stiffness: 400 }}
                />
              ))}
            </svg>
          </div>

          {/* DNA Anatomy — what it's made of */}
          <div className="space-y-2">
            <p className="text-[10px] text-text-muted uppercase tracking-wider font-medium">
              What DNA Tracks
            </p>
            <div className="space-y-1.5">
              {[
                {
                  icon: <Activity className="h-3.5 w-3.5 text-velocity-500" />,
                  label: 'Execution Patterns',
                  desc: 'Input shapes, latency curves, error clusters',
                  color: 'bg-velocity-500/10',
                },
                {
                  icon: <Zap className="h-3.5 w-3.5 text-warning" />,
                  label: 'Performance Bottlenecks',
                  desc: 'Hot paths, cold starts, memory pressure',
                  color: 'bg-warning/10',
                },
                {
                  icon: <Shield className="h-3.5 w-3.5 text-success" />,
                  label: 'Reliability Signals',
                  desc: 'Success rate, error patterns, edge cases',
                  color: 'bg-success/10',
                },
                {
                  icon: <GitBranch className="h-3.5 w-3.5 text-info" />,
                  label: 'Code Mutations',
                  desc: 'AI-proposed optimizations with canary deploy',
                  color: 'bg-info/10',
                },
              ].map((item, i) => (
                <motion.div
                  key={item.label}
                  initial={{ opacity: 0, x: -8 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: 0.6 + i * 0.1, duration: 0.3 }}
                  className="flex items-start gap-2.5 rounded-lg bg-bg-tertiary/40 px-3 py-2 border border-border-subtle/50"
                >
                  <div className={cn('p-1 rounded shrink-0 mt-0.5', item.color)}>
                    {item.icon}
                  </div>
                  <div className="min-w-0">
                    <p className="text-[11px] font-medium text-text-primary">{item.label}</p>
                    <p className="text-[10px] text-text-muted leading-snug">{item.desc}</p>
                  </div>
                </motion.div>
              ))}
            </div>
          </div>

          {/* Evolution cycle */}
          <div className="space-y-2">
            <p className="text-[10px] text-text-muted uppercase tracking-wider font-medium">
              Evolution Cycle
            </p>
            <div className="flex items-center gap-1 justify-center">
              {[
                { label: 'Observe', icon: <Activity className="h-3 w-3" /> },
                { label: 'Analyze', icon: <Zap className="h-3 w-3" /> },
                { label: 'Propose', icon: <Sparkles className="h-3 w-3" /> },
                { label: 'Deploy', icon: <CheckCircle2 className="h-3 w-3" /> },
              ].map((step, i, arr) => (
                <React.Fragment key={step.label}>
                  <motion.div
                    initial={{ opacity: 0, scale: 0.8 }}
                    animate={{ opacity: 1, scale: 1 }}
                    transition={{ delay: 1.0 + i * 0.1 }}
                    className="flex flex-col items-center gap-0.5"
                  >
                    <div className="h-6 w-6 rounded-full bg-velocity-500/10 flex items-center justify-center text-velocity-500">
                      {step.icon}
                    </div>
                    <span className="text-[8px] text-text-muted font-medium">{step.label}</span>
                  </motion.div>
                  {i < arr.length - 1 && (
                    <motion.div
                      initial={{ opacity: 0, scaleX: 0 }}
                      animate={{ opacity: 0.3, scaleX: 1 }}
                      transition={{ delay: 1.1 + i * 0.1 }}
                      className="flex-1 h-px bg-gradient-to-r from-velocity-500/40 to-velocity-500/10 max-w-[24px] mt-[-10px]"
                    />
                  )}
                </React.Fragment>
              ))}
            </div>
          </div>

          {/* Mutation types preview */}
          <div className="flex flex-wrap gap-1 justify-center">
            {[
              { type: 'Latency', color: 'text-velocity-500 bg-velocity-500/10 border-velocity-500/20' },
              { type: 'Memory', color: 'text-info bg-info/10 border-info/20' },
              { type: 'Errors', color: 'text-error bg-error/10 border-error/20' },
              { type: 'Reliability', color: 'text-success bg-success/10 border-success/20' },
              { type: 'Hot Path', color: 'text-warning bg-warning/10 border-warning/20' },
            ].map((tag, i) => (
              <motion.span
                key={tag.type}
                initial={{ opacity: 0, y: 4 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 1.4 + i * 0.06 }}
                className={cn(
                  'inline-flex items-center rounded-full px-2 py-0.5 text-[9px] font-medium border',
                  tag.color
                )}
              >
                {tag.type}
              </motion.span>
            ))}
          </div>

          {/* Settings link — DNA is enabled from Platform Settings */}
          <motion.div
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 1.8, duration: 0.4 }}
          >
            <Link to={`/u/${username}/settings`}>
              <Button
                variant="outline"
                size="sm"
                className="w-full text-xs gap-1.5 group"
              >
                <Sparkles className="h-3 w-3" />
                Enable in Platform Settings
                <ChevronRight className="h-3 w-3 ml-auto transition-transform group-hover:translate-x-0.5" />
              </Button>
            </Link>
            <p className="text-center text-[9px] text-text-muted mt-1.5">
              DNA is configured per-account in settings
            </p>
          </motion.div>
        </div>
      </motion.div>
    );
  }

  const hasActiveProposal = recentMutations.some((m) => m.status === 'proposed');

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4 }}
      className={cn(
        'rounded-xl border border-border-subtle bg-card overflow-hidden',
        className
      )}
    >
      {/* Header strip with gradient */}
      <div className="relative h-1 bg-gradient-to-r from-velocity-500 via-brand-500 to-velocity-400" />

      <div className="p-4 space-y-4">
        {/* Title row */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="p-1.5 rounded-lg bg-velocity-500/10">
              <Dna className="h-4 w-4 text-velocity-500" />
            </div>
            <div>
              <p className="text-xs font-semibold text-text-primary">Function DNA</p>
              <p className="text-[10px] text-text-muted">Evolving since gen {profile.generation}</p>
            </div>
          </div>
          <Badge
            variant="outline"
            className={cn(
              'text-[10px] font-mono px-1.5 py-0',
              profile.evolution_enabled
                ? 'border-success/30 text-success'
                : 'border-text-muted/30 text-text-muted'
            )}
          >
            {profile.evolution_enabled ? '● Active' : '○ Paused'}
          </Badge>
        </div>

        {/* Fitness orb + generation bar */}
        <div className="flex items-center gap-4">
          <FitnessOrb score={profile.fitness_score} size={88} />
          <div className="flex-1 space-y-2">
            <div>
              <p className={cn('text-sm font-semibold', fitnessColor(profile.fitness_score))}>
                {fitnessLabel(profile.fitness_score)}
              </p>
              <p className="text-[10px] text-text-muted">
                Generation {profile.generation}
              </p>
            </div>
            <GenerationBar generation={profile.generation} fitness={profile.fitness_score} />
            {profile.dna_hash && (
              <p className="text-[9px] text-text-muted font-mono truncate" title={profile.dna_hash}>
                <Lock className="h-2.5 w-2.5 inline mr-0.5 -mt-px" />
                {profile.dna_hash.length > 24 ? profile.dna_hash.slice(0, 24) + '…' : profile.dna_hash}
              </p>
            )}
          </div>
        </div>

        {/* Metrics grid */}
        <div className="grid grid-cols-2 gap-2.5">
          <MicroMetric
            icon={<Activity className="h-3.5 w-3.5" />}
            label="Executions"
            value={formatNumber(profile.total_executions)}
          />
          <MicroMetric
            icon={<Zap className="h-3.5 w-3.5" />}
            label="Avg Latency"
            value={profile.avg_latency_ms != null ? `${Math.round(profile.avg_latency_ms)}ms` : '—'}
          />
          <MicroMetric
            icon={<Shield className="h-3.5 w-3.5" />}
            label="Success"
            value={`${(profile.success_rate * 100).toFixed(1)}%`}
            trend={profile.success_rate >= 0.99 ? 'up' : profile.success_rate < 0.95 ? 'down' : undefined}
          />
          <MicroMetric
            icon={<GitBranch className="h-3.5 w-3.5" />}
            label="Mutations"
            value={String(profile.total_mutations)}
          />
        </div>

        {/* Bottleneck tags */}
        {profile.bottleneck_signature.length > 0 && (
          <div className="space-y-1.5">
            <p className="text-[10px] text-text-muted uppercase tracking-wider">Bottlenecks</p>
            <div className="flex flex-wrap gap-1">
              {profile.bottleneck_signature.slice(0, 3).map((b, i) => (
                <span
                  key={i}
                  className={cn(
                    'inline-flex items-center rounded-full px-2 py-0.5 text-[9px] font-medium border',
                    b.severity === 'critical' || b.severity === 'high'
                      ? 'bg-error/10 text-error border-error/20'
                      : b.severity === 'medium'
                        ? 'bg-warning/10 text-warning border-warning/20'
                        : 'bg-info/10 text-info border-info/20'
                  )}
                >
                  {b.type}
                  <span className="ml-1 opacity-60">{Math.round(b.frequency * 100)}%</span>
                </span>
              ))}
            </div>
          </div>
        )}

        {/* Recent mutations */}
        {recentMutations.length > 0 && (
          <div className="space-y-1.5">
            <p className="text-[10px] text-text-muted uppercase tracking-wider">Recent Activity</p>
            <div className="space-y-1">
              {recentMutations.slice(0, 2).map((m) => (
                <RecentMutationPill key={m.id} mutation={m} />
              ))}
            </div>
          </div>
        )}

        {/* Active proposal callout */}
        <AnimatePresence>
          {hasActiveProposal && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              exit={{ opacity: 0, height: 0 }}
              className="overflow-hidden"
            >
              <div className="rounded-lg bg-velocity-500/8 border border-velocity-500/20 px-3 py-2 flex items-center gap-2">
                <Sparkles className="h-3.5 w-3.5 text-velocity-500 shrink-0" />
                <p className="text-[11px] text-velocity-500 font-medium">
                  New evolution proposed — review now
                </p>
              </div>
            </motion.div>
          )}
        </AnimatePresence>

        {/* CTA link to full DNA page */}
        <Link to={`/functions/${functionId}/dna`} className="block">
          <Button
            variant="outline"
            size="sm"
            className="w-full text-xs gap-1.5 group"
          >
            <Dna className="h-3 w-3" />
            View Full DNA Dashboard
            <ChevronRight className="h-3 w-3 ml-auto transition-transform group-hover:translate-x-0.5" />
          </Button>
        </Link>
      </div>
    </motion.div>
  );
}
