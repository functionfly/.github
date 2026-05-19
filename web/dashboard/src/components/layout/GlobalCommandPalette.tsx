import * as React from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Command,
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
  CommandShortcut,
} from '@/components/ui/command';
import { ROUTES } from '@/lib/constants';
import { conversationsApi, type Conversation } from '@/api/conversations';
import {
  BarChart3,
  Cloud,
  Code,
  FunctionSquare,
  LayoutDashboard,
  Loader2,
  MessageSquare,
  Search,
  Settings,
  Sparkles,
  Zap,
  Bot,
  Bug,
  Rocket,
  Wrench,
  Brain,
  Users,
  DollarSign,
  Shield,
  Building2,
} from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';

interface GlobalCommandPaletteProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

const CONVERSATION_TYPE_ICONS: Record<string, React.ComponentType<{ className?: string }>> = {
  dm: MessageSquare,
  function_thread: FunctionSquare,
  issue_thread: Bug,
  fix_mode: Wrench,
  bounty_thread: DollarSign,
  org_thread: Building2,
  security_disclosure: Shield,
};

export function GlobalCommandPalette({ open, onOpenChange }: GlobalCommandPaletteProps) {
  const navigate = useNavigate();
  const { username } = useParams<{ username?: string }>();
  const [searchQuery, setSearchQuery] = React.useState('');
  const [conversations, setConversations] = React.useState<Conversation[]>([]);
  const [loading, setLoading] = React.useState(false);

  const goTo = React.useCallback(
    (path: string) => {
      onOpenChange(false);
      setSearchQuery('');
      navigate(path);
    },
    [navigate, onOpenChange],
  );

  // Search conversations when query changes
  React.useEffect(() => {
    if (!open || !searchQuery.trim() || !username) {
      setConversations([]);
      return;
    }

    const timer = setTimeout(async () => {
      setLoading(true);
      try {
        const res = await conversationsApi.searchMessages(username, {
          q: searchQuery,
          limit: 5,
        });
        // Extract unique conversation IDs from messages, then fetch conversations
        const convIds = [...new Set(res.messages.map((m) => m.conversation_id))];
        const convs = await conversationsApi.listConversations(username, { limit: 20 });
        const filtered = convs.conversations.filter(
          (c) =>
            convIds.includes(c.id) ||
            c.type.toLowerCase().includes(searchQuery.toLowerCase()) ||
            c.participant_ids.some((p) => p.toLowerCase().includes(searchQuery.toLowerCase())),
        );
        setConversations(filtered.slice(0, 5));
      } catch {
        setConversations([]);
      } finally {
        setLoading(false);
      }
    }, 300);

    return () => clearTimeout(timer);
  }, [searchQuery, open, username]);

  // Reset on close
  React.useEffect(() => {
    if (!open) {
      setSearchQuery('');
      setConversations([]);
    }
  }, [open]);

  const handleOpenChange = (nextOpen: boolean) => {
    onOpenChange(nextOpen);
    if (!nextOpen) {
      setSearchQuery('');
      setConversations([]);
    }
  };

  return (
    <CommandDialog open={open} onOpenChange={handleOpenChange}>
      <CommandInput
        placeholder="Search functions, conversations, pages..."
        value={searchQuery}
        onValueChange={setSearchQuery}
      />
      <CommandList>
        <CommandEmpty>
          {loading ? (
            <div className="flex items-center justify-center gap-2 py-6 text-sm text-muted-foreground">
              <Loader2 className="h-4 w-4 animate-spin" />
              Searching...
            </div>
          ) : (
            'No results found.'
          )}
        </CommandEmpty>

        {/* Conversation search results */}
        {conversations.length > 0 && (
          <CommandGroup heading="Conversations">
            {conversations.map((c) => {
              const Icon = CONVERSATION_TYPE_ICONS[c.type] || MessageSquare;
              return (
                <CommandItem
                  key={c.id}
                  value={`conversation-${c.id}`}
                  onSelect={() =>
                    goTo(
                      username
                        ? `/u/${username}/conversations/${c.id}`
                        : `/conversations/${c.id}`,
                    )
                  }
                >
                  <Icon className="mr-2 h-4 w-4" />
                  <div className="flex-1 min-w-0">
                    <span className="block truncate text-sm">
                      {c.type.replace(/_/g, ' ')}
                    </span>
                    <span className="block truncate text-xs text-muted-foreground">
                      {formatDistanceToNow(new Date(c.updated_at), { addSuffix: true })}
                    </span>
                  </div>
                  {(c.unread_count ?? 0) > 0 && (
                    <span className="ml-auto shrink-0 flex h-5 min-w-5 items-center justify-center rounded-full bg-brand-500 px-1.5 text-[10px] font-semibold text-white tabular-nums">
                      {c.unread_count}
                    </span>
                  )}
                </CommandItem>
              );
            })}
          </CommandGroup>
        )}

        {conversations.length > 0 && <CommandSeparator />}

        {/* Empty query: show navigation */}
        {!searchQuery.trim() && (
          <>
            <CommandGroup heading="Navigation">
              <CommandItem onSelect={() => goTo(ROUTES.DASHBOARD)}>
                <LayoutDashboard className="mr-2 h-4 w-4" />
                <span>Overview</span>
              </CommandItem>
              <CommandItem onSelect={() => goTo(ROUTES.FUNCTIONS)}>
                <FunctionSquare className="mr-2 h-4 w-4" />
                <span>Functions</span>
              </CommandItem>
              <CommandItem onSelect={() => goTo(ROUTES.PROVIDERS)}>
                <Cloud className="mr-2 h-4 w-4" />
                <span>Providers</span>
              </CommandItem>
              <CommandItem onSelect={() => goTo(ROUTES.ANALYTICS)}>
                <BarChart3 className="mr-2 h-4 w-4" />
                <span>Analytics</span>
              </CommandItem>
              <CommandItem onSelect={() => goTo('/devops')}>
                <Rocket className="mr-2 h-4 w-4" />
                <span>DevOps</span>
              </CommandItem>
              <CommandItem onSelect={() => goTo('/security')}>
                <Shield className="mr-2 h-4 w-4" />
                <span>Security</span>
              </CommandItem>
              <CommandItem onSelect={() => goTo('/code-intelligence')}>
                <Brain className="mr-2 h-4 w-4" />
                <span>Code Intelligence</span>
              </CommandItem>
              <CommandItem onSelect={() => goTo('/collaboration')}>
                <Users className="mr-2 h-4 w-4" />
                <span>Collaboration</span>
              </CommandItem>
              <CommandItem onSelect={() => goTo('/memory')}>
                <Brain className="mr-2 h-4 w-4" />
                <span>Memory</span>
              </CommandItem>
              <CommandItem onSelect={() => goTo('/simulation')}>
                <Gauge className="mr-2 h-4 w-4" />
                <span>R-Sim</span>
              </CommandItem>
              <CommandItem onSelect={() => goTo('/robotics')}>
                <Bot className="mr-2 h-4 w-4" />
                <span>Robotics</span>
              </CommandItem>
              <CommandItem onSelect={() => goTo('/marketplace-economy')}>
                <DollarSign className="mr-2 h-4 w-4" />
                <span>Marketplace Economy</span>
              </CommandItem>
              <CommandItem onSelect={() => goTo('/adaptive-ux')}>
                <Brain className="mr-2 h-4 w-4" />
                <span>Adaptive UX</span>
              </CommandItem>
              <CommandItem onSelect={() => goTo('/universal-runtime')}>
                <Boxes className="mr-2 h-4 w-4" />
                <span>Universal Runtime</span>
              </CommandItem>
              <CommandItem onSelect={() => goTo("/data-visualization")}>
                <BarChart3 className="mr-2 h-4 w-4" />
                <span>Data Visualization</span>
              </CommandItem>
              <CommandItem onSelect={() => goTo('/futuristic')}>
                <Sparkles className="mr-2 h-4 w-4" />
                <span>Futuristic</span>
              </CommandItem>
              <CommandItem
                onSelect={() =>
                  goTo(username ? `/u/${username}/agents` : ROUTES.AGENT_LIST)
                }
              >
                <Bot className="mr-2 h-4 w-4" />
                <span>Agents</span>
              </CommandItem>
              <CommandItem onSelect={() => goTo(ROUTES.SETTINGS)}>
                <Settings className="mr-2 h-4 w-4" />
                <span>Settings</span>
              </CommandItem>
            </CommandGroup>

            <CommandSeparator />

            <CommandGroup heading="Conversations">
              <CommandItem
                onSelect={() =>
                  goTo(username ? `/u/${username}/conversations` : '/conversations')
                }
              >
                <MessageSquare className="mr-2 h-4 w-4" />
                <span>All Conversations</span>
                <CommandShortcut>\u2318\u21e7M</CommandShortcut>
              </CommandItem>
            </CommandGroup>

            <CommandSeparator />

            <CommandGroup heading="Quick Actions">
              <CommandItem
                onSelect={() => {
                  onOpenChange(false);
                  goTo(ROUTES.FUNCTIONS);
                }}
              >
                <Sparkles className="mr-2 h-4 w-4" />
                <span>New Function</span>
                <CommandShortcut>\u2318N</CommandShortcut>
              </CommandItem>
              <CommandItem
                onSelect={() => {
                  onOpenChange(false);
                  goTo(ROUTES.DASHBOARD);
                }}
              >
                <Zap className="mr-2 h-4 w-4" />
                <span>Function Marketplace</span>
              </CommandItem>
              <CommandItem
                onSelect={() => {
                  onOpenChange(false);
                  goTo(ROUTES.FUNCTIONS);
                }}
              >
                <Code className="mr-2 h-4 w-4" />
                <span>Browse Registry</span>
              </CommandItem>
            </CommandGroup>

            <CommandSeparator />

            <CommandGroup heading="Help">
              <CommandItem
                onSelect={() => {
                  onOpenChange(false);
                  window.open('/docs', '_blank');
                }}
              >
                <Search className="mr-2 h-4 w-4" />
                <span>Search Documentation</span>
              </CommandItem>
            </CommandGroup>
          </>
        )}

        {/* Typed query: show function/provider shortcuts */}
        {searchQuery.trim() && (
          <>
            <CommandGroup heading="Pages">
              <CommandItem onSelect={() => goTo(ROUTES.FUNCTIONS)}>
                <FunctionSquare className="mr-2 h-4 w-4" />
                <span>Browse Functions</span>
              </CommandItem>
              <CommandItem
                onSelect={() =>
                  goTo(username ? `/u/${username}/agents` : ROUTES.AGENT_LIST)
                }
              >
                <Bot className="mr-2 h-4 w-4" />
                <span>Browse Agents</span>
              </CommandItem>
              <CommandItem onSelect={() => goTo(ROUTES.PROVIDERS)}>
                <Cloud className="mr-2 h-4 w-4" />
                <span>Browse Providers</span>
              </CommandItem>
            </CommandGroup>
          </>
        )}
      </CommandList>
    </CommandDialog>
  );
}
