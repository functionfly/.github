/**
 * @functionfly/ui-security
 * Security Components - Full implementation
 */

import React, { useState, useCallback, useMemo } from "react";
import { cn } from "@functionfly/ui-core";
import {
  Shield,
  AlertTriangle,
  AlertCircle,
  CheckCircle2,
  XCircle,
  Clock,
  Eye,
  EyeOff,
  Lock,
  Unlock,
  Key,
  User,
  Users,
  FileCode,
  FileText,
  Activity,
  TrendingUp,
  TrendingDown,
  BarChart3,
  PieChart,
  LineChart,
  Globe,
  Server,
  MonitorIcon,
  Bug,
  TrafficCone,
  ChevronRight,
  ChevronDown,
  Search,
  Filter,
  Download,
  Upload,
  Settings,
  RefreshCw,
  Plus,
  Minus,
  Check,
  X,
  Info,
  ArrowRight,
  ArrowUpRight,
  ArrowDownRight,
  Zap,
  Database,
  Network,
  Container,
  Terminal,
  Fingerprint,
  BadgeCheck,
  ShieldCheck,
  ShieldAlert,
  Scan,
  Radar,
  Target,
  Crosshair,
  EyeIcon,
  Microscope,
  Scale,
  Clipboard,
  ClipboardCheck,
  History,
  Timer,
  MapPin,
  KeyRound,
  CreditCard,
  TimerIcon,
  BugIcon,
  Code2,
  Braces,
  Circle,
  Square,
  Hexagon,
  type LucideIcon,
} from "lucide-react";

// ============================================================================
// Permission Matrix
// ============================================================================

interface Permission {
  id: string;
  name: string;
  resource: string;
  action: "read" | "write" | "delete" | "admin";
}

interface Role {
  id: string;
  name: string;
  permissions: string[];
  userCount: number;
  isSystem?: boolean;
}

interface PermissionMatrixProps {
  roles: Role[];
  permissions: Permission[];
  selectedRoleId?: string | null;
  onRoleSelect?: (role: Role) => void;
  onPermissionToggle?: (roleId: string, permissionId: string) => void;
  showUserCount?: boolean;
  className?: string;
}

