// Type declarations for @functionfly/* packages
// These provide types when packages aren't available via node_modules

// ui-core
declare module '@functionfly/ui-core' {
  export const StudioShell: any;
  export const TitleBar: any;
  export const ComplexityToggle: any;
  export const LeftSidebar: any;
  export const RightPanel: any;
  export const Panel: any;
  export type PanelConfig = any;
  export const ResizablePanelGroup: any;
  export const ResizablePanel: any;
  export const ResizableHandle: any;
  export const THEME_CONFIG: any;
  export const Tabs: any;
  export const TabsList: any;
  export const TabsTrigger: any;
  export const TabsContent: any;
  export const GlassCard: any;
  export const Badge: any;
  export const Tooltip: any;
  export const Spinner: any;
  export const Button: any;
  export const Input: any;
  export const Select: any;
  export const SelectContent: any;
  export const SelectItem: any;
  export const SelectTrigger: any;
  export const SelectValue: any;
  export const DropdownMenu: any;
  export const DropdownMenuContent: any;
  export const DropdownMenuItem: any;
  export const DropdownMenuSeparator: any;
  export const DropdownMenuTrigger: any;
  export const Slider: any;
  export const cn: (...args: any[]) => string;
}

// ui-graph
declare module '@functionfly/ui-graph' {
  export const FunctionCanvas: any;
  export const ExecutionNode: any;
  export const ExecutionEdge: any;
  export const ExecutionTimeline: any;
  export const generateSampleGraph: () => { nodes: any[]; edges: any[] };
  export const GraphContext: any;
  export const RuntimeGraph: any;
  export const DynamicPort: any;
  export const TokenFlowRenderer: any;
  export const NodeInspector: any;
  export const GraphMiniMap: any;
  export const ExecutionReplayControls: any;
  export const ExecutionHeatmap: any;
  export const GraphViewport: any;
  export const WorkflowForkManager: any;
  export const LiveNodeTelemetry: any;
  export const GraphStateDiffViewer: any;
  export const DependencyExplorer: any;
  export const OrchestrationLayerView: any;
  export const ConditionalBranchRenderer: any;
  export const LoopVisualizer: any;
  export const FailurePropagationMap: any;
  export const ResourceConsumptionOverlay: any;
  export const NodeVersionHistory: any;
  export const ExecutionTraceExplorer: any;
  export const DistributedRuntimeMap: any;
  export const CanvasViewport: any;
  export type NodeData = any;
  export type EdgeData = any;
  export type ExecutionEvent = any;
  export type ViewMode = any;
  export type RuntimeStatus = any;
  export type RuntimeMetrics = any;
  export type NodeType = any;
  export type NodeInspectorProps = any;
  export type GraphMiniMapProps = any;
  export type ExecutionReplayControlsProps = any;
  export type ExecutionHeatmapProps = any;
  export type WorkflowForkManagerProps = any;
  export type LiveNodeTelemetryProps = any;
  export type GraphStateDiffViewerProps = any;
  export type DependencyExplorerProps = any;
  export type OrchestrationLayerViewProps = any;
  export type ConditionalBranchRendererProps = any;
  export type LoopVisualizerProps = any;
  export type FailurePropagationMapProps = any;
  export type ResourceConsumptionOverlayProps = any;
  export type NodeVersionHistoryProps = any;
  export type ExecutionTraceExplorerProps = any;
  export type DistributedRuntimeMapProps = any;
  export type RuntimeGraphProps = any;
  export const getStatusColor: (status: string) => string;
}

// ui-ai
declare module '@functionfly/ui-ai' {
  export const AICommandPalette: any;
  export const PromptComposer: any;
  export const AgentChatPanel: any;
  export const ReasoningStream: any;
  export const PromptHistory: any;
  export const ConversationThread: any;
  export const ExecutionNarrator: any;
  export const GoalPlanner: any;
  export const IntentTranslator: any;
  export const PromptTemplateLibrary: any;
  export const ToolInvocationFeed: any;
  export const AIConfidenceMeter: any;
  export const MultiAgentConversationView: any;
  export const AgentConversationTimeline: any;
  export type AICommand = any;
  export type AgentChatMessage = any;
}

