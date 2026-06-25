import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { reputationApi, type ReputationScore } from '@/api/reputation';
import { identityApi } from '@/api/identity';
import { employeesApi } from '@/api/employees';
import { Trophy, Medal, Star, ChevronRight } from 'lucide-react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts';

const categories = ['Engineering', 'Innovation', 'Leadership', 'Reliability', 'Mentorship', 'Collaboration'];

const categoryColors: Record<string, string> = {
  Engineering: 'bg-blue-500/20 text-blue-400',
  Innovation: 'bg-yellow-500/20 text-yellow-400',
  Leadership: 'bg-purple-500/20 text-purple-400',
  Reliability: 'bg-green-500/20 text-green-400',
  Mentorship: 'bg-pink-500/20 text-pink-400',
  Collaboration: 'bg-cyan-500/20 text-cyan-400',
};

const chartLineColors: Record<string, string> = {
  Engineering: '#60A5FA',
  Innovation: '#FBBF24',
  Leadership: '#A78BFA',
  Reliability: '#34D399',
  Mentorship: '#F472B6',
  Collaboration: '#22D3EE',
};

const rankIcons = (rank: number) => {
  if (rank === 1) return <Medal className="h-5 w-5 text-yellow-400" />;
  if (rank === 2) return <Medal className="h-5 w-5 text-gray-400" />;
  if (rank === 3) return <Medal className="h-5 w-5 text-amber-600" />;
  return <span className="text-sm text-gray-500">#{rank}</span>;
};

