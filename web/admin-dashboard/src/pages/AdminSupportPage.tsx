import { adminApiClient } from '@/lib/api/adminClient';
import {
  Activity,
  AlertTriangle,
  CheckCircle,
  Clock,
  Headphones,
  MessageSquare,
  TrendingUp,
  Users,
} from 'lucide-react';
import { useEffect, useState } from 'react';

interface SupportMetrics {
  total_conversations: number;
  active_conversations: number;
  pending_conversations: number;
  resolved_conversations: number;
  escalated_conversations: number;
  emergency_requests: number;
  pending_emergencies: number;
  average_resolution_time: number;
  online_staff_count: number;
}

interface ConversationSummary {
  id: string;
  user_id: string;
  type: string;
  status: string;
  priority: string;
  title: string;
  ai_handled: boolean;
  is_emergency: boolean;
  created_at: string;
  updated_at: string;
  resolved_at?: string;
}

interface StaffStatus {
  staff_id: string;
  is_online: boolean;
  current_conversation_id?: string;
  last_seen: string;
}

export function AdminSupportPage() {
  const [metrics, setMetrics] = useState<SupportMetrics | null>(null);
  const [conversations, setConversations] = useState<ConversationSummary[]>([]);
  const [staffStatus, setStaffStatus] = useState<StaffStatus[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'overview' | 'conversations' | 'staff'>('overview');

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    setError(null);
    try {
      const [metricsRes, convRes, staffRes] = await Promise.all([
        adminApiClient.get<SupportMetrics>('/support/metrics'),
        adminApiClient.get<{ conversations: ConversationSummary[] }>('/support/conversations'),
        adminApiClient.get<{ staff: StaffStatus[] }>('/support/staff'),
      ]);
      setMetrics(metricsRes.data);
      setConversations(convRes.data.conversations || []);
      setStaffStatus(staffRes.data.staff || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load support data');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-6">
        <div className="flex items-center gap-3 text-red-800">
          <AlertTriangle className="w-5 h-5" />
          <span>{error}</span>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Headphones className="w-8 h-8 text-blue-600" />
          <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">Live Support System</h1>
        </div>
        <button
          onClick={loadData}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          <Activity className="w-4 h-4" />
          Refresh
        </button>
      </div>

      {/* Tab Navigation */}
      <div className="flex gap-4 border-b border-gray-200 dark:border-gray-700">
        <button
          onClick={() => setActiveTab('overview')}
          className={`pb-3 px-4 font-medium transition-colors ${
            activeTab === 'overview'
              ? 'border-b-2 border-blue-600 text-blue-600'
              : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'
          }`}
        >
          Overview
        </button>
        <button
          onClick={() => setActiveTab('conversations')}
          className={`pb-3 px-4 font-medium transition-colors ${
            activeTab === 'conversations'
              ? 'border-b-2 border-blue-600 text-blue-600'
              : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'
          }`}
        >
          Conversations ({conversations.length})
        </button>
        <button
          onClick={() => setActiveTab('staff')}
          className={`pb-3 px-4 font-medium transition-colors ${
            activeTab === 'staff'
              ? 'border-b-2 border-blue-600 text-blue-600'
              : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'
          }`}
        >
          Staff ({staffStatus.filter((s) => s.is_online).length} online)
        </button>
      </div>

      {/* Overview Tab */}
      {activeTab === 'overview' && metrics && (
        <>
          {/* Metrics Cards */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <MetricCard
              title="Total Conversations"
              value={metrics.total_conversations}
              icon={MessageSquare}
              color="blue"
            />
            <MetricCard
              title="Active Now"
              value={metrics.active_conversations}
              icon={Activity}
              color="green"
            />
            <MetricCard
              title="Pending"
              value={metrics.pending_conversations}
              icon={Clock}
              color="yellow"
            />
            <MetricCard
              title="Resolved"
              value={metrics.resolved_conversations}
              icon={CheckCircle}
              color="emerald"
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <MetricCard
              title="Escalated"
              value={metrics.escalated_conversations}
              icon={TrendingUp}
              color="purple"
            />
            <MetricCard
              title="Emergency Requests"
              value={metrics.emergency_requests}
              icon={AlertTriangle}
              color="red"
            />
            <MetricCard
              title="Pending Emergencies"
              value={metrics.pending_emergencies}
              icon={AlertTriangle}
              color="orange"
            />
            <MetricCard
              title="Online Staff"
              value={metrics.online_staff_count}
              icon={Users}
              color="indigo"
            />
          </div>

          {/* AI Performance */}
          <div className="bg-white dark:bg-gray-800 rounded-xl shadow-sm p-6">
            <h2 className="text-xl font-semibold mb-4 text-gray-900 dark:text-gray-100">AI Support Performance</h2>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div className="text-center p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <div className="text-3xl font-bold text-blue-600 dark:text-blue-400">
                  {metrics.resolved_conversations > 0
                    ? Math.round(
                        ((metrics.resolved_conversations - metrics.escalated_conversations) /
                          metrics.resolved_conversations) *
                          100
                      )
                    : 0}
                  %
                </div>
                <div className="text-sm text-gray-600 dark:text-gray-400 mt-1">AI Resolution Rate</div>
              </div>
              <div className="text-center p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <div className="text-3xl font-bold text-green-600 dark:text-green-400">
                  {metrics.average_resolution_time > 0
                    ? `${Math.round(metrics.average_resolution_time / 60)}m`
                    : 'N/A'}
                </div>
                <div className="text-sm text-gray-600 dark:text-gray-400 mt-1">Avg Resolution Time</div>
              </div>
              <div className="text-center p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <div className="text-3xl font-bold text-purple-600 dark:text-purple-400">
                  {metrics.emergency_requests > 0
                    ? Math.round((metrics.pending_emergencies / metrics.emergency_requests) * 100)
                    : 0}
                  %
                </div>
                <div className="text-sm text-gray-600 dark:text-gray-400 mt-1">Emergency Response Rate</div>
              </div>
            </div>
          </div>

          {/* How It Works */}
          <div className="bg-white dark:bg-gray-800 rounded-xl shadow-sm p-6">
            <h2 className="text-xl font-semibold mb-4 text-gray-900 dark:text-gray-100">How AI + Human Co-Pilot Works</h2>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div className="flex flex-col items-center text-center p-4">
                <div className="w-12 h-12 bg-blue-100 dark:bg-blue-900/30 rounded-full flex items-center justify-center mb-3">
                  <MessageSquare className="w-6 h-6 text-blue-600 dark:text-blue-400" />
                </div>
                <h3 className="font-semibold mb-2 text-gray-900 dark:text-gray-100">1. AI First Response</h3>
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  When users click "Get Live Help", AI automatically gathers context: function code,
                  logs, errors, and environment settings.
                </p>
              </div>
              <div className="flex flex-col items-center text-center p-4">
                <div className="w-12 h-12 bg-purple-100 dark:bg-purple-900/30 rounded-full flex items-center justify-center mb-3">
                  <TrendingUp className="w-6 h-6 text-purple-600 dark:text-purple-400" />
                </div>
                <h3 className="font-semibold mb-2 text-gray-900 dark:text-gray-100">2. Smart Escalation</h3>
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  AI handles common issues. Complex problems are escalated to human engineers when
                  confidence is low or user requests it.
                </p>
              </div>
              <div className="flex flex-col items-center text-center p-4">
                <div className="w-12 h-12 bg-green-100 dark:bg-green-900/30 rounded-full flex items-center justify-center mb-3">
                  <CheckCircle className="w-6 h-6 text-green-600 dark:text-green-400" />
                </div>
                <h3 className="font-semibold mb-2 text-gray-900 dark:text-gray-100">3. Learning Loop</h3>
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  Every solved issue improves AI. Staff fixes become automated solutions, reducing
                  future support burden.
                </p>
              </div>
            </div>
          </div>
        </>
      )}

      {/* Conversations Tab */}
      {activeTab === 'conversations' && (
        <div className="bg-white dark:bg-gray-800 rounded-xl shadow-sm overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Type
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Title
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Priority
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    AI Handled
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Created
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                {conversations.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="px-6 py-12 text-center text-gray-500 dark:text-gray-400">
                      No conversations yet
                    </td>
                  </tr>
                ) : (
                  conversations.map((conv) => (
                    <tr key={conv.id} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span
                          className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                            conv.type === 'support_emergency'
                              ? 'bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-400'
                              : conv.type === 'support_human'
                                ? 'bg-purple-100 dark:bg-purple-900/30 text-purple-800 dark:text-purple-400'
                                : 'bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-400'
                          }`}
                        >
                          {conv.type.replace('support_', '')}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <div className="text-sm font-medium text-gray-900 dark:text-gray-100">{conv.title}</div>
                        <div className="text-xs text-gray-500 dark:text-gray-400">ID: {conv.id.slice(0, 8)}...</div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span
                          className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                            conv.status === 'active'
                              ? 'bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-400'
                              : conv.status === 'pending'
                                ? 'bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-400'
                                : conv.status === 'resolved'
                                  ? 'bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-400'
                                  : 'bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-400'
                          }`}
                        >
                          {conv.status}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span
                          className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                            conv.priority === 'critical'
                              ? 'bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-400'
                              : conv.priority === 'high'
                                ? 'bg-orange-100 dark:bg-orange-900/30 text-orange-800 dark:text-orange-400'
                                : conv.priority === 'normal'
                                  ? 'bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-400'
                                  : 'bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-400'
                          }`}
                        >
                          {conv.priority}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {conv.ai_handled ? (
                          <span className="text-green-600 dark:text-green-400 text-sm">✓ AI Resolved</span>
                        ) : (
                          <span className="text-gray-500 dark:text-gray-400 text-sm">Human</span>
                        )}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                        {new Date(conv.created_at).toLocaleString()}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Staff Tab */}
      {activeTab === 'staff' && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {staffStatus.length === 0 ? (
            <div className="col-span-full bg-white dark:bg-gray-800 rounded-xl shadow-sm p-12 text-center text-gray-500 dark:text-gray-400">
              No staff members registered
            </div>
          ) : (
            staffStatus.map((staff) => (
              <div key={staff.staff_id} className="bg-white dark:bg-gray-800 rounded-xl shadow-sm p-6">
                <div className="flex items-center gap-4">
                  <div
                    className={`w-12 h-12 rounded-full flex items-center justify-center ${
                      staff.is_online ? 'bg-green-100 dark:bg-green-900/30' : 'bg-gray-100 dark:bg-gray-700'
                    }`}
                  >
                    <Users
                      className={`w-6 h-6 ${staff.is_online ? 'text-green-600 dark:text-green-400' : 'text-gray-400'}`}
                    />
                  </div>
                  <div className="flex-1">
                    <div className="font-medium text-gray-900 dark:text-gray-100">Staff Member</div>
                    <div className="text-sm text-gray-500 dark:text-gray-400">ID: {staff.staff_id.slice(0, 8)}...</div>
                  </div>
                  <div
                    className={`px-3 py-1 rounded-full text-xs font-medium ${
                      staff.is_online ? 'bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-400' : 'bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-400'
                    }`}
                  >
                    {staff.is_online ? 'Online' : 'Offline'}
                  </div>
                </div>
                {staff.is_online && staff.current_conversation_id && (
                  <div className="mt-4 pt-4 border-t border-gray-100 dark:border-gray-700">
                    <div className="text-sm text-gray-600 dark:text-gray-400">
                      Currently handling conversation:{' '}
                      <span className="font-mono text-xs">
                        {staff.current_conversation_id.slice(0, 8)}...
                      </span>
                    </div>
                  </div>
                )}
                <div className="mt-4 text-xs text-gray-400 dark:text-gray-500">
                  Last seen: {new Date(staff.last_seen).toLocaleString()}
                </div>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
}

interface MetricCardProps {
  title: string;
  value: number;
  icon: React.ComponentType<{ className?: string }>;
  color: 'blue' | 'green' | 'yellow' | 'emerald' | 'purple' | 'red' | 'orange' | 'indigo';
}

function MetricCard({ title, value, icon: Icon, color }: MetricCardProps) {
  const colorClasses = {
    blue: 'bg-blue-50 text-blue-600',
    green: 'bg-green-50 text-green-600',
    yellow: 'bg-yellow-50 text-yellow-600',
    emerald: 'bg-emerald-50 text-emerald-600',
    purple: 'bg-purple-50 text-purple-600',
    red: 'bg-red-50 text-red-600',
    orange: 'bg-orange-50 text-orange-600',
    indigo: 'bg-indigo-50 text-indigo-600',
  };

  return (
    <div className="bg-white rounded-xl shadow-sm p-6">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-gray-600">{title}</p>
          <p className="text-3xl font-bold mt-2">{value}</p>
        </div>
        <div className={`p-3 rounded-lg ${colorClasses[color]}`}>
          <Icon className="w-6 h-6" />
        </div>
      </div>
    </div>
  );
}
