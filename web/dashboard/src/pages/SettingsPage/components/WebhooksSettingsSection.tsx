import {
  functionWebhooksApi,
  type FunctionWebhook,
  type WebhookDelivery,
} from '@/api/function-webhooks';
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
import { useQuery } from '@tanstack/react-query';
import { CheckCircle2, ChevronRight, RefreshCw, Trash2, Webhook, XCircle } from 'lucide-react';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { formatDate, getApiErrorMessage } from '../settings-utils';

const EVENT_TYPE_OPTIONS = [
  'function.executed',
  'function.failed',
  'function.completed',
  'function.deployed',
  'graph.executed',
  'graph.failed',
];

export function WebhooksSettingsSection() {
  const { t } = useTranslation();
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [createUrl, setCreateUrl] = useState('');
  const [createEventTypes, setCreateEventTypes] = useState<string[]>(['function.executed']);
  const [createSecret, setCreateSecret] = useState('');
  const [creating, setCreating] = useState(false);

  const [viewDeliveryKey, setViewDeliveryKey] = useState<{
    webhookId: string;
    delivery: WebhookDelivery;
  } | null>(null);
  const [testWebhookId, setTestWebhookId] = useState<string | null>(null);
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<{ success: boolean; delivery_id?: string } | null>(
    null
  );

  const [deleteWebhookId, setDeleteWebhookId] = useState<string | null>(null);
  const [deleteWebhookName, setDeleteWebhookName] = useState('');
  const [deleting, setDeleting] = useState(false);

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['function-webhooks'],
    queryFn: async () => {
      try {
        return await functionWebhooksApi.list();
      } catch (e: unknown) {
        const status = (e as { response?: { status?: number } })?.response?.status;
        if (status === 404) return { subscriptions: [], total_count: 0, page: 1, page_size: 20 };
        throw e;
      }
    },
    enabled: true,
    retry: false,
  });

  const webhooks: FunctionWebhook[] = data?.subscriptions ?? [];

  const handleOpenCreateModal = () => {
    setCreateUrl('');
    setCreateEventTypes(['function.executed']);
    setCreateSecret('');
    setCreateModalOpen(true);
  };

  const handleCreateSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!createUrl.trim()) {
      toast.error(t('webhooksSettings.toastUrlRequired'));
      return;
    }
    setCreating(true);
    try {
      await functionWebhooksApi.create({
        url: createUrl.trim(),
        event_types: createEventTypes,
        secret: createSecret || undefined,
      });
      refetch();
      setCreateModalOpen(false);
      toast.success(t('webhooksSettings.toastCreated'));
    } catch (err: unknown) {
      const msg = getApiErrorMessage(err, { default: t('webhooksSettings.errorFailedToCreate') });
      toast.error(msg);
    } finally {
      setCreating(false);
    }
  };

  const handleToggleEventType = (eventType: string) => {
    setCreateEventTypes((prev) =>
      prev.includes(eventType) ? prev.filter((e) => e !== eventType) : [...prev, eventType]
    );
  };

  const handleTestClick = (webhook: FunctionWebhook) => {
    setTestWebhookId(webhook.id);
    setTestResult(null);
    setTesting(true);
    functionWebhooksApi
      .test(webhook.id)
      .then((result) => {
        setTestResult(result);
        if (result.success) {
          toast.success(t('webhooksSettings.testSuccess'));
        } else {
          toast.error(t('webhooksSettings.testFailed'));
        }
      })
      .catch(() => {
        setTestResult({ success: false });
        toast.error(t('webhooksSettings.testFailed'));
      })
      .finally(() => {
        setTesting(false);
      });
  };

  const handleViewDeliveriesClick = async (webhook: FunctionWebhook) => {
    try {
      const deliveries = await functionWebhooksApi.listDeliveries(webhook.id);
      if (deliveries.deliveries.length > 0) {
        setViewDeliveryKey({ webhookId: webhook.id, delivery: deliveries.deliveries[0] });
      } else {
        toast.info(t('webhooksSettings.noDeliveriesYet'));
      }
    } catch {
      toast.error(t('webhooksSettings.errorLoadingDeliveries'));
    }
  };

  const handleDeleteClick = (webhook: FunctionWebhook) => {
    setDeleteWebhookId(webhook.id);
    setDeleteWebhookName(webhook.url);
  };

  const handleDeleteConfirm = async () => {
    if (!deleteWebhookId) return;
    setDeleting(true);
    try {
      await functionWebhooksApi.delete(deleteWebhookId);
      refetch();
      setDeleteWebhookId(null);
      setDeleteWebhookName('');
      toast.success(t('webhooksSettings.toastDeleted'));
    } catch (err: unknown) {
      const msg = getApiErrorMessage(err, { default: t('webhooksSettings.errorFailedToDelete') });
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
            {t('webhooksSettings.title')}
          </h3>
          <p className="text-sm mt-1" style={{ color: 'var(--text-dim)' }}>
            {t('webhooksSettings.description')}
          </p>
        </div>
        <div className="space-y-4">
          {isLoading ? (
            <p className="text-sm" style={{ color: 'var(--text-dim)' }}>
              {t('webhooksSettings.loading')}
            </p>
          ) : webhooks.length === 0 ? (
            <div
              className="rounded-lg p-6 text-center"
              style={{ background: 'var(--panel-raised)', border: '1px solid var(--panel-edge)' }}
            >
              <p className="text-sm" style={{ color: 'var(--text-dim)' }}>
                {t('webhooksSettings.noWebhooksYet')}
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
                <Webhook className="h-4 w-4" />
                {t('webhooksSettings.addWebhook')}
              </Button>
            </div>
          ) : (
            <>
              {webhooks.map((webhook) => (
                <div
                  key={webhook.id}
                  className="flex flex-wrap items-center justify-between gap-3 rounded-lg p-4"
                  style={{
                    background: 'var(--panel-raised)',
                    border: '1px solid var(--panel-edge)',
                  }}
                >
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                      <h4 className="font-medium truncate" style={{ color: 'var(--text)' }}>
                        {webhook.url}
                      </h4>
                      <span
                        className="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium"
                        style={{
                          background: webhook.active ? 'rgba(143, 255, 208, 0.06)' : 'var(--panel)',
                          color: webhook.active ? 'var(--status-ok)' : 'var(--text-dim)',
                        }}
                      >
                        {webhook.active
                          ? t('webhooksSettings.active')
                          : t('webhooksSettings.inactive')}
                      </span>
                    </div>
                    <p className="text-sm" style={{ color: 'var(--text-dim)' }}>
                      {t('webhooksSettings.events', { events: webhook.event_types.join(', ') })}
                    </p>
                    <p className="text-xs" style={{ color: 'var(--text-faint)' }}>
                      {t('webhooksSettings.createdAt', { date: formatDate(webhook.created_at) })}
                    </p>
                  </div>
                  <div className="flex flex-wrap items-center gap-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleTestClick(webhook)}
                      title={t('webhooksSettings.test')}
                      disabled={testing && testWebhookId === webhook.id}
                      style={{ color: 'var(--text-dim)' }}
                    >
                      {testResult && testWebhookId === webhook.id ? (
                        testResult.success ? (
                          <CheckCircle2 className="h-4 w-4" style={{ color: 'var(--status-ok)' }} />
                        ) : (
                          <XCircle className="h-4 w-4" style={{ color: 'var(--status-revoked)' }} />
                        )
                      ) : (
                        <RefreshCw
                          className={`h-4 w-4 ${testing && testWebhookId === webhook.id ? 'animate-spin' : ''}`}
                        />
                      )}
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleViewDeliveriesClick(webhook)}
                      title={t('webhooksSettings.viewDeliveries')}
                      style={{ color: 'var(--text-dim)' }}
                    >
                      <ChevronRight className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleDeleteClick(webhook)}
                      title={t('webhooksSettings.delete')}
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
                <Webhook className="h-4 w-4" />
                {t('webhooksSettings.addWebhook')}
              </Button>
            </>
          )}
        </div>
      </div>

      <Dialog open={createModalOpen} onOpenChange={(open) => !open && setCreateModalOpen(false)}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{t('webhooksSettings.createDialogTitle')}</DialogTitle>
            <DialogDescription>{t('webhooksSettings.createDialogDescription')}</DialogDescription>
          </DialogHeader>
          <form onSubmit={handleCreateSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="webhook-url">{t('webhooksSettings.urlLabel')}</Label>
              <Input
                id="webhook-url"
                type="url"
                value={createUrl}
                onChange={(e) => setCreateUrl(e.target.value)}
                placeholder="https://example.com/webhook"
                required
              />
            </div>
            <div className="space-y-2">
              <Label style={{ color: 'var(--text)' }}>
                {t('webhooksSettings.eventTypesLabel')}
              </Label>
              <div className="flex flex-wrap gap-2">
                {EVENT_TYPE_OPTIONS.map((eventType) => {
                  const isSelected = createEventTypes.includes(eventType);
                  return (
                    <button
                      key={eventType}
                      type="button"
                      onClick={() => handleToggleEventType(eventType)}
                      className="inline-flex items-center justify-center gap-1.5 rounded-md border px-3 py-1.5 text-xs font-medium transition-all duration-200"
                      style={{
                        borderColor: isSelected ? 'var(--accent)' : 'var(--panel-edge)',
                        background: isSelected ? 'rgba(143, 255, 208, 0.06)' : 'var(--panel)',
                        color: isSelected ? 'var(--accent)' : 'var(--text-dim)',
                        boxShadow: isSelected ? '0 0 12px rgba(143, 255, 208, 0.15)' : 'none',
                      }}
                    >
                      {isSelected && (
                        <svg className="h-3 w-3" fill="currentColor" viewBox="0 0 8 8">
                          <path d="M6.4 1.6a.4.4 0 01.6.6L3.2 7.6a.4.4 0 01-.6 0L1.6 6.4a.4.4 0 01.6-.6L3 6.2 5.8 2.2a.4.4 0 01.6-.6z" />
                        </svg>
                      )}
                      {eventType}
                    </button>
                  );
                })}
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="webhook-secret" style={{ color: 'var(--text)' }}>
                {t('webhooksSettings.secretLabel')}
              </Label>
              <Input
                id="webhook-secret"
                type="password"
                value={createSecret}
                onChange={(e) => setCreateSecret(e.target.value)}
                placeholder={t('webhooksSettings.secretPlaceholder')}
              />
              <p className="text-xs" style={{ color: 'var(--text-faint)' }}>
                {t('webhooksSettings.secretHint')}
              </p>
            </div>
            <DialogFooter>
              <Button
                type="button"
                onClick={() => setCreateModalOpen(false)}
                style={{ borderColor: 'var(--steel)', color: 'var(--text)' }}
              >
                {t('webhooksSettings.cancel')}
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
                {creating ? t('webhooksSettings.creating') : t('webhooksSettings.create')}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={!!deleteWebhookId} onOpenChange={(open) => !open && setDeleteWebhookId(null)}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{t('webhooksSettings.deleteDialogTitle')}</DialogTitle>
            <DialogDescription>
              {t('webhooksSettings.deleteDialogDescription', { url: deleteWebhookName })}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setDeleteWebhookId(null)}
              style={{ borderColor: 'var(--steel)', color: 'var(--text)' }}
            >
              {t('webhooksSettings.cancel')}
            </Button>
            <Button variant="destructive" onClick={handleDeleteConfirm} disabled={deleting}>
              {deleting ? t('webhooksSettings.deleting') : t('webhooksSettings.delete')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
