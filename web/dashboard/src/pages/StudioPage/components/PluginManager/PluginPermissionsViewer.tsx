import { useState } from "react";
import { GlassCard, Badge, Button, Spinner } from "@functionfly/ui-core";
import { Shield, AlertTriangle, Check, X, Lock } from "lucide-react";
import { type Plugin, usePluginPermissions, useSetPermission } from "@/hooks/usePlugin";
import { useQueryClient } from "@tanstack/react-query";
import { pluginKeys } from "@/hooks/usePlugin";

interface PluginPermissionsViewerProps {
  plugins: Plugin[];
}

const permissionTypes = [
  { key: "network", label: "Network", description: "Access to network resources" },
  { key: "filesystem", label: "Filesystem", description: "Read/write access to filesystem" },
  { key: "agents", label: "Agents", description: "Create and manage AI agents" },
  { key: "memory", label: "Memory", description: "Access to memory systems" },
  { key: "terminal", label: "Terminal", description: "Execute terminal commands" },
  { key: "gpu", label: "GPU", description: "Access to GPU compute" },
  { key: "webhooks", label: "Webhooks", description: "Register webhook endpoints" },
  { key: "api_keys", label: "API Keys", description: "Read API keys" },
  { key: "secrets", label: "Secrets", description: "Access to secrets vault" },
];

export function PluginPermissionsViewer({ plugins }: PluginPermissionsViewerProps) {
  const [selectedPlugin, setSelectedPlugin] = useState<Plugin | null>(plugins[0] || null);
  const { data, isLoading } = usePluginPermissions(selectedPlugin?.id || "");
  const setPermissionMutation = useSetPermission();
  const queryClient = useQueryClient();

  const handleTogglePermission = async (permType: string, granted: boolean) => {
    if (!selectedPlugin) return;

    await setPermissionMutation.mutateAsync({
      pluginId: selectedPlugin.id,
      data: {
        permission_type: permType,
        permission_action: "access",
        granted,
      },
    });

    queryClient.invalidateQueries({ queryKey: pluginKeys.permissions(selectedPlugin.id) });
  };

  const getRiskLevel = (permType: string): "high" | "medium" | "low" => {
    const highRisk = ["terminal", "gpu", "api_keys", "secrets"];
    const mediumRisk = ["network", "filesystem", "agents"];
    if (highRisk.includes(permType)) return "high";
    if (mediumRisk.includes(permType)) return "medium";
    return "low";
  };

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
            <GlassCard className="p-4">
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-2">
                  <Shield className="w-5 h-5 text-white/60" />
                  <h3 className="font-medium text-white">Permissions</h3>
                </div>
                {isLoading && <Spinner className="w-4 h-4" />}
              </div>

              <div className="space-y-3">
                {permissionTypes.map((perm) => {
                  const granted = data?.permissions?.some(
                    (p) => p.permission_type === perm.key && p.granted
                  ) || false;
                  const risk = getRiskLevel(perm.key);

                  return (
                    <div
                      key={perm.key}
                      className="flex items-center justify-between p-3 bg-white/5 rounded-lg"
                    >
                      <div className="flex items-center gap-3">
                        <div className={`p-2 rounded-lg ${
                          risk === "high" ? "bg-red-500/20" :
                          risk === "medium" ? "bg-yellow-500/20" : "bg-green-500/20"
                        }`}>
                          {risk === "high" ? (
                            <AlertTriangle className="w-4 h-4 text-red-400" />
                          ) : (
                            <Lock className="w-4 h-4 text-green-400" />
                          )}
                        </div>
                        <div>
                          <div className="flex items-center gap-2">
                            <span className="font-medium text-white">{perm.label}</span>
                            <Badge
                              className={`text-xs ${
                                risk === "high" ? "bg-red-500/20 text-red-400" :
                                risk === "medium" ? "bg-yellow-500/20 text-yellow-400" :
                                "bg-green-500/20 text-green-400"
                              }`}
                            >
                              {risk}
                            </Badge>
                          </div>
                          <p className="text-xs text-white/60">{perm.description}</p>
                        </div>
                      </div>
                      <Button
                        size="sm"
                        variant={granted ? "default" : "outline"}
                        onClick={() => handleTogglePermission(perm.key, !granted)}
                      >
                        {granted ? (
                          <>
                            <Check className="w-4 h-4 mr-1" /> Granted
                          </>
                        ) : (
                          <>
                            <X className="w-4 h-4 mr-1" /> Denied
                          </>
                        )}
                      </Button>
                    </div>
                  );
                })}
              </div>
            </GlassCard>
          ) : (
            <div className="flex items-center justify-center h-64 text-white/40">
              Select a plugin to view permissions
            </div>
          )}
        </div>
      </div>
    </div>
  );
}