/**
 * Centralized breadcrumb registry.
 *
 * Maps route prefixes → breadcrumb label + icon.
 * Both Breadcrumb.tsx and PageHeader.tsx derive from this so they stay in sync.
 *
 * Covers all sidebar nav routes plus commonly-linked pages.
 * Entity-level routes (e.g. /functions/:id) get a dynamic last crumb from the parent.
 */

import {
  Activity,
  BarChart3,
  Bot,
  Brain,
  Building2,
  CheckCircle,
  CreditCard,
  Database,
  Dna,
  Home,
  History,
  LayoutDashboard,
  LayoutGrid,
  LifeBuoy,
  MessageSquare,
  Network,
  Plug,
  Plus,
  Settings,
  Shield,
  Sparkles,
  Star,
  TrendingUp,
  Users,
  Wallet,
  Wand,
  type LucideIcon,
} from 'lucide-react';
import { ROUTES } from './constants';

export interface BreadcrumbEntry {
  label: string;
  path?: string;
  icon?: LucideIcon;
}

/**
 * Route prefix → breadcrumb config.
 * Order matters: more-specific prefixes must come before less-specific ones.
 */
export const BREADCRUMB_MAP: Record<string, BreadcrumbEntry> = {
  // ─── Top-level / home ──────────────────────────────────────────────────────
  '/': { label: 'Home', icon: Home },
  '/dashboard': { label: 'Home', path: ROUTES.DASHBOARD, icon: Home },

  // ─── Discover section ──────────────────────────────────────────────────────
  '/functions/discovery': { label: 'Discover', path: ROUTES.DISCOVER, icon: LayoutGrid },
  '/functions/hot': { label: 'Hot', path: '/functions/hot', icon: Star },
  '/functions/trending': { label: 'Trending', path: '/functions/trending', icon: TrendingUp },
  '/functions/explore/new': { label: 'New', path: '/functions/explore/new', icon: Sparkles },
  '/functions/favorites': { label: 'Favorites', path: '/functions/favorites', icon: Star },
  '/notifications': { label: 'Notifications', path: '/notifications', icon: Activity },
  '/conversations': { label: 'Conversations', path: ROUTES.CONVERSATIONS, icon: MessageSquare },

  // ─── Build section ──────────────────────────────────────────────────────────
  '/functions/my': { label: 'My Functions', path: '/functions/my', icon: LayoutGrid },
  '/ai/composer': { label: 'AI Composer', path: '/ai/composer', icon: Sparkles },
  '/studio': { label: 'Studio', path: ROUTES.STUDIO, icon: Wand },
  '/frg': { label: 'Graph Editor', path: ROUTES.FRG, icon: Network },
  '/state': { label: 'State', path: ROUTES.STATE, icon: Database },
  '/github': { label: 'GitHub Import', path: '/github', icon: LayoutGrid },
  '/functions/paste': { label: 'Paste Code', path: '/functions/paste', icon: LayoutGrid },

  // ─── Deploy section ─────────────────────────────────────────────────────────
  '/apps': { label: 'Apps', path: ROUTES.APPS, icon: Building2 },
  '/agents': { label: 'Agents', path: ROUTES.AGENT_LIST, icon: Bot },
  '/providers': { label: 'Providers', path: ROUTES.PROVIDERS, icon: LayoutGrid },
  '/sdk-integrations': { label: 'SDK', path: ROUTES.SDK_INTEGRATIONS, icon: LayoutGrid },
  '/vault': { label: 'Vault', path: ROUTES.VAULT, icon: Shield },
  '/api-keys': { label: 'API Keys', path: ROUTES.API_KEYS, icon: Shield },
  '/bundles': { label: 'Bundles', path: '/bundles', icon: LayoutGrid },

  // ─── Operate section ────────────────────────────────────────────────────────
  '/analytics': { label: 'Analytics', path: ROUTES.ANALYTICS, icon: BarChart3 },
  '/agent-observability': { label: 'Observability', path: ROUTES.AGENT_OBSERVABILITY, icon: Activity },
  '/usage': { label: 'Usage', path: ROUTES.USAGE, icon: BarChart3 },
  '/brain': { label: 'Brain', path: '/brain', icon: Brain },
  '/wallet': { label: 'Wallet', path: '/wallet', icon: Wallet },
  '/status': { label: 'Status', path: '/status', icon: Activity },
  '/dna/overview': { label: 'Function DNA', path: '/dna/overview', icon: Dna },
  '/time-machine': { label: 'Time Machine', path: ROUTES.TIME_MACHINE, icon: History },
  '/certification': { label: 'Certification', path: '/certification', icon: CheckCircle },
  '/credentials': { label: 'My Credentials', path: '/credentials', icon: CheckCircle },

  // ─── Advanced section ───────────────────────────────────────────────────────
  '/evolution': { label: 'Evolution', path: ROUTES.EVOLUTION, icon: Sparkles },
  '/marketplace': { label: 'Marketplace', path: ROUTES.MARKETPLACE, icon: Shield },
  '/agent-memories': { label: 'Memory', path: ROUTES.AGENT_MEMORIES, icon: Database },
  '/teams': { label: 'Teams', path: ROUTES.TEAMS, icon: Users },
  '/decisions': { label: 'Decisions', path: ROUTES.DECISIONS, icon: CheckCircle },
  '/state-fabric': { label: 'State Fabric', path: ROUTES.STATE_FABRIC, icon: Network },

  // ─── Account section ────────────────────────────────────────────────────────
  '/settings': { label: 'Settings', path: ROUTES.SETTINGS, icon: Settings },
  '/billing': { label: 'Billing', path: '/billing', icon: CreditCard },
  '/founders': { label: 'Founders', path: ROUTES.FOUNDERS, icon: Sparkles },
  '/enterprise/support': { label: 'Support', path: ROUTES.ENTERPRISE_SUPPORT, icon: LifeBuoy },

  // ─── MCP ────────────────────────────────────────────────────────────────────
  '/mcp': { label: 'MCP', path: ROUTES.MCP, icon: Plug },

  // ─── Overview (special — goes to Dashboard) ────────────────────────────────
  '/overview': { label: 'Overview', path: ROUTES.OVERVIEW, icon: LayoutDashboard },
};

