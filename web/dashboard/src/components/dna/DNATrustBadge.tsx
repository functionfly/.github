import React from 'react';
import { motion } from 'framer-motion';
import { cn } from '@/lib/utils';
import { Dna, TrendingUp, Sparkles } from 'lucide-react';
import { fitnessColor, fitnessLabel, formatNumber } from './DNAHelix';

// ──────────────────────────────────────────────────────────────────────────────
// DNATrustBadge — compact badge for marketplace function cards
// ──────────────────────────────────────────────────────────────────────────────

export interface DNATrustBadgeProps {
  generation: number;
  fitnessScore: number;
  totalMutations: number;
  totalExecutions?: number;
  variant?: 'compact' | 'full' | 'micro';
  className?: string;
}

export function DNATrustBadge({
  generation,
  fitnessScore,
  totalMutations,
  totalExecutions,
  variant = 'compact',
  className,
}: DNATrustBadgeProps) {
  // Micro: just an icon + generation number
  if (variant === 'micro') {
    return (
      <span
        className={cn(
          'inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-semibold',
          'bg-velocity-500/10 text-velocity-500 border border-velocity-500/20',
          className
        )}
        title={`Generation ${generation} — ${fitnessLabel(fitnessScore)}`}
      >
        <Dna className="h-3 w-3" />
        Gen {generation}
      </span>
    );
  }

  // Compact: generation + fitness + mutation count
  if (variant === 'compact') {
    return (
      <div
        className={cn(
          'inline-flex items-center gap-2 rounded-lg px-2.5 py-1.5 text-xs',
          'border border-border-subtle bg-card',
          className
        )}
      >
        <div className="flex items-center gap-1 text-velocity-500">
          <Dna className="h-3.5 w-3.5" />
          <span className="font-semibold font-mono">Gen {generation}</span>
        </div>

        <div className="w-px h-3 bg-border-subtle" />

        <div className="flex items-center gap-1">
          <span
            className={cn(
              'inline-block w-1.5 h-1.5 rounded-full',
              fitnessScore >= 85 ? 'bg-success' :
              fitnessScore >= 65 ? 'bg-velocity-500' :
              fitnessScore >= 40 ? 'bg-warning' : 'bg-error'
            )}
          />
          <span className={cn('font-mono', fitnessColor(fitnessScore))}>
            {Math.round(fitnessScore)}
          </span>
        </div>

        <div className="w-px h-3 bg-border-subtle" />

        <span className="text-text-muted">
          {totalMutations} evolution{totalMutations !== 1 ? 's' : ''}
        </span>
      </div>
    );
  }

  // Full: detailed trust card for function detail pages
  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className={cn(
        'rounded-xl border border-border-subtle bg-card p-4 space-y-3',
        className
      )}
    >
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="p-1.5 rounded-lg bg-velocity-500/10">
            <Dna className="h-4 w-4 text-velocity-500" />
          </div>
          <div>
            <p className="text-xs font-medium text-text-primary">Function DNA</p>
            <p className="text-[10px] text-text-muted">Living code that evolves</p>
          </div>
        </div>
        <span className={cn('text-xs font-semibold px-2 py-0.5 rounded-full', fitnessColor(fitnessScore), 'bg-current/10')}>
          {fitnessLabel(fitnessScore)}
        </span>
      </div>

      <div className="grid grid-cols-3 gap-2">
        <div className="text-center">
          <p className="text-lg font-bold font-mono text-velocity-500">{generation}</p>
          <p className="text-[10px] text-text-muted">Generation</p>
        </div>
        <div className="text-center">
          <p className={cn('text-lg font-bold font-mono', fitnessColor(fitnessScore))}>
            {Math.round(fitnessScore)}
          </p>
          <p className="text-[10px] text-text-muted">Fitness</p>
        </div>
        <div className="text-center">
          <p className="text-lg font-bold font-mono text-text-primary">{totalMutations}</p>
          <p className="text-[10px] text-text-muted">Evolutions</p>
        </div>
      </div>

      {totalExecutions != null && (
        <div className="flex items-center justify-center gap-1 text-[10px] text-text-muted pt-1 border-t border-border-subtle">
          <Sparkles className="h-3 w-3" />
          Evolved from {formatNumber(totalExecutions)} executions
        </div>
      )}

      {generation > 1 && (
        <div className="flex items-center justify-center gap-1 text-[10px] text-success">
          <TrendingUp className="h-3 w-3" />
          {totalMutations}x improved since v1
        </div>
      )}
    </motion.div>
  );
}
