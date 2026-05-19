/**
 * Studio Layout System Components
 * 
 * Core layout components for the Studio workspace with docking, panels,
 * and resizable layouts. All components are wired together through
 * the StudioStore for global state management.
 */

// Re-export the main studio components
export { StudioShell, useStudioContext } from './StudioShell'
export type { StudioShellProps } from './StudioShell'

export { AdaptiveWorkspace, useWorkspaceMode } from './AdaptiveWorkspace'
export type { AdaptiveWorkspaceProps } from './AdaptiveWorkspace'

export { WorkspaceViewport } from './WorkspaceViewport'
export type { WorkspaceViewportProps } from './WorkspaceViewport'

export { WorkspaceScene } from './WorkspaceScene'
export type { WorkspaceSceneProps } from './WorkspaceScene'

// Panel System Components
export { DockLayout } from './DockLayout'
export type { DockLayoutProps, DockPosition, DockPanel } from './DockLayout'

export { FloatingPanelSystem } from './FloatingPanelSystem'
export type { FloatingPanelSystemProps, FloatingPanel } from './FloatingPanelSystem'

export { ResizablePanel } from './ResizablePanel'
export type { ResizablePanelProps } from './ResizablePanel'

export { SnapGridLayout } from './SnapGridLayout'
export type { SnapGridLayoutProps, GridItem } from './SnapGridLayout'

export { SplitViewManager } from './SplitViewManager'
export type { SplitViewManagerProps, SplitView } from './SplitViewManager'

// Workspace Components
export { SmartSidebar } from './SmartSidebar'
export type { SmartSidebarProps, SidebarSection } from './SmartSidebar'

export { ContextSurface } from './ContextSurface'
export type { ContextSurfaceProps } from './ContextSurface'

export { WindowManager } from './WindowManager'
export type { WindowManagerProps, WindowState } from './WindowManager'

export { WorkspaceTabs } from './WorkspaceTabs'
export type { WorkspaceTabsProps, WorkspaceTab } from './WorkspaceTabs'

export { MultiMonitorWorkspace, useScreenDetection } from './MultiMonitorWorkspace'
export type { MultiMonitorWorkspaceProps, Monitor } from './MultiMonitorWorkspace'

// Immersive Mode Components
export { ImmersiveMode } from './ImmersiveMode'
export type { ImmersiveModeProps } from './ImmersiveMode'

export { CommandOverlay } from './CommandOverlay'
export type { CommandOverlayProps, CommandItem } from './CommandOverlay'

// Navigation Components
export { StudioTopbar } from './StudioTopbar'
export type { StudioTopbarProps } from './StudioTopbar'

export { ActivityRail } from './ActivityRail'
export type { ActivityRailProps, ActivityItem } from './ActivityRail'

export { QuickActionRail } from './QuickActionRail'
export type { QuickActionRailProps, QuickAction } from './QuickActionRail'

export { BottomTelemetryBar } from './BottomTelemetryBar'
export type { BottomTelemetryBarProps, TelemetryMetric } from './BottomTelemetryBar'

// Studio Store
export { useStudioStore } from '@/stores/studioStore'
export type { StudioPanel, StudioFloatingPanel, StudioWorkspaceTab } from '@/stores/studioStore'
