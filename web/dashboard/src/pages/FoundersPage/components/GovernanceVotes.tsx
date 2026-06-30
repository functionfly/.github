import { Vote as VoteIcon } from 'lucide-react';
import { useCallback, useState } from 'react';
import type { Vote, VoteOption } from '../hooks/useFounderConsole';

interface GovernanceVotesProps {
  votes: Vote[];
  castVote: (voteId: string, optionId: string) => Promise<void>;
  totalFounders: number;
}

export function GovernanceVotes({ votes, castVote, totalFounders }: GovernanceVotesProps) {
  if (votes.length === 0) return null;

  return (
    <section className="founders-section">
      <div className="founders-section__title">
        <VoteIcon size={14} />
        Governance Votes
      </div>
      <div className="vote-grid">
        {votes.map((vote) => (
          <VoteCard
            key={vote.id}
            vote={vote}
            castVote={castVote}
            totalFounders={totalFounders}
          />
        ))}
      </div>
    </section>
  );
}

interface VoteCardProps {
  vote: Vote;
  castVote: (voteId: string, optionId: string) => Promise<void>;
  totalFounders: number;
}

function VoteCard({ vote, castVote, totalFounders }: VoteCardProps) {
  const [selectedOption, setSelectedOption] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showResults, setShowResults] = useState(vote.has_voted);

  const totalVotes = vote.total_votes ?? 0;

  const handleVote = useCallback(
    async (optionId: string) => {
      if (vote.has_voted || isSubmitting) return;
      setIsSubmitting(true);
      setSelectedOption(optionId);
      try {
        await castVote(vote.id, optionId);
        setShowResults(true);
      } catch {
        setSelectedOption(null);
      } finally {
        setIsSubmitting(false);
      }
    },
    [vote.id, vote.has_voted, isSubmitting, castVote]
  );

  return (
    <div className="vote-panel">
      <div className="vote-panel__header">
        <h3 className="vote-panel__title">{vote.title}</h3>
        <span className={`vote-panel__status vote-panel__status--${vote.status}`}>
          {vote.status}
        </span>
      </div>
      {vote.description && (
        <p className="vote-panel__description">{vote.description}</p>
      )}
      <div className="vote-panel__options">
        {vote.options.map((opt) => (
          <VoteOptionRow
            key={opt.id}
            option={opt}
            vote={vote}
            selectedOption={selectedOption}
            isSubmitting={isSubmitting}
            showResults={showResults}
            results={vote.results}
            totalVotes={totalVotes}
            onVote={handleVote}
          />
        ))}
      </div>
      <div className="vote-panel__footer">
        <span className="vote-panel__meta">
          {totalVotes} of {totalFounders.toLocaleString()} founders voted
        </span>
        {vote.has_voted && (
          <span className="vote-panel__voted-badge">
            <VoteIcon size={12} /> Your vote is recorded
          </span>
        )}
      </div>
    </div>
  );
}

interface VoteOptionRowProps {
  option: VoteOption;
  vote: Vote;
  selectedOption: string | null;
  isSubmitting: boolean;
  showResults: boolean;
  results?: Record<string, number>;
  totalVotes: number;
  onVote: (optionId: string) => void;
}

function VoteOptionRow({
  option,
  vote,
  selectedOption,
  isSubmitting,
  showResults,
  results,
  totalVotes,
  onVote,
}: VoteOptionRowProps) {
  const isMyVote = vote.my_vote === option.id || selectedOption === option.id;
  const voteCount = results?.[option.id] ?? 0;
  const pct = totalVotes > 0 ? (voteCount / totalVotes) * 100 : 0;

  if (showResults) {
    return (
      <div className={`vote-panel__result ${isMyVote ? 'vote-panel__result--mine' : ''}`}>
        <div className="vote-panel__result-header">
          <span className="vote-panel__result-label">
            {option.label}
            {isMyVote && <span className="vote-panel__your-tag">YOUR VOTE</span>}
          </span>
          <span className="vote-panel__result-pct">{pct.toFixed(0)}%</span>
        </div>
        <div className="vote-panel__result-bar">
          <div
            className="vote-panel__result-fill"
            style={{ width: `${pct}%` }}
          />
        </div>
        <span className="vote-panel__result-count">{voteCount} votes</span>
      </div>
    );
  }

  return (
    <button
      className={`vote-panel__option ${selectedOption === option.id ? 'vote-panel__option--selected' : ''}`}
      onClick={() => onVote(option.id)}
      disabled={isSubmitting || vote.has_voted}
    >
      <span className="vote-panel__option-label">{option.label}</span>
      {isSubmitting && selectedOption === option.id && (
        <span className="vote-panel__option-spinner" />
      )}
    </button>
  );
}
