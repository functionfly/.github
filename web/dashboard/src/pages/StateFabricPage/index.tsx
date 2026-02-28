import { useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Plus,
  Database,
  Zap,
  Network,
  Activity,
  Settings,
  MoreVertical,
  Search,
  RefreshCw,
  AlertTriangle,
  Loader2,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { StatCard } from "@/components/common/StatCard";
import { StatusBadge } from "@/components/common/StatusBadge";
import { LoadingSpinner } from "@/components/common/LoadingSpinner";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  useStateFabrics,
  useDeleteStateFabric,
} from "@/hooks/useStateFabric";
import type { StateFabric } from "@/types";

const getTypeIcon = (type: string) => {
  switch (type) {
    case "session":
      return "👤";
    case "catalog":
      return "📦";
    case "cache":
      return "⚡";
    case "workflow":
      return "🔄";
    default:
      return "🧵";
  }
};

const getTypeColor = (type: string) => {
  switch (type) {
    case "session":
      return "bg-blue-500/10 border-blue-500/20";
    case "catalog":
      return "bg-green-500/10 border-green-500/20";
    case "cache":
      return "bg-yellow-500/10 border-yellow-500/20";
    case "workflow":
      return "bg-purple-500/10 border-purple-500/20";
    default:
      return "bg-gray-500/10 border-gray-500/20";
  }
};

const getTypeLabel = (type: string) => {
  switch (type) {
    case "session":
      return "Session Store";
    case "catalog":
      return "Data Catalog";
    case "cache":
      return "Cache Layer";
    case "workflow":
      return "Workflow Engine";
    default:
      return "Custom Fabric";
  }
};

