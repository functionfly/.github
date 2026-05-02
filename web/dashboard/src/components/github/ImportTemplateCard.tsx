import { motion } from 'framer-motion';
import {
  FileText,
  Edit3,
  Trash2,
  Play,
  Code2,
  Globe,
  Lock,
  EyeOff,
  RefreshCw,
  Star,
} from 'lucide-react';
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import type { GitHubTemplate } from '@/types/github';

function getConfigSummary(config: Record<string, unknown>): { runtime?: string; visibility?: string; autoSync?: boolean } {
  return {
    runtime: typeof config.runtime === 'string' ? config.runtime : undefined,
    visibility: typeof config.visibility === 'string' ? config.visibility : undefined,
    autoSync: typeof config.auto_sync === 'boolean' ? config.auto_sync : undefined,
  };
}

function getVisibilityIcon(visibility?: string) {
  switch (visibility) {
    case 'public':
      return Globe;
    case 'private':
      return Lock;
    case 'unlisted':
      return EyeOff;
    default:
      return Lock;
  }
}

interface ImportTemplateCardProps {
  template: GitHubTemplate;
  onApply?: (template: GitHubTemplate) => void;
  onEdit?: (template: GitHubTemplate) => void;
  onDelete?: (template: GitHubTemplate) => void;
  className?: string;
}

export function ImportTemplateCard({
  template,
  onApply,
  onEdit,
  onDelete,
  className,
}: ImportTemplateCardProps) {
  const summary = getConfigSummary(template.config);
  const VisIcon = getVisibilityIcon(summary.visibility);

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
      className={className}
    >
      <Card className="group hover:-translate-y-0.5 hover:shadow-lg transition-all duration-200 h-full flex flex-col">
        <CardHeader className="pb-3">
          <div className="flex items-start justify-between gap-2">
            <div className="flex items-center gap-2 min-w-0">
              <div className="h-8 w-8 rounded-lg bg-brand-500/10 flex items-center justify-center shrink-0">
                <FileText className="h-4 w-4 text-brand-500" />
              </div>
              <div className="min-w-0">
                <CardTitle className="text-sm truncate flex items-center gap-1.5">
                  {template.name}
                  {template.is_default && (
                    <Star className="h-3 w-3 text-amber-500 fill-amber-500 shrink-0" />
                  )}
                </CardTitle>
              </div>
            </div>
          </div>
          {template.description && (
            <p className="text-xs text-text-secondary line-clamp-2 mt-1">{template.description}</p>
          )}
        </CardHeader>

        <CardContent className="flex-1 space-y-3">
          <div className="flex flex-wrap gap-1.5">
            {summary.runtime && (
              <Badge variant="secondary" className="text-[10px] gap-1">
                <Code2 className="h-2.5 w-2.5" />
                {summary.runtime}
              </Badge>
            )}
            {summary.visibility && (
              <Badge variant="secondary" className="text-[10px] gap-1">
                <VisIcon className="h-2.5 w-2.5" />
                {summary.visibility}
              </Badge>
            )}
            {summary.autoSync && (
              <Badge variant="secondary" className="text-[10px] gap-1">
                <RefreshCw className="h-2.5 w-2.5" />
                auto-sync
              </Badge>
            )}
          </div>

          <div className="flex items-center gap-2 text-xs text-text-muted">
            <span>Used {template.usage_count} time{template.usage_count !== 1 ? 's' : ''}</span>
          </div>
        </CardContent>

        <CardFooter className="pt-3 border-t border-border-subtle gap-2">
          <Button
            variant="default"
            size="sm"
            onClick={() => onApply?.(template)}
            className="flex-1"
            aria-label={`Apply template ${template.name}`}
          >
            <Play className="h-3.5 w-3.5 mr-1.5" />
            Apply
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => onEdit?.(template)}
            aria-label={`Edit template ${template.name}`}
          >
            <Edit3 className="h-3.5 w-3.5" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onDelete?.(template)}
            className="text-red-500 hover:text-red-600 hover:bg-red-500/10"
            aria-label={`Delete template ${template.name}`}
          >
            <Trash2 className="h-3.5 w-3.5" />
          </Button>
        </CardFooter>
      </Card>
    </motion.div>
  );
}
