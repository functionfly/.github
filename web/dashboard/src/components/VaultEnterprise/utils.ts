/**
 * Shared utilities for the Vault Enterprise UI.
 */

import type { VaultPlan } from "@/types/vault-enterprise";

export function formatPlan(plan: VaultPlan): string {
  return plan.charAt(0).toUpperCase() + plan.slice(1);
}

export function formatRelativeTime(iso: string | undefined): string {
  if (!iso) return "—";
  const ms = Date.now() - new Date(iso).getTime();
  if (ms < 0) return "in the future";
  const sec = Math.floor(ms / 1000);
  if (sec < 60) return `${sec}s ago`;
  const min = Math.floor(sec / 60);
  if (min < 60) return `${min}m ago`;
  const hr = Math.floor(min / 60);
  if (hr < 24) return `${hr}h ago`;
  const d = Math.floor(hr / 24);
  return `${d}d ago`;
}

export function formatBytes(n: number): string {
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  return `${(n / (1024 * 1024)).toFixed(1)} MB`;
}

export function percent(numerator: number, denominator: number): number {
  if (denominator <= 0) return 0;
  return Math.min(100, Math.max(0, Math.round((numerator / denominator) * 100)));
}

export function statusColor(status: string): "default" | "secondary" | "destructive" | "outline" {
  switch (status) {
    case "active":
    case "approved":
    case "success":
      return "default";
    case "pending":
    case "expiring_soon":
      return "secondary";
    case "expired":
    case "revoked":
    case "denied":
    case "failed":
      return "destructive";
    default:
      return "outline";
  }
}
