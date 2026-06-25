/**
 * Brain Page - Connectors Tab
 * Manage integrations with external services that feed the Brain
 */

import { useState, useCallback, useEffect } from 'react';
import {
  GitBranch,
  MessageSquare,
  Mail,
  FileText,
  Zap,
  Link2,
  Unlink,
  RefreshCw,
  Loader2,
  Check,
  X,
  AlertCircle,
  ExternalLink,
  Plus,
  Settings,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { EmptyState } from '@/components/ui/empty-state';
import { toast } from 'sonner';
import type { Connector, UserConnector } from '@/api/connectors';
import { connectorsApi } from '@/api/connectors';

const connectorIcons: Record<string, React.ReactNode> = {
  github: <GitBranch className="w-5 h-5" />,
  notion: <FileText className="w-5 h-5" />,
  slack: <MessageSquare className="w-5 h-5" />,
  gmail: <Mail className="w-5 h-5" />,
  linear: <Zap className="w-5 h-5" />,
};

const connectorColors: Record<string, { bg: string; border: string; text: string }> = {
  github: { bg: 'bg-gray-500/10', border: 'border-gray-500/30', text: 'text-gray-400' },
  notion: { bg: 'bg-gray-500/10', border: 'border-gray-500/30', text: 'text-gray-300' },
  slack: { bg: 'bg-purple-500/10', border: 'border-purple-500/30', text: 'text-purple-400' },
  gmail: { bg: 'bg-red-500/10', border: 'border-red-500/30', text: 'text-red-400' },
  linear: { bg: 'bg-blue-500/10', border: 'border-blue-500/30', text: 'text-blue-400' },
};

const connectorDescriptions: Record<string, string> = {
  github: 'Issues, PRs, commits, and discussions',
  notion: 'Pages, databases, and comments',
  slack: 'Channels, messages, and reactions',
  gmail: 'Emails, labels, and threads',
  linear: 'Issues, projects, and cycles',
};

interface ConnectorsTabProps {
  onConnectorLinked?: () => void;
}

export function ConnectorsTab({ onConnectorLinked }: ConnectorsTabProps) {
  const [catalog, setCatalog] = useState<Connector[]>([]);
  const [userConnectors, setUserConnectors] = useState<UserConnector[]>([]);
  const [loading, setLoading] = useState(true);
  const [linkingSlug, setLinkingSlug] = useState<string | null>(null);
  const [syncingId, setSyncingId] = useState<string | null>(null);
  const [unlinkingId, setUnlinkingId] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  const loadCatalog = useCallback(async () => {
    try {
      const connectors = await connectorsApi.listCatalog();
      setCatalog(connectors);
    } catch (err) {
      console.error('Failed to load catalog:', err);
      toast.error('Failed to load connector catalog');
    }
  }, []);

  const loadUserConnectors = useCallback(async () => {
    try {
      const connectors = await connectorsApi.listUserConnectors();
      setUserConnectors(connectors);
    } catch (err) {
      console.error('Failed to load user connectors:', err);
    }
  }, []);

  useEffect(() => {
    const loadAll = async () => {
      setLoading(true);
      await Promise.all([loadCatalog(), loadUserConnectors()]);
      setLoading(false);
    };
    loadAll();
  }, [loadCatalog, loadUserConnectors]);

  const handleLinkConnector = async (slug: string) => {
    setLinkingSlug(slug);
    try {
      const { oauth_url } = await connectorsApi.getConnectorOAuthUrl(slug);
      // Open OAuth URL in popup
      const width = 600;
      const height = 700;
      const left = window.screenX + (window.outerWidth - width) / 2;
      const top = window.screenY + (window.outerHeight - height) / 2;
      const popup = window.open(
        oauth_url,
        'oauth',
        `width=${width},height=${height},left=${left},top=${top},popup=yes`
      );

      // Listen for OAuth callback
      const messageHandler = async (event: MessageEvent) => {
        if (event.data?.type === 'oauth_callback') {
          window.removeEventListener('message', messageHandler);
          if (event.data.status === 'success') {
            toast.success(`${event.data.connector_name} connected successfully`);
            await loadUserConnectors();
            await loadCatalog();
            onConnectorLinked?.();
          } else {
            toast.error(`Failed to connect: ${event.data.message}`);
          }
          popup?.close();
        }
      };
      window.addEventListener('message', messageHandler);

      // Fallback: check if popup was closed without callback
      const checkClosed = setInterval(() => {
        if (popup?.closed) {
          clearInterval(checkClosed);
          window.removeEventListener('message', messageHandler);
          // Reload user connectors in case it was linked
          loadUserConnectors();
          loadCatalog();
        }
      }, 1000);
    } catch (err) {
      console.error('Failed to start OAuth flow:', err);
      toast.error('Failed to start OAuth flow');
    } finally {
      setLinkingSlug(null);
    }
  };

  const handleUnlinkConnector = async (connectorId: string) => {
    setUnlinkingId(connectorId);
    try {
      await connectorsApi.unlinkConnector(connectorId);
      toast.success('Connector unlinked');
      await loadUserConnectors();
      await loadCatalog();
    } catch (err) {
      console.error('Failed to unlink connector:', err);
      toast.error('Failed to unlink connector');
    } finally {
      setUnlinkingId(null);
    }
  };

  const handleSyncConnector = async (connectorId: string) => {
    setSyncingId(connectorId);
    try {
      const result = await connectorsApi.triggerSync(connectorId);
      toast.success(result.message);
    } catch (err) {
      console.error('Failed to sync connector:', err);
      toast.error('Failed to trigger sync');
    } finally {
      setSyncingId(null);
    }
  };

  const filteredCatalog = catalog.filter(
    (c) =>
      c.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      c.slug.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const linkedSlugs = new Set(userConnectors.map((uc) => uc.connector_slug));

  const getStatusBadge = (status: UserConnector['status']) => {
    switch (status) {
      case 'active':
        return (
          <Badge className="bg-emerald-500/10 text-emerald-400 border-emerald-500/30">
            <Check className="w-3 h-3 mr-1" />
            Active
          </Badge>
        );
      case 'sync_error':
        return (
          <Badge className="bg-red-500/10 text-red-400 border-red-500/30">
            <AlertCircle className="w-3 h-3 mr-1" />
            Error
          </Badge>
        );
      case 'revoked':
        return (
          <Badge className="bg-amber-500/10 text-amber-400 border-amber-500/30">
            <AlertCircle className="w-3 h-3 mr-1" />
            Revoked
          </Badge>
        );
      case 'disabled':
        return (
          <Badge className="bg-gray-500/10 text-gray-400 border-gray-500/30">
            Disabled
          </Badge>
        );
      default:
        return null;
    }
  };

  const formatLastSync = (dateStr?: string) => {
    if (!dateStr) return 'Never';
    const date = new Date(dateStr);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const mins = Math.floor(diff / 60000);
    if (mins < 1) return 'Just now';
    if (mins < 60) return `${mins}m ago`;
    const hours = Math.floor(mins / 60);
    if (hours < 24) return `${hours}h ago`;
    const days = Math.floor(hours / 24);
    return `${days}d ago`;
  };

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <Card key={i} className="border-white/[0.06] bg-white/[0.02]">
              <CardContent className="p-4">
                <div className="h-20 bg-white/[0.05] animate-pulse rounded" />
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h2 className="text-xl font-semibold text-text-primary flex items-center gap-2">
            <GitBranch className="w-5 h-5 text-indigo-400" />
            Connectors
          </h2>
          <p className="text-sm text-text-secondary mt-1">
            Connect external accounts to feed your Brain with signals
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              loadCatalog();
              loadUserConnectors();
            }}
            className="border-white/10 hover:bg-white/5"
          >
            <RefreshCw className="w-3.5 h-3.5 mr-1.5" />
            Refresh
          </Button>
        </div>
      </div>

      {/* Linked Connectors */}
      {userConnectors.length > 0 && (
        <div className="space-y-4">
          <h3 className="text-sm font-medium text-text-secondary">Your Connected Accounts</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {userConnectors.map((uc) => {
              const color = connectorColors[uc.connector_slug] || {
                bg: 'bg-white/5',
                border: 'border-white/10',
                text: 'text-gray-400',
              };
              const icon = connectorIcons[uc.connector_slug] || <Link2 className="w-5 h-5" />;

              return (
                <Card
                  key={uc.id}
                  className={`border-white/[0.06] bg-white/[0.02] ${uc.status !== 'active' ? 'opacity-70' : ''}`}
                >
                  <CardContent className="p-4">
                    <div className="flex items-start justify-between">
                      <div className="flex items-center gap-3">
                        <div className={`p-2 rounded-lg ${color.bg}`}>{icon}</div>
                        <div>
                          <p className="font-medium text-text-primary">{uc.display_name}</p>
                          <p className="text-xs text-text-secondary">{uc.connector_name}</p>
                        </div>
                      </div>
                      {getStatusBadge(uc.status)}
                    </div>
                    <div className="mt-3 flex items-center justify-between text-xs text-text-secondary">
                      <span>Last sync: {formatLastSync(uc.last_sync_at)}</span>
                      {uc.sync_error && (
                        <span className="text-red-400 truncate max-w-[150px]" title={uc.sync_error}>
                          {uc.sync_error}
                        </span>
                      )}
                    </div>
                    <div className="mt-3 flex items-center gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        className="flex-1 border-white/10 hover:bg-white/5"
                        onClick={() => handleSyncConnector(uc.id)}
                        disabled={syncingId === uc.id || uc.status !== 'active'}
                      >
                        {syncingId === uc.id ? (
                          <Loader2 className="w-3.5 h-3.5 animate-spin" />
                        ) : (
                          <RefreshCw className="w-3.5 h-3.5" />
                        )}
                        Sync
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="text-red-400 hover:text-red-300 hover:bg-red-500/10"
                        onClick={() => handleUnlinkConnector(uc.id)}
                        disabled={unlinkingId === uc.id}
                      >
                        {unlinkingId === uc.id ? (
                          <Loader2 className="w-3.5 h-3.5 animate-spin" />
                        ) : (
                          <Unlink className="w-3.5 h-3.5" />
                        )}
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              );
            })}
          </div>
        </div>
      )}

      {/* Available Connectors */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-medium text-text-secondary">Available Connectors</h3>
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-secondary" />
            <Input
              placeholder="Search connectors..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-9 w-[200px] h-9 bg-white/[0.03] border-white/[0.06] text-sm"
            />
          </div>
        </div>

        {filteredCatalog.length === 0 ? (
          <EmptyState
            variant="card"
            icon={<GitBranch className="w-8 h-8" />}
            title="No connectors found"
            description="Try adjusting your search query"
          />
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filteredCatalog.map((connector) => {
              const isLinked = linkedSlugs.has(connector.slug);
              const color = connectorColors[connector.slug] || {
                bg: 'bg-white/5',
                border: 'border-white/10',
                text: 'text-gray-400',
              };
              const icon = connectorIcons[connector.slug] || <Link2 className="w-5 h-5" />;
              const description = connectorDescriptions[connector.slug] || connector.scopes;

              return (
                <Card
                  key={connector.id}
                  className={`border-white/[0.06] bg-white/[0.02] hover:bg-white/[0.04] transition-colors ${
                    isLinked ? 'ring-1 ring-indigo-500/30' : ''
                  }`}
                >
                  <CardContent className="p-4">
                    <div className="flex items-start justify-between mb-3">
                      <div className="flex items-center gap-3">
                        <div className={`p-2 rounded-lg ${color.bg}`}>{icon}</div>
                        <div>
                          <p className="font-medium text-text-primary">{connector.name}</p>
                          <p className="text-xs text-text-secondary">{description}</p>
                        </div>
                      </div>
                      {isLinked && (
                        <Badge className="bg-indigo-500/10 text-indigo-400 border-indigo-500/30">
                          <Check className="w-3 h-3 mr-1" />
                          Linked
                        </Badge>
                      )}
                    </div>
                    <Button
                      variant={isLinked ? 'ghost' : 'default'}
                      size="sm"
                      className={`w-full ${isLinked ? 'border-white/10' : ''}`}
                      onClick={() => handleLinkConnector(connector.slug)}
                      disabled={linkingSlug === connector.slug}
                    >
                      {linkingSlug === connector.slug ? (
                        <Loader2 className="w-3.5 h-3.5 animate-spin" />
                      ) : isLinked ? (
                        <>
                          <Link2 className="w-3.5 h-3.5 mr-1.5" />
                          Reconnect
                        </>
                      ) : (
                        <>
                          <Plus className="w-3.5 h-3.5 mr-1.5" />
                          Connect
                        </>
                      )}
                    </Button>
                  </CardContent>
                </Card>
              );
            })}
          </div>
        )}
      </div>

      {/* Help Text */}
      <Card className="border-white/[0.06] bg-white/[0.02]">
        <CardContent className="p-4">
          <div className="flex items-start gap-3">
            <AlertCircle className="w-5 h-5 text-indigo-400 flex-shrink-0 mt-0.5" />
            <div className="text-sm">
              <p className="text-text-primary font-medium mb-1">How connectors work</p>
              <p className="text-text-secondary">
                Connectors authenticate with your external accounts and periodically sync data to
                your Brain. Each signal extracted from your connected accounts is scored for
                importance and stored securely. Your Brain uses these signals to provide smarter
                context to AI agents.
              </p>
              <div className="mt-3 flex items-center gap-4 text-xs text-text-secondary">
                <a
                  href="https://docs.functionfly.com/brain/connectors"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-indigo-400 hover:text-indigo-300 flex items-center gap-1"
                >
                  Documentation
                  <ExternalLink className="w-3 h-3" />
                </a>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function Search({ className }: { className?: string }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
    >
      <circle cx="11" cy="11" r="8" />
      <path d="m21 21-4.3-4.3" />
    </svg>
  );
}
