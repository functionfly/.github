import {
  AffiliateCommission,
  FounderReferralCode,
  FounderReferralStats,
  getMyAffiliateCommissions,
  getMyReferralCode,
  getMyReferralStats,
} from '@/api/billing';
import { apiClient } from '@/api/client';
import { API_URLS } from '@/lib/api-urls';
import { useCallback, useEffect, useState } from 'react';

export interface FounderStatus {
  is_founder: boolean;
  founder_number: number | null;
  total_founders: number;
  max_founders: number;
  benefits: {
    permanent_badge: boolean;
    voting_rights: boolean;
    lifetime_commissions: boolean;
    early_access: boolean;
  };
}

export interface VoteOption {
  id: string;
  label: string;
}

export interface Vote {
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
  starts_at?: string;
  ends_at?: string;
}

export interface EarlyAccessFeature {
  slug: string;
  name: string;
  description: string;
  is_claimed: boolean;
  launched_at?: string;
}

export interface FounderRank {
  founder_number: number;
  rank: number;
  percentile: number;
  total_founders: number;
}

export interface LeaderboardEntry {
  founder_number: number;
  display_name: string;
  avatar_url?: string;
  total_earnings_cents: number;
  referral_count: number;
  member_since: string;
}

export interface FounderConsoleData {
  status: FounderStatus | null;
  rank: FounderRank | null;
  votes: Vote[];
  features: EarlyAccessFeature[];
  referralCode: FounderReferralCode | null;
  referralStats: FounderReferralStats | null;
  commissions: AffiliateCommission[];
  leaderboard: LeaderboardEntry[];
  loading: boolean;
  error: string | null;
  hasUnvotedVotes: boolean;
  castVote: (voteId: string, optionId: string) => Promise<void>;
  claimFeature: (slug: string) => Promise<void>;
  refresh: () => Promise<void>;
}

export function useFounderConsole(): FounderConsoleData {
  const [status, setStatus] = useState<FounderStatus | null>(null);
  const [rank, setRank] = useState<FounderRank | null>(null);
  const [votes, setVotes] = useState<Vote[]>([]);
  const [features, setFeatures] = useState<EarlyAccessFeature[]>([]);
  const [referralCode, setReferralCode] = useState<FounderReferralCode | null>(null);
  const [referralStats, setReferralStats] = useState<FounderReferralStats | null>(null);
  const [commissions, setCommissions] = useState<AffiliateCommission[]>([]);
  const [leaderboard, setLeaderboard] = useState<LeaderboardEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    try {
      const [
        statusRes,
        rankRes,
        votesRes,
        featuresRes,
        referralCodeRes,
        referralStatsRes,
        commissionsRes,
        leaderboardRes,
      ] = await Promise.all([
        apiClient.get<FounderStatus>(API_URLS.founders.status),
        apiClient.get<FounderRank>(API_URLS.founders.myRank),
        apiClient.get<{ votes: Vote[] }>(API_URLS.founders.votes),
        apiClient.get<{ features: EarlyAccessFeature[] }>(API_URLS.founders.earlyAccess),
        getMyReferralCode(),
        getMyReferralStats(),
        getMyAffiliateCommissions(),
        apiClient.get<{ entries: LeaderboardEntry[] }>(API_URLS.founders.leaderboard),
      ]);

      if (statusRes) setStatus(statusRes);
      if (rankRes) setRank(rankRes);
      if (votesRes) setVotes(votesRes.votes || []);
      if (featuresRes) setFeatures(featuresRes.features || []);
      if (referralCodeRes) setReferralCode(referralCodeRes);
      if (referralStatsRes) setReferralStats(referralStatsRes);
      if (commissionsRes) setCommissions(commissionsRes);
      if (leaderboardRes) setLeaderboard(leaderboardRes.entries || []);
    } catch {
      setError('Failed to load founders data');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const castVote = useCallback(
    async (voteId: string, optionId: string) => {
      await apiClient.post(`/v1/founders/votes/${voteId}`, { option_id: optionId });
      const voteDetail = await apiClient.get<{ vote: { results: Record<string, number>; total_votes: number } }>(
        `/v1/founders/votes/${voteId}`
      );
      setVotes((prev) =>
        prev.map((v) =>
          v.id === voteId
            ? {
                ...v,
                has_voted: true,
                my_vote: optionId,
                results: voteDetail?.vote?.results,
                total_votes: voteDetail?.vote?.total_votes,
              }
            : v
        )
      );
    },
    []
  );

  const claimFeature = useCallback(async (slug: string) => {
    await apiClient.post(`/v1/founders/early-access/${slug}`);
    setFeatures((prev) =>
      prev.map((f) => (f.slug === slug ? { ...f, is_claimed: true } : f))
    );
  }, []);

  const hasUnvotedVotes = votes.some((v) => v.status === 'active' && !v.has_voted);

  return {
    status,
    rank,
    votes,
    features,
    referralCode,
    referralStats,
    commissions,
    leaderboard,
    loading,
    error,
    hasUnvotedVotes,
    castVote,
    claimFeature,
    refresh: fetchData,
  };
}
