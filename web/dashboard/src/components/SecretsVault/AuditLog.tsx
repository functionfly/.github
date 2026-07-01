/**
 * AuditLog - View vault audit trail
 * Displays audit log entries with filtering and pagination
 */

import { useState, useMemo } from "react";
import {
  Shield,
  Search,
  Download,
  Filter,
  Clock,
  User,
  Key,
  FileEdit,
  Trash2,
  Eye,
  XCircle,
  CheckCircle,
  ChevronDown,
  ChevronUp,
  Loader2,
  AlertTriangle,
} from "lucide-react";
import { format, formatDistanceToNow } from "date-fns";
import { cn } from "@/lib/utils";
import { useVaultAuditLog, useSecretAuditLog } from "@/hooks/useVault";
import type { AuditAction, AuditLogEntry, ActorType } from "@/types/vault";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";

// Action icon mapping
const actionIcons: Record<AuditAction, typeof Eye> = {
  create: Key,
  read: Eye,
  update: FileEdit,
  delete: Trash2,
  use: CheckCircle,
  revoke: XCircle,
  version: FileEdit,
  rollback: XCircle,
};

// Action label mapping
const actionLabels: Record<AuditAction, string> = {
  create: "Created",
  read: "Accessed",
  update: "Updated",
  delete: "Deleted",
  use: "Used",
  revoke: "Revoked",
  version: "Versioned",
  rollback: "Rolled back",
};

// Action badge variant mapping
const actionVariants: Record<
  AuditAction,
  "default" | "secondary" | "outline" | "success" | "warning" | "error"
> = {
  create: "success",
  read: "secondary",
  update: "warning",
  delete: "error",
  use: "default",
  revoke: "error",
  version: "warning",
  rollback: "error",
};

// Actor type icon mapping
const actorTypeIcons: Record<ActorType, typeof User> = {
  user: User,
  token: Key,
  system: Shield,
  api_key: Key,
};

// Actor type label mapping
const actorTypeLabels: Record<ActorType, string> = {
  user: "User",
  token: "Token",
  system: "System",
  api_key: "API Key",
};

export interface AuditLogProps {
  secretId?: string; // If provided, shows audit log for specific secret
  className?: string;
  limit?: number;
}

