import { useState, useCallback } from 'react';
import { motion } from 'framer-motion';
import { Code2, FileCode, Package } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Checkbox } from '@/components/ui/checkbox';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { VisibilitySelector } from '@/components/github/VisibilitySelector';
import type { DetectedFunction } from '@/types/github';

function getConfidenceColor(confidence: number): string {
  if (confidence >= 80) return 'bg-emerald-500';
  if (confidence >= 50) return 'bg-amber-500';
  return 'bg-red-500';
}

function getConfidenceTextColor(confidence: number): string {
  if (confidence >= 80) return 'text-emerald-500';
  if (confidence >= 50) return 'text-amber-500';
  return 'text-red-500';
}

function getRuntimeBadge(runtime: string): { label: string; color: string } {
  const lower = runtime.toLowerCase();
  if (lower.includes('node') || lower.includes('javascript') || lower.includes('typescript'))
    return { label: 'Node.js', color: 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20' };
  if (lower.includes('python'))
    return { label: 'Python', color: 'bg-blue-500/10 text-blue-500 border-blue-500/20' };
  if (lower.includes('go'))
    return { label: 'Go', color: 'bg-cyan-500/10 text-cyan-500 border-cyan-500/20' };
  if (lower.includes('rust'))
    return { label: 'Rust', color: 'bg-orange-500/10 text-orange-500 border-orange-500/20' };
  if (lower.includes('ruby'))
    return { label: 'Ruby', color: 'bg-red-500/10 text-red-500 border-red-500/20' };
  return { label: runtime, color: 'bg-gray-500/10 text-gray-500 border-gray-500/20' };
}

interface DetectedFunctionCardProps {
  fn: DetectedFunction;
  selected: boolean;
  visibility: 'public' | 'private' | 'unlisted';
  editedName?: string;
  onSelect: (selected: boolean) => void;
  onNameChange: (name: string) => void;
  onVisibilityChange: (visibility: 'public' | 'private' | 'unlisted') => void;
  className?: string;
}

export function DetectedFunctionCard({
  fn,
  selected,
  visibility,
  editedName,
  onSelect,
  onNameChange,
  onVisibilityChange,
  className,
}: DetectedFunctionCardProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [localName, setLocalName] = useState(editedName ?? fn.name);
  const runtimeBadge = getRuntimeBadge(fn.runtime);

  const handleNameSave = useCallback(() => {
    setIsEditing(false);
    if (localName.trim() && localName !== fn.name) {
      onNameChange(localName.trim());
    } else {
      setLocalName(fn.name);
    }
  }, [localName, fn.name, onNameChange]);

  const deps = (fn.dependencies as string[] | null) ?? null;

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.2 }}
      className={cn(
        'rounded-xl border bg-card p-4 transition-all duration-200',
        selected ? 'border-brand-500/50 shadow-sm shadow-brand-500/10' : 'border-border-subtle',
        className
      )}
    >
      <div className="flex items-start gap-3">
        <Checkbox
          checked={selected}
          onCheckedChange={(checked) => onSelect(!!checked)}
          aria-label={`Select ${fn.name}`}
          className="mt-0.5"
        />

        <div className="flex-1 min-w-0 space-y-2">
          <div className="flex items-center gap-2 flex-wrap">
            {isEditing ? (
              <Input
                value={localName}
                onChange={(e) => setLocalName(e.target.value)}
                onBlur={handleNameSave}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') handleNameSave();
                  if (e.key === 'Escape') {
                    setLocalName(fn.name);
                    setIsEditing(false);
                  }
                }}
                className="h-7 text-sm font-semibold max-w-[200px]"
                autoFocus
                aria-label="Edit function name"
              />
            ) : (
              <button
                onClick={() => setIsEditing(true)}
                className="text-sm font-semibold text-text-primary hover:text-brand-500 transition-colors cursor-text"
                title="Click to edit name"
              >
                {editedName ?? fn.name}
              </button>
            )}

            <Badge variant="outline" className={cn('text-[10px] border', runtimeBadge.color)}>
              {runtimeBadge.label}
            </Badge>
          </div>

          <div className="flex items-center gap-2 text-xs text-text-muted">
            <FileCode className="h-3 w-3 shrink-0" />
            <span className="truncate font-mono">{fn.entry_point}</span>
            {fn.sub_directory && (
              <>
                <span className="text-border-subtle">·</span>
                <span className="truncate">{fn.sub_directory}</span>
              </>
            )}
          </div>

          <div className="flex items-center gap-3">
            <div className="flex-1 max-w-[160px]">
              <div className="flex items-center justify-between mb-1">
                <span className="text-[10px] text-text-muted">Confidence</span>
                <span className={cn('text-[10px] font-mono font-bold', getConfidenceTextColor(fn.confidence))}>
                  {fn.confidence}%
                </span>
              </div>
              <Progress
                value={fn.confidence}
                className="h-1.5"
                indicatorClassName={getConfidenceColor(fn.confidence)}
              />
            </div>

            <VisibilitySelector
              value={visibility}
              onChange={onVisibilityChange}
              size="sm"
            />
          </div>

          {deps && deps.length > 0 && (
            <div className="flex items-center gap-1.5 flex-wrap">
              <Package className="h-3 w-3 text-text-muted shrink-0" />
              {deps.slice(0, 3).map((dep) => (
                <Badge key={dep} variant="secondary" className="text-[10px] font-mono">
                  {dep}
                </Badge>
              ))}
              {deps.length > 3 && (
                <Badge variant="secondary" className="text-[10px]">
                  +{deps.length - 3} more
                </Badge>
              )}
            </div>
          )}
        </div>
      </div>
    </motion.div>
  );
}
