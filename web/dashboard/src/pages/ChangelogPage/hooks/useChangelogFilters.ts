import { useState, useEffect } from 'react';
import { ChangelogEntry } from '@/api/content';

export const useChangelogFilters = (changelogEntries: ChangelogEntry[]) => {
  const [filteredEntries, setFilteredEntries] = useState<ChangelogEntry[]>([]);

  // Search and filter states
  const [searchTerm, setSearchTerm] = useState('');
  const [releaseTypeFilter, setReleaseTypeFilter] = useState<string>('all');
  const [categoryFilter, setCategoryFilter] = useState<string>('all');
  const [dateFrom, setDateFrom] = useState<string>('');
  const [dateTo, setDateTo] = useState<string>('');

  // Filter entries based on search and filters
  useEffect(() => {
    let filtered = [...changelogEntries];

    // Search filter
    if (searchTerm) {
      const searchLower = searchTerm.toLowerCase();
      filtered = filtered.filter(entry =>
        entry.title.toLowerCase().includes(searchLower) ||
        entry.description.toLowerCase().includes(searchLower) ||
        entry.changes.some(change =>
          change.category.toLowerCase().includes(searchLower) ||
          change.items.some(item => item.toLowerCase().includes(item))
        )
      );
    }

    // Release type filter
    if (releaseTypeFilter !== 'all') {
      filtered = filtered.filter(entry => entry.type === releaseTypeFilter);
    }

    // Category filter
    if (categoryFilter !== 'all') {
      filtered = filtered.filter(entry =>
        entry.changes.some(change => change.category.toLowerCase() === categoryFilter.toLowerCase())
      );
    }

    // Date range filter
    if (dateFrom) {
      const fromDate = new Date(dateFrom);
      filtered = filtered.filter(entry => new Date(entry.date) >= fromDate);
    }
    if (dateTo) {
      const toDate = new Date(dateTo);
      toDate.setHours(23, 59, 59, 999); // Include the entire day
      filtered = filtered.filter(entry => new Date(entry.date) <= toDate);
    }

    setFilteredEntries(filtered);
  }, [changelogEntries, searchTerm, releaseTypeFilter, categoryFilter, dateFrom, dateTo]);

  const clearFilters = () => {
    setSearchTerm('');
    setReleaseTypeFilter('all');
    setCategoryFilter('all');
    setDateFrom('');
    setDateTo('');
  };

  const hasActiveFilters = !!(searchTerm || releaseTypeFilter !== 'all' || categoryFilter !== 'all' || dateFrom || dateTo);

  return {
    filteredEntries,
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
    clearFilters,
    hasActiveFilters: Boolean(hasActiveFilters)
  };
};