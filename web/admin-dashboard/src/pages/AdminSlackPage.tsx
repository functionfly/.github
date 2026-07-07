import { useState } from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { CheckCircle2, XCircle, AlertTriangle, Send, RefreshCw } from 'lucide-react';

interface SlackConfig {
  enabled: boolean;
  alert_channel?: string;
  report_channel?: string;
  channel_routing?: Record<string, string>;
  severity_config?: {
    critical: boolean;
    high: boolean;
    medium: boolean;
    low: boolean;
  };
  quiet_hours?: {
    enabled: boolean;
    start: string;
    end: string;
    timezone: string;
  };
}

interface SlackChannel {
  id: string;
  name: string;
}

const COMPONENT_OPTIONS = [
  { id: 'api', name: 'API' },
  { id: 'database', name: 'Database' },
  { id: 'cache', name: 'Cache' },
  { id: 'ai-service', name: 'AI Service' },
  { id: 'embeddings', name: 'Embeddings' },
  { id: 'state-fabric', name: 'State Fabric' },
  { id: 'microvm', name: 'MicroVM Runtime' },
  { id: 'queue', name: 'Queue Worker' },
  { id: 'function-backup', name: 'Function Backup' },
  { id: 'email', name: 'Email Delivery' },
  { id: 'billing', name: 'Billing' },
  { id: 'storage', name: 'Object Storage' },
  { id: 'cdn', name: 'CDN' },
  { id: 'pgbouncer', name: 'Connection Pool' },
  { id: 'recommendations', name: 'Recommendations' },
  { id: 'verification', name: 'Verification Pipeline' },
  { id: 'trust-api', name: 'Trust API' },
  { id: 'support', name: 'Support System' },
  { id: 'registry', name: 'Function Registry' },
  { id: 'health-monitor', name: 'Health Monitor' },
];

