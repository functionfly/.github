/**
 * SecretAuditDrawer - Slide-out drawer for audit logs
 *
 * Displays a comprehensive audit trail in a slide-out drawer format with
 * real-time updates indicator, export options (CSV, JSON), filtering by
 * action type, user, date range, and search within audit entries.
 * Supports pagination and includes empty and loading states.
 *
 * @example
 * ```tsx
 * // Basic usage
 * <SecretAuditDrawer
 *   isOpen={isDrawerOpen}
 *   onClose={() => setIsDrawerOpen(false)}
 *   secretId="secret-123"
 * />
 *
 * // With vault-wide audit logs
 * <SecretAuditDrawer
 *   isOpen={isOpen}
 *   onClose={onClose}
 *   title="Vault Activity"
 * />
 *
 * // Loading state
 * <SecretAuditDrawer isOpen={isOpen} isLoading onClose={onClose} />
 * ```
 */

import { useState, useMemo, useCallback } from "react";
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
  X,
  FileJson,
  FileSpreadsheet,
  Radio,
  Calendar,
} from "lucide-react";
import { format, formatDistanceToNow, isWithinInterval, parseISO } from "date-fns";
import { cn } from "@/lib/utils";
import type { AuditAction, AuditLogEntry, ActorType } from "@/types/vault";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Separator } from "@/components/ui/separator";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
  SheetFooter,
} from "@/components/ui/sheet";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Calendar as CalendarComponent } from "@/components/ui/calendar";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

/** Export format options */
export type ExportFormat = "csv" | "json";

/** Date range filter */
export interface DateRange {
  from?: Date;
  to?: Date;
}

export interface SecretAuditDrawerProps {
  /** Whether the drawer is open */
  isOpen: boolean;
  /** Callback when drawer closes */
  onClose: () => void;
  /** Secret ID to filter audit logs (optional - shows vault-wide if not provided) */
  secretId?: string;
  /** Custom title for the drawer */
  title?: string;
  /** Audit log entries (controlled mode) */
  entries?: AuditLogEntry[];
  /** Whether the component is loading */
  isLoading?: boolean;
  /** Error message if loading failed */
  error?: Error | null;
  /** Whether real-time updates are enabled */
  isLive?: boolean;
  /** Callback when refresh is requested */
  onRefresh?: () => void;
  /** Callback when export is requested */
  onExport?: (format: ExportFormat, entries: AuditLogEntry[]) => void;
  /** Number of entries per page */
  pageSize?: number;
  /** Additional CSS classes */
  className?: string;
}

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

