import { cn } from '@/lib/utils';
import { Link2 } from 'lucide-react';
import { HashBlock } from '../primitives/HashBlock';
import { VerificationBadge } from '../primitives/VerificationBadge';

export interface ExecutionRootBadgeProps {
  /** Execution root hash */
  hash: string;
  /** Whether the execution is verified */
  verified: boolean;
  /** Whether the hash is anchored to blockchain */
  anchored?: boolean;
  /** Blockchain name (if anchored) */
  chain?: string;
  /** Block number (if anchored) */
  blockNumber?: number;
  /** Transaction hash (if anchored) */
  txHash?: string;
  /** Truncate hash display */
  truncate?: boolean;
  /** Custom className */
  className?: string;
}

export function ExecutionRootBadge({
  hash,
  verified,
  anchored = false,
  chain,
  blockNumber,
  txHash,
  truncate = true,
  className,
}: ExecutionRootBadgeProps) {
  return (
    <div className={cn('space-y-2', className)}>
      <div className="flex items-center gap-2 mb-2">
        <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
          Execution Root Hash
        </span>
        <VerificationBadge status={verified ? 'verified' : 'pending'} size="sm" showIcon={false} />
        {anchored && (
          <span className="inline-flex items-center gap-1 text-xs text-brand-500 dre-link-accent">
            <Link2 className="h-3 w-3" />
            Anchored
          </span>
        )}
      </div>

      <HashBlock hash={hash} truncate={truncate} truncateChars={16} verified={verified} />

      {/* Anchor Info */}
      {anchored && chain && (
        <div className="flex items-center gap-4 text-xs text-muted-foreground mt-2">
          <span className="font-medium">{chain}</span>
          {blockNumber && <span>Block #{blockNumber.toLocaleString()}</span>}
          {txHash && (
            <a
              href={`https://${chain.toLowerCase()}.com/tx/${txHash}`}
              target="_blank"
              rel="noopener noreferrer"
              className="dre-link-accent text-brand-500 hover:underline focus:outline-none focus:ring-2 focus:ring-brand-500/30 rounded"
            >
              View Transaction →
            </a>
          )}
        </div>
      )}
    </div>
  );
}
