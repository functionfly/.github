import { useState, useMemo } from 'react';
import { Link } from 'react-router-dom';
import {
  Vote as VoteIcon,
  ChevronRight,
  Clock,
  CheckCircle2,
  XCircle,
  BarChart3,
} from 'lucide-react';
import type { Vote } from '../hooks/useFounderConsole';

type VoteTab = 'active' | 'passed' | 'all';

interface VoteHubProps {
  votes: Vote[];
  totalFounders: number;
}

export function VoteHub({ votes, totalFounders }: VoteHubProps) {
  const [activeTab, setActiveTab] = useState<VoteTab>('active');

  const filteredVotes = useMemo(() => {
    switch (activeTab) {
      case 'active':
        return votes.filter((v) => v.status === 'active');
      case 'passed':
        return votes.filter((v) => v.status === 'passed' || v.status === 'closed');
      case 'all':
      default:
        return votes;
    }
  }, [votes, activeTab]);

  const activeCount = votes.filter((v) => v.status === 'active').length;
  const unvotedActive = votes.filter((v) => v.status === 'active' && !v.has_voted).length;

  if (votes.length === 0) return null;

  return (
    <section className="founders-section vote-hub">
      <div className="vote-hub__header">
        <div className="founders-section__title">
          <VoteIcon size={14} />
          Governance Hub
        </div>
        <Link to="/founders" className="vote-hub__view-all">
          <BarChart3 size={12} />
          {votes.length} proposals
        </Link>
      </div>

      {/* Tabs */}
      <div className="vote-hub__tabs">
        <button
          className={`vote-hub__tab ${activeTab === 'active' ? 'vote-hub__tab--active' : ''}`}
          onClick={() => setActiveTab('active')}
        >
          Active
          {activeCount > 0 && (
            <span className="vote-hub__tab-count">{activeCount}</span>
          )}
        </button>
        <button
          className={`vote-hub__tab ${activeTab === 'passed' ? 'vote-hub__tab--active' : ''}`}
          onClick={() => setActiveTab('passed')}
        >
          Results
        </button>
        <button
          className={`vote-hub__tab ${activeTab === 'all' ? 'vote-hub__tab--active' : ''}`}
          onClick={() => setActiveTab('all')}
        >
          All
        </button>
      </div>

      {/* Unvoted banner */}
      {unvotedActive > 0 && activeTab === 'active' && (
        <div className="vote-hub__banner">
          <VoteIcon size={14} />
          <span>
            {unvotedActive} proposal{unvotedActive !== 1 ? 's' : ''} awaiting your vote
          </span>
        </div>
      )}

      {/* Proposal list */}
      <div className="vote-hub__list">
        {filteredVotes.length === 0 ? (
          <div className="vote-hub__empty">
            <p>No {activeTab === 'passed' ? 'completed' : activeTab} proposals</p>
          </div>
        ) : (
          filteredVotes.map((vote) => (
            <ProposalRow
              key={vote.id}
              vote={vote}
              totalFounders={totalFounders}
            />
          ))
        )}
      </div>
    </section>
  );
}

interface ProposalRowProps {
  vote: Vote;
  totalFounders: number;
}

function ProposalRow({ vote, totalFounders }: ProposalRowProps) {
  const totalVotes = vote.total_votes ?? 0;
  const participationPct = totalFounders > 0 ? (totalVotes / totalFounders) * 100 : 0;
  const isActive = vote.status === 'active';
  const hasVoted = vote.has_voted;

  const leadingOption = vote.results
    ? Object.entries(vote.results)
        .sort(([, a], [, b]) => b - a)
        .map(([id, count]) => {
          const opt = vote.options.find((o) => o.id === id);
          return { label: opt?.label ?? id, count, pct: totalVotes > 0 ? (count / totalVotes) * 100 : 0 };
        })[0]
    : null;

  return (
    <Link
      to={`/founders/votes/${vote.id}`}
      className={`vote-hub__row ${isActive && !hasVoted ? 'vote-hub__row--needs-vote' : ''}`}
    >
      <div className="vote-hub__row-left">
        <div className="vote-hub__row-status">
          {isActive ? (
            <Clock size={14} className="vote-hub__row-status-icon vote-hub__row-status-icon--active" />
          ) : hasVoted ? (
            <CheckCircle2 size={14} className="vote-hub__row-status-icon vote-hub__row-status-icon--done" />
          ) : (
            <XCircle size={14} className="vote-hub__row-status-icon" />
          )}
        </div>
        <div className="vote-hub__row-content">
          <h3 className="vote-hub__row-title">{vote.title}</h3>
          <div className="vote-hub__row-meta">
            <span className={`vote-panel__status vote-panel__status--${vote.status}`}>
              {vote.status}
            </span>
            <span className="vote-hub__row-votes">
              {totalVotes} vote{totalVotes !== 1 ? 's' : ''}
            </span>
            {isActive && !hasVoted && (
              <span className="vote-hub__row-action-badge">Vote now</span>
            )}
            {hasVoted && (
              <span className="vote-hub__row-voted-badge">
                <CheckCircle2 size={10} /> Voted
              </span>
            )}
          </div>
        </div>
      </div>

      <div className="vote-hub__row-right">
        {leadingOption && (
          <div className="vote-hub__row-leading">
            <span className="vote-hub__row-leading-label">{leadingOption.label}</span>
            <span className="vote-hub__row-leading-pct">{leadingOption.pct.toFixed(0)}%</span>
          </div>
        )}
        <div className="vote-hub__row-participation">
          <div className="vote-hub__row-participation-bar">
            <div
              className="vote-hub__row-participation-fill"
              style={{ width: `${Math.min(participationPct, 100)}%` }}
            />
          </div>
          <span className="vote-hub__row-participation-label">
            {participationPct.toFixed(0)}% participation
          </span>
        </div>
        <ChevronRight size={16} className="vote-hub__row-chevron" />
      </div>
    </Link>
  );
}
