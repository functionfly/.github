import { galleryApi, type GalleryFunction } from '@/api/composer';
import { useQuery } from '@tanstack/react-query';
import { useMemo } from 'react';
import { mapGalleryFunction } from './constants';

/** Registry search API max per request */
export const GALLERY_PAGE_SIZE = 100;

export interface GalleryData {
  functions: GalleryFunction[];
  totalCount: number;
  page: number;
  totalPages: number;
  pageSize: number;
  isLoading: boolean;
  isError: boolean;
}

export interface GalleryStats {
  totalRemixes: number;
  distinctRuntimes: number;
  isLoading: boolean;
}

export function useGalleryStats(functions: GalleryFunction[] = []): GalleryStats {
  const { data, isLoading, isError } = useQuery({
    queryKey: ['gallery', 'stats'],
    queryFn: () => galleryApi.getStats(),
    staleTime: 5 * 60 * 1000,
    retry: 1,
  });

  const pageDistinctRuntimes = useMemo(
    () => new Set(functions.map((f) => f.runtime).filter(Boolean)).size,
    [functions]
  );
  const pageTotalRemixes = useMemo(
    () => functions.reduce((sum, f) => sum + (f.remix_count || 0), 0),
    [functions]
  );

  const useFallback = isError || (!isLoading && data == null);

  return {
    totalRemixes: data?.total_remixes ?? (useFallback ? pageTotalRemixes : 0),
    distinctRuntimes: data?.distinct_runtimes ?? (useFallback ? pageDistinctRuntimes : 0),
    isLoading,
  };
}

function filterByRuntime(functions: GalleryFunction[], runtime: string): GalleryFunction[] {
  if (!runtime || runtime === 'all') return functions;
  return functions.filter((fn) => fn.runtime === runtime);
}

export function useGalleryData(
  category: string,
  debouncedQuery: string,
  page: number,
  runtime: string = 'all'
): GalleryData {
  const safePage = Math.max(1, page);

  const { data, isLoading, isError } = useQuery({
    queryKey: ['gallery', 'functions', category, debouncedQuery, safePage],
    queryFn: async () => {
      const result = await galleryApi.search({
        category: category !== 'all' ? category : undefined,
        query: debouncedQuery.trim() || undefined,
        sort_by: debouncedQuery.trim() ? undefined : 'popular',
        limit: GALLERY_PAGE_SIZE,
        offset: (safePage - 1) * GALLERY_PAGE_SIZE,
      });
      return {
        ...result,
        results: result.results.map(mapGalleryFunction),
      };
    },
    staleTime: 2 * 60 * 1000,
    retry: 1,
  });

  const totalCount = data?.total_count ?? 0;
  const totalPages = Math.max(1, Math.ceil(totalCount / GALLERY_PAGE_SIZE));
  const functions = filterByRuntime(data?.results ?? [], runtime);

  return {
    functions,
    totalCount,
    page: Math.min(safePage, totalPages),
    totalPages,
    pageSize: GALLERY_PAGE_SIZE,
    isLoading,
    isError,
  };
}

export function getFeatured(functions: GalleryFunction[], count = 6): GalleryFunction[] {
  return [...functions].sort((a, b) => (b.trust_score || 0) - (a.trust_score || 0)).slice(0, count);
}

export function getTrending(functions: GalleryFunction[], count = 8): GalleryFunction[] {
  return [...functions]
    .sort((a, b) => (b.popularity_score || 0) - (a.popularity_score || 0))
    .slice(0, count);
}

export function galleryPageRange(page: number, pageSize: number, total: number) {
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
