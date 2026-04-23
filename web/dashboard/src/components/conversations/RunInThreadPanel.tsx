import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { Play, Loader2, Share2, AlertCircle } from "lucide-react";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { registryApi, type RegistryExecutionResponse } from "@/api/registry";
import { conversationsApi, type MessageEmbeddings } from "@/api/conversations";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { cn } from "@/lib/utils";

function truncate(str: string, maxLen: number): string {
  if (str.length <= maxLen) return str;
  return str.slice(0, maxLen) + "…";
}

function summarize(obj: unknown): string {
  if (obj === null || obj === undefined) return "";
  try {
    const s = typeof obj === "string" ? obj : JSON.stringify(obj);
    return truncate(s, 120);
  } catch {
    return String(obj).slice(0, 120);
  }
}

export interface RunInThreadPanelProps {
  username: string;
  conversationId: string;
  /** Pre-filled function author (e.g. from context) */
  defaultAuthor?: string;
  /** Pre-filled function name */
  defaultName?: string;
  /** Pre-filled version */
  defaultVersion?: string;
  /** Callback after snippet was added to thread */
  onSnippetAdded?: () => void;
  className?: string;
}

export function RunInThreadPanel({
  username,
  conversationId,
  defaultAuthor = "",
  defaultName = "",
  defaultVersion = "",
  onSnippetAdded,
  className,
}: RunInThreadPanelProps) {
  const queryClient = useQueryClient();
  const [author, setAuthor] = useState(defaultAuthor);
  const [name, setName] = useState(defaultName);
  const [version, setVersion] = useState(defaultVersion);
  const [inputText, setInputText] = useState("{}");
  const [lastResult, setLastResult] = useState<RegistryExecutionResponse | null>(null);
  const [lastError, setLastError] = useState<string | null>(null);

  const runMutation = useMutation({
    mutationFn: async () => {
      let input: unknown;
      try {
        input = JSON.parse(inputText);
      } catch {
        throw new Error("Invalid JSON input");
      }
      const req = { input, ...(version ? { version } : {}) };
      if (version) {
        return registryApi.executeFunctionVersion(author, name, version, req);
      }
      return registryApi.executeFunction(author, name, req);
    },
    onSuccess: (data) => {
      setLastResult(data);
      setLastError(null);
      toast.success("Execution completed");
    },
    onError: (err: Error) => {
      setLastResult(null);
      setLastError(err.message || "Execution failed");
      toast.error(err.message || "Execution failed");
    },
  });

  const addSnippetMutation = useMutation({
    mutationFn: async () => {
      if (!lastResult) return;
      const embeddings: MessageEmbeddings = {
        function_ref: { author, name, version: version || lastResult.version },
        input_summary: summarize(JSON.parse(inputText)),
        output_summary: summarize(lastResult.data),
      };
      if (lastResult.execution_id) embeddings.execution_id = lastResult.execution_id;
      await conversationsApi.createMessage(username, conversationId, {
        content: "",
        embeddings,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["conversation-messages", username, conversationId] });
      toast.success("Execution snippet added to thread");
      onSnippetAdded?.();
    },
    onError: (err: Error) => {
      toast.error(err.message || "Failed to add snippet");
    },
  });

  const canRun = author.trim() && name.trim();
  const canAddSnippet = lastResult?.ok && lastResult.data !== undefined;

  return (
    <Card className={cn("border-border/60", className)}>
      <CardHeader className="pb-2">
        <div className="flex items-center gap-2">
          <Play className="h-4 w-4 text-muted-foreground" />
          <span className="text-sm font-medium">Run in thread</span>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="grid grid-cols-2 gap-2">
          <div>
            <Label className="text-xs">Author</Label>
            <Input
              placeholder="e.g. acme"
              value={author}
              onChange={(e) => setAuthor(e.target.value)}
              className="h-8 mt-0.5"
            />
          </div>
          <div>
            <Label className="text-xs">Name</Label>
            <Input
              placeholder="e.g. slugify"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="h-8 mt-0.5"
            />
          </div>
        </div>
        <div>
          <Label className="text-xs">Version (optional)</Label>
          <Input
            placeholder="latest if empty"
            value={version}
            onChange={(e) => setVersion(e.target.value)}
            className="h-8 mt-0.5"
          />
        </div>
        <div>
          <Label className="text-xs">Input (JSON)</Label>
          <textarea
            placeholder='{"key": "value"}'
            value={inputText}
            onChange={(e) => setInputText(e.target.value)}
            className="mt-0.5 w-full min-h-[80px] rounded-md border border-input bg-background px-3 py-2 font-mono text-xs"
            spellCheck={false}
          />
        </div>
        <div className="flex gap-2">
          <Button
            size="sm"
            onClick={() => runMutation.mutate()}
            disabled={!canRun || runMutation.isPending}
            className="gap-1"
          >
            {runMutation.isPending ? (
              <Loader2 className="h-3.5 w-3.5 animate-spin" />
            ) : (
              <Play className="h-3.5 w-3.5" />
            )}
            Run
          </Button>
          {canAddSnippet && (
            <Button
              size="sm"
              variant="outline"
              onClick={() => addSnippetMutation.mutate()}
              disabled={addSnippetMutation.isPending}
              className="gap-1"
            >
              {addSnippetMutation.isPending ? (
                <Loader2 className="h-3.5 w-3.5 animate-spin" />
              ) : (
                <Share2 className="h-3.5 w-3.5" />
              )}
              Add to thread as snippet
            </Button>
          )}
        </div>
        {lastError && (
          <div className="flex items-start gap-2 rounded-md bg-destructive/10 border border-destructive/20 p-2 text-sm">
            <AlertCircle className="h-4 w-4 text-destructive shrink-0 mt-0.5" />
            <span className="text-destructive">{lastError}</span>
          </div>
        )}
        {lastResult?.ok && lastResult.data !== undefined && (
          <div className="rounded-md bg-muted/50 border border-border/60 p-2">
            <Label className="text-xs text-muted-foreground">Output</Label>
            <pre className="mt-1 font-mono text-xs break-words whitespace-pre-wrap">
              {typeof lastResult.data === "object"
                ? JSON.stringify(lastResult.data, null, 2)
                : String(lastResult.data)}
            </pre>
            <p className="text-xs text-muted-foreground mt-1">
              {lastResult.duration_ms}ms
              {lastResult.execution_id && ` · ID: ${lastResult.execution_id}`}
            </p>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
