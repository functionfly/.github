import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { featureFlagsApi, type FeatureFlag } from '@/api/feature_flags';
import { Flag, Plus, Search } from 'lucide-react';

const flagTypeColors: Record<string, string> = {
  boolean: 'bg-blue-500/20 text-blue-400',
  multivariate: 'bg-purple-500/20 text-purple-400',
  percentage: 'bg-cyan-500/20 text-cyan-400',
};

export function FeatureFlagsPage() {
  const queryClient = useQueryClient();
  const [showCreate, setShowCreate] = useState(false);
  const [editing, setEditing] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const [form, setForm] = useState({ key: '', name: '', flag_type: 'boolean', description: '' });
  const [editRollout, setEditRollout] = useState(0);

  const { data, isLoading } = useQuery({
    queryKey: ['feature-flags'],
    queryFn: () => featureFlagsApi.list(),
  });

  const createMutation = useMutation({
    mutationFn: (data: Partial<FeatureFlag>) => featureFlagsApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['feature-flags'] });
      setShowCreate(false);
      setForm({ key: '', name: '', flag_type: 'boolean', description: '' });
    },
  });

  const toggleMutation = useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) => featureFlagsApi.update(id, { is_enabled: enabled }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['feature-flags'] }),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<FeatureFlag> }) => featureFlagsApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['feature-flags'] });
      setEditing(null);
    },
  });

  const flags = (data?.data?.flags || []).filter((f) =>
    !search || f.name.toLowerCase().includes(search.toLowerCase()) || f.key.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Flag className="h-6 w-6 text-cyan-400" />
          <h1 className="text-2xl font-bold">Feature Flags</h1>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          <Plus className="h-4 w-4" />
          Create Flag
        </button>
      </div>

      <div className="flex items-center gap-3">
        <Search className="h-4 w-4 text-gray-400" />
        <input
          type="text"
          placeholder="Search flags..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-64 rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
        />
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : flags.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
          <Flag className="mb-4 h-12 w-12 text-gray-600" />
          <p className="text-gray-400">{search ? 'No flags match your search' : 'No feature flags yet'}</p>
        </div>
      ) : (
        <div className="space-y-3">
          {flags.map((flag) => (
            <div key={flag.id} className="rounded-xl border border-gray-800 bg-gray-900 p-4">
              <div className="flex items-center justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-3">
                    <h3 className="font-medium text-gray-100">{flag.name}</h3>
                    <span className={`rounded-full px-2 py-0.5 text-xs ${flagTypeColors[flag.flag_type] || 'bg-gray-500/20 text-gray-400'}`}>
                      {flag.flag_type}
                    </span>
                    <code className="rounded bg-gray-800 px-2 py-0.5 text-xs text-gray-400">{flag.key}</code>
                  </div>
                  {flag.description && <p className="mt-1 text-sm text-gray-400">{flag.description}</p>}
                  <div className="mt-2 flex items-center gap-4 text-xs text-gray-500">
                    <span>Rollout: {flag.rollout_pct}%</span>
                    {Object.keys(flag.variants).length > 0 && (
                      <span>Variants: {Object.entries(flag.variants).map(([k, v]) => `${k}(${v}%)`).join(', ')}</span>
                    )}
                    <span>{new Date(flag.created_at).toLocaleDateString()}</span>
                  </div>
                </div>
                <div className="flex items-center gap-3">
                  <button
                    onClick={() => {
                      setEditing(flag.id === editing ? null : flag.id);
                      setEditRollout(flag.rollout_pct);
                    }}
                    className="rounded-lg px-3 py-1.5 text-xs text-gray-400 hover:bg-gray-800 hover:text-gray-200"
                  >
                    Edit
                  </button>
                  <button
                    onClick={() => toggleMutation.mutate({ id: flag.id, enabled: !flag.is_enabled })}
                    className={`relative h-6 w-11 rounded-full transition-colors ${flag.is_enabled ? 'bg-green-600' : 'bg-gray-700'}`}
                  >
                    <span className={`absolute top-0.5 h-5 w-5 rounded-full bg-white transition-transform ${flag.is_enabled ? 'left-[22px]' : 'left-0.5'}`} />
                  </button>
                </div>
              </div>

              {editing === flag.id && (
                <div className="mt-4 border-t border-gray-700 pt-4 space-y-3">
                  <div>
                    <label className="mb-1 block text-xs text-gray-400">Rollout Percentage</label>
                    <div className="flex items-center gap-3">
                      <input
                        type="range"
                        min={0}
                        max={100}
                        value={editRollout}
                        onChange={(e) => setEditRollout(Number(e.target.value))}
                        className="flex-1"
                      />
                      <span className="w-12 text-right text-sm text-gray-200">{editRollout}%</span>
                    </div>
                  </div>
                  <div className="flex justify-end gap-2">
                    <button
                      onClick={() => setEditing(null)}
                      className="rounded-lg px-3 py-1.5 text-xs text-gray-400 hover:text-gray-200"
                    >
                      Cancel
                    </button>
                    <button
                      onClick={() => updateMutation.mutate({ id: flag.id, data: { rollout_pct: editRollout } })}
                      className="rounded-lg bg-blue-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-blue-700"
                    >
                      Save
                    </button>
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Create Feature Flag</h2>
            <input
              type="text"
              placeholder="Key (e.g. new_dashboard)"
              value={form.key}
              onChange={(e) => setForm({ ...form, key: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              autoFocus
            />
            <input
              type="text"
              placeholder="Display name"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
            />
            <select
              value={form.flag_type}
              onChange={(e) => setForm({ ...form, flag_type: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
            >
              <option value="boolean">Boolean</option>
              <option value="multivariate">Multivariate</option>
              <option value="percentage">Percentage</option>
            </select>
            <textarea
              placeholder="Description (optional)"
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
              className="mb-4 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              rows={3}
            />
            <div className="flex justify-end gap-3">
              <button onClick={() => setShowCreate(false)} className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200">Cancel</button>
              <button
                onClick={() => createMutation.mutate({
                  key: form.key,
                  name: form.name,
                  flag_type: form.flag_type,
                  description: form.description || undefined,
                  is_enabled: false,
                  rollout_pct: 0,
                  variants: {},
                  target_audience: {},
                })}
                disabled={!form.key.trim() || !form.name.trim()}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                Create
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
