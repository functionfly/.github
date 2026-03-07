/**
 * Admin Auth Audit Page
 * View authentication audit logs with filtering capabilities
 */

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { 
  Search, 
  Download, 
  Eye, 
  AlertTriangle, 
  CheckCircle, 
  Filter,
  Calendar,
  User,
  Building2,
  Globe,
  RefreshCw
} from 'lucide-react';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { AUDIT_EVENT_TYPES } from '@/lib/constants';
import type { AuditEvent, AuditLogFilters } from '@/types';

const ACTION_ICONS: Record<string, string> = {
  login: '🔓',
  logout: '🔐',
  mfa_enable: '🛡️',
  mfa_disable: '⚠️',
  sso_login: '🔗',
  password_change: '🔑',
  password_reset: '🔄',
  user_create: '👤',
  user_update: '✏️',
  user_delete: '🗑️',
  tenant_create: '🏢',
  tenant_update: '🏗️',
  tenant_suspend: '⛔',
  api_key_create: '🗝️',
  api_key_revoke: '❌',
  session_revoke: '🚫',
  ip_allowlist_create: '📋',
  ip_allowlist_update: '📝',
  siem_config_create: '📊',
  passkey_register: '🔐',
  passkey_delete: '🗑️',
  default: '📝',
};

