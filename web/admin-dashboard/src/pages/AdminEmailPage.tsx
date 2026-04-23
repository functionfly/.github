/**
 * Admin Email Page
 * Tabbed interface: Newsletter (subscribers & campaigns) and Email Settings
 */

import { useState, useEffect, useCallback } from 'react';
import { Mail, Settings, Send, Users, Plus, Trash2, Send as SendIcon, BarChart3, X } from 'lucide-react';
import { adminApiClient } from '@/lib/api/adminClient';

type TabId = 'newsletter' | 'settings';

interface Subscriber {
  id: string;
  email: string;
  name: string;
  status: string;
  source: string;
  subscribed_at: string;
}

interface NewsletterStats {
  total_subscribers: number;
  active_subscribers: number;
  unsubscribed: number;
  bounced: number;
  subscribers_last_30_days: number;
}

export function AdminEmailPage() {
  const [activeTab, setActiveTab] = useState<TabId>('newsletter');

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Email</h1>
        <p className="mt-2 text-gray-600 dark:text-gray-400">
          Manage newsletter subscribers, campaigns, and platform email settings.
        </p>
      </div>

      <div className="border-b border-gray-200 dark:border-gray-700">
        <nav className="flex gap-6">
          <button
            type="button"
            onClick={() => setActiveTab('newsletter')}
            className={`flex items-center gap-2 pb-3 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'newsletter'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-200 dark:hover:border-gray-600'
            }`}
          >
            <Send className="w-4 h-4" />
            Newsletter
          </button>
          <button
            type="button"
            onClick={() => setActiveTab('settings')}
            className={`flex items-center gap-2 pb-3 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'settings'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-200 dark:hover:border-gray-600'
            }`}
          >
            <Settings className="w-4 h-4" />
            Email Settings
          </button>
        </nav>
      </div>

      {activeTab === 'newsletter' && <NewsletterTab />}
      {activeTab === 'settings' && <EmailSettingsTab />}
    </div>
  );
}

