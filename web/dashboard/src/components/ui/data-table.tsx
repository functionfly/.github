import * as React from "react";
import {
  useReactTable,
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  getPaginationRowModel,
  flexRender,
  type ColumnDef,
  type ColumnFiltersState,
  type SortingState,
  type VisibilityState,
  type ColumnResizeMode,
  type RowSelectionState,
} from "@tanstack/react-table";
import { Download, Filter, Settings2, ChevronDown, X } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "./button";
import { Checkbox } from "./checkbox";
import { Input } from "./input";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "./dropdown-menu";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "./table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "./select";

// ============================================================================
// Types
// ============================================================================

export interface DataTableProps<TData> {
  data: TData[];
  columns: ColumnDef<TData, any>[];
  enableRowSelection?: boolean;
  enableColumnResize?: boolean;
  enableColumnVisibility?: boolean;
  enableExport?: boolean;
  enableGlobalFilter?: boolean;
  enableColumnFilters?: boolean;
  onRowSelectionChange?: (selectedRows: TData[]) => void;
  onBulkAction?: (action: string, selectedRows: TData[]) => void;
  bulkActions?: { label: string; value: string; variant?: "default" | "destructive" }[];
  exportFileName?: string;
  className?: string;
  emptyState?: React.ReactNode;
  isLoading?: boolean;
}

// ============================================================================
// Export Utilities
// ============================================================================

function downloadCSV(data: any[], columns: ColumnDef<any, any>[], filename: string) {
  const headers = columns
    .filter((col) => col.id !== "select")
    .map((col) => {
      const header = col.header;
      if (typeof header === "string") return header;
      if (typeof header === "function") {
        const result = header({ column: col as any, header: col.header as any });
        if (typeof result === "string") return result;
        return col.id || "Column";
      }
      return col.id || "Column";
    });

  const rows = data.map((row) =>
    columns
      .filter((col) => col.id !== "select")
      .map((col) => {
        const accessor = col.accessorKey as string;
        const value = accessor ? (row as any)[accessor] : "";
        // Escape CSV values
        if (typeof value === "string" && (value.includes(",") || value.includes('"') || value.includes("\n"))) {
          return `"${value.replace(/"/g, '""')}"`;
        }
        return value ?? "";
      })
  );

  const csvContent = [headers.join(","), ...rows.map((row) => row.join(","))].join("\n");
  const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" });
  const link = document.createElement("a");
  link.href = URL.createObjectURL(blob);
  link.download = `${filename}.csv`;
  link.click();
}

// ============================================================================
// Main Component
// ============================================================================

