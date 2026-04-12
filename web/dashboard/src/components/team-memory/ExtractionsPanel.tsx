import { CheckCircle, XCircle, Brain, AlertCircle } from 'lucide-react';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { useMemoryExtractions, useApproveExtraction, useRejectExtraction } from '@/hooks/use-team-memory';
import { type MemoryExtraction } from '@/services/team-memory.service';

interface ExtractionsPanelProps {
  teamId: string;
}

const typeLabels: Record<string, string> = {
  decision: 'Decision',
  preference: 'Preference',
  process: 'Process',
  client_context: 'Client Context',
};

const typeColors: Record<string, string> = {
  decision: 'bg-blue-500/10 text-blue-500',
  preference: 'bg-purple-500/10 text-purple-500',
  process: 'bg-orange-500/10 text-orange-500',
  client_context: 'bg-green-500/10 text-green-500',
};

export function ExtractionsPanel({ teamId }: ExtractionsPanelProps) {
  const { data: extractions, isLoading } = useMemoryExtractions(teamId, 'pending');
  const approveMutation = useApproveExtraction(teamId);
  const rejectMutation = useRejectExtraction(teamId);

  if (isLoading) {
    return <div className="text-center py-12">Loading pending extractions...</div>;
  }

  if (!extractions || extractions.length === 0) {
    return (
      <div className="text-center py-12 text-muted-foreground">
        <CheckCircle className="h-12 w-12 mx-auto mb-4 opacity-50" />
        <p className="text-lg font-medium">No pending extractions</p>
        <p className="text-sm">
          AI-extracted memories will appear here for your review. High-confidence
          extractions (≥90%) are auto-approved.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <AlertCircle className="h-4 w-4" />
        <span>
          Review AI-extracted memories before they&apos;re added to team knowledge.
          You can approve or reject each extraction.
        </span>
      </div>

      <div className="grid gap-4">
        {extractions.map((extraction: MemoryExtraction) => (
          <ExtractionCard
            key={extraction.id}
            extraction={extraction}
            onApprove={() => approveMutation.mutate(extraction.id)}
            onReject={() => rejectMutation.mutate({ extractionId: extraction.id })}
            isApproving={approveMutation.isPending}
            isRejecting={rejectMutation.isPending}
          />
        ))}
      </div>
    </div>
  );
}

interface ExtractionCardProps {
  extraction: MemoryExtraction;
  onApprove: () => void;
  onReject: () => void;
  isApproving: boolean;
  isRejecting: boolean;
}

function ExtractionCard({
  extraction,
  onApprove,
  onReject,
  isApproving,
  isRejecting,
}: ExtractionCardProps) {
  const confidence = Math.round(extraction.confidence * 100);
  const isHighConfidence = extraction.confidence >= 0.9;

  return (
    <Card className={isHighConfidence ? 'border-primary/50' : ''}>
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-2">
            <Brain className="h-5 w-5 text-primary" />
            <Badge className={typeColors[extraction.memory_type] || 'bg-gray-500/10'}>
              {typeLabels[extraction.memory_type] || extraction.memory_type}
            </Badge>
            {isHighConfidence && (
              <Badge variant="outline" className="text-primary border-primary">
                Auto-apply eligible
              </Badge>
            )}
          </div>
          <span className="text-sm text-muted-foreground">
            {confidence}% confidence
          </span>
        </div>
        <CardTitle className="text-lg">{extraction.summary}</CardTitle>
        {extraction.category && (
          <CardDescription>Category: {extraction.category}</CardDescription>
        )}
      </CardHeader>

      <CardContent className="pb-3 space-y-3">
        <ContentPreview content={extraction.content} type={extraction.memory_type} />

        {extraction.rationale && (
          <div className="bg-muted p-3 rounded text-sm">
            <p className="font-medium text-muted-foreground mb-1">
              AI Rationale:
            </p>
            <p>{extraction.rationale}</p>
          </div>
        )}
      </CardContent>

      <CardFooter className="flex justify-end gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={onReject}
          disabled={isRejecting || isApproving}
        >
          <XCircle className="h-4 w-4 mr-2" />
          {isRejecting ? 'Rejecting...' : 'Reject'}
        </Button>
        <Button
          size="sm"
          onClick={onApprove}
          disabled={isRejecting || isApproving}
        >
          <CheckCircle className="h-4 w-4 mr-2" />
          {isApproving ? 'Approving...' : 'Approve'}
        </Button>
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
        <div className="space-y-1 text-sm">
          {content.rationale && (
            <p>
              <span className="font-medium">Rationale:</span>{' '}
              {content.rationale}
            </p>
          )}
          {content.decision_maker && (
            <p>
              <span className="font-medium">Decision by:</span>{' '}
              {content.decision_maker}
            </p>
          )}
        </div>
      );
    case 'preference':
      return (
        <div className="space-y-1 text-sm">
          {content.subject && content.value && (
            <p>
              <span className="font-medium">{content.subject}:</span>{' '}
              {content.value}
            </p>
          )}
          {content.context && (
            <p className="text-muted-foreground">Context: {content.context}</p>
          )}
        </div>
      );
    case 'process':
      return (
        <div className="space-y-1 text-sm">
          {content.name && (
            <p className="font-medium">{content.name}</p>
          )}
          {content.steps && Array.isArray(content.steps) && (
            <p className="text-muted-foreground">
              {content.steps.length} steps
            </p>
          )}
        </div>
      );
    case 'client_context':
      return (
        <div className="space-y-1 text-sm">
          {content.client_name && (
            <p className="font-medium">{content.client_name}</p>
          )}
          {content.notes && (
            <p className="text-muted-foreground line-clamp-2">{content.notes}</p>
          )}
        </div>
      );
    default:
      return (
        <p className="text-sm text-muted-foreground">
          {JSON.stringify(content).slice(0, 200)}
          {JSON.stringify(content).length > 200 ? '...' : ''}
        </p>
      );
  }
}
