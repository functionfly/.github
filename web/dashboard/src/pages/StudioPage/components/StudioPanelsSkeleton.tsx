import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

interface StudioPanelSkeletonProps {
  className?: string;
  showHeader?: boolean;
  headerTitle?: string;
}

export function StudioPanelSkeleton({
  className,
  showHeader = true,
  headerTitle = "Loading...",
}: StudioPanelSkeletonProps) {
  return (
    <div className={cn("p-3 space-y-4", className)}>
      {showHeader && (
        <div className="border-b border-border-subtle pb-3">
          <Skeleton className="h-4 w-32 mb-1" />
          <Skeleton className="h-3 w-48" />
        </div>
      )}
      <div className="space-y-3">{headerTitle && <div className="text-xs text-text-muted">{headerTitle}</div>}</div>
    </div>
  );
}

export function AgentsPanelSkeleton() {
  return (
    <div className="flex-1 overflow-y-auto">
      <div className="p-3 space-y-4">
        {/* AgentDock skeleton */}
        <div className="space-y-2">
          {[1, 2, 3].map((i) => (
            <div key={i} className="flex items-center gap-3 p-2 rounded-lg border border-border-subtle">
              <Skeleton className="size-8 rounded-full" />
              <div className="flex-1 space-y-1.5">
                <Skeleton className="h-3 w-24" />
                <Skeleton className="h-2 w-16" />
              </div>
              <Skeleton className="h-5 w-12 rounded-full" />
            </div>
          ))}
        </div>

        {/* AgentLifecyclePanel skeleton */}
        <div className="space-y-3">
          <Skeleton className="h-4 w-full" />
          <div className="flex gap-2">
            <Skeleton className="h-8 flex-1 rounded-lg" />
            <Skeleton className="h-8 flex-1 rounded-lg" />
          </div>
        </div>

        {/* AgentMemoryViewer skeleton */}
        <div className="space-y-2">
          <Skeleton className="h-3 w-24" />
          <Skeleton className="h-20 w-full rounded-lg" />
        </div>
      </div>
    </div>
  );
}

export function RuntimePanelSkeleton() {
  return (
    <div className="p-3 space-y-4">
      <div className="border-b border-border-subtle pb-3">
        <Skeleton className="h-4 w-32 mb-1" />
        <Skeleton className="h-3 w-48" />
      </div>

      {/* RuntimeTargetSelector skeleton */}
      <div className="space-y-2">
        <Skeleton className="h-24 w-full rounded-lg" />
      </div>

      {/* RuntimeCapabilityMatrix skeleton */}
      <div className="grid grid-cols-2 gap-2">
        {[1, 2, 3, 4].map((i) => (
          <Skeleton key={i} className="h-12 rounded-lg" />
        ))}
      </div>

      {/* Capabilities */}
      <div className="border-t border-border-subtle pt-4 space-y-2">
        <Skeleton className="h-3 w-32" />
        <div className="grid grid-cols-2 gap-2">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <Skeleton key={i} className="h-8 rounded-lg" />
          ))}
        </div>
      </div>

      {/* WasmExecutionPanel skeleton */}
      <Skeleton className="h-32 w-full rounded-lg" />
    </div>
  );
}

export function MarketplacePanelSkeleton() {
  return (
    <div className="p-3 space-y-3">
      {/* FunctionMarketplace search skeleton */}
      <div className="space-y-2">
        <Skeleton className="h-8 w-full rounded-lg" />
        <div className="flex gap-2">
          <Skeleton className="h-6 w-20 rounded-full" />
          <Skeleton className="h-6 w-20 rounded-full" />
          <Skeleton className="h-6 w-20 rounded-full" />
        </div>
      </div>

      <div className="border-t border-border-subtle pt-3 mt-3 space-y-3">
        <Skeleton className="h-3 w-24" />

        {/* CreatorProfile skeleton */}
        <Skeleton className="h-24 w-full rounded-lg" />

        {/* Stats grid skeleton */}
        <div className="grid grid-cols-3 gap-2">
          <Skeleton className="h-16 rounded-lg" />
          <Skeleton className="h-16 rounded-lg" />
          <Skeleton className="h-16 rounded-lg" />
        </div>

        {/* UsageBillingPanel skeleton */}
        <Skeleton className="h-24 w-full rounded-lg" />

        {/* SubscriptionManager skeleton */}
        <Skeleton className="h-20 w-full rounded-lg" />

        {/* Royalties and Leaderboard skeleton */}
        <div className="grid grid-cols-2 gap-2">
          <Skeleton className="h-32 rounded-lg" />
          <Skeleton className="h-32 rounded-lg" />
        </div>

        {/* RevenueAnalytics skeleton */}
        <Skeleton className="h-40 w-full rounded-lg" />

        {/* LicenseManager skeleton */}
        <Skeleton className="h-16 w-full rounded-lg" />

        {/* MonetizationOptimizer skeleton */}
        <Skeleton className="h-24 w-full rounded-lg" />

        {/* AssetPricingEditor skeleton */}
        <Skeleton className="h-16 w-full rounded-lg" />
      </div>
    </div>
  );
}

