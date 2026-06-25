import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { careerApi, type CareerPath } from '@/api/career';
import { TrendingUp, Target, Award, ChevronRight, AlertCircle, CheckCircle } from 'lucide-react';

const trackColors: Record<string, string> = {
  engineering: 'bg-blue-500/20 text-blue-400',
  product: 'bg-purple-500/20 text-purple-400',
  design: 'bg-pink-500/20 text-pink-400',
  leadership: 'bg-amber-500/20 text-amber-400',
};

export function CareerPage() {
  const queryClient = useQueryClient();
  const [trackFilter, setTrackFilter] = useState('');
  const [showTarget, setShowTarget] = useState<string | null>(null);
  const [targetDate, setTargetDate] = useState('');

  const { data: pathsData, isLoading: pathsLoading } = useQuery({
    queryKey: ['career', 'paths', trackFilter],
    queryFn: () => careerApi.listPaths(trackFilter ? { track: trackFilter } : undefined),
  });

  const { data: progressData } = useQuery({
    queryKey: ['career', 'progress'],
    queryFn: () => careerApi.getMyProgress(),
  });

  const setTargetMutation = useMutation({
    mutationFn: ({ pathId, date }: { pathId: string; date?: string }) => careerApi.setTarget(pathId, date),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['career'] });
      setShowTarget(null);
      setTargetDate('');
    },
  });

  const paths = pathsData?.data?.paths || [];
  const progress = progressData?.data?.progress || [];

  const grouped = paths.reduce<Record<string, CareerPath[]>>((acc, path) => {
    (acc[path.track] = acc[path.track] || []).push(path);
    return acc;
  }, {});

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Career Navigation</h1>

      {progress.length > 0 && (
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
          <h2 className="mb-4 flex items-center gap-2 text-lg font-semibold">
            <Target className="h-5 w-5 text-blue-400" />
            My Career Targets
          </h2>
          <div className="space-y-3">
            {progress.map((p) => (
              <div key={p.id} className="flex items-center justify-between rounded-lg border border-gray-700 bg-gray-800 p-3">
                <div>
                  <p className="font-medium text-gray-200">{p.career_path_id}</p>
                  <p className="text-sm capitalize text-gray-500">{p.status.replace('_', ' ')}</p>
                </div>
                {p.target_date && (
                  <span className="text-xs text-gray-400">Target: {new Date(p.target_date).toLocaleDateString()}</span>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="flex items-center gap-3">
        <select
          value={trackFilter}
          onChange={(e) => setTrackFilter(e.target.value)}
          className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
        >
          <option value="">All Tracks</option>
          <option value="engineering">Engineering</option>
          <option value="product">Product</option>
          <option value="design">Design</option>
          <option value="leadership">Leadership</option>
        </select>
      </div>

      {pathsLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : Object.keys(grouped).length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
          <TrendingUp className="mb-4 h-12 w-12 text-gray-600" />
          <p className="text-gray-400">No career paths available</p>
        </div>
      ) : (
        Object.entries(grouped).map(([track, trackPaths]) => (
          <div key={track}>
            <h2 className="mb-3 flex items-center gap-2 text-lg font-semibold capitalize">
              <span className={`rounded px-2 py-0.5 text-xs ${trackColors[track] || 'bg-gray-500/20 text-gray-400'}`}>
                {track}
              </span>
            </h2>
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
              {trackPaths.map((path) => (
                <div key={path.id} className="rounded-xl border border-gray-800 bg-gray-900 p-5">
                  <div className="mb-2 flex items-start justify-between">
                    <h3 className="font-semibold text-gray-100">{path.title}</h3>
                    <span className="rounded bg-gray-800 px-2 py-0.5 text-xs text-gray-400">L{path.level}</span>
                  </div>
                  {path.description && (
                    <p className="mb-3 line-clamp-2 text-sm text-gray-400">{path.description}</p>
                  )}
                  {path.requirements?.skills && path.requirements.skills.length > 0 && (
                    <div className="mb-3 flex flex-wrap gap-1.5">
                      {path.requirements.skills.map((skill) => (
                        <span key={skill} className="rounded bg-gray-800 px-2 py-0.5 text-xs text-gray-400">{skill}</span>
                      ))}
                    </div>
                  )}
                  <div className="mb-3 flex items-center gap-3 text-xs text-gray-500">
                    {path.requirements?.years_exp != null && (
                      <span>{path.requirements.years_exp}+ years</span>
                    )}
                    {path.salary_range_min_cents != null && path.salary_range_max_cents != null && (
                      <span>
                        ${(path.salary_range_min_cents / 1000).toFixed(0)}k - ${(path.salary_range_max_cents / 1000).toFixed(0)}k
                      </span>
                    )}
                  </div>
                  <button
                    onClick={() => setShowTarget(path.id)}
                    className="flex w-full items-center justify-center gap-2 rounded-lg border border-blue-600/30 bg-blue-600/10 px-4 py-2 text-sm font-medium text-blue-400 hover:bg-blue-600/20"
                  >
                    <Target className="h-4 w-4" />
                    Set as Target
                  </button>
                </div>
              ))}
            </div>
          </div>
        ))
      )}

      {showTarget && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-sm rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Set Career Target</h2>
            <label className="mb-2 block text-sm text-gray-400">Target date (optional)</label>
            <input
              type="date"
              value={targetDate}
              onChange={(e) => setTargetDate(e.target.value)}
              className="mb-4 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
            />
            <div className="flex justify-end gap-3">
              <button
                onClick={() => { setShowTarget(null); setTargetDate(''); }}
                className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200"
              >
                Cancel
              </button>
              <button
                onClick={() => setTargetMutation.mutate({ pathId: showTarget, date: targetDate || undefined })}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
              >
                Set Target
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
