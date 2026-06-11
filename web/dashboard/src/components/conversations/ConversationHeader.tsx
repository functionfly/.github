import type { Conversation } from '@/api/conversations';
import { formatParticipantLine } from '@/components/conversations/constants';
import { Button } from '@/components/ui/button';
import { CheckCircle, Coins, Search } from 'lucide-react';

export interface ConversationHeaderProps {
  conversation: Conversation;
  currentUserId?: string;
  displayForParticipantId: (id: string) => string;
  onSearch: () => void;
  onResolve: () => void;
  onBounty: () => void;
  resolvePending: boolean;
}

export function ConversationHeader({
  conversation,
  currentUserId,
  displayForParticipantId,
  onSearch,
  onResolve,
  onBounty,
  resolvePending,
}: ConversationHeaderProps) {
  return (
    <div className="conv-aviation-header border-b border-border px-4 py-2 flex items-center justify-between gap-2">
      <div className="flex flex-col min-w-0 gap-0.5">
        <span className="conv-aviation-header-title text-sm font-medium truncate">
          {formatParticipantLine(
            conversation.participant_ids,
            currentUserId,
            displayForParticipantId
          )}
        </span>
        <span className="conv-aviation-header-subtitle text-xs text-muted-foreground capitalize">
          {conversation.type.replace(/_/g, ' ')}
        </span>
      </div>
      <div className="conv-aviation-header-actions flex gap-1">
        <Button
          size="sm"
          variant="ghost"
          className="conv-aviation-btn gap-1"
          onClick={onSearch}
          title="Search messages"
        >
          <Search className="h-3.5 w-3.5" />
        </Button>
        {!conversation.resolved_at && (
          <Button
            size="sm"
            variant="outline"
            className="conv-aviation-btn-primary gap-1"
            onClick={onResolve}
            disabled={resolvePending}
          >
            <CheckCircle className="h-3.5 w-3.5" />
            Resolve
          </Button>
        )}
        <Button
          size="sm"
          variant="ghost"
          className="conv-aviation-btn gap-1"
          onClick={onBounty}
        >
          <Coins className="h-3.5 w-3.5" />
          Bounty
        </Button>
      </div>
    </div>
  );
}
