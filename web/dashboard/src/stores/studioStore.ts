import { create } from 'zustand'
import { immer } from 'zustand/middleware/immer'

export interface StudioPanel {
  id: string
  title: string
  position: 'left' | 'right' | 'top' | 'bottom'
  isVisible: boolean
  isCollapsed: boolean
  size: number
}

export interface StudioFloatingPanel {
  id: string
  title: string
  x: number
  y: number
  width: number
  height: number
  zIndex: number
  isMaximized: boolean
  isMinimized: boolean
}

export interface StudioWorkspaceTab {
  id: string
  label: string
  icon?: string
  isClosable: boolean
}

interface StudioState {
  // Panels
  panels: StudioPanel[]
  floatingPanels: StudioFloatingPanel[]
  
  // Tabs
  tabs: StudioWorkspaceTab[]
  activeTabId: string
  
  // View state
  mode: 'compact' | 'standard' | 'expanded' | 'ultrawide'
  zoomLevel: number
  showMinimap: boolean
  gridEnabled: boolean
  snapToGrid: boolean
  
  // Immersive mode
  immersiveMode: boolean
  
  // Sidebar
  sidebarCollapsed: boolean
  
  // Actions
  setMode: (mode: 'compact' | 'standard' | 'expanded' | 'ultrawide') => void
  setZoomLevel: (level: number) => void
  toggleMinimap: () => void
  toggleGrid: () => void
  toggleSnapToGrid: () => void
  toggleImmersiveMode: () => void
  toggleSidebar: () => void
  
  // Panel actions
  setPanelVisible: (id: string, visible: boolean) => void
  setPanelCollapsed: (id: string, collapsed: boolean) => void
  setPanelSize: (id: string, size: number) => void
  
  // Floating panel actions
  addFloatingPanel: (panel: Omit<StudioFloatingPanel, 'id' | 'zIndex'>) => void
  removeFloatingPanel: (id: string) => void
  updateFloatingPanel: (id: string, updates: Partial<StudioFloatingPanel>) => void
  
  // Tab actions
  addTab: (tab: Omit<StudioWorkspaceTab, 'id'>) => void
  removeTab: (id: string) => void
  setActiveTab: (id: string) => void
}

export const useStudioStore = create<StudioState>()(
  immer((set) => ({
    // Initial state
    panels: [
      { id: 'files', title: 'Files', position: 'left', isVisible: true, isCollapsed: false, size: 250 },
      { id: 'properties', title: 'Properties', position: 'right', isVisible: true, isCollapsed: false, size: 300 },
      { id: 'output', title: 'Output', position: 'bottom', isVisible: true, isCollapsed: false, size: 200 },
      { id: 'problems', title: 'Problems', position: 'bottom', isVisible: false, isCollapsed: false, size: 150 },
    ],
    floatingPanels: [],
    tabs: [
      { id: 'main', label: 'Main Workspace', isClosable: false },
    ],
    activeTabId: 'main',
    mode: 'standard',
    zoomLevel: 1,
    showMinimap: true,
    gridEnabled: true,
    snapToGrid: true,
    immersiveMode: false,
    sidebarCollapsed: false,

    // Actions
    setMode: (mode) => set((state) => { state.mode = mode }),
    setZoomLevel: (level) => set((state) => { state.zoomLevel = level }),
    toggleMinimap: () => set((state) => { state.showMinimap = !state.showMinimap }),
    toggleGrid: () => set((state) => { state.gridEnabled = !state.gridEnabled }),
    toggleSnapToGrid: () => set((state) => { state.snapToGrid = !state.snapToGrid }),
    toggleImmersiveMode: () => set((state) => { state.immersiveMode = !state.immersiveMode }),
    toggleSidebar: () => set((state) => { state.sidebarCollapsed = !state.sidebarCollapsed }),

    // Panel actions
    setPanelVisible: (id, visible) => set((state) => {
      const panel = state.panels.find(p => p.id === id)
      if (panel) panel.isVisible = visible
    }),
    setPanelCollapsed: (id, collapsed) => set((state) => {
      const panel = state.panels.find(p => p.id === id)
      if (panel) panel.isCollapsed = collapsed
    }),
    setPanelSize: (id, size) => set((state) => {
      const panel = state.panels.find(p => p.id === id)
      if (panel) panel.size = size
    }),

    // Floating panel actions
    addFloatingPanel: (panel) => set((state) => {
      const id = `float-${Date.now()}`
      state.floatingPanels.push({ ...panel, id, zIndex: state.floatingPanels.length + 1 })
    }),
    removeFloatingPanel: (id) => set((state) => {
      state.floatingPanels = state.floatingPanels.filter(p => p.id !== id)
    }),
    updateFloatingPanel: (id, updates) => set((state) => {
      const panel = state.floatingPanels.find(p => p.id === id)
      if (panel) Object.assign(panel, updates)
    }),

    // Tab actions
    addTab: (tab) => set((state) => {
      const id = `tab-${Date.now()}`
      state.tabs.push({ ...tab, id })
    }),
    removeTab: (id) => set((state) => {
      state.tabs = state.tabs.filter(t => t.id !== id)
      if (state.activeTabId === id && state.tabs.length > 0) {
        state.activeTabId = state.tabs[0].id
      }
    }),
    setActiveTab: (id) => set((state) => { state.activeTabId = id }),
  }))
)
