/**
 * Admin Sidebar Component
 */

import { ROUTES } from '@/lib/constants';
import clsx from 'clsx';
import {
  Activity,
  AlertTriangle,
  BarChart3,
  BookOpen,
  Boxes,
  Building2,
  Calendar,
  CircleDot,
  Cloud,
  CreditCard,
  Factory,
  FileText,
  Landmark,
  LayoutDashboard,
  Mail,
  MessageSquare,
  PanelTop,
  RotateCcw,
  Settings,
  Shield,
  TrendingUp,
  Users,
  Wrench,
  X,
  Zap,
} from 'lucide-react';
import { Link, useLocation } from 'react-router-dom';

interface AdminSidebarProps {
  isOpen: boolean;
  onClose: () => void;
}

const NAV_ITEMS = [
  {
    label: 'Dashboard',
    path: ROUTES.ADMIN_DASHBOARD,
    icon: LayoutDashboard,
  },
  {
    label: 'Tenants',
    path: ROUTES.ADMIN_TENANTS,
    icon: Building2,
  },
  {
    label: 'Users',
    path: ROUTES.ADMIN_USERS,
    icon: Users,
  },
  {
    label: 'Signup Invites',
    path: ROUTES.ADMIN_SIGNUP_INVITES,
    icon: Shield,
  },
  {
    label: 'Billing',
    path: ROUTES.ADMIN_BILLING,
    icon: CreditCard,
  },
  {
    label: 'Audit Log',
    path: ROUTES.ADMIN_AUDIT,
    icon: FileText,
  },
  {
    label: 'System',
    path: ROUTES.ADMIN_SYSTEM,
    icon: Settings,
  },
  {
    label: 'Backends',
    path: ROUTES.ADMIN_BACKENDS,
    icon: Boxes,
  },
  {
    label: 'Providers',
    path: ROUTES.ADMIN_PROVIDERS,
    icon: Zap,
  },
  {
    label: 'Content',
    path: ROUTES.ADMIN_CONTENT,
    icon: PanelTop,
  },
  {
    label: 'Blog',
    path: ROUTES.ADMIN_BLOG,
    icon: BookOpen,
  },
  {
    label: 'Functions',
    path: ROUTES.ADMIN_FUNCTIONS,
    icon: BarChart3,
  },
  {
    label: 'Registry',
    path: ROUTES.ADMIN_REGISTRY,
    icon: FileText,
  },
  {
    label: 'State Fabric',
    path: ROUTES.ADMIN_STATE_FABRIC,
    icon: CircleDot,
  },
  {
    label: 'Feedback',
    path: ROUTES.ADMIN_FEEDBACK,
    icon: MessageSquare,
  },
  {
    label: 'Features',
    path: ROUTES.ADMIN_FEATURES,
    icon: Settings,
  },
  {
    label: 'Status',
    path: ROUTES.ADMIN_STATUS,
    icon: Landmark,
  },
  {
    label: 'Status Incidents',
    path: ROUTES.ADMIN_STATUS_INCIDENTS,
    icon: AlertTriangle,
  },
  {
    label: 'Redirects',
    path: ROUTES.ADMIN_REDIRECTS,
    icon: RotateCcw,
  },
  {
    label: 'Email',
    path: ROUTES.ADMIN_EMAIL,
    icon: Mail,
  },
  {
    label: 'Content Calendar',
    path: ROUTES.ADMIN_CONTENT_CALENDAR,
    icon: Calendar,
  },
  {
    label: 'Trust Dashboard',
    path: ROUTES.ADMIN_TRUST_DASHBOARD,
    icon: Shield,
  },
  {
    label: 'Execution Audit',
    path: ROUTES.ADMIN_EXECUTION_AUDIT,
    icon: BarChart3,
  },
  {
    label: 'Fraud Detection',
    path: ROUTES.ADMIN_FRAUD_DETECTION,
    icon: Shield,
  },
  {
    label: 'Economic Leaderboard',
    path: ROUTES.ADMIN_ECONOMIC_LEADERBOARD,
    icon: TrendingUp,
  },
  {
    label: 'Factory',
    path: ROUTES.ADMIN_FACTORY,
    icon: Factory,
  },
  {
    label: 'Maintenance',
    path: ROUTES.ADMIN_MAINTENANCE,
    icon: Wrench,
  },
  {
    label: 'Cache',
    path: ROUTES.ADMIN_CACHE,
    icon: Boxes,
  },
  {
    label: 'Monitoring',
    path: ROUTES.ADMIN_MONITORING,
    icon: Activity,
  },
  {
    label: 'Cloudflare',
    path: ROUTES.ADMIN_CLOUDFLARE_ANALYTICS,
    icon: Cloud,
  },
  {
    label: 'Support',
    path: ROUTES.ADMIN_SUPPORT,
    icon: MessageSquare,
  },
];

export function AdminSidebar({ isOpen, onClose }: AdminSidebarProps) {
  const location = useLocation();

  const isActive = (path: string) => {
    return location.pathname === path || location.pathname.startsWith(`${path}/`);
  };

  return (
    <>
      {/* Mobile overlay */}
      {isOpen && <div className="fixed inset-0 bg-black/50 z-40 md:hidden" onClick={onClose} />}

      {/* Sidebar */}
      <aside
        className={clsx(
          'fixed md:static left-0 top-0 h-screen w-64 bg-gray-900 text-white z-50 md:z-auto transition-transform duration-300 ease-in-out',
          isOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0'
        )}
      >
        <div className="flex flex-col h-full">
          {/* Logo */}
          <div className="flex items-center justify-between px-6 py-4 border-b border-gray-800">
            <h2 className="text-2xl font-bold">FlyAdmin</h2>
            <button onClick={onClose} className="md:hidden">
              <X className="w-5 h-5" />
            </button>
          </div>

          {/* Navigation */}
          <nav className="flex-1 overflow-y-auto px-4 py-8">
            <div className="space-y-2">
              {NAV_ITEMS.map((item) => {
                const Icon = item.icon;
                const active = isActive(item.path);

                return (
                  <Link
                    key={item.path}
                    to={item.path}
                    className={clsx(
                      'flex items-center gap-3 px-4 py-3 rounded-lg transition-colors duration-200',
                      active ? 'bg-blue-600 text-white' : 'text-gray-300 hover:bg-gray-800'
                    )}
                  >
                    <Icon className="w-5 h-5" />
                    <span className="font-medium">{item.label}</span>
                  </Link>
                );
              })}
            </div>
          </nav>

          {/* Footer */}
          <div className="px-6 py-4 border-t border-gray-800">
            <p className="text-xs text-gray-400">
              FunctionFly LLC · Admin © {new Date().getFullYear()}
            </p>
          </div>
        </div>
      </aside>
    </>
  );
}
