/**
 * @functionfly/ui-robotics
 * Robotics & Physical Systems Components
 */

// ============================================================================
// Robot Fleet Dashboard
// ============================================================================

export type RobotStatus = 'online' | 'offline' | 'busy' | 'error' | 'maintenance';
export type RobotType = 'ground' | 'aerial' | 'aquatic' | 'stationary';

export interface Robot {
  id: string;
  name: string;
  type: RobotType;
  status: RobotStatus;
  batteryLevel: number;
  position: { x: number; y: number };
  lastSeen: number;
  metadata?: Record<string, unknown>;
}

export interface Fleet {
  id: string;
  name: string;
  robots: Robot[];
  totalCount: number;
  onlineCount: number;
  busyCount: number;
  errorCount: number;
}

export interface RobotFleetDashboardProps {
  fleet: Fleet;
  selectedRobotId?: string | null;
  onRobotSelect?: (robot: Robot) => void;
  onRobotHover?: (robot: Robot | null) => void;
  className?: string;
}

// ============================================================================
// Sensor Telemetry Panel
// ============================================================================

export interface SensorReading {
  id: string;
  name: string;
  value: number;
  unit: string;
  timestamp: number;
  status: 'normal' | 'warning' | 'critical';
}

export interface SensorTelemetryPanelProps {
  robotId: string;
  robotName: string;
  readings: SensorReading[];
  onRefresh?: () => void;
  className?: string;
}

// ============================================================================
// Robot Command Center
// ============================================================================

export interface Command {
  id: string;
  name: string;
  description: string;
  type: 'move' | 'stop' | 'patrol' | 'return' | 'custom';
  status: 'pending' | 'sent' | 'acknowledged' | 'completed' | 'failed';
  issuedAt: number;
  acknowledgedAt?: number;
  completedAt?: number;
}

export interface RobotCommandCenterProps {
  robotId: string;
  robotName: string;
  commands: Command[];
  onSendCommand?: (command: Command) => void;
  onCancelCommand?: (commandId: string) => void;
  className?: string;
}

// ============================================================================
// Physical Environment Map
// ============================================================================

export interface MapWaypoint {
  id: string;
  name: string;
  x: number;
  y: number;
  type: 'charging' | 'checkpoint' | 'target' | 'hazard';
}

export interface Obstacle {
  id: string;
  type: 'static' | 'dynamic';
  position: { x: number; y: number };
  dimensions: { width: number; height: number };
  label?: string;
}

export interface PhysicalEnvironmentMapProps {
  robots: Robot[];
  waypoints: MapWaypoint[];
  obstacles: Obstacle[];
  selectedRobotId?: string | null;
  onRobotSelect?: (robot: Robot) => void;
  onWaypointClick?: (waypoint: MapWaypoint) => void;
  className?: string;
}

// ============================================================================
// Drone Flight Overlay
// ============================================================================

export interface FlightPath {
  id: string;
  name: string;
  points: Array<{ x: number; y: number; altitude: number }>;
  status: 'planned' | 'active' | 'completed';
  estimatedDuration: number;
  totalDistance: number;
}

export interface DroneFlightOverlayProps {
  droneId: string;
  droneName: string;
  currentPosition: { x: number; y: number; altitude: number };
  flightPaths: FlightPath[];
  activePathId?: string | null;
  onPathSelect?: (path: FlightPath) => void;
  className?: string;
}

// ============================================================================
// Robot Vision Stream
// ============================================================================

export interface VisionFrame {
  id: string;
  timestamp: number;
  width: number;
  height: number;
  objects: Array<{
    label: string;
    confidence: number;
    boundingBox: { x: number; y: number; width: number; height: number };
  }>;
}

export interface RobotVisionStreamProps {
  robotId: string;
  robotName: string;
  frames: VisionFrame[];
  isStreaming?: boolean;
  onFrameSelect?: (frame: VisionFrame) => void;
  className?: string;
}

// ============================================================================
// Device Mesh Viewer
// ============================================================================

export interface MeshNode {
  id: string;
  name: string;
  type: 'robot' | 'sensor' | 'gateway' | 'controller';
  status: RobotStatus;
  connections: string[];
  metrics?: {
    cpu?: number;
    memory?: number;
    latency?: number;
  };
}

export interface DeviceMeshViewerProps {
  nodes: MeshNode[];
  selectedNodeId?: string | null;
  onNodeSelect?: (node: MeshNode) => void;
  onNodeHover?: (node: MeshNode | null) => void;
  className?: string;
}

// ============================================================================
// Actuator Control Panel
// ============================================================================

export interface Actuator {
  id: string;
  name: string;
  type: 'servo' | 'motor' | 'pneumatic' | 'hydraulic';
  currentPosition: number;
  targetPosition: number;
  speed: number;
  torque?: number;
  status: RobotStatus;
}

export interface ActuatorControlPanelProps {
  robotId: string;
  robotName: string;
  actuators: Actuator[];
  onActuatorControl?: (actuator: Actuator, targetPosition: number) => void;
  onEmergencyStop?: () => void;
  className?: string;
}

// ============================================================================
// Edge Device Monitor
// ============================================================================

export interface EdgeMetric {
  id: string;
  name: string;
  value: number;
  unit: string;
  trend: 'up' | 'down' | 'stable';
  threshold?: { warning: number; critical: number };
}

export interface EdgeDevice {
  id: string;
  name: string;
  type: 'compute' | 'inference' | 'gateway';
  status: RobotStatus;
  ipAddress: string;
  metrics: EdgeMetric[];
  uptime: number;
}

export interface EdgeDeviceMonitorProps {
  devices: EdgeDevice[];
  selectedDeviceId?: string | null;
  onDeviceSelect?: (device: EdgeDevice) => void;
  onDeviceRestart?: (deviceId: string) => void;
  className?: string;
}

// ============================================================================
// Robotic Workflow Designer
// ============================================================================

export interface WorkflowStep {
  id: string;
  name: string;
  type: 'navigation' | 'manipulation' | 'inspection' | 'monitoring' | 'charging';
  duration: number;
  robotType?: RobotType;
  params?: Record<string, unknown>;
}

export interface RoboticWorkflow {
  id: string;
  name: string;
  steps: WorkflowStep[];
  totalDuration: number;
  assignedRobots: string[];
  status: 'draft' | 'scheduled' | 'running' | 'paused' | 'completed' | 'cancelled';
}

export interface RoboticWorkflowDesignerProps {
  workflows: RoboticWorkflow[];
  selectedWorkflowId?: string | null;
  selectedStepId?: string | null;
  onWorkflowSelect?: (workflow: RoboticWorkflow) => void;
  onStepSelect?: (step: WorkflowStep) => void;
  onWorkflowCreate?: () => void;
  onStepAdd?: (workflowId: string, step: WorkflowStep) => void;
  onStepRemove?: (workflowId: string, stepId: string) => void;
  className?: string;
}
