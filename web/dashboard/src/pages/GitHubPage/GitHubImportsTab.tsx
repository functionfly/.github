import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  useGitHubImports,
  useCancelImport,
  useRetryImport,
  useResyncImport,
} from '@/hooks/useGitHubImport';
import type { GitHubImport } from '@/types/github';
import { motion } from 'framer-motion';
import {
  AlertCircle,
  CheckCircle,
  ChevronLeft,
  ChevronRight,
  Clock,
  Eye,
  FileCode,
  GitBranch,
  Loader2,
  MoreHorizontal,
  Play,
  RefreshCw,
  Search,
  Trash2,
  XCircle,
} from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';

const STATUS_FILTER_OPTIONS = [
  { value: 'all', label: 'All' },
  { value: 'completed', label: 'Completed' },
  { value: 'failed', label: 'Failed' },
  { value: 'in-progress', label: 'In Progress' },
] as const;

const IMPORT_STATUSES_IN_PROGRESS = ['pending', 'scanning', 'configuring', 'fetching', 'building', 'publishing'];

function getStatusBadgeVariant(status: GitHubImport['status']) {
  switch (status) {
    case 'completed':
      return 'default';
    case 'failed':
      return 'destructive';
    case 'cancelled':
      return 'outline';
    default:
      return 'secondary';
  }
}

