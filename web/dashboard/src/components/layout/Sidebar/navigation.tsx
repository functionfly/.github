/**
 * Sidebar navigation data and configuration.
 *
 * All hardcoded nav labels and routes — no user-controlled content flows here.
 * HTML escaping in translateLabel provides defense-in-depth against future XSS.
 */

import {
  Activity,
  Award,
  BarChart3,
  BadgeCheck,
  Bell,
  Bot,
  Brain,
  Building2,
  CheckCircle,
  Cloud,
  Code,
  CreditCard,
  Database,
  Dna,
  Flame,
  History,
  Key,
  KeyRound,
  LayoutGrid,
  LifeBuoy,
  Link2,
  MessageSquare,
  Network,
  Package,
  PieChart,
  Plug,
  Rocket,
  Search,
  Settings,
  Shield,
  Sparkles,
  Star,
  TrendingUp,
  Users,
  Wallet,
  Wand,
  Workflow,
  type LucideIcon,
} from 'lucide-react';
import { ROUTES } from '@/lib/constants';

// ============================================================================
// Types
// ============================================================================

export interface NavItem {
  path: string;
  label: string;
  icon: LucideIcon | React.ComponentType;
  badge?: 'new' | 'beta' | number;
  shortcut?: string;
  description?: string;
  onboardingHint?: string;
}

export interface NavSection {
  id: string;
  title: string;
  icon: LucideIcon;
  items: NavItem[];
  collapsible?: boolean;
}

// Props shared across sidebar entry points
export interface SidebarProps {
  isOpen: boolean;
  onClose: () => void;
}

// ============================================================================
// GitHub Icon (custom SVG component)
// ============================================================================

export const GitHubIcon = () => (
  <svg role="img" viewBox="0 0 24 24" className="w-5 h-5" xmlns="http://www.w3.org/2000/svg">
    <path
      fill="currentColor"
      d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"
    />
  </svg>
);

// ============================================================================
// Navigation Sections
// ============================================================================

