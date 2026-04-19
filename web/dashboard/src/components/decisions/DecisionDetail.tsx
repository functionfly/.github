import type { TeamDecision } from '@/api/decisions';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Separator } from '@/components/ui/separator';
import { format, formatDistanceToNow } from 'date-fns';
import { ArrowRight, Calendar, CheckCircle, Pencil, Tag, Trash2 } from 'lucide-react';
import { DecisionStatusBadge } from './DecisionStatusBadge';

interface DecisionDetailProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  decision: TeamDecision | null;
  onEdit?: (decision: TeamDecision) => void;
  onDelete?: (decisionId: string) => void;
  onApprove?: (decision: TeamDecision) => void;
}

export function DecisionDetail({
  open,
  onOpenChange,
  decision,
  onEdit,
  onDelete,
  onApprove,
}: DecisionDetailProps) {
  if (!decision) return null;

  const canApprove = decision.status === 'pending' && onApprove;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[700px] max-h-[90vh] overflow-y-auto">
        <DialogHeader className="space-y-4">
          <div className="flex items-start justify-between gap-4">
            <div className="flex items-center gap-3">
              <DecisionStatusBadge status={decision.status} />
              {decision.source_type === 'ai_extracted' && (
                <Badge
                  variant="outline"
                  className="bg-purple-500/10 text-purple-500 border-purple-500/20"
                >
                  AI Extracted
                </Badge>
              )}
            </div>

            <div className="flex gap-2">
              {onEdit && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    onOpenChange(false);
                    onEdit(decision);
                  }}
                >
                  <Pencil className="mr-2 h-4 w-4" />
                  Edit
                </Button>
              )}
              {canApprove && (
                <Button
                  variant="default"
                  size="sm"
                  onClick={() => {
                    onOpenChange(false);
                    onApprove(decision);
                  }}
                >
                  <CheckCircle className="mr-2 h-4 w-4" />
                  Approve
                </Button>
              )}
              {onDelete && (
                <Button
                  variant="destructive"
                  size="sm"
                  onClick={() => {
                    if (confirm('Are you sure you want to delete this decision?')) {
                      onOpenChange(false);
                      onDelete(decision.id);
                    }
                  }}
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  Delete
                </Button>
              )}
            </div>
          </div>

          <DialogTitle className="text-2xl font-bold">{decision.title}</DialogTitle>

          <div className="flex flex-wrap items-center gap-4 text-sm text-muted-foreground">
            <div className="flex items-center gap-1">
              <Calendar className="h-4 w-4" />
              Recorded{' '}
              {formatDistanceToNow(new Date(decision.created_at), {
                addSuffix: true,
              })}
            </div>
            {decision.approved_by && decision.approved_at && (
              <div className="flex items-center gap-1 text-green-500">
                <CheckCircle className="h-4 w-4" />
                Approved{' '}
                {formatDistanceToNow(new Date(decision.approved_at), {
                  addSuffix: true,
                })}
              </div>
            )}
          </div>
        </DialogHeader>

        <div className="space-y-6 pt-4">
          {/* Description */}
          {decision.description && (
            <div>
              <h3 className="font-semibold mb-2">Summary</h3>
              <p className="text-muted-foreground">{decision.description}</p>
            </div>
          )}

          {/* Rationale */}
          {decision.rationale && (
            <>
              <Separator />
              <div>
                <h3 className="font-semibold mb-2 flex items-center gap-2">Why This Decision?</h3>
                <p className="text-muted-foreground">{decision.rationale}</p>
              </div>
            </>
          )}

          {/* Outcome */}
          {decision.outcome && (
            <>
              <Separator />
              <div>
                <h3 className="font-semibold mb-2 flex items-center gap-2">What Was Decided</h3>
                <p className="text-muted-foreground">{decision.outcome}</p>
              </div>
            </>
          )}

          {/* Alternatives */}
          {decision.alternatives && decision.alternatives.length > 0 && (
            <>
              <Separator />
              <div>
                <h3 className="font-semibold mb-2">Alternatives Considered</h3>
                <ul className="space-y-2">
                  {decision.alternatives.map((alt, i) => (
                    <li key={i} className="flex items-start gap-2 text-muted-foreground">
                      <ArrowRight className="h-4 w-4 mt-0.5 shrink-0" />
                      {alt}
                    </li>
                  ))}
                </ul>
              </div>
            </>
          )}

          {/* Tags */}
          {decision.tags && decision.tags.length > 0 && (
            <>
              <Separator />
              <div>
                <h3 className="font-semibold mb-2 flex items-center gap-2">
                  <Tag className="h-4 w-4" />
                  Tags
                </h3>
                <div className="flex flex-wrap gap-2">
                  {decision.tags.map((tag) => (
                    <Badge key={tag} variant="secondary">
                      {tag}
                    </Badge>
                  ))}
                </div>
              </div>
            </>
          )}

          {/* Metadata */}
          <Separator />
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <p className="text-muted-foreground">Created</p>
              <p>{format(new Date(decision.created_at), 'PPpp')}</p>
            </div>
            <div>
              <p className="text-muted-foreground">Last Updated</p>
              <p>{format(new Date(decision.updated_at), 'PPpp')}</p>
            </div>
            {decision.approved_at && (
              <div>
                <p className="text-muted-foreground">Approved At</p>
                <p>{format(new Date(decision.approved_at), 'PPpp')}</p>
              </div>
            )}
            <div>
              <p className="text-muted-foreground">Importance</p>
              <p>
                {decision.importance_score < 0.3
                  ? 'Low'
                  : decision.importance_score < 0.7
                    ? 'Medium'
                    : 'High'}{' '}
                ({Math.round(decision.importance_score * 100)}%)
              </p>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
