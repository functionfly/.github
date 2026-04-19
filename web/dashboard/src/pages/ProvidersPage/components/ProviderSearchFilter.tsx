import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import { Search, SlidersHorizontal, X, CheckCircle2, Circle, LayoutGrid, List } from 'lucide-react';
import { useState } from 'react';

type FilterStatus = 'all' | 'connected' | 'available' | 'degraded';
type SortOption = 'name' | 'status' | 'recent' | 'regions';
type ViewMode = 'grid' | 'list';

interface ProviderSearchFilterProps {
  searchQuery: string;
  onSearchChange: (query: string) => void;
  filterStatus: FilterStatus;
  onFilterStatusChange: (status: FilterStatus) => void;
  sortBy: SortOption;
  onSortChange: (sort: SortOption) => void;
  viewMode: ViewMode;
  onViewModeChange: (mode: ViewMode) => void;
  connectedCount: number;
  availableCount: number;
  degradedCount: number;
  totalCount: number;
  className?: string;
}

const STATUS_FILTERS: { value: FilterStatus; label: string; icon: React.ReactNode }[] = [
  { value: 'all', label: 'All', icon: <LayoutGrid className="w-3.5 h-3.5" /> },
  { value: 'connected', label: 'Connected', icon: <CheckCircle2 className="w-3.5 h-3.5" /> },
  { value: 'available', label: 'Available', icon: <Circle className="w-3.5 h-3.5" /> },
  { value: 'degraded', label: 'Degraded', icon: <SlidersHorizontal className="w-3.5 h-3.5" /> },
];

const SORT_OPTIONS: { value: SortOption; label: string }[] = [
  { value: 'name', label: 'Name' },
  { value: 'status', label: 'Status' },
  { value: 'recent', label: 'Recently Used' },
  { value: 'regions', label: 'Most Regions' },
];

export function ProviderSearchFilter({
  searchQuery,
  onSearchChange,
  filterStatus,
  onFilterStatusChange,
  sortBy,
  onSortChange,
  viewMode,
  onViewModeChange,
  connectedCount,
  availableCount,
  degradedCount,
  totalCount,
  className,
}: ProviderSearchFilterProps) {
  const [showFilters, setShowFilters] = useState(false);

  const hasActiveFilters = filterStatus !== 'all' || searchQuery;

  const getCountForFilter = (filter: FilterStatus) => {
    switch (filter) {
      case 'connected':
        return connectedCount;
      case 'available':
        return availableCount;
      case 'degraded':
        return degradedCount;
      case 'all':
      default:
        return totalCount;
    }
  };

  const clearFilters = () => {
    onSearchChange('');
    onFilterStatusChange('all');
  };

  return (
    <div className={`space-y-3 ${className}`}>
      {/* Search and Actions Row */}
      <div className="flex flex-col sm:flex-row gap-3">
        {/* Search Input */}
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
          <Input
            placeholder="Search providers by name..."
            value={searchQuery}
            onChange={(e) => onSearchChange(e.target.value)}
            className="pl-9 bg-bg-secondary border-border-subtle focus:border-border-default"
          />
          {searchQuery && (
            <button
              onClick={() => onSearchChange('')}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-text-muted hover:text-text-secondary"
            >
              <X className="w-4 h-4" />
            </button>
          )}
        </div>

        {/* Filter Toggle & View Toggle */}
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowFilters(!showFilters)}
            className={`gap-2 border-border-default ${hasActiveFilters ? 'bg-blue-50 dark:bg-blue-950/30 border-blue-200 dark:border-blue-800' : ''}`}
          >
            <SlidersHorizontal className="w-4 h-4" />
            Filters
            {hasActiveFilters && (
              <Badge variant="secondary" className="ml-1 h-5 px-1.5">
                1
              </Badge>
            )}
          </Button>

          <div className="flex border border-border-default rounded-md overflow-hidden">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onViewModeChange('grid')}
              className={`rounded-none ${viewMode === 'grid' ? 'bg-bg-secondary' : ''}`}
            >
              <LayoutGrid className="w-4 h-4" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onViewModeChange('list')}
              className={`rounded-none ${viewMode === 'list' ? 'bg-bg-secondary' : ''}`}
            >
              <List className="w-4 h-4" />
            </Button>
          </div>
        </div>
      </div>

      {/* Expanded Filters */}
      {showFilters && (
        <div className="p-4 rounded-lg bg-bg-secondary/50 border border-border-subtle animate-in slide-in-from-top-2">
          <div className="flex flex-wrap items-center gap-4">
            {/* Status Filter */}
            <div className="flex items-center gap-2">
              <span className="text-sm text-text-secondary whitespace-nowrap">Status:</span>
              <div className="flex flex-wrap gap-1">
                {STATUS_FILTERS.map((filter) => (
                  <button
                    key={filter.value}
                    onClick={() => onFilterStatusChange(filter.value)}
                    className={`flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-xs font-medium transition-colors ${
                      filterStatus === filter.value
                        ? 'bg-bg-tertiary border border-border-subtle text-text-primary'
                        : 'hover:bg-bg-secondary text-text-secondary'
                    }`}
                  >
                    {filter.icon}
                    <span>{filter.label}</span>
                    <Badge variant="outline" className="ml-1 h-4 px-1 text-[10px]">
                      {getCountForFilter(filter.value)}
                    </Badge>
                  </button>
                ))}
              </div>
            </div>

            <div className="w-px h-6 bg-border-subtle hidden sm:block" />

            {/* Sort Dropdown */}
            <div className="flex items-center gap-2">
              <span className="text-sm text-text-secondary whitespace-nowrap">Sort by:</span>
              <Select value={sortBy} onValueChange={(v) => onSortChange(v as SortOption)}>
                <SelectTrigger className="w-[140px] h-8 text-xs bg-bg-tertiary border-border-subtle">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {SORT_OPTIONS.map((option) => (
                    <SelectItem key={option.value} value={option.value} className="text-xs">
                      {option.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Clear Filters */}
            {hasActiveFilters && (
              <>
                <div className="w-px h-6 bg-border-subtle hidden sm:block" />
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={clearFilters}
                  className="text-text-secondary hover:text-text-primary"
                >
                  <X className="w-4 h-4 mr-1" />
                  Clear
                </Button>
              </>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
