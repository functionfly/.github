import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import {
  getKeyHistory,
  getSignerStatus,
  rotateKey,
  testSigner,
  type SigningKeyRecord,
  type SignerStatus,
  type SignerTestResult,
} from '@/api/trustapi';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Cloud, Code, Copy, Key, Shield, ShieldCheck, ShieldX, Loader2 } from 'lucide-react';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';

const SIGNER_BACKEND_ICONS: Record<string, typeof Code> = {
  software: Code,
  pkcs11: Shield,
  awskms: Cloud,
};

const SIGNER_BACKEND_LABELS: Record<string, string> = {
  software: 'Software',
  pkcs11: 'PKCS#11 HSM',
  awskms: 'AWS KMS',
};

function truncateHex(hex: string, startChars = 8, endChars = 6): string {
  if (hex.length <= startChars + endChars + 3) return hex;
  return `${hex.slice(0, startChars)}...${hex.slice(-endChars)}`;
}

function formatDate(dateString: string): string {
  return new Intl.DateTimeFormat('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  }).format(new Date(dateString));
}

interface SignerSettingsCardProps {
  userPlan: string;
}

export function SignerSettingsCard({ userPlan }: SignerSettingsCardProps) {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [testResult, setTestResult] = useState<SignerTestResult | null>(null);
  const [showRotateConfirm, setShowRotateConfirm] = useState(false);

  const { data: signerStatus, isLoading: statusLoading } = useQuery({
    queryKey: ['trustapi', 'signer', 'status'],
    queryFn: getSignerStatus,
    retry: false,
  });

  const { data: keyHistory, isLoading: keyHistoryLoading } = useQuery({
    queryKey: ['trustapi', 'signer', 'keys'],
    queryFn: getKeyHistory,
    retry: false,
  });

  const testMutation = useMutation({
    mutationFn: testSigner,
    onSuccess: (result) => {
      setTestResult(result);
      if (result.pass) {
        toast.success(t('signerSettings.testSuccess'));
      } else {
        toast.error(result.error || t('signerSettings.testFailed'));
      }
    },
    onError: (err) => {
      toast.error(t('signerSettings.testError'));
      console.error('Signer test error:', err);
    },
  });

  const rotateMutation = useMutation({
    mutationFn: rotateKey,
    onSuccess: () => {
      toast.success(t('signerSettings.rotateSuccess'));
      queryClient.invalidateQueries({ queryKey: ['trustapi', 'signer'] });
      setShowRotateConfirm(false);
    },
    onError: (err) => {
      toast.error(t('signerSettings.rotateError'));
      console.error('Key rotation error:', err);
    },
  });

  const isProPlan = ['pro', 'enterprise', 'agent_enterprise'].includes(userPlan?.toLowerCase());

  const handleCopyPublicKey = async (hex: string) => {
    try {
      await navigator.clipboard.writeText(hex);
      toast.success(t('common.copied'));
    } catch {
      toast.error(t('signerSettings.copyFailed'));
    }
  };

  const handleTestSigner = () => {
    setTestResult(null);
    testMutation.mutate();
  };

  const handleRotateKey = () => {
    rotateMutation.mutate();
  };

  const BackendIcon = signerStatus ? SIGNER_BACKEND_ICONS[signerStatus.backend] || Shield : Shield;

  return (
    <Card className="ff-card-velocity">
      <CardHeader>
        <CardTitle className="font-display flex items-center gap-2">
          <Key className="h-5 w-5 text-brand-500" />
          {t('signerSettings.title')}
        </CardTitle>
        <CardDescription className="text-text-secondary">
          {t('signerSettings.description')}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {statusLoading ? (
          <div className="flex items-center justify-center p-4">
            <Loader2 className="h-6 w-6 animate-spin text-brand-500" />
          </div>
        ) : !signerStatus ? (
          <div className="p-4 rounded-lg bg-amber-500/10 border border-amber-500/20">
            <p className="text-amber-500 text-sm">{t('signerSettings.notAvailable')}</p>
          </div>
        ) : (
          <>
            <div className="flex items-start justify-between p-4 rounded-lg bg-linear-to-r from-brand-500/10 to-brand-600/10 border border-border-default">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-bg-secondary">
                  <BackendIcon className="h-5 w-5 text-brand-500" />
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <h3 className="font-semibold font-display text-text-primary">
                      {SIGNER_BACKEND_LABELS[signerStatus.backend] || signerStatus.backend}
                    </h3>
                    <div
                      className={`w-2 h-2 rounded-full ${
                        signerStatus.healthy ? 'bg-green-500' : 'bg-red-500'
                      }`}
                    />
                  </div>
                  <p className="text-sm text-text-muted mt-0.5">
                    {t('signerSettings.algorithm')}: {signerStatus.algorithm}
                  </p>
                </div>
              </div>
              <Badge
                variant={signerStatus.healthy ? 'success' : 'secondary'}
                className={signerStatus.healthy ? 'ff-badge-success' : ''}
              >
                {signerStatus.healthy ? t('signerSettings.healthy') : t('signerSettings.unhealthy')}
              </Badge>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div className="p-3 rounded-lg bg-bg-secondary border border-border-default">
                <p className="text-xs text-text-muted mb-1">{t('signerSettings.keyId')}</p>
                <p className="font-mono text-sm text-text-primary truncate">{signerStatus.key_id}</p>
              </div>
              <div className="p-3 rounded-lg bg-bg-secondary border border-border-default">
                <p className="text-xs text-text-muted mb-1">{t('signerSettings.latency')}</p>
                <p className="font-mono text-sm text-text-primary">
                  {signerStatus.latency_ms != null ? `${signerStatus.latency_ms}ms` : '—'}
                </p>
              </div>
            </div>

            <div className="p-3 rounded-lg bg-bg-secondary border border-border-default">
              <div className="flex items-center justify-between mb-2">
                <p className="text-xs text-text-muted">{t('signerSettings.publicKey')}</p>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-6 px-2 text-xs"
                  onClick={() => handleCopyPublicKey(signerStatus.public_key_hex)}
                >
                  <Copy className="h-3 w-3 mr-1" />
                  {t('common.copy')}
                </Button>
              </div>
              <p className="font-mono text-xs text-text-primary break-all">
                {truncateHex(signerStatus.public_key_hex, 16, 12)}
              </p>
            </div>

            {testResult && (
              <div
                className={`p-4 rounded-lg border ${
                  testResult.pass
                    ? 'bg-green-500/10 border-green-500/20'
                    : 'bg-red-500/10 border-red-500/20'
                }`}
              >
                <div className="flex items-center gap-2 mb-2">
                  {testResult.pass ? (
                    <ShieldCheck className="h-4 w-4 text-green-500" />
                  ) : (
                    <ShieldX className="h-4 w-4 text-red-500" />
                  )}
                  <span
                    className={`font-medium text-sm ${
                      testResult.pass ? 'text-green-400' : 'text-red-400'
                    }`}
                  >
                    {testResult.pass ? t('signerSettings.testPassed') : t('signerSettings.testFailed')}
                  </span>
                </div>
                {testResult.error && (
                  <p className="text-xs text-red-400 mb-2">{testResult.error}</p>
                )}
                <div className="grid grid-cols-2 gap-2 text-xs text-text-muted">
                  <span>
                    {t('signerSettings.signLatency')}: {testResult.sign_latency_ms}ms
                  </span>
                  <span>
                    {t('signerSettings.verifyLatency')}: {testResult.verify_latency_ms}ms
                  </span>
                </div>
              </div>
            )}

            <div className="flex items-center gap-3">
              <Button
                variant="outline"
                size="sm"
                onClick={handleTestSigner}
                disabled={testMutation.isPending}
              >
                {testMutation.isPending ? (
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                ) : (
                  <ShieldCheck className="h-4 w-4 mr-2" />
                )}
                {t('signerSettings.testButton')}
              </Button>

              {isProPlan ? (
                <AlertDialog open={showRotateConfirm} onOpenChange={setShowRotateConfirm}>
                  <AlertDialogTrigger asChild>
                    <Button variant="outline" size="sm">
                      <Key className="h-4 w-4 mr-2" />
                      {t('signerSettings.rotateButton')}
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent
                    style={{
                      background: 'var(--panel)',
                      borderColor: 'var(--panel-edge)',
                      borderRadius: 'var(--radius-lg)',
                    }}
                  >
                    <AlertDialogHeader>
                      <AlertDialogTitle>{t('signerSettings.rotateConfirmTitle')}</AlertDialogTitle>
                      <AlertDialogDescription style={{ color: 'var(--text-dim)' }}>
                        {t('signerSettings.rotateConfirmDescription')}
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>{t('common.cancel')}</AlertDialogCancel>
                      <AlertDialogAction
                        onClick={handleRotateKey}
                        disabled={rotateMutation.isPending}
                        style={{ backgroundColor: 'var(--red-600)' }}
                      >
                        {rotateMutation.isPending ? (
                          <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                        ) : null}
                        {t('signerSettings.rotateButton')}
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              ) : (
                <AlertDialog>
                  <AlertDialogTrigger asChild>
                    <Button variant="outline" size="sm" disabled>
                      <Key className="h-4 w-4 mr-2" />
                      {t('signerSettings.rotateButton')}
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent
                    style={{
                      background: 'var(--panel)',
                      borderColor: 'var(--panel-edge)',
                      borderRadius: 'var(--radius-lg)',
                    }}
                  >
                    <AlertDialogHeader>
                      <AlertDialogTitle>{t('signerSettings.rotateRequiresPro')}</AlertDialogTitle>
                      <AlertDialogDescription style={{ color: 'var(--text-dim)' }}>
                        {t('signerSettings.rotateProDescription')}
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogAction asChild>
                        <Button variant="outline" onClick={() => window.location.href = '/settings/billing'}>
                          {t('signerSettings.upgradePlan')}
                        </Button>
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              )}
            </div>

            {keyHistory && keyHistory.keys.length > 0 && (
              <div className="mt-6">
                <h4 className="font-medium text-sm text-text-primary mb-3">
                  {t('signerSettings.keyHistory')}
                </h4>
                <div className="rounded-lg border border-border-default overflow-hidden">
                  <table className="w-full text-sm">
                    <thead className="bg-bg-secondary">
                      <tr>
                        <th className="px-3 py-2 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                          {t('signerSettings.keyIdCol')}
                        </th>
                        <th className="px-3 py-2 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                          {t('signerSettings.algorithmCol')}
                        </th>
                        <th className="px-3 py-2 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                          {t('signerSettings.activatedCol')}
                        </th>
                        <th className="px-3 py-2 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                          {t('signerSettings.statusCol')}
                        </th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-border-default">
                      {keyHistoryLoading ? (
                        <tr>
                          <td colSpan={4} className="px-3 py-4 text-center text-text-muted">
                            <Loader2 className="h-4 w-4 animate-spin mx-auto" />
                          </td>
                        </tr>
                      ) : (
                        keyHistory.keys.map((key: SigningKeyRecord) => (
                          <tr key={key.id} className="bg-bg-primary">
                            <td className="px-3 py-2">
                              <span className="font-mono text-xs text-text-primary">
                                {key.key_id.slice(0, 12)}...
                              </span>
                            </td>
                            <td className="px-3 py-2 text-text-secondary text-xs">
                              {key.algorithm}
                            </td>
                            <td className="px-3 py-2 text-text-secondary text-xs">
                              {formatDate(key.activated_at)}
                            </td>
                            <td className="px-3 py-2">
                              <Badge
                                variant={key.is_active ? 'success' : 'secondary'}
                                className={key.is_active ? 'ff-badge-success' : ''}
                              >
                                {key.is_active
                                  ? t('signerSettings.active')
                                  : t('signerSettings.inactive')}
                              </Badge>
                            </td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            )}
          </>
        )}
      </CardContent>
    </Card>
  );
}
