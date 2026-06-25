import { useState } from 'react';
import { PageLayout } from '@/components/layout/PageLayout';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import { usePlan } from '@/hooks/usePlan';
import {
  useEnterpriseAuditLogs,
  useEnterpriseAuditFilters,
  useDownloadEnterpriseAuditExport,
  type AuditLogParams,
} from '@/hooks/useEnterpriseAudit';
import {
  Download,
  FileText,
  Filter,
  Search,
  ChevronLeft,
  ChevronRight,
  CheckCircle,
  XCircle,
  Clock,
  AlertTriangle,
} from 'lucide-react';
import { useNavigate } from 'react-router-dom';

function formatTimestamp(timestamp: string): string {
  const date = new Date(timestamp);
  return date.toLocaleString();
}

function formatRelativeTime(timestamp: string): string {
  const now = new Date();
  const date = new Date(timestamp);
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffMins < 1) return 'just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;
  return date.toLocaleDateString();
}

function StatusBadge({ success }: { success: boolean }) {
  if (success) {
    return (
      <Badge variant="outline" className="text-green-400 border-green-400/50 gap-1">
        <CheckCircle className="w-3 h-3" />
        Success
      </Badge>
    );
  }
  return (
    <Badge variant="outline" className="text-red-400 border-red-400/50 gap-1">
      <XCircle className="w-3 h-3" />
      Failed
    </Badge>
  );
}

function AuditLogRow({ log }: { log: ReturnType<typeof useEnterpriseAuditLogs>['data']['logs'][number] }) {
  return (
    <tr className="border-b border-white/5 hover:bg-white/5 transition-colors">
      <td className="py-3 px-4">
        <div className="flex flex-col">
          <span className="text-sm text-white">{formatRelativeTime(log.created_at)}</span>
          <span className="text-xs text-text-muted">{formatTimestamp(log.created_at)}</span>
        </div>
      </td>
      <td className="py-3 px-4">
        <div className="flex flex-col">
          <span className="text-sm text-white">{log.actor_name || 'System'}</span>
          <span className="text-xs text-text-muted">
            {log.actor_type} {log.actor_id && `(${log.actor_id.slice(0, 8)}...)`}
          </span>
        </div>
      </td>
      <td className="py-3 px-4">
        <Badge variant="outline" className="text-amber-400 border-amber-400/50">
          {log.action}
        </Badge>
      </td>
      <td className="py-3 px-4">
        <div className="flex flex-col">
          <span className="text-sm text-white">{log.resource_type}</span>
          <span className="text-xs text-text-muted">
            {log.service_area}
            {log.resource_id && ` • ${log.resource_id.slice(0, 8)}...`}
          </span>
        </div>
      </td>
      <td className="py-3 px-4">
        <StatusBadge success={log.success} />
      </td>
    </tr>
  );
}

function AuditLogTableSkeleton() {
  return (
    <>
      {[...Array(5)].map((_, i) => (
        <tr key={i} className="border-b border-white/5">
          <td className="py-3 px-4">
            <Skeleton className="h-10 w-32" />
          </td>
          <td className="py-3 px-4">
            <Skeleton className="h-10 w-40" />
          </td>
          <td className="py-3 px-4">
            <Skeleton className="h-6 w-20" />
          </td>
          <td className="py-3 px-4">
            <Skeleton className="h-10 w-40" />
          </td>
          <td className="py-3 px-4">
            <Skeleton className="h-6 w-16" />
          </td>
        </tr>
      ))}
    </>
  );
}

