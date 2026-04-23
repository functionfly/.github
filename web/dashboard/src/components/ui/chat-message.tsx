import * as React from 'react';
import { cn } from '@/lib/utils';
import { formatDistanceToNow } from 'date-fns';
import { SmilePlus } from 'lucide-react';
import { ReactionBar } from './reaction-bar';

export interface ChatMessageReaction {
  emoji: string;
  count: number;
  users: string[];
}

export interface ChatMessageAttachment {
  id: string;
  filename: string;
  content_type: string;
  size_bytes: number;
  storage_url: string;
  thumbnail_url?: string | null;
}

export interface ChatMessageProps {
  content: string;
  author: string;
  authorDisplayName?: string;
  timestamp: string;
  isOwn?: boolean;
  isDeleted?: boolean;
  isEdited?: boolean;
  attachments?: ChatMessageAttachment[];
  reactions?: ChatMessageReaction[];
  onReply?: () => void;
  onReact?: (emoji: string) => void;
  onRemoveReact?: (emoji: string) => void;
  actions?: React.ReactNode;
  className?: string;
  children?: React.ReactNode;
}

export function ChatMessage({
  content,
  authorDisplayName,
  timestamp,
  isOwn = false,
  isDeleted = false,
  isEdited = false,
  attachments,
  reactions,
  onReact,
  onRemoveReact,
  actions,
  className,
  children,
}: ChatMessageProps) {
  const [showReactPicker, setShowReactPicker] = React.useState(false);

  if (isDeleted) {
    return (
      <div
        className={cn(
          'flex flex-col gap-1 max-w-[85%]',
          isOwn && 'self-end items-end',
          !isOwn && 'self-start items-start',
          className,
        )}
      >
        <div className="rounded-lg px-3 py-2 text-sm italic text-muted-foreground bg-muted/30 border border-border/40">
          Message deleted
        </div>
      </div>
    );
  }

  return (
    <div
      className={cn(
        'flex flex-col gap-2 max-w-[85%] group',
        isOwn && 'self-end items-end',
        !isOwn && 'self-start items-start',
        className,
      )}
    >
      {authorDisplayName && !isOwn && (
        <span className="text-xs font-medium text-muted-foreground">
          {authorDisplayName}
        </span>
      )}

      <div className="relative">
        <div
          className={cn(
            'rounded-lg px-3 py-2 text-sm break-words',
            isOwn
              ? 'bg-brand-500/15 text-brand-foreground border border-brand-500/30'
              : 'bg-muted/60 text-foreground border border-border/60',
          )}
        >
          {content && <p className="whitespace-pre-wrap">{content}</p>}
          {isEdited && (
            <span className="block text-[10px] text-muted-foreground mt-1">
              (edited)
            </span>
          )}
        </div>

        {actions && (
          <div
            className={cn(
              'absolute top-0.5',
              isOwn ? '-left-7' : '-right-7',
            )}
          >
            {actions}
          </div>
        )}
      </div>

      {attachments && attachments.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {attachments.map((att) => (
            <a
              key={att.id}
              href={att.storage_url}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-2 rounded-md border border-border/60 bg-muted/40 px-2 py-1.5 text-xs hover:bg-muted/60 transition-colors"
            >
              {att.content_type.startsWith('image/') && att.thumbnail_url ? (
                <img
                  src={att.thumbnail_url}
                  alt={att.filename}
                  className="h-8 w-8 rounded object-cover"
                />
              ) : null}
              <span className="truncate max-w-[150px]">{att.filename}</span>
              <span className="text-muted-foreground">
                {(att.size_bytes / 1024).toFixed(0)}KB
              </span>
            </a>
          ))}
        </div>
      )}

      {children}

      {reactions && reactions.length > 0 && (
        <ReactionBar
          reactions={reactions}
          onReact={onReact}
          onRemoveReact={onRemoveReact}
        />
      )}

      <div className="flex items-center gap-2">
        <span className="text-[10px] text-muted-foreground px-1">
          {formatDistanceToNow(new Date(timestamp), { addSuffix: true })}
        </span>
        {onReact && (
          <button
            onClick={() => setShowReactPicker(!showReactPicker)}
            className="opacity-0 group-hover:opacity-100 transition-opacity text-muted-foreground hover:text-foreground"
            title="Add reaction"
          >
            <SmilePlus className="h-3.5 w-3.5" />
          </button>
        )}
      </div>

      {showReactPicker && (
        <div className="flex gap-1 rounded-lg border border-border bg-card px-2 py-1 shadow-md text-sm">
          {['👍', '❤️', '😂', '🎉', '👀', '🚀'].map((emoji) => (
            <button
              key={emoji}
              onClick={() => {
                onReact?.(emoji);
                setShowReactPicker(false);
              }}
              className="hover:scale-125 transition-transform p-0.5"
            >
              {emoji}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
