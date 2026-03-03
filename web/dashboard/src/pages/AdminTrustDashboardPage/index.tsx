import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import {
  Shield,
  AlertTriangle,
  TrendingUp,
  Users,
  ArrowLeft,
  RefreshCw,
  AlertCircle,
  Clock,
  MapPin,
  Activity,
  ChevronRight,
  CheckCircle,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { StatCard } from "@/components/common/StatCard";
import { oversightApi, type TrustDashboardData, type ReputationFarmAlert } from "@/api/admin";

// Helper functions
const getTrustScoreColor = (score: number): string => {
  if (score >= 70) return "text-emerald-400";
  if (score >= 50) return "text-blue-400";
  if (score >= 40) return "text-amber-400";
  return "text-red-400";
};

const getTrustScoreBg = (score: number): string => {
  if (score >= 70) return "bg-emerald-500/10";
  if (score >= 50) return "bg-blue-500/10";
  if (score >= 40) return "bg-amber-500/10";
  return "bg-red-500/10";
};

const getSeverityColor = (severity: string): string => {
  switch (severity) {
    case "high":
      return "bg-red-500/10 text-red-400 border-red-500/20";
    case "medium":
      return "bg-amber-500/10 text-amber-400 border-amber-500/20";
    default:
      return "bg-blue-500/10 text-blue-400 border-blue-500/20";
  }
};

const getAlertTypeIcon = (type: string) => {
  switch (type) {
    case "ip_range":
      return MapPin;
    case "rapid_rating":
      return Activity;
    case "new_account_cluster":
      return Users;
    default:
      return AlertCircle;
  }
};

export function AdminTrustDashboardPage() {
  const navigate = useNavigate();

  const { data, isLoading, refetch } = useQuery<TrustDashboardData>({
    queryKey: ["admin-trust-dashboard"],
    queryFn: () => oversightApi.getTrustDashboard(),
  });

  const totalFunctions = data
    ? data.distribution.excellent + data.distribution.good + data.distribution.fair + data.distribution.poor
    : 0;

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
  };

  const formatDate = (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toLocaleDateString();
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <Button
            variant="ghost"
            onClick={() => navigate("/admin")}
            className="text-text-muted hover:text-text-primary hover:bg-bg-hover"
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back
          </Button>
          <div>
            <h1 className="text-2xl font-bold text-text-primary">Trust Dashboard</h1>
            <p className="text-sm text-text-secondary">Monitor trust distribution and detect anomalies</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => refetch()}
            className="border-border-default hover:bg-bg-hover text-text-secondary"
          >
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </Button>
        </div>
      </div>

      {/* Stats Overview */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard
          title="Total Functions"
          value={isLoading ? "—" : totalFunctions.toLocaleString()}
          icon={<Shield className="w-5 h-5 text-brand-500" />}
          trend="neutral"
          change={{ value: 0, label: "monitored" }}
        />
        <StatCard
          title="Excellent Trust"
          value={isLoading ? "—" : data?.distribution.excellent.toLocaleString() || "0"}
          icon={<CheckCircle className="w-5 h-5 text-emerald-400" />}
          trend="up"
          change={{ value: 12, label: "this week" }}
        />
        <StatCard
          title="High Risk"
          value={isLoading ? "—" : data?.highRiskFunctions.length.toString() || "0"}
          icon={<AlertTriangle className="w-5 h-5 text-red-400" />}
          trend="down"
          change={{ value: -2, label: "vs last week" }}
        />
        <StatCard
          title="Active Alerts"
          value={isLoading ? "—" : (data?.reputationFarmingAlerts.length || 0).toString()}
          icon={<AlertCircle className="w-5 h-5 text-amber-400" />}
          trend="neutral"
          change={{ value: 0, label: "requiring review" }}
        />
      </div>

      {/* Trust Distribution */}
      <Card className="glass-card">
        <CardHeader>
          <CardTitle className="text-text-primary flex items-center gap-2">
            <Shield className="w-5 h-5 text-brand-500" />
            Trust Distribution Overview
          </CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="h-32 flex items-center justify-center">
              <Skeleton className="h-24 w-full" />
            </div>
          ) : (
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="p-4 rounded-lg bg-emerald-500/10 border border-emerald-500/20">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm text-emerald-400 font-medium">Excellent</span>
                  <CheckCircle className="w-4 h-4 text-emerald-400" />
                </div>
                <p className="text-3xl font-bold text-emerald-400">{data?.distribution.excellent}</p>
                <p className="text-xs text-text-muted mt-1">
                  {totalFunctions > 0 ? ((data!.distribution.excellent / totalFunctions) * 100).toFixed(1) : 0}% of total
                </p>
              </div>
              <div className="p-4 rounded-lg bg-blue-500/10 border border-blue-500/20">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm text-blue-400 font-medium">Good</span>
                  <Shield className="w-4 h-4 text-blue-400" />
                </div>
                <p className="text-3xl font-bold text-blue-400">{data?.distribution.good}</p>
                <p className="text-xs text-text-muted mt-1">
                  {totalFunctions > 0 ? ((data!.distribution.good / totalFunctions) * 100).toFixed(1) : 0}% of total
                </p>
              </div>
              <div className="p-4 rounded-lg bg-amber-500/10 border border-amber-500/20">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm text-amber-400 font-medium">Fair</span>
                  <AlertCircle className="w-4 h-4 text-amber-400" />
                </div>
                <p className="text-3xl font-bold text-amber-400">{data?.distribution.fair}</p>
                <p className="text-xs text-text-muted mt-1">
                  {totalFunctions > 0 ? ((data!.distribution.fair / totalFunctions) * 100).toFixed(1) : 0}% of total
                </p>
              </div>
              <div className="p-4 rounded-lg bg-red-500/10 border border-red-500/20">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm text-red-400 font-medium">Poor</span>
                  <AlertTriangle className="w-4 h-4 text-red-400" />
                </div>
                <p className="text-3xl font-bold text-red-400">{data?.distribution.poor}</p>
                <p className="text-xs text-text-muted mt-1">
                  {totalFunctions > 0 ? ((data!.distribution.poor / totalFunctions) * 100).toFixed(1) : 0}% of total
                </p>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* High-Risk Functions */}
        <Card className="glass-card">
          <CardHeader>
            <CardTitle className="text-text-primary flex items-center gap-2">
              <AlertTriangle className="w-5 h-5 text-red-400" />
              High-Risk Functions
              <Badge variant="secondary" className="bg-red-500/10 text-red-400 border-red-500/20">
                Score {'<'} 40
              </Badge>
            </CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            {isLoading ? (
              <div className="p-6 space-y-4">
                {Array.from({ length: 3 }).map((_, i) => (
                  <Skeleton key={i} className="h-20 w-full" />
                ))}
              </div>
            ) : data?.highRiskFunctions.length === 0 ? (
              <div className="text-center py-8">
                <CheckCircle className="w-12 h-12 text-emerald-400 mx-auto mb-3" />
                <p className="text-text-primary font-medium">No high-risk functions</p>
                <p className="text-sm text-text-muted">All functions have trust scores above 40</p>
              </div>
            ) : (
              <div className="divide-y divide-border-subtle">
                {data?.highRiskFunctions.map((fn) => (
                  <div
                    key={fn.id}
                    className="p-4 hover:bg-bg-hover transition-colors cursor-pointer"
                    onClick={() => navigate(`/admin/registry/functions/${fn.id}`)}
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 mb-1">
                          <p className="font-medium text-text-primary truncate">{fn.name}</p>
                          <Badge variant="secondary" className="text-xs">
                            {fn.tenant}
                          </Badge>
                        </div>
                        <div className="flex flex-wrap gap-1 mt-2">
                          {fn.riskFactors.map((factor, idx) => (
                            <span
                              key={idx}
                              className="text-xs px-2 py-0.5 rounded-full bg-red-500/10 text-red-400"
                            >
                              {factor}
                            </span>
                          ))}
                        </div>
                      </div>
                      <div className="flex items-center gap-3 ml-4">
                        <div className={`text-center px-3 py-1 rounded-lg ${getTrustScoreBg(fn.trustScore)}`}>
                          <p className={`text-lg font-bold ${getTrustScoreColor(fn.trustScore)}`}>{fn.trustScore}</p>
                          <p className="text-xs text-text-muted">score</p>
                        </div>
                        <ChevronRight className="w-5 h-5 text-text-muted" />
                      </div>
                    </div>
                    <p className="text-xs text-text-muted mt-2">
                      Last updated: {formatDate(fn.lastUpdated)} at {formatTime(fn.lastUpdated)}
                    </p>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Suspicious Trust Spikes */}
        <Card className="glass-card">
          <CardHeader>
            <CardTitle className="text-text-primary flex items-center gap-2">
              <TrendingUp className="w-5 h-5 text-amber-400" />
              Suspicious Trust Spikes
              <Badge variant="secondary" className="bg-amber-500/10 text-amber-400 border-amber-500/20">
                {'>'}20 pts in 24h
              </Badge>
            </CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            {isLoading ? (
              <div className="p-6 space-y-4">
                {Array.from({ length: 3 }).map((_, i) => (
                  <Skeleton key={i} className="h-24 w-full" />
                ))}
              </div>
            ) : data?.trustSpikes.length === 0 ? (
              <div className="text-center py-8">
                <CheckCircle className="w-12 h-12 text-emerald-400 mx-auto mb-3" />
                <p className="text-text-primary font-medium">No suspicious spikes</p>
                <p className="text-sm text-text-muted">No unusual trust score increases detected</p>
              </div>
            ) : (
              <div className="divide-y divide-border-subtle">
                {data?.trustSpikes.map((spike) => (
                  <div key={spike.id} className="p-4 hover:bg-bg-hover transition-colors">
                    <div className="flex items-start justify-between">
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 mb-1">
                          <p className="font-medium text-text-primary">{spike.functionName}</p>
                          <Badge variant="secondary" className="text-xs">
                            {spike.tenant}
                          </Badge>
                        </div>
                        <div className="flex items-center gap-3 mt-2">
                          <div className="flex items-center gap-1">
                            <span className="text-sm text-text-muted">{spike.previousScore}</span>
                            <TrendingUp className="w-4 h-4 text-amber-400" />
                            <span className="text-sm font-medium text-amber-400">{spike.newScore}</span>
                          </div>
                          <Badge className="bg-amber-500/10 text-amber-400 border-amber-500/20">
                            +{spike.spikeAmount} pts
                          </Badge>
                        </div>
                      </div>
                      <div className="flex items-center gap-2 ml-4">
                        <Clock className="w-4 h-4 text-text-muted" />
                        <span className="text-xs text-text-muted">{formatTime(spike.detectedAt)}</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Reputation Farming Alerts */}
      <Card className="glass-card">
        <CardHeader>
          <CardTitle className="text-text-primary flex items-center gap-2">
            <Users className="w-5 h-5 text-purple-400" />
            Reputation Farming Alerts
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="p-6 space-y-4">
              {Array.from({ length: 3 }).map((_, i) => (
                <Skeleton key={i} className="h-24 w-full" />
              ))}
            </div>
          ) : data?.reputationFarmingAlerts.length === 0 ? (
            <div className="text-center py-8">
              <CheckCircle className="w-12 h-12 text-emerald-400 mx-auto mb-3" />
              <p className="text-text-primary font-medium">No farming patterns detected</p>
              <p className="text-sm text-text-muted">No suspicious reputation manipulation activity</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 p-6">
              {data?.reputationFarmingAlerts.map((alert) => {
                const Icon = getAlertTypeIcon(alert.type);
                return (
                  <div
                    key={alert.id}
                    className={`p-4 rounded-lg border ${getSeverityColor(alert.severity)}`}
                  >
                    <div className="flex items-start justify-between mb-3">
                      <div className="flex items-center gap-2">
                        <Icon className="w-4 h-4" />
                        <span className="font-medium capitalize">{alert.type.replace("_", " ")}</span>
                      </div>
                      <Badge className={getSeverityColor(alert.severity)}>{alert.severity}</Badge>
                    </div>
                    <p className="text-sm text-text-secondary mb-3">{alert.description}</p>
                    <div className="space-y-2 text-xs">
                      {alert.details.ipRange && (
                        <div className="flex justify-between">
                          <span className="text-text-muted">IP Range:</span>
                          <span className="font-mono">{alert.details.ipRange}</span>
                        </div>
                      )}
                      {alert.details.accountCount && (
                        <div className="flex justify-between">
                          <span className="text-text-muted">Accounts:</span>
                          <span>{alert.details.accountCount}</span>
                        </div>
                      )}
                      {alert.details.ratingCount && (
                        <div className="flex justify-between">
                          <span className="text-text-muted">Ratings:</span>
                          <span>{alert.details.ratingCount}</span>
                        </div>
                      )}
                      {alert.details.timeWindow && (
                        <div className="flex justify-between">
                          <span className="text-text-muted">Window:</span>
                          <span>{alert.details.timeWindow}</span>
                        </div>
                      )}
                      <div className="flex justify-between">
                        <span className="text-text-muted">Functions:</span>
                        <span>{alert.affectedFunctions.length}</span>
                      </div>
                    </div>
                    <p className="text-xs text-text-muted mt-3 pt-3 border-t border-border-subtle">
                      Detected at {formatTime(alert.detectedAt)}
                    </p>
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
