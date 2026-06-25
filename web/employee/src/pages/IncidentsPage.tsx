import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { incidentsApi, type Incident } from '@/api/incidents';
import { Siren, Plus, Filter, MessageSquare, UserPlus, FileText, Clock, AlertTriangle, ChevronRight } from 'lucide-react';

const severityColors: Record<string, string> = {
  critical: 'bg-red-500/20 text-red-400',
  high: 'bg-orange-500/20 text-orange-400',
  medium: 'bg-yellow-500/20 text-yellow-400',
  low: 'bg-gray-500/20 text-gray-400',
};

const statusColors: Record<string, string> = {
  open: 'bg-blue-500/20 text-blue-400',
  investigating: 'bg-purple-500/20 text-purple-400',
  monitoring: 'bg-cyan-500/20 text-cyan-400',
  resolved: 'bg-green-500/20 text-green-400',
  closed: 'bg-gray-500/20 text-gray-400',
};

function formatDuration(minutes?: number) {
  if (!minutes) return '-';
  if (minutes < 60) return `${minutes}m`;
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  return m > 0 ? `${h}h ${m}m` : `${h}h`;
}

export function IncidentsPage() {
  const queryClient = useQueryClient();
  const [statusFilter, setStatusFilter] = useState('');
  const [severityFilter, setSeverityFilter] = useState('');
  const [selected, setSelected] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [showAddEvent, setShowAddEvent] = useState(false);
  const [showAddResponder, setShowAddResponder] = useState(false);
  const [eventBody, setEventBody] = useState('');
  const [responderId, setResponderId] = useState('');
  const [form, setForm] = useState({ title: '', severity: 'medium', description: '' });

  const { data, isLoading } = useQuery({
    queryKey: ['incidents', statusFilter, severityFilter],
    queryFn: () => incidentsApi.list({
      ...(statusFilter && { status: statusFilter }),
      ...(severityFilter && { severity: severityFilter }),
    }),
  });

  const { data: detailData } = useQuery({
    queryKey: ['incident', selected],
    queryFn: () => incidentsApi.get(selected!),
    enabled: !!selected,
  });

  const { data: eventsData } = useQuery({
    queryKey: ['incident-events', selected],
    queryFn: () => incidentsApi.listEvents(selected!),
    enabled: !!selected,
  });

  const { data: postmortemData } = useQuery({
    queryKey: ['incident-postmortem', selected],
    queryFn: () => incidentsApi.getPostmortem(selected!),
    enabled: !!selected,
  });

  const createMutation = useMutation({
    mutationFn: (data: Partial<Incident>) => incidentsApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['incidents'] });
      setShowCreate(false);
      setForm({ title: '', severity: 'medium', description: '' });
    },
  });

  const eventMutation = useMutation({
    mutationFn: ({ id, body }: { id: string; body: string }) => incidentsApi.addEvent(id, body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['incident-events'] });
      setShowAddEvent(false);
      setEventBody('');
    },
  });

  const responderMutation = useMutation({
    mutationFn: ({ id, employeeId }: { id: string; employeeId: string }) => incidentsApi.addResponder(id, employeeId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['incident', selected] });
      setShowAddResponder(false);
      setResponderId('');
    },
  });

  const incidents = data?.data?.incidents || [];
  const incident = detailData?.data?.incident;
  const events = eventsData?.data?.events || [];
  const postmortem = postmortemData?.data?.postmortem;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Siren className="h-6 w-6 text-red-400" />
          <h1 className="text-2xl font-bold">Incident Command Center</h1>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-2 rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700"
        >
          <Plus className="h-4 w-4" />
          Declare Incident
        </button>
      </div>

      <div className="flex items-center gap-3">
        <Filter className="h-4 w-4 text-gray-400" />
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
        >
          <option value="">All Statuses</option>
          <option value="open">Open</option>
          <option value="investigating">Investigating</option>
          <option value="monitoring">Monitoring</option>
          <option value="resolved">Resolved</option>
          <option value="closed">Closed</option>
        </select>
        <select
          value={severityFilter}
          onChange={(e) => setSeverityFilter(e.target.value)}
          className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
        >
          <option value="">All Severities</option>
          <option value="critical">Critical</option>
          <option value="high">High</option>
          <option value="medium">Medium</option>
          <option value="low">Low</option>
        </select>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : incidents.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
          <Siren className="mb-4 h-12 w-12 text-gray-600" />
          <p className="text-gray-400">No incidents found</p>
        </div>
      ) : (
        <div className="space-y-3">
          {incidents.map((inc) => (
            <button
              key={inc.id}
              onClick={() => setSelected(inc.id === selected ? null : inc.id)}
              className={`w-full rounded-xl border p-4 text-left transition-colors ${
                inc.id === selected
                  ? 'border-blue-600 bg-gray-800'
                  : 'border-gray-800 bg-gray-900 hover:bg-gray-800'
              }`}
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <AlertTriangle className={`h-5 w-5 ${inc.severity === 'critical' ? 'text-red-400' : inc.severity === 'high' ? 'text-orange-400' : 'text-yellow-400'}`} />
                  <div>
                    <h3 className="font-medium text-gray-100">{inc.title}</h3>
                    <div className="mt-1 flex items-center gap-2 text-xs text-gray-500">
                      <span className={`rounded-full px-2 py-0.5 ${severityColors[inc.severity] || ''}`}>{inc.severity}</span>
                      <span className={`rounded-full px-2 py-0.5 ${statusColors[inc.status] || ''}`}>{inc.status}</span>
                      <span className="flex items-center gap-1"><Clock className="h-3 w-3" />{formatDuration(inc.duration_minutes)}</span>
                      <span>{new Date(inc.detected_at).toLocaleString()}</span>
                    </div>
                  </div>
                </div>
                <ChevronRight className={`h-4 w-4 text-gray-500 transition-transform ${inc.id === selected ? 'rotate-90' : ''}`} />
              </div>
            </button>
          ))}
        </div>
      )}

      {selected && incident && (
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-6 space-y-6">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-gray-100">{incident.title}</h2>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setShowAddEvent(true)}
                className="flex items-center gap-2 rounded-lg bg-gray-800 px-3 py-2 text-sm text-gray-300 hover:bg-gray-700"
              >
                <MessageSquare className="h-4 w-4" />
                Add Event
              </button>
              <button
                onClick={() => setShowAddResponder(true)}
                className="flex items-center gap-2 rounded-lg bg-gray-800 px-3 py-2 text-sm text-gray-300 hover:bg-gray-700"
              >
                <UserPlus className="h-4 w-4" />
                Add Responder
              </button>
            </div>
          </div>

          {incident.description && <p className="text-sm text-gray-400">{incident.description}</p>}

          <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
            <div className="rounded-lg bg-gray-800 p-3">
              <span className="text-xs text-gray-500">Severity</span>
              <p className={`mt-1 text-sm font-medium ${severityColors[incident.severity]?.split(' ')[1] || ''}`}>{incident.severity}</p>
            </div>
            <div className="rounded-lg bg-gray-800 p-3">
              <span className="text-xs text-gray-500">Status</span>
              <p className={`mt-1 text-sm font-medium ${statusColors[incident.status]?.split(' ')[1] || ''}`}>{incident.status}</p>
            </div>
            <div className="rounded-lg bg-gray-800 p-3">
              <span className="text-xs text-gray-500">Duration</span>
              <p className="mt-1 text-sm font-medium text-gray-100">{formatDuration(incident.duration_minutes)}</p>
            </div>
            <div className="rounded-lg bg-gray-800 p-3">
              <span className="text-xs text-gray-500">Detected</span>
              <p className="mt-1 text-sm font-medium text-gray-100">{new Date(incident.detected_at).toLocaleString()}</p>
            </div>
          </div>

          {incident.root_cause && (
            <div>
              <h3 className="mb-2 text-sm font-medium text-gray-300">Root Cause</h3>
              <p className="text-sm text-gray-400">{incident.root_cause}</p>
            </div>
          )}

          {incident.impact && (
            <div>
              <h3 className="mb-2 text-sm font-medium text-gray-300">Impact</h3>
              <p className="text-sm text-gray-400">{incident.impact}</p>
            </div>
          )}

          <div>
            <h3 className="mb-3 text-sm font-medium text-gray-300">Timeline</h3>
            {events.length === 0 ? (
              <p className="text-sm text-gray-500">No events recorded</p>
            ) : (
              <div className="space-y-3">
                {events.map((ev) => (
                  <div key={ev.id} className="flex gap-3">
                    <div className="flex flex-col items-center">
                      <div className="h-2 w-2 rounded-full bg-blue-500" />
                      <div className="w-px flex-1 bg-gray-700" />
                    </div>
                    <div className="pb-3">
                      <div className="flex items-center gap-2 text-xs text-gray-500">
                        <span className="rounded-full bg-gray-700 px-2 py-0.5 text-gray-300">{ev.event_type}</span>
                        <span>{new Date(ev.created_at).toLocaleString()}</span>
                      </div>
                      <p className="mt-1 text-sm text-gray-300">{ev.body}</p>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {postmortem && (
            <div>
              <div className="mb-3 flex items-center gap-2">
                <FileText className="h-4 w-4 text-purple-400" />
                <h3 className="text-sm font-medium text-gray-300">Postmortem</h3>
                <span className={`rounded-full px-2 py-0.5 text-xs ${postmortem.status === 'published' ? 'bg-green-500/20 text-green-400' : 'bg-yellow-500/20 text-yellow-400'}`}>{postmortem.status}</span>
              </div>
              <div className="space-y-3 rounded-lg bg-gray-800 p-4">
                <div>
                  <span className="text-xs text-gray-500">Summary</span>
                  <p className="mt-1 text-sm text-gray-300">{postmortem.summary}</p>
                </div>
                <div>
                  <span className="text-xs text-gray-500">Root Cause</span>
                  <p className="mt-1 text-sm text-gray-300">{postmortem.root_cause}</p>
                </div>
                {postmortem.action_items.length > 0 && (
                  <div>
                    <span className="text-xs text-gray-500">Action Items</span>
                    <ul className="mt-1 space-y-1">
                      {postmortem.action_items.map((item, i) => (
                        <li key={i} className="flex items-center gap-2 text-sm text-gray-300">
                          <span className={`h-2 w-2 rounded-full ${item.status === 'done' ? 'bg-green-500' : item.status === 'in_progress' ? 'bg-yellow-500' : 'bg-gray-500'}`} />
                          {item.title}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      )}

      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Declare Incident</h2>
            <input
              type="text"
              placeholder="Incident title"
              value={form.title}
              onChange={(e) => setForm({ ...form, title: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              autoFocus
            />
            <select
              value={form.severity}
              onChange={(e) => setForm({ ...form, severity: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
            >
              <option value="critical">Critical</option>
              <option value="high">High</option>
              <option value="medium">Medium</option>
              <option value="low">Low</option>
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
                onClick={() => createMutation.mutate({ title: form.title, severity: form.severity, description: form.description || undefined })}
                disabled={!form.title.trim()}
                className="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-50"
              >
                Declare
              </button>
            </div>
          </div>
        </div>
      )}

      {showAddEvent && selected && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Add Event</h2>
            <textarea
              placeholder="Event details"
              value={eventBody}
              onChange={(e) => setEventBody(e.target.value)}
              className="mb-4 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              rows={3}
              autoFocus
            />
            <div className="flex justify-end gap-3">
              <button onClick={() => { setShowAddEvent(false); setEventBody(''); }} className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200">Cancel</button>
              <button
                onClick={() => eventMutation.mutate({ id: selected, body: eventBody })}
                disabled={!eventBody.trim()}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                Add Event
              </button>
            </div>
          </div>
        </div>
      )}

      {showAddResponder && selected && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-sm rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Add Responder</h2>
            <input
              type="text"
              placeholder="Employee ID"
              value={responderId}
              onChange={(e) => setResponderId(e.target.value)}
              className="mb-4 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              autoFocus
            />
            <div className="flex justify-end gap-3">
              <button onClick={() => { setShowAddResponder(false); setResponderId(''); }} className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200">Cancel</button>
              <button
                onClick={() => responderMutation.mutate({ id: selected, employeeId: responderId })}
                disabled={!responderId.trim()}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                Add
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
