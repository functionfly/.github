import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { Activity, Database, Server, Cpu, HardDrive, Zap, AlertTriangle, CheckCircle, Clock, ArrowLeft } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { LoadingSpinner } from "@/components/common/LoadingSpinner";
import { AnalyticsManagement } from "./components/AnalyticsManagement";
import { DatabaseMonitoring } from "./components/DatabaseMonitoring";
import { SectionErrorBoundary } from "./components/SectionErrorBoundary";
import { healthApi, incidentsApi, type SystemHealth, type Incident } from "@/api/admin";

export function AdminSystemPage() {
  const navigate = useNavigate();
  const [isRefreshing, setIsRefreshing] = useState(false);

  const { data: systemHealth, isLoading: healthLoading, error: healthError, refetch } = useQuery({
    queryKey: ['system-health'],
    queryFn: () => healthApi.getSystemHealth(),
    refetchInterval: 30000, // Refetch every 30 seconds
  });

  const { data: incidentsData, isLoading: incidentsLoading, error: incidentsError } = useQuery({
    queryKey: ['incidents'],
    queryFn: () => incidentsApi.listIncidents({ limit: 10 }),
    refetchInterval: 60000, // Refetch every minute
  });

  const handleRefreshHealth = async () => {
    setIsRefreshing(true);
    await refetch();
    setIsRefreshing(false);
  };

  const getHealthStatusColor = (status: string, healthy: boolean) => {
    if (!healthy) return "text-red-400";
    switch (status) {
      case "ok":
        return "text-emerald-400";
      case "degraded":
        return "text-amber-400";
      default:
        return "text-text-secondary";
    }
  };

  const getIncidentSeverityColor = (severity: string) => {
    switch (severity) {
      case "critical":
        return "bg-red-500/10 text-red-400 border-red-500/20";
      case "high":
        return "bg-orange-500/10 text-orange-400 border-orange-500/20";
      case "medium":
        return "bg-amber-500/10 text-amber-400 border-amber-500/20";
      case "low":
        return "bg-blue-500/10 text-blue-400 border-blue-500/20";
      default:
        return "bg-gray-500/10 text-text-secondary border-gray-500/20";
    }
  };

  const getIncidentStatusColor = (status: string) => {
    switch (status) {
      case "resolved":
        return "bg-emerald-500/10 text-emerald-400";
      case "investigating":
        return "bg-blue-500/10 text-blue-400";
      case "monitoring":
        return "bg-amber-500/10 text-amber-400";
      default:
        return "bg-gray-500/10 text-text-secondary";
    }
  };

  return (
    <SectionErrorBoundary sectionTitle="System Health Page">
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
          <h1 className="text-2xl font-bold text-text-primary">System Health</h1>
          <p className="text-text-secondary">Monitor system performance and health status</p>
        </div>
        <Button
          onClick={handleRefreshHealth}
          disabled={isRefreshing}
          variant="outline"
          className="border-border-subtle hover:bg-bg-hover"
        >
          <Activity className="w-4 h-4 mr-2" />
          {isRefreshing ? "Refreshing..." : "Refresh"}
        </Button>
      </div>

      {/* Overall Status */}
      {healthLoading ? (
        <Card className="card">
          <CardContent className="card-content flex items-center justify-center py-8">
            <LoadingSpinner />
          </CardContent>
        </Card>
      ) : healthError ? (
        <Card className="card">
          <CardContent className="card-content">
            <div className="text-center py-8">
              <AlertTriangle className="w-12 h-12 text-red-400 mx-auto mb-4" />
              <p className="text-text-primary font-medium">Failed to load system health</p>
              <p className="text-text-secondary">Please try refreshing the page</p>
            </div>
          </CardContent>
        </Card>
      ) : systemHealth ? (
        <Card className="card">
          <CardContent className="card-content">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <div className={`w-3 h-3 rounded-full ${
                  systemHealth.status === "healthy" ? "bg-emerald-400" : "bg-red-400"
                }`} />
                <div>
                  <h2 className="text-xl font-semibold text-text-primary">
                    System Status: {systemHealth.status === "healthy" ? "Healthy" : "Unhealthy"}
                  </h2>
                  <p className="text-text-secondary">
                    Last updated: {new Date(systemHealth.timestamp).toLocaleString()}
                  </p>
                </div>
              </div>
              <Badge className="badge-primary">
                Version {systemHealth.version}
              </Badge>
            </div>
          </CardContent>
        </Card>
      ) : null}

      {/* Health Checks */}
      {systemHealth?.checks && typeof systemHealth.checks === "object" && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {Object.entries(systemHealth.checks).map(([name, check]) => (
            <Card key={name} className="card">
              <CardContent className="card-content">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    {name === "database" && <Database className={`w-5 h-5 ${getHealthStatusColor(check.status, check.healthy)}`} />}
                    {name === "api" && <Server className={`w-5 h-5 ${getHealthStatusColor(check.status, check.healthy)}`} />}
                    {name === "repository" && <HardDrive className={`w-5 h-5 ${getHealthStatusColor(check.status, check.healthy)}`} />}
                    {name === "system" && <Cpu className={`w-5 h-5 ${getHealthStatusColor(check.status, check.healthy)}`} />}
                    <div>
                      <p className="font-medium text-text-primary capitalize">{name}</p>
                      <p className="text-sm text-text-secondary">{check.status}</p>
                    </div>
                  </div>
                  {check.healthy ? (
                    <CheckCircle className="w-5 h-5 text-emerald-400" />
                  ) : (
                    <AlertTriangle className="w-5 h-5 text-red-400" />
                  )}
                </div>
                {'response_time_ms' in check && check.response_time_ms && (
                  <div className="mt-3">
                    <div className="flex justify-between text-sm">
                      <span className="text-text-secondary">Response Time</span>
                      <span className="text-text-primary">{check.response_time_ms}ms</span>
                    </div>
                  </div>
                )}
                {'goroutines' in check && check.goroutines && (
                  <div className="mt-3">
                    <div className="flex justify-between text-sm">
                      <span className="text-text-secondary">Goroutines</span>
                      <span className="text-text-primary">{check.goroutines}</span>
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Performance Metrics */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card className="card">
          <CardHeader className="card-header">
            <CardTitle>Performance Metrics</CardTitle>
          </CardHeader>
          <CardContent className="card-content space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-sm text-text-secondary">Uptime</p>
                <p className="text-xl font-semibold text-text-primary">99.98%</p>
              </div>
              <div>
                <p className="text-sm text-text-secondary">Avg Response Time</p>
                <p className="text-xl font-semibold text-text-primary">45ms</p>
              </div>
              <div>
                <p className="text-sm text-text-secondary">Requests/min</p>
                <p className="text-xl font-semibold text-text-primary">1,250</p>
              </div>
              <div>
                <p className="text-sm text-text-secondary">Error Rate</p>
                <p className="text-xl font-semibold text-text-primary">0.02%</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="card">
          <CardHeader className="card-header">
            <CardTitle>Resource Usage</CardTitle>
          </CardHeader>
          <CardContent className="card-content space-y-4">
            <div>
              <div className="flex justify-between text-sm mb-2">
                <span className="text-text-secondary">Memory Usage</span>
                <span className="text-text-primary">72%</span>
              </div>
              <Progress value={72} className="h-2" />
            </div>
            <div>
              <div className="flex justify-between text-sm mb-2">
                <span className="text-text-secondary">CPU Usage</span>
                <span className="text-text-primary">34%</span>
              </div>
              <Progress value={34} className="h-2" />
            </div>
            <div>
              <div className="flex justify-between text-sm mb-2">
                <span className="text-text-secondary">Disk Usage</span>
                <span className="text-text-primary">45%</span>
              </div>
              <Progress value={45} className="h-2" />
            </div>
            <div className="pt-2 border-t border-white/8">
              <div className="flex justify-between text-sm">
                <span className="text-text-secondary">Active Connections</span>
                <span className="text-text-primary">89</span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Active Incidents */}
      <Card className="card">
        <CardHeader className="card-header">
          <CardTitle>Active Incidents</CardTitle>
        </CardHeader>
        <CardContent className="card-content">
          {!Array.isArray(incidentsData?.incidents) || incidentsData.incidents.length === 0 ? (
            <div className="text-center py-8">
              <CheckCircle className="w-12 h-12 text-emerald-400 mx-auto mb-4" />
              <p className="text-text-primary font-medium">All Systems Operational</p>
              <p className="text-text-secondary">No active incidents at this time</p>
            </div>
          ) : (
            <div className="space-y-4">
              {incidentsData.incidents.map((incident) => (
                <Alert key={incident.id} className={`border ${getIncidentSeverityColor(incident.severity)}`}>
                  <AlertTriangle className="h-4 w-4" />
                  <AlertDescription>
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-1">
                          <h4 className="font-medium text-text-primary">{incident.title}</h4>
                          <Badge className={getIncidentStatusColor(incident.status)}>
                            {incident.status}
                          </Badge>
                        </div>
                        <p className="text-sm text-text-secondary mb-2">{incident.description}</p>
                        <div className="flex items-center gap-4 text-xs text-text-muted">
                          <span className="flex items-center gap-1">
                            <Clock className="w-3 h-3" />
                            {new Date(incident.created_at).toLocaleString()}
                          </span>
                          {incident.resolved_at && (
                            <span>Resolved: {new Date(incident.resolved_at).toLocaleString()}</span>
                          )}
                        </div>
                      </div>
                    </div>
                  </AlertDescription>
                </Alert>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Analytics Management */}
      <SectionErrorBoundary sectionTitle="Analytics Management">
        <AnalyticsManagement />
      </SectionErrorBoundary>

      {/* Database Monitoring */}
      <SectionErrorBoundary sectionTitle="Database Monitoring">
        <DatabaseMonitoring />
      </SectionErrorBoundary>
      </div>
    </SectionErrorBoundary>
  );
}
