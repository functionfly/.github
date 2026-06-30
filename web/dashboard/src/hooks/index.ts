// Auth & User Hooks
export {
  dnaKeys,
  useDNAProfile,
  useDNAMutations,
  useDNAMutation,
  useDNAInsights,
  useEnterpriseDNAInsights,
  useAcceptDNAVariant,
  useRejectDNAVariant,
  useTriggerDNAAnalysis,
  useToggleDNAEvolution,
} from './useFunctionDNA';

export {
  useAddSkill,
  useChangePassword,
  useChangeUsername,
  useDeleteAccount,
  useLoginHistory,
  useMe,
  usePublicProfile,
  useRemoveSkill,
  useReportProfile,
  useRevokeOtherSessions,
  useRevokeSession,
  userKeys,
  useUpdateMe,
  useUserAchievements,
  useUserActivity,
  useUserAnalytics,
  useUserContributions,
  useUsernameChangeEligibility,
  useUsernameChangeHistory,
  useUserSessions,
  useUserSkills,
} from './useUsers';

export {
  authKeys,
  useAdminAPI,
  useAuth,
  useAuthenticatedRequest,
  useCheckUsernameAvailability,
  useDisableMFA,
  useEnableMFA,
  useMagicLink,
  useMFAStatus,
  useOAuthProviders,
  useOAuthUrl,
  usePasswordResetConfirm,
  usePasswordResetRequest,
  useResendVerification,
  useSetupMFA,
  useSignIn,
  useSignOut,
  useSignUp,
  useVerifyMagicLink,
  useVerifyMFASetup,
} from './useAuth';

export { useLoginForm, useSignupForm } from './useAuthForms';
export { useProfileUpdates } from './useProfile';
export { useUsernameValidation } from './useUsernameValidation';
export { useTenantUserStatus } from './useUserStatus';

// Team & Organization Hooks
export {
  teamKeys,
  useAcceptInvite,
  useCancelInvite,
  useCreateTeam,
  useDeleteTeam,
  useInviteMember,
  useRemoveMember,
  useResendInvite,
  useTeam,
  useTeamInvites,
  useTeamMembers,
  useTeams,
  useUpdateMember,
  useUpdateTeam,
} from './useTeams';

// Billing Hooks
export {
  billingKeys,
  useBundle,
  useBundles,
  useCancelSubscription,
  useConvertToPaid,
  useCreateBillingPortal,
  useCreateCheckout,
  useCreateStateFabricAddOnCheckout,
  useDeferredBillingStatus,
  useFounderModeStatus,
  useInvoices,
  usePlatformFees,
  useStateFabricAddOns,
  useStateFabricEntitlements,
  useSubscription,
  useTopUpWallet,
  useUsage,
  useWallet,
} from './useBilling';

// Payout Hooks
export {
  payoutKeys,
  useConnectAccountStatus,
  usePayoutBalance,
  usePayoutLedger,
  usePayoutRequests,
  useRefreshConnectAccount,
  useRequestPayout,
  useStartConnectOnboarding,
} from './usePayouts';

// App & Backend Hooks
export {
  appKeys,
  useApp,
  useApps,
  useAppStatus,
  useBackends,
  useCreateApp,
  useCreateBackend,
  useDeployBackendOptions,
} from './useApps';

// Function Hooks
export {
  functionKeys,
  useCreateFunction,
  useDeleteFunction,
  useDeployFunction,
  useDeploymentLogs,
  useFunction,
  useFunctionDeployments,
  useFunctionLogs,
  useFunctionMetrics,
  useFunctions,
  useFunctionTrustScore,
  useTestFunction,
  useUpdateFunction,
} from './useFunctions';

// Registry Hooks
export {
  registryKeys,
  useExecuteRegistryFunction,
  useFunctionSettings,
  usePublishRegistryFunction,
  useRegistryFunction,
  useRegistryFunctions,
  useRegistryFunctionStats,
  useRegistryFunctionVersions,
  useRegistryReviews,
  useRegistrySearch,
  useReplay,
  useSubmitRegistryRating,
  useSubmitRegistryReview,
  useTestRegistryFunction,
  useUpdateFunctionSettings,
} from './useRegistry';

