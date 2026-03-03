import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import {
  Search,
  Eye,
  ArrowLeft,
  RefreshCw,
  ChevronLeft,
  ChevronRight,
  Clock,
  Copy,
  CheckCircle,
  XCircle,
  AlertCircle,
  Filter,
  Calendar,
  Hash,
  Building2,
  FunctionSquare,
  Fingerprint,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { toast } from "sonner";
import { oversightApi, type ExecutionAuditData, type ExecutionRecord } from "@/api/admin";

const mockTenants = [
  "All Tenants",
  "acme-corp",
  "startup-xyz",
  "data-analytics-co",
  "dev-studio-x",
  "ai-solutions-ltd",
  "doc-tools-inc",
  "cloud-services-pro",
  "enterprise-solutions",
];

// Helper functions
const getStatusBadge = (status: string) => {
  switch (status) {
    case "success":
      return (
        <Badge className="bg-emerald-500/10 text-emerald-400 border-emerald-500/20 flex items-center gap-1">
          <CheckCircle className="w-3 h-3" />
          Success
        </Badge>
      );
    case "error":
      return (
        <Badge className="bg-red-500/10 text-red-400 border-red-500/20 flex items-center gap-1">
          <XCircle className="w-3 h-3" />
          Error
        </Badge>
      );
    case "timeout":
      return (
        <Badge className="bg-amber-500/10 text-amber-400 border-amber-500/20 flex items-center gap-1">
          <AlertCircle className="w-3 h-3" />
          Timeout
        </Badge>
      );
    default:
      return <Badge variant="secondary">{status}</Badge>;
  }
};

const truncateHash = (hash: string, length: number = 16) => {
  if (hash.length <= length * 2 + 4) return hash;
  return `${hash.slice(0, length)}...${hash.slice(-length)}`;
};

const formatBytes = (bytes: number) => {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`;
};

const formatDuration = (ms: number) => {
  if (ms < 1000) return `${ms}ms`;
  return `${(ms / 1000).toFixed(2)}s`;
};

export function AdminExecutionAuditPage() {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState("");
  const [tenantFilter, setTenantFilter] = useState("All Tenants");
  const [statusFilter, setStatusFilter] = useState("all");
  const [hashFilter, setHashFilter] = useState("");
  const [dateRange, setDateRange] = useState<{ from?: string; to?: string }>({});
  const [currentPage, setCurrentPage] = useState(1);
  const [selectedExecution, setSelectedExecution] = useState<ExecutionRecord | null>(null);
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);
  const pageSize = 10;

  const { data, isLoading, refetch } = useQuery<ExecutionAuditData>({
    queryKey: ["admin-execution-audit", currentPage, tenantFilter, statusFilter, hashFilter, searchTerm],
    queryFn: () => oversightApi.getExecutionAudit({
      page: currentPage,
      pageSize,
      search: searchTerm || undefined,
      tenant: tenantFilter !== "All Tenants" ? tenantFilter : undefined,
      status: statusFilter !== "all" ? statusFilter : undefined,
      hash: hashFilter || undefined,
    }),
  });

  const handleCopyHash = (hash: string) => {
    navigator.clipboard.writeText(hash);
    toast.success("Hash copied to clipboard");
  };

  const handleViewDetails = (execution: ExecutionRecord) => {
    setSelectedExecution(execution);
    setIsDetailsOpen(true);
  };

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    return {
      date: date.toLocaleDateString(),
      time: date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" }),
    };
  };

  const totalPages = data ? Math.ceil(data.total / pageSize) : 0;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <Button
            variant="ghost"
            onClick={() => navigate("/admin")}
            className="text-text-muted hover:text-text-primary hover:bg-bg-hover"
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back
          </Button>
          <div>
            <h1 className="text-2xl font-bold text-text-primary">Execution Audit Console</h1>
            <p className="text-sm text-text-secondary">Searchable audit trail for all function executions</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => refetch()}
            className="border-border-default hover:bg-bg-hover text-text-secondary"
          >
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </Button>
        </div>
      </div>

      {/* Search & Filters */}
      <Card className="glass-card">
        <CardHeader>
          <CardTitle className="text-text-primary flex items-center gap-2">
            <Filter className="w-5 h-5 text-brand-500" />
            Search & Filters
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Search Bar */}
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
            <Input
              placeholder="Search by function name, tenant, or any execution field..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-10 bg-bg-secondary border-border-default text-text-primary"
            />
          </div>

          {/* Advanced Filters */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            <div className="relative">
              <Hash className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
              <Input
                placeholder="Execution Root Hash..."
                value={hashFilter}
                onChange={(e) => setHashFilter(e.target.value)}
                className="pl-10 bg-bg-secondary border-border-default text-text-primary"
              />
            </div>

            <div className="relative">
              <Building2 className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
              <Select value={tenantFilter} onValueChange={setTenantFilter}>
                <SelectTrigger className="pl-10 bg-bg-secondary border-border-default text-text-primary">
                  <SelectValue placeholder="Select tenant" />
                </SelectTrigger>
                <SelectContent className="bg-bg-tertiary border-border-default">
                  {mockTenants.map((tenant) => (
                    <SelectItem key={tenant} value={tenant}>
                      {tenant}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="relative">
              <FunctionSquare className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
              <Input
                placeholder="Function name..."
                className="pl-10 bg-bg-secondary border-border-default text-text-primary"
              />
            </div>

            <div className="relative">
              <Select value={statusFilter} onValueChange={setStatusFilter}>
                <SelectTrigger className="bg-bg-secondary border-border-default text-text-primary">
                  <SelectValue placeholder="Status" />
                </SelectTrigger>
                <SelectContent className="bg-bg-tertiary border-border-default">
                  <SelectItem value="all">All Status</SelectItem>
                  <SelectItem value="success">Success</SelectItem>
                  <SelectItem value="error">Error</SelectItem>
                  <SelectItem value="timeout">Timeout</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* Date Range */}
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <Calendar className="w-4 h-4 text-text-muted" />
              <span className="text-sm text-text-secondary">Date Range:</span>
            </div>
            <div className="flex items-center gap-2">
              <Input
                type="date"
                value={dateRange.from || ""}
                onChange={(e) => setDateRange((prev) => ({ ...prev, from: e.target.value }))}
                className="bg-bg-secondary border-border-default text-text-primary w-auto"
              />
              <span className="text-text-muted">to</span>
              <Input
                type="date"
                value={dateRange.to || ""}
                onChange={(e) => setDateRange((prev) => ({ ...prev, to: e.target.value }))}
                className="bg-bg-secondary border-border-default text-text-primary w-auto"
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Results Table */}
      <Card className="glass-card">
        <CardHeader>
          <CardTitle className="text-text-primary flex items-center gap-2">
            <Fingerprint className="w-5 h-5 text-brand-500" />
            Execution Results
            <span className="text-sm font-normal text-text-muted">
              ({isLoading ? "..." : data?.total || 0} total)
            </span>
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="p-6 space-y-4">
              {Array.from({ length: 5 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : data?.executions.length === 0 ? (
            <div className="text-center py-12">
              <Search className="w-12 h-12 text-text-muted mx-auto mb-4" />
              <h3 className="text-lg font-semibold text-text-primary mb-2">No executions found</h3>
              <p className="text-text-secondary">Try adjusting your search or filters</p>
            </div>
          ) : (
            <>
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow className="border-border-subtle hover:bg-transparent">
                      <TableHead className="text-text-secondary">Execution Hash</TableHead>
                      <TableHead className="text-text-secondary">Tenant</TableHead>
                      <TableHead className="text-text-secondary">Function</TableHead>
                      <TableHead className="text-text-secondary">Timestamp</TableHead>
                      <TableHead className="text-text-secondary">Node Signature</TableHead>
                      <TableHead className="text-text-secondary">Status</TableHead>
                      <TableHead className="text-text-secondary">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {data?.executions.map((execution) => {
                      const timestamp = formatTimestamp(execution.timestamp);
                      return (
                        <TableRow key={execution.id} className="border-border-subtle">
                          <TableCell>
                            <div className="flex items-center gap-2">
                              <code className="text-xs font-mono text-text-secondary">
                                {truncateHash(execution.executionRootHash, 12)}
                              </code>
                              <Button
                                variant="ghost"
                                size="sm"
                                className="h-6 w-6 p-0 text-text-muted hover:text-text-primary"
                                onClick={() => handleCopyHash(execution.executionRootHash)}
                              >
                                <Copy className="w-3 h-3" />
                              </Button>
                            </div>
                          </TableCell>
                          <TableCell>
                            <Badge variant="secondary" className="text-xs">
                              {execution.tenant}
                            </Badge>
                          </TableCell>
                          <TableCell className="font-medium text-text-primary">
                            {execution.functionName}
                          </TableCell>
                          <TableCell>
                            <div className="flex items-center gap-1 text-sm text-text-secondary">
                              <Clock className="w-3 h-3" />
                              <span>{timestamp.date}</span>
                              <span className="text-text-muted">{timestamp.time}</span>
                            </div>
                          </TableCell>
                          <TableCell>
                            <code className="text-xs font-mono text-text-muted">
                              {truncateHash(execution.nodeSignature, 8)}
                            </code>
                          </TableCell>
                          <TableCell>{getStatusBadge(execution.status)}</TableCell>
                          <TableCell>
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-8 text-text-muted hover:text-text-primary"
                              onClick={() => handleViewDetails(execution)}
                            >
                              <Eye className="w-4 h-4 mr-1" />
                              View
                            </Button>
                          </TableCell>
                        </TableRow>
                      );
                    })}
                  </TableBody>
                </Table>
              </div>

              {/* Pagination */}
              <div className="flex items-center justify-between px-6 py-4 border-t border-border-subtle">
                <div className="text-sm text-text-muted">
                  Showing {(currentPage - 1) * pageSize + 1} to{" "}
                  {Math.min(currentPage * pageSize, data?.total || 0)} of {data?.total || 0} results
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                    disabled={currentPage === 1}
                    className="border-border-default hover:bg-bg-hover text-text-secondary"
                  >
                    <ChevronLeft className="w-4 h-4" />
                  </Button>
                  <span className="text-sm text-text-secondary">
                    Page {currentPage} of {totalPages}
                  </span>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
                    disabled={currentPage >= totalPages}
                    className="border-border-default hover:bg-bg-hover text-text-secondary"
                  >
                    <ChevronRight className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {/* Execution Details Dialog */}
      <Dialog open={isDetailsOpen} onOpenChange={setIsDetailsOpen}>
        <DialogContent className="max-w-2xl bg-bg-tertiary border-border-default">
          <DialogHeader>
            <DialogTitle className="text-text-primary flex items-center gap-2">
              <Fingerprint className="w-5 h-5 text-brand-500" />
              Execution Details
            </DialogTitle>
          </DialogHeader>
          {selectedExecution && (
            <div className="space-y-6">
              {/* Status & Basic Info */}
              <div className="flex items-center justify-between">
                <div>{getStatusBadge(selectedExecution.status)}</div>
                <div className="text-sm text-text-muted">
                  {formatTimestamp(selectedExecution.timestamp).date}{" "}
                  {formatTimestamp(selectedExecution.timestamp).time}
                </div>
              </div>

              {/* Execution Hash */}
              <div className="p-4 rounded-lg bg-bg-secondary">
                <p className="text-xs text-text-muted mb-1">Execution Root Hash</p>
                <div className="flex items-center justify-between">
                  <code className="text-sm font-mono text-text-primary break-all">
                    {selectedExecution.executionRootHash}
                  </code>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleCopyHash(selectedExecution.executionRootHash)}
                  >
                    <Copy className="w-4 h-4" />
                  </Button>
                </div>
              </div>

              {/* Details Grid */}
              <div className="grid grid-cols-2 gap-4">
                <div className="p-3 rounded-lg bg-bg-secondary">
                  <p className="text-xs text-text-muted mb-1">Tenant</p>
                  <p className="font-medium text-text-primary">{selectedExecution.tenant}</p>
                </div>
                <div className="p-3 rounded-lg bg-bg-secondary">
                  <p className="text-xs text-text-muted mb-1">Function</p>
                  <p className="font-medium text-text-primary">{selectedExecution.functionName}</p>
                </div>
                <div className="p-3 rounded-lg bg-bg-secondary">
                  <p className="text-xs text-text-muted mb-1">Duration</p>
                  <p className="font-medium text-text-primary">{formatDuration(selectedExecution.duration)}</p>
                </div>
                <div className="p-3 rounded-lg bg-bg-secondary">
                  <p className="text-xs text-text-muted mb-1">Input Size</p>
                  <p className="font-medium text-text-primary">{formatBytes(selectedExecution.inputSize)}</p>
                </div>
              </div>

              {/* Node Signature */}
              <div className="p-3 rounded-lg bg-bg-secondary">
                <p className="text-xs text-text-muted mb-1">Node Signature</p>
                <code className="text-sm font-mono text-text-secondary">
                  {selectedExecution.nodeSignature}
                </code>
              </div>

              {/* Error Message (if applicable) */}
              {selectedExecution.errorMessage && (
                <div className="p-4 rounded-lg bg-red-500/10 border border-red-500/20">
                  <p className="text-xs text-red-400 mb-1 flex items-center gap-1">
                    <AlertCircle className="w-3 h-3" />
                    Error Message
                  </p>
                  <p className="text-sm text-text-primary">{selectedExecution.errorMessage}</p>
                </div>
              )}
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