export function SwarmPanelSkeleton() {
  return (
    <div className="p-3 space-y-4">
      <div className="border-b border-border-subtle pb-3">
        <Skeleton className="h-4 w-32 mb-1" />
        <Skeleton className="h-3 w-48" />
      </div>

      {/* Stats grid skeleton */}
      <div className="grid grid-cols-2 gap-2 mb-4">
        <Skeleton className="h-16 rounded-lg" />
        <Skeleton className="h-16 rounded-lg" />
      </div>

      {/* SwarmCoordinator skeleton */}
      <Skeleton className="h-40 w-full rounded-lg" />

      {/* Topology views skeleton */}
      <div className="border-t border-border-subtle pt-4 space-y-2">
        <Skeleton className="h-3 w-24" />
        <div className="grid grid-cols-3 gap-2">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <Skeleton key={i} className="h-8 rounded-lg" />
          ))}
        </div>
      </div>
    </div>
  );
}

export function SkillsPanelSkeleton() {
  return (
    <div className="p-3 space-y-4">
      <div className="border-b border-border-subtle pb-3">
        <Skeleton className="h-4 w-32 mb-1" />
        <Skeleton className="h-3 w-48" />
      </div>

      {/* Stats grid skeleton */}
      <div className="grid grid-cols-2 gap-2 mb-4">
        <Skeleton className="h-16 rounded-lg" />
        <Skeleton className="h-16 rounded-lg" />
      </div>

      {/* Skill Graph skeleton */}
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <Skeleton className="size-4" />
          <Skeleton className="h-3 w-20" />
        </div>
        <Skeleton className="h-40 w-full rounded-lg" />
      </div>

      {/* Dependency Map skeleton */}
      <div className="border-t border-border-subtle pt-4 space-y-2">
        <div className="flex items-center gap-2">
          <Skeleton className="size-4" />
          <Skeleton className="h-3 w-28" />
        </div>
        <Skeleton className="h-40 w-full rounded-lg" />
      </div>
    </div>
  );
}

export function CanvasPanelSkeleton() {
  return (
    <div className="p-3 space-y-3">
      {/* GlassCard skeleton */}
      <div className="p-3 rounded-lg border border-border-subtle">
        <div className="flex items-center gap-2 mb-2">
          <Skeleton className="size-4" />
          <Skeleton className="h-4 w-24" />
          <Skeleton className="h-5 w-16 rounded-full ml-auto" />
        </div>
        <Skeleton className="h-3 w-48 mb-3" />

        {/* Tabs skeleton */}
        <div className="flex gap-2 mb-2">
          <Skeleton className="h-6 flex-1 rounded" />
          <Skeleton className="h-6 flex-1 rounded" />
          <Skeleton className="h-6 flex-1 rounded" />
        </div>

        {/* FunctionCanvas placeholder skeleton */}
        <Skeleton className="h-[300px] w-full rounded-lg" />
      </div>

      {/* AgentLifecyclePanel skeleton */}
      <div className="space-y-2">
        <Skeleton className="h-4 w-full" />
        <div className="flex gap-2">
          <Skeleton className="h-8 flex-1 rounded-lg" />
          <Skeleton className="h-8 flex-1 rounded-lg" />
        </div>
      </div>

      {/* AgentMemoryViewer skeleton */}
      <div className="space-y-2">
        <Skeleton className="h-3 w-24" />
        <Skeleton className="h-20 w-full rounded-lg" />
      </div>
    </div>
  );
}

export function ExecutionPanelSkeleton() {
  return (
    <div className="flex-1 flex items-center justify-center">
      <div className="text-center space-y-4">
        <Skeleton className="size-16 rounded-full mx-auto" />
        <div className="space-y-2">
          <Skeleton className="h-4 w-48 mx-auto" />
          <Skeleton className="h-3 w-32 mx-auto" />
        </div>
        <Skeleton className="h-8 w-32 mx-auto rounded-lg" />
      </div>
    </div>
  );
}

