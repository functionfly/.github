import { GlassCard, Badge } from "@functionfly/ui-core";
import { Check, X, Minus } from "lucide-react";

interface CompatibilityMatrixProps {
  compatibility?: Record<string, unknown>;
  runtimeVersion?: string;
  platformVersion?: string;
}

interface Platform {
  id: string;
  name: string;
  icon: string;
  supported: boolean;
  minVersion?: string;
}

export function ExtensionCompatibilityMatrix({
  compatibility,
  runtimeVersion = "2.0+",
  platformVersion = "2024.1+",
}: CompatibilityMatrixProps) {
  const platforms: Platform[] = [
    { id: "wasm", name: "WebAssembly", icon: "W", supported: true, minVersion: "1.0" },
    { id: "nodejs", name: "Node.js", icon: "N", supported: true, minVersion: "16.0" },
    { id: "bun", name: "Bun", icon: "B", supported: true, minVersion: "1.0" },
    { id: "deno", name: "Deno", icon: "D", supported: true, minVersion: "1.28" },
    { id: "python", name: "Python", icon: "Py", supported: false },
    { id: "ruby", name: "Ruby", icon: "Rb", supported: false },
    { id: "go", name: "Go", icon: "Go", supported: false },
    { id: "rust", name: "Rust", icon: "Rs", supported: false },
  ];

  const features = [
    { id: "sandboxed", name: "Sandboxed Execution", description: "Runs in isolated environment" },
    { id: "network", name: "Network Access", description: "Can make external requests" },
    { id: "filesystem", name: "Filesystem Access", description: "Can read/write files" },
    { id: "envvars", name: "Environment Variables", description: "Can access environment" },
    { id: "webhooks", name: "Webhooks", description: "Supports webhook callbacks" },
    { id: "streaming", name: "Streaming", description: "Supports response streaming" },
    { id: "streaming", name: "Long-running", description: "Supports long-running tasks" },
    { id: "async", name: "Async/Await", description: "Supports async operations" },
  ];

  const supportedFeatures = compatibility?.features as string[] || ["sandboxed", "network", "webhooks", "async"];

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-white">Compatibility Matrix</h3>
        <Badge variant="outline" size="sm" className="text-white/60 border-white/20">
          Runtime {runtimeVersion}
        </Badge>
      </div>

      <GlassCard className="p-4">
        <div className="space-y-3">
          <div className="grid grid-cols-4 gap-2 text-[10px] text-white/40 uppercase tracking-wider pb-2 border-b border-white/10">
            <div>Platform</div>
            <div>Status</div>
            <div>Min Version</div>
            <div>Features</div>
          </div>
          {platforms.map((platform) => (
            <div key={platform.id} className="grid grid-cols-4 gap-2 items-center py-1">
              <div className="flex items-center gap-2">
                <div className="w-6 h-6 rounded bg-white/10 flex items-center justify-center text-[10px] font-bold text-white">
                  {platform.icon}
                </div>
                <span className="text-sm text-white">{platform.name}</span>
              </div>
              <div>
                {platform.supported ? (
                  <span className="inline-flex items-center gap-1 text-xs text-green-400">
                    <Check className="w-3 h-3" /> Supported
                  </span>
                ) : (
                  <span className="inline-flex items-center gap-1 text-xs text-white/40">
                    <X className="w-3 h-3" /> N/A
                  </span>
                )}
              </div>
              <div className="text-xs text-white/60">
                {platform.minVersion || "-"}
              </div>
              <div className="text-xs text-white/60">
                {platform.supported ? `${supportedFeatures.length}+` : "-"}
              </div>
            </div>
          ))}
        </div>
      </GlassCard>

      <GlassCard className="p-4">
        <h4 className="text-xs font-medium text-white/60 uppercase tracking-wider mb-3">Feature Support</h4>
        <div className="grid grid-cols-2 gap-2">
          {features.map((feature) => {
            const supported = supportedFeatures.includes(feature.id);
            return (
              <div
                key={feature.id}
                className="flex items-center gap-2 p-2 rounded bg-white/5"
              >
                <div className={`w-5 h-5 rounded flex items-center justify-center ${
                  supported ? "bg-green-500/20 text-green-400" : "bg-white/10 text-white/40"
                }`}>
                  {supported ? (
                    <Check className="w-3 h-3" />
                  ) : (
                    <Minus className="w-3 h-3" />
                  )}
                </div>
                <div>
                  <div className="text-xs font-medium text-white">{feature.name}</div>
                  <div className="text-[10px] text-white/60">{feature.description}</div>
                </div>
              </div>
            );
          })}
        </div>
      </GlassCard>

      <div className="flex items-center gap-2 text-xs text-white/60">
        <span className="w-2 h-2 rounded-full bg-green-400" />
        <span>Fully compatible</span>
        <span className="mx-2">•</span>
        <span className="w-2 h-2 rounded-full bg-yellow-400" />
        <span>Partially compatible</span>
        <span className="mx-2">•</span>
        <span className="w-2 h-2 rounded-full bg-red-400" />
        <span>Incompatible</span>
      </div>
    </div>
  );
}