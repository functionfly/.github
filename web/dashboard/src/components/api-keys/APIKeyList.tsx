import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { API_KEY_TYPE_LABELS, APIKey, APIKeyFilters, APIKeyType } from '@/types/api-key';
import { useQuery } from '@tanstack/react-query';
import { teamsApi } from '@/api/teams';
import {
  AlertCircle,
  CheckSquare,
  ChevronLeft,
  ChevronRight,
  Filter,
  Key,
  RefreshCw,
  RotateCcw,
  Search,
  Square,
  Trash2,
} from 'lucide-react';
import { useCallback, useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import { toast } from 'sonner';

interface APIKeyListProps {
  apiKeys: APIKey[];
  isLoading?: boolean;
  isError?: boolean;
  error?: Error | null;
  onRetry?: () => void;
  total?: number;
  page?: number;
  pageSize?: number;
  filters?: APIKeyFilters;
  onFiltersChange?: (filters: APIKeyFilters) => void;
  onPageChange?: (page: number) => void;
  onCreateNew?: () => void;
  onRotate?: (key: APIKey) => void;
  onDelete?: (key: APIKey) => void;
  // Bulk operations
  selectedKeys?: Set<string>;
  onSelectionChange?: (selected: Set<string>) => void;
  onBulkDelete?: (ids: string[]) => void | Promise<void>;
  onBulkRotate?: (ids: string[]) => void | Promise<void>;
}

const SEARCH_DEBOUNCE_MS = 300;

export function APIKeyList({
  apiKeys,
  isLoading = false,
  isError = false,
  error = null,
  onRetry,
  total = 0,
  page = 1,
  pageSize = 10,
  filters,
  onFiltersChange,
  onPageChange,
  onCreateNew,
  onRotate,
  onDelete,
  selectedKeys,
  onSelectionChange,
  onBulkDelete,
  onBulkRotate,
}: APIKeyListProps) {
  const [searchValue, setSearchValue] = useState(filters?.search || '');
  const [localSelected, setLocalSelected] = useState<Set<string>>(new Set());
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const [bulkAction, setBulkAction] = useState<'delete' | 'rotate' | null>(null);

  // Keep local state in sync if upstream filters change (e.g. on reset).
  useEffect(() => {
    setSearchValue(filters?.search || '');
  }, [filters?.search]);

  useEffect(() => {
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
  }, []);

  // Clear selection when page changes so stale selections don't persist.
  useEffect(() => {
    setLocalSelected(new Set());
    setBulkAction(null);
  }, [page, filters]);

  // Effective selected set (controlled or internal).
  const activeSelected = selectedKeys ?? localSelected;
  const setActiveSelected = onSelectionChange ?? setLocalSelected;

  const allCurrentIds = useCallback(() => new Set(apiKeys.map((k) => k.id)), [apiKeys]);

  const isAllSelected = apiKeys.length > 0 && activeSelected.size === apiKeys.length;
  const isSomeSelected = activeSelected.size > 0 && !isAllSelected;

  const toggleOne = (id: string) => {
    const next = new Set(activeSelected);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    setActiveSelected(next);
  };

  const toggleAll = () => {
    if (isAllSelected) {
      setActiveSelected(new Set());
      return;
    }
    setActiveSelected(allCurrentIds());
  };

  const handleSearch = useCallback(
    (value: string) => {
      setSearchValue(value);
      if (debounceRef.current) clearTimeout(debounceRef.current);
      debounceRef.current = setTimeout(() => {
        onFiltersChange?.({ ...filters, search: value, page: 1 });
      }, SEARCH_DEBOUNCE_MS);
    },
    [filters, onFiltersChange]
  );

  const handleTypeFilter = (value: string) => {
    const newFilters: APIKeyFilters = {
      ...filters,
      key_type: value === 'all' ? undefined : (value as APIKeyType),
      page: 1,
    };
    onFiltersChange?.(newFilters);
  };

  const handleStatusFilter = (value: string) => {
    const newFilters: APIKeyFilters = {
      ...filters,
      is_active: value === 'all' ? undefined : value === 'active',
      page: 1,
    };
    onFiltersChange?.(newFilters);
  };

  const { data: teamsData } = useQuery({ queryKey: ['teams'], queryFn: () => teamsApi.list() });
  const teams = teamsData?.teams ?? [];

  const handleTeamFilter = (value: string) => {
    const newFilters: APIKeyFilters = {
      ...filters,
      team_id: value === 'all' ? undefined : value,
      page: 1,
    };
    onFiltersChange?.(newFilters);
  };

  const handlePageChange = (newPage: number) => {
    onPageChange?.(newPage);
  };

  const executeBulkDelete = async () => {
    if (!onBulkDelete || activeSelected.size === 0) return;
    setBulkAction('delete');
    try {
      await onBulkDelete(Array.from(activeSelected));
      setActiveSelected(new Set());
      toast.success('Selected API keys deleted');
    } catch {
      toast.error('Some keys could not be deleted');
    } finally {
      setBulkAction(null);
    }
  };

  const executeBulkRotate = async () => {
    if (!onBulkRotate || activeSelected.size === 0) return;
    setBulkAction('rotate');
    try {
      await onBulkRotate(Array.from(activeSelected));
      setActiveSelected(new Set());
      toast.success('Selected API keys rotated');
    } catch {
      toast.error('Some keys could not be rotated');
    } finally {
      setBulkAction(null);
    }
  };

  // Forward bulk actions to parent if provided via props; otherwise fall back to
  // a noop so the UI still renders.
  const handleBulkDelete = async () => {
    if (onBulkDelete) {
      await executeBulkDelete();
      return;
    }
    // Fallback: iterate single-delete callback if present.
    if (onDelete && activeSelected.size > 0) {
      setBulkAction('delete');
      for (const id of Array.from(activeSelected)) {
        const key = apiKeys.find((k) => k.id === id);
        if (key) onDelete(key);
      }
      setActiveSelected(new Set());
      setBulkAction(null);
    }
  };

  const handleBulkRotate = async () => {
    if (onBulkRotate) {
      await executeBulkRotate();
      return;
    }
    if (onRotate && activeSelected.size > 0) {
      setBulkAction('rotate');
      // Rotate one at a time to respect per-key rate limits; pause briefly
      // between calls so the user-session limiter is not tripped.
      for (const id of Array.from(activeSelected)) {
        const key = apiKeys.find((k) => k.id === id);
        if (key) onRotate(key);
        await new Promise((r) => setTimeout(r, 350));
      }
      setActiveSelected(new Set());
      setBulkAction(null);
    }
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return 'Never';
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getKeyTypeBadgeVariant = (type: string) => {
    switch (type) {
      case 'platform':
        return 'default';
      case 'function':
        return 'secondary';
      default:
        return 'outline';
    }
  };

  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  if (isError) {
    return (
      <div
        role="alert"
        className="flex flex-col items-center justify-center gap-3 py-12 border border-red-200 dark:border-red-800 rounded-lg bg-red-50 dark:bg-red-950"
      >
        <AlertCircle className="w-10 h-10 text-red-600 dark:text-red-400" />
        <h3 className="text-lg font-semibold text-red-700 dark:text-red-300">
          Failed to load API keys
        </h3>
        <p className="text-sm text-red-600 dark:text-red-400 max-w-md text-center">
          {error?.message || 'An unexpected error occurred.'}
        </p>
        {onRetry && (
          <Button variant="outline" onClick={onRetry} className="mt-2">
            <RefreshCw className="w-4 h-4 mr-2" />
            Try again
          </Button>
        )}
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Bulk action toolbar */}
      {activeSelected.size > 0 && (
        <div className="flex items-center justify-between rounded-lg border bg-muted/40 p-2">
          <span className="text-sm font-medium px-2">{activeSelected.size} selected</span>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={handleBulkRotate}
              disabled={bulkAction !== null}
            >
              <RotateCcw className="w-4 h-4 mr-2" />
              {bulkAction === 'rotate' ? 'Rotating…' : 'Rotate Selected'}
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="text-red-600 hover:text-red-600"
              onClick={handleBulkDelete}
              disabled={bulkAction !== null}
            >
              <Trash2 className="w-4 h-4 mr-2" />
              {bulkAction === 'delete' ? 'Deleting…' : 'Delete Selected'}
            </Button>
            <Button variant="ghost" size="sm" onClick={() => setActiveSelected(new Set())}>
              Clear
            </Button>
          </div>
        </div>
      )}

      {/* Toolbar */}
      <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
        <div className="flex flex-1 gap-4 w-full sm:w-auto">
          <div className="relative flex-1 sm:w-64">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <Input
              placeholder="Search API keys..."
              value={searchValue}
              onChange={(e) => handleSearch(e.target.value)}
              className="pl-9"
              aria-label="Search API keys"
            />
          </div>
          <Select value={filters?.key_type || 'all'} onValueChange={handleTypeFilter}>
            <SelectTrigger className="w-[140px]">
              <Filter className="w-4 h-4 mr-2" />
              <SelectValue placeholder="Type" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Types</SelectItem>
              <SelectItem value="platform">Platform</SelectItem>
              <SelectItem value="function">Function</SelectItem>
              <SelectItem value="agent">Agent</SelectItem>
              <SelectItem value="environment">Environment</SelectItem>
              <SelectItem value="oauth">OAuth</SelectItem>
              <SelectItem value="trust">Trust API</SelectItem>
            </SelectContent>
          </Select>
          <Select
            value={
              filters?.is_active === undefined ? 'all' : filters.is_active ? 'active' : 'inactive'
            }
            onValueChange={handleStatusFilter}
          >
            <SelectTrigger className="w-[140px]">
              <SelectValue placeholder="Status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Status</SelectItem>
              <SelectItem value="active">Active</SelectItem>
              <SelectItem value="inactive">Inactive</SelectItem>
            </SelectContent>
          </Select>
          {teams.length > 0 && (
            <Select value={filters?.team_id || 'all'} onValueChange={handleTeamFilter}>
              <SelectTrigger className="w-[160px]">
                <SelectValue placeholder="Team" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Teams</SelectItem>
                <SelectItem value="personal">Personal</SelectItem>
                {teams.map((team) => (
                  <SelectItem key={team.id} value={team.id}>{team.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </div>
      </div>

      {/* Table */}
      {apiKeys.length === 0 ? (
        <div className="apikey-empty">
          <Key className="w-12 h-12" />
          <h3 className="apikey-empty-title">No API Keys Found</h3>
          <p className="apikey-empty-text">
            {filters?.search || filters?.key_type
              ? 'Try adjusting your search or filters'
              : 'Create your first API key to get started'}
          </p>
        </div>
      ) : (
        <>
          <div className="border rounded-lg">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[40px]">
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-6 w-6"
                      onClick={toggleAll}
                      aria-label={isAllSelected ? 'Deselect all' : 'Select all'}
                    >
                      {isAllSelected ? (
                        <CheckSquare className="h-4 w-4" />
                      ) : isSomeSelected ? (
                        <CheckSquare className="h-4 w-4 text-muted-foreground" />
                      ) : (
                        <Square className="h-4 w-4 text-muted-foreground" />
                      )}
                    </Button>
                  </TableHead>
                  <TableHead>Name</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Rate Limit</TableHead>
                  <TableHead>Last Used</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {apiKeys.map((apiKey) => (
                  <TableRow
                    key={apiKey.id}
                    className={activeSelected.has(apiKey.id) ? 'bg-muted/40' : undefined}
                  >
                    <TableCell>
                      <Checkbox
                        checked={activeSelected.has(apiKey.id)}
                        onCheckedChange={() => toggleOne(apiKey.id)}
                        aria-label={`Select ${apiKey.name}`}
                      />
                    </TableCell>
                    <TableCell>
                      <Link
                        to={`/dashboard/api-keys/${apiKey.id}`}
                        className="font-medium hover:underline"
                      >
                        {apiKey.name}
                      </Link>
                      {apiKey.description && (
                        <p className="text-xs text-muted-foreground truncate max-w-[200px]">
                          {apiKey.description}
                        </p>
                      )}
                    </TableCell>
                    <TableCell>
                      <Badge variant={getKeyTypeBadgeVariant(apiKey.key_type)}>
                        {API_KEY_TYPE_LABELS[apiKey.key_type] ?? apiKey.key_type}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant={apiKey.is_active ? 'default' : 'secondary'}>
                        {apiKey.is_active ? 'Active' : 'Inactive'}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {apiKey.rate_limit_rpm
                        ? `${apiKey.rate_limit_rpm.toLocaleString()}/min`
                        : '-'}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {formatDate(apiKey.last_used_at)}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {formatDate(apiKey.created_at)}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end gap-2">
                        <Link to={`/dashboard/api-keys/${apiKey.id}`}>
                          <Button variant="ghost" size="sm">
                            View
                          </Button>
                        </Link>
                        {onRotate && (
                          <Button variant="ghost" size="sm" onClick={() => onRotate(apiKey)}>
                            Rotate
                          </Button>
                        )}
                        {onDelete && (
                          <Button
                            variant="ghost"
                            size="sm"
                            className="text-red-600 hover:text-red-600"
                            onClick={() => onDelete(apiKey)}
                          >
                            Delete
                          </Button>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between">
              <p className="text-sm text-muted-foreground">
                Showing {(page - 1) * pageSize + 1} to {Math.min(page * pageSize, total)} of {total}{' '}
                results
              </p>
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handlePageChange(page - 1)}
                  disabled={page <= 1}
                >
                  <ChevronLeft className="w-4 h-4" />
                  Previous
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handlePageChange(page + 1)}
                  disabled={page >= totalPages}
                >
                  Next
                  <ChevronRight className="w-4 h-4" />
                </Button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
