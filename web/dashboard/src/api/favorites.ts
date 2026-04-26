import { apiClient } from './client';
import { API_URLS } from '@/lib/api-urls';

export interface FavoriteItem {
  id: string;
  function_id: string;
  position: number;
  created_at: string;
}

export interface ListFavoritesResponse {
  favorites: FavoriteItem[];
  total: number;
}

export interface ToggleFavoriteResponse {
  favorited: boolean;
}

export interface AddFavoriteRequest {
  function_id: string;
  position?: number;
}

export interface FavoriteStatusResponse {
  function_id: string;
  favorited: boolean;
}

export const favoritesApi = {
  list: async (page = 1, limit = 20): Promise<ListFavoritesResponse> => {
    const response = await apiClient.get<ListFavoritesResponse>(
      `${API_URLS.user.favorites.list}?page=${page}&limit=${limit}`
    );
    return response;
  },

  add: async (functionId: string, position?: number): Promise<{ message: string; function_id: string; favorited: boolean }> => {
    const response = await apiClient.post<{ message: string; function_id: string; favorited: boolean }>(
      API_URLS.user.favorites.add,
      { function_id: functionId, position }
    );
    return response;
  },

  remove: async (functionId: string): Promise<{ message: string; function_id: string; favorited: boolean }> => {
    const response = await apiClient.delete<{ message: string; function_id: string; favorited: boolean }>(
      API_URLS.user.favorites.remove(functionId)
    );
    return response;
  },

  toggle: async (functionId: string): Promise<ToggleFavoriteResponse> => {
    const response = await apiClient.post<ToggleFavoriteResponse>(
      API_URLS.user.favorites.toggle(functionId)
    );
    return response;
  },

  check: async (functionId: string): Promise<FavoriteStatusResponse> => {
    const response = await apiClient.get<FavoriteStatusResponse>(
      API_URLS.user.favorites.check(functionId)
    );
    return response;
  },

  updatePosition: async (functionId: string, position: number): Promise<{ message: string }> => {
    const response = await apiClient.patch<{ message: string }>(
      API_URLS.user.favorites.updatePosition(functionId),
      { position }
    );
    return response;
  },
};
