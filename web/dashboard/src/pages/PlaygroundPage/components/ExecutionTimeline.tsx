import { motion } from 'framer-motion';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { TFunction } from 'i18next';
import { useTranslation } from 'react-i18next';
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

function getPhases(result: ExecutionResult, t: TFunction): Phase[] {
  const total = result.duration_ms || 1;

  // Estimate phase breakdown (real data would come from backend)
  const networkPercent = Math.min(15, (5 / total) * 100);
  const queuePercent = Math.min(10, (2 / total) * 100);
  const executePercent = 100 - networkPercent - queuePercent;

  return [
    {
      label: t('playground.network'),
      shortLabel: 'Net',
      color: 'text-blue-500 dark:text-blue-400',
      bgColor: 'bg-blue-500 dark:bg-blue-600',
      estimatedPercent: networkPercent,
      description: t('playground.networkDescription', { ms: Math.round((networkPercent / 100) * total) }),
    },
    {
      label: t('playground.queue'),
      shortLabel: 'Q',
      color: 'text-yellow-500 dark:text-yellow-400',
      bgColor: 'bg-yellow-500 dark:bg-yellow-600',
      estimatedPercent: queuePercent,
      description: t('playground.queueDescription', { ms: Math.round((queuePercent / 100) * total) }),
    },
    {
      label: t('playground.execute'),
      shortLabel: 'Exec',
      color: result.ok ? 'text-green-500 dark:text-green-400' : 'text-red-500 dark:text-red-400',
      bgColor: result.ok ? 'bg-green-500 dark:bg-green-600' : 'bg-red-500 dark:bg-red-600',
      estimatedPercent: executePercent,
      description: t('playground.executeDescription', { ms: Math.round((executePercent / 100) * total) }),
    },
  ];
}

export function ExecutionTimeline({ result }: ExecutionTimelineProps) {
  const { t } = useTranslation();
  const phases = getPhases(result, t);

  return (
    <div className="space-y-4">
      {/* Total duration */}
      <div className="flex items-center justify-between text-sm">
        <span className="text-text-secondary font-medium">{t('playground.totalDuration')}</span>
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
                    <span className="absolute inset-0 flex items-center justify-center text-[10px] font-medium text-white dark:text-gray-100">
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
          <p className="text-xs text-text-muted">{t('playground.status')}</p>
          <p
            className={`text-sm font-medium ${
              result.ok ? 'text-green-400' : 'text-red-400'
            }`}
          >
            {result.ok ? t('playground.successCheck') : t('playground.failedCheck')}
          </p>
        </div>

        <div className="space-y-1">
          <p className="text-xs text-text-muted">{t('playground.cache')}</p>
          <p className="text-sm font-medium">
            {result.cached ? (
              <span className="text-amber-400">{t('playground.hit')}</span>
            ) : (
              <span className="text-text-secondary">{t('playground.miss')}</span>
            )}
          </p>
        </div>

        {result.execution_id && (
          <div className="space-y-1 col-span-2">
            <p className="text-xs text-text-muted">{t('playground.executionId')}</p>
            <p className="text-xs font-mono text-text-secondary truncate">
              {result.execution_id}
            </p>
          </div>
        )}

        <div className="space-y-1">
          <p className="text-xs text-text-muted">{t('playground.version')}</p>
          <p className="text-sm font-mono text-text-secondary">v{result.version}</p>
        </div>
      </div>
    </div>
  );
}