export function SimulationPanelSkeleton() {
  return (
    <div className="p-3 space-y-4 overflow-y-auto">
      {/* SimulationControlCenter skeleton */}
      <Skeleton className="h-32 w-full rounded-lg" />

      {/* Metrics grid skeleton */}
      <div className="grid grid-cols-2 gap-3">
        <Skeleton className="h-40 rounded-lg" />
        <Skeleton className="h-40 rounded-lg" />
      </div>

      {/* Latency and Cost skeleton */}
      <div className="grid grid-cols-2 gap-3">
        <Skeleton className="h-40 rounded-lg" />
        <Skeleton className="h-40 rounded-lg" />
      </div>

      {/* HallucinationRiskAnalyzer skeleton */}
      <Skeleton className="h-32 w-full rounded-lg" />

      {/* Stress, Scaling, Agent skeletons */}
      <div className="grid grid-cols-3 gap-3">
        <Skeleton className="h-40 rounded-lg" />
        <Skeleton className="h-40 rounded-lg" />
        <Skeleton className="h-40 rounded-lg" />
      </div>
    </div>
  );
}

export function GhostPanelSkeleton() {
  return (
    <div className="p-3 space-y-4">
      {/* GhostModeOrchestrator skeleton */}
      <Skeleton className="h-40 w-full rounded-lg" />

      {/* Conversation Timeline skeleton */}
      <div className="border-t border-border-subtle pt-4 space-y-2">
        <div className="flex items-center gap-2">
          <Skeleton className="size-3" />
          <Skeleton className="h-3 w-36" />
        </div>
        <Skeleton className="h-24 w-full rounded-lg" />
      </div>

      {/* MultiAgentConversationView skeleton */}
      <div className="border-t border-border-subtle pt-4 space-y-2">
        <Skeleton className="h-3 w-44" />
        <Skeleton className="h-32 w-full rounded-lg" />
      </div>
    </div>
  );
}

export function TasksPanelSkeleton() {
  return (
    <div className="p-3 space-y-4">
      {/* AutonomousTaskBoard skeleton */}
      <div className="space-y-3">
        <Skeleton className="h-6 w-full rounded-lg" />
        <div className="grid grid-cols-4 gap-2">
          <div className="space-y-2">
            <Skeleton className="h-4 w-16" />
            <Skeleton className="h-20 w-full rounded-lg" />
            <Skeleton className="h-20 w-full rounded-lg" />
          </div>
          <div className="space-y-2">
            <Skeleton className="h-4 w-16" />
            <Skeleton className="h-20 w-full rounded-lg" />
          </div>
          <div className="space-y-2">
            <Skeleton className="h-4 w-16" />
            <Skeleton className="h-20 w-full rounded-lg" />
            <Skeleton className="h-20 w-full rounded-lg" />
          </div>
          <div className="space-y-2">
            <Skeleton className="h-4 w-16" />
            <Skeleton className="h-20 w-full rounded-lg" />
          </div>
        </div>
      </div>
    </div>
  );
}

export function DevOpsPanelSkeleton() {
  return (
    <div className="p-3 space-y-4">
      <div className="border-b border-border-subtle pb-3">
        <Skeleton className="h-4 w-32 mb-1" />
        <Skeleton className="h-3 w-48" />
      </div>

      {/* Stats grid skeleton */}
      <div className="grid grid-cols-3 gap-2 mb-4">
        <Skeleton className="h-16 rounded-lg" />
        <Skeleton className="h-16 rounded-lg" />
        <Skeleton className="h-16 rounded-lg" />
      </div>

      {/* DeploymentPipeline skeleton */}
      <Skeleton className="h-40 w-full rounded-lg" />

      {/* Environments skeleton */}
      <div className="border-t border-border-subtle pt-4 space-y-2">
        <Skeleton className="h-3 w-20" />
        <Skeleton className="h-32 w-full rounded-lg" />
      </div>

      {/* Cloud Regions skeleton */}
      <div className="border-t border-border-subtle pt-4 space-y-2">
        <Skeleton className="h-3 w-24" />
        <Skeleton className="h-32 w-full rounded-lg" />
      </div>
    </div>
  );
}

export function MemoryPanelSkeleton() {
  return (
    <div className="p-3 space-y-4">
      <div className="border-b border-border-subtle pb-3">
        <Skeleton className="h-4 w-32 mb-1" />
        <Skeleton className="h-3 w-48" />
      </div>

      {/* Stats grid skeleton */}
      <div className="grid grid-cols-3 gap-2 mb-4">
        <Skeleton className="h-16 rounded-lg" />
        <Skeleton className="h-16 rounded-lg" />
        <Skeleton className="h-16 rounded-lg" />
      </div>

      {/* MemoryGraph skeleton */}
      <Skeleton className="h-64 w-full rounded-lg" />

      {/* SemanticMemoryViewer skeleton */}
      <div className="border-t border-border-subtle pt-4 space-y-2">
        <Skeleton className="h-3 w-28" />
        <Skeleton className="h-32 w-full rounded-lg" />
      </div>

      {/* LongTermContextExplorer skeleton */}
      <div className="border-t border-border-subtle pt-4 space-y-2">
        <Skeleton className="h-3 w-32" />
        <Skeleton className="h-40 w-full rounded-lg" />
      </div>
    </div>
  );
}

