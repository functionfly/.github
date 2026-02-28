import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { Search, Filter, Eye, Download, AlertTriangle, CheckCircle, XCircle, ArrowLeft } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { LoadingSpinner } from "@/components/common/LoadingSpinner";
import { auditApi, type AuditEvent } from "@/api/admin";

export function AdminAuditPage() {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState("");
  const [actionFilter, setActionFilter] = useState<string>("all");
  const [resourceFilter, setResourceFilter] = useState<string>("all");
  const [successFilter, setSuccessFilter] = useState<string>("all");
  const [selectedEvent, setSelectedEvent] = useState<AuditEvent | null>(null);
  const [limit] = useState(50);
  const [offset, setOffset] = useState(0);

  const { data, isLoading, error } = useQuery({
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

  const filteredEvents = events.filter((event) => {
    const matchesSearch =
      event.actor_email?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      event.action.toLowerCase().includes(searchTerm.toLowerCase()) ||
      event.resource_type.toLowerCase().includes(searchTerm.toLowerCase());

    return matchesSearch;
  });

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString();
  };

  const getActionBadgeColor = (action: string) => {
    if (action.includes("create")) return "bg-emerald-500/10 text-emerald-400";
    if (action.includes("update")) return "bg-blue-500/10 text-blue-400";
    if (action.includes("delete")) return "bg-red-500/10 text-red-400";
    if (action.includes("suspend")) return "bg-amber-500/10 text-amber-400";
    return "bg-gray-500/10 text-text-secondary";
  };

  const getResourceIcon = (resourceType: string) => {
    switch (resourceType) {
      case "tenant":
        return "🏢";
      case "user":
        return "👤";
      case "billing":
      case "invoice":
      case "subscription":
        return "💳";
      case "deployment":
        return "🚀";
      default:
        return "📝";
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <LoadingSpinner />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-center">
          <AlertTriangle className="w-12 h-12 text-red-400 mx-auto mb-4" />
          <p className="text-text-primary font-medium">Failed to load audit events</p>
          <p className="text-text-secondary">Please try again later</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <Button
          variant="ghost"
          onClick={() => navigate('/admin')}
          className="text-text-muted hover:text-text-primary hover:bg-bg-hover"
        >
          <ArrowLeft className="w-4 h-4 mr-2" />
          Back to Dashboard
        </Button>
        <div className="flex-1 text-center">
          <h1 className="text-2xl font-bold text-text-primary">Audit Log</h1>
          <p className="text-text-secondary">Monitor all admin actions and system events</p>
        </div>
        <Button variant="outline" className="border-border-subtle hover:bg-bg-hover">
          <Download className="w-4 h-4 mr-2" />
          Export
        </Button>
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="pt-6">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
            <div className="lg:col-span-2">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-text-muted" />
                <Input
                  placeholder="Search audit events..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-10 bg-bg-secondary border-white/8 text-text-primary"
                />
              </div>
            </div>

            <Select value={actionFilter} onValueChange={setActionFilter}>
              <SelectTrigger className="bg-bg-secondary border-white/8 text-text-primary">
                <SelectValue placeholder="Action" />
              </SelectTrigger>
              <SelectContent className="bg-[#1a1a25] border-white/8">
                <SelectItem value="all">All Actions</SelectItem>
                <SelectItem value="tenant.update">Tenant Update</SelectItem>
                <SelectItem value="user.create">User Create</SelectItem>
                <SelectItem value="billing.invoice.paid">Invoice Paid</SelectItem>
              </SelectContent>
            </Select>

            <Select value={resourceFilter} onValueChange={setResourceFilter}>
              <SelectTrigger className="bg-bg-secondary border-white/8 text-text-primary">
                <SelectValue placeholder="Resource" />
              </SelectTrigger>
              <SelectContent className="bg-[#1a1a25] border-white/8">
                <SelectItem value="all">All Resources</SelectItem>
                <SelectItem value="tenant">Tenant</SelectItem>
                <SelectItem value="user">User</SelectItem>
                <SelectItem value="billing">Billing</SelectItem>
                <SelectItem value="deployment">Deployment</SelectItem>
              </SelectContent>
            </Select>

            <Select value={successFilter} onValueChange={setSuccessFilter}>
              <SelectTrigger className="bg-bg-secondary border-white/8 text-text-primary">
                <SelectValue placeholder="Status" />
              </SelectTrigger>
              <SelectContent className="bg-[#1a1a25] border-white/8">
                <SelectItem value="all">All Status</SelectItem>
                <SelectItem value="success">Success</SelectItem>
                <SelectItem value="failure">Failure</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* Audit Events Table */}
      <Card>
        <CardHeader>
          <CardTitle>Audit Events ({filteredEvents.length})</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow className="border-white/8">
                  <TableHead className="text-text-secondary">Time</TableHead>
                  <TableHead className="text-text-secondary">Actor</TableHead>
                  <TableHead className="text-text-secondary">Action</TableHead>
                  <TableHead className="text-text-secondary">Resource</TableHead>
                  <TableHead className="text-text-secondary">Status</TableHead>
                  <TableHead className="text-text-secondary">Details</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredEvents.map((event) => (
                  <TableRow key={event.id} className="border-white/8">
                    <TableCell className="text-text-primary">
                      {formatTimestamp(event.timestamp)}
                    </TableCell>
                    <TableCell className="text-text-primary">
                      <div>
                        <div className="font-medium">{event.actor_email || "System"}</div>
                        <div className="text-xs text-text-muted">
                          {event.ip_address || "N/A"}
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge className={getActionBadgeColor(event.action)}>
                        {event.action}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-text-primary">
                      <div className="flex items-center gap-2">
                        <span>{getResourceIcon(event.resource_type)}</span>
                        <span>{event.resource_type}</span>
                        {event.resource_id && (
                          <span className="text-xs text-text-muted">
                            {event.resource_id.slice(-8)}
                          </span>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      {event.success ? (
                        <CheckCircle className="w-4 h-4 text-emerald-400" />
                      ) : (
                        <XCircle className="w-4 h-4 text-red-400" />
                      )}
                    </TableCell>
                    <TableCell>
                      <Dialog>
                        <DialogTrigger asChild>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setSelectedEvent(event)}
                            className="text-text-secondary hover:text-text-primary"
                          >
                            <Eye className="w-4 h-4" />
                          </Button>
                        </DialogTrigger>
                        <DialogContent className="bg-[#1a1a25] border-white/8 max-w-2xl">
                          <DialogHeader>
                            <DialogTitle className="text-text-primary">Audit Event Details</DialogTitle>
                          </DialogHeader>
                          <div className="space-y-4">
                            <div className="grid grid-cols-2 gap-4">
                              <div>
                                <label className="text-sm font-medium text-text-secondary">Event ID</label>
                                <p className="text-text-primary font-mono">{event.id}</p>
                              </div>
                              <div>
                                <label className="text-sm font-medium text-text-secondary">Timestamp</label>
                                <p className="text-text-primary">{formatTimestamp(event.timestamp)}</p>
                              </div>
                              <div>
                                <label className="text-sm font-medium text-text-secondary">Actor</label>
                                <p className="text-text-primary">{event.actor_email || "System"}</p>
                              </div>
                              <div>
                                <label className="text-sm font-medium text-text-secondary">Action</label>
                                <p className="text-text-primary">{event.action}</p>
                              </div>
                              <div>
                                <label className="text-sm font-medium text-text-secondary">Resource</label>
                                <p className="text-text-primary">{event.resource_type} {event.resource_id?.slice(-8)}</p>
                              </div>
                              <div>
                                <label className="text-sm font-medium text-text-secondary">IP Address</label>
                                <p className="text-text-primary font-mono">{event.ip_address || "N/A"}</p>
                              </div>
                            </div>

                            {event.beforeState && (
                              <div>
                                <label className="text-sm font-medium text-text-secondary">Before State</label>
                                <pre className="text-xs text-text-primary bg-bg-secondary p-2 rounded mt-1 overflow-x-auto">
                                  {JSON.stringify(event.beforeState, null, 2)}
                                </pre>
                              </div>
                            )}

                            {event.afterState && (
                              <div>
                                <label className="text-sm font-medium text-text-secondary">After State</label>
                                <pre className="text-xs text-text-primary bg-bg-secondary p-2 rounded mt-1 overflow-x-auto">
                                  {JSON.stringify(event.afterState, null, 2)}
                                </pre>
                              </div>
                            )}
                          </div>
                        </DialogContent>
                      </Dialog>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
