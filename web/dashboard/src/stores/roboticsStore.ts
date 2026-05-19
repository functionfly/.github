/**
 * Robotics Store
 * Global state management for robotics and physical systems components
 */

import { create } from 'zustand'
import { immer } from 'zustand/middleware/immer'

// ============================================================================
// Types (mirrored from @functionfly/ui-robotics for local store)
// ============================================================================

export type RobotStatus = 'online' | 'offline' | 'busy' | 'error' | 'maintenance'
export type RobotType = 'ground' | 'aerial' | 'aquatic' | 'stationary'

export interface Robot {
  id: string
  name: string
  type: RobotType
  status: RobotStatus
  batteryLevel: number
  position: { x: number; y: number }
  lastSeen: number
  metadata?: Record<string, unknown>
}

export interface RobotAlert {
  id: string
  robotId: string
  robotName: string
  type: 'battery_low' | 'offline' | 'error' | 'maintenance_required'
  severity: 'info' | 'warning' | 'critical'
  message: string
  timestamp: number
  acknowledged: boolean
}

export type RoboticsView = 'fleet' | 'sensor' | 'command' | 'map' | 'drone' | 'vision' | 'mesh' | 'actuator' | 'edge' | 'workflow'

// ============================================================================
// State Interface
// ============================================================================

export interface RoboticsState {
  robots: Robot[]
  selectedRobotId: string | null
  activeView: RoboticsView
  isConnected: boolean
  alerts: RobotAlert[]
}

// ============================================================================
// Actions Interface
// ============================================================================

interface RoboticsActions {
  selectRobot: (robotId: string | null) => void
  setActiveView: (view: RoboticsView) => void
  toggleConnection: () => void
  dismissAlert: (alertId: string) => void
  updateRobotStatus: (robotId: string, status: Partial<Robot>) => void
  setRobots: (robots: Robot[]) => void
  addAlert: (alert: RobotAlert) => void
  clearAlerts: () => void
}

// ============================================================================
// Store
// ============================================================================

export const useRoboticsStore = create<RoboticsState & RoboticsActions>()(
  immer((set) => ({
    // ============================================================================
    // Initial State
    // ============================================================================

    robots: [],
    selectedRobotId: null,
    activeView: 'fleet',
    isConnected: false,
    alerts: [],

    // ============================================================================
    // Actions
    // ============================================================================

    selectRobot: (robotId) =>
      set((state) => {
        state.selectedRobotId = robotId
      }),

    setActiveView: (view) =>
      set((state) => {
        state.activeView = view
      }),

    toggleConnection: () =>
      set((state) => {
        state.isConnected = !state.isConnected
      }),

    dismissAlert: (alertId) =>
      set((state) => {
        const alert = state.alerts.find((a) => a.id === alertId)
        if (alert) {
          alert.acknowledged = true
        }
      }),

    updateRobotStatus: (robotId, statusUpdate) =>
      set((state) => {
        const robot = state.robots.find((r) => r.id === robotId)
        if (robot) {
          Object.assign(robot, statusUpdate)
        }
      }),

    setRobots: (robots) =>
      set((state) => {
        state.robots = robots
      }),

    addAlert: (alert) =>
      set((state) => {
        state.alerts.push(alert)
      }),

    clearAlerts: () =>
      set((state) => {
        state.alerts = []
      }),
  }))
)

// ============================================================================
// Selectors
// ============================================================================

export const selectSelectedRobot = (state: RoboticsState) =>
  state.robots.find((r) => r.id === state.selectedRobotId) || null

export const selectOnlineRobots = (state: RoboticsState) =>
  state.robots.filter((r) => r.status === 'online')

export const selectAlertCount = (state: RoboticsState) =>
  state.alerts.filter((a) => !a.acknowledged).length

// ============================================================================
// Custom Hooks for Robotics Features
// ============================================================================

export const useRoboticsFleet = () =>
  useRoboticsStore((state) => ({
    robots: state.robots,
    selectedRobotId: state.selectedRobotId,
    onlineCount: state.robots.filter((r) => r.status === 'online').length,
    busyCount: state.robots.filter((r) => r.status === 'busy').length,
    errorCount: state.robots.filter((r) => r.status === 'error').length,
  }))

export const useRobotTelemetry = () =>
  useRoboticsStore((state) => ({
    robots: state.robots,
    selectedRobotId: state.selectedRobotId,
  }))

export const useRobotCommands = () =>
  useRoboticsStore((state) => ({
    selectedRobotId: state.selectedRobotId,
    isConnected: state.isConnected,
  }))

export const useEnvironmentMap = () =>
  useRoboticsStore((state) => ({
    robots: state.robots,
    selectedRobotId: state.selectedRobotId,
  }))

export const useDroneFlight = () =>
  useRoboticsStore((state) => ({
    robots: state.robots.filter((r) => r.type === 'aerial'),
    selectedRobotId: state.selectedRobotId,
  }))

export const useVisionStream = () =>
  useRoboticsStore((state) => ({
    robots: state.robots,
    selectedRobotId: state.selectedRobotId,
  }))

export const useDeviceMesh = () =>
  useRoboticsStore((state) => ({
    robots: state.robots,
    alerts: state.alerts,
  }))

export const useActuatorControl = () =>
  useRoboticsStore((state) => ({
    selectedRobotId: state.selectedRobotId,
    isConnected: state.isConnected,
  }))

export const useEdgeMonitor = () =>
  useRoboticsStore((state) => ({
    isConnected: state.isConnected,
    alerts: state.alerts,
  }))

export const useWorkflowDesigner = () =>
  useRoboticsStore((state) => ({
    robots: state.robots,
    selectedRobotId: state.selectedRobotId,
  }))
