import { useState } from "react";
import { GlassCard, Badge, Button, Spinner } from "@functionfly/ui-core";
import { Clock, RotateCcw, ChevronDown, ChevronUp, FileText, GitCompare, X, Check } from "lucide-react";
import { type Plugin, usePluginVersions, useRollbackPlugin } from "@/hooks/usePlugin";
import { useQueryClient } from "@tanstack/react-query";
import { pluginKeys } from "@/hooks/usePlugin";
import { formatDistanceToNow } from "date-fns";
import { cn } from "@/lib/utils";

interface PluginVersionManagerProps {
  plugins: Plugin[];
}

interface VersionCompare {
  version1: string;
  version2: string;
  changelog1?: string;
  changelog2?: string;
}

export function PluginVersionManager({ plugins }: PluginVersionManagerProps) {
  const [selectedPlugin, setSelectedPlugin] = useState<Plugin | null>(plugins[0] || null);
  const [compareMode, setCompareMode] = useState(false);
  const [compareVersions, setCompareVersions] = useState<VersionCompare[]>([]);
  const { data, isLoading } = usePluginVersions(selectedPlugin?.id || "");
  const rollbackMutation = useRollbackPlugin();
  const queryClient = useQueryClient();

  const handleRollback = async (version: string) => {
    if (!selectedPlugin) return;
    if (confirm(`Rollback to version ${version}? This will revert all changes made in versions after ${version}.`)) {
      await rollbackMutation.mutateAsync({ pluginId: selectedPlugin.id, toVersion: version });
      queryClient.invalidateQueries({ queryKey: pluginKeys.versions(selectedPlugin.id) });
    }
  };

  const toggleCompareVersion = (version: string) => {
    if (compareVersions.find((v) => v.version1 === version)) {
      setCompareVersions(compareVersions.filter((v) => v.version1 !== version));
    } else if (compareVersions.find((v) => v.version2 === version)) {
      setCompareVersions(compareVersions.filter((v) => v.version2 !== version));
    } else if (compareVersions.length === 0) {
      setCompareVersions([{ version1: version, version2: "", changelog1: data?.versions?.find((v) => v.version === version)?.changelog }]);
    } else if (compareVersions.length === 1 && compareVersions[0].version2 === "") {
      setCompareVersions([{ ...compareVersions[0], version2: version, changelog2: data?.versions?.find((v) => v.version === version)?.changelog }]);
    }
  };

  const isVersionSelected = (version: string) => {
    return compareVersions.some((v) => v.version1 === version || v.version2 === version);
  };

  const renderChangelog = (changelog?: string) => {
    if (!changelog) return <p className="text-xs text-white/40 italic">No changelog available for this version</p>;
    return (
      <div className="space-y-2">
        {changelog.split("\n").filter(line => line.trim()).map((line, i) => (
          <div key={i} className={cn(
            "text-sm",
            line.startsWith("## ") ? "text-white font-medium mt-3" :
            line.startsWith("- ") || line.startsWith("* ") ? "text-white/70 pl-3" :
            "text-white/60"
          )}>
            {line}
          </div>
        ))}
      </div>
    );
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Clock className="w-5 h-5 text-white/60" />
          <h3 className="text-sm font-medium text-white/60">Version History</h3>
        </div>
        <Button
          variant={compareMode ? "default" : "outline"}
          size="sm"
          onClick={() => { setCompareMode(!compareMode); setCompareVersions([]); }}
        >
          <GitCompare className="w-4 h-4 mr-2" />
          {compareMode ? "Exit Compare" : "Compare Versions"}
        </Button>
      </div>

      <div className="flex gap-4">
        <div className="w-64 space-y-2">
          <h3 className="text-sm font-medium text-white/60">Select Plugin</h3>
          {plugins.map((plugin) => (
            <button
              key={plugin.id}
              onClick={() => { setSelectedPlugin(plugin); setCompareVersions([]); }}
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
            <GlassCard className="p-4">
              {compareMode ? (
                <div className="space-y-4">
                  <div className="flex items-center gap-2 p-3 bg-orange-500/10 border border-orange-500/20 rounded-lg">
                    <GitCompare className="w-5 h-5 text-orange-400" />
                    <div>
                      <p className="text-sm text-white font-medium">Compare Mode</p>
                      <p className="text-xs text-white/60">Select two versions to compare their changelogs</p>
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div className="p-4 bg-white/5 rounded-lg border border-white/10">
                      <div className="flex items-center justify-between mb-3">
                        <span className="text-sm text-white/60">Version 1</span>
                        {compareVersions[0]?.version1 && (
                          <Badge className="bg-orange-500/20 text-orange-400 border-orange-500/30">
                            v{compareVersions[0].version1}
                          </Badge>
                        )}
                      </div>
                      {compareVersions.length === 0 || !compareVersions[0]?.version1 ? (
                        <p className="text-sm text-white/40 italic">Select first version</p>
                      ) : (
                        renderChangelog(compareVersions[0].changelog1)
                      )}
                    </div>

                    <div className="p-4 bg-white/5 rounded-lg border border-white/10">
                      <div className="flex items-center justify-between mb-3">
                        <span className="text-sm text-white/60">Version 2</span>
                        {compareVersions[0]?.version2 && (
                          <Badge className="bg-blue-500/20 text-blue-400 border-blue-500/30">
                            v{compareVersions[0].version2}
                          </Badge>
                        )}
                      </div>
                      {compareVersions.length === 0 || !compareVersions[0]?.version2 ? (
                        <p className="text-sm text-white/40 italic">Select second version</p>
                      ) : (
                        renderChangelog(compareVersions[0].changelog2)
                      )}
                    </div>
                  </div>
                </div>
              ) : (
                <div className="space-y-3">
                  {data?.versions && data.versions.length > 0 ? (
                    data.versions.map((version, index) => {
                      const isCurrent = index === 0 && version.version === selectedPlugin.version;
                      const isSelected = isVersionSelected(version.version);
                      return (
                        <div
                          key={version.id}
                          className={cn(
                            "flex items-center justify-between p-3 rounded-lg transition-colors",
                            isSelected ? "bg-orange-500/10 border border-orange-500/30" : "bg-white/5"
                          )}
                        >
                          <div className="flex items-center gap-3">
                            <button
                              onClick={() => toggleCompareVersion(version.version)}
                              className={cn(
                                "w-6 h-6 rounded border flex items-center justify-center transition-colors",
                                isSelected
                                  ? "bg-orange-500/20 border-orange-500/50 text-orange-400"
                                  : "border-white/20 text-white/40 hover:border-white/40 hover:text-white/60"
                              )}
                            >
                              {isSelected && <Check className="w-4 h-4" />}
                            </button>
                            <div className={`w-8 h-8 rounded-full flex items-center justify-center ${
                              isCurrent
                                ? "bg-green-500/20 text-green-400"
                                : "bg-white/10 text-white/40"
                            }`}>
                              {index + 1}
                            </div>
                            <div>
                              <div className="flex items-center gap-2">
                                <span className="font-medium text-white">v{version.version}</span>
                                {isCurrent && (
                                  <Badge className="bg-green-500/20 text-green-400 text-xs">Current</Badge>
                                )}
                              </div>
                              <p className="text-xs text-white/60">
                                {formatDistanceToNow(new Date(version.release_at), { addSuffix: true })}
                              </p>
                            </div>
                          </div>
                          <div className="flex items-center gap-2">
                            {version.changelog && (
                              <details className="group">
                                <summary className="cursor-pointer text-white/40 hover:text-white/60 text-xs flex items-center gap-1">
                                  <FileText className="w-3 h-3" />
                                  Changelog
                                </summary>
                                <div className="mt-2 p-3 bg-black/20 rounded border border-white/10 max-w-md">
                                  {renderChangelog(version.changelog)}
                                </div>
                              </details>
                            )}
                            {!isCurrent && (
                              <Button
                                size="sm"
                                variant="outline"
                                onClick={() => handleRollback(version.version)}
                                disabled={rollbackMutation.isPending}
                              >
                                <RotateCcw className="w-4 h-4 mr-1" />
                                Rollback
                              </Button>
                            )}
                          </div>
                        </div>
                      );
                    })
                  ) : (
                    <div className="flex items-center justify-center h-32 text-white/40">
                      No version history available
                    </div>
                  )}
                </div>
              )}
            </GlassCard>
          ) : (
            <div className="flex items-center justify-center h-64 text-white/40">
              Select a plugin to view version history
            </div>
          )}
        </div>
      </div>
    </div>
  );
}