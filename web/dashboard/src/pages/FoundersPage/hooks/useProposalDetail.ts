import { apiClient } from '@/api/client';
import { useQuery, useQueryClient, useMutation } from '@tanstack/react-query';
import type { VoteOption } from './useFounderConsole';

export interface ProposalDetail {
  id: string;
  title: string;
  description: string;
  vote_type: string;
  status: string;
  options: VoteOption[];
  has_voted: boolean;
  my_vote?: string;
  results?: Record<string, number>;
  total_votes?: number;
  change_diff?: ChangeDiff;
  quorum: number;
  created_at: string;
}

export interface ChangeDiff {
  summary?: string;
  category?: string;
  changes?: DiffEntry[];
  impact?: string;
  rationale?: string;
}

export interface DiffEntry {
  field: string;
  label?: string;
  before?: string;
  after?: string;
  type?: 'added' | 'removed' | 'modified';
}

export function useProposalDetail(voteId: string | undefined) {
  const queryClient = useQueryClient();

  const {
    data: proposal,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['founder-vote-detail', voteId],
    queryFn: async (): Promise<ProposalDetail> => {
      const res = await apiClient.get<{ vote: ProposalDetail }>(
        `/v1/founders/votes/${voteId}`
      );
      return res.vote;
    },
    enabled: !!voteId,
    staleTime: 30 * 1000,
  });

  const castVoteMutation = useMutation({
    mutationFn: async (optionId: string) => {
      await apiClient.post(`/v1/founders/votes/${voteId}`, { option_id: optionId });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['founder-vote-detail', voteId] });
      queryClient.invalidateQueries({ queryKey: ['founders-votes'] });
    },
  });

  return {
    proposal,
    isLoading,
    error: error ? (error as Error).message : null,
    castVote: castVoteMutation.mutateAsync,
    isVoting: castVoteMutation.isPending,
  };
}