// ui-agent
declare module '@functionfly/ui-agent' {
  export const AgentDock: any;
  export const AgentCard: any;
  export const AgentLifecyclePanel: any;
  export const AgentMemoryViewer: any;
  export const AgentPermissionEditor: any;
  export const AgentToolchainEditor: any;
  export const AgentBudgetMeter: any;
  export const SwarmCoordinator: any;
  export const AgentSkillGraph: any;
  export const AgentDependencyMap: any;
  export const AutonomousTaskBoard: any;
  export const AgentConsensusViewer: any;
  export const AgentConflictResolver: any;
  export const AgentRuntimeInspector: any;
  export type AgentData = any;
  export type AgentPermission = any;
  export type AgentTool = any;
}

// ui-observability
declare module '@functionfly/ui-observability' {
  export const LiveTelemetryPanel: any;
  export const MetricCard: any;
  export const TokenUsageStream: any;
  export const CostHeatmap: any;
  export const LatencyGraph: any;
  export const ExecutionProfiler: any;
  export const InferenceTraceViewer: any;
  export type TelemetryMetric = any;
  export type TokenUsage = any;
}

// ui-marketplace
declare module '@functionfly/ui-marketplace' {
  export const FunctionMarketplace: any;
  export const FunctionCard: any;
  export const VerifiedBadge: any;
  export const TrustScoreMeter: any;
  export const InstallFunctionModal: any;
  export const RevenueAnalytics: any;
  export const SubscriptionManager: any;
  export const UsageBillingPanel: any;
  export const CreatorProfile: any;
  export const MarketplaceLeaderboard: any;
  export const FunctionRoyaltiesPanel: any;
  export const AssetPricingEditor: any;
  export const MonetizationOptimizer: any;
  export const LicenseManager: any;
  export type FunctionCardData = any;
}

// ui-visualization
declare module '@functionfly/ui-visualization' {
  export const NeuralExecutionMap: any;
  export const ParticleFlow: any;
}

// ui-runtime
declare module '@functionfly/ui-runtime' {
  export const RuntimeTargetSelector: any;
  export const RuntimeCard: any;
  export const WasmExecutionPanel: any;
  export const ServerlessRuntimeViewer: any;
  export const EdgeRuntimeMap: any;
  export const GPUKernelInspector: any;
  export const CrossCloudTopologyMap: any;
  export const ModelRoutingVisualizer: any;
  export const InferenceProviderSelector: any;
  export const RuntimeCapabilityMatrix: any;
  export type RuntimeDescriptor = any;
  export type RuntimeSelection = any;
}

// ui-security
declare module '@functionfly/ui-security' {
  export const PermissionMatrix: any;
  export const ThreatDetectionRadar: any;
  export const SecurityTimeline: any;
  export const SandboxBoundaryViewer: any;
  export const ZeroTrustFlowViewer: any;
}

// ui-collaboration
declare module '@functionfly/ui-collaboration' {
  export const SessionReplayViewer: any;
  export const LivePresence: any;
  export const LivePresenceLayer: any;
  export const CollaboratorCursor: any;
  export const SharedMemoryBoard: any;
  export const TeamActivityFeed: any;
  export const ConflictResolutionPanel: any;
  export const VoiceSessionPanel: any;
  export const AIHumanTaskBoard: any;
  export const AIHumanTaskAssignmentBoard: any;
  export const SharedExecutionView: any;
  export const CollaborativeGraphEditor: any;
  export const RealtimeAnnotationSystem: any;
  export const AsyncReviewTimeline: any;
  export const CollaborativePromptEditor: any;
  export const LivePairProgrammingView: any;
  export type Collaborator = any;
}

// ui-editor
declare module '@functionfly/ui-editor' {
  export const SemanticCodeEditor: any;
  export const RUNTIME_MONACO_LANG: Record<string, string>;
}

// ui-simulation
declare module '@functionfly/ui-simulation' {
  export type SimulationConfig = any;
  export type SimulationResult = any;
  export type CostEstimate = any;
  export const SimulationControlCenter: any;
  export const ExecutionForecastPanel: any;
  export const FailureProbabilityMap: any;
  export const LatencyPredictionGraph: any;
  export const CostSimulationChart: any;
  export const HallucinationRiskAnalyzer: any;
  export const StressTestRunner: any;
  export const ScalingForecastMap: any;
  export const AgentBehaviorPredictor: any;
  export const WorkflowOutcomeSimulator: any;
  export const ResourceCollisionDetector: any;
}