export const PermissionMatrix: React.FC<PermissionMatrixProps> = ({
  roles,
  permissions,
  selectedRoleId,
  onRoleSelect,
  onPermissionToggle,
  showUserCount = true,
  className,
}) => {
  const selectedRole = roles.find((r) => r.id === selectedRoleId);

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        <Shield className="w-5 h-5 text-aviation-cyan" />
        <span className="text-sm font-medium">Permission Matrix</span>
      </div>
      <div className="flex-1 overflow-auto">
        <table className="w-full text-xs">
          <thead className="sticky top-0 bg-aviation-bg-secondary">
            <tr>
              <th className="px-4 py-2 text-left font-medium text-aviation-text-muted">
                Role
              </th>
              {permissions.slice(0, 8).map((p) => (
                <th
                  key={p.id}
                  className="px-3 py-2 text-center font-medium text-aviation-text-muted truncate max-w-[80px]"
                  title={p.name}
                >
                  {p.name.substring(0, 8)}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {roles.map((role) => (
              <tr
                key={role.id}
                className={cn(
                  "cursor-pointer transition-colors",
                  selectedRoleId === role.id
                    ? "bg-aviation-cyan/10"
                    : "hover:bg-aviation-bg-secondary",
                )}
                onClick={() => onRoleSelect?.(role)}
              >
                <td className="px-4 py-2">
                  <div className="flex items-center gap-2">
                    <BadgeCheck className="w-4 h-4 text-aviation-text-muted" />
                    <div>
                      <div className="font-medium text-aviation-text-primary">
                        {role.name}
                      </div>
                      {showUserCount && (
                        <div className="text-[10px] text-aviation-text-muted">
                          {role.userCount} users
                        </div>
                      )}
                    </div>
                  </div>
                </td>
                {permissions.slice(0, 8).map((p) => {
                  const hasPermission = role.permissions.includes(p.id);
                  return (
                    <td key={p.id} className="px-3 py-2 text-center">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          onPermissionToggle?.(role.id, p.id);
                        }}
                        className={cn(
                          "w-6 h-6 rounded flex items-center justify-center transition-colors",
                          hasPermission
                            ? "bg-green-500/20 text-green-400"
                            : "bg-aviation-bg-instrument text-aviation-text-muted",
                        )}
                      >
                        {hasPermission ? (
                          <Check className="w-4 h-4" />
                        ) : (
                          <X className="w-4 h-4" />
                        )}
                      </button>
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};

// ============================================================================
// Threat Detection Radar
// ============================================================================

interface Threat {
  id: string;
  type:
    | "malware"
    | "intrusion"
    | "dos"
    | "data-breach"
    | "insider"
    | "phishing";
  severity: "low" | "medium" | "high" | "critical";
  source: string;
  target: string;
  timestamp: number;
  status: "detected" | "investigating" | "mitigated" | "false-positive";
}

interface ThreatSector {
  id: string;
  name: string;
  angle: number;
  radius: number;
  severity: "low" | "medium" | "high" | "critical";
  threats: Threat[];
}

interface ThreatDetectionRadarProps {
  sectors: ThreatSector[];
  selectedThreatId?: string | null;
  onThreatSelect?: (threat: Threat) => void;
  showLegend?: boolean;
  className?: string;
}

export const ThreatDetectionRadar: React.FC<ThreatDetectionRadarProps> = ({
  sectors,
  selectedThreatId,
  onThreatSelect,
  showLegend = true,
  className,
}) => {
  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case "critical":
        return "text-red-500";
      case "high":
        return "text-orange-500";
      case "medium":
        return "text-amber-500";
      default:
        return "text-green-500";
    }
  };

  const getSeverityBg = (severity: string) => {
    switch (severity) {
      case "critical":
        return "bg-red-500";
      case "high":
        return "bg-orange-500";
      case "medium":
        return "bg-amber-500";
      default:
        return "bg-green-500";
    }
  };

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      <div className="flex items-center justify-between px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center gap-2">
          <Radar className="w-5 h-5 text-aviation-cyan" />
          <span className="text-sm font-medium">Threat Detection Radar</span>
        </div>
        <div className="flex items-center gap-2 text-xs">
          <span className="text-aviation-text-muted">
            {sectors.reduce((sum, s) => sum + s.threats.length, 0)} threats
          </span>
        </div>
      </div>
      <div className="flex-1 overflow-auto p-4">
        {/* Simplified radar visualization */}
        <div className="relative w-full aspect-square max-w-[400px] mx-auto">
          {/* Radar circles */}
          <div className="absolute inset-0 rounded-full border border-aviation-border-panel" />
          <div className="absolute inset-[15%] rounded-full border border-aviation-border-panel" />
          <div className="absolute inset-[30%] rounded-full border border-aviation-border-panel" />
          <div className="absolute inset-[45%] rounded-full border border-aviation-border-panel" />
          <div className="absolute inset-0 flex items-center justify-center">
            <Shield className="w-12 h-12 text-aviation-cyan" />
          </div>

          {/* Threat dots */}
          {sectors.map((sector, idx) => {
            const angleRad = (sector.angle - 90) * (Math.PI / 180);
            const distance = sector.radius * (1 - 0.2);
            const x = 50 + distance * Math.cos(angleRad);
            const y = 50 + distance * Math.sin(angleRad);

            return sector.threats.slice(0, 3).map((threat, tIdx) => (
              <div
                key={`${threat.id}-${tIdx}`}
                className={cn(
                  "absolute w-3 h-3 rounded-full cursor-pointer transition-transform hover:scale-150",
                  getSeverityBg(threat.severity),
                  selectedThreatId === threat.id && "ring-2 ring-white",
                )}
                style={{
                  left: `${x}%`,
                  top: `${y}%`,
                  transform: "translate(-50%, -50%)",
                }}
                onClick={() => onThreatSelect?.(threat)}
                title={`${threat.type} - ${threat.severity}`}
              />
            ));
          })}
        </div>

        {/* Threat list */}
        <div className="mt-4 space-y-2">
          {sectors
            .flatMap((s) => s.threats)
            .slice(0, 5)
            .map((threat) => (
              <div
                key={threat.id}
                className={cn(
                  "flex items-center gap-3 px-3 py-2 rounded cursor-pointer",
                  selectedThreatId === threat.id
                    ? "bg-aviation-cyan/10"
                    : "hover:bg-aviation-bg-secondary",
                )}
                onClick={() => onThreatSelect?.(threat)}
              >
                <div
                  className={cn(
                    "w-2 h-2 rounded-full",
                    getSeverityBg(threat.severity),
                  )}
                />
                <div className="flex-1 min-w-0">
                  <div className="text-xs font-medium capitalize">
                    {threat.type}
                  </div>
                  <div className="text-[10px] text-aviation-text-muted truncate">
                    {threat.source} → {threat.target}
                  </div>
                </div>
                <span
                  className={cn(
                    "text-[10px] px-1.5 py-0.5 rounded capitalize",
                    getSeverityColor(threat.severity),
                    "bg-current/10",
                  )}
                >
                  {threat.severity}
                </span>
              </div>
            ))}
        </div>

        {showLegend && (
          <div className="flex items-center justify-center gap-4 mt-4 text-xs text-aviation-text-muted">
            <span className="flex items-center gap-1">
              <div className="w-2 h-2 rounded-full bg-red-500" /> Critical
            </span>
            <span className="flex items-center gap-1">
              <div className="w-2 h-2 rounded-full bg-orange-500" /> High
            </span>
            <span className="flex items-center gap-1">
              <div className="w-2 h-2 rounded-full bg-amber-500" /> Medium
            </span>
            <span className="flex items-center gap-1">
              <div className="w-2 h-2 rounded-full bg-green-500" /> Low
            </span>
          </div>
        )}
      </div>
    </div>
  );
};

// ============================================================================
// Security Timeline
// ============================================================================

interface SecurityEvent {
  id: string;
  type:
    | "login"
    | "logout"
    | "permission-change"
    | "resource-access"
    | "config-change"
    | "alert";
  severity: "info" | "warning" | "error" | "critical";
  message: string;
  timestamp: number;
  userName?: string;
  ipAddress?: string;
}

interface SecurityTimelineProps {
  events: SecurityEvent[];
  selectedEventId?: string | null;
  onEventSelect?: (event: SecurityEvent) => void;
  className?: string;
}

export const SecurityTimeline: React.FC<SecurityTimelineProps> = ({
  events,
  selectedEventId,
  onEventSelect,
  className,
}) => {
  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case "critical":
        return "text-red-400 bg-red-500/20";
      case "error":
        return "text-red-400 bg-red-500/20";
      case "warning":
        return "text-amber-400 bg-amber-500/20";
      default:
        return "text-blue-400 bg-blue-500/20";
    }
  };

  const getEventIcon = (type: string) => {
    switch (type) {
      case "login":
        return <Key className="w-4 h-4" />;
      case "logout":
        return <Lock className="w-4 h-4" />;
      case "permission-change":
        return <BadgeCheck className="w-4 h-4" />;
      case "resource-access":
        return <Eye className="w-4 h-4" />;
      case "config-change":
        return <Settings className="w-4 h-4" />;
      default:
        return <AlertCircle className="w-4 h-4" />;
    }
  };

  const formatTime = (ts: number) => {
    const date = new Date(ts);
    return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
  };

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        <Timeline className="w-5 h-5 text-aviation-cyan" />
        <span className="text-sm font-medium">Security Timeline</span>
      </div>
      <div className="flex-1 overflow-auto">
        <div className="relative">
          <div className="absolute left-6 top-0 bottom-0 w-0.5 bg-aviation-border-panel" />
          <div className="space-y-1 p-2">
            {events.map((event) => {
              const isSelected = event.id === selectedEventId;
              return (
                <div
                  key={event.id}
                  className={cn(
                    "relative flex items-start gap-3 px-3 py-2 rounded-lg cursor-pointer transition-colors",
                    isSelected
                      ? "bg-aviation-cyan/10"
                      : "hover:bg-aviation-bg-secondary",
                  )}
                  onClick={() => onEventSelect?.(event)}
                >
                  <div
                    className={cn(
                      "relative z-10 flex-shrink-0 w-10 h-10 rounded-full flex items-center justify-center",
                      getSeverityColor(event.severity),
                    )}
                  >
                    {getEventIcon(event.type)}
                  </div>
                  <div className="flex-1 min-w-0 pt-1">
                    <div className="text-xs text-aviation-text-primary">
                      {event.message}
                    </div>
                    <div className="flex items-center gap-2 mt-1 text-[10px] text-aviation-text-muted">
                      <span>{formatTime(event.timestamp)}</span>
                      {event.userName && <span>· {event.userName}</span>}
                      {event.ipAddress && <span>· {event.ipAddress}</span>}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
};

const Timeline: React.FC<{ className?: string }> = ({ className }) => (
  <History className={className} />
);

// ============================================================================
// Sandbox Boundary Viewer
// ============================================================================

interface SandboxBoundary {
  id: string;
  name: string;
  type: "function" | "container" | "vm" | "process";
  boundaries: Array<{
    id: string;
    type: "network" | "filesystem" | "memory" | "cpu";
    status: "enforced" | "relaxed" | "bypassed";
  }>;
  status: "active" | "inactive" | "breached";
}

interface SandboxBoundaryViewerProps {
  boundaries: SandboxBoundary[];
  selectedBoundaryId?: string | null;
  onBoundarySelect?: (boundary: SandboxBoundary) => void;
  className?: string;
}

export const SandboxBoundaryViewer: React.FC<SandboxBoundaryViewerProps> = ({
  boundaries,
  selectedBoundaryId,
  onBoundarySelect,
  className,
}) => {
  const getStatusColor = (status: string) => {
    switch (status) {
      case "active":
        return "text-green-400 bg-green-500/20";
      case "breached":
        return "text-red-400 bg-red-500/20";
      default:
        return "text-gray-400 bg-gray-500/20";
    }
  };

  const getBoundaryIcon = (type: string) => {
    switch (type) {
      case "function":
        return <Zap className="w-4 h-4" />;
      case "container":
        return <Container className="w-4 h-4" />;
      case "vm":
        return <MonitorIcon className="w-4 h-4" />;
      default:
        return <Square className="w-4 h-4" />;
    }
  };

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        <Shield className="w-5 h-5 text-aviation-cyan" />
        <span className="text-sm font-medium">Sandbox Boundaries</span>
      </div>
      <div className="flex-1 overflow-auto p-4">
        <div className="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-3">
          {boundaries.map((boundary) => {
            const isSelected = boundary.id === selectedBoundaryId;
            return (
              <div
                key={boundary.id}
                className={cn(
                  "p-4 rounded-lg border cursor-pointer transition-colors",
                  isSelected
                    ? "border-aviation-cyan bg-aviation-cyan/10"
                    : "border-aviation-border-panel hover:border-aviation-text-muted",
                )}
                onClick={() => onBoundarySelect?.(boundary)}
              >
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-2">
                    {getBoundaryIcon(boundary.type)}
                    <span className="text-sm font-medium">{boundary.name}</span>
                  </div>
                  <span
                    className={cn(
                      "px-1.5 py-0.5 rounded text-[10px] capitalize",
                      getStatusColor(boundary.status),
                    )}
                  >
                    {boundary.status}
                  </span>
                </div>
                <div className="space-y-2">
                  {boundary.boundaries.map((b) => (
                    <div
                      key={b.id}
                      className="flex items-center justify-between"
                    >
                      <span className="text-xs text-aviation-text-muted capitalize">
                        {b.type}
                      </span>
                      <span
                        className={cn(
                          "px-1.5 py-0.5 rounded text-[10px]",
                          b.status === "enforced"
                            ? "text-green-400 bg-green-500/20"
                            : b.status === "relaxed"
                              ? "text-amber-400 bg-amber-500/20"
                              : "text-red-400 bg-red-500/20",
                        )}
                      >
                        {b.status}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// API Exposure Analyzer
// ============================================================================

interface APIEndpoint {
  id: string;
  path: string;
  method: "GET" | "POST" | "PUT" | "DELETE" | "PATCH";
  exposure: "public" | "authenticated" | "internal" | "private";
  risk: "low" | "medium" | "high" | "critical";
  callsLast7Days?: number;
}

interface APIExposureAnalyzerProps {
  endpoints: APIEndpoint[];
  selectedEndpointId?: string | null;
  onEndpointSelect?: (endpoint: APIEndpoint) => void;
  className?: string;
}

export const APIExposureAnalyzer: React.FC<APIExposureAnalyzerProps> = ({
  endpoints,
  selectedEndpointId,
  onEndpointSelect,
  className,
}) => {
  const getMethodColor = (method: string) => {
    switch (method) {
      case "GET":
        return "text-green-400 bg-green-500/20";
      case "POST":
        return "text-blue-400 bg-blue-500/20";
      case "PUT":
        return "text-amber-400 bg-amber-500/20";
      case "DELETE":
        return "text-red-400 bg-red-500/20";
      default:
        return "text-gray-400 bg-gray-500/20";
    }
  };

  const getRiskColor = (risk: string) => {
    switch (risk) {
      case "critical":
        return "text-red-400";
      case "high":
        return "text-orange-400";
      case "medium":
        return "text-amber-400";
      default:
        return "text-green-400";
    }
  };

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        <Globe className="w-5 h-5 text-aviation-cyan" />
        <span className="text-sm font-medium">API Exposure</span>
        <span className="text-xs text-aviation-text-muted ml-auto">
          {endpoints.length} endpoints
        </span>
      </div>
      <div className="flex-1 overflow-auto">
        {endpoints.map((endpoint) => {
          const isSelected = endpoint.id === selectedEndpointId;
          return (
            <div
              key={endpoint.id}
              className={cn(
                "flex items-center gap-3 px-4 py-3 border-b border-aviation-border-panel cursor-pointer",
                isSelected
                  ? "bg-aviation-cyan/10"
                  : "hover:bg-aviation-bg-secondary",
              )}
              onClick={() => onEndpointSelect?.(endpoint)}
            >
              <span
                className={cn(
                  "px-1.5 py-0.5 rounded text-[10px] font-mono",
                  getMethodColor(endpoint.method),
                )}
              >
                {endpoint.method}
              </span>
              <code className="flex-1 text-xs font-mono truncate">
                {endpoint.path}
              </code>
              <span
                className={cn(
                  "px-1.5 py-0.5 rounded text-[10px] capitalize",
                  getRiskColor(endpoint.risk),
                  "bg-current/10",
                )}
              >
                {endpoint.risk}
              </span>
              <span
                className={cn(
                  "px-1.5 py-0.5 rounded text-[10px] capitalize",
                  "text-aviation-text-muted bg-aviation-bg-secondary",
                )}
              >
                {endpoint.exposure}
              </span>
            </div>
          );
        })}
      </div>
    </div>
  );
};

// ============================================================================
// Credential Access Monitor
// ============================================================================

interface CredentialAccess {
  id: string;
  credentialName: string;
  accessedBy: string;
  accessedAt: number;
  accessType: "read" | "write" | "delete" | "use";
  success: boolean;
  ipAddress?: string;
}

interface CredentialAccessMonitorProps {
  accesses: CredentialAccess[];
  selectedAccessId?: string | null;
  onAccessSelect?: (access: CredentialAccess) => void;
  className?: string;
}

export const CredentialAccessMonitor: React.FC<
  CredentialAccessMonitorProps
> = ({ accesses, selectedAccessId, onAccessSelect, className }) => {
  const formatTime = (ts: number) =>
    new Date(ts).toLocaleString([], {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        <Key className="w-5 h-5 text-aviation-amber" />
        <span className="text-sm font-medium">Credential Access</span>
      </div>
      <div className="flex-1 overflow-auto">
        {accesses.map((access) => {
          const isSelected = access.id === selectedAccessId;
          return (
            <div
              key={access.id}
              className={cn(
                "flex items-center gap-3 px-4 py-3 border-b border-aviation-border-panel cursor-pointer",
                isSelected
                  ? "bg-aviation-amber/10"
                  : "hover:bg-aviation-bg-secondary",
              )}
              onClick={() => onAccessSelect?.(access)}
            >
              <div
                className={cn(
                  "w-2 h-2 rounded-full",
                  access.success ? "bg-green-500" : "bg-red-500",
                )}
              />
              <div className="flex-1 min-w-0">
                <div className="text-xs font-medium">
                  {access.credentialName}
                </div>
                <div className="text-[10px] text-aviation-text-muted">
                  {access.accessedBy} · {access.accessType}
                </div>
              </div>
              <div className="text-right">
                <div className="text-xs text-aviation-text-muted">
                  {formatTime(access.accessedAt)}
                </div>
                {access.ipAddress && (
                  <div className="text-[10px] text-aviation-text-dim">
                    {access.ipAddress}
                  </div>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

// ============================================================================
// Runtime Isolation Map
// ============================================================================

interface RuntimeIsolate {
  id: string;
  name: string;
  type: "v8-isolate" | "wasm-instance" | "container" | "process";
  status: "running" | "paused" | "terminated" | "crashed";
  memoryUsage?: number;
  networkAllowed: boolean;
}

interface RuntimeIsolationMapProps {
  isolates: RuntimeIsolate[];
  selectedIsolateId?: string | null;
  onIsolateSelect?: (isolate: RuntimeIsolate) => void;
  className?: string;
}

export const RuntimeIsolationMap: React.FC<RuntimeIsolationMapProps> = ({
  isolates,
  selectedIsolateId,
  onIsolateSelect,
  className,
}) => {
  const getStatusColor = (status: string) => {
    switch (status) {
      case "running":
        return "text-green-400";
      case "paused":
        return "text-amber-400";
      case "crashed":
        return "text-red-400";
      default:
        return "text-gray-400";
    }
  };

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        <Container className="w-5 h-5 text-aviation-cyan" />
        <span className="text-sm font-medium">Runtime Isolation</span>
      </div>
      <div className="flex-1 overflow-auto p-4">
        <div className="space-y-2">
          {isolates.map((isolate) => {
            const isSelected = isolate.id === selectedIsolateId;
            return (
              <div
                key={isolate.id}
                className={cn(
                  "p-3 rounded-lg border cursor-pointer transition-colors",
                  isSelected
                    ? "border-aviation-cyan bg-aviation-cyan/10"
                    : "border-aviation-border-panel hover:border-aviation-text-muted",
                )}
                onClick={() => onIsolateSelect?.(isolate)}
              >
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm font-medium">{isolate.name}</span>
                  <span
                    className={cn(
                      "text-[10px] capitalize",
                      getStatusColor(isolate.status),
                    )}
                  >
                    {isolate.status}
                  </span>
                </div>
                <div className="flex items-center gap-4 text-[10px] text-aviation-text-muted">
                  <span className="capitalize">
                    {isolate.type.replace("-", " ")}
                  </span>
                  {isolate.memoryUsage && (
                    <span>{isolate.memoryUsage}% memory</span>
                  )}
                  <span
                    className={
                      isolate.networkAllowed ? "text-green-400" : "text-red-400"
                    }
                  >
                    {isolate.networkAllowed ? "Network ✓" : "Network ✗"}
                  </span>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Compliance Dashboard
// ============================================================================

interface ComplianceFramework {
  id: string;
  name: string;
  version: string;
  status: "compliant" | "partial" | "non-compliant" | "audit-required";
  score: number;
  requirements: Array<{ name: string; status: string }>;
}

interface ComplianceDashboardProps {
  frameworks: ComplianceFramework[];
  selectedFrameworkId?: string | null;
  onFrameworkSelect?: (framework: ComplianceFramework) => void;
  className?: string;
}

export const ComplianceDashboard: React.FC<ComplianceDashboardProps> = ({
  frameworks,
  selectedFrameworkId,
  onFrameworkSelect,
  className,
}) => {
  const getStatusColor = (status: string) => {
    switch (status) {
      case "compliant":
        return "text-green-400 bg-green-500/20";
      case "partial":
        return "text-amber-400 bg-amber-500/20";
      case "non-compliant":
        return "text-red-400 bg-red-500/20";
      default:
        return "text-blue-400 bg-blue-500/20";
    }
  };

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        <ShieldCheck className="w-5 h-5 text-aviation-cyan" />
        <span className="text-sm font-medium">Compliance</span>
      </div>
      <div className="flex-1 overflow-auto p-4">
        <div className="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-3">
          {frameworks.map((fw) => {
            const isSelected = fw.id === selectedFrameworkId;
            return (
              <div
                key={fw.id}
                className={cn(
                  "p-4 rounded-lg border cursor-pointer transition-colors",
                  isSelected
                    ? "border-aviation-cyan bg-aviation-cyan/10"
                    : "border-aviation-border-panel hover:border-aviation-text-muted",
                )}
                onClick={() => onFrameworkSelect?.(fw)}
              >
                <div className="flex items-center justify-between mb-3">
                  <span className="text-sm font-medium">{fw.name}</span>
                  <span
                    className={cn(
                      "px-1.5 py-0.5 rounded text-[10px]",
                      getStatusColor(fw.status),
                    )}
                  >
                    {fw.status}
                  </span>
                </div>
                <div className="relative w-full h-2 bg-aviation-bg-secondary rounded-full overflow-hidden mb-2">
                  <div
                    className="h-full bg-aviation-cyan transition-all"
                    style={{ width: `${fw.score}%` }}
                  />
                </div>
                <div className="flex items-center justify-between text-xs">
                  <span className="text-aviation-text-muted">Score</span>
                  <span className="font-mono">{fw.score}%</span>
                </div>
                <div className="mt-2 pt-2 border-t border-aviation-border-panel">
                  <div className="text-[10px] text-aviation-text-muted">
                    {fw.requirements.length} requirements
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Zero Trust Flow Viewer
// ============================================================================

interface ZeroTrustFlow {
  id: string;
  source: { type: string; identity: string; trustLevel: number };
  destination: { type: string; identity: string; sensitivity: string };
  status: "allowed" | "denied" | "pending";
  timestamp: number;
}

interface ZeroTrustFlowViewerProps {
  flows: ZeroTrustFlow[];
  selectedFlowId?: string | null;
  onFlowSelect?: (flow: ZeroTrustFlow) => void;
  className?: string;
}

export const ZeroTrustFlowViewer: React.FC<ZeroTrustFlowViewerProps> = ({
  flows,
  selectedFlowId,
  onFlowSelect,
  className,
}) => {
  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        <Shield className="w-5 h-5 text-aviation-cyan" />
        <span className="text-sm font-medium">Zero Trust Flows</span>
      </div>
      <div className="flex-1 overflow-auto">
        {flows.map((flow) => {
          const isSelected = flow.id === selectedFlowId;
          return (
            <div
              key={flow.id}
              className={cn(
                "flex items-center gap-3 px-4 py-3 border-b border-aviation-border-panel cursor-pointer",
                isSelected
                  ? "bg-aviation-cyan/10"
                  : "hover:bg-aviation-bg-secondary",
              )}
              onClick={() => onFlowSelect?.(flow)}
            >
              <div
                className={cn(
                  "w-2 h-2 rounded-full",
                  flow.status === "allowed"
                    ? "bg-green-500"
                    : flow.status === "denied"
                      ? "bg-red-500"
                      : "bg-amber-500",
                )}
              />
              <div className="flex-1 min-w-0">
                <div className="text-xs font-medium truncate">
                  {flow.source.identity}
                </div>
                <div className="text-[10px] text-aviation-text-muted truncate">
                  {flow.destination.identity}
                </div>
              </div>
              <div className="text-right">
                <div className="text-[10px] text-aviation-text-muted capitalize">
                  {flow.status}
                </div>
                <div className="text-[10px] text-aviation-text-dim">
                  {Math.round(flow.source.trustLevel * 100)}% trust
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

// ============================================================================
// Malicious Execution Detector
// ============================================================================

interface MaliciousExecution {
  id: string;
  processName: string;
  detectedAt: number;
  threatType: string;
  confidence: number;
  severity: "low" | "medium" | "high" | "critical";
  status: "detected" | "blocked" | "terminated";
}

interface MaliciousExecutionDetectorProps {
  executions: MaliciousExecution[];
  selectedExecutionId?: string | null;
  onExecutionSelect?: (execution: MaliciousExecution) => void;
  onBlock?: (executionId: string) => void;
  className?: string;
}

export const MaliciousExecutionDetector: React.FC<
  MaliciousExecutionDetectorProps
> = ({
  executions,
  selectedExecutionId,
  onExecutionSelect,
  onBlock,
  className,
}) => {
  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case "critical":
        return "text-red-400 bg-red-500/20";
      case "high":
        return "text-orange-400 bg-orange-500/20";
      default:
        return "text-amber-400 bg-amber-500/20";
    }
  };

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        <Bug className="w-5 h-5 text-red-400" />
        <span className="text-sm font-medium">Malicious Executions</span>
      </div>
      <div className="flex-1 overflow-auto">
        {executions.map((exec) => {
          const isSelected = exec.id === selectedExecutionId;
          return (
            <div
              key={exec.id}
              className={cn(
                "flex items-center gap-3 px-4 py-3 border-b border-aviation-border-panel",
                isSelected
                  ? "bg-red-500/10"
                  : "hover:bg-aviation-bg-secondary cursor-pointer",
              )}
              onClick={() => onExecutionSelect?.(exec)}
            >
              <div
                className={cn("p-2 rounded", getSeverityColor(exec.severity))}
              >
                <Bug className="w-4 h-4" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="text-xs font-medium">{exec.processName}</div>
                <div className="text-[10px] text-aviation-text-muted">
                  {exec.threatType} · {Math.round(exec.confidence * 100)}%
                  confidence
                </div>
              </div>
              <div className="flex items-center gap-2">
                <span
                  className={cn(
                    "px-1.5 py-0.5 rounded text-[10px] capitalize",
                    getSeverityColor(exec.severity),
                  )}
                >
                  {exec.severity}
                </span>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    onBlock?.(exec.id);
                  }}
                  className="p-1.5 hover:bg-red-500/20 rounded text-red-400"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

// ============================================================================
// Audit Trail Explorer
// ============================================================================

interface AuditEntry {
  id: string;
  action: string;
  actor: { type: string; name: string };
  timestamp: number;
  metadata?: Record<string, unknown>;
}

interface AuditTrailExplorerProps {
  entries: AuditEntry[];
  selectedEntryId?: string | null;
  onEntrySelect?: (entry: AuditEntry) => void;
  className?: string;
}

export const AuditTrailExplorer: React.FC<AuditTrailExplorerProps> = ({
  entries,
  selectedEntryId,
  onEntrySelect,
  className,
}) => {
  const formatTime = (ts: number) =>
    new Date(ts).toLocaleString([], {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        <Clipboard className="w-5 h-5 text-aviation-cyan" />
        <span className="text-sm font-medium">Audit Trail</span>
      </div>
      <div className="flex-1 overflow-auto">
        {entries.map((entry) => {
          const isSelected = entry.id === selectedEntryId;
          return (
            <div
              key={entry.id}
              className={cn(
                "px-4 py-3 border-b border-aviation-border-panel cursor-pointer",
                isSelected
                  ? "bg-aviation-cyan/10"
                  : "hover:bg-aviation-bg-secondary",
              )}
              onClick={() => onEntrySelect?.(entry)}
            >
              <div className="flex items-center justify-between mb-1">
                <span className="text-xs font-medium">{entry.action}</span>
                <span className="text-[10px] text-aviation-text-muted">
                  {formatTime(entry.timestamp)}
                </span>
              </div>
              <div className="text-[10px] text-aviation-text-muted">
                {entry.actor.type}: {entry.actor.name}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

// ============================================================================
// Encryption Status Panel
// ============================================================================

interface EncryptionStatus {
  id: string;
  name: string;
  type: "at-rest" | "in-transit" | "in-use";
  algorithm?: string;
  status: "encrypted" | "decrypted" | "key-missing" | "error";
  lastVerified?: number;
}

interface EncryptionStatusPanelProps {
  statuses: EncryptionStatus[];
  selectedStatusId?: string | null;
  onStatusSelect?: (status: EncryptionStatus) => void;
  className?: string;
}

export const EncryptionStatusPanel: React.FC<EncryptionStatusPanelProps> = ({
  statuses,
  selectedStatusId,
  onStatusSelect,
  className,
}) => {
  const getStatusColor = (status: string) => {
    switch (status) {
      case "encrypted":
        return "text-green-400 bg-green-500/20";
      case "decrypted":
        return "text-amber-400 bg-amber-500/20";
      case "key-missing":
        return "text-red-400 bg-red-500/20";
      default:
        return "text-gray-400 bg-gray-500/20";
    }
  };

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        <Lock className="w-5 h-5 text-aviation-cyan" />
        <span className="text-sm font-medium">Encryption Status</span>
      </div>
      <div className="flex-1 overflow-auto p-4">
        <div className="grid grid-cols-[repeat(auto-fill,minmax(200px,1fr))] gap-3">
          {statuses.map((status) => {
            const isSelected = status.id === selectedStatusId;
            return (
              <div
                key={status.id}
                className={cn(
                  "p-3 rounded-lg border cursor-pointer transition-colors",
                  isSelected
                    ? "border-aviation-cyan bg-aviation-cyan/10"
                    : "border-aviation-border-panel hover:border-aviation-text-muted",
                )}
                onClick={() => onStatusSelect?.(status)}
              >
                <div className="flex items-center justify-between mb-2">
                  <span className="text-xs font-medium">{status.name}</span>
                  <span
                    className={cn(
                      "px-1.5 py-0.5 rounded text-[10px] capitalize",
                      getStatusColor(status.status),
                    )}
                  >
                    {status.status}
                  </span>
                </div>
                <div className="text-[10px] text-aviation-text-muted capitalize">
                  {status.type.replace("-", " ")}
                </div>
                {status.algorithm && (
                  <div className="text-[10px] text-aviation-text-dim mt-1">
                    {status.algorithm}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Suspicious Behavior Timeline
// ============================================================================

interface SuspiciousBehavior {
  id: string;
  type: "anomaly" | "signature-match" | "heuristic" | "behavioral";
  description: string;
  severity: "low" | "medium" | "high" | "critical";
  detectedAt: number;
  score: number;
}

interface SuspiciousBehaviorTimelineProps {
  behaviors: SuspiciousBehavior[];
  selectedBehaviorId?: string | null;
  onBehaviorSelect?: (behavior: SuspiciousBehavior) => void;
  className?: string;
}

export const SuspiciousBehaviorTimeline: React.FC<
  SuspiciousBehaviorTimelineProps
> = ({ behaviors, selectedBehaviorId, onBehaviorSelect, className }) => {
  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case "critical":
        return "text-red-400 bg-red-500/20";
      case "high":
        return "text-orange-400 bg-orange-500/20";
      default:
        return "text-amber-400 bg-amber-500/20";
    }
  };

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        <Eye className="w-5 h-5 text-amber-400" />
        <span className="text-sm font-medium">Suspicious Behavior</span>
      </div>
      <div className="flex-1 overflow-auto">
        {behaviors.map((behavior) => {
          const isSelected = behavior.id === selectedBehaviorId;
          return (
            <div
              key={behavior.id}
              className={cn(
                "px-4 py-3 border-b border-aviation-border-panel cursor-pointer",
                isSelected
                  ? "bg-amber-500/10"
                  : "hover:bg-aviation-bg-secondary",
              )}
              onClick={() => onBehaviorSelect?.(behavior)}
            >
              <div className="flex items-center justify-between mb-1">
                <span className="text-xs font-medium capitalize">
                  {behavior.type}
                </span>
                <span
                  className={cn(
                    "px-1.5 py-0.5 rounded text-[10px] capitalize",
                    getSeverityColor(behavior.severity),
                  )}
                >
                  {behavior.severity}
                </span>
              </div>
              <div className="text-xs text-aviation-text-primary mb-1">
                {behavior.description}
              </div>
              <div className="flex items-center justify-between text-[10px] text-aviation-text-muted">
                <span>Score: {behavior.score}</span>
                <span>{new Date(behavior.detectedAt).toLocaleString()}</span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

// ============================================================================
// Vulnerability Scanner
// ============================================================================

interface Vulnerability {
  id: string;
  cveId?: string;
  name: string;
  severity: "critical" | "high" | "medium" | "low" | "info";
  status: "open" | "in-remediation" | "resolved" | "false-positive";
  cvssScore?: number;
}

interface VulnerabilityScannerProps {
  vulnerabilities: Vulnerability[];
  selectedVulnerabilityId?: string | null;
  onVulnerabilitySelect?: (vulnerability: Vulnerability) => void;
  className?: string;
}

export const VulnerabilityScanner: React.FC<VulnerabilityScannerProps> = ({
  vulnerabilities,
  selectedVulnerabilityId,
  onVulnerabilitySelect,
  className,
}) => {
  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case "critical":
        return "text-red-400 bg-red-500/20";
      case "high":
        return "text-orange-400 bg-orange-500/20";
      case "medium":
        return "text-amber-400 bg-amber-500/20";
      default:
        return "text-blue-400 bg-blue-500/20";
    }
  };

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        <Scan className="w-5 h-5 text-aviation-cyan" />
        <span className="text-sm font-medium">Vulnerabilities</span>
        <span className="text-xs text-aviation-text-muted ml-auto">
          {vulnerabilities.length} found
        </span>
      </div>
      <div className="flex-1 overflow-auto">
        {vulnerabilities.map((vuln) => {
          const isSelected = vuln.id === selectedVulnerabilityId;
          return (
            <div
              key={vuln.id}
              className={cn(
                "px-4 py-3 border-b border-aviation-border-panel cursor-pointer",
                isSelected
                  ? "bg-aviation-cyan/10"
                  : "hover:bg-aviation-bg-secondary",
              )}
              onClick={() => onVulnerabilitySelect?.(vuln)}
            >
              <div className="flex items-center justify-between mb-1">
                <span className="text-xs font-medium">{vuln.name}</span>
                <span
                  className={cn(
                    "px-1.5 py-0.5 rounded text-[10px] capitalize",
                    getSeverityColor(vuln.severity),
                  )}
                >
                  {vuln.severity}
                </span>
              </div>
              <div className="flex items-center justify-between">
                {vuln.cveId && (
                  <span className="text-[10px] font-mono text-aviation-cyan">
                    {vuln.cveId}
                  </span>
                )}
                {vuln.cvssScore && (
                  <span className="text-[10px] text-aviation-text-muted">
                    CVSS: {vuln.cvssScore}
                  </span>
                )}
                <span
                  className={cn(
                    "px-1.5 py-0.5 rounded text-[10px] capitalize",
                    "text-aviation-text-muted bg-aviation-bg-secondary",
                  )}
                >
                  {vuln.status}
                </span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

// ============================================================================
// Policy Engine Viewer
// ============================================================================

interface SecurityPolicy {
  id: string;
  name: string;
  description: string;
  type: "access" | "network" | "data" | "audit";
  status: "active" | "draft" | "disabled";
  rules: Array<{ id: string; condition: string; action: string }>;
  hitCount?: number;
}

interface PolicyEngineViewerProps {
  policies: SecurityPolicy[];
  selectedPolicyId?: string | null;
  onPolicySelect?: (policy: SecurityPolicy) => void;
  onPolicyToggle?: (policyId: string) => void;
  className?: string;
}

export const PolicyEngineViewer: React.FC<PolicyEngineViewerProps> = ({
  policies,
  selectedPolicyId,
  onPolicySelect,
  onPolicyToggle,
  className,
}) => {
  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        <Scale className="w-5 h-5 text-aviation-cyan" />
        <span className="text-sm font-medium">Policy Engine</span>
      </div>
      <div className="flex-1 overflow-auto">
        {policies.map((policy) => {
          const isSelected = policy.id === selectedPolicyId;
          return (
            <div
              key={policy.id}
              className={cn(
                "p-4 border-b border-aviation-border-panel cursor-pointer",
                isSelected
                  ? "bg-aviation-cyan/10"
                  : "hover:bg-aviation-bg-secondary",
              )}
              onClick={() => onPolicySelect?.(policy)}
            >
              <div className="flex items-center justify-between mb-2">
                <div>
                  <div className="text-sm font-medium">{policy.name}</div>
                  <div className="text-[10px] text-aviation-text-muted capitalize">
                    {policy.type}
                  </div>
                </div>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    onPolicyToggle?.(policy.id);
                  }}
                  className={cn(
                    "p-2 rounded transition-colors",
                    policy.status === "active"
                      ? "text-green-400 bg-green-500/20"
                      : "text-gray-400 bg-aviation-bg-secondary",
                  )}
                >
                  {policy.status === "active" ? (
                    <ShieldCheck className="w-4 h-4" />
                  ) : (
                    <ShieldAlert className="w-4 h-4" />
                  )}
                </button>
              </div>
              <p className="text-xs text-aviation-text-muted mb-3">
                {policy.description}
              </p>
              <div className="flex items-center justify-between text-[10px] text-aviation-text-dim">
                <span>{policy.rules.length} rules</span>
                {policy.hitCount && <span>{policy.hitCount} hits</span>}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
