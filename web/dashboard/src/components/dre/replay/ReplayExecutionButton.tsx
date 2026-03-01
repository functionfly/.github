import { Play, AlertTriangle, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export type ReplayMode = "strict" | "lite" | "debug";

export interface ReplayExecutionButtonProps {
  /** Whether the button is disabled */
  disabled?: boolean;
  /** Whether the button is loading */
  loading?: boolean;
  /** Capsule version to replay */
  capsuleVersion?: string;
  /** Selected replay mode */
  mode?: ReplayMode;
  /** Callback when mode selection changes */
  onModeChange?: (mode: ReplayMode) => void;
  /** Callback when button is clicked */
  onClick?: () => void;
  /** Show determinism warnings */
  showWarning?: boolean;
  /** Warning message */
  warningMessage?: string;
  /** Custom className */
  className?: string;
}

const modeLabels: Record<ReplayMode, { label: string; description: string }> = {
  strict: {
    label: "Strict",
    description: "Full deterministic replay with all checks",
  },
  lite: {
    label: "Lite",
    description: "Fast replay with minimal verification",
  },
  debug: {
    label: "Debug",
    description: "Detailed debugging with step-by-step trace",
  },
};

export function ReplayExecutionButton({
  disabled = false,
  loading = false,
  capsuleVersion,
  mode = "strict",
  onModeChange,
  onClick,
  showWarning = false,
  warningMessage,
  className,
}: ReplayExecutionButtonProps) {
  return (
    <div className={cn("space-y-3", className)}>
      {/* Mode Selector */}
      <div className="flex flex-wrap gap-2">
        {(Object.keys(modeLabels) as ReplayMode[]).map((m) => (
          <Button
            key={m}
            variant={mode === m ? "default" : "outline"}
            size="sm"
            onClick={() => onModeChange?.(m)}
            disabled={disabled || loading}
            className={cn(
              mode === m && "ring-2 ring-brand-500/50"
            )}
          >
            {modeLabels[m].label}
          </Button>
        ))}
      </div>

      {/* Warning */}
      {showWarning && (
        <div className="flex items-start gap-2 p-3 bg-yellow-500/10 border border-yellow-500/20 rounded-md">
          <AlertTriangle className="h-4 w-4 text-yellow-500 shrink-0 mt-0.5" />
          <p className="text-xs text-yellow-500">
            {warningMessage || "This execution may have non-deterministic components. Replay results may differ."}
          </p>
        </div>
      )}

      {/* Main Button */}
      <Button
        onClick={onClick}
        disabled={disabled || loading}
        className="w-full gap-2"
      >
        {loading ? (
          <Loader2 className="h-4 w-4 animate-spin" />
        ) : (
          <Play className="h-4 w-4" />
        )}
        Replay in Deterministic Sandbox
      </Button>

      {/* Version Info */}
      {capsuleVersion && (
        <p className="text-xs text-muted-foreground text-center">
          Using capsule version {capsuleVersion}
        </p>
      )}
    </div>
  );
}
