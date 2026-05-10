import React, { useMemo, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';
import type { DNAMutation, MutationImpact } from '@/types/dna';
import { MUTATION_TYPE_META, MUTATION_STATUS_META } from '@/types/dna';
import {
  Check,
  X,
  ChevronLeft,
  ChevronRight,
  Copy,
  Check as CheckIcon,
  ArrowRight,
  TrendingUp,
  TrendingDown,
} from 'lucide-react';

// ──────────────────────────────────────────────────────────────────────────────
// Diff Parser — splits unified diff into structured lines
// ──────────────────────────────────────────────────────────────────────────────

interface DiffLine {
  type: 'add' | 'remove' | 'context' | 'header';
  content: string;
  oldLineNum?: number;
  newLineNum?: number;
}

function parseDiff(diffText: string): DiffLine[] {
  if (!diffText) return [];

  const lines = diffText.split('\n');
  const result: DiffLine[] = [];
  let oldLine = 0;
  let newLine = 0;

  for (const line of lines) {
    if (line.startsWith('@@')) {
      const match = line.match(/@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@/);
      if (match) {
        oldLine = parseInt(match[1], 10);
        newLine = parseInt(match[2], 10);
      }
      result.push({ type: 'header', content: line });
    } else if (line.startsWith('+')) {
      result.push({ type: 'add', content: line.slice(1), newLineNum: newLine++ });
    } else if (line.startsWith('-')) {
      result.push({ type: 'remove', content: line.slice(1), oldLineNum: oldLine++ });
    } else {
      result.push({ type: 'context', content: line.startsWith(' ') ? line.slice(1) : line, oldLineNum: oldLine++, newLineNum: newLine++ });
    }
  }

  return result;
}

// ──────────────────────────────────────────────────────────────────────────────
// Impact Comparison
// ──────────────────────────────────────────────────────────────────────────────

function ImpactComparison({
  estimated,
  actual,
}: {
  estimated: MutationImpact;
  actual: MutationImpact | null;
}) {
  const metrics = [
    { label: 'Latency', key: 'latency_improvement_pct' as const, icon: '⚡' },
    { label: 'Memory', key: 'memory_reduction_pct' as const, icon: '🧠' },
    { label: 'Reliability', key: 'reliability_improvement_pct' as const, icon: '🛡️' },
  ];

  return (
    <div className="grid grid-cols-3 gap-4">
      {metrics.map(({ label, key, icon }) => {
        const est = estimated[key];
        const act = actual?.[key];
        return (
          <div key={key} className="text-center space-y-1">
            <span className="text-lg">{icon}</span>
            <p className="text-[10px] text-text-muted uppercase tracking-wider">{label}</p>
            <div className="flex items-center justify-center gap-1">
              <span
                className={cn(
                  'text-sm font-mono font-semibold',
                  est > 0 ? 'text-success' : est < 0 ? 'text-error' : 'text-text-muted'
                )}
              >
                {est > 0 ? '+' : ''}{est.toFixed(1)}%
              </span>
              {act != null && (
                <>
                  <ArrowRight className="h-3 w-3 text-text-muted" />
                  <span
                    className={cn(
                      'text-sm font-mono font-bold',
                      act > 0 ? 'text-success' : act < 0 ? 'text-error' : 'text-text-muted'
                    )}
                  >
                    {act > 0 ? '+' : ''}{act.toFixed(1)}%
                  </span>
                  {act > est ? (
                    <TrendingUp className="h-3 w-3 text-success" />
                  ) : act < est ? (
                    <TrendingDown className="h-3 w-3 text-error" />
                  ) : null}
                </>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
}

// ──────────────────────────────────────────────────────────────────────────────
// Code Panel — renders code with line numbers
// ──────────────────────────────────────────────────────────────────────────────

function CodePanel({
  code,
  title,
  className,
}: {
  code: string;
  title: string;
  className?: string;
}) {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(code);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className={cn('flex flex-col rounded-lg border border-border-subtle overflow-hidden', className)}>
      <div className="flex items-center justify-between px-3 py-1.5 bg-bg-tertiary border-b border-border-subtle">
        <span className="text-[10px] font-medium text-text-muted uppercase tracking-wider">{title}</span>
        <button onClick={handleCopy} className="text-text-muted hover:text-text-primary transition-colors">
          {copied ? <CheckIcon className="h-3 w-3 text-success" /> : <Copy className="h-3 w-3" />}
        </button>
      </div>
      <pre className="flex-1 overflow-auto p-3 text-xs font-mono leading-relaxed bg-bg-secondary">
        {code.split('\n').map((line, i) => (
          <div key={i} className="flex">
            <span className="w-8 flex-shrink-0 text-right pr-3 text-text-muted select-none opacity-50">
              {i + 1}
            </span>
            <code className="flex-1 text-text-primary">{line}</code>
          </div>
        ))}
      </pre>
    </div>
  );
}

// ──────────────────────────────────────────────────────────────────────────────
// Diff Lines Panel — renders unified diff
// ──────────────────────────────────────────────────────────────────────────────

function DiffLinesPanel({ diff, className }: { diff: string; className?: string }) {
  const lines = useMemo(() => parseDiff(diff), [diff]);

  const lineColors = {
    add: 'bg-success/8 text-success',
    remove: 'bg-error/8 text-error',
    context: '',
    header: 'bg-velocity-500/8 text-velocity-500',
  };

  const linePrefix = {
    add: '+',
    remove: '-',
    context: ' ',
    header: '',
  };

  return (
    <div className={cn('rounded-lg border border-border-subtle overflow-hidden', className)}>
      <div className="px-3 py-1.5 bg-bg-tertiary border-b border-border-subtle">
        <span className="text-[10px] font-medium text-text-muted uppercase tracking-wider">Unified Diff</span>
      </div>
      <pre className="overflow-auto text-xs font-mono leading-relaxed max-h-[400px]">
        {lines.map((line, i) => (
          <div key={i} className={cn('flex px-3 py-0.5', lineColors[line.type])}>
            {line.type !== 'header' && (
              <>
                <span className="w-8 flex-shrink-0 text-right pr-2 text-text-muted select-none opacity-40">
                  {line.oldLineNum ?? ''}
                </span>
                <span className="w-8 flex-shrink-0 text-right pr-2 text-text-muted select-none opacity-40">
                  {line.newLineNum ?? ''}
                </span>
              </>
            )}
            <span className="w-4 flex-shrink-0 text-center select-none opacity-60">
              {linePrefix[line.type]}
            </span>
            <code className="flex-1 whitespace-pre">{line.content}</code>
          </div>
        ))}
      </pre>
    </div>
  );
}

// ──────────────────────────────────────────────────────────────────────────────
// Main DNAVariantDiff Component
// ──────────────────────────────────────────────────────────────────────────────

export interface DNAVariantDiffProps {
  mutation: DNAMutation;
  onAccept?: (canaryPercentage: number) => void;
  onReject?: (reason: string) => void;
  isAccepting?: boolean;
  isRejecting?: boolean;
  className?: string;
}

export function DNAVariantDiff({
  mutation,
  onAccept,
  onReject,
  isAccepting,
  isRejecting,
  className,
}: DNAVariantDiffProps) {
  const [viewMode, setViewMode] = useState<'split' | 'diff'>('diff');
  const [canaryPct, setCanaryPct] = useState(10);
  const [rejectReason, setRejectReason] = useState('');
  const [showReject, setShowReject] = useState(false);

  const typeMeta = MUTATION_TYPE_META[mutation.mutation_type];
  const statusMeta = MUTATION_STATUS_META[mutation.status];
  const isProposed = mutation.status === 'proposed';

  return (
    <div className={cn('space-y-4', className)}>
      {/* Header */}
      <div className="flex items-center justify-between flex-wrap gap-3">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <h3 className="text-base font-semibold text-text-primary">{typeMeta.label}</h3>
            <Badge variant={statusMeta.variant}>{statusMeta.label}</Badge>
            <span className="text-xs text-text-muted font-mono">Gen {mutation.generation}</span>
          </div>
          {mutation.trigger_reason && (
            <p className="text-sm text-text-secondary">{mutation.trigger_reason}</p>
          )}
        </div>

        {mutation.confidence != null && (
          <div className="text-right">
            <p className="text-[10px] text-text-muted uppercase tracking-wider">Confidence</p>
            <p className="text-lg font-bold font-mono text-velocity-500">
              {Math.round(mutation.confidence * 100)}%
            </p>
          </div>
        )}
      </div>

      {/* Impact comparison */}
      {mutation.estimated_impact && (
        <div className="rounded-xl border border-border-subtle bg-card p-4">
          <ImpactComparison estimated={mutation.estimated_impact} actual={mutation.actual_impact} />
        </div>
      )}

      {/* View mode toggle */}
      {mutation.diff && mutation.original_code && mutation.mutated_code && (
        <div className="flex items-center gap-1 rounded-lg bg-bg-tertiary p-0.5 w-fit">
          <button
            onClick={() => setViewMode('diff')}
            className={cn(
              'px-3 py-1 text-xs font-medium rounded-md transition-all',
              viewMode === 'diff'
                ? 'bg-card text-text-primary shadow-sm'
                : 'text-text-muted hover:text-text-secondary'
            )}
          >
            Unified Diff
          </button>
          <button
            onClick={() => setViewMode('split')}
            className={cn(
              'px-3 py-1 text-xs font-medium rounded-md transition-all',
              viewMode === 'split'
                ? 'bg-card text-text-primary shadow-sm'
                : 'text-text-muted hover:text-text-secondary'
            )}
          >
            Side by Side
          </button>
        </div>
      )}

      {/* Code view */}
      {viewMode === 'diff' && mutation.diff ? (
        <DiffLinesPanel diff={mutation.diff} />
      ) : viewMode === 'split' && mutation.original_code && mutation.mutated_code ? (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-3">
          <CodePanel code={mutation.original_code} title="Original" />
          <CodePanel code={mutation.mutated_code} title="Evolved" />
        </div>
      ) : mutation.original_code ? (
        <CodePanel code={mutation.original_code} title="Current Code" />
      ) : null}

      {/* Metadata */}
      <div className="flex flex-wrap gap-x-4 gap-y-1 text-[10px] text-text-muted">
        {mutation.model_used && (
          <span>Model: <span className="font-mono text-text-secondary">{mutation.model_used}</span></span>
        )}
        {mutation.executions_analyzed && (
          <span>Executions: <span className="font-mono text-text-secondary">{mutation.executions_analyzed.toLocaleString()}</span></span>
        )}
        {mutation.analysis_window_hours && (
          <span>Window: <span className="font-mono text-text-secondary">{mutation.analysis_window_hours}h</span></span>
        )}
        {mutation.original_hash && (
          <span>Original: <span className="font-mono text-text-secondary">{mutation.original_hash.slice(0, 12)}</span></span>
        )}
        {mutation.mutated_hash && (
          <span>Evolved: <span className="font-mono text-text-secondary">{mutation.mutated_hash.slice(0, 12)}</span></span>
        )}
      </div>

      {/* Accept / Reject actions */}
      {isProposed && (onAccept || onReject) && (
        <div className="rounded-xl border border-border-subtle bg-card p-4 space-y-3">
          <div className="flex items-center gap-4 flex-wrap">
            {onAccept && (
              <div className="flex items-center gap-2">
                <label className="text-xs text-text-muted">Canary %</label>
                <select
                  value={canaryPct}
                  onChange={(e) => setCanaryPct(Number(e.target.value))}
                  className="rounded-lg border border-border-subtle bg-bg-secondary px-2 py-1 text-xs text-text-primary"
                >
                  {[5, 10, 25, 50, 100].map((v) => (
                    <option key={v} value={v}>{v}%</option>
                  ))}
                </select>
                <button
                  onClick={() => onAccept(canaryPct)}
                  disabled={isAccepting}
                  className={cn(
                    'inline-flex items-center gap-1.5 rounded-lg px-4 py-2 text-sm font-medium transition-all',
                    'bg-gradient-to-r from-success to-emerald-400 text-white',
                    'hover:shadow-lg hover:shadow-success/20 disabled:opacity-50'
                  )}
                >
                  <Check className="h-4 w-4" />
                  {isAccepting ? 'Accepting...' : 'Accept & Deploy'}
                </button>
              </div>
            )}

            {onReject && !showReject && (
              <button
                onClick={() => setShowReject(true)}
                disabled={isRejecting}
                className="inline-flex items-center gap-1.5 rounded-lg px-4 py-2 text-sm font-medium border border-border-subtle bg-bg-tertiary text-text-secondary hover:text-text-primary hover:border-border-default transition-all disabled:opacity-50"
              >
                <X className="h-4 w-4" />
                Reject
              </button>
            )}
          </div>

          {/* Reject reason form */}
          <AnimatePresence>
            {showReject && (
              <motion.div
                initial={{ height: 0, opacity: 0 }}
                animate={{ height: 'auto', opacity: 1 }}
                exit={{ height: 0, opacity: 0 }}
                className="overflow-hidden"
              >
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={rejectReason}
                    onChange={(e) => setRejectReason(e.target.value)}
                    placeholder="Reason for rejection (optional)"
                    className="flex-1 rounded-lg border border-border-subtle bg-bg-secondary px-3 py-2 text-sm text-text-primary placeholder:text-text-muted focus:outline-none focus:ring-2 focus:ring-velocity-500/30"
                  />
                  <button
                    onClick={() => {
                      onReject?.(rejectReason);
                      setShowReject(false);
                    }}
                    disabled={isRejecting}
                    className="inline-flex items-center gap-1.5 rounded-lg px-3 py-2 text-sm font-medium bg-error/10 text-error hover:bg-error/20 transition-colors disabled:opacity-50"
                  >
                    Confirm
                  </button>
                  <button
                    onClick={() => setShowReject(false)}
                    className="rounded-lg px-3 py-2 text-sm text-text-muted hover:text-text-primary transition-colors"
                  >
                    Cancel
                  </button>
                </div>
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      )}
    </div>
  );
}
