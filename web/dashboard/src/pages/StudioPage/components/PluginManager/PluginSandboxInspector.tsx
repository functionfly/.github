import { useState } from "react";
import { GlassCard, Badge, Button, Spinner, Slider, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@functionfly/ui-core";
import { Activity, Cpu, MemoryStick, Clock, Globe, Shield, DollarSign, Zap, Save } from "lucide-react";
import { type Plugin, usePluginSandbox, useUpdateSandbox, type SandboxTier } from "@/hooks/usePlugin";
import { useQueryClient } from "@tanstack/react-query";
import { pluginKeys } from "@/hooks/usePlugin";
import { cn } from "@/lib/utils";

interface PluginSandboxInspectorProps {
  plugins: Plugin[];
}

type SandboxPreset = "development" | "production" | "enterprise" | "custom";

interface PresetConfig {
  label: string;
  tier: SandboxTier;
  cpu: number;
  memory: number;
  timeout: number;
  rateLimit: number;
  estimatedCost: number;
}

const tierInfo: Record<SandboxTier, { label: string; description: string; color: string }> = {
  wasm: { label: "WASM", description: "Fastest and safest - UI plugins and utilities", color: "bg-green-500/20 text-green-400" },
  worker: { label: "Worker", description: "Isolated worker - AI tools and graph nodes", color: "bg-blue-500/20 text-blue-400" },
  microvm: { label: "MicroVM", description: "Full isolation - untrusted third-party code", color: "bg-yellow-500/20 text-yellow-400" },
  enterprise: { label: "Enterprise", description: "Maximum isolation - regulated workloads", color: "bg-red-500/20 text-red-400" },
};

const PRESETS: Record<SandboxPreset, PresetConfig> = {
  development: {
    label: "Development",
    tier: "wasm",
    cpu: 0.5,
    memory: 128,
    timeout: 30,
    rateLimit: 100,
    estimatedCost: 0,
  },
  production: {
    label: "Production",
    tier: "worker",
    cpu: 2,
    memory: 512,
    timeout: 60,
    rateLimit: 500,
    estimatedCost: 15,
  },
  enterprise: {
    label: "Enterprise",
    tier: "microvm",
    cpu: 4,
    memory: 2048,
    timeout: 300,
    rateLimit: 1000,
    estimatedCost: 75,
  },
  custom: {
    label: "Custom",
    tier: "worker",
    cpu: 1,
    memory: 256,
    timeout: 30,
    rateLimit: 100,
    estimatedCost: 5,
  },
};

export function PluginSandboxInspector({ plugins }: PluginSandboxInspectorProps) {
  const [selectedPlugin, setSelectedPlugin] = useState<Plugin | null>(plugins[0] || null);
  const [activePreset, setActivePreset] = useState<SandboxPreset>("production");
  const { data: sandbox, isLoading } = usePluginSandbox(selectedPlugin?.id || "");
  const updateSandboxMutation = useUpdateSandbox();
  const queryClient = useQueryClient();

  const currentPreset = sandbox?.sandbox ? PRESETS.custom : PRESETS[activePreset];

  const estimateMonthlyCost = (tier: SandboxTier, cpu: number, memory: number, rateLimit: number) => {
    const baseCost: Record<SandboxTier, number> = {
      wasm: 0,
      worker: 5,
      microvm: 25,
      enterprise: 50,
    };
    const cpuCost = cpu * 3;
    const memoryCost = (memory / 256) * 2;
    const rateCost = (rateLimit / 100) * 1;
    return +(baseCost[tier] + cpuCost + memoryCost + rateCost).toFixed(2);
  };

  const handlePresetChange = async (preset: SandboxPreset) => {
    if (!selectedPlugin) return;
    setActivePreset(preset);
    const config = PRESETS[preset];
    await updateSandboxMutation.mutateAsync({
      pluginId: selectedPlugin.id,
      data: {
        tier: config.tier,
        cpu_limit: config.cpu,
        memory_limit_mb: config.memory,
        timeout_seconds: config.timeout,
        rate_limit_rpm: config.rateLimit,
      },
    });
    queryClient.invalidateQueries({ queryKey: pluginKeys.sandbox(selectedPlugin.id) });
  };

  const handleUpdateTier = async (tier: SandboxTier) => {
    if (!selectedPlugin) return;
    setActivePreset("custom");
    await updateSandboxMutation.mutateAsync({
      pluginId: selectedPlugin.id,
      data: { tier },
    });
    queryClient.invalidateQueries({ queryKey: pluginKeys.sandbox(selectedPlugin.id) });
  };

  const handleUpdateResource = async (key: string, value: number) => {
    if (!selectedPlugin) return;
    setActivePreset("custom");
    const updateData: any = { [key]: value };
    await updateSandboxMutation.mutateAsync({
      pluginId: selectedPlugin.id,
      data: updateData,
    });
    queryClient.invalidateQueries({ queryKey: pluginKeys.sandbox(selectedPlugin.id) });
  };

  const getEffectiveValue = (key: keyof typeof PRESETS.production) => {
    return sandbox?.sandbox?.[key] ?? PRESETS[activePreset][key];
  };

  const estimatedCost = sandbox?.sandbox
    ? estimateMonthlyCost(
        sandbox.sandbox.tier as SandboxTier,
        sandbox.sandbox.cpu_limit,
        sandbox.sandbox.memory_limit_mb,
        sandbox.sandbox.rate_limit_rpm || 100
      )
    : estimateMonthlyCost(
        getEffectiveValue("tier") as SandboxTier,
        getEffectiveValue("cpu"),
        getEffectiveValue("memory"),
        getEffectiveValue("rateLimit")
      );

  return (
    <div className="space-y-4">
      <div className="flex gap-4">
        <div className="w-64 space-y-2">
          <h3 className="text-sm font-medium text-white/60">Select Plugin</h3>
          {plugins.map((plugin) => (
            <button
              key={plugin.id}
              onClick={() => setSelectedPlugin(plugin)}
              className={`w-full text-left p-3 rounded-lg border transition-colors ${
                selectedPlugin?.id === plugin.id
                  ? "bg-white/10 border-white/20"
                  : "bg-white/5 border-white/10 hover:bg-white/10"
              }`}
            >
              <div className="font-medium text-white text-sm">{plugin.name}</div>
              <div className="text-xs text-white/60">v{plugin.version}</div>
            </button>
          ))}
        </div>

        <div className="flex-1">
          {selectedPlugin ? (
            <GlassCard className="p-4 space-y-4">
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-2">
                  <Shield className="w-5 h-5 text-white/60" />
                  <h3 className="font-medium text-white">Sandbox Configuration</h3>
                </div>
                {isLoading && <Spinner className="w-4 h-4" />}
              </div>

              <div className="p-4 bg-gradient-to-r from-orange-500/10 to-red-500/10 border border-orange-500/20 rounded-lg">
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-2">
                    <DollarSign className="w-5 h-5 text-orange-400" />
                    <span className="text-sm text-white font-medium">Estimated Monthly Cost</span>
                  </div>
                  <span className="text-2xl font-bold text-white">${estimatedCost}</span>
                </div>
                <p className="text-xs text-white/60">
                  Based on current resource configuration. Actual usage may vary.
                </p>
              </div>

              <div>
                <label className="text-sm text-white/60 mb-3 block">Quick Presets</label>
                <div className="grid grid-cols-3 gap-2">
                  {(Object.keys(PRESETS) as SandboxPreset[]).filter(p => p !== "custom").map((preset) => (
                    <button
                      key={preset}
                      onClick={() => handlePresetChange(preset)}
                      className={cn(
                        "p-3 rounded-lg border text-left transition-all",
                        activePreset === preset
                          ? "bg-white/10 border-white/30"
                          : "bg-white/5 border-white/10 hover:bg-white/10"
                      )}
                    >
                      <div className="text-sm font-medium text-white">{PRESETS[preset].label}</div>
                      <div className="text-xs text-white/60 mt-1">
                        ${PRESETS[preset].estimatedCost}/mo
                      </div>
                      <div className="flex items-center gap-1 mt-2">
                        <Badge className={cn("text-xs", tierInfo[PRESETS[preset].tier].color)}>
                          {PRESETS[preset].tier}
                        </Badge>
                      </div>
                    </button>
                  ))}
                </div>
              </div>

              <div>
                <label className="text-sm text-white/60 mb-2 block">Sandbox Tier</label>
                <Select
                  value={sandbox?.sandbox?.tier || activePreset}
                  onValueChange={(v) => handleUpdateTier(v as SandboxTier)}
                >
                  <SelectTrigger className="bg-white/5 border-white/10">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {Object.entries(tierInfo).map(([tier, info]) => (
                      <SelectItem key={tier} value={tier}>
                        <div className="flex items-center gap-2">
                          <span className={cn("px-2 py-0.5 rounded text-xs", info.color)}>{info.label}</span>
                          <span className="text-white/60 text-xs">{info.description}</span>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                {sandbox?.sandbox?.tier && (
                  <p className="mt-2 text-xs text-white/40">
                    {tierInfo[sandbox.sandbox.tier as SandboxTier]?.description}
                  </p>
                )}
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <label className="text-sm text-white/60 flex items-center gap-1">
                      <Cpu className="w-4 h-4" /> CPU Limit
                    </label>
                    <span className="text-sm text-white">{getEffectiveValue("cpu")} cores</span>
                  </div>
                  <Slider
                    value={[getEffectiveValue("cpu")]}
                    min={0.1}
                    max={4}
                    step={0.1}
                    onValueChange={([v]) => handleUpdateResource("cpu_limit", v)}
                    className="bg-white/10"
                  />
                </div>

                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <label className="text-sm text-white/60 flex items-center gap-1">
                      <MemoryStick className="w-4 h-4" /> Memory Limit
                    </label>
                    <span className="text-sm text-white">{getEffectiveValue("memory")} MB</span>
                  </div>
                  <Slider
                    value={[getEffectiveValue("memory")]}
                    min={64}
                    max={2048}
                    step={64}
                    onValueChange={([v]) => handleUpdateResource("memory_limit_mb", v)}
                    className="bg-white/10"
                  />
                </div>

                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <label className="text-sm text-white/60 flex items-center gap-1">
                      <Clock className="w-4 h-4" /> Timeout
                    </label>
                    <span className="text-sm text-white">{getEffectiveValue("timeout")}s</span>
                  </div>
                  <Slider
                    value={[getEffectiveValue("timeout")]}
                    min={5}
                    max={300}
                    step={5}
                    onValueChange={([v]) => handleUpdateResource("timeout_seconds", v)}
                    className="bg-white/10"
                  />
                </div>

                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <label className="text-sm text-white/60 flex items-center gap-1">
                      <Activity className="w-4 h-4" /> Rate Limit
                    </label>
                    <span className="text-sm text-white">{getEffectiveValue("rateLimit")} rpm</span>
                  </div>
                  <Slider
                    value={[getEffectiveValue("rateLimit")]}
                    min={10}
                    max={1000}
                    step={10}
                    onValueChange={([v]) => handleUpdateResource("rate_limit_rpm", v)}
                    className="bg-white/10"
                  />
                </div>
              </div>

              <div>
                <label className="text-sm text-white/60 mb-2 block flex items-center gap-1">
                  <Globe className="w-4 h-4" /> Allowed Domains
                </label>
                <div className="p-3 bg-white/5 rounded-lg min-h-[60px]">
                  {sandbox?.sandbox?.allowed_domains?.length ? (
                    <div className="flex flex-wrap gap-2">
                      {sandbox.sandbox.allowed_domains.map((domain, i) => (
                        <Badge key={i} variant="outline" className="text-white/60 border-white/20">
                          {domain}
                        </Badge>
                      ))}
                    </div>
                  ) : (
                    <p className="text-xs text-white/40">No domain restrictions</p>
                  )}
                </div>
              </div>
            </GlassCard>
          ) : (
            <div className="flex items-center justify-center h-64 text-white/40">
              Select a plugin to configure sandbox
            </div>
          )}
        </div>
      </div>
    </div>
  );
}