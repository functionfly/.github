import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { pushApi, type NotificationPreference } from '@/api/push';
import { BellRing, Save, Mail, MessageSquare, Bell, Hash, Clock } from 'lucide-react';
import { toast } from 'sonner';

const channels = [
  { key: 'in_app', label: 'In-App', icon: Bell },
  { key: 'email', label: 'Email', icon: Mail },
  { key: 'push', label: 'Push', icon: BellRing },
  { key: 'slack', label: 'Slack', icon: Hash },
];

const eventTypes = [
  { key: 'task_assigned', label: 'Task Assigned' },
  { key: 'task_completed', label: 'Task Completed' },
  { key: 'mention', label: 'Mention' },
  { key: 'comment', label: 'Comment' },
  { key: 'approval_required', label: 'Approval Required' },
  { key: 'system_alert', label: 'System Alert' },
  { key: 'learning_reminder', label: 'Learning Reminder' },
  { key: 'performance_review', label: 'Performance Review' },
];

export function NotificationSettingsPage() {
  const queryClient = useQueryClient();
  const [activeChannel, setActiveChannel] = useState('in_app');
  const [quietHours, setQuietHours] = useState({ start: '22:00', end: '08:00' });
  const [localPrefs, setLocalPrefs] = useState<Record<string, Record<string, boolean>>>({});

  const { data, isLoading } = useQuery({
    queryKey: ['notification-preferences'],
    queryFn: () => pushApi.getPreferences(),
  });

  useEffect(() => {
    if (!data?.data?.preferences) return;
    const grouped: Record<string, Record<string, boolean>> = {};
    for (const pref of data.data.preferences) {
      if (!grouped[pref.channel]) grouped[pref.channel] = {};
      grouped[pref.channel][pref.event_type] = pref.is_enabled;
      if (pref.quiet_hours_start) setQuietHours((prev) => ({ ...prev, start: pref.quiet_hours_start! }));
      if (pref.quiet_hours_end) setQuietHours((prev) => ({ ...prev, end: pref.quiet_hours_end! }));
    }
    setLocalPrefs(grouped);
  }, [data?.data?.preferences]);

  const saveMutation = useMutation({
    mutationFn: async () => {
      const prefs = data?.data?.preferences || [];
      const updates: Promise<unknown>[] = [];
      for (const channel of Object.keys(localPrefs)) {
        for (const eventType of Object.keys(localPrefs[channel])) {
          const existing = prefs.find((p) => p.channel === channel && p.event_type === eventType);
          if (existing) {
            updates.push(
              pushApi.updatePreference(existing.id, {
                is_enabled: localPrefs[channel][eventType],
                quiet_hours_start: quietHours.start,
                quiet_hours_end: quietHours.end,
              }),
            );
          }
        }
      }
      await Promise.all(updates);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notification-preferences'] });
      toast.success('Preferences saved');
    },
    onError: () => toast.error('Failed to save preferences'),
  });

  const togglePref = (channel: string, eventType: string) => {
    setLocalPrefs((prev) => ({
      ...prev,
      [channel]: {
        ...prev[channel],
        [eventType]: !prev[channel]?.[eventType],
      },
    }));
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <BellRing className="h-6 w-6 text-orange-400" />
          <h1 className="text-2xl font-bold">Notification Settings</h1>
        </div>
        <button
          onClick={() => saveMutation.mutate()}
          disabled={saveMutation.isPending}
          className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
        >
          <Save className="h-4 w-4" />
          {saveMutation.isPending ? 'Saving...' : 'Save'}
        </button>
      </div>

      <div className="flex gap-1 rounded-lg border border-gray-800 bg-gray-900 p-1">
        {channels.map((ch) => {
          const Icon = ch.icon;
          return (
            <button
              key={ch.key}
              onClick={() => setActiveChannel(ch.key)}
              className={`flex flex-1 items-center justify-center gap-2 rounded-md px-3 py-2.5 text-sm font-medium transition-colors ${
                activeChannel === ch.key
                  ? 'bg-blue-600 text-white'
                  : 'text-gray-400 hover:bg-gray-800 hover:text-gray-200'
              }`}
            >
              <Icon className="h-4 w-4" />
              {ch.label}
            </button>
          );
        })}
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : (
        <div className="rounded-xl border border-gray-800 bg-gray-900">
          <div className="border-b border-gray-800 px-6 py-4">
            <h2 className="text-sm font-semibold text-gray-300">
              {channels.find((c) => c.key === activeChannel)?.label} Notifications
            </h2>
            <p className="text-xs text-gray-500">Toggle which events trigger notifications for this channel</p>
          </div>
          <div className="divide-y divide-gray-800">
            {eventTypes.map((evt) => {
              const isEnabled = localPrefs[activeChannel]?.[evt.key] ?? false;
              return (
                <div key={evt.key} className="flex items-center justify-between px-6 py-4">
                  <span className="text-sm text-gray-200">{evt.label}</span>
                  <button
                    onClick={() => togglePref(activeChannel, evt.key)}
                    className={`relative h-6 w-11 rounded-full transition-colors ${
                      isEnabled ? 'bg-blue-600' : 'bg-gray-700'
                    }`}
                  >
                    <span
                      className={`absolute left-0.5 top-0.5 h-5 w-5 rounded-full bg-white shadow transition-transform ${
                        isEnabled ? 'translate-x-5' : ''
                      }`}
                    />
                  </button>
                </div>
              );
            })}
          </div>
        </div>
      )}

      <div className="rounded-xl border border-gray-800 bg-gray-900 p-6">
        <div className="flex items-center gap-2 text-gray-300">
          <Clock className="h-4 w-4" />
          <h2 className="text-sm font-semibold">Quiet Hours</h2>
        </div>
        <p className="mt-1 text-xs text-gray-500">Suppress non-critical notifications during these hours</p>
        <div className="mt-4 flex items-center gap-4">
          <div>
            <label className="mb-1 block text-xs text-gray-500">Start</label>
            <input
              type="time"
              value={quietHours.start}
              onChange={(e) => setQuietHours({ ...quietHours, start: e.target.value })}
              className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-200"
            />
          </div>
          <span className="mt-5 text-gray-600">to</span>
          <div>
            <label className="mb-1 block text-xs text-gray-500">End</label>
            <input
              type="time"
              value={quietHours.end}
              onChange={(e) => setQuietHours({ ...quietHours, end: e.target.value })}
              className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-200"
            />
          </div>
        </div>
      </div>
    </div>
  );
}
