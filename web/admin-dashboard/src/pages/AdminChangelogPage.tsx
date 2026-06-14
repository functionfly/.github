/**
 * Admin Changelog Page
 * Manage changelog entries with full CRUD operations and AI generation
 */

import { adminApiClient } from '@/lib/api/adminClient';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Calendar,
  CheckCircle2,
  Clock,
  Edit2,
  FileText,
  GitCommit,
  GitPullRequest,
  Loader2,
  Plus,
  RefreshCw,
  Rocket,
  Sparkles,
  Trash2,
  X,
  Zap,
} from 'lucide-react';
import { useEffect, useState } from 'react';

interface ChangelogEntry {
  id: string;
  version: string;
  date: string;
  type: 'major' | 'minor' | 'patch';
  title: string;
  description: string;
  changes: ChangelogChange[];
  release_url?: string;
  github_id?: string;
  is_published: boolean;
  created_at: string;
  updated_at: string;
}

interface ChangelogChange {
  id: string;
  entry_id: string;
  category: string;
  icon: string;
  items: string[];
  created_at: string;
  updated_at: string;
}

const CHANGE_TYPE_COLORS: Record<string, string> = {
  major: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
  minor: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',
  patch: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
};

const CHANGE_TYPE_ICONS: Record<string, typeof Rocket> = {
  major: Rocket,
  minor: Zap,
  patch: CheckCircle2,
};

const CATEGORY_OPTIONS = [
  { value: 'added', label: 'Added' },
  { value: 'changed', label: 'Changed' },
  { value: 'deprecated', label: 'Deprecated' },
  { value: 'removed', label: 'Removed' },
  { value: 'fixed', label: 'Fixed' },
  { value: 'security', label: 'Security' },
  { value: 'performance', label: 'Performance' },
  { value: 'documentation', label: 'Documentation' },
  { value: 'breaking', label: 'Breaking' },
];

export default function AdminChangelogPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
          <FileText className="w-7 h-7" />
          Changelog
        </h1>
        <p className="mt-2 text-gray-600 dark:text-gray-400">
          Manage changelog entries, release notes, and version history.
        </p>
      </div>

      <ChangelogActions />
      <ChangelogEntriesList />
    </div>
  );
}

function ChangelogActions() {
  const queryClient = useQueryClient();
  const [showGenerateModal, setShowGenerateModal] = useState(false);
  const [syncStatus, setSyncStatus] = useState<'idle' | 'syncing' | 'success' | 'error'>('idle');

  const syncMutation = useMutation({
    mutationFn: async () => {
      const res = await adminApiClient.post<{ message: string }>('/content/sync/github-releases', {});
      return res;
    },
    onSuccess: () => {
      setSyncStatus('success');
      queryClient.invalidateQueries({ queryKey: ['admin-changelog-entries'] });
      setTimeout(() => setSyncStatus('idle'), 3000);
    },
    onError: () => {
      setSyncStatus('error');
      setTimeout(() => setSyncStatus('idle'), 3000);
    },
  });

  return (
    <div className="flex items-center gap-3">
      <button
        type="button"
        onClick={() => setShowGenerateModal(true)}
        className="flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-purple-600 to-indigo-600 text-white rounded-lg hover:from-purple-700 hover:to-indigo-700 shadow-sm"
      >
        <Sparkles className="w-4 h-4" />
        Generate with AI
      </button>
      <button
        type="button"
        onClick={() => syncMutation.mutate()}
        disabled={syncStatus === 'syncing'}
        className="flex items-center gap-2 px-4 py-2 bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 border border-gray-300 dark:border-gray-600"
      >
        {syncStatus === 'syncing' ? (
          <Loader2 className="w-4 h-4 animate-spin" />
        ) : syncStatus === 'success' ? (
          <CheckCircle2 className="w-4 h-4 text-green-600" />
        ) : syncStatus === 'error' ? (
          <X className="w-4 h-4 text-red-600" />
        ) : (
          <RefreshCw className="w-4 h-4" />
        )}
        {syncStatus === 'syncing' ? 'Syncing…' : syncStatus === 'success' ? 'Synced!' : syncStatus === 'error' ? 'Failed' : 'Sync from GitHub'}
      </button>

      {showGenerateModal && (
        <GenerateChangelogModal onClose={() => setShowGenerateModal(false)} />
      )}
    </div>
  );
}

