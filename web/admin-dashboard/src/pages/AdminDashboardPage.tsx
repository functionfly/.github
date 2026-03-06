/**
 * Admin Dashboard Page
 * Main dashboard with overview statistics
 */

import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { TrendingUp, Users, Building2, Activity } from 'lucide-react';

interface QuickStats {
  total_tenants: number;
  total_users: number;
  active_sessions: number;
  pending_incidents: number;
}

interface DashboardActivity {
  timestamp: string;
  action: string;
  user_email: string;
  resource_type: string;
}

export function AdminDashboardPage() {
  const { data: statsData, isLoading: statsLoading } = useQuery({
    queryKey: ['admin-stats'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<QuickStats>('/dashboard/quick-stats');
      } catch {
        return {
          data: {
            total_tenants: 0,
            total_users: 0,
            active_sessions: 0,
            pending_incidents: 0,
          },
          success: false,
        };
      }
    },
    staleTime: 1000 * 60, // 1 minute
  });

  const { data: activityData, isLoading: activityLoading } = useQuery({
    queryKey: ['admin-activity'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<DashboardActivity[]>('/dashboard/activity');
      } catch {
        return { data: [], success: false };
      }
    },
    staleTime: 1000 * 30, // 30 seconds
  });

  if (statsLoading || activityLoading) {
    return <LoadingScreen />;
  }

  const stats = statsData?.data || {
    total_tenants: 0,
    total_users: 0,
    active_sessions: 0,
    pending_incidents: 0,
  };

  const activities = activityData?.data || [];

  return (
    <div className="space-y-8">
      {/* Page Header */}
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Admin Dashboard</h1>
        <p className="mt-2 text-gray-600">Welcome back! Here's an overview of your platform.</p>
      </div>

      {/* Quick Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          label="Total Tenants"
          value={stats.total_tenants}
          icon={Building2}
          color="blue"
        />
        <StatCard
          label="Total Users"
          value={stats.total_users}
          icon={Users}
          color="green"
        />
        <StatCard
          label="Active Sessions"
          value={stats.active_sessions}
          icon={Activity}
          color="purple"
        />
        <StatCard
          label="Pending Incidents"
          value={stats.pending_incidents}
          icon={TrendingUp}
          color="red"
        />
      </div>

      {/* Recent Activity */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">Recent Activity</h2>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-200 bg-gray-50">
                <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">
                  Action
                </th>
                <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">
                  User
                </th>
                <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">
                  Resource
                </th>
                <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">
                  Time
                </th>
              </tr>
            </thead>
            <tbody>
              {activities.length === 0 ? (
                <tr>
                  <td colSpan={4} className="px-6 py-8 text-center text-gray-500">
                    No recent activity
                  </td>
                </tr>
              ) : (
                activities.map((activity, idx) => (
                  <tr key={idx} className="border-b border-gray-100 hover:bg-gray-50">
                    <td className="px-6 py-4 text-sm text-gray-900">{activity.action}</td>
                    <td className="px-6 py-4 text-sm text-gray-600">{activity.user_email}</td>
                    <td className="px-6 py-4 text-sm text-gray-600">
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                        {activity.resource_type}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-500">
                      {new Date(activity.timestamp).toLocaleDateString()}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

interface StatCardProps {
  label: string;
  value: number;
  icon: React.ElementType;
  color: 'blue' | 'green' | 'purple' | 'red';
}

function StatCard({ label, value, icon: Icon, color }: StatCardProps) {
  const colorMap = {
    blue: 'bg-blue-100 text-blue-600',
    green: 'bg-green-100 text-green-600',
    purple: 'bg-purple-100 text-purple-600',
    red: 'bg-red-100 text-red-600',
  };

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm text-gray-600 font-medium">{label}</p>
          <p className="mt-2 text-3xl font-bold text-gray-900">{value.toLocaleString()}</p>
        </div>
        <div className={`${colorMap[color]} p-3 rounded-lg`}>
          <Icon className="w-6 h-6" />
        </div>
      </div>
    </div>
  );
}
