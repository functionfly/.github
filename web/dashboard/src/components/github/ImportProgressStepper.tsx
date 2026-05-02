import { useMemo } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  ScanSearch,
  Download,
  Hammer,
  Rocket,
  Check,
  X,
  Loader2,
  ExternalLink,
  RotateCcw,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { cn } from '@/lib/utils';
import { useImportProgress, useRetryImport } from '@/hooks/useGitHubImport';
import type { GitHubImport } from '@/types/github';

interface StepConfig {
  id: string;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
  statuses: GitHubImport['status'][];
}

const STEPS: StepConfig[] = [
  { id: 'scanning', label: 'Scanning', icon: ScanSearch, statuses: ['pending', 'scanning'] },
  { id: 'fetching', label: 'Fetching', icon: Download, statuses: ['fetching'] },
  { id: 'building', label: 'Building', icon: Hammer, statuses: ['configuring', 'building'] },
  { id: 'publishing', label: 'Publishing', icon: Rocket, statuses: ['publishing'] },
];

type StepStatus = 'pending' | 'active' | 'complete' | 'error';

function getStepStatus(stepIndex: number, currentStatus: GitHubImport['status'] | undefined, error: boolean): StepStatus {
  if (error) {
    const errorStepIndex = STEPS.findIndex((s) => s.statuses.includes(currentStatus ?? 'pending'));
    if (stepIndex < errorStepIndex) return 'complete';
    if (stepIndex === errorStepIndex) return 'error';
    return 'pending';
  }

  if (currentStatus === 'completed') return 'complete';

  const currentStepIndex = STEPS.findIndex((s) => s.statuses.includes(currentStatus ?? 'pending'));

  if (stepIndex < currentStepIndex) return 'complete';
  if (stepIndex === currentStepIndex) return 'active';
  return 'pending';
}

interface StepCircleProps {
  status: StepStatus;
  icon: React.ComponentType<{ className?: string }>;
  index: number;
}

function StepCircle({ status, icon: Icon, index }: StepCircleProps) {
  return (
    <div
      className={cn(
        'relative flex h-10 w-10 items-center justify-center rounded-full border-2 transition-all duration-300',
        status === 'pending' && 'border-border-subtle bg-transparent text-text-muted',
        status === 'active' && 'border-[#00D4FF] bg-[#00D4FF]/10 text-[#00D4FF]',
        status === 'complete' && 'border-emerald-500 bg-emerald-500 text-white',
        status === 'error' && 'border-red-500 bg-red-500/10 text-red-500'
      )}
    >
      {status === 'complete' && <Check className="h-5 w-5" />}
      {status === 'error' && <X className="h-5 w-5" />}
      {status === 'active' && <Loader2 className="h-5 w-5 animate-spin" />}
      {status === 'pending' && <Icon className="h-4 w-4" />}
    </div>
  );
}

interface ImportProgressStepperProps {
  importId: string;
  functionId?: string;
  functionName?: string;
  onComplete?: () => void;
  className?: string;
}

