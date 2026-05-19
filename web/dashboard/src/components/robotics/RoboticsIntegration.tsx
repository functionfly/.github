/**
 * Robotics Integration Component
 * Unified panel that wires all robotics sub-components together
 */

import React, { useState, useEffect, useMemo } from 'react'
import { cn } from '@functionfly/ui-core'
import {
  Bot,
  Radio,
  Crosshair,
  Map,
  Navigation,
  Eye,
  Network,
  Cpu,
  Server,
  Workflow,
  Gauge,
  Wifi,
  WifiOff,
  AlertTriangle,
  ChevronRight,
  ChevronLeft,
  Plus,
  Settings,
} from 'lucide-react'
import {
  RobotFleetDashboard,
  SensorTelemetryPanel,
  RobotCommandCenter,
  PhysicalEnvironmentMap,
  DroneFlightOverlay,
  RobotVisionStream,
  DeviceMeshViewer,
  ActuatorControlPanel,
  EdgeDeviceMonitor,
  RoboticWorkflowDesigner,
  type Robot,
  type Fleet,
  type SensorReading,
  type Command,
  type MapWaypoint,
  type Obstacle,
  type FlightPath,
  type VisionFrame,
  type MeshNode,
  type Actuator,
  type EdgeDevice,
  type RoboticWorkflow,
} from '@functionfly/ui-robotics'
import { useRoboticsStore } from '@/stores/roboticsStore'

// Panel navigation items
const PANEL_NAV_ITEMS = [
  { id: 'fleet', label: 'Fleet Dashboard', icon: Bot },
  { id: 'sensor', label: 'Sensor Telemetry', icon: Radio },
  { id: 'command', label: 'Command Center', icon: Crosshair },
  { id: 'map', label: 'Environment Map', icon: Map },
  { id: 'drone', label: 'Drone Flight', icon: Navigation },
  { id: 'vision', label: 'Vision Stream', icon: Eye },
  { id: 'mesh', label: 'Device Mesh', icon: Network },
  { id: 'actuator', label: 'Actuator Control', icon: Cpu },
  { id: 'edge', label: 'Edge Monitor', icon: Server },
  { id: 'workflow', label: 'Workflow Designer', icon: Workflow },
] as const

type PanelId = typeof PANEL_NAV_ITEMS[number]['id']

// Mock data generators
const generateMockRobots = (): Robot[] => [
  { id: 'r1', name: 'Alpha-1', type: 'ground', status: 'online', batteryLevel: 85, position: { x: 120, y: 80 }, lastSeen: Date.now() },
  { id: 'r2', name: 'Alpha-2', type: 'ground', status: 'online', batteryLevel: 72, position: { x: 180, y: 150 }, lastSeen: Date.now() },
  { id: 'r3', name: 'Beta-1', type: 'aerial', status: 'busy', batteryLevel: 94, position: { x: 280, y: 100 }, lastSeen: Date.now() },
  { id: 'r4', name: 'Beta-2', type: 'aerial', status: 'online', batteryLevel: 67, position: { x: 350, y: 180 }, lastSeen: Date.now() },
  { id: 'r5', name: 'Gamma-1', type: 'stationary', status: 'maintenance', batteryLevel: 100, position: { x: 80, y: 220 }, lastSeen: Date.now() - 300000 },
  { id: 'r6', name: 'Delta-1', type: 'aquatic', status: 'error', batteryLevel: 15, position: { x: 250, y: 250 }, lastSeen: Date.now() },
]

const generateMockFleet = (robots: Robot[]): Fleet => ({
  id: 'fleet-1',
  name: 'Main Fleet',
  robots,
  totalCount: robots.length,
  onlineCount: robots.filter(r => r.status === 'online').length,
  busyCount: robots.filter(r => r.status === 'busy').length,
  errorCount: robots.filter(r => r.status === 'error').length,
})

