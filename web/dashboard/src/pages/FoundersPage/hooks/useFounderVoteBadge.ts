import { apiClient } from '@/api/client';
import { API_URLS } from '@/lib/api-urls';
import { useEffect, useState } from 'react';

interface VoteCheck {
  has_voted: boolean;
}

export function useFounderVoteBadge(): boolean {
  const [hasUnvoted, setHasUnvoted] = useState(false);

  useEffect(() => {
    let cancelled = false;

    const check = async () => {
      try {
        const res = await apiClient.get<{ votes: VoteCheck[] }>(API_URLS.founders.votes);
        if (!cancelled && res?.votes) {
          setHasUnvoted(res.votes.some((v) => !v.has_voted));
        }
      } catch {
        // silent — user may not be a founder
      }
    };

    check();
    return () => {
      cancelled = true;
    };
  }, []);

  return hasUnvoted;
}
