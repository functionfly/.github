import React from "react";
import {
  RobotFleetDashboard,
  SensorTelemetryPanel,
  DeviceMeshViewer,
} from "@functionfly/ui-robotics";
import { Bot, Wifi, AlertTriangle } from "lucide-react";

export function RoboticsPanel() {
  return (
    <div className="p-3 space-y-4">
      <div className="border-b border-border-subtle pb-3">
        <h3 className="text-sm font-medium mb-1">Robotics Control</h3>
        <p className="text-xs text-text-muted">Manage robots, drones, and physical agents</p>
      </div>

      <div className="grid grid-cols-3 gap-2 mb-4">
        <div className="bg-bg-primary rounded-lg border border-border-subtle p-3 text-center">
          <Bot className="size-5 text-brand-400 mx-auto mb-1" />
          <div className="text-lg font-semibold">24</div>
          <div className="text-[10px] text-text-muted">Robots</div>
        </div>
        <div className="bg-bg-primary rounded-lg border border-border-subtle p-3 text-center">
          <Wifi className="size-5 text-success mx-auto mb-1" />
          <div className="text-lg font-semibold">22</div>
          <div className="text-[10px] text-text-muted">Online</div>
        </div>
        <div className="bg-bg-primary rounded-lg border border-border-subtle p-3 text-center">
          <AlertTriangle className="size-5 text-warning mx-auto mb-1" />
          <div className="text-lg font-semibold">2</div>
          <div className="text-[10px] text-text-muted">Alerts</div>
        </div>
      </div>

      <div className="h-80">
        <RobotFleetDashboard
          fleet={{
            id: "fleet-1",
            name: "Warehouse Fleet",
            robots: [
              {
                id: "r1",
                name: "Inspector Bot",
                type: "ground",
                status: "online",
                batteryLevel: 85,
                position: { x: 120, y: 80 },
                lastSeen: Date.now(),
              },
              {
                id: "r2",
                name: "Delivery Drone",
                type: "aerial",
                status: "busy",
                batteryLevel: 62,
                position: { x: 250, y: 150 },
                lastSeen: Date.now() - 30000,
              },
              {
                id: "r3",
                name: "Crawler Unit",
                type: "ground",
                status: "offline",
                batteryLevel: 0,
                position: { x: 80, y: 200 },
                lastSeen: Date.now() - 3600000,
              },
              {
                id: "r4",
                name: "Scanner A1",
                type: "stationary",
                status: "online",
                batteryLevel: 100,
                position: { x: 300, y: 100 },
                lastSeen: Date.now(),
              },
            ],
            totalCount: 4,
            onlineCount: 2,
            busyCount: 1,
            errorCount: 1,
          }}
          onRobotSelect={(robot) => console.log("Selected robot:", robot)}
          onRobotHover={(robot) => console.log("Hover robot:", robot)}
        />
      </div>

      <div className="border-t border-border-subtle pt-4">
        <h4 className="text-xs font-medium mb-2">Sensor Telemetry</h4>
        <SensorTelemetryPanel
          robotId="r1"
          robotName="Inspector Bot"
          readings={[
            {
              id: "s1",
              name: "LIDAR Distance",
              value: 12.5,
              unit: "m",
              timestamp: Date.now(),
              status: "normal",
            },
            {
              id: "s2",
              name: "Ultrasonic",
              value: 0.3,
              unit: "m",
              timestamp: Date.now(),
              status: "normal",
            },
            {
              id: "s3",
              name: "Temperature",
              value: 42.5,
              unit: "°C",
              timestamp: Date.now(),
              status: "warning",
            },
            {
              id: "s4",
              name: "IMU Acceleration",
              value: 9.8,
              unit: "m/s²",
              timestamp: Date.now(),
              status: "normal",
            },
          ]}
          onRefresh={() => console.log("Refresh sensors")}
        />
      </div>

      <div className="border-t border-border-subtle pt-4">
        <h4 className="text-xs font-medium mb-2">Device Mesh</h4>
        <div className="h-64">
          <DeviceMeshViewer
            nodes={[
              {
                id: "r1",
                name: "Inspector Bot",
                type: "robot",
                status: "online",
                connections: ["g1", "s1"],
                metrics: { cpu: 45, memory: 60, latency: 12 },
              },
              {
                id: "r2",
                name: "Delivery Drone",
                type: "robot",
                status: "busy",
                connections: ["g1"],
                metrics: { cpu: 78, memory: 82, latency: 8 },
              },
              {
                id: "g1",
                name: "Edge Gateway",
                type: "gateway",
                status: "online",
                connections: ["r1", "r2", "c1"],
                metrics: { cpu: 32, memory: 45, latency: 2 },
              },
              {
                id: "c1",
                name: "Controller Hub",
                type: "controller",
                status: "online",
                connections: ["g1"],
                metrics: { cpu: 12, memory: 25, latency: 1 },
              },
              {
                id: "s1",
                name: "Temp Sensor",
                type: "sensor",
                status: "online",
                connections: ["r1"],
                metrics: { cpu: 5, latency: 45 },
              },
            ]}
            onNodeSelect={(node) => console.log("Selected mesh node:", node)}
            onNodeHover={(node) => console.log("Hover mesh node:", node)}
          />
        </div>
      </div>
    </div>
  );
}
