import type { ConversationMessage } from '@/api/conversations';
import { Button } from '@/components/ui/button';
import { SkeletonMessages } from '@/components/ui/skeleton-chat';
import { cn } from '@/lib/utils';
import { ArrowDown, Loader2 } from 'lucide-react';
import { useCallback, useEffect, useRef, useState } from 'react';
import { groupMessagesByDate } from './conversation-ui';
import { ExecutableMessage } from './ExecutableMessage';
import type { UserLookupEntry } from './ParticipantAvatar';

export interface ConversationMessageListProps {
  messages: ConversationMessage[];
  isLoading: boolean;
  isLoadingMore?: boolean;
  hasMore?: boolean;
  onLoadOlder?: () => void;
  isOwn: (authorId: string) => boolean;
  username?: string;
  currentUserId?: string;
  userById: Map<string, UserLookupEntry>;
  displayForParticipantId: (id: string) => string;
}

export function ConversationMessageList({
  messages,
  isLoading,
  isLoadingMore = false,
  hasMore = false,
  onLoadOlder,
  isOwn,
  username,
  currentUserId,
  userById,
  displayForParticipantId,
}: ConversationMessageListProps) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const bottomRef = useRef<HTMLDivElement>(null);
  const loadOlderRequestedRef = useRef(false);
  const [showJump, setShowJump] = useState(false);
  const prevLengthRef = useRef(messages.length);

  useEffect(() => {
    if (!isLoadingMore) {
      loadOlderRequestedRef.current = false;
    }
  }, [isLoadingMore]);

  const scrollToBottom = useCallback((behavior: ScrollBehavior = 'smooth') => {
    bottomRef.current?.scrollIntoView({ behavior });
  }, []);

  useEffect(() => {
    if (!isLoading && messages.length > 0 && messages.length > prevLengthRef.current) {
      const el = scrollRef.current;
      if (el) {
        const nearBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 120;
        if (nearBottom || prevLengthRef.current === 0) {
          scrollToBottom(prevLengthRef.current === 0 ? 'auto' : 'smooth');
        } else {
          setShowJump(true);
        }
      }
    }
    prevLengthRef.current = messages.length;
  }, [messages.length, isLoading, scrollToBottom]);

  const handleScroll = () => {
    const el = scrollRef.current;
    if (!el) return;
    const nearBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 80;
    setShowJump((prev) => (prev === !nearBottom ? prev : !nearBottom));

    if (
      hasMore &&
      onLoadOlder &&
      el.scrollTop < 80 &&
      !isLoadingMore &&
      !loadOlderRequestedRef.current
    ) {
      loadOlderRequestedRef.current = true;
      onLoadOlder();
    }
  };

  if (isLoading) {
    return (
      <div className="flex-1 p-4 overflow-y-auto">
        <SkeletonMessages count={6} />
      </div>
    );
  }

  const groups = groupMessagesByDate(messages);

  return (
    <div className="conv-message-list-wrap flex-1 min-h-0 flex flex-col">
      {hasMore && (
        <div className="flex justify-center py-2 border-b border-border/50">
          <Button
            variant="ghost"
            size="sm"
            className="h-7 text-xs"
            onClick={onLoadOlder}
            disabled={isLoadingMore}
          >
            {isLoadingMore ? (
              <>
                <Loader2 className="h-3 w-3 mr-1 animate-spin" />
                Loading…
              </>
            ) : (
              'Load older messages'
            )}
          </Button>
        </div>
      )}
      <div ref={scrollRef} className="flex-1 overflow-y-auto p-4" onScroll={handleScroll}>
        <div className="flex flex-col gap-1 min-h-full">
          {groups.map((group) => (
            <div key={group.label}>
              <div className="conv-date-separator">
                <span>{group.label}</span>
              </div>
              <div className="space-y-4">
                {group.messages.map((m) => {
                  const user = userById.get(m.author_id.toLowerCase());
                  const authorDisplayName = isOwn(m.author_id)
                    ? undefined
                    : displayForParticipantId(m.author_id);
                  return (
                    <ExecutableMessage
                      key={m.id}
                      message={m}
                      isOwn={isOwn(m.author_id)}
                      authorDisplayName={authorDisplayName}
                      authorAvatarUrl={user?.avatar}
                      authorInitials={
                        user?.username?.slice(0, 2) ||
                        user?.name?.slice(0, 2) ||
                        m.author_id.slice(0, 2)
                      }
                      username={username}
                      currentUserId={currentUserId}
                    />
                  );
                })}
              </div>
            </div>
          ))}
          <div ref={bottomRef} className="h-px shrink-0" aria-hidden />
        </div>
      </div>
      {showJump && (
        <Button
          size="sm"
          className={cn('conv-jump-latest')}
          onClick={() => {
            scrollToBottom();
            setShowJump(false);
          }}
        >
          <ArrowDown className="h-3.5 w-3.5 mr-1" />
          Jump to latest
        </Button>
      )}
    </div>
  );
}