const generateMockSensorReadings = (): SensorReading[] => [
  { id: 's1', name: 'Temperature', value: 42.5, unit: '°C', timestamp: Date.now(), status: 'normal' },
  { id: 's2', name: 'Pressure', value: 101.3, unit: 'kPa', timestamp: Date.now(), status: 'normal' },
  { id: 's3', name: 'Humidity', value: 65.2, unit: '%', timestamp: Date.now(), status: 'warning' },
  { id: 's4', name: 'IMU Accelerometer', value: 0.02, unit: 'g', timestamp: Date.now(), status: 'normal' },
  { id: 's5', name: 'Ultrasonic Distance', value: 125.0, unit: 'cm', timestamp: Date.now(), status: 'normal' },
]

const generateMockCommands = (): Command[] => [
  { id: 'c1', name: 'Move Forward', description: 'Navigate 10m forward', type: 'move', status: 'completed', issuedAt: Date.now() - 120000, completedAt: Date.now() - 115000 },
  { id: 'c2', name: 'Turn Left', description: 'Rotate 90 degrees left', type: 'move', status: 'acknowledged', issuedAt: Date.now() - 60000, acknowledgedAt: Date.now() - 59000 },
  { id: 'c3', name: 'Stop', description: 'Immediate halt', type: 'stop', status: 'pending', issuedAt: Date.now() - 30000 },
]

const generateMockWaypoints = (): MapWaypoint[] => [
  { id: 'w1', name: 'Charging Station A', x: 50, y: 50, type: 'charging' },
  { id: 'w2', name: 'Checkpoint Alpha', x: 150, y: 100, type: 'checkpoint' },
  { id: 'w3', name: 'Target Zone B', x: 300, y: 200, type: 'target' },
  { id: 'w4', name: 'Hazard Area', x: 200, y: 250, type: 'hazard' },
]

const generateMockObstacles = (): Obstacle[] => [
  { id: 'o1', type: 'static', position: { x: 100, y: 150 }, dimensions: { width: 40, height: 30 }, label: 'Wall' },
  { id: 'o2', type: 'dynamic', position: { x: 250, y: 120 }, dimensions: { width: 25, height: 25 }, label: 'Moving Object' },
]

const generateMockFlightPaths = (): FlightPath[] => [
  {
    id: 'fp1',
    name: 'Survey Route A',
    points: [
      { x: 50, y: 150, altitude: 50 },
      { x: 150, y: 100, altitude: 75 },
      { x: 300, y: 150, altitude: 100 },
      { x: 350, y: 200, altitude: 80 },
    ],
    status: 'active',
    estimatedDuration: 600,
    totalDistance: 450,
  },
  {
    id: 'fp2',
    name: 'Survey Route B',
    points: [
      { x: 350, y: 200, altitude: 80 },
      { x: 280, y: 250, altitude: 60 },
      { x: 150, y: 280, altitude: 40 },
    ],
    status: 'planned',
    estimatedDuration: 420,
    totalDistance: 320,
  },
]

const generateMockVisionFrames = (): VisionFrame[] => [
  {
    id: 'vf1',
    timestamp: Date.now() - 2000,
    width: 640,
    height: 480,
    objects: [
      { label: 'person', confidence: 0.92, boundingBox: { x: 0.1, y: 0.2, width: 0.15, height: 0.4 } },
      { label: 'vehicle', confidence: 0.85, boundingBox: { x: 0.6, y: 0.5, width: 0.3, height: 0.35 } },
    ],
  },
  {
    id: 'vf2',
    timestamp: Date.now() - 1000,
    width: 640,
    height: 480,
    objects: [
      { label: 'person', confidence: 0.89, boundingBox: { x: 0.12, y: 0.18, width: 0.14, height: 0.42 } },
    ],
  },
]

const generateMockMeshNodes = (): MeshNode[] => [
  { id: 'n1', name: 'Gateway-1', type: 'gateway', status: 'online', connections: ['n2', 'n3'], metrics: { cpu: 45, memory: 38, latency: 12 } },
  { id: 'n2', name: 'Robot-Alpha', type: 'robot', status: 'online', connections: ['n1', 'n4'], metrics: { cpu: 72, memory: 65, latency: 8 } },
  { id: 'n3', name: 'Sensor-1', type: 'sensor', status: 'online', connections: ['n1'], metrics: { cpu: 12, memory: 8, latency: 5 } },
  { id: 'n4', name: 'Controller-1', type: 'controller', status: 'busy', connections: ['n2', 'n5'], metrics: { cpu: 89, memory: 72, latency: 15 } },
  { id: 'n5', name: 'Robot-Beta', type: 'robot', status: 'online', connections: ['n4'], metrics: { cpu: 55, memory: 48, latency: 10 } },
]

