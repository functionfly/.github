import { CheckCircle, XCircle, AlertTriangle, ExternalLink, FileText } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { HashBlock } from "../primitives/HashBlock";
import { cn } from "@/lib/utils";
import { DriftCategory } from "../execution/HashDiffViewer";

export interface ReplayResultCardProps {
  /** Replay root hash */
  replayRootHash: string;
  /** Original execution root hash */
  originalRootHash?: string;
  /** Whether roots match */
  match: boolean;
  /** Drift category if mismatch */
  driftCategory?: DriftCategory;
  /** URL to drift report */
  driftReportUrl?: string;
  /** Replay node ID */
  replayNodeId?: string;
  /** Replay duration in ms */
  durationMs?: number;
  /** Callback when view details is clicked */
  onViewDetails?: () => void;
  /** Custom className */
  className?: string;
}

const driftCategoryLabels: Record<string, { label: string; color: string }> = {
  instruction: { label: "Instruction Mismatch", color: "text-red-500" },
  memory: { label: "Memory Drift", color: "text-orange-500" },
  resource: { label: "Resource Variance", color: "text-yellow-500" },
  output: { label: "Output Difference", color: "text-red-500" },
  environment: { label: "Environment Drift", color: "text-orange-500" },
  dependency: { label: "Dependency Mismatch", color: "text-yellow-500" },
};

export function ReplayResultCard({
  replayRootHash,
  originalRootHash,
  match,
  driftCategory,
  driftReportUrl,
  replayNodeId,
  durationMs,
  onViewDetails,
  className,
}: ReplayResultCardProps) {
  return (
    <Card className={className}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base flex items-center gap-2">
            Replay Result
          </CardTitle>
          {match ? (
            <Badge className="bg-green-500/10 text-green-500 border-green-500/20">
              <CheckCircle className="h-3.5 w-3.5 mr-1" />
              Match
            </Badge>
          ) : (
            <Badge className="bg-red-500/10 text-red-500 border-red-500/20">
              <XCircle className="h-3.5 w-3.5 mr-1" />
              Drift Detected
            </Badge>
          )}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Replay Root Hash */}
        <HashBlock
          hash={replayRootHash}
          label="Replay Root Hash"
          truncate
          truncateChars={12}
          verified={match}
          invalid={!match}
        />

        {/* Original Root Hash (if different) */}
        {originalRootHash && !match && (
          <HashBlock
            hash={originalRootHash}
            label="Original Root Hash"
            truncate
            truncateChars={12}
          />
        )}

        {/* Drift Info */}
        {!match && driftCategory && (
          <div className="flex items-center gap-2 p-3 bg-red-500/10 border border-red-500/20 rounded-lg">
            <AlertTriangle className="h-5 w-5 text-red-500 shrink-0" />
            <div className="flex-1">
              <span className={cn("font-medium", driftCategoryLabels[driftCategory]?.color)}>
                {driftCategoryLabels[driftCategory]?.label || driftCategory}
              </span>
            </div>
            {driftReportUrl && (
              <a
                href={driftReportUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="text-blue-500 hover:underline text-sm flex items-center gap-1"
              >
                View Report
                <ExternalLink className="h-3 w-3" />
              </a>
            )}
          </div>
        )}

        {/* Metadata */}
        <div className="flex flex-wrap gap-4 text-sm">
          {replayNodeId && (
            <div>
              <span className="text-muted-foreground">Node: </span>
              <span className="font-mono text-xs">{replayNodeId}</span>
            </div>
          )}
          {durationMs !== undefined && (
            <div>
              <span className="text-muted-foreground">Duration: </span>
              <span>{durationMs}ms</span>
            </div>
          )}
        </div>

        {/* Actions */}
        {onViewDetails && (
          <Button variant="outline" size="sm" onClick={onViewDetails} className="gap-2">
            <FileText className="h-4 w-4" />
            View Full Details
          </Button>
        )}
      </CardContent>
    </Card>
  );
}
