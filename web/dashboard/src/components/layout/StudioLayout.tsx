import React, { useCallback, useMemo } from 'react'
import { Outlet } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { StudioShell } from '@/components/studio/StudioShell'
import { StudioTopbar } from '@/components/studio/StudioTopbar'
import { SmartSidebar } from '@/components/studio/SmartSidebar'
import { ActivityRail } from '@/components/studio/ActivityRail'
import { QuickActionRail } from '@/components/studio/QuickActionRail'
import { BottomTelemetryBar } from '@/components/studio/BottomTelemetryBar'
import { DockLayout } from '@/components/studio/DockLayout'
import { WorkspaceTabs } from '@/components/studio/WorkspaceTabs'
import { useStudioStore } from '@/stores/studioStore'
import { Activity, Workflow, Bot, Settings, Play, Pause, Save, Plus, GitBranch, Cpu, FileText, Terminal, Sparkles, DollarSign, Brain, Boxes, BarChart3, Gauge } from 'lucide-react'

interface StudioLayoutProps {
  children?: React.ReactNode
}

/**
 * StudioLayout - Full Studio layout with all shell components wired up
 * Uses StudioStore for global state management
 */
export function StudioLayout({ children }: StudioLayoutProps) {
  const {
    panels,
    tabs,
    activeTabId,
    sidebarCollapsed,
    toggleSidebar,
    setActiveTab,
  } = useStudioStore()

// Default sidebar sections
   const sidebarSections = useMemo(() => [
     { id: 'studio', label: 'Studio', path: '/studio', icon: <Activity className="w-5 h-5" /> },
     { id: 'workflows', label: 'Workflows', path: '/studio/workflows', icon: <GitBranch className="w-5 h-5" />, badge: 3 },
     { id: 'agents', label: 'Agents', path: '/studio/agents', icon: <Bot className="w-5 h-5" />, badge: 1 },
     { id: 'functions', label: 'Functions', path: '/studio/functions', icon: <Cpu className="w-5 h-5" /> },
     { id: 'analytics', label: 'Analytics', path: '/studio/analytics', icon: <Activity className="w-5 h-5" /> },
     { id: 'settings', label: 'Settings', path: '/studio/settings', icon: <Settings className="w-5 h-5" /> },
    {
      id: 'labs',
      label: 'Labs',
      icon: <Sparkles className="w-5 h-5" />,
      children: [
        { id: 'simulation', label: 'R-Sim', icon: <Gauge className="w-5 h-5" />, path: '/simulation' },
        { id: 'robotics', label: 'Robotics', icon: <Bot className="w-5 h-5" />, path: '/robotics' },
        { id: 'marketplace-economy', label: 'Marketplace Economy', icon: <DollarSign className="w-5 h-5" />, path: '/marketplace-economy' },
        { id: 'adaptive-ux', label: 'Adaptive UX', icon: <Brain className="w-5 h-5" />, path: '/adaptive-ux' },
        { id: 'universal-runtime', label: 'Universal Runtime', icon: <Boxes className="w-5 h-5" />, path: '/universal-runtime' },
        { id: 'data-visualization', label: 'Data Visualization', icon: <BarChart3 className="w-5 h-5" />, path: '/data-visualization' },
        { id: 'futuristic', label: 'Futuristic', icon: <Sparkles className="w-5 h-5" />, path: '/futuristic' },
      ],
    },
  ], [])

// Default activity items
   const activityItems = useMemo(() => [
     { id: 'studio', label: 'Studio', icon: <Activity className="w-5 h-5" /> },
     { id: 'workflows', label: 'Workflows', icon: <GitBranch className="w-5 h-5" />, badge: 3 },
     { id: 'agents', label: 'Agents', icon: <Bot className="w-5 h-5" />, badge: 1 },
     { id: 'functions', label: 'Functions', icon: <Cpu className="w-5 h-5" /> },
   ], [])

  // Quick actions
  const quickActions = useMemo(() => [
    { id: 'new', label: 'New', icon: <Plus className="w-4 h-4" />, onClick: () => console.log('New') },
    { id: 'save', label: 'Save', icon: <Save className="w-4 h-4" />, onClick: () => console.log('Save') },
    { id: 'run', label: 'Run', icon: <Play className="w-4 h-4" />, onClick: () => console.log('Run'), variant: 'primary' as const },
    { id: 'debug', label: 'Debug', icon: <Pause className="w-4 h-4" />, onClick: () => console.log('Debug') },
  ], [])

  // Telemetry metrics
  const metrics = useMemo(() => [
    { id: 'latency', label: 'Latency', value: '42', unit: 'ms', icon: <Activity className="w-3 h-3" />, status: 'success' as const },
    { id: 'requests', label: 'Requests', value: '1.2K', unit: '/hr', icon: <Activity className="w-3 h-3" />, status: 'info' as const },
    { id: 'cost', label: 'Cost', value: '$24.50', icon: <Play className="w-3 h-3" />, status: 'warning' as const },
    { id: 'uptime', label: 'Uptime', value: '99.9', unit: '%', icon: <Activity className="w-3 h-3" />, status: 'success' as const },
  ], [])

  // Map store panels to DockLayout panels
  const dockPanels = useMemo(() => panels.map(panel => ({
    ...panel,
    content: panel.id === 'files' 
      ? <FileExplorerContent />
      : panel.id === 'output'
      ? <TerminalContent />
      : <div className="p-4 text-sm text-aviation-text-muted">{panel.title}</div>,
  })), [panels])

  // Map store tabs to WorkspaceTabs
  const workspaceTabs = useMemo(() => tabs.map(tab => ({
    ...tab,
    content: <div className="p-8 text-aviation-text-muted">{tab.label}</div>,
  })), [tabs])

  return (
    <StudioShell enableGrid showMinimap>
      <div className="flex flex-col h-full">
        {/* Top Bar */}
        <StudioTopbar onMenuClick={() => toggleSidebar()} />

        {/* Main Content */}
        <div className="flex flex-1 overflow-hidden">
{/* Activity Rail - Left */}
           <ActivityRail
             items={activityItems}
             position="left"
             defaultActiveId="studio"
             onItemClick={(id) => console.log(`Activity: ${id}`)}
          />

          {/* Sidebar */}
          <SmartSidebar
            sections={sidebarSections}
            className="h-full"
            defaultCollapsed={sidebarCollapsed}
          />

          {/* Content Area */}
          <div className="flex-1 flex flex-col overflow-hidden">
            {children ? (
              <div className="flex-1 overflow-hidden">{children}</div>
            ) : (
              <>
                {/* Quick Actions Bar */}
                <div className="px-4 py-2 border-b border-aviation-border-panel bg-aviation-bg-secondary/50">
                  <QuickActionRail actions={quickActions} position="left" />
                </div>

                {/* Main Workspace with Tabs */}
                <div className="flex-1 overflow-hidden">
                  <WorkspaceTabs
                    tabs={workspaceTabs}
                    defaultActiveTab={activeTabId}
                    onTabChange={(id) => setActiveTab(id)}
                  />
                </div>

                {/* Dock Panels */}
                <DockLayout
                  panels={dockPanels}
                  className="flex-1"
                />
              </>
            )}

            {/* Telemetry Bar */}
            <BottomTelemetryBar metrics={metrics} showTimestamp />
          </div>
        </div>
      </div>
    </StudioShell>
  )
}

// Placeholder content components
function FileExplorerContent() {
  return (
    <div className="p-2">
      <div className="text-xs font-medium text-aviation-text-muted mb-2 px-2">EXPLORER</div>
      <div className="space-y-1">
        {['src', 'components', 'hooks', 'utils'].map(folder => (
          <div key={folder} className="flex items-center gap-2 px-2 py-1 text-sm text-aviation-text-secondary hover:bg-aviation-bg-instrument rounded cursor-pointer">
            <FileText className="w-4 h-4 text-aviation-amber" />
            {folder}
          </div>
        ))}
      </div>
    </div>
  )
}

function TerminalContent() {
  return (
    <div className="p-2 font-mono text-xs">
      <div className="text-aviation-text-muted mb-1">OUTPUT</div>
      <div className="text-aviation-green">&gt; Ready</div>
    </div>
  )
}
