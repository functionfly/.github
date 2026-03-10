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
    <div className="container mx-auto py-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-brand-100 rounded-lg">
            <Database className="h-6 w-6 text-brand-600" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-text-primary">Simple State</h1>
            <p className="text-sm text-text-secondary">
              Manage your key-value state storage
            </p>
          </div>
        </div>
        <Button onClick={handleCreate} className="gap-2">
          <Plus className="h-4 w-4" />
          {t("Create State")}
        </Button>
      </div>

      {/* Search and Filters */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex gap-4">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-text-muted" />
              <Input
                placeholder="Search by path or key..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10"
              />
            </div>
            <Input
              placeholder="Filter by prefix..."
              value={prefixFilter}
              onChange={(e) => setPrefixFilter(e.target.value)}
              className="w-64"
            />
          </div>
        </CardContent>
      </Card>

      {/* States Table */}
      <Card>
        <CardHeader>
          <CardTitle>States ({filteredStates?.length || 0})</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-4">
              {[...Array(5)].map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : error ? (
            <div className="text-center py-8 text-error">
              <p>Failed to load states: {(error as Error).message}</p>
            </div>
          ) : filteredStates?.length === 0 ? (
            <div className="text-center py-8">
              <Database className="h-12 w-12 mx-auto text-text-muted mb-4" />
              <p className="text-text-secondary mb-4">No states found</p>
              <Button onClick={handleCreate} variant="outline">
                Create your first state
              </Button>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Path</TableHead>
                  <TableHead>Key</TableHead>
                  <TableHead>Value</TableHead>
                  <TableHead>Version</TableHead>
                  <TableHead>Updated</TableHead>
                  <TableHead className="w-[50px]"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredStates?.map((state) => (
                  <TableRow
                    key={state.path}
                    className="cursor-pointer hover:bg-muted/50"
                    onClick={() => handleView(state.path)}
                  >
                    <TableCell className="font-mono text-sm">
                      {state.path}
                    </TableCell>
                    <TableCell className="font-medium">{state.key}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <span className="text-sm text-text-secondary max-w-[200px] truncate">
                          {formatValue(state.value)}
                        </span>
                        <Badge className={getValueTypeColor(state.value)}>
                          {getValueType(state.value)}
                        </Badge>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">v{state.version}</Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1 text-sm text-text-secondary">
                        <Clock className="h-3 w-3" />
                        {new Date(state.updatedAt).toLocaleDateString()}
                      </div>
                    </TableCell>
                    <TableCell onClick={(e) => e.stopPropagation()}>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                            <MoreVertical className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem
                            onClick={() => handleView(state.path)}
                            className="cursor-pointer"
                          >
                            <ChevronRight className="mr-2 h-4 w-4" />
                            View Details
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={() => handleDelete(state.path)}
                            className="cursor-pointer text-error focus:text-error"
                          >
                            <Trash2 className="mr-2 h-4 w-4" />
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
