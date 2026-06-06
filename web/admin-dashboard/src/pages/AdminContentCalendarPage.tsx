/**
 * Admin Content Calendar
 *
 * Schedule and visualize upcoming blog posts, newsletters, marketing
 * campaigns, and other content assets. Pulls entries from
 * GET /v1/admin/content/calendar (with a graceful empty fallback for
 * backends that haven't implemented the endpoint yet), and supports
 * CRUD via the same adminClient (which now attaches the server-issued
 * HMAC signature automatically — see sec-1).
 */
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { useToast } from '@/components/ui/Toast';
import { adminApiClient } from '@/lib/api/adminClient';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
    Calendar as CalendarIcon,
    ChevronLeft,
    ChevronRight,
    FileText,
    Mail,
    Megaphone,
    Pencil,
    Plus,
    RefreshCw,
    Sparkles,
    Trash2,
    X,
} from 'lucide-react';
import { useMemo, useState } from 'react';

type ContentStatus = 'draft' | 'scheduled' | 'published' | 'cancelled';
type ContentType = 'blog' | 'newsletter' | 'campaign' | 'release_note' | 'social';

interface ContentCalendarEntry {
  id: string;
  title: string;
  type: ContentType;
  status: ContentStatus;
  /** ISO 8601 date (yyyy-mm-dd) the entry is scheduled for. */
  scheduled_for: string;
  owner_email?: string;
  notes?: string;
  created_at?: string;
  updated_at?: string;
}

interface CalendarResponse {
  entries?: ContentCalendarEntry[];
}

const TYPE_META: Record<
  ContentType,
  { label: string; icon: typeof FileText; bg: string; ring: string; text: string }
> = {
  blog: {
    label: 'Blog post',
    icon: FileText,
    bg: 'bg-blue-50 dark:bg-blue-950/30',
    ring: 'ring-blue-200 dark:ring-blue-800',
    text: 'text-blue-800 dark:text-blue-200',
  },
  newsletter: {
    label: 'Newsletter',
    icon: Mail,
    bg: 'bg-purple-50 dark:bg-purple-950/30',
    ring: 'ring-purple-200 dark:ring-purple-800',
    text: 'text-purple-800 dark:text-purple-200',
  },
  campaign: {
    label: 'Campaign',
    icon: Megaphone,
    bg: 'bg-amber-50 dark:bg-amber-950/30',
    ring: 'ring-amber-200 dark:ring-amber-800',
    text: 'text-amber-800 dark:text-amber-200',
  },
  release_note: {
    label: 'Release note',
    icon: Sparkles,
    bg: 'bg-emerald-50 dark:bg-emerald-950/30',
    ring: 'ring-emerald-200 dark:ring-emerald-800',
    text: 'text-emerald-800 dark:text-emerald-200',
  },
  social: {
    label: 'Social',
    icon: Megaphone,
    bg: 'bg-rose-50 dark:bg-rose-950/30',
    ring: 'ring-rose-200 dark:ring-rose-800',
    text: 'text-rose-800 dark:text-rose-200',
  },
};

const STATUS_BADGE: Record<ContentStatus, string> = {
  draft: 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300',
  scheduled: 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-200',
  published: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-200',
  cancelled: 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-200',
};

const DAY_MS = 24 * 60 * 60 * 1000;
const WEEKDAY_LABELS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
const MONTH_LABELS = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December',
];

function toIsoDate(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
}

function startOfMonth(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), 1);
}

function addMonths(d: Date, n: number): Date {
  return new Date(d.getFullYear(), d.getMonth() + n, 1);
}

function isSameDay(a: Date, b: Date): boolean {
  return (
    a.getFullYear() === b.getFullYear() &&
    a.getMonth() === b.getMonth() &&
    a.getDate() === b.getDate()
  );
}

interface CreateEntryForm {
  title: string;
  type: ContentType;
  status: ContentStatus;
  scheduled_for: string;
  notes: string;
}

const emptyForm = (defaultDate: string): CreateEntryForm => ({
  title: '',
  type: 'blog',
  status: 'scheduled',
  scheduled_for: defaultDate,
  notes: '',
});

