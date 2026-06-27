import { deployKeysApi, type DeployKey } from '@/api/deploy-keys';
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
import { useQuery } from '@tanstack/react-query';
import { CheckCircle2, Key, RefreshCw, Trash2 } from 'lucide-react';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { formatDate, getApiErrorMessage } from '../settings-utils';

export function DeployKeysSettingsSection() {
  const { t } = useTranslation();
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [createName, setCreateName] = useState('');
  const [createPublicKey, setCreatePublicKey] = useState('');
  const [creating, setCreating] = useState(false);

  const [verifyKeyId, setVerifyKeyId] = useState<string | null>(null);
  const [verifying, setVerifying] = useState(false);
  const [verificationSuccess, setVerificationSuccess] = useState(false);

  const [deleteKeyId, setDeleteKeyId] = useState<string | null>(null);
  const [deleteKeyName, setDeleteKeyName] = useState('');
  const [deleting, setDeleting] = useState(false);

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['deploy-keys'],
    queryFn: async () => {
      try {
        return await deployKeysApi.list();
      } catch (e: unknown) {
        const status = (e as { response?: { status?: number } })?.response?.status;
        if (status === 404) return { deploy_keys: [], total_count: 0, page: 1, page_size: 20 };
        throw e;
      }
    },
    enabled: true,
    retry: false,
  });

  const deployKeys: DeployKey[] = data?.deploy_keys ?? [];

  const handleOpenCreateModal = () => {
    setCreateName('');
    setCreatePublicKey('');
    setCreateModalOpen(true);
  };

  const handleCreateSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!createName.trim()) {
      toast.error(t('deployKeysSettings.toastNameRequired'));
      return;
    }
    if (!createPublicKey.trim()) {
      toast.error(t('deployKeysSettings.toastPublicKeyRequired'));
      return;
    }
    setCreating(true);
    try {
      await deployKeysApi.create({ name: createName.trim(), public_key: createPublicKey.trim() });
      refetch();
      setCreateModalOpen(false);
      toast.success(t('deployKeysSettings.toastCreated'));
    } catch (err: unknown) {
      const msg = getApiErrorMessage(err, { default: t('deployKeysSettings.errorFailedToCreate') });
      toast.error(msg);
    } finally {
      setCreating(false);
    }
  };

  const handleVerifyClick = (key: DeployKey) => {
    setVerifyKeyId(key.id);
    setVerifying(true);
    setVerificationSuccess(false);
    deployKeysApi
      .verify(key.id)
      .then(() => {
        setVerificationSuccess(true);
        toast.success(t('deployKeysSettings.verificationSuccess'));
      })
      .catch(() => {
        toast.error(t('deployKeysSettings.verificationFailed'));
      })
      .finally(() => {
        setVerifying(false);
      });
  };

  const handleVerifyClose = () => {
    setVerifyKeyId(null);
    setVerificationSuccess(false);
  };

  const handleDeleteClick = (key: DeployKey) => {
    setDeleteKeyId(key.id);
    setDeleteKeyName(key.name);
  };

  const handleDeleteConfirm = async () => {
    if (!deleteKeyId) return;
    setDeleting(true);
    try {
      await deployKeysApi.delete(deleteKeyId);
      refetch();
      setDeleteKeyId(null);
      setDeleteKeyName('');
      toast.success(t('deployKeysSettings.toastDeleted'));
    } catch (err: unknown) {
      const msg = getApiErrorMessage(err, { default: t('deployKeysSettings.errorFailedToDelete') });
      toast.error(msg);
    } finally {
      setDeleting(false);
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
            {t('deployKeysSettings.title')}
          </h3>
          <p className="text-sm mt-1" style={{ color: 'var(--text-dim)' }}>
            {t('deployKeysSettings.description')}
          </p>
        </div>
        <div className="space-y-4">
          {isLoading ? (
            <p className="text-sm" style={{ color: 'var(--text-dim)' }}>
              {t('deployKeysSettings.loading')}
            </p>
          ) : deployKeys.length === 0 ? (
            <div
              className="rounded-lg p-6 text-center"
              style={{ background: 'var(--panel-raised)', border: '1px solid var(--panel-edge)' }}
            >
              <p className="text-sm" style={{ color: 'var(--text-dim)' }}>
                {t('deployKeysSettings.noKeysYet')}
              </p>
              <Button
                className="mt-4 gap-2"
                onClick={handleOpenCreateModal}
                disabled={creating}
                style={{
                  background: 'linear-gradient(180deg, #ffffff, #d8dee2)',
                  color: 'var(--text-on-light)',
                  boxShadow: 'var(--shadow-btn-primary-rest)',
                }}
              >
                <Key className="h-4 w-4" />
                {t('deployKeysSettings.addDeployKey')}
              </Button>
            </div>
          ) : (
            <>
              {deployKeys.map((key) => (
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
                      {key.name}
                    </h4>
                    <p className="text-sm" style={{ color: 'var(--text-dim)' }}>
                      {t('deployKeysSettings.fingerprint', { fingerprint: key.fingerprint })}
                    </p>
                    <p className="text-xs" style={{ color: 'var(--text-faint)' }}>
                      {t('deployKeysSettings.createdAt', { date: formatDate(key.created_at) })}
                      {key.last_used_at &&
                        ` · ${t('deployKeysSettings.lastUsed', { date: formatDate(key.last_used_at) })}`}
                      {key.expires_at &&
                        ` · ${t('deployKeysSettings.expiresAt', { date: formatDate(key.expires_at) })}`}
                    </p>
                  </div>
                  <div className="flex flex-wrap items-center gap-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleVerifyClick(key)}
                      title={t('deployKeysSettings.verify')}
                      disabled={verifying && verifyKeyId === key.id}
                      style={{ color: 'var(--text-dim)' }}
                    >
                      {verificationSuccess && verifyKeyId === key.id ? (
                        <CheckCircle2 className="h-4 w-4" style={{ color: 'var(--status-ok)' }} />
                      ) : (
                        <RefreshCw
                          className={`h-4 w-4 ${verifying && verifyKeyId === key.id ? 'animate-spin' : ''}`}
                        />
                      )}
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleDeleteClick(key)}
                      title={t('deployKeysSettings.delete')}
                      style={{ color: 'var(--status-revoked)' }}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              ))}
              <Button
                className="gap-2"
                onClick={handleOpenCreateModal}
                disabled={creating}
                style={{
                  background: 'linear-gradient(180deg, #ffffff, #d8dee2)',
                  color: 'var(--text-on-light)',
                  boxShadow: 'var(--shadow-btn-primary-rest)',
                }}
              >
                <Key className="h-4 w-4" />
                {t('deployKeysSettings.addDeployKey')}
              </Button>
            </>
          )}
        </div>
      </div>

      <Dialog open={createModalOpen} onOpenChange={(open) => !open && setCreateModalOpen(false)}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{t('deployKeysSettings.createDialogTitle')}</DialogTitle>
            <DialogDescription>{t('deployKeysSettings.createDialogDescription')}</DialogDescription>
          </DialogHeader>
          <form onSubmit={handleCreateSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="deploy-key-name">{t('deployKeysSettings.nameLabel')}</Label>
              <Input
                id="deploy-key-name"
                value={createName}
                onChange={(e) => setCreateName(e.target.value)}
                placeholder={t('deployKeysSettings.namePlaceholder')}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="deploy-key-public">{t('deployKeysSettings.publicKeyLabel')}</Label>
              <Textarea
                id="deploy-key-public"
                value={createPublicKey}
                onChange={(e) => setCreatePublicKey(e.target.value)}
                placeholder={t('deployKeysSettings.publicKeyPlaceholder')}
                rows={4}
                required
              />
            </div>
            <DialogFooter>
              <Button
                type="button"
                onClick={() => setCreateModalOpen(false)}
                style={{ borderColor: 'var(--steel)', color: 'var(--text)' }}
              >
                {t('deployKeysSettings.cancel')}
              </Button>
              <Button
                type="submit"
                disabled={creating}
                style={{
                  background: 'linear-gradient(180deg, #ffffff, #d8dee2)',
                  color: 'var(--text-on-light)',
                  boxShadow: 'var(--shadow-btn-primary-rest)',
                }}
              >
                {creating ? t('deployKeysSettings.creating') : t('deployKeysSettings.create')}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={!!deleteKeyId} onOpenChange={(open) => !open && setDeleteKeyId(null)}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{t('deployKeysSettings.deleteDialogTitle')}</DialogTitle>
            <DialogDescription>
              {t('deployKeysSettings.deleteDialogDescription', { name: deleteKeyName })}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setDeleteKeyId(null)}
              style={{ borderColor: 'var(--steel)', color: 'var(--text)' }}
            >
              {t('deployKeysSettings.cancel')}
            </Button>
            <Button variant="destructive" onClick={handleDeleteConfirm} disabled={deleting}>
              {deleting ? t('deployKeysSettings.deleting') : t('deployKeysSettings.delete')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