export {
  catalogKeys,
  useFunctionCatalog,
  useFunctionDetail,
  useFunctionSearch,
  useRecentFunctions,
} from './useFunctionCatalog';

// Deployment Hooks
export {
  deploymentKeys,
  useDeploy,
  useDeployment,
  useDeployments,
  useRollbackDeployment,
} from './useDeployments';

// Agent Hooks
export {
  agentKeys,
  useAgent,
  useAgentPolicy,
  useAgentQuota,
  useAgents,
  useAgentUsage,
  useDeleteAgent,
  useRegisterAgent,
  useUpdateAgent,
  useUpdateAgentPolicy,
} from './useAgent';

export {
  agentMemoryKeys,
  useAgentMemories,
  useAgentMemory,
  useCreateAgentMemory,
  useDeleteAgentMemory,
  useRebuildIndex,
  useSearchAgentMemories,
  useUpdateAgentMemory,
} from './useAgentMemory';

// Agent Search Hooks
export {
  searchKeys,
  useSearchTools,
  useSearchStats,
  useExecuteSearchTool,
} from './useAgentSearch';

// Agent Swarm Hooks
export {
  swarmKeys,
  useAgentChildren,
  useAgentParent,
  useSwarmHealth,
  useAgentInbox,
  useSwarmStats,
  useSpawnChildAgent,
  useSendAgentMessage,
} from './useAgentSwarm';

// Security Alert Hooks
export {
  securityAlertKeys,
  useSecurityAlerts,
  useAcknowledgeAlert,
  useResolveAlert,
  useKillSwitch,
} from './useSecurityAlerts';

// State & Fabric Hooks
export {
  stateKeys,
  useCreateSnapshot,
  useCreateState,
  useDeleteState,
  useDeleteStateValue,
  useGrantPermission,
  usePatchStateValue,
  useRestoreSnapshot,
  useSetStateValue,
  useState,
  useStateHistory,
  useStatePermissions,
  useStates,
  useStateSnapshots,
  useStateValue,
  useTimeTravel,
  useUpdateState,
} from './useState';

export {
  stateFabricKeys,
  useStateFabric,
  useStateFabricEventLogs,
  useStateFabricMetrics,
  useStateFabricPipelines,
  useStateFabricReplays,
  useStateFabrics,
  useStateFabricSnapshots,
  useStateFabricStores,
  useStateFabricTriggers,
} from './useStateFabric';

// Provider Hooks
export {
  providerKeys,
  useConnectedProviders,
  useConnectProvider,
  useDisconnectProvider,
  useTestProviderConnection,
} from './useProviders';

// API Key Hooks
export {
  apiKeyKeys,
  useAddAPIKeyEnvironment,
  useAddAPIKeyPermission,
  useAPIKey,
  useAPIKeyEnvironments,
  useAPIKeyPermissions,
  useAPIKeys,
  useCreateAPIKey,
  useDeleteAPIKey,
  useRemoveAPIKeyEnvironment,
  useRemoveAPIKeyPermission,
  useRotateAPIKey,
  useUpdateAPIKey,
} from './useApiKeys';

// Security Hooks
export {
  securityKeys,
  useComplianceFrameworks,
  useIncidentResponse,
  useSecurityContacts,
  useSecurityFAQ,
  useSecurityIncidents,
  useSecurityMeasures,
  useSecurityMetrics,
  useSecurityResources,
  useServiceStatus,
  useSSLCertificates,
} from './useSecurity';

// Content Hooks
export {
  contentKeys,
  useAdminBlogPosts,
  useAdminChangelogEntries,
  useBlogCategories,
  useBlogPost,
  useBlogPosts,
  useChangelogEntries,
  useCreateBlogPost,
  useCreateChangelogEntry,
  useDeleteBlogPost,
  useDeleteChangelogEntry,
  useGenerateBlogContent,
  useGenerateChangelogContent,
  useSyncGitHubReleases,
  useSyncSanityPosts,
  useUpdateBlogPost,
  useUpdateChangelogEntry,
} from './useContent';