const generateMockActuators = (): Actuator[] => [
  { id: 'a1', name: 'Left Wheel', type: 'motor', currentPosition: 45, targetPosition: 50, speed: 100, torque: 80, status: 'online' },
  { id: 'a2', name: 'Right Wheel', type: 'motor', currentPosition: 48, targetPosition: 50, speed: 98, torque: 78, status: 'online' },
  { id: 'a3', name: 'Arm Joint', type: 'servo', currentPosition: 90, targetPosition: 85, speed: 50, torque: 40, status: 'online' },
]

const generateMockEdgeDevices = (): EdgeDevice[] => [
  { id: 'e1', name: 'Edge-Compute-1', type: 'compute', status: 'online', ipAddress: '192.168.1.101', uptime: 86400, metrics: [
    { id: 'm1', name: 'CPU', value: 45, unit: '%', trend: 'stable', threshold: { warning: 70, critical: 90 } },
    { id: 'm2', name: 'Memory', value: 62, unit: '%', trend: 'up', threshold: { warning: 75, critical: 90 } },
  ]},
  { id: 'e2', name: 'Inference-1', type: 'inference', status: 'busy', ipAddress: '192.168.1.102', uptime: 43200, metrics: [
    { id: 'm3', name: 'CPU', value: 88, unit: '%', trend: 'up', threshold: { warning: 70, critical: 90 } },
    { id: 'm4', name: 'Memory', value: 45, unit: '%', trend: 'stable', threshold: { warning: 75, critical: 90 } },
  ]},
]

const generateMockWorkflows = (): RoboticWorkflow[] => [
  {
    id: 'wf1',
    name: 'Morning Patrol',
    steps: [
      { id: 's1', name: 'Start Navigation', type: 'navigation', duration: 60, robotType: 'ground' },
      { id: 's2', name: 'Inspect Perimeter', type: 'inspection', duration: 300, robotType: 'ground' },
      { id: 's3', name: 'Monitor Sensors', type: 'monitoring', duration: 180 },
      { id: 's4', name: 'Return to Base', type: 'navigation', duration: 120, robotType: 'ground' },
    ],
    totalDuration: 660,
    assignedRobots: ['r1', 'r2'],
    status: 'scheduled',
  },
]

interface RoboticsIntegrationProps {
  className?: string
  initialView?: PanelId
}