/**
 * Find the best BreadcrumbEntry for a given pathname.
 * Tries exact match first, then falls back to longest prefix match.
 */
export function getBreadcrumbForPath(pathname: string): BreadcrumbEntry | null {
  if (BREADCRUMB_MAP[pathname]) return BREADCRUMB_MAP[pathname];

  const segments = pathname.split('/').filter(Boolean);
  while (segments.length > 0) {
    const prefix = '/' + segments.join('/');
    if (BREADCRUMB_MAP[prefix]) return BREADCRUMB_MAP[prefix];
    segments.pop();
  }
  return null;
}

/**
 * Generate a flat crumb list for a given pathname.
 * Returns array from Home → parent → current page.
 * The final crumb's `path` is undefined (it's the active page).
 */
export function generateBreadcrumbs(pathname: string): BreadcrumbEntry[] {
  const segments = pathname.split('/').filter(Boolean);
  const crumbs: BreadcrumbEntry[] = [{ label: 'Home', path: ROUTES.DASHBOARD, icon: Home }];

  const pathSoFar = segments
    .map((_, i) => '/' + segments.slice(0, i + 1).join('/'))
    .filter(Boolean);

  for (const path of pathSoFar) {
    const entry = getBreadcrumbForPath(path);
    if (!entry) continue;

    const isLast = path === pathname;
    if (isLast) {
      crumbs.push({ label: entry.label, icon: entry.icon });
    } else {
      crumbs.push({ label: entry.label, path: entry.path ?? path, icon: entry.icon });
    }
  }

  return crumbs;
}
