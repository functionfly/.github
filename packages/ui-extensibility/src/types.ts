/**
 * @functionfly/ui-extensibility
 * Type definitions
 */

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
  size: number;
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
  memoryLimit: number;
  cpuLimit: number;
  networkAccess: boolean;
  filesystemAccess: boolean;
  status: "active" | "suspended" | "terminated";
}

export interface ExtensionMetrics {
  extensionId: string;
  executions: number;
  avgDuration: number;
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