// Action color mapping for icons
const actionColors: Record<AuditAction, string> = {
  create: "text-success bg-success-glow",
  read: "text-brand-500 bg-brand-500/10",
  update: "text-warning bg-warning-glow",
  delete: "text-error bg-error-glow",
  use: "text-brand-500 bg-brand-500/10",
  revoke: "text-error bg-error-glow",
  version: "text-warning bg-warning-glow",
  rollback: "text-error bg-error-glow",
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

/**
 * Default empty audit entries for demo/loading states
 */
const defaultEntries: AuditLogEntry[] = [];

/**
 * SecretAuditDrawer component
 */
export function SecretAuditDrawer({
  isOpen,
  onClose,
  secretId,
  title = secretId ? "Secret Audit Log" : "Vault Audit Log",
  entries: controlledEntries,
  isLoading = false,
  error = null,
  isLive = false,
  onRefresh,
  onExport,
  pageSize = 25,
  className,
}: SecretAuditDrawerProps) {
  const [searchQuery, setSearchQuery] = useState("");
  const [actionFilter, setActionFilter] = useState<AuditAction | "all">("all");
  const [actorFilter, setActorFilter] = useState<ActorType | "all">("all");
  const [dateRange, setDateRange] = useState<DateRange>({});
  const [currentPage, setCurrentPage] = useState(1);
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());
  const [isExportMenuOpen, setIsExportMenuOpen] = useState(false);

  // Use controlled entries or default empty array
  const entries = controlledEntries ?? defaultEntries;

  // Filter entries
  const filteredEntries = useMemo(() => {
    return entries.filter((entry) => {
      // Action filter
      if (actionFilter !== "all" && entry.action !== actionFilter) {
        return false;
      }

      // Actor filter
      if (actorFilter !== "all" && entry.actor_type !== actorFilter) {
        return false;
      }

      // Date range filter
      if (dateRange.from || dateRange.to) {
        const entryDate = parseISO(entry.created_at);
        const from = dateRange.from;
        const to = dateRange.to;

        if (from && to) {
          if (!isWithinInterval(entryDate, { start: from, end: to })) {
            return false;
          }
        } else if (from && entryDate < from) {
          return false;
        } else if (to && entryDate > to) {
          return false;
        }
      }

      // Search filter
      if (searchQuery) {
        const query = searchQuery.toLowerCase();
        const matchesSearch =
          entry.actor_id.toLowerCase().includes(query) ||
          entry.secret_id?.toLowerCase().includes(query) ||
          entry.request_id?.toLowerCase().includes(query) ||
          entry.error_message?.toLowerCase().includes(query) ||
          entry.ip_address?.toLowerCase().includes(query);

        if (!matchesSearch) return false;
      }

      return true;
    });
  }, [entries, actionFilter, actorFilter, dateRange, searchQuery]);

  // Paginated entries
  const paginatedEntries = useMemo(() => {
    const start = (currentPage - 1) * pageSize;
    return filteredEntries.slice(start, start + pageSize);
  }, [filteredEntries, currentPage, pageSize]);

  // Total pages
  const totalPages = Math.ceil(filteredEntries.length / pageSize);

  // Toggle row expansion
  const toggleRow = useCallback((id: string) => {
    setExpandedRows((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(id)) {
        newSet.delete(id);
      } else {
        newSet.add(id);
      }
      return newSet;
    });
  }, []);

  // Export to CSV
  const exportToCSV = useCallback(() => {
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
      "Request ID",
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
          entry.request_id || "",
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

    onExport?.("csv", filteredEntries);
    setIsExportMenuOpen(false);
  }, [filteredEntries, onExport]);

  // Export to JSON
  const exportToJSON = useCallback(() => {
    if (!filteredEntries.length) return;

    const jsonContent = JSON.stringify(filteredEntries, null, 2);
    const blob = new Blob([jsonContent], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `audit-log-${format(new Date(), "yyyy-MM-dd")}.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);

    onExport?.("json", filteredEntries);
    setIsExportMenuOpen(false);
  }, [filteredEntries, onExport]);

  // Clear all filters
  const clearFilters = useCallback(() => {
    setSearchQuery("");
    setActionFilter("all");
    setActorFilter("all");
    setDateRange({});
    setCurrentPage(1);
  }, []);

  // Check if any filters are active
  const hasActiveFilters =
    searchQuery || actionFilter !== "all" || actorFilter !== "all" || dateRange.from || dateRange.to;

  // Reset page when filters change
  useMemo(() => {
    setCurrentPage(1);
  }, [searchQuery, actionFilter, actorFilter, dateRange]);

  return (
    <Sheet open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <SheetContent className={cn("w-full sm:max-w-2xl flex flex-col", className)}>
        <SheetHeader className="space-y-2">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <SheetTitle>{title}</SheetTitle>
              {isLive && (
                <Badge variant="success" className="gap-1 animate-pulse">
                  <Radio className="h-3 w-3" />
                  Live
                </Badge>
              )}
            </div>
            <div className="flex items-center gap-2">
              {onRefresh && (
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={onRefresh}
                  disabled={isLoading}
                  aria-label="Refresh audit log"
                >
                  <Loader2
                    className={cn("h-4 w-4", isLoading && "animate-spin")}
                  />
                </Button>
              )}
              <Button variant="ghost" size="icon" onClick={onClose} aria-label="Close audit drawer">
                <X className="h-4 w-4" />
              </Button>
            </div>
          </div>
          <SheetDescription>
            {secretId
              ? `View audit trail for secret ${secretId.slice(0, 8)}...`
              : "View complete vault audit trail with filtering and export options."}
          </SheetDescription>
        </SheetHeader>

        <Separator className="my-4" />

        {/* Filters */}
        <div className="space-y-4 shrink-0">
          <div className="flex flex-col gap-3">
            {/* Search */}
            <div className="relative">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-text-muted" />
              <Input
                placeholder="Search by actor, secret ID, IP..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10"
              />
              {searchQuery && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="absolute right-1 top-1/2 -translate-y-1/2 h-6 w-6"
                  onClick={() => setSearchQuery("")}
                >
                  <X className="h-3 w-3" />
                </Button>
              )}
            </div>

            {/* Filter row */}
            <div className="flex flex-wrap gap-2">
              <Select
                value={actionFilter}
                onValueChange={(value) => setActionFilter(value as AuditAction | "all")}
              >
                <SelectTrigger className="w-[130px]">
                  <Filter className="h-4 w-4 mr-2" />
                  <SelectValue placeholder="Action" />
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

              <Select
                value={actorFilter}
                onValueChange={(value) => setActorFilter(value as ActorType | "all")}
              >
                <SelectTrigger className="w-[130px]">
                  <User className="h-4 w-4 mr-2" />
                  <SelectValue placeholder="Actor" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Actors</SelectItem>
                  <SelectItem value="user">User</SelectItem>
                  <SelectItem value="token">Token</SelectItem>
                  <SelectItem value="system">System</SelectItem>
                  <SelectItem value="api_key">API Key</SelectItem>
                </SelectContent>
              </Select>

              <Popover>
                <PopoverTrigger asChild>
                  <Button variant="outline" className="gap-2">
                    <Calendar className="h-4 w-4" />
                    {dateRange.from && dateRange.to
                      ? `${format(dateRange.from, "MMM d")} - ${format(dateRange.to, "MMM d")}`
                      : dateRange.from
                      ? `From ${format(dateRange.from, "MMM d")}`
                      : dateRange.to
                      ? `Until ${format(dateRange.to, "MMM d")}`
                      : "Date Range"}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0" align="start">
                  <CalendarComponent
                    initialFocus
                    mode="range"
                    defaultMonth={dateRange.from}
                    selected={{
                      from: dateRange.from,
                      to: dateRange.to,
                    }}
                    onSelect={(range) =>
                      setDateRange({
                        from: range?.from,
                        to: range?.to,
                      })
                    }
                    numberOfMonths={2}
                  />
                </PopoverContent>
              </Popover>

              {hasActiveFilters && (
                <Button variant="ghost" size="sm" onClick={clearFilters}>
                  Clear filters
                </Button>
              )}
            </div>
          </div>
        </div>

        <Separator className="my-4" />

        {/* Content area */}
        <div className="flex-1 overflow-y-auto min-h-0">
          {isLoading ? (
            // Loading state
            <div className="space-y-3">
              {Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="flex items-center gap-4 p-3 rounded-lg border">
                  <Skeleton className="h-10 w-10 rounded-lg" />
                  <div className="flex-1 space-y-2">
                    <Skeleton className="h-4 w-32" />
                    <Skeleton className="h-3 w-48" />
                  </div>
                  <Skeleton className="h-6 w-16" />
                </div>
              ))}
            </div>
          ) : error ? (
            // Error state
            <div className="rounded-lg border border-error/20 bg-error-glow p-8 text-center">
              <AlertTriangle className="mx-auto h-12 w-12 text-error mb-4" />
              <h3 className="text-lg font-semibold text-error mb-2">
                Failed to load audit log
              </h3>
              <p className="text-text-secondary mb-4">
                {error.message || "An unexpected error occurred"}
              </p>
              <Button onClick={onRefresh} variant="outline">
                Retry
              </Button>
            </div>
          ) : filteredEntries.length === 0 ? (
            // Empty state
            <div className="rounded-lg border border-border-subtle bg-card p-12 text-center">
              <div className="mx-auto h-16 w-16 rounded-full bg-gradient-to-br from-brand-500/20 to-purple-500/20 flex items-center justify-center mb-4">
                <Shield className="h-8 w-8 text-brand-500" />
              </div>
              <h3 className="text-lg font-semibold text-card-foreground mb-2">
                No audit entries found
              </h3>
              <p className="text-text-secondary max-w-sm mx-auto">
                {hasActiveFilters
                  ? "No entries match your current filters. Try adjusting your search criteria."
                  : "Audit log entries will appear here when secrets are created, accessed, or modified."}
              </p>
            </div>
          ) : (
            // Audit entries list
            <div className="space-y-2">
              {paginatedEntries.map((entry) => {
                const ActionIcon = actionIcons[entry.action];
                const ActorIcon = actorTypeIcons[entry.actor_type];
                const isExpanded = expandedRows.has(entry.id);

                return (
                  <div
                    key={entry.id}
                    className={cn(
                      "rounded-lg border border-border-subtle p-3 transition-colors",
                      "hover:border-border-hover cursor-pointer",
                      !entry.success && "bg-error-glow/20 border-error/30",
                      isExpanded && "border-border-hover"
                    )}
                    onClick={() => toggleRow(entry.id)}
                  >
                    <div className="flex items-start gap-3">
                      {/* Action icon */}
                      <div
                        className={cn(
                          "h-10 w-10 rounded-lg flex items-center justify-center shrink-0",
                          actionColors[entry.action]
                        )}
                      >
                        <ActionIcon className="h-5 w-5" />
                      </div>

                      {/* Main content */}
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center justify-between gap-2">
                          <div className="flex items-center gap-2 flex-wrap">
                            <Badge variant={actionVariants[entry.action]}>
                              {actionLabels[entry.action]}
                            </Badge>
                            <span className="text-sm text-text-muted">
                              {formatDistanceToNow(new Date(entry.created_at), {
                                addSuffix: true,
                              })}
                            </span>
                          </div>
                          {entry.success ? (
                            <CheckCircle className="h-4 w-4 text-success shrink-0" />
                          ) : (
                            <XCircle className="h-4 w-4 text-error shrink-0" />
                          )}
                        </div>

                        <div className="mt-2 flex items-center gap-3 text-sm">
                          <div className="flex items-center gap-1.5 text-text-secondary">
                            <ActorIcon className="h-3.5 w-3.5" />
                            <span className="truncate max-w-[120px]" title={entry.actor_id}>
                              {entry.actor_id.slice(0, 12)}...
                            </span>
                            <Badge variant="outline" className="text-[10px] px-1 py-0">
                              {actorTypeLabels[entry.actor_type]}
                            </Badge>
                          </div>
                          {entry.secret_id && (
                            <div className="flex items-center gap-1.5 text-text-secondary">
                              <Key className="h-3.5 w-3.5" />
                              <code className="text-xs bg-bg-secondary px-1.5 py-0.5 rounded">
                                {entry.secret_id.slice(0, 8)}...
                              </code>
                            </div>
                          )}
                        </div>

                        {/* Expanded details */}
                        {isExpanded && (
                          <div className="mt-3 pt-3 border-t border-border-subtle space-y-2">
                            <div className="grid grid-cols-2 gap-2 text-xs">
                              <div>
                                <span className="text-text-muted">Timestamp:</span>
                                <p className="text-text-primary">
                                  {format(new Date(entry.created_at), "MMM d, yyyy HH:mm:ss")}
                                </p>
                              </div>
                              <div>
                                <span className="text-text-muted">Request ID:</span>
                                <p className="text-text-primary font-mono">
                                  {entry.request_id || "—"}
                                </p>
                              </div>
                              {entry.ip_address && (
                                <div>
                                  <span className="text-text-muted">IP Address:</span>
                                  <p className="text-text-primary">{entry.ip_address}</p>
                                </div>
                              )}
                              {entry.user_agent && (
                                <div className="col-span-2">
                                  <span className="text-text-muted">User Agent:</span>
                                  <p className="text-text-primary truncate">{entry.user_agent}</p>
                                </div>
                              )}
                            </div>
                            {entry.error_message && (
                              <div className="rounded bg-error-glow/30 p-2 text-xs text-error">
                                <strong>Error:</strong> {entry.error_message}
                              </div>
                            )}
                            {entry.metadata && Object.keys(entry.metadata).length > 0 && (
                              <div>
                                <span className="text-text-muted text-xs">Metadata:</span>
                                <pre className="mt-1 text-xs bg-bg-secondary p-2 rounded overflow-x-auto">
                                  {JSON.stringify(entry.metadata, null, 2)}
                                </pre>
                              </div>
                            )}
                          </div>
                        )}
                      </div>

                      {/* Expand indicator */}
                      <div className="shrink-0 text-text-muted">
                        {isExpanded ? (
                          <ChevronUp className="h-4 w-4" />
                        ) : (
                          <ChevronDown className="h-4 w-4" />
                        )}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>

        <Separator className="my-4" />

        {/* Footer with pagination and export */}
        <SheetFooter className="flex-col sm:flex-row gap-4 sm:justify-between shrink-0">
          <div className="flex items-center gap-2 text-sm text-text-muted">
            <span>
              Showing {filteredEntries.length > 0 ? (currentPage - 1) * pageSize + 1 : 0} -{" "}
              {Math.min(currentPage * pageSize, filteredEntries.length)} of{" "}
              {filteredEntries.length}
            </span>
            {entries.length !== filteredEntries.length && (
              <span className="text-text-secondary">
                (filtered from {entries.length})
              </span>
            )}
          </div>

          <div className="flex items-center gap-2">
            {/* Pagination */}
            {totalPages > 1 && (
              <div className="flex items-center gap-1">
                <Button
                  variant="outline"
                  size="icon"
                  className="h-8 w-8"
                  onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                  disabled={currentPage === 1}
                  aria-label="Previous page"
                >
                  <ChevronDown className="h-4 w-4 rotate-90" />
                </Button>
                <span className="text-sm text-text-secondary px-2">
                  {currentPage} / {totalPages}
                </span>
                <Button
                  variant="outline"
                  size="icon"
                  className="h-8 w-8"
                  onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
                  disabled={currentPage === totalPages}
                  aria-label="Next page"
                >
                  <ChevronDown className="h-4 w-4 -rotate-90" />
                </Button>
              </div>
            )}

            {/* Export dropdown */}
            <Popover open={isExportMenuOpen} onOpenChange={setIsExportMenuOpen}>
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  disabled={filteredEntries.length === 0}
                  className="gap-2"
                >
                  <Download className="h-4 w-4" />
                  Export
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-40" align="end">
                <div className="space-y-1">
                  <Button
                    variant="ghost"
                    className="w-full justify-start gap-2"
                    onClick={exportToCSV}
                  >
                    <FileSpreadsheet className="h-4 w-4" />
                    Export CSV
                  </Button>
                  <Button
                    variant="ghost"
                    className="w-full justify-start gap-2"
                    onClick={exportToJSON}
                  >
                    <FileJson className="h-4 w-4" />
                    Export JSON
                  </Button>
                </div>
              </PopoverContent>
            </Popover>
          </div>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
}
