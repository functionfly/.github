import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import {
  Search,
  Eye,
  Download,
  AlertTriangle,
  CheckCircle,
  XCircle,
  ArrowLeft,
  RefreshCw,
  ChevronLeft,
  ChevronRight,
  Shield,
  Clock,
  User,
  Filter,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Skeleton } from "@/components/ui/skeleton";
import { auditApi, type AuditEvent } from "@/api/admin";
import { cn } from "@/lib/utils";

const ACTION_COLORS: Record<string, string> = {
  create: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20",
  update: "bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/20",
  delete: "bg-red-500/10 text-red-600 dark:text-red-400 border-red-500/20",
  suspend: "bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20",
  login: "bg-purple-500/10 text-purple-600 dark:text-purple-400 border-purple-500/20",
  default: "bg-slate-500/10 text-slate-600 dark:text-slate-400 border-slate-500/20",
};

const RESOURCE_ICONS: Record<string, string> = {
  tenant: "🏢",
  user: "👤",
  billing: "💳",
  invoice: "💳",
  subscription: "💳",
  deployment: "🚀",
  function: "⚡",
  registry: "📦",
};

function getActionColor(action: string): string {
  for (const [key, color] of Object.entries(ACTION_COLORS)) {
    if (action.includes(key)) return color;
  }
  return ACTION_COLORS.default;
}

