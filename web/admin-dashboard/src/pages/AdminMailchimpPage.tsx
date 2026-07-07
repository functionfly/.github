/**
 * Admin Mailchimp Page
 * Manage Mailchimp integration, sync status, and audience.
 */

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { RefreshCw, Users, Mail, CheckCircle, AlertCircle, ExternalLink, Clock } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { useToastHelpers } from '@/components/ui/Toast';
import { LoadingScreen } from '@/components/common/LoadingScreen';

interface MailchimpStats {
  platform_subscribers: number;
  mailchimp_stats?: {
    member_count: number;
    unsubscribe_count: number;
    cleaned_count: number;
    open_rate: number;
    click_rate: number;
  };
  sync_queue_length: number;
  sync_enabled: boolean;
}

interface Subscriber {
  email: string;
  name: string;
  status: string;
  mailchimp_status?: string;
  mailchimp_sync_status?: string;
  subscribed_at: string;
  mailchimp_last_synced?: string;
  email_frequency?: string;
}

interface SubscriberListResponse {
  subscribers: Subscriber[];
  total: number;
  limit: number;
  offset: number;
}

export function AdminMailchimpPage() {
  const queryClient = useQueryClient();
  const toast = useToastHelpers();
  const [activeTab, setActiveTab] = useState<'overview' | 'subscribers' | 'sync'>('overview');

  // Fetch Mailchimp stats
  const { data: statsResponse, isLoading: statsLoading, error: statsError, refetch: refetchStats } = useQuery({
    queryKey: ['admin-mailchimp-stats'],
    queryFn: async () => {
      try {
        const res = await adminApiClient.get<MailchimpStats>('/admin/mailchimp/stats');
        return res;
      } catch (err) {
        console.error('Failed to fetch Mailchimp stats:', err);
        return { data: null };
      }
    },
  });

  // Fetch subscribers
  const { data: subscribersResponse, isLoading: subscribersLoading, refetch: refetchSubscribers } = useQuery({
    queryKey: ['admin-mailchimp-subscribers'],
    queryFn: async () => {
      try {
        const res = await adminApiClient.get<SubscriberListResponse>('/admin/mailchimp/subscribers', {
          params: { limit: 50, offset: 0 },
        });
        return res;
      } catch (err) {
        console.error('Failed to fetch subscribers:', err);
        return { data: { subscribers: [], total: 0, limit: 50, offset: 0 } };
      }
    },
    enabled: activeTab === 'subscribers',
  });

  // Trigger sync mutation
  const syncMutation = useMutation({
    mutationFn: (fullSync: boolean) =>
      adminApiClient.post('/admin/mailchimp/sync', { full_sync: fullSync }),
    onSuccess: (response: unknown) => {
      const res = response as { data?: { synced_count?: number; failed_count?: number } };
      queryClient.invalidateQueries({ queryKey: ['admin-mailchimp-stats'] });
      queryClient.invalidateQueries({ queryKey: ['admin-mailchimp-subscribers'] });
      toast.success(
        `Sync completed: ${res?.data?.synced_count ?? 0} synced, ${res?.data?.failed_count ?? 0} failed`
      );
    },
    onError: () => {
      toast.error('Failed to trigger sync');
    },
  });

  const stats = statsResponse?.data;
  const subscribers = subscribersResponse?.data?.subscribers ?? [];
  const totalSubscribers = subscribersResponse?.data?.total ?? 0;

  if (statsLoading && !stats) {
    return <LoadingScreen />;
  }

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Mailchimp Integration</h1>
          <p className="text-muted-foreground">Manage your Mailchimp audience sync and settings</p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => refetchStats()}
            disabled={statsLoading}
          >
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => window.open('https://mailchimp.com', '_blank')}
          >
            <ExternalLink className="w-4 h-4 mr-2" />
            Open Mailchimp
          </Button>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 border-b">
        <button
          className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
            activeTab === 'overview'
              ? 'border-primary text-primary'
              : 'border-transparent text-muted-foreground hover:text-foreground'
          }`}
          onClick={() => setActiveTab('overview')}
        >
          Overview
        </button>
        <button
          className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
            activeTab === 'subscribers'
              ? 'border-primary text-primary'
              : 'border-transparent text-muted-foreground hover:text-foreground'
          }`}
          onClick={() => setActiveTab('subscribers')}
        >
          Subscribers
        </button>
        <button
          className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
            activeTab === 'sync'
              ? 'border-primary text-primary'
              : 'border-transparent text-muted-foreground hover:text-foreground'
          }`}
          onClick={() => setActiveTab('sync')}
        >
          Sync Status
        </button>
      </div>

      {/* Overview Tab */}
      {activeTab === 'overview' && (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {/* Platform Subscribers */}
          <Card className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Platform Subscribers</p>
                <p className="text-2xl font-bold">{stats?.platform_subscribers ?? 0}</p>
              </div>
              <Users className="w-8 h-8 text-muted-foreground" />
            </div>
          </Card>

          {/* Mailchimp Members */}
          <Card className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Mailchimp Members</p>
                <p className="text-2xl font-bold">{stats?.mailchimp_stats?.member_count ?? 'N/A'}</p>
              </div>
              <Mail className="w-8 h-8 text-muted-foreground" />
            </div>
          </Card>

          {/* Sync Queue */}
          <Card className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Pending Sync</p>
                <p className="text-2xl font-bold">{stats?.sync_queue_length ?? 0}</p>
              </div>
              <Clock className="w-8 h-8 text-muted-foreground" />
            </div>
          </Card>

          {/* Sync Status */}
          <Card className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Sync Status</p>
                <div className="flex items-center gap-1">
                  {stats?.sync_enabled ? (
                    <>
                      <CheckCircle className="w-4 h-4 text-green-500" />
                      <span className="text-sm font-medium text-green-500">Enabled</span>
                    </>
                  ) : (
                    <>
                      <AlertCircle className="w-4 h-4 text-yellow-500" />
                      <span className="text-sm font-medium text-yellow-500">Disabled</span>
                    </>
                  )}
                </div>
              </div>
            </div>
          </Card>
        </div>
      )}

      {/* Subscribers Tab */}
      {activeTab === 'subscribers' && (
        <Card>
          <div className="p-4 border-b">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold">Subscribers ({totalSubscribers})</h2>
              <Button variant="outline" size="sm" onClick={() => refetchSubscribers()}>
                <RefreshCw className="w-4 h-4 mr-2" />
                Refresh
              </Button>
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b bg-muted/50">
                  <th className="px-4 py-3 text-left font-medium">Email</th>
                  <th className="px-4 py-3 text-left font-medium">Name</th>
                  <th className="px-4 py-3 text-left font-medium">Status</th>
                  <th className="px-4 py-3 text-left font-medium">Mailchimp</th>
                  <th className="px-4 py-3 text-left font-medium">Sync</th>
                  <th className="px-4 py-3 text-left font-medium">Frequency</th>
                  <th className="px-4 py-3 text-left font-medium">Subscribed</th>
                </tr>
              </thead>
              <tbody>
                {subscribersLoading ? (
                  <tr>
                    <td colSpan={7} className="px-4 py-8 text-center text-muted-foreground">
                      Loading...
                    </td>
                  </tr>
                ) : subscribers.length === 0 ? (
                  <tr>
                    <td colSpan={7} className="px-4 py-8 text-center text-muted-foreground">
                      No subscribers found
                    </td>
                  </tr>
                ) : (
                  subscribers.map((subscriber) => (
                    <tr key={subscriber.email} className="border-b">
                      <td className="px-4 py-3">{subscriber.email}</td>
                      <td className="px-4 py-3">{subscriber.name || '-'}</td>
                      <td className="px-4 py-3">
                        <span
                          className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
                            subscriber.status === 'active'
                              ? 'bg-green-100 text-green-800'
                              : subscriber.status === 'pending'
                              ? 'bg-yellow-100 text-yellow-800'
                              : 'bg-gray-100 text-gray-800'
                          }`}
                        >
                          {subscriber.status}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        {subscriber.mailchimp_status ? (
                          <span className="inline-flex items-center gap-1">
                            <CheckCircle className="w-3 h-3 text-green-500" />
                            {subscriber.mailchimp_status}
                          </span>
                        ) : (
                          <span className="text-muted-foreground">-</span>
                        )}
                      </td>
                      <td className="px-4 py-3">
                        <span
                          className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
                            subscriber.mailchimp_sync_status === 'synced'
                              ? 'bg-green-100 text-green-800'
                              : subscriber.mailchimp_sync_status === 'pending'
                              ? 'bg-yellow-100 text-yellow-800'
                              : subscriber.mailchimp_sync_status === 'failed'
                              ? 'bg-red-100 text-red-800'
                              : 'bg-gray-100 text-gray-800'
                          }`}
                        >
                          {subscriber.mailchimp_sync_status || 'none'}
                        </span>
                      </td>
                      <td className="px-4 py-3">{subscriber.email_frequency || 'weekly'}</td>
                      <td className="px-4 py-3">
                        {new Date(subscriber.subscribed_at).toLocaleDateString()}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {/* Sync Tab */}
      {activeTab === 'sync' && (
        <div className="space-y-4">
          <Card className="p-6">
            <h2 className="text-lg font-semibold mb-4">Manual Sync</h2>
            <p className="text-sm text-muted-foreground mb-4">
              Trigger a manual sync to push pending subscribers to Mailchimp. This will process up to 1000
              subscribers that have a pending or failed sync status.
            </p>
            <div className="flex gap-2">
              <Button
                variant="outline"
                onClick={() => syncMutation.mutate(false)}
                disabled={syncMutation.isPending}
              >
                <RefreshCw className="w-4 h-4 mr-2" />
                {syncMutation.isPending ? 'Syncing...' : 'Sync Pending Subscribers'}
              </Button>
              <Button
                variant="outline"
                onClick={() => syncMutation.mutate(true)}
                disabled={syncMutation.isPending}
              >
                <RefreshCw className="w-4 h-4 mr-2" />
                {syncMutation.isPending ? 'Syncing...' : 'Full Sync (All Active)'}
              </Button>
            </div>
          </Card>

          <Card className="p-6">
            <h2 className="text-lg font-semibold mb-4">Sync Configuration</h2>
            <dl className="space-y-2 text-sm">
              <div className="flex justify-between">
                <dt className="text-muted-foreground">Sync Enabled</dt>
                <dd>
                  {stats?.sync_enabled ? (
                    <span className="text-green-500">Yes</span>
                  ) : (
                    <span className="text-yellow-500">No</span>
                  )}
                </dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-muted-foreground">Pending Jobs</dt>
                <dd>{stats?.sync_queue_length ?? 0}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-muted-foreground">Mailchimp List ID</dt>
                <dd className="font-mono text-xs">Configured in environment</dd>
              </div>
            </dl>
          </Card>
        </div>
      )}

      {/* Mailchimp Stats (if available) */}
      {stats?.mailchimp_stats && (
        <Card className="p-6">
          <h2 className="text-lg font-semibold mb-4">Mailchimp Audience Stats</h2>
          <div className="grid gap-4 md:grid-cols-4">
            <div>
              <p className="text-sm text-muted-foreground">Total Members</p>
              <p className="text-xl font-bold">{stats.mailchimp_stats.member_count}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Unsubscribed</p>
              <p className="text-xl font-bold">{stats.mailchimp_stats.unsubscribe_count}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Cleaned (Bounced)</p>
              <p className="text-xl font-bold">{stats.mailchimp_stats.cleaned_count}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Open Rate</p>
              <p className="text-xl font-bold">
                {stats.mailchimp_stats.open_rate ? `${stats.mailchimp_stats.open_rate.toFixed(1)}%` : 'N/A'}
              </p>
            </div>
          </div>
        </Card>
      )}
    </div>
  );
}
