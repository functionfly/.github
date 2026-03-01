import { CheckCircle, Circle, Loader2, AlertCircle } from "lucide-react";
import { cn } from "@/lib/utils";

export type ReplayStage = 
  | "pending"
  | "capsule_init"
  | "dependency_load"
  | "instruction_replay"
  | "trace_verification"
  | "root_match"
  | "complete"
  | "error";

export interface ReplayProgressTimelineProps {
  /** Current stage */
  currentStage: ReplayStage;
  /** Progress percentage (0-100) */
  progress: number;
  /** Error message if any */
  error?: string;
  /** Custom className */
  className?: string;
}

interface StageInfo {
  key: ReplayStage;
  label: string;
  description: string;
}

const stages: StageInfo[] = [
  { key: "capsule_init", label: "Capsule Init", description: "Initializing deterministic capsule" },
  { key: "dependency_load", label: "Dependency Load", description: "Loading dependencies in sandbox" },
  { key: "instruction_replay", label: "Instruction Replay", description: "Replaying instructions" },
  { key: "trace_verification", label: "Trace Verification", description: "Verifying execution trace" },
  { key: "root_match", label: "Root Match", description: "Comparing root hashes" },
];

const stageOrder: ReplayStage[] = [
  "capsule_init",
  "dependency_load",
  "instruction_replay",
  "trace_verification",
  "root_match",
];

function getStageStatus(stage: ReplayStage, currentStage: ReplayStage): "complete" | "current" | "pending" | "error" {
  if (stage === "error" || currentStage === "error") return "error";
  
  const stageIndex = stageOrder.indexOf(stage);
  const currentIndex = stageOrder.indexOf(currentStage);
  
  if (stageIndex < currentIndex) return "complete";
  if (stageIndex === currentIndex) return "current";
  return "pending";
}

export function ReplayProgressTimeline({
  currentStage,
  progress,
  error,
  className,
}: ReplayProgressTimelineProps) {
  return (
    <div className={cn("space-y-4", className)}>
      {/* Progress Bar */}
      <div className="space-y-2">
        <div className="flex justify-between text-sm">
          <span className="text-muted-foreground">Progress</span>
          <span className="font-medium">{progress}%</span>
        </div>
        <div className="h-2 bg-bg-secondary rounded-full overflow-hidden">
          <div
            className={cn(
              "h-full rounded-full transition-all duration-500",
              currentStage === "error"
                ? "bg-red-500"
                : currentStage === "complete"
                ? "bg-green-500"
                : "bg-brand-500"
            )}
            style={{ width: `${progress}%` }}
          />
        </div>
      </div>

      {/* Timeline */}
      <div className="relative">
        {/* Connector Line */}
        <div className="absolute left-[11px] top-6 bottom-6 w-0.5 bg-border-subtle" />

        <div className="space-y-4">
          {stages.map((stage) => {
            const status = getStageStatus(stage.key, currentStage);
            
            return (
              <div key={stage.key} className="flex items-start gap-3">
                {/* Icon */}
                <div
                  className={cn(
                    "relative z-10 flex items-center justify-center w-6 h-6 rounded-full border-2 shrink-0",
                    status === "complete" && "bg-green-500 border-green-500",
                    status === "current" && "bg-brand-500 border-brand-500",
                    status === "pending" && "bg-bg-primary border-border-subtle",
                    status === "error" && "bg-red-500 border-red-500"
                  )}
                >
                  {status === "complete" && (
                    <CheckCircle className="h-4 w-4 text-white" />
                  )}
                  {status === "current" && (
                    <Loader2 className="h-3.5 w-3.5 text-white animate-spin" />
                  )}
                  {status === "pending" && (
                    <Circle className="h-3 w-3 text-muted-foreground" />
                  )}
                  {status === "error" && (
                    <AlertCircle className="h-4 w-4 text-white" />
                  )}
                </div>

                {/* Content */}
                <div className="flex-1 pt-0.5">
                  <div
                    className={cn(
                      "text-sm font-medium",
                      status === "current" && "text-brand-500",
                      status === "complete" && "text-green-500",
                      status === "error" && "text-red-500",
                      status === "pending" && "text-muted-foreground"
                    )}
                  >
                    {stage.label}
                  </div>
                  <div className="text-xs text-muted-foreground">
                    {stage.description}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Error Display */}
      {error && (
        <div className="flex items-start gap-2 p-3 bg-red-500/10 border border-red-500/20 rounded-md">
          <AlertCircle className="h-4 w-4 text-red-500 shrink-0 mt-0.5" />
          <p className="text-xs text-red-500">{error}</p>
        </div>
      )}
    </div>
  );
}
