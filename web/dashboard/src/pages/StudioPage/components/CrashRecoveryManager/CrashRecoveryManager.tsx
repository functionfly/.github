import { useState } from "react";
import { GlassCard, Badge, Button } from "@functionfly/ui-core";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";
import {
  Shield, AlertTriangle, CheckCircle2, RotateCcw, Clock, FileWarning,
  Save, Trash2, RefreshCw, Info, XCircle
} from "lucide-react";

interface RecoveryPoint {
  id: string;
  timestamp: Date;
  type: "auto" | "manual" | "crash";
  description: string;
  status: "available" | "restored" | "corrupted" | "archived";
  size: string;
}

interface CrashReport {
  id: string;
  timestamp: Date;
  type: "crash" | "freeze" | "memory" | "disk";
  message: string;
  stackTrace?: string;
  resolved: boolean;
}

const mockRecoveryPoints: RecoveryPoint[] = [
  {
    id: "rp-1",
    timestamp: new Date(Date.now() - 3600000),
    type: "auto",
    description: "Auto-save recovery point",
    status: "available",
    size: "12.4 MB",
  },
  {
    id: "rp-2",
    timestamp: new Date(Date.now() - 7200000),
    type: "manual",
    description: "Before plugin installation",
    status: "available",
    size: "18.2 MB",
  },
  {
    id: "rp-3",
    timestamp: new Date(Date.now() - 86400000),
    type: "crash",
    description: "Crash recovery - unexpected shutdown",
    status: "restored",
    size: "15.7 MB",
  },
  {
    id: "rp-4",
    timestamp: new Date(Date.now() - 172800000),
    type: "auto",
    description: "Scheduled auto-save",
    status: "archived",
    size: "14.1 MB",
  },
];

const mockCrashReports: CrashReport[] = [
  {
    id: "cr-1",
    timestamp: new Date(Date.now() - 86400000 * 2),
    type: "freeze",
    message: "UI became unresponsive during graph rendering",
    resolved: true,
  },
  {
    id: "cr-2",
    timestamp: new Date(Date.now() - 86400000 * 5),
    type: "memory",
    message: "Memory usage exceeded threshold",
    resolved: true,
  },
  {
    id: "cr-3",
    timestamp: new Date(Date.now() - 86400000 * 7),
    type: "crash",
    message: "Plugin caused segmentation fault",
    stackTrace: "at PluginHost.processMessage (plugin-host.ts:342)\nat NodeGraph.render (renderer.ts:156)",
    resolved: true,
  },
];

