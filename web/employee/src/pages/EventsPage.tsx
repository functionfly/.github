import { useState } from 'react';
import { Radio, Filter, Search, ChevronRight, Clock } from 'lucide-react';

interface Event {
  id: string;
  event_type: string;
  resource_type: string;
  resource_id: string;
  actor_id: string;
  payload: Record<string, unknown>;
  timestamp: string;
}

const eventTypeColors: Record<string, string> = {
  created: 'bg-green-500/20 text-green-400',
  updated: 'bg-blue-500/20 text-blue-400',
  deleted: 'bg-red-500/20 text-red-400',
  login: 'bg-purple-500/20 text-purple-400',
  logout: 'bg-gray-500/20 text-gray-400',
  permission_changed: 'bg-yellow-500/20 text-yellow-400',
  incident_declared: 'bg-red-500/20 text-red-400',
  workflow_completed: 'bg-cyan-500/20 text-cyan-400',
};

const mockEvents: Event[] = [
  { id: '1', event_type: 'created', resource_type: 'project', resource_id: 'proj-001', actor_id: 'emp-001', payload: { name: 'FWOS Core', status: 'active' }, timestamp: '2026-06-23T10:30:00Z' },
  { id: '2', event_type: 'login', resource_type: 'session', resource_id: 'sess-abc', actor_id: 'emp-002', payload: { ip: '192.168.1.100', user_agent: 'Chrome/126' }, timestamp: '2026-06-23T10:25:00Z' },
  { id: '3', event_type: 'updated', resource_type: 'employee', resource_id: 'emp-003', actor_id: 'emp-001', payload: { field: 'role', old: 'developer', new: 'senior_developer' }, timestamp: '2026-06-23T10:20:00Z' },
  { id: '4', event_type: 'incident_declared', resource_type: 'incident', resource_id: 'inc-001', actor_id: 'emp-004', payload: { title: 'API outage', severity: 'critical' }, timestamp: '2026-06-23T10:15:00Z' },
  { id: '5', event_type: 'permission_changed', resource_type: 'role', resource_id: 'role-dev', actor_id: 'emp-001', payload: { permission: 'admin.access', granted: true }, timestamp: '2026-06-23T10:10:00Z' },
  { id: '6', event_type: 'created', resource_type: 'document', resource_id: 'doc-010', actor_id: 'emp-005', payload: { title: 'Q3 Planning', type: 'meeting_notes' }, timestamp: '2026-06-23T10:05:00Z' },
  { id: '7', event_type: 'workflow_completed', resource_type: 'lifecycle', resource_id: 'wf-inst-003', actor_id: 'system', payload: { workflow: 'onboarding', employee: 'emp-006' }, timestamp: '2026-06-23T10:00:00Z' },
  { id: '8', event_type: 'deleted', resource_type: 'task', resource_id: 'task-042', actor_id: 'emp-002', payload: { title: 'Deprecated cleanup' }, timestamp: '2026-06-23T09:55:00Z' },
  { id: '9', event_type: 'updated', resource_type: 'feature_flag', resource_id: 'ff-001', actor_id: 'emp-001', payload: { key: 'new_dashboard', rollout_pct: 50 }, timestamp: '2026-06-23T09:50:00Z' },
  { id: '10', event_type: 'logout', resource_type: 'session', resource_id: 'sess-def', actor_id: 'emp-003', payload: {}, timestamp: '2026-06-23T09:45:00Z' },
];

