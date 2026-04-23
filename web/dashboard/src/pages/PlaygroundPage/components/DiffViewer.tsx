import { useState } from 'react';
import { motion } from 'framer-motion';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Plus, Minus, Equal, GitCompare } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { cn } from '@/lib/utils';
import { usePlaygroundStore, ExecutionHistoryItem } from '../store/playgroundStore';
import { diffJson, DiffNode, countDiffChanges } from '../utils/jsonDiff';

interface DiffViewerProps {
  className?: string;
}

function DiffNodeRow({ node, depth }: { node: DiffNode; depth: number }) {
  const [isExpanded, setIsExpanded] = useState(true);

  const typeStyles = {
    added: 'bg-green-500/10 border-l-2 border-green-500',
    removed: 'bg-red-500/10 border-l-2 border-red-500',
    changed: 'bg-yellow-500/10 border-l-2 border-yellow-500',
    unchanged: '',
  };

  const typeIcons = {
    added: <Plus className="w-3 h-3 text-green-400 shrink-0" />,
    removed: <Minus className="w-3 h-3 text-red-400 shrink-0" />,
    changed: <Equal className="w-3 h-3 text-yellow-400 shrink-0" />,
    unchanged: <span className="w-3 h-3 shrink-0" />,
  };

  const hasChildren = node.children && node.children.length > 0;

  return (
    <div>
      <div
        className={cn(
          'flex items-start gap-2 py-0.5 px-2 text-xs font-mono rounded-sm',
          typeStyles[node.type],
          hasChildren && 'cursor-pointer'
        )}
        style={{ paddingLeft: `${depth * 16 + 8}px` }}
        onClick={() => hasChildren && setIsExpanded((p) => !p)}
      >
        {typeIcons[node.type]}

        <span className="text-purple-300">{node.key}</span>
        <span className="text-text-muted">:</span>

        {node.type === 'changed' && !hasChildren && (
          <span className="flex items-center gap-1.5">
            <span className="text-red-400 line-through">
              {JSON.stringify(node.leftValue)}
            </span>
            <span className="text-text-muted">→</span>
            <span className="text-green-400">{JSON.stringify(node.rightValue)}</span>
          </span>
        )}

        {node.type === 'added' && (
          <span className="text-green-400">{JSON.stringify(node.rightValue)}</span>
        )}

        {node.type === 'removed' && (
          <span className="text-red-400 line-through">
            {JSON.stringify(node.leftValue)}
          </span>
        )}

        {node.type === 'unchanged' && !hasChildren && (
          <span className="text-text-muted">{JSON.stringify(node.leftValue)}</span>
        )}

        {hasChildren && (
          <span className="text-text-muted">{isExpanded ? '{...}' : `{${node.children!.length} fields}`}</span>
        )}
      </div>

      {hasChildren && isExpanded && (
        <div>
          {node.children!.map((child, i) => (
            <DiffNodeRow key={i} node={child} depth={depth + 1} />
          ))}
        </div>
      )}
    </div>
  );
}

export function DiffViewer({ className }: DiffViewerProps) {
  const { t } = useTranslation();
  const { executionHistory, executionResult, diffBaseItem, setDiffBaseItem } =
    usePlaygroundStore();

  const currentResult = executionResult;

  if (!currentResult) {
    return (
      <div className={cn('flex flex-col items-center justify-center py-12 text-center', className)}>
        <GitCompare className="w-10 h-10 text-text-muted mb-3" />
        <p className="text-sm text-text-muted">{t('playground.runFunctionToCompare')}</p>
      </div>
    );
  }

  if (executionHistory.length < 2) {
    return (
      <div className={cn('flex flex-col items-center justify-center py-12 text-center', className)}>
        <GitCompare className="w-10 h-10 text-text-muted mb-3" />
        <p className="text-sm text-text-muted">{t('playground.runTwiceToCompare')}</p>
      </div>
    );
  }

  const baseItem = diffBaseItem || executionHistory[1]; // Default to second-most-recent
  const diffNodes = diffJson(baseItem.result.data, currentResult.data);
  const { added, removed, changed } = countDiffChanges(diffNodes);

  return (
    <div className={cn('space-y-3', className)}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-xs text-text-muted">{t('playground.comparingWith')}</span>
          <select
            className="text-xs bg-bg-tertiary border border-border-subtle rounded px-2 py-1 text-text-primary"
            value={baseItem.id}
            onChange={(e) => {
              const item = executionHistory.find((h) => h.id === e.target.value);
              if (item) setDiffBaseItem(item);
            }}
          >
            {executionHistory.slice(1).map((item) => (
              <option key={item.id} value={item.id}>
                {new Date(item.timestamp).toLocaleTimeString()} —{' '}
                {item.result.ok ? '✓' : '✗'} {item.result.duration_ms}ms
              </option>
            ))}
          </select>
        </div>

        {/* Change summary */}
        <div className="flex items-center gap-2">
          {added > 0 && (
            <Badge className="text-[10px] bg-green-500/10 text-green-400 border-green-500/20">
              +{added}
            </Badge>
          )}
          {removed > 0 && (
            <Badge className="text-[10px] bg-red-500/10 text-red-400 border-red-500/20">
              -{removed}
            </Badge>
          )}
          {changed > 0 && (
            <Badge className="text-[10px] bg-yellow-500/10 text-yellow-400 border-yellow-500/20">
              ~{changed}
            </Badge>
          )}
          {added === 0 && removed === 0 && changed === 0 && (
            <Badge className="text-[10px] bg-bg-tertiary text-text-muted border-border-subtle">
              {t('playground.noChanges')}
            </Badge>
          )}
        </div>
      </div>

      {/* Diff tree */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        className="bg-bg-secondary rounded-md p-2 overflow-auto max-h-96"
      >
        {diffNodes.length === 0 ? (
          <p className="text-xs text-text-muted text-center py-4">{t('playground.resultsIdentical')}</p>
        ) : (
          diffNodes.map((node, i) => (
            <DiffNodeRow key={i} node={node} depth={0} />
          ))
        )}
      </motion.div>

      {/* Legend */}
      <div className="flex items-center gap-4 text-xs text-text-muted">
        <div className="flex items-center gap-1">
          <Plus className="w-3 h-3 text-green-400" />
          <span>{t('playground.added')}</span>
        </div>
        <div className="flex items-center gap-1">
          <Minus className="w-3 h-3 text-red-400" />
          <span>{t('playground.removed')}</span>
        </div>
        <div className="flex items-center gap-1">
          <Equal className="w-3 h-3 text-yellow-400" />
          <span>{t('playground.changed')}</span>
        </div>
      </div>
    </div>
  );
}
