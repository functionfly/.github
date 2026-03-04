/**
 * ThreadListPage - Browse threads
 */

import { useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { FlywheelPageLayout, FlywheelSection } from '@/components/flywheel/layout/FlywheelLayout';
import { ThreadCard, ThreadCardSkeleton } from '@/components/flywheel/thread/ThreadCard';
import { useThreads, useCategories } from '@/api/flywheel';
import type { ThreadFilters } from '@/components/flywheel/types';
import {
  Search,
  Plus,
  SlidersHorizontal,
  X,
} from 'lucide-react';

export default function ThreadListPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [showFilters, setShowFilters] = useState(false);

  // Parse filters from URL
  const filters: ThreadFilters = {
    search: searchParams.get('search') || undefined,
    type: (searchParams.get('type') as ThreadFilters['type']) || undefined,
    status: (searchParams.get('status') as ThreadFilters['status']) || undefined,
    category: searchParams.get('category') || undefined,
    tags: searchParams.getAll('tags') || undefined,
    sortBy: (searchParams.get('sort') as ThreadFilters['sortBy']) || 'recent',
  };

  const { data: threadsData, isLoading } = useThreads(filters, 20);
  const { data: categoriesData } = useCategories();

  const updateFilter = (key: string, value: string | undefined) => {
    const newParams = new URLSearchParams(searchParams);
    if (value) {
      newParams.set(key, value);
    } else {
      newParams.delete(key);
    }
    setSearchParams(newParams);
  };

  const clearFilters = () => {
    setSearchParams(new URLSearchParams());
  };

  const hasActiveFilters = Object.values(filters).some(Boolean);

  return (
    <FlywheelPageLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 className="text-2xl font-bold text-white">Threads</h1>
            <p className="text-slate-400">Browse and search community discussions</p>
          </div>
          <Button
            onClick={() => window.location.href = '/flywheel/threads/new'}
            className="bg-indigo-600 hover:bg-indigo-500"
          >
            <Plus className="mr-2 h-4 w-4" />
            New Thread
          </Button>
        </div>

        {/* Search & Filters */}
        <div className="space-y-4">
          {/* Search Bar */}
          <div className="flex gap-2">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
              <Input
                type="search"
                placeholder="Search threads..."
                value={filters.search || ''}
                onChange={(e) => updateFilter('search', e.target.value || undefined)}
                className="pl-10 border-slate-800 bg-slate-900 text-slate-200 focus-visible:ring-indigo-500"
              />
            </div>
            <Button
              variant="outline"
              onClick={() => setShowFilters(!showFilters)}
              className={cn(
                'border-slate-800',
                showFilters && 'bg-slate-800 text-slate-200'
              )}
            >
              <SlidersHorizontal className="mr-2 h-4 w-4" />
              Filters
            </Button>
          </div>

          {/* Filter Options */}
          {showFilters && (
            <div className="flex flex-wrap gap-3 rounded-lg border border-slate-800 bg-slate-900/50 p-4">
              <Select
                value={filters.type || 'all'}
                onValueChange={(v) => updateFilter('type', v === 'all' ? undefined : v)}
              >
                <SelectTrigger className="w-40 border-slate-700 bg-slate-800 text-slate-200">
                  <SelectValue placeholder="Type" />
                </SelectTrigger>
                <SelectContent className="bg-slate-900 border-slate-800">
                  <SelectItem value="all" className="text-slate-200 focus:bg-slate-800">All Types</SelectItem>
                  <SelectItem value="problem" className="text-slate-200 focus:bg-slate-800">Problem</SelectItem>
                  <SelectItem value="discussion" className="text-slate-200 focus:bg-slate-800">Discussion</SelectItem>
                  <SelectItem value="challenge" className="text-slate-200 focus:bg-slate-800">Challenge</SelectItem>
                </SelectContent>
              </Select>

              <Select
                value={filters.status || 'all'}
                onValueChange={(v) => updateFilter('status', v === 'all' ? undefined : v)}
              >
                <SelectTrigger className="w-40 border-slate-700 bg-slate-800 text-slate-200">
                  <SelectValue placeholder="Status" />
                </SelectTrigger>
                <SelectContent className="bg-slate-900 border-slate-800">
                  <SelectItem value="all" className="text-slate-200 focus:bg-slate-800">All Statuses</SelectItem>
                  <SelectItem value="open" className="text-slate-200 focus:bg-slate-800">Open</SelectItem>
                  <SelectItem value="in_progress" className="text-slate-200 focus:bg-slate-800">In Progress</SelectItem>
                  <SelectItem value="resolved" className="text-slate-200 focus:bg-slate-800">Resolved</SelectItem>
                </SelectContent>
              </Select>

              <Select
                value={filters.category || 'all'}
                onValueChange={(v) => updateFilter('category', v === 'all' ? undefined : v)}
              >
                <SelectTrigger className="w-48 border-slate-700 bg-slate-800 text-slate-200">
                  <SelectValue placeholder="Category" />
                </SelectTrigger>
                <SelectContent className="bg-slate-900 border-slate-800">
                  <SelectItem value="all" className="text-slate-200 focus:bg-slate-800">All Categories</SelectItem>
                  {categoriesData?.categories.map((cat) => (
                    <SelectItem
                      key={cat.id}
                      value={cat.slug}
                      className="text-slate-200 focus:bg-slate-800"
                    >
                      {cat.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <Select
                value={filters.sortBy || 'recent'}
                onValueChange={(v) => updateFilter('sort', v)}
              >
                <SelectTrigger className="w-40 border-slate-700 bg-slate-800 text-slate-200">
                  <SelectValue placeholder="Sort by" />
                </SelectTrigger>
                <SelectContent className="bg-slate-900 border-slate-800">
                  <SelectItem value="recent" className="text-slate-200 focus:bg-slate-800">Most Recent</SelectItem>
                  <SelectItem value="popular" className="text-slate-200 focus:bg-slate-800">Most Popular</SelectItem>
                  <SelectItem value="replies" className="text-slate-200 focus:bg-slate-800">Most Replies</SelectItem>
                  <SelectItem value="views" className="text-slate-200 focus:bg-slate-800">Most Views</SelectItem>
                </SelectContent>
              </Select>

              {hasActiveFilters && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={clearFilters}
                  className="text-slate-500 hover:text-slate-300"
                >
                  <X className="mr-1 h-4 w-4" />
                  Clear
                </Button>
              )}
            </div>
          )}

          {/* Active Filter Tags */}
          {hasActiveFilters && (
            <div className="flex flex-wrap gap-2">
              {filters.search && (
                <FilterTag
                  label={`Search: ${filters.search}`}
                  onRemove={() => updateFilter('search', undefined)}
                />
              )}
              {filters.type && (
                <FilterTag
                  label={`Type: ${filters.type}`}
                  onRemove={() => updateFilter('type', undefined)}
                />
              )}
              {filters.status && (
                <FilterTag
                  label={`Status: ${filters.status}`}
                  onRemove={() => updateFilter('status', undefined)}
                />
              )}
              {filters.category && (
                <FilterTag
                  label={`Category: ${filters.category}`}
                  onRemove={() => updateFilter('category', undefined)}
                />
              )}
            </div>
          )}
        </div>

        {/* Results */}
        <FlywheelSection>
          {isLoading ? (
            <div className="space-y-4">
              <ThreadCardSkeleton />
              <ThreadCardSkeleton />
              <ThreadCardSkeleton />
            </div>
          ) : threadsData?.threads.length ? (
            <div className="space-y-4">
              {threadsData.threads.map((thread) => (
                <ThreadCard key={thread.id} thread={thread} />
              ))}
            </div>
          ) : (
            <div className="rounded-xl border border-slate-800 bg-slate-900 py-12 text-center">
              <Search className="mx-auto h-12 w-12 text-slate-600" />
              <h3 className="mt-4 text-lg font-medium text-white">No threads found</h3>
              <p className="mt-1 text-slate-400">
                {hasActiveFilters
                  ? 'Try adjusting your filters'
                  : 'Be the first to start a discussion!'}
              </p>
            </div>
          )}
        </FlywheelSection>
      </div>
    </FlywheelPageLayout>
  );
}

function FilterTag({ label, onRemove }: { label: string; onRemove: () => void }) {
  return (
    <span className="inline-flex items-center gap-1 rounded-full border border-slate-700 bg-slate-800 px-3 py-1 text-sm text-slate-300">
      {label}
      <button
        onClick={onRemove}
        className="ml-1 rounded-full p-0.5 hover:bg-slate-700"
      >
        <X className="h-3 w-3" />
      </button>
    </span>
  );
}
