import { useMemo } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';

/** Flat item for display (from backend category + measures or wrapped data) */
interface FeatureMeasureItem {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
}

/** Backend flat format from DB: { measures: Array<{ id, name, description, category, enabled }> } */
interface SecurityMeasureFlat {
  id?: string;
  name?: string;
  description?: string;
  category?: string;
  enabled?: boolean;
}

/** Legacy: { measures: Array<{ category, icon, measures: string[] }> } */
interface SecurityMeasureCategory {
  category?: string;
  icon?: string;
  measures?: string[];
}

/** Normalize backend response into a flat list for the UI */
function normalizeMeasures(raw: unknown): FeatureMeasureItem[] {
  if (!raw || typeof raw !== 'object') return [];
  const obj = raw as Record<string, unknown>;
  // Wrapped admin response: { data: FeatureMeasureItem[] }
  const wrapped = obj.data;
  if (Array.isArray(wrapped) && wrapped.length > 0) {
    return wrapped.map((m: Record<string, unknown>, idx: number) => ({
      id: String((m.id as string) ?? idx),
      name: String(m.name ?? 'Unnamed measure'),
      description: String(m.description ?? ''),
      enabled: Boolean(m.enabled),
    }));
  }
  const measuresArr = obj.measures as (SecurityMeasureFlat | SecurityMeasureCategory)[] | undefined;
  if (!Array.isArray(measuresArr) || measuresArr.length === 0) return [];
  // New flat format from DB: each item has id, name, description, enabled
  const first = measuresArr[0] as Record<string, unknown>;
  if (first && typeof first.enabled === 'boolean' && (first.id != null || first.name != null)) {
    return (measuresArr as SecurityMeasureFlat[]).map((m, idx) => ({
      id: String(m.id ?? idx),
      name: String(m.name ?? 'Unnamed measure'),
      description: String(m.description ?? m.category ?? ''),
      enabled: Boolean(m.enabled),
    }));
  }
  // Legacy category format: { category, icon, measures: string[] }
  const flat: FeatureMeasureItem[] = [];
  let id = 0;
  for (const cat of measuresArr as SecurityMeasureCategory[]) {
    const list = cat.measures ?? [];
    const categoryName = cat.category ?? 'General';
    for (const measure of list) {
      flat.push({
        id: String(id++),
        name: measure,
        description: categoryName,
        enabled: true,
      });
    }
  }
  return flat;
}

export function AdminFeaturesPage() {
  const queryClient = useQueryClient();
  const { data, isLoading } = useQuery({
    queryKey: ['admin-feature-measures'],
    queryFn: async () => {
      try {
        const res = await adminApiClient.get<unknown>('/security/measures');
        return res;
      } catch {
        return { data: [], success: false, timestamp: new Date().toISOString() };
      }
    },
  });

  const updateEnabled = useMutation({
    mutationFn: async ({ id, enabled }: { id: string; enabled: boolean }) => {
      await adminApiClient.patch<{ ok: boolean }>(`/security/measures/${id}`, { enabled });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-feature-measures'] });
    },
  });

  const measures = useMemo(() => normalizeMeasures(data), [data]);
  const canToggle = measures.length > 0 && measures.some((m) => m.id && m.id.length === 36);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Features</h1>
        <p className="mt-2 text-gray-600">Platform security and operational feature controls.</p>
      </div>

      <div className="bg-white rounded-lg border border-gray-200 divide-y divide-gray-100">
        {isLoading ? (
          <div className="p-6 text-gray-500">Loading…</div>
        ) : measures.length === 0 ? (
          <div className="p-6 text-gray-500">No feature measures returned.</div>
        ) : (
          measures.map((measure) => (
            <div key={measure.id} className="p-6 flex items-center justify-between gap-4">
              <div className="min-w-0 flex-1">
                <h2 className="text-sm font-semibold text-gray-900">{measure.name}</h2>
                <p className="text-sm text-gray-600 mt-1">{measure.description || 'No description provided.'}</p>
              </div>
              <div className="flex items-center gap-2 shrink-0">
                {canToggle ? (
                  <button
                    type="button"
                    onClick={() => updateEnabled.mutate({ id: measure.id, enabled: !measure.enabled })}
                    disabled={updateEnabled.isPending && (updateEnabled.variables?.id === measure.id)}
                    className={`text-xs font-medium px-3 py-1.5 rounded border transition-colors ${
                      measure.enabled
                        ? 'bg-green-50 border-green-200 text-green-800 hover:bg-green-100'
                        : 'bg-gray-50 border-gray-200 text-gray-600 hover:bg-gray-100'
                    } disabled:opacity-50`}
                  >
                    {(updateEnabled.isPending && updateEnabled.variables?.id === measure.id) ? '…' : measure.enabled ? 'Enabled' : 'Disabled'}
                  </button>
                ) : (
                  <span className={`text-xs px-2 py-1 rounded ${measure.enabled ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-700'}`}>
                    {measure.enabled ? 'Enabled' : 'Disabled'}
                  </span>
                )}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
