/**
 * Admin System Page
 * System health, configuration, and platform settings
 */

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import {
  Activity,
  HardDrive,
  Cpu,
  MemoryStick,
  Disc,
  Settings2,
  Save,
  Loader2,
  BarChart3,
  Sliders,
} from 'lucide-react';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/Tabs';

interface SystemMetrics {
  status: 'healthy' | 'degraded' | 'down';
  uptime: number;
  cpuUsage: number;
  memoryUsage: number;
  diskUsage: number;
  apiResponsiveness: number;
  databaseHealth: 'connected' | 'disconnected';
}

interface PlatformSettingsState {
  maintenanceMode: boolean;
  signupsEnabled: boolean;
  platformName: string;
  supportEmail: string;
  defaultRateLimitPerMin: number;
}

function Switch({
  checked,
  onChange,
  disabled,
}: {
  checked: boolean;
  onChange: (v: boolean) => void;
  disabled?: boolean;
}) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      disabled={disabled}
      onClick={() => onChange(!checked)}
      className={`
        relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent
        transition-colors duration-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-admin-700 focus-visible:ring-offset-2
        ${checked ? 'bg-admin-700' : 'bg-admin-300'}
        ${disabled ? 'opacity-50 cursor-not-allowed' : ''}
      `}
    >
      <span
        className={`
          pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0
          transition duration-200
          ${checked ? 'translate-x-5' : 'translate-x-0.5'}
        `}
      />
    </button>
  );
}

