import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Card } from '@/components/ui/card';
import { Search, Filter, X, ChevronDown } from 'lucide-react';

interface SearchAndFiltersProps {
  searchTerm: string;
  setSearchTerm: (value: string) => void;
  releaseTypeFilter: string;
  setReleaseTypeFilter: (value: string) => void;
  categoryFilter: string;
  setCategoryFilter: (value: string) => void;
  dateFrom: string;
  setDateFrom: (value: string) => void;
  dateTo: string;
  setDateTo: (value: string) => void;
  showFilters: boolean;
  setShowFilters: (value: boolean) => void;
  hasActiveFilters: boolean;
  clearFilters: () => void;
  filteredEntriesCount: number;
  totalEntriesCount: number;
}

const SearchAndFilters = ({
  searchTerm,
  setSearchTerm,
  releaseTypeFilter,
  setReleaseTypeFilter,
  categoryFilter,
  setCategoryFilter,
  dateFrom,
  setDateFrom,
  dateTo,
  setDateTo,
  showFilters,
  setShowFilters,
  hasActiveFilters,
  clearFilters,
  filteredEntriesCount,
  totalEntriesCount
}: SearchAndFiltersProps) => {
  return (
    <div className="max-w-6xl mx-auto mb-12">
      <div className="space-y-6">
        {/* Search Bar */}
        <div className="relative group">
          <div className="absolute inset-0 bg-gradient-to-r from-brand-500/10 to-purple-500/10 rounded-xl blur-xl group-hover:blur-2xl transition-all duration-300"></div>
          <div className="relative glass-card p-1 rounded-xl glow">
            <div className="relative bg-bg-elevated/80 backdrop-blur-sm rounded-lg p-4">
              <Search className="absolute left-6 top-1/2 transform -translate-y-1/2 text-brand-500 h-5 w-5 animate-pulse-glow" />
              <Input
                placeholder="Search changelog entries..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-14 pr-4 py-3 text-lg border-0 bg-transparent focus:ring-0 placeholder:text-text-muted/60 focus:placeholder:text-text-muted/40 transition-all duration-300"
              />
            </div>
          </div>
        </div>

        {/* Filter Toggle & Results */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Button
              variant="outline"
              onClick={() => setShowFilters(!showFilters)}
              className="btn-secondary flex items-center gap-2 hover:scale-105 transition-all duration-300 hover:glow-sm"
            >
              <Filter className="h-4 w-4" />
              Advanced Filters
              <ChevronDown className={`h-4 w-4 transition-transform duration-300 ${showFilters ? 'rotate-180' : ''}`} />
            </Button>

            {hasActiveFilters && (
              <Badge variant="secondary" className="bg-brand-500/20 text-brand-500 border-brand-500/30 animate-pulse-glow">
                {filteredEntriesCount} of {totalEntriesCount} filtered
              </Badge>
            )}
          </div>

          {hasActiveFilters && (
            <Button
              variant="ghost"
              onClick={clearFilters}
              className="text-text-secondary hover:text-error bg-error/5 hover:bg-error/10 border border-error/20 hover:border-error/30 transition-all duration-300 hover:scale-105"
            >
              <X className="h-4 w-4 mr-2" />
              Clear All
            </Button>
          )}
        </div>

        {/* Advanced Filters */}
        {showFilters && (
          <Card className="glass-card glow animate-slide-down p-6 border-border-subtle/50">
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
              {/* Release Type Filter */}
              <div className="space-y-3">
                <label className="text-sm font-semibold text-text-primary flex items-center gap-2">
                  <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
                  Release Type
                </label>
                <Select value={releaseTypeFilter} onValueChange={setReleaseTypeFilter}>
                  <SelectTrigger className="glass-card border-border-subtle/50 hover:border-brand-500/50 transition-all duration-300 hover:glow-sm">
                    <SelectValue placeholder="All types" />
                  </SelectTrigger>
                  <SelectContent className="glass-card border-border-subtle/50">
                    <SelectItem value="all">All types</SelectItem>
                    <SelectItem value="major">Major Release</SelectItem>
                    <SelectItem value="minor">Minor Release</SelectItem>
                    <SelectItem value="patch">Patch Release</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {/* Category Filter */}
              <div className="space-y-3">
                <label className="text-sm font-semibold text-text-primary flex items-center gap-2">
                  <div className="w-2 h-2 bg-blue-500 rounded-full animate-pulse"></div>
                  Category
                </label>
                <Select value={categoryFilter} onValueChange={setCategoryFilter}>
                  <SelectTrigger className="glass-card border-border-subtle/50 hover:border-brand-500/50 transition-all duration-300 hover:glow-sm">
                    <SelectValue placeholder="All categories" />
                  </SelectTrigger>
                  <SelectContent className="glass-card border-border-subtle/50">
                    <SelectItem value="all">All categories</SelectItem>
                    <SelectItem value="security">Security</SelectItem>
                    <SelectItem value="features">Features</SelectItem>
                    <SelectItem value="bug fixes">Bug Fixes</SelectItem>
                    <SelectItem value="performance">Performance</SelectItem>
                    <SelectItem value="api">API Changes</SelectItem>
                    <SelectItem value="documentation">Documentation</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {/* Date From */}
              <div className="space-y-3">
                <label className="text-sm font-semibold text-text-primary flex items-center gap-2">
                  <div className="w-2 h-2 bg-purple-500 rounded-full animate-pulse"></div>
                  From Date
                </label>
                <Input
                  type="date"
                  value={dateFrom}
                  onChange={(e) => setDateFrom(e.target.value)}
                  className="glass-card border-border-subtle/50 hover:border-brand-500/50 transition-all duration-300 hover:glow-sm"
                />
              </div>

              {/* Date To */}
              <div className="space-y-3">
                <label className="text-sm font-semibold text-text-primary flex items-center gap-2">
                  <div className="w-2 h-2 bg-orange-500 rounded-full animate-pulse"></div>
                  To Date
                </label>
                <Input
                  type="date"
                  value={dateTo}
                  onChange={(e) => setDateTo(e.target.value)}
                  className="glass-card border-border-subtle/50 hover:border-brand-500/50 transition-all duration-300 hover:glow-sm"
                />
              </div>
            </div>
          </Card>
        )}

        {/* Results Count */}
        <div className="text-center">
          <div className="inline-flex items-center gap-2 px-4 py-2 bg-bg-glass/50 rounded-lg border border-border-subtle/30 text-text-secondary text-sm animate-fade-in">
            <div className="w-2 h-2 bg-brand-500 rounded-full animate-pulse-glow"></div>
            Showing <span className="font-semibold text-brand-500">{filteredEntriesCount}</span> of <span className="font-semibold text-text-primary">{totalEntriesCount}</span> entries
            {hasActiveFilters && (
              <Badge variant="outline" className="ml-2 bg-warning/10 text-warning border-warning/20">
                Filtered
              </Badge>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default SearchAndFilters;