function getStatusIcon(status: GitHubImport['status']) {
  switch (status) {
    case 'completed':
      return <CheckCircle className="w-3.5 h-3.5 text-green-500" />;
    case 'failed':
      return <XCircle className="w-3.5 h-3.5 text-red-500" />;
    case 'cancelled':
      return <XCircle className="w-3.5 h-3.5 text-text-muted" />;
    default:
      return <Loader2 className="w-3.5 h-3.5 text-blue-500 animate-spin" />;
  }
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return '-';
  return new Date(dateStr).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function GitHubImportsTab() {
  const navigate = useNavigate();
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [page, setPage] = useState(1);
  const perPage = 20;

  const importsParams = useMemo(() => {
    const params: Record<string, unknown> = { page, per_page: perPage };
    if (statusFilter !== 'all' && statusFilter !== 'in-progress') {
      params.status = statusFilter;
    }
    return params;
  }, [page, statusFilter]);

  const { data, isLoading, error } = useGitHubImports(importsParams as any);
  const cancelMutation = useCancelImport();
  const retryMutation = useRetryImport();
  const resyncMutation = useResyncImport();

  const imports = data?.imports ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.ceil(total / perPage);

  const filteredImports = useMemo(() => {
    let result = imports;
    if (statusFilter === 'in-progress') {
      result = result.filter((imp) => IMPORT_STATUSES_IN_PROGRESS.includes(imp.status));
    }
    if (searchQuery) {
      const q = searchQuery.toLowerCase();
      result = result.filter(
        (imp) =>
          imp.function_name.toLowerCase().includes(q) ||
          imp.source_path.toLowerCase().includes(q)
      );
    }
    return result;
  }, [imports, statusFilter, searchQuery]);

  useEffect(() => {
    setPage(1);
  }, [statusFilter]);

  if (error) {
    return (
      <div className="text-center py-12">
        <AlertCircle className="w-8 h-8 text-red-500 mx-auto mb-3" />
        <p className="text-text-secondary">Failed to load imports</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Toolbar */}
      <div className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
          <Input
            placeholder="Search by function name..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>

        <Select value={statusFilter} onValueChange={setStatusFilter}>
          <SelectTrigger className="w-[160px]">
            <SelectValue placeholder="Status" />
          </SelectTrigger>
          <SelectContent>
            {STATUS_FILTER_OPTIONS.map((opt) => (
              <SelectItem key={opt.value} value={opt.value}>
                {opt.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Loading */}
      {isLoading && (
        <div className="space-y-2">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-16 rounded-lg" />
          ))}
        </div>
      )}

      {/* Empty State */}
      {!isLoading && filteredImports.length === 0 && (
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-center py-16"
        >
          <FileCode className="w-12 h-12 text-text-muted mx-auto mb-4" />
          <h3 className="text-lg font-semibold text-text-primary mb-2">No imports yet</h3>
          <p className="text-text-secondary mb-4">
            Import functions from your GitHub repositories
          </p>
          <Button onClick={() => navigate('/github')} variant="outline">
            Browse Repositories
          </Button>
        </motion.div>
      )}

      {/* Table */}
      {!isLoading && filteredImports.length > 0 && (
        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }}>
          <div className="border border-border-default rounded-lg overflow-hidden">
            <table className="w-full">
              <thead>
                <tr className="bg-bg-secondary text-xs text-text-muted uppercase tracking-wider">
                  <th className="text-left px-4 py-3 font-medium">Repository</th>
                  <th className="text-left px-4 py-3 font-medium">Function</th>
                  <th className="text-left px-4 py-3 font-medium">Visibility</th>
                  <th className="text-left px-4 py-3 font-medium">Status</th>
                  <th className="text-left px-4 py-3 font-medium">Last Sync</th>
                  <th className="text-left px-4 py-3 font-medium">Created</th>
                  <th className="text-right px-4 py-3 font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredImports.map((imp) => (
                  <tr
                    key={imp.id}
                    className="border-t border-border-default hover:bg-bg-secondary/50 transition-colors"
                  >
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <GitBranch className="w-3.5 h-3.5 text-text-muted shrink-0" />
                        <span className="text-sm text-text-primary truncate max-w-[180px]">
                          {imp.source_path}
                        </span>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <span className="text-sm font-medium text-text-primary">
                        {imp.function_name}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <Badge
                        variant="outline"
                        className={
                          imp.visibility === 'public'
                            ? 'text-green-500 border-green-500/30'
                            : imp.visibility === 'private'
                              ? 'text-amber-500 border-amber-500/30'
                              : 'text-text-muted border-border-default'
                        }
                      >
                        {imp.visibility}
                      </Badge>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-1.5">
                        {getStatusIcon(imp.status)}
                        <Badge variant={getStatusBadgeVariant(imp.status)} className="text-xs">
                          {imp.status}
                        </Badge>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-sm text-text-muted">
                      {formatDate(imp.updated_at)}
                    </td>
                    <td className="px-4 py-3 text-sm text-text-muted">
                      {formatDate(imp.created_at)}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="icon" className="h-8 w-8">
                            <MoreHorizontal className="w-4 h-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          {(imp.function_author && imp.function_name) || imp.function_id ? (
                            <DropdownMenuItem
                              onClick={() =>
                                imp.function_author && imp.function_name
                                  ? navigate(`/functions/${imp.function_author}/${imp.function_name}`)
                                  : navigate(`/functions/${imp.function_id}`)
                              }
                            >
                              <Eye className="w-4 h-4 mr-2" />
                              View Function
                            </DropdownMenuItem>
                          ) : null}
                          <DropdownMenuItem
                            onClick={() => navigate(`/github?tab=imports&sync=${imp.id}`)}
                          >
                            <Clock className="w-4 h-4 mr-2" />
                            View Sync History
                          </DropdownMenuItem>
                          {imp.status === 'completed' && (
                            <DropdownMenuItem
                              onClick={() => resyncMutation.mutate(imp.id)}
                              disabled={resyncMutation.isPending}
                            >
                              <RefreshCw className="w-4 h-4 mr-2" />
                              Resync
                            </DropdownMenuItem>
                          )}
                          {IMPORT_STATUSES_IN_PROGRESS.includes(imp.status) && (
                            <DropdownMenuItem
                              onClick={() => cancelMutation.mutate(imp.id)}
                              disabled={cancelMutation.isPending}
                            >
                              <XCircle className="w-4 h-4 mr-2" />
                              Cancel
                            </DropdownMenuItem>
                          )}
                          {imp.status === 'failed' && (
                            <DropdownMenuItem
                              onClick={() => retryMutation.mutate(imp.id)}
                              disabled={retryMutation.isPending}
                            >
                              <Play className="w-4 h-4 mr-2" />
                              Retry
                            </DropdownMenuItem>
                          )}
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            onClick={() => {}}
                            className="text-red-500 focus:text-red-500"
                          >
                            <Trash2 className="w-4 h-4 mr-2" />
                            Delete
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </motion.div>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <span className="text-sm text-text-muted">
            Page {page} of {totalPages} ({total} imports)
          </span>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page <= 1}
            >
              <ChevronLeft className="w-4 h-4" />
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages}
            >
              <ChevronRight className="w-4 h-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