export function EnterpriseAuditPage() {
  const { isEnterprise } = usePlan();
  const navigate = useNavigate();

  const [params, setParams] = useState<AuditLogParams>({
    limit: 20,
    offset: 0,
  });
  const [searchQuery, setSearchQuery] = useState('');
  const [showFilters, setShowFilters] = useState(false);
  const [filters, setFilters] = useState<{
    service_area?: string;
    action?: string;
    resource_type?: string;
    success?: boolean;
  }>({});

  const { data: logsData, isLoading, error } = useEnterpriseAuditLogs(params);
  const { data: filtersData } = useEnterpriseAuditFilters();
  const downloadExport = useDownloadEnterpriseAuditExport();

  const handleSearch = () => {
    setParams((prev) => ({
      ...prev,
      search: searchQuery || undefined,
      offset: 0,
    }));
  };

  const handleFilterChange = (key: string, value: string | boolean | undefined) => {
    setFilters((prev) => ({
      ...prev,
      [key]: value || undefined,
    }));
  };

  const applyFilters = () => {
    setParams((prev) => ({
      ...prev,
      ...filters,
      offset: 0,
    }));
  };

  const handleExport = async (format: 'json' | 'csv' | 'cef') => {
    const filename = `enterprise-audit-${new Date().toISOString().split('T')[0]}.${format}`;
    await downloadExport(
      {
        from: params.start_time,
        to: params.end_time,
        format,
        service_area: filters.service_area,
        action: filters.action,
      },
      filename,
    );
  };

  const handlePageChange = (newOffset: number) => {
    setParams((prev) => ({
      ...prev,
      offset: newOffset,
    }));
  };

  const totalPages = logsData ? Math.ceil(logsData.total / (params.limit || 20)) : 0;
  const currentPage = Math.floor((params.offset || 0) / (params.limit || 20)) + 1;

  if (!isEnterprise) {
    return (
      <PageLayout title="Audit Logs">
        <Card className="border-dashed border-white/20">
          <CardContent className="flex flex-col items-center justify-center py-16 text-center">
            <div className="w-16 h-16 rounded-full bg-amber-500/10 flex items-center justify-center mb-4">
              <FileText className="w-8 h-8 text-amber-400" />
            </div>
            <h2 className="text-xl font-semibold text-white mb-2">Enterprise Feature</h2>
            <p className="text-text-secondary mb-6 max-w-md">
              Audit logs are available exclusively for Enterprise plan customers. Upgrade to access
              detailed audit trails and compliance reporting.
            </p>
            <Button
              onClick={() => navigate('/pricing')}
              className="bg-gradient-to-r from-amber-500 to-yellow-500"
            >
              View Enterprise Plans
            </Button>
          </CardContent>
        </Card>
      </PageLayout>
    );
  }

  return (
    <PageLayout title="Audit Logs">
      <p className="text-text-secondary mb-6">
        View and export audit trails for compliance and security monitoring
      </p>

      <div className="space-y-6">
        <Card>
          <CardContent className="p-4">
            <div className="flex flex-col sm:flex-row gap-4">
              <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
                <Input
                  placeholder="Search audit logs..."
                  className="pl-10"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
                />
              </div>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  className="gap-2"
                  onClick={() => setShowFilters(!showFilters)}
                >
                  <Filter className="w-4 h-4" />
                  Filter
                </Button>
                <Select onValueChange={(v) => handleExport(v as 'json' | 'csv' | 'cef')}>
                  <SelectTrigger className="w-[120px]">
                    <Download className="w-4 h-4 mr-2" />
                    <SelectValue placeholder="Export" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="json">JSON</SelectItem>
                    <SelectItem value="csv">CSV</SelectItem>
                    <SelectItem value="cef">CEF</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            {showFilters && (
              <div className="mt-4 pt-4 border-t border-white/10 grid grid-cols-2 md:grid-cols-4 gap-4">
                <Select
                  value={filters.service_area}
                  onValueChange={(v) => handleFilterChange('service_area', v)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Service Area" />
                  </SelectTrigger>
                  <SelectContent>
                    {filtersData?.service_areas?.map((sa) => (
                      <SelectItem key={sa} value={sa}>
                        {sa}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>

                <Select
                  value={filters.action}
                  onValueChange={(v) => handleFilterChange('action', v)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Action" />
                  </SelectTrigger>
                  <SelectContent>
                    {filtersData?.actions?.map((action) => (
                      <SelectItem key={action} value={action}>
                        {action}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>

                <Select
                  value={filters.resource_type}
                  onValueChange={(v) => handleFilterChange('resource_type', v)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Resource Type" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="function">Function</SelectItem>
                    <SelectItem value="secret">Secret</SelectItem>
                    <SelectItem value="user">User</SelectItem>
                    <SelectItem value="api_key">API Key</SelectItem>
                    <SelectItem value="team">Team</SelectItem>
                  </SelectContent>
                </Select>

                <Select
                  value={filters.success === undefined ? 'all' : filters.success ? 'true' : 'false'}
                  onValueChange={(v) =>
                    handleFilterChange('success', v === 'all' ? undefined : v === 'true')
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Status" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All</SelectItem>
                    <SelectItem value="true">Success</SelectItem>
                    <SelectItem value="false">Failed</SelectItem>
                  </SelectContent>
                </Select>

                <div className="col-span-2 md:col-span-4 flex justify-end gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setFilters({});
                      setParams((prev) => ({
                        ...prev,
                        service_area: undefined,
                        action: undefined,
                        resource_type: undefined,
                        success: undefined,
                        offset: 0,
                      }));
                    }}
                  >
                    Clear Filters
                  </Button>
                  <Button size="sm" onClick={applyFilters}>
                    Apply Filters
                  </Button>
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="text-white">
                Audit Logs
                {logsData && (
                  <span className="text-text-muted text-sm font-normal ml-2">
                    ({logsData.total.toLocaleString()} total)
                  </span>
                )}
              </CardTitle>
            </div>
          </CardHeader>
          <CardContent>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-white/10">
                    <th className="text-left py-3 px-4 text-sm font-medium text-text-secondary">
                      Timestamp
                    </th>
                    <th className="text-left py-3 px-4 text-sm font-medium text-text-secondary">
                      User
                    </th>
                    <th className="text-left py-3 px-4 text-sm font-medium text-text-secondary">
                      Action
                    </th>
                    <th className="text-left py-3 px-4 text-sm font-medium text-text-secondary">
                      Resource
                    </th>
                    <th className="text-left py-3 px-4 text-sm font-medium text-text-secondary">
                      Status
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {isLoading && <AuditLogTableSkeleton />}
                  {error && (
                    <tr>
                      <td colSpan={5} className="py-8 text-center text-red-400">
                        <AlertTriangle className="w-8 h-8 mx-auto mb-2" />
                        <p>Failed to load audit logs</p>
                      </td>
                    </tr>
                  )}
                  {!isLoading && !error && logsData?.logs.length === 0 && (
                    <tr>
                      <td colSpan={5} className="py-16 text-center">
                        <div className="flex flex-col items-center justify-center">
                          <Clock className="w-12 h-12 text-text-muted mb-4" />
                          <h3 className="text-lg font-semibold text-white mb-2">
                            No Audit Logs Found
                          </h3>
                          <p className="text-text-secondary max-w-md">
                            {Object.values(filters).some(Boolean)
                              ? 'No logs match your current filters. Try adjusting your search criteria.'
                              : 'No audit activity has been recorded yet.'}
                          </p>
                        </div>
                      </td>
                    </tr>
                  )}
                  {!isLoading &&
                    !error &&
                    logsData?.logs.map((log) => <AuditLogRow key={log.id} log={log} />)}
                </tbody>
              </table>
            </div>

            {logsData && logsData.total > (params.limit || 20) && (
              <div className="mt-4 flex items-center justify-between">
                <span className="text-sm text-text-muted">
                  Showing {(params.offset || 0) + 1} -{' '}
                  {Math.min((params.offset || 0) + (params.limit || 20), logsData.total)} of{' '}
                  {logsData.total.toLocaleString()}
                </span>
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handlePageChange((params.offset || 0) - (params.limit || 20))}
                    disabled={(params.offset || 0) === 0}
                  >
                    <ChevronLeft className="w-4 h-4" />
                    Previous
                  </Button>
                  <span className="text-sm text-text-secondary">
                    Page {currentPage} of {totalPages}
                  </span>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handlePageChange((params.offset || 0) + (params.limit || 20))}
                    disabled={(params.offset || 0) + (params.limit || 20) >= logsData.total}
                  >
                    Next
                    <ChevronRight className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </PageLayout>
  );
}
