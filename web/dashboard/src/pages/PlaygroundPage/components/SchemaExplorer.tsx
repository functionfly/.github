import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import { AnimatePresence, motion } from 'framer-motion';
import { AlertCircle, Check, ChevronDown, ChevronRight, Copy } from 'lucide-react';
import { useState } from 'react';
import { SchemaNode, flattenSchema, getTypeBadgeColor } from '../utils/schemaHelpers';

interface SchemaExplorerProps {
  schema: Record<string, unknown>;
  onFieldClick?: (path: string) => void;
  className?: string;
}

interface SchemaNodeItemProps {
  node: SchemaNode;
  depth: number;
  onFieldClick?: (path: string) => void;
}

function SchemaNodeItem({ node, depth, onFieldClick }: SchemaNodeItemProps) {
  const [isExpanded, setIsExpanded] = useState(depth < 1);
  const [copied, setCopied] = useState(false);
  const hasChildren = (node.properties && node.properties.length > 0) || node.items;

  const handleCopyPath = async (e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await navigator.clipboard.writeText(node.path);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      // ignore
    }
  };

  const handleClick = () => {
    if (hasChildren) {
      setIsExpanded((prev) => !prev);
    }
    onFieldClick?.(node.path);
  };

  const typeLabel = Array.isArray(node.type) ? node.type.join(' | ') : node.type;
  const typeBadgeColor = getTypeBadgeColor(node.type);

  return (
    <div>
      <div
        className={cn(
          'group flex items-start gap-2 py-1.5 px-2 rounded hover:bg-bg-tertiary/50 cursor-pointer text-xs',
          'transition-colors'
        )}
        style={{ paddingLeft: `${depth * 16 + 8}px` }}
        onClick={handleClick}
      >
        {/* Expand icon */}
        <span className="w-3.5 h-3.5 flex items-center justify-center shrink-0 mt-0.5">
          {hasChildren ? (
            isExpanded ? (
              <ChevronDown className="w-3 h-3 text-text-muted" />
            ) : (
              <ChevronRight className="w-3 h-3 text-text-muted" />
            )
          ) : (
            <span className="w-3 h-3" />
          )}
        </span>

        {/* Field name */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-1.5 flex-wrap">
            <span className="font-mono text-text-primary font-medium">{node.key}</span>

            {node.required && (
              <span className="text-red-400 text-[10px]" title="Required">
                *
              </span>
            )}

            <Badge
              variant="outline"
              className={cn('text-[10px] px-1 py-0 h-4 border', typeBadgeColor)}
            >
              {typeLabel}
            </Badge>

            {node.enum && (
              <Badge
                variant="outline"
                className="text-[10px] px-1 py-0 h-4 border bg-indigo-500/10 text-indigo-400 border-indigo-500/20"
              >
                enum
              </Badge>
            )}
          </div>

          {node.description && (
            <p className="text-text-muted mt-0.5 text-[11px] leading-relaxed">{node.description}</p>
          )}

          {node.example !== undefined && (
            <p className="text-text-muted mt-0.5 text-[11px]">
              <span className="text-text-secondary">Example:</span>{' '}
              <span className="font-mono text-green-400">
                {typeof node.example === 'string'
                  ? `"${node.example}"`
                  : JSON.stringify(node.example)}
              </span>
            </p>
          )}

          {node.enum && (
            <div className="flex flex-wrap gap-1 mt-1">
              {node.enum.map((v, i) => (
                <span
                  key={i}
                  className="font-mono text-[10px] px-1 py-0.5 bg-bg-tertiary rounded border border-border-subtle text-text-secondary"
                >
                  {JSON.stringify(v)}
                </span>
              ))}
            </div>
          )}
        </div>

        {/* Copy path button */}
        <button
          onClick={handleCopyPath}
          className="opacity-0 group-hover:opacity-100 p-0.5 rounded hover:bg-bg-secondary transition-all shrink-0"
          title={`Copy path: ${node.path}`}
        >
          {copied ? (
            <Check className="w-3 h-3 text-green-400" />
          ) : (
            <Copy className="w-3 h-3 text-text-muted" />
          )}
        </button>
      </div>

      {/* Children */}
      {hasChildren && isExpanded && (
        <AnimatePresence>
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            transition={{ duration: 0.15 }}
          >
            {node.properties?.map((child) => (
              <SchemaNodeItem
                key={child.path}
                node={child}
                depth={depth + 1}
                onFieldClick={onFieldClick}
              />
            ))}
            {node.items && (
              <div
                className="flex items-center gap-2 py-1 px-2 text-xs text-text-muted"
                style={{ paddingLeft: `${(depth + 1) * 16 + 8}px` }}
              >
                <span className="font-mono">items:</span>
                <Badge
                  variant="outline"
                  className={cn(
                    'text-[10px] px-1 py-0 h-4 border',
                    getTypeBadgeColor(node.items.type)
                  )}
                >
                  {Array.isArray(node.items.type) ? node.items.type.join(' | ') : node.items.type}
                </Badge>
                {node.items.description && (
                  <span className="text-text-muted">{node.items.description}</span>
                )}
              </div>
            )}
          </motion.div>
        </AnimatePresence>
      )}
    </div>
  );
}

export function SchemaExplorer({ schema, onFieldClick, className }: SchemaExplorerProps) {
  const nodes = flattenSchema(schema);

  if (nodes.length === 0) {
    return (
      <div className={cn('flex flex-col items-center justify-center py-8 text-center', className)}>
        <AlertCircle className="w-8 h-8 text-text-muted mb-2" />
        <p className="text-sm text-text-muted">No schema available</p>
      </div>
    );
  }

  return (
    <div className={cn('space-y-0.5', className)}>
      {nodes.map((node) => (
        <SchemaNodeItem key={node.path} node={node} depth={0} onFieldClick={onFieldClick} />
      ))}
    </div>
  );
}
