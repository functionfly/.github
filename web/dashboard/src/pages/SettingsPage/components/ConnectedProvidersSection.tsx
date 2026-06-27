import { providersApi } from '@/api/providers';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { useQuery } from '@tanstack/react-query';
import { AlertTriangle, Cloud, ExternalLink, Loader2, RefreshCw, Trash2 } from 'lucide-react';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { toast } from 'sonner';
import { formatDate } from '../settings-utils';

interface ProviderCredential {
  id: string;
  name: string;
  status: string;
  connectedAt: string;
  maskedApiKey: string;
  isStale?: boolean;
  lastUsedAt?: string;
}

export function ConnectedProvidersSection() {
  const { t } = useTranslation();
  const [disconnectingId, setDisconnectingId] = useState<string | null>(null);
  const [confirmDisconnectId, setConfirmDisconnectId] = useState<string | null>(null);
  const [confirmDisconnectName, setConfirmDisconnectName] = useState('');

  const {
    data: providersData,
    isLoading,
    refetch: refetchProviders,
  } = useQuery({
    queryKey: ['provider-credentials'],
    queryFn: async () => {
      try {
        return await providersApi.getProviderCredentials();
      } catch (e: unknown) {
        const status = (e as { response?: { status?: number } })?.response?.status;
        if (status === 404) return [];
        throw e;
      }
    },
    enabled: true,
    retry: false,
  });

  const providers: ProviderCredential[] = Array.isArray(providersData) ? providersData : [];

  const handleDisconnectClick = (provider: ProviderCredential) => {
    setConfirmDisconnectId(provider.id);
    setConfirmDisconnectName(provider.name);
  };

  const handleDisconnectConfirm = async () => {
    if (!confirmDisconnectId) return;
    setDisconnectingId(confirmDisconnectId);
    try {
      await providersApi.disconnectProvider(confirmDisconnectId);
      refetchProviders();
      toast.success(t('connectedProvidersSettings.toastDisconnected'));
      setConfirmDisconnectId(null);
      setConfirmDisconnectName('');
    } catch (err: unknown) {
      toast.error(t('connectedProvidersSettings.errorFailedToDisconnect'));
    } finally {
      setDisconnectingId(null);
    }
  };

  const getStatusBadge = (status: string, isStale: boolean) => {
    if (status === 'active' || status === 'online') {
      return isStale ? (
        <Badge variant="secondary">{t('connectedProvidersSettings.stale')}</Badge>
      ) : (
        <Badge className="ff-badge-success">{t('connectedProvidersSettings.active')}</Badge>
      );
    }
    if (status === 'error' || status === 'degraded') {
      return <Badge variant="destructive">{t('connectedProvidersSettings.error')}</Badge>;
    }
    return <Badge variant="secondary">{status}</Badge>;
  };

  const getProviderIcon = (name: string) => {
    const iconProps = { className: 'h-4 w-4' };
    switch (name) {
      case 'cloudflare':
      case 'workers':
        return <Cloud {...iconProps} />;
      case 'vercel':
        return (
          <svg viewBox="0 0 24 24" fill="currentColor" className="h-4 w-4">
            <path d="M24 22.525h-9.943l-4.358-14.4a1.317 1.317 0 00-2.365-.001L2.058 22.52H0l8.018-24.002a1.318 1.318 0 012.366 0L24 22.525z" />
          </svg>
        );
      case 'fly':
        return (
          <svg viewBox="0 0 24 24" fill="currentColor" className="h-4 w-4">
            <path d="M12.005 0L2.755 6.003V18.02L12.005 24l9.25-5.988V6.033L12.005 0zm4.5 12.005l-4.5 2.988-4.5-2.988v-5.01L12.005 4.995l4.5 2.995v5.01z" />
          </svg>
        );
      case 'functionfly-edge':
        return <Cloud {...iconProps} />;
      default:
        return <Cloud {...iconProps} />;
    }
  };

  return (
    <>
      <div
        className="rounded-lg p-5"
        style={{
          background: 'var(--panel)',
          border: '1px solid var(--panel-edge)',
          boxShadow: 'var(--shadow-chamber)',
        }}
      >
        <div className="mb-4">
          <h3 className="font-display text-lg font-semibold" style={{ color: 'var(--text)' }}>
            {t('connectedProvidersSettings.title')}
          </h3>
          <p className="text-sm mt-1" style={{ color: 'var(--text-dim)' }}>
            {t('connectedProvidersSettings.description')}
          </p>
        </div>
        <div className="space-y-4">
          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin" style={{ color: 'var(--text-dim)' }} />
            </div>
          ) : providers.length === 0 ? (
            <div
              className="rounded-lg p-6 text-center"
              style={{ background: 'var(--panel-raised)', border: '1px solid var(--panel-edge)' }}
            >
              <p className="text-sm" style={{ color: 'var(--text-dim)' }}>
                {t('connectedProvidersSettings.noProvidersYet')}
              </p>
              <p className="mt-1 text-sm" style={{ color: 'var(--text-dim)' }}>
                {t('connectedProvidersSettings.connectFirst')}
              </p>
              <Button
                className="mt-4 gap-2"
                asChild
                style={{
                  background: 'linear-gradient(180deg, #ffffff, #d8dee2)',
                  color: 'var(--text-on-light)',
                  boxShadow: 'var(--shadow-btn-primary-rest)',
                }}
              >
                <Link to="/providers">
                  {t('connectedProvidersSettings.connectProviders')}
                  <ExternalLink className="h-4 w-4" />
                </Link>
              </Button>
            </div>
          ) : (
            <>
              {providers.map((provider) => (
                <div
                  key={provider.id}
                  className="flex flex-wrap items-center justify-between gap-3 rounded-lg p-4"
                  style={{
                    background: 'var(--panel-raised)',
                    border: '1px solid var(--panel-edge)',
                  }}
                >
                  <div className="flex items-center gap-3 min-w-0">
                    <div
                      className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl"
                      style={{ background: 'var(--panel)' }}
                    >
                      {getProviderIcon(provider.name)}
                    </div>
                    <div className="min-w-0">
                      <h4 className="font-medium capitalize" style={{ color: 'var(--text)' }}>
                        {provider.name === 'functionfly-edge'
                          ? 'FunctionFly Edge'
                          : provider.name === 'workers'
                            ? 'Cloudflare Workers'
                            : provider.name}
                      </h4>
                      <p className="text-sm" style={{ color: 'var(--text-dim)' }}>
                        {t('connectedProvidersSettings.connectedLabel', {
                          date: provider.connectedAt ? formatDate(provider.connectedAt) : '—',
                        })}
                        {provider.lastUsedAt &&
                          ` · ${t('connectedProvidersSettings.lastUsedLabel', { date: formatDate(provider.lastUsedAt) })}`}
                      </p>
                    </div>
                  </div>
                  <div className="flex flex-wrap items-center gap-2">
                    {getStatusBadge(provider.status, provider.isStale)}
                    <code
                      className="rounded px-3 py-1 text-sm"
                      style={{ background: 'var(--panel)', color: 'var(--text-dim)' }}
                    >
                      {provider.maskedApiKey}
                    </code>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleDisconnectClick(provider)}
                      title={t('connectedProvidersSettings.disconnectProvider')}
                      style={{ color: 'var(--status-revoked)' }}
                      disabled={disconnectingId === provider.id}
                    >
                      {disconnectingId === provider.id ? (
                        <Loader2 className="h-4 w-4 animate-spin" />
                      ) : (
                        <Trash2 className="h-4 w-4" />
                      )}
                    </Button>
                  </div>
                </div>
              ))}
              <div
                className="flex items-center justify-between rounded-lg p-3"
                style={{ background: 'var(--panel-raised)', border: '1px solid var(--panel-edge)' }}
              >
                <p className="text-sm" style={{ color: 'var(--text-dim)' }}>
                  {t('connectedProvidersSettings.manageProvidersNote')}
                </p>
                <Button
                  variant="outline"
                  size="sm"
                  className="gap-2"
                  asChild
                  style={{ borderColor: 'var(--steel)', color: 'var(--text)' }}
                >
                  <Link to="/providers">
                    <RefreshCw className="h-4 w-4" />
                    {t('connectedProvidersSettings.manageProviders')}
                  </Link>
                </Button>
              </div>
            </>
          )}
        </div>
      </div>

      <Dialog
        open={!!confirmDisconnectId}
        onOpenChange={(open) => !open && setConfirmDisconnectId(null)}
      >
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{t('connectedProvidersSettings.disconnectDialogTitle')}</DialogTitle>
            <DialogDescription>
              {t('connectedProvidersSettings.disconnectDialogDescription', {
                name: confirmDisconnectName,
              })}
            </DialogDescription>
          </DialogHeader>
          <div
            className="flex items-center gap-2 rounded-lg p-3 text-sm"
            style={{
              background: 'rgba(232, 196, 104, 0.06)',
              border: '1px solid rgba(232, 196, 104, 0.3)',
              color: 'var(--status-pending)',
            }}
          >
            <AlertTriangle className="h-5 w-5 shrink-0" />
            <span>{t('connectedProvidersSettings.disconnectWarning')}</span>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setConfirmDisconnectId(null)}
              style={{ borderColor: 'var(--steel)', color: 'var(--text)' }}
            >
              {t('connectedProvidersSettings.cancel')}
            </Button>
            <Button
              variant="destructive"
              onClick={handleDisconnectConfirm}
              disabled={!!disconnectingId}
            >
              {disconnectingId
                ? t('connectedProvidersSettings.disconnecting')
                : t('connectedProvidersSettings.disconnect')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
