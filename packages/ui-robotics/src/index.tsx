/**
 * @functionfly/ui-robotics
 * Robotics & Physical Systems Components
 */

import React, { useState, useMemo, useCallback } from "react";
import { cn } from "@functionfly/ui-core";
import {
  Bot,
  Radio,
  Antenna,
  MapPin,
  Navigation,
  Play,
  Pause,
  Square,
  RotateCcw,
  AlertTriangle,
  AlertCircle,
  CheckCircle,
  XCircle,
  Battery,
  BatteryCharging,
  Signal,
  Wifi,
  Eye,
  Camera,
  Monitor,
  Cpu,
  MemoryStick,
  Activity,
  Clock,
  Target,
  Map,
  Waypoints,
  ArrowUp,
  ArrowDown,
  ArrowLeft,
  ArrowRight,
  Move,
  Zap,
  Shield,
  Server,
  Network,
  Gauge,
  Timer,
  RefreshCw,
  Plus,
  Minus,
  X,
  Check,
  Settings,
  MoreHorizontal,
  ChevronRight,
  ChevronDown,
  Layers,
  Hexagon,
  Circle,
  Box,
  type LucideIcon,
} from "lucide-react";

// ============================================================================
// Robot Fleet Dashboard
// ============================================================================

import { type RobotStatus, type RobotType } from "./types";
export type { RobotStatus, RobotType } from "./types";
export type { Robot } from "./types";
export type { Fleet } from "./types";
export type { RobotFleetDashboardProps } from "./types";
export type { SensorReading } from "./types";
export type { SensorTelemetryPanelProps } from "./types";
export type { Command } from "./types";
export type { RobotCommandCenterProps } from "./types";
export type { MapWaypoint } from "./types";
export type { Obstacle } from "./types";
export type { PhysicalEnvironmentMapProps } from "./types";
export type { FlightPath } from "./types";
export type { DroneFlightOverlayProps } from "./types";
export type { VisionFrame } from "./types";
export type { RobotVisionStreamProps } from "./types";
export type { MeshNode } from "./types";
export type { DeviceMeshViewerProps } from "./types";
export type { Actuator } from "./types";
export type { ActuatorControlPanelProps } from "./types";
export type { EdgeMetric } from "./types";
export type { EdgeDevice } from "./types";
export type { EdgeDeviceMonitorProps } from "./types";
export type { WorkflowStep } from "./types";
export type { RoboticWorkflow } from "./types";
export type { RoboticWorkflowDesignerProps } from "./types";

interface Robot {
  id: string;
  name: string;
  type: RobotType;
  status: RobotStatus;
  batteryLevel: number;
  position: { x: number; y: number };
  lastSeen: number;
  metadata?: Record<string, unknown>;
}

interface Fleet {
  id: string;
  name: string;
  robots: Robot[];
  totalCount: number;
  onlineCount: number;
  busyCount: number;
  errorCount: number;
}

interface RobotFleetDashboardProps {
  fleet: Fleet;
  selectedRobotId?: string | null;
  onRobotSelect?: (robot: Robot) => void;
  onRobotHover?: (robot: Robot | null) => void;
  className?: string;
}

