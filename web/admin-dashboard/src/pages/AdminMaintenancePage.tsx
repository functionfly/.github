import { adminApiClient } from '@/lib/api/adminClient';
import { API_ROUTES } from '@/lib/constants';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  AlertTriangle,
  Calendar,
  Clock,
  Edit,
  FileText,
  Plus,
  Power,
  PowerOff,
  Trash2,
} from 'lucide-react';
import { useState } from 'react';

// Types
interface MaintenanceWindow {
  id: string;
  name: string;
  description: string;
  scheduled_start: string;
  scheduled_end: string;
  status: 'scheduled' | 'active' | 'completed' | 'cancelled';
  affected_services: string[];
  created_at: string;
  created_by: string;
}

interface MaintenanceTemplate {
  id: string;
  name: string;
  description: string;
  default_duration_minutes: number;
  affected_services: string[];
}

interface MaintenanceStatus {
  enabled: boolean;
  message: string;
  scheduled_windows: MaintenanceWindow[];
  templates: MaintenanceTemplate[];
}

type RawMaintenanceStatus = {
  enabled?: boolean;
  message?: string | null;
  // backend currently returns many more fields (name, template, etc.)
};

/**
 * Admin Maintenance Page
 * Manages maintenance mode and scheduled maintenance windows
 */
