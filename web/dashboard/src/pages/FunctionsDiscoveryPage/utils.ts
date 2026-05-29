export const DISCOVERY_PAGE_SIZE = 24;

export const PAGINATED_FILTERS = new Set(['hot', 'trending', 'new', 'popular']);

export type DiscoverySortBy = 'popular' | 'recent' | 'rating' | 'name' | 'hot' | 'trending';

export const DISCOVERY_SORT_OPTIONS: { value: DiscoverySortBy; label: string }[] = [
  { value: 'hot', label: 'Hot Right Now' },
  { value: 'trending', label: 'Trending This Week' },
  { value: 'popular', label: 'Most Popular' },
  { value: 'recent', label: 'Newest' },
  { value: 'rating', label: 'Highest Rated' },
  { value: 'name', label: 'Name (A–Z)' },
];

export function defaultSortForFilter(filter: string): DiscoverySortBy {
  switch (filter) {
    case 'hot':
      return 'hot';
    case 'trending':
      return 'trending';
    case 'new':
      return 'recent';
    case 'popular':
    default:
      return 'popular';
  }
}

export function parseDiscoverySort(value: string | null, filter: string): DiscoverySortBy {
  const valid = DISCOVERY_SORT_OPTIONS.map((o) => o.value);
  if (value && valid.includes(value as DiscoverySortBy)) {
    return value as DiscoverySortBy;
  }
  return defaultSortForFilter(filter);
}

export function discoveryPageRange(page: number, pageSize: number, total: number) {
  if (total === 0) return { start: 0, end: 0 };
  const start = (page - 1) * pageSize + 1;
  const end = Math.min(page * pageSize, total);
  return { start, end };
}

/** Compact page list with ellipsis for large page counts */
export function getVisiblePages(current: number, total: number): (number | 'ellipsis')[] {
  if (total <= 7) {
    return Array.from({ length: total }, (_, i) => i + 1);
  }

  const pages: (number | 'ellipsis')[] = [1];

  if (current > 3) pages.push('ellipsis');

  const rangeStart = Math.max(2, current - 1);
  const rangeEnd = Math.min(total - 1, current + 1);

  for (let p = rangeStart; p <= rangeEnd; p++) {
    pages.push(p);
  }

  if (current < total - 2) pages.push('ellipsis');

  pages.push(total);
  return pages;
}
