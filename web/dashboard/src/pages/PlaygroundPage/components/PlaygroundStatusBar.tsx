import { motion } from 'framer-motion';
import { Clock, Zap, Database, Hash, Keyboard } from 'lucide-react';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { useTranslation } from 'react-i18next';
import { usePlaygroundStore } from '../store/playgroundStore';
import { usePlaygroundState } from '../hooks/usePlaygroundState';
import { KEYBOARD_SHORTCUTS } from '../hooks/usePlaygroundKeyboard';

export function PlaygroundStatusBar() {
  const { t } = useTranslation();
  const { executionResult, functionInfo } = usePlaygroundStore();
  const { averageLatency, successRate } = usePlaygroundState();

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.2, delay: 0.2 }}
      className="flex items-center gap-4 px-4 py-1.5 border-t border-border-subtle bg-bg-secondary dark:bg-bg-tertiary text-xs text-text-muted"
    >
      {/* Last run time */}
      {executionResult && (
        <div className="flex items-center gap-1.5">
          <Clock className="w-3 h-3" />
          <span>
            {t('playground.lastRun')}{' '}
            <span
              className={
                executionResult.ok ? 'text-green-500 dark:text-green-400' : 'text-red-500 dark:text-red-400'
              }
            >
              {executionResult.duration_ms}ms
            </span>
          </span>
        </div>
      )}

      {/* Average latency */}
      {averageLatency !== null && (
        <div className="flex items-center gap-1.5">
          <Zap className="w-3 h-3" />
          <span>{t('playground.avg')} {averageLatency}ms</span>
        </div>
      )}

      {/* Success rate */}
      {successRate !== null && (
        <div className="flex items-center gap-1.5">
          <span
            className={
              successRate >= 90
                ? 'text-green-500 dark:text-green-400'
                : successRate >= 70
                ? 'text-yellow-500 dark:text-yellow-400'
                : 'text-red-500 dark:text-red-400'
            }
            >
              {t('playground.successRate', { rate: successRate })}
            </span>
        </div>
      )}

      {/* Cache status */}
      {executionResult?.cached && (
        <div className="flex items-center gap-1.5 text-amber-500 dark:text-amber-400">
          <Database className="w-3 h-3" />
          <span>{t('playground.cached')}</span>
        </div>
      )}

      {/* Version */}
      {executionResult?.version && (
        <div className="flex items-center gap-1.5">
          <Hash className="w-3 h-3" />
          <span>v{executionResult.version}</span>
        </div>
      )}

      <div className="flex-1" />

      {/* Function info */}
      {functionInfo && (
        <span className="text-text-muted">
          {functionInfo.author}/{functionInfo.name}
        </span>
      )}

      {/* Keyboard shortcuts hint */}
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <button className="flex items-center gap-1 hover:text-text-secondary transition-colors">
              <Keyboard className="w-3 h-3" />
              <span>{t('playground.shortcuts')}</span>
            </button>
          </TooltipTrigger>
          <TooltipContent side="top" className="p-3 max-w-xs">
            <div className="space-y-1.5">
               <p className="text-xs font-medium mb-2">{t('playground.keyboardShortcuts')}</p>
              {KEYBOARD_SHORTCUTS.map((shortcut, i) => (
                <div key={i} className="flex items-center justify-between gap-4 text-xs">
                  <span className="text-text-muted">{shortcut.description}</span>
                  <div className="flex items-center gap-0.5">
                    {shortcut.keys.map((key, j) => (
                      <kbd
                        key={j}
                        className="px-1 py-0.5 bg-bg-tertiary border border-border-subtle rounded text-[10px] font-mono"
                      >
                        {key}
                      </kbd>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </motion.div>
  );
}
