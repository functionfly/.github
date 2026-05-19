import { useState } from "react";
import { GlassCard, Button, Badge, Input } from "@functionfly/ui-core";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";
import {
  Camera, Clock, RotateCcw, Download, Trash2, Upload, Search,
  Calendar, HardDrive, Check, AlertCircle, ChevronRight, Layers
} from "lucide-react";

interface WorkspaceSnapshot {
  id: string;
  name: string;
  description: string;
  createdAt: Date;
  size: string;
  type: "auto" | "manual";
  status: "ready" | "creating" | "restoring" | "error";
  components: {
    graphs: number;
    workflows: number;
    plugins: number;
    settings: boolean;
  };
}

const mockSnapshots: WorkspaceSnapshot[] = [
  {
    id: "snap-1",
    name: "Pre-feature branch",
    description: "Snapshot before neural network visualization update",
    createdAt: new Date(Date.now() - 86400000 * 2),
    size: "24.5 MB",
    type: "manual",
    status: "ready",
    components: { graphs: 12, workflows: 8, plugins: 5, settings: true },
  },
  {
    id: "snap-2",
    name: "Auto-save 2024-01-15 14:32",
    description: "Automatic snapshot during editing session",
    createdAt: new Date(Date.now() - 86400000),
    size: "18.2 MB",
    type: "auto",
    status: "ready",
    components: { graphs: 10, workflows: 6, plugins: 5, settings: true },
  },
  {
    id: "snap-3",
    name: "Clean state",
    description: "After plugin cleanup and optimization",
    createdAt: new Date(Date.now() - 86400000 * 5),
    size: "15.7 MB",
    type: "manual",
    status: "ready",
    components: { graphs: 8, workflows: 5, plugins: 3, settings: true },
  },
];