// ui-ghost
declare module '@functionfly/ui-ghost' {
  export type GhostPhase = "planning" | "provisioning" | "building" | "deploying" | "monitoring" | "complete" | "error" | "paused";
  export type GhostBuild = any;
  export type GhostTask = any;
  export type AgentConversationMessage = any;
  export type AgentDecisionPoint = any;
  export const GhostModeOrchestrator: any;
  export const AgentConversationTimeline: any;
  export const MultiAgentConversationView: any;
}

// ui-extensibility
declare module '@functionfly/ui-extensibility' {
  export const ExtensionManager: any;
  export const ExtensionDetailPanel: any;
  export const HookSystemVisualizer: any;
  export const SandboxMonitor: any;
  export const ExtensionSDKDebugger: any;
  export type Extension = any;
}

// ui-futuristic
declare module '@functionfly/ui-futuristic' {
  export const OrbitCommandLayer: any;
  export const QuantumWorkspaceTransition: any;
  export const HolographicPanel: any;
  export const CinematicFocusMode: any;
  export const AIThoughtWave: any;
  export const GlassExecutionCard: any;
  export const TokenStormRenderer: any;
  export const SwarmMindVisualizer: any;
  export const AmbientTelemetryLayer: any;
  export const DigitalTwinViewport: any;
  export const AmbientEffects: any;
  export const OrbitCommand: any;
  export const QuantumTransition: any;
  export const HolographicDisplay: any;
  export const CinematicFocus: any;
  export const AIThoughtWaveVisualizer: any;
  export const TokenStreamDisplay: any;
  export const SwarmAgentMonitor: any;
  export const TelemetryMetricsPanel: any;
  export const DigitalTwinView: any;
}

// ui-data-visualization
declare module '@functionfly/ui-data-visualization' {
  export const StreamingLineChart: any;
  export const RealtimeScatterPlot: any;
  export const ThreeDTopologyChart: any;
  export const ExecutionSunburst: any;
  export const DependencyTreemap: any;
  export const CircularFlow: any;
  export const WaterfallChart: any;
  export const CostDistribution: any;
  export const SemanticCluster: any;
  export const AgentInteractionGraph: any;
}

// ui-adaptive-ux
declare module '@functionfly/ui-adaptive-ux' {
  export const AdaptiveComplexityLayer: any;
  export const ContextAwareToolbar: any;
  export const PredictiveActionBar: any;
  export const SmartWorkspaceRecommendations: any;
  export const LearningModeOverlay: any;
  export const BeginnerSimplificationView: any;
  export const ExpertSystemView: any;
  export const CognitiveLoadBalancer: any;
  export const AttentionFocusOverlay: any;
  export const WorkflowOptimizationHints: any;
  export const AdaptiveLayoutEngine: any;
  export type UserSkillLevel = 'beginner' | 'intermediate' | 'expert';
}

// ui-devops
declare module '@functionfly/ui-devops' {
  export const DeploymentPipeline: any;
  export const EnvironmentManager: any;
  export const CloudRegionSelector: any;
  export const RuntimeTargetSelector: any;
  export const ContainerTopologyView: any;
  export const EdgeDeploymentMap: any;
  export const ContainerLifecyclePanel: any;
  export const SecretVaultManager: any;
  export const InfrastructureDiffViewer: any;
  export const ResourceScaler: any;
  export const TrafficBalancerView: any;
  export const RollbackManager: any;
  export const DeploymentSimulation: any;
  export const BuildArtifactExplorer: any;
  export const ClusterHealthMonitor: any;
  export const ColdStartAnalyzer: any;
  export const ServerlessExecutionMap: any;
  export const PipelineOrchestrator: any;
  export const RegionSelector: any;
  export const RuntimeMetricsPanel: any;
  export const KubernetesDashboard: any;
  export const EdgeDeploymentPanel: any;
  export const ContainerRegistryViewer: any;
  export const SecretsVaultViewer: any;
  export const ResourceAllocationPanel: any;
  export const TrafficRouterViewer: any;
}

