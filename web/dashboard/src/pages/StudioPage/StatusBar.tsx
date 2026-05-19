import React from "react";
import {
  Wifi,
  WifiOff,
  Cloud,
  CloudOff,
  Database,
  Cpu,
  Activity,
  Clock,
  CheckCircle,
  AlertCircle,
} from "lucide-react";
import { cn } from "@functionfly/ui-core";

interface StatusBarProps {
  isConnected: boolean;
  isSynced: boolean;
  runtimeName: string;
  latencyMs: number;
  lastSyncTime: Date | null;
  activeAgents: number;
  activeExecutions: number;
}

export function StatusBar({
  isConnected,
  isSynced,
  runtimeName,
  latencyMs,
  lastSyncTime,
  activeAgents,
  activeExecutions,
}: StatusBarProps) {
  return (
    <div className="h-6 bg-bg-tertiary border-t border-border-subtle flex items-center justify-between px-3 text-[10px] text-text-muted">
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-1.5">
          {isConnected ? (
            <Wifi className="size-3 text-success" />
          ) : (
            <WifiOff className="size-3 text-error" />
          )}
          <span>{isConnected ? "Connected" : "Offline"}</span>
        </div>

        <div className="flex items-center gap-1.5">
          {isSynced ? (
            <Cloud className="size-3 text-success" />
          ) : (
            <CloudOff className="size-3 text-warning" />
          )}
          <span>{isSynced ? "Synced" : "Syncing..."}</span>
        </div>

        <div className="flex items-center gap-1.5">
          <Cpu className="size-3 text-brand-400" />
          <span>{runtimeName}</span>
        </div>

        <div className="flex items-center gap-1.5">
          <Activity className="size-3 text-warning" />
          <span>{latencyMs}ms</span>
        </div>
      </div>

      <div className="flex items-center gap-4">
        {lastSyncTime && (
          <div className="flex items-center gap-1.5">
            <Clock className="size-3" />
            <span>Last sync: {formatTime(lastSyncTime)}</span>
          </div>
        )}

        <div className="flex items-center gap-1.5">
          <div
            className={cn(
              "w-1.5 h-1.5 rounded-full",
              activeExecutions > 0 ? "bg-success animate-pulse" : "bg-text-muted"
            )}
          />
          <span>
            {activeExecutions} exec{activeExecutions !== 1 ? "s" : ""}
          </span>
        </div>

        <div className="flex items-center gap-1.5">
          <span>
            {activeAgents} agent{activeAgents !== 1 ? "s" : ""}
          </span>
        </div>
      </div>
    </div>
  );
}

function formatTime(date: Date): string {
  const now = new Date();
  const diff = now.getTime() - date.getTime();

  if (diff < 60000) {
    return "just now";
  }
  if (diff < 3600000) {
    const mins = Math.floor(diff / 60000);
    return `${mins}m ago`;
  }
  return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
}