import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import {
  Bot,
  AlertTriangle,
  MapPin,
  RefreshCw,
  ArrowLeft,
  Shield,
  Users,
  Activity,
  Clock,
  AlertCircle,
  CheckCircle,
  Globe,
  Network,
  Zap,
  ChevronRight,
  Eye,
  Ban,
  Search,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { StatCard } from "@/components/common/StatCard";
import { toast } from "sonner";
import { oversightApi, type FraudDetectionData, type BotPattern } from "@/api/admin";

// Helper functions
const getConfidenceColor = (score: number): string => {
  if (score >= 90) return "text-red-400";
  if (score >= 75) return "text-amber-400";
  return "text-blue-400";
};

const getConfidenceBg = (score: number): string => {
  if (score >= 90) return "bg-red-500/10";
  if (score >= 75) return "bg-amber-500/10";
  return "bg-blue-500/10";
};

const getRiskLevelBadge = (level: string) => {
  switch (level) {
    case "high":
      return <Badge className="bg-red-500/10 text-red-400 border-red-500/20">High Risk</Badge>;
    case "medium":
      return <Badge className="bg-amber-500/10 text-amber-400 border-amber-500/20">Medium Risk</Badge>;
    default:
      return <Badge className="bg-blue-500/10 text-blue-400 border-blue-500/20">Low Risk</Badge>;
  }
};

const getPatternTypeLabel = (type: string): string => {
  const labels: Record<string, string> = {
    coordinated_voting: "Coordinated Voting",
    automated_execution: "Automated Execution",
    synthetic_traffic: "Synthetic Traffic",
    credential_stuffing: "Credential Stuffing",
  };
  return labels[type] || type;
};

const getPatternTypeIcon = (type: string) => {
  switch (type) {
    case "coordinated_voting":
      return Users;
    case "automated_execution":
      return Bot;
    case "synthetic_traffic":
      return Activity;
    case "credential_stuffing":
      return Shield;
    default:
      return AlertCircle;
  }
};

export function AdminFraudDetectionPage() {
  const navigate = useNavigate();

  const { data, isLoading, refetch } = useQuery<FraudDetectionData>({
    queryKey: ["admin-fraud-detection"],
    queryFn: () => oversightApi.getFraudDetection(),
  });

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
  };

  const handleInvestigate = async (id: string, type: string) => {
    try {
      await oversightApi.investigateEntity(type, id);
      toast.success(`Investigation opened for ${type} ${id}`);
    } catch (error) {
      toast.error(`Failed to open investigation for ${type} ${id}`);
    }
  };

  const handleBlock = async (id: string, type: string) => {
    try {
      await oversightApi.blockEntity(type, id);
      toast.success(`${type} ${id} has been blocked`);
    } catch (error) {
      toast.error(`Failed to block ${type} ${id}`);
    }
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
            <h1 className="text-2xl font-bold text-text-primary">Fraud Detection Panel</h1>
            <p className="text-sm text-text-secondary">AI-powered fraud detection and anomaly analysis</p>
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

      {/* Summary Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard
          title="Bot Patterns"
          value={isLoading ? "—" : data?.summary.totalBotPatterns.toString() || "0"}
          icon={<Bot className="w-5 h-5 text-red-400" />}
          trend="up"
          change={{ value: 2, label: "new today" }}
        />
        <StatCard
          title="High Risk Clusters"
          value={isLoading ? "—" : data?.summary.highRiskClusters.toString() || "0"}
          icon={<Network className="w-5 h-5 text-amber-400" />}
          trend="neutral"
          change={{ value: 0, label: "active" }}
        />
        <StatCard
          title="Suspicious Tenants"
          value={isLoading ? "—" : data?.summary.suspiciousTenants.toString() || "0"}
          icon={<AlertTriangle className="w-5 h-5 text-purple-400" />}
          trend="up"
          change={{ value: 3, label: "flagged" }}
        />
        <StatCard
          title="Wash Usage"
          value={isLoading ? "—" : data?.summary.washUsageDetected.toString() || "0"}
          icon={<Zap className="w-5 h-5 text-blue-400" />}
          trend="down"
          change={{ value: -1, label: "resolved" }}
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Bot Pattern Detection */}
        <Card className="glass-card">
          <CardHeader>
            <CardTitle className="text-text-primary flex items-center gap-2">
              <Bot className="w-5 h-5 text-red-400" />
              Bot Pattern Detection
            </CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            {isLoading ? (
              <div className="p-6 space-y-4">
                {Array.from({ length: 3 }).map((_, i) => (
                  <Skeleton key={i} className="h-32 w-full" />
                ))}
              </div>
            ) : data?.botPatterns.length === 0 ? (
              <div className="text-center py-8">
                <CheckCircle className="w-12 h-12 text-emerald-400 mx-auto mb-3" />
                <p className="text-text-primary font-medium">No bot patterns detected</p>
                <p className="text-sm text-text-muted">All traffic appears to be human-generated</p>
              </div>
            ) : (
              <div className="divide-y divide-border-subtle">
                {data?.botPatterns.map((pattern) => {
                  const Icon = getPatternTypeIcon(pattern.patternType);
                  return (
                    <div key={pattern.id} className="p-4 hover:bg-bg-hover transition-colors">
                      <div className="flex items-start justify-between mb-3">
                        <div className="flex items-center gap-2">
                          <div className={`p-2 rounded-lg ${getConfidenceBg(pattern.confidenceScore)}`}>
                            <Icon className={`w-4 h-4 ${getConfidenceColor(pattern.confidenceScore)}`} />
                          </div>
                          <div>
                            <p className="font-medium text-text-primary">
                              {getPatternTypeLabel(pattern.patternType)}
                            </p>
                            <p className="text-xs text-text-muted">{pattern.id}</p>
                          </div>
                        </div>
                        <Badge className={`${getConfidenceBg(pattern.confidenceScore)} ${getConfidenceColor(pattern.confidenceScore)} border-0`}>
                          {pattern.confidenceScore}% confidence
                        </Badge>
                      </div>
                      <p className="text-sm text-text-secondary mb-3">{pattern.pattern}</p>
                      <div className="flex items-center gap-4 text-xs text-text-muted">
                        <span>{pattern.affectedFunctions.length} functions</span>
                        <span>{pattern.affectedTenants.length} tenants</span>
                        <span className="flex items-center gap-1">
                          <Clock className="w-3 h-3" />
                          {formatTime(pattern.detectedAt)}
                        </span>
                      </div>
                      <div className="flex gap-2 mt-3">
                        <Button
                          variant="outline"
                          size="sm"
                          className="border-border-default hover:bg-bg-hover"
                          onClick={() => handleInvestigate(pattern.id, "pattern")}
                        >
                          <Eye className="w-3 h-3 mr-1" />
                          Investigate
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          className="border-red-500/20 text-red-400 hover:bg-red-500/10"
                          onClick={() => handleBlock(pattern.id, "pattern")}
                        >
                          <Ban className="w-3 h-3 mr-1" />
                          Block
                        </Button>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Fake Tenant Diversity */}
        <Card className="glass-card">
          <CardHeader>
            <CardTitle className="text-text-primary flex items-center gap-2">
              <Users className="w-5 h-5 text-purple-400" />
              Fake Tenant Diversity Alerts
            </CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            {isLoading ? (
              <div className="p-6 space-y-4">
                {Array.from({ length: 2 }).map((_, i) => (
                  <Skeleton key={i} className="h-40 w-full" />
                ))}
              </div>
            ) : data?.fakeDiversityAlerts.length === 0 ? (
              <div className="text-center py-8">
                <CheckCircle className="w-12 h-12 text-emerald-400 mx-auto mb-3" />
                <p className="text-text-primary font-medium">No fake diversity detected</p>
                <p className="text-sm text-text-muted">All tenant accounts appear legitimate</p>
              </div>
            ) : (
              <div className="divide-y divide-border-subtle">
                {data?.fakeDiversityAlerts.map((alert) => (
                  <div key={alert.id} className="p-4 hover:bg-bg-hover transition-colors">
                    <div className="flex items-start justify-between mb-3">
                      <div className="flex items-center gap-2">
                        <AlertTriangle className="w-4 h-4 text-purple-400" />
                        <span className="font-medium text-text-primary">Artificial Diversity Cluster</span>
                      </div>
                      {getRiskLevelBadge(alert.riskLevel)}
                    </div>
                    <div className="space-y-3">
                      <div>
                        <p className="text-xs text-text-muted mb-1">Suspected Tenant Group</p>
                        <div className="flex flex-wrap gap-1">
                          {alert.tenantGroup.map((tenant) => (
                            <Badge key={tenant} variant="secondary" className="text-xs">
                              {tenant}
                            </Badge>
                          ))}
                        </div>
                      </div>
                      <div>
                        <p className="text-xs text-text-muted mb-1">Indicators</p>
                        <div className="flex flex-wrap gap-1">
                          {alert.indicators.map((indicator, idx) => (
                            <span
                              key={idx}
                              className="text-xs px-2 py-0.5 rounded-full bg-purple-500/10 text-purple-400"
                            >
                              {indicator}
                            </span>
                          ))}
                        </div>
                      </div>
                      <p className="text-xs text-text-muted">
                        Detected at {formatTime(alert.detectedAt)}
                      </p>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* IP Clustering */}
      <Card className="glass-card">
        <CardHeader>
          <CardTitle className="text-text-primary flex items-center gap-2">
            <Globe className="w-5 h-5 text-blue-400" />
            IP Clustering Analysis
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="p-6 space-y-4">
              {Array.from({ length: 3 }).map((_, i) => (
                <Skeleton key={i} className="h-16 w-full" />
              ))}
            </div>
          ) : data?.ipClusters.length === 0 ? (
            <div className="text-center py-8">
              <CheckCircle className="w-12 h-12 text-emerald-400 mx-auto mb-3" />
              <p className="text-text-primary font-medium">No suspicious IP clusters</p>
              <p className="text-sm text-text-muted">All IP ranges show normal usage patterns</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow className="border-border-subtle hover:bg-transparent">
                    <TableHead className="text-text-secondary">IP Range</TableHead>
                    <TableHead className="text-text-secondary">Associated Tenants</TableHead>
                    <TableHead className="text-text-secondary">Risk Level</TableHead>
                    <TableHead className="text-text-secondary">Common Patterns</TableHead>
                    <TableHead className="text-text-secondary">First Seen</TableHead>
                    <TableHead className="text-text-secondary">Last Seen</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data?.ipClusters.map((cluster) => (
                    <TableRow key={cluster.id} className="border-border-subtle">
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <MapPin className="w-4 h-4 text-text-muted" />
                          <code className="text-sm font-mono text-text-primary">{cluster.ipRange}</code>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="flex flex-wrap gap-1">
                          {cluster.associatedTenants.map((tenant) => (
                            <Badge key={tenant} variant="secondary" className="text-xs">
                              {tenant}
                            </Badge>
                          ))}
                        </div>
                      </TableCell>
                      <TableCell>{getRiskLevelBadge(cluster.riskLevel)}</TableCell>
                      <TableCell>
                        <div className="flex flex-wrap gap-1">
                          {cluster.commonPatterns.map((pattern, idx) => (
                            <span
                              key={idx}
                              className="text-xs px-2 py-0.5 rounded-full bg-bg-secondary text-text-secondary"
                            >
                              {pattern}
                            </span>
                          ))}
                        </div>
                      </TableCell>
                      <TableCell className="text-sm text-text-muted">
                        {new Date(cluster.firstSeen).toLocaleDateString()}
                      </TableCell>
                      <TableCell className="text-sm text-text-muted">
                        {new Date(cluster.lastSeen).toLocaleDateString()}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Wash Usage Detection */}
      <Card className="glass-card">
        <CardHeader>
          <CardTitle className="text-text-primary flex items-center gap-2">
            <Activity className="w-5 h-5 text-amber-400" />
            Wash-Usage Detection
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="p-6 space-y-4">
              {Array.from({ length: 3 }).map((_, i) => (
                <Skeleton key={i} className="h-20 w-full" />
              ))}
            </div>
          ) : data?.washUsagePatterns.length === 0 ? (
            <div className="text-center py-8">
              <CheckCircle className="w-12 h-12 text-emerald-400 mx-auto mb-3" />
              <p className="text-text-primary font-medium">No wash usage detected</p>
              <p className="text-sm text-text-muted">No reciprocal execution patterns found</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow className="border-border-subtle hover:bg-transparent">
                    <TableHead className="text-text-secondary">Tenant Pair</TableHead>
                    <TableHead className="text-text-secondary">Function</TableHead>
                    <TableHead className="text-text-secondary">Pattern</TableHead>
                    <TableHead className="text-text-secondary">Confidence</TableHead>
                    <TableHead className="text-text-secondary">Reciprocal Execs</TableHead>
                    <TableHead className="text-text-secondary">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data?.washUsagePatterns.map((pattern) => (
                    <TableRow key={pattern.id} className="border-border-subtle">
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <Badge variant="secondary" className="text-xs">{pattern.tenantA}</Badge>
                          <ChevronRight className="w-4 h-4 text-text-muted" />
                          <Badge variant="secondary" className="text-xs">{pattern.tenantB}</Badge>
                        </div>
                      </TableCell>
                      <TableCell className="font-medium text-text-primary">{pattern.function}</TableCell>
                      <TableCell className="text-sm text-text-secondary max-w-xs truncate">
                        {pattern.pattern}
                      </TableCell>
                      <TableCell>
                        <Badge className={`${getConfidenceBg(pattern.confidence)} ${getConfidenceColor(pattern.confidence)} border-0`}>
                          {pattern.confidence}%
                        </Badge>
                      </TableCell>
                      <TableCell className="font-mono text-text-primary">
                        {pattern.reciprocalExecutions.toLocaleString()}
                      </TableCell>
                      <TableCell>
                        <div className="flex gap-2">
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-8 text-text-muted hover:text-text-primary"
                            onClick={() => handleInvestigate(pattern.id, "wash usage")}
                          >
                            <Search className="w-4 h-4 mr-1" />
                            Details
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-8 text-red-400 hover:text-red-500 hover:bg-red-500/10"
                            onClick={() => handleBlock(pattern.id, "wash usage")}
                          >
                            <Ban className="w-4 h-4 mr-1" />
                            Block
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
