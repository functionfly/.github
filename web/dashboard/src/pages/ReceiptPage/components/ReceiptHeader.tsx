// ReceiptHeader — function name, author, runtime, verified shield.
import { CheckCircle2, ExternalLink } from "lucide-react";
import { Link } from "react-router-dom";

import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";

import { getRuntimeStyle } from "../lib/runtime-badge";
import type { Receipt } from "../types";

interface ReceiptHeaderProps {
  receipt: Receipt;
}

export function ReceiptHeader({ receipt }: ReceiptHeaderProps) {
  const { function: fn, execution } = receipt;
  const runtime = getRuntimeStyle(fn.runtime);
  const RuntimeIcon = runtime.icon;
  const authorInitial = (fn.author || "?").charAt(0).toUpperCase();
  const isVerified = execution.verification?.status === "verified";

  return (
    <header className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
      <div className="flex items-start gap-4">
        <Avatar className="h-12 w-12 border border-border/50">
          <AvatarFallback className="bg-gradient-to-br from-primary/20 to-primary/5 text-base font-semibold">
            {authorInitial}
          </AvatarFallback>
        </Avatar>
        <div className="space-y-1.5">
          <div className="flex flex-wrap items-center gap-2">
            <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl">
              {fn.author}/{fn.name}
            </h1>
            <Badge variant="outline" className={`gap-1.5 ${runtime.className}`}>
              <RuntimeIcon className="h-3 w-3" aria-hidden />
              {runtime.label}
            </Badge>
            <Badge variant="secondary" className="font-mono text-xs">
              v{fn.version}
            </Badge>
            {isVerified ? (
              <Badge
                variant="outline"
                className="gap-1 border-emerald-500/30 bg-emerald-500/10 text-emerald-400"
                title="This execution was replay-verified by the platform"
              >
                <CheckCircle2 className="h-3 w-3" aria-hidden /> Verified
              </Badge>
            ) : null}
          </div>
          {fn.description ? (
            <p className="max-w-2xl text-sm text-muted-foreground">{fn.description}</p>
          ) : null}
          <div className="flex items-center gap-3 text-sm text-muted-foreground">
            <Link
              to={`/u/${encodeURIComponent(fn.author)}`}
              className="inline-flex items-center gap-1 hover:text-foreground"
            >
              by {fn.author}
              <ExternalLink className="h-3 w-3" aria-hidden />
            </Link>
            <Link
              to={`/fx/${encodeURIComponent(fn.author)}/${encodeURIComponent(fn.name)}`}
              className="inline-flex items-center gap-1 hover:text-foreground"
            >
              View function
              <ExternalLink className="h-3 w-3" aria-hidden />
            </Link>
          </div>
        </div>
      </div>
    </header>
  );
}

export default ReceiptHeader;