export const RoboticsIntegration: React.FC<RoboticsIntegrationProps> = ({ 
  className,
  initialView = 'fleet' 
}) => {
  const {
    robots,
    selectedRobotId,
    activeView,
    isConnected,
    alerts,
    setRobots,
    selectRobot,
    setActiveView,
    toggleConnection,
  } = useRoboticsStore()

  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const [activePanel, setActivePanel] = useState<PanelId>(initialView)

  // Initialize with mock data
  useEffect(() => {
    if (robots.length === 0) {
      setRobots(generateMockRobots())
    }
  }, [robots.length, setRobots])

  // Sync activeView with activePanel
  useEffect(() => {
    setActiveView(activeView)
  }, [activeView, setActiveView])

  const selectedRobot = robots.find(r => r.id === selectedRobotId) || robots[0]
  const mockFleet = useMemo(() => generateMockFleet(robots), [robots])
  const mockReadings = useMemo(() => generateMockSensorReadings(), [])
  const mockCommands = useMemo(() => generateMockCommands(), [])
  const mockWaypoints = useMemo(() => generateMockWaypoints(), [])
  const mockObstacles = useMemo(() => generateMockObstacles(), [])
  const mockFlightPaths = useMemo(() => generateMockFlightPaths(), [])
  const mockVisionFrames = useMemo(() => generateMockVisionFrames(), [])
  const mockMeshNodes = useMemo(() => generateMockMeshNodes(), [])
  const mockActuators = useMemo(() => generateMockActuators(), [])
  const mockEdgeDevices = useMemo(() => generateMockEdgeDevices(), [])
  const mockWorkflows = useMemo(() => generateMockWorkflows(), [])

  const handlePanelChange = (panelId: PanelId) => {
    setActivePanel(panelId)
    setActiveView(panelId as any)
  }

  const renderPanel = () => {
    switch (activePanel) {
      case 'fleet':
        return (
          <RobotFleetDashboard
            fleet={mockFleet}
            selectedRobotId={selectedRobotId}
            onRobotSelect={(robot) => selectRobot(robot.id)}
            className="h-full"
          />
        )
      case 'sensor':
        return (
          <SensorTelemetryPanel
            robotId={selectedRobot?.id || 'r1'}
            robotName={selectedRobot?.name || 'Robot'}
            readings={mockReadings}
            onRefresh={() => {}}
            className="h-full"
          />
        )
      case 'command':
        return (
          <RobotCommandCenter
            robotId={selectedRobot?.id || 'r1'}
            robotName={selectedRobot?.name || 'Robot'}
            commands={mockCommands}
            onSendCommand={(cmd) => console.log('Send command:', cmd)}
            onCancelCommand={(id) => console.log('Cancel command:', id)}
            className="h-full"
          />
        )
      case 'map':
        return (
          <PhysicalEnvironmentMap
            robots={robots}
            waypoints={mockWaypoints}
            obstacles={mockObstacles}
            selectedRobotId={selectedRobotId}
            onRobotSelect={(robot) => selectRobot(robot.id)}
            onWaypointClick={(wp) => console.log('Waypoint clicked:', wp)}
            className="h-full"
          />
        )
      case 'drone':
        return (
          <DroneFlightOverlay
            droneId={selectedRobot?.id || 'r3'}
            droneName={selectedRobot?.name || 'Drone'}
            currentPosition={selectedRobot?.position ? { ...selectedRobot.position, altitude: 75 } : { x: 280, y: 100, altitude: 75 }}
            flightPaths={mockFlightPaths}
            activePathId="fp1"
            onPathSelect={(path) => console.log('Path selected:', path)}
            className="h-full"
          />
        )
      case 'vision':
        return (
          <RobotVisionStream
            robotId={selectedRobot?.id || 'r1'}
            robotName={selectedRobot?.name || 'Robot'}
            frames={mockVisionFrames}
            isStreaming={true}
            onFrameSelect={(frame) => console.log('Frame selected:', frame)}
            className="h-full"
          />
        )
      case 'mesh':
        return (
          <DeviceMeshViewer
            nodes={mockMeshNodes}
            selectedNodeId={selectedRobotId}
            onNodeSelect={(node) => console.log('Node selected:', node)}
            onNodeHover={(node) => console.log('Node hovered:', node)}
            className="h-full"
          />
        )
      case 'actuator':
        return (
          <ActuatorControlPanel
            robotId={selectedRobot?.id || 'r1'}
            robotName={selectedRobot?.name || 'Robot'}
            actuators={mockActuators}
            onActuatorControl={(actuator, position) => console.log('Actuator control:', actuator, position)}
            onEmergencyStop={() => console.log('Emergency stop')}
            className="h-full"
          />
        )
      case 'edge':
        return (
          <EdgeDeviceMonitor
            devices={mockEdgeDevices}
            selectedDeviceId={selectedRobotId}
            onDeviceSelect={(device) => console.log('Device selected:', device)}
            onDeviceRestart={(deviceId) => console.log('Device restart:', deviceId)}
            className="h-full"
          />
        )
      case 'workflow':
        return (
          <RoboticWorkflowDesigner
            workflows={mockWorkflows}
            selectedWorkflowId="wf1"
            selectedStepId="s1"
            onWorkflowSelect={(workflow) => console.log('Workflow selected:', workflow)}
            onStepSelect={(step) => console.log('Step selected:', step)}
            onWorkflowCreate={() => console.log('Create workflow')}
            onStepAdd={(workflowId, step) => console.log('Add step:', workflowId, step)}
            onStepRemove={(workflowId, stepId) => console.log('Remove step:', workflowId, stepId)}
            className="h-full"
          />
        )
      default:
        return (
          <RobotFleetDashboard
            fleet={mockFleet}
            selectedRobotId={selectedRobotId}
            onRobotSelect={(robot) => selectRobot(robot.id)}
            className="h-full"
          />
        )
    }
  }

  const unacknowledgedAlerts = alerts.filter(a => !a.acknowledged)

  return (
    <div className={cn('flex h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Navigation Sidebar */}
      <div className={cn(
        'flex flex-col border-r border-aviation-border-panel transition-all duration-300',
        sidebarCollapsed ? 'w-12' : 'w-56'
      )}>
        {/* Collapse Toggle */}
        <div className="flex items-center justify-end px-2 py-2 border-b border-aviation-border-panel">
          <button
            onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
            className="p-1.5 hover:bg-aviation-bg-instrument rounded transition-colors"
          >
            {sidebarCollapsed ? <ChevronRight className="w-4 h-4" /> : <ChevronLeft className="w-4 h-4" />}
          </button>
        </div>

        {/* Navigation Items */}
        <nav className="flex-1 overflow-auto py-2">
          {PANEL_NAV_ITEMS.map((item) => {
            const Icon = item.icon
            const isActive = activePanel === item.id
            return (
              <button
                key={item.id}
                onClick={() => handlePanelChange(item.id)}
                className={cn(
                  'flex items-center gap-3 w-full px-3 py-2 text-left transition-colors',
                  isActive ? 'bg-aviation-cyan/20 text-aviation-cyan border-l-2 border-aviation-cyan' : 'text-aviation-text-muted hover:text-aviation-text-primary hover:bg-aviation-bg-secondary',
                  sidebarCollapsed && 'justify-center px-0'
                )}
                title={sidebarCollapsed ? item.label : undefined}
              >
                <Icon className="w-4 h-4 flex-shrink-0" />
                {!sidebarCollapsed && <span className="text-sm truncate">{item.label}</span>}
              </button>
            )
          })}
        </nav>

        {/* Status Indicator */}
        {!sidebarCollapsed && (
          <div className="px-3 py-2 border-t border-aviation-border-panel">
            <button
              onClick={toggleConnection}
              className={cn(
                'flex items-center gap-2 w-full px-2 py-1.5 rounded transition-colors',
                isConnected ? 'bg-green-500/20 hover:bg-green-500/30' : 'bg-red-500/20 hover:bg-red-500/30'
              )}
            >
              {isConnected ? <Wifi className="w-3 h-3 text-green-400" /> : <WifiOff className="w-3 h-3 text-red-400" />}
              <span className={cn('text-xs', isConnected ? 'text-green-400' : 'text-red-400')}>
                {isConnected ? 'Connected' : 'Disconnected'}
              </span>
            </button>
          </div>
        )}
      </div>

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-aviation-border-panel bg-aviation-bg-secondary">
          <div className="flex items-center gap-2">
            <Bot className="w-5 h-5 text-aviation-cyan" />
            <span className="text-sm font-medium">Robotics Integration</span>
            {unacknowledgedAlerts.length > 0 && (
              <div className="flex items-center gap-1.5 px-2 py-1 bg-red-500/20 rounded border border-red-500/50">
                <AlertTriangle className="w-3 h-3 text-red-400" />
                <span className="text-xs text-red-400">{unacknowledgedAlerts.length} Alert{unacknowledgedAlerts.length > 1 ? 's' : ''}</span>
              </div>
            )}
          </div>
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-2 text-xs text-aviation-text-muted">
              <Gauge className="w-4 h-4" />
              <span>{PANEL_NAV_ITEMS.find(i => i.id === activePanel)?.label}</span>
              {selectedRobot && (
                <>
                  <span className="text-aviation-text-muted mx-1">•</span>
                  <span className="text-aviation-text-dim">{selectedRobot.name}</span>
                </>
              )}
            </div>
          </div>
        </div>

        {/* Content Panel */}
        <div className="flex-1 overflow-hidden">
          {renderPanel()}
        </div>
      </div>
    </div>
  )
}

export default RoboticsIntegration
