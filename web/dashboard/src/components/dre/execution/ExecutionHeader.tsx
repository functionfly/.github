import { MapPin, Server, Hash } from "lucide-react";
import { DeterminismBadge } from "../primitives/DeterminismBadge";
import { VerificationBadge } from "../primitives/VerificationBadge";
import { ProtocolVersionTag } from "../primitives/ProtocolVersionTag";
import { cn } from "@/lib/utils";

export interface ExecutionHeaderProps {
  /** Execution ID */
  executionId: string;
  /** Function ID (fx:// URI) */
  functionId?: string;
  /** Determinism tier */
  determinismTier: "full" | "lite" | "partial" | "drifted";
  /** Trust score at execution (0-100) */
  trustScore: number;
  /** Region where execution happened */
  region?: string;
  /** Node ID */
  nodeId?: string;
  /** Verification status */
  verified: boolean;
  /** Protocol version */
  protocolVersion: string;
  /** Is latest protocol version */
  isLatestProtocol?: boolean;
  /** Custom className */
  className?: string;
}

export function ExecutionHeader({
  executionId,
  functionId,
  determinismTier,
  trustScore,
  region,
  nodeId,
  verified,
  protocolVersion,
  isLatestProtocol = false,
  className,
}: ExecutionHeaderProps) {
  return (
    <div className={cn("space-y-4", className)}>
      {/* Title Row */}
      <div className="flex flex-wrap items-center gap-3">
        <h2 className="text-xl font-semibold">Execution Details</h2>
        <VerificationBadge
          status={verified ? "verified" : "pending"}
          size="sm"
        />
        <DeterminismBadge tier={determinismTier} size="sm" />
      </div>

      {/* Metadata Grid */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
        {/* Execution ID */}
        <div className="space-y-1">
          <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
            Execution ID
          </span>
          <div className="flex items-center gap-1.5">
            <Hash className="h-3.5 w-3.5 text-muted-foreground" />
            <span className="font-mono text-xs truncate" title={executionId}>
              {executionId.slice(0, 12)}...
            </span>
          </div>
        </div>

        {/* Function ID */}
        {functionId && (
          <div className="space-y-1">
            <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
              Function
            </span>
            <div className="font-mono text-xs truncate" title={functionId}>
              {functionId}
            </div>
          </div>
        )}

        {/* Region */}
        {region && (
          <div className="space-y-1">
            <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
              Region
            </span>
            <div className="flex items-center gap-1.5">
              <MapPin className="h-3.5 w-3.5 text-muted-foreground" />
              <span>{region}</span>
            </div>
          </div>
        )}

        {/* Node */}
        {nodeId && (
          <div className="space-y-1">
            <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
              Node
            </span>
            <div className="flex items-center gap-1.5">
              <Server className="h-3.5 w-3.5 text-muted-foreground" />
              <span className="font-mono text-xs truncate" title={nodeId}>
                {nodeId.slice(0, 8)}...
              </span>
            </div>
          </div>
        )}
      </div>

      {/* Trust Score & Protocol */}
      <div className="flex flex-wrap items-center gap-6">
        {/* Trust Score */}
        <div className="space-y-1">
          <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
            Trust Score
          </span>
          <div className="flex items-center gap-2">
            <div className="w-24 h-2 bg-bg-secondary rounded-full overflow-hidden">
              <div
                className={cn(
                  "h-full rounded-full transition-all",
                  trustScore >= 80
                    ? "bg-green-500"
                    : trustScore >= 50
                    ? "bg-yellow-500"
                    : "bg-red-500"
                )}
                style={{ width: `${trustScore}%` }}
              />
            </div>
            <span className="text-sm font-medium">{trustScore}%</span>
          </div>
        </div>

        {/* Protocol Version */}
        <div className="space-y-1">
          <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
            Protocol
          </span>
          <ProtocolVersionTag
            version={protocolVersion}
            latest={isLatestProtocol}
          />
        </div>
      </div>
    </div>
  );
}
