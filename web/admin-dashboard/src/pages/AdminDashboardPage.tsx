/**
 * Admin Dashboard Page
 * Main dashboard with overview statistics, section cards, and activity
 */

import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { ROUTES } from '@/lib/constants';
import {
  TrendingUp,
  Users,
  Building2,
  Activity,
  CreditCard,
  Shield,
  Settings,
  FileText,
  Mail,
  Calendar,
  MessageSquare,
  BarChart3,
  CircleDot,
  AlertTriangle,
  Zap,
  PanelTop,
  RotateCcw,
  Landmark,
  BookOpen,
} from 'lucide-react';

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

  const sectionCards = [
    { title: 'Tenants', description: 'Manage multi-tenant organizations', path: ROUTES.ADMIN_TENANTS, icon: Building2, color: 'blue' },
    { title: 'Users', description: 'User management and permissions', path: ROUTES.ADMIN_USERS, icon: Users, color: 'green' },
    { title: 'Billing', description: 'Subscriptions and revenue', path: ROUTES.ADMIN_BILLING, icon: CreditCard, color: 'purple' },
    { title: 'Features', description: 'Tier-specific features', path: ROUTES.ADMIN_FEATURES, icon: Shield, color: 'amber' },
    { title: 'Audit Log', description: 'Security and system events', path: ROUTES.ADMIN_AUDIT, icon: FileText, color: 'red' },
    { title: 'System', description: 'Configuration and maintenance', path: ROUTES.ADMIN_SYSTEM, icon: Settings, color: 'slate' },
    { title: 'Backends', description: 'Platform backends', path: ROUTES.ADMIN_BACKENDS, icon: BarChart3, color: 'indigo' },
    { title: 'Providers', description: 'Provider management', path: ROUTES.ADMIN_PROVIDERS, icon: Zap, color: 'violet' },
    { title: 'Redirects', description: 'URL redirects', path: ROUTES.ADMIN_REDIRECTS, icon: RotateCcw, color: 'orange' },
    { title: 'Email', description: 'Newsletter, campaigns & email settings', path: ROUTES.ADMIN_EMAIL, icon: Mail, color: 'indigo' },
    { title: 'Content Calendar', description: 'Publication schedule', path: ROUTES.ADMIN_CONTENT_CALENDAR, icon: Calendar, color: 'orange' },
    { title: 'Content', description: 'Blog and content management', path: ROUTES.ADMIN_CONTENT, icon: PanelTop, color: 'teal' },
    { title: 'Blog', description: 'Posts, categories, settings & analytics', path: ROUTES.ADMIN_BLOG, icon: BookOpen, color: 'teal' },
    { title: 'Feedback', description: 'User feedback and tickets', path: ROUTES.ADMIN_FEEDBACK, icon: MessageSquare, color: 'pink' },
    { title: 'Functions', description: 'All functions across tenants', path: ROUTES.ADMIN_FUNCTIONS, icon: BarChart3, color: 'violet' },
    { title: 'Registry', description: 'Function registry moderation', path: ROUTES.ADMIN_REGISTRY, icon: FileText, color: 'cyan' },
    { title: 'State Fabric', description: 'State fabrics across tenants', path: ROUTES.ADMIN_STATE_FABRIC, icon: CircleDot, color: 'amber' },
    { title: 'Status', description: 'Platform health', path: ROUTES.ADMIN_STATUS, icon: Landmark, color: 'gray' },
    { title: 'Status Incidents', description: 'Incident management', path: ROUTES.ADMIN_STATUS_INCIDENTS, icon: AlertTriangle, color: 'red' },
    { title: 'Trust Dashboard', description: 'Trust and safety indicators', path: ROUTES.ADMIN_TRUST_DASHBOARD, icon: Shield, color: 'emerald' },
    { title: 'Execution Audit', description: 'Execution audit trail', path: ROUTES.ADMIN_EXECUTION_AUDIT, icon: BarChart3, color: 'blue' },
    { title: 'Fraud Detection', description: 'Fraud and bot detection', path: ROUTES.ADMIN_FRAUD_DETECTION, icon: Shield, color: 'rose' },
    { title: 'Economic Leaderboard', description: 'Revenue leaders', path: ROUTES.ADMIN_ECONOMIC_LEADERBOARD, icon: TrendingUp, color: 'violet' },
  ];

  const colorClasses: Record<string, string> = {
    blue: 'bg-blue-100 text-blue-600 border-blue-200 hover:border-blue-400',
    green: 'bg-green-100 text-green-600 border-green-200 hover:border-green-400',
    purple: 'bg-purple-100 text-purple-600 border-purple-200 hover:border-purple-400',
    amber: 'bg-amber-100 text-amber-600 border-amber-200 hover:border-amber-400',
    red: 'bg-red-100 text-red-600 border-red-200 hover:border-red-400',
    slate: 'bg-slate-100 text-slate-600 border-slate-200 hover:border-slate-400',
    indigo: 'bg-indigo-100 text-indigo-600 border-indigo-200 hover:border-indigo-400',
    violet: 'bg-violet-100 text-violet-600 border-violet-200 hover:border-violet-400',
    orange: 'bg-orange-100 text-orange-600 border-orange-200 hover:border-orange-400',
    teal: 'bg-teal-100 text-teal-600 border-teal-200 hover:border-teal-400',
    pink: 'bg-pink-100 text-pink-600 border-pink-200 hover:border-pink-400',
    cyan: 'bg-cyan-100 text-cyan-600 border-cyan-200 hover:border-cyan-400',
    emerald: 'bg-emerald-100 text-emerald-600 border-emerald-200 hover:border-emerald-400',
    rose: 'bg-rose-100 text-rose-600 border-rose-200 hover:border-rose-400',
    gray: 'bg-gray-100 text-gray-600 border-gray-200 hover:border-gray-400',
  };

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

      {/* Section Cards */}
      <div>
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Admin Sections</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
          {sectionCards.map((card) => {
            const Icon = card.icon;
            const colorClass = colorClasses[card.color] || colorClasses.slate;
            return (
              <Link
                key={card.path}
                to={card.path}
                className={`flex items-start gap-3 rounded-lg border-2 p-4 transition-colors ${colorClass}`}
              >
                <div className="flex-shrink-0 p-2 rounded-lg bg-white/80">
                  <Icon className="w-5 h-5" />
                </div>
                <div className="min-w-0 flex-1">
                  <p className="font-semibold text-gray-900">{card.title}</p>
                  <p className="text-sm text-gray-600 line-clamp-2">{card.description}</p>
                </div>
              </Link>
            );
          })}
        </div>
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
