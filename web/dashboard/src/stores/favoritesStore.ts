import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import { favoritesApi, FavoriteItem } from '@/api/favorites';

interface FavoritesState {
  favorites: FavoriteItem[];
  total: number;
  isLoading: boolean;
  error: string | null;

  fetchFavorites: (page?: number, limit?: number) => Promise<void>;
  addFavorite: (functionId: string) => Promise<void>;
  removeFavorite: (functionId: string) => Promise<void>;
  toggleFavorite: (functionId: string) => Promise<boolean>;
  checkFavorite: (functionId: string) => Promise<boolean>;
  updatePosition: (functionId: string, position: number) => Promise<void>;
  isFavorite: (functionId: string) => boolean;
  clearError: () => void;
}

export const useFavoritesStore = create<FavoritesState>()(
  immer((set, get) => ({
    favorites: [],
    total: 0,
    isLoading: false,
    error: null,

    fetchFavorites: async (page = 1, limit = 50) => {
      set((state) => {
        state.isLoading = true;
        state.error = null;
      });

      try {
        const response = await favoritesApi.list(page, limit);
        set((state) => {
          state.favorites = response.favorites;
          state.total = response.total;
          state.isLoading = false;
        });
      } catch (err) {
        set((state) => {
          state.isLoading = false;
          state.error = err instanceof Error ? err.message : 'Failed to fetch favorites';
        });
      }
    },

    addFavorite: async (functionId: string) => {
      try {
        await favoritesApi.add(functionId);
        set((state) => {
          const newFavorite: FavoriteItem = {
            id: crypto.randomUUID(),
            function_id: functionId,
            position: state.favorites.length,
            created_at: new Date().toISOString(),
          };
          state.favorites.push(newFavorite);
          state.total += 1;
        });
      } catch (err) {
        set((state) => {
          state.error = err instanceof Error ? err.message : 'Failed to add favorite';
        });
        throw err;
      }
    },

    removeFavorite: async (functionId: string) => {
      try {
        await favoritesApi.remove(functionId);
        set((state) => {
          state.favorites = state.favorites.filter((f) => f.function_id !== functionId);
          state.total = Math.max(0, state.total - 1);
        });
      } catch (err) {
        set((state) => {
          state.error = err instanceof Error ? err.message : 'Failed to remove favorite';
        });
        throw err;
      }
    },

    toggleFavorite: async (functionId: string) => {
      const isFav = get().isFavorite(functionId);
      try {
        const response = await favoritesApi.toggle(functionId);
        set((state) => {
          if (response.favorited) {
            const newFavorite: FavoriteItem = {
              id: crypto.randomUUID(),
              function_id: functionId,
              position: state.favorites.length,
              created_at: new Date().toISOString(),
            };
            state.favorites.push(newFavorite);
            state.total += 1;
          } else {
            state.favorites = state.favorites.filter((f) => f.function_id !== functionId);
            state.total = Math.max(0, state.total - 1);
          }
        });
        return response.favorited;
      } catch (err) {
        set((state) => {
          state.error = err instanceof Error ? err.message : 'Failed to toggle favorite';
        });
        throw err;
      }
    },

    checkFavorite: async (functionId: string) => {
      try {
        const response = await favoritesApi.check(functionId);
        return response.favorited;
      } catch (err) {
        set((state) => {
          state.error = err instanceof Error ? err.message : 'Failed to check favorite';
        });
        return false;
      }
    },

    updatePosition: async (functionId: string, position: number) => {
      try {
        await favoritesApi.updatePosition(functionId, position);
        set((state) => {
          const fav = state.favorites.find((f) => f.function_id === functionId);
          if (fav) {
            fav.position = position;
          }
        });
      } catch (err) {
        set((state) => {
          state.error = err instanceof Error ? err.message : 'Failed to update position';
        });
        throw err;
      }
    },

    isFavorite: (functionId: string) => {
      return get().favorites.some((f) => f.function_id === functionId);
    },

    clearError: () => {
      set((state) => {
        state.error = null;
      });
    },
  }))
);
