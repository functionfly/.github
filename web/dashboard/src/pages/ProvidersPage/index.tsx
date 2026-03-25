import { ProviderIcon } from '@/components/common/ProviderIcon';
import { StatusBadge } from '@/components/common/StatusBadge';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { PROVIDERS, PROVIDER_EXTERNAL_DASHBOARD_URL, ROUTES } from '@/lib/constants';
import { useProvidersStore } from '@/stores/providersStore';
import {
  AlertCircle,
  Check,
  ExternalLink,
  FunctionSquare,
  Globe,
  Loader2,
  Plus,
  X,
  Zap,
} from 'lucide-react';
import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';

// Provider brand colors for accents (keys match PROVIDERS constant IDs)
const providerAccents: Record<string, { border: string; glow: string; text: string }> = {
  workers: { border: '#f48120', glow: 'rgba(244, 129, 32, 0.15)', text: '#f48120' },
  vercel: { border: '#171717', glow: 'rgba(23, 23, 23, 0.15)', text: '#171717' },
  fly: { border: '#7b68ee', glow: 'rgba(123, 104, 238, 0.15)', text: '#7b68ee' },
  deno: { border: '#0a0a0a', glow: 'rgba(10, 10, 10, 0.15)', text: '#3c3c3c' },
  'functionfly-edge': { border: '#6366f1', glow: 'rgba(99, 102, 241, 0.2)', text: '#6366f1' },
};

interface ConnectionStatusProps {
  connected: boolean;
  status?: string;
}

function ConnectionStatus({ connected, status }: ConnectionStatusProps) {
  if (connected) {
    return (
      <div className="flex items-center gap-2">
        <Badge
          variant="outline"
          className="border-emerald-500/40 bg-emerald-50 dark:bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 font-medium"
        >
          <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 mr-1.5" />
          Connected
        </Badge>
        {status && <StatusBadge status={status as 'online' | 'offline' | 'degraded' | 'pending'} />}
      </div>
    );
  }

  return (
    <Badge
      variant="outline"
      className="border-amber-600/30 bg-amber-50 dark:bg-amber-500/10 text-amber-800 dark:text-amber-400 font-medium"
    >
      <span className="w-1.5 h-1.5 rounded-full bg-amber-500 mr-1.5 animate-pulse" />
      Not Connected
    </Badge>
  );
}

type ProviderConfig = (typeof PROVIDERS)[keyof typeof PROVIDERS];

interface ProviderCardProps {
  provider: ProviderConfig;
  connected: boolean;
  status: string;
  onDisconnect: () => void;
  onConnect: (providerId: string, apiKey?: string) => Promise<void>;
  isDisconnecting: boolean;
}

function ProviderCard({
  provider,
  connected,
  status,
  onDisconnect,
  onConnect,
  isDisconnecting,
}: ProviderCardProps) {
  const accent = providerAccents[provider.id] || {
    border: '#6366f1',
    glow: 'rgba(99, 102, 241, 0.2)',
    text: '#6366f1',
  };
  const isFunctionFly = (provider.id as string) === 'functionfly-edge';
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const providerDesc = (provider as any).description;

  return (
    <Card
      className="group relative overflow-hidden transition-all duration-300"
      style={{
        borderColor: connected ? `${accent.border}40` : undefined,
      }}
    >
      {/* Brand accent top border */}
      <div
        className="absolute top-0 left-0 right-0 h-1 transition-all duration-300"
        style={{
          backgroundColor: accent.border,
          opacity: connected ? 1 : 0.3,
        }}
      />

      {/* Subtle brand glow on hover */}
      <div
        className="absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity duration-500 pointer-events-none"
        style={{
          background: `radial-gradient(600px circle at 50% 0%, ${accent.glow}, transparent 50%)`,
        }}
      />

      <CardContent className="p-6 relative">
        {/* Header: Icon + Name + Status */}
        <div className="flex items-start justify-between mb-5">
          <div className="flex items-center gap-4">
            <div
              className="w-14 h-14 rounded-2xl flex items-center justify-center transition-transform duration-300 group-hover:scale-105"
              style={{
                backgroundColor: `${accent.border}15`,
              }}
            >
              <ProviderIcon provider={provider.id} size="lg" />
            </div>
            <div>
              <h3 className="font-semibold text-lg text-text-primary mb-0.5">{provider.name}</h3>
              <ConnectionStatus connected={connected} status={status} />
            </div>
          </div>
        </div>

        {/* Description for FunctionFly Edge */}
        {isFunctionFly && (
          <p className="text-sm text-text-secondary mb-4 leading-relaxed">
            {providerDesc ||
              "Host your edge functions on FunctionFly's infrastructure — no deployment required"}
          </p>
        )}

        {/* Compact regions display */}
        <div className="flex items-center gap-2 text-sm text-text-secondary mb-5">
          <Globe className="w-4 h-4 text-text-muted" />
          <span className="font-medium">{provider.regions.length} regions</span>
          <span className="text-text-muted">•</span>
          <span className="text-text-muted truncate">
            {provider.regions.slice(0, 3).join(', ')}
            {provider.regions.length > 3 && ` +${provider.regions.length - 3} more`}
          </span>
        </div>

        {/* Connection Details (only when connected) */}
        {connected && (
          <div className="mb-5 p-3 rounded-lg bg-bg-secondary/50 border border-border-subtle">
            <div className="flex items-center gap-2 text-sm">
              <Zap className="w-4 h-4" style={{ color: accent.text }} />
              <span className="text-text-secondary">
                Ready to deploy • Primary region:{' '}
                <span className="font-medium text-text-primary">{provider.regions[0]}</span>
              </span>
            </div>
          </div>
        )}

        {/* Action Buttons */}
        <div className="flex gap-3">
          {connected ? (
            <>
              {isFunctionFly ? (
                <Button
                  variant="outline"
                  className="flex-1 gap-2 border-border-default hover:border-border-subtle hover:bg-bg-secondary transition-colors"
                  asChild
                >
                  <Link to={ROUTES.FUNCTIONS}>
                    <FunctionSquare className="w-4 h-4" />
                    Configure
                  </Link>
                </Button>
              ) : (
                <Button
                  variant="outline"
                  className="flex-1 gap-2 border-border-default hover:border-border-subtle hover:bg-bg-secondary transition-colors"
                  asChild
                >
                  <a
                    href={
                      PROVIDER_EXTERNAL_DASHBOARD_URL[
                        provider.id as keyof typeof PROVIDER_EXTERNAL_DASHBOARD_URL
                      ]
                    }
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    <ExternalLink className="w-4 h-4" />
                    Configure
                  </a>
                </Button>
              )}
              <Button
                variant="outline"
                size="icon"
                className="border-border-default text-text-secondary hover:text-error hover:border-error/30 hover:bg-error/5 transition-colors"
                onClick={onDisconnect}
                disabled={isDisconnecting}
              >
                {isDisconnecting ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  <X className="w-4 h-4" />
                )}
              </Button>
            </>
          ) : (
            <ConnectDialog provider={provider} accent={accent} onConnect={onConnect} />
          )}
        </div>
      </CardContent>
    </Card>
  );
}

