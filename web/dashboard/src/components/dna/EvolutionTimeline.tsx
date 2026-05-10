import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';
import type { DNAMutation, MutationStatus, MutationType } from '@/types/dna';
import {
  MUTATION_TYPE_META,
  MUTATION_STATUS_META,
} from '@/types/dna';
import {
  Zap,
  Cpu,
  ShieldAlert,
  Shield,
  GitBranch,
  ChevronDown,
  ChevronRight,
  Clock,
  CheckCircle2,
  XCircle,
  Loader2,
  RotateCcw,
  ArrowRight,
  Dna,
} from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';

// ──────────────────────────────────────────────────────────────────────────────
// Icon resolver (DRY — shared with DNAHelix via type metadata)
// ──────────────────────────────────────────────────────────────────────────────

const TYPE_ICONS: Record<MutationType, React.ReactNode> = {
  optimize_latency: <Zap className="h-4 w-4" />,
  reduce_memory: <Cpu className="h-4 w-4" />,
  fix_error_pattern: <ShieldAlert className="h-4 w-4" />,
  improve_reliability: <Shield className="h-4 w-4" />,
  refactor_hotpath: <GitBranch className="h-4 w-4" />,
};

const STATUS_ICONS: Record<MutationStatus, React.ReactNode> = {
  proposed: <Clock className="h-3.5 w-3.5" />,
  accepted: <CheckCircle2 className="h-3.5 w-3.5" />,
  rejected: <XCircle className="h-3.5 w-3.5" />,
  deploying: <Loader2 className="h-3.5 w-3.5 animate-spin" />,
  deployed: <CheckCircle2 className="h-3.5 w-3.5" />,
  rolled_back: <RotateCcw className="h-3.5 w-3.5" />,
};

// ──────────────────────────────────────────────────────────────────────────────
// Impact Bar
// ──────────────────────────────────────────────────────────────────────────────

function ImpactBar({ label, value, suffix = '%' }: { label: string; value: number; suffix?: string }) {
  const color = value > 0 ? 'bg-success' : value < 0 ? 'bg-error' : 'bg-text-muted';
  const absValue = Math.abs(value);
  return (
    <div className="space-y-1">
      <div className="flex justify-between text-[10px]">
        <span className="text-text-muted">{label}</span>
        <span className={cn('font-mono', value > 0 ? 'text-success' : value < 0 ? 'text-error' : 'text-text-muted')}>
          {value > 0 ? '+' : ''}{value.toFixed(1)}{suffix}
        </span>
      </div>
      <div className="h-1 rounded-full bg-border-subtle overflow-hidden">
        <motion.div
          className={cn('h-full rounded-full', color)}
          initial={{ width: 0 }}
          animate={{ width: `${Math.min(absValue, 100)}%` }}
          transition={{ duration: 0.6, ease: 'easeOut' }}
        />
      </div>
    </div>
  );
}

// ──────────────────────────────────────────────────────────────────────────────
// Timeline Entry
// ──────────────────────────────────────────────────────────────────────────────

interface TimelineEntryProps {
  mutation: DNAMutation;
  isExpanded: boolean;
  onToggle: () => void;
  onViewDiff?: (mutationId: string) => void;
}

