// ReceiptForkCTA — "Deploy your own function — free" hero card with the
// 1→2→3 explainer. The distribution lever.
import { ExternalLink, GitBranch, Loader2 } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

import { useReceiptFork } from "../hooks/useReceiptFork";
import type { Receipt } from "../types";

interface ReceiptForkCTAProps {
  receipt: Receipt;
  isAuthenticated: boolean;
  appOrigin?: string;
}

export function ReceiptForkCTA({ receipt, isAuthenticated, appOrigin }: ReceiptForkCTAProps) {
  const navigate = useNavigate();
  const fork = useReceiptFork({ receiptId: receipt.id, isAuthenticated, appOrigin });
  const link = fork.buildForkLink();

  const onFork = () => {
    if (!link) {
      // Payload still loading — wait one tick.
      toast.message("Loading source…");
      return;
    }
    if (link.target === "editor") {
      navigate(link.href);
    } else {
      // Sign-up flow with a `next=` back to the editor
      window.location.href = link.href;
    }
  };

  return (
    <Card
      className="overflow-hidden border-primary/30 bg-gradient-to-br from-primary/10 via-primary/5 to-transparent"
      data-testid="receipt-fork-cta"
    >
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-lg">
          <GitBranch className="h-5 w-5 text-primary" aria-hidden />
          Deploy your own function — free
        </CardTitle>
        <CardDescription>
          Fork {receipt.function.author}/{receipt.function.name} into your own account. No credit card required.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-5">
        <ol className="grid grid-cols-1 gap-3 sm:grid-cols-3">
          {[
            { n: 1, t: "Fork", d: "One click. The source is copied to your account." },
            { n: 2, t: "Edit", d: "Tweak the code, change the schema, ship your variant." },
            { n: 3, t: "Deploy", d: "Your function goes live with its own public URL and receipt." },
          ].map((s) => (
            <li
              key={s.n}
              className="flex flex-col gap-1 rounded-lg border border-border/40 bg-background/60 p-3"
            >
              <span className="text-xs font-mono text-muted-foreground">Step {s.n}</span>
              <span className="text-sm font-semibold">{s.t}</span>
              <span className="text-xs text-muted-foreground">{s.d}</span>
            </li>
          ))}
        </ol>
        <div className="flex flex-wrap items-center justify-between gap-2">
          <div className="text-xs text-muted-foreground">
            {link ? (
              <>
                Source: <span className="font-mono">{link.sourceBytes.toLocaleString()} bytes</span>
                {link.target !== "editor" ? (
                  <>
                    {" "}
                    · You&apos;ll be asked to sign up first.
                  </>
                ) : null}
              </>
            ) : (
              <span className="inline-flex items-center gap-1.5">
                <Loader2 className="h-3 w-3 animate-spin" aria-hidden /> Loading source…
              </span>
            )}
          </div>
          <Button
            onClick={onFork}
            disabled={fork.isLoading}
            size="lg"
            className="gap-2"
            data-testid="receipt-fork-submit"
          >
            <GitBranch className="h-4 w-4" aria-hidden />
            Fork this function
            <ExternalLink className="h-3.5 w-3.5 opacity-70" aria-hidden />
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

export default ReceiptForkCTA;
