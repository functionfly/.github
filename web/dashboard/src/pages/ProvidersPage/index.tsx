import { ProviderIcon } from '@/components/common/ProviderIcon';
import { StatusBadge } from '@/components/common/StatusBadge';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Skeleton } from '@/components/ui/skeleton';
import { PROVIDERS, PROVIDER_EXTERNAL_DASHBOARD_URL, ROUTES } from '@/lib/constants';
import { useProvidersStore } from '@/stores/providersStore';
import {
  AlertCircle,
  AlertTriangle,
  Check,
  Clock,
  ExternalLink,
  FunctionSquare,
  Globe,
  Loader2,
  Plus,
  RefreshCw,
  Shield,
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

// Loading skeleton for provider cards
function ProviderCardSkeleton() {
  return (
    <Card className="relative overflow-hidden">
      <div className="absolute top-0 left-0 right-0 h-1 bg-slate-200 dark:bg-slate-700" />
      <CardContent className="p-6">
        <div className="flex items-start gap-4 mb-5">
          <Skeleton className="w-14 h-14 rounded-2xl" />
          <div className="flex-1">
            <Skeleton className="h-5 w-32 mb-2" />
            <Skeleton className="h-4 w-20" />
          </div>
        </div>
        <Skeleton className="h-4 w-40 mb-5" />
        <Skeleton className="h-10 w-full" />
      </CardContent>
    </Card>
  );
}

interface ProviderCardProps {
  provider: ProviderConfig;
  connected: boolean;
  status: string;
  onDisconnect: () => void;
  onConnect: (providerId: string, apiKey?: string) => Promise<void>;
  onTestConnection?: () => Promise<void>;
  isDisconnecting: boolean;
  isTestingConnection?: boolean;
  lastUsedAt?: string;
  isStale?: boolean;
  connectionTestResult?: 'success' | 'error' | null;
}

function formatRelativeTime(dateString?: string): string | null {
  if (!dateString) return null;
  const date = new Date(dateString);
  const now = new Date();
  const diffInMs = now.getTime() - date.getTime();
  const diffInDays = Math.floor(diffInMs / (1000 * 60 * 60 * 24));
  const diffInHours = Math.floor(diffInMs / (1000 * 60 * 60));
  const diffInMinutes = Math.floor(diffInMs / (1000 * 60));

  if (diffInMinutes < 1) return 'Just now';
  if (diffInMinutes < 60) return `${diffInMinutes}m ago`;
  if (diffInHours < 24) return `${diffInHours}h ago`;
  if (diffInDays < 7) return `${diffInDays}d ago`;
  return date.toLocaleDateString();
}

function ProviderCard({
  provider,
  connected,
  status,
  onDisconnect,
  onConnect,
  onTestConnection,
  isDisconnecting,
  isTestingConnection,
  lastUsedAt,
  isStale,
  connectionTestResult,
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

        {/* Stale Connection Warning */}
        {connected && (status === 'degraded' || isStale) && (
          <div className="mb-5 p-3 rounded-lg bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-800/50">
            <div className="flex items-start gap-2">
              <AlertTriangle className="w-4 h-4 text-amber-600 dark:text-amber-400 mt-0.5 shrink-0" />
              <div className="text-sm">
                <p className="text-amber-800 dark:text-amber-300 font-medium">
                  Connection may be stale
                </p>
                <p className="text-amber-700 dark:text-amber-400 text-xs mt-0.5">
                  This provider hasn't been used recently or the connection test failed. Click
                  "Test" to verify your credentials.
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Last Used Timestamp */}
        {connected && lastUsedAt && (
          <div className="mb-5 flex items-center gap-2 text-xs text-text-tertiary">
            <Clock className="w-3.5 h-3.5" />
            <span>Last used {formatRelativeTime(lastUsedAt)}</span>
          </div>
        )}

        {/* Connection Test Result */}
        {connectionTestResult && (
          <div
            className={`mb-4 p-2.5 rounded-lg text-sm animate-in slide-in-from-top-1 duration-200 ${
              connectionTestResult === 'success'
                ? 'bg-emerald-50 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-800/50'
                : 'bg-error/10 border border-error/20'
            }`}
          >
            <div className="flex items-center gap-2">
              {connectionTestResult === 'success' ? (
                <>
                  <Check className="w-4 h-4 text-emerald-600 dark:text-emerald-400" />
                  <span className="text-emerald-700 dark:text-emerald-400 font-medium">
                    Connection successful
                  </span>
                </>
              ) : (
                <>
                  <AlertCircle className="w-4 h-4 text-error" />
                  <span className="text-error font-medium">Connection test failed</span>
                </>
              )}
            </div>
          </div>
        )}

        {/* Action Buttons */}
        <div className="flex gap-2">
          {connected ? (
            <>
              {/* Test Connection Button */}
              {onTestConnection && (
                <Button
                  variant="outline"
                  size="sm"
                  className={`flex-1 gap-1.5 border-border-default hover:border-border-subtle hover:bg-bg-secondary transition-colors ${
                    connectionTestResult === 'success'
                      ? 'border-emerald-500/30 bg-emerald-50/50 dark:bg-emerald-950/20'
                      : connectionTestResult === 'error'
                        ? 'border-error/30 bg-error/5'
                        : ''
                  }`}
                  onClick={onTestConnection}
                  disabled={isTestingConnection}
                  title="Test provider connection"
                >
                  {isTestingConnection ? (
                    <Loader2 className="w-3.5 h-3.5 animate-spin" />
                  ) : connectionTestResult === 'success' ? (
                    <Check className="w-3.5 h-3.5 text-emerald-600 dark:text-emerald-400" />
                  ) : (
                    <Shield className="w-3.5 h-3.5" />
                  )}
                  Test
                </Button>
              )}

              {isFunctionFly ? (
                <Button
                  variant="outline"
                  size="sm"
                  className="flex-1 gap-1.5 border-border-default hover:border-border-subtle hover:bg-bg-secondary transition-colors"
                  asChild
                >
                  <Link to={ROUTES.FUNCTIONS}>
                    <FunctionSquare className="w-3.5 h-3.5" />
                    Configure
                  </Link>
                </Button>
              ) : (
                <Button
                  variant="outline"
                  size="sm"
                  className="flex-1 gap-1.5 border-border-default hover:border-border-subtle hover:bg-bg-secondary transition-colors"
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
                    <ExternalLink className="w-3.5 h-3.5" />
                    Configure
                  </a>
                </Button>
              )}
              <Button
                variant="outline"
                size="icon"
                className="border-border-default text-text-secondary hover:text-error hover:border-error/30 hover:bg-error/5 transition-colors shrink-0"
                onClick={onDisconnect}
                disabled={isDisconnecting}
                title="Disconnect provider"
                aria-label={`Disconnect ${provider.name} provider`}
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

interface DisconnectConfirmationDialogProps {
  providerName: string;
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  isDisconnecting: boolean;
}

function DisconnectConfirmationDialog({
  providerName,
  isOpen,
  onClose,
  onConfirm,
  isDisconnecting,
}: DisconnectConfirmationDialogProps) {
  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="bg-bg-tertiary border-border-subtle sm:max-w-md">
        <DialogHeader>
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded-xl flex items-center justify-center bg-amber-100 dark:bg-amber-900/30">
              <AlertTriangle className="w-5 h-5 text-amber-600 dark:text-amber-400" />
            </div>
            <div>
              <DialogTitle className="text-text-primary text-lg">Disconnect Provider</DialogTitle>
            </div>
          </div>
          <DialogDescription className="text-text-secondary">
            Are you sure you want to disconnect <strong>{providerName}</strong>? This action cannot
            be undone.
          </DialogDescription>
        </DialogHeader>

        <Alert className="bg-amber-50 dark:bg-amber-950/30 border-amber-200 dark:border-amber-800/50">
          <AlertCircle className="h-4 w-4 text-amber-600 dark:text-amber-400" />
          <AlertTitle className="text-amber-800 dark:text-amber-300">Warning</AlertTitle>
          <AlertDescription className="text-amber-700 dark:text-amber-400 text-sm">
            Disconnecting this provider will:
            <ul className="list-disc list-inside mt-1 space-y-0.5">
              <li>Prevent new deployments to {providerName}</li>
              <li>Remove stored API credentials (they cannot be recovered)</li>
              <li>Keep existing deployments running but disable updates</li>
            </ul>
          </AlertDescription>
        </Alert>

        <DialogFooter className="gap-2 sm:gap-0">
          <Button
            variant="outline"
            onClick={onClose}
            disabled={isDisconnecting}
            className="border-border-default"
          >
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={onConfirm}
            disabled={isDisconnecting}
            className="gap-2"
          >
            {isDisconnecting ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                Disconnecting...
              </>
            ) : (
              <>
                <X className="w-4 h-4" />
                Disconnect
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
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
  const [validationError, setValidationError] = useState<string | null>(null);
  const isFunctionFly = provider.id === 'functionfly-edge';

  const handleConnect = async () => {
    // Clear previous validation errors
    setValidationError(null);

    // Client-side validation
    if (!isFunctionFly) {
      if (!apiKey.trim()) {
        setValidationError('API key is required');
        return;
      }
      if (apiKey.length < 10) {
        setValidationError('API key must be at least 10 characters');
        return;
      }
    }

    setIsConnecting(true);
    try {
      await onConnect(provider.id, isFunctionFly ? undefined : apiKey);
      setIsOpen(false);
      setApiKey('');
      setValidationError(null);
    } catch (error) {
      // Error is handled by parent - could also set validationError here if needed
      const errorMessage = error instanceof Error ? error.message : 'Failed to connect provider';
      setValidationError(errorMessage);
    } finally {
      setIsConnecting(false);
    }
  };

  const handleClose = () => {
    setIsOpen(false);
    setApiKey('');
    setValidationError(null);
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && handleClose()}>
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
          {/* Validation Error */}
          {validationError && (
            <div className="p-3 rounded-lg bg-error/10 border border-error/20 animate-in slide-in-from-top-1 duration-200">
              <div className="flex items-start gap-2">
                <AlertCircle className="w-4 h-4 text-error mt-0.5 shrink-0" />
                <p className="text-sm text-error">{validationError}</p>
              </div>
            </div>
          )}

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
                  onChange={(e) => {
                    setApiKey(e.target.value);
                    if (validationError) setValidationError(null);
                  }}
                  className="bg-bg-secondary border-border-subtle focus:border-border-default"
                  disabled={isConnecting}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && !isConnecting) {
                      e.preventDefault();
                      handleConnect();
                    }
                  }}
                />
              </div>
              <div className="flex items-start gap-2 p-3 rounded-lg bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-800/50">
                <Shield className="w-4 h-4 text-amber-600 dark:text-amber-400 mt-0.5 shrink-0" />
                <p className="text-xs text-amber-950 dark:text-amber-100">
                  This allows FunctionFly to deploy to your {provider.name} account. You can revoke
                  access anytime. API keys are encrypted with AES-256-GCM.
                </p>
              </div>
            </>
          )}

          <button
            type="button"
            onClick={handleConnect}
            disabled={(!isFunctionFly && !apiKey.trim()) || isConnecting}
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
  const [disconnectConfirmOpen, setDisconnectConfirmOpen] = useState(false);
  const [disconnectingProvider, setDisconnectingProvider] = useState<{
    id: string;
    name: string;
  } | null>(null);
  const [validationError, setValidationError] = useState<string | null>(null);
  const [testingProvider, setTestingProvider] = useState<string | null>(null);
  const [connectionTestResults, setConnectionTestResults] = useState<
    Record<string, 'success' | 'error' | null>
  >({});
  const {
    providers,
    error,
    isLoading,
    fetchProviders,
    connectProvider,
    disconnectProvider,
    testConnection,
    clearError,
    startHealthCheckPolling,
    stopHealthCheckPolling,
  } = useProvidersStore();

  useEffect(() => {
    fetchProviders();
  }, [fetchProviders]);

  // Start health check polling when providers are loaded
  useEffect(() => {
    if (providers.length > 0) {
      const stopPolling = startHealthCheckPolling(5 * 60 * 1000); // 5 minute interval
      return () => {
        stopPolling();
      };
    }
  }, [providers.length, startHealthCheckPolling]);

  const handleConnect = async (providerId: string, key?: string) => {
    const isFunctionFly = providerId === 'functionfly-edge';
    const providerKey = key ?? '';

    // Clear previous validation errors
    setValidationError(null);
    clearError();

    // Client-side validation
    if (!isFunctionFly) {
      if (!providerKey.trim()) {
        setValidationError('API key is required');
        throw new Error('API key is required');
      }
      if (providerKey.length < 10) {
        setValidationError('API key must be at least 10 characters');
        throw new Error('API key must be at least 10 characters');
      }
    }

    setConnecting(true);

    try {
      await connectProvider({ providerId, apiKey: providerKey });
      // Refresh from server to get accurate state
      await fetchProviders();
      setApiKey('');
      setValidationError(null);
      // Clear any connection test results for this provider
      setConnectionTestResults((prev) => ({ ...prev, [providerId]: null }));
    } catch (error) {
      console.error('Failed to connect provider:', error);
      throw error;
    } finally {
      setConnecting(false);
    }
  };

  const handleTestConnection = async (providerId: string) => {
    setTestingProvider(providerId);
    setConnectionTestResults((prev) => ({ ...prev, [providerId]: null }));

    try {
      const isSuccess = await testConnection(providerId);
      const status = isSuccess ? 'success' : 'error';
      setConnectionTestResults((prev) => ({ ...prev, [providerId]: status }));

      // Auto-clear success message after 3 seconds
      if (isSuccess) {
        setTimeout(() => {
          setConnectionTestResults((prev) => ({ ...prev, [providerId]: null }));
        }, 3000);
      }
    } catch (error) {
      setConnectionTestResults((prev) => ({ ...prev, [providerId]: 'error' }));
    } finally {
      setTestingProvider(null);
    }
  };

  const openDisconnectConfirm = (catalogProviderId: string) => {
    const row = providers.find((p) => p.name === catalogProviderId);
    if (!row) {
      console.warn('Disconnect: no connected provider for', catalogProviderId);
      return;
    }
    setDisconnectingProvider({ id: row.id, name: row.name });
    setDisconnectConfirmOpen(true);
  };

  const handleDisconnectConfirm = async () => {
    if (!disconnectingProvider) return;

    setDisconnecting(disconnectingProvider.name);
    clearError();

    try {
      await disconnectProvider(disconnectingProvider.id);
      setDisconnectConfirmOpen(false);
      setDisconnectingProvider(null);
    } catch (error) {
      console.error('Failed to disconnect provider:', error);
    } finally {
      setDisconnecting(null);
    }
  };

  const handleDisconnectCancel = () => {
    setDisconnectConfirmOpen(false);
    setDisconnectingProvider(null);
  };

  const isConnected = (catalogProviderId: string) =>
    providers.some((p) => p.name === catalogProviderId);

  const getProviderStatus = (catalogProviderId: string) => {
    const connected = providers.find((p) => p.name === catalogProviderId);
    return connected?.status || 'pending';
  };

  const getProviderData = (catalogProviderId: string) => {
    return providers.find((p) => p.name === catalogProviderId);
  };

  return (
    <div className="space-y-8 animate-in fade-in duration-300">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-text-primary tracking-tight">Providers</h1>
          <p className="text-text-secondary mt-1">Connect and manage your deployment targets</p>
        </div>
        <div className="flex items-center gap-3">
          {isLoading && (
            <div className="flex items-center gap-2 text-sm text-text-tertiary">
              <Loader2 className="w-4 h-4 animate-spin" />
              <span>Loading...</span>
            </div>
          )}
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
      </div>

      {/* Error Message with Retry */}
      {error && (
        <div className="p-4 bg-error/10 border border-error/20 rounded-lg animate-in slide-in-from-top-2 duration-200">
          <div className="flex items-start gap-3">
            <AlertCircle className="w-5 h-5 text-error mt-0.5 shrink-0" />
            <div className="flex-1">
              <p className="text-error font-medium">Failed to load providers</p>
              <p className="text-error/80 text-sm mt-1">{error}</p>
              <div className="flex gap-2 mt-3">
                <Button
                  variant="outline"
                  size="sm"
                  className="border-error/30 text-error hover:bg-error/10 hover:text-error"
                  onClick={() => fetchProviders()}
                  disabled={isLoading}
                >
                  <RefreshCw className={`w-4 h-4 mr-1.5 ${isLoading ? 'animate-spin' : ''}`} />
                  Retry
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-error/70 hover:text-error hover:bg-error/5"
                  onClick={clearError}
                >
                  Dismiss
                </Button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Providers Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-5">
        {isLoading && providers.length === 0 ? (
          // Loading skeletons
          <>
            <ProviderCardSkeleton />
            <ProviderCardSkeleton />
            <ProviderCardSkeleton />
            <ProviderCardSkeleton />
            <ProviderCardSkeleton />
          </>
        ) : (
          Object.values(PROVIDERS).map((provider) => {
            const connected = isConnected(provider.id);
            const status = getProviderStatus(provider.id);
            const providerData = getProviderData(provider.id);

            return (
              <ProviderCard
                key={provider.id}
                provider={provider}
                connected={connected}
                status={status}
                onDisconnect={() => openDisconnectConfirm(provider.id)}
                onConnect={async (pid, key) => handleConnect(pid, key)}
                onTestConnection={connected ? () => handleTestConnection(provider.id) : undefined}
                isDisconnecting={disconnecting === provider.id}
                isTestingConnection={testingProvider === provider.id}
                lastUsedAt={providerData?.lastUsedAt}
                isStale={providerData?.isStale}
                connectionTestResult={connectionTestResults[provider.id]}
              />
            );
          })
        )}
      </div>

      {/* Disconnect Confirmation Dialog */}
      {disconnectingProvider && (
        <DisconnectConfirmationDialog
          providerName={
            PROVIDERS[disconnectingProvider.name.toUpperCase() as keyof typeof PROVIDERS]?.name ||
            disconnectingProvider.name
          }
          isOpen={disconnectConfirmOpen}
          onClose={handleDisconnectCancel}
          onConfirm={handleDisconnectConfirm}
          isDisconnecting={!!disconnecting}
        />
      )}
    </div>
  );
}
