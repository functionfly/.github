/**
 * Robotics Components Index
 * Re-exports all robotics-related components and hooks
 */

export {
  RoboticsIntegration,
} from './RoboticsIntegration'

// Re-export from @functionfly/ui-robotics
export {
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
  type RobotStatus,
  type RobotType,
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
  type RobotFleetDashboardProps,
  type SensorTelemetryPanelProps,
  type RobotCommandCenterProps,
  type PhysicalEnvironmentMapProps,
  type DroneFlightOverlayProps,
  type RobotVisionStreamProps,
  type DeviceMeshViewerProps,
  type ActuatorControlPanelProps,
  type EdgeDeviceMonitorProps,
  type RoboticWorkflowDesignerProps,
} from '@functionfly/ui-robotics'

// Re-export from robotics store
export {
  useRoboticsStore,
  selectSelectedRobot,
  selectOnlineRobots,
  selectAlertCount,
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
