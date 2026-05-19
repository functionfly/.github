import { useState } from "react";
import { GlassCard, Badge, Button } from "@functionfly/ui-core";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Progress } from "@/components/ui/progress";
import { cn } from "@/lib/utils";
import {
  FlaskConical, Sparkles, Beaker, Zap, Rocket, Eye, EyeOff,
  Clock, RotateCcw, Check, X, AlertTriangle, Star, GitBranch
} from "lucide-react";

interface ExperimentalFeature {
  id: string;
  name: string;
  description: string;
  category: "ai" | "ui" | "performance" | "workflow";
  status: "enabled" | "disabled" | "beta" | "new";
  impact: "low" | "medium" | "high";
  flag: string;
  sinceVersion?: string;
}

const experimentalFeatures: ExperimentalFeature[] = [
  {
    id: "ef-1",
    name: "Neural Graph Completion",
    description: "AI-powered automatic node connection suggestions using local language model",
    category: "ai",
    status: "beta",
    impact: "medium",
    flag: "FF_NEURAL_COMPLETION",
    sinceVersion: "2.4.0",
  },
  {
    id: "ef-2",
    name: "Real-time Collaboration",
    description: "Multi-user editing with presence indicators and live cursors",
    category: "workflow",
    status: "beta",
    impact: "high",
    flag: "FF_COLLAB",
  },
  {
    id: "ef-3",
    name: "WebGPU Renderer",
    description: "Hardware-accelerated graph rendering using WebGPU API",
    category: "performance",
    status: "new",
    impact: "high",
    flag: "FF_WEBGPU",
    sinceVersion: "2.5.0",
  },
  {
    id: "ef-4",
    name: "Fluid Animation Mode",
    description: "Enhanced spring-based animations for node transitions",
    category: "ui",
    status: "enabled",
    impact: "low",
    flag: "FF_FLUID_ANIM",
  },
  {
    id: "ef-5",
    name: "Streaming Execution",
    description: "Live output streaming for long-running graph executions",
    category: "workflow",
    status: "beta",
    impact: "medium",
    flag: "FF_STREAM_EXEC",
  },
  {
    id: "ef-6",
    name: "Adaptive Node Sizing",
    description: "Automatically resize nodes based on content",
    category: "ui",
    status: "disabled",
    impact: "low",
    flag: "FF_ADAPTIVE_NODES",
  },
  {
    id: "ef-7",
    name: "Incremental Compilation",
    description: "Faster graph updates by only recompiling changed sections",
    category: "performance",
    status: "enabled",
    impact: "medium",
    flag: "FF_INCREMENTAL_COMP",
  },
  {
    id: "ef-8",
    name: "Voice Control",
    description: "Control graphs using voice commands",
    category: "ai",
    status: "new",
    impact: "low",
    flag: "FF_VOICE_CONTROL",
    sinceVersion: "2.5.0",
  },
];

