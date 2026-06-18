/**
 * Dashboard Components Index
 *
 * Unified exports for all dashboard components including:
 * - Studio Layout System (studio/)
 * - Registry & Marketplace (registry/)
 * - Graph Runtime (graph/)
 * - Observability (observability/)
 * - AI Command System (ai/)
 */

// Studio Layout System
export { StudioShell, useStudioContext } from './studio'
export type { StudioShellProps } from './studio'
export { AdaptiveWorkspace, useWorkspaceMode } from './studio'
export type { AdaptiveWorkspaceProps } from './studio'
export { WorkspaceViewport } from './studio'
export type { WorkspaceViewportProps } from './studio'
export { WorkspaceScene } from './studio'
export type { WorkspaceSceneProps } from './studio'
export { DockLayout } from './studio'
export type { DockLayoutProps, DockPosition, DockPanel } from './studio'
export { FloatingPanelSystem } from './studio'
export type { FloatingPanelSystemProps, FloatingPanel } from './studio'
export { ResizablePanel } from './studio'
export type { ResizablePanelProps } from './studio'
export { SnapGridLayout } from './studio'
export type { SnapGridLayoutProps, GridItem } from './studio'
export { SplitViewManager } from './studio'
export type { SplitViewManagerProps, SplitView } from './studio'
export { SmartSidebar } from './studio'
export type { SmartSidebarProps, SidebarSection } from './studio'
export { ContextSurface } from './studio'
export type { ContextSurfaceProps } from './studio'
export { WindowManager } from './studio'
export type { WindowManagerProps, WindowState } from './studio'
export { WorkspaceTabs } from './studio'
export type { WorkspaceTabsProps, WorkspaceTab } from './studio'
export { MultiMonitorWorkspace, useScreenDetection } from './studio'
export type { MultiMonitorWorkspaceProps, Monitor } from './studio'
export { ImmersiveMode } from './studio'
export type { ImmersiveModeProps } from './studio'
export { CommandOverlay } from './studio'
export type { CommandOverlayProps, CommandItem } from './studio'
export { StudioTopbar } from './studio'
export type { StudioTopbarProps } from './studio'
export { ActivityRail } from './studio'
export type { ActivityRailProps, ActivityItem } from './studio'
export { QuickActionRail } from './studio'
export type { QuickActionRailProps, QuickAction } from './studio'
export { BottomTelemetryBar } from './studio'
export type { BottomTelemetryBarProps } from './studio'
export { useStudioStore } from '@/stores/studioStore'

// Registry & Marketplace Components
export { useRegistryStore, selectFilteredFunctions, selectSortedFunctions } from '@/stores/registryStore'

// Graph Runtime Components
export { useGraphRuntimeStore } from '@/stores/graphRuntimeStore'

// Observability Components
export { ObservabilityIntegration } from './observability'
export { default as AgentRunTimeline } from './observability/AgentRunTimeline'
export { default as DecisionGraphViewer } from './observability/DecisionGraphViewer'
export { default as EventDetailPanel } from './observability/EventDetailPanel'
export { default as CostBreakdownPanel } from './observability/CostBreakdownPanel'
export { default as ReplayControls } from './observability/ReplayControls'
export { default as RealtimeStream } from './observability/RealtimeStream'
export { default as AtlasStatusBadge } from './observability/AtlasStatusBadge'
export { default as SpanNavigator } from './observability/SpanNavigator'
export { default as AtlasConfigPanel } from './observability/AtlasConfigPanel'
export { default as SpanDetailPanel } from './observability/SpanDetailPanel'
export { default as GraphNodeDetail } from './observability/GraphNodeDetail'
export { default as PaginationControls } from './observability/PaginationControls'
export { default as DataExport } from './observability/DataExport'
export { default as AutoRefreshToggle } from './observability/AutoRefreshToggle'
export { default as ReconnectButton } from './observability/ReconnectButton'
export { default as ConfirmDialog } from './observability/ConfirmDialog'
export { default as RunMetadataPanel } from './observability/RunMetadataPanel'
export { RunsListSkeleton, EventsListSkeleton, SpanTreeSkeleton, GraphSkeleton, StatsSkeleton } from './observability/Skeletons'
export type {
  TokenUsage,
  InferenceSpan,
  ErrorNode,
  TraceSpan,
  ReplayRequest,
  LogEntry,
  Anomaly,
  Incident,
  ResourceForecastPoint,
} from '@/stores/observabilityStore'
export { useObservabilityStore, selectCriticalMetrics, selectRecentLogs, selectActiveIncidents } from '@/stores/observabilityStore'

// AI Command System
export { useAICommandStore } from '@/stores/aiCommandStore'

// Visualization Components
export { useVisualizationStore } from '@/stores/visualizationStore'

// Code Intelligence Components
export { useCodeIntelligenceStore } from '@/stores/codeIntelligenceStore'

// DevOps Components
export { useDevOpsStore } from '@/stores/devopsStore'

// Security Components
export { useSecurityStore } from '@/stores/securityStore'

// Collaboration Components
export { useCollaborationStore, selectActivePresences, selectSpeakingParticipants, selectUnresolvedConflicts } from '@/stores/collaborationStore'

// Memory Components
export { useMemoryStore } from '@/stores/memoryStore'

// Simulation Components
export { useSimulationStore } from '@/stores/simulationStore'

// Robotics Components
export { RoboticsIntegration } from './robotics/RoboticsIntegration'
export type {
  Robot,
  RobotStatus,
  RobotType,
  Fleet,
  SensorReading,
  Command,
  MapWaypoint,
  Obstacle,
  FlightPath,
  VisionFrame,
  MeshNode as RoboticsMeshNode,
  Actuator,
  EdgeDevice,
  RoboticWorkflow,
  RobotFleetDashboardProps,
  SensorTelemetryPanelProps,
  RobotCommandCenterProps,
  PhysicalEnvironmentMapProps,
  DroneFlightOverlayProps,
  RobotVisionStreamProps,
  DeviceMeshViewerProps,
  ActuatorControlPanelProps,
  EdgeDeviceMonitorProps,
  RoboticWorkflowDesignerProps,
} from '@functionfly/ui-robotics'
export {
  useRoboticsStore,
  selectSelectedRobot,
  selectOnlineRobots,
  selectAlertCount as selectRoboticsAlertCount,
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
} from '@/stores/roboticsStore'

// Marketplace Economy Components
export { useMarketplaceEconomyStore } from '@/stores/marketplaceEconomyStore'

// Adaptive UX Components
export { useAdaptiveUXStore } from '@/stores/adaptiveUXStore'

// Universal Runtime Components
export { useUniversalRuntimeStore } from '@/stores/universalRuntimeStore'

// Data Visualization Components
export { useDataVisualizationStore } from '@/stores/dataVisualizationStore'

// Futuristic Signature Components
export { useFuturisticStore } from '@/stores/futuristicStore'