// Decision Hooks
export {
  decisionKeys,
  useApproveDecision,
  useCreateDecision,
  useDecision,
  useDecisions,
  useDeleteDecision,
  useSearchDecisions,
  useUpdateDecision,
} from './useDecisions';

// Conversation Hooks
export {
  conversationKeys,
  useBounties,
  useClaimBounty,
  useConversation,
  useConversations,
  useCreateBounty,
  useCreateConversation,
  useCreateMessage,
  useMarkConversationRead,
  useMessages,
  useRealtimeMessages,
  useResolveConversation,
} from './useConversations';

export { useConversationWebSocket } from './useConversationWebSocket';

// Dashboard Hooks
export {
  dashboardKeys,
  useDashboardActivity,
  useDashboardExecutionRate,
  useDashboardHealthStatus,
  useDashboardMemory,
  useDashboardMetrics,
  useDashboardUsage,
} from './useDashboard';

// Analytics Hooks
export { analyticsKeys, useAnalyticsSettings, useUpdateAnalyticsSettings } from './useAnalytics';

// Admin Hooks
export {
  adminKeys,
  useActivateTenant,
  useAdminAnalytics,
  useAdminInvoices,
  useAuditEvent,
  useAuditEvents,
  useCoupons,
  useCreateCoupon,
  useCreatePricingTier,
  useDeletePricingTier,
  usePricingTiers,
  useSuspendTenant,
  useSystemHealth,
  useTenant,
  useTenants,
  useUpdatePricingTier,
  useUpdateTenant,
} from './useAdmin';

// Notification Hooks
export { useUserNotifications } from './useNotifications';

export { useNotificationUnreadPolling } from './useNotificationUnreadPolling';

// Vault Hooks
export {
  useCreateSecret,
  useDecryptSecret,
  useDeleteSecret,
  useGenerateToken,
  useRevokeToken,
  useSecretAuditLog,
  useSecretTokens,
  useUpdateSecret,
  useVaultAuditLog,
  useVaultSecret,
  useVaultSecrets,
  vaultKeys,
} from './useVault';

// UI/UX Hooks
export { usePageTitle, formatPageTitle } from './usePageTitle.tsx';
export { useCookieConsent } from './useCookieConsent';
export { useInfiniteScroll } from './useInfiniteScroll';
export { useKeyboardNavigation } from './useKeyboardNavigation';
export { useNavigationStatus, useStatusBadge } from './useNavigationStatus';
export { useUserPresence } from './usePresence';
export { useSwipeGesture } from './useSwipeGesture';
export { useDirection } from './useDirection';
export { useSyncUserLanguage } from './useSyncUserLanguage';
export { useWebVitals } from './useWebVitals';
export { useBreadcrumbs } from './useBreadcrumbs';

// Form Hooks
export { useFormWithDraft } from './useFormWithDraft';
export { useFormWithValidation } from './useFormWithValidation';

// Realtime Hooks
export { useNotificationRealtime } from './useNotificationRealtime';
export { useRealtime } from './useRealtime';
export { useCustomStatus, CUSTOM_STATUS_OPTIONS } from './useCustomStatus';
export type { CustomStatus, CustomStatusValue } from './useCustomStatus';
export { useRealtimeSubscription } from './useRealtimeSubscription';

// Status & Activity Hooks
export { useActivityFeed } from './useActivityFeed';
export {
  useFollowFunction,
  useFollowUser,
  useFunctionFollowers,
  useFunctionFollowStatus,
  useMyFollowedFunctions,
  useMyFollowStats,
  useUnfollowFunction,
  useUnfollowUser,
  useUserFollowers,
  useUserFollowing,
  useUserFollowStatus,
} from './useFollow';
export { useNewsletter } from './useNewsletter';
export { usePlan } from './usePlan';
export {
  useProvisioningStatus,
  useProvisionBundle,
  useRetryProvisioning,
  useIsProvisioned,
  useProvisionedBundle,
} from './useProvisioning';
export { useSignupConfig } from './useSignupConfig';
export { useStatus } from './useStatus';
export { useStatusCheck, useStatusHealthCheck, useStatusWebSocket } from './useStatusWebSocket';

