import { NavLink, Outlet } from 'react-router-dom';
import {
  LayoutDashboard,
  User,
  FolderKanban,
  CheckSquare,
  GraduationCap,
  BookOpen,
  DollarSign,
  Network,
  Users,
  Bot,
  Target,
  Clock,
  BarChart3,
  Bell,
  LogOut,
  Lightbulb,
  Store,
  TrendingUp,
  HeartHandshake,
  FileText,
  Gauge,
  HeartPulse,
  GitBranch,
  Trophy,
  Award,
  Brain,
  Siren,
  GitPullRequest,
  Flag,
  ShieldCheck,
  KeyRound,
  Radio,
  Mail,
  Monitor,
  Key,
  Wallet,
  BellRing,
  MessageCircle,
  GitMerge,
  PenLine,
  Lock,
  Upload,
  Package,
} from 'lucide-react';
import { useAuthStore } from '@/stores/authStore';

const navItems = [
  { to: '/', icon: LayoutDashboard, label: 'Dashboard' },
  { to: '/profile', icon: User, label: 'Profile' },
  { to: '/projects', icon: FolderKanban, label: 'Projects' },
  { to: '/tasks', icon: CheckSquare, label: 'Tasks' },
  { to: '/learning', icon: GraduationCap, label: 'Learning' },
  { to: '/knowledge', icon: BookOpen, label: 'Knowledge' },
  { to: '/compensation', icon: DollarSign, label: 'Compensation' },
  { to: '/orgchart', icon: Network, label: 'Org Chart' },
  { to: '/team', icon: Users, label: 'Team' },
  { to: '/ai-assistant', icon: Bot, label: 'AI Assistant' },
  { to: '/performance', icon: Target, label: 'Performance' },
  { to: '/time-tracking', icon: Clock, label: 'Time Tracking' },
  { to: '/analytics', icon: BarChart3, label: 'Analytics' },
  { to: '/innovation', icon: Lightbulb, label: 'Innovation' },
  { to: '/marketplace', icon: Store, label: 'Marketplace' },
  { to: '/career', icon: TrendingUp, label: 'Career' },
  { to: '/mentorship', icon: HeartHandshake, label: 'Mentorship' },
  { to: '/documents', icon: FileText, label: 'Documents' },
  { to: '/mission-control', icon: Gauge, label: 'Mission Control' },
  { to: '/team-health', icon: HeartPulse, label: 'Team Health' },
  { to: '/skills-graph', icon: GitBranch, label: 'Skills Graph' },
  { to: '/reputation', icon: Trophy, label: 'Reputation' },
  { to: '/badges', icon: Award, label: 'Badges' },
  { to: '/memory', icon: Brain, label: 'Memory' },
  { to: '/incidents', icon: Siren, label: 'Incidents' },
  { to: '/lifecycle', icon: GitPullRequest, label: 'Lifecycle' },
  { to: '/feature-flags', icon: Flag, label: 'Feature Flags' },
  { to: '/data-classification', icon: ShieldCheck, label: 'Classification' },
  { to: '/certificates', icon: KeyRound, label: 'Certificates' },
  { to: '/events', icon: Radio, label: 'Events' },
  { to: '/email', icon: Mail, label: 'Email' },
  { to: '/devices', icon: Monitor, label: 'Devices' },
  { to: '/sso-provisioning', icon: Key, label: 'SSO' },
  { to: '/wallet-badge', icon: Wallet, label: 'Wallet Badge' },
  { to: '/notification-settings', icon: BellRing, label: 'Notifications' },
  { to: '/feedback-rounds', icon: MessageCircle, label: 'Feedback Rounds' },
  { to: '/goal-cascade', icon: GitMerge, label: 'Goal Cascade' },
  { to: '/signatures', icon: PenLine, label: 'Signatures' },
  { to: '/certificate-pki', icon: Lock, label: 'Certificate PKI' },
  { to: '/org-import', icon: Upload, label: 'Org Import' },
  { to: '/packages', icon: Package, label: 'Packages' },
  { to: '/passport', icon: BookOpen, label: 'Passport' },
  { to: '/wallet', icon: Wallet, label: 'Wallet' },
  { to: '/career-timeline', icon: Clock, label: 'Timeline' },
];

export function EmployeeLayout() {
  const { user, logout } = useAuthStore();

  return (
    <div className="flex h-screen bg-gray-950 text-gray-100">
      {/* Sidebar */}
      <aside className="flex w-64 flex-col border-r border-gray-800 bg-gray-900">
        <div className="flex h-16 items-center gap-2 border-b border-gray-800 px-4">
          <div className="flex h-8 w-8 items-center justify-center rounded bg-blue-600 font-bold text-white">
            F
          </div>
          <span className="text-lg font-semibold">FWOS</span>
        </div>

        <nav className="flex-1 overflow-y-auto px-3 py-4">
          <ul className="space-y-1">
            {navItems.map((item) => (
              <li key={item.to}>
                <NavLink
                  to={item.to}
                  end={item.to === '/'}
                  className={({ isActive }) =>
                    `flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
                      isActive
                        ? 'bg-blue-600/20 text-blue-400'
                        : 'text-gray-400 hover:bg-gray-800 hover:text-gray-200'
                    }`
                  }
                >
                  <item.icon className="h-4 w-4" />
                  {item.label}
                </NavLink>
              </li>
            ))}
          </ul>
        </nav>

        <div className="border-t border-gray-800 p-4">
          <button
            onClick={logout}
            className="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm text-gray-400 hover:bg-gray-800 hover:text-gray-200"
          >
            <LogOut className="h-4 w-4" />
            Sign Out
          </button>
        </div>
      </aside>

      {/* Main content */}
      <div className="flex flex-1 flex-col overflow-hidden">
        {/* Top bar */}
        <header className="flex h-16 items-center justify-between border-b border-gray-800 bg-gray-900 px-6">
          <div />
          <div className="flex items-center gap-4">
            <button className="relative rounded-lg p-2 text-gray-400 hover:bg-gray-800 hover:text-gray-200">
              <Bell className="h-5 w-5" />
              <span className="absolute right-1 top-1 h-2 w-2 rounded-full bg-blue-500" />
            </button>
            <div className="flex items-center gap-3">
              <div className="h-8 w-8 rounded-full bg-gray-700">
                {user?.avatar_url && (
                  <img
                    src={user.avatar_url}
                    alt={user.name}
                    className="h-full w-full rounded-full object-cover"
                  />
                )}
              </div>
              <span className="text-sm font-medium">{user?.name || 'Employee'}</span>
            </div>
          </div>
        </header>

        {/* Page content */}
        <main className="flex-1 overflow-y-auto p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