export function AdminAuthAuditPage() {
  const [searchTerm, setSearchTerm] = useState('');
  const [eventTypeFilter, setEventTypeFilter] = useState<string>('all');
  const [successFilter, setSuccessFilter] = useState<string>('all');
  const [selectedEvent, setSelectedEvent] = useState<AuditEvent | null>(null);
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);
  const [showFilters, setShowFilters] = useState(false);
  
  // Date range filters
  const [startDate, setStartDate] = useState<string>('');
  const [endDate, setEndDate] = useState<string>('');

  // Fetch auth audit events
  const { data: eventsResponse, isLoading, isError, refetch, isFetching } = useQuery({
    queryKey: ['admin-auth-audit-events', eventTypeFilter, successFilter, startDate, endDate],
    queryFn: async () => {
      try {
        const params = new URLSearchParams();
        if (eventTypeFilter !== 'all') params.append('event_type', eventTypeFilter);
        if (successFilter !== 'all') params.append('success', successFilter);
        if (startDate) params.append('start_date', startDate);
        if (endDate) params.append('end_date', endDate);
        
        return await adminApiClient.get<AuditEvent[]>(`/admin/auth-audit?${params.toString()}`);
      } catch {
        return { data: [], success: false };
      }
    },
    staleTime: 1000 * 30, // 30 seconds
  });

  const events = eventsResponse?.data || [];

  const filteredEvents = events.filter((event) => {
    const matchesSearch = searchTerm
      ? event.event_type?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        event.user_email?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        event.tenant_name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        event.ip_address?.includes(searchTerm) ||
        event.action?.toLowerCase().includes(searchTerm.toLowerCase())
      : true;

    return matchesSearch;
  });

  const getActionIcon = (eventType: string) => {
    const normalized = eventType?.toLowerCase() || '';
    if (ACTION_ICONS[normalized]) return ACTION_ICONS[normalized];
    if (normalized.includes('login')) return ACTION_ICONS.login;
    if (normalized.includes('logout')) return ACTION_ICONS.logout;
    if (normalized.includes('mfa')) return ACTION_ICONS.mfa_enable;
    if (normalized.includes('password')) return ACTION_ICONS.password_change;
    if (normalized.includes('user')) return ACTION_ICONS.user_create;
    if (normalized.includes('tenant')) return ACTION_ICONS.tenant_create;
    return ACTION_ICONS.default;
  };

  const handleDownloadCsv = () => {
    const csv = [
      ['Timestamp', 'Event Type', 'User Email', 'Tenant', 'IP Address', 'Success', 'Action'],
      ...filteredEvents.map((e) => [
        new Date(e.created_at || e.timestamp).toISOString(),
        e.event_type || '',
        e.user_email || 'System',
        e.tenant_name || '-',
        e.ip_address || '-',
        e.success ? 'Yes' : 'No',
        e.action || '',
      ]),
    ]
      .map((row) => row.map((cell) => `"${cell}"`).join(','))
      .join('\n');

    const blob = new Blob([csv], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `auth-audit-${new Date().toISOString().split('T')[0]}.csv`;
    a.click();
  };

  const clearFilters = () => {
    setEventTypeFilter('all');
    setSuccessFilter('all');
    setStartDate('');
    setEndDate('');
    setSearchTerm('');
  };

  if (isLoading) {
    return <LoadingScreen />;
  }

  if (isError) {
    return (
      <div className="p-8 bg-red-50 border border-red-200 rounded-lg dark:bg-red-900/20 dark:border-red-800">
        <h3 className="font-semibold text-red-900 dark:text-red-200">Error loading audit events</h3>
        <p className="text-red-700 dark:text-red-300 mt-2">Failed to fetch audit data.</p>
        <button 
          onClick={() => refetch()}
          className="mt-4 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Auth Audit Log</h1>
          <p className="mt-2 text-gray-600 dark:text-gray-400">
            View authentication events and security audit logs
          </p>
        </div>

        <div className="flex gap-2">
          <button
            onClick={() => refetch()}
            disabled={isFetching}
            className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 dark:border-gray-600 dark:hover:bg-gray-800 transition-colors"
          >
            <RefreshCw className={`w-5 h-5 ${isFetching ? 'animate-spin' : ''}`} />
            Refresh
          </button>
          <button
            onClick={handleDownloadCsv}
            className="flex items-center gap-2 px-4 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700 transition-colors"
          >
            <Download className="w-5 h-5" />
            Export CSV
          </button>
        </div>
      </div>

      {/* Search and Filters */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
        <div className="flex gap-4">
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-3 w-5 h-5 text-gray-400" />
            <input
              type="text"
              placeholder="Search by user, tenant, IP, or event type..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent dark:bg-gray-700 dark:text-white"
            />
          </div>

          <button
            onClick={() => setShowFilters(!showFilters)}
            className={`flex items-center gap-2 px-4 py-2 border rounded-lg transition-colors ${
              showFilters 
                ? 'bg-blue-50 border-blue-300 text-blue-700 dark:bg-blue-900/20 dark:border-blue-700 dark:text-blue-300' 
                : 'border-gray-300 hover:bg-gray-50 dark:border-gray-600 dark:hover:bg-gray-700'
            }`}
          >
            <Filter className="w-5 h-5" />
            Filters
          </button>
        </div>

        {/* Expanded Filters */}
        {showFilters && (
          <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              {/* Event Type Filter */}
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Event Type
                </label>
                <select
                  value={eventTypeFilter}
                  onChange={(e) => setEventTypeFilter(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                >
                  {AUDIT_EVENT_TYPES.map((type) => (
                    <option key={type.value} value={type.value}>
                      {type.label}
                    </option>
                  ))}
                </select>
              </div>

              {/* Success Filter */}
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Result
                </label>
                <select
                  value={successFilter}
                  onChange={(e) => setSuccessFilter(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                >
                  <option value="all">All Results</option>
                  <option value="success">Success</option>
                  <option value="failure">Failure</option>
                </select>
              </div>

              {/* Start Date */}
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  <Calendar className="w-4 h-4 inline mr-1" />
                  Start Date
                </label>
                <input
                  type="date"
                  value={startDate}
                  onChange={(e) => setStartDate(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                />
              </div>

              {/* End Date */}
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  <Calendar className="w-4 h-4 inline mr-1" />
                  End Date
                </label>
                <input
                  type="date"
                  value={endDate}
                  onChange={(e) => setEndDate(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                />
              </div>
            </div>

            <div className="mt-4 flex justify-end">
              <button
                onClick={clearFilters}
                className="text-sm text-blue-600 hover:text-blue-700 dark:text-blue-400"
              >
                Clear all filters
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Events Table */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900">
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">Event</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">User</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">Tenant</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">IP Address</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">Result</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">Timestamp</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">Actions</th>
            </tr>
          </thead>
          <tbody>
            {filteredEvents.length === 0 ? (
              <tr>
                <td colSpan={7} className="px-6 py-8 text-center text-gray-500 dark:text-gray-400">
                  No events found matching your criteria
                </td>
              </tr>
            ) : (
              filteredEvents.map((event, idx) => (
                <tr 
                  key={event.id || idx} 
                  className="border-b border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700/50"
                >
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2">
                      <span className="text-xl">{getActionIcon(event.event_type)}</span>
                      <div>
                        <p className="font-medium text-gray-900 dark:text-white">
                          {event.event_type || event.action}
                        </p>
                        {event.action && event.event_type && event.action !== event.event_type && (
                          <p className="text-xs text-gray-500 dark:text-gray-400">{event.action}</p>
                        )}
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2">
                      <User className="w-4 h-4 text-gray-400" />
                      <span className="text-sm text-gray-600 dark:text-gray-300">
                        {event.user_email || event.actor_email || 'System'}
                      </span>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2">
                      <Building2 className="w-4 h-4 text-gray-400" />
                      <span className="text-sm text-gray-600 dark:text-gray-300">
                        {event.tenant_name || '-'}
                      </span>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2">
                      <Globe className="w-4 h-4 text-gray-400" />
                      <span className="text-sm text-gray-600 dark:text-gray-300 font-mono">
                        {event.ip_address || '-'}
                      </span>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    {event.success ? (
                      <span className="inline-flex items-center gap-1 px-2 py-1 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 rounded text-sm">
                        <CheckCircle className="w-4 h-4" />
                        Success
                      </span>
                    ) : (
                      <span className="inline-flex items-center gap-1 px-2 py-1 bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400 rounded text-sm">
                        <AlertTriangle className="w-4 h-4" />
                        Failed
                      </span>
                    )}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-600 dark:text-gray-300">
                    <div>
                      <p>{new Date(event.created_at || event.timestamp).toLocaleDateString()}</p>
                      <p className="text-xs text-gray-500 dark:text-gray-400">
                        {new Date(event.created_at || event.timestamp).toLocaleTimeString()}
                      </p>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <button
                      onClick={() => {
                        setSelectedEvent(event);
                        setIsDetailsOpen(true);
                      }}
                      className="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700"
                    >
                      <Eye className="w-5 h-5" />
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Summary Stats */}
      <div className="grid grid-cols-4 gap-4">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
          <p className="text-gray-600 dark:text-gray-400 text-sm">Total Events</p>
          <p className="text-2xl font-bold text-gray-900 dark:text-white">{filteredEvents.length}</p>
        </div>
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
          <p className="text-gray-600 dark:text-gray-400 text-sm">Successful</p>
          <p className="text-2xl font-bold text-green-600 dark:text-green-400">
            {filteredEvents.filter((e) => e.success).length}
          </p>
        </div>
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
          <p className="text-gray-600 dark:text-gray-400 text-sm">Failed</p>
          <p className="text-2xl font-bold text-red-600 dark:text-red-400">
            {filteredEvents.filter((e) => !e.success).length}
          </p>
        </div>
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
          <p className="text-gray-600 dark:text-gray-400 text-sm">Unique Users</p>
          <p className="text-2xl font-bold text-blue-600 dark:text-blue-400">
            {new Set(filteredEvents.map((e) => e.user_email || e.actor_email).filter(Boolean)).size}
          </p>
        </div>
      </div>

      {/* Details Modal */}
      {isDetailsOpen && selectedEvent && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-2xl font-bold text-gray-900 dark:text-white">Event Details</h2>
              <button
                onClick={() => setIsDetailsOpen(false)}
                className="text-gray-500 hover:text-gray-700 dark:hover:text-gray-300"
              >
                ✕
              </button>
            </div>

            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <h3 className="font-semibold text-gray-900 dark:text-white">Event Type</h3>
                  <p className="text-gray-600 dark:text-gray-300">{selectedEvent.event_type}</p>
                </div>

                <div>
                  <h3 className="font-semibold text-gray-900 dark:text-white">Action</h3>
                  <p className="text-gray-600 dark:text-gray-300">{selectedEvent.action}</p>
                </div>

                <div>
                  <h3 className="font-semibold text-gray-900 dark:text-white">User Email</h3>
                  <p className="text-gray-600 dark:text-gray-300">
                    {selectedEvent.user_email || selectedEvent.actor_email || 'System'}
                  </p>
                </div>

                <div>
                  <h3 className="font-semibold text-gray-900 dark:text-white">Tenant</h3>
                  <p className="text-gray-600 dark:text-gray-300">
                    {selectedEvent.tenant_name || '-'}
                  </p>
                </div>

                <div>
                  <h3 className="font-semibold text-gray-900 dark:text-white">IP Address</h3>
                  <p className="text-gray-600 dark:text-gray-300 font-mono">
                    {selectedEvent.ip_address || '-'}
                  </p>
                </div>

                <div>
                  <h3 className="font-semibold text-gray-900 dark:text-white">Timestamp</h3>
                  <p className="text-gray-600 dark:text-gray-300">
                    {new Date(selectedEvent.created_at || selectedEvent.timestamp).toLocaleString()}
                  </p>
                </div>

                <div>
                  <h3 className="font-semibold text-gray-900 dark:text-white">Result</h3>
                  <p className={selectedEvent.success ? 'text-green-600 font-medium' : 'text-red-600 font-medium'}>
                    {selectedEvent.success ? 'Success' : 'Failed'}
                  </p>
                </div>

                <div>
                  <h3 className="font-semibold text-gray-900 dark:text-white">Resource</h3>
                  <p className="text-gray-600 dark:text-gray-300">
                    {selectedEvent.resource_type}
                    {selectedEvent.resource_id ? `:${selectedEvent.resource_id}` : ''}
                  </p>
                </div>
              </div>

              {selectedEvent.before_state && (
                <div>
                  <h3 className="font-semibold text-gray-900 dark:text-white mb-2">Before State</h3>
                  <pre className="bg-gray-50 dark:bg-gray-900 p-3 rounded text-xs overflow-auto border border-gray-200 dark:border-gray-700">
                    {JSON.stringify(selectedEvent.before_state, null, 2)}
                  </pre>
                </div>
              )}

              {selectedEvent.after_state && (
                <div>
                  <h3 className="font-semibold text-gray-900 dark:text-white mb-2">After State</h3>
                  <pre className="bg-gray-50 dark:bg-gray-900 p-3 rounded text-xs overflow-auto border border-gray-200 dark:border-gray-700">
                    {JSON.stringify(selectedEvent.after_state, null, 2)}
                  </pre>
                </div>
              )}

              {selectedEvent.metadata && Object.keys(selectedEvent.metadata).length > 0 && (
                <div>
                  <h3 className="font-semibold text-gray-900 dark:text-white mb-2">Metadata</h3>
                  <pre className="bg-gray-50 dark:bg-gray-900 p-3 rounded text-xs overflow-auto border border-gray-200 dark:border-gray-700">
                    {JSON.stringify(selectedEvent.metadata, null, 2)}
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

export default AdminAuthAuditPage;
