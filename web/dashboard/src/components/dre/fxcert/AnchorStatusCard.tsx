import { Link2, ExternalLink, CheckCircle } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

export interface AnchorStatusCardProps {
  /** Whether the certificate is anchored */
  anchored: boolean;
  /** Blockchain name */
  chain?: string;
  /** Block number */
  blockNumber?: number;
  /** Transaction hash */
  txHash?: string;
  /** Anchor timestamp */
  timestamp?: string;
  /** Custom className */
  className?: string;
}

export function AnchorStatusCard({
  anchored,
  chain,
  blockNumber,
  txHash,
  timestamp,
  className,
}: AnchorStatusCardProps) {
  const getExplorerUrl = () => {
    if (!chain || !txHash) return null;
    
    const explorers: Record<string, string> = {
      ethereum: `https://etherscan.io/tx/${txHash}`,
      polygon: `https://polygonscan.com/tx/${txHash}`,
      arbitrum: `https://arbiscan.io/tx/${txHash}`,
      optimism: `https://optimistic.etherscan.io/tx/${txHash}`,
      solana: `https://solscan.io/tx/${txHash}`,
    };
    
    return explorers[chain.toLowerCase()] || `https://${chain.toLowerCase()}.com/tx/${txHash}`;
  };

  const explorerUrl = getExplorerUrl();

  return (
    <Card className={cn(className)}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base flex items-center gap-2">
            <Link2 className="h-4 w-4" />
            Blockchain Anchor
          </CardTitle>
          {anchored ? (
            <Badge className="bg-green-500/10 text-green-500 border-green-500/20">
              <CheckCircle className="h-3.5 w-3.5 mr-1" />
              Anchored
            </Badge>
          ) : (
            <Badge variant="outline">Not Anchored</Badge>
          )}
        </div>
      </CardHeader>
      <CardContent>
        {anchored && chain ? (
          <div className="space-y-3">
            {/* Chain */}
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Chain</span>
              <Badge variant="secondary">{chain}</Badge>
            </div>

            {/* Block Number */}
            {blockNumber && (
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Block</span>
                <span className="font-mono text-sm">#{blockNumber.toLocaleString()}</span>
              </div>
            )}

            {/* Transaction Hash */}
            {txHash && (
              <div className="space-y-1">
                <span className="text-sm text-muted-foreground">Transaction</span>
                <div className="flex items-center gap-2">
                  <code className="flex-1 font-mono text-xs bg-bg-secondary px-2 py-1 rounded break-all">
                    {txHash}
                  </code>
                  {explorerUrl && (
                    <a
                      href={explorerUrl}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="shrink-0 p-1 hover:bg-bg-secondary rounded transition-colors"
                    >
                      <ExternalLink className="h-4 w-4 text-muted-foreground" />
                    </a>
                  )}
                </div>
              </div>
            )}

            {/* Timestamp */}
            {timestamp && (
              <div className="flex items-center justify-between pt-2 border-t border-border-subtle">
                <span className="text-sm text-muted-foreground">Anchored At</span>
                <span className="text-sm">
                  {new Date(timestamp).toLocaleString()}
                </span>
              </div>
            )}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">
            This execution certificate has not been anchored to a blockchain.
          </p>
        )}
      </CardContent>
    </Card>
  );
}
