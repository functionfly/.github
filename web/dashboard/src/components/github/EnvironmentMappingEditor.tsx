import { useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Plus, Trash2, GitBranch, ArrowRight } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { cn } from '@/lib/utils';

const ENVIRONMENT_OPTIONS = [
  { value: 'production', label: 'Production' },
  { value: 'staging', label: 'Staging' },
  { value: 'development', label: 'Development' },
  { value: 'preview', label: 'Preview' },
];

interface Mapping {
  id: string;
  branch: string;
  environment: string;
}

function generateId(): string {
  return Math.random().toString(36).substring(2, 9);
}

interface EnvironmentMappingEditorProps {
  value: Record<string, string>;
  onChange: (mappings: Record<string, string>) => void;
  className?: string;
}

export function EnvironmentMappingEditor({ value, onChange, className }: EnvironmentMappingEditorProps) {
  const mappings: Mapping[] = Object.entries(value).map(([branch, environment], i) => ({
    id: `mapping-${i}-${branch}`,
    branch,
    environment,
  }));

  const addRow = useCallback(() => {
    onChange({ ...value, '': 'development' });
  }, [value, onChange]);

  const removeRow = useCallback(
    (branch: string) => {
      const next = { ...value };
      delete next[branch];
      onChange(next);
    },
    [value, onChange]
  );

  const updateBranch = useCallback(
    (oldBranch: string, newBranch: string) => {
      const next: Record<string, string> = {};
      for (const [k, v] of Object.entries(value)) {
        if (k === oldBranch) {
          next[newBranch] = v;
        } else {
          next[k] = v;
        }
      }
      onChange(next);
    },
    [value, onChange]
  );

  const updateEnvironment = useCallback(
    (branch: string, environment: string) => {
      onChange({ ...value, [branch]: environment });
    },
    [value, onChange]
  );

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
      className={className}
    >
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm flex items-center gap-2">
              <GitBranch className="h-4 w-4 text-text-muted" />
              Branch Environment Mapping
            </CardTitle>
            <Button
              variant="ghost"
              size="sm"
              onClick={addRow}
              aria-label="Add branch mapping"
            >
              <Plus className="h-3.5 w-3.5 mr-1" />
              Add
            </Button>
          </div>
        </CardHeader>

        <CardContent>
          {mappings.length === 0 ? (
            <div className="text-center py-6 text-xs text-text-muted">
              <p>No environment mappings configured.</p>
              <p className="mt-1">Functions will be deployed to the default environment.</p>
            </div>
          ) : (
            <div className="space-y-2">
              <div className="grid grid-cols-[1fr_auto_1fr_auto] gap-2 items-center text-[10px] text-text-muted uppercase tracking-wider px-1">
                <span>Branch</span>
                <span />
                <span>Environment</span>
                <span />
              </div>

              <AnimatePresence initial={false}>
                {mappings.map((mapping) => (
                  <motion.div
                    key={mapping.id}
                    initial={{ opacity: 0, height: 0 }}
                    animate={{ opacity: 1, height: 'auto' }}
                    exit={{ opacity: 0, height: 0 }}
                    transition={{ duration: 0.2 }}
                    className="overflow-hidden"
                  >
                    <div className="grid grid-cols-[1fr_auto_1fr_auto] gap-2 items-center">
                      <Input
                        value={mapping.branch}
                        onChange={(e) => updateBranch(mapping.branch, e.target.value)}
                        placeholder="e.g. main"
                        className="h-9 text-sm font-mono"
                        aria-label="Branch name"
                      />

                      <ArrowRight className="h-4 w-4 text-text-muted shrink-0" />

                      <Select
                        value={mapping.environment}
                        onValueChange={(env) => updateEnvironment(mapping.branch, env)}
                      >
                        <SelectTrigger className="h-9" aria-label="Target environment">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {ENVIRONMENT_OPTIONS.map((env) => (
                            <SelectItem key={env.value} value={env.value}>
                              {env.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>

                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-9 w-9 shrink-0 text-text-muted hover:text-red-500"
                        onClick={() => removeRow(mapping.branch)}
                        aria-label={`Remove ${mapping.branch} mapping`}
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </Button>
                    </div>
                  </motion.div>
                ))}
              </AnimatePresence>
            </div>
          )}
        </CardContent>
      </Card>
    </motion.div>
  );
}
