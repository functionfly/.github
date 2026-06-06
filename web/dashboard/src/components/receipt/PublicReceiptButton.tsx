// Shared UI primitives for the Execution Receipt feature.
//
// These components render the "View public receipt" CTA that appears in
// the playground, replay, function page, and other surfaces. Keep them
// tiny and free of dashboard-specific dependencies so they can be
// imported from anywhere.

import { ExternalLink, Link2, Share2 } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { API_BASE_URL, ROUTE_BUILDERS } from '@/lib/constants';

import { encodeInputForUrl } from '@/pages/ReceiptPage/lib/og-meta';

// PublicReceiptButton — primary CTA. Opens the public receipt in a new
// tab. Use on the playground output panel and the replay page header.
export function PublicReceiptButton({
  publicId,
  variant = 'default',
  size = 'sm',
  className,
  label = 'View public receipt',
}: {
  publicId: string;
  variant?: 'default' | 'outline' | 'ghost';
  size?: 'sm' | 'default' | 'lg';
  className?: string;
  label?: string;
}) {
  const url = ROUTE_BUILDERS.receipt(publicId);
  return (
    <Button
      asChild
      variant={variant}
      size={size}
      className={className}
      data-testid="public-receipt-button"
    >
      <a href={url} target="_blank" rel="noopener noreferrer" className="gap-1.5">
        <ExternalLink className="h-3.5 w-3.5" aria-hidden />
        {label}
      </a>
    </Button>
  );
}

// PublicReceiptLink — inline text link. Use inside dense surfaces (the
// playground output headers tab) where a full button would crowd the
// layout. Opens in a new tab — use a regular anchor, not react-router
// <Link>, so the user keeps their current tab.
export function PublicReceiptLink({
  publicId,
  className,
  children,
}: {
  publicId: string;
  className?: string;
  children?: React.ReactNode;
}) {
  const url = ROUTE_BUILDERS.receipt(publicId);
  return (
    <a
      href={url}
      target="_blank"
      rel="noopener noreferrer"
      className={className ?? 'inline-flex items-center gap-1 text-xs text-indigo-400 hover:text-indigo-300'}
    >
      {children ?? (
        <>
          <Link2 className="h-3 w-3" aria-hidden />
          Public receipt
        </>
      )}
    </a>
  );
}

// ReceiptShareMenu — compact share row (Copy link / X). Mirrors the
// in-page bar on /r/:id but smaller so it fits in playground headers and
// function page sidebars.
export function ReceiptShareMenu({
  publicId,
  functionName,
  functionAuthor,
  className,
}: {
  publicId: string;
  functionName: string;
  functionAuthor: string;
  className?: string;
}) {
  const url = `${typeof window !== 'undefined' ? window.location.origin : API_BASE_URL}${ROUTE_BUILDERS.receipt(publicId)}`;
  const text = `I just ran ${functionAuthor}/${functionName} on @functionfly`;
  const tweetIntent = `https://twitter.com/intent/tweet?${new URLSearchParams({ text, url }).toString()}`;

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(url);
    } catch {
      // ignore
    }
  };

  return (
    <div className={className ?? 'flex items-center gap-1.5'}>
      <Button
        asChild
        variant="ghost"
        size="sm"
        className="h-7 gap-1 text-xs text-muted-foreground hover:text-foreground"
        aria-label="Open public receipt in a new tab"
      >
        <a href={url} target="_blank" rel="noopener noreferrer">
          <ExternalLink className="h-3 w-3" aria-hidden /> Open
        </a>
      </Button>
      <Button
        type="button"
        variant="ghost"
        size="sm"
        onClick={copy}
        className="h-7 gap-1 text-xs text-muted-foreground hover:text-foreground"
        aria-label="Copy receipt link"
      >
        <Link2 className="h-3 w-3" aria-hidden /> Copy
      </Button>
      <Button
        asChild
        variant="ghost"
        size="sm"
        className="h-7 gap-1 text-xs text-muted-foreground hover:text-foreground"
        aria-label="Tweet this receipt"
      >
        <a href={tweetIntent} target="_blank" rel="noopener noreferrer">
          <Share2 className="h-3 w-3" aria-hidden /> Tweet
        </a>
      </Button>
    </div>
  );
}

// buildRunWithInputLink — builds a /run/:author/:name?input=<base64> URL
// for the "Run again with this input" CTA. Exported here so other
// surfaces (e.g. the function page's executions tab) can offer the
// same affordance without duplicating the encode logic.
export function buildRunWithInputLink(
  author: string,
  name: string,
  input: unknown,
): string {
  const encoded = encodeInputForUrl(input);
  if (!encoded) return ROUTE_BUILDERS.playground(author, name);
  return `${ROUTE_BUILDERS.playground(author, name)}?input=${encoded}`;
}
