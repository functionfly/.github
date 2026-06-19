import { useStudioRuntimes } from "@/hooks/useStudio";
import type { RuntimeSelection, RuntimeDescriptor } from "@functionfly/ui-runtime";
import {
  RuntimeCapabilityMatrix,
  RuntimeTargetSelector,
  WasmExecutionPanel,
} from "@functionfly/ui-runtime";
import { Cpu, Globe, Shield, Zap } from "lucide-react";
import { RuntimePanelSkeleton } from "../components/StudioPanelsSkeleton";

interface RuntimePanelProps {
  selectedRuntime: RuntimeSelection | null;
  onSelect: (runtimeId: string) => void;
}

const CAPABILITIES = [
  { id: "sandboxed", label: "Sandboxed", icon: Shield, color: "text-success" },
  { id: "fast-cold-start", label: "Fast Cold Start", icon: Zap, color: "text-warning" },
  { id: "portable", label: "Portable", icon: Globe, color: "text-brand-400" },
  { id: "full-node-api", label: "Full Node API", icon: Cpu, color: "text-info" },
  { id: "fast-startup", label: "Fast Startup", icon: Zap, color: "text-success" },
  { id: "secure-by-default", label: "Secure by Default", icon: Shield, color: "text-success" },
  { id: "incremental", label: "Incremental Exec", icon: Cpu, color: "text-brand-400" },
];

export function RuntimePanel({ selectedRuntime, onSelect }: RuntimePanelProps) {
  const { runtimes, isLoading, error } = useStudioRuntimes();

  if (isLoading) {
    return <RuntimePanelSkeleton />;
  }

  return (
    <div className="p-3 space-y-4">
      <div className="border-b border-border-subtle pb-3">
        <h3 className="text-sm font-medium mb-1">Runtime Selection</h3>
        <p className="text-xs text-text-muted">Choose execution environment for your agents</p>
      </div>

      {error ? (
        <div className="text-xs text-error px-2 py-1 bg-error/10 rounded">
          Failed to load runtimes
        </div>
      ) : (
        <>
          <RuntimeTargetSelector
            runtimes={runtimes as RuntimeDescriptor[]}
            selectedId={selectedRuntime?.runtimeId}
            onSelect={onSelect}
            className="mb-4"
          />

          <RuntimeCapabilityMatrix
            runtimes={runtimes as RuntimeDescriptor[]}
            features={[
              "sandboxed",
              "fast-cold-start",
              "portable",
              "full-node-api",
              "native-modules",
              "npm-packages",
              "fast-startup",
              "bun-packages",
              "secure-by-default",
              "typescript-native",
              "web-standards",
              "incremental",
              "faas-optimized",
            ]}
            className="mb-4"
          />
        </>
      )}

      <div className="border-t border-border-subtle pt-4">
        <h4 className="text-xs font-medium mb-2">Available Capabilities</h4>
        <div className="grid grid-cols-2 gap-2">
          {CAPABILITIES.map((cap) => (
            <div
              key={cap.id}
              className="flex items-center gap-2 px-2 py-1.5 bg-bg-primary rounded-lg border border-border-subtle"
            >
              <cap.icon className={`size-3 ${cap.color}`} />
              <span className="text-[10px]">{cap.label}</span>
            </div>
          ))}
        </div>
      </div>

      <WasmExecutionPanel
        runtimeId={selectedRuntime?.runtimeId || "wasm"}
        status="idle"
        logs={[]}
        className="mb-4"
      />
    </div>
  );
}