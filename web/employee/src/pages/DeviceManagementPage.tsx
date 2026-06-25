import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { devicesApi, type Device } from '@/api/devices';
import {
  Monitor,
  Smartphone,
  Laptop,
  Server,
  Plus,
  Search,
  Clock,
  CheckCircle,
  AlertTriangle,
  HelpCircle,
} from 'lucide-react';
import { toast } from 'sonner';

const deviceTypeIcons: Record<string, typeof Monitor> = {
  desktop: Monitor,
  laptop: Laptop,
  mobile: Smartphone,
  tablet: Smartphone,
  server: Server,
};

const complianceColors: Record<string, string> = {
  compliant: 'bg-green-500/20 text-green-400',
  non_compliant: 'bg-red-500/20 text-red-400',
  unknown: 'bg-gray-500/20 text-gray-400',
};

const complianceIcons: Record<string, typeof CheckCircle> = {
  compliant: CheckCircle,
  non_compliant: AlertTriangle,
  unknown: HelpCircle,
};

const statusColors: Record<string, string> = {
  active: 'bg-green-500/20 text-green-400',
  inactive: 'bg-gray-500/20 text-gray-400',
  retired: 'bg-red-500/20 text-red-400',
  lost: 'bg-red-500/20 text-red-400',
};

export function DeviceManagementPage() {
  const queryClient = useQueryClient();
  const [showRegister, setShowRegister] = useState(false);
  const [detailId, setDetailId] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const [typeFilter, setTypeFilter] = useState('');
  const [form, setForm] = useState({
    device_name: '',
    device_type: 'laptop',
    serial_number: '',
    os: '',
    os_version: '',
    manufacturer: '',
    model: '',
  });

  const { data, isLoading } = useQuery({
    queryKey: ['devices'],
    queryFn: () => devicesApi.list(),
  });

  const { data: detailData } = useQuery({
    queryKey: ['devices', detailId],
    queryFn: () => devicesApi.get(detailId!),
    enabled: !!detailId,
  });

  const registerMutation = useMutation({
    mutationFn: () => devicesApi.register(form),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['devices'] });
      toast.success('Device registered');
      setShowRegister(false);
      setForm({ device_name: '', device_type: 'laptop', serial_number: '', os: '', os_version: '', manufacturer: '', model: '' });
    },
    onError: () => toast.error('Failed to register device'),
  });

  const devices = (data?.data?.devices || []).filter((d) => {
    if (typeFilter && d.device_type !== typeFilter) return false;
    if (
      search &&
      !d.device_name.toLowerCase().includes(search.toLowerCase()) &&
      !(d.serial_number && d.serial_number.toLowerCase().includes(search.toLowerCase()))
    )
      return false;
    return true;
  });

  const compliantCount = (data?.data?.devices || []).filter((d) => d.compliance_status === 'compliant').length;
  const nonCompliantCount = (data?.data?.devices || []).filter((d) => d.compliance_status === 'non_compliant').length;
  const types = [...new Set((data?.data?.devices || []).map((d) => d.device_type))].sort();

  const detail = detailData?.data?.device;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Monitor className="h-6 w-6 text-indigo-400" />
          <h1 className="text-2xl font-bold">Device Management</h1>
        </div>
        <button
          onClick={() => setShowRegister(true)}
          className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          <Plus className="h-4 w-4" />
          Register Device
        </button>
      </div>

      <div className="grid grid-cols-3 gap-3">
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <div className="flex items-center gap-2 text-gray-400">
            <Monitor className="h-4 w-4 text-indigo-400" />
            <span className="text-sm">Total Devices</span>
          </div>
          <p className="mt-1 text-2xl font-bold">{data?.data?.devices?.length || 0}</p>
        </div>
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <div className="flex items-center gap-2 text-gray-400">
            <CheckCircle className="h-4 w-4 text-green-400" />
            <span className="text-sm">Compliant</span>
          </div>
          <p className="mt-1 text-2xl font-bold text-green-400">{compliantCount}</p>
        </div>
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <div className="flex items-center gap-2 text-gray-400">
            <AlertTriangle className="h-4 w-4 text-red-400" />
            <span className="text-sm">Non-Compliant</span>
          </div>
          <p className="mt-1 text-2xl font-bold text-red-400">{nonCompliantCount}</p>
        </div>
      </div>

      <div className="flex items-center gap-3">
        <Search className="h-4 w-4 text-gray-400" />
        <input
          type="text"
          placeholder="Search devices..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-64 rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
        />
        <select
          value={typeFilter}
          onChange={(e) => setTypeFilter(e.target.value)}
          className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-200"
        >
          <option value="">All Types</option>
          {types.map((t) => (
            <option key={t} value={t}>
              {t}
            </option>
          ))}
        </select>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : devices.length === 0 ? (
        <div className="rounded-xl border border-gray-800 bg-gray-900 py-16 text-center text-gray-500">
          No devices found
        </div>
      ) : (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {devices.map((device) => {
            const TypeIcon = deviceTypeIcons[device.device_type] || Monitor;
            const ComplianceIcon = complianceIcons[device.compliance_status] || HelpCircle;
            return (
              <button
                key={device.id}
                onClick={() => setDetailId(device.id)}
                className="rounded-xl border border-gray-800 bg-gray-900 p-4 text-left transition-colors hover:border-gray-700"
              >
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-3">
                    <div className="rounded-lg bg-indigo-500/10 p-2">
                      <TypeIcon className="h-5 w-5 text-indigo-400" />
                    </div>
                    <div>
                      <p className="font-medium">{device.device_name}</p>
                      <p className="text-xs text-gray-500 capitalize">{device.device_type}</p>
                    </div>
                  </div>
                  <span
                    className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ${complianceColors[device.compliance_status] || 'bg-gray-500/20 text-gray-400'}`}
                  >
                    <ComplianceIcon className="h-3 w-3" />
                    {device.compliance_status.replace('_', ' ')}
                  </span>
                </div>
                <div className="mt-3 flex items-center gap-4 text-xs text-gray-500">
                  {device.os && (
                    <span>
                      {device.os} {device.os_version}
                    </span>
                  )}
                  {device.manufacturer && <span>{device.manufacturer}</span>}
                  {device.serial_number && (
                    <code className="rounded bg-gray-800 px-1.5 py-0.5 text-[10px]">{device.serial_number}</code>
                  )}
                </div>
                <div className="mt-2 flex items-center gap-2">
                  <span
                    className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ${statusColors[device.status] || 'bg-gray-500/20 text-gray-400'}`}
                  >
                    {device.status}
                  </span>
                  {device.last_seen_at && (
                    <span className="flex items-center gap-1 text-xs text-gray-500">
                      <Clock className="h-3 w-3" />
                      {new Date(device.last_seen_at).toLocaleDateString()}
                    </span>
                  )}
                </div>
              </button>
            );
          })}
        </div>
      )}

      {detail && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
          <div className="w-full max-w-lg rounded-xl border border-gray-700 bg-gray-900 p-6">
            <div className="flex items-start justify-between">
              <h2 className="text-lg font-semibold">{detail.device_name}</h2>
              <button onClick={() => setDetailId(null)} className="text-gray-400 hover:text-gray-200">
                ✕
              </button>
            </div>
            <div className="mt-4 space-y-3 text-sm">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <span className="text-gray-500">Type</span>
                  <p className="capitalize">{detail.device_type}</p>
                </div>
                <div>
                  <span className="text-gray-500">Status</span>
                  <p className="capitalize">{detail.status}</p>
                </div>
                <div>
                  <span className="text-gray-500">Serial</span>
                  <p className="font-mono text-xs">{detail.serial_number || '—'}</p>
                </div>
                <div>
                  <span className="text-gray-500">Compliance</span>
                  <p className="capitalize">{detail.compliance_status.replace('_', ' ')}</p>
                </div>
                <div>
                  <span className="text-gray-500">OS</span>
                  <p>
                    {detail.os || '—'} {detail.os_version || ''}
                  </p>
                </div>
                <div>
                  <span className="text-gray-500">Manufacturer</span>
                  <p>
                    {detail.manufacturer || '—'} {detail.model || ''}
                  </p>
                </div>
                <div>
                  <span className="text-gray-500">Enrolled</span>
                  <p>{detail.enrolled_at ? new Date(detail.enrolled_at).toLocaleDateString() : '—'}</p>
                </div>
                <div>
                  <span className="text-gray-500">Last Seen</span>
                  <p>{detail.last_seen_at ? new Date(detail.last_seen_at).toLocaleString() : '—'}</p>
                </div>
              </div>
            </div>
            <div className="mt-6 flex justify-end">
              <button
                onClick={() => setDetailId(null)}
                className="rounded-lg border border-gray-700 px-4 py-2 text-sm text-gray-300 hover:bg-gray-800"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}

      {showRegister && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
          <div className="w-full max-w-lg rounded-xl border border-gray-700 bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Register Device</h2>
            <div className="space-y-4">
              <div>
                <label className="mb-1 block text-sm text-gray-400">Device Name</label>
                <input
                  type="text"
                  value={form.device_name}
                  onChange={(e) => setForm({ ...form, device_name: e.target.value })}
                  placeholder="e.g. MacBook Pro - Engineering"
                  className="w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
                />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="mb-1 block text-sm text-gray-400">Device Type</label>
                  <select
                    value={form.device_type}
                    onChange={(e) => setForm({ ...form, device_type: e.target.value })}
                    className="w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-200"
                  >
                    <option value="laptop">Laptop</option>
                    <option value="desktop">Desktop</option>
                    <option value="mobile">Mobile</option>
                    <option value="tablet">Tablet</option>
                    <option value="server">Server</option>
                  </select>
                </div>
                <div>
                  <label className="mb-1 block text-sm text-gray-400">Serial Number</label>
                  <input
                    type="text"
                    value={form.serial_number}
                    onChange={(e) => setForm({ ...form, serial_number: e.target.value })}
                    placeholder="SN-12345"
                    className="w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="mb-1 block text-sm text-gray-400">OS</label>
                  <input
                    type="text"
                    value={form.os}
                    onChange={(e) => setForm({ ...form, os: e.target.value })}
                    placeholder="macOS, Windows, Linux"
                    className="w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
                  />
                </div>
                <div>
                  <label className="mb-1 block text-sm text-gray-400">OS Version</label>
                  <input
                    type="text"
                    value={form.os_version}
                    onChange={(e) => setForm({ ...form, os_version: e.target.value })}
                    placeholder="14.0"
                    className="w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="mb-1 block text-sm text-gray-400">Manufacturer</label>
                  <input
                    type="text"
                    value={form.manufacturer}
                    onChange={(e) => setForm({ ...form, manufacturer: e.target.value })}
                    placeholder="Apple, Dell, Lenovo"
                    className="w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
                  />
                </div>
                <div>
                  <label className="mb-1 block text-sm text-gray-400">Model</label>
                  <input
                    type="text"
                    value={form.model}
                    onChange={(e) => setForm({ ...form, model: e.target.value })}
                    placeholder="MacBook Pro 16"
                    className="w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
                  />
                </div>
              </div>
            </div>
            <div className="mt-6 flex justify-end gap-3">
              <button
                onClick={() => setShowRegister(false)}
                className="rounded-lg border border-gray-700 px-4 py-2 text-sm text-gray-300 hover:bg-gray-800"
              >
                Cancel
              </button>
              <button
                onClick={() => registerMutation.mutate()}
                disabled={!form.device_name || registerMutation.isPending}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                {registerMutation.isPending ? 'Registering...' : 'Register'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
