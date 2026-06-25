import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { universityRankingsApi } from '@/api/universityRankings';

const KEY = 'university-rankings';

export function useUniversityLeaderboard(params?: { country?: string; limit?: number; category?: string }) {
  return useQuery({
    queryKey: [KEY, 'leaderboard', params],
    queryFn: () => universityRankingsApi.getLeaderboard(params),
    staleTime: 60_000,
  });
}

export const UNIVERSITY_CATEGORIES = [
  { slug: 'composite', label: 'Composite' },
  { slug: 'agents', label: 'Agent Capital' },
  { slug: 'automation', label: 'Automation Capital' },
  { slug: 'startups', label: 'Startup Capital' },
  { slug: 'open_source', label: 'Open Source Capital' },
  { slug: 'robotics', label: 'Robotics Capital' },
] as const;

export function useUniversityDetail(slug: string | undefined) {
  return useQuery({
    queryKey: [KEY, 'detail', slug],
    queryFn: () => universityRankingsApi.getDetail(slug!),
    enabled: !!slug,
    staleTime: 60_000,
  });
}

export function useMyUniversity() {
  return useQuery({
    queryKey: [KEY, 'me'],
    queryFn: () => universityRankingsApi.getMy(),
    staleTime: 60_000,
  });
}

export function useResolveUniversity() {
  return useMutation({
    mutationFn: (vars: { input: string; slug?: string }) =>
      universityRankingsApi.resolve(vars.input, vars.slug),
  });
}

export function useUniversityOptOut() {
  return useQuery({
    queryKey: [KEY, 'opt-out'],
    queryFn: () => universityRankingsApi.getOptOut(),
    staleTime: 60_000,
  });
}

export function useSetUniversityOptOut() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (optedOut: boolean) => universityRankingsApi.setOptOut(optedOut),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: [KEY, 'opt-out'] });
      qc.invalidateQueries({ queryKey: [KEY, 'me'] });
    },
  });
}