export function RoboticsPanelSkeleton() {
  return (
    <div className="p-3 space-y-4">
      <div className="border-b border-border-subtle pb-3">
        <Skeleton className="h-4 w-32 mb-1" />
        <Skeleton className="h-3 w-48" />
      </div>

      {/* Stats grid skeleton */}
      <div className="grid grid-cols-3 gap-2 mb-4">
        <Skeleton className="h-16 rounded-lg" />
        <Skeleton className="h-16 rounded-lg" />
        <Skeleton className="h-16 rounded-lg" />
      </div>

      {/* RobotFleetDashboard skeleton */}
      <Skeleton className="h-80 w-full rounded-lg" />

      {/* Sensor Telemetry skeleton */}
      <div className="border-t border-border-subtle pt-4 space-y-2">
        <Skeleton className="h-3 w-28" />
        <Skeleton className="h-24 w-full rounded-lg" />
      </div>

      {/* Device Mesh skeleton */}
      <div className="border-t border-border-subtle pt-4 space-y-2">
        <Skeleton className="h-3 w-24" />
        <Skeleton className="h-64 w-full rounded-lg" />
      </div>
    </div>
  );
}

export function TelemetryPanelSkeleton() {
  return (
    <div className="p-3 space-y-4">
      {/* LiveTelemetryPanel skeleton */}
      <Skeleton className="h-40 w-full rounded-lg" />

      {/* Cost Heatmap skeleton */}
      <div className="border-t border-border-subtle pt-4 space-y-2">
        <div className="flex items-center gap-2">
          <Skeleton className="size-3" />
          <Skeleton className="h-3 w-20" />
        </div>
        <Skeleton className="h-32 w-full rounded-lg" />
      </div>

      {/* Latency Distribution skeleton */}
      <div className="border-t border-border-subtle pt-4 space-y-2">
        <div className="flex items-center gap-2">
          <Skeleton className="size-3" />
          <Skeleton className="h-3 w-28" />
        </div>
        <Skeleton className="h-32 w-full rounded-lg" />
      </div>
    </div>
  );
}

export function VisualizationPanelSkeleton() {
  return (
    <div className="h-full flex flex-col">
      {/* Header skeleton */}
      <div className="flex items-center justify-between px-4 py-2 border-b border-border-subtle">
        <Skeleton className="h-4 w-32" />
        <div className="flex items-center gap-2">
          <Skeleton className="h-6 w-16 rounded" />
          <Skeleton className="size-6 rounded" />
          <Skeleton className="size-6 rounded" />
        </div>
      </div>

      {/* Content skeleton */}
      <div className="flex-1 flex items-center justify-center">
        <Skeleton className="size-32 rounded-full" />
      </div>
    </div>
  );
}

export function ProfilerPanelSkeleton() {
  return (
    <div className="p-3 space-y-4">
      {/* Stats grid skeleton */}
      <div className="grid grid-cols-2 gap-2">
        <Skeleton className="h-16 rounded-lg" />
        <Skeleton className="h-16 rounded-lg" />
        <Skeleton className="h-16 rounded-lg" />
        <Skeleton className="h-16 rounded-lg" />
      </div>

      {/* ExecutionProfiler skeleton */}
      <div className="border-t border-border-subtle pt-4 space-y-2">
        <Skeleton className="h-48 w-full rounded-lg" />
      </div>
    </div>
  );
}

export function CollabPanelSkeleton() {
  return (
    <div className="p-3 space-y-4">
      {/* Activity skeleton */}
      <div className="space-y-2">
        <Skeleton className="h-4 w-24" />
        <div className="space-y-2">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="flex items-center gap-2">
              <Skeleton className="size-6 rounded-full" />
              <div className="flex-1 space-y-1">
                <Skeleton className="h-3 w-full" />
                <Skeleton className="h-2 w-16" />
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Comments skeleton */}
      <div className="border-t border-border-subtle pt-4 space-y-2">
        <Skeleton className="h-4 w-20" />
        <Skeleton className="h-24 w-full rounded-lg" />
      </div>

      {/* Annotations skeleton */}
      <div className="border-t border-border-subtle pt-4 space-y-2">
        <Skeleton className="h-4 w-24" />
        <Skeleton className="h-24 w-full rounded-lg" />
      </div>

      {/* GraphEdits skeleton */}
      <div className="border-t border-border-subtle pt-4 space-y-2">
        <Skeleton className="h-4 w-22" />
        <Skeleton className="h-24 w-full rounded-lg" />
      </div>
    </div>
  );
}
