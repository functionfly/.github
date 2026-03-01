import { CheckCircle, XCircle, Clock, Shield } from "lucide-react";
import { cn } from "@/lib/utils";

export interface SignatureVerificationBadgeProps {
  /** Node signature status */
  nodeVerified: boolean;
  /** Platform signature status */
  platformVerified: boolean;
  /** Whether the certificate is expired */
  expired?: boolean;
  /** Custom className */
  className?: string;
}

export function SignatureVerificationBadge({
  nodeVerified,
  platformVerified,
  expired = false,
  className,
}: SignatureVerificationBadgeProps) {
  const allVerified = nodeVerified && platformVerified && !expired;
  const someVerified = nodeVerified || platformVerified;

  return (
    <div className={cn("flex flex-wrap gap-2", className)}>
      {/* Node Signature */}
      <div
        className={cn(
          "flex items-center gap-1.5 px-2 py-1 rounded-md text-xs font-medium border",
          nodeVerified
            ? "bg-green-500/10 text-green-500 border-green-500/20"
            : "bg-red-500/10 text-red-500 border-red-500/20"
        )}
      >
        {nodeVerified ? (
          <CheckCircle className="h-3.5 w-3.5" />
        ) : (
          <XCircle className="h-3.5 w-3.5" />
        )}
        <span>Node Signature</span>
      </div>

      {/* Platform Signature */}
      <div
        className={cn(
          "flex items-center gap-1.5 px-2 py-1 rounded-md text-xs font-medium border",
          platformVerified
            ? "bg-green-500/10 text-green-500 border-green-500/20"
            : "bg-red-500/10 text-red-500 border-red-500/20"
        )}
      >
        {platformVerified ? (
          <CheckCircle className="h-3.5 w-3.5" />
        ) : (
          <XCircle className="h-3.5 w-3.5" />
        )}
        <span>Platform Signature</span>
      </div>

      {/* Expiry Status */}
      {expired && (
        <div className="flex items-center gap-1.5 px-2 py-1 rounded-md text-xs font-medium border bg-yellow-500/10 text-yellow-500 border-yellow-500/20">
          <Clock className="h-3.5 w-3.5" />
          <span>Expired</span>
        </div>
      )}

      {/* Overall Status */}
      {allVerified && (
        <div className="flex items-center gap-1.5 px-2 py-1 rounded-md text-xs font-medium border bg-blue-500/10 text-blue-500 border-blue-500/20">
          <Shield className="h-3.5 w-3.5" />
          <span>Fully Verified</span>
        </div>
      )}
    </div>
  );
}
