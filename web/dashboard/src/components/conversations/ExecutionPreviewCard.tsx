import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { Play, Hash, AlertCircle, Loader2 } from "lucide-react";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { dreApi, type ExecutionDetail } from "@/api/dre";
import { ExecutionRootBadge, VerificationBadge } from "@/components/dre";
import { cn } from "@/lib/utils";

export interface ExecutionPreviewCardProps {
  /** Function author (registry) */
  author: string;
  /** Function name */
  name: string;
  /** Execution ID if known */
  executionId?: string;
  /** Execution root hash if known (used when executionId not set) */
  executionRootHash?: string;
  /** Optional input summary from message embeddings */
  inputSummary?: string;
  /** Optional output summary from message embeddings */
  outputSummary?: string;
  /** Error message if execution failed (from embeddings or fetch) */
  errorTrace?: string;
  /** Compact layout for inline in message */
  compact?: boolean;
  className?: string;
}

function useExecutionDetail(
  author: string,
  name: string,
  executionId?: string,
  executionRootHash?: string
) {
  return useQuery({
    queryKey: ["execution-preview", author, name, executionId ?? executionRootHash],
    queryFn: async (): Promise<ExecutionDetail | null> => {
      if (executionId) {
        const res = await dreApi.getExecution(author, name, executionId);
        return res.execution as unknown as ExecutionDetail;
      }
      if (executionRootHash) {
        const res = await dreApi.getExecutionByHash(author, name, executionRootHash);
        return res.execution as unknown as ExecutionDetail;
      }
      return null;
    },
    enabled: Boolean(author && name && (executionId || executionRootHash)),
  });
}

export function ExecutionPreviewCard({
  author,
  name,
  executionId,
  executionRootHash,
  inputSummary,
  outputSummary,
  errorTrace,
  compact = false,
  className,
}: ExecutionPreviewCardProps) {
  const { data: execution, isLoading, error } = useExecutionDetail(
    author,
    name,
    executionId,
    executionRootHash
  );

  const functionRef = `fx://${author}/${name}`;
  const hash = execution?.execution_root_hash ?? executionRootHash ?? "";
  const verified = Boolean(execution?.replay_verified_at);
  const determinismTier = execution?.determinism_tier ?? "full";
  const isDeterministic = determinismTier === "full";
  const execId = execution?.execution_id ?? executionId;

  if (isLoading && !execution && !inputSummary && !outputSummary && !hash) {
    return (
      <Card className={cn("border-border/60", className)}>
        <CardContent className="flex items-center gap-2 py-4">
          <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
          <span className="text-sm text-muted-foreground">Loading execution…</span>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className={cn("border-border/60 bg-card/80", compact && "shadow-sm", className)}>
      <CardHeader className={cn("pb-2", compact && "py-3")}>
        <div className="flex flex-wrap items-center gap-2">
          <span className="text-xs font-medium text-muted-foreground">Execution</span>
          <code className="text-xs bg-muted px-1.5 py-0.5 rounded">{functionRef}</code>
          {execution?.version && (
            <Badge variant="secondary" className="text-xs">
              v{execution.version}
            </Badge>
          )}
          {isDeterministic ? (
            <VerificationBadge status="verified" size="sm" showIcon={false} />
          ) : (
            <Badge variant="outline" className="text-amber-600 border-amber-500/50 text-xs">
              Non-deterministic
            </Badge>
          )}
        </div>
      </CardHeader>
      <CardContent className={cn("space-y-3", compact && "space-y-2 pt-0")}>
        {(inputSummary || outputSummary) && (
          <div className="grid grid-cols-1 gap-2 text-sm">
            {inputSummary && (
              <div>
                <span className="text-muted-foreground">Input: </span>
                <span className="font-mono text-xs break-all">{inputSummary}</span>
              </div>
            )}
            {outputSummary && (
              <div>
                <span className="text-muted-foreground">Output: </span>
                <span className="font-mono text-xs break-all">{outputSummary}</span>
              </div>
            )}
          </div>
        )}
        {hash && (
          <div className="flex items-start gap-2">
            <Hash className="h-4 w-4 text-muted-foreground shrink-0 mt-0.5" />
            <ExecutionRootBadge
              hash={hash}
              verified={verified}
              truncate={true}
              className="flex-1 min-w-0"
            />
          </div>
        )}
        {errorTrace && (
          <div className="flex items-start gap-2 rounded-md bg-destructive/10 border border-destructive/20 p-2 text-sm">
            <AlertCircle className="h-4 w-4 text-destructive shrink-0 mt-0.5" />
            <pre className="whitespace-pre-wrap break-words font-mono text-xs text-destructive">
              {errorTrace}
            </pre>
          </div>
        )}
        {error && (
          <div className="flex items-center gap-2 text-sm text-destructive">
            <AlertCircle className="h-4 w-4 shrink-0" />
            <span>Could not load execution details.</span>
          </div>
        )}
        {execId && (
          <div className="flex gap-2 pt-1">
            <Button variant="outline" size="sm" asChild className="gap-1">
              <Link to={`/replay/${execId}`}>
                <Play className="h-3.5 w-3.5" />
                Replay
              </Link>
            </Button>
            <Button variant="ghost" size="sm" asChild className="gap-1">
              <Link to={`/registry/${author}/${name}/executions`}>
                View all executions
              </Link>
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
