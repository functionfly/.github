import { Trophy, Lock } from 'lucide-react';
import type { Achievement } from '@/api/identity';

interface AchievementGridProps {
  achievements: Achievement[];
  className?: string;
}

const tierColors: Record<number, string> = {
  1: 'border-gray-600 bg-gray-800/50',
  2: 'border-green-600/50 bg-green-900/10',
  3: 'border-blue-600/50 bg-blue-900/10',
  4: 'border-purple-600/50 bg-purple-900/10',
  5: 'border-yellow-500/50 bg-yellow-900/10',
};

const tierLabels: Record<number, string> = {
  1: 'Common',
  2: 'Uncommon',
  3: 'Rare',
  4: 'Epic',
  5: 'Legendary',
};

export function AchievementGrid({ achievements, className = '' }: AchievementGridProps) {
  if (achievements.length === 0) {
    return (
      <div className={`rounded-xl border border-gray-800 bg-gray-900 p-5 ${className}`}>
        <h3 className="mb-4 flex items-center gap-2 text-sm font-medium text-gray-300">
          <Trophy className="h-4 w-4 text-yellow-400" />
          Achievements
        </h3>
        <p className="py-4 text-center text-sm text-gray-500">No achievements available</p>
      </div>
    );
  }

  return (
    <div className={`rounded-xl border border-gray-800 bg-gray-900 p-5 ${className}`}>
      <h3 className="mb-4 flex items-center gap-2 text-sm font-medium text-gray-300">
        <Trophy className="h-4 w-4 text-yellow-400" />
        Achievements ({achievements.filter((a) => a.earned).length}/{achievements.length})
      </h3>

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
        {achievements.map((achievement) => (
          <div
            key={achievement.id}
            className={`relative rounded-lg border p-3 transition-colors ${
              tierColors[achievement.tier] || tierColors[1]
            } ${!achievement.earned ? 'opacity-60' : ''}`}
          >
            <div className="flex items-center gap-2">
              {achievement.earned ? (
                <span className="text-lg">{achievement.icon || '🏆'}</span>
              ) : (
                <Lock className="h-4 w-4 text-gray-600" />
              )}
              <div className="min-w-0 flex-1">
                <p className="truncate text-xs font-semibold text-gray-100">{achievement.name}</p>
                <p className="text-xs text-gray-500">{tierLabels[achievement.tier] || 'Common'}</p>
              </div>
            </div>

            {achievement.progress !== undefined && achievement.threshold && !achievement.earned && (
              <div className="mt-2">
                <div className="h-1 rounded-full bg-gray-700">
                  <div
                    className="h-full rounded-full bg-blue-500"
                    style={{ width: `${Math.min((achievement.progress / achievement.threshold) * 100, 100)}%` }}
                  />
                </div>
                <p className="mt-1 text-xs text-gray-500">
                  {achievement.progress}/{achievement.threshold}
                </p>
              </div>
            )}

            {achievement.earned && achievement.earned_at && (
              <p className="mt-1 text-xs text-gray-500">
                {new Date(achievement.earned_at).toLocaleDateString()}
              </p>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
