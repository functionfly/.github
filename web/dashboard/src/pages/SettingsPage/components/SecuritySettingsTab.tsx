import { authApi, type MFASetupResponse, type MFAStatusResponse } from '@/api/auth';
import { usersApi, type SessionItem } from '@/api/users';
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
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Switch } from '@/components/ui/switch';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useAuthStore } from '@/stores/authStore';
import { Icon } from '@iconify/react';
import {
  AlertTriangle,
  CheckCircle2,
  Clock,
  Download,
  Globe,
  KeyRound,
  Laptop,
  ListFilter,
  LogOut,
  Monitor,
  RefreshCw,
  Server,
  ShieldAlert,
  ShieldCheck,
  Smartphone,
  Tablet,
  Trash2,
} from 'lucide-react';
import { QRCodeSVG } from 'qrcode.react';
import { differenceInMinutes } from 'date-fns';
import { useEffect, useMemo, useState } from 'react';
import { Trans, useTranslation } from 'react-i18next';
import { toast } from 'sonner';

function detectCurrentDeviceLabel(): string {
  if (typeof navigator === 'undefined') return 'Current browser';
  const ua = navigator.userAgent;
  if (ua.includes('Edg/')) return 'Edge (this device)';
  if (ua.includes('Chrome/') && !ua.includes('Edg/')) return 'Chrome (this device)';
  if (ua.includes('Firefox/')) return 'Firefox (this device)';
  if (ua.includes('Safari/') && !ua.includes('Chrome/')) return 'Safari (this device)';
  if (ua.includes('Windows')) return 'Desktop (Windows)';
  if (ua.includes('Mac')) return 'Desktop (macOS)';
  if (ua.includes('Linux')) return 'Desktop (Linux)';
  if (ua.includes('Android') || ua.includes('iPhone')) return 'Mobile (this device)';
  return 'Current browser';
}

function getSessionSimpleIcon(deviceLabel: string): string | null {
  const normalized = deviceLabel.toLowerCase();
  if (normalized.includes('chrome')) return 'simple-icons:googlechrome';
  if (normalized.includes('firefox')) return 'simple-icons:firefoxbrowser';
  if (normalized.includes('safari')) return 'simple-icons:safari';
  if (normalized.includes('edge')) return 'simple-icons:microsoftedge';
  if (normalized.includes('opera')) return 'simple-icons:opera';
  if (normalized.includes('brave')) return 'simple-icons:brave';
  if (
    normalized.includes('mobile') ||
    normalized.includes('iphone') ||
    normalized.includes('android')
  ) {
    return null;
  }
  return null;
}

function getSessionDeviceIcon(deviceLabel: string, size: 'sm' | 'lg' = 'sm'): React.ReactNode {
  const normalized = deviceLabel.toLowerCase();
  const sm = size === 'sm';
  const baseClass = sm ? 'h-4 w-4' : 'h-5 w-5';

  if (normalized.includes('mac') || normalized.includes('darwin')) {
    return <Laptop className={baseClass} />;
  }
  if (normalized.includes('windows')) {
    return <Monitor className={baseClass} />;
  }
  if (normalized.includes('linux')) {
    return <Server className={baseClass} />;
  }
  if (normalized.includes('mobile') || normalized.includes('iphone') || normalized.includes('android') || normalized.includes('tablet') || normalized.includes('ipad')) {
    return <Smartphone className={baseClass} />;
  }
  if (normalized.includes('tablet')) {
    return <Tablet className={baseClass} />;
  }
  if (normalized.includes('chrome')) return <Icon icon="simple-icons:googlechrome" className={baseClass} />;
  if (normalized.includes('firefox')) return <Icon icon="simple-icons:firefoxbrowser" className={baseClass} />;
  if (normalized.includes('safari')) return <Icon icon="simple-icons:safari" className={baseClass} />;
  if (normalized.includes('edge')) return <Icon icon="simple-icons:microsoftedge" className={baseClass} />;

  return <Monitor className={`${baseClass} text-text-muted`} />;
}

