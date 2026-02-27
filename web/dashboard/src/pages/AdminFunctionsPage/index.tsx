import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Search,
  MoreVertical,
  Code,
  Play,
  Pause,
  Trash2,
  RefreshCw,
  Eye,
  Server,
  Clock,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Loader2,
  ArrowLeft,
} from "lucide-react";
import { toast } from "sonner";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogFooter,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { Skeleton } from "@/components/ui/skeleton";
import { StatCard } from "@/components/common/StatCard";
import { LoadingSpinner } from "@/components/common/LoadingSpinner";
import {
  adminFunctionsApi,
  type AdminFunction,
  type AdminFunctionDeployment,
  type AdminFunctionLog,
  type AdminRegistryFunction,
} from "@/api/admin";
import { registryApi, type RegistryFunction } from "@/api/registry";

// Unified list item: tenant function or registry function (so admin sees "all" functions)
export type AdminFunctionListItem =
  | (AdminFunction & { source: "tenant" })
  | (Pick<AdminRegistryFunction, "id" | "name" | "author" | "visibility" | "created_at" | "updated_at"> & {
      source: "registry";
      status: "published";
      tenant_id: string;
      version: string;
      region: string;
    });

const statusColors: Record<string, string> = {
  draft: "bg-gray-500/10 text-gray-400",
  deploying: "bg-yellow-500/10 text-yellow-400",
  deployed: "bg-emerald-500/10 text-emerald-400",
  failed: "bg-red-500/10 text-red-400",
  published: "bg-emerald-500/10 text-emerald-400",
};

const statusIcons: Record<string, React.ReactNode> = {
  draft: <Clock className="w-4 h-4" />,
  deploying: <Loader2 className="w-4 h-4 animate-spin" />,
  deployed: <CheckCircle className="w-4 h-4" />,
  failed: <XCircle className="w-4 h-4" />,
  published: <CheckCircle className="w-4 h-4" />,
};

