import { useEffect, useMemo, useState } from 'react';
import type { FounderStatus, Vote } from '../hooks/useFounderConsole';

interface ActivityTickerProps {
  status: FounderStatus;
  votes: Vote[];
}

export function ActivityTicker({ status, votes }: ActivityTickerProps) {
  const messages = useMemo(() => {
    const msgs: string[] = [];
    const activeVotes = votes.filter((v) => v.status === 'active');
    if (activeVotes.length > 0) {
      msgs.push(`${activeVotes.length} active proposal${activeVotes.length > 1 ? 's' : ''} awaiting your vote`);
    }
    const totalVoters = votes.reduce((sum, v) => sum + (v.total_votes ?? 0), 0);
    if (totalVoters > 0) {
      msgs.push(`${totalVoters.toLocaleString()} votes cast across all proposals`);
    }
    const slotsLeft = status.max_founders - status.total_founders;
    msgs.push(`${status.total_founders.toLocaleString()} founders — ${slotsLeft.toLocaleString()} slots remaining`);
    if (status.total_founders > 9000) {
      msgs.push('Final founder slots filling fast — limited availability');
    }
    const unclaimed = votes.filter((v) => !v.has_voted && v.status === 'active').length;
    if (unclaimed > 0) {
      msgs.push(`You have ${unclaimed} unvoted proposal${unclaimed > 1 ? 's' : ''}`);
    }
    msgs.push('Your voting weight shapes the future of FunctionFly');
    return msgs;
  }, [status, votes]);

  const [index, setIndex] = useState(0);

  useEffect(() => {
    if (messages.length <= 1) return;
    const interval = setInterval(() => {
      setIndex((prev) => (prev + 1) % messages.length);
    }, 4000);
    return () => clearInterval(interval);
  }, [messages.length]);

  if (messages.length === 0) return null;

  return (
    <div className="founders-ticker">
      <div className="founders-ticker__dot" />
      <span className="founders-ticker__label">LIVE</span>
      <span className="founders-ticker__message" key={index}>
        {messages[index]}
      </span>
    </div>
  );
}