// Database Hooks
export {
  useDatabaseAlerts,
  useDatabaseChanges,
  useDatabaseHealth,
  useDatabaseMetrics,
} from './useDatabase';

// Factory Hooks
export {
  factoryKeys,
  useApproveOpportunity,
  useFactoryConfig,
  useFactoryFunctions,
  useFactoryOpportunities,
  useFactoryStatus,
  usePendingReviews,
  useRejectOpportunity,
  useTriggerPipelineRun,
  useUpdateFactoryConfig,
} from './useFactory';

// Utility Hooks
export {
  useApproveExtraction,
  useCreateMemory,
  useDeleteMemory,
  useMemoryExtractions,
  useRejectExtraction,
  useSearchMemories,
  useTeamMemories,
  useUpdateMemory,
  useValidateMemory,
} from './use-team-memory';
export { useAccessControl } from './useAccessControl';
export { useContextualActions } from './useContextualActions';
export { useNewFunction } from './useNewFunction';
export { useUndoRedo } from './useUndoRedo';

// Environment Hooks
export {
  activeEnvironmentKeys,
  useActiveEnvironment,
  useEnvironmentQueryKey,
  useEnvironmentSelectorVisibility,
  type Environment,
} from './useActiveEnvironment';

// AI Command System Hooks
export {
  useAICommandSystem,
  useAIConfidence,
  useActiveAgents,
  useToolInvocations,
  useGoalProgress,
} from './useAICommandSystem';

// Graph Runtime Hooks
export {
  useGraphRuntime,
  useNodeSelection,
  useExecutionReplay,
  useRuntimeMetrics,
  useNodeInspector,
  useResourceConsumption,
  useDistributedRuntime,
  useFailurePropagation,
  useLoopVisualization,
  useNodeVersions,
} from './useGraphRuntime';

// Registry Hooks
export {
  useRegistryStore,
  selectFilteredFunctions,
  selectSortedFunctions,
} from '@/stores/registryStore';

// Observability Hooks
export {
  useObservabilityStore,
  selectCriticalMetrics,
  selectRecentLogs,
  selectActiveIncidents,
} from '@/stores/observabilityStore';

// Visualization Hooks
export {
  useVisualizationStore,
  selectActiveNodes,
  selectCriticalNodes,
  selectHealthyRegions,
} from '@/stores/visualizationStore';

// Code Intelligence Hooks
export {
  useCodeIntelligence,
  useCodeEditor,
  useASTExplorer,
  useDependencies,
  useDiffViewer,
  useArchitectureMap,
  useSmartRefactor,
  useCodeGeneration,
  useInlineAI,
  useIntentExplorer,
  useSemanticSearch,
  useCodeLineage,
  useRiskAnalyzer,
  useImportGraph,
  useExecutionEditor,
  useCompletionInspector,
  useRefactorSimulation,
  useArchitectureConstraints,
  useCodeOwnership,
  useCodeIntelligenceUI,
} from './useCodeIntelligence';

// DevOps Hooks
export {
  useDevOpsStore,
  usePipeline,
  useEnvironments,
  useCloudRegions,
  useRuntimeTargets,
  useKubernetes,
  useEdgeLocations,
  useContainers,
  useVaults,
  useScalableResources,
  useTrafficBalancer,
  useRollbackManager,
  useBuildArtifacts,
  useClusterHealth,
  useColdStartAnalyzer,
  useServerlessExecutionMap,
  useDevOpsUI,
} from './useDevOps';