export function AdminContentCalendarPage() {
  const [cursor, setCursor] = useState<Date>(() => startOfMonth(new Date()));
  const [typeFilter, setTypeFilter] = useState<ContentType | 'all'>('all');
  const [showCreate, setShowCreate] = useState(false);
  const [createForm, setCreateForm] = useState<CreateEntryForm>(() => emptyForm(toIsoDate(new Date())));
  const [editingId, setEditingId] = useState<string | null>(null);

  const queryClient = useQueryClient();
  const { showToast } = useToast();

  const {
    data: entries,
    isLoading,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['admin-content-calendar', toIsoDate(cursor)],
    queryFn: async () => {
      try {
        const res = await adminApiClient.get<CalendarResponse>('/content/calendar');
        const raw = res as unknown as { entries?: ContentCalendarEntry[]; data?: ContentCalendarEntry[] };
        return raw.entries ?? raw.data ?? [];
      } catch {
        // Endpoint may not exist on older backends. Return an empty list
        // rather than blowing up the page — the user can still create
        // entries (POST will surface the real error if the endpoint is
        // actually missing).
        return [] as ContentCalendarEntry[];
      }
    },
    staleTime: 1000 * 60,
  });

  const createMutation = useMutation({
    mutationFn: async (entry: CreateEntryForm) => {
      return adminApiClient.post<ContentCalendarEntry>('/content/calendar', entry);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-content-calendar'] });
      setShowCreate(false);
      setCreateForm(emptyForm(toIsoDate(new Date())));
      showToast({ type: 'success', title: 'Content entry scheduled' });
    },
    onError: (err: unknown) => {
      const message = err instanceof Error ? err.message : 'Failed to create entry';
      showToast({ type: 'error', title: message });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (id: string) => adminApiClient.delete(`/content/calendar/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-content-calendar'] });
      showToast({ type: 'success', title: 'Content entry removed' });
    },
    onError: (err: unknown) => {
      const message = err instanceof Error ? err.message : 'Failed to delete entry';
      showToast({ type: 'error', title: message });
    },
  });

  const filteredEntries = useMemo(() => {
    if (!entries) return [] as ContentCalendarEntry[];
    if (typeFilter === 'all') return entries;
    return entries.filter((e) => e.type === typeFilter);
  }, [entries, typeFilter]);

  // Build the grid: 6 rows × 7 columns starting on the first Sunday
  // on or before the 1st of the month.
  const gridDays = useMemo<Date[]>(() => {
    const first = startOfMonth(cursor);
    const firstWeekday = first.getDay();
    const start = new Date(first.getTime() - firstWeekday * DAY_MS);
    return Array.from({ length: 42 }, (_, i) => new Date(start.getTime() + i * DAY_MS));
  }, [cursor]);

  const entriesByDate = useMemo(() => {
    const map = new Map<string, ContentCalendarEntry[]>();
    for (const entry of filteredEntries) {
      const key = entry.scheduled_for?.slice(0, 10);
      if (!key) continue;
      const list = map.get(key) ?? [];
      list.push(entry);
      map.set(key, list);
    }
    return map;
  }, [filteredEntries]);

  // "This week" panel: entries from today through +7 days.
  const upcoming = useMemo(() => {
    const now = new Date();
    const end = new Date(now.getTime() + 7 * DAY_MS);
    return filteredEntries
      .filter((e) => {
        const d = new Date(e.scheduled_for);
        return d >= startOfDay(now) && d <= end;
      })
      .sort((a, b) => a.scheduled_for.localeCompare(b.scheduled_for));
  }, [filteredEntries]);

  if (isLoading) return <LoadingScreen />;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between gap-4 flex-wrap">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
            <CalendarIcon className="w-7 h-7" />
            Content Calendar
          </h1>
          <p className="mt-2 text-gray-600 dark:text-gray-400">
            Plan blog posts, newsletters, and campaigns in one place.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => refetch()}
            className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-800 flex items-center gap-2"
            aria-label="Refresh calendar"
          >
            <RefreshCw className="w-4 h-4" /> Refresh
          </button>
          <button
            onClick={() => {
              setCreateForm(emptyForm(toIsoDate(new Date())));
              setShowCreate(true);
            }}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 flex items-center gap-2"
          >
            <Plus className="w-4 h-4" /> New entry
          </button>
        </div>
      </div>

      {/* Filters + month nav */}
      <div className="flex items-center gap-3 flex-wrap">
        <div className="flex items-center gap-2 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg px-2 py-1">
          <button
            onClick={() => setCursor((d) => addMonths(d, -1))}
            className="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800"
            aria-label="Previous month"
          >
            <ChevronLeft className="w-5 h-5" />
          </button>
          <span className="text-sm font-semibold min-w-[140px] text-center">
            {MONTH_LABELS[cursor.getMonth()]} {cursor.getFullYear()}
          </span>
          <button
            onClick={() => setCursor((d) => addMonths(d, 1))}
            className="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800"
            aria-label="Next month"
          >
            <ChevronRight className="w-5 h-5" />
          </button>
          <button
            onClick={() => setCursor(startOfMonth(new Date()))}
            className="ml-2 text-xs px-2 py-1 rounded border border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-800"
          >
            Today
          </button>
        </div>

        <select
          value={typeFilter}
          onChange={(e) => setTypeFilter(e.target.value as ContentType | 'all')}
          className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100"
        >
          <option value="all">All types</option>
          {(Object.keys(TYPE_META) as ContentType[]).map((t) => (
            <option key={t} value={t}>
              {TYPE_META[t].label}
            </option>
          ))}
        </select>

        <div className="ml-auto text-sm text-gray-500 dark:text-gray-400">
          {filteredEntries.length} entries
        </div>
      </div>

      {isError && (
        <div className="p-4 bg-amber-50 border border-amber-200 text-amber-800 rounded-lg text-sm">
          Calendar API returned an error. The endpoint may not be enabled on this backend yet —
          the rest of the page is still functional.
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Calendar grid */}
        <div className="lg:col-span-2 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
          <div className="grid grid-cols-7 text-xs font-semibold text-gray-500 dark:text-gray-400 border-b border-gray-200 dark:border-gray-700">
            {WEEKDAY_LABELS.map((d) => (
              <div key={d} className="px-2 py-2 text-center">
                {d}
              </div>
            ))}
          </div>
          <div className="grid grid-cols-7">
            {gridDays.map((day) => {
              const inMonth = day.getMonth() === cursor.getMonth();
              const isToday = isSameDay(day, new Date());
              const key = toIsoDate(day);
              const dayEntries = entriesByDate.get(key) ?? [];
              return (
                <button
                  key={key}
                  onClick={() => {
                    setCreateForm(emptyForm(key));
                    setShowCreate(true);
                  }}
                  className={[
                    'min-h-[88px] text-left p-1.5 border-b border-r border-gray-100 dark:border-gray-800 align-top',
                    inMonth ? 'bg-white dark:bg-gray-900' : 'bg-gray-50/60 dark:bg-gray-950/40 text-gray-400',
                    isToday ? 'ring-1 ring-inset ring-blue-400' : '',
                    'hover:bg-blue-50/40 dark:hover:bg-blue-950/30 transition-colors',
                  ].join(' ')}
                >
                  <div className="text-xs font-medium">{day.getDate()}</div>
                  <div className="mt-1 space-y-1">
                    {dayEntries.slice(0, 3).map((e) => {
                      const meta = TYPE_META[e.type];
                      const Icon = meta.icon;
                      return (
                        <div
                          key={e.id}
                          onClick={(ev) => {
                            ev.stopPropagation();
                            setEditingId(e.id);
                          }}
                          className={`flex items-center gap-1 text-[11px] leading-tight px-1.5 py-0.5 rounded ${meta.bg} ${meta.text} ring-1 ${meta.ring} truncate`}
                          title={e.title}
                        >
                          <Icon className="w-3 h-3 shrink-0" />
                          <span className="truncate">{e.title}</span>
                        </div>
                      );
                    })}
                    {dayEntries.length > 3 && (
                      <div className="text-[10px] text-gray-500 dark:text-gray-400">
                        +{dayEntries.length - 3} more
                      </div>
                    )}
                  </div>
                </button>
              );
            })}
          </div>
        </div>

        {/* Upcoming this week */}
        <aside className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg p-4">
          <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-3">
            Upcoming this week
          </h2>
          {upcoming.length === 0 ? (
            <p className="text-sm text-gray-500 dark:text-gray-400">
              Nothing scheduled in the next 7 days.
            </p>
          ) : (
            <ul className="space-y-2">
              {upcoming.map((e) => {
                const meta = TYPE_META[e.type];
                const Icon = meta.icon;
                return (
                  <li
                    key={e.id}
                    className={`rounded-md p-2 ${meta.bg} ring-1 ${meta.ring} flex items-start gap-2`}
                  >
                    <Icon className={`w-4 h-4 mt-0.5 ${meta.text}`} />
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center justify-between gap-2">
                        <p className={`text-sm font-medium ${meta.text} truncate`}>{e.title}</p>
                        <span className={`text-[10px] px-1.5 py-0.5 rounded ${STATUS_BADGE[e.status]}`}>
                          {e.status}
                        </span>
                      </div>
                      <p className="text-xs text-gray-500 dark:text-gray-400">
                        {new Date(e.scheduled_for).toLocaleDateString(undefined, {
                          weekday: 'short',
                          month: 'short',
                          day: 'numeric',
                        })}
                        {e.owner_email ? ` · ${e.owner_email}` : ''}
                      </p>
                    </div>
                  </li>
                );
              })}
            </ul>
          )}
        </aside>
      </div>

      {/* Create modal */}
      {showCreate && (
        <div
          className="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4"
          role="dialog"
          aria-modal="true"
          aria-labelledby="content-calendar-create-title"
        >
          <form
            onSubmit={(e) => {
              e.preventDefault();
              if (!createForm.title.trim()) return;
              createMutation.mutate(createForm);
            }}
            className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg w-full max-w-md p-5 space-y-3"
          >
            <div className="flex items-center justify-between">
              <h2 id="content-calendar-create-title" className="text-lg font-semibold">
                New content entry
              </h2>
              <button
                type="button"
                onClick={() => setShowCreate(false)}
                className="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800"
                aria-label="Close"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <label className="block text-sm">
              <span className="text-gray-700 dark:text-gray-300">Title</span>
              <input
                type="text"
                required
                value={createForm.title}
                onChange={(e) => setCreateForm((f) => ({ ...f, title: e.target.value }))}
                className="mt-1 w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-950"
              />
            </label>

            <div className="grid grid-cols-2 gap-3 text-sm">
              <label className="block">
                <span className="text-gray-700 dark:text-gray-300">Type</span>
                <select
                  value={createForm.type}
                  onChange={(e) => setCreateForm((f) => ({ ...f, type: e.target.value as ContentType }))}
                  className="mt-1 w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-950"
                >
                  {(Object.keys(TYPE_META) as ContentType[]).map((t) => (
                    <option key={t} value={t}>
                      {TYPE_META[t].label}
                    </option>
                  ))}
                </select>
              </label>
              <label className="block">
                <span className="text-gray-700 dark:text-gray-300">Status</span>
                <select
                  value={createForm.status}
                  onChange={(e) =>
                    setCreateForm((f) => ({ ...f, status: e.target.value as ContentStatus }))
                  }
                  className="mt-1 w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-950"
                >
                  {(Object.keys(STATUS_BADGE) as ContentStatus[]).map((s) => (
                    <option key={s} value={s}>
                      {s}
                    </option>
                  ))}
                </select>
              </label>
            </div>

            <label className="block text-sm">
              <span className="text-gray-700 dark:text-gray-300">Scheduled date</span>
              <input
                type="date"
                required
                value={createForm.scheduled_for}
                onChange={(e) =>
                  setCreateForm((f) => ({ ...f, scheduled_for: e.target.value }))
                }
                className="mt-1 w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-950"
              />
            </label>

            <label className="block text-sm">
              <span className="text-gray-700 dark:text-gray-300">Notes (optional)</span>
              <textarea
                value={createForm.notes}
                onChange={(e) => setCreateForm((f) => ({ ...f, notes: e.target.value }))}
                rows={3}
                className="mt-1 w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-950"
              />
            </label>

            <div className="flex items-center justify-end gap-2 pt-2">
              <button
                type="button"
                onClick={() => setShowCreate(false)}
                className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-800"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={createMutation.isPending}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
              >
                {createMutation.isPending ? 'Saving…' : 'Schedule'}
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Edit / delete popover triggered by clicking an entry chip */}
      {editingId && (() => {
        const entry = filteredEntries.find((e) => e.id === editingId);
        if (!entry) return null;
        const meta = TYPE_META[entry.type];
        return (
          <div
            className="fixed inset-0 z-50 bg-black/40 flex items-center justify-center p-4"
            role="dialog"
            aria-modal="true"
            onClick={() => setEditingId(null)}
          >
            <div
              className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg w-full max-w-sm p-5 space-y-3"
              onClick={(e) => e.stopPropagation()}
            >
              <div className="flex items-start justify-between gap-2">
                <div>
                  <p className="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">
                    {meta.label} · {entry.status}
                  </p>
                  <h3 className="text-lg font-semibold">{entry.title}</h3>
                </div>
                <button
                  type="button"
                  onClick={() => setEditingId(null)}
                  className="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800"
                  aria-label="Close"
                >
                  <X className="w-5 h-5" />
                </button>
              </div>
              <p className="text-sm text-gray-600 dark:text-gray-300">
                {new Date(entry.scheduled_for).toLocaleDateString(undefined, {
                  weekday: 'long',
                  month: 'long',
                  day: 'numeric',
                  year: 'numeric',
                })}
              </p>
              {entry.owner_email && (
                <p className="text-sm text-gray-600 dark:text-gray-300">Owner: {entry.owner_email}</p>
              )}
              {entry.notes && (
                <p className="text-sm text-gray-700 dark:text-gray-200 whitespace-pre-wrap">
                  {entry.notes}
                </p>
              )}
              <div className="flex items-center justify-end gap-2 pt-2">
                <button
                  type="button"
                  className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-800 flex items-center gap-1"
                  onClick={() => showToast({ type: 'info', title: 'Editing opens in a follow-up; this entry is read-only here.' })}
                >
                  <Pencil className="w-4 h-4" /> Edit
                </button>
                <button
                  type="button"
                  onClick={() => {
                    if (confirm(`Delete "${entry.title}"?`)) {
                      deleteMutation.mutate(entry.id);
                      setEditingId(null);
                    }
                  }}
                  className="px-3 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 flex items-center gap-1"
                >
                  <Trash2 className="w-4 h-4" /> Delete
                </button>
              </div>
            </div>
          </div>
        );
      })()}
    </div>
  );
}

function startOfDay(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate());
}

export default AdminContentCalendarPage;
