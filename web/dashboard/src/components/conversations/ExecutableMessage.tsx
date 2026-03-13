import type { ConversationMessage, MessageEmbeddings } from "@/api/conversations";
import { ExecutionPreviewCard } from "./ExecutionPreviewCard";
import { cn } from "@/lib/utils";

export interface ExecutableMessageProps {
  message: ConversationMessage;
  isOwn?: boolean;
  authorDisplayName?: string;
  className?: string;
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
}: ExecutableMessageProps) {
  const emb = parseEmbeddings(message.embeddings);
  const functionRef = emb?.function_ref;
  const hasExecutionRef =
    emb && (emb.execution_id || emb.execution_root_hash || functionRef);

  return (
    <div
      className={cn(
        "flex flex-col gap-2 max-w-[85%]",
        isOwn && "self-end items-end",
        !isOwn && "self-start items-start",
        className
      )}
    >
      {authorDisplayName && !isOwn && (
        <span className="text-xs font-medium text-muted-foreground">{authorDisplayName}</span>
      )}
      <div
        className={cn(
          "rounded-lg px-3 py-2 text-sm break-words",
          isOwn
            ? "bg-brand-500/15 text-brand-foreground border border-brand-500/30"
            : "bg-muted/60 text-foreground border border-border/60"
        )}
      >
        {message.content && <p className="whitespace-pre-wrap">{message.content}</p>}
      </div>
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
    </div>
  );
}
