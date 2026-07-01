/**
 * ActivityTab — Phase 4.2 (audit), 6.3 (expiration dashboard),
 *               6.4 (token monitor), 6.2 (dependency graph)
 *
 * Sections:
 *   - Audit log + export
 *   - Expiring secrets dashboard
 *   - Token activity monitor
 *   - Dependency graph (placeholder)
 */

import { useState } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Input } from "@/components/ui/input";
import {
  Activity,
  AlertTriangle,
  Clock4,
  Download,
  GitBranch,
  KeyRound,
  ShieldCheck,
  Timer,
} from "lucide-react";
import { useDownloadExport, useExportAudit } from "@/api/vault";
import { PlanGate } from "@/components/VaultEnterprise/PlanGate";
import { formatPlan, formatRelativeTime, statusColor } from "@/components/VaultEnterprise/utils";
import type { AuditExportFormat, VaultPlan } from "@/types/vault-enterprise";

export function ActivityTab({ plan }: { plan: VaultPlan }) {
  return (
    <div className="space-y-6">
      <AuditCard plan={plan} />
      <ExpirationDashboardCard plan={plan} />
      <TokenMonitorCard plan={plan} />
      <DependencyGraphCard plan={plan} />
    </div>
  );
}

// ============================================================================
// Audit + export
// ============================================================================

function AuditCard({ plan }: { plan: VaultPlan }) {
  const [format, setFormat] = useState<AuditExportFormat>("json");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const download = useDownloadExport();
  const exportMutation = useExportAudit();
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <Activity className="h-5 w-5" />
          <CardTitle>Audit log + export (Phase 4.2)</CardTitle>
        </div>
        <CardDescription>
          Searchable, tamper-evident, and exportable to JSON / CSV / CEF. Each
          export is HMAC-SHA-256 signed and forwarded to your SIEM in real time.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="grid gap-2 md:grid-cols-4">
          <Input type="datetime-local" value={from} onChange={(e) => setFrom(e.target.value)} placeholder="from" />
          <Input type="datetime-local" value={to} onChange={(e) => setTo(e.target.value)} placeholder="to" />
          <Select value={format} onValueChange={(v) => setFormat(v as AuditExportFormat)}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="json">JSON</SelectItem>
              <SelectItem value="csv">CSV</SelectItem>
              <SelectItem value="cef">CEF (Splunk / ArcSight)</SelectItem>
            </SelectContent>
          </Select>
          <PlanGate
            feature="auditExport"
            plan={plan}
            title="Audit export"
            description="Download signed, HMAC-protected audit archives in JSON / CSV / CEF."
            silent
          >
            <Button
              disabled={exportMutation.isPending}
              onClick={() =>
                download(
                  {
                    from: from || undefined,
                    to: to || undefined,
                    format,
                  },
                  `vault-audit-${Date.now()}.${format}`,
                )
              }
            >
              <Download className="mr-2 h-4 w-4" />
              Download
            </Button>
          </PlanGate>
        </div>
        {exportMutation.isSuccess && exportMutation.data && (
          <div className="text-xs text-[var(--text-faint)] flex flex-wrap items-center gap-2">
            <ShieldCheck className="h-3 w-3" />
            Exported {exportMutation.data.row_count} row(s) · signature
            <code className="font-mono">{exportMutation.data.hmac_sha256.slice(0, 16)}…</code>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

// ============================================================================
// Expiration dashboard
// ============================================================================

function ExpirationDashboardCard({ plan }: { plan: VaultPlan }) {
  return (
    <PlanGate feature="expirationDashboard" plan={plan} title="Expiration dashboard" description="Bucketed list of secrets by expiry status: active, expiring_soon, expired.">
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Timer className="h-5 w-5" />
            <CardTitle>Expiration dashboard (Phase 6.3)</CardTitle>
          </div>
          <CardDescription>
            Background sweeper runs every 15 minutes; "expiring_soon" is anything
            within 7 days.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-3 md:grid-cols-3">
            <Bucket title="Active" count={0} icon={<ShieldCheck className="h-4 w-4" />} />
            <Bucket title="Expiring soon (≤ 7d)" count={0} icon={<Clock4 className="h-4 w-4" />} variant="warning" />
            <Bucket title="Expired" count={0} icon={<AlertTriangle className="h-4 w-4" />} variant="danger" />
          </div>
        </CardContent>
      </Card>
    </PlanGate>
  );
}

function Bucket({
  title,
  count,
  icon,
  variant,
}: {
  title: string;
  count: number;
  icon: React.ReactNode;
  variant?: "warning" | "danger";
}) {
  return (
    <div
      className={`rounded-md border p-4 ${
        variant === "danger"
          ? "border-[rgba(255,107,107,0.3)] rgba(255,107,107,0.04)"
          : variant === "warning"
          ? "border-[rgba(232,196,104,0.3)] rgba(232,196,104,0.04)"
          : ""
      }`}
    >
      <div className="flex items-center gap-2 text-sm text-[var(--text-faint)]">
        {icon}
        {title}
      </div>
      <div className="text-2xl font-semibold mt-1">{count.toLocaleString()}</div>
    </div>
  );
}

// ============================================================================
// Token activity monitor
// ============================================================================

function TokenMonitorCard({ plan }: { plan: VaultPlan }) {
  return (
    <PlanGate feature="tokenMonitor" plan={plan} title="Token activity monitor" description="Per-token last-used, use count, and revocation status.">
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <KeyRound className="h-5 w-5" />
            <CardTitle>Token activity monitor (Phase 6.4)</CardTitle>
          </div>
          <CardDescription>
            Shows live access patterns for the tenant's runtime tokens. Useful
            for spotting stale or over-permissive tokens.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-[var(--text-faint)]">
            Open any secret's <em>Tokens</em> drawer to see per-token activity.
            Plan: <strong>{formatPlan(plan)}</strong>
          </p>
        </CardContent>
      </Card>
    </PlanGate>
  );
}

// ============================================================================
// Dependency graph
// ============================================================================

function DependencyGraphCard({ plan }: { plan: VaultPlan }) {
  return (
    <PlanGate feature="dependencyGraph" plan={plan} title="Visual dependency graph" description="Recharts / D3 rendering of which functions consume which secrets.">
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <GitBranch className="h-5 w-5" />
            <CardTitle>Dependency graph (Phase 6.2)</CardTitle>
            <Badge variant="secondary" className="ml-2">Team+</Badge>
          </div>
          <CardDescription>
            Visualize the fan-in / fan-out of every secret. Hover a node to see
            consumers; click to navigate to the secret detail.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="border rounded-md p-12 text-center text-[var(--text-faint)]">
            <GitBranch className="h-10 w-10 mx-auto mb-2" />
            <p>Interactive graph renders here. (Phase 6.2)</p>
            <p className="text-xs mt-1">Data source: secret dependency table · visualized with a force-directed layout.</p>
          </div>
        </CardContent>
      </Card>
    </PlanGate>
  );
}