function TimelineEntry({ mutation, isExpanded, onToggle, onViewDiff }: TimelineEntryProps) {
  const meta = MUTATION_TYPE_META[mutation.mutation_type];
  const statusMeta = MUTATION_STATUS_META[mutation.status];

  return (
    <motion.div
      layout
      className={cn(
        'relative rounded-xl border transition-all duration-200',
        isExpanded
          ? 'border-velocity-500/30 bg-card shadow-sm'
          : 'border-border-subtle bg-card hover:border-border-default'
      )}
    >
      {/* Timeline connector dot */}
      <div
        className={cn(
          'absolute -left-[21px] top-5 w-3 h-3 rounded-full border-2',
          mutation.status === 'deployed' ? 'bg-success border-success' :
          mutation.status === 'proposed' ? 'bg-velocity-500 border-velocity-500' :
          mutation.status === 'rejected' ? 'bg-error border-error' :
          'bg-bg-tertiary border-border-default'
        )}
      />

      {/* Header */}
      <button
        onClick={onToggle}
        className="w-full flex items-center gap-3 p-4 text-left"
      >
        <div className={cn('flex-shrink-0', meta.color)}>
          {TYPE_ICONS[mutation.mutation_type]}
        </div>

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-sm font-medium text-text-primary">{meta.label}</span>
            <Badge variant={statusMeta.variant} className="text-[10px]">
              {STATUS_ICONS[mutation.status]}
              <span className="ml-1">{statusMeta.label}</span>
            </Badge>
            <span className="text-[10px] text-text-muted font-mono">
              Gen {mutation.generation}
            </span>
          </div>
          {mutation.trigger_reason && !isExpanded && (
            <p className="text-xs text-text-muted truncate mt-0.5">
              {mutation.trigger_reason}
            </p>
          )}
        </div>

        <div className="flex items-center gap-3 flex-shrink-0">
          {mutation.confidence != null && (
            <span
              className={cn(
                'text-xs font-mono px-2 py-0.5 rounded-full',
                mutation.confidence >= 0.8
                  ? 'bg-success/10 text-success'
                  : mutation.confidence >= 0.5
                  ? 'bg-warning/10 text-warning'
                  : 'bg-error/10 text-error'
              )}
            >
              {Math.round(mutation.confidence * 100)}%
            </span>
          )}
          <span className="text-[10px] text-text-muted hidden sm:block">
            {formatDistanceToNow(new Date(mutation.created_at), { addSuffix: true })}
          </span>
          {isExpanded ? (
            <ChevronDown className="h-4 w-4 text-text-muted" />
          ) : (
            <ChevronRight className="h-4 w-4 text-text-muted" />
          )}
        </div>
      </button>

      {/* Expanded details */}
      <AnimatePresence>
        {isExpanded && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.2 }}
            className="overflow-hidden"
          >
            <div className="px-4 pb-4 space-y-4 border-t border-border-subtle pt-3">
              {/* Trigger reason */}
              {mutation.trigger_reason && (
                <p className="text-sm text-text-secondary">{mutation.trigger_reason}</p>
              )}

              {/* Estimated impact */}
              {mutation.estimated_impact && (
                <div className="space-y-2">
                  <h5 className="text-[10px] text-text-muted uppercase tracking-wider font-medium">
                    {mutation.actual_impact ? 'Actual Impact' : 'Estimated Impact'}
                  </h5>
                  <div className="grid grid-cols-3 gap-3">
                    <ImpactBar
                      label="Latency"
                      value={(mutation.actual_impact ?? mutation.estimated_impact).latency_improvement_pct}
                    />
                    <ImpactBar
                      label="Memory"
                      value={(mutation.actual_impact ?? mutation.estimated_impact).memory_reduction_pct}
                    />
                    <ImpactBar
                      label="Reliability"
                      value={(mutation.actual_impact ?? mutation.estimated_impact).reliability_improvement_pct}
                    />
                  </div>
                </div>
              )}

              {/* Metadata */}
              <div className="flex flex-wrap gap-x-4 gap-y-1 text-[10px] text-text-muted">
                {mutation.model_used && (
                  <span>Model: <span className="text-text-secondary font-mono">{mutation.model_used}</span></span>
                )}
                {mutation.executions_analyzed && (
                  <span>Analyzed: <span className="text-text-secondary font-mono">{mutation.executions_analyzed.toLocaleString()}</span> executions</span>
                )}
                {mutation.analysis_window_hours && (
                  <span>Window: <span className="text-text-secondary font-mono">{mutation.analysis_window_hours}h</span></span>
                )}
                {mutation.accepted_by && (
                  <span>Accepted by: <span className="text-text-secondary">{mutation.accepted_by}</span></span>
                )}
                {mutation.rejected_reason && (
                  <span>Reason: <span className="text-text-secondary">{mutation.rejected_reason}</span></span>
                )}
              </div>

              {/* View diff button */}
              {mutation.diff && onViewDiff && (
                <button
                  onClick={() => onViewDiff(mutation.id)}
                  className="inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs font-medium bg-velocity-500/10 text-velocity-500 hover:bg-velocity-500/20 transition-colors"
                >
                  View Code Diff <ArrowRight className="h-3 w-3" />
                </button>
              )}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  );
}

// ──────────────────────────────────────────────────────────────────────────────
// EvolutionTimeline
// ──────────────────────────────────────────────────────────────────────────────

export interface EvolutionTimelineProps {
  mutations: DNAMutation[];
  loading?: boolean;
  onViewDiff?: (mutationId: string) => void;
  className?: string;
}

export function EvolutionTimeline({ mutations, loading, onViewDiff, className }: EvolutionTimelineProps) {
  const [expandedId, setExpandedId] = useState<string | null>(null);

  if (loading) {
    return (
      <div className={cn('space-y-3', className)}>
        {[1, 2, 3].map((i) => (
          <div key={i} className="h-16 rounded-xl bg-bg-tertiary animate-pulse" />
        ))}
      </div>
    );
  }

  if (mutations.length === 0) {
    return (
      <div className={cn('text-center py-12', className)}>
        <Dna className="h-8 w-8 text-text-muted mx-auto mb-3" />
        <p className="text-sm text-text-muted">No evolution history yet</p>
        <p className="text-xs text-text-muted mt-1">
          Mutations will appear here as your function runs and evolves
        </p>
      </div>
    );
  }

  return (
    <div className={cn('relative', className)}>
      {/* Timeline line */}
      <div className="absolute left-[7px] top-4 bottom-4 w-px bg-border-subtle" />

      <div className="space-y-3">
        {mutations.map((m) => (
          <TimelineEntry
            key={m.id}
            mutation={m}
            isExpanded={expandedId === m.id}
            onToggle={() => setExpandedId(expandedId === m.id ? null : m.id)}
            onViewDiff={onViewDiff}
          />
        ))}
      </div>
    </div>
  );
}


