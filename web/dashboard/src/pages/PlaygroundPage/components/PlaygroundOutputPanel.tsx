import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  CheckCircle2,
  XCircle,
  Clock,
  Database,
  Activity,
  GitCompare,
  Play,
  Copy,
  Check,
  Search,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { cn } from '@/lib/utils';
import { usePlaygroundStore, OutputTab } from '../store/playgroundStore';
import { JsonTreeViewer } from './JsonTreeViewer';
import { ExecutionTimeline } from './ExecutionTimeline';
import { DiffViewer } from './DiffViewer';
import { LatencyChart } from './LatencyChart';

interface PlaygroundOutputPanelProps {
  className?: string;
}

function EmptyState() {
  const { t } = useTranslation();
  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      className="flex flex-col items-center justify-center h-full py-16 text-center"
    >
      <div className="w-16 h-16 rounded-full bg-indigo-500/10 flex items-center justify-center mb-4">
        <Play className="w-7 h-7 text-indigo-400" />
      </div>
      <p className="text-sm font-medium text-text-secondary">{t('playground.runFunctionToSeeResults')}</p>
      <p className="text-xs text-text-muted mt-1">
        {t('playground.pressOrClickRun', { shortcut: '⌘↵' })}
      </p>
    </motion.div>
  );
}

function LoadingState() {
  const { t } = useTranslation();
  return (
    <div className="flex flex-col items-center justify-center h-full py-16">
      <div className="w-full max-w-xs space-y-3">
        <div className="h-2 bg-bg-tertiary rounded-full overflow-hidden">
          <motion.div
            className="h-full bg-indigo-500 rounded-full"
            initial={{ width: '0%' }}
            animate={{ width: '100%' }}
            transition={{ duration: 2, ease: 'easeInOut', repeat: Infinity }}
          />
        </div>
        <p className="text-xs text-text-muted text-center">{t('playground.executingFunction')}</p>
      </div>
    </div>
  );
}

