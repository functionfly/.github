/**
 * FactoryDashboard Component
 * User-facing AI Function Factory pipeline monitoring dashboard
 */

import { useState, useEffect, useCallback } from 'react';
import {
  Activity,
  Play,
  RefreshCw,
  CheckCircle,
  XCircle,
  Clock,
  AlertTriangle,
  Filter,
  Pause,
  TrendingUp,
  Sparkles,
  ListFilter,
  Check,
  X,
} from 'lucide-react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { cn } from '@/lib/utils';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  factoryApi,
  type FactoryStatus,
  type FactoryRun,
  type Opportunity,
} from '@/api/factory';

interface FactoryDashboardProps {
  className?: string;
}

const OPPORTUNITY_STATUS_OPTIONS = [
  { value: 'all', label: 'All' },
  { value: 'pending_review', label: 'Pending Review' },
  { value: 'approved', label: 'Approved' },
  { value: 'rejected', label: 'Rejected' },
];

const SOURCE_FILTER_OPTIONS = [
  { value: 'all', label: 'All Sources' },
  { value: 'github', label: 'GitHub' },
  { value: 'npm', label: 'NPM' },
  { value: 'stackoverflow', label: 'Stack Overflow' },
  { value: 'community', label: 'Community' },
];

export function FactoryDashboard({ className }: FactoryDashboardProps) {
  const { t } = useTranslation();
  const [factoryStatus, setFactoryStatus] = useState<FactoryStatus | null>(null);
  const [opportunities, setOpportunities] = useState<Opportunity[]>([]);
  const [opportunitiesTotal, setOpportunitiesTotal] = useState(0);
  const [isLoading, setIsLoading] = useState(true);
  const [isTriggering, setIsTriggering] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [statusFilter, setStatusFilter] = useState('all');
  const [sourceFilter, setSourceFilter] = useState('all');
  const [activeTab, setActiveTab] = useState('overview');

  const fetchFactoryStatus = useCallback(async () => {
    try {
      const data = await factoryApi.getStatus();
      setFactoryStatus(data);
    } catch (err) {
      console.error('Failed to fetch factory status:', err);
      toast.error('Failed to load factory status');
    }
  }, []);

  const fetchOpportunities = useCallback(async () => {
    try {
      const params: { status?: string; source?: string; limit: number } = {
        limit: 20,
      };
      if (statusFilter !== 'all') {
        params.status = statusFilter;
      }
      if (sourceFilter !== 'all') {
        params.source = sourceFilter;
      }
      const data = await factoryApi.listOpportunities(params);
      setOpportunities(data.opportunities);
      setOpportunitiesTotal(data.total);
    } catch (err) {
      console.error('Failed to fetch opportunities:', err);
      toast.error('Failed to load opportunities');
    }
  }, [statusFilter, sourceFilter]);

  const fetchAllData = useCallback(async () => {
    setIsLoading(true);
    await Promise.all([fetchFactoryStatus(), fetchOpportunities()]);
    setIsLoading(false);
  }, [fetchFactoryStatus, fetchOpportunities]);

  useEffect(() => {
    fetchAllData();
  }, [fetchAllData]);

  useEffect(() => {
    if (autoRefresh) {
      const interval = setInterval(fetchFactoryStatus, 15000);
      return () => clearInterval(interval);
    }
  }, [autoRefresh, fetchFactoryStatus]);

  const handleTriggerPipeline = async () => {
    setIsTriggering(true);
    try {
      const response = await factoryApi.triggerPipelineRun();
      toast.success('Pipeline run started', {
        description: `Run ID: ${response.run.id}`,
      });
      fetchFactoryStatus();
    } catch (err) {
      console.error('Failed to trigger pipeline:', err);
      toast.error('Failed to start pipeline run');
    } finally {
      setIsTriggering(false);
    }
  };

  const handleApproveOpportunity = async (id: string) => {
    try {
      await factoryApi.approveOpportunity(id);
      toast.success('Opportunity approved');
      fetchOpportunities();
    } catch (err) {
      console.error('Failed to approve opportunity:', err);
      toast.error('Failed to approve opportunity');
    }
  };

  const handleRejectOpportunity = async (id: string, reason: string) => {
    try {
      await factoryApi.rejectOpportunity(id, reason);
      toast.success('Opportunity rejected');
      fetchOpportunities();
    } catch (err) {
      console.error('Failed to reject opportunity:', err);
      toast.error('Failed to reject opportunity');
    }
  };

  const getStatusBadgeVariant = (status: string) => {
    switch (status) {
      case 'completed':
      case 'approved':
        return 'success';
      case 'running':
        return 'default';
      case 'failed':
      case 'rejected':
        return 'destructive';
      case 'pending_review':
        return 'warning';
      default:
        return 'secondary';
    }
  };

  const calculateSuccessRate = (run?: FactoryRun) => {
    if (!run || run.functions_created === 0) return 0;
    const total = run.functions_created + run.functions_failed;
    if (total === 0) return 0;
    return Math.round((run.functions_created / total) * 100);
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return 'N/A';
    return new Date(dateString).toLocaleString();
  };

  const formatRelativeTime = (dateString?: string) => {
    if (!dateString) return 'N/A';
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    return `${diffDays}d ago`;
  };

  const latestRun = factoryStatus?.latest_run;

  return (
    <div className={cn('space-y-6', className)}>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold flex items-center gap-2">
            <Sparkles className="h-6 w-6 text-brand-500" />
            AI Function Factory
          </h1>
          <p className="text-sm text-text-secondary mt-1">
            Monitor pipeline runs and manage opportunities
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setAutoRefresh(!autoRefresh)}
            className={cn('flex items-center gap-2', autoRefresh && 'bg-success/10 border-success/30')}
          >
            {autoRefresh ? <Play className="h-4 w-4" /> : <Pause className="h-4 w-4" />}
            {autoRefresh ? 'Auto' : 'Paused'}
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={fetchAllData}
            disabled={isLoading}
            className="flex items-center gap-2"
          >
            <RefreshCw className={cn('h-4 w-4', isLoading && 'animate-spin')} />
            Refresh
          </Button>
          <Button
            size="sm"
            onClick={handleTriggerPipeline}
            isLoading={isTriggering}
            className="flex items-center gap-2"
          >
            <Play className="h-4 w-4" />
            Run Pipeline
          </Button>
        </div>
      </div>

      {latestRun?.status === 'failed' && latestRun.error && (
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>{latestRun.error}</AlertDescription>
        </Alert>
      )}

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="opportunities">Opportunities</TabsTrigger>
          <TabsTrigger value="runs">Pipeline Runs</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-6">
          {isLoading && !factoryStatus ? (
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              {[1, 2, 3, 4].map((i) => (
                <Card key={i}>
                  <CardContent className="pt-6">
                    <Skeleton className="h-20" />
                  </CardContent>
                </Card>
              ))}
            </div>
          ) : (
            <>
              <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                <Card>
                  <CardContent className="pt-6">
                    <div className="flex items-start justify-between">
                      <div>
                        <p className="text-sm text-text-secondary">Total Runs</p>
                        <p className="text-2xl font-bold">{factoryStatus?.totals.runs ?? 0}</p>
                      </div>
                      <div className="p-2 bg-blue-500/10 rounded-lg">
                        <Activity className="h-5 w-5 text-blue-500" />
                      </div>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardContent className="pt-6">
                    <div className="flex items-start justify-between">
                      <div>
                        <p className="text-sm text-text-secondary">Published</p>
                        <p className="text-2xl font-bold text-success">
                          {factoryStatus?.totals.published ?? 0}
                        </p>
                      </div>
                      <div className="p-2 bg-success/10 rounded-lg">
                        <CheckCircle className="h-5 w-5 text-success" />
                      </div>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardContent className="pt-6">
                    <div className="flex items-start justify-between">
                      <div>
                        <p className="text-sm text-text-secondary">Opportunities</p>
                        <p className="text-2xl font-bold">
                          {factoryStatus?.totals.opportunities ?? 0}
                        </p>
                      </div>
                      <div className="p-2 bg-purple-500/10 rounded-lg">
                        <TrendingUp className="h-5 w-5 text-purple-500" />
                      </div>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardContent className="pt-6">
                    <div className="flex items-start justify-between">
                      <div>
                        <p className="text-sm text-text-secondary">Auto-Publish</p>
                        <p
                          className={cn(
                            'text-2xl font-bold',
                            factoryStatus?.totals.autopublish ? 'text-success' : 'text-warning'
                          )}
                        >
                          {factoryStatus?.totals.autopublish ? 'On' : 'Off'}
                        </p>
                      </div>
                      <div
                        className={cn(
                          'p-2 rounded-lg',
                          factoryStatus?.totals.autopublish ? 'bg-success/10' : 'bg-warning/10'
                        )}
                      >
                        {factoryStatus?.totals.autopublish ? (
                          <CheckCircle className="h-5 w-5 text-success" />
                        ) : (
                          <XCircle className="h-5 w-5 text-warning" />
                        )}
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </div>

              {latestRun && (
                <Card>
                  <CardHeader>
                    <div className="flex items-center justify-between">
                      <CardTitle className="flex items-center gap-2">
                        <Clock className="h-5 w-5" />
                        Latest Pipeline Run
                      </CardTitle>
                      <Badge variant={getStatusBadgeVariant(latestRun.status)}>
                        {latestRun.status}
                      </Badge>
                    </div>
                    <CardDescription>Run ID: {latestRun.id}</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
                      <div className="space-y-1">
                        <p className="text-sm text-text-secondary">Started</p>
                        <p className="font-medium">{formatDate(latestRun.started_at)}</p>
                      </div>
                      <div className="space-y-1">
                        <p className="text-sm text-text-secondary">Completed</p>
                        <p className="font-medium">
                          {latestRun.completed_at ? formatDate(latestRun.completed_at) : 'In Progress'}
                        </p>
                      </div>
                      <div className="space-y-1">
                        <p className="text-sm text-text-secondary">Success Rate</p>
                        <p
                          className={cn(
                            'font-medium',
                            calculateSuccessRate(latestRun) >= 80
                              ? 'text-success'
                              : calculateSuccessRate(latestRun) >= 50
                                ? 'text-warning'
                                : 'text-error'
                          )}
                        >
                          {calculateSuccessRate(latestRun)}%
                        </p>
                      </div>
                      <div className="space-y-1">
                        <p className="text-sm text-text-secondary">Functions Created</p>
                        <p className="font-medium text-brand-500">{latestRun.functions_created}</p>
                      </div>
                    </div>

                    <div className="grid grid-cols-5 gap-4 pt-4 border-t">
                      <div className="text-center">
                        <p className="text-2xl font-bold">{latestRun.opportunities_found}</p>
                        <p className="text-xs text-text-secondary">Found</p>
                      </div>
                      <div className="text-center">
                        <p className="text-2xl font-bold text-success">
                          {latestRun.opportunities_approved}
                        </p>
                        <p className="text-xs text-text-secondary">Approved</p>
                      </div>
                      <div className="text-center">
                        <p className="text-2xl font-bold text-error">
                          {latestRun.opportunities_rejected}
                        </p>
                        <p className="text-xs text-text-secondary">Rejected</p>
                      </div>
                      <div className="text-center">
                        <p className="text-2xl font-bold text-brand-500">
                          {latestRun.functions_created}
                        </p>
                        <p className="text-xs text-text-secondary">Created</p>
                      </div>
                      <div className="text-center">
                        <p className="text-2xl font-bold text-error">
                          {latestRun.functions_failed}
                        </p>
                        <p className="text-xs text-text-secondary">Failed</p>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              )}

              {factoryStatus?.config && (
                <Card>
                  <CardHeader>
                    <CardTitle>Configuration</CardTitle>
                    <CardDescription>Current factory pipeline settings</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                      <div className="space-y-1">
                        <p className="text-sm text-text-secondary">Discovery Batch Size</p>
                        <p className="font-medium">{factoryStatus.config.discovery_batch_size}</p>
                      </div>
                      <div className="space-y-1">
                        <p className="text-sm text-text-secondary">Min Quality Score</p>
                        <p className="font-medium">{factoryStatus.config.minimum_quality_score}%</p>
                      </div>
                      <div className="space-y-1">
                        <p className="text-sm text-text-secondary">Min Test Score</p>
                        <p className="font-medium">{factoryStatus.config.minimum_test_score}%</p>
                      </div>
                      <div className="space-y-1">
                        <p className="text-sm text-text-secondary">Max Opportunities</p>
                        <p className="font-medium">
                          {factoryStatus.config.max_opportunities_per_run}
                        </p>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              )}
            </>
          )}
        </TabsContent>

        <TabsContent value="opportunities" className="space-y-4">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="flex items-center gap-2">
                  <ListFilter className="h-5 w-5" />
                  Recent Opportunities
                </CardTitle>
                <span className="text-sm text-text-secondary">
                  {opportunitiesTotal} total
                </span>
              </div>
              <CardDescription>Review and manage discovered opportunities</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex gap-4 mb-6">
                <div className="flex items-center gap-2">
                  <span className="text-sm text-text-secondary">Status:</span>
                  <div className="flex gap-1">
                    {OPPORTUNITY_STATUS_OPTIONS.map((option) => (
                      <Button
                        key={option.value}
                        variant={statusFilter === option.value ? 'default' : 'outline'}
                        size="sm"
                        onClick={() => setStatusFilter(option.value)}
                      >
                        {option.label}
                      </Button>
                    ))}
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-sm text-text-secondary">Source:</span>
                  <div className="flex gap-1">
                    {SOURCE_FILTER_OPTIONS.map((option) => (
                      <Button
                        key={option.value}
                        variant={sourceFilter === option.value ? 'default' : 'outline'}
                        size="sm"
                        onClick={() => setSourceFilter(option.value)}
                      >
                        {option.label}
                      </Button>
                    ))}
                  </div>
                </div>
              </div>

              {isLoading ? (
                <div className="space-y-2">
                  {[1, 2, 3, 4, 5].map((i) => (
                    <Skeleton key={i} className="h-16" />
                  ))}
                </div>
              ) : opportunities.length === 0 ? (
                <div className="text-center py-8 text-text-secondary">
                  <AlertTriangle className="h-8 w-8 mx-auto mb-2 opacity-50" />
                  <p>No opportunities found</p>
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Title</TableHead>
                      <TableHead>Source</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Quality</TableHead>
                      <TableHead>Test</TableHead>
                      <TableHead>Created</TableHead>
                      <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {opportunities.map((opp) => (
                      <TableRow key={opp.id}>
                        <TableCell className="font-medium max-w-xs truncate">
                          {opp.title}
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline">{opp.source}</Badge>
                        </TableCell>
                        <TableCell>
                          <Badge variant={getStatusBadgeVariant(opp.review_status)}>
                            {opp.review_status.replace('_', ' ')}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          {opp.quality_score != null ? (
                            <span
                              className={cn(
                                opp.quality_score >= 80
                                  ? 'text-success'
                                  : opp.quality_score >= 50
                                    ? 'text-warning'
                                    : 'text-error'
                              )}
                            >
                              {opp.quality_score}%
                            </span>
                          ) : (
                            <span className="text-text-secondary">-</span>
                          )}
                        </TableCell>
                        <TableCell>
                          {opp.test_score != null ? (
                            <span
                              className={cn(
                                opp.test_score >= 80
                                  ? 'text-success'
                                  : opp.test_score >= 50
                                    ? 'text-warning'
                                    : 'text-error'
                              )}
                            >
                              {opp.test_score}%
                            </span>
                          ) : (
                            <span className="text-text-secondary">-</span>
                          )}
                        </TableCell>
                        <TableCell className="text-text-secondary">
                          {formatRelativeTime(opp.created_at)}
                        </TableCell>
                        <TableCell className="text-right">
                          {opp.review_status === 'pending_review' && (
                            <div className="flex justify-end gap-2">
                              <Button
                                size="icon"
                                variant="ghost"
                                className="h-8 w-8 text-success hover:text-success hover:bg-success/10"
                                onClick={() => handleApproveOpportunity(opp.id)}
                              >
                                <Check className="h-4 w-4" />
                              </Button>
                              <Button
                                size="icon"
                                variant="ghost"
                                className="h-8 w-8 text-error hover:text-error hover:bg-error/10"
                                onClick={() =>
                                  handleRejectOpportunity(opp.id, 'Does not meet criteria')
                                }
                              >
                                <X className="h-4 w-4" />
                              </Button>
                            </div>
                          )}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="runs" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Activity className="h-5 w-5" />
                Pipeline Run History
              </CardTitle>
              <CardDescription>Historical runs and performance metrics</CardDescription>
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <div className="space-y-2">
                  {[1, 2, 3].map((i) => (
                    <Skeleton key={i} className="h-20" />
                  ))}
                </div>
              ) : latestRun ? (
                <div className="space-y-4">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Run ID</TableHead>
                        <TableHead>Status</TableHead>
                        <TableHead>Started</TableHead>
                        <TableHead>Success Rate</TableHead>
                        <TableHead>Functions</TableHead>
                        <TableHead>Opportunities</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      <TableRow>
                        <TableCell className="font-mono text-xs">{latestRun.id}</TableCell>
                        <TableCell>
                          <Badge variant={getStatusBadgeVariant(latestRun.status)}>
                            {latestRun.status}
                          </Badge>
                        </TableCell>
                        <TableCell>{formatDate(latestRun.started_at)}</TableCell>
                        <TableCell>
                          <span
                            className={cn(
                              calculateSuccessRate(latestRun) >= 80
                                ? 'text-success'
                                : calculateSuccessRate(latestRun) >= 50
                                  ? 'text-warning'
                                  : 'text-error'
                            )}
                          >
                            {calculateSuccessRate(latestRun)}%
                          </span>
                        </TableCell>
                        <TableCell>
                          <span className="text-success">{latestRun.functions_created}</span>
                          {' / '}
                          <span className="text-error">{latestRun.functions_failed}</span>
                        </TableCell>
                        <TableCell>
                          <span>{latestRun.opportunities_found}</span>
                          {' → '}
                          <span className="text-success">{latestRun.opportunities_approved}</span>
                          {' / '}
                          <span className="text-error">{latestRun.opportunities_rejected}</span>
                        </TableCell>
                      </TableRow>
                    </TableBody>
                  </Table>
                </div>
              ) : (
                <div className="text-center py-8 text-text-secondary">
                  <Activity className="h-8 w-8 mx-auto mb-2 opacity-50" />
                  <p>No pipeline runs yet</p>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}

export default FactoryDashboard;
