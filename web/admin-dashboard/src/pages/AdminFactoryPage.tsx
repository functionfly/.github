/**
 * Admin Factory Page
 * AI Function Factory status monitoring and review queue management
 * 
 * This page is organized into modular tabs:
 * - Status: Factory health, stats, and latest run info
 * - Opportunities: Browse all discovered opportunities
 * - Reviews: Pending opportunities awaiting approval
 * - Settings: Factory configuration
 */

import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Activity,
  Play,
  RefreshCw,
  AlertTriangle,
  Settings,
  Search,
  Eye,
  Package,
} from 'lucide-react';
import { useToastHelpers } from '@/components/ui/Toast';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import {
  factoryApi,
  type FactoryStatus,
  type PendingReview,
  type FactoryConfig,
  type Opportunity,
  type OpportunityListResponse,
  type PublishedFunction,
} from '@/lib/api/factory';
import {
  FactoryStatusTab,
  FactoryReviewsTab,
  FactorySettingsTab,
  FactoryOpportunitiesTab,
  FactoryFunctionsTab,
  FactoryRejectDialog,
  FactoryOpportunitySlideOver,
} from '@/components/factory';
import clsx from 'clsx';

// Tab type definition
type FactoryTab = 'status' | 'opportunities' | 'reviews' | 'functions' | 'settings';

// Form type for settings
type FactorySettingsForm = Partial<FactoryConfig>;

/**
 * AdminFactoryPage - Main factory management page
 */
