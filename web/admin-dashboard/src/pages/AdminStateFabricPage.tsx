import { useState, useEffect } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { LayoutDashboard, Settings } from 'lucide-react';

interface StateFabricItem {
  id: string;
  name?: string;
  tenant_id?: string;
  tenantId?: string;
  status?: 'running' | 'suspended' | string;
}

// API returns totalFabrics, activeFabrics (no suspended count; we derive it)
interface StateFabricStats {
  total?: number;
  totalFabrics?: number;
  active?: number;
  activeFabrics?: number;
  suspended?: number;
}

export interface StateFabricPlatformSettings {
  maxFabricsPerTenant?: number;
  defaultSnapshotRetentionDays?: number;
  allowPublicPipelines?: boolean;
  maintenanceMode?: boolean;
}

type TabId = 'overview' | 'settings';

export function AdminStateFabricPage() {
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState<TabId>('overview');

  // API returns { totalFabrics, activeFabrics, totalStores, totalPipelines, totalEvents, storageUsed }
  const { data: statsResponse, isLoading: loadingStats } = useQuery({
    queryKey: ['admin-state-fabrics-stats'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<StateFabricStats>('/state-fabrics/stats');
      } catch {
        return { totalFabrics: 0, activeFabrics: 0, totalStores: 0, totalPipelines: 0, totalEvents: 0, storageUsed: 0 };
      }
    },
  });

  // API returns { fabrics, total }
  const { data: listResponse, isLoading: loadingList } = useQuery({
    queryKey: ['admin-state-fabrics-list'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<StateFabricItem[]>('/state-fabrics');
      } catch {
        return { fabrics: [], total: 0 };
      }
    },
  });

  const suspendMutation = useMutation({
    mutationFn: (id: string) => adminApiClient.post(`/state-fabrics/${id}/suspend`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin-state-fabrics-list'] }),
  });

  const resumeMutation = useMutation({
    mutationFn: (id: string) => adminApiClient.post(`/state-fabrics/${id}/resume`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin-state-fabrics-list'] }),
  });

  const { data: settingsResponse, isLoading: loadingSettings } = useQuery({
    queryKey: ['admin-state-fabrics-settings'],
    queryFn: async () => adminApiClient.get<StateFabricPlatformSettings>('/state-fabrics/settings'),
    enabled: activeTab === 'settings',
  });

  const updateSettingsMutation = useMutation({
    mutationFn: (payload: StateFabricPlatformSettings) =>
      adminApiClient.patch<StateFabricPlatformSettings>('/state-fabrics/settings', payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin-state-fabrics-settings'] }),
  });

  const rawSettings = settingsResponse as StateFabricPlatformSettings | { data?: StateFabricPlatformSettings } | undefined;
  const settings: StateFabricPlatformSettings =
    rawSettings && typeof rawSettings === 'object' && 'data' in rawSettings && rawSettings.data
      ? rawSettings.data
      : (rawSettings as StateFabricPlatformSettings) ?? {};

  const loadingOverview = loadingStats || loadingList;
  if (activeTab === 'overview' && loadingOverview) {
    return <LoadingScreen />;
  }
  if (activeTab === 'settings' && loadingSettings) {
    return <LoadingScreen />;
  }

  const rawStats = statsResponse as Record<string, number> | undefined;
  const totalFabrics = rawStats?.totalFabrics ?? rawStats?.total ?? 0;
  const activeFabrics = rawStats?.activeFabrics ?? rawStats?.active ?? 0;
  const suspendedCount = totalFabrics - activeFabrics;

  const rawList = listResponse as { fabrics?: StateFabricItem[]; data?: StateFabricItem[] } | undefined;
  const fabrics = rawList?.fabrics ?? rawList?.data ?? [];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">State Fabric</h1>
        <p className="mt-2 text-gray-600">Monitor and control state fabrics across tenants.</p>
      </div>

      <div className="border-b border-gray-200">
        <nav className="flex gap-6">
          <button
            type="button"
            onClick={() => setActiveTab('overview')}
            className={`flex items-center gap-2 pb-3 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'overview'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            <LayoutDashboard className="w-4 h-4" />
            Overview
          </button>
          <button
            type="button"
            onClick={() => setActiveTab('settings')}
            className={`flex items-center gap-2 pb-3 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'settings'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            <Settings className="w-4 h-4" />
            Settings
          </button>
        </nav>
      </div>

      {activeTab === 'overview' && (
        <>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <StatCard label="Total" value={totalFabrics || fabrics.length} />
            <StatCard label="Active" value={totalFabrics ? activeFabrics : fabrics.filter((f) => f.status !== 'suspended').length} />
            <StatCard label="Suspended" value={totalFabrics ? suspendedCount : fabrics.filter((f) => f.status === 'suspended').length} />
          </div>

          <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
            <table className="w-full">
              <thead>
                <tr className="bg-gray-50 border-b border-gray-200">
                  <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">ID</th>
                  <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Name</th>
                  <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Tenant</th>
                  <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Status</th>
                  <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Action</th>
                </tr>
              </thead>
              <tbody>
                {fabrics.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="px-6 py-8 text-center text-gray-500">No state fabrics found.</td>
                  </tr>
                ) : (
                  fabrics.map((fabric) => {
                    const suspended = fabric.status === 'suspended';
                    return (
                      <tr key={fabric.id} className="border-b border-gray-100 hover:bg-gray-50">
                        <td className="px-6 py-4 text-sm text-gray-900">{fabric.id}</td>
                        <td className="px-6 py-4 text-sm text-gray-600">{fabric.name || '-'}</td>
                        <td className="px-6 py-4 text-sm text-gray-600">{fabric.tenantId ?? fabric.tenant_id ?? '-'}</td>
                        <td className="px-6 py-4 text-sm text-gray-600">{fabric.status || 'unknown'}</td>
                        <td className="px-6 py-4 text-sm">
                          <button
                            type="button"
                            disabled={suspendMutation.isPending || resumeMutation.isPending}
                            onClick={() => suspended ? resumeMutation.mutate(fabric.id) : suspendMutation.mutate(fabric.id)}
                            className="px-3 py-1 rounded bg-blue-100 text-blue-800 hover:bg-blue-200 disabled:opacity-50"
                          >
                            {suspended ? 'Resume' : 'Suspend'}
                          </button>
                        </td>
                      </tr>
                    );
                  })
                )}
              </tbody>
            </table>
          </div>
        </>
      )}

      {activeTab === 'settings' && (
        <StateFabricSettingsPanel
          settings={settings}
          onSave={(payload) => updateSettingsMutation.mutate(payload)}
          isSaving={updateSettingsMutation.isPending}
        />
      )}
    </div>
  );
}

function StatCard({ label, value }: { label: string; value: number }) {
  return (
    <div className="bg-white rounded-lg border border-gray-200 p-4">
      <p className="text-sm text-gray-600">{label}</p>
      <p className="text-2xl font-bold text-gray-900">{value}</p>
    </div>
  );
}

function StateFabricSettingsPanel({
  settings,
  onSave,
  isSaving,
}: {
  settings: StateFabricPlatformSettings;
  onSave: (payload: StateFabricPlatformSettings) => void;
  isSaving: boolean;
}) {
  const [maxFabricsPerTenant, setMaxFabricsPerTenant] = useState(settings.maxFabricsPerTenant ?? 10);
  const [defaultSnapshotRetentionDays, setDefaultSnapshotRetentionDays] = useState(
    settings.defaultSnapshotRetentionDays ?? 30
  );
  const [allowPublicPipelines, setAllowPublicPipelines] = useState(settings.allowPublicPipelines ?? false);
  const [maintenanceMode, setMaintenanceMode] = useState(settings.maintenanceMode ?? false);

  useEffect(() => {
    setMaxFabricsPerTenant(settings.maxFabricsPerTenant ?? 10);
    setDefaultSnapshotRetentionDays(settings.defaultSnapshotRetentionDays ?? 30);
    setAllowPublicPipelines(settings.allowPublicPipelines ?? false);
    setMaintenanceMode(settings.maintenanceMode ?? false);
  }, [settings.maxFabricsPerTenant, settings.defaultSnapshotRetentionDays, settings.allowPublicPipelines, settings.maintenanceMode]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSave({
      maxFabricsPerTenant: Number(maxFabricsPerTenant) || 10,
      defaultSnapshotRetentionDays: Number(defaultSnapshotRetentionDays) || 30,
      allowPublicPipelines: !!allowPublicPipelines,
      maintenanceMode: !!maintenanceMode,
    });
  };

  return (
    <div className="bg-white rounded-lg border border-gray-200 p-6 max-w-2xl">
      <h2 className="text-lg font-semibold text-gray-900 mb-4">Platform settings</h2>
      <p className="text-sm text-gray-600 mb-6">
        Configure default limits and behavior for state fabrics across all tenants.
      </p>
      <form onSubmit={handleSubmit} className="space-y-5">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Max fabrics per tenant
          </label>
          <input
            type="number"
            min={1}
            max={1000}
            value={maxFabricsPerTenant}
            onChange={(e) => setMaxFabricsPerTenant(Number(e.target.value) || 10)}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
          <p className="text-xs text-gray-500 mt-1">Maximum number of state fabrics a tenant can create (1–1000).</p>
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Default snapshot retention (days)
          </label>
          <input
            type="number"
            min={1}
            max={365}
            value={defaultSnapshotRetentionDays}
            onChange={(e) => setDefaultSnapshotRetentionDays(Number(e.target.value) || 30)}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
          <p className="text-xs text-gray-500 mt-1">How long snapshots are kept before cleanup (1–365).</p>
        </div>
        <div className="flex items-center gap-3">
          <input
            type="checkbox"
            id="allowPublicPipelines"
            checked={allowPublicPipelines}
            onChange={(e) => setAllowPublicPipelines(e.target.checked)}
            className="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
          />
          <label htmlFor="allowPublicPipelines" className="text-sm font-medium text-gray-700">
            Allow public pipelines
          </label>
        </div>
        <div className="flex items-center gap-3">
          <input
            type="checkbox"
            id="maintenanceMode"
            checked={maintenanceMode}
            onChange={(e) => setMaintenanceMode(e.target.checked)}
            className="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
          />
          <label htmlFor="maintenanceMode" className="text-sm font-medium text-gray-700">
            Maintenance mode (disable new fabric creation)
          </label>
        </div>
        <div className="pt-2">
          <button
            type="submit"
            disabled={isSaving}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
          >
            {isSaving ? 'Saving...' : 'Save settings'}
          </button>
        </div>
      </form>
    </div>
  );
}
