import { useState } from "react";
import { Link } from "react-router-dom";
import {
  Search,
  Filter,
  Plus,
  Key,
  ChevronLeft,
  ChevronRight,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import {
  APIKey,
  APIKeyFilters,
  API_KEY_TYPE_LABELS,
  APIKeyType,
} from "@/types/api-key";
import { cn } from "@/lib/utils";

interface APIKeyListProps {
  apiKeys: APIKey[];
  isLoading?: boolean;
  total?: number;
  page?: number;
  pageSize?: number;
  filters?: APIKeyFilters;
  onFiltersChange?: (filters: APIKeyFilters) => void;
  onPageChange?: (page: number) => void;
  onCreateNew?: () => void;
  onRotate?: (key: APIKey) => void;
  onDelete?: (key: APIKey) => void;
}

export function APIKeyList({
  apiKeys,
  isLoading = false,
  total = 0,
  page = 1,
  pageSize = 10,
  filters,
  onFiltersChange,
  onPageChange,
  onCreateNew,
  onRotate,
  onDelete,
}: APIKeyListProps) {
  const [searchValue, setSearchValue] = useState(filters?.search || "");
  const totalPages = Math.ceil(total / pageSize);

  const handleSearch = (value: string) => {
    setSearchValue(value);
    onFiltersChange?.({ ...filters, search: value, page: 1 });
  };

  const handleTypeFilter = (value: string) => {
    const newFilters: APIKeyFilters = {
      ...filters,
      key_type: value === "all" ? undefined : (value as APIKeyType),
      page: 1,
    };
    onFiltersChange?.(newFilters);
  };

  const handleStatusFilter = (value: string) => {
    const newFilters: APIKeyFilters = {
      ...filters,
      is_active: value === "all" ? undefined : value === "active",
      page: 1,
    };
    onFiltersChange?.(newFilters);
  };

  const handlePageChange = (newPage: number) => {
    onPageChange?.(newPage);
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return "Never";
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const getKeyTypeBadgeVariant = (type: string) => {
    switch (type) {
      case "platform":
        return "default";
      case "function":
        return "secondary";
      default:
        return "outline";
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    );
  }

  return (
    <div className="space-y-4">
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
            />
          </div>
          <Select
            value={filters?.key_type || "all"}
            onValueChange={handleTypeFilter}
          >
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
            </SelectContent>
          </Select>
          <Select
            value={
              filters?.is_active === undefined
                ? "all"
                : filters.is_active
                ? "active"
                : "inactive"
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
        </div>
        {onCreateNew && (
          <Button onClick={onCreateNew}>
            <Plus className="w-4 h-4 mr-2" />
            Create API Key
          </Button>
        )}
      </div>

      {/* Table */}
      {apiKeys.length === 0 ? (
        <div className="text-center py-12 border rounded-lg">
          <Key className="w-12 h-12 mx-auto mb-4 text-muted-foreground opacity-50" />
          <h3 className="text-lg font-semibold mb-2">No API Keys Found</h3>
          <p className="text-muted-foreground mb-4">
            {filters?.search || filters?.key_type
              ? "Try adjusting your search or filters"
              : "Create your first API key to get started"}
          </p>
          {!filters?.search && !filters?.key_type && onCreateNew && (
            <Button onClick={onCreateNew}>
              <Plus className="w-4 h-4 mr-2" />
              Create API Key
            </Button>
          )}
        </div>
      ) : (
        <>
          <div className="border rounded-lg">
            <Table>
              <TableHeader>
                <TableRow>
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
                  <TableRow key={apiKey.id}>
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
                        {API_KEY_TYPE_LABELS[apiKey.key_type]}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant={apiKey.is_active ? "default" : "secondary"}>
                        {apiKey.is_active ? "Active" : "Inactive"}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {apiKey.rate_limit_rpm
                        ? `${apiKey.rate_limit_rpm.toLocaleString()}/min`
                        : "-"}
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
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => onRotate(apiKey)}
                          >
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
                Showing {(page - 1) * pageSize + 1} to{" "}
                {Math.min(page * pageSize, total)} of {total} results
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
