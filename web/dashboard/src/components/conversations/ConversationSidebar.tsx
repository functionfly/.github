import type { Conversation } from '@/api/conversations';
import { formatParticipantLine } from '@/components/conversations/constants';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { SkeletonConversationList } from '@/components/ui/skeleton-chat';
import { cn } from '@/lib/utils';
import { formatDistanceToNow } from 'date-fns';
import { MessageSquare, Plus, CheckCircle } from 'lucide-react';
import { Link } from 'react-router-dom';
import { useState, useCallback, useRef, useEffect } from 'react';
import { conversationsApi } from '@/api/conversations';

export interface ConversationSidebarProps {
  conversations: Conversation[];
  loading: boolean;
  currentUsername: string;
  activeConversationId?: string;
  currentUserId?: string;
  displayForParticipantId: (id: string) => string;
  onNewConversation: () => void;
}

export function ConversationSidebar({
  conversations,
  loading,
  currentUsername,
  activeConversationId,
  currentUserId,
  displayForParticipantId,
  onNewConversation,
}: ConversationSidebarProps) {
  const [hoveredId, setHoveredId] = useState<string | null>(null);
  const [previewText, setPreviewText] = useState<Record<string, string>>({});
  const previewCache = useRef<Set<string>>(new Set());

  const fetchPreview = useCallback(
    async (conversationId: string) => {
      if (previewCache.current.has(conversationId)) return;
      previewCache.current.add(conversationId);
      try {
        const res = await conversationsApi.listMessages(currentUsername, conversationId, {
          limit: 1,
          offset: 0,
        });
        if (res.messages.length > 0) {
          const msg = res.messages[0];
          const text = msg.deleted_at
            ? 'Message deleted'
            : msg.content.length > 60
              ? msg.content.slice(0, 60) + '\u2026'
              : msg.content;
          setPreviewText((prev) => ({ ...prev, [conversationId]: text }));
        }
      } catch {
        // Silently ignore preview fetch failures
      }
    },
    [currentUsername],
  );

  useEffect(() => {
    if (hoveredId && !previewText[hoveredId]) {
      fetchPreview(hoveredId);
    }
  }, [hoveredId, fetchPreview, previewText]);

  return (
    <aside className="w-72 border-r border-border bg-muted/20 flex flex-col">
      <div className="p-3 border-b border-border flex items-center justify-between">
        <h2 className="font-semibold text-sm">Messages</h2>
        <Button
          variant="ghost"
          size="icon"
          className="h-8 w-8"
          onClick={onNewConversation}
          title="New conversation"
        >
          <Plus className="h-4 w-4" />
        </Button>
      </div>
      <ScrollArea className="flex-1">
        {loading ? (
          <SkeletonConversationList count={6} />
        ) : conversations.length === 0 ? (
          <div className="p-4 text-center text-sm text-muted-foreground">
            <MessageSquare className="h-8 w-8 mx-auto mb-2 opacity-50" />
            <p>No conversations yet.</p>
            <p className="text-xs mt-1">Start one from a function or profile.</p>
          </div>
        ) : (
          <ul className="p-1">
            {conversations.map((c) => (
              <li key={c.id}>
                <Link
                  to={`/u/${currentUsername}/conversations/${c.id}`}
                  onMouseEnter={() => setHoveredId(c.id)}
                  onMouseLeave={() => setHoveredId(null)}
                  className={cn(
                    'flex flex-col gap-0.5 rounded-lg px-3 py-2 text-left transition-colors',
                    activeConversationId === c.id
                      ? 'bg-brand-500/15 border border-brand-500/30'
                      : 'hover:bg-muted/60',
                  )}
                >
                  <div className="flex items-start justify-between gap-2">
                    <div className="min-w-0 flex-1 flex flex-col gap-0.5">
                      <span className="text-xs text-muted-foreground capitalize flex items-center gap-1">
                        {c.resolved_at && (
                          <CheckCircle className="h-3 w-3 text-green-500 shrink-0" />
                        )}
                        {c.type.replace(/_/g, ' ')}
                      </span>
                      <span className="text-sm font-medium truncate">
                        {formatParticipantLine(
                          c.participant_ids,
                          currentUserId,
                          displayForParticipantId,
                        )}
                      </span>
                      <span className="text-xs text-muted-foreground">
                        {formatDistanceToNow(new Date(c.updated_at), { addSuffix: true })}
                      </span>
                      {/* Hover preview of last message */}
                      {hoveredId === c.id && previewText[c.id] && (
                        <span className="text-xs text-muted-foreground/80 italic truncate mt-0.5">
                          {previewText[c.id]}
                        </span>
                      )}
                    </div>
                    {(c.unread_count ?? 0) > 0 && (
                      <span
                        className="mt-0.5 shrink-0 flex h-5 min-w-5 items-center justify-center rounded-full bg-brand-500 px-1.5 text-[10px] font-semibold text-white tabular-nums"
                        aria-label={`${c.unread_count} unread`}
                      >
                        {(c.unread_count ?? 0) > 99 ? '99+' : c.unread_count}
                      </span>
                    )}
                  </div>
                </Link>
              </li>
            ))}
          </ul>
        )}
      </ScrollArea>
    </aside>
  );
}
