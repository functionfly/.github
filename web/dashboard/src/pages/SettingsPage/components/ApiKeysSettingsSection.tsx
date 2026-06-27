import { apiKeysApi } from '@/api/apikeys';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import type { APIKey, CreateAPIKeyRequest } from '@/types/api-key';
import { useQuery } from '@tanstack/react-query';
import { AlertTriangle, ExternalLink, Key, RefreshCw, Trash2 } from 'lucide-react';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { toast } from 'sonner';
import { formatDate, getApiErrorMessage } from '../settings-utils';

export function ApiKeysSettingsSection() {
  const { t } = useTranslation();
  const [generatingKey, setGeneratingKey] = useState(false);
  const [createKeyModalOpen, setCreateKeyModalOpen] = useState(false);
  const [createKeyName, setCreateKeyName] = useState('');
  const [createKeyDescription, setCreateKeyDescription] = useState('');
  const [createKeyExpiresAt, setCreateKeyExpiresAt] = useState('');
  const [createKeyStep, setCreateKeyStep] = useState<'form' | 'key'>('form');
  const [createKeyPlaintext, setCreateKeyPlaintext] = useState<string | null>(null);

  const [rotateKeyId, setRotateKeyId] = useState<string | null>(null);
  const [rotateKeyName, setRotateKeyName] = useState('');
  const [rotateNewPlaintext, setRotateNewPlaintext] = useState<string | null>(null);
  const [rotating, setRotating] = useState(false);

  const [revokeKeyId, setRevokeKeyId] = useState<string | null>(null);
  const [revokeKeyName, setRevokeKeyName] = useState('');
  const [revoking, setRevoking] = useState(false);

  const {
    data: apiKeysData,
    isLoading: apiKeysLoading,
    refetch: refetchApiKeys,
  } = useQuery({
    queryKey: ['api-keys'],
    queryFn: async () => {
      try {
        const res = await apiKeysApi.list();
        return res.data as { data?: APIKey[]; meta?: { total?: number } };
      } catch (e: unknown) {
        const status = (e as { response?: { status?: number } })?.response?.status;
        if (status === 404) return { data: [] };
        throw e;
      }
    },
    enabled: true,
    retry: false,
  });

  const apiKeys: APIKey[] = Array.isArray((apiKeysData as { data?: APIKey[] })?.data)
    ? (apiKeysData as { data: APIKey[] }).data
    : [];

  const handleOpenCreateKeyModal = () => {
    setCreateKeyStep('form');
    setCreateKeyPlaintext(null);
    setCreateKeyName('');
    setCreateKeyDescription('');
    setCreateKeyExpiresAt('');
    setCreateKeyModalOpen(true);
  };

  const handleCreateKeySubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!createKeyName.trim()) {
      toast.error(t('apiKeysSettings.toastNameRequired'));
      return;
    }
    setGeneratingKey(true);
    try {
      const body: CreateAPIKeyRequest = {
        name: createKeyName.trim(),
        key_type: 'platform',
      };
      if (createKeyDescription.trim()) body.description = createKeyDescription.trim();
      if (createKeyExpiresAt) body.expires_at = new Date(createKeyExpiresAt).toISOString();
      const res = await apiKeysApi.create(body);
      const plaintext = (res.data as { data?: { plaintext?: string } })?.data?.plaintext ?? null;
      setCreateKeyPlaintext(plaintext);
      setCreateKeyStep('key');
      refetchApiKeys();
      toast.success(t('apiKeysSettings.toastCreated'));
    } catch (err: unknown) {
      const msg = getApiErrorMessage(err, {
        401: t('apiKeysSettings.errorSignIn'),
        403: t('apiKeysSettings.errorNoPermission'),
        404: t('apiKeysSettings.errorRouteNotFound'),
        default: t('apiKeysSettings.errorFailedToGenerate'),
      });
      toast.error(msg);
    } finally {
      setGeneratingKey(false);
    }
  };

  const handleCreateKeyDone = () => {
    setCreateKeyModalOpen(false);
    setCreateKeyStep('form');
    setCreateKeyPlaintext(null);
    setCreateKeyName('');
    setCreateKeyDescription('');
    setCreateKeyExpiresAt('');
  };

  const handleCopyKey = async (value: string) => {
    try {
      await navigator.clipboard.writeText(value);
      toast.success(t('apiKeysSettings.toastCopied'));
    } catch {
      toast.error(t('apiKeysSettings.toastCopyFailed', 'Failed to copy'));
    }
  };

  const handleRotateClick = (key: APIKey) => {
    setRotateKeyId(key.id);
    setRotateKeyName(key.name);
    setRotateNewPlaintext(null);
  };

  const handleRotateConfirm = async () => {
    if (!rotateKeyId) return;
    setRotating(true);
    try {
      const res = await apiKeysApi.rotate(rotateKeyId, { reason: 'manual' });
      const plaintext = (res.data as { data?: { plaintext?: string } })?.data?.plaintext ?? null;
      setRotateNewPlaintext(plaintext);
      refetchApiKeys();
      toast.success(t('apiKeysSettings.toastRotated'));
    } catch (err: unknown) {
      const msg = getApiErrorMessage(err, { default: t('apiKeysSettings.errorFailedToRotate') });
      toast.error(msg);
    } finally {
      setRotating(false);
    }
  };

  const handleRotateClose = () => {
    setRotateKeyId(null);
    setRotateKeyName('');
    setRotateNewPlaintext(null);
  };

  const handleRevokeClick = (key: APIKey) => {
    setRevokeKeyId(key.id);
    setRevokeKeyName(key.name);
  };

  const handleRevokeConfirm = async () => {
    if (!revokeKeyId) return;
    setRevoking(true);
    try {
      await apiKeysApi.delete(revokeKeyId);
      refetchApiKeys();
      toast.success(t('apiKeysSettings.toastRevoked'));
      setRevokeKeyId(null);
      setRevokeKeyName('');
    } catch (err: unknown) {
      const msg = getApiErrorMessage(err, { default: t('apiKeysSettings.errorFailedToRevoke') });
      toast.error(msg);
    } finally {
      setRevoking(false);
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
            {t('apiKeysSettings.title')}
          </h3>
          <p className="text-sm mt-1" style={{ color: 'var(--text-dim)' }}>
            {t('apiKeysSettings.description')}
          </p>
        </div>
        <div className="space-y-4">
          {apiKeysLoading ? (
            <p className="text-sm" style={{ color: 'var(--text-dim)' }}>
              {t('apiKeysSettings.loadingApiKeys')}
            </p>
          ) : apiKeys.length === 0 ? (
            <div
              className="rounded-lg p-6 text-center"
              style={{ background: 'var(--panel-raised)', border: '1px solid var(--panel-edge)' }}
            >
              <p className="text-sm" style={{ color: 'var(--text-dim)' }}>
                {t('apiKeysSettings.noApiKeysYet')}
              </p>
              <p className="mt-1 text-sm" style={{ color: 'var(--text-dim)' }}>
                {t('apiKeysSettings.generateFirst')}
              </p>
              <Button
                className="mt-4 gap-2"
                onClick={handleOpenCreateKeyModal}
                disabled={generatingKey}
                style={{
                  background: 'linear-gradient(180deg, #ffffff, #d8dee2)',
                  color: 'var(--text-on-light)',
                  boxShadow: 'var(--shadow-btn-primary-rest)',
                }}
              >
                <Key className="h-4 w-4" />
                {generatingKey
                  ? t('apiKeysSettings.generating')
                  : t('apiKeysSettings.generateFirstButton')}
              </Button>
            </div>
          ) : (
            <>
              {apiKeys.map((key: APIKey) => {
                const prefix = key.prefix ?? key.key_prefix ?? 'ffp_';
                const lastUsed = key.last_used_at
                  ? formatDate(key.last_used_at)
                  : t('apiKeysSettings.never');
                const expiresAt = key.expires_at ? formatDate(key.expires_at) : null;
                return (
                  <div
                    key={key.id}
                    className="flex flex-wrap items-center justify-between gap-3 rounded-lg p-4"
                    style={{
                      background: 'var(--panel-raised)',
                      border: '1px solid var(--panel-edge)',
                    }}
                  >
                    <div className="min-w-0">
                      <h4 className="font-medium" style={{ color: 'var(--text)' }}>
                        {key.name || t('apiKeysSettings.defaultKeyName')}
                      </h4>
                      <p className="text-sm" style={{ color: 'var(--text-dim)' }}>
                        {t('apiKeysSettings.createdLabel', {
                          created: key.created_at ? formatDate(key.created_at) : '—',
                          lastUsed,
                        })}
                        {expiresAt &&
                          ` · ${t('apiKeysSettings.expiresLabel', { expires: expiresAt })}`}
                      </p>
                    </div>
                    <div className="flex flex-wrap items-center gap-2">
                      <span
                        className="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium"
                        style={{
                          background: key.is_active ? 'rgba(143, 255, 208, 0.06)' : 'var(--panel)',
                          color: key.is_active ? 'var(--status-ok)' : 'var(--text-dim)',
                        }}
                      >
                        {key.is_active
                          ? t('apiKeysSettings.active')
                          : t('apiKeysSettings.inactive')}
                      </span>
                      <code
                        className="rounded px-3 py-1 text-sm"
                        style={{ background: 'var(--panel)', color: 'var(--text-dim)' }}
                      >
                        {prefix}••••••••••••
                      </code>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleCopyKey(prefix)}
                        title={t('apiKeysSettings.copyPrefix')}
                        style={{ color: 'var(--text-dim)' }}
                      >
                        {t('apiKeysSettings.copyPrefix')}
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleRotateClick(key)}
                        title={t('apiKeysSettings.rotateKey')}
                        disabled={rotating}
                        style={{ color: 'var(--text-dim)' }}
                      >
                        <RefreshCw className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleRevokeClick(key)}
                        title={t('apiKeysSettings.revokeKey')}
                        style={{ color: 'var(--status-revoked)' }}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                );
              })}
              <Button
                className="gap-2"
                onClick={handleOpenCreateKeyModal}
                disabled={generatingKey}
                style={{
                  background: 'linear-gradient(180deg, #ffffff, #d8dee2)',
                  color: 'var(--text-on-light)',
                  boxShadow: 'var(--shadow-btn-primary-rest)',
                }}
              >
                <Key className="h-4 w-4" />
                {generatingKey
                  ? t('apiKeysSettings.generating')
                  : t('apiKeysSettings.generateNewKey')}
              </Button>
            </>
          )}
          <p className="text-xs" style={{ color: 'var(--text-faint)' }}>
            {t('apiKeysSettings.forPermissions')}
            <Link
              to="/dashboard/api-keys"
              style={{ color: 'var(--accent)', textDecoration: 'underline' }}
            >
              {t('apiKeysSettings.apiKeysInDashboard')}
            </Link>
            <ExternalLink className="ml-0.5 inline h-3 w-3" />
          </p>
        </div>
      </div>

      <Dialog open={createKeyModalOpen} onOpenChange={(open) => !open && handleCreateKeyDone()}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>
              {createKeyStep === 'form'
                ? t('apiKeysSettings.createDialogTitle')
                : t('apiKeysSettings.copyDialogTitle')}
            </DialogTitle>
            <DialogDescription>
              {createKeyStep === 'form'
                ? t('apiKeysSettings.createDialogDescription')
                : t('apiKeysSettings.copyDialogDescription')}
            </DialogDescription>
          </DialogHeader>
          {createKeyStep === 'form' ? (
            <form onSubmit={handleCreateKeySubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="api-key-name">{t('apiKeysSettings.nameLabel')}</Label>
                <Input
                  id="api-key-name"
                  value={createKeyName}
                  onChange={(e) => setCreateKeyName(e.target.value)}
                  placeholder={t('apiKeysSettings.namePlaceholder')}
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="api-key-desc">{t('apiKeysSettings.descriptionLabel')}</Label>
                <Textarea
                  id="api-key-desc"
                  value={createKeyDescription}
                  onChange={(e) => setCreateKeyDescription(e.target.value)}
                  placeholder={t('apiKeysSettings.descriptionPlaceholder')}
                  rows={2}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="api-key-expires">{t('apiKeysSettings.expiresLabel')}</Label>
                <Input
                  id="api-key-expires"
                  type="date"
                  value={createKeyExpiresAt}
                  onChange={(e) => setCreateKeyExpiresAt(e.target.value)}
                />
              </div>
              <DialogFooter>
                <Button
                  type="button"
                  onClick={handleCreateKeyDone}
                  style={{ borderColor: 'var(--steel)', color: 'var(--text)' }}
                >
                  {t('apiKeysSettings.cancel')}
                </Button>
                <Button
                  type="submit"
                  disabled={generatingKey}
                  style={{
                    background: 'linear-gradient(180deg, #ffffff, #d8dee2)',
                    color: 'var(--text-on-light)',
                    boxShadow: 'var(--shadow-btn-primary-rest)',
                  }}
                >
                  {generatingKey ? t('apiKeysSettings.creating') : t('apiKeysSettings.createKey')}
                </Button>
              </DialogFooter>
            </form>
          ) : (
            <div className="space-y-4">
              <div
                className="flex items-center gap-2 rounded-lg p-3 text-sm"
                style={{
                  background: 'rgba(232, 196, 104, 0.06)',
                  border: '1px solid rgba(232, 196, 104, 0.3)',
                  color: 'var(--status-pending)',
                }}
              >
                <AlertTriangle className="h-5 w-5 shrink-0" />
                <span>{t('apiKeysSettings.copyKeyNowWarning')}</span>
              </div>
              {createKeyPlaintext && (
                <div className="flex gap-2">
                  <Input readOnly value={createKeyPlaintext} className="font-mono text-sm" />
                  <Button
                    type="button"
                    onClick={() => handleCopyKey(createKeyPlaintext)}
                    style={{ borderColor: 'var(--steel)', color: 'var(--text)' }}
                  >
                    {t('apiKeysSettings.copy')}
                  </Button>
                </div>
              )}
              <DialogFooter>
                <Button
                  onClick={handleCreateKeyDone}
                  style={{ borderColor: 'var(--steel)', color: 'var(--text)' }}
                >
                  {t('apiKeysSettings.done')}
                </Button>
              </DialogFooter>
            </div>
          )}
        </DialogContent>
      </Dialog>

      <Dialog open={!!rotateKeyId} onOpenChange={(open) => !open && handleRotateClose()}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{t('apiKeysSettings.rotateDialogTitle')}</DialogTitle>
            <DialogDescription>
              {rotateNewPlaintext
                ? t('apiKeysSettings.rotateNewKeyDescription')
                : t('apiKeysSettings.rotateConfirmDescription', { name: rotateKeyName })}
            </DialogDescription>
          </DialogHeader>
          {rotateNewPlaintext ? (
            <div className="space-y-4">
              <div
                className="flex items-center gap-2 rounded-lg p-3 text-sm"
                style={{
                  background: 'rgba(232, 196, 104, 0.06)',
                  border: '1px solid rgba(232, 196, 104, 0.3)',
                  color: 'var(--status-pending)',
                }}
              >
                <AlertTriangle className="h-5 w-5 shrink-0" />
                <span>{t('apiKeysSettings.copyNewKeyWarning')}</span>
              </div>
              <div className="flex gap-2">
                <Input readOnly value={rotateNewPlaintext} className="font-mono text-sm" />
                <Button
                  type="button"
                  onClick={() => handleCopyKey(rotateNewPlaintext)}
                  style={{ borderColor: 'var(--steel)', color: 'var(--text)' }}
                >
                  {t('apiKeysSettings.copy')}
                </Button>
              </div>
              <DialogFooter>
                <Button
                  onClick={handleRotateClose}
                  style={{ borderColor: 'var(--steel)', color: 'var(--text)' }}
                >
                  {t('apiKeysSettings.done')}
                </Button>
              </DialogFooter>
            </div>
          ) : (
            <DialogFooter>
              <Button
                variant="outline"
                onClick={handleRotateClose}
                style={{ borderColor: 'var(--steel)', color: 'var(--text)' }}
              >
                {t('apiKeysSettings.cancel')}
              </Button>
              <Button
                onClick={handleRotateConfirm}
                disabled={rotating}
                style={{
                  background: 'linear-gradient(180deg, #ffffff, #d8dee2)',
                  color: 'var(--text-on-light)',
                  boxShadow: 'var(--shadow-btn-primary-rest)',
                }}
              >
                {rotating ? t('apiKeysSettings.rotating') : t('apiKeysSettings.rotateKey')}
              </Button>
            </DialogFooter>
          )}
        </DialogContent>
      </Dialog>

      <Dialog
        open={!!revokeKeyId}
        onOpenChange={(open) => !open && (setRevokeKeyId(null), setRevokeKeyName(''))}
      >
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{t('apiKeysSettings.revokeDialogTitle')}</DialogTitle>
            <DialogDescription>
              {t('apiKeysSettings.revokeDialogDescription', { name: revokeKeyName })}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => {
                setRevokeKeyId(null);
                setRevokeKeyName('');
              }}
              style={{ borderColor: 'var(--steel)', color: 'var(--text)' }}
            >
              {t('apiKeysSettings.cancel')}
            </Button>
            <Button variant="destructive" onClick={handleRevokeConfirm} disabled={revoking}>
              {revoking ? t('apiKeysSettings.revoking') : t('apiKeysSettings.revokeKey')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