export function AdminAuditPage() {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState("");
  const [actionFilter, setActionFilter] = useState<string>("all");
  const [resourceFilter, setResourceFilter] = useState<string>("all");
  const [successFilter, setSuccessFilter] = useState<string>("all");
  const [selectedEvent, setSelectedEvent] = useState<AuditEvent | null>(null);
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);
  const [limit] = useState(50);
  const [offset, setOffset] = useState(0);

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['audit-events', limit, offset, actionFilter, resourceFilter, successFilter],
    queryFn: () => auditApi.listAuditEvents({
      limit,
      offset,
      action: actionFilter !== "all" ? actionFilter : undefined,
      resource_type: resourceFilter !== "all" ? resourceFilter : undefined,
      success: successFilter === "success" ? true : successFilter === "failure" ? false : undefined,
    }),
  });

  const events = data?.events || [];
  const total = data?.total || 0;

  const filteredEvents = events.filter((event) => {
    if (!searchTerm) return true;
    const lower = searchTerm.toLowerCase();
    return (
      event.actor_email?.toLowerCase().includes(lower) ||
      event.action.toLowerCase().includes(lower) ||
      event.resource_type.toLowerCase().includes(lower)
    );
  });

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    return {
      date: date.toLocaleDateString(),
      time: date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
    };
  };

  const handleViewDetails = (event: AuditEvent) => {
    setSelectedEvent(event);
    setIsDetailsOpen(true);
  };

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-center">
          <AlertTriangle className="w-12 h-12 text-red-500 mx-auto mb-4" />
          <p className="text-text-primary font-medium">Failed to load audit events</p>
          <p className="text-text-secondary text-sm mt-1">Please try again later</p>
          <Button onClick={() => refetch()} className="mt-4 bg-brand-500 hover:bg-brand-600 text-white">
            Retry
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <Button
            variant="ghost"
            onClick={() => navigate('/admin')}
            className="text-text-muted hover:text-text-primary hover:bg-bg-hover"
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back
          </Button>
          <div>
            <h1 className="text-2xl font-bold text-text-primary">Audit Log</h1>
            <p className="text-sm text-text-secondary">Monitor all admin actions and system events</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => refetch()}
            className="border-border-default hover:bg-bg-hover text-text-secondary"
          >
            <RefreshCw className="w-4 h-4" />
          </Button>
          <Button
            variant="outline"
            size="sm"
            className="border-border-default hover:bg-bg-hover text-text-secondary"
          >
            <Download className="w-4 h-4 mr-2" />
            Export
          </Button>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {[
          { label: "Total Events", value: total, icon: Shield, color: "text-blue-500", bg: "bg-blue-500/10" },
          { label: "Successful", value: events.filter(e => e.success).length, icon: CheckCircle, color: "text-emerald-500", bg: "bg-emerald-500/10" },
          { label: "Failed", value: events.filter(e => !e.success).length, icon: XCircle, color: "text-red-500", bg: "bg-red-500/10" },
          { label: "Unique Actors", value: new Set(events.map(e => e.actor_email)).size, icon: User, color: "text-purple-500", bg: "bg-purple-500/10" },
        ].map((stat) => (
          <Card key={stat.label} className="glass-card">
            <CardContent className="p-4">
              <div className="flex items-center justify-between mb-1">
                <p className="text-xs text-text-secondary">{stat.label}</p>
                <div className={cn("p-1.5 rounded-lg", stat.bg)}>
                  <stat.icon className={cn("w-3.5 h-3.5", stat.color)} />
                </div>
              </div>
              <p className="text-xl font-bold text-text-primary">{isLoading ? "—" : stat.value}</p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Filters */}
      <Card className="glass-card">
        <CardContent className="p-4">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-3">
            <div className="lg:col-span-2 relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
              <Input
                placeholder="Search by actor, action, or resource..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10 bg-bg-secondary border-border-default text-text-primary"
              />
            </div>

            <Select value={actionFilter} onValueChange={setActionFilter}>
              <SelectTrigger className="bg-bg-secondary border-border-default text-text-primary">
                <SelectValue placeholder="Action" />
              </SelectTrigger>
              <SelectContent className="bg-bg-tertiary border-border-default">
                <SelectItem value="all">All Actions</SelectItem>
                <SelectItem value="tenant.update">Tenant Update</SelectItem>
                <SelectItem value="user.create">User Create</SelectItem>
                <SelectItem value="user.update">User Update</SelectItem>
                <SelectItem value="user.delete">User Delete</SelectItem>
                <SelectItem value="billing.invoice.paid">Invoice Paid</SelectItem>
              </SelectContent>
            </Select>

            <Select value={resourceFilter} onValueChange={setResourceFilter}>
              <SelectTrigger className="bg-bg-secondary border-border-default text-text-primary">
                <SelectValue placeholder="Resource" />
              </SelectTrigger>
              <SelectContent className="bg-bg-tertiary border-border-default">
                <SelectItem value="all">All Resources</SelectItem>
                <SelectItem value="tenant">Tenant</SelectItem>
                <SelectItem value="user">User</SelectItem>
                <SelectItem value="billing">Billing</SelectItem>
                <SelectItem value="deployment">Deployment</SelectItem>
                <SelectItem value="function">Function</SelectItem>
              </SelectContent>
            </Select>

            <Select value={successFilter} onValueChange={setSuccessFilter}>
              <SelectTrigger className="bg-bg-secondary border-border-default text-text-primary">
                <SelectValue placeholder="Status" />
              </SelectTrigger>
              <SelectContent className="bg-bg-tertiary border-border-default">
                <SelectItem value="all">All Status</SelectItem>
                <SelectItem value="success">Success</SelectItem>
                <SelectItem value="failure">Failure</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* Audit Events */}
      <Card className="glass-card">
        <CardHeader className="pb-3">
          <CardTitle className="text-text-primary">
            Audit Events
            <span className="ml-2 text-sm font-normal text-text-muted">
              ({isLoading ? "..." : filteredEvents.length} shown)
            </span>
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {/* Table Header */}
          <div className="hidden lg:grid grid-cols-[1fr_1.5fr_1.5fr_1fr_80px_60px] gap-4 px-6 py-3 border-b border-border-subtle text-xs font-semibold text-text-muted uppercase tracking-wider">
            <span>Time</span>
            <span>Actor</span>
            <span>Action</span>
            <span>Resource</span>
            <span>Status</span>
            <span>Details</span>
          </div>

          <div className="divide-y divide-border-subtle">
            {isLoading ? (
              Array.from({ length: 8 }).map((_, i) => (
                <div key={i} className="flex items-center gap-4 px-6 py-4">
                  <Skeleton className="h-4 w-24" />
                  <Skeleton className="h-4 w-32" />
                  <Skeleton className="h-6 w-28" />
                  <Skeleton className="h-4 w-20" />
                  <Skeleton className="h-5 w-5 rounded-full" />
                </div>
              ))
            ) : filteredEvents.length === 0 ? (
              <div className="text-center py-12 text-text-muted">
                <Shield className="w-10 h-10 mx-auto mb-3 opacity-30" />
                <p className="font-medium">No audit events found</p>
                <p className="text-sm mt-1">Try adjusting your filters</p>
              </div>
            ) : (
              filteredEvents.map((event) => {
                const { date, time } = formatTimestamp(event.timestamp);
                return (
                  <div
                    key={event.id}
                    className="flex flex-col lg:grid lg:grid-cols-[1fr_1.5fr_1.5fr_1fr_80px_60px] gap-2 lg:gap-4 px-6 py-4 hover:bg-bg-hover transition-colors"
                  >
                    {/* Time */}
                    <div className="flex items-center gap-2">
                      <Clock className="w-3.5 h-3.5 text-text-muted flex-shrink-0" />
                      <div>
                        <p className="text-xs font-medium text-text-primary">{time}</p>
                        <p className="text-xs text-text-muted">{date}</p>
                      </div>
                    </div>

                    {/* Actor */}
                    <div className="flex items-center gap-2">
                      <div className="w-6 h-6 rounded-full bg-brand-500/10 flex items-center justify-center flex-shrink-0">
                        <User className="w-3 h-3 text-brand-500" />
                      </div>
                      <div className="min-w-0">
                        <p className="text-sm text-text-primary truncate">{event.actor_email || "System"}</p>
                        {event.ip_address && (
                          <p className="text-xs text-text-muted font-mono">{event.ip_address}</p>
                        )}
                      </div>
                    </div>

                    {/* Action */}
                    <div className="flex items-center">
                      <Badge className={cn("text-xs border font-medium", getActionColor(event.action))}>
                        {event.action}
                      </Badge>
                    </div>

                    {/* Resource */}
                    <div className="flex items-center gap-2">
                      <span className="text-base">{RESOURCE_ICONS[event.resource_type] || "📝"}</span>
                      <div className="min-w-0">
                        <p className="text-sm text-text-primary capitalize">{event.resource_type}</p>
                        {event.resource_id && (
                          <p className="text-xs text-text-muted font-mono">{event.resource_id.slice(-8)}</p>
                        )}
                      </div>
                    </div>

                    {/* Status */}
                    <div className="flex items-center">
                      {event.success ? (
                        <div className="flex items-center gap-1.5">
                          <CheckCircle className="w-4 h-4 text-emerald-500" />
                          <span className="text-xs text-emerald-600 dark:text-emerald-400 hidden lg:block">OK</span>
                        </div>
                      ) : (
                        <div className="flex items-center gap-1.5">
                          <XCircle className="w-4 h-4 text-red-500" />
                          <span className="text-xs text-red-600 dark:text-red-400 hidden lg:block">Fail</span>
                        </div>
                      )}
                    </div>

                    {/* Details */}
                    <div className="flex items-center">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleViewDetails(event)}
                        className="text-text-muted hover:text-text-primary h-8 w-8 p-0"
                      >
                        <Eye className="w-4 h-4" />
                      </Button>
                    </div>
                  </div>
                );
              })
            )}
          </div>

          {/* Pagination */}
          {!isLoading && total > limit && (
            <div className="flex items-center justify-between px-6 py-4 border-t border-border-subtle">
              <p className="text-sm text-text-muted">
                Showing {offset + 1}–{Math.min(offset + limit, total)} of {total}
              </p>
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setOffset(Math.max(0, offset - limit))}
                  disabled={offset === 0}
                  className="border-border-default hover:bg-bg-hover text-text-secondary"
                >
                  <ChevronLeft className="w-4 h-4" />
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setOffset(offset + limit)}
                  disabled={offset + limit >= total}
                  className="border-border-default hover:bg-bg-hover text-text-secondary"
                >
                  <ChevronRight className="w-4 h-4" />
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Event Details Dialog */}
      <Dialog open={isDetailsOpen} onOpenChange={setIsDetailsOpen}>
        <DialogContent className="bg-bg-tertiary border-border-default max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="text-text-primary flex items-center gap-2">
              <Shield className="w-5 h-5 text-brand-500" />
              Audit Event Details
            </DialogTitle>
          </DialogHeader>
          {selectedEvent && (
            <div className="space-y-4">
              {/* Status Banner */}
              <div className={cn(
                "flex items-center gap-3 p-3 rounded-lg border",
                selectedEvent.success
                  ? "bg-emerald-500/10 border-emerald-500/20"
                  : "bg-red-500/10 border-red-500/20"
              )}>
                {selectedEvent.success ? (
                  <CheckCircle className="w-5 h-5 text-emerald-500" />
                ) : (
                  <XCircle className="w-5 h-5 text-red-500" />
                )}
                <span className={cn(
                  "font-medium",
                  selectedEvent.success ? "text-emerald-600 dark:text-emerald-400" : "text-red-600 dark:text-red-400"
                )}>
                  {selectedEvent.success ? "Action completed successfully" : "Action failed"}
                </span>
              </div>

              <div className="grid grid-cols-2 gap-4">
                {[
                  { label: "Event ID", value: selectedEvent.id, mono: true },
                  { label: "Timestamp", value: new Date(selectedEvent.timestamp).toLocaleString(), mono: false },
                  { label: "Actor", value: selectedEvent.actor_email || "System", mono: false },
                  { label: "Action", value: selectedEvent.action, mono: true },
                  { label: "Resource Type", value: selectedEvent.resource_type, mono: false },
                  { label: "Resource ID", value: selectedEvent.resource_id || "N/A", mono: true },
                  { label: "IP Address", value: selectedEvent.ip_address || "N/A", mono: true },
                  { label: "Tenant ID", value: selectedEvent.tenant_id || "N/A", mono: true },
                ].map((field) => (
                  <div key={field.label} className="space-y-1">
                    <p className="text-xs font-medium text-text-muted uppercase tracking-wider">{field.label}</p>
                    <p className={cn(
                      "text-sm text-text-primary bg-bg-secondary px-2 py-1.5 rounded border border-border-subtle break-all",
                      field.mono && "font-mono"
                    )}>
                      {field.value}
                    </p>
                  </div>
                ))}
              </div>

              {selectedEvent.beforeState && (
                <div className="space-y-1">
                  <p className="text-xs font-medium text-text-muted uppercase tracking-wider">Before State</p>
                  <pre className="text-xs text-text-primary bg-bg-secondary p-3 rounded border border-border-subtle overflow-x-auto">
                    {JSON.stringify(selectedEvent.beforeState, null, 2)}
                  </pre>
                </div>
              )}

              {selectedEvent.afterState && (
                <div className="space-y-1">
                  <p className="text-xs font-medium text-text-muted uppercase tracking-wider">After State</p>
                  <pre className="text-xs text-text-primary bg-bg-secondary p-3 rounded border border-border-subtle overflow-x-auto">
                    {JSON.stringify(selectedEvent.afterState, null, 2)}
                  </pre>
                </div>
              )}

              <div className="flex justify-end">
                <Button variant="outline" onClick={() => setIsDetailsOpen(false)} className="border-border-default">
                  Close
                </Button>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
