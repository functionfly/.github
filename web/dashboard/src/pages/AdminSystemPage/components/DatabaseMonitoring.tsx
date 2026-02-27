import { useState, useEffect } from "react";
import { Database, Activity, AlertTriangle, CheckCircle, Zap, Users, Clock, TrendingUp, BarChart3 } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Progress } from "@/components/ui/progress";
import { useDatabaseHealth, useDatabaseAlerts, useDatabaseMetrics } from "@/hooks/useRealtime";


export function DatabaseMonitoring() {
  const { health: dbHealth, loading: healthLoading, error: healthError, isRealtimeConnected: realtimeConnected } = useDatabaseHealth();
  const { alerts, loading: alertsLoading } = useDatabaseAlerts();
  const { metrics, loading: metricsLoading, error: metricsError } = useDatabaseMetrics('1h');

  const isLoading = healthLoading || alertsLoading || metricsLoading;
  const error = healthError || metricsError;

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'healthy':
        return 'text-emerald-400';
      case 'degraded':
        return 'text-amber-400';
      case 'unhealthy':
      case 'error':
        return 'text-red-400';
      default:
        return 'text-gray-400';
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'bg-red-500/10 text-red-400 border-red-500/20';
      case 'high':
        return 'bg-orange-500/10 text-orange-400 border-orange-500/20';
      case 'medium':
        return 'bg-amber-500/10 text-amber-400 border-amber-500/20';
      case 'low':
        return 'bg-blue-500/10 text-blue-400 border-blue-500/20';
      default:
        return 'bg-gray-500/10 text-gray-400 border-gray-500/20';
    }
  };

  const formatBytes = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
    return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`;
  };

  if (isLoading) {
    return (
      <Card className="card">
        <CardContent className="card-content">
          <div className="flex items-center justify-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-emerald-400"></div>
            <span className="ml-3 text-text-secondary">Loading database monitoring...</span>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className="card">
        <CardContent className="card-content">
          <Alert className="border-red-500/20 bg-red-500/10">
            <AlertTriangle className="h-4 w-4 text-red-400" />
            <AlertDescription className="text-red-400">
              Failed to load database monitoring: {error}
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    );
  }

  if (!dbHealth) return null;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Database className={`w-6 h-6 ${getStatusColor(dbHealth.status)}`} />
          <div>
            <h2 className="text-xl font-semibold text-white">Database Monitoring</h2>
            <div className="flex items-center gap-2 mt-1">
              <div className={`w-2 h-2 rounded-full ${
                dbHealth.status === 'healthy' ? 'bg-emerald-400' :
                dbHealth.status === 'degraded' ? 'bg-amber-400' : 'bg-red-400'
              }`} />
              <span className="text-sm text-text-secondary capitalize">{dbHealth.status}</span>
              {realtimeConnected && (
                <Badge className="bg-emerald-500/10 text-emerald-400 border-emerald-500/20 text-xs">
                  <Activity className="w-3 h-3 mr-1" />
                  Live
                </Badge>
              )}
            </div>
          </div>
        </div>
        <Button
          variant="outline"
          size="sm"
          className="btn-outline"
        >
          <BarChart3 className="w-4 h-4 mr-2" />
          View Details
        </Button>
      </div>

      {/* Connection Pool Status */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="card">
          <CardContent className="card-content">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <Users className="w-5 h-5 text-blue-400" />
                <div>
                  <p className="font-medium text-white">Active Connections</p>
                  <p className="text-sm text-text-secondary">Pool utilization</p>
                </div>
              </div>
              <div className="text-right">
                <p className="text-xl font-semibold text-white">{dbHealth.connections.active}</p>
                <p className="text-xs text-text-secondary">of {dbHealth.connections.max}</p>
              </div>
            </div>
            <div className="mt-3">
              <Progress value={(dbHealth.connections.active / dbHealth.connections.max) * 100} className="h-2" />
            </div>
          </CardContent>
        </Card>

        <Card className="card">
          <CardContent className="card-content">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <Zap className="w-5 h-5 text-amber-400" />
                <div>
                  <p className="font-medium text-white">Query Throughput</p>
                  <p className="text-sm text-text-secondary">Queries/min</p>
                </div>
              </div>
              <div className="text-right">
                <p className="text-xl font-semibold text-white">{dbHealth.performance.throughput.toLocaleString()}</p>
                <p className="text-xs text-emerald-400">+12%</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="card">
          <CardContent className="card-content">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <Clock className="w-5 h-5 text-purple-400" />
                <div>
                  <p className="font-medium text-white">Avg Response Time</p>
                  <p className="text-sm text-text-secondary">Query latency</p>
                </div>
              </div>
              <div className="text-right">
                <p className="text-xl font-semibold text-white">{dbHealth.performance.avgQueryTime}ms</p>
                <p className="text-xs text-text-secondary">Slow: {dbHealth.performance.slowQueries}</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="card">
          <CardContent className="card-content">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <TrendingUp className={`w-5 h-5 ${
                  dbHealth.replication.status === 'healthy' ? 'text-emerald-400' :
                  dbHealth.replication.status === 'lagging' ? 'text-amber-400' : 'text-red-400'
                }`} />
                <div>
                  <p className="font-medium text-white">Replication Lag</p>
                  <p className="text-sm text-text-secondary">Sync status</p>
                </div>
              </div>
              <div className="text-right">
                <p className="text-xl font-semibold text-white">{dbHealth.replication.lag}ms</p>
                <p className={`text-xs capitalize ${getStatusColor(dbHealth.replication.status)}`}>
                  {dbHealth.replication.status}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Storage Usage */}
      <Card className="card">
        <CardHeader className="card-header">
          <CardTitle>Storage Usage</CardTitle>
        </CardHeader>
        <CardContent className="card-content">
          <div className="space-y-4">
            <div>
              <div className="flex justify-between text-sm mb-2">
                <span className="text-text-secondary">Database Size</span>
                <span className="text-white">{dbHealth.storage.used}GB of {dbHealth.storage.total}GB</span>
              </div>
              <Progress value={(dbHealth.storage.used / dbHealth.storage.total) * 100} className="h-3" />
            </div>
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span className="text-text-secondary">Growth Rate:</span>
                <span className="text-white ml-2">+{dbHealth.storage.growthRate}GB/week</span>
              </div>
              <div>
                <span className="text-text-secondary">Days to Full:</span>
                <span className="text-white ml-2">
                  {Math.round((dbHealth.storage.total - dbHealth.storage.used) / (dbHealth.storage.growthRate / 7))} days
                </span>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Active Alerts */}
      <Card className="card">
        <CardHeader className="card-header">
          <CardTitle>Database Alerts</CardTitle>
        </CardHeader>
        <CardContent className="card-content">
          {alerts.length === 0 ? (
            <div className="text-center py-8">
              <CheckCircle className="w-12 h-12 text-emerald-400 mx-auto mb-4" />
              <p className="text-white font-medium">All Systems Operational</p>
              <p className="text-text-secondary">No active database alerts</p>
            </div>
          ) : (
            <div className="space-y-4">
              {alerts.map((alert) => (
                <Alert key={alert.id} className={`border ${getSeverityColor(alert.severity)}`}>
                  <AlertTriangle className="h-4 w-4" />
                  <AlertDescription>
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-1">
                          <h4 className="font-medium text-white">{alert.title}</h4>
                          {alert.resolved && (
                            <Badge className="bg-emerald-500/10 text-emerald-400 border-emerald-500/20 text-xs">
                              Resolved
                            </Badge>
                          )}
                        </div>
                        <p className="text-sm text-text-secondary mb-2">{alert.message}</p>
                        <div className="flex items-center gap-4 text-xs text-text-muted">
                          <span className="flex items-center gap-1">
                            <Clock className="w-3 h-3" />
                            {new Date(alert.timestamp).toLocaleString()}
                          </span>
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

      {/* Recent Metrics Chart Placeholder */}
      <Card className="card">
        <CardHeader className="card-header">
          <CardTitle>Query Performance Trends</CardTitle>
        </CardHeader>
        <CardContent className="card-content">
          <div className="h-48 flex items-center justify-center border border-dashed border-white/8 rounded-lg">
            <div className="text-center">
              <BarChart3 className="w-12 h-12 text-text-secondary mx-auto mb-3" />
              <p className="text-text-secondary">Performance chart visualization</p>
              <p className="text-sm text-text-muted">Integration with charting library needed</p>
            </div>
          </div>
          <div className="mt-4 grid grid-cols-4 gap-4 text-sm">
            <div>
              <span className="text-text-secondary">Peak Connections:</span>
              <span className="text-white ml-2">
                {metrics.length > 0 ? Math.max(...metrics.map(m => m.connections)) : 0}
              </span>
            </div>
            <div>
              <span className="text-text-secondary">Avg Response Time:</span>
              <span className="text-white ml-2">
                {metrics.length > 0 ? Math.round(metrics.reduce((sum, m) => sum + m.avgResponseTime, 0) / metrics.length) : 0}ms
              </span>
            </div>
            <div>
              <span className="text-text-secondary">Total Queries:</span>
              <span className="text-white ml-2">
                {metrics.length > 0 ? metrics.reduce((sum, m) => sum + m.queryCount, 0).toLocaleString() : 0}
              </span>
            </div>
            <div>
              <span className="text-text-secondary">Avg Error Rate:</span>
              <span className="text-white ml-2">
                {metrics.length > 0 ? (metrics.reduce((sum, m) => sum + m.errorRate, 0) / metrics.length * 100).toFixed(2) : 0}%
              </span>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}