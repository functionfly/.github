import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { Check, Copy } from 'lucide-react';
import { useState } from 'react';

export interface HashBlockProps {
  /** The hash value to display */
  hash: string;
  /** Label for the hash */
  label?: string;
  /** Whether to truncate the hash display */
  truncate?: boolean;
  /** Number of characters to show on each side when truncated */
  truncateChars?: number;
  /** Custom className */
  className?: string;
  /** Callback when hash is clicked */
  onClick?: () => void;
  /** Whether the hash is verified (adds green indicator) */
  verified?: boolean;
  /** Whether the hash is invalid (adds red indicator) */
  invalid?: boolean;
}

export function HashBlock({
  hash,
  label,
  truncate = false,
  truncateChars = 8,
  className,
  onClick,
  verified = false,
  invalid = false,
}: HashBlockProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = (e: React.MouseEvent) => {
    e.stopPropagation();
    navigator.clipboard.writeText(hash);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const displayHash =
    truncate && hash.length > truncateChars * 2 + 3
      ? `${hash.slice(0, truncateChars)}...${hash.slice(-truncateChars)}`
      : hash;

  return (
    <div className={cn('flex flex-col gap-1', className)}>
      {label && (
        <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
          {label}
        </span>
      )}
      <div
        className={cn(
          'flex items-center gap-2 bg-bg-secondary rounded-md px-3 py-2 font-mono text-sm text-foreground',
          onClick && 'cursor-pointer hover:bg-bg-tertiary transition-colors',
          verified && 'border-l-2 border-l-green-500',
          invalid && 'border-l-2 border-l-red-500'
        )}
        onClick={onClick}
        title={truncate && hash.length > (truncateChars ?? 8) * 2 + 3 ? hash : undefined}
      >
        <span className={cn('flex-1 break-all', truncate && 'select-none')}>{displayHash}</span>
        <Button
          variant="ghost"
          size="icon"
          className="h-6 w-6 shrink-0"
          onClick={handleCopy}
          title={copied ? 'Copied!' : 'Copy to clipboard'}
        >
          {copied ? (
            <Check className="h-3.5 w-3.5 text-green-500" />
          ) : (
            <Copy className="h-3.5 w-3.5 text-muted-foreground" />
          )}
        </Button>
        {verified && (
          <span className="shrink-0">
            <svg
              className="h-4 w-4 text-green-500"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
          </span>
        )}
        {invalid && (
          <span className="shrink-0">
            <svg
              className="h-4 w-4 text-red-500"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
          </span>
        )}
      </div>
    </div>
  );
}
