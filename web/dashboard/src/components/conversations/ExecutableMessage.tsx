import type { ConversationMessage, MessageEmbeddings, ReactionSummary } from "@/api/conversations";
import { ExecutionPreviewCard } from "./ExecutionPreviewCard";
import { MessageActions } from "./MessageActions";
import { MessageAttachments } from "./MessageAttachments";
import { ReactionBar } from "@/components/ui/reaction-bar";
import { ReactionPicker } from "./ReactionPicker";
import { cn } from "@/lib/utils";
import { formatDistanceToNow } from "date-fns";
import { Check, CheckCheck } from "lucide-react";
import { useState } from "react";

export interface ExecutableMessageProps {
  message: ConversationMessage;
  isOwn?: boolean;
  authorDisplayName?: string;
  className?: string;
  username?: string;
  currentUserId?: string;
  readReceipts?: { user_id: string; read_at: string }[];
  onReactionAdd?: (reaction: string) => void;
  onReactionRemove?: (reaction: string) => void;
}

const UUID_REGEX = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
function isUUID(s: string): boolean {
  return UUID_REGEX.test(s);
}

function parseEmbeddings(embeddings: ConversationMessage["embeddings"]): MessageEmbeddings | null {
  if (!embeddings || typeof embeddings !== "object") return null;
  const e = embeddings as Record<string, unknown>;
  if (e.execution_id || e.execution_root_hash || e.function_ref) return e as MessageEmbeddings;
  return null;
}

export function ExecutableMessage({
  message,
  isOwn = false,
  authorDisplayName,
  className,
  username,
  currentUserId,
  readReceipts,
  onReactionAdd,
  onReactionRemove,
}: ExecutableMessageProps) {
  const emb = parseEmbeddings(message.embeddings);
  const functionRef = emb?.function_ref;
  const hasExecutionRef =
    emb && (emb.execution_id || emb.execution_root_hash || functionRef);
  const [showReactionPicker, setShowReactionPicker] = useState(false);

  const hasUserReacted = (reaction: string) => {
    return message.reactions?.some(
      (r) => r.reaction === reaction && r.user_ids.includes(currentUserId || "")
    ) ?? false;
  };

  const handleReactionClick = (reaction: string) => {
    if (hasUserReacted(reaction)) {
      onReactionRemove?.(reaction);
    } else {
      onReactionAdd?.(reaction);
    }
  };

  const reactions = message.reactions?.map((r) => ({
    emoji: r.reaction,
    count: r.count,
    users: r.user_ids,
  })) ?? [];

  if (message.deleted_at) {
    return (
      <div
        className={cn(
          "flex flex-col gap-1 max-w-[85%]",
          isOwn && "self-end items-end",
          !isOwn && "self-start items-start",
          className
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
        "flex flex-col gap-2 max-w-[85%] group",
        isOwn && "self-end items-end",
        !isOwn && "self-start items-start",
        className
      )}
    >
      {authorDisplayName && !isOwn && (
        <span className="text-xs font-medium text-muted-foreground">{authorDisplayName}</span>
      )}
      <div className="relative">
        <div
          className={cn(
            "rounded-lg px-3 py-2 text-sm break-words",
            isOwn
              ? "bg-brand-500/15 text-brand-foreground border border-brand-500/30"
              : "bg-muted/60 text-foreground border border-border/60"
          )}
        >
          {message.content && <p className="whitespace-pre-wrap">{message.content}</p>}
          {message.edited_at && (
            <span className="block text-[10px] text-muted-foreground mt-1">(edited)</span>
          )}
        </div>
        {username && (
          <div className={cn(
            "absolute top-0.5",
            isOwn ? "-left-7" : "-right-7"
          )}>
            <MessageActions
              messageId={message.id}
              conversationId={message.conversation_id}
              username={username}
              currentContent={message.content}
              isOwn={isOwn}
              reactions={message.reactions}
              currentUserId={currentUserId}
            />
          </div>
        )}
      </div>
      {message.attachments && message.attachments.length > 0 && username && (
        <MessageAttachments
          attachments={message.attachments}
          username={username}
          conversationId={message.conversation_id}
          messageId={message.id}
          currentUserId={currentUserId}
        />
      )}
      {hasExecutionRef && functionRef && (
        <ExecutionPreviewCard
          author={functionRef.author}
          name={functionRef.name}
          executionId={emb!.execution_id && isUUID(emb!.execution_id) ? emb!.execution_id : undefined}
          executionRootHash={emb!.execution_root_hash}
          inputSummary={emb!.input_summary}
          outputSummary={emb!.output_summary}
          compact
          className="w-full max-w-md"
        />
      )}

      {reactions.length > 0 && (
        <ReactionBar
          reactions={reactions}
          onReact={handleReactionClick}
          onRemoveReact={handleReactionClick}
          className="mt-1"
        />
      )}

      <div className="flex items-center gap-2">
        <span className="text-[10px] text-muted-foreground px-1">
          {formatDistanceToNow(new Date(message.created_at), { addSuffix: true })}
        </span>

        {isOwn && readReceipts && readReceipts.length > 0 && (
          <div className="flex items-center gap-0.5 text-[10px] text-muted-foreground">
            <CheckCheck className="h-3 w-3 text-brand-500" />
            <span>{readReceipts.length}</span>
          </div>
        )}

        {!isOwn && readReceipts && readReceipts.length > 0 && (
          <div className="flex items-center gap-0.5 text-[10px] text-muted-foreground/60">
            <Check className="h-3 w-3" />
          </div>
        )}

        {onReactionAdd && (
          <button
            onClick={() => setShowReactionPicker(!showReactionPicker)}
            className="opacity-0 group-hover:opacity-100 transition-opacity text-muted-foreground hover:text-foreground p-0.5"
            title="Add reaction"
          >
            <span className="text-xs">+</span>
          </button>
        )}
      </div>

      {showReactionPicker && onReactionAdd && (
        <div className="flex gap-1 rounded-lg border border-border bg-card px-2 py-1 shadow-md text-sm">
          {['👍', '❤️', '😂', '🎉', '👀', '🚀'].map((emoji) => (
            <button
              key={emoji}
              onClick={() => {
                handleReactionClick(emoji);
                setShowReactionPicker(false);
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
