/**
 * Functions Tab Component
 *
 * Displays user's published functions with search and filter capabilities.
 */

import { FunctionCard } from '@/components/functions/FunctionCard';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { formatNumber } from '@/lib/utils';
import type { FunctionFilters, UserProfile } from '@/types';
import { motion } from 'framer-motion';
import { Filter, Package, Search } from 'lucide-react';
import { useMemo, useState } from 'react';
import { tabContentVariants } from '../../animations';

export interface FunctionsTabProps {
  profile: UserProfile;
}

export function FunctionsTab({ profile }: FunctionsTabProps) {
  const [filters, setFilters] = useState<FunctionFilters>({
    search: '',
    sortBy: 'popular',
  });

  const filteredFunctions = useMemo(() => {
    let result = [...profile.publishedFunctions];

    if (filters.search) {
      const searchLower = filters.search.toLowerCase();
      result = result.filter(
        (f) =>
          f.name.toLowerCase().includes(searchLower) ||
          f.description.toLowerCase().includes(searchLower) ||
          f.tags?.some((t) => t.toLowerCase().includes(searchLower))
      );
    }

    result.sort((a, b) => {
      switch (filters.sortBy) {
        case 'popular':
          return (b.metrics.executionCount || 0) - (a.metrics.executionCount || 0);
        case 'recent':
          return new Date(b.lastUpdated || 0).getTime() - new Date(a.lastUpdated || 0).getTime();
        case 'name':
          return a.name.localeCompare(b.name);
        case 'rating':
          return (b.rating?.average || 0) - (a.rating?.average || 0);
        default:
          return 0;
      }
    });

    return result;
  }, [profile.publishedFunctions, filters]);

  return (
    <motion.div
      variants={tabContentVariants}
      initial="hidden"
      animate="visible"
      exit="exit"
      className="space-y-6 px-4 md:px-8 pb-8"
    >
      {/* Filter Bar */}
      <Card className="border-border-subtle">
        <CardContent className="p-4">
          <div className="flex flex-col sm:flex-row gap-4">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
              <Input
                placeholder="Search functions..."
                value={filters.search}
                onChange={(e) => setFilters((prev) => ({ ...prev, search: e.target.value }))}
                className="pl-9"
              />
            </div>
            <Select
              value={filters.sortBy}
              onValueChange={(value) =>
                setFilters((prev) => ({ ...prev, sortBy: value as FunctionFilters['sortBy'] }))
              }
            >
              <SelectTrigger className="w-full sm:w-40">
                <Filter className="w-4 h-4 mr-2" />
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="popular">Most Popular</SelectItem>
                <SelectItem value="recent">Recently Updated</SelectItem>
                <SelectItem value="name">Name</SelectItem>
                <SelectItem value="rating">Highest Rated</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* Statistics Summary */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <Card className="border-border-subtle">
          <CardContent className="p-4 text-center">
            <p className="text-2xl font-bold font-mono tabular-nums text-text-primary">
              {profile.publishedFunctions.length}
            </p>
            <p className="text-sm text-text-muted">Total Functions</p>
          </CardContent>
        </Card>
        <Card className="border-border-subtle">
          <CardContent className="p-4 text-center">
            <p className="text-2xl font-bold font-mono tabular-nums text-text-primary">
              {formatNumber(profile.stats.totalExecutions)}
            </p>
            <p className="text-sm text-text-muted">Total Executions</p>
          </CardContent>
        </Card>
        <Card className="border-border-subtle">
          <CardContent className="p-4 text-center">
            <p className="text-2xl font-bold font-mono tabular-nums text-text-primary">
              {formatNumber(profile.stats.totalViews)}
            </p>
            <p className="text-sm text-text-muted">Total Views</p>
          </CardContent>
        </Card>
        <Card className="border-border-subtle">
          <CardContent className="p-4 text-center">
            <p className="text-2xl font-bold font-mono tabular-nums text-text-primary">
              {(
                profile.publishedFunctions.reduce((sum, f) => sum + (f.rating?.average || 0), 0) /
                  profile.publishedFunctions.length || 0
              ).toFixed(1)}
            </p>
            <p className="text-sm text-text-muted">Avg Rating</p>
          </CardContent>
        </Card>
      </div>

      {/* Functions Grid */}
      {filteredFunctions.length > 0 ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filteredFunctions.map((fn) => (
            <FunctionCard key={fn.id} data={fn} variant="compact" />
          ))}
        </div>
      ) : (
        <div className="text-center py-16">
          <Package className="w-16 h-16 mx-auto text-text-muted mb-4" />
          <h3 className="text-lg font-medium text-text-primary mb-2">No functions found</h3>
          <p className="text-text-muted">Try adjusting your search or filters</p>
        </div>
      )}
    </motion.div>
  );
}