interface ConnectDialogProps {
  provider: ProviderConfig;
  accent: { border: string; glow: string; text: string };
  onConnect: (providerId: string, apiKey?: string) => Promise<void>;
}

function ConnectDialog({ provider, accent, onConnect }: ConnectDialogProps) {
  const [apiKey, setApiKey] = useState('');
  const [isConnecting, setIsConnecting] = useState(false);
  const [isOpen, setIsOpen] = useState(false);
  const isFunctionFly = provider.id === 'functionfly-edge';

  const handleConnect = async () => {
    setIsConnecting(true);
    try {
      await onConnect(provider.id, isFunctionFly ? undefined : apiKey);
      setIsOpen(false);
      setApiKey('');
    } catch (error) {
      // Error is handled by parent
    } finally {
      setIsConnecting(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger asChild>
        <Button
          variant="outline"
          className="w-full gap-2 border-border-default hover:border-border-subtle hover:bg-bg-secondary transition-all duration-200"
          onClick={() => setIsOpen(true)}
        >
          <Plus className="w-4 h-4" />
          Connect
        </Button>
      </DialogTrigger>
      <DialogContent className="bg-bg-tertiary border-border-subtle sm:max-w-md">
        <DialogHeader>
          <div className="flex items-center gap-3 mb-2">
            <div
              className="w-10 h-10 rounded-xl flex items-center justify-center"
              style={{ backgroundColor: `${accent.border}15` }}
            >
              <ProviderIcon provider={provider.id} size="md" />
            </div>
            <div>
              <DialogTitle className="text-text-primary text-lg">
                {isFunctionFly ? 'Enable' : 'Connect'} {provider.name}
              </DialogTitle>
            </div>
          </div>
          <DialogDescription className="text-text-secondary">
            {isFunctionFly
              ? 'Enable FunctionFly Edge to deploy to our managed infrastructure. No external account required.'
              : `Enter your API credentials to connect ${provider.name}. Credentials are encrypted with AES-256.`}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 pt-2">
          {isFunctionFly ? (
            <div className="p-4 rounded-lg bg-bg-secondary border border-border-subtle">
              <div className="flex items-start gap-3">
                <div
                  className="w-8 h-8 rounded-lg flex items-center justify-center shrink-0"
                  style={{ backgroundColor: `${accent.border}15` }}
                >
                  <Zap className="w-4 h-4" style={{ color: accent.text }} />
                </div>
                <div>
                  <h4 className="text-sm font-medium text-text-primary mb-1">Ready to Deploy</h4>
                  <p className="text-sm text-text-secondary">
                    FunctionFly Edge is our managed hosting. Select your preferred regions and
                    deploy immediately.
                  </p>
                </div>
              </div>
            </div>
          ) : (
            <>
              <div className="space-y-2">
                <Label htmlFor={`apiKey-${provider.id}`} className="text-text-primary">
                  API Key
                </Label>
                <Input
                  id={`apiKey-${provider.id}`}
                  type="password"
                  placeholder={`Enter your ${provider.name} API key`}
                  value={apiKey}
                  onChange={(e) => setApiKey(e.target.value)}
                  className="bg-bg-secondary border-border-subtle focus:border-border-default"
                />
              </div>
              <div className="flex items-start gap-2 p-3 rounded-lg bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-800/50">
                <AlertCircle className="w-4 h-4 text-amber-600 dark:text-amber-400 mt-0.5 shrink-0" />
                <p className="text-xs text-amber-950 dark:text-amber-100">
                  This allows FunctionFly to deploy to your {provider.name} account. You can revoke
                  access anytime.
                </p>
              </div>
            </>
          )}

          <button
            type="button"
            onClick={handleConnect}
            disabled={(!isFunctionFly && !apiKey) || isConnecting}
            className="w-full flex items-center justify-center gap-2 px-4 py-2.5 rounded-md font-medium transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed hover:opacity-90 active:scale-[0.98]"
            style={{
              backgroundColor: accent.border,
              color:
                provider.id === 'vercel' || provider.id === 'deno' || provider.id === 'workers'
                  ? 'white'
                  : 'white',
            }}
          >
            {isConnecting ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                {isFunctionFly ? 'Enabling...' : 'Connecting...'}
              </>
            ) : (
              <>
                <Check className="w-4 h-4" />
                {isFunctionFly ? 'Enable Provider' : 'Connect Provider'}
              </>
            )}
          </button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

export function ProvidersPage() {
  const [apiKey, setApiKey] = useState('');
  const [connecting, setConnecting] = useState(false);
  const [disconnecting, setDisconnecting] = useState<string | null>(null);
  const {
    providers,
    error,
    fetchProviders,
    connectProvider,
    disconnectProvider,
    testConnection,
    clearError,
  } = useProvidersStore();

  useEffect(() => {
    fetchProviders();
  }, [fetchProviders]);

  const handleConnect = async (providerId: string, key?: string) => {
    const isFunctionFly = providerId === 'functionfly-edge';
    const providerKey = key ?? '';

    if (!isFunctionFly && !providerKey.trim()) return;

    setConnecting(true);
    clearError();

    try {
      await connectProvider({ providerId, apiKey: providerKey });
      // Refresh from server to get accurate state
      await fetchProviders();
      setApiKey('');
    } catch (error) {
      console.error('Failed to connect provider:', error);
      throw error;
    } finally {
      setConnecting(false);
    }
  };

  const handleDisconnect = async (catalogProviderId: string) => {
    // API stores each row with a DB id; `ConnectedProvider.name` is the provider slug (workers, functionfly-edge, …).
    const row = providers.find((p) => p.name === catalogProviderId);
    if (!row) {
      console.warn('Disconnect: no connected provider for', catalogProviderId);
      return;
    }
    setDisconnecting(catalogProviderId);
    clearError();

    try {
      await disconnectProvider(row.id);
    } catch (error) {
      console.error('Failed to disconnect provider:', error);
    } finally {
      setDisconnecting(null);
    }
  };

  const isConnected = (catalogProviderId: string) =>
    providers.some((p) => p.name === catalogProviderId);

  const getProviderStatus = (catalogProviderId: string) => {
    const connected = providers.find((p) => p.name === catalogProviderId);
    return connected?.status || 'pending';
  };

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-text-primary tracking-tight">Providers</h1>
          <p className="text-text-secondary mt-1">Connect and manage your deployment targets</p>
        </div>
        <div className="flex items-center gap-2 text-sm text-text-secondary">
          <span className="flex items-center gap-1.5">
            <span className="w-2 h-2 rounded-full bg-emerald-500" />
            {providers.length} connected
          </span>
          <span className="text-text-muted">•</span>
          <span className="flex items-center gap-1.5">
            <span className="w-2 h-2 rounded-full bg-amber-500" />
            {Object.keys(PROVIDERS).length - providers.length} available
          </span>
        </div>
      </div>

      {/* Error Message */}
      {error && (
        <div className="p-4 bg-error/10 border border-error/20 rounded-lg">
          <p className="text-error">{error}</p>
          <Button
            variant="ghost"
            size="sm"
            className="mt-2 text-error hover:text-error hover:bg-error/10"
            onClick={clearError}
          >
            Dismiss
          </Button>
        </div>
      )}

      {/* Providers Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-5">
        {Object.values(PROVIDERS).map((provider) => {
          const connected = isConnected(provider.id);
          const status = getProviderStatus(provider.id);

          return (
            <ProviderCard
              key={provider.id}
              provider={provider}
              connected={connected}
              status={status}
              onDisconnect={() => handleDisconnect(provider.id)}
              onConnect={async (pid, key) => handleConnect(pid, key)}
              isDisconnecting={disconnecting === provider.id}
            />
          );
        })}
      </div>
    </div>
  );
}
