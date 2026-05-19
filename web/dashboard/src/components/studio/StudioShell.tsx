import React, { useState, useCallback, useMemo, createContext, useContext } from 'react'
import { cn } from '@/lib/utils'
import { useThemeStore } from '@/stores/themeStore'
import { motion, AnimatePresence } from 'framer-motion'
import { StudioTopbar } from './StudioTopbar'
import { SmartSidebar } from './SmartSidebar'
import { ActivityRail } from './ActivityRail'
import { QuickActionRail } from './QuickActionRail'
import { BottomTelemetryBar } from './BottomTelemetryBar'
import { CommandOverlay, CommandItem } from './CommandOverlay'
import { ImmersiveMode } from './ImmersiveMode'
import { DockLayout, DockPanel } from './DockLayout'
import { FloatingPanelSystem, FloatingPanel } from './FloatingPanelSystem'
import { WorkspaceTabs, WorkspaceTab } from './WorkspaceTabs'
import { SplitViewManager, SplitView } from './SplitViewManager'
import { ResizablePanel } from './ResizablePanel'
import { Activity, Workflow, Bot, Settings, Play, Pause, Save, Plus, GitBranch, Cpu, Sparkles } from 'lucide-react'
import { useAICommandSystem } from '@/hooks/useAICommandSystem'

export interface StudioShellProps {
  children?: React.ReactNode
  className?: string
  enableGrid?: boolean
  showMinimap?: boolean
  showAIPanel?: boolean
}

// Studio Context for global state
interface StudioContextValue {
  immersiveMode: boolean
  setImmersiveMode: (v: boolean) => void
  commandPaletteOpen: boolean
  setCommandPaletteOpen: (v: boolean) => void
  sidebarCollapsed: boolean
  setSidebarCollapsed: (v: boolean) => void
  floatingPanels: FloatingPanel[]
  addFloatingPanel: (panel: Omit<FloatingPanel, 'id' | 'zIndex'>) => void
  removeFloatingPanel: (id: string) => void
  updateFloatingPanel: (id: string, updates: Partial<FloatingPanel>) => void
  // AI
  aiCommandOpen: boolean
  setAICommandOpen: (v: boolean) => void
}

const StudioContext = createContext<StudioContextValue | null>(null)

export function useStudioContext() {
  const ctx = useContext(StudioContext)
  if (!ctx) throw new Error('useStudioContext must be used within StudioShell')
  return ctx
}

/**
 * StudioShell - Main container for the Studio workspace
 * Provides the foundational layout structure with aviation theming
 * Includes AI Command System integration
 */
export function StudioShell({
  children,
  className,
  enableGrid = true,
  showMinimap = false,
  showAIPanel = true,
}: StudioShellProps) {
  const { isDarkMode } = useThemeStore()
  const ai = useAICommandSystem()
  const [isInitialized, setIsInitialized] = useState(false)
  const [immersiveMode, setImmersiveMode] = useState(false)
  const [commandPaletteOpen, setCommandPaletteOpen] = useState(false)
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const [floatingPanels, setFloatingPanels] = useState<FloatingPanel[]>([])

  React.useEffect(() => {
    setIsInitialized(true)
  }, [])

  const addFloatingPanel = useCallback((panel: Omit<FloatingPanel, 'id' | 'zIndex'>) => {
    const id = `float-${Date.now()}`
    setFloatingPanels(prev => [...prev, { ...panel, id, zIndex: prev.length + 1 }])
  }, [])

  const removeFloatingPanel = useCallback((id: string) => {
    setFloatingPanels(prev => prev.filter(p => p.id !== id))
  }, [])

  const updateFloatingPanel = useCallback((id: string, updates: Partial<FloatingPanel>) => {
    setFloatingPanels(prev => prev.map(p => p.id === id ? { ...p, ...updates } : p))
  }, [])

  const contextValue = useMemo(() => ({
    immersiveMode,
    setImmersiveMode,
    commandPaletteOpen,
    setCommandPaletteOpen,
    sidebarCollapsed,
    setSidebarCollapsed,
    floatingPanels,
    addFloatingPanel,
    removeFloatingPanel,
    updateFloatingPanel,
    aiCommandOpen: false,
    setAICommandOpen: () => {},
  }), [immersiveMode, commandPaletteOpen, sidebarCollapsed, floatingPanels, addFloatingPanel, removeFloatingPanel, updateFloatingPanel])

  // Default command palette commands
  const defaultCommands: CommandItem[] = useMemo(() => [
    { id: 'ai-command', title: 'AI Command Panel', group: 'AI', icon: <Sparkles className="size-4" />, onSelect: () => ai.openCommandPalette() },
    { id: 'new-workflow', title: 'New Workflow', group: 'Workflows', onSelect: () => console.log('New workflow') },
    { id: 'new-agent', title: 'New Agent', group: 'Agents', onSelect: () => console.log('New agent') },
    { id: 'save', title: 'Save All', group: 'General', keywords: ['save', 'commit'], onSelect: () => console.log('Save all') },
    { id: 'immersive', title: 'Toggle Immersive Mode', group: 'View', keywords: ['focus', 'fullscreen'], onSelect: () => setImmersiveMode(m => !m) },
    { id: 'command-palette', title: 'Open Command Palette', group: 'Navigation', keywords: ['cmd', 'k'], onSelect: () => setCommandPaletteOpen(true) },
    { id: 'toggle-sidebar', title: 'Toggle Sidebar', group: 'View', onSelect: () => setSidebarCollapsed(c => !c) },
  ], [ai, setImmersiveMode, setCommandPaletteOpen, setSidebarCollapsed])

  return (
    <StudioContext.Provider value={contextValue}>
      <ImmersiveMode isEnabled={immersiveMode} onToggle={() => setImmersiveMode(m => !m)}>
        <div
          className={cn(
            'h-screen w-screen overflow-hidden',
            'bg-aviation-bg-primary text-aviation-text-primary',
            isDarkMode ? 'dark' : '',
            className
          )}
          data-studio-shell
        >
          {/* Background Effects */}
          <div className="absolute inset-0 overflow-hidden pointer-events-none">
            <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-brand-500/10 rounded-full blur-[128px] animate-float" />
            <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-purple-500/10 rounded-full blur-[128px] animate-float-rotate" />
          </div>

          {/* Grid Background */}
          {enableGrid && (
            <div
              className="absolute inset-0 opacity-20 pointer-events-none"
              style={{
                backgroundImage: `
                  linear-gradient(rgba(255,255,255,0.02) 1px, transparent 1px),
                  linear-gradient(90deg, rgba(255,255,255,0.02) 1px, transparent 1px)
                `,
                backgroundSize: '40px 40px',
              }}
            />
          )}

          {/* Main Content */}
          <AnimatePresence>
            {isInitialized && (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ duration: 0.3 }}
                className="relative h-full w-full"
              >
                {children}
              </motion.div>
            )}
          </AnimatePresence>

          {/* Command Palette */}
          <CommandOverlay
            isOpen={commandPaletteOpen}
            onClose={() => setCommandPaletteOpen(false)}
            commands={defaultCommands}
          />
        </div>
      </ImmersiveMode>
    </StudioContext.Provider>
  )
}
