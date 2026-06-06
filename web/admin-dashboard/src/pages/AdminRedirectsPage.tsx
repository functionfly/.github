/**
 * Admin Redirects Page
 *
 * Manage HTTP redirect rules (from → to, status code, whether to preserve
 * the path/query). Hits the `/v1/admin/redirects` endpoints. Like the
 * content calendar, gracefully handles backends that haven't implemented
 * the endpoint yet by returning an empty list and surfacing the actual
 * error on create/update/delete mutations.
 */
import { useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Plus,
  Trash2,
  Pencil,
  X,
  RefreshCw,
  ArrowRight,
  Search,
} from 'lucide-react';
import { adminApiClient } from '@/lib/api/adminClient';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { useToast } from '@/components/ui/Toast';

type RedirectStatus = 301 | 302 | 307 | 308;

interface Redirect {
  id: string;
  from_path: string;
  to_url: string;
  status_code: RedirectStatus;
  preserve_query?: boolean;
  enabled?: boolean;
  notes?: string;
  created_at?: string;
  updated_at?: string;
}

interface RedirectsResponse {
  redirects?: Redirect[];
}

const STATUS_CODE_OPTIONS: RedirectStatus[] = [301, 302, 307, 308];

const STATUS_CODE_LABEL: Record<RedirectStatus, string> = {
  301: '301 — Permanent',
  302: '302 — Temporary',
  307: '307 — Temporary (preserve method)',
  308: '308 — Permanent (preserve method)',
};

interface RedirectForm {
  from_path: string;
  to_url: string;
  status_code: RedirectStatus;
  preserve_query: boolean;
  enabled: boolean;
  notes: string;
}

const emptyForm = (): RedirectForm => ({
  from_path: '/',
  to_url: 'https://',
  status_code: 301,
  preserve_query: true,
  enabled: true,
  notes: '',
});

function normalizePath(path: string): string {
  let p = path.trim();
  if (!p.startsWith('/')) p = `/${p}`;
  // Drop a trailing slash except for the bare "/" case.
  if (p.length > 1 && p.endsWith('/')) p = p.slice(0, -1);
  return p;
}

