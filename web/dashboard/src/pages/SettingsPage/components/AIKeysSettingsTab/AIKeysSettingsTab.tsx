import { useState } from 'react';
import { Trans, useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { useAIKeys, useSupportedProviders, useConnectAIKey, useDisconnectAIKey, useTestAIKey, useRotateAIKey } from '@/api/ai-keys';
import type { AIProviderKey, SupportedProvider } from '@/types/ai-keys';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger } from '@/components/ui/alert-dialog';
import { ProviderIcon } from '@/components/common/ProviderIcon';
import { Brain, Eye, EyeOff, Key, Loader2, Plus, RefreshCw, Shield, TestTube, Trash2 } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';

const TOKEN_PLAN_REGIONS = [
  { id: 'cn', label: 'China', flag: 'CN' },
  { id: 'sgp', label: 'Singapore (US/APAC)', flag: 'SG' },
  { id: 'eu', label: 'Europe', flag: 'EU' },
];

function StatusBadge({ status }: { status: string }) {
  const styles: Record<string, { bg: string; color: string; border: string }> = {
    active: { bg: 'rgba(143,255,208,0.06)', color: 'var(--status-ok)', border: 'rgba(143,255,208,0.3)' },
    degraded: { bg: 'rgba(232,196,104,0.06)', color: 'var(--status-pending)', border: 'rgba(232,196,104,0.3)' },
    expired: { bg: 'rgba(255,107,107,0.06)', color: 'var(--status-revoked)', border: 'rgba(255,107,107,0.3)' },
    revoked: { bg: 'rgba(255,107,107,0.06)', color: 'var(--status-revoked)', border: 'rgba(255,107,107,0.3)' },
  };
  const s = styles[status] ?? styles.active;
  return (
    <span
      className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-[var(--radius-sm)] text-xs font-medium capitalize"
      style={{ background: s.bg, color: s.color, border: `1px solid ${s.border}`, fontFamily: 'var(--font-mono)', letterSpacing: '0.04em' }}
    >
      <span className="w-1.5 h-1.5 rounded-full" style={{ background: s.color }} />
      {status}
    </span>
  );
}

function ConnectedKeyCard({
  keyEntry,
  onTest,
  onRotate,
  onDisconnect,
  isTesting,
}: {
  keyEntry: AIProviderKey;
  onTest: () => void;
  onRotate: (newKey: string) => void;
  onDisconnect: () => void;
  isTesting: boolean;
}) {
  const { t } = useTranslation();
  const [showRotate, setShowRotate] = useState(false);
  const [newKey, setNewKey] = useState('');

  const isTokenPlan = keyEntry.provider === 'mimo-token-plan' || keyEntry.provider === 'minimax-token-plan';
  const keyPrefix = keyEntry.provider === 'mimo-token-plan' ? 'tp-' : 'sk-';
  const regionLabel = TOKEN_PLAN_REGIONS.find((r) => r.id === keyEntry.token_plan_region);
  const displayName: Record<string, string> = {
    'mimo-token-plan': 'MiMo Token Plan',
    'minimax-token-plan': 'MiniMax Token Plan',
  };

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span className="w-10 h-10 rounded-lg flex items-center justify-center bg-muted">
              <ProviderIcon provider={keyEntry.provider} size="xl" />
            </span>
            <div>
              <CardTitle className="text-base capitalize flex items-center gap-2">
                {displayName[keyEntry.provider] ?? keyEntry.provider}
                {regionLabel && (
                  <Badge variant="outline" className="text-xs font-normal">
                    {regionLabel.flag} {regionLabel.label}
                  </Badge>
                )}
              </CardTitle>
              <p className="text-sm text-muted-foreground font-mono">
                {keyPrefix}...{keyEntry.key_last4}
              </p>
            </div>
          </div>
          <StatusBadge status={keyEntry.status} />
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        {keyEntry.health_message && keyEntry.status !== 'active' && (
          <p className="text-sm" style={{ color: 'var(--status-pending)' }}>{keyEntry.health_message}</p>
        )}
        <div className="flex flex-wrap gap-2 text-xs" style={{ color: 'var(--text-faint)' }}>
          {keyEntry.last_health_check && (
            <span>Checked {formatDistanceToNow(new Date(keyEntry.last_health_check), { addSuffix: true })}</span>
          )}
          {keyEntry.last_used_at && (
            <span>Used {formatDistanceToNow(new Date(keyEntry.last_used_at), { addSuffix: true })}</span>
          )}
        </div>

        {showRotate && (
          <div className="flex gap-2">
            <Input
              type="password"
              placeholder={t('aiKeysSettings.newApiKey')}
              value={newKey}
              onChange={(e) => setNewKey(e.target.value)}
              className="font-mono text-sm"
            />
            <Button
              size="sm"
              onClick={() => {
                onRotate(newKey);
                setNewKey('');
                setShowRotate(false);
              }}
              disabled={!newKey}
            >
              {t('aiKeysSettings.save')}
            </Button>
            <Button size="sm" variant="ghost" onClick={() => setShowRotate(false)}>
              {t('aiKeysSettings.cancel')}
            </Button>
          </div>
        )}

        <div className="flex gap-2">
          <Button size="sm" variant="outline" onClick={onTest} disabled={isTesting}>
            <TestTube className="h-3.5 w-3.5 mr-1" />
            {isTesting ? t('aiKeysSettings.testing') : t('aiKeysSettings.test')}
          </Button>
          <Button size="sm" variant="outline" onClick={() => setShowRotate(!showRotate)}>
            <RefreshCw className="h-3.5 w-3.5 mr-1" />
            {t('aiKeysSettings.rotate')}
          </Button>
          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button size="sm" variant="destructive">
                <Trash2 className="h-3.5 w-3.5 mr-1" />
                {t('aiKeysSettings.disconnect')}
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent style={{ background: 'var(--panel)', borderColor: 'var(--panel-edge)', borderRadius: 'var(--radius-lg)' }}>
              <AlertDialogHeader>
                <AlertDialogTitle style={{ fontFamily: 'var(--font-display)' }}>{t('aiKeysSettings.disconnectTitle', { provider: keyEntry.provider })}</AlertDialogTitle>
                <AlertDialogDescription style={{ color: 'var(--text-dim)' }}>
                  {t('aiKeysSettings.disconnectDesc', { provider: keyEntry.provider })}
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>{t('aiKeysSettings.cancel')}</AlertDialogCancel>
                <AlertDialogAction onClick={onDisconnect}>{t('aiKeysSettings.disconnect')}</AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </div>
      </CardContent>
    </Card>
  );
}