export function AdminMaintenancePage() {
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState<'status' | 'schedule' | 'templates'>('status');

  const normalizeMaintenanceStatus = (raw: RawMaintenanceStatus): MaintenanceStatus => {
    return {
      enabled: Boolean(raw.enabled),
      message: typeof raw.message === 'string' ? raw.message : '',
      scheduled_windows: [],
      templates: [],
    };
  };

  // Fetch maintenance status
  const { data: status, isLoading: statusLoading } = useQuery<MaintenanceStatus>({
    queryKey: ['maintenance-status'],
    queryFn: async () => {
      // Backend returns a raw JSON object (not the { data, success } envelope).
      // React Query requires queryFn to return a non-undefined value.
      const rawUnknown = await adminApiClient.get<unknown>(API_ROUTES.ADMIN_MAINTENANCE);
      const raw =
        rawUnknown && typeof rawUnknown === 'object'
          ? (rawUnknown as RawMaintenanceStatus)
          : ({} as RawMaintenanceStatus);
      return normalizeMaintenanceStatus(raw);
    },
  });

  // Enable maintenance mutation
  const enableMaintenanceMutation = useMutation({
    mutationFn: async (data: { message: string; duration_minutes: number }) => {
      return adminApiClient.post(API_ROUTES.ADMIN_MAINTENANCE, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['maintenance-status'] });
    },
  });

  // Disable maintenance mutation
  const disableMaintenanceMutation = useMutation({
    mutationFn: async () => {
      return adminApiClient.delete(API_ROUTES.ADMIN_MAINTENANCE);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['maintenance-status'] });
    },
  });

  const handleEnableMaintenance = () => {
    enableMaintenanceMutation.mutate({
      message: 'System maintenance in progress',
      duration_minutes: 60,
    });
  };

  const handleDisableMaintenance = () => {
    disableMaintenanceMutation.mutate();
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">Maintenance Mode</h1>
          <p className="text-gray-600 dark:text-gray-400 mt-1">
            Manage system maintenance windows and scheduled maintenance
          </p>
        </div>
      </div>

      {/* Status Card */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div
              className={`p-3 rounded-full ${status?.enabled ? 'bg-yellow-100 dark:bg-yellow-900/30' : 'bg-green-100 dark:bg-green-900/30'}`}
            >
              <AlertTriangle
                className={`w-6 h-6 ${status?.enabled ? 'text-yellow-600 dark:text-yellow-400' : 'text-green-600 dark:text-green-400'}`}
              />
            </div>
            <div>
              <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">
                {status?.enabled ? 'Maintenance Mode Active' : 'System Operational'}
              </h2>
              {status?.enabled && status?.message && (
                <p className="text-gray-600 dark:text-gray-400">{status.message}</p>
              )}
            </div>
          </div>
          {status?.enabled ? (
            <button
              onClick={handleDisableMaintenance}
              disabled={disableMaintenanceMutation.isPending}
              className="flex items-center gap-2 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 disabled:opacity-50"
            >
              <PowerOff className="w-4 h-4" />
              Disable Maintenance
            </button>
          ) : (
            <button
              onClick={handleEnableMaintenance}
              disabled={enableMaintenanceMutation.isPending}
              className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
            >
              <Power className="w-4 h-4" />
              Enable Maintenance
            </button>
          )}
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200 dark:border-gray-700">
        <nav className="-mb-px flex space-x-8">
          <button
            onClick={() => setActiveTab('status')}
            className={`py-4 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'status'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-200 dark:hover:border-gray-600'
            }`}
          >
            <Clock className="w-4 h-4 inline mr-2" />
            Scheduled Windows
          </button>
          <button
            onClick={() => setActiveTab('schedule')}
            className={`py-4 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'schedule'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-200 dark:hover:border-gray-600'
            }`}
          >
            <Calendar className="w-4 h-4 inline mr-2" />
            Schedule New
          </button>
          <button
            onClick={() => setActiveTab('templates')}
            className={`py-4 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'templates'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-200 dark:hover:border-gray-600'
            }`}
          >
            <FileText className="w-4 h-4 inline mr-2" />
            Templates
          </button>
        </nav>
      </div>

      {/* Tab Content */}
      {activeTab === 'status' && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
          {statusLoading ? (
            <div className="p-8 text-center text-gray-500 dark:text-gray-400">Loading...</div>
          ) : status?.scheduled_windows?.length === 0 ? (
            <div className="p-8 text-center text-gray-500 dark:text-gray-400">No scheduled maintenance windows</div>
          ) : (
            <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
              <thead className="bg-gray-50 dark:bg-gray-800">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Name
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Start Time
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    End Time
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Affected Services
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                {status?.scheduled_windows?.map((window) => (
                  <tr key={window.id}>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm font-medium text-gray-900 dark:text-gray-100">{window.name}</div>
                      <div className="text-sm text-gray-500 dark:text-gray-400">{window.description}</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span
                        className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                          window.status === 'active'
                            ? 'bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-400'
                            : window.status === 'scheduled'
                              ? 'bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-400'
                              : window.status === 'completed'
                                ? 'bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-400'
                                : 'bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-400'
                        }`}
                      >
                        {window.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                      {new Date(window.scheduled_start).toLocaleString()}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                      {new Date(window.scheduled_end).toLocaleString()}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                      {window.affected_services.join(', ')}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <button className="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300 mr-4">
                        <Edit className="w-4 h-4" />
                      </button>
                      <button className="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300">
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {activeTab === 'schedule' && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <form className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Window Name</label>
              <input
                type="text"
                className="mt-1 block w-full rounded-md border-gray-300 dark:border-gray-600 shadow-sm focus:border-blue-500 focus:ring-blue-500 dark:bg-gray-700"
                placeholder="e.g., Database Maintenance"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Description</label>
              <textarea
                rows={3}
                className="mt-1 block w-full rounded-md border-gray-300 dark:border-gray-600 shadow-sm focus:border-blue-500 focus:ring-blue-500 dark:bg-gray-700"
                placeholder="Describe the maintenance work..."
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Start Time</label>
                <input
                  type="datetime-local"
                  className="mt-1 block w-full rounded-md border-gray-300 dark:border-gray-600 shadow-sm focus:border-blue-500 focus:ring-blue-500 dark:bg-gray-700"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">End Time</label>
                <input
                  type="datetime-local"
                  className="mt-1 block w-full rounded-md border-gray-300 dark:border-gray-600 shadow-sm focus:border-blue-500 focus:ring-blue-500 dark:bg-gray-700"
                />
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Affected Services</label>
              <input
                type="text"
                className="mt-1 block w-full rounded-md border-gray-300 dark:border-gray-600 shadow-sm focus:border-blue-500 focus:ring-blue-500 dark:bg-gray-700"
                placeholder="e.g., API, Database, Cache (comma-separated)"
              />
            </div>
            <div className="flex justify-end">
              <button
                type="submit"
                className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
              >
                <Plus className="w-4 h-4" />
                Schedule Maintenance
              </button>
            </div>
          </form>
        </div>
      )}

      {activeTab === 'templates' && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
          {status?.templates?.length === 0 ? (
            <div className="p-8 text-center text-gray-500 dark:text-gray-400">No maintenance templates configured</div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 p-6">
              {status?.templates?.map((template) => (
                <div key={template.id} className="border rounded-lg p-4 border-gray-200 dark:border-gray-700">
                  <h3 className="font-semibold text-lg text-gray-900 dark:text-gray-100">{template.name}</h3>
                  <p className="text-gray-600 dark:text-gray-400 text-sm mt-1">{template.description}</p>
                  <div className="mt-3 flex items-center justify-between text-sm text-gray-500 dark:text-gray-400">
                    <span>Duration: {template.default_duration_minutes} min</span>
                    <span>{template.affected_services.length} services</span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default AdminMaintenancePage;
