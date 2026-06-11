import { Filter, X } from 'lucide-react';
import { Button } from '@/components/ui';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';

interface SearchFiltersProps {
  toolName: string;
  filters: Record<string, unknown>;
  onFiltersChange: (filters: Record<string, unknown>) => void;
}

export function SearchFilters({ toolName, filters, onFiltersChange }: SearchFiltersProps) {
  const updateFilter = (key: string, value: unknown) => {
    onFiltersChange({ ...filters, [key]: value });
  };

  const clearFilters = () => {
    onFiltersChange({});
  };

  const hasFilters = Object.keys(filters).length > 0;

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline" size="sm" className="gap-2">
          <Filter className="h-4 w-4" />
          Filters
          {hasFilters && (
            <span className="ml-1 rounded-full bg-primary text-primary-foreground text-xs px-1.5 py-0.5">
              {Object.keys(filters).length}
            </span>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80" align="end">
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h4 className="font-medium">Search Filters</h4>
            {hasFilters && (
              <Button
                variant="ghost"
                size="sm"
                onClick={clearFilters}
                className="h-8 px-2"
              >
                <X className="h-4 w-4 mr-1" />
                Clear
              </Button>
            )}
          </div>

          {toolName === 'search.web' && (
            <>
              <div className="space-y-2">
                <label className="text-sm font-medium">Date Range</label>
                <Select
                  value={(filters.dateRange as string) || 'any'}
                  onValueChange={(v) => updateFilter('dateRange', v)}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="any">Any time</SelectItem>
                    <SelectItem value="day">Past 24 hours</SelectItem>
                    <SelectItem value="week">Past week</SelectItem>
                    <SelectItem value="month">Past month</SelectItem>
                    <SelectItem value="year">Past year</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Max Results</label>
                <Select
                  value={String(filters.maxResults || 10)}
                  onValueChange={(v) => updateFilter('maxResults', parseInt(v))}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="5">5 results</SelectItem>
                    <SelectItem value="10">10 results</SelectItem>
                    <SelectItem value="20">20 results</SelectItem>
                    <SelectItem value="50">50 results</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </>
          )}

          {toolName === 'search.news' && (
            <>
              <div className="space-y-2">
                <label className="text-sm font-medium">Sort By</label>
                <Select
                  value={(filters.sortBy as string) || 'relevance'}
                  onValueChange={(v) => updateFilter('sortBy', v)}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="relevance">Relevance</SelectItem>
                    <SelectItem value="date">Date</SelectItem>
                    <SelectItem value="popularity">Popularity</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Language</label>
                <Select
                  value={(filters.language as string) || 'en'}
                  onValueChange={(v) => updateFilter('language', v)}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="en">English</SelectItem>
                    <SelectItem value="es">Spanish</SelectItem>
                    <SelectItem value="fr">French</SelectItem>
                    <SelectItem value="de">German</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </>
          )}

          {toolName === 'search.docs' && (
            <div className="space-y-2">
              <label className="text-sm font-medium">Source Type</label>
              <Select
                value={(filters.sourceType as string) || 'all'}
                onValueChange={(v) => updateFilter('sourceType', v)}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Sources</SelectItem>
                  <SelectItem value="github">GitHub</SelectItem>
                  <SelectItem value="official">Official Docs</SelectItem>
                  <SelectItem value="stackoverflow">Stack Overflow</SelectItem>
                </SelectContent>
              </Select>
            </div>
          )}

          {toolName === 'search.company' && (
            <div className="space-y-3">
              <label className="text-sm font-medium">Include Additional Data</label>
              <div className="flex flex-col gap-2">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={filters.includeNews as boolean || false}
                    onChange={(e) => updateFilter('includeNews', e.target.checked)}
                    className="rounded border-input"
                  />
                  <span className="text-sm">News articles</span>
                </label>
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={filters.includeFunding as boolean || false}
                    onChange={(e) => updateFilter('includeFunding', e.target.checked)}
                    className="rounded border-input"
                  />
                  <span className="text-sm">Funding information</span>
                </label>
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={filters.includeTechnologies as boolean || false}
                    onChange={(e) => updateFilter('includeTechnologies', e.target.checked)}
                    className="rounded border-input"
                  />
                  <span className="text-sm">Technologies used</span>
                </label>
              </div>
            </div>
          )}
        </div>
      </PopoverContent>
    </Popover>
  );
}