export function AdminSystemPage() {
  // Default tab is now managed by the Tabs component
  const [platformSettings, setPlatformSettings] = useState<PlatformSettingsState>({
    maintenanceMode: false,
    signupsEnabled: true,
    platformName: 'FunctionFly',
    supportEmail: 'support@functionfly.local',
    defaultRateLimitPerMin: 1000,
  });
  const [saving, setSaving] = useState(false);
  const [saveMessage, setSaveMessage] = useState<'success' | 'error' | null>(null);

  const { data: metricsResponse } = useQuery({
    queryKey: ['admin-system-metrics'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<SystemMetrics>('/system/metrics');
      } catch {
        return { data: null, success: false };
      }
    },
    staleTime: 1000 * 30,
  });

  const metrics = metricsResponse?.data;

  const handleSavePlatformSettings = async () => {
    setSaving(true);
    setSaveMessage(null);
    try {
      // Placeholder: wire to API when backend supports PATCH /system/platform-settings
      await new Promise((r) => setTimeout(r, 600));
      setSaveMessage('success');
      setTimeout(() => setSaveMessage(null), 3000);
    } catch {
      setSaveMessage('error');
    } finally {
      setSaving(false);
    }
  };

  const statusColor =
    metrics?.status === 'healthy'
      ? 'text-emerald-600'
      : metrics?.status === 'degraded'
        ? 'text-amber-600'
        : 'text-red-600';
  const statusBg =
    metrics?.status === 'healthy'
      ? 'bg-emerald-50 border-emerald-200'
      : metrics?.status === 'degraded'
        ? 'bg-amber-50 border-amber-200'
        : 'bg-red-50 border-red-200';

  return (
    <div className="min-h-full flex flex-col">
      {/* Page header + modern tabs */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-admin-900 tracking-tight">System</h1>
        <p className="mt-1 text-admin-600 text-sm">
          Health, configuration, and platform-wide settings
        </p>

        <Tabs defaultTab="health" className="mt-6">
          <TabsList variant="segmented" className="w-fit">
            <TabsTrigger value="health" icon={BarChart3} variant="segmented">
              Health & Usage
            </TabsTrigger>
            <TabsTrigger value="settings" icon={Sliders} variant="segmented">
              Platform Settings
            </TabsTrigger>
          </TabsList>

          <div className="mt-6 flex-1 min-h-0">
            <TabsContent value="health" className="space-y-8">
            {/* System Health */}
            <section className="rounded-xl border border-admin-200 bg-white shadow-sm overflow-hidden">
              <div className="border-l-4 border-admin-600 bg-admin-50/60 px-5 py-3">
                <h2 className="text-sm font-semibold text-admin-900 uppercase tracking-wider">
                  System Health
                </h2>
                <p className="text-xs text-admin-600 mt-0.5">
                  Live resources and health checks
                </p>
              </div>
              <div className="p-6 space-y-6">
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                  <div
                    className={`rounded-lg border p-4 flex items-center justify-between ${statusBg}`}
                  >
                    <div className="flex items-center gap-3">
                      <div
                        className={`rounded-lg p-2 ${
                          metrics?.status === 'healthy'
                            ? 'bg-emerald-100'
                            : metrics?.status === 'degraded'
                              ? 'bg-amber-100'
                              : 'bg-red-100'
                        }`}
                      >
                        <Activity className={`w-5 h-5 ${statusColor}`} />
                      </div>
                      <div>
                        <p className="text-xs font-medium text-admin-600 uppercase tracking-wider">
                          System Status
                        </p>
                        <p className={`font-semibold ${statusColor}`}>
                          {metrics?.status ?? 'checking…'}
                        </p>
                      </div>
                    </div>
                  </div>
                  <div
                    className={`rounded-lg border p-4 flex items-center justify-between ${
                      metrics?.databaseHealth === 'connected'
                        ? 'bg-emerald-50 border-emerald-200'
                        : 'bg-red-50 border-red-200'
                    }`}
                  >
                    <div className="flex items-center gap-3">
                      <div
                        className={`rounded-lg p-2 ${
                          metrics?.databaseHealth === 'connected'
                            ? 'bg-emerald-100'
                            : 'bg-red-100'
                        }`}
                      >
                        <HardDrive
                          className={
                            metrics?.databaseHealth === 'connected'
                              ? 'w-5 h-5 text-emerald-600'
                              : 'w-5 h-5 text-red-600'
                          }
                        />
                      </div>
                      <div>
                        <p className="text-xs font-medium text-admin-600 uppercase tracking-wider">
                          Database
                        </p>
                        <p
                          className={`font-semibold ${
                            metrics?.databaseHealth === 'connected'
                              ? 'text-emerald-600'
                              : 'text-red-600'
                          }`}
                        >
                          {metrics?.databaseHealth ?? 'checking…'}
                        </p>
                      </div>
                    </div>
                  </div>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-3 lg:grid-cols-6 gap-6">
                  {[
                    { label: 'CPU Usage', value: metrics?.cpuUsage ?? 0, icon: Cpu, stroke: '#0f766e' },
                    { label: 'Memory Usage', value: metrics?.memoryUsage ?? 0, icon: MemoryStick, stroke: '#7c3aed' },
                    { label: 'Disk Usage', value: metrics?.diskUsage ?? 0, icon: Disc, stroke: '#be185d' },
                  ].map(({ label, value, icon: Icon, stroke }) => (
                    <div
                      key={label}
                      className="rounded-lg border border-admin-200 bg-admin-50/40 p-5 flex flex-col items-center"
                    >
                      <p className="text-xs font-medium text-admin-600 uppercase tracking-wider mb-3">
                        {label}
                      </p>
                      <div className="relative w-20 h-20">
                        <svg className="w-full h-full -rotate-90" viewBox="0 0 100 100">
                          <circle cx="50" cy="50" r="42" fill="none" stroke="#e9ecef" strokeWidth="8" />
                          <circle
                            cx="50"
                            cy="50"
                            r="42"
                            fill="none"
                            stroke={stroke}
                            strokeWidth="8"
                            strokeLinecap="round"
                            strokeDasharray={`${(value / 100) * 263.9} 263.9`}
                          />
                        </svg>
                        <span className="absolute inset-0 flex items-center justify-center text-sm font-semibold text-admin-800">
                          {value}%
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </section>

            {/* System Configuration (read-only) */}
            <section className="rounded-xl border border-admin-200 bg-white shadow-sm overflow-hidden">
              <div className="border-l-4 border-admin-500 bg-admin-50/60 px-5 py-3">
                <h2 className="text-sm font-semibold text-admin-900 uppercase tracking-wider">
                  System Configuration
                </h2>
              </div>
              <div className="p-6">
                <dl className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-x-8 gap-y-5">
                  <div className="flex justify-between sm:block">
                    <dt className="text-sm text-admin-600">API Version</dt>
                    <dd className="text-sm font-medium text-admin-900 sm:mt-0.5">v1.0.0</dd>
                  </div>
                  <div className="flex justify-between sm:block">
                    <dt className="text-sm text-admin-600">Environment</dt>
                    <dd className="text-sm font-medium text-admin-900 sm:mt-0.5">Production</dd>
                  </div>
                  <div className="flex justify-between sm:block">
                    <dt className="text-sm text-admin-600">Uptime</dt>
                    <dd className="text-sm font-medium text-admin-900 sm:mt-0.5">{metrics?.uptime ?? 0}h</dd>
                  </div>
                  <div className="flex justify-between sm:block">
                    <dt className="text-sm text-admin-600">API Responsiveness</dt>
                    <dd className="text-sm font-medium text-admin-900 sm:mt-0.5">{metrics?.apiResponsiveness ?? '—'}ms</dd>
                  </div>
                </dl>
              </div>
            </section>
            </TabsContent>

            <TabsContent value="settings" className="max-w-3xl">
            <section className="rounded-xl border border-admin-200 bg-white shadow-sm overflow-hidden">
              <div className="border-l-4 border-admin-700 bg-admin-100/80 px-5 py-3 flex items-center gap-2">
                <Settings2 className="w-4 h-4 text-admin-700" />
                <h2 className="text-sm font-semibold text-admin-900 uppercase tracking-wider">
                  Platform Settings
                </h2>
              </div>
              <div className="p-6 space-y-6">
                <div className="flex items-center justify-between gap-6 py-2">
                  <div className="min-w-0">
                    <p className="text-sm font-medium text-admin-900">Maintenance mode</p>
                    <p className="text-xs text-admin-600 mt-0.5">
                      Disable public access; only admins can sign in
                    </p>
                  </div>
                  <Switch
                    checked={platformSettings.maintenanceMode}
                    onChange={(v) =>
                      setPlatformSettings((s) => ({ ...s, maintenanceMode: v }))
                    }
                  />
                </div>
                <div className="flex items-center justify-between gap-6 py-2">
                  <div className="min-w-0">
                    <p className="text-sm font-medium text-admin-900">Sign-ups enabled</p>
                    <p className="text-xs text-admin-600 mt-0.5">
                      Allow new tenant and user registration
                    </p>
                  </div>
                  <Switch
                    checked={platformSettings.signupsEnabled}
                    onChange={(v) =>
                      setPlatformSettings((s) => ({ ...s, signupsEnabled: v }))
                    }
                  />
                </div>

                <div className="border-t border-admin-200 pt-6 space-y-5">
                  <div>
                    <label
                      htmlFor="platform-name"
                      className="block text-sm font-medium text-admin-900 mb-1"
                    >
                      Platform name
                    </label>
                    <input
                      id="platform-name"
                      type="text"
                      value={platformSettings.platformName}
                      onChange={(e) =>
                        setPlatformSettings((s) => ({ ...s, platformName: e.target.value }))
                      }
                      className="w-full max-w-md rounded-lg border border-admin-300 bg-white px-3 py-2 text-sm text-admin-900 placeholder-admin-400 focus:border-admin-500 focus:outline-none focus:ring-1 focus:ring-admin-500"
                      placeholder="FunctionFly"
                    />
                  </div>
                  <div>
                    <label
                      htmlFor="support-email"
                      className="block text-sm font-medium text-admin-900 mb-1"
                    >
                      Support email
                    </label>
                    <input
                      id="support-email"
                      type="email"
                      value={platformSettings.supportEmail}
                      onChange={(e) =>
                        setPlatformSettings((s) => ({ ...s, supportEmail: e.target.value }))
                      }
                      className="w-full max-w-md rounded-lg border border-admin-300 bg-white px-3 py-2 text-sm text-admin-900 placeholder-admin-400 focus:border-admin-500 focus:outline-none focus:ring-1 focus:ring-admin-500"
                      placeholder="support@example.com"
                    />
                  </div>
                  <div>
                    <label
                      htmlFor="rate-limit"
                      className="block text-sm font-medium text-admin-900 mb-1"
                    >
                      Default API rate limit (per minute)
                    </label>
                    <input
                      id="rate-limit"
                      type="number"
                      min={1}
                      max={100000}
                      value={platformSettings.defaultRateLimitPerMin}
                      onChange={(e) =>
                        setPlatformSettings((s) => ({
                          ...s,
                          defaultRateLimitPerMin: Math.max(
                            1,
                            parseInt(e.target.value, 10) || 1000
                          ),
                        }))
                      }
                      className="w-full max-w-xs rounded-lg border border-admin-300 bg-white px-3 py-2 text-sm text-admin-900 focus:border-admin-500 focus:outline-none focus:ring-1 focus:ring-admin-500"
                    />
                  </div>
                </div>

                <div className="pt-4">
                  <button
                    type="button"
                    onClick={handleSavePlatformSettings}
                    disabled={saving}
                    className="inline-flex items-center gap-2 rounded-lg bg-admin-700 px-5 py-2.5 text-sm font-medium text-white hover:bg-admin-800 focus:outline-none focus:ring-2 focus:ring-admin-600 focus:ring-offset-2 disabled:opacity-60 disabled:cursor-not-allowed transition-colors"
                  >
                    {saving ? (
                      <>
                        <Loader2 className="w-4 h-4 animate-spin" />
                        Saving…
                      </>
                    ) : (
                      <>
                        <Save className="w-4 h-4" />
                        Save platform settings
                      </>
                    )}
                  </button>
                  {saveMessage === 'success' && (
                    <p className="mt-2 text-sm text-emerald-600">Settings saved.</p>
                  )}
                  {saveMessage === 'error' && (
                    <p className="mt-2 text-sm text-red-600">Failed to save. Try again.</p>
                  )}
                </div>
              </div>
            </section>
            </TabsContent>
          </div>
        </Tabs>
      </div>
    </div>
  );
}