export function EventsPage() {
  const [search, setSearch] = useState('');
  const [typeFilter, setTypeFilter] = useState('');
  const [resourceFilter, setResourceFilter] = useState('');
  const [selected, setSelected] = useState<string | null>(null);

  const events = mockEvents.filter((e) => {
    if (typeFilter && e.event_type !== typeFilter) return false;
    if (resourceFilter && e.resource_type !== resourceFilter) return false;
    if (search && !e.event_type.includes(search) && !e.resource_type.includes(search) && !e.resource_id.includes(search)) return false;
    return true;
  });

  const selectedEvent = events.find((e) => e.id === selected);

  const uniqueTypes = [...new Set(mockEvents.map((e) => e.event_type))];
  const uniqueResources = [...new Set(mockEvents.map((e) => e.resource_type))];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Radio className="h-6 w-6 text-emerald-400" />
          <h1 className="text-2xl font-bold">Event Bus Viewer</h1>
        </div>
        <div className="flex items-center gap-2">
          <span className="flex h-2 w-2 animate-pulse rounded-full bg-emerald-500" />
          <span className="text-xs text-gray-400">Live</span>
        </div>
      </div>

      <div className="flex items-center gap-3">
        <Filter className="h-4 w-4 text-gray-400" />
        <select
          value={typeFilter}
          onChange={(e) => setTypeFilter(e.target.value)}
          className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
        >
          <option value="">All Event Types</option>
          {uniqueTypes.map((t) => (
            <option key={t} value={t}>{t}</option>
          ))}
        </select>
        <select
          value={resourceFilter}
          onChange={(e) => setResourceFilter(e.target.value)}
          className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
        >
          <option value="">All Resources</option>
          {uniqueResources.map((r) => (
            <option key={r} value={r}>{r}</option>
          ))}
        </select>
        <div className="flex items-center gap-2">
          <Search className="h-4 w-4 text-gray-400" />
          <input
            type="text"
            placeholder="Search events..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-48 rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
          />
        </div>
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-2">
          {events.length === 0 ? (
            <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
              <Radio className="mb-4 h-12 w-12 text-gray-600" />
              <p className="text-gray-400">No events match filters</p>
            </div>
          ) : (
            events.map((ev) => (
              <button
                key={ev.id}
                onClick={() => setSelected(ev.id === selected ? null : ev.id)}
                className={`w-full rounded-xl border p-3 text-left transition-colors ${
                  ev.id === selected
                    ? 'border-blue-600 bg-gray-800'
                    : 'border-gray-800 bg-gray-900 hover:bg-gray-800'
                }`}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <span className={`rounded-full px-2 py-0.5 text-xs ${eventTypeColors[ev.event_type] || 'bg-gray-500/20 text-gray-400'}`}>
                      {ev.event_type}
                    </span>
                    <span className="text-sm text-gray-300">{ev.resource_type}</span>
                    <code className="rounded bg-gray-800 px-1.5 py-0.5 text-xs text-gray-500">{ev.resource_id}</code>
                  </div>
                  <div className="flex items-center gap-2 text-xs text-gray-500">
                    <Clock className="h-3 w-3" />
                    {new Date(ev.timestamp).toLocaleTimeString()}
                    <ChevronRight className={`h-3 w-3 transition-transform ${ev.id === selected ? 'rotate-90' : ''}`} />
                  </div>
                </div>
              </button>
            ))
          )}
        </div>

        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <h3 className="mb-3 text-sm font-medium text-gray-300">Event Detail</h3>
          {selectedEvent ? (
            <div className="space-y-3">
              <div>
                <span className="text-xs text-gray-500">Event Type</span>
                <p className={`mt-0.5 text-sm ${eventTypeColors[selectedEvent.event_type]?.split(' ')[1] || 'text-gray-300'}`}>{selectedEvent.event_type}</p>
              </div>
              <div>
                <span className="text-xs text-gray-500">Resource</span>
                <p className="mt-0.5 text-sm text-gray-200">{selectedEvent.resource_type} / {selectedEvent.resource_id}</p>
              </div>
              <div>
                <span className="text-xs text-gray-500">Actor</span>
                <p className="mt-0.5 text-sm text-gray-200">{selectedEvent.actor_id}</p>
              </div>
              <div>
                <span className="text-xs text-gray-500">Timestamp</span>
                <p className="mt-0.5 text-sm text-gray-200">{new Date(selectedEvent.timestamp).toLocaleString()}</p>
              </div>
              <div>
                <span className="text-xs text-gray-500">Payload</span>
                <pre className="mt-1 overflow-x-auto rounded-lg bg-gray-800 p-3 text-xs text-gray-300">
                  {JSON.stringify(selectedEvent.payload, null, 2)}
                </pre>
              </div>
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-8">
              <Radio className="mb-2 h-8 w-8 text-gray-600" />
              <p className="text-sm text-gray-500">Select an event to view details</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