export function StateFabricPage() {
  const navigate = useNavigate();
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [typeFilter, setTypeFilter] = useState<string>("all");
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [fabricToDelete, setFabricToDelete] = useState<StateFabric | null>(null);

  const { data: fabrics, isLoading, error, refetch } = useStateFabrics();
  const deleteFabric = useDeleteStateFabric();

  const filteredFabrics = fabrics?.filter((fabric) => {
    const matchesSearch =
      fabric.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      fabric.description.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesStatus =
      statusFilter === "all" || fabric.status === statusFilter;
    const matchesType = typeFilter === "all" || fabric.type === typeFilter;
    return matchesSearch && matchesStatus && matchesType;
  });

  const stats = {
    total: fabrics?.length || 0,
    active: fabrics?.filter((f) => f.status === "online").length || 0,
    stores: fabrics?.reduce((acc, f) => acc + (f.stores?.length || 0), 0) || 0,
    pipelines:
      fabrics?.reduce((acc, f) => acc + (f.pipelines?.length || 0), 0) || 0,
  };

  const handleCreateFabric = () => {
    navigate("/state-fabric/new");
  };

  const handleViewFabric = (id: string) => {
    navigate(`/state-fabric/${id}`);
  };

  const handleEditFabric = (id: string) => {
    navigate(`/state-fabric/${id}/edit`);
  };

  const handleDeleteClick = (fabric: StateFabric) => {
    setFabricToDelete(fabric);
    setDeleteDialogOpen(true);
  };

  const handleConfirmDelete = async () => {
    if (fabricToDelete) {
      await deleteFabric.mutateAsync(fabricToDelete.id);
      setDeleteDialogOpen(false);
      setFabricToDelete(null);
    }
  };

  const handleCancelDelete = () => {
    setDeleteDialogOpen(false);
    setFabricToDelete(null);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-96">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-text-primary">
              State Fabric
            </h1>
            <p className="text-text-secondary">
              Manage state and data orchestration across your applications
            </p>
          </div>
        </div>
        <Card className="p-12 text-center">
          <div className="text-red-400 mb-4">Failed to load state fabrics</div>
          <Button onClick={() => refetch()} variant="outline">
            <RefreshCw className="w-4 h-4 mr-2" />
            Retry
          </Button>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-text-primary">State Fabric</h1>
          <p className="text-text-secondary">
            Manage state and data orchestration across your applications
          </p>
        </div>
        <Button className="gap-2" onClick={handleCreateFabric}>
          <Plus className="w-4 h-4" />
          Create State Fabric
        </Button>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="Total Fabrics"
          value={stats.total.toString()}
          icon={<Database className="w-5 h-5 text-blue-500" />}
          trend="neutral"
        />
        <StatCard
          title="Active Fabrics"
          value={stats.active.toString()}
          change={{
            value: stats.total > 0 ? Math.round((stats.active / stats.total) * 100) : 0,
            label: "of total",
          }}
          icon={<Activity className="w-5 h-5 text-green-500" />}
          trend="up"
        />
        <StatCard
          title="Total Stores"
          value={stats.stores.toString()}
          icon={<Network className="w-5 h-5 text-purple-500" />}
          trend="neutral"
        />
        <StatCard
          title="Active Pipelines"
          value={stats.pipelines.toString()}
          icon={<Zap className="w-5 h-5 text-yellow-500" />}
          trend="neutral"
        />
      </div>

      {/* Toolbar */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
          <Input
            type="text"
            placeholder="Search state fabrics..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>
        <div className="flex gap-2">
          <Select value={statusFilter} onValueChange={setStatusFilter}>
            <SelectTrigger className="w-[140px]">
              <SelectValue placeholder="All Status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Status</SelectItem>
              <SelectItem value="online">Online</SelectItem>
              <SelectItem value="degraded">Degraded</SelectItem>
              <SelectItem value="offline">Offline</SelectItem>
            </SelectContent>
          </Select>
          <Select value={typeFilter} onValueChange={setTypeFilter}>
            <SelectTrigger className="w-[140px]">
              <SelectValue placeholder="All Types" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Types</SelectItem>
              <SelectItem value="session">Session</SelectItem>
              <SelectItem value="catalog">Catalog</SelectItem>
              <SelectItem value="cache">Cache</SelectItem>
              <SelectItem value="workflow">Workflow</SelectItem>
              <SelectItem value="custom">Custom</SelectItem>
            </SelectContent>
          </Select>
          <Button variant="outline" size="icon" onClick={() => refetch()}>
            <RefreshCw className="w-4 h-4" />
          </Button>
        </div>
      </div>

      {/* State Fabrics Grid */}
      {filteredFabrics && filteredFabrics.length > 0 ? (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {filteredFabrics.map((fabric) => (
            <Card
              key={fabric.id}
              className="hover:border-brand-500/30 transition-colors"
            >
              <CardHeader className="pb-3">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-3">
                    <div
                      className={`w-10 h-10 rounded-lg flex items-center justify-center text-lg ${getTypeColor(
                        fabric.type
                      )}`}
                    >
                      {getTypeIcon(fabric.type)}
                    </div>
                    <div>
                      <CardTitle className="text-lg text-text-primary">
                        {fabric.name}
                      </CardTitle>
                      <p className="text-sm text-text-secondary">
                        {fabric.description}
                      </p>
                      <Badge variant="secondary" className="mt-1 text-xs">
                        {getTypeLabel(fabric.type)}
                      </Badge>
                    </div>
                  </div>
                  <StatusBadge status={fabric.status} />
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                {/* Metrics */}
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-1">
                    <p className="text-xs text-text-muted uppercase tracking-wide">
                      Throughput
                    </p>
                    <p className="text-lg font-semibold text-text-primary">
                      {fabric.metrics?.operationsPerSecond
                        ? `${fabric.metrics.operationsPerSecond.toFixed(1)} ops/sec`
                        : "N/A"}
                    </p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-xs text-text-muted uppercase tracking-wide">
                      Latency
                    </p>
                    <p className="text-lg font-semibold text-text-primary">
                      {fabric.metrics?.averageLatency
                        ? `${fabric.metrics.averageLatency.toFixed(0)}ms`
                        : "N/A"}
                    </p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-xs text-text-muted uppercase tracking-wide">
                      Stores
                    </p>
                    <p className="text-lg font-semibold text-text-primary">
                      {fabric.stores?.length || 0}
                    </p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-xs text-text-muted uppercase tracking-wide">
                      Pipelines
                    </p>
                    <p className="text-lg font-semibold text-text-primary">
                      {fabric.pipelines?.length || 0}
                    </p>
                  </div>
                </div>

                {/* Footer */}
                <div className="flex items-center justify-between pt-4 border-t border-border-subtle">
                  <p className="text-xs text-text-muted">
                    Updated {new Date(fabric.updatedAt).toLocaleDateString()}
                  </p>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="text-text-secondary"
                      >
                        <MoreVertical className="w-4 h-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent
                      align="end"
                      className="bg-bg-tertiary border-white/8"
                    >
                      <DropdownMenuItem
                        className="gap-2"
                        onClick={() => handleViewFabric(fabric.id)}
                      >
                        <Activity className="w-4 h-4" />
                        View Details
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        className="gap-2"
                        onClick={() => handleEditFabric(fabric.id)}
                      >
                        <Settings className="w-4 h-4" />
                        Configure
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        className="gap-2 text-red-400"
                        onClick={() => handleDeleteClick(fabric)}
                      >
                        <Database className="w-4 h-4" />
                        Delete
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      ) : (
        <Card className="p-12 text-center">
          <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-bg-tertiary flex items-center justify-center">
            <Database className="w-8 h-8 text-text-muted" />
          </div>
          <h3 className="text-lg font-medium text-text-primary mb-2">
            No state fabrics yet
          </h3>
          <p className="text-text-secondary mb-6">
            {searchQuery
              ? "No fabrics match your search."
              : "Create your first state fabric to get started with data orchestration."}
          </p>
          <Button onClick={handleCreateFabric}>
            <Plus className="w-4 h-4 mr-2" />
            Create State Fabric
          </Button>
        </Card>
      )}

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <AlertTriangle className="w-5 h-5 text-red-500" />
              Delete State Fabric
            </DialogTitle>
            <DialogDescription>
              Are you sure you want to delete "{fabricToDelete?.name}"? This action cannot be undone.
              All associated stores and pipelines will be permanently removed.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={handleCancelDelete}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleConfirmDelete}
              disabled={deleteFabric.isPending}
            >
              {deleteFabric.isPending ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Deleting...
                </>
              ) : (
                "Delete"
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
