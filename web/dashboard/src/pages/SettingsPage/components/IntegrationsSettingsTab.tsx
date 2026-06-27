import { connectorsApi, type Connector, type UserConnector } from '@/api/connectors';
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
} from '@/components/ui/dialog';
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet';
import { Switch } from '@/components/ui/switch';
import { useToast } from '@/components/ui/use-toast';
import { cn } from '@/lib/utils';
import { AnimatePresence, motion } from 'framer-motion';
import {
  AlertCircle,
  AlertTriangle,
  CheckCircle2,
  Clock,
  FileText,
  Info,
  Link2,
  Loader2,
  Mail,
  MessageSquare,
  RefreshCw,
  Settings2,
  Shield,
  Unlink,
  Zap,
} from 'lucide-react';
import { useCallback, useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { siGithub, siLinear } from 'simple-icons';

// ─── OAuth Callback Message Types ─────────────────────────────────────────────

interface OAuthCallbackMessage {
  type: 'oauth_callback';
  status: 'success' | 'error';
  connector_slug?: string;
  connector_name?: string;
  message?: string;
}

// ─── Types ───────────────────────────────────────────────────────────────────

type SyncFrequency = '5m' | '15m' | '1h' | '6h' | '24h';

interface ConnectorSettings {
  displayName: string;
  syncFrequency: SyncFrequency;
  autoSync: boolean;
}

interface OAuthUrlResponse {
  oauth_url: string;
  state: string;
}

// ─── Constants ────────────────────────────────────────────────────────────────

const SYNC_FREQUENCIES: { value: SyncFrequency; label: string }[] = [
  { value: '5m', label: 'Every 5 min' },
  { value: '15m', label: 'Every 15 min' },
  { value: '1h', label: 'Every hour' },
  { value: '6h', label: 'Every 6 hours' },
  { value: '24h', label: 'Once a day' },
];

const connectorIcons: Record<string, React.ReactNode> = {
  github: (
    <svg viewBox="0 0 24 24" className="w-5 h-5" fill="currentColor" aria-hidden="true">
      <path d={siGithub.path} />
    </svg>
  ),
  notion: <FileText className="w-5 h-5" aria-hidden="true" />,
  slack: <MessageSquare className="w-5 h-5" aria-hidden="true" />,
  gmail: <Mail className="w-5 h-5" aria-hidden="true" />,
  linear: (
    <svg viewBox="0 0 24 24" className="w-5 h-5" fill="currentColor" aria-hidden="true">
      <path d={siLinear.path} />
    </svg>
  ),
};

const connectorColors: Record<string, string> = {
  github: 'from-gray-700 to-gray-900',
  notion: 'from-gray-800 to-black',
  slack: 'from-purple-600 to-purple-800',
  gmail: 'from-red-600 to-red-800',
  linear: 'from-blue-600 to-blue-800',
};

const connectorAccents: Record<string, string> = {
  github: 'text-gray-400',
  notion: 'text-blue-400',
  slack: 'text-purple-400',
  gmail: 'text-red-400',
  linear: 'text-blue-400',
};

const statusConfig: Record<string, { label: string; color: string; icon: React.ReactNode }> = {
  active: {
    label: 'Active',
    color: 'bg-green-500/10 text-green-400 border-green-500/20',
    icon: <CheckCircle2 className="w-3.5 h-3.5" aria-hidden="true" />,
  },
  disabled: {
    label: 'Disabled',
    color: 'bg-yellow-500/10 text-yellow-400 border-yellow-500/20',
    icon: <AlertCircle className="w-3.5 h-3.5" aria-hidden="true" />,
  },
  sync_error: {
    label: 'Sync Error',
    color: 'bg-red-500/10 text-red-400 border-red-500/20',
    icon: <AlertCircle className="w-3.5 h-3.5" aria-hidden="true" />,
  },
  revoked: {
    label: 'Revoked',
    color: 'bg-yellow-500/10 text-yellow-400 border-yellow-500/20',
    icon: <AlertCircle className="w-3.5 h-3.5" aria-hidden="true" />,
  },
};

// ─── Helpers ─────────────────────────────────────────────────────────────────

function formatRelativeTime(dateStr?: string): string {
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
}

function openOAuthPopup(url: string, width = 500, height = 700): Window | null {
  const left = Math.round(window.screenX + (window.outerWidth - width) / 2);
  const top = Math.round(window.screenY + (window.outerHeight - height) / 2);
  return window.open(
    url,
    'oauth_popup',
    `width=${width},height=${height},left=${left},top=${top},toolbar=no,menubar=no`
  );
}

// ─── Connector Row Component ──────────────────────────────────────────────────

interface ConnectorRowProps {
  connector: Connector;
  userConnector: UserConnector | undefined;
  onToggle: (enabled: boolean) => void;
  onConfigure: () => void;
  onSync: () => void;
  onDisconnect: () => void;
  syncing: boolean;
  toggling: boolean;
}

function ConnectorRow({
  connector,
  userConnector,
  onToggle,
  onConfigure,
  onSync,
  onDisconnect,
  syncing,
  toggling,
}: ConnectorRowProps) {
  const { t } = useTranslation();
  const isLinked = !!userConnector;
  const isEnabled = isLinked && userConnector?.status !== 'disabled';
  const status = userConnector ? statusConfig[userConnector.status] || statusConfig.active : null;
  const icon = connectorIcons[connector.slug] || <Link2 className="w-5 h-5" aria-hidden="true" />;
  const gradient = connectorColors[connector.slug] || 'from-gray-600 to-gray-800';
  const accent = connectorAccents[connector.slug] || 'text-brand-400';

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -12 }}
      layout
    >
      <div
        className={cn(
          'relative flex items-center gap-4 rounded-xl border p-4 transition-colors duration-200',
          isLinked
            ? 'border-white/[0.06] bg-white/[0.02] hover:bg-white/[0.04]'
            : 'border-white/[0.04] bg-white/[0.01] hover:bg-white/[0.03]'
        )}
      >
        {/* Gradient accent bar */}
        {isLinked && (
          <div
            className={cn(
              'absolute top-0 left-0 h-full w-0.5 bg-gradient-to-b rounded-full',
              gradient
            )}
          />
        )}

        {/* Icon */}
        <div
          className={cn(
            'flex items-center justify-center w-11 h-11 rounded-xl shrink-0',
            isLinked ? `bg-gradient-to-br ${gradient} text-white` : 'bg-bg-tertiary text-text-muted'
          )}
        >
          {icon}
        </div>

        {/* Name + meta */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium text-text-primary truncate">{connector.name}</span>
            {!isLinked && (
              <Badge
                variant="outline"
                className="text-[10px] bg-white/5 border-white/10 text-text-muted shrink-0"
              >
                Not connected
              </Badge>
            )}
            {status && isLinked && (
              <Badge variant="outline" className={cn('shrink-0', status.color)}>
                <span className="flex items-center gap-1">
                  {status.icon}
                  {status.label}
                </span>
              </Badge>
            )}
          </div>
          <div className="flex items-center gap-3 mt-0.5">
            <span className={cn('text-xs', accent)}>
              {connector.scopes.split(',').length} permissions
            </span>
            {isLinked && userConnector?.last_sync_at && (
              <>
                <span className="text-text-muted">·</span>
                <span className="text-xs text-text-muted flex items-center gap-1">
                  <Clock className="w-3 h-3" aria-hidden="true" />
                  {t('integrationsSettings.lastSynced', {
                    time: formatRelativeTime(userConnector.last_sync_at),
                  })}
                </span>
              </>
            )}
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2 shrink-0">
          {isLinked && (
            <>
              {/* Sync */}
              <Button
                variant="ghost"
                size="sm"
                onClick={onSync}
                disabled={syncing || !isEnabled}
                style={{ color: 'var(--text-dim)' }}
                aria-label={t('integrationsSettings.syncNow')}
              >
                {syncing ? (
                  <Loader2 className="w-4 h-4 animate-spin" aria-hidden="true" />
                ) : (
                  <RefreshCw className="w-4 h-4" aria-hidden="true" />
                )}
              </Button>

              {/* Configure */}
              <Button
                variant="ghost"
                size="sm"
                onClick={onConfigure}
                style={{ color: 'var(--text-dim)' }}
                aria-label={t('integrationsSettings.configure')}
              >
                <Settings2 className="w-4 h-4" aria-hidden="true" />
              </Button>

              {/* Toggle */}
              <Switch
                checked={isEnabled}
                onCheckedChange={onToggle}
                disabled={toggling}
                aria-label={t('integrationsSettings.toggle', { name: connector.name })}
              />

              {/* Disconnect */}
              <Button
                variant="ghost"
                size="sm"
                onClick={onDisconnect}
                style={{ color: 'var(--status-revoked)' }}
                aria-label={t('integrationsSettings.disconnect')}
              >
                <Unlink className="w-4 h-4" aria-hidden="true" />
              </Button>
            </>
          )}

          {!isLinked && (
            <Button
              size="sm"
              onClick={onConfigure}
              className="gap-1.5 text-xs"
              style={{
                background: 'linear-gradient(180deg, #ffffff, #d8dee2)',
                color: 'var(--text-on-light)',
                boxShadow: 'var(--shadow-btn-primary-rest)',
              }}
              aria-label={t('integrationsSettings.connect', { name: connector.name })}
            >
              <Zap className="w-3.5 h-3.5" aria-hidden="true" />
              {t('integrationsSettings.connect')}
            </Button>
          )}
        </div>
      </div>
    </motion.div>
  );
}

