import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { cityRankingsApi } from '@/api/cityRankings';

const KEY = 'city-rankings';

export function useCityLeaderboard(params?: { country?: string; limit?: number }) {
  return useQuery({
    queryKey: [KEY, 'leaderboard', params],
    queryFn: () => cityRankingsApi.getLeaderboard(params),
    staleTime: 60_000,
  });
}

export function useCityMovers(direction: 'gainers' | 'losers' = 'gainers') {
  return useQuery({
    queryKey: [KEY, 'movers', direction],
    queryFn: () => cityRankingsApi.getMovers(direction),
    staleTime: 60_000,
  });
}

export function useMyCity() {
  return useQuery({
    queryKey: [KEY, 'me'],
    queryFn: () => cityRankingsApi.getMyCity(),
    staleTime: 60_000,
  });
}

export function useCityOptOut() {
  return useQuery({
    queryKey: [KEY, 'opt-out'],
    queryFn: () => cityRankingsApi.getOptOut(),
    staleTime: 60_000,
  });
}

export function useSetCityOptOut() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (optedOut: boolean) => cityRankingsApi.setOptOut(optedOut),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: [KEY, 'opt-out'] });
      qc.invalidateQueries({ queryKey: [KEY, 'me'] });
    },
  });
}