export function SecuritySettingsTab() {
  const { t } = useTranslation();
  const [status, setStatus] = useState<MFAStatusResponse | null>(null);
  const [loadingStatus, setLoadingStatus] = useState(true);
  const [working, setWorking] = useState(false);

  const [setupData, setSetupData] = useState<MFASetupResponse | null>(null);
  const [verificationCode, setVerificationCode] = useState('');
  const [recoveryCodes, setRecoveryCodes] = useState<string[]>([]);
  const [sessions, setSessions] = useState<SessionItem[]>([]);
  const [loadingSessions, setLoadingSessions] = useState(true);
  const [revokingSessionId, setRevokingSessionId] = useState<string | null>(null);
  const [revokingOthers, setRevokingOthers] = useState(false);
  const [exportingData, setExportingData] = useState(false);
  const [deleteConfirmation, setDeleteConfirmation] = useState('');
  const [deletingAccount, setDeletingAccount] = useState(false);
  const user = useAuthStore((state) => state.user);
  const logout = useAuthStore((state) => state.logout);

  const [disablePassword, setDisablePassword] = useState('');
  const [disableCode, setDisableCode] = useState('');
  const [showMfaSecret, setShowMfaSecret] = useState(false);

  const [timeoutModalOpen, setTimeoutModalOpen] = useState(false);
  const [selectedTimeout, setSelectedTimeout] = useState<string>('7d');
  const [savingTimeout, setSavingTimeout] = useState(false);
  const [rememberDevices, setRememberDevices] = useState(true);
  const [loadingSettings, setLoadingSettings] = useState(true);

  const timeoutOptions = [
    { value: '1h', label: '1 hour', description: 'Sign out after 1 hour of inactivity' },
    { value: '24h', label: '24 hours', description: 'Sign out after 1 day of inactivity' },
    { value: '7d', label: '7 days', description: 'Sign out after 1 week of inactivity' },
    { value: '30d', label: '30 days', description: 'Sign out after 30 days of inactivity' },
    { value: 'never', label: 'Never', description: 'Do not automatically sign out' },
  ];

  useEffect(() => {
    const loadSettings = async () => {
      try {
        const data = await usersApi.getMySettings();
        const settings = (data as { settings?: Record<string, unknown> }).settings ?? {};
        if (settings.sessionTimeout) {
          setSelectedTimeout(settings.sessionTimeout as string);
        }
        if (typeof settings.rememberDevices === 'boolean') {
          setRememberDevices(settings.rememberDevices);
        }
      } catch {
        // use defaults
      } finally {
        setLoadingSettings(false);
      }
    };
    void loadSettings();
  }, []);

  const handleSaveTimeout = async () => {
    setSavingTimeout(true);
    try {
      await usersApi.updateMySecuritySettings({ sessionTimeout: selectedTimeout });
      toast.success(t('securitySettings.toastTimeoutUpdated', 'Session timeout updated'));
    } catch {
      toast.error(t('securitySettings.toastTimeoutFailed', 'Failed to update session timeout'));
    } finally {
      setSavingTimeout(false);
      setTimeoutModalOpen(false);
    }
  };

  const handleRememberDevicesChange = async (checked: boolean) => {
    setRememberDevices(checked);
    try {
      await usersApi.updateMySecuritySettings({ rememberDevices: checked });
      toast.success(t('securitySettings.toastRememberDevicesUpdated', 'Trusted device setting updated'));
    } catch {
      setRememberDevices(!checked);
      toast.error(t('securitySettings.toastRememberDevicesFailed', 'Failed to update trusted device setting'));
    }
  };

  const qrCodeValue = useMemo(() => {
    return setupData?.qr_code_url ?? '';
  }, [setupData?.qr_code_url]);

  const loadStatus = async () => {
    setLoadingStatus(true);
    try {
      const nextStatus = await authApi.getMFAStatus();
      setStatus(nextStatus);
    } catch {
      toast.error(t('securitySettings.toastFailedToLoadMfaStatus'));
    } finally {
      setLoadingStatus(false);
    }
  };

  useEffect(() => {
    void loadStatus();
  }, []);

  const loadSessions = async () => {
    setLoadingSessions(true);
    try {
      const data = await usersApi.listSessions();
      setSessions(data.sessions ?? []);
    } catch {
      toast.error(t('securitySettings.toastFailedToLoadSessions'));
    } finally {
      setLoadingSessions(false);
    }
  };

  useEffect(() => {
    void loadSessions();
  }, []);

  const handleStartSetup = async () => {
    setWorking(true);
    try {
      const data = await authApi.setupMFA();
      setSetupData(data);
      setRecoveryCodes(data.backup_codes ?? []);
      setVerificationCode('');
      setShowMfaSecret(false);
      toast.success(t('securitySettings.toastMfaSetupStarted'));
    } catch {
      toast.error(t('securitySettings.toastFailedToStartMfaSetup'));
    } finally {
      setWorking(false);
    }
  };

  const handleEnableMFA = async () => {
    if (!verificationCode.trim()) {
      toast.error(t('securitySettings.toastEnter6DigitCode'));
      return;
    }
    setWorking(true);
    try {
      const verify = await authApi.verifyMFASetupCode(verificationCode.trim());
      if (!verify.verified) {
        toast.error(t('securitySettings.toastVerificationCodeInvalid'));
        return;
      }
      await authApi.enableMFA();
      toast.success(t('securitySettings.toastMfaEnabled'));
      setSetupData(null);
      setVerificationCode('');
      await loadStatus();
    } catch {
      toast.error(t('securitySettings.toastFailedToEnableMfa'));
    } finally {
      setWorking(false);
    }
  };

  const handleDisableMFA = async () => {
    if (!disablePassword.trim() || !disableCode.trim()) {
      toast.error(t('securitySettings.toastPasswordAndMfaRequired'));
      return;
    }
    setWorking(true);
    try {
      await authApi.disableMFA(disablePassword.trim(), disableCode.trim());
      toast.success(t('securitySettings.toastMfaDisabled'));
      setDisablePassword('');
      setDisableCode('');
      await loadStatus();
    } catch {
      toast.error(t('securitySettings.toastFailedToDisableMfa'));
    } finally {
      setWorking(false);
    }
  };

  const handleRevokeSession = async (sessionId: string) => {
    setRevokingSessionId(sessionId);
    try {
      await usersApi.revokeSession(sessionId);
      toast.success(t('securitySettings.toastSessionRevoked'));
      await loadSessions();
    } catch {
      toast.error(t('securitySettings.toastFailedToRevokeSession'));
    } finally {
      setRevokingSessionId(null);
    }
  };

  const handleRevokeOthers = async () => {
    setRevokingOthers(true);
    try {
      await usersApi.revokeOtherSessions();
      toast.success(t('securitySettings.toastOtherSessionsRevoked'));
      await loadSessions();
    } catch {
      toast.error(t('securitySettings.toastFailedToRevokeOtherSessions'));
    } finally {
      setRevokingOthers(false);
    }
  };

  const handleExportData = async () => {
    setExportingData(true);
    try {
      const [me, settings, activeSessions] = await Promise.all([
        usersApi.getMe(),
        usersApi.getMySettings(),
        usersApi.listSessions(),
      ]);

      const payload = {
        exportedAt: new Date().toISOString(),
        account: me,
        settings: settings.settings ?? {},
        sessions: activeSessions.sessions ?? [],
      };

      const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const anchor = document.createElement('a');
      const date = new Date().toISOString().slice(0, 10);
      const username = user?.username || 'account';
      anchor.href = url;
      anchor.download = `functionfly-${username}-data-export-${date}.json`;
      document.body.appendChild(anchor);
      anchor.click();
      document.body.removeChild(anchor);
      URL.revokeObjectURL(url);
      toast.success(t('securitySettings.toastDataExportReady'));
    } catch {
      toast.error(t('securitySettings.toastFailedToExportData'));
    } finally {
      setExportingData(false);
    }
  };

  const handleDeleteAccount = async () => {
    if (deleteConfirmation !== 'DELETE') {
      toast.error(t('securitySettings.toastTypeDeleteToConfirm'));
      return;
    }

    setDeletingAccount(true);
    try {
      await usersApi.deleteMe();
      toast.success(t('securitySettings.toastAccountDeleted'));
      await logout();
      window.location.href = '/';
      return;
    } catch {
      const subject = encodeURIComponent('Account deletion request');
      const body = encodeURIComponent(
        `Please delete my FunctionFly account.\n\nUsername: ${user?.username || 'unknown'}\nEmail: ${
          user?.email || 'unknown'
        }\n\nI understand this action is permanent.`
      );
      window.location.href = `mailto:support@functionfly.com?subject=${subject}&body=${body}`;
      toast.info(t('securitySettings.toastDirectDeleteUnavailable'));
    } finally {
      setDeletingAccount(false);
      setDeleteConfirmation('');
    }
  };

  const handleDownloadRecoveryCodes = async () => {
    if (recoveryCodes.length === 0) return;

    const username = user?.username || 'account';
    const date = new Date().toISOString().slice(0, 10);
    const content = [
      'FunctionFly MFA Recovery Codes',
      '',
      'Each code can be used once.',
      `GeneratedAt: ${new Date().toISOString()}`,
      `Username: ${username}`,
      '',
      recoveryCodes.join('\n'),
      '',
    ].join('\n');

    const blob = new Blob([content], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement('a');
    anchor.href = url;
    anchor.download = `functionfly-${username}-mfa-recovery-codes-${date}.txt`;
    document.body.appendChild(anchor);
    anchor.click();
    document.body.removeChild(anchor);
    URL.revokeObjectURL(url);

    toast.success(t('securitySettings.toastRecoveryCodesDownloaded'));
  };

  const displaySessions: SessionItem[] = useMemo(() => {
    if (sessions.length > 0) return sessions;
    return [
      {
        id: 'current-browser-fallback',
        device: detectCurrentDeviceLabel(),
        ip: 'Current network',
        location: '',
        lastActive: 'Now',
        currentSession: true,
      },
    ];
  }, [sessions]);

  return (
    <div className="settings-page space-y-6">
      <Card className="settings-panel">
        <CardHeader>
          <CardTitle className="font-display flex items-center gap-2">
            <ShieldCheck className="h-5 w-5 text-brand-500" />
            {t('securitySettings.mfaTitle')}
          </CardTitle>
          <CardDescription className="text-text-secondary">
            {t('securitySettings.mfaDescription')}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {loadingStatus ? (
            <p className="text-sm text-text-muted">{t('securitySettings.loadingSecurityStatus')}</p>
          ) : (
            <div className="flex flex-wrap items-center gap-2">
              <Badge variant={status?.enabled ? 'default' : 'secondary'} className={status?.enabled ? 'ff-badge-primary' : ''}>
                {status?.enabled ? t('securitySettings.mfaEnabledBadge') : t('securitySettings.mfaDisabledBadge')}
              </Badge>
              {status?.required ? (
                <Badge
                  variant="outline"
                  className="border-amber-500/50 text-amber-600 dark:text-amber-400"
                >
                  {t('securitySettings.requiredForYourRole')}
                </Badge>
              ) : null}
              {typeof status?.backup_codes_remaining === 'number' ? (
                <Badge variant="outline">{t('securitySettings.backupCodesCount', { count: status.backup_codes_remaining })}</Badge>
              ) : null}
            </div>
          )}

          {!status?.enabled ? (
            <div className="rounded-lg border border-border-default bg-bg-secondary p-4 space-y-3">
              <p className="text-sm text-text-secondary">
                {t('securitySettings.enableMfaDescription')}
              </p>
              {!setupData ? (
                <Button onClick={handleStartSetup} disabled={working || loadingStatus} className="ff-btn-velocity">
                  {working ? t('securitySettings.starting') : t('securitySettings.setUpAuthenticatorApp')}
                </Button>
              ) : (
                <div className="space-y-4">
                  <div className="flex flex-col gap-4 md:flex-row md:items-start">
                    <div className="rounded-lg border border-border-default bg-bg-primary p-3">
                      {qrCodeValue ? (
                        <QRCodeSVG
                          value={qrCodeValue}
                          size={180}
                          bgColor="transparent"
                          fgColor="var(--text-primary)"
                          level="M"
                        />
                      ) : (
                        <div className="h-[180px] w-[180px] flex items-center justify-center text-xs text-text-muted">
                          {t('securitySettings.loadingQr')}
                        </div>
                      )}
                    </div>
                    <div className="space-y-2">
                      <p className="text-sm text-text-secondary">
                        {t('securitySettings.scanQrCode')}
                      </p>
                      <div className="flex items-center gap-2 flex-wrap">
                        <p className="text-xs text-text-muted">{t('securitySettings.manualKey')}</p>
                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          onClick={() => setShowMfaSecret((v) => !v)}
                          className="h-7 px-2 text-xs"
                        >
                          {showMfaSecret ? t('securitySettings.hide') : t('securitySettings.show')}
                        </Button>
                      </div>
                      {showMfaSecret ? (
                        <code className="inline-block break-all rounded bg-bg-primary px-2 py-1 text-xs">
                          {setupData.secret}
                        </code>
                      ) : (
                        <code className="inline-block break-all rounded bg-bg-primary px-2 py-1 text-xs text-text-muted">
                          {t('securitySettings.secretHidden')}
                        </code>
                      )}
                    </div>
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="mfa-verify-code">{t('securitySettings.verificationCode')}</Label>
                    <Input
                      id="mfa-verify-code"
                      placeholder="123456"
                      inputMode="numeric"
                      value={verificationCode}
                      onChange={(e) =>
                        setVerificationCode(e.target.value.replace(/[^\d]/g, '').slice(0, 6))
                      }
                    />
                  </div>
                  <div className="flex gap-2">
                    <Button onClick={handleEnableMFA} disabled={working} className="ff-btn-velocity">
                      {working ? t('securitySettings.enabling') : t('securitySettings.enableMfa')}
                    </Button>
                    <Button
                      variant="outline"
                      onClick={() => {
                        setSetupData(null);
                        setVerificationCode('');
                      }}
                      disabled={working}
                    >
                      {t('securitySettings.cancel')}
                    </Button>
                  </div>
                </div>
              )}
            </div>
          ) : (
            <div className="rounded-lg border border-border-default bg-bg-secondary p-4 space-y-4">
              <div className="flex items-start gap-2 text-sm text-text-secondary">
                <CheckCircle2 className="mt-0.5 h-4 w-4 text-green-500" />
                <p>{t('securitySettings.accountProtectedWithMfa')}</p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="disable-password">{t('securitySettings.currentPassword')}</Label>
                <Input
                  id="disable-password"
                  type="password"
                  value={disablePassword}
                  onChange={(e) => setDisablePassword(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="disable-code">{t('securitySettings.authenticatorCode')}</Label>
                <Input
                  id="disable-code"
                  inputMode="numeric"
                  placeholder="123456"
                  value={disableCode}
                  onChange={(e) => setDisableCode(e.target.value.replace(/[^\d]/g, '').slice(0, 6))}
                />
              </div>
              <Button
                variant="destructive"
                onClick={handleDisableMFA}
                disabled={working || status?.required}
              >
                {working ? t('securitySettings.disabling') : t('securitySettings.disableMfa')}
              </Button>
              {status?.required ? (
                <p className="flex items-center gap-2 text-xs text-amber-600 dark:text-amber-400">
                  <AlertTriangle className="h-3.5 w-3.5" />
                  {t('securitySettings.mfaRequiredForRole')}
                </p>
              ) : null}
            </div>
          )}
        </CardContent>
      </Card>

      {recoveryCodes.length > 0 ? (
        <Card className="settings-panel">
          <CardHeader>
            <CardTitle className="font-display flex items-center gap-2">
              <KeyRound className="h-5 w-5 text-brand-500" />
              {t('securitySettings.recoveryCodesTitle')}
            </CardTitle>
            <CardDescription className="text-text-secondary">
              {t('securitySettings.recoveryCodesDescription')}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="mb-3 flex items-center justify-end">
              <Button
                variant="outline"
                size="sm"
                onClick={() => void handleDownloadRecoveryCodes()}
              >
                {t('securitySettings.downloadCodes')}
              </Button>
            </div>
            <div className="grid gap-2 sm:grid-cols-2">
              {recoveryCodes.map((code) => (
                <code
                  key={code}
                  className="rounded border border-border-default bg-bg-secondary px-3 py-2 text-xs font-mono"
                >
                  {code}
                </code>
              ))}
            </div>
          </CardContent>
        </Card>
      ) : null}

      <Card className="settings-panel">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Monitor className="h-5 w-5 text-brand-500" />
              <CardTitle className="font-display">{t('securitySettings.sessionsDevicesTitle', 'Sessions & Devices')}</CardTitle>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-xs text-text-muted">
                {displaySessions.filter(s => !s.currentSession).length} {t('securitySettings.otherDevices', 'other devices')}
              </span>
            </div>
          </div>
          <CardDescription className="text-text-secondary">
            {t('securitySettings.sessionsDevicesDescription', 'Manage active sessions across all your devices. Revoke any session you don\'t recognize.')}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <Tabs defaultValue="active" className="w-full">
            <TabsList className="grid w-full grid-cols-3 h-auto rounded-lg border border-border-default bg-bg-secondary/80 p-1 gap-1">
              <TabsTrigger
                value="active"
                className="flex items-center gap-1.5 rounded-md px-3 py-2 text-xs font-medium data-[state=active]:bg-brand-500/10 data-[state=active]:text-brand-500 data-[state=active]:shadow-glow-sm transition-all"
              >
                <Server className="h-3.5 w-3.5" />
                {t('securitySettings.tabActive', 'Active')}
                <Badge variant="secondary" className="ml-1 text-xs px-1.5 py-0.5">
                  {displaySessions.filter(s => !s.currentSession).length + 1}
                </Badge>
              </TabsTrigger>
              <TabsTrigger
                value="history"
                className="flex items-center gap-1.5 rounded-md px-3 py-2 text-xs font-medium data-[state=active]:bg-brand-500/10 data-[state=active]:text-brand-500 data-[state=active]:shadow-glow-sm transition-all"
              >
                <Globe className="h-3.5 w-3.5" />
                {t('securitySettings.tabHistory', 'History')}
              </TabsTrigger>
              <TabsTrigger
                value="settings"
                className="flex items-center gap-1.5 rounded-md px-3 py-2 text-xs font-medium data-[state=active]:bg-brand-500/10 data-[state=active]:text-brand-500 data-[state=active]:shadow-glow-sm transition-all"
              >
                <ShieldCheck className="h-3.5 w-3.5" />
                {t('securitySettings.tabSecurity', 'Security')}
              </TabsTrigger>
            </TabsList>

            <TabsContent value="active" className="mt-4 space-y-3">
              <div className="flex items-center justify-between rounded-lg border border-border-default bg-bg-secondary/50 px-4 py-3">
                <div className="flex items-center gap-3">
                  <div className="flex h-10 w-10 items-center justify-center rounded-full bg-brand-500/10">
                    <Monitor className="h-5 w-5 text-brand-500" />
                  </div>
                  <div className="space-y-0.5">
                    <p className="text-sm font-medium text-text-primary">{t('securitySettings.currentSessionLabel', 'Current Session')}</p>
                    <p className="text-xs text-text-muted">
                      {t('securitySettings.thisDeviceNote', 'This is your current device and cannot be revoked')}
                    </p>
                  </div>
                </div>
                <Badge variant="outline" className="ff-status-active border-green-500/40 text-green-600 dark:text-green-400">
                  <span className="flex items-center gap-1">
                    <span className="h-1.5 w-1.5 rounded-full bg-green-500 animate-pulse" />
                    {t('securitySettings.activeBadge', 'Active')}
                  </span>
                </Badge>
              </div>

              {displaySessions.filter(s => !s.currentSession).length > 0 ? (
                <div className="space-y-2">
                  <div className="flex items-center justify-between px-1">
                    <span className="text-xs font-medium text-text-muted uppercase tracking-wide">
                      {t('securitySettings.signedInElsewhere', 'Signed in elsewhere')}
                    </span>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={handleRevokeOthers}
                      disabled={revokingOthers || loadingSessions}
                      className="text-xs text-destructive hover:text-destructive h-7 px-2"
                    >
                      <LogOut className="h-3 w-3 mr-1" />
                      {revokingOthers ? t('securitySettings.signingOut') : t('securitySettings.signOutAllOther', 'Sign out all')}
                    </Button>
                  </div>
                  {displaySessions.filter(s => !s.currentSession).map((session) => (
                    <div
                      key={session.id}
                      className="group flex items-center justify-between rounded-lg border border-border-default bg-bg-secondary p-3 hover:border-brand-500/30 transition-colors"
                    >
                      <div className="flex items-center gap-3 min-w-0">
                        <div className="flex h-9 w-9 items-center justify-center rounded-full bg-bg-tertiary shrink-0">
                          {getSessionDeviceIcon(session.device, 'lg')}
                        </div>
                        <div className="min-w-0">
                          <div className="flex items-center gap-2">
                            <p className="text-sm font-medium text-text-primary truncate">{session.device}</p>
                            {differenceInMinutes(new Date(), new Date(session.lastActive)) < 60 && (
                              <Badge variant="outline" className="text-[10px] px-1.5 py-0 border-amber-500/40 text-amber-600 dark:text-amber-400">
                                {t('securitySettings.recentBadge', 'Recent')}
                              </Badge>
                            )}
                          </div>
                          <div className="flex items-center gap-2 text-xs text-text-muted mt-0.5">
                            <span className="font-mono">{session.ip}</span>
                            {session.location && (
                              <>
                                <span>·</span>
                                <span>{session.location}</span>
                              </>
                            )}
                          </div>
                          <p className="text-xs text-text-muted/70 mt-0.5">
                            {t('securitySettings.lastActiveAt', 'Last active')} {session.lastActive}
                          </p>
                        </div>
                      </div>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleRevokeSession(session.id)}
                        disabled={revokingSessionId === session.id}
                        className="shrink-0 opacity-0 group-hover:opacity-100 transition-opacity text-xs h-8"
                      >
                        {revokingSessionId === session.id ? (
                          t('securitySettings.revoking')
                        ) : (
                          <>
                            <LogOut className="h-3 w-3 mr-1" />
                            {t('securitySettings.signOut', 'Sign out')}
                          </>
                        )}
                      </Button>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="flex flex-col items-center justify-center rounded-lg border border-dashed border-border-default bg-bg-secondary/30 py-8 px-4 text-center">
                  <Server className="h-8 w-8 text-text-muted mb-2" />
                  <p className="text-sm text-text-muted">{t('securitySettings.noOtherSessions', 'No other active sessions')}</p>
                  <p className="text-xs text-text-muted/70 mt-1">{t('securitySettings.allDevicesListedNote', 'All your devices will appear here when you sign in')}</p>
                </div>
              )}
            </TabsContent>

            <TabsContent value="history" className="mt-4 space-y-3">
              <div className="flex items-center justify-between rounded-lg border border-border-default bg-bg-secondary/50 px-4 py-3">
                <div className="flex items-center gap-3">
                  <div className="flex h-10 w-10 items-center justify-center rounded-full bg-blue-500/10">
                    <Globe className="h-5 w-5 text-blue-500" />
                  </div>
                  <div className="space-y-0.5">
                    <p className="text-sm font-medium text-text-primary">{t('securitySettings.signInHistoryTitle', 'Sign-in History')}</p>
                    <p className="text-xs text-text-muted">
                      {t('securitySettings.signInHistoryDescription', 'View your historical login activity across all devices')}
                    </p>
                  </div>
                </div>
                <Button variant="outline" size="sm" className="text-xs">
                  <Download className="h-3 w-3 mr-1" />
                  {t('securitySettings.downloadHistory', 'Download')}
                </Button>
              </div>
              <div className="rounded-lg border border-border-default bg-bg-secondary/30 p-6 text-center">
                <ListFilter className="h-8 w-8 text-text-muted mx-auto mb-2" />
                <p className="text-sm text-text-muted">{t('securitySettings.historyComingSoon', 'Full sign-in history coming soon')}</p>
                <p className="text-xs text-text-muted/70 mt-1">{t('securitySettings.historyNote', 'We track device, IP, location, and timestamp for each sign-in')}</p>
              </div>
            </TabsContent>

            <TabsContent value="settings" className="mt-4 space-y-4">
              <div className="rounded-lg border border-border-default bg-bg-secondary/50 p-4 space-y-4">
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <p className="text-sm font-medium text-text-primary">{t('securitySettings.rememberDevicesTitle', 'Remember trusted devices')}</p>
                    <p className="text-xs text-text-muted">
                      {t('securitySettings.rememberDevicesDescription', 'Allow sessions on recognized devices for 30 days')}
                    </p>
                  </div>
                  <Switch checked={rememberDevices} onCheckedChange={handleRememberDevicesChange} />
                </div>
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <p className="text-sm font-medium text-text-primary">{t('securitySettings.sessionTimeoutTitle', 'Session timeout')}</p>
                    <p className="text-xs text-text-muted">
                      {t('securitySettings.sessionTimeoutDescription', 'Automatically sign out after period of inactivity')}
                    </p>
                  </div>
                  <Button variant="outline" size="sm" className="text-xs" onClick={() => setTimeoutModalOpen(true)}>
                    {t('securitySettings.configureTimeout', 'Configure')}
                  </Button>
                </div>
              </div>
              <div className="rounded-lg border border-amber-500/20 bg-amber-500/5 p-4">
                <div className="flex items-start gap-3 mb-4">
                  <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-amber-500/10 border border-amber-500/20">
                    <ShieldAlert className="h-4 w-4 text-amber-500" />
                  </div>
                  <div className="space-y-1 min-w-0">
                    <p className="text-sm font-semibold text-amber-600 dark:text-amber-400">
                      {t('securitySettings.suspiciousActivityTitle', 'Suspicious activity?')}
                    </p>
                    <p className="text-xs text-text-muted leading-relaxed">
                      {t('securitySettings.suspiciousActivityDescription', 'If you see unrecognized sessions, revoke them immediately and change your password')}
                    </p>
                  </div>
                </div>
                <Button
                  variant="destructive"
                  size="sm"
                  onClick={handleRevokeOthers}
                  disabled={revokingOthers || loadingSessions}
                  className="w-full gap-2 text-xs font-medium shadow-glow-sm"
                >
                  <LogOut className="h-3.5 w-3.5" />
                  {revokingOthers
                    ? t('securitySettings.signingOut')
                    : t('securitySettings.revokeAllOtherSessions', 'Revoke all other sessions')}
                </Button>
              </div>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>

      <Card className="settings-panel">
        <CardHeader>
          <CardTitle className="font-display flex items-center gap-2">
            <Download className="h-5 w-5 text-brand-500" />
            {t('securitySettings.dataExportTitle')}
          </CardTitle>
          <CardDescription className="text-text-secondary">
            {t('securitySettings.dataExportDescription')}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <p className="text-sm text-text-secondary">
            {t('securitySettings.dataExportBody')}
          </p>
          <Button onClick={handleExportData} disabled={exportingData} className="ff-btn-velocity">
            {exportingData ? t('securitySettings.preparingExport') : t('securitySettings.exportMyData')}
          </Button>
        </CardContent>
      </Card>

      <Card className="settings-panel border-red-500/30">
        <CardHeader>
          <CardTitle className="font-display flex items-center gap-2 text-red-600 dark:text-red-400">
            <ShieldAlert className="h-5 w-5" />
            {t('securitySettings.dangerZoneTitle')}
          </CardTitle>
          <CardDescription className="text-text-secondary">
            {t('securitySettings.dangerZoneDescription')}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <p className="text-sm text-text-secondary">
            {t('securitySettings.deleteAccountWarning')}
          </p>
          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button variant="destructive">
                <Trash2 className="mr-2 h-4 w-4" />
                {t('securitySettings.deleteAccount')}
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>{t('securitySettings.deleteAccountConfirm')}</AlertDialogTitle>
                <AlertDialogDescription>
                  <Trans i18nKey="securitySettings.deleteAccountDescription">
                    This action is permanent. Type <strong>DELETE</strong> to confirm.
                  </Trans>
                </AlertDialogDescription>
              </AlertDialogHeader>
              <div className="space-y-2">
                <Label htmlFor="delete-confirmation">{t('securitySettings.confirmation')}</Label>
                <Input
                  id="delete-confirmation"
                  placeholder={t('securitySettings.typeDelete')}
                  value={deleteConfirmation}
                  onChange={(e) => setDeleteConfirmation(e.target.value)}
                />
              </div>
              <AlertDialogFooter>
                <AlertDialogCancel onClick={() => setDeleteConfirmation('')}>
                  {t('securitySettings.cancel')}
                </AlertDialogCancel>
                <AlertDialogAction
                  className="bg-red-600 hover:bg-red-700"
                  onClick={(e) => {
                    e.preventDefault();
                    void handleDeleteAccount();
                  }}
                  disabled={deletingAccount || deleteConfirmation !== 'DELETE'}
                >
                  {deletingAccount ? t('securitySettings.deleting') : t('securitySettings.permanentlyDelete')}
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>

          <Dialog open={timeoutModalOpen} onOpenChange={setTimeoutModalOpen}>
            <DialogContent className="sm:max-w-md">
              <DialogHeader>
                <div className="flex items-center gap-3">
                  <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-brand-500/10">
                    <Clock className="h-5 w-5 text-brand-500" />
                  </div>
                  <div>
                    <DialogTitle>{t('securitySettings.sessionTimeoutTitle', 'Session Timeout')}</DialogTitle>
                    <DialogDescription className="text-xs mt-0.5">
                      {t('securitySettings.sessionTimeoutModalDescription', 'Choose how long before your session expires due to inactivity')}
                    </DialogDescription>
                  </div>
                </div>
              </DialogHeader>
              <form onSubmit={(e) => { e.preventDefault(); void handleSaveTimeout(); }}>
              <div className="space-y-2">
                {timeoutOptions.map((option) => (
                  <button
                    type="button"
                    key={option.value}
                    onClick={() => setSelectedTimeout(option.value)}
                    className={`w-full flex items-start gap-3 rounded-lg border p-3 text-left transition-all ${
                      selectedTimeout === option.value
                        ? 'border-brand-500/50 bg-brand-500/5 ring-1 ring-brand-500/30'
                        : 'border-border-default bg-bg-secondary hover:border-brand-500/30'
                    }`}
                  >
                    <div className={`mt-0.5 h-4 w-4 shrink-0 rounded-full border-2 ${
                      selectedTimeout === option.value
                        ? 'border-brand-500 bg-brand-500'
                        : 'border-border-default'
                    }`}>
                      {selectedTimeout === option.value && (
                        <div className="flex items-center justify-center">
                          <CheckCircle2 className="h-2.5 w-2.5 text-white" />
                        </div>
                      )}
                    </div>
                    <div className="space-y-0.5 min-w-0">
                      <p className="text-sm font-medium text-text-primary">{option.label}</p>
                      <p className="text-xs text-text-muted">{option.description}</p>
                    </div>
                  </button>
                ))}
              </div>
              <DialogFooter className="gap-2 sm:gap-0 mt-4">
                <Button type="button" variant="outline" size="sm" onClick={() => setTimeoutModalOpen(false)} className="text-xs">
                  {t('securitySettings.cancel', 'Cancel')}
                </Button>
                <Button
                  type="submit"
                  size="sm"
                  disabled={savingTimeout}
                  className="ff-btn-velocity text-xs gap-1.5"
                >
                  {savingTimeout ? (
                    <>
                      <RefreshCw className="h-3 w-3 animate-spin" />
                      {t('securitySettings.saving', 'Saving...')}
                    </>
                  ) : (
                    <>
                      <CheckCircle2 className="h-3 w-3" />
                      {t('securitySettings.saveTimeout', 'Save')}
                    </>
                  )}
                </Button>
              </DialogFooter>
            </form>
            </DialogContent>
          </Dialog>
        </CardContent>
      </Card>
    </div>
  );
}
