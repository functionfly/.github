import { useState } from "react";
import { Plus, Database, HardDrive, Trash2, Settings, Server, Lock } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogFooter,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useCreateStore, useDeleteStore } from "@/hooks/useStateFabric";
import { useStateFabricEntitlements } from "@/hooks/useBilling";
import type { StateFabricStore } from "@/types";

interface StoreConfigurationProps {
  fabricId: string;
  stores: StateFabricStore[];
}

interface StoreTypeOption {
  value: StateFabricStore["type"];
  label: string;
  icon: string;
  addonId?: "ai_memory_pack";
}

const BASE_STORE_TYPES: StoreTypeOption[] = [
  { value: "memory", label: "In-Memory", icon: "💾" },
  { value: "persistent", label: "Persistent", icon: "💿" },
  { value: "cache", label: "Cache Layer", icon: "⚡" },
  { value: "queue", label: "Message Queue", icon: "📬" },
];

const ADDON_STORE_TYPES: StoreTypeOption[] = [
  { value: "vector", label: "Vector Index", icon: "🧭", addonId: "ai_memory_pack" },
  { value: "embedding", label: "Embeddings", icon: "🫀", addonId: "ai_memory_pack" },
  { value: "ai-memory", label: "AI Memory", icon: "🧠", addonId: "ai_memory_pack" },
];

const ALL_STORE_TYPES: StoreTypeOption[] = [...BASE_STORE_TYPES, ...ADDON_STORE_TYPES];

const storeTypeIcons: Record<string, string> = ALL_STORE_TYPES.reduce(
  (acc, t) => {
    acc[t.value] = t.icon;
    return acc;
  },
  {} as Record<string, string>,
);

const storeTypeLabels: Record<string, string> = ALL_STORE_TYPES.reduce(
  (acc, t) => {
    acc[t.value] = t.label;
    return acc;
  },
  {} as Record<string, string>,
);

const statusColors: Record<string, string> = {
  active: "bg-green-500/10 text-green-400",
  inactive: "bg-gray-500/10 text-gray-400",
  error: "bg-red-500/10 text-red-400",
};

