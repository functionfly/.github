import { useState } from 'react';
import { motion } from 'framer-motion';
import { AlertTriangle, ArrowRight, SkipForward, GitBranch, Tag } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Label } from '@/components/ui/label';
import { cn } from '@/lib/utils';
import type { ImportConflict } from '@/types/github';

type Resolution = 'overwrite' | 'rename' | 'new_version' | 'skip';

interface ResolutionOption {
  value: Resolution;
  label: string;
  description: string;
  icon: React.ComponentType<{ className?: string }>;
  variant?: 'default' | 'destructive';
}

const RESOLUTION_OPTIONS: ResolutionOption[] = [
  {
    value: 'overwrite',
    label: 'Overwrite',
    description: 'Replace the existing function with this import',
    icon: ArrowRight,
    variant: 'destructive',
  },
  {
    value: 'rename',
    label: 'Rename',
    description: 'Import as a new function with a different name',
    icon: GitBranch,
  },
  {
    value: 'new_version',
    label: 'New Version',
    description: 'Create a new version of the existing function',
    icon: Tag,
  },
  {
    value: 'skip',
    label: 'Skip',
    description: 'Skip importing this function',
    icon: SkipForward,
  },
];

interface ConflictResolutionDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  conflict: ImportConflict;
  onResolve: (conflict: ImportConflict, resolution: Resolution) => void;
  className?: string;
}

export function ConflictResolutionDialog({
  open,
  onOpenChange,
  conflict,
  onResolve,
  className,
}: ConflictResolutionDialogProps) {
  const [resolution, setResolution] = useState<Resolution>('new_version');

  const handleApply = () => {
    onResolve(conflict, resolution);
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className={cn('sm:max-w-md', className)}>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <AlertTriangle className="h-5 w-5 text-amber-500" />
            Resolve Conflict
          </DialogTitle>
          <DialogDescription>
            A function named <span className="font-mono font-semibold text-text-primary">{conflict.function_name}</span> already
            exists. Choose how to resolve this conflict.
          </DialogDescription>
        </DialogHeader>

        <motion.div
          initial={{ opacity: 0, y: 5 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.2 }}
          className="space-y-4"
        >
          <div className="rounded-lg border border-border-subtle bg-bg-secondary/50 p-3 space-y-1">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium text-text-primary">{conflict.function_name}</span>
              <Badge variant="outline" className="text-xs font-mono">
                v{conflict.existing_version}
              </Badge>
            </div>
            <p className="text-xs text-text-muted">
              Function ID: <span className="font-mono">{conflict.existing_function_id}</span>
            </p>
          </div>

          <RadioGroup
            value={resolution}
            onValueChange={(val) => setResolution(val as Resolution)}
            className="space-y-2"
          >
            {RESOLUTION_OPTIONS.map((option) => {
              const Icon = option.icon;
              const isSelected = resolution === option.value;
              return (
                <label
                  key={option.value}
                  className={cn(
                    'flex items-start gap-3 rounded-lg border p-3 cursor-pointer transition-all duration-200',
                    isSelected
                      ? 'border-brand-500/50 bg-brand-500/5'
                      : 'border-border-subtle hover:border-border-default'
                  )}
                >
                  <RadioGroupItem value={option.value} className="mt-0.5" />
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <Icon
                        className={cn(
                          'h-4 w-4',
                          option.variant === 'destructive' ? 'text-red-500' : 'text-text-secondary'
                        )}
                      />
                      <span className="text-sm font-medium text-text-primary">{option.label}</span>
                    </div>
                    <p className="text-xs text-text-muted mt-0.5">{option.description}</p>
                  </div>
                </label>
              );
            })}
          </RadioGroup>
        </motion.div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} aria-label="Cancel">
            Cancel
          </Button>
          <Button
            onClick={handleApply}
            aria-label="Apply resolution"
          >
            Apply Resolution
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
