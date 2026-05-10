import { FunctionCodeViewer } from '@/components/functions';
import { DNAWidget } from '@/components/dna';
import { Button } from '@/components/ui/button';
import { useDNAProfile, useDNAMutations } from '@/hooks/useFunctionDNA';
import { motion } from 'framer-motion';
import { AlertCircle, Code2, Loader2, RefreshCw } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';
import type { FunctionInfo } from './types';

interface SourceSectionProps {
  functionInfo: FunctionInfo;
  sourceCode: string | null | undefined;
  isLoadingSource: boolean;
  sourceError?: Error | null;
  onRetrySource?: () => void;
}

export function SourceSection({
  functionInfo,
  sourceCode,
  isLoadingSource,
  sourceError,
  onRetrySource,
}: SourceSectionProps) {
  const { data: dnaProfile, isLoading: dnaLoading } = useDNAProfile(functionInfo.id || '', 'registry');
  const { data: dnaMutationsData } = useDNAMutations(functionInfo.id || '', { limit: 3 });

  if (!sourceCode && !isLoadingSource && !functionInfo.id) return null;

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, delay: 0.6 }}
      className="function-page-section"
    >
      <div className="grid grid-cols-1 lg:grid-cols-[1fr_320px] gap-6">
        {(sourceCode || isLoadingSource || sourceError) && (
          <div>
            <div className="mb-4 flex items-center gap-3">
              <div className="w-10 h-10 rounded-xl bg-brand-500/10 flex items-center justify-center border border-brand-500/20">
                <Code2 className="w-5 h-5 text-brand-500" />
              </div>
              <div>
                <h2 className="text-xl font-bold text-foreground">Function Source</h2>
                <p className="text-sm text-muted-foreground">Verified implementation source code</p>
              </div>
            </div>
            {isLoadingSource ? (
              <div className="flex items-center justify-center h-32 rounded-xl border border-border-subtle bg-bg-secondary">
                <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
              </div>
            ) : sourceError ? (
              <div className="flex flex-col items-center justify-center h-32 rounded-xl border border-border-subtle bg-bg-secondary gap-3">
                <AlertCircle className="w-6 h-6 text-muted-foreground" />
                <p className="text-sm text-muted-foreground">Source unavailable</p>
                {onRetrySource && (
                  <Button variant="outline" size="sm" onClick={onRetrySource} className="gap-2">
                    <RefreshCw className="w-4 h-4" />
                    Retry
                  </Button>
                )}
              </div>
            ) : sourceCode ? (
              <FunctionCodeViewer
                code={sourceCode}
                runtime={functionInfo.runtime || 'text'}
                functionName={functionInfo.name || 'function'}
                version={functionInfo.version}
                lastModified={
                  functionInfo.updated_at
                    ? formatDistanceToNow(new Date(functionInfo.updated_at), { addSuffix: true })
                    : undefined
                }
              />
            ) : null}
          </div>
        )}

        {functionInfo.id && (dnaLoading || dnaProfile !== null) && (
          <div className={sourceCode ? '' : 'lg:col-start-1 lg:col-end-2 lg:max-w-sm'}>
            <DNAWidget
              functionId={functionInfo.id}
              functionSlug={`${functionInfo.author}/${functionInfo.name}`}
              profile={dnaProfile ?? null}
              recentMutations={dnaMutationsData?.mutations || []}
              isLoading={dnaLoading}
            />
          </div>
        )}
      </div>
    </motion.div>
  );
}