export function ExperimentalFeatureLab() {
  const [activeTab, setActiveTab] = useState("features");
  const [features, setFeatures] = useState<ExperimentalFeature[]>(experimentalFeatures);
  const [showOnlyEnabled, setShowOnlyEnabled] = useState(false);

  const toggleFeature = (id: string) => {
    setFeatures((prev) =>
      prev.map((f) =>
        f.id === id
          ? { ...f, status: f.status === "enabled" ? "disabled" : "enabled" }
          : f
      )
    );
  };

  const enabledCount = features.filter((f) => f.status === "enabled").length;
  const betaCount = features.filter((f) => f.status === "beta").length;
  const newCount = features.filter((f) => f.status === "new").length;

  const filteredFeatures = showOnlyEnabled
    ? features.filter((f) => f.status === "enabled")
    : features;

  const getCategoryIcon = (category: ExperimentalFeature["category"]) => {
    switch (category) {
      case "ai":
        return <Sparkles className="w-4 h-4 text-purple-400" />;
      case "ui":
        return <Beaker className="w-4 h-4 text-blue-400" />;
      case "performance":
        return <Zap className="w-4 h-4 text-yellow-400" />;
      case "workflow":
        return <GitBranch className="w-4 h-4 text-emerald-400" />;
    }
  };

  const getStatusBadge = (status: ExperimentalFeature["status"]) => {
    switch (status) {
      case "enabled":
        return <Badge className="text-[10px] bg-emerald-500/20 text-emerald-400 border-emerald-500/30">Enabled</Badge>;
      case "disabled":
        return <Badge variant="outline" className="text-[10px] text-white/40">Disabled</Badge>;
      case "beta":
        return <Badge className="text-[10px] bg-orange-500/20 text-orange-400 border-orange-500/30">Beta</Badge>;
      case "new":
        return <Badge className="text-[10px] bg-blue-500/20 text-blue-400 border-blue-500/30">New</Badge>;
    }
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between p-5 border-b border-white/10">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-purple-500/20 flex items-center justify-center">
            <FlaskConical className="w-5 h-5 text-purple-400" />
          </div>
          <div>
            <h2 className="text-xl font-semibold text-white">Experimental Feature Lab</h2>
            <p className="text-sm text-white/60">Try out cutting-edge features before release</p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <Badge variant="outline" className="text-emerald-400 border-emerald-400/30">
            <Check className="w-3 h-3 mr-1" />
            {enabledCount} enabled
          </Badge>
        </div>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col">
        <div className="px-5 pt-5">
          <TabsList className="inline-flex h-auto flex-wrap gap-1 rounded-xl border border-white/10 bg-white/5 p-1.5 text-white/60">
            <TabsTrigger
              value="features"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <FlaskConical className="h-4 w-4 shrink-0" />
              Features
            </TabsTrigger>
            <TabsTrigger
              value="stats"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <Rocket className="h-4 w-4 shrink-0" />
              Usage Stats
            </TabsTrigger>
            <TabsTrigger
              value="flags"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <Zap className="h-4 w-4 shrink-0" />
              Feature Flags
            </TabsTrigger>
          </TabsList>
        </div>

        <div className="flex-1 overflow-auto p-4">
          <TabsContent value="features" className="mt-0">
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <button
                    onClick={() => setShowOnlyEnabled(!showOnlyEnabled)}
                    className={cn(
                      "flex items-center gap-2 px-3 py-1.5 rounded-lg border transition-colors text-sm",
                      showOnlyEnabled
                        ? "bg-emerald-500/20 border-emerald-500/30 text-emerald-400"
                        : "bg-white/5 border-white/10 text-white/60 hover:text-white"
                    )}
                  >
                    {showOnlyEnabled ? <Eye className="w-4 h-4" /> : <EyeOff className="w-4 h-4" />}
                    {showOnlyEnabled ? "Showing enabled" : "Show all"}
                  </button>
                  <div className="flex items-center gap-2 text-xs text-white/50">
                    <span className="flex items-center gap-1">
                      <Star className="w-3 h-3 text-blue-400" />
                      {newCount} new
                    </span>
                    <span className="flex items-center gap-1">
                      <Sparkles className="w-3 h-3 text-orange-400" />
                      {betaCount} beta
                    </span>
                  </div>
                </div>
              </div>

              <div className="grid gap-3">
                {filteredFeatures.map((feature) => (
                  <GlassCard key={feature.id} className="p-4">
                    <div className="flex items-start justify-between gap-4">
                      <div className="flex items-start gap-4 flex-1 min-w-0">
                        <div className="w-10 h-10 rounded-xl bg-white/5 flex items-center justify-center shrink-0">
                          {getCategoryIcon(feature.category)}
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2 mb-1 flex-wrap">
                            <h3 className="font-medium text-white">{feature.name}</h3>
                            {getStatusBadge(feature.status)}
                            <Badge
                              variant="outline"
                              className={cn(
                                "text-[10px]",
                                feature.impact === "low" && "text-white/40 border-white/20",
                                feature.impact === "medium" && "text-yellow-400 border-yellow-400/30",
                                feature.impact === "high" && "text-red-400 border-red-400/30"
                              )}
                            >
                              {feature.impact} impact
                            </Badge>
                            {feature.sinceVersion && (
                              <Badge variant="outline" className="text-[10px] text-white/40 border-white/20">
                                v{feature.sinceVersion}
                              </Badge>
                            )}
                          </div>
                          <p className="text-sm text-white/60 line-clamp-2">{feature.description}</p>
                          <p className="text-xs text-white/40 font-mono mt-2">
                            {feature.flag}
                          </p>
                        </div>
                      </div>
                      <div className="shrink-0">
                        <Switch
                          checked={feature.status === "enabled"}
                          onCheckedChange={() => toggleFeature(feature.id)}
                        />
                      </div>
                    </div>
                  </GlassCard>
                ))}
              </div>
            </div>
          </TabsContent>

          <TabsContent value="stats" className="mt-0">
            <div className="space-y-6">
              <div className="grid grid-cols-3 gap-4">
                <GlassCard className="p-4 text-center">
                  <p className="text-3xl font-bold text-white">{enabledCount}</p>
                  <p className="text-sm text-white/60">Features Enabled</p>
                </GlassCard>
                <GlassCard className="p-4 text-center">
                  <p className="text-3xl font-bold text-orange-400">{betaCount}</p>
                  <p className="text-sm text-white/60">In Beta</p>
                </GlassCard>
                <GlassCard className="p-4 text-center">
                  <p className="text-3xl font-bold text-blue-400">{newCount}</p>
                  <p className="text-sm text-white/60">New Features</p>
                </GlassCard>
              </div>

              <GlassCard className="p-5">
                <h3 className="font-semibold text-white mb-4">Feature Adoption</h3>
                <div className="space-y-3">
                  {features.filter(f => f.status === "enabled").slice(0, 5).map((feature) => (
                    <div key={feature.id} className="flex items-center gap-4">
                      <span className="text-sm text-white/80 w-48 truncate">{feature.name}</span>
                      <Progress value={Math.random() * 100} className="flex-1 h-2" />
                      <span className="text-xs text-white/60 w-12 text-right">
                        {Math.floor(Math.random() * 100)}%
                      </span>
                    </div>
                  ))}
                </div>
              </GlassCard>
            </div>
          </TabsContent>

          <TabsContent value="flags" className="mt-0">
            <div className="space-y-4">
              <GlassCard className="p-5">
                <h3 className="font-semibold text-white mb-4">Feature Flag Reference</h3>
                <div className="space-y-2 font-mono text-sm">
                  {features.map((feature) => (
                    <div
                      key={feature.id}
                      className="flex items-center justify-between p-3 rounded-lg bg-white/5"
                    >
                      <code className="text-orange-400">{feature.flag}</code>
                      <span className="text-white/60">{feature.name}</span>
                    </div>
                  ))}
                </div>
              </GlassCard>

              <GlassCard className="p-5">
                <div className="flex items-center gap-2 mb-3">
                  <AlertTriangle className="w-5 h-5 text-yellow-400" />
                  <h3 className="font-semibold text-white">Warning</h3>
                </div>
                <p className="text-sm text-white/60">
                  Experimental features may be unstable or removed in future releases.
                  Do not enable features in production environments unless explicitly advised.
                </p>
              </GlassCard>
            </div>
          </TabsContent>
        </div>
      </Tabs>
    </div>
  );
}