function GenerateChangelogModal({ onClose }: { onClose: () => void }) {
  const queryClient = useQueryClient();
  const [version, setVersion] = useState('');
  const [type, setType] = useState<'major' | 'minor' | 'patch'>('minor');
  const [topic, setTopic] = useState('');
  const [generated, setGenerated] = useState<{ title: string; description: string } | null>(null);
  const [isGenerating, setIsGenerating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const generateMutation = useMutation({
    mutationFn: async () => {
      setIsGenerating(true);
      setError(null);
      const res = await adminApiClient.post<{ title: string; description: string }>(
        '/content/generate/changelog',
        { version, type, topic }
      );
      return res;
    },
    onSuccess: (data) => {
      setGenerated(data);
      setIsGenerating(false);
    },
    onError: (err) => {
      setError(err.message || 'Failed to generate changelog content');
      setIsGenerating(false);
    },
  });

  const createAndFillMutation = useMutation({
    mutationFn: async () => {
      if (!generated) return;
      const entry = {
        version,
        date: new Date().toISOString().split('T')[0],
        type,
        title: generated.title,
        description: generated.description,
        is_published: false,
        changes: [],
      };
      const res = await adminApiClient.post<ChangelogEntry>('/content/changelog', entry);
      return res;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-changelog-entries'] });
      onClose();
    },
  });

  const handleGenerate = () => {
    generateMutation.mutate();
  };

  return (
    <div
      className="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4"
      onClick={onClose}
    >
      <div
        className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg w-full max-w-lg"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="px-5 py-4 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Sparkles className="w-5 h-5 text-purple-600" />
            <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
              Generate Changelog with AI
            </h2>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="p-5 space-y-4">
          <p className="text-sm text-gray-600 dark:text-gray-400">
            Use AI to generate a changelog entry title and description. The content is based on the free OpenRouter model.
          </p>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Version
              </label>
              <input
                type="text"
                value={version}
                onChange={(e) => setVersion(e.target.value)}
                placeholder="e.g. 2.0.0"
                className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 bg-white dark:bg-gray-800 focus:border-purple-500 focus:ring-1 focus:ring-purple-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Release Type
              </label>
              <select
                value={type}
                onChange={(e) => setType(e.target.value as 'major' | 'minor' | 'patch')}
                className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 bg-white dark:bg-gray-800 focus:border-purple-500 focus:ring-1 focus:ring-purple-500"
              >
                <option value="major">Major</option>
                <option value="minor">Minor</option>
                <option value="patch">Patch</option>
              </select>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Topic / Focus (optional)
            </label>
            <input
              type="text"
              value={topic}
              onChange={(e) => setTopic(e.target.value)}
              placeholder="e.g. New dashboard features, Performance improvements"
              className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 bg-white dark:bg-gray-800 focus:border-purple-500 focus:ring-1 focus:ring-purple-500"
            />
          </div>

          {error && (
            <div className="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-red-700 dark:text-red-300 text-sm">
              {error}
            </div>
          )}

          {generated && (
            <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg space-y-3">
              <div>
                <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 uppercase mb-1">
                  Generated Title
                </label>
                <p className="text-gray-900 dark:text-gray-100 font-medium">{generated.title}</p>
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 uppercase mb-1">
                  Generated Description
                </label>
                <p className="text-gray-700 dark:text-gray-300 text-sm">{generated.description}</p>
              </div>
            </div>
          )}
        </div>

        <div className="px-5 py-4 border-t border-gray-200 dark:border-gray-700 flex items-center justify-between">
          <button
            type="button"
            onClick={handleGenerate}
            disabled={isGenerating}
            className="flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-purple-600 to-indigo-600 text-white rounded-lg hover:from-purple-700 hover:to-indigo-700 disabled:opacity-50"
          >
            {isGenerating ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                Generating…
              </>
            ) : (
              <>
                <Sparkles className="w-4 h-4" />
                Generate
              </>
            )}
          </button>

          {generated && (
            <button
              type="button"
              onClick={() => createAndFillMutation.mutate()}
              disabled={createAndFillMutation.isPending}
              className="flex items-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50"
            >
              {createAndFillMutation.isPending ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : (
                <Plus className="w-4 h-4" />
              )}
              Create Entry
            </button>
          )}
        </div>
      </div>
    </div>
  );
}

