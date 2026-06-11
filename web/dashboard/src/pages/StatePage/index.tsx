import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { Database, Search, Plus, ChevronRight, Clock, Trash2, MoreVertical } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  useStates,
  useDeleteState,
} from "@/hooks/useState";
import type { SimpleState } from "@/types";
import "./styles.css";

export function StatePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [searchQuery, setSearchQuery] = useState("");
  const [prefixFilter, setPrefixFilter] = useState("");

  const { data: states, isLoading, error } = useStates({
    prefix: prefixFilter || undefined,
  });
  const deleteState = useDeleteState();

  const filteredStates = states?.filter((state) =>
    state.path.toLowerCase().includes(searchQuery.toLowerCase()) ||
    state.key.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleCreate = () => {
    navigate("/state/new");
  };

  const handleView = (path: string) => {
    navigate(`/state/${encodeURIComponent(path)}`);
  };

  const handleDelete = async (path: string) => {
    if (window.confirm(`Are you sure you want to delete "${path}"?`)) {
      await deleteState.mutateAsync(path);
    }
  };

  const formatValue = (value: unknown): string => {
    if (value === null) return "null";
    if (typeof value === "object") return JSON.stringify(value);
    return String(value);
  };

  const getValueType = (value: unknown): string => {
    if (value === null) return "null";
    if (Array.isArray(value)) return "array";
    return typeof value;
  };

  const getValueTypeColor = (value: unknown): string => {
    const type = getValueType(value);
    switch (type) {
      case "string":
        return "bg-blue-100 text-blue-800";
      case "number":
        return "bg-green-100 text-green-800";
      case "boolean":
        return "bg-purple-100 text-purple-800";
      case "object":
        return "bg-orange-100 text-orange-800";
      case "array":
        return "bg-teal-100 text-teal-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  };

  return (
    <div className="state-container state-content">
      {/* Header */}
      <div className="state-header">
        <div className="state-header-left">
          <div className="state-icon-container">
            <Database className="state-icon" />
          </div>
          <div>
            <h1 className="state-title">Simple State</h1>
            <p className="state-subtitle">
              Manage your key-value state storage
            </p>
          </div>
        </div>
        <div className="state-header-actions">
          <Button onClick={handleCreate} className="btn-state-primary">
            <Plus className="h-4 w-4" />
            {t("Create State")}
          </Button>
        </div>
      </div>

      {/* Search and Filters */}
      <Card className="state-search-card">
        <CardContent className="pt-6">
          <div className="flex gap-4">
            <div className="relative flex-1">
              <Search className="state-search-icon absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4" />
              <Input
                placeholder="Search by path or key..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="state-search-input pl-10"
              />
            </div>
            <Input
              placeholder="Filter by prefix..."
              value={prefixFilter}
              onChange={(e) => setPrefixFilter(e.target.value)}
              className="state-filter-input"
            />
          </div>
        </CardContent>
      </Card>

      {/* States Table */}
      <Card className="state-table-card">
        <CardHeader className="state-table-header">
          <CardTitle className="state-table-title">States ({filteredStates?.length || 0})</CardTitle>
        </CardHeader>
        <CardContent className="state-table-content">
          {isLoading ? (
            <div className="state-loading-container">
              {[...Array(5)].map((_, i) => (
                <Skeleton key={i} className="state-skeleton" />
              ))}
            </div>
          ) : error ? (
            <div className="state-error">
              <p>Failed to load states: {(error as Error).message}</p>
            </div>
          ) : filteredStates?.length === 0 ? (
            <div className="state-empty-state">
              <Database className="state-empty-icon" />
              <p className="state-empty-description mb-4">No states found</p>
              <Button onClick={handleCreate} variant="outline" className="btn-state-outline">
                Create your first state
              </Button>
            </div>
          ) : (
            <Table className="state-table">
              <TableHeader>
                <TableRow className="state-table-header-row">
                  <TableHead className="state-table-header-cell">Path</TableHead>
                  <TableHead className="state-table-header-cell">Key</TableHead>
                  <TableHead className="state-table-header-cell">Value</TableHead>
                  <TableHead className="state-table-header-cell">Version</TableHead>
                  <TableHead className="state-table-header-cell">Updated</TableHead>
                  <TableHead className="state-table-header-cell state-table-cell-actions"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredStates?.map((state) => (
                  <TableRow
                    key={state.path}
                    className="state-table-body-row"
                    onClick={() => handleView(state.path)}
                  >
                    <TableCell className="state-table-cell state-table-cell-path">
                      {state.path}
                    </TableCell>
                    <TableCell className="state-table-cell state-table-cell-key">{state.key}</TableCell>
                    <TableCell className="state-table-cell">
                      <div className="state-table-cell-value">
                        <span className="state-table-cell-value-text">
                          {formatValue(state.value)}
                        </span>
                        <Badge className={getValueTypeColor(state.value)}>
                          {getValueType(state.value)}
                        </Badge>
                      </div>
                    </TableCell>
                    <TableCell className="state-table-cell">
                      <Badge className="state-version-badge">v{state.version}</Badge>
                    </TableCell>
                    <TableCell className="state-table-cell">
                      <div className="state-table-cell-updated">
                        <Clock className="h-3 w-3" />
                        {new Date(state.updatedAt).toLocaleDateString()}
                      </div>
                    </TableCell>
                    <TableCell className="state-table-cell state-table-cell-actions" onClick={(e) => e.stopPropagation()}>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm" className="btn-state-ghost h-8 w-8 p-0">
                            <MoreVertical className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end" className="state-dropdown-content">
                          <DropdownMenuItem
                            onClick={() => handleView(state.path)}
                            className="state-dropdown-item cursor-pointer"
                          >
                            <ChevronRight className="state-dropdown-item-icon mr-2 h-4 w-4" />
                            View Details
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={() => handleDelete(state.path)}
                            className="state-dropdown-item state-dropdown-item-danger cursor-pointer"
                          >
                            <Trash2 className="state-dropdown-item-icon mr-2 h-4 w-4" />
                            Delete
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

export default StatePage;