export const navigationSections: NavSection[] = [
  {
    id: 'discover',
    title: 'Discover',
    icon: Search,
    collapsible: true,
    items: [
      {
        path: ROUTES.DISCOVER,
        label: 'Discover',
        icon: Code,
        shortcut: 'G',
        description: 'Browse functions and marketplace',
        onboardingHint: 'Start here to explore functions',
      },
      {
        path: '/functions/hot',
        label: 'Hot',
        icon: Flame,
        shortcut: 'H',
        description: 'Trending hot functions right now',
      },
      {
        path: '/functions/trending',
        label: 'Trending',
        icon: TrendingUp,
        shortcut: 'T',
        description: 'Functions gaining popularity',
      },
      {
        path: '/functions/explore/new',
        label: 'New',
        icon: Sparkles,
        shortcut: 'N',
        description: 'Recently added functions',
      },
      {
        path: '/functions/favorites',
        label: 'Favorites',
        icon: Star,
        description: 'Your starred functions',
      },
      {
        path: '/notifications',
        label: 'Notifications',
        icon: Bell,
        description: 'View your notifications',
      },
      {
        path: ROUTES.CONVERSATIONS,
        label: 'Conversations',
        icon: MessageSquare,
        shortcut: 'C',
        description: 'Chat and message history',
      },
    ],
  },
  {
    id: 'build',
    title: 'Build',
    icon: Workflow,
    collapsible: true,
    items: [
      {
        path: '/functions/my',
        label: 'My Functions',
        icon: Code,
        description: 'Functions you created',
        onboardingHint: 'Create your first function here',
      },
      {
        path: '/ai/composer',
        label: 'AI Composer',
        icon: Sparkles,
        badge: 'new',
        description: 'AI-powered function generation',
      },
      {
        path: '/studio',
        label: 'Studio',
        icon: Wand,
        badge: 'beta',
        shortcut: 'S',
        description: 'AI-powered code & function studio',
      },
      {
        path: '/frg',
        label: 'Graph Editor',
        icon: Network,
        badge: 'beta',
        shortcut: 'R',
        description: 'Visual function graph editor',
      },
      {
        path: ROUTES.MCP,
        label: 'MCP',
        icon: Plug,
        shortcut: 'M',
        description: 'Model Context Protocol integrations',
      },
      {
        path: ROUTES.STATE,
        label: 'State',
        icon: Database,
        description: 'Function state management',
      },
      {
        path: '/github',
        label: 'GitHub Import',
        icon: GitHubIcon,
        badge: 'new',
        description: 'Import functions from GitHub repositories',
      },
      {
        path: '/functions/paste',
        label: 'Paste Code',
        icon: Code,
        description: 'Paste and import code snippets',
      },
    ],
  },
  {
    id: 'deploy',
    title: 'Deploy',
    icon: Rocket,
    collapsible: true,
    items: [
      {
        path: ROUTES.APPS,
        label: 'Apps',
        icon: Building2,
        description: 'Your applications',
        onboardingHint: 'Deploy your first app',
      },
      {
        path: '/agents',
        label: 'Agents',
        icon: Bot,
        shortcut: 'A',
        badge: 'new',
        description: 'Manage AI agents',
      },
      {
        path: ROUTES.PROVIDERS,
        label: 'Providers',
        icon: Cloud,
        shortcut: 'P',
        description: 'Cloud providers',
      },
      {
        path: ROUTES.SDK_INTEGRATIONS,
        label: 'SDK',
        icon: LayoutGrid,
        description: 'SDK integrations',
      },
      {
        path: ROUTES.VAULT,
        label: 'Vault',
        icon: Key,
        description: 'Secure secret storage',
      },
      {
        path: ROUTES.API_KEYS,
        label: 'API Keys',
        icon: KeyRound,
        description: 'API key management',
      },
      {
        path: '/bundles',
        label: 'Bundles',
        icon: LayoutGrid,
        badge: 'new',
        description: 'Backend-in-a-Box pricing bundles',
      },
      {
        path: '/bundles/mine',
        label: 'My Bundles',
        icon: Package,
        description: 'Your deployed bundles and billing',
      },
    ],
  },
  {
    id: 'operate',
    title: 'Operate',
    icon: Activity,
    collapsible: true,
    items: [
      {
        path: ROUTES.ANALYTICS,
        label: 'Analytics',
        icon: BarChart3,
        shortcut: 'Y',
        description: 'Performance analytics',
      },
      {
        path: '/agent-observability',
        label: 'Observability',
        icon: Activity,
        description: 'Agent traces, metrics, and debugging',
      },
      {
        path: ROUTES.USAGE,
        label: 'Usage',
        icon: PieChart,
        description: 'Resource usage & cost analytics',
      },
      {
        path: '/brain',
        label: 'Brain',
        icon: Brain,
        badge: 'new',
        description: 'AI memory that learns from your connected accounts',
      },
      {
        path: '/wallet',
        label: 'Wallet',
        icon: Wallet,
        description: 'Platform wallet & credits',
      },
      {
        path: '/status',
        label: 'Status',
        icon: Activity,
        description: 'System status',
      },
      {
        path: '/dna/overview',
        label: 'Function DNA',
        icon: Dna,
        badge: 'new',
        description: 'Living code that evolves itself',
      },
      {
        path: '/time-machine',
        label: 'Time Machine',
        icon: History,
        badge: 'new',
        description: 'Rewind and fix production bugs',
      },
      {
        path: '/certification',
        label: 'Certification',
        icon: Award,
        badge: 'new',
        description: 'Earn verifiable developer credentials',
      },
      {
        path: '/credentials',
        label: 'My Credentials',
        icon: BadgeCheck,
        description: 'View earned certifications',
      },
    ],
  },
  {
    id: 'advanced',
    title: 'Advanced',
    icon: Shield,
    collapsible: true,
    items: [
      {
        path: ROUTES.EVOLUTION,
        label: 'Evolution',
        icon: Sparkles,
        badge: 'beta',
        description: 'Agent evolution tracking',
      },
      {
        path: ROUTES.AGENT_MEMORIES,
        label: 'Memory',
        icon: Database,
        description: 'Agent memory and context storage',
      },
      {
        path: ROUTES.TEAMS,
        label: 'Teams',
        icon: Users,
        shortcut: 'M',
        description: 'Manage your teams',
      },
      {
        path: ROUTES.DECISIONS,
        label: 'Decisions',
        icon: CheckCircle,
        badge: 'new',
        description: 'Team decision recorder',
      },
      {
        path: ROUTES.STATE_FABRIC,
        label: 'State Fabric',
        icon: Network,
        badge: 'beta',
        description: 'Distributed state management',
      },
    ],
  },
  {
    id: 'account',
    title: 'Account',
    icon: Settings,
    collapsible: false,
    items: [
      {
        path: ROUTES.SETTINGS,
        label: 'Settings',
        icon: Settings,
        shortcut: 'S',
        description: 'Account settings',
      },
      {
        path: '/billing',
        label: 'Billing',
        icon: CreditCard,
        description: 'Manage subscription, invoices, and payments',
      },
      {
        path: ROUTES.FOUNDERS,
        label: 'Founders',
        icon: Sparkles,
        shortcut: 'F',
        description: 'DAO governance console — vote, track platform state, earn',
      },
      {
        path: ROUTES.ENTERPRISE_SUPPORT,
        label: 'Support',
        icon: LifeBuoy,
        description: 'Help and support center',
      },
    ],
  },
];

