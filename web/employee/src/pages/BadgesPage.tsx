import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { badgesApi, type DigitalBadge } from '@/api/badges';
import { identityApi, type Achievement } from '@/api/identity';
import { Award, Plus, Star, Search, Trophy, Lock } from 'lucide-react';

const categoryColors: Record<string, string> = {
  achievement: 'bg-yellow-500/20 text-yellow-400',
  milestone: 'bg-blue-500/20 text-blue-400',
  skill: 'bg-purple-500/20 text-purple-400',
  collaboration: 'bg-green-500/20 text-green-400',
  innovation: 'bg-pink-500/20 text-pink-400',
  leadership: 'bg-cyan-500/20 text-cyan-400',
};

const tierColors: Record<number, string> = {
  1: 'border-gray-600',
  2: 'border-green-600/50',
  3: 'border-blue-600/50',
  4: 'border-purple-600/50',
  5: 'border-yellow-500/50',
};

const tierLabels: Record<number, string> = {
  1: 'Common',
  2: 'Uncommon',
  3: 'Rare',
  4: 'Epic',
  5: 'Legendary',
};

export function BadgesPage() {
  const queryClient = useQueryClient();
  const [categoryFilter, setCategoryFilter] = useState('');
  const [search, setSearch] = useState('');
  const [showAward, setShowAward] = useState(false);
  const [awardForm, setAwardForm] = useState({ employeeId: '', badgeId: '' });
  const [view, setView] = useState<'badges' | 'achievements'>('badges');

  const { data: badgesData, isLoading } = useQuery({
    queryKey: ['badges'],
    queryFn: () => badgesApi.list(),
  });

  const { data: myBadgesData } = useQuery({
    queryKey: ['badges', 'mine'],
    queryFn: () => badgesApi.getMyBadges(),
  });

  const { data: achievementsData, isLoading: achievementsLoading } = useQuery({
    queryKey: ['achievements'],
    queryFn: () => identityApi.getAchievements(),
  });

  const awardMutation = useMutation({
    mutationFn: ({ employeeId, badgeId }: { employeeId: string; badgeId: string }) =>
      badgesApi.award(employeeId, badgeId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['badges'] });
      setShowAward(false);
      setAwardForm({ employeeId: '', badgeId: '' });
    },
  });

  const badges = badgesData?.data?.badges || [];
  const myBadges = myBadgesData?.data?.badges || [];
  const myBadgeIds = new Set(myBadges.map((b) => b.badge_id));

  const achievements = achievementsData?.data?.definitions || [];
  const achievementProgress = achievementsData?.data?.progress || [];
  const progressMap = new Map(achievementProgress.map((p) => [p.achievement_id, p]));

  const achievementsWithProgress = achievements.map((a) => {
    const prog = progressMap.get(a.id);
    return {
      ...a,
      earned: prog?.awarded ?? a.earned ?? false,
      progress: prog?.current_value ?? a.progress,
    };
  });

  const categories = [...new Set(badges.map((b) => b.category))].sort();

  const filtered = badges.filter((b) => {
    if (categoryFilter && b.category !== categoryFilter) return false;
    if (search && !b.name.toLowerCase().includes(search.toLowerCase())) return false;
    return true;
  });

  const filteredAchievements = achievementsWithProgress.filter((a) => {
    if (search && !a.name.toLowerCase().includes(search.toLowerCase())) return false;
    return true;
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Award className="h-6 w-6 text-yellow-400" />
          <h1 className="text-2xl font-bold">Digital Badges</h1>
        </div>
        <button
          onClick={() => setShowAward(true)}
          className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          <Plus className="h-4 w-4" />
          Award Badge
        </button>
      </div>

      <div className="flex gap-2">
        <button
          onClick={() => setView('badges')}
          className={`rounded-lg px-4 py-2 text-sm font-medium transition-colors ${
            view === 'badges' ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
          }`}
        >
          Badges
        </button>
        <button
          onClick={() => setView('achievements')}
          className={`rounded-lg px-4 py-2 text-sm font-medium transition-colors ${
            view === 'achievements' ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
          }`}
        >
          Achievements
        </button>
      </div>

      {view === 'badges' && myBadges.length > 0 && (
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
          <h3 className="mb-4 text-sm font-medium text-gray-300">My Earned Badges ({myBadges.length})</h3>
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
            {myBadges.map((eb) => (
              <div key={eb.id} className="flex flex-col items-center rounded-lg bg-gray-800/50 p-3 text-center">
                {eb.badge?.icon_url ? (
                  <img src={eb.badge.icon_url} alt={eb.badge.name} className="mb-2 h-10 w-10" />
                ) : (
                  <Award className="mb-2 h-10 w-10 text-yellow-400" />
                )}
                <span className="text-xs font-medium text-gray-100">{eb.badge?.name || eb.badge_id}</span>
                <span className="mt-1 text-xs text-gray-500">{new Date(eb.awarded_at).toLocaleDateString()}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="flex flex-wrap items-center gap-3">
        <div className="relative flex-1 sm:max-w-xs">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-500" />
          <input
            type="text"
            placeholder="Search..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full rounded-lg border border-gray-700 bg-gray-800 py-2 pl-9 pr-3 text-sm text-gray-100 placeholder-gray-500"
          />
        </div>
        {view === 'badges' && (
          <select
            value={categoryFilter}
            onChange={(e) => setCategoryFilter(e.target.value)}
            className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
          >
            <option value="">All Categories</option>
            {categories.map((cat) => (
              <option key={cat} value={cat}>{cat}</option>
            ))}
          </select>
        )}
      </div>

      {view === 'badges' ? (
        isLoading ? (
          <div className="flex justify-center py-12">
            <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
          </div>
        ) : filtered.length === 0 ? (
          <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
            <Award className="mb-4 h-12 w-12 text-gray-600" />
            <p className="text-gray-400">No badges found</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {filtered.map((badge) => {
              const earned = myBadgeIds.has(badge.id);
              return (
                <div
                  key={badge.id}
                  className={`rounded-xl border p-5 transition-colors ${
                    earned
                      ? 'border-yellow-500/30 bg-yellow-500/5'
                      : 'border-gray-800 bg-gray-900'
                  }`}
                >
                  <div className="mb-3 flex items-start justify-between">
                    <div className="flex items-center gap-3">
                      {badge.icon_url ? (
                        <img src={badge.icon_url} alt={badge.name} className="h-10 w-10" />
                      ) : (
                        <Award className={`h-10 w-10 ${earned ? 'text-yellow-400' : 'text-gray-600'}`} />
                      )}
                      <div>
                        <h3 className="font-semibold text-gray-100">{badge.name}</h3>
                        <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${categoryColors[badge.category] || 'bg-gray-500/20 text-gray-400'}`}>
                          {badge.category}
                        </span>
                      </div>
                    </div>
                    {earned && (
                      <span className="rounded-full bg-yellow-500/20 px-2 py-0.5 text-xs font-medium text-yellow-400">
                        Earned
                      </span>
                    )}
                  </div>
                  {badge.description && (
                    <p className="mb-3 text-sm text-gray-400">{badge.description}</p>
                  )}
                  <div className="flex items-center gap-1 text-xs text-gray-500">
                    <Star className="h-3 w-3" />
                    <span>{badge.points} points</span>
                  </div>
                </div>
              );
            })}
          </div>
        )
      ) : achievementsLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : filteredAchievements.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
          <Trophy className="mb-4 h-12 w-12 text-gray-600" />
          <p className="text-gray-400">No achievements found</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {filteredAchievements.map((achievement) => (
            <div
              key={achievement.id}
              className={`rounded-xl border p-5 transition-colors ${
                tierColors[achievement.tier] || tierColors[1]
              } ${achievement.earned ? 'bg-gray-900' : 'bg-gray-900/80'}`}
            >
              <div className="mb-3 flex items-start justify-between">
                <div className="flex items-center gap-3">
                  {achievement.earned ? (
                    <span className="text-2xl">{achievement.icon || '🏆'}</span>
                  ) : (
                    <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-gray-800">
                      <Lock className="h-5 w-5 text-gray-600" />
                    </div>
                  )}
                  <div>
                    <h3 className="font-semibold text-gray-100">{achievement.name}</h3>
                    <div className="flex items-center gap-2">
                      <span className="text-xs text-gray-500">{tierLabels[achievement.tier] || 'Common'}</span>
                      {achievement.category && (
                        <span className="rounded-full bg-gray-800 px-2 py-0.5 text-xs text-gray-400">
                          {achievement.category}
                        </span>
                      )}
                    </div>
                  </div>
                </div>
                {achievement.earned && (
                  <span className="rounded-full bg-yellow-500/20 px-2 py-0.5 text-xs font-medium text-yellow-400">
                    Earned
                  </span>
                )}
              </div>

              {achievement.description && (
                <p className="mb-3 text-sm text-gray-400">{achievement.description}</p>
              )}

              {achievement.progress !== undefined && achievement.threshold && !achievement.earned && (
                <div className="mb-2">
                  <div className="mb-1 flex items-center justify-between text-xs">
                    <span className="text-gray-500">Progress</span>
                    <span className="text-gray-400">{achievement.progress}/{achievement.threshold}</span>
                  </div>
                  <div className="h-2 rounded-full bg-gray-800">
                    <div
                      className="h-full rounded-full bg-blue-500 transition-all"
                      style={{ width: `${Math.min((achievement.progress / achievement.threshold) * 100, 100)}%` }}
                    />
                  </div>
                </div>
              )}

              <div className="flex items-center justify-between text-xs text-gray-500">
                <span>{achievement.points} points</span>
                {achievement.earned && achievement.earned_at && (
                  <span>{new Date(achievement.earned_at).toLocaleDateString()}</span>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {showAward && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Award Badge</h2>
            <input
              type="text"
              placeholder="Employee ID"
              value={awardForm.employeeId}
              onChange={(e) => setAwardForm({ ...awardForm, employeeId: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              autoFocus
            />
            <select
              value={awardForm.badgeId}
              onChange={(e) => setAwardForm({ ...awardForm, badgeId: e.target.value })}
              className="mb-4 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
            >
              <option value="">Select Badge</option>
              {badges.map((b) => (
                <option key={b.id} value={b.id}>{b.name} ({b.points} pts)</option>
              ))}
            </select>
            <div className="flex justify-end gap-3">
              <button
                onClick={() => { setShowAward(false); setAwardForm({ employeeId: '', badgeId: '' }); }}
                className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200"
              >
                Cancel
              </button>
              <button
                onClick={() => awardMutation.mutate(awardForm)}
                disabled={!awardForm.employeeId.trim() || !awardForm.badgeId}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                Award
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
