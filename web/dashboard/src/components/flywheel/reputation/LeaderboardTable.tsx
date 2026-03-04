/**
 * LeaderboardTable - Rankings display
 */

import { Link } from 'react-router-dom';
import { cn } from '@/lib/utils';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { ReputationBadgeInline } from './ReputationBadge';
import type { LeaderboardEntry } from '../types';

interface LeaderboardTableProps {
  entries: LeaderboardEntry[];
  myRank?: LeaderboardEntry;
  compact?: boolean;
  showTrend?: boolean;
  className?: string;
}

export function LeaderboardTable({
  entries,
  myRank,
  compact = false,
  showTrend = true,
  className,
}: LeaderboardTableProps) {
  return (
    <div className={cn('rounded-xl border border-slate-800 bg-slate-900 overflow-hidden', className)}>
      <table className="w-full">
        <thead>
          <tr className="border-b border-slate-800 bg-slate-950/50">
            <th className="px-4 py-3 text-left text-xs font-medium text-slate-500">Rank</th>
            <th className="px-4 py-3 text-left text-xs font-medium text-slate-500">User</th>
            {!compact && (
              <>
                <th className="px-4 py-3 text-left text-xs font-medium text-slate-500">Tier</th>
                {showTrend && <th className="px-4 py-3 text-center text-xs font-medium text-slate-500">Trend</th>}
              </>
            )}
            <th className="px-4 py-3 text-right text-xs font-medium text-slate-500">Score</th>
          </tr>
        </thead>
        <tbody>
          {entries.map((entry, index) => (
            <LeaderboardRow
              key={entry.user.id}
              entry={entry}
              compact={compact}
              showTrend={showTrend}
              isTopThree={index < 3}
            />
          ))}

          {myRank && !entries.find(e => e.user.id === myRank.user.id) && (
            <>
              <tr className="border-t border-slate-800">
                <td colSpan={compact ? 3 : showTrend ? 5 : 4} className="py-2 text-center text-xs text-slate-600">
                  • • •
                </td>
              </tr>
              <LeaderboardRow
                entry={myRank}
                compact={compact}
                showTrend={showTrend}
                isHighlighted
              />
            </>
          )}
        </tbody>
      </table>
    </div>
  );
}

function LeaderboardRow({
  entry,
  compact,
  showTrend,
  isTopThree,
  isHighlighted,
}: {
  entry: LeaderboardEntry;
  compact: boolean;
  showTrend: boolean;
  isTopThree?: boolean;
  isHighlighted?: boolean;
}) {
  const rankIcons = ['🥇', '🥈', '🥉'];

  const trendIcons = {
    up: { icon: '↑', color: 'text-emerald-400' },
    down: { icon: '↓', color: 'text-red-400' },
    same: { icon: '→', color: 'text-slate-500' },
  };
  const trend = trendIcons[entry.trend];

  return (
    <tr
      className={cn(
        'border-b border-slate-800/50 transition-colors hover:bg-slate-800/50',
        isHighlighted && 'bg-indigo-500/10'
      )}
    >
      <td className="px-4 py-3">
        <div className="flex items-center gap-2">
          {entry.rank <= 3 ? (
            <span className="text-xl">{rankIcons[entry.rank - 1]}</span>
          ) : (
            <span className="w-6 text-center text-sm text-slate-500">{entry.rank}</span>
          )}
        </div>
      </td>

      <td className="px-4 py-3">
        <Link
          to={`/flywheel/reputation/${entry.user.id}`}
          className="flex items-center gap-2 hover:opacity-80"
        >
          <Avatar className="h-8 w-8 border border-slate-800">
            <AvatarImage src={entry.user.avatarUrl} alt={entry.user.username} />
            <AvatarFallback className="bg-slate-800 text-xs text-slate-400">
              {entry.user.username.slice(0, 2).toUpperCase()}
            </AvatarFallback>
          </Avatar>
          <div>
            <p className="font-medium text-slate-200">@{entry.user.username}</p>
            {compact && entry.user.reputation && (
              <ReputationBadgeInline
                score={entry.user.reputation.overallScore}
                type="overall"
                tier={entry.user.reputation.tier.level}
              />
            )}
          </div>
        </Link>
      </td>

      {!compact && (
        <>
          <td className="px-4 py-3">
            <div className="flex items-center gap-1">
              {Array.from({ length: entry.tier }).map((_, i) => (
                <span key={i} className="text-amber-400 text-xs">⭐</span>
              ))}
            </div>
          </td>

          {showTrend && (
            <td className="px-4 py-3 text-center">
              {entry.previousRank && entry.previousRank !== entry.rank && (
                <span className={cn('text-sm font-medium', trend.color)}>
                  {trend.icon}
                  {Math.abs(entry.previousRank - entry.rank)}
                </span>
              )}
            </td>
          )}
        </>
      )}

      <td className="px-4 py-3 text-right">
        <span className="font-mono font-medium text-white">
          {entry.score.toLocaleString()}
        </span>
      </td>
    </tr>
  );
}

