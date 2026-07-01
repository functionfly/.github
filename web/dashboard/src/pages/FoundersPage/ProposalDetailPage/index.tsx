import { useState, useCallback } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { usePageTitle } from '@/hooks';
import { useFounderConsole } from '../hooks/useFounderConsole';
import { useProposalDetail, type ChangeDiff, type DiffEntry } from '../hooks/useProposalDetail';
import {
  ArrowLeft,
  CheckCircle2,
  Circle,
  Clock,
  Shield,
  Vote as VoteIcon,
  AlertCircle,
  Loader2,
  TrendingUp,
  FileDiff,
  Lightbulb,
  Zap,
} from 'lucide-react';

import '@/styles/sc-founders.css';

export default function ProposalDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  usePageTitle('Proposal');

  const { proposal, isLoading, error, castVote, isVoting } = useProposalDetail(id);
  const { status: founderStatus } = useFounderConsole();

  const [selectedOption, setSelectedOption] = useState<string | null>(null);
  const [voteSuccess, setVoteSuccess] = useState(false);

  const handleVote = useCallback(
    async (optionId: string) => {
      if (!proposal?.status || proposal.status !== 'active' || proposal.has_voted || isVoting) return;
      setSelectedOption(optionId);
      try {
        await castVote(optionId);
        setVoteSuccess(true);
      } catch {
        setSelectedOption(null);
      }
    },
    [proposal, isVoting, castVote]
  );

  if (isLoading) {
    return (
      <div className="founders-page">
        <div className="founders-loading">
          <div className="founders-loading__spinner" />
        </div>
      </div>
    );
  }

  if (error || !proposal) {
    return (
      <div className="founders-page">
        <div className="founders-error">
          <div className="founders-error__icon">
            <AlertCircle size={24} />
          </div>
          <p className="founders-error__message">{error || 'Proposal not found'}</p>
          <Link to="/founders" className="proposal-detail__back-link">
            Back to Founders
          </Link>
        </div>
      </div>
    );
  }

  const totalVotes = proposal.total_votes ?? 0;
  const totalFounders = founderStatus?.total_founders ?? 10000;
  const participationPct = totalFounders > 0 ? (totalVotes / totalFounders) * 100 : 0;
  const quorumPct = proposal.quorum > 0 ? (totalVotes / proposal.quorum) * 100 : 0;
  const quorumMet = proposal.quorum === 0 || totalVotes >= proposal.quorum;
  const isActive = proposal.status === 'active';
  const hasVoted = proposal.has_voted || voteSuccess;
  const showResults = hasVoted || !isActive;

  const winningOption = showResults && proposal.results
    ? Object.entries(proposal.results).sort(([, a], [, b]) => b - a)[0]
    : null;

  return (
    <div className="founders-page">
      {/* Back nav */}
      <div className="proposal-detail__nav">
        <button
          className="proposal-detail__back-btn"
          onClick={() => navigate('/founders')}
        >
          <ArrowLeft size={16} />
          Back to Founders Console
        </button>
        <div className="proposal-detail__breadcrumb">
          <Link to="/founders">Founders</Link>
          <span>/</span>
          <span>Proposal</span>
        </div>
      </div>

      {/* Header */}
      <div className="proposal-detail__header">
        <div className="proposal-detail__header-top">
          <span className={`vote-panel__status vote-panel__status--${proposal.status}`}>
            {proposal.status}
          </span>
          {proposal.change_diff?.category && (
            <span className="proposal-detail__category">
              {proposal.change_diff.category}
            </span>
          )}
          <span className="proposal-detail__date">
            <Clock size={12} />
            {new Date(proposal.created_at).toLocaleDateString('en-US', {
              year: 'numeric',
              month: 'long',
              day: 'numeric',
            })}
          </span>
        </div>
        <h1 className="proposal-detail__title">{proposal.title}</h1>
        {proposal.description && (
          <p className="proposal-detail__description">{proposal.description}</p>
        )}
      </div>

      {/* Stats bar */}
      <div className="proposal-detail__stats">
        <div className="proposal-detail__stat">
          <VoteIcon size={14} />
          <span className="proposal-detail__stat-value">{totalVotes}</span>
          <span className="proposal-detail__stat-label">votes cast</span>
        </div>
        <div className="proposal-detail__stat">
          <TrendingUp size={14} />
          <span className="proposal-detail__stat-value">{participationPct.toFixed(1)}%</span>
          <span className="proposal-detail__stat-label">participation</span>
        </div>
        {proposal.quorum > 0 && (
          <div className="proposal-detail__stat">
            <Shield size={14} />
            <span className="proposal-detail__stat-value">
              {quorumMet ? (
                <span className="proposal-detail__quorum-met">Met</span>
              ) : (
                `${quorumPct.toFixed(0)}%`
              )}
            </span>
            <span className="proposal-detail__stat-label">
              quorum ({proposal.quorum} votes)
            </span>
          </div>
        )}
        <div className="proposal-detail__stat">
          <Zap size={14} />
          <span className="proposal-detail__stat-value">{proposal.options.length}</span>
          <span className="proposal-detail__stat-label">options</span>
        </div>
      </div>

      {/* Change diff section */}
      {proposal.change_diff && hasContent(proposal.change_diff) && (
        <div className="proposal-detail__section">
          <div className="proposal-detail__section-header">
            <FileDiff size={16} />
            <h2>Proposed Changes</h2>
          </div>

          {proposal.change_diff.summary && (
            <div className="proposal-detail__diff-summary">
              {proposal.change_diff.summary}
            </div>
          )}

          {proposal.change_diff.changes && proposal.change_diff.changes.length > 0 && (
            <div className="proposal-detail__diff-table">
              {proposal.change_diff.changes.map((entry, i) => (
                <DiffRow key={i} entry={entry} />
              ))}
            </div>
          )}

          {proposal.change_diff.impact && (
            <div className="proposal-detail__impact">
              <div className="proposal-detail__impact-label">
                <TrendingUp size={14} />
                Expected Impact
              </div>
              <p>{proposal.change_diff.impact}</p>
            </div>
          )}

          {proposal.change_diff.rationale && (
            <div className="proposal-detail__rationale">
              <div className="proposal-detail__rationale-label">
                <Lightbulb size={14} />
                Rationale
              </div>
              <p>{proposal.change_diff.rationale}</p>
            </div>
          )}
        </div>
      )}

      {/* Voting section */}
      <div className="proposal-detail__section">
        <div className="proposal-detail__section-header">
          <VoteIcon size={16} />
          <h2>Cast Your Vote</h2>
          {hasVoted && (
            <span className="proposal-detail__voted-badge">
              <CheckCircle2 size={14} />
              Vote recorded
            </span>
          )}
        </div>

        {/* Participation bar */}
        <div className="proposal-detail__participation">
          <div className="proposal-detail__participation-header">
            <span>{totalVotes} of {totalFounders.toLocaleString()} founders voted</span>
            <span>{participationPct.toFixed(1)}%</span>
          </div>
          <div className="proposal-detail__participation-bar">
            <div
              className="proposal-detail__participation-fill"
              style={{ width: `${Math.min(participationPct, 100)}%` }}
            />
            {proposal.quorum > 0 && (
              <div
                className="proposal-detail__quorum-line"
                style={{ left: `${Math.min((proposal.quorum / totalFounders) * 100, 100)}%` }}
                title={`Quorum: ${proposal.quorum} votes`}
              />
            )}
          </div>
        </div>

        {/* Options */}
        <div className="proposal-detail__options">
          {proposal.options.map((opt) => {
            const voteCount = proposal.results?.[opt.id] ?? 0;
            const pct = totalVotes > 0 ? (voteCount / totalVotes) * 100 : 0;
            const isMyVote = proposal.my_vote === opt.id || selectedOption === opt.id;
            const isWinner = winningOption && winningOption[0] === opt.id && showResults;

            if (showResults) {
              return (
                <div
                  key={opt.id}
                  className={`proposal-detail__result ${isMyVote ? 'proposal-detail__result--mine' : ''} ${isWinner ? 'proposal-detail__result--winner' : ''}`}
                >
                  <div className="proposal-detail__result-header">
                    <span className="proposal-detail__result-label">
                      {isMyVote && <CheckCircle2 size={14} />}
                      {opt.label}
                      {isMyVote && <span className="vote-panel__your-tag">YOUR VOTE</span>}
                      {isWinner && <span className="proposal-detail__winner-tag">LEADING</span>}
                    </span>
                    <span className="proposal-detail__result-pct">{pct.toFixed(1)}%</span>
                  </div>
                  <div className="proposal-detail__result-bar">
                    <div
                      className="proposal-detail__result-fill"
                      style={{ width: `${pct}%` }}
                    />
                  </div>
                  <span className="proposal-detail__result-count">{voteCount} votes</span>
                </div>
              );
            }

            return (
              <button
                key={opt.id}
                className={`proposal-detail__option ${selectedOption === opt.id ? 'proposal-detail__option--selected' : ''}`}
                onClick={() => handleVote(opt.id)}
                disabled={isVoting || hasVoted}
              >
                <Circle size={18} className="proposal-detail__option-radio" />
                <span className="proposal-detail__option-label">{opt.label}</span>
                {isVoting && selectedOption === opt.id && (
                  <Loader2 size={16} className="animate-spin" />
                )}
              </button>
            );
          })}
        </div>

        {voteSuccess && (
          <div className="proposal-detail__vote-success">
            <CheckCircle2 size={16} />
            Your vote has been recorded on-chain. Results are now visible.
          </div>
        )}
      </div>
    </div>
  );
}

function DiffRow({ entry }: { entry: DiffEntry }) {
  const typeClass = entry.type || 'modified';
  return (
    <div className={`proposal-detail__diff-row proposal-detail__diff-row--${typeClass}`}>
      <div className="proposal-detail__diff-field">
        <span className={`proposal-detail__diff-badge proposal-detail__diff-badge--${typeClass}`}>
          {typeClass === 'added' ? '+' : typeClass === 'removed' ? '−' : '~'}
        </span>
        {entry.label || entry.field}
      </div>
      <div className="proposal-detail__diff-values">
        {entry.before && (
          <div className="proposal-detail__diff-before">
            <span className="proposal-detail__diff-label">Before</span>
            <code>{entry.before}</code>
          </div>
        )}
        {entry.after && (
          <div className="proposal-detail__diff-after">
            <span className="proposal-detail__diff-label">After</span>
            <code>{entry.after}</code>
          </div>
        )}
      </div>
    </div>
  );
}

function hasContent(diff: ChangeDiff): boolean {
  return !!(
    diff.summary ||
    (diff.changes && diff.changes.length > 0) ||
    diff.impact ||
    diff.rationale
  );
}
