/**
 * MCP Center - Function Table Component
 * Displays MCP-enabled functions in a sortable, filterable table
 */

import { useState } from 'react';
import { Link } from 'react-router-dom';
import { ArrowUpDown, Shield, MoreHorizontal, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
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
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Switch } from '@/components/ui/switch';
import type { MCPFunction, MCPFunctionFilter, MCPFunctionSort } from '../types';
import { MCP_FUNCTION_FILTERS, MCP_FUNCTION_SORTS } from '../constants';
import { useToggleMCPEnabled } from '../hooks';
import { MCPSettingsDialog } from './MCPSettingsDialog';

interface FunctionTableProps {
  functions: MCPFunction[];
  isLoading: boolean;
  filter: MCPFunctionFilter;
  sort: MCPFunctionSort;
  search: string;
  onFilterChange: (filter: MCPFunctionFilter) => void;
  onSortChange: (sort: MCPFunctionSort) => void;
  onSearchChange: (search: string) => void;
}

export function FunctionTable({
  functions,
  isLoading,
  filter,
  sort,
  search,
  onFilterChange,
  onSortChange,
  onSearchChange,
}: FunctionTableProps) {
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [editSettingsFn, setEditSettingsFn] = useState<{ author: string; name: string } | null>(
    null
  );
  const toggleMCP = useToggleMCPEnabled();

  const toggleSelect = (id: string) => {
    const newSelected = new Set(selectedIds);
    if (newSelected.has(id)) {
      newSelected.delete(id);
    } else {
      newSelected.add(id);
    }
    setSelectedIds(newSelected);
  };

  const toggleSelectAll = () => {
    if (selectedIds.size === functions.length) {
      setSelectedIds(new Set());
    } else {
      setSelectedIds(new Set(functions.map((fn) => fn.id)));
    }
  };

  const formatLastInvoked = (date: string | null | undefined) => {
    if (!date) return 'Never';
    const d = new Date(date);
    const now = new Date();
    const diff = now.getTime() - d.getTime();
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);

    if (minutes < 1) return 'Just now';
    if (minutes < 60) return `${minutes}m ago`;
    if (hours < 24) return `${hours}h ago`;
    if (days < 7) return `${days}d ago`;
    return d.toLocaleDateString();
  };

  const getToolName = (fn: MCPFunction) => {
    return fn.tool_name_override || `${fn.author}__${fn.name}`;
  };

  return (
    <div className="space-y-4">
      {/* Filter Bar */}
      <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
        <div className="flex flex-wrap gap-2 items-center">
          {/* Search */}
          <Input
            placeholder="Search functions..."
            value={search}
            onChange={(e) => onSearchChange(e.target.value)}
            className="w-[200px] h-9"
          />

          {/* Filter Pills */}
          <div className="flex gap-1 bg-muted p-1 rounded-lg">
            {MCP_FUNCTION_FILTERS.map((f) => (
              <button
                key={f.value}
                onClick={() => onFilterChange(f.value as MCPFunctionFilter)}
                className={`px-3 py-1.5 text-sm rounded-md transition-colors ${
                  filter === f.value
                    ? 'bg-card text-text-primary shadow-sm'
                    : 'text-text-secondary hover:text-text-primary'
                }`}
              >
                {f.label}
              </button>
            ))}
          </div>
        </div>

        {/* Sort */}
        <Select value={sort} onValueChange={(v) => onSortChange(v as MCPFunctionSort)}>
          <SelectTrigger className="w-[160px] h-9">
            <ArrowUpDown className="h-4 w-4 mr-2" />
            <SelectValue placeholder="Sort by" />
          </SelectTrigger>
          <SelectContent>
            {MCP_FUNCTION_SORTS.map((s) => (
              <SelectItem key={s.value} value={s.value}>
                {s.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Bulk Actions */}
      {selectedIds.size > 0 && (
        <div className="flex items-center gap-2 p-3 bg-muted rounded-lg">
          <span className="text-sm text-text-secondary">{selectedIds.size} selected</span>
          <Button
            size="sm"
            variant="outline"
            onClick={() => {
              /* bulk enable */
            }}
          >
            Enable Selected
          </Button>
          <Button
            size="sm"
            variant="outline"
            onClick={() => {
              /* bulk disable */
            }}
          >
            Disable Selected
          </Button>
          <Button size="sm" variant="ghost" onClick={() => setSelectedIds(new Set())}>
            Clear selection
          </Button>
        </div>
      )}

      {/* Table */}
      <div className="mcp-table-container">
        <Table className="mcp-table">
          <TableHeader>
            <TableRow className="bg-muted/50">
              <TableHead className="w-[40px]">
                <input
                  type="checkbox"
                  checked={selectedIds.size === functions.length && functions.length > 0}
                  onChange={toggleSelectAll}
                  className="rounded"
                />
              </TableHead>
              <TableHead>Function</TableHead>
              <TableHead>Tool Name</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="text-right">Invocations</TableHead>
              <TableHead>Last Invoked</TableHead>
              <TableHead>Verified</TableHead>
              <TableHead className="w-[80px]">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={8} className="text-center py-12">
                  <Loader2 className="h-6 w-6 animate-spin mx-auto text-muted-foreground" />
                </TableCell>
              </TableRow>
            ) : functions.length === 0 ? (
              <TableRow>
                <TableCell colSpan={8} className="text-center py-12">
                  <p className="text-text-secondary">No MCP functions found</p>
                  <p className="text-sm text-muted-foreground mt-1">
                    Try adjusting your filters or search query
                  </p>
                </TableCell>
              </TableRow>
            ) : (
              functions.map((fn) => (
                <TableRow key={fn.id}>
                  <TableCell>
                    <input
                      type="checkbox"
                      checked={selectedIds.has(fn.id)}
                      onChange={() => toggleSelect(fn.id)}
                      className="rounded"
                    />
                  </TableCell>
                  <TableCell>
                    <Link
                      to={`/functions/${fn.author}/${fn.name}`}
                      className="font-medium text-text-primary hover:text-[var(--accent)]"
                    >
                      {fn.author}/{fn.name}
                    </Link>
                  </TableCell>
                  <TableCell>
                    <code className="text-sm bg-muted px-2 py-1 rounded font-mono">
                      {getToolName(fn)}
                    </code>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <Switch
                        checked={fn.enabled}
                        onCheckedChange={(enabled) =>
                          toggleMCP.mutate({ functionId: fn.id, enabled })
                        }
                        disabled={toggleMCP.isPending}
                      />
                      <Badge
                        variant={fn.enabled ? 'default' : 'secondary'}
                        className={fn.enabled ? 'bg-emerald-600' : ''}
                      >
                        {fn.enabled ? 'Enabled' : 'Disabled'}
                      </Badge>
                    </div>
                  </TableCell>
                  <TableCell className="text-right font-mono text-sm">
                    {(fn.invocation_count || 0).toLocaleString()}
                  </TableCell>
                  <TableCell className="text-sm text-text-secondary">
                    {formatLastInvoked(fn.last_invoked_at)}
                  </TableCell>
                  <TableCell>
                    {fn.verified_mcp ? (
                      <Badge variant="default" className="bg-emerald-600 gap-1">
                        <Shield className="h-3 w-3" /> Verified
                      </Badge>
                    ) : (
                      <span className="text-xs text-muted-foreground">—</span>
                    )}
                  </TableCell>
                  <TableCell>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon" className="h-8 w-8">
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem
                          onClick={() => setEditSettingsFn({ author: fn.author, name: fn.name })}
                        >
                          Edit MCP Settings
                        </DropdownMenuItem>
                        <DropdownMenuItem asChild>
                          <Link to={`/functions/${fn.author}/${fn.name}`}>View Function</Link>
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          onClick={() =>
                            toggleMCP.mutate({ functionId: fn.id, enabled: !fn.enabled })
                          }
                        >
                          {fn.enabled ? 'Disable MCP' : 'Enable MCP'}
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* MCP Settings Dialog */}
      {editSettingsFn && (
        <MCPSettingsDialog
          open={!!editSettingsFn}
          onOpenChange={(open) => !open && setEditSettingsFn(null)}
          author={editSettingsFn.author}
          name={editSettingsFn.name}
        />
      )}
    </div>
  );
}
