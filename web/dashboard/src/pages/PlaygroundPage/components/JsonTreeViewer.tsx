import { useState, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { ChevronRight, ChevronDown, Copy, Check } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { cn } from '@/lib/utils';
import { getValueType, getTypeColor } from '../utils/schemaHelpers';

interface JsonTreeViewerProps {
  data: unknown;
  searchQuery?: string;
  className?: string;
  maxDepth?: number;
}

interface JsonNodeProps {
  keyName: string | null;
  value: unknown;
  depth: number;
  path: string;
  searchQuery?: string;
  maxDepth: number;
  isLast?: boolean;
}

function CopyButton({ value }: { value: unknown }) {
  const { t } = useTranslation();
  const [copied, setCopied] = useState(false);

  const handleCopy = async (e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      const text = typeof value === 'string' ? value : JSON.stringify(value, null, 2);
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      // ignore
    }
  };

  return (
    <button
      onClick={handleCopy}
      className="opacity-0 group-hover:opacity-100 ml-1.5 p-0.5 rounded hover:bg-bg-tertiary transition-all"
      title={t('playground.copyValue')}
    >
      {copied ? (
        <Check className="w-3 h-3 text-green-400" />
      ) : (
        <Copy className="w-3 h-3 text-text-muted" />
      )}
    </button>
  );
}

function JsonNode({
  keyName,
  value,
  depth,
  path,
  searchQuery,
  maxDepth,
  isLast = true,
}: JsonNodeProps) {
  const { t } = useTranslation();
  const [isExpanded, setIsExpanded] = useState(depth < 2);
  const valueType = getValueType(value);
  const isExpandable = valueType === 'object' || valueType === 'array';
  const isObject = valueType === 'object' && value !== null;
  const isArray = valueType === 'array';

  const toggle = useCallback(() => {
    if (isExpandable) setIsExpanded((prev) => !prev);
  }, [isExpandable]);

  const typeColorClass = getTypeColor(valueType);

  const renderPrimitive = () => {
    if (value === null) return <span className="text-text-muted">null</span>;
    if (value === undefined) return <span className="text-text-muted">undefined</span>;
    if (typeof value === 'string')
      return <span className="text-green-500 dark:text-green-400">"{value}"</span>;
    if (typeof value === 'number')
      return <span className="text-blue-500 dark:text-blue-400">{value}</span>;
    if (typeof value === 'boolean')
      return <span className="text-yellow-500 dark:text-yellow-400">{value.toString()}</span>;
    return <span className={typeColorClass}>{String(value)}</span>;
  };

  const entries = isObject
    ? Object.entries(value as Record<string, unknown>)
    : isArray
    ? (value as unknown[]).map((v, i) => [String(i), v] as [string, unknown])
    : [];

  const collapsedPreview = isObject
    ? `{${entries.length} ${entries.length === 1 ? 'key' : 'keys'}}`
    : isArray
    ? `[${(value as unknown[]).length} items]`
    : '';

  const indent = depth * 16;

  return (
    <div className="group">
      <div
        className={cn(
          'flex items-start gap-1 py-0.5 px-1 rounded hover:bg-bg-tertiary/50 cursor-default text-xs font-mono',
          isExpandable && 'cursor-pointer'
        )}
        style={{ paddingLeft: `${indent + 4}px` }}
        onClick={toggle}
      >
        {/* Expand/collapse icon */}
        <span className="w-3.5 h-3.5 flex items-center justify-center shrink-0 mt-0.5">
          {isExpandable ? (
            isExpanded ? (
              <ChevronDown className="w-3 h-3 text-text-muted" />
            ) : (
              <ChevronRight className="w-3 h-3 text-text-muted" />
            )
          ) : null}
        </span>

        {/* Key */}
        {keyName !== null && (
          <>
            <span className="text-purple-500 dark:text-purple-300 shrink-0">{keyName}</span>
            <span className="text-text-muted shrink-0">:</span>
          </>
        )}

        {/* Value */}
        {isExpandable ? (
          isExpanded ? (
            <span className="text-text-muted">{isObject ? '{' : '['}</span>
          ) : (
            <span className="text-text-muted">{collapsedPreview}</span>
          )
        ) : (
          <span className="flex items-center">
            {renderPrimitive()}
            {!isLast && <span className="text-text-muted">,</span>}
            <CopyButton value={value} />
          </span>
        )}
      </div>

      {/* Children */}
      {isExpandable && isExpanded && depth < maxDepth && (
        <AnimatePresence>
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            transition={{ duration: 0.15 }}
          >
            {entries.map(([k, v], i) => (
              <JsonNode
                key={k}
                keyName={isArray ? null : k}
                value={v}
                depth={depth + 1}
                path={`${path}.${k}`}
                searchQuery={searchQuery}
                maxDepth={maxDepth}
                isLast={i === entries.length - 1}
              />
            ))}
            <div
              className="text-xs font-mono text-text-muted py-0.5 px-1"
              style={{ paddingLeft: `${indent + 4}px` }}
            >
              <span className="w-3.5 inline-block" />
              {isObject ? '}' : ']'}
              {!isLast && ','}
            </div>
          </motion.div>
        </AnimatePresence>
      )}

      {isExpandable && isExpanded && depth >= maxDepth && (
        <div
          className="text-xs font-mono text-text-muted py-0.5 px-1"
          style={{ paddingLeft: `${(depth + 1) * 16 + 4}px` }}
        >
          ... ({t('playground.maxDepthReached')})
        </div>
      )}
    </div>
  );
}

export function JsonTreeViewer({
  data,
  searchQuery,
  className,
  maxDepth = 10,
}: JsonTreeViewerProps) {
  if (data === null || data === undefined) {
    return (
      <div className={cn('text-xs font-mono text-text-muted p-2', className)}>
        null
      </div>
    );
  }

  return (
    <div
      className={cn(
        'text-xs font-mono overflow-auto bg-bg-secondary rounded-md p-2',
        className
      )}
    >
      <JsonNode
        keyName={null}
        value={data}
        depth={0}
        path="root"
        searchQuery={searchQuery}
        maxDepth={maxDepth}
        isLast={true}
      />
    </div>
  );
}