export function AdminFunctionsPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [searchTerm, setSearchTerm] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [selectedFunction, setSelectedFunction] = useState<AdminFunctionListItem | null>(null);
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [functionToDelete, setFunctionToDelete] = useState<AdminFunctionListItem | null>(null);

  // Fetch tenant functions (deployed per-tenant)
  const { data: tenantData, isLoading: tenantLoading, refetch: refetchTenant } = useQuery({
    queryKey: ["admin-functions", statusFilter, searchTerm],
    queryFn: () =>
      adminFunctionsApi.listFunctions({
        status: statusFilter !== "all" ? statusFilter : undefined,
        search: searchTerm || undefined,
        limit: 100,
      }),
  });

  // Fetch registry functions via public API (GET /v1/registry/functions); admin registry endpoint may not exist
  const { data: registryData, isLoading: registryLoading, refetch: refetchRegistry } = useQuery({
    queryKey: ["registry-functions-for-admin-list"],
    queryFn: async () => {
      const res = await registryApi.getFunctions({ limit: 500 });
      return { functions: res.functions ?? [], total: (res as { total?: number }).total ?? res.functions?.length ?? 0 };
    },
  });

  const refetch = () => {
    refetchTenant();
    refetchRegistry();
  };

  // Merge into one list: tenant functions + registry functions (normalized to same shape for display)
  const tenantFunctions: AdminFunctionListItem[] = (tenantData?.functions || []).map((f) => ({
    ...f,
    source: "tenant" as const,
  }));
  const registryAsListItems: AdminFunctionListItem[] = (registryData?.functions || []).map((f: RegistryFunction & { version?: string }) => ({
    id: f.id || `registry-${f.author}-${f.name}`,
    name: f.name,
    author: f.author,
    visibility: (f.visibility ?? "public") as "public" | "private" | "unlisted",
    created_at: f.created_at ?? new Date().toISOString(),
    updated_at: (f as { updated_at?: string }).updated_at ?? f.created_at ?? new Date().toISOString(),
    source: "registry" as const,
    status: "published" as const,
    tenant_id: "",
    version: f.latest_version ?? f.version ?? "1.0.0",
    region: "global",
  }));
  const allFunctions: AdminFunctionListItem[] = [...tenantFunctions, ...registryAsListItems];
  const functionsData = {
    functions: allFunctions,
    total: (tenantData?.total ?? 0) + (registryData?.total ?? 0),
  };
  const functionsLoading = tenantLoading || registryLoading;

  // Fetch deployments for selected function (tenant functions only)
  const { data: deploymentsData, isLoading: deploymentsLoading } = useQuery({
    queryKey: ["admin-function-deployments", selectedFunction?.id],
    queryFn: () => adminFunctionsApi.getFunctionDeployments(selectedFunction!.id),
    enabled: !!selectedFunction && selectedFunction.source === "tenant",
  });

  // Fetch logs for selected function (tenant functions only)
  const { data: logsData, isLoading: logsLoading } = useQuery({
    queryKey: ["admin-function-logs", selectedFunction?.id],
    queryFn: () => adminFunctionsApi.getFunctionLogs(selectedFunction!.id, { limit: 50 }),
    enabled: !!selectedFunction && selectedFunction.source === "tenant",
  });

  // Toggle function status mutation
  const toggleStatusMutation = useMutation({
    mutationFn: ({ functionId, enabled }: { functionId: string; enabled: boolean }) =>
      adminFunctionsApi.toggleFunctionStatus(functionId, enabled),
    onSuccess: () => {
      toast.success("Function status updated");
      queryClient.invalidateQueries({ queryKey: ["admin-functions"] });
      queryClient.invalidateQueries({ queryKey: ["registry-functions-for-admin-list"] });
    },
    onError: () => {
      toast.error("Failed to update function status");
    },
  });

  // Delete function mutation
  const deleteFunctionMutation = useMutation({
    mutationFn: (functionId: string) => adminFunctionsApi.deleteFunction(functionId),
    onSuccess: () => {
      toast.success("Function deleted successfully");
      queryClient.invalidateQueries({ queryKey: ["admin-functions"] });
      queryClient.invalidateQueries({ queryKey: ["registry-functions-for-admin-list"] });
      setIsDeleteDialogOpen(false);
      setFunctionToDelete(null);
    },
    onError: () => {
      toast.error("Failed to delete function");
    },
  });

  const functions = functionsData?.functions || [];
  const totalFunctions = functionsData?.total || 0;

  const filteredFunctions = functions.filter((fn) => {
    const searchLower = searchTerm?.toLowerCase() ?? "";
    const matchesSearch =
      !searchTerm ||
      fn.name.toLowerCase().includes(searchLower) ||
      (fn.source === "tenant" && "tenant_name" in fn && fn.tenant_name?.toLowerCase().includes(searchLower)) ||
      (fn.source === "tenant" && fn.tenant_id?.toLowerCase().includes(searchLower)) ||
      (fn.source === "registry" && "author" in fn && fn.author?.toLowerCase().includes(searchLower));
    const matchesStatus =
      statusFilter === "all" ||
      fn.status === statusFilter ||
      (statusFilter === "deployed" && fn.status === "published");
    return matchesSearch && matchesStatus;
  });

  const stats = {
    total: totalFunctions,
    deployed:
      functions.filter((f) => f.status === "deployed").length +
      functions.filter((f) => f.status === "published").length,
    failed: functions.filter((f) => f.status === "failed").length,
    draft: functions.filter((f) => f.status === "draft").length,
  };

  const handleViewDetails = (fn: AdminFunctionListItem) => {
    setSelectedFunction(fn);
    setIsDetailsOpen(true);
  };

  const handleToggleStatus = (fn: AdminFunctionListItem) => {
    if (fn.source === "registry") return;
    const newStatus = fn.status === "deployed" ? "draft" : "deployed";
    toggleStatusMutation.mutate({ functionId: fn.id, enabled: newStatus === "deployed" });
  };

  const handleDeleteClick = (fn: AdminFunctionListItem) => {
    if (fn.source === "registry") return;
    setFunctionToDelete(fn);
    setIsDeleteDialogOpen(true);
  };

  const confirmDelete = () => {
    if (functionToDelete && functionToDelete.source === "tenant") {
      deleteFunctionMutation.mutate(functionToDelete.id);
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
          <h1 className="text-2xl font-bold text-text-primary">Functions</h1>
          <p className="text-text-secondary">
            Manage all functions across all tenants
          </p>
        </div>
        <Button
          onClick={() => refetch()}
          variant="outline"
          className="border-border-subtle hover:bg-bg-hover"
        >
          <RefreshCw className="w-4 h-4 mr-2" />
          Refresh
        </Button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <StatCard
          title="Total Functions"
          value={stats.total}
          icon={<Code className="w-5 h-5 text-[#6366f1]" />}
          trend="neutral"
          change={{ value: 0, label: "total" }}
        />
        <StatCard
          title="Deployed"
          value={stats.deployed}
          icon={<CheckCircle className="w-5 h-5 text-emerald-400" />}
          trend="up"
          change={{ value: 0, label: "active" }}
        />
        <StatCard
          title="Failed"
          value={stats.failed}
          icon={<XCircle className="w-5 h-5 text-red-400" />}
          trend="down"
          change={{ value: 0, label: "needs attention" }}
        />
        <StatCard
          title="Draft"
          value={stats.draft}
          icon={<Clock className="w-5 h-5 text-gray-400" />}
          trend="neutral"
          change={{ value: 0, label: "not deployed" }}
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
                  placeholder="Search functions by name or tenant..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-10"
                />
              </div>
            </div>
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-full sm:w-[180px]">
                <SelectValue placeholder="Filter by status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Status</SelectItem>
                <SelectItem value="deployed">Deployed</SelectItem>
                <SelectItem value="published">Published (Registry)</SelectItem>
                <SelectItem value="draft">Draft</SelectItem>
                <SelectItem value="deploying">Deploying</SelectItem>
                <SelectItem value="failed">Failed</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* Functions List */}
      <Card>
        <CardHeader>
          <CardTitle className="text-text-primary">
            All Functions ({filteredFunctions.length})
          </CardTitle>
        </CardHeader>
        <CardContent>
          {functionsLoading ? (
            <div className="space-y-4">
              {Array.from({ length: 5 }).map((_, i) => (
                <Skeleton key={i} className="h-20 w-full" />
              ))}
            </div>
          ) : filteredFunctions.length === 0 ? (
            <div className="text-center py-12">
              <Code className="h-12 w-12 text-text-muted mx-auto mb-4" />
              <h3 className="text-lg font-semibold text-text-primary mb-2">
                No functions found
              </h3>
              <p className="text-text-secondary">
                {searchTerm
                  ? "Try adjusting your search or filters"
                  : "No functions exist in the system yet"}
              </p>
            </div>
          ) : (
            <div className="space-y-4">
              {filteredFunctions.map((fn) => (
                <div
                  key={fn.id}
                  className="flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-border-subtle hover:bg-bg-hover transition-colors"
                >
                  <div className="flex items-center gap-4">
                    <div className="w-10 h-10 rounded-lg bg-[#6366f1]/10 flex items-center justify-center">
                      <Code className="w-5 h-5 text-[#6366f1]" />
                    </div>
                    <div>
                      <div className="flex items-center gap-2">
                        <p className="font-medium text-text-primary">{fn.name}</p>
                        <Badge variant="outline" className="text-xs border-border-subtle text-text-muted">
                          {fn.source === "tenant" ? "Tenant" : "Registry"}
                        </Badge>
                      </div>
                      <p className="text-sm text-text-muted">
                        {fn.source === "tenant"
                          ? `Tenant: ${"tenant_name" in fn && fn.tenant_name ? fn.tenant_name : fn.tenant_id?.slice(0, 8) + "..."}`
                          : "author" in fn
                            ? `Author: ${fn.author}`
                            : ""}
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center gap-4">
                    <div className="text-right">
                      <Badge
                        className={statusColors[fn.status] ?? "bg-gray-500/10 text-gray-400"}
                      >
                        {statusIcons[fn.status]}
                        <span className="ml-1">{fn.status}</span>
                      </Badge>
                      <p className="text-xs text-text-muted mt-1">
                        v{fn.version} • {fn.region}
                      </p>
                    </div>

                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-text-muted hover:text-text-primary"
                        >
                          <MoreVertical className="w-4 h-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent className="bg-bg-tertiary border-border-subtle">
                        <DropdownMenuItem
                          className="text-text-primary hover:bg-bg-hover"
                          onClick={() => handleViewDetails(fn)}
                        >
                          <Eye className="w-4 h-4 mr-2" />
                          View Details
                        </DropdownMenuItem>
                        {fn.source === "tenant" && (
                          <>
                            <DropdownMenuItem
                              className="text-text-primary hover:bg-bg-hover"
                              onClick={() => handleToggleStatus(fn)}
                            >
                              {fn.status === "deployed" ? (
                                <>
                                  <Pause className="w-4 h-4 mr-2" />
                                  Disable
                                </>
                              ) : (
                                <>
                                  <Play className="w-4 h-4 mr-2" />
                                  Enable
                                </>
                              )}
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              className="text-red-400 hover:bg-red-500/10"
                              onClick={() => handleDeleteClick(fn)}
                            >
                              <Trash2 className="w-4 h-4 mr-2" />
                              Delete
                            </DropdownMenuItem>
                          </>
                        )}
                        {fn.source === "registry" && (
                          <DropdownMenuItem
                            className="text-text-primary hover:bg-bg-hover"
                            onClick={() => navigate("/admin/registry")}
                          >
                            <Server className="w-4 h-4 mr-2" />
                            Manage in Registry
                          </DropdownMenuItem>
                        )}
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Function Details Dialog */}
      <Dialog open={isDetailsOpen} onOpenChange={setIsDetailsOpen}>
        <DialogContent className="bg-bg-tertiary border-border-subtle max-w-4xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="text-text-primary flex items-center gap-2">
              <Code className="w-5 h-5 text-[#6366f1]" />
              {selectedFunction?.name}
            </DialogTitle>
          </DialogHeader>

          {selectedFunction && (
            <Tabs defaultValue="details" className="w-full">
              <TabsList className="bg-bg-secondary">
                <TabsTrigger value="details" className="text-text-primary">
                  Details
                </TabsTrigger>
                {selectedFunction.source === "tenant" && (
                  <>
                    <TabsTrigger value="deployments" className="text-text-primary">
                      Deployments
                    </TabsTrigger>
                    <TabsTrigger value="logs" className="text-text-primary">
                      Logs
                    </TabsTrigger>
                  </>
                )}
              </TabsList>

              <TabsContent value="details" className="space-y-4 mt-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label className="text-text-muted">Function ID</Label>
                    <p className="text-text-primary font-mono text-sm">
                      {selectedFunction.id}
                    </p>
                  </div>
                  <div>
                    <Label className="text-text-muted">Status</Label>
                    <p className="text-text-primary">
                      <Badge
                        className={statusColors[selectedFunction.status] ?? "bg-gray-500/10 text-gray-400"}
                      >
                        {selectedFunction.status}
                      </Badge>
                    </p>
                  </div>
                  {selectedFunction.source === "tenant" ? (
                    <>
                      <div>
                        <Label className="text-text-muted">Tenant</Label>
                        <p className="text-text-primary">
                          {"tenant_name" in selectedFunction && selectedFunction.tenant_name
                            ? selectedFunction.tenant_name
                            : selectedFunction.tenant_id}
                        </p>
                      </div>
                      <div>
                        <Label className="text-text-muted">Region</Label>
                        <p className="text-text-primary">{selectedFunction.region}</p>
                      </div>
                      <div>
                        <Label className="text-text-muted">Version</Label>
                        <p className="text-text-primary">{selectedFunction.version}</p>
                      </div>
                      <div>
                        <Label className="text-text-muted">Providers</Label>
                        <p className="text-text-primary">
                          {"providers" in selectedFunction && selectedFunction.providers?.join(", ")}
                        </p>
                      </div>
                    </>
                  ) : (
                    <>
                      <div>
                        <Label className="text-text-muted">Author</Label>
                        <p className="text-text-primary">
                          {"author" in selectedFunction ? selectedFunction.author : "—"}
                        </p>
                      </div>
                      <div>
                        <Label className="text-text-muted">Visibility</Label>
                        <p className="text-text-primary">
                          {"visibility" in selectedFunction ? selectedFunction.visibility : "—"}
                        </p>
                      </div>
                      <div>
                        <Label className="text-text-muted">Version</Label>
                        <p className="text-text-primary">{selectedFunction.version}</p>
                      </div>
                      <div className="col-span-2">
                        <Button
                          variant="outline"
                          className="border-border-subtle"
                          onClick={() => {
                            setIsDetailsOpen(false);
                            navigate("/admin/registry");
                          }}
                        >
                          <Server className="w-4 h-4 mr-2" />
                          Manage in Registry
                        </Button>
                      </div>
                    </>
                  )}
                  <div>
                    <Label className="text-text-muted">Created At</Label>
                    <p className="text-text-primary">
                      {new Date(selectedFunction.created_at).toLocaleString()}
                    </p>
                  </div>
                  <div>
                    <Label className="text-text-muted">Updated At</Label>
                    <p className="text-text-primary">
                      {new Date(selectedFunction.updated_at).toLocaleString()}
                    </p>
                  </div>
                </div>

                {selectedFunction.source === "tenant" && "env_vars" in selectedFunction && (
                  <div>
                    <Label className="text-text-muted">Environment Variables</Label>
                    {selectedFunction.env_vars && selectedFunction.env_vars.length > 0 ? (
                      <div className="mt-2 space-y-2">
                        {selectedFunction.env_vars.map((env, i) => (
                          <div
                            key={i}
                            className="flex justify-between p-2 bg-bg-secondary rounded text-sm"
                          >
                            <span className="text-text-primary font-mono">{env.key}</span>
                            <span className="text-text-muted">
                              {env.is_secret ? "••••••••" : env.value}
                            </span>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <p className="text-text-muted mt-2">No environment variables</p>
                    )}
                  </div>
                )}

                {selectedFunction.source === "tenant" && "code" in selectedFunction && (
                  <div>
                    <Label className="text-text-muted">Code Preview</Label>
                    <Textarea
                      readOnly
                      value={selectedFunction.code}
                      className="mt-2 h-48 font-mono text-sm"
                    />
                  </div>
                )}
              </TabsContent>

              {selectedFunction.source === "tenant" && (
              <>
              <TabsContent value="deployments" className="mt-4">
                {deploymentsLoading ? (
                  <LoadingSpinner />
                ) : deploymentsData?.deployments && deploymentsData.deployments.length > 0 ? (
                  <div className="space-y-3">
                    {deploymentsData.deployments.map((deployment) => (
                      <div
                        key={deployment.id}
                        className="p-4 bg-bg-secondary rounded-lg border border-border-subtle"
                      >
                        <div className="flex justify-between items-start">
                          <div>
                            <p className="text-text-primary font-medium">
                              v{deployment.version}
                            </p>
                            <p className="text-sm text-text-muted">
                              {deployment.provider} • {deployment.region}
                            </p>
                          </div>
                          <Badge
                            className={
                              deployment.status === "success"
                                ? "bg-emerald-500/10 text-emerald-400"
                                : deployment.status === "failed"
                                ? "bg-red-500/10 text-red-400"
                                : "bg-yellow-500/10 text-yellow-400"
                            }
                          >
                            {deployment.status}
                          </Badge>
                        </div>
                        {deployment.deployed_url && (
                          <p className="text-sm text-text-muted mt-2">
                            URL: {deployment.deployed_url}
                          </p>
                        )}
                        {deployment.error_message && (
                          <p className="text-sm text-red-400 mt-2">
                            Error: {deployment.error_message}
                          </p>
                        )}
                        <p className="text-xs text-text-muted mt-2">
                          {new Date(deployment.created_at).toLocaleString()}
                        </p>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-text-muted text-center py-8">
                    No deployments found
                  </p>
                )}
              </TabsContent>

              <TabsContent value="logs" className="mt-4">
                {logsLoading ? (
                  <LoadingSpinner />
                ) : logsData?.logs && logsData.logs.length > 0 ? (
                  <div className="space-y-2 max-h-96 overflow-y-auto">
                    {logsData.logs.map((log) => (
                      <div
                        key={log.id}
                        className={`p-3 rounded-lg border ${
                          log.level === "error"
                            ? "bg-red-500/5 border-red-500/20"
                            : log.level === "warn"
                            ? "bg-yellow-500/5 border-yellow-500/20"
                            : "bg-bg-secondary border-border-subtle"
                        }`}
                      >
                        <div className="flex justify-between items-start">
                          <Badge
                            variant="outline"
                            className={
                              log.level === "error"
                                ? "text-red-400 border-red-400/30"
                                : log.level === "warn"
                                ? "text-yellow-400 border-yellow-400/30"
                                : "text-blue-400 border-blue-400/30"
                            }
                          >
                            {log.level}
                          </Badge>
                          <span className="text-xs text-text-muted">
                            {new Date(log.timestamp).toLocaleString()}
                          </span>
                        </div>
                        <p className="text-sm text-text-primary mt-2">{log.message}</p>
                        {log.metadata && (
                          <pre className="text-xs text-text-muted mt-2 overflow-x-auto">
                            {JSON.stringify(log.metadata, null, 2)}
                          </pre>
                        )}
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-text-muted text-center py-8">No logs found</p>
                )}
              </TabsContent>
              </>
              )}
            </Tabs>
          )}
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={isDeleteDialogOpen} onOpenChange={setIsDeleteDialogOpen}>
        <DialogContent className="bg-bg-tertiary border-border-subtle">
          <DialogHeader>
            <DialogTitle className="text-text-primary flex items-center gap-2">
              <AlertTriangle className="w-5 h-5 text-red-400" />
              Delete Function
            </DialogTitle>
          </DialogHeader>
          <p className="text-text-secondary">
            Are you sure you want to delete{" "}
            <span className="text-text-primary font-medium">{functionToDelete?.name}</span>?
            This action cannot be undone.
          </p>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsDeleteDialogOpen(false)}
              className="border-border-subtle text-text-primary hover:bg-bg-hover"
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={confirmDelete}
              disabled={deleteFunctionMutation.isPending}
              className="bg-red-600 hover:bg-red-700"
            >
              {deleteFunctionMutation.isPending ? "Deleting..." : "Delete"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
