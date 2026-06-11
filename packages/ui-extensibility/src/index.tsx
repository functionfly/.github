/**
 * @functionfly/ui-extensibility
 * Extension SDK - Plugin API, Hook System, and Extension Manager UI
 */

import * as React from "react";
import { cn } from "@functionfly/ui-core";
import {
  Puzzle,
  Plug,
  Unplug,
  Eye,
  EyeOff,
  Settings,
  Power,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Loader2,
  Package,
  Code,
  Lock,
  Zap,
  Monitor,
  ArrowRight,
  RefreshCw,
  ChevronDown,
  ChevronRight,
  FileCode,
  Shield,
} from "lucide-react";

// ============================================================================
// Types
// ============================================================================

export interface Extension {
  id: string;
  name: string;
  version: string;
  description: string;
  author: {
    name: string;
    url?: string;
  };
  icon?: string;
  status: "installed" | "enabled" | "disabled" | "error" | "updating";
  permissions: string[];
  hooks: string[];
  size: number; // KB
  installedAt: string;
  updatedAt: string;
  homepage?: string;
  category: "tool" | "theme" | "integration" | "ai" | "runtime" | "custom";
  error?: string;
}

export interface ExtensionHook {
  id: string;
  name: string;
  description: string;
  extensionId: string;
  events: string[];
  enabled: boolean;
}

export interface ExtensionSandbox {
  id: string;
  extensionId: string;
  permissions: string[];
  memoryLimit: number; // MB
  cpuLimit: number; // percentage
  networkAccess: boolean;
  filesystemAccess: boolean;
  status: "active" | "suspended" | "terminated";
}

export interface ExtensionMetrics {
  extensionId: string;
  executions: number;
  avgDuration: number; // ms
  errors: number;
  lastUsed: string;
}

export interface ExtensionSettings {
  extensionId: string;
  config: Record<string, unknown>;
  ui?: {
    showInToolbar?: boolean;
    toolbarPriority?: number;
    keyboardShortcuts?: string[];
  };
}

export interface ExtensionManagerProps {
  extensions: Extension[];
  onInstall?: (extensionId: string) => void;
  onUninstall?: (extensionId: string) => void;
  onEnable?: (extensionId: string) => void;
  onDisable?: (extensionId: string) => void;
  onConfigure?: (extensionId: string) => void;
  className?: string;
}

export interface ExtensionDetailPanelProps {
  extension: Extension;
  hooks: ExtensionHook[];
  sandbox: ExtensionSandbox;
  metrics?: ExtensionMetrics;
  onEnable?: () => void;
  onDisable?: () => void;
  onConfigure?: () => void;
  onUninstall?: () => void;
  onToggleHook?: (hookId: string) => void;
  className?: string;
}

export interface HookSystemVisualizerProps {
  hooks: ExtensionHook[];
  onHookClick?: (hookId: string) => void;
  className?: string;
}

export interface SandboxMonitorProps {
  sandboxes: ExtensionSandbox[];
  onSandboxAction?: (sandboxId: string, action: "suspend" | "terminate" | "resume") => void;
  className?: string;
}

export interface ExtensionMarketplaceBrowserProps {
  onInstall?: (extensionId: string) => void;
  className?: string;
}

export interface ExtensionSDKDebuggerProps {
  extensionId: string;
  logs: Array<{ timestamp: string; level: "info" | "warn" | "error"; message: string; context?: string }>;
  onClearLogs?: () => void;
  className?: string;
}

// ============================================================================
// ExtensionManager
// ============================================================================

const statusConfig: Record<string, { color: string; icon: React.ReactNode }> = {
  installed: { color: "#6b7280", icon: <Package className="size-3" /> },
  enabled: { color: "#10b981", icon: <CheckCircle className="size-3" /> },
  disabled: { color: "#6b7280", icon: <EyeOff className="size-3" /> },
  error: { color: "#ef4444", icon: <AlertTriangle className="size-3" /> },
  updating: { color: "#3b82f6", icon: <Loader2 className="size-3 animate-spin" /> },
};

const categoryColors: Record<string, string> = {
  tool: "#3b82f6",
  theme: "#8b5cf6",
  integration: "#10b981",
  ai: "#f97316",
  runtime: "#06b6d4",
  custom: "#6b7280",
};

