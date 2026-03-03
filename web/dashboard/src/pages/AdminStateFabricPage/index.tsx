import { useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Database,
  Search,
  Filter,
  MoreVertical,
  Pause,
  Play,
  AlertTriangle,
  ArrowLeft,
  RefreshCw,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { LoadingSpinner } from "@/components/common/LoadingSpinner";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  useAdminStateFabrics,
  useAdminStateFabricStats,
  useSuspendStateFabric,
  useResumeStateFabric,
} from "@/hooks/useAdminStateFabric";
import { StatCard } from "@/components/common/StatCard";

export function AdminStateFabricPage() {
  const navigate = useNavigate();
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");

  const { data: fabricsData, isLoading, error, refetch } = useAdminStateFabrics();
  const { data: stats } = useAdminStateFabricStats();
  const suspendFabric = useSuspendStateFabric();
  const resumeFabric = useResumeStateFabric();

  const fabrics = fabricsData?.fabrics || [];
  const total = fabricsData?.total || 0;

  const filteredFabrics = fabrics.filter((fabric) => {
    const matchesSearch =
      fabric.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      fabric.tenantId?.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesStatus =
      statusFilter === "all" || fabric.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  const handleSuspend = async (fabricId: string) => {
    const reason = prompt("Enter reason for suspension:");
    if (reason) {
      await suspendFabric.mutateAsync({ fabricId, reason });
    }
  };

  const handleResume = async (fabricId: string) => {
    await resumeFabric.mutateAsync(fabricId);
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "online":
        return "bg-green-500/10 text-green-400 border-green-500/20";
      case "degraded":
        return "bg-yellow-500/10 text-yellow-400 border-yellow-500/20";
      case "offline":
        return "bg-red-500/10 text-red-400 border-red-500/20";
      default:
        return "bg-gray-500/10 text-text-secondary border-gray-500/20";
    }
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
        <Button variant="ghost" onClick={() => navigate("/admin")}>
          <ArrowLeft className="w-4 h-4 mr-2" />
          Back to Admin
        </Button>
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
      <div className="flex items-center justify-between">
        <Button variant="ghost" onClick={() => navigate("/admin")}>
          <ArrowLeft className="w-4 h-4 mr-2" />
          Back to Admin
        </Button>
        <div className="flex-1 text-center">
          <h1 className="text-2xl font-bold text-text-primary">
            State Fabric Management
          </h1>
          <p className="text-text-secondary">
            Manage all state fabrics across tenants
          </p>
        </div>
        <Button onClick={() => refetch()} variant="outline">
          <RefreshCw className="w-4 h-4 mr-2" />
          Refresh
        </Button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="Total Fabrics"
          value={stats?.totalFabrics?.toString() || "0"}
          icon={<Database className="w-5 h-5 text-blue-500" />}
          trend="neutral"
        />
        <StatCard
          title="Active Fabrics"
          value={stats?.activeFabrics?.toString() || "0"}
          icon={<Play className="w-5 h-5 text-green-500" />}
          trend="up"
        />
        <StatCard
          title="Total Stores"
          value={stats?.totalStores?.toString() || "0"}
          icon={<Database className="w-5 h-5 text-purple-500" />}
          trend="neutral"
        />
        <StatCard
          title="Total Pipelines"
          value={stats?.totalPipelines?.toString() || "0"}
          icon={<RefreshCw className="w-5 h-5 text-yellow-500" />}
          trend="neutral"
        />
      </div>

      {/* Toolbar */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
          <Input
            placeholder="Search fabrics by name or tenant..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="px-3 py-2 bg-bg-secondary border border-border-subtle rounded-lg text-text-primary text-sm"
        >
          <option value="all">All Status</option>
          <option value="online">Online</option>
          <option value="degraded">Degraded</option>
          <option value="offline">Offline</option>
        </select>
      </div>

      {/* Fabrics Table */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center justify-between">
            <span>All State Fabrics</span>
            <span className="text-sm text-text-muted font-normal">
              {total.toLocaleString()} total
            </span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          {filteredFabrics.length === 0 ? (
            <div className="py-8 text-center text-text-muted">
              No state fabrics found
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-border-subtle">
                    <th className="text-left py-3 px-4 text-text-muted font-medium">
                      Name
                    </th>
                    <th className="text-left py-3 px-4 text-text-muted font-medium">
                      Tenant
                    </th>
                    <th className="text-left py-3 px-4 text-text-muted font-medium">
                      Type
                    </th>
                    <th className="text-left py-3 px-4 text-text-muted font-medium">
                      Status
                    </th>
                    <th className="text-left py-3 px-4 text-text-muted font-medium">
                      Stores
                    </th>
                    <th className="text-left py-3 px-4 text-text-muted font-medium">
                      Pipelines
                    </th>
                    <th className="text-left py-3 px-4 text-text-muted font-medium">
                      Created
                    </th>
                    <th className="text-right py-3 px-4 text-text-muted font-medium">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {filteredFabrics.map((fabric) => (
                    <tr
                      key={fabric.id}
                      className="border-b border-border-subtle last:border-0 hover:bg-bg-secondary/50"
                    >
                      <td className="py-3 px-4">
                        <div>
                          <p className="font-medium text-text-primary">
                            {fabric.name}
                          </p>
                          <p className="text-sm text-text-muted">
                            {fabric.description}
                          </p>
                        </div>
                      </td>
                      <td className="py-3 px-4">
                        <span className="text-sm font-mono">
                          {fabric.tenantId?.slice(0, 8)}...
                        </span>
                      </td>
                      <td className="py-3 px-4">
                        <span className="capitalize">{fabric.type}</span>
                      </td>
                      <td className="py-3 px-4">
                        <Badge className={getStatusColor(fabric.status)}>
                          {fabric.status}
                        </Badge>
                      </td>
                      <td className="py-3 px-4">{fabric.stores?.length || 0}</td>
                      <td className="py-3 px-4">
                        {fabric.pipelines?.length || 0}
                      </td>
                      <td className="py-3 px-4">
                        {new Date(fabric.createdAt).toLocaleDateString()}
                      </td>
                      <td className="py-3 px-4 text-right">
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="icon" aria-label="State fabric options">
                              <MoreVertical className="w-4 h-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem
                              onClick={() =>
                                navigate(`/state-fabric/${fabric.id}`)
                              }
                            >
                              View Details
                            </DropdownMenuItem>
                            {fabric.status === "online" ? (
                              <DropdownMenuItem
                                onClick={() => handleSuspend(fabric.id)}
                                className="text-yellow-400"
                              >
                                <Pause className="w-4 h-4 mr-2" />
                                Suspend
                              </DropdownMenuItem>
                            ) : (
                              <DropdownMenuItem
                                onClick={() => handleResume(fabric.id)}
                                className="text-green-400"
                              >
                                <Play className="w-4 h-4 mr-2" />
                                Resume
                              </DropdownMenuItem>
                            )}
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
