import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Plus, Search, MoreVertical, Users, Activity, DollarSign, ArrowLeft } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { StatCard } from "@/components/common/StatCard";
import { LoadingSpinner } from "@/components/common/LoadingSpinner";
import { tenantApi, type Tenant } from "@/api/admin";

export function AdminTenantsPage() {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [newTenantName, setNewTenantName] = useState("");

  const queryClient = useQueryClient();

  const { data: tenantsData, isLoading: tenantsLoading, error: tenantsError } = useQuery({
    queryKey: ['tenants'],
    queryFn: () => tenantApi.listTenants(),
  });

  const createTenantMutation = useMutation({
    mutationFn: (name: string) => tenantApi.createTenant(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenants'] });
      setNewTenantName("");
      setIsCreateDialogOpen(false);
    },
  });

  const updateTenantMutation = useMutation({
    mutationFn: ({ tenantId, updates }: { tenantId: string; updates: Partial<Tenant> }) =>
      tenantApi.updateTenant(tenantId, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenants'] });
    },
  });

  const tenants = tenantsData?.tenants || [];

  const filteredTenants = tenants.filter((tenant) => {
    const matchesSearch = tenant.name.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesStatus = statusFilter === "all" || tenant.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  const handleCreateTenant = () => {
    if (newTenantName.trim()) {
      createTenantMutation.mutate(newTenantName.trim());
    }
  };

  const handleSuspendTenant = (tenantId: string) => {
    const tenant = tenants.find(t => t.id === tenantId);
    if (tenant) {
      const newStatus = tenant.status === "active" ? "suspended" : "active";
      updateTenantMutation.mutate({ tenantId, updates: { status: newStatus } });
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <Button
          variant="ghost"
          onClick={() => navigate('/admin')}
          className="text-text-muted hover:text-text-primary hover:bg-bg-hover"
        >
          <ArrowLeft className="w-4 h-4 mr-2" />
          Back to Dashboard
        </Button>
        <div className="flex-1 text-center">
          <h1 className="text-2xl font-bold text-text-primary">Tenants</h1>
          <p className="text-text-secondary">Manage all tenants and their subscriptions</p>
        </div>
        <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
          <DialogTrigger asChild>
            <Button className="bg-brand-500 hover:bg-brand-600 text-white">
              <Plus className="w-4 h-4 mr-2" />
              Add Tenant
            </Button>
          </DialogTrigger>
          <DialogContent className="bg-bg-tertiary border-border-default">
            <DialogHeader>
              <DialogTitle className="text-text-primary">Create New Tenant</DialogTitle>
            </DialogHeader>
            <div className="space-y-4">
              <div>
                <Label htmlFor="tenant-name" className="text-text-primary">Tenant Name</Label>
                <Input
                  id="tenant-name"
                  placeholder="Enter tenant name"
                  value={newTenantName}
                  onChange={(e) => setNewTenantName(e.target.value)}
                  className="bg-bg-secondary border-border-default text-text-primary"
                />
              </div>
              <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={() => setIsCreateDialogOpen(false)}>
                  Cancel
                </Button>
                <Button
                  onClick={handleCreateTenant}
                  disabled={!newTenantName.trim() || createTenantMutation.isPending}
                  className="bg-brand-500 hover:bg-brand-600 text-white"
                >
                  {createTenantMutation.isPending ? "Creating..." : "Create Tenant"}
                </Button>
              </div>
            </div>
          </DialogContent>
        </Dialog>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <StatCard
          title="Total Tenants"
          value={tenants.length}
          change={{ value: 0, label: "from last month" }}
          icon={<Users className="w-5 h-5 text-brand-500" />}
          trend="neutral"
        />
        <StatCard
          title="Active Tenants"
          value={tenants.filter(t => t.status === "active").length}
          change={{ value: 0, label: "from last month" }}
          icon={<Activity className="w-5 h-5 text-brand-500" />}
          trend="neutral"
        />
        <StatCard
          title="Suspended Tenants"
          value={tenants.filter(t => t.status === "suspended").length}
          change={{ value: 0, label: "from last month" }}
          icon={<DollarSign className="w-5 h-5 text-brand-500" />}
          trend="neutral"
        />
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex flex-col sm:flex-row gap-4">
            <div className="flex-1">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-text-muted" />
                <Input
                  placeholder="Search tenants..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-10 bg-bg-secondary border-border-default text-text-primary"
                />
              </div>
            </div>
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-full sm:w-[180px] bg-bg-secondary border-border-default text-text-primary">
                <SelectValue placeholder="Filter by status" />
              </SelectTrigger>
              <SelectContent className="bg-bg-tertiary border-border-default">
                <SelectItem value="all">All Status</SelectItem>
                <SelectItem value="active">Active</SelectItem>
                <SelectItem value="suspended">Suspended</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* Tenants List */}
      <Card>
        <CardHeader>
          <CardTitle>All Tenants ({filteredTenants.length})</CardTitle>
        </CardHeader>
        <CardContent>
          {tenantsLoading ? (
            <div className="flex items-center justify-center py-8">
              <LoadingSpinner />
            </div>
          ) : tenantsError ? (
            <div className="text-center py-8">
              <p className="text-red-400">Failed to load tenants</p>
              <p className="text-text-secondary">Please try again later</p>
            </div>
          ) : (
          <div className="space-y-4">
            {filteredTenants.map((tenant) => (
              <div
                key={tenant.id}
                className="flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-border-default hover:bg-bg-hover transition-colors"
              >
                <div className="flex items-center gap-4">
                  <div className="w-10 h-10 rounded-lg bg-brand-500/10 flex items-center justify-center">
                    <Users className="w-5 h-5 text-brand-500" />
                  </div>
                  <div>
                    <p className="font-medium text-text-primary">{tenant.name}</p>
                    <p className="text-sm text-text-secondary">ID: {tenant.id.slice(0, 8)}...</p>
                  </div>
                </div>

                <div className="flex items-center gap-4">
                  <div className="text-right">
                    <Badge
                      variant={tenant.status === "active" ? "default" : "secondary"}
                      className={tenant.status === "active"
                        ? "bg-emerald-500/10 text-emerald-400"
                        : "bg-red-500/10 text-red-400"
                      }
                    >
                      {tenant.status}
                    </Badge>
                    <p className="text-xs text-text-muted mt-1">
                      Plan: {tenant.plan || "None"}
                    </p>
                  </div>

                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="sm" className="text-text-muted hover:text-text-primary">
                        <MoreVertical className="w-4 h-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent className="bg-bg-tertiary border-border-default">
                      <DropdownMenuItem className="text-text-primary hover:bg-bg-hover">
                        View Details
                      </DropdownMenuItem>
                      <DropdownMenuItem className="text-text-primary hover:bg-bg-hover">
                        Edit Tenant
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        className="text-red-400 hover:bg-red-500/10"
                        onClick={() => handleSuspendTenant(tenant.id)}
                        disabled={updateTenantMutation.isPending}
                      >
                        {updateTenantMutation.isPending ? "Updating..." : (tenant.status === "active" ? "Suspend" : "Unsuspend")}
                      </DropdownMenuItem>
                      <DropdownMenuItem className="text-red-400 hover:bg-red-500/10">
                        Delete Tenant
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </div>
            ))}
          </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
