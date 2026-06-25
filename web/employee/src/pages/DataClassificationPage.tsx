import { useState } from 'react';
import { ShieldCheck, Plus, Search } from 'lucide-react';

interface ClassificationEntry {
  id: string;
  resource_type: string;
  resource_id: string;
  resource_name: string;
  classification: string;
  classified_by: string;
  classified_at: string;
  notes?: string;
}

const classificationColors: Record<string, string> = {
  public: 'bg-green-500/20 text-green-400',
  internal: 'bg-blue-500/20 text-blue-400',
  confidential: 'bg-yellow-500/20 text-yellow-400',
  restricted: 'bg-red-500/20 text-red-400',
  founder: 'bg-purple-500/20 text-purple-400',
};

const classificationBorders: Record<string, string> = {
  public: 'border-green-500/30',
  internal: 'border-blue-500/30',
  confidential: 'border-yellow-500/30',
  restricted: 'border-red-500/30',
  founder: 'border-purple-500/30',
};

const mockData: ClassificationEntry[] = [
  { id: '1', resource_type: 'document', resource_id: 'doc-001', resource_name: 'Company Handbook', classification: 'internal', classified_by: 'admin', classified_at: '2026-01-15T10:00:00Z', notes: 'Standard internal document' },
  { id: '2', resource_type: 'document', resource_id: 'doc-002', resource_name: 'API Keys Vault', classification: 'restricted', classified_by: 'admin', classified_at: '2026-01-20T14:00:00Z' },
  { id: '3', resource_type: 'project', resource_id: 'proj-001', resource_name: 'FWOS Core', classification: 'confidential', classified_by: 'security-team', classified_at: '2026-02-01T09:00:00Z', notes: 'Contains proprietary algorithms' },
  { id: '4', resource_type: 'document', resource_id: 'doc-003', resource_name: 'Public Blog Post', classification: 'public', classified_by: 'marketing', classified_at: '2026-02-10T16:00:00Z' },
  { id: '5', resource_type: 'project', resource_id: 'proj-002', resource_name: 'Founder Roadmap', classification: 'founder', classified_by: 'founder', classified_at: '2026-02-15T11:00:00Z', notes: 'Eyes-only strategic plans' },
  { id: '6', resource_type: 'document', resource_id: 'doc-004', resource_name: 'Incident Playbook', classification: 'confidential', classified_by: 'security-team', classified_at: '2026-03-01T08:00:00Z' },
  { id: '7', resource_type: 'dataset', resource_id: 'ds-001', resource_name: 'Customer Analytics', classification: 'restricted', classified_by: 'data-team', classified_at: '2026-03-05T13:00:00Z' },
  { id: '8', resource_type: 'document', resource_id: 'doc-005', resource_name: 'Open Source License', classification: 'public', classified_by: 'legal', classified_at: '2026-03-10T10:00:00Z' },
];

const levelOrder = ['public', 'internal', 'confidential', 'restricted', 'founder'];

export function DataClassificationPage() {
  const [search, setSearch] = useState('');
  const [showClassify, setShowClassify] = useState(false);
  const [filterLevel, setFilterLevel] = useState('');
  const [form, setForm] = useState({ resource_type: 'document', resource_name: '', classification: 'internal', notes: '' });

  const entries = mockData.filter((e) => {
    if (filterLevel && e.classification !== filterLevel) return false;
    if (search && !e.resource_name.toLowerCase().includes(search.toLowerCase())) return false;
    return true;
  });

  const grouped = levelOrder.reduce((acc, level) => {
    const items = entries.filter((e) => e.classification === level);
    if (items.length > 0) acc.push({ level, items });
    return acc;
  }, [] as { level: string; items: ClassificationEntry[] }[]);

  const counts = levelOrder.map((level) => ({
    level,
    count: mockData.filter((e) => e.classification === level).length,
  }));

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <ShieldCheck className="h-6 w-6 text-yellow-400" />
          <h1 className="text-2xl font-bold">Data Classification</h1>
        </div>
        <button
          onClick={() => setShowClassify(true)}
          className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          <Plus className="h-4 w-4" />
          Classify Resource
        </button>
      </div>

      <div className="grid grid-cols-5 gap-3">
        {counts.map(({ level, count }) => (
          <button
            key={level}
            onClick={() => setFilterLevel(filterLevel === level ? '' : level)}
            className={`rounded-xl border p-3 text-center transition-colors ${
              filterLevel === level ? classificationBorders[level] : 'border-gray-800'
            } bg-gray-900`}
          >
            <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${classificationColors[level]}`}>{level}</span>
            <p className="mt-1 text-2xl font-bold text-gray-100">{count}</p>
          </button>
        ))}
      </div>

      <div className="flex items-center gap-3">
        <Search className="h-4 w-4 text-gray-400" />
        <input
          type="text"
          placeholder="Search resources..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-64 rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
        />
      </div>

      {grouped.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
          <ShieldCheck className="mb-4 h-12 w-12 text-gray-600" />
          <p className="text-gray-400">No classified resources found</p>
        </div>
      ) : (
        <div className="space-y-6">
          {grouped.map(({ level, items }) => (
            <div key={level}>
              <div className="mb-3 flex items-center gap-2">
                <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${classificationColors[level]}`}>{level}</span>
                <span className="text-sm text-gray-500">{items.length} resource{items.length !== 1 ? 's' : ''}</span>
              </div>
              <div className="space-y-2">
                {items.map((entry) => (
                  <div key={entry.id} className={`rounded-xl border p-4 ${classificationBorders[level]} bg-gray-900`}>
                    <div className="flex items-center justify-between">
                      <div>
                        <h3 className="font-medium text-gray-100">{entry.resource_name}</h3>
                        <div className="mt-1 flex items-center gap-3 text-xs text-gray-500">
                          <span className="rounded bg-gray-800 px-2 py-0.5 text-gray-400">{entry.resource_type}</span>
                          <span>ID: {entry.resource_id}</span>
                          <span>By: {entry.classified_by}</span>
                          <span>{new Date(entry.classified_at).toLocaleDateString()}</span>
                        </div>
                        {entry.notes && <p className="mt-1 text-sm text-gray-400">{entry.notes}</p>}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      {showClassify && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Classify Resource</h2>
            <select
              value={form.resource_type}
              onChange={(e) => setForm({ ...form, resource_type: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
            >
              <option value="document">Document</option>
              <option value="project">Project</option>
              <option value="dataset">Dataset</option>
              <option value="api">API</option>
              <option value="secret">Secret</option>
            </select>
            <input
              type="text"
              placeholder="Resource name"
              value={form.resource_name}
              onChange={(e) => setForm({ ...form, resource_name: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              autoFocus
            />
            <select
              value={form.classification}
              onChange={(e) => setForm({ ...form, classification: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
            >
              {levelOrder.map((l) => (
                <option key={l} value={l}>{l.charAt(0).toUpperCase() + l.slice(1)}</option>
              ))}
            </select>
            <textarea
              placeholder="Notes (optional)"
              value={form.notes}
              onChange={(e) => setForm({ ...form, notes: e.target.value })}
              className="mb-4 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              rows={3}
            />
            <div className="flex justify-end gap-3">
              <button onClick={() => setShowClassify(false)} className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200">Cancel</button>
              <button
                onClick={() => setShowClassify(false)}
                disabled={!form.resource_name.trim()}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                Classify
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
