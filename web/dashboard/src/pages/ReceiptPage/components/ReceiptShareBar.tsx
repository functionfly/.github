// ReceiptShareBar — copy link / tweet / QR. All three are non-blocking
// fire-and-forget actions so the user can keep exploring the receipt.
import { Check, Copy } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";

import { XBrandIcon } from "../lib/brand-icons";
import type { Receipt } from "../types";

interface ReceiptShareBarProps {
  receipt: Receipt;
}

function copy(value: string, label: string) {
  return async () => {
    try {
      await navigator.clipboard.writeText(value);
      toast.success(`${label} copied`);
    } catch {
      toast.error("Copy failed", { description: "Your browser blocked clipboard access." });
    }
  };
}

export function ReceiptShareBar({ receipt }: ReceiptShareBarProps) {
  const [copiedLink, setCopiedLink] = useState(false);

  const onCopyLink = async () => {
    try {
      await navigator.clipboard.writeText(receipt.share.url);
      setCopiedLink(true);
      window.setTimeout(() => setCopiedLink(false), 1500);
      toast.success("Receipt link copied");
    } catch {
      toast.error("Copy failed");
    }
  };

  return (
    <div className="flex flex-wrap items-center gap-2" data-testid="receipt-share-bar">
      <Button
        type="button"
        variant="outline"
        size="sm"
        onClick={onCopyLink}
        className="gap-1.5"
        aria-label="Copy receipt link"
      >
        {copiedLink ? <Check className="h-3.5 w-3.5" aria-hidden /> : <Copy className="h-3.5 w-3.5" aria-hidden />}
        {copiedLink ? "Copied" : "Copy link"}
      </Button>
      <Button
        type="button"
        variant="outline"
        size="sm"
        asChild
        className="gap-1.5"
        aria-label="Share on X"
      >
        <a href={receipt.share.tweet_intent_url} target="_blank" rel="noopener noreferrer">
          <XBrandIcon size={14} className="h-3.5 w-3.5" aria-hidden /> Tweet
        </a>
      </Button>
      <Button
        type="button"
        variant="ghost"
        size="sm"
        onClick={copy(receipt.share.embed_url, "Embed URL")}
        className="gap-1.5 text-muted-foreground"
        aria-label="Copy embed URL"
      >
        Embed
      </Button>
    </div>
  );
}

export default ReceiptShareBar;