export function StoreConfiguration({ fabricId, stores }: StoreConfigurationProps) {
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [newStoreName, setNewStoreName] = useState("");
  const [newStoreType, setNewStoreType] = useState<StateFabricStore["type"]>("memory");
  const [newStoreSize, setNewStoreSize] = useState("100");
  const [newStoreRegion, setNewStoreRegion] = useState("us-east-1");

  const createStore = useCreateStore(fabricId);
  const deleteStore = useDeleteStore(fabricId);
  const { data: entitlements } = useStateFabricEntitlements();
  const hasAddon = (addonId: string) => (entitlements?.addon_ids ?? []).includes(addonId);

  const availableTypes = ALL_STORE_TYPES.filter((t) => !t.addonId || hasAddon(t.addonId));
  const isAiMemoryEnabled = hasAddon("ai_memory_pack");

  const handleCreate = async () => {
    if (!newStoreName.trim()) return;
    const selected = ALL_STORE_TYPES.find((t) => t.value === newStoreType);
    if (selected?.addonId && !isAiMemoryEnabled) {
      return;
    }
    await createStore.mutateAsync({
      name: newStoreName,
      type: newStoreType,
      maxSize: parseInt(newStoreSize) * 1024 * 1024,
      region: newStoreRegion,
    });
    setIsCreateOpen(false);
    resetForm();
  };

  const resetForm = () => {
    setNewStoreName("");
    setNewStoreType("memory");
    setNewStoreSize("100");
    setNewStoreRegion("us-east-1");
  };

  const handleDelete = async (storeId: string) => {
    if (confirm("Are you sure you want to delete this store? All data will be lost.")) {
      await deleteStore.mutateAsync(storeId);
    }
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold text-text-primary">Stores</h3>
        <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
          <DialogTrigger asChild>
            <Button size="sm">
              <Plus className="w-4 h-4 mr-2" />
              Add Store
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create New Store</DialogTitle>
            </DialogHeader>
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="name">Name</Label>
                <Input
                  id="name"
                  value={newStoreName}
                  onChange={(e) => setNewStoreName(e.target.value)}
                  placeholder="Enter store name"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="type">Type</Label>
                <Select
                  value={newStoreType}
                  onValueChange={(v) => setNewStoreType(v as StateFabricStore["type"])}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {availableTypes.map((storeType) => (
                      <SelectItem key={storeType.value} value={storeType.value}>
                        <span className="mr-2">{storeType.icon}</span>
                        {storeType.label}
                      </SelectItem>
                    ))}
                    {ADDON_STORE_TYPES.filter((t) => !hasAddon(t.addonId)).length > 0 && (
                      <SelectItem value="__addon_placeholder" disabled>
                        <span className="text-text-muted flex items-center gap-1">
                          <Lock className="w-3 h-3" />
                          AI Memory store types require AI Memory Pack add-on
                        </span>
                      </SelectItem>
                    )}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label htmlFor="size">Max Size (MB)</Label>
                <Input
                  id="size"
                  type="number"
                  value={newStoreSize}
                  onChange={(e) => setNewStoreSize(e.target.value)}
                  placeholder="100"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="region">Region</Label>
                <Input
                  id="region"
                  value={newStoreRegion}
                  onChange={(e) => setNewStoreRegion(e.target.value)}
                  placeholder="us-east-1"
                />
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setIsCreateOpen(false)}>
                Cancel
              </Button>
              <Button
                onClick={handleCreate}
                disabled={!newStoreName.trim() || (!!ADDON_STORE_TYPES.find((t) => t.value === newStoreType && !isAiMemoryEnabled))}
              >
                Create
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {/* Store List */}
      {stores.length === 0 ? (
        <Card className="p-8 text-center">
          <p className="text-text-muted mb-4">No stores configured yet</p>
          <Button onClick={() => setIsCreateOpen(true)}>
            <Plus className="w-4 h-4 mr-2" />
            Add Your First Store
          </Button>
        </Card>
      ) : (
        <div className="grid gap-4">
          {stores.map((store) => {
            const isAiStore = ["vector", "embedding", "ai-memory"].includes(store.type);
            return (
              <Card key={store.id} className={isAiStore && !isAiMemoryEnabled ? "border-brand-500/30" : undefined}>
                <CardHeader className="pb-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="w-10 h-10 rounded-lg bg-bg-secondary flex items-center justify-center text-lg">
                        {storeTypeIcons[store.type] || "💾"}
                      </div>
                      <div>
                        <CardTitle className="text-lg">{store.name}</CardTitle>
                        <div className="flex items-center gap-2 mt-1">
                          <Badge variant="secondary" className="text-xs">
                            {storeTypeLabels[store.type]}
                          </Badge>
                          <Badge className={`text-xs ${statusColors[store.status]}`}>
                            {store.status}
                          </Badge>
                          {isAiStore && !isAiMemoryEnabled && (
                            <Badge className="text-xs bg-brand-500/10 text-brand-400 border-brand-500/20">
                              <Lock className="w-3 h-3 mr-1" />
                              AI Memory Pack
                            </Badge>
                          )}
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <Button variant="ghost" size="icon" aria-label="Store settings">
                        <Settings className="w-4 h-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleDelete(store.id)}
                        aria-label="Delete store"
                      >
                        <Trash2 className="w-4 h-4 text-red-400" />
                      </Button>
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="space-y-4">
                  {/* Storage Usage */}
                  <div>
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-sm text-text-muted">Storage Usage</span>
                      <span className="text-sm font-medium">
                        {formatBytes(store.size)} / {formatBytes(store.maxSize)}
                      </span>
                    </div>
                    <Progress value={(store.size / store.maxSize) * 100} />
                  </div>

                  {/* Metrics */}
                  <div className="grid grid-cols-3 gap-4 pt-4 border-t border-border-subtle">
                    <div>
                      <p className="text-xs text-text-muted">Region</p>
                      <p className="font-medium text-sm">{store.region}</p>
                    </div>
                    <div>
                      <p className="text-xs text-text-muted">Provider</p>
                      <p className="font-medium text-sm">{store.provider || "Default"}</p>
                    </div>
                    <div>
                      <p className="text-xs text-text-muted">Created</p>
                      <p className="font-medium text-sm">
                        {new Date(store.createdAt).toLocaleDateString()}
                      </p>
                    </div>
                  </div>

                  {/* Performance */}
                  <div className="grid grid-cols-2 gap-4 pt-4 border-t border-border-subtle">
                    <div className="flex items-center gap-2">
                      <Server className="w-4 h-4 text-text-muted" />
                      <div>
                        <p className="text-xs text-text-muted">Throughput</p>
                        <p className="font-medium">
                          {store.throughput?.toFixed(1) || 0} ops/sec
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <HardDrive className="w-4 h-4 text-text-muted" />
                      <div>
                        <p className="text-xs text-text-muted">Latency</p>
                        <p className="font-medium">{store.latency?.toFixed(0) || 0} ms</p>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
}
