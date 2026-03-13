import type { Conversation } from "@/api/conversations";
import { ExecutionPreviewCard } from "./ExecutionPreviewCard";
import { Button } from "@/components/ui/button";
import { CheckCircle } from "lucide-react";
import { cn } from "@/lib/utils";

export interface FixModeMetadata {
  problem_statement?: string;
  reproduction_steps?: string[];
  execution_hash_refs?: Array<{ author: string; name: string; execution_root_hash?: string; execution_id?: string }>;
  suggested_patch?: string;
}

export interface FixModeLayoutProps {
  conversation: Conversation;
  onAcceptSolution?: () => void;
  isResolved?: boolean;
  className?: string;
}

export function FixModeLayout({
  conversation,
  onAcceptSolution,
  isResolved,
  className,
}: FixModeLayoutProps) {
  const meta = (conversation.metadata || {}) as FixModeMetadata;
  const problem = meta.problem_statement;
  const steps = meta.reproduction_steps ?? [];
  const refs = meta.execution_hash_refs ?? [];
  const patch = meta.suggested_patch;

  return (
    <div className={cn("space-y-4", className)}>
      {problem && (
        <section>
          <h3 className="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-1">
            Problem
          </h3>
          <p className="text-sm whitespace-pre-wrap rounded-md bg-muted/50 p-3">{problem}</p>
        </section>
      )}
      {steps.length > 0 && (
        <section>
          <h3 className="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-1">
            Reproduction
          </h3>
          <ol className="list-decimal list-inside space-y-1 text-sm">
            {steps.map((s, i) => (
              <li key={i}>{s}</li>
            ))}
          </ol>
        </section>
      )}
      {refs.length > 0 && (
        <section>
          <h3 className="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-2">
            Execution refs
          </h3>
          <div className="space-y-2">
            {refs.map((r, i) => (
              <ExecutionPreviewCard
                key={i}
                author={r.author}
                name={r.name}
                executionId={r.execution_id}
                executionRootHash={r.execution_root_hash}
                compact
              />
            ))}
          </div>
        </section>
      )}
      {patch && (
        <section>
          <h3 className="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-1">
            Suggested patch
          </h3>
          <pre className="text-xs font-mono whitespace-pre-wrap rounded-md bg-muted/50 p-3 overflow-x-auto">
            {patch}
          </pre>
        </section>
      )}
      {!isResolved && onAcceptSolution && (
        <Button onClick={onAcceptSolution} className="gap-1">
          <CheckCircle className="h-4 w-4" />
          Accept solution
        </Button>
      )}
    </div>
  );
}
