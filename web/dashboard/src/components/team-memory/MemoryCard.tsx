import { useState } from 'react';
import {
  Brain,
  CheckCircle,
  Shield,
  AlertCircle,
  MoreHorizontal,
  Check,
  X,
} from 'lucide-react';
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Badge } from '@/components/ui/badge';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import {
  useValidateMemory,
  useDeleteMemory,
} from '@/hooks/use-team-memory';
import { type TeamMemory } from '@/services/team-memory.service';
import { cn } from '@/lib/utils';

interface MemoryCardProps {
  memory: TeamMemory;
  teamId: string;
}

const typeIcons = {
  decision: CheckCircle,
  preference: Brain,
  process: Brain,
  client_context: Brain,
};

const typeColors = {
  decision: 'bg-blue-500/10 text-blue-500 border-blue-500/20',
  preference: 'bg-purple-500/10 text-purple-500 border-purple-500/20',
  process: 'bg-orange-500/10 text-orange-500 border-orange-500/20',
  client_context: 'bg-green-500/10 text-green-500 border-green-500/20',
};

const typeLabels = {
  decision: 'Decision',
  preference: 'Preference',
  process: 'Process',
  client_context: 'Client',
};

export function MemoryCard({ memory, teamId }: MemoryCardProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const validateMutation = useValidateMemory(teamId);
  const deleteMutation = useDeleteMemory(teamId);

  const Icon = typeIcons[memory.memory_type] || Brain;

  const handleValidate = () => {
    validateMutation.mutate({ memoryId: memory.id, validated: !memory.is_validated });
  };

  const handleDelete = () => {
    if (confirm('Are you sure you want to delete this memory?')) {
      deleteMutation.mutate(memory.id);
    }
  };

  return (
    <Card
      className={cn(
        'relative transition-all',
        !memory.is_validated && 'border-dashed',
        memory.is_encrypted && 'border-amber-500/50'
      )}
    >
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-2">
            <div className={cn('p-2 rounded-lg', typeColors[memory.memory_type])}>
              <Icon className="h-4 w-4" />
            </div>
            <Badge variant="outline" className={typeColors[memory.memory_type]}>
              {typeLabels[memory.memory_type]}
            </Badge>
          </div>
          <div className="flex items-center gap-2">
            {memory.is_encrypted && (
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger>
                    <Shield className="h-4 w-4 text-amber-500" />
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>Encrypted - requires team passphrase to decrypt</p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            )}
            {!memory.is_validated && (
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger>
                    <AlertCircle className="h-4 w-4 text-yellow-500" />
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>Unvalidated - pending review</p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            )}
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm">
                  <MoreHorizontal className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={handleValidate}>
                  {memory.is_validated ? (
                    <>
                      <X className="h-4 w-4 mr-2" />
                      Unvalidate
                    </>
                  ) : (
                    <>
                      <Check className="h-4 w-4 mr-2" />
                      Validate
                    </>
                  )}
                </DropdownMenuItem>
                <DropdownMenuItem onClick={handleDelete} className="text-destructive">
                  <X className="h-4 w-4 mr-2" />
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
        <CardTitle className="text-lg mt-2">
          {memory.summary || 'Untitled Memory'}
        </CardTitle>
        {memory.category && (
          <CardDescription className="text-xs">
            Category: {memory.category}
          </CardDescription>
        )}
      </CardHeader>

      <CardContent className="pb-3">
        {memory.is_encrypted ? (
          <div className="flex items-center gap-2 text-sm text-amber-600 bg-amber-50 p-3 rounded">
            <Shield className="h-4 w-4" />
            <span>Encrypted content - decrypt to view</span>
          </div>
        ) : memory.content ? (
          <div className="space-y-2">
            <ContentPreview content={memory.content} type={memory.memory_type} />
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">No content</p>
        )}
      </CardContent>

      <CardFooter className="pt-0 flex items-center justify-between text-xs text-muted-foreground">
        <div className="flex items-center gap-4">
          <span>Confidence: {Math.round(memory.confidence_score * 100)}%</span>
          <span>Importance: {Math.round(memory.importance_score * 100)}%</span>
        </div>
        <div className="flex items-center gap-2">
          <span>Accessed {memory.access_count} times</span>
          <span>•</span>
          <span>{new Date(memory.updated_at).toLocaleDateString()}</span>
        </div>
      </CardFooter>
    </Card>
  );
}

function ContentPreview({
  content,
  type,
}: {
  content: Record<string, any>;
  type: string;
}) {
  switch (type) {
    case 'decision':
      return (
        <div className="space-y-1">
          {content.rationale && (
            <p className="text-sm line-clamp-3">{content.rationale}</p>
          )}
          {content.decision_maker && (
            <p className="text-xs text-muted-foreground">
              Decision by: {content.decision_maker}
            </p>
          )}
        </div>
      );
    case 'preference':
      return (
        <div className="space-y-1">
          {content.subject && content.value && (
            <p className="text-sm">
              <strong>{content.subject}:</strong> {content.value}
            </p>
          )}
          {content.context && (
            <p className="text-xs text-muted-foreground">Context: {content.context}</p>
          )}
        </div>
      );
    case 'process':
      return (
        <div className="space-y-1">
          {content.name && (
            <p className="text-sm font-medium">{content.name}</p>
          )}
          {content.steps && Array.isArray(content.steps) && (
            <p className="text-xs text-muted-foreground">
              {content.steps.length} steps
            </p>
          )}
        </div>
      );
    case 'client_context':
      return (
        <div className="space-y-1">
          {content.client_name && (
            <p className="text-sm font-medium">{content.client_name}</p>
          )}
          {content.notes && (
            <p className="text-sm text-muted-foreground line-clamp-2">
              {content.notes}
            </p>
          )}
        </div>
      );
    default:
      return (
        <p className="text-sm text-muted-foreground line-clamp-3">
          {JSON.stringify(content)}
        </p>
      );
  }
}
