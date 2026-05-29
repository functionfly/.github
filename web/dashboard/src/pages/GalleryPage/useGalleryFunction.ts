import { galleryApi, type GalleryFunction } from '@/api/composer';
import { useQuery } from '@tanstack/react-query';
import { mapGalleryFunction } from './constants';

export function galleryFunctionPath(author: string, name: string): string {
  return `/gallery/${encodeURIComponent(author)}/${encodeURIComponent(name)}`;
}

export function useGalleryFunction(author: string | undefined, name: string | undefined) {
  return useQuery({
    queryKey: ['gallery', 'function', author, name],
    enabled: !!author && !!name,
    staleTime: 2 * 60 * 1000,
    retry: 1,
    queryFn: async (): Promise<GalleryFunction> => {
      if (!author || !name) throw new Error('Missing author or name');
      const fn = await galleryApi.getFunction(author, name);
      return mapGalleryFunction(fn);
    },
  });
}

export function useRelatedFunctions(fn: GalleryFunction | undefined, limit = 4) {
  return useQuery({
    queryKey: ['gallery', 'related', fn?.category, fn?.id],
    enabled: !!fn,
    staleTime: 5 * 60 * 1000,
    queryFn: async (): Promise<GalleryFunction[]> => {
      if (!fn) return [];

      const result = await galleryApi.search({
        category: fn.category || undefined,
        sort_by: 'popular',
        limit: 20,
      });
      return result.results
        .map(mapGalleryFunction)
        .filter((f) => f.id !== fn.id)
        .slice(0, limit);
    },
  });
}
