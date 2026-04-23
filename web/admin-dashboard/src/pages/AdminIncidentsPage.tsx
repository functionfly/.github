/**
 * Admin Incidents Page
 * Manage and track service incidents via /v1/admin/incidents
 */

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { LoadingScreen } from '@/components/common/LoadingScreen';

interface Incident {
  id: string;
  title: string;
  severity: 'critical' | 'high' | 'medium' | 'low';
  status: 'resolved' | 'investigating' | 'monitoring';
  description: string;
  created_at: string;
  resolved_at?: string;
  updated_at: string;
  affected_components?: string[];
}

const severityColors: Record<string, string> = {
  critical: 'bg-red-100 text-red-800',
  high: 'bg-orange-100 text-orange-800',
  medium: 'bg-yellow-100 text-yellow-800',
  low: 'bg-gray-100 text-gray-800',
};

const statusColors: Record<string, string> = {
  investigating: 'bg-red-100 text-red-800',
  identified: 'bg-amber-100 text-amber-800',
  monitoring: 'bg-blue-100 text-blue-800',
  resolved: 'bg-green-100 text-green-800',
};

export function AdminIncidentsPage() {
  const queryClient = useQueryClient();
  const [statusFilter, setStatusFilter] = useState<string>('all');

  const { data: response, isLoading } = useQuery({
    queryKey: ['admin-incidents', statusFilter],
    queryFn: async () => {
      const params = statusFilter !== 'all' ? `?status=${statusFilter}` : '';
      return await adminApiClient.get<{ incidents: Incident[] }>(`/incidents${params}`);
    },
  });

  const resolveMutation = useMutation({
    mutationFn: (incidentId: string) =>
      adminApiClient.post<Incident>(`/incidents/${incidentId}/resolve`, {}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-incidents'] });
    },
  });

  const incidents: Incident[] = response?.data?.incidents ?? [];
  const activeCount = incidents.filter((i) => i.status !== 'resolved').length;

  if (isLoading) {
    return <LoadingScreen />;
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">Incident Management</h1>
          <p className="mt-2 text-gray-600 dark:text-gray-400">Manage and track service incidents.</p>
        </div>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div className="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-4">
          <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Total</p>
          <p className="mt-1 text-2xl font-bold text-gray-900 dark:text-gray-100">{incidents.length}</p>
        </div>
        <div className="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-4">
          <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Active</p>
          <p className="mt-1 text-2xl font-bold text-red-600 dark:text-red-400">{activeCount}</p>
        </div>
        <div className="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-4">
          <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Critical</p>
          <p className="mt-1 text-2xl font-bold text-red-700 dark:text-red-500">
            {incidents.filter((i) => i.severity === 'critical').length}
          </p>
        </div>
        <div className="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-4">
          <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Resolved</p>
          <p className="mt-1 text-2xl font-bold text-green-600 dark:text-green-400">
            {incidents.filter((i) => i.status === 'resolved').length}
          </p>
        </div>
      </div>

      <div className="flex gap-2">
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-sm text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
        >
          <option value="all">All statuses</option>
          <option value="investigating">Investigating</option>
          <option value="monitoring">Monitoring</option>
          <option value="resolved">Resolved</option>
        </select>
      </div>

      <div className="overflow-hidden rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead className="bg-gray-50 dark:bg-gray-800">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500 dark:text-gray-400">
                Title
              </th>
              <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500 dark:text-gray-400">
                Status
              </th>
              <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500 dark:text-gray-400">
                Severity
              </th>
              <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500 dark:text-gray-400">
                Created
              </th>
              <th className="px-4 py-3 text-right text-xs font-medium uppercase text-gray-500 dark:text-gray-400">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 dark:divide-gray-700 bg-white dark:bg-gray-800">
            {incidents.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">
                  No incidents found.
                </td>
              </tr>
            ) : (
              incidents.map((incident) => (
                <tr key={incident.id}>
                  <td className="px-4 py-3">
                    <p className="font-medium text-gray-900 dark:text-gray-100">{incident.title}</p>
                    <p className="text-xs text-gray-500 dark:text-gray-400 line-clamp-1">{incident.description}</p>
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`inline-flex rounded px-2 py-1 text-xs font-medium ${
                        statusColors[incident.status] ?? 'bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-400'
                      }`}
                    >
                      {incident.status}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`inline-flex rounded px-2 py-1 text-xs font-medium ${
                        severityColors[incident.severity] ?? 'bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-400'
                      }`}
                    >
                      {incident.severity}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">
                    {new Date(incident.created_at).toLocaleString()}
                  </td>
                  <td className="px-4 py-3 text-right">
                    {incident.status !== 'resolved' && (
                      <button
                        type="button"
                        onClick={() => resolveMutation.mutate(incident.id)}
                        disabled={resolveMutation.isPending}
                        className="text-sm font-medium text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 disabled:opacity-50"
                      >
                        Resolve
                      </button>
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
