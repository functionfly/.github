import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { ROUTES } from '@/lib/constants';
import { cn } from '@/lib/utils';
import {
  BarChart3,
  Cloud,
  Code,
  Command,
  FunctionSquare,
  LayoutDashboard,
  Search,
  Settings,
} from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';

interface SearchResult {
  id: string;
  type: 'function' | 'provider' | 'page' | 'setting';
  title: string;
  description?: string;
  path: string;
  icon: React.ComponentType<{ className?: string }>;
  badge?: string;
}

interface GlobalSearchProps {
  isOpen: boolean;
  onClose: () => void;
}

const SEARCH_ITEMS: SearchResult[] = [
  // Pages
  {
    id: 'dashboard',
    type: 'page',
    title: 'Function Marketplace',
    description: 'Browse and deploy serverless functions from the registry',
    path: ROUTES.DASHBOARD,
    icon: Code,
  },
  {
    id: 'overview',
    type: 'page',
    title: 'Overview',
    description: 'Usage, activity, and health for your workspace',
    path: ROUTES.OVERVIEW,
    icon: LayoutDashboard,
  },
  {
    id: 'functions',
    type: 'page',
    title: 'Functions',
    description: 'Manage and deploy your serverless functions',
    path: ROUTES.FUNCTIONS,
    icon: FunctionSquare,
  },
  {
    id: 'providers',
    type: 'page',
    title: 'Providers',
    description: 'Configure cloud providers and regions',
    path: ROUTES.PROVIDERS,
    icon: Cloud,
  },
  {
    id: 'analytics',
    type: 'page',
    title: 'Analytics',
    description: 'View performance metrics and insights',
    path: ROUTES.ANALYTICS,
    icon: BarChart3,
  },
  {
    id: 'settings',
    type: 'page',
    title: 'Settings',
    description: 'Account and application settings',
    path: ROUTES.SETTINGS,
    icon: Settings,
  },
  // Example functions (in real app, this would come from API)
  {
    id: 'func-1',
    type: 'function',
    title: 'user-auth-api',
    description: 'Authentication service with JWT tokens',
    path: '/functions/user-auth-api',
    icon: FunctionSquare,
    badge: 'Active',
  },
  {
    id: 'func-2',
    type: 'function',
    title: 'payment-processor',
    description: 'Stripe payment processing webhook',
    path: '/functions/payment-processor',
    icon: FunctionSquare,
    badge: 'Active',
  },
  {
    id: 'func-3',
    type: 'function',
    title: 'email-service',
    description: 'SendGrid email notification service',
    path: '/functions/email-service',
    icon: FunctionSquare,
    badge: 'Failed',
  },
  // Providers
  {
    id: 'provider-vercel',
    type: 'provider',
    title: 'Vercel',
    description: 'Connected to Vercel with 3 regions',
    path: '/providers/vercel',
    icon: Cloud,
    badge: 'Connected',
  },
  {
    id: 'provider-cloudflare',
    type: 'provider',
    title: 'Cloudflare Workers',
    description: 'Connected to Cloudflare with 5 regions',
    path: '/providers/cloudflare',
    icon: Cloud,
    badge: 'Connected',
  },
];

const TYPE_LABELS = {
  page: 'Page',
  function: 'Function',
  provider: 'Provider',
  setting: 'Setting',
};

const TYPE_COLORS = {
  page: 'bg-blue-500/20 text-blue-400',
  function: 'bg-purple-500/20 text-purple-400',
  provider: 'bg-green-500/20 text-green-400',
  setting: 'bg-orange-500/20 text-orange-400',
};

