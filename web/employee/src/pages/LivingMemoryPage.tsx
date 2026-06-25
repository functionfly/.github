import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { livingMemoryApi, type LivingMemoryEntry } from '@/api/living_memory';
import { Brain, Search, Plus, Clock, Users, Tag, FileText } from 'lucide-react';

const typeColors: Record<string, string> = {
  Meeting: 'bg-blue-500/20 text-blue-400',
  Decision: 'bg-yellow-500/20 text-yellow-400',
  Design: 'bg-purple-500/20 text-purple-400',
  Lesson: 'bg-green-500/20 text-green-400',
  Discovery: 'bg-pink-500/20 text-pink-400',
  Note: 'bg-gray-500/20 text-gray-400',
};

const importanceColors: Record<string, string> = {
  critical: 'bg-red-500/20 text-red-400',
  high: 'bg-orange-500/20 text-orange-400',
  normal: 'bg-blue-500/20 text-blue-400',
  low: 'bg-gray-500/20 text-gray-400',
};

const memoryTypes = ['', 'Meeting', 'Decision', 'Design', 'Lesson', 'Discovery', 'Note'];

export function LivingMemoryPage() {
  const queryClient = useQueryClient();
  const [query, setQuery] = useState('');
  const [typeFilter, setTypeFilter] = useState('');
  const [showCreate, setShowCreate] = useState(false);
  const [form, setForm] = useState({
    title: '',
    body: '',
    memory_type: 'Note',
    importance: 'normal',
    tags: '',
    participants: '',
  });

  const { data: searchData, isLoading: searchLoading } = useQuery({
    queryKey: ['memory', 'search', query, typeFilter],
    queryFn: () => livingMemoryApi.search(query, typeFilter ? { memory_type: typeFilter } : undefined),
    enabled: query.length > 0,
  });

  const { data: listData, isLoading: listLoading } = useQuery({
    queryKey: ['memory', 'list', typeFilter],
    queryFn: () => livingMemoryApi.list(typeFilter ? { memory_type: typeFilter } : undefined),
    enabled: query.length === 0,
  });

  const createMutation = useMutation({
    mutationFn: (data: Partial<LivingMemoryEntry>) => livingMemoryApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['memory'] });
      setShowCreate(false);
      setForm({ title: '', body: '', memory_type: 'Note', importance: 'normal', tags: '', participants: '' });
    },
  });

  const entries = query.length > 0 ? (searchData?.data?.results || []) : (listData?.data?.entries || []);
  const isLoading = query.length > 0 ? searchLoading : listLoading;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Brain className="h-6 w-6 text-purple-400" />
          <h1 className="text-2xl font-bold">Living Memory</h1>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          <Plus className="h-4 w-4" />
          New Entry
        </button>
      </div>

      <div className="flex flex-wrap items-center gap-3">
        <div className="relative flex-1 sm:max-w-md">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-500" />
          <input
            type="text"
            placeholder="Search memories..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="w-full rounded-lg border border-gray-700 bg-gray-800 py-2 pl-9 pr-3 text-sm text-gray-100 placeholder-gray-500"
          />
        </div>
        <select
          value={typeFilter}
          onChange={(e) => setTypeFilter(e.target.value)}
          className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
        >
          <option value="">All Types</option>
          {memoryTypes.filter(Boolean).map((t) => (
            <option key={t} value={t}>{t}</option>
          ))}
        </select>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : entries.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
          <Brain className="mb-4 h-12 w-12 text-gray-600" />
          <p className="text-gray-400">{query ? 'No results found' : 'No memory entries yet'}</p>
        </div>
      ) : (
        <div className="space-y-3">
          {entries.map((entry) => (
            <div key={entry.id} className="rounded-xl border border-gray-800 bg-gray-900 p-5">
              <div className="mb-2 flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <h3 className="font-semibold text-gray-100">{entry.title}</h3>
                    <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${typeColors[entry.memory_type] || 'bg-gray-500/20 text-gray-400'}`}>
                      {entry.memory_type}
                    </span>
                    <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${importanceColors[entry.importance] || ''}`}>
                      {entry.importance}
                    </span>
                  </div>
                  <p className="mt-2 text-sm text-gray-400 line-clamp-2">{entry.body}</p>
                </div>
              </div>
              <div className="flex flex-wrap items-center gap-4 text-xs text-gray-500">
                <span className="flex items-center gap-1">
                  <Clock className="h-3 w-3" />
                  {new Date(entry.created_at).toLocaleDateString()}
                </span>
                {entry.participants.length > 0 && (
                  <span className="flex items-center gap-1">
                    <Users className="h-3 w-3" />
                    {entry.participants.length} participants
                  </span>
                )}
                <span className="flex items-center gap-1">
                  <FileText className="h-3 w-3" />
                  {entry.view_count} views
                </span>
                {entry.tags.length > 0 && (
                  <span className="flex items-center gap-1">
                    <Tag className="h-3 w-3" />
                    {entry.tags.join(', ')}
                  </span>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-lg rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">New Memory Entry</h2>
            <input
              type="text"
              placeholder="Title"
              value={form.title}
              onChange={(e) => setForm({ ...form, title: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              autoFocus
            />
            <textarea
              placeholder="Body"
              value={form.body}
              onChange={(e) => setForm({ ...form, body: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              rows={5}
            />
            <div className="mb-3 grid grid-cols-2 gap-3">
              <select
                value={form.memory_type}
                onChange={(e) => setForm({ ...form, memory_type: e.target.value })}
                className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
              >
                {memoryTypes.filter(Boolean).map((t) => (
                  <option key={t} value={t}>{t}</option>
                ))}
              </select>
              <select
                value={form.importance}
                onChange={(e) => setForm({ ...form, importance: e.target.value })}
                className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
              >
                <option value="low">Low</option>
                <option value="normal">Normal</option>
                <option value="high">High</option>
                <option value="critical">Critical</option>
              </select>
            </div>
            <input
              type="text"
              placeholder="Tags (comma-separated)"
              value={form.tags}
              onChange={(e) => setForm({ ...form, tags: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
            />
            <input
              type="text"
              placeholder="Participants (comma-separated IDs)"
              value={form.participants}
              onChange={(e) => setForm({ ...form, participants: e.target.value })}
              className="mb-4 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
            />
            <div className="flex justify-end gap-3">
              <button
                onClick={() => { setShowCreate(false); setForm({ title: '', body: '', memory_type: 'Note', importance: 'normal', tags: '', participants: '' }); }}
                className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200"
              >
                Cancel
              </button>
              <button
                onClick={() => createMutation.mutate({
                  title: form.title,
                  body: form.body,
                  memory_type: form.memory_type,
                  importance: form.importance,
                  tags: form.tags.split(',').map((t) => t.trim()).filter(Boolean),
                  participants: form.participants.split(',').map((p) => p.trim()).filter(Boolean),
                })}
                disabled={!form.title.trim() || !form.body.trim()}
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
