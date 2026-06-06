import { Badge, Button, DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from "@functionfly/ui-core";
import { MoreHorizontal, Play, Pause, Trash2, Settings, Shield, Activity, Clock, AlertTriangle } from "lucide-react";
import { type Plugin, useEnablePlugin, useDisablePlugin, usePausePlugin, useUninstallPlugin, useSetPluginError } from "@/hooks/usePlugin";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

interface PluginCardProps {
  plugin: Plugin;
  onConfigure?: (plugin: Plugin) => void;
  onViewPermissions?: (plugin: Plugin) => void;
  onViewSandbox?: (plugin: Plugin) => void;
}

export function PluginCard({ plugin, onConfigure, onViewPermissions, onViewSandbox }: PluginCardProps) {
  const enableMutation = useEnablePlugin();
  const disableMutation = useDisablePlugin();
  const pauseMutation = usePausePlugin();
  const uninstallMutation = useUninstallPlugin();
  const setErrorMutation = useSetPluginError();

  const statusColors = {
    enabled: "bg-green-500/20 text-green-400 border-green-500/30",
    disabled: "bg-white/10 text-white/60 border-white/20",
    error: "bg-red-500/20 text-red-400 border-red-500/30",
    paused: "bg-yellow-500/20 text-yellow-400 border-yellow-500/30",
  };

  const typeColors = {
    ui: "bg-blue-500/20 text-blue-400 border-blue-500/30",
    graph: "bg-purple-500/20 text-purple-400 border-purple-500/30",
    ai_tool: "bg-orange-500/20 text-orange-400 border-orange-500/30",
    runtime: "bg-cyan-500/20 text-cyan-400 border-cyan-500/30",
    infrastructure: "bg-pink-500/20 text-pink-400 border-pink-500/30",
    marketplace: "bg-emerald-500/20 text-emerald-400 border-emerald-500/30",
  };

  const handleEnable = () => enableMutation.mutate(plugin.id);
  const handleDisable = () => disableMutation.mutate(plugin.id);
  const handlePause = () => pauseMutation.mutate(plugin.id);
  const handleUninstall = () => {
    if (confirm(`Are you sure you want to uninstall "${plugin.name}"?`)) {
      uninstallMutation.mutate(plugin.id);
    }
  };

  const handleReportError = () => {
    const errorMsg = prompt(`Report an error for "${plugin.name}":`);
    if (errorMsg && errorMsg.trim()) {
      setErrorMutation.mutate({ pluginId: plugin.id, error: errorMsg.trim() });
    }
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i];
  };

  return (
    <div className="bg-white/5 border border-white/10 rounded-lg p-4 hover:bg-white/10 transition-colors">
      <div className="flex items-start justify-between gap-2">
        <div className="flex items-center gap-3 min-w-0">
          {plugin.icon_url ? (
            <img src={plugin.icon_url} alt={plugin.name} className="w-10 h-10 rounded-lg object-cover shrink-0" />
          ) : (
            <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-purple-500/30 to-blue-500/30 flex items-center justify-center text-lg font-bold text-white/80 shrink-0">
              {plugin.name.charAt(0).toUpperCase()}
            </div>
          )}
          <div className="min-w-0">
            <h3 className="font-medium text-white truncate">{plugin.name}</h3>
            <p className="text-sm text-white/60">v{plugin.version}</p>
          </div>
        </div>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button className="p-1 rounded text-white/60 hover:text-white hover:bg-white/10">
              <MoreHorizontal className="w-4 h-4" />
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="bg-black/90 border-white/10">
            {plugin.status !== "enabled" && (
              <DropdownMenuItem onClick={handleEnable}>
                <Play className="w-4 h-4 mr-2" /> Enable
              </DropdownMenuItem>
            )}
            {plugin.status === "enabled" && (
              <DropdownMenuItem onClick={handleDisable}>
                <Pause className="w-4 h-4 mr-2" /> Disable
              </DropdownMenuItem>
            )}
            {plugin.status === "enabled" && (
              <DropdownMenuItem onClick={handlePause}>
                <Clock className="w-4 h-4 mr-2" /> Pause
              </DropdownMenuItem>
            )}
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={() => onConfigure?.(plugin)}>
              <Settings className="w-4 h-4 mr-2" /> Configure
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => onViewPermissions?.(plugin)}>
              <Shield className="w-4 h-4 mr-2" /> Permissions
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => onViewSandbox?.(plugin)}>
              <Activity className="w-4 h-4 mr-2" /> Sandbox
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={handleReportError}>
              <AlertTriangle className="w-4 h-4 mr-2" /> Report Error
            </DropdownMenuItem>
            <DropdownMenuItem onClick={handleUninstall} className="text-red-400">
              <Trash2 className="w-4 h-4 mr-2" /> Uninstall
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      <p className="mt-3 text-sm text-white/60 line-clamp-2">
        {plugin.description || "No description provided"}
      </p>

      <div className="mt-4 flex items-center gap-2 flex-wrap">
        <Badge className={cn("text-xs", statusColors[plugin.status])}>
          {plugin.status}
        </Badge>
        <Badge className={cn("text-xs", typeColors[plugin.plugin_type])}>
          {plugin.plugin_type.replace("_", " ")}
        </Badge>
        <span className="text-xs text-white/40 ml-auto">
          {formatBytes(plugin.size_bytes)}
        </span>
      </div>

      {plugin.error && (
        <div className="mt-3 p-2 bg-red-500/10 border border-red-500/20 rounded text-xs text-red-400">
          {plugin.error}
        </div>
      )}
    </div>
  );
}