export function WorkspaceSnapshotManager() {
  const [activeTab, setActiveTab] = useState("snapshots");
  const [searchQuery, setSearchQuery] = useState("");
  const [snapshots, setSnapshots] = useState<WorkspaceSnapshot[]>(mockSnapshots);
  const [selectedSnapshot, setSelectedSnapshot] = useState<string | null>(null);
  const [isCreating, setIsCreating] = useState(false);

  const filteredSnapshots = snapshots.filter(
    (s) =>
      s.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      s.description.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleCreateSnapshot = async () => {
    setIsCreating(true);
    const newSnapshot: WorkspaceSnapshot = {
      id: `snap-${Date.now()}`,
      name: `Snapshot ${new Date().toLocaleString()}`,
      description: "Manual workspace snapshot",
      createdAt: new Date(),
      size: "Calculating...",
      type: "manual",
      status: "creating",
      components: { graphs: 12, workflows: 8, plugins: 5, settings: true },
    };
    setSnapshots([newSnapshot, ...snapshots]);
    await new Promise((r) => setTimeout(r, 2000));
    setSnapshots((prev) =>
      prev.map((s) =>
        s.id === newSnapshot.id ? { ...s, status: "ready", size: "22.3 MB" } : s
      )
    );
    setIsCreating(false);
  };

  const handleRestore = async (snapshotId: string) => {
    setSnapshots((prev) =>
      prev.map((s) =>
        s.id === snapshotId ? { ...s, status: "restoring" } : s
      )
    );
    await new Promise((r) => setTimeout(r, 2000));
    setSnapshots((prev) =>
      prev.map((s) =>
        s.id === snapshotId ? { ...s, status: "ready" } : s
      )
    );
  };

  const handleDelete = (snapshotId: string) => {
    setSnapshots((prev) => prev.filter((s) => s.id !== snapshotId));
  };

  const formatDate = (date: Date) => {
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const days = Math.floor(diff / 86400000);
    if (days === 0) return "Today";
    if (days === 1) return "Yesterday";
    if (days < 7) return `${days} days ago`;
    return date.toLocaleDateString();
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between p-5 border-b border-white/10">
        <div>
          <h2 className="text-xl font-semibold text-white">Workspace Snapshots</h2>
          <p className="text-sm text-white/60">Save and restore workspace states</p>
        </div>
        <div className="flex items-center gap-3">
          <Badge variant="outline" className="text-white/60 border-white/20">
            <HardDrive className="w-3 h-3 mr-1" />
            {snapshots.length} snapshots
          </Badge>
          <Button
            onClick={handleCreateSnapshot}
            disabled={isCreating}
            className="gap-2 bg-gradient-to-r from-orange-500 to-red-500 hover:from-orange-400 hover:to-red-400"
          >
            <Camera className="w-4 h-4" />
            {isCreating ? "Creating..." : "Create Snapshot"}
          </Button>
        </div>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col">
        <div className="px-5 pt-5">
          <TabsList className="inline-flex h-auto flex-wrap gap-1 rounded-xl border border-white/10 bg-white/5 p-1.5 text-white/60">
            <TabsTrigger
              value="snapshots"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <Camera className="h-4 w-4 shrink-0" />
              Snapshots
            </TabsTrigger>
            <TabsTrigger
              value="scheduled"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <Clock className="h-4 w-4 shrink-0" />
              Scheduled
            </TabsTrigger>
            <TabsTrigger
              value="storage"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <HardDrive className="h-4 w-4 shrink-0" />
              Storage
            </TabsTrigger>
          </TabsList>
        </div>

        <div className="flex-1 overflow-auto p-4">
          <TabsContent value="snapshots" className="mt-0">
            <div className="space-y-4">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-white/40" />
                <Input
                  placeholder="Search snapshots..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-9 bg-white/5 border-white/10"
                />
              </div>

              <div className="space-y-3">
                {filteredSnapshots.map((snapshot) => (
                  <GlassCard
                    key={snapshot.id}
                    className={cn(
                      "p-4 transition-all duration-200 cursor-pointer",
                      selectedSnapshot === snapshot.id && "ring-2 ring-orange-500/30"
                    )}
                    onClick={() => setSelectedSnapshot(
                      selectedSnapshot === snapshot.id ? null : snapshot.id
                    )}
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex items-start gap-4">
                        <div
                          className={cn(
                            "w-10 h-10 rounded-xl flex items-center justify-center",
                            snapshot.type === "auto"
                              ? "bg-blue-500/20 text-blue-400"
                              : "bg-orange-500/20 text-orange-400"
                          )}
                        >
                          <Camera className="w-5 h-5" />
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2 mb-1">
                            <h3 className="font-medium text-white truncate">{snapshot.name}</h3>
                            <Badge
                              variant="outline"
                              className={cn(
                                "text-xs",
                                snapshot.type === "auto"
                                  ? "text-blue-400 border-blue-400/30"
                                  : "text-orange-400 border-orange-400/30"
                              )}
                            >
                              {snapshot.type}
                            </Badge>
                            {snapshot.status !== "ready" && (
                              <Badge
                                variant="outline"
                                className={cn(
                                  "text-xs",
                                  snapshot.status === "creating"
                                    ? "text-yellow-400 border-yellow-400/30"
                                    : snapshot.status === "restoring"
                                    ? "text-purple-400 border-purple-400/30"
                                    : "text-red-400 border-red-400/30"
                                )}
                              >
                                {snapshot.status}
                              </Badge>
                            )}
                          </div>
                          <p className="text-sm text-white/60 line-clamp-1 mb-2">
                            {snapshot.description}
                          </p>
                          <div className="flex items-center gap-4 text-xs text-white/50">
                            <span className="flex items-center gap-1">
                              <Calendar className="w-3 h-3" />
                              {formatDate(snapshot.createdAt)}
                            </span>
                            <span className="flex items-center gap-1">
                              <HardDrive className="w-3 h-3" />
                              {snapshot.size}
                            </span>
                            <span className="flex items-center gap-1">
                              <Layers className="w-3 h-3" />
                              {snapshot.components.graphs} graphs
                            </span>
                          </div>
                        </div>
                      </div>

                      {selectedSnapshot === snapshot.id && (
                        <div className="flex items-center gap-2" onClick={(e) => e.stopPropagation()}>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => handleRestore(snapshot.id)}
                            disabled={snapshot.status !== "ready"}
                            className="gap-1 text-emerald-400 border-emerald-400/30 hover:bg-emerald-400/10"
                          >
                            <RotateCcw className="w-3 h-3" />
                            Restore
                          </Button>
                          <Button
                            size="icon"
                            variant="ghost"
                            onClick={() => handleDelete(snapshot.id)}
                            className="text-red-400 hover:text-red-300 hover:bg-red-400/10"
                          >
                            <Trash2 className="w-4 h-4" />
                          </Button>
                        </div>
                      )}
                    </div>

                    {selectedSnapshot === snapshot.id && (
                      <div className="mt-4 pt-4 border-t border-white/10">
                        <h4 className="text-xs font-semibold text-white/60 mb-3">INCLUDED COMPONENTS</h4>
                        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
                          <div className="p-3 rounded-lg bg-white/5">
                            <p className="text-2xl font-bold text-white">{snapshot.components.graphs}</p>
                            <p className="text-xs text-white/60">Graphs</p>
                          </div>
                          <div className="p-3 rounded-lg bg-white/5">
                            <p className="text-2xl font-bold text-white">{snapshot.components.workflows}</p>
                            <p className="text-xs text-white/60">Workflows</p>
                          </div>
                          <div className="p-3 rounded-lg bg-white/5">
                            <p className="text-2xl font-bold text-white">{snapshot.components.plugins}</p>
                            <p className="text-xs text-white/60">Plugins</p>
                          </div>
                          <div className="p-3 rounded-lg bg-white/5">
                            <p className="text-2xl font-bold text-white">
                              {snapshot.components.settings ? "Yes" : "No"}
                            </p>
                            <p className="text-xs text-white/60">Settings</p>
                          </div>
                        </div>
                      </div>
                    )}
                  </GlassCard>
                ))}
              </div>

              {filteredSnapshots.length === 0 && (
                <GlassCard className="flex flex-col items-center justify-center h-48">
                  <Camera className="w-10 h-10 text-white/30 mb-3" />
                  <p className="text-white/60">No snapshots found</p>
                  <p className="text-sm text-white/40">Create a snapshot to save your workspace state</p>
                </GlassCard>
              )}
            </div>
          </TabsContent>

          <TabsContent value="scheduled" className="mt-0">
            <div className="space-y-4">
              <GlassCard className="p-6">
                <div className="flex items-center gap-4 mb-4">
                  <div className="w-12 h-12 rounded-xl bg-emerald-500/20 flex items-center justify-center">
                    <Clock className="w-6 h-6 text-emerald-400" />
                  </div>
                  <div>
                    <h3 className="font-semibold text-white">Automatic Snapshots</h3>
                    <p className="text-sm text-white/60">Create snapshots on a schedule</p>
                  </div>
                </div>
                <div className="space-y-3">
                  <div className="flex items-center justify-between p-3 rounded-lg bg-white/5">
                    <span className="text-sm text-white/80">Every hour</span>
                    <Badge className="bg-emerald-500/20 text-emerald-400 border-emerald-500/30">Enabled</Badge>
                  </div>
                  <div className="flex items-center justify-between p-3 rounded-lg bg-white/5">
                    <span className="text-sm text-white/80">Before major operations</span>
                    <Badge className="bg-emerald-500/20 text-emerald-400 border-emerald-500/30">Enabled</Badge>
                  </div>
                  <div className="flex items-center justify-between p-3 rounded-lg bg-white/5">
                    <span className="text-sm text-white/80">Daily at midnight</span>
                    <Badge variant="outline" className="text-white/40 border-white/20">Disabled</Badge>
                  </div>
                </div>
              </GlassCard>
            </div>
          </TabsContent>

          <TabsContent value="storage" className="mt-0">
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <GlassCard className="p-4">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm text-white/60">Used Storage</span>
                    <HardDrive className="w-4 h-4 text-white/40" />
                  </div>
                  <p className="text-2xl font-bold text-white">58.4 MB</p>
                  <p className="text-xs text-white/40">of 500 MB</p>
                </GlassCard>
                <GlassCard className="p-4">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm text-white/60">Total Snapshots</span>
                    <Camera className="w-4 h-4 text-white/40" />
                  </div>
                  <p className="text-2xl font-bold text-white">{snapshots.length}</p>
                  <p className="text-xs text-white/40">11.7% of limit</p>
                </GlassCard>
              </div>
              <Button variant="outline" className="w-full gap-2 text-red-400 border-red-400/30 hover:bg-red-400/10">
                <Trash2 className="w-4 h-4" />
                Clear All Snapshots
              </Button>
            </div>
          </TabsContent>
        </div>
      </Tabs>
    </div>
  );
}