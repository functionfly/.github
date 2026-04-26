import { useCallback, useEffect, useState } from 'react';
import { useFavoritesStore } from '@/stores/favoritesStore';

export function useFavoriteToggle(functionId: string) {
  const [localFavorited, setLocalFavorited] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  const isFavoriteStore = useFavoritesStore((state) => state.isFavorite);
  const toggleFavoriteStore = useFavoritesStore((state) => state.toggleFavorite);
  const checkFavoriteStore = useFavoritesStore((state) => state.checkFavorite);

  const isServerFavorited = isFavoriteStore(functionId);

  useEffect(() => {
    setLocalFavorited(isServerFavorited);
  }, [isServerFavorited]);

  const toggle = useCallback(async () => {
    setIsLoading(true);
    try {
      const result = await toggleFavoriteStore(functionId);
      setLocalFavorited(result);
    } catch {
      setLocalFavorited((prev) => !prev);
    } finally {
      setIsLoading(false);
    }
  }, [functionId, toggleFavoriteStore]);

  const checkFavorite = useCallback(async () => {
    const result = await checkFavoriteStore(functionId);
    setLocalFavorited(result);
    return result;
  }, [functionId, checkFavoriteStore]);

  return {
    isFavorited: localFavorited,
    isLoading,
    toggle,
    checkFavorite,
  };
}

export function useFavoriteCheck(functionIds: string[]) {
  const [favoritesMap, setFavoritesMap] = useState<Record<string, boolean>>({});
  const [isLoading, setIsLoading] = useState(true);

  const fetchFavorites = useFavoritesStore((state) => state.fetchFavorites);

  useEffect(() => {
    const checkFavorites = async () => {
      setIsLoading(true);
      try {
        await fetchFavorites(1, 100);
      } finally {
        setIsLoading(false);
      }
    };
    checkFavorites();
  }, [fetchFavorites]);

  useEffect(() => {
    const map: Record<string, boolean> = {};
    functionIds.forEach((id) => {
      map[id] = useFavoritesStore.getState().isFavorite(id);
    });
    setFavoritesMap(map);
  }, [functionIds, useFavoritesStore.getState().favorites]);

  return { favoritesMap, isLoading };
}
