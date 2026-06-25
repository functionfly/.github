import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { timeTrackingApi, type TimeEntry, type PTORequest } from '@/api/time_tracking';
import { Clock, Calendar, Plus, Trash2, Send, CheckCircle, XCircle } from 'lucide-react';
import { formatDate } from '@/lib/utils';

const ptoStatusColors: Record<string, string> = {
  pending: 'bg-yellow-500/20 text-yellow-400',
  approved: 'bg-green-500/20 text-green-400',
  rejected: 'bg-red-500/20 text-red-400',
  cancelled: 'bg-gray-500/20 text-gray-400',
};

function getWeekDays(date: Date): Date[] {
  const start = new Date(date);
  start.setDate(start.getDate() - start.getDay() + 1);
  return Array.from({ length: 7 }, (_, i) => {
    const d = new Date(start);
    d.setDate(d.getDate() + i);
    return d;
  });
}

function toDateString(d: Date): string {
  return d.toISOString().split('T')[0];
}

export function TimeTrackingPage() {
  const queryClient = useQueryClient();
  const [tab, setTab] = useState<'timesheet' | 'pto'>('timesheet');
  const [weekOffset, setWeekOffset] = useState(0);
  const [showEntryModal, setShowEntryModal] = useState(false);
  const [showPTOModal, setShowPTOModal] = useState(false);

  const baseDate = new Date();
  baseDate.setDate(baseDate.getDate() + weekOffset * 7);
  const weekDays = getWeekDays(baseDate);
  const weekStart = toDateString(weekDays[0]);
  const weekEnd = toDateString(weekDays[6]);

  const [newEntry, setNewEntry] = useState({ date: toDateString(new Date()), hours: 8, description: '', entry_type: 'work', is_billable: true, project_id: '' });
  const [newPTO, setNewPTO] = useState({ pto_type: 'vacation', start_date: '', end_date: '', reason: '' });

  const { data: entriesData } = useQuery({
    queryKey: ['time-entries', { start_date: weekStart, end_date: weekEnd }],
    queryFn: () => timeTrackingApi.listEntries({ start_date: weekStart, end_date: weekEnd }),
  });

  const { data: ptoData } = useQuery({
    queryKey: ['pto', 'requests'],
    queryFn: () => timeTrackingApi.listPTO(),
  });

  const createEntryMutation = useMutation({
    mutationFn: (data: Partial<TimeEntry>) => timeTrackingApi.createEntry(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['time-entries'] });
      setShowEntryModal(false);
      setNewEntry({ date: toDateString(new Date()), hours: 8, description: '', entry_type: 'work', is_billable: true, project_id: '' });
    },
  });

  const deleteEntryMutation = useMutation({
    mutationFn: (id: string) => timeTrackingApi.deleteEntry(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['time-entries'] }),
  });

  const requestPTOMutation = useMutation({
    mutationFn: (data: Partial<PTORequest>) => timeTrackingApi.requestPTO(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pto'] });
      setShowPTOModal(false);
      setNewPTO({ pto_type: 'vacation', start_date: '', end_date: '', reason: '' });
    },
  });

  const entries = entriesData?.data?.entries || [];
  const ptoRequests = ptoData?.data?.requests || [];

  const hoursByDay: Record<string, number> = {};
  entries.forEach((e) => {
    hoursByDay[e.date] = (hoursByDay[e.date] || 0) + e.hours;
  });
  const totalHours = entries.reduce((sum, e) => sum + e.hours, 0);
  const billableHours = entries.filter((e) => e.is_billable).reduce((sum, e) => sum + e.hours, 0);

  const tabs = [
    { id: 'timesheet' as const, label: 'Timesheet', icon: Clock },
    { id: 'pto' as const, label: 'PTO', icon: Calendar },
  ];

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Time Tracking</h1>

      <div className="flex gap-1 rounded-lg bg-gray-900 p-1">
        {tabs.map((t) => (
          <button
            key={t.id}
            onClick={() => setTab(t.id)}
            className={`flex flex-1 items-center justify-center gap-2 rounded-md px-4 py-2 text-sm font-medium ${
              tab === t.id ? 'bg-blue-600 text-white' : 'text-gray-400 hover:text-gray-200'
            }`}
          >
            <t.icon className="h-4 w-4" />
            {t.label}
          </button>
        ))}
      </div>

      {/* Timesheet Tab */}
      {tab === 'timesheet' && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <button onClick={() => setWeekOffset(weekOffset - 1)} className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-300 hover:bg-gray-700">
                &larr;
              </button>
              <span className="text-sm font-medium text-gray-300">
                {formatDate(weekStart)} &mdash; {formatDate(weekEnd)}
              </span>
              <button onClick={() => setWeekOffset(weekOffset + 1)} className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-300 hover:bg-gray-700">
                &rarr;
              </button>
              {weekOffset !== 0 && (
                <button onClick={() => setWeekOffset(0)} className="text-xs text-blue-400 hover:text-blue-300">
                  This Week
                </button>
              )}
            </div>
            <button
              onClick={() => setShowEntryModal(true)}
              className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
            >
              <Plus className="h-4 w-4" />
              Add Entry
            </button>
          </div>

          <div className="grid grid-cols-3 gap-4">
            <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
              <p className="text-sm text-gray-400">Total Hours</p>
              <p className="text-2xl font-bold text-gray-100">{totalHours.toFixed(1)}</p>
            </div>
            <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
              <p className="text-sm text-gray-400">Billable</p>
              <p className="text-2xl font-bold text-green-400">{billableHours.toFixed(1)}</p>
            </div>
            <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
              <p className="text-sm text-gray-400">Non-Billable</p>
              <p className="text-2xl font-bold text-gray-400">{(totalHours - billableHours).toFixed(1)}</p>
            </div>
          </div>

          <div className="rounded-xl border border-gray-800 bg-gray-900">
            <div className="grid grid-cols-7 border-b border-gray-800">
              {weekDays.map((d) => {
                const ds = toDateString(d);
                const isToday = toDateString(new Date()) === ds;
                return (
                  <div key={ds} className={`p-3 text-center text-sm ${isToday ? 'bg-blue-600/10 text-blue-400' : 'text-gray-400'}`}>
                    <p className="font-medium">{d.toLocaleDateString('en-US', { weekday: 'short' })}</p>
                    <p className="text-xs">{d.getDate()}</p>
                    <p className="mt-1 text-xs font-medium">{hoursByDay[ds]?.toFixed(1) || '0'}h</p>
                  </div>
                );
              })}
            </div>
            <div className="p-4">
              {entries.length === 0 ? (
                <p className="py-4 text-center text-sm text-gray-500">No time entries this week</p>
              ) : (
                <div className="space-y-2">
                  {entries.map((entry) => (
                    <div key={entry.id} className="flex items-center gap-3 rounded-lg border border-gray-800 bg-gray-800/50 p-3">
                      <div className="flex-1">
                        <p className="text-sm text-gray-200">{entry.description || 'No description'}</p>
                        <p className="text-xs text-gray-500">{entry.date} &middot; {entry.entry_type}</p>
                      </div>
                      <span className={`rounded px-2 py-0.5 text-xs ${entry.is_billable ? 'bg-green-500/20 text-green-400' : 'bg-gray-500/20 text-gray-400'}`}>
                        {entry.is_billable ? 'Billable' : 'Non-billable'}
                      </span>
                      <span className="text-sm font-medium text-gray-200">{entry.hours}h</span>
                      <button
                        onClick={() => deleteEntryMutation.mutate(entry.id)}
                        className="rounded p-1 text-gray-600 hover:bg-gray-700 hover:text-red-400"
                      >
                        <Trash2 className="h-4 w-4" />
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* PTO Tab */}
      {tab === 'pto' && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="grid grid-cols-3 gap-4">
              {['vacation', 'sick', 'personal'].map((type) => {
                const count = ptoRequests.filter((r) => r.pto_type === type && r.status === 'approved').reduce((s, r) => s + r.days, 0);
                return (
                  <div key={type} className="rounded-xl border border-gray-800 bg-gray-900 p-3">
                    <p className="text-xs capitalize text-gray-400">{type}</p>
                    <p className="text-lg font-bold text-gray-100">{count} days</p>
                  </div>
                );
              })}
            </div>
            <button
              onClick={() => setShowPTOModal(true)}
              className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
            >
              <Plus className="h-4 w-4" />
              Request PTO
            </button>
          </div>

          {ptoRequests.length === 0 ? (
            <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
              <Calendar className="mb-4 h-12 w-12 text-gray-600" />
              <p className="text-gray-400">No PTO requests</p>
            </div>
          ) : (
            <div className="space-y-3">
              {ptoRequests.map((req) => (
                <div key={req.id} className="rounded-xl border border-gray-800 bg-gray-900 p-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="font-medium text-gray-100 capitalize">{req.pto_type}</span>
                        <span className={`rounded-full px-2 py-0.5 text-xs ${ptoStatusColors[req.status] || ''}`}>
                          {req.status}
                        </span>
                      </div>
                      <p className="mt-1 text-sm text-gray-400">
                        {formatDate(req.start_date)} &mdash; {formatDate(req.end_date)} ({req.days} day{req.days !== 1 ? 's' : ''})
                      </p>
                      {req.reason && <p className="mt-1 text-sm text-gray-500">{req.reason}</p>}
                      {req.notes && <p className="mt-1 text-xs text-gray-600">Note: {req.notes}</p>}
                    </div>
                    <div className="flex items-center gap-2">
                      {req.status === 'approved' && <CheckCircle className="h-5 w-5 text-green-400" />}
                      {req.status === 'rejected' && <XCircle className="h-5 w-5 text-red-400" />}
                      {req.status === 'pending' && <Clock className="h-5 w-5 text-yellow-400" />}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Add Entry Modal */}
      {showEntryModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Add Time Entry</h2>
            <input
              type="date"
              value={newEntry.date}
              onChange={(e) => setNewEntry({ ...newEntry, date: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
            />
            <div className="mb-3 grid grid-cols-2 gap-3">
              <input
                type="number"
                placeholder="Hours"
                min="0.5"
                step="0.5"
                value={newEntry.hours}
                onChange={(e) => setNewEntry({ ...newEntry, hours: parseFloat(e.target.value) || 0 })}
                className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
              />
              <select
                value={newEntry.entry_type}
                onChange={(e) => setNewEntry({ ...newEntry, entry_type: e.target.value })}
                className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
              >
                <option value="work">Work</option>
                <option value="meeting">Meeting</option>
                <option value="training">Training</option>
                <option value="other">Other</option>
              </select>
            </div>
            <input
              type="text"
              placeholder="Description"
              value={newEntry.description}
              onChange={(e) => setNewEntry({ ...newEntry, description: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
            />
            <label className="mb-4 flex items-center gap-2 text-sm text-gray-400">
              <input
                type="checkbox"
                checked={newEntry.is_billable}
                onChange={(e) => setNewEntry({ ...newEntry, is_billable: e.target.checked })}
                className="rounded border-gray-700 bg-gray-800"
              />
              Billable
            </label>
            <div className="flex justify-end gap-3">
              <button onClick={() => setShowEntryModal(false)} className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200">Cancel</button>
              <button
                onClick={() => createEntryMutation.mutate(newEntry)}
                disabled={newEntry.hours <= 0}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                Add
              </button>
            </div>
          </div>
        </div>
      )}

      {/* PTO Request Modal */}
      {showPTOModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Request PTO</h2>
            <select
              value={newPTO.pto_type}
              onChange={(e) => setNewPTO({ ...newPTO, pto_type: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
            >
              <option value="vacation">Vacation</option>
              <option value="sick">Sick</option>
              <option value="personal">Personal</option>
              <option value="bereavement">Bereavement</option>
              <option value="jury_duty">Jury Duty</option>
            </select>
            <div className="mb-3 grid grid-cols-2 gap-3">
              <div>
                <label className="mb-1 block text-xs text-gray-400">Start Date</label>
                <input
                  type="date"
                  value={newPTO.start_date}
                  onChange={(e) => setNewPTO({ ...newPTO, start_date: e.target.value })}
                  className="w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
                />
              </div>
              <div>
                <label className="mb-1 block text-xs text-gray-400">End Date</label>
                <input
                  type="date"
                  value={newPTO.end_date}
                  onChange={(e) => setNewPTO({ ...newPTO, end_date: e.target.value })}
                  className="w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
                />
              </div>
            </div>
            <textarea
              placeholder="Reason (optional)"
              value={newPTO.reason}
              onChange={(e) => setNewPTO({ ...newPTO, reason: e.target.value })}
              className="mb-4 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
              rows={3}
            />
            <div className="flex justify-end gap-3">
              <button onClick={() => setShowPTOModal(false)} className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200">Cancel</button>
              <button
                onClick={() => requestPTOMutation.mutate(newPTO)}
                disabled={!newPTO.start_date || !newPTO.end_date}
                className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                <Send className="h-4 w-4" />
                Submit
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
