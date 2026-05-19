import { useState } from "react";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@functionfly/ui-core";
import { PluginCard } from "./PluginCard";
import { PluginPermissionsViewer } from "./PluginPermissionsViewer";
import { PluginSandboxInspector } from "./PluginSandboxInspector";
import { PluginVersionManager } from "./PluginVersionManager";
import { PluginTelemetryPanel } from "./PluginTelemetryPanel";
import { PluginUpdateCenter } from "./PluginUpdateCenter";
import { usePlugins, useInstallPlugin, useEnablePlugin, useDisablePlugin, useConfigurePlugin, type Plugin } from "@/hooks/usePlugin";
import { Spinner, GlassCard, Button, Input, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@functionfly/ui-core";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { ConfigurePluginDialog } from "./ConfigurePluginDialog";
import { Plus, Search, Filter, Grid, List, Puzzle, Store, Shield, Container, GitBranch, BarChart3, CheckSquare, Square, X, Link, Trash2, Play, Pause } from "lucide-react";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

export function PluginManager() {
  const [activeTab, setActiveTab] = useState("installed");
  const [searchQuery, setSearchQuery] = useState("");
  const [typeFilter, setTypeFilter] = useState<string>("all");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");
  const [selectedPlugins, setSelectedPlugins] = useState<Set<string>>(new Set());
  const [showInstallUrlDialog, setShowInstallUrlDialog] = useState(false);
  const [installUrl, setInstallUrl] = useState("");
  const [selectedPluginForConfig, setSelectedPluginForConfig] = useState<Plugin | null>(null);

  const { data, isLoading } = usePlugins({
    search: searchQuery || undefined,
    type: typeFilter !== "all" ? (typeFilter as any) : undefined,
    status: statusFilter !== "all" ? (statusFilter as any) : undefined,
  });

  const installMutation = useInstallPlugin();
  const enableMutation = useEnablePlugin();
  const disableMutation = useDisablePlugin();
  const configureMutation = useConfigurePlugin();

  const plugins = data?.plugins || [];
  const enabledPlugins = plugins.filter((p: Plugin) => p.status === "enabled");
  const disabledPlugins = plugins.filter((p: Plugin) => p.status === "disabled");
  const errorPlugins = plugins.filter((p: Plugin) => p.status === "error");

  const handleInstallPlugin = async (pluginData: any) => {
    await installMutation.mutateAsync(pluginData);
  };

  const handleConfigurePlugin = (plugin: Plugin) => {
    setSelectedPluginForConfig(plugin);
  };

  const handleSavePluginConfig = async (pluginId: string, config: Record<string, string>) => {
    await configureMutation.mutateAsync({ pluginId, config });
  };

  const handleInstallFromUrl = async () => {
    if (!installUrl.trim()) {
      toast.error("Please enter a plugin URL");
      return;
    }
    toast.info("Installing plugin from URL... (Feature coming soon)");
    setShowInstallUrlDialog(false);
    setInstallUrl("");
  };

  const togglePluginSelection = (pluginId: string) => {
    const newSelected = new Set(selectedPlugins);
    if (newSelected.has(pluginId)) {
      newSelected.delete(pluginId);
    } else {
      newSelected.add(pluginId);
    }
    setSelectedPlugins(newSelected);
  };

  const toggleSelectAll = () => {
    if (selectedPlugins.size === plugins.length) {
      setSelectedPlugins(new Set());
    } else {
      setSelectedPlugins(new Set(plugins.map((p: Plugin) => p.id)));
    }
  };

  const handleBulkEnable = async () => {
    for (const pluginId of selectedPlugins) {
      await enableMutation.mutateAsync(pluginId);
    }
    setSelectedPlugins(new Set());
    toast.success(`Enabled ${selectedPlugins.size} plugins`);
  };

  const handleBulkDisable = async () => {
    for (const pluginId of selectedPlugins) {
      await disableMutation.mutateAsync(pluginId);
    }
    setSelectedPlugins(new Set());
    toast.success(`Disabled ${selectedPlugins.size} plugins`);
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between p-5 border-b border-white/10">
        <div>
          <h2 className="text-xl font-semibold text-white">Plugin Manager</h2>
          <p className="text-sm text-white/60">Manage plugins, permissions, and sandbox settings</p>
        </div>
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-white/5 border border-white/10">
            <span className="w-2 h-2 rounded-full bg-emerald-400" />
            <span className="text-sm text-white/80 font-medium">{plugins.length}</span>
            <span className="text-sm text-white/50">plugins</span>
          </div>
          <Button
            variant="outline"
            onClick={() => setShowInstallUrlDialog(true)}
            className="h-9 px-4 inline-flex items-center gap-2 border-white/20 text-white/80 hover:text-white hover:bg-white/10 text-sm whitespace-nowrap shrink-0"
          >
            <Link className="w-4 h-4 shrink-0" />
            <span>Install from URL</span>
          </Button>
          <Button
            className="bg-gradient-to-r from-orange-500 to-red-500 hover:from-orange-400 hover:to-red-400 text-white shadow-lg shadow-orange-500/25 hover:shadow-orange-500/40 transition-all duration-200 hover:scale-[1.02] active:scale-[0.98] rounded-lg h-9 px-4 text-sm whitespace-nowrap"
          >
            <Plus className="w-4 h-4 shrink-0" />
            Add Plugin
          </Button>
        </div>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col">
        <div className="px-5 pt-5">
          <TabsList className="inline-flex h-auto flex-wrap gap-1 rounded-xl border border-white/10 bg-white/5 p-1.5 text-white/60 backdrop-blur-sm">
            <TabsTrigger
              value="installed"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10 data-[state=active]:shadow-sm hover:text-white/80"
            >
              <Puzzle className="h-4 w-4 shrink-0" />
              Installed
            </TabsTrigger>
            <TabsTrigger
              value="store"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10 data-[state=active]:shadow-sm hover:text-white/80"
            >
              <Store className="h-4 w-4 shrink-0" />
              Plugin Store
            </TabsTrigger>
            <TabsTrigger
              value="permissions"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10 data-[state=active]:shadow-sm hover:text-white/80"
            >
              <Shield className="h-4 w-4 shrink-0" />
              Permissions
            </TabsTrigger>
            <TabsTrigger
              value="sandbox"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10 data-[state=active]:shadow-sm hover:text-white/80"
            >
              <Container className="h-4 w-4 shrink-0" />
              Sandbox
            </TabsTrigger>
            <TabsTrigger
              value="versions"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10 data-[state=active]:shadow-sm hover:text-white/80"
            >
              <GitBranch className="h-4 w-4 shrink-0" />
              Versions
            </TabsTrigger>
            <TabsTrigger
              value="telemetry"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10 data-[state=active]:shadow-sm hover:text-white/80"
            >
              <BarChart3 className="h-4 w-4 shrink-0" />
              Telemetry
            </TabsTrigger>
          </TabsList>
        </div>

        <div className="flex-1 overflow-auto p-4">
          <TabsContent value="installed" className="h-full">
            <div className="space-y-4">
              <div className="flex items-center gap-4 flex-wrap">
                <div className="relative flex-1 min-w-[200px]">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-white/40" />
                  <Input
                    placeholder="Search plugins..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="pl-9 bg-white/5 border-white/10"
                  />
                </div>
                <Select value={typeFilter} onValueChange={setTypeFilter}>
                  <SelectTrigger className="w-[150px] bg-white/5 border-white/10">
                    <SelectValue placeholder="Type" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Types</SelectItem>
                    <SelectItem value="ui">UI</SelectItem>
                    <SelectItem value="graph">Graph</SelectItem>
                    <SelectItem value="ai_tool">AI Tool</SelectItem>
                    <SelectItem value="runtime">Runtime</SelectItem>
                    <SelectItem value="infrastructure">Infrastructure</SelectItem>
                  </SelectContent>
                </Select>
                <Select value={statusFilter} onValueChange={setStatusFilter}>
                  <SelectTrigger className="w-[150px] bg-white/5 border-white/10">
                    <SelectValue placeholder="Status" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Status</SelectItem>
                    <SelectItem value="enabled">Enabled</SelectItem>
                    <SelectItem value="disabled">Disabled</SelectItem>
                    <SelectItem value="error">Error</SelectItem>
                    <SelectItem value="paused">Paused</SelectItem>
                  </SelectContent>
                </Select>
                <div className="flex items-center gap-0 p-0.5 rounded-lg bg-white/5 border border-white/10">
                  <Button
                    size="icon"
                    variant="ghost"
                    className={cn(
                      "h-8 w-8 rounded-md transition-all duration-200",
                      viewMode === "grid"
                        ? "bg-white/10 text-white shadow-sm"
                        : "text-white/40 hover:text-white/70 hover:bg-white/5"
                    )}
                    onClick={() => setViewMode("grid")}
                  >
                    <Grid className="w-4 h-4" />
                  </Button>
                  <Button
                    size="icon"
                    variant="ghost"
                    className={cn(
                      "h-8 w-8 rounded-md transition-all duration-200",
                      viewMode === "list"
                        ? "bg-white/10 text-white shadow-sm"
                        : "text-white/40 hover:text-white/70 hover:bg-white/5"
                    )}
                    onClick={() => setViewMode("list")}
                  >
                    <List className="w-4 h-4" />
                  </Button>
                </div>
              </div>

              {plugins.length > 0 && (
                <div className="flex items-center gap-3 p-3 bg-white/5 rounded-lg border border-white/10">
                  <button
                    onClick={toggleSelectAll}
                    className="flex items-center gap-2 text-sm text-white/80 hover:text-white"
                  >
                    {selectedPlugins.size === plugins.length ? (
                      <CheckSquare className="w-4 h-4 text-orange-400" />
                    ) : (
                      <Square className="w-4 h-4" />
                    )}
                    Select All
                  </button>

                  {selectedPlugins.size > 0 && (
                    <>
                      <div className="w-px h-4 bg-white/20" />
                      <span className="text-sm text-white/60">{selectedPlugins.size} selected</span>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={handleBulkEnable}
                        className="text-emerald-400 hover:text-emerald-300 hover:bg-emerald-400/10"
                      >
                        <Play className="w-3 h-3 mr-1" />
                        Enable
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={handleBulkDisable}
                        className="text-yellow-400 hover:text-yellow-300 hover:bg-yellow-400/10"
                      >
                        <Pause className="w-3 h-3 mr-1" />
                        Disable
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => setSelectedPlugins(new Set())}
                        className="text-red-400 hover:text-red-300 hover:bg-red-400/10"
                      >
                        <X className="w-3 h-3 mr-1" />
                        Clear
                      </Button>
                    </>
                  )}
                </div>
              )}

              {isLoading ? (
                <div className="flex items-center justify-center h-64">
                  <Spinner className="w-8 h-8" />
                </div>
              ) : plugins.length === 0 ? (
                <GlassCard className="flex flex-col items-center justify-center h-64">
                  <p className="text-white/60 mb-2">No plugins installed</p>
                  <p className="text-sm text-white/40">Install plugins from the Plugin Store</p>
                </GlassCard>
              ) : (
                <div className={cn(
                  "grid gap-4",
                  viewMode === "grid" ? "grid-cols-1 md:grid-cols-2 lg:grid-cols-3" : "grid-cols-1"
                )}>
                  {plugins.map((plugin) => (
                    <div key={plugin.id} className="relative">
                      {viewMode === "list" && (
                        <button
                          onClick={() => togglePluginSelection(plugin.id)}
                          className="absolute left-2 top-1/2 -translate-y-1/2 z-10"
                        >
                          {selectedPlugins.has(plugin.id) ? (
                            <CheckSquare className="w-4 h-4 text-orange-400" />
                          ) : (
                            <Square className="w-4 h-4 text-white/40" />
                          )}
                        </button>
                      )}
                      <PluginCard
                        key={plugin.id}
                        plugin={plugin}
                        onConfigure={handleConfigurePlugin}
                      />
                    </div>
                  ))}
                </div>
              )}
            </div>
          </TabsContent>

          <TabsContent value="store">
            <PluginUpdateCenter onInstall={handleInstallPlugin} installedPlugins={plugins} />
          </TabsContent>

          <TabsContent value="permissions">
            <PluginPermissionsViewer plugins={plugins} />
          </TabsContent>

          <TabsContent value="sandbox">
            <PluginSandboxInspector plugins={plugins} />
          </TabsContent>

          <TabsContent value="versions">
            <PluginVersionManager plugins={plugins} />
          </TabsContent>

          <TabsContent value="telemetry">
            <PluginTelemetryPanel plugins={plugins} />
          </TabsContent>
        </div>
      </Tabs>

      <Dialog open={showInstallUrlDialog} onOpenChange={setShowInstallUrlDialog}>
        <DialogContent className="bg-[#1a1a2e] border-white/10">
          <DialogHeader>
            <DialogTitle className="text-white">Install Plugin from URL</DialogTitle>
            <DialogDescription className="text-white/60">
              Enter the URL of a plugin manifest or package to install a custom plugin.
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <Input
              placeholder="https://example.com/plugin/manifest.json"
              value={installUrl}
              onChange={(e) => setInstallUrl(e.target.value)}
              className="bg-white/5 border-white/10 text-white"
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowInstallUrlDialog(false)}>
              Cancel
            </Button>
            <Button onClick={handleInstallFromUrl}>
              <Link className="w-4 h-4 mr-2" />
              Install
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {selectedPluginForConfig && (
        <ConfigurePluginDialog
          plugin={selectedPluginForConfig}
          open={!!selectedPluginForConfig}
          onOpenChange={(open) => !open && setSelectedPluginForConfig(null)}
          onSaveConfig={handleSavePluginConfig}
        />
      )}
    </div>
  );
}