export function CrashRecoveryManager() {
  const [activeTab, setActiveTab] = useState("recovery");
  const [recoveryPoints] = useState<RecoveryPoint[]>(mockRecoveryPoints);
  const [crashReports] = useState<CrashReport[]>(mockCrashReports);
  const [isRecovering, setIsRecovering] = useState(false);

  const handleRestore = async (pointId: string) => {
    setIsRecovering(true);
    await new Promise((r) => setTimeout(r, 2000));
    setIsRecovering(false);
  };

  const formatTimestamp = (date: Date) => {
    return date.toLocaleString("en-US", {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const getStatusIcon = (status: RecoveryPoint["status"]) => {
    switch (status) {
      case "available":
        return <CheckCircle2 className="w-4 h-4 text-emerald-400" />;
      case "restored":
        return <RefreshCw className="w-4 h-4 text-blue-400" />;
      case "corrupted":
        return <XCircle className="w-4 h-4 text-red-400" />;
      case "archived":
        return <FileWarning className="w-4 h-4 text-yellow-400" />;
    }
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between p-5 border-b border-white/10">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-emerald-500/20 flex items-center justify-center">
            <Shield className="w-5 h-5 text-emerald-400" />
          </div>
          <div>
            <h2 className="text-xl font-semibold text-white">Crash Recovery</h2>
            <p className="text-sm text-white/60">Protect your work from unexpected failures</p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <Badge variant="outline" className="text-white/60 border-white/20">
            <CheckCircle2 className="w-3 h-3 mr-1" />
            Protection Active
          </Badge>
        </div>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col">
        <div className="px-5 pt-5">
          <TabsList className="inline-flex h-auto flex-wrap gap-1 rounded-xl border border-white/10 bg-white/5 p-1.5 text-white/60">
            <TabsTrigger
              value="recovery"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <RotateCcw className="h-4 w-4 shrink-0" />
              Recovery Points
            </TabsTrigger>
            <TabsTrigger
              value="crashes"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <AlertTriangle className="h-4 w-4 shrink-0" />
              Crash Reports
            </TabsTrigger>
            <TabsTrigger
              value="settings"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <Shield className="h-4 w-4 shrink-0" />
              Protection Settings
            </TabsTrigger>
          </TabsList>
        </div>

        <div className="flex-1 overflow-auto p-4">
          <TabsContent value="recovery" className="mt-0">
            <div className="space-y-4">
              <div className="flex items-center justify-between mb-4">
                <p className="text-sm text-white/60">
                  {recoveryPoints.filter((r) => r.status === "available").length} recovery points available
                </p>
                <Button size="sm" variant="outline" className="gap-2">
                  <Save className="w-4 h-4" />
                  Create Recovery Point
                </Button>
              </div>

              <div className="space-y-3">
                {recoveryPoints.map((point) => (
                  <GlassCard key={point.id} className="p-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-4">
                        <div
                          className={cn(
                            "w-10 h-10 rounded-xl flex items-center justify-center",
                            point.type === "auto" && "bg-blue-500/20 text-blue-400",
                            point.type === "manual" && "bg-orange-500/20 text-orange-400",
                            point.type === "crash" && "bg-red-500/20 text-red-400"
                          )}
                        >
                          {point.type === "auto" && <Clock className="w-5 h-5" />}
                          {point.type === "manual" && <Save className="w-5 h-5" />}
                          {point.type === "crash" && <AlertTriangle className="w-5 h-5" />}
                        </div>
                        <div>
                          <div className="flex items-center gap-2 mb-1">
                            <h3 className="font-medium text-white">{point.description}</h3>
                            {getStatusIcon(point.status)}
                          </div>
                          <div className="flex items-center gap-3 text-xs text-white/50">
                            <span>{formatTimestamp(point.timestamp)}</span>
                            <span>•</span>
                            <span>{point.size}</span>
                            <Badge
                              variant="outline"
                              className={cn(
                                "text-[10px]",
                                point.type === "auto" && "text-blue-400 border-blue-400/30",
                                point.type === "manual" && "text-orange-400 border-orange-400/30",
                                point.type === "crash" && "text-red-400 border-red-400/30"
                              )}
                            >
                              {point.type}
                            </Badge>
                          </div>
                        </div>
                      </div>
                      <div className="flex items-center gap-2">
                        {point.status === "available" && (
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => handleRestore(point.id)}
                            disabled={isRecovering}
                            className="gap-1"
                          >
                            <RotateCcw className="w-3 h-3" />
                            Restore
                          </Button>
                        )}
                      </div>
                    </div>
                  </GlassCard>
                ))}
              </div>
            </div>
          </TabsContent>

          <TabsContent value="crashes" className="mt-0">
            <div className="space-y-4">
              {crashReports.length === 0 ? (
                <GlassCard className="flex flex-col items-center justify-center h-48">
                  <CheckCircle2 className="w-10 h-10 text-emerald-400 mb-3" />
                  <p className="text-white/80 font-medium">No crash reports</p>
                  <p className="text-sm text-white/60">Your studio is running smoothly</p>
                </GlassCard>
              ) : (
                crashReports.map((report) => (
                  <GlassCard key={report.id} className="p-4">
                    <div className="flex items-start justify-between">
                      <div className="flex items-start gap-4">
                        <div
                          className={cn(
                            "w-10 h-10 rounded-xl flex items-center justify-center shrink-0",
                            report.resolved ? "bg-emerald-500/20 text-emerald-400" : "bg-red-500/20 text-red-400"
                          )}
                        >
                          {report.resolved ? <CheckCircle2 className="w-5 h-5" /> : <AlertTriangle className="w-5 h-5" />}
                        </div>
                        <div>
                          <div className="flex items-center gap-2 mb-1">
                            <h3 className="font-medium text-white">{report.message}</h3>
                            {report.resolved && (
                              <Badge className="text-[10px] bg-emerald-500/20 text-emerald-400 border-emerald-500/30">
                                Resolved
                              </Badge>
                            )}
                          </div>
                          <p className="text-sm text-white/60 mb-2">{formatTimestamp(report.timestamp)}</p>
                          {report.stackTrace && (
                            <div className="p-3 rounded-lg bg-black/30 border border-white/10">
                              <p className="text-xs font-mono text-white/70 overflow-x-auto">{report.stackTrace}</p>
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                  </GlassCard>
                ))
              )}
            </div>
          </TabsContent>

          <TabsContent value="settings" className="mt-0">
            <div className="space-y-4 max-w-xl">
              <GlassCard className="p-5">
                <h3 className="font-semibold text-white mb-4">Recovery Settings</h3>
                <div className="space-y-4">
                  <div className="flex items-center justify-between p-4 rounded-lg bg-white/5">
                    <div className="flex items-center gap-3">
                      <Clock className="w-5 h-5 text-blue-400" />
                      <div>
                        <p className="text-sm font-medium text-white">Auto-save interval</p>
                        <p className="text-xs text-white/60">Create recovery points automatically</p>
                      </div>
                    </div>
                    <Badge className="bg-blue-500/20 text-blue-400 border-blue-500/30">5 minutes</Badge>
                  </div>
                  <div className="flex items-center justify-between p-4 rounded-lg bg-white/5">
                    <div className="flex items-center gap-3">
                      <Save className="w-5 h-5 text-orange-400" />
                      <div>
                        <p className="text-sm font-medium text-white">Max recovery points</p>
                        <p className="text-xs text-white/60">Keep recent recovery points</p>
                      </div>
                    </div>
                    <Badge className="bg-orange-500/20 text-orange-400 border-orange-500/30">10</Badge>
                  </div>
                  <div className="flex items-center justify-between p-4 rounded-lg bg-white/5">
                    <div className="flex items-center gap-3">
                      <Trash2 className="w-5 h-5 text-red-400" />
                      <div>
                        <p className="text-sm font-medium text-white">Auto-cleanup</p>
                        <p className="text-xs text-white/60">Remove old recovery points</p>
                      </div>
                    </div>
                    <Badge className="bg-emerald-500/20 text-emerald-400 border-emerald-500/30">Enabled</Badge>
                  </div>
                </div>
              </GlassCard>

              <GlassCard className="p-5">
                <div className="flex items-center gap-3 mb-4">
                  <Info className="w-5 h-5 text-blue-400" />
                  <h3 className="font-semibold text-white">About Crash Recovery</h3>
                </div>
                <p className="text-sm text-white/60 leading-relaxed">
                  Crash Recovery automatically saves your workspace state at regular intervals.
                  If Studio crashes or closes unexpectedly, you can restore your work from the
                  last available recovery point. Recovery points are stored locally and encrypted.
                </p>
              </GlassCard>
            </div>
          </TabsContent>
        </div>
      </Tabs>
    </div>
  );
}