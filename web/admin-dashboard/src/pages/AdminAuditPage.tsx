/**
 * Admin Audit Page
 * View audit log of all admin actions
 */

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { Search, Download, Eye, AlertTriangle, CheckCircle } from 'lucide-react';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import type { AuditEvent } from '@/types';

const ACTION_ICONS: Record<string, string> = {
  create: '✨',
  update: '✏️',
  delete: '🗑️',
  suspend: '⛔',
  login: '🔓',
  logout: '🔐',
  default: '📝',
};

export function AdminAuditPage() {
  const [searchTerm, setSearchTerm] = useState('');
  const [actionFilter, setActionFilter] = useState<string>('all');
  const [successFilter, setSuccessFilter] = useState<string>('all');
  const [selectedEvent, setSelectedEvent] = useState<AuditEvent | null>(null);
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);

  // Fetch audit events (API returns { events, limit, offset, filters })
  const { data: eventsResponse, isLoading, isError, refetch } = useQuery({
    queryKey: ['admin-audit-events'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<AuditEvent[]>('/audit-events?limit=100');
      } catch {
        return { events: [], limit: 100, offset: 0, filters: {} };
      }
    },
    staleTime: 1000 * 30, // 30 seconds
  });

  const raw = eventsResponse as { events?: AuditEvent[]; data?: AuditEvent[] } | undefined;
  const events = raw?.events ?? raw?.data ?? [];

  const filteredEvents = events.filter((event) => {
    const action = event.action ?? '';
    const resourceType = event.resource_type ?? '';
    const matchesSearch = searchTerm
      ? action.toLowerCase().includes(searchTerm.toLowerCase()) ||
        event.actor_email?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        resourceType.toLowerCase().includes(searchTerm.toLowerCase())
      : true;

    const matchesAction = actionFilter === 'all' || action.includes(actionFilter);
    const matchesSuccess =
      successFilter === 'all' || (successFilter === 'success' ? event.success : !event.success);

    return matchesSearch && matchesAction && matchesSuccess;
  });

  const getActionIcon = (action: string) => {
    const normalized = action.toLowerCase();
    if (normalized.includes('create')) return ACTION_ICONS.create;
    if (normalized.includes('update')) return ACTION_ICONS.update;
    if (normalized.includes('delete')) return ACTION_ICONS.delete;
    if (normalized.includes('suspend')) return ACTION_ICONS.suspend;
    if (normalized.includes('login')) return ACTION_ICONS.login;
    if (normalized.includes('logout')) return ACTION_ICONS.logout;
    return ACTION_ICONS.default;
  };

  const handleDownloadCsv = () => {
    const csv = [
      ['Timestamp', 'Actor', 'Action', 'Resource', 'Success'],
      ...filteredEvents.map((e) => [
        new Date(e.timestamp ?? 0).toISOString(),
        e.actor_email || 'System',
        e.action ?? '',
        `${e.resource_type ?? ''}${e.resource_id ? `:${e.resource_id}` : ''}`,
        e.success ? 'Yes' : 'No',
      ]),
    ]
      .map((row) => row.map((cell) => `"${cell}"`).join(','))
      .join('\n');

    const blob = new Blob([csv], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `audit-events-${new Date().toISOString().split('T')[0]}.csv`;
    a.click();
  };

  if (isLoading) {
    return <LoadingScreen />;
  }

  if (isError) {
    return (
      <div className="p-8 bg-red-50 border border-red-200 rounded-lg">
        <h3 className="font-semibold text-red-900">Error loading audit events</h3>
        <p className="text-red-700 mt-2">Failed to fetch audit data.</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">Audit Log</h1>
          <p className="mt-2 text-gray-600 dark:text-gray-400">View all admin actions and system events</p>
        </div>

        <button
          onClick={handleDownloadCsv}
          className="flex items-center gap-2 px-4 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700 transition-colors"
        >
          <Download className="w-5 h-5" />
          Export CSV
        </button>
      </div>

      {/* Filters */}
      <div className="flex gap-4">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-3 w-5 h-5 text-gray-400" />
          <input
            type="text"
            placeholder="Search by action, user, or resource..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          />
        </div>

        <select
          value={actionFilter}
          onChange={(e) => setActionFilter(e.target.value)}
          className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
        >
          <option value="all">All Actions</option>
          <option value="create">Create</option>
          <option value="update">Update</option>
          <option value="delete">Delete</option>
          <option value="login">Login</option>
        </select>

        <select
          value={successFilter}
          onChange={(e) => setSuccessFilter(e.target.value)}
          className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
        >
          <option value="all">All Results</option>
          <option value="success">Success</option>
          <option value="failure">Failure</option>
        </select>

        <button
          onClick={() => refetch()}
          className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors text-gray-900 dark:text-gray-100"
        >
          Refresh
        </button>
      </div>

      {/* Events List */}
      <div className="space-y-2">
        {filteredEvents.length === 0 ? (
          <div className="p-8 text-center text-gray-500 dark:text-gray-400 bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700">
            No events found
          </div>
        ) : (
          filteredEvents.map((event, idx) => (
            <div
              key={idx}
              onClick={() => {
                setSelectedEvent(event);
                setIsDetailsOpen(true);
              }}
              className="bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700 p-4 hover:shadow-md transition-shadow cursor-pointer"
            >
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-3">
                    <span className="text-xl">{getActionIcon(event.action ?? '')}</span>
                    <div>
                      <p className="font-medium text-gray-900 dark:text-gray-100">{event.action ?? '—'}</p>
                      <p className="text-sm text-gray-600 dark:text-gray-400">
                        {event.actor_email || 'System'} •{' '}
                        {event.resource_type ?? '—'}
                        {event.resource_id ? `:${event.resource_id}` : ''}
                      </p>
                    </div>
                  </div>
                </div>

                <div className="flex items-center gap-3">
                  {event.success ? (
                    <CheckCircle className="w-5 h-5 text-green-600" />
                  ) : (
                    <AlertTriangle className="w-5 h-5 text-red-600" />
                  )}

                  <div className="text-right">
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      {new Date(event.timestamp ?? 0).toLocaleDateString()}
                    </p>
                    <p className="text-xs text-gray-500 dark:text-gray-500">
                      {new Date(event.timestamp ?? 0).toLocaleTimeString()}
                    </p>
                  </div>

                  <Eye className="w-5 h-5 text-gray-400" />
                </div>
              </div>
            </div>
          ))
        )}
      </div>

      {/* Details Modal */}
      {isDetailsOpen && selectedEvent && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-gray-900 rounded-lg shadow-xl p-6 max-w-2xl w-full max-h-96 overflow-y-auto">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100">Event Details</h2>
              <button
                onClick={() => setIsDetailsOpen(false)}
                className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
              >
                ✕
              </button>
            </div>

            <div className="space-y-4">
              <div>
                <h3 className="font-semibold text-gray-900 dark:text-gray-100">Action</h3>
                <p className="text-gray-600 dark:text-gray-400">{selectedEvent.action}</p>
              </div>

              <div>
                <h3 className="font-semibold text-gray-900 dark:text-gray-100">Actor</h3>
                <p className="text-gray-600 dark:text-gray-400">{selectedEvent.actor_email || 'System'}</p>
              </div>

              <div>
                <h3 className="font-semibold text-gray-900 dark:text-gray-100">Resource</h3>
                <p className="text-gray-600 dark:text-gray-400">
                  {selectedEvent.resource_type}
                  {selectedEvent.resource_id ? `:${selectedEvent.resource_id}` : ''}
                </p>
              </div>

              <div>
                <h3 className="font-semibold text-gray-900 dark:text-gray-100">Timestamp</h3>
                <p className="text-gray-600 dark:text-gray-400">
                  {new Date(selectedEvent.timestamp).toLocaleString()}
                </p>
              </div>

              <div>
                <h3 className="font-semibold text-gray-900 dark:text-gray-100">Result</h3>
                <p className={selectedEvent.success ? 'text-green-600 dark:text-green-400 font-medium' : 'text-red-600 dark:text-red-400 font-medium'}>
                  {selectedEvent.success ? 'Success' : 'Failed'}
                </p>
              </div>

              {selectedEvent.ip_address && (
                <div>
                  <h3 className="font-semibold text-gray-900 dark:text-gray-100">IP Address</h3>
                  <p className="text-gray-600 dark:text-gray-400">{selectedEvent.ip_address}</p>
                </div>
              )}

              {selectedEvent.before_state && (
                <div>
                  <h3 className="font-semibold text-gray-900 dark:text-gray-100">Before</h3>
                  <pre className="bg-gray-50 dark:bg-gray-800 p-3 rounded text-xs overflow-auto text-gray-900 dark:text-gray-100">
                    {JSON.stringify(selectedEvent.before_state, null, 2)}
                  </pre>
                </div>
              )}

              {selectedEvent.after_state && (
                <div>
                  <h3 className="font-semibold text-gray-900 dark:text-gray-100">After</h3>
                  <pre className="bg-gray-50 dark:bg-gray-800 p-3 rounded text-xs overflow-auto text-gray-900 dark:text-gray-100">
                    {JSON.stringify(selectedEvent.after_state, null, 2)}
                  </pre>
                </div>
              )}
            </div>

            <button
              onClick={() => setIsDetailsOpen(false)}
              className="w-full mt-6 px-4 py-2 bg-gray-200 dark:bg-gray-700 text-gray-800 dark:text-gray-200 rounded-lg hover:bg-gray-300 dark:hover:bg-gray-600"
            >
              Close
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