export function ReputationPage() {
  const [activeCategory, setActiveCategory] = useState('Engineering');

  const { data: leaderboardData, isLoading: lbLoading } = useQuery({
    queryKey: ['reputation', 'leaderboard', activeCategory],
    queryFn: () => reputationApi.getLeaderboard(activeCategory),
  });

  const { data: myScoresData } = useQuery({
    queryKey: ['reputation', 'me'],
    queryFn: () => reputationApi.get('me'),
  });

  const { data: employeeData } = useQuery({
    queryKey: ['employee', 'me'],
    queryFn: () => employeesApi.list({ limit: 1 }).then((r) => ({ data: { employee: r.data.employees[0] } })),
  });

  const employeeId = employeeData?.data?.employee?.id || '';

  const { data: trendsData } = useQuery({
    queryKey: ['reputation-trends', employeeId],
    queryFn: () => identityApi.getReputationTrends(employeeId),
    enabled: !!employeeId,
  });

  const leaderboard = leaderboardData?.data?.leaderboard || [];
  const myScores = myScoresData?.data?.scores || [];
  const myActiveScore = myScores.find((s) => s.category === activeCategory);
  const trends = trendsData?.data?.trends || [];

  const trendChartData = (() => {
    const dateMap: Record<string, Record<string, number>> = {};
    for (const trend of trends) {
      for (const point of trend.history) {
        const dateKey = new Date(point.recorded_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
        if (!dateMap[dateKey]) dateMap[dateKey] = {};
        dateMap[dateKey][trend.category] = point.score;
      }
    }
    return Object.entries(dateMap)
      .map(([date, scores]) => ({ date, ...scores }))
      .slice(-12);
  })();

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Trophy className="h-6 w-6 text-yellow-400" />
        <h1 className="text-2xl font-bold">Reputation</h1>
      </div>

      <div className="flex flex-wrap gap-2">
        {categories.map((cat) => (
          <button
            key={cat}
            onClick={() => setActiveCategory(cat)}
            className={`rounded-lg px-4 py-2 text-sm font-medium transition-colors ${
              activeCategory === cat
                ? 'bg-blue-600 text-white'
                : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
            }`}
          >
            {cat}
          </button>
        ))}
      </div>

      {trendChartData.length > 0 && (
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
          <h3 className="mb-4 text-sm font-medium text-gray-300">Reputation Trends</h3>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={trendChartData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#1F2937" />
                <XAxis dataKey="date" tick={{ fill: '#9CA3AF', fontSize: 11 }} stroke="#374151" />
                <YAxis tick={{ fill: '#9CA3AF', fontSize: 11 }} stroke="#374151" />
                <Tooltip
                  contentStyle={{
                    backgroundColor: '#111827',
                    border: '1px solid #374151',
                    borderRadius: '8px',
                    fontSize: '12px',
                  }}
                  labelStyle={{ color: '#D1D5DB' }}
                />
                <Legend wrapperStyle={{ fontSize: '11px' }} />
                {categories.map((cat) => (
                  <Line
                    key={cat}
                    type="monotone"
                    dataKey={cat}
                    stroke={chartLineColors[cat]}
                    strokeWidth={2}
                    dot={false}
                    connectNulls
                  />
                ))}
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
            <h3 className="mb-4 text-sm font-medium text-gray-300">{activeCategory} Leaderboard</h3>
            {lbLoading ? (
              <div className="flex justify-center py-8">
                <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
              </div>
            ) : leaderboard.length === 0 ? (
              <p className="py-8 text-center text-sm text-gray-500">No leaderboard data</p>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-gray-800 text-left text-xs text-gray-500">
                      <th className="pb-3 pr-4 w-16">Rank</th>
                      <th className="pb-3 pr-4">Employee</th>
                      <th className="pb-3 pr-4">Score</th>
                      <th className="pb-3">Percentile</th>
                    </tr>
                  </thead>
                  <tbody>
                    {leaderboard.map((entry) => (
                      <tr key={entry.employee_id} className="border-b border-gray-800/50">
                        <td className="py-3 pr-4">{rankIcons(entry.rank)}</td>
                        <td className="py-3 pr-4 font-medium text-gray-100">{entry.ffid}</td>
                        <td className="py-3 pr-4">
                          <div className="flex items-center gap-2">
                            <span className="text-gray-100">{entry.score}</span>
                            <Star className="h-3 w-3 text-yellow-400" />
                          </div>
                        </td>
                        <td className="py-3">
                          <div className="flex items-center gap-2">
                            <div className="h-1.5 w-20 rounded-full bg-gray-800">
                              <div className="h-full rounded-full bg-blue-500" style={{ width: `${entry.rank <= 10 ? (11 - entry.rank) * 10 : Math.max(10, 100 - entry.rank)}%` }} />
                            </div>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>

        <div className="space-y-4">
          <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
            <h3 className="mb-4 text-sm font-medium text-gray-300">My Reputation</h3>
            {myScores.length === 0 ? (
              <p className="py-4 text-center text-sm text-gray-500">No reputation data</p>
            ) : (
              <div className="space-y-3">
                {myScores.map((score) => (
                  <div key={score.id} className="flex items-center justify-between rounded-lg bg-gray-800/50 px-3 py-2">
                    <span className="text-sm text-gray-300">{score.category}</span>
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium text-gray-100">{score.score}</span>
                      <span className="text-xs text-gray-500">P{score.percentile}</span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {myActiveScore && (
            <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
              <h3 className="mb-4 text-sm font-medium text-gray-300">{activeCategory} Breakdown</h3>
              <div className="space-y-3">
                {Object.entries(myActiveScore.components).map(([key, value]) => (
                  <div key={key}>
                    <div className="mb-1 flex items-center justify-between text-xs">
                      <span className="text-gray-400 capitalize">{key.replace(/_/g, ' ')}</span>
                      <span className="text-gray-300">{value}</span>
                    </div>
                    <div className="h-1.5 rounded-full bg-gray-800">
                      <div className="h-full rounded-full bg-blue-500" style={{ width: `${Math.min(value, 100)}%` }} />
                    </div>
                  </div>
                ))}
              </div>
              <div className="mt-4 flex items-center justify-between border-t border-gray-800 pt-3">
                <span className="text-sm text-gray-400">Overall</span>
                <div className="flex items-center gap-2">
                  <span className="text-lg font-bold text-gray-100">{myActiveScore.score}</span>
                  <span className="text-xs text-gray-500">Rank #{myActiveScore.rank}</span>
                </div>
              </div>
            </div>
          )}

          <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
            <h3 className="mb-3 text-sm font-medium text-gray-300">Category Legend</h3>
            <div className="space-y-2">
              {categories.map((cat) => (
                <div key={cat} className="flex items-center gap-2">
                  <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${categoryColors[cat]}`}>
                    {cat}
                  </span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