// ─── Config Drawer Component ──────────────────────────────────────────────────

interface ConfigDrawerProps {
  connector: Connector;
  userConnector: UserConnector | undefined;
  open: boolean;
  onClose: () => void;
  onSave: (settings: ConnectorSettings) => Promise<void>;
  onConnect: () => void;
  onDisconnect: () => void;
  saving: boolean;
  connecting: boolean;
}

function ConfigDrawer({
  connector,
  userConnector,
  open,
  onClose,
  onSave,
  onConnect,
  onDisconnect,
  saving,
  connecting,
}: ConfigDrawerProps) {
  const { t } = useTranslation();
  const isLinked = !!userConnector;
  const isEnabled = isLinked && userConnector?.status !== 'disabled';
  const gradient = connectorColors[connector.slug] || 'from-gray-600 to-gray-800';
  const icon = connectorIcons[connector.slug] || <Link2 className="w-5 h-5" aria-hidden="true" />;

  const [displayName, setDisplayName] = useState(userConnector?.display_name || connector.name);
  const [syncFrequency, setSyncFrequency] = useState<SyncFrequency>(
    (userConnector as any)?.sync_frequency || '1h'
  );
  const [autoSync, setAutoSync] = useState((userConnector as any)?.auto_sync ?? true);

  // Reset form when userConnector changes
  useEffect(() => {
    setDisplayName(userConnector?.display_name || connector.name);
    setSyncFrequency((userConnector as any)?.sync_frequency || '1h');
    setAutoSync((userConnector as any)?.auto_sync ?? true);
  }, [userConnector, connector.name]);

  const handleSave = async () => {
    await onSave({ displayName, syncFrequency, autoSync });
  };

  return (
    <Sheet open={open} onOpenChange={(v) => !v && onClose()}>
      <SheetContent
        side="right"
        className="w-full sm:max-w-md bg-slate-900 border-white/10 flex flex-col"
      >
        <SheetHeader className="shrink-0">
          <div className="flex items-center gap-3">
            <div
              className={cn(
                'flex items-center justify-center w-10 h-10 rounded-xl bg-gradient-to-br text-white',
                gradient
              )}
            >
              {icon}
            </div>
            <div>
              <SheetTitle className="text-text-primary text-base">{connector.name}</SheetTitle>
              <SheetDescription className="text-xs text-text-muted mt-0.5">
                {isLinked
                  ? t('integrationsSettings.configTitleConnected')
                  : t('integrationsSettings.configTitleAvailable')}
              </SheetDescription>
            </div>
          </div>
        </SheetHeader>

        <div className="flex-1 overflow-y-auto space-y-5 py-6">
          {/* Status */}
          {isLinked && (
            <div className="flex items-center justify-between rounded-lg border border-white/[0.06] bg-white/[0.02] p-3">
              <div className="flex items-center gap-2">
                <div
                  className={cn(
                    'w-2 h-2 rounded-full',
                    isEnabled ? 'bg-green-400 animate-pulse' : 'bg-yellow-400'
                  )}
                />
                <span className="text-sm text-text-primary">
                  {isEnabled ? 'Sync active' : 'Sync paused'}
                </span>
              </div>
              {userConnector?.last_sync_at && (
                <span className="text-xs text-text-muted">
                  {t('integrationsSettings.lastSynced', {
                    time: formatRelativeTime(userConnector.last_sync_at),
                  })}
                </span>
              )}
            </div>
          )}

          {/* Connect CTA */}
          {!isLinked && (
            <div className="space-y-3">
              <div className="rounded-lg border border-indigo-500/20 bg-indigo-500/5 p-3 flex items-start gap-2">
                <Shield className="w-4 h-4 text-indigo-400 mt-0.5 shrink-0" aria-hidden="true" />
                <div className="text-xs text-text-secondary">
                  <span className="text-indigo-300 font-medium">
                    {t('integrationsSettings.zeroKnowledgeNotice')}
                  </span>{' '}
                  {t('integrationsSettings.configEncryptNote')}
                </div>
              </div>

              <div className="rounded-lg border border-white/[0.06] bg-white/[0.02] p-3 space-y-2">
                <h4 className="text-xs font-medium text-text-primary">
                  {t('integrationsSettings.requestedPermissions')}
                </h4>
                <div className="flex flex-wrap gap-1.5">
                  {connector.scopes.split(',').map((scope) => (
                    <Badge
                      key={scope}
                      variant="outline"
                      className="text-[10px] bg-white/5 border-white/10"
                    >
                      {scope.trim()}
                    </Badge>
                  ))}
                </div>
              </div>
            </div>
          )}

          {/* Settings (only when linked) */}
          {isLinked && (
            <div className="space-y-4">
              {/* Display name */}
              <div className="space-y-1.5">
                <label htmlFor="display-name" className="text-xs font-medium text-text-secondary">
                  {t('integrationsSettings.displayName')}
                </label>
                <input
                  id="display-name"
                  type="text"
                  value={displayName}
                  onChange={(e) => setDisplayName(e.target.value)}
                  className="w-full h-9 rounded-md border border-white/[0.06] bg-white/[0.03] px-3 text-sm text-text-primary placeholder:text-text-muted focus:outline-none focus:ring-2 focus:ring-brand-500/50 focus:border-brand-500/50 transition-colors"
                  placeholder={connector.name}
                />
              </div>

              {/* Sync frequency */}
              <div className="space-y-1.5">
                <label htmlFor="sync-frequency" className="text-xs font-medium text-text-secondary">
                  {t('integrationsSettings.syncFrequency')}
                </label>
                <select
                  id="sync-frequency"
                  value={syncFrequency}
                  onChange={(e) => setSyncFrequency(e.target.value as SyncFrequency)}
                  className="w-full h-9 rounded-md border border-white/[0.06] bg-white/[0.03] px-3 text-sm text-text-primary focus:outline-none focus:ring-2 focus:ring-brand-500/50 transition-colors appearance-none cursor-pointer"
                >
                  {SYNC_FREQUENCIES.map((f) => (
                    <option key={f.value} value={f.value}>
                      {f.label}
                    </option>
                  ))}
                </select>
              </div>

              {/* Auto sync */}
              <div className="flex items-center justify-between rounded-lg border border-white/[0.06] bg-white/[0.02] p-3">
                <div>
                  <p className="text-sm text-text-primary">{t('integrationsSettings.autoSync')}</p>
                  <p className="text-xs text-text-muted mt-0.5">
                    {t('integrationsSettings.autoSyncDescription')}
                  </p>
                </div>
                <Switch
                  checked={autoSync}
                  onCheckedChange={setAutoSync}
                  aria-label={t('integrationsSettings.autoSync')}
                />
              </div>

              {/* Permissions */}
              <div className="space-y-1.5">
                <p className="text-xs font-medium text-text-secondary">
                  {t('integrationsSettings.permissionsGranted')}
                </p>
                <div className="flex flex-wrap gap-1.5">
                  {connector.scopes.split(',').map((scope) => (
                    <Badge
                      key={scope}
                      variant="outline"
                      className="text-[10px] bg-white/5 border-white/10"
                    >
                      {scope.trim()}
                    </Badge>
                  ))}
                </div>
              </div>

              {/* Sync error */}
              {userConnector?.sync_error && (
                <div className="flex items-start gap-2 rounded-lg border border-red-500/20 bg-red-500/5 p-3">
                  <AlertCircle
                    className="w-4 h-4 text-red-400 mt-0.5 shrink-0"
                    aria-hidden="true"
                  />
                  <div>
                    <p className="text-xs font-medium text-red-400">
                      {t('integrationsSettings.syncError')}
                    </p>
                    <p className="text-xs text-red-300/70 mt-0.5">{userConnector.sync_error}</p>
                  </div>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="shrink-0 pt-4 border-t border-white/[0.06] space-y-2">
          {isLinked ? (
            <>
              <Button onClick={handleSave} disabled={saving} className="w-full">
                {saving ? (
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" aria-hidden="true" />
                ) : null}
                {t('integrationsSettings.saveSettings')}
              </Button>
              <Button
                variant="ghost"
                onClick={onDisconnect}
                className="w-full text-red-400 hover:text-red-300 hover:bg-red-500/10"
              >
                <Unlink className="w-4 h-4 mr-2" aria-hidden="true" />
                {t('integrationsSettings.disconnectConnector')}
              </Button>
            </>
          ) : (
            <Button onClick={onConnect} disabled={connecting} className="w-full gap-2">
              {connecting ? (
                <Loader2 className="w-4 h-4 animate-spin" aria-hidden="true" />
              ) : (
                <Zap className="w-4 h-4" aria-hidden="true" />
              )}
              {t('integrationsSettings.authorizeAndConnect')}
            </Button>
          )}
        </div>
      </SheetContent>
    </Sheet>
  );
}

// ─── Disconnect Confirm Dialog ────────────────────────────────────────────────

interface DisconnectDialogProps {
  open: boolean;
  connectorName: string;
  onConfirm: () => void;
  onCancel: () => void;
  disconnecting: boolean;
}

function DisconnectDialog({
  open,
  connectorName,
  onConfirm,
  onCancel,
  disconnecting,
}: DisconnectDialogProps) {
  const { t } = useTranslation();
  return (
    <Dialog open={open} onOpenChange={(v) => !v && onCancel()}>
      <DialogContent className="bg-slate-900 border-white/10">
        <DialogHeader>
          <DialogTitle>
            {t('integrationsSettings.disconnectConfirmTitle', { name: connectorName })}
          </DialogTitle>
          <DialogDescription>
            {t('integrationsSettings.disconnectConfirmDescription', { name: connectorName })}
          </DialogDescription>
        </DialogHeader>
        <div className="flex items-center gap-2 rounded-lg border border-amber-500/50 bg-amber-500/10 p-3 text-sm text-amber-400">
          <AlertTriangle className="h-5 w-5 shrink-0" aria-hidden="true" />
          <span>{t('integrationsSettings.disconnectWarning')}</span>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onCancel}>
            {t('integrationsSettings.cancel')}
          </Button>
          <Button variant="destructive" onClick={onConfirm} disabled={disconnecting}>
            {disconnecting ? (
              <Loader2 className="w-4 h-4 mr-2 animate-spin" aria-hidden="true" />
            ) : (
              <Unlink className="w-4 h-4 mr-2" aria-hidden="true" />
            )}
            {t('integrationsSettings.disconnect')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// ─── Main Component ───────────────────────────────────────────────────────────

export function IntegrationsSettingsTab() {
  const { t } = useTranslation();
  const { toast } = useToast();

  const [catalog, setCatalog] = useState<Connector[]>([]);
  const [userConnectors, setUserConnectors] = useState<UserConnector[]>([]);
  const [loading, setLoading] = useState(true);

  // Drawer state
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [selectedConnector, setSelectedConnector] = useState<Connector | null>(null);
  const [saving, setSaving] = useState(false);
  const [connecting, setConnecting] = useState(false);
  const [togglingId, setTogglingId] = useState<string | null>(null);
  const [syncingId, setSyncingId] = useState<string | null>(null);

  // Disconnect dialog
  const [disconnectOpen, setDisconnectOpen] = useState(false);
  const [disconnectingId, setDisconnectingId] = useState<string | null>(null);
  const [disconnectingName, setDisconnectingName] = useState('');
  const [disconnecting, setDisconnecting] = useState(false);

  // Popup ref for OAuth
  const oauthPopupRef = useRef<Window | null>(null);
  const oauthCheckIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const [cat, uc] = await Promise.all([
        connectorsApi.listCatalog(),
        connectorsApi.listUserConnectors(),
      ]);
      setCatalog(cat ?? []);
      setUserConnectors(uc ?? []);
    } catch (err) {
      console.error('Failed to load connectors:', err);
      toast({
        title: t('integrationsSettings.loadError'),
        description: t('integrationsSettings.loadErrorDescription'),
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  }, [t, toast]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  // Cleanup OAuth popup interval on unmount
  useEffect(() => {
    return () => {
      if (oauthCheckIntervalRef.current) {
        clearInterval(oauthCheckIntervalRef.current);
      }
    };
  }, []);

  // ─── OAuth postMessage listener ────────────────────────────────────────────
  // The backend callback page sends a postMessage when OAuth completes,
  // allowing instant feedback without waiting for popup.close polling.
  useEffect(() => {
    const handleMessage = (event: MessageEvent) => {
      // Validate the message is an OAuth callback
      const data = event.data as OAuthCallbackMessage;
      if (!data || data.type !== 'oauth_callback') return;

      // Clear the popup polling interval since we got a direct message
      if (oauthCheckIntervalRef.current) {
        clearInterval(oauthCheckIntervalRef.current);
        oauthCheckIntervalRef.current = null;
      }

      if (data.status === 'success') {
        toast({
          title: t('integrationsSettings.connected', {
            name: data.connector_name || data.connector_slug || 'Connector',
          }),
          description: data.message || t('integrationsSettings.connectedDescription'),
        });
        loadData();
      } else {
        toast({
          title: t('integrationsSettings.oauthFailed'),
          description: data.message || t('integrationsSettings.oauthFailedDescription'),
          variant: 'destructive',
        });
      }
    };

    window.addEventListener('message', handleMessage);
    return () => window.removeEventListener('message', handleMessage);
  }, [toast, t, loadData]);

  // ─── OAuth Connect ──────────────────────────────────────────────────────────

  const handleConnectOAuth = async (connector: Connector) => {
    setSelectedConnector(connector);
    setConnecting(true);
    try {
      // Get OAuth URL from backend
      const { oauth_url } = await connectorsApi.getConnectorOAuthUrl(connector.slug);

      // Open our callback page which redirects to OAuth provider
      // The callback page will redirect to the OAuth URL, and the provider
      // will redirect back to our API which handles the callback
      const callbackUrl = `/connectors/callback?slug=${encodeURIComponent(connector.slug)}&oauth_url=${encodeURIComponent(oauth_url)}`;
      const popup = openOAuthPopup(callbackUrl);
      if (!popup) {
        toast({
          title: t('integrationsSettings.popupBlocked'),
          description: t('integrationsSettings.popupBlockedDescription'),
          variant: 'destructive',
        });
        setConnecting(false);
        return;
      }
      oauthPopupRef.current = popup;

      // Poll for popup closure as fallback (postMessage is primary)
      oauthCheckIntervalRef.current = setInterval(() => {
        if (popup.closed) {
          clearInterval(oauthCheckIntervalRef.current!);
          oauthCheckIntervalRef.current = null;
          loadData();
          oauthPopupRef.current = null;
        }
      }, 1000);
    } catch (err: any) {
      console.error('Failed to get OAuth URL:', err);
      // Surface the backend error message if available
      const backendMessage = err?.response?.data?.message || err?.message;
      toast({
        title: t('integrationsSettings.oauthInitFailed'),
        description: backendMessage || t('integrationsSettings.oauthInitFailedDescription'),
        variant: 'destructive',
      });
    } finally {
      setConnecting(false);
    }
  };

  // ─── Toggle enable/disable ───────────────────────────────────────────────────

  const handleToggle = async (connector: Connector, enabled: boolean) => {
    const uc = userConnectors.find((u) => u.connector_slug === connector.slug);
    if (!uc) return;

    setTogglingId(uc.id);
    try {
      await connectorsApi.updateConnector(uc.id, {
        enabled,
        sync_frequency: (uc as any).sync_frequency || '1h',
        auto_sync: (uc as any).auto_sync ?? true,
      });
      setUserConnectors((prev) =>
        prev.map((u) => (u.id === uc.id ? { ...u, status: enabled ? 'active' : 'disabled' } : u))
      );
      toast({
        title: enabled
          ? t('integrationsSettings.enabled', { name: connector.name })
          : t('integrationsSettings.disabled', { name: connector.name }),
      });
    } catch (err) {
      console.error('Failed to toggle connector:', err);
      toast({
        title: t('integrationsSettings.toggleFailed'),
        description: t('integrationsSettings.toggleFailedDescription'),
        variant: 'destructive',
      });
    } finally {
      setTogglingId(null);
    }
  };

  // ─── Save settings ──────────────────────────────────────────────────────────

  const handleSaveSettings = async (settings: ConnectorSettings) => {
    const uc = userConnectors.find((u) => u.connector_slug === selectedConnector?.slug);
    if (!uc) return;

    setSaving(true);
    try {
      const updated = await connectorsApi.updateConnector(uc.id, {
        display_name: settings.displayName,
        sync_frequency: settings.syncFrequency,
        auto_sync: settings.autoSync,
        enabled: uc.status !== 'disabled',
      });
      setUserConnectors((prev) => prev.map((u) => (u.id === uc.id ? updated : u)));
      toast({
        title: t('integrationsSettings.settingsSaved'),
        description: t('integrationsSettings.settingsSavedDescription'),
      });
      setDrawerOpen(false);
    } catch (err) {
      console.error('Failed to save settings:', err);
      toast({
        title: t('integrationsSettings.saveFailed'),
        variant: 'destructive',
      });
    } finally {
      setSaving(false);
    }
  };

  // ─── Sync ───────────────────────────────────────────────────────────────────

  const handleSync = async (uc: UserConnector) => {
    setSyncingId(uc.id);
    try {
      await connectorsApi.triggerSync(uc.id);
      toast({
        title: t('integrationsSettings.syncStarted', {
          name: uc.display_name || uc.connector_name,
        }),
        description: t('integrationsSettings.syncStartedDescription'),
      });
      // Reload to get updated last_sync_at
      await loadData();
    } catch (err) {
      console.error('Failed to trigger sync:', err);
      toast({
        title: t('integrationsSettings.syncFailed', {
          name: uc.display_name || uc.connector_name,
        }),
        variant: 'destructive',
      });
    } finally {
      setSyncingId(null);
    }
  };

  // ─── Disconnect ─────────────────────────────────────────────────────────────

  const handleDisconnectRequest = (uc: UserConnector) => {
    setDisconnectingId(uc.id);
    setDisconnectingName(uc.display_name || uc.connector_name);
    setDisconnectOpen(true);
  };

  const handleDisconnectConfirm = async () => {
    if (!disconnectingId) return;
    setDisconnecting(true);
    try {
      await connectorsApi.unlinkConnector(disconnectingId);
      setUserConnectors((prev) => prev.filter((u) => u.id !== disconnectingId));
      toast({
        title: t('integrationsSettings.disconnected', { name: disconnectingName }),
      });
      setDisconnectOpen(false);
      if (selectedConnector && drawerOpen) {
        setDrawerOpen(false);
      }
    } catch (err) {
      console.error('Failed to unlink connector:', err);
      toast({
        title: t('integrationsSettings.disconnectFailed', { name: disconnectingName }),
        variant: 'destructive',
      });
    } finally {
      setDisconnecting(false);
      setDisconnectingId(null);
      setDisconnectingName('');
    }
  };

  // ─── Configure (open drawer) ────────────────────────────────────────────────

  const handleConfigure = (connector: Connector) => {
    setSelectedConnector(connector);
    setDrawerOpen(true);
  };

  const getUserConnector = (slug: string) => userConnectors.find((u) => u.connector_slug === slug);

  // ─── Render ─────────────────────────────────────────────────────────────────

  if (loading) {
    return (
      <div className="space-y-3">
        <div className="h-16 rounded-xl bg-white/[0.02] border border-white/[0.04] animate-pulse" />
        {[1, 2, 3, 4, 5].map((i) => (
          <div
            key={i}
            className="h-20 rounded-xl bg-white/[0.02] border border-white/[0.04] animate-pulse"
          />
        ))}
      </div>
    );
  }

  return (
    <div className="settings-page space-y-6">
      {/* Security notice */}
      <Card className="settings-panel border-indigo-500/20 bg-indigo-500/5">
        <CardContent className="p-3 flex items-start gap-3">
          <Shield className="w-4 h-4 text-indigo-400 mt-0.5 shrink-0" aria-hidden="true" />
          <div className="text-xs text-text-secondary">
            <span className="text-indigo-300 font-medium">
              {t('integrationsSettings.zeroKnowledgeNotice')}
            </span>{' '}
            {t('integrationsSettings.securityDescription')}
          </div>
        </CardContent>
      </Card>

      {/* Connector rows */}
      <div className="space-y-2">
        <div className="flex items-center justify-between px-1">
          <h2 className="text-sm font-medium text-text-secondary">
            {t('integrationsSettings.brainConnectors')}
          </h2>
          <span className="text-xs text-text-muted">
            {userConnectors.length} / {t('integrationsSettings.connectorsLinked', { max: 10 })}
          </span>
        </div>

        <AnimatePresence mode="popLayout">
          {catalog.map((connector) => (
            <ConnectorRow
              key={connector.id}
              connector={connector}
              userConnector={getUserConnector(connector.slug)}
              onToggle={(enabled) => handleToggle(connector, enabled)}
              onConfigure={() => handleConfigure(connector)}
              onSync={() => {
                const uc = getUserConnector(connector.slug);
                if (uc) handleSync(uc);
              }}
              onDisconnect={() => {
                const uc = getUserConnector(connector.slug);
                if (uc) handleDisconnectRequest(uc);
              }}
              syncing={syncingId === getUserConnector(connector.slug)?.id}
              toggling={togglingId === getUserConnector(connector.slug)?.id}
            />
          ))}
        </AnimatePresence>
      </div>

      {/* Info callout */}
      <div className="flex items-start gap-2 rounded-lg border border-white/[0.06] bg-white/[0.02] p-3">
        <Info className="w-4 h-4 text-text-muted mt-0.5 shrink-0" aria-hidden="true" />
        <p className="text-xs text-text-muted">{t('integrationsSettings.brainInfoCallout')}</p>
      </div>

      {/* Config drawer */}
      {selectedConnector && (
        <ConfigDrawer
          connector={selectedConnector}
          userConnector={getUserConnector(selectedConnector.slug)}
          open={drawerOpen}
          onClose={() => setDrawerOpen(false)}
          onSave={handleSaveSettings}
          onConnect={() => handleConnectOAuth(selectedConnector)}
          onDisconnect={() => {
            const uc = getUserConnector(selectedConnector.slug);
            if (uc) handleDisconnectRequest(uc);
          }}
          saving={saving}
          connecting={connecting}
        />
      )}

      {/* Disconnect confirm */}
      <DisconnectDialog
        open={disconnectOpen}
        connectorName={disconnectingName}
        onConfirm={handleDisconnectConfirm}
        onCancel={() => {
          setDisconnectOpen(false);
          setDisconnectingId(null);
          setDisconnectingName('');
        }}
        disconnecting={disconnecting}
      />
    </div>
  );
}
