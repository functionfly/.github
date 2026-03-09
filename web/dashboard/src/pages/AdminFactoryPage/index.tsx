import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Activity,
  Play,
  RefreshCw,
  CheckCircle,
  XCircle,
  AlertCircle,
  Clock,
  TrendingUp,
  Package,
  Search,
  ChevronRight,
  Zap,
  AlertTriangle,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import {
  factoryApi,
  type FactoryStatus,
  type PendingReview,
  type FactoryRun,
} from "@/api/factory";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

/**
 * AdminFactoryPage - Factory status monitoring and review queue management
 *
 * Displays:
 * - Factory health status and statistics
 * - Pipeline run statistics
 * - Recent pipeline runs
 * - Manual pipeline run trigger
 * - Review queue management UI
 */
export function AdminFactoryPage() {
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState("status");
  const [selectedReview, setSelectedReview] = useState<PendingReview | null>(null);
  const [rejectReason, setRejectReason] = useState("");
  const [showRejectDialog, setShowRejectDialog] = useState(false);

  // Fetch factory status (queryFn must not return undefined for React Query v5)
  const {
    data: factoryStatus,
    isLoading: statusLoading,
    error: statusError,
    refetch: refetchStatus,
  } = useQuery({
    queryKey: ["factory-status"],
    queryFn: async () => {
      const data = await factoryApi.getStatus();
      return data ?? null;
    },
    refetchInterval: 30000, // Refresh every 30 seconds
  });

  // Fetch pending reviews (queryFn must not return undefined for React Query v5)
  const {
    data: pendingReviews,
    isLoading: reviewsLoading,
    refetch: refetchReviews,
  } = useQuery({
    queryKey: ["factory-pending-reviews"],
    queryFn: async () => {
      const data = await factoryApi.listPendingReviews({ limit: 50 });
      return data ?? { reviews: [], total: 0, limit: 50, offset: 0 };
    },
    refetchInterval: 15000, // Refresh every 15 seconds
  });

  // Pipeline run mutation
  const runPipelineMutation = useMutation({
    mutationFn: factoryApi.triggerPipelineRun,
    onSuccess: (data) => {
      toast.success("Pipeline run initiated", {
        description: data.run?.id ? `Run ID: ${data.run.id}` : "Pipeline started successfully",
      });
      refetchStatus();
    },
    onError: (error: Error) => {
      toast.error("Failed to start pipeline", {
        description: error.message || "An error occurred",
      });
    },
  });

  // Approve opportunity mutation
  const approveMutation = useMutation({
    mutationFn: ({ id }: { id: string }) => factoryApi.approveOpportunity(id),
    onSuccess: () => {
      toast.success("Opportunity approved");
      queryClient.invalidateQueries({ queryKey: ["factory-pending-reviews"] });
      queryClient.invalidateQueries({ queryKey: ["factory-status"] });
      setSelectedReview(null);
    },
    onError: (error: Error) => {
      toast.error("Failed to approve opportunity", {
        description: error.message,
      });
    },
  });

  // Reject opportunity mutation
  const rejectMutation = useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) =>
      factoryApi.rejectOpportunity(id, reason),
    onSuccess: () => {
      toast.success("Opportunity rejected");
      queryClient.invalidateQueries({ queryKey: ["factory-pending-reviews"] });
      queryClient.invalidateQueries({ queryKey: ["factory-status"] });
      setShowRejectDialog(false);
      setSelectedReview(null);
      setRejectReason("");
    },
    onError: (error: Error) => {
      toast.error("Failed to reject opportunity", {
        description: error.message,
      });
    },
  });

  const handleRunPipeline = () => {
    runPipelineMutation.mutate();
  };

  const handleApprove = (review: PendingReview) => {
    approveMutation.mutate({ id: review.id });
  };

  const handleRejectClick = (review: PendingReview) => {
    setSelectedReview(review);
    setShowRejectDialog(true);
  };

  const handleConfirmReject = () => {
    if (!selectedReview || !rejectReason.trim()) return;
    rejectMutation.mutate({ id: selectedReview.id, reason: rejectReason });
  };

  const getStatusColor = (status?: string) => {
    switch (status) {
      case "healthy":
      case "completed":
      case "approved":
        return "text-emerald-600 dark:text-emerald-400";
      case "running":
        return "text-blue-600 dark:text-blue-400";
      case "failed":
      case "rejected":
        return "text-red-600 dark:text-red-400";
      default:
        return "text-yellow-600 dark:text-yellow-400";
    }
  };

  const getStatusBadgeVariant = (status?: string) => {
    switch (status) {
      case "healthy":
      case "completed":
      case "approved":
        return "default";
      case "running":
        return "outline";
      case "failed":
      case "rejected":
        return "destructive";
      default:
        return "secondary";
    }
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return "N/A";
    return new Date(dateString).toLocaleString();
  };

  const formatRelativeTime = (dateString?: string) => {
    if (!dateString) return "N/A";
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffMins < 1) return "Just now";
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    return `${diffDays}d ago`;
  };

  // Calculate success rate
  const calculateSuccessRate = (run?: FactoryRun) => {
    if (!run || run.functions_created === 0) return 0;
    const total = run.functions_created + run.functions_failed;
    if (total === 0) return 0;
    return Math.round((run.functions_created / total) * 100);
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-foreground">AI Function Factory</h1>
          <p className="text-muted-foreground">
            Monitor factory status and manage opportunity reviews
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => refetchStatus()}
            disabled={statusLoading}
          >
            <RefreshCw className={cn("h-4 w-4 mr-2", statusLoading && "animate-spin")} />
            Refresh
          </Button>
          <Button
            onClick={handleRunPipeline}
            disabled={runPipelineMutation.isPending}
          >
            <Play className={cn("h-4 w-4 mr-2", runPipelineMutation.isPending && "animate-spin")} />
            {runPipelineMutation.isPending ? "Starting..." : "Run Pipeline"}
          </Button>
        </div>
      </div>

      {/* Error Alert */}
      {statusError && (
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>
            Failed to load factory status: {(statusError as Error).message}
          </AlertDescription>
        </Alert>
      )}

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="status">Status</TabsTrigger>
          <TabsTrigger value="reviews">
            Reviews
            {pendingReviews?.total ? (
              <Badge variant="destructive" className="ml-2 h-5 px-1.5">
                {pendingReviews.total}
              </Badge>
            ) : null}
          </TabsTrigger>
        </TabsList>

        {/* Status Tab */}
        <TabsContent value="status" className="space-y-6 mt-6">
          {statusLoading ? (
            <StatusSkeleton />
          ) : factoryStatus ? (
            <>
              {/* Stats Cards */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <StatCard
                  title="Total Runs"
                  value={factoryStatus.totals.runs}
                  icon={Activity}
                  description="Pipeline executions"
                />
                <StatCard
                  title="Published Functions"
                  value={factoryStatus.totals.published}
                  icon={Package}
                  description="Functions in registry"
                />
                <StatCard
                  title="Opportunities"
                  value={factoryStatus.totals.opportunities}
                  icon={Search}
                  description="Discovered opportunities"
                />
                <StatCard
                  title="Auto-Publish"
                  value={factoryStatus.totals.autopublish ? "Enabled" : "Disabled"}
                  icon={Zap}
                  description={`Quality: ${factoryStatus.totals.quality_minimum}% | Tests: ${factoryStatus.totals.test_minimum}%`}
                  status={factoryStatus.totals.autopublish ? "success" : "warning"}
                />
              </div>

              {/* Latest Run Card */}
              {factoryStatus.latest_run && (
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                      <Clock className="h-5 w-5" />
                      Latest Pipeline Run
                    </CardTitle>
                    <CardDescription>
                      ID: {factoryStatus.latest_run.id}
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                      <div className="space-y-1">
                        <p className="text-sm text-muted-foreground">Status</p>
                        <Badge variant={getStatusBadgeVariant(factoryStatus.latest_run.status)}>
                          {factoryStatus.latest_run.status}
                        </Badge>
                      </div>
                      <div className="space-y-1">
                        <p className="text-sm text-muted-foreground">Started</p>
                        <p className="font-medium">{formatDate(factoryStatus.latest_run.started_at)}</p>
                      </div>
                      <div className="space-y-1">
                        <p className="text-sm text-muted-foreground">Completed</p>
                        <p className="font-medium">
                          {factoryStatus.latest_run.completed_at
                            ? formatDate(factoryStatus.latest_run.completed_at)
                            : "In Progress"}
                        </p>
                      </div>
                      <div className="space-y-1">
                        <p className="text-sm text-muted-foreground">Success Rate</p>
                        <p className={cn(
                          "font-medium",
                          calculateSuccessRate(factoryStatus.latest_run) >= 80
                            ? "text-emerald-600"
                            : calculateSuccessRate(factoryStatus.latest_run) >= 50
                            ? "text-yellow-600"
                            : "text-red-600"
                        )}>
                          {calculateSuccessRate(factoryStatus.latest_run)}%
                        </p>
                      </div>
                    </div>

                    <div className="mt-6 grid grid-cols-2 md:grid-cols-5 gap-4 pt-4 border-t">
                      <div className="text-center">
                        <p className="text-2xl font-bold">{factoryStatus.latest_run.opportunities_found}</p>
                        <p className="text-xs text-muted-foreground">Found</p>
                      </div>
                      <div className="text-center">
                        <p className="text-2xl font-bold text-emerald-600">
                          {factoryStatus.latest_run.opportunities_approved}
                        </p>
                        <p className="text-xs text-muted-foreground">Approved</p>
                      </div>
                      <div className="text-center">
                        <p className="text-2xl font-bold text-red-600">
                          {factoryStatus.latest_run.opportunities_rejected}
                        </p>
                        <p className="text-xs text-muted-foreground">Rejected</p>
                      </div>
                      <div className="text-center">
                        <p className="text-2xl font-bold text-blue-600">
                          {factoryStatus.latest_run.functions_created}
                        </p>
                        <p className="text-xs text-muted-foreground">Created</p>
                      </div>
                      <div className="text-center">
                        <p className="text-2xl font-bold text-red-600">
                          {factoryStatus.latest_run.functions_failed}
                        </p>
                        <p className="text-xs text-muted-foreground">Failed</p>
                      </div>
                    </div>

                    {factoryStatus.latest_run.error && (
                      <Alert variant="destructive" className="mt-4">
                        <AlertTriangle className="h-4 w-4" />
                        <AlertDescription>{factoryStatus.latest_run.error}</AlertDescription>
                      </Alert>
                    )}
                  </CardContent>
                </Card>
              )}

              {/* Config Card */}
              <Card>
                <CardHeader>
                  <CardTitle>Configuration</CardTitle>
                  <CardDescription>Current factory settings</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                    <div className="space-y-1">
                      <p className="text-sm text-muted-foreground">Agent ID</p>
                      <p className="font-mono text-sm">{factoryStatus.config.agent_id}</p>
                    </div>
                    <div className="space-y-1">
                      <p className="text-sm text-muted-foreground">Discovery Batch Size</p>
                      <p className="font-medium">{factoryStatus.config.discovery_batch_size}</p>
                    </div>
                    <div className="space-y-1">
                      <p className="text-sm text-muted-foreground">Min Quality Score</p>
                      <p className="font-medium">{factoryStatus.config.minimum_quality_score}%</p>
                    </div>
                    <div className="space-y-1">
                      <p className="text-sm text-muted-foreground">Min Test Score</p>
                      <p className="font-medium">{factoryStatus.config.minimum_test_score}%</p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </>
          ) : null}
        </TabsContent>

        {/* Reviews Tab */}
        <TabsContent value="reviews" className="space-y-6 mt-6">
          <Card>
            <CardHeader>
              <CardTitle>Pending Reviews</CardTitle>
              <CardDescription>
                Opportunities awaiting manual review before publishing
              </CardDescription>
            </CardHeader>
            <CardContent>
              {reviewsLoading ? (
                <div className="space-y-4">
                  {[1, 2, 3].map((i) => (
                    <Skeleton key={i} className="h-20 w-full" />
                  ))}
                </div>
              ) : pendingReviews?.reviews.length === 0 ? (
                <div className="text-center py-12">
                  <CheckCircle className="h-12 w-12 text-emerald-500 mx-auto mb-4" />
                  <h3 className="text-lg font-medium">All Caught Up!</h3>
                  <p className="text-muted-foreground">
                    No pending reviews at the moment.
                  </p>
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Title</TableHead>
                      <TableHead>Source</TableHead>
                      <TableHead>Quality</TableHead>
                      <TableHead>Tests</TableHead>
                      <TableHead>Submitted</TableHead>
                      <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {pendingReviews?.reviews.map((review) => (
                      <TableRow key={review.id}>
                        <TableCell className="font-medium max-w-xs truncate">
                          {review.title}
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline">{review.source}</Badge>
                        </TableCell>
                        <TableCell>
                          <span
                            className={cn(
                              review.quality_score
                                ? review.quality_score >= 80
                                  ? "text-emerald-600"
                                  : review.quality_score >= 60
                                  ? "text-yellow-600"
                                  : "text-red-600"
                                : "text-muted-foreground"
                            )}
                          >
                            {review.quality_score ? `${review.quality_score}%` : "N/A"}
                          </span>
                        </TableCell>
                        <TableCell>
                          <span
                            className={cn(
                              review.test_score
                                ? review.test_score >= 80
                                  ? "text-emerald-600"
                                  : review.test_score >= 60
                                  ? "text-yellow-600"
                                  : "text-red-600"
                                : "text-muted-foreground"
                            )}
                          >
                            {review.test_score ? `${review.test_score}%` : "N/A"}
                          </span>
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {formatRelativeTime(review.review_requested_at || review.created_at)}
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex items-center justify-end gap-2">
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => handleRejectClick(review)}
                            >
                              <XCircle className="h-4 w-4 mr-1" />
                              Reject
                            </Button>
                            <Button
                              size="sm"
                              onClick={() => handleApprove(review)}
                              disabled={approveMutation.isPending}
                            >
                              <CheckCircle className="h-4 w-4 mr-1" />
                              Approve
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Reject Dialog */}
      <Dialog open={showRejectDialog} onOpenChange={setShowRejectDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Reject Opportunity</DialogTitle>
            <DialogDescription>
              Please provide a reason for rejecting this opportunity.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="reject-reason">Rejection Reason</Label>
              <Textarea
                id="reject-reason"
                placeholder="Enter the reason for rejection..."
                value={rejectReason}
                onChange={(e) => setRejectReason(e.target.value)}
                rows={4}
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowRejectDialog(false)}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleConfirmReject}
              disabled={!rejectReason.trim() || rejectMutation.isPending}
            >
              {rejectMutation.isPending ? "Rejecting..." : "Reject Opportunity"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

/**
 * StatCard - Display a single statistic
 */
interface StatCardProps {
  title: string;
  value: string | number;
  icon: React.ComponentType<{ className?: string }>;
  description?: string;
  status?: "success" | "warning" | "error";
}

function StatCard({ title, value, icon: Icon, description, status }: StatCardProps) {
  return (
    <Card>
      <CardContent className="pt-6">
        <div className="flex items-start justify-between">
          <div className="space-y-1">
            <p className="text-sm text-muted-foreground">{title}</p>
            <p className={cn(
              "text-2xl font-bold",
              status === "success" && "text-emerald-600 dark:text-emerald-400",
              status === "warning" && "text-yellow-600 dark:text-yellow-400",
              status === "error" && "text-red-600 dark:text-red-400"
            )}>
              {value}
            </p>
            {description && (
              <p className="text-xs text-muted-foreground">{description}</p>
            )}
          </div>
          <div className={cn(
            "p-2 rounded-lg",
            status === "success" && "bg-emerald-500/10 text-emerald-600",
            status === "warning" && "bg-yellow-500/10 text-yellow-600",
            status === "error" && "bg-red-500/10 text-red-600",
            !status && "bg-brand-500/10 text-brand-600"
          )}>
            <Icon className="h-5 w-5" />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

/**
 * StatusSkeleton - Loading skeleton for status tab
 */
function StatusSkeleton() {
  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {[1, 2, 3, 4].map((i) => (
          <Skeleton key={i} className="h-32" />
        ))}
      </div>
      <Skeleton className="h-64" />
      <Skeleton className="h-48" />
    </div>
  );
}

export default AdminFactoryPage;