export function ExtensionManager({
  extensions,
  onInstall,
  onUninstall,
  onEnable,
  onDisable,
  onConfigure,
  className,
}: ExtensionManagerProps) {
  const [filter, setFilter] = React.useState<Extension["status"] | "all">("all");
  const [search, setSearch] = React.useState("");
  const [selectedExt, setSelectedExt] = React.useState<string | null>(null);

  const filtered = extensions.filter((ext) => {
    if (filter !== "all" && ext.status !== filter) return false;
    if (search && !ext.name.toLowerCase().includes(search.toLowerCase())) return false;
    return true;
  });

  const stats = {
    total: extensions.length,
    enabled: extensions.filter((e) => e.status === "enabled").length,
    disabled: extensions.filter((e) => e.status === "disabled").length,
    error: extensions.filter((e) => e.status === "error").length,
  };

  return (
    <div className={cn("space-y-4", className)}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
            <Puzzle className="size-4 text-brand-400" /> Extensions
          </h4>
          <p className="text-[10px] text-text-muted">Manage plugins and hooks</p>
        </div>
        <div className="flex items-center gap-2">
          {(["all", "enabled", "disabled", "error"] as const).map((f) => (
            <button
              key={f}
              onClick={() => setFilter(f)}
              className={cn(
                "px-2 py-0.5 text-[10px] rounded capitalize transition-colors",
                filter === f ? "bg-brand-500/20 text-brand-400 font-medium" : "text-text-muted hover:text-text-secondary"
              )}
            >
              {f}
            </button>
          ))}
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-2">
        <div className="p-2 bg-bg-secondary rounded-lg text-center">
          <div className="text-lg font-bold text-text-primary">{stats.total}</div>
          <div className="text-[9px] text-text-muted">Total</div>
        </div>
        <div className="p-2 bg-bg-secondary rounded-lg text-center">
          <div className="text-lg font-bold text-emerald-400">{stats.enabled}</div>
          <div className="text-[9px] text-text-muted">Enabled</div>
        </div>
        <div className="p-2 bg-bg-secondary rounded-lg text-center">
          <div className="text-lg font-bold text-text-muted">{stats.disabled}</div>
          <div className="text-[9px] text-text-muted">Disabled</div>
        </div>
        <div className="p-2 bg-bg-secondary rounded-lg text-center">
          <div className="text-lg font-bold text-red-400">{stats.error}</div>
          <div className="text-[9px] text-text-muted">Error</div>
        </div>
      </div>

      {/* Search */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-text-muted" />
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search extensions..."
          className="w-full pl-9 pr-4 py-2 text-sm bg-bg-primary border border-border-subtle rounded-lg text-text-primary focus:outline-none focus:border-brand-500"
        />
      </div>

      {/* Extension list */}
      <div className="space-y-2 max-h-80 overflow-y-auto">
        {filtered.map((ext) => {
          const status = statusConfig[ext.status] || statusConfig.installed;
          const isSelected = selectedExt === ext.id;

          return (
            <div
              key={ext.id}
              onClick={() => setSelectedExt(isSelected ? null : ext.id)}
              className={cn(
                "p-3 rounded-lg border cursor-pointer transition-all",
                isSelected ? "border-brand-500 bg-brand-500/5" : "border-border-subtle hover:border-border-default",
                ext.status === "error" && "border-red-500/30"
              )}
            >
              <div className="flex items-center gap-3">
                <div
                  className="size-8 rounded-lg flex items-center justify-center text-xs"
                  style={{ backgroundColor: `${categoryColors[ext.category] || "#6b7280"}20`, color: categoryColors[ext.category] || "#6b7280" }}
                >
                  {ext.icon ? <img src={ext.icon} className="size-5" alt={ext.name} /> : <Puzzle className="size-4" />}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-0.5">
                    <span className="text-sm font-medium text-text-primary">{ext.name}</span>
                    <span className="text-[9px] text-text-muted">v{ext.version}</span>
                  </div>
                  <p className="text-[10px] text-text-muted line-clamp-1">{ext.description}</p>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-[10px]" style={{ color: status.color }}>
                    {status.icon}
                  </span>
                  <span className="text-[9px] text-text-muted">{ext.size}KB</span>
                </div>
              </div>

              {/* Expanded actions */}
              {isSelected && (
                <div className="mt-3 pt-3 border-t border-border-subtle flex gap-2 flex-wrap">
                  {ext.status === "disabled" || ext.status === "installed" ? (
                    <button
                      onClick={(e) => { e.stopPropagation(); onEnable?.(ext.id); }}
                      className="flex items-center gap-1 px-3 py-1.5 text-xs bg-emerald-500/20 text-emerald-400 rounded hover:bg-emerald-500/30"
                    >
                      <Plug className="size-3" /> Enable
                    </button>
                  ) : (
                    <button
                      onClick={(e) => { e.stopPropagation(); onDisable?.(ext.id); }}
                      className="flex items-center gap-1 px-3 py-1.5 text-xs bg-amber-500/20 text-amber-400 rounded hover:bg-amber-500/30"
                    >
                      <Unplug className="size-3" /> Disable
                    </button>
                  )}
                  <button
                    onClick={(e) => { e.stopPropagation(); onConfigure?.(ext.id); }}
                    className="flex items-center gap-1 px-3 py-1.5 text-xs bg-bg-secondary text-text-secondary rounded hover:bg-bg-hover"
                  >
                    <Settings className="size-3" /> Configure
                  </button>
                  <button
                    onClick={(e) => { e.stopPropagation(); onUninstall?.(ext.id); }}
                    className="flex items-center gap-1 px-3 py-1.5 text-xs bg-red-500/20 text-red-400 rounded hover:bg-red-500/30"
                  >
                    <XCircle className="size-3" /> Uninstall
                  </button>
                </div>
              )}

              {/* Error message */}
              {ext.status === "error" && ext.error && (
                <div className="mt-2 p-2 bg-red-500/10 rounded text-[10px] text-red-400 flex items-center gap-1">
                  <AlertTriangle className="size-3 shrink-0" />
                  {ext.error}
                </div>
              )}
            </div>
          );
        })}
      </div>

      {filtered.length === 0 && (
        <div className="text-center py-8 text-[11px] text-text-muted">
          No extensions found
        </div>
      )}
    </div>
  );
}

// ============================================================================
// ExtensionDetailPanel
// ============================================================================

export function ExtensionDetailPanel({
  extension,
  hooks,
  sandbox,
  metrics,
  onEnable,
  onDisable,
  onConfigure,
  onUninstall,
  onToggleHook,
  className,
}: ExtensionDetailPanelProps) {
  const [activeTab, setActiveTab] = React.useState<"overview" | "hooks" | "sandbox" | "logs">("overview");

  return (
    <div className={cn("space-y-4", className)}>
      <div className="flex items-center gap-4">
        <div
          className="size-12 rounded-lg flex items-center justify-center"
          style={{ backgroundColor: `${categoryColors[extension.category]}20`, color: categoryColors[extension.category] }}
        >
          {extension.icon ? <img src={extension.icon} className="size-8" alt={extension.name} /> : <Puzzle className="size-6" />}
        </div>
        <div className="flex-1">
          <div className="flex items-center gap-2">
            <h4 className="text-base font-semibold text-text-primary">{extension.name}</h4>
            <span className="text-[10px] px-1.5 py-0.5 bg-bg-tertiary text-text-muted rounded">v{extension.version}</span>
          </div>
          <p className="text-[11px] text-text-muted">{extension.description}</p>
        </div>
      </div>

      {/* Status */}
      <div className="flex items-center gap-2 p-3 bg-bg-secondary rounded-lg">
        <span className={cn(
          "text-[10px] px-2 py-0.5 rounded-full capitalize",
          extension.status === "enabled" ? "bg-emerald-500/20 text-emerald-400" :
          extension.status === "error" ? "bg-red-500/20 text-red-400" :
          "bg-gray-500/20 text-gray-400"
        )}>
          {extension.status}
        </span>
        <span className="text-[10px] text-text-muted capitalize">{extension.category}</span>
        <span className="text-[10px] text-text-muted ml-auto">Installed {new Date(extension.installedAt).toLocaleDateString()}</span>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 border-b border-border-subtle">
        {(["overview", "hooks", "sandbox", "logs"] as const).map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={cn(
              "px-3 py-1.5 text-xs font-medium capitalize transition-colors",
              activeTab === tab ? "text-brand-500 border-b-2 border-brand-500" : "text-text-muted hover:text-text-primary"
            )}
          >
            {tab}
          </button>
        ))}
      </div>

      {/* Tab content */}
      {activeTab === "overview" && (
        <div className="space-y-3">
          <div>
            <span className="text-[10px] font-medium text-text-muted">Author</span>
            <p className="text-sm text-text-primary">{extension.author.name}</p>
          </div>
          <div>
            <span className="text-[10px] font-medium text-text-muted">Permissions</span>
            <div className="flex flex-wrap gap-1 mt-1">
              {extension.permissions.map((perm) => (
                <span key={perm} className="text-[9px] px-1.5 py-0.5 bg-bg-tertiary text-text-muted rounded flex items-center gap-1">
                  <Lock className="size-2.5" /> {perm}
                </span>
              ))}
            </div>
          </div>
          <div>
            <span className="text-[10px] font-medium text-text-muted">Hooks</span>
            <p className="text-sm text-text-primary">{extension.hooks.join(", ")}</p>
          </div>
          {metrics && (
            <div className="grid grid-cols-4 gap-2">
              <div className="p-2 bg-bg-secondary rounded-lg text-center">
                <div className="text-sm font-bold text-text-primary">{metrics.executions}</div>
                <div className="text-[9px] text-text-muted">Executions</div>
              </div>
              <div className="p-2 bg-bg-secondary rounded-lg text-center">
                <div className="text-sm font-bold text-text-primary">{metrics.avgDuration}ms</div>
                <div className="text-[9px] text-text-muted">Avg Duration</div>
              </div>
              <div className="p-2 bg-bg-secondary rounded-lg text-center">
                <div className="text-sm font-bold text-error">{metrics.errors}</div>
                <div className="text-[9px] text-text-muted">Errors</div>
              </div>
              <div className="p-2 bg-bg-secondary rounded-lg text-center">
                <div className="text-sm font-bold text-text-muted">{new Date(metrics.lastUsed).toLocaleDateString()}</div>
                <div className="text-[9px] text-text-muted">Last Used</div>
              </div>
            </div>
          )}
        </div>
      )}

      {activeTab === "hooks" && (
        <div className="space-y-2">
          {hooks.map((hook) => (
            <div key={hook.id} className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
              <div className="flex items-center justify-between mb-1">
                <div className="flex items-center gap-2">
                  <Zap className="size-3 text-brand-400" />
                  <span className="text-sm font-medium text-text-primary">{hook.name}</span>
                </div>
                <button
                  onClick={() => onToggleHook?.(hook.id)}
                  className={cn(
                    "relative w-9 h-5 rounded-full transition-colors",
                    hook.enabled ? "bg-brand-500" : "bg-bg-tertiary"
                  )}
                >
                  <div className={cn(
                    "absolute top-0.5 w-4 h-4 rounded-full bg-white transition-transform",
                    hook.enabled ? "left-4.5" : "left-0.5"
                  )} />
                </button>
              </div>
              <p className="text-[10px] text-text-muted">{hook.description}</p>
              <div className="flex gap-1 mt-1.5">
                {hook.events.map((event) => (
                  <span key={event} className="text-[9px] px-1 py-0.5 bg-bg-tertiary text-text-muted rounded">
                    {event}
                  </span>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      {activeTab === "sandbox" && (
        <div className="space-y-3">
          <div className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm font-medium text-text-primary">Sandbox Status</span>
              <span className={cn(
                "text-[10px] px-2 py-0.5 rounded-full capitalize",
                sandbox.status === "active" ? "bg-emerald-500/20 text-emerald-400" :
                sandbox.status === "suspended" ? "bg-amber-500/20 text-amber-400" :
                "bg-red-500/20 text-red-400"
              )}>
                {sandbox.status}
              </span>
            </div>
            <div className="grid grid-cols-3 gap-3 text-[10px]">
              <div>
                <span className="text-text-muted">Memory</span>
                <div className="text-text-primary font-mono">{sandbox.memoryLimit}MB</div>
              </div>
              <div>
                <span className="text-text-muted">CPU</span>
                <div className="text-text-primary font-mono">{sandbox.cpuLimit}%</div>
              </div>
              <div>
                <span className="text-text-muted">Network</span>
                <div className={sandbox.networkAccess ? "text-emerald-400" : "text-red-400"}>
                  {sandbox.networkAccess ? "Allowed" : "Blocked"}
                </div>
              </div>
            </div>
          </div>
          <div>
            <span className="text-[10px] font-medium text-text-muted">Permissions</span>
            <div className="flex flex-wrap gap-1 mt-1">
              {sandbox.permissions.map((perm) => (
                <span key={perm} className="text-[9px] px-1.5 py-0.5 bg-bg-tertiary text-text-muted rounded">
                  {perm}
                </span>
              ))}
            </div>
          </div>
          <div className="flex items-center gap-2 text-[10px]">
            <Shield className="size-3 text-brand-400" />
            <span className="text-text-muted">Filesystem access:</span>
            <span className={sandbox.filesystemAccess ? "text-emerald-400" : "text-red-400"}>
              {sandbox.filesystemAccess ? "Allowed" : "Blocked"}
            </span>
          </div>
        </div>
      )}

      {activeTab === "logs" && (
        <div className="p-3 bg-bg-primary rounded-lg border border-border-subtle font-mono text-[10px] max-h-48 overflow-y-auto">
          <div className="text-text-muted text-center py-4">No logs yet</div>
        </div>
      )}

      {/* Actions */}
      <div className="flex gap-2">
        {extension.status === "disabled" ? (
          <button onClick={onEnable} className="flex-1 py-2 text-xs bg-emerald-500 text-white rounded-lg hover:bg-emerald-600">
            Enable
          </button>
        ) : (
          <button onClick={onDisable} className="flex-1 py-2 text-xs bg-amber-500/20 text-amber-400 rounded-lg hover:bg-amber-500/30">
            Disable
          </button>
        )}
        <button onClick={onConfigure} className="flex-1 py-2 text-xs bg-bg-secondary text-text-secondary rounded-lg hover:bg-bg-hover">
          Configure
        </button>
        <button onClick={onUninstall} className="flex-1 py-2 text-xs bg-red-500/20 text-red-400 rounded-lg hover:bg-red-500/30">
          Uninstall
        </button>
      </div>
    </div>
  );
}

// ============================================================================
// HookSystemVisualizer
// ============================================================================

export function HookSystemVisualizer({ hooks, onHookClick, className }: HookSystemVisualizerProps) {
  const hookTypes = React.useMemo(() => {
    const types: Record<string, ExtensionHook[]> = {};
    for (const hook of hooks) {
      for (const event of hook.events) {
        if (!types[event]) types[event] = [];
        types[event].push(hook);
      }
    }
    return types;
  }, [hooks]);

  const eventColors: Record<string, string> = {
    "onExecutionStart": "#3b82f6",
    "onExecutionEnd": "#10b981",
    "onError": "#ef4444",
    "onAgentSpawn": "#8b5cf6",
    "onAgentTerminate": "#f97316",
    "onMetric": "#06b6d4",
  };

  return (
    <div className={cn("space-y-4", className)}>
      <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
        <Zap className="size-4 text-brand-400" /> Hook System
      </h4>

      <div className="space-y-3">
        {Object.entries(hookTypes as Record<string, ExtensionHook[]>).map(([event, eventHooks]) => (
          <div key={event} className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
            <div className="flex items-center gap-2 mb-2">
              <div
                className="size-2 rounded-full"
                style={{ backgroundColor: eventColors[event] || "#6b7280" }}
              />
              <span className="text-xs font-medium text-text-primary">{event}</span>
              <span className="text-[9px] text-text-muted ml-auto">{eventHooks.length} hook(s)</span>
            </div>
            <div className="space-y-1.5">
              {eventHooks.map((hook) => (
                <div
                  key={hook.id}
                  onClick={() => onHookClick?.(hook.id)}
                  className="flex items-center justify-between p-2 bg-bg-primary rounded border border-border-subtle hover:border-border-default cursor-pointer"
                >
                  <div className="flex items-center gap-2">
                    <span className="text-[10px] text-text-secondary">{hook.name}</span>
                    <span className="text-[9px] text-text-muted">by {hook.extensionId}</span>
                  </div>
                  <span className={cn(
                    "text-[9px] px-1.5 py-0.5 rounded-full",
                    hook.enabled ? "bg-emerald-500/20 text-emerald-400" : "bg-gray-500/20 text-gray-400"
                  )}>
                    {hook.enabled ? "active" : "disabled"}
                  </span>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>

      {hooks.length === 0 && (
        <div className="text-center py-8 text-[11px] text-text-muted">
          No hooks registered
        </div>
      )}
    </div>
  );
}

// ============================================================================
// SandboxMonitor
// ============================================================================

export function SandboxMonitor({
  sandboxes,
  onSandboxAction,
  className,
}: SandboxMonitorProps) {
  return (
    <div className={cn("space-y-4", className)}>
      <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
        <Monitor className="size-4 text-brand-400" /> Sandbox Monitor
      </h4>

      <div className="space-y-2">
        {sandboxes.map((sandbox) => (
          <div key={sandbox.id} className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-2">
                <span className={cn(
                  "size-2 rounded-full",
                  sandbox.status === "active" ? "bg-emerald-400" :
                  sandbox.status === "suspended" ? "bg-amber-400" : "bg-red-400"
                )} />
                <span className="text-sm font-medium text-text-primary">{sandbox.extensionId}</span>
              </div>
              <div className="flex gap-1">
                {sandbox.status === "active" && (
                  <button
                    onClick={() => onSandboxAction?.(sandbox.id, "suspend")}
                    className="text-[9px] px-2 py-0.5 bg-amber-500/20 text-amber-400 rounded hover:bg-amber-500/30"
                  >
                    Suspend
                  </button>
                )}
                {sandbox.status === "suspended" && (
                  <button
                    onClick={() => onSandboxAction?.(sandbox.id, "resume")}
                    className="text-[9px] px-2 py-0.5 bg-emerald-500/20 text-emerald-400 rounded hover:bg-emerald-500/30"
                  >
                    Resume
                  </button>
                )}
                <button
                  onClick={() => onSandboxAction?.(sandbox.id, "terminate")}
                  className="text-[9px] px-2 py-0.5 bg-red-500/20 text-red-400 rounded hover:bg-red-500/30"
                >
                  Terminate
                </button>
              </div>
            </div>

            <div className="grid grid-cols-3 gap-2 text-[10px]">
              <div>
                <span className="text-text-muted">Memory</span>
                <div className="text-text-primary font-mono">{sandbox.memoryLimit}MB</div>
              </div>
              <div>
                <span className="text-text-muted">CPU</span>
                <div className="text-text-primary font-mono">{sandbox.cpuLimit}%</div>
              </div>
              <div>
                <span className="text-text-muted">Network</span>
                <div className={sandbox.networkAccess ? "text-emerald-400" : "text-red-400"}>
                  {sandbox.networkAccess ? "Yes" : "No"}
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>

      {sandboxes.length === 0 && (
        <div className="text-center py-8 text-[11px] text-text-muted">
          No active sandboxes
        </div>
      )}
    </div>
  );
}

// ============================================================================
// ExtensionSDKDebugger
// ============================================================================

export function ExtensionSDKDebugger({
  extensionId,
  logs,
  onClearLogs,
  className,
}: ExtensionSDKDebuggerProps) {
  return (
    <div className={cn("space-y-3", className)}>
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
          <Code className="size-4 text-brand-400" /> SDK Debugger
        </h4>
        <button onClick={onClearLogs} className="text-[10px] text-text-muted hover:text-text-secondary">
          Clear logs
        </button>
      </div>

      <div className="p-2 bg-bg-primary rounded-lg border border-border-subtle font-mono text-[10px] max-h-64 overflow-y-auto">
        {logs.length === 0 ? (
          <div className="text-center py-4 text-text-muted">No logs</div>
        ) : (
          <div className="space-y-0.5">
            {logs.map((log, i) => (
              <div
                key={i}
                className={cn(
                  "flex gap-2 py-0.5",
                  log.level === "error" && "text-red-400",
                  log.level === "warn" && "text-amber-400"
                )}
              >
                <span className="text-text-muted shrink-0">{log.timestamp.split("T")[1]?.slice(0, 12) || ""}</span>
                <span className="uppercase font-bold w-12 shrink-0">{log.level}</span>
                <span className="text-text-secondary">{log.message}</span>
                {log.context && <span className="text-text-muted">[{log.context}]</span>}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

// ============================================================================
// Helper
// ============================================================================

function Search({ className }: { className?: string }) {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={className}>
      <circle cx="11" cy="11" r="8" />
      <path d="m21 21-4.35-4.35" />
    </svg>
  );
}

// ============================================================================
// Index
// ============================================================================