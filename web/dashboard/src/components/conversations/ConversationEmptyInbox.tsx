import type { Conversation } from '@/api/conversations';
import { Button } from '@/components/ui/button';
import { MessageSquare, Plus, Users } from 'lucide-react';

interface ConversationEmptyInboxProps {
  conversations: Conversation[];
  onNewConversation: () => void;
  displayForParticipantId: (id: string) => string;
  currentUserId?: string;
  onSelectConversation: (id: string) => void;
}

export function ConversationEmptyInbox({
  conversations,
  onNewConversation,
  displayForParticipantId,
  currentUserId,
  onSelectConversation,
}: ConversationEmptyInboxProps) {
  const unreadTotal = conversations.reduce((n, c) => n + (c.unread_count ?? 0), 0);
  const recent = conversations.filter((c) => !c.resolved_at).slice(0, 5);

  return (
    <div className="conv-empty-inbox">
      <div className="rounded-full bg-muted p-4">
        <MessageSquare className="h-8 w-8 text-muted-foreground" />
      </div>
      <div>
        <h2 className="text-lg font-semibold">Your conversations</h2>
        <p className="text-sm text-muted-foreground mt-1 max-w-md">
          {unreadTotal > 0
            ? `${unreadTotal} unread message${unreadTotal === 1 ? '' : 's'} across your threads.`
            : 'Executable threads for functions, fixes, bounties, and DMs — all in one place.'}
        </p>
      </div>
      <div className="flex flex-wrap gap-2 justify-center">
        <Button onClick={onNewConversation}>
          <Plus className="h-4 w-4 mr-1" />
          Start conversation
        </Button>
      </div>
      {recent.length > 0 && (
        <div className="w-full max-w-md mt-4 text-left">
          <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground mb-2">
            Recent threads
          </p>
          <ul className="space-y-1">
            {recent.map((c) => {
              const others = (c.participant_ids ?? []).filter(
                (id) => id.toLowerCase() !== currentUserId?.toLowerCase()
              );
              const label = others.map(displayForParticipantId).join(', ') || 'Thread';
              return (
                <li key={c.id}>
                  <button
                    type="button"
                    className="w-full flex items-center gap-2 rounded-md px-3 py-2 text-sm hover:bg-muted text-left"
                    onClick={() => onSelectConversation(c.id)}
                  >
                    <Users className="h-4 w-4 shrink-0 text-muted-foreground" />
                    <span className="truncate">{label}</span>
                    {(c.unread_count ?? 0) > 0 && (
                      <span className="ml-auto text-xs font-semibold text-primary">
                        {c.unread_count}
                      </span>
                    )}
                  </button>
                </li>
              );
            })}
          </ul>
        </div>
      )}
    </div>
  );
}