// Security Hooks
export {
  useSecurityStore,
  usePermissionMatrix,
  useThreats,
  useSecurityTimeline,
  useSandboxBoundaries,
  useAPIExposure,
  useCredentialAccess,
  useRuntimeIsolation,
  useCompliance,
  useZeroTrust,
  useMaliciousExecutions,
  useAuditTrail,
  useEncryptionStatus,
  useSuspiciousBehavior,
  useVulnerabilities,
  useSecurityPolicies,
  useSecurityUI,
} from '@/stores/securityStore';

// Collaboration Hooks
export {
  useCollaborationStore,
  selectActivePresences,
  selectSpeakingParticipants,
  selectUnresolvedConflicts,
} from '@/stores/collaborationStore';

export type {
  CollaboratorPresence,
  VoiceSession,
  VoiceParticipant,
  ExecutionBookmark,
  GraphNode,
  GraphEdge,
  GraphOperation,
  Annotation,
  SessionRecording,
  SessionEvent,
  ActivityItem,
  MemoryCard,
  ConflictResolution,
  ConflictMarker,
  ReviewSession,
  ReviewComment,
  PromptSegment,
  PairProgrammingSession,
  DriverNavigator,
  TaskAssignment,
  TaskAssignee,
  CodePosition,
  CodeRange,
} from '@/stores/collaborationStore';

// Robotics Hooks
export {
  useRoboticsStore,
  useRoboticsFleet,
  useRobotTelemetry,
  useRobotCommands,
  useEnvironmentMap,
  useDroneFlight,
  useVisionStream,
  useDeviceMesh,
  useActuatorControl,
  useEdgeMonitor,
  useWorkflowDesigner,
} from '@/stores/roboticsStore';

// Marketplace Economy Hooks
export {
  useMarketplaceEconomyStore,
  useRevenueAnalytics,
  useSubscriptions,
  useUsageBilling,
  useLicenses,
  useCreatorProfile,
  useLeaderboard,
  useRoyalties,
  useAssetPricing,
  useConversionAnalytics,
  useOptimizer,
  useTrendRadar,
} from '@/stores/marketplaceEconomyStore';

// Adaptive UX Hooks
export {
  useAdaptiveUXStore,
  useComplexityLayers,
  useContextToolbar,
  usePredictiveActions,
  useWorkspaceRecommendations,
  useLearningMode,
  useBeginnerView,
  useExpertView,
  useCognitiveLoad,
  useAttentionFocus,
  useWorkflowHints,
} from '@/stores/adaptiveUXStore';

// Universal Runtime Hooks
export {
  useUniversalRuntimeStore,
  useWasmExecution,
  useGPUKernel,
  useServerlessRuntime,
  useBrowserAgent,
  useEdgeRuntime,
  useHybridOrchestrator,
  useCrossCloudTopology,
  useModelRouting,
  useInferenceSelector,
  useCapabilityMatrix,
} from '@/stores/universalRuntimeStore';

// Data Visualization Hooks
export {
  useDataVisualizationStore,
  useStreamingLineChart,
  useRealtimeScatterPlot,
  use3DTopologyChart,
  useExecutionSunburst,
  useDependencyTreemap,
  useCircularFlow,
  useWaterfallChart,
  useCostDistribution,
  useSemanticCluster,
  useAgentInteractionGraph,
} from '@/stores/dataVisualizationStore';

// Futuristic Hooks
export {
  useFuturisticStore,
  useOrbitCommand,
  useQuantumTransition,
  useHolographic,
  useCinematicFocus,
  useAIThoughtWave,
  useTokenStorm,
  useSwarmMind,
  useAmbientTelemetry,
  useDigitalTwin,
} from '@/stores/futuristicStore';

// Enterprise Audit Hooks
export {
  enterpriseAuditKeys,
  useEnterpriseAuditLogs,
  useEnterpriseAuditFilters,
  useExportEnterpriseAudit,
  useDownloadEnterpriseAuditExport,
  type AuditLogParams,
  type AuditExportParams,
  type AuditExportResult,
} from './useEnterpriseAudit';