export function AdminRedirectsPage() {
  const [search, setSearch] = useState('');
  const [editing, setEditing] = useState<Redirect | null>(null);
  const [creating, setCreating] = useState(false);
  const [createForm, setCreateForm] = useState<RedirectForm>(emptyForm());
  const [editForm, setEditForm] = useState<RedirectForm>(emptyForm());

  const queryClient = useQueryClient();
  const { showToast } = useToast();

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['admin-redirects'],
    queryFn: async () => {
      try {
        const res = await adminApiClient.get<RedirectsResponse>('/redirects');
        const raw = res as unknown as { redirects?: Redirect[]; data?: Redirect[] };
        return raw.redirects ?? raw.data ?? [];
      } catch {
        return [] as Redirect[];
      }
    },
    staleTime: 1000 * 60,
  });

  const createMutation = useMutation({
    mutationFn: async (form: RedirectForm) =>
      adminApiClient.post<Redirect>('/redirects', {
        from_path: normalizePath(form.from_path),
        to_url: form.to_url.trim(),
        status_code: form.status_code,
        preserve_query: form.preserve_query,
        enabled: form.enabled,
        notes: form.notes || null,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-redirects'] });
      setCreating(false);
      setCreateForm(emptyForm());
      showToast({ type: 'success', title: 'Redirect created' });
    },
    onError: (err: unknown) => {
      const message = err instanceof Error ? err.message : 'Failed to create redirect';
      showToast({ type: 'error', title: message });
    },
  });

  const updateMutation = useMutation({
    mutationFn: async (vars: { id: string; form: RedirectForm }) =>
      adminApiClient.put<Redirect>(`/redirects/${vars.id}`, {
        from_path: normalizePath(vars.form.from_path),
        to_url: vars.form.to_url.trim(),
        status_code: vars.form.status_code,
        preserve_query: vars.form.preserve_query,
        enabled: vars.form.enabled,
        notes: vars.form.notes || null,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-redirects'] });
      setEditing(null);
      showToast({ type: 'success', title: 'Redirect updated' });
    },
    onError: (err: unknown) => {
      const message = err instanceof Error ? err.message : 'Failed to update redirect';
      showToast({ type: 'error', title: message });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (id: string) => adminApiClient.delete(`/redirects/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-redirects'] });
      showToast({ type: 'success', title: 'Redirect deleted' });
    },
    onError: (err: unknown) => {
      const message = err instanceof Error ? err.message : 'Failed to delete redirect';
      showToast({ type: 'error', title: message });
    },
  });

  const toggleEnabled = (r: Redirect) => {
    updateMutation.mutate({
      id: r.id,
      form: {
        from_path: r.from_path,
        to_url: r.to_url,
        status_code: r.status_code,
        preserve_query: r.preserve_query ?? true,
        enabled: !(r.enabled ?? true),
        notes: r.notes ?? '',
      },
    });
  };

  const filtered = useMemo(() => {
    if (!data) return [] as Redirect[];
    const q = search.trim().toLowerCase();
    if (!q) return data;
    return data.filter(
      (r) =>
        r.from_path.toLowerCase().includes(q) ||
        r.to_url.toLowerCase().includes(q) ||
        (r.notes ?? '').toLowerCase().includes(q)
    );
  }, [data, search]);

  if (isLoading) return <LoadingScreen />;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between gap-4 flex-wrap">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">URL Redirects</h1>
          <p className="mt-2 text-gray-600 dark:text-gray-400">
            Manage HTTP redirect rules for marketing campaigns, renamed pages, and SEO migrations.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => refetch()}
            className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-800 flex items-center gap-2"
          >
            <RefreshCw className="w-4 h-4" /> Refresh
          </button>
          <button
            onClick={() => {
              setCreateForm(emptyForm());
              setCreating(true);
            }}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 flex items-center gap-2"
          >
            <Plus className="w-4 h-4" /> New redirect
          </button>
        </div>
      </div>

      <div className="flex items-center gap-2 max-w-md">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-3 w-4 h-4 text-gray-400" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Filter by from, to, or notes…"
            className="w-full pl-9 pr-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-900 text-sm"
          />
        </div>
        <div className="text-sm text-gray-500 dark:text-gray-400">
          {filtered.length} of {data?.length ?? 0}
        </div>
      </div>

      {isError && (
        <div className="p-4 bg-amber-50 border border-amber-200 text-amber-800 rounded-lg text-sm">
          The redirects API returned an error. The endpoint may not be enabled on this backend —
          create/update/delete are still attempted and will surface the underlying error.
        </div>
      )}

      <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-950/40 text-gray-600 dark:text-gray-300">
            <tr>
              <th className="text-left font-medium px-4 py-2">From</th>
              <th className="text-left font-medium px-4 py-2">To</th>
              <th className="text-left font-medium px-4 py-2">Status</th>
              <th className="text-left font-medium px-4 py-2">Enabled</th>
              <th className="text-right font-medium px-4 py-2 w-40">Actions</th>
            </tr>
          </thead>
          <tbody>
            {filtered.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-gray-500 dark:text-gray-400">
                  No redirects configured yet.
                </td>
              </tr>
            ) : (
              filtered.map((r) => (
                <tr
                  key={r.id}
                  className="border-t border-gray-100 dark:border-gray-800 hover:bg-gray-50/60 dark:hover:bg-gray-950/40"
                >
                  <td className="px-4 py-2 font-mono text-xs text-gray-800 dark:text-gray-200">
                    {r.from_path}
                  </td>
                  <td className="px-4 py-2 max-w-[24rem]">
                    <div className="flex items-center gap-1.5 text-gray-700 dark:text-gray-200">
                      <ArrowRight className="w-3.5 h-3.5 text-gray-400 shrink-0" />
                      <span className="truncate">{r.to_url}</span>
                    </div>
                    {r.notes && (
                      <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5 truncate">
                        {r.notes}
                      </p>
                    )}
                  </td>
                  <td className="px-4 py-2 text-xs text-gray-700 dark:text-gray-200">
                    {r.status_code}
                    {r.preserve_query && (
                      <span className="ml-1 text-[10px] text-gray-500 dark:text-gray-400">
                        (preserve query)
                      </span>
                    )}
                  </td>
                  <td className="px-4 py-2">
                    <label className="inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={r.enabled ?? true}
                        onChange={() => toggleEnabled(r)}
                        className="sr-only peer"
                      />
                      <span className="w-9 h-5 bg-gray-200 dark:bg-gray-700 rounded-full peer peer-checked:bg-blue-600 relative transition-colors after:content-[''] after:absolute after:top-0.5 after:left-0.5 after:w-4 after:h-4 after:rounded-full after:bg-white after:transition-transform peer-checked:after:translate-x-4" />
                    </label>
                  </td>
                  <td className="px-4 py-2 text-right">
                    <div className="inline-flex items-center gap-1">
                      <button
                        onClick={() => {
                          setEditing(r);
                          setEditForm({
                            from_path: r.from_path,
                            to_url: r.to_url,
                            status_code: r.status_code,
                            preserve_query: r.preserve_query ?? true,
                            enabled: r.enabled ?? true,
                            notes: r.notes ?? '',
                          });
                        }}
                        className="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-600 dark:text-gray-300"
                        aria-label="Edit redirect"
                      >
                        <Pencil className="w-4 h-4" />
                      </button>
                      <button
                        onClick={() => {
                          if (confirm(`Delete redirect ${r.from_path}?`)) {
                            deleteMutation.mutate(r.id);
                          }
                        }}
                        className="p-1.5 rounded hover:bg-red-50 dark:hover:bg-red-950/40 text-red-600"
                        aria-label="Delete redirect"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {creating && (
        <RedirectFormModal
          title="New redirect"
          form={createForm}
          onChange={setCreateForm}
          onCancel={() => setCreating(false)}
          onSubmit={() => createMutation.mutate(createForm)}
          submitting={createMutation.isPending}
        />
      )}

      {editing && (
        <RedirectFormModal
          title={`Edit ${editing.from_path}`}
          form={editForm}
          onChange={setEditForm}
          onCancel={() => setEditing(null)}
          onSubmit={() => updateMutation.mutate({ id: editing.id, form: editForm })}
          submitting={updateMutation.isPending}
        />
      )}
    </div>
  );
}

interface RedirectFormModalProps {
  title: string;
  form: RedirectForm;
  onChange: (form: RedirectForm) => void;
  onCancel: () => void;
  onSubmit: () => void;
  submitting: boolean;
}

function RedirectFormModal({
  title,
  form,
  onChange,
  onCancel,
  onSubmit,
  submitting,
}: RedirectFormModalProps) {
  const fromPath = normalizePath(form.from_path);
  const isValid =
    fromPath.length > 1 &&
    /^https?:\/\//i.test(form.to_url.trim()) &&
    STATUS_CODE_OPTIONS.includes(form.status_code);

  return (
    <div
      className="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="redirect-form-title"
    >
      <form
        onSubmit={(e) => {
          e.preventDefault();
          if (isValid) onSubmit();
        }}
        className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg w-full max-w-lg p-5 space-y-3"
      >
        <div className="flex items-center justify-between">
          <h2 id="redirect-form-title" className="text-lg font-semibold">
            {title}
          </h2>
          <button
            type="button"
            onClick={onCancel}
            className="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800"
            aria-label="Close"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-[1fr_auto_1fr] gap-2 items-end">
          <label className="block text-sm">
            <span className="text-gray-700 dark:text-gray-300">From path</span>
            <input
              type="text"
              required
              value={form.from_path}
              onChange={(e) => onChange({ ...form, from_path: e.target.value })}
              placeholder="/old-page"
              className="mt-1 w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-950 font-mono text-sm"
            />
            <span className="text-[10px] text-gray-500">Normalized: {fromPath}</span>
          </label>
          <div className="hidden sm:flex pb-2 text-gray-400">
            <ArrowRight className="w-5 h-5" />
          </div>
          <label className="block text-sm">
            <span className="text-gray-700 dark:text-gray-300">To URL</span>
            <input
              type="url"
              required
              value={form.to_url}
              onChange={(e) => onChange({ ...form, to_url: e.target.value })}
              placeholder="https://example.com/new-page"
              className="mt-1 w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-950 text-sm"
            />
          </label>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-sm">
          <label className="block">
            <span className="text-gray-700 dark:text-gray-300">Status code</span>
            <select
              value={form.status_code}
              onChange={(e) =>
                onChange({ ...form, status_code: Number(e.target.value) as RedirectStatus })
              }
              className="mt-1 w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-950"
            >
              {STATUS_CODE_OPTIONS.map((s) => (
                <option key={s} value={s}>
                  {STATUS_CODE_LABEL[s]}
                </option>
              ))}
            </select>
          </label>
          <div className="space-y-2 pt-6">
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={form.preserve_query}
                onChange={(e) => onChange({ ...form, preserve_query: e.target.checked })}
              />
              <span>Preserve query string</span>
            </label>
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={form.enabled}
                onChange={(e) => onChange({ ...form, enabled: e.target.checked })}
              />
              <span>Enabled</span>
            </label>
          </div>
        </div>

        <label className="block text-sm">
          <span className="text-gray-700 dark:text-gray-300">Notes (optional)</span>
          <textarea
            value={form.notes}
            onChange={(e) => onChange({ ...form, notes: e.target.value })}
            rows={2}
            className="mt-1 w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-950"
            placeholder="Why does this redirect exist? (for the team)"
          />
        </label>

        <div className="flex items-center justify-end gap-2 pt-2">
          <button
            type="button"
            onClick={onCancel}
            className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-800"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={!isValid || submitting}
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
          >
            {submitting ? 'Saving…' : 'Save'}
          </button>
        </div>
      </form>
    </div>
  );
}

export default AdminRedirectsPage;
