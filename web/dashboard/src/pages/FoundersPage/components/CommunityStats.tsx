import { Users } from 'lucide-react';
import type { FounderStatus, LeaderboardEntry } from '../hooks/useFounderConsole';

interface CommunityStatsProps {
  status: FounderStatus;
  leaderboard: LeaderboardEntry[];
}

export function CommunityStats({ status, leaderboard }: CommunityStatsProps) {
  const pctFilled = status.max_founders > 0
    ? (status.total_founders / status.max_founders) * 100
    : 0;
  const slotsRemaining = status.max_founders - status.total_founders;

  return (
    <section className="founders-section">
      <div className="founders-section__title">
        <Users size={14} />
        Community
      </div>

      <div className="founders-chamber founders-chamber--medium">
        <div className="community-stats__bar-header">
          <span className="community-stats__count">
            <span className="community-stats__count-value">{status.total_founders.toLocaleString()}</span>
            <span className="community-stats__count-sep"> of </span>
            <span className="community-stats__count-max">{status.max_founders.toLocaleString()}</span>
          </span>
          <span className="community-stats__remaining">
            {slotsRemaining.toLocaleString()} slots remaining
          </span>
        </div>
        <div className="community-stats__progress">
          <div
            className="community-stats__progress-fill"
            style={{ width: `${pctFilled}%` }}
          />
        </div>
      </div>

      {leaderboard.length > 0 && (
        <div className="founders-chamber founders-chamber--large" style={{ marginTop: '1rem', overflow: 'auto' }}>
          <table className="founders-table">
            <thead>
              <tr>
                <th>#</th>
                <th>Founder</th>
                <th>Joined</th>
              </tr>
            </thead>
            <tbody>
              {leaderboard.slice(0, 10).map((entry) => (
                <tr key={entry.founder_number}>
                  <td style={{ fontFamily: 'var(--font-mono)', fontWeight: 600 }}>
                    {entry.founder_number}
                  </td>
                  <td>{entry.display_name}</td>
                  <td style={{ fontFamily: 'var(--font-mono)', color: 'var(--text-faint)', fontSize: 12 }}>
                    {entry.member_since}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}