// ============================================================================
// Translation keys map
// All labels are hardcoded here (not user-controlled) so XSS risk is LOW.
// We still escape output as defense-in-depth.
// ============================================================================

export const NAV_LABEL_KEYS: Record<string, string> = {
  'Discover': 'nav.discover',
  'Hot': 'nav.hot',
  'Trending': 'nav.trending',
  'New': 'nav.new',
  'Notifications': 'nav.notifications',
  'Conversations': 'nav.conversations',
  'My Functions': 'nav.myFunctions',
  'AI Composer': 'nav.aiComposer',
  'Graph Editor': 'nav.graphEditor',
  'MCP': 'nav.mcp',
  'State': 'nav.state',
  'Apps': 'nav.apps',
  'Agents': 'nav.agents',
  'Providers': 'nav.providers',
  'SDK': 'nav.sdk',
  'Secrets': 'nav.secrets',
  'API Keys': 'nav.apiKeys',
  'Bundles': 'nav.bundles',
  'Analytics': 'nav.analytics',
  'Observability': 'nav.observability',
  'Usage': 'nav.usage',
  'Wallet': 'nav.wallet',
  'Status': 'nav.status',
  'Time Machine': 'nav.timeMachine',
  'Evolution': 'nav.evolution',
  'Marketplace': 'nav.marketplace',
  'Brain': 'nav.brain',
  'Memory': 'nav.memory',
  'Teams': 'nav.teams',
  'Decisions': 'nav.decisions',
  'State Fabric': 'nav.stateFabric',
  'GitHub Import': 'nav.githubImport',
  'Settings': 'nav.settings',
  'Billing': 'nav.billing',
  'Founders': 'nav.founders',
  'Support': 'nav.support',
  'Recent': 'nav.recent',
  'Getting Started': 'nav.gettingStarted',
  'Sign Out': 'nav.signOut',
  'Search Results': 'nav.search',
  'Favorites': 'nav.favorites',
};

// ============================================================================
// Animation variants (shared across sidebar)
// ============================================================================

export const SECTION_VARIANTS = {
  collapsed: { height: 0, opacity: 0 },
  expanded: { height: 'auto' as const, opacity: 1 },
};

export const LG_BREAKPOINT = 1024;