function NewsletterTab() {
  const [subscribers, setSubscribers] = useState<Subscriber[]>([]);
  const [stats, setStats] = useState<NewsletterStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showAddModal, setShowAddModal] = useState(false);
  const [newEmail, setNewEmail] = useState('');
  const [newName, setNewName] = useState('');
  const [adding, setAdding] = useState(false);

  const fetchSubscribers = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await adminApiClient.get<{ subscribers: Subscriber[]; total: number }>(
        '/newsletter/subscribers?limit=50'
      );
      if (response.data) {
        setSubscribers(response.data.subscribers || []);
      }
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to fetch subscribers');
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchStats = useCallback(async () => {
    try {
      const response = await adminApiClient.get<NewsletterStats>('/newsletter/stats');
      if (response.data) {
        setStats(response.data);
      }
    } catch (err: any) {
      console.error('Failed to fetch newsletter stats:', err);
    }
  }, []);

  useEffect(() => {
    fetchSubscribers();
    fetchStats();
  }, [fetchSubscribers, fetchStats]);

  const handleAddSubscriber = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newEmail.trim()) return;

    try {
      setAdding(true);
      await adminApiClient.post('/newsletter/subscribers', {
        email: newEmail.trim(),
        name: newName.trim(),
      });
      setShowAddModal(false);
      setNewEmail('');
      setNewName('');
      fetchSubscribers();
      fetchStats();
    } catch (err: any) {
      alert(err.response?.data?.error || 'Failed to add subscriber');
    } finally {
      setAdding(false);
    }
  };

  const handleDeleteSubscriber = async (id: string) => {
    if (!confirm('Are you sure you want to delete this subscriber?')) return;

    try {
      await adminApiClient.delete(`/newsletter/subscribers/${id}`);
      fetchSubscribers();
      fetchStats();
    } catch (err: any) {
      alert(err.response?.data?.error || 'Failed to delete subscriber');
    }
  };

  return (
    <div className="space-y-6">
      {/* Stats Cards */}
      {stats && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <StatCard
            label="Total Subscribers"
            value={stats.total_subscribers}
            icon={<Users className="w-5 h-5" />}
            color="blue"
          />
          <StatCard
            label="Active"
            value={stats.active_subscribers}
            icon={<Users className="w-5 h-5" />}
            color="green"
          />
          <StatCard
            label="Unsubscribed"
            value={stats.unsubscribed}
            icon={<X className="w-5 h-5" />}
            color="gray"
          />
          <StatCard
            label="New (30 days)"
            value={stats.subscribers_last_30_days}
            icon={<BarChart3 className="w-5 h-5" />}
            color="purple"
          />
        </div>
      )}

      {/* Subscribers List */}
      <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Mail className="w-5 h-5 text-indigo-600" />
            <h2 className="text-lg font-semibold text-gray-900 dark:text-white">Subscribers</h2>
          </div>
          <button
            onClick={() => setShowAddModal(true)}
            className="flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm font-medium hover:bg-indigo-700 dark:bg-indigo-700 dark:hover:bg-indigo-800"
          >
            <Plus className="w-4 h-4" />
            Add Subscriber
          </button>
        </div>

        {loading ? (
          <div className="p-8 text-center text-gray-500 dark:text-gray-400">Loading subscribers...</div>
        ) : error ? (
          <div className="p-8 text-center text-red-500 dark:text-red-400">{error}</div>
        ) : subscribers.length === 0 ? (
          <div className="p-8 text-center text-gray-500 dark:text-gray-400">
            No subscribers yet. Add your first subscriber to get started.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50 dark:bg-gray-900">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Email
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Name
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Source
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Subscribed
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                {subscribers.map((subscriber) => (
                  <tr key={subscriber.id} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">
                      {subscriber.email}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                      {subscriber.name || '-'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span
                        className={`inline-flex px-2 py-1 text-xs font-medium rounded-full ${
                          subscriber.status === 'active'
                            ? 'bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-400'
                            : subscriber.status === 'unsubscribed'
                            ? 'bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-400'
                            : 'bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-400'
                        }`}
                      >
                        {subscriber.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                      {subscriber.source}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                      {new Date(subscriber.subscribed_at).toLocaleDateString()}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm">
                      <button
                        onClick={() => handleDeleteSubscriber(subscriber.id)}
                        className="text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300"
                        title="Delete subscriber"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Add Subscriber Modal */}
      {showAddModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-md w-full mx-4">
            <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Add Subscriber</h3>
              <button onClick={() => setShowAddModal(false)} className="text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300">
                <X className="w-5 h-5" />
              </button>
            </div>
            <form onSubmit={handleAddSubscriber} className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Email</label>
                <input
                  type="email"
                  value={newEmail}
                  onChange={(e) => setNewEmail(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
                  placeholder="subscriber@example.com"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Name (optional)
                </label>
                <input
                  type="text"
                  value={newName}
                  onChange={(e) => setNewName(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
                  placeholder="John Doe"
                />
              </div>
              <div className="flex justify-end gap-3 pt-4">
                <button
                  type="button"
                  onClick={() => setShowAddModal(false)}
                  className="px-4 py-2 text-sm font-medium text-gray-700 hover:text-gray-900 dark:text-gray-300 dark:hover:text-white"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={adding}
                  className="px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm font-medium hover:bg-indigo-700 disabled:opacity-50 dark:bg-indigo-700 dark:hover:bg-indigo-800"
                >
                  {adding ? 'Adding...' : 'Add Subscriber'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

function StatCard({
  label,
  value,
  icon,
  color,
}: {
  label: string;
  value: number;
  icon: React.ReactNode;
  color: 'blue' | 'green' | 'gray' | 'purple';
}) {
  const colorClasses = {
    blue: 'bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400',
    green: 'bg-green-50 dark:bg-green-900/30 text-green-600 dark:text-green-400',
    gray: 'bg-gray-50 dark:bg-gray-700 text-gray-600 dark:text-gray-400',
    purple: 'bg-purple-50 dark:bg-purple-900/30 text-purple-600 dark:text-purple-400',
  };

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm text-gray-500 dark:text-gray-400">{label}</p>
          <p className="text-2xl font-bold text-gray-900 dark:text-white mt-1">{value}</p>
        </div>
        <div className={`p-3 rounded-lg ${colorClasses[color]}`}>{icon}</div>
      </div>
    </div>
  );
}

function EmailSettingsTab() {
  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6 max-w-2xl">
      <div className="flex items-center gap-2 mb-6">
        <Settings className="w-5 h-5 text-indigo-600" />
        <h2 className="text-lg font-semibold text-gray-900 dark:text-white">Email Settings</h2>
      </div>
      <p className="text-gray-600 dark:text-gray-400 mb-6">
        Configure how the platform sends transactional and marketing email. Some options are set via server environment variables.
      </p>

      <div className="space-y-6">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">From address</label>
          <input
            type="email"
            placeholder="noreply@functionfly.dev"
            className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
            readOnly
            disabled
          />
          <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">Set via FROM_EMAIL on the server.</p>
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Email transport</label>
          <input
            type="text"
            placeholder="Resend (RESEND_API_KEY) or SMTP_HOST"
            className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 bg-gray-50 dark:bg-gray-700"
            readOnly
            disabled
          />
          <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
            Production uses <strong className="font-medium">Resend</strong> when <code className="text-xs bg-gray-100 dark:bg-gray-700 px-1 rounded">RESEND_API_KEY</code> is set; otherwise{' '}
            <code className="text-xs bg-gray-100 dark:bg-gray-700 px-1 rounded">SMTP_*</code>. Webhooks:{' '}
            <code className="text-xs bg-gray-100 dark:bg-gray-700 px-1 rounded">RESEND_WEBHOOK_SECRET</code> for{' '}
            <code className="text-xs bg-gray-100 dark:bg-gray-700 px-1 rounded">/v1/webhooks/resend</code>.
          </p>
        </div>
        <div className="rounded-lg bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 p-4 text-sm text-amber-800 dark:text-amber-200">
          <strong>Note:</strong> To change Resend, SMTP, or from-address, update environment variables on the orchestrator API server and restart.
        </div>
      </div>
    </div>
  );
}