export function ImportProgressStepper({
  importId,
  functionId,
  functionName,
  onComplete,
  className,
}: ImportProgressStepperProps) {
  const { progress, complete, error, status } = useImportProgress(importId);
  const retryMutation = useRetryImport();

  const currentImportStatus = progress?.stage as GitHubImport['status'] | undefined;
  const isComplete = status === 'completed' || complete !== null;
  const isError = status === 'failed' || error !== null;

  const activeStepIndex = useMemo(() => {
    if (isComplete) return STEPS.length - 1;
    return STEPS.findIndex((s) => s.statuses.includes(currentImportStatus ?? 'pending'));
  }, [currentImportStatus, isComplete]);

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
      className={className}
    >
      <Card>
        <CardHeader className="pb-4">
          <CardTitle className="text-base flex items-center gap-2">
            Importing {functionName ?? 'Function'}
            {isComplete && (
              <Badge variant="success" className="text-xs">Complete</Badge>
            )}
            {isError && (
              <Badge variant="error" className="text-xs">Failed</Badge>
            )}
          </CardTitle>
        </CardHeader>

        <CardContent className="space-y-6">
          <div className="flex items-center justify-between">
            {STEPS.map((step, index) => {
              const stepStatus = getStepStatus(index, currentImportStatus, isError);
              return (
                <div key={step.id} className="flex items-center flex-1 last:flex-initial">
                  <div className="flex flex-col items-center gap-1.5">
                    <StepCircle status={stepStatus} icon={step.icon} index={index} />
                    <span
                      className={cn(
                        'text-[11px] font-medium',
                        stepStatus === 'active' && 'text-[#00D4FF]',
                        stepStatus === 'complete' && 'text-emerald-500',
                        stepStatus === 'error' && 'text-red-500',
                        stepStatus === 'pending' && 'text-text-muted'
                      )}
                    >
                      {step.label}
                    </span>
                  </div>
                  {index < STEPS.length - 1 && (
                    <div
                      className={cn(
                        'h-0.5 flex-1 mx-2 rounded transition-colors duration-300 mt-[-20px]',
                        index < activeStepIndex || isComplete ? 'bg-emerald-500' : 'bg-border-subtle'
                      )}
                    />
                  )}
                </div>
              );
            })}
          </div>

          {progress && !isComplete && !isError && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="space-y-2"
            >
              <div className="flex items-center justify-between text-xs">
                <span className="text-text-secondary">{progress.message}</span>
                <span className="text-text-muted font-mono">{progress.progress}%</span>
              </div>
              <Progress
                value={progress.progress}
                className="h-2"
                indicatorClassName="bg-[#00D4FF]"
              />
            </motion.div>
          )}

          <AnimatePresence>
            {isComplete && complete && (
              <motion.div
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.95 }}
                className="rounded-lg border border-emerald-500/30 bg-emerald-500/5 p-4 space-y-3"
              >
                <div className="flex items-center gap-2 text-emerald-500">
                  <Check className="h-5 w-5" />
                  <span className="text-sm font-semibold">Import Complete</span>
                </div>
                <div className="grid grid-cols-3 gap-3 text-xs">
                  <div>
                    <span className="text-text-muted">Files</span>
                    <p className="font-semibold text-text-primary">{complete.files_imported ?? 0}</p>
                  </div>
                  <div>
                    <span className="text-text-muted">Function</span>
                    <p className="font-semibold text-text-primary font-mono">{complete.function_name ?? '—'}</p>
                  </div>
                  <div>
                    <span className="text-text-muted">Commit</span>
                    <p className="font-semibold text-text-primary font-mono">{complete.commit_sha?.slice(0, 8) ?? '—'}</p>
                  </div>
                </div>
                {functionId && (
                  <Button variant="outline" size="sm" className="w-full" asChild>
                    <a href={`/functions/${functionId}`} aria-label="View imported function">
                      <ExternalLink className="h-3.5 w-3.5 mr-1.5" />
                      View Function
                    </a>
                  </Button>
                )}
              </motion.div>
            )}
          </AnimatePresence>

          <AnimatePresence>
            {isError && error && (
              <motion.div
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.95 }}
                className="rounded-lg border border-red-500/30 bg-red-500/5 p-4 space-y-3"
              >
                <div className="flex items-center gap-2 text-red-500">
                  <X className="h-5 w-5" />
                  <span className="text-sm font-semibold">Import Failed</span>
                </div>
                <p className="text-xs text-text-secondary">{error.message}</p>
                {error.details?.failed_stage && (
                  <p className="text-[10px] text-text-muted">
                    Failed at stage: <span className="font-mono">{error.details.failed_stage}</span>
                  </p>
                )}
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => retryMutation.mutate(importId)}
                  disabled={retryMutation.isPending}
                  className="w-full"
                  aria-label="Retry import"
                >
                  <RotateCcw className={cn('h-3.5 w-3.5 mr-1.5', retryMutation.isPending && 'animate-spin')} />
                  Retry
                </Button>
              </motion.div>
            )}
          </AnimatePresence>

          {status === 'connecting' && (
            <div className="flex items-center justify-center gap-2 text-xs text-text-muted py-2">
              <Loader2 className="h-3.5 w-3.5 animate-spin" />
              Connecting to import stream...
            </div>
          )}
        </CardContent>
      </Card>
    </motion.div>
  );
}
