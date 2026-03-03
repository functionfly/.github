import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import {
  DollarSign,
  TrendingUp,
  AlertTriangle,
  ArrowLeft,
  RefreshCw,
  Trophy,
  Activity,
  Users,
  CheckCircle,
  ChevronRight,
  Search,
  Ban,
  Eye,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { StatCard } from "@/components/common/StatCard";
import { toast } from "sonner";
import { oversightApi, type EconomicLeaderboardData, type RevenueGenerator } from "@/api/admin";

export function AdminEconomicLeaderboardPage() {
  const navigate = useNavigate();

  const { data, isLoading, refetch } = useQuery<EconomicLeaderboardData>({
    queryKey: ["admin-economic-leaderboard"],
    queryFn: () => oversightApi.getEconomicLeaderboard(),
  });

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: "USD",
      minimumFractionDigits: 2,
    }).format(amount);
  };

  const handleInvestigate = async (id: string) => {
    try {
      await oversightApi.investigateEntity("revenue", id);
      toast.success(`Investigation opened for revenue entry ${id}`);
    } catch (error) {
      toast.error(`Failed to open investigation`);
    }
  };

  const handleBlock = async (id: string) => {
    try {
      await oversightApi.blockEntity("account", id);
      toast.success(`Account ${id} has been blocked`);
    } catch (error) {
      toast.error(`Failed to block account`);
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
            <h1 className="text-2xl font-bold text-text-primary">Economic Leaderboard</h1>
            <p className="text-sm text-text-secondary">Monitor top performers and detect economic manipulation</p>
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
          title="Total Revenue (30d)"
          value={isLoading ? "—" : "$1.2M"}
          icon={<DollarSign className="w-5 h-5 text-emerald-400" />}
          trend="up"
          change={{ value: 12.5, label: "vs last month" }}
        />
        <StatCard
          title="Top Performers"
          value={isLoading ? "—" : data?.topRevenueGenerators.length.toString() || "0"}
          icon={<Trophy className="w-5 h-5 text-amber-400" />}
          trend="neutral"
          change={{ value: 0, label: "active functions" }}
        />
        <StatCard
          title="Suspicious Growth"
          value={isLoading ? "—" : data?.suspiciousGrowth.length.toString() || "0"}
          icon={<AlertTriangle className="w-5 h-5 text-red-400" />}
          trend="down"
          change={{ value: -2, label: "vs last week" }}
        />
        <StatCard
          title="Boosting Detected"
          value={isLoading ? "—" : data?.artificialBoosting.length.toString() || "0"}
          icon={<Activity className="w-5 h-5 text-purple-400" />}
          trend="neutral"
          change={{ value: 0, label: "cases pending" }}
        />
      </div>

      {/* Top Revenue Generators */}
      <Card className="glass-card">
        <CardHeader>
          <CardTitle className="text-text-primary flex items-center gap-2">
            <Trophy className="w-5 h-5 text-amber-400" />
            Top Revenue Generators (30 Days)
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="p-6 space-y-4">
              {Array.from({ length: 5 }).map((_, i) => (
                <Skeleton key={i} className="h-16 w-full" />
              ))}
            </div>
          ) : data?.topRevenueGenerators.length === 0 ? (
            <div className="text-center py-8">
              <CheckCircle className="w-12 h-12 text-emerald-400 mx-auto mb-3" />
              <p className="text-text-primary font-medium">No revenue data available</p>
              <p className="text-sm text-text-muted">No revenue-generating functions found</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow className="border-border-subtle hover:bg-transparent">
                    <TableHead className="text-text-secondary">Rank</TableHead>
                    <TableHead className="text-text-secondary">Tenant / Function</TableHead>
                    <TableHead className="text-text-secondary">Revenue (30d)</TableHead>
                    <TableHead className="text-text-secondary">Executions</TableHead>
                    <TableHead className="text-text-secondary">Growth</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data?.topRevenueGenerators.map((generator: RevenueGenerator) => (
                    <TableRow key={generator.id} className="border-border-subtle">
                      <TableCell>
                        <div className="flex items-center gap-2">
                          {generator.rank <= 3 ? (
                            <Badge className={`
                              ${generator.rank === 1 ? "bg-amber-500/20 text-amber-400" : ""}
                              ${generator.rank === 2 ? "bg-slate-500/20 text-slate-400" : ""}
                              ${generator.rank === 3 ? "bg-orange-500/20 text-orange-400" : ""}
                            `}>
                              #{generator.rank}
                            </Badge>
                          ) : (
                            <span className="text-text-muted">#{generator.rank}</span>
                          )}
                        </div>
                      </TableCell>
                      <TableCell className="font-medium text-text-primary">
                        {generator.tenantFunction}
                      </TableCell>
                      <TableCell className="font-mono text-emerald-400">
                        {formatCurrency(generator.revenue30d)}
                      </TableCell>
                      <TableCell className="text-text-secondary">
                        {generator.executionCount.toLocaleString()}
                      </TableCell>
                      <TableCell>
                        <div className={`flex items-center gap-1 ${
                          generator.growthRate >= 0 ? "text-emerald-400" : "text-red-400"
                        }`}>
                          <TrendingUp className="w-4 h-4" />
                          <span>{generator.growthRate > 0 ? "+" : ""}{generator.growthRate.toFixed(1)}%</span>
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

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Suspicious Growth Alerts */}
        <Card className="glass-card">
          <CardHeader>
            <CardTitle className="text-text-primary flex items-center gap-2">
              <AlertTriangle className="w-5 h-5 text-amber-400" />
              Suspicious Growth Alerts
            </CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            {isLoading ? (
              <div className="p-6 space-y-4">
                {Array.from({ length: 3 }).map((_, i) => (
                  <Skeleton key={i} className="h-24 w-full" />
                ))}
              </div>
            ) : data?.suspiciousGrowth.length === 0 ? (
              <div className="text-center py-8">
                <CheckCircle className="w-12 h-12 text-emerald-400 mx-auto mb-3" />
                <p className="text-text-primary font-medium">No suspicious growth detected</p>
                <p className="text-sm text-text-muted">All growth patterns appear natural</p>
              </div>
            ) : (
              <div className="divide-y divide-border-subtle">
                {data?.suspiciousGrowth.map((alert) => (
                  <div key={alert.id} className="p-4 hover:bg-bg-hover transition-colors">
                    <div className="flex items-start justify-between mb-2">
                      <div>
                        <p className="font-medium text-text-primary">{alert.tenantFunction}</p>
                        <Badge className="bg-amber-500/10 text-amber-400 border-amber-500/20 mt-1">
                          {alert.pattern}
                        </Badge>
                      </div>
                      <AlertTriangle className="w-5 h-5 text-amber-400" />
                    </div>
                    <p className="text-sm text-text-secondary mb-2">{alert.details}</p>
                    <p className="text-xs text-text-muted">
                      Detected: {new Date(alert.detectedAt).toLocaleString()}
                    </p>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Artificial Boosting Detection */}
        <Card className="glass-card">
          <CardHeader>
            <CardTitle className="text-text-primary flex items-center gap-2">
              <Activity className="w-5 h-5 text-purple-400" />
              Artificial Boosting Detection
            </CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            {isLoading ? (
              <div className="p-6 space-y-4">
                {Array.from({ length: 3 }).map((_, i) => (
                  <Skeleton key={i} className="h-24 w-full" />
                ))}
              </div>
            ) : data?.artificialBoosting.length === 0 ? (
              <div className="text-center py-8">
                <CheckCircle className="w-12 h-12 text-emerald-400 mx-auto mb-3" />
                <p className="text-text-primary font-medium">No boosting detected</p>
                <p className="text-sm text-text-muted">No artificial manipulation found</p>
              </div>
            ) : (
              <div className="divide-y divide-border-subtle">
                {data?.artificialBoosting.map((boosting) => (
                  <div key={boosting.id} className="p-4 hover:bg-bg-hover transition-colors">
                    <div className="flex items-start justify-between mb-2">
                      <div>
                        <p className="font-medium text-text-primary">{boosting.function}</p>
                        <p className="text-sm text-text-secondary">{boosting.detectedPattern}</p>
                      </div>
                      <Badge className={`${
                        boosting.confidence >= 80 ? "bg-red-500/10 text-red-400" :
                        boosting.confidence >= 60 ? "bg-amber-500/10 text-amber-400" :
                        "bg-blue-500/10 text-blue-400"
                      }`}>
                        {boosting.confidence}% confidence
                      </Badge>
                    </div>
                    <div className="flex items-center gap-2 mt-2">
                      <Users className="w-4 h-4 text-text-muted" />
                      <span className="text-sm text-text-muted">
                        {boosting.relatedAccounts.length} related accounts
                      </span>
                    </div>
                    <div className="flex gap-2 mt-3">
                      <Button
                        variant="outline"
                        size="sm"
                        className="border-border-default hover:bg-bg-hover"
                        onClick={() => handleInvestigate(boosting.id)}
                      >
                        <Eye className="w-3 h-3 mr-1" />
                        Investigate
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        className="border-red-500/20 text-red-400 hover:bg-red-500/10"
                        onClick={() => handleBlock(boosting.id)}
                      >
                        <Ban className="w-3 h-3 mr-1" />
                        Block
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