function ConnectKeyDialog({
  providers,
  connectedProviders,
  onConnect,
  isConnecting,
}: {
  providers: SupportedProvider[];
  connectedProviders: Set<string>;
  onConnect: (provider: string, apiKey: string, region?: string) => void;
  isConnecting: boolean;
}) {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  const [selected, setSelected] = useState('');
  const [apiKey, setApiKey] = useState('');
  const [showKey, setShowKey] = useState(false);
  const [region, setRegion] = useState('');

  const available = providers.filter((p) => !connectedProviders.has(p.id));
  const selectedProvider = providers.find((p) => p.id === selected);

  const handleSubmit = () => {
    if (selected && apiKey) {
      if (selected === 'mimo-token-plan' && !region) return;
      onConnect(selected, apiKey, region || undefined);
      setApiKey('');
      setSelected('');
      setRegion('');
      setOpen(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
          <Plus className="h-4 w-4 mr-2" />
          {t('aiKeysSettings.connectKey')}
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-lg" style={{ background: 'var(--panel)', borderColor: 'var(--panel-edge)', borderRadius: 'var(--radius-lg)', boxShadow: 'var(--shadow-chamber)' }}>
        <DialogHeader>
          <DialogTitle style={{ fontFamily: 'var(--font-display)', fontSize: '18px' }}>{t('aiKeysSettings.connectDialogTitle')}</DialogTitle>
        </DialogHeader>
        <div className="space-y-5 pt-1">
          {/* Provider grid */}
          <div>
            <Label style={{ color: 'var(--text)', fontFamily: 'var(--font-mono)', fontSize: '11px', fontWeight: 600, letterSpacing: '0.06em', textTransform: 'uppercase' }}>
              {t('aiKeysSettings.provider')}
            </Label>
            <div className="grid grid-cols-3 gap-2 mt-2">
              {available.map((p) => {
                const isActive = selected === p.id;
                return (
                  <button
                    key={p.id}
                    onClick={() => setSelected(p.id)}
                    className="flex flex-col items-center gap-2 p-3 rounded-[var(--radius)] transition-all"
                    style={{
                      background: isActive ? 'rgba(143,255,208,0.06)' : 'var(--panel-raised)',
                      border: isActive ? '1px solid rgba(143,255,208,0.3)' : '1px solid var(--panel-edge)',
                      boxShadow: isActive ? '0 0 0 1px rgba(143,255,208,0.1)' : 'none',
                    }}
                  >
                    <span
                      className="w-10 h-10 rounded-[var(--radius)] flex items-center justify-center"
                      style={{
                        background: isActive ? 'rgba(143,255,208,0.08)' : 'var(--panel)',
                        border: '1px solid var(--panel-edge)',
                      }}
                    >
                      <ProviderIcon provider={p.id} size="xl" />
                    </span>
                    <p className="text-xs font-medium capitalize" style={{ color: isActive ? 'var(--status-ok)' : 'var(--text-dim)' }}>
                      {p.name}
                    </p>
                  </button>
                );
              })}
            </div>
          </div>

          {selectedProvider && (
            <>
              {selectedProvider.is_meta_provider && (
                <div className="rounded-lg bg-blue-50 dark:bg-blue-950/30 p-3 text-sm text-blue-700 dark:text-blue-300">
                  <Trans i18nKey="aiKeysSettings.metaProviderDesc" components={{ 1: <strong /> }} values={{ name: selectedProvider.name }} />
                </div>
              )}
              {selected === 'mimo-token-plan' && (
                <div>
                  <Label>{t('aiKeysSettings.region', { defaultValue: 'Region' })}</Label>
                  <div className="grid grid-cols-3 gap-2 mt-2">
                    {TOKEN_PLAN_REGIONS.map((r) => (
                      <button
                        key={r.id}
                        onClick={() => setRegion(r.id)}
                        className={`p-2 rounded-lg border text-center text-sm transition-colors ${
                          region === r.id
                            ? 'border-primary bg-primary/10'
                            : 'border-border hover:border-primary/50'
                        }`}
                      >
                        <span className="text-xs font-bold">{r.flag}</span>
                        <p className="text-xs mt-1">{r.label}</p>
                      </button>
                    ))}
                  </div>
                  <p className="text-xs text-muted-foreground mt-1.5">
                    {t('aiKeysSettings.regionHint', { defaultValue: 'Select the cluster matching your Token Plan subscription.' })}
                  </p>
                </div>
              )}

              {/* API Key input */}
              <div className="space-y-2">
                <Label htmlFor="api-key" style={{ color: 'var(--text)', fontFamily: 'var(--font-mono)', fontSize: '11px', fontWeight: 600, letterSpacing: '0.06em', textTransform: 'uppercase' }}>
                  {t('aiKeysSettings.apiKey')}
                </Label>
                <div className="relative">
                  <Input
                    id="api-key"
                    type={showKey ? 'text' : 'password'}
                    placeholder={selectedProvider.key_format}
                    value={apiKey}
                    onChange={(e) => setApiKey(e.target.value)}
                    className="font-mono text-sm pr-10"
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' && apiKey && !isConnecting) {
                        e.preventDefault();
                        handleSubmit();
                      }
                    }}
                  />
                  <button
                    type="button"
                    onClick={() => setShowKey(!showKey)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 transition-colors"
                    style={{ color: 'var(--text-faint)' }}
                  >
                    {showKey ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </button>
                </div>
                <p className="text-xs" style={{ color: 'var(--text-faint)' }}>
                  Format: <code className="font-mono" style={{ color: 'var(--text-dim)' }}>{selectedProvider.key_format}</code>
                </p>
              </div>

              {/* Security note */}
              <div className="flex items-start gap-2.5 p-3 rounded-[var(--radius)]" style={{ background: 'rgba(143,255,208,0.03)', border: '1px solid rgba(143,255,208,0.1)' }}>
                <Shield className="h-4 w-4 mt-0.5 shrink-0" style={{ color: 'var(--status-ok)' }} />
                <p className="text-xs leading-relaxed" style={{ color: 'var(--text-dim)' }}>
                  {t('aiKeysSettings.keyEncrypted')}
                </p>
              </div>

              {/* Submit */}
              <Button onClick={handleSubmit} disabled={!apiKey || isConnecting || (selected === 'mimo-token-plan' && !region)} className="w-full gap-2">
                {isConnecting ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin" />
                    {t('aiKeysSettings.validating')}
                  </>
                ) : (
                  <>
                    <Key className="h-4 w-4" />
                    {t('aiKeysSettings.testAndConnect')}
                  </>
                )}
              </Button>
            </>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}

export function AIKeysSettingsTab() {
  const { t } = useTranslation();
  const { data: keys = [], isLoading: keysLoading } = useAIKeys();
  const { data: providers = [] } = useSupportedProviders();
  const connectMutation = useConnectAIKey();
  const disconnectMutation = useDisconnectAIKey();
  const testMutation = useTestAIKey();
  const rotateMutation = useRotateAIKey();
  const [testingProvider, setTestingProvider] = useState<string | null>(null);

  const connectedProviders = new Set(keys.map((k) => k.provider));

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold" style={{ fontFamily: 'var(--font-display)' }}>{t('aiKeysSettings.title')}</h3>
          <p className="text-sm" style={{ color: 'var(--text-dim)' }}>
            {t('aiKeysSettings.description')}
          </p>
        </div>
        <ConnectKeyDialog
          providers={providers}
          connectedProviders={connectedProviders}
          onConnect={(provider, apiKey, region) => connectMutation.mutate({ provider, apiKey, region })}
          isConnecting={connectMutation.isPending}
        />
      </div>

      <div className="h-px" style={{ background: 'var(--panel-edge)' }} />

      {keysLoading ? (
        <div className="text-center py-8" style={{ color: 'var(--text-faint)' }}>{t('aiKeysSettings.loadingKeys')}</div>
      ) : keys.length === 0 ? (
        <div className="text-center py-12 px-6" style={{ background: 'var(--panel-raised)', borderRadius: 'var(--radius-lg)', border: '1px solid var(--panel-edge)' }}>
          <div
            style={{
              width: 64,
              height: 64,
              borderRadius: 'var(--radius-lg)',
              background: 'rgba(139, 92, 246, 0.1)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              margin: '0 auto 16px',
            }}
          >
            <Brain className="h-8 w-8" style={{ color: 'rgba(139, 92, 246, 0.8)' }} />
          </div>
          <h3 className="text-lg font-semibold mb-2" style={{ fontFamily: 'var(--font-display)', color: 'var(--text)' }}>
            Bring Your Own AI Keys
          </h3>
          <p className="text-sm max-w-md mx-auto mb-6" style={{ color: 'var(--text-dim)' }}>
            Connect your first AI provider to enable AI-powered agents and workflows.
            Your keys are encrypted and never stored in plaintext.
          </p>
          <div className="flex flex-wrap justify-center gap-3 mb-6">
            {[
              { name: 'OpenAI', desc: 'GPT-4o, o1', badge: 'Most Popular' },
              { name: 'Anthropic', desc: 'Claude 3.5', badge: 'Best Value' },
              { name: 'OpenRouter', desc: '100+ models', badge: 'Most Models' },
            ].map((p) => (
              <div
                key={p.name}
                className="px-4 py-2 rounded-lg text-left"
                style={{
                  background: 'var(--panel)',
                  border: '1px solid var(--panel-edge)',
                  minWidth: 140,
                }}
              >
                <p className="text-sm font-medium capitalize" style={{ color: 'var(--text)' }}>{p.name}</p>
                <p className="text-xs" style={{ color: 'var(--text-faint)' }}>{p.desc}</p>
              </div>
            ))}
          </div>
          <ConnectKeyDialog
            providers={providers}
            connectedProviders={connectedProviders}
            onConnect={(provider, apiKey, region) => connectMutation.mutate({ provider, apiKey, region })}
            isConnecting={connectMutation.isPending}
          />
        </div>
      ) : (
        <div className="grid gap-4">
          {keys.map((key) => (
            <ConnectedKeyCard
              key={key.id}
              keyEntry={key}
              onTest={() => {
                setTestingProvider(key.provider);
                testMutation.mutate(key.provider, {
                  onSuccess: () => {
                    toast.success(t('aiKeysSettings.testSuccess', { provider: key.provider, defaultValue: `${key.provider} key is working` }));
                  },
                  onError: (err: Error) => {
                    toast.error(err.message || t('aiKeysSettings.testFailed', { defaultValue: 'Key test failed' }));
                  },
                  onSettled: () => setTestingProvider(null),
                });
              }}
              onRotate={(newKey) => rotateMutation.mutate({ provider: key.provider, apiKey: newKey })}
              onDisconnect={() => disconnectMutation.mutate(key.provider)}
              isTesting={testingProvider === key.provider}
            />
          ))}
        </div>
      )}
    </div>
  );
}