export function AdminSlackPage() {
  const [webhookUrl, setWebhookUrl] = useState('');
  const [botToken, setBotToken] = useState('');
  const [signingSecret, setSigningSecret] = useState('');
  const [alertChannel, setAlertChannel] = useState('');
  const [reportChannel, setReportChannel] = useState('');
  const [enabled, setEnabled] = useState(false);
  const [channelRouting, setChannelRouting] = useState<Record<string, string>>({});
  const [severityConfig, setSeverityConfig] = useState({
    critical: true,
    high: true,
    medium: true,
    low: false,
  });
  const [quietHours, setQuietHours] = useState({
    enabled: false,
    start: '22:00',
    end: '08:00',
    timezone: 'UTC',
  });

  const { data: configData, isLoading, refetch } = useQuery({
    queryKey: ['slack-config'],
    queryFn: async () => {
      const response = await adminApiClient.get<SlackConfig>('/admin/slack/config');
      return response.data;
    },
  });

  const { data: channelsData } = useQuery({
    queryKey: ['slack-channels'],
    queryFn: async () => {
      const response = await adminApiClient.get<{ channels: SlackChannel[] }>('/admin/slack/channels');
      return response.data;
    },
    enabled: enabled,
  });

  const saveMutation = useMutation({
    mutationFn: async (config: Partial<SlackConfig>) => {
      await adminApiClient.put('/admin/slack/config', config);
    },
    onSuccess: () => {
      refetch();
    },
  });

  const testMutation = useMutation({
    mutationFn: async () => {
      await adminApiClient.post('/admin/slack/test', {});
    },
  });

  const handleSave = () => {
    saveMutation.mutate({
      enabled,
      webhook_url: webhookUrl,
      bot_token: botToken,
      signing_secret: signingSecret,
      alert_channel: alertChannel,
      report_channel: reportChannel,
      channel_routing: channelRouting,
      severity_config: severityConfig,
      quiet_hours: quietHours,
    });
  };

  const handleTest = () => {
    testMutation.mutate();
  };

  if (isLoading) {
    return <LoadingScreen />;
  }

  const channels = channelsData?.channels || [];

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Slack Integration</h1>
          <p className="text-gray-500 mt-1">Configure Slack notifications for platform status</p>
        </div>
        <div className="flex items-center gap-4">
          <span className={`inline-flex items-center gap-2 px-3 py-1 rounded-full text-sm font-medium ${
            enabled ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-500'
          }`}>
            {enabled ? <CheckCircle2 className="h-4 w-4" /> : <XCircle className="h-4 w-4" />}
            {enabled ? 'Connected' : 'Disabled'}
          </span>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow p-6 space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-lg font-semibold">Enable Slack Integration</h2>
            <p className="text-sm text-gray-500">Send status updates and alerts to Slack</p>
          </div>
          <button
            type="button"
            onClick={() => setEnabled(!enabled)}
            className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
              enabled ? 'bg-blue-600' : 'bg-gray-200'
            }`}
          >
            <span
              className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                enabled ? 'translate-x-6' : 'translate-x-1'
              }`}
            />
          </button>
        </div>
      </div>

      {enabled && (
        <>
          <div className="bg-white rounded-lg shadow p-6 space-y-6">
            <h2 className="text-lg font-semibold">API Credentials</h2>
            <div className="grid grid-cols-1 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Webhook URL
                </label>
                <input
                  type="url"
                  value={webhookUrl}
                  onChange={(e) => setWebhookUrl(e.target.value)}
                  placeholder="https://hooks.slack.com/services/..."
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <p className="text-xs text-gray-500 mt-1">Incoming webhook URL for alerts</p>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Bot Token
                </label>
                <input
                  type="password"
                  value={botToken}
                  onChange={(e) => setBotToken(e.target.value)}
                  placeholder="xoxb-..."
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <p className="text-xs text-gray-500 mt-1">Bot token for slash commands (xoxb-...)</p>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Signing Secret
                </label>
                <input
                  type="password"
                  value={signingSecret}
                  onChange={(e) => setSigningSecret(e.target.value)}
                  placeholder="..."
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <p className="text-xs text-gray-500 mt-1">Request verification for slash commands</p>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6 space-y-6">
            <h2 className="text-lg font-semibold">Channels</h2>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Alert Channel
                </label>
                <select
                  value={alertChannel}
                  onChange={(e) => setAlertChannel(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="">Select channel</option>
                  {channels.map((ch) => (
                    <option key={ch.id} value={ch.id}>
                      #{ch.name}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Report Channel
                </label>
                <select
                  value={reportChannel}
                  onChange={(e) => setReportChannel(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="">Select channel</option>
                  {channels.map((ch) => (
                    <option key={ch.id} value={ch.id}>
                      #{ch.name}
                    </option>
                  ))}
                </select>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6 space-y-6">
            <h2 className="text-lg font-semibold">Per-Service Channel Routing</h2>
            <p className="text-sm text-gray-500">Route alerts for specific services to different channels</p>
            <div className="space-y-3">
              {COMPONENT_OPTIONS.map((comp) => (
                <div key={comp.id} className="flex items-center gap-4">
                  <span className="w-40 text-sm font-medium text-gray-700">{comp.name}</span>
                  <input
                    type="text"
                    value={channelRouting[comp.id] || ''}
                    onChange={(e) => setChannelRouting({ ...channelRouting, [comp.id]: e.target.value })}
                    placeholder="Channel ID (optional)"
                    className="flex-1 px-3 py-1 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              ))}
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6 space-y-6">
            <h2 className="text-lg font-semibold">Severity Thresholds</h2>
            <p className="text-sm text-gray-500">Choose which severity levels trigger Slack alerts</p>
            <div className="space-y-3">
              {Object.entries(severityConfig).map(([severity, enabled]) => (
                <div key={severity} className="flex items-center justify-between">
                  <span className="text-sm font-medium text-gray-700 capitalize">{severity}</span>
                  <button
                    type="button"
                    onClick={() => setSeverityConfig({ ...severityConfig, [severity]: !enabled })}
                    className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                      enabled ? 'bg-blue-600' : 'bg-gray-200'
                    }`}
                  >
                    <span
                      className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                        enabled ? 'translate-x-6' : 'translate-x-1'
                      }`}
                    />
                  </button>
                </div>
              ))}
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6 space-y-6">
            <h2 className="text-lg font-semibold">Quiet Hours</h2>
            <p className="text-sm text-gray-500">Suppress non-critical alerts during specified hours</p>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-gray-700">Enable Quiet Hours</span>
                <button
                  type="button"
                  onClick={() => setQuietHours({ ...quietHours, enabled: !quietHours.enabled })}
                  className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                    quietHours.enabled ? 'bg-blue-600' : 'bg-gray-200'
                  }`}
                >
                  <span
                    className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                      quietHours.enabled ? 'translate-x-6' : 'translate-x-1'
                    }`}
                  />
                </button>
              </div>
              {quietHours.enabled && (
                <div className="grid grid-cols-3 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Start</label>
                    <input
                      type="time"
                      value={quietHours.start}
                      onChange={(e) => setQuietHours({ ...quietHours, start: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">End</label>
                    <input
                      type="time"
                      value={quietHours.end}
                      onChange={(e) => setQuietHours({ ...quietHours, end: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Timezone</label>
                    <select
                      value={quietHours.timezone}
                      onChange={(e) => setQuietHours({ ...quietHours, timezone: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      <option value="UTC">UTC</option>
                      <option value="America/New_York">America/New_York</option>
                      <option value="America/Chicago">America/Chicago</option>
                      <option value="America/Denver">America/Denver</option>
                      <option value="America/Los_Angeles">America/Los_Angeles</option>
                      <option value="Europe/London">Europe/London</option>
                      <option value="Europe/Paris">Europe/Paris</option>
                      <option value="Asia/Tokyo">Asia/Tokyo</option>
                    </select>
                  </div>
                </div>
              )}
            </div>
          </div>

          <div className="flex justify-end gap-4">
            <button
              type="button"
              onClick={handleTest}
              disabled={testMutation.isPending || !webhookUrl}
              className="inline-flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {testMutation.isPending ? (
                <RefreshCw className="h-4 w-4 animate-spin" />
              ) : (
                <Send className="h-4 w-4" />
              )}
              Send Test Message
            </button>
            <button
              type="button"
              onClick={handleSave}
              disabled={saveMutation.isPending}
              className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-md text-sm font-medium hover:bg-blue-700 disabled:opacity-50"
            >
              {saveMutation.isPending && <RefreshCw className="h-4 w-4 animate-spin" />}
              Save Configuration
            </button>
          </div>
        </>
      )}
    </div>
  );
}