function ChangelogEntriesList() {
  const queryClient = useQueryClient();
  const [showForm, setShowForm] = useState(false);
  const [editingEntry, setEditingEntry] = useState<ChangelogEntry | null>(null);
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null);
  const [filterPublished, setFilterPublished] = useState<boolean | null>(null);

  const {
    data: listData,
    isLoading,
    error: queryError,
  } = useQuery({
    queryKey: ['admin-changelog-entries', filterPublished],
    queryFn: async () => {
      const params = new URLSearchParams();
      params.set('limit', '50');
      if (filterPublished !== null) {
        params.set('published_only', filterPublished ? 'true' : 'false');
      }
      const res = await adminApiClient.get<{
        data: { entries: ChangelogEntry[]; limit: number; offset: number };
      }>(`/content/changelog?${params}`);
      return res;
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => adminApiClient.delete(`/content/changelog/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-changelog-entries'] });
      setDeleteConfirm(null);
    },
  });

  const entries = listData?.data?.entries ?? [];

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="w-6 h-6 animate-spin text-gray-400" />
        <span className="ml-2 text-gray-500">Loading changelog entries…</span>
      </div>
    );
  }

  if (queryError) {
    return (
      <div className="space-y-6">
        <div className="flex justify-between items-center">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Changelog entries</h2>
          <button
            type="button"
            onClick={() => {
              setEditingEntry(null);
              setShowForm(true);
            }}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            <Plus className="w-4 h-4" />
            New entry
          </button>
        </div>
        <div className="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-lg p-4 text-amber-800 dark:text-amber-200">
          <p className="font-medium">Unable to load changelog entries</p>
          <p className="text-sm mt-1">Please try refreshing the page or check your connection.</p>
        </div>
        {showForm && (
          <ChangelogEntryForm
            entry={editingEntry ?? undefined}
            onClose={() => {
              setShowForm(false);
              setEditingEntry(null);
            }}
            onSaved={() => {
              queryClient.invalidateQueries({ queryKey: ['admin-changelog-entries'] });
              setShowForm(false);
              setEditingEntry(null);
            }}
          />
        )}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div className="flex items-center gap-4">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Changelog entries</h2>
          <div className="flex items-center gap-2">
            <select
              value={filterPublished === null ? 'all' : filterPublished ? 'published' : 'drafts'}
              onChange={(e) => {
                const val = e.target.value;
                setFilterPublished(val === 'all' ? null : val === 'published');
              }}
              className="rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-1.5 text-sm text-gray-900 dark:text-gray-100 bg-white dark:bg-gray-800 focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
            >
              <option value="all">All entries</option>
              <option value="published">Published</option>
              <option value="drafts">Drafts</option>
            </select>
          </div>
        </div>
        <button
          type="button"
          onClick={() => {
            setEditingEntry(null);
            setShowForm(true);
          }}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          <Plus className="w-4 h-4" />
          New entry
        </button>
      </div>

      {showForm && (
        <ChangelogEntryForm
          entry={editingEntry ?? undefined}
          onClose={() => {
            setShowForm(false);
            setEditingEntry(null);
          }}
          onSaved={() => {
            queryClient.invalidateQueries({ queryKey: ['admin-changelog-entries'] });
            setShowForm(false);
            setEditingEntry(null);
          }}
        />
      )}

      <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
        {entries.length === 0 ? (
          <div className="px-6 py-12 text-center text-gray-500 dark:text-gray-400">
            <FileText className="w-12 h-12 mx-auto mb-4 text-gray-300 dark:text-gray-600" />
            <p className="font-medium text-gray-900 dark:text-gray-100">No changelog entries yet</p>
            <p className="text-sm mt-1">Create your first entry with "New entry" or use "Generate with AI".</p>
          </div>
        ) : (
          <div className="divide-y divide-gray-100 dark:divide-gray-700">
            {entries.map((entry) => (
              <ChangelogEntryRow
                key={entry.id}
                entry={entry}
                onEdit={() => {
                  setEditingEntry(entry);
                  setShowForm(true);
                }}
                onDelete={() => setDeleteConfirm(entry.id)}
                deleteConfirm={deleteConfirm}
                onConfirmDelete={() => deleteMutation.mutate(entry.id)}
                onCancelDelete={() => setDeleteConfirm(null)}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function ChangelogEntryRow({
  entry,
  onEdit,
  onDelete,
  deleteConfirm,
  onConfirmDelete,
  onCancelDelete,
}: {
  entry: ChangelogEntry;
  onEdit: () => void;
  onDelete: () => void;
  deleteConfirm: string | null;
  onConfirmDelete: () => void;
  onCancelDelete: () => void;
}) {
  const TypeIcon = CHANGE_TYPE_ICONS[entry.type] ?? FileText;
  const typeColor = CHANGE_TYPE_COLORS[entry.type] ?? 'bg-gray-100 text-gray-700';

  return (
    <div className="px-6 py-4 hover:bg-gray-50/60 dark:hover:bg-gray-800/40">
      <div className="flex items-start gap-4">
        <div className={`shrink-0 w-10 h-10 rounded-lg flex items-center justify-center ${typeColor}`}>
          <TypeIcon className="w-5 h-5" />
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <h3 className="text-base font-semibold text-gray-900 dark:text-gray-100">{entry.title}</h3>
            <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${typeColor}`}>
              {entry.type}
            </span>
            <span className="text-xs text-gray-500 dark:text-gray-400 flex items-center gap-1">
              <GitCommit className="w-3 h-3" />
              {entry.version}
            </span>
            {!entry.is_published && (
              <span className="text-xs font-medium px-2 py-0.5 rounded-full bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400">
                Draft
              </span>
            )}
          </div>
          <p className="mt-1 text-sm text-gray-600 dark:text-gray-400 line-clamp-2">{entry.description}</p>
          <div className="mt-2 flex items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
            <span className="flex items-center gap-1">
              <Calendar className="w-3 h-3" />
              {new Date(entry.date).toLocaleDateString()}
            </span>
            <span className="flex items-center gap-1">
              <Clock className="w-3 h-3" />
              Updated {new Date(entry.updated_at).toLocaleDateString()}
            </span>
            {entry.changes.length > 0 && (
              <span className="flex items-center gap-1">
                <GitPullRequest className="w-3 h-3" />
                {entry.changes.length} change{entry.changes.length !== 1 ? 's' : ''}
              </span>
            )}
          </div>
        </div>
        <div className="shrink-0 flex items-center gap-2">
          <button
            type="button"
            onClick={onEdit}
            className="p-2 text-gray-600 hover:text-blue-600 dark:text-gray-400 dark:hover:text-blue-400 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700"
            title="Edit entry"
          >
            <Edit2 className="w-4 h-4" />
          </button>
          {deleteConfirm === entry.id ? (
            <span className="flex items-center gap-1">
              <button
                type="button"
                onClick={onConfirmDelete}
                className="px-3 py-1.5 text-sm font-medium text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300"
              >
                Confirm
              </button>
              <button
                type="button"
                onClick={onCancelDelete}
                className="px-3 py-1.5 text-sm text-gray-600 dark:text-gray-400"
              >
                Cancel
              </button>
            </span>
          ) : (
            <button
              type="button"
              onClick={onDelete}
              className="p-2 text-gray-600 hover:text-red-600 dark:text-gray-400 dark:hover:text-red-400 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700"
              title="Delete entry"
            >
              <Trash2 className="w-4 h-4" />
            </button>
          )}
        </div>
      </div>
    </div>
  );
}

function ChangelogEntryForm({
  entry,
  onClose,
  onSaved,
}: {
  entry?: ChangelogEntry;
  onClose: () => void;
  onSaved: () => void;
}) {
  const [version, setVersion] = useState(entry?.version ?? '');
  const [date, setDate] = useState(entry?.date ?? new Date().toISOString().split('T')[0]);
  const [type, setType] = useState<'major' | 'minor' | 'patch'>(entry?.type ?? 'minor');
  const [title, setTitle] = useState(entry?.title ?? '');
  const [description, setDescription] = useState(entry?.description ?? '');
  const [isPublished, setIsPublished] = useState(entry?.is_published ?? false);
  const [releaseUrl, setReleaseUrl] = useState(entry?.release_url ?? '');
  const [changes, setChanges] = useState<{ category: string; items: string[] }[]>(
    entry?.changes.map((c) => ({ category: c.category, items: c.items })) ?? []
  );
  const [showChangeForm, setShowChangeForm] = useState(false);
  const [newChangeCategory, setNewChangeCategory] = useState('added');
  const [newChangeItems, setNewChangeItems] = useState('');

  useEffect(() => {
    if (entry) {
      setVersion(entry.version);
      setDate(entry.date);
      setType(entry.type);
      setTitle(entry.title);
      setDescription(entry.description);
      setIsPublished(entry.is_published);
      setReleaseUrl(entry.release_url ?? '');
      setChanges(entry.changes.map((c) => ({ category: c.category, items: c.items })));
    }
  }, [entry]);

  const createMutation = useMutation({
    mutationFn: (payload: Record<string, unknown>) => adminApiClient.post<ChangelogEntry>('/content/changelog', payload),
    onSuccess: () => onSaved(),
  });

  const updateMutation = useMutation({
    mutationFn: (payload: Record<string, unknown>) =>
      adminApiClient.patch<ChangelogEntry>(`/content/changelog/${entry!.id}`, payload),
    onSuccess: () => onSaved(),
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const payload = {
      version,
      date,
      type,
      title,
      description,
      is_published: isPublished,
      release_url: releaseUrl || undefined,
      changes: changes.map((c) => ({ category: c.category, items: c.items })),
    };
    if (entry) {
      updateMutation.mutate(payload);
    } else {
      createMutation.mutate(payload);
    }
  };

  const addChange = () => {
    if (newChangeItems.trim()) {
      setChanges([
        ...changes,
        {
          category: newChangeCategory,
          items: newChangeItems.split('\n').filter((s) => s.trim()),
        },
      ]);
      setNewChangeItems('');
      setShowChangeForm(false);
    }
  };

  const removeChange = (index: number) => {
    setChanges(changes.filter((_, i) => i !== index));
  };

  const saving = createMutation.isPending || updateMutation.isPending;

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
      <div className="flex justify-between items-center mb-4">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
          {entry ? 'Edit entry' : 'New changelog entry'}
        </h3>
        <button
          type="button"
          onClick={onClose}
          className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
        >
          <X className="w-5 h-5" />
        </button>
      </div>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="grid grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Version *
            </label>
            <input
              type="text"
              value={version}
              onChange={(e) => setVersion(e.target.value)}
              placeholder="e.g. 2.0.0"
              required
              className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Date *</label>
            <input
              type="date"
              value={date}
              onChange={(e) => setDate(e.target.value)}
              required
              className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Type *</label>
            <select
              value={type}
              onChange={(e) => setType(e.target.value as 'major' | 'minor' | 'patch')}
              className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
            >
              <option value="major">Major</option>
              <option value="minor">Minor</option>
              <option value="patch">Patch</option>
            </select>
          </div>
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Title *</label>
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="e.g. Release v2.0 with new dashboard"
            required
            className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Description *
          </label>
          <textarea
            rows={3}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Brief description of this release"
            required
            className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Release URL
          </label>
          <input
            type="url"
            value={releaseUrl}
            onChange={(e) => setReleaseUrl(e.target.value)}
            placeholder="https://github.com/..."
            className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
          />
        </div>

        <div>
          <div className="flex items-center justify-between mb-2">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Changes</label>
            <button
              type="button"
              onClick={() => setShowChangeForm(!showChangeForm)}
              className="text-sm text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300"
            >
              {showChangeForm ? 'Cancel' : '+ Add change'}
            </button>
          </div>
          {showChangeForm && (
            <div className="mb-3 p-4 bg-gray-50 dark:bg-gray-900 rounded-lg space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Category
                </label>
                <select
                  value={newChangeCategory}
                  onChange={(e) => setNewChangeCategory(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 bg-white dark:bg-gray-800 focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
                >
                  {CATEGORY_OPTIONS.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Items (one per line)
                </label>
                <textarea
                  rows={4}
                  value={newChangeItems}
                  onChange={(e) => setNewChangeItems(e.target.value)}
                  placeholder="New feature added&#10;Bug fix applied"
                  className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-800"
                />
              </div>
              <button
                type="button"
                onClick={addChange}
                className="px-3 py-1.5 bg-blue-600 text-white text-sm rounded-lg hover:bg-blue-700"
              >
                Add
              </button>
            </div>
          )}
          {changes.length > 0 ? (
            <div className="space-y-3">
              {changes.map((change, idx) => (
                <div
                  key={idx}
                  className="flex items-start gap-3 p-3 bg-gray-50 dark:bg-gray-900 rounded-lg"
                >
                  <div className="flex-1">
                    <span className="text-xs font-medium text-gray-600 dark:text-gray-400 uppercase">
                      {change.category}
                    </span>
                    <ul className="mt-1 space-y-1">
                      {change.items.map((item, i) => (
                        <li key={i} className="text-sm text-gray-700 dark:text-gray-300">
                          {item}
                        </li>
                      ))}
                    </ul>
                  </div>
                  <button
                    type="button"
                    onClick={() => removeChange(idx)}
                    className="text-gray-400 hover:text-red-500"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-gray-500 dark:text-gray-400">No changes added yet.</p>
          )}
        </div>

        <div className="flex items-center gap-3">
          <input
            type="checkbox"
            id="is_published"
            checked={isPublished}
            onChange={(e) => setIsPublished(e.target.checked)}
            className="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
          />
          <label htmlFor="is_published" className="text-sm font-medium text-gray-700 dark:text-gray-300">
            Published
          </label>
        </div>
        <div className="flex gap-3 pt-2">
          <button
            type="submit"
            disabled={saving}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
          >
            {saving ? 'Saving…' : entry ? 'Update entry' : 'Create entry'}
          </button>
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  );
}
