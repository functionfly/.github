import { useParams, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import {
  ArrowLeft,
  Building2,
  Loader2,
  AlertCircle,
  Server,
  Clock,
  CheckCircle2,
  XCircle,
  RefreshCw,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { appsApi } from "@/api/apps";
import type { AppStatus, BackendStatus } from "@/types";
import { ROUTES } from "@/lib/constants";
import { cn } from "@/lib/utils";

function BackendStatusCard({ backendStatus }: { backendStatus: BackendStatus }) {
  const { backend, circuitState, latestHealthCheck } = backendStatus;

  const getStatusIcon = () => {
    if (!latestHealthCheck) {
      return <Clock className="w-4 h-4 text-muted-foreground" />;
    }
    if (latestHealthCheck.ok) {
      return <CheckCircle2 className="w-4 h-4 text-emerald-500" />;
    }
    return <XCircle className="w-4 h-4 text-red-500" />;
  };

  const getStatusColor = () => {
    if (circuitState?.state === "open") return "bg-red-500";
    if (circuitState?.state === "half-open") return "bg-amber-500";
    if (latestHealthCheck?.ok === false) return "bg-red-500";
    if (latestHealthCheck?.ok) return "bg-emerald-500";
    return "bg-gray-500";
  };

  return (
    <div className="flex items-center gap-4 p-4 border rounded-lg">
      <div className={cn("w-3 h-3 rounded-full", getStatusColor())} />
      <div className="flex-1">
        <div className="flex items-center gap-2">
          <Server className="w-4 h-4 text-muted-foreground" />
          <span className="font-medium">{backend.provider}</span>
          <Badge variant="outline">{backend.region}</Badge>
        </div>
        <p className="text-sm text-muted-foreground truncate">{backend.url}</p>
      </div>
      <div className="flex items-center gap-2">
        {getStatusIcon()}
        {latestHealthCheck && (
          <span className="text-xs text-muted-foreground">
            {latestHealthCheck.latencyMs}ms
          </span>
        )}
      </div>
    </div>
  );
}

function EmptyBackends() {
  return (
    <div className="text-center py-8 text-muted-foreground">
      <Server className="w-8 h-8 mx-auto mb-2 opacity-50" />
      <p>No backends configured</p>
      <p className="text-sm">Add a backend to start deploying your app</p>
    </div>
  );
}

function LoadingState() {
  return (
    <div className="flex items-center justify-center py-12">
      <Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
    </div>
  );
}

function ErrorState({ message, onRetry }: { message: string; onRetry: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center">
      <AlertCircle className="w-8 h-8 text-red-500 mb-2" />
      <p className="text-red-500 mb-4">{message}</p>
      <Button variant="outline" onClick={onRetry}>
        <RefreshCw className="w-4 h-4 mr-2" />
        Retry
      </Button>
    </div>
  );
}

export function AppDetailPage() {
  const { appId } = useParams<{ appId: string }>();

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ["app-status", appId],
    queryFn: async () => {
      if (!appId) throw new Error("App ID is required");
      const response = await appsApi.getStatus(appId);
      return response;
    },
    enabled: !!appId,
  });

  if (isLoading) {
    return <LoadingState />;
  }

  if (error || !data) {
    return (
      <ErrorState
        message={error instanceof Error ? error.message : "Failed to load app details"}
        onRetry={() => refetch()}
      />
    );
  }

  const { app, backends } = data as AppStatus;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Link to={ROUTES.APPS} aria-label="Back to apps">
          <Button variant="ghost" size="icon" aria-label="Back to apps">
            <ArrowLeft className="w-5 h-5" />
          </Button>
        </Link>
        <div className="flex-1">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 rounded-lg bg-brand-100 dark:bg-brand-900 flex items-center justify-center">
              <Building2 className="w-6 h-6 text-brand-600" />
            </div>
            <div>
              <h1 className="text-2xl font-bold">{app.name}</h1>
              <p className="text-muted-foreground">{app.slug}</p>
            </div>
          </div>
        </div>
        <div className="text-right">
          <p className="text-sm text-muted-foreground">Created</p>
          <p className="text-sm">{new Date(app.createdAt).toLocaleDateString()}</p>
        </div>
      </div>

      {/* Overview Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Backends
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{backends.length}</div>
            <p className="text-xs text-muted-foreground">connected providers</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Healthy Backends
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              {backends.filter((b) => b.latestHealthCheck?.ok).length}
            </div>
            <p className="text-xs text-muted-foreground">operational</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Circuit Breaker
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              {backends.filter((b) => b.circuitState?.state === "open").length}
            </div>
            <p className="text-xs text-muted-foreground">open circuits</p>
          </CardContent>
        </Card>
      </div>

      {/* Backends Panel */}
      <Card>
        <CardHeader>
          <CardTitle>Backends</CardTitle>
        </CardHeader>
        <CardContent>
          {backends.length > 0 ? (
            <div className="space-y-3">
              {backends.map((backendStatus) => (
                <BackendStatusCard key={backendStatus.backend.id} backendStatus={backendStatus} />
              ))}
            </div>
          ) : (
            <EmptyBackends />
          )}
        </CardContent>
      </Card>

      {/* Quick Actions */}
      <div className="flex gap-3">
        <Button variant="outline">
          <Server className="w-4 h-4 mr-2" />
          Add Backend
        </Button>
        <Button variant="outline">
          <RefreshCw className="w-4 h-4 mr-2" />
          Deploy
        </Button>
      </div>
    </div>
  );
}
