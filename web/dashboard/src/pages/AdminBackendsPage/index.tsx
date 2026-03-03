import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  ArrowLeft,
  RefreshCw,
  Server,
  CheckCircle,
  XCircle,
  AlertCircle,
  Database,
  Globe,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { backendsApi, type PlatformBackend } from "@/api/admin";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

export function AdminBackendsPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [isRefreshing, setIsRefreshing] = useState(false);

  // Fetch backends data
  const {
    data: backendsData,
    isLoading: backendsLoading,
    error: backendsError,
    refetch,
  } = useQuery({
    queryKey: ["admin-platform-backends"],
    queryFn: () => backendsApi.listPlatformBackends(),
  });

  // Update backend enabled status mutation
  const updateBackendMutation = useMutation({
    mutationFn: ({ backendId, enabled }: { backendId: string; enabled: boolean }) =>
      backendsApi.updateBackendEnabled(backendId, enabled),
    onSuccess: (updatedBackend) => {
      // Update the cache with the new backend data
      queryClient.setQueryData(
        ["admin-platform-backends"],
        (oldData: { backends: PlatformBackend[] } | undefined) => {
          if (!oldData) return oldData;
          return {
            backends: oldData.backends.map((backend) =>
              backend.id === updatedBackend.id ? updatedBackend : backend
            ),
          };
        }
      );
      toast.success(`Backend ${updatedBackend.enabled ? "enabled" : "disabled"} successfully`);
    },
    onError: (error) => {
      toast.error("Failed to update backend status");
      console.error("Failed to update backend:", error);
    },
  });

  const handleRefresh = async () => {
    setIsRefreshing(true);
    await refetch();
    setIsRefreshing(false);
  };

  const handleToggleEnabled = async (backend: PlatformBackend) => {
    try {
      await updateBackendMutation.mutateAsync({
        backendId: backend.id,
        enabled: !backend.enabled,
      });
    } catch (error) {
      // Error is handled in the mutation
    }
  };

  const getStatusIcon = (enabled: boolean) => {
    return enabled ? (
      <CheckCircle className="h-4 w-4 text-emerald-600 dark:text-emerald-400" />
    ) : (
      <XCircle className="h-4 w-4 text-red-600 dark:text-red-400" />
    );
  };

  const getStatusBadge = (enabled: boolean) => {
    return (
      <Badge
        variant={enabled ? "default" : "secondary"}
        className={cn(
          "flex items-center gap-1",
          enabled
            ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20"
            : "bg-red-500/10 text-red-600 dark:text-red-400 border-red-500/20"
        )}
      >
        {getStatusIcon(enabled)}
        {enabled ? "Enabled" : "Disabled"}
      </Badge>
    );
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

  if (backendsLoading) {
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

  if (backendsError) {
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
            <h1 className="text-2xl font-bold">Platform Backends</h1>
            <p className="text-muted-foreground">
              Manage platform backend enable/disable status
            </p>
          </div>
        </div>

        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            Failed to load platform backends. Please try again.
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

  const backends = backendsData?.backends || [];
  const enabledCount = backends.filter((b) => b.enabled).length;
  const totalCount = backends.length;

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
            <h1 className="text-2xl font-bold">Platform Backends</h1>
            <p className="text-muted-foreground">
              Manage platform backend enable/disable status
            </p>
          </div>
        </div>

        <div className="flex items-center gap-4">
          <div className="text-sm text-muted-foreground">
            {enabledCount} of {totalCount} backends enabled
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
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
        <Card>
          <CardContent className="flex items-center gap-4 p-6">
            <div className="p-2 bg-blue-500/10 rounded-lg">
              <Server className="h-6 w-6 text-blue-600 dark:text-blue-400" />
            </div>
            <div>
              <div className="text-2xl font-bold">{totalCount}</div>
              <div className="text-sm text-muted-foreground">Total Backends</div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="flex items-center gap-4 p-6">
            <div className="p-2 bg-emerald-500/10 rounded-lg">
              <CheckCircle className="h-6 w-6 text-emerald-600 dark:text-emerald-400" />
            </div>
            <div>
              <div className="text-2xl font-bold">{enabledCount}</div>
              <div className="text-sm text-muted-foreground">Enabled</div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="flex items-center gap-4 p-6">
            <div className="p-2 bg-red-500/10 rounded-lg">
              <XCircle className="h-6 w-6 text-red-600 dark:text-red-400" />
            </div>
            <div>
              <div className="text-2xl font-bold">{totalCount - enabledCount}</div>
              <div className="text-sm text-muted-foreground">Disabled</div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Backends Table */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Database className="h-5 w-5" />
            Platform Backends
          </CardTitle>
        </CardHeader>
        <CardContent>
          {backends.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <Server className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No platform backends found</p>
            </div>
          ) : (
            <div className="space-y-4">
              {backends.map((backend) => (
                <div
                  key={backend.id}
                  className="flex items-center justify-between p-4 border rounded-lg hover:bg-muted/50 transition-colors"
                >
                  <div className="flex items-center gap-4">
                    <div className="p-2 bg-muted rounded-lg">
                      <Globe className="h-4 w-4" />
                    </div>

                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2 mb-1">
                        <span className="font-medium truncate">
                          {backend.app_name}
                        </span>
                        <span className="text-sm text-muted-foreground">•</span>
                        <span className="text-sm text-muted-foreground truncate">
                          {backend.tenant_name}
                        </span>
                      </div>

                      <div className="flex items-center gap-4 text-sm text-muted-foreground">
                        <span className="flex items-center gap-1">
                          <Server className="h-3 w-3" />
                          {backend.provider} / {backend.region}
                        </span>
                        <span className="truncate max-w-xs">
                          {backend.url}
                        </span>
                        {backend.priority && (
                          <Badge variant="outline" className="text-xs">
                            Priority {backend.priority}
                          </Badge>
                        )}
                      </div>

                      <div className="text-xs text-muted-foreground mt-1">
                        Created {formatDate(backend.created_at)}
                        {backend.updated_at !== backend.created_at && (
                          <> • Updated {formatDate(backend.updated_at)}</>
                        )}
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center gap-4">
                    {getStatusBadge(backend.enabled)}

                    <Switch
                      checked={backend.enabled}
                      onCheckedChange={() => handleToggleEnabled(backend)}
                      disabled={updateBackendMutation.isPending}
                      aria-label={`Toggle ${backend.app_name} backend`}
                    />
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