export function AdminFactoryPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const toast = useToastHelpers();
  const [activeTab, setActiveTab] = useState<FactoryTab>('status');
  const [selectedReview, setSelectedReview] = useState<PendingReview | null>(null);
  const [rejectReason, setRejectReason] = useState('');
  const [showRejectDialog, setShowRejectDialog] = useState(false);

  // Opportunity slide-over state
  const [selectedOpportunity, setSelectedOpportunity] = useState<Opportunity | null>(null);
  const [showOpportunitySlideOver, setShowOpportunitySlideOver] = useState(false);

  // Opportunities pagination state
  const [opportunityOffset, setOpportunityOffset] = useState(0);
  const opportunityLimit = 20;

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        if (showOpportunitySlideOver) setShowOpportunitySlideOver(false);
        else if (showRejectDialog) setShowRejectDialog(false);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [showOpportunitySlideOver, showRejectDialog]);

  // Fetch factory status
  const {
    data: factoryStatus,
    isLoading: statusLoading,
    error: statusError,
    refetch: refetchStatus,
  } = useQuery({
    queryKey: ['factory-status'],
    queryFn: async (): Promise<FactoryStatus | null> => {
      const data = await factoryApi.getStatus();
      return data ?? null;
    },
    refetchInterval: 30000,
  });

  // Fetch pending reviews
  const {
    data: pendingReviews,
    isLoading: reviewsLoading,
    refetch: refetchReviews,
  } = useQuery({
    queryKey: ['factory-pending-reviews'],
    queryFn: async () => {
      const data = await factoryApi.listPendingReviews({ limit: 50 });
      return data ?? { reviews: [], total: 0, limit: 50, offset: 0 };
    },
    refetchInterval: 15000,
  });

  // Fetch factory config for Settings tab
  const {
    data: factoryConfig,
    isLoading: configLoading,
    error: configError,
  } = useQuery({
    queryKey: ['factory-config'],
    queryFn: async (): Promise<FactoryConfig | null> => {
      const data = await factoryApi.getConfig();
      return data ?? null;
    },
    enabled: activeTab === 'settings',
  });

  // Fetch all opportunities for Opportunities tab
  const {
    data: opportunitiesData,
    isLoading: opportunitiesLoading,
    refetch: refetchOpportunities,
  } = useQuery({
    queryKey: ['factory-opportunities', opportunityOffset],
    queryFn: async (): Promise<OpportunityListResponse> => {
      const data = await factoryApi.listOpportunities({
        limit: opportunityLimit,
        offset: opportunityOffset,
      });
      return data ?? { opportunities: [], total: 0, limit: opportunityLimit, offset: opportunityOffset };
    },
    enabled: activeTab === 'opportunities',
  });

  // Settings form state
  const [settingsForm, setSettingsForm] = useState<FactorySettingsForm>({});
  const [settingsFormDirty, setSettingsFormDirty] = useState(false);

  useEffect(() => {
    if (factoryConfig && activeTab === 'settings') {
      setSettingsForm({
        agent_id: factoryConfig.agent_id,
        discovery_batch_size: factoryConfig.discovery_batch_size,
        minimum_quality_score: factoryConfig.minimum_quality_score,
        minimum_test_score: factoryConfig.minimum_test_score,
        require_all_tests_pass: factoryConfig.require_all_tests_pass,
        auto_publish: factoryConfig.auto_publish,
        max_opportunities_per_run: factoryConfig.max_opportunities_per_run,
        retry_attempts: factoryConfig.retry_attempts,
        retry_backoff_ms: factoryConfig.retry_backoff_ms,
        schedule_enabled: factoryConfig.schedule_enabled ?? false,
        schedule_cron: factoryConfig.schedule_cron ?? '0 0 * * *',
        schedule_timezone: factoryConfig.schedule_timezone ?? 'UTC',
        notification_webhook_url: factoryConfig.notification_webhook_url ?? '',
        rate_limit_per_hour: factoryConfig.rate_limit_per_hour ?? 10,
        max_concurrent_runs: factoryConfig.max_concurrent_runs ?? 1,
        dry_run_mode: factoryConfig.dry_run_mode ?? false,
        discovery_sources: factoryConfig.discovery_sources ?? [],
        feature_flags: factoryConfig.feature_flags ?? {},
        approval_required_above_quality: factoryConfig.approval_required_above_quality ?? 0,
        approval_required_above_test: factoryConfig.approval_required_above_test ?? 0,
        log_level: factoryConfig.log_level ?? 'info',
        notify_on_failure: factoryConfig.notify_on_failure ?? true,
        notify_on_review_required: factoryConfig.notify_on_review_required ?? true,
        discovery_cooldown_minutes: factoryConfig.discovery_cooldown_minutes ?? 60,
        max_versions_per_function: factoryConfig.max_versions_per_function ?? 5,
      });
      setSettingsFormDirty(false);
    }
  }, [factoryConfig, activeTab]);

  // Update config mutation
  const updateConfigMutation = useMutation({
    mutationFn: (payload: FactorySettingsForm) => factoryApi.updateConfig(payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['factory-config'] });
      queryClient.invalidateQueries({ queryKey: ['factory-status'] });
      setSettingsFormDirty(false);
      toast.success('Configuration saved successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to save configuration: ${error.message}`);
    },
  });

  // Pipeline run mutation
  const runPipelineMutation = useMutation({
    mutationFn: factoryApi.triggerPipelineRun,
    onSuccess: (data) => {
      toast.success(`Pipeline run initiated: ${data.run?.id || 'Started successfully'}`);
      refetchStatus();
    },
    onError: (error: Error) => {
      toast.error(`Failed to start pipeline: ${error.message}`);
    },
  });

  // Approve opportunity mutation
  const approveMutation = useMutation({
    mutationFn: (id: string) => factoryApi.approveOpportunity(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ['factory-pending-reviews'] });
      queryClient.invalidateQueries({ queryKey: ['factory-status'] });
      queryClient.invalidateQueries({ queryKey: ['factory-opportunities'] });
      setSelectedReview(null);
      setSelectedOpportunity(null);
      toast.success('Opportunity approved!', {
        description: (
          <a
            href="/functions"
            className="font-medium underline hover:no-underline"
            onClick={(e) => {
              e.preventDefault();
              window.location.href = '/functions';
            }}
          >
            View in Functions &amp; Registry →
          </a>
        ),
        duration: 8000,
      });
    },
    onError: (error: Error) => {
      toast.error(`Failed to approve: ${error.message}`);
    },
  });

  // Reject opportunity mutation
  const rejectMutation = useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) =>
      factoryApi.rejectOpportunity(id, reason),
    onSuccess: () => {
      toast.success('Opportunity rejected');
      queryClient.invalidateQueries({ queryKey: ['factory-pending-reviews'] });
      queryClient.invalidateQueries({ queryKey: ['factory-status'] });
      queryClient.invalidateQueries({ queryKey: ['factory-opportunities'] });
      setShowRejectDialog(false);
      setSelectedReview(null);
      setRejectReason('');
    },
    onError: (error: Error) => {
      toast.error(`Failed to reject: ${error.message}`);
    },
  });

  // Handlers
  const handleRunPipeline = () => {
    runPipelineMutation.mutate();
  };

  const handleApproveReview = (review: PendingReview) => {
    approveMutation.mutate(review.id);
  };

  const handleRejectReview = (review: PendingReview) => {
    setSelectedReview(review);
    setShowRejectDialog(true);
  };

  const handleBulkApprove = (ids: string[]) => {
    toast.success(`Approving ${ids.length} opportunities...`);
    ids.forEach((id, index) => {
      setTimeout(() => {
        approveMutation.mutate(id);
      }, index * 500);
    });
  };

  const handleBulkReject = (reviews: PendingReview[]) => {
    if (reviews.length === 1) {
      setSelectedReview(reviews[0]);
      setShowRejectDialog(true);
    } else {
      toast.error('Please reject items one at a time. Bulk reject is not supported.');
    }
  };

  const handleConfirmReject = () => {
    if (!selectedReview || !rejectReason.trim()) return;
    rejectMutation.mutate({ id: selectedReview.id, reason: rejectReason });
  };

  const handleUndoReject = (id: string) => {
    toast.info('Rejection undone');
  };

  const handleApproveOpportunity = (id: string) => {
    approveMutation.mutate(id);
    setShowOpportunitySlideOver(false);
  };

  const handleRejectOpportunity = (id: string) => {
    const opp = opportunitiesData?.opportunities.find((o) => o.id === id);
    if (opp) {
      setSelectedReview({
        id: opp.id,
        source: opp.source,
        status: opp.status,
        review_status: opp.review_status,
        quality_score: opp.quality_score,
        test_score: opp.test_score,
        title: opp.title,
        description: opp.description,
        created_at: opp.created_at,
        review_requested_at: opp.review_requested_at,
      } as PendingReview);
    }
    setShowOpportunitySlideOver(false);
    setShowRejectDialog(true);
  };

  const handleViewOpportunity = (opportunity: Opportunity) => {
    setSelectedOpportunity(opportunity);
    setShowOpportunitySlideOver(true);
  };

  const handleViewFactoryFunction = (fn: PublishedFunction) => {
    navigate(`/functions/${fn.id}`);
  };

  const handleOpportunityPageChange = (newOffset: number) => {
    setOpportunityOffset(newOffset);
  };

  const updateSetting = <K extends keyof FactorySettingsForm>(
    key: K,
    value: FactorySettingsForm[K]
  ) => {
    setSettingsForm((prev) => ({ ...prev, [key]: value }));
    setSettingsFormDirty(true);
  };

  const handleSaveSettings = () => {
    updateConfigMutation.mutate(settingsForm);
  };

  // Tab counts for badges
  const pendingReviewCount = pendingReviews?.total ?? 0;

  if (statusLoading) {
    return <LoadingScreen />;
  }

  return (
    <div className="p-6 max-w-7xl mx-auto">
      {/* Page Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">AI Function Factory</h1>
          <p className="text-gray-600 dark:text-gray-400 mt-1">
            Monitor factory status and manage opportunity reviews
          </p>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={() => refetchStatus()}
            disabled={statusLoading}
            className="flex items-center gap-2 px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors disabled:opacity-50 text-gray-700 dark:text-gray-300"
          >
            <RefreshCw className={clsx('h-4 w-4', statusLoading && 'animate-spin')} />
            Refresh
          </button>
          <button
            onClick={handleRunPipeline}
            disabled={runPipelineMutation.isPending}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 dark:bg-blue-600 text-white rounded-lg hover:bg-blue-700 dark:hover:bg-blue-700 transition-colors disabled:opacity-50"
          >
            <Play className={clsx('h-4 w-4', runPipelineMutation.isPending && 'animate-spin')} />
            {runPipelineMutation.isPending ? 'Starting...' : 'Run Pipeline'}
          </button>
        </div>
      </div>

      {/* Error Alert */}
      {statusError && (
        <div className="mb-6 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
          <div className="flex items-center gap-2 text-red-800 dark:text-red-300">
            <AlertTriangle className="h-5 w-5" />
            <span>Failed to load factory status: {(statusError as Error).message}</span>
          </div>
        </div>
      )}

      {/* Tabs */}
      <div className="border-b border-gray-200 dark:border-gray-700 mb-6">
        <nav className="-mb-px flex space-x-8">
          <button
            onClick={() => setActiveTab('status')}
            className={clsx(
              'py-4 px-1 border-b-2 font-medium text-sm transition-colors',
              activeTab === 'status'
                ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600'
            )}
          >
            <Activity className="inline h-4 w-4 mr-1.5" />
            Status
          </button>
          <button
            onClick={() => setActiveTab('opportunities')}
            className={clsx(
              'py-4 px-1 border-b-2 font-medium text-sm transition-colors flex items-center gap-2',
              activeTab === 'opportunities'
                ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600'
            )}
          >
            <Search className="h-4 w-4" />
            Opportunities
            {factoryStatus?.totals.opportunities ? (
              <span className="ml-1.5 px-2 py-0.5 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 text-xs rounded-full">
                {factoryStatus.totals.opportunities}
              </span>
            ) : null}
          </button>
          <button
            onClick={() => setActiveTab('reviews')}
            className={clsx(
              'py-4 px-1 border-b-2 font-medium text-sm transition-colors flex items-center gap-2',
              activeTab === 'reviews'
                ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600'
            )}
          >
            <Eye className="h-4 w-4" />
            Reviews
            {pendingReviewCount > 0 ? (
              <span className="ml-1.5 px-2 py-0.5 bg-red-100 dark:bg-red-900/50 text-red-600 dark:text-red-400 text-xs rounded-full">
                {pendingReviewCount}
              </span>
            ) : null}
          </button>
          <button
            onClick={() => setActiveTab('functions')}
            className={clsx(
              'py-4 px-1 border-b-2 font-medium text-sm transition-colors flex items-center gap-2',
              activeTab === 'functions'
                ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600'
            )}
          >
            <Package className="h-4 w-4" />
            Functions
          </button>
          <button
            onClick={() => setActiveTab('settings')}
            className={clsx(
              'py-4 px-1 border-b-2 font-medium text-sm transition-colors flex items-center gap-2',
              activeTab === 'settings'
                ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600'
            )}
          >
            <Settings className="h-4 w-4" />
            Settings
          </button>
        </nav>
      </div>

      {/* Tab Content */}
      {activeTab === 'status' && (
        <FactoryStatusTab
          factoryStatus={factoryStatus ?? null}
          isLoading={statusLoading}
          onRefresh={refetchStatus}
        />
      )}

      {activeTab === 'opportunities' && (
        <FactoryOpportunitiesTab
          opportunities={opportunitiesData?.opportunities ?? []}
          total={opportunitiesData?.total ?? 0}
          limit={opportunityLimit}
          offset={opportunityOffset}
          isLoading={opportunitiesLoading}
          onRefresh={refetchOpportunities}
          onViewDetails={handleViewOpportunity}
          onPageChange={handleOpportunityPageChange}
        />
      )}

      {activeTab === 'reviews' && (
        <FactoryReviewsTab
          reviews={pendingReviews?.reviews ?? []}
          total={pendingReviews?.total ?? 0}
          isLoading={reviewsLoading}
          onRefresh={refetchReviews}
          onApprove={handleApproveReview}
          onReject={handleRejectReview}
          onBulkApprove={handleBulkApprove}
          onBulkReject={handleBulkReject}
          isApproving={approveMutation.isPending}
        />
      )}

      {activeTab === 'functions' && (
        <FactoryFunctionsTab onViewFunction={handleViewFactoryFunction} />
      )}

      {activeTab === 'settings' && (
        <>
          {configLoading ? (
            <div className="text-center py-12 text-gray-500 dark:text-gray-400">Loading configuration...</div>
          ) : configError ? (
            <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
              <p className="text-red-800 dark:text-red-300">Failed to load configuration: {(configError as Error).message}</p>
            </div>
          ) : (
            <FactorySettingsTab
              settingsForm={settingsForm}
              isDirty={settingsFormDirty}
              isSaving={updateConfigMutation.isPending}
              onSettingChange={updateSetting}
              onSave={handleSaveSettings}
            />
          )}
        </>
      )}

      {/* Reject Dialog */}
      <FactoryRejectDialog
        open={showRejectDialog}
        reason={rejectReason}
        isPending={rejectMutation.isPending}
        onOpenChange={setShowRejectDialog}
        onReasonChange={setRejectReason}
        onConfirm={handleConfirmReject}
        onUndo={handleUndoReject}
      />

      {/* Opportunity Slide-Over */}
      <FactoryOpportunitySlideOver
        opportunity={selectedOpportunity}
        open={showOpportunitySlideOver}
        onOpenChange={setShowOpportunitySlideOver}
        onApprove={handleApproveOpportunity}
        onReject={handleRejectOpportunity}
        isApproving={approveMutation.isPending}
        isRejecting={rejectMutation.isPending}
      />
    </div>
  );
}

export default AdminFactoryPage;
