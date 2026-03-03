import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  ArrowLeft,
  RefreshCw,
  Cloud,
  CheckCircle,
  XCircle,
  AlertTriangle,
  Users,
  Building2,
  MoreVertical,
  Share2,
  Trash2,
  Zap,
  Server,
  Globe,
  Activity,
  BarChart3,
  Code,
  TrendingUp,
  Shield,
  Database,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { Progress } from "@/components/ui/progress";
import { adminProvidersApi, type AdminProvider, backendsApi } from "@/api/admin";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

// Mock data for FunctionFly Edge - in a real implementation, this would come from API
const mockEdgeData = {
  infrastructure: {
    regions: [
      { id: "ff-us-east", name: "US East", status: "healthy", backends: 3, load: 65 },
      { id: "ff-us-west", name: "US West", status: "healthy", backends: 2, load: 45 },
      { id: "ff-eu-west", name: "Europe West", status: "degraded", backends: 2, load: 80 },
      { id: "ff-apac", name: "Asia Pacific", status: "healthy", backends: 1, load: 30 },
    ],
    totalBackends: 8,
    activeFunctions: 1247,
    totalRequests: 2456789,
    avgLatency: 45,
  },
  functions: [
    {
      id: "func-1",
      name: "user-auth-api",
      author: "platform",
      region: "ff-us-east",
      status: "active",
      requests24h: 45230,
      avgLatency: 23,
      errorRate: 0.1,
    },
    {
      id: "func-2",
      name: "data-processor",
      author: "platform",
      region: "ff-eu-west",
      status: "active",
      requests24h: 12890,
      avgLatency: 67,
      errorRate: 0.3,
    },
    {
      id: "func-3",
      name: "image-optimizer",
      author: "platform",
      region: "ff-us-west",
      status: "degraded",
      requests24h: 8934,
      avgLatency: 120,
      errorRate: 2.1,
    },
  ],
};

