// ReceiptRunPanel — "Run with this input" form. The killer feature: lets
// any visitor re-execute the function with one click.
import { AlertCircle, Loader2, Play, Zap } from "lucide-react";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Textarea } from "@/components/ui/textarea";

import { useReceiptRun } from "../hooks/useReceiptRun";
import { prettyJSON } from "../lib/schema-render";
import type { Receipt } from "../types";

interface ReceiptRunPanelProps {
  receipt: Receipt;
  /** App origin (window.location.origin by default). */
  appOrigin?: string;
  /** Whether the current user is signed in. Used for the paid gate. */
  isAuthenticated: boolean;
}

export function ReceiptRunPanel({ receipt, appOrigin, isAuthenticated }: ReceiptRunPanelProps) {
  const navigate = useNavigate();
  const [override, setOverride] = useState<string>(prettyJSON(receipt.execution.input, 2));
  const [hasEdited, setHasEdited] = useState(false);
  const runMut = useReceiptRun();
  const origin = appOrigin ?? (typeof window !== "undefined" ? window.location.origin : "");

  const isPaid = receipt.is_paid;
  const canRunPublic = receipt.can_run && receipt.function.visibility === "public";

  const handleRun = async () => {
    let parsed: unknown = null;
    if (hasEdited) {
      try {
        parsed = override.trim() === "" ? null : JSON.parse(override);
      } catch (err) {
        toast.error("Invalid JSON", {
          description: (err as Error).message,
        });
        return;
      }
    }
    try {
      const result = await runMut.mutateAsync({ receiptId: receipt.id, input: parsed });
      if (result.ok && result.execution_id) {
        toast.success("New receipt generated", {
          description: "Click to view the fresh run.",
          action: {
            label: "Open",
            onClick: () => navigate(`/r/${result.execution_id}`),
          },
        });
      } else {
        toast.error("Run failed", {
          description: result.error?.message ?? "Unknown error",
        });
      }
    } catch (err) {
      toast.error("Run failed", {
        description: (err as Error).message,
      });
    }
  };

  if (!isPaid) {
    return (
      <Card data-testid="receipt-run-panel">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <Play className="h-4 w-4" aria-hidden /> Run with this input
          </CardTitle>
          <CardDescription>
            Anyone can re-execute a public function. The new run gets its own shareable receipt.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <Textarea
            value={override}
            onChange={(e) => {
              setOverride(e.target.value);
              setHasEdited(true);
            }}
            rows={Math.min(10, Math.max(4, override.split("\n").length))}
            className="font-mono text-xs"
            spellCheck={false}
            aria-label="Function input (JSON)"
            data-testid="receipt-run-input"
          />
          <div className="flex items-center justify-between">
            <p className="text-xs text-muted-foreground">
              Pre-filled with the recorded input. Edit and re-run, or click Run as-is.
            </p>
            <Button
              onClick={handleRun}
              disabled={runMut.isPending || !canRunPublic}
              className="gap-2"
              data-testid="receipt-run-submit"
            >
              {runMut.isPending ? (
                <Loader2 className="h-4 w-4 animate-spin" aria-hidden />
              ) : (
                <Zap className="h-4 w-4" aria-hidden />
              )}
              {runMut.isPending ? "Running…" : "Run function"}
            </Button>
          </div>
          {runMut.isError ? (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertTitle>Run failed</AlertTitle>
              <AlertDescription>{(runMut.error as Error).message}</AlertDescription>
            </Alert>
          ) : null}
        </CardContent>
      </Card>
    );
  }

  // Paid function gate
  return (
    <Card data-testid="receipt-run-panel-paid">
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          <Zap className="h-4 w-4 text-amber-500" aria-hidden /> Paid function
        </CardTitle>
        <CardDescription>
          This function costs ${receipt.price_per_call_usd.toFixed(4)} per call. Sign in to run it.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {isAuthenticated ? (
          <Button
            onClick={handleRun}
            disabled={runMut.isPending}
            className="gap-2"
            data-testid="receipt-run-submit"
          >
            {runMut.isPending ? <Loader2 className="h-4 w-4 animate-spin" aria-hidden /> : <Zap className="h-4 w-4" aria-hidden />}
            Run for ${receipt.price_per_call_usd.toFixed(4)}
          </Button>
        ) : (
          <Button asChild>
            <a href={`${origin}/auth/login?next=${encodeURIComponent(typeof window !== "undefined" ? window.location.pathname : "/")}`}>
              Sign in to run
            </a>
          </Button>
        )}
      </CardContent>
    </Card>
  );
}

export default ReceiptRunPanel;