export function AuditLog({
  secretId,
  className,
  limit: initialLimit = 50,
}: AuditLogProps) {
  const [searchQuery, setSearchQuery] = useState("");
  const [actionFilter, setActionFilter] = useState<AuditAction | "all">("all");
  const [limit, setLimit] = useState(initialLimit);
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());

  // Use appropriate hook based on whether we're viewing all logs or secret-specific
  const {
    data: allEntries,
    isLoading: isLoadingAll,
    error: errorAll,
  } = useVaultAuditLog(secretId ? undefined : limit);

  const {
    data: secretEntries,
    isLoading: isLoadingSecret,
    error: errorSecret,
  } = useSecretAuditLog(secretId || "", secretId ? limit : undefined);

  const entriesData = secretId ? secretEntries : allEntries;
  const entriesArray = entriesData?.entries ?? [];
  const isLoading = secretId ? isLoadingSecret : isLoadingAll;
  const error = secretId ? errorSecret : errorAll;

  // Filter entries
  const filteredEntries = useMemo(() => {
    if (!entriesArray) return [];

    return entriesArray.filter((entry) => {
      // Action filter
      if (actionFilter !== "all" && entry.action !== actionFilter) {
        return false;
      }

      // Search filter
      if (searchQuery) {
        const query = searchQuery.toLowerCase();
        const matchesSearch =
          entry.actor_id.toLowerCase().includes(query) ||
          entry.secret_id?.toLowerCase().includes(query) ||
          entry.request_id?.toLowerCase().includes(query) ||
          entry.error_message?.toLowerCase().includes(query);

        if (!matchesSearch) return false;
      }

      return true;
    });
  }, [entriesArray, actionFilter, searchQuery]);

  // Toggle row expansion
  const toggleRow = (id: string) => {
    setExpandedRows((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(id)) {
        newSet.delete(id);
      } else {
        newSet.add(id);
      }
      return newSet;
    });
  };

  // Export to CSV
  const handleExport = () => {
    if (!filteredEntries.length) return;

    const headers = [
      "Timestamp",
      "Action",
      "Actor ID",
      "Actor Type",
      "Secret ID",
      "Success",
      "Error Message",
      "IP Address",
    ];

    const csvContent = [
      headers.join(","),
      ...filteredEntries.map((entry) =>
        [
          format(new Date(entry.created_at), "yyyy-MM-dd HH:mm:ss"),
          entry.action,
          entry.actor_id,
          entry.actor_type,
          entry.secret_id || "",
          entry.success ? "Yes" : "No",
          entry.error_message || "",
          entry.ip_address || "",
        ].join(",")
      ),
    ].join("\n");

    const blob = new Blob([csvContent], { type: "text/csv" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `audit-log-${format(new Date(), "yyyy-MM-dd")}.csv`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  // Loading state
  if (isLoading) {
    return (
      <div className={cn("space-y-4", className)}>
        <div className="flex items-center justify-between">
          <Skeleton className="h-10 w-64" />
          <Skeleton className="h-10 w-32" />
        </div>
        <div className="rounded-lg border border-[var(--panel-edge)]-subtle">
          <div className="space-y-2 p-4">
            {Array.from({ length: 5 }).map((_, i) => (
              <Skeleton key={i} className="h-12 w-full" />
            ))}
          </div>
        </div>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div
        className={cn(
          "rounded-lg border border-error/20 bg-error-glow p-8 text-center",
          className
        )}
      >
        <AlertTriangle className="mx-auto h-12 w-12 text-error mb-4" />
        <h3 className="text-lg font-semibold text-error mb-2">
          Failed to load audit log
        </h3>
        <p className="text-[var(--text-dim)] mb-4">
          {error instanceof Error ? error.message : "An unexpected error occurred"}
        </p>
        <Button onClick={() => window.location.reload()} variant="outline">
          Retry
        </Button>
      </div>
    );
  }

  return (
    <div className={cn("space-y-4", className)}>
      {/* Header with filters */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-3 flex-1">
          <div className="relative flex-1 max-w-md">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--text-faint)]" />
            <Input
              placeholder="Search audit log..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>
          <Select
            value={actionFilter}
            onValueChange={(value) =>
              setActionFilter(value as AuditAction | "all")
            }
          >
            <SelectTrigger className="w-[140px]">
              <Filter className="h-4 w-4 mr-2" />
              <SelectValue placeholder="Filter action" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Actions</SelectItem>
              <SelectItem value="create">Created</SelectItem>
              <SelectItem value="read">Accessed</SelectItem>
              <SelectItem value="update">Updated</SelectItem>
              <SelectItem value="delete">Deleted</SelectItem>
              <SelectItem value="use">Used</SelectItem>
              <SelectItem value="revoke">Revoked</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <Button
          variant="outline"
          onClick={handleExport}
          disabled={!filteredEntries?.length}
          className="gap-2"
        >
          <Download className="h-4 w-4" />
          Export CSV
        </Button>
      </div>

      {/* Empty state */}
      {filteredEntries?.length === 0 && (
        <div className="rounded-lg border border-[var(--panel-edge)]-subtle bg-[var(--panel-raised)] p-12 text-center">
          <div className="mx-auto h-16 w-16 rounded-full bg-gradient-to-br from-brand-500/20 to-purple-500/20 flex items-center justify-center mb-4">
            <Shield className="h-8 w-8 text-[var(--status-ok)]" />
          </div>
          <h3 className="text-lg font-semibold text-card-foreground mb-2">
            No audit entries found
          </h3>
          <p className="text-[var(--text-dim)] max-w-sm mx-auto">
            {searchQuery || actionFilter !== "all"
              ? "No entries match your current filters. Try adjusting your search criteria."
              : "Audit log entries will appear here when secrets are created, accessed, or modified."}
          </p>
        </div>
      )}

      {/* Audit log table */}
      {filteredEntries && filteredEntries.length > 0 && (
        <div className="rounded-lg border border-[var(--panel-edge)]-subtle overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Timestamp</TableHead>
                <TableHead>Action</TableHead>
                <TableHead>Actor</TableHead>
                <TableHead>Secret</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="w-16"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredEntries.map((entry) => {
                const ActionIcon = actionIcons[entry.action];
                const ActorIcon = actorTypeIcons[entry.actor_type];
                const isExpanded = expandedRows.has(entry.id);

                return (
                  <TableRow
                    key={entry.id}
                    className={cn(
                      "cursor-pointer",
                      !entry.success && "bg-error-glow/30"
                    )}
                    onClick={() => toggleRow(entry.id)}
                  >
                    <TableCell>
                      <div className="flex flex-col">
                        <span className="text-sm text-card-foreground">
                          {formatDistanceToNow(new Date(entry.created_at), {
                            addSuffix: true,
                          })}
                        </span>
                        <span className="text-xs text-[var(--text-faint)]">
                          {format(new Date(entry.created_at), "HH:mm:ss")}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <div
                          className={cn(
                            "h-8 w-8 rounded-lg flex items-center justify-center",
                            {
                              "bg-success-glow": entry.action === "create",
                              "rgba(143,255,208,0.06)": entry.action === "read",
                              "bg-warning-glow": entry.action === "update",
                              "bg-error-glow":
                                entry.action === "delete" ||
                                entry.action === "revoke",
                            }
                          )}
                        >
                          <ActionIcon
                            className={cn("h-4 w-4", {
                              "text-success": entry.action === "create",
                              "text-[var(--status-ok)]": entry.action === "read",
                              "text-warning": entry.action === "update",
                              "text-error":
                                entry.action === "delete" ||
                                entry.action === "revoke",
                            })}
                          />
                        </div>
                        <Badge variant={actionVariants[entry.action]}>
                          {actionLabels[entry.action]}
                        </Badge>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <ActorIcon className="h-3.5 w-3.5 text-[var(--text-faint)]" />
                        <div className="flex flex-col">
                          <span
                            className="text-sm text-card-foreground truncate max-w-[120px]"
                            title={entry.actor_id}
                          >
                            {entry.actor_id.slice(0, 8)}...
                          </span>
                          <span className="text-xs text-[var(--text-faint)]">
                            {actorTypeLabels[entry.actor_type]}
                          </span>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      {entry.secret_id ? (
                        <code className="text-xs bg-bg-secondary px-2 py-1 rounded">
                          {entry.secret_id.slice(0, 8)}...
                        </code>
                      ) : (
                        <span className="text-[var(--text-faint)]">—</span>
                      )}
                    </TableCell>
                    <TableCell>
                      {entry.success ? (
                        <Badge variant="success" className="gap-1">
                          <CheckCircle className="h-3 w-3" />
                          Success
                        </Badge>
                      ) : (
                        <Badge variant="error" className="gap-1">
                          <XCircle className="h-3 w-3" />
                          Failed
                        </Badge>
                      )}
                    </TableCell>
                    <TableCell>
                      <button
                        className="p-1 text-[var(--text-faint)] hover:text-[var(--text)]"
                        aria-label={isExpanded ? "Collapse audit log details" : "Expand audit log details"}
                        onClick={(e) => {
                          e.stopPropagation(); // Prevent row click from firing
                          toggleRow(entry.id);
                        }}
                      >
                        {isExpanded ? (
                          <ChevronUp className="h-4 w-4" />
                        ) : (
                          <ChevronDown className="h-4 w-4" />
                        )}
                      </button>
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>

          {/* Load more button */}
          {entriesArray && entriesArray.length >= limit && (
            <div className="p-4 border-t border-[var(--panel-edge)]-subtle text-center">
              <Button
                variant="outline"
                onClick={() => setLimit((prev) => prev + 50)}
              >
                Load More
              </Button>
            </div>
          )}
        </div>
      )}

      {/* Entry count */}
      <p className="text-sm text-[var(--text-faint)] text-center">
        Showing {filteredEntries?.length || 0} of {entriesArray?.length || 0} entries
      </p>
    </div>
  );
}