export function PlaygroundOutputPanel({ className }: PlaygroundOutputPanelProps) {
  const { t } = useTranslation();
  const {
    executionResult,
    isExecuting,
    activeOutputTab,
    setActiveOutputTab,
    settings,
  } = usePlaygroundStore();

  const [searchQuery, setSearchQuery] = useState('');
  const [copiedAll, setCopiedAll] = useState(false);

  const handleCopyAll = async () => {
    if (!executionResult?.data) return;
    try {
      await navigator.clipboard.writeText(JSON.stringify(executionResult.data, null, 2));
      setCopiedAll(true);
      setTimeout(() => setCopiedAll(false), 2000);
    } catch {
      // ignore
    }
  };

  return (
    <div className={cn('flex flex-col h-full', className)}>
      <Tabs
        value={activeOutputTab}
        onValueChange={(v) => setActiveOutputTab(v as OutputTab)}
        className="flex flex-col h-full"
      >
        {/* Tab bar */}
        <div className="flex items-center border-b border-border-subtle px-3 pt-1 bg-bg-secondary shrink-0">
          <TabsList className="h-8 bg-transparent gap-0 p-0">
            <TabsTrigger
              value="response"
              className="h-8 px-3 text-xs rounded-none border-b-2 border-transparent data-[state=active]:border-indigo-500 data-[state=active]:text-indigo-400 data-[state=active]:bg-transparent gap-1.5"
            >
              {executionResult ? (
                executionResult.ok ? (
                  <CheckCircle2 className="w-3.5 h-3.5 text-green-400" />
                ) : (
                  <XCircle className="w-3.5 h-3.5 text-red-400" />
                )
              ) : null}
              {t('playground.response')}
            </TabsTrigger>

            {settings.showHeaders && (
              <TabsTrigger
                value="headers"
                className="h-8 px-3 text-xs rounded-none border-b-2 border-transparent data-[state=active]:border-indigo-500 data-[state=active]:text-indigo-400 data-[state=active]:bg-transparent gap-1.5"
              >
                <Database className="w-3.5 h-3.5" />
                {t('playground.headers')}
              </TabsTrigger>
            )}

            {settings.showTimeline && (
              <TabsTrigger
                value="timeline"
                className="h-8 px-3 text-xs rounded-none border-b-2 border-transparent data-[state=active]:border-indigo-500 data-[state=active]:text-indigo-400 data-[state=active]:bg-transparent gap-1.5"
              >
                <Activity className="w-3.5 h-3.5" />
                {t('playground.timeline')}
              </TabsTrigger>
            )}

            <TabsTrigger
              value="diff"
              className="h-8 px-3 text-xs rounded-none border-b-2 border-transparent data-[state=active]:border-indigo-500 data-[state=active]:text-indigo-400 data-[state=active]:bg-transparent gap-1.5"
            >
              <GitCompare className="w-3.5 h-3.5" />
              {t('playground.diff')}
            </TabsTrigger>
          </TabsList>

          {/* Status badge */}
          {executionResult && (
            <div className="ml-2 flex items-center gap-2">
              <Badge
                variant="outline"
                className={cn(
                  'text-[10px] px-1.5 py-0 h-5 border',
                  executionResult.ok
                    ? 'bg-green-500/10 text-green-400 border-green-500/20'
                    : 'bg-red-500/10 text-red-400 border-red-500/20'
                )}
              >
                {executionResult.ok ? '200 OK' : t('playground.error')}
              </Badge>
              <span className="text-xs text-text-muted flex items-center gap-1">
                <Clock className="w-3 h-3" />
                {executionResult.duration_ms}ms
              </span>
              {executionResult.cached && (
                <Badge
                  variant="outline"
                  className="text-[10px] px-1.5 py-0 h-5 border bg-amber-500/10 text-amber-400 border-amber-500/20"
                >
                  {t('playground.cached')}
                </Badge>
              )}
            </div>
          )}

          <div className="ml-auto text-xs text-text-muted pr-1">{t('playground.output')}</div>
        </div>

        {/* Tab content */}
        <div className="flex-1 overflow-hidden">
          {/* Response tab */}
          <TabsContent value="response" className="h-full m-0 flex flex-col" forceMount hidden={activeOutputTab !== 'response'}>
            {isExecuting ? (
              <LoadingState />
            ) : !executionResult ? (
              <EmptyState />
            ) : executionResult.ok ? (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className="flex flex-col h-full"
              >
                {/* Search + copy toolbar */}
                <div className="flex items-center gap-2 px-3 py-2 border-b border-border-subtle shrink-0">
                  <div className="relative flex-1">
                    <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-3 h-3 text-text-muted" />
                    <Input
                      value={searchQuery}
                      onChange={(e) => setSearchQuery(e.target.value)}
                      placeholder={t('playground.searchResponse')}
                      className="h-7 pl-7 text-xs bg-bg-tertiary border-border-subtle"
                    />
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={handleCopyAll}
                    className="h-7 gap-1.5 text-xs text-text-muted hover:text-text-primary shrink-0"
                  >
                    {copiedAll ? (
                      <Check className="w-3 h-3 text-green-400" />
                    ) : (
                      <Copy className="w-3 h-3" />
                    )}
                    {copiedAll ? t('playground.copied') : t('playground.copyAll')}
                  </Button>
                </div>

                {/* JSON tree */}
                <div className="flex-1 overflow-auto p-2">
                  <JsonTreeViewer
                    data={executionResult.data}
                    searchQuery={searchQuery}
                    className="min-h-full"
                  />
                </div>

                {/* Latency sparkline */}
                <div className="border-t border-border-subtle p-3 shrink-0">
                  <LatencyChart />
                </div>
              </motion.div>
            ) : (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className="p-4"
              >
                <div className="border border-red-500/20 bg-red-500/5 rounded-lg p-4">
                  <div className="flex items-start gap-3">
                    <XCircle className="w-5 h-5 text-red-400 shrink-0 mt-0.5" />
                    <div>
                      <p className="text-sm font-medium text-red-400">
                        {executionResult.error?.code || t('playground.executionFailed')}
                      </p>
                      <p className="text-sm text-red-300/80 mt-1">
                        {executionResult.error?.message || t('playground.unknownError')}
                      </p>
                    </div>
                  </div>
                </div>
              </motion.div>
            )}
          </TabsContent>

          {/* Headers tab */}
          <TabsContent value="headers" className="h-full m-0 overflow-auto" forceMount hidden={activeOutputTab !== 'headers'}>
            {!executionResult ? (
              <EmptyState />
            ) : (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className="p-4 space-y-3"
              >
                <div className="grid grid-cols-2 gap-3">
                  {[
                    { label: t('playground.status'), value: executionResult.ok ? t('playground.success') : t('playground.error'), color: executionResult.ok ? 'text-green-400' : 'text-red-400' },
                    { label: t('playground.duration'), value: `${executionResult.duration_ms}ms`, color: 'text-text-primary' },
                    { label: t('playground.cache'), value: executionResult.cached ? t('playground.hit') : t('playground.miss'), color: executionResult.cached ? 'text-amber-400' : 'text-text-secondary' },
                    { label: t('playground.version'), value: `v${executionResult.version}`, color: 'text-text-primary' },
                  ].map((item) => (
                    <div key={item.label} className="bg-bg-tertiary rounded-lg p-3">
                      <p className="text-xs text-text-muted mb-1">{item.label}</p>
                      <p className={cn('text-sm font-medium font-mono', item.color)}>
                        {item.value}
                      </p>
                    </div>
                  ))}
                </div>

                {executionResult.execution_id && (
                  <div className="bg-bg-tertiary rounded-lg p-3">
                    <p className="text-xs text-text-muted mb-1">{t('playground.executionId')}</p>
                    <p className="text-xs font-mono text-text-secondary break-all">
                      {executionResult.execution_id}
                    </p>
                  </div>
                )}
              </motion.div>
            )}
          </TabsContent>

          {/* Timeline tab */}
          <TabsContent value="timeline" className="h-full m-0 overflow-auto" forceMount hidden={activeOutputTab !== 'timeline'}>
            {!executionResult ? (
              <EmptyState />
            ) : (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className="p-4"
              >
                <ExecutionTimeline result={executionResult} />
              </motion.div>
            )}
          </TabsContent>

          {/* Diff tab */}
          <TabsContent value="diff" className="h-full m-0 overflow-auto" forceMount hidden={activeOutputTab !== 'diff'}>
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="p-4"
            >
              <DiffViewer />
            </motion.div>
          </TabsContent>
        </div>
      </Tabs>
    </div>
  );
}