export function AdminProvidersPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [activeTab, setActiveTab] = useState("user-providers");

  // Fetch providers data
  const {
    data: providersData,
    isLoading: providersLoading,
    error: providersError,
    refetch,
  } = useQuery({
    queryKey: ["admin-providers"],
    queryFn: () => adminProvidersApi.listProviders(),
  });

  // Fetch FunctionFly Edge backends
  const { data: edgeBackendsData, isLoading: edgeBackendsLoading } = useQuery({
    queryKey: ["admin-platform-backends"],
    queryFn: () => backendsApi.listPlatformBackends(),
    select: (data) => ({
      ...data,
      backends: data.backends.filter(b => b.provider === "functionfly-edge"),
    }),
  });

  // Update provider mutation
  const updateProviderMutation = useMutation({
    mutationFn: ({ providerId, updates }: {
      providerId: string;
      updates: Parameters<typeof adminProvidersApi.updateProvider>[1]
    }) =>
      adminProvidersApi.updateProvider(providerId, updates),
    onSuccess: (updatedProvider) => {
      // Update the cache with the new provider data
      queryClient.setQueryData(
        ["admin-providers"],
        (oldData: { providers: AdminProvider[] } | undefined) => {
          if (!oldData) return oldData;
          return {
            providers: oldData.providers.map((provider) =>
              provider.id === updatedProvider.id ? updatedProvider : provider
            ),
          };
        }
      );
      toast.success("Provider updated successfully");
    },
    onError: (error) => {
      toast.error("Failed to update provider");
      console.error("Failed to update provider:", error);
    },
  });

  // Delete provider mutation
  const deleteProviderMutation = useMutation({
    mutationFn: (providerId: string) =>
      adminProvidersApi.deleteProvider(providerId),
    onSuccess: (_, providerId) => {
      // Remove from cache
      queryClient.setQueryData(
        ["admin-providers"],
        (oldData: { providers: AdminProvider[] } | undefined) => {
          if (!oldData) return oldData;
          return {
            providers: oldData.providers.filter((provider) => provider.id !== providerId),
          };
        }
      );
      toast.success("Provider deactivated successfully");
    },
    onError: (error) => {
      toast.error("Failed to deactivate provider");
      console.error("Failed to deactivate provider:", error);
    },
  });

  const handleRefresh = async () => {
    setIsRefreshing(true);
    await refetch();
    setIsRefreshing(false);
  };

  const handleToggleStatus = async (provider: AdminProvider) => {
    const newStatus = provider.status === 'active' ? 'inactive' : 'active';
    try {
      await updateProviderMutation.mutateAsync({
        providerId: provider.id,
        updates: { status: newStatus },
      });
    } catch (error) {
      // Error is handled in the mutation
    }
  };

  const handleToggleShared = async (provider: AdminProvider) => {
    try {
      await updateProviderMutation.mutateAsync({
        providerId: provider.id,
        updates: { is_shared: !provider.is_shared },
      });
    } catch (error) {
      // Error is handled in the mutation
    }
  };

  const handleDeleteProvider = async (provider: AdminProvider) => {
    if (confirm(`Are you sure you want to deactivate the ${provider.provider} provider for ${provider.user_email}?`)) {
      try {
        await deleteProviderMutation.mutateAsync(provider.id);
      } catch (error) {
        // Error is handled in the mutation
      }
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'active':
        return <CheckCircle className="h-4 w-4 text-emerald-600 dark:text-emerald-400" />;
      case 'inactive':
        return <XCircle className="h-4 w-4 text-red-600 dark:text-red-400" />;
      case 'error':
        return <AlertTriangle className="h-4 w-4 text-amber-600 dark:text-amber-400" />;
      default:
        return <XCircle className="h-4 w-4 text-gray-600 dark:text-gray-400" />;
    }
  };

  const getStatusBadge = (status: string) => {
    const config = {
      active: {
        label: "Active",
        className: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20"
      },
      inactive: {
        label: "Inactive",
        className: "bg-red-500/10 text-red-600 dark:text-red-400 border-red-500/20"
      },
      error: {
        label: "Error",
        className: "bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20"
      },
    };

    const statusConfig = config[status as keyof typeof config] || config.inactive;

    return (
      <Badge
        variant="outline"
        className={cn("flex items-center gap-1", statusConfig.className)}
      >
        {getStatusIcon(status)}
        {statusConfig.label}
      </Badge>
    );
  };

  const getProviderIcon = (provider: string) => {
    // You could add specific icons for each provider here
    return <Cloud className="h-4 w-4" />;
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const formatNumber = (num: number) => {
    return new Intl.NumberFormat().format(num);
  };

  const getEdgeStatusBadge = (status: string) => {
    const config = {
      healthy: { label: "Healthy", className: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20" },
      degraded: { label: "Degraded", className: "bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20" },
      unhealthy: { label: "Unhealthy", className: "bg-red-500/10 text-red-600 dark:text-red-400 border-red-500/20" },
    };

    const statusConfig = config[status as keyof typeof config] || config.healthy;

    return (
      <Badge variant="outline" className={cn("flex items-center gap-1", statusConfig.className)}>
        {status === "healthy" && <CheckCircle className="h-3 w-3" />}
        {status === "degraded" && <AlertTriangle className="h-3 w-3" />}
        {status === "unhealthy" && <XCircle className="h-3 w-3" />}
        {statusConfig.label}
      </Badge>
    );
  };

  const getLoadColor = (load: number) => {
    if (load < 50) return "text-emerald-600";
    if (load < 80) return "text-amber-600";
    return "text-red-600";
  };

  if (providersLoading) {
    return (
      <div className="container mx-auto px-6 py-8">
        <div className="flex items-center gap-4 mb-6">
          <Skeleton className="h-10 w-10" />
          <div>
            <Skeleton className="h-8 w-64 mb-2" />
            <Skeleton className="h-4 w-96" />
          </div>
        </div>

        <Card>
          <CardHeader>
            <Skeleton className="h-6 w-48" />
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="flex items-center justify-between p-4 border rounded-lg">
                  <div className="flex items-center gap-4">
                    <Skeleton className="h-8 w-8" />
                    <div>
                      <Skeleton className="h-4 w-32 mb-1" />
                      <Skeleton className="h-3 w-24" />
                    </div>
                  </div>
                  <Skeleton className="h-6 w-16" />
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (providersError) {
    return (
      <div className="container mx-auto px-6 py-8">
        <div className="flex items-center gap-4 mb-6">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => navigate("/admin")}
            className="shrink-0"
          >
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <h1 className="text-2xl font-bold">Provider Management</h1>
            <p className="text-muted-foreground">
              Manage user provider connections and access
            </p>
          </div>
        </div>

        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>
            Failed to load providers. Please try again.
          </AlertDescription>
        </Alert>

        <div className="mt-4">
          <Button onClick={handleRefresh} disabled={isRefreshing}>
            <RefreshCw className={cn("h-4 w-4 mr-2", isRefreshing && "animate-spin")} />
            Retry
          </Button>
        </div>
      </div>
    );
  }

  const providers = providersData?.providers || [];
  const activeCount = providers.filter((p) => p.status === 'active').length;
  const inactiveCount = providers.filter((p) => p.status === 'inactive').length;
  const errorCount = providers.filter((p) => p.status === 'error').length;
  const sharedCount = providers.filter((p) => p.is_shared).length;

  return (
    <div className="container mx-auto px-6 py-8">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-4">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => navigate("/admin")}
            className="shrink-0"
          >
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <h1 className="text-2xl font-bold">Provider Management</h1>
            <p className="text-muted-foreground">
              Manage user provider connections and FunctionFly Edge infrastructure
            </p>
          </div>
        </div>

        <Button
          variant="outline"
          size="sm"
          onClick={handleRefresh}
          disabled={isRefreshing}
        >
          <RefreshCw className={cn("h-4 w-4 mr-2", isRefreshing && "animate-spin")} />
          Refresh
        </Button>
      </div>

      {/* Main Content Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="user-providers">User Providers</TabsTrigger>
          <TabsTrigger value="functionfly-edge">FunctionFly Edge</TabsTrigger>
        </TabsList>

        {/* User Providers Tab */}
        <TabsContent value="user-providers" className="space-y-6">
          {/* Stats Cards */}
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
            <Card>
              <CardContent className="flex items-center gap-4 p-6">
                <div className="p-2 bg-blue-500/10 rounded-lg">
                  <Cloud className="h-6 w-6 text-blue-600 dark:text-blue-400" />
                </div>
                <div>
                  <div className="text-2xl font-bold">{providers.length}</div>
                  <div className="text-sm text-muted-foreground">Total Providers</div>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="flex items-center gap-4 p-6">
                <div className="p-2 bg-emerald-500/10 rounded-lg">
                  <CheckCircle className="h-6 w-6 text-emerald-600 dark:text-emerald-400" />
                </div>
                <div>
                  <div className="text-2xl font-bold">{activeCount}</div>
                  <div className="text-sm text-muted-foreground">Active</div>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="flex items-center gap-4 p-6">
                <div className="p-2 bg-amber-500/10 rounded-lg">
                  <AlertTriangle className="h-6 w-6 text-amber-600 dark:text-amber-400" />
                </div>
                <div>
                  <div className="text-2xl font-bold">{errorCount}</div>
                  <div className="text-sm text-muted-foreground">With Errors</div>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="flex items-center gap-4 p-6">
                <div className="p-2 bg-purple-500/10 rounded-lg">
                  <Share2 className="h-6 w-6 text-purple-600 dark:text-purple-400" />
                </div>
                <div>
                  <div className="text-2xl font-bold">{sharedCount}</div>
                  <div className="text-sm text-muted-foreground">Shared</div>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Providers Table */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Cloud className="h-5 w-5" />
                Provider Connections
              </CardTitle>
            </CardHeader>
            <CardContent>
          {providers.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <Cloud className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No provider connections found</p>
            </div>
          ) : (
            <div className="space-y-4">
              {providers.map((provider) => (
                <div
                  key={provider.id}
                  className="flex items-center justify-between p-4 border rounded-lg hover:bg-muted/50 transition-colors"
                >
                  <div className="flex items-center gap-4">
                    <div className="p-2 bg-muted rounded-lg">
                      {getProviderIcon(provider.provider)}
                    </div>

                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2 mb-1">
                        <span className="font-medium capitalize">
                          {provider.provider}
                        </span>
                        {provider.is_shared && (
                          <Badge variant="outline" className="text-xs">
                            <Share2 className="h-3 w-3 mr-1" />
                            Shared
                          </Badge>
                        )}
                      </div>

                      <div className="flex items-center gap-4 text-sm text-muted-foreground">
                        <span className="flex items-center gap-1">
                          <Users className="h-3 w-3" />
                          {provider.user_email}
                        </span>
                        <span className="flex items-center gap-1">
                          <Building2 className="h-3 w-3" />
                          {provider.tenant_name}
                        </span>
                      </div>

                      <div className="text-xs text-muted-foreground mt-1">
                        Connected {formatDate(provider.created_at)}
                        {provider.updated_at !== provider.created_at && (
                          <> • Updated {formatDate(provider.updated_at)}</>
                        )}
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center gap-4">
                    {getStatusBadge(provider.status)}

                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="sm">
                          <MoreVertical className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem
                          onClick={() => handleToggleStatus(provider)}
                          disabled={updateProviderMutation.isPending}
                        >
                          {provider.status === 'active' ? 'Deactivate' : 'Activate'}
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          onClick={() => handleToggleShared(provider)}
                          disabled={updateProviderMutation.isPending}
                        >
                          {provider.is_shared ? 'Unshare' : 'Share'} with Team
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          onClick={() => handleDeleteProvider(provider)}
                          disabled={deleteProviderMutation.isPending}
                          className="text-red-600 dark:text-red-400"
                        >
                          <Trash2 className="h-4 w-4 mr-2" />
                          Deactivate Provider
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
        </TabsContent>

        {/* FunctionFly Edge Tab */}
        <TabsContent value="functionfly-edge" className="space-y-6">
          {/* Edge Stats Overview */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
            <Card>
              <CardContent className="flex items-center gap-4 p-6">
                <div className="p-2 bg-purple-500/10 rounded-lg">
                  <Server className="h-6 w-6 text-purple-600 dark:text-purple-400" />
                </div>
                <div>
                  <div className="text-2xl font-bold">{mockEdgeData.infrastructure.totalBackends}</div>
                  <div className="text-sm text-muted-foreground">Active Backends</div>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="flex items-center gap-4 p-6">
                <div className="p-2 bg-blue-500/10 rounded-lg">
                  <Code className="h-6 w-6 text-blue-600 dark:text-blue-400" />
                </div>
                <div>
                  <div className="text-2xl font-bold">{formatNumber(mockEdgeData.infrastructure.activeFunctions)}</div>
                  <div className="text-sm text-muted-foreground">Active Functions</div>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="flex items-center gap-4 p-6">
                <div className="p-2 bg-emerald-500/10 rounded-lg">
                  <Activity className="h-6 w-6 text-emerald-600 dark:text-emerald-400" />
                </div>
                <div>
                  <div className="text-2xl font-bold">{formatNumber(mockEdgeData.infrastructure.totalRequests)}</div>
                  <div className="text-sm text-muted-foreground">Requests (24h)</div>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="flex items-center gap-4 p-6">
                <div className="p-2 bg-amber-500/10 rounded-lg">
                  <Zap className="h-6 w-6 text-amber-600 dark:text-amber-400" />
                </div>
                <div>
                  <div className="text-2xl font-bold">{mockEdgeData.infrastructure.avgLatency}ms</div>
                  <div className="text-sm text-muted-foreground">Avg Latency</div>
                </div>
              </CardContent>
            </Card>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Regional Status */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Globe className="h-5 w-5" />
                  Regional Status
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                {mockEdgeData.infrastructure.regions.map((region) => (
                  <div key={region.id} className="flex items-center justify-between p-3 border rounded-lg">
                    <div className="flex items-center gap-3">
                      <div className="w-2 h-2 rounded-full bg-emerald-500" />
                      <div>
                        <div className="font-medium">{region.name}</div>
                        <div className="text-sm text-muted-foreground">
                          {region.backends} backends • {region.load}% load
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      {getEdgeStatusBadge(region.status)}
                      <div className={cn("text-sm font-medium", getLoadColor(region.load))}>
                        {region.load}%
                      </div>
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>

            {/* Edge Backends */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Server className="h-5 w-5" />
                  Edge Backends
                </CardTitle>
              </CardHeader>
              <CardContent>
                {edgeBackendsLoading ? (
                  <div className="space-y-4">
                    {Array.from({ length: 3 }).map((_, i) => (
                      <div key={i} className="flex items-center justify-between p-4 border rounded-lg">
                        <div className="flex items-center gap-4">
                          <Skeleton className="h-8 w-8" />
                          <div>
                            <Skeleton className="h-4 w-32 mb-1" />
                            <Skeleton className="h-3 w-24" />
                          </div>
                        </div>
                        <Skeleton className="h-6 w-16" />
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="space-y-4">
                    {edgeBackendsData?.backends.map((backend) => (
                      <div
                        key={backend.id}
                        className="flex items-center justify-between p-4 border rounded-lg hover:bg-muted/50 transition-colors"
                      >
                        <div className="flex items-center gap-4">
                          <div className="p-2 bg-muted rounded-lg">
                            <Server className="h-4 w-4" />
                          </div>
                          <div>
                            <div className="font-medium">
                              {backend.app_name} - {backend.region}
                            </div>
                            <div className="text-sm text-muted-foreground">
                              {backend.url} • Updated {new Date(backend.updated_at).toLocaleDateString()}
                            </div>
                          </div>
                        </div>
                        <div className="flex items-center gap-2">
                          {getEdgeStatusBadge(backend.enabled ? "healthy" : "unhealthy")}
                          <Badge variant="outline">
                            {backend.priority ? `Priority ${backend.priority}` : "Standard"}
                          </Badge>
                        </div>
                      </div>
                    ))}
                    {(!edgeBackendsData?.backends || edgeBackendsData.backends.length === 0) && (
                      <div className="text-center py-8 text-muted-foreground">
                        <Server className="h-12 w-12 mx-auto mb-4 opacity-50" />
                        <p>No FunctionFly Edge backends found</p>
                      </div>
                    )}
                  </div>
                )}
              </CardContent>
            </Card>
          </div>

          {/* Platform Functions */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Code className="h-5 w-5" />
                Platform Functions
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {mockEdgeData.functions.map((func) => (
                  <div
                    key={func.id}
                    className="flex items-center justify-between p-4 border rounded-lg hover:bg-muted/50 transition-colors"
                  >
                    <div className="flex items-center gap-4">
                      <div className="p-2 bg-muted rounded-lg">
                        <Code className="h-4 w-4" />
                      </div>
                      <div>
                        <div className="font-medium">{func.name}</div>
                        <div className="text-sm text-muted-foreground">
                          {func.region} • {formatNumber(func.requests24h)} requests/24h
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center gap-4">
                      <div className="text-right">
                        <div className="text-sm font-medium">{func.avgLatency}ms</div>
                        <div className="text-xs text-muted-foreground">
                          {func.errorRate}% errors
                        </div>
                      </div>
                      {getEdgeStatusBadge(func.status)}
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}