export const RobotFleetDashboard: React.FC<RobotFleetDashboardProps> = ({
  fleet,
  selectedRobotId,
  onRobotSelect,
  onRobotHover,
  className,
}) => {
  const [hoveredRobot, setHoveredRobot] = useState<string | null>(null);

  const getStatusColor = (status: RobotStatus) => {
    switch (status) {
      case "online":
        return "text-green-400";
      case "busy":
        return "text-aviation-cyan";
      case "offline":
        return "text-aviation-text-muted";
      case "error":
        return "text-red-400";
      case "maintenance":
        return "text-amber-400";
      default:
        return "text-aviation-text-muted";
    }
  };

  const getStatusBg = (status: RobotStatus) => {
    switch (status) {
      case "online":
        return "bg-green-400/20";
      case "busy":
        return "bg-aviation-cyan/20";
      case "offline":
        return "bg-aviation-bg-instrument";
      case "error":
        return "bg-red-400/20";
      case "maintenance":
        return "bg-amber-400/20";
      default:
        return "bg-aviation-bg-instrument";
    }
  };

  const getTypeIcon = (
    type: RobotType,
  ): React.ComponentType<{ className?: string }> => {
    switch (type) {
      case "ground":
        return Bot;
      case "aerial":
        return Navigation;
      case "aquatic":
        return Waves;
      case "stationary":
        return Cpu;
      default:
        return Bot;
    }
  };

  const RobotIcon = getTypeIcon(fleet.robots[0]?.type || "ground");

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Bot className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">
              {fleet.name}
            </h3>
          </div>
          <div className="flex items-center gap-3 text-xs">
            <span className="flex items-center gap-1">
              <span className="w-2 h-2 rounded-full bg-green-400" />
              <span className="text-aviation-text-primary">
                {fleet.onlineCount}
              </span>
            </span>
            <span className="flex items-center gap-1">
              <span className="w-2 h-2 rounded-full bg-aviation-cyan" />
              <span className="text-aviation-text-primary">
                {fleet.busyCount}
              </span>
            </span>
            <span className="flex items-center gap-1">
              <span className="w-2 h-2 rounded-full bg-red-400" />
              <span className="text-aviation-text-primary">
                {fleet.errorCount}
              </span>
            </span>
          </div>
        </div>
      </div>

      {/* Stats Row */}
      <div className="px-4 py-3 border-b border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="grid grid-cols-4 gap-4 text-xs">
          <div>
            <span className="text-aviation-text-dim">Total Units</span>
            <div className="text-lg font-bold text-aviation-text-primary">
              {fleet.totalCount}
            </div>
          </div>
          <div>
            <span className="text-aviation-text-dim">Online</span>
            <div className="text-lg font-bold text-green-400">
              {fleet.onlineCount}
            </div>
          </div>
          <div>
            <span className="text-aviation-text-dim">Busy</span>
            <div className="text-lg font-bold text-aviation-cyan">
              {fleet.busyCount}
            </div>
          </div>
          <div>
            <span className="text-aviation-text-dim">Errors</span>
            <div className="text-lg font-bold text-red-400">
              {fleet.errorCount}
            </div>
          </div>
        </div>
      </div>

      {/* Robot Grid */}
      <div className="flex-1 overflow-auto p-4">
        <div className="grid grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3">
          {fleet.robots.map((robot) => {
            const Icon = getTypeIcon(robot.type);
            const isSelected = selectedRobotId === robot.id;
            const isHovered = hoveredRobot === robot.id;

            return (
              <div
                key={robot.id}
                onClick={() => onRobotSelect?.(robot)}
                onMouseEnter={() => {
                  setHoveredRobot(robot.id);
                  onRobotHover?.(robot);
                }}
                onMouseLeave={() => {
                  setHoveredRobot(null);
                  onRobotHover?.(null);
                }}
                className={cn(
                  "p-3 rounded-lg border cursor-pointer transition-all",
                  isSelected
                    ? "bg-aviation-bg-instrument border-aviation-cyan"
                    : "bg-aviation-bg-secondary border-aviation-border-panel hover:border-aviation-text-muted",
                )}
              >
                <div className="flex items-center justify-between mb-2">
                  <Icon
                    className={cn("w-5 h-5", getStatusColor(robot.status))}
                  />
                  <span
                    className={cn(
                      "px-1.5 py-0.5 text-[10px] rounded uppercase",
                      getStatusBg(robot.status),
                      getStatusColor(robot.status),
                    )}
                  >
                    {robot.status}
                  </span>
                </div>
                <div className="text-sm font-medium text-aviation-text-primary mb-1">
                  {robot.name}
                </div>
                <div className="flex items-center gap-2 text-[10px] text-aviation-text-dim mb-2">
                  <span className="uppercase">{robot.type}</span>
                  <span>•</span>
                  <span>ID: {robot.id.slice(0, 6)}</span>
                </div>
                <div className="flex items-center gap-2">
                  <Battery
                    className={cn(
                      "w-3 h-3",
                      robot.batteryLevel > 20
                        ? "text-green-400"
                        : "text-red-400",
                    )}
                  />
                  <div className="flex-1 h-1.5 bg-aviation-bg-instrument rounded-full overflow-hidden">
                    <div
                      className={cn(
                        "h-full transition-all",
                        robot.batteryLevel > 50
                          ? "bg-green-400"
                          : robot.batteryLevel > 20
                            ? "bg-amber-400"
                            : "bg-red-400",
                      )}
                      style={{ width: `${robot.batteryLevel}%` }}
                    />
                  </div>
                  <span className="text-[10px] text-aviation-text-dim">
                    {robot.batteryLevel}%
                  </span>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Footer */}
      <div className="px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="flex items-center justify-between text-xs text-aviation-text-dim">
          <span>{fleet.robots.length} robots in fleet</span>
          <span>Last updated: {new Date().toLocaleTimeString()}</span>
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Sensor Telemetry Panel
// ============================================================================

interface SensorReading {
  id: string;
  name: string;
  value: number;
  unit: string;
  timestamp: number;
  status: "normal" | "warning" | "critical";
}

interface SensorTelemetryPanelProps {
  robotId: string;
  robotName: string;
  readings: SensorReading[];
  onRefresh?: () => void;
  className?: string;
}

export const SensorTelemetryPanel: React.FC<SensorTelemetryPanelProps> = ({
  robotId,
  robotName,
  readings,
  onRefresh,
  className,
}) => {
  const [expandedId, setExpandedId] = useState<string | null>(null);

  const getStatusColor = (status: SensorReading["status"]) => {
    switch (status) {
      case "normal":
        return { text: "text-green-400", bg: "bg-green-400/20" };
      case "warning":
        return { text: "text-amber-400", bg: "bg-amber-400/20" };
      case "critical":
        return { text: "text-red-400", bg: "bg-red-400/20" };
    }
  };

  const getSensorIcon = (name: string): LucideIcon => {
    if (name.toLowerCase().includes("temp")) return Thermometer as LucideIcon;
    if (name.toLowerCase().includes("pressure")) return Gauge as LucideIcon;
    if (name.toLowerCase().includes("humidity")) return Droplet as LucideIcon;
    if (
      name.toLowerCase().includes("gps") ||
      name.toLowerCase().includes("position")
    )
      return MapPin;
    if (
      name.toLowerCase().includes("imu") ||
      name.toLowerCase().includes("acceler")
    )
      return Activity;
    if (
      name.toLowerCase().includes("ultrasonic") ||
      name.toLowerCase().includes("distance")
    )
      return Radio;
    if (
      name.toLowerCase().includes("camera") ||
      name.toLowerCase().includes("vision")
    )
      return Camera as LucideIcon;
    return Sensor as LucideIcon;
  };

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Radio className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">
              Sensor Telemetry
            </h3>
          </div>
          <button
            onClick={onRefresh}
            className="p-1.5 hover:bg-aviation-bg-instrument rounded transition-colors"
          >
            <RefreshCw className="w-4 h-4 text-aviation-text-muted" />
          </button>
        </div>
        <div className="flex items-center gap-2 mt-2 text-xs text-aviation-text-dim">
          <Bot className="w-3 h-3" />
          <span>{robotName}</span>
          <span className="text-aviation-text-muted">•</span>
          <span className="font-mono">{robotId.slice(0, 8)}</span>
        </div>
      </div>

      {/* Sensor List */}
      <div className="flex-1 overflow-y-auto">
        {readings.map((reading) => {
          const colors = getStatusColor(reading.status);
          const Icon = getSensorIcon(reading.name);
          const isExpanded = expandedId === reading.id;

          return (
            <div
              key={reading.id}
              onClick={() => setExpandedId(isExpanded ? null : reading.id)}
              className="p-4 border-b border-aviation-border-panel cursor-pointer hover:bg-aviation-bg-secondary transition-colors"
            >
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <Icon className={cn("w-4 h-4", colors.text)} />
                  <span className="text-sm text-aviation-text-primary">
                    {reading.name}
                  </span>
                </div>
                <div
                  className={cn(
                    "flex items-center gap-1.5 px-2 py-1 rounded",
                    colors.bg,
                  )}
                >
                  <span className={cn("text-sm font-bold", colors.text)}>
                    {reading.value.toFixed(1)}
                  </span>
                  <span className="text-[10px] text-aviation-text-dim">
                    {reading.unit}
                  </span>
                </div>
              </div>
              <div className="flex items-center justify-between text-[10px] text-aviation-text-dim">
                <span>{new Date(reading.timestamp).toLocaleTimeString()}</span>
                <span className={cn("uppercase", colors.text)}>
                  {reading.status}
                </span>
              </div>
            </div>
          );
        })}
      </div>

      {/* Summary */}
      <div className="px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="grid grid-cols-3 gap-3 text-xs">
          <div className="flex items-center gap-1.5">
            <div className="w-2 h-2 rounded-full bg-green-400" />
            <span className="text-aviation-text-dim">
              {readings.filter((r) => r.status === "normal").length} Normal
            </span>
          </div>
          <div className="flex items-center gap-1.5">
            <div className="w-2 h-2 rounded-full bg-amber-400" />
            <span className="text-aviation-text-dim">
              {readings.filter((r) => r.status === "warning").length} Warning
            </span>
          </div>
          <div className="flex items-center gap-1.5">
            <div className="w-2 h-2 rounded-full bg-red-400" />
            <span className="text-aviation-text-dim">
              {readings.filter((r) => r.status === "critical").length} Critical
            </span>
          </div>
        </div>
      </div>
    </div>
  );
};

// Helper component for Waves icon - typed as LucideIcon for compatibility
const Waves: React.FC<{ className?: string }> & { displayName?: string } = ({
  className,
}) => (
  <svg
    className={className}
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
  >
    <path d="M2 12c.6-.5 1.3-.8 2-.8s1.4.3 2 .8c1.2 1 2.4 2 4 2 1.3 0 2.5-.5 3.5-1.3" />
    <path d="M2 6c.6-.5 1.3-.8 2-.8s1.4.3 2 .8c1.2 1 2.4 2 4 2 1.3 0 2.5-.5 3.5-1.3" />
    <path d="M2 18c.6-.5 1.3-.8 2-.8s1.4.3 2 .8c1.2 1 2.4 2 4 2 1.3 0 2.5-.5 3.5-1.3" />
  </svg>
);

const Sensor: React.FC<{ className?: string }> & { displayName?: string } = ({
  className,
}) => (
  <svg
    className={className}
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
  >
    <circle cx="12" cy="12" r="3" />
    <path d="M12 2v4M12 18v4M2 12h4M18 12h4" />
  </svg>
);

const Thermometer: React.FC<{ className?: string }> & {
  displayName?: string;
} = ({ className }) => (
  <svg
    className={className}
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
  >
    <path d="M14 14.76V3.5a2.5 2.5 0 0 0-5 0v11.26a4.5 4.5 0 1 0 5 0z" />
  </svg>
);

const Droplet: React.FC<{ className?: string }> & { displayName?: string } = ({
  className,
}) => (
  <svg
    className={className}
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
  >
    <path d="M12 2.69l5.66 5.66a8 8 0 1 1-11.31 0z" />
  </svg>
);

// ============================================================================
// Robot Command Center
// ============================================================================

interface Command {
  id: string;
  name: string;
  description: string;
  type: "move" | "stop" | "patrol" | "return" | "custom";
  status: "pending" | "sent" | "acknowledged" | "completed" | "failed";
  issuedAt: number;
  acknowledgedAt?: number;
  completedAt?: number;
}

interface RobotCommandCenterProps {
  robotId: string;
  robotName: string;
  commands: Command[];
  onSendCommand?: (command: Command) => void;
  onCancelCommand?: (commandId: string) => void;
  className?: string;
}

export const RobotCommandCenter: React.FC<RobotCommandCenterProps> = ({
  robotId,
  robotName,
  commands,
  onSendCommand,
  onCancelCommand,
  className,
}) => {
  const [selectedCommand, setSelectedCommand] = useState<string | null>(null);

  const getStatusColor = (status: Command["status"]) => {
    switch (status) {
      case "pending":
        return "text-amber-400";
      case "sent":
        return "text-aviation-cyan";
      case "acknowledged":
        return "text-purple-400";
      case "completed":
        return "text-green-400";
      case "failed":
        return "text-red-400";
    }
  };

  const getCommandIcon = (type: Command["type"]) => {
    switch (type) {
      case "move":
        return Navigation;
      case "stop":
        return Square;
      case "patrol":
        return Waypoints;
      case "return":
        return ArrowUp;
      case "custom":
        return Zap;
    }
  };

  const quickCommands: Command[] = [
    {
      id: "stop-all",
      name: "Emergency Stop",
      description: "Immediately halt all operations",
      type: "stop",
      status: "pending",
      issuedAt: Date.now(),
    },
    {
      id: "return-base",
      name: "Return to Base",
      description: "Navigate back to charging station",
      type: "return",
      status: "pending",
      issuedAt: Date.now(),
    },
    {
      id: "patrol",
      name: "Start Patrol",
      description: "Begin scheduled patrol route",
      type: "patrol",
      status: "pending",
      issuedAt: Date.now(),
    },
  ];

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Crosshair className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">
              Command Center
            </h3>
          </div>
          <div className="flex items-center gap-2 px-2 py-1 bg-aviation-bg-instrument rounded">
            <span className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />
            <span className="text-xs text-aviation-text-primary">
              Connected
            </span>
          </div>
        </div>
        <div className="flex items-center gap-2 mt-2 text-xs text-aviation-text-dim">
          <Bot className="w-3 h-3" />
          <span>{robotName}</span>
        </div>
      </div>

      {/* Quick Commands */}
      <div className="px-4 py-3 border-b border-aviation-border-panel bg-aviation-bg-secondary">
        <span className="text-xs text-aviation-text-dim mb-2 block">
          Quick Commands
        </span>
        <div className="flex flex-wrap gap-2">
          {quickCommands.map((cmd) => {
            const Icon = getCommandIcon(cmd.type);
            return (
              <button
                key={cmd.id}
                onClick={() => onSendCommand?.(cmd)}
                className={cn(
                  "flex items-center gap-2 px-3 py-2 rounded-lg border border-aviation-border-panel transition-colors",
                  cmd.type === "stop"
                    ? "bg-red-500/20 hover:bg-red-500/30 border-red-500/50"
                    : "bg-aviation-bg-instrument hover:bg-aviation-bg-panel",
                )}
              >
                <Icon
                  className={cn(
                    "w-4 h-4",
                    cmd.type === "stop" ? "text-red-400" : "text-aviation-cyan",
                  )}
                />
                <span
                  className={cn(
                    "text-xs",
                    cmd.type === "stop"
                      ? "text-red-400"
                      : "text-aviation-text-primary",
                  )}
                >
                  {cmd.name}
                </span>
              </button>
            );
          })}
        </div>
      </div>

      {/* Command History */}
      <div className="flex-1 overflow-y-auto">
        <div className="px-4 py-2 border-b border-aviation-border-panel bg-aviation-bg-secondary">
          <span className="text-xs text-aviation-text-dim">
            Command History
          </span>
        </div>
        {commands.map((cmd) => {
          const Icon = getCommandIcon(cmd.type);
          const isSelected = selectedCommand === cmd.id;

          return (
            <div
              key={cmd.id}
              onClick={() => setSelectedCommand(isSelected ? null : cmd.id)}
              className={cn(
                "p-4 border-b border-aviation-border-panel cursor-pointer transition-colors",
                isSelected
                  ? "bg-aviation-bg-instrument"
                  : "hover:bg-aviation-bg-secondary",
              )}
            >
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <Icon className="w-4 h-4 text-aviation-cyan" />
                  <span className="text-sm font-medium text-aviation-text-primary">
                    {cmd.name}
                  </span>
                </div>
                <span
                  className={cn(
                    "text-xs font-medium uppercase",
                    getStatusColor(cmd.status),
                  )}
                >
                  {cmd.status}
                </span>
              </div>
              <p className="text-xs text-aviation-text-dim mb-2">
                {cmd.description}
              </p>
              <div className="flex items-center justify-between text-[10px] text-aviation-text-dim">
                <span>
                  Issued: {new Date(cmd.issuedAt).toLocaleTimeString()}
                </span>
                {cmd.completedAt && (
                  <span>
                    Completed: {new Date(cmd.completedAt).toLocaleTimeString()}
                  </span>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

const Crosshair: React.FC<{ className?: string }> = ({ className }) => (
  <svg
    className={className}
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
  >
    <circle cx="12" cy="12" r="10" />
    <line x1="22" y1="12" x2="18" y2="12" />
    <line x1="6" y1="12" x2="2" y2="12" />
    <line x1="12" y1="6" x2="12" y2="2" />
    <line x1="12" y1="22" x2="12" y2="18" />
  </svg>
);

// ============================================================================
// Physical Environment Map
// ============================================================================

interface MapWaypoint {
  id: string;
  name: string;
  x: number;
  y: number;
  type: "charging" | "checkpoint" | "target" | "hazard";
}

interface Obstacle {
  id: string;
  type: "static" | "dynamic";
  position: { x: number; y: number };
  dimensions: { width: number; height: number };
  label?: string;
}

interface PhysicalEnvironmentMapProps {
  robots: Robot[];
  waypoints: MapWaypoint[];
  obstacles: Obstacle[];
  selectedRobotId?: string | null;
  onRobotSelect?: (robot: Robot) => void;
  onWaypointClick?: (waypoint: MapWaypoint) => void;
  className?: string;
}

export const PhysicalEnvironmentMap: React.FC<PhysicalEnvironmentMapProps> = ({
  robots,
  waypoints,
  obstacles,
  selectedRobotId,
  onRobotSelect,
  onWaypointClick,
  className,
}) => {
  const [hoveredRobot, setHoveredRobot] = useState<string | null>(null);

  const getWaypointColor = (type: MapWaypoint["type"]) => {
    switch (type) {
      case "charging":
        return "#22c55e";
      case "checkpoint":
        return "#3b82f6";
      case "target":
        return "#f59e0b";
      case "hazard":
        return "#ef4444";
    }
  };

  const getWaypointIcon = (type: MapWaypoint["type"]): LucideIcon => {
    switch (type) {
      case "charging":
        return BatteryCharging;
      case "checkpoint":
        return CheckCircle;
      case "target":
        return Target;
      case "hazard":
        return AlertTriangle;
    }
  };

  const getStatusColor = (status: RobotStatus) => {
    switch (status) {
      case "online":
        return "text-green-400";
      case "busy":
        return "text-aviation-cyan";
      case "offline":
        return "text-aviation-text-muted";
      case "error":
        return "text-red-400";
      case "maintenance":
        return "text-amber-400";
      default:
        return "text-aviation-text-muted";
    }
  };

  return (
    <div
      className={cn(
        "relative h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      {/* Header */}
      <div className="absolute top-3 left-3 flex items-center gap-2 z-10">
        <div className="flex items-center gap-1.5 px-2 py-1 bg-aviation-bg-instrument rounded border border-aviation-border-panel">
          <Map className="w-4 h-4 text-aviation-cyan" />
          <span className="text-xs text-aviation-text-primary font-medium">
            Environment Map
          </span>
        </div>
      </div>

      {/* Legend */}
      <div className="absolute top-3 right-3 flex items-center gap-3 px-2 py-1 bg-aviation-bg-instrument rounded border border-aviation-border-panel z-10">
        {[
          { type: "charging", color: "#22c55e" },
          { type: "checkpoint", color: "#3b82f6" },
          { type: "target", color: "#f59e0b" },
          { type: "hazard", color: "#ef4444" },
        ].map((item) => (
          <div key={item.type} className="flex items-center gap-1">
            <div
              className="w-2 h-2 rounded"
              style={{ backgroundColor: item.color }}
            />
            <span className="text-[10px] text-aviation-text-dim capitalize">
              {item.type}
            </span>
          </div>
        ))}
      </div>

      {/* Map SVG */}
      <svg className="w-full h-full" viewBox="0 0 400 300">
        {/* Grid pattern */}
        <defs>
          <pattern
            id="grid"
            width="20"
            height="20"
            patternUnits="userSpaceOnUse"
          >
            <path
              d="M 20 0 L 0 0 0 20"
              fill="none"
              stroke="rgba(255,255,255,0.05)"
              strokeWidth="0.5"
            />
          </pattern>
        </defs>
        <rect width="100%" height="100%" fill="url(#grid)" />

        {/* Obstacles */}
        {obstacles.map((obs) => (
          <g key={obs.id}>
            <rect
              x={obs.position.x - obs.dimensions.width / 2}
              y={obs.position.y - obs.dimensions.height / 2}
              width={obs.dimensions.width}
              height={obs.dimensions.height}
              className={cn(
                "fill-aviation-bg-instrument stroke-2",
                obs.type === "dynamic"
                  ? "stroke-aviation-amber"
                  : "stroke-aviation-border-panel",
              )}
              strokeDasharray={obs.type === "dynamic" ? "4 2" : undefined}
            />
            {obs.label && (
              <text
                x={obs.position.x}
                y={obs.position.y + obs.dimensions.height / 2 + 12}
                textAnchor="middle"
                className="text-[8px] fill-aviation-text-dim"
              >
                {obs.label}
              </text>
            )}
          </g>
        ))}

        {/* Waypoints */}
        {waypoints.map((wp) => {
          const Icon = getWaypointIcon(wp.type);
          return (
            <g
              key={wp.id}
              onClick={() => onWaypointClick?.(wp)}
              className="cursor-pointer"
            >
              <circle
                cx={wp.x}
                cy={wp.y}
                r="12"
                fill={getWaypointColor(wp.type)}
                fillOpacity="0.2"
                stroke={getWaypointColor(wp.type)}
                strokeWidth="2"
              />
              <text
                x={wp.x}
                y={wp.y + 4}
                textAnchor="middle"
                className="text-[8px] fill-aviation-text-primary font-medium"
              >
                {wp.name.slice(0, 3)}
              </text>
            </g>
          );
        })}

        {/* Robot positions */}
        {robots.map((robot) => {
          const isSelected = selectedRobotId === robot.id;
          const isHovered = hoveredRobot === robot.id;
          const getRobotColor = (status: RobotStatus) => {
            switch (status) {
              case "online":
                return "#22c55e";
              case "busy":
                return "#06b6d4";
              case "offline":
                return "#6b7280";
              case "error":
                return "#ef4444";
              case "maintenance":
                return "#f59e0b";
            }
          };

          return (
            <g
              key={robot.id}
              onClick={() => onRobotSelect?.(robot)}
              onMouseEnter={() => setHoveredRobot(robot.id)}
              onMouseLeave={() => setHoveredRobot(null)}
              className="cursor-pointer"
            >
              {/* Selection ring */}
              {(isSelected || isHovered) && (
                <circle
                  cx={robot.position.x}
                  cy={robot.position.y}
                  r="18"
                  fill="none"
                  stroke={getRobotColor(robot.status)}
                  strokeWidth="2"
                  strokeDasharray="4 2"
                />
              )}

              {/* Robot body */}
              <circle
                cx={robot.position.x}
                cy={robot.position.y}
                r="10"
                fill="rgba(6, 182, 212, 0.3)"
                stroke={getRobotColor(robot.status)}
                strokeWidth={isSelected ? 3 : 2}
              />

              {/* Robot icon */}
              <Bot
                x={robot.position.x - 6}
                y={robot.position.y - 6}
                className="w-3 h-3"
                style={{ color: getRobotColor(robot.status) }}
              />
            </g>
          );
        })}
      </svg>

      {/* Hover tooltip */}
      {hoveredRobot && robots.find((r) => r.id === hoveredRobot) && (
        <div className="absolute bottom-3 left-3 right-3 p-3 bg-aviation-bg-secondary/90 rounded-lg border border-aviation-border-panel backdrop-blur-sm z-10">
          <div className="flex items-center justify-between">
            <div>
              <h4 className="text-sm font-medium text-aviation-text-primary">
                {robots.find((r) => r.id === hoveredRobot)?.name}
              </h4>
              <div className="flex items-center gap-2 mt-1 text-[10px] text-aviation-text-dim">
                <span className="uppercase">
                  {robots.find((r) => r.id === hoveredRobot)?.type}
                </span>
                <span>•</span>
                <span>
                  {robots.find((r) => r.id === hoveredRobot)?.batteryLevel}%
                  battery
                </span>
              </div>
            </div>
            <div
              className={cn(
                "text-sm font-bold uppercase",
                getStatusColor(
                  robots.find((r) => r.id === hoveredRobot)!.status,
                ),
              )}
            >
              {robots.find((r) => r.id === hoveredRobot)?.status}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

// ============================================================================
// Drone Flight Overlay
// ============================================================================

interface FlightPath {
  id: string;
  name: string;
  points: Array<{ x: number; y: number; altitude: number }>;
  status: "planned" | "active" | "completed";
  estimatedDuration: number;
  totalDistance: number;
}

interface DroneFlightOverlayProps {
  droneId: string;
  droneName: string;
  currentPosition: { x: number; y: number; altitude: number };
  flightPaths: FlightPath[];
  activePathId?: string | null;
  onPathSelect?: (path: FlightPath) => void;
  className?: string;
}

export const DroneFlightOverlay: React.FC<DroneFlightOverlayProps> = ({
  droneId,
  droneName,
  currentPosition,
  flightPaths,
  activePathId,
  onPathSelect,
  className,
}) => {
  const getStatusColor = (status: FlightPath["status"]) => {
    switch (status) {
      case "active":
        return "text-green-400";
      case "planned":
        return "text-aviation-cyan";
      case "completed":
        return "text-aviation-text-muted";
    }
  };

  return (
    <div
      className={cn(
        "relative h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      {/* Header */}
      <div className="absolute top-3 left-3 flex items-center gap-2 z-10">
        <div className="flex items-center gap-1.5 px-2 py-1 bg-aviation-bg-instrument rounded border border-aviation-border-panel">
          <Navigation className="w-4 h-4 text-aviation-cyan" />
          <span className="text-xs text-aviation-text-primary font-medium">
            Flight Overlay
          </span>
        </div>
        <div className="flex items-center gap-1.5 px-2 py-1 bg-green-500/20 rounded border border-green-500/50">
          <div className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />
          <span className="text-xs text-green-400">LIVE</span>
        </div>
      </div>

      {/* Drone Info */}
      <div className="absolute top-3 right-3 px-2 py-1 bg-aviation-bg-instrument rounded border border-aviation-border-panel z-10">
        <div className="text-xs text-aviation-text-dim">Altitude</div>
        <div className="text-sm font-bold text-aviation-cyan">
          {currentPosition.altitude}m
        </div>
      </div>

      {/* Flight Map SVG */}
      <svg className="w-full h-full" viewBox="0 0 400 300">
        {/* Grid */}
        <defs>
          <pattern
            id="flightGrid"
            width="20"
            height="20"
            patternUnits="userSpaceOnUse"
          >
            <path
              d="M 20 0 L 0 0 0 20"
              fill="none"
              stroke="rgba(255,255,255,0.03)"
              strokeWidth="0.5"
            />
          </pattern>
          <linearGradient id="pathGradient" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stopColor="#06b6d4" stopOpacity="0.3" />
            <stop offset="100%" stopColor="#06b6d4" stopOpacity="1" />
          </linearGradient>
        </defs>
        <rect width="100%" height="100%" fill="url(#flightGrid)" />

        {/* Altitude grid lines */}
        {[0, 50, 100, 150, 200].map((alt) => (
          <g key={alt}>
            <line
              x1="0"
              y1={300 - (alt / 200) * 300}
              x2="400"
              y2={300 - (alt / 200) * 300}
              stroke="rgba(6, 182, 212, 0.1)"
              strokeWidth="1"
              strokeDasharray="4 4"
            />
            <text
              x="395"
              y={300 - (alt / 200) * 300 - 2}
              textAnchor="end"
              className="text-[8px] fill-aviation-text-dim"
            >
              {alt}m
            </text>
          </g>
        ))}

        {/* Flight paths */}
        {flightPaths.map((path) => {
          const isActive = activePathId === path.id;
          const pathPoints = path.points
            .map((p) => `${p.x},${300 - (p.altitude / 200) * 300}`)
            .join(" ");

          return (
            <g
              key={path.id}
              onClick={() => onPathSelect?.(path)}
              className="cursor-pointer"
            >
              {/* Path area fill */}
              <polygon
                points={`0,300 ${pathPoints} 400,300`}
                fill={
                  isActive
                    ? "rgba(6, 182, 212, 0.1)"
                    : "rgba(107, 114, 128, 0.05)"
                }
              />

              {/* Path line */}
              <polyline
                points={pathPoints}
                fill="none"
                stroke={isActive ? "#06b6d4" : "#6b7280"}
                strokeWidth={isActive ? 3 : 2}
                strokeDasharray={path.status === "planned" ? "6 3" : undefined}
              />

              {/* Waypoints */}
              {path.points.map((point, i) => (
                <g key={i}>
                  <circle
                    cx={point.x}
                    cy={300 - (point.altitude / 200) * 300}
                    r={i === 0 || i === path.points.length - 1 ? 5 : 3}
                    fill={isActive ? "#06b6d4" : "#6b7280"}
                  />
                  {isActive && (
                    <text
                      x={point.x}
                      y={300 - (point.altitude / 200) * 300 - 8}
                      textAnchor="middle"
                      className="text-[8px] fill-aviation-text-dim"
                    >
                      {point.altitude}m
                    </text>
                  )}
                </g>
              ))}
            </g>
          );
        })}

        {/* Current drone position */}
        <g className="animate-pulse">
          <circle
            cx={currentPosition.x}
            cy={300 - (currentPosition.altitude / 200) * 300}
            r="8"
            fill="rgba(34, 197, 94, 0.3)"
            stroke="#22c55e"
            strokeWidth="2"
          />
          <circle
            cx={currentPosition.x}
            cy={300 - (currentPosition.altitude / 200) * 300}
            r="3"
            fill="#22c55e"
          />
        </g>
      </svg>

      {/* Path List */}
      <div className="absolute bottom-3 left-3 right-3 p-3 bg-aviation-bg-secondary/90 rounded-lg border border-aviation-border-panel backdrop-blur-sm z-10">
        <div className="flex items-center gap-4 text-xs">
          {flightPaths.map((path) => (
            <div
              key={path.id}
              onClick={() => onPathSelect?.(path)}
              className={cn(
                "flex items-center gap-2 px-2 py-1 rounded cursor-pointer",
                activePathId === path.id
                  ? "bg-aviation-cyan/20"
                  : "hover:bg-aviation-bg-instrument",
              )}
            >
              <Navigation
                className={cn("w-3 h-3", getStatusColor(path.status))}
              />
              <span className="text-aviation-text-primary">{path.name}</span>
              <span
                className={cn(
                  "text-[10px] uppercase",
                  getStatusColor(path.status),
                )}
              >
                {path.status}
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Robot Vision Stream
// ============================================================================

interface VisionFrame {
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

interface RobotVisionStreamProps {
  robotId: string;
  robotName: string;
  frames: VisionFrame[];
  isStreaming?: boolean;
  onFrameSelect?: (frame: VisionFrame) => void;
  className?: string;
}

export const RobotVisionStream: React.FC<RobotVisionStreamProps> = ({
  robotId,
  robotName,
  frames,
  isStreaming = false,
  onFrameSelect,
  className,
}) => {
  const [selectedFrame, setSelectedFrame] = useState<string | null>(null);

  const latestFrame = frames[frames.length - 1];

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Eye className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">
              Vision Stream
            </h3>
          </div>
          <div className="flex items-center gap-2">
            {isStreaming ? (
              <div className="flex items-center gap-1.5 px-2 py-1 bg-red-500/20 rounded border border-red-500/50">
                <div className="w-2 h-2 rounded-full bg-red-500 animate-pulse" />
                <span className="text-xs text-red-400">Recording</span>
              </div>
            ) : (
              <div className="flex items-center gap-1.5 px-2 py-1 bg-aviation-bg-instrument rounded">
                <span className="w-2 h-2 rounded-full bg-aviation-text-muted" />
                <span className="text-xs text-aviation-text-dim">Idle</span>
              </div>
            )}
          </div>
        </div>
        <div className="flex items-center gap-2 mt-2 text-xs text-aviation-text-dim">
          <Camera className="w-3 h-3" />
          <span>{robotName}</span>
          {latestFrame && <span className="text-aviation-text-muted">•</span>}
          {latestFrame && (
            <span>
              {latestFrame.width}x{latestFrame.height}
            </span>
          )}
        </div>
      </div>

      {/* Vision Canvas */}
      <div className="flex-1 relative bg-black">
        {latestFrame ? (
          <svg
            className="w-full h-full"
            viewBox={`0 0 ${latestFrame.width} ${latestFrame.height}`}
            preserveAspectRatio="xMidYMid meet"
          >
            {/* Simulated camera feed background */}
            <rect width="100%" height="100%" fill="#1a1a2e" />

            {/* Grid overlay */}
            {[...Array(10)].map((_, i) => (
              <g key={i}>
                <line
                  x1={(i * latestFrame.width) / 10}
                  y1="0"
                  x2={(i * latestFrame.width) / 10}
                  y2={latestFrame.height}
                  stroke="rgba(6, 182, 212, 0.1)"
                  strokeWidth="1"
                />
                <line
                  x1="0"
                  y1={(i * latestFrame.height) / 10}
                  x2={latestFrame.width}
                  y2={(i * latestFrame.height) / 10}
                  stroke="rgba(6, 182, 212, 0.1)"
                  strokeWidth="1"
                />
              </g>
            ))}

            {/* Detected objects */}
            {latestFrame.objects.map((obj, i) => {
              const bx = obj.boundingBox.x * latestFrame.width;
              const by = obj.boundingBox.y * latestFrame.height;
              const bw = obj.boundingBox.width * latestFrame.width;
              const bh = obj.boundingBox.height * latestFrame.height;

              const color =
                obj.confidence > 0.8
                  ? "#22c55e"
                  : obj.confidence > 0.5
                    ? "#f59e0b"
                    : "#ef4444";

              return (
                <g key={i}>
                  <rect
                    x={bx}
                    y={by}
                    width={bw}
                    height={bh}
                    fill="none"
                    stroke={color}
                    strokeWidth="2"
                  />
                  <rect
                    x={bx}
                    y={by}
                    width={bw}
                    height={bh}
                    fill={color}
                    fillOpacity="0.1"
                  />
                  <text
                    x={bx + 4}
                    y={by + 14}
                    className="text-[10px] fill-white font-medium"
                    style={{ textShadow: "0 1px 2px rgba(0,0,0,0.8)" }}
                  >
                    {obj.label} {Math.round(obj.confidence * 100)}%
                  </text>
                </g>
              );
            })}
          </svg>
        ) : (
          <div className="flex items-center justify-center h-full">
            <div className="text-center text-aviation-text-muted">
              <Camera className="w-12 h-12 mx-auto mb-2 opacity-50" />
              <p className="text-sm">No vision frames available</p>
            </div>
          </div>
        )}

        {/* Timestamp overlay */}
        {latestFrame && (
          <div className="absolute bottom-3 left-3 px-2 py-1 bg-black/60 rounded backdrop-blur-sm">
            <span className="text-xs text-aviation-text-dim font-mono">
              {new Date(latestFrame.timestamp).toLocaleTimeString()}
            </span>
          </div>
        )}

        {/* Object count */}
        {latestFrame && latestFrame.objects.length > 0 && (
          <div className="absolute top-3 right-3 px-2 py-1 bg-black/60 rounded backdrop-blur-sm">
            <span className="text-xs text-aviation-cyan">
              {latestFrame.objects.length} objects detected
            </span>
          </div>
        )}
      </div>

      {/* Frame thumbnails */}
      <div className="px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="flex items-center gap-2 text-xs text-aviation-text-dim mb-2">
          <span>Recent Frames</span>
          <span className="text-aviation-text-muted">({frames.length})</span>
        </div>
        <div className="flex gap-2 overflow-x-auto">
          {frames.slice(-6).map((frame) => (
            <div
              key={frame.id}
              onClick={() => onFrameSelect?.(frame)}
              className={cn(
                "flex-shrink-0 w-16 h-12 rounded border cursor-pointer transition-colors overflow-hidden",
                selectedFrame === frame.id
                  ? "border-aviation-cyan"
                  : "border-aviation-border-panel hover:border-aviation-text-muted",
              )}
            >
              <svg
                viewBox={`0 0 ${frame.width} ${frame.height}`}
                className="w-full h-full"
              >
                <rect width="100%" height="100%" fill="#1a1a2e" />
                {frame.objects.slice(0, 2).map((obj, i) => (
                  <rect
                    key={i}
                    x={obj.boundingBox.x * frame.width}
                    y={obj.boundingBox.y * frame.height}
                    width={obj.boundingBox.width * frame.width}
                    height={obj.boundingBox.height * frame.height}
                    fill="none"
                    stroke="#22c55e"
                    strokeWidth="1"
                  />
                ))}
              </svg>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Device Mesh Viewer
// ============================================================================

interface MeshNode {
  id: string;
  name: string;
  type: "robot" | "sensor" | "gateway" | "controller";
  status: RobotStatus;
  connections: string[];
  metrics?: {
    cpu?: number;
    memory?: number;
    latency?: number;
  };
}

interface DeviceMeshViewerProps {
  nodes: MeshNode[];
  selectedNodeId?: string | null;
  onNodeSelect?: (node: MeshNode) => void;
  onNodeHover?: (node: MeshNode | null) => void;
  className?: string;
}

export const DeviceMeshViewer: React.FC<DeviceMeshViewerProps> = ({
  nodes,
  selectedNodeId,
  onNodeSelect,
  onNodeHover,
  className,
}) => {
  const [hoveredNode, setHoveredNode] = useState<string | null>(null);

  const getNodeColor = (status: RobotStatus) => {
    switch (status) {
      case "online":
        return "#22c55e";
      case "busy":
        return "#06b6d4";
      case "offline":
        return "#6b7280";
      case "error":
        return "#ef4444";
      case "maintenance":
        return "#f59e0b";
    }
  };

  const getNodeIcon = (type: MeshNode["type"]): LucideIcon => {
    switch (type) {
      case "robot":
        return Bot;
      case "sensor":
        return Radio;
      case "gateway":
        return Network;
      case "controller":
        return Cpu;
    }
  };

  const nodePositions = useMemo(() => {
    const positions: Record<string, { x: number; y: number }> = {};
    const centerX = 200;
    const centerY = 150;

    nodes.forEach((node, index) => {
      const angle = (2 * Math.PI * index) / nodes.length;
      const radius = 70 + (nodes.length > 6 ? 20 : 0);
      positions[node.id] = {
        x: centerX + radius * Math.cos(angle),
        y: centerY + radius * Math.sin(angle),
      };
    });
    return positions;
  }, [nodes]);

  const connections = useMemo(() => {
    const conns: Array<{ from: string; to: string }> = [];
    nodes.forEach((node) => {
      node.connections.forEach((targetId) => {
        if (nodes.find((n) => n.id === targetId)) {
          conns.push({ from: node.id, to: targetId });
        }
      });
    });
    return conns;
  }, [nodes]);

  return (
    <div
      className={cn(
        "relative h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      {/* Header */}
      <div className="absolute top-3 left-3 flex items-center gap-2 z-10">
        <div className="flex items-center gap-1.5 px-2 py-1 bg-aviation-bg-instrument rounded border border-aviation-border-panel">
          <Network className="w-4 h-4 text-aviation-cyan" />
          <span className="text-xs text-aviation-text-primary font-medium">
            Device Mesh
          </span>
        </div>
      </div>

      {/* Legend */}
      <div className="absolute top-3 right-3 flex items-center gap-2 px-2 py-1 bg-aviation-bg-instrument rounded border border-aviation-border-panel z-10 text-[10px]">
        {[
          { type: "robot", icon: Bot },
          { type: "sensor", icon: Radio },
          { type: "gateway", icon: Network },
          { type: "controller", icon: Cpu },
        ].map((item) => (
          <div key={item.type} className="flex items-center gap-1">
            <item.icon className="w-3 h-3 text-aviation-text-dim" />
            <span className="text-aviation-text-dim capitalize">
              {item.type}
            </span>
          </div>
        ))}
      </div>

      {/* Mesh SVG */}
      <svg className="w-full h-full" viewBox="0 0 400 300">
        {/* Connection lines */}
        {connections.map((conn, i) => {
          const fromPos = nodePositions[conn.from];
          const toPos = nodePositions[conn.to];
          if (!fromPos || !toPos) return null;

          return (
            <line
              key={i}
              x1={fromPos.x}
              y1={fromPos.y}
              x2={toPos.x}
              y2={toPos.y}
              stroke="rgba(6, 182, 212, 0.3)"
              strokeWidth="2"
            />
          );
        })}

        {/* Nodes */}
        {nodes.map((node) => {
          const pos = nodePositions[node.id];
          if (!pos) return null;

          const isSelected = selectedNodeId === node.id;
          const isHovered = hoveredNode === node.id;
          const Icon = getNodeIcon(node.type);
          const color = getNodeColor(node.status);

          return (
            <g
              key={node.id}
              onClick={() => onNodeSelect?.(node)}
              onMouseEnter={() => {
                setHoveredNode(node.id);
                onNodeHover?.(node);
              }}
              onMouseLeave={() => {
                setHoveredNode(null);
                onNodeHover?.(null);
              }}
              className="cursor-pointer"
            >
              {/* Selection/hover ring */}
              {(isSelected || isHovered) && (
                <circle
                  cx={pos.x}
                  cy={pos.y}
                  r="25"
                  fill="none"
                  stroke={color}
                  strokeWidth="2"
                  strokeDasharray="4 2"
                />
              )}

              {/* Node body */}
              <circle
                cx={pos.x}
                cy={pos.y}
                r="18"
                fill="rgba(6, 182, 212, 0.1)"
                stroke={color}
                strokeWidth="2"
              />

              {/* Icon */}
              <g transform={`translate(${pos.x - 10}, ${pos.y - 10})`}>
                <Icon className="w-5 h-5" style={{ color }} />
              </g>

              {/* Label */}
              <text
                x={pos.x}
                y={pos.y + 32}
                textAnchor="middle"
                className="text-[9px] fill-aviation-text-primary"
              >
                {node.name.length > 10
                  ? node.name.slice(0, 10) + "..."
                  : node.name}
              </text>

              {/* Metrics badge */}
              {node.metrics && (
                <g transform={`translate(${pos.x + 12}, ${pos.y - 8})`}>
                  {node.metrics.cpu && (
                    <rect
                      x="0"
                      y="0"
                      width="20"
                      height="10"
                      rx="2"
                      fill="rgba(0,0,0,0.6)"
                    />
                  )}
                  {node.metrics.cpu && (
                    <text
                      x="10"
                      y="7"
                      textAnchor="middle"
                      className="text-[7px] fill-aviation-cyan"
                    >
                      {node.metrics.cpu}%
                    </text>
                  )}
                </g>
              )}
            </g>
          );
        })}
      </svg>

      {/* Node details */}
      {hoveredNode && nodes.find((n) => n.id === hoveredNode) && (
        <div className="absolute bottom-3 left-3 right-3 p-3 bg-aviation-bg-secondary/90 rounded-lg border border-aviation-border-panel backdrop-blur-sm z-10">
          <div className="flex items-center justify-between">
            <div>
              <h4 className="text-sm font-medium text-aviation-text-primary">
                {nodes.find((n) => n.id === hoveredNode)?.name}
              </h4>
              <div className="flex items-center gap-2 mt-1 text-[10px] text-aviation-text-dim">
                <span className="uppercase">
                  {nodes.find((n) => n.id === hoveredNode)?.type}
                </span>
                <span>•</span>
                <span>
                  {nodes.find((n) => n.id === hoveredNode)?.connections.length}{" "}
                  connections
                </span>
              </div>
            </div>
            <div className="text-right">
              {nodes.find((n) => n.id === hoveredNode)?.metrics && (
                <div className="text-xs text-aviation-text-primary">
                  CPU: {nodes.find((n) => n.id === hoveredNode)?.metrics?.cpu}%
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

// ============================================================================
// Actuator Control Panel
// ============================================================================

interface Actuator {
  id: string;
  name: string;
  type: "servo" | "motor" | "pneumatic" | "hydraulic";
  currentPosition: number;
  targetPosition: number;
  speed: number;
  torque?: number;
  status: RobotStatus;
}

interface ActuatorControlPanelProps {
  robotId: string;
  robotName: string;
  actuators: Actuator[];
  onActuatorControl?: (actuator: Actuator, targetPosition: number) => void;
  onEmergencyStop?: () => void;
  className?: string;
}

export const ActuatorControlPanel: React.FC<ActuatorControlPanelProps> = ({
  robotId,
  robotName,
  actuators,
  onActuatorControl,
  onEmergencyStop,
  className,
}) => {
  const [selectedActuator, setSelectedActuator] = useState<string | null>(null);
  const [targetValue, setTargetValue] = useState<number>(0);

  const getTypeIcon = (type: Actuator["type"]): LucideIcon => {
    switch (type) {
      case "servo":
        return Gauge;
      case "motor":
        return RotateCcw;
      case "pneumatic":
        return Wind as LucideIcon;
      case "hydraulic":
        return Droplet as LucideIcon;
    }
  };

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Settings className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">
              Actuator Control
            </h3>
          </div>
          <button
            onClick={onEmergencyStop}
            className="flex items-center gap-1.5 px-3 py-1.5 bg-red-500/20 border border-red-500/50 rounded hover:bg-red-500/30 transition-colors"
          >
            <Square className="w-4 h-4 text-red-400" />
            <span className="text-xs text-red-400 font-medium">E-STOP</span>
          </button>
        </div>
        <div className="flex items-center gap-2 mt-2 text-xs text-aviation-text-dim">
          <Bot className="w-3 h-3" />
          <span>{robotName}</span>
        </div>
      </div>

      {/* Actuator List */}
      <div className="flex-1 overflow-y-auto">
        {actuators.map((actuator) => {
          const Icon = getTypeIcon(actuator.type);
          const isSelected = selectedActuator === actuator.id;
          const isMoving =
            Math.abs(actuator.currentPosition - actuator.targetPosition) > 1;

          return (
            <div
              key={actuator.id}
              onClick={() =>
                setSelectedActuator(isSelected ? null : actuator.id)
              }
              className={cn(
                "p-4 border-b border-aviation-border-panel cursor-pointer transition-colors",
                isSelected
                  ? "bg-aviation-bg-instrument"
                  : "hover:bg-aviation-bg-secondary",
              )}
            >
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2">
                  <Icon className="w-4 h-4 text-aviation-cyan" />
                  <span className="text-sm font-medium text-aviation-text-primary">
                    {actuator.name}
                  </span>
                </div>
                <span
                  className={cn(
                    "px-2 py-0.5 text-[10px] rounded uppercase",
                    isMoving
                      ? "bg-aviation-cyan/20 text-aviation-cyan"
                      : "bg-aviation-bg-instrument text-aviation-text-dim",
                  )}
                >
                  {isMoving ? "Moving" : "Idle"}
                </span>
              </div>

              {/* Position Display */}
              <div className="mb-3">
                <div className="flex items-center justify-between text-xs text-aviation-text-dim mb-1">
                  <span>Position</span>
                  <span>{actuator.currentPosition.toFixed(1)}°</span>
                </div>
                <div className="relative h-2 bg-aviation-bg-instrument rounded-full overflow-hidden">
                  <div
                    className="absolute h-full bg-aviation-cyan rounded-full transition-all"
                    style={{
                      width: `${(actuator.currentPosition / 360) * 100}%`,
                    }}
                  />
                  <div
                    className="absolute h-full bg-aviation-amber rounded-full opacity-50"
                    style={{
                      width: `${(actuator.targetPosition / 360) * 100}%`,
                    }}
                  />
                </div>
                <div className="flex items-center justify-between text-[10px] text-aviation-text-dim mt-1">
                  <span>Target: {actuator.targetPosition.toFixed(1)}°</span>
                  <span>Speed: {actuator.speed}°/s</span>
                </div>
              </div>

              {/* Control Slider */}
              {isSelected && (
                <div className="mt-3 pt-3 border-t border-aviation-border-panel">
                  <div className="flex items-center gap-3">
                    <input
                      type="range"
                      min="0"
                      max="360"
                      value={targetValue}
                      onChange={(e) => setTargetValue(parseInt(e.target.value))}
                      className="flex-1 h-2 bg-aviation-bg-instrument rounded-full appearance-none cursor-pointer [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:w-4 [&::-webkit-slider-thumb]:h-4 [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-aviation-cyan"
                    />
                    <span className="text-sm font-mono text-aviation-text-primary w-12 text-right">
                      {targetValue}°
                    </span>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        onActuatorControl?.(actuator, targetValue);
                      }}
                      className="px-3 py-1.5 bg-aviation-cyan text-aviation-bg-primary rounded text-xs hover:bg-aviation-cyan/90 transition-colors"
                    >
                      Set
                    </button>
                  </div>

                  {actuator.torque && (
                    <div className="mt-2 text-xs text-aviation-text-dim">
                      Torque: {actuator.torque} N·m
                    </div>
                  )}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
};

const Wind: React.FC<{ className?: string }> & { displayName?: string } = ({
  className,
}) => (
  <svg
    className={className}
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
  >
    <path d="M9.59 4.59A2 2 0 1 1 11 8H2m10.59 11.41A2 2 0 1 0 14 16H2m15.73-8.27A2.5 2.5 0 1 1 19.5 12H2" />
  </svg>
);

// ============================================================================
// Edge Device Monitor
// ============================================================================

interface EdgeMetric {
  id: string;
  name: string;
  value: number;
  unit: string;
  trend: "up" | "down" | "stable";
  threshold?: { warning: number; critical: number };
}

interface EdgeDevice {
  id: string;
  name: string;
  type: "compute" | "inference" | "gateway";
  status: RobotStatus;
  ipAddress: string;
  metrics: EdgeMetric[];
  uptime: number;
}

interface EdgeDeviceMonitorProps {
  devices: EdgeDevice[];
  selectedDeviceId?: string | null;
  onDeviceSelect?: (device: EdgeDevice) => void;
  onDeviceRestart?: (deviceId: string) => void;
  className?: string;
}

export const EdgeDeviceMonitor: React.FC<EdgeDeviceMonitorProps> = ({
  devices,
  selectedDeviceId,
  onDeviceSelect,
  onDeviceRestart,
  className,
}) => {
  const getStatusColor = (status: RobotStatus) => {
    switch (status) {
      case "online":
        return "text-green-400";
      case "busy":
        return "text-aviation-cyan";
      case "offline":
        return "text-aviation-text-muted";
      case "error":
        return "text-red-400";
      case "maintenance":
        return "text-amber-400";
    }
  };

  const getTypeIcon = (type: EdgeDevice["type"]): LucideIcon => {
    switch (type) {
      case "compute":
        return Cpu as LucideIcon;
      case "inference":
        return Brain as LucideIcon;
      case "gateway":
        return Server as LucideIcon;
    }
  };

  const formatUptime = (seconds: number) => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    if (days > 0) return `${days}d ${hours}h`;
    return `${hours}h`;
  };

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Server className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">
              Edge Devices
            </h3>
          </div>
          <span className="text-xs text-aviation-text-dim">
            {devices.length} devices
          </span>
        </div>
      </div>

      {/* Device List */}
      <div className="flex-1 overflow-y-auto">
        {devices.map((device) => {
          const Icon = getTypeIcon(device.type);
          const isSelected = selectedDeviceId === device.id;
          const cpuMetric = device.metrics.find((m) =>
            m.name.toLowerCase().includes("cpu"),
          );
          const memMetric = device.metrics.find((m) =>
            m.name.toLowerCase().includes("memory"),
          );

          return (
            <div
              key={device.id}
              onClick={() => onDeviceSelect?.(device)}
              className={cn(
                "p-4 border-b border-aviation-border-panel cursor-pointer transition-colors",
                isSelected
                  ? "bg-aviation-bg-instrument"
                  : "hover:bg-aviation-bg-secondary",
              )}
            >
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2">
                  <Icon
                    className={cn("w-5 h-5", getStatusColor(device.status))}
                  />
                  <div>
                    <div className="text-sm font-medium text-aviation-text-primary">
                      {device.name}
                    </div>
                    <div className="text-[10px] text-aviation-text-dim font-mono">
                      {device.ipAddress}
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <span
                    className={cn(
                      "text-xs uppercase font-medium",
                      getStatusColor(device.status),
                    )}
                  >
                    {device.status}
                  </span>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      onDeviceRestart?.(device.id);
                    }}
                    className="p-1.5 hover:bg-aviation-bg-panel rounded transition-colors"
                  >
                    <RotateCcw className="w-3 h-3 text-aviation-text-muted" />
                  </button>
                </div>
              </div>

              {/* Metrics Grid */}
              <div className="grid grid-cols-4 gap-2">
                {cpuMetric && (
                  <div className="px-2 py-1.5 bg-aviation-bg-instrument rounded">
                    <div className="text-[10px] text-aviation-text-dim">
                      CPU
                    </div>
                    <div
                      className={cn(
                        "text-sm font-bold",
                        cpuMetric.value > 80
                          ? "text-red-400"
                          : "text-aviation-text-primary",
                      )}
                    >
                      {cpuMetric.value.toFixed(0)}%
                    </div>
                  </div>
                )}
                {memMetric && (
                  <div className="px-2 py-1.5 bg-aviation-bg-instrument rounded">
                    <div className="text-[10px] text-aviation-text-dim">
                      Memory
                    </div>
                    <div
                      className={cn(
                        "text-sm font-bold",
                        memMetric.value > 80
                          ? "text-red-400"
                          : "text-aviation-text-primary",
                      )}
                    >
                      {memMetric.value.toFixed(0)}%
                    </div>
                  </div>
                )}
                <div className="px-2 py-1.5 bg-aviation-bg-instrument rounded">
                  <div className="text-[10px] text-aviation-text-dim">Type</div>
                  <div className="text-sm font-bold text-aviation-text-primary capitalize">
                    {device.type.slice(0, 4)}
                  </div>
                </div>
                <div className="px-2 py-1.5 bg-aviation-bg-instrument rounded">
                  <div className="text-[10px] text-aviation-text-dim">
                    Uptime
                  </div>
                  <div className="text-sm font-bold text-aviation-cyan">
                    {formatUptime(device.uptime)}
                  </div>
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

const Brain: React.FC<{ className?: string }> = ({ className }) => (
  <svg
    className={className}
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
  >
    <path d="M12 5a3 3 0 1 0-5.997.125 4 4 0 0 0-2.526 5.77 4 4 0 0 0 .556 6.588A4 4 0 1 0 12 18Z" />
    <path d="M12 5a3 3 0 1 1 5.997.125 4 4 0 0 1 2.526 5.77 4 4 0 0 1-.556 6.588A4 4 0 1 1 12 18Z" />
    <path d="M15 13a4.5 4.5 0 0 1-3-4 4.5 4.5 0 0 1-3 4" />
    <path d="M17.599 6.5a3 3 0 0 0 .399-1.375" />
    <path d="M6.003 5.125A3 3 0 0 0 6.401 6.5" />
    <path d="M3.477 10.896a4 4 0 0 1 .585-.396" />
    <path d="M19.938 10.5a4 4 0 0 1 .585.396" />
    <path d="M6 18a4 4 0 0 1-1.967-.516" />
    <path d="M19.967 17.484A4 4 0 0 1 18 18" />
  </svg>
);

// ============================================================================
// Robotic Workflow Designer
// ============================================================================

interface WorkflowStep {
  id: string;
  name: string;
  type:
    | "navigation"
    | "manipulation"
    | "inspection"
    | "monitoring"
    | "charging";
  duration: number;
  robotType?: RobotType;
  params?: Record<string, unknown>;
}

interface RoboticWorkflow {
  id: string;
  name: string;
  steps: WorkflowStep[];
  totalDuration: number;
  assignedRobots: string[];
  status:
    | "draft"
    | "scheduled"
    | "running"
    | "paused"
    | "completed"
    | "cancelled";
}

interface RoboticWorkflowDesignerProps {
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

export const RoboticWorkflowDesigner: React.FC<
  RoboticWorkflowDesignerProps
> = ({
  workflows,
  selectedWorkflowId,
  selectedStepId,
  onWorkflowSelect,
  onStepSelect,
  onWorkflowCreate,
  onStepAdd,
  onStepRemove,
  className,
}) => {
  const [isAddingStep, setIsAddingStep] = useState(false);

  const getStatusColor = (status: RoboticWorkflow["status"]) => {
    switch (status) {
      case "draft":
        return "text-aviation-text-muted";
      case "scheduled":
        return "text-aviation-cyan";
      case "running":
        return "text-green-400";
      case "paused":
        return "text-amber-400";
      case "completed":
        return "text-purple-400";
      case "cancelled":
        return "text-red-400";
    }
  };

  const getStepIcon = (type: WorkflowStep["type"]): LucideIcon => {
    switch (type) {
      case "navigation":
        return Navigation;
      case "manipulation":
        return Settings;
      case "inspection":
        return Eye;
      case "monitoring":
        return Activity;
      case "charging":
        return BatteryCharging;
    }
  };

  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}m ${secs}s`;
  };

  const selectedWorkflow = workflows.find((w) => w.id === selectedWorkflowId);

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Waypoints className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">
              Workflow Designer
            </h3>
          </div>
          <button
            onClick={onWorkflowCreate}
            className="flex items-center gap-1.5 px-3 py-1.5 bg-aviation-cyan/20 border border-aviation-cyan/50 rounded hover:bg-aviation-cyan/30 transition-colors"
          >
            <Plus className="w-4 h-4 text-aviation-cyan" />
            <span className="text-xs text-aviation-cyan">New Workflow</span>
          </button>
        </div>
      </div>

      {/* Workflow List */}
      <div className="px-4 py-2 border-b border-aviation-border-panel bg-aviation-bg-secondary">
        <span className="text-xs text-aviation-text-dim">Workflows</span>
      </div>
      <div className="max-h-48 overflow-y-auto border-b border-aviation-border-panel">
        {workflows.map((workflow) => {
          const isSelected = selectedWorkflowId === workflow.id;

          return (
            <div
              key={workflow.id}
              onClick={() => onWorkflowSelect?.(workflow)}
              className={cn(
                "px-4 py-3 border-b border-aviation-border-panel cursor-pointer transition-colors",
                isSelected
                  ? "bg-aviation-bg-instrument"
                  : "hover:bg-aviation-bg-secondary",
              )}
            >
              <div className="flex items-center justify-between mb-1">
                <span className="text-sm font-medium text-aviation-text-primary">
                  {workflow.name}
                </span>
                <span
                  className={cn(
                    "text-[10px] uppercase font-medium",
                    getStatusColor(workflow.status),
                  )}
                >
                  {workflow.status}
                </span>
              </div>
              <div className="flex items-center gap-3 text-[10px] text-aviation-text-dim">
                <span>{workflow.steps.length} steps</span>
                <span>•</span>
                <span>{formatDuration(workflow.totalDuration)}</span>
                <span>•</span>
                <span>{workflow.assignedRobots.length} robots</span>
              </div>
            </div>
          );
        })}
      </div>

      {/* Step Editor */}
      {selectedWorkflow ? (
        <div className="flex-1 overflow-y-auto">
          <div className="px-4 py-2 border-b border-aviation-border-panel bg-aviation-bg-secondary flex items-center justify-between">
            <span className="text-xs text-aviation-text-dim">Steps</span>
            <button
              onClick={() => setIsAddingStep(true)}
              className="flex items-center gap-1 px-2 py-1 text-[10px] text-aviation-cyan hover:bg-aviation-bg-instrument rounded transition-colors"
            >
              <Plus className="w-3 h-3" />
              Add Step
            </button>
          </div>

          <div className="p-4">
            {selectedWorkflow.steps.map((step, index) => {
              const Icon = getStepIcon(step.type);
              const isSelected = selectedStepId === step.id;

              return (
                <div key={step.id} className="flex items-center gap-3 mb-3">
                  {/* Step number */}
                  <div className="flex-shrink-0 w-6 h-6 rounded-full bg-aviation-cyan/20 flex items-center justify-center">
                    <span className="text-xs font-bold text-aviation-cyan">
                      {index + 1}
                    </span>
                  </div>

                  {/* Step card */}
                  <div
                    onClick={() => onStepSelect?.(step)}
                    className={cn(
                      "flex-1 p-3 rounded-lg border cursor-pointer transition-colors",
                      isSelected
                        ? "bg-aviation-bg-instrument border-aviation-cyan"
                        : "bg-aviation-bg-secondary border-aviation-border-panel hover:border-aviation-text-muted",
                    )}
                  >
                    <div className="flex items-center gap-2 mb-1">
                      <Icon className="w-4 h-4 text-aviation-cyan" />
                      <span className="text-sm font-medium text-aviation-text-primary">
                        {step.name}
                      </span>
                    </div>
                    <div className="flex items-center gap-2 text-[10px] text-aviation-text-dim">
                      <span className="uppercase">{step.type}</span>
                      <span>•</span>
                      <span>{formatDuration(step.duration)}</span>
                      {step.robotType && (
                        <>
                          <span>•</span>
                          <span className="uppercase">{step.robotType}</span>
                        </>
                      )}
                    </div>
                  </div>

                  {/* Remove button */}
                  <button
                    onClick={() => onStepRemove?.(selectedWorkflow.id, step.id)}
                    className="flex-shrink-0 p-1.5 hover:bg-red-500/20 rounded transition-colors"
                  >
                    <X className="w-4 h-4 text-red-400" />
                  </button>

                  {/* Connector line */}
                  {index < selectedWorkflow.steps.length - 1 && (
                    <div className="absolute left-[52px] top-full w-0.5 h-3 bg-aviation-border-panel" />
                  )}
                </div>
              );
            })}
          </div>
        </div>
      ) : (
        <div className="flex-1 flex items-center justify-center">
          <div className="text-center text-aviation-text-muted">
            <Waypoints className="w-8 h-8 mx-auto mb-2 opacity-50" />
            <p className="text-sm">Select a workflow to edit</p>
          </div>
        </div>
      )}
    </div>
  );
};