// ui-universal-runtime
declare module '@functionfly/ui-universal-runtime' {
  export const UniversalRuntimeOrchestrator: any;
  export const WASMExecutionPanel: any;
  export const GPUKernelInspector: any;
  export const ServerlessRuntimeViewer: any;
  export const BrowserAgentRuntime: any;
  export const EdgeRuntimePanel: any;
  export const HybridRuntimeCoordinator: any;
  export const CrossCloudTopologyMap: any;
  export const ModelRoutingVisualizer: any;
  export const InferenceProviderSelector: any;
}

// ui-robotics
declare module '@functionfly/ui-robotics' {
  export const RobotFleetDashboard: any;
  export const SensorTelemetryPanel: any;
  export const RobotCommandCenter: any;
  export const PhysicalEnvironmentMap: any;
  export const DroneFlightOverlay: any;
  export const RobotVisionStream: any;
  export const DeviceMeshViewer: any;
  export const ActuatorControlPanel: any;
  export const EdgeDeviceMonitor: any;
  export const RoboticWorkflowDesigner: any;
  export type Robot = any;
  export type Fleet = any;
  export type SensorReading = any;
  export type Command = any;
  export type MapWaypoint = any;
  export type Obstacle = any;
  export type FlightPath = any;
  export type VisionFrame = any;
  export type MeshNode = any;
  export type Actuator = any;
  export type EdgeDevice = any;
  export type RoboticWorkflow = any;
}

// ui-memory
declare module '@functionfly/ui-memory' {
  export const MemoryGraph: any;
  export const SemanticMemoryViewer: any;
  export const LongTermContextExplorer: any;
  export const MemoryRecallTimeline: any;
  export const KnowledgeClusterMap: any;
  export const MemoryDecayVisualizer: any;
  export const VectorEmbeddingExplorer: any;
  export const SharedAgentMemoryPanel: any;
  export const MemoryMergeTool: any;
  export const ConversationMemoryTree: any;
  export const MemoryAccessMonitor: any;
}

// ui-marketplace-economy
declare module '@functionfly/ui-marketplace-economy' {
  export const CreatorEconomy: any;
  export const RevenueAnalytics: any;
  export const SubscriptionManager: any;
  export const UsageBillingPanel: any;
  export const LicenseManager: any;
  export const CreatorProfile: any;
  export const MarketplaceLeaderboard: any;
  export const FunctionRoyaltiesPanel: any;
  export const AssetPricingEditor: any;
  export const SalesConversionAnalytics: any;
  export const MonetizationOptimizer: any;
  export const MarketplaceTrendRadar: any;
  export type Subscription = any;
  export type License = any;
  export type RoyaltyRecord = any;
  export type LeaderboardEntry = any;
  export type PricingTier = any;
  export type ConversionFunnelStep = any;
  export type OptimizationSuggestion = any;
  export type TrendItem = any;
}

// ui-code-intelligence
declare module '@functionfly/ui-code-intelligence' {
  export const CodeIntelligenceDashboard: any;
  export const SemanticCodeEditor: any;
  export const ASTExplorer: any;
  export const DependencyHeatmap: any;
  export const MultiFileDiffViewer: any;
  export const ArchitectureMap: any;
  export const SmartRefactorPanel: any;
  export const CodeGenerationPreview: any;
  export const InlineAIAssistant: any;
  export const CodeIntentExplorer: any;
  export const SemanticSearchPanel: any;
  export const CodeLineageViewer: any;
  export const CodeRiskAnalyzer: any;
  export const ImportGraphViewer: any;
  export const ExecutionAwareEditor: any;
  export const AICompletionInspector: any;
  export const RefactorSimulationViewer: any;
  export const ArchitectureConstraintPanel: any;
  export const CodeOwnershipMap: any;
}

// shared
declare module '@functionfly/shared' {
  export const cn: (...args: any[]) => string;
}

// shared/theme
declare module '@functionfly/shared/theme' {
  export const initTheme: any;
  export const setTheme: any;
  export const subscribe: any;
  export type ThemeState = any;
}