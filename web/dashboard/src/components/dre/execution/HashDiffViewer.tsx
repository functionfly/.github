import { useState } from "react";
import { AlertTriangle, CheckCircle, XCircle, ArrowLeftRight } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export type DriftCategory = "instruction" | "memory" | "resource" | "output" | "environment" | "dependency" | null;

export interface HashDiffViewerProps {
  /** Hash from original execution */
  hash1: string;
  /** Label for hash1 */
  label1?: string;
  /** Hash from replay execution */
  hash2: string;
  /** Label for hash2 */
  label2?: string;
  /** Whether hashes match */
  match: boolean;
  /** Drift category if not matching */
  driftCategory?: DriftCategory;
  /** Divergent component name */
  divergentComponent?: string;
  /** Instruction mismatch index (if applicable) */
  mismatchIndex?: number;
  /** Callback when drift is selected */
  onDriftSelect?: (category: DriftCategory) => void;
  /** Custom className */
  className?: string;
}

export function HashDiffViewer({
  hash1,
  label1 = "Original",
  hash2,
  label2 = "Replay",
  match,
  driftCategory,
  divergentComponent,
  mismatchIndex,
  onDriftSelect,
  className,
}: HashDiffViewerProps) {
  const [copied1, setCopied1] = useState(false);
  const [copied2, setCopied2] = useState(false);

  const copyHash = (hash: string, setCopied: React.Dispatch<React.SetStateAction<boolean>>) => {
    navigator.clipboard.writeText(hash);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  // Find the first position where hashes differ
  const findDiffPosition = () => {
    if (match) return -1;
    const minLen = Math.min(hash1.length, hash2.length);
    for (let i = 0; i < minLen; i++) {
      if (hash1[i] !== hash2[i]) return i;
    }
    return minLen;
  };

  const diffPosition = findDiffPosition();

  const renderHashWithDiff = (hash: string, label: string, copied: boolean, setCopied: React.Dispatch<React.SetStateAction<boolean>>) => (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
          {label}
        </span>
        <Button
          variant="ghost"
          size="sm"
          onClick={() => copyHash(hash, setCopied)}
          className="h-6 px-2 text-xs"
        >
          {copied ? "Copied!" : "Copy"}
        </Button>
      </div>
      <div className="bg-bg-secondary rounded-md p-3 font-mono text-sm break-all">
        {!match && diffPosition > 0 ? (
          <>
            <span className="text-muted-foreground">{hash.slice(0, diffPosition)}</span>
            <span className="bg-red-500/20 text-red-500 font-bold">{hash[diffPosition]}</span>
            <span className="text-muted-foreground">{hash.slice(diffPosition + 1)}</span>
          </>
        ) : (
          hash
        )}
      </div>
    </div>
  );

  return (
    <Card className={className}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base flex items-center gap-2">
            <ArrowLeftRight className="h-4 w-4" />
            Hash Comparison
          </CardTitle>
          {match ? (
            <div className="flex items-center gap-2 text-green-500">
              <CheckCircle className="h-5 w-5" />
              <span className="text-sm font-medium">Match</span>
            </div>
          ) : (
            <div className="flex items-center gap-2 text-red-500">
              <XCircle className="h-5 w-5" />
              <span className="text-sm font-medium">Drift Detected</span>
            </div>
          )}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Hash Display */}
        <div className="grid md:grid-cols-2 gap-4">
          {renderHashWithDiff(hash1, label1, copied1, setCopied1)}
          {renderHashWithDiff(hash2, label2, copied2, setCopied2)}
        </div>

        {/* Drift Details */}
        {!match && driftCategory && (
          <div className="mt-4 p-4 bg-red-500/10 border border-red-500/20 rounded-lg">
            <div className="flex items-start gap-3">
              <AlertTriangle className="h-5 w-5 text-red-500 shrink-0 mt-0.5" />
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <span className="font-medium text-red-500">Drift Type:</span>
                  <span className="capitalize">{driftCategory}</span>
                </div>
                {divergentComponent && (
                  <div className="flex items-center gap-2">
                    <span className="text-sm text-muted-foreground">Component:</span>
                    <span className="text-sm font-medium">{divergentComponent}</span>
                  </div>
                )}
                {mismatchIndex !== undefined && (
                  <div className="flex items-center gap-2">
                    <span className="text-sm text-muted-foreground">Mismatch at index:</span>
                    <span className="text-sm font-mono">{mismatchIndex}</span>
                  </div>
                )}
                {onDriftSelect && (
                  <Button
                    variant="outline"
                    size="sm"
                    className="mt-2"
                    onClick={() => onDriftSelect(driftCategory)}
                  >
                    View Full Report
                  </Button>
                )}
              </div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