/**
 * Compact leaderboard for sidebars
 */
export function LeaderboardCompact({
  entries,
  className,
}: {
  entries: LeaderboardEntry[];
  className?: string;
}) {
  return (
    <div className={cn('space-y-2', className)}>
      {entries.map((entry, index) => (
        <Link
          key={entry.user.id}
          to={`/flywheel/reputation/${entry.user.id}`}
          className="flex items-center gap-3 rounded-lg border border-slate-800 bg-slate-900 p-2 transition-colors hover:border-slate-700"
        >
          <span className="w-6 text-center text-sm font-medium text-slate-500">
            {index < 3 ? ['🥇', '🥈', '🥉'][index] : entry.rank}
          </span>
          <Avatar className="h-8 w-8 border border-slate-800">
            <AvatarImage src={entry.user.avatarUrl} alt={entry.user.username} />
            <AvatarFallback className="bg-slate-800 text-xs text-slate-400">
              {entry.user.username.slice(0, 2).toUpperCase()}
            </AvatarFallback>
          </Avatar>
          <div className="min-w-0 flex-1">
            <p className="truncate text-sm font-medium text-slate-200">
              @{entry.user.username}
            </p>
          </div>
          <span className="font-mono text-sm font-medium text-white">
            {entry.score.toLocaleString()}
          </span>
        </Link>
      ))}
    </div>
  );
}

/**
 * Leaderboard podium for top 3
 */
export function LeaderboardPodium({ entries }: { entries: LeaderboardEntry[] }) {
  const topThree = entries.slice(0, 3);
  const [second, first, third] = [
    topThree[1],
    topThree[0],
    topThree[2],
  ];

  return (
    <div className="flex items-end justify-center gap-4 py-8">
      {/* 2nd Place */}
      {second && (
        <div className="flex flex-col items-center">
          <div className="relative">
            <Avatar className="h-16 w-16 border-2 border-slate-400 ring-4 ring-slate-400/20">
              <AvatarImage src={second.user.avatarUrl} alt={second.user.username} />
              <AvatarFallback className="bg-slate-800 text-slate-400">
                {second.user.username.slice(0, 2).toUpperCase()}
              </AvatarFallback>
            </Avatar>
            <span className="absolute -top-2 left-1/2 -translate-x-1/2 text-2xl">🥈</span>
          </div>
          <div className="mt-3 h-24 w-24 rounded-t-lg bg-gradient-to-t from-slate-400 to-slate-300 flex items-end justify-center pb-2">
            <div className="text-center">
              <p className="text-xs font-medium text-slate-800">@{second.user.username}</p>
              <p className="font-mono text-sm font-bold text-slate-900">
                {second.score.toLocaleString()}
              </p>
            </div>
          </div>
        </div>
      )}

      {/* 1st Place */}
      {first && (
        <div className="flex flex-col items-center">
          <div className="relative">
            <Avatar className="h-20 w-20 border-2 border-amber-400 ring-4 ring-amber-400/20">
              <AvatarImage src={first.user.avatarUrl} alt={first.user.username} />
              <AvatarFallback className="bg-slate-800 text-slate-400">
                {first.user.username.slice(0, 2).toUpperCase()}
              </AvatarFallback>
            </Avatar>
            <span className="absolute -top-2 left-1/2 -translate-x-1/2 text-3xl">🥇</span>
          </div>
          <div className="mt-3 h-32 w-28 rounded-t-lg bg-gradient-to-t from-amber-500 to-amber-300 flex items-end justify-center pb-2">
            <div className="text-center">
              <p className="text-xs font-medium text-amber-900">@{first.user.username}</p>
              <p className="font-mono text-lg font-bold text-amber-900">
                {first.score.toLocaleString()}
              </p>
            </div>
          </div>
        </div>
      )}

      {/* 3rd Place */}
      {third && (
        <div className="flex flex-col items-center">
          <div className="relative">
            <Avatar className="h-16 w-16 border-2 border-orange-600 ring-4 ring-orange-600/20">
              <AvatarImage src={third.user.avatarUrl} alt={third.user.username} />
              <AvatarFallback className="bg-slate-800 text-slate-400">
                {third.user.username.slice(0, 2).toUpperCase()}
              </AvatarFallback>
            </Avatar>
            <span className="absolute -top-2 left-1/2 -translate-x-1/2 text-2xl">🥉</span>
          </div>
          <div className="mt-3 h-16 w-24 rounded-t-lg bg-gradient-to-t from-orange-700 to-orange-600 flex items-end justify-center pb-2">
            <div className="text-center">
              <p className="text-xs font-medium text-orange-100">@{third.user.username}</p>
              <p className="font-mono text-sm font-bold text-orange-100">
                {third.score.toLocaleString()}
              </p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