export function DataTable<TData>({
  data,
  columns,
  enableRowSelection = true,
  enableColumnResize = true,
  enableColumnVisibility = true,
  enableExport = true,
  enableGlobalFilter = true,
  enableColumnFilters = true,
  onRowSelectionChange,
  onBulkAction,
  bulkActions = [],
  exportFileName = "export",
  className,
  emptyState,
  isLoading,
}: DataTableProps<TData>) {
  // State
  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>([]);
  const [globalFilter, setGlobalFilter] = React.useState("");
  const [columnVisibility, setColumnVisibility] = React.useState<VisibilityState>({});
  const [rowSelection, setRowSelection] = React.useState<RowSelectionState>({});
  const [columnResizeMode] = React.useState<ColumnResizeMode>("onChange");
  const [pagination, setPagination] = React.useState({ pageIndex: 0, pageSize: 10 });

  // Build columns with selection column
  const tableColumns = React.useMemo<ColumnDef<TData, any>[]>(() => {
    if (!enableRowSelection) return columns;

    const selectColumn: ColumnDef<TData, any> = {
      id: "select",
      header: ({ table }) => (
        <Checkbox
          checked={table.getIsAllPageRowsSelected() || (table.getIsSomePageRowsSelected() && "indeterminate")}
          onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
          aria-label="Select all"
          className="translate-y-[2px]"
        />
      ),
      cell: ({ row }) => (
        <Checkbox
          checked={row.getIsSelected()}
          onCheckedChange={(value) => row.toggleSelected(!!value)}
          aria-label="Select row"
          className="translate-y-[2px]"
        />
      ),
      enableSorting: false,
      enableHiding: false,
      size: 50,
    };

    return [selectColumn, ...columns];
  }, [columns, enableRowSelection]);

  // Initialize table
  const table = useReactTable({
    data,
    columns: tableColumns,
    state: {
      sorting,
      columnFilters,
      globalFilter,
      columnVisibility,
      rowSelection,
      pagination,
    },
    enableRowSelection,
    columnResizeMode,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onGlobalFilterChange: setGlobalFilter,
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
    onPaginationChange: setPagination,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
  });

  // Notify parent of selection changes
  React.useEffect(() => {
    if (onRowSelectionChange) {
      const selectedRows = table.getFilteredSelectedRowModel().rows.map((row) => row.original);
      onRowSelectionChange(selectedRows);
    }
  }, [rowSelection, onRowSelectionChange, table]);

  // Get selected rows for bulk actions
  const selectedRows = React.useMemo(() => {
    return table.getFilteredSelectedRowModel().rows.map((row) => row.original);
  }, [rowSelection, table]);

  const hasSelection = selectedRows.length > 0;

  // Handle export
  const handleExport = () => {
    const exportData = hasSelection ? selectedRows : data;
    downloadCSV(exportData, columns, exportFileName);
  };

  // Clear all filters
  const clearFilters = () => {
    setColumnFilters([]);
    setGlobalFilter("");
  };

  const hasActiveFilters = columnFilters.length > 0 || globalFilter !== "";

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="flex items-center gap-4">
          <div className="h-10 w-[300px] animate-pulse rounded bg-muted" />
          <div className="h-10 w-[120px] animate-pulse rounded bg-muted" />
        </div>
        <div className="rounded-md border">
          <div className="h-[300px] animate-pulse bg-muted" />
        </div>
      </div>
    );
  }

  if (!isLoading && data.length === 0 && emptyState) {
    return <>{emptyState}</>;
  }

  return (
    <div className={cn("space-y-4", className)}>
      {/* Toolbar */}
      <div className="flex flex-wrap items-center justify-between gap-4">
        {/* Left: Search and Filters */}
        <div className="flex flex-wrap items-center gap-2">
          {enableGlobalFilter && (
            <div className="relative">
              <Input
                placeholder="Search all columns..."
                value={globalFilter}
                onChange={(e) => setGlobalFilter(e.target.value)}
                className="w-[250px] pl-3"
              />
              {globalFilter && (
                <button
                  onClick={() => setGlobalFilter("")}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                >
                  <X className="h-4 w-4" />
                </button>
              )}
            </div>
          )}

          {/* Column Filter Dropdown */}
          {enableColumnFilters && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="sm" className="gap-2">
                  <Filter className="h-4 w-4" />
                  Filter
                  {columnFilters.length > 0 && (
                    <span className="ml-1 rounded-full bg-brand-500 px-1.5 py-0.5 text-xs text-white">
                      {columnFilters.length}
                    </span>
                  )}
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="start" className="w-[300px]">
                <div className="p-2">
                  <p className="mb-2 text-sm font-semibold">Column Filters</p>
                  {table.getAllColumns().filter((col) => col.getCanFilter()).length === 0 ? (
                    <p className="text-sm text-muted-foreground">No filterable columns</p>
                  ) : (
                    <div className="space-y-2">
                      {table.getAllColumns().map((column) => {
                        if (!column.getCanFilter() || column.id === "select") return null;
                        const header = column.columnDef.header;
                        const columnName =
                          typeof header === "string"
                            ? header
                            : column.id;
                        return (
                          <div key={column.id} className="flex items-center gap-2">
                            <label className="text-sm min-w-[100px]">{columnName}</label>
                            <Input
                              placeholder={`Filter ${columnName}...`}
                              value={(column.getFilterValue() as string) ?? ""}
                              onChange={(e) => column.setFilterValue(e.target.value)}
                              className="flex-1 h-8"
                            />
                          </div>
                        );
                      })}
                    </div>
                  )}
                  {hasActiveFilters && (
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={clearFilters}
                      className="mt-3 w-full gap-2"
                    >
                      <X className="h-4 w-4" />
                      Clear all filters
                    </Button>
                  )}
                </div>
              </DropdownMenuContent>
            </DropdownMenu>
          )}

          {hasActiveFilters && (
            <Button variant="ghost" size="sm" onClick={clearFilters} className="gap-2">
              <X className="h-4 w-4" />
              Clear
            </Button>
          )}
        </div>

        {/* Right: Bulk Actions, Column Visibility, Export */}
        <div className="flex items-center gap-2">
          {/* Bulk Actions */}
          {hasSelection && bulkActions.length > 0 && (
            <div className="flex items-center gap-2 border-r border-border pr-2">
              <span className="text-sm text-muted-foreground">
                {selectedRows.length} selected
              </span>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="secondary" size="sm" className="gap-2">
                    Actions
                    <ChevronDown className="h-4 w-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  {bulkActions.map((action) => (
                    <DropdownMenuCheckboxItem
                      key={action.value}
                      onSelect={() => onBulkAction?.(action.value, selectedRows)}
                    >
                      {action.label}
                    </DropdownMenuCheckboxItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => table.toggleAllRowsSelected(false)}
              >
                Cancel
              </Button>
            </div>
          )}

          {/* Column Visibility */}
          {enableColumnVisibility && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="sm" className="gap-2">
                  <Settings2 className="h-4 w-4" />
                  Columns
                  <ChevronDown className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-[200px]">
                {table
                  .getAllColumns()
                  .filter((column) => column.getCanHide())
                  .map((column) => {
                    const header = column.columnDef.header;
                    const columnName =
                      typeof header === "string"
                        ? header
                        : column.id === "select"
                          ? "Select"
                          : column.id;
                    return (
                      <DropdownMenuCheckboxItem
                        key={column.id}
                        className="capitalize"
                        checked={column.getIsVisible()}
                        onCheckedChange={(value) => column.toggleVisibility(!!value)}
                      >
                        {columnName}
                      </DropdownMenuCheckboxItem>
                    );
                  })}
              </DropdownMenuContent>
            </DropdownMenu>
          )}

          {/* Export */}
          {enableExport && (
            <Button variant="outline" size="sm" onClick={handleExport} className="gap-2">
              <Download className="h-4 w-4" />
              Export{hasSelection ? ` (${selectedRows.length})` : ""}
            </Button>
          )}
        </div>
      </div>

      {/* Selected Count */}
      {hasSelection && !bulkActions.length && (
        <div className="flex items-center gap-2 rounded-md bg-muted px-3 py-2 text-sm">
          <Checkbox checked={true} className="translate-y-[2px]" />
          <span>
            {selectedRows.length} row{selectedRows.length === 1 ? "" : "s"} selected
          </span>
          <Button
            variant="ghost"
            size="sm"
            className="h-auto px-2 py-0 text-xs"
            onClick={() => table.toggleAllRowsSelected(false)}
          >
            Clear
          </Button>
        </div>
      )}

      {/* Table */}
      <div className="rounded-md border">
        <div className="relative w-full overflow-auto">
          <Table>
            <TableHeader>
              {table.getHeaderGroups().map((headerGroup) => (
                <TableRow key={headerGroup.id}>
                  {headerGroup.headers.map((header) => {
                    const isResizing = header.column.getIsResizing();
                    return (
                      <TableHead
                        key={header.id}
                        style={{
                          width: header.getSize(),
                          position: "relative",
                        }}
                        className={cn(
                          "relative select-none",
                          header.column.getCanSort() && "cursor-pointer",
                          isResizing && "bg-muted"
                        )}
                        onClick={header.column.getToggleSortingHandler()}
                      >
                        {header.isPlaceholder ? null : (
                          <div className="flex items-center gap-2">
                            {flexRender(header.column.columnDef.header, header.getContext())}
                            {header.column.getIsSorted() === "asc" && <span>↑</span>}
                            {header.column.getIsSorted() === "desc" && <span>↓</span>}
                          </div>
                        )}

                        {/* Resize Handle */}
                        {enableColumnResize && header.column.getCanResize() && (
                          <div
                            onMouseDown={header.getResizeHandler()}
                            onTouchStart={header.getResizeHandler()}
                            className={cn(
                              "absolute right-0 top-0 h-full w-1 cursor-col-resize touch-none select-none bg-border opacity-0 hover:opacity-100",
                              isResizing && "opacity-100 bg-brand-500"
                            )}
                            style={{ transform: isResizing ? "translateX(50%)" : undefined }}
                          />
                        )}
                      </TableHead>
                    );
                  })}
                </TableRow>
              ))}
            </TableHeader>
            <TableBody>
              {table.getRowModel().rows?.length ? (
                table.getRowModel().rows.map((row) => (
                  <TableRow
                    key={row.id}
                    data-state={row.getIsSelected() && "selected"}
                    className={cn(
                      "transition-colors",
                      row.getIsSelected() && "bg-muted/50"
                    )}
                  >
                    {row.getVisibleCells().map((cell) => (
                      <TableCell
                        key={cell.id}
                        style={{ width: cell.column.getSize() }}
                      >
                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                      </TableCell>
                    ))}
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell colSpan={columns.length + (enableRowSelection ? 1 : 0)} className="h-32 text-center">
                    No results found.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>

      {/* Pagination */}
      <div className="flex items-center justify-between">
        <div className="text-sm text-muted-foreground">
          Showing {table.getRowModel().rows.length} of {data.length} results
          {hasSelection && ` (${selectedRows.length} selected)`}
        </div>
        <div className="flex items-center gap-2">
          <Select
            value={`${table.getState().pagination.pageSize}`}
            onValueChange={(value) => table.setPageSize(Number(value))}
          >
            <SelectTrigger className="h-8 w-[80px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {[10, 20, 50, 100].map((size) => (
                <SelectItem key={size} value={`${size}`}>
                  {size}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <div className="flex items-center gap-1">
            <Button
              variant="outline"
              size="sm"
              onClick={() => table.setPageIndex(0)}
              disabled={!table.getCanPreviousPage()}
            >
              First
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => table.previousPage()}
              disabled={!table.getCanPreviousPage()}
            >
              Previous
            </Button>
            <span className="px-2 text-sm">
              Page {table.getState().pagination.pageIndex + 1} of {table.getPageCount() || 1}
            </span>
            <Button
              variant="outline"
              size="sm"
              onClick={() => table.nextPage()}
              disabled={!table.getCanNextPage()}
            >
              Next
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => table.setPageIndex(table.getPageCount() - 1)}
              disabled={!table.getCanNextPage()}
            >
              Last
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}

export default DataTable;