export function GlobalSearch({ isOpen, onClose }: GlobalSearchProps) {
  const [query, setQuery] = useState('');
  const [selectedIndex, setSelectedIndex] = useState(0);

  const filteredResults = useMemo(() => {
    if (!query.trim()) return SEARCH_ITEMS.slice(0, 8); // Show recent/popular items when no query

    return SEARCH_ITEMS.filter(
      (item) =>
        item.title.toLowerCase().includes(query.toLowerCase()) ||
        (item.description && item.description.toLowerCase().includes(query.toLowerCase()))
    );
  }, [query]);

  // Reset selection when results change
  useEffect(() => {
    setSelectedIndex(0);
  }, [filteredResults]);

  // Reset query when dialog opens
  useEffect(() => {
    if (isOpen) {
      setQuery('');
      setSelectedIndex(0);
    }
  }, [isOpen]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setSelectedIndex((prev) => Math.min(prev + 1, filteredResults.length - 1));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setSelectedIndex((prev) => Math.max(prev - 1, 0));
    } else if (e.key === 'Enter') {
      e.preventDefault();
      if (filteredResults[selectedIndex]) {
        onClose();
        // In a real app, you'd navigate here
        window.location.href = filteredResults[selectedIndex].path;
      }
    }
  };

  const handleResultClick = (result: SearchResult) => {
    onClose();
    // In a real app, use React Router navigation
    window.location.href = result.path;
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl bg-bg-secondary border-white/12 p-0 gap-0">
        <DialogHeader className="px-6 py-4 border-b border-white/8">
          <DialogTitle className="sr-only">Global Search</DialogTitle>
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-text-muted" />
            <Input
              placeholder="Search functions, providers, pages..."
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onKeyDown={handleKeyDown}
              className="pl-10 pr-12 bg-transparent border-none text-white placeholder:text-text-muted focus:ring-0"
              autoFocus
            />
            <div className="absolute right-3 top-1/2 -translate-y-1/2 flex items-center gap-1">
              <kbd className="px-2 py-1 text-xs bg-white/10 rounded border border-white/20">
                <Command className="w-3 h-3 inline mr-1" />K
              </kbd>
            </div>
          </div>
        </DialogHeader>

        <div className="max-h-96 overflow-y-auto">
          {filteredResults.length === 0 ? (
            <div className="px-6 py-8 text-center text-text-muted">
              <Search className="w-12 h-12 mx-auto mb-4 opacity-50" />
              <p>No results found for "{query}"</p>
            </div>
          ) : (
            <div className="py-2">
              {filteredResults.map((result, index) => {
                const Icon = result.icon;
                const isSelected = index === selectedIndex;

                return (
                  <button
                    key={result.id}
                    onClick={() => handleResultClick(result)}
                    className={cn(
                      'w-full px-6 py-3 flex items-center gap-4 hover:bg-white/5 transition-colors text-left',
                      isSelected && 'bg-white/5'
                    )}
                  >
                    <div className="flex-shrink-0">
                      <Icon className="w-5 h-5 text-text-secondary" />
                    </div>

                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        <h3 className="text-sm font-medium text-white truncate">{result.title}</h3>
                        <Badge
                          variant="secondary"
                          className={cn('text-xs px-2 py-0.5', TYPE_COLORS[result.type])}
                        >
                          {TYPE_LABELS[result.type]}
                        </Badge>
                        {result.badge && (
                          <Badge
                            variant="outline"
                            className={cn(
                              'text-xs px-2 py-0.5 border-white/20',
                              result.badge === 'Active' && 'text-green-400',
                              result.badge === 'Failed' && 'text-red-400',
                              result.badge === 'Connected' && 'text-blue-400'
                            )}
                          >
                            {result.badge}
                          </Badge>
                        )}
                      </div>
                      {result.description && (
                        <p className="text-xs text-text-muted truncate">{result.description}</p>
                      )}
                    </div>
                  </button>
                );
              })}
            </div>
          )}
        </div>

        <div className="px-6 py-3 border-t border-white/8 bg-bg-tertiary">
          <div className="flex items-center justify-between text-xs text-text-muted">
            <div className="flex items-center gap-4">
              <span>↑↓ to navigate</span>
              <span>↵ to select</span>
              <span>esc to close</span>
            </div>
            <span>{filteredResults.length} results</span>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
