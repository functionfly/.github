import { motion } from 'framer-motion';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { ExecutionResult } from '../store/playgroundStore';

interface ExecutionTimelineProps {
  result: ExecutionResult;
}

interface Phase {
  label: string;
  shortLabel: string;
  color: string;
  bgColor: string;
  estimatedPercent: number;
  description: string;
}

function getPhases(result: ExecutionResult): Phase[] {
  const total = result.duration_ms || 1;

  // Estimate phase breakdown (real data would come from backend)
  const networkPercent = Math.min(15, (5 / total) * 100);
  const queuePercent = Math.min(10, (2 / total) * 100);
  const executePercent = 100 - networkPercent - queuePercent;

  return [
    {
      label: 'Network',
      shortLabel: 'Net',
      color: 'text-blue-400',
      bgColor: 'bg-blue-500',
      estimatedPercent: networkPercent,
      description: `~${Math.round((networkPercent / 100) * total)}ms — Request routing and network overhead`,
    },
    {
      label: 'Queue',
      shortLabel: 'Q',
      color: 'text-yellow-400',
      bgColor: 'bg-yellow-500',
      estimatedPercent: queuePercent,
      description: `~${Math.round((queuePercent / 100) * total)}ms — Function scheduling and queue wait`,
    },
    {
      label: 'Execute',
      shortLabel: 'Exec',
      color: result.ok ? 'text-green-400' : 'text-red-400',
      bgColor: result.ok ? 'bg-green-500' : 'bg-red-500',
      estimatedPercent: executePercent,
      description: `~${Math.round((executePercent / 100) * total)}ms — Function execution time`,
    },
  ];
}

export function ExecutionTimeline({ result }: ExecutionTimelineProps) {
  const phases = getPhases(result);

  return (
    <div className="space-y-4">
      {/* Total duration */}
      <div className="flex items-center justify-between text-sm">
        <span className="text-text-secondary font-medium">Total Duration</span>
        <span
          className={`font-mono font-semibold ${
            result.ok ? 'text-green-400' : 'text-red-400'
          }`}
        >
          {result.duration_ms}ms
        </span>
      </div>

      {/* Timeline bar */}
      <TooltipProvider>
        <div className="relative h-8 rounded-md overflow-hidden bg-bg-tertiary flex">
          {phases.map((phase, i) => (
            <Tooltip key={i}>
              <TooltipTrigger asChild>
                <motion.div
                  initial={{ width: 0 }}
                  animate={{ width: `${phase.estimatedPercent}%` }}
                  transition={{ duration: 0.6, delay: i * 0.1, ease: 'easeOut' }}
                  className={`${phase.bgColor} opacity-80 hover:opacity-100 transition-opacity cursor-default relative`}
                  style={{ minWidth: '2px' }}
                >
                  {phase.estimatedPercent > 8 && (
                    <span className="absolute inset-0 flex items-center justify-center text-[10px] font-medium text-white">
                      {phase.shortLabel}
                    </span>
                  )}
                </motion.div>
              </TooltipTrigger>
              <TooltipContent>
                <div className="text-xs">
                  <p className="font-medium">{phase.label}</p>
                  <p className="text-text-muted">{phase.description}</p>
                </div>
              </TooltipContent>
            </Tooltip>
          ))}
        </div>
      </TooltipProvider>

      {/* Phase legend */}
      <div className="flex flex-wrap gap-3">
        {phases.map((phase, i) => (
          <div key={i} className="flex items-center gap-1.5 text-xs">
            <div className={`w-2.5 h-2.5 rounded-sm ${phase.bgColor} opacity-80`} />
            <span className={phase.color}>{phase.label}</span>
            <span className="text-text-muted">
              {Math.round((phase.estimatedPercent / 100) * result.duration_ms)}ms
            </span>
          </div>
        ))}
      </div>

      {/* Metadata */}
      <div className="grid grid-cols-2 gap-3 pt-2 border-t border-border-subtle">
        <div className="space-y-1">
          <p className="text-xs text-text-muted">Status</p>
          <p
            className={`text-sm font-medium ${
              result.ok ? 'text-green-400' : 'text-red-400'
            }`}
          >
            {result.ok ? '✓ Success' : '✗ Failed'}
          </p>
        </div>

        <div className="space-y-1">
          <p className="text-xs text-text-muted">Cache</p>
          <p className="text-sm font-medium">
            {result.cached ? (
              <span className="text-amber-400">Hit</span>
            ) : (
              <span className="text-text-secondary">Miss</span>
            )}
          </p>
        </div>

        {result.execution_id && (
          <div className="space-y-1 col-span-2">
            <p className="text-xs text-text-muted">Execution ID</p>
            <p className="text-xs font-mono text-text-secondary truncate">
              {result.execution_id}
            </p>
          </div>
        )}

        <div className="space-y-1">
          <p className="text-xs text-text-muted">Version</p>
          <p className="text-sm font-mono text-text-secondary">v{result.version}</p>
        </div>
      </div>
